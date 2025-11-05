INSERT INTO decisions (
    action, symbol, size, order_type, limit_price, tp1, tp2, tp3, sl
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at;


