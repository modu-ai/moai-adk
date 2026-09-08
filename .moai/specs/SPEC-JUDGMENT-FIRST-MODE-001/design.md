# SPEC-JUDGMENT-FIRST-MODE-001 — Design

Design is stated at the level of observable behavior and record shape. Concrete function names and
signatures are deferred to the run phase.

## 1. Why a mode axis rather than a new rule

The codebase already carries exactly one clause able to suppress the `(권장)` /
`(Recommended)` label: **Adaptive strength** (`askuser-protocol.md` § Recommendation Placement
Principles, principle 5). It gates suppression on *estimated user proficiency*.

That is the right mechanism sited on the wrong input. Proficiency is a system inference — the very
kind of model-derived default the issue's second condition ("an LLM 'best practice' is not a
policy") warns against. Generalizing the same clause so its input can also be an **explicit
operator setting** costs one axis and reuses the mechanism, its placement, and its surrounding
guards.

Authoring a second suppression clause beside it would give the tree two rules that can both
suppress the label, differing in trigger, with no stated precedence. That is the shape this design
refuses.

Consequently the mode is **not** auto-selected. Inferring `pull` from proficiency would collapse
the two inputs back together and reintroduce a system-chosen default in the exact place this SPEC
removes one.

## 2. Mode resolution

```
.moai/config/sections/interview.yaml
        │  recommendation_mode: push | pull   (absent ⇒ push)
        ▼
InterviewConfig.RecommendationMode  (internal/config)
        │
        ├──► orchestrator: read at output-composition time, per surface
        └──► PreToolUse observer: recorded verbatim on every observation row
```

Resolution rules:

| Value in YAML | Resolved | Recorded |
|---|---|---|
| absent / empty | `push` | `push` |
| `push` | `push` | `push` |
| `pull` | `pull` | `pull` |
| anything else | `push` | the raw value, so the misconfiguration is visible |

**Mid-session flip.** Because the orchestrator reads the mode at output-composition time rather
than latching it at session start, a mode change takes effect at the **next** composition and
never rewrites, recalls, or re-renders output already emitted. A question already on screen keeps
the labels it was composed with; the next one carries the new mode. There is no in-flight state to
migrate and none is authored (spec.md §E.3).

The unrecognized value resolves rather than errors, because a config typo must not make the tool
unusable — but it is recorded, because a silently-corrected typo is a mode the operator thinks is
on and is not. That combination is what makes AC-JFM-002 worth having.

## 3. What `pull` withholds, and what it never withholds

```
observation ──► explanation ──► enumerated options ──► [ recommendation ]
   always          always              always            push: emitted
                                                         pull: withheld until asked
```

`pull` moves exactly the last box. Everything to its left is unchanged in both modes, which is
what "Detect → Explain → Ask, but never decide" means operationally.

Specifically **not** touched by the mode:

- The Report-Before-Ask gate — the report still precedes the question in the same turn.
- Requested-Deliverable Primacy — a requested report still ends the turn.
- The neutral-description rule — option descriptions were already required to be factual and
  non-persuasive; under `pull` they carry the entire information load, which makes that existing
  rule *more* load-bearing, not less.
- The Implementation Kickoff Approval gate — still mandatory and score-independent. Under `pull`
  the gate asks the same question with an unlabeled first option.

The last point is the design's sharpest edge and is stated deliberately: `pull` does not remove a
gate, it removes the anchor **inside** the gate.

## 4. Per-surface shape

| Surface | `push` (unchanged) | `pull` |
|---|---|---|
| S1 option label | first option carries `(권장)` / `(Recommended)` | no option carries it; option order carries no preference claim |
| S2 Discovery banner | `⏭️ Recommended action:` line rendered | line omitted; `🔍 Scope` / `📊 Findings` / `⚠️ Drift` unchanged |
| S3 Epic Stats / Status | `⏭️ Next:` line rendered | line omitted; the stats themselves unchanged |
| S4 Insight banner | `What` / `Why` / `Alternatives` / `Implications` | same four fields, plus the literal field `Your call: [what the reader decides]` placed after `Implications:`, inside the same banner frame |
| S5 Error Recovery | `A. Retry as-is` fixed first | ordered by increasing cost to the user — `Pause` → `Retry as-is` → `Alt approach` → `Abort+preserve` |
| S6 `/clear` guidance | "Recommend `/clear`" | state the measured threshold crossing and the available actions; no directive phrasing |

S2 and S3 are **omission**, not replacement. Substituting a differently-worded directive would
keep the anchor and change only its label.

S6 is the one surface that must state the mode locally: it sits outside the AskUserQuestion
channel, so it inherits none of that channel's guards and cannot be covered by a cross-reference
alone.

## 5. The doctrine-edit shape

Each of the six surfaces receives the same three-part edit, so the convention reads identically
wherever it appears:

1. the existing text, unchanged, presented as the **push branch**;
2. the **pull branch**, stating what is withheld;
3. the **on-request clause**, stating that an explicit user request for analysis or for a
   recommendation restores the push-branch output on that surface.

