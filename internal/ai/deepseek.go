package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log" 
	"net/http"
	"time"
)

const (
		deepSeekEndpoint = "https://api.deepseek.com/v1/chat/completions"
		deepSeekModel = "deepseek-chat"
) 

type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callDeepSeek(prompt string, apiKey string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}
	if apiKey == "" {
		return "", fmt.Errorf("API key is empty")
	}

	requestBody := DeepSeekRequest{
		Model: deepSeekModel,
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	log.Printf("DeepSeek request body size: %d bytes", len(jsonData))

	req, err := http.NewRequest(
		"POST",
		deepSeekEndpoint,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request to DeepSeek API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	log.Printf(
		"DeepSeek API response status: %d, body length: %d",
		resp.StatusCode,
		len(body),
	)

	if resp.StatusCode != http.StatusOK {
		log.Printf("DeepSeek API error response: %s", string(body))
		return "", fmt.Errorf(
			"deepseek API returned status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var deepSeekResponse DeepSeekResponse
	if err := json.Unmarshal(body, &deepSeekResponse); err != nil {
		return "", fmt.Errorf("unmarshaling response: %w", err)
	}

	if len(deepSeekResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from DeepSeek API")
	}

	return deepSeekResponse.Choices[0].Message.Content, nil
}
