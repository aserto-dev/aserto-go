package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/grpc/authorizer"
	middleware "github.com/aserto-dev/aserto-go/middleware/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewAuthorizaionMiddleware(userKey interface{}) (*middleware.Middleware, error) {
	ctx := context.Background()

	// Create an authorization client
	authzClient, err := authorizer.New(
		ctx,
		client.WithAddr("localhost:8282"),
		client.WithTenantID("0fb5d7eb-8190-4f9d-ac7f-db0ba8374cb7"),
		client.WithInsecure(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create authorizer client")
	}

	// Create and configure authorization middleware
	mw := middleware.New(
		authzClient,
		middleware.Policy{ID: "messageboards", Decision: "allowed"},
	)
	mw.Identity.Subject().FromContextValue(userKey)
	mw.WithPolicyFromURL("messageboards").WithResourceMapper(resourceContext)

	return mw, nil
}

func resourceContext(r *http.Request) *structpb.Struct {
	board := getBoard(r.Context())
	if board == nil {
		return nil
	}

	resource := map[string]interface{}{
		"board": structToMap(board),
	}

	if messageID, err := messageIDFromString(mux.Vars(r)["messageID"]); err == nil {
		if message, ok := board.Messages[messageID]; ok {
			resource["message"] = structToMap(message)
		}
	}

	resourceStruct, err := structpb.NewStruct(resource)
	if err != nil {
		return nil
	}

	return resourceStruct
}

func structToMap(data interface{}) map[string]interface{} {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	mapData := make(map[string]interface{})

	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil
	}

	return mapData
}
