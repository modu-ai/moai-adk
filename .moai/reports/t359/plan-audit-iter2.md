# SPEC Review Report: SPEC-TODO-LANDING-EVIDENCE-001 — iteration 2

Card: **t359** · Iteration: **2/3** · Tier **L** (PASS threshold **0.85**)
Tree: worktree `.claude/worktrees/t359`, branch `WT-landing-evidence`, HEAD `483cea858`
Subject: SPEC v0.2.0 (`483cea858`), revised in response to iteration 1 (`2f1c36151`, FAIL 0.80)
Auditor: plan-auditor (Claude anchor only; no cross-model backend consulted)

**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.89** (Tier L threshold 0.85) — up from 0.80, no score regression, no STOP signal
**Blocking defects: 3** (2 major, 1 minor) · non-blocking: 4
**Iteration-1 regression: 13 of 13 defects RESOLVED**, one with a same-class recurrence in new text

Reasoning context ignored per M1 Context Isolation. The coordinator supplied two measurements to
build on; I spot-checked both rather than accepting them — `PRLinkOutcome`
(`internal/kanban/prlink.go:101-114`) does carry exactly `CardID`, `Kind`, `PRs`, `PRState`,
`Confidence` and no delivering-commit field, and both cross-referenced SPECs still read
`in-progress`. Everything else below was measured in this tree in this run.

