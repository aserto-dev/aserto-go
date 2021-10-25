package authorizer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	authz "github.com/aserto-dev/aserto-go/pkg/authorizer"
)

func ReadDecisions(reader io.Reader) (authz.DecisionResults, error) {
	body, err := unmarshalObject(reader)
	if err != nil {
		return nil, err
	}

	return unmarshalDecisionResults(body["decisions"])
}

func ReadDecisionTree(reader io.Reader) (*authz.DecisionTree, error) {
	body, err := unmarshalObject(reader)
	if err != nil {
		return nil, err
	}

	return unmarshalDecisionTree(body)
}

func unmarshalObject(reader io.Reader) (map[string]interface{}, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}

	err = json.Unmarshal(content, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func unmarshalDecisionResults(jsonDecisions interface{}) (authz.DecisionResults, error) {
	decisions, ok := jsonDecisions.([]interface{})
	if !ok {
		return nil, authz.ErrUnexpectedJSONSchema
	}

	results := authz.DecisionResults{}

	for _, d := range decisions {
		decision, ok := d.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: decision should be an object, found %v", authz.ErrUnexpectedJSONSchema, d)
		}

		name, ok := decision["decision"]
		if !ok {
			return nil, fmt.Errorf("%w: missing 'decision' key: %v", authz.ErrUnexpectedJSONSchema, decision)
		}

		is, ok := decision["is"]
		if !ok {
			return nil, fmt.Errorf("%w: missing 'is' key: %v", authz.ErrUnexpectedJSONSchema, decision)
		}

		results[name.(string)] = is.(bool)
	}

	return results, nil
}

func unmarshalDecisionTree(jsonTree interface{}) (*authz.DecisionTree, error) {
	tree, ok := jsonTree.(map[string]interface{})
	if !ok {
		return nil, authz.ErrUnexpectedJSONSchema
	}

	root, err := unmarshalStringMapValue(tree, "path_root")
	if err != nil {
		return nil, fmt.Errorf("%w: path_root", err)
	}

	path, ok := tree["path"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: path", authz.ErrUnexpectedJSONSchema)
	}

	return &authz.DecisionTree{Root: root, Path: path}, nil
}

func unmarshalStringMapValue(obj map[string]interface{}, key string) (string, error) {
	if _, ok := obj[key]; !ok {
		return "", fmt.Errorf("%w: missing key '%s'", authz.ErrUnexpectedJSONSchema, key)
	}

	val, ok := obj[key].(string)
	if !ok {
		return "", fmt.Errorf(
			"%w: unexpected value in '%s'. expected string, found '%v'",
			authz.ErrUnexpectedJSONSchema,
			key,
			obj[key],
		)
	}

	return val, nil
}
