---
name: hns-lsel-curator
description: >
  Local Self-Evolution Loop (LSEL) curator — the CLUSTER + drain engine for the
  GOOS-local PROPOSE→APPLY seam closure (SPEC-LSEL-LOCAL-EVOLUTION-001). Companion-offset
  drain of .moai/lessons-inbox.jsonl with a drain-side severity filter that drops the ~65%
  Bash-timeout/sandbox noise, event_key clustering with a frequency gate, and a
  Generative-Agents-style 1-10 importance score. Candidates stage at
  .moai/state/lsel/clusters.json. M1 = drain only (NO PROPOSE, NO APPLY, NO memory/ writes).
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.2.0"
  category: "harness"
  status: "active"
  updated: "2026-08-26"
  tags: "lsel,self-evolution,drain,cluster,harness,dogfood"
---

# hns-lsel-curator — LSEL CLUSTER + drain engine

> **Namespace:** `hns-lsel-*` is user-owned dogfood (CLAUDE.local.md §24). This skill is
> NOT mirrored into `internal/template/templates/` — it lives only in this repo. Graduation
> to `moai-lsel-*` + 16-language distribution is a separate SPEC (out of scope per spec.md §G).
>
> **M1 scope:** drain + cluster + stage candidates. NO APPROVE, NO APPLY (M3).
> **M2 scope:** drain + cluster + **PROPOSE shadow** (no APPROVE, no APPLY). The PROPOSE stage
> emits shadow proposals + self-critiques; APPROVE/APPLY land in M3 via the fresh
> `hns-lsel-applier` path. M2 does NOT write to `memory/` — the first `feedback_*.md` topic
> file is an M3+ deliverable after APPROVE.

## What this skill does

The MoAI-ADK repo accumulates tool-failure stubs in `.moai/lessons-inbox.jsonl` (624 stubs at
M1 start, re-measured — a moving target). The constitution names the orchestrator as the drain
actor, but until this skill there was **zero mechanical drain code** — the drain existed only as
a doctrine paragraph (`moai-constitution.md:147`). This skill closes that gap in user-owned
surfaces, without touching the frozen Go applier (`internal/harness/applier.go:22` —
its write-flag stays `false`; REQ-LSEL-003: bypass, never unfreeze).

The drain is split into a **mechanical core** (`drain.sh`, deterministic, testable) and a
**model-mediated layer** (this SKILL.md + your judgment, invoked for M2+ importance refinement
and proposal drafting).

## The mechanical core — `drain.sh`

`drain.sh` is a portable bash + jq script that lives next to this SKILL.md. It performs the
deterministic half of the drain:

```
drain.sh --inbox <path-to-lessons-inbox.jsonl> --state-dir <path-to-lsel-state>
```

Pipeline (REQ-LSEL-009 + AC-LSEL-009 / AC-LSEL-010):

1. **Companion offset** — read `<state-dir>/drain-offset.json` (seed `{"offset":0}` if absent).
   The inbox is append-only and is NEVER mutated; the offset marks consumed stubs
   (SPEC-HARNESS-RATCHET-REWIRE-001 D3 companion-offset pattern).
2. **Slice** — read stubs from the offset onwards (`tail -n +<offset+1>`).
3. **Drain-side severity filter** (AC-LSEL-010) — discard noise BEFORE clustering:
   - `tool_failure:Bash:UnknownFailure` — the opaque ~65% timeout/sandbox bucket (the dominant
     noise share; report §2).
   - `tool_failure:Bash:SandboxViolation` — environment constraint, not a code defect.
   - any `*:TimeoutError` (Bash + MCP timeouts).
   The filter is drain-side because `internal/hook/failure_observer.go` (the inbox writer) is
   OUTSIDE the six loop-writable surfaces (plan.md §F.1 [DECISION RESOLVED]), so the loop cannot
   edit the writer — it filters on read instead.
4. **Cluster** by `event_key` with frequency count, first/last seen, and up to 3 sample summaries.
5. **Singleton gate** — discard clusters with `frequency < 2` (single-occurrence noise per the
   constitution Lessons Protocol drain paragraph).
6. **Importance** — score each survivor with a Generative-Agents-style 1-10 gate:
   `importance = min(10, frequency)` (frequency as proxy; the model augments this in M2+ with a
   severity hint and retrieval-weighted judgment).
7. **Emit** candidates to `<state-dir>/clusters.json`; advance the companion offset.

### `clusters.json` schema

