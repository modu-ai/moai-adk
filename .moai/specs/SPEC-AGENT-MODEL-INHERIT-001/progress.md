# Progress — SPEC-AGENT-MODEL-INHERIT-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1246. Branch `WT-agent-model-inherit`, base develop `d6992e3a0`. Tier L.
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md; run-entry touch set `.moai/reports/t1246/touch-set.{sh,txt}`.
- Inventory measured 2026-09-26 (research.md). Budget baseline: 77539 / 77600 tokens, headroom 61.
- Run gate: `git merge-base --is-ancestor WT-rules-diet develop` exit 0 (REQ-AMI-001; exit 1 on 2026-09-26); t1257 overlap re-measured at run entry (REQ-AMI-002; C-collated intersection 0 at `024b95f77`, touch set 272 paths).
- Operator questions — RESOLVED 2026-09-26 (answered by the operator in the lane window, relayed by the coordinator):
  - Q1 — web console: delete the agent-settings tab entirely (UI + API). → REQ-AMI-011, design D7.
  - Q2 — leftover user config keys: `moai update` removes them and lists them in the update report. → REQ-AMI-014, design D8/D14.
  - Q3 — extra scope included: remove `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and dynamic-workflow `agent()` model/effort values. → REQ-AMI-007, REQ-AMI-013, design D12.
  - Q4 — `init`/`update --profile`: keep the flag, no-op, emit a deprecation warning. → REQ-AMI-016, design D10; extended by measurement to `--model-policy` / `--high` / `--medium-alias` / `--low` (design D13).
  - Q5 — harness v4 manifest specialist `model`/`effort`: optional; `/moai:harness` stops emitting them on generation; existing manifests still parse. → REQ-AMI-003, REQ-AMI-006, design D11.
  - Q6 — docs-site: in scope as M8 of this SPEC. → REQ-AMI-025.
  - Q7 — RESOLVED 2026-09-26 (operator, lane window, relayed by the coordinator): the init/update wizard "agent model policy" question is DELETED; the main-session policy stays in `moai profile setup`. → REQ-AMI-015, design D13.
- plan_status: audit-ready — plan phase closed 2026-09-26 on plan-audit iter-5 PASS-WITH-DEBT.
- plan-audit verdict history: iter-1 FAIL 0.76 → iter-2 FAIL 0.82 → iter-3 FAIL 0.83 → iter-4 FAIL 0.86 → iter-5 **PASS-WITH-DEBT 0.90** (Tier L threshold 0.85). Iterations 4 and 5 ran beyond the Tier L 3-iteration cap on explicit operator approval.
- plan-audit iter-5: PASS-WITH-DEBT 0.90 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-5.md`, commit `b46ea465d`). No SPEC body edit after this verdict; the debt below is carried into run entry.
- plan-audit iter-4: FAIL 0.86 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-4.md`, commit `4cc702dcf`) — blocked by V1/V2; V1, V2 and V5 addressed in commit `077c4f8f6`.
- Run-entry debt (resolve at M0, before the first edit commit; record resolution in §E.2):
  - W1 — `internal/cli/init_test.go:419` `TestValidateInitFlags_InvalidProfile` and the invalid-value half of the `ModelPolicyVocabulary` test at `:476` assert rejection of an invalid `--profile` / `--model-policy` value, which design D10 retires (the value is no longer validated); both will fail at M4. At run entry add `internal/cli/init_test.go` to the touch set (`.moai/reports/t1246/touch-set.sh`) and to the plan §G test-file table under M4 (adapt: expect the deprecation warning and exit 0).
  - W2 — `research.md:L314` states the host-(b) version-match predicate without the `verr == nil` conjunct; the authoritative form is `verr == nil && packageVersion == projectVersion && !forceUpdate` (update_template_sync.go:771-773, as design D14 states). Implement against D14.
  - W3 — `plan.md:L93` cites `wizard/translations.go:89-91` / `:179-181` / `:269-271` for the init-wizard question translations; those ranges stop at the opening lines — each `"model_policy": {` block runs to its closing brace. Remove the whole block per locale.
  - Optional carry-over (plan-audit iter-4): V3 — dead agentfm selectors in `internal/web/assets/app.js:390-480` (`wireProfileMatrix`, `reapplyHaikuLocks`, `wireHaikuEffortLock`) are residue once M2 removes the panel; remove them in M2 if in scope, otherwise note in §E.2. V4 — AC-AMI-006's "no finding for `user-pinned.md`" depends on fixture fields (LR-05 warns on a missing `isolation:`); give the fixture the fields that avoid LR-05, or read the clause as "no LR-03/LR-12/LR-13 finding and no error".
- plan-audit iter-3: FAIL 0.83 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-3.md`, commit `d8261313c`). The operator explicitly approved ONE extra revision round beyond the Tier L 3-iteration cap (2026-09-26, lane window, relayed by the coordinator). R1–R9 addressed in spec v0.5.0.
- plan-audit iter-2: FAIL 0.82 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-2.md`, commit `c87803153`). N1–N9 addressed in spec v0.4.0.
- plan-audit iter-1: FAIL 0.76 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-1.md`, commit `5faf93bb8`). D1–D16 addressed in spec v0.3.0 (see spec.md HISTORY).
- Counts: 25 requirements (REQ-AMI-001..025, contiguous), 25 acceptance criteria (AC-AMI-001..025) — both at the Tier L ceiling of 25.

