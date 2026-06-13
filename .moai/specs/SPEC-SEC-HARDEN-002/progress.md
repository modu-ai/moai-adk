# SPEC-SEC-HARDEN-002 — Progress

## §A. Plan-phase signal

```yaml
spec_id: SPEC-SEC-HARDEN-002
tier: M
cycle_type: tdd
status: draft
era: V3R6
predecessor: SPEC-SEC-HARDEN-001   # completed (terminal) — fast-follow
scope_origin:
  - SPEC-SEC-HARDEN-001/spec.md §F.3 (uncovered cli package group)
  - SPEC-SEC-HARDEN-001/progress.md "Known limitations" D2 (deferred MEDIUM redirect operators)
baseline:
  origin_main: 443fad912
  local_head: 639fd7a5e   # unpushed SPEC-MERGE-METHOD-CONFIG-001 run commit — PRESERVE
  git_divergence: "0 1"   # benign local-ahead
plan_complete_at: 2026-06-13
plan_status: audit-ready
```

### Milestone map (priority-ordered)

| Milestone | Finding | Severity | Files | AC group |
|-----------|---------|----------|-------|----------|
| M1 | `validateSpecID` helper (enabler) | — | `internal/cli/<sanitizer>.go` + test | AC-SEC2-M1-001..004 |
| M2a | A-F1 worktree-new path traversal | HIGH | `internal/cli/worktree/new.go` + test | AC-SEC2-M2-001,002,004 |
| M2b | A-F3 git argv option smuggling | MEDIUM | `internal/core/git/worktree.go` + test | AC-SEC2-M2-003 |
| M3 | A-F4 spec read-path traversal | LOW | THREE files: `spec_view.go:30` + `spec_status.go:52` + `spec_close.go:98` (CLI `args[0]` guard) + tests. `spec_drift.go` EXCLUDED (no positional SPEC-ID); `closer.go:173` is `spec close`'s deeper transitive sink, NOT modified | AC-SEC2-M3-001..003 |
| M4 | D2 (SEC-HARDEN-001) redirect operators | MEDIUM | `internal/permission/stack.go` + `stack_sec_harden_test.go` | AC-SEC2-M4-001..004 |

### SPEC-ID Pre-Write Self-Check (recorded for audit)

decomposition: SPEC ✓ | SEC ✓ | HARDEN ✓ | 002 ✓ → PASS

(canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; `002` = digit-only end anchor, no alpha suffix; uniqueness verified — `.moai/specs/SPEC-SEC-HARDEN-002/` did not previously exist.)

### Plan-phase artifacts created

- `.moai/specs/SPEC-SEC-HARDEN-002/spec.md` (GEARS REQ-SEC2-M1..M4, §F h3 exclusions, 12-field canonical frontmatter, status: draft)
- `.moai/specs/SPEC-SEC-HARDEN-002/plan.md` (Section A-H, PRESERVE list, file-touch map, milestones)
- `.moai/specs/SPEC-SEC-HARDEN-002/acceptance.md` (Given-When-Then + grep-verifiable RED/NO-REG matrix)
- `.moai/specs/SPEC-SEC-HARDEN-002/progress.md` (this file)

### Plan-auditor verdict (Phase 2.3)

PASS-WITH-DEBT 0.83 (Tier M threshold 0.80 — passes). MUST-PASS 4/4 PASS. Security-critical milestones (M2a/M2b/M4) confirmed accurate. 3 corrections applied post-audit:

- **D1 (SHOULD-FIX)**: M3 over-claimed source anchors. Corrected to exactly THREE files (`spec_view.go:30`, `spec_status.go:52`, `spec_close.go:98` — all guarded at the CLI `args[0]` boundary). `spec_drift.go` DROPPED (no positional SPEC-ID, repo-wide drift command). `spec_close` guard is at the CLI handler, with the deeper transitive path-join sink noted as `internal/spec/closer.go:173` (NOT the guard site, NOT modified). AC-SEC2-M3-002 rewritten to assert the guard call in the 3 CLI handlers (robust positive assertion).
- **D2 (MINOR)**: 4 non-canonical `(Event-detected)` GEARS labels (REQ-SEC2-M1-003 / M2-001 / M2-003 / M4-001) replaced with canonical `(Event-driven)`.
- **D3 (MINOR)**: brittle negative-grep AC (AC-SEC2-M1-004 `grep -c 'strings.Contains.*\.\.'` absence check) replaced with a robust positive assertion (assert `validateSpecID` invoked in each call site).

