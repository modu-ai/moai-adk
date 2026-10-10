## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-10-10T09:40:46Z
- plan_status: audit-ready
- plan_audit: iteration 1 returned FAIL with eight blocking findings (`.moai/reports/t1663/plan-audit.md`). Revision 0.1.1 answers FLR-B1 to FLR-B8 and point (d); the next audit is iteration 2.
- renewal_period: working value 5 minutes (`workflow.factory_lease.renew_interval`, decision-index Q1). Bound: no longer than one half of `workflow.factory_lease.duration`, which is 7 minutes 30 seconds at the 15-minute default; the whole-minute maximum at the default is 7 minutes. The value is ruled at decision-index Q1 (d-20261010T095818Z-c501).
- stall_window: counted from the last observed channel change (REQ-FLR-013). The choice is ruled at decision-index Q6 (d-20261010T095731Z-5ddc).

Debts from the iteration 2 verdict PASS-WITH-DEBT (audited at 3c100b4e6; verdict file `.moai/reports/t1663/plan-audit-iter2.md`, committed at 88ed24d68):

- debt: AC-FLR-011-RECORD dispose_in=run Observe one refused renewal's snapshot in the run phase's §E.2 evidence, or keep the model-written limit named in progress.md.
- debt: AC-FLR-014-SEAM dispose_in=run Make the AC-FLR-014 Given reachable through the existing factoryLaneRuleForSource surface with no new symbol, so its predicted RED is observed rather than TOOL_FAILURE.
- debt: AC-FLR-006-DOMAIN dispose_in=run Add an unparseable-duration subtest and name the sub-two-minute case (no whole-minute renew interval exists) in the refusal text.
- debt: Q5-CITE dispose_in=sync Correct decision-index Q5 to cite draft spec.md:147 and auto-semantics.md:108, not draft REQ-FRR-005, which carries no 7,27,47 literal.

## §E.2 Run-phase Evidence

### M1 block (card t1663; M1 status: RED recorded, GREEN blocked on the read-path ruling)

Scope and constraint
- Lane update applied: no key is registered in any config registry (card t1636 owns registration). Not made: the shipped template and rendered `workflow.yaml` keys, the shipped-key inventory entries, the symmetry case, and the `WorkflowConfig` struct field. Evidence at `75d30ad4b`: `git diff --name-only b292eab19 HEAD` lists the two files `spec.md` and `internal/homestate/factory_lease_config_test.go`. Nothing changed under `internal/config/`, under the template or rendered config, or under `.moai/specs/SPEC-FACTORY-RECORD-REPAIR-001/`.
- Pre-flight baseline, HEAD `b292eab19`, this run: `go build ./internal/...` exit 0; `GOOS=windows GOARCH=amd64 go build ./internal/config/... ./internal/homestate/...` exit 0; `golangci-lint run --timeout=2m ./internal/config/... ./internal/homestate/...` printed `0 issues.`, exit 0. Retired-marker grep on both packages: the hits are comments and run-retirement identifiers, none on the lease path. Subagent-boundary grep `grep -rn 'AskUserQuestion' internal/hook | grep -v _test.go | grep -v '// '`: one match, `internal/hook/pre_tool.go:856:	if input.ToolName == "AskUserQuestion" {`, a tool-name comparison in the PreToolUse guard and not a prompt. `internal/hook` is outside M1 and was not touched.

Commits (local, not pushed)
- `75d30ad4b` feat(SPEC-FACTORY-LEASE-RENEW-001): M1 RED refusal tests for lease keys (card t1663). Carries the `draft -> in-progress` transition on `spec.md` and the RED tests.

RED (HEAD `75d30ad4b`, this run; `go test ./internal/homestate/ -run '^(TestFactoryOpenRefusesNonPositiveLeaseDuration|TestFactoryOpenRefusesSubTwoMinuteDuration|TestFactoryOpenRefusesRenewIntervalOutsideBound|TestFactoryOpenAcceptsRenewIntervalAtBound)$' -v -count=1`, exit 1)

```
--- FAIL: TestFactoryOpenRefusesNonPositiveLeaseDuration (subtests zero, negative, unparseable)
    factory_lease_config_test.go:65: OpenFactory accepted the configuration; want a refusal
--- FAIL: TestFactoryOpenRefusesSubTwoMinuteDuration
    factory_lease_config_test.go:81: OpenFactory accepted the configuration; want a refusal
--- FAIL: TestFactoryOpenRefusesRenewIntervalOutsideBound (subtests over-bound-at-10m-duration, not-whole-minute-at-15m-duration, over-bound-at-15m-duration, above-59-minutes-at-4h-duration)
    factory_lease_config_test.go:122: OpenFactory accepted the configuration; want a refusal
--- PASS: TestFactoryOpenAcceptsRenewIntervalAtBound (subtests 5m-at-10m-duration, 7m-at-15m-duration)
```

The PASS line is vacuous on the unchanged tree, because no code reads the key. It is not a witness for AC-FLR-007 and is not recorded as GREEN.

Package coverage, same tree: `go test -count=1 -cover ./internal/homestate/` printed `coverage: 82.0% of statements`. The only failing functions are the three RED functions above. The QG-3 target of 85% is not met, and the production code is unchanged in this pass.

GREEN: none recorded. The GREEN of AC-FLR-006 and AC-FLR-007 is blocked on the read path (below).

AC matrix, M1 (this run; baseline attribution HEAD `75d30ad4b`)

