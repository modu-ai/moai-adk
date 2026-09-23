package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/glmcred"
)

// SPEC-APPJS-FIRE-GUARD-001 (card t1060) — runtime fire guard for the
// console's app.js button handlers.
//
// The static sibling guard SPEC-APPJS-IIFE-GUARD-001
// (appjs_iife_scope_test.go) proves handlers never cross the IIFE boundary
// syntactically. This driver is orthogonal to it: it proves the handlers
// actually FIRE in a real Chrome against the real server surface. Neither
// guard's green implies the other's — a static-scope pass says nothing about
// runtime firing, and a fire-guard red does not pinpoint the offending
// identifier (static-scope stays the owner of precise cause reports).
//
// What this guard does NOT cover (spec.md §F): interactions outside the
// committed probe's manifest; paths the probe's single scenario never
// exercises; browsers other than Chrome; static-cause attribution.
//
// Architecture (plan §A): the committed probe
// testdata/appjs_fire_probe.py drives headless Chrome over CDP against an
// in-process server (NewServer + Handler, this package's test surface). The
// driver owns process lifetime through t.Cleanup — server, Chrome, and probe
// all die with the test, on every exit path.
//
// Sandbox note (REQ-AFG-012): the manifest exercises reversible effects only
// (visibility/label/tab/swap — save-family controls are excluded in the
// probe), so serving the real repo root is safe. The ProjectRoot field below
// is the wiring point: if a manifest entry ever needs a persistence-family
// control, the driver MUST switch it to a one-off project copy instead of
// loosening the manifest.

const fireGuardGateEnv = "MOAI_BROWSER_GUARD"

// fireProbeReport is the subset of the probe's JSON report the driver asserts
// on. The probe prints the full report to stdout and exits 0/1/2.
type fireProbeReport struct {
	Exit             int                `json:"exit"`
	Failures         []fireProbeFailure `json:"failures"`
	MissingSelectors []fireProbeMissing `json:"missing_selectors"`

	P1LoadReferenceErrors []string `json:"p1_load_referenceerrors"`
	P5SwapReferenceErrors []string `json:"p5_swap_referenceerrors"`
	P7LoadReferenceErrors []string `json:"p7_load_referenceerrors"`

	P6PopoverAfterSwapFired bool `json:"p6_popover_after_swap_fired"`
}

type fireProbeFailure struct {
	Entry    string `json:"entry"`
	Reason   string `json:"reason"`
	Selector string `json:"selector"`
	Detail   string `json:"detail"`
	Check    string `json:"check"`
}

type fireProbeMissing struct {
	Entry    string `json:"entry"`
	Selector string `json:"selector"`
}

// TestAppJsFireManifestInventoryCount is the manifest's continued-firing
// guard (verification-completeness §1.3): it runs ungated on every `go test`
// of this package. The probe declares INVENTORY_TOTAL — the number of
// click/change-family addEventListener groups in app.js its manifest covers.
// When app.js gains or loses a registration group, the live count drifts from
// the declared total and this test goes red: a stale manifest is never
// silently green — the same selector-survival contract the probe enforces at
// runtime, applied to the manifest itself.
func TestAppJsFireManifestInventoryCount(t *testing.T) {
	t.Parallel()
	src := readEmbeddedAsset(t, "app.js")
	re := regexp.MustCompile(`addEventListener\((['"])(click|submit|change|input)`)
	live := len(re.FindAllString(src, -1))

	probeSrc, err := os.ReadFile(filepath.Join("testdata", "appjs_fire_probe.py"))
	if err != nil {
		t.Fatalf("read committed probe: %v", err)
	}
	m := regexp.MustCompile(`(?m)^INVENTORY_TOTAL = (\d+)$`)
	mm := m.FindStringSubmatch(string(probeSrc))
	if mm == nil {
		t.Fatal("committed probe does not declare INVENTORY_TOTAL — the manifest count cross-check cannot run")
	}
	var declared int
	if _, err := fmt.Sscanf(mm[1], "%d", &declared); err != nil {
		t.Fatalf("parse INVENTORY_TOTAL %q: %v", mm[1], err)
	}
	if declared < 1 {
		t.Fatalf("INVENTORY_TOTAL = %d: a zero-entry manifest measures nothing (REQ-AFG-003)", declared)
	}
	if live != declared {
		t.Fatalf("app.js has %d click/change-family addEventListener groups but the probe manifest declares INVENTORY_TOTAL=%d — update testdata/appjs_fire_probe.py (entries or reasoned exclusions) to cover the new surface", live, declared)
	}
}

