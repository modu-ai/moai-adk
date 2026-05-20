// Mandatory-checksum tests for SPEC-V3R5-SECURITY-CRIT-001 P0-3 (CWE-345).
//
// Threat: prior to this SPEC, checker.buildVersionInfo silently set
// info.Checksum = "" when checksums.txt download failed, and
// updater.Download then skipped verification when Checksum == "". A network
// attacker who could intercept the checksums.txt fetch could deliver a
// tampered binary with no signal to the user.
//
// Fix: checker.queryRelease MUST return ErrChecksumUnavailable when the
// checksums.txt URL is missing OR the download fails after all retries.
// updater.Download MUST also return ErrChecksumUnavailable when given a
// VersionInfo with an empty Checksum, as a defense-in-depth guard.

package update

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestCheckLatestChecksumDownloadFailureAborts covers AC-SEC-009:
// When the release asset list includes a checksums.txt URL but the download
// returns 5xx (or 404), CheckLatest MUST return ErrChecksumUnavailable.
// Silent skip is forbidden.
//
// NOTE: This test supersedes the legacy
// TestChecker_CheckLatest_ChecksumDownloadFailed which previously asserted
// the unsafe "graceful degradation" behaviour. The legacy test was
// rewritten as part of this SPEC (M3 GREEN), and this test locks in the
// secure behaviour.
func TestCheckLatestChecksumDownloadFailureAborts(t *testing.T) {
	t.Parallel()

	// Counter to verify checksums.txt is retried (REQ-SEC-003-003) before
	// the request finally errors out.
	var attempts int32

	checksumsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer checksumsTS.Close()

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	archiveName := fmt.Sprintf("moai-adk_1.2.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: archiveName, BrowserDownloadURL: "https://example.com/moai.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: checksumsTS.URL},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)

	// AC-SEC-009 binary check: must return an error (not silent skip).
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Fatal("AC-SEC-009 regression: CheckLatest succeeded despite checksums.txt 503; expected ErrChecksumUnavailable")
	}
	if !errors.Is(err, ErrChecksumUnavailable) {
		t.Errorf("error should wrap ErrChecksumUnavailable; got %v", err)
	}

	// AC-SEC-010 sanity: retry count > 1 confirms exponential backoff loop
	// engaged. Exact attempt count is asserted in
	// TestChecksumDownloadRetryAttempts so this test only requires >1.
	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("expected ≥2 attempts (retry loop), got %d", got)
	}
}

// TestCheckLatestNoChecksumsAssetAborts covers REQ-SEC-003-002:
// When the release does not include a checksums.txt asset at all,
// CheckLatest MUST still return ErrChecksumUnavailable rather than
// silently producing a VersionInfo with empty Checksum.
func TestCheckLatestNoChecksumsAssetAborts(t *testing.T) {
	t.Parallel()

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	archiveName := fmt.Sprintf("moai-adk_1.2.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: archiveName, BrowserDownloadURL: "https://example.com/moai.tar.gz"},
			// NOTE: no checksums.txt entry
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)

	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Fatal("expected error when checksums.txt absent; got nil")
	}
	if !errors.Is(err, ErrChecksumUnavailable) {
		t.Errorf("error should wrap ErrChecksumUnavailable; got %v", err)
	}
}

// TestChecksumDownloadRetryAttempts covers AC-SEC-010:
// downloadChecksumWithRetry must attempt downloads up to maxAttempts times
// (3 by default) before returning. Attempts must use exponential backoff
// but the test uses a tiny base interval to keep total runtime under 1s.
func TestChecksumDownloadRetryAttempts(t *testing.T) {
	t.Parallel()

	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	// downloadChecksumWithRetry is the implementation hook the test
	// exercises directly so we do not pay the 2s/4s production backoff.
	c := NewChecker("http://example.com", http.DefaultClient).(*checker)
	_, err := c.downloadChecksumWithRetry(ts.URL, "moai-adk_1.2.0_test.tar.gz", 3, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("attempts = %d, want 3 (maxAttempts)", got)
	}
}

// TestChecksumDownloadRetrySuccess covers happy-path: downloadChecksumWithRetry
// succeeds on a retry without re-trying after success.
func TestChecksumDownloadRetrySuccess(t *testing.T) {
	t.Parallel()

	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`abc123  moai-adk_1.2.0_test.tar.gz` + "\n"))
	}))
	defer ts.Close()

	c := NewChecker("http://example.com", http.DefaultClient).(*checker)
	got, err := c.downloadChecksumWithRetry(ts.URL, "moai-adk_1.2.0_test.tar.gz", 3, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if got != "abc123" {
		t.Errorf("checksum = %q, want %q", got, "abc123")
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("attempts = %d, want 2 (success on 2nd)", got)
	}
}

// TestDownloadAndVerifyEmptyChecksumRejected covers AC-SEC-011:
// Updater.Download must refuse to proceed when given a VersionInfo whose
// Checksum field is empty. This is defense-in-depth: even if a future
// checker bypass slipped past the M3 fix, the updater itself blocks the
// unsafe write.
//
// Verification: ErrChecksumUnavailable is returned, AND no HTTP GET is
// performed against version.URL (the binary URL).
func TestDownloadAndVerifyEmptyChecksumRejected(t *testing.T) {
	t.Parallel()

	var binaryRequests int32
	binTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&binaryRequests, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("malicious payload"))
	}))
	defer binTS.Close()

	updater := NewUpdater("/tmp/moai-test-bin", http.DefaultClient)
	info := &VersionInfo{
		Version:  "v1.2.0",
		URL:      binTS.URL,
		Checksum: "", // ← the dangerous case
	}

	_, err := updater.Download(context.Background(), info)
	if err == nil {
		t.Fatal("AC-SEC-011 regression: Download proceeded with empty Checksum")
	}
	if !errors.Is(err, ErrChecksumUnavailable) {
		t.Errorf("error should wrap ErrChecksumUnavailable; got %v", err)
	}

	// Confirm no binary HTTP request was issued.
	if n := atomic.LoadInt32(&binaryRequests); n != 0 {
		t.Errorf("binary URL was fetched %d times; expected 0 (must abort before download)", n)
	}
}

// TestQueryReleaseChecksumGraceMessage covers REQ-SEC-003-006:
// The wrapped error message must contain a recognisable phrase so the CLI
// can surface a useful explanation to the user.
func TestQueryReleaseChecksumGraceMessage(t *testing.T) {
	t.Parallel()

	checksumsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer checksumsTS.Close()

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	archiveName := fmt.Sprintf("moai-adk_1.2.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: archiveName, BrowserDownloadURL: "https://example.com/moai.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: checksumsTS.URL},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "checksum") {
		t.Errorf("error message %q should mention 'checksum' for user diagnostic", err.Error())
	}
}