```json
{
  "drained_at": "2026-08-04T08:41:00Z",
  "offset_before": 0,
  "offset_after": 624,
  "total_read": 624,
  "noise_discarded": 533,
  "singletons_discarded": 4,
  "candidates": [
    {
      "event_key": "tool_failure:Agent:UnknownFailure",
      "frequency": 41,
      "first_seen": "...",
      "last_seen": "...",
      "sample_summaries": ["...", "...", "..."],
      "source": "tool:Agent",
      "importance": 10
    }
  ]
}
```

### Empty-delta no-op

If the inbox has not grown past the offset, `drain.sh` writes an empty-candidate `clusters.json`
and leaves the offset unchanged. Not a failure (acceptance.md §E edge case).

## The model-mediated layer (you, when invoked)

`drain.sh` produces the deterministic candidate set. When this skill is invoked for a real
curation pass (M2+), your job on top of the mechanical output is:

- **Read `clusters.json`** and rank candidates by `importance` then `frequency`.
- **Augment importance** with a severity hint the mechanical core cannot see: a recurring
  `Bash:ExitError` cluster points at a real command-shape defect (high signal); a recurring
  `Agent:ContextCancelled` cluster may be session-teardown noise (lower signal). Record the
  rationale in the candidate's prose when you draft the M2 proposal — do NOT rewrite
  `clusters.json` (it is the mechanical artifact; your augmentation lives in the proposal).
- **Do NOT write to `memory/` in M1.** Candidates stage in `clusters.json` only. The first
  `feedback_*.md` topic file is produced by the M2 PROPOSE stage after retrieval-before-propose
  and self-critique (REQ-LSEL-010).

## What this skill does NOT do (M1 boundaries)

- **No APPROVE / APPLY** — the parallel user-owned applier (`hns-lsel-applier`) is M3.
- **No edits to frozen doctrine** — `.claude/rules/moai/**`, `CLAUDE.md`,
  `internal/template/templates/**`, retained agents, `moai-*` skills, and the frozen Go
  applier / `curator_dispatch.go` are all byte-for-byte untouched (REQ-LSEL-001 / §B.3).
- **No new `.moai/config/sections/` file** — loop state lives under `.moai/state/lsel/`
  (a new section file would be wiped on `moai update`; plan.md §B.4 / AP-LSEL-005).
- **No orchestrator-only synchronous user-question channel** — this is a subagent-owned
  mechanism skill; it never invokes the orchestrator's user gate. On a missing input,
  return a structured blocker report; the orchestrator runs the user gate (CLAUDE.md §8).

## Durable operations — the session-start trigger (`session_drain.sh`)

The drain's original "schedule" was a session-scoped `/loop` recipe that died with its
owning session (2026-08-04) and executed nothing even while alive — the inbox stalled
for 3 weeks with nobody notified (SPEC-LSEL-DRAIN-STALL-001 §B). The trigger is now
mechanical:

**ALL drains route through `session_drain.sh`** (next to this SKILL.md), never
`drain.sh` directly. The wrapper adds what the frozen core deliberately lacks: an
exclusive drain lock (contention = safe no-op), an **unconditional** archive of any
existing `clusters.json` to `clusters-history/` BEFORE any overwrite (a direct
`drain.sh` call bypasses archiving and can silently discard staged candidates —
`drain.sh` overwrites `clusters.json` on both the drain path and the no-op path), a
one-line status, and fail-open (any internal error degrades to a stderr notice and
exit 0 — the hook never blocks session start).

```bash
session_drain.sh [--inbox <path>] [--state-dir <dir>]   # defaults: live paths
```

**PROPOSE reads the archived copies** in `.moai/state/lsel/clusters-history/` (newest
first), NOT the live `clusters.json` — under per-session-start drains the live file is
ephemeral: the next no-op session start overwrites it with `candidates: []`. The
wrapper's archive is what survives.

**Local wiring is a maintainer-machine deliverable** (applied in M2, NOT carried by
the PR — a tracked `.claude/settings.json` entry would be wiped by every
`moai update`, so tracked wiring is affirmatively wrong): add BOTH `session_drain.sh`
and `backlog_check.sh` to `.claude/settings.local.json` `.hooks.SessionStart` with an
explicit `"timeout": 30` each (the measured full-backlog drain is <1s for a 1.1MB
inbox; 30s matches the live SessionStart hook precedent). Wiring verification:
`jq '.hooks.SessionStart' .claude/settings.local.json`.

