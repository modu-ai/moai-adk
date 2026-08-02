package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

// SPEC-UPDATE-VERSION-FLAG-001 — moai update --version <tag>.
//
// This file adds a --version <tag> flag to `moai update` that installs a
// specific GitHub release tag (stable / rc / previous version) by resolving
// <tag> to a release via GET /repos/modu-ai/moai-adk/releases/tags/<tag> on the
// api.github.com allowlist, then downloading the matching binary asset through
// the EXISTING checksum-verified download path (internal/update.Updater), and
// atomically replacing the running binary.
//
// Reuse, do NOT fork (AP-UVF-001): installVersionTag reuses update.NewChecker,
// update.NewUpdater, and update.NewRollback — the same objects EnsureUpdate
// constructs for the default update path. The only new logic is tag
// validation/normalization, the tag-resolution URL shape, and the
// backup→download→replace orchestration sequence that mirrors
// orchestratorImpl.Update steps 2-4 without the IsUpdateAvailable semver gate
// (--version installs an explicit tag regardless of newer/older).

// versionInstallHTTPClient is a test seam: when non-nil, runVersionBranch
// passes it to installVersionTag so tests can redirect api.github.com to a
// local httptest mock without making real network calls. Production leaves it
// nil (installVersionTag falls back to http.DefaultClient).
var versionInstallHTTPClient *http.Client

// versionInstallBinaryPath is a test seam overriding os.Executable() so the
// branch can install into a temp binary under test. Production leaves it empty.
var versionInstallBinaryPath string

// tagCharsetRegexp restricts a --version tag to alphanumerics, dot, and hyphen,
// with an optional single leading v/V prefix. It rejects URL/path
// metacharacters (slash, question, hash, plus, whitespace) BEFORE URL
// construction so a tag value cannot inject path traversal or query fragments
// into the GitHub API URL.
//
// @MX:ANCHOR: [AUTO] tagCharsetRegexp is the path-traversal defense for --version tag values
// @MX:REASON: SPEC-UPDATE-VERSION-FLAG-001 plan §B#4 (charset validation elevated to HARD constraint in delegation §D#4). A tag carrying ../, ?, #, +, or whitespace would otherwise be spliced raw into the release-by-tag URL.
var tagCharsetRegexp = regexp.MustCompile(`^v?[0-9A-Za-z.\-]+$`)

// ErrInvalidTag is returned when a --version value fails charset or
// v-prefix-normalization validation.
var ErrInvalidTag = fmt.Errorf("invalid --version tag")

// normalizeVersionTag validates and canonicalizes a raw --version value to the
// GitHub release tag form.
//
//   - "3.0.0"    → "v3.0.0"  (bare semver gets a leading v, REQ-UVF-003)
//   - "v3.0.0"   → "v3.0.0"  (already canonical)
//   - "V3.0.0"   → "v3.0.0"  (uppercase V normalized to lowercase v)
//   - "v3.1.0-rc1" → "v3.1.0-rc1"
//   - "go-v3.0.0" → error    (dev-branch territory, REQ-UVF-012 / spec §F.2)
//   - "../x", "v3?y", "v3#f", "v3+b", "v 3" → error (charset / path-traversal)
//
// The v-prefix discipline reuses normalizeVersionMajor's stripping approach
// (internal/cli/v2_detection.go) but produces the full canonical v-prefixed
// tag rather than just the major integer.
func normalizeVersionTag(rawTag string) (string, error) {
	s := strings.TrimSpace(rawTag)
	if s == "" {
		return "", fmt.Errorf("%w: empty tag", ErrInvalidTag)
	}

	// Reject the dev-branch "go-v" / "go" prefix outright — those belong to the
	// dev-branch self-update path (REQ-UVF-012), not to an explicit tag install.
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "go-v") || strings.HasPrefix(lower, "go") && len(s) > 2 && s[2] >= '0' && s[2] <= '9' {
		return "", fmt.Errorf("%w: %q (dev-branch go-v tags are not installable via --version; use default `moai update`)", ErrInvalidTag, rawTag)
	}

	// Charset gate (path-traversal defense, §D#4). Applied to the original
	// value so a leading "v" is tolerated but metacharacters anywhere are not.
	if !tagCharsetRegexp.MatchString(s) {
		return "", fmt.Errorf("%w: %q (allowed charset: letters, digits, '.', '-'; rejected metacharacters: / ? # + whitespace)", ErrInvalidTag, rawTag)
	}

	// Normalize an uppercase V prefix to lowercase v; add a v prefix when absent.
	switch {
	case strings.HasPrefix(s, "v"):
		return s, nil
	case strings.HasPrefix(s, "V"):
		return "v" + s[1:], nil
	default:
		return "v" + s, nil
	}
}

