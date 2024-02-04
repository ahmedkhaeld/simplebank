package db

import (
	"context"
	"testing"
	"time"

	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createUserHelper(t *testing.T) User {
	hashPass, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	ctx := context.Background()

	arg := CreateUserParams{
		Email:    util.RandomEmail(),
		Password: hashPass,
		Username: util.RandomAccountOwner(),
		FullName: util.RandomAccountOwner(),
	}
	user, err := testQueries.CreateUser(ctx, arg)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.FullName, user.FullName)
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user

}

func TestUserAccount(t *testing.T) {
	createUserHelper(t)
}

func TestGetUser(t *testing.T) {
	createdUser := createUserHelper(t)

	ctx := context.Background()
	fetchedUser, err := testQueries.GetUser(ctx, createdUser.Username)
	require.NoError(t, err)
	require.NotNil(t, fetchedUser)
	require.Equal(t, createdUser.Username, fetchedUser.Username)
	require.Equal(t, createdUser.Email, fetchedUser.Email)
	require.Equal(t, createdUser.FullName, fetchedUser.FullName)
	require.Equal(t, createdUser.Password, fetchedUser.Password)
	require.WithinDuration(t, createdUser.CreatedAt, fetchedUser.CreatedAt, time.Second)
	require.WithinDuration(t, createdUser.PasswordChangedAt, fetchedUser.PasswordChangedAt, time.Second)

}
