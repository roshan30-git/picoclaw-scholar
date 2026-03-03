package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

	// Enable WAL and 5s timeout to prevent locking when concurrent reads/writes occur
	connectionString := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", connectionString)
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
		`CREATE TABLE IF NOT EXISTS deadlines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			due_date DATETIME NOT NULL,
			status TEXT DEFAULT 'pending',
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS learning_profile (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic TEXT NOT NULL,
			avg_score REAL DEFAULT 0,
			attempts INTEGER DEFAULT 0,
			pace_label TEXT DEFAULT 'medium'
		)`,
		`CREATE TABLE IF NOT EXISTS chat_summaries (
			chat_id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, t := range tables {
		if _, err := conn.Exec(t); err != nil {
			return nil, fmt.Errorf("create table: %w", err)
		}
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Conn() *sql.DB {
	return db.conn
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

func (db *DB) SaveChatSummary(chatID, content string) error {
	_, err := db.conn.Exec(`
		INSERT INTO chat_summaries (chat_id, content) 
		VALUES (?, ?) 
		ON CONFLICT(chat_id) DO UPDATE SET 
		content = excluded.content, 
		updated_at = CURRENT_TIMESTAMP`, chatID, content)
	return err
}

func (db *DB) GetLatestChatSummary(chatID string) string {
	var content string
	err := db.conn.QueryRow(`SELECT content FROM chat_summaries WHERE chat_id = ?`, chatID).Scan(&content)
	if err != nil {
		return ""
	}
	return content
}
