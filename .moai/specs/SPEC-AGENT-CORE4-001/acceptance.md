---
id: SPEC-AGENT-CORE4-001
title: "Acceptance: agent core four"
version: "0.1.0"
created: 2026-10-18
---

# acceptance.md — SPEC-AGENT-CORE4-001

Document-level tree pin: every RED-now cell below was measured on tree `ae34f26831`
(branch `WT-3-3-10`, clean) and binds every criterion that carries no pin of its own.
Classes: **blocking** (release-blocking, RED-now observed) and **guard** (regression-guard).
KB = 1,000 bytes. A `go test ... -v` criterion is green only when the output names the listed
`--- PASS:` lines; `ok ... [no tests to run]` is red by the empty-sweep rule
(`verification-completeness.md` §1.1) — that is the RED-now form for every test-selector row,
observed in this run (ledger L-12).

## §1 Acceptance criteria

Each criterion: Given/When/Then, the single-invocation command, the green output, and the
RED-now ledger id. Green path = the milestone (plan.md) that flips it.

**AC-AC4-001** (blocking; REQ-AC4-005, 002) — Given the planner and auditor are rendered, When the
plan pair is measured, Then the injected bytes are at most 56,000.
Command: `moai agent size --phase plan --json`. Green (M3, measured from in-memory renders; copies agree from M8): exit 0, JSON with
`"phase":"plan"`, `"total_bytes"` ≤ 56000, `"within_budget":true`, and components listing
planner, auditor, moai-agent-spec, moai-agent-audit, and `memory_cap`. RED-now: L-01.

**AC-AC4-002** (guard; REQ-AC4-005) — Given documenter and auditor, When the sync pair is measured,
Then at most 46,000 bytes. Command: `moai agent size --phase sync --json`. Green (M3): exit 0,
`"within_budget":true`. RED-now: L-01 (same missing verb). Not stage-completion: zero headroom
(design §3).

**AC-AC4-003** (blocking; REQ-AC4-004) — Given the neutral specs, When C1 is compared with C2,
Then no file differs. Command: `diff -rq .claude/agents/moai internal/template/templates/.claude/agents/moai`.
Green (M8 end state, intermediate M2 for the carried four): empty stdout, exit 0. RED-now: L-02.

**AC-AC4-004** (blocking; REQ-AC4-004) — When the project Codex mirror is compared with the
template Codex set, Then they are identical. Command:
`diff -rq .codex/agents/moai internal/template/templates/.codex/agents/moai`. Green (M8): empty,
exit 0. RED-now: L-03.

**AC-AC4-005** (blocking; REQ-AC4-004) — When the read-only emission check runs, Then it reports
all four copies equal to a fresh render. Command: `make agents-emit-check`. Green (M1 extended, M8
final): stdout contains `specs=9 copies=4 diff=0`, exit 0. RED-now: L-04 (passes today with no
summary line).

**AC-AC4-006** (blocking; REQ-AC4-004) — Given a one-byte edit to a rendered copy, When the check
runs, Then it fails and names the file. Command:
`go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1`. Green (M1):
`--- PASS: TestEmitCheckFailsOnHandEdit ` (name followed by a space) with subtests `claude_project`, `claude_template`,
`codex_template`, `codex_project`. RED-now: L-12.

**AC-AC4-007** (blocking; REQ-AC4-001, 003) — When the template Codex directory is listed, Then it
holds exactly the nine resulting-roster files (the arithmetic is in §3). Command:
`ls internal/template/templates/.codex/agents/moai`. Green (M8): `auditor.toml builder-harness.toml
documenter.toml e2e-tester.toml implementer.toml manager-design.toml planner.toml super-advisor.toml
worker.toml`, exit 0. RED-now: L-03 (12 files, none of the five new).

**AC-AC4-008** (blocking; REQ-AC4-003) — When the project Codex bodies are scanned for Claude-only
notation, Then every file reports 0. Command:
`grep -cE '\.claude/|Skill\(|Agent\(' .codex/agents/moai/*.toml`. Green (M8; the new planner/implementer/documenter/worker files are 0 from M3): nine lines, each
ending `:0`. RED-now: L-05 (11 files, counts 7 to 27).

