package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store represents a database connection to a database server
// and a mock database for testing purposes
type Store interface {
	Querier
	TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, args CreateUserTxParams) (CreateuserTxResult, error)
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		Queries: New(db),
		db:      db,
	}
}

// execTransaction a wrapper that executes the given function within a transaction.
func (s *SQLStore) execTransaction(ctx context.Context, fn func(*Queries) error) error {
	// Begin a transaction.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Create a new instance of the Queries struct using the transaction.
	q := New(tx)

	// Execute the given function within the transaction.
	err = fn(q)

	if err != nil {
		// Rollback the transaction if an error occurs.
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	// Commit the transaction.
	return tx.Commit()
}
