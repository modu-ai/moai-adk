# Progress — SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (final revision under the operator-approved one-time extension)
plan_complete_at: 2026-09-11
card: t656
tier: M
artifact_count: 4 (spec.md, plan.md, acceptance.md, progress.md)
spec_version: 0.4.0
era: V3R6
base_tree: 04a8ab731 (internal/ identical to 81c1d58f9 — git diff --stat empty)
branch: WT-update-value-merge
depends_on:
  - SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001
counts:
  requirements: 16 (Tier M ceiling 16)
  acceptance_criteria: 16 (Tier M ceiling 16)
  mutant_rows: 27
  needs_clarification_markers: 0
audits:
  iter1: {verdict: FAIL, score: 0.73, report: .moai/reports/t656/plan-audit-iter1.md}
  iter2: {verdict: FAIL, score: 0.75, report: .moai/reports/t656/plan-audit-iter2.md}
  extension: operator-approved one-time third revision 2026-09-11
decisions:
  A1_B1_C1: confirmed (operator, before plan phase)
  F05_capture_only_when_deploy_wrote: confirmed 2026-09-11 — REQ-USB-016, AC-USB-014; D7 prefers manifest provenance + hash (N-11)
  D5_promotion_rule: decided 2026-09-11 — refined rule; REQ-USB-005 restated as three GEARS sentences (N-01); AC-USB-016 adopted with R3 != R2 (N-03)
  D5_record: option (a)'s purpose preserved; earlier abort explanation superseded (wrong premise)
  D5_implementation: "운영자 규칙의 구현 방식, 리드 수용 (2026-09-11)" — two-point judgement; leftover judged before any step of the next flow removes or rewrites the live .claude/settings.json (N-02, operator direction); positions update.go before :384, init.go before :867; supersedes "before the next flow writes its own staging copy"
  D5_normal_end_signal: merge preserve path taken (merge.go :197-204, :217-225, :229-237), not byte compare (N-10)
  D5_disclosed_deviation: an intervening write to the live file after an abort turns promote into discard (fail-safe) — spec §E, plan D5 (N-08)
  D6_sibling_relation: decided 2026-09-11 — sibling REQ-UMC-010 scoped; wording unchanged this round
  F17_user_deleted_key: accepted 2026-09-11 as a known limitation — spec §B.5, §E
run_phase_verification_items:
  - M1: whether RestoreMoaiConfig writes .claude/settings.json (read so far: restore_entry.go:47-79 calls only RestoreMoaiConfig; auditor read restore.go as .moai/config-only)
red_now_observed: none (plan phase forbids go test — recorded at run-phase M1)

## §E.2 Run-phase Evidence

Run phase started 2026-09-11 (Implementation Kickoff Approval granted by the operator via the lead). Worktree `.claude/worktrees/t656`, branch `WT-update-value-merge`, run base HEAD `41a470641` (local develop `0db675bed` absorbed). Long outputs live under `.moai/reports/t656/run/`.

### §E.2.0 Plan-audit iteration-3 minors (N3-01..N3-06) — disposition

The SPEC ownership matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Forbidden ownership crossings) forbids manager-develop from editing `spec.md` / `plan.md` / `acceptance.md` body content. Every N3 fix below is a body edit of `plan.md` or `acceptance.md`, so none is applied to those files in the run phase. Each is recorded as **debt for manager-spec** together with the run-phase handling that keeps the implementation and its evidence correct without the wording change.

