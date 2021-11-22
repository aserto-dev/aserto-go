package http // nolint:testpackage  // testing unexported logic

import (
	"testing"

	"gotest.tools/assert"
)

type TestCase struct {
	name     string
	hostname string
	level    int
	expected string
}

func TestHostnameSegment(t *testing.T) {
	testCases := []TestCase{
		{"should accept a valid positive index", "user.example.com", 0, "user"},
		{"should accept a valid negative index", "com.example.user", -1, "user"},
		{"should be empty if index is too high", "user.example.com", 5, ""},
		{"should be empty if index is too low", "user.example.com", -5, ""},
		{"should be empty hostname is empty", "", 0, ""},
	}

	for _, test := range testCases {
		t.Run(test.name, hostnameSegmentTest(test))
	}
}

func hostnameSegmentTest(test TestCase) func(*testing.T) {
	return func(t *testing.T) {
		actual := hostnameSegment(test.hostname, test.level)
		assert.Equal(t, test.expected, actual)
	}
}
