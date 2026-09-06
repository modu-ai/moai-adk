# SPEC-JUDGMENT-FIRST-MODE-001 — Implementation Plan

Milestones are ordered by **decision reversibility**: the decisions most likely to change on
review come first (the user-facing output convention and the Frozen-clause conditionalization),
and the mechanical distribution work is last. Review attention should concentrate on M1 and M2.

**One milestone breaks that ordering deliberately: M0.** The observer is mechanical and
precedent-shaped — by reversibility it belongs late, and in the 0.1.0 plan it sat at M4. It moved
to the front because of a measurable ordering consequence surfaced by the iteration-1 audit (D4):
the vacuity falsifier is read from rows the observer wrote, and a falsifier is worthless unless the
detector that wrote those rows has been observed **firing**. The only window in which a live
`label_present: true` row can be recorded is one where the repository is still at
`recommendation_mode: push` and the doctrine is still unamended — that is, **before** M1. Build the
observer after the convention lands and that window no longer exists at any point in the plan. M0
is therefore not "the easy part first"; it is the only position from which REQ-JFM-024's
pre-landing baseline is obtainable at all.

## §A Context

Card t401. Issue `modu-ai/moai-adk#1683` item #2 ("Analysis is Pull, Not Push"), adopted first and
narrowly. Base commit `ad272be20`, worktree `.claude/worktrees/t401`, branch `WT-analysis-pull`.
Full framing in `spec.md` §A; evidence under `.moai/reports/t401/`.

## §B Known Issues Entering the Milestones

1. **Two `[HARD]` clauses will contradict each other unless M1 handles both.**
   `askuser-protocol.md:64` and `.claude/skills/moai/workflows/run.md:137` both mandate a
   `(Recommended)`-first option. Conditioning only the first leaves the second as a live
   contradiction at the one mandatory human gate. `run.md` is not one of the six nominated
   surfaces — it is a downstream consumer of S1 — so its one-line mode reference is **flagged for
   operator review** rather than assumed (spec.md §E.1).
2. **S6 has no inherited guard.** `context-window-management.md:84` sits outside the
   AskUserQuestion channel, so the mode must be stated locally there. It cannot be covered by a
   cross-reference to `askuser-protocol.md` alone.
3. **The banner surfaces are localized.** `moai.md` §8 carries 4-locale header translation tables.
   A withholding branch that only exists in the English column is a partial landing.
4. **`.claude/settings.json` is rendered from `settings.json.tmpl`.** Editing the rendered file
   without its `.tmpl` sibling is reverted by the next `moai update`.
5. **The observer's question-type field is not free — and the SPEC no longer depends on it.**
   `AskUserQuestion` payloads carry no "question type" field; "decision-type, options derived from
   investigation" is a doctrine concept. Rather than have AC-JFM-018 filter on a field that does
   not exist, the requirement itself was widened: REQ-JFM-005 binds every `AskUserQuestion` call
   under `pull`, and AC-JFM-018's denominator is **every `mode == "pull"` row, unfiltered**
   (design.md §6.3 is the single rule; this file and `acceptance.md` follow it). M0 still records
   `question_type` descriptively where it can honestly derive one, but nothing selects on it.
6. **A falsifier read from an unproven detector asserts nothing.** `violations == 0` is satisfied
   by an observer whose label detector never fires. REQ-JFM-024 therefore requires both directions
   of the detector to be shown — a unit assertion on constructed payloads, and a live pre-landing
   `push`-mode row with `label_present: true`. This is what forces M0 to the front.

## §C Pre-flight

Run before M1, in one batch:

```bash
git rev-parse --short HEAD && git branch --show-current
git status --porcelain
go build ./... && go vet ./internal/config/... ./internal/hook/...
grep -c 'Adaptive strength' .claude/rules/moai/core/askuser-protocol.md
grep -n 'CONST-V3R5-035' -A6 .claude/rules/moai/core/zone-registry.md
```

Expected: HEAD on `WT-analysis-pull`, clean tree apart from the SPEC artifacts, build and vet
green, one Adaptive-strength occurrence, the registry entry present with `canary_gate: true`.

