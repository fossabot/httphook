package httphook_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/hourglassdesign/httphook"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func createEndpoint() {
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
	})

	go http.ListenAndServe(":8080", nil)
}

func TestNew(t *testing.T) {
	tt := []struct {
		Name     string
		Endpoint string
		Levels   []logrus.Level
	}{
		{Name: "test-service", Endpoint: "localhost:8080/logs", Levels: logrus.AllLevels},
	}

	for _, tc := range tt {
		hook := httphook.New(tc.Name, tc.Endpoint, tc.Levels)

		assert.NotNil(t, hook)
		assert.Equal(t, tc.Levels, hook.Levels())
	}
}

func TestHook_Fire(t *testing.T) {
	createEndpoint()

	tt := []struct {
		Name           string
		Endpoint       string
		ExpectedError  string
		ExpectedStatus int
		BeforeError    string
		AfterError     string
		Levels         []logrus.Level
		Entry          logrus.Entry
	}{
		// TEST CASE 1: Valid payload
		{
			Name:           "test-service",
			Endpoint:       "http://localhost:8080/logs",
			Levels:         logrus.AllLevels,
			ExpectedStatus: http.StatusOK,
			Entry: logrus.Entry{
				Message: "test-message",
				Data: logrus.Fields{
					"test-key": "test-value",
				},
				Time: time.Now(),
			},
		},
		// TEST CASE 2: Payload cannot be marshalled
		{
			Entry: logrus.Entry{
				Data: logrus.Fields{
					"test-key": make(chan int),
				},
			},
			ExpectedError: "failed to marshal payload due to error json: unsupported type: chan int",
		},
		// TEST CASE 3: Invalid request URI.
		{
			Entry: logrus.Entry{
				Message: "test-message",
				Data: logrus.Fields{
					"test-key": "test-value",
				},
				Time: time.Now(),
			},
			ExpectedError: "failed to build request due to error parse :: missing protocol scheme",
			Endpoint:      ":",
		},
		// TEST CASE 4: Error in BeforePost handler.
		{
			Entry: logrus.Entry{
				Message: "test-message",
				Data: logrus.Fields{
					"test-key": "test-value",
				},
				Time: time.Now(),
			},
			ExpectedError: "before-error",
			BeforeError:   "before-error",
		},
		// TEST CASE 5: Error in AfterPost handler.
		{
			Name:           "test-service",
			Endpoint:       "http://localhost:8080/logs",
			Levels:         logrus.AllLevels,
			ExpectedStatus: http.StatusOK,
			Entry: logrus.Entry{
				Message: "test-message",
				Data: logrus.Fields{
					"test-key": "test-value",
				},
				Time: time.Now(),
			},
			ExpectedError: "after-error",
			AfterError:    "after-error",
		},
		// TEST CASE 6: Endpoint returns a 404 with no after handler.
		{
			Name:     "test-service",
			Endpoint: "http://localhost:8080/invalid",
			Levels:   logrus.AllLevels,
			Entry: logrus.Entry{
				Message: "test-message",
				Data: logrus.Fields{
					"test-key": "test-value",
				},
				Time: time.Now(),
			},
			ExpectedError:  "failed to post payload, the server responded with a status of 404",
			ExpectedStatus: http.StatusNotFound,
		},
		// TEST CASE 7: Endpoint handler does not exist.
		{
			Name:     "test-service",
			Endpoint: "http://localhost:8081/invalid",
			Levels:   logrus.AllLevels,
			Entry: logrus.Entry{
				Message: "test-message",
				Data: logrus.Fields{
					"test-key": "test-value",
				},
				Time: time.Now(),
			},
			ExpectedError:  "failed to perform request due to error Post http://localhost:8081/invalid: dial tcp [::1]:8081: connectex: No connection could be made because the target machine actively refused it.",
			ExpectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tt {
		hook := httphook.New(tc.Name, tc.Endpoint, tc.Levels)

		assert.NotNil(t, hook)
		assert.Equal(t, tc.Levels, hook.Levels())

		hook.BeforePost = func(req *http.Request) error {
			method := req.Method
			name := req.Header.Get("service-name")
			content := req.Header.Get("content-type")

			assert.Equal(t, "POST", method)
			assert.Equal(t, tc.Name, name)
			assert.Equal(t, "application/json", content)

			if tc.BeforeError != "" {
				return errors.New(tc.BeforeError)
			}

			return nil
		}

		hook.AfterPost = func(res *http.Response) error {
			assert.Equal(t, tc.ExpectedStatus, res.StatusCode)

			if tc.AfterError != "" {
				return errors.New(tc.AfterError)
			}

			return nil
		}

		if err := hook.Fire(&tc.Entry); err != nil {
			assert.Equal(t, tc.ExpectedError, err.Error())
		}
	}
}
