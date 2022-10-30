package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/aserto-dev/aserto-go/v2/authorizer/grpc"
	"github.com/aserto-dev/aserto-go/v2/client"
	"github.com/aserto-dev/aserto-go/v2/middleware"
	"github.com/aserto-dev/aserto-go/v2/middleware/http/std"
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

	mw := std.New(
		authClient,
		middleware.Policy{
			Name:          "local",
			Decision:      "allowed",
			InstanceLabel: "label",
		},
	)
	mw.Identity.Mapper(func(r *http.Request, identity middleware.Identity) {
		if username, _, ok := r.BasicAuth(); ok {
			identity.Subject().ID(username)
		}
	})
	mw.WithPolicyFromURL("example")

	router := mux.NewRouter()
	router.HandleFunc("/api/{asset}", Handler).Methods("GET", "POST", "DELETE")

	router.Use(mw.Handler)
	start(router)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`"Permission granted"`))
}

func start(h http.Handler) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Println("Staring server on", addr)

	srv := http.Server{
		Handler: h,
		Addr:    addr,
	}
	log.Fatal(srv.ListenAndServe())
}
