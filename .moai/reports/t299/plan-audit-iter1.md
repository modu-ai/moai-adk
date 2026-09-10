# SPEC Review Report: SPEC-SYNC-SHA-SLOT-FORMAT-001

Iteration: 1/1 (Tier S ceiling)
Verdict: **FAIL**
Overall Score: **0.80**
Audited in: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t299`, branch `WT-sha-slot-format`, HEAD `a6bbbf82b`
Corpus state at audit: only untracked path under `.moai/specs` is this SPEC's own directory
(`git status --porcelain .moai/specs` → `?? .moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/`), so the
corpus did not move between authoring and audit.

Reasoning context ignored per M1 Context Isolation. The dispatch's framings were treated as claims to
test, not as inputs.

The score alone clears the Tier S PASS threshold (0.75). The FAIL is driven by **D1**, an internal
contradiction between AC-SSF-006 and REQ-SSF-005 that makes a MUST acceptance criterion unsatisfiable
by a correct implementation and voids the §D.3 severity rationale. No M5 must-pass criterion failed.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE 'REQ-SSF-[0-9]+' spec.md | sort -u` → REQ-SSF-001
  … 008, sequential, no gaps, no duplicates, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-SSF-*` in
  `spec.md` §C), not the verification layer. 001/002 ubiquitous (`The system shall define/expose`),
  003/004/006 event-driven (`When …, the … shall`), 005 state-driven (`While a slot holds …, the lint
  shall not`), 007/008 unwanted (`shall not be added` / `shall not alter`). The `Given/When/Then`
  entries in `acceptance.md` are the correct verification-layer format and are graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types:
  `id`, `title`, `version: "0.1.0"` (quoted), `status: draft`, `created: 2026-08-29`,
  `updated: 2026-08-29`, `author`, `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`,
  `tags` (comma-separated string). No rejected snake_case alias (`created_at` / `updated_at` /
  `labels` / `spec_id`) present. Extras (`era`, `tier`, `related_specs`) are additive.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own Go internals
  (`internal/spec`), a single-language surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — referenced IDs resolved and status-read:
  `grep -m1 -H '^status:' .moai/specs/SPEC-MOVING-REF-GUARD-001/spec.md .moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001/spec.md`
  → `in-progress` / `completed`. `SPEC-BACKLOG-LOCK-BUDGET-001` → `implemented`,
  `SPEC-V3R6-AUDIT-MODEL-PIN-001` → `implemented`, `SPEC-V3R6-SESSION-LEGACY-COVERAGE-001` →
  `completed`. None is `retired` / `superseded` / `archived`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md plan.md acceptance.md` →
  `0` on all three. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -c 'NEEDS CLARIFICATION' plan.md` → `0`. No
  `research.md` exists (Tier S).

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | Every REQ has a single reading; §D.1's grammar is stated as a formal production. Deduction: three REQs name internal identifiers rather than behaviour (D7), and §A/§acceptance disagree on the slot's section number (D9). |
| Completeness | 0.75 | 0.75 | All required sections present; three `### Out of Scope — <topic>` H3 sub-headings each with specific bullets (`spec.md` §E). Deductions: REQ-SSF-007 carries no AC (D3), the `mx_commit_sha` inheritance is named but never measured (D4), and §B.6's derivation is incomplete — it never intersects the flagged set with REQ-SSF-005's exemption (D1). |
| Testability | 0.80 | 0.75-1.0 | Unusually strong: every AC names a deciding command **and** a planted mutation, and the AC-SSF-001/002/003 triad genuinely excludes the switched-off rule (verified by construction below). Deduction: AC-SSF-006 is a MUST criterion whose stated numbers a correct implementation cannot produce (D1). |
| Traceability | 0.80 | 0.75 | REQ-001→AC-001/002/003, 002→AC-007, 003→AC-004, 004→AC-001, 005→AC-003, 006→AC-002, 008→AC-009. Deductions: REQ-SSF-007 uncovered (D3); AC-SSF-008 is an orphan with no REQ (D8). |

Arithmetic mean 0.80.

---

## What reproduced (measured, not assumed)

