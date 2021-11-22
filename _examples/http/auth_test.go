package main // nolint:testpackage  // whitebox tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	Anonymous = ""
	Jeff      = "jeffh@acmecorp.com"
	Ian       = "iant@acmecorp.com"
	Kim       = "kimr@acmecorp.com"
)

func TestAuthorization(t *testing.T) {
	cmd := exec.Command("aserto", "developer", "start", "local", "--src-path=policy")

	err := cmd.Start()
	if err != nil {
		t.Error("failed to start aserto onebox:", err)
	}

	defer stopProcess(cmd)

	go main()

	r := runner{router: createRouter()}

	t.Run(
		"Should start with no boards",
		r.Test(
			Anonymous,
			httptest.NewRequest(http.MethodGet, "/boards", nil),
			AssertBoardCount(0),
		),
	)

	t.Run(
		"Jeff should be able to create a board",
		r.Test(
			Jeff,
			httptest.NewRequest(http.MethodPost, "/boards?name=first_board", nil),
			func(t *testing.T, status int, body []byte) {
				assert.Equal(t, http.StatusOK, status)

				var board Board

				err = json.Unmarshal(body, &board)
				assert.NoError(t, err, "can't load JSON response")

				assert.Equal(t, "first_board", board.Name)
				assert.Equal(t, Jeff, board.Owner)
			},
		),
	)

	t.Run(
		"Should have one board",
		r.Test(
			Ian,
			GetBoardsRequest(),
			AssertBoardCount(1),
		),
	)

	t.Run(
		"There should be no messages in the board",
		r.Test(
			Anonymous,
			httptest.NewRequest(http.MethodGet, "/boards/1/messages", nil),
			AssertMessageCount(0),
		),
	)

	t.Run(
		"Ian should be able to post a message",
		r.Test(
			Ian,
			httptest.NewRequest(http.MethodPost, "/boards/1/messages", strings.NewReader(`{"message": "hello"}`)),
			func(t *testing.T, status int, body []byte) {
				assert.Equal(t, http.StatusOK, status)
			},
		),
	)

	t.Run(
		"There should be one messages in the board",
		r.Test(
			Anonymous,
			httptest.NewRequest(http.MethodGet, "/boards/1/messages", nil),
			AssertMessageCount(1),
		),
	)

	t.Run(
		"Kim should not be able to delete Ian's message",
		r.Test(
			Kim,
			httptest.NewRequest(http.MethodDelete, "/boards/1/messages/2", nil),
			func(t *testing.T, status int, _ []byte) {
				assert.Equal(t, http.StatusUnauthorized, status)
			},
		),
	)

	t.Run(
		"Board owner should be able to delete Ian's message",
		r.Test(
			Jeff,
			httptest.NewRequest(http.MethodDelete, "/boards/1/messages/2", nil),
			func(t *testing.T, status int, _ []byte) {
				assert.Equal(t, http.StatusOK, status)
			},
		),
	)

	t.Run(
		"There should be no messages in the board",
		r.Test(
			Anonymous,
			httptest.NewRequest(http.MethodGet, "/boards/1/messages", nil),
			AssertMessageCount(0),
		),
	)
}

func GetBoardsRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/boards", nil)
}

func AssertBoardCount(expected int) ResponseValidator {
	return func(t *testing.T, status int, body []byte) {
		assert.Equal(t, http.StatusOK, status)

		var boards []Board

		err := json.Unmarshal(body, &boards)
		assert.NoError(t, err, "can't load JSON response")

		assert.Equal(t, expected, len(boards), fmt.Sprintf("there should be %d boards", expected))
	}
}

func AssertMessageCount(expected int) ResponseValidator {
	return func(t *testing.T, status int, body []byte) {
		assert.Equal(t, http.StatusOK, status)

		var messages []Message

		err := json.Unmarshal(body, &messages)
		assert.NoError(t, err, "can't load JSON response")

		assert.Equal(t, expected, len(messages), fmt.Sprintf("there should be %d messages", expected))
	}
}

type runner struct {
	router *mux.Router
}

type ResponseValidator func(*testing.T, int, []byte)

func (r runner) Test(user string, req *http.Request, validate ResponseValidator) func(*testing.T) {
	return func(t *testing.T) {
		if user != "" {
			req.SetBasicAuth(user, user)
		}

		w := httptest.NewRecorder()

		r.router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err, "failed to read response body")

		validate(t, res.StatusCode, body)
	}
}

func stopProcess(c *exec.Cmd) error {
	if err := c.Process.Signal(os.Interrupt); err != nil {
		if err := c.Process.Kill(); err != nil {
			return errors.Wrap(err, "unable to stop or kill process")
		}
	}

	if err := c.Wait(); err != nil {
		log.Print("failed to wait for process:", err)
	}

	return nil
}
