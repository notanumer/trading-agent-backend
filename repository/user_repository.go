package repository

import (
	"context"
	"deepseek-trader/models"
	_ "embed"

	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/user/create.sql
	createUserSQL string

	//go:embed sql/user/find_by_email.sql
	findByEmailSQL string
)

type UserRepository struct{ db *sqlx.DB }

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	return r.db.QueryRowxContext(ctx, createUserSQL, u.Email, u.PasswordHash).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User

	if err := r.db.GetContext(ctx, &u, findByEmailSQL, email); err != nil {
		return models.User{}, err
	}

	return u, nil
}
