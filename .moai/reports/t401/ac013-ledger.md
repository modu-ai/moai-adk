# AC-JFM-013 classification ledger — SPEC-JUDGMENT-FIRST-MODE-001 (card t401)

Sweep window: **the enclosing markdown block of the matched line**, drawn by the mechanical
boundary rule in `acceptance.md` AC-JFM-013 (blank line / heading / code fence / list-item start /
table row). This ledger was **re-swept in full under that window**; the previous revision was built
under the retired matched-physical-line window.

Sweep command (verbatim):

```
grep -rn -E '\(Recommended\)|\(권장\)' .claude/rules .claude/skills .claude/output-styles \
  | grep -iE '\bfirst\b|첫 |먼저' > .moai/reports/t401/ac013-candidates.txt
```

| measurement | value |
|---|---|
| tree measured | worktree `WT-analysis-pull`, HEAD `6839f717a`, clean tree |
| swept count (`wc -l < ac013-candidates.txt`) | **26** |
| ledger rows below | **26** |
| rows classed `conditioned` | **11** |
| rows classed `unconditioned-by-design` | **15** |
| rows **escalated** (unclassified remainder) | **0** |

**Counts revised by the M3 `harness.md` conditioning edit** (operator-approved resolution ① of
§ Escalation), measured on the post-edit working tree whose parent is `846b38b28`. The escalation
is discharged: **rows 18 and 20 both move to `conditioned`**, carrier
`.claude/skills/moai/workflows/harness.md` `:75` + `:190` and the template mirror. The delta from
the pre-edit revision is therefore `9 / 16 / 1` → `11 / 15 / 0` — **two** rows moved, not one. Row
20 (escalated → conditioned) is the predicted movement; row 18 (`unconditioned-by-design` →
`conditioned`) is the second, and it follows from the disposition rule the row itself already
carried ("this row's disposition follows row 20's resolution"), not from a reclassification of its
merits. A prediction of `escalated → 0` alone would have left the `conditioned` /
`unconditioned-by-design` split at 10 / 16, which does not sum to 26; the two-row movement is what
makes the counts close.

**The counts differ from the SPEC's prediction (8 / 18 / 0), and the difference is reported as a
difference, not reconciled.** The prediction was deduced from line ⊆ block monotonicity, which
holds — every row `conditioned` under the line window is still `conditioned`. The two movements are
not window effects:

- **Row 4 (`zone-registry.md:869`) moved `unconditioned-by-design` → `conditioned`** because
  `acceptance.md` v0.2.4 now names it in the required-`conditioned` list and admits it on the
  **carrier form**. The block window alone would not have moved it: the block is the whole
  `CONST-V3R5-035` entry (`864`-`870`) and carries no mode reference, so form (a) fails and form
  (b) carries it.
- **Row 20 (`harness.md:190`) moved `unconditioned-by-design` → escalated** on re-reading its
  enclosing block. See § Escalation.

**Window effect, measured.** Under the block window only **3** of 26 rows have a block wider than
the matched line: row 4 (`864`-`870`), row 7 (`263`-`269`, inside a code fence), and row 23 (the
wrapped `[HARD]` paragraph the window repair was made for). The other 23 blocks are exactly the
matched line, because a blank line or a list-item start bounds them immediately.

**A limitation of the window worth recording.** Boundary rule 4 stops the upward scan **at** a
list-item start, so a list item's **parent lead-in paragraph is never inside the block**. The
window therefore does *not* mechanically capture the lead-in that row 12 turned out to depend on.
For this re-sweep the lead-in was read by hand for all 26 rows regardless — that reading is what
produced the row-16 confirmation and the row-20 escalation below. The window and the lead-in are
independent obligations; satisfying the first does not satisfy the second.

**Classification test used.** REQ-JFM-016 names *clauses **mandating** a `(Recommended)` / `(권장)`
first option*. The discriminant applied throughout is therefore **modal force**, which is textual
and reads off the block:

- A clause carrying a **mandate** (`MUST`, or an unqualified `[HARD]` directive) that neither
  carries a mode reference nor cites an SSOT that delivers one → `conditioned`, or escalated when
  conditioning it needs a file outside this SPEC's edit surface.
