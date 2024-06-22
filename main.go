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
	ChatID    int64
	UserName  string
	Date      string
	BeginTime int
	EndTime   int
}

type CBData struct {
	Type  string `json:"type" validate:"oneof=date begin end none"`
	Value string `json:"value"` // date = YYYY-MM-DD, begin = HH, end = HH
}

var state = make(map[int64]map[string][]Record) // { chatID: []Record }

func main() {
	// load .env and get the bot token
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
	fmt.Println(data)
	chatId := update.CallbackQuery.From.ID
	userName := update.CallbackQuery.From.UserName
	msgId := update.CallbackQuery.Message.MessageID
	var text string

	switch {
	case strings.Fields(data)[0] == "none":
	case strings.Fields(data)[0] == "past":
	case data == "Choose date":
		text := "Choose a date"
		msg := tgbotapi.NewMessage(chatId, text)
		dateKeyboard := GenerateMonthlyCalendar(time.Now())
		msg.ReplyMarkup = dateKeyboard
		msg.ParseMode = "MarkdownV2"
		sendMessage(msg)
	case strings.Fields(data)[0] == "date":
		year := strings.Fields(data)[1]
		month := strings.Fields(data)[2]
		day := strings.Fields(data)[3]
		if _, exists := state[chatId]; !exists {
			state[chatId] = make(map[string][]Record)
		}

		if _, exists := state[chatId][userName]; !exists {
			state[chatId][userName] = make([]Record, 0, 20)
		}
		// TODO: first check if that date already exists in the state
		state[chatId][userName] = append(state[chatId][userName], Record{
			ChatID:   chatId,
			UserName: userName,
			Date:     fmt.Sprintf("%s %s %s", year, month, day),
		})
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
		if _, exists := state[chatID][userName]; !exists {
			msg.ReplyMarkup = GenInitialMenu()
		}
		msg.Text = "Hello, " + userName
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

// func sendAutoDeletingMsg(msg tgbotapi.MessageConfig, delay time.Duration) {
// 	sentMsg, err := bot.Send(msg)
// 	if err != nil {
// 		log.Panicf("Send message error: %v", err)
// 	}
// 	time.Sleep(delay * time.Second)
// 	deleteConfig := tgbotapi.NewDeleteMessage(sentMsg.Chat.ID, sentMsg.MessageID)

// 	if _, err := bot.Request(deleteConfig); err != nil {
// 		log.Println("Error deleting message:", err)
// 	}
// }
