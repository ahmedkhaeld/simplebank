package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	dbSource = "postgresql://root:password@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries

// TestMain sets up the test environment by creating a database connection and
// populating the testQueries instance variable with a new instance of Queries.
// It then invokes the m.Run method to run the tests and returns the exit code.
func TestMain(m *testing.M) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbSource)
	if err != nil {
		log.Fatal("Failed to connect", err)
	}
	defer conn.Close(ctx)
	testQueries = New(conn)

	os.Exit(m.Run())

}
