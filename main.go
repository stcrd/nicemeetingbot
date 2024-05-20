package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var bot *tgbotapi.BotAPI

type button struct {
	name string
	data string
}

func startMenu() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{
			name: "Hello",
			data: "hello",
		},
		{
			name: "Goodbye",
			data: "goodbye",
		},
	}

	buttons := make([][]tgbotapi.InlineKeyboardButton, len(states))
	for index, state := range states {
		buttons[index] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(state.name, state.data))
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tgToken := os.Getenv("TG_TOKEN")

	bot, err = tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to start listening for updates %v", err)
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			cbHandler(update)
		} else if update.Message.IsCommand() {
			cmdHandler(update)
		} else {
			log.Println("Unknown update")
		}
	}
}

func cbHandler(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatId := update.CallbackQuery.From.ID
	userName := update.CallbackQuery.From.UserName
	var text string
	switch data {
	case "hello":
		text = fmt.Sprintf("Hello, %v", userName)
	case "goodbye":
		text = fmt.Sprintf("Goodbye, %v", userName)
	default:
		text = "Unknown command"
	}
	msg := tgbotapi.NewMessage(chatId, text)
	sendMessage(msg)
}

func cmdHandler(update tgbotapi.Update) {
	command := update.Message.Command()
	switch command {
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select an action")
		msg.ReplyMarkup = startMenu()
		msg.ParseMode = "Markdown"
		sendMessage(msg)
	}
}

func sendMessage(msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Panicf("Send message error: %v", err)
	}
}
