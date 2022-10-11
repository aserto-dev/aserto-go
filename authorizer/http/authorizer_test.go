package http_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aserto-dev/aserto-go/authorizer/http"
	"github.com/aserto-dev/aserto-go/client"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

func Example() {
	ctx := context.Background()

	// Create new authorizer client.
	authorizer, err := http.New(
		client.WithAPIKeyAuth("<Aserto authorizer API key"),
	)
	if err != nil {
		log.Fatal("Failed to create authorizer:", err)
	}

	// Make an authorization call.
	result, err := authorizer.Is(
		ctx,
		&authz.IsRequest{
			PolicyContext: &api.PolicyContext{
				Name:          "<Aserto Policy Name>",
				Path:          "<Policy path (e.g. 'peoplefinder.GET.users')",
				Decisions:     []string{"<authorization decisions (e.g. 'allowed')>"},
				InstanceLabel: "<Aserto Policy Intsance Label>",
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
