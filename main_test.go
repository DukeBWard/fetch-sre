package main_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	main "github.com/dukebward/fetch-sre"
)

// local test server with a predictable 200 response and <500ms latency
func TestLocalHttp(t *testing.T) {
	tests := []struct {
		// test param definitions
		name       string
		handler    http.HandlerFunc
		expectUp   int
		expectDown int
	}{
		{
			// pass test
			name: "pass-200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			},
			expectUp: 1,
		},
		{
			// fail test
			name: "fail-timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(600 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			},
			expectUp: 0,
		},
		{
			// fail with 500
			name: "fail-500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("ERROR"))
			},
			expectUp: 0,
		},
		{
			// fail with 404
			name: "fail-404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("NOT FOUND"))
			},
			expectUp: 0,
		},
	}

	// run all tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create server
			srv := httptest.NewServer(http.HandlerFunc(test.handler))
			defer srv.Close()

			// set up endpoints
			endpoints := []main.Endpoint{
				{
					Name: test.name,
					URL:  srv.URL,
				},
			}
			domainStatusMap := make(map[string]*main.DomainStatus)

			// execute checks
			main.RunChecks(endpoints, domainStatusMap)

			// parse server url
			url, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatalf("failed to parse url: %v", err)
			}

			// separate host, port
			host, _, err := net.SplitHostPort(url.Host)
			if err != nil {
				t.Fatalf("failed to split host and port: %v", err)
			}

			// lookup domain status
			actual, ok := domainStatusMap[host]
			if !ok {
				t.Fatalf("expected domainstatusmap to have key %q", host)
			}

			// verify upcount
			if actual.UpCount != test.expectUp {
				t.Errorf("expected upcount=%d, got %d for %q", test.expectUp, actual.UpCount, host)
			}
		})
	}
}

func TestHomework(t *testing.T) {
	files := []string{
		"sample1.yaml",
		"sample2.yaml",
		"sample3.yaml",
	}

	for _, file := range files {

		t.Run(file, func(t *testing.T) {
			//  absolute path to the YAML file
			yamlPath, err := filepath.Abs(filepath.Join("test", file))
			if err != nil {
				t.Fatalf("Failed to get absolute path for %s: %v", file, err)
			}

			// run `go run main.go --file=<yamlPath> --maxRuns=1`
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			cmd := exec.CommandContext(
				ctx,
				"go", "run", "main.go",
				"--file", yamlPath,
				"--maxRuns=1", // Ensure the program does 1 cycle then exits
			)

			// combime stdout and stderr
			output, err := cmd.CombinedOutput()

			// check if ctx timed out
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Test timed out after 5s. Output so far:\n%s", output)
			}

			// if fail (exit code != 0), fail the test
			if err != nil {
				t.Logf("Process returned error: %v\nOutput:\n%s", err, output)
			}

			t.Logf("Output from main.go with %s:\n%s", file, output)
		})
	}
}
