package httphook_test

import (
	"net/http"

	"github.com/hourglassdesign/httphook"
	"github.com/sirupsen/logrus"
)

func ExampleNew() {
	// Create a hook to post logs via HTTP.
	hook := httphook.New(
		"service-name",
		"localhost:8080/logs",
		logrus.AllLevels,
	)

	// Optionally perform processing on the HTTP request before a log is posted to the
	// configured endpoint. This is useful for adding custom headers etc.
	hook.BeforePost = func(req *http.Request) error {
		return nil
	}

	// Optionally perform processing on the HTTP response when a log is posted to the
	// configured endpoint. This is useful for error handling scenarios & debugging.
	hook.AfterPost = func(res *http.Response) error {
		return nil
	}

	// Register the hook with logrus.
	logrus.AddHook(hook)
}