`CONST-V3R5-035` is handled by exactly this shape: part 1 is the current clause verbatim, so the
registry's `clause:` string still matches and the canary gate stays satisfiable.

## 6. The runtime observer

### 6.1 Placement

`internal/hook/pre_tool.go` already carries two observation branches with the required properties —
`Agent|Task` (agent-model) and `SendMessage` (stop-guard). Both sit **after** every established
deny path, both parse `input.ToolInput`, both can only add a `SystemMessage`. The
`AskUserQuestion` branch is the third of the same kind, placed after `SendMessage`, returning no
deny in either mode.

`.claude/settings.json` gains a fourth `PreToolUse` matcher block, `AskUserQuestion`, routing to
`handle-pre-tool.sh` exactly as the existing `SendMessage|TaskStop` block does.

### 6.2 Record shape

One JSONL row per `AskUserQuestion` issuance, appended under `.moai/logs/`:

| Field | Source | Notes |
|---|---|---|
| timestamp | UTC, RFC3339 | |
| session id | `input.SessionID` | absent ⇒ row still written, field empty |
| mode | resolved config | `push` / `pull` / raw unrecognized value |
| label_present | payload scan | true when any option label carries `(권장)` or `(Recommended)` |
| option_count | payload | the honest structural fact |
| question_count | payload | the honest structural fact |
| question_type | **see 6.3** | |

### 6.3 The honesty constraint on `question_type`

"Decision-type question whose options derive from investigation results" is a **doctrine concept**.
The `AskUserQuestion` payload carries question text and options; it does not carry a type tag, and
whether a report preceded the call in the same turn is not visible to a PreToolUse hook.

The design therefore fixes this rule for the run phase: the observer records **what the payload
actually contains**, and any classification it emits is labeled as derived, with its derivation
stated. An inferred classification presented as an observation would make the falsifier itself
vacuous — the precise failure mode the criterion exists to prevent.

**The denominator is settled here, once, and the other artifacts follow it.** Iteration-1 audit
finding D3 recorded three incompatible denominators across `acceptance.md`, this file, and
`plan.md`. There is now exactly one rule:

> **AC-JFM-018's denominator is every recorded row with `mode == "pull"`. No `question_type`
> filter is applied, in any branch.**

No conditional, no fallback, no "filtered when derivable". The `question_type` field is retained
in the row shape (§6.2) as a descriptive record only — never as a selector — precisely so a later
reader cannot reintroduce the filter. The widening is not a concession to a missing field; it
follows the requirement, which was itself widened for the same reason: REQ-JFM-005 now binds
**every** `AskUserQuestion` call under `pull`, not only decision-type ones (spec.md §B). Rule and
measurement therefore have the same scope, and the criterion cannot be satisfied by a sample the
filter emptied.

**Detector integrity (REQ-JFM-024).** `violations == 0` over any denominator is satisfiable by an
observer whose detector never fires. The row-shape unit test asserts both directions of
`label_present` on constructed payloads, and the pre-landing `push`-mode window (M0) must contain
at least one live row with `label_present: true` before the `pull` window is opened. Only after
both are on record does a zero-violation `pull` window mean anything.

### 6.4 Fail-open

Every uncertainty allows the call: unparseable payload, absent or unreadable config, unwritable
log path, missing session id. No error propagates to the caller and no established decision is
displaced. This is not defensive habit — an observer that can block the question channel can
deadlock the only path to the user.

## 7. The two-axis regression argument

| Axis | Answers | Fails when |
|---|---|---|
| Static (CI grep guard) | "does the convention exist, and are doctrine and banner templates consistent?" | text drifts apart |
| Runtime (observer JSONL) | "is the convention actually followed?" | the label appears under `pull` |

Only the runtime axis can fail while the text is perfect. That asymmetry is the whole reason both
are required: a static guard alone certifies a convention nobody follows.

The static guard is itself demonstrated in both directions (mutation RED, clean GREEN) for the same
reason — a guard observed only passing has not been shown to be able to fail.

## 8. Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Rewrite `CONST-V3R5-035` outright | Frozen zone with `canary_gate: true`; conditioning preserves the clause string and the gate |
| A separate `anchoring.yaml` config section | The setting configures interview output; `interview.yaml` already exists and already carries the interview's behavioral knobs |
| A `workflow.*` guard-style key | The guard precedents gate a **deny**. This gates an output convention and denies nothing; siting it under `workflow` would imply a blocking layer that does not exist |
| Deny `AskUserQuestion` calls carrying a label under `pull` | The question channel is the only path to the user; a deny there can deadlock the session. CONST-7 |
| Change the auditor report contract | On the two most common plan-audit paths the report never reaches a user (spec.md §A.3) |
| Auto-select the mode from proficiency | Reintroduces a system-chosen default where this SPEC removes one (§1) |
| Replace the withheld banner lines with a neutral directive | Keeps the anchor, changes its wording |
