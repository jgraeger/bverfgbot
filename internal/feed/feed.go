package feed

import (
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

type Feed struct {
	mu     sync.RWMutex
	closed bool

	url             string
	refreshInterval time.Duration
	subscriptions   []chan gofeed.Feed
}

func NewSubscription(url string) (*Feed, error) {
	return nil, nil
}

func (f *Feed) Subscribe() chan gofeed.Feed {
	f.mu.Lock()
	defer f.mu.Unlock()

	ch := make(chan gofeed.Feed, 1)
	f.subscriptions = append(f.subscriptions, ch)
	return ch
}

func (f *Feed) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.closed {
		f.closed = true
		for _, ch := range f.subscriptions {
			close(ch)
		}
	}
}

func (f *Feed) publish(feed gofeed.Feed) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return
	}

	for _, ch := range f.subscriptions {
		go func(ch chan gofeed.Feed) {
			ch <- feed
		}(ch)
	}
}
