# t462 — Codex-axis e2e measurement: run-phase inventory re-measurement

Run-phase tree SHA: `bd7c58201` (branch `WT-codex-e2e`, worktree `.claude/worktrees/t462`)
Re-measured: 2026-09-03, run phase (M2). Plan-phase baseline: `inventory-baseline.md` @ `e9c6a8564`.

## 0. Base movement and adjacent-card state (REQ-CEM-009 / AC-CEM-010)

The plan-phase base moved: the lane absorbed local develop `1b9c02991` before run-phase entry
(commit `bd7c58201`, "absorb local develop (1b9c02991, t456+t468)"). All run-phase measurements in
this file and in `execution-log.md` are pinned to `bd7c58201`; no plan-phase count is re-cited
without this run's own measurement (AP-3).

Adjacent cards:
- **t451 (doctor codex-wiring silence repair): LANDED.** 5 commits in `e9c6a8564..bd7c58201`
  name t451 (`d592b0551` merge, `04f89d39d`, `c2679d712`, `fe6208602`, `b1f04b1dd`) — "moai
  doctor reports unwired codex projects and stale skill registrations".
- **t452 (codex skill-axis wholesale absence): NOT LANDED** — 0 commits in the range name t452
  (`git log --oneline e9c6a8564..bd7c58201 | grep -ci 't451\|t452'` → 5, all t451).

Range contents: 134 commits `e9c6a8564..bd7c58201`; test files added in range include
`internal/codexwiring/skills_test.go`, `internal/cli/doctor_golden_test.go` (modified, gained
codex refs), plus non-codex files (`gate_precommit_test.go` ×2, `precommit_gate_scope_e2e_test.go`,
`update_gate_yaml_preserve_test.go`, `section_gate_test.go`, `landed_test.go`,
`project_continuation_pipeline_signal_test.go`, `vci_ordering_clause_guard_test.go`,
`gate_panel_test.go`).

## 1. Axis 1 — filename glob (spec.md §A commands, verbatim)

```
find internal/cli -name '*codex*_test.go' | wc -l          → 38   (unchanged vs plan)
find internal -name '*codex*_test.go' | wc -l              → 41   (unchanged vs plan)
find internal/codexwiring -name '*_test.go' | wc -l        → 7    (DRIFT: was 6)
```

**Drift statement (axis 1)**: +1 file — `internal/codexwiring/skills_test.go` added in the
absorbed range (codex path-shape / stale-skill work, t451/t468 family). 41 codex-named repo-wide
files unchanged. Axis-1 total 47 → **48**.

## 2. Axis 2 — dependency/symbol sweep (spec.md §A commands, verbatim)

```
find internal/codexadapter -name '*_test.go' | wc -l       → 7    (unchanged)
grep -rlE 'runCodexReviewGate|CodexAudit|codex_task|codex_setup|codex_job|codex-review-gate|CodexWiring|resolveCodexHomeDir|codexCmd|codexadapter|codexwiring|codexRunner' internal/ --include='*_test.go' \
  | grep -v -E '/[a-z0-9_]*codex[a-z0-9_]*_test\.go$' \
  | grep -v '^internal/codexwiring/\|^internal/codexadapter/'      → 22 files (unchanged)
```

**Drift statement (axis 2)**: none — 7 + 22 identical to plan baseline.

## 3. Axis 3 — lexicon delta (spec.md §A command shape)

Pool = `grep -ril 'codex' internal/ --include='*_test.go'` minus filename-axis names minus
codex packages (73 files) minus dependency-axis files (22) = **51 delta files** (DRIFT: was 50).
Per-file `grep -ci codex` classification (rule: ≥5 = behavioral, 1–4 = incidental):

- Behavioral: **28** files (was 27)
- Incidental: **23** files (unchanged)

New behavioral file: `internal/cli/doctor_golden_test.go` — 7 refs (was not in the plan delta;
gained codex references from t451's doctor wiring work). Top of the distribution is unchanged:
`mcp_convergence_participant_test.go` (29), `agentemit/agentemit_test.go` (25),
`agentemit/golden_test.go` (22), `config/audit_models_test.go` (22),
`review_gate_config_key_test.go` (20), `wizard/mcp_audit_test.go` (17).

**Drift statement (axis 3)**: +1 behavioral file (`internal/cli/doctor_golden_test.go`, 7 refs),
incidental set byte-identical in count.

## 4. Union and execution surface

**Union @ `bd7c58201`: 128 test files** = 48 (filename: 41 codex-named + 7 codexwiring) + 29
(dependency: 7 codexadapter + 22 symbol) + 51 (lexicon delta). Plan baseline was 126; drift +2
(`skills_test.go`, `doctor_golden_test.go`), both explained by landed adjacent work.

Every union file sits in a package reached by the M4 recursive execution patterns:

| Execution run (plan M4) | Recursive pattern | Union files covered by construction |
|---|---|---|
| Step 1 | `./internal/codexwiring/... ./internal/codexadapter/...` | 7 + 7 |
| Step 2 | `./internal/cli/...` (standalone, `-timeout 1800s`) | 38 codex-named + 16 symbol + wizard subpackage + lexicon files under `internal/cli/**` |
| Step 3a | `./internal/config/... ./internal/core/project/... ./internal/github/workflow/... ./internal/hook/... ./internal/mcp/... ./internal/sessionmsg/... ./internal/settings/... ./internal/spec/... ./internal/web/...` | the dependency + lexicon files in those 9 packages (incl. `internal/web/mcp_codex_surface_test.go`, `codex_card_sentinel_test.go`) |
| Step 3b | `./internal/template/...` | `codex_agents_deploy_test.go`, `agentemit/*`, template lexicon files |

No union file falls outside the four runs above — checked by construction (each axis's file list
was intersected with the run patterns when assembling this table; the pool/delta lists are
preserved at `/tmp/t462-{lexicon-pool,dep-axis,lexicon-delta}.txt` from this run).

## 5. Machine-state ground truth (re-verified at run phase, `bd7c58201` window)

- `shasum ~/.codex/config.toml` → `ad8c8593a5d89937b9786f1b706384a532361120` (before all
  integration-style checks; re-checked after — see `execution-log.md` §1)
- `ls ~/.codex/skills/` → `.system`, `hatch-pet` (zero moai skills deployed — CONFIRMED, unchanged)
- `command -v codex` → `/Users/goos/.local/bin/codex` (`codex-cli 0.152.1`) — functional binary on
  PATH, so the OPT-OUT live gate matters (M1-D2: `MOAI_SKIP_LIVE_CODEX=1` on every run)
- Installed moai binary: `/Users/goos/go/bin/moai`, v3.2.0-rc.0 — NOT this tree's build; excluded
  from the M3 positive control per plan M3 (AP-6)
