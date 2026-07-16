# Acceptance Criteria — SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001

> Every AC is reproduction-test-shaped: a Go test asserting an OBSERVABLE check status (`CheckOK` / `CheckWarn` / `CheckFail`) or `warnCount`, run against a `t.TempDir()` fixture — NOT a token-presence grep. Both defects have deterministic reproduction steps, so the reproduction IS the RED test (fails before the fix, passes after). All temp dirs use `t.TempDir()` (auto-cleanup; CLAUDE.local.md §6). All ACs authored run-phase; shapes below are the contract.

## §D AC Matrix

| AC | REQ(s) | Defect | Shape | RED before fix? |
|----|--------|--------|-------|-----------------|
| AC-DFS-001 | REQ-DFS-002, 004 | A | telemetry-only `.moai/harness/` → `CheckOK` | YES |
| AC-DFS-002 | REQ-DFS-003 | A | full baseline harness → L1-L6 battery runs | no (preserve) |
| AC-DFS-003 | REQ-DFS-003 | A | partial baseline (misconfigured) → still `CheckFail` | no (guard) |
| AC-DFS-004 | REQ-DFS-005, 006 | B | fresh-install embedded skills → 0 unknown | YES |
| AC-DFS-005 | REQ-DFS-001, 005 | B | every embedded manifest `moai-*` skill → non-WARN (anti-drift) | YES |
| AC-DFS-006 | REQ-DFS-007 | B | bogus `moai-*` dir → `CheckWarn`, count==1 | no (guard) |
| AC-DFS-007 | REQ-DFS-002, 005 | A+B | doctor golden snapshot regenerated & passing | YES |
| AC-DFS-008 | REQ-DFS-008 | preserve | CLEAN-REINSTALL preservation suites green, update path un-diffed | no (guard) |
| AC-DFS-009 | (cross) | all | full `go test ./...` green + cross-platform build | no (gate) |

## §D.1 Scenario Detail (Given-When-Then)

### AC-DFS-001 — Defect A reproduction: telemetry-only directory reports OK (RED)

