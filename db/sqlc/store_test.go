package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(conn)

	// prepare two  mock accounts records in the Accounts table, which the transfer will occur between them.
	// each Account Record should ID, Balance
	beforeFromAccountRecord := createAccountHelper(t)
	beforeToAccountRecord := createAccountHelper(t)

	// prepare amount of money, The two records will exchange
	amount := int64(100)

	fmt.Println("Befor Transaction occurred")
	fmt.Printf("Account: %v send an amount of $%v to Account %v\n", beforeFromAccountRecord.Owner, amount, beforeToAccountRecord.Owner)
	fmt.Printf("Sender Balance: $%v Receiver Balance  $%v\n", beforeFromAccountRecord.Balance, beforeToAccountRecord.Balance)

	fmt.Printf("After the Tranfer occurs\n")
	expectedFrom := beforeFromAccountRecord.Balance - amount
	expectedTo := beforeToAccountRecord.Balance + amount

	fmt.Printf("Sender Expected Balance: $%v , Receiver Expected Balance: $%v\n", expectedFrom, expectedTo)

	/*
	 TransferResultRecord returned contains
	 - Transfer
	 - FromAccount
	 - ToAccount
	 - FromEntry
	 - ToEntry
	*/
	transferResultRecord, err := store.TransferTx(context.Background(), TransferTxParams{
		FromAccountID: beforeFromAccountRecord.ID,
		ToAccountID:   beforeToAccountRecord.ID,
		Amount:        amount,
	})

	require.NoError(t, err)
	require.NotEmpty(t, transferResultRecord)

	/* 1. Check Transfer integrity
	********************************
	* transfer is not empty
	* transfer.ID          is not zero
	* transfer.CreatedAt   is not zero
	* transfer.Amount      match the amount
	* transfer.FromAccount match the beforeFromAccountRecord
	* transfer.ToAccount   match the beforeToAccountRecord
	 */
	transfer := transferResultRecord.Transfer
	require.NotEmpty(t, transfer)
	require.NotZero(t, transfer.ID)
	require.Equal(t, amount, transfer.Amount)
	require.NotZero(t, transfer.CreatedAt)
	require.Equal(t, beforeFromAccountRecord.ID, transfer.FromAccountID)
	require.Equal(t, beforeToAccountRecord.ID, transfer.ToAccountID)
	// now check if we Call db to get the transfer record no error
	_, err = store.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	/* 2. Check the From Entry integrity
	***********************************
	* from entry             is not empty
	* from entry.ID          is not zero
	* from entry.CreatedAt   is not zero
	* from -ve entry.Amount      match the amount
	* from entry.FromAccount match the beforeFromAccountRecord
	 */
	fromEntry := transferResultRecord.FromEntry
	require.NotEmpty(t, fromEntry)
	require.NotZero(t, fromEntry.ID)
	require.NotZero(t, fromEntry.CreatedAt)
	require.Equal(t, -amount, fromEntry.Amount)
	require.Equal(t, beforeFromAccountRecord.ID, fromEntry.AccountID)
	_, err = store.GetEntry(context.Background(), fromEntry.ID)
	require.NoError(t, err)

	/* 3. Check the To Entry integrity
	***********************************
	* to entry             is not empty
	* to entry.ID          is not zero
	* to entry.CreatedAt   is not zero
	* to entry.Amount      match the amount
	* to entry.ToAccount match the beforeToAccountRecord
	 */
	toEntry := transferResultRecord.ToEntry
	require.NotEmpty(t, toEntry)
	require.NotZero(t, toEntry.ID)
	require.NotZero(t, toEntry.CreatedAt)
	require.Equal(t, amount, toEntry.Amount)
	require.Equal(t, beforeToAccountRecord.ID, toEntry.AccountID)
	_, err = store.GetEntry(context.Background(), toEntry.ID)
	require.NoError(t, err)

	/* 4. Check the From Account integrity
	 ***********************************
	 * from account             is not empty
	 * from account.ID          is not zero
	 * from account.ID    match the beforeFromAccountRecord.ID
	 */
	fromAccount := transferResultRecord.FromAccount
	require.NotEmpty(t, fromAccount)
	require.NotZero(t, fromAccount.ID)
	require.Equal(t, beforeFromAccountRecord.ID, fromAccount.ID)
	fmt.Print("********************************Money Sent ********************************\n")
	fmt.Printf("From Account Balance: %v\n", fromAccount.Balance)

	/* 5. Check the To Account
	********************************
	* to account             is not empty
	* to account.ID          is not zero
	* to account.ID    match the beforeToAccountRecord.ID
	 */
	toAccount := transferResultRecord.ToAccount
	require.NotEmpty(t, toAccount)
	require.NotZero(t, toAccount.ID)
	require.Equal(t, beforeToAccountRecord.ID, toAccount.ID)
	fmt.Printf("To Account Balance: %v\n", toAccount.Balance)

	// Check the acounts' balance:
	// calculate the balance change within the Transaction,  the outgoing amount, the incoming amount
	// outgoing amount should be equal
	outgoingAmount := beforeFromAccountRecord.Balance - fromAccount.Balance
	incomingAmount := toAccount.Balance - beforeToAccountRecord.Balance
	require.Equal(t, outgoingAmount, incomingAmount)
	require.True(t, outgoingAmount > 0)
	require.True(t, outgoingAmount%amount == 0)

	k := int64(outgoingAmount / amount)
	require.True(t, k >= 1 && k <= amount)

	// get the updated account from the database
	afterFromAccountRecord, err := store.GetAccount(context.Background(), beforeFromAccountRecord.ID)
	require.NoError(t, err)
	afterToAccountRecord, err := store.GetAccount(context.Background(), beforeToAccountRecord.ID)
	require.NoError(t, err)

	fmt.Println(">>> after transfer", "from", afterFromAccountRecord.Balance, "to", afterToAccountRecord.Balance)

	// after number of transactions
	// balance of FromAccount should be decreased by  amount
	// balance of ToAccount should be increased by amount
	require.Equal(t, beforeFromAccountRecord.Balance-amount, afterFromAccountRecord.Balance)
	require.Equal(t, beforeToAccountRecord.Balance+amount, afterToAccountRecord.Balance)

}