**Removed dead anchors:** the former `CLAUDE.local.md` section-28 anchor references in
`backlog_check.sh` (header comment + reminder body — both occurrences) were removed in
SPEC-LSEL-DRAIN-STALL-001 M1; the operating instructions now live in this section,
mirrored into the restored CLAUDE.local.md LSEL section (M2 local deliverable).

## Verification (run before declaring a drain complete)

```bash
# 1. The drain mechanics + the wrapper (fixture-based characterization tests —
#    AC-LSEL-009/010 and AC-LDS-001..006 + the mutant probe):
.claude/skills/hns-lsel-curator/drain_test.sh
.claude/skills/hns-lsel-curator/session_drain_test.sh

# 2. A real drain of the live backlog — VIA THE WRAPPER (all drains are
#    wrapper-mediated; re-measure first, the inbox is a moving target). Capture
#    BEFORE the drain, then judge the ARCHIVED copy: a later no-op session-start
#    drain overwrites the live clusters.json with candidates: [], but the wrapper
#    archives unconditionally before any overwrite, so the bulk-drain result
#    survives in clusters-history/.
OFFSET_BEFORE=$(jq -r .offset .moai/state/lsel/drain-offset.json)
LIVE_COUNT=$(wc -l < .moai/lessons-inbox.jsonl | tr -d ' ')
.claude/skills/hns-lsel-curator/session_drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel
ARCHIVE=$(ls -t .moai/state/lsel/clusters-history/clusters-*.json | head -1)
jq --argjson n "$LIVE_COUNT" --argjson b "$OFFSET_BEFORE" \
   '(.offset_after == $n) and ((.candidates // []) | length >= 1) and (.total_read == ($n - $b))' "$ARCHIVE"
# must print: true  (offset==live AND candidates>=1 AND self-consistent — the
# AC-LDS-010 predicate; the session_drain_test.sh mutant probe proves it rejects
# an offset-only-advance fake)

# 3. M1 invariant — zero memory/ writes from the drain:
find memory -newer <drain-start-timestamp> -name 'feedback_*' 2>/dev/null | wc -l   # must be 0
```

## Characterization test

`drain_test.sh` (next to this SKILL.md) is the TDD RED→GREEN harness. It builds a synthetic
inbox with known noise + signal stubs, runs `drain.sh`, and asserts the drain semantics:
noise excluded pre-cluster, signal clustered with correct frequencies, singletons discarded,
offset advanced, candidates emitted, zero `memory/` writes, idempotent re-drain. Run it after
any edit to `drain.sh`. `session_drain_test.sh` (same directory) is the wrapper harness —
five paths (drain / lock contention / archive-before-overwrite / no-op / fail-open) plus the
mutant probe. Run it after any edit to `session_drain.sh`.

## Cross-references

- **SPEC:** `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/{spec,plan,acceptance,progress}.md`
- **Design report (SSOT):** `.moai/reports/moai-local-self-evolution-design-20260804.html`
  §6 stage 2 (CLUSTER), §10 P1, §11 mustFix B#1/B#3.
- **Frozen applier (reference only):** `internal/harness/applier.go:22`
  (the write-flag, kept `false`), `internal/harness/curator_dispatch.go`.
- **Constitution drain paragraph (the "0 Go code" stub this skill replaces):**
  `.claude/rules/moai/core/moai-constitution.md:147`.
- **Namespace guard:** `internal/template/split_namespace_test.go`,
  `internal/template/internal_content_leak_test.go` (extended in M2 — AC-LSEL-006).

---

## PROPOSE stage (M2 — shadow proposals)

The PROPOSE stage consumes the candidate clusters **archived** in
`.moai/state/lsel/clusters-history/` (newest archive — the live `clusters.json` is
ephemeral under per-session-start drains: the next no-op drain overwrites it with
`candidates: []`; SPEC-LSEL-DRAIN-STALL-001 REQ-LDS-010) and
emits **shadow proposals** — one per candidate worth acting on — at
`.moai/state/lsel/proposals/<proposal-id>/`. M2 proposals are SHADOW only: no APPROVE, no APPLY.
APPROVE/APPLY land in M3 via the fresh `hns-lsel-applier` path (NOT via the dead
`moai-harness-learner` Tier-4 flow — see "Tier-4 finding" below).

### Retrieval-before-propose (Reflexion)

BEFORE drafting a proposal, retrieve relevant `feedback_*.md` topic files from
`~/.claude/projects/<hash>/memory/`. The retrieval grounds the proposal in prior
lessons (Reflexion-style) and is evidenced in the proposal's `retrieval_evidence`
block. A proposal without retrieval evidence is malformed and MUST NOT be emitted.

