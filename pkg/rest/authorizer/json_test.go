package authorizer_test

import (
	"strings"
	"testing"

	"github.com/aserto-dev/aserto-go"
	"github.com/aserto-dev/aserto-go/pkg/rest/authorizer"
)

func TestReadDecisions(t *testing.T) {
	decisionsJSON := `{
	"decisions": [
		{
			"decision": "allowed",
			"is":       true
		},
		{
			"decision": "visible",
			"is":       false
		}
	]
	}
	`
	results, err := authorizer.ReadDecisions(strings.NewReader(decisionsJSON))
	if err != nil {
		t.Error(err)
	}

	t.Run("allowed", resultsContainDecision(results, "allowed", true))
	t.Run("not visible", resultsContainDecision(results, "visible", false))
}

func resultsContainDecision(results aserto.DecisionResults, rule string, expectedDecision bool) func(*testing.T) {
	check := func(t *testing.T) {
		if decision, ok := results[rule]; !ok {
			t.Errorf("results missing decision for '%s'", rule)
		} else if decision != expectedDecision {
			t.Errorf("unexpected decision for '%s': %v", rule, decision)
		}
	}

	return check
}

func TestReadDecisionTree(t *testing.T) {
	decisionTreeJSON := `
	{
	  "path_root": "peoplefinder.POST.api.users",
	  "path": {
		"peoplefinder.POST.api.users": {
		  "allowed": false,
		  "enabled": false,
		  "visible": false
		},
		"peoplefinder.POST.api.users.__id": {
		  "allowed": false,
		  "enabled": true,
		  "visible": true
		}
	  }
	}
	`
	results, err := authorizer.ReadDecisionTree(strings.NewReader(decisionTreeJSON))
	if err != nil {
		t.Error(err)
	}

	if results.Root != "peoplefinder.POST.api.users" {
		t.Errorf("wrong path root: %s", results.Root)
	}
	path, ok := results.Path["peoplefinder.POST.api.users"]
	if !ok {
		t.Error("unexpected path type")
	}
	decisions, ok := path.(map[string]interface{})
	if !ok {
		t.Error("unexpected path element type")
	}
	t.Run("not allowed", assertDecision(decisions, "allowed", false))
	t.Run("not enabled", assertDecision(decisions, "enabled", false))
	t.Run("not visible", assertDecision(decisions, "visible", false))
}

func assertDecision(decisions map[string]interface{}, key string, expected bool) func(*testing.T) {
	assert := func(t *testing.T) {
		val, ok := decisions[key]
		if !ok {
			t.Errorf("missing key '%s'", key)
		}
		decision, ok := val.(bool)
		if !ok {
			t.Error("unexpected decision type")
		}
		if decision != expected {
			t.Errorf("expected '%v' found '%v'", expected, decision)
		}
	}
	return assert
}