- **Given** a `t.TempDir()` project root with NO baseline harness files, and a `.moai/harness/usage-log.jsonl` file containing at least one `tool_failure` observation line (reproducing issue #1087 step 2),
- **When** `runHarnessCheck(projectRoot)` is invoked,
- **Then** the returned `DiagnosticCheck.Status == uikit.CheckOK` and its message indicates "no harness configured",
- **And** the status is NOT `CheckFail` and the message does NOT contain `L5 missing`.
- **RED**: on the current code this test FAILS (returns `CheckFail` with `L5:FAIL`); it PASSES after REQ-DFS-002.

### AC-DFS-002 — Defect A: genuine harness still evaluated (preserve)

- **Given** a `t.TempDir()` project with ALL 7 baseline harness files present in `.moai/harness/` (`main.md`, `plan-extension.md`, `run-extension.md`, `sync-extension.md`, `chaining-rules.yaml`, `interview-results.md`, `README.md`) AND a `usage-log.jsonl` also present,
- **When** `runHarnessCheck(projectRoot)` is invoked,
- **Then** the check proceeds through the full L1-L6 battery (the returned message contains the `L1: … L6:` status string, i.e. the check is NOT short-circuited to "no harness configured"),
- **And** the telemetry file's presence does not alter the evaluation outcome.
- **Preserve**: this MUST pass both before and after the fix (the fix must not suppress genuine harnesses).

### AC-DFS-003 — Defect A edge: partial baseline still FAILs (guard)

- **Given** a `t.TempDir()` project with SOME but not all 7 baseline harness files (e.g. only `main.md` present) AND `usage-log.jsonl` present,
- **When** `runHarnessCheck(projectRoot)` is invoked,
- **Then** the harness is treated as "configured" (≥1 baseline file present) and the L1-L6 battery runs, producing `CheckFail` with an `L5 missing:` detail for the absent baseline files.
- **Guard**: prevents the telemetry-exclusion fix from masking a genuinely-misconfigured harness as "not configured".

### AC-DFS-004 — Defect B reproduction: template-fresh project reports zero unknown (RED)

- **Given** a `t.TempDir()` project whose `.claude/skills/` is populated with the embedded template's `moai-*` skill directories (via the template deploy mechanism, or by materializing the embedded `moai-*` skill set),
- **When** `checkSkillsAllowlist(projectRoot, false)` is invoked,
- **Then** the returned `warnCount == 0` and `DiagnosticCheck.Status == uikit.CheckOK`,
- **And** the message does NOT contain `unknown moai- skill(s)`.
- **RED**: on the current code this test FAILS (reports the 10 known-drift skills as unknown → `CheckWarn`); it PASSES after REQ-DFS-005.

### AC-DFS-005 — Defect B anti-drift invariant: every embedded skill is known (RED)

- **Given** the embedded template's authoritative `moai-*` skill set (enumerated from the embedded manifest/FS at test time),
- **When** each embedded skill name is passed to `classifySkill(name)`,
- **Then** none classifies as `"WARN"` — every embedded `moai-*` skill is a KNOWN skill by construction.
- **RED**: on the current code this fails for the 10 drift skills; PASSES after derivation. **Anti-drift**: this test re-fails automatically if a future skill is added to templates without the derivation picking it up — it is the guard against re-drift.

### AC-DFS-006 — Defect B: genuine-unknown still warns (guard)

- **Given** a `t.TempDir()` skills dir populated with the full embedded `moai-*` set PLUS one bogus directory `moai-nonexistent-xyz`,
- **When** `checkSkillsAllowlist(projectRoot, false)` is invoked,
- **Then** `warnCount == 1` (only the bogus directory) and `Status == uikit.CheckWarn`,
- **And** `classifySkill("moai-nonexistent-xyz") == "WARN"` while every embedded skill classifies non-WARN.
- **Guard**: proves the manifest derivation did not disable the check — genuine stale/third-party `moai-*` skills are still surfaced.

### AC-DFS-007 — Doctor golden snapshot regenerated (RED → green)

- **Given** the doctor golden test (`internal/cli/doctor_golden_test.go`) captures the Skills Allowlist and/or Harness 5-Layer rendered output,
- **When** the golden snapshot is regenerated after the fixes and the golden test runs,
- **Then** `go test -run TestDoctorGolden ./internal/cli/` passes,
- **And** each golden delta corresponds to a corrected signal (Skills Allowlist OK instead of 10-unknown warning; Harness 5-Layer OK instead of telemetry-induced FAIL) — verified by inspecting the diff, not blind-accepting.
- **RED**: pre-fix golden (if it encodes the buggy strings) fails after the source fix until regenerated.

### AC-DFS-008 — Preservation: CLEAN-REINSTALL-002 contract intact (guard)

- **Given** the existing update/clean-reinstall preservation test suites (`internal/cli/update_preserve_*_test.go`, `update_namespace_harness_v*_test.go`, and the CLEAN-REINSTALL-002 test set),
- **When** the full `internal/cli` + `internal/harness` test packages are run after the fixes,
- **Then** every preservation test passes UNCHANGED (no test modified to accommodate the doctor fix),
- **And** `git diff --stat` shows NO change to `internal/cli/update.go` or `internal/harness/applier.go` (REQ-DFS-008).
- **Verification**: `go test ./internal/cli/... ./internal/harness/...` exit 0 + `git diff --name-only` excludes update.go / applier.go.

### AC-DFS-009 — Full suite + cross-platform (gate)

- **Given** all fixes and new tests are in place,
- **When** the full verification batch runs,
- **Then** `go test ./...` exits 0, `golangci-lint run` reports no NEW findings vs baseline, `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.

## §D.2 Definition of Done

- [ ] AC-DFS-001 .. AC-DFS-009 all PASS with cited command output (verification-claim-integrity §3 5-section format in the run-phase §E report).
- [ ] Both RED reproduction tests (AC-DFS-001, AC-DFS-004, AC-DFS-005) demonstrably failed on the pre-fix tree and pass post-fix.
- [ ] No modification to `internal/cli/update.go`, `internal/harness/applier.go`, or the telemetry write path (REQ-DFS-008 + §C Out of Scope).
- [ ] @MX tags added per plan.md §G (2 ANCHOR + 3 NOTE).
- [ ] Doctor golden snapshot regenerated with each delta traced to a corrected signal.
- [ ] Coverage for `internal/cli` doctor files ≥ 85% (or characterization-preserved where pre-existing).
- [ ] Conventional Commits per milestone; direct-to-main per Hybrid Trunk (Tier M, Route A).

## §D.3 Traceability

- REQ-DFS-001 → AC-DFS-004, AC-DFS-005 (shared root invariant, both false-signal loci)
- REQ-DFS-002 → AC-DFS-001, AC-DFS-007
- REQ-DFS-003 → AC-DFS-002, AC-DFS-003
- REQ-DFS-004 → AC-DFS-001
- REQ-DFS-005 → AC-DFS-004, AC-DFS-005, AC-DFS-007
- REQ-DFS-006 → AC-DFS-004
- REQ-DFS-007 → AC-DFS-006
- REQ-DFS-008 → AC-DFS-008
- cross-cutting gate → AC-DFS-009
