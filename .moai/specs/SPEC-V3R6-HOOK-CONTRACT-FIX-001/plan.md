# SPEC-V3R6-HOOK-CONTRACT-FIX-001 — Implementation Plan

## 1. Scope Recap (LEAN Tier S)

This SPEC is **Tier S** per the LEAN classification (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier):

- LOC scope: ~120 LOC across in-scope packages (2 new test functions ~80 LOC + 1 source edit ~5 LOC + working-tree cleanup ~0 LOC + optional doc polish ~30 LOC)
- Files affected: 3-5 (within the Tier S < 5 budget): `internal/cli/hook_test.go` (or new file), `internal/template/settings_test.go` (or new file), `internal/hook/subagent_stop.go`, `internal/hook/subagent_stop_test.go`, and the working-tree cleanup target `internal/hook/.moai/`
- Artifact set: spec.md + plan.md + acceptance.md (3 files per Tier S OR M; this SPEC uses 3-file Tier S form for clarity since AC count is 14 and benefits from a separate acceptance.md)
- plan-auditor PASS threshold: 0.75
- Section A-E delegation template: OPTIONAL (Tier S) — this SPEC uses a minimal in-line delegation if run-phase chooses Agent delegation; orchestrator-direct execution is also acceptable

## 2. Implementation Approach

### 2.1 Strategy: Defensive Consolidation

PR #1044 (commit `1b376eaef`) already applied the active-creator contract fix. This SPEC's role is **defensive consolidation**: lock in the contract with executable tests, fix the related working-tree leak at source, and verify the documentation crosswalk.

The work decomposes into 5 milestones, each independently verifiable, executable in a single run-phase session:

| Milestone | Description | Files touched | LOC | Verifies |
|-----------|-------------|---------------|-----|----------|
| M1 | Plain-text stdout regression test | `internal/cli/hook_test.go` (new file or extension) | ~50 | REQ-HCF-001, REQ-HCF-004, AC-HCF-001/002/004 |
| M2 | settings.json no-WorktreeCreate guard | `internal/template/settings_test.go` (new file or extension) | ~30 | REQ-HCF-003, AC-HCF-003 |
| M3 | subagent_stop.go CLAUDE_PROJECT_DIR fallback | `internal/hook/subagent_stop.go` + `subagent_stop_test.go` | ~5 + ~30 | REQ-HCF-005, AC-HCF-005 |
| M4 | Working-tree cleanup | `internal/hook/.moai/` removal | 0 | REQ-HCF-006, AC-HCF-006 |
| M5 | Documentation crosswalk verification (no edits unless drift) | Inspection only across 8+ files | 0-30 (drift fix-up only) | REQ-HCF-008/009, AC-HCF-008/009 |

### 2.2 Execution order

