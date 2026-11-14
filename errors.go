package types

import (
	"fmt"
	"net/http"
)

const (
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderRequestID     = "X-Request-ID"
)

type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type ErrorItem struct {
	Code            string       `json:"code"`
	Message         string       `json:"message"`
	CustomerActions []string     `json:"customerActions,omitempty"`
	SupportActions  []string     `json:"supportActions,omitempty"`
	Fields          []FieldError `json:"fields,omitempty"`
}

type ErrorEnvelope struct {
	RequestID     string      `json:"requestId,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
	Errors        []ErrorItem `json:"errors"`
}

type APIError struct {
	Status   int
	Envelope ErrorEnvelope
	Body     string
}

func (e *APIError) Error() string {
	if len(e.Envelope.Errors) > 0 {
		x := e.Envelope.Errors[0]
		return fmt.Sprintf("%d %s: %s", e.Status, x.Code, x.Message)
	}
	return fmt.Sprintf("%d %s", e.Status, http.StatusText(e.Status))
}

func (e *APIError) Code() string {
	if len(e.Envelope.Errors) == 0 {
		return ""
	}
	return e.Envelope.Errors[0].Code
}

func (e *APIError) NotFound() bool { return e.Status == http.StatusNotFound }

func (e *APIError) Conflict() bool { return e.Status == http.StatusConflict }

func (e *APIError) Invalid() bool { return e.Status == http.StatusBadRequest }

func (e *APIError) Unauthorized() bool {
	return e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden
}
