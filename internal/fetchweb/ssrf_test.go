package fetchweb

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
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

// F1b (t1968 sync-audit iter2): the shared address space 100.64.0.0/10
// (RFC 6598 CGNAT — the Tailscale mesh range, whose nodes are reachable
// internal machines, not public web servers) is NOT public: a redirect
// resolving there is refused at the hop's dial, exactly like the private and
// link-local ranges above. The refusal asserts the policy's own message —
// "non-public address" — so a kernel "connection refused" from an unguarded
// dial can never masquerade as a policy refusal.
func TestPageClient_RefusesCGNATRedirectTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://100.64.0.1/", http.StatusFound)
	}))
	defer srv.Close()
	// The context bounds the unguarded case (the RED state this test was
	// authored against): without the range refusal the hop dials for real, and
	// a silent target must not stretch the test to the client timeout.
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, err := FetchPage(ctx, srv.URL, newPageClient(harnessIPPolicy))
	if err == nil || !strings.Contains(err.Error(), "non-public address") {
		t.Fatalf("err = %v, want a dial-level non-public-address refusal of the CGNAT redirect target", err)
	}
}

// F1b: the IPv4 limited-broadcast address 255.255.255.255 is not a public
// web origin either; same dial-level refusal on the redirect hop.
func TestPageClient_RefusesBroadcastRedirectTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://255.255.255.255/", http.StatusFound)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, err := FetchPage(ctx, srv.URL, newPageClient(harnessIPPolicy))
	if err == nil || !strings.Contains(err.Error(), "non-public address") {
		t.Fatalf("err = %v, want a dial-level non-public-address refusal of the broadcast redirect target", err)
	}
}

// F1b unit boundaries: the /10 prefix is exact — its in-range edges and the
// limited-broadcast address are refused; 100.63.x and 100.128.x (public) and
// ordinary public addresses stay allowed. No existing refusal is weakened.
func TestRefuseNonPublicIP_CGNATAndBroadcastBoundaries(t *testing.T) {
	for _, s := range []string{
		"100.64.0.0", "100.64.0.1", "100.100.100.200", "100.127.255.255",
		"255.255.255.255",
	} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err == nil {
			t.Errorf("%s must be refused (CGNAT shared space / limited broadcast), got allowed", s)
		}
	}
	for _, s := range []string{"100.63.255.255", "100.128.0.0", "8.8.8.8", "93.184.216.34"} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err != nil {
			t.Errorf("%s is public, want allowed, got %v", s, err)
		}
	}
	// The pre-existing refusal set stays live alongside the new ranges.
	for _, s := range []string{"127.0.0.1", "10.0.0.1", "192.168.1.1", "169.254.169.254", "0.0.0.0", "::1", "fe80::1"} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err == nil {
			t.Errorf("%s must stay refused (pre-existing set weakened)", s)
		}
	}
}