## §D Constraints

Carried verbatim from `spec.md` §D (CONST-1 … CONST-7). Two bind every milestone:

- **Template-First**: the mirror lands in the **same milestone** as the local file, never later.
- **Neutrality**: template-side text is authored neutral from the start — no SPEC ID, no REQ
  token, no date, no SHA, no `/Users/` path, no `CLAUDE.local.md` reference.

## §E Self-Verification

Each milestone below names its own verification command. In addition, at the end of every
milestone:

```bash
go build ./...
git status --porcelain
```

and at the end of M5 only, the affected-package suite (`internal/config`, `internal/hook`,
`internal/template`). The full suite is CI's judgment, not the lane's.

## §F Milestones

### M0 — The runtime observer and the pre-landing baseline window (Priority: High; must precede M1)

Sibling of the existing `Agent|Task` (agent-model) and `SendMessage` (stop-guard) observation
branches in `internal/hook/pre_tool.go`. Observation only — no deny path in either mode.

| File | Change | Mirror |
|---|---|---|
| `internal/hook/askuser_observer.go` (new) | Parse `input.ToolInput`, detect a `(권장)` / `(Recommended)` suffix on any option label, resolve the mode, append one JSONL row under `.moai/logs/` | — |
| `internal/hook/pre_tool.go` | Add the `input.ToolName == "AskUserQuestion"` observation branch after the `SendMessage` branch; it returns no deny | — |
| `internal/hook/askuser_observer_test.go` (new) | Row shape; **both** detector directions (a payload carrying `(권장)` / `(Recommended)` ⇒ `label_present: true`, one without ⇒ `false`); never-denies table; fail-open table (unparseable payload / absent config / unwritable path) | — |
| `.claude/settings.json` | Fourth `PreToolUse` matcher block `AskUserQuestion`, routing to `handle-pre-tool.sh`, in the shape of the existing `SendMessage\|TaskStop` block | `internal/template/templates/.claude/settings.json.tmpl` |

**The baseline window is the point of the milestone, not a by-product.** With the observer live and
this repository still at the default `recommendation_mode: push` — M3 has not run, M1/M2 have not
amended the doctrine — collect a `push`-mode control sample of at least 5 rows containing **at
least one** row with `label_present: true`. That row is the live evidence that the detector fires,
recorded before the convention exists to suppress it. Copy the window's rows to
`.moai/reports/t401/baseline-push-window.jsonl` and cite that path in `progress.md` §E.2; the live
log is not a durable artifact.

**The party that collects the window is the session that actually asks — and whose observer is
actually wired.** Under the kanban division of labour the asking party is the **lead** session,
which owns the operator channel — not the card's lane: a lane never prompts the operator, so it
issues no `AskUserQuestion` calls and can produce no rows at all. But asking is necessary, not
sufficient. The rows exist only where the observer is live **in that session**: a session that was
already running when the `AskUserQuestion` matcher was added to the settings does not pick the
matcher up, so it asks and records nothing. The collecting session is therefore one started
**after** the matcher landed. Rows accumulate under the asking session's `CLAUDE_PROJECT_DIR` (its
cwd when that variable is unset), at `.moai/logs/askuser-observations.jsonl`, which is generally
**not** the card worktree — hence the copy step above, and hence the provenance record that
accompanies it (§D.6, REQ-JFM-025). Read an empty log in a session that demonstrably asked as
evidence of an **unwired observer**, never as evidence that no labelled option was offered: those
are a real absence and a vacuous green, and confusing them is the reading error this milestone
exists to prevent. Manufacturing questions in order to fill the window is prohibited: a window
filled that way is decoration, not a sample, and it is the very shape of vacuous evidence this SPEC
exists to remove.

**Do not proceed to M1 until the baseline row exists.** Opening the `pull` window (M6) without it
leaves the falsifier unable to distinguish a followed convention from an inert detector — the
mutant D4 named.

