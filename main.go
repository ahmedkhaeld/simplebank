package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"

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
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var Env = util.Env{}

func init() {
	var err error
	Env, err = util.LoadEnv(".")
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to load environment")
	}
}

func main() {

	conn, err := sql.Open(Env.DBDriver, Env.DBSource)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to open database connection")
	}
	if Env.Environment == "development" {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	runDBMigration(Env.MigrationURL, Env.DBSource)

	store := db.NewStore(conn)

	go runGatewayServer(Env, store)

	runGrpcServer(Env, store)

}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		zlog.Fatal().Err(err).Msg("can't create migration instance")
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		zlog.Fatal().Err(err).Msg("can't migrate database up")
	}
	zlog.Info().Msg("db migrated successfully")
}

func runGrpcServer(env util.Env, store db.Store) {

	//create and initialize a new gRPC server

	server, err := gapi.NewServer(env, store)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to initialize gRPC server")
	}

	gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(gprcLogger)

	//Register SimpleBank server with the gRPC server
	pb.RegisterSimpleBankServer(grpcServer, server)

	// allow clients to introspect the server services and methods
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", env.GRPCServerAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msg("cannot create listener")
	}
	zlog.Info().Msgf("start gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			zlog.Fatal().Err(err).Msg("server stopped")
		}
		zlog.Fatal().Err(err).Msg("cannot serve gRPC server")
	}

}

func runGatewayServer(env util.Env, store db.Store) {

	//create and initialize a new gRPC server
	server, err := gapi.NewServer(env, store)
	if err != nil {
		zlog.Fatal().Err(err).Msg("cannot create server")
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
		zlog.Fatal().Err(err).Msg("cannot register simple bank handler")

	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", env.HTTPServerAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msg("cannot create listener")
	}

	zlog.Info().Msgf("start HTTP gateway server at %s", listener.Addr().String())

	handler := gapi.HttpLogger(mux)

	err = http.Serve(listener, handler)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			zlog.Fatal().Err(err).Msg("server stopped")
		}
		zlog.Fatal().Err(err).Msg("cannot serve HTTP gateway server")
	}

}

func runGinServer(env util.Env, store db.Store) {
	server, err := api.NewServer(env, store)
	if err != nil {
		zlog.Fatal().Err(err).Msg("cannot create server")

	}

	err = server.Start(Env.HTTPServerAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msg("cannot start server")
	}
	zlog.Info().Msg("HTTP server started")
}
