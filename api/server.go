package api

import (
	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{
		store:  store,
		router: gin.Default(),
	}

	server.router.POST("/api/accounts", server.CreateAccount)
	server.router.GET("api/accounts", server.ListAccounts)
	server.router.GET("api/accounts/:id", server.GetAccount)

	return server

}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func httpResponse(ctx *gin.Context, r Response) {
	ctx.JSON(r.Status, gin.H{"Error": r.Error, "Data": r.Data, "Message": r.Message})
}

type Response struct {
	Error    string `json:"error"`
	Data     any    `json:"data"`
	Status   int    `json:"status"`
	Message  any    `json:"message"`
	Feedback any    `json:"feedback"`
}
