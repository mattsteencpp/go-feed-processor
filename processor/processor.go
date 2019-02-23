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

type Item struct {
	MatchDetails string     "-"
	Included     bool       "-"
	Title        string     `xml:"title"`
	Link         string     "-"
	RawLink      string     `xml:"link"`
	ItunesLink   ItunesLink `xml:"enclosure"`
	Date         string     `xml:"pubDate"`
	Published    string     `xml:"published"`
	ID           string     "-"
	RawID        string     `xml:"id"`
	GUID         string     `xml:"guid"`
	Author       string     "-"
	RawAuthor    string     `xml:"author>name"`
	Creator      string     `xml:"creator"` // ignore dc xmlns (tag is dc:creator)
	Content      string     `xml:"encoded"` // ignore content xmlns (tag is content:encoded)
}

// alternate feed structure; currently used only by The Atlantic
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
	Items   []Item
}

type Entry struct {
	Included    bool       "-"
	Title       string     `xml:"title"`
	Link        string     `xml:"origLink"` // ignore feedburner xmlns (tag is feedburner:origLink)
	ItunesLink  ItunesLink `xml:"enclosure"`
	Date        string     `xml:"published"`
	ID          string     `xml:"id"`
	Author      string     `xml:"author>name"`
	RawContents []Content  `xml:"content"`
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
	Link              string
	FeedType          string
	ShouldFindAuthor  bool
	ShouldFindContent bool
	ShouldFindID      bool
	Files             []File
}

type File struct {
	Filename            string
	IncludeAll          bool
	EpisodeRegex        string
	Titles              []string
	Content             []string
	Authors             []string
	Links               []string
	ExpectedTestMatches int
}

func GetFeedBody(feedLocation string) ([]byte, error) {
	var body []byte
	var err error
	if strings.Contains(feedLocation, "http") {
		resp, err := http.Get(feedLocation)
		if err != nil {
			fmt.Printf("There was an error fetching the feed\n")
			return nil, err
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("There was an error getting the feed response from the HTTP request\n")
			return nil, err
		}
	} else {
		body, err = ioutil.ReadFile(feedLocation)
		if err != nil {
			fmt.Printf("There was an error reading the feed from file\n")
			return nil, err
		}
	}
	return body, nil
}

func GetConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("error:", err)
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
		return config, err
	}
	return config, nil
}

func processItemElement(item *Item) {
	item.Link = item.RawLink
	if item.Link == "" {
		item.Link = item.ItunesLink.Url // freakonomics, fresh air, planet money
	}
	item.ID = item.RawID
	if item.ID == "" {
		item.ID = item.GUID // arseblog, freakonomics, fresh air, planet money
	}
	item.Author = item.RawAuthor
	if item.Author == "" {
		item.Author = item.Creator // arseblog, tal
	}
}

func convertEntryToItem(entry *Entry, item *Item) {
	item.Title = entry.Title
	item.Link = entry.Link
	item.Date = entry.Date
	item.Author = entry.Author
	item.ID = entry.ID
	for i := 0; i < len(entry.RawContents); i++ {
		if entry.RawContents[i].Type == "html" {
			item.Content = entry.RawContents[i].Value
		}
	}
}

func ParseFeed(config Config, body []byte) {
	items := GetItems(config, body)
	processItems(config, items)
}

func GetItems(config Config, body []byte) (items []Item) {
	if config.FeedType == "standard" {
		var rss RSS
		xml.Unmarshal(body, &rss)

		if len(rss.Channel.Items) == 0 {
			fmt.Println("No items found of feed type standard")
		}

		for i := 0; i < len(rss.Channel.Items); i++ {
			processItemElement(&rss.Channel.Items[i])
		}
		return rss.Channel.Items
	} else if config.FeedType == "alternate" {
		var feed Feed
		xml.Unmarshal(body, &feed)

		if len(feed.Entries) == 0 {
			fmt.Println("No items found for feed type alternate")
		}

		feed.Items = make([]Item, len(feed.Entries))

		for i := 0; i < len(feed.Entries); i++ {
			convertEntryToItem(&feed.Entries[i], &feed.Items[i])
		}
		return feed.Items
	}
	return nil
}

func processItems(config Config, items []Item) {
	for i := 0; i < len(config.Files); i++ {
		file := config.Files[i]
		fmt.Printf("Looking for items for file %s\n", file.Filename)

		matchingItems := GetMatchingItems(file, items)
		for j := 0; j < len(matchingItems); j++ {
			item := matchingItems[j]
			printMatchingItem(item)
		}

		fmt.Printf("\n\n")
	}
}

func GetMatchingItems(file File, items []Item) (matchingItems []*Item) {
	matchingItems = make([]*Item, 0)
	for i := 0; i < len(items); i++ {
		item := &items[i]
		if item.Included {
			continue
		}
		if file.IncludeAll {
			matchingItems = includeItem(matchingItems, item, "all", "all")
		} else {
			if len(file.Titles) > 0 {
				for k := 0; k < len(file.Titles); k++ {
					if strings.Contains(item.Title, file.Titles[k]) {
						matchingItems = includeItem(matchingItems, item, "title", file.Titles[k])
					}
				}
			}
			if !item.Included && len(file.Authors) > 0 {
				for k := 0; k < len(file.Authors); k++ {
					if strings.Contains(item.Author, file.Authors[k]) {
						matchingItems = includeItem(matchingItems, item, "author", file.Authors[k])
					}
				}
			}
			if !item.Included && len(file.Content) > 0 {
				for k := 0; k < len(file.Content); k++ {
					if strings.Contains(item.Content, file.Content[k]) {
						matchingItems = includeItem(matchingItems, item, "content", file.Content[k])
					}
				}
			}
			if !item.Included && len(file.Links) > 0 {
				for k := 0; k < len(file.Links); k++ {
					if strings.Contains(item.Link, file.Links[k]) {
						matchingItems = includeItem(matchingItems, item, "link", file.Links[k])
					}
				}
			}
		}
	}
	return matchingItems
}

func includeItem(matchingItems []*Item, item *Item, matchType string, match string) (updatedMatchingItems []*Item) {
	item.MatchDetails = fmt.Sprintf("   Matched on %s: %s\n", matchType, match)
	matchingItems = append(matchingItems, item)

	return matchingItems
}

func printMatchingItem(item *Item) {
	fmt.Printf("Item found: %s by %s (%s)\n", item.Title, item.Author, item.Link)
	fmt.Printf("   ID: %s\n", item.ID)
	fmt.Printf("   Author: %s\n", item.Author)
	fmt.Printf("   Published: %s\n", item.Date)
	if len(item.Content) > 100 {
		fmt.Printf("   Content: %s...\n", item.Content[0:100])
	} else {
		fmt.Printf("   Content: %s...\n", item.Content)
	}
	fmt.Printf("   %s\n", item.MatchDetails)
	item.Included = true
}
