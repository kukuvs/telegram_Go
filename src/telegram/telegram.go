package telegram

import (
	"log"

	"github.com/kukuvs/telegram_Go/src/mistral" // Импорт клиента Mistral

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// Bot структура для Telegram бота
type Bot struct {
	BotClient  *telego.Bot
	MistralAPI *mistral.MistralClient
}

// Обработка сообщений
func (b *Bot) HandleMessages() {
	updates, err := b.BotClient.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil {
			userMessage := update.Message.Text
			response, err := b.MistralAPI.GenerateText(userMessage)
			if err != nil {
				log.Printf("Error while sending request to Mistral: %v", err)
				continue
			}

			msg := &telego.SendMessageParams{
				ChatID: telegoutil.ID(update.Message.Chat.ID),
				Text:   response,
			}

			_, err = b.BotClient.SendMessage(msg)
			if err != nil {
				log.Printf("Error while sending message: %v", err)
			}
		}
	}
}
