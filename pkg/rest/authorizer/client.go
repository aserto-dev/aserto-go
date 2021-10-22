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

	"google.golang.org/grpc/credentials"
)

const defaultAuthorizer = "authorizer.prod.aserto.com"

type Resource map[string]interface{}

type DecisionResults map[string]bool

func NewDecisionResults(jsonDecisions interface{}) (DecisionResults, error) {
	decisions, ok := jsonDecisions.([]interface{})
	if !ok {
		return nil, errors.New("unexpected JSON schema")
	}

	results := DecisionResults{}
	for _, d := range decisions {
		decision, ok := d.(map[string]interface{})
		name, ok := decision["decision"]
		if !ok {
			return nil, errors.New(fmt.Sprintf("missing 'decision' key: %v", decision))
		}

		is, ok := decision["is"]
		if !ok {
			return nil, errors.New(fmt.Sprintf("missing 'is' key: %v", decision))
		}
		results[name.(string)] = is.(bool)
	}

	return results, nil

}

type Authorizer interface {
	Decide(ctx context.Context, string, decisions []string, identity, policyPath string, resource Resource) (DecisionResults, error)
}

type Options struct {
	Credentials  credentials.PerRPCCredentials
	Server       string
	TenantID     string
	PolicyID     string
	PolicyPath   string
	IdentityType string
	Identity     string
}

type RestAuthorizer struct {
	options Options
	client  *http.Client
}

func NewRestAuthorizer(opts Options) (*RestAuthorizer, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return &RestAuthorizer{options: opts, client: client}, nil
}

func (authz RestAuthorizer) Decide(
	ctx context.Context,
	decisions []string,
	identity string,
	policyPath string,
	resource Resource,
) (DecisionResults, error) {
	url := fmt.Sprintf("https://%s/api/v1/authz/is", authz.options.Server)
	body, err := json.Marshal(map[string]interface{}{
		"identityContext": map[string]interface{}{
			"type":     authz.options.IdentityType,
			"identity": identity,
		},
		"policyContext": map[string]interface{}{
			"id":        authz.options.PolicyID,
			"path":      policyPath,
			"decisions": decisions,
		},
		"resourceContext": map[string]interface{}(resource),
	})
	if err != nil {
		return nil, err
	}

	return authz.postRequest(url, body)
}

func (authz RestAuthorizer) postRequest(url string, body []byte) (DecisionResults, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(body)))
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(
			fmt.Sprintf("http request failed. status: %s. body: %s",
				resp.Status,
				tryReadText(resp.Body),
			),
		)
	}

	return readDecisions(resp.Body)
}

func readDecisions(body io.Reader) (DecisionResults, error) {
	content, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(content), &m)
	if err != nil {
		return nil, err
	}

	return NewDecisionResults(m["decisions"])
}

func tryReadText(reader io.Reader) string {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return ""
	}
	return string(content)
}

func (authz RestAuthorizer) addRequestHeaders(req *http.Request) (err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Aserto-Tenant-Id", authz.options.TenantID)
	err = authz.addAuthenticationHeader(req)
	return
}

func (authz RestAuthorizer) addAuthenticationHeader(req *http.Request) (err error) {
	headerMap, err := authz.options.Credentials.GetRequestMetadata(context.Background())
	if err == nil {
		for key, val := range headerMap {
			req.Header.Set(key, val)
		}
	}
	return
}
