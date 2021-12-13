package http_test

import (
	"context"
	"log"
	"net/http"

	"github.com/aserto-dev/aserto-go/authorizer/grpc"
	"github.com/aserto-dev/aserto-go/client"
	mw "github.com/aserto-dev/aserto-go/middleware/http"
	"github.com/gorilla/mux"
)

func Example_gorillamux() {
	// Using Aserto middleware with gorilla/mux.

	Hello := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`"hello"`)); err != nil {
			log.Println("Failed to write HTTP response:", err)
		}
	}

	ctx := context.Background()

	// Create authorizer client.
	authorizer, err := grpc.New(
		ctx,
		client.WithAPIKeyAuth("<Aserto authorizer API Key>"),
		client.WithTenantID("<Aserto tenant ID>"),
	)
	if err != nil {
		log.Fatal("Failed to create authorizer client:", err)
	}

	// Create HTTP middleware.
	middleware := mw.New(
		authorizer,
		mw.Policy{
			ID:       "<Aserto policy ID>",
			Decision: "<authorization decision (e.g. 'allowed')",
		},
	)

	// Create ServeMux.
	r := mux.NewRouter()
	r.Use(middleware.Handler)

	// Define HTTP route with middleware.
	r.Handle("/", http.HandlerFunc(Hello))

	// Start server.
	log.Fatal(http.ListenAndServe(":8080", nil))

}
