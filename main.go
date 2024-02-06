package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ahmedkhaeld/simplebank/api"
	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/util"
)

var Env = util.Env{}

func init() {
	var err error
	Env, err = util.LoadEnv(".")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	conn, err := sql.Open(Env.DBDriver, Env.DBSource)
	if err != nil {
		log.Fatal("Failed to connect", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(Env, store)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Start(Env.ServerAddress)
	if err != nil {
		log.Fatal("error starting server", err)
	}

}