| id | Where | Wording debt (manager-spec) | Run-phase handling (no SPEC body edit) |
|---|---|---|---|
| N3-01 | acceptance.md:124, :285 (M-D5g-w) | Split the row into M-D5g-wb (judgement moved to the backup step → killed only by `update_leftover_version_skip`) and M-D5g-wd (judgement moved after deploy → killed by `update_leftover_abort` `a == 2` and by `update_leftover_version_skip`) | The slot-request mutant list carries the two variants as separate mutants with their own killing cells |
| N3-02 | acceptance.md:115-116 | Add a retired deny entry to the `update_leftover_version_skip` cell and a mutant row M-D5g-s | The run-phase test gives the leftover R2 and the live file the retired entry `Write(./secrets/**)` (a strictly stronger fixture of the same cell); M-D5g-s is in the slot-request mutant list |
| N3-03 | plan.md:151 (D3) vs :207 (M3) | Name the preserve-path signal channel in D3 | Implemented as a sibling function `MergeUserFilesWithOutcome` that returns a per-path outcome; `MergeUserFiles` keeps its signature and becomes a thin wrapper, so D3 ("signature and base injection unchanged") still holds |
| N3-04 | plan.md:208 vs :217, acceptance.md:81 | State the pre-merge observation hook in one milestone | The hook is introduced once, in the `internal/cli` wiring (M4), next to its only consumer AC-USB-005 — consistent with acceptance.md:81. AC-USB-006/016 observe between flows and need no hook |
| N3-05 | acceptance.md:232, :235 | Relabel the c2/c5 next-flow cells: they are green only against the empty-promotion stub with base selection in place, not before implementation | The RED/GREEN evidence below records these cells against that stub, not against the pre-implementation tree |
| N3-06 | acceptance.md:135-137 | Optional: add a leftover-promotion-failure cell | A `backup` package test drives the leftover judgement with a directory planted at the canonical path and asserts a nil-free, non-blocking result with exactly one `settings-snapshot-promote-failed:` line; the `runUpdate` call-site cell is in the slot request |

### §E.2.1 M1 — baseline (this run, tree `41a470641`)

```
$ go test ./internal/cli/update/merge/... ./internal/cli/update/backup/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.656s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.323s
exit=0

$ go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.581s	coverage: 92.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.197s	coverage: 90.1% of statements
exit=0

$ grep -n "^\.moai/cache/$" .gitignore internal/template/templates/.gitignore
internal/template/templates/.gitignore:241:.moai/cache/
.gitignore:352:.moai/cache/
```

Coverage baseline for the DoD "no lower than M1": merge 92.1%, backup 90.1%.

### §E.2.2 M1 — RED-now, merge base (AC-USB-001/003/012/013)

Tree: HEAD `0083064c8` with the new test file `internal/cli/update/merge/settings_snapshot_base_test.go` uncommitted and **no production change**. Full output: `.moai/reports/t656/run/m1-red-merge.txt`.

```
$ go test ./internal/cli/update/merge/ -run 'TestMergeUserFiles_Snapshot' -count=1 -v
    settings_snapshot_base_test.go:105: statusLine.command = old, want new
    settings_snapshot_base_test.go:108: env.PATH = /old, want /new
    settings_snapshot_base_test.go:111: permissions.deny = [A], want [A B]
    --- FAIL: TestMergeUserFiles_SnapshotBaseDeliversTemplateValueChange/with_canonical (0.00s)
    --- PASS: TestMergeUserFiles_SnapshotBaseDeliversTemplateValueChange/control_no_canonical (0.00s)
--- PASS: TestMergeUserFiles_SnapshotBaseKeepsUserEdit (0.00s)
    settings_snapshot_base_test.go:158: conflict line count = 0, want 1:
--- FAIL: TestMergeUserFiles_SnapshotBaseBothChangedReportsConflict (0.00s)
--- PASS: TestMergeUserFiles_SnapshotFallbackMatchesDerivedBase (0.02s)   [5 subtests PASS]
--- PASS: TestMergeUserFiles_SnapshotBaseAddsNewTemplateKey (0.00s)
--- PASS: TestMergeUserFiles_SnapshotBaseScopedToSettingsJSON (0.00s)
    settings_snapshot_base_test.go:249: user-deleted statusLine came back: map[model:sonnet statusLine:map[command:x]]
    --- FAIL: TestMergeUserFiles_SnapshotBaseHonorsUserKeyDeletion/with_canonical (0.00s)
    --- PASS: TestMergeUserFiles_SnapshotBaseHonorsUserKeyDeletion/control_no_canonical (0.00s)
    settings_snapshot_base_test.go:275: conflict line count = 0, want 1:
--- FAIL: TestMergeUserFiles_SnapshotBaseArrayIsWholeLeaf (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.601s
exit=1
```

Each RED is a value assertion, for the stated reason (the derived base reads every shared leaf as "user changed"). The guards AC-002/004/010/011 and both control cells pass before implementation, as acceptance.md §C predicts.

### §E.2.3 M1 — restore behaviour (plan.md D5 verification item)

`internal/cli/update/backup/settings_snapshot_restore_test.go` drives `RestoreFromBackupDir` over a sections backup and over a legacy (no `sections/`) backup that carries `in-memory-backups/.claude/settings.json` and a root `.claude/settings.json`; a control asserts the restore wrote under `.moai/config` (so the check is not vacuous).