### Next phase

GATE-2 (plan-to-implement human gate) → `/moai run SPEC-SEC-HARDEN-002` (cycle_type=tdd, Mode 5 sub-agent likely).

---

## §E — Phase 0.95 Mode Selection — Decision: sub-agent

GATE-2: user-approved (run-phase entry confirmed).

| Input parameter | Value |
|-----------------|-------|
| tier | M |
| scope (files) | ~7 (validate_spec_id.go + new.go + worktree.go + spec_view/status/close.go + stack.go) + tests |
| domain count | 2 (Go source: cli/permission/core-git; SPEC artifacts) |
| file language mix | 100% Go source + Go tests |
| concurrency benefit | LOW — coding-heavy, sequential RED→GREEN milestones with shared validateSpecID dependency |
| Agent Teams prereqs | not evaluated (coding-heavy → Mode 5 default) |

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | semantic security fix, not a typo |
| 2 background | no | write-heavy (RED→GREEN code changes) |
| 3 agent-team | no | coding-heavy; Agent Teams reserved for multi-domain research-heavy work |
| 4 parallel | no | coding-heavy per Anthropic coding-task parallelism caveat; M2 depends on M1 helper |
| 5 sub-agent | **yes** | single sequential agent per milestone; M1→M2→M3→M4 share the validateSpecID dependency chain |
| 6 workflow | no | not ≥30-file mechanical transform; security-semantic coding work |

**Decision: sub-agent** (Mode 5, sequential per-milestone)

Justification: This is coding-heavy security work where M2/M3 depend on the M1 `validateSpecID` helper (shared dependency chain) and each milestone is RED-first reproduction. Per Anthropic's coding-task parallelism caveat (most coding tasks have fewer truly parallelizable subtasks than research), Mode 5 sequential sub-agent is the correct default. The orchestrator dispatched this run-phase to a single `manager-develop` sub-agent (cycle_type=tdd).

---

## §F — Phase 0.5 Plan Audit Gate verdict (re-affirmed at run entry)

| Field | Value |
|-------|-------|
| verdict | PASS (skip-eligible) |
| overall score | 0.91 |
| Tier M threshold | 0.80 |
| skip rationale | score ≥ 0.90 AND no plan-PR commit landed since verdict → Phase 0.5 re-execution skipped per spec-workflow.md § Plan Audit Gate skip policy |
| must-pass | 4/4 PASS |

---

## §G — Run-phase Evidence (RED→GREEN per milestone)

### M1 — validateSpecID helper

- RED: `go test -run 'TestValidateSpecID' ./internal/cli` → `undefined: validateSpecID` → build failed (helper not yet defined). Confirmed reproduction.
- GREEN: added `internal/cli/validate_spec_id.go` (single shared helper, modeled on validateSkillID). `go test -run 'TestValidateSpecID' ./internal/cli` → `ok`.
- AC-SEC2-M1-001/002/003/004: PASS. `grep -rc 'func validateSpecID' internal/cli/` → `validate_spec_id.go:1` (exactly one definition).
- status transition: draft → in-progress (this M1 commit).

### M2b — git worktree-add `--` argv separator (A-F3, MEDIUM)

- RED: `go test -run 'TestBuildWorktreeAddArgs' ./internal/core/git` → `undefined: buildWorktreeAddArgs` → build failed. Confirmed.
- GREEN: extracted pure `buildWorktreeAddArgs(branchExists, path, branch)` helper in `internal/core/git/worktree.go`; inserts `--` before user-derived operands (`worktree add -- <path> <branch>` / `worktree add -b <branch> -- <path>`). `Add` refactored to call it. Tests pass; full `internal/core/git` suite green (real `TestWorktreeAdd_*` integration tests prove behavior preservation).
- AC-SEC2-M2-003: PASS. `grep -n '"--"' internal/core/git/worktree.go` → 2 matches.

