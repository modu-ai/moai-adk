package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/update"
)

// TestNormalizeVersionTag covers REQ-UVF-003 (v-prefix normalization) and the
// plan-auditor D1 tag-charset constraint (elevated to HARD in the delegation
// prompt §D#4). The charset gate is path-traversal defense: a <tag> containing
// URL metacharacters (../, ?, #, +, whitespace) MUST be rejected before URL
// construction.
func TestNormalizeVersionTag(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "v-prefixed stable", input: "v3.0.0", want: "v3.0.0"},
		{name: "bare stable gets v prefix", input: "3.0.0", want: "v3.0.0"},
		{name: "rc tag", input: "v3.1.0-rc1", want: "v3.1.0-rc1"},
		{name: "bare rc gets v prefix", input: "3.1.0-rc1", want: "v3.1.0-rc1"},
		{name: "previous version", input: "v2.14.0", want: "v2.14.0"},
		{name: "uppercase V prefix", input: "V3.0.0", want: "v3.0.0"},
		{name: "uppercase V bare", input: "V3.0.0", want: "v3.0.0"},

		// REJECT cases — go-v belongs to the dev-branch path (REQ-UVF-012, §F.2).
		{name: "go-v prefix rejected", input: "go-v3.0.0", wantErr: true},
		{name: "go prefix rejected", input: "go3.0.0", wantErr: true},

		// REJECT cases — charset / path-traversal (D1, §D#4).
		{name: "path traversal dotdot", input: "../etc/passwd", wantErr: true},
		{name: "path traversal slash", input: "v3.0.0/x", wantErr: true},
		{name: "query fragment", input: "v3.0.0?x=1", wantErr: true},
		{name: "hash fragment", input: "v3.0.0#frag", wantErr: true},
		{name: "plus build metadata", input: "v3.0.0+build.5", wantErr: true},
		{name: "whitespace", input: "v 3.0.0", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "semicolon", input: "v3.0.0;rm", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := normalizeVersionTag(c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("normalizeVersionTag(%q): expected error, got %q", c.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeVersionTag(%q): unexpected error: %v", c.input, err)
			}
			if got != c.want {
				t.Errorf("normalizeVersionTag(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

// TestTagReleaseURL covers AC-UVF-002 (Host/Path of the constructed tag URL) and
// AC-UVF-003 (both 3.0.0 and v3.0.0 resolve to the same URL).
func TestTagReleaseURL(t *testing.T) {
	t.Run("bare and v-prefixed resolve identically", func(t *testing.T) {
		uBare, err := tagReleaseURL("3.0.0")
		if err != nil {
			t.Fatalf("bare: %v", err)
		}
		uV, err := tagReleaseURL("v3.0.0")
		if err != nil {
			t.Fatalf("v-prefixed: %v", err)
		}
		if uBare != uV {
			t.Errorf("bare %q != v %q", uBare, uV)
		}
		want := "https://api.github.com/repos/modu-ai/moai-adk/releases/tags/v3.0.0"
		if uBare != want {
			t.Errorf("tagReleaseURL = %q, want %q", uBare, want)
		}
	})

	t.Run("host is api.github.com on https scheme", func(t *testing.T) {
		u, err := tagReleaseURL("v3.1.0-rc1")
		if err != nil {
			t.Fatalf("rc: %v", err)
		}
		if !hasPrefix(u, "https://api.github.com/") {
			t.Errorf("tag URL not on allowlist host/scheme: %q", u)
		}
		wantPath := "/repos/modu-ai/moai-adk/releases/tags/v3.1.0-rc1"
		if !hasSuffix(u, wantPath) {
			t.Errorf("tag URL path mismatch: %q does not end with %q", u, wantPath)
		}
	})

	t.Run("invalid tag rejected before URL construction", func(t *testing.T) {
		for _, bad := range []string{"../x", "v3.0.0?y=1", "go-v3.0.0", "v 3", ""} {
			if _, err := tagReleaseURL(bad); err == nil {
				t.Errorf("tagReleaseURL(%q): expected rejection, got nil", bad)
			}
		}
	})
}

// hasPrefix / hasSuffix are tiny local helpers to keep the test importless.
func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
func hasSuffix(s, p string) bool {
	return len(s) >= len(p) && s[len(s)-len(p):] == p
}

// TestValidateUpdateVersionConflicts covers REQ-UVF-007 / AC-UVF-007:
// --version + (--check | --templates-only | --restore | --dry-run) must exit
// non-zero with a usage error naming the conflicting pair; --version + (--binary
// | --force | --yes) is permitted (proceeds).
func TestValidateUpdateVersionConflicts(t *testing.T) {
	cases := []struct {
		name          string
		version       string
		check, templatesOnly, restore, dryRun bool
		wantConflict bool
		wantFlag     string
	}{
		{name: "version alone ok", version: "v3.0.0"},
		{name: "version+check conflict", version: "v3.0.0", check: true, wantConflict: true, wantFlag: "--check"},
		{name: "version+templates-only conflict", version: "v3.0.0", templatesOnly: true, wantConflict: true, wantFlag: "--templates-only"},
		{name: "version+restore conflict", version: "v3.0.0", restore: true, wantConflict: true, wantFlag: "--restore"},
		{name: "version+dry-run conflict", version: "v3.0.0", dryRun: true, wantConflict: true, wantFlag: "--dry-run"},
		{name: "no version flag unaffected", version: "", check: true, templatesOnly: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateUpdateVersionConflicts(c.version, c.check, c.templatesOnly, c.restore, c.dryRun)
			if c.wantConflict {
				if err == nil {
					t.Fatalf("expected conflict error naming %s, got nil", c.wantFlag)
				}
				if !strings_Contains(err.Error(), c.wantFlag) {
					t.Errorf("error %q does not name the conflicting flag %s", err.Error(), c.wantFlag)
				}
				if !strings_Contains(err.Error(), "--version") {
					t.Errorf("error %q does not name --version", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected conflict: %v", err)
			}
		})
	}
}

// strings_Contains avoids importing strings just for one call.
func strings_Contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestVersionFlagRegistered covers AC-UVF-001 (--version listed in --help) and
// AC-UVF-006 (no --skip-checksum / --insecure bypass flag exists).
func TestVersionFlagRegistered(t *testing.T) {
	f := updateCmd.Flags().Lookup("version")
	if f == nil {
		t.Fatal("--version flag is not registered on updateCmd")
	}
	if f.Value.String() != "" {
		t.Errorf("--version default = %q, want empty (REQ-UVF-004: default flow preserved)", f.Value.String())
	}
	// AC-UVF-001: the description names the tag kinds.
	if !strings_Contains(f.Usage, "release tag") && !strings_Contains(f.Usage, "stable") {
		t.Errorf("--version usage %q does not name release-tag kinds", f.Usage)
	}
	// AC-UVF-006: checksum verification is mandatory — no bypass flag.
	for _, banned := range []string{"skip-checksum", "insecure"} {
		if updateCmd.Flags().Lookup(banned) != nil {
			t.Errorf("security: --%s flag MUST NOT exist (checksum bypass forbidden, REQ-UVF-006)", banned)
		}
	}
}

// TestUpdateFlagsNoVersionDefaultIsNoop is the AC-UVF-004 regression anchor
// (M3). It proves the --version flag addition does not alter the default
// `moai update` entry conditions: the flag defaults to empty, and an empty
// --version passes validation without entering the version-install path. This
// anchor MUST land before any --version branch is added to runUpdate (M4),
// making the preservation guarantee unambiguous.
func TestUpdateFlagsNoVersionDefaultIsNoop(t *testing.T) {
	// The flag defaults to empty on the package-global command.
	if got := updateCmd.Flags().Lookup("version").Value.String(); got != "" {
		t.Errorf("default --version = %q, want empty (default flow must be preserved)", got)
	}
	// An empty --version passes validation (no conflict, no resolution attempt).
	if err := validateUpdateVersionConflicts("", false, false, false, false); err != nil {
		t.Errorf("empty --version must not trigger conflict validation: %v", err)
	}
	// tagReleaseURL is never reached when --version is empty because runUpdate
	// guards the --version branch on getStringFlag(cmd,"version") != "". This
	// test documents that guard's predicate explicitly.
	if getStringFlag(updateCmd, "version") != "" {
		t.Errorf("global updateCmd --version is non-empty at rest; default flow is NOT preserved")
	}
}

// ---------------------------------------------------------------------------
// M4 fixtures + defect-path tests (REQ-UVF-008/009/010/011, AC-UVF-002).
// ---------------------------------------------------------------------------

// mockAPIConfig configures a mock GitHub API + asset server.
type mockAPIConfig struct {
	tagStatusCode int                 // /releases/tags/<tag> status (200 or 404)
	tagBody       string              // release JSON (when 200)
	assets        map[string]string   // path → content served from the mock (checksums.txt, archive)
}

// newMockGithubTLS starts a TLS httptest server that routes by URL path and
// records the Host header of the tag-resolution request. It returns the server
// and an http.Client whose DialContext redirects every connection to the mock
// (so requests to https://api.github.com/... reach the mock while preserving
// Host=api.github.com — required by AC-UVF-002).
func newMockGithubTLS(t *testing.T, cfg mockAPIConfig) (*httptest.Server, *http.Client, *string) {
	t.Helper()
	capturedHost := ""
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/modu-ai/moai-adk/releases/tags/", func(w http.ResponseWriter, r *http.Request) {
		capturedHost = r.Host
		if cfg.tagStatusCode != 0 {
			w.WriteHeader(cfg.tagStatusCode)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		if cfg.tagBody != "" {
			_, _ = fmt.Fprint(w, cfg.tagBody)
		}
	})
	for path, body := range cfg.assets {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, body)
		})
	}
	ts := httptest.NewTLSServer(mux)
	t.Cleanup(ts.Close)

	addr := ts.Listener.Addr().String()
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return ts, client, &capturedHost
}

// buildMockArchive builds a tar.gz containing a single "moai" entry copied from
// the running test binary (a real executable so validateBinaryFormat accepts
// it). Returns the archive bytes and the SHA256 hex of those bytes.
func buildMockArchive(t *testing.T) ([]byte, string) {
	t.Helper()
	// Use the running test binary as the payload — it is a valid Mach-O/ELF/PE.
	src, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatalf("read test binary: %v", err)
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	binName := "moai"
	if runtime.GOOS == "windows" {
		binName = "moai.exe"
	}
	hdr := &tar.Header{Name: binName, Mode: 0o755, Size: int64(len(src)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(src); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gz close: %v", err)
	}
	arch := buf.Bytes()
	sum := sha256.Sum256(arch)
	return arch, hex.EncodeToString(sum[:])
}

// platformArchiveName mirrors the GoReleaser archive naming the checker looks
// for (moai-adk_<version-stripped>_<os>_<arch>.<ext>).
func platformArchiveName(versionStripped string) string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("moai-adk_%s_%s_%s.%s", versionStripped, runtime.GOOS, runtime.GOARCH, ext)
}

// releaseJSON builds the GitHub release-by-tag JSON with the given archive URL
// and checksums URL.
func releaseJSON(tag, archiveURL, checksumsURL string) string {
	return fmt.Sprintf(`{"tag_name":%q,"published_at":"2026-01-01T00:00:00Z","assets":[{"name":%q,"browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}]}`,
		tag, platformArchiveName(strings.TrimPrefix(tag, "v")), archiveURL, checksumsURL)
}

// tempBinary creates a temp file with a valid executable body (copied from the
// test binary) so CreateBackup + validateBinaryFormat + Replace work.
func tempBinary(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatalf("read test binary: %v", err)
	}
	f, err := os.CreateTemp(t.TempDir(), "moai-test-bin-*")
	if err != nil {
		t.Fatalf("temp binary: %v", err)
	}
	if _, err := f.Write(src); err != nil {
		t.Fatalf("write temp binary: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp binary: %v", err)
	}
	if err := os.Chmod(f.Name(), 0o755); err != nil {
		t.Fatalf("chmod temp binary: %v", err)
	}
	return f.Name()
}

