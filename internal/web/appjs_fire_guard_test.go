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
	"sort"
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
// Sandbox wiring (REQ-AFG-012 / REQ-AFG-014, card t1106): that directive has
// now been EXECUTED. A manifest entry did come to need a persistence-family
// control — the validation-reject submit — and the answer was a one-off
// project copy, not a looser manifest. What runs where:
//
//   - real repo root (startFireGuardServer, ProjectRoot: findRepoRoot): every
//     entry WITHOUT the sandbox-serving marker. Unchanged, to the byte.
//   - disposable copy (startFireGuardSandboxServer, a t.TempDir()-derived
//     root): the marked entries, and only those. The marker IS the routing
//     key, so the two families cannot silently swap places.
//
// Limit (spec.md §F 7) — two-surfaces: this guard does NOT claim the two
// server surfaces are equivalent. The copy carries one branch, the reject
// path; everything else stays on the real root. That asymmetry is how the
// blast radius stays zero, and reading a green here as "the copy serves the
// same thing" would be reading a claim the guard never makes.

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

	// card t1106 — the driven/excluded accounting REQ-AFG-014 (1) requires of
	// every run, and the sandbox-serving entry's observations.
	ReductionDeclared bool     `json:"reduction_declared"`
	DrivenCount       int      `json:"driven_count"`
	DrivenEntries     []string `json:"driven_entries"`
	ExcludedEntries   []string `json:"excluded_entries"`

	SandboxInvalidSet      string            `json:"s_invalid_set"`
	SandboxPaint           *firePaintReport  `json:"s_paint"`
	SandboxLoadRefErrors   []string          `json:"s_load_referenceerrors"`
	SandboxWindowRefErrors []string          `json:"s_window_referenceerrors"`
	SandboxChangedPaths    []fireChangedPath `json:"s_changed_paths"`
	SandboxSnapshotFiles   int               `json:"s_snapshot_files"`
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
	python3 := requirePython3(t)
	if out, err := exec.Command(python3, "-c", "import websockets").CombinedOutput(); err != nil {
		t.Skipf("python3 lacks the websockets module (%s) — the probe speaks CDP over websockets; pip install websockets (version pinned in the test-browser CI job) to run the app.js fire guard", strings.TrimSpace(string(out)))
	}
	return chromePath
}

// requirePython3 resolves the interpreter that executes the committed
// appjs_fire_probe.py, and skips — naming the missing prerequisite — when it
// is absent from PATH. Every test that shells out to the probe calls this
// first, gated or not: without it an absent interpreter surfaces as a
// t.Fatalf whose message accuses the committed manifest ("--lint-manifest
// rejected the committed manifest"), which reads as a real defect on a
// machine that simply has no python3. A missing prerequisite is "not measured
// here", never a failure of the thing under test.
func requirePython3(t *testing.T) (python3 string) {
	t.Helper()
	python3, err := exec.LookPath("python3")
	if err != nil {
		t.Skipf("python3 not found in PATH — it executes the committed appjs_fire_probe.py; install python3 to run the app.js fire guard")
	}
	return python3
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
	url, _ := startFireGuardServerAt(t, findRepoRoot(t))
	return url
}

// startFireGuardServerAt boots a console server on the given project root and
// returns its base URL together with the root it actually serves. The root is
// returned rather than assumed so AC-AFG-013 can MEASURE which tree each
// server serves instead of reading the wiring and trusting it.
func startFireGuardServerAt(t *testing.T, projectRoot string) (baseURL, servedRoot string) {
	t.Helper()
	// Seed a stored GLM key through the test hook: #glmKeyReveal renders only
	// when a key is configured, so without this the glm_reveal entry passes on
	// a machine with an operator key and misses everywhere else (card t1087).
	t.Setenv(glmcred.EnvTestGLMKey, fireGuardFakeGLMKey)
	cfg := Config{
		Port:           0,
		NoOpen:         true,
		ProjectRoot:    projectRoot,
		ProfileBaseDir: t.TempDir(),
	}
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	hs := httptest.NewServer(srv.Handler())
	t.Cleanup(hs.Close)
	return hs.URL, cfg.ProjectRoot
}

