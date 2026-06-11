# SPEC-PREPUSH-SAVE-WIRING-001 — Implementation Plan (Tier S)

> Tier S (Simple). 2-artifact set (spec.md + plan.md); AC enumerated inline in spec.md §D and
> expanded here in § Acceptance. Minimal manager-develop delegation form (~500-800 tokens) is
> sufficient — full Section A-E template is OPTIONAL for Tier S
> (per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability).

---

## A. Context

- **Location**: main checkout `/Users/goos/MoAI/moai-adk-go` (Step 1 plan-phase rule — main checkout).
- **Module**: `internal/config` (single package).
- **The whole change**: add ONE `saveSection(... "git-strategy.yaml", gitStrategyFileWrapper{...})`
  call inside `ConfigManager.Save()`, mirroring the adjacent git-convention WRITE leg, plus
  round-trip + no-regression unit tests.
- **Wrapper reuse**: `gitStrategyFileWrapper` already exists (`types.go:1076-1081`) from
  SPEC-PREPUSH-LOADER-WIRING-001's READ leg — REUSE it; do NOT create a new type.
- **Export seam already complete**: `SetSection`/`GetSection` `case "git_strategy"` already exist
  (`manager.go:261-262`, `:307-312`). The ONLY missing leg is Save(). This SPEC adds exactly that.
- **Chain position**: 4th and final PREPUSH SPEC; the inverse of LOADER-WIRING-001 (which added the
  READ leg and explicitly deferred this WRITE leg via its AC-PLW-008).

## B. Known Issues (filtered to relevant — Tier S)

- **B4 Frontmatter canonical schema** — spec.md uses `created:`/`updated:`/`tags:` (verified; no
  snake_case alias). `tier: S` set in frontmatter to avoid the backward-compat Tier-L threshold.
- **B6 spec-lint heading** — spec.md §C uses `### Out of Scope` (h3, dash bullets) under
  `## C. Exclusions (What NOT to Build)` to satisfy `OutOfScopeRule` (avoids `MissingExclusions`
  ERROR — the recurring lint trap; mirrors the LOADER-WIRING-001 §C structure that passed lint).
- **B8 / B10 working-tree hygiene & scope discipline** — change is bounded to `manager.go` +
  one `_test.go`. Do NOT touch `loader.go`, `types.go`, `validation.go`, `defaults.go`, templates,
  or any `SetSection`/`GetSection` case. No runtime-managed file edits.
- **B9 commit + push** — Hybrid Trunk 1-person OSS: manager-develop commits + pushes main directly,
  Conventional Commits (`feat(SPEC-PREPUSH-SAVE-WIRING-001): M1 ...`), never `--no-verify`.

## C. Pre-flight (착수 전 — single most-relevant baseline)

```bash
# 1. Confirm the WRITE leg is genuinely absent (expect 0 — this SPEC makes it ≥1)
grep -c 'git-strategy.yaml' internal/config/manager.go        # expect 0 before, ≥1 after

# 2. Confirm wrapper already exists (expect 1 — reuse, do NOT recreate)
grep -c 'type gitStrategyFileWrapper' internal/config/types.go  # expect 1

# 3. Baseline: existing Save tests green before change
go test -run 'TestConfigManagerSave' ./internal/config/...
```

## D. Constraints (DO NOT VIOLATE)

- PRESERVE: `loader.go`, `types.go`, `validation.go`, `defaults.go`, all templates, and every
  `SetSection`/`GetSection` case — byte-unchanged.
- The new WRITE leg MUST reuse `gitStrategyFileWrapper` (no new type) and the filename
  `git-strategy.yaml` (exact match to the READ leg, so the round-trip is symmetric).
- Place the new `saveSection` call adjacent to the git-convention leg (`manager.go:185-188`) for
  natural grouping; keep the `fmt.Errorf("save git strategy config: %w", err)` idiom parity.
- No new validator; no `defaults.go` change; no template addition; no `Save()` refactor into a loop.
- Conventional Commits; no `--no-verify`; no force-push to main.

## E. Self-Verification (manager-develop deliverables)

- **E1 AC matrix** — run each AC verification command (see § Acceptance) and report PASS/FAIL +
  actual output.