- A clause that **identifies which option slot** something occupies, **illustrates** a pattern in a
  code fence, or **asserts** the shape of one concrete option set → not the clause class
  REQ-JFM-016 names → `unconditioned-by-design`. Under `pull` such a clause resolves to its
  slot/ordering content with no label, so it is satisfied rather than contradicted.

Neither ground is membership in §B.1's coordinate table, and neither is "the file belongs to
another milestone" — both remain classifications this run phase may not make.

**Carrier form.** A row is `conditioned` when EITHER (a) a mode reference appears in its block, OR
(b) the clause is one this SPEC forbids editing and the row **names the coordinate where the
conditioning actually lives**. Two rows take form (b): row 8 (`branch-origin-protocol.md:25`,
carrier `:26`) and row 4 (`zone-registry.md:869`, carriers `branch-origin-protocol.md:25-26` and
`spec-assembly.md:353`). Form (b) requires no edit to `zone-registry.md`, so AC-JFM-012's
empty-diff invariant is untouched.

**The third ground recorded in the previous revision is now gone.** Row 4 was previously classed
`unconditioned-by-design` on an explicit SPEC exclusion — a ground the ledger's two-class contract
did not describe. Under the carrier form it is `conditioned`, so the two-class contract holds and
no row uses a third ground.

## Rows

