package authorizer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	op "github.com/aserto-dev/aserto-go/options"
	"google.golang.org/grpc/credentials"
)

type RestAuthorizer struct {
	options  op.ConnectionOptions
	client   *http.Client
	defaults AuthorizerParams
}

var _ Authorizer = (*RestAuthorizer)(nil)

var ErrHTTPFailure = errors.New("http error response")

func NewRestAuthorizer(opts ...op.ConnectionOption) (*RestAuthorizer, error) {
	options := &op.ConnectionOptions{}

	for _, opt := range opts {
		opt(options)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: ConfigureTLS(options),
		},
	}

	return &RestAuthorizer{options: *options, client: client}, nil
}

func ConfigureTLS(options *op.ConnectionOptions) *tls.Config {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if options.Insecure {
		config.InsecureSkipVerify = true
	}

	return config
}

func (authz *RestAuthorizer) Decide(
	ctx context.Context,
	params ...AuthorizerParam,
) (DecisionResults, error) {
	args, err := authz.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/is", authz.address())
	body, err := json.Marshal(map[string]interface{}{
		"identityContext": map[string]interface{}{
			"type":     args.IdentityType,
			"identity": args.Identity,
		},
		"policyContext": map[string]interface{}{
			"id":        args.PolicyID,
			"path":      args.PolicyPath,
			"decisions": args.Decisions,
		},
		"resourceContext": map[string]interface{}(*args.Resource),
	})
	if err != nil {
		return nil, err
	}

	resp, err := authz.postRequest(ctx, url, body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ReadDecisions(resp.Body)
}

func (authz *RestAuthorizer) DecisionTree(
	ctx context.Context,
	sep PathSeparator,
	params ...AuthorizerParam,
) (*DecisionTree, error) {
	args, err := authz.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/decisiontree", authz.address())
	body, err := json.Marshal(map[string]interface{}{
		"identityContext": map[string]interface{}{
			"type":     args.IdentityType,
			"identity": args.Identity,
		},
		"policyContext": map[string]interface{}{
			"id":        args.PolicyID,
			"path":      args.PolicyPath,
			"decisions": args.Decisions,
		},
		"resourceContext": map[string]interface{}(*args.Resource),
		"options": map[string]interface{}{
			"pathSeparator": sep,
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := authz.postRequest(ctx, url, body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ReadDecisionTree(resp.Body)
}

func (authz *RestAuthorizer) Options(params ...AuthorizerParam) error {
	for _, param := range params {
		param(&authz.defaults)
	}
	return nil
}

func (authz *RestAuthorizer) address() string {
	return authz.options.Address
}

func (authz *RestAuthorizer) tenantID() string {
	return authz.options.TenantID
}

func (authz *RestAuthorizer) credentials() credentials.PerRPCCredentials {
	return authz.options.Creds
}

func (authz *RestAuthorizer) postRequest(ctx context.Context, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if authz.addRequestHeaders(req) != nil {
		return nil, err
	}

	resp, err := authz.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		return nil,
			fmt.Errorf("%w: http request failed. status: %s. body: %s",
				ErrHTTPFailure,
				resp.Status,
				tryReadText(resp.Body),
			)
	}

	return resp, nil
}

func (authz *RestAuthorizer) addRequestHeaders(req *http.Request) (err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Aserto-Tenant-Id", authz.tenantID())
	err = authz.addAuthenticationHeader(req)

	return
}

func (authz *RestAuthorizer) addAuthenticationHeader(req *http.Request) (err error) {
	headerMap, err := authz.credentials().GetRequestMetadata(context.Background())
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
		return ""
	}

	return string(content)
}
