package agent

import (
	"bytes"
	"context"
	_ "embed" //nolint:gci
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"deepseek-trader/config"
)

//go:embed promt.txt
var systemPrompt string

type DeepseekAgent struct {
	http *http.Client
	cfg  *config.Settings
}

func NewDeepseekAgent(cfg *config.Settings) *DeepseekAgent {
	return &DeepseekAgent{http: &http.Client{Timeout: 20 * time.Second}, cfg: cfg}
}

func (a *DeepseekAgent) Decide(ctx context.Context, snap Snapshot) (Decision, error) {
	if a.cfg.DeepseekAPIKey == "" {
		return Decision{Action: "none"}, nil
	}

	prompt := buildPrompt(snap)

	requestMessages := []RequestMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	request := DeepseekRequest{
		Model:          a.cfg.DeepseekModel,
		Messages:       requestMessages,
		Temperature:    0.2,
		ResponseFormat: ResponseFormat{Type: "json_object"},
	}

	b, err := json.Marshal(request)
	if err != nil {
		return Decision{Action: "none"}, err
	}

	url := a.buildUrl()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return Decision{Action: "none"}, err
	}

	req.Header.Set("Authorization", "Bearer "+a.cfg.DeepseekAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.http.Do(req)
	if err != nil {
		return Decision{Action: "none"}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Decision{Action: "none"}, errors.New("deepseek http error")
	}

	var out DeepseekResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Decision{Action: "none"}, err
	}

	if len(out.Choices) == 0 {
		return Decision{Action: "none"}, nil
	}

	var dec Decision
	if err := json.Unmarshal([]byte(out.Choices[0].Message.Content), &dec); err != nil {
		return Decision{Action: "none"}, nil
	}

	if dec.Action == "" {
		dec.Action = "none"
	}

	if dec.Symbol == "" {
		dec.Symbol = "BTCUSDT"
	}

	if dec.Order == "" {
		dec.Order = "market"
	}

	return dec, nil
}

func buildPrompt(s Snapshot) string {
	b, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	fmt.Println(string(b))
	return "Input Snapshot (JSON): " + string(b)
}

func (a *DeepseekAgent) buildUrl() string {
	url := a.cfg.DeepseekBaseURL
	if url == "" {
		url = "https://api.deepseek.com"
	}

	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	url += "/v1/chat/completions"
	return url
}