| AC | Status | Evidence |
|---|---|---|
| AC-FLR-005 | NOT WRITTEN | The configured subtests need the read path. The default subtests pass on the unchanged tree, and their RED needs a default-mutant that was not run. |
| AC-FLR-006 | RED observed; GREEN blocked | `TestFactoryOpenRefusesNonPositiveLeaseDuration`, subtests zero, negative, unparseable (the unparseable arm is AC-FLR-006-DOMAIN) |
| AC-FLR-006-DOMAIN, sub-two-minute case | RED observed; GREEN blocked | `TestFactoryOpenRefusesSubTwoMinuteDuration` |
| AC-FLR-007 | RED observed; GREEN blocked; accept arm vacuous | `TestFactoryOpenRefusesRenewIntervalOutsideBound`; `TestFactoryOpenAcceptsRenewIntervalAtBound` passes without exercising the key |
| AC-FLR-008 | NOT EXECUTED | Registration-free (fixture only), but it needs the factory package under the heavy-package slot and a mutant run. Deferred. |
| AC-FLR-009 (M1 part) | NOT EXECUTED | The configured subtest needs the read path. A defaults-only status surface is the failure pattern AC-FLR-009 names, and it was not written. |

Blocker (structured, E7)

| Parameter | Type | Expected values | Rationale |
|---|---|---|---|
| Read path for `workflow.factory_lease.duration` and `workflow.factory_lease.renew_interval` | ruling (lane or leader) | (a) typed fields `WorkflowConfig.FactoryLease` in `internal/config/types.go`; (b) an untyped reader in `internal/homestate` over the project `workflow.yaml`; (c) wait for card t1636 | OpenFactory must read the keys before any refusal is possible. Plan B3 counts the struct field as registration, so (a) is held. (b) is a second read path that t1636 would have to reconcile. |

Open decisions (the `Operator verdict:` lines in `decision-index.md` stay empty; this lane writes none)
- Q1 (FOUNDER): the lane update confirms the two defaults (15 minutes and 5 minutes) and the bound (`renew_interval` a whole number of minutes from 1 to 59, and at most half of `duration`). The M1 tests assert them. Not implemented: blocked on the read path. The armed form (`*/N`, decision-index Q1 sub-question d) is not in the update and is not implemented in M1. Verdict owed.
- Q2 (FOUNDER): the lane update confirms refuse at open with no clamp, and it requires the refusal text to state the allowed range and the config key path. The M1 tests assert the key path and the allowed-range phrases. Not implemented: blocked. Verdict owed.
- Q3 (FOUNDER): implemented as the body states (REQ-FLR-004, refuse with no write). The Default row says (b), reap-then-refuse, which conflicts with the body. The body governs per the run brief. Not in M1 scope (M2). Verdict owed.
- Q5 (DEFAULT-APPLIED at plan close, option a): the draft folder is unchanged at `75d30ad4b`. The applied default stands.
- Q6 (FOUNDER): implemented as the body states (REQ-FLR-013, `last_change_at`). The Default row says (a), the previous-snapshot comparison, which cannot give a stall verdict below N. The body governs per the run brief. Not in M1 scope (M3). Verdict owed.

Slot commands
- None. No `internal/cli`, `internal/factory`, or `internal/hook` test ran in this pass, so no `moai slot acquire` or `moai slot release` was issued.

Open for the next pass (not started): AC-FLR-005 (both halves), AC-FLR-008 with its mutant, AC-FLR-009 (M1 part), and the M1 GREEN after the ruling. M2 to M5 not started.

### M1 resume checkpoint (ruling d-20261010T100853Z-98a1; card t1663)

Commit `bbde389f9` (feat(SPEC-FACTORY-LEASE-RENEW-001): M1 workflow.factory_lease keys in config and anti-rot entries (card t1663)) carries resume items 1 and 3. The RED block above is unchanged. Baseline attribution: HEAD `0662f6238` before the change, `bbde389f9` after.

Items 1 and 3, observed:
- `make build` at `0662f6238`: exit 0, with the tail line `catalog.yaml updated successfully (24330 bytes)`. `git status --short` after the build lists only the five committed files.
- `go test -count=1 -cover ./internal/config/`: before the change at `0662f6238`, `coverage: 84.8% of statements`; after the change at `bbde389f9`, `ok  	github.com/modu-ai/moai-adk/internal/config	5.190s	coverage: 84.8% of statements`. The anti-rot test `TestShippedConfigKeysHaveReaders` passes with the two inventory entries.
- `golangci-lint run --timeout=2m ./internal/config/...`: `0 issues.`, exit 0.
- `GOOS=windows GOARCH=amd64 go build ./...` at `bbde389f9`: exit 0.

Blocker (resume item 2, the workflow symmetry case). Observed at `0662f6238` plus the case, before it was reverted (`go test -count=1 -cover ./internal/config/`):

```
--- FAIL: TestStructYAMLSymmetry_Workflow (0.00s)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.todo, side=go-only (key "todo" exists in struct but not in workflow.yaml)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.subagent_write_guard, side=go-only (key "subagent_write_guard" exists in struct but not in workflow.yaml)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.agent_stop_guard, side=go-only (key "agent_stop_guard" exists in struct but not in workflow.yaml)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.integration_lock, side=go-only (key "integration_lock" exists in struct but not in workflow.yaml)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.settings_drift_gate, side=go-only (key "settings_drift_gate" exists in struct but not in workflow.yaml)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.anchor_relocation_guard, side=go-only (key "anchor_relocation_guard" exists in struct but not in workflow.yaml)
    audit_struct_yaml_symmetry_test.go:267: CONFIG_STRUCT_YAML_MISMATCH: field=WorkflowConfig.session_worktree, side=go-only (key "session_worktree" exists in struct but not in workflow.yaml)
FAIL
coverage: 84.8% of statements
FAIL	github.com/modu-ai/moai-adk/internal/config	17.144s
```