// tagReleaseURL validates and normalizes a raw --version value and constructs
// the GitHub API "release by tag" URL on the api.github.com allowlist. The URL
// stays on the compiled githubReleasesURL constant (REQ-UVF-005) — no
// environment input reaches it.
func tagReleaseURL(rawTag string) (string, error) {
	canonical, err := normalizeVersionTag(rawTag)
	if err != nil {
		return "", err
	}
	return githubReleasesURL + "/tags/" + canonical, nil
}

// versionInstallResult summarizes a --version install outcome.
type versionInstallResult struct {
	NewVersion   string
	RollbackPath string
}

// installVersionTag resolves rawTag to a GitHub release, downloads+verifies the
// binary through the existing checksum-verified download path, and atomically
// replaces the binary at binaryPath. It reuses update.NewChecker /
// update.NewUpdater / update.NewRollback (AP-UVF-001 — no parallel downloader).
//
// httpClient is injected so tests can redirect the api.github.com host to a
// local httptest mock while preserving the request Host header (AC-UVF-002).
// A nil client uses http.DefaultClient.
//
// Phase-named errors (AC-UVF-011): resolution-phase failures (tag 404, no
// checksums.txt, network) are wrapped with "resolution"; download-phase
// failures (network, checksum mismatch) with "download".
func installVersionTag(ctx context.Context, rawTag, binaryPath string, httpClient *http.Client) (*versionInstallResult, error) {
	tagURL, err := tagReleaseURL(rawTag)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	// Phase 1: tag resolution. CheckLatest surfaces HTTP 404 (REQ-UVF-008),
	// missing checksums.txt (ErrChecksumUnavailable), and network errors
	// (REQ-UVF-011 resolution phase).
	checker := update.NewChecker(tagURL, httpClient)
	info, err := checker.CheckLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("--version resolution (tag=%s): %w", rawTag, err)
	}

	// REQ-UVF-009: a release with no matching binary asset for the running
	// platform must fail clearly, naming the tag, platform, and present assets.
	if strings.TrimSpace(info.URL) == "" {
		return nil, fmt.Errorf("--version %s: no installable binary asset for %s/%s (available assets did not match)", rawTag, runtime.GOOS, runtime.GOARCH)
	}

	// Phase 2: backup + download+verify + atomic replace. Reuses the same
	// update.Updater / update.Rollback types EnsureUpdate constructs.
	rollback := update.NewRollback(binaryPath)
	backupPath, err := rollback.CreateBackup()
	if err != nil {
		return nil, fmt.Errorf("--version backup: %w", err)
	}

	updater := update.NewUpdater(binaryPath, httpClient)
	downloadPath, err := updater.Download(ctx, info)
	if err != nil {
		// REQ-UVF-010: checksum mismatch discards the downloaded bytes (the
		// updater cleans up its temp file on error) and never reaches Replace,
		// so the running binary path is byte-identical pre/post. Wrap with the
		// download phase for AC-UVF-011.
		return nil, rollbackOrFail(backupPath, fmt.Errorf("--version download (tag=%s): %w", rawTag, err))
	}

	if err := updater.Replace(ctx, downloadPath); err != nil {
		return nil, rollbackOrFail(backupPath, fmt.Errorf("--version replace (tag=%s): %w", rawTag, err))
	}

	return &versionInstallResult{NewVersion: info.Version, RollbackPath: backupPath}, nil
}

// rollbackOrFail attempts a rollback restore on failure and returns the
// (possibly wrapped) original error. Mirrors orchestratorImpl.attemptRollback.
func rollbackOrFail(backupPath string, originalErr error) error {
	if restoreErr := update.NewRollback("").Restore(backupPath); restoreErr != nil {
		// Best-effort: keep the original error primary; note the rollback path.
		return fmt.Errorf("%w (rollback also failed: %v; manual recovery from %s)", originalErr, restoreErr, backupPath)
	}
	return originalErr
}

// runtimeGOOS / runtimeGOARCH are thin wrappers so tests can assert on the
// reported platform string without importing runtime at every call site.

// validateUpdateVersionConflicts enforces the --version mutual-exclusion matrix
// (REQ-UVF-007 / AC-UVF-007). --version is mutually exclusive with --check,
// --templates-only, --restore, and --dry-run. Combinations with --binary,
// --force, and --yes are permitted.
func validateUpdateVersionConflicts(versionTag string, check, templatesOnly, restore, dryRun bool) error {
	if versionTag == "" {
		return nil
	}
	type conflict struct {
		flag string
		on   bool
	}
	// Order is deterministic for stable error messages.
	conflicts := []conflict{
		{"--check", check},
		{"--templates-only", templatesOnly},
		{"--restore", restore},
		{"--dry-run", dryRun},
	}
	for _, c := range conflicts {
		if c.on {
			return fmt.Errorf("--version and %s are mutually exclusive", c.flag)
		}
	}
	return nil
}

