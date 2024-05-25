package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	baseURL    = "https://api.openai.com/v1"
	apiKeyName = "OPENAI_API_KEY"
)

type Model string

const (
	ModelGPT35Turbo Model = "gpt-3.5-turbo"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    Model     `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"context_length_exceeded"`
}

type ChatResponse struct {
	Error   *Error   `json:"error"`
	Model   Model    `json:"model"`
	Choices []Choice `json:"choices"`
}

func openAIRequest(ctx context.Context, method string, path string, req any, resp any) error {

	apiKey := os.Getenv(apiKeyName)
	if apiKey == "" {
		return fmt.Errorf("an environment variable named %s is required", apiKeyName)
	}

	content, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewBuffer(content))
	if err != nil {
		return err
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	err = json.NewDecoder(httpResp.Body).Decode(resp)
	if err != nil {
		return err
	}
	return nil
}
