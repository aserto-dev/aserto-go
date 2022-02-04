package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aserto-dev/aserto-go/authorizer/grpc"
	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/http/ginz"
	"github.com/gin-gonic/gin"
)

const port = 8080

func main() {
	ctx := context.Background()
	authClient, err := grpc.New(
		ctx,
		client.WithAddr("localhost:8282"),
	)
	if err != nil {
		log.Fatalln("Failed to create authorizer client:", err)
	}

	mw := ginz.New(
		authClient,
		middleware.Policy{
			ID:       "local",
			Decision: "allowed",
		},
	)
	mw.Identity.Mapper(func(r *http.Request, identity middleware.Identity) {
		if username, _, ok := r.BasicAuth(); ok {
			identity.Subject().ID(username)
		}
	})
	mw.WithPolicyFromURL("example")

	router := gin.Default()
	router.Use(mw.Handler)
	router.GET("/api/:asset", Handler)
	router.POST("/api/:asset", Handler)
	router.DELETE("/api/:asset", Handler)

	router.Run(fmt.Sprintf(":%d", port))
}

func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, "Permission granted")
}
