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

// generate a slice of days padded on both sides
// with 0 values for better visibility
func genMonthDays(t time.Time) []int {
	month := t.Month()
	year := t.Year()
	firstDayOfMonth := t.AddDate(0, 0, -t.Day()+1)
	lastDayOfMonth := t.AddDate(0, 1, -t.Day())
	firstDayWeekday := firstDayOfMonth.Weekday() // Sun = 0, Mon = 1, etc
	lastDayWeekday := lastDayOfMonth.Weekday()

	res := make([]int, 0, 35) // max 5 * 7 = 35 cells
	constantDays := []int{
		1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14,
		15, 16, 17, 18, 19, 20, 21,
		22, 23, 24, 25, 26, 27, 28,
	}

	for i := 1; i < int(firstDayWeekday); i++ {
		res = append(res, 0)
	}
	res = append(res, constantDays...)

	switch month {
	case time.April, time.June, time.September, time.November:
		res = append(res, []int{29, 30}...)
	case time.February:
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			// leap year
			res = append(res, 29)
		}
	default:
		res = append(res, []int{29, 30, 31}...)
	}

	for j := int(lastDayWeekday); j < 7; j++ {
		res = append(res, 0)
	}
	return res
}

func generateMonthlyCalendar(t time.Time) tgbotapi.InlineKeyboardMarkup {
	currDayOfMonth := t.Day()
	var keyboard tgbotapi.InlineKeyboardMarkup
	cells := genMonthDays(t)
	dayIndex := 0
	var text string
	var data string
	rowCount := 5
	columnCount := 7

	firstRow := generateWeekdayNames()
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, firstRow)

	for i := 0; i < rowCount; i++ {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := 1; j <= columnCount && dayIndex < len(cells); j++ {
			if cells[dayIndex] != 0 {
				text = fmt.Sprint(cells[dayIndex])
				if cells[dayIndex] >= currDayOfMonth {
					data = "date " + t.Month().String() + " " + fmt.Sprint(cells[dayIndex])
				} else {
					data = "past"
				}
			} else {
				text = " "
				data = "none"
			}
			btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
			row = append(row, btn)
			dayIndex++
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	return keyboard
}

func generateWeekdayNames() []tgbotapi.InlineKeyboardButton {
	weekdays := []tgbotapi.InlineKeyboardButton{}
	dayNames := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	for i := range dayNames {
		weekdays = append(weekdays, tgbotapi.NewInlineKeyboardButtonData(dayNames[i], "none"))
	}
	return weekdays
}

func callbackHandler(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatId := update.CallbackQuery.From.ID
	var text string

	switch {
	case strings.Fields(data)[0] == "none":
	case strings.Fields(data)[0] == "past":
		text = "Choose a valid date starting from today"
	case strings.Fields(data)[0] == "date":
		month := strings.Fields(data)[1]
		day := strings.Fields(data)[2]
		text = fmt.Sprintf("You have chosen: %s %s", month, day)
	default:
		text = "Unknown command"
	}
	msg := tgbotapi.NewMessage(chatId, text)
	sendMessage(msg)
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
		msg.Text = "Choose a date"
		msg.ReplyMarkup = generateMonthlyCalendar(time.Now())
		msg.ParseMode = "Markdown"
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
