package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Handle interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Store struct {
	*Queries
	db Handle
}

func NewStore(db Handle) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (s *Store) BeginFunc(ctx context.Context, fn func(s *Store) error) error {
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		return fn(&Store{
			Queries: New(tx),
			db:      tx,
		})
	})
}
