package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/ahmedkhaeld/simplebank/util"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var testQueries *Queries
var conn *sql.DB

// TestMain sets up the test environment by creating a database connection and
// populating the testQueries instance variable with a new instance of Queries.
// It then invokes the m.Run method to run the tests and returns the exit code.
func TestMain(m *testing.M) {
	var err error
	env, err := util.LoadEnv("../..")
	if err != nil {
		log.Fatal(err)
	}
	conn, err = sql.Open(env.DBDriver, env.DBSource)
	if err != nil {
		log.Fatal("Failed to connect", err)
	}
	testQueries = New(conn)

	os.Exit(m.Run())

}