```
$ go test ./internal/cli/update/backup/ -run 'TestRestoreFromBackupDir_NeverWritesLiveSettingsJSON' -count=1 -v
    --- PASS: TestRestoreFromBackupDir_NeverWritesLiveSettingsJSON/sections_backup (0.01s)
    --- PASS: TestRestoreFromBackupDir_NeverWritesLiveSettingsJSON/legacy_backup_without_sections (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.514s
exit=0
```

Finding: the restore command never writes the live `.claude/settings.json` — every write target is confined to `.moai/config` by `RestoreTargetContained` (`backup.go:310`), so a backed-up settings file lands under `.moai/config/in-memory-backups/…`. D5 case 5 therefore arises only from a hand revert, and the rule stands as written.

### §E.2.4 M2 — snapshot data model (backup)

RED-stub: skeleton functions returning zero values, tests written against them. Tree: HEAD `0083064c8` + uncommitted skeleton + tests. Full output: `.moai/reports/t656/run/m2-red-backup.txt`.

```
$ go test ./internal/cli/update/backup/ -run '<M2 tests>' -count=1 -v
    settings_snapshot_test.go:117: LoadSettingsSnapshot ok = false, want true
    settings_snapshot_test.go:137: .moai/cache/template-snapshot/claude/settings.json.pending is absent, want {"a":2}
    settings_snapshot_test.go:199: .moai/cache/template-snapshot/claude/settings.json is absent, want {"a":2,"K":1}
--- FAIL: TestInitSettingsSnapshot_SkippedDeployRecordsNothing/fresh_dir
    --- PASS: .../untracked_existing   --- PASS: .../user_modified_existing
    settings_snapshot_test.go:223: "settings-snapshot-write-failed:" lines = 0, want 1:
    settings_snapshot_test.go:251: .moai/cache/template-snapshot/claude/settings.json = {"a":1}, want {"a":2}
    settings_snapshot_test.go:291: .moai/cache/template-snapshot/claude/settings.json = {"a":1}, want {"a":2}
    settings_snapshot_test.go:328: "settings-snapshot-promote-failed:" lines = 0, want 1:
--- PASS: TestSectionsSnapshot_UnaffectedBySettingsSubpath (0.03s)   [cell A, cell B]
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.536s
exit=1
```

GREEN after implementation (committed as `e27866e4c`):

```
$ go test ./internal/cli/update/backup/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.820s
exit=0
```

AC-USB-014 is driven here, through the real non-force deployer (`template.NewDeployer`) plus the staging helper, not through a stub: `fresh_dir` records the render (RED above), the two existing-file cells record nothing and leave the user file byte-identical. The init-flow wiring order is covered by AC-USB-007 `init` (slot).

### §E.2.5 M3 — base selection and flow seam (merge)

GREEN for AC-001/003/012/013 after base selection (the whole merge package, existing tests included):

```
$ go test ./internal/cli/update/merge/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.741s
exit=0
```

RED-stub for the seam: `MergeUserFilesAndSettleSnapshot` with an empty settle (no promotion). Full output: `.moai/reports/t656/run/m3-red-stub-flow.txt`.

```
$ go test ./internal/cli/update/merge/ -run 'TestSettingsSnapshotFlow' -count=1 -v
    settings_snapshot_flow_test.go:167: model = sonnet, want haiku
    --- FAIL: TestSettingsSnapshotFlow_TwoCycles_BaseIsPreviousRender/single_call_order
    --- FAIL: TestSettingsSnapshotFlow_TwoCycles_BaseIsPreviousRender/split_deploy_then_restore_order
    settings_snapshot_flow_test.go:199: canonical snapshot = {"a":1}, want {"a":2,"K":1}
    settings_snapshot_flow_test.go:225: canonical snapshot = {"a":2,"K":1}, want {"a":3,"K":1,"L":1}
    settings_snapshot_flow_test.go:236: canonical snapshot = {"a":1}, want {"a":3,"K":1,"L":1}
    settings_snapshot_flow_test.go:249: canonical snapshot = {"a":1}, want {"a":2,"K":1}
    --- FAIL: .../c1 .../c2 .../c3 .../c4 .../c5 .../c8
    --- PASS: .../c6_init_writes_promotes_despite_bundle_rewrite
    settings_snapshot_flow_test.go:265: promote-failed lines = 0, want 1:
--- FAIL: TestSettingsSnapshotFlow_PromoteFailureDoesNotBlock
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.899s
exit=1
```

