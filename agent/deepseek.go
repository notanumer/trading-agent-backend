package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"deepseek-trader/config"
)

type DeepseekAgent struct {
	http *http.Client
	cfg  *config.Settings
}

func NewDeepseekAgent(cfg *config.Settings) *DeepseekAgent {
	return &DeepseekAgent{http: &http.Client{Timeout: 20 * time.Minute}, cfg: cfg}
}

func (a *DeepseekAgent) Decide(ctx context.Context, snap Snapshot) (Decision, error) {
	if a.cfg.DeepseekAPIKey == "" {
		return Decision{Action: "none"}, nil
	}

	systemPrompt := fmt.Sprintf(
		systemPromptTemplate,
		formatFloat(snap.Balance),
		formatFloat(snap.PnL),
		formatFloat(snap.ROE*100),
		len(snap.Trades),
	)

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
	// Build summary section
	summaryBuf := bytes.Buffer{}

	// Recent trades
	if len(s.Trades) > 0 {
		summaryBuf.WriteString("## Recent Trades\n")
		summaryBuf.WriteString("Total trades executed: ")
		summaryBuf.WriteString(formatInt(len(s.Trades)))
		summaryBuf.WriteString("\n\n")
	}

	// Current prices
	if len(s.CoinsMids) > 0 {
		summaryBuf.WriteString("## Current Mid Prices\n```\n")
		for coin, price := range s.CoinsMids {
			summaryBuf.WriteString(coin)
			summaryBuf.WriteString(": $")
			summaryBuf.WriteString(price)
			summaryBuf.WriteString("\n")
		}
		summaryBuf.WriteString("```\n\n")
	}

	// Order books
	if len(s.OrderBooks) > 0 {
		summaryBuf.WriteString("## Order Book Data\n")
		summaryBuf.WriteString("Available order books: ")
		summaryBuf.WriteString(formatInt(len(s.OrderBooks)))
		summaryBuf.WriteString(" symbols\n\n")
	}

	// Candles
	if len(s.CandleSnapshots) > 0 {
		summaryBuf.WriteString("## Candlestick Data (15m intervals)\n")
		for coin, candles := range s.CandleSnapshots {
			if len(candles) == 0 {
				continue
			}
			summaryBuf.WriteString("- ")
			summaryBuf.WriteString(coin)
			summaryBuf.WriteString(": ")
			summaryBuf.WriteString(formatInt(len(candles)))
			summaryBuf.WriteString(" candles\n")
		}
		summaryBuf.WriteString("\n")
	}

	// Recent decisions
	if len(s.Decisions) > 0 {
		summaryBuf.WriteString("## Recent Decisions\n")
		summaryBuf.WriteString("Previous decisions count: ")
		summaryBuf.WriteString(formatInt(len(s.Decisions)))
		summaryBuf.WriteString("\n\n")
	}

	// Marshal full JSON
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		jsonData, _ = json.Marshal(s)
	}

	// Format using template
	return fmt.Sprintf(
		userPromptTemplate,
		formatFloat(s.Balance),
		formatFloat(s.PnL),
		formatFloat(s.ROE*100),
		summaryBuf.String(),
		string(jsonData),
	)
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

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatInt(i int) string {
	return strconv.Itoa(i)
}
