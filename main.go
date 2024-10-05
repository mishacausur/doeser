package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID      int64  `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment,omitempty"`
	Repeat  string `db:"repeat" json:"repeat,omitempty"`
}

func createTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Decoding JSON Error", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(task.Title) == "" {
		http.Error(w, "Task name cannot be empty", http.StatusBadRequest)
		return
	}
	now := time.Now().Format("20060102")

	if task.Date == "" {
		task.Date = now
	} else {
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, "Некорректный формат даты", http.StatusBadRequest)
			return
		}

		if parsedDate.Format("20060102") < now {
			if strings.TrimSpace(task.Repeat) == "" {
				task.Date = now
			}
		}
	}
	id, err := createTaskInDB(db, task)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	fmt.Println(id)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(id)
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
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		createTask(w, r, db)
	})
	http.HandleFunc("/api/nextdate", nextDateHandler)

	error := http.ListenAndServe(":"+port, nil)

	if error != nil {
		panic(error)
	}
}
