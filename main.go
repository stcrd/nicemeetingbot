package main

/*
Assumptions:
1 - Only one chat room for now
2 - Users select only one date
3 - Users select only one start time and only one end time
*/

/*
TODOs:
1 - Fix field types in UserState
2 - Save the generated monthly calendar in cache
3 - Fix reset command so that it does not generate a new message, but rather replaces the existing one
4 - Summary should contain day of week in addition to the date
5 - Implement meeting time calculation
6 - Implement DB interaction
7 - Dockerize
8 - Deploy
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var bot *tgbotapi.BotAPI
var dateKeyboard tgbotapi.InlineKeyboardMarkup
const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

type Event struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type UserState struct {
	Name string // initial, dateSelected, startSelected, endSelected, confirmed, meetingSet
	Date time.Time
	TimeStart time.Time
	TimeEnd time.Time
	Confirmation bool
}

type ChatState struct {
	UserStates map[string]UserState // map of userNames to UserStates
}

var State = make(map[int64]ChatState) // chatId > ChatState

func callbackHandler(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.From.ID
	userName := update.CallbackQuery.From.UserName
	msgID := update.CallbackQuery.Message.MessageID
	var msg tgbotapi.MessageConfig
	msg.ChatID = chatID
	var event Event
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		log.Printf("Unmarshal error: %v\n", err)
		return
	}
	oldState := State[chatID].UserStates[userName]
	userState := State[chatID].UserStates[userName]

	// fix errors below, unmarshall etc.
	switch event.Name {
	case "none": // handling this case to make buttons inactive
	case "date":
		dateReceived, err := time.Parse(dateLayout, event.Data)
		if err != nil {
			log.Printf("Error parsing date: %v\n", err)
		}
		userState.Date = dateReceived
		userState.Name = "dateSelected"
	case "timestart":
		timeStart, err := time.Parse(timeLayout, event.Data)
		if err != nil {
			log.Printf("Error parsing timestart: %v\n", err)
		}
		userState.TimeStart = timeStart
		userState.Name = "startSelected"
	case "timeend":
		timeEnd, err := time.Parse(timeLayout, event.Data)
		if err != nil {
			log.Printf("Error parsing timeend: %v\n", err)
		}
		userState.TimeEnd = timeEnd
		userState.Name = "endSelected"
	case "confirm":
		userState.Confirmation = true
		userState.Name = "confirmed"
	case "back":
		switch userState.Name {
		case "confirmed":
			userState.Confirmation = false
			userState.Name = "endSelected"
		case "endSelected":
			userState.TimeEnd = time.Time{}
			userState.Name = "startSelected"
		case "startSelected":
			userState.TimeStart = time.Time{}
			userState.Name = "dateSelected"
		case "dateSelected":
			userState.Date = time.Time{}
			userState.Name = "initial"
		case "initial": // nothing happens
		}
	default:
		log.Printf("Unknown command from user %s in chat %d\n", userName, chatID)
	}
	if userState != oldState {
		State[chatID].UserStates[userName] = userState
		newText, newKeyboard :=  GenCurrentMsg(State[chatID].UserStates[userName])
		newMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, newText, newKeyboard)
		sendMessage(newMsg)
	}
	fmt.Printf("%+v\n", State)
}

func commandHandler(update tgbotapi.Update) {
	command := update.Message.Command()
	userName := update.Message.From.UserName
	chatID := update.Message.Chat.ID
	// msgID := update.Message.MessageID
	var msg tgbotapi.MessageConfig
	msg.ChatID = chatID

	switch  {
	case command == "start" || command == "reset":
		userStates := make(map[string]UserState)
		State[chatID] = ChatState{UserStates: userStates}
		State[chatID].UserStates[userName] = UserState{} // initiate the userName key in the map
		newText, newKeyboard := GenCurrentMsg(State[chatID].UserStates[userName])
		msg.Text = newText
		msg.ReplyMarkup = newKeyboard
		sendMessage(msg)
	default:
		sendMessage(tgbotapi.NewMessage(chatID, "Unknown command"))
	}
}

func sendMessage(msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send message error: %v", err)
	}
}

func main() {
	// load .env and get the bot token
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	tgToken := os.Getenv("TG_TOKEN")

	// initialize the bot
	bot, err = tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Printf("Error initializing the bot: %v\n", err)
	}

	// create updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Failed to start listening for updates %v", err)
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
