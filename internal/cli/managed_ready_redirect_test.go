package cli

// managed_ready_redirect_test.go — card t1410, F13 readiness redirect guard.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestManagedCodexAppReadyRedirectGuard pins F13: a loopback endpoint cannot
// bounce the readiness probe to a host the loopback rule would refuse.
func TestManagedCodexAppReadyRedirectGuard(t *testing.T) {
	wsOf := func(srv *httptest.Server) string { return strings.Replace(srv.URL, "http://", "ws://", 1) }
	redirectTo := func(target string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target, http.StatusFound)
		}))
	}
	refused := map[string]string{
		"non-loopback IP":   "http://10.1.2.3:9/readyz",
		"localhost by name": "http://localhost:9/readyz",
	}
	for name, target := range refused {
		t.Run(name, func(t *testing.T) {
			srv := redirectTo(target)
			defer srv.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()
			start := time.Now()
			err := managedCodexAppReady(ctx, wsOf(srv))
			if !errors.Is(err, errManagedCodexNonLoopback) {
				t.Fatalf("redirect to %s = %v, want the loopback refusal", target, err)
			}
			if elapsed := time.Since(start); elapsed > 700*time.Millisecond {
				t.Fatalf("refusal took %v, want a prompt return", elapsed)
			}
		})
	}
	t.Run("loopback redirect still allowed", func(t *testing.T) {
		ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer ok.Close()
		srv := redirectTo(ok.URL + "/readyz")
		defer srv.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := managedCodexAppReady(ctx, wsOf(srv)); err != nil {
			t.Fatalf("redirect to another loopback literal = %v, want nil", err)
		}
	})
}
