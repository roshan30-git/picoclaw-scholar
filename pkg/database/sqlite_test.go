package database

import (
	"testing"
)

func TestSaveNote(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Conn().Close()

	topic := "Test Topic"
	content := "Test Content"
	source := "Test Source"

	err = db.SaveNote(topic, content, source)
	if err != nil {
		t.Fatalf("SaveNote failed: %v", err)
	}

	var gotTopic, gotContent, gotSource string
	err = db.Conn().QueryRow("SELECT topic, content, source FROM notes").Scan(&gotTopic, &gotContent, &gotSource)
	if err != nil {
		t.Fatalf("Failed to query notes table: %v", err)
	}

	if gotTopic != topic {
		t.Errorf("got topic %q, want %q", gotTopic, topic)
	}
	if gotContent != content {
		t.Errorf("got content %q, want %q", gotContent, content)
	}
	if gotSource != source {
		t.Errorf("got source %q, want %q", gotSource, source)
	}
}
