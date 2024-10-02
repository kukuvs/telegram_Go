package telegram

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kukuvs/telegram_Go/src/mistral" // Импорт клиента Mistral
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// Bot структура для Telegram бота
type Bot struct {
	BotClient  *telego.Bot
	MistralAPI *mistral.MistralClient
}

// MistralResponse структура для разбора ответа от Mistral API
type MistralResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// HandleMessages обрабатывает входящие сообщения
func (b *Bot) HandleMessages() {
	updates, err := b.BotClient.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil {
			if update.Message.Document != nil {
				// Обработка документа
				fileID := update.Message.Document.FileID
				file, err := b.BotClient.GetFile(&telego.GetFileParams{FileID: fileID})
				if err != nil {
					log.Printf("Ошибка при получении информации о файле: %v", err)
					continue
				}

				// Скачивание файла
				resp, err := http.Get("https://api.telegram.org/file/bot" + b.BotClient.Token() + "/" + file.FilePath)
				if err != nil {
					log.Printf("Ошибка при скачивании файла: %v", err)
					continue
				}
				defer resp.Body.Close()

				// Сохранение файла на диск
				out, err := os.Create("downloaded_file")
				if err != nil {
					log.Printf("Ошибка при создании файла: %v", err)
					continue
				}
				defer out.Close()

				_, err = io.Copy(out, resp.Body)
				if err != nil {
					log.Printf("Ошибка при копировании файла: %v", err)
					continue
				}

				// Чтение содержимого файла
				content, err := os.ReadFile("downloaded_file")
				if err != nil {
					log.Printf("Ошибка при чтении файла: %v", err)
					continue
				}

				// Отправка содержимого файла пользователю
				msg := &telego.SendMessageParams{
					ChatID: telegoutil.ID(update.Message.Chat.ID),
					Text:   string(content),
				}

				_, err = b.BotClient.SendMessage(msg)
				if err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
				}
			} else {
				userMessage := update.Message.Text

				// Запрос к Mistral API
				response, err := b.MistralAPI.GenerateText(userMessage)
				if err != nil {
					log.Printf("Ошибка при запросе к Mistral: %v", err)
					continue
				}

				// Парсинг ответа от Mistral
				var mistralResp MistralResponse
				err = json.Unmarshal([]byte(response), &mistralResp)
				if err != nil {
					log.Printf("Ошибка при разборе ответа от Mistral: %v", err)
					continue
				}

				// Извлекаем текст ответа
				if len(mistralResp.Choices) > 0 {
					content := mistralResp.Choices[0].Message.Content

					// Отправляем ответ пользователю
					msg := &telego.SendMessageParams{
						ChatID: telegoutil.ID(update.Message.Chat.ID),
						Text:   content,
					}

					_, err = b.BotClient.SendMessage(msg)
					if err != nil {
						log.Printf("Ошибка при отправке сообщения: %v", err)
					}
				}
			}
		}
	}
}