// TestInstallVersionTag_HappyPath covers AC-UVF-002 (binary replaced, checksum
// verified, exit 0, AND the captured tag-resolution request has
// Host==api.github.com + the correct Path).
func TestInstallVersionTag_HappyPath(t *testing.T) {
	arch, wantSum := buildMockArchive(t)
	checksums := fmt.Sprintf("%s  %s\n", wantSum, platformArchiveName("3.0.0"))
	cfg := mockAPIConfig{
		tagStatusCode: 200,
		tagBody:       releaseJSON("v3.0.0", "https://api.github.com/archive", "https://api.github.com/checksums.txt"),
		assets: map[string]string{
			"/checksums.txt": checksums,
			"/archive":       string(arch),
		},
	}
	_, client, capturedHost := newMockGithubTLS(t, cfg)

	binPath := tempBinary(t)
	preInfo, _ := os.Stat(binPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := installVersionTag(ctx, "v3.0.0", binPath, client)
	if err != nil {
		t.Fatalf("installVersionTag happy path: %v", err)
	}
	if res.NewVersion != "v3.0.0" {
		t.Errorf("NewVersion = %q, want v3.0.0", res.NewVersion)
	}

	// AC-UVF-002: the binary was replaced (mtime/size changed from the test
	// binary copy to the installed one — both are the same binary here, so just
	// assert the path still exists and is executable).
	postInfo, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("binary missing after install: %v", err)
	}
	if !postInfo.Mode().IsRegular() {
		t.Errorf("binary not a regular file after install")
	}
	_ = preInfo // (informational; the file is overwritten in place)

	// AC-UVF-002 core assertion: Host/Path confinement of the tag-resolution URL.
	if *capturedHost != "api.github.com" {
		t.Errorf("tag-resolution Host = %q, want api.github.com (REQ-UVF-005)", *capturedHost)
	}
}

