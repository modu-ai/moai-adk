package homestate

import (
	"net/url"
	"testing"
)

func TestRuntimeCensusDSN_WindowsDrivePathIsNotTheAuthority(t *testing.T) {
	for _, p := range []string{`C:\Users\dev\factory.db`, "C:/x/factory.db"} {
		got := runtimeCensusDSN(p)
		u, err := url.Parse(got)
		if err != nil {
			t.Fatalf("runtimeCensusDSN(%q) does not parse: %v", p, err)
		}
		if u.Host != "" || len(u.Path) < 3 || u.Path[:3] != "/C:" {
			t.Fatalf("runtimeCensusDSN(%q) = %q: host %q path %q, want no authority and a /C: rooted path", p, got, u.Host, u.Path)
		}
		if u.Query().Get("mode") != "ro" {
			t.Fatalf("runtimeCensusDSN(%q) = %q lost mode=ro", p, got)
		}
	}
}
