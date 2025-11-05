package hyperliquid

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
	Type string `json:"type"`
	User string `json:"user"`
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