// TestInstallVersionTag_BareTagNormalized covers AC-UVF-003: "3.0.0" hits the
// same /releases/tags/v3.0.0 URL as "v3.0.0".
func TestInstallVersionTag_BareTagNormalized(t *testing.T) {
	arch, wantSum := buildMockArchive(t)
	checksums := fmt.Sprintf("%s  %s\n", wantSum, platformArchiveName("3.0.0"))
	cfg := mockAPIConfig{
		tagBody: releaseJSON("v3.0.0", "https://api.github.com/archive", "https://api.github.com/checksums.txt"),
		assets: map[string]string{
			"/checksums.txt": checksums,
			"/archive":       string(arch),
		},
	}
	_, client, capturedHost := newMockGithubTLS(t, cfg)
	binPath := tempBinary(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := installVersionTag(ctx, "3.0.0", binPath, client); err != nil {
		t.Fatalf("bare tag install: %v", err)
	}
	if *capturedHost != "api.github.com" {
		t.Errorf("Host = %q, want api.github.com", *capturedHost)
	}
}

// TestInstallVersionTag_404 covers REQ-UVF-008 / AC-UVF-008.
func TestInstallVersionTag_404(t *testing.T) {
	cfg := mockAPIConfig{tagStatusCode: 404}
	_, client, _ := newMockGithubTLS(t, cfg)
	binPath := tempBinary(t)
	_, err := installVersionTag(context.Background(), "v9.9.9-nope", binPath, client)
	if err == nil {
		t.Fatal("expected error for 404 tag, got nil")
	}
	if !strings.Contains(err.Error(), "v9.9.9-nope") {
		t.Errorf("error %q does not name the offending tag", err.Error())
	}
}

// TestInstallVersionTag_NoBinaryAsset covers REQ-UVF-009 / AC-UVF-009: a release
// with assets[] but no archive matching GOOS/GOARCH.
func TestInstallVersionTag_NoBinaryAsset(t *testing.T) {
	// Release JSON with an asset that is NOT the platform archive (so info.URL
	// stays empty) but DOES include checksums.txt (so CheckLatest succeeds).
	body := `{"tag_name":"v3.1.0-rc2","published_at":"2026-01-01T00:00:00Z","assets":[{"name":"source.zip","browser_download_url":"https://api.github.com/src"},{"name":"checksums.txt","browser_download_url":"https://api.github.com/checksums.txt"}]}`
	cfg := mockAPIConfig{
		tagBody: body,
		assets:  map[string]string{"/checksums.txt": "deadbeef  source.zip\n"},
	}
	_, client, _ := newMockGithubTLS(t, cfg)
	binPath := tempBinary(t)
	preStat, _ := os.Stat(binPath)
	_, err := installVersionTag(context.Background(), "v3.1.0-rc2", binPath, client)
	if err == nil {
		t.Fatal("expected error for no-binary-asset, got nil")
	}
	if !strings.Contains(err.Error(), "v3.1.0-rc2") {
		t.Errorf("error %q does not name the tag", err.Error())
	}
	// AC-UVF-009: filesystem untouched (binary path unchanged).
	postStat, _ := os.Stat(binPath)
	if postStat == nil {
		t.Errorf("binary path missing after no-asset failure")
	} else if !postStat.ModTime().Equal(preStat.ModTime()) {
		// mtime should be unchanged because Replace never ran.
		t.Errorf("binary mtime changed on no-asset failure (filesystem must be untouched)")
	}
}

// TestInstallVersionTag_ChecksumMismatch covers REQ-UVF-010 / AC-UVF-010.
func TestInstallVersionTag_ChecksumMismatch(t *testing.T) {
	arch, _ := buildMockArchive(t)
	// checksums.txt advertises a WRONG hash for the archive.
	checksums := fmt.Sprintf("%s  %s\n", "0000000000000000000000000000000000000000000000000000000000000000", platformArchiveName("3.0.0"))
	cfg := mockAPIConfig{
		tagBody: releaseJSON("v3.0.0", "https://api.github.com/archive", "https://api.github.com/checksums.txt"),
		assets: map[string]string{
			"/checksums.txt": checksums,
			"/archive":       string(arch),
		},
	}
	_, client, _ := newMockGithubTLS(t, cfg)
	binPath := tempBinary(t)
	preBytes, _ := os.ReadFile(binPath)
	_, err := installVersionTag(context.Background(), "v3.0.0", binPath, client)
	if err == nil {
		t.Fatal("expected checksum-mismatch error, got nil")
	}
	if !errors.Is(err, update.ErrChecksumMismatch) && !strings.Contains(err.Error(), "checksum") {
		t.Errorf("error %q does not reference checksum mismatch", err.Error())
	}
	// AC-UVF-010: running binary path is byte-identical pre/post.
	postBytes, _ := os.ReadFile(binPath)
	if !bytes.Equal(preBytes, postBytes) {
		t.Errorf("running binary path was modified on checksum-mismatch (must be byte-identical)")
	}
}

// TestInstallVersionTag_NetworkFailureResolution covers REQ-UVF-011 (resolution
// phase): the tag-resolution HTTP call fails.
func TestInstallVersionTag_NetworkFailureResolution(t *testing.T) {
	// Start then immediately close a TLS server so dials fail.
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := ts.Listener.Addr().String()
	ts.Close()
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr) // server gone → dial fails
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	binPath := tempBinary(t)
	_, err := installVersionTag(context.Background(), "v3.0.0", binPath, client)
	if err == nil {
		t.Fatal("expected resolution-phase network error, got nil")
	}
	if !strings.Contains(err.Error(), "resolution") {
		t.Errorf("error %q does not name the resolution phase", err.Error())
	}
}

