package fetchweb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// ssrf_test.go — F1 (t1968 sync-audit). The guarded page client's IP and
// redirect defenses, proven against the test harness ONLY: every server here
// is an httptest instance bound to 127.0.0.1, reached through the
// newPageClient construction seam with the loopback-admitting harness policy
// — EXCEPT the loopback-refusal case itself, which runs the production
// NewPageClient unchanged. No real metadata or private endpoint is contacted
// (REQ-FW-010): refusals fire at dial/redirect time, before any connect.

// TestPageClient_RefusesLoopbackTarget: the production client (strict policy)
// refuses a loopback target — the shape an agentic caller would use to steer
// the fetch at a local service.
func TestPageClient_RefusesLoopbackTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body><p>never served</p></body></html>"))
	}))
	defer srv.Close()
	_, err := FetchPage(t.Context(), srv.URL, NewPageClient())
	if err == nil || !strings.Contains(err.Error(), "refus") {
		t.Fatalf("err = %v, want a dial-level refusal of the loopback target", err)
	}
}

// TestPageClient_HarnessAllowedTargetFetches: the positive control — under
// the harness policy (loopback admitted, everything else still refused) an
// allowed target fetches normally, so the guard passes what it should.
func TestPageClient_HarnessAllowedTargetFetches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body><p>harness page</p></body></html>"))
	}))
	defer srv.Close()
	page, err := FetchPage(t.Context(), srv.URL, newPageClient(harnessIPPolicy))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(page, "harness page") {
		t.Errorf("page = %q, want the served body through the guarded client", page)
	}
}

// TestPageClient_RefusesRedirectToPrivateTarget: a redirect whose target
// resolves into a refused range (RFC1918 192.168.0.1) is refused at the
// redirect hop's dial — DNS rebinding across hops cannot slip past, every
// hop's connected IP is validated.
func TestPageClient_RefusesRedirectToPrivateTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://192.168.0.1/", http.StatusFound)
	}))
	defer srv.Close()
	_, err := FetchPage(t.Context(), srv.URL, newPageClient(harnessIPPolicy))
	if err == nil || !strings.Contains(err.Error(), "refus") {
		t.Fatalf("err = %v, want a dial-level refusal of the private redirect target", err)
	}
}

// TestPageClient_RefusesLinkLocalRedirectTarget: the same refusal for the
// link-local range — 169.254.169.254, the cloud metadata address.
func TestPageClient_RefusesLinkLocalRedirectTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer srv.Close()
	_, err := FetchPage(t.Context(), srv.URL, newPageClient(harnessIPPolicy))
	if err == nil || !strings.Contains(err.Error(), "refus") {
		t.Fatalf("err = %v, want a dial-level refusal of the link-local redirect target", err)
	}
}

// TestPageClient_RefusesNonHTTPRedirectScheme: a redirect to a non-HTTP(S)
// scheme (ftp here) is refused by the redirect policy before any dial of the
// new scheme's handler.
func TestPageClient_RefusesNonHTTPRedirectScheme(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "ftp://127.0.0.1/pub", http.StatusFound)
	}))
	defer srv.Close()
	_, err := FetchPage(t.Context(), srv.URL, newPageClient(harnessIPPolicy))
	if err == nil || !strings.Contains(err.Error(), "non-HTTP(S)") {
		t.Fatalf("err = %v, want a non-HTTP(S) scheme redirect refusal", err)
	}
}

// TestPageClient_RefusesRedirectChainOverCap: an endless redirect chain is
// cut at MaxRedirectHops.
func TestPageClient_RefusesRedirectChainOverCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := 0
		if v, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
			n = v
		}
		http.Redirect(w, r, fmt.Sprintf("/%d", n+1), http.StatusFound)
	}))
	defer srv.Close()
	_, err := FetchPage(t.Context(), srv.URL+"/0", newPageClient(harnessIPPolicy))
	if err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("err = %v, want a redirect-cap refusal", err)
	}
}