**AC-AC4-009** (blocking; REQ-AC4-003, 018) — When the planner Codex definition is read, Then it
carries its skill body inline exactly once. Command:
`grep -c 'moai-agent-spec:begin' .codex/agents/moai/planner.toml`. Green (M3): `1`, exit 0.
RED-now: L-06 (file absent, exit 2).

**AC-AC4-010** (blocking; REQ-AC4-006) — Given the acting agent is the auditor, When it writes a
file, Then only targets under the tree's `.moai/reports/` pass. Command:
`go test ./internal/hook -run '^TestAuditorWriteGuard$' -v -count=1`. Green (M4): PASS lines for
`allows_verdict_path`, `allows_nested_report_dir`, `denies_source_path`, `denies_spec_path`,
`denies_dotdot_traversal`, `denies_symlink_escape`, `denies_prefix_lookalike_reportsX`,
`denies_agent_memory_path`, `ignores_other_agents` (negative control: an implementer may write
code), `fails_open_when_path_unparseable`, `sentinel_prefix_AUDITOR_WRITE_VIOLATION`. RED-now: L-12.

**AC-AC4-011** (blocking; REQ-AC4-007) — When the auditor's Codex render is emitted, Then it is
read-only, `write-path-scope` is recorded UNSUPPORTED (not enforced), and no row is silently
dropped. Command: `go test ./internal/template/agentemit -run '^TestAuditorCodexContract$' -v -count=1`.
Green (M3): PASS lines `sandbox_read_only`, `write_path_scope_unsupported_recorded`,
`no_contract_row_dropped`. RED-now: L-12.

**AC-AC4-012** (blocking; REQ-AC4-003, 001) — When the auditor's rubric is compared with the
retired plan/sync auditor definitions, Then every rubric heading survives in the auditor or its
references. Command: `go test ./internal/template/agentemit -run '^TestAuditorRubricHeadingsPreserved$' -v -count=1`.
Green (M3): `--- PASS` with a subtest per old heading set (`plan_rubric`, `plan_checklist`,
`sync_protocol`) and a non-zero heading count logged. RED-now: L-12.

**AC-AC4-013** (blocking; REQ-AC4-008) — When a write would leave an agent MEMORY.md above 8,000
bytes, Then it is refused; at exactly 8,000 it passes; the auditor renders no `memory:` field; the
doctor flags over-cap files. Command:
`go test ./internal/hook -run '^TestAgentMemoryCap$' -v -count=1`. Green (M4): PASS lines
`write_8000_allowed`, `write_8001_denied_with_MEMORY_CAP_VIOLATION`, `edit_applied_result_over_cap_denied`,
`other_agent_dir_same_rule`, `topic_file_not_capped`, `auditor_has_no_memory_field`,
`doctor_flags_over_cap`. RED-now: L-12.

**AC-AC4-014** (blocking; REQ-AC4-009) — When an agent is spawned, Then the audit row carries
agent, phase, effort, effort_source and prompt_bytes, never the prompt text, and no unattributed
stop row is written. Command: `go test ./internal/hook -run '^TestAgentSpawnAuditRecord$' -v -count=1`.
Green (M4): PASS lines `fields_present`, `auditor_phase_from_prompt_line`, `auditor_phase_unknown`,
`effort_source_declared_frontmatter_session`, `prompt_text_absent`, `unattributed_stop_not_written`.
RED-now: L-12 (and `grep -c 'prompt_bytes' internal/hook/agent_model_guard.go` → `0`, exit 1).

**AC-AC4-015** (blocking; REQ-AC4-010) — When the project worker definition is read, Then its
tools line grants no spawn capability. Command: `grep -E '^tools: ' .claude/agents/moai/worker.md`.
Green (M3): exactly `tools: Read, Write, Edit, Bash, Grep, Glob`, exit 0. RED-now: L-06 (absent,
exit 2).

**AC-AC4-016** (blocking; REQ-AC4-012) — When an Agent spawn names a retired name after cutover,
Then it is denied with the sentinel and the table's replacement; every other name passes. Command:
`go test ./internal/hook -run '^TestRetiredAgentSpawnRefused$' -v -count=1`. Green (M8): PASS for one
subtest per retired name (eight), each asserting `RETIRED_AGENT_REJECTED:` plus the replacement
text from `retired_names.yaml`; plus `planner_passes`, `worker_passes`, `explore_passes`,
`general_purpose_passes`, `unknown_name_fails_open`, `payload_not_rewritten`. RED-now: L-12.

