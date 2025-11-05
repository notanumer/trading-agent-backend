-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS decisions (
    id SERIAL PRIMARY KEY,
    action TEXT NOT NULL,
    symbol TEXT NOT NULL,
    size NUMERIC NOT NULL,
    order_type TEXT NOT NULL,
    limit_price NUMERIC NOT NULL,
    tp1 NUMERIC NOT NULL,
    tp2 NUMERIC NOT NULL,
    tp3 NUMERIC NOT NULL,
    sl NUMERIC NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS decisions;
-- +goose StatementEnd
