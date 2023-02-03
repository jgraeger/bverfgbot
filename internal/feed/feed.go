package feed

import (
	"bytes"
	"context"
	"crypto/sha1"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"go.uber.org/atomic"
)

const (
	fetchTimeout           = 10 * time.Second
	defaultRefreshInterval = 5 * time.Second

	rssTimestampFormat = "Mon, 2 Jan 2006 15:04:05 -0700"
)

type Feed struct {
	mu     sync.RWMutex
	closed bool
	ctx    context.Context

	url             string
	http            http.Client
	refreshInterval atomic.Duration
	lastChecksum    []byte // SHA1 hash of the last feed processed to detect changes
	parser          *gofeed.Parser
	subscriptions   []chan gofeed.Feed
}

func NewFeed(ctx context.Context, url string) *Feed {
	feed := &Feed{
		ctx:             ctx,
		url:             url,
		refreshInterval: *atomic.NewDuration(defaultRefreshInterval),
		lastChecksum:    make([]byte, sha1.Size),
		parser:          gofeed.NewParser(),
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

	req, err := http.NewRequest("GET", f.url, nil)
	if err != nil {
		log.Printf("error creating request: %v", err)
		return
	}

	res, err := f.http.Do(req.WithContext(ctx))
	if err != nil {
		log.Printf("error fetching feed: %v", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("unexpected status code while fetching feed %d", res.StatusCode)
		return
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading feed body: %v", err)
		return
	}

	hash := sha1.Sum(content)
	if bytes.Equal(f.lastChecksum, hash[:]) {
		// Feed are the same as last time
		return
	}
	f.lastChecksum = hash[:]

	feed, err := f.parser.Parse(bytes.NewReader(content))
	if err != nil {
		log.Printf("error parsing feed %v: %v\n", f.url, err)
		return
	}

	go f.publish(*feed)
}
