package token

import (
	"testing"
	"time"

	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {

	maker, err := NewPasetoMaker(util.RandomString(32))
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
	expireTime, _ := payload.GetExpirationTime()

	require.Equal(t, username, sub)
	require.WithinDuration(t, expiredAt, expireTime.Time, time.Second)

}
