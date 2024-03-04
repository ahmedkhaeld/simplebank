package api

import (
	"fmt"
	"net/http"

	customValidator "github.com/ahmedkhaeld/simplebank/api/custom-validators"
	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/token"
	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	env        util.Env
}

func NewServer(env util.Env, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(env.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)

	}
	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		env:        env,
	}

	//register the custom validators to gin
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", customValidator.ValidateCurrency)
	}

	server.setupRouter()

	return server, nil

}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) setupRouter() {

	router := gin.Default()
	apiRoutes := router.Group("/api")

	apiRoutes.POST("/users", server.CreateUser)
	apiRoutes.POST("/users/login", server.LoginUser)
	apiRoutes.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := apiRoutes.Group("/").Use(BearerMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.CreateAccount)
	authRoutes.GET("/accounts", server.ListAccounts)
	authRoutes.GET("/accounts/:id", server.GetAccount)
	authRoutes.POST("/transfers", server.CreateTransfer)

	server.router = router

}

func httpResponse(ctx *gin.Context, r Response) {
	ctx.JSON(r.Status, gin.H{"Error": r.Error, "Data": r.Data, "Message": r.Message})
}

func httpUnauthorized(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, Response{
		Status: http.StatusUnauthorized,
		Error:  err.Error(),
	})
}

type Response struct {
	Error    string `json:"error"`
	Data     any    `json:"data"`
	Status   int    `json:"status"`
	Message  any    `json:"message"`
	Feedback any    `json:"feedback"`
}
