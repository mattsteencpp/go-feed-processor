package main

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
	Channel Channel `xml:"channel"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Items []Item `xml:"item"`
}

type Item struct {
	XMLName xml.Name `xml:"item"`
	Included bool
	Title    string   `xml:"title"`
	Link     string   `xml:"link"`
	Date     string   `xml:"pubDate"`
	Author   string   `xml:"author"`
	Contents string   `xml:"content"`
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
	Contents	[]string
	Authors		[]string
}

func get_feed_body(url string) ([]byte) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("There was an error/n")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
}

// idea: would be cool to record every time we match on a specific property
// then I could clean up the config to remove outdated exclusions
// - store in dynamodb (free up to 25GB)

// potential issue: unmarshal ignores data that doesn't match a defined attribute.
// that makes it easy to lose metadata or other properties


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

func main() {
	config := get_config("/home/msteen/personal/go-feed-processor/lambdas/arseblog.json")
	body := get_feed_body(config.Link)
	var rss RSS
	xml.Unmarshal(body, &rss)
	if len(rss.Channel.Items) == 0 {
		fmt.Printf("No items found\n")
	}
	for j := 0; j < len(config.Files); j++ {
		file := config.Files[j]
		fmt.Printf("Looking for items for file %s\n", file.Filename)
		for i := 0; i < len(rss.Channel.Items); i++ {
			// make this a pointer or something?
			item := rss.Channel.Items[i]
			// why isn't this working?!?!
			if item.Included {
				continue
			}
			if file.Include_All {
				fmt.Printf("Item found: %s (%s)\n", item.Title, item.Link)
			} else {
				if len(file.Titles) > 0 {
					for k := 0; k < len(file.Titles); k++ {
						if strings.Contains(item.Title, file.Titles[k]) {
							fmt.Printf("Item found: %s (%s)\n", item.Title, item.Link)
							item.Included = true
							rss.Channel.Items[i] = item
						}
					}
				}
				if !item.Included && len(file.Authors) > 0 {
					for k := 0; k < len(file.Authors); k++ {
						if strings.Contains(item.Author, file.Authors[k]) {
							fmt.Printf("Item found: %s (%s)\n", item.Title, item.Link)
							item.Included = true
							rss.Channel.Items[i] = item
						}
					}
				}
				if !item.Included && len(file.Contents) > 0 {
					for k := 0; k < len(file.Contents); k++ {
						if strings.Contains(item.Contents, file.Contents[k]) {
							fmt.Printf("Item found: %s (%s)\n", item.Title, item.Link)
							item.Included = true
							rss.Channel.Items[i] = item
						}
					}
				}
			}
		}
		fmt.Printf("\n\n")
	}
}
