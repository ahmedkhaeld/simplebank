package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

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

	go runGatewayServer(Env, store)

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

func runGatewayServer(env util.Env, store db.Store) {

	//create and initialize a new gRPC server
	server, err := gapi.NewServer(env, store)
	if err != nil {
		log.Fatal("can not create grpc server", err)
	}

	// jsonOptions to enable snake case for fields of the proto messages for the grpc-gateway server
	jsonOptions := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOptions)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("can not register gateway server", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", env.HTTPServerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("http gateway is being started on %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatalf("failed to serve from http gateway: %v", err)
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
