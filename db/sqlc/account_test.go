package db

import (
	"context"
	"testing"
	"time"

	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createAccountHelper(t *testing.T) Account {
	ctx := context.Background()
	arg := CreateAccountParams{
		Owner:    util.RandomAccountOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomAccountCurrency(),
	}
	account, err := testQueries.CreateAccount(ctx, arg)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createAccountHelper(t)
}

func TestGetAccount(t *testing.T) {
	createdAccount := createAccountHelper(t)

	ctx := context.Background()
	fetchedAccount, err := testQueries.GetAccount(ctx, createdAccount.ID)

	require.NoError(t, err)
	require.NotNil(t, fetchedAccount)
	require.Equal(t, createdAccount.ID, fetchedAccount.ID)
	require.Equal(t, createdAccount.Owner, fetchedAccount.Owner)
	require.Equal(t, createdAccount.Balance, fetchedAccount.Balance)
	require.Equal(t, createdAccount.Currency, fetchedAccount.Currency)
	require.WithinDuration(t, createdAccount.CreatedAt.Time, fetchedAccount.CreatedAt.Time, time.Second)

}

func TestUpdateAccount(t *testing.T) {
	createdAccount := createAccountHelper(t)

	ctx := context.Background()
	arg := UpdateAccountParams{
		ID:      createdAccount.ID,
		Balance: util.RandomMoney(),
	}
	updatedAccount, err := testQueries.UpdateAccount(ctx, arg)
	require.NoError(t, err)
	require.NotNil(t, updatedAccount)

	require.Equal(t, createdAccount.ID, updatedAccount.ID)
	require.Equal(t, createdAccount.Owner, updatedAccount.Owner)
	require.Equal(t, arg.Balance, updatedAccount.Balance)
	require.Equal(t, createdAccount.Currency, updatedAccount.Currency)
	require.WithinDuration(t, createdAccount.CreatedAt.Time, updatedAccount.CreatedAt.Time, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	createdAccount := createAccountHelper(t)

	ctx := context.Background()
	err := testQueries.DeleteAccount(ctx, createdAccount.ID)
	require.NoError(t, err)

	isAccountExist, err := testQueries.GetAccount(ctx, createdAccount.ID)
	require.Error(t, err)

	require.Empty(t, isAccountExist)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createAccountHelper(t)
	}
	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
