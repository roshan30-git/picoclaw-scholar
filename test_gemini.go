package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/roshan30-git/picoclaw-scholar/pkg/providers"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

func main() {
	godotenv.Load()
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		log.Fatal("no key")
	}

	provider, err := providers.NewGeminiProvider(key)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Println("Sending chat to Gemini...")
	resp, err := provider.Chat(ctx, []tools.Message{{Role: "user", Content: "hello"}}, nil, "", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Resp:", resp.Content)
}
