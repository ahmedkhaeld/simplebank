package token

import (
	"testing"
	"time"

	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/golang-jwt/jwt/v5"

	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {

	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomAccountOwner()
	duration := time.Minute

	isssuedAt := time.Now()
	expiredAt := isssuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	sub, _ := payload.GetSubject()
	expiresAt, _ := payload.GetExpirationTime()

	require.Equal(t, username, sub)
	require.WithinDuration(t, expiredAt, expiresAt.Time, time.Second)

}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomAccountOwner()
	duration := time.Minute

	token, err := maker.CreateToken(username, -duration)
	require.NoError(t, err)
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
}

func TestInvalidJWTAlgNone(t *testing.T) {
	payload, err := NewPayload(util.RandomAccountOwner(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	tokenString, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(tokenString)
	require.Error(t, err)
	require.Nil(t, payload)
	require.EqualError(t, err, ErrUnexpectedMethod.Error())

}