// TestTransferTx runs n concurrent transfer transactions
func TestConcurrentTransferTx(t *testing.T) {
	store := NewStore(conn)

	// prepare two  mock accounts records in the Accounts table, which the transfer will occur between them.
	// each Account Record should ID, Balance
	beforeFromAccountRecord := createAccountHelper(t)
	beforeToAccountRecord := createAccountHelper(t)

	// prepare amount of money, The two records will exchange
	amount := int64(100)
	n := 5 // how many times to transfer certain amount of money

	fmt.Println("Befor Transactions occurred")
	fmt.Printf("Account: %v send Total amount of $%v*%v to Account %v\n", beforeFromAccountRecord.Owner, amount, n, beforeToAccountRecord.Owner)
	fmt.Printf("Sender Balance: $%v Receiver Balance  $%v\n", beforeFromAccountRecord.Balance, beforeToAccountRecord.Balance)

	fmt.Printf("After the Tranfer occurs for #%v\n", n)
	expectedFrom := beforeFromAccountRecord.Balance - amount*int64(n)
	expectedTo := beforeToAccountRecord.Balance + amount*int64(n)

	fmt.Printf("Sender Expected Balance: $%v , Receiver Expected Balance: $%v\n", expectedFrom, expectedTo)

	// create channels for errors and results
	chErrs := make(chan error)
	chResults := make(chan TransferTxResult)

	// run n concurrent transfer transactions
	for i := 0; i < n; i++ {
		go func() {
			ctx := context.Background()
			// run transfer transaction
			transferResultRecord, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: beforeFromAccountRecord.ID,
				ToAccountID:   beforeToAccountRecord.ID,
				Amount:        amount,
			})
			chErrs <- err
			chResults <- transferResultRecord
		}()
	}

	// existed track the [n] transfer operation  status, true occured, false didn't occur
	// capture the status of the transfer operation with its index will help to ensure eac go routine completes one time
	existed := make(map[int64]bool)

	// collect results and check for errors
	for i := 0; i < n; i++ {

		// get error from channel
		err := <-chErrs
		require.NoError(t, err)

		// get result from channel
		transferResultRecord := <-chResults
		require.NotEmpty(t, transferResultRecord)

		/* 1. Check Transfer integrity
		********************************
		* transfer is not empty
		* transfer.ID          is not zero
		* transfer.CreatedAt   is not zero
		* transfer.Amount      match the amount
		* transfer.FromAccount match the beforeFromAccountRecord
		* transfer.ToAccount   match the beforeToAccountRecord
		 */
		transfer := transferResultRecord.Transfer
		require.NotEmpty(t, transfer)
		require.NotZero(t, transfer.ID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.CreatedAt)
		require.Equal(t, beforeFromAccountRecord.ID, transfer.FromAccountID)
		require.Equal(t, beforeToAccountRecord.ID, transfer.ToAccountID)
		// now check if we Call db to get the transfer record no error
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		/* 2. Check the From Entry integrity
		***********************************
		* from entry             is not empty
		* from entry.ID          is not zero
		* from entry.CreatedAt   is not zero
		* from -ve entry.Amount      match the amount
		* from entry.FromAccount match the beforeFromAccountRecord
		 */
		fromEntry := transferResultRecord.FromEntry
		require.NotEmpty(t, fromEntry)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		require.Equal(t, -amount, fromEntry.Amount)
		require.Equal(t, beforeFromAccountRecord.ID, fromEntry.AccountID)
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		/* 3. Check the To Entry integrity
		***********************************
		* to entry             is not empty
		* to entry.ID          is not zero
		* to entry.CreatedAt   is not zero
		* to entry.Amount      match the amount
		* to entry.ToAccount match the beforeToAccountRecord
		 */
		toEntry := transferResultRecord.ToEntry
		require.NotEmpty(t, toEntry)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		require.Equal(t, amount, toEntry.Amount)
		require.Equal(t, beforeToAccountRecord.ID, toEntry.AccountID)
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		/* 4. Check the From Account integrity
		 ***********************************
		 * from account             is not empty
		 * from account.ID          is not zero
		 * from account.ID    match the beforeFromAccountRecord.ID
		 */
		fromAccount := transferResultRecord.FromAccount
		require.NotEmpty(t, fromAccount)
		require.NotZero(t, fromAccount.ID)
		require.Equal(t, beforeFromAccountRecord.ID, fromAccount.ID)

		/* 5. Check the To Account
		********************************
		* to account             is not empty
		* to account.ID          is not zero
		* to account.ID    match the beforeToAccountRecord.ID
		 */
		toAccount := transferResultRecord.ToAccount
		require.NotEmpty(t, toAccount)
		require.NotZero(t, toAccount.ID)
		require.Equal(t, beforeToAccountRecord.ID, toAccount.ID)

		// Check the acounts' balance:
		// calculate the balance change within the Transaction,  the outgoing amount, the incoming amount
		// outgoing amount should be equal
		outgoingAmount := beforeFromAccountRecord.Balance - fromAccount.Balance
		incomingAmount := toAccount.Balance - beforeToAccountRecord.Balance
		require.Equal(t, outgoingAmount, incomingAmount)
		require.Equal(t, outgoingAmount, incomingAmount)
		require.True(t, outgoingAmount > 0)
		require.True(t, outgoingAmount%amount == 0)

		// as each transfer operation identified with an index from the n
		// we should ensure that index has not been done before through the concurrent routines
		//If a duplicate k is encountered, it would indicate a potential issue in transfer logic.
		k := int64(outgoingAmount / amount)      // 100/100=1; 200/100=2; 300/100=3; 400/100=4; 500/100=5
		require.True(t, k >= 1 && k <= int64(n)) // ensure each k  equals to n
		require.NotContains(t, existed, k)       //Tracks seen k values in the existed map to identify transfer o duplicates.
		existed[k] = true

	}

	// get the updated account from the database
	afterFromAccountRecord, err := store.GetAccount(context.Background(), beforeFromAccountRecord.ID)
	require.NoError(t, err)
	afterToAccountRecord, err := store.GetAccount(context.Background(), beforeToAccountRecord.ID)
	require.NoError(t, err)

	fmt.Println(">>> after transfer", "from", afterFromAccountRecord.Balance, "to", afterToAccountRecord.Balance)

	// after number of transactions
	// balance of FromAccount should be decreased by  amount * number of transfers completed
	// balance of ToAccount should be increased by amount * number of transfers completed
	require.Equal(t, beforeFromAccountRecord.Balance-amount*int64(n), afterFromAccountRecord.Balance)
	require.Equal(t, beforeToAccountRecord.Balance+amount*int64(n), afterToAccountRecord.Balance)

}
