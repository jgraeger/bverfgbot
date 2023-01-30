package bverfg

import (
	"fmt"
	"sort"

	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"
)

// FeedTranslator is a translator for the gofeed package, providing
// a convenient use for the shitty government site builder rss feeds.
type FeedTranslator struct {
	dt *gofeed.DefaultRSSTranslator
}

func NewFeedTranslator() *FeedTranslator {
	return &FeedTranslator{
		dt: &gofeed.DefaultRSSTranslator{},
	}
}

func (ft *FeedTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
	rss, isRSSFeed := feed.(*rss.Feed)
	if !isRSSFeed {
		return nil, fmt.Errorf("feed didn't match expected RSS format")
	}

	f, err := ft.dt.Translate(rss)
	if err != nil {
		return nil, err
	}

	if len(f.Items) == 0 {
		return f, nil
	}

	// Sort items from latest to oldest
	sort.Sort(sort.Reverse(f))
	latestItem := f.Items[0]

	if latestItem.Published != "" {
		f.Published = latestItem.Published
	}

	return f, nil
}
