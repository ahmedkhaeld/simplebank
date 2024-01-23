package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomString generates a random string of length n using the characters in the `alphabet` constant.
func RandomString(n int) string {
	// sb is the string builder that will hold the random string.
	var sb strings.Builder

	// k is the length of the alphabet constant.
	k := len(alphabet)

	// Loop n times to generate the random string.
	for i := 0; i < n; i++ {
		// Pick a random index in the alphabet constant using rand.Intn.
		randIndex := rand.Intn(k)

		// Append the character at the picked index to the string builder.
		sb.WriteByte(alphabet[randIndex])
	}

	// Return the generated random string.
	return sb.String()
}

// RandomInt generates a random int64 number between min (inclusive) and max (exclusive).
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomAccountOwner generates a random string of length 6 as an account owner.
func RandomAccountOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random int64 number between min (inclusive) and max (exclusive).
func RandomMoney() int64 {
	return RandomInt(0, 10000)
}

// RandomAccountCurrency generates a random currency code for an account.
func RandomAccountCurrency() string {
	// currencies is a slice of strings containing the possible currency codes.
	currencies := []string{"USD", "EUR", "GBP", "EGP"}

	// n is the length of the currencies slice.
	n := len(currencies)

	// rand.Intn returns a random integer in the range [0, n-1].
	return currencies[rand.Intn(n)]
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