Per the retry-loop contract this is a **delta re-audit**: scoped to the v0.1.0→v0.2.0 change plus a
regression pass over all thirteen iteration-1 defects. Structural must-pass checks were re-run in
full because the counts moved.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -o 'REQ-TLE-[0-9]*' spec.md | sort -u` → `REQ-TLE-001`
  … `REQ-TLE-021`, 21 ids, sequential, no gap, no duplicate, uniform 3-digit padding.
  `grep -c '^- \*\*REQ-TLE-'` = 21, so every id is also a definition. The two additions
  (020, 021) extend the sequence rather than interleaving.
- **[PASS] MP-2 GEARS compliance — requirement layer only.** Both new requirements conform:
  **REQ-TLE-020** (`spec.md:462`) is Event-driven with a chained second `when` clause
  ("**When** the operator supplies `--sha` … ; **when** either check fails or cannot be run …"),
  a GEARS-legal compound; **REQ-TLE-021** is Ubiquitous ("The two doctrine surfaces … shall carry
  byte-identical … rows"). The nineteen carried-over requirements are unchanged in pattern except
  REQ-TLE-001 (reduced per D12, still Ubiquitous) and REQ-TLE-019 (widened to a tuple assertion,
  still Ubiquitous). No `should` / `may` / weasel term in any REQ line; no `IF/THEN`.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present, correct types;
  `version: "0.2.0"` quoted semver, `status: draft`, ISO dates. `grep -nE
  '^(created_at|updated_at|labels|spec_id):'` → `rc=1`, no rejected alias.
- **[N/A] MP-4 Section 22 language neutrality.** Single-language SPEC (Go plus two markdown
  doctrine files); no multi-language tooling claim. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** Referenced SPEC set unchanged; all directories
  exist; measured statuses `in-progress` ×2 (`SPEC-KANBAN-QUEUE-PR-SYNC-001`,
  `SPEC-KANBAN-TODO-CLI-001`), `completed` ×3 (`SPEC-TODO-LANDING-STATE-001`,
  `SPEC-TODO-ANALYSIS-001`, plus `SPEC-TODO-ARCHIVE-QUERY-001` /
  `SPEC-TODO-DESTRUCTIVE-GUARD-001`). None `retired` / `superseded` / `archived` → no D7 BLOCKING.
  The iteration-1 D1 misdescription that ran alongside this check is now corrected (see regression
  table).
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate.** `grep -rn '\[NEEDS CLARIFICATION' ` over all five artifacts
  → `rc=1`.

All seven pass, as in iteration 1. Nothing in the revision touched the firewall.

---

## Regression check — all 13 iteration-1 defects

| # | Iteration-1 defect | Status | Evidence |
|---|---|---|---|
| D1 | Two SPECs asserted `completed`; both read `in-progress` | **RESOLVED** | `grep -n 'completed' spec.md` now returns only lines about `SPEC-TODO-LANDING-STATE-001` and `SPEC-TODO-ANALYSIS-001`, both of which measurably ARE `completed`. `spec.md:120` §A.3b carries the correction and re-grounds the binding on two measured properties instead of a lifecycle field; both were verified — `internal/kanban/prlink.go:101-114` carries five fields and no delivering commit, `internal/cli/todo_pr.go:1-15` states the writes-nothing property of the file. The HISTORY 0.1.0 row is annotated rather than rewritten, which preserves the audit trail |
| D2 | False ground for declining `--sha` validation | **RESOLVED** (requirement level) | §B.3.1 withdraws the ground explicitly ("**That ground is false and is withdrawn**") with a three-row table separating the grep predicate from `rev-parse` / `merge-base`, and adds REQ-TLE-020 rather than re-declining. See Hunt 1 — the criterion that enforces it has a new defect (E1) |
| D3 | §C.7 claimed AC-TLE-015/016 verify the doctrine text | **RESOLVED** | The claim is withdrawn in terms ("That was false"), replaced by an explicit verified-vs-DoD split and a new REQ-TLE-021 / AC-TLE-021 that does a real cross-artifact comparison. §D.3 now marks the uncovered DoD line in place |
| D4 | AC-TLE-015 left fields 1-5 unpinned | **RESOLVED** | Fields 1-5 now asserted "each asserted individually, not as a joined string", with a RED that fails on a confidence/state swap while every other clause still passes — a genuinely independent failure mode |
| D5 | AC-TLE-019 plants only on `items` | **RESOLVED** | Split into 019a/b/c; 019b plants on `archived_items` **alone**. Reproduced independently — see Hunt 2 |
| D6 | AC-TLE-014's project-root hash includes `.git/` | **RESOLVED** | `.git/` excluded, with two positive controls (non-empty set; queue DB present). The cited mutation path is exact: `internal/kanban/prlink_landed.go:150-154` (`out, err := q.Run("git", args...)`) reached via `internal/cli/todo_pr.go:57-60` (`exec.CommandContext(ctx, name, args...).Output()`) |
| D7 | §G deferred the downgrade proof to a criterion that does not close it | **RESOLVED** | §G no longer parks it: "**No pre-change binary is ever run against a post-change database, and no criterion in this SPEC closes that**", and REQ-TLE-018 "stays in this list after AC-TLE-018 passes". AC-TLE-018 itself says "this criterion narrows the gap, it does not close it". The parking spot is gone. Two *new* weaknesses in the added clause (b) are E2 below — new defects, not this one persisting |
| D8 | §E mapped REQ-TLE-004 → M1 | **RESOLVED** | §E row reads M3; `plan.md` M1.4 additionally disclaims it in text ("**AC-TLE-004 is NOT claimed here**") rather than silently dropping "in part" |
| D9 | AC-TLE-005's Given omitted the `spec_id` its Then asserts | **RESOLVED** | Given now reads "a card **carrying a `spec_id` that resolves to a fixture SPEC**", and additionally gained a reachable `--sha`, which the new REQ-TLE-020 makes necessary |
| D10 | Redundant `AskUserQuestion` grep conjunct | **RESOLVED** | Conjunct removed; the inherited guard is cited with its two controls, verified at `internal/cli/todo_test.go:451-480` (`filepath.Glob("todo*.go")`, `scanned < 2` positive control, synthetic-violation negative control) |
| D11 | §R.8's "verbatim" awk block was not verbatim | **RESOLVED**, but the class **RECURS** | §R.8 re-pasted complete and line-numbered; I re-measured lines 292-298 and they match byte for byte. However §R.10.1 / §A.3b introduce a new `$ grep` block whose four output lines have `.moai/specs/` stripped from every path, undisclosed → E4 |
| D12 | REQ-TLE-001 prescribed the DDL mechanism | **RESOLVED** | Reduced to the observable shape, with an explicit pointer sending the statement form to `design.md` §2 and the no-rebuild obligation to REQ-TLE-003 |
| D13 | Column-*name*-set guard misses type / `NOT NULL` drift | **RESOLVED — closed, not recorded** | REQ-TLE-019 now asserts `(name, type, notnull, dflt_value)` tuples. Verified empirically — see Hunt 2 |

**No iteration-1 defect is unresolved**, so the retry-loop's automatic-FAIL clause does not fire.
Eleven are cleanly closed; D7's fix is a genuine correction of the defect I raised (the parking
spot) even though the criterion it added carries new weaknesses; D11 is closed where I raised it and
reintroduced in new text. No defect appeared unchanged across both iterations, so there is no
stagnation signal.

---

## Hunt 1 — REQ-TLE-020: does the added check earn its keep?

**Verdict: the requirement is right and its central argument holds. The criterion's
attribution-boundary clause — the one the SPEC calls "the mechanical guarantee" — cannot fail.**

**The "record checks itself" argument holds.** §B.3.1 argument 3 is: a record asserts *this commit
delivered the card, observed against this ref*; a SHA unreachable from that ref contradicts the
record's own `ref` field. That is a coherence check on the record, not a check of the operator
against the machine, and it is structurally sound — the contradiction exists between two fields of
one record and is decidable without reference to any card. The supporting table is accurate: the
card id is genuinely not an input to `git rev-parse --verify <sha>^{commit}` or
`git merge-base --is-ancestor <sha> <ref>`, whereas it is the whole input to the grep predicate
(`LandedGrepArgs(ref, cardID)` builds `--grep=\b` + cardID + `\b`, `prlink_landed.go:96-108`).
Argument 1 is the strongest of the three and is the one that actually answers iteration-1's D2: the
SPEC's own "a stored mis-attribution is worse than a rendered one" applies to storage regardless of
authorship, so condemning machine mis-attribution while waving through operator mis-typing was
never an argument. Withdrawing the false ground and *adding* the check rather than re-declining it
is the right move, and §D's no-new-read-cost exclusion is untouched because the check runs in the
write verb.

**The mechanical boundary is not mechanical.** `acceptance.md:324` asserts the boundary by
"running (a) with the card **renamed** between two invocations and observing the same accept/reject
outcome", and calls this clause "not decorative … an implementation that fed the card token into
the validation would produce a card-dependent result and fail the rename assertion."

Measured: the grep token is the card **id**, not the card **text** —
`--grep=\b` + `cardID` + `\b` at `prlink_landed.go:106`. And no `moai todo` verb changes a card's
id: the sixteen documented verbs are `add analyze done drop edit export history list move next pr
relate undone undrop unrelate why`, registered at `internal/cli/todo.go:148-152`; `edit` changes
text only, and there is no rename-id verb at all. So a "renamed" card keeps the id the leaking
implementation would key on, the validation result is unchanged, and the assertion passes with the
leak intact. The clause tests the one variable the predicate does not read.

Worse, the criterion's own stated intent names the right variable and then measures the wrong one:
"the validation result is **independent of which card the record belongs to**". Varying *which
card* means two different card ids, not one card with two texts. The fix is a one-line substitution
and it does kill the leak: record the same `--sha` against **two different cards** and assert an
identical accept/reject outcome — a leaking implementation greps for `t1` in one call and `t2` in
the other and cannot produce the same answer for both unless the ref happens to satisfy both, which
the fixture controls. **E1, blocking.**

**On the declared residual.** §G's replacement text is honest where it matters — a typo landing on
another reachable commit is stored, marked `operator`, and no machine check can separate it from a
correct record, "because telling them apart is exactly the card-to-commit attribution REQ-1.10
forbids the machine to attempt". That reasoning is correct. But the residual then states the cost
as "nothing flags it; and it is corrected only when a human notices", and the SPEC does not
consider the cheapest remaining mitigation: rendering the commit's **own subject line** beside the
SHA on the `todo pr` row. Displaying a commit's subject makes no claim about which card it
delivered — it is the same class of non-attributing fact as `ref_head` — and it is precisely what
would let the human notice. The claim "no machine *check* can detect it" is true; the implied
"therefore nothing more can be done" is not. **E7, non-blocking** — this is a design option the
orchestrator may decline, not a correctness defect.

---

## Hunt 2 — AC-TLE-019a/b/c: are the three plants genuinely independent?

**Verdict: yes, all three. Reproduced with my own probe; D13 is genuinely closed rather than
recorded.**

I planted each case myself rather than reading the author's evidence file
(probe `internal/kanban/zz_audit2_t359_probe_test.go`, deleted after the run; tree verified clean):

```
$ go test ./internal/kanban/ -run 'TestAudit2_' -count=1 -v
=== RUN   TestAudit2_019b_ArchivedOnlyPlant
    items  names before="seq id text added_at spec_id state" after="seq id text added_at spec_id state"  ITEMS-ONLY GUARD TRIPS=false
    arch   names before="seq id text added_at spec_id state position" after="seq id text added_at spec_id state position bogus2"  ARCHIVED GUARD TRIPS=true