// provisionFireGuardSandbox builds the disposable project copy the
// validation-reject entry is served from (REQ-AFG-014 (1)) and returns its
// root.
//
// Lifetime is bound to t.TempDir() and to nothing else (REQ-AFG-014 (3)):
// the framework removes it on EVERY exit path, panics and early failures
// included. There is deliberately no defer and no trailing removal statement
// here — a trailing cleanup line is one the process may never reach, and
// AC-AFG-011 (d) checks this file for exactly that shape.
//
// Contents: .moai/config is copied because it is what POST /save's write
// seams (SyncToProjectConfig, writeProjectConfig) target, so the byte
// invariance assertion has something real to be invariant about. Nothing else
// is copied — the reject path never reads further, and a smaller copy is a
// smaller surface to keep unchanged.
func provisionFireGuardSandbox(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "sandbox-project")
	src := filepath.Join(findRepoRoot(t), ".moai", "config")
	dst := filepath.Join(root, ".moai", "config")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("create sandbox root: %v", err)
	}
	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		t.Fatalf("copy .moai/config into the sandbox root: %v", err)
	}
	return root
}

// startFireGuardSandboxServer boots the SECOND server instance — the one that
// serves the disposable copy. It is additive: startFireGuardServer's real-root
// wiring above is untouched, so every unmarked manifest entry keeps serving
// exactly the tree it served before this card (plan §A0, option (a)).
func startFireGuardSandboxServer(t *testing.T) (baseURL, sandboxRoot string) {
	t.Helper()
	root := provisionFireGuardSandbox(t)
	url, served := startFireGuardServerAt(t, root)
	return url, served
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
func runFireGuardProbe(t *testing.T, scriptPath, cdpPort, baseURL, label string, extra ...string) (int, fireProbeReport) {
	t.Helper()
	requirePython3(t)
	port := strings.TrimPrefix(strings.TrimPrefix(baseURL, "http://"), "https://")
	if host, p, err := net.SplitHostPort(port); err == nil && host == "127.0.0.1" {
		port = p
	}
	args := append([]string{scriptPath, "--cdp-port", cdpPort}, extra...)
	args = append(args, port, label)
	cmd := exec.Command("python3", args...)
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

	// card t1106: this cycle drives the real-root family and says so. The
	// reduction is an affirmative declaration, never inferred from a missing
	// sandbox base — and the driven/excluded accounting below is what stops
	// "declare everything away" from reading as a pass.
	exitCode, report := runFireGuardProbe(t, probePath, cdpPort, baseURL, "t1060-driver-runtime",
		"--primary-entries-only")

	if exitCode != 0 {
		t.Fatalf("fire probe exited %d (want 0); failures=%s missing=%s",
			exitCode, mustJSON(t, report.Failures), mustJSON(t, report.MissingSelectors))
	}
	// (a) the driven set is every entry WITHOUT the sandbox-serving marker —
	// judged as a predicate, not against a count frozen at amendment time, so
	// a later entry cannot be quietly declared away. The marked family is
	// excluded here and judged by AC-AFG-010/011/013 in their own runs.
	wantDriven, wantExcluded := fireManifestFamilies(t)
	if !report.ReductionDeclared {
		t.Error("(a) the run does not carry the reduction declaration — absence of a sandbox base must never imply a narrowed cycle")
	}
	if report.DrivenCount != len(report.DrivenEntries) {
		t.Errorf("(a) driven_count=%d disagrees with driven_entries=%v", report.DrivenCount, report.DrivenEntries)
	}
	if report.DrivenCount == 0 {
		t.Fatal("(a) the run drove nothing — an empty cycle is not a passing cycle")
	}
	if got := sortedCopy(report.DrivenEntries); !slicesEqual(got, wantDriven) {
		t.Errorf("(a) driven set %v != the unmarked manifest family %v — every unmarked entry must be driven, and no marked entry may be", got, wantDriven)
	}
	if got := sortedCopy(report.ExcludedEntries); !slicesEqual(got, wantExcluded) {
		t.Errorf("(a) excluded set %v != the marked family %v — exclusion is a reported event, and it must be exactly the sandbox-serving family", got, wantExcluded)
	}
	// every driven entry fired — the probe judge folds missing selectors
	// and collapsed indicators into failures.
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

	// Same cycle shape as the forward run: the real-root family, declared.
	exitCode, report := runFireGuardProbe(t, tamperedPath, cdpPort, baseURL, "t1060-driver-selector-miss",
		"--primary-entries-only")

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

// ── card t1106 — validation-reject submit surface ───────────────────────────

// fireGuardProbePath returns the committed probe's absolute path.
func fireGuardProbePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(findRepoRoot(t), "internal", "web", "testdata", "appjs_fire_probe.py")
}

// TestAppJsFireSandboxPairing is AC-AFG-012: the inseparability of the
// `validation-reject` effect kind and its TWO manifest markers is a MACHINE
// judgement, in both directions. (a) the committed manifest passes its own
// self-check, and (b) a synthetic entry that declares `validation-reject`
// while missing either marker is rejected with exit 1 that NAMES the missing
// condition. Forward-only would pass a rule that rejects nothing.
//
// Ungated on purpose: this is a data-loss guard (REQ-AFG-014), so it runs on
// every `go test` of the package, with no browser and no server.
func TestAppJsFireSandboxPairing(t *testing.T) {
	t.Parallel()
	requirePython3(t)
	probe := fireGuardProbePath(t)

	// (a) forward: the committed manifest satisfies the rule.
	out, err := exec.Command("python3", probe, "--lint-manifest").CombinedOutput()
	if err != nil {
		t.Fatalf("--lint-manifest rejected the committed manifest: %v\n%s", err, out)
	}

	// (b) reverse: each synthetic entry is missing one (or both) markers.
	cases := []struct {
		name      string
		entry     map[string]any
		wantNamed []string
	}{
		{
			name: "missing sandbox-serving marker",
			entry: map[string]any{
				"id": "synthetic_missing_sandbox", "line_group": nil, "page": "/settings",
				"selector": "#settings-form", "effect": "validation-reject",
				"requires_no_write_assertion": true,
				"check":                       "synthetic pairing fixture",
			},
			wantNamed: []string{"requires_sandbox_serving"},
		},
		{
			name: "missing no-write-assertion marker",
			entry: map[string]any{
				"id": "synthetic_missing_nowrite", "line_group": nil, "page": "/settings",
				"selector": "#settings-form", "effect": "validation-reject",
				"requires_sandbox_serving": true,
				"check":                    "synthetic pairing fixture",
			},
			wantNamed: []string{"requires_no_write_assertion"},
		},
		{
			name: "missing both markers",
			entry: map[string]any{
				"id": "synthetic_missing_both", "line_group": nil, "page": "/settings",
				"selector": "#settings-form", "effect": "validation-reject",
				"check": "synthetic pairing fixture",
			},
			wantNamed: []string{"requires_sandbox_serving", "requires_no_write_assertion"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := mustJSON(t, tc.entry)
			out, err := exec.Command("python3", probe, "--lint-manifest", "--extra-entry", payload).CombinedOutput()
			if err == nil {
				t.Fatalf("synthetic %s was ACCEPTED (exit 0) — an entry may not claim validation-reject without both markers; output:\n%s", tc.name, out)
			}
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("synthetic %s: want exit 1, got %v\n%s", tc.name, err, out)
			}
			for _, want := range tc.wantNamed {
				if !strings.Contains(string(out), want) {
					t.Errorf("rejection does not NAME the missing condition %q — the report must say which condition is absent; output:\n%s", want, out)
				}
			}
		})
	}
}

