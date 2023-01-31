package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jgraeger/bverfgbot/internal/bverfg"
	"github.com/jgraeger/bverfgbot/internal/feed"
	"github.com/jgraeger/bverfgbot/internal/telegram"
)

const (
	defaultPort = "8000"

	decisionFeedURL = "https://www.bundesverfassungsgericht.de/SiteGlobals/Functions/RSSFeed/DE/Entscheidungen/RSSEntscheidungen.xml"
	pressFeedURL    = "https://www.bundesverfassungsgericht.de/SiteGlobals/Functions/RSSFeed/DE/Pressemitteilungen/RSSPressemitteilungen.xml"
)

var (
	debugFlag bool
)

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "enable debug mode")
}

type serveCfg struct {
	Addr     string
	BotToken string
	DSN      string
}

func serve(ctx context.Context, cfg serveCfg) error {
	// Start bot API
	bot, err := telegram.NewBot(ctx, cfg.BotToken, cfg.DSN)
	if err != nil {
		log.Fatalln("error creating telegram bot", err)
	}
	bot.DoNothing()

	testFeed := feed.NewFeed(ctx, decisionFeedURL)
	testFeed.SetRefreshInterval(5 * time.Second)
	testFeed.SetTranslator(bverfg.NewFeedTranslator())

	feedCh := testFeed.Subscribe()

	fmt.Println("Starting...")
	<-feedCh
	fmt.Println("Started...")
	for {
		select {
		case feed := <-feedCh:
			log.Println("new feed item received...")
			item := feed.Items[0]
			if item != nil {
				log.Println("notify bot users about:", item)
				if err := bot.NotifyDecision(item); err != nil {
					log.Println("Error sending decision notification:", err)
				}
			}
		case <-ctx.Done():
			log.Println("server received shutdown signal")
			return nil
		}
	}
}

func main() {
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("bot token not set")
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("db dsn not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	serveCfg := serveCfg{
		Addr:     fmt.Sprintf(":%s", port),
		BotToken: token,
		DSN:      dsn,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := <-shutdownCh
		log.Printf("received os interrupt:%+v\n", sig)
		cancel()
	}()

	if err := serve(ctx, serveCfg); err != nil {
		log.Fatalf("failed to serve: %+v", err)
	}
}
