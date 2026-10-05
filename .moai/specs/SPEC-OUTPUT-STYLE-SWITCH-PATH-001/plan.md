# Implementation Plan — SPEC-OUTPUT-STYLE-SWITCH-PATH-001

Card t906. Tier M. Ordered by decision-reversibility: the two content decisions
this card carries — what the added sentence says in each persona, and whether
`moai.md` gains guidance it has never had — lead; the mechanical mirror and
measurement work comes last.

## §A. Context

Five switch-guidance sites across four files teach one route to switching output
styles; the runtime documents two. See `spec.md` §A for the measurements — they
are not restated here.

Card **t878** measured the command live on 2.1.275 and named these files as its
follow-up (finding F3). Nothing in this card revisits t878's own surface.

## §B. The decisions this card makes (highest reversibility first)

### B.1 What the added text says, per persona

This is the reversible decision, and it is a **writing** decision rather than a
structural one: the mechanics (which lines, which files, both sides) are fully
determined by the measurement, and only the wording is open.

The constraint that does the work is REQ-OSP-004 crossed with REQ-OSP-005: three
files, three registers, and the two edited files must not end up carrying the
same sentence. The failure mode is specific and cheap to fall into — write one
good sentence naming both routes, then paste it at all five sites. That satisfies
"names both routes" perfectly and produces a warm persona speaking in the
tutor's voice.

Shape per file, as a direction rather than as copy the run phase must transcribe:

- **`moai-easy.md`** (4 sites) — colloquial and second-person, matching the existing "go ahead and switch to **MoAI**" voice. The four sites are not interchangeable: line 21 is a redirect *away* to a more powerful persona, 493 is a FAQ answer about switching back, 496 is a redirect to MoAI-Learn, and 533 is the standing "switch any time" line. The command belongs most naturally on the standing line and on the FAQ answer; whether all four sites carry both routes or only the ones where a command reads naturally is a judgement the run phase makes **within** REQ-OSP-001, which requires both routes at each of the five sites — so all four do carry both, and the variation is in phrasing, not in coverage.
- **`moai-learn.md`** (1 site) — line 35 is a `[HARD]` bullet that quotes an instruction to the reader (`"Switch to MoAI via /config → Output style → MoAI"`). The addition goes inside that quoted instruction, so the tutor's voice and the bullet's structure both survive.
- **`moai.md`** — no site; see B.2.

**Anti-pattern, stated because it is the likely one:** writing `/config` → Output
style **or** `/output-style <style>` as a bare parenthetical appended to every
site. It is the paste failure wearing a different hat — REQ-OSP-005 catches it
between the two files, but not between the four sites inside `moai-easy.md`.
Those four are a review concern, named here so review knows to look.

### B.2 Whether `moai.md` gains switch guidance — RESOLVED

**Ruling (lead, 2026-09-18, recorded before Implementation Kickoff Approval):
option (가) — four files changed. `moai.md` (both copies) stays byte-unchanged.**

The question this section carried was whether `moai.md` should gain output-style
switch guidance it does not currently carry. It is settled, and the plan above is
unchanged by the answer: the ruling confirms the scope this plan was already
written to rather than re-scoping it.

**The deciding predicate — the same one that cut docs-site out of scope: a file
is not opened to fix a defect it does not have.** `moai.md` carries no switch
guidance (`spec.md` §A.3 — zero switch-guidance sites; all seven of its `/config`
occurrences are the `language.yaml` configuration path), so there is no
incomplete guidance there to complete. Adding guidance would be a **new surface,
not a correction** — the same category as
`docs-site/content/*/claude-code/foundations/interactive-mode.md`, which §A.5
measured as teaching neither route and which `spec.md` §D excludes on exactly
this reasoning. One predicate, applied twice, giving the same answer both times.

**The distinction the lane raised is real and was NOT dismissed.** docs-site is
silent on a *different* topic; `moai.md` is silent on the *same* topic its two
sibling personas do address. That asymmetry is a genuine finding, and it has been
**split into its own card, t936**, framed as *"of the three personas, only
`moai.md` has no exit route"*. t936 carries this card's measurements as its
evidence: switch-guidance sites **4 / 1 / 0** across `moai-easy` / `moai-learn` /
`moai`; the sibling-persona names `moai-easy` and `moai-learn` returning **grep 0
hits** inside `moai.md`; and both contrasted against the `moai-easy.md` §14 switch
table and the `moai-learn.md` line 35 pointer.

