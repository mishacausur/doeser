package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func createDatabase(db *sql.DB) error {

	createTableSQL := `
    CREATE TABLE scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT CHECK (length(repeat) <= 128)
    );`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("Ошибка создания таблицы: %v", err)
	}

	createIndexSQL := `
    CREATE INDEX idx_scheduler_date ON scheduler (date);`

	_, err = db.Exec(createIndexSQL)
	if err != nil {
		return fmt.Errorf("Ошибка создания индекса: %v", err)
	}

	return nil
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

	if repeat == "y" {
		nextDate = taskDate
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
			if nextDate.Year() > now.Year()+100 {
				return "", fmt.Errorf("Некорректная дата -- превышение лимита лет")
			}
		}
	} else if strings.HasPrefix(repeat, "d ") {
		fields := strings.Fields(repeat)
		if len(fields) != 2 {
			return "", fmt.Errorf("Некорректный формат повторения")
		}
		numberStrings := fields[1]
		days, err := strconv.Atoi(numberStrings)
		if err != nil {
			return "", fmt.Errorf("Некорректный формат повторения")
		}
		if days <= 0 || days > 400 {
			return "", fmt.Errorf("Некорректный формат повторения")
		}
		nextDate = taskDate
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(taskDate.AddDate(5, 0, 0)) {
				return "", fmt.Errorf("Не удалось получить следующую дату")
			}
		}
	} else {
		return "", nil
	}
	return nextDate.Format(layout), nil
}

func main() {

	dbFile := os.Getenv("TODO_DBFILE")

	if dbFile == "" {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")

	}

	_, err := os.Stat(dbFile)
	var install bool

	if os.IsNotExist(err) {
		install = true
	} else if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if install {
		if err := createDatabase(db); err != nil {
			log.Fatal(err)
		}
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	indexPage := http.FileServer(http.Dir("./web"))

	http.Handle("/", indexPage)

	error := http.ListenAndServe(":"+port, nil)

	if error != nil {
		panic(error)
	}
}
