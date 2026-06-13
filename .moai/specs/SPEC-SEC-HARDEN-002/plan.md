# SPEC-SEC-HARDEN-002 — Implementation Plan

## §A. Context

- **Project root**: `/Users/goos/MoAI/moai-adk-go` (Go CLI tool `moai-adk-go`).
- **Branch / baseline**: `main`. Baseline `origin/main` = `443fad912`; local HEAD = `639fd7a5e` (an unpushed SPEC-MERGE-METHOD-CONFIG-001 run commit — **PRESERVE, do NOT touch**). git divergence `0 1` (benign local-ahead).
- **cycle_type**: `tdd` (reproduction-first; each fix RED-first against the current allow/traversal behavior).
- **Tier**: M (7-10 files, ~150-300 LOC, 3-artifact set: spec.md + plan.md + acceptance.md).
- **Predecessor**: SPEC-SEC-HARDEN-001 (status **completed** — terminal; do NOT re-open). Scope origin: -001 spec.md §F.3 (uncovered `cli` package group) + -001 progress.md "Known limitations" D2 (deferred MEDIUM redirect operators).
- **SPEC artifacts**: `.moai/specs/SPEC-SEC-HARDEN-002/{spec,plan,acceptance,progress}.md`.

### §A.1 Existing infrastructure (PRESERVE + EXTEND map)

| Asset | Disposition | Why |
|-------|-------------|-----|
| `internal/cli/update_archive.go:99-122` `validateSkillID` | **PRESERVE (reference model)** | The new `validateSpecID` is modeled on it; `validateSkillID` itself is untouched (still used by `migrate restore-skill`). |
| `internal/permission/stack.go:172-201` `hasUnquotedShellSeparator` | **EXTEND (additive only)** | M4 adds `>` and `<` to the existing unquoted-separator `case` at line 189. The D1 unterminated-quote guard (lines 198-200) is PRESERVED. |
| `internal/permission/stack.go:127-137` `:*` prefix branch | **PRESERVE** | Already calls `hasUnquotedShellSeparator`; M4 extension flows through unchanged. |
| `internal/permission/stack_sec_harden_test.go` | **EXTEND** | Add M4 redirect cases; the existing M1-001..005 + D1 8-case suite stays green. |
| `internal/cli/worktree/new.go:108,141` | **EXTEND** | Apply M1 helper before `MkdirAll`/`Add` (M2a). |
| `internal/core/git/worktree.go:46,52` | **EXTEND** | Insert `--` before the first user-derived operand in the git `worktree add` argv (M2b). |
| `internal/cli/spec_view.go:30` (`args[0]`) + `spec_status.go:52` (`args[0]`) + `spec_close.go:98` (`args[0]`) — THREE files | **EXTEND** | Apply M1 helper at each CLI `args[0]` boundary (M3). `spec_view`/`spec_status` join at the CLI handler; `spec_close` passes `specID` to `spec.Close` whose path-join sink is at `internal/spec/closer.go:173` — guard at the CLI boundary (`spec_close.go:98`), NOT at closer.go. `spec_drift.go` EXCLUDED (no positional SPEC-ID — repo-wide drift command). `internal/spec/closer.go` is NOT modified. |

## §B. Known Issues (auto-injected, domain-filtered)

- **B3 — Subagent boundary (C-HRA-008)**: All touched CLI files (`internal/cli`, `internal/cli/worktree`) MUST NOT call `AskUserQuestion` / `mcp__askuser`. The new validation paths return structured stderr errors + non-zero exit, NEVER prompt. Verify: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli internal/cli/worktree internal/permission internal/core/git | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0 matches. (`internal/cli/worktree/new_test.go::TestNew_NoAskUserQuestion` is the canonical static guard — keep green.)
- **B1 — Cross-platform build tags**: No new `syscall` usage expected. M2b argv change and M1 string validation are platform-neutral. Verify `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Path-separator rejection MUST reject both `/` and `\` (per `validateSkillID` `ContainsAny(skillID, "/\\")`).
- **B2 — Cross-SPEC policy conflict scan**: M4 extends the SEC-HARDEN-001 M1/D1 scanner. The D1 fix (`f25772d1c`) is on `main` already; M4 is purely additive to the same `case`. Verify the D1 8-case suite is present and green BEFORE editing (`go test -run TestMatches_UnterminatedQuoteBypass ./internal/permission`).
- **B5 — CI 3-tier awareness**: spec-lint, golangci-lint, Test (per OS) each fail independently. The §F Exclusions use `### F.N Out of scope` h3 sub-sections with dash bullets to satisfy spec-lint `OutOfScopeRule` (avoid `MissingExclusions`).
- **B6 — spec-lint heading**: §F authored as `### F.N Out of scope` h3 (NOT bare `## Exclusions` h2). Already done in spec.md.
- **B8 / B10 — Working-tree hygiene + scope discipline**: PRESERVE the unpushed `639fd7a5e` commit, the uncommitted `SPEC-MERGE-METHOD-CONFIG-001/spec.md` lint fix, untracked dirs `SPEC-COMPLETION-MARKER-RETIRE-001` / `SPEC-HOOK-DISCIPLINE-WIRING-001`, and all `.moai/research/` `.moai/reports/` artifacts. `git add` only this SPEC's run-phase files. No runtime-managed file edits.
- **B9 — Commit + push self-perform (Hybrid Trunk)**: manager-develop performs M1-M4 commits + push on `main` directly (Tier M, per git-workflow-doctrine). Conventional Commits (`feat(SPEC-SEC-HARDEN-002): M{N} <subject>`). `--no-verify` PROHIBITED. If a parallel-session race is detected at push, return a blocker report to the orchestrator (do NOT force-push).
- **B11 — AskUserQuestion forbidden**: subagent boundary. Any blocker → structured blocker report, never a prose question.