The `question_type` decision is settled here: the observer records what the payload actually
contains. Where no defensible classification exists the field is omitted. Nothing selects on it
either way (§B item 5), so an honest omission costs the falsifier nothing.

Verify:

```bash
go test ./internal/hook/... -run 'AskUserQuestionObserver' -count=1 -v
go test ./internal/hook/... -count=1
python3 -c "import json;print([m.get('matcher') for m in json.load(open('.claude/settings.json'))['hooks']['PreToolUse']])"
jq -s '{n: length, positives: ([.[] | select(.label_present==true)] | length)}' .moai/reports/t401/baseline-push-window.jsonl
```

Covers: REQ-JFM-017, 018, 021, 024, 025. AC-JFM-014, 015, 016, 023.

### M1 — The pull convention in doctrine (Priority: High; highest change likelihood)

Generalize the existing Adaptive-strength clause into the mode axis, and conditionalize the Frozen
clause without rewriting it.

| File | Change | Mirror |
|---|---|---|
| `.claude/rules/moai/core/askuser-protocol.md` | Generalize § Recommendation Placement Principles principle 5 into the two-value mode; add the pull-mode withholding rule (S1), the on-request emission rule, and the three adopted conditions | `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` |
| `.claude/rules/moai/development/branch-origin-protocol.md` | Keep the current `(권장)`-first text as the push-mode branch; add the pull-mode branch beside it | `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` |
| `.claude/rules/moai/core/zone-registry.md` | **No change to the `CONST-V3R5-035` `clause:` string.** Touch only if an annotation is required, and never the clause text | `internal/template/templates/.claude/rules/moai/core/zone-registry.md` |
| `.claude/skills/moai/workflows/run.md` | One-line mode reference on the `(Recommended)`-first Kickoff clause (§B item 1 — operator-review flagged) | `internal/template/templates/.claude/skills/moai/workflows/run.md` |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | One-line mode reference on **both** `(권장)`-first clauses: the Implementation Kickoff Approval clause (`:212` — the same clause as `run.md:137`, one file over) and the Phase 13 BODP-gate first-option rule (`:353` — the sole implementing site of Frozen `CONST-V3R5-035`, whose doctrine site `branch-origin-protocol.md:25` this milestone already conditions). Operator-review flagged, as `run.md:137` is | `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` |

Verify:

```bash
grep -n 'recommendation_mode' .claude/rules/moai/core/askuser-protocol.md
git diff ad272be20 -- .claude/rules/moai/core/zone-registry.md          # expect: no clause: change
git diff ad272be20 -- .claude/rules/moai/workflow/orchestration-mode-selection.md   # expect: empty
diff <(sed -n '/Report-Before-Ask/,/^## /p' .claude/rules/moai/core/askuser-protocol.md) \
     <(git show ad272be20:.claude/rules/moai/core/askuser-protocol.md | sed -n '/Report-Before-Ask/,/^## /p')
```

**The AC-JFM-013 sweep, with its window defined.** The 0.1.0 criterion turned on "within the
enclosing clause" and never said what a clause is, leaving the sweep unbounded. 0.2.2 fixed the
window at **the matched physical line** and justified it with the claim that "every candidate this
tree actually contains is a single-line paragraph or list item (measured)". **That claim was
false.** The run phase falsified it: `plan/spec-assembly.md:212` is a continuation line of a
wrapped `[HARD]` paragraph spanning roughly `:208`-`:214`, so the line window cut a clause in half
— and the only route to a PASS was rewriting the paragraph as one long line so the token would land
inside the measuring instrument. A criterion whose only pass route is writing to fit the instrument
proves nothing.

The window is therefore **the enclosing markdown block of the matched line**, with the boundary
rule stated mechanically in acceptance.md AC-JFM-013 (blank line / heading / code fence / list
marker / table row). The window is a superset of the line, so it only widens what counts as
conditioning; it never narrows it. The long line the run phase wrote at `spec-assembly.md:212`
stays as it is — it is the evidence for this repair, not a defect to tidy. The candidate set is
mechanical:

