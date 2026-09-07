package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Fixtures
//
// The corpus is synthetic and lives in t.TempDir(); the real .moai/specs/ tree
// is never read, so no assertion here depends on the repository's own standing
// inventory size (REQ-SLGS-003 — no warning-count literal is frozen anywhere).
//
// Two injection vectors, both confirmed by measurement before use:
//   - a SPEC-*/ directory carrying no spec.md yields exactly one NON-ADVISORY
//     warning, SpecsDirMissingSpecFile (a directory-level finding, so the era
//     demotion that marks per-SPEC findings advisory never reaches it);
//   - a modern-era spec.md with its "Out of Scope" section removed yields one
//     ERROR-severity MissingExclusions.
// ---------------------------------------------------------------------------

const slOutOfScopeHeading = "## 5. Scope and Out of Scope"

const slCleanSpecMD = `---
id: SPEC-PROBE-001
title: "Probe"
version: "0.1.0"
status: in-progress
created: 2026-09-01
updated: 2026-09-01
author: probe
priority: P1
phase: "v3.1.0"
module: "internal/probe"
lifecycle: spec-anchored
tags: "probe"
tier: S
---

# SPEC: Probe

## 2. Requirements

- **REQ-PRB-001** (Ubiquitous): The system shall do the thing.

## 3. Acceptance

### AC-PRB-001 — thing happens (maps REQ-PRB-001)

- Given a thing
- When it runs
- Then it happens

## 5. Scope and Out of Scope

### 5.2 Out of Scope

- nothing else
`

const slProgressMD = `# progress

## §E.2 Run-phase Evidence

none

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: "a1b2c3d4e5f6"
`

// slNewCorpus creates a project root holding .moai/specs/ with one baseline SPEC.
func slNewCorpus(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-PROBE-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir corpus: %v", err)
	}
	slWriteFile(t, filepath.Join(specDir, "spec.md"), slCleanSpecMD)
	slWriteFile(t, filepath.Join(specDir, "progress.md"), slProgressMD)
	return root
}

func slWriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// slInjectWarning adds one SpecsDirMissingSpecFile finding.
func slInjectWarning(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("inject warning: %v", err)
	}
	// A file inside keeps the directory non-empty and realistic.
	slWriteFile(t, filepath.Join(dir, "research.md"), "# umbrella doc\n")
	return dir
}

// slInjectError removes the Out of Scope section, producing MissingExclusions.
func slInjectError(t *testing.T, root string) {
	t.Helper()
	idx := strings.Index(slCleanSpecMD, slOutOfScopeHeading)
	if idx < 0 {
		t.Fatalf("fixture drift: %q not found in the clean SPEC body", slOutOfScopeHeading)
	}
	slWriteFile(t, filepath.Join(root, ".moai", "specs", "SPEC-PROBE-001", "spec.md"), slCleanSpecMD[:idx])
}

func slRestoreClean(t *testing.T, root string) {
	t.Helper()
	slWriteFile(t, filepath.Join(root, ".moai", "specs", "SPEC-PROBE-001", "spec.md"), slCleanSpecMD)
}

func slFileSHA256(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("hash %s: %v", path, err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// slRunLint executes the real cobra command in `dir`, returning combined
// stdout, stderr and the exit code the process would have used.
func slRunLint(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	t.Chdir(dir)

	cmd := newSpecLintCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	combined := out.String() + errBuf.String()
	if err == nil {
		return combined, 0
	}
	var ec *exitCodeError
	if errors.As(err, &ec) {
		return combined + "\n" + err.Error(), ec.ExitCode()
	}
	return combined + "\n" + err.Error(), 1
}

// slCaptureBaseline captures the corpus's current state as the baseline file.
func slCaptureBaseline(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "baseline.json")
	out, code := slRunLint(t, root, "--baseline", path, "--update-baseline", "--reason", "initial capture")
	if code != 0 {
		t.Fatalf("baseline capture failed (rc=%d):\n%s", code, out)
	}
	fi, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("baseline file not written: %v", statErr)
	}
	if fi.Size() == 0 {
		t.Fatalf("baseline file is empty")
	}
	return path
}

