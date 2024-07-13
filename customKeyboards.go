package main

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var calendarCache = make(map[string]map[string]tgbotapi.InlineKeyboardMarkup) // { year: { month: calendar } }

var BackBtn = tgbotapi.NewInlineKeyboardButtonData("Back", "Back")
var FwdBtn = tgbotapi.NewInlineKeyboardButtonData("Fwd", "Fwd")

func GenerateMonthlyCalendar(t time.Time) tgbotapi.InlineKeyboardMarkup {
	var dateKeyboard tgbotapi.InlineKeyboardMarkup
	var text, data string
	year := fmt.Sprint(t.Year())
	month := t.Month().String()

	// if already exists in the cache map, just return it
	if monthCalendar, exists := calendarCache[year][month]; exists {
		return monthCalendar
	}
	currDayOfMonth := t.Day()
	cells := genMonthDays(t)
	dayIndex := 0
	rowCount := 5
	columnCount := 7

	// fill first row with weekday names
	weekdays := []tgbotapi.InlineKeyboardButton{}
	dayNames := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	for i := range dayNames {
		weekdays = append(weekdays, tgbotapi.NewInlineKeyboardButtonData(dayNames[i], "none"))
	}
	dateKeyboard.InlineKeyboard = append(dateKeyboard.InlineKeyboard, weekdays)

	for i := 0; i < rowCount; i++ {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := 1; j <= columnCount && dayIndex < len(cells); j++ {
			if cells[dayIndex] != 0 {
				text = fmt.Sprint(cells[dayIndex])
				if cells[dayIndex] >= currDayOfMonth {
					data = fmt.Sprintf("date %s %s %s", year, month, text)
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
		dateKeyboard.InlineKeyboard = append(dateKeyboard.InlineKeyboard, row)
	}
	if _, exists := calendarCache[year]; !exists {
		calendarCache[year] = make(map[string]tgbotapi.InlineKeyboardMarkup)
	}
	calendarCache[year][month] = dateKeyboard
	dateKeyboard.InlineKeyboard = append(dateKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(BackBtn))
	return dateKeyboard
}

func UpdateMonthlyCalendar(oldDateKeyboard tgbotapi.InlineKeyboardMarkup, date string) tgbotapi.InlineKeyboardMarkup {
	lng := len(oldDateKeyboard.InlineKeyboard)
	for i := 0; i < lng; i++ {
		for j := 0; j < len(oldDateKeyboard.InlineKeyboard[i]); j++ {
			if oldDateKeyboard.InlineKeyboard[i][j].Text == date {
				prevText := oldDateKeyboard.InlineKeyboard[i][j].Text
				oldDateKeyboard.InlineKeyboard[i][j].Text = prevText + "📌"
				break
			}
		}
	}
	oldDateKeyboard.InlineKeyboard = append(oldDateKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(BackBtn))
	return oldDateKeyboard
}

func GenHours() tgbotapi.InlineKeyboardMarkup {
	var keyboard tgbotapi.InlineKeyboardMarkup

	hourStrs := []string{
		"10:00", "11:00", "12:00", "13:00",
		"14:00", "15:00", "16:00", "17:00",
		"18:00", "19:00", "20:00", "21:00",
	}
	// Generate 3 x 4 grid
	for i := 0; i < 3; i++ {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := i * 4; j < i*4+4; j++ {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(hourStrs[j], "time "+hourStrs[j]))
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(BackBtn))
	return keyboard
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

func GenInitialMenu() tgbotapi.InlineKeyboardMarkup {
	var keyboard tgbotapi.InlineKeyboardMarkup
	var firstRow []tgbotapi.InlineKeyboardButton
	var secondRow []tgbotapi.InlineKeyboardButton

	var chooseDateBtn = tgbotapi.NewInlineKeyboardButtonData("Choose date", "Choose date")

	firstRow = append(firstRow, chooseDateBtn)
	secondRow = append(secondRow, BackBtn, FwdBtn)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, firstRow)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, secondRow)
	return keyboard
}

func GenCurrentMsg(currUserState UserState) {

}