No `factory_lease` line appears in the list, so the new key is symmetric. The seven fields are pre-existing `WorkflowConfig` tags with no key in `workflow.yaml`. Making the case pass needs either those seven keys added to the template (registry scope) or the symmetry test scoped down (a test change). Neither is covered by the ruling. The case was reverted before commit `bbde389f9`, which restored `audit_struct_yaml_symmetry_test.go` to HEAD. The diff being reverted held only the two additions made for this case.

Not started, held behind the blocker: resume items 4 (OpenFactory reads and refuses), 5 (factory status JSON), and 6 (queue claim and the AC-FLR-008 mutant). Their GREEN outputs are not recorded here. AC-FLR-005, AC-FLR-006, AC-FLR-007, AC-FLR-008, and AC-FLR-009 remain RED or unexecuted, as the M1 block above states.

### Gap: no struct-to-YAML drift check for factory_lease (ruling d-20261010T104937Z-2161)

Appended after the resume checkpoint above. The checkpoint text, including the blocker from `cc22b6622`, is not rewritten.

- Gap: there is no struct-to-YAML drift check for `factory_lease`. The workflow-section symmetry case (`TestStructYAMLSymmetry_Workflow`) is withheld by ruling d-20261010T104937Z-2161 (option c), so `checkSymmetry` does not compare `WorkflowConfig.FactoryLease` with the `factory_lease` block of the shipped template. The symmetry helper and `symmetryCases` are unchanged.
- Covered instead by: the two anti-rot inventory entries (class W, evidence reader) in `internal/config/testdata/shipped_key_inventory.yaml`, and by this card's own loader and default tests. Those tests are added with item 4 and must cover both keys, `workflow.factory_lease.duration` and `workflow.factory_lease.renew_interval`.
- Handed to t1636: the seven pre-existing `WorkflowConfig` asymmetries (todo, subagent_write_guard, agent_stop_guard, integration_lock, settings_drift_gate, anchor_relocation_guard, session_worktree). They are a config-registry finding, not part of this card.

### M1 resume evidence: items 4 to 6 (decision d-20261010T100853Z-98a1; card t1663)

Commits, in order: `def48caa5` (item 4, OpenFactory and the lease writes), `2aa971514` (item 5, factory status), `66be95abb` (item 6, the AC-FLR-008 test). The Gap entry is in `046bc437f`. Nothing is pushed. The operator verdict lines in decision-index.md stay empty; this lane writes none.

Entry point, recorded and not changed. `config.LoadFactoryLease` reads through `loadYAMLFile` over `<projectRoot>/.moai/config/sections`, the entry point that `LoadQuotaGate` and `LoadSlotLeaseDefaultMaxDuration` use. That entry point does not consult `MOAI_CONFIG_DIR` (`EnvConfigDir`); `ConfigManager.LoadRaw` and `Load` do. The lease keys therefore do not follow the override, as the other workflow keys on this path do. `TestLoadFactoryLeaseIgnoresEnvConfigDir` records this.

Refusal policy, recorded. A `workflow.yaml` that cannot be read or decoded refuses the open with "could not be read". The sibling per-key loaders fall back to defaults; the lease keys do not, because a configured lease must not fall back silently.

RED, mutant, and GREEN, by AC:

- AC-FLR-005 (`TestCardLeaseUsesConfiguredDuration`). RED on `cc22b6622` plus the new test, before item 4: `lease expiry = "2026-09-26T09:15:00Z", want "2026-09-26T09:20:00Z" (write instant plus 20m0s)` for `acquisition/configured`, `transition/configured`, and `renewal/configured`. The default arms pass on the unchanged tree. Mutant (default arms green on the unchanged tree; `DefaultFactoryLeaseDuration = 10 * time.Minute`): `lease expiry = "2026-09-26T09:10:00Z", want "2026-09-26T09:15:00Z" (write instant plus 15m0s)` for `acquisition/default`, `transition/default`, and `renewal/default`. `defaults.go` restored from HEAD and verified clean. GREEN on `def48caa5`: all six arms PASS.
- AC-FLR-006, AC-FLR-006-DOMAIN, AC-FLR-007. RED recorded earlier at `75d30ad4b` (the block above, unchanged). GREEN on `def48caa5`: `TestFactoryOpenRefusesNonPositiveLeaseDuration` (zero, negative, unparseable), `TestFactoryOpenRefusesSubTwoMinuteDuration`, and `TestFactoryOpenRefusesRenewIntervalOutsideBound` (four arms) PASS. The AC-FLR-007 accept arm passed on the unchanged tree and is not a witness; its mutant (`renew >= bound`) gave RED: `OpenFactory refused a bound-satisfying configuration: ... got "5m"`. Restored. GREEN on `def48caa5`.
- AC-FLR-008 (`TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration`). Green on the unchanged claim code: `--- PASS: TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration (0.01s)`. Mutant (surrogate: the claim expiry taken from 20m at `backlog_store.go:1244`, because the claim takes no project root and cannot read the key in place): `claim expiry = 2026-10-10T11:27:40Z, want within [...] (15 minutes after the claim, not the card key's 20 minutes)`. `backlog_store.go` restored from HEAD. GREEN on `66be95abb`, with `TestBacklogClaim_ReclaimsExpired` and `TestBacklogClaim_RenewAfterExpiry` also PASS.
- AC-FLR-009 (`TestFactoryStatusReportsLeaseKeys`, M1 part). RED on the tree before item 5: `lease_duration = <nil>, want "15m0s"` (defaults) and `lease_duration = <nil>, want "20m0s"` (configured). GREEN on `2aa971514`: both subtests PASS.
- Loader and default tests (`internal/config/loader_factory_lease_test.go`, covering both keys). Named mutants: renewal default 4m gave `DefaultFactoryLeaseRenewInterval = 4m0s, want 5m0s`; yaml tag `renew_interval_mutant` gave `RenewInterval:5m0s, want 4m`. Both restored from HEAD.

