package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSSERequest(t *testing.T) {
	// Create a test handler that just returns OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with validation middleware
	handler := ValidateSSERequest(testHandler)

	tests := []struct {
		name           string
		queryString    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid datastar parameter",
			queryString:    "datastar=%7B%22theme%22%3A%22dark%22%7D",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "no parameters",
			queryString:    "",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "invalid parameter",
			queryString:    "invalid=test",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameter",
		},
		{
			name:           "multiple invalid parameters",
			queryString:    "invalid1=test&invalid2=test",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameter",
		},
		{
			name:           "valid and invalid parameters mixed",
			queryString:    "datastar=%7B%22theme%22%3A%22dark%22%7D&invalid=test",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameter",
		},
		{
			name:           "datastar parameter too large",
			queryString:    "datastar=" + strings.Repeat("a", 8193),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Datastar state too large",
		},
		{
			name:           "query string too large",
			queryString:    "datastar=" + strings.Repeat("a", 10001),
			expectedStatus: http.StatusRequestURITooLong,
			expectedBody:   "Query string too large",
		},
		{
			name:           "multiple datastar values",
			queryString:    "datastar=value1&datastar=value2",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid datastar parameter",
		},
		{
			name:           "malformed query string",
			queryString:    "datastar=%",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid query parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/sse/test?"+tt.queryString, nil)
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestValidateSSERequest_EdgeCases(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ValidateSSERequest(testHandler)

	t.Run("empty datastar parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sse/test?datastar=", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("datastar with complex JSON", func(t *testing.T) {
		// Simulate a real Datastar query parameter with nested JSON
		complexJSON := `{"theme":"dark","isStarting":false,"canStartGame":true,"validationMessage":"","cardId":"card123","accordionLeader":false}`
		encodedJSON := "datastar=" + strings.ReplaceAll(strings.ReplaceAll(complexJSON, `"`, "%22"), ":", "%3A")
		
		req := httptest.NewRequest("GET", "/sse/test?"+encodedJSON, nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("datastar with invalid signal name", func(t *testing.T) {
		// Try to inject an unauthorized signal
		invalidJSON := `{"theme":"dark","isStarting":false,"maliciousSignal":"hack"}`
		encodedJSON := "datastar=" + strings.ReplaceAll(strings.ReplaceAll(invalidJSON, `"`, "%22"), ":", "%3A")
		
		req := httptest.NewRequest("GET", "/sse/test?"+encodedJSON, nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid signal in datastar")
	})

	t.Run("datastar with all valid signals", func(t *testing.T) {
		// Test with many valid signals
		validJSON := `{"theme":"dark","isStarting":false,"canStartGame":true,"countdown":5,"accordionLeader":true,"allowLeaderless":false,"qrCode":"test"}`
		encodedJSON := "datastar=" + url.QueryEscape(validJSON)
		
		req := httptest.NewRequest("GET", "/sse/test?"+encodedJSON, nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("datastar with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sse/test?datastar=%7Binvalid%20json", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid datastar JSON")
	})
}