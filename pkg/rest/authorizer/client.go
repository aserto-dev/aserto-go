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

	authz "github.com/aserto-dev/aserto-go/pkg/authorizer"
)

type RestAuthorizer struct {
	options authz.Options
	client  *http.Client
}

var _ authz.Authorizer = (*RestAuthorizer)(nil)

var ErrHTTPFailure = errors.New("http error response")

func NewRestAuthorizer(opts ...authz.Option) (*RestAuthorizer, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	options := &authz.Options{}

	for _, opt := range opts {
		opt(options)
	}

	return &RestAuthorizer{options: *options, client: client}, nil
}

func (authorizer *RestAuthorizer) Decide(
	ctx context.Context,
	params ...authz.Param,
) (authz.DecisionResults, error) {
	args, err := authorizer.options.Defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/is", authorizer.options.Server)
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
	sep authz.PathSeparator,
	params ...authz.Param,
) (*authz.DecisionTree, error) {
	args, err := authorizer.options.Defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/decisiontree", authorizer.options.Server)
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

func (authorizer *RestAuthorizer) Options(params ...authz.Param) error {
	for _, param := range params {
		param(&authorizer.options.Defaults)
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
	req.Header.Set("Aserto-Tenant-Id", authorizer.options.TenantID)
	err = authorizer.addAuthenticationHeader(req)

	return
}

func (authorizer *RestAuthorizer) addAuthenticationHeader(req *http.Request) (err error) {
	headerMap, err := authorizer.options.Credentials.GetRequestMetadata(context.Background())
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
