package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Config struct {
	TelegramToken string
	MistralApiKey string
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var tokens []string
	for scanner.Scan() {
		tokens = append(tokens, strings.TrimSpace(scanner.Text()))
	}

	if len(tokens) < 2 {
		log.Fatal("Invalid config file format")
	}

	return &Config{
		TelegramToken: tokens[0],
		MistralApiKey: tokens[1],
	}, nil
}
