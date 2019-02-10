package processor

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// XML data structures for feeds

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Items   []Item   `xml:"item"`
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Item   `xml:"entry"`
}

type Item struct {
	Included    bool       "-"
	Title       string     `xml:"title"`
	Link        string     "-"
	RawLink     string     `xml:"link"`
	OrigLink    string     `xml:"origLink"` // ignore feedburner xmlns (tag is feedburner:origLink)
	ItunesLink  ItunesLink `xml:"enclosure"`
	Date        string     "-"
	RawDate     string     `xml:"pubDate"`
	Published   string     `xml:"published"`
	ID          string     "-"
	RawID       string     `xml:"id"`
	GUID        string     `xml:"guid"`
	Author      string     "-"
	RawAuthor   string     `xml:"author>name"`
	Creator     string     `xml:"creator"` // ignore dc xmlns (tag is dc:creator)
	Content     string     "-"
	RawContents []Content  `xml:"content"`
	EnContent   string     `xml:"encoded"` // ignore content xmlns (tag is content:encoded)
}

type ItunesLink struct {
	Url string `xml:"url,attr"`
}

// we have to get tricky here because of the optional media:content tag
type Content struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",innerxml"` //chardata"`
}

// JSON data structures for config files

type Config struct {
	Link  string
	Files []File
}

type File struct {
	Filename     string
	IncludeAll   bool
	EpisodeRegex string
	Titles       []string
	Content      []string
	Authors      []string
	Links        []string
}

func GetFeedBody(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("There was an error\n")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
}

func GetConfig(filename string) Config {
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

func includeItem(item *Item, matchType string, match string) {
	fmt.Printf("Item found: %s by %s (%s)\n", item.Title, item.Author, item.Link)
	fmt.Printf("   ID: %s\n", item.ID)
	fmt.Printf("   Author: %s\n", item.Author)
	fmt.Printf("   Published: %s\n", item.Date)
	fmt.Printf("   Content: %s\n", item.Content)
	fmt.Printf("   Matched on %s: %s\n", matchType, match)
	item.Included = true
}

func processItemElement(item *Item) {
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

func ParseFeed(config Config, body []byte) {
	var rss RSS
	xml.Unmarshal(body, &rss)

	if len(rss.Channel.Items) > 0 {
		for i := 0; i < len(rss.Channel.Items); i++ {
			processItemElement(&rss.Channel.Items[i])
		}
		processItems(config, rss.Channel.Items)
	} else {
		var feed Feed
		xml.Unmarshal(body, &feed)
		if len(feed.Entries) > 0 {
			for i := 0; i < len(feed.Entries); i++ {
				processItemElement(&feed.Entries[i])
			}
			processItems(config, feed.Entries)
		} else {
			fmt.Printf("No items found in feed\n")
		}
	}
}

func processItems(config Config, items []Item) {
	for j := 0; j < len(config.Files); j++ {
		file := config.Files[j]
		fmt.Printf("Looking for items for file %s\n", file.Filename)
		for i := 0; i < len(items); i++ {
			item := &items[i]
			if item.Included {
				continue
			}
			if file.IncludeAll {
				includeItem(item, "all", "all")
			} else {
				if len(file.Titles) > 0 {
					for k := 0; k < len(file.Titles); k++ {
						if strings.Contains(item.Title, file.Titles[k]) {
							includeItem(item, "title", file.Titles[k])
						}
					}
				}
				if !item.Included && len(file.Authors) > 0 {
					for k := 0; k < len(file.Authors); k++ {
						if strings.Contains(item.Author, file.Authors[k]) {
							includeItem(item, "author", file.Authors[k])
						}
					}
				}
				if !item.Included && len(file.Content) > 0 {
					for k := 0; k < len(file.Content); k++ {
						if strings.Contains(item.Content, file.Content[k]) {
							includeItem(item, "content", file.Content[k])
						}
					}
				}
				if !item.Included && len(file.Links) > 0 {
					for k := 0; k < len(file.Links); k++ {
						if strings.Contains(item.Link, file.Links[k]) {
							includeItem(item, "link", file.Links[k])
						}
					}
				}
			}
		}
		fmt.Printf("\n\n")
	}
}
