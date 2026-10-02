---
id: SPEC-JEV-AUTO-EXCEPTION-001
title: "Linked amendment of the Jev display-only principle for the todo --auto selection-order exception — Jev-side surfaces, two completed SPECs, and a linkage guard"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1403)
priority: P2
phase: "v3.2.0 target"
module: "internal/jev, internal/cli, internal/template, .claude/rules, .moai/config, .moai/specs, .moai/docs"
lifecycle: spec-anchored
tags: "jev, display-only, doctrine-amendment, linkage-guard, todo-auto, selection-order, card-t1403"
tier: M
card: t1403
depends_on: [SPEC-TODO-AUTO-PRIORITY-001]
related_specs: [SPEC-JEV-CORE-001, SPEC-MANAGER-TODO-001, SPEC-AUTONOMY-GATE-REWIRE-001, SPEC-JEV-OPTIN-MEASURE-001, SPEC-JEV-GUARD-001, SPEC-JEV-CONSUMERS-001, SPEC-JEV-GOAL-DIST-001]
---

# SPEC: linked amendment of the Jev display-only principle for the `todo --auto` exception

## HISTORY

- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1403; worktree
  `.moai/worktrees/t1403`, branch `WT-jev-auto-exception`, HEAD `c50da9c2f`).
  Tier M: spec.md + plan.md + acceptance.md, plus research.md carrying the
  inventory and baseline evidence (verbatim output above 50 lines). Operator
  decisions 5 and 6 of `SPEC-TODO-AUTO-PRIORITY-001` §B.1 are inherited as
  binding input and are not re-opened here.

## §A Context

### A.1 Problem — the split state is live on develop now

`SPEC-TODO-AUTO-PRIORITY-001` (card t1400, merged at `c50da9c2f`) gave the
`moai todo --auto` cycle an `auto-scoped ranking exception`: it may rank the
queued candidates it is about to accept using a Jev answer, which changes its
selection order only. Operator decision 5 of that SPEC amended three documents
(`kanban-dispatch.md`, `workflows/gtd.md`, `manager-todo.md`, live and mirror)
and deferred every other surface that still states the Jev display-only principle
to a follow-up card. This SPEC is that card (t1403). Until it lands, one tree
carries two incompatible statements:

- **The landed side says an exception exists.** `.claude/agents/moai/manager-todo.md:7`
  reads "Jev as a display-only signal for dispatch order and priority, except for
  the `--auto` cycle's own selection order"; `kanban-dispatch.md:31` and
  `workflows/gtd.md` carry the two literals `auto-scoped ranking exception` and
  `selection order only`; `internal/cli/todo_auto.go:27-30` says the same.
