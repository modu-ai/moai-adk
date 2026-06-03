package web

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

// waitForAddr blocks until the server's listener is bound (Addr non-empty) or
// fails the test after a short timeout.
func waitForAddr(t *testing.T, srv *Server) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if srv.Addr() != "" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("server listener did not bind within timeout")
}

// httpGet performs a GET with a bounded timeout.
func httpGet(t *testing.T, url string) (*http.Response, error) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	return client.Get(url)
}

// atoiOrZero parses s as an int, returning 0 on error.
func atoiOrZero(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