Milestones can execute in parallel where safe:
- M1 + M2 + M3 + M5 can run in any order (no file overlap)
- M4 (cleanup) should run AFTER M1/M3 so that the test runs do not accidentally re-create the leak (REQ-HCF-005 fix in M3 makes new leaks impossible from `subagent_stop.go`, but existing tests that haven't been re-checked could still create one)

Recommended order: **M3 → M1 → M2 → M5 → M4 (last)**. This way M4 is the final, idempotent commit.

### 2.3 Atomic commit strategy

3-4 commits recommended:
- Commit 1: M3 (source fix) — single Go source change, easy to revert
- Commit 2: M1 + M2 (new regression tests) — adds tests, no behavior change
- Commit 3: M4 (working-tree cleanup) — pure deletion, no source change
- Commit 4 (only if M5 found drift): documentation fix-up

Each commit follows Conventional Commits + `🗿 MoAI <email@mo.ai.kr>` trailer.

## 3. Detailed Milestone Plan

### M1 — Plain-text stdout regression test

**Goal**: Add executable spec for the plain-text stdout contract in `writeHookOutput()`.

**Target file**: `internal/cli/hook_test.go` (existing) OR new file `internal/cli/hook_writehookoutput_test.go`. Choice depends on existing test file size; if `hook_test.go` already exists and is small, extend; otherwise create new.

**Implementation sketch**:

The test captures stdout via `os.Pipe()` redirection inside the test scope. Standard Go pattern:

1. Save original `os.Stdout`
2. Create pipe (`os.Pipe()`)
3. Replace `os.Stdout` with pipe writer
4. Invoke `writeHookOutput(event, input, output)` (function under test)
5. Close writer side, read from reader side into a buffer
6. Restore `os.Stdout`
7. Assert buffer contents

Three subtests required:
- `t.Run("WorktreeCreate emits plain-text path", ...)`: event=WorktreeCreate, WorktreePath=`/test/path` → expect `/test/path\n`
- `t.Run("WorktreeRemove emits plain-text path", ...)`: event=WorktreeRemove, WorktreePath=`/test/path` → expect `/test/path\n`
- `t.Run("Empty WorktreePath produces empty stdout", ...)`: two sub-cases (Create + Remove), both empty WorktreePath → expect `""` exactly

**Risk**: If `writeHookOutput()` is unexported (currently lowercase `w`), the test MUST live in the same package (`package cli` not `cli_test`). Verify package declaration.

**Alternative implementation**: Refactor `writeHookOutput()` to accept an `io.Writer` parameter so the test can pass `&bytes.Buffer{}` directly. This is cleaner but expands scope. Recommend the os.Pipe approach (no source change).

**Acceptance**: AC-HCF-001, AC-HCF-002, AC-HCF-004 all PASS.

### M2 — settings.json no-WorktreeCreate guard

**Goal**: CI guard test fails if `WorktreeCreate` or `WorktreeRemove` re-appears as a top-level hooks key.

**Target file**: `internal/template/settings_test.go` (existing) OR new file `internal/template/settings_no_worktree_keys_test.go`.

**Implementation sketch**:

1. Read both files into memory:
   - `.claude/settings.json` (local) — resolve via `filepath.Join(repoRoot, ".claude/settings.json")`. The repoRoot helper likely exists in `settings_test.go` already; reuse.
   - `internal/template/templates/.claude/settings.json.tmpl` (template) — same.
2. Parse each as JSON (or scan for the literal pattern with a regex anchored at the `hooks` object key).
3. Assert `result.Hooks["WorktreeCreate"]` is absent AND `result.Hooks["WorktreeRemove"]` is absent.
4. Fail with a descriptive error if either is found.

**Subtle issue**: The `.json.tmpl` file is a Go template with template-language tokens (`{{...}}`). Direct JSON parsing may fail. Two options:
- (a) Render the template first (with stub TemplateContext), then parse JSON
- (b) Use a regex/grep approach: search for the literal string `"WorktreeCreate"` followed by `:` (with optional whitespace) at any line — simpler, more brittle to comments

Recommended: **option (b)** with a comment-stripping pre-pass. Test does:
```
content := readFile(...)
// Strip JSON-style line comments (template may have //... lines)
stripped := commentStripperRegex.ReplaceAllString(content, "")
if regexp.MustCompile(`"WorktreeCreate"\s*:`).MatchString(stripped) {
    t.Fatal("WorktreeCreate hook re-registered in settings.json - violates REQ-HCF-003")
}
```

**Acceptance**: AC-HCF-003 PASS.

### M3 — subagent_stop.go CLAUDE_PROJECT_DIR fallback

**Goal**: Insert `$CLAUDE_PROJECT_DIR` lookup between `input.CWD` check and `os.Getwd()` fallback in `dispatchCapture()`.

**Target file**: `internal/hook/subagent_stop.go` lines 204-210.

**Current code** (subagent_stop.go:204-210):
```go
func (h *subagentStopHandler) dispatchCapture(input *HookInput) {
	// Determine observations path relative to project CWD.
	projectDir := input.CWD
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	obsPath := filepath.Join(projectDir, ".moai", "harness", "observations.yaml")
```

**Target code**:
```go
func (h *subagentStopHandler) dispatchCapture(input *HookInput) {
	// Determine observations path. Resolution order: input.CWD → $CLAUDE_PROJECT_DIR → os.Getwd().
	// REQ-HCF-005: env var lookup prevents the internal/hook/.moai/ working-tree leak
	// observed when unit tests omit input.CWD; os.Getwd() then returns the test's package dir.
	projectDir := input.CWD
	if projectDir == "" {
		projectDir = os.Getenv("CLAUDE_PROJECT_DIR")
	}
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	obsPath := filepath.Join(projectDir, ".moai", "harness", "observations.yaml")
```

**Test file**: `internal/hook/subagent_stop_test.go` — add new test `TestDispatchCapture_UsesClaudeProjectDirWhenCwdEmpty`.

**Test sketch**:

```go
func TestDispatchCapture_UsesClaudeProjectDirWhenCwdEmpty(t *testing.T) {
    // Set the env var to a known value
    t.Setenv("CLAUDE_PROJECT_DIR", "/test/project")

    // Construct input with empty CWD
    input := &HookInput{CWD: "", AgentName: "test-agent", SessionID: "sess-test"}

    // Invoke dispatchCapture; verify the constructed ObservationsPath used the env var.
    // Since dispatchCapture is unexported and the capturer.New(...) is called inline,
    // we need a way to observe the path.

    // Approach 1: Inject capture.New via a package-level var that the test can override.
    // Approach 2: Test via the side effect — call dispatchCapture, then assert that
    //   the file at "/test/project/.moai/harness/observations.yaml" was attempted to be created
    //   (or that no file appeared under the test's t.TempDir()).
    // Approach 3: Refactor dispatchCapture into two functions: pathResolver(input) and emitObservation(path).
    //   Test pathResolver in isolation.

    // Approach 3 is cleanest. Plan-phase recommends approach 3; final decision is run-phase.

    // Pseudo-code with approach 3:
    got := resolveObservationsPath(input)
    want := filepath.Join("/test/project", ".moai", "harness", "observations.yaml")
    if got != want {
        t.Errorf("resolveObservationsPath() = %q, want %q", got, want)
    }
}
```

**Refactor sketch (approach 3)**:

```go
// New helper, package-private:
func resolveObservationsPath(input *HookInput) string {
    projectDir := input.CWD
    if projectDir == "" {
        projectDir = os.Getenv("CLAUDE_PROJECT_DIR")
    }
    if projectDir == "" {
        projectDir, _ = os.Getwd()
    }
    return filepath.Join(projectDir, ".moai", "harness", "observations.yaml")
}

// dispatchCapture becomes:
func (h *subagentStopHandler) dispatchCapture(input *HookInput) {
    obsPath := resolveObservationsPath(input)
    capturer := capture.New(capture.Config{ObservationsPath: obsPath})
    // ... rest unchanged
}
```

This refactor is in scope per Tier S (small, well-contained). The new helper is testable in isolation without mocking.

**Acceptance**: AC-HCF-005 PASS.

### M4 — Working-tree cleanup

**Goal**: Remove the leaked `internal/hook/.moai/` directory and confirm idempotency.

**Pre-flight**:
```bash
# Confirm the dir is untracked (not in git history)
git ls-files internal/hook/.moai/ | wc -l
# Expected: 0
```

**Execution**:
```bash
rm -rf internal/hook/.moai/
```

**Post-condition verification**:
```bash
test ! -e internal/hook/.moai/   # exit 0
git status --porcelain internal/hook/.moai/   # empty output
```

**Acceptance**: AC-HCF-006 PASS.

### M5 — Documentation crosswalk verification

**Goal**: Verify the post-PR-#1044 documentation state is consistent across 8+ files.

**Inspection points**:

1. `.claude/rules/moai/core/hooks-system.md` — search for §WorktreeCreate and WorktreeRemove Hooks section
2. `internal/template/templates/.claude/rules/moai/core/hooks-system.md` (template) — same
3. `.claude/rules/moai/workflow/worktree-integration.md` — same
4. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` (template) — same
5. `docs-site/content/ko/advanced/hooks-reference.md` — 20-row handler table
6. `docs-site/content/en/advanced/hooks-reference.md` — same
7. `docs-site/content/ja/advanced/hooks-reference.md` — same
8. `docs-site/content/zh/advanced/hooks-reference.md` — same
9. `internal/template/settings_test.go:512` — `const expectedCount = 20`

**Method**: Run the AC-HCF-008 verification commands. Capture stdout for each.

**Edge case — drift found**: If any inspection point shows drift (e.g., a locale has 21 rows by mistake), the fix-up edit is in scope as a Commit 4. The drift is added to progress.md.

**Edge case — no drift**: All 8+ files are already consistent post-PR-#1044. AC-HCF-008 is a pass-by-inspection with the verification commands as evidence.

**Acceptance**: AC-HCF-008, AC-HCF-009 PASS.

## 4. Traceability Matrix (REQ ↔ AC)

| REQ | Description | Mapped ACs |
|-----|-------------|------------|
| REQ-HCF-001 | writeHookOutput plain-text contract | AC-HCF-001, AC-HCF-002, AC-HCF-004 |
| REQ-HCF-002 | Handle() returns empty HookOutput | AC-HCF-014 |
| REQ-HCF-003 | settings.json no-WorktreeCreate | AC-HCF-003 |
| REQ-HCF-004 | Plain-text stdout regression test exists | AC-HCF-001, AC-HCF-002, AC-HCF-004 |
| REQ-HCF-005 | CLAUDE_PROJECT_DIR resolution order | AC-HCF-005 |
| REQ-HCF-006 | observations.yaml leak removed | AC-HCF-006 |
| REQ-HCF-007 | Shell wrappers + CLI preserved | AC-HCF-007 |
| REQ-HCF-008 | Documentation crosswalk consistent | AC-HCF-008 |
| REQ-HCF-009 | Opt-in active-creator documented | AC-HCF-009 |
| (Constraint 4.4) | No AskUserQuestion in subagent domain | AC-HCF-010 |
| (Constraint 4.1) | Cross-platform build PASS | AC-HCF-011 |
| (Constraint 4.1) | No NEW test regressions | AC-HCF-012 |
| (Success 8.1) | No NEW spec-lint findings | AC-HCF-013 |
| (Constraint 4.4) | Handle() byte-identity | AC-HCF-014 |

100% REQ coverage: 9/9 REQs have ≥1 mapped AC. 100% AC traceability: 14/14 ACs trace back to ≥1 REQ or explicit constraint.

## 5. Technical Approach Details

### 5.1 Test placement strategy

For M1 (writeHookOutput test): the function is currently package-private (`writeHookOutput`, lowercase). The test MUST live in `package cli`. Two viable file locations:
- Extend `internal/cli/hook_test.go` if it exists (need to verify)
- Create new file `internal/cli/hook_writehookoutput_test.go`

Run-phase MUST `ls internal/cli/hook_test.go` first and decide based on existence.

For M2 (settings.json guard): same pattern; either extend existing `internal/template/settings_test.go` or create new file.

For M3 (resolveObservationsPath test): extend `internal/hook/subagent_stop_test.go` if exists, or create.

### 5.2 stdout capture pattern

Two standard Go patterns for capturing stdout in tests:

**Pattern A — os.Pipe()**:
```go
oldStdout := os.Stdout
r, w, _ := os.Pipe()
os.Stdout = w

// invoke function under test

w.Close()
os.Stdout = oldStdout
buf, _ := io.ReadAll(r)
got := string(buf)
```

**Pattern B — function refactor for io.Writer injection**:
Refactor `writeHookOutput` to accept `io.Writer` as parameter, default to `os.Stdout` at the call site. Test passes `&bytes.Buffer{}`.

Pattern A keeps source unchanged but has goroutine-safety concerns (parallel tests). Pattern B is cleaner but expands scope by 1 LOC at the call site.

**Recommendation**: Pattern A for M1 since the change is test-only. If Pattern A proves flaky in parallel test runs, escalate to Pattern B as a fix-up commit.

### 5.3 settings.json JSON-vs-template parsing

The `.tmpl` file is a Go template, so direct `json.Unmarshal` fails on `{{...}}` tokens.

Three options:
- (a) Render the template with a stub `TemplateContext`, then parse JSON. Requires importing the template package.
- (b) Regex-only check for `"WorktreeCreate"\s*:` and `"WorktreeRemove"\s*:` patterns.
- (c) Render via Go's `text/template` with a stub `funcs` map.

**Recommendation**: Option (b) for both files (consistency). The check is a substring guard; we don't need full JSON parsing to detect the key. False-positive risk (e.g., string `"WorktreeCreate"` appearing in a comment) is mitigated by the line-comment stripper.

### 5.4 Working-tree cleanup safety

`internal/hook/.moai/` is verified untracked at run-phase start. Pre-flight `git ls-files` MUST return empty before `rm -rf` executes. If somehow tracked (unexpected), use `git rm -r` instead.

The cleanup is idempotent and safe to re-run.

## 6. Risk Register

### R-HCF-001: Future re-registration via CI bypass

**Description**: A future PR adds `WorktreeCreate` to `settings.json` and skips CI via `--no-verify` or admin merge.

**Probability**: Low (CLAUDE.local.md §23 forbids `--no-verify`; admin merge requires explicit decision).

**Impact**: High (regression returns; manager-develop blocked again).

**Mitigation**:
- REQ-HCF-003 CI guard test catches it at the next CI run on the next PR
- Branch protection per CLAUDE.local.md §23.2 makes CI green a soft requirement (4 status checks)
- Reviewers see the change in PR diff

### R-HCF-002: writeHookOutput() refactor changes plain-text to JSON

**Description**: A future refactor (e.g., extracting writeHookOutput into a per-event strategy) accidentally routes WorktreeCreate through the JSON path.

**Probability**: Medium (the function has clear special-case branching; refactor temptation exists).

**Impact**: High (`{}` regression returns immediately).

**Mitigation**:
- REQ-HCF-004 plain-text-stdout test (AC-HCF-001/002/004) catches it at the refactor PR's CI
- The comment in writeHookOutput() (lines 187-196) explicitly documents the contract

### R-HCF-003: CLAUDE_PROJECT_DIR misconfigured pointing to wrong path

**Description**: In some user environment, `$CLAUDE_PROJECT_DIR` is set but points to a stale or wrong path.

**Probability**: Low (Claude Code sets this consistently).

**Impact**: Medium (`.moai/harness/observations.yaml` could be created in the wrong project's tree).

**Mitigation**:
- `input.CWD` (set by Claude Code in production) takes precedence over the env var, so production paths are unaffected
- Only test/CLI invocations without `input.CWD` exercise the env-var branch
- Documentation in the comment block (M3 target code) explains the resolution order

### R-HCF-004: Test pollution from other test files re-creating internal/hook/.moai/

**Description**: A test in `internal/hook/` (other than `subagent_stop_test.go`) might also trigger `subagent_stop.go` indirectly and re-create the leak.

**Probability**: Low (only `subagent_stop_test.go` directly exercises `dispatchCapture`; integration tests at higher levels use real `input.CWD`).

**Impact**: Medium (manual cleanup needed after each `go test`).

**Mitigation**:
- M3 fix at source eliminates the leak for ALL invocations (not just unit tests)
- Pre-commit cleanup check could be added later but is out of scope for this SPEC

### R-HCF-005: AC-HCF-008 documentation drift on hooks-reference.md tables

**Description**: One of the 4 locales of `hooks-reference.md` may have drifted (e.g., 21 rows on a locale due to a docs-only PR after PR #1044).

**Probability**: Low (PR #1044 explicitly fixed all 4 locales).

**Impact**: Low (drift fix is mechanical; 1 commit).

**Mitigation**:
- M5 inspection step verifies on every run-phase entry
- If drift found, fix-up commit added before merge

### R-HCF-006: Pattern A stdout capture fails in parallel tests

**Description**: M1 uses `os.Stdout` global mutation, which is unsafe for `t.Parallel()`.

**Probability**: Medium (some test runners default to parallelism).

**Impact**: Medium (flaky test).

**Mitigation**:
- Tests in M1 do NOT call `t.Parallel()`. This is the convention for stdout-mutating tests.
- If parallel-safety becomes a requirement, refactor M1 to Pattern B (io.Writer injection). Tracked as follow-up if needed.

## 7. Pre-flight Baseline (already captured in spec.md § 10)

| Metric | Pre-run baseline (HEAD `58a235e06`) |
|--------|--------------------------------------|
| `go test ./internal/cli/...` exit | Expected 0 |
| `go test ./internal/hook/...` exit | Expected 0 |
| `go test ./internal/template/...` exit | 1 (3 baseline FAIL from template_mirror_drift_audit) |
| `go build ./...` exit | 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` exit | 0 |
| `internal/hook/.moai/` exists | YES (3025 bytes in observations.yaml) |
| `internal/template/settings_test.go:512` value | `const expectedCount = 20` |
| settings.json + .tmpl WorktreeCreate keys | 0 occurrences (PR #1044 state) |

Run-phase MUST re-capture this baseline and confirm matches before applying changes (Section C pre-flight per `manager-develop-prompt-template.md`).

## 8. Deferred / Follow-up Work

| Item | Reason for deferral |
|------|---------------------|
| Broader `os.Getwd()` audit across `internal/hook/` | Out of scope. Only `subagent_stop.go` has the documented leak. Other handlers (compact.go, file_changed.go, instructions_loaded.go) use `input.CWD` correctly without `os.Getwd()` fallback. Optional follow-up SPEC `SPEC-V3R6-HOOK-OBS-PATH-AUDIT-001` (provisional). |
| `internal/hook/subagent_boundary_test.go` (general C-HRA-008 guard) | Out of scope. Broader subagent-boundary enforcement is owned by a different SPEC. This SPEC validates the boundary at AC-HCF-010 via grep only. |
| Refactoring `writeHookOutput()` into strategy pattern | Out of scope. Current in-line dispatch is the canonical form. |
| Removing handlers + shell wrappers + CLI subcommands | Explicitly forbidden by REQ-HCF-007 PRESERVE. They remain opt-in infrastructure. |

## 9. Definition of Done

The implementation is complete when:

1. All 14 acceptance criteria pass (AC-HCF-001 .. AC-HCF-014) — verified via the verification commands in acceptance.md
2. Cross-platform build PASS (AC-HCF-011)
3. C-HRA-008 boundary grep returns 0 matches (AC-HCF-010)
4. Handle() byte-identity verified (AC-HCF-014)
5. Working-tree leak cleanup verified absent (AC-HCF-006)
6. SPEC frontmatter `status: draft` → `status: implemented`, `version: 0.1.0` → `version: 0.2.0`
7. progress.md captures: pre-flight baseline, AC matrix, commits SHA, any deviations from plan
8. Run PR opened, CI green (or baseline residual 3 FAIL only), merged to main

## 10. Open Questions

None blocking. Two design choices are deferred to run-phase author judgment:

1. **stdout capture pattern A vs B (M1)** — recommend A unless parallelism required
2. **subagent_stop.go refactor extent (M3)** — recommend approach 3 (extract `resolveObservationsPath` helper) for testability; alternative is inline test with side-effect inspection

Neither blocks plan-phase completion. Both are run-phase implementation choices.