## §C. Pre-flight (run-phase entry — execute before any code change)

```bash
# 1. Branch + baseline
git branch --show-current                       # expect: main
git rev-parse HEAD                               # expect: 639fd7a5e (preserve)

# 2. Cross-platform build baseline
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. SEC-HARDEN-001 D1 scanner present + green BEFORE M4 edit
go test -run 'TestMatches_UnterminatedQuoteBypass' ./internal/permission

# 5. PRESERVE targets visible (must NOT be modified)
git status --porcelain | grep -E 'MERGE-METHOD-CONFIG|COMPLETION-MARKER-RETIRE|HOOK-DISCIPLINE-WIRING'

# 6. validateSkillID reference model present (M1 model)
grep -n 'func validateSkillID' internal/cli/update_archive.go
```

## §D. Constraints (DO NOT VIOLATE)

- **PRESERVE** (no modification): the unpushed `639fd7a5e` commit; uncommitted `.moai/specs/SPEC-MERGE-METHOD-CONFIG-001/spec.md`; untracked dirs `SPEC-COMPLETION-MARKER-RETIRE-001`, `SPEC-HOOK-DISCIPLINE-WIRING-001`; all `.moai/research/`, `.moai/reports/` artifacts; runtime-managed `.moai/state/`, `.moai/cache/`, `.moai/logs/`.
- **PRESERVE** `validateSkillID` (reference model only — do NOT modify; still used by `migrate restore-skill`).
- **PRESERVE** the SEC-HARDEN-001 D1 unterminated-quote guard (`stack.go:198-200`) — M4 is additive to the separator `case`, NOT a rewrite.
- **Do NOT** over-constrain the documented `--path` escape hatch beyond the `..`-after-`filepath.Clean` rejection (REQ-SEC2-M2-004).
- **Do NOT** ship a `$`-blacklist for D3 `${IFS}` (out of scope, §F.1 — dependency-blocked on `mvdan.cc/sh`).
- **Do NOT** re-allow `&>` / `>&` / `2>&1` (already denied via `&` branch; leave conservatively denied per REQ-SEC2-M4-004).
- **Forbidden commands**: `--no-verify`, `git commit --amend` (on pushed history), force-push to main.
- **Required**: Conventional Commits, `🗿 MoAI` trailer, `Authored-By-Agent: manager-develop` body trailer (ownership matrix — `draft → in-progress` on M1 commit).
- **Binary constraint**: C-HRA-008 grep 0 matches across all touched packages.

## §E. Self-Verification Deliverables (manager-develop reports on completion)

