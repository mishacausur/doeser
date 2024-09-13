package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
