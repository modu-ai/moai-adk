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

Sync scope: **CHANGELOG emission only.** No docs-site surface exists for this change (see
`docs_site_surface` below). SPEC body content (spec.md / plan.md / acceptance.md §A-§H) untouched;
frontmatter `status:` deliberately left at `in-progress` pending a lead ruling on whether a
partial-scope SPEC advances.

```yaml
sync_status: partial — write side documented; M0 gate open, M3 delegated to t562
sync_complete_at: 2026-09-08
sync_commit_sha: pending-backfill-t540   # a commit cannot carry its own SHA; backfill on next touch
sync_measured_at_head: e5df637bc
run_commit_shas: [2d66ebad4, dc71e8ef9]
changelog_entry_position: "CHANGELOG.md ## [Unreleased] → ### Fixed, first bullet"
b12_self_test_a: pass   # pre-emission `grep -c 'SPEC-CODEX-SKILL-PATH-SLASH-001' CHANGELOG.md` → 0 before the write, 1 after
b12_self_test_b: pass   # distinct AC ids in acceptance.md → 8 (AC-CSPS-001..008); entry states 5 PASS / 1 partial / 2 open == 8
b12_self_test_c: pass   # every file path cited in the entry verified present via ls at e5df637bc
ac_total_distinct: 8
ac_pass_count: 5        # 002, 003, 006, 007, 008 (fully closed at the AC-id level)
ac_partial_count: 1     # 004 — arm B PASS; arms A / A' / C open (read side, t562)
ac_open_count: 2        # 001 (Windows resolution gate), 005 (read side, t562)
ac_fail_count: 0
open_gate:
  id: AC-CSPS-001
  state: OPEN — not attempted, not simulated
  reason: >-
    Requires observing how Codex resolves a forward-slash path on a Windows host.
    No Windows host exists in this worktree, so the behaviour was neither measured nor
    inferred-as-measured. plan.md §H: "M0 is the gate; nothing lands before it."
  never_record_as: [satisfied, not-applicable, waived]
delegated_scope:
  m3_reader_side:
    to_card: t562
    to_spec: SPEC-CODEX-SKILL-PATH-READBACK-001
    covers: [AC-CSPS-004 arms A / A' / C, AC-CSPS-005]
    files_untouched_here: [internal/cli/codex_skills_prune.go, internal/cli/doctor_codex.go]
    classification: delegated scope, NOT t540 debt
    verification_gap: >-
      The sibling SPEC directory does not exist in THIS worktree (it lives on t562's branch),
      so its stated scope was taken from the dispatch and could not be read here.
sync_phase_verification:                # re-measured by manager-docs this run at e5df637bc
  scoped_tests:
    cmd: "go test ./internal/cli/ -run 'TestConfigPath|TestUpsertCodexSkillDisable' -timeout 1800s"
    out: "ok  github.com/modu-ai/moai-adk/internal/cli\t0.776s"
  go_vet: "go vet ./internal/cli/... → rc=0"
  build_host: "go build ./... → rc=0"
  build_windows: "GOOS=windows GOARCH=amd64 go build ./... → rc=0"
  lint: "golangci-lint run ./internal/cli/... → rc=0, `0 issues.`"
docs_site_surface: none
docs_site_evidence: >-
  grep -rln 'skills.config|skills disable|codex skills' docs-site/content/ → 4 hits, all the
  same page in the 4 locales (utility-commands/moai-clean.md), and that page documents the
  READER (prune) side only: the `[[skills.config]]` ghost-registration model and the
  seven never-deleted path classes. This card changes the WRITER only, so nothing a user
  reads there changes. cli-reference/ carries no `skills` page at all. No page invented.
locales_touched: 0
frontmatter_status_transitions:
  spec_md: "in-progress → implemented (lead ruling 2026-09-08); STOPS here, NOT completed"
  terminal_state_rationale: >-
    acceptance.md:10 pins AC-CSPS-001 as Blocking `YES — nothing lands before it`, and that
    gate is unmeasured. The 3-phase close presumes run actually closed; marking `completed`
    would assert the blocking gate closed when it is open — an unobserved completion claim.
    `implemented` is the honest terminal state: the delivered scope (M1+M2, write side) is
    done and the close remains blocked on the gate. Contract-precondition shortfall, NOT a
    contract deviation.
  plan_md / acceptance_md: untouched (no frontmatter edit, no body edit)
spec_body_modified: false              # spec.md / plan.md / acceptance.md bodies untouched
pushed: false                          # nothing pushed; working tree left dirty for the lane to commit
t502_debt:
  F4: CLOSED — both branches of the `cannot carry verbatim` guard now covered
  F5: unchanged — separate policy card
```