- **The not-yet-amended side says none does.** `.moai/config/sections/workflow.yaml:227`
  (template `:229`): "The capability is display-only: an answer is … never an
  input to a completion verdict, a merge approval, a queue mutation"; the
  `internal/jev/jev.go:24` package comment; the `internal/cli/mcp_jev.go:8`
  comment; the MCP tools catalogue rows `:139` and `:233` (live and mirror);
  `SPEC-JEV-CORE-001` REQ-JEVC-011 (`spec.md:81`, "The capability is display-only,
  with exactly one exception"), REQ-JEVC-012 (`spec.md:99`, no consultation "as an
  input" for "a `moai todo` … mutation"), and the two `Out of Scope — authority`
  bullets (`spec.md:160-161`); `SPEC-MANAGER-TODO-001` REQ-MT-014 and REQ-MT-015
  (`spec.md:69`, `:71`).
- **The consumer is wired.** `internal/cli/todo_auto_rank.go:474-481` builds
  `jev.New(true)` and calls `client.Ask`; the consumer-set guard in
  `internal/cli/doctor_jev_test.go` declares it ("a Jev answer orders candidates
  and nothing else; no ordering accuracy is claimed"). It ships behind
  `workflow.jev.enabled`, whose code default is `false`
  (`internal/config/defaults.go:1180-1182`).

A reader of the Jev-side text is therefore told that a `moai todo` pick can never
take a Jev answer as input, while the code and three documents say it can, for
ordering. **This SPEC removes that split** by amending the Jev-side surfaces in one
linked change and adding a guard that fails if they ever diverge again.

### A.2 Verified basis (this tree, HEAD `c50da9c2f`)

Full evidence with commands and verbatim output is in `research.md`; the facts the
requirements stand on:

- **Six surfaces named by the card, four more found.** The card lists S1-S6
  (§B.2). A repository sweep (69 `display-only` hits in 43 tracked files outside
  `.moai/specs/`, `.moai/reports/` and `CHANGELOG.md`, plus a synonym sweep) found
  four further class-(i) surfaces, X1-X4, none on the card (§B.3).
- **Four guards already pin the first amendment's text and must stay green
  unmodified** (§B.8): `TestJevAmendmentLinkage`, `TestJevDoctrineAmendment`,
  `TestMCPToolCatalogueDocsStayMirrorIdentical`, and the consumer-set guards.
  Baseline at this HEAD: all PASS (research.md §R3).
- **The shipping-gate guard does not touch the ranking path.**
  `TestNoConsumerCallPathShips` scans non-test Go files under `internal/` for the
  literals `NearDuplicateMark`, `LaneQuestionRoute`, `SkillSuggest`; it PASSes at
  this HEAD with the ranking consumer present (§B.6).
- **The exception is not reachable through the MCP tool.** The ranking consumer
  calls the client directly (`todo_auto_rank.go:474-481`); the tool handler
  `handleJevAsk` (`mcp_jev.go:67`) is referenced from `mcp_jev.go` only (§B.4).

### A.3 Why the exception is order-only — the property the amendment text states

The amended passages may assert only what the landed behavior already makes true.
`SPEC-TODO-AUTO-PRIORITY-001` fixes it: the ranking stage runs after the existing
eligibility filters (REQ-TAP-001); the `blocked` exclusion applies on both ranking
sources (REQ-TAP-006); a candidate set larger than the request bound places the
surplus after the Jev-ordered candidates, so nothing is dropped (REQ-TAP-002); an
answer set that fails validation makes the whole Jev result unavailable and the
cycle falls back to recorded priority over the same set (REQ-TAP-003, -011); the
stage writes no queue field (REQ-TAP-010). The candidate **set** is therefore fixed
by mechanical filters before any answer is read, and a Jev answer can only permute
it. That is the property that separates this exception from a Jev answer deciding a
queue change, and it is the property the amended text names. This SPEC cites those
requirements and does not re-derive them; eight behavior tests that evidence them
were re-run at this HEAD and PASS (research.md §R3).

### A.4 Why the first exception's guard cannot simply be extended

The contract-mode Kickoff exception (commit `185569ef3`, card t1236) was a
multi-surface amendment protected by `TestJevAmendmentLinkage`
(`internal/contract/kickoff/activation_test.go:208`). `linkageFindings` requires its
markers to be all present or all absent **and** to first appear in one commit
(`firstCommit` = the oldest `git log -S<token> -- <path>` hit). Appending the new
surfaces to `linkageMarkers` would compare their first commit with the old
amendment's commit and fail permanently; the old guard also keys part of its set to
`JevDoctrineAmended = true` in `kickoff.go`, which this exception does not touch.
The new amendment therefore needs a sibling guard with its own marker set (§B.5).

## §B Decisions

### B.1 Inherited operator decisions (binding input — do not re-open)

From `SPEC-TODO-AUTO-PRIORITY-001` §B.1, recorded 2026-10-02:

5. **Doctrine scope.** That SPEC amended only `kanban-dispatch.md`,
   `workflows/gtd.md` and `manager-todo.md`; the Jev-side surfaces belong to this
   follow-up card, to be implemented as a linked amendment.
6. **Jev ordering consumer and accuracy.** The ordering consumer ships only behind
   the default-off gate `workflow.jev.enabled`; enabling it is the operator's act.
   No requirement or criterion of this SPEC claims or implies ordering accuracy.

The wording constraint of that SPEC's §B.5 also binds here: each amended passage
carries the literals `auto-scoped ranking exception` and `selection order only` in
one paragraph, keeps every prohibition on every other surface, and does not state
or imply that the principle is amended outside the `--auto` cycle.

### B.2 Inventory result and classification rule

Every restatement of the principle found in the repository is classified:
**(i)** states the principle and contradicts the exception — in scope;
**(ii)** historical record (a SPEC body other than the two named, a CHANGELOG
entry, a report, a generated map) — left untouched; **(iii)** already compatible,
or a statement about a different object — left untouched, with the reason.

| ID | Surface (verified line, this tree) | Class | Disposition |
|---|---|---|---|
| S1 | `internal/jev/jev.go:24-28` package comment (also `:27`, a reference to a test file that does not exist) | (i) card | amend, comment only |
| S2 | `.moai/config/sections/workflow.yaml:227-232`; template `:229-234` — the `workflow.jev` comment | (i) card | amend both |
| S3 | `internal/cli/mcp_jev.go:8-10` doc comment | (i) card | amend, comment only |
| S4 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` `:139` and `:233` rows; template mirror, same lines | (i) card | amend both copies, byte-identical |
| S5 | `.moai/specs/SPEC-JEV-CORE-001/spec.md` REQ-JEVC-011 `:81`, REQ-JEVC-012 `:99`, authority bullets `:160-161`, HISTORY, frontmatter | (i) card | amend, v0.4.0 |
| S6 | `.moai/specs/SPEC-MANAGER-TODO-001/spec.md` REQ-MT-014 `:69`, REQ-MT-015 `:71`, HISTORY, frontmatter | (i) card | amend |
| X1 | `.moai/docs/jev-local-operations.md:28-32` — the doctrine guide the first amendment's linkage guard names; "예외는 한 곳뿐이다" | (i) **extension** | additive paragraph only |
| X2 | `.claude/rules/moai/development/agent-authoring.md:147`; template `:147` — "consults Jev as a display-only signal" | (i) **extension** | amend both |
| X3 | `.claude/skills/moai/SKILL.md:180`; template `:180` — "never reorder by inferred priority" | (i) **extension** | amend both; regenerates the skill's catalog hash |
| X4 | `CLAUDE.md:63`; template `:63` — "Jev display-only consultation" | (i) **extension, weakest** | amend both, bounded growth |

The (ii) and (iii) hits and the reason each is left alone are enumerated in
§D and in `research.md` §R1.

### B.3 Scope extensions beyond the six card-named surfaces

The card names S1-S6. X1-X4 are included because the instruction for this SPEC is
to include class-(i) hits in the same doctrine family; each is **marked as a scope
extension for confirmation** and the SPEC is written so cutting one removes its
row and nothing else (the guard's marker registry is derived from the confirmed
surface list, `plan.md` §B D-3).

- **X1** is not optional in practice: it is the doctrine guide whose pinned
  paragraph states "the exception is one place only", and the first amendment's own
  guard lists it. Because `TestJevDoctrineAmendment/local-guide` requires that
  paragraph verbatim, X1 can be amended **additively only** (§B.8 P3).
- **X2, X3, X4** state the principle for `manager-todo` or the pick. Evidence and
  cost: X3 edits a skill directory, so `make build` regenerates a hash in
  `internal/template/catalog.yaml`, which then belongs to the same commit; X4 edits
  the always-loaded instruction file (15,573 bytes in both copies), so growth is
  bounded in `plan.md` §D. X3 and X4 are the first cut candidates.

### B.4 The tool is not widened: capability versus call path

The exception is an in-process consumer of `internal/jev`; it never goes through
the `jev_ask` MCP tool (`todo_auto_rank.go:474-481` versus `handleJevAsk`,
`mcp_jev.go:67`). The tool therefore **stays display-only for every agent that
calls it**. Consequences, all decided here:

- The tool's registered description string (`mcp_jev.go:48`, "DISPLAY-ONLY: …"),
  its input schema, and the `internal/mcp/catalog.go:97` entry and comment stay
  unchanged. No test pins the description string (§R3), so this is a choice, not a
  constraint; it is made because the string is agent-visible tool metadata and an
  exception it cannot honour would invite calls it must refuse.
- The doc comment at `mcp_jev.go:8` is amended to say the exception exists and does
  not reach this tool, so a reader of the wrapper is not misled in either direction.
- The catalogue rows (S4) record the exception because they describe the
  capability's authority; the first exception was recorded there the same way.
- The user-facing docs-site pages for the tool (four locales) describe the tool
  and stay as they are (§D); this follows the first exception's precedent.

### B.5 Guard design — a sibling, not an extension

Decision: a new test file `internal/template/jev_auto_exception_test.go` in the
existing `template_test` package, reusing its helpers `grRoot`, `grRead`, `grGit`,
`grReqBody` and `grSection`. It carries two tests:

- `TestJevAutoExceptionLinkage` — the analogue of `TestJevAmendmentLinkage`.
  **Marker registry:** the universal token `auto-scoped ranking exception` in every
  confirmed surface file (S1-S6, X1-X4, live and mirror) plus the arming constant
  `jevAutoExceptionAmended = true` in the test file itself. **Anchors:** the three
  landed documents (`kanban-dispatch.md`, `gtd.md`, `manager-todo.md`) carry the same
  token but predate this amendment, so they are *presence-only* anchors and are kept
  out of the first-commit comparison. The checks: markers all present or all absent;
  all first appearing in one commit; and no dangling amendment (a marker present
  while an anchor lacks the token).
- `TestJevAutoExceptionWording` — the analogue of `TestJevDoctrineAmendment`:
  per surface group, every passage that says "display-only" carries both literals in
  the same passage, keeps its closed targets, and a fixture without the literal is
  rejected.

Why a sibling and why here. Appending to `linkageMarkers` fails (§A.4). Refactoring
`linkageFindings` to take a marker set would modify a guard that must stay
byte-for-byte green, to save about a dozen lines of helper. The `template_test`
package already holds the wording guard for the first amendment and the file
helpers; `internal/contract/kickoff` is about the contract-mode Kickoff and
`internal/jev` is restricted to the standard library. The universal token is
template-neutral (no SPEC id, requirement token, date or hash), so it may sit in
mirrors; SPEC files carry it in the same paragraph as their version marker.

**Staged landing is resolved by an arming constant, not by a weaker rule.**
`jevAutoExceptionAmended` ships `false` in the guard's own commit (M1), when every
other marker is absent, so the tree check passes; the single linked commit flips it
to `true` and, being a marker, fails the guard if it lands without the others. This
is the same device `JevDoctrineAmended` is for the first guard. The guard commit
precedes the marker commit in the commit graph (REQ-JAE-011), which is the only
witness of ordering (`verification-claim-integrity.md` §2.3).

### B.6 `REQ-JEVO-009` accuracy-label set — excluded, with the interaction measured

`SPEC-JEV-OPTIN-MEASURE-001` REQ-JEVO-009 ("a consumer whose measured accuracy does
not beat its own constant-answer baseline shall not be shipped") and
`SPEC-JEV-GUARD-001` REQ-JEVG-001 bar a *named* consumer call path until its
measurement runs. The labelled accuracy set for an ordering question does not
exist. **Default taken: out of scope.** It is a measurement and shipping-gate axis,
not a doctrine-text axis, and decision 6 already fixed that no accuracy is claimed.

Interaction checked, not inferred: at this HEAD
`go test -count=1 -v -run '^TestNoConsumerCallPathShips$' ./internal/jevmeasure/`
prints `--- PASS: TestNoConsumerCallPathShips (0.19s)`. The test walks non-test Go
files under `internal/` for `NearDuplicateMark`, `LaneQuestionRoute` and
`SkillSuggest`; the ranking consumer is not among the three markers it scans for,
which is why the test passes with the consumer shipped. This SPEC's only Go edits
are comments, and REQ-JAE-008 forbids those three literals in them. **No evidence
found that inclusion is unavoidable.**

### B.7 Amending two completed SPECs

`SPEC-JEV-CORE-001` and `SPEC-MANAGER-TODO-001` are `status: completed`. The
amendment follows the exact style the first exception used on `SPEC-JEV-CORE-001`
(v0.3.0): a `[AMENDED <landing date> — v0.4.0; see HISTORY]` marker beside the
existing ones on REQ-JEVC-011 and REQ-JEVC-012, an exception paragraph written like
"The one gate exception (v0.3.0)", the exception named inside the two existing
authority bullets (their count stays two), a new HISTORY row, a version bump, and
`status: completed` unchanged. It does **not** use the heavier `completed →
in-progress (amendment)` transition with `amendment_of:` that the frontmatter schema
describes: the v0.2.0 and v0.3.0 amendments set the precedent for narrow in-place
amendments of this SPEC without it. Flagged for confirmation (plan.md §B A-2).
Authoring ownership: SPEC bodies belong to `manager-spec`, so the run phase
re-delegates those two files to it (the D-NEW-1 pattern), then `manager-develop`
writes everything else; the first amendment was assembled the same way
(`SPEC-AUTONOMY-GATE-REWIRE-001/design.md` §11.2).

### B.8 Pinned text the amendment must preserve

| Pin | Guard (package) | What it requires of the edited text |
|---|---|---|
| P1 | `TestJevDoctrineAmendment/spec` (`internal/template`) | `SPEC-JEV-CORE-001` stays `status: completed`; REQ-JEVC-011 and -012 keep `[AMENDED 2026-09-26`, `contract-mode Kickoff`, `llm+jev`; the region from `### Out of Scope — authority` to the next `## ` heading holds **exactly two** `- ` items, each naming `contract-mode Kickoff` and `llm+jev`; HISTORY keeps a row with `2026-09-26`, `0.3.0`, `Jev-alone`, `cross-check`; a requirement paragraph ends at the next `\n**REQ-` or heading |
| P2 | `TestJevDoctrineAmendment/rules-and-config` | each catalogue row (`\| \`mcp__moai__jev_ask\` \|`, `\| Judgment (gated) \|`) and each `workflow.yaml` jev comment (from `# jev: ` to `\n    jev:`) names `contract-mode Kickoff` **exactly once**, keeps `llm+jev`, `never decides alone` and the closed targets (`completion predicate`/`completion verdict`, `merge approval`/`merge`, `queue mutation`) |
| P3 | `TestJevDoctrineAmendment/local-guide` | `.moai/docs/jev-local-operations.md` contains the §29 amendment text of `SPEC-AUTONOMY-GATE-REWIRE-001/design.md` **verbatim**; `AGENTS.local.md` §29 keeps its pointer and the markers `판단 자료일 뿐`, `판정 근거로 쓰지 않는다` |
| P4 | `TestJevAmendmentLinkage` (`internal/contract/kickoff`) | tokens `[AMENDED 2026-09-26`, `contract-mode Kickoff` (both catalogue copies, both `workflow.yaml` copies), the guide token, `JevDoctrineAmended = true` keep their original first commit; `kickoff.go` is not touched |
| P5 | `TestMCPToolCatalogueDocsStayMirrorIdentical` (`internal/cli`) | the two catalogue copies stay byte-identical |
| P6 | `TestNoConsumerCallPathShips` (`internal/jevmeasure`) | no non-test Go file under `internal/` gains `NearDuplicateMark`, `LaneQuestionRoute` or `SkillSuggest` |
| P7 | `TestPackageImports_AreStandardLibraryOnly` (`internal/jev`) | `jev.go` imports stay standard library |
| P8 | `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestAutoRankAgentDoctrine` (`internal/cli`) | the three landed documents are not edited |
| P9 | `TestTemplateNoInternalContentLeak` and the neutrality tests (`internal/template`) | template mirrors carry no SPEC id, requirement token, ISO date or commit hash |

## §C Requirements

Verification layer: `acceptance.md`. The requirement layer below is GEARS.

- **REQ-JAE-001** (Ubiquitous) — Every in-scope surface that states a Jev answer is
  display-only (inventory S1-S6 and the confirmed extensions X1-X4, §B.2-B.3) shall,
  in the same passage, name the `todo --auto` exception with the literal
  `auto-scoped ranking exception` and bound it with the literal
  `selection order only`, and shall identify it as the `--auto` cycle's own ranking
  of the queued candidates it is about to accept. A long-form passage (S1, S2, S3,
  S5, S6, X1) shall also state that mechanical filters fix the candidate set before
  any Jev answer is read and that the capability sits behind the default-off
  `workflow.jev.enabled` gate; a short-form passage (an S4 table row, X2, X3, X4)
  need not. A passage is the unit the surface's existing guard already scopes: a
  comment block, a table row, a requirement paragraph, or a doctrine paragraph.

- **REQ-JAE-002** (Ubiquitous) — Every amended passage shall keep each prohibition it
  carried before the amendment on every other surface (completion verdict or
  predicate, merge approval, operator gate, user-surface behaviour change,
  CodeRabbit slot-wait adjudication, and any change to a card other than the
  cycle's own existing pick, unpick and done transitions), shall not state or imply
  that the Jev display-only principle is amended anywhere outside the `--auto`
  cycle, and shall not claim or imply that the Jev ordering is accurate or measured.

- **REQ-JAE-003** (Unwanted) — No amended passage shall state or imply that the
  exception is reachable through the `jev_ask` MCP tool; the tool's registered name,
  description string, input schema and catalog entry shall remain unchanged, and a
  passage that describes the tool shall say the exception is an in-process consumer
  that never reaches it.

- **REQ-JAE-004** (Ubiquitous) — For every live/mirror pair the mirror shall carry the
  same amended wording as the live copy (byte-identical where the pair is
  byte-identical today, which is the catalogue pair), and every file under
  `internal/template/templates/` shall carry no SPEC id, requirement token, ISO
  date or commit hash.

- **REQ-JAE-005** (Where) — **Where** an existing guard pins text in a passage this
  amendment touches (§B.8 P1-P9), the amendment shall keep the pinned text, counts
  and file identities, so that every one of those guards passes without being
  modified.

- **REQ-JAE-006** (Event-driven) — **When** the linked amendment lands,
  `SPEC-JEV-CORE-001` shall carry version `"0.4.0"` with a refreshed `updated`, an
  `[AMENDED <landing date> — v0.4.0; see HISTORY]` marker and an exception paragraph
  on REQ-JEVC-011 and on REQ-JEVC-012, the exception named inside both existing
  `Out of Scope — authority` bullets (still exactly two), a HISTORY row recording the
  amendment, `status: completed`, no requirement removed or renumbered, and
  `sync_commit_sha` in its `progress.md` unchanged.

- **REQ-JAE-007** (Event-driven) — **When** the linked amendment lands,
  `SPEC-MANAGER-TODO-001` shall carry on REQ-MT-014 and on REQ-MT-015 a scope
  sentence stating that each governs consultation of the local Jev scripts, which
  stays display-only, and that the `--auto` cycle's own Jev ranking is the single
  exception under `SPEC-TODO-AUTO-PRIORITY-001`; plus a HISTORY row, a version
  bump, `status: completed`, and no requirement removed or renumbered.

- **REQ-JAE-008** (Unwanted) — The amendment shall not change shipped behavior:
  every changed line of a non-test Go file shall be a comment line; the gate default
  (`Enabled: false`), `todo_auto_rank.go` and the MCP registration shall be
  untouched; `internal/jev` shall remain standard-library-only; the edited comments
  shall contain none of `NearDuplicateMark`, `LaneQuestionRoute`, `SkillSuggest`; and
  no agent definition shall change, so no `make agents-emit` is owed.

- **REQ-JAE-009** (Ubiquitous) — A new automated check, `TestJevAutoExceptionLinkage`,
  shall fail when (a) the marker registry is partially present, (b) its tokens first
  appear in more than one commit, or (c) a marker is present while a landed anchor
  lacks the token; it shall carry a falsifier subtest for each failure, a passing
  all-in-one-commit subtest and a tree subtest, shall include the arming constant
  `jevAutoExceptionAmended` in the registry, and shall leave `TestJevAmendmentLinkage`
  and its marker list unmodified.

- **REQ-JAE-010** (Ubiquitous) — A new automated check, `TestJevAutoExceptionWording`,
  shall verify per surface group that every passage stating a Jev answer is
  display-only also carries both literals and its closed targets, and shall carry a
  falsifier subtest that rejects a passage lacking the literal, one that drops a
  closed target, and one whose bound literal sits in another paragraph.

- **REQ-JAE-011** (Event-driven) — **When** the amendment is implemented, the guard
  (with `jevAutoExceptionAmended = false`) shall land in a commit that is an ancestor
  of the commit that first adds any marker token; every marker token, including
  `jevAutoExceptionAmended = true`, shall first appear in one commit; the two
  completed-SPEC edits shall be written by `manager-spec` and every other surface by
  `manager-develop`, one writer at a time; and the commit shall be staged by explicit
  pathspec.

- **REQ-JAE-012** (Event-driven) — **When** the linked commit has landed, a sweep of
  tracked files outside `.moai/specs/`, `.moai/reports/` and `CHANGELOG.md` for
  display-only restatements (the two patterns in `plan.md` §F) shall return no
  passage that states a Jev answer is display-only without the exception unless
  `research.md` classifies it (ii) or (iii), and the sweep shall be accompanied by a
  positive-control hit.

### C.1 Traceability

REQ-JAE-001 → AC-JAE-001 + AC-JAE-002 + AC-JAE-003 + AC-JAE-004 + AC-JAE-005 +
AC-JAE-006 + AC-JAE-007 + AC-JAE-014 · REQ-JAE-002 → AC-JAE-001 .. AC-JAE-007 ·
REQ-JAE-003 → AC-JAE-001 + AC-JAE-003 + AC-JAE-010 · REQ-JAE-004 → AC-JAE-002 +
AC-JAE-003 + AC-JAE-007 + AC-JAE-008 · REQ-JAE-005 → AC-JAE-002 + AC-JAE-003 +
AC-JAE-004 + AC-JAE-006 + AC-JAE-009 · REQ-JAE-006 → AC-JAE-004 · REQ-JAE-007 →
AC-JAE-005 · REQ-JAE-008 → AC-JAE-001 + AC-JAE-010 · REQ-JAE-009 → AC-JAE-011 ·
REQ-JAE-010 → AC-JAE-012 · REQ-JAE-011 → AC-JAE-013 · REQ-JAE-012 → AC-JAE-014
(bodies and `**Covers**` clauses in `acceptance.md`).

## §D Out of Scope

### Out of Scope — Jev accuracy and the shipping gate

- The labelled accuracy set and the measurement that `SPEC-JEV-OPTIN-MEASURE-001`
  REQ-JEVO-009 and `SPEC-JEV-GUARD-001` REQ-JEVG-001 require before a named consumer
  ships. This SPEC measures nothing, certifies no ordering accuracy, and claims none
  in any requirement, criterion or amended passage (§B.6). Enabling
  `workflow.jev.enabled` stays the operator's act.
- Any edit to `TestNoConsumerCallPathShips` or to the `gate_demo_test.go` marker list.

### Out of Scope — the MCP tool surface and its user-facing docs

- The `jev_ask` registered name, description string (`mcp_jev.go:48`), input schema,
  the `internal/mcp/catalog.go:97` entry and comment, and the four locale pages
  `docs-site/content/{en,ko,ja,zh}/guides/mcp-server.md` (`:208`/`:212`). They
  describe the tool, the tool stays display-only (§B.4), and the first exception left
  them unchanged. A docs-site pass would owe the four-locale same-PR obligation and
  is a separate card if the operator wants it.

### Out of Scope — behavior and configuration

- The ranking code (`todo_auto_rank.go`, `todo_auto.go`), the gate default, the
  credential path, the request bounds, and any change to what the `--auto` cycle
  does. This SPEC edits comments, prose, requirements text and one new test file.

### Out of Scope — historical records and generated maps (class ii)

- `CHANGELOG.md` entries (lines `24`, `35`, `52`, `72`, `124`, `132`, `205`) and the
  bodies, plans, designs and progress records of every other SPEC, including
  `SPEC-TODO-AUTO-PRIORITY-001` §B.5, §D and §G R-6, which truthfully describe the
  interim state at the time; this SPEC's HISTORY and §A.1 supersede them for the
  present. Tracked `.moai/reports/` content is likewise a record.
- `.moai/project/codemaps/docs-truth.md:43`, a generated map that follows
  `CLAUDE.md` on regeneration.

### Out of Scope — surfaces already compatible (class iii)

- `manager-todo.md` (and its template and `.codex` emission), `kanban-dispatch.md`,
  `workflows/gtd.md` — amended by the predecessor card.
- `auto-semantics.md` §12 (`:217`) and its mirror: a statement about the `jev_ask`
  tool as the lane watchdog uses it, which stays true.
- `moai-jev-skill-suggestion/SKILL.md` and its mirror: the contract of a different
  consumer (skill suggestion), unaffected.
- `AGENTS.local.md` §29: the leader's use of the local scripts, whose markers are
  pinned by P3; `internal/cli/todo_auto.go` `:201/:231/:367` and
  `internal/cli/todo_jev_finding.go:5-6`: the display-only script line and the
  admission-path finding, both unchanged in contract;
  `internal/contract/kickoff/kickoff.go:21`; `internal/mission/*` and
  `internal/closure/*` receipt comments; the first amendment's tests.
- Hits that use "display-only" for a different object: hook `MessageDisplay`
  (`hooks-system.md:85`), the loop completion sentence (`loop.md`), `internal/mx`,
  `internal/web`.
- `.claude/skills/moai-kanban-foreman/SKILL.md:69` ("serial consumption in queue
  order is authorized"): states what the batch approval grants, not the cycle's pick
  order; observed, not changed.

### Out of Scope — the first exception's guards

- No edit to `TestJevAmendmentLinkage`, `activation_test.go`, `TestJevDoctrineAmendment`,
  `contract_mode_blocks_test.go`, `kickoff.go` or the first amendment's text,
  including the pinned guide paragraph that still says "one place" (§G R-4).

## §G Gaps and Residual Risks

- **R-1 — Interim state until the linked commit lands.** By design the split of §A.1
  persists until then; the guard, committed first, is inert (`false`, all markers
  absent) in that window.
- **R-2 — A full revert goes unseen.** If the single linked commit were reverted
  whole, the markers and the arming constant disappear together and the guard passes
  while the landed anchors again contradict the Jev side. `TestJevAmendmentLinkage`
  has the same property. Not mitigated here.
- **R-3 — Behavior claims are cited, not re-derived.** The amended text rests on
  REQ-TAP-001/-002/-003/-006/-010/-011. Eight behavior tests were re-run at this
  HEAD and PASS; the remainder of the 15-name ranking suite was not re-run.
- **R-4 — The pinned guide paragraph still reads "one place".** P3 forbids editing
  it, so X1 adds a following paragraph that scopes the old sentence to the Kickoff
  cross-check. A reader who stops after the old sentence is still misled; the full
  fix edits `SPEC-AUTONOMY-GATE-REWIRE-001/design.md` and the guide together and is
  not in scope.
- **R-5 — Literal check, not a semantic one.** The wording guard proves two literals
  share a passage; it cannot detect a sentence that generalizes the exception in
  other words (inherited from the predecessor's R-6).
- **R-6 — Extensions may be cut.** X3 costs a catalog-hash regeneration and X4 adds
  bytes to an always-loaded file; both are cut candidates and neither is required by
  a guard.
- **R-7 — Tool-level docs-site pages keep saying "display-only".** Accurate for the
  tool; a reader of the pages alone is not told the capability has an in-process
  exception (§D).
- **R-8 — The sweep is a pattern, not a proof.** Two regexes bound the inventory;
  restatements in other words would be missed (`research.md` §R1 records the two
  sweeps and the one hit the synonym sweep added).
