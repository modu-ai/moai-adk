---
id: SPEC-SYNC-SHA-SLOT-FORMAT-001
title: "sync_commit_sha slot format: one value grammar, one shared predicate, and a guard that survives the backfill window"
version: "0.3.0"
status: completed
created: 2026-08-29
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "lint, spec-lifecycle, sync, commit-sha, closer, backfill"
era: V3R6
tier: S
related_specs: [SPEC-MOVING-REF-GUARD-001, SPEC-V3R6-LIFECYCLE-REDESIGN-001]
---

# SPEC-SYNC-SHA-SLOT-FORMAT-001 — sync_commit_sha slot format

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.2.1 | 2026-08-29 | REQ-SSF-007's t357 contingency **resolved by observation, not deleted**. The t357 lane reports, through the lead, that t357 M2 sets `Severity: SeverityError` directly at its own rule's emission site (`REQ-AST-001-006`, guarded by `TestArtifactStatus_SurvivesEraDemotion`), with no global warning->error promotion, no `Advisory`, and no `eraDemotableCodes` entry of its own — so it promotes at neither layer the contingency table named as a hazard, this rule's `Finding.Severity` stays `warning`, and REQ-SSF-007 holds as written. The mechanism paragraph and the two-layer table are kept because they record what was at stake; the lane's "stop and report" instruction is marked historical. Recorded as a **reported** observation with its provenance — no t357 SPEC directory exists in this tree, so nothing was verified here. Mirrored into `plan.md` §C.4 and `acceptance.md` AC-SSF-010. §D.3 additionally records the contrast the operator asked to keep: t299 and t357 M2 reached opposite severities through the same count-first procedure (5 findings / 0 non-advisory -> `warning`; 392 day-one violations -> `error`, on the ground that the same commit removes all 392). No requirement, criterion, or measured figure changed. | manager-spec |
| 0.2.0 | 2026-08-29 | Plan audit iter-1 FAIL (0.80, `.moai/reports/t299/plan-audit-iter1.md`) remediated. **D1 (critical)**: §B.6 derived its enforcement cost from §B.2's twelve non-SHA values without passing them through REQ-SSF-005's own exemption — seven are recognized placeholders and are not findings. The corrected prediction is **5 findings, 0 non-advisory**, re-derived here with the auditor's classifier and named per file and line. The sharpest consequence is that the two `implemented` SPECs §B.6 had nominated as the non-advisory pair both hold `pending-backfill-sync` and are therefore **exempt**: the old prediction and the exemption cancelled exactly, which is why the error was invisible from inside the derivation. §D.3, AC-SSF-006, `plan.md` §C.3 and §F M4 all inherited the figure and are corrected with it. **D2**: §D.3's severity rationale named those two exempt SPECs as its exemplars and so refuted itself; rewritten against "0 non-advisory, 5 advisory, all on closed history", which strengthens the `warning` verdict rather than weakening it. **D3**: AC-SSF-010 added to decide REQ-SSF-007 mechanically, and REQ-SSF-007's justification is now stated as **conditional** on the rule's severity remaining `warning` at the `Finding` level, with the t357 contingency and the lane's response named. **D4**: the `mx_commit_sha` inheritance is measured (85 slots, 9 non-SHA, three of them deliberate declarations) and accepted in an explicit Out-of-Scope bullet, with the blast radius measured rather than asserted — all three owning SPECs are `completed`. **D5**: §B.3's command now excludes this SPEC's own directory, which was self-matching twice. **D6**: the annotation figure is restated with its correct denominator (105 of 346 values; 99 of the 334 conforming). D7/D8 recorded, no change; D9 is a pre-existing repository-wide inconsistency and stays recorded. A point the audit implied but did not state is now explicit in §A and §D.2: the lint rule will never flag the motivating t354 slot, because that slot is an admitted placeholder — the closer-side inversion is what repairs it. | manager-spec |
| 0.1.0 | 2026-08-29 | Initial plan-phase authoring for card t299, in worktree `.claude/worktrees/t299` at HEAD `a6bbbf82b`. Every figure in §B was measured in this tree; none was carried from the dispatch. Two dispatch claims were re-derived and one was **sharpened**: the root cause is not that `needsSHABackfill`'s allowlist is merely closed, but that the allowlist and the doctrine disagree — the `pending-backfill-*` spelling `spec-frontmatter-schema.md` §D3 itself prescribes is **absent** from that allowlist, so the closer's blind spot is aimed precisely at the sanctioned pattern (§B.3). A fourth independent value reader was found beyond the three named in the dispatch (§B.4). | manager-spec |