// _ pins the update import; installVersionTag consumes update.NewChecker /
// update.NewUpdater / update.NewRollback from this file.
var _ = update.ErrChecksumMismatch

// compareVersionLoose compares two version strings (with optional go-v / v
// prefixes) by [major, minor, patch]. Returns -1, 0, or 1. Pre-release suffixes
// are stripped before numeric comparison. It mirrors update.compareSemver but
// is local because the update-package helper is unexported.
func compareVersionLoose(a, b string) int {
	a = stripVersionPrefix(a)
	b = stripVersionPrefix(b)
	ap := parseSemverLoose(a)
	bp := parseSemverLoose(b)
	for i := 0; i < 3; i++ {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	return 0
}

func stripVersionPrefix(s string) string {
	for _, p := range []string{"go-v", "v"} {
		if strings.HasPrefix(s, p) {
			return strings.TrimPrefix(s, p)
		}
	}
	return s
}

func parseSemverLoose(v string) [3]int {
	var parts [3]int
	for i, seg := range strings.SplitN(v, ".", 3) {
		if i >= 3 {
			break
		}
		if idx := strings.IndexAny(seg, "-+"); idx >= 0 {
			seg = seg[:idx]
		}
		n := 0
		for _, r := range seg {
			if r < '0' || r > '9' {
				break
			}
			n = n*10 + int(r-'0')
		}
		parts[i] = n
	}
	return parts
}

// isVersionDowngrade reports whether requestedTag is older than currentVersion
// (REQ-UVF-013). Used to gate the interactive downgrade confirmation.
func isVersionDowngrade(requestedTag, currentVersion string) bool {
	return compareVersionLoose(requestedTag, currentVersion) < 0
}

// runVersionBranch is the moai update --version <tag> code path. It:
//  1. Rejects a poisoned MOAI_UPDATE_URL fail-closed (AC-UVF-005) before any
//     --version HTTP call, reusing the existing validateUpdateURL allowlist gate.
//  2. Confirms a downgrade interactively (REQ-UVF-013) unless --yes or non-TTY.
//  3. Installs the resolved tag via the checksum-verified download path
//     (installVersionTag → update.Updater.Download/Replace).
//  4. Re-execs into the newly installed binary unless --binary (REQ-UVF-014).
//
// It runs BEFORE the default binary-update + template-sync flow so a non-project
// cwd does not block a binary-only --version install (parity with --binary).
func runVersionBranch(cmd *cobra.Command, versionTag string) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	// AC-UVF-005: reject a poisoned MOAI_UPDATE_URL fail-closed before any
	// --version HTTP call. validateUpdateURL is the existing allowlist gate;
	// --version constructs its own tag URL from compiled constants, but the env
	// var must still be validated so a poisoned environment cannot survive into
	// the install.
	if envURL := os.Getenv(config.EnvUpdateURL); envURL != "" {
		if err := validateUpdateURL(envURL); err != nil {
			return fmt.Errorf("%s rejected: %w", config.EnvUpdateURL, err)
		}
	}

	binaryPath, err := func() (string, error) {
		if versionInstallBinaryPath != "" {
			return versionInstallBinaryPath, nil
		}
		return os.Executable()
	}()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	// REQ-UVF-013: downgrade confirmation. --yes OR non-TTY stdin skips the
	// prompt and proceeds.
	currentVersion := version.GetVersion()
	assumeYes := getBoolFlag(cmd, "yes")
	if !assumeYes && isatty.IsTerminal(os.Stdin.Fd()) && isVersionDowngrade(versionTag, currentVersion) {
		var confirm bool
		prompt := huh.NewConfirm().
			Title(fmt.Sprintf("Downgrade %s → %s?", currentVersion, versionTag)).
			Description("The requested tag is older than the running version.").
			Value(&confirm)
		form := huh.NewForm(huh.NewGroup(prompt)).WithTheme(moaiHuhTheme())
		if err := form.Run(); err != nil {
			return fmt.Errorf("downgrade confirmation: %w", err)
		}
		if !confirm {
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Downgrade aborted", Theme: &th}))
			return nil
		}
	}

	_, _ = fmt.Fprintln(out, tui.KV("Installing version", versionTag, tui.KVOpts{Theme: &th, KeyWidth: 16}))
	_, _ = fmt.Fprintln(out, tui.CheckLine("run", "Resolving tag + downloading binary", "", "", &th))

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	res, err := installVersionTag(ctx, versionTag, binaryPath, versionInstallHTTPClient)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Installed " + res.NewVersion, Theme: &th}))

	// REQ-UVF-014: re-exec into the newly installed binary so subsequent template
	// sync (when not suppressed by --binary) runs against the new binary's
	// embedded templates. --binary skips re-exec (binary-only install).
	if !getBoolFlag(cmd, "binary") {
		if err := reexecNewBinary(); err != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Re-exec", "failed", err.Error(), &th))
		}
	}
	return nil
}