// TestAppJsFireSandboxRouting is AC-AFG-013: the disposable copy serves the
// NEW entry and nothing else, measured in BOTH directions. A predicate that
// only checks "a sandbox server exists" passes the world where every entry
// silently moved onto the copy — the exact blast radius option (b) was
// rejected for (plan §A0). So this test asserts, mechanically:
//
//	(a) the primary server's ProjectRoot is still findRepoRoot,
//	(b) every entry WITHOUT the sandbox-serving marker routes to the primary
//	    base and every entry WITH it routes to the sandbox base — red in
//	    either direction — with both families non-empty,
//	(c) the sandbox server's root is a disposable copy, not the repo root.
//
// Ungated: routing is a wiring property, so it needs no browser.
func TestAppJsFireSandboxRouting(t *testing.T) {
	requirePython3(t)
	primaryBase, primaryRoot := startFireGuardServerAt(t, findRepoRoot(t))
	sandboxBase, sandboxRoot := startFireGuardSandboxServer(t)

	if primaryRoot != findRepoRoot(t) {
		t.Errorf("(a) primary server ProjectRoot = %q, want findRepoRoot %q — unmarked entries must keep serving the real repo root", primaryRoot, findRepoRoot(t))
	}
	if sandboxRoot == findRepoRoot(t) {
		t.Fatalf("(c) sandbox server serves the real repo root %q — the disposable copy is the whole point", sandboxRoot)
	}

	out, err := exec.Command("python3", fireGuardProbePath(t),
		"--print-routing", "--base-url", primaryBase, "--sandbox-base-url", sandboxBase).Output()
	if err != nil {
		t.Fatalf("--print-routing failed: %v\n%s", err, out)
	}
	var routing struct {
		Routes  map[string]string `json:"routes"`
		Markers map[string]bool   `json:"sandbox_marked"`
	}
	if err := json.Unmarshal(out, &routing); err != nil {
		t.Fatalf("parse routing report: %v\n%s", err, out)
	}
	if len(routing.Routes) == 0 {
		t.Fatal("routing report is empty — nothing was measured")
	}
	var marked, unmarked int
	for id, base := range routing.Routes {
		if routing.Markers[id] {
			marked++
			if base != sandboxBase {
				t.Errorf("(b) marked entry %q routes to %q, want the sandbox base %q — a marked entry must never touch the real repo root", id, base, sandboxBase)
			}
			continue
		}
		unmarked++
		if base != primaryBase {
			t.Errorf("(b) unmarked entry %q routes to %q, want the primary base %q — blast radius must stay zero", id, base, primaryBase)
		}
	}
	if marked == 0 {
		t.Error("(b) no entry carries the sandbox-serving marker — the sandbox direction of this judgement is vacuous")
	}
	if unmarked == 0 {
		t.Error("(b) every entry carries the sandbox-serving marker — the primary direction of this judgement is vacuous")
	}
}

