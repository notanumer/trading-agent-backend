package hyperliquid

import (
	"encoding/json"
	"fmt"
)

type UserFeeResponse struct {
	UserCrossRate          string                 `json:"userCrossRate"`
	UserAddRate            string                 `json:"userAddRate"`
	ActiveReferralDiscount string                 `json:"activeReferralDiscount"`
	ActiveStakingDiscount  *ActiveStakingDiscount `json:"activeStakingDiscount"`
}

type ActiveStakingDiscount struct {
	Discount string `json:"discount"`
}

type LiveStats struct {
	Balance float64
	PnL     float64
	ROE     float64
}

type UserFees struct {
	UserCrossRate         float64
	UserAddRate           float64
	ReferralDiscount      float64
	StakingActiveDiscount float64
}

type Payload struct {
	Type string  `json:"type"`
	User string  `json:"user"`
	Coin *string `json:"coin,omitempty"`
}

type UserFillsResponse struct {
	UserFills []UserFill `json:"userFills"`
}

type UserFill struct {
	Coin          string `json:"coin"`
	Px            string `json:"px"`
	Sz            string `json:"Sz"`
	Time          int64  `json:"time"`
	Side          string `json:"side"`
	StartPosition string `json:"startPosition"`
	Dir           string `json:"dir"`
	ClosedPnl     string `json:"closedPnl"`
}

// ExchangeMeta represents the full JSON structure with instruments and margin tables.
type ExchangeMeta struct {
	Universe        []Instrument       `json:"universe"`
	MarginTables    []MarginTableEntry `json:"marginTables"`
	CollateralToken int                `json:"collateralToken"`
}

// Instrument describes a tradable instrument in the universe.
type Instrument struct {
	SzDecimals    int    `json:"szDecimals"`
	Name          string `json:"name"`
	MaxLeverage   int    `json:"maxLeverage"`
	MarginTableID int    `json:"marginTableId"`
	IsDelisted    bool   `json:"isDelisted,omitempty"`
	OnlyIsolated  bool   `json:"onlyIsolated,omitempty"`
	MarginMode    string `json:"marginMode,omitempty"`
}

// MarginTable is a set of tiers for leverage by notional size.
type MarginTable struct {
	Description string       `json:"description"`
	MarginTiers []MarginTier `json:"marginTiers"`
}

type MarginTier struct {
	LowerBound  string `json:"lowerBound"`
	MaxLeverage int    `json:"maxLeverage"`
}

// MarginTableEntry maps a numeric table id to its table.
// It unmarshals from a JSON array pair: [id, {table}].
type MarginTableEntry struct {
	ID    int
	Table MarginTable
}

func (e *MarginTableEntry) UnmarshalJSON(data []byte) error {
	var pair []json.RawMessage
	if err := json.Unmarshal(data, &pair); err != nil {
		return err
	}
	if len(pair) != 2 {
		return fmt.Errorf("MarginTableEntry: expected 2 elements, got %d", len(pair))
	}
	if err := json.Unmarshal(pair[0], &e.ID); err != nil {
		return err
	}
	if err := json.Unmarshal(pair[1], &e.Table); err != nil {
		return err
	}
	return nil
}

// OrderBookSnapshot represents the book snapshot with two sides in levels.
// levels[0] and levels[1] are arrays of price levels for each side.
type OrderBookSnapshot struct {
	Coin   string             `json:"coin"`
	Time   int64              `json:"time"`
	Levels [][]OrderBookLevel `json:"levels"`
}

type OrderBookLevel struct {
	Px string `json:"px"`
	Sz string `json:"sz"`
	N  int    `json:"n"`
}

// Candle represents one OHLCV bar entry.
type Candle struct {
	StartTime int64  `json:"t"`
	EndTime   int64  `json:"T"`
	Symbol    string `json:"s"`
	Interval  string `json:"i"`
	Open      string `json:"o"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Volume    string `json:"v"`
	NumTrades int    `json:"n"`
}

type CandleSnapshotRequest struct {
	Type string      `json:"type"`
	Req  RequestBody `json:"req"`
}

type RequestBody struct {
	Coin      string `json:"coin"`
	Interval  string `json:"interval"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}
