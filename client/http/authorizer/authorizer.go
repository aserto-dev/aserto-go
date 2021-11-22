package authorizer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/internal"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type AuthorizerClient = authz.AuthorizerClient

// Error codes for REST authorizer.
var (
	ErrHTTPFailure  = errors.New("received http failure response")
	ErrNotSupported = errors.New("unsupported feature")
)

type authorizer struct {
	httpClient *http.Client
	options    *client.ConnectionOptions
}

// New returns a new REST authorizer with the specified options.
func New(opts ...client.ConnectionOption) (AuthorizerClient, error) {
	options := client.NewConnectionOptions(opts...)

	tlsConf, err := internal.TLSConfig(options.Insecure)
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

func (a *authorizer) postAPIRequest(
	ctx context.Context,
	endpoint string,
	message proto.Message,
	opts []grpc.CallOption,
) ([]byte, error) {
	if len(opts) > 0 {
		return nil, ErrNotSupported
	}

	resp, err := a.postRequest(ctx, a.endpointURL(endpoint), message)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (a *authorizer) serverAddress() string {
	if a.options.Address != "" {
		return a.options.Address
	}

	return internal.HostedAuthorizerHostname
}

func (a *authorizer) endpointURL(endpoint string) string {
	return fmt.Sprintf("https://%s/api/v1/authz/%s", a.serverAddress(), endpoint)
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

		return nil,
			errors.Wrap(
				ErrHTTPFailure,
				fmt.Sprintf("http request failed. status: '%s'. body: '%s'",
					resp.Status,
					tryReadText(resp.Body),
				),
			)
	}

	return resp, nil
}

func (a *authorizer) addRequestHeaders(req *http.Request) (err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Aserto-Tenant-Id", a.options.TenantID)
	err = a.addAuthenticationHeader(req)

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
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Sprintf("failed to read response body: %s", err.Error())
	}

	return string(content)
}
