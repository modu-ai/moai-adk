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

// P2 (card t1968 codex review, repair round): refuseNonPublicIP misses the
// remaining IANA-registered special-purpose ranges — IPv4 192.0.0.0/24 (IETF
// Protocol Assignments), 192.88.99.0/24 (Deprecated 6to4 Relay Anycast),
// 198.18.0.0/15 (Benchmarking), 240.0.0.0/4 (Reserved), and IPv6 100::/64
// (Discard-Only), 2001::/23 (IETF Protocol Assignments, incl. Teredo
// 2001::/32), 2001:db8::/32 (Documentation), 2002::/16 (6to4). Verified
// against the IANA IPv4/IPv6 Special-Purpose Address Registries (retrieved
// 2026-10-28). None of these is a public web origin; the benchmarking range
// especially is a documented bypass target. 64:ff9b::/96 (NAT64) is NOT here:
// F1 (card t1968 iter15) decides it per embedded IPv4 — see
// TestRefuseNonPublicIP_NAT64EmbeddedPolicy.
func TestRefuseNonPublicIP_IANASpecialPurposeRanges(t *testing.T) {
	for _, s := range []string{
		"198.18.0.1",       // 198.18.0.0/15 benchmarking
		"198.19.255.1",     // 198.18.0.0/15 far edge
		"192.0.0.1",        // 192.0.0.0/24 IETF protocol assignments
		"192.88.99.1",      // 192.88.99.0/24 deprecated 6to4 relay anycast
		"240.0.0.1",        // 240.0.0.0/4 reserved
		"250.1.2.3",        // 240.0.0.0/4 far edge
		"100::1",           // discard-only
		"2001::1",          // 2001::/23 IETF protocol assignments (Teredo)
		"2001:db8::1",      // documentation
		"2001:db8:ffff::1", // documentation far edge
		"2002::1",          // 6to4
		"2002:c000:0201::1",
	} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err == nil {
			t.Errorf("%s must be refused (IANA special-purpose range), got allowed", s)
		}
	}
	// The boundary OUTSIDE each new range stays public: 198.16.x and 198.20.x
	// flank the /15 benchmarking block, 2001:0200:: sits just past 2001::/23
	// (APNIC unicast), 2001:dc8:: just past the documentation /32, and 2003::
	// past the 6to4 /16. Over-refusing here would block real destinations.
	for _, s := range []string{"198.16.255.1", "198.20.0.1", "2001:0200::1", "2001:dc8::1", "2003::1", "8.8.8.8"} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err != nil {
			t.Errorf("%s is public, want allowed, got %v", s, err)
		}
	}
}

// F1 (card t1968 iter15, operator-approved repair): RFC 6052 NAT64. An
// address inside the well-known prefix 64:ff9b::/96 embeds an IPv4
// destination in its low 32 bits, and a DNS64/NAT64 IPv6-only client
// legitimately dials such an address for a PUBLIC embedded IPv4 — so the
// prefix must not be refused wholesale; the EMBEDDED address decides,
// through the same public-address predicate (no second policy). BOTH
// directions are asserted: a public embed passes (the base 41fe0c5396
// behavior the wholesale refusal regressed), and embeds of every refused
// range — loopback, RFC1918, CGNAT, link-local, IANA special — stay
// refused. 64:ff9b::192.0.2.128 moves here from the IANA-range test: its
// embedded 192.0.2.128 (RFC 5737 documentation) passes the pre-existing
// IPv4 predicate by the documented round-6 scope decision, so the NAT64
// form passes with it.
func TestRefuseNonPublicIP_NAT64EmbeddedPolicy(t *testing.T) {
	for _, s := range []string{
		"64:ff9b::93.184.216.34",    // public embed (example.com), dotted form
		"64:ff9b:0:0:0:0:5db8:d822", // same address, hex form
		"64:ff9b::192.0.2.128",      // RFC 5737 docs embed: passes the IPv4 predicate
	} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err != nil {
			t.Errorf("%s embeds a public IPv4, want allowed, got %v", s, err)
		}
	}
	for _, s := range []string{
		"64:ff9b::127.0.0.1",       // embedded loopback — the hole the wholesale refusal closed
		"64:ff9b::10.0.0.1",        // embedded RFC1918 private
		"64:ff9b::192.168.1.1",     // embedded RFC1918 private
		"64:ff9b::100.64.0.1",      // embedded RFC 6598 CGNAT
		"64:ff9b::169.254.169.254", // embedded link-local (cloud metadata)
		"64:ff9b::198.18.0.1",      // embedded IANA benchmarking special
	} {
		if err := refuseNonPublicIP(net.ParseIP(s)); err == nil {
			t.Errorf("%s embeds a refused IPv4, want refused, got allowed", s)
		}
	}
}

// F1: the dial guard routes NAT64 forms through that same policy — the
// Control hook parses the dial address and hands the IP to refuseNonPublicIP
// (production path of NewPageClient), so the embedded address, not the
// prefix, decides before any connect.
func TestDialControl_NAT64RoutesThroughPolicy(t *testing.T) {
	hook := dialControl(refuseNonPublicIP)
	if err := hook("tcp", "[64:ff9b::93.184.216.34]:443", nil); err != nil {
		t.Errorf("dial of a NAT64 public embed must pass the guard, got %v", err)
	}
	if err := hook("tcp", "[64:ff9b::127.0.0.1]:80", nil); err == nil || !strings.Contains(err.Error(), "non-public") {
		t.Errorf("err = %v, want a non-public refusal of the NAT64 loopback embed", err)
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
