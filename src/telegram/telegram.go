package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kukuvs/telegram_Go/src/mistral"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// Bot структура для Telegram бота
type Bot struct {
	BotClient  *telego.Bot
	MistralAPI *mistral.MistralClient
}

// HandleMessages обрабатывает входящие сообщения
func (b *Bot) HandleMessages(ctx context.Context) {
	updates, err := b.BotClient.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Fatalf("Ошибка получения обновлений: %v", err)
		return
	}

	var wg sync.WaitGroup
	messageChan := make(chan telego.Update)

	go func() {
		defer close(messageChan)
		for update := range updates {
			if update.Message != nil {
				messageChan <- update
			}
		}
	}()

	for update := range messageChan {
		wg.Add(1)
		go func(update telego.Update) {
			defer wg.Done()
			if update.Message.Document != nil {
				b.handleDocument(ctx, update)
			} else {
				b.handleTextMessage(ctx, update.Message)
			}
		}(update)
	}

	wg.Wait()
}

// handleDocument обрабатывает сообщения с документами
func (b *Bot) handleDocument(ctx context.Context, update telego.Update) {
	fileID := update.Message.Document.FileID
	fileName := update.Message.Document.FileName

	if !isSupportedFileType(fileName) {
		b.sendError(update.Message.Chat.ID, "Поддерживаются только .txt файлы", nil)
		return
	}

	file, err := b.BotClient.GetFile(&telego.GetFileParams{FileID: fileID})
	if err != nil {
		b.sendError(update.Message.Chat.ID, "Ошибка при получении информации о файле", err)
		return
	}

	fileContent, err := b.downloadFile(ctx, file.FilePath)
	if err != nil {
		b.sendError(update.Message.Chat.ID, "Ошибка при скачивании файла", err)
		return
	}

	combinedContent := string(fileContent)
	if update.Message.Caption != "" {
		combinedContent = update.Message.Caption + "\n" + combinedContent
	}

	response, err := b.MistralAPI.GenerateText(combinedContent)
	if err != nil {
		b.sendError(update.Message.Chat.ID, "Ошибка при запросе к Mistral API", err)
		return
	}

	b.sendMessage(update.Message.Chat.ID, response)
}

// handleTextMessage обрабатывает текстовые сообщения
func (b *Bot) handleTextMessage(ctx context.Context, message *telego.Message) {
	response, err := b.MistralAPI.GenerateText(message.Text)
	if err != nil {
		b.sendError(message.Chat.ID, "Ошибка при запросе к Mistral API", err)
		return
	}

	b.sendMessage(message.Chat.ID, response)
}

// isSupportedFileType returns true if the given file name has a supported extension
func isSupportedFileType(fileName string) bool {
	fileName = strings.ToLower(fileName)
	ext := filepath.Ext(fileName)
	return ext == ".txt"
}

// downloadFile загружает файл по заданному пути
func (b *Bot) downloadFile(ctx context.Context, filePath string) ([]byte, error) {
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.BotClient.Token(), filePath)

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе файла: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка при получении файла: статус %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// sendMessage отправляет сообщение, автоматически разбивая длинные сообщения
func (b *Bot) sendMessage(chatID int64, content string) {
	const maxMessageLength = 4096
	for len(content) > 0 {
		part := content
		if len(content) > maxMessageLength {
			part = content[:maxMessageLength]
			content = content[maxMessageLength:]
		} else {
			content = ""
		}

		msg := &telego.SendMessageParams{
			ChatID: telegoutil.ID(chatID),
			Text:   part,
		}

		if _, err := b.BotClient.SendMessage(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
	}
}

// sendError отправляет сообщение об ошибке в чат
func (b *Bot) sendError(chatID int64, message string, err error) {
	var errorMessage string
	if err != nil {
		errorMessage = fmt.Sprintf("%s: %v", message, err)
	} else {
		errorMessage = message
	}
	b.sendMessage(chatID, errorMessage)
}
