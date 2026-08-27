package cli

// codex_init_test.go — SPEC-CODEX-INIT-001 run-phase tests, gate side:
// AC-CI-001 (24), AC-CI-002 (32), AC-CI-003 (20), AC-CI-004 (20+40),
// AC-CI-010 (12) — 148 cells. Every subtest name carries its axis values
// (state / injected state / disk / verb / spawn / fixture) so the `-v`
// subtest count maps 1:1 onto the acceptance cell table (AC-CI-009).
//
// Disciplines observed here: RED-before-GREEN (the gate units were absent —
// acceptance R1-R3), no `-run` selector as a verdict basis, SNAP-based
// no-write proofs per definition 4, launch counting per definition 1 (both
// sites), and seam-only filesystem access on the judged paths.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── fixtures — wiring states (acceptance common fixture table) ────────────

type codexStateFixture string

const (
	stateNotWiredMissing codexStateFixture = "not_wired_missing" // S1
	stateNotWiredEmpty   codexStateFixture = "not_wired_empty"   // S2
	statePartialHooks    codexStateFixture = "partial_hooks"     // S3
	statePartialConfig   codexStateFixture = "partial_config"    // S4
	stateInvalid         codexStateFixture = "invalid_whitelist" // S5
	stateWired           codexStateFixture = "wired"             // S6
)

// codexAllIncompleteStates is S1..S5 in fixture-table order.
var codexAllIncompleteStates = []codexStateFixture{
	stateNotWiredMissing, stateNotWiredEmpty, statePartialHooks, statePartialConfig, stateInvalid,
}

func codexLayWiringState(t *testing.T, root string, st codexStateFixture) {
	t.Helper()
	lay := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	switch st {
	case stateNotWiredMissing:
		// no .codex at all
	case stateNotWiredEmpty:
		if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
			t.Fatal(err)
		}
	case statePartialHooks:
		lay(".codex/hooks.json", `{"hooks":{}}`)
	case statePartialConfig:
		lay(".codex/config.toml", "[mcp_servers.moai]\ncommand = \"moai\"\n")
	case stateInvalid:
		// parseable JSON whose top-level key fails the whitelist.
		lay(".codex/hooks.json", `{"bogus_top_level": true}`)
		lay(".codex/config.toml", "[mcp_servers.moai]\ncommand = \"moai\"\n")
	case stateWired:
		lay(".codex/hooks.json", `{"hooks":{}}`)
		lay(".codex/config.toml", "[mcp_servers.moai]\ncommand = \"moai\"\n")
	default:
		t.Fatalf("unknown wiring fixture %q", st)
	}
}

// codexFixtureStatus maps a fixture to the status the REAL classifier must
// return for it — the cells hold the disk and the classifier to this.
func codexFixtureStatus(st codexStateFixture) string {
	switch st {
	case stateNotWiredMissing, stateNotWiredEmpty:
		return codexWiringStatusNotWired
	case statePartialHooks, statePartialConfig:
		return codexWiringStatusPartial
	case stateInvalid:
		return codexWiringStatusInvalid
	default:
		return codexWiringStatusWired
	}
}

// ─── sandbox — one root under SNAP (definition 4) ──────────────────────────