| # | Coordinate | Class | Reason |
|---|---|---|---|
| 1 | `.claude/rules/moai/core/askuser-protocol-reference.md:34` | unconditioned-by-design | Block = the line (heading above, blank below). Detail companion of `askuser-protocol.md` § Recommendation Placement Principles, paths-scoped to it and loaded only with it. Its MUST governs *which* option may carry the label when one exists, not that one must exist; under `pull` no option carries a label and it is satisfied vacuously. |
| 2 | `.claude/rules/moai/core/askuser-protocol-reference.md:49` | unconditioned-by-design | Block = the line. Same companion. This row *is* the adaptive-strength principle the mode axis generalizes, and already describes a label-omitting branch. |
| 3 | `.claude/rules/moai/core/askuser-protocol-reference.md:122` | unconditioned-by-design | Block = the line (list item). Same companion; restates the bias-prevention rule for `preview` content, whose home (§ Option Description Standards) AC-JFM-010 requires byte-unchanged. Same vacuous-satisfaction ground as row 6. |
| 4 | `.claude/rules/moai/core/zone-registry.md:869` | **conditioned** | Block = `864`-`870`, the whole `CONST-V3R5-035` entry; it carries no mode reference, so form (a) fails. Admitted on **form (b)**: REQ-JFM-015 / spec.md §B.2 forbid editing the Frozen `clause:` string, and the conditioning lives at its doctrine site (`branch-origin-protocol.md:25-26`, row 8/9) and its implementing site (`spec-assembly.md:353`, row 24), both `conditioned` below. Same shape and same carrier chain as row 8; classing the two differently was an artifact of the retired line window. Requires no edit to `zone-registry.md`. |
| 5 | `.claude/rules/moai/core/askuser-protocol.md:64` | **conditioned** | Block = the line (list item 3). S1's first-named coordinate (spec.md §B.1). Mode reference in block. AC-JFM-004 does not reach it — that criterion is satisfied inside § Recommendation Placement Principles, while `:64` lives in § Socratic Interview Structure. |
| 6 | `.claude/rules/moai/core/askuser-protocol.md:83` | unconditioned-by-design | Block = the line (blank above and below). The § Option Description Standards bias-prevention clause. It mandates **no** first-option label — it constrains *where* a recommendation signal may be carried when one exists — so it is not the clause class REQ-JFM-016 names; under `pull` no option carries a preference claim, so there is no signal for it to misplace and it is satisfied vacuously. Independently, AC-JFM-010 requires this section byte-unchanged from `ad272be20`. ~~Not an S1 coordinate (§B.1 lists `:64`, `:217`, `:245`, `:253`).~~ **[STRUCK]** — membership test, see the strike note. |
| 7 | `.claude/rules/moai/core/askuser-protocol.md:265` | **conditioned** | Block = `263`-`269`, the fenced First-Action Sequence. Mode reference in block. |
| 8 | `.claude/rules/moai/development/branch-origin-protocol.md:25` | **conditioned** | Block = the line (list item). Doctrine site of Frozen `CONST-V3R5-035`; text kept verbatim as the `push` branch per REQ-JFM-015 / §B.2. **Form (b)** — carrier is the adjacent pull branch at `:26` (row 9). |
| 9 | `.claude/rules/moai/development/branch-origin-protocol.md:26` | **conditioned** | Block = the line (list item). The pull-mode branch added by M1. Self-conditioning. |
| 10 | `.claude/rules/moai/workflow/archived-agent-rejection.md:127` | unconditioned-by-design | Block = the fenced recovery-procedure listing. A code-fenced **illustration** of a recovery round, not prose doctrine: it carries no MUST and creates no mandate. Correction to the previous revision, which said it "inherits the conditioned SSOT" — **no SSOT citation exists anywhere in this block**; that was an inference. The surviving ground is the modal one. |
| 11 | `.claude/rules/local/ci-watch-protocol.md:98` | unconditioned-by-design | Block = the line (blank above and below). Carries a MUST, but states it "per `.claude/rules/moai/core/askuser-protocol.md`" — an unqualified whole-file SSOT citation, and that file's § Socratic Interview Structure constraint 3 is conditioned (row 5), so the citation delivers the conditioning. Also a `rules/local/` dev-only file that is never distributed. |
| 12 | `.claude/skills/moai/SKILL.md:350` | **conditioned** | Block = the line (list item); its `[HARD] Beginner-Friendly Option Design:` lead-in at `:348`-`:349` sits **outside** the block and was read separately. Independent `[HARD]` mandate with no SSOT citation and maximal reach — it meets REQ-JFM-016's reachability test. Mode reference in block. M1 scope expansion approved by the operator; the lead-in header, bullet order, and second bullet are untouched. |
| 13 | `.claude/skills/moai/workflows/moai.md:205` | unconditioned-by-design | Block = the line (list item). **Slot identification, not a mandate**: the clause's normative subject is that the sync chain is *surfaced rather than fired silently*, and `"(Recommended)"` names which slot of the existing next-step question it occupies. Under `pull` it resolves to "the chain occupies the first option, unlabelled" and the anti-silent-chain requirement is preserved intact. (Previous reason said "descriptive"; sharpened to the consequence argument.) |
| 14 | `.claude/skills/moai/workflows/run.md:97` | unconditioned-by-design | Block = the line. The same `single-phase` chaining contract as row 13, one file over. Same slot-identification ground. |
| 15 | `.claude/skills/moai/workflows/run.md:137` | **conditioned** | Block = the line. The Implementation Kickoff Approval `[HARD]` clause (spec.md §E.1). Mode reference in block; the gate stays mandatory and score-independent in both modes. |
| 16 | `.claude/skills/moai/workflows/harness-build-entry.md:75` | unconditioned-by-design | Block = the line (numbered step 2). Its **parent lead-in at `:73` was read** and cites the SSOT unqualified — "conduct AskUserQuestion Socratic rounds per `.claude/rules/moai/core/askuser-protocol.md`" — so this step is a restatement of that file's § Socratic Interview Structure constraint set, whose constraint 3 is conditioned (row 5). The citation delivers the conditioning. Flagged as most-suspect before the sweep; the reading **confirmed** the class rather than overturning it. |
| 17 | `.claude/skills/moai/workflows/harness-build-entry.md:120` | unconditioned-by-design | Block = the line (blank above and below). A **non-mandate directive** introducing the Phase 7 option list — "with the canonical four-option pattern (first option `(권장)` / `(Recommended)`)". No MUST; the mandate in this section is that the gate *fires* (the `[HARD]` two lines above), which `pull` leaves untouched. See § Residual risk — this row is the closest call in the ledger. |
| 18 | `.claude/skills/moai/workflows/harness.md:75` | **conditioned** | Block = the line. The `### Canonical Four-Option Pattern` **template lead-in** to a fenced example; a non-mandate directive. Its normative force is supplied by row 20, which invokes this pattern with a MUST — so this row's disposition follows row 20's, and row 20 is now `conditioned`. **Resolved by the M3 edit**: the lead-in now names itself the `push`-mode form and carries the `pull` branch inline, so a reader arriving from row 20 under `pull` is not sent to an unconditioned template. The fenced example below it is left exactly as it is — a code-fence illustration, the `unconditioned-by-design` class of rows 21-22, and it is now explicitly scoped by the lead-in that introduces it. |
| 19 | `.claude/skills/moai/workflows/harness.md:131` | unconditioned-by-design | Block = the line (sub-list item). Names which option is **first** for a specific verb (`Continue (권장)` on `rollback` vs `Abort (권장)` on `apply`); the subject is option ordering for that gate. Under `pull` the ordering claim survives and the labels drop. |
| 20 | `.claude/skills/moai/workflows/harness.md:190` | **conditioned** (was ESCALATED) | Block = the line (numbered step 5). Carried an explicit **MUST** — "The first option `Apply (권장)` MUST carry the `(권장)` / `(Recommended)` suffix per `.claude/rules/moai/core/askuser-protocol.md` **§ Option Description Standards**". The previous revision classed it `unconditioned-by-design` because a citation was present. **The citation does not deliver conditioning**: it is section-scoped to § Option Description Standards, the one section AC-JFM-010 requires **byte-unchanged**, so following it under `pull` yields no withholding instruction and the MUST stands unconditioned. It cannot be satisfied vacuously the way row 6 is — it positively requires the suffix. **Resolved by the M3 edit**: the MUST is now scoped to `push`, the `pull` branch withholds the suffix from every option, and the mode branch cites § Recommendation Placement Principles → Recommendation mode — a section that does carry the branch — while the § Option Description Standards citation is retained for the `push` clause it correctly governs. The pinned section is untouched (measured: identical SHA256 against `ad272be20`). See § Escalation (discharged). |
| 21 | `.claude/skills/moai/workflows/feedback.md:124` | unconditioned-by-design | Block = the line. An **assertion about one concrete option set** ("`(Recommended)` is carried by the first option only"), describing the submission gate's own table immediately above. Not a general mandate. |
| 22 | `.claude/skills/moai/workflows/run/mode-orchestration.md:79` | unconditioned-by-design | Block = the line (list item). The same `single-phase` chaining contract as rows 13-14, third file. Same slot-identification ground. |
| 23 | `.claude/skills/moai/workflows/plan/spec-assembly.md:212` | **conditioned** | Block = the wrapped `[HARD]` paragraph (`208`-`214`) — **the case the window repair was made for**; under the retired line window the token had to sit on one physical line. Mode reference in block. Same Implementation Kickoff Approval clause as row 15. The long line the run phase wrote is left exactly as it is, per AC-JFM-013. |
| 24 | `.claude/skills/moai/workflows/plan/spec-assembly.md:353` | **conditioned** | Block = the line (sub-list item). Sole implementing site of Frozen `CONST-V3R5-035` (the Phase 13 BODP gate). Mode reference in block, so the doctrine site (row 8) and its implementation are conditioned together and the change is not half-applied. |
| 25 | `.claude/skills/hns-workflow-ci-loop/SKILL.md:198` | unconditioned-by-design | Block = the line (checklist item). A **verification checklist assertion** about what `EmitReadyToMergeReport` emits, in a dev-only, non-distributed `hns-*` skill governing `scripts/ci-watch/`. Not a mandate on AskUserQuestion composition. Correction to the previous revision, which said it "inherits the conditioned SSOT" — **no citation exists in this block**; the surviving ground is the modal one. |
| 26 | `.claude/output-styles/moai/moai.md:65` | unconditioned-by-design | Block = the line (numbered step 1). **The M1 hedge is removed** — this file is M2's edit surface and M2 has now run, so the row is classed on its merits rather than deferred. It is a **non-mandate restatement** of the Socratic round constraints (≤4 questions, ≤4 options, user language, first option marked), the same modal class as rows 16-18, inside a Process list whose step 0 immediately above cites `askuser-protocol.md`. M2's edit surface is S2-S5 (the Discovery / Epic / Insight / Error-Recovery banner rules), which this line is not; conditioning it would be an edit outside the enumerated surface. See § Residual risk. |

