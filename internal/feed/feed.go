package feed

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"go.uber.org/atomic"
)

const (
	fetchTimeout           = 10 * time.Second
	defaultRefreshInterval = 5 * time.Second
)

type Feed struct {
	mu     sync.RWMutex
	closed bool
	ctx    context.Context

	url             string
	refreshInterval atomic.Duration
	lastTimestamp   atomic.Time
	parser          *gofeed.Parser
	subscriptions   []chan gofeed.Feed
}

func NewFeed(ctx context.Context, url string) *Feed {
	feed := &Feed{
		ctx:             ctx,
		url:             url,
		refreshInterval: *atomic.NewDuration(defaultRefreshInterval),
		parser:          gofeed.NewParser(),
		lastTimestamp:   *atomic.NewTime(time.Time{}),
	}

	return feed
}

func (f *Feed) SetRefreshInterval(d time.Duration) {
	f.refreshInterval.Store(d)
}

func (f *Feed) SetTranslator(t gofeed.Translator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.parser.RSSTranslator = t
}

func (f *Feed) Subscribe() chan gofeed.Feed {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Lazy start fetching loop on the first subscription
	if len(f.subscriptions) == 0 {
		defer func() {
			go f.fetchLoop()
		}()
	}

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

func (f *Feed) fetchLoop() {
	for {
		select {
		case <-time.After(f.refreshInterval.Load()):
			f.fetchFeed()
		case <-f.ctx.Done():
			log.Println("shutdown feed:", f.url)
			f.Close()
			return
		}
	}
}

func (f *Feed) fetchFeed() {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	feed, err := f.parser.ParseURLWithContext(f.url, ctx)
	if err != nil {
		log.Printf("error parsing feed %v: %v\n", f.url, err)
		return
	}

	fetchedDate, err := time.Parse(time.RFC1123Z, feed.Published)
	if err != nil {
		log.Println("failed to parse date as RFC1123Z:", feed.Published)
		return
	}

	if fetchedDate.After(f.lastTimestamp.Load()) {
		f.lastTimestamp.Store(fetchedDate)
		go f.publish(*feed)
	}
}
