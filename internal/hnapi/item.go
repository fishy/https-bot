package hnapi

import (
	"html"
	"regexp"
)

var urlRE = regexp.MustCompile(
	`<a href="(.+?)" rel="nofollow">`,
)

// Item defines a firebase hacker news API item's json format.
//
// Ref: https://github.com/HackerNews/API/blob/master/README.md#items.
type Item struct {
	// Fields we care about in https-bot
	ID      int64  `json:"id"`
	By      string `json:"by"`
	Text    string `json:"text,omitempty"`
	URL     string `json:"url,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
	Dead    bool   `json:"dead,omitempty"`

	// Fields we don't care about in https-bot
	Type        string          `json:"type"`
	Time        TimestampSecond `json:"time"`
	Parent      int64           `json:"parent,omitempty"`
	Poll        int64           `json:"poll,omitempty"`
	Kids        []int64         `json:"kids,omitempty"`
	Score       int64           `json:"score,omitempty"`
	Title       string          `json:"title,omitempty"`
	Parts       []int64         `json:"parts,omitempty"`
	Descendants int64           `json:"descendants,omitempty"`
}

// URLs returns all the urls from this item.
//
// This could include the URL field, and the URLs in Text field.
func (i *Item) URLs() []string {
	var urls []string
	if i.URL != "" {
		urls = append(urls, i.URL)
	}
	matches := urlRE.FindAllStringSubmatch(html.UnescapeString(i.Text), -1)
	for _, match := range matches {
		urls = append(urls, match[1])
	}
	return urls
}
