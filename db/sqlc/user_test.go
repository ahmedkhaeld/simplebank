package db

import (
	"context"
	"database/sql"
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

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := createUserHelper(t)
	newFullName := util.RandomAccountOwner()
	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.Password, updatedUser.Password)

}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createUserHelper(t)
	newEmail := util.RandomEmail()

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Password, updatedUser.Password)

}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createUserHelper(t)

	newPassword := util.RandomString(6)
	newHash, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Password: sql.NullString{
			String: newHash,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, newHash, oldUser.Password)
	require.Equal(t, newHash, updatedUser.Password)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)

}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createUserHelper(t)

	newFullName := util.RandomAccountOwner()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)
	newHash, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
		Password: sql.NullString{
			String: newHash,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.NotEqual(t, oldUser.Password, updatedUser.Password)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newHash, updatedUser.Password)

}
