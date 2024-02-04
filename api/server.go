package api

import (
	customValidator "github.com/ahmedkhaeld/simplebank/api/custom-validators"
	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{
		store:  store,
		router: gin.Default(),
	}

	//register the custom validators to gin
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", customValidator.ValidateCurrency)
	}

	server.router.POST("/api/accounts", server.CreateAccount)
	server.router.GET("api/accounts", server.ListAccounts)
	server.router.GET("api/accounts/:id", server.GetAccount)

	server.router.POST("/api/transfers", server.CreateTransfer)

	server.router.POST("/api/users", server.CreateUser)

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
