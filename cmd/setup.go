package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

	config := map[string]interface{}{
		"gemini_api_key": apiKey,
		"whatsapp": map[string]string{
			"owner_number": phone,
		},
		"scheduler": map[string]string{
			"daily_quiz_time": quizTime,
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile("config.json", data, 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ config.json created successfully!")
	fmt.Println("Run `go run cmd/main.go` to start StudyClaw.")
}
