# httphook
[![CircleCI](https://img.shields.io/circleci/project/github/hourglassdesign/httphook.svg)](https://circleci.com/gh/hourglassdesign/httphook)
[![GoDoc](https://godoc.org/github.com/hourglassdesign/httphook?status.svg)](http://godoc.org/github.com/hourglassdesign/httphook)
[![Go Report Card](https://goreportcard.com/badge/github.com/hourglassdesign/httphook)](https://goreportcard.com/report/github.com/hourglassdesign/httphook)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/hourglassdesign/httphook/release/LICENSE)

A simple logrus hook for forwarding logs via HTTP.

## usage
```go
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
```