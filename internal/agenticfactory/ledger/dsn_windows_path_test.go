package ledger

import (
	"net/url"
	"testing"
)

// A Windows drive path must reach the driver as a path, never as the URI
// authority ("invalid uri authority: C:", card t2118).
func TestLedgerDSN_WindowsDrivePathIsNotTheAuthority(t *testing.T) {
	for _, p := range []string{`C:\Users\dev\x.db`, "C:/x/y.db"} {
		u, err := url.Parse(ledgerDSN(p))
		if err != nil {
			t.Fatalf("ledgerDSN(%q) does not parse: %v", p, err)
		}
		if u.Host != "" {
			t.Fatalf("ledgerDSN(%q) = %q: host %q, want no authority", p, ledgerDSN(p), u.Host)
		}
		if u.Scheme != "file" || len(u.Path) < 3 || u.Path[:3] != "/C:" {
			t.Fatalf("ledgerDSN(%q) = %q: path %q, want it rooted at /C:", p, ledgerDSN(p), u.Path)
		}
	}
}
