# progress.md — SPEC-CC2219-UPSTREAM-ALIGN-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-07-25 (spec.md v0.1.0 draft, plan.md, acceptance.md, this skeleton). Evidence SSOT: `.moai/research/cc-update-2.1.207-to-2.1.219.md`. GD-3 excluded (PR #1146). M1 carries a pre-execution decision gate (REQ-GD1-005 sync-auditor pilot option a/b).

## §E.2 Run-phase Evidence

Run-phase executed 2026-07-25 in worktree `.moai/worktrees/cc2219` (branch `feat/SPEC-CC2219-UPSTREAM-ALIGN-001`, base origin/main 714270085). REQ-GD1-005 decision: **option (a)** — pilot retired, `Agent` removed from sync-auditor `tools` (user-resolved gate, injected by orchestrator; not re-asked).

Milestone commits: M1 `1fe7aaad9` (GD-1/GD-2 doctrine + pilot retire + spec.md in-progress/0.1.1), M2 `478c629ee` (GD-4 Go), M3 `8ce874111` (GD-4 docs sweep), M4 `f5481e212` (GD-5/6/7/8/9), M5 `30180bd10` (docs-site 4-locale).

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-GD1-001 | PASS | `grep -rn 'default changed to **off**\|defaults to \`0\` = nesting disabled\|runtime default-off' CLAUDE.md .claude/rules/ .claude/agents/ internal/template/templates/...` | 0 matches (`rc=1`) |
| AC-GD1-002 | PASS | `grep -rn 'double guarantee' CLAUDE.md .claude/rules/ .claude/agents/ internal/template/templates/` | 0 matches (`rc=1`) |
| AC-GD1-003 | PASS | `grep -rln 'MAX_SUBAGENT_SPAWN_DEPTH=1' <4 files>` | all 4 files returned (CLAUDE.md, agent-authoring.md, both mirrors) |
| AC-GD1-004 | PASS | `grep -c 'single depth-1' CLAUDE.md internal/template/templates/CLAUDE.md` | `1` / `1` |
| AC-GD1-005 | PASS (option a) | `grep -n '^tools:' sync-auditor.md` (live+T) | `tools: Read, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill` — no `Agent`; pilot section rewritten RETIRED |
| AC-GD1-006 | PASS | `grep -c 'SPEC-SUBAGENT-NESTING-DOCTRINE-001' CLAUDE.md` → 1; `grep -rn 'SPEC-SUBAGENT-NESTING' internal/template/templates/` | live 1 (dated supersession note); template 0 matches (`rc=1`) |
| AC-GD2-001 | PASS | `grep -rn 'mode: "plan"' <doctrine surfaces live+T>` | 0 matches (`rc=1`) |
| AC-GD2-002 | PASS | `grep -c 'tool restriction' sync-auditor.md worktree-integration.md` | 1 / 3 (live; mirrors carry same edits) |
| AC-GD2-003 | PASS | `grep -c 'changelog/doc-sourced' CLAUDE.md` | 1 (v2.1.213 provenance caveat in §4 Watch note) |
| AC-GD4-001 | PASS | `grep -n '"opus"' internal/template/model_policy.go` | `"opus": ModelIDOpus5` (= `claude-opus-5`); `ModelIDOpus48: "opus"` in ModelDeprecatedCanonicalIDs; `go test ./internal/template/` → `ok ... 1.483s` |
| AC-GD4-002 | PASS | `go build ./... && go test ./...` (also targeted `./internal/cli/ ./internal/settings/ ./internal/web/`) | exit 0; full suite 107 packages `ok`, 0 FAIL |
| AC-GD4-003a | PASS | `grep -rn 'claude-opus-4-8' internal/web/ --include='*.go' --include='*.templ' \| grep -v _test.go` | 0 matches (`rc=1`) |
| AC-GD4-003b | PASS | `go test ./internal/web/` (appbar_context_test asserts `claude-opus-5`) | `ok github.com/modu-ai/moai-adk/internal/web` |
| AC-GD4-004 | PASS | `grep -rn 'opus = Opus 4.8' .claude/rules/` → 0; `grep -c 'Opus 5 (1M)' context-window-management.md` → 1; `grep -c '\| Opus/Fable (256K)'` → 0 | table gained Opus 5 (1M) row; 256K row disambiguated to `Fable (256K)` |
| AC-GD4-005 | PASS | `grep -n xhigh quality.yaml.tmpl` | `"xhigh" and "max" are available on Opus 5, Sonnet 5, Opus 4.8, and Opus 4.7.` |
| AC-D-001 | PASS | `grep -n 'HUMAN-ONLY (manual-invocation-only' native-invocation-model.md` | L39 `/code-review` (v2.1.215), L42 `/deep-research` (v2.1.218); Axis A L67/L87 annotated (Live+T) |
| AC-D-002 | PASS | `grep -c 'manual invocation only\|invoked manually' CLAUDE.md dynamic-workflows.md` | 1 / 1 (Live+T) |
| AC-D-003 | PASS | `grep -c 'unrestricted\|workflowSizeGuideline' dynamic-workflows.md settings-management.md` | ≥1 each; explicit `medium` default (<15 agents) stated; settings key table row added (Live+T) |
| AC-D-004 | PASS | `grep -c 'DirectoryAdded' hooks-system.md` (live+T) | 3 / 3; entry notes official doc lags + no MoAI handler wired |
| AC-E-001 | PASS | `grep -c '/subtask' agent-authoring.md` (live+T) | 1 / 1; `/fork` = background session copy; 2.1.212/2.1.213 upstream inconsistency recorded |
| AC-E-002 | PASS | `grep -c 'background: false' skill-authoring.md claude-code-skills-official.md` | 1 / 1 (Live+T) |
| AC-F-001 | PASS | family greps over `docs-site/content/{en,ko,ja,zh}` + README.md + README.ko.md | F1 nesting: drift found in `claude-code/agentic/sub-agents.md` → 4-locale parity edit landed (M5 `30180bd10`). F2 `mode: "plan"`: **0 matches**. F3 opus-alias-default-4.8 claim: **0 matches** (remaining Opus 4.8 mentions are model-specific facts/benchmarks, not alias claims). F4 `/deep-research` auto-invoke claim: **0 matches**. 0-match evidence recorded per REQ-F-002 |
| AC-F-002 | PASS | `hugo --quiet` in docs-site | exit 0, no warnings |
| AC-X-001 | PASS | every `.claude/` edit paired with template mirror in same commit; `make build` | exit 0 each milestone (M1/M3/M4); harness.yaml + supersession note = documented local-only/sanitized deltas |
| AC-X-002 | PASS | `go test ./internal/template/ -run 'Leak\|Neutrality'` | exit 0; no SPEC ID/date/SHA introduced into templates (SPEC-ID template grep 0) |
| AC-X-003 | PASS (observation) | `git show origin/main:.claude/settings.json \| grep -c 'startup\|resume\|clear\|compact\|fork'` | 1 — PR #1146 fork matcher present on origin/main (ancestor 714270085); no GD-3 edit made |
| AC-X-004 | PASS | `./bin/moai spec lint \| grep CC2219-UPSTREAM` | 0 findings for this SPEC (repo-wide grandfathered warnings excluded) |
| AC-X-005 | PASS | `grep -c 'does not run two write-capable agents concurrently' CLAUDE.md templates/CLAUDE.md` | 1 / 1 — concurrency safeguard prose survived verbatim |

