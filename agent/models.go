package agent

type Decision struct {
	Action     string  `json:"action"` // buy|sell|none
	Symbol     string  `json:"symbol"` // e.g., BTCUSDT
	Size       float64 `json:"size"`   // in base units
	Order      string  `json:"order"`  // market|limit
	LimitPrice float64 `json:"limitPrice,omitempty"`
}

type Snapshot struct {
	Balance float64 `json:"balance"`
	PnL     float64 `json:"pnl"`
	ROE     float64 `json:"roe"`
	Trades  []any   `json:"trades"` // minimal for now
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