## §A. Problem Statement

`.moai/specs/<ID>/progress.md` carries a `sync_commit_sha:` slot whose contract is that it holds the
commit that carried the sync phase. Nothing enforces that contract. The slot accepts any text, and
the text that lands there when the real value is not yet knowable — a placeholder — is supposed to be
temporary. Some of it is not: prose sits in the slot permanently, and no signal is raised.

The permanence is the defect, not the placeholder. A commit cannot cite its own hash, so a
placeholder in the phase commit followed by a backfill in a later commit is the **documented and
sanctioned** pattern (`spec-frontmatter-schema.md` § SHA placeholder backfill exemption, "D3"). Any
guard that forbids the intermediate state forbids the only workable procedure.

What is missing is any mechanism that distinguishes *transiently placeholdered* from *permanently
prose-occupied*, on either side of the field:

- **Write side.** `internal/spec/closer.go` `needsSHABackfill` decides whether the closer repairs the
  slot from a **closed four-value allowlist**. Whether a SPEC gets repaired therefore depends on how
  its placeholder happens to be **spelled**. §B.3 shows the allowlist and the doctrine's own
  prescribed spelling do not intersect.
- **Read side.** No lint rule validates the value's shape at all. The field is read only as an
  opaque string, by four separate parsers that disagree about what counts as a value (§B.4).

### What this SPEC does, and what it deliberately does not

It fixes the **format contract and its two guards**: one value grammar, one shared predicate used by
both the closer and a new lint rule, and a warning that fires on a slot occupied by something that is
neither a commit SHA nor the canonical backfill placeholder.

**Which guard handles the motivating case, stated plainly so a reader does not conclude the card
missed it.** The motivating instance — `SPEC-BACKLOG-LOCK-BUDGET-001` holding `pending-backfill-sync`
forever — is repaired by the **closer-side inversion (REQ-SSF-003)**, and the lint rule will **never**
flag that slot. It cannot: `pending-backfill-sync` is a recognized placeholder, and REQ-SSF-005
requires silence on exactly that value. The two guards answer different questions and the division is
deliberate:

| Surface | Question | Motivating case |
|---|---|---|
| closer (REQ-SSF-003) | is this slot still owed a real SHA? | **yes — this is what repairs it** |
| lint (REQ-SSF-004) | is this slot occupied by something that is neither a SHA nor a sanctioned placeholder? | silent, correctly |

The lint rule is a **backstop** for the class the closer cannot help — a value that is neither a SHA
nor a placeholder, so no close will ever be triggered to repair it and no reader can tell what the
slot was supposed to say. §B.6 measures that class at five slots.

It does **not** repair the corpus. The twelve non-conforming values measured in §B.2 are left as
they are, and §G records why each class is somebody else's decision.

## §B. Measured Baseline

Every figure below was measured in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t299`, branch
`WT-sha-slot-format`, HEAD `a6bbbf82b`, on 2026-08-29. Each row names the command that produced it.
`a6bbbf82b` is `origin/develop` at measurement time; it is recorded here as a **frozen literal**, and
every count below is decided against that literal rather than against the moving ref.

### B.1 — The tree this was measured in

```
git rev-parse --show-toplevel  → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t299
git branch --show-current      → WT-sha-slot-format
git rev-parse --short HEAD     → a6bbbf82b
```

### B.2 — The corpus

| # | Measurement | Count | Command |
|---|---|---|---|
| 1 | `sync_commit_sha:` lines in `.moai/specs/*/progress.md` | **346** | `grep -h '^sync_commit_sha:' .moai/specs/*/progress.md \| wc -l` |
| 2 | (1) whose value token is a 7-40 hex SHA (quoted or bare) | **334** | (1) piped through `grep -cE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]\|$)'` after stripping the key |
| 3 | (1) whose value token is **not** a SHA | **12** | (1) piped through `grep -vE …` (same pattern, inverted) |
| 4 | (1) carrying a trailing annotation of any form | **105** | (1) piped through `grep -c '#'` after stripping the key |

