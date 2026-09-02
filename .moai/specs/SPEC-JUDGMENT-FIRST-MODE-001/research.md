# SPEC-JUDGMENT-FIRST-MODE-001 — Research

All findings below were measured read-only in `.claude/worktrees/t401` at commit `ad272be20`
(= `origin/develop`). Source reports: `.moai/reports/t401/preflight.md`,
`.moai/reports/t401/lens-recommendation-surfaces.md`, `.moai/reports/t401/lens-audit-contract.md`.

## 1. Surface inventory

77 recommendation-emission coordinates were collected across `.claude/rules/moai/**`,
`.claude/output-styles/moai/*.md`, `.claude/agents/moai/*.md`, `.claude/skills/moai/**`,
`.moai/config/sections/*.yaml`, and `internal/config/*.go`, classified:

- **A — emission mandated**: the coordinates the SPEC's six surfaces are drawn from, plus the
  auditor and advisor output contracts.
- **B — permitted**: emission allowed but not required.
- **C — existing anti-anchoring guards**: the ten clauses that already constrain how a
  recommendation may be phrased.

Class C is why this SPEC adds an axis rather than a guard: the guards exist and are good. What is
missing is the ability to withhold, not the ability to phrase neutrally.

## 2. Negative finding 1 — no judgment-first construct exists

A grep sweep for `anchor(ing)`, `independent judg(e)ment`, `withhold`, `push, not pull`,
`pull, not push`, `unprompted`, `solicit`, `on request only`, `only when asked`, and
`권장|recommend|⏭️|prescri|non-binding` returned **zero** judgment-first or
recommendation-withholding constructs.

The four nearest clauses each put **evidence** before the recommendation, never **user judgment**:

| Clause | What it does | Why it is not judgment-first |
|---|---|---|
| Report-Before-Ask `[HARD]` | report precedes the question | same turn — the user's first act is still choosing among labeled options |
| Requested-Deliverable Primacy `[HARD]` | ends the turn on a requested report | fires only when the user explicitly asked for a report; does not suppress the `[HARD]` Discovery `⏭️ Recommended action:` field |
| Adaptive strength (principle 5) | can omit the label | gated on **estimated proficiency**, a system inference — not user judgment |
| Implementation Kickoff Approval | mandatory human gate | approves options that already carry the label |

Adaptive strength is the only existing mechanism able to suppress the label. That is why the design
generalizes it rather than adding a sibling.

## 3. Negative finding 2 — no config key controls emission

All 33 files under `.moai/config/sections/` and every `internal/config/*.go` were read. Nothing
toggles recommendation emission. The nearest keys control something else:

| Key | Actually controls |
|---|---|
| `learning.auto_apply` | whether Tier-4 proposals auto-apply — false makes recommendations *more* visible |
| `interview.enabled` / `plan.max_rounds` / `skip_conditions` | round counts; removing rounds removes the label incidentally, not by intent |
| `AuditGateOff/Advisory/Required` | whether an audit **blocks**; the auditor's `### Recommendations` emits regardless |
| `gate.yaml` advisory, `sunset.yaml action: advisor` | quality-gate blocking mode |
| `ralph.yaml` | auto-apply without confirmation |
| `mcp-matrix.yaml` | a static recommendation matrix — data, not a toggle |

## 4. Precedent for the flag shape

Three landed default-off flags were read for wiring shape:

| Flag | Type | Default | Local dogfood |
|---|---|---|---|
| `Workflow.BranchGuard.Enabled` | `BranchGuardConfig` in `internal/config/types.go`, default in `defaults.go` | false | on |
| `workflow.integration_lock.enabled` | `IntegrationLockConfig`, same shape | false | on |
| `workflow.agent_stop_guard.enabled` | `AgentStopGuardConfig`, plus a default-false pin test and a loader round-trip test | false | on |

All three gate a **deny**. The recommendation mode gates an **output convention** and denies
nothing, which is why it is sited on `InterviewConfig` rather than `WorkflowConfig`. The test shape
is copied directly from `internal/config/workflow_agent_stop_guard_test.go`: a shipped-default pin
plus a section-file round-trip.

`InterviewConfig` today carries `ClarityThreshold`, `Enabled`, `Plan`, `Project`, and
`SkipConditions`; `defaultInterviewConfig()` mirrors the template file. Adding one string field
plus its default is additive.

## 5. Hook feasibility — pre-verified

- `.claude/settings.json` registers three `PreToolUse` matcher blocks (`Write|Edit|Bash`,
  `Agent|Task`, `SendMessage|TaskStop`), all routing to `handle-pre-tool.sh`. The matcher is a
  tool-name regex, so an `AskUserQuestion` block is the same shape as the existing
  `SendMessage|TaskStop` block.
