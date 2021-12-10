package grpc_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aserto-dev/aserto-go/authorizer/grpc"
	"github.com/aserto-dev/aserto-go/client"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

func Example() {
	ctx := context.Background()

	// Create new authorizer client.
	authorizer, err := grpc.New(
		ctx,
		client.WithAPIKeyAuth("<Aserto authorizer API key"),
		client.WithTenantID("<Aserto tenant ID>"),
	)
	if err != nil {
		log.Fatal("Failed to create authorizer:", err)
	}

	// Make an authorization call.
	result, err := authorizer.Is(
		ctx,
		&authz.IsRequest{
			PolicyContext: &api.PolicyContext{
				Id:        "<Aserto Policy ID>",
				Path:      "<Policy path (e.g. 'peoplefinder.GET.users')",
				Decisions: []string{"<authorization decisions (e.g. 'allowed')>"},
			},
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
				Identity: "<user id>",
			},
		},
	)
	if err != nil {
		log.Fatal("Failed to make authorization call:", err)
	}

	// Check the authorizer's decision.
	for _, decision := range result.Decisions {
		if decision.Decision == "allowed" { // "allowed" is just an example. Your policy may have different rules.
			if decision.Is {
				fmt.Println("Access granted")
			} else {
				fmt.Println("Access denied")
			}
		}
	}
}
