package main

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
	"fmt"
	"os"
	"path/filepath"
	"github.com/mattsteencpp/go-feed-processor/processor"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("You must give the name of a feed to process.\n")
		return
	}
	configFilename := os.Args[1] + ".json"
	config := processor.GetConfig(filepath.Join("/home/msteen/personal/go/src/github.com/mattsteencpp/go-feed-processor/config/", configFilename))

	body := processor.GetFeedBody(config.Link)

	processor.ParseFeed(config, body)
}