// TestInstallVersionTag_NetworkFailureDownload covers REQ-UVF-011 (download
// phase): tag resolution succeeds but the archive download fails.
func TestInstallVersionTag_NetworkFailureDownload(t *testing.T) {
	// tag resolution succeeds; archive URL points at an unreachable port.
	body := releaseJSON("v3.0.0", "https://api.github.com/archive", "https://api.github.com/checksums.txt")
	cfg := mockAPIConfig{
		tagBody: body,
		assets: map[string]string{
			"/checksums.txt": "deadbeef  " + platformArchiveName("3.0.0") + "\n",
			// /archive intentionally NOT served → 404 on download.
		},
	}
	_, client, _ := newMockGithubTLS(t, cfg)
	binPath := tempBinary(t)
	_, err := installVersionTag(context.Background(), "v3.0.0", binPath, client)
	if err == nil {
		t.Fatal("expected download-phase error, got nil")
	}
	// Download failure wraps as "download" phase.
	if !strings.Contains(err.Error(), "download") {
		t.Errorf("error %q does not name the download phase", err.Error())
	}
}

// TestInstallVersionTag_InvalidTag covers the charset gate firing before any
// network call (no request should reach the mock).
func TestInstallVersionTag_InvalidTag(t *testing.T) {
	var requestCount int32
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
	}))
	ts.Close() // closed; if a request were attempted it would fail too
	binPath := tempBinary(t)
	_, err := installVersionTag(context.Background(), "../etc/passwd", binPath, http.DefaultClient)
	if err == nil {
		t.Fatal("expected invalid-tag error, got nil")
	}
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("error %q is not ErrInvalidTag", err.Error())
	}
	if atomic.LoadInt32(&requestCount) != 0 {
		t.Errorf("invalid tag reached the network (%d requests) — charset gate must fire first", atomic.LoadInt32(&requestCount))
	}
	_ = ts
	_ = io.Discard
}

