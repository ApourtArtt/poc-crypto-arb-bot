-- name: InsertCoin :one
INSERT INTO "coins" ("name", "base")
VALUES ($1, $2)
ON CONFLICT DO NOTHING
RETURNING id;

-- name: SelectExchanges :many
SELECT * FROM exchanges;

-- name: SelectAllCoins :many
SELECT * FROM coins;

-- name: InsertCoinExchange :exec
INSERT INTO "exchange_coins" ("coin_id", "exchange_id", "name", "base")
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: SelectExchangeCoinFromCoinID :many
SELECT base, e.name FROM exchange_coins ec
JOIN exchanges e ON e.id = ec.exchange_id
WHERE coin_id = $1;

-- name: InsertTicker :exec
INSERT INTO "exchange_tickers" ("exchange_id", "base_exch_coin_id", "quote_exch_coin_id")
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: SelectExchangeCoinIDFromBase :one
SELECT id FROM exchange_coins
WHERE base = $1 AND exchange_id = $2
LIMIT 1;

-- name: SelectExchangeTickers :many
SELECT et.id, e.name as exchange_name, et.base_exch_coin_id, et.quote_exch_coin_id, ec1.base as base, ec2.base as quote
FROM "exchange_tickers" et
JOIN "exchange_coins" ec1 ON ec1.exchange_id = et.exchange_id AND ec1.id = et.base_exch_coin_id
JOIN "exchange_coins" ec2 ON ec2.exchange_id = et.exchange_id AND ec2.id = et.quote_exch_coin_id
JOIN "exchanges" e ON e.id = et.exchange_id;;

-- name: SelectExchangeCoins :many
SELECT ec.id, ec.coin_id, ec.exchange_id, ec.name, ec.base, e.name AS exchange_name
FROM "exchange_coins" ec
LEFT JOIN "exchanges" e ON ec.exchange_id = e.id;