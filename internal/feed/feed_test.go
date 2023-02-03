package feed_test

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/jgraeger/bverfgbot/internal/feed"
)

const (
	numTestFeedItems = 10
	autoRotateAfter  = 50
)

var (
	tpl *template.Template
)

func TestFeed(t *testing.T) {
	srv, err := NewFakeServer(false)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := feed.NewFeed(ctx, srv.server.URL)
	feed.SetRefreshInterval(2 * time.Millisecond)
	recvCh := feed.Subscribe()

	// Read initial feed
	<-recvCh

	// No new data if no refresh
	select {
	case <-recvCh:
		t.Fatal("Received feed data without feed change")
	case <-time.After(10 * time.Millisecond):
		break
	}

	srv.rotateFeed()
	select {
	case <-recvCh:
		break
	case <-time.After(5 * time.Millisecond):
		t.Fatal("didn't receive new feed data after update")
	}
}

func BenchmarkFeed(b *testing.B) {
	srv, err := NewFakeServer(true)
	if err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	feed := feed.NewFeed(ctx, srv.server.URL)
	feed.SetRefreshInterval(0)
	recvCh := feed.Subscribe()
	for i := 0; i < b.N; i++ {
		<-recvCh
	}
	cancel()
}

type feedItem struct {
	Title       string    `faker:"sentence"`
	Link        string    `faker:"url"`
	Description string    `faker:"paragraph"`
	Date        time.Time `faked:"timestamp"`
}

type tplData struct {
	Published time.Time
	Items     []feedItem
}

func init() {
	var err error
	tpl, err = template.New("feed").Parse(tplStr)
	if err != nil {
		log.Panicf("Error parsing feed test template: %v", err)
	}
}

func generateFakeFeed(numItems uint) (tplData, error) {
	tplData := tplData{
		Published: time.Now(),
		Items:     make([]feedItem, numItems),
	}

	for i := 0; i < int(numItems); i++ {
		if err := faker.FakeData(&tplData.Items[i]); err != nil {
			return tplData, fmt.Errorf("generating fake feed item: %w", err)
		}
	}
	faker.GetDateTimer()
	return tplData, nil
}

func WriteFakeFeed(w io.Writer, f tplData) error {
	return tpl.Execute(w, f)
}

type fakeServer struct {
	mu       sync.RWMutex
	server   *httptest.Server
	feedData tplData
}

func NewFakeServer(autorotate bool) (*fakeServer, error) {
	feedData, err := generateFakeFeed(numTestFeedItems)
	if err != nil {
		return nil, fmt.Errorf("create initial fake feed: %w", err)
	}

	f := &fakeServer{feedData: feedData}
	f.server = httptest.NewServer(func() http.HandlerFunc {
		reqs := 0
		return func(w http.ResponseWriter, r *http.Request) {
			if autorotate && reqs%autoRotateAfter == 0 {
				f.rotateFeed()
			}
			w.WriteHeader(http.StatusOK)
			f.mu.RLock()
			defer f.mu.RUnlock()
			WriteFakeFeed(w, f.feedData)
			reqs++
		}
	}())

	return f, nil
}

func (f *fakeServer) rotateFeed() error {
	newFeed, err := generateFakeFeed(numTestFeedItems)
	if err != nil {
		return fmt.Errorf("generate feed data for rotation: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.feedData = newFeed
	return nil
}

const tplStr = `
<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
<channel>
<atom:link href="https://www.bundesverfassungsgericht.de/SiteGlobals/Functions/RSSFeed/DE/Entscheidungen/RSSEntscheidungen.xml" rel="self" type="application/rss+xml"/>
<title>Bundesverfassungsgericht Newsfeed</title>
<link>http://www.bundesverfassungsgericht.de</link>
<description>Aktuelle Entscheidungen des Bundesverfassungsgerichts</description>
<language>de-de</language>
<copyright>Copyright by Bundesverfassungsgericht. Alle Rechte vorbehalten</copyright>
<category>Newspapers</category>
<generator>Government Site Builder</generator>
<docs>http://blogs.law.harvard.edu/tech/rss</docs>
<ttl>60</ttl>
<pubDate>{{ .Published.Format "Mon, 2 Jan 2006 15:04:05 -0700" }}</pubDate>
{{range .Items }}
<item>
<title>{{ .Title }}</title>
<guid>{{ .Link }}</guid>
<link>{{ .Link }}</link>
<description>{{ .Description }}</description>
<pubDate>{{ .Date.Format "Mon, 2 Jan 2006 15:04:05 -0700" }}</pubDate>
</item>
{{end}}
</channel>
</rss>
`