// ---------------------------------------------------------------------------
// AC-SLGS-005 — a new non-advisory warning turns the gate red and names the
// rule; removing it returns the gate to green.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_InjectedWarningTurnsRedAndNamesTheRule(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	// Precondition: the unchanged corpus is green against its own baseline.
	// Without this the red below could come from a broken capture rather than
	// from the injection.
	if preOut, preCode := slRunLint(t, root, "--baseline", baseline); preCode != 0 {
		t.Fatalf("precondition: unchanged corpus must be green, rc=%d\n%s", preCode, preOut)
	}

	dir := slInjectWarning(t, root, "SPEC-INJECTED-001")

	redOut, redCode := slRunLint(t, root, "--baseline", baseline)
	if redCode != 1 {
		t.Fatalf("injected violation must exit 1, got %d\n%s", redCode, redOut)
	}
	if !strings.Contains(redOut, "SpecsDirMissingSpecFile") {
		t.Errorf("output must name the rule that grew:\n%s", redOut)
	}
	if !strings.Contains(redOut, "+1") {
		t.Errorf("output must show the per-rule delta (+1):\n%s", redOut)
	}
	if !strings.Contains(strings.ToUpper(redOut), "EXCEEDED") {
		t.Errorf("output must state the baseline was exceeded:\n%s", redOut)
	}

	// Remove the violation — the gate returns to green.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove injected dir: %v", err)
	}
	greenOut, greenCode := slRunLint(t, root, "--baseline", baseline)
	if greenCode != 0 {
		t.Fatalf("removing the violation must return rc=0, got %d\n%s", greenCode, greenOut)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-006 — the standing inventory passes and stays observable.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_UnchangedCorpusPassesAndPrintsWarningTotal(t *testing.T) {
	root := slNewCorpus(t)
	// Give the corpus a standing non-advisory inventory so the pass is not
	// vacuous: something must actually be absorbed by the baseline.
	slInjectWarning(t, root, "SPEC-STANDING-001")
	slInjectWarning(t, root, "SPEC-STANDING-002")

	baseline := slCaptureBaseline(t, root)

	// The captured baseline must actually record the standing inventory —
	// a baseline of zero rules would pass for the wrong reason.
	b := slReadBaseline(t, baseline)
	if b.Rules["SpecsDirMissingSpecFile"] != 2 {
		t.Fatalf("precondition: baseline must record the standing inventory, got %v", b.Rules)
	}

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 0 {
		t.Fatalf("standing inventory must pass, rc=%d\n%s", code, out)
	}
	if !strings.Contains(out, "warning(s)") {
		t.Errorf("the warning total must remain visible while green:\n%s", out)
	}
	if !strings.Contains(strings.ToUpper(out), "OK") {
		t.Errorf("output must state the baseline verdict:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-007 — a decrease passes, is reported, and does NOT rewrite the file.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_DecreasePassesAndLeavesFileByteIdentical(t *testing.T) {
	root := slNewCorpus(t)
	dir := slInjectWarning(t, root, "SPEC-STANDING-001")
	slInjectWarning(t, root, "SPEC-STANDING-002")
	baseline := slCaptureBaseline(t, root)

	before := slFileSHA256(t, baseline)

	// Fix one violation: the count drops from 2 to 1.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove: %v", err)
	}

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 0 {
		t.Fatalf("a decrease must pass, rc=%d\n%s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "decreas") {
		t.Errorf("the improvement must be reported:\n%s", out)
	}

	after := slFileSHA256(t, baseline)
	if before != after {
		t.Errorf("the baseline file must not shrink implicitly — a silent rewrite bypasses the rebaseline audit.\nbefore=%s\nafter =%s", before, after)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-008 — rebaseline is explicit, requires a non-empty reason, and
// records tree SHA + date + reason.
// ---------------------------------------------------------------------------

func TestSpecLintUpdateBaseline_RejectsMissingOrEmptyReason(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)
	before := slFileSHA256(t, baseline)

	cases := []struct {
		name string
		args []string
	}{
		{"reason omitted", []string{"--baseline", baseline, "--update-baseline"}},
		{"reason empty", []string{"--baseline", baseline, "--update-baseline", "--reason", ""}},
		{"reason whitespace only", []string{"--baseline", baseline, "--update-baseline", "--reason", "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, code := slRunLint(t, root, tc.args...)
			if code != 3 {
				t.Errorf("must be rejected with exit 3, got %d\n%s", code, out)
			}
			if after := slFileSHA256(t, baseline); after != before {
				t.Errorf("a rejected rebaseline must leave the file byte-unchanged")
			}
		})
	}
}

func TestSpecLintUpdateBaseline_RecordsTreeSHADateAndReason(t *testing.T) {
	root := slNewCorpus(t)
	slInjectWarning(t, root, "SPEC-STANDING-001")

	path := filepath.Join(t.TempDir(), "baseline.json")
	out, code := slRunLint(t, root, "--baseline", path, "--update-baseline", "--reason", "absorb t518 population shift")
	if code != 0 {
		t.Fatalf("rebaseline must succeed, rc=%d\n%s", code, out)
	}

	b := slReadBaseline(t, path)
	if strings.TrimSpace(b.Reason) != "absorb t518 population shift" {
		t.Errorf("reason not recorded: %q", b.Reason)
	}
	if strings.TrimSpace(b.TreeSHA) == "" {
		t.Errorf("tree_sha not recorded: %+v", b)
	}
	if len(b.UpdatedAt) != len("2026-09-08") {
		t.Errorf("updated_at must be an ISO date, got %q", b.UpdatedAt)
	}
	if b.Rules["SpecsDirMissingSpecFile"] != 1 {
		t.Errorf("recomputed counts missing: %v", b.Rules)
	}
	// The same three fields must be reported to the operator, not only written.
	for _, want := range []string{"absorb t518 population shift", b.TreeSHA, b.UpdatedAt} {
		if !strings.Contains(out, want) {
			t.Errorf("rebaseline output must report %q:\n%s", want, out)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-009 — an error fails regardless of the baseline, and the reported
// cause is the error, not a baseline increase. The cause distinction is what
// makes the error-BEFORE-baseline ordering observable: both exit 1.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_ErrorFailsRegardlessOfBaseline(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	if preOut, preCode := slRunLint(t, root, "--baseline", baseline); preCode != 0 {
		t.Fatalf("precondition: corpus must be green before the error injection, rc=%d\n%s", preCode, preOut)
	}

	slInjectError(t, root)

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("an error must fail the gate even with a passing baseline, rc=%d\n%s", code, out)
	}
	if !strings.Contains(out, "MissingExclusions") {
		t.Errorf("the error must be visible in the output:\n%s", out)
	}
	if !strings.Contains(strings.ToUpper(out), "ERROR-GATED") {
		t.Errorf("the reported cause must be the error, not a baseline increase:\n%s", out)
	}
	if strings.Contains(strings.ToUpper(out), "EXCEEDED") {
		t.Errorf("an error-gated failure must not be reported as a baseline exceedance:\n%s", out)
	}

	slRestoreClean(t, root)
	if postOut, postCode := slRunLint(t, root, "--baseline", baseline); postCode != 0 {
		t.Fatalf("removing the error must return rc=0, got %d\n%s", postCode, postOut)
	}
}

// The error-BEFORE-baseline ORDERING is only observable when both causes are
// present at once: with an error alone, an implementation that consults the
// baseline first still falls through to the error branch and reports the same
// thing. Added after a mutant that swapped the two branches survived the whole
// suite — this is the case that kills it.
func TestSpecLintBaseline_ErrorOutranksASimultaneousIncrease(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	slInjectWarning(t, root, "SPEC-BOTH-001") // a baseline increase
	slInjectError(t, root)                    // and an error, at the same time

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("both causes present must exit 1, rc=%d\n%s", code, out)
	}
	// Precondition: the increase really is present, or this asserts nothing.
	if !strings.Contains(out, "SpecsDirMissingSpecFile") {
		t.Fatalf("precondition: the injected increase must be in the report:\n%s", out)
	}
	if !strings.Contains(out, "MissingExclusions") {
		t.Fatalf("precondition: the injected error must be in the report:\n%s", out)
	}

	if !strings.Contains(strings.ToUpper(out), "ERROR-GATED") {
		t.Errorf("the error must be reported as the cause, ahead of the ratchet:\n%s", out)
	}
	if strings.Contains(strings.ToUpper(out), "EXCEEDED") {
		t.Errorf("the ratchet verdict must not preempt the error cause:\n%s", out)
	}
}

// An error also gates the rebaseline path — the baseline records no errors, so
// a rebaseline must not read as a clean run while one stands.
func TestSpecLintUpdateBaseline_ErrorStillExitsOne(t *testing.T) {
	root := slNewCorpus(t)
	slInjectError(t, root)

	path := filepath.Join(t.TempDir(), "baseline.json")
	out, code := slRunLint(t, root, "--baseline", path, "--update-baseline", "--reason", "capture")
	if code != 1 {
		t.Fatalf("a standing error must exit 1 even on rebaseline, rc=%d\n%s", code, out)
	}
	// Guard against a wrong-reason pass: an argument error also exits non-zero.
	// Observed during RED — this test went green purely on "unknown flag".
	if strings.Contains(out, "unknown flag") {
		t.Fatalf("rc=1 came from an argument error, not from the standing error:\n%s", out)
	}
	if !strings.Contains(out, "MissingExclusions") {
		t.Errorf("the standing error must be visible in the output:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Flag contract + boundary cases — no silent fallback anywhere.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_FlagContractRejections(t *testing.T) {
	root := slNewCorpus(t)
	existing := slCaptureBaseline(t, root)
	missing := filepath.Join(t.TempDir(), "absent.json")

	cases := []struct {
		name     string
		args     []string
		wantWord string
	}{
		{"missing baseline file is an argument error, never a silent fallback",
			[]string{"--baseline", missing}, "absent.json"},
		{"--reason without --update-baseline",
			[]string{"--baseline", existing, "--reason", "x"}, "--update-baseline"},
		{"--update-baseline without --baseline",
			[]string{"--update-baseline", "--reason", "x"}, "--baseline"},
		{"--baseline with --json",
			[]string{"--baseline", existing, "--json"}, "--json"},
		{"--baseline with --sarif",
			[]string{"--baseline", existing, "--sarif"}, "--sarif"},
		{"--baseline with --strict",
			[]string{"--baseline", existing, "--strict"}, "--strict"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, code := slRunLint(t, root, tc.args...)
			if code != 3 {
				t.Fatalf("want exit 3, got %d\n%s", code, out)
			}
			if !strings.Contains(out, tc.wantWord) {
				t.Errorf("the message must name %q:\n%s", tc.wantWord, out)
			}
		})
	}
}

// Boundary: a rule that fires for the first time (absent from the baseline)
// reads as recorded 0 and is an increase.
func TestSpecLintBaseline_RuleAbsentFromBaselineIsAnIncrease(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	b := slReadBaseline(t, baseline)
	if b.Rules["SpecsDirMissingSpecFile"] != 0 {
		t.Fatalf("precondition: the rule must be ABSENT from this baseline, got %v", b.Rules)
	}

	slInjectWarning(t, root, "SPEC-FIRSTFIRE-001")
	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("a rule firing for the first time must be red, rc=%d\n%s", code, out)
	}
	if !strings.Contains(out, "recorded 0") {
		t.Errorf("an absent rule must read as recorded 0:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Backward compatibility — with no new flag, behaviour is unchanged.
// ---------------------------------------------------------------------------

func TestSpecLint_DefaultAndStrictBehaviourUnchanged(t *testing.T) {
	root := slNewCorpus(t)
	slInjectWarning(t, root, "SPEC-STANDING-001")

	defOut, defCode := slRunLint(t, root)
	if defCode != 0 {
		t.Errorf("default mode must not gate on warnings, rc=%d\n%s", defCode, defOut)
	}
	if !strings.Contains(defOut, "warning(s)") {
		t.Errorf("the default summary line must be preserved:\n%s", defOut)
	}

	strictOut, strictCode := slRunLint(t, root, "--strict")
	if strictCode != 1 {
		t.Errorf("--strict must still escalate a non-advisory warning, rc=%d\n%s", strictCode, strictOut)
	}

	jsonOut, jsonCode := slRunLint(t, root, "--json")
	if jsonCode != 0 {
		t.Errorf("--json must still exit 0 on a warning-only corpus, got %d\n%s", jsonCode, jsonOut)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-005 cross-process — the in-process tests above exercise the cobra
// command, but only a real binary invocation can catch a wiring defect between
// the process entry point, the flag, and the exit-code mapping.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_CrossProcessBothDirections(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-process test builds the CLI binary; skipped under -short")
	}
	bin := slBuildMoaiBinary(t)
	root := slNewCorpus(t)

	baselinePath := filepath.Join(t.TempDir(), "baseline.json")
	if capOut, capCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath, "--update-baseline", "--reason", "cross-process capture"); capCode != 0 {
		t.Fatalf("capture failed rc=%d\n%s", capCode, capOut)
	}

	if preOut, preCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath); preCode != 0 {
		t.Fatalf("precondition: unchanged corpus must be green, rc=%d\n%s", preCode, preOut)
	}

	dir := slInjectWarning(t, root, "SPEC-XPROC-001")
	redOut, redCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath)
	if redCode != 1 {
		t.Fatalf("cross-process: injected violation must exit 1, got %d\n%s", redCode, redOut)
	}
	if !strings.Contains(redOut, "SpecsDirMissingSpecFile") {
		t.Errorf("cross-process output must name the rule:\n%s", redOut)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove: %v", err)
	}
	cleanOut, cleanCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath)
	if cleanCode != 0 {
		t.Fatalf("cross-process: removal must return rc=0, got %d\n%s", cleanCode, cleanOut)
	}
}

func slBuildMoaiBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "moai-test-bin")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/modu-ai/moai-adk/cmd/moai")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build moai binary: %v\n%s", err, out)
	}
	return bin
}

func slRunBinary(t *testing.T, bin, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run %s: %v\n%s", bin, err, out)
	return "", -1
}

// slBaselineFile mirrors the on-disk record so the test decodes it without
// importing the producer's own marshaller — a serialization defect cannot then
// hide behind a symmetric bug in the reader.
type slBaselineFile struct {
	Version   int            `json:"version"`
	UpdatedAt string         `json:"updated_at"`
	TreeSHA   string         `json:"tree_sha"`
	Reason    string         `json:"reason"`
	Rules     map[string]int `json:"rules"`
}

func slReadBaseline(t *testing.T, path string) slBaselineFile {
	t.Helper()
	var b slBaselineFile
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("decode baseline: %v\n%s", err, data)
	}
	return b
}
