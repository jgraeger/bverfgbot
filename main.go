package main

import (
	"flag"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	decisionFeedURL = "http://www.bundesverfassungsgericht.de/rss/entscheidungen/"
	pressFeedURL    = "https://www.bundesverfassungsgericht.de/SiteGlobals/Functions/RSSFeed/DE/Pressemitteilungen/RSSPressemitteilungen.xml"
)

var (
	debugFlag bool
)

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "enable debug mode")
}

func main() {
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("bot token not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("creating api client: %w", err)
	}

	if debugFlag {
		bot.Debug = true
		log.Printf("authorized on account: %s", bot.Self.UserName)
	}

}
