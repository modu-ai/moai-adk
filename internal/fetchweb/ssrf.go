// ssrf.go — F1 (t1968 sync-audit). Dial-level and redirect-level defenses
// for the agentic fetch surface: `moai fetch web` and the MCP fetch_web tool
// take their URL from MODEL OUTPUT, so a chosen or redirected URL must never
// reach loopback, private, or link-local space (including the cloud metadata
// range 169.254.0.0/16), whose content would then be forwarded verbatim to
// the external scoring endpoint.
//
// Defense in depth, two layers:
//
//   - DIAL: the client's dialer carries a Control hook that validates every
//     address AFTER DNS resolution and BEFORE the connect — so it sees every
//     redirect hop and every re-resolution, not just the URL's first lookup.
//   - REDIRECT: a CheckRedirect policy caps the chain and refuses any hop
//     whose scheme is not http/https; each hop's host is re-validated at dial
//     time by the Control hook.
//
// FetchPage itself stays client-agnostic (its injection seam is what lets
// tests run against fake servers, REQ-FW-010); the guard ships in the client
// the production call sites construct — NewPageClient — so both the CLI
// command and the MCP tool carry it identically.
package fetchweb

import (
	"fmt"
	"net"
	"net/http"
	"syscall"
)

// MaxRedirectHops caps the redirect chain the guarded page client follows
// (F1b). The stock client would follow ten; five is the whole range a
// legitimate page-handoff needs, and a shorter cap bounds redirect-based
// probing. Exported so the contract is nameable where the client is wired.
const MaxRedirectHops = 5

// ipPolicy decides whether the page fetcher may connect to ip. A nil ip is
// always refused (fail closed): it means the dialed address did not parse.
type ipPolicy func(ip net.IP) error

// refuseNonPublicIP is the production policy (F1a): the fetch surface
// connects to PUBLIC addresses only. Refused ranges: loopback, RFC1918
// private and RFC4193 unique-local (net.IP.IsPrivate covers both), the RFC
// 6598 shared address space 100.64.0.0/10 (F1b — CGNAT carrier ranges and
// the Tailscale mesh; reachable internal nodes, not public web servers),
// link-local unicast and multicast (169.254.0.0/16 — the cloud metadata
// range — and fe80::/10), unspecified, multicast, and the IPv4
// limited-broadcast address. IPv4-mapped IPv6 forms normalize through
// net.ParseIP before the checks, so ::ffff:127.0.0.1 is caught.
func refuseNonPublicIP(ip net.IP) error {
	if ip != nil && !ip.IsLoopback() && !ip.IsPrivate() && !isSharedAddressSpace(ip) &&
		!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsUnspecified() &&
		!ip.IsMulticast() && !ip.Equal(net.IPv4bcast) {
		return nil
	}
	return fmt.Errorf("non-public address %s refused", ip)
}

// isSharedAddressSpace reports whether ip is inside the RFC 6598 shared
// address space 100.64.0.0/10 (F1b): 100.64.0.0–100.127.255.255. No net.IP
// helper covers it — IsPrivate stops at RFC1918/RFC4193 — so the /10 prefix
// is matched on the 4-byte form directly (first octet 100, top two bits of
// the second octet fixed at 01).
func isSharedAddressSpace(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 100 && ip4[1]&0xC0 == 0x40
	}
	return false
}

// harnessIPPolicy is the TEST-HARNESS policy (the override seam): loopback is
// admitted because httptest servers bind 127.0.0.1, and NOTHING else changes
// — private, link-local, and every other refused range stay refused under it,
// which is what the redirect-refusal tests prove against.
func harnessIPPolicy(ip net.IP) error {
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return refuseNonPublicIP(ip)
}

// dialControl builds the Control hook that applies policy to every address
// the dialer is about to connect to. SplitHostPort/ParseIP failure also
// refuses (fail closed): the hook must never wave through an address it
// could not interpret.
func dialControl(policy ipPolicy) func(network, address string, _ syscall.RawConn) error {
	return func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("fetch web: refusing dial to unparseable address %q", address)
		}
		if err := policy(net.ParseIP(host)); err != nil {
			return fmt.Errorf("fetch web: %w", err)
		}
		return nil
	}
}

// checkRedirect is the redirect policy (F1b): cap the chain, refuse
// non-HTTP(S) schemes on every hop. Host validation happens at dial time —
// the hop dials again, so the Control hook re-validates its resolved IP.
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= MaxRedirectHops {
		return fmt.Errorf("fetch web: redirect chain past the %d-hop cap refused", MaxRedirectHops)
	}
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return fmt.Errorf("fetch web: redirect to non-HTTP(S) scheme %q refused", req.URL.Scheme)
	}
	return nil
}

// NewPageClient is the guarded page client the production call sites (the
// CLI command and the MCP fetch_web tool) construct: the fetch timeout, a
// redirect chain capped at MaxRedirectHops with per-hop scheme checks, and a
// dialer whose Control hook validates every resolved IP against
// refuseNonPublicIP — every redirect hop and every DNS re-resolution
// included. No proxy is configured: through a proxy the dial connects to the
// PROXY's address, and the dial-level target validation would never see the
// real destination — the guard would be silently void.
//
// @MX:ANCHOR: [AUTO] SSRF guard SSOT — both production fetch surfaces construct their page client here
// @MX:REASON: [AUTO] a call site building its own http.Client would drop the dial-level IP validation and redirect policy (F1); NewPageClient is the single construction point the CLI command and MCP fetch_web tool share.
func NewPageClient() *http.Client {
	return newPageClient(refuseNonPublicIP)
}

// newPageClient is the construction seam: NewPageClient with a replaceable IP
// policy, so the test harness can admit loopback (httptest binds 127.0.0.1)
// while keeping every other refused range live. Production code passes
// refuseNonPublicIP and nothing else.
func newPageClient(policy ipPolicy) *http.Client {
	dialer := &net.Dialer{
		Timeout:   FetchTimeout,
		KeepAlive: 30_000_000_000, // 30s, mirroring http.DefaultTransport's dialer
		Control:   dialControl(policy),
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = dialer.DialContext
	t.DialTLSContext = nil // nil rides DialContext, so the Control hook guards HTTPS dials too
	t.Proxy = nil          // see NewPageClient: a proxy hop would bypass dial-level validation
	return &http.Client{
		Timeout:       FetchTimeout,
		Transport:     t,
		CheckRedirect: checkRedirect,
	}
}
