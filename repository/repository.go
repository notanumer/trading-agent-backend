package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Wallets *WalletRepository
	Trades  *TradeRepository
	Stats   *StatsRepository
	Users   *UserRepository
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Wallets: &WalletRepository{db: db},
		Trades:  &TradeRepository{db: db},
		Stats:   &StatsRepository{db: db},
		Users:   &UserRepository{db: db},
	}
}