--- PASS: TestAudit2_019b_ArchivedOnlyPlant (0.01s)
=== RUN   TestAudit2_019c_TupleDrift
    nullable  names="seq id text added_at spec_id state landing"
    notnull   names="seq id text added_at spec_id state landing"
    NAME-SET ASSERTION SEPARATES THE TWO = false
    nullable  tuples="... state:TEXT:1:<NULL> landing:TEXT:0:<NULL>"
    notnull   tuples="... state:TEXT:1:<NULL> landing:TEXT:1:''"
    TUPLE ASSERTION SEPARATES THE TWO     = true
--- PASS: TestAudit2_019c_TupleDrift (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.573s
```

- **019b is independent.** With the plant on `archived_items` alone, an `items`-only assertion sees
  no change (`ITEMS-ONLY GUARD TRIPS=false`) while an `archived_items` assertion does
  (`=true`). A guard extended for one table therefore passes 019a and dies on 019b, exactly as the
  criterion claims. This is the iteration-1 D5 hole closed at the mechanism level, not restated.
- **019c is independent, and the plant is plantable.** SQLite **accepts**
  `ALTER TABLE items ADD COLUMN landing TEXT NOT NULL DEFAULT ''` (the ALTER did not error, so the
  mutant can actually be applied — a plant SQLite rejected would make the criterion unreachable).
  A name-set assertion cannot separate the two variants (`false`); a
  `(name, type, notnull, dflt_value)` tuple assertion does (`true`), the discriminating tuple being
  `landing:TEXT:0:<NULL>` versus `landing:TEXT:1:''`. So REQ-TLE-019's widening is load-bearing and
  iteration-1's D13 is closed rather than merely recorded.
- **019a is unchanged from the already-measured baseline**, which I reproduced in iteration 1.

**The decay hand-off is stated where run-phase will see it.** `acceptance.md` carries a **Decay
note** inside AC-TLE-019 itself — "The guard is owned by `SPEC-TODO-ARCHIVE-QUERY-001`. If that
SPEC's owner extends it independently on `develop`, this measurement decays and the run-phase MUST
re-measure rather than cite this criterion's recorded RED" — and `research.md` §R.10.5 repeats it.
That is the right placement: inside the criterion the run-phase reads, not only in a research
appendix. This directly answers a residual I raised at the end of iteration 1.

One honest limit the SPEC states and I confirm: §R.10.5 records that the author did **not** re-run
my `archived_items` probe and cites my report instead. Citing a second observer's measurement while
labelling it as cited rather than re-measured is the correct handling.

---

## Hunt 4 — AC-TLE-018 clause (b): narrowing, or relocation?

**Verdict: a genuine but thin narrowing. §G's wording is accurate about what it does not prove, and
silent about two structural weaknesses — one of which is a real drift hazard.**

**What clause (b) genuinely adds.** It reaches `backlog_sqlite.go:278-299` — DDL exec →
`schemaVersion` read → version switch — which clause (a)'s bare SQL statements never touch. The
concrete fact it establishes is worth having: re-executing `CREATE TABLE IF NOT EXISTS` against a
database whose `items` already carries an extra column does not error, and the version read still
returns `"1"`. Nothing else in the SPEC asserts that. So it is not pure relocation.

**Weakness 1 — no independent failure mode.** The criterion's RED lists exactly one mutation that
reds (b): bumping `backlogSchemaVersion`. That same mutation already reds clause (a)'s version
assertion. The criterion presents this as a strength ("the two clauses fail independently, which is
the point of adding (b)") — they fail *separately* on one mutation, which is not the same as having
independent failure modes. I could construct no mutation permitted by this SPEC that reds (b) alone:
the ALTER is additive, `CREATE TABLE IF NOT EXISTS` is a no-op on existing tables, and REQ-TLE-001
forbids the index and the version bump that would otherwise reach the reconstructed path. Clause (b)
is therefore closer to a coverage assertion than a detector. Not a defect on its own — a coverage
assertion over an otherwise-unreached path is legitimate — but the criterion overstates it.

**Weakness 2 — the frozen replica has no drift detector. This is the real problem.** `design.md`
§2 rules that "The DDL const itself is **not** edited to include `landing`" (verified at
`design.md:57-60`), and REQ-TLE-018 forbids the version bump. So the "frozen test-local copy … as
they stand at HEAD `e50964ad3`" is, today, **byte-identical to the live `backlogDDL` const and the
live switch**. Nothing asserts it stays so. A later change that edits `backlogDDL` — for a new
table, a new index, a changed CHECK — leaves the hand-copied replica silently stale: the test keeps
passing against a snapshot of code that no longer exists, while reporting that it exercises "the
pre-change open path". A frozen replica whose divergence from its original is undetectable is the
same failure shape as a guard that has stopped guarding, and it is invisible by construction.
**E2, blocking.** The fix is cheap and does not sacrifice the freeze: assert the frozen copy equals
the live const at test time, so a divergence becomes a deliberate act (a test failure a human then
resolves by updating the replica *and* re-reasoning about downgrade), rather than a silent one.

**Is §G accurate?** Yes, on its own terms. It says clause (b) exercises the reconstructed path,
that the reconstruction "is compiled from today's source, so a divergence between the
reconstruction and a real released binary would be invisible to it", and that REQ-TLE-018 "remains
an argued claim with a partial demonstration … and it stays in this list after AC-TLE-018 passes."
Every clause of that is true and it is the honest position. What it does not say is that the
replica can also diverge from *tomorrow's* source with equal invisibility — a different hazard from
the released-binary one it does name, and the one E2 addresses.

---

## Hunts 3, 5, 6 — briefly

**Hunt 3 — D6's positive controls.** They work as claimed. An exclusion broad enough to hash
nothing fails the non-empty check; one that swallowed `.moai/state/kanban/` fails the containment
check. The cited mutation path is exact (verified above), and the RED now names a plant location
that is both outside the queue directory and outside `.git/` (`.moai/cache/`), so it lands inside
the hashed set — the widening half A paid for is kept and only `.git/` is carved back out. One
residual worth stating rather than filing: the controls constrain the exclusion's *breadth*, not
its *shape* — an exclusion that additionally carved out exactly the mutant's directory would pass
both controls. In practice the run-phase must observe the RED, which self-corrects that case, so
this is a note, not a defect.

**Hunt 5 — the two new requirements.** REQ-TLE-021 is sound and its criterion covers both halves;
I checked the premise it depends on and the two doctrine files are currently **byte-identical in
full** (`diff` returns 0 lines), so byte-identity of the two rows is achievable and does not
collide with template neutrality. REQ-TLE-020 has one coverage hole: its when-clause binds "either
check fails **or cannot be run**", and `design.md` §3 rules that an unrunnable check (no git,
unresolvable ref) is also exit 1 — but AC-TLE-020 tests only (a) reachable, (b) non-existent,
(c) unreachable. The cannot-be-run branch has no criterion. **E3, blocking (minor)** — this is the
"criterion covers only half its requirement" class from iteration 1, recurring in new territory,
and the fix is one added case. AC-TLE-020 is otherwise the strongest new criterion in the set: its
REDs come from ordinary inputs rather than a planted mutant, which §D.2 correctly notes is a
stronger position, and the abbreviated-vs-resolved-SHA assertion is a real trap that a naive
implementation would fall into.

**Hunt 6 — recorded-vs-closed.** Checked each iteration-1 item that was answered with "record it".
D7's §G entry is genuinely rewritten, not softened. D13 was **closed**, not recorded — verified
empirically in Hunt 2. §G gained four residual entries (reachable-but-wrong SHA, guard-sees-tuples-
not-whole-schema, doctrine-prose-uncovered, 21-of-25 budget) and none of the v0.1.0 entries was
quietly dropped: the seventh-column contract change, the stale `spec_status`, the un-enumerable
consumers, the migration-parity-read-not-exercised, and the SQLite-observed-on-one-machine items
all survive verbatim or expanded. The `§D` exclusions (six `### Out of Scope —` H3s) are unchanged
and still carry bullets.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric band | Evidence | Δ from iter 1 |
|---|---|---|---|---|
| Clarity | 0.90 | 0.75 (well above band; the 1.0 band requires no ambiguity anywhere) | All four iteration-1 clarity defects closed; §A.3b, §B.3.1 and §C.7's verified-vs-DoD split are precise and correctly grounded. Residual: E5 (`plan.md` §B numbered 1,2,3,5,4), E6 (AC-TLE-021 cites a value another test measures), E4 (undisclosed path trimming at `research.md:184`) | +0.10 |
| Completeness | 0.90 | 0.75 (well above band) | 21/21, five artifacts, six out-of-scope H3s, §G expanded to four new residuals with none dropped, §D.3 marks its one uncovered line in place. Residual: E3's untested when-branch | +0.05 |
| Testability | 0.85 | 0.75 (above band) | 019a/b/c independence verified empirically; AC-TLE-015 pins all seven fields individually; AC-TLE-014 has working positive controls; AC-TLE-020's REDs are ordinary inputs. Residual: E1 (a clause that cannot fail, on a Blocking-data criterion), E2 (undetectable replica drift), E3 | +0.10 |
| Traceability | 0.90 | 0.75 (well above band) | 21↔21 bidirectional in §E, both new rows mapped (020→M3, 021→M5), 019 shown as (a/b/c), the REQ-TLE-004 milestone cell corrected and additionally disclaimed in `plan.md` M1.4. Residual: REQ-TLE-020's cannot-be-run branch untraced (E3) | +0.10 |

Aggregate (harmonic mean): **0.8870 → 0.89**. Arithmetic mean 0.8875.
**0.89 ≥ 0.85**, so the Tier L threshold is met.

**Score trajectory: 0.80 → 0.89.** No regression, so the LEAN STOP-on-regression clause does not
fire and no scope-reduction question is warranted.

---

## Defects Found (iteration 2)

**E1** — `.moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/acceptance.md:L323-325` (AC-TLE-020,
attribution-boundary clause; its defence at `:334`) — The clause asserts the boundary by "running (a) with the card
**renamed** between two invocations", but the predicate it is guarding against keys on the card
**id**, not its text: `LandedGrepArgs` builds `--grep=\b` + `cardID` + `\b`
(`internal/kanban/prlink_landed.go:96-108`), and no `moai todo` verb changes a card's id (16 verbs
registered at `internal/cli/todo.go:148-152`; `edit` changes text only). An implementation that
leaked the card token into validation would return the same result before and after a text rename,
so the assertion passes with the leak intact. The criterion's own stated intent — "independent of
**which card** the record belongs to" — names the correct variable, which the test then does not
vary. — Severity: **major** — Class: **blocking** — Required fix: replace the rename with a
two-card comparison — record the same `--sha` against two different card ids and assert an
identical accept/reject outcome and identical stderr classification. Fix in the SPEC before
run-phase reaches M3; do not defer it to implementation.

