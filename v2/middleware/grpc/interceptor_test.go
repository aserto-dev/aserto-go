package grpc_test

import (
	"context"
	"fmt"
	"testing"

	grpcmw "github.com/aserto-dev/aserto-go/v2/middleware/grpc"
	"github.com/aserto-dev/aserto-go/v2/middleware/internal/mock"
	"github.com/aserto-dev/aserto-go/v2/middleware/internal/test"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type TestCase struct {
	*test.Case
	expectedErr error
	middleware  *grpcmw.Middleware
}

type testOptions struct {
	test.Options
	expectedErr error
	callback    func(*grpcmw.Middleware)
}

const DefaultPolicyPath = "policy.path"

func NewTest(t *testing.T, name string, options *testOptions) *TestCase {
	if options.ExpectedRequest == nil && options.PolicyPath == "" {
		options.PolicyPath = DefaultPolicyPath
	}

	base := test.NewTest(t, name, &options.Options)

	mw := grpcmw.New(base.Client, test.Policy(DefaultPolicyPath))

	if options.callback == nil {
		mw.Identity.Subject().ID(test.DefaultUsername)
	} else {
		options.callback(mw)
	}

	return &TestCase{Case: base, middleware: mw, expectedErr: options.expectedErr}
}

func TestAuthorizer(t *testing.T) {
	tests := []*TestCase{
		NewTest(
			t,
			"authorized decisions should succeed",
			&testOptions{},
		),
		NewTest(
			t,
			"unauthorized decisions should err",
			&testOptions{
				Options: test.Options{
					Reject: true,
				},
				expectedErr: aerr.ErrAuthorizationFailed,
			},
		),
		NewTest(
			t,
			"policy mapper should override policy path",
			&testOptions{
				Options: test.Options{
					PolicyPath: test.OverridePolicyPath,
				},
				callback: func(mw *grpcmw.Middleware) {
					mw.WithPolicyPathMapper(
						func(_ context.Context, _ interface{}) string {
							return test.OverridePolicyPath
						},
					).Identity.Subject().ID(test.DefaultUsername)
				},
			},
		),
	}

	for _, test := range tests {
		for runnerName, runner := range runners() {
			t.Run(
				fmt.Sprintf("%s: %s", test.Name, runnerName),
				testCase(test, runner),
			)
		}
	}
}

func runners() map[string]func(*grpcmw.Middleware) error {
	return map[string]func(*grpcmw.Middleware) error{"unary": runUnary, "stream": runStream}
}

type testRunner func(*grpcmw.Middleware) error

func testCase(testCase *TestCase, runner testRunner) func(*testing.T) {
	return func(t *testing.T) {
		err := runner(testCase.middleware)
		if testCase.expectedErr == nil {
			assert.NoError(t, err)
		} else {
			assert.ErrorIs(t, err, testCase.expectedErr)
		}
	}
}

func runUnary(mw *grpcmw.Middleware) error {
	_, err := mw.Unary()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{},
		func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, nil
		},
	)

	return err
}

func runStream(mw *grpcmw.Middleware) error {
	return mw.Stream()(
		nil,
		&mock.ServerStream{Ctx: context.Background()},
		&grpc.StreamServerInfo{},
		func(_ interface{}, _ grpc.ServerStream) error {
			return nil
		},
	)
}
