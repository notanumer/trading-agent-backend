package services

import (
    "fmt"
    "sort"
    "time"
)

type TradeSummary struct {
    Time        time.Time `json:"time"`
    Coin        string    `json:"coin"`
    Direction   string    `json:"direction"` // Open Long | Close Long
    Price       float64   `json:"price"`
    Size        float64   `json:"size"`
    TradeValue  float64   `json:"tradeValue"`
    Fee         float64   `json:"fee"`
    ClosedPnL   float64   `json:"closedPnl"`
}

type FeeParams struct {
    Maker float64
    Taker float64
    Discount float64 // 0..1 combined referral+staking
}

// BuildTradeSummary converts HL historicalOrders into rows with PnL using FIFO average.
// side B -> buy (open long), side A -> sell (close long)
func BuildTradeSummary(raw []map[string]any, fee FeeParams) []TradeSummary {
    // sort by statusTimestamp/ timestamp ascending
    type row struct{
        ts time.Time
        item map[string]any
    }
    rows := make([]row, 0, len(raw))
    for _, it := range raw {
        ord, _ := it["order"].(map[string]any)
        if ord == nil { continue }
        var ts int64
        if v, ok := it["statusTimestamp"].(float64); ok { ts = int64(v) } else if v2, ok2 := ord["timestamp"].(float64); ok2 { ts = int64(v2) }
        rows = append(rows, row{ts: time.UnixMilli(ts), item: it})
    }
    sort.Slice(rows, func(i,j int) bool { return rows[i].ts.Before(rows[j].ts) })

    type pos struct{ qty float64; avg float64 }
    coinPos := map[string]*pos{}
    out := make([]TradeSummary, 0, len(rows))

    for _, r := range rows {
        it := r.item
        ord, _ := it["order"].(map[string]any)
        coin, _ := ord["coin"].(string)
        side, _ := ord["side"].(string)
        pxStr, _ := ord["limitPx"].(string)
        szStr, _ := ord["origSz"].(string)
        if szStr == "" { szStr, _ = ord["sz"].(string) }
        price := parseF(pxStr)
        qty := parseF(szStr)
        if qty == 0 { continue }
        p := coinPos[coin]
        if p == nil { p = &pos{}; coinPos[coin] = p }
        rate := fee.Maker
        // decide maker/taker by orderType / tif / isTrigger
        orderType, _ := ord["orderType"].(string)
        tif, _ := ord["tif"].(string)
        isTrigger, _ := ord["isTrigger"].(bool)
        if orderType == "Market" || tif == "FrontendMarket" || isTrigger { rate = fee.Taker }
        effRate := rate * (1 - fee.Discount)
        if effRate < 0 { effRate = 0 }
        calcFee := price * qty * effRate
        if side == "B" { // buy -> open long (increase pos)
            newQty := p.qty + qty
            if newQty <= 0 { newQty = 0 }
            if newQty == 0 { p.avg = 0 } else { p.avg = (p.avg*p.qty + price*qty) / newQty }
            p.qty = newQty
            out = append(out, TradeSummary{Time: r.ts, Coin: coin, Direction: "Open Long", Price: price, Size: qty, TradeValue: price*qty, Fee: calcFee, ClosedPnL: 0})
        } else { // side A -> sell -> close long (reduce)
            realized := 0.0
            closeQty := qty
            if closeQty > p.qty { closeQty = p.qty }
            realized += (price - p.avg) * closeQty
            p.qty -= closeQty
            if p.qty == 0 { p.avg = 0 }
            out = append(out, TradeSummary{Time: r.ts, Coin: coin, Direction: "Close Long", Price: price, Size: qty, TradeValue: price*qty, Fee: calcFee, ClosedPnL: realized})
        }
    }
    return out
}

func parseF(s string) float64 {
    if s == "" { return 0 }
    var f float64
    _, _ = fmt.Sscan(s, &f)
    return f
}


