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
func (s *SQLStore) TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTransaction(ctx, func(q *Queries) error {
		var err error
		// createTransferAndEntries creates a transfer record and two entries for the transfer.
		if err = s.createTransferAndEntries(ctx, q, &result, args); err != nil {
			return err
		}

		// updateAccountBalances updates the balances of the from and to accounts based on the transfer parameters.
		if args.FromAccountID < args.ToAccountID {
			// If the from account ID is less than the to account ID, a transfer is made from the from account to the to account.
			err = s.updateAccountBalances(ctx, q, &result, args.FromAccountID, -args.Amount, args.ToAccountID, args.Amount)
			if err != nil {
				return err
			}
		} else {
			// If the from account ID is greater than the to account ID, a transfer is made from the to account to the from account.
			err = s.updateAccountBalances(ctx, q, &result, args.ToAccountID, args.Amount, args.FromAccountID, -args.Amount)
			if err != nil {
				return err
			}
		}

		return err
	})

	return result, err
}

func (s *SQLStore) createTransferAndEntries(ctx context.Context, q *Queries, result *TransferTxResult, args TransferTxParams) error {
	var err error

	createParams := CreateTransferParams(args)
	result.Transfer, err = q.CreateTransfer(ctx, createParams)
	if err != nil {
		return err
	}

	result.FromEntry, err = s.createEntry(ctx, q, args.FromAccountID, -args.Amount)
	if err != nil {
		return err
	}

	result.ToEntry, err = s.createEntry(ctx, q, args.ToAccountID, args.Amount)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLStore) createEntry(ctx context.Context, q *Queries, accountID int64, amount int64) (Entry, error) {
	return q.CreateEntry(ctx, CreateEntryParams{
		AccountID: accountID,
		Amount:    amount,
	})
}

func (s *SQLStore) updateAccountBalances(ctx context.Context, q *Queries, result *TransferTxResult, fromAccountID, fromAmount, toAccountID, toAmount int64) error {
	var err error

	result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
		Amount: fromAmount,
		ID:     fromAccountID,
	})
	if err != nil {
		return err
	}

	result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
		Amount: toAmount,
		ID:     toAccountID,
	})
	if err != nil {
		return err
	}

	return nil
}