**t936 carries a [HARD] precondition of its own**: decide first whether the
silence is a *defect* or a *deliberate design choice* — a professional persona may
intentionally not advertise switching away from itself. The `git log` of the three
files shows when the asymmetry began and is the evidence for that decision. If the
silence turns out to be deliberate, then **recording that fact is t936's
deliverable** — there is no edit to make, and the record is what stops the question
being reopened a third time.

That judgment is kept **out of t906 deliberately**: mixing "is this persona's
silence intentional?" with "fill in guidance that is incomplete" would make the
change unreviewable — a reviewer would have to adjudicate a design question while
reading a four-file correction diff.

Consequences for this SPEC, all already in place and now asserted rather than
assumed: the editable population stays at the five measured sites in four files
(`spec.md` §C); `spec.md` §D's exclusion of `moai.md` stands on the predicate
above rather than pending an answer; and **both `moai.md` copies are asserted
byte-unchanged by AC-OSP-006**, whose paired empty/non-empty diff makes an
incidental edit fail a criterion instead of passing unnoticed.

### B.3 Both sides carry the same bytes

Not a decision so much as a measured property with a consequence. The three
pairs are byte-identical today (`spec.md` §A.1), so each edit is authored once
and applied to both copies with identical content — not authored separately and
reconciled, which is how the two sides drift a character at a time.

This property is **local to this file group**. `.claude/**` and
`internal/template/templates/.claude/**` diverge deliberately elsewhere in this
repository, so nothing here licenses a `cp` between the trees for any other path.

## §C. Pre-flight

- Re-measure the baselines in the run-phase tree rather than reading them from `spec.md`: anchored `/output-style` counts (0 / 0 / 0 both trees), `/config` counts (7 / 7 / 7), `language.yaml` counts (7 / 3 / 6), and the three `cmp` results.
- Confirm §B.2's ruling is the operative scope before the first edit: four files changed, both `moai.md` copies untouched. The clarification is **resolved** (§B.2, lead ruling 2026-09-18, recorded in `progress.md` §E.1); the run phase reads that ruling, it does not re-open the question or supply a default of its own.
- Record the `internal/template` package's **pre-existing** test state before changing anything. The CI guard workflow states the package carries failures unrelated to this work, so a package-wide green is not the bar — the two isolated `-run` targets are (§E).
- Re-read `git rev-parse --short HEAD` and `git branch --show-current` immediately before any commit; never a value read earlier in the turn.

## §D. Milestones

### M1 — the five sites, both sides

Files:

- `.claude/output-styles/moai/moai-easy.md` — lines 21, 493, 496, 533.
- `.claude/output-styles/moai/moai-learn.md` — line 35.
- `internal/template/templates/.claude/output-styles/moai/moai-easy.md` — the same four edits, same bytes.
- `internal/template/templates/.claude/output-styles/moai/moai-learn.md` — the same edit, same bytes.
- Both `moai.md` copies — UNCHANGED, verified by diff (AC-OSP-006).

Line numbers are the base-tree coordinates; they move as the file is edited.
Locate each site by its content (`/config` adjacent to switch prose), not by
re-applying a stale line number.

The template-side text is written to the neutrality rules from the start
(REQ-OSP-008) rather than written freely and scrubbed afterwards — the working
copy and mirror carry identical bytes, so a class that fails the template guard
fails on both sides or neither, and scrubbing one side breaks C1.

### M2 — evidence

Files: `.moai/reports/t906/` only — no source change.

- The anchored `/output-style` per-file counts, before and after, on **both** trees.
- The three `cmp` results, after.
- `/config` and `language.yaml` per-file counts, before and after.
- `git diff --stat` naming exactly the four changed files, and the empty `moai.md` diff beside a non-empty diff on a file that did change (the C7 non-empty-sweep control).
- The two isolated neutrality-guard runs, verbatim, with exit codes.
- The `moai.md` unanchored-vs-anchored pair (2 vs 0), re-measured, keeping the §A.2 false positive documented rather than folkloric.
- The added sentences quoted side by side, so the register decision is reviewable as text rather than as an assertion.

