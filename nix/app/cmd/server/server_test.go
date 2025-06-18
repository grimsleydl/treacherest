package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupServer(t *testing.T) {
	// Create the server
	handler := SetupServer()
	
	if handler == nil {
		t.Fatal("SetupServer returned nil handler")
	}
	
	// Test that basic routes work
	testCases := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/room/INVALID", http.StatusNotFound},
		{"GET", "/game/INVALID", http.StatusNotFound},
		{"POST", "/room/new", http.StatusBadRequest}, // No player name
		{"GET", "/static/test.js", http.StatusNotFound}, // No actual file
	}
	
	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			
			handler.ServeHTTP(w, req)
			
			if w.Code != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, w.Code)
			}
		})
	}
}

func TestMainPackageSetup(t *testing.T) {
	// We can't easily test main() since it calls ListenAndServe
	// But we've tested SetupServer which contains all the logic
	// This test just ensures main.go compiles and the setup works
	
	t.Run("main function setup", func(t *testing.T) {
		// The fact that we can call SetupServer proves main.go works
		handler := SetupServer()
		if handler == nil {
			t.Fatal("SetupServer failed in main package")
		}
	})
}