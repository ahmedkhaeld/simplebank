package db

import "context"

// CreateUserTxParams Packages user creation parameters and post-creation actions into a CreateUserTxParams structure.
type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error // executed after user is created inside the same transaction, error decides whether to commit or rollback the tx
}

type CreateuserTxResult struct {
	User User
}

// CreateUserTx
//
// Returns an error if any step fails, causing transaction rollback.
// Within the execTransaction function,
// the CreateUser function is called creates a new user in the db and returns a User;
// AfterCreate function is called with the result of the CreateUser call as an argument.
// This function can be used to perform additional operations after the user is created,
// such as sending an email to the user.
func (s *SQLStore) CreateUserTx(ctx context.Context, args CreateUserTxParams) (CreateuserTxResult, error) {
	var result CreateuserTxResult

	err := s.execTransaction(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, args.CreateUserParams)
		if err != nil {
			return err
		}
		return args.AfterCreate(result.User)
	})

	return result, err
}
