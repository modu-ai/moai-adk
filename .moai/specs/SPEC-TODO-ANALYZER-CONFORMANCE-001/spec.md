---
id: SPEC-TODO-ANALYZER-CONFORMANCE-001
title: "Todo-analyzer documentation conformance: the findings.source enum documents the measured jev value, and the spec_id pick-time promise is abolished in favor of a record-only posture"
version: "0.2.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec (card t1311)
priority: P3
phase: "v3.2.0 target"
module: ".claude/skills/moai/workflows/gtd.md, internal/template/templates/.claude/skills/moai/workflows/gtd.md, internal/cli"
lifecycle: spec-anchored
tags: "kanban, todo, findings, source-enum, jev, spec-id, documentation-conformance, template-mirror"
tier: S
related_specs:
  - SPEC-JEV-CONSUMERS-001
  - SPEC-TODO-ARCHIVE-QUERY-001
  - SPEC-TODO-QUEUE-HOME-CANON-001
---

# SPEC: Todo-analyzer documentation conformance (findings.source enum + spec_id posture)

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-29 | Initial plan-phase authoring (card t1311, Class C, todo 로직 개선 review §P5 — `.moai/reports/todo-logic-review-20260929.md`, primary-checkout-local). Measured in worktree `.moai/worktrees/t1311`, branch `WT-jev-enum-backfill`, HEAD `68e37864a` (local develop tip). Live-queue figures (findings 3/3 source=jev at confidence 0.82/0.28/0.59; picked 9/9 spec_id empty; t472 axis D 53/53 persistent) are carried from the review report's 2026-09-29 measurement of the live queue `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` (seq=1305) and cited, not re-measured. Every `file:line` citation below was measured in this tree at plan time. |
| 0.2.0 | 2026-09-29 | **Plan-audit iteration 1 remediation** (PASS 0.875 vs Tier S 0.75, iteration 1/1 — no re-audit mandated; 3 fix-before-run findings, applied as delta on top of committed `25996dad4`, no REQ or decision changes — REQ-2 abolition stands, audit-verified grounded). **D1 (major)**: AC-TAC-006's diff-scope check "no `internal/` path outside the new test file" was unachievable — the M2 mirror edit itself lands under `internal/template/templates/`; the check now names the exact expected 3-file diff set. **D2 (minor)**: AC→REQ linkage made machine-readable — AC entries converted to leading-list-marker form citing their REQs (spec-lint REQ collection is blind without a leading list marker — t1020/t1057 lesson family), AC-TAC-006 now cites REQ-TAC-005; §B REQ entries likewise carry leading list markers. **D3 (minor)**: REQ-TAC-004's three-carrier enumeration replaced with the audit-recommended thin form ("the association lives in the dispatch and commit carriers") — kanban-dispatch mandates four carriers (PR title included), so the enumeration under-counted; the thin form resolves D3 without doctrine duplication. **D4 (adopted)**: the two `-run` patterns anchored to verified test-name literals (`TestTodoSkillDocument`, `TestJevFinding`, `TestSchemaFreeze` — names grep-verified this tree). **D5 (adopted)**: mid-command line-wraps inside AC checkable text unwrapped. §F open questions closed by the audit: keep P3; take the thin AC-TAC-004 form. spec.md only; plan.md untouched per the delta scope. |

> **Provenance discipline.** Doc-surface and code citations were measured in this worktree at HEAD
> `68e37864a`. Queue-population figures are attributed to the review report's measurement, named as
> such wherever used.

## §A Context

### A.1 Two documented behaviors diverge from measured reality

The todo queue's analysis layer keeps `findings` — relation records about card pairs. The queue's
user-facing documentation (`.claude/skills/moai/workflows/gtd.md` § "Reading the records", and its
template mirror `internal/template/templates/.claude/skills/moai/workflows/gtd.md`) documents the
`source` field as a two-value enum:

> gtd.md:258-259 (both surfaces, measured): "`source` is `mechanical` (measured text similarity)
> or `agent` (a judgement written through `relate`), and a mechanical finding with no agent finding
> on the same pair renders marked `machine-only` ..."