## Escalation — row 20 (`harness.md:190`) — DISCHARGED

**Outcome.** The operator approved **resolution 1** below. The M3 edit conditions `:190` and `:75`
in `.claude/skills/moai/workflows/harness.md` and its template mirror; rows 20 and 18 are now
`conditioned` and the escalated count is 0. Resolutions 2 and 3 were considered and rejected —
2 because re-pointing the citation leaves the `MUST … suffix` wording itself unconditioned, so the
`pull`-mode conflict survives the re-pointing; 3 because it is a SPEC-body decision and would admit
a carve-out where a direct fix was available. The record below is preserved as the reasoning that
produced the choice, not as an open item.

**Why it is not classifiable in either class.** It carries an explicit `MUST` that the first option
carry the `(권장)` / `(Recommended)` suffix, so it is squarely the clause class REQ-JFM-016 names —
`unconditioned-by-design` is unavailable. Conditioning it requires editing
`.claude/skills/moai/workflows/harness.md`, which is in neither M1's nor M2's declared edit
surface — and classing it `unconditioned-by-design` **because** the file belongs to another
milestone is precisely the scope-based classification a run phase may not make. It therefore
escalates, exactly as row 12 did before the operator approved M1's scope expansion.

**Why the previous revision missed it.** The reason recorded was "states it 'per
askuser-protocol.md § Option Description Standards' — an explicit SSOT citation, so it inherits."
The citation exists, but its *target* was never checked. § Option Description Standards is the one
section AC-JFM-010 requires byte-unchanged from `ad272be20`; it therefore carries no mode branch
and never will under this SPEC. A reader under `pull` who follows the citation finds the
bias-prevention rule and no withholding instruction, and emits the label.

