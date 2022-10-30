/*
Package http is used to create an AuthorizerClient that communicates with the authorizer using HTTP.

AuthorizerClient is the low-level interface that exposes the raw authorization API.
*/
package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/aserto-dev/aserto-go/v2/client"
	"github.com/aserto-dev/aserto-go/v2/internal/hosted"
	"github.com/aserto-dev/aserto-go/v2/internal/tlsconf"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type AuthorizerClient = authz.AuthorizerClient

// ErrHttp is returned in response to failed HTTP requests to the authorizer.
type ErrHTTP struct {
	// Status text (e.g. "200 OK")
	Status string

	// Status code
	StatusCode int

	// Response body decoded as a string.
	Body string
}

// Error returns a string representation of the HTTP error.
func (e *ErrHTTP) Error() string {
	return fmt.Sprintf("status: %s. body: %s", e.Status, e.Body)
}

// ErrNotSupported is returned when gRPC options are passed to the HTTP client.
var ErrNotSupported = errors.New("unsupported feature")

type authorizer struct {
	httpClient *http.Client
	options    *client.ConnectionOptions
}

// New returns a new REST authorizer with the specified options.
func New(opts ...client.ConnectionOption) (AuthorizerClient, error) {
	options, err := client.NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	tlsConf, err := tlsconf.TLSConfig(options.Insecure, "")
	if err != nil {
		return nil, err
	}

	httpc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
		},
	}

	return &authorizer{options: options, httpClient: httpc}, nil
}

func (a *authorizer) DecisionTree(
	ctx context.Context,
	in *authz.DecisionTreeRequest,
	opts ...grpc.CallOption,
) (*authz.DecisionTreeResponse, error) {
	respBody, err := a.postAPIRequest(ctx, "decisiontree", in, opts)
	if err != nil {
		return nil, err
	}

	var tree authz.DecisionTreeResponse
	if err := protojson.Unmarshal(respBody, &tree); err != nil {
		return nil, err
	}

	return &tree, nil
}

func (a *authorizer) Is(
	ctx context.Context,
	in *authz.IsRequest,
	opts ...grpc.CallOption,
) (*authz.IsResponse, error) {
	respBody, err := a.postAPIRequest(ctx, "is", in, opts)
	if err != nil {
		return nil, err
	}

	var response authz.IsResponse
	if err := protojson.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *authorizer) Query(
	ctx context.Context,
	in *authz.QueryRequest,
	opts ...grpc.CallOption,
) (*authz.QueryResponse, error) {
	respBody, err := a.postAPIRequest(ctx, "query", in, opts)
	if err != nil {
		return nil, err
	}

	var response authz.QueryResponse
	if err := protojson.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *authorizer) Compile(
	ctx context.Context,
	in *authz.CompileRequest,
	opts ...grpc.CallOption,
) (*authz.CompileResponse, error) {
	respBody, err := a.postAPIRequest(ctx, "compile", in, opts)
	if err != nil {
		return nil, err
	}

	var response authz.CompileResponse
	if err := protojson.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *authorizer) GetPolicy(
	ctx context.Context,
	in *authz.GetPolicyRequest,
	opts ...grpc.CallOption,
) (*authz.GetPolicyResponse, error) {
	var paths []string

	if in.FieldMask != nil {
		paths = in.FieldMask.Paths
	}

	respBody, err := a.getAPIRequest(ctx, "policies/"+in.Id, paths, opts)
	if err != nil {
		return nil, err
	}

	var response authz.GetPolicyResponse
	if err := protojson.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *authorizer) ListPolicies(
	ctx context.Context,
	in *authz.ListPoliciesRequest,
	opts ...grpc.CallOption,
) (*authz.ListPoliciesResponse, error) {
	var paths []string

	if in.FieldMask != nil {
		paths = in.FieldMask.Paths
	}

	respBody, err := a.getAPIRequest(ctx, "policies", paths, opts)
	if err != nil {
		return nil, err
	}

	var response authz.ListPoliciesResponse
	if err := protojson.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *authorizer) Info(
	ctx context.Context,
	in *authz.InfoRequest,
	opts ...grpc.CallOption,
) (*authz.InfoResponse, error) {
	var paths []string

	respBody, err := a.getAPIRequest(ctx, "info", paths, opts)
	if err != nil {
		return nil, err
	}

	var response authz.InfoResponse
	if err := protojson.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *authorizer) getAPIRequest(
	ctx context.Context,
	endpoint string,
	fieldMask []string,
	opts []grpc.CallOption,
) ([]byte, error) {
	if len(opts) > 0 {
		return nil, ErrNotSupported
	}

	endpointURL := a.endpointURL(endpoint)
	if len(fieldMask) > 0 {
		endpointURL += "?field_mask="
		for _, field := range fieldMask {
			endpointURL = endpointURL + field + ","
		}

		endpointURL = strings.TrimSuffix(endpointURL, ",")
	}

	resp, err := a.getRequest(ctx, endpointURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (a *authorizer) postAPIRequest(
	ctx context.Context,
	endpoint string,
	message proto.Message,
	opts []grpc.CallOption,
) ([]byte, error) {
	if len(opts) > 0 {
		return nil, ErrNotSupported
	}

	resp, err := a.postRequest(ctx, a.endpointAuthzURL(endpoint), message)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (a *authorizer) baseURL() string {
	if a.options.URL != nil {
		return a.options.URL.String()
	}

	address := a.options.Address
	if address == "" {
		address = hosted.HostedAuthorizerHostname
	}

	return fmt.Sprintf("https://%s", address)
}

func (a *authorizer) endpointAuthzURL(endpoint string) string {
	return fmt.Sprintf("%s/api/v2/authz/%s", a.baseURL(), endpoint)
}

func (a *authorizer) endpointURL(endpoint string) string {
	return fmt.Sprintf("%s/api/v2/%s", a.baseURL(), endpoint)
}

func (a *authorizer) postRequest(ctx context.Context, url string, message proto.Message) (*http.Response, error) {
	body, err := protojson.Marshal(message)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if a.addRequestHeaders(req) != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		return nil, &ErrHTTP{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       tryReadText(resp.Body),
		}
	}

	return resp, nil
}

func (a *authorizer) getRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if a.addRequestHeaders(req) != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		return nil, &ErrHTTP{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       tryReadText(resp.Body),
		}
	}

	return resp, nil
}

func (a *authorizer) addRequestHeaders(req *http.Request) (err error) {
	req.Header.Set("Content-Type", "application/json")

	if a.options.TenantID != "" {
		req.Header.Set("Aserto-Tenant-Id", a.options.TenantID)
	}

	if a.options.Creds != nil {
		err = a.addAuthenticationHeader(req)
	}

	return
}

func (a *authorizer) addAuthenticationHeader(req *http.Request) (err error) {
	headerMap, err := a.options.Creds.GetRequestMetadata(context.Background())
	if err == nil {
		for key, val := range headerMap {
			req.Header.Set(key, val)
		}
	}

	return
}

func tryReadText(reader io.Reader) string {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Sprintf("failed to read response body: %s", err.Error())
	}

	return string(content)
}
