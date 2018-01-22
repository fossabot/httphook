// Package httphook contains types for creating a logrus hook to send logs
// via HTTP to a configured endpoint. See https://github.com/sirupsen/logrus
// for more details.
package httphook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	// Hook handles a log entry from logrus and attempts to post it to a configured
	// endpoint.
	Hook struct {
		client   http.Client
		levels   []logrus.Level
		endpoint string
		name     string

		// Method to be executed before performing the request.
		BeforePost BeforeFunc

		// Method to be executed after performing the request.
		AfterPost AfterFunc
	}

	// BeforeFunc is a convenience wrapper for the method executed
	// before a request is made.
	BeforeFunc func(req *http.Request) error

	// AfterFunc is a convenience wrapper for the method executed
	// after a request is made.
	AfterFunc func(res *http.Response) error

	// Log represents the format in which we want to post logging entries to the endpoint.
	Log struct {
		// The log's message string.
		Message string `json:"message"`

		// The log's additional metadata.
		Fields logrus.Fields `json:"fields"`

		// The time at which the log happened.
		Timestamp time.Time `json:"timestamp"`
	}
)

// New creates an instance of the Hook type, specifying the name of the application producing
// logs, the endpoint to post logs to & the logging levels for which requests will be made.
func New(name, endpoint string, levels []logrus.Level) *Hook {
	return &Hook{
		client:   http.Client{},
		levels:   levels,
		endpoint: endpoint,
		name:     name,
	}
}

// Levels returns a slice of all levels handled by this hook.
func (h Hook) Levels() []logrus.Level {
	return h.levels
}

// Fire handles forwarding a given logging entry to the configured destination via a
// HTTP POST request. If it fails to send the payload, an error is returned.
func (h Hook) Fire(entry *logrus.Entry) error {
	log := Log{
		Message:   entry.Message,
		Fields:    entry.Data,
		Timestamp: entry.Time,
	}

	// Convert the log to JSON.
	payload, err := json.Marshal(log)

	if err != nil {
		return fmt.Errorf("failed to marshal payload due to error %v", err)
	}

	body := bytes.NewBuffer(payload)

	// Create a request
	req, err := http.NewRequest("POST", h.endpoint, body)

	if err != nil {
		return fmt.Errorf("failed to build request due to error %v", err)
	}

	// Set appropriate headers so we can identify the service.
	req.Header.Add("service-name", h.name)
	req.Header.Add("content-type", "application/json")

	// Run the custom before post handler if it has been configured.
	if h.BeforePost != nil {
		if err := h.BeforePost(req); err != nil {
			return err
		}
	}

	// Send the request.
	resp, err := h.client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to perform request due to error %v", err)
	}

	// Run the custom after post handler if it has been configured.
	if h.AfterPost != nil {
		if err := h.AfterPost(resp); err != nil {
			return err
		}
	}

	// Return an error if the status code is greater than 201.
	if resp.StatusCode > http.StatusCreated {
		return fmt.Errorf("failed to post payload, the server responded with a status of %v", resp.StatusCode)
	}

	// Otherwise, success.
	return nil
}
