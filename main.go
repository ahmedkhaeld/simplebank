package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ahmedkhaeld/simplebank/api"
	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/gapi"
	"github.com/ahmedkhaeld/simplebank/pb"
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

	runGrpcServer(Env, store)

}

func runGrpcServer(env util.Env, store db.Store) {

	//create and initialize a new gRPC server
	server, err := gapi.NewServer(env, store)
	if err != nil {
		log.Fatal("can not create grpc server", err)
	}
	grpcServer := grpc.NewServer()

	//Register SimpleBank server with the gRPC server
	pb.RegisterSimpleBankServer(grpcServer, server)

	// allow clients to introspect the server services and methods
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", env.GRPCServerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC server listening on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func runGinServer(env util.Env, store db.Store) {
	server, err := api.NewServer(env, store)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Start(Env.HTTPServerAddress)
	if err != nil {
		log.Fatal("error starting server", err)
	}
}
