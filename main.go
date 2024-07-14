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
5 - Fix adding back and forward buttons, maybe make it like a footer

*/

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var bot *tgbotapi.BotAPI
var dateKeyboard tgbotapi.InlineKeyboardMarkup

type UserState struct {
	Date string
	TimeStart string
	TimeEnd string
	Confirmation string
}

type ChatState struct {
	UserStates map[string]UserState // map of userNames to UserStates
}

var State = make(map[int64]ChatState) // chatId > ChatState

func GetUserState(chatID int64, userName string) UserState {
	return State[chatID].UserStates[userName]
}

// generate updated keyboard based on the user state
func genKeyboard(chatID int64, msgID int, userName string) tgbotapi.InlineKeyboardMarkup {
	var msg tgbotapi.MessageConfig
	msg.ChatID = chatID
	// userState := GetUserState(chatID, userName)

	// msgID equals 0 only after start or reset command
	if msgID == 0 {
		return GenInitialMenu()
	}
	fmt.Println(chatID, msgID, userName)

	return tgbotapi.InlineKeyboardMarkup{}
}

func callbackHandler(update tgbotapi.Update) {
	fmt.Printf("%+v\n", State)
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.From.ID
	userName := update.CallbackQuery.From.UserName
	msgID := update.CallbackQuery.Message.MessageID
	var text string
	var msg tgbotapi.MessageConfig
	msg.ChatID = chatID

	switch {
	case strings.Fields(data)[0] == "none": // handling this case to make buttons inactive
	case strings.Fields(data)[0] == "past": // handling this case to make buttons inactive
	case strings.Fields(data)[0] == "date":
		year := strings.Fields(data)[1]
		month := strings.Fields(data)[2]
		day := strings.Fields(data)[3]
		State[chatID].UserStates[userName] = UserState{
			Date: fmt.Sprintf("%s %s %s", year, month, day),
		}
		msg.Text, msg.ReplyMarkup = GenCurrentMsg(State[chatID].UserStates[userName], msgID, chatID)
		sendMessage(msg)
	case strings.Fields(data)[0] == "timestart":
		timeStart := strings.Fields(data)[1]
		userState := State[chatID].UserStates[userName]
		userState.TimeStart = timeStart
		State[chatID].UserStates[userName] = userState
		msg.Text, msg.ReplyMarkup = GenCurrentMsg(State[chatID].UserStates[userName], msgID, chatID)
		sendMessage(msg)	
	case strings.Fields(data)[0] == "timeend":
		timeEnd := strings.Fields(data)[1]
		userState := State[chatID].UserStates[userName]
		userState.TimeEnd = timeEnd
		State[chatID].UserStates[userName] = userState
		msg.Text, msg.ReplyMarkup = GenCurrentMsg(State[chatID].UserStates[userName], msgID, chatID)
		sendMessage(msg)	
	case strings.Fields(data)[0] == "Back":
		fmt.Println("Back button pressed")
	case strings.Fields(data)[0] == "Fwd":
		fmt.Println("Forward button pressed")
	default:
		text = "Unknown command"
		msg := tgbotapi.NewMessage(chatID, text)
		sendMessage(msg)
	}

}

func commandHandler(update tgbotapi.Update) {
	command := update.Message.Command()
	userName := update.Message.From.UserName
	chatID := update.Message.Chat.ID
	msgID := update.Message.MessageID
	var msg tgbotapi.MessageConfig
	msg.ChatID = chatID

	switch  {
	case command == "start" || command == "reset":
		userStates := make(map[string]UserState)
		State[chatID] = ChatState{UserStates: userStates}
		State[chatID].UserStates[userName] = UserState{} // initiate the userName key in the map
		msg.Text, msg.ReplyMarkup = GenCurrentMsg(State[chatID].UserStates[userName], msgID, chatID)
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
