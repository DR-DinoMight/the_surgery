package main

import (
	"database/sql"
	"fmt"
	"net/http"

	helpers "github.com/DR-DinoMight/the_surgery/helpers"
	chat "github.com/DR-DinoMight/the_surgery/routes"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//check if sqllite3 db exists, if not create it
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		panic(err)
	}

	helpers.CtrlK()

	// Create table if doesn't exist or trucate existing table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message TEXT,
		user TEXT,
		colour INTEGER,
		timestamp TEXT,
		created_at TEXT DEFAULT (datetime('now')));`)
	if err != nil {
		fmt.Println("Error Creating Table")
		panic(err)
	}
	// Create table for actions
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action TEXT,
		type TEXT,
		actioned_at TEXT,
		created_at TEXT DEFAULT (datetime('now')));
	`)
	if err != nil {
		fmt.Println("Error Creating Table")
		panic(err)
	}

	http.HandleFunc("/webhook/chat", chat.ChatWebhook)
	http.HandleFunc("/messages", chat.MessageHandler)

	http.Handle("/", http.FileServer(http.Dir("static")))

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