Edge case D.1 (`opus[1m]` picker): **kept** for back-compat (decision recorded as code comment in `ModelAliasPickerValues`).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25
run_commit_sha: 5f47749af
run_status: audit-ready
ac_pass_count: 27
ac_fail_count: 0
preserve_list_post_run_count: 0 violations (scope held to plan.md §D; no .moai/specs historical artifacts edited beyond this SPEC dir)
l44_pre_commit_fetch: n/a (worktree branch; base origin/main 714270085 verified at spawn)
l44_post_push_fetch: n/a (push deferred per orchestrator instruction)
new_warnings_or_lints_introduced: 0 (spec lint 0 findings for this SPEC; neutrality guards green)
cross_platform_build:
  darwin: exit 0 (go build ./...)
  windows: exit 0 (GOOS=windows GOARCH=amd64 go build ./...)
total_run_phase_files: 55 (M1 14 + M2 6 + M3 25 + M4 16 + M5 4, minus overlaps; per-commit stats in §E.2)
m1_to_mN_commit_strategy: one commit per milestone (M1..M5) + M6 evidence commit; no push, no PR (deferred to sync/manager-git)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-25
sync_commit_sha: pending-backfill-cc2219-sync
sync_status: audit-ready
changelog_entry_position: Unreleased > Added (single entry, top of section)
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (merged 3-phase close; no separate implemented intermediate commit)"
b12_self_test_a: "grep -c 'SPEC-CC2219-UPSTREAM-ALIGN-001' CHANGELOG.md (pre-emission) -> 0; no duplicate entry risk"
b12_self_test_b: "acceptance.md AC bullet count (27 unique AC IDs) matches progress.md §E.3 ac_pass_count: 27"
b12_self_test_c: "file paths cited in CHANGELOG entry verified via ls/grep: internal/template/model_policy.go, internal/web appbar, CLAUDE.md, .claude/rules/moai/development/{agent-authoring,agent-patterns,skill-authoring}.md, .claude/rules/moai/workflow/{orchestration-mode-selection,dynamic-workflows,native-invocation-model,hooks-system}.md — all present"
```
