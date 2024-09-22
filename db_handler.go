package main

import (
	"database/sql"
	"fmt"
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

func createTaskInDB(db *sql.DB, task Task) (string, error) {
	valid, err := isDateValid(task.Date)

	if !valid {
		return "", fmt.Errorf("Дата задачи должна быть равна или больше текущей даты.")
	}
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	fmt.Println(task.Date, task.Title, task.Comment, task.Repeat)
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", fmt.Errorf("Ошибка при добавлении задачи в базу данных: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("Ошибка при получении ID задачи: %v", err)
	}
	return fmt.Sprintf("%d", id), nil
}