The code and the live queue disagree. `internal/kanban/backlog_store.go:141-158` defines THREE
source constants — `mechanical`, `agent`, and `BacklogSourceJev = "jev"` (SPEC-JEV-CONSUMERS-001,
REQ-JEVN-002), with the code comment explaining why a model answer is deliberately neither of the
other two. The live queue carries 3 findings, all `source=jev` (t1153↔t1145 p=0.82, t1194↔t1172
p=0.28, t1298↔t1297 p=0.59 — review report §큐 실측). The documentation names a two-value world the
emitter left one value ago; a consumer reading the doc would misclassify every real finding.

Second divergence: the same doc promises that `spec_id` is "filled in when the item is picked"
(gtd.md:252-253, both surfaces). Measured practice has never matched: 9/9 picked cards carry an
empty spec_id today, and the same gap was 53/53 persistent at t472 axis D. The pick-time fill is a
design promise the practice has rejected twice, months apart.

### A.2 What the code actually does (the wording must describe this)

`internal/cli/todo_jev_finding.go` is the sole `jev` emission site: at card ADMISSION (`todo add`
path, REQ-JEVN-001 — admission only; the `analyze` re-sweep does not consult it), when the
`workflow.jev.enabled` capability gate is on (shipped default off), a TypeSafe System One answer is
asked which existing queued card, if any, is a near-duplicate of the newly admitted one; when the
answer names a real candidate, one `BacklogFinding` with `Source: jev`, `Relation: near-duplicate`,
and `Score: <model probability>` is appended beside the others. It is record-only: no caller acts
on it (REQ-JEVC-011/012), the render label is the model-confidence form `jev p=%.2f`
(`jevFindingSignalFragment`, todo_jev_finding.go:252-254) — deliberately NOT the word "score",
because a calibrated model confidence must not read as a measured similarity. A `jev` finding is
also NOT part of the `machine-only` mark logic the doc describes for the mechanical/agent pair —
`HasFindingForPairAnySource` suppression (todo_jev_finding.go:127-135) treats it as a third thing.

### A.3 The spec_id decision (REQ-2)

**Decision: option (b) — explicit abolition of the pick-time promise. `spec_id` stays in the
schema as a record-only, optional annotation; the documentation stops promising that a pick fills
it. `moai gtd next <n> --spec <SPEC-ID>` remains a supported record verb.**

Rationale, grounded in the measured evidence:

1. **The practice is unanimous and persistent.** 9/9 picked cards empty today; 53/53 at t472 axis
   D. A design that practice has rejected across two measurements months apart is not a bug in the
   practice — it is a wrong promise in the design.
2. **The card↔SPEC association already has mandatory carriers.** The dispatch's `card:` / `spec:`
   fields, the card id in every commit message on the card's branch, the card id in the PR title,
   and the evidence path (`.moai/reports/<card-id>/…`) are all [HARD] obligations
   (`kanban-dispatch.md`). The queue column is a redundant carrier that nothing fills.
3. **Auto-backfill needs machinery that does not exist.** Plan completion is an agent-driven
   phase, not a CLI verb — an automatic backfill would require a new hook or command to bridge
   "SPEC authored" to "queue row updated", plus resolution logic (which queue? the home DB
   resolved per project) for a linkage the carriers above already deliver. The cost exceeds the
   conformance need this card was issued for.
4. **The existing promise is internally wrong anyway.** gtd.md:252-253 says a `null` spec_id is
   "what distinguishes a backlog item from a card already on the board" — it is the `state` field
   (`queued` | `picked`), not spec_id, that does that. Abolishing the promise removes a second
   factual error with it.

The alternative (a) auto-backfill at plan completion is recorded as REJECTED in §E Exclusions.

## §B Requirements (GEARS)

- REQ-TAC-001 (Event-driven): **When** a reader consults the queue documentation's `findings`
field description in `.claude/skills/moai/workflows/gtd.md` § "Reading the records", the
documentation shall enumerate all three `source` values the analyzer emits — `mechanical`,
`agent`, and `jev` — with no fourth value implied.