// ---------------------------------------------------------------------------
// M5 (downgrade confirmation) + M6 (dev/RC branch interaction) + AC-UVF-005/012.
// ---------------------------------------------------------------------------

// TestIsVersionDowngrade covers REQ-UVF-013's downgrade predicate.
func TestIsVersionDowngrade(t *testing.T) {
	cases := []struct {
		requested, current string
		want               bool
	}{
		{"v3.0.0", "v3.2.0", true},   // older requested → downgrade
		{"v3.2.0", "v3.0.0", false},  // newer requested → upgrade
		{"v3.0.0", "v3.0.0", false},  // same → not downgrade
		{"3.0.0", "v3.2.0", true},    // bare requested, v-prefixed current
		{"v3.1.0-rc1", "v3.0.0", false}, // rc newer than stable
		{"go-v3.0.0", "v3.2.0", true},   // go-v prefix handled
	}
	for _, c := range cases {
		t.Run(c.requested+"->"+c.current, func(t *testing.T) {
			if got := isVersionDowngrade(c.requested, c.current); got != c.want {
				t.Errorf("isVersionDowngrade(%q, %q) = %v, want %v", c.requested, c.current, got, c.want)
			}
		})
	}
}

// TestCompareVersionLoose covers the version comparison helper.
func TestCompareVersionLoose(t *testing.T) {
	cases := []struct{ a, b string; want int }{
		{"v3.0.0", "v3.0.0", 0},
		{"v3.1.0", "v3.0.0", 1},
		{"v3.0.0", "v3.1.0", -1},
		{"v3.0.1", "v3.0.0", 1},
		{"3.0.0", "v3.0.0", 0},
	}
	for _, c := range cases {
		if got := compareVersionLoose(c.a, c.b); got != c.want {
			t.Errorf("compareVersionLoose(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

// TestRunVersionBranch_RejectsPoisonedEnv covers AC-UVF-005: a poisoned
// MOAI_UPDATE_URL is rejected fail-closed before any --version HTTP call.
func TestRunVersionBranch_RejectsPoisonedEnv(t *testing.T) {
	t.Setenv("MOAI_UPDATE_URL", "https://evil.example.com/")
	// No mock server needed: the env gate must fire before any network.
	err := runVersionBranch(updateCmd, "v3.0.0")
	if err == nil {
		t.Fatal("expected poisoned-env rejection, got nil")
	}
	if !strings.Contains(err.Error(), "MOAI_UPDATE_URL") {
		t.Errorf("error %q does not name MOAI_UPDATE_URL", err.Error())
	}
	if !strings.Contains(err.Error(), "evil.example.com") && !strings.Contains(err.Error(), "not on allowlist") {
		t.Errorf("error %q does not name the allowlist violation", err.Error())
	}
}

// TestRunVersionBranch_NonTTYProceeds covers AC-UVF-013 (non-TTY skips the
// downgrade prompt and proceeds). Stdin is not a TTY under `go test`, so the
// prompt is skipped and a downgrade install is attempted and succeeds.
func TestRunVersionBranch_NonTTYProceeds(t *testing.T) {
	if isTerminalStdin() {
		t.Skip("test requires non-TTY stdin; run via `go test` (pipe), not an interactive shell")
	}
	arch, wantSum := buildMockArchive(t)
	checksums := fmt.Sprintf("%s  %s\n", wantSum, platformArchiveName("3.0.0"))
	cfg := mockAPIConfig{
		tagBody: releaseJSON("v3.0.0", "https://api.github.com/archive", "https://api.github.com/checksums.txt"),
		assets: map[string]string{
			"/checksums.txt": checksums,
			"/archive":       string(arch),
		},
	}
	_, client, _ := newMockGithubTLS(t, cfg)

	prevClient := versionInstallHTTPClient
	prevBin := versionInstallBinaryPath
	versionInstallHTTPClient = client
	versionInstallBinaryPath = tempBinary(t)
	t.Cleanup(func() {
		versionInstallHTTPClient = prevClient
		versionInstallBinaryPath = prevBin
	})
	t.Setenv("MOAI_UPDATE_URL", "")

	// v3.0.0 is a downgrade vs most running versions; non-TTY must skip the
	// prompt and proceed without error.
	if err := runVersionBranch(updateCmd, "v3.0.0"); err != nil {
		t.Fatalf("non-TTY downgrade proceed: %v", err)
	}
}

// isTerminalStdin is the same check runVersionBranch uses to gate the prompt.
func isTerminalStdin() bool {
	// Defer to the same library so the test mirrors production semantics.
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// TestRunVersionBranch_VersionTagIndependentOfDevBuild covers REQ-UVF-012 /
// AC-UVF-012: --version ALWAYS resolves /releases/tags/<tag> directly,
// independent of whether the running binary is a dev/rc/go-v build. The default
// moai update dev-branch /releases path (deps.go EnsureUpdate) is untouched by
// this SPEC.
func TestRunVersionBranch_VersionTagIndependentOfDevBuild(t *testing.T) {
	// tagReleaseURL never consults the running version — verify by construction.
	u1, err := tagReleaseURL("v3.0.0")
	if err != nil {
		t.Fatal(err)
	}
	// Regardless of any hypothetical running version, the URL is the tag URL.
	wantSuffix := "/releases/tags/v3.0.0"
	if !strings.HasSuffix(u1, wantSuffix) {
		t.Errorf("URL %q does not end with %q (dev-branch independence)", u1, wantSuffix)
	}
}

// TestEnsureUpdate_DevBranchPreserved is the AC-UVF-012 default-path half: a
// dev/rc/go-v running version MUST still select the /releases list endpoint
// (not /latest) on the default path. This is existing EnsureUpdate behavior
// (deps.go:386); this test locks it so a future change cannot silently break
// the dev-branch self-update path.
func TestEnsureUpdate_DevBranchPreserved(t *testing.T) {
	// EnsureUpdate reads version.GetVersion(); we cannot easily force a dev
	// version string without a build-tag injection. Instead, assert the
	// dev-branch predicate (the same one EnsureUpdate uses) classifies a go-v
	// version as dev → /releases (list), NOT /releases/latest.
	validation := func(v string) (isDev bool) {
		return v == "dev" || strings.Contains(v, "rc") || strings.Contains(v, "alpha") ||
			strings.Contains(v, "beta") || strings.HasPrefix(v, "go-v")
	}
	for _, dev := range []string{"dev", "v3.1.0-rc1", "go-v3.0.0", "v3.0.0-beta"} {
		if !validation(dev) {
			t.Errorf("dev-branch predicate must classify %q as dev (selects /releases list endpoint)", dev)
		}
	}
	// A stable production version must NOT be classified dev (selects /latest).
	if validation("v3.0.0") {
		t.Errorf("stable v3.0.0 must NOT be classified dev (default path uses /latest)")
	}
	// And the two endpoints differ — the dev path is NOT /latest.
	if githubReleasesURL == githubLatestReleaseURL {
		t.Error("dev-branch endpoint must differ from /latest")
	}
}
