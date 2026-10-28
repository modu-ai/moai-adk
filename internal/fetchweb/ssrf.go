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
// range — and fe80::/10), unspecified, multicast, the IPv4
// limited-broadcast address, and the remaining IANA-registered
// special-purpose ranges (isIANASpecial — P2, card t1968 codex review).
// IPv4-mapped IPv6 forms normalize through net.ParseIP before the checks,
// so ::ffff:127.0.0.1 is caught. An address inside the RFC 6052 NAT64
// well-known prefix 64:ff9b::/96 is decided on its EMBEDDED IPv4 (F1, card
// t1968 iter15): the embedded address is the real destination a DNS64
// client reaches, so it — not the prefix — runs through this same
// predicate; an embedded loopback/private/CGNAT/link-local stays refused,
// an embedded public address passes.
func refuseNonPublicIP(ip net.IP) error {
	if ip != nil {
		if embedded := nat64Embedded(ip); embedded != nil {
			return refuseNonPublicIP(embedded)
		}
		if !ip.IsLoopback() && !ip.IsPrivate() && !isSharedAddressSpace(ip) &&
			!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsUnspecified() &&
			!ip.IsMulticast() && !ip.Equal(net.IPv4bcast) && !isIANASpecial(ip) {
			return nil
		}
	}
	return fmt.Errorf("non-public address %s refused", ip)
}

// ianaIPv4Special holds the IANA-registered IPv4 special-purpose ranges no
// net.IP helper covers (P2, card t1968 codex review; verified against the
// IANA IPv4 Special-Purpose Address Registry,
// iana.org/assignments/iana-ipv4-special-registry — retrieved 2026-10-28):
// 192.0.0.0/24 IETF Protocol Assignments, 192.88.99.0/24 Deprecated (6to4
// Relay Anycast), 198.18.0.0/15 Benchmarking (a documented SSRF filter
// bypass target), 240.0.0.0/4 Reserved. The registry's other IPv4 entries
// are already refused above: loopback, RFC1918, CGNAT, link-local,
// multicast, unspecified, limited broadcast. 192.0.2.0/24 and friends (RFC
// 5737 documentation) sit outside this round's reviewed additions and stay
// governed by the pre-existing policy.
var ianaIPv4Special = []struct {
	prefix [4]byte
	mask   [4]byte
}{
	{prefix: [4]byte{192, 0, 0, 0}, mask: [4]byte{255, 255, 255, 0}},   // 192.0.0.0/24
	{prefix: [4]byte{192, 88, 99, 0}, mask: [4]byte{255, 255, 255, 0}}, // 192.88.99.0/24
	{prefix: [4]byte{198, 18, 0, 0}, mask: [4]byte{255, 254, 0, 0}},    // 198.18.0.0/15
	{prefix: [4]byte{240, 0, 0, 0}, mask: [4]byte{240, 0, 0, 0}},       // 240.0.0.0/4
}

// ianaIPv6Special holds the IANA-registered IPv6 special-purpose ranges no
// net.IP helper covers (same review; verified against the IANA IPv6
// Special-Purpose Address Registry — retrieved 2026-10-28): 100::/64
// Discard-Only, 2001::/23 IETF Protocol Assignments (Teredo 2001::/32
// included), 2001:db8::/32 Documentation, 2002::/16 6to4. 2001:db8::/32
// sits OUTSIDE 2001::/23 (0x0db8 > 0x01ff), so it needs its own entry.
// Unique-local (fc00::/7) is IsPrivate; link-local fe80::/10 and multicast
// ff00::/8 are covered above. 64:ff9b::/96 (NAT64) is NOT here — F1 (card
// t1968 iter15) removed the wholesale refusal: an address in that prefix
// embeds a per-address IPv4 destination, decided by nat64Embedded below.
var ianaIPv6Special = []struct {
	prefix [16]byte
	mask   [16]byte
}{
	// 100::/64 — first 8 bytes.
	{prefix: [16]byte{0x01, 0x00}, mask: [16]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	// 2001::/23 — first 23 bits: mask ff ff fe 00.
	{prefix: [16]byte{0x20, 0x01}, mask: [16]byte{0xff, 0xff, 0xfe}},
	// 2001:db8::/32 — first 4 bytes.
	{prefix: [16]byte{0x20, 0x01, 0x0d, 0xb8}, mask: [16]byte{0xff, 0xff, 0xff, 0xff}},
	// 2002::/16 — first 2 bytes.
	{prefix: [16]byte{0x20, 0x02}, mask: [16]byte{0xff, 0xff}},
}

// isIANASpecial reports whether ip is inside one of the IANA-registered
// special-purpose ranges above — prefix/mask byte comparison in the
// isSharedAddressSpace style, table-driven. IPv4-mapped IPv6 forms take the
// IPv4 table via To4, matching the policy's normalization.
func isIANASpecial(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		for _, r := range ianaIPv4Special {
			if matchesSpecialPrefix(ip4, r.prefix[:], r.mask[:]) {
				return true
			}
		}
		return false
	}
	ip16 := ip.To16()
	for _, r := range ianaIPv6Special {
		if matchesSpecialPrefix(ip16, r.prefix[:], r.mask[:]) {
			return true
		}
	}
	return false
}

// matchesSpecialPrefix reports whether every masked byte of ip equals the
// prefix entry.
func matchesSpecialPrefix(ip, prefix, mask []byte) bool {
	if len(ip) != len(prefix) {
		return false
	}
	for i := range prefix {
		if ip[i]&mask[i] != prefix[i] {
			return false
		}
	}
	return true
}

// nat64Range is the RFC 6052 NAT64 well-known prefix 64:ff9b::/96 (first 12
// bytes fixed). It sat in ianaIPv6Special until F1 (card t1968 iter15): the
// prefix embeds a per-address IPv4 destination in its low 32 bits, so a
// wholesale refusal blocked PUBLIC embeds a DNS64/NAT64 IPv6-only client
// legitimately dials. Same prefix/mask shape as the tables above.
var nat64Range = struct {
	prefix [16]byte
	mask   [16]byte
}{
	prefix: [16]byte{0x00, 0x64, 0xff, 0x9b}, mask: [16]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
}

// nat64Embedded reports whether ip sits inside 64:ff9b::/96 and, if so,
// returns the IPv4 address embedded in its low 32 bits for the caller to
// run through the SAME public-address predicate — no second policy. IPv4
// and IPv4-mapped forms carry no NAT64 prefix and return nil, so the
// predicate's recursion terminates: the extracted IPv4's 16-byte form is
// ::ffff:-mapped, which never matches the 64:ff9b:: prefix.
func nat64Embedded(ip net.IP) net.IP {
	if ip.To4() != nil {
		return nil
	}
	ip16 := ip.To16()
	if ip16 == nil || !matchesSpecialPrefix(ip16, nat64Range.prefix[:], nat64Range.mask[:]) {
		return nil
	}
	return net.IPv4(ip16[12], ip16[13], ip16[14], ip16[15])
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