c6 is green against this stub because init has no merge: the cell calls `backup.SettleSettingsSnapshot` directly, whose RED was observed in §E.2.4 (`TestSettleSettingsSnapshot/not_preserved_promotes`). The c2/c5 next-flow values are recorded against this stub, per N3-05.

GREEN after wiring the settle:

```
$ go test ./internal/cli/update/merge/ ./internal/cli/update/backup/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	1.092s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.257s
exit=0
```

### §E.2.6 Disclosed implementation decisions (lead review)

- **D7 render-byte source.** `StageDeployedSettingsSnapshot` takes the manifest entry the deployer wrote and stages the file only when the entry is template-managed and its `template_hash` (the SHA-256 of the bytes the deployer rendered and wrote) equals the file's hash. The staged bytes are therefore exactly the deployer's render; the forbidden form — reading the disk file with no provenance proof (M-14) — is not used. Taking the bytes out of the deployer's return value would need a change to `internal/template` (a `DeployResult` field), which is outside this SPEC's module; this choice keeps the SPEC's module boundary and the plan's D7 option 2 (manifest record + hash). If the lead reads plan.md D7 ("렌더 바이트는 배포기에서 직접 받는다") as requiring the return-value route, that is a scope expansion to `internal/template` and needs a re-delegation.
- **Stateless seam.** The lifecycle is three stateless calls (`backup.JudgeLeftoverSettingsSnapshot`, `backup.StageDeployedSettingsSnapshot`, `merge.MergeUserFilesAndSettleSnapshot` / `backup.SettleSettingsSnapshot`); the only state is the staging file on disk plus the merge outcome. Consequences for the mutant table: M-06c, M-06t and M-D5e collapse into one code mutation at seam level (staging promotes itself); M-D5f / M-D5g are placement mutants of the `internal/cli` call sites and are observed in the slot as M-D5f, M-D5g-wb, M-D5g-wd, M-D5g-s.
- **Preserve-path channel (N3-03).** `MergeUserFilesWithOutcome` returns a per-path `MergeOutcome`; `MergeUserFiles` keeps its signature. A merge that returns an error before writing anything (manifest or embedded-FS load failure) is not a preserve path: the live file still holds the render, so the staging copy is promoted.

### §E.2.7 M5 (part) — mutants observed without the slot

Each mutant was applied to the committed tree `c581a6127`, the named test run with `-count=1`, and the file reverted (`git restore` or an inverse edit); `git status --short internal/` was empty afterwards. Verbatim output: `.moai/reports/t656/run/mut-<ID>.txt`.

| ID | Mutation (file) | Killing cell(s) observed RED | exit |
|---|---|---|---|
| M-01 | canonical branch disabled (`merge.go`) | AC-001 `with_canonical`, AC-003, AC-012 `with_canonical`, AC-013 | 1 |
| M-02 | canonical present → write the new render wholesale (`merge.go`) | AC-002 (`model = sonnet, want opus`), AC-003, AC-012, AC-013 | 1 |
| M-04 | unusable canonical → merge returns an error (`merge.go`) | AC-004 `unreadable_dir`, `invalid_json`, `json_array`, `json_null` (`absent` stays green, as it must) | 1 |
| M-06c / M-06t / M-D5e | staging promotes itself right after the write (`settings_snapshot.go`) | AC-006 `single_call_order` and `split_deploy_then_restore_order` (`newKey = <nil>, want 1`), AC-016 c2/c4/c5, AC-008 `promote_failure` (2 lines), and the unit cell `RecordsRenderWithoutTouchingCanonical` | 1 |
| M-08p | promotion failure printed with the write-failed prefix (`settings_snapshot.go`) | AC-008 `promote_failure` (seam) + both backup promote-failure cells | 1 |
| M-08s | staging failure printed with the sections wording (`settings_snapshot.go`) | AC-008 `helper` (0 prefixed lines; sections wording present) | 1 |
| M-09 | `HasSnapshot` inspects the whole snapshot root (`snapshot.go`) | AC-009 cell B (`HasSnapshot = true with only the settings copy present`); cell A stays green | 1 |
| M-11 | canonical base for every `.json` (`merge.go`) | AC-011 (`.mcp.json statusLine.command = new, want old`) | 1 |
| M-14 | staging reads the disk file with no manifest proof (`settings_snapshot.go`) | AC-014 `untracked_existing`, `user_modified_existing`; stale-hash unit cell | 1 |
| M-D5a | always promote — leftover and flow end (`settings_snapshot.go`) | AC-016 c2 next flow, c5 next flow (`a = 1`, `K = <nil>`) | 1 |
| M-D5b | promote only when the merge wrote a result (`settings_snapshot_flow.go`) | AC-016 c3, c8. c4 and c6 stay green in this design: c4's next-flow promotion goes through the leftover judgement and c6 calls the settle directly with no merge | 1 |
| M-D5c | flow-end promotion only when live == staging bytes (`settings_snapshot.go`) | AC-016 c1, c6 (and c4/c5 next-flow canonical) | 1 |
| M-D5d | leftover always discarded (`settings_snapshot.go`) | AC-016 c4 next flow (`a = 2, want 3`) | 1 |
| M-D5i | preserve judged by live == pre-flow user bytes (`settings_snapshot_flow.go`) | AC-016 c3 (`canonical snapshot = {"a":1}`) | 1 |

