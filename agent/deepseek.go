package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"deepseek-trader/config"
)

type DecisionAgent interface {
	Decide(ctx context.Context, snap Snapshot) (Decision, error)
}

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
	reqBody := map[string]any{
		"model":           a.cfg.DeepseekModel,
		"messages":        []map[string]string{{"role": "system", "content": "You are a crypto trading agent. Output ONLY compact JSON."}, {"role": "user", "content": prompt}},
		"temperature":     0.2,
		"response_format": map[string]string{"type": "json_object"},
	}
	b, _ := json.Marshal(reqBody)
	url := a.cfg.DeepseekBaseURL
	if url == "" {
		url = "https://api.deepseek.com"
	}
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	url += "/v1/chat/completions"
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
	b, _ := json.Marshal(s)
	return "Given this trading snapshot (JSON), decide next action. " +
		"Return ONLY JSON: {\"action\":\"buy|sell|none\",\"symbol\":\"BTCUSDT\",\"size\":0.001,\"order\":\"market|limit\",\"limitPrice\":12345}. Snapshot: " + string(b)
}
