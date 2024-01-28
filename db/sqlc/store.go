package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

// execTransaction a wrapper that executes the given function within a transaction.
func (s *Store) execTransaction(ctx context.Context, fn func(*Queries) error) error {
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

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another.
//
// 1- create a transfer record for from account, to account, amount of being transferred
//
// 2- create two entries one for from account with balance decreased, and for to account with balance increased
//
// 3. update the from account balance, and update the to account balance
func (s *Store) TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTransaction(ctx, func(q *Queries) error {
		var err error
		// Convert params to CreateTransferParams
		createParams := CreateTransferParams(args) // Direct conversion

		result.Transfer, err = q.CreateTransfer(ctx, createParams)
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.FromAccountID,
			Amount:    -args.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.ToAccountID,
			Amount:    args.Amount,
		})
		if err != nil {
			return err
		}

		//move money out from the FromAccount
		result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			Amount: -args.Amount,
			ID:     args.FromAccountID,
		})
		if err != nil {
			return err
		}
		//move money in to the ToAccount
		result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			Amount: args.Amount,
			ID:     args.ToAccountID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