Row 4 is why the grammar in §D.1 must admit a trailing annotation: **105 of the 346 values** carry
one, and **99 of the 334 conforming ones** do (the difference is six annotated non-SHA values, which
row 4's `#` count does not separate). A grammar rejecting annotated values would therefore flag
**99 of 334 — nearly a third — of a healthy corpus.** Three further values carry a non-`#` tail (a
parenthetical or an em-dash clause) and are likewise not defects.

```
… | grep -E '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)' | grep -c '#'   → 99
```

### B.3 — The write-side blind spot is aimed at the sanctioned spelling

`needsSHABackfill` (`internal/spec/closer.go:397`) returns `true` for exactly four values:

```go
switch v { case "", "(this commit)", "(pending)", "<pending>": return true }
```

`spec-frontmatter-schema.md` §D3 prescribes the placeholder as **`pending-backfill-*`**. That family
is **not in the allowlist**. The closer therefore does not repair the very pattern the doctrine tells
authors to write.

The effect is masked, and the masking is itself instructive. `state.SyncCommitSHA` is populated by
`extractProgressField` → `cleanFieldValue` (`internal/spec/era.go:252`), which owns a **different**
placeholder set — `"" / null / none / tbd / <pending> / pending` — and maps those to the empty
string. So `null` and a bare `pending` do get repaired, by accident, through a normalizer written for
era classification. `pending-backfill-sync` is in neither set and is repaired by neither path.

Family spellings actually present across `progress.md` files:

```
grep -rH 'pending-backfill' .moai/specs/*/progress.md \
  | grep -v 'SPEC-SYNC-SHA-SLOT-FORMAT-001' \
  | grep -oE 'pending-backfill[a-zA-Z0-9-]*' | sort | uniq -c
  → 28 distinct spellings, led by `pending-backfill` (29) and `pending-backfill-sync` (24)
```

**The exclusion in the second line is load-bearing for reproducibility.** This SPEC's own
`progress.md` §E.1 quotes the measurement command verbatim, and the naive glob matches that file,
adding two self-matches: without the exclusion the same command now returns `31`, and a reader would
get a figure differing from the one recorded here with no explanation. The distinct-spelling count
(28) and `pending-backfill-sync` (24) are unaffected either way.

Twenty-eight spellings is what an unenforced format produces. It is also the direct measurement of
why an enumerated allowlist cannot be the closer's test: the set is open, and every new suffix an
author invents silently opts that SPEC out of repair.

### B.4 — Four readers, three notions of "a value"

| Reader | Location | What it treats as the value |
|---|---|---|
| era classification | `era.go:139` via `cleanFieldValue` | quotes stripped; `null`/`none`/`tbd`/`pending`/`<pending>` → empty |
| lifecycle drift audit | `audit.go:360` via `cleanFieldValue` | same as above |
| close / backfill | `closer.go:618` + `needsSHABackfill` | same normalization, then a **different** four-value allowlist |
| Epic status render | `status.go:261` `syncShaYAMLPattern` | `^\s*sync_commit_sha\s*:\s*"?(.+?)"?\s*$` — its own regex; captures the whole tail |

Observed consequence of the fourth reader, from `moai spec audit --filter-spec SPEC-V3R6-AUDIT-MODEL-PIN-001`
run in this tree: the emitted `sync_commit_sha` detail is

```
pending-backfill-sync"   # D3 self-reference exemption — a commit cannot know its own SHA; …
```

— a stray closing quote and a full prose sentence, because `cleanFieldValue` trims quotes from the
ends of the **whole string** and the string ends in the annotation. The drift report is currently
publishing an unparsed line as if it were a SHA.

### B.5 — The twelve, and where they sit

The twelve non-conforming values live in **eleven** SPECs (`SPEC-V3R6-SESSION-LEGACY-COVERAGE-001`
holds two). Their `spec.md` statuses, measured by `grep -m1 '^status:'` on each:

| status | SPECs | violations |
|---|---|---|
| `completed` | 9 | 10 |
| `implemented` | 2 — `SPEC-BACKLOG-LOCK-BUDGET-001`, `SPEC-V3R6-AUDIT-MODEL-PIN-001` | 2 |

Both `implemented` SPECs classify **era V3R6**, verified independently rather than assumed —
`mcp__moai__spec_audit` with `project_root` set to this worktree returns
`"era":"V3R6","heuristic_matched":"H-4"` for each. Neither is grandfathered.

### B.5.1 — Twelve non-SHA values are not twelve findings

The twelve of §B.2 are the values that are **not a SHA**. REQ-SSF-005 exempts a recognized backfill
placeholder, so the flagged set is the twelve **minus** the placeholders. Classifying all 346 slots
with a classifier implementing §D.1 exactly (`python3 .moai/reports/t299/grammar_check.py`, the plan
auditor's, re-run in this tree):

```
total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
```

**Seven of the twelve are recognized placeholders and are correctly silent:**
`SPEC-AUDIT-SNAPSHOT-001` (`pending-backfill-SYNC`), `SPEC-BACKLOG-LOCK-BUDGET-001`,
`SPEC-INFINITE-GOAL-001`, `SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-V3R6-AUDIT-MODEL-PIN-001`,
`SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001`, `SPEC-VERIFICATION-COMPLETENESS-001`.

**Five are findings**, and each is named with its line:

| File | Line | Value token |
|---|---|---|
| `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md` | 47 | `null` |
| `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md` | 259 | `<pending>` |
| `.moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/progress.md` | 103 | `TBD` |
| `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/progress.md` | 45 | `null` |
| `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/progress.md` | 149 | `pending` |

All five owning `spec.md` files carry `status: completed` (`grep -m1 -H '^status:'` over the four
SPECs).

**Both `implemented` SPECs of §B.5 are in the exempt seven, not the flagged five.** Each holds
`pending-backfill-sync`. This is stated explicitly because v0.1.0 of this SPEC predicted them as the
*only* two non-advisory findings: the prediction and the exemption cancel exactly, which is what made
the arithmetic error invisible from inside the derivation that produced it.

### B.6 — Predicted enforcement cost, and the derivation that decides it

`internal/spec/lint.go:220` computes `demote := isGrandfatheredSpecDir(…) || terminalStatusEnum[doc.Frontmatter.Status]`,
and `applyEraDemotion` sets `Advisory = true` on **every** warning of a demoted document.
`Report.HasErrors` (line 61) escalates under `--strict` only for a warning that is **not** advisory.
`terminalStatusEnum` (line 1134) contains `completed`.

The attachment question the dispatch flagged is real and is answered here by reading the loop, not by
assuming: `demote` is computed **per document** from `spec.md`'s frontmatter and applied to
`docFindings`, the batch of every finding that document's rules produced. A rule that reads a sibling
`progress.md` (the `MovingRefUnpinnedRule` precedent, `lint_movingref.go:129`) sets `Finding.File` to
the sibling path but its finding is still in that batch. `progress.md` carrying no frontmatter status
is therefore not an escape: the demotion keys on the owning `spec.md`, and the finding is demoted with
it.

**Predicted: 5 total findings, 0 non-advisory.** The flagged set is the five of §B.5.1, every one of
them in a `completed` SPEC, and `completed` is in `terminalStatusEnum` — so all five are demoted to
advisory and `--strict` escalates none of them. This SPEC's rule contributes **nothing** to the strict
exit status on the corpus as it stands.

This is a prediction derived from code plus a classifier run over the corpus, not a measurement of a
rule that does not yet exist. AC-SSF-006 converts it into a measurement, and a divergence from
**5 / 0** is a finding to report rather than a number to adjust.

The contrast with `SPEC-MOVING-REF-GUARD-001` is the reason this SPEC can afford `warning` severity
at all: that rule faced 42 candidate lines with no terminal-status shelter, and reddening them would
have made bulk suppression the rational response. Here the shelter covers the entire flagged set.

## §C. Requirements

**REQ-SSF-001** — The system shall define a single value grammar for the `sync_commit_sha` slot,
consisting of a **value token** optionally preceded by a quote, optionally followed by a trailing
annotation, where the value token is either a commit SHA or the canonical backfill placeholder.

**REQ-SSF-002** — The system shall expose one shared predicate, in `internal/spec`, that decides
whether a value token is a commit SHA, and both the closer's backfill decision and the lint rule's
finding decision shall be derived from that single predicate.

**REQ-SSF-003** — When the closer evaluates a `sync_commit_sha` value whose value token is not a
commit SHA, the closer shall treat the slot as requiring backfill, **irrespective of the token's
spelling**.

**REQ-SSF-004** — When a `sync_commit_sha` slot holds a value token that is neither a commit SHA nor
a recognized backfill placeholder, the lint shall emit a `SyncSHASlotFormat` finding at `warning`
severity naming the file, the line, and the canonical placeholder spelling.

**REQ-SSF-005** — While a slot holds a recognized backfill placeholder, the lint shall **not** emit a
finding for that slot — the D3 backfill window is a sanctioned intermediate state, not a defect.

**REQ-SSF-006** — When a `sync_commit_sha` value carries a trailing annotation after its value token,
the annotation shall not affect either decision.

**REQ-SSF-007** — The `SyncSHASlotFormat` code shall **not** be added to `eraDemotableCodes`; the
existing warning-demotion path already covers grandfathered and terminal-status SPECs, and
`eraDemotableCodes` governs the demotion of **errors** only.

> **[HARD] This requirement's justification is conditional, and the condition is named so a later
> reader can check it rather than inherit it.** It holds **only while the rule's severity is
> `warning` at the `Finding` level.** The mechanism: `lint.go:284` gates on
> `f.Severity == SeverityError && eraDemotableCodes[f.Code]`, and a warning is already made advisory
> by the branch at `lint.go:288` — so for a warning the map entry is inert, and an inert entry in a
> policy map reads as intent to a later maintainer.
>
> **If the findings ever become `SeverityError`, this requirement inverts.** `eraDemotableCodes` would
> then be the *only* remaining shelter, and forbidding it would turn the five findings on closed
> history (§B.5.1) into hard `--strict` errors — precisely the outcome §D.3 says the rule exists to
> avoid, and the one that makes `lint.skip` the rational response.
>
> **The named contingency: card t357 M2.** `plan.md` §C.4 records, from the lead's dispatch and
> **not** independently verified (no t357 SPEC directory exists in this tree to read), that t357 M2
> raises `--strict` behavior by promoting warnings to errors. Which layer it promotes at decides
> whether this requirement survives:
>
> | t357 promotes at | Consequence here |
> |---|---|
> | `Report.HasErrors` (escalation-time) | `Finding.Severity` stays `warning`; the entry stays inert; REQ-SSF-007 holds unchanged |
> | `Finding.Severity` (finding-time) | the demotion path no longer shelters these findings; REQ-SSF-007's premise is void |
>
> **RESOLVED by observation on 2026-08-29 — the contingency did not materialize. It is kept here
> rather than deleted so a later reader can see what was at stake and re-check it.** The t357 lane
> reports, through the lead, that card t357 M2 sets `Severity: SeverityError` **directly at its own
> rule's emission site** (`REQ-AST-001-006`, guarded by `TestArtifactStatus_SurvivesEraDemotion`). It
> introduces no global `SeverityWarning` -> error promotion, sets no `Advisory`, and adds no code of
> its own to `eraDemotableCodes`. t357 therefore promotes at **neither** of the two layers the table
> above anticipates as a hazard: this rule's `Finding.Severity` stays `warning`, the warning-demotion
> path still shelters it, and an `eraDemotableCodes` entry for `SyncSHASlotFormat` would still be
> inert. **REQ-SSF-007's premise holds, and the requirement stands as written.**
>
> **This is a reported observation, not one this lane measured.** No t357 SPEC directory exists in
> this tree, and nothing above was read from `internal/spec` here; it is attributed to the t357 lane
> via the lead and recorded with that provenance so a later reader re-verifies it against t357's
> landed diff rather than inheriting it. The resolution **strengthens** AC-SSF-010's premise; it does
> not relax the criterion, which still requires the code's absence from the map.
>
> **Historical — the response that would have applied.** Had t357 promoted at `Finding.Severity`, the
> lane's instruction was to read which layer it promotes at and **stop and report to the lead** rather
> than adding the map entry on its own authority: adding it would have satisfied REQ-SSF-007's inverse
> without the requirement having been changed, and the choice between "shelter these five" and "keep
> the map errors-only" is a policy decision the operator owns. That response is no longer live.
> AC-SSF-010 decides the requirement as written; it never decided the contingency.

**REQ-SSF-008** — The system shall not alter `cleanFieldValue`, whose `(this commit)`-as-value
behavior the era-classification heuristics depend on.

## §D. Design Decisions

### D.1 — The value grammar

```
line   := "sync_commit_sha:" WS* value
value  := QUOTE? token QUOTE? ( WS+ annotation )?
token  := SHA | PLACEHOLDER
SHA    := [0-9a-fA-F]{7,40}
PLACEHOLDER := "pending-backfill" ( "-" [A-Za-z0-9-]+ )?
```

**The token is the first whitespace-delimited run.** Everything after the first whitespace run is a
free-form annotation and is ignored by both decisions. This is chosen over enumerating separators
(`#`, ` — `, ` (…)`) because all three forms occur in the corpus (§B.2 row 4 plus three non-`#`
tails) and a separator enumeration would have to grow every time an author picks a fourth. A
first-token rule admits all of them without naming any.

**The canonical placeholder is `pending-backfill`.** Three reasons, in order of weight: §D3 of the
schema already prescribes the `pending-backfill-*` family, so this spelling is not a new invention
but the existing doctrine written down precisely; it is the single most-used spelling in the corpus
(29 occurrences, §B.3); and the suffixed forms stay admitted, so the twenty-eight existing spellings
do not become findings on the first run.

**Quote handling is single-leading/single-trailing on the token, not on the line.** This is the
narrow repair of the §B.4 artifact: `"pending-backfill-sync"   # …` yields the token
`pending-backfill-sync`, where the current whole-string trim yields `pending-backfill-sync"   # …`.

**Known limitation (L1): a 7-40 character all-hex English word parses as a SHA.** `defaced` is the
worked example. The class is accepted rather than defended against: the alternative is a dictionary,
and the false negative costs one unflagged slot while a dictionary would cost a maintenance surface.

**Known limitation (L2): the rule reads shape, never reachability.** It cannot tell a SHA that names
a real commit from seven plausible hex characters. Reachability verification is a separate concern and
is not claimed here.

### D.2 — Enforcement on both sides, from one predicate

Both surfaces change, and they share `isCommitSHAToken`. The dispatch asked for a justification
against the observed drift; §B.4 is that justification measured rather than argued — four readers
already exist and three notions of "a value" already exist among them, produced by exactly the
practice of writing a fresh test at each site. A fifth independent test is the same defect again.

**The closer's test is inverted, and the inversion is the substance of the fix.**
`needsSHABackfill` becomes `!isCommitSHAToken(token(value))` — a **positive, closed** test for what a
SHA is, rather than an **open** enumeration of what a placeholder might be spelled as. The set of
non-SHAs needs no maintenance; the set of placeholder spellings provably does (28 and counting).

**On the previously-caught set this is a strict widening**, and AC-SSF-005 measures it: `""`,
`(this commit)`, `(pending)`, `<pending>` all fail `isCommitSHAToken`, so nothing that was repaired
stops being repaired.

**On the previously-uncaught set it is not automatically benign, and for `mx_commit_sha` it is not.**
`needsSHABackfill` is shared with `closer.go:332`, which on a `true` writes the literal
`l60MxBackfillPlaceholder = "(this commit)"` over whatever the slot held. Measured:

```
grep -h '^mx_commit_sha:' .moai/specs/*/progress.md | wc -l                                     → 85
… | sed 's/^mx_commit_sha:[[:space:]]*//' | grep -vcE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'   → 9
```

Six of the nine (`null` ×3, `(this commit)` ×3) are repaired today and are unaffected. **Three are
deliberate declarations that the inversion newly brings into scope**, and a later close would
overwrite each with `(this commit)`:

| Value | Owning SPEC | `spec.md` status |
|---|---|---|
| `<NA>` | `SPEC-CLIFIX-CONCURRENCY-001` | `completed` |
| `_<pending Mx-phase>_` | `SPEC-V3R6-BASH-RISK-GOVERNANCE-001` | `completed` |
| `_(not applicable — this SPEC removes the Mx-phase concept; REQ-LR-004/007)_` | `SPEC-V3R6-LIFECYCLE-REDESIGN-001` | `completed` |

Overwriting the third would replace a reasoned declaration with a placeholder. The exposure is
**prospective, not live**: all three owning SPECs are already `completed`, so the close path has no
reason to run against them again — the blast radius today is measured at zero, and what remains is a
future SPEC writing a declaration of this shape and then closing. §E accepts that behavior explicitly
rather than leaving "inherits the widening" as the whole account of a write path.

The third value is also the reason the Mx concept is retired: `SPEC-V3R6-LIFECYCLE-REDESIGN-001` is
the SPEC that removed the phase, so the population that could grow this class is shrinking rather
than growing.

### D.3 — Severity: `warning`, and not in `eraDemotableCodes`

`warning`, on the §B.6 derivation as corrected: **0 non-advisory findings, 5 advisory, all five on
closed history.** No open card is put into the strict path by this rule, and no lane inherits work
from it on day one.

`error` would invert exactly that. The five findings sit in `completed` SPECs, and
`applyEraDemotion` shelters a warning unconditionally but demotes an **error** only when its code is
in `eraDemotableCodes` (`lint.go:284`) — which REQ-SSF-007 declines. An `error` severity would
therefore put five closed SPECs' history into the strict path with no shelter at all, and
`lint.skip` becomes the rational response: the outcome the rule exists to prevent.

Note what this rationale does **not** claim, because v0.1.0 did claim it and it was false: it does not
rest on the two `implemented` SPECs. Both hold `pending-backfill-sync` and are exempt under
REQ-SSF-005 (§B.5.1). A slot that is genuinely mid-backfill is not flagged at all, so it cannot be an
exemplar of enforcement cost.

`eraDemotableCodes` is deliberately not touched. That map demotes **errors** to warnings on protected
SPECs (`lint.go:284`); a warning is already marked advisory by the `case f.Severity == SeverityWarning`
branch immediately below it. Adding the code would be inert, and inert entries in a policy map read as
intent. This follows `MovingRefUnpinnedRule`, which records the same choice for the same reason.

**A contrast worth keeping: card t357 M2 reached the opposite severity through the same procedure.**
Both cards chose a severity by first counting the day-one violations and only then picking the
severity that count could carry. This card measured **5 findings / 0 non-advisory** and chose
`warning`. t357 M2 — reported through the lead, **not** verified from this tree — measured **392
day-one violations** and chose `error`. The two verdicts are not in tension, and t357's ground is
**not** "no violations exist": it is that the **same commit** removes all 392, so its rule lands on a
corpus already clean. That ground is load-bearing in exactly the way this card's is — if the cleanup
were dropped from that commit, t357's severity decision would have to be revisited immediately, just
as `warning` here would have to be revisited if the five findings sat on open cards rather than closed
history. What a later reader should carry from the comparison is the procedure, not either verdict:
count the day-one cost first, then choose the severity that cost can bear.

### D.4 — t354 coordination

`SPEC-BACKLOG-LOCK-BUDGET-001` is card t354's SPEC. It is `status: implemented` with AC-BLB-006 FAIL,
so it is an **open card**, and its own close is what must populate its slot. Reaching into it from
this card would repair a value whose owner is mid-work and would remove the very condition their close
procedure is meant to satisfy.

The recommendation carried to the operator: **t299 defines the format and lands the guards; t354
backfills its own value at its own close.** No contrary evidence was found — `moai spec audit` on that
SPEC reports the drift with the remediation `moai spec close SPEC-BACKLOG-LOCK-BUDGET-001 --backfill-only`,
addressed to that card.

## §E. Scope Boundaries

### Out of Scope — corpus repair

- The twelve non-conforming values measured in §B.2 are **not** rewritten by this card. The guard is
  forward-looking; the existing values are observed and recorded, not repaired.
- The ten slots in `completed` SPECs are **observed-but-not-repaired**. Whether closed history is
  rewritten to satisfy a rule written afterwards is an operator decision, not a lane's, and it is the
  same judgment `applyEraDemotion` already encodes by demoting terminal-status findings.

### Out of Scope — another open card's SPEC

- `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354) is **not** touched. Its slot stays its own card's to
  backfill, per §D.4. Editing another open card's SPEC directory is a scope violation regardless of
  how mechanical the edit looks.
- `SPEC-V3R6-AUDIT-MODEL-PIN-001` is likewise left alone; it is `implemented` and its close owns its
  slot.

### Out of Scope — adjacent machinery

- `cleanFieldValue` and the era-classification heuristics are **not** altered (REQ-SSF-008). The
  `(this commit)`-as-value behavior is load-bearing for H-3/H-4 and changing it would move era
  classification for an unmeasured set of SPECs.
- The `mx_commit_sha` field is not in scope, and **no lint rule is written for it**. It shares
  `needsSHABackfill`, so it inherits the closer-side widening; the inheritance is measured in §D.2
  rather than assumed.
- **Three `mx_commit_sha` declarations are knowingly brought into the backfill predicate's scope, and
  the behavior is accepted rather than guarded**: `<NA>` (`SPEC-CLIFIX-CONCURRENCY-001`),
  `_<pending Mx-phase>_` (`SPEC-V3R6-BASH-RISK-GOVERNANCE-001`), and
  `_(not applicable — this SPEC removes the Mx-phase concept; REQ-LR-004/007)_`
  (`SPEC-V3R6-LIFECYCLE-REDESIGN-001`). After the inversion a subsequent close would overwrite each
  with `(this commit)`. This is accepted on a measured basis, not waved through: all three owning
  SPECs are `status: completed`, so no close is expected against them and the live blast radius is
  **zero**; the Mx phase itself is retired by the third SPEC in that list, so the class shrinks rather
  than grows. Should a future SPEC need a durable "not applicable" declaration in this field, that is
  a `mx`-path requirement to raise as its own card — this one does not pre-empt it.
- The fourth reader (`internal/epic/status.go` `syncShaYAMLPattern`, §B.4) is **not** unified into the
  shared predicate by this card. It is recorded as a follow-up candidate: it is a render path, not a
  decision path, and folding it in would widen the diff past the format contract.
- Reachability verification of SHA values (does the commit exist, is it an ancestor) is not in scope.

## §F. Dependencies and Cross-References

- `.claude/rules/moai/development/spec-frontmatter-schema.md` § SHA placeholder backfill exemption (D3)
  — the doctrine this grammar writes down precisely. Local copy and
  `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` are
  **byte-identical** at `a6bbbf82b` (`diff` → exit 0), so a doc edit obliges both plus `make build`.
- `SPEC-MOVING-REF-GUARD-001` — the file-layout, severity, and `eraDemotableCodes` precedent.
- `internal/spec/lint.go:220` `applyEraDemotion` — the demotion path §B.6's prediction is derived from.
