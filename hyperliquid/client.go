package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"deepseek-trader/config"

	"github.com/ethereum/go-ethereum/crypto"
	hl "github.com/sonirico/go-hyperliquid"
)

type Client struct {
	cfg           *config.Settings
	walletAddress string
	httpClient    *http.Client
	ex            *hl.Exchange
}

func NewClient(cfg *config.Settings) *Client {
	lc := &Client{
		cfg:           cfg,
		walletAddress: cfg.APIWallet,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}

	if cfg.APISecret != "" {
		if pk, err := crypto.HexToECDSA(cfg.APISecret); err == nil {
			lc.ex = hl.NewExchange(context.Background(), pk, cfg.HLBaseURL, nil, "", "", nil)
		}
	}

	return lc
}

func (c *Client) GetLiveStats(ctx context.Context) (*LiveStats, error) {
	if c.walletAddress == "" {
		return nil, errors.New("wallet address is required (set API_KEY or API_SECRET)")
	}
	if st, ok := c.tryFetchPortfolio(ctx); ok {
		return st, nil
	}

	return nil, errors.New("failed to fetch live stats")
}

// HistoricalOrders fetches user's historical orders raw list from HL info API.
func (c *Client) HistoricalOrders(ctx context.Context) ([]UserFill, error) {
	url := fmt.Sprintf("%s/info", c.cfg.HLBaseURL)
	payload := Payload{Type: "userFills", User: c.walletAddress}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch historical orders")
	}

	var out []UserFill
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) UserFees(ctx context.Context) (*UserFees, error) {
	if c.walletAddress == "" {
		return nil, errors.New("wallet address is required")
	}
	url := strings.TrimRight(c.cfg.HLBaseURL, "/") + "/info"
	payload := map[string]any{"type": "userFees", "user": c.walletAddress}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch user fees")
	}

	var out UserFeeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	toF := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	fees := &UserFees{
		UserCrossRate:    toF(out.UserCrossRate),
		UserAddRate:      toF(out.UserAddRate),
		ReferralDiscount: toF(out.ActiveReferralDiscount),
	}

	if out.ActiveStakingDiscount != nil {
		fees.StakingActiveDiscount = toF(out.ActiveStakingDiscount.Discount)
	}

	return fees, nil
}

// PlaceOrder: оформляет лимитный ордер GTC
func (c *Client) PlaceOrder(ctx context.Context, symbol, side string, qty, price float64) (string, error) {
	if c.ex == nil {
		return "", errors.New("exchange client not initialized; set API_SECRET")
	}
	coin := normalizeSymbol(symbol)
	isBuy := strings.EqualFold(side, "BUY")
	req := hl.CreateOrderRequest{
		Coin:      coin,
		IsBuy:     isBuy,
		Size:      qty,
		Price:     price,
		OrderType: hl.OrderType{Limit: &hl.LimitOrderType{Tif: "Gtc"}},
	}
	resp, err := c.ex.Order(ctx, req, nil)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(resp.Filled.Oid), nil
}

func (c *Client) SetWalletAddress(addr string) {
	c.walletAddress = strings.TrimSpace(addr)
}

func (c *Client) CoinsMids(ctx context.Context) (map[string]string, error) {
	url := fmt.Sprintf("%s/info", c.cfg.HLBaseURL)
	payload := Payload{Type: "allMids", User: c.walletAddress}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch user fees")
	}

	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}

// Вспомогательная нормализация тикера
func normalizeSymbol(sym string) string {
	s := strings.ToUpper(strings.TrimSpace(sym))
	s = strings.TrimSuffix(s, "USDT")
	return s
}

func (c *Client) tryFetchPortfolio(ctx context.Context) (*LiveStats, bool) {
	raw, ok := c.fetchPortfolioRaw(ctx)
	if !ok {
		return nil, false
	}

	labelTo, ok := c.parsePortfolioResponse(raw)
	if !ok || len(labelTo) == 0 {
		return nil, false
	}

	pick := c.selectPreferredData(labelTo)
	if pick == nil {
		return nil, false
	}

	bal, ok1 := c.getLastValue(pick, "accountValueHistory")
	pnl, ok2 := c.getLastValue(pick, "pnlHistory")
	if !ok1 {
		return nil, false
	}

	roe := 0.0
	if bal != 0 && ok2 {
		roe = (pnl / bal) * 100
	}

	return &LiveStats{Balance: bal, PnL: pnl, ROE: roe}, true
}

// fetchPortfolioRaw выполняет HTTP-запрос и возвращает тело ответа.
func (c *Client) fetchPortfolioRaw(ctx context.Context) ([]byte, bool) {
	url := strings.TrimRight(c.cfg.HLBaseURL, "/") + "/info"
	payload := map[string]any{"type": "portfolio", "user": c.walletAddress}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false
	}

	raw, err := io.ReadAll(resp.Body)
	return raw, err == nil
}

// parsePortfolioResponse парсит JSON и строит мапу label → объект.
func (c *Client) parsePortfolioResponse(raw []byte) (map[string]map[string]any, bool) {
	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, false
	}

	labelTo := make(map[string]map[string]any)
	for _, item := range arr {
		tup, ok := item.([]any)
		if !ok || len(tup) != 2 {
			continue
		}
		lbl, _ := tup[0].(string)
		obj, _ := tup[1].(map[string]any)
		if lbl != "" && obj != nil {
			labelTo[lbl] = obj
		}
	}
	return labelTo, true
}

// selectPreferredData выбирает первый подходящий набор по приоритету.
func (c *Client) selectPreferredData(labelTo map[string]map[string]any) map[string]any {
	prefs := []string{"perpDay", "day", "perpWeek", "week", "perpMonth", "month", "perpAllTime", "allTime"}
	for _, p := range prefs {
		if o, ok := labelTo[p]; ok {
			return o
		}
	}
	// fallback: любой первый
	for _, o := range labelTo {
		return o
	}
	return nil
}

// getLastValue извлекает последнее значение из истории по ключу.
func (c *Client) getLastValue(data map[string]any, key string) (float64, bool) {
	seq, ok := data[key]
	if !ok {
		return 0, false
	}
	list, ok := seq.([]any)
	if !ok || len(list) == 0 {
		return 0, false
	}
	last := list[len(list)-1]
	pair, ok := last.([]any)
	if !ok || len(pair) != 2 {
		return 0, false
	}
	switch v := pair[1].(type) {
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	case float64:
		return v, true
	}
	return 0, false
}
