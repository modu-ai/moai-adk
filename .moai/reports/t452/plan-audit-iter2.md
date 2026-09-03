# SPEC Review Report: SPEC-CODEX-SKILL-LOADER-001 (card t452)

Iteration: 2/2 (Tier M ceiling — this is the final iteration)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.86** — Tier M PASS threshold 0.80. Score gap from iter-1 (0.79) is **closed**;
two blocking-class findings remain and are enumerated as the debt.

Reasoning context ignored per M1 Context Isolation.

## Hashes read

Verified before scoring; all four match the frozen set the lead named, so no writer was active:

| artifact | sha256 |
|---|---|
| spec.md | `8b7425bbad6ca0597902eb48e30b87955047bdc10cc24146eb2089cb0238b796` |
| plan.md | `8b192b7b79f9dab3e3de4bd37d903d6114f48d03226fb6825ac4676d618996d8` |
| acceptance.md | `70256ddfa41031f93c37ccd58ae89e9955b2c6b8d1397aaa89abc32b77a21520` |
| progress.md | `f8b7844e27cb8c0a07c8ff557586dd826dbe2512986f94df18cf53541abb0ef9` (unchanged since iter-1) |

Per the Retry Loop Contract this is a delta re-audit: the iter-1 defect list plus a regression check,
plus a search for defects planted by the fix pass.

## Must-Pass Results (re-measured at the frozen bytes)

- **[PASS] MP-1** — `grep -oE 'REQ-CSL-[0-9]{3}' spec.md | sort -u | wc -l` → 13;
  `grep -cE '^- \*\*REQ-CSL-'` → 13. Contiguous 001..013, no renumbering from iter-1.
- **[PASS] MP-2** — the 13 `REQ-CSL-*` entries in spec.md §C all match a GEARS pattern. The three
  rewritten ones keep their modality: REQ-CSL-001 `While` (spec.md:134), REQ-CSL-005 unwanted
  (`…해서는 안 된다`, :138), REQ-CSL-008 `When` (:144). Requirement layer only.
- **[PASS] MP-3** — 12 canonical fields + `tier: M`; `grep -cE '^(id|title|…|tier):'` → 13.
- **[N/A] MP-4** — single-domain SPEC; no multi-language tooling surface.
- **[PASS] MP-5 D7** — referenced SPECs unchanged and all `completed`; no BLOCKING.
- **[PASS] MP-6 D8** — `grep -c syscall` → 0 in all three artifacts. Auto-PASS.
- **[PASS] MP-7** — no `[NEEDS CLARIFICATION` markers.

Structural invariants also re-measured: AC ids 13 unique / 13 headings; 4 `### Out of Scope —`
sub-headings; `grep -c 'origin/main\|origin/develop'` → 0 in all three artifacts (moving-ref
hygiene still clean); `moai spec lint` → `0 error(s), 13 warning(s)`, all the same
`CoverageIncomplete` class whose linter-limitation status I established with a control group in
iter-1.

## Regression check — iter-1 defects

| iter-1 | Status | Evidence |
|---|---|---|
| D1 (AC-CSL-005 impossible) | **RESOLVED** | acceptance.md:43 |
| D2 (vacuous-PASS mandate) | **RESOLVED** | acceptance.md:91, :98, :106-107 |
| D3 (roots do not discriminate) | **RESOLVED, with residual** → N2 | plan.md:52-82 |
| D4 (no run-provenance) | **RESOLVED** | acceptance.md:15, :30-31; spec.md:134 |
| D5 (absence claim wider than selector) | **RESOLVED** | spec.md:45 |
| D6 (the `11` literal) | **RESOLVED** | acceptance.md:53; spec.md:144; plan.md:7 |
| D7 (Tier asymmetry) | **RECORDED as agreed** | plan.md:22 |

