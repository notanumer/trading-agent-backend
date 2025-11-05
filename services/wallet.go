package services

import (
	"context"
	"errors"

	"deepseek-trader/config"
	"deepseek-trader/hyperliquid"
	"deepseek-trader/models"
	cryptoutil "deepseek-trader/pkg/crypto"
	"deepseek-trader/repository"
)

type WalletService struct {
	repo *repository.WalletRepository
	hl   *hyperliquid.Client
	cfg  *config.Settings
}

func NewWalletService(repo *repository.WalletRepository, hl *hyperliquid.Client, cfg *config.Settings) *WalletService {
	return &WalletService{repo: repo, hl: hl, cfg: cfg}
}

type ConnectRequest struct {
	Address string `json:"address"`
	APIKey  string `json:"api_key"`
}

func (s *WalletService) Connect(ctx context.Context, req ConnectRequest, userID int64) (*models.Wallet, error) {
	if req.Address == "" {
		return nil, errors.New("address required")
	}

	// Try to validate by fetching balance (sandbox/live)
	// Update runtime wallet address for live client
	s.hl.SetWalletAddress(req.Address)

	encrypted := ""
	if req.APIKey != "" {
		encKey := []byte(s.cfg.SecretKey)
		var err error
		encrypted, err = cryptoutil.EncryptString(encKey, req.APIKey)
		if err != nil {
			return nil, err
		}
	}

	w := &models.Wallet{Address: req.Address, APIKey: encrypted, UserID: &userID}
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	w.APIKey = "" // don't return secret
	return w, nil
}

func (s *WalletService) FindLatestByUser(ctx context.Context, userID int64) (models.Wallet, error) {
	return s.repo.FindLatestByUser(ctx, userID)
}

func (s *WalletService) Disconnect(ctx context.Context, userID int64) error {
	s.hl.SetWalletAddress("")
	return s.repo.DeleteByUser(ctx, userID)
}