- **E1**: AC binary PASS/FAIL matrix (every AC-SEC2-M{1,2,3,4}-* with verification command + actual output). SSOT = acceptance.md.
- **E2**: Cross-platform build (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`) → both exit 0.
- **E3**: Coverage `go test -cover ./internal/cli/... ./internal/cli/worktree/... ./internal/core/git/... ./internal/permission/...` — no regression vs baseline; new helper ≥ 85%.
- **E4**: Subagent boundary grep (B3 command above) → 0 matches.
- **E5**: Lint status (NEW vs baseline) — `golangci-lint run --timeout=2m`.
- **E6**: New commit SHAs + `git push origin main` result.
- **E7**: Blocker report (if any) — structured, AskUserQuestion never called.

## §F. Milestones (priority-ordered; no time estimates)

> Reproduction-first: each milestone lands the RED test FIRST (demonstrating the CURRENT allow/traversal behavior), then the minimal fix, then NO-REG verification.

### M1 — `validateSpecID` sanitizer helper (enabler) — Priority: High

- **RED**: table-driven test asserting the helper (not yet existing) rejects `..`, path separator (`/`, `\`), absolute path; accepts a legitimate `SPEC-SEC-HARDEN-002`-style ID.
- **GREEN**: add ONE shared `validateSpecID` helper in `internal/cli`, modeled on `validateSkillID` (reject `filepath.IsAbs`, `strings.Contains(id, "..")`, `strings.ContainsAny(id, "/\\")`), returning a structured error.
- **AC**: AC-SEC2-M1-001..004.
- **Files**: `internal/cli/<sanitizer>.go` (new or existing CLI file — §F.4 run-phase decision) + `_test.go`.

### M2 — Apply at `worktree new` + `--` argv separator — Priority: High (closes the HIGH)

- **M2a (RED)**: test invoking `runNew` with `'../../../../tmp/evil'` asserting the CURRENT code creates a dir outside `~/.moai/worktrees/<project>/` (FAIL pre-fix). **GREEN**: call `validateSpecID(specID)` immediately after `specID := args[0]` (new.go:108), before path construction / `MkdirAll` / `Add` — reject → no dir created.
- **M2b (RED)**: test asserting the git `worktree add` argv accepts a `-`-leading operand (`'--upload-pack=x'`) as an option. **GREEN**: insert `--` before the first user-derived operand in `worktree.go:46` (`add path branch` → `add -- path branch`) and `worktree.go:52` (`add -b branch path` → `add -b branch -- path`, or equivalent placement preserving `-b branch` as the option pair). NO-REG: legitimate path/branch still create the worktree.
- **AC**: AC-SEC2-M2-001..004.
- **Files**: `internal/cli/worktree/new.go` (+ test), `internal/core/git/worktree.go` (+ test).

### M3 — Apply at spec view/status/close read boundaries (THREE files) — Priority: Medium

- **RED**: test invoking the three positional-SPEC-ID entry points with `'../../../../etc'` asserting the CURRENT code constructs an out-of-`.moai/specs/` read path / passes the traversal to `spec.Close` (FAIL pre-fix).
- **GREEN**: call `validateSpecID(specID)` at the CLI `args[0]` boundary of exactly THREE subcommands:
  - `spec_view.go:30` (after `specID := args[0]`, before the `:49` `filepath.Join` read-path)
  - `spec_status.go:52` (after `specID := args[0]`, before the `:77` `filepath.Join` read-path)
  - `spec_close.go:98` (after `specID := args[0]`, before calling `spec.Close(specID, opts)` — the path-join sink is the deeper transitive `internal/spec/closer.go:173`; guarding at the CLI handler stops traversal before it reaches that sink). `internal/spec/closer.go` is NOT modified.
  - `spec_drift.go` is EXCLUDED — repo-wide drift command, RunE/PostRunE only, no `args[0]`.
- **AC**: AC-SEC2-M3-001..003.
- **Files**: `internal/cli/spec_view.go`, `internal/cli/spec_status.go`, `internal/cli/spec_close.go` (+ tests) — THREE files. NOT `spec_drift.go`, NOT `closer.go`.

### M4 — Permission redirect-operator scanner extension (SEC-HARDEN-001 D2) — Priority: Medium

- **RED**: extend `internal/permission/stack_sec_harden_test.go` with cases asserting the CURRENT scanner resolves `go test > /etc/cron.d/payload`, `go test >> ~/.bashrc`, `go test 2> /tmp/x`, `go test < /etc/shadow` to ALLOW (FAIL pre-fix).
- **GREEN**: add `>` and `<` to the unquoted-separator `case` in `hasUnquotedShellSeparator` (stack.go:189) — additive only, same quote-aware semantics. NO-REG: quoted `go test "a > b"` / `go test -run 'TestGreater>'` STAY ALLOW; D1 8-case + existing separator suite stay green; `&>`/`>&`/`2>&1` stay denied.
- **AC**: AC-SEC2-M4-001..004.
- **Files**: `internal/permission/stack.go`, `internal/permission/stack_sec_harden_test.go`.

## §G. Anti-Patterns (avoid)

- Per-call-site bespoke SPEC-ID validation instead of ONE shared helper (violates REQ-SEC2-M1-004).
- `$`-blacklist hack for D3 (out of scope; high false-positive).
- Rewriting `hasUnquotedShellSeparator` instead of an additive `case` extension (regression risk; the D1 fix shape is the precedent).
- Over-constraining `--path` (documented escape hatch).
- Touching the PRESERVE list (unpushed commit / sibling SPEC artifacts / research / reports).
- Asserting the fix without a RED-first test (reproduction-first mandate).

## §H. Cross-References

- Predecessor: `.moai/specs/SPEC-SEC-HARDEN-001/spec.md` §F.3 + `progress.md` "Known limitations" D2/D3.
- D1 fix precedent: SEC-HARDEN-001 `f25772d1c` (quote-aware scanner hardening — M4 mirrors its shape).
- Reference model: `internal/cli/update_archive.go:99-122` `validateSkillID`.
- Conventions: `internal/cli/CLAUDE.md` (subagent boundary, `filepath.Abs` rule, exit-code discipline).
- Delegation template: `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier M → full Section A-E).
- Status ownership: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix.
