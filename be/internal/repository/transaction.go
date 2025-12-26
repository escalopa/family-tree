package repository

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const txKey contextKey = "tx"

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type TransactionManager struct {
	db *pgxpool.Pool
}

func NewTransactionManager(db *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return doWithQuerier(ctx, tm.db, fn)
}

func doWithQuerier(ctx context.Context, db *pgxpool.Pool, fn func(ctx context.Context) error) error {
	querier := getQuerier(ctx, nil) // nil means not in a transaction
	if querier != nil {
		return fn(ctx)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	defer tx.Rollback(ctx)

	txCtx := setQuerier(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.NewDatabaseError(err)
	}

	return nil
}

func getQuerier(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}

func setQuerier(ctx context.Context, tx Querier) context.Context {
	return context.WithValue(ctx, txKey, tx)
}
