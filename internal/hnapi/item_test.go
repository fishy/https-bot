package hnapi_test

import (
	"reflect"
	"testing"

	"github.com/fishy/https-bot/internal/hnapi"
)

func TestItemURLs(t *testing.T) {
	item := &hnapi.Item{
		// From https://hacker-news.firebaseio.com/v0/item/26597278.json?print=pretty
		Text: "If you&#x27;re willing to share your HN username, and an email address there&#x27;s <a href=\"http:&#x2F;&#x2F;www.hnreplies.com&#x2F;\" rel=\"nofollow\">http:&#x2F;&#x2F;www.hnreplies.com&#x2F;</a>. I&#x27;ve used it for several years. Previous discussions <a href=\"https:&#x2F;&#x2F;hn.algolia.com&#x2F;?dateRange=all&amp;page=0&amp;prefix=false&amp;query=hnreplies.com&amp;sort=byPopularity&amp;type=story\" rel=\"nofollow\">https:&#x2F;&#x2F;hn.algolia.com&#x2F;?dateRange=all&amp;page=0&amp;prefix=false&amp;qu...</a>",
	}
	urls := item.URLs()
	expect := []string{
		`http://www.hnreplies.com/`,
		`https://hn.algolia.com/?dateRange=all&page=0&prefix=false&query=hnreplies.com&sort=byPopularity&type=story`,
	}
	if !reflect.DeepEqual(urls, expect) {
		t.Errorf("URLs expected %#v, got %#v", expect, urls)
	}
}
