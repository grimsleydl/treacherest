package main

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestMainSubprocess tests the main function using subprocess
func TestMainSubprocess(t *testing.T) {
	if os.Getenv("BE_SUBPROCESS") == "1" {
		// We're in the subprocess, run main
		main()
		return
	}

	// Test with default port
	t.Run("default port", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMainSubprocess")
		cmd.Env = append(os.Environ(), "BE_SUBPROCESS=1")

		// Capture output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		// Give server time to start and check output
		time.Sleep(200 * time.Millisecond)

		// Read the output to find the port
		scanner := bufio.NewScanner(stdout)
		serverStarted := false
		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "Starting server on :8080") {
					serverStarted = true
					break
				}
			}
		}()

		// Wait a bit for the log line
		time.Sleep(100 * time.Millisecond)

		if !serverStarted {
			// Server might have started without us catching the log
			// Try to connect anyway
		}

		// Test that server is accessible
		resp, err := http.Get("http://localhost:8080/")
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Clean up
		cancel()
		cmd.Wait()
	})

	// Test with custom PORT env var
	t.Run("custom port", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMainSubprocess")
		cmd.Env = append(os.Environ(), "BE_SUBPROCESS=1", "PORT=8081")

		// Capture output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		// Give server time to start and check output
		time.Sleep(200 * time.Millisecond)

		// Read the output to find the port
		scanner := bufio.NewScanner(stdout)
		serverStarted := false
		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "Starting server on :8081") {
					serverStarted = true
					break
				}
			}
		}()

		// Wait a bit for the log line
		time.Sleep(100 * time.Millisecond)

		if !serverStarted {
			// Server might have started without us catching the log
			// Try to connect anyway
		}

		// Test that server is accessible on custom port
		resp, err := http.Get("http://localhost:8081/")
		if err != nil {
			t.Fatalf("Failed to connect to server on custom port: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Clean up
		cancel()
		cmd.Wait()
	})
}

// TestMainFunctionErrors tests error handling in main
func TestMainFunctionErrors(t *testing.T) {
	if os.Getenv("BE_SUBPROCESS_ERROR") == "1" {
		// Try to start server on an invalid port
		os.Setenv("PORT", "99999")
		main()
		return
	}

	t.Run("invalid port", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainFunctionErrors")
		cmd.Env = append(os.Environ(), "BE_SUBPROCESS_ERROR=1")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatal("Expected main to fail with invalid port")
		}

		// Check that it logged the error
		if !strings.Contains(string(output), "listen tcp") {
			t.Fatalf("Expected listen error, got: %s", output)
		}
	})
}

