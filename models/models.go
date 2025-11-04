package models

import "time"

type Wallet struct {
    ID        int64     `db:"id" json:"id"`
    Address   string    `db:"address" json:"address"`
    APIKey    string    `db:"api_key" json:"-"`
    UserID    *int64    `db:"user_id" json:"userId,omitempty"`
    CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Trade struct {
    ID        int64     `db:"id" json:"id"`
    Symbol    string    `db:"symbol" json:"symbol"`
    Side      string    `db:"side" json:"side"`
    Qty       float64   `db:"qty" json:"qty"`
    Price     float64   `db:"price" json:"price"`
    PnL       float64   `db:"pnl" json:"pnl"`
    CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Stats struct {
    ID        int64     `db:"id" json:"id"`
    Balance   float64   `db:"balance" json:"balance"`
    PnL       float64   `db:"pnl" json:"pnl"`
    ROE       float64   `db:"roe" json:"roe"`
    CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type User struct {
    ID           int64     `db:"id" json:"id"`
    Email        string    `db:"email" json:"email"`
    PasswordHash string    `db:"password_hash" json:"-"`
    CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}


