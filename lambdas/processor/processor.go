package main

// TODO: put most code into a lib; have a driver in a separate file so it can be
// easily switched out for a lambda driver

// TODO: add tests

// TODO?
// consider adding a setting for whether the feed uses atom, itunes, or not instead of trying raw every time

// TODO
// tal and planet money - need episode support
// (other feeds work...)

// TODO: would be cool to record every time we match on a specific property
// then I could clean up the config to remove outdated exclusions
// - store in dynamodb (free up to 25GB)
// need to parse dates before inserting...

// TODO: generate feed files
// need to handle updates (check id/guid)
// need to handle limited sizes (for some feeds)
// potential issue: unmarshal ignores data that doesn't match a defined attribute.
// that makes it easy to lose metadata or other properties - more important for 
// podcasts than other feeds...

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// XML data structures for feeds

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel `xml:"channel"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Items []Item     `xml:"item"`
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Item   `xml:"entry"`
}

type Item struct {
	Included     bool	    "-"
	Title        string     `xml:"title"`
	Link         string     "-"
	RawLink      string     `xml:"link"`
	OrigLink     string     `xml:"origLink"`  // ignore feedburner xmlns (tag is feedburner:origLink)
	ItunesLink   ItunesLink `xml:"enclosure"`
	Date         string     "-"
	RawDate      string     `xml:"pubDate"`
	Published    string     `xml:"published"`
	ID           string     "-"
	RawID        string     `xml:"id"`
	GUID         string     `xml:"guid"`
	Author       string     "-"
	RawAuthor    string     `xml:"author>name"`
	Creator      string     `xml:"creator"`  // ignore dc xmlns (tag is dc:creator)
	Content      string     "-"
	RawContents  []Content  `xml:"content"`
	EnContent    string     `xml:"encoded"` // ignore content xmlns (tag is content:encoded)
}

type ItunesLink struct {
	Url      string    `xml:"url,attr"`
}

// we have to get tricky here because of the optional media:content tag
type Content struct {
	Type       string   `xml:"type,attr"`
	Value	   string   `xml:",innerxml"` //chardata"`
}

// JSON data structures for config files

type Config struct {
	Link	string
	Files	[]File
}

type File struct {
	Filename	string
	Include_All bool
	Titles		[]string
	Content	    []string
	Authors		[]string
	Links		[]string
}

func get_feed_body(url string) ([]byte) {
	fmt.Print("url: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("There was an error\n")
	}
	fmt.Printf("%s\n\n", resp)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
}


func get_config(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("error:", err)
	}
	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	return config
}


func include_item(item *Item, match_type string, match string) {
	fmt.Printf("Item found: %s by %s (%s)\n", item.Title, item.Author, item.Link)
	fmt.Printf("   ID: %s\n", item.ID)
	fmt.Printf("   Author: %s\n", item.Author)
	fmt.Printf("   Published: %s\n", item.Date)
	fmt.Printf("   Content: %s\n", item.Content)
	fmt.Printf("   Matched on %s: %s\n", match_type, match)
	item.Included = true
}


func process_item_element(item *Item) {
	item.Link = item.RawLink
	if item.Link == "" {
		item.Link = item.OrigLink
	}
	if item.Link == "" {
        item.Link = item.ItunesLink.Url
	}
	item.Date = item.RawDate
	if item.Date == "" {
		item.Date = item.Published
	}
	item.ID = item.RawID
	if item.ID == "" {
		item.ID = item.GUID
	}
	item.Content = item.EnContent
	if item.Content == "" {
		for i := 0; i < len(item.RawContents); i++ {
			if item.RawContents[i].Type == "html" {
				item.Content = item.RawContents[i].Value
			}
		}
	}
	item.Author = item.RawAuthor
	if item.Author == "" {
		item.Author = item.Creator
	}
}


func parse_feed(config Config, body []byte) {
	var rss RSS
	xml.Unmarshal(body, &rss)

	if len(rss.Channel.Items) > 0 {
		for i := 0; i < len(rss.Channel.Items); i++ {
			process_item_element(&rss.Channel.Items[i])
		}
		process_items(config, rss.Channel.Items)
	} else {
		var feed Feed
		xml.Unmarshal(body, &feed)
		if len(feed.Entries) > 0 {
			for i := 0; i < len(feed.Entries); i++ {
				process_item_element(&feed.Entries[i])
			}
			process_items(config, feed.Entries)
		} else {
			fmt.Printf("No items found in feed\n")
		}
	}
}


func process_items(config Config, items []Item) {
	for j := 0; j < len(config.Files); j++ {
		file := config.Files[j]
		fmt.Printf("Looking for items for file %s\n", file.Filename)
		for i := 0; i < len(items); i++ {
			item := &items[i]
			if item.Included {
				continue
			}
			if file.Include_All {
				include_item(item, "all", "all")
			} else {
				if len(file.Titles) > 0 {
					for k := 0; k < len(file.Titles); k++ {
						if strings.Contains(item.Title, file.Titles[k]) {
							include_item(item, "title", file.Titles[k])
						}
					}
				}
				if !item.Included && len(file.Authors) > 0 {
					for k := 0; k < len(file.Authors); k++ {
						if strings.Contains(item.Author, file.Authors[k]) {
							include_item(item, "author", file.Authors[k])
						}
					}
				}
				if !item.Included && len(file.Content) > 0 {
					for k := 0; k < len(file.Content); k++ {
						if strings.Contains(item.Content, file.Content[k]) {
							include_item(item, "content", file.Content[k])
						}
					}
				}
				if !item.Included && len(file.Links) > 0 {
					for k := 0; k < len(file.Links); k++ {
						if strings.Contains(item.Link, file.Links[k]) {
							include_item(item, "link", file.Links[k])
						}
					}
				}
			}
		}
		fmt.Printf("\n\n")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("You must give the name of a feed to process.\n")
		return
	}
	config_filename := os.Args[1] + ".json"
	config := get_config(filepath.Join("/home/msteen/personal/go-feed-processor/lambdas/", config_filename))

	body := get_feed_body(config.Link)

	parse_feed(config, body)
}