```bash
grep -rn -E '\(Recommended\)|\(권장\)' .claude/rules .claude/skills .claude/output-styles \
  | grep -iE '\bfirst\b|첫 |먼저' > .moai/reports/t401/ac013-candidates.txt
wc -l < .moai/reports/t401/ac013-candidates.txt          # swept count — MUST be > 0
grep -c 'recommendation_mode' .moai/reports/t401/ac013-candidates.txt
```

Requiring a mode reference **inside every candidate's own window** would be wrong —
`zone-registry.md:869` is the Frozen `clause:` string that REQ-JFM-015 / spec.md §B.2 forbids
editing, and its conditionalization lives in `branch-origin-protocol.md` instead. So the criterion
is a **classification ledger**, not a zero-count: `.moai/reports/t401/ac013-ledger.md` carries one
row per candidate, each classed `conditioned` or `unconditioned-by-design: <reason>`, with the row
count equal to the swept count and no unclassified remainder. A row is `conditioned` either on a
window-local mode reference or on a **named carrier** — a coordinate this SPEC forbids editing,
whose conditioning text lives at a coordinate the ledger row names (acceptance.md AC-JFM-013,
carrier form). The coordinates this SPEC names — `askuser-protocol.md:64` (S1's first coordinate,
one of the two contradicting `[HARD]` clauses named above), any other `askuser-protocol.md` S1 row,
`run.md:137`, `plan/spec-assembly.md:212`, `plan/spec-assembly.md:353`, and `zone-registry.md:869`
(carrier form) — must all land in `conditioned`. Coordinates are identified by anchor text;
line numbers are a locating aid measured at `82edb9109`.