### M3 — spec view/status/close read-path traversal guard (A-F4, LOW)

- RED: `go test -run 'TestSpec(View|Status|Close)_TraversalRejected' ./internal/cli` → all 3 FAIL (traversal `../../../../etc` resolved to out-of-`.moai/specs/` paths / passed to spec.Close). Confirmed defect.
- GREEN: `validateSpecID(specID)` applied at the CLI `args[0]` boundary of EXACTLY THREE handlers — `viewAcceptanceCriteria` (spec_view.go), `updateSpecStatus` (spec_status.go), spec close RunE (spec_close.go, before `spec.Close`). All 3 + NO-REG test pass.
- AC-SEC2-M3-002: PASS. guard grep = 1 each in the 3 files; `spec_drift.go` = 0 (excluded); `closer.go` NOT in `git diff --name-only` (unmodified).
- Bonus evidence: post-fix M3 close test no longer leaks `internal/cli/etc.lock` (pre-fix RED run created it because `spec.Close` reached the lock-creation sink; post-fix guard rejects before that sink).

### M4 — permission redirect-operator scanner extension (D2, MEDIUM)

- RED: `go test -run 'TestMatches.*Redirect' ./internal/permission` → 4 FAIL (`go test > /etc/cron.d/payload`, `>> ~/.bashrc`, `2> /tmp/x`, `< /etc/shadow` all resolved to ALLOW=true). Confirmed write-primitive escalation.
- GREEN: added `>` and `<` to the unquoted-separator `case` in `hasUnquotedShellSeparator` (stack.go:189), additive only. D1 unterminated-quote guard (`return inSingle || inDouble`) preserved verbatim. Full `internal/permission` suite green (D1 8-case + QuotedRedirect NO-REG + existing separator suite all pass).
- AC-SEC2-M4-001/002/003/004: PASS. `grep "c == '>'"` + `grep "c == '<'"` → present; `grep 'return inSingle || inDouble'` → 1 (D1 guard preserved); `&>`/`>&`/`2>&1` stay denied via existing `&` branch.

### M2a — worktree-new path traversal guard (A-F1, HIGH) — BLOCKED (cross-package import cycle)

- RED tests written + confirmed (traversal `../../../../tmp/evil` creates worktree dir OUTSIDE root; `--path ../../etc/evil` accepted). Reverted from commit pending AC resolution.
- BLOCKER: `validateSpecID` is defined in package `cli` per AC-SEC2-M1-004 line-17 grep (`grep -n 'func validateSpecID' internal/cli/*.go` → exactly 1, non-recursive glob pins definition to `internal/cli/*.go`). But `new.go` is in package `worktree`, and `internal/cli` ALREADY imports `internal/cli/worktree` — so `worktree` CANNOT import `cli` (import cycle). The AC's line-36 grep (`validateSpecID(` called in `internal/cli/worktree/new.go`) + line-17 grep (definition in `internal/cli/*.go`) are mutually exclusive under Go import rules.
- Resolution requires a manager-spec AC amendment (e.g., relocate the shared helper to a leaf package both import, OR scope the line-17 grep to allow `internal/cli/worktree/`). manager-develop returns a blocker report rather than editing acceptance.md (owned by manager-spec) or shipping a design that fails a literal AC grep.

### NO-REG / cross-platform / lint / boundary (post-M2b/M3/M4)

- Full suite (touched packages): `cli` (minus 3 pre-existing `TestDoctor_*` golden failures — confirmed pre-existing on baseline `192bd5f81` with changes stashed; environmental `.claude/hooks/` golden mismatch, unrelated to this SPEC), `cli/worktree`, `cli/harness`, `cli/pr`, `cli/wizard`, `permission`, `core/git` all green.
- Cross-platform: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Lint: `golangci-lint run ./internal/cli/... ./internal/permission/... ./internal/core/git/...` → 0 issues.
- C-HRA-008: 0 AskUserQuestion matches in touched source files.
- Coverage: permission 90.2%, core/git 88.1%.
