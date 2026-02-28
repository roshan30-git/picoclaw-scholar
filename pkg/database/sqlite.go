package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	conn.SetMaxOpenConns(1)

	tables := []string{
		`CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic TEXT NOT NULL,
			content TEXT NOT NULL,
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS quiz_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic TEXT NOT NULL,
			score INTEGER,
			total INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS embeddings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			note_id INTEGER,
			vector_json TEXT,
			FOREIGN KEY(note_id) REFERENCES notes(id)
		)`,
	}

	for _, t := range tables {
		if _, err := conn.Exec(t); err != nil {
			return nil, fmt.Errorf("create table: %w", err)
		}
	}

	return &DB{conn: conn}, nil
}

func (db *DB) SaveNote(topic, content, source string) error {
	_, err := db.conn.Exec(`INSERT INTO notes (topic, content, source) VALUES (?, ?, ?)`, topic, content, source)
	return err
}

func (db *DB) SaveEmbedding(noteID int64, vector []float64) error {
	data, _ := json.Marshal(vector)
	_, err := db.conn.Exec(`INSERT INTO embeddings (note_id, vector_json) VALUES (?, ?)`, noteID, string(data))
	return err
}

func (db *DB) GetNotesForTopic(topic string) ([]string, error) {
	query := "%" + topic + "%"
	rows, err := db.conn.Query(`SELECT content FROM notes WHERE topic LIKE ? OR content LIKE ? LIMIT 10`, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		results = append(results, c)
	}
	return results, nil
}

func (db *DB) QueryContext(topic string) (string, error) {
	notes, err := db.GetNotesForTopic(topic)
	if err != nil {
		return "", err
	}
	return strings.Join(notes, "\n---\n"), nil
}

func (db *DB) SaveQuizScore(topic string, score, total int) error {
	_, err := db.conn.Exec(`INSERT INTO quiz_history (topic, score, total) VALUES (?, ?, ?)`, topic, score, total)
	return err
}
