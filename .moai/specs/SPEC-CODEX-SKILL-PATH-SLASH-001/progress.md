# SPEC-CODEX-SKILL-PATH-SLASH-001 — Progress

Card t540 · branch `WT-codex-path-escape` · Tier M · cycle_type tdd.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-08
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
blocking_gate: AC-CSPS-001 (Codex-on-Windows slash resolution — NOT measurable in this worktree)
```

## §E.2 Run-phase Evidence

Run scope: **M1 + M2 only.** M0 and M3 are NOT in this run (operator decision: "M1+M2 first,
landing deferred"). Full evidence with verbatim command output: `.moai/reports/t540/run-m1-m2.md`.

Commits: `2d66ebad4` (M1 — separator seam), `dc71e8ef9` (M2 — publisher conversion + comparison).
Card base re-derived at read time: `9ce7926377236aa837294cce9a214cb545751e7c`.

### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CSPS-001 | **OPEN — not attempted** | (requires a Windows host) | no Windows host in this worktree; not measured, not simulated |
| AC-CSPS-002 | PASS | `go test ./internal/cli/ -run 'TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm' -timeout 1800s -v` | `--- PASS: TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm (0.00s)` — RED before: `action = 3 (… cannot carry verbatim …), want appended` |
| AC-CSPS-003 arm 1 | PASS (regression guard, no RED claimed) | `go test ./internal/cli/ -run 'TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars' -timeout 1800s -v` | `--- PASS` + 6 subtests (`{/,\\}×{quote,lf,cr}`) |
| AC-CSPS-003 arm 2 | PASS (regression guard, no RED claimed) | `go test ./internal/cli/ -run 'TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost' -timeout 1800s -v` | `--- PASS: TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost (0.00s)`; discriminates the unconditional-`ReplaceAll` mutant |
| AC-CSPS-004 arm B | PASS | `go test ./internal/cli/ -run 'TestConfigPathFromConfigPathRestoresHostSeparator' -timeout 1800s -v` | `--- PASS` — RED before: build failure (undefined), then identity stub `= "C:/Users/u/SKILL.md", want "C:\\Users\\u\\SKILL.md"` |
| AC-CSPS-004 arms A / A' / C | **OPEN — M3, not in this run** | — | `codex_skills_prune.go` / `doctor_codex.go` untouched |
| AC-CSPS-005 | **OPEN — M3, not in this run** | — | not attempted |
| AC-CSPS-006 | PASS | `go test ./internal/cli/... -timeout 1800s -v > ac-006.log; /usr/bin/grep -c -- '--- PASS: '` | `rc=0`, AFTER `6902` vs BASE `6886`, `--- FAIL: ` count `0`; predicate `AFTER>=BEFORE AND AFTER>0` satisfied; +16 delta == tests added |
| AC-CSPS-007 arm 1 | PASS | `go test ./internal/cli/ -run 'TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry' -timeout 1800s -v` | `--- PASS` — RED before: `action = 3 (… cannot carry verbatim …), want updated` |
| AC-CSPS-007 arm 2 | PASS | `go test ./internal/cli/ -run 'TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates' -timeout 1800s -v` | `--- PASS` — RED before: reason was the verbatim-refusal, not the 2-entry duplicate skip |
| AC-CSPS-008 | PASS (filter-liveness discriminator only; the parser-touch mutant is M4's) | `git diff --name-only <re-derived base>..HEAD -- internal/codexwiring/skills.go` | empty probe under a non-zero control (13 rows at `dc71e8ef9`; the control grows as later evidence commits land, the probe stays empty) |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| `internal/codexwiring/skills.go` unmodified (REQ-CSPS-007) | HOLDS | empty probe, non-zero control (§2.9 of the report) |
| No unconditional backslash replacement (REQ-CSPS-010) | HOLDS | mutant 2 CAUGHT by AC-CSPS-003 arm 2 |
| Identity on a `/`-separator host (REQ-CSPS-006) | HOLDS | `TestConfigPathIsIdentityOnSlashSeparator`; AC-CSPS-006 delta is additions only |
| No escape sequence emitted (REQ-CSPS-003) | HOLDS | AC-CSPS-002 asserts the emitted content contains no `\` at all |
| Real `~/.codex/config.toml` never written (REQ-CSPS-009) | HOLDS | this run is pure-function / in-memory `[]byte` only; no `CODEX_HOME` set, no config written |
| Identity-stub seam cannot pass silently | HOLDS | mutant 1 CAUGHT by AC-CSPS-002, 004 arm B, 007 both arms |

### t502 debt

- **F4 — CLOSED.** The `:245` guard had zero coverage (`/usr/bin/grep -rn 'cannot carry verbatim'
  --include='*.go' internal/` → the production line only). Both branches now carry tests: the refuse
  branch (AC-CSPS-003) and the newly-opened publish branch (AC-CSPS-002).
- **F5 — unchanged by this run.** No `skip` branch's return code was touched; it remains a separate
  policy card.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: partial — M1 and M2 complete, M0 and M3 open, landing deferred
run_complete_at: 2026-09-08
run_commit_sha: dc71e8ef9
run_commit_shas: [2d66ebad4, dc71e8ef9]
card_base_rederived: 9ce7926377236aa837294cce9a214cb545751e7c
ac_pass_count: 8      # 002, 003 arm1, 003 arm2, 004 arm B, 006, 007 arm1, 007 arm2, 008
ac_fail_count: 0
ac_open_count: 5      # 001 (Windows gate), 004 arms A/A'/C, 005
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed (git fetch origin develop at pre-flight and before the scope check)
l44_post_push_fetch: n/a — nothing pushed; landing is deferred pending AC-CSPS-001
new_warnings_or_lints_introduced: 0   # golangci-lint ./internal/cli/... → "0 issues."; go vet → rc=0
cross_platform_build:
  host: rc=0            # go build ./...
  windows: rc=0         # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 3   # internal/cli/codex_config_path.go (new), codex_config_path_test.go (new), codex_skills_disable_path_test.go (new) + codex_skills_disable.go (modified)
m1_to_mN_commit_strategy: one commit per milestone (M1, M2); every subject carries t540
blocking_before_landing: AC-CSPS-001 (unmeasured), M3 (reader-side conversion not implemented)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