## §E.2 Run-phase Evidence

### M0 — run entry and baseline (2026-09-27, measured on HEAD `5509ea71e`)

Tree: branch `WT-agent-model-inherit`, HEAD `5509ea71e` = merge of `dc8d18bcd` (plan close) and
develop `7fe658815` (the absorbed develop SHA). No file edited before these measurements; raw
outputs under `.moai/state/verify/t1246/` (local, not committed).

Run gate (REQ-AMI-001 / AC-AMI-001):

| Command | Output |
|---|---|
| `git merge-base --is-ancestor WT-rules-diet develop; echo $?` | `rules-diet-ancestor exit=0` (`WT-rules-diet` = `8fb81c948`, develop = `7fe658815`) |

Overlap gate (REQ-AMI-002 / AC-AMI-002):

| Command | Output |
|---|---|
| `sh .moai/reports/t1246/touch-set.sh > .moai/state/verify/t1246/touch-set.now.txt` | exit 0, **272** paths; `LC_ALL=C diff` against the committed `touch-set.txt` → exit 0 (byte-identical); `LC_ALL=C sort -c` → exit 0 |
| `git merge-base HEAD WT-role-naming-docs` | `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c` (`WT-role-naming-docs` tip `637513578`) |
| `git diff --name-only b59a5d69c… WT-role-naming-docs \| LC_ALL=C sort` | 40 paths: 34 under `.moai/reports/t1257/`, 6 under `.moai/specs/SPEC-ROLE-NAMING-DOCS-001/` |
| `LC_ALL=C comm -12 t1257.txt touch-set.now.txt \| wc -l` | **0** — gate open |

Touch-set distribution (`cut -d/ -f1-2 | LC_ALL=C sort | uniq -c`): 22 `.claude/agents`,
2 `.claude/commands`, 1 `.claude/hooks`, 12 `.claude/rules`, 8 `.claude/skills`, 5 `.claude/workflows`,
1 `.moai/docs`, 2 `.moai/project`, 1 `CHANGELOG.md`, 62 `docs-site/content`, 38 `internal/cli`,
12 `internal/config`, 10 `internal/harness`, 6 `internal/hook`, 5 `internal/settings`, 2 `internal/spec`,
54 `internal/template`, 29 `internal/web` — identical to the plan-time distribution (research.md §I).

Re-measured inventory (research.md §B–§K), this tree:

