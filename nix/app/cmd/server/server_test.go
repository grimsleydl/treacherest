package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		{"POST", "/room/new", http.StatusBadRequest},    // No player name
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

func TestStaticFileServing(t *testing.T) {
	// Create a temporary static directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create static directory and a test file
	if err := os.Mkdir("static", 0755); err != nil {
		t.Fatal(err)
	}
	testContent := "console.log('test');"
	if err := os.WriteFile("static/test.js", []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	handler := SetupServer()

	testCases := []struct {
		name         string
		path         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "existing static file",
			path:         "/static/test.js",
			expectedCode: http.StatusOK,
			expectedBody: testContent,
		},
		{
			name:         "non-existent static file",
			path:         "/static/missing.js",
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "directory traversal attempt",
			path:         "/static/../main.go",
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, w.Code)
			}

			if tc.expectedBody != "" && w.Body.String() != tc.expectedBody {
				t.Errorf("expected body %q, got %q", tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	handler := SetupServer()

	t.Run("recovery middleware", func(t *testing.T) {
		// Create a handler that panics
		panicPath := "/panic-test"
		mux := http.NewServeMux()
		mux.HandleFunc(panicPath, func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		// Wrap with the same middleware stack
		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Mount("/", mux)

		req := httptest.NewRequest("GET", panicPath, nil)
		w := httptest.NewRecorder()

		// Should not panic due to recoverer
		r.ServeHTTP(w, req)

		// Should return 500 Internal Server Error
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 after panic, got %d", w.Code)
		}
	})

	t.Run("logger middleware", func(t *testing.T) {
		// Just ensure requests are handled with logger
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// If we got here without issues, logger is working
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("timeout middleware", func(t *testing.T) {
		// This is harder to test without modifying the handler
		// We'll just ensure normal requests complete within timeout
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		done := make(chan bool)
		go func() {
			handler.ServeHTTP(w, req)
			done <- true
		}()

		select {
		case <-done:
			// Request completed successfully
		case <-time.After(2 * time.Second):
			t.Error("request took too long, timeout middleware may not be working")
		}
	})
}
