package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dbSource = "postgresql://root:password@localhost:5432/simple_bank?sslmode=disable"
	dbDriver = "pgx"
)

var testQueries *Queries
var conn *sql.DB

// TestMain sets up the test environment by creating a database connection and
// populating the testQueries instance variable with a new instance of Queries.
// It then invokes the m.Run method to run the tests and returns the exit code.
func TestMain(m *testing.M) {
	var err error
	conn, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Failed to connect", err)
	}
	testQueries = New(conn)

	os.Exit(m.Run())

}
