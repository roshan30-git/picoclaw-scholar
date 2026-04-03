package study

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/roshan30-git/picoclaw-scholar/pkg/database"
)

func BenchmarkPYQInserts(b *testing.B) {
	dbPath := "test_bench.db"
	db, err := database.New(dbPath)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer os.Remove(dbPath)
	defer db.Conn().Close()

	subject := "Computer Science"
	year := 2023
	numQuestions := 100
	questions := make([]string, numQuestions)
	for i := 0; i < numQuestions; i++ {
		questions[i] = fmt.Sprintf("Question %d for %s", i, subject)
	}

	b.Run("N+1 Inserts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, q := range questions {
				_, err := db.Conn().Exec("INSERT INTO pyq_bank (subject, question_text, year) VALUES (?, ?, ?)", subject, q, year)
				if err != nil {
					b.Fatalf("Insert failed: %v", err)
				}
			}
			// Cleanup for next iteration
			db.Conn().Exec("DELETE FROM pyq_bank")
		}
	})

	b.Run("Batch Inserts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := db.SavePYQs(subject, questions, year)
			if err != nil {
				b.Fatalf("Batch insert failed: %v", err)
			}
			// Cleanup for next iteration
			db.Conn().Exec("DELETE FROM pyq_bank")
		}
	})
}

func TestScrapePastPapers(t *testing.T) {
	dbPath := "test_scrape.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer os.Remove(dbPath)
	defer db.Conn().Close()

	scraper := NewPYQScraper(db)
	err = scraper.ScrapePastPapers(context.Background(), "Physics")
	if err != nil {
		t.Errorf("ScrapePastPapers failed: %v", err)
	}

	// Verify data was inserted
	var count int
	err = db.Conn().QueryRow("SELECT COUNT(*) FROM pyq_bank WHERE subject = ?", "Physics").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query pyq_bank: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 questions in database, got %d", count)
	}
}