// requireFireGuardPrereqs enforces the REQ-AFG-001 gate: the full cycle runs
// only on explicit opt-in with chrome, python3, and websockets all present.
// Every skip names the missing prerequisite in English (error_messages: en) —
// a reasonless skip is indistinguishable from a pass to whoever reads the
// guard's survival.
func requireFireGuardPrereqs(t *testing.T) (chromePath string) {
	t.Helper()
	if os.Getenv(fireGuardGateEnv) != "1" {
		t.Skipf("%s is not set to 1 — the app.js runtime fire guard runs only on explicit opt-in (set %s=1); this skip records \"not measured here\", not a pass", fireGuardGateEnv, fireGuardGateEnv)
	}
	chromePath, err := findChrome()
	if err != nil {
		t.Skipf("%v — install Google Chrome (or point MOAI_BROWSER_GUARD_CHROME at a chrome binary) to run the app.js fire guard", err)
	}
	python3, err := exec.LookPath("python3")
	if err != nil {
		t.Skipf("python3 not found in PATH — it executes the committed appjs_fire_probe.py; install python3 to run the app.js fire guard")
	}
	if out, err := exec.Command(python3, "-c", "import websockets").CombinedOutput(); err != nil {
		t.Skipf("python3 lacks the websockets module (%s) — the probe speaks CDP over websockets; pip install websockets (version pinned in the test-browser CI job) to run the app.js fire guard", strings.TrimSpace(string(out)))
	}
	return chromePath
}

// findChrome locates a chrome binary: MOAI_BROWSER_GUARD_CHROME first, then
// the usual PATH names, then platform install locations. The error names
// `chrome` because the skip reason must name the missing prerequisite.
func findChrome() (string, error) {
	if p := os.Getenv("MOAI_BROWSER_GUARD_CHROME"); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
		return "", fmt.Errorf("chrome: MOAI_BROWSER_GUARD_CHROME=%q does not point at an executable", p)
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chrome"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	candidates := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("chrome executable not found (searched MOAI_BROWSER_GUARD_CHROME, PATH google-chrome/chromium, standard install locations)")
}

// findRepoRoot walks up from the working directory until it finds go.mod —
// `go test` runs with the package directory as CWD, and the served console
// needs the repo root so /settings, /todo, and /specs render their real
// interactive surface (identical to what the real binary serves).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the working directory — cannot locate the repo root to serve")
		}
		dir = parent
	}
}

// fireGuardFakeGLMKey is an obviously fake value; it only has to make the
// reveal control render and never reaches any GLM endpoint.
const fireGuardFakeGLMKey = "test-fire-guard-fake-glm-key"

// startFireGuardServer boots the in-process console server on a random
// loopback port and registers its teardown in t.Cleanup. It returns the base
// URL Chrome will load.
func startFireGuardServer(t *testing.T) string {
	t.Helper()
	// Seed a stored GLM key through the test hook: #glmKeyReveal renders only
	// when a key is configured, so without this the glm_reveal entry passes on
	// a machine with an operator key and misses everywhere else (card t1087).
	t.Setenv(glmcred.EnvTestGLMKey, fireGuardFakeGLMKey)
	cfg := Config{
		Port:           0,
		NoOpen:         true,
		ProjectRoot:    findRepoRoot(t),
		ProfileBaseDir: t.TempDir(),
	}
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	hs := httptest.NewServer(srv.Handler())
	t.Cleanup(hs.Close)
	return hs.URL
}