Mutants restored, verified. `defaults.go`, `types.go`, and `backlog_store.go` were restored from HEAD and are absent from `git status` afterwards. The accept-arm line in `factory_lease.go` reads `if !wholeMinutes || renew > bound {` again. No mutant marker remains in `internal/config` or `internal/homestate`.

Coverage (`go test -count=1 -cover`, scoped to the package):

| Point | internal/config | internal/homestate |
|---|---|---|
| Before item 4 (`0662f6238`, `75d30ad4b`) | 84.8% | 82.0% |
| After item 4 (`def48caa5`) | 84.9% | 82.4% |
| After item 5 (`2aa971514`) | 84.9% | 82.4% |
| After item 6 (`66be95abb`) | 84.9% | 82.4% |

Gap: internal/config (84.9%) and internal/homestate (82.4%) stay below 85%. Gap: internal/cli and internal/factory were not measured as whole packages; the brief limits these heavy packages to named tests, so no package figure exists for them.

Lint, per commit. `def48caa5`: `golangci-lint run --timeout=2m ./internal/config/... ./internal/homestate/...` printed `0 issues.`. `2aa971514`: homestate and cli each printed `0 issues.`. `66be95abb`: factory printed `0 issues.`. `GOOS=windows GOARCH=amd64 go build ./...` exited 0 before each of the three commits.

Slot (go-heavy). Each named internal/cli and internal/factory run and the factory lint ran between `moai slot acquire --resource go-heavy --max-duration 20m` and `moai slot release --resource go-heavy`. The cli lint for item 5 ran outside the slot; this is disclosed, and the slot rule in the brief covers test runs.

E4, pre-existing. `grep -rn 'AskUserQuestion' internal/hook | grep -v _test.go | grep -v '// '` returns `internal/hook/pre_tool.go:856: if input.ToolName == "AskUserQuestion" {` on `0662f6238` and on this tree; M1 does not touch internal/hook.

### M2 block (card t1663; RED recorded, GREEN pending)

Governing decisions: decision-index Q3 (REQ-FLR-004, refusal with no write), ruling d-20261010T095731Z-5ddc. Leader ruling d-20261010T111725Z-3ef9 changes the fail-open behaviour of the M1 keys: (a) an undecodable `workflow.yaml` falls back to the defaults (15 minutes, 5 minutes) with a warning that names the file and the error, and no longer refuses the open; (b) a value of a `factory_lease` key that the bounds reject refuses the open, and the refusal names the key path, the allowed range, and the config file. The ruling also sets the coverage rule (per changed function, decision d-20261010T085601Z-991b) and accepts the AC-FLR-008 proxy mutant of the M1 block as evidence.

Commits (local, not pushed; the HEAD at which each output was taken is named beside it)
- `c8fda81b1` test(SPEC-FACTORY-LEASE-RENEW-001): M2 RED tests (card t1663). Test files only.
- `49cdb0af1` test(SPEC-FACTORY-LEASE-RENEW-001): M2 RED, non-holder legs as subtests (card t1663). Test file only.

Pre-M2 baseline and the old behaviour (HEAD `55f5887ec`, before any edit, this run)

`go test ./internal/homestate/ -run '^TestRenewLeaseRefusalsAndExpiry$' -v -count=1`, exit 0 (baseline, §C.3):

```
=== RUN   TestRenewLeaseRefusalsAndExpiry
--- PASS: TestRenewLeaseRefusalsAndExpiry (1.71s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	2.051s
```

`go test ./internal/config/ -run '^TestLoadFactoryLeaseMalformedFileIsAnError$' -v -count=1`, exit 0. This is the M1 record of the old behaviour that ruling part (a) replaces: a malformed `workflow.yaml` refused the open.

```
=== RUN   TestLoadFactoryLeaseMalformedFileIsAnError
--- PASS: TestLoadFactoryLeaseMalformedFileIsAnError (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.350s
```

The M1 test is replaced in `c8fda81b1` by `TestLoadFactoryLeaseMalformedFileFallsBackWithWarning`; the homestate refusal test `TestOpenFactoryRefusesAWorkflowFileThatCannotBeRead` is replaced by `TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead`. The old names are retired, and the old outputs above are the record of them.

RED (observed at HEAD `c8fda81b1`, after the RED commit; every RED is an assertion failure, and no TOOL_FAILURE occurred)

Rewritten reaped-state assertion, on the pre-M2 `RenewLease` (`go test ./internal/homestate/ -run '^TestRenewLeaseRefusalsAndExpiry$' -v -count=1`, exit 1). The RED is observed before any edit to `card_transition.go`.

```
=== RUN   TestRenewLeaseRefusalsAndExpiry
    f7_record_failure_test.go:41: card after expired renewal = assigned holder="" v2, want run, still held by worker-1, version v1 unchanged (REQ-FLR-004: refused with no write)
--- FAIL: TestRenewLeaseRefusalsAndExpiry (0.87s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/homestate	1.219s
```

AC-FLR-004 and the fail-open homestate tests (`go test ./internal/homestate/ -run '^(TestRenewRefusesExpiredWithoutWrite|TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead|TestFactoryLeaseRefusalNamesTheConfigFile)$' -v -count=1`, exit 1)