**Consequence for the criterion.** AC-JFM-013 requires "no unclassified remainder". With one row
escalated, that clause was **not satisfied** at HEAD `846b38b28`. It is satisfied after the M3
edit: 26 rows, 0 escalated. Three resolutions were available and the choice was the operator's:

1. Expand M2's (or a follow-up milestone's) edit surface to include `harness.md`, condition `:190`,
   and — because row 18 is the pattern template `:190` invokes — condition or re-ground `:75` in the
   same edit. This is the row-12 precedent.
2. Re-point `:190`'s citation from § Option Description Standards to the unqualified file or to
   § Recommendation Placement Principles, which does carry the mode branch. This is a smaller edit
   to the same out-of-scope file and changes what a reader is sent to read.
3. Amend the SPEC to admit a modal-force carve-out for a MUST whose only defect is a mis-scoped
   citation. This is a SPEC-body decision, not a run-phase one.

## Residual risk — the two closest calls

Both are rows where the modal-force test returns `unconditioned-by-design` but a literal reader
under `pull` could still emit the label. They are recorded rather than smoothed away.

- **Row 17 (`harness-build-entry.md:120`)** — the Phase 7 harness approval gate. Two lines above it,
  a `[HARD]` states the gate is mandatory and score-independent "parallel to the Implementation
  Kickoff Approval human gate" — the very gate this SPEC conditions in two other files (rows 15,
  23). The `(권장)` mention is a parenthetical pattern reference with no MUST, which is why it
  classes as it does, but the sibling-gate parallel is a real argument for conditioning it. The
  file is outside M1's and M2's edit surfaces.
- **Row 26 (`.claude/output-styles/moai/moai.md:65`)** — the modal test puts it with rows 16-18,
  but it sits in an **always-loaded output style**, which is the strongest form of REQ-JFM-016's
  reachability test, and its only nearby citation (step 0 at `:64`) is section-scoped to
  § ToolSearch Preload Procedure rather than to the Socratic constraints it restates. Conditioning
  it is a one-clause parenthetical edit to a file M2 already touches. It was **not** made, because
  `:65` is outside M2's enumerated S2-S5 surface and minimum-scope discipline governs. If the
  operator prefers it conditioned, the edit is small and this row moves to `conditioned`.

## Resolutions carried forward from the previous revision

**Row 12 — resolved by an approved scope expansion.** Originally escalated: classing it
`conditioned` needed an out-of-scope edit, and classing it `unconditioned-by-design` would have
rested on scope absence. The operator approved adding `.claude/skills/moai/SKILL.md` and its
template mirror to M1's edit surface, on the same reachability basis §E.1 uses to admit
`run.md:137` and `spec-assembly.md:212`.