// launchFireGuardChrome starts headless Chrome with an ephemeral CDP port and
// registers its teardown in t.Cleanup. It returns the CDP port read from the
// browser's DevToolsActivePort file (the documented discovery mechanism for
// --remote-debugging-port=0). Teardown kills and reaps Chrome's whole process
// group, not just the parent, so no child still writes into the profile when
// its TempDir is removed (card t1098). Chrome's combined stdout/stderr is
// captured and its tail logged when the test fails.
func launchFireGuardChrome(t *testing.T, chromePath string) string {
	t.Helper()
	profile := t.TempDir()
	cmd := exec.Command(chromePath,
		"--headless=new",
		"--remote-debugging-port=0",
		"--user-data-dir="+profile,
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-gpu",
		"about:blank",
	)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	fireGuardSetProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start headless Chrome: %v", err)
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	// Cleanup is LIFO: this runs before the profile TempDir's RemoveAll.
	t.Cleanup(func() {
		fireGuardKillAndReap(t, cmd, done)
		if t.Failed() {
			t.Logf("chrome output tail:\n%s", fireGuardOutputTail(&output))
		}
	})

	portFile := filepath.Join(profile, "DevToolsActivePort")
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(portFile); err == nil {
			if line, _, ok := strings.Cut(strings.TrimSpace(string(data)), "\n"); ok && line != "" {
				if _, err := net.LookupPort("tcp", line); err == nil {
					return line
				}
			}
		}
		select {
		case <-done:
			t.Fatalf("headless Chrome exited before exposing a CDP port; chrome output tail:\n%s", fireGuardOutputTail(&output))
		case <-time.After(100 * time.Millisecond):
		}
	}
	t.Fatal("headless Chrome did not write DevToolsActivePort within 20s")
	return ""
}

// fireGuardOutputTail returns the last 4096 bytes of Chrome's captured output.
// Callers read it only after the Wait goroutine closed done, so the exec copy
// goroutines have finished writing.
func fireGuardOutputTail(b *bytes.Buffer) string {
	const tailMax = 4096
	out := b.Bytes()
	if len(out) > tailMax {
		out = out[len(out)-tailMax:]
	}
	return string(out)
}

// runFireGuardProbe executes the committed (or a caller-supplied) probe
// script and returns its exit code plus the parsed report. Probe stdout is
// the evidence surface — it is echoed verbatim into the test log so a red
// run carries its own report.
func runFireGuardProbe(t *testing.T, scriptPath, cdpPort, baseURL, label string) (int, fireProbeReport) {
	t.Helper()
	port := strings.TrimPrefix(strings.TrimPrefix(baseURL, "http://"), "https://")
	if host, p, err := net.SplitHostPort(port); err == nil && host == "127.0.0.1" {
		port = p
	}
	cmd := exec.Command("python3", scriptPath, "--cdp-port", cdpPort, port, label)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	t.Logf("probe %s stdout:\n%s", label, stdout.String())
	if stderr.Len() > 0 {
		t.Logf("probe %s stderr:\n%s", label, stderr.String())
	}
	var report fireProbeReport
	_ = json.Unmarshal([]byte(stdout.String()), &report) // machine-fault JSON (exit 2) has no exit field
	exitCode := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			t.Fatalf("probe %s failed to start: %v", label, runErr)
		}
		exitCode = exitErr.ExitCode()
	}
	if exitCode == 2 {
		return exitCode, report // machine fault: caller decides how to surface it
	}
	if report.Exit != exitCode {
		t.Fatalf("probe %s exit code %d disagrees with report exit field %d", label, exitCode, report.Exit)
	}
	return exitCode, report
}

