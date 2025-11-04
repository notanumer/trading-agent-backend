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