- REQ-TAC-002 (Ubiquitous): The `jev` wording shall describe WHEN the value is emitted: at card
admission only, when the `workflow.jev.enabled` capability gate is on, and the model answer names
an existing card as a near-duplicate; and shall carry the two properties the code pins — the score
is a model confidence rendered `p=`, never labelled a measured similarity, and a `jev` finding is
a record a person reads, never an input to a queue mutation, dispatch decision, or completion
verdict.

- REQ-TAC-003 (Ubiquitous): The template mirror
(`internal/template/templates/.claude/skills/moai/workflows/gtd.md`) shall carry the identical
enum wording as the live skill document, while remaining internally neutral (no SPEC ID, no REQ
token, no internal date, no commit SHA — the template-neutrality contract enforced by
`todo_skill_doc_test.go`).

- REQ-TAC-004 (Ubiquitous): The `spec_id` field description on both surfaces shall state the
record-only posture: it is an optional annotation attached only when `next --spec` is explicitly
given; an empty spec_id on a picked card is the normal state; the card↔SPEC association lives in
the dispatch and commit carriers — and the "filled in when the item is picked" promise shall not
appear on either surface.

- REQ-TAC-005 (Event-driven): **When** the doc surfaces are edited in a future change, a
doc-parity guard test in `internal/cli` (modeled on `TestTodoSkillDocumentsHistoryVerb`,
`todo_skill_doc_test.go`) shall fail unless BOTH gtd.md surfaces enumerate the `jev` source value
and neither carries the abolished pick-time promise — so the two-cell conformance cannot silently
regress.

## §C Acceptance Criteria (inline, Tier S)

Two-cell discipline: every AC names the RED-now state observed on this tree at plan time, and the
green path naming the milestone that flips it. Each AC cites its covering REQ(s) in the corpus
canonical `(maps REQ-…)` list-marker form so spec-lint collects the linkage (D2).

- AC-TAC-001: maps REQ-TAC-001, REQ-TAC-002 — live surface documents jev. Given this tree at HEAD
  `68e37864a`, When the source-enum sentence of `.claude/skills/moai/workflows/gtd.md` § "Reading
  the records" is read: **RED-now (observed)** — lines 258-259 name only `mechanical` and
  `agent`; `jev` appears in gtd.md only at line 330 (the `--auto` display-only mention); grep
  evidence: `grep -n "jev\|Jev" .claude/skills/moai/workflows/gtd.md` → single hit at :330.
  **Green path (M2)**: the findings bullet enumerates `mechanical`, `agent`, and `jev`, and the
  `jev` clause states the admission-only, gate-on, model-confidence, record-only properties of
  §A.2. Then: PASS when `grep -n "jev\|Jev" .claude/skills/moai/workflows/gtd.md` returns a hit
  inside the findings field description AND the enumeration sentence contains all three values.

- AC-TAC-002: maps REQ-TAC-003 — mirror parity. Given the template mirror
  `internal/template/templates/.claude/skills/moai/workflows/gtd.md`: **RED-now (observed)** —
  mirror lines 258-259 carry the same two-value enum wording as the live surface. **Green path
  (M2)**: the mirror's findings bullet carries the identical (modulo mirror-neutrality) enum
  wording. Then: PASS when `grep -c "jev" internal/template/templates/.claude/skills/moai/workflows/gtd.md`
  over the findings-description region is ≥ 1 on both surfaces' enum sentences.

- AC-TAC-003: maps REQ-TAC-003 — mirror neutrality preserved. Given the edited mirror: **RED-now
  (observed)** — the neutrality scan in `TestTodoSkillDocumentsHistoryVerb`
  (`todo_skill_doc_test.go:45-50`) passes today; it must still pass after the edit. **Green path
  (M2/M3)**: the mirror carries no SPEC ID, REQ token, internal date, or commit SHA. Then: PASS
  when `go test ./internal/cli/ -run '^TestTodoSkillDocumentsHistoryVerb$'` exits 0 post-edit.

- AC-TAC-004: maps REQ-TAC-004 — pick-time promise abolished. Given both gtd.md surfaces: **RED-now
  (observed)** — `grep -n "filled in when the item is picked" .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md`
  returns one hit each, both at :252 (exit=0 measured). **Green path (M1)**: the grep returns
  zero hits on both surfaces; the replacement bullet states the record-only posture. Then: PASS
  when the same grep exits 1 (zero matches) on both paths AND the record-only wording is present
  on both.

