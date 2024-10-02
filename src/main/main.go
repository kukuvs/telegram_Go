package main

import (
	"log"

	"github.com/kukuvs/telegram_Go/src/config"
	"github.com/kukuvs/telegram_Go/src/mistral"
	"github.com/kukuvs/telegram_Go/src/telegram"

	"github.com/mymmrac/telego"
)

func main() {
	// Загружаем конфиг из файла
	cfg, err := config.LoadConfig("../1.txt")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Создаем нового бота
	botClient, err := telego.NewBot(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Создаем клиента для Mistral API
	mistralClient := &mistral.MistralClient{
		ApiKey: cfg.MistralApiKey,
	}

	// Создаем Telegram-бота
	bot := &telegram.Bot{
		BotClient:  botClient,
		MistralAPI: mistralClient,
	}

	// Запускаем обработку сообщений
	bot.HandleMessages()
}
