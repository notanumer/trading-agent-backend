package services

import (
	"fmt"
	"sort"
	"time"
)

type TradeSummary struct {
	Time       time.Time `json:"time"`
	Coin       string    `json:"coin"`
	Direction  string    `json:"direction"` // Open Long | Close Long
	Price      float64   `json:"price"`
	Size       float64   `json:"size"`
	TradeValue float64   `json:"tradeValue"`
	Fee        float64   `json:"fee"`
	ClosedPnL  float64   `json:"closedPnl"`
}

type FeeParams struct {
	Maker    float64
	Taker    float64
	Discount float64 // 0..1 combined referral+staking
}

// BuildTradeSummary converts HL historicalOrders into rows with PnL using FIFO average.
// side B -> buy (open long), side A -> sell (close long)
func BuildTradeSummary(raw []map[string]any, fee FeeParams) []TradeSummary {
	rows := extractAndSortRows(raw)
	coinPos := make(map[string]*position)
	out := make([]TradeSummary, 0, len(rows))

	for _, r := range rows {
		summary := processOrder(r, coinPos, fee)
		if summary != nil {
			out = append(out, *summary)
		}
	}
	return out
}

// -- Вспомогательные типы и функции --

type position struct {
	qty float64
	avg float64
}

type orderRow struct {
	ts   time.Time
	item map[string]any
}

func extractAndSortRows(raw []map[string]any) []orderRow {
	rows := make([]orderRow, 0, len(raw))
	for _, it := range raw {
		ord, ok := it["order"].(map[string]any)
		if !ok {
			continue
		}
		ts := extractTimestamp(it, ord)
		rows = append(rows, orderRow{ts: ts, item: it})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ts.Before(rows[j].ts)
	})
	return rows
}

func extractTimestamp(it, ord map[string]any) time.Time {
	var ts int64
	if v, ok := it["statusTimestamp"].(float64); ok {
		ts = int64(v)
	} else if v2, ok := ord["timestamp"].(float64); ok {
		ts = int64(v2)
	}
	return time.UnixMilli(ts)
}

func getEffectiveFeeRate(ord map[string]any, fee FeeParams) float64 {
	orderType, _ := ord["orderType"].(string)
	tif, _ := ord["tif"].(string)
	isTrigger, _ := ord["isTrigger"].(bool)

	rate := fee.Maker
	if orderType == "Market" || tif == "FrontendMarket" || isTrigger {
		rate = fee.Taker
	}
	effRate := rate * (1 - fee.Discount)
	if effRate < 0 {
		effRate = 0
	}
	return effRate
}

func processOrder(r orderRow, coinPos map[string]*position, fee FeeParams) *TradeSummary {
	ord, ok := r.item["order"].(map[string]any)
	if !ok {
		return nil
	}

	coin, _ := ord["coin"].(string)
	side, _ := ord["side"].(string)
	pxStr, _ := ord["limitPx"].(string)
	szStr, _ := ord["origSz"].(string)
	if szStr == "" {
		szStr, _ = ord["sz"].(string)
	}

	price := parseF(pxStr)
	qty := parseF(szStr)
	if qty == 0 {
		return nil
	}

	p, exists := coinPos[coin]
	if !exists {
		p = &position{}
		coinPos[coin] = p
	}

	effRate := getEffectiveFeeRate(ord, fee)
	calcFee := price * qty * effRate

	if side == "B" {
		return handleBuyOrder(r.ts, coin, price, qty, calcFee, p)
	}

	return handleSellOrder(r.ts, coin, price, qty, calcFee, p)
}

func handleBuyOrder(ts time.Time, coin string, price, qty, fee float64, p *position) *TradeSummary {
	newQty := p.qty + qty
	if newQty <= 0 {
		newQty = 0
	}
	if newQty == 0 {
		p.avg = 0
	} else {
		p.avg = (p.avg*p.qty + price*qty) / newQty
	}
	p.qty = newQty

	return &TradeSummary{
		Time:       ts,
		Coin:       coin,
		Direction:  "Open Long",
		Price:      price,
		Size:       qty,
		TradeValue: price * qty,
		Fee:        fee,
		ClosedPnL:  0,
	}
}

func handleSellOrder(ts time.Time, coin string, price, qty, fee float64, p *position) *TradeSummary {
	closeQty := qty
	if closeQty > p.qty {
		closeQty = p.qty
	}
	realized := (price - p.avg) * closeQty
	p.qty -= closeQty
	if p.qty == 0 {
		p.avg = 0
	}

	return &TradeSummary{
		Time:       ts,
		Coin:       coin,
		Direction:  "Close Long",
		Price:      price,
		Size:       qty, // исходный qty, как в оригинале
		TradeValue: price * qty,
		Fee:        fee,
		ClosedPnL:  realized,
	}
}

func parseF(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	_, _ = fmt.Sscan(s, &f)
	return f
}
