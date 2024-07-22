package main

import (
	"fmt"
	"time"
	"strings"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

 // { year: { month: calendar } }
var calendarCache = make(map[string]map[string]tgbotapi.InlineKeyboardMarkup)

var BackBtn = tgbotapi.NewInlineKeyboardButtonData("Back", "back")
var Footer = tgbotapi.NewInlineKeyboardRow(BackBtn)

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
					data = "none"
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
	return dateKeyboard
}

func GenHours(timeType string, minStartTime int) tgbotapi.InlineKeyboardMarkup {
	var keyboard tgbotapi.InlineKeyboardMarkup

	hours := []int{
		10, 11, 12, 13,
		14, 15, 16, 17,
		18, 19, 20, 21,
	}

	// Generate 3 x 4 grid
	for i := 0; i < 3; i++ {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := i * 4; j < i*4+4; j++ {
			timeBtnText := fmt.Sprintf("%d:00", hours[j])
			var timeBtnData string
			if hours[j] < minStartTime {
				timeBtnData = "none"
			} else {
				timeBtnData = fmt.Sprintf("time%s %d", timeType, hours[j])
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(timeBtnText, timeBtnData))
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
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

func GenCurrentMsg(currUserState UserState) (string, tgbotapi.InlineKeyboardMarkup) {
	var keyboard tgbotapi.InlineKeyboardMarkup
	var msgText string
	if currUserState.Date == "" {
		year := fmt.Sprint(time.Now().Year())
		month := time.Now().Month().String()
		msgText = month + " " + year
		keyboard = GenerateMonthlyCalendar(time.Now())
	} else if currUserState.TimeStart == "" {
		msgText = "Pick a starting time"
		keyboard = GenHours("start", 10)
	} else if currUserState.TimeEnd == "" {
		minTime, err := strconv.Atoi(currUserState.TimeStart)
		if err != nil {
			fmt.Printf("Error converting str to int: %v\n", err)
		}
		msgText = fmt.Sprintf("Now pick an ending time later than %s:00", currUserState.TimeStart)
		keyboard = GenHours("end", minTime)
	} else if currUserState.Confirmation == "" {
		msgText = "Your selection"
		day := strings.Fields(currUserState.Date)[2]
		month := strings.Fields(currUserState.Date)[1]
		year := strings.Fields(currUserState.Date)[0]
		dateBtn := tgbotapi.NewInlineKeyboardButtonData(day + " " + month + " " + year, "none")
		intervalBtnText := fmt.Sprintf("%s:00...%s:00", currUserState.TimeStart, currUserState.TimeEnd)
		intervalBtn := tgbotapi.NewInlineKeyboardButtonData(intervalBtnText, "none")
		dateAndTimeRow := tgbotapi.NewInlineKeyboardRow(dateBtn, intervalBtn)
		confirmButtonRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Confirm", "confirm"))
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, dateAndTimeRow)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, confirmButtonRow)
	} else if currUserState.Confirmation == "confirmed" {
		msgText = "Waiting for other participants..."	
		hourglassBtn := tgbotapi.NewInlineKeyboardButtonData("⏳🤖⏳", "none")
		hourglassRow := tgbotapi.NewInlineKeyboardRow(hourglassBtn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, hourglassRow)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, Footer)
	return msgText, keyboard
}
