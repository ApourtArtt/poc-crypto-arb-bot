package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	conn    *pgxpool.Pool
	Queries *Queries
}

func NewDatabase(dbName, username, password string) (*Database, error) {
	conn, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, "127.0.0.1", "5555", dbName))
	if err != nil {
		return nil, err
	}

	return &Database{
		conn:    conn,
		Queries: New(conn),
	}, nil
}

func (d *Database) Close() {
	d.conn.Close()
}
