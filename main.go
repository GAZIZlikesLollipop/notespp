package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

type Note struct {
	Id      int64
	Name    string
	Content string
}

func createNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Ошибка чтения тела запроса: ", err)
		w.Write([]byte("Ошибка чтения тела запроса"))
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &note); err != nil {
		fmt.Println("Ошибка декодирвания тела запроса: ", err)
		w.Write([]byte("Ошибка декодирвания тела запроса"))
		return
	}
	if _, err := db.Exec(context.Background(), "INSERT INTO notes (name,content) VALUES ($1, $2)", note.Name, note.Content); err != nil {
		fmt.Println("Ошибка создания заметки: ", err)
		w.Write([]byte("Ошибка создания заметки"))
		return
	}
	w.Write([]byte("Успешное создание заметки"))
}

func getNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	userId := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	row := db.QueryRow(context.Background(), "SELECT * FROM notes WHERE id=$1", userId)
	if err := row.Scan(&note.Id, &note.Name, &note.Content); err != nil {
		fmt.Println("Ошибка чтения заметки: ", err)
		w.Write([]byte("Ошибка чтения заметки"))
		return
	}
	data, err := json.Marshal(note)
	if err != nil {
		fmt.Println("Ошибка сериализации заметки: ", err)
		w.Write([]byte("Ошибка сериализации заметки"))
		return
	}
	w.Write(data)
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	var notes []Note
	rows, err := db.Query(context.Background(), "SELECT * FROM notes")
	if err != nil {
		fmt.Println("Ошибка чтения рядов: ", err)
		w.Write([]byte("Ошибка чтения рядов"))
		return
	}

	defer rows.Close()

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.Id, &note.Name, &note.Content); err != nil {
			fmt.Println("Ошибка чтения рядa: ", err)
			w.Write([]byte("Ошибка чтения рядa"))
			return
		}
		notes = append(notes, note)
	}

	result, err := json.Marshal(notes)
	if err != nil {
		fmt.Println("Ошибка преобразования данных: ", err)
		w.Write([]byte("Ошибка преобразования данных"))
		return
	}

	w.Write([]byte(result))
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if _, err := db.Exec(context.Background(), "DELETE FROM notes WHERE id = $1", userId); err != nil {
		fmt.Println("Ошибка удаления записи: ", err)
		w.Write([]byte("Ошибка удаления записи"))
		return
	}
	w.Write([]byte("Успешное удаление заметки"))
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	var reqStr string
	var reqData []interface{}
	userId := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Ошибка обнволвение записи: ", err)
		w.Write([]byte("Ошибка обнволвение записи"))
		return
	}
	if err := json.Unmarshal(body, &note); err != nil {
		fmt.Println("Ошибка обнволвение записи: ", err)
		w.Write([]byte("Ошибка обнволвение записи"))
		return
	}
	if note.Name != "" {
		reqStr += "name = $2,"
		reqData = append(reqData, note.Name)
	}
	if note.Content != "" {
		if note.Name != "" {
			reqStr += " content = $3"
		} else {
			reqStr += " content = $2"
		}
		reqData = append(reqData, note.Content)
	}
	if len(reqData) < 1 {
		w.Write([]byte("Ничего не обнвиолось"))
		return
	}
	if len(reqData) > 1 {
		if _, err := db.Exec(context.Background(), fmt.Sprintf("UPDATE notes SET %s WHERE id = $1", reqStr), userId, reqData[0], reqData[1]); err != nil {
			fmt.Println("Ошибка обнволвение записи: ", err)
			w.Write([]byte("Ошибка обнволвение записи"))
			return
		}
	} else {
		if _, err := db.Exec(context.Background(), fmt.Sprintf("UPDATE notes SET %s WHERE id = $1", reqStr), userId, reqData[0]); err != nil {
			fmt.Println("Ошибка обнволвение записи: ", err)
			w.Write([]byte("Ошибка обнволвение записи"))
			return
		}
	}
	w.Write([]byte("Успешное обновление заметки"))
}

func main() {
	url := "postgres://postgres:12345678@localhost:5432/notesdb"
	var err error
	db, err = pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatalln("Ошибка подключения к базе данных: ", err)
	}
	_, err = db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS notes (id SERIAL PRIMARY KEY,name TEXT NOT NULL,content TEXT )")
	if err != nil {
		log.Fatalln("Ошибка при созании таблицы: ", err)
	}
	http.HandleFunc("/create", createNote)
	http.HandleFunc("/notes", getNotes)
	http.HandleFunc("/notes/", getNote)
	http.HandleFunc("/delete/", deleteNote)
	http.HandleFunc("/update/", updateNote)
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
