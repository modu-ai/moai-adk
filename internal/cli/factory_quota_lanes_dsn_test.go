// factory_quota_lanes_dsn_test.go — card t2118: the QAS registry fingerprint
// helper must hand the SQLite driver a drive path as a path, never as the URI
// authority (`file://C:/…` fails with "invalid uri authority: C:" on Windows).
package cli

import (
	"net/url"
	"strings"
	"testing"
)

func TestQASReadOnlyDSN_DrivePathIsNotTheAuthority(t *testing.T) {
	for _, in := range []string{
		`C:\Users\RUNNER~1\AppData\Local\Temp\x\factory.db`,
		`C:/Users/RUNNER~1/AppData/Local/Temp/x/factory.db`,
	} {
		dsn := qasReadOnlyDSN(in)
		if !strings.HasPrefix(dsn, "file:///C:/") {
			t.Errorf("qasReadOnlyDSN(%q) = %q, want the drive path after file:///", in, dsn)
		}
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatalf("parse %q: %v", dsn, err)
		}
		if u.Host != "" {
			t.Errorf("qasReadOnlyDSN(%q) = %q has URI authority %q, want none", in, dsn, u.Host)
		}
		q := u.Query()
		if q.Get("mode") != "ro" || q.Get("_pragma") != "query_only(ON)" {
			t.Errorf("qasReadOnlyDSN(%q) = %q lost its mode=ro / query_only pragma", in, dsn)
		}
	}
}