Every figure below was re-derived in this tree at `a6bbbf82b` with the SPEC's own cited command.

| Claim | Cited | Re-derived | Command |
|---|---|---|---|
| §B.2 r1 total slots | 346 | **346** | `grep -h '^sync_commit_sha:' .moai/specs/*/progress.md \| wc -l` |
| §B.2 r2 SHA-token | 334 | **334** | above `\| sed 's/^sync_commit_sha:[[:space:]]*//' \| grep -cE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]\|$)'` |
| §B.2 r3 non-SHA | 12 | **12**, in **11** SPECs (`SPEC-V3R6-SESSION-LEGACY-COVERAGE-001` holds two) | same, inverted |
| §B.2 r4 `#`-annotated | 105 | **105** | same `\| grep -c '#'` |
| §B.2 prose: three non-`#` tails | 3 | **3** (`SPEC-SESSIONSTART-PERF-001`, `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001`, `SPEC-ZONE-REGISTRY-RESYNC-001`) | `python3 .moai/reports/t299/tails.py` |
| §B.3 distinct spellings | 28 | **28** | `grep -rh 'pending-backfill' .moai/specs/*/progress.md \| grep -oE 'pending-backfill[a-zA-Z0-9-]*' \| sort -u \| wc -l` |
| §B.3 `pending-backfill-sync` | 24 | **24** | same, `uniq -c` |
| §B.3 `pending-backfill` | 29 | **31** as cited → see **D5** (29 excluding this SPEC's own self-matching progress.md) | same |
| §B.3 the allowlist | four values | **verified verbatim** — `internal/spec/closer.go:397`, `switch v { case "", "(this commit)", "(pending)", "<pending>": return true }` | `sed -n '397,405p' internal/spec/closer.go` |
| §B.3 D3 prescribes `pending-backfill-*` | yes | **verified** — `spec-frontmatter-schema.md:140` carries the exemption and the literal `pending-backfill-*` | `grep -n 'SHA placeholder backfill' .claude/rules/moai/development/spec-frontmatter-schema.md` |
| §B.4 four readers | era.go:139, era.go:252, audit.go:360, closer.go:618, status.go:261 | **all five line citations exact**; `cleanFieldValue` set is `"" / null / none / tbd / <pending> / pending` as stated | `sed -n` at each |
| §B.4 live parsing artifact | stray quote + prose in the drift detail | **reproduced verbatim** — `mcp__moai__spec_audit(filter_spec=SPEC-V3R6-AUDIT-MODEL-PIN-001, project_root=<this worktree>)` returns `"sync_commit_sha":"pending-backfill-sync\"   # D3 self-reference exemption — …"` | MCP call |
| §B.5 status table | 9 completed / 2 implemented | **9 / 2**, and the two `implemented` are exactly `SPEC-BACKLOG-LOCK-BUDGET-001` and `SPEC-V3R6-AUDIT-MODEL-PIN-001` | `grep -m1 -H '^status:'` over the 11 owning `spec.md` |
| §B.5 both implemented classify V3R6 | H-4 | **both** return `"era":"V3R6","heuristic_matched":"H-4 (§E.2 + §E.4 + sync_commit_sha)"` | two `mcp__moai__spec_audit` calls with `project_root` |
| §B.6 demotion attachment derivation | sibling-file finding is demoted by the owning spec.md | **CORRECT** — `lint.go:220` computes `demote` per document and `lint.go:221` applies it to `docFindings`, the whole batch, with no reference to `Finding.File`; `applyEraDemotion` (277) sets `Advisory` on every warning at 288; `HasErrors` escalates only `!f.Advisory` at 61; `terminalStatusEnum` (1134) contains `completed`, not `implemented` | `sed -n '205,300p' internal/spec/lint.go`, `sed -n '1128,1142p'` |
| §D.4 t354 remediation string | `moai spec close SPEC-BACKLOG-LOCK-BUDGET-001 --backfill-only` | **verified verbatim** in the audit output | MCP call |
| §F byte-identity of schema + mirror | identical, `diff` exit 0 | **verified**, exit 0, both **24677** bytes | `diff .claude/rules/moai/development/spec-frontmatter-schema.md internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md; echo $?` + `wc -c` |

**§D.1 grammar tested against the corpus, not reasoned about.** A classifier implementing §D.1
exactly (token = first whitespace-delimited run; one leading and one trailing quote stripped from the
token; `^[0-9a-fA-F]{7,40}$`; `^pending-backfill(-[A-Za-z0-9-]+)?$`) was run over all 346 slots
(`python3 .moai/reports/t299/grammar_check.py`):

```
total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
```

- **Admit side: no misclassification.** All 334 values the SPEC counts as conforming are admitted by
  the token rule; the three non-`#` tails (parenthetical, em-dash) are admitted, so the first-token
  choice over a separator enumeration is vindicated by measurement.
- **Reject side: the grammar itself is sound** — the 5 flagged tokens are `null`, `<pending>`, `TBD`,
  `null`, `pending`, all genuinely prose-occupied.
- **L1 has no live instance**: `^[a-fA-F]{7,40}$` matches **0** of the 334 tokens, so the accepted
  hex-word false-negative class is currently empty. Recording it as a known limitation is honest, not
  speculative.

**The falsifiability contract holds where the lead asked it to.** The AC-SSF-001/002/003 triad does
mechanically exclude a switched-off rule: (a) demands one finding on prose, (b) demands zero on four
SHA shapes, (c) demands zero on two placeholder shapes. A rule flagging everything fails (b)+(c); a
rule flagging nothing fails (a). Each leg carries a named mutation that must turn it red. This part of
the SPEC is better than the norm and should survive the revision unchanged.

---

## Defects Found

**D1. AC-SSF-006 asserts numbers that REQ-SSF-005 forbids — the corpus yields 5 findings and 0
non-advisory, not 12 and 2.**
`acceptance.md` AC-SSF-006; `spec.md` §B.6, §D.3; `plan.md` §C.3, §F M4 — **Severity: critical** —
**Class: blocking**.

§B.6 derives its enforcement cost from the 12 non-SHA values of §B.2 and their owning statuses (§B.5:
9 `completed` + 2 `implemented`), concluding "Predicted non-advisory findings on the corpus: 2".
That derivation silently assumes all 12 become findings. REQ-SSF-005 says the opposite: a recognized
backfill placeholder is **not** flagged. Seven of the twelve are exactly that.

Establishing command — the §D.1 classifier plus the status grep:

```
$ python3 .moai/reports/t299/grammar_check.py
total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
--- FLAGGED ---
SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md 47   token='null'
SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md 259  token='<pending>'
SPEC-V3R6-SPEC-ID-VALIDATION-001/progress.md     103   token='TBD'
SPEC-V3R6-SPEC-LINT-CLEANUP-001/progress.md       45   token='null'
SPEC-V3R6-TEST-REFACTOR-001/progress.md          149   token='pending'
--- PLACEHOLDER (admitted by REQ-SSF-005) ---
SPEC-AUDIT-SNAPSHOT-001, SPEC-BACKLOG-LOCK-BUDGET-001, SPEC-INFINITE-GOAL-001,
SPEC-UPDATE-YAML-PRESERVE-001, SPEC-V3R6-AUDIT-MODEL-PIN-001,
SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001, SPEC-VERIFICATION-COMPLETENESS-001

$ grep -m1 -H '^status:' .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/spec.md \
    .moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/spec.md \
    .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md \
    .moai/specs/SPEC-V3R6-TEST-REFACTOR-001/spec.md
→ completed ×4
```

All five flagged slots belong to `completed` SPECs, and `terminalStatusEnum` (`lint.go:1134`)
contains `completed`, so `applyEraDemotion` marks every one of them `Advisory` and `HasErrors`
(`lint.go:61`) escalates none of them. **Correct prediction: 12 → 5 total findings, 2 → 0
non-advisory.**

The two SPECs §B.6 nominates as the non-advisory pair are precisely the two the rule must stay silent
about: both hold `pending-backfill-sync`. The prediction and the exemption cancel each other exactly.

**Required fix:** recompute §B.6 against REQ-SSF-005's exemption; restate the prediction as *5 total,
0 non-advisory*, naming the five files and lines; update AC-SSF-006's expected values to 5 and 0 and
its "Fails when" clause (a rule attaching to the wrong document would yield 5 non-advisory, not 12);
update `plan.md` §C.3 and §F M4 (M4 says "Record the twelve findings' locations" — there are five
findings and twelve values, and the inventory of the other seven is a separate, still-useful list).

---

**D2. §D.3's severity rationale is self-refuting.**
`spec.md` §D.3 — **Severity: major** — **Class: blocking**.

The rationale reads: "`warning`, on the §B.6 derivation: the predicted non-advisory count is 2, which
is a number a lane can close, and both are `implemented` SPECs whose slots are genuinely mid-backfill."
A slot that is *genuinely mid-backfill* is, by REQ-SSF-005, not flagged at all — the sentence names
its two exemplars and both are exempt. The severity decision (`warning`) is very likely still right,
and D1's corrected figure makes it *more* affordable, not less; but the argument as written cannot be
the reason.

**Establishing command:** the D1 block above (the two `implemented` SPECs appear in the PLACEHOLDER
list, not the FLAGGED list).

**Required fix:** rewrite §D.3's first paragraph against the corrected derivation. The honest form is
"0 non-advisory findings, 5 advisory, all on closed history" — which strengthens the case for
`warning` while removing the false claim that two open cards are in the strict path.

---

**D3. REQ-SSF-007 has no acceptance criterion, and its justification is conditional on an
interaction the SPEC itself records as unverified.**
`spec.md` REQ-SSF-007, §D.3; `plan.md` §C.4 — **Severity: major** — **Class: blocking**.

Two distinct problems in one requirement.

*Traceability:* no AC asserts that `SyncSHASlotFormat` is absent from `eraDemotableCodes`. The
property is a normative REQ and is decidable in one line
(`grep -A6 'var eraDemotableCodes' internal/spec/lint.go`), so its absence from `acceptance.md` is a
gap, not a judgement call. Establishing command:

```
$ grep -oE 'REQ-SSF-00[0-9]' acceptance.md | sort -u   → (REQ tokens do not appear in acceptance.md at all;
                                                          coverage was mapped by reading each AC)
```
— REQ-SSF-007 is the only requirement with no criterion mapping to it.

*Premise:* REQ-SSF-007's stated reason is that the entry "would be inert", because `eraDemotableCodes`
demotes **errors** only (verified: `lint.go:284` matches `f.Severity == SeverityError &&
eraDemotableCodes[f.Code]`) and a warning is already made advisory by the branch at `lint.go:288`.
That is true **at `a6bbbf82b`**. `plan.md` §C.4 records that card t357 M2 "raises `--strict` behavior
by promoting warnings to errors" and that this is carried from the dispatch, **not** independently
verified (no t357 SPEC directory exists in this tree — confirmed: `ls .moai/specs | grep -i strict` →
no match). If t357's promotion operates on `Finding.Severity` rather than on `Report.HasErrors`, then
after t357 lands all five findings are `SeverityError`, `eraDemotableCodes` becomes the *only*
applicable shelter, and REQ-SSF-007 has forbidden it — turning five findings on closed history into
hard `--strict` errors, which is the exact outcome §D.3 says the rule exists to avoid, and which makes
`lint.skip` the rational response.

To the lead's question — *is the t357 dependency handled or merely mentioned?* — it is **mentioned
well and handled weakly**. `plan.md` §C.4 discloses it and instructs the lane to report rather than
absorb a divergence, which is correct discipline. What is missing is that the disclosure treats t357
as affecting only a *number* (§B.6's prediction), when it also affects a *decision* (REQ-SSF-007) whose
sole justification it can invalidate.

**Required fix:** (a) add an AC that decides REQ-SSF-007 mechanically; (b) state in REQ-SSF-007 or
§D.3 which of the two promotion mechanisms t357 uses, or record explicitly that the requirement's
justification holds only while the rule's severity is `warning` at the `Finding` level, and name what
the lane does if t357 lands first.

---

**D4. The `mx_commit_sha` inheritance is named but never measured, and it is not a pure widening
there.**
`spec.md` §E "Out of Scope — adjacent machinery", §D.2; `plan.md` §F M2 — **Severity: major** —
**Class: optional**.

§D.2's "This is a strict widening. … Nothing that was repaired stops being repaired" is a claim about
the previously-caught set only. The uncaught set also changes, and for `mx_commit_sha` the change is
not benign: `needsSHABackfill(state.MxCommitSHA)` at `closer.go:332` writes the literal
`l60MxBackfillPlaceholder = "(this commit)"` over whatever was there.

```
$ grep -h '^mx_commit_sha:' .moai/specs/*/progress.md | wc -l                                  → 85
$ … | sed 's/^mx_commit_sha:[[:space:]]*//' | grep -vcE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)' → 9
$ … | sort | uniq -c | sort -rn
   3 null                3 (this commit)                 1 <NA>
   1 _<pending Mx-phase>_
   1 _(not applicable — this SPEC removes the Mx-phase concept; REQ-LR-004/007)_
```

`null` and `(this commit)` are already repaired today. The other three are not: after the inversion,
a close would overwrite `_(not applicable — this SPEC removes the Mx-phase concept)_` with
`(this commit)`, replacing a deliberate declaration with a placeholder. Only SPECs subsequently closed
are affected, so the blast radius is small — but it is unmeasured in the SPEC, and no AC covers the
`mx` path at all.

**Required fix:** either measure and accept the three values explicitly (an `Out of Scope` bullet
naming them is sufficient), or add an AC pinning `mx` behavior. Do not leave "inherits the widening as
a side effect" as the whole account of a write path.

---

**D5. §B.3's `pending-backfill (29)` no longer reproduces with the command §B.3 cites.**
`spec.md` §B.3, §D.1 — **Severity: minor** — **Class: optional**.

```
$ grep -rh 'pending-backfill' .moai/specs/*/progress.md | grep -oE 'pending-backfill[a-zA-Z0-9-]*' | sort | uniq -c | sort -rn | head -2
  31 pending-backfill
  24 pending-backfill-sync
$ grep -o 'pending-backfill[a-zA-Z0-9-]*' .moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/progress.md | wc -l
   2
```

The SPEC's own `progress.md` §E.1 quotes the measurement command verbatim, and the command's glob
includes that file, so the artifact now contributes two self-matches. 31 − 2 = 29: the figure was
correct for the corpus at the moment it was taken and is correct today for the corpus excluding this
card's own artifact. The defect is that a reader running the cited command gets a different number
with no explanation — in a SPEC whose §B opens "Each row names the command that produced it", that is
a reproducibility gap rather than a wrong number. `pending-backfill-sync (24)` and the 28 distinct
spellings are unaffected either way.

**Required fix:** exclude this SPEC's own directory in the cited command, or annotate the row with the
self-match count.

---

**D6. §B.2 row 4's prose miscounts its own denominator.**
`spec.md` §B.2 (row-4 commentary); `acceptance.md` AC-SSF-002 rationale — **Severity: minor** —
**Class: optional**.

Both read "105 of 346 conforming values" / "105 of 346 conforming corpus values carrying an
annotation". 346 is the total, not the conforming count (334), and 105 counts `#` across all values
including six non-conforming ones.

```
$ python3 .moai/reports/t299/annot.py
total 346 | any-value with #: 105 | SHA-token values with #: 99
```

The load-bearing conclusion is untouched (99/334 = 29.6%, still "nearly a third").

**Required fix:** state it as "105 of 346 values, of which 99 of the 334 conforming ones".

---

**D7. Three requirements carry implementation detail.**
`spec.md` REQ-SSF-002 (`internal/spec`), REQ-SSF-007 (`eraDemotableCodes`), REQ-SSF-008
(`cleanFieldValue`) — **Severity: minor** — **Class: optional**.

Checklist item RQ-4 forbids naming internal identifiers in requirements. The convention is
established here (the `SPEC-MOVING-REF-GUARD-001` precedent does the same) and for a SPEC whose
subject *is* an internal contract the alternative is circumlocution, so this is recorded rather than
pressed. Reported for completeness, not as a change request.

---

**D8. AC-SSF-008 is an orphan, and its deciding command does not decide half of what it asserts.**
`acceptance.md` AC-SSF-008 — **Severity: minor** — **Class: optional**.

No REQ states the one-line-registry property; AC-SSF-008 (SHOULD) introduces it. Its command
(`git diff a6bbbf82b -- internal/spec/lint.go`) shows the diff, but "no existing rule entry is
reordered or removed" is then a human read of that diff rather than a binary outcome. Acceptable for a
SHOULD; noted because every other AC in this document is mechanically decidable and this one breaks
the pattern.

---

**D9. The slot's section number is stated three different ways.**
`plan.md` §A ("`progress.md` §E.4"); `acceptance.md` + `closer.go` comments (§E.2);
`spec-frontmatter-schema.md:140` (§E.3 / §E.4) — **Severity: minor** — **Class: optional**.

Pre-existing repository-wide inconsistency, not introduced here, and it decides nothing in this SPEC.
Recorded so the run phase does not spend a turn reconciling it.

---

## Recommendation

**FAIL.** The revision needed is small, mechanical, and confined to four locations. Nothing in the
design is wrong: the grammar is measurably sound on both the admit and reject sides, the demotion
derivation the dispatch flagged as suspect is **correct**, the falsifiability triad genuinely works,
and every §B figure except one reproduces exactly. What fails is the arithmetic layered on top of the
grammar — §B.6 was derived from §B.2's twelve without passing them through REQ-SSF-005's own
exemption, and three downstream surfaces inherited the error.

Ordered fixes for manager-spec:

1. **Recompute §B.6** (`spec.md` §B.6) against REQ-SSF-005: **5 total findings, 0 non-advisory**.
   Name the five — `SPEC-V3R6-SESSION-LEGACY-COVERAGE-001:47` and `:259`,
   `SPEC-V3R6-SPEC-ID-VALIDATION-001:103`, `SPEC-V3R6-SPEC-LINT-CLEANUP-001:45`,
   `SPEC-V3R6-TEST-REFACTOR-001:149` — and record that the seven placeholder-holding slots (including
   both `implemented` SPECs) are silent by design, which is REQ-SSF-005 working.
2. **Rewrite §D.3's first paragraph** so the `warning` choice rests on the corrected figure rather
   than on two SPECs the rule will not flag.
3. **Update AC-SSF-006** (`acceptance.md`): expected 5 total / 0 non-advisory; revise the "Fails
   when" clause — under a wrong-attachment implementation the non-advisory count would be 5, not 12.
   Keep the "report, do not tune" wording verbatim; it is the best clause in the document.
4. **Update `plan.md`** §C.3's "Predicted non-advisory findings: 2" and §F M4's "the twelve
   findings' locations" (five findings; twelve values; the seven-item placeholder inventory is worth
   keeping as a separate list for the operator decision in §E).
5. **Add an AC for REQ-SSF-007** and state which promotion mechanism t357 M2 uses — or record
   explicitly that REQ-SSF-007's justification is scoped to a `Finding`-level `warning` severity, and
   name the lane's action if t357 lands first (D3).
6. **Measure the `mx_commit_sha` widening** or scope it out by naming the three affected values (D4).
7. Optional, cheap: D5 (self-matching command), D6 (denominator).

Scope discipline was checked and is **clean**: `spec.md` §E, `plan.md` §D, and the DoD all forbid
corpus repair, forbid touching `SPEC-BACKLOG-LOCK-BUDGET-001`, and forbid altering `cleanFieldValue`,
and no requirement smuggles any of it back in — REQ-SSF-008 and AC-SSF-009 pin the last one with an
empty-diff command. REQ-SSF-002's "one shared predicate" is **achievable as specified**: the four
readers of §B.4 are value *extractors*, not SHA predicates, so the predicate needs exactly the two
decision call sites AC-SSF-007 names (`closer.go`, the new lint file). No fifth call site is hidden.
One implementation note the SPEC does not state and the run phase should: the closer feeds the shared
splitter a **`cleanFieldValue`-normalized** string while the lint rule feeds it a **raw** line — both
work under §D.1, but the SPEC never says which input the predicate is contracted against.
