package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

type Note struct {
	Id      int64
	Name    string
	Content string
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ishlab ketdi!")
	w.Write([]byte("HEllo bro"))
}

func main() {
	// url := "postgres://username:password@localhost:5432/database_name"
	url := "postgres://postgres:@localhost:5432/notesdb"
	var err error
	db, err = pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatalln("Ошибка подключения к базе данных: ", err)
	}
	_, err = db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS notes (id SERIAL PRIMARY KEY,name TEXT NOT NULL,content TEXT )")
	if err != nil {
		log.Fatalln("Ошибка при созании таблицы: ", err)
	}
	http.HandleFunc("/notes", getNotes)
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