### AC-CSPS-001 — M0-on-CI feasibility finding (partial; distribution + infrastructure only)

The gate stays **OPEN and unmeasured**. This sub-block records only what has been measured about
whether the gate *could* be executed on a Windows CI runner — it does NOT move the gate, and it is
never a basis for recording AC-CSPS-001 as satisfied, waived, or not-applicable.

The AC's own Given clause admits a CI runner: `acceptance.md:23` reads **"a Windows host (physical,
VM, or a Windows CI runner) with codex-cli installed"**. `acceptance.md:44`'s Gap paragraph scopes
its limitation to this tree — "No Windows host is available in this worktree. This AC is not
satisfiable **here** and must be executed **elsewhere**" — so the limitation is worktree-scoped, not
gate-wide.

```yaml
m0_ci_feasibility:
  status: PARTIAL — distribution axis and infrastructure axis satisfied; install/run axis UNVERIFIED
  gate_state_unchanged: true   # AC-CSPS-001 remains OPEN, unmeasured; never satisfied/waived/n-a

  distribution_axis:
    measured_by: lane (this run) — not re-measured by manager-docs
    satisfied: true
    evidence:
      - cmd: "npm view @openai/codex version os cpu bin"
        out: "version = '0.153.4'; bin = { codex: 'bin/codex.js' }"
        rc: 0
      - cmd: "npm view @openai/codex optionalDependencies --json"
        out: >-
          includes "@openai/codex-win32-x64": "npm:@openai/codex@0.153.4-win32-x64"
          and "@openai/codex-win32-arm64": "npm:@openai/codex@0.153.4-win32-arm64"
        rc: 0
      - cmd: "codex --version   # local reference host"
        out: "codex-cli 0.153.4"
    conclusion: >-
      A Windows codex-cli distribution exists for the SAME version measured locally (0.153.4),
      on both win32-x64 and win32-arm64.

  infrastructure_axis:
    measured_by: manager-docs (this run, in this worktree, at HEAD e5df637bc)
    satisfied: true
    evidence:
      - ".github/workflows/ci.yml:376 → os: [ubuntu-latest, macos-latest, windows-latest]"
      - ".github/workflows/release-pr-multi-os.yml:91 → os: [ubuntu-latest, macos-latest, windows-latest]"
      - ".github/workflows/test-install.yml:100 → os: [ubuntu-latest, macos-latest, windows-latest]"
      - ".github/workflows/test-install.yml:146 → runs-on: windows-latest"
      - ".github/workflows/test-install.yml:196 → runs-on: windows-latest"
    conclusion: This repository has Windows CI runners.

  install_run_axis:
    status: UNVERIFIED — GAP
    unverified:
      - whether codex-cli actually installs and runs on `windows-latest`
      - whether `codex debug prompt-input` executes there without interactive auth
    note: >-
      Neither was attempted. The distribution existing is not the same claim as it installing;
      a runner existing is not the same claim as codex running on it.

  do_not_claim:
    claims:
      - "M0 is achievable"
      - "M0 is feasible overall"
      - "M0 is ready to run"
    reason: >-
      Those assertions require the install/run axis, which is unmeasured. Only two of the three
      axes have been measured.
```