- `internal/hook/pre_tool.go` already carries the two sibling observation branches (`Agent|Task`
  at the agent-model check, `SendMessage` at the stop-guard check). Both sit after every
  established deny, both parse `input.ToolInput`, both contribute only a `SystemMessage`.
- `internal/hook/agent_stop_guard.go` demonstrates `input.ToolInput` payload extraction and JSONL
  audit appending with fail-open on every uncertainty.

Two mirrors are load-bearing: `.claude/settings.json` is rendered from
`internal/template/templates/.claude/settings.json.tmpl`, and `handle-pre-tool.sh` exists as
`.claude/hooks/moai/handle-pre-tool.sh` plus `…/handle-pre-tool.sh.tmpl`.

## 6. Static-guard precedent

`.github/workflows/template-neutrality-check.yaml` is the shape to copy: path-filtered
`pull_request` and `push` triggers, a `concurrency` group with `cancel-in-progress`,
`permissions: contents: read`, and an **isolated** test target so unrelated pre-existing failures
in the same package do not mask the guard's verdict.

**Mirror status of the new workflow — corrected premise.** An earlier draft of this section
asserted that `.github/` is not part of the distributed template. Measured, that is false:
`find internal/template/templates/.github -type f` returns four files —
`actions/detect-language/action.yml`, `branch-protection.json.gtmpl`, `labels.yml`, and
`workflows/label-sync.yml`. So the template does carry a `.github/` subtree, and it carries exactly
one workflow. What is true, and what the conclusion actually rests on, is narrower: this
repository's own 21 CI workflows under `.github/workflows/` are unmirrored — `label-sync.yml` is
the sole template-side workflow, and it is a distributed-project scaffold rather than a mirror of
a moai-adk CI job. `judgment-first-consistency.yaml` is a repository-local guard over this
repository's own doctrine tree, so it has **no template counterpart**, and REQ-JFM-021 binds only
files "that ha[ve] a template counterpart". The file therefore takes no mirror — by the
counterpart clause, not by a blanket `.github/` exemption that does not exist.

## 7. The plan-audit exposure measurement

The card's premise ("observation and recommendation always emitted together") was tested against
`plan-auditor.md` and `.claude/skills/moai/plan/spec-assembly.md`:

- The report's required sections end with `## Recommendation`, and there is **no omission branch**.
- The defect record format itself embeds `Required fix:` — a defect without one does not exist in
  the format.
- But on **iter-1 PASS** the report body is never shown (one log line only), and on **iter-1~2
  FAIL** it routes straight back to `manager-spec`. The user sees the auditor's recommendations on
  exactly three paths: the Kickoff-turn HTML, the 3×FAIL escalation, and a run-gate block.
- `plan-auditor`'s `tools:` list contains no `AskUserQuestion` — it is structurally unable to ask.

Hence the correction recorded in `spec.md` §A.3, and the scoping decision that follows from it.

## 8. Two concerns from the issue that are already satisfied

- **`zero-flag ≠ pass`.** No clause says "zero defects = PASS". PASS requires all must-pass
  PASS/N-A **and** score ≥ tier threshold **and** an evidence citation on every PASS. Zero defects
  satisfies none of those. A counter-guard already exists in the other direction ("a long list of
  optional findings does not by itself justify a FAIL").
- **Verdict vocabulary.** `PASS | FAIL` are the only auditor tokens; `UNVERIFIED` counts as FAIL
  at must-pass; `PASS-WITH-DEBT` is an orchestrator-offered user option, not an auditor token.

Both are excluded in `spec.md` §F: restating an already-binding property is text without effect.

## 9. Tier evidence

| Signal | Measurement | Tier implication |
|---|---|---|
| Files affected | ~20 (6 local doctrine/config/settings + 6 template mirrors + 4 Go files incl. 2 new + 2 new Go test files + 1 CI workflow) | > 15 ⇒ L |
| LOC | doctrine prose across 6 files plus an estimated 300-500 LOC of Go (observer, config field, tests) | mid-band; not decisive alone |
| Constitutional | touches a Frozen-zone clause (`CONST-V3R5-035`, `canary_gate: true`) and a `[HARD]` clause in the always-loaded `askuser-protocol.md` | explicit L trigger |

Tier **L**, on the file-count signal and the constitutional signal independently. PASS threshold
0.85; REQ ceiling 25 (23 used); AC ceiling 25 (22 used).
