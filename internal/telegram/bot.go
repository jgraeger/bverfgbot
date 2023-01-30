package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/mmcdole/gofeed"
)

const (
	botTimeout = 30
)

type Bot struct {
	ctx context.Context

	api *tgbotapi.BotAPI
	db  *pgx.Conn
}

func NewBot(ctx context.Context, token string, postgresDSN string) (*Bot, error) {
	dbConn, err := pgx.Connect(ctx, postgresDSN)
	if err != nil {
		return nil, err
	}

	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Println("telegram bot authorized on account:", botApi.Self.UserName)

	bot := &Bot{
		ctx: ctx,
		api: botApi,
		db:  dbConn,
	}

	go bot.mainLoop()

	return bot, nil
}

func (b *Bot) mainLoop() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = botTimeout

	updateChan := b.api.GetUpdatesChan(updateConfig)

	eightAm, err := time.Parse(time.RFC1123Z, "Tue, 31 Jan 2023 07:00:00 +0000")
	if err != nil {
		panic(err)
	}
	fmt.Println(time.Until(eightAm))
	timer := time.NewTimer(time.Until(eightAm))

	for {
		select {
		case u := <-updateChan:
			if u.Message != nil {
				b.handleMessage(*u.Message)
			} else if u.ChatMember != nil {
				b.handleChatMember(*u.ChatMember)
			}
		case <-timer.C:
			b.SendToAll(muellerMsg)
			timer.Stop()
		case <-b.ctx.Done():
			log.Printf("shutdown telegram loop")
			b.shutdown()
			return
		}
	}
}

func (b *Bot) handleMessage(msg tgbotapi.Message) {
	_, err := b.db.Exec(b.ctx, storeChatQuery, msg.Chat.ID, msg.From.FirstName, msg.From.LastName)
	if err != nil {
		fmt.Println("error inserting chat into db", err)
		return
	}

	var responseText string

	switch msg.Text {
	case "/start":
		responseText, err = getWelcomeMessage(MessageConfig{FirstName: msg.From.FirstName})
		if err != nil {
			log.Println("template error:", err)
			return
		}
	default:
		return
	}

	response := tgbotapi.NewMessage(msg.Chat.ID, responseText)
	response.ReplyToMessageID = msg.MessageID

	if _, err := b.api.Send(response); err != nil {
		log.Println("error sending message:", err)
	}
}

func (b *Bot) handleChatMember(update tgbotapi.ChatMemberUpdated) {
	_, err := b.db.Exec(b.ctx, storeChatQuery, update.Chat.ID, update.From.FirstName, update.From.LastName)
	if err != nil {
		fmt.Println("error inserting chat into db", err)
		return
	}

	responseText, err := getWelcomeMessage(MessageConfig{FirstName: update.From.FirstName})
	if err != nil {
		log.Println("template error:", err)
		return
	}

	response := tgbotapi.NewMessage(update.Chat.ID, responseText)
	if _, err := b.api.Send(response); err != nil {
		log.Println("error sending message:", err)
	}
}

func (b *Bot) DoNothing() {}

func (b *Bot) NotifyDecision(item *gofeed.Item) error {
	msgString, err := buildDecisionMessage(item)
	if err != nil {
		return err
	}

	return b.SendToAll(msgString)
}

func (b *Bot) SendToAll(msg string) error {
	rows, err := b.db.Query(b.ctx, getAllQuery)
	if err != nil {
		return fmt.Errorf("error sending to all users: %w", err)
	}

	var chats []int64
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}
		chats = append(chats, id)
	}

	sent := 0
	for _, chatId := range chats {
		tgMsg := tgbotapi.NewMessage(chatId, msg)
		tgMsg.ParseMode = tgbotapi.ModeHTML

		if _, err := b.api.Send(tgMsg); err != nil {
			log.Println("error sending msg:", err)
		}

		sent++
		if sent%30 == 0 {
			<-time.After(1 * time.Second)
		}
	}

	return nil
}

func (b *Bot) shutdown() {
	b.db.Close(context.Background())
}
