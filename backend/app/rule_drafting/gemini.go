package rule_drafting

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

const geminiBase = "https://generativelanguage.googleapis.com/v1beta/models/"

// Gemini calls the Gemini REST API (free tier) in JSON mode with temperature 0.
type Gemini struct {
	APIKey string
	Model  string
	// BaseURL replaces the Google endpoint; tests point it at httptest.
	BaseURL string
	Client  *http.Client
}

// NewGemini returns a client with a 20 second timeout.
func NewGemini(apiKey, model string) *Gemini {
	return &Gemini{APIKey: apiKey, Model: model, Client: &http.Client{Timeout: 20 * time.Second}}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction geminiContent   `json:"systemInstruction"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  struct {
		ResponseMimeType string  `json:"responseMimeType"`
		Temperature      float64 `json:"temperature"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

// ErrProviderQuota is returned when the provider says the free quota is used up.
var ErrProviderQuota = errors.New("model provider quota exhausted")

func (g *Gemini) Complete(ctx context.Context, system, user string) (string, error) {
	req := geminiRequest{
		SystemInstruction: geminiContent{Parts: []geminiPart{{Text: system}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: user}}}},
	}
	req.GenerationConfig.ResponseMimeType = "application/json"
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	base := g.BaseURL
	if base == "" {
		base = geminiBase
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+g.Model+":generateContent", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", g.APIKey)

	resp, err := g.Client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", ErrProviderQuota
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini returned status %d", resp.StatusCode)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("gemini reply: %w", err)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini returned no text")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