// TestAppJsHandlersFireRuntime is AC-AFG-001's green path: with the gate on
// and all prerequisites present, the committed probe runs the full cycle
// against the in-process server surface and must report (a) a non-zero
// manifest with every entry fired, (b) zero ReferenceErrors on the load and
// swap windows, and (c) at least one indicator exercised AFTER the hx-boost
// swap. The three assertions hang together — indicator assertions alone would
// let an empty manifest pass, ReferenceError assertions alone would let a
// run where no button was ever clicked pass.
func TestAppJsHandlersFireRuntime(t *testing.T) {
	chromePath := requireFireGuardPrereqs(t)
	baseURL := startFireGuardServer(t)
	cdpPort := launchFireGuardChrome(t, chromePath)

	probePath := filepath.Join(findRepoRoot(t), "internal", "web", "testdata", "appjs_fire_probe.py")

	// The lint self-check rides every gated run: a manifest that fails its own
	// structural contract (allowlist, coverage count, post-swap entry) must
	// fail here first.
	lint := exec.Command("python3", probePath, "--lint-manifest")
	if out, err := lint.CombinedOutput(); err != nil {
		t.Fatalf("--lint-manifest rejected the committed manifest: %v\n%s", err, out)
	}

	exitCode, report := runFireGuardProbe(t, probePath, cdpPort, baseURL, "t1060-driver-runtime")

	if exitCode != 0 {
		t.Fatalf("fire probe exited %d (want 0); failures=%s missing=%s",
			exitCode, mustJSON(t, report.Failures), mustJSON(t, report.MissingSelectors))
	}
	// (a) every manifest entry fired — the probe judge folds missing selectors
	// and collapsed indicators into failures; entry count > 0 is enforced
	// upstream by the lint self-check above and the ungated inventory test.
	if len(report.Failures) != 0 {
		t.Fatalf("probe reported failures despite exit 0: %s", mustJSON(t, report.Failures))
	}
	if len(report.MissingSelectors) != 0 {
		t.Fatalf("probe reported missing selectors despite exit 0: %s", mustJSON(t, report.MissingSelectors))
	}
	// (b) zero ReferenceErrors on load and swap windows, asserted directly.
	if len(report.P1LoadReferenceErrors) != 0 {
		t.Errorf("load ReferenceErrors on /settings: %v", report.P1LoadReferenceErrors)
	}
	if len(report.P5SwapReferenceErrors) != 0 {
		t.Errorf("ReferenceErrors around the hx-boost swap: %v", report.P5SwapReferenceErrors)
	}
	if len(report.P7LoadReferenceErrors) != 0 {
		t.Errorf("load ReferenceErrors on /specs: %v", report.P7LoadReferenceErrors)
	}
	// (c) an indicator exercised after the hx-boost swap (REQ-AFG-007).
	if !report.P6PopoverAfterSwapFired {
		t.Error("no indicator fired after the hx-boost swap (p6_popover_after_swap_fired=false) — REQ-AFG-007 requires a post-swap indicator")
	}
}

// TestAppJsHandlersFireSelectorMiss is AC-AFG-004's green path: a manifest
// selector that matches nothing must turn the run red AND be named. The
// subtest tampers a one-shot copy of the committed probe (the committed file
// itself is never modified), points it at the same live surface, and asserts
// exit 1 with the stale selector reported by entry id.
func TestAppJsHandlersFireSelectorMiss(t *testing.T) {
	chromePath := requireFireGuardPrereqs(t)
	baseURL := startFireGuardServer(t)
	cdpPort := launchFireGuardChrome(t, chromePath)

	committed := filepath.Join(findRepoRoot(t), "internal", "web", "testdata", "appjs_fire_probe.py")
	src, err := os.ReadFile(committed)
	if err != nil {
		t.Fatalf("read committed probe: %v", err)
	}
	const liveSelector, staleSelector = "[data-copy]", "[data-copy-nonexistent]"
	tampered := strings.ReplaceAll(string(src), liveSelector, staleSelector)
	if !strings.Contains(tampered, staleSelector) {
		t.Fatalf("tamper produced no change — the probe no longer carries %q; update the selector-miss fixture", liveSelector)
	}
	tamperedPath := filepath.Join(t.TempDir(), "appjs_fire_probe_stale.py")
	if err := os.WriteFile(tamperedPath, []byte(tampered), 0o600); err != nil {
		t.Fatalf("write tampered probe copy: %v", err)
	}

	exitCode, report := runFireGuardProbe(t, tamperedPath, cdpPort, baseURL, "t1060-driver-selector-miss")

	if exitCode != 1 {
		t.Fatalf("stale-selector probe exited %d (want 1) — a manifest selector matching nothing must be red, never a quiet zero; report=%s", exitCode, mustJSON(t, report))
	}
	named := false
	for _, m := range report.MissingSelectors {
		if m.Entry == "copy_button" || strings.Contains(m.Selector, staleSelector) {
			named = true
		}
	}
	if !named {
		t.Fatalf("stale-selector report does not name the failed selector (copy_button/%s): missing=%s", staleSelector, mustJSON(t, report.MissingSelectors))
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
