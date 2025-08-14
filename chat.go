package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- Gemini API Structures ---

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// --- API Call Logic ---

func callGemini(ctx context.Context, apiKey string, history []ChatMessage, userMsg string) (string, error) {
	if apiKey == "" {
		return "", errors.New("GEMINI_API_KEY environment variable not set")
	}

	var contents []geminiContent
	for _, msg := range history {
		role := "user"
		if msg.Role == "model" {
			role = "model"
		}
		contents = append(contents, geminiContent{Role: role, Parts: []geminiPart{{Text: msg.Content}}})
	}
	contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{{Text: userMsg}}})

	reqBody := geminiRequest{Contents: contents}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + apiKey
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(data))
	}

	var gr geminiResponse
	if err := json.Unmarshal(data, &gr); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if gr.Error != nil {
		return "", errors.New(gr.Error.Message)
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content in gemini response")
	}

	return gr.Candidates[0].Content.Parts[0].Text, nil
}

