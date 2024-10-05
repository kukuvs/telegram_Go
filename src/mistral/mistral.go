package mistral

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MistralClient структура для работы с API Mistral
type MistralClient struct {
	ApiKey string
}

// GenerateText выполняет запрос к Mistral для генерации текста по prompt
func (client *MistralClient) GenerateText(prompt string) (string, error) {
	ctx := context.Background() // Создаем контекст для запроса
	return client.GenerateTextWithContext(ctx, prompt)
}

// GenerateTextWithContext позволяет использовать контекст для отмены долгих запросов
func (client *MistralClient) GenerateTextWithContext(ctx context.Context, prompt string) (string, error) {
	url := "https://api.mistral.ai/v1/chat/completions"

	// Добавляем в начало каждого запроса
	var prePrompt string = "отвечай коротко и на русском ели тебя не просят иного при написании кода добавляй полноценные описания функций и классов так же комментарии ( если что-то уже написанно в коде в виде комментария не нужно повторять это ещё раз) старайся как можно точнее вести диалог не сворачивая на другую тему если есть предложения по улучшению того или иного кода или текста говори их "

	// Формируем тело запроса с учётом нужной модели и параметров
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "mistral-large-latest",
		"messages":    []map[string]string{{"role": "user", "content": prePrompt + " " + prompt}},
		"temperature": 0.7,
		"max_tokens":  100000,
	})
	if err != nil {
		return "", err
	}

	// Создаем запрос с использованием контекста
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Устанавливаем необходимые заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiKey))

	// Выполняем запрос и обрабатываем результат
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при запросе к Mistral API: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Mistral API вернул статус: %v", resp.Status)
	}

	// Читаем тело ответа
	var responseData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", fmt.Errorf("ошибка при разборе JSON: %v", err)
	}

	// Возвращаем сгенерированный контент, если он доступен
	if len(responseData.Choices) > 0 {
		return responseData.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("не удалось получить контент от Mistral API")
}