```
=== RUN   TestFactoryLeaseRefusalNamesTheConfigFile
    factory_lease_refusal_path_test.go:24: refusal does not name the config file: factory lease configuration refused: workflow.factory_lease.duration must be a positive duration (greater than zero); got "0s"
--- FAIL: TestFactoryLeaseRefusalNamesTheConfigFile (0.00s)
=== RUN   TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead
    factory_lease_resolve_test.go:126: OpenFactory on an undecodable workflow file: err = factory lease configuration refused: the workflow configuration could not be read: parse workflow.yaml: config: invalid YAML syntax, want the open to fall back to the defaults
--- FAIL: TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead (0.00s)
=== RUN   TestRenewRefusesExpiredWithoutWrite
    factory_renew_expiry_test.go:37: refused renewal changed the card: v1 run holder="worker-1" expiry=2026-09-26T08:58:00Z → v2 assigned holder="" expiry=
--- FAIL: TestRenewRefusesExpiredWithoutWrite (0.80s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/homestate	1.004s
```

The AC-FLR-004 RED is the reaping branch: the version is bumped to v2 and the lease is cleared, which is the write that Q3 forbids. Its RED input is the pre-M2 branch, not a mutant, because the criterion is red on the unchanged tree.

Fail-open config (`go test ./internal/config/ -run '^TestLoadFactoryLeaseMalformedFileFallsBackWithWarning$' -v -count=1`, exit 1)

```
=== RUN   TestLoadFactoryLeaseMalformedFileFallsBackWithWarning
    loader_factory_lease_test.go:107: LoadFactoryLease on a malformed file: err = parse workflow.yaml: config: invalid YAML syntax, want the defaults and a warning
--- FAIL: TestLoadFactoryLeaseMalformedFileFallsBackWithWarning (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/config	0.338s
```

Catalog pins (`go test ./internal/mcp/ -run '^TestMoaiMCPTools_(CatalogSize|WriteCapableSet)$' -v -count=1`, exit 1)

```
=== RUN   TestMoaiMCPTools_CatalogSize
    catalog_test.go:30: catalog declares 47 tools, want 48
--- FAIL: TestMoaiMCPTools_CatalogSize (0.00s)
=== RUN   TestMoaiMCPTools_WriteCapableSet
    catalog_test.go:84: write-capable tool count = 20 ([audit_multi codex_audit codex_job_cancel codex_role_audit codex_task factory_complete factory_decide factory_msg_list factory_msg_receipt factory_msg_send factory_next factory_stage glm_job_cancel glm_task goal_arm session_msg_poll session_msg_register session_msg_send todo_add verify_snapshot]), want 21
--- FAIL: TestMoaiMCPTools_WriteCapableSet (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/mcp	0.248s
```

CLI acceptance tests (`go test ./internal/cli/ -run '^(TestFactoryRenewExtendsLiveLease|TestFactoryRenewMatchesMCP|TestNonHolderRenewRefusedCLIAndMCP|TestFactoryRenewRefusalIsMachineReadable)$' -v -count=1`, exit 1, slot `go-heavy` held by this session; the first run at `c8fda81b1`)

```
=== RUN   TestFactoryRenewExtendsLiveLease
    factory_renew_test.go:115: renew: unknown command "renew" for "factory"
--- FAIL: TestFactoryRenewExtendsLiveLease (1.03s)
=== RUN   TestFactoryRenewMatchesMCP
=== RUN   TestFactoryRenewMatchesMCP/live_lease:_the_same_expiry_fields
    factory_renew_test.go:137: MCP tool factory_renew is not bound by registerFactoryCardMCPTools
=== RUN   TestFactoryRenewMatchesMCP/expired_lease:_the_same_refusal_class
    factory_renew_test.go:155: MCP tool factory_renew is not bound by registerFactoryCardMCPTools
--- FAIL: TestFactoryRenewMatchesMCP (1.19s)
    --- FAIL: TestFactoryRenewMatchesMCP/live_lease:_the_same_expiry_fields (0.65s)
    --- FAIL: TestFactoryRenewMatchesMCP/expired_lease:_the_same_refusal_class (0.54s)
=== RUN   TestNonHolderRenewRefusedCLIAndMCP
    factory_renew_test.go:180: MCP tool factory_renew is not bound by registerFactoryCardMCPTools
--- FAIL: TestNonHolderRenewRefusedCLIAndMCP (0.75s)
=== RUN   TestFactoryRenewRefusalIsMachineReadable
    factory_renew_test.go:215: renew --json printed no JSON outcome (stdout "", error unknown command "renew" for "factory"): unexpected end of JSON input
--- FAIL: TestFactoryRenewRefusalIsMachineReadable (0.53s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.589s
```

The first run reached only the MCP leg of `TestNonHolderRenewRefusedCLIAndMCP`, because that test stopped at its first failure. Commit `49cdb0af1` splits the legs into subtests, and the re-run at HEAD `49cdb0af1` observes both:

```
=== RUN   TestNonHolderRenewRefusedCLIAndMCP
=== RUN   TestNonHolderRenewRefusedCLIAndMCP/verb
    factory_renew_test.go:182: renew --json printed no JSON outcome (stdout "", error unknown command "renew" for "factory"): unexpected end of JSON input
=== RUN   TestNonHolderRenewRefusedCLIAndMCP/MCP_tool
    factory_renew_test.go:191: MCP tool factory_renew is not bound by registerFactoryCardMCPTools
--- FAIL: TestNonHolderRenewRefusedCLIAndMCP (1.09s)
    --- FAIL: TestNonHolderRenewRefusedCLIAndMCP/verb (0.00s)
    --- FAIL: TestNonHolderRenewRefusedCLIAndMCP/MCP_tool (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.162s
```