**E2** — `acceptance.md:L250-273` (AC-TLE-018 clause (b)) — The "frozen test-local copy" of
`backlogDDL`, the `schemaVersion` read, and the version switch has no drift detector. `design.md:57-60`
rules the DDL const is not edited and REQ-TLE-018 forbids the version bump, so the replica is today
byte-identical to the live code; nothing asserts it remains so. A later edit to `backlogDDL` leaves
the replica silently stale while the test keeps passing and keeps reporting that it exercises "the
pre-change open path". §G names the released-binary divergence but not this one.
— Severity: **major** — Class: **blocking** — Required fix: assert in the test that the frozen copy
equals the live `backlogDDL` const (and the live switch's accepted-version set), so a divergence
fails loudly and becomes a deliberate act; add a sentence to §G naming forward drift as a distinct
hazard from the released-binary one.

**E3** — `spec.md:L462-468` (REQ-TLE-020) vs `acceptance.md:L313-334` (AC-TLE-020) — The
requirement binds "**when** either check fails **or cannot be run**, the verb shall write nothing
and exit 1", and `design.md` §3 rules explicitly that no-git or an unresolvable ref is also exit 1.
AC-TLE-020's Given supplies only (a) reachable, (b) non-existent, (c) unreachable. The
cannot-be-run branch — the one that decides behaviour on a machine without git, which is a real
operating condition this codebase otherwise handles by degrading (`todo_pr.go` fail-open) — has no
criterion. — Severity: **minor** — Class: **blocking** — Required fix: add case (d) to AC-TLE-020's
Given (git unavailable or ref unresolvable) with the same Then: exit 1, `SELECT landing IS NULL`
returns 1, stderr names the unrunnable check.

**E4** — `research.md:L184-196` (§R.10.1) and `spec.md:L120-136` (§A.3b) — The `$ grep -m1
'^status:' …` block presents four output lines with the `.moai/specs/` prefix stripped from every
path, under a shell prompt and without disclosure. I re-ran the command: the real output is
`.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:status: in-progress` and so on. The statuses and the
surprising non-argument ordering are both reproduced faithfully — only the paths were trimmed —
so the substance is correct. But this is the same class as iteration-1's D11, which this revision
fixed in §R.8, and §R.10.2 shows the author knows the convention ("Comment lines elided for width").
— Severity: **minor** — Class: **non-blocking** — Required fix: re-paste the block with full paths,
or add the one-line elision disclosure §R.10.2 already uses.

**E5** — `plan.md` §B Known issues — The list is numbered 1, 2, 3, **5**, **4**: the new git-subprocess
item was inserted as item 5 ahead of the existing `--json` item 4. Purely cosmetic; it does make the
§B cross-references ambiguous if anything later cites "§B.4". — Severity: **minor** —
Class: **non-blocking** — Required fix: renumber the inserted item to 4 and the `--json` item to 5.

**E6** — `acceptance.md:L343` (AC-TLE-021) — "the stated column count equals **the field count
AC-TLE-015 measures** on the rendered row (7)". A test cannot read another test's runtime value; it
must either re-render and count, or compare against the literal 7. The requirement (REQ-TLE-021)
says "the count the rendered row actually carries", which is the stronger reading — but the
criterion does not say which of the two it does, and the parenthetical `(7)` invites the weaker
one. If it compares prose to a literal, the criterion is a doc-consistency check rather than the
prose-versus-behaviour tie the surrounding paragraph claims. — Severity: **minor** —
Class: **non-blocking** — Required fix: state that AC-TLE-021 renders the row itself and counts its
fields, and drop or demote the `(7)` to a parenthetical note of today's value.

**E7** — `spec.md:L661-670` (§G, the reachable-but-wrong-SHA residual) — The residual states the
cost as "nothing flags it; and it is corrected only when a human notices". The machine-detection
claim is correct and correctly reasoned. But the SPEC does not consider the cheapest available
mitigation: rendering the named commit's **own subject line** beside the SHA on the `todo pr` row.
That display makes no card-to-commit attribution claim — it is the same class of non-attributing
fact as `ref_head` — and it is exactly what would let a human notice a typo. — Severity: **minor**
— Class: **non-blocking** — Required fix: none required; either add the subject to the render
(a small extension to REQ-TLE-015/016's cell format), or amend §G to record that this mitigation
was considered and declined, so the next reader does not re-derive it.

---

## Recommendation

**PASS-WITH-DEBT at 0.89** against the Tier L threshold of 0.85. All seven must-pass criteria pass;
all thirteen iteration-1 defects are resolved; the score improved by 0.09 with no regression, so no
STOP signal and no scope-reduction question. The revision is unusually disciplined: it withdrew two
false claims in terms rather than softening them, closed D13 by strengthening a requirement rather
than by recording a residual, and put the guard-ownership decay note inside the criterion the
run-phase actually reads.

The debt is three blocking defects, none of which is a must-pass failure and none of which requires
rethinking a decision:

1. **E1 — fix before run-phase reaches M3.** This is the one that matters. AC-TLE-020 is the
   criterion protecting the SPEC's most load-bearing boundary, and its boundary clause currently
   cannot fail. A one-line substitution (two card ids instead of one renamed card) makes it real.
2. **E2 — fix in the SPEC now.** One added assertion converts an invisible drift into a loud one.
3. **E3 — fix in the SPEC now.** One added case closes half a when-clause.

The four non-blocking findings (E4-E7) are surfaced for the orchestrator's discretion and should
not by themselves drive an iteration 3. Under the M6 finding-consumption rule, routing them into a
revision round would buy less than it costs.

Proceeding to the Implementation Kickoff Approval gate is appropriate once E1, E2, and E3 are
closed — they are SPEC edits, not implementation work, and all three are small enough that a
confirming re-audit can be scoped to the enumerated delta rather than a fresh full pass.

---

## Gaps — what this audit did NOT observe

Stated explicitly so the empty set is not implied.

- **No run-phase behaviour was verified.** `moai todo landed`, `internal/kanban/landing_evidence.go`,
  and the extended guard remain unwritten. Every judgment about them is a judgment about the
  *criterion*.
- **I did not execute `git rev-parse --verify` or `git merge-base --is-ancestor` against a
  fixture.** REQ-TLE-020's commands are judged from their documented semantics and from the
  measured fact that neither takes a card id. The unreachable-commit fixture is AC-TLE-020's to
  build in run-phase, as `research.md` §R.10.5 records.
- **I did not re-run the 019a `items` plant in this iteration.** I reproduced it in iteration 1 at
  HEAD `b2d30deb2`; this iteration's probe covered 019b and 019c, which were the new claims.
- **I did not re-run `go test ./internal/cli/...` or `./internal/kanban/...` in full.** My probe was
  scoped to `-run 'TestAudit2_'`. The pre-edit baseline over both packages is `plan.md` §C's
  run-phase-entry obligation; `go test ./...` is prohibited on this machine and was not attempted.
- **I did not re-verify the ~30 `file:line` pins carried over unchanged from v0.1.0.** I verified
  every pin that is new or changed in v0.2.0 (`prlink.go:101-114`, `prlink_landed.go:96-108` and
  `:150-154`, `todo_pr.go:57-60`, `todo_test.go:451-480`, `backlog_sqlite.go:292-298`,
  `design.md:57-60`) plus the two doctrine files. The carried-over pins were verified in
  iteration 1 against the same source tree, which has not moved.
- **I did not measure whether `.git` in a git *worktree* (a file, not a directory) would defeat
  AC-TLE-014's `.git/` exclusion.** The criterion's Given has the test create its own repository,
  so a real `.git/` directory is the expected fixture shape; the worktree case is out of the
  criterion's scope as written.
- **No cross-model backend was consulted.** `mcp__moai__audit_multi` / `codex_audit` / `glm_audit`
  were not invoked; this is a Claude-anchor-only verdict.

## Residual risk

- **E1's fix is small enough to look done without being done.** Substituting "two cards" for
  "renamed card" changes one sentence; the confirming re-audit must check that the fixture actually
  uses two distinct card *ids* and that both calls supply the same `--sha`, or the clause stays
  vacuous under new wording.
- **Two blocking defects this iteration are of the same class as iteration 1's** (a clause that
  cannot fail; a criterion covering half its requirement). The class is recurring in new text as
  fast as it is closed in old text, which suggests the risk sits in how new criteria are drafted
  rather than in any particular criterion. A third iteration, if one happens, should sample the
  newest criteria first.
- **My Hunt 2 measurement decays.** It was taken at HEAD `483cea858` in this worktree, and the
  guard belongs to `SPEC-TODO-ARCHIVE-QUERY-001`. The SPEC's own decay note is correct and the
  run-phase must re-measure rather than cite either observation.
- **A PASS-WITH-DEBT verdict is not a licence to skip the Implementation Kickoff Approval gate.**
  That gate is score-independent; this verdict is an input to it, never a substitute for it.