**AC-AC4-017** (blocking; REQ-AC4-012) — When the rejection rule is read, Then it carries one table
row per retired name. Command:
`grep -cE '^\| (13|14|15|16|17|18|19|20) \|' .claude/rules/moai/workflow/archived-agent-rejection.md`.
Green (M7): `8`, exit 0. RED-now: L-07 (`0`, exit 1).

**AC-AC4-018** (blocking; REQ-AC4-013, 020) — When the tree is swept for retired names, Then every
hit lies in an allowlisted path with a reason, and the docs-site and README entries name their
owner card. Command: `go test ./internal/template -run '^TestRetiredNamesOnlyInAllowlist$' -v -count=1`.
Green (M7, M8): `--- PASS` subtests `rules`, `skills`, `hooks`, `config`, `go_production`,
`template_mirrors`, `tests_and_fixtures`, `allowlist_entries_have_reason`; reported count of
non-allowlisted hits `0` (baseline 977 files, research §3). RED-now: L-12.

**AC-AC4-019** (blocking; REQ-AC4-014) — Given a universe of ten names, When a file lists seven of
them, Then the sweep reports it; at today's thirteen the threshold stays 10. Command:
`go test ./internal/harness/rosterguard -run '^TestSweepThresholdScalesWithUniverse$' -v -count=1`.
Green (M3): PASS `threshold_13_is_10`, `threshold_10_is_7`, `seven_of_ten_listing_found`,
`six_of_ten_listing_not_found`. RED-now: L-12.

**AC-AC4-020** (blocking; REQ-AC4-015) — When ownership is linted, Then trailers `planner`,
`implementer`, `documenter` are accepted for their predecessors' transitions and legacy values
remain accepted. Command: `go test ./internal/spec -run '^TestOwnershipTrailerAcceptsCoreAndLegacy$' -v -count=1`.
Green (M7): PASS `core_names_accepted_per_transition`, `legacy_names_still_accepted`,
`wrong_owner_still_flagged`, `manager_git_stays_ownerless`. RED-now: L-12.

**AC-AC4-021** (blocking; REQ-AC4-017) — When the skills check runs, Then it reports no exposed
`moai-agent-*` skill. Command: `moai doctor --check skills`. Green (M6): output contains
`moai-agent exposure: 0 exposed`, exit 0. RED-now: L-08 (no such line).

**AC-AC4-022** (blocking when a live probe ran; REQ-AC4-018, 019) — When the preload probe evidence
is read, Then it records the shipped mode and a passing reply. Command:
`grep -c 'PROBE_RESULT: preload=honored' .moai/reports/t1649/preload-probe.md`. Green (M6): `1`,
exit 0, with the file also naming mode, Claude Code version, nonce, and reply. If no live probe
could run, the file records `PROBE_RESULT: not-run` and the shipped mode is H3 (decision Q9); the
criterion then reads `grep -c 'PROBE_RESULT: not-run'` → `1` and is a guard. RED-now: L-09
(file absent).

**AC-AC4-023** (blocking; REQ-AC4-011, 020) — When the SPEC's commits are listed in graph order,
Then the baseline, replacement-path verification and removal appear in that order. Command:
`git log --reverse --format=%s --grep=SPEC-AGENT-CORE4-001`. Green (M8): subjects in order
`feat(SPEC-AGENT-CORE4-001): M0 baseline`, `... M1 ...` through `... M6 ...`, `... M7a/M7b/M7c ...`,
`... M8 ...`, with M5 before M8 and M0 the first (the baseline lands alone, `verification-claim-integrity.md` §2.3). RED-now: L-10 (empty).

**AC-AC4-024** (blocking; REQ-AC4-010) — When the run workflow is read, Then it carries the large-scope
fan-out clause. Command: `grep -c '^## Tier L Fan-Out' .claude/skills/moai/workflows/run.md`.
Green (M5): `1`, exit 0. RED-now: L-11 (`0`, exit 1).

