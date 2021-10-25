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
)

type RestAuthorizer struct {
	options Options
	client  *http.Client
}

var _ Authorizer = (*RestAuthorizer)(nil)

var ErrHTTPFailure = errors.New("http error response")

func NewRestAuthorizer(opts *Options) (*RestAuthorizer, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return &RestAuthorizer{options: *opts, client: client}, nil
}

func (authz *RestAuthorizer) Decide(
	ctx context.Context,
	params ...Param,
) (DecisionResults, error) {
	args, err := authz.options.defaults.applyOverrides(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/is", authz.options.server)
	body, err := json.Marshal(map[string]interface{}{
		"identityContext": map[string]interface{}{
			"type":     args.identityType,
			"identity": args.identity,
		},
		"policyContext": map[string]interface{}{
			"id":        args.policyID,
			"path":      args.policyPath,
			"decisions": args.decisions,
		},
		"resourceContext": map[string]interface{}(*args.resource),
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
	params ...Param,
) (*DecisionTree, error) {
	args, err := authz.options.defaults.applyOverrides(params...)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/authz/decisiontree", authz.options.server)
	body, err := json.Marshal(map[string]interface{}{
		"identityContext": map[string]interface{}{
			"type":     args.identityType,
			"identity": args.identity,
		},
		"policyContext": map[string]interface{}{
			"id":        args.policyID,
			"path":      args.policyPath,
			"decisions": args.decisions,
		},
		"resourceContext": map[string]interface{}(*args.resource),
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
	req.Header.Set("Aserto-Tenant-Id", authz.options.tenantID)
	err = authz.addAuthenticationHeader(req)

	return
}

func (authz *RestAuthorizer) addAuthenticationHeader(req *http.Request) (err error) {
	headerMap, err := authz.options.credentials.GetRequestMetadata(context.Background())
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
