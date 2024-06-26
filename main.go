package main

/*
Assumptions:
1 - Only one chat room for now
2 - Users select only one date
3 - Users select only one start time and only one end time
*/

/*
TODOs:
1 - Replace keyboard when sending a reply
2 - Make two time keyboards, for start and end: add timeStart and timeEnd instead of time in the data?
3 - Deal with back button: maintain the state of the selection process?
4 - Last keyboard with 'Confirm' and 'Change' buttons?

*/

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

type UserState struct {
	Date string
	TimeStart string
	TimeEnd string
}

type ChatState struct {
	UserStates map[string]UserState // map of userNames to UserStates
}

var State = make(map[int64]ChatState) // chatId > ChatState

// generate updated keyboard based on the user state
// func updateKeyboard(chatId int64, msgId int) tgbotapi.InlineKeyboardMarkup {
// 		
// }

func callbackHandler(update tgbotapi.Update) {
	fmt.Printf("%v\n", State)
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
		State[chatId].UserStates[userName] = UserState{
			Date: fmt.Sprintf("%s %s %s", year, month, day),
		}

		// TODO: implement toggling
		updatedCalendar := UpdateMonthlyCalendar(dateKeyboard, day)
		fmt.Printf("updatedCalendar: %v\n", updatedCalendar.InlineKeyboard)
		msg := tgbotapi.NewEditMessageReplyMarkup(chatId, msgId, updatedCalendar)
		sendMessage(msg)

		text = fmt.Sprintf("Choose time slots for: %s %s", month, day)
		msg2 := tgbotapi.NewMessage(chatId, text)
		msg2.ReplyMarkup = GenHours()
		sendMessage(msg2)
	case strings.Fields(data)[0] == "time":
		fmt.Println("Time field pressed")
	case strings.Fields(data)[0] == "Back":
		fmt.Println("Back button pressed")
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

	if _, exists := State[chatID]; !exists {
		userStates := make(map[string]UserState)
		State[chatID] = ChatState{UserStates: userStates}
	}

	if _, exists := State[chatID].UserStates[userName]; !exists {
		State[chatID].UserStates[userName] = UserState{}
	}

	switch command {
	case "start":
		State[chatID].UserStates[userName] = UserState{} // initiate the userName key in the map
		msg.ReplyMarkup = GenInitialMenu()
		msg.Text = "Hello, " + userName // should this also reset the user state?
	case "reset":
		State[chatID].UserStates[userName] = UserState{}
		msg.ReplyMarkup = GenInitialMenu()
		msg.Text = "Hello, " + userName // should this also reset the user state?
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