## §2 Evidence ledger (RED-now, tree ae34f26831)

```text
L-01  cmd: moai agent size --phase plan --json
      stdout: ERROR / Unknown flag: --phase. / Try --help for usage.
      exit: 1   (reason: the `size` verb does not exist; `moai agent` has `lint` only)
L-02  cmd: diff -rq .claude/agents/moai internal/template/templates/.claude/agents/moai
      stdout: 7 lines "Files ... differ": builder-harness.md e2e-tester.md manager-design.md
              manager-develop.md manager-spec.md plan-auditor.md super-advisor.md
      exit: 1   (reason: install-root path prefix and one content drift, research §2)
L-03  cmd: diff -rq .codex/agents/moai internal/template/templates/.codex/agents/moai
      stdout: 7 "differ" lines + "Only in internal/template/templates/.codex/agents/moai: manager-todo.toml"
      exit: 1
      cmd: ls internal/template/templates/.codex/agents/moai
      stdout: builder-harness e2e-tester manager-design manager-develop manager-docs manager-git
              manager-lead manager-spec manager-todo plan-auditor super-advisor sync-auditor (.toml) = 12
      exit: 0   (reason: pre-change roster; none of the nine target names but five)
L-04  cmd: make agents-emit-check
      stdout: ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.305s
      exit: 0   (reason red: the required summary line `specs=9 copies=4 diff=0` is absent)
L-05  cmd: grep -cE '\.claude/|Skill\(|Agent\(' .codex/agents/moai/*.toml
      stdout: manager-design 10, builder-harness 17, manager-develop 26, e2e-tester 7, manager-git 9,
              manager-lead 24, manager-docs 17, manager-spec 27, super-advisor 10, sync-auditor 12,
              plan-auditor 17
      exit: 0   (reason red: counts non-zero)
L-06  cmd: grep -c 'moai-agent-spec:begin' .codex/agents/moai/planner.toml
      stdout: ugrep: warning: .codex/agents/moai/planner.toml: No such file or directory
      exit: 2
      cmd: grep -E '^tools: ' .claude/agents/moai/worker.md
      stdout: ugrep: warning: .claude/agents/moai/worker.md: No such file or directory
      exit: 2
L-07  cmd: grep -cE '^\| (13|14|15|16|17|18|19|20) \|' .claude/rules/moai/workflow/archived-agent-rejection.md
      stdout: 0
      exit: 1
L-08  cmd: moai doctor --check skills
      stdout: System Diagnostics box, "Pass 0 Warn 0 Fail 0"; no `moai-agent exposure` line
      exit: 0   (reason red: required line absent)
L-09  cmd: ls .moai/reports/t1649/preload-probe.md
      stdout: ls: .moai/reports/t1649/preload-probe.md: No such file or directory
      exit: 1
L-10  cmd: git log --reverse --format=%s --grep=SPEC-AGENT-CORE4-001   (run before the plan commit)
      stdout: (empty)
      exit: 0
L-11  cmd: grep -c '^## Tier L Fan-Out' .claude/skills/moai/workflows/run.md
      stdout: 0
      exit: 1
L-12  cmd: go test <pkg> -run '^<Selector>$' -v -count=1   for the ten selectors in AC-006, 010, 011,
            012, 013, 014, 016, 018, 019, 020 (packages hook, agentemit, template, rosterguard, spec)
      stdout: "testing: warning: no tests to run / PASS / ok <pkg> 0.4-0.9s [no tests to run]"
      exit: 0   (reason red: swept set is empty, so no listed `--- PASS:` line exists)
```

L-10 note: after the plan-phase commit of this SPEC the grep also matches that plan commit; the
green list in AC-AC4-023 is therefore read from the first `M0` subject onward.

## §3 The Codex count, reconciled (card text "TOML 12/12")

| Set | Count | Members |
|---|---:|---|
| Definitions at HEAD (C2) | 12 | builder-harness, e2e-tester, manager-design, manager-develop, manager-docs, manager-git, manager-lead, manager-spec, manager-todo, plan-auditor, super-advisor, sync-auditor |
| Retired outright | −3 | manager-lead, manager-todo, manager-git |
| Absorbed into the core four | −5 | manager-spec, manager-develop, manager-docs, plan-auditor, sync-auditor |
| New | +5 | planner, implementer, documenter, auditor, worker |
| **Resulting roster** | **9** | planner, implementer, documenter, auditor, worker, builder-harness, e2e-tester, manager-design, super-advisor |