### Proposal payload schema (AC-LSEL-011)

Each proposal lives at `.moai/state/lsel/proposals/<id>/` and contains exactly:

| File | Purpose |
|------|---------|
| `proposal.md` | YAML-frontmatter payload + prose body |
| `diff.patch` | The proposed edit (unified diff; NOT applied in M2) |
| `self-critique.md` | Model-performed critique against frozen doctrine |

`proposal.md` YAML frontmatter carries the full schema (8 required keys):

```yaml
---
proposal_id: lsel-001
target_surface: <one of the 6 evolvable surfaces, spec.md §B.3>
rationale: |
  <what + why>
WHY-not-just-WHAT: |
  <the reasoning, not just the change — catches "what" proposals that skip the "why">
prediction: <a FALSIFIABLE expected effect — the verify_command must be able to falsify it>
verify_command: <a runnable command that, if green, confirms the prediction>
blast_radius: <which surfaces the diff touches; used by the CSA forced-gate match>
memory_type: semantic|procedural|episodic   # CoALA taxonomy
retrieval_evidence:
  - <path to a feedback_*.md retrieved before drafting>
status: blocked   # blocked | ready — blocked if self-critique has an UNRESOLVED objection
---
```

### Self-critique gate

`self-critique.md` is model-performed (NOT a mechanical doctrine checker — report §13 caveat 3:
the model can rationalize; the frozen allowlist + `/moai gate` are the real safety floor). It
lists objections against frozen doctrine; each objection is marked RESOLVED or UNRESOLVED. A
proposal with ANY UNRESOLVED objection is `status: blocked` and MUST NOT proceed to APPROVE.
A proposal that never converges stays blocked; the curator returns a blocker report and the
orchestrator surfaces it (acceptance.md §E edge case — not a ship-blocker for M2; it proves
the gate fires).

### Tier-4 finding (AC-LSEL-012 — do NOT wire the dead flow)

