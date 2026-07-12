# research.md — Shared Investigation for Epic AGENTIC-CORE

> **Shared research artifact** for the 3-SPEC Epic **AGENTIC-CORE**
> (`SPEC-ANALYZE-FIRST-ROUTING-001`, `SPEC-GOAL-ENGINE-001`, `SPEC-LOOP-SWEEP-001`).
> Cross-referenced by the other two SPECs — do NOT duplicate this content there.
>
> Citations use content-token anchors plus approximate line numbers. Line numbers
> drift; the content token is the durable anchor. Every finding below was observed
> this session against the live tree; where a claim could not be mechanically
> confirmed it is flagged as an assumption, per
> `.claude/rules/moai/core/verification-claim-integrity.md`.

## §A — Purpose and Epic Scope

Epic **AGENTIC-CORE** reforms four `/moai` role boundaries plus the default
orchestration pipeline:

- `/moai fix` — **UNCHANGED**: single-pass localized defect fix (deterministic
  Agentless pipeline preserved). Only cross-reference TEXT is touched where loop
  semantics change.
- `/moai review` — **UNCHANGED**: read-only findings/report.
- `/moai goal` — **NEW**: condition-declared universal agentic loop
  (`SPEC-GOAL-ENGINE-001`).
- `/moai loop` — **REDEFINED**: project-wide improvement sweep built ON the goal
  engine as a preset (`SPEC-LOOP-SWEEP-001`).
- **Analyze-First routing reform** — language-independent intent analysis as the
  default orchestration pipeline (`SPEC-ANALYZE-FIRST-ROUTING-001`).