The two `spec-assembly.md` coordinates are admitted on REQ-JFM-016's **reachability** test, the
same consequence test that admits `run.md:137` and `branch-origin-protocol.md:25` (both of which
are also outside §B.1's S1-S6 table). `:212` is the Implementation Kickoff Approval clause —
identical to `run.md:137`, one file over — and `:353` is the sole implementing site of the Frozen
`CONST-V3R5-035` clause whose doctrine site this milestone already conditions. A run phase MUST
NOT class a swept row `unconditioned-by-design` on coordinate-table absence; that judgment is a
blocker report to the orchestrator, not a run-phase classification (see acceptance.md AC-JFM-013).

The second sweep stage is **case-insensitive** by design. `askuser-protocol.md:64` reads
`**First option label**: MUST carry the `(권장)` …` with a capital `F`; a case-sensitive
`\bfirst\b` never matched it, so the SPEC's own first-named coordinate sat outside the swept set.
Measured in this tree at `ad272be20`: case-sensitive → `23` rows, case-insensitive → `25`, the two
added rows being `askuser-protocol.md:64` and `plan/spec-assembly.md:353`.

Covers: REQ-JFM-005, 010, 011, 012, 013, 014, 015, 016, 021, 022. AC-JFM-004, 009, 010, 011, 012, 013.

### M2 — The banner surfaces (Priority: High; user-facing render change)

| File | Change | Mirror |
|---|---|---|
| `.claude/output-styles/moai/moai.md` | Pull-mode withholding branch on the Discovery `⏭️ Recommended action:` rule (S2) and the Epic Stats / Epic Status `⏭️ Next:` rules (S3); user-judgment slot on the Insight banner (S4); preference-neutral option ordering on Error Recovery (S5). All 4 locale columns of the affected header tables kept in step | `internal/template/templates/.claude/output-styles/moai/moai.md` |
| `.claude/rules/moai/workflow/context-window-management.md` | Mode-conditional phrasing on the pre-clear announcement step 4 (S6) | `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` |

Decide before editing: whether `moai-easy.md` and `moai-learn.md` carry the same banner rules and
therefore need the same branch. Measure with `grep -n 'Recommended action' .claude/output-styles/moai/*.md`
and record the answer in `progress.md` §E.2 rather than assuming.

Verify:

```bash
grep -c 'recommendation_mode' .claude/output-styles/moai/moai.md
grep -n 'recommendation_mode' .claude/rules/moai/workflow/context-window-management.md
git diff ad272be20 --stat -- .claude/output-styles/moai/ .claude/rules/moai/workflow/context-window-management.md
```

Covers: REQ-JFM-006, 007, 008, 009, 021, 022. AC-JFM-005, 006, 007, 008.

### M3 — The config key (Priority: High; type-interface change)

Follow the landed default-off precedent (`Workflow.BranchGuard.Enabled`,
`workflow.integration_lock.enabled`, `workflow.agent_stop_guard.enabled`), but site the key under
`interview` because it configures interview output rather than a guard.

| File | Change | Mirror |
|---|---|---|
| `.moai/config/sections/interview.yaml` | Add `recommendation_mode: pull` (this repository dogfoods) | — (local dogfood value; the template default differs by design) |
| `internal/template/templates/.moai/config/sections/interview.yaml` | Add `recommendation_mode: push` with a neutral comment | (is the mirror) |
| `internal/config/types.go` | `RecommendationMode string \`yaml:"recommendation_mode"\`` on `InterviewConfig` | — |
| `internal/config/defaults.go` | `defaultInterviewConfig` sets `push` | — |
| `internal/config/interview_recommendation_mode_test.go` (new) | Default-`push` pin, `pull` round-trip, unrecognized-value fallback | — |

Verify:

```bash
go test ./internal/config/... -run 'RecommendationMode' -count=1
go test ./internal/config/... -count=1
```

Covers: REQ-JFM-001, 002, 003, 004, 021, 022. AC-JFM-001, 002, 003.

### M4 — The static CI guard (Priority: Medium)

| File | Change | Mirror |
|---|---|---|
| `.github/workflows/judgment-first-consistency.yaml` (new) | Path-triggered guard asserting the doctrine text and the §8 banner templates stay consistent with the pull convention; shaped like `template-neutrality-check.yaml` (path filters, `concurrency`, `permissions: contents: read`, isolated target) | — (no template counterpart; see below) |

**Why no mirror — the accurate reason.** Not "`.github/` is outside the distributed template": the
template does carry `.github/`, and `find internal/template/templates/.github -type f` returns four
files including `workflows/label-sync.yml` (research.md §6). The reason is REQ-JFM-021's
counterpart clause. This repository's 21 CI workflows are unmirrored; the one template-side
workflow is a scaffold shipped to user projects, not a mirror of a moai-adk job. A repository-local
doctrine guard has **no counterpart to match**, so the clause does not fire on it.

Demonstrate the guard in **both** directions: a scratch mutation that makes doctrine and banner
inconsistent must turn the job RED, and the clean tree must turn it GREEN. A guard shown only
green is vacuous.

Verify:

```bash
# local shape check
yq '.on.pull_request.paths, .jobs' .github/workflows/judgment-first-consistency.yaml
# mutation direction, on a scratch commit, reverted before merge
```

Covers: REQ-JFM-019. AC-JFM-017.

### M5 — Distribution close-out (Priority: Low; mechanical)

| File | Change |
|---|---|
| `internal/template/templates/**` | Confirm every M0-M3 mirror landed; sweep for neutrality violations |
| embedded FS | `make build` |

Verify:

```bash
git diff --name-only ad272be20 | grep -cE '^\.claude/|^\.moai/config/'   # swept count — MUST be > 0
git diff --name-only ad272be20 | grep -E '^\.claude/|^\.moai/config/'    # each has a template twin
grep -rn 'SPEC-JUDGMENT-FIRST-MODE\|REQ-JFM\|/Users/\|CLAUDE.local' internal/template/templates/
go test ./internal/template/... -run TestTemplateNeutralityAudit -count=1
make build && git status --porcelain
go test ./internal/config/... ./internal/hook/... ./internal/template/... -count=1
```

**The corrected `.sh` / `.sh.tmpl` drift loop.** The 0.1.0 recipe was defective and the audit
measured it: `[ -f "$b" ] && diff -q … || echo DRIFT` reports a `.tmpl` with **no base sibling** as
drift, because the `[ -f ]` guard fails into the `||` branch. The template hook directory holds 35
`.tmpl` files and 12 `.sh` files, of which only 4 form pairs, so the loop printed 31 spurious DRIFT
lines on an untouched tree — including `handle-pre-tool.sh`, the very wrapper the criterion named.
The corrected loop skips an absent base sibling instead of misreporting it, and prints its swept
count so an empty sweep cannot masquerade as a pass:

```bash
n=0; d=0
for f in internal/template/templates/.claude/hooks/moai/*.tmpl; do
  b=${f%.tmpl}
  [ -f "$b" ] || continue
  n=$((n+1))
  diff -q "$b" "$f" >/dev/null || { d=$((d+1)); echo "DRIFT $(basename "$b")"; }
done
echo "swept=$n drift=$d"
```

Covers: REQ-JFM-021, 022, 023. AC-JFM-019, 020, 021, 022.

### M6 — Vacuity falsifier collection (Priority: High for closure; gated on M0-M3)

AC-JFM-018 cannot be verified until the dogfood sample exists. This milestone is a **collection
window**, not an implementation step: with this repository at `recommendation_mode: pull` and the
observer live, accumulate at least 20 recorded `AskUserQuestion` calls and read the log.

Entry condition, checked before the window opens: M0's pre-landing baseline row exists
(`label_present: true`, `mode: push`). Without it the window measures nothing (§B item 6).

Verify:

```bash
jq -s '[.[] | select(.mode=="pull")] | {n: length, violations: ([.[] | select(.label_present==true)] | length)}' \
  .moai/reports/t401/pull-window.jsonl
```

The verify reads the **exported** artifact, not this tree's `.moai/logs/`. Two reasons, and both are
load-bearing. The observer's live log name is `askuser-observations.jsonl`
(`internal/hook/askuser_observer.go`, `askUserObservationFileName`) — `<observer>` was an unresolved
placeholder. And the log lands under the *asking* session's project directory, so reading this tree's
`.moai/logs/` yields a permanently empty file whenever the asking session lives elsewhere; `jq` over
an empty file returns `violations: 0`, which reads as "no violations" while nothing was measured at
all. The exported artifact carries its provenance record alongside it (§D.6, REQ-JFM-025).

No `question_type` filter appears here, in `acceptance.md`, or in `design.md` §6.3 — one rule, one
denominator: every `pull`-mode row (design.md §6.3 is the owner).

A sample below the floor is a **gap**, recorded as such. It is never reported as a pass.

Covers: REQ-JFM-020, 025. AC-JFM-018. (Entry-gated on AC-JFM-023, whose green path is M0.)

## §G Anti-Patterns to Avoid

- **Authoring a parallel mechanism.** The mode generalizes the existing Adaptive-strength
  principle. A second, separate suppression clause is the failure this SPEC exists to avoid.
- **Rewriting the Frozen clause.** `CONST-V3R5-035` keeps its text as the push-mode branch. Editing
  the registry `clause:` string breaks the canary gate.
- **Editing the auditor report contract.** Out of scope (spec.md §A.3, CONST-4) — on the two most
  common plan-audit paths no user ever reads it.
- **Landing the mirror "next milestone".** `moai update` wipes `.claude/rules/moai`,
  `.claude/skills/moai*`, and `.moai/config` wholesale; a local-only file is destroyed silently.
- **Adding a deny path to the observer.** CONST-7. The observer measures; it does not gate.
- **Claiming AC-JFM-018 from doctrine text.** The whole point of that criterion is that text
  cannot satisfy it.
- **Turning the mode on in the template.** CONST-1. The distributed default is `push`.
- **Running the full test suite locally.** Affected packages only; CI owns the full verdict.

## §H Cross-References

- `spec.md` — requirements, exclusions, boundary notes
- `acceptance.md` — the AC matrix and the Definition of Done
- `design.md` — the mode-resolution flow, the observer's record shape, and the doctrine-edit shape
- `research.md` — the measured surface inventory and the two negative findings
- `.moai/reports/t401/` — preflight and the two lens reports
