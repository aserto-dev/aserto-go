package authorizer

import (
	"testing"
)

func TestNewDecisionResults(t *testing.T) {
	decisions := []map[string]interface{}{
		{
			"decision": "allowed",
			"is":       true,
		},
		{
			"decision": "visible",
			"is":       false,
		},
	}

	results, err := NewDecisionResults(decisions)
	if err != nil {
		t.Error(err)
	}

	t.Run("test allowed", resultsContainDecision(results, "allowed", true))
	t.Run("test not visible", resultsContainDecision(results, "visible", false))

}

func resultsContainDecision(results DecisionResults, rule string, expectedDecision bool) func(*testing.T) {
	check := func(t *testing.T) {
		if decision, ok := results[rule]; !ok {
			t.Errorf("results missing decision for '%s'", rule)
		} else if decision != expectedDecision {
			t.Errorf("unexpected decision for '%s': %v", rule, decision)
		}

	}
	return check
}