// firePaintReport is the probe's paint judgement for the validation-reject
// banner. Node existence alone is NOT a pass (REQ-AFG-015): a layout box must
// exist, no ancestor may hide it, and the text must be non-empty.
type firePaintReport struct {
	Found    bool    `json:"found"`
	Box      bool    `json:"box"`
	Hidden   bool    `json:"hidden"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	TextLen  int     `json:"text_len"`
	Text     string  `json:"text"`
	HiddenBy string  `json:"hidden_by"`
}

type fireChangedPath struct {
	Path   string `json:"path"`
	Change string `json:"change"`
}

// startFireGuardBrowserStack boots the prerequisites every gated
// sandbox-serving test needs: the real-root server, the second server on the
// disposable copy, and headless Chrome. It returns the two bases, the sandbox
// root, and the CDP port.
func startFireGuardBrowserStack(t *testing.T) (primaryBase, sandboxBase, sandboxRoot, cdpPort string) {
	t.Helper()
	chromePath := requireFireGuardPrereqs(t)
	primaryBase = startFireGuardServer(t)
	sandboxBase, sandboxRoot = startFireGuardSandboxServer(t)
	cdpPort = launchFireGuardChrome(t, chromePath)
	return primaryBase, sandboxBase, sandboxRoot, cdpPort
}

// TestAppJsFireValidationRejectPaints is AC-AFG-010: submitting a
// validation-failing value must leave the reject banner PAINTED — the card
// t1105 fix delivered a 2xx so htmx would swap the response in, and this is
// the guard that the swapped-in banner actually reaches the screen.
//
// The predicate is deliberately not "the node exists": the response-body layer
// already proves the banner is in the markup (transport400_swap_contract_test.go
// and siblings). What this adds is reach — a layout box, no hiding ancestor,
// non-empty text, and zero ReferenceErrors in the same window. The
// --inject-banner-hidden mutant below is what keeps that predicate honest: a
// banner rendered hidden must NOT pass.
func TestAppJsFireValidationRejectPaints(t *testing.T) {
	primaryBase, sandboxBase, sandboxRoot, cdpPort := startFireGuardBrowserStack(t)
	probe := fireGuardProbePath(t)

	exitCode, report := runFireGuardProbe(t, probe, cdpPort, primaryBase, "t1106-reject-paints",
		"--sandbox-base-url", sandboxBase, "--sandbox-root", sandboxRoot)
	if exitCode != 0 {
		t.Fatalf("probe exited %d (want 0); failures=%s", exitCode, mustJSON(t, report.Failures))
	}
	if report.SandboxInvalidSet != "bogus" {
		t.Fatalf("the probe did not place a validation-failing value in the form (s_invalid_set=%q) — a submit that could have SUCCEEDED measures the wrong path", report.SandboxInvalidSet)
	}
	paint := report.SandboxPaint
	if paint == nil {
		t.Fatal("probe reported no paint judgement for the validation-reject banner")
	}
	if !paint.Found {
		t.Fatalf("no banner node after the reject render: %s", mustJSON(t, paint))
	}
	if !paint.Box {
		t.Errorf("(a) banner has no layout box (%gx%g) — it is in the DOM but not on the screen", paint.Width, paint.Height)
	}
	if paint.Hidden {
		t.Errorf("(a) banner is hidden by an ancestor (%s) — reaching the DOM is not reaching the screen", paint.HiddenBy)
	}
	if paint.TextLen == 0 {
		t.Error("(b) banner text is empty — an empty banner tells the user nothing")
	}
	if len(report.SandboxLoadRefErrors) != 0 || len(report.SandboxWindowRefErrors) != 0 {
		t.Errorf("(c) ReferenceErrors in the reject window: load=%v submit=%v", report.SandboxLoadRefErrors, report.SandboxWindowRefErrors)
	}

	// Mutant probe (AC-AFG-010): weaken the surface, not the predicate — the
	// same run with the banner forced hidden must go red. A predicate that
	// only checks node existence passes this mutant, and is unadoptable.
	mutantExit, mutantReport := runFireGuardProbe(t, probe, cdpPort, primaryBase, "t1106-reject-paints-mutant",
		"--sandbox-base-url", sandboxBase, "--sandbox-root", sandboxRoot, "--inject-banner-hidden")
	if mutantExit != 1 {
		t.Fatalf("a banner rendered hidden exited %d (want 1) — the paint predicate is too shallow to adopt; report=%s", mutantExit, mustJSON(t, mutantReport.SandboxPaint))
	}
}

// TestAppJsFireValidationRejectNoWrites is AC-AFG-011: the submit happens on a
// disposable copy and the copy stays byte-identical across it, judged in BOTH
// directions.
//
//	(a) the serving root is the disposable copy, not findRepoRoot,
//	(b)(c) the probe compares the copy's bytes either side of the submit and
//	       exits 1 naming the path when one changes — OBSERVED by injecting a
//	       one-byte write under .moai/config/sections/, the subtree
//	       handleSave's write seams actually reach, so the reverse direction
//	       cannot be satisfied by an implementation that excludes it,
//	(d) the copy's root derives from t.TempDir() and its lifetime hangs on no
//	    defer and no trailing removal statement — cleanup that a panic or an
//	    early failure can skip is not cleanup.
func TestAppJsFireValidationRejectNoWrites(t *testing.T) {
	primaryBase, sandboxBase, sandboxRoot, cdpPort := startFireGuardBrowserStack(t)
	probe := fireGuardProbePath(t)

	// (a) + (d) — the root the second server serves.
	if sandboxRoot == findRepoRoot(t) {
		t.Fatalf("(a) the submit is being served from the real repo root %q", sandboxRoot)
	}
	tempParent := filepath.Dir(t.TempDir())
	if got := filepath.Dir(filepath.Dir(sandboxRoot)); got != tempParent {
		t.Errorf("(d) sandbox root %q is not derived from t.TempDir() (%q != %q) — its removal must be the framework's job, on every exit path", sandboxRoot, got, tempParent)
	}
	driverSrc, err := os.ReadFile(filepath.Join(findRepoRoot(t), "internal", "web", "appjs_fire_guard_test.go"))
	if err != nil {
		t.Fatalf("read driver source: %v", err)
	}
	// The needles are assembled rather than written out: a literal
	// "os.Remove"+"All" in this file would be matched by this very scan, and
	// the test would fail on its own fixture instead of on the driver.
	for _, forbidden := range []string{"defer os." + "Remove", "os." + "RemoveAll"} {
		if strings.Contains(string(driverSrc), forbidden) {
			t.Errorf("(d) driver carries %q — the sandbox copy's lifetime must hang on t.TempDir()/t.Cleanup alone; a trailing removal is a line the process may never reach", forbidden)
		}
	}

	// (b) forward: the real submit changes nothing.
	exitCode, report := runFireGuardProbe(t, probe, cdpPort, primaryBase, "t1106-reject-nowrites",
		"--sandbox-base-url", sandboxBase, "--sandbox-root", sandboxRoot)
	if exitCode != 0 {
		t.Fatalf("probe exited %d (want 0); failures=%s", exitCode, mustJSON(t, report.Failures))
	}
	if report.SandboxSnapshotFiles == 0 {
		t.Fatal("the byte-invariance comparison covered zero files — an invariance claim over an empty set asserts nothing")
	}
	if len(report.SandboxChangedPaths) != 0 {
		t.Fatalf("the validation-reject path wrote to the sandbox: %s", mustJSON(t, report.SandboxChangedPaths))
	}

	// (c) reverse: a single injected byte under the subtree the write seams
	// reach must be caught and NAMED.
	const injected = ".moai/config/sections/quality.yaml"
	mutantExit, mutantReport := runFireGuardProbe(t, probe, cdpPort, primaryBase, "t1106-reject-nowrites-mutant",
		"--sandbox-base-url", sandboxBase, "--sandbox-root", sandboxRoot, "--inject-sandbox-write", injected)
	if mutantExit != 1 {
		t.Fatalf("a one-byte write under %s exited %d (want 1) — the no-write assertion does not actually watch the paths the write seams reach", injected, mutantExit)
	}
	named := false
	for _, c := range mutantReport.SandboxChangedPaths {
		if c.Path == injected {
			named = true
		}
	}
	if !named {
		t.Fatalf("the failure does not NAME the changed path %s: %s", injected, mustJSON(t, mutantReport.SandboxChangedPaths))
	}
}

// fireManifestFamilies reads the committed manifest and splits it the way the
// routing key does: entries WITHOUT the sandbox-serving marker, and entries
// WITH it. The split is derived from the manifest rather than written down
// here, so a manifest that grows a tenth entry moves this expectation with it
// instead of leaving a frozen number behind.
func fireManifestFamilies(t *testing.T) (unmarked, marked []string) {
	t.Helper()
	requirePython3(t)
	out, err := exec.Command("python3", fireGuardProbePath(t),
		"--print-routing", "--base-url", "http://primary.invalid", "--sandbox-base-url", "http://sandbox.invalid").Output()
	if err != nil {
		t.Fatalf("read manifest families via --print-routing: %v\n%s", err, out)
	}
	var routing struct {
		Markers map[string]bool `json:"sandbox_marked"`
	}
	if err := json.Unmarshal(out, &routing); err != nil {
		t.Fatalf("parse routing report: %v\n%s", err, out)
	}
	if len(routing.Markers) == 0 {
		t.Fatal("the manifest reports no entries — nothing to split")
	}
	for id, isMarked := range routing.Markers {
		if isMarked {
			marked = append(marked, id)
			continue
		}
		unmarked = append(unmarked, id)
	}
	return sortedCopy(unmarked), sortedCopy(marked)
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestAppJsFireReductionDeclaration pins the affirmative half of REQ-AFG-014
// (1): the ABSENCE of a sandbox base must not be read as a reduction. Without
// the declaration, a manifest entry that requires sandbox serving and has
// nowhere to be served from is a caller-wiring fault — exit 2, with the entry
// NAMED — never a silent skip and never a fallback onto the real repo root.
//
// Ungated: the contract is decided before any server or browser is contacted,
// so an unreachable port is enough to reach it.
func TestAppJsFireReductionDeclaration(t *testing.T) {
	t.Parallel()
	requirePython3(t)
	_, marked := fireManifestFamilies(t)
	if len(marked) == 0 {
		t.Skip("no sandbox-serving entry in the manifest — this contract has nothing to bind; recorded as not measured, not as a pass")
	}

	out, err := exec.Command("python3", fireGuardProbePath(t),
		"--cdp-port", "1", "--base-url", "http://127.0.0.1:1", "1", "t1106-no-declaration").CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 2 {
		t.Fatalf("want exit 2 (caller wiring fault), got %v\n%s", err, out)
	}
	// Exit 2 alone does not pin this: every machine fault is exit 2, so an
	// implementation that skipped the marked entry and then tripped over an
	// unreachable browser would look identical. The named cause is what
	// separates them.
	if !strings.Contains(string(out), "--sandbox-base-url") {
		t.Errorf("the fault does not name the missing wiring (--sandbox-base-url) — any other exit 2 would read the same: %s", out)
	}
	for _, id := range marked {
		if !strings.Contains(string(out), id) {
			t.Errorf("the fault does not NAME the unservable entry %q: %s", id, out)
		}
	}
}
