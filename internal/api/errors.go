package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type APIError struct {
	StatusCode        int
	Code              string
	Message           string
	RequestID         string
	DocsURL           string
	Param             string
	RetryAfterSeconds int // populated on 429
}

func (e *APIError) Error() string {
	base := fmt.Sprintf("monit.rs API error: %s (status=%d, code=%q)", e.Message, e.StatusCode, e.Code)
	if e.RequestID != "" {
		base += fmt.Sprintf(" request_id=%s", e.RequestID)
	}
	return base
}

func (e *APIError) IsNotFound() bool    { return e.StatusCode == 404 }
func (e *APIError) IsRateLimited() bool { return e.StatusCode == 429 }
func (e *APIError) IsScopeError() bool  { return e.StatusCode == 403 && e.Code == "scope_missing" }
func (e *APIError) IsAuthError() bool   { return e.StatusCode == 401 }

type errorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		DocsURL   string `json:"docs_url"`
		RequestID string `json:"request_id"`
		Param     string `json:"param"`
	} `json:"error"`
}

func parseErrorResponse(resp *http.Response) *APIError {
	e := &APIError{StatusCode: resp.StatusCode, Message: resp.Status}
	e.RequestID = resp.Header.Get("Request-Id")
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if n, err := strconv.Atoi(ra); err == nil {
			e.RetryAfterSeconds = n
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return e
	}
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error.Code != "" {
		e.Code = env.Error.Code
		if env.Error.Message != "" {
			e.Message = env.Error.Message
		}
		if env.Error.RequestID != "" {
			e.RequestID = env.Error.RequestID
		}
		e.DocsURL = env.Error.DocsURL
		e.Param = env.Error.Param
	}
	return e
}
