package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// Функция для отправки запроса к Mistral AI
func sendMistralRequest(prompt string, apiKey string) (string, error) {
	url := "https://api.mistral.ai/v1/chat/completions" // Заменить на правильный endpoint

	// Создаем тело запроса в формате JSON
	requestBody, err := json.Marshal(map[string]string{
		"prompt": prompt,
	})
	if err != nil {
		return "", err
	}

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Добавляем необходимые заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Функция для обработки сообщений от пользователей
func handleMessages(bot *telego.Bot, apiKey string) {
	// Устанавливаем обработчик для получения обновлений
	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil {
			// Получаем текст сообщения от пользователя
			userMessage := update.Message.Text

			// Отправляем запрос к Mistral AI
			response, err := sendMistralRequest(userMessage, apiKey)
			if err != nil {
				log.Printf("Error while sending request to Mistral: %v", err)
				continue
			}

			// Создаем сообщение для отправки пользователю
			msg := &telego.SendMessageParams{
				ChatID: telegoutil.ID(update.Message.Chat.ID),
				Text:   response,
			}

			// Отправляем сообщение
			_, err = bot.SendMessage(msg)
			if err != nil {
				log.Printf("Error while sending message: %v", err)
			}
		}
	}
}

func main() {
	// Загружаем токен бота и API ключ Mistral из переменных окружения (или можно захардкодить)
	botToken := "YOUR_TELEGRAM_BOT_TOKEN"   // Заменить на реальный токен
	mistralApiKey := "YOUR_MISTRAL_API_KEY" // Заменить на реальный API ключ

	// Создаем нового бота
	bot, err := telego.NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Запускаем обработку сообщений
	handleMessages(bot, mistralApiKey)
}