One correction to the escalation's own record is preserved: it stated the mirror "was
byte-identical to the live file at `ad272be20`". **That was wrong, and it was an inference rather
than a measurement.** Measured at `ad272be20`, the pair carries pre-existing intentional
differences — the live copy uses `${CLAUDE_SKILL_DIR}` where the template uses literal
`.claude/skills/moai/` paths, and the live copy carries a `Last Updated:` line the template
deliberately lacks. Mirroring by `cp` would have destroyed them and injected an internal date into
the template, violating REQ-JFM-022. The edit was applied to each copy by hand.

**Row 6 — membership clause struck by orchestrator adjudication.** The reason originally closed
with "Not an S1 coordinate (§B.1 lists `:64`, `:217`, `:245`, `:253`)", which is a **membership
test** — the classification `plan.md` forbids a run phase from making. It is struck above rather
than deleted, so the record shows a strike. Ground (a) survives on a consequence test: `:83` is the
**Bias prevention** clause of § Option Description Standards. It does not mandate that a
first-option label exist; it governs where a recommendation signal may live when one does, barring
the description from carrying it. Under `pull` the label is withheld and no option carries a
preference claim, so there is no signal to misplace — the clause is satisfied vacuously rather than
contradicted. Under `pull` plus on-request emission the labelled form returns and the clause
applies unchanged.

## Open question against the criterion's own text

AC-JFM-013's required-`conditioned` list includes "any other `askuser-protocol.md` (S1) row". The
sweep contains two other `askuser-protocol.md` rows: `:265` (row 7, `conditioned`) and `:83`
(row 6, `unconditioned-by-design`). Read literally, the clause requires `:83` to be `conditioned` —
which **AC-JFM-010 forbids**, since it requires § Option Description Standards byte-unchanged. The
conflict resolves only by reading "(S1)" as narrowing the set to §B.1's coordinate table, and the
same criterion elsewhere calls using that table as the scope test **circular**. Recorded as a
SPEC-body question, not resolved here.

### Resolution of the open question (2026-09-07, session following the lead's continuation order)

Measured before writing: the sweep's `askuser-protocol.md` rows are exactly three — `:64`
(`**First option label**`, `conditioned` via M1's push-mode branch), `:83` (`**Bias prevention**`),
`:265` (`Step 2: Compose AskUserQuestion round`, `conditioned` via M1's push-mode branch). All
three verified by direct read at HEAD `a12beb541`.

The question resolves into two parts:

1. **`:83` does not need conditioning, and no edit may give it one.** The Bias prevention clause
   does not mandate a first-option label; it regulates where the recommendation signal may live
   *when one exists* ("conveyed exclusively through the label suffix on the first option"). Under
   `pull` no option carries a preference claim, so the clause is satisfied vacuously — the same
   consequence-test verdict row 6 already carries. Conditioning it would require editing a clause
   AC-JFM-010 protects byte-unchanged, turning AC-JFM-010 into a violated criterion for no
   behavioral gain. Row 6's `unconditioned-by-design` stands.

2. **The clause "any other `askuser-protocol.md` (S1) row" is defective in both readings, and the
   fix is a SPEC-body edit — manager-spec's surface, not the run phase's.** Read as "every
   `askuser-protocol.md` row in the sweep" it demands `:83` conditioning — the impossible
   direction AC-JFM-010 forbids. Read as "any row at §B.1's S1 coordinates" it revives the
   membership test 0.2.2 explicitly removed. The clause's practical intent is already
   discharged — all three swept coordinates are classified, two `conditioned`, reasons recorded —
   so what remains is wording precision only. Logged to the lead as a SPEC-body blocker: replace
   the clause with the reachability form (or pin the three coordinates explicitly) at the next
   manager-spec touch.

**Disposition: the question is CLOSED as resolved-as-measured.** AC-JFM-013 is satisfied on the
sweep + ledger the criterion defines — every swept candidate carries a class, no unclassified
remainder, all named coordinates `conditioned`. The wording defect is a SPEC-body debt carried to
the lead, not an unmet criterion condition; reading it as blocking would be the wrong-reason red
(the work this criterion measures is done; the debt lives in the sentence describing it).
