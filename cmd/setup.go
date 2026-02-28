package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🦞 StudyClaw Setup Wizard")
	fmt.Println("========================")
	fmt.Println()

	fmt.Print("Enter your Gemini API key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	fmt.Print("Enter your WhatsApp phone number (e.g. +91XXXXXXXXXX): ")
	phone, _ := reader.ReadString('\n')
	phone = strings.TrimSpace(phone)

	fmt.Print("Daily quiz time (24h format, e.g. 20:00): ")
	quizTime, _ := reader.ReadString('\n')
	quizTime = strings.TrimSpace(quizTime)
	if quizTime == "" {
		quizTime = "20:00"
	}

	fmt.Printf("Telegram Mini App URL (default: %s): ", config.DefaultTelegramWebAppURL)
	webAppURL, _ := reader.ReadString('\n')
	webAppURL = strings.TrimSpace(webAppURL)
	if webAppURL == "" {
		webAppURL = config.DefaultTelegramWebAppURL
	}

	cfgData := map[string]interface{}{
		"gemini_api_key": apiKey,
		"whatsapp": map[string]string{
			"owner_number": phone,
		},
		"scheduler": map[string]string{
			"daily_quiz_time": quizTime,
		},
		"telegram_webapp_url": webAppURL,
	}

	data, _ := json.MarshalIndent(cfgData, "", "  ")
	if err := os.WriteFile("config.json", data, 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ config.json created successfully!")
	fmt.Println("Run `go run cmd/main.go` to start StudyClaw.")
}