12 − 3 − 5 + 5 = 9. The card's "12/12" cannot be the post-change file count under any reading of
the retire list; the only reading that yields 12 is the pre-change state (12 definitions, 12
fresh TOMLs — and even that is 11/12 in the project mirror today). This SPEC therefore measures
**9/9 fresh** (AC-AC4-004, 005, 007) and flags the discrepancy to the leader (decision-index Q2).
If the three carried agents (e2e-tester, manager-design, super-advisor) move to the seed catalog
in the sibling card, the count falls to 6.

## §4 Mutant closure (verification-completeness §2)

| Criterion | Mutant that would satisfy it while violating the requirement | Closed by |
|---|---|---|
| 001 | verb prints a constant under 56,000 | Go test cross-checks each component against `wc -c` of the named file (design §3) and fails if a component is missing from the list |
| 003 | delete the differing files so the diff is empty | 007 pins the file set; 005 pins `specs=9` |
| 005 | print the summary line unconditionally | 006 (hand edit → non-zero) |
| 008 | strip the tokens from the body by regex after render | 009 pins inline skill content; 018 sweeps sources |
| 010 | deny every write | `allows_verdict_path` subtests |
| 010 | allow every write | the seven deny subtests |
| 010 | key on a name other than the generated verdict-only set | `ignores_other_agents` + a table test that the set equals the specs' capability tokens |
| 013 | cap at 8,001 or per-line | boundary subtests at 8,000 / 8,001 |
| 016 | deny every spawn | the five pass-through subtests |
| 018 | allowlist everything | `allowlist_entries_have_reason` + a fixture that plants an unlisted hit and expects failure |
| 023 | rewrite history order | ordering is read from the commit graph, not from message text; M5 → M8 ancestry also asserted by `git merge-base --is-ancestor` in the M8 gate |

## §5 Edge cases

- A spawn names `Planner` or `PLANNER` (case): the runtime matches exact names; the hook table
  matches exact retired names only and never rejects a non-retired spelling.
- A named teammate spawn (`name` argument) of a retired agent: refused identically; the stop-guard
  registry is untouched.
- An auditor spawned without a `phase:` line: record phase `unknown`; the body returns a blocker
  report; the guard still applies.
- `MEMORY.md` write by the PostToolUse mirror inside a worktree: the cap applies to the tree being
  written; the mirror copies topic files, never the index.
- A symlink inside `.moai/reports/` pointing outside: denied by symlink resolution.
- `moai update` on a project that still holds old definitions: the deploy removes retired
  definitions only through the existing `moai update` managed-file path (design §6; Gap G-4).

## §6 Quality gates and Definition of Done

Quality gates (guards, run at every milestone commit; none may be skipped or relaxed):

| Gate | Command | Expected |
|---|---|---|
| Roster guard | `go test ./internal/harness/rosterguard/... -count=1` | `ok` (baseline `ok` 32.845s) |
| Emission and embed checks | `make agents-emit-check embed-manifest-check commands-emit-check` | exit 0 |
| No skipped test introduced | `git log -p --format= -G't\.Skip\(' ae34f26831..HEAD -- '*_test.go'` | empty stdout |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Coverage of touched packages | `go test -cover ./internal/hook/... ./internal/template/agentemit/... ./internal/harness/rosterguard/...` | ≥ 85% each, per quality.yaml |
| Template neutrality | `go test ./internal/template -run '^(TestTemplateNoInternalContentLeak|TestDeployedRoleRuleEntryPointMarkers)$' -count=1` | `ok` with non-empty sweep |

Definition of Done: all 24 criteria green except 002 and (when `not-run`) 022 recorded as guards;
the five quality gates green; `.moai/reports/t1649/` holds the M0 baseline and the M5 verification
and the preload probe; progress.md §E.2/§E.3 populated by the run phase; the `AGENTS.local.md`
§2.0 follow-up edit and the docs-site/README deferral are recorded as open items for the owning
cards.