Deviation from the acceptance.md §E prediction: M-D5b is killed by c3 and c8 here, not by c4/c6 (reason in the table). The mutant is killed either way.

### §E.2.8 M4 wiring — prepared, not run (lead slot required)

Wiring (`go vet ./internal/cli/` exit 0, `gofmt -l` empty; never compiled into a test binary, never run):

| Point | File:line |
|---|---|
| ① leftover judgement, update | `internal/cli/update.go:384` (after the `--dry-run` return, above the deny-rule strip at :404) |
| ① leftover judgement, init | `internal/cli/init.go:875` (above `executor.Execute` at :877) |
| ② staging, template sync | `internal/cli/update_template_sync.go:371` (right after the Deploy Templates deploy) |
| ② staging, clean reinstall | `internal/cli/update_clean_install.go:467` (right after the Step 5 deploy) |
| ② staging, init | `internal/cli/init.go:891` (after `executor.Execute` succeeds, before the autonomy bundle) |
| ③ settle, template sync | `internal/cli/update_template_sync.go:551` (outside the `configBackupPath` block, unconditional) |
| ③ settle, clean reinstall | `internal/cli/update_clean_install.go:515` (before the deny-rule strip at :538) |
| ③ settle, init | `internal/cli/init.go:921` (flow end, after the bundle) |
| D8 seams | `internal/cli/update_settings_snapshot.go` — `preMergeSettingsSnapshotHook`, `applyAutonomyTierBundleFn`, `mergeUserFilesSettlingSnapshot` |
| B7 stale comment | `internal/cli/update_clean_install.go:394-399` rewritten |

Tests written (not run): `internal/cli/update_settings_snapshot_test.go` — AC-005, AC-007 (six cells + the init source-position substitute), AC-008 `clean_reinstall`, N3-06 `update_leftover_promote_failure`. Commands, the RED-stub procedure, regression set and the 15 slot mutants: `.moai/reports/t656/run/slot-request.md`.

Scope checks (this run):

```
$ go list -deps ./internal/cli/update/merge/... | grep -c "^github.com/modu-ai/moai-adk/internal/cli$"
0
$ go list -deps ./internal/cli/update/backup/... | grep -c "^github.com/modu-ai/moai-adk/internal/cli$"
0
$ go list -deps -test ./internal/cli/update/merge/ ./internal/cli/update/backup/ | grep -c "^github.com/modu-ai/moai-adk/internal/cli$"
0
$ golangci-lint run ./internal/cli/update/...
0 issues.
$ GOOS=windows GOARCH=amd64 go build ./internal/cli/update/...
exit=0
$ go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.901s	coverage: 92.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.613s	coverage: 90.3% of statements
```

Coverage is at or above the M1 baseline (merge 92.1 → 92.9, backup 90.1 → 90.3). `internal/merge` and `internal/template/templates/` are untouched (`git diff --stat 41a470641 HEAD -- internal/merge internal/template/templates` is empty — re-measure after the final commit per AC-USB-015).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — M4/M5-cli/M6 evidence from the lead slot is still owed; see §E.2.8>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
