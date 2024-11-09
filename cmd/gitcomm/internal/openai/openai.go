package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ktappdev/gitcomm/config"
	"io"
	"net/http"
)

type Client struct {
	apiKey string
	apiURL string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		apiURL: config.OpenAIAPIURL,
		client: &http.Client{},
	}
}

func (c *Client) SendPrompt(prompt string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	})

	req, _ := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return extractContent(result)
}

func extractContent(result map[string]interface{}) (string, error) {
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is not a string")
	}

	return content, nil
}
