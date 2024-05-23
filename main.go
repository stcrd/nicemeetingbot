package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var bot *tgbotapi.BotAPI
var dateKeyboard tgbotapi.InlineKeyboardMarkup

type Record struct {
	ChatID    string `validate:"required"`
	UserID    string `validate:"required"`
	Date      int    `validate:"required,gte=1,lte=31"`
	BeginTime int    `validate:"required,gte=10,lte=21"`
	EndTime   int    `validate:"required,gte=10,lte=21"`
}

type CBData struct {
	Type  string `json:"type" validate:"oneof=date begin end none"`
	Value string `json:"value"` // date = YYYY-MM-DD, begin = HH, end = HH
}

func main() {
	// load .env and get the token
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tgToken := os.Getenv("TG_TOKEN")

	// initialize the bot
	bot, err = tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic(err)
	}

	// create updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to start listening for updates %v", err)
	}

	// listen to updates from the channel
	for update := range updates {
		if update.CallbackQuery != nil {
			callbackHandler(update)
		} else if update.Message.IsCommand() {
			commandHandler(update)
		} else {
			log.Println("Unknown update")
		}
	}
}

func callbackHandler(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatId := update.CallbackQuery.From.ID
	msgId := update.CallbackQuery.Message.MessageID
	// userId := chatId // user ID and chat ID are the same in private chats
	var text string

	switch {
	case strings.Fields(data)[0] == "none":
	case strings.Fields(data)[0] == "past":
		text = "Choose a valid date starting from today"
		msg := tgbotapi.NewMessage(chatId, text)
		sendMessage(msg)
	case strings.Fields(data)[0] == "date":
		month := strings.Fields(data)[1]
		day := strings.Fields(data)[2]
		// TODO: implement toggling
		updatedCalendar := UpdateMonthlyCalendar(dateKeyboard, day)
		msg := tgbotapi.NewEditMessageReplyMarkup(chatId, msgId, updatedCalendar)
		sendMessage(msg)

		text = fmt.Sprintf("Choose time slots for: %s %s", month, day)
		msg2 := tgbotapi.NewMessage(chatId, text)
		msg2.ReplyMarkup = GenHours()
		sendMessage(msg2)
	default:
		text = "Unknown command"
		msg := tgbotapi.NewMessage(chatId, text)
		sendMessage(msg)
	}

}

func commandHandler(update tgbotapi.Update) {
	command := update.Message.Command()
	userName := update.Message.From.UserName
	chatID := update.Message.Chat.ID
	var msg tgbotapi.MessageConfig
	msg.ChatID = chatID

	switch command {
	case "start":
		msg.Text = "Hello, " + userName
	case "choosedate":
		msg.Text = "__Choose a date__"
		GenerateMonthlyCalendar(time.Now(), &dateKeyboard)
		msg.ReplyMarkup = dateKeyboard
		msg.ParseMode = "MarkdownV2"
	case "choosetime": // TODO
		msg.Text = "Not implemented yet"
	case "changedate": // TODO
		msg.Text = "Not implemented yet"
	case "changetime": // TODO
		msg.Text = "Not implemented yet"
	default:
		msg.Text = "Unknown command"
	}
	sendMessage(msg)
}

func sendMessage(msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Panicf("Send message error: %v", err)
	}
}
