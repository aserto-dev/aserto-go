package authorizer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

func ReadDecisions(reader io.Reader) (DecisionResults, error) {
	body, err := unmarshalObject(reader)
	if err != nil {
		return nil, err
	}

	return unmarshalDecisionResults(body["decisions"])
}

func ReadDecisionTree(reader io.Reader) (*DecisionTree, error) {
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
	err = json.Unmarshal([]byte(content), &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func unmarshalDecisionResults(jsonDecisions interface{}) (DecisionResults, error) {
	decisions, ok := jsonDecisions.([]interface{})
	if !ok {
		return nil, UnexpectedJSONSchema
	}

	results := DecisionResults{}
	for _, d := range decisions {
		decision, ok := d.(map[string]interface{})
		name, ok := decision["decision"]
		if !ok {
			return nil, fmt.Errorf("missing 'decision' key: %v", decision)
		}

		is, ok := decision["is"]
		if !ok {
			return nil, fmt.Errorf("missing 'is' key: %v", decision)
		}
		results[name.(string)] = is.(bool)
	}

	return results, nil

}

func unmarshalDecisionTree(jsonTree interface{}) (*DecisionTree, error) {
	tree, ok := jsonTree.(map[string]interface{})
	if !ok {
		return nil, UnexpectedJSONSchema
	}
	root, err := unmarshalStringMapValue(tree, "path_root")
	if err != nil {
		return nil, fmt.Errorf("%w: path_root", err)
	}
	path, ok := tree["path"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: path", UnexpectedJSONSchema)
	}
	return &DecisionTree{Root: root, Path: path}, nil
}

func unmarshalStringMapValue(obj map[string]interface{}, key string) (string, error) {
	if _, ok := obj[key]; !ok {
		return "", fmt.Errorf("%w: missing key '%s'", UnexpectedJSONSchema, key)
	}
	if val, ok := obj[key].(string); !ok {
		return "", fmt.Errorf(
			"%w: unexpected value in '%s'. expected string, found '%v'",
			UnexpectedJSONSchema,
			key,
			obj[key],
		)
	} else {
		return val, nil
	}
}
