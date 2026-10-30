// screens_error_refine_test.go: B13. A failed SPEC scan reached the 500 body of
// /, /factory and /specs through err.Error() unrefined (handleMonitor already
// refined). Written first as the afred RED (nine failing subtests), green with
// the three bodies passing through refineText.
//
// Contract: a value of each category (e-mail, home path, token) inside the scan
// error must not appear in the response body; the status and the prefix stay.
// Sentinel on failure: AF_SCREEN_ERROR_REFINE.
package web

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestScreenScanErrorIsRefined(t *testing.T) {
	for path, prefix := range map[string]string{
		"/":        "overview unavailable",
		"/factory": "factory unavailable",
		"/specs":   "spec list unavailable",
	} {
		for cat, payload := range screenPayloads {
			t.Run(path+"/"+cat, func(t *testing.T) {
				a := newApp(Config{ProjectRoot: t.TempDir()})
				a.specs = newSpecCache(func(string) ([]SpecRowVM, map[string][]FindingVM, error) {
					return nil, nil, errors.New("scan " + payload + " failed")
				}, specCacheMaxAge)
				rec := serveGet(t, a.routes(), path)
				if rec.Code != http.StatusInternalServerError || !strings.HasPrefix(rec.Body.String(), prefix) {
					t.Fatalf("AF_SCREEN_ERROR_REFINE: %s = %d %q, want 500 starting %q", path, rec.Code, rec.Body.String(), prefix)
				}
				if strings.Contains(rec.Body.String(), payload) {
					t.Errorf("AF_SCREEN_ERROR_REFINE: the %s reached the 500 body of %s", cat, path)
				}
			})
		}
	}
}
