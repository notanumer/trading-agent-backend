select 
    id,
    action,
    symbol, 
    size, 
    order_type, 
    limit_price, 
    tp1, 
    tp2, 
    tp3, 
    sl, 
    created_at 
from decisions 
order by id desc 
limit $1;