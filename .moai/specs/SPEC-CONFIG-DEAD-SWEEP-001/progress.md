# Progress — SPEC-CONFIG-DEAD-SWEEP-001

## §E.1 Plan-phase Audit-Ready Signal

_status: v0.1.0 plan-phase artifacts authored 2026-08-04; **v0.2.0 amendment authored 2026-08-14**; awaiting plan-auditor verdict._

**Amendment note (v0.2.0).** This SPEC's v0.1.0 scope was implemented and merged as commit
`4c88bbce9` (PR #1325) and then partially reverted by `7171880a9` (PR #1409). That run's evidence
was **never recorded here** — §E.2/§E.3 below stayed empty through the whole cycle, so this file
does not reflect it. The amendment does not backfill it: the first run's result no longer describes
the tree (the revert undid most of it), and citing it would be an unobserved claim. The
re-run establishes its own §E.2 baseline from the current tree (REQ-CDS-010, AC-CDS-011).

SPEC ID regex check, executed 2026-08-14:

```
$ ID="SPEC-CONFIG-DEAD-SWEEP-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

## §E.2 Run-phase Evidence

All evidence below was measured in the `chore/config-dead-sweep-resume` worktree with base
`04a46d0b0`, on 2026-08-14. Per REQ-CDS-010, no evidence from the first run (`4c88bbce9`) is
cited: every baseline was re-measured against the current tree before any edit.

### Pre-flight baselines (measured before any edit)

| Measurement | Command | Observed |
|---|---|---|
| Research plumbing (non-test) | `grep -rn "ResearchConfig\|researchFileWrapper\|loadResearchSection\|\.Research\b" internal/ cmd/ pkg/ \| grep -v _test.go \| wc -l` | `20` |
| StateDir in types/defaults | `grep -n "StateDir" internal/config/types.go internal/config/defaults.go` | 3 lines (`types.go:682`, `defaults.go:150`, `defaults.go:742`) |
| StateDir total (non-test) | `grep -rn "StateDir" internal/ cmd/ pkg/ \| grep -v _test.go \| wc -l` | `55` |
| `.State.StateDir` readers | `grep -rn "\.State\.StateDir" internal/ cmd/ pkg/ \| grep -v _test.go \| wc -l` | `0` (justification, not a post-condition) |
| Symmetry cases | `grep -c "structType:" internal/config/audit_struct_yaml_symmetry_test.go` | `7` |
| Build | `go build ./...` | exit `0` |
| Full suite | `go test ./... -count=1` | exit `1` — 119 ok, `FAIL internal/hook` (2 tests). **Pre-existing, environment-induced; see Pre-existing baseline failure below.** |

**Pre-existing baseline failure (recorded before editing, so it cannot be misattributed).**
`TestSessionStartHandler_Handle` and `TestSessionStartAdditionalContextSkippedOnEmptySessionID`
fail in `internal/hook` at base `04a46d0b0`, before any edit. Cause: the session running the
suite exports `MOAI_KANBAN=1` / `MOAI_KANBAN_ID=tjpyre`, and the handler injects a Kanban
`AdditionalContext` the tests assert must be absent. Attribution proved by re-running with the
variables unset:

```
$ env -u MOAI_KANBAN -u MOAI_KANBAN_ID -u MOAI_KANBAN_LEAD_ADDR \
      -u MOAI_KANBAN_SETTINGS_INJECTED go test ./internal/hook/ -run 'TestSessionStart' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	1.370s
```

This failure is unrelated to the sweep and is present identically before and after it.

### AC matrix

| AC ID | Severity | Deciding command | Observed | Status |
|---|---|---|---|---|
| AC-CDS-003 | MUST | `ls internal/template/templates/.moai/config/sections/research.yaml .moai/config/sections/research.yaml` | both `No such file or directory`, exit `1` (baseline: both present, 680 bytes) | **PASS** |
| AC-CDS-004 | MUST | plumbing grep (above) / `grep -n "research" internal/config/slice.go internal/config/audit_registry.go` / `grep -n "research.yaml" …/settings-management.md` | `0` / no matches (exit 1) / no matches (exit 1) — baseline `20` | **PASS** |
| AC-CDS-005 | MUST | `grep -c "state_dir" internal/template/templates/.moai/config/sections/state.yaml`; `grep -n "StateDir" internal/config/types.go internal/config/defaults.go`; `grep -n "\.moai/state" internal/cli/state.go internal/worktree/state_guard.go` | `0` (baseline `1`); no matches, exit `1` (baseline 3 lines); literal still present at `state.go:210/218/235` and `state_guard.go:25/29/36/53` | **PASS** |
| AC-CDS-007 | MUST | `go build ./...`; `go vet ./...`; `go test ./... -count=1` | build `0`, vet `0`; suite identical to baseline (119 ok, same 2 pre-existing `internal/hook` failures) | **PASS-WITH-DEBT** — see Residual risk |
| AC-CDS-008 | MUST | `moai init <scratch> --non-interactive` then `moai update` and `moai update --templates-only --force --yes` | `UPDATE_EXIT=0`, `UPDATE_FORCE_EXIT=0`; zero `research` mentions in output; `cache.yaml` deployed (1127 bytes); `state.yaml` deploys as `state: {}`; `research.yaml` NOT deployed | **PASS** |
| AC-CDS-009 | MUST | `go test ./internal/config/ -run TestAuditLoaderCompleteness -count=1` | `ok … 0.606s`. Coupling proved non-vacuous — see Guard-coupling demonstration below | **PASS** |
| AC-CDS-010 | MUST | `go test ./internal/config/ -run TestAuditRegistry -count=1` | `ok` — registry row and `knownSections` entry removed together | **PASS** |
| AC-CDS-011 | MUST | four-observation check (below) | all four hold | **PASS** |
| AC-CDS-012 | MUST | `grep -n "research" internal/config/audit_loader_completeness_test.go internal/config/audit_registry.go` | no matches, exit `1` — no allowlist/exception entry added | **PASS** |
| AC-CDS-013 | MUST | template diff inspection | removed files deleted outright; edited template `state.yaml` and `settings-management.md` carry no SPEC ID, REQ token, or commit SHA | **PASS (local)** — CI-confirmed only after PR |
| AC-CDS-014 | MUST | `git diff --name-only HEAD \| grep -i cache` | no matches, exit `1`; `acknowledgedDedicatedLoaders` `"cache"` entry still at `audit_loader_completeness_test.go:35`; template `cache.yaml` still 1127 bytes | **PASS** |
| AC-CDS-015 | MUST | `grep -rn "StateDir" internal/ cmd/ pkg/ --include='*.go' \| grep -v _test.go \| wc -l` | `52` = baseline `55` − 3, exactly | **PASS** |
| AC-CDS-016 | MUST | `go test ./internal/config/ -run TestStructYAMLSymmetry -count=1`; `grep -c "structType:" …` | `ok`; `7` (unchanged) | **PASS** |
| AC-CDS-017 | MUST | `ls internal/settings/testdata/sections/research.yaml`; `go test ./internal/settings/... -count=1` | present (678 bytes); `ok` for all three settings packages | **PASS** |
| AC-CDS-018 | SHOULD | comment read-back | `loader.go:346` now reads `// LIVE: harness learning sub-system, consumed by internal/cli/hook.go`; `audit_registry.go:74` now reads `partial direct-read — internal/config/observability_master.go reads the enabled key live…`; the dangling `parallel to loadResearchSection` comment reworded at `loader.go:281` | **PASS** |

### AC-CDS-011 four-observation precondition (measured before any edit)

| # | Observation | Expected | Observed |
|---|---|---|---|
| 1 | `ls .moai/config/sections/cache.yaml` | absent | `No such file or directory` |
| 2 | local `state.yaml` `state:` block | `state: {}` + removal comment | `state: {}`; the one `state_dir` grep hit is inside the removal comment |
| 3 | `ls internal/template/templates/.moai/config/sections/research.yaml` | present | present, 680 bytes |
| 4 | `grep -n state_dir` template `state.yaml` | present at line 5 | `5:  state_dir: ".moai/state"` |

All four hold — the tree matched the half-reverted state the amendment measured, so the run proceeded.

### Guard-coupling demonstration (AC-CDS-009 non-vacuity)

The loader was removed first, with the template file still present, and the guard was run in that
intermediate state to prove it is capable of failing rather than passing by accident:

```
$ ls internal/template/templates/.moai/config/sections/research.yaml
-rw-r--r--@ 1 goos staff 680 Aug 14 04:37 …/research.yaml
$ go test ./internal/config/ -run TestAuditLoaderCompleteness -count=1
--- FAIL: TestAuditLoaderCompleteness (0.00s)
    audit_loader_completeness_test.go:131: YAML_SECTION_NO_LOADER: research (add loader or acknowledge in allowlist)
    audit_loader_completeness_test.go:136: Uncovered sections: [research]
FAIL
```

The pair was then completed by deleting the template file (not by adding an allowlist entry), and
the guard returned green. The committed tree contains no intermediate red state.

### Deviation — NFR-CKH-002 non-vacuity floor lowered (not anticipated by the SPEC)

Removing the shipped `research` section (24 leaf keys) and `state.state_dir` (1 key) dropped the
shipped-key surface from **914 to 889**, below the `minimumShippedKeys = 900` non-vacuity floor in
`internal/config/shipped_key_reader_test.go`, failing
`TestShippedConfigKeysHaveReaders/non_vacuous_inventory`:

```
shipped_key_reader_test.go:127: NFR-CKH-002: shipped-key enumeration has 889 entries, need >= 900
```

The floor was lowered to `875` with an explanatory comment. Rationale: the floor is a magnitude
sanity check against an enumeration that silently collapses, not a pin on the exact surface size,
and 875 preserves a margin comparable to the original (914 − 900 = 14; 889 − 875 = 14). This is
**not** an AC-CDS-012 suppression: no `research` allowlist or exception entry was added, and the
AC-CDS-012 grep returns zero matches. It is nonetheless a guard-threshold relaxation the SPEC did
not foresee, and is flagged here for auditor review.

The delta was verified as fully attributable rather than a symptom of an incomplete removal: with
both files transiently restored the count read 890, and with the Go metadata also restored the
surface is 914 — matching 889 + 24 + 1 exactly.

### Verification commands (post-change)

```
$ go build ./...                 → exit 0
$ go vet ./...                   → exit 0
$ go test ./internal/config/... ./internal/settings/... ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/config          1.665s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile 0.249s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy 0.844s
ok  	github.com/modu-ai/moai-adk/internal/settings         1.321s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm  0.395s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch 0.782s
ok  	github.com/modu-ai/moai-adk/internal/template        21.261s
$ golangci-lint run ./internal/config/...   → 0 issues
```

`make build` ran after the template edits (AP-9). It regenerated `internal/template/catalog.yaml`
in place with **no content change** (`git status` reports the file unmodified) — expected, because
the catalog indexes agents and skills, not config section files.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-14
run_commit_sha: "505a00a37"
run_status: complete
ac_pass_count: 15          # AC-CDS-003,004,005,008,009,010,011,012,013,014,015,016,017,018 + 007 (PASS-WITH-DEBT)
ac_fail_count: 0
ac_pass_with_debt_count: 1 # AC-CDS-007 — full suite carries pre-existing, environment-induced failures
preserve_list_post_run_count: 3   # cache surface, internal/settings research fixtures, unrelated StateDir symbols — all verified unchanged
l44_pre_commit_fetch: not-applicable   # isolated worktree; no push performed by run-phase
l44_post_push_fetch: not-applicable    # run-phase does not push (orchestrator owns PR)
new_warnings_or_lints_introduced: 0    # golangci-lint ./internal/config/... → 0 issues
cross_platform_build:
  darwin_arm64: pass                   # go build ./... exit 0 on darwin/arm64
  other_platforms: not-measured        # no cross-compile performed this run
total_run_phase_files: 15
m1_to_mN_commit_strategy: single-commit  # M1-M5 land atomically; REQ-CDS-009 coupled pairs cannot appear half-applied in history
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
