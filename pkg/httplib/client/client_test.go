package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type TestResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, Init) {
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	init := Init{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	return server, init
}

func TestGetJSON(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   *TestResponse
		expectedError  error
		init           Init
	}{
		{
			name: "successful request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				enc := json.NewEncoder(w)
				_ = enc.Encode(TestResponse{
					Message: "success",
					Code:    200,
				})

			}),
			expectedStatus: http.StatusOK,
			expectedBody: &TestResponse{
				Message: "success",
				Code:    200,
			},
		},
		{
			name: "server error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				enc := json.NewEncoder(w)
				_ = enc.Encode(TestResponse{
					Message: "server error",
					Code:    500,
				})
			}),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  ErrRequestFailed,
		},
		{
			name: "invalid json response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("invalid json"))
			}),
			expectedStatus: http.StatusOK,
			expectedError:  errors.New("failed to unmarshal response: invalid character 'i' looking for beginning of value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, init := setupTestServer(t, tt.handler)
			if tt.init.Headers != nil {
				init.Headers = tt.init.Headers
			}

			resp, err := GetJSON[TestResponse](context.Background(), "/test", init)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error containing %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != nil {
				if resp.Body.Message != tt.expectedBody.Message {
					t.Errorf("expected message %q, got %q", tt.expectedBody.Message, resp.Body.Message)
				}
				if resp.Body.Code != tt.expectedBody.Code {
					t.Errorf("expected code %d, got %d", tt.expectedBody.Code, resp.Body.Code)
				}
			}
		})
	}
}

func TestPostJSON(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		requestBody    interface{}
		expectedStatus int
		expectedBody   *TestResponse
		expectedError  error
	}{
		{
			name: "successful post",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				enc := json.NewEncoder(w)
				_ = enc.Encode(TestResponse{
					Message: "created",
					Code:    201,
				})
			}),
			requestBody: TestResponse{
				Message: "test",
				Code:    100,
			},
			expectedStatus: http.StatusOK,
			expectedBody: &TestResponse{
				Message: "created",
				Code:    201,
			},
		},
		{
			name: "bad request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				enc := json.NewEncoder(w)
				_ = enc.Encode(TestResponse{
					Message: "bad request",
					Code:    400,
				})
			}),
			requestBody:    TestResponse{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrRequestFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, init := setupTestServer(t, tt.handler)
			init.Body = tt.requestBody

			resp, err := PostJSON[TestResponse](context.Background(), "/test", init)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error containing %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != nil {
				if resp.Body.Message != tt.expectedBody.Message {
					t.Errorf("expected message %q, got %q", tt.expectedBody.Message, resp.Body.Message)
				}
				if resp.Body.Code != tt.expectedBody.Code {
					t.Errorf("expected code %d, got %d", tt.expectedBody.Code, resp.Body.Code)
				}
			}
		})
	}
}

func TestGetBytes(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   []byte
		expectedError  error
	}{
		{
			name: "successful request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("test data"))
			}),
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("test data"),
		},
		{
			name: "empty response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
			expectedError:  ErrInvalidResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, init := setupTestServer(t, tt.handler)

			resp, err := GetBytes(context.Background(), "/test", init)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error containing %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != nil {
				if string(resp.Body) != string(tt.expectedBody) {
					t.Errorf("expected body %q, got %q", string(tt.expectedBody), string(resp.Body))
				}
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		url           string
		init          Init
		expectedError error
	}{
		{
			name:          "nil context",
			ctx:           nil,
			url:           "http://example.com/test",
			expectedError: ErrNilContext,
		},
		{
			name:          "empty url",
			ctx:           context.Background(),
			url:           "",
			expectedError: ErrEmptyURL,
		},
		{
			name: "negative timeout",
			ctx:  context.Background(),
			url:  "http://example.com/test",
			init: Init{
				Timeout: -1 * time.Second,
				BaseURL: "http://example.com",
			},
			expectedError: ErrInvalidTimeout,
		},
		{
			name: "negative retry count",
			ctx:  context.Background(),
			url:  "http://example.com/test",
			init: Init{
				RetryCount: -1,
				BaseURL:    "http://example.com",
			},
			expectedError: ErrInvalidRetryCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetJSON[TestResponse](tt.ctx, tt.url, tt.init)

			if err == nil {
				t.Errorf("expected error %v, got nil", tt.expectedError)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError.Error()) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}