**D1 — closing the impossible direction did not open the vacuous one.** This was the specific
failure mode to check, and the fix handles it explicitly: acceptance.md:45 pins an empty introduced-key
set to **해당 없음, not PASS**, in its own sentence ("스윕한 것이 없는 초록은 아무것도 주장하지 않는다").
:46 names the 7 excluded keys individually and cites where their measurement lives, and REQ-CSL-005
(spec.md:138) was narrowed to match. I verified the cited record exists rather than accepting it:
`internal/template/agentemit/agents-codex.yaml:27` carries the `fields:` section and `:88` the
`classes:` section, with `P-01`/`P-02`/`P-03`/`P-04 MEASURED` rationale at `:17`, `:28`, `:35`, `:46`.
The citation is real. (One consequence of that record's version scope is N3 below.)

**D2 — both places agree, and the canonical-on-divergence clause exists.** acceptance.md:106 gives the
branch-B row as `AC-CSL-010·011 = PASS / AC-CSL-012·013 = 해당 없음`; :107 forbids writing PASS there
and states "두 자리가 어긋나면 AC 본문이 정본이다" — the canonical-on-divergence clause the lead asked me to
confirm. The AC bodies carry the same conclusion independently: :90-91 (012: swept files ≥1 plus a
control, else 해당 없음) and :97-98 (013: at least one of the two target sets non-empty, else 해당 없음).
No divergence between the two places today.

**D3 — the five roots do genuinely discriminate now.** plan.md:56-60 lists five roots each with a
distinct marker name; `.claude/skills/` is added as R5 and plan.md:62 states exactly why its absence
was the attribution hole ("R5 를 빼면 R1 의 양성을 R1 에 귀속시킬 수 없다"). Rule 2 (:66) mandates one root
populated per run with per-root distinct names; the defective pre-built fixture is banned at :68 and
again at plan.md:27, with the ban's reason ("판별 불가") correctly separated from the filesystem
link-traversal fact. Rule 3 (:70-78) requires the signal be chosen and recorded before the first run.
Rule 4 (:80-82) defines the no-marker negative control and its consequence. The residual is N2.

## Defects Found (this iteration)

**N1 — `git diff` with no base ref: the D1 fix's own selector goes empty once the work is committed — `acceptance.md:43`, also `:89` and `:82` — Severity: critical — Class: blocking. NEW, planted by this fix pass.**
AC-CSL-005's `When` is `git diff -- internal/template/templates/.codex/agents/moai/ | grep '^+[a-z_]* *=' | …`.
Bare `git diff` compares the **working tree to the index**. REQ-CSL-008 (`spec.md:144`) has the run-phase
regenerate and commit, and acceptance is evaluated at completion — at which point that command prints
nothing. It also prints nothing for changes merely staged. An empty result is then routed by
acceptance.md:45 to **해당 없음** — so a run that *did* introduce a new `skills` field records "this SPEC
introduced no new field". The D1 fix's own vacuity guard converts the selector's blindness into a
confident wrong answer, which is worse than the impossible direction it replaced.
The same missing base afflicts **AC-CSL-012:89** (`git diff --name-only -- internal/template/templates/`
→ swept count 0 → 해당 없음 on the branch where it should be PASS) and, in prose form,
**AC-CSL-011:82** ("이 SPEC 의 변경 집합" names no base at all).
This is the `verification-claim-integrity.md` §2.1 shape: a claim decided against an unstated,
moving coordinate. Its own R2 remedy applies directly.
**Required fix:** capture `BASELINE_SHA=$(git rev-parse HEAD)` at §C pre-flight, before the first
run-phase commit, record the resolved value in `progress.md`, and decide all three criteria against
`git diff $BASELINE_SHA -- <path>`. (Minor, same edit: `[a-z_]*` will not match a key containing a
digit; widen to `[a-z0-9_]*`.)

**N2 — the negative control catches only false positives; a signal that never fires reads as branch B — `plan.md:80-82`, `acceptance.md:16` — Severity: major — Class: blocking. Residual of D3.**
Rule 4 is sound in the direction it covers: signal present with no marker ⇒ signal invalid ⇒ reselect
and re-run. It has no counterpart in the other direction. If the chosen signal simply does not work —
and plan.md:76 flags S3 as exactly that risk, "존재 여부 자체가 미관측" — then every one of the five roots
reads 미적재, the negative control passes (no marker, no signal: as expected), and M0 records **branch B**
on a broken instrument. Branch B is the leading outcome (spec.md §A.5, plan.md:15), so this failure mode
produces the *expected* answer, which is the hardest kind to notice. The instrument's red has been
observed; its green never has (`verification-completeness.md` §1.1).
A positive control is available and cheap: R2 (`$CODEX_HOME/skills`) is the one root codex's **own
installer** documents (spec.md §A.5 P3). It is not a guarantee, but it is the strongest prior in the
candidate set.
**Required fix:** add Rule 5 — if R2 also reads 미적재, the signal is suspect and MUST be reselected
before a branch is recorded; a branch-B verdict in which no root produced the signal is recorded as
**inconclusive**, not as branch B. Bound the reselection at 3 (S1/S2/S3 exhausted ⇒ blocker report),
since Rule 4 currently allows unbounded reselect-and-rerun.

**N3 — `codex_measured_version` restamped to 0.152.1 over field semantics measured at 0.147.0 — `spec.md:140` (REQ-CSL-009), `acceptance.md:68-72`, against `agents-codex.yaml:13` — Severity: major — Class: optional (branch-A only).**
The manifest states its own field's meaning at `agents-codex.yaml:13`: "codex-cli version **the field
semantics below** were measured against", and the preamble at `:7-9` explains the purpose — "so a future
upgrade knows which probe run to repeat". Those semantics are the `P-01`..`P-04 MEASURED` rationale,
all recorded at 0.147.0 (`:17`, `:28`, `:35`, `:46`). REQ-CSL-009 restamps the field to this round's
0.152.1 while the D1 fix (`acceptance.md:46`) states in writing that the 7 prior fields are **not**
re-judged. After a branch-A run the manifest therefore asserts a coverage it does not have, and the
next reader loses the very signal the field exists to carry. The two edits are individually correct
and jointly produce this; iter-1 did not catch it because the non-re-judgment was implicit then.
**Required fix:** scope the stamp rather than widening it — either add a sibling
`skills_field_measured_version` for what this SPEC measures, or leave `codex_measured_version` at
0.147.0 and record 0.152.1 in the `skill-loader` row's own rationale. Do not restamp the shared field
without re-measuring what it covers.

## Judgment the lead asked for: 012/013 = 해당 없음 on the leading branch

**Acceptable, and the alternative is worse — not a new defect.** The constraints are not weakened by
being unjudged there; their **subject cannot exist** on that path. Branch B's REQ-CSL-003 forbids
touching `agents-codex.yaml`, so nothing lands under `internal/template/templates/` for neutrality to
be about, and no Go code is added for the `os.Stat` clause to read. Unreachability is a stronger
guarantee than a passing check, and recording PASS over an empty set would be precisely the
unobserved-verification claim D2 was raised to remove.

The residual worth naming — that a branch-B record would show no trace the two constraints were
considered — **is already closed** by acceptance.md:104, which requires every one of the 13 to be
recorded as PASS / FAIL / 해당 없음 and treats non-recording as FAIL. The 해당 없음 is therefore written
down with the rest, not silently dropped. No action needed.

## HARD constraint traceability — all five still hold

| constraint | requirement | judging AC | branch-B status |
|---|---|---|---|
| (a) no codex flow in this repo; `/tmp` + isolated `CODEX_HOME` | REQ-CSL-010 | AC-CSL-010 | PASS (probe alone gives it a subject) |
| (b) honour `CODEX_HOME`, no home hardcode, no `os.Stat` collapse | REQ-CSL-013 | AC-CSL-013 | 해당 없음 (judged above as correct) |
| (c) where/when/whose-consent writes | REQ-CSL-010 + 011, §B D5 | AC-CSL-010 + 011 | PASS |
| (d) Template-First: `make build` + `make agents-emit` | REQ-CSL-008 | AC-CSL-008 | 해당 없음 (no manifest edit on B) |
| (e) template content neutrality | REQ-CSL-012 | AC-CSL-012 | 해당 없음 (judged above as correct) |

## Category Scores

| Dimension | Score | iter-1 | Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.90 | 0.80 | 1.0 | M0's four rules are stated as rules with their reasons (plan.md:50-82); the canonical-on-divergence clause removes the one place two documents could disagree (acceptance.md:107). |
| Completeness | 0.85 | 0.75 | 1.0 | Fifth root added with its attribution argument (plan.md:60-62); signal candidates enumerated; negative control defined; Tier asymmetry recorded (plan.md:22). Deduction: no positive control (N2). |
| Testability | 0.70 | 0.60 | 0.75 | D1 closed without opening vacuity; D2 closed in both places; five recorded elements per probe run (acceptance.md:15). Deductions: N1 (unstated diff base defeats the D1 fix) and N2 (instrument's green never observed). |
| Traceability | 1.00 | 1.00 | 1.0 | 13 REQ ↔ 13 AC bijection intact through the rewrite; §E table unchanged; all five HARD constraints trace to a requirement and a judging AC. |

Aggregate **0.8625 → 0.86**, above the Tier M threshold of 0.80.

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria clear, the aggregate clears the Tier M threshold, and
every iter-1 defect is resolved — I verified each against the frozen bytes rather than against the
change report, including the one cited record (`agents-codex.yaml` `fields:`/`classes:` rationale) that
the D1 fix depends on existing. No defect from iter-1 recurs, so there is no stagnation.

The debt is two items, both cheap and both **before run-phase M0 begins**, since each governs how M0 is
run or recorded:

1. **N1** — freeze `BASELINE_SHA` at §C pre-flight and decide AC-CSL-005 / 012 / 011 against
   `git diff $BASELINE_SHA`. Without this the D1 fix reports 해당 없음 for work it was written to catch.
2. **N2** — add the R2 positive control and the inconclusive verdict, and bound signal reselection at 3.
   Without this a broken signal produces branch B, which is the answer the SPEC already expects.

**N3** is branch-A-only and can be taken with the M3 edit rather than before M0.

This is iteration 2 of the Tier M ceiling of 2, so there is no iteration 3 to schedule. Route the two
debt items as a pre-M0 fix rather than a re-audit: neither changes a requirement, a criterion's
identity, or the branch structure, so re-auditing the whole SPEC would cost more than it returns.
