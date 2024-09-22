package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func isDateValid(dateString string) (bool, error) {
	layout := "20060102"
	taskDate, err := time.Parse(layout, dateString)
	if err != nil {
		return false, fmt.Errorf("Некорректный формат даты: %v", err)
	}
	now := time.Now().Format(layout)
	currentDate, err := time.Parse(layout, now)
	if err != nil {
		return false, fmt.Errorf("Ошибка при определении текущей даты: %v", err)
	}
	return !taskDate.Before(currentDate), nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	layout := "20060102"
	taskDate, err := time.Parse(layout, date)

	if err != nil {
		return "", fmt.Errorf("Ошибка парсинга даты: %v", err)
	}
	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", nil
	}

	var nextDate time.Time

	switch {
	// Повторение ежегодно
	case repeat == "y":
		nextDate = taskDate
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
			if nextDate.Year() > now.Year()+100 {
				return "", fmt.Errorf("превышен лимит повторений по годам")
			}
		}
	// Повторение каждые N дней
	case strings.HasPrefix(repeat, "d "):

		fields := strings.Fields(repeat)
		if len(fields) != 2 {
			return "", fmt.Errorf("некорректный формат правила 'd N'")
		}
		days, err := strconv.Atoi(fields[1])
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("некорректное число дней в правиле повторения")
		}
		nextDate = taskDate
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(taskDate.AddDate(5, 0, 0)) {
				return "", fmt.Errorf("превышен лимит повторений по дням")
			}
		}
	// Повторение по дням недели
	case strings.HasPrefix(repeat, "w "):
		fields := strings.Fields(repeat)
		if len(fields) != 2 {
			return "", fmt.Errorf("некорректный формат правила 'w'")
		}
		weekdaysStr := strings.Split(fields[1], ",")
		weekdays := make([]time.Weekday, 0, len(weekdaysStr))
		for _, wdStr := range weekdaysStr {
			wdStr = strings.TrimSpace(wdStr)
			wdNum, err := strconv.Atoi(wdStr)
			if err != nil || wdNum < 1 || wdNum > 7 {
				return "", fmt.Errorf("некорректный номер дня недели: %s", wdStr)
			}
			var wd time.Weekday
			if wdNum == 7 {
				wd = time.Sunday
			} else {
				wd = time.Weekday(wdNum)
			}
			weekdays = append(weekdays, wd)
		}
		nextDate = now.AddDate(0, 0, 1)
		found := false
		for i := 0; i < 14; i++ {
			for _, wd := range weekdays {
				if nextDate.Weekday() == wd {
					if nextDate.After(now) {
						found = true
						break
					}
				}
			}
			if found {
				break
			}
			nextDate = nextDate.AddDate(0, 0, 1)
		}
		if !found {
			return "", fmt.Errorf("не удалось найти подходящий день недели")
		}
	// Повторение по дням месяца
	case strings.HasPrefix(repeat, "m "):
		fields := strings.Fields(repeat)
		if len(fields) < 2 || len(fields) > 3 {
			return "", fmt.Errorf("некорректный формат правила 'm'")
		}

		daysStr := strings.Split(fields[1], ",")
		daysOfMonth := make([]int, 0, len(daysStr))
		for _, dayStr := range daysStr {
			dayStr = strings.TrimSpace(dayStr)
			day, err := strconv.Atoi(dayStr)
			if err != nil || (day == 0) || (day < -2) || day > 31 {
				return "", fmt.Errorf("некорректный день месяца: %s", dayStr)
			}
			daysOfMonth = append(daysOfMonth, day)
		}
		monthsOfYear := make([]time.Month, 0)
		if len(fields) == 3 {
			monthsStr := strings.Split(fields[2], ",")
			for _, monthStr := range monthsStr {
				monthStr = strings.TrimSpace(monthStr)
				monthNum, err := strconv.Atoi(monthStr)
				if err != nil || monthNum < 1 || monthNum > 12 {
					return "", fmt.Errorf("некорректный номер месяца: %s", monthStr)
				}
				monthsOfYear = append(monthsOfYear, time.Month(monthNum))
			}
		} else {
			for m := time.January; m <= time.December; m++ {
				monthsOfYear = append(monthsOfYear, m)
			}
		}
		nextDate = now.AddDate(0, 0, 1)
		for i := 0; i < 366; i++ {
			if !containsMonth(monthsOfYear, nextDate.Month()) {
				nextDate = nextDate.AddDate(0, 0, 1)
				continue
			}
			if containsDayOfMonth(daysOfMonth, nextDate) && nextDate.After(now) {
				break
			}
			nextDate = nextDate.AddDate(0, 0, 1)
		}
		if !nextDate.After(now) {
			return "", fmt.Errorf("не удалось найти подходящий день месяца")
		}
	default:
		return "", nil
	}

	return nextDate.Format(layout), nil
}

func containsMonth(months []time.Month, month time.Month) bool {
	for _, m := range months {
		if m == month {
			return true
		}
	}
	return false
}

func containsDayOfMonth(days []int, date time.Time) bool {
	day := date.Day()
	lastDay := lastDayOfMonth(date.Year(), date.Month())
	for _, d := range days {
		switch {
		case d > 0 && d == day:
			return true
		case d == -1 && day == lastDay:
			return true
		case d == -2 && day == lastDay-1:
			return true
		}
	}
	return false
}

func lastDayOfMonth(year int, month time.Month) int {
	nextMonth := month + 1
	if nextMonth > 12 {
		nextMonth = 1
		year++
	}
	firstOfNextMonth := time.Date(year, nextMonth, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}
