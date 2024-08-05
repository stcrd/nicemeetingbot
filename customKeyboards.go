package main

import (
	"fmt"
	"time"
	"encoding/json"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

 // { year: { month: calendar } }
var calendarCache = make(map[int]map[time.Month]tgbotapi.InlineKeyboardMarkup)

var Footer = tgbotapi.NewInlineKeyboardRow(GenerateBackBtn())
var NoneEvent = `{"name":"none","data":"none"}`

func GenerateBackBtn() tgbotapi.InlineKeyboardButton {
	var backEvent Event
	backEvent.Name = "back"
	backEvent.Data = "back"
	backJsonData, err := json.Marshal(&backEvent)
	if err != nil {
		log.Printf("Error marshaling back button: %v", err)
	}
	backBtnData := string(backJsonData)
	return tgbotapi.NewInlineKeyboardButtonData("Back", backBtnData)
}

func GenerateMonthlyCalendar(t time.Time) tgbotapi.InlineKeyboardMarkup {
	var dateKeyboard tgbotapi.InlineKeyboardMarkup
	var text, data string
	year := t.Year()
	month := t.Month()

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
		weekdays = append(weekdays, tgbotapi.NewInlineKeyboardButtonData(dayNames[i], NoneEvent))
	}
	dateKeyboard.InlineKeyboard = append(dateKeyboard.InlineKeyboard, weekdays)

	var event Event
	event.Name = "date"
	for i := 0; i < rowCount; i++ {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := 1; j <= columnCount && dayIndex < len(cells); j++ {
			if cells[dayIndex] != 0 {
				text = fmt.Sprint(cells[dayIndex])
				if cells[dayIndex] >= currDayOfMonth {
					event.Data = fmt.Sprintf("%d-%02d-%02d", year, month, cells[dayIndex])
					jsonData, err := json.Marshal(&event)
					if err != nil {
						log.Printf("Error marshaling date event: %v", err)
					}
					data = string(jsonData)
				} else {
					data = NoneEvent
				}
			} else {
				text = " "
				data = NoneEvent
			}
			btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
			row = append(row, btn)
			dayIndex++
		}
		dateKeyboard.InlineKeyboard = append(dateKeyboard.InlineKeyboard, row)
	}
	if _, exists := calendarCache[year]; !exists {
		calendarCache[year] = make(map[time.Month]tgbotapi.InlineKeyboardMarkup)
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
	var event Event
	event.Name = fmt.Sprintf("time%s", timeType)
	for i := 0; i < 3; i++ {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := i * 4; j < i*4+4; j++ {
			timeBtnText := fmt.Sprintf("%d:00", hours[j])
			var timeBtnData string
			if hours[j] < minStartTime {
				timeBtnData = NoneEvent
			} else {
				event.Data = timeBtnText
				jsonData, err := json.Marshal(&event)
				if err != nil {
					log.Printf("Error marshaling time %s event: %v", timeType, err)
				}
				timeBtnData = string(jsonData) 
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
	if currUserState.Date.IsZero() {
		year := fmt.Sprint(time.Now().Year())
		month := time.Now().Month().String()
		msgText = month + " " + year
		keyboard = GenerateMonthlyCalendar(time.Now())
	} else if currUserState.TimeStart.IsZero() {
		msgText = "Pick a starting time"
		keyboard = GenHours("start", 10)
	} else if currUserState.TimeEnd.IsZero() {
		minTime := currUserState.TimeStart.Hour()
		msgText = fmt.Sprint("Now pick an ending time later than: ", currUserState.TimeStart.Format(timeLayout))
		keyboard = GenHours("end", minTime)
	} else if !currUserState.Confirmation {
		msgText = "Your selection"
		day := currUserState.Date.Day()
		month := currUserState.Date.Month().String()
		year := currUserState.Date.Year()
		dateStr := fmt.Sprintf("%d-%s-%d", day, month, year)
		dateBtn := tgbotapi.NewInlineKeyboardButtonData(dateStr, NoneEvent)
		intervalBtnText := fmt.Sprintf("%s...%s", currUserState.TimeStart.Format(timeLayout), currUserState.TimeEnd.Format(timeLayout))
		intervalBtn := tgbotapi.NewInlineKeyboardButtonData(intervalBtnText, NoneEvent)
		dateAndTimeRow := tgbotapi.NewInlineKeyboardRow(dateBtn, intervalBtn)

		var event Event
		event.Name = "confirm"
		event.Data = "confirm"
		jsonData, err := json.Marshal(&event)
		if err != nil {
			log.Printf("Error marshaling confirm: %v", err)
		}
		confirmBtnData := string(jsonData)
		confirmButtonRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Confirm", confirmBtnData ))
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, dateAndTimeRow)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, confirmButtonRow)
	} else if currUserState.Confirmation {
		msgText = "Waiting for other participants..."	
		hourglassBtn := tgbotapi.NewInlineKeyboardButtonData("⏳🤖⏳", NoneEvent)
		hourglassRow := tgbotapi.NewInlineKeyboardRow(hourglassBtn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, hourglassRow)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, Footer)
	return msgText, keyboard
}