Mutant rule. No M2 criterion is green on the unchanged tree: AC-FLR-001, AC-FLR-002, AC-FLR-003, AC-FLR-004, and the verb half of AC-FLR-011 each fail before any edit, so no M2 mutant is required. The rewritten reaped-state assertion is the one M2 case whose unchanged neighbours (the missing-card, unleased-card, and zero-clock assertions) stay green; they are not new criteria and are not re-recorded as RED.

Boundary of ruling (a), recorded for the leader's confirmation. The warning and fallback apply to every failure of `loadYAMLFile`, which is either a read failure or a decode failure of the whole file. A decode failure therefore includes a YAML type mismatch inside a `factory_lease` key (for example a sequence where a duration string is expected), because such a value stops the file from decoding. A scalar that the duration parser rejects (for example `banana`) is a value the file parsed, so it is case (b): a refusal.

Warning mechanism (ruling part a, chosen for the GREEN step): the repo's config fallback path, `slog.Warn(msg, "path", <file>, "error", err)`, the same call the workflow section uses at `internal/config/loader.go:240`. The default slog handler writes it to stderr. No new logging surface is added.

Slot (go-heavy): acquired by this session for the `go vet ./internal/cli/` compile check (until 11:46:13Z) and released; re-acquired for the CLI RED runs (until 11:47:24Z) and released after the split run.

### M2 GREEN block (card t1663; HEAD `e46a0b836` for each verbatim output below unless a line names another SHA)

Commits (local, not pushed; every subject names card t1663):
- `c8fda81b1` test: M2 RED tests (test files only).
- `49cdb0af1` test: M2 RED, non-holder legs as subtests (test file only).
- `8c2c645d1` docs: M2 RED record in section E.2 (the block above).
- `90a732255` fix: an undecodable `workflow.yaml` falls back to the defaults with a warning; out-of-range lease keys name the config file. Cites d-20261010T111725Z-3ef9.
- `953e092b3` fix: RenewLease refuses an expired lease with no write. Cites d-20261010T095731Z-5ddc.
- `4750b8b57` feat: the `moai factory renew <card> [--run] [--json]` verb (lane-only, beside stage; one implementation `factoryRenewCard`).
- `08ac017e6` feat: the `factory_renew` MCP tool, declared write-capable in `internal/mcp/catalog.go`, through the same `factoryRenewCard`; the catalog pins of `c8fda81b1` go GREEN here.
- `1f15ee972` docs: the MCP catalogue rule and its template mirror name the tool (byte-identical; `make build` ran; no tracked file regenerated).
- `e46a0b836` test: coverage tests for the changed homestate functions below 85% (decision d-20261010T085601Z-991b).

Ruling d-20261010T111725Z-3ef9, applied (part a: fail-open; part b: refusal names the config file):
- Part (a) mechanism: `config.LoadFactoryLease` returns the defaults (15m and 5m) when `workflow.yaml` cannot be read or decoded, and emits `slog.Warn("failed to load factory lease config, using defaults", "path", <workflow.yaml path>, "error", <err>)`, the config fallback path the workflow section uses (stderr through the default slog handler). The function's error return is removed in `90a732255`, because a file-level failure no longer produces an error. The sibling loaders are unchanged.
- Part (a) boundary (for the leader to confirm): every failure of the file read or decode is file-level. A YAML type mismatch inside a `factory_lease` key stops the decode, so it falls back. A scalar that the duration parser rejects (for example `banana`) is a value the file parsed, so it refuses the open.
- Part (b): `resolveFactoryLease` still refuses out-of-range and unparseable values, with no clamp. `loadFactoryLease` appends ` (config: <root>/.moai/config/sections/workflow.yaml)` to the refusal. The key path and the allowed range were already named.
- Per the ruling, the M1 record of the old refusal behaviour stays above, and the AC-FLR-008 proxy mutant is accepted as evidence without an exact mutant.

Mutants (M2): none executed. No M2 criterion is green on the unchanged tree (see the RED block), so the mutant rule does not apply to M2.

GREEN, verbatim (HEAD `e46a0b836`, this run)

`go test ./internal/homestate/ -run '^(TestRenewLeaseRefusalsAndExpiry|TestRenewRefusesExpiredWithoutWrite|TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead|TestFactoryLeaseRefusalNamesTheConfigFile|TestLeaseAcquisitionRefusesAnUnreadableRoster|TestOpenFactoryRefusesAnUnopenableRecord)$' -v -count=1`, exit 0:

```
=== RUN   TestRenewLeaseRefusalsAndExpiry
--- PASS: TestRenewLeaseRefusalsAndExpiry (0.52s)
=== RUN   TestOpenFactoryRefusesAnUnopenableRecord
--- PASS: TestOpenFactoryRefusesAnUnopenableRecord (0.39s)
=== RUN   TestLeaseAcquisitionRefusesAnUnreadableRoster
--- PASS: TestLeaseAcquisitionRefusesAnUnreadableRoster (0.47s)
=== RUN   TestFactoryLeaseRefusalNamesTheConfigFile
--- PASS: TestFactoryLeaseRefusalNamesTheConfigFile (0.00s)
=== RUN   TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead
--- PASS: TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead (0.23s)
=== RUN   TestRenewRefusesExpiredWithoutWrite
--- PASS: TestRenewRefusesExpiredWithoutWrite (0.45s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	2.206s
```

`go test ./internal/config/ -run '^(TestLoadFactoryLeaseMalformedFileFallsBackWithWarning|TestLoadFactoryLeaseReadsBothKeys)$' -v -count=1`, exit 0:

```
=== RUN   TestLoadFactoryLeaseReadsBothKeys
--- PASS: TestLoadFactoryLeaseReadsBothKeys (0.00s)
=== RUN   TestLoadFactoryLeaseMalformedFileFallsBackWithWarning
--- PASS: TestLoadFactoryLeaseMalformedFileFallsBackWithWarning (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.298s
```