- **E2 cross-platform** — `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0
  (no syscall surface added, so this is a sanity check).
- **E3 coverage** — `go test -cover ./internal/config/...` does not regress; the new `saveSection`
  call line is covered by the round-trip test.
- **E5 lint** — `golangci-lint run ./internal/config/...` clean (no NEW issues).
- **E6 branch + push** — feat/SPEC-PREPUSH-SAVE-WIRING-001 commits, pushed to main per Hybrid Trunk.

## F. Milestones

| M | Subject | Files | Done when |
|---|---------|-------|-----------|
| M1 | Add git-strategy WRITE leg to `Save()` | `internal/config/manager.go` | One `saveSection(... "git-strategy.yaml", gitStrategyFileWrapper{GitStrategy: m.config.GitStrategy})` call added adjacent to the git-convention leg, mirroring its error-wrap idiom. `go build ./...` exit 0. |
| M2 | Round-trip + no-regression tests | `internal/config/manager_test.go` (or a focused `manager_save_git_strategy_test.go`) | New test: `SetSection("git_strategy", <non-default>)` → `Save()` → fresh `Load()` recovers the value (AC-PSW-002). Existing `TestConfigManagerSaveAndReloadRoundTrip` / `TestConfigManagerSaveCreatesDirectory` extended or asserted still-green (AC-PSW-004). |
| M3 | Verify + commit + push | (no new files) | All AC PASS; `go test ./internal/config/...` full suite green; lint clean; commit + push main. |

> No Rounds (SSE-stall split) needed — well under the 30-task threshold. M1→M2→M3 sequential.

## G. Anti-Patterns (avoid)

- Creating a second wrapper type instead of reusing `gitStrategyFileWrapper` (the READ leg already
  defines it — a duplicate would diverge the round-trip).
- Refactoring `Save()` into a `for`-loop over a section list "while here" — out of scope; keep the
  explicit 6th call (Scope Discipline / Surgical Changes).
- Adding a new validator on the WRITE side — validation runs at Load; adding it here is scope creep.
- Adding a `git-strategy.yaml.tmpl` — Save() creates the file from in-memory config; a template is
  not required and would be a separate template-mirror concern.
- Using a different filename than `git-strategy.yaml` — would break the READ/WRITE round-trip
  symmetry that is the entire point of this SPEC.

## H. Cross-References

- `.moai/specs/SPEC-PREPUSH-LOADER-WIRING-001/` — READ leg (predecessor); AC-PLW-008 deferred this WRITE leg
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix
- `internal/config/CLAUDE.md` — module conventions (section-file layout, wrapper pattern)

---

## Acceptance (Given-When-Then, Tier S inline expansion of spec.md §D)

### AC-PSW-001 — Save() persists git-strategy.yaml (REQ-PSW-001, REQ-PSW-002)

**Given** an initialized `ConfigManager` (`Load()` called),
**When** `Save()` runs,
**Then** `<root>/.moai/config/sections/git-strategy.yaml` exists on disk, and `manager.go`'s
`Save()` references the filename.

```bash
# Source-level (inverse of LOADER-WIRING-001 AC-PLW-008 which asserted 0):
grep -c 'git-strategy.yaml' internal/config/manager.go
# Expected: ≥ 1 (Save() now persists git-strategy.yaml alongside the other 5 sections)
```

```go
// Behavioral: after Save(), the file must exist (extend TestConfigManagerSaveCreatesDirectory
// or add a focused assertion).
for _, f := range []string{"user.yaml", "language.yaml", "quality.yaml",
    "git-convention.yaml", "llm.yaml", "git-strategy.yaml"} {
    if _, err := os.Stat(filepath.Join(sectionsDir, f)); os.IsNotExist(err) {
        t.Errorf("Save() did not create file: %s", f)
    }
}
```

---

### AC-PSW-002 — round-trip preserves a non-default git_strategy value (REQ-PSW-004)

**Given** an initialized `ConfigManager`,
**When** a non-default `GitStrategyConfig` is set via `SetSection("git_strategy", ...)`, `Save()`d,
and reloaded into a fresh manager,
**Then** the non-default value survives the round-trip.

```go
// Default Mode is "team" (defaults.go:235); use "personal" as the non-default probe.
_ = m.SetSection("git_strategy", config.GitStrategyConfig{
    Mode:     "personal",
    Provider: "github",
    // (other fields may be zero — Save persists the struct as-is)
})
if err := m.Save(); err != nil { t.Fatalf("Save() error: %v", err) }

