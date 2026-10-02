# SPEC-LAUNCHER-ENTRY-FLAGS-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02
spec_version: "0.8.0"   # plan-audit iteration 2 (FAIL 0.79) findings D29-D45 addressed on top of v0.7.0 (D1-D28); Q20-Q22 and Q24 open and non-gating; Tier L
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md]
operator_verdicts_recorded: 20  # decision-index Q1-Q6, Q8-Q15, Q17-Q19, Q23 (Q1 superseded by Q10; Q9 count sentence superseded by Q17); Q7 closed as moot; Q23 (role-declaration carrier) answered after plan-audit iteration 2
orchestrator_rulings_recorded: 1   # Q16 (OD-15): raised as a ruling, CONFIRMED by the operator
open_questions: 4               # Q20-Q22 raised by the plan audit and Q24 raised by the compile proof, all non-gating (smallest-footprint reading written); Q18 (OD-17) settled earlier as Option X
author_choices_listed: 12       # spec.md §D, decision-index Q19 — all accepted (two steps)
requirements: 25                # Tier L ceiling 25 (at the ceiling; REQ-019 split into REQ-019 and REQ-025 in v0.8.0)
criteria: 25                    # Tier L ceiling 25 (at the ceiling)
milestones: 13                  # M0-M11 with M5 split into M5a/M5b; integration units: M0-M1, M2+M3+M4, then M5a, M5b, M6, M7, M8, M9, M10, M11 each alone
tree_measured: a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2
audited_head: 5445e296caa48e6ab9821afe808eac2e7385e897   # plan-audit iteration 2; content outside this SPEC directory equals tree_measured (PV-73); iteration 1 audited b9242da00ce489c4f26efb5f6d447ccafef08c54 (PV-71)
branch: WT-launcher-entry-flags
```

Plan-phase signal: plan-audit iteration 1 returned FAIL (0.72, audited at `b9242da00ce489c4f26efb5f6d447ccafef08c54`; report `.moai/reports/t1399/plan-audit-iter1.md`) and iteration 2 returned FAIL (0.79, audited at `5445e296caa48e6ab9821afe808eac2e7385e897`; report `.moai/reports/t1399/plan-audit-iter2.md`; both local and gitignored); v0.8.0 addresses findings D29-D45 and replaces the caller-grep deletion-order table with a committed cumulative compile proof (PV-73 to PV-90); iteration 3 is the last audit allowed and has not run. Every operator verdict is recorded and confirmed (Q23, the role-declaration carrier, was added after iteration 2); Q20-Q22 (raised by iteration 1) and Q24 (raised by the compile proof) are open and non-gating.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Plan-phase Premise Verification

Every row is a measurement taken in this plan run against tree `a6d3e6fd4` (branch `WT-launcher-entry-flags`; `git rev-parse --short HEAD` re-read before the final revision: `a6d3e6fd4`; `git status --short` showed only this SPEC directory untracked). Binary rows used a binary built from this tree (`go build ./cmd/moai`, exit 0; scratchpad path `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/0dcdf2d5-df5c-4da1-8870-24c2a5861303/scratchpad/moai-t1399`, built from `a6d3e6fd4`) or `go run ./cmd/moai`; the judging build and the tree measured are the same HEAD. Rows PV-1..PV-37 were taken in earlier revisions of this plan run against the same tree; the tree has not moved since. Rows marked "v0.1.0" measured the superseded verb-less design and stay true.

| ID | Command | Observed | Exit |
|----|---------|----------|------|
| PV-1 | `moai` (no arguments; tree-built binary) | banner + help text printed | 0 |
| PV-2 (v0.1.0) | `go run ./cmd/moai -f` | stdout empty; stderr `Unknown shorthand flag: 'f' in -f.` | 1 |
| PV-3 (v0.1.0) | `go run ./cmd/moai -l` | stdout empty; stderr `Unknown shorthand flag: 'l' in -l.` | 1 |
| PV-4 | `go test ./internal/cli -run '^(TestParseLauncherEntryMarksAutoAssignedNumbers\|TestCCFactoryLaneJoinsDiscoveredLeader\|TestGLMFactoryLaneJoinsDiscoveredLeader\|TestCodexFactoryEntryParsingUsesLaneOnly\|TestCodexFactoryLegacyEntryIsRefused\|TestCGRetiredEntryAndModeHaveZeroEffects\|TestCGRetirementCompleteEntryShapesAndCounters)$' -v -count=1` | seven PASS results (swept count 7), `ok github.com/modu-ai/moai-adk/internal/cli 12.521s` | 0 |
| PV-5 | `moai gpt`; `moai cg` (tree-built binary) | `Unknown command "gpt" for "moai"`, exit 1; `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, exit 1 | 1 / 1 |
| PV-6 (v0.1.0) | `go run ./cmd/moai -f lane`; `go run ./cmd/moai -l lane-2` | stdout empty; unknown-shorthand diagnostic (`f` / `l`) | 1 / 1 |
| PV-7 (v0.1.0) | `grep -c -e '--lane' internal/cli/root.go`; `grep -c -e '--factory' internal/cli/root.go` | `0`; `0` | 1 / 1 |
| PV-8 | `grep -rl -e '--lane' internal/cli` | only todo/gtd command files and their tests (`--lane <label>` is a subcommand-local flag there) | 0 |
| PV-9 | bounded search for a default-backend setting (`grep -rIn 'default_backend\|DefaultBackend\|launch\.yaml' --include='*.go' internal`, tests filtered with a pipe) | only the last-used-profile ledger hits; no backend key. Moot: the verb carries the backend. | 0 |
| PV-10 | local-versus-template markdown pairs compared with `cmp -s` | `kanban-dispatch-mechanics.md` and `kanban-dispatch-detail.md` identical; `kanban-dispatch.md`, `orchestration-mode-selection.md`, `cross-session-messaging-detail.md`, `manager-lead.md` differ | n/a |
| PV-11 | `go build ./cmd/moai` | exit 0 (build log empty) | 0 |
| PV-12 | `go test ./internal/hook -run '^(TestUnbindNoticeRebindLinePresence\|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide\|TestFactoryGuideNamesWorkerJoinInEveryLocale)$' -v -count=1` | three PASS results (swept count 3), `ok github.com/modu-ai/moai-adk/internal/hook 2.303s` | 0 |
| PV-13 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` (tree-built binary) | see PV-46 for the v0.4.0 run | 0 |
| PV-14 | mutation probe on a scratchpad COPY of the SPEC directory: rewrite every `REQ-005` reference in acceptance.md to `REQ-0XX`, then lint | still `No findings` — the lint does not move when a requirement loses all coverage (measured on v0.1.0, v0.2.0, v0.3.0; re-measured in PV-46) | 0 |
| PV-15 | by hand: `grep -c "REQ-<n>" acceptance.md` per requirement | see PV-47 for the v0.4.0 count | 0 |
| PV-16 | `go run ./cmd/moai cc -f 3`; `moai glm -f 3`; `moai codex -f 3`; `moai codex -f lane-2`; `moai codex -f` | `-f <N>` is ALREADY refused on all three verbs: exit 1, stdout empty, one line, nothing launched. cc/glm line: `-F/--Factory takes no argument (the factory leader), the role token -f lane, which joins this session to a running factory as the next free lane, or a lane label (e.g. -f lane-2) that launches exactly that one lane, got "3".` codex line (all three shapes): `FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader` | 1 each |
| PV-17 | `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata\|TestGLM_FactoryLeadRunIsJoinableByLane)$' -v -count=1` | two PASS results, `ok ... 2.134s`; the first launches `-f` (leader) only; the second launches `-f` then `-f lane-3`, so it is re-pinned | 0 |
| PV-18 | `grep -rlE` of removed-lane-form patterns over AGENTS.md, READMEs, docs-site/content, .claude/agents, .claude/rules, internal/template/templates | 29 paths (the v0.2.0 inventory; superseded by PV-45 for the Tier L scope) | 0 |
| PV-19 | `go run ./cmd/moai cc -l`; `moai codex -l`; `moai cc -l lane-2`; `moai cc -f -l` | stdout empty, exit 1: `--Leader requires a leader label.` (cc, codex, and `cc -f -l`); `--Leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target.` (`cc -l lane-2`) | 1 each |
| PV-20 | `grep -c -e '-f lane'` over `cc.go`, `glm.go`, `codex_launcher.go`; over `session_start_factory.go`, `session_start_factory_i18n.go`, `factory.go`, `factory_card.go`; and `grep -c -e 'cc -f lane' internal/hook/session_stale_run.go` | cc 9, glm 8, codex_launcher 10; session_start_factory 3, session_start_factory_i18n 12, factory 21, factory_card 2; stale_run 4 | 0 |
| PV-21 | `grep -rIlE` of the removed-lane-form pattern over the whole tree excluding `.git`, worktrees, reports, node_modules, specs, testdata | 136 paths: `.moai/specs` 68, docs/instruction 29, Go 14, tests 18, frozen fixtures 3, CHANGELOG and generated codemaps 4 | 0 |
| PV-22 | `go test ./internal/cli -run '^(TestCC_KanbanFlagStrippedBeforeLaunch\|TestGLM_KanbanFlagParity\|TestPrepareKanbanSettingsWritesTransientFile)$' -v -count=1` | three PASS results, `ok ... 0.921s` | 0 |
| PV-23 | `go test ./internal/template -run '^(TestWorkflowRulePathsPinned\|TestContractModeAlwaysLoadedBudget\|TestContractModeConstitutionDriftNotIncreased\|TestContractModeLocalTemplateParity\|TestRuleTemplateMirrorDrift\|TestDeclaredRuleMirrorForks\|TestAllSkillsInCatalog\|TestCatalogHashCoversSkillSubfiles)$' -v -count=1` | six PASS, TWO SKIP (`TestContractModeConstitutionDriftNotIncreased`, `TestContractModeAlwaysLoadedBudget`; reason unobserved — a skip is not a pass), `ok ... 0.254s` | 0 |
| PV-24 | `grep -rlE` per marker constant over `internal` and `cmd`, non-test files | `MOAI_KANBAN_ID` 22; `_LEAD_ADDR` 9; `_LEAD_NAME` 8; `_SETTINGS_INJECTED` 6; `_BACKEND` 12; `_CARD` 7; `MOAI_KANBAN` 8; `_SPEC` 6; `_LABEL` 8. Union of the six factory-read markers by name or constant: 95 files (34 non-test) — SUPERSEDED in v0.8.0 by PV-87: 92 files (31 production, 61 test). `internal/discovery` reads `MOAI_KANBAN_ID` from a live leader process's environment. | 0 |
| PV-25 | `grep -rln` of the board-family symbols outside `internal/kanban`, non-test | only `todo_autodone.go:306` and `todo_landed.go:337`, both `ReadPrimarySpecStatus` (not board); the board state store has no non-test caller outside the package | 0 |
| PV-26 | `moai constitution validate` (tree-built binary) | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`; `4 retired entry/entries skipped` | 0 |
| PV-27 | `grep -c` over `zone-registry.md` for `orchestrator-class\|kanban companion\|Selection Decision Tree\|manager-lead` and `grep -n -i 'companion\|lane\|factory\|foreman\|kanban'`; `grep -c 'ZONE:Frozen'` over the three kanban rules | 0 hits in the registry for each; 0 `[ZONE:Frozen]` in `kanban-dispatch.md`, `-detail.md`, `-mechanics.md` | 0 |
| PV-28 | `moai constitution guard --help`; `moai constitution amend --help` | `guard` takes `--violations` (rule IDs), not a file diff; `amend` takes `--rule CONST-V3R2-NNN` (required), `--before`, `--after`, `--evidence`, `--dry-run` | 0 |
| PV-29 | `grep -rlE -i 'kanban (mode\|companion\|chain\|board\|foreman)\|moai (cc\|glm) -k\|-k --name' .claude internal/template/templates AGENTS.md CLAUDE.md` | 25 paths (v0.3.0 inventory; superseded by PV-45) | 0 |
| PV-30 | `grep -rlE -i 'kanban mode\|moai (cc\|glm) -k\|(-f\|--factory)[ =]lane\|-l, --leader\|codex +-f'` over the READMEs, `docs-site/{content,data,i18n,layouts}`, `AGENTS.md` | 50 paths (v0.3.0 inventory; superseded by PV-45) | 0 |
| PV-31 | `grep -rlE -e '-k --name\|moai (cc\|glm) -k\|(-f\|--factory)[ =]lane\|-l, --leader' internal cmd --include='*.go' --exclude='*_test.go' --exclude-dir=testdata` | 16 paths: `internal/config/{envkeys,defaults}.go`; `internal/cli/{factory_card,glm,cc,factory,factory_lane_relaunch,codex_launcher,codex_factory,kanban}.go`; `internal/kanban/bootstrap.go`; `internal/hook/{session_start_factory_i18n,session_start_factory,session_stale_run,session_start_kanban,session_start_kanban_i18n}.go` — identical when widened from four packages to all of `internal` and `cmd` | 0 |
| PV-32 | `grep -rlE` of the kanban chain and companion identifiers over `internal` and `cmd`, `--exclude='*_test.go'` | 15 paths (listed in acceptance.md RED-K3) | 0 |
| PV-33 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict`, v0.3.0 text | `No findings`; mutant (every `REQ-005` in acceptance.md rewritten) also `No findings` — superseded by PV-46 | 0 / 0 |
| PV-34 | by hand, v0.3.0 | superseded by PV-47 | 0 |
| PV-35 | counts: `grep -rlE '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go' \| wc -l` (and the same with `--include='*_test.go'`); `grep -rIn 'kanban\.' internal cmd --include='*.go' \| wc -l` | 179 importing files, of which 113 test (so 66 production, one under `cmd`); 1,689 qualified references | 0 |
| PV-36 | `go run ./cmd/moai codex -k` | stdout empty, exit 1, stderr `KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead` | 1 |
| PV-37 | `grep -c`/`ls` single-cell checks: `grep -c 'Chain session board' internal/web/screens.templ`; `grep -c 'moai-kanban-foreman' internal/template/catalog.yaml`; `ls .claude/loop.md .claude/skills/moai-kanban-foreman/SKILL.md`; `grep -n 'kanbanBootstrapNotice' internal/hook/session_start.go` | `1`; `2`; both files exist (1254 and 11181 bytes); lines `587` and `599` | 0 |
| PV-38 | `moai constitution validate` re-run; `grep -n -i kanban CLAUDE.md .claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/agent-common-protocol.md`; `grep -n 'agent-common-protocol' .claude/rules/moai/core/zone-registry.md`; read of `zone-registry.md:370-469` | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`, 4 retired skipped, exit 0. The four sentences: `CLAUDE.md:61`, `moai-constitution.md:11`, `agent-common-protocol.md:27` and `:75` (line 27 also cites `kanban-dispatch-mechanics.md` and `kanban-dispatch.md`). Registered `agent-common-protocol.md` entries: `CONST-V3R2-036..038` (Frozen, anchor `#user-interaction-boundary`, three short clauses), `-039` (Evolvable, `#language-handling`, clause = the header sentence "All agents receive and respond in user's configured conversation_language."), `-040..-046`. Line 27 is in the `#user-interaction-boundary` section; line 75 in the `#language-handling` body. None of the four sentences' text is a registered clause string. | 0 |
| PV-39 | `go test ./internal/hook -run '^(TestFactoryLeadNoticeWorkerCountDrivesLineCount\|TestFactoryLeadNoticeAllSlotsClaimed)$' -v -count=1` | two PASS results (swept count 2), `ok github.com/modu-ai/moai-adk/internal/hook 1.639s` — the two tests the leader-notice change replaces or deletes | 0 |
| PV-40 | `grep -c 'EnvMoaiKanbanLabel'` over `codex_launcher.go factory_m5_test.go codex_factory_retire_test.go codex_debug_trace_test.go factory_m4_test.go envkeys.go`; read of `factory_m5_test.go:51-158` and `factory_card.go:50-100` | launcher 4, `factory_m5_test.go` 2, `codex_factory_retire_test.go` 1, `codex_debug_trace_test.go` 1, `factory_m4_test.go` 1, `envkeys.go` 4. `TestSD_AC003_CodexRelaunchPerCard` asserts the label PRESENT (`:145-147`) under a comment that the card verbs read it; `factory_card.go:62,88-91` read `MOAI_FACTORY_WORKER`. No production reader of the label found. | 0 |
| PV-41 | `grep -rnE 'VerifyRung\|DeepScanDir\|VerifyReentries\|\.Rung\b' internal --include='*.go' --exclude='*_test.go'`; `grep -rn 'verify exit gate\|VerifyExitGate\|Verify Exit Gate' internal --include='*.go'`; `grep -rn 'factory_chain' .claude internal/template/templates internal docs-site/content` (md/go/yaml); `grep -c factory_chain`, `grep -c '.moai/state/factory'`, `grep -c 'chain head'`; `grep -rn -A12 'func RecordPath' internal/kanban` | chain fields: only `internal/kanban/record.go` lines (declarations and comments), no writer or reader elsewhere; verify-gate in Go: no output; `factory_chain`: `factory.md:64,66` in the local copy and the template copy only (2 lines per file); `factory.md` `.moai/state/factory` count 1; `moai.md` `chain head` count 1; `RecordPath` = `filepath.Join(RuntimeStateDirForRoot(projectRoot), sessionID+".json")` (`record.go:189-191`) | 0 / 1 / 0 |
| PV-42 | `grep -c "'factory\.\|\"factory\.\|factory\.[a-zA-Z]*:" internal/web/assets/i18n.js`; `grep -n 'case "factory"\|case "kanban"' internal/web/icons_templ.go internal/web/icons.templ` | `0` existing factory i18n keys (so the key-family rename cannot collide); icon switch has only `case "kanban"` (`icons_templ.go:128`, `icons.templ:72`) | 1 / 0 |
| PV-43 | `go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader\|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1` | two PASS results (swept count 2), `ok github.com/modu-ai/moai-adk/internal/discovery 0.555s` | 0 |
| PV-44 | `grep -rlE '\b(LoadBoard\|WriteBoardState\|AcquireBoardLock\|RecoverBoard\|ParseColumn\|TransitionIntoRun\|BoardState)\b' internal cmd --include='*.go'`; the same symbols in `integration_lock_mutation.go`, `status_read_test.go`, `admission_test.go`; `find internal -name 'board_store*.go' -o -name 'board_lock*.go' -o -name 'board_recover*.go' -o -name board.go -o -name column.go -o -name reconcile.go` | 24 paths, all under `internal/kanban` (verbatim in acceptance.md RED-K4): the 10 board files, the board tests, six further tests (`admission_test.go`, `status_read_test.go:324`, `fix2_probe_test.go`, `fix3_wedge_test.go`, `f1_traversal_test.go`, `kanban_helper_test.go`), and a comment at `integration_lock_mutation.go:24`; the `find` lists 18 paths (10 production, 8 tests) | 0 |
| PV-45 | counts, each one command piped to `wc -l`: non-test Go/templ/js carrying the word (plain `grep -rliE kanban ...`): 190; same under the AC-018 pattern: 189; test files under that pattern: 308 (310 plain); `find internal cmd -iname '*kanban*'`: 22; docs-scope files under the plain Latin word: 151; under the combined pattern with `claude-code` excluded: 156; Korean/Japanese/Chinese spellings in non-test source: 8 files (`internal/web/{app.go,screens.templ,screens_templ.go,viewmodel_ops.go,screens.go,assets/i18n.js}`, `internal/hook/{session_start_factory_i18n,session_start_kanban_i18n}.go`) | counts as listed; the mirrored `claude-code/` docs contain no Latin occurrence and two generic native-language uses (a Chinese "monitoring dashboard", a Korean Agent Teams analogy) | 0 |
| PV-46 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` on the v0.4.0 text (tree-built binary, scratchpad); then on a scratchpad COPY (`scratchpad/probe4/SPEC-LAUNCHER-ENTRY-FLAGS-001`) with every `REQ-005` in acceptance.md rewritten to `REQ-0XX` (`grep -c 'REQ-005'` on the copy: 0) | original: `✓ No findings — all SPEC documents are valid`, exit 0; mutant: the same `✓ No findings`, exit 0. The lint does NOT move when a requirement loses all coverage, so its green is not coverage evidence (PV-14, now measured on four revisions). Also run: `[[ "SPEC-LAUNCHER-ENTRY-FLAGS-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS \|\| echo FAIL` → `PASS`; spec frontmatter `id`, `version: "0.4.0"`, `status: draft`, `phase: "v3.2.0 target"`, `tier: L`; the four sibling artifacts carry no `status:` field (counts 0 each) | 0 / 0 |
| PV-46b | v0.5.0 re-run after the Q18 revision: `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict`; then a scratchpad copy (`scratchpad/probe5-copy`) with every `REQ-019` in acceptance.md rewritten to `REQ-0XX`, linted with `--strict` | original: `✓ No findings — all SPEC documents are valid`, exit 0; mutant (the requirement that carries the Q18 change loses all coverage): the same `✓ No findings`, exit 0 — the lint is again not coverage evidence. Hand count re-run (PV-47): 25 criteria, 25 `Verifies` lines, 25 `Command` lines, 24 requirement lines, every REQ-001..REQ-024 on a `Verifies` line (REQ-001 twice, REQ-022 with REQ-023, REQ-023 again). Ceilings respected: 24 of 25 requirements, 25 of 25 criteria — the Q18 change folded into REQ-019 and AC-020 and the AC-022 stop condition, adding no REQ or AC | 0 / 0 |
| PV-47 | by hand: `grep -c '^### AC-' acceptance.md`; `grep -c '^- Verifies' acceptance.md`; `grep -c '^- Command:' acceptance.md`; `grep -o '^- Verifies REQ-[0-9]*\(, REQ-[0-9]*\)\?' acceptance.md`; `grep -c '^- REQ-' spec.md` | 25 criteria; 25 `Verifies` lines; 25 `Command` lines; 24 requirement lines. Every requirement REQ-001..REQ-024 appears on a `Verifies` line: REQ-001 twice (AC-001 cc/glm, AC-002 codex), REQ-022 with REQ-023 on AC-023, REQ-023 again on AC-024, every other requirement once on AC-003..AC-021, AC-022, AC-025 (AC→REQ map in `spec.md` §C matches). Ceilings: 24 of 25 requirements, 25 of 25 criteria. | 0 |
| PV-48 | `grep -n 'Env.* = "MOAI_KANBAN' internal/config/envkeys.go`; `grep -rn '"MOAI_KANBAN' internal cmd --include='*.go' --exclude='*_test.go'` | literals at `envkeys.go:182` `MOAI_KANBAN`, `:187` `_SPEC`, `:195` `_ID`, `:205` `_LABEL`, `:216` `_SETTINGS_INJECTED`, `:225` `_LEAD_ADDR`, `:238` `_BACKEND`, `:253` `_CARD`, `:273` `_LEAD_NAME`; and `internal/codexwiring/configtoml.go:21` (the MCP env allowlist naming `MOAI_KANBAN_ID` and `MOAI_KANBAN_BACKEND`); no other non-test string literal carries a marker name | 0 |
| PV-49 | `grep -n -i 'kanban\|factory-mode' docs-site/vercel.json`; read of `vercel.json:170-212`; `grep -n -i kanban docs-site/data/menu/main.yaml docs-site/layouts/index.html docs-site/i18n/en.yaml`; `grep -rn -i kanban docs-site/content/en/{advanced,multi-llm,core-concepts}/_meta.yaml`; `ls docs-site/content/*/advanced/factory-mode.md` | rules `:183-201` redirect `/advanced/factory-mode` and `/multi-llm/factory-mode` (locale and bare forms) to the kanban pages; `:202-211` `manager-kanban` → `manager-lead`; menu `main.yaml:159-162,493-496,725-728`; banner `layouts/index.html:56-76` and `i18n/en.yaml:83-85`; `_meta.yaml` listings; `advanced/factory-mode.md` exists in en (11,173 B), ko, ja, zh — so the factory-mode page is shadowed by the redirect today | 0 |
| PV-50 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0, no output | 0 |
| PV-51 | `grep -n '/loop' internal/hook/session_start_factory_i18n.go`; `grep -c` of notice string-table fields | the queue sentence at `:89` (en: "the kanban foreman loop (bare `/loop`)"), `:139` (ko), `:184` (ja), `:229` (zh); 34 references to `leaderFreeSlots`, `leaderSlotsNone`, `laneJoin`, `laneJoinNoCount`, `leaderManual`, `entryGuide` (this row first read 28; that figure did not reproduce and PV-68 measured 34); `session_start_factory.go:229` builds `FactoryFreeSlots`, `:240` `leaderFreeSlots` | 0 |
| PV-52 | `wc -c` of the rule files; read of `update_archive.go:25-70`; bounded search `grep -rniE 'legacyRule\|obsolete\|removedFiles\|deprecatedFiles\|staleRule\|retiredRule\|legacyRuleFiles' internal/cli internal/template --include='*.go' --exclude='*_test.go' -l` | `kanban-dispatch.md` 26,352 B local and 26,030 B template, `-detail.md` 39,965 B, `-mechanics.md` 18,909 B; `legacySkillIDs` archives retired skills to `.moai/archive/skills/v2.16/` and its guard test `TestLegacySkillIDsNotEmbedded` keeps the list disjoint from the embedded skills; the rule-file search found only `internal/cli/memory.go` (unrelated) — **SUPERSEDED in v0.5.0 by PV-54..PV-59: that search used the wrong terms and missed the managed-root clean, so the conclusion "no stale-rule cleanup exists" was wrong** | 0 |
| PV-53 | `grep -rn '"kanban' internal cmd --include='*.go' --exclude='*_test.go'`; `find internal cmd docs-site/content -iname '*kanban*'`; `grep -rnwE 'factory( :=\|,\| =)' internal cmd --include='*.go'` | the string-literal lines beginning with "kanban" (the generated web templates, the web events and screens, error texts in `internal/kanban`, two hook timing laps, a coverage key — summarized in research.md §R3; not counted); the file/directory names listed in RED-N2 plus 12 docs pages; four non-LSP/TUI local identifiers named `factory` (`factory_handoff_recover.go:20`, `role_naming_m3_notice_test.go:129`, `stale_run_m1_test.go:118`, `runtime_census_test.go:75`) | 0 |

| PV-54 | `grep -rliE 'namespace.protect\|user-owned namespace\|protected namespace\|namespace protection' internal .claude/rules .moai/docs CLAUDE.md AGENTS.md --include='*.go' --include='*.md'`; read of `update_namespace_protect.go`, `update_destructive_registry.go`, `update_cleanup.go:1-452`, `update_residue_cleanup.go:1-160`, `v2_detection.go:1-160`, `deploy.go:60-210`, `update_template_sync.go:360-470`, `update_archive.go:25-140,285-350`, `plan.go:152-259`; `grep -rn 'archiveLegacySkills(\|removeDeprecatedFile\|scanDeprecatedPaths(' internal/cli` | the namespace-protection contract (`SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001`): `IsUserOwnedNamespace` classifies harness and user skills/agents; `backupUserOwnedNamespace` to `.moai/backups/update-<ISO>/`; `assertNoUserOwnedNamespaceTouch` sentinel `UPDATE_USER_NAMESPACE_VIOLATION`. The managed targets include `.claude/rules/moai` (`deploy.go:75-78`); the clean is a template-sync stage (`update_template_sync.go:388-421`); `legacySkillIDs` archive runs inside that stage before the removal and its failure only warns (`:402-407`). `defs.DeprecatedPaths` is acted on only by clean-reinstall and the v3 residue sweep | 0 |
| PV-55 | scratch probe (a `_test.go` in `internal/cli` created for the run and removed after it; `git status --short` afterwards: only the SPEC directory untracked): `t.TempDir()` project with `.claude/rules/moai/workflow/kanban-dispatch.md` ("USER MODIFIED carried") and `.claude/rules/moai/workflow/zz-retired-probe.md` ("USER MODIFIED not carried"); `deploy.CleanMoaiManagedPaths(root, &out, template.EmbeddedTemplates())`; `go test ./internal/cli -run '^TestZZT1399Probe$' -v -count=1` | PASS; live tree: both files gone; backup files: `[.moai-backups/20261002_121511/pre-clean/.claude/rules/moai/workflow/zz-retired-probe.md]` only; output: `✓ Removed .claude/rules/moai (backed up 1 unmanaged file(s))`; so a user-modified template-carried rule is overwritten without a backup today, while a file the template does not carry is backed up. Probe essentials (to recreate it): write the two files, build `tmplFS` with `template.EmbeddedTemplates()`, call the clean, log `os.Stat` of both paths and the files under `.moai-backups` | 0 |
| PV-56 | `go test ./internal/cli/update/deploy -run '^(TestCleanMoaiManagedPaths_BackupsUnmanagedFiles\|TestCleanMoaiManagedPaths_BackupFailureAbortsRemoval)$' -v -count=1`; read of `deploy_preclean_backup_test.go:40-101` | two PASS results (swept count 2), `ok .../internal/cli/update/deploy 0.549s`; the first asserts the unmanaged file reaches the backup byte-identical, the managed file does not, the root is removed, and the output contains "backed up 1 unmanaged file" | 0 |
| PV-57 | read of `update.go:480-500`, `update_residue_cleanup.go:50-160`, `v2_detection.go:82-147`, `defs/dirs.go:25-68`; `grep -n 'DeprecatedPaths'` | the v3 residue branch `return cleanupErr` aborts the update; its sweep deletes after backup regardless of modification (`os.RemoveAll`); a registered path that exists sets `V2DetectedViaDeprecatedPath`, and `IsV2 = !V3VersionConfirmed && (...)`; the registry is a 40-entry slice guarded by `dirs_test.go` and `TestDeprecatedPaths_NoTemplateCollision` | 0 |
| PV-58 | `git log --oneline -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md \| wc -l`; `git tag --list 'v3*' \| wc -l`; `wc -c` of the rule files; `grep -c -e 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' -e 'preload `AskUserQuestion` via `ToolSearch`' .claude/rules/moai/core/agent-common-protocol.md` | 65 revisions; 18 v3 tags; sizes in PV-52; the registered Frozen clause strings (`Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.` and ``preload `AskUserQuestion` via `ToolSearch` ``) match 2 lines (CONST-V3R2-038's string is a substring of the first) | 0 |
| PV-59 | `/private/tmp/.../moai-t1399 constitution validate` re-run in the final revision; `git status --short` | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`, exit 0; only the SPEC directory untracked | 0 |

Rows PV-60 to PV-72 are the v0.7.0 re-measurements of the plan-audit iteration 1 findings. They were taken in this revision against HEAD `b9242da00ce489c4f26efb5f6d447ccafef08c54`, whose content outside this SPEC directory equals tree `a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2` (PV-71). The binary rows used `scratchpad/moai-t1399`, built from `a6d3e6fd4` earlier in this plan run (its `version` prints no commit; the tree content it was built from is the code content of HEAD, PV-71); no measurement here was replaced by reading source because of a refusal — one compound shell command was refused by the worktree guard and re-run as plain commands (Gaps).

| ID | Command | Observed | Exit |
|----|---------|----------|------|
| PV-60 | `go test -overlay <scratch>/probe7/overlay.json ./internal/cli -run '^TestZZT1399EntryProbe$' -v -count=1` — a scratch test (source essentials: for each of 18 cc/glm argument lists call `parseLauncherEntry(args)`, then `parseFactoryLaneLabel(entry.Rest)` and `resolveFactoryBranch(entry.FactoryEnabled, isLane)` and print one `CCPARSE` line; for 6 codex lists call `runCodex(&cobra.Command{Use: "codex"}, args)` with captured stderr and print one `CODEX` line; the overlay maps `internal/cli/zz_t1399_probe_test.go` to the scratch file, so nothing is written into the tree; `git status --short` empty afterwards) | `-f --name lane-2`, `-f -n lane-2`, `--factory --name=lane-2` → `factoryEnabled=true laneLabel="lane-2" isLane=true branch=2`; `-f --name leader-r7` → `isLane=false branch=1`; `--lane`, `--lane lane-2`, `--lane=lane-2`, `--lane 3`, `--lane leader-2` → `factoryEnabled=false`, the tokens left in `rest`, `branch=0`, no error; `--lane -f` → `factoryEnabled=true rest=["--lane"] branch=1`; `-l` → parse error `--leader requires a leader label`; `-l lane-2`, `-l 3`, `-l=lane-2`, `-l leader-2` → `--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target`; `-l -f`, `-l -k`, `-l --name lane-2` → `--leader requires a leader label`; codex `--lane`, `--lane lane-2`, `--lane=lane-2`, `--lane 3`, `--lane leader-2` → exit 1, stderr `unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane\|lane-<n>]] [--factory-run <id>] [-- codex-args...] \| moai codex status \| moai codex app`; codex `--lane -f` → exit 1, `FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; …`; `ok … 0.857s` | 0 |
| PV-61 | `grep -n '^func \|^type \|^const \|^var ' internal/cli/kanban.go`; `grep -rn 'exportKanbanLaunchFacts\|exportFactoryLaunchFacts' internal cmd --include='*.go'`; `grep -rnE '\b(parseKanbanFlag\|enterKanbanMode\|…)\b'` over the nine kanban-only declarations and the three constants (non-test); the same for the factory-shared helpers; read of `kanban.go:500-575`, `factory.go:325-410`, `glm.go` and `cc.go` branch structure | 25 funcs, 2 types, 2 single constants, 3 constant blocks: 27 func/type declarations, 9 kanban-only (`parseKanbanFlag`, `enterKanbanMode`, `enterKanbanCompanionMode`, `companionRegistryPath`, `resolveCompanionName`, `parseCompanionLabel`, `rejectKanbanOnCG`, `kanbanBranch`, `resolveKanbanBranch`) and 18 factory-shared. `exportKanbanLaunchFacts` callers: `glm.go:271` (factory leader branch, `case factoryBranchLeader` at :256), `glm.go:304` (factory lane branch, :282), `glm.go:334,352` and `cc.go:300,320` (kanban branches), `cc.go:222,260` and `codex_launcher.go:951`, `codex_factory.go:144` through `exportFactoryLaunchFacts` (`kanban.go:565-567`), tests `kanban_launch_facts_test.go:48,70,97`. `parseLauncherEntry` refuses `-k` with `-f` (`factory.go:363-366`), so `entry.Spec` is empty on every factory path. The factory-shared helpers have factory callers directly (`factory.go:311,315,379,391,653,677,680,732,733`; `cc.go:211,224,225`; `glm.go:257,273,274`) or through `appendLeaderName` and `resolveLeaderName` (`resolveLeaderName`, `noteLegacyLeaderRegistryEntries`, `claimName`, `leaderRegistryPath`, `leaderNameArgs`, `allDigits`) | 0 |
| PV-62 | `grep -rn 'EnvMoaiKanbanLabel\|EnvMoaiKanbanSpec\|EnvMoaiKanban\b' internal cmd --include='*.go' --exclude='*_test.go'`; the same with `--include='*_test.go'`; `grep -rlE` with `\| wc -l` for both | non-test readers (lines in research.md §R16): `envkeys.go` definitions and comments, `launcher_blockcap_infinite.go:57`, `codex_launcher.go:317,319,320,959,960,1037`, `kanban.go:194-204,321,334,343,545,553`, `ptycaptest/harness.go:67,68,70`, `session_start_kanban.go:56,59,183`, `session_start_record.go:110,191,198`, comments `factory.go:665,699`; counts: 8 non-test files and 24 test files (13 `internal/cli`, 11 `internal/hook`, names in research.md §R16) | 0 |
| PV-63 | `grep -rnE '\b(CompanionRoles\|CompanionLauncher\|companionLaunchers\|CompanionLabel\|CompanionNumberLabel\|SplitCompanionLabel\|isCompanionRole\|LeaderSocketPath\|LeaderLabel\|SplitLeaderLabel)\b' internal cmd --include='*.go' --exclude='*_test.go'` minus `bootstrap.go`; `grep -rnE '\b(kanbanLeaderNotice\|kanbanCompanionNotice\|kanbanBootstrapNotice\|kanbanBootstrapNoticeForSource\|kanbanRoleFromEnv)\b'`; `grep -rhoE 'kanban\.[A-Za-z]+' internal/web internal/statusline internal/cli/doctor*.go` (non-test); `grep -rn 'kanban-dispatch\|moai-kanban-foreman' internal cmd --include='*.go' --exclude='*_test.go'`; `grep -rnoE` of the RED-K3 identifiers per file; read of `record.go:125-175`, `bootstrap.go:236-395`, `role.go` | companion-symbol readers: `cli/kanban.go:211,380-385,590`, `hook/session_start_kanban.go:131-151,233`, `hook/session_start_record.go:192`, `kanban/record.go:151` (`WithRole` → `isCompanionRole`), comment `role.go:61`; leader-label readers kept: `factory_discovery.go:167`, `factory.go:177,313,320,515`, `codex_factory.go:60`, `session_start_factory.go:196`, `kanban.go:240,422`; hook notice builders: `session_start.go:587,599` and comments; web/statusline/doctor use `RecordPath`, `Record`, `ReadAll`, `LoadFactoryRegistry`, backlog and landing symbols only; 13 comment lines in 11 non-test Go files name `kanban-dispatch.md` or `moai-kanban-foreman`; the RED-K3 per-file token map in research.md §R16 | 0 |
| PV-64 | `git ls-files AGENTS.local.md CLAUDE.local.md AGENTS.md CLAUDE.md`; `git check-ignore -v CLAUDE.local.md`; `grep -n -i 'kanban\|-f lane\|moai cc -k\|moai cc -f' AGENTS.md AGENTS.local.md`; `grep -n 'AGENTS.local' CLAUDE.md`; `grep -rn -i 'kanban\|-f lane' internal/template/templates/AGENTS.md.tmpl`; the three RED-D5, RED-D7, and RED-D1 commands (acceptance.md ledger) | `AGENTS.local.md`, `AGENTS.md`, `CLAUDE.md` tracked; `CLAUDE.local.md` ignored (`.gitignore:276`) and absent from the worktree; `AGENTS.md:125` "**Codex factory lanes (`-f lane`)**" (no word); `AGENTS.local.md:217` "Kanban(`moai cc -k`) / Factory(`moai cc -f N`) …"; `CLAUDE.md:167,172` import `@AGENTS.local.md`; `AGENTS.md.tmpl` no match (exit 1); removed-form grep 46 files (AGENTS.md and AGENTS.local.md among them); word grep with `AGENTS.local.md` added 157 files (`AGENTS.local.md` listed once); `find` outside `internal` and `cmd`: 19 paths (3 image files, 1 skill directory, 12 docs pages, 3 rules) | 0 |
| PV-65 | `env MOAI_GR_BASE=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 go test ./internal/template -run '^TestContract' -v -count=1` (output to a scratch file); read of `contract_mode_guided_test.go:1-60,640-740` | `TestContractModeConstitutionDriftNotIncreased` PASS (6.21 s, base and current non-OK pairs `[]`), `TestContractModeAlwaysLoadedBudget` PASS (0.51 s, +0 characters); `TestContractModeGuidedPreservation` FAIL (text outside the contract blocks differs from the base in several skill documents; the first mismatch printed is `internal/template/templates/.claude/skills/moai/SKILL.md` at line 169; "stripped 40 blocks across 26 copies"), `TestContractModeChangeSetAllowlist` FAIL (`changed path outside the allowlist: .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/acceptance.md`), `TestContractModeEmitterSites` FAIL (unclassified Kickoff documents at the base: `moai-mcp-tools-catalogue.md`, `auto-semantics.md`, `contract-autonomy.md` and their template copies); the file header says a test skips when `MOAI_GR_BASE` is unset and an acceptance run must set it; `grKickoffClasses` (`:652-691`) is keyed by base-ref paths and `kanban-dispatch.md` is at `:679` | 1 (the three FAILs; unrelated to this SPEC) |
| PV-66 | `hugo --source docs-site --minify --gc --destination <scratch>/public` (output to a scratch file); `sh <scratch>/probe7/parity.sh` (the `.claude/skills/hns-oss-docs-verify/SKILL.md` §4 ratchet with scratch outputs); `grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md`; `git status --short` | Hugo exit 0, summary table ends `Total in 3099 ms`, page counts 189, 187, 187, 187, `grep -c -i -e WARN -e ERROR` 0; ratchet: 51 divergent pages, 51 baseline lines, `comm -23` empty, `comm -13` empty, 155 ko pages compared; README H2 counts 12, 12, 12, 12; the tree unchanged | 0 |
| PV-67 | `grep -n '^status:\|^id:'` over `SPEC-KANBAN-BOOTSTRAP-001`, `SPEC-KANBAN-WORKTREE-001`, `SPEC-WEB-TODO-QUEUE-001`, `SPEC-KANBAN-BOARD-001`, `SPEC-KANBAN-RENAME-001`; `grep -rn 'REQ-TOSQ-018' .moai/specs/SPEC-TODO-SQLITE-001`; `grep -rln 'TOSQ-018\|literal-cleanliness' internal --include='*.go'`; read of `state_dir.go:15-25`, `events.go:30-36`, `backlog_sqlite_test.go` | BOOTSTRAP-001 and WORKTREE-001 `draft`; WEB-TODO-QUEUE-001, KANBAN-BOARD-001, KANBAN-RENAME-001 `completed`; REQ-TOSQ-018 = "no active template source under … `state/kanban`" with an allowlist for the intentional old-layout fallback reader; no Go test enforces the sweep (the only Go hits are the comment in `state_dir.go:19-22` and an unrelated `AC-TOSQ-018` label in `backlog_sqlite_test.go`); `events.go:33-34` "SSE event KEY stays "kanban" — it is a frontend-visible contract" | 0 |
| PV-68 | `git diff --quiet a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- .claude/rules/moai/core/zone-registry.md`; the same over `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/spec.md`; `grep -cE '\b(leaderFreeSlots\|leaderSlotsNone\|laneJoin\|laneJoinNoCount\|leaderManual\|entryGuide)\b' internal/hook/session_start_factory_i18n.go`; `grep -n -B1 -A4 '^paths:'` over the three rules; read of `moai-constitution.md:9-13`, `zone-registry.md:281-300`; `moai constitution validate --help` (audit report E-9) | zone-registry exit 0; spec.md control exit 1; i18n references 34; the three `paths:` globs name `**/kanban-dispatch*.md` at `kanban-dispatch-detail.md:3`, `kanban-dispatch-mechanics.md:3`, `cross-session-messaging-detail.md:3`; `moai-constitution.md:11` is the kanban sentence and `:12` the Frozen `[ZONE:Frozen] [HARD] AskUserQuestion is the sole user-facing question channel` bullet, registered as `CONST-V3R2-025` (`#moai-orchestrator`, `zone-registry.md:283-289`) | 0 / 1 |
| PV-69 | `grep -n -e '-f lane' internal/hook/session_start_factory.go internal/hook/session_stale_run.go internal/cli/factory_card.go`; `grep -c -e '-f lane' internal/hook/session_start_factory_i18n.go`; `grep -c -i kanban internal/cli/update_archive.go internal/kanban/state_dir.go`; the widened AC-024 pattern over non-test Go (`-k --name\|moai (cc\|glm) -k\|(-f\|--factory)[ =]lane\|-l, --leader\|(-f\|--factory)[ =](<N>\|N\b\|[0-9])`) | stale-run hint `session_stale_run.go:92,103,114,125` (`'moai cc -f lane-<n>'` in four locales), card errors `factory_card.go:95,98` (`rejoin with -f lane`), notice comments `session_start_factory.go:162,170,258` and the per-lane line at `:207` (RED-7), 12 matches in `session_start_factory_i18n.go`; word-bearing lines: `update_archive.go` 0, `state_dir.go` 2; the widened pattern lists the same 16 files as RED-S1 (the extra `-f <N>` hits are `envkeys.go:278`, `factory.go:620,621,630,631,710,860`, all in files already listed) | 0 |
| PV-70 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` (scratch binary); the plan-auditor traceability verb (`.claude/agents/moai/plan-auditor.md` Group 4, copied verbatim to `<scratch>/probe7/trace.sh`) on `spec.md` and `acceptance.md`; the same two on a scratch COPY (`<scratch>/probe7/mut1`) in which every `REQ-019` outside its definition line is rewritten to `REQ-0XX` in all of `acceptance.md` and `spec.md` (`grep -c 'REQ-019'`: acceptance 0, spec 1); hand counts `grep -c '^### AC-'`, `'^- Verifies'`, `'^- Command:'` in acceptance.md and `grep -c '^- \*\*REQ-'` in spec.md; `grep -o '^- Verifies REQ-…'` | lint original `✓ No findings`, exit 0; verb on the original `COLLECTED: 24 REQ definitions (acceptance input: read)` with no `UNCOVERED` and no `ORPHAN` line; verb on the mutant `COLLECTED: 24`, `ORPHAN: REQ-0` (the rewritten token), `UNCOVERED: REQ-019` — the verb moves when a requirement loses coverage, so its silence on the original is measured, not blind; lint on the mutant still `✓ No findings`, exit 0 — the lint remains silent on lost coverage, so its green is not coverage evidence. A first mutant that rewrote `REQ-019` only in acceptance.md and the §C table left the verb silent, because history lines of spec.md name `REQ-019` beside `AC-020` and the verb counts any line that names both — a reminder that its "mapped" is generous and the hand count stays the evidence. Hand counts: 25 criteria, 25 `Verifies` lines, 25 `Command` lines, 24 requirement lines; `Verifies` lines cover REQ-001 (twice), 002 to 016, 017 with 024, 018, 019, 020, 021, 022 with 023, 023, 022 — every REQ-001..REQ-024 at least once; ceilings 24 of 25 requirements, 25 of 25 criteria | 0 / 0 / 0 |
| PV-71 | `git rev-parse a6d3e6fd4`; `git rev-parse HEAD`; `git diff --quiet a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- . ':!.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001'` | `a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2`; `b9242da00ce489c4f26efb5f6d447ccafef08c54`; exit 0 — no content differs outside this SPEC directory | 0 |
| PV-72 | read of `internal/kanban/bootstrap.go:236-395` (`NextFactoryLaneNumber` at `:370-384`), `internal/kanban/factory_slots.go:100-210`, `internal/cli/factory.go:374-395`; `grep -n 'never backfills' README.md` | the lane label is one past the highest LIVE canonical claim after `PruneFactoryDeadClaims`, 1 when nothing is claimed; the `-f lane` parse computes it at `factory.go:386` and the branch claims it in one IMMEDIATE SQLite transaction (`ClaimFactoryLane`, dead claims removed first); `README.md:80` states the same rule (a dead lane's claim no longer blocks its number; auto-assignment never backfills a gap) | 0 |


Rows PV-73 to PV-90 are the v0.8.0 evidence for plan-audit iteration 2 (FAIL 0.79; findings D29-D45). Every measurement was taken in this revision on HEAD `5445e296caa48e6ab9821afe808eac2e7385e897` (the commit iteration 2 audited), whose content outside this SPEC directory equals tree `a6d3e6fd4` (PV-73), with Go `go1.26.8` (the toolchain the module selects). The compile proof builds only a SCRATCH COPY of the Go tree; `git status --short` after the first probe run showed only the new untracked `probe/` directory. Where a command was refused rather than run, the Gaps paragraph says so.

| ID | Command | Observed | Exit |
|----|---------|----------|------|
| PV-73 | `git rev-parse HEAD`; `git diff --quiet a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- . ':!.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001'`; `go version` in the tree; `diff -rq` of the pristine scratch copy P0 (the tree the stage data was derived from) against the checkout for `internal`, `cmd`, `e2e`, `pkg`, `scripts`, `test`, and `diff -q` for `go.mod` and `go.sum`; baselines on P0: `go build -gcflags=-e ./...`, `GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...`, `go vet ./...`, `GOOS=windows GOARCH=amd64 go vet ./...` | HEAD `5445e296caa48e6ab9821afe808eac2e7385e897`; `git diff --quiet` exit 0 (nothing differs outside this SPEC directory); `go version go1.26.8 darwin/arm64`; every `diff` empty; baselines: host build exit 0 and empty, windows build exit 0 and empty, host vet exit 0 and empty, windows vet exit 1 with exactly three lines — `# github.com/modu-ai/moai-adk/internal/cli/worktree`, `# [github.com/modu-ai/moai-adk/internal/cli/worktree]`, `vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs` — a failure that predates this SPEC and is the baseline of every windows vet block below | 0 / 0 / 0 / 0 / 0 / 0 / 1 (baseline) |
| PV-74 | the committed runner, one invocation: `go run probe.go -src <checkout> -work <scratch>/replay-final -data .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe -from M0 -to M11` (run from the probe directory; vet on) — it copies the Go tree to the scratch directory, applies M2, M3, M4, M5a, M5b, M6a, M6b from `probe/patches/` and M7, M8, M9, M10 programmatically, and after every stage runs host and windows `go build -gcflags=-e ./...` and `go vet ./...` | every stage prints four blocks that are empty with exit 0, except the windows vet block, which prints exactly the baseline of PV-73 (exit 1); the stage table and the verbatim output follow this table. The header comment of `probe.go` was reworded after this run started (no code line changed). | 0 (the runner) |
| PV-75 | negative control for the lock substrate: on the M5b tree of the committed stage data, delete the ten files the v0.7.0 plan listed as the board family (`board.go`, `board_store.go`, `board_recover.go`, `board_lock.go`, `board_lock_unix.go`, `board_lock_windows.go`, `board_lock_clear_unix.go`, `board_lock_clear_windows.go`, `column.go`, `reconcile.go`) and run `go build -gcflags=-e ./...` on the host and on `GOOS=windows` | 25 compile errors on each OS (listings below the table): `backlog_store.go` 6, `integration_lock_mutation.go` 7, `slot_lease.go` 7, `role.go` 1 (`BoardDir`), and 4 in the per-OS mutation files — `integration_lock_mutation_unix.go` and `slot_lease_mutation_unix.go` (`ClearStaleReport`) on the host, `integration_lock_mutation_windows.go` and `slot_lease_mutation_windows.go` (`ClearStaleReport` and `clearStaleLockAtPath`) on windows | 1 / 1 |
| PV-76 | negative control for the locale helpers: on the M5b tree, delete `internal/hook/session_start_lang.go` (the file that holds `langEnglish` and `operatorLang`) and run `go build -gcflags=-e ./internal/hook/` | 8 compile errors (listing below the table): `session_stale_run.go`, `session_start_factory_i18n.go`, `factory_messages.go`, `session_start.go` | 1 |
| PV-77 | negative control for the `-l` short: on the tree of stage M2 (pristine P0 plus `probe/patches/M2.patch`), delete the constant `leadFlagShort` and run `go build -gcflags=-e ./...`; the alias and socket-directory collisions of rows 5 and 6 of design.md §7.1 were observed as redeclaration errors in an interim replay of the same stage data before the corrections entered it (`internal/cli/codex_launcher.go:769:2: codexFactoryRefusalDiag redeclared in this block`; `internal/kanban/bootstrap.go:309:2: factorySocketDir redeclared in this block`) | 4 compile errors, all `internal/cli/codex_factory.go` lines 42, 43, 50, 51: `undefined: leadFlagShort` (the Codex entry parse still reads it until M3); the two interim redeclaration errors as quoted; deleting the nine `kanban.go` declarations alone (interim tree, before the retired-entry file existed) failed with `internal/cli/codex_launcher.go:771:27: undefined: kanbanUnsupportedBackendSentinel` and `:834:17`, `:834:45`, `:835:29`, `:835:78` `undefined: kanbanFlagShort` / `kanbanFlagLong`, plus the callers in `cc.go`, `glm.go`, and `factory.go` (the tokens therefore move to the retired-entry file in M5a) | 1 |
| PV-78 | negative controls for the test helpers: on the M5b tree, delete `internal/hook/session_start_env_helper_test.go` and run `go vet ./internal/hook/`; on the M6b tree, delete `internal/kanban/test_helpers_test.go` and `test_helpers_windows_test.go` and run `go test -c -o /dev/null -gcflags=-e ./internal/kanban/` on the host and on `GOOS=windows` (the test compile lists every error; `go vet` stops at the first) | `clearKanbanEnv` is called by `session_start_additional_context_test.go`, `session_start_factory_provider_test.go`, `session_start_factory_test.go`, and more (first ten errors listed below); `runtimeIsWindows` by `f3_f4_probe_test.go`, `integration_lock_cross_test.go`, `slot_lease_cross_test.go`, `status_read_test.go`; `runGitAt` by `f3_f4_probe_test.go`; `deadPID` by `integration_lock_rotation_test.go`; windows only, `deadPIDWin` by `integration_lock_mutation_windows_test.go:52`; `readFileBytes` is called by no retained test once the board cases are gone, so it is not needed | 1 / 1 / 1 |
| PV-79 | the M7 stage of PV-74; the interim collision list from the replay of the same data before the corrections: `bootstrap.go:309:2: factorySocketDir redeclared` (`kanbanSocketDir` left alive by M6), `codex_launcher.go:769:2: codexFactoryRefusalDiag redeclared` (alias `codexKanbanRefusalDiag` left alive by M5a), file-name collision `kanban_dispatch_test.go` to `factory_dispatch_test.go` (already exists) | the M7 runner note in the stage table; after the corrections no collision remains and the stage is clean | 0 |
| PV-80 | negative control for `role.go`: on the M6b tree, delete the whole file `internal/kanban/role.go` and run `go build -gcflags=-e ./...` | 10 compile errors in the first failing package (listing below the table): `bootstrap.go` 4 uses of `RoleLeader`, `record.go:151` `RoleLeader` and `RoleLane`, `todo_owner_label.go` `IsLegacyLeaderSpelling`, `RoleLeader`, `legacyLeaderSpelling`, `RoleLane` — the four symbols that stay | 1 |
| PV-81 | the M8 stage of PV-74 | the M8 runner note in the stage table: package clauses renamed, qualified references renamed, shadowing local occurrences renamed; the single shadowing site is `internal/hook/stale_run_m1_test.go:118` (`factoryRun`, four occurrences); the three other design.md §4.7 sites are in files that do not import the package, so the Gap of design.md §4.7 is closed | 0 |
| PV-82 | the M9 stage of PV-74; interim prototype of the same edits on the M9 tree before the view-model edits entered the stage: delete `ChainRoles`, then delete `ChainVM`, `RoleVM`, `buildChain`, `chainRoleRecords`, `chainCardID`, and `go build -gcflags=-e ./internal/web/` | deleting `ChainRoles` alone: `viewmodel_ops.go:262:43`, `:263:23` (in `chainRoleRecords`) and `:330:23` (in `buildChain`) undefined; deleting the model: `viewmodel_ops.go:110:13` (`OverviewVM.Chain`), `:119:13` (the screen model's `Roles`), `:615:13` and `:677:11` (`buildChain` in the Overview and factory builders), `:615:45` and `:677:48` (`chainCardID`), `:628:78` (`buildAttention`'s parameter), `:676:18` (`chainRoleRecords`) — the Overview builder and `buildAttention` are the second reader; after the view-model edits the stage is clean and five web test files lose the cases that use the model | 1 |
| PV-83 | `go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/entry_probe.go -root .` (the committed program; one invocation from the repository root; it writes the committed probe test into a temporary directory outside the tree and runs `go test -overlay`) | the verbatim output is below the table (the same rows as RED-11 to RED-14); `git status --short` after the first run listed only the new untracked `probe/` directory, so the probe wrote nothing into the tree | 0 |
| PV-84 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` (scratch binary `moai-v8`, built by `go build -o ... ./cmd/moai` from this tree; its `version` prints `moai-adk v3.1.3` and no commit, so its judging build is identified by the tree it was built from, HEAD `5445e296caa48e6ab9821afe808eac2e7385e897`, per verification-claim-integrity §2.2); the plan-auditor traceability verb (`plan-auditor.md` Group 4, saved verbatim as `scratchpad/probe7/trace.sh`) on `spec.md` and `acceptance.md`; hand counts; the plan-auditor CN-4 ordering verb (Group 6, extracted verbatim from `plan-auditor.md` into a scratch script) on `plan.md` and `acceptance.md` | lint: `✓ No findings — all SPEC documents are valid`, exit 0; collection verb: `COLLECTED: 25 REQ definitions (acceptance input: read)`, no `UNCOVERED` and no `ORPHAN` line, exit 0; positive control on a coverage-scrubbed scratch copy (every `REQ-025` outside its definition line rewritten to `REQ-0XX` in `spec.md` and `acceptance.md`): `COLLECTED: 25 REQ definitions`, `ORPHAN: REQ-0`, `UNCOVERED: REQ-025`, so the silence on the real files is measured and not blind; hand counts: 25 criteria, 25 `Verifies` lines, 25 `Command` lines, 25 requirement lines (ceilings 25 and 25, both at the ceiling); `Verifies` lines naming each requirement: REQ-001:2 REQ-002:1 REQ-003:1 REQ-004:1 REQ-005:1 REQ-006:1 REQ-007:1 REQ-008:1 REQ-009:1 REQ-010:1 REQ-011:1 REQ-012:1 REQ-013:1 REQ-014:1 REQ-015:1 REQ-016:1 REQ-017:1 REQ-018:1 REQ-019:1 REQ-020:1 REQ-021:1 REQ-022:2 REQ-023:2 REQ-024:1 REQ-025:1 (every REQ-001..REQ-025 at least once; REQ-001, REQ-022, REQ-023 twice by design); CN-4 verb: `COLLECTED: 12 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7 M8 M9 M10 M11), 35 exit bindings, 46 ordering candidates` — no `CONFLICT`, no `GAP` (the verb reads M5a and M5b as the one label M5, so it collects 12 labels for the 13 milestones; its 46 candidates were read one by one and none orders an acceptance criterion across the plan order, PV-88) | 0 |
| PV-85 | the RED-now commands of every criterion this revision changed (AC-003, AC-004, AC-005, AC-012, AC-018, AC-020, AC-023), re-run on this tree: the launcher cells with the tree binary, the grep and `find` cells through a script that prints exit codes and line counts (the shell's `grep` is a function that supports `-P`; the two PCRE cells were re-run in that shell) | listing below the table: every cell reproduces its ledger value (RED-1, RED-2, RED-3, RED-5, RED-6, RED-10, RED-K2 exit 1 with the ledger stderr; RED-K3 15; RED-K4 27 (new pattern); the RED-K4 board-family `find` 18 and its positive control 4; RED-N1 189; RED-N2 22; RED-N3 179; RED-C1 exit 1 three lines; RED-C2 two lines; RED-C3 39965, 18909, 26030; RED-C4 three counts of 1; RED-D1 157; RED-D2 4; RED-D3 0 exit 1; RED-D5 46 with the new pattern and 46 with the old (unchanged union); RED-D8 16; RED-D7 19; RED-S1 16) | 1 (RED-1, RED-2, RED-3, RED-5, RED-6, RED-10, RED-K2) / 0 (the grep and `find` cells) / 1 (RED-C1, RED-D3) |
| PV-86 | baselines on the pre-change tree: `go test ./internal/kanban -run '^(TestBoardLockWaitBudgetDerivedFromNamedInputs\|TestBoardLockWaitBudgetCoversSerializedMutations\|TestBoardLockRetryWaitIsNotLockstep\|TestBacklogLockStuckHolderSurfacesBoundedNamedError)$' -v -count=1`; `go test ./internal/kanban -run 'Foreman' -v -count=1`; `go test ./internal/template -run '^TestBacklogJSONDisclosure_' -v -count=1` | the four wait-budget tests PASS (0.00 s, 0.00 s, 0.00 s, 3.32 s, `ok ... 4.318s`); the seven foreman tests PASS (`TestForemanQueueWatchResolvesCanonicalStateDir`, `TestForemanQueueWatch_FiresOnMutation`, `_ShippedJSONTargetIsSilent`, `_FiresWithStaleJSONPresent`, `_WatchTargetsAgree`, `_DBOnlyTargetMissesWALDeferral`, `_SeesWALDeferredCommit`; `ok ... 129.305s`); the two template tests PASS (`ok ... 0.989s`); the tests that read the foreman skill by path are `foreman_queue_watch_test.go:66-67`, `foreman_queue_statement_test.go:36`, and `backlog_json_disclosure_mirror_test.go:26`; the three catalog tests name the id in comments only | 0 / 0 / 0 |
| PV-87 | `grep -rlE 'EnvMoaiKanban(ID\|LeadAddr\|LeadName\|SettingsInjected\|Backend\|Card)\b\|MOAI_KANBAN_(ID\|LEAD_ADDR\|LEAD_NAME\|SETTINGS_INJECTED\|BACKEND\|CARD)\b' internal cmd --include='*.go'` (each count by a separate `wc -l`, tests by `grep -c '_test.go'`); the constants-only variant; the literals-only variant | union 92 files, 61 test, 31 production; constants only 85 files, 56 test, 29 production; literals only 27 files; so seven files (two production, five test) spell only the literal | 0 |
| PV-88 | the cross-artifact ordering check by hand (table below): every ordering obligation in the acceptance surface against the milestone order of plan.md §D | every obligation is on the correct side; the CN-4 verb prints no `CONFLICT` | 0 |
| PV-89 | stale-label sweep over the seven artifacts (commands and results below the table) | swept all seven artifacts and the probe directory for the old counts (24 requirements, 95/34 files), the old M5a/M5b/M6 deletion lists (ten board files, 13 tests, the four-constant budget), and the stale milestone flips (RED-N2 at M7, AC-018 flips at M7 and M8); results in the list below the table; the historical evidence rows PV-24, PV-44, PV-46b, PV-47, PV-51, PV-70 keep their original figures and carry a superseded note where a figure changed | 0 |
| PV-90 | green-path observations of the acceptance commands that name deletions and renames, run on the scratch tree the committed runner leaves after stage M11 (`replay-final/tree`; the paths read `internal/factory` there): AC-012's identifier grep, symbol grep, board file-name `find`, kept-file `find`, and substrate consumers; AC-018's file-name `find`, old-import grep, and package directory; AC-019's panel string and legacy-route file; AC-020's renamed rule paths, old rule path, catalog counts, archive-list count, and skill directories | listing below the table: the board file names are absent (0 lines), the ten kept files are present (10 lines), the substrate consumers carry the re-homed names (`backlog_store.go` 2, `integration_lock_mutation.go` 2, `slot_lease.go` 3), `role.go` keeps `RoleLeader`, no file name carries the word, the old import path and `internal/kanban` are gone, `Chain session board` is gone from `screens.templ` and `legacy_routes.go` exists, the three renamed rule files exist and the old one does not, the catalog has 0 old and 2 new entries, the archive list carries the old id once, and the skill directory is renamed; two greps are NOT empty on this tree because the probe rewords no comment: AC-012's identifier grep still lists 7 files and its symbol grep 2 files, every match a comment (26 comment lines in `envkeys.go`, `factory_launch_helpers.go`, `factory.go`, `factory_settings.go`, `session_start_factory.go`, `factory/role.go`, `factory/bootstrap.go`, plus `state_lock.go:22` and `integration_lock_mutation.go:24`), which the plan assigns to M5a, M5b, and M6 | 0 |

### PV-74 — per-stage result and verbatim runner output

Stage table (read from the runner output below by `build_progress.py`-style parsing: a cell says `clean` only for an empty block with exit 0; the windows vet column says `baseline failure only` where the block equals the three baseline lines of PV-73 and nothing else):

| Stage | host build | windows build | host vet | windows vet | runner note |
|---|---|---|---|---|---|
| M0 | clean | clean | clean | baseline failure only | no Go change (printed by the runner) |
| M1 | clean | clean | clean | baseline failure only | no Go change (printed by the runner) |
| M2 | clean | clean | clean | baseline failure only |  |
| M3 | clean | clean | clean | baseline failure only |  |
| M4 | clean | clean | clean | baseline failure only |  |
| M5a | clean | clean | clean | baseline failure only |  |
| M5b | clean | clean | clean | baseline failure only |  |
| M6a | clean | clean | clean | baseline failure only |  |
| M6b | clean | clean | clean | baseline failure only |  |
| M7 | clean | clean | clean | baseline failure only | identifiers renamed: 515 |
| M8 | clean | clean | clean | baseline failure only | package clauses renamed: 156, qualifiers renamed: 1830, shadowing locals renamed: 4 |
| M9 | clean | clean | clean | baseline failure only | web identifiers renamed: 91 (14 distinct) |
| M10 | clean | clean | clean | baseline failure only |  |
| M11 | clean | clean | clean | baseline failure only | no Go change (printed by the runner) |

The windows vet block of every stage is identical to the baseline of PV-73 (a failure in `internal/cli/worktree`, a package this SPEC does not touch); no block names a package this SPEC changes. Stages M0, M1, and M11 print no Go change by design. The runner's own output, verbatim:

```
##### stage M0
(no Go change: no code, baseline only)
=== [M0] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M0] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M0] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M0] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M1
(no Go change: adds test files only; the additions are not authored by this probe)
=== [M1] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M1] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M1] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M1] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M2
=== [M2] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M2] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M2] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M2] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M3
=== [M3] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M3] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M3] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M3] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M4
=== [M4] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M4] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M4] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M4] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M5a
=== [M5a] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M5a] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M5a] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M5a] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M5b
=== [M5b] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M5b] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M5b] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M5b] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M6a
=== [M6a] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M6a] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M6a] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M6a] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M6b
=== [M6b] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M6b] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M6b] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M6b] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M7
identifiers renamed: 515
=== [M7] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M7] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M7] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M7] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M8
package clauses renamed: 156, qualifiers renamed: 1830, shadowing locals renamed: 4
=== [M8] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M8] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M8] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M8] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M9
web identifiers renamed: 91 (14 distinct)
templ generate exit 0
(✓) Post-generation event received, processing... [ updates=0 needsRestart=true needsBrowserReload=true ]
(✓) Post-generation event received, processing... [ updates=1 needsRestart=false needsBrowserReload=false ]
(✓) Complete [ updates=1 duration=111.639958ms ]
=== [M9] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M9] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M9] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M9] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M10
=== [M10] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M10] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M10] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M10] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M11
(no Go change: documentation only)
=== [M11] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M11] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M11] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M11] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
runner-exit=0
```

### PV-75 — 25 compile errors when the ten files are deleted without the re-home

Host build:

```
# github.com/modu-ai/moai-adk/internal/kanban
internal/kanban/backlog_store.go:1254:40: undefined: BoardLock
internal/kanban/backlog_store.go:1259:29: undefined: boardLockWaitBudget
internal/kanban/backlog_store.go:1261:16: undefined: acquireBoardLockImpl
internal/kanban/backlog_store.go:1263:12: undefined: BoardLock
internal/kanban/backlog_store.go:1265:7: undefined: IsBoardLockHeld
internal/kanban/backlog_store.go:1272:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation.go:95:51: undefined: boardLockImpl
internal/kanban/integration_lock_mutation.go:97:29: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:99:16: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:103:7: undefined: IsBoardLockHeld
internal/kanban/integration_lock_mutation.go:119:26: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:126:73: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:128:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation_unix.go:18:53: undefined: ClearStaleReport
internal/kanban/integration_lock_mutation_unix.go:19:10: undefined: ClearStaleReport
internal/kanban/slot_lease.go:463:49: undefined: boardLockImpl
internal/kanban/slot_lease.go:465:29: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:467:16: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:471:7: undefined: IsBoardLockHeld
internal/kanban/slot_lease.go:477:26: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:482:67: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:484:14: undefined: boardLockRetryWait
internal/kanban/slot_lease_mutation_unix.go:9:51: undefined: ClearStaleReport
internal/kanban/slot_lease_mutation_unix.go:10:10: undefined: ClearStaleReport
internal/kanban/role.go:86:23: undefined: BoardDir
```

`GOOS=windows GOARCH=amd64` build:

```
# github.com/modu-ai/moai-adk/internal/kanban
internal/kanban/backlog_store.go:1254:40: undefined: BoardLock
internal/kanban/backlog_store.go:1259:29: undefined: boardLockWaitBudget
internal/kanban/backlog_store.go:1261:16: undefined: acquireBoardLockImpl
internal/kanban/backlog_store.go:1263:12: undefined: BoardLock
internal/kanban/backlog_store.go:1265:7: undefined: IsBoardLockHeld
internal/kanban/backlog_store.go:1272:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation.go:95:51: undefined: boardLockImpl
internal/kanban/integration_lock_mutation.go:97:29: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:99:16: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:103:7: undefined: IsBoardLockHeld
internal/kanban/integration_lock_mutation.go:119:26: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:126:73: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:128:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation_windows.go:24:56: undefined: ClearStaleReport
internal/kanban/integration_lock_mutation_windows.go:25:9: undefined: clearStaleLockAtPath
internal/kanban/slot_lease.go:463:49: undefined: boardLockImpl
internal/kanban/slot_lease.go:465:29: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:467:16: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:471:7: undefined: IsBoardLockHeld
internal/kanban/slot_lease.go:477:26: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:482:67: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:484:14: undefined: boardLockRetryWait
internal/kanban/slot_lease_mutation_windows.go:14:54: undefined: ClearStaleReport
internal/kanban/slot_lease_mutation_windows.go:15:9: undefined: clearStaleLockAtPath
internal/kanban/role.go:86:23: undefined: BoardDir
```

### PV-76 — 8 compile errors when `langEnglish` and `operatorLang` are not moved first

```
# github.com/modu-ai/moai-adk/internal/hook
internal/hook/session_stale_run.go:83:2: undefined: langEnglish
internal/hook/session_stale_run.go:136:25: undefined: langEnglish
internal/hook/session_start_factory_i18n.go:72:2: undefined: langEnglish
internal/hook/session_start_factory_i18n.go:258:24: undefined: langEnglish
internal/hook/factory_messages.go:69:78: undefined: langEnglish
internal/hook/session_start.go:511:91: undefined: langEnglish
internal/hook/session_start.go:523:74: undefined: operatorLang
internal/hook/session_start.go:542:52: undefined: operatorLang
```

### PV-78 — test-helper controls

`internal/hook` without `session_start_env_helper_test.go` (`go vet ./internal/hook/`, first errors):

```
# github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
internal/hook/session_start_additional_context_test.go:147:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_provider_test.go:24:5: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:32:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:46:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:89:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:103:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:120:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:157:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:200:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:244:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:244:2: too many errors
```

`internal/kanban` without `test_helpers_test.go` and `test_helpers_windows_test.go`, host test compile:

```
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/f3_f4_probe_test.go:18:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:22:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:23:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:24:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:29:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:30:21: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:63:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:67:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:68:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:69:2: undefined: runGitAt
internal/kanban/integration_lock_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_cross_test.go:311:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_rotation_test.go:113:10: undefined: deadPID
internal/kanban/slot_lease_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:109:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:139:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:162:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:196:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:224:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:288:5: undefined: runtimeIsWindows
```

the same on `GOOS=windows` (the extra error is `deadPIDWin`):

```
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/f3_f4_probe_test.go:18:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:22:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:23:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:24:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:29:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:30:21: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:63:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:67:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:68:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:69:2: undefined: runGitAt
internal/kanban/integration_lock_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_cross_test.go:311:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_mutation_windows_test.go:52:41: undefined: deadPIDWin
internal/kanban/integration_lock_rotation_test.go:113:10: undefined: deadPID
internal/kanban/slot_lease_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:109:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:139:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:162:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:196:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:224:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:288:5: undefined: runtimeIsWindows
```

### PV-80 — deleting the whole file `role.go` on the M6b tree

```
# github.com/modu-ai/moai-adk/internal/kanban
internal/kanban/bootstrap.go:89:9: undefined: RoleLeader
internal/kanban/bootstrap.go:97:9: undefined: RoleLeader
internal/kanban/bootstrap.go:119:14: undefined: RoleLeader
internal/kanban/bootstrap.go:123:23: undefined: RoleLeader
internal/kanban/record.go:151:13: undefined: RoleLeader
internal/kanban/record.go:151:35: undefined: RoleLane
internal/kanban/todo_owner_label.go:35:7: undefined: IsLegacyLeaderSpelling
internal/kanban/todo_owner_label.go:38:10: undefined: RoleLeader
internal/kanban/todo_owner_label.go:38:49: undefined: legacyLeaderSpelling
internal/kanban/todo_owner_label.go:47:10: undefined: RoleLane
```

### PV-83 — the entry-parse probe, verbatim

```
=== RUN   TestZZT1399EntryProbe
CCPARSE args=["-f" "--name" "lane-2"] factoryEnabled=true rest=["--name" "lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["-f" "-n" "lane-2"] factoryEnabled=true rest=["-n" "lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["--factory" "--name=lane-2"] factoryEnabled=true rest=["--name=lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["-f" "--name" "leader-r7"] factoryEnabled=true rest=["--name" "leader-r7"] laneLabel="" isLane=false branch=1
CCPARSE args=["--lane"] factoryEnabled=false rest=["--lane"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "lane-2"] factoryEnabled=false rest=["--lane" "lane-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane=lane-2"] factoryEnabled=false rest=["--lane=lane-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "3"] factoryEnabled=false rest=["--lane" "3"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "leader-2"] factoryEnabled=false rest=["--lane" "leader-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "-f"] factoryEnabled=true rest=["--lane"] laneLabel="" isLane=false branch=1
CCPARSE args=["-l"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "lane-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "3"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l=lane-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "-f"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "-k"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "--name" "lane-2"] parseErr="--leader requires a leader label"
CCPARSE args=["-f" "-l"] parseErr="--leader requires a leader label"
CCPARSE args=["--lane" "-k"] factoryEnabled=false rest=["--lane"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "--name" "lane-2"] factoryEnabled=false rest=["--lane" "--name" "lane-2"] laneLabel="lane-2" isLane=true branch=0
CCPARSE args=["--leader" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["--leader=leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-f" "--leader" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-f" "--leader=leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CODEX args=["--lane"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane=lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "3"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "leader-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "-f"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--lane" "-k"] err= exitCode=1 stderr="KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead\n" stdoutLen=0
CODEX args=["--lane" "--name" "lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["-f" "-l"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--leader" "leader-2"] err=--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target exitCode=-1 stderr="" stdoutLen=0
CODEX args=["-f" "--leader" "leader-2"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
--- PASS: TestZZT1399EntryProbe (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.196s
exit 0
```

### PV-85 — RED-now re-runs of the changed criteria, verbatim

Launcher cells (tree binary):

```
RED-1 moai cc -l: exit=1 stdout-bytes=0 stderr=ERROR --Leader requires a leader label.
RED-2 moai codex -l: exit=1 stdout-bytes=0 stderr=ERROR --Leader requires a leader label.
RED-3 moai cc -l lane-2: exit=1 stdout-bytes=0 stderr=ERROR --Leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target.
RED-5 moai codex -f 3: exit=1 stdout-bytes=0 stderr=FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader
RED-6 moai cc -f 3: exit=1 stdout-bytes=0 stderr=ERROR -F/--Factory takes no argument (the factory leader), the role token -f lane, which joins this session to a running factory as the next free lane, or a lane label (e.g. -f lane-2) that launches exactly that one lane, got "3".
RED-10 moai cc -f -l: exit=1 stdout-bytes=0 stderr=ERROR --Leader requires a leader label.
RED-K2 moai codex -k: exit=1 stdout-bytes=0 stderr=KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead
```

Grep and `find` cells (script output: exit code and stdout line count; the two PCRE cells RED-N1 and RED-D1 printed `exit=0` with 189 and 157 lines in the interactive shell):

```
RED-K3: exit=0 stdout-lines=15 
RED-K4: exit=0 stdout-lines=27 
RED-K4-find: exit=0 stdout-lines=18 
AC-012-positive-control-find(today): exit=0 stdout-lines=4 
RED-N1: exit=2 stdout-lines=0 
RED-N2: exit=0 stdout-lines=22 
RED-N3: exit=0 stdout-lines=179 
RED-C1: exit=1 stdout-lines=3 
RED-C2: exit=0 stdout-lines=2 internal/template/catalog.yaml:41:            - name: moai-kanban-foreman | internal/template/catalog.yaml:43:              path: templates/.claude/skills/moai-kanban-foreman/
RED-C3: exit=0 stdout-lines=3 39965 | 18909 | 26030
RED-C4: exit=0 stdout-lines=3 .claude/rules/moai/workflow/kanban-dispatch-detail.md:1 | .claude/rules/moai/workflow/kanban-dispatch-mechanics.md:1 | .claude/rules/moai/workflow/cross-session-messaging-detail.md:1
RED-D1: exit=2 stdout-lines=0 
RED-D2: exit=0 stdout-lines=1 4
RED-D3: exit=1 stdout-lines=1 0
RED-D5: exit=0 stdout-lines=46 
RED-D5-old-pattern: exit=0 stdout-lines=46 
RED-D8: exit=0 stdout-lines=16 
RED-D7: exit=0 stdout-lines=19 
RED-S1: exit=0 stdout-lines=16 
docs-line-cites: exit=0 stdout-lines=2 38:| `-l, --lead <name>` | With `-f lane` / `-f lane-<n>`: which leader session the record-absence verification targets (default `leader`; the former spelling `lead` is refused). When the run's record is missing or retired while a live leader exists, the join verifies that leader (pid + process-start) and restores its run | | 80:Grow a run one lane at a time with `moai cc -f lane` (auto-join the next free number) or `moai cc -f lane-<n>` (that number exactly). Both forms already name the lane, so passing `--name`/`-n` alongside them is an error. An explicitly-picked number that collides with a live lane bumps to the next free number. A number is otherwise skipped only while a live session holds it — a dead lane's claim no longer blocks its number (an explicit pick reuses it right away), but `-f lane` auto-assignment always takes one past the highest live number and never backfills a gap. Lane ownership is recorded in `~/.moai/db/<project-key>/factory/factory.db` — or, when the launch directory is a temporary one (no absolute `MOAI_HOME` override), project-local under `<base>/.moai/db/<project-key>/factory/`, the same exception the backlog queue follows; a legacy `.moai/state/factory/workers.json` is imported once and retained only as rollback evidence. A lane runs up to 10 concurrent `Agent()` subagents, and write-capable spawns are isolated in their own worktree. Never bring every lane up at once — start the first, confirm it is actually producing output, then activate the rest. Cards are never split across lanes. `-k` still drives the three-role kanban chain; one launch takes one entry token, so `-k` with `-f` is an error. CG is retired; use `moai migrate cg` to preview explicit migration choices. A factory run now records the process identity of the session that owns it, so a run whose leader has died is retired automatically the next time a lane joins instead of leaving that join stuck on `AMBIGUOUS_FACTORY`. The inverse is covered too: when a lane joins and the run's record is missing or retired while a live leader session exists, the join verifies that leader (pid plus process-start fingerprint, targeted with `-l/--lead`, default `leader`), restores its run record, and joins anyway — two or more verified leaders fail closed naming each candidate. `moai factory runs` lists every run with its owner's liveness, and `moai factory runs --retire <run-id>` retires one by hand — refused unless that run's owner is actually dead.
```

### PV-90 — green-path observations on the modeled final tree, verbatim

```
AC-012 chain/companion identifiers (non-test): exit=0 stdout-lines=7
AC-012 board API and carrier symbols: exit=0 stdout-lines=2
AC-012 board file names: exit=0 stdout-lines=0
AC-012 kept files: exit=0 stdout-lines=10
AC-012 substrate consumers: exit=0 stdout-lines=3 | internal/factory/backlog_store.go:2 | internal/factory/integration_lock_mutation.go:2 | internal/factory/slot_lease.go:3
AC-012 role.go keeps RoleLeader: exit=0 stdout-lines=1 | 2
AC-018 file names: exit=0 stdout-lines=0
AC-018 old import path: exit=1 stdout-lines=0
AC-018 package directory: exit=1 stdout-lines=2 | ls: internal/kanban: No such file or directory | internal/factory
AC-019 panel string: exit=1 stdout-lines=1 | 0
AC-019 legacy route file: exit=0 stdout-lines=1 | internal/web/legacy_routes.go
AC-020 renamed rules (template): exit=0 stdout-lines=3
AC-020 old rule path: exit=1 stdout-lines=1 | ls: internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md: No such file or directory
AC-020 catalog old/new: exit=0 stdout-lines=2 | 0 | 2
AC-020 archive list: exit=0 stdout-lines=1 | 1
AC-020 skill dir: exit=1 stdout-lines=2 | ls: internal/template/templates/.claude/skills/moai-kanban-foreman: No such file or directory | internal/template/templates/.claude/skills/moai-factory-foreman
```

### PV-88 — cross-artifact ordering check, by hand

Every clause that orders work in the acceptance surface (the Definition of Done, and each criterion's Given, When, and Then), read against the order of plan.md §D and, where the clause concerns a deletion or a rename, against the compile proof of PV-74. The CN-4 verb (PV-84) flags candidate records mechanically; this table is the by-hand reading of each obligation, including the ones the verb cannot see (a clause spread across two sentences, or a milestone bound by prose).

| # | Obligation (where it is stated) | Plan binding | Satisfied |
|---|---------------------------------|--------------|-----------|
| 1 | The AC-004 golden is committed alone BEFORE any change (AC-004 Given; Definition of Done 6) | M1 commits the golden alone; M2 is the first milestone that changes a parse | yes |
| 2 | AC-015's eight mutant reds (a)-(h) and AC-017's mutant red are recorded BEFORE M5a starts (AC-015, AC-017, Definition of Done 2) | M1 authors the net and the tolerance test and records the reds; M5a is the first removal | yes |
| 3 | Mutants (g) and (h) are re-observed red AFTER the M2 re-pin of the lane tests (AC-015, Definition of Done 2) | M2 states the re-observation as its closing step | yes |
| 4 | RED-11, RED-12, RED-14 are re-observed on the milestone tree BEFORE the cc/glm parse changes; RED-13 and the codex rows of RED-14 BEFORE the codex parse changes (AC-006, AC-003, AC-005, Definition of Done 1) | M2 and M3 each open with that observation as their first act (the M3 first act was added in v0.8.0) | yes |
| 5 | The update fixture of AC-020 is observed RED on the pre-rename template BEFORE the rename commit (AC-020, Definition of Done 3) | M10 commit 1 is the fixture test alone, commit 2 the rename; the commit graph is the witness | yes |
| 6 | `moai constitution validate` runs BEFORE the first constitution edit and AFTER each of the four (AC-022) | M10 commit 2 | yes |
| 7 | The three constants `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL` are deleted only AFTER their last readers (AC-012, AC-013, Definition of Done 7) | M5a removes the launcher readers, M5b removes the hook readers and then the constants; compile proof PV-74 stages M5a and M5b | yes |
| 8 | `leadFlagShort` is deleted only AFTER the Codex parse stops reading it (Definition of Done 7) | M3 deletes it with its last reader; PV-77 is the negative control for M2 | yes |
| 9 | The shared lock substrate is re-homed BEFORE the board is deleted (AC-012, Definition of Done 7) | M6 step 1 (stage M6a), then step 2 (stage M6b); PV-75 is the negative control | yes |
| 10 | `langEnglish` and `operatorLang` move BEFORE the kanban notice file is deleted; the test helpers move BEFORE the test files that define them (AC-012, AC-014) | M5b step 1; M5b and M6 helper moves; PV-76 and PV-78 are the negative controls | yes |
| 11 | Renames come AFTER removals so nothing is renamed and then deleted (design.md §1) | M7 to M10 follow M6; `internal/web` is renamed at M9, after the Go-wide M7 and M8, and its view model is deleted in the same stage; no symbol renamed at M7 is deleted later (the chain view model is not renamed at M7) | yes |
| 12 | The package path is `internal/kanban` through M7 and `internal/factory` from M8 on (AC-012, AC-016, AC-020) | M8 renames the package; M9 imports it | yes |
| 13 | The web sources carry the word until M9 lands, the template mirror paths until M10 lands, so AC-018's full command closes at M10 (AC-018) | M7, M8, M9, M10 each clear their part; AC-018 is bound to M10 | yes |
| 14 | AC-013's grep half is read AFTER the hook readers are gone (AC-013) | M5a test half, M5b grep half | yes |
| 15 | The foreman path-pinned tests are re-pointed in the SAME commit as the skill rename (AC-020) | M10 commit 2 | yes |
| 16 | The net runs at every milestone from M2 through M9 and the windows build at every Go milestone (AC-015, AC-018) | the verification gate of plan.md §D | yes |
| 17 | The four locales and each page land in ONE commit; AC-023 and AC-025 close at sync-audit (AC-023, AC-025) | M11 | yes |
| 18 | M2, M3, and M4 merge as one integration unit; M6 and M10 hold two commits in the stated order (spec.md §E, plan.md §D) | plan.md §D integration units | yes |

Sweep results:

- old counts and labels: `grep -rnE '24 requirements\|24 of 25\|REQ-001\.\.REQ-024\|M11 scope\|95 files\|34 production\|M7 flips RED-N2\|flips at M7, M8\|28 references' *.md \| grep -v '^progress.md:\(66\\|94\\|90\\|89\\|125\)' \| cut -c1-80` -> progress.md:119:| PV-70 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict / spec.md:31:- 2026-10-02 (v0.7.0): Plan-audit iteration 1 (FAIL, 0.72; findings D
- old deletion lists: `grep -rnE '10 board files and 13 tests\|13 board tests\|four-constant budget\|board family \(10 files\)' *.md \| cut -c1-80` -> no output (exit 0)
- requirement count lines: `grep -n 'this SPEC carries' acceptance.md spec.md \| cut -c1-150` -> acceptance.md:3:All criteria are mechanically verifiable. Tier L ceiling: 25 requirements and 25 criteria; this SPEC carries 25 requirements and 25 cr / spec.md:156:Full Given-When-Then scenarios, the RED-now evidence ledger, and the Definition of Done live in `acceptance.md`. Tier L ceiling: 25 requir

Reading of the sweep: the two hits of the first command are the v0.7.0 history line of `spec.md` (line 31) and the historical row PV-70, which record the v0.7.0 counts and stay as history; the second command is empty (the old M5a/M5b/M6 deletion lists are gone); the third shows both ceiling statements at 25 requirements and 25 criteria.

Re-run of the RED-now cells of the criteria this revision changed, on HEAD `b9242da00ce489c4f26efb5f6d447ccafef08c54` (this run): RED-1, RED-3, RED-10 (`cc -l`, `cc -l lane-2`, `cc -f -l`: exit 1, the same stderr lines as the ledger), RED-2, RED-5 (`codex -l` exit 1 `--Leader requires a leader label.`; `codex -f 3` exit 1 with the `FACTORY_MODE_UNSUPPORTED_BACKEND` line), RED-6, RED-4 (`--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (2.92s)`), RED-K3 (the same 15 files), RED-K4 (24), RED-8 (4), RED-8b (`envkeys.go`, `codex_launcher.go`), RED-N1 (189), RED-N2 (22), RED-N3 (179), RED-C1 (three `No such file`, exit 1), RED-C2 (`catalog.yaml:41,43`), RED-C3 (39965, 18909, 26030 bytes), RED-D1 (157), RED-D2 (4), RED-D3 (0, exit 1), RED-D4 (four zeros, exit 1), RED-D5 (46), RED-D6 (four `No such file`, exit 1), RED-D7 (19), RED-C4 (three counts of 1, exit 0), RED-11 to RED-13 (PV-60), RED-S1 (the same 16 files), BASE-4 (windows build exit 0), BASE-5 (2). The AC-015 cli selector with both settings-test names swept 4 tests today (`TestCCFactoryEntryRecordsFailOpenRunMetadata`, `TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestPrepareKanbanSettingsWritesTransientFile`) and sweeps 8 once the four net tests of M1 exist. The cells of AC-009, AC-010, AC-011, AC-014, AC-019, AC-021 were not changed in this revision and were not re-run.

Reading of PV-46/PV-47 (and PV-70, which supersedes them for v0.7.0): REQ-to-AC coverage is established by PV-47, not by the lint. Counts at authoring: 24 requirements (Tier L ceiling 25) and 25 criteria (Tier L ceiling 25 — at the ceiling; no further criterion without merging or splitting the SPEC).

Reading of PV-84 (supersedes PV-70 for v0.8.0): REQ-to-AC coverage is again established by the hand count, not by the lint; the lint reports `No findings` and the collection verb reports 25 definitions with no `UNCOVERED` line, and the verb's `mapped` is generous (it counts any line that names both an AC and the REQ), so the per-requirement `Verifies` counts of PV-84 are the evidence. Counts at authoring: 25 requirements (Tier L ceiling 25, at the ceiling after the REQ-019 split) and 25 criteria (ceiling 25, at the ceiling; the new `-l` and `--leader` shapes were folded into AC-003 and AC-005).

Gaps: see `plan.md` §G (sixteen plan-level items) and `research.md` §R14 (twenty-one items) — none is asserted as fact in `spec.md`. Gaps recorded by the v0.8.0 revision specifically: (1) the compile proof models each milestone's DELETIONS and the edits that keep the build, not the added code, and prints M0, M1, and M11 as stages with no Go change; (2) the semantic re-pin of every test the probe retargets (the removed constants are mapped to surviving factory markers as a compile model) and the per-function classification of the cases the probe deleted are the run phase's; (3) the five lock-coupled tests of `internal/kanban` are modeled by deletion although two of them (the errno classification and the cross-process exclusion) are to be re-pointed, and whether retained tests already cover the substrate was not measured; (4) the runtime effects of the M10 renames (embedded template paths, catalog hash, foreman path pins) are not compile effects and are gated by AC-018 to AC-022, not by the probe; (5) PV-77 and PV-79 quote two redeclaration errors and one file-name collision from an interim replay of the stage data before the corrections entered it, and PV-82 quotes the view-model reader errors from the M9 prototype tree, so those rows are attributable to the same edit set but not to the committed stage data end to end; (6) the committed runner's own run (PV-74) began before the header comment of `probe.go` was reworded; no code line changed, and the file was deliberately left as run (it is not gofmt-aligned in its `var` block) so that the recorded run stays attributable to it; (7) the plan-audit reports `.moai/reports/t1399/plan-audit-iter1.md` and `plan-audit-iter2.md` are local gitignored files; both were read and neither was edited; (8) the worktree guard refused several compound shell commands and the runs were re-issued as plain commands or scripts written with the Write tool, with no measurement replaced by source reading; (9) the Overview attention row (Q24) and the file `role.go` (verdict 20) are findings of the compile proof that the verdicts did not state — they are written to the smallest-footprint reading and returned for acknowledgement.

## §J Plan-audit iteration 3 outcome and carried debt

Appended after plan-audit iteration 3 (the last audit allowed). This section is append-only: no earlier line of this file was edited, and the five plan-artifact files (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`) were not touched, because the audit verdict is bound to their hash. Where this section contradicts an earlier line (the §E.1 prose says iteration 3 "has not run"; Gap (6) of §G says `probe.go` was left unformatted), this section is the later record and the earlier line is left as written history.

### J.1 Audit iterations

| Iteration | Verdict | Score | Audited at | Report (local, gitignored) |
|-----------|---------|-------|------------|----------------------------|
| 1 | FAIL | 0.72 | `b9242da00ce489c4f26efb5f6d447ccafef08c54` | `.moai/reports/t1399/plan-audit-iter1.md` |
| 2 | FAIL | 0.79 | `5445e296caa48e6ab9821afe808eac2e7385e897` | `.moai/reports/t1399/plan-audit-iter2.md` |
| 3 | PASS-WITH-DEBT | 0.88 (Tier L threshold 0.85) | `7e1ae0d808b180f266c71761290ab6ed96ba335f` | `.moai/reports/t1399/plan-audit-iter3.md` |

Iteration 3 reports MUST-FIX 0, SHOULD-FIX 8, ADVISORY 6 (14 findings, all Class optional). The audit-ready signal of §E.1 (`plan_status: audit-ready`, `plan_complete_at: 2026-10-02`) was not edited: it carries no `audited_sha` field, its `audited_head` field names the iteration 2 commit, and editing an existing line would break the append-only constraint of this section. The iteration 3 audited commit is recorded here instead.

### J.2 What the auditor re-ran, as the report states it

Source: `.moai/reports/t1399/plan-audit-iter3.md` § 4.2 and § 4.3 (read in this run; the report is the auditor's measurement, not mine).

- The committed compile-proof runner was re-run over all 14 stages (M0..M11 with M6a and M6b), four checks per stage. Every stage printed empty build and vet blocks with exit 0 on the host and for windows, except the windows vet block, which printed exactly the pre-existing three-line baseline (`internal/cli/worktree/sweep_test.go:1687`, `undefined: parseLsofCWDs`). The output equals the recorded PV-74 block except the `templ` generator's own timing lines at M9.
- The auditor made the proof go red on purpose: one line (`internal/kanban/state_lock_unix.go`) appended to a scratch copy of `M6b.rm`, so a re-homed lock file is deleted with no replacement. The host build failed with five `undefined: acquireStateLockImpl` errors; the windows build stayed clean (the windows file survived).
- A pristine-tree overlay control with the ten v0.7.0 "board family" files mapped to empty printed 25 `undefined:` lines, equal to PV-75.

### J.3 Carried debt

Owners are as the report assigns them. Where the report names no owner the row says `unassigned`; this section does not invent one. None of the plan-text fixes below can be applied now, because the five hash-subject files are frozen; every such fix is carried as run-phase or Kickoff work.

| ID | Class | One-line description | Owner (as the report gives it) | The report's minimal fix |
|----|-------|----------------------|--------------------------------|--------------------------|
| D-A1 | SHOULD-FIX (major) | `probe/probe.go` was not gofmt-clean (blank line missing before the M9 and M10 banners), and the CI format gate runs over tracked `.go` files | the orchestrator, before the card branch is merged or pushed | `gofmt -w probe/probe.go`, note it in progress.md Gap (6), re-run `gofmt -l` and the probe's M0 stage — EXECUTED in this run except the M0 re-run, see J.4 |
| D-A2 | SHOULD-FIX (major) | AC-018's exact-four grep still returns 123 files / 502 lines on the modeled final tree (22 files cite `SPEC-KANBAN-*`); the comment and citation sweep is unscheduled | M7 (comments), M10 (final sweep) | add a comment-and-citation sweep step to M7 with the measured count and a stated rewrite rule for completed-SPEC citations, or narrow REQ-017 and AC-018 to non-comment tokens |
| D-A3 | SHOULD-FIX (major) | run-time package-path strings in tests are ungated: `TestBacklogJSONLiteralStaysSeamScoped` fails on the modeled tree and `migrate_home_state_test.go:68,992,1059` carry `"./internal/kanban"` and are named in no artifact; Gap 13's claim is false | M8 | add the three `migrate_home_state_test.go` sites to design §4.7, add scoped lease-guarded test selectors to AC-018's command, correct Gap 13 |
| D-A4 | SHOULD-FIX (minor) | `-l -k` and `--lane -k`: REQ-002 and REQ-010 prescribe different one-line content with no stated precedence; AC-003's two rows go red at M5a unless re-pinned | M5a | state that `-k` is refused first and the line names both facts; add AC-003 to M5a's re-pin list |
| D-A5 | SHOULD-FIX (minor) | AC-003 covers `--name lane-2` but not `-n lane-2` or `--name=lane-2` beside `-l` or `--lane` (a shallow-mutant gap) | M2 | add the `-n` and `--name=` shapes and their `--lane` twins to AC-003's shape list and subtest count |
| D-A6 | SHOULD-FIX (minor) | AC-012 floors only the four wait-budget tests; the errno classification and cross-process exclusion tests could be deleted and AC-012 would still pass | M6 | add a scoped, lease-guarded selector over the retained consumer cross-process tests to AC-012's command, with a swept count |
| D-A7 | SHOULD-FIX (minor) | AC-017's M1-authored fixture may name `MOAI_KANBAN` and `MOAI_KANBAN_LABEL` as constants that M5b deletes; Gap 12 says the design reads none | M1 | state that M1-authored tests name those markers only as string literals; say so in Gap 12 |
| D-A8 | SHOULD-FIX (minor) | `graph-freshness.yml` runs `moai graph check` on every push to `develop`; the tracked codemaps cite `internal/kanban` 41 times, so the check is expected (inferred, not measured) to go red after the rename | M11 / sync | schedule `/moai codemaps` plus the stamp in the same batch as M8, or record that a red graph-freshness check is accepted until the debt card lands |
| D-A9 | ADVISORY | stale labels: `design.md:1` says v0.7.0; `design.md:96` and `research.md:146` print a closed gap as open; `design.md:47` and `research.md` §R4 say "two `-k` constants" where `plan.md:168` lists four; `decision-index.md:25` Q18 row maps to REQ-019 which REQ-025 now carries | unassigned (the report gives none) | the report names no fix beyond the label corrections it lists; not applicable now (frozen files, and `decision-index.md` is not among the files this task may touch) |
| D-A10 | ADVISORY | REQ-025 is printed between REQ-019 and REQ-020 (`spec.md:147`) | unassigned (the report gives none) | move it to the end of §B.4 or renumber |
| D-A11 | ADVISORY | AC-023's added alternative misses a page that writes the retired short as separate code spans ("`-l`, `--lead <name>`") | unassigned (the report gives none) | the report states the gap and gives no fix |
| D-A12 | ADVISORY | the proof and AC-018 cover darwin and windows only; linux-only files (`internal/discovery/leader_readers_linux.go`, `internal/session/proc_info_linux.go`) were measured clean only at the M11 tree | unassigned (the report's owner label is absent; its fix names M7) | add `GOOS=linux GOARCH=amd64 go build ./...` to the per-milestone gate at M7 (the marker-constant rename) |
| D-A13 | ADVISORY | Q20-Q22 and Q24 stay open and non-gating; REQ-018 states the Q24 reading normatively, so an answer of B at Kickoff changes REQ-018 and M9 | Kickoff | record the closures at Kickoff or sync |
| D-A14 | ADVISORY | REQ-019 still bundles several concerns (regression rows D12 and D45, partially resolved) | unassigned (the report gives none) | the report states the bundle and gives no fix |

Row count: 14 (D-A1..D-A8 SHOULD-FIX, D-A9..D-A14 ADVISORY). Owner column: nine rows carry an owner the report names (D-A1, D-A2, D-A3, D-A4, D-A5, D-A6, D-A7, D-A8, D-A13); five carry `unassigned` (D-A9, D-A10, D-A11, D-A12, D-A14), of which D-A12's fix names M7.

### J.4 Fix applied in this run

D-A1 only: `gofmt -w` on `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/probe.go`, run in this worktree at HEAD `7e1ae0d808b180f266c71761290ab6ed96ba335f` (this run, this tree). Observed:

- before: `gofmt -l .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe` printed `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/probe.go`;
- after: the same command printed nothing, exit 0;
- `go vet` over `probe/probe.go` and over `probe/entry_probe.go` (each a `//go:build ignore` file, vetted by path): exit 0 for both;
- the diff of the probe directory is exactly two inserted blank lines, one before the `// ---- M9: web console` banner and one before the `// ---- M10: rules, skills, catalog` banner; no code line changed. `entry_probe.go` was not listed by `gofmt -l` and was not changed; `entry_probe_test.go.txt` is not a Go file and was not touched.

This supersedes the "left as run" statement of Gap (6) in §G for the formatting only: the file's formatting changed by those two blank lines, its code did not, and the recorded PV-74 run was taken before them. The `var` block alignment that Gap (6) mentions needed no change (the formatter produced no diff there).

### J.5 Run-phase obligations (not requirement changes)

The following bind the run phase without changing a requirement or a criterion:

1. **AC-018 sweep sizing (D-A2).** At M1, size the exact-four-files grep sweep behind the audit's measured 123 files / 502 lines (41 files with a non-comment match; 22 files carrying `SPEC-KANBAN-*` citations) and record the measured edit list the plan already promises; the gofmt/format gate for tracked `.go` files (`make fmt-check`, which runs `gofmt -l` over `git ls-files '*.go'`) stays green in every milestone.
2. **Path-string tests (D-A3).** The run-time path-string tests the report names are added to the milestone that renames their path strings (M8): `TestBacklogJSONLiteralStaysSeamScoped` in `internal/factory`; the three `internal/cli` tests the auditor saw go red once the non-test path strings were renamed (`TestHomeStateVerifiedLiveGateCannotBypassOrReplay`, `TestHomeStateVerdictEvidenceValidator`, `TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases`), caused by `internal/cli/migrate_home_state_test.go:68,992,1059`; the fourth failure the auditor saw (`TestHomeStateValidationCommandWrappersAndHelperFailures`, `head="" err=exit status 128`) failed identically before the rename on a scratch tree without a repository, so it is an environment artifact of that scratch tree and not a SPEC obligation.

### J.6 Gaps of this section

- The M0 sanity re-run of the probe that D-A1's fix suggests was not executed in this run (the instruction scoped this run to formatting plus this section); the `go vet` exit 0 over both probe files is the check that was observed. The auditor's own full 14-stage re-run (J.2) was taken against the pre-format file and the diff is two blank lines, so the program is unchanged, but that is a reading of the diff, not a re-measurement.
- The report's counts (123 files, 502 lines, 41 codemaps citations, the four red tests) are the auditor's measurements; none was re-measured in this run.
- The plan-text fixes D-A2..D-A7, D-A9..D-A12 and D-A14 were not applied, by instruction (hash-bound files); `decision-index.md` (D-A9's Q18 row) was outside the two files this run was permitted to touch.
