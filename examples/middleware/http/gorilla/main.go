package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/aserto-dev/aserto-go/authorizer/grpc"
	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/middleware"
	authzmw "github.com/aserto-dev/aserto-go/middleware/http"
)

const port = 8080

func main() {
	ctx := context.Background()
	authClient, err := grpc.New(
		ctx,
		client.WithAddr("localhost:8282"),
		client.WithInsecure(true),
	)
	if err != nil {
		log.Fatalln("Failed to create authorizer client:", err)
	}

	mw := authzmw.New(
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
	mw.WithResourceMapper(func(*http.Request) *structpb.Struct {
		resource, err := structpb.NewStruct(map[string]interface{}{"asset": "secret"})
		if err != nil {
			log.Print("Error creating resource:", err)
			return nil
		}
		return resource
	})

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
