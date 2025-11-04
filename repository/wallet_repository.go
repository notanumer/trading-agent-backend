package repository

import (
	"context"

	"deepseek-trader/models"
	_ "embed"

	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/wallet/create.sql
	createWalletSQL string
	//go:embed sql/wallet/find_by_user.sql
	findByUserSQL string
	//go:embed sql/wallet/delete_by_user.sql
	deleteByUserSQL string
)

type WalletRepository struct {
	db *sqlx.DB
}

func (r *WalletRepository) Create(ctx context.Context, w *models.Wallet) error {
	return r.db.
		QueryRowxContext(ctx, createWalletSQL, w.Address, w.APIKey, w.UserID).
		Scan(&w.ID, &w.CreatedAt)
}

func (r *WalletRepository) FindLatestByUser(ctx context.Context, userID int64) (models.Wallet, error) {
	var w models.Wallet

	if err := r.db.GetContext(ctx, &w, findByUserSQL, userID); err != nil {
		return models.Wallet{}, err
	}
	return w, nil
}

func (r *WalletRepository) DeleteByUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, deleteByUserSQL, userID)
	return err
}