The dependency chain is `ANALYZE-FIRST-ROUTING-001` → `GOAL-ENGINE-001` →
`LOOP-SWEEP-001` (see each SPEC's `depends_on` frontmatter).

## §B — Routing / Trigger Findings

### B.1 Skill/agent triggering is MODEL-SIDE SEMANTIC matching, not literal keywords

`.claude/rules/moai/development/agent-authoring.md`:
- `:49` — `description` field is defined as "When Claude should delegate to this
  agent" (a semantic intent, not a keyword list).
- `:205` — "Include relevant trigger keywords" (documented as *hints*, not a
  required literal match).
- `:215` — "Use normal tool-trigger phrasing … not 'CRITICAL: you MUST' —
  aggressive language overtriggers tools/skills on the latest models."

**Implication**: the "MUST INVOKE when ANY of these keywords appear" +
per-language keyword blocks in the agent bodies are counterproductive on
Opus 4.7+/4.8 — they inflate always-loaded context AND risk over-triggering,
contradicting the project's own authoring rule at `:215`.

### B.2 Agent description bloat (8 of 9 agents carry keyword blocks)

`grep -l "MUST INVOKE" .claude/agents/moai/*.md` returns 8 files:
`builder-harness`, `manager-spec`, `manager-docs`, `manager-git`,
`manager-develop`, `super-advisor`, `sync-auditor`, `plan-auditor`.
`manager-design.md` carries NO trigger description (prose scope only) — the lone
exception, and the diet target's converse (it needs concise trigger prose ADDED).

Observed agent body sizes (`wc -c`, this session):

| Agent | chars |
|-------|------:|
| super-advisor | 6257 |
| manager-git | 9335 |
| manager-docs | 10863 |
| manager-design | 11257 |
| sync-auditor | 11686 |
| builder-harness | 12109 |
| manager-develop | 20055 |
| manager-spec | 23314 |
| plan-auditor | 34984 |

The keyword blocks (EN/KO/JA/ZH) are the compressible surface. Prior-session
estimate: ~16.7K chars of trigger-block text across the 8 agents, ~5-6K tokens
loaded EVERY session. JA/ZH blocks are asymmetric (e.g. builder-harness EN 26 vs
JA/ZH 14 keywords) — an inconsistency independent of the diet.
**NOTE (measurement caveat)**: the ~16.7K figure is a prior-session estimate,
NOT re-measured this session. The run-phase author MUST re-measure the exact
trigger-block char total before/after and report the delta (per
`verification-claim-integrity.md` — attribute the number to a fresh command).

### B.3 `/moai` SKILL.md Intent Router is English-only

`.claude/skills/moai/SKILL.md`:
- `:36` — `## Intent Router` heading.
- `:266` — "Apply the Intent Router (Priority 1 through Priority 4) to determine
  the target workflow. If ambiguous, use AskUserQuestion to clarify."

Structure (from the router body): P1 first-word literal match on English
subcommand tokens; P3 nine intent buckets keyed on English-only cue words; P4
AskUserQuestion fallback. No KO/JA/ZH cues anywhere. A Korean or Japanese request
without an English subcommand token falls straight through P1/P3 to the P4
AskUserQuestion fallback more often than an English request would.

**P3 collision**: the cue word `lint` appears in BOTH the `gate` bucket and the
`fix` bucket (assumption from prior investigation — the run-phase author MUST
re-confirm the exact bucket membership by reading the P3 body before resolving
the collision).

### B.4 CLAUDE.md §2 has only a soft routing hint

`CLAUDE.md` §2 "Request Processing Pipeline" Phase 1 says only "Detect technology
keywords for agent matching." The actual [HARD] Intent Router lives INSIDE the
`/moai` skill (B.3) — so a non-`/moai` natural-language request is routed by the
soft CLAUDE.md hint, not the structured router. This is the gap
`SPEC-ANALYZE-FIRST-ROUTING-001` closes: Analyze-First must be the DEFAULT
main-session behavior, with or without `/moai`.

### B.5 Skill descriptions English-only by policy; `triggers:` under-used

- `.claude/rules/moai/development/coding-standards.md` § Language Policy requires
  all skill/agent/command instruction docs in English. Skill *descriptions* are
  therefore English-only by policy (not a defect).
- `skill-authoring.md` documents a `triggers:` frontmatter block, but only ~2/32
  skills use it (assumption — re-confirm count at run-phase via
  `grep -l "^triggers:" .claude/skills/*/SKILL.md`). Since matching is model-side
  semantic (B.1), `triggers:` is optional metadata, NOT a matcher.
- `moai-meta-harness` skill is deprecated but still fully listed (cleanup target,
  low priority).
- `moai-foundation-core` drift: local frontmatter `related-skills:
  "moai-foundation-context"` (an absorbed/retired skill) vs template value
  `"moai-foundation-cc, moai-foundation-thinking"`. Hygiene fix — adopt the
  template value locally.

### B.6 SKILL.md team/*.md path references — verify dead

The SKILL.md body references `team/*.md` paths. With the Agent Teams static layer
retired (`spec-workflow.md` § Agent Teams Variant — RETIRED), these MAY be dead
references. The run-phase author MUST `grep -n "team/" .claude/skills/moai/SKILL.md`
and confirm each path resolves or is dead before editing.

## §C — loop / fix / goal Findings

### C.1 Two unrelated "loop" surfaces

**(a) `/moai loop` skill** — `.claude/skills/moai/workflows/loop.md`:
- `:54` — `/moai loop` occupies the **goal-based** quadrant; iterate until a
  mechanical completion predicate holds or the ceiling is reached.
- `:79`, `:122-123` — completion predicate "zero errors + tests passing +
  coverage threshold"; coverage-only gap routes to `go test -cover` + `/moai gate`.
- `:82`, `:179` — Step 1.5 **Independent Final Pass** is the ONLY success-exit
  confirmation path (mechanical predicate confirmed twice).
- `:173`, `:186-194` — ceiling exit emits a **5-section verdict**
  (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per
  `verification-claim-integrity.md` §3).
- `:198`, `:205-208` — residue persisted to `.moai/state/loop-verdict-<id>.json`;
  `exit_kind` enum = `ceiling | manual-residue`; `ceiling_source` =
  `flag | ralph | loop_prevention`.
- `:231` — iteration-ceiling precedence: CLI `--max` > ralph.yaml
  `loop.max_iterations` (shipped **10**) > workflow.yaml
  `loop_prevention.max_iterations`. The 50-iteration memory checkpoint (Step 2)
  is an orthogonal memory-pressure safeguard, not a fourth ceiling.

**(b) Go CLI `moai loop`** — `internal/cli/loop.go` + `internal/loop` +
`internal/ralph`: a Spec→Plan→Impl→Sync lifecycle controller, Go-only feedback,
max 5, with ZERO cross-reference to the skill. Config axes:
`workflow.agentic_loop.max_iterations` (10) vs
`loop_prevention.max_iterations` (100), guarded by
`internal/config/agentic_loop_distinctness_test.go` (must stay green).

**Reconciliation is `SPEC-LOOP-SWEEP-001` REQ-LSW-004** — minimal: rename the Go
CLI help text to clarify it is the SPEC-lifecycle controller, document the two
surfaces, defer engine unification.

### C.2 `/moai fix` is a single-pass Agentless pipeline (PRESERVED)

`.claude/skills/moai/workflows/fix.md`:
- `:44` — `## Pipeline Contract (Agentless Classification)`.
- `:48`, `:52` — "Agentless fixed-pipeline"; "No LLM-driven control flow"
  (Agent() delegation exists within a phase, never to select the next phase).
- `:56` — "Repeatability: Even when the parent supplies `--mode loop`, the
  pipeline runs once per invocation. Re-entry requires explicit user re-invocation."
- `:62` — occupies the **turn-based** quadrant.
- `:181` — `@MX:WARN` guarding the static lookup table; LLM-decided dispatch
  fails `TestAgentlessUtilityNoLLMControlFlow`.
- `:267`, `:271` — residue handoff sets `exit_kind: "one-shot-residue"`
  (third value alongside the base `ceiling | manual-residue` enum),
  `iterations_used: 1`, and recommends `/moai loop`.

**Epic decision**: `/moai fix` behavior is UNCHANGED. Only the residue-recommends
text (`fix.md` ~`:265-273`) is updated where loop's queue semantics change
(`SPEC-LOOP-SWEEP-001` REQ-LSW-003).

### C.3 Four-quadrant loop taxonomy (cross-referenced across doctrine)

The quadrant taxonomy appears at `loop.md:54-60`, `fix.md:62-68`, and is
cross-referenced by `goal-directive.md` (§Comparing Autonomous-Continuation
Approaches), `native-invocation-model.md`, `cadence-bridge.md`,
`spec-workflow.md` § Subcommand Classification, `CLAUDE.md:46`. Quadrants:

| Quadrant | Surface | Ends when |
|----------|---------|-----------|
| turn-based | `/moai fix` | one pass per invocation |
| goal-based | `/moai loop` (native `/goal`) | mechanical predicate / model condition |
| time-based | `cadence-bridge.md` recipes | user cancels the schedule |
| proactive | `moai-workflow-ci-loop` | CI-triggered |

`spec-workflow.md` § Subcommand Classification: `/moai fix` = Pipeline
(Agentless); `/moai loop` = Multi-Agent (alias for `/moai run --mode loop`).
`SPEC-LOOP-SWEEP-001` REQ-LSW-004 re-expresses this taxonomy as
"goal engine + presets."

### C.4 Native `/goal` is HUMAN-ONLY — Axis B justifies `/moai goal`

`.claude/rules/moai/workflow/native-invocation-model.md`:
- `:29`, `:44` — `/goal` is HUMAN-ONLY (built-in, no `Skill`-tool bridge; the
  model cannot set it on the user's behalf). `/clear` and `/compact` are the
  other two HUMAN-ONLY commands.
- `:62` — **Axis B**: "Where the nearest native equivalent is HUMAN-ONLY … a
  MoAI subcommand that automates that capability inside the pipeline is NOT
  redundant reinvention — it is the ONLY pipeline path."
- `:71-73` — worked illustration: "MoAI does not currently reimplement any of the
  three; the Axis B principle is stated here for completeness so a future
  subcommand facing a genuinely HUMAN-ONLY native counterpart has a recorded
  justification path."

**`/moai goal` (`SPEC-GOAL-ENGINE-001`) IS that future subcommand.** It is the
MoAI-owned, PROGRAMMATIC reimplementation of native `/goal` semantics, justified
by Axis B. `SPEC-GOAL-ENGINE-001` REQ-GLE-005 updates the `:71-73` illustration
("now `/moai goal` does").

`goal-directive.md` semantics (native `/goal`): a session-scoped completion
condition; after each turn a small fast model (Haiku by default) checks whether
the condition holds against what Claude surfaced in the transcript; the evaluator
does NOT run tools or read files; condition up to 4000 chars; a turn/time bound
(`stop after N turns`) is recommended; requires v2.1.139+, workspace trust, hooks
enabled. `/goal` never bypasses Implementation Kickoff Approval.

### C.5 Stop-hook mechanism (the substrate `/goal` wraps)

- Claude Code Stop hooks receive stdin JSON including `session_id`; an exit-0
  stdout JSON `{"decision":"block","reason":"…"}` continues the turn (per Claude
  Code hook semantics: stdout JSON honored only on exit 0). Prompt-type Stop
  hooks (`type: prompt`) are the mechanism native `/goal` wraps.
- Infra: `.claude/hooks/moai/handle-*.sh` wrappers → `moai hook <event>`;
  registration in `settings.json.tmpl`; MoAI 5s hook timeout policy
  (goal evaluation may need longer — a per-hook timeout override is a
  `SPEC-GOAL-ENGINE-001` design decision, plan.md § NEEDS CLARIFICATION).
- Existing hook-verb infra confirms the pattern: a new verb `moai hook
  stop-goal` (or folding into an existing stop handler) is feasible.

## §D — Overlap SPEC Reconciliation

Three existing SPECs touch adjacent surfaces. Frontmatter observed this session:

### D.1 `SPEC-AUTONOMY-RUN-GOAL-001` (status: `completed`, tier M)

Delivers run-phase autonomy: Mode 6 (workflow) catalog addition +
`/goal ac_converge` wrapping in `.claude/skills/moai/workflows/run.md`
§ Run-phase Autonomy, with GATE-2 (Implementation Kickoff Approval) preservation.
It wraps the **native** `/goal` at the run-phase boundary.

**Relationship (NO conflict)**: `SPEC-GOAL-ENGINE-001` builds a **programmatic
`/moai goal`** — a distinct, MoAI-owned surface. It does NOT modify the
`run.md § Run-phase Autonomy (/goal ac_converge)` section that AUTONOMY-RUN-GOAL
owns. `SPEC-GOAL-ENGINE-001` REQ-GLE-006 (Analyze-First termination judge) must
preserve the `run.md ac_converge` wiring point verbatim and cross-reference it,
NOT overwrite it. The `ac_converge` `/goal` remains a native-`/goal` wrapping;
`/moai goal` is the programmatic sibling.

### D.2 `SPEC-HARNESS-EVOLVE-001` (status: `draft`, tier M, era V3R6)

"Routing Observation Ledger — Loop 0 (Generator) of the self-evolving harness."
Modules: `internal/harness/routing, internal/cli, .claude/skills/moai`. It
observes routing decisions via a Stop-hook JSONL ledger — a DIFFERENT concern
(observation/telemetry) from AGENTIC-CORE's routing REFORM (behavior change).

**Potential collision**: both `SPEC-HARNESS-EVOLVE-001` and
`SPEC-GOAL-ENGINE-001` add Stop-hook surfaces. `SPEC-GOAL-ENGINE-001` plan.md
§ Dependencies flags this: the two Stop-hook verbs (`stop-goal` vs the routing
ledger observer) must not clobber each other's `settings.json.tmpl` registration.
Both should register as separate Stop-hook entries (Stop hooks compose — multiple
entries run). `SPEC-ANALYZE-FIRST-ROUTING-001` REQ-AFR-001 must keep the CLAUDE.md
§2 pipeline enumeration that HARNESS-EVOLVE's Phase −1/Ω attachment points wrap
(do NOT delete the enumeration those phases anchor to).

### D.3 `SPEC-CADENCE-BRIDGE-001` (status: `completed`, tier S, era V3R6)

"AUTOMATE Bridge — sanctioned cadence recipes composing native `/loop` and cron
with read-only `/moai` entry points." Owns the **time-based** quadrant (C.3).

**Relationship (NO conflict)**: `SPEC-LOOP-SWEEP-001` redefines the **goal-based**
quadrant (`/moai loop`). REQ-LSW-004 must state that `/moai loop` remains NOT
cadence-eligible (cadence recipes never enter run-phase, never commit — a HARD
invariant CADENCE-BRIDGE owns). LOOP-SWEEP updates the taxonomy prose but does
NOT change cadence eligibility.

## §E — Open Decisions (summary; full markers in each plan.md)

The following require user resolution before run-phase (surfaced as
`[NEEDS CLARIFICATION]` markers in the respective plan.md, per
`.claude/skills/moai-workflow-spec/SKILL.md` § NEEDS CLARIFICATION convention —
markers live ONLY in plan.md/research.md):

1. **[SPEC-GOAL-ENGINE-001] Tier-2 model-condition evaluation mechanism** —
   prompt-type Stop hook config vs orchestrator self-eval via block-reason claim.
2. **[SPEC-GOAL-ENGINE-001] Stop-hook verb placement** — new `handle-stop-goal.sh`
   wrapper + `moai hook stop-goal` verb vs folding into an existing stop handler.
3. **[SPEC-GOAL-ENGINE-001] goal-eval hook timeout** — per-hook override above the
   MoAI 5s policy default.
4. **[SPEC-GOAL-ENGINE-001] native-/goal-active detection** — how `stop-goal`
   detects and yields to an active native `/goal`.
5. **[SPEC-LOOP-SWEEP-001] `run --mode loop` alias disposition** — keep the alias
   or retire it now that loop is a goal preset.

## §F — Constraints Inherited by All Three SPECs

- GEARS requirements with stable IDs; ACs verify **reachability** not token
  presence (no vacuous compound greps `A|B|C ≥ N`; each AC = a single
  discriminating check, baseline 0 → post ≥ 1 where applicable; router-registration
  ACs pinned SEPARATELY from implementation-file ACs; template-mirror ACs use
  per-file existence + content checks).
- Template-First (`CLAUDE.local.md` §2 + §25 neutrality): every changed `.claude/`
  file mirrored in `internal/template/templates/.claude/` with NO internal SPEC
  IDs in the template body; `make build` passes.
- docs-site 4-locale sync is a DEFERRED follow-up (each plan.md § Deferred), NOT
  run-phase scope.
- Coverage 85%+ for critical Go packages (`SPEC-GOAL-ENGINE-001` only).
