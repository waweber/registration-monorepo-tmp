package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBSession struct {
	pgx.Tx
}

func NewDBSession(context context.Context, pool *pgxpool.Pool) (DBSession, error) {
	tx, err := pool.Begin(context)
	return DBSession{Tx: tx}, err
}