`go test ./internal/mcp/ -run '^(TestMoaiMCPTools_CatalogSize|TestMoaiMCPTools_WriteCapableSet)$' -v -count=1`, exit 0:

```
=== RUN   TestMoaiMCPTools_CatalogSize
--- PASS: TestMoaiMCPTools_CatalogSize (0.00s)
=== RUN   TestMoaiMCPTools_WriteCapableSet
--- PASS: TestMoaiMCPTools_WriteCapableSet (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/mcp	0.235s
```

`go test ./internal/cli/ -run '^(TestFactoryRenewExtendsLiveLease|TestFactoryRenewMatchesMCP|TestNonHolderRenewRefusedCLIAndMCP|TestFactoryRenewRefusalIsMachineReadable|TestMoaiMCPServer_RegistrationMatchesCatalog|TestSD_AC012_StageAppliesEdgeAndRenews)$' -v -count=1`, exit 0, slot `go-heavy` held by this session for this run:

```
=== RUN   TestSD_AC012_StageAppliesEdgeAndRenews
=== RUN   TestSD_AC012_StageAppliesEdgeAndRenews/holder_applies_plan_→_plan-audit,_actor_recorded,_lease_renewed
=== RUN   TestSD_AC012_StageAppliesEdgeAndRenews/another_lane_is_refused_by_F1_verbatim
--- PASS: TestSD_AC012_StageAppliesEdgeAndRenews (2.96s)
    --- PASS: TestSD_AC012_StageAppliesEdgeAndRenews/holder_applies_plan_→_plan-audit,_actor_recorded,_lease_renewed (1.21s)
    --- PASS: TestSD_AC012_StageAppliesEdgeAndRenews/another_lane_is_refused_by_F1_verbatim (1.75s)
=== RUN   TestFactoryRenewExtendsLiveLease
--- PASS: TestFactoryRenewExtendsLiveLease (1.19s)
=== RUN   TestFactoryRenewMatchesMCP
=== RUN   TestFactoryRenewMatchesMCP/live_lease:_the_same_expiry_fields
=== RUN   TestFactoryRenewMatchesMCP/expired_lease:_the_same_refusal_class
--- PASS: TestFactoryRenewMatchesMCP (1.47s)
    --- PASS: TestFactoryRenewMatchesMCP/live_lease:_the_same_expiry_fields (0.72s)
    --- PASS: TestFactoryRenewMatchesMCP/expired_lease:_the_same_refusal_class (0.75s)
=== RUN   TestNonHolderRenewRefusedCLIAndMCP
=== RUN   TestNonHolderRenewRefusedCLIAndMCP/verb
=== RUN   TestNonHolderRenewRefusedCLIAndMCP/MCP_tool
--- PASS: TestNonHolderRenewRefusedCLIAndMCP (1.09s)
    --- PASS: TestNonHolderRenewRefusedCLIAndMCP/verb (0.19s)
    --- PASS: TestNonHolderRenewRefusedCLIAndMCP/MCP_tool (0.13s)
=== RUN   TestFactoryRenewRefusalIsMachineReadable
--- PASS: TestFactoryRenewRefusalIsMachineReadable (0.99s)
=== RUN   TestMoaiMCPServer_RegistrationMatchesCatalog
--- PASS: TestMoaiMCPServer_RegistrationMatchesCatalog (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	8.699s
```

Earlier GREEN, observed at the commit named (each before the next commit): at `953e092b3`, `--- PASS: TestRenewLeaseRefusalsAndExpiry (0.75s)` and `--- PASS: TestRenewRefusesExpiredWithoutWrite (0.70s)`; at `4750b8b57`, `--- PASS: TestFactoryRenewExtendsLiveLease (1.13s)` before the SilenceUsage fix, and then `--- PASS: TestFactoryRenewExtendsLiveLease (1.19s)` and `--- PASS: TestFactoryRenewRefusalIsMachineReadable (0.89s)` after it (the commit includes the fix); at `08ac017e6`, the MCP subtests above passed in the same run shape, and `TestMoaiMCPTools_CatalogSize`, `TestMoaiMCPTools_WriteCapableSet`, `TestMCPConsoleRendersAllTools`, `TestMCPConsoleToolCountMatchesCatalog`, and `TestMCPToggleRowWriteCapableBadge` (internal/web) and the internal/settings package passed.

Rule mirror (HEAD `1f15ee972`): `go test ./internal/template/ -run '^(TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks)$' -v -count=1 | tail -4` printed `--- PASS: TestRuleTemplateMirrorDrift/verification-plan-contract.md (0.00s)`, `--- PASS: TestRuleTemplateMirrorDrift/cross-session-messaging.md (0.00s)`, `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/template	0.381s`. The `moai-mcp-tools-catalogue.md` subtest sits inside that set.

Race, scoped (QG-2, partial): `go test -race ./internal/homestate/ -run '^(TestRenewLeaseRefusalsAndExpiry|TestRenewRefusesExpiredWithoutWrite|TestCardLeaseUsesConfiguredDuration|TestOpenFactoryFallsBackOnAWorkflowFileThatCannotBeRead|TestLeaseAcquisitionRefusesAnUnreadableRoster)$' -count=1 -v`, exit 0: every named test `--- PASS`, `ok  	github.com/modu-ai/moai-adk/internal/homestate	8.152s`. The race run of `./internal/factory/` is not done (Gap).

Coverage, per changed function (decision d-20261010T085601Z-991b). Profiles are in the session scratchpad (`config.cover`, `homestate2.cover`, `homestate3.cover`), and each function is read with `go tool cover -func`.

