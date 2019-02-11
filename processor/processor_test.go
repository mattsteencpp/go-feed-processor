package processor

import "testing"

func TestGetConfig(t *testing.T) {
	config, err := GetConfig("/home/msteen/personal/go/src/github.com/mattsteencpp/go-feed-processor/config/test_config.json")
	if err != nil {
		t.Errorf("Could not load config")
	}

	expectedLink := "http://feed.thisamericanlife.org/talpodcast"

	if config.Link != expectedLink {
		t.Errorf("config.Link == %q, got %q", expectedLink, config.Link)
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

/* func TestGetFeedBody(t *testing.T) {

}
*/

/*
func TestParseFeed(t *testing.T) {

}
*/
