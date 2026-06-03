# Design — SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001

> Tier M design note. Records the single load-bearing design decision (the wiring-mechanism choice for
> the orphaned installers) and the structure of the `main.md` router. This is a WHAT/HOW-boundary doc for
> plan-phase; concrete function signatures are deferred to run-phase.

## A. The Core Design Decision — Where the Installers Are Called

The diagnosis established that `InjectMarker` (`internal/harness/layer3.go`) and `ScaffoldHarnessDir`
(`internal/harness/layer5.go`) work but are never invoked. The central design question is **where to
anchor their call path** so the wiring is testable and cannot silently rot again.

### A.1 Options

| Option | Mechanism | Testability | Dead-code recurrence risk | Verdict |
|--------|-----------|-------------|---------------------------|---------|
| **A — thin CLI install command** | Add `moai harness install --spec-id --domain` (or extend an existing harness subcommand) that calls `InjectMarker` + `ScaffoldHarnessDir`. Phase 7 of `project/meta-harness.md` instructs the orchestrator to run it. | High — a Go entry point with table-driven tests + the smoke gate both exercise the call path | Low — a CLI command with tests cannot be silently removed without a red test | **Recommended** |
| **B — skill-body-only orchestration** | Phase 7 skill body instructs the orchestrator to drive `InjectMarker` via ad-hoc Bash/CLI, no new Go call site | Low — the call path lives only in prose; the original dead-code failure mode is exactly "a flow that was supposed to call it but didn't" | High — recurrence is the documented root cause | Rejected as primary |

### A.2 Decision (recommended, to be confirmed in M2)

**Option A.** A thin `moai harness install` CLI surface that wraps the two existing functions:

- It is the smallest change that gives the wiring a **test anchor** — the diagnosis's root cause is a
  "completed" installer that was never called and never verified. A CLI command + its test + the Phase-6
  smoke gate form a triple guard against recurrence.
- It honors the subagent boundary (REQ-HAW-003): the command takes positional/flag inputs
  (`--spec-id`, `--domain`, import paths), never `AskUserQuestion`. Per `internal/cli/CLAUDE.md`, CLI code
  is subagent-context and the orchestrator owns user interaction.
- Phase 7 of `project/meta-harness.md` then becomes a thin orchestrator instruction: "run
  `moai harness install` with the generated SPEC ID + domain", keeping the skill body declarative.
- It does NOT rewrite the installers (EX-6 / D-4) — it is a caller.

Option B remains the fallback if M2 surfaces a reason the CLI surface is undesirable (e.g. the harness
generation is entirely orchestrator-side with no natural CLI moment); the run-phase will confirm.

## B. main.md Router Structure (REQ-HAW-006)

The current `mainMD()` builder (`internal/harness/layer5.go`) already emits `**Domain**:` metadata, a
`## Domain Summary` section, and a `## Linked Files` list. The router change is **additive, not a rewrite**:
it inserts ONE new `## Task-Shape Routing` table between the existing Domain Summary and Linked Files
sections so the orchestrator can route at activation time. Domain Summary + Linked Files are kept as-is;
the scaffolding algorithm is untouched (D-4):

```
# Harness Main
<!-- 진입점: CLAUDE.md @import가 이 파일을 따라옵니다. -->

**Domain**: <domain>
**SPEC**: <spec-id>
**Updated**: <date>

## Domain Summary
<one-paragraph domain summary>

## Task-Shape Routing
| Task shape | Route to specialist |
|------------|---------------------|
| <observable task-shape 1> | .claude/agents/harness/<role-1>.md |
| <observable task-shape 2> | .claude/agents/harness/<role-2>.md |

## Linked Files
- plan-extension.md — Plan phase chain
- run-extension.md — Run phase chain
- sync-extension.md — Sync phase chain
- chaining-rules.yaml — machine-readable rules
- interview-results.md — original interview answers
```

The `## Task-Shape Routing` table is the ONLY new element. The `**Domain**:` metadata, the `## Domain
Summary` section, and the `## Linked Files` list ALL already exist in the current `mainMD()` builder and
are preserved verbatim. The change is confined to inserting the routing table into `mainMD()` body
construction — the `ScaffoldHarnessDir` file-writing algorithm is untouched (D-4).

## C. Smoke Gate Extension (REQ-HAW-010..014, plus REQ-HAW-013b)

`runHarnessCheck` already performs L1-L5. The smoke gate adds three agent-frontmatter checks atop the
existing layers:

- **Agent description check** (REQ-HAW-012): iterate `.claude/agents/harness/*.md`, parse the
  `description:` frontmatter field, FAIL if empty.
- **Missing-`skills:` check** (REQ-HAW-013b / AC-HAW-015): FAIL if a generated agent OMITS the `skills:`
  frontmatter key entirely. This is the runtime enforcement of REQ-HAW-008's emission contract — without
  it, a `skills:`-less agent would pass silently and reproduce the auto-discovery failure mode the SPEC
  exists to close.
- **Dangling skill reference check** (REQ-HAW-013): for each generated agent's `skills:` frontmatter
  entry that IS present, resolve the referenced skill directory; FAIL if a `my-harness-*` reference points
  to a non-existent dir. References to template-distributed `moai-*` skills are NOT dangling (EC-4).
  (Distinct from the missing-`skills:` check above: absent key vs present-but-dangling key.)

These are added as either a new `checkLayer6...` helper or folded into the existing aggregation, preserving
the L1-L5 status string semantics (REQ-HAW-014). L3 (marker pairing) and L5 (`main.md`) already satisfy
REQ-HAW-010/011 — the design reuses them rather than duplicating.

## D. Why This Is Low-Risk

- All three mechanisms (`InjectMarker`, `ScaffoldHarnessDir`, `runHarnessCheck`) exist and are unit-tested.
- The change is additive: a caller + a router body restructure + two smoke checks.
- No prefix migration (EX-1), no update-protection change (EX-2), no external-project edits (EX-3).
- The biggest blast radius is `doctor_harness.go` (shared diagnosis), mitigated by preserving L1-L5 and
  TDD-extending only.

## E. Open Questions (resolve in M2/run-phase)

- OQ-1: Confirm Option A vs B (M2).
- OQ-2: If Option A, decide whether `moai harness install` is a new subcommand or an extension of an
  existing `harness` subcommand (`internal/cli/harness/`).
- OQ-3: Whether the agent-description + dangling-skill checks warrant a distinct `L6:` status token or fold
  into the existing aggregation (cosmetic; both satisfy REQ-HAW-014).

## F. Cross-References

- plan.md §F M2/M3/M5
- spec.md §C REQ-HAW-001..014
- `internal/harness/layer3.go` (`InjectMarker`), `internal/harness/layer5.go` (`mainMD`, `ScaffoldHarnessDir`)
- `internal/cli/doctor_harness.go` (`runHarnessCheck`), `internal/cli/CLAUDE.md` (subagent boundary)
- `.moai/docs/harness-delivery-strategy.md` §6.5 (Option A rationale; namespace dependency clarification)