| Package | Changed function | Before the coverage tests | After (HEAD `e46a0b836`) |
|---|---|---|---|
| internal/config | LoadFactoryLease | 100.0% | 100.0% |
| internal/config | NewDefaultWorkflowConfig | 100.0% | 100.0% |
| internal/homestate | OpenFactory | 76.9% | 92.3% |
| internal/homestate | OpenFactoryBounded | 76.9% | 92.3% |
| internal/homestate | planTransition | 84.0% | 85.3% |
| internal/homestate | RenewLease | 89.3% | 89.3% |
| internal/homestate | loadFactoryLease, resolveFactoryLease, EffectiveFactoryLease, refuseFactoryLease, CardLeaseDuration, RenewInterval | 100.0% | 100.0% |

Package totals, context only: internal/config 84.9%; internal/homestate 82.4% before the coverage tests and 82.7% after. internal/cli is not measured here (CI measures it, per the ruling). internal/mcp has no changed function (its change is catalog data).

Lint, `golangci-lint run --timeout=5m` (or `10m` for cli) per package, HEAD `e46a0b836`: `./internal/config/...` → `0 issues.`; `./internal/homestate/...` → `0 issues.`; `./internal/mcp/...` → `0 issues.`; `./internal/cli/...` → `0 issues.` (the cli run held the slot).

Cross-platform build: `GOOS=windows GOARCH=amd64 go build ./...` at HEAD `e46a0b836` printed no output and reported no error.

Formatting: `gofmt -l` over every Go file M2 changed or added printed no output.

Boundary grep (E4), canonical form `grep -rn 'AskUserQuestion' internal/hook --include='*.go' | grep -v _test.go | grep -v '// '`: one match, `internal/hook/pre_tool.go:856: if input.ToolName == "AskUserQuestion" {`, pre-existing (it predates M2, and internal/hook was not touched). The same filter over internal/cli, internal/homestate, internal/config, and internal/mcp returns only help-text lines in unrelated files (`harness.go`, `pr_watch_cmd.go`, `harness_mute.go`, `harness_clusters.go`, `agentlint/agent_lint.go`, `harness/propose.go`, `harness/execute.go`, `harness/install.go`). None of them is a file M2 touched.

Template-First: `make build` at the template edit (tree `08ac017e6` plus the template change, dirty) exited 0 with `catalog.yaml updated successfully (24330 bytes)` and built `bin/moai` (git-ignored). The build regenerated no tracked file (`git status --short` listed only the template file). The local mirror was copied from the template and is byte-identical (`diff -q` printed nothing).

Rule-loading budget: `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` is path-scoped (`paths:` frontmatter). It went from 27029 to 27354 bytes (+325). No always-loaded file changed.

Slot (go-heavy, identity 3e1318d7; each window released by `moai slot release --resource go-heavy`):
- acquired until 11:46:13Z → `go vet ./internal/cli/` (compile check) → released.
- acquired until 11:47:24Z → the CLI RED runs at `c8fda81b1` and the split run at `49cdb0af1` → released.
- acquired until 11:53:23Z → the verb GREEN and regression set → released.
- acquired until 11:56:02Z → the MCP GREEN set → released.
- acquired until 12:09:00Z → `golangci-lint run ./internal/cli/...` → released.
- acquired until 12:10:25Z → the final CLI GREEN record run → released.

Gaps (not observed, or not done in M2)
- AC-FLR-011 procedure half (the watchdog text and its `renew_result` copy step): M3, not run here.
- Console labels: `internal/web/assets/i18n.js` has no `f.mcp.tools.factory_renew.enabled.*` keys in the four locales. The brief bars locale strings in this lane, so the keys are not added; the console shows the tool without a translated label until a locale owner adds them. No test requires the keys (`TestMCPConsoleRendersAllTools` passes without them).
- docs-site pages (`docs-site/content/{en,ja,ko,zh}/guides/mcp-server.md`) list the `project_root` tools and still say twenty-two; they are not updated (outside the M2 list; the four-locale same-PR rule applies to that follow-up).
- `.moai/policies/core/moai-mcp-tools-catalogue.md` is an older, different copy of the catalogue rule; it is not updated (outside the M2 list).
- Race on `./internal/factory/` (QG-2 names it) not run: M2 changed no file in internal/factory.
- internal/cli coverage: not measured (CI, per the ruling).
- `/verify` was not run before the code commits `90a732255`, `953e092b3`, `4750b8b57`, and `08ac017e6`. The equivalent checks named above (scoped tests, golangci-lint per package, the Windows build) were run after each commit as recorded, but `/verify` itself was not invoked.
- Trailer layout: each commit ends with the `🗿 MoAI` line followed by the `Authored-By-Agent: manager-develop` trailer, the layout of the M1 commits. The brief reads as the trailer before the 🗿 line; the M1 layout was kept so the authorship trailer stays a trailer block.
- The lease at 12:30:00Z (run v14) was not renewed by this lane; every commit above landed locally before that time, and nothing was pushed.

Residual risk (what could still be wrong despite the observations)
- The fail-open boundary: a YAML type mismatch inside a `factory_lease` key falls back to the defaults instead of refusing (see the boundary note above). If the leader reads that case as part (b), `loadFactoryLease` needs a per-key decode.
- The non-zero exit of `moai factory renew` on any non-renewed outcome is a design choice the SPEC does not state. The watchdog (M3) treats a refusal as a result, not a stop, so it does not depend on the exit status, but a script that stops on non-zero exits will stop on a refusal.
- `factory_renew` has no i18n label in the console (Gap above).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-10T09:51:05Z SPEC-FACTORY-LEASE-RENEW-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1663/.moai/reports/t1663/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1663/.moai/reports/t1663/plan-audit.md