func codexNewSandbox(t *testing.T) (sbx, proj string) {
	t.Helper()
	sbx = t.TempDir()
	proj = filepath.Join(sbx, "proj")
	for _, d := range []string{proj, filepath.Join(sbx, "home"), filepath.Join(sbx, "codex-home")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return sbx, proj
}

type codexSnapEntry struct {
	Mode os.FileMode
	SHA  string // regular files only
}

func codexSnapSandbox(t *testing.T, sbx string) map[string]codexSnapEntry {
	t.Helper()
	snap := map[string]codexSnapEntry{}
	err := filepath.WalkDir(sbx, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if path == sbx {
			return nil
		}
		rel, rerr := filepath.Rel(sbx, path)
		if rerr != nil {
			return rerr
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		entry := codexSnapEntry{Mode: info.Mode()}
		if info.Mode().IsRegular() {
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			sum := sha256.Sum256(data)
			entry.SHA = hex.EncodeToString(sum[:])
		}
		snap[filepath.ToSlash(rel)] = entry
		return nil
	})
	if err != nil {
		t.Fatalf("snap sandbox: %v", err)
	}
	return snap
}

// codexSnapOutsideProject keeps every entry NOT under proj/ — the parent
// directory, the fake home, the fake CODEX_HOME, and any escape target.
func codexSnapOutsideProject(snap map[string]codexSnapEntry) map[string]codexSnapEntry {
	out := map[string]codexSnapEntry{}
	for rel, e := range snap {
		if rel == "proj" || strings.HasPrefix(rel, "proj/") {
			continue
		}
		out[rel] = e
	}
	return out
}

func codexAssertSnapEqual(t *testing.T, label string, want, got map[string]codexSnapEntry) {
	t.Helper()
	for rel, w := range want {
		g, ok := got[rel]
		if !ok {
			t.Errorf("%s: entry %q vanished", label, rel)
			continue
		}
		if w.Mode != g.Mode || w.SHA != g.SHA {
			t.Errorf("%s: entry %q changed (mode %v→%v sha %q→%q)", label, rel, w.Mode, g.Mode, w.SHA, g.SHA)
		}
	}
	for rel := range got {
		if _, ok := want[rel]; !ok {
			t.Errorf("%s: entry %q appeared", label, rel)
		}
	}
}

// ─── filesystem seam recorder ──────────────────────────────────────────────

type codexFSCall struct {
	Kind  string // lstat | read | createtemp | rename | evalsymlinks
	Path  string
	Path2 string // rename destination
}

type codexFSRecorder struct {
	calls          []codexFSCall
	failCreateTemp error
}

// withCodexFSRecorder wraps ALL five contract filesystem seams with a
// recording delegate-to-real pair. This is the counting surface the
// acceptance's "filesystem seam calls" read: the gate section must total 0
// (AC-CI-001/002), reads toward a refused path must total 0 (AC-CI-011),
// and write-mode paths must never name a target or leave the project.
func withCodexFSRecorder(t *testing.T, failCreateTemp error) *codexFSRecorder {
	t.Helper()
	rec := &codexFSRecorder{failCreateTemp: failCreateTemp}
	prevLstat, prevRead, prevTemp, prevRename, prevEval := codexLstatFn, codexReadFileFn, codexCreateTempFn, codexRenameFn, codexEvalSymlinksFn
	codexLstatFn = func(path string) (os.FileInfo, error) {
		rec.calls = append(rec.calls, codexFSCall{Kind: "lstat", Path: path})
		return os.Lstat(path)
	}
	codexReadFileFn = func(path string) ([]byte, error) {
		rec.calls = append(rec.calls, codexFSCall{Kind: "read", Path: path})
		return os.ReadFile(path)
	}
	codexCreateTempFn = func(dir, pattern string) (*os.File, error) {
		rec.calls = append(rec.calls, codexFSCall{Kind: "createtemp", Path: filepath.Join(dir, pattern)})
		if rec.failCreateTemp != nil {
			return nil, rec.failCreateTemp
		}
		return os.CreateTemp(dir, pattern)
	}
	codexRenameFn = func(oldname, newname string) error {
		rec.calls = append(rec.calls, codexFSCall{Kind: "rename", Path: oldname, Path2: newname})
		return os.Rename(oldname, newname)
	}
	codexEvalSymlinksFn = func(path string) (string, error) {
		rec.calls = append(rec.calls, codexFSCall{Kind: "evalsymlinks", Path: path})
		return filepath.EvalSymlinks(path)
	}
	t.Cleanup(func() {
		codexLstatFn, codexReadFileFn, codexCreateTempFn, codexRenameFn, codexEvalSymlinksFn = prevLstat, prevRead, prevTemp, prevRename, prevEval
	})
	return rec
}

func (r *codexFSRecorder) total() int { return len(r.calls) }

// readsOf counts read-mode calls whose path contains substr.
func (r *codexFSRecorder) readsOf(substr string) int {
	n := 0
	for _, c := range r.calls {
		if c.Kind == "read" && strings.Contains(c.Path, substr) {
			n++
		}
	}
	return n
}

// writePaths lists every write-mode principal path (temp targets and rename
// destinations).
func (r *codexFSRecorder) writePaths() []string {
	var out []string
	for _, c := range r.calls {
		switch c.Kind {
		case "createtemp":
			out = append(out, c.Path)
		case "rename":
			out = append(out, c.Path2)
		}
	}
	return out
}

// ─── gate harness — every seam the gate can touch, stubbed at once ─────────

type codexGateHarness struct {
	launch *codexLaunchCapture
	fs     *codexFSRecorder
	order  []string // "generator" | "contract", in call order

	classifierCalls int
	injectedState   string // empty → count + delegate to the REAL classifier

	capable       bool
	capableCalls  int
	promptCalls   int
	promptAccepts bool

	genCalls int
	genReqs  []codexGeneratorRequest
	genErr   error

	contractCalls int
	contractReqs  []codexContractRequest
	contractReal  bool
}

type codexHarnessOpts struct {
	injectedState  string
	capable        bool
	promptAccepts  bool
	genErr         error
	failCreateTemp error
	contractReal   bool
}

func withCodexGateHarness(t *testing.T, projectRoot string, opts codexHarnessOpts) *codexGateHarness {
	t.Helper()
	h := &codexGateHarness{
		launch:        withCodexLaunchCapture(t),
		fs:            withCodexFSRecorder(t, opts.failCreateTemp),
		capable:       opts.capable,
		promptAccepts: opts.promptAccepts,
		genErr:        opts.genErr,
		injectedState: opts.injectedState,
		contractReal:  opts.contractReal,
	}

	prevClassifier := codexWiringClassifierFn
	codexWiringClassifierFn = func(root string) codexWiringInfo {
		h.classifierCalls++
		if h.injectedState != "" {
			return codexWiringInfo{Status: h.injectedState, Detail: "(injected)"}
		}
		return classifyCodexWiring(root)
	}
	t.Cleanup(func() { codexWiringClassifierFn = prevClassifier })

	prevCapable := codexPromptCapableFn
	codexPromptCapableFn = func() bool {
		h.capableCalls++
		return h.capable
	}
	t.Cleanup(func() { codexPromptCapableFn = prevCapable })

	prevPrompt := codexOfferPromptFn
	codexOfferPromptFn = func(out io.Writer, in io.Reader, info codexWiringInfo) bool {
		h.promptCalls++
		return h.promptAccepts
	}
	t.Cleanup(func() { codexOfferPromptFn = prevPrompt })

	prevGen := codexInitGeneratorFn
	codexInitGeneratorFn = func(req codexGeneratorRequest) error {
		h.genCalls++
		h.genReqs = append(h.genReqs, req)
		h.order = append(h.order, "generator")
		return h.genErr
	}
	t.Cleanup(func() { codexInitGeneratorFn = prevGen })

	prevContract := codexContractFn
	codexContractFn = func(req codexContractRequest) error {
		h.contractCalls++
		h.contractReqs = append(h.contractReqs, req)
		h.order = append(h.order, "contract")
		if h.contractReal {
			return secureCodexInstructionContract(req)
		}
		return nil
	}
	t.Cleanup(func() { codexContractFn = prevContract })

	withCodexProjectRoot(t, projectRoot)

	// --spawn cells cross checkSpawnPrereqs before the spawn seam — pin
	// in-tmux so the seam (not the precondition) bounds the cell.
	prevInTmux := inTmuxFn
	inTmuxFn = func() bool { return true }
	t.Cleanup(func() { inTmuxFn = prevInTmux })

	return h
}

// withCodexGateOpen pins the gate's classifier to wired: launcher-SPEC
// launch-mechanics tests exercise the launch path directly, past the offer.
func withCodexGateOpen(t *testing.T) {
	t.Helper()
	prev := codexWiringClassifierFn
	codexWiringClassifierFn = func(string) codexWiringInfo {
		return codexWiringInfo{Status: codexWiringStatusWired, Detail: "(gate pinned open)"}
	}
	t.Cleanup(func() { codexWiringClassifierFn = prev })
}

// runCodexInitCell runs one verb×spawn combination on a fresh command.
func runCodexInitCell(t *testing.T, verb string, spawn bool) (stdout, stderr string, exitCode int) {
	t.Helper()
	args := []string{verb}
	if spawn {
		args = append(args, "--spawn")
	}
	out, errOut, err := runCodexCmd(t, args...)
	return out, errOut, codexExitCodeOf(t, err)
}

func codexExitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var ec *exitCodeError
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	t.Fatalf("unexpected non-exit-code error: %v", err)
	return 0
}