Evidence is **exported to the primary checkout** before it is cited:
`.moai/reports/**` is gitignored, so an in-worktree copy reaches no clone and the
citation stops resolving once this worktree is disposed.

## §E. Self-verification

- `go test ./internal/template/... -run 'TestTemplateNeutralityAudit' -v` — exit 0.
- `go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak' -v` — exit 0.
- `go test ./internal/template/...` — run, and its result compared against the pre-flight baseline from §C. A pre-existing failure is not this card's to fix and not this card's to absorb: the comparison is what separates the two. Any failure **new** against the baseline is a blocker.
- No `go build`, no `make build`, no `moai doctor` run. The embed refresh is owed by the change and is discharged at batch close by the lead (REQ-OSP-010); running it here would also make every subsequent local measurement depend on a binary this card did not pin.

## §F. Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| One sentence pasted into both personas | REQ-OSP-001 satisfied, REQ-OSP-004 violated; the warm persona speaks in the tutor's voice | REQ-OSP-005; AC-OSP-011's byte-inequality check; §B.1's named anti-pattern |
| Four `moai-easy` sites given the same parenthetical | Same failure, invisible to REQ-OSP-005 (which compares across files) | Review of the M2 side-by-side quotes; named in §B.1 rather than left to chance |
| `/config` replaced instead of joined | One incomplete instruction traded for another | REQ-OSP-002; AC-OSP-004's unchanged 7 / 7 / 7 |
| A `language.yaml` path caught by a careless `/config` sweep | A configuration reference silently rewritten | REQ-OSP-006; AC-OSP-005's unchanged 7 / 3 / 6 |
| Only one side of a pair edited | Working copy and template drift; `moai update` later reverts the user-visible half | REQ-OSP-007; AC-OSP-003's three `cmp` runs |
| `moai.md` edited on the strength of the card's framing | The scope the §B.2 ruling excluded is entered anyway, and the t936 design question is decided inside a correction diff | §B.2 ruling (option 가); AC-OSP-006's empty diff, paired with a non-empty one so the assertion cannot pass vacuously |
| Template text carries a rejected neutrality class | CI guard fails on a docs-only change | REQ-OSP-008; AC-OSP-007's two isolated runs; write-to-the-rule in M1 |
| `make build` run here | Contradicts the lead's batch-close ruling, and pins measurements to an unpinned binary | REQ-OSP-010; §E states the omission explicitly so it reads as a decision, not an oversight |
| A pre-existing `internal/template` failure read as this card's regression | A clean change blocked, or a real regression absorbed into "pre-existing" | §C baseline recorded before any edit; §E compares rather than asserts |

## §G. Anti-patterns

- Editing `moai.md` because the card text says "the three persona files". The measurement says otherwise, and the measurement is what the SPEC records.
- Treating the unanchored `grep -c '/output-style'` count as evidence. It counts a file path (`spec.md` §A.2).
- Claiming `/output-style <style>` performs a switch. The probe observed the command's usage output, not a completed switch (`spec.md` §A.4).
- `cp`-ing between `.claude/` and `internal/template/templates/.claude/` as a general habit. The byte-identity is a property of this file group only.
- Adding the guidance to docs-site because it is adjacent. That is a new four-locale surface under the same-PR obligation, and it is a separate card.
- Running `make build` "to be safe".

## §H. Cross-references

- `.claude/output-styles/moai/{moai,moai-easy,moai-learn}.md` — the working copies
- `internal/template/templates/.claude/output-styles/moai/` — the three mirrors
- `.github/workflows/template-neutrality-check.yaml` — the guard, and the note that the `internal/template` package carries pre-existing failures
- `.moai/reports/t878/verdict.md` (primary checkout) — finding F3, which named these files as the follow-up
- `docs-site/content/*/claude-code/foundations/interactive-mode.md` — the four-locale page named in the proposed follow-up card
