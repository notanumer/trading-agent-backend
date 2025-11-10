package agent

import "deepseek-trader/hyperliquid"

type Decision struct {
	Action     string  `json:"action"` // buy|sell|none
	Symbol     string  `json:"symbol"` // e.g., BTCUSDT
	Size       float64 `json:"size"`   // in base units
	Order      string  `json:"order"`  // market|limit
	LimitPrice float64 `json:"limitPrice"`
	Targets    Targets `json:"targets"`
}

type Targets struct {
	TP1 float64 `json:"tp1"`
	TP2 float64 `json:"tp2"`
	TP3 float64 `json:"tp3"`
	SL  float64 `json:"sl"`
}

type Snapshot struct {
	Balance         float64                         `json:"balance"`
	PnL             float64                         `json:"pnl"`
	ROE             float64                         `json:"roe"`
	Trades          []interface{}                   `json:"trades"` // minimal for now
	CoinsMids       map[string]string               `json:"coinsMids"`
	Runtime         Runtime                         `json:"runtime,omitempty"`
	Decisions       []interface{}                   `json:"decisions"`
	Meta            hyperliquid.ExchangeMeta        `json:"meta"`
	OrderBooks      []hyperliquid.OrderBookSnapshot `json:"orderBooks"`
	CandleSnapshots map[string][]hyperliquid.Candle `json:"candleSnapshots"`
}

type Runtime struct {
	Date        string      `json:"date,omitempty"`
	Signature   string      `json:"signature,omitempty"`
	AgentConfig AgentConfig `json:"agentConfig,omitempty"`
}

type AgentConfig struct {
	MaxSteps    int     `json:"maxSteps,omitempty"`
	MaxRetries  int     `json:"maxRetries,omitempty"`
	BaseDelay   float64 `json:"baseDelay,omitempty"`
	InitialCash float64 `json:"initialCash,omitempty"`
}

type DeepseekRequest struct {
	Model          string           `json:"model"`
	Messages       []RequestMessage `json:"messages"`
	Temperature    float64          `json:"temperature"`
	ResponseFormat ResponseFormat   `json:"response_format"`
}

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type DeepseekResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Message struct {
	Content string `json:"content"`
}

var Coins = []string{
	"BTC",
	"ETH",
	"SOL",
	"XRP",
	"DOGE",
	"ADA",
	"DOT",
	"LINK",
	"WLD",
}

// FilterCoinsMids returns a new map that contains only entries whose keys are present in the allowed coins list.
func FilterCoinsMids(all map[string]string) map[string]string {
	if len(all) == 0 {
		return nil
	}

	allowed := make(map[string]struct{}, len(Coins))
	for _, c := range Coins {
		allowed[c] = struct{}{}
	}

	out := make(map[string]string, len(Coins))
	for k, v := range all {
		if _, ok := allowed[k]; ok {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
