package processor

import (
	"path/filepath"
	"strings"
	"testing"
)

// confirm that GetConfig reads a config file from disk and parses the json properties successfully
func TestGetConfig(t *testing.T) {
	filename := "/home/msteen/personal/go/src/github.com/mattsteencpp/go-feed-processor/config/test_config.json"
	config, err := GetConfig(filename)
	if err != nil {
		t.Errorf("Could not load config")
	}

	expectedLink := "http://feed.thisamericanlife.org/talpodcast"

	if config.Link != expectedLink {
		t.Errorf("config.Link == %q, got %q", expectedLink, config.Link)
	}

	expectedFeedType := "standard"

	if config.FeedType != expectedFeedType {
		t.Errorf("config.FeedType == %q, got %q", expectedFeedType, config.FeedType)
	}

	if len(config.Files) != 2 {
		t.Errorf("len(config.Files) == %d, got %d", len(config.Files), 2)
	}

	var expectedRegex string
	var expectedIncludeAll bool
	var expectedTitles []string

	for i := 0; i < len(config.Files); i++ {
		file := config.Files[i]
		if file.Filename == "excluded" {
			expectedRegex = "#?(\\d*):.*"
			expectedIncludeAll = false
			expectedTitles = []string{"Boring", "Repeat"}
		} else if file.Filename == "main" {
			expectedRegex = ""
			expectedIncludeAll = true
			expectedTitles = nil
		} else {
			t.Errorf("Unexpected filename found: %q. Should be either 'excluded' or 'main'i", file.Filename)
		}

		if file.EpisodeRegex != expectedRegex {
			t.Errorf("(file %q) file.EpisodeRegex == %q, got %q", file.Filename, expectedRegex, file.EpisodeRegex)
		}

		if file.IncludeAll != expectedIncludeAll {
			t.Errorf("(file %q) file.IncludeAll == %t, got %t", file.Filename, expectedIncludeAll, file.IncludeAll)
		}

		if len(file.Titles) != len(expectedTitles) {
			t.Errorf("(file %q) len(file.Titles) == %d, got %d", file.Filename, len(expectedTitles), len(file.Titles))
		}
	}
}

// confirm that GetFeedBody, given a url, fetches an xml document from that url
func TestGetFeedBody(t *testing.T) {
	url := "http://feed.thisamericanlife.org/talpodcast"
	body, err := GetFeedBody(url)

	if err != nil {
		t.Errorf("Could not load feed")
	}

	minLength := 10000
	if len(body) < minLength {
		t.Errorf("len(body) > %d, got %d", minLength, len(body))
	}

	expectedStrings := [...]string{"<?xml", "This American Life"}
	for i := 0; i < len(expectedStrings); i++ {
		expectedString := expectedStrings[i]
		if !strings.Contains(string(body), expectedString) {
			t.Errorf("body contains %q, not found. body: %q", expectedString, body)
		}
	}
}

// confirm that all feeds can be parsed from file
func TestParsing(t *testing.T) {
	feeds := []string{"arseblog", "atlantic", "freakonomics", "fresh_air", "hacker_news", "planet_money", "tal"}
	for i := 0; i < len(feeds); i++ {
		// load actual config
		configFilename := feeds[i] + ".json"
		config, err := GetConfig(filepath.Join("/home/msteen/personal/go/src/github.com/mattsteencpp/go-feed-processor/config/", configFilename))
		if err != nil {
			t.Errorf("An error occurred reading the config file %s", configFilename)
		}

		// load sample file (instead of url)
		feedFilename := feeds[i] + ".xml"
		body, err := GetFeedBody(filepath.Join("/home/msteen/personal/go/src/github.com/mattsteencpp/go-feed-processor/fixtures/", feedFilename))
		if err != nil {
			t.Errorf("An error occurred reading the example feed file %s", feedFilename)
		}

		items := GetItems(config, body)

		// parse items for each file in config
		for j := 0; j < len(config.Files); j++ {
			file := config.Files[j]

			matchingItems := GetMatchingItems(file, items)

			for k := 0; k < len(matchingItems); k++ {
				item := matchingItems[k]
				if item.Link == "" {
					t.Errorf("No link found for item in feed %s", feeds[i])
				}
				if item.Title == "" {
					t.Errorf("No title found for item in feed %s, %v", feeds[i], item)
				}
				if config.ShouldFindAuthor && item.Author == "" {
					t.Errorf("No author found for item in feed %s: %s", feeds[i], item.Title)
				}
				if config.ShouldFindContent && item.Content == "" {
					t.Errorf("No content found for item in feed %s", feeds[i])
				}
				if item.Date == "" {
					t.Errorf("No date found for item in feed %s", feeds[i])
				}
				if config.ShouldFindID && item.ID == "" {
					t.Errorf("No ID found for item in feed %s", feeds[i])
				}
			}

			if len(matchingItems) != file.ExpectedTestMatches {
				t.Errorf("Unexpected item count for feed %s, file %s. Found %d items, expected %d", feeds[i], file.Filename, len(matchingItems), file.ExpectedTestMatches)
			}
		}
	}
}
