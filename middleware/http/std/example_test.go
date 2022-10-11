package std_test

import (
	"context"
	"log"
	"net/http"

	"github.com/aserto-dev/aserto-go/authorizer/grpc"
	"github.com/aserto-dev/aserto-go/client"
	mw "github.com/aserto-dev/aserto-go/middleware/http/std"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(`"hello"`)); err != nil {
		log.Println("Failed to write HTTP response:", err)
	}
}

func Example() {
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
			Name:          "<Aserto policy Name>",
			Decision:      "<authorization decision (e.g. 'allowed')",
			InstanceLabel: "<Aserto  policy instance label>",
		},
	)

	// Define HTTP route.
	http.Handle(
		"/",
		middleware.Handler(http.HandlerFunc(Hello)), // Attach middleware to route.
	)

	// Start server.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