// codexWantIncompleteGate holds the assertions shared by every cell whose
// gate outcome is a proposal (offer issued, nothing launched, no disk touch).
func codexWantIncompleteGate(t *testing.T, h *codexGateHarness, wantStatus string, stdout, stderr string, exitCode int) {
	t.Helper()
	if h.promptCalls != 1 {
		t.Errorf("prompt calls = %d, want 1", h.promptCalls)
	}
	if h.classifierCalls < 1 {
		t.Errorf("classifier calls = %d, want >= 1", h.classifierCalls)
	}
	if n := h.launch.count(); n != 0 {
		t.Errorf("launches (both sites) = %d, want 0", n)
	}
	if n := h.fs.total(); n != 0 {
		t.Errorf("gate-section filesystem seam calls = %d, want 0 (%v)", n, h.fs.calls)
	}
	if exitCode != codexInitDeclinedExitCode {
		t.Errorf("exit code = %d, want %d (cancel)", exitCode, codexInitDeclinedExitCode)
	}
	combined := stdout + stderr
	if !strings.Contains(combined, wantStatus) {
		t.Errorf("output does not name the wiring state %q: %q", wantStatus, combined)
	}
	if !strings.Contains(combined, codexWiringAction) {
		t.Errorf("output does not name the remedy %q: %q", codexWiringAction, combined)
	}
}

