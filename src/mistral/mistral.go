package mistral

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// MistralClient структура для работы с API
type MistralClient struct {
	ApiKey string
}

// Запрос на генерацию текста по prompt
func (client *MistralClient) GenerateText(prompt string) (string, error) {
	url := "https://api.mistral.ai/v1/chat/completions"

	script := "отвечай коротко и на русском ели тебя не просят иного при написании кода добавляй полноценные описания функций и классов так же комментарии ( если что-то уже написанно в коде в виде комментария не нужно повторять это ещё раз) старайся как можно точнее вести диалог не сворачивая на другую тему если есть предложения по улучшению того или иного кода или текста говори их и учти то общение происходит через telegram "

	// Формируем тело запроса с правильной структурой
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "mistral-large-latest", // Указываем идентификатор модели
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": script + prompt,
			},
		},
		"temperature": 0.7,    // Опциональные параметры
		"max_tokens":  100000, // Ограничение на количество токенов
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiKey))

	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