- AC-TAC-005: maps REQ-TAC-005 — parity guard. Given `internal/cli/todo_jev_finding_test.go`
  (measured: no doc-surface coverage today — the guard is genuinely new): **RED-now (observed)**
  — no test in `internal/cli` asserts the gtd.md surfaces' findings-source or spec_id wording;
  reverting the doc edit would break nothing. **Green path (M3)**: a new doc-parity test named
  `TestTodoSkillDocumentsJevSource` (pattern of `todo_skill_doc_test.go`) fails when either
  surface drops the `jev` enumeration or regains the pick-time promise. Then: PASS when the
  seeded-regression proof holds — doc regression in a scratch copy → test fails; restore →
  `go test ./internal/cli/ -run '^TestTodoSkillDocumentsJevSource$'` exits 0 on the edited tree.

- AC-TAC-006: maps REQ-TAC-005 — regression containment and diff scope. Given this SPEC is
  documentation plus one new test file: **RED-now (observed)** — tree clean at plan HEAD
  `68e37864a` (git status clean, measured). **Green path (M3)**: the anchored affected-package
  measurements `go test ./internal/cli/ -run '^(TestTodoSkillDocumentsHistoryVerb|TestTodoSkillDocumentsJevSource|TestJevFinding_WrittenAtAdmission)$'`
  and `go test ./internal/kanban/ -run '^TestSchemaFreezeRecordsTransitionStamps$'` exit 0
  (test-name literals verified to exist this tree for the first, third, and fourth anchors; the
  second is the M3-created guard pinned by AC-TAC-005). Then: PASS when those exit 0 AND
  `git diff --name-only` against the base names exactly the expected 3-file set —
  `.claude/skills/moai/workflows/gtd.md`, `internal/template/templates/.claude/skills/moai/workflows/gtd.md`,
  and the one new `internal/cli/*_test.go` — and nothing else.

## §D Constraints

- Template-First (`CLAUDE.local.md` §2): the mirror under `internal/template/templates/` is the
  distribution source; live and mirror are edited in the same change, and `make build` re-embeds
  before verification. The emitted `.toml` agent layer is untouched (no agent file changes).
- Doc wording stays template-neutral (no card ids, SPEC ids, dates, SHAs in the mirror) — the
  live surface MAY carry them.
- No schema change: `items.spec_id` stays a column; `backlog_schema_freeze_test.go` is untouched.
- The jev wording must not weaken the REQ-JEVN-003 property: `jev` must never be documented as or
  conflated with `agent` (the machine-only mark logic).

## §E Exclusions

### Out of Scope — analyzer behavior
- No change to `BacklogFinding` emission, dedup keys, `HasFindingForPairAnySource`, or the
  `analyze` re-sweep — the code is conformant; only the documentation is not.
- No removal of `items.spec_id`, no migration, and REJECTED ALTERNATIVE (a): no auto-backfill of
  spec_id at plan completion (rationale §A.3.3).

### Out of Scope — sibling review items
- Review §P1 (stale-store notice), §P2 (primary-checkout doc staleness), §P3 (hold state — landed
  via SPEC-TODO-HOLD-STATE-001), §P4 (blocks relation), §P6 (session records): separate cards.

### Out of Scope — other spec_id doc surfaces
- `.claude/skills/moai-kanban-foreman/SKILL.md:157` lists `spec_id` among record fields and makes
  no pick-time promise — no edit required.
- `.moai/docs/todo-queue-storage.md` § "Factory assignment provenance" describes what happens
  WHEN `next --spec` IS given (the `card.assigned` event) — accurate as written; unchanged.

## §F Open Questions

Both open questions from v0.1.0 were closed by plan-audit iteration 1 (no open questions remain):

1. Priority P3 vs the review's "Low" triage of §P5 — CLOSED: keep P3 (audit adjudication).
2. AC-TAC-004 carrier enumeration vs thin form — CLOSED: take the thin form ("the association
   lives in the dispatch and commit carriers"), applied to REQ-TAC-004 in v0.2.0 (D3).
