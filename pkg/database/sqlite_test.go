package database

import (
	"testing"
)

func TestSaveNote(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}
	defer db.conn.Close()

	topic := "Biology"
	content := "Mitochondria is the powerhouse of the cell."
	source := "textbook"

	err = db.SaveNote(topic, content, source)
	if err != nil {
		t.Errorf("SaveNote failed: %v", err)
	}

	// Verify using direct query
	var savedTopic, savedContent, savedSource string
	err = db.conn.QueryRow("SELECT topic, content, source FROM notes WHERE topic = ?", topic).Scan(&savedTopic, &savedContent, &savedSource)
	if err != nil {
		t.Fatalf("failed to query saved note: %v", err)
	}

	if savedTopic != topic {
		t.Errorf("expected topic %s, got %s", topic, savedTopic)
	}
	if savedContent != content {
		t.Errorf("expected content %s, got %s", content, savedContent)
	}
	if savedSource != source {
		t.Errorf("expected source %s, got %s", source, savedSource)
	}

	// Verify using GetNotesForTopic
	notes, err := db.GetNotesForTopic(topic)
	if err != nil {
		t.Errorf("GetNotesForTopic failed: %v", err)
	}
	found := false
	for _, n := range notes {
		if n == content {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetNotesForTopic did not return the saved content")
	}
}