| Item | Command (abridged) | Now | Plan-time |
|---|---|---|---|
| local moai agent frontmatter | `grep -rlE / -rnE '^(model\|effort):' .claude/agents/moai` | 12 files / 24 lines | 12 / 24 |
| local harness agent frontmatter | same over `.claude/agents/harness` | 10 / 20 | 10 / 20 |
| template agent frontmatter | same over `internal/template/templates/.claude/agents` | 12 / 24 | 12 / 24 |
| codex toml `model` / `model_reasoning_effort` lines | `grep -cE '^model ' / '^model_reasoning_effort'` over the 12 toml | 0 / 12 | 0 / 12 |
| `ResolveAgentModelEffort` files (non-test / test, definer included) | `grep -rlF … --include='*.go' internal cmd pkg` | 9 / 9 | 8 consumers + definer |
| `ResolveHarnessAgentModelEffort` | same | 3 / 2 | 3 |
| `DefaultProfileMatrix` | same | 4 / 7 | 3 + definer |
| `ProfileMatrixAgents` | same | 7 / 4 | 6 + definer |
| `AgentGroup(` | same | 3 / 0 | 2 + definer |
| `ApplyProfile` | same | 5 / 1 | 4 + definer |
| `EffectiveProfile` | same | 6 / 1 | 5 + definer |
| `ResolveGLMReasoningForModel` | same | 3 / 2 | 2 + definer |
| `retainedAgentNames` | same | 2 / 1 | profile.go + registry.go |
| guard surface `.go` files | `grep -rlE 'agent_model_guard\|AgentModelGuard\|agent-model-audit' --include='*.go'` | 11 | 11 (research §C list) |
| `agent_model_guard` in template YAML | `grep -rnE … --include='*.yaml' internal/template/templates` | 0 | 0 |
| resolver/guard test pattern | `grep -rlE 'ProfileMatrix\|ResolveAgentModelEffort\|agent_overrides\|AgentModelGuard\|agentModel\|profile_matrix' --include='*_test.go'` | 29 | 29 |
| `CLAUDE_CODE_SUBAGENT_MODEL` non-test hits | `grep -rnE … internal cmd pkg \| grep -v _test.go` | 1 (scrub list) | 1 |
| AC-AMI-007(d) `judge_effort` files | `grep -rlE 'judge_effort\|JUDGE_EFFORT' …` | 3 | 3 |
| AC-AMI-007(c) prose hits | research §F regex | 6 | 6 |
| docs-site pages (AC-AMI-025 pattern) | research §F narrowed pattern | 52 | 52 |
| `[HARD]` lines naming effort/model | `grep -rnE '\[HARD\].*(effort\|model)' .claude/rules` | 17 | 17 |

Always-loaded budget baseline (REQ-AMI-024 / AC-AMI-024 "before"):

```
$ unset MOAI_KANBAN … && go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v
    token_budget_guard_test.go:70: always-loaded surface = 65591 tokens (budget 77600, headroom 12009, 16 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.152s
```

Before-headroom re-baselined after t1175: **12009** (plan-time 61 at `d6992e3a0`). The test lives in
`internal/config`; `go test ./internal/template/ -run TestAlwaysLoadedTokenBudget` reports
`[no tests to run]`.

Run-entry debt resolution:

- **W1 — resolved by record.** `internal/cli/init_test.go` is added to the run touch set as an
  explicit addendum (touch set for this run = the 272 generated paths + `internal/cli/init_test.go`
  = 273; `grep -c 'internal/cli/init_test.go' touch-set.now.txt` → 0 before the addendum). The
  generator script and plan.md §G are plan artifacts and are not edited in run; the M4 test-file
  row is carried here instead: at M4, `TestValidateInitFlags_InvalidProfile` (:419), the invalid
  half of `TestValidateInitFlags_ModelPolicyVocabulary` (:451) and `TestInitCmd_ProfilePersistence`
  (:490, asserts `profile:` persisted) are adapted to expect the deprecation warning, exit 0, and no
  `profile:` / `performance_tier:` write (D10/D13).
- **W2** — carried to M1 host (b): implemented against design D14's
  `verr == nil && packageVersion == projectVersion && !forceUpdate`.
- **W3** — carried to M4 (whole `"model_policy": {…}` block per locale).
- **V3 / V4** — carried to M2 / the AC-AMI-006 fixture.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
