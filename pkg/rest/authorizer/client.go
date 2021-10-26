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

	"github.com/aserto-dev/aserto-go"
)

type RestAuthorizer struct {
	options  aserto.ConnectionOptions
	client   *http.Client
	defaults aserto.AuthorizerParams
}

var _ aserto.Authorizer = (*RestAuthorizer)(nil)

var ErrHTTPFailure = errors.New("http error response")

func NewRestAuthorizer(opts ...aserto.ConnectionOption) (*RestAuthorizer, error) {
	options := &aserto.ConnectionOptions{}

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

func ConfigureTLS(options *aserto.ConnectionOptions) *tls.Config {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if options.Insecure {
		config.InsecureSkipVerify = true
	}

	return config
}

func (authorizer *RestAuthorizer) Decide(
	ctx context.Context,
	params ...aserto.AuthorizerParam,
) (aserto.DecisionResults, error) {
	args, err := authorizer.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/is", authorizer.options.Address)
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

	resp, err := authorizer.postRequest(ctx, url, body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ReadDecisions(resp.Body)
}

func (authorizer *RestAuthorizer) DecisionTree(
	ctx context.Context,
	sep aserto.PathSeparator,
	params ...aserto.AuthorizerParam,
) (*aserto.DecisionTree, error) {
	args, err := authorizer.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/decisiontree", authorizer.options.Address)
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

	resp, err := authorizer.postRequest(ctx, url, body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ReadDecisionTree(resp.Body)
}

func (authorizer *RestAuthorizer) Options(params ...aserto.AuthorizerParam) error {
	for _, param := range params {
		param(&authorizer.defaults)
	}
	return nil
}

func (authorizer *RestAuthorizer) postRequest(ctx context.Context, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if authorizer.addRequestHeaders(req) != nil {
		return nil, err
	}

	resp, err := authorizer.client.Do(req)
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

func (authorizer *RestAuthorizer) addRequestHeaders(req *http.Request) (err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Aserto-Tenant-Id", string(authorizer.options.TenantID))
	err = authorizer.addAuthenticationHeader(req)

	return
}

func (authorizer *RestAuthorizer) addAuthenticationHeader(req *http.Request) (err error) {
	headerMap, err := authorizer.options.Creds.GetRequestMetadata(context.Background())
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
