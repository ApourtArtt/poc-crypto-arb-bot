// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: 000001.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const insertCoin = `-- name: InsertCoin :one
INSERT INTO "coins" ("name", "base")
VALUES ($1, $2)
ON CONFLICT DO NOTHING
RETURNING id
`

type InsertCoinParams struct {
	Name string `json:"name"`
	Base string `json:"base"`
}

func (q *Queries) InsertCoin(ctx context.Context, arg InsertCoinParams) (uuid.UUID, error) {
	row := q.db.QueryRow(ctx, insertCoin, arg.Name, arg.Base)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const insertCoinExchange = `-- name: InsertCoinExchange :exec
INSERT INTO "exchange_coins" ("coin_id", "exchange_id", "name", "base")
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING
`

type InsertCoinExchangeParams struct {
	CoinID     uuid.UUID `json:"coin_id"`
	ExchangeID uuid.UUID `json:"exchange_id"`
	Name       string    `json:"name"`
	Base       string    `json:"base"`
}

func (q *Queries) InsertCoinExchange(ctx context.Context, arg InsertCoinExchangeParams) error {
	_, err := q.db.Exec(ctx, insertCoinExchange,
		arg.CoinID,
		arg.ExchangeID,
		arg.Name,
		arg.Base,
	)
	return err
}

const insertTicker = `-- name: InsertTicker :exec
INSERT INTO "exchange_tickers" ("exchange_id", "base_exch_coin_id", "quote_exch_coin_id")
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`

type InsertTickerParams struct {
	ExchangeID      uuid.UUID `json:"exchange_id"`
	BaseExchCoinID  uuid.UUID `json:"base_exch_coin_id"`
	QuoteExchCoinID uuid.UUID `json:"quote_exch_coin_id"`
}

func (q *Queries) InsertTicker(ctx context.Context, arg InsertTickerParams) error {
	_, err := q.db.Exec(ctx, insertTicker, arg.ExchangeID, arg.BaseExchCoinID, arg.QuoteExchCoinID)
	return err
}

const selectAllCoins = `-- name: SelectAllCoins :many
SELECT id, name, base FROM coins
`

func (q *Queries) SelectAllCoins(ctx context.Context) ([]Coin, error) {
	rows, err := q.db.Query(ctx, selectAllCoins)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Coin{}
	for rows.Next() {
		var i Coin
		if err := rows.Scan(&i.ID, &i.Name, &i.Base); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectExchangeCoinFromCoinID = `-- name: SelectExchangeCoinFromCoinID :many
SELECT base, e.name FROM exchange_coins ec
JOIN exchanges e ON e.id = ec.exchange_id
WHERE coin_id = $1
`

type SelectExchangeCoinFromCoinIDRow struct {
	Base string `json:"base"`
	Name string `json:"name"`
}

func (q *Queries) SelectExchangeCoinFromCoinID(ctx context.Context, coinID uuid.UUID) ([]SelectExchangeCoinFromCoinIDRow, error) {
	rows, err := q.db.Query(ctx, selectExchangeCoinFromCoinID, coinID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SelectExchangeCoinFromCoinIDRow{}
	for rows.Next() {
		var i SelectExchangeCoinFromCoinIDRow
		if err := rows.Scan(&i.Base, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectExchangeCoinIDFromBase = `-- name: SelectExchangeCoinIDFromBase :one
SELECT id FROM exchange_coins
WHERE base = $1 AND exchange_id = $2
LIMIT 1
`

type SelectExchangeCoinIDFromBaseParams struct {
	Base       string    `json:"base"`
	ExchangeID uuid.UUID `json:"exchange_id"`
}

func (q *Queries) SelectExchangeCoinIDFromBase(ctx context.Context, arg SelectExchangeCoinIDFromBaseParams) (uuid.UUID, error) {
	row := q.db.QueryRow(ctx, selectExchangeCoinIDFromBase, arg.Base, arg.ExchangeID)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const selectExchangeCoins = `-- name: SelectExchangeCoins :many
SELECT ec.id, ec.coin_id, ec.exchange_id, ec.name, ec.base, e.name AS exchange_name
FROM "exchange_coins" ec
LEFT JOIN "exchanges" e ON ec.exchange_id = e.id
`

type SelectExchangeCoinsRow struct {
	ID           uuid.UUID   `json:"id"`
	CoinID       uuid.UUID   `json:"coin_id"`
	ExchangeID   uuid.UUID   `json:"exchange_id"`
	Name         string      `json:"name"`
	Base         string      `json:"base"`
	ExchangeName pgtype.Text `json:"exchange_name"`
}

func (q *Queries) SelectExchangeCoins(ctx context.Context) ([]SelectExchangeCoinsRow, error) {
	rows, err := q.db.Query(ctx, selectExchangeCoins)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SelectExchangeCoinsRow{}
	for rows.Next() {
		var i SelectExchangeCoinsRow
		if err := rows.Scan(
			&i.ID,
			&i.CoinID,
			&i.ExchangeID,
			&i.Name,
			&i.Base,
			&i.ExchangeName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectExchangeTickers = `-- name: SelectExchangeTickers :many
SELECT et.id, e.name as exchange_name, et.base_exch_coin_id, et.quote_exch_coin_id, ec1.base as base, ec2.base as quote
FROM "exchange_tickers" et
JOIN "exchange_coins" ec1 ON ec1.exchange_id = et.exchange_id AND ec1.id = et.base_exch_coin_id
JOIN "exchange_coins" ec2 ON ec2.exchange_id = et.exchange_id AND ec2.id = et.quote_exch_coin_id
JOIN "exchanges" e ON e.id = et.exchange_id
`

type SelectExchangeTickersRow struct {
	ID              uuid.UUID `json:"id"`
	ExchangeName    string    `json:"exchange_name"`
	BaseExchCoinID  uuid.UUID `json:"base_exch_coin_id"`
	QuoteExchCoinID uuid.UUID `json:"quote_exch_coin_id"`
	Base            string    `json:"base"`
	Quote           string    `json:"quote"`
}

func (q *Queries) SelectExchangeTickers(ctx context.Context) ([]SelectExchangeTickersRow, error) {
	rows, err := q.db.Query(ctx, selectExchangeTickers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SelectExchangeTickersRow{}
	for rows.Next() {
		var i SelectExchangeTickersRow
		if err := rows.Scan(
			&i.ID,
			&i.ExchangeName,
			&i.BaseExchCoinID,
			&i.QuoteExchCoinID,
			&i.Base,
			&i.Quote,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectExchanges = `-- name: SelectExchanges :many
SELECT id, name, url, ticker_format FROM exchanges
`

func (q *Queries) SelectExchanges(ctx context.Context) ([]Exchange, error) {
	rows, err := q.db.Query(ctx, selectExchanges)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Exchange{}
	for rows.Next() {
		var i Exchange
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Url,
			&i.TickerFormat,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