// codexWantLaunched holds the assertions shared by every wired cell.
func codexWantLaunched(t *testing.T, h *codexGateHarness, spawn bool, exitCode int) {
	t.Helper()
	if h.promptCalls != 0 {
		t.Errorf("prompt calls = %d, want 0 (wired launches without an offer)", h.promptCalls)
	}
	if h.classifierCalls < 1 {
		t.Errorf("classifier calls = %d, want >= 1", h.classifierCalls)
	}
	if h.genCalls != 0 {
		t.Errorf("generator calls = %d, want 0 on the wired path", h.genCalls)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	direct, spawnN := 1, 0
	if spawn {
		direct, spawnN = 0, 1
	}
	codexWantLaunches(t, h.launch, 1, direct, spawnN)
	if n := h.fs.total(); n != 0 {
		t.Errorf("gate-section filesystem seam calls = %d, want 0 (%v)", n, h.fs.calls)
	}
}

// ─── AC-CI-001 — state × verb × spawn matrix (24 cells) ────────────────────

// TestCodexInitGateStateMatrix runs the 6-state × 2-verb × 2-spawn
// crossproduct against the REAL classifier on laid-out disk states. The
// prompt seam is pinned to DECLINE — the accept path belongs to
// TestCodexInitAcceptDelegation.
func TestCodexInitGateStateMatrix(t *testing.T) {
	states := append([]codexStateFixture{stateWired}, codexAllIncompleteStates...)
	verbs := []string{"cli", "app"}
	type outcome struct {
		prompts, launches, classifier, fsCalls int
	}
	byState := map[codexStateFixture][]outcome{}
	for _, st := range states {
		for _, verb := range verbs {
			for _, spawn := range []bool{false, true} {
				name := fmt.Sprintf("state=%s/verb=%s/spawn=%t", st, verb, spawn)
				t.Run(name, func(t *testing.T) {
					_, proj := codexNewSandbox(t)
					codexLayWiringState(t, proj, st)
					h := withCodexGateHarness(t, proj, codexHarnessOpts{capable: true, promptAccepts: false})
					stdout, stderr, code := runCodexInitCell(t, verb, spawn)

					if st == stateWired {
						codexWantLaunched(t, h, spawn, code)
					} else {
						codexWantIncompleteGate(t, h, codexFixtureStatus(st), stdout, stderr, code)
					}

					// The judgement must come from the disk state through the
					// classifier — every cell observes the fixture's status.
					byState[st] = append(byState[st], outcome{
						prompts: h.promptCalls, launches: h.launch.count(),
						classifier: h.classifierCalls, fsCalls: h.fs.total(),
					})
				})
			}
		}
	}
	// All four verb×spawn combinations of one state behaved identically —
	// neither the verb nor --spawn may move the verdict.
	for st, outs := range byState {
		if len(outs) != 4 {
			t.Fatalf("state %s: %d outcomes, want 4", st, len(outs))
		}
		for i, o := range outs[1:] {
			if o != outs[0] {
				t.Errorf("state %s: outcome %d differs from the first (%+v vs %+v) — the verdict split by verb or spawn", st, i+1, o, outs[0])
			}
		}
	}
}

// ─── AC-CI-002 — injected-state contradiction matrix (32 cells) ────────────

// TestCodexInitGateInjectedState injects each of the four state tokens while
// the DISK carries a contradicting (or, for one pair per state, agreeing)
// fixture, across both verbs and both spawn modes. The gate must follow the
// returned state alone: injected≠wired proposes (launch 0) even on a fully
// wired disk; injected=wired launches even on a bare disk.
func TestCodexInitGateInjectedState(t *testing.T) {
	type pair struct {
		injected string
		disk     codexStateFixture
	}
	pairs := []pair{
		{codexWiringStatusNotWired, stateWired},
		{codexWiringStatusPartial, stateWired},
		{codexWiringStatusInvalid, stateWired},
		{codexWiringStatusWired, stateNotWiredMissing},
		{codexWiringStatusNotWired, stateNotWiredEmpty},
		{codexWiringStatusPartial, stateNotWiredMissing},
		{codexWiringStatusInvalid, stateNotWiredMissing},
		{codexWiringStatusWired, stateNotWiredEmpty},
	}
	verbs := []string{"cli", "app"}
	for _, p := range pairs {
		for _, verb := range verbs {
			for _, spawn := range []bool{false, true} {
				name := fmt.Sprintf("injected=%s/disk=%s/verb=%s/spawn=%t",
					strings.ReplaceAll(p.injected, " ", "_"), p.disk, verb, spawn)
				t.Run(name, func(t *testing.T) {
					_, proj := codexNewSandbox(t)
					codexLayWiringState(t, proj, p.disk)
					h := withCodexGateHarness(t, proj, codexHarnessOpts{
						injectedState: p.injected, capable: true, promptAccepts: false,
					})
					_, _, code := runCodexInitCell(t, verb, spawn)

					if h.classifierCalls < 1 {
						t.Errorf("classifier calls = %d, want >= 1", h.classifierCalls)
					}
					if p.injected == codexWiringStatusWired {
						// launched, no proposal — even though the disk below
						// is bare/incomplete.
						codexWantLaunched(t, h, spawn, code)
						return
					}
					// proposal, no launch — even though disk may be fully
					// wired.
					if h.promptCalls != 1 {
						t.Errorf("prompt calls = %d, want 1", h.promptCalls)
					}
					if n := h.launch.count(); n != 0 {
						t.Errorf("launches = %d, want 0", n)
					}
					if exitCode := code; exitCode != codexInitDeclinedExitCode {
						t.Errorf("exit code = %d, want %d", exitCode, codexInitDeclinedExitCode)
					}
					if n := h.fs.total(); n != 0 {
						t.Errorf("gate-section filesystem seam calls = %d, want 0 (%v)", n, h.fs.calls)
					}
				})
			}
		}
	}
}

// ─── AC-CI-003 — decline path (20 cells) ───────────────────────────────────

// TestCodexInitDecline: S1..S5 × 2 verbs × 2 spawn, operator declines.
// Nothing may be written (whole-sandbox SNAP equality — not just the
// project), nothing launched, generator never called, cancel exit code.
func TestCodexInitDecline(t *testing.T) {
	verbs := []string{"cli", "app"}
	for _, st := range codexAllIncompleteStates {
		for _, verb := range verbs {
			for _, spawn := range []bool{false, true} {
				name := fmt.Sprintf("state=%s/verb=%s/spawn=%t", st, verb, spawn)
				t.Run(name, func(t *testing.T) {
					sbx, proj := codexNewSandbox(t)
					codexLayWiringState(t, proj, st)
					before := codexSnapSandbox(t, sbx)
					h := withCodexGateHarness(t, proj, codexHarnessOpts{capable: true, promptAccepts: false})
					stdout, stderr, code := runCodexInitCell(t, verb, spawn)

					if h.genCalls != 0 {
						t.Errorf("generator calls = %d, want 0 on decline", h.genCalls)
					}
					if h.contractCalls != 0 {
						t.Errorf("contract calls = %d, want 0 on decline", h.contractCalls)
					}
					if n := h.launch.count(); n != 0 {
						t.Errorf("launches = %d, want 0", n)
					}
					if exitCode := code; exitCode != codexInitDeclinedExitCode {
						t.Errorf("exit code = %d, want %d (cancel, not error)", exitCode, codexInitDeclinedExitCode)
					}
					codexAssertSnapEqual(t, "decline sandbox", before, codexSnapSandbox(t, sbx))
					for _, p := range h.fs.writePaths() {
						if !strings.HasPrefix(p, proj+string(filepath.Separator)) {
							t.Errorf("write-mode seam call outside the project: %q", p)
						}
					}
					combined := stdout + stderr
					if !strings.Contains(combined, codexFixtureStatus(st)) || !strings.Contains(combined, codexWiringAction) {
						t.Errorf("output must name the state and the remedy: %q", combined)
					}
				})
			}
		}
	}
}

// ─── AC-CI-004 (accept) — generator delegation, 20 cells ───────────────────

// TestCodexInitAcceptDelegation: S1..S5 × 2 verbs × 2 spawn, operator
// accepts. The generator must run EXACTLY once with exactly the codex agent
// selection; the contract runs once, AFTER the generator, and really writes
// the instruction files; the launch then happens exactly once. The spawn
// pair of each (state, verb) must capture identical contract arguments and
// produce identical instruction files — the contract step is not a function
// of --spawn.
func TestCodexInitAcceptDelegation(t *testing.T) {
	verbs := []string{"cli", "app"}
	type cellResult struct {
		contractRootRel string // captured ProjectRoot with the per-cell sandbox prefix stripped
		contractCalls   int
		agents, claude  []byte
	}
	byPair := map[string][]cellResult{}
	for _, st := range codexAllIncompleteStates {
		for _, verb := range verbs {
			for _, spawn := range []bool{false, true} {
				name := fmt.Sprintf("state=%s/verb=%s/spawn=%t", st, verb, spawn)
				t.Run(name, func(t *testing.T) {
					sbx, proj := codexNewSandbox(t)
					codexLayWiringState(t, proj, st)
					before := codexSnapSandbox(t, sbx)
					h := withCodexGateHarness(t, proj, codexHarnessOpts{
						capable: true, promptAccepts: true, contractReal: true,
					})
					_, _, code := runCodexInitCell(t, verb, spawn)

					// generator: exactly once, exactly the codex selection.
					if h.genCalls != 1 {
						t.Errorf("generator calls = %d, want 1", h.genCalls)
					}
					if h.genCalls == 1 {
						req := h.genReqs[0]
						if req.Agent != codexGeneratorAgentCodex {
							t.Errorf("generator agent = %q, want exactly %q", req.Agent, codexGeneratorAgentCodex)
						}
						if req.ProjectRoot != proj {
							t.Errorf("generator project root = %q, want %q", req.ProjectRoot, proj)
						}
					}
					// contract: exactly once, AFTER the generator.
					if h.contractCalls != 1 {
						t.Errorf("contract calls = %d, want 1", h.contractCalls)
					}
					if len(h.order) != 2 || h.order[0] != "generator" || h.order[1] != "contract" {
						t.Errorf("call order = %v, want [generator contract]", h.order)
					}
					// disk: the instruction files really exist, with exactly
					// one executing import.
					agentsPath := filepath.Join(proj, codexAgentsRelPath)
					claudePath := filepath.Join(proj, codexClaudeRelPath)
					agents, aerr := os.ReadFile(agentsPath)
					if aerr != nil {
						t.Fatalf("AGENTS.md missing after acceptance: %v", aerr)
					}
					claude, cerr := os.ReadFile(claudePath)
					if cerr != nil {
						t.Fatalf("CLAUDE.md missing after acceptance: %v", cerr)
					}
					if got := codexTestExecImports(t, claudePath, codexLinkAgentsDirective); got != 1 {
						t.Errorf("executing @AGENTS.md imports in CLAUDE.md = %d, want 1", got)
					}
					if got := codexTestExecImports(t, agentsPath, codexLinkLocalDirective); got != 0 {
						t.Errorf("executing @CLAUDE.local.md imports in AGENTS.md = %d, want 0 (no local file exists)", got)
					}
					// launch: exactly once, on the requested site.
					direct, spawnN := 1, 0
					if spawn {
						direct, spawnN = 0, 1
					}
					codexWantLaunches(t, h.launch, 1, direct, spawnN)
					// no wiring write from THIS SPEC's code paths: no
					// write-mode seam call INSIDE the .codex directory, none
					// outside proj.
					codexDirPrefix := filepath.Join(proj, ".codex") + string(filepath.Separator)
					for _, p := range h.fs.writePaths() {
						if strings.HasPrefix(p, codexDirPrefix) {
							t.Errorf("write-mode seam call under .codex/: %q", p)
						}
						if !strings.HasPrefix(p, proj+string(filepath.Separator)) {
							t.Errorf("write-mode seam call outside the project: %q", p)
						}
					}
					codexAssertSnapEqual(t, "outside-project snapshot", codexSnapOutsideProject(before), codexSnapOutsideProject(codexSnapSandbox(t, sbx)))
					if code != 0 {
						t.Errorf("exit code = %d, want 0 on the accepted path", code)
					}

					key := fmt.Sprintf("%s/%s", st, verb)
					byPair[key] = append(byPair[key], cellResult{
						contractRootRel: strings.TrimPrefix(h.contractReqs[0].ProjectRoot, sbx+string(filepath.Separator)),
						contractCalls:   h.contractCalls,
						agents:          agents, claude: claude,
					})
				})
			}
		}
	}
	// Pair comparison: same state+verb, spawn off vs on — identical captured
	// contract arguments (compared relative to each cell's own sandbox),
	// identical call counts, identical file bytes.
	for key, results := range byPair {
		if len(results) != 2 {
			t.Fatalf("pair %s: %d results, want 2", key, len(results))
		}
		a, b := results[0], results[1]
		if a.contractRootRel != b.contractRootRel {
			t.Errorf("pair %s: contract ProjectRoot differs by spawn (%q vs %q)", key, a.contractRootRel, b.contractRootRel)
		}
		if a.contractCalls != 1 || b.contractCalls != 1 {
			t.Errorf("pair %s: contract calls differ by spawn (%d vs %d)", key, a.contractCalls, b.contractCalls)
		}
		if string(a.agents) != string(b.agents) || string(a.claude) != string(b.claude) {
			t.Errorf("pair %s: instruction file bytes differ by spawn — the contract read the flag", key)
		}
	}
}

// ─── AC-CI-004 (prompt) — non-interactive refusal, 40 cells ────────────────

// TestCodexInitPromptIssuance: S1..S5 × 2 verbs × 2 spawn × {capable,
// incapable}. The decision source is the INJECTED capability alone — the
// capable-20 cells run with stdin as a pipe yet must still issue exactly one
// prompt, which is what kills a stdin-kind gate.
func TestCodexInitPromptIssuance(t *testing.T) {
	verbs := []string{"cli", "app"}
	for _, capable := range []bool{false, true} {
		for _, st := range codexAllIncompleteStates {
			for _, verb := range verbs {
				for _, spawn := range []bool{false, true} {
					name := fmt.Sprintf("capable=%t/state=%s/verb=%s/spawn=%t", capable, st, verb, spawn)
					t.Run(name, func(t *testing.T) {
						sbx, proj := codexNewSandbox(t)
						codexLayWiringState(t, proj, st)
						before := codexSnapSandbox(t, sbx)
						h := withCodexGateHarness(t, proj, codexHarnessOpts{capable: capable, promptAccepts: false})
						stdout, stderr, code := runCodexInitCell(t, verb, spawn)

						if h.genCalls != 0 {
							t.Errorf("generator calls = %d, want 0 in the prompt cells", h.genCalls)
						}
						if n := h.launch.count(); n != 0 {
							t.Errorf("launches = %d, want 0", n)
						}
						if !capable {
							if h.promptCalls != 0 {
								t.Errorf("prompt calls = %d, want 0 — an unanswerable prompt must never be issued", h.promptCalls)
							}
							if exitCode := code; exitCode != codexInitFailureExitCode {
								t.Errorf("exit code = %d, want %d (non-success)", exitCode, codexInitFailureExitCode)
							}
							codexAssertSnapEqual(t, "non-interactive sandbox", before, codexSnapSandbox(t, sbx))
							combined := stdout + stderr
							if !strings.Contains(combined, codexFixtureStatus(st)) || !strings.Contains(combined, codexWiringAction) {
								t.Errorf("output must name the state and the remedy: %q", combined)
							}
							return
						}
						// capable: exactly one prompt (declining), despite the
						// pipe stdin — proves the decision came from the seam.
						if h.promptCalls != 1 {
							t.Errorf("prompt calls = %d, want 1", h.promptCalls)
						}
						if exitCode := code; exitCode != codexInitDeclinedExitCode {
							t.Errorf("exit code = %d, want %d", exitCode, codexInitDeclinedExitCode)
						}
					})
				}
			}
		}
	}
}

// ─── AC-CI-010 — failure paths (12 cells) ──────────────────────────────────

// codexLayUserInstructionFiles writes the three instruction files with user
// content and NO links — the E-fixture precondition (existing content is
// what makes preservation observable).
func codexLayUserInstructionFiles(t *testing.T, proj string) (agents, claude, local []byte) {
	t.Helper()
	agents = []byte("# My agents notes\n\ncustom instruction prose\n")
	claude = []byte("# My claude notes\n\nmore prose\n")
	local = []byte("local secrets — uncommitted guidance\n")
	for name, content := range map[string][]byte{
		codexAgentsRelPath: agents, codexClaudeRelPath: claude, codexLocalInstructionName: local,
	} {
		if err := os.WriteFile(filepath.Join(proj, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return agents, claude, local
}

// TestCodexInitFailurePaths: E1 (generator fails), E2 (contract write
// fails), E3 (generator fails after partial output) × 2 verbs × 2 spawn.
// A launch after any failure delivers exactly the unwired session this SPEC
// exists to prevent — so every cell launches NOTHING, exits non-success,
// names the failure point, preserves all three instruction files
// byte-for-byte, and never calls the contract after a failed generator.
func TestCodexInitFailurePaths(t *testing.T) {
	verbs := []string{"cli", "app"}
	type failureFixture struct {
		name     string
		opts     func(t *testing.T) codexHarnessOpts
		layDisk  func(t *testing.T, proj string)
		wantCall string // the underlying failure text the output must name
	}
	fixtures := []failureFixture{
		{
			name: "e1_generator_fails",
			opts: func(t *testing.T) codexHarnessOpts {
				return codexHarnessOpts{capable: true, promptAccepts: true, genErr: errors.New("generator boom (e1)")}
			},
			layDisk:  func(t *testing.T, proj string) {},
			wantCall: "generator boom (e1)",
		},
		{
			name: "e2_contract_write_fails",
			opts: func(t *testing.T) codexHarnessOpts {
				return codexHarnessOpts{
					capable: true, promptAccepts: true,
					failCreateTemp: errors.New("read-only target directory (e2)"), contractReal: true,
				}
			},
			layDisk:  func(t *testing.T, proj string) {},
			wantCall: "stage AGENTS.md",
		},
		{
			name: "e3_partial_output_then_fails",
			opts: func(t *testing.T) codexHarnessOpts {
				return codexHarnessOpts{capable: true, promptAccepts: true, genErr: errors.New("generator stopped midway (e3)")}
			},
			layDisk: func(t *testing.T, proj string) {
				// the generator's partial produce: config.toml landed, hooks
				// did not — a state that must NOT read as wired.
				if err := os.MkdirAll(filepath.Join(proj, ".codex"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(proj, ".codex/config.toml"), []byte("[mcp_servers.moai]\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantCall: "generator stopped midway (e3)",
		},
	}
	for _, fx := range fixtures {
		for _, verb := range verbs {
			for _, spawn := range []bool{false, true} {
				name := fmt.Sprintf("fixture=%s/verb=%s/spawn=%t", fx.name, verb, spawn)
				t.Run(name, func(t *testing.T) {
					sbx, proj := codexNewSandbox(t)
					codexLayWiringState(t, proj, stateNotWiredMissing)
					fx.layDisk(t, proj)
					agents, claude, local := codexLayUserInstructionFiles(t, proj)
					before := codexSnapSandbox(t, sbx)
					h := withCodexGateHarness(t, proj, fx.opts(t))
					stdout, stderr, code := runCodexInitCell(t, verb, spawn)

					if n := h.launch.count(); n != 0 {
						t.Errorf("launches = %d, want 0 after a failed initialization", n)
					}
					if exitCode := code; exitCode != codexInitFailureExitCode {
						t.Errorf("exit code = %d, want %d (non-success)", exitCode, codexInitFailureExitCode)
					}
					// all three instruction files byte-identical: a contract
					// that truncates before failing would zero them here.
					for name, want := range map[string][]byte{
						codexAgentsRelPath: agents, codexClaudeRelPath: claude, codexLocalInstructionName: local,
					} {
						got, err := os.ReadFile(filepath.Join(proj, name))
						if err != nil {
							t.Fatalf("read %s after failure: %v", name, err)
						}
						if string(got) != string(want) {
							t.Errorf("%s changed across the failed initialization (%d → %d bytes)", name, len(want), len(got))
						}
					}
					// a failed generator never reaches the contract step.
					if fx.name != "e2_contract_write_fails" && h.contractCalls != 0 {
						t.Errorf("contract calls = %d, want 0 after a generator failure", h.contractCalls)
					}
					// no write-mode call ever named an instruction target.
					for _, p := range h.fs.writePaths() {
						if strings.HasSuffix(p, codexAgentsRelPath) || strings.HasSuffix(p, codexClaudeRelPath) || strings.HasSuffix(p, codexLocalInstructionName) {
							t.Errorf("write-mode seam call named an instruction target: %q", p)
						}
						if !strings.HasPrefix(p, proj+string(filepath.Separator)) {
							t.Errorf("write-mode seam call outside the project: %q", p)
						}
					}
					codexAssertSnapEqual(t, "outside-project snapshot", codexSnapOutsideProject(before), codexSnapOutsideProject(codexSnapSandbox(t, sbx)))
					combined := stdout + stderr
					if !strings.Contains(combined, "codex init failed") || !strings.Contains(combined, fx.wantCall) {
						t.Errorf("output must name the failure point (%q): %q", fx.wantCall, combined)
					}
					if !strings.Contains(combined, codexWiringAction) {
						t.Errorf("output must name the remedy: %q", combined)
					}
				})
			}
		}
	}
}

// ─── default seam implementations ──────────────────────────────────────────

// TestCodexGateDefaultSeams exercises the gate's DEFAULT seam bodies, which
// every counted cell stubs out: the prompt accepts only y/yes, the generator
// default really delegates to codexwiring.Wire, and the capability probe
// runs without touching global state. Deliberately subtest-free — the 354
// cell count maps onto axis-named subtests only.
func TestCodexGateDefaultSeams(t *testing.T) {
	var out bytes.Buffer
	if !defaultCodexOfferPrompt(&out, strings.NewReader("y\n"), codexWiringInfo{Status: codexWiringStatusNotWired}) {
		t.Errorf("prompt declined a `y` answer")
	}
	if !strings.Contains(out.String(), codexWiringAction) {
		t.Errorf("offer text does not name the remedy: %q", out.String())
	}
	out.Reset()
	if defaultCodexOfferPrompt(&out, strings.NewReader("n\n"), codexWiringInfo{}) {
		t.Errorf("prompt accepted an `n` answer")
	}
	out.Reset()
	if defaultCodexOfferPrompt(&out, strings.NewReader(""), codexWiringInfo{}) {
		t.Errorf("prompt accepted EOF")
	}

	proj := t.TempDir()
	if err := defaultCodexInitGenerator(codexGeneratorRequest{ProjectRoot: proj, Agent: codexGeneratorAgentCodex}); err != nil {
		t.Fatalf("default generator failed: %v", err)
	}
	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml"} {
		if _, err := os.Stat(filepath.Join(proj, rel)); err != nil {
			t.Errorf("default generator did not produce %s: %v", rel, err)
		}
	}
	_ = defaultCodexPromptCapable() // must not panic; value is environment-dependent
}

// codexTestExecImports counts executing `@directive` lines per acceptance
// definition 5, implemented independently from the production scanner so the
// two implementations cross-check each other.
func codexTestExecImports(t *testing.T, path, directive string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	fence, comment, n := false, false, 0
	for _, ln := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(ln, "```") || strings.HasPrefix(ln, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		body := strings.TrimRight(ln, "\r")
		if comment {
			if i := strings.Index(body, "-->"); i >= 0 {
				comment = false
				body = body[i+3:]
			} else {
				continue
			}
		}
		if i := strings.Index(body, "<!--"); i >= 0 {
			if j := strings.Index(body[i:], "-->"); j >= 0 {
				body = body[i+j+3:]
			} else {
				comment = true
				continue
			}
		}
		if strings.HasPrefix(body, ">") {
			continue
		}
		if strings.HasPrefix(body, "@") && strings.TrimRight(body, " \t") == directive {
			n++
		}
	}
	return n
}

// codexTestRawOccurrences counts raw substring occurrences (the I6/I7
// companion count).
func codexTestRawOccurrences(t *testing.T, path, needle string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Count(string(data), needle)
}
