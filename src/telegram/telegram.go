package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kukuvs/telegram_Go/src/mistral"
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

// HandleMessages обрабатывает входящие сообщения с использованием Long Polling
func (b *Bot) HandleMessages(ctx context.Context) {
	// Получаем обновления через Long Polling
	updates, err := b.BotClient.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Обрабатываем каждое обновление
	for update := range updates {
		if update.Message != nil {
			if update.Message.Document != nil {
				go b.handleDocument(ctx, update) // Асинхронная обработка документа
			} else {
				go b.handleTextMessage(ctx, update.Message) // Асинхронная обработка текстового сообщения
			}
		}
	}
}

// handleDocument обрабатывает сообщения с документами
func (b *Bot) handleDocument(ctx context.Context, update telego.Update) {
	fileID := update.Message.Document.FileID
	// Получаем информацию о файле по его FileID
	file, err := b.BotClient.GetFile(&telego.GetFileParams{FileID: fileID})
	if err != nil {
		errorMessage := fmt.Sprintf("Ошибка при получении информации о файле: %v", err)
		b.sendMessage(update.Message.Chat.ID, errorMessage)
		return
	}

	// Скачиваем файл
	fileContent, err := b.downloadFile(ctx, file.FilePath)
	if err != nil {
		errorMessage := fmt.Sprintf("Ошибка при скачивании файла: %v", err)
		b.sendMessage(update.Message.Chat.ID, errorMessage)
		return
	}

	// Отправляем содержимое файла пользователю
	b.sendMessage(update.Message.Chat.ID, string(fileContent))
}

// handleTextMessage обрабатывает текстовые сообщения
func (b *Bot) handleTextMessage(ctx context.Context, message *telego.Message) {
	// Запрос к Mistral API для генерации ответа
	response, err := b.MistralAPI.GenerateText(message.Text)
	if err != nil {
		errorMessage := fmt.Sprintf("Ошибка при запросе к Mistral API: %v", err)
		b.sendMessage(message.Chat.ID, errorMessage)
		return
	}

	// Разбиваем длинный ответ на несколько сообщений, если он превышает лимит Telegram
	b.sendLongMessage(message.Chat.ID, response)
}

// downloadFile загружает файл по заданному пути
func (b *Bot) downloadFile(ctx context.Context, filePath string) ([]byte, error) {
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.BotClient.Token(), filePath)

	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе файла: %v", err)
	}
	defer resp.Body.Close()

	// Чтение содержимого файла
	return io.ReadAll(resp.Body)
}

// sendMessage отправляет сообщение в чат
func (b *Bot) sendMessage(chatID int64, content string) {
	msg := &telego.SendMessageParams{
		ChatID: telegoutil.ID(chatID),
		Text:   content,
	}

	_, err := b.BotClient.SendMessage(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

// sendLongMessage отправляет длинное сообщение, разбивая его на несколько частей, если необходимо
func (b *Bot) sendLongMessage(chatID int64, content string) {
	const maxMessageLength = 4096 // Максимальная длина сообщения в Telegram
	// Разбиваем сообщение на части, если оно превышает максимальную длину
	for len(content) > 0 {
		// Определяем длину текущего сообщения
		part := content
		if len(content) > maxMessageLength {
			part = content[:maxMessageLength]
			content = content[maxMessageLength:]
		} else {
			content = ""
		}

		// Отправляем каждую часть сообщения
		b.sendMessage(chatID, part)
	}
}
