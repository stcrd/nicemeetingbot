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

type button struct {
	name string
	data string
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

	// bot.Debug = true

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

func genMonthDays(t time.Time) []int {
	month := t.Month()
	year := t.Year()
	firstDayOfMonth := t.AddDate(0, 0, -t.Day()+1)
	lastDayOfMonth := t.AddDate(0, 1, -t.Day())
	firstDayWeekday := firstDayOfMonth.Weekday() // Sun = 0, Mon = 1, etc
	lastDayWeekday := lastDayOfMonth.Weekday()

	res := []int{}
	constantDays := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28}

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
					data = "date " + fmt.Sprint(cells[dayIndex])
				} else {
					data = "inactive"
				}
			} else {
				text = " "
				data = "inactive"
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
		weekdays = append(weekdays, tgbotapi.NewInlineKeyboardButtonData(dayNames[i], "inactive"))
	}
	return weekdays
}

func cbHandler(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatId := update.CallbackQuery.From.ID
	userName := update.CallbackQuery.From.UserName
	var text string
	switch {
	case data == "hello":
		text = fmt.Sprintf("Hello, %v", userName)
	case data == "goodbye":
		text = fmt.Sprintf("Goodbye, %v", userName)
	case data == "inactive":
		text = "Choose a valid date starting from today"
	case strings.Contains(data, "date"):
		month := time.Now().Month().String() // assuming current month is selected
		day := strings.Fields(data)[1]
		text = fmt.Sprintf("You have chosen: %s %s", month, day)
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
	case "choosedate":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose a date")
		now := time.Now()
		msg.ReplyMarkup = generateMonthlyCalendar(now)
		msg.ParseMode = "Markdown"
		sendMessage(msg)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
		sendMessage(msg)
	}
}

func sendMessage(msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Panicf("Send message error: %v", err)
	}
}