**Finding (verified 2026-08-04 via `tier4_firing_test.sh`): the `moai-harness-learner` Tier-4
synchronous-user-question flow is DEAD at the production invocation layer.** The CLI (`moai harness apply`)
prints a stub string and never invokes the learner skill; `CuratorDispatch` has 0 production
callers (the audit's cautionary precedent); the frozen applier's write-flag (`false`
at `internal/harness/applier.go:22`) is the apply dead-switch; and NO mechanical trigger causes the orchestrator to surface a Tier-4
proposal (the audit's exact failure mode, report §11 mustFix B#1).

Per acceptance.md §E edge case, M2 does NOT wire the PROPOSE→APPROVE handoff to depend on the
Tier-4 flow. APPROVE routes via the M3 fresh path (`hns-lsel-applier` + `decision.json` with a
synchronous-approval marker). M2 emits shadow proposals only. This finding is recorded in
`tier4_firing_test.sh` and cited in the M2 wiring commit.

## CSA forced-gate categories (AC-LSEL-005 / REQ-LSEL-005)

The APPROVE stage (M3, `hns-lsel-applier`) forces a synchronous user-question gate
(orchestrator-run) — regardless of proposer confidence — for any proposal whose blast
radius touches one of the SIX CSA
forced-gate categories:

1. **INVARANTS kernel** — the read-only goal kernel block at the top of `CLAUDE.local.md`.
2. **security/validation exception** bands — input-validation carve-outs, error-handling that
   prevents data loss, OWASP measures.
3. **HIGH-fan-in references** — `@MX:ANCHOR` functions with fan_in ≥ 3 callers.
4. **Bash risk path** — the destructive-primitive set + `BASH_SUBCOMMAND_SOFT_CAP` compound
   commands (coding-standards.md § Bash Risk-Amplifier Doctrine).
5. **`permissions.allow`** additions — explicit security-exception band; per-line synchronous
   approval (every added allow entry is its own forced gate).
6. **execution-meta files** — the four execution-meta categories named in REQ-LSEL-002/005:
   (i) the frozen allowlist meta file at `.claude/lsel/frozen-allowlist.json`, (ii) an applier
   or curator skill body (`hns-lsel-applier/`, `hns-lsel-curator/`), (iii) the apply hook
   script (`lsel-apply.sh` and wrappers), (iv) the `settings.local.json` hook-registration
   subblock.

**Bother-cost-exemption:** forced gates are **bother-cost-exempt** — the bother-cost gating
rule applies ONLY to routine-tier proposals. A forced-gate proposal always triggers a
synchronous user-question gate (orchestrator-run) regardless of bother-cost state.

**Mechanical enforcement (D3):** the applier (`hns-lsel-applier` driving `lsel-apply.sh`,
M3) intercepts every proposal matching the four execution-meta categories and REFUSES to
write unless the proposal's `decision.json` carries an explicit synchronous-approval marker
(an approval artifact produced by the orchestrator's synchronous user-question gate).
A match with no marker aborts the apply,
appends a rejection row to `.moai/logs/lsel-reject.log` naming the matched category, and
writes nothing. Proposals matching none of the four categories proceed through the routine
bother-cost path. This mechanical interception is what makes the self-amending-handcuffs
defense defensible without resting on the regex paradox alone.

`csa_refusal_test.sh` (next to this SKILL.md) is the fixture test for the refusal rule.

---

## REFLECTION stage (M4 — REQ-LSEL-014 / AC-LSEL-016)

The periodic consolidation pass that prevents un-refined accumulation — the
dominant failure mode the design report §10 P4 names: "no consolidation / decay /
pruning → wrong-lesson retrieval". Without REFLECTION, concrete topic files pile
up and retrieval surfaces stale concrete incidents instead of the principle they
collectively support.

### Threshold-fired, not wall-clock-fired

REFLECTION fires when the **accumulated importance** of concrete `feedback_*.md`
topic files clears the threshold (default ~150), NOT on a monthly cron. This is
the Vectorize 4-lever model (importance-gate / merge / decay / evict) the design
report §10 P4 cites: importance is assigned write-time, and the reflection
threshold is an accumulation signal, not a calendar one. A single-topic cohort
below the threshold is a clean no-op (acceptance.md §E edge case).

### The mechanical core — `reflect.sh`

`reflect.sh --memory-dir <m> [--threshold 150] [--min-topics 3]`:

1. Reads the active `feedback_*.md` topic files (maxdepth 1 — never the
   `_archive/` cold tier).
2. Sums their frontmatter `importance`. If `count < min-topics` OR `sum <
   threshold` → clean no-op (exit 0).
3. Synthesizes ONE `feedback_*_principle_*.md` carrying:
   - a `memory_type` label (CoALA taxonomy — `semantic` for a feedback principle;
     `procedural` would route to a `hns-*` skill body instead).
   - the shared theme drawn from the source descriptions (the retrieval cue).
   - `source_count` + `synthesized_at` for the audit trail.
4. **Moves** the originals to `memory/_archive/` (cold tier) — **NEVER deleted**
   (report §10 P4: "축출 ≠ 보관 — 보관은 성능용, 하드 삭제는 규정 준수용; MoAI의
   '삭제 말고 보관' 규칙이 옳음이 입증된다" — archive preserves the audit trail).

### Decay-weighted retrieval

The originals relocate to `_archive/`, so the active recall set (the
`memory/` directory the recall layer scans first) holds the synthesized
principle, NOT the stale concrete originals. A retrieval probe for a related cue
returns the principle ranked ABOVE the archived originals — this is the
decay-weighted retrieval AC-LSEL-016 clause requires. The principle's
`description` is crafted to match the shared cue; the archived originals stay
discoverable (cold tier) but no longer dominate the top of the recall set.

### Model-mediated layer (you, when invoked)

`reflect.sh` performs the mechanical synthesis (deterministic). When this skill
runs a real reflection pass, your job on top is:
- **Read the synthesized principle** and refine its prose into a genuine
  abstract statement (the mechanical core aggregates descriptions; you write the
  actual principle).
- **Confirm the memory_type** — if the consolidated knowledge is procedural (a
  how-to that belongs in a `hns-*` skill body), stamp `memory_type: procedural`
  and route the synthesis to the skill rather than leaving it as a `feedback_*`.
- **Do NOT delete the archive** — originals in `_archive/` are the audit trail.
  If the hot tier (active `feedback_*.md`) approaches the 50-file cap, prefer
  archiving more concrete topics over deleting them (moai-memory.md § Memory
  Hygiene).

### Verification (run before declaring a reflection pass complete)

```bash
# M4 REFLECTION characterization test (AC-LSEL-016) — hermetic temp memory dir:
.claude/skills/hns-lsel-curator/reflect_test.sh

# cold-tier growth vs hot-tier (post-M4 audit, acceptance.md §H):
ls memory/_archive/ | wc -l   # archived originals
ls memory/feedback_*.md | wc -l   # active hot tier
```