m2 := NewConfigManager()
cfg, err := m2.Load(root)               // or LoadRaw if validation requires full struct
if err != nil { t.Fatalf("Load() after Save() error: %v", err) }
if cfg.GitStrategy.Mode != "personal" {
    t.Errorf("GitStrategy.Mode round-trip: got %q, want %q", cfg.GitStrategy.Mode, "personal")
}
```

> Note for run-phase: if `Load()` validation rejects a sparsely-populated `GitStrategyConfig`,
> seed the probe from `NewDefaultGitStrategyConfig()` and mutate ONE field (e.g. set
> `g.Team.Hooks.PrePush = "enforce"` where default is `"warn"`, or `g.Mode = "personal"`), then
> assert that single field survives. Use `LoadRaw` (validation-skipping) if needed — mirror
> whichever sibling round-trip test pattern keeps the assertion isolated to git_strategy.

---

### AC-PSW-003 — written file uses the existing wrapper / git_strategy: key (REQ-PSW-003)

**Given** the implemented WRITE leg,
**When** `types.go` is inspected and `git-strategy.yaml` is written,
**Then** the same `gitStrategyFileWrapper` (top-level key `git_strategy:`) is used — no new wrapper.

```bash
grep -c 'type gitStrategyFileWrapper' internal/config/types.go   # Expected: 1 (reused, not duplicated)
```

```go
// The written YAML must carry the git_strategy: top-level key (symmetric with the READ leg).
data, _ := os.ReadFile(filepath.Join(sectionsDir, "git-strategy.yaml"))
if !strings.Contains(string(data), "git_strategy:") {
    t.Errorf("git-strategy.yaml missing top-level git_strategy: key; got:\n%s", data)
}
```

---

### AC-PSW-004 — no regression to the 5 existing saved sections (REQ-PSW-005)

**Given** the WRITE leg added,
**When** the existing Save tests run,
**Then** `TestConfigManagerSaveAndReloadRoundTrip` and `TestConfigManagerSaveCreatesDirectory`
still pass; the user/language/quality/git-convention/llm round-trips are unchanged.

```bash
go test -run 'TestConfigManagerSave' ./internal/config/... -v
# Expected: all existing Save tests PASS (no behavioral change to the prior 5 sections)
go test ./internal/config/...
# Expected: FULL package suite green
```

---

### AC-PSW-005 — MUST: no scope creep beyond the mirror (REQ-PSW-006)

**Given** the change,
**When** the diff is audited,
**Then** only `internal/config/manager.go` + one `_test.go` are modified; no new validator, no
`defaults.go`/`loader.go`/`validation.go`/template/`SetSection`/`GetSection` edit.

```bash
git diff --name-only origin/main...HEAD
# Expected: only internal/config/manager.go and internal/config/*_test.go

# validator count unchanged (no new validate* function added):
git diff origin/main...HEAD -- internal/config/validation.go | wc -l   # Expected: 0

# loader / defaults / types untouched on the change commit (wrapper reuse, not new):
git diff origin/main...HEAD -- internal/config/loader.go internal/config/defaults.go | wc -l  # Expected: 0
```

---

## Definition of Done

- [ ] AC-PSW-001 through AC-PSW-005 all pass.
- [ ] `go test ./internal/config/...` (FULL suite) green.
- [ ] `golangci-lint run ./internal/config/...` clean.
- [ ] `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- [ ] Coverage for `internal/config` does not regress.
- [ ] No change to `loader.go`, `types.go`, `validation.go`, `defaults.go`, templates, or
      `SetSection`/`GetSection` (footprint bound per REQ-PSW-006).

## Quality Gate Criteria (TRUST 5)

- **Tested**: round-trip test (set→save→reload) + existing Save no-regression tests; the new
  `saveSection` line is covered.
- **Readable**: WRITE leg mirrors the git-convention leg verbatim in shape; self-documenting.
- **Unified**: identical `saveSection` + `fmt.Errorf("save ...: %w", err)` idiom as siblings.
- **Secured**: atomic temp+rename write preserved (no new write path); no secrets touched.
- **Trackable**: Conventional Commit `feat(SPEC-PREPUSH-SAVE-WIRING-001): ...`; status transitions
  per the ownership matrix.
