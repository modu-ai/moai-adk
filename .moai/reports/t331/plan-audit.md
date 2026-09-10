# SPEC Review Report: SPEC-TODO-LANDING-STATE-001 (card t331)

Iteration: 1/2 (Tier M ceiling, `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.74** (Tier M PASS threshold: 0.80)

Audited tree: worktree `.claude/worktrees/t331`, branch `WT-card-landing-state`, HEAD `e497c3dba`.
Artifacts read: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set — complete).
Reasoning context ignored per M1 Context Isolation; every claim below was re-measured in this tree.

Refs at audit time (they moved since authoring — see D8):
`origin/main` → `origin/develop` divergence `0 334` (the SPEC pinned `0 329`).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-TLS-001`..`REQ-TLS-026`, 26 ids, sequential, no
  gaps, no duplicates, uniform 3-digit padding (`grep -o 'REQ-TLS-[0-9]\{3\}' spec.md | sort -u`).
  `AC-TLS-001`..`AC-TLS-016` likewise.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md §C`), not the verification layer. All 26 entries match a GEARS pattern: Ubiquitous
  (`spec.md:411`, `:423`, `:445`), Where (`:413`, `:429`, `:462`), When (`:425`), While (`:436`),
  Unwanted/negative (`:427`, `:439`, `:469`). The Given/When/Then entries in `acceptance.md` are the
  correct verification-layer format and are graded under Group 4, not here.
  One blemish, not a failure: `REQ-TLS-012` (`spec.md:440-441`) mixes a `shall not` clause with a
  declarative sentence ("the permissive policy … is preserved unchanged").
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-13`): `id` / `title` / `version:"0.1.0"` / `status:draft` / `created` / `updated` /
  `author` / `priority:P1` / `phase:"v3.1.4 target"` / `module` / `lifecycle:spec-anchored` /
  `tags`. No rejected snake_case alias. `phase` is a release-target label, not a prohibited
  lifecycle token. `id` matches the implementation regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`
  (`internal/spec/lint.go:715`), so `progress.md:9`'s self-check claim is verified true. `tier:` and
  `related_specs:` are extra; `tier` is a documented optional field, `related_specs` is not in the
  schema (harmless).
- **[N/A] MP-4 Section 22 language neutrality** — single-project Go/CLI scope. The one
  template-bound artifact (`internal/template/templates/.claude/skills/moai/workflows/todo.md`) is
  language-neutral doctrine and REQ-TLS-025 keeps the mirror in the same change. See D17 for the
  neutrality-CI-guard omission (optional).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — every referenced SPEC exists and none is
  retired / superseded / archived: `SPEC-KANBAN-QUEUE-PR-SYNC-001` `in-progress`;
  `SPEC-TODO-DESTRUCTIVE-GUARD-001`, `SPEC-TODO-SQLITE-001`, `SPEC-WORKTREE-BASEREF-001` all
  `completed`. No D7-status BLOCKING. The **semantic** cross-SPEC conflicts I found (D1, D2) are
  recorded under Group 6 Consistency, not as D7-status findings.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-PASS per D8-4.
- **[FAIL] MP-7 clarification gate** —
  `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-LANDING-STATE-001/` returns three matches:
  `plan.md:54`, `plan.md:61`, `plan.md:65`. (`research.md` does not exist — Tier M — so the grep
  subject is `plan.md` alone; that half is not N/A.) See D9.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.70 | between 0.50 and 0.75 | `spec.md:448` "the observed commit SHA" is singular where the predicate is many-to-one — t293 matches 9 commits on `origin/develop`, t322 matches 24 (measured this run). A reasonable engineer implements "first match"; `SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:253-255` names that exact reading as wrong. `spec.md:222-224` "with the observed commit named" is ambiguous on the same axis. Everything else in the document is unusually precise. |
| Completeness | 0.75 | 0.75 | All sections present: HISTORY `:24`, WHY `§A:43`, WHAT/HOW `§B:241`, REQUIREMENTS `§C:407`, Out of Scope `§D:485` with four `### Out of Scope — <topic>` H3 sub-headings (`:490`, `:506`, `:514`, `:525`) each carrying specific `-` bullets, Traceability `§E:541`, Cross-refs `§F:563`; ACs in `acceptance.md` per Tier M. Deductions: the REQ budget breach (D4) and the un-required C2 behaviour (D6). |
| Testability | 0.75 | 0.75 | 14 of 16 ACs carry a re-runnable command and a stated GREEN; zero weasel words (`grep -Ei '(appropriate\|adequate\|reasonable\|proper)' acceptance.md` → no match). Deductions: `acceptance.md:58` (AC-TLS-013) cannot go red and fails the mutant probe (D5); `acceptance.md:55` cites an address where its RED evidence is not (D7). |
| Traceability | 0.75 | 0.75 | `spec.md §E:543-559` covers all 26 REQs with at least one AC — no uncovered requirement. Deductions: `AC-TLS-016` (`acceptance.md:61`) is absent from the table and traces to a `§A.6` table row rather than a REQ (D6); four `file:line` citations do not resolve to the text they carry (D7). |

Aggregate = 0.74 (arithmetic and harmonic means agree to two places). Below the Tier M threshold of
0.80, and MP-7 fails independently.

---

## What I re-measured and CONFIRMED

Stated first so the FAIL is not read as doubt about the thesis. The **root-cause claim is correct**,
independently reproduced at HEAD `e497c3dba`:

| Claim | SPEC (@`3de2f85a2`) | This audit (now) | Verdict |
|---|---|---|---|
| `LandedRef = "origin/main"` at `prlink_landed.go:28` | quoted | line 28 exact; its only behavioural use is `:44` | CONFIRMED |
| the project integrates on `develop` | `worktree_base_branch: develop` | `git-strategy.yaml:5`, committed at `3de2f85a2`, clean in this tree | CONFIRMED |
| main lags develop | `0 329` | `0 334`; `merge-base --is-ancestor origin/main origin/develop` rc 0 | CONFIRMED (number moved — D8) |
| t293 0/9 | 0/9 | **0/9** | CONFIRMED |
| t310 0/6 | 0/6 | **0/6** | CONFIRMED |
| t322 0/5 | 0/5 | **0/24** — same direction, count moved | CONFIRMED (D8) |
| t200 1/1, reads `landed` | 1/1 | **1/1**, `294b4b6ab` on both refs | CONFIRMED |
| t200 squash-blindness | `7fc161b36` rc 1, `294b4b6ab` rc 0 | **rc 1 / rc 0**, subjects match | CONFIRMED |
| unresolvable-ref probe | rc 128 | **rc 128**, verbatim `fatal: ambiguous argument` | CONFIRMED |
| t338 zero commits | 0 | **0** | CONFIRMED |
| C7 exemplars promoted to main | asserted, unmeasured | t241 main=1, t278 main=2 | CONFIRMED (was unmeasured in the artifact) |
| C5 exemplars sync-closed | asserted, unmeasured | t293 `812ee01fc`, t322 `5e194bba2` sync-close commits | CONFIRMED (was unmeasured) |
| F6: both outcomes exit 0 | 2 tests PASS | 3 tests PASS this run (`0.13s` / `0.13s` / `0.21s`); `todo.go:417-431` returns `nil` on both paths | CONFIRMED |

**What would falsify the root-cause claim**, stated so a later reader can attack it: (a) any commit
naming t293 / t310 / t322 appearing on `origin/main` — measured 0 for all three; (b) the landed leg
of `ResolveCardPRLink` not routing through `LandedRef` — `prlink.go:175-185` shows it does, via
`GitLandedQuerier.Landed` → `LandedGrepArgs` → `LandedRef`; (c) `git_strategy.worktree_base_branch`
not naming `develop` in the committed tree — it does. None falsifies it. The example set is right:
**no card the SPEC names as a false-not-landed is absent from `develop`.**

Also confirmed: AC-TLS-015's mirror claim — `diff` of the two `todo.md` copies is **empty**, 241
lines each, so "mirror at the same line numbers" is true.

---

## Defects Found

**D1. `spec.md:296-308` (§B.2 M1) — the deviation's primary justification misattributes the freeze to REQ-TODO-013, and a completed sibling SPEC has already ruled the other way. — Severity: critical — Class: blocking**
The SPEC deviates from the dispatching lead's explicit `ALTER TABLE ADD COLUMN` instruction and rests
M1 on: "**per-item fields are frozen** … REQ-TODO-013 freezes the second [face]". The quoted code
comments are verbatim-accurate (`backlog_store.go:43-45` and `:62-63`, both verified). The
**inference is not**. REQ-TODO-013's actual text (`.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:59`)
reads: *"The backlog store shall preserve the existing version-1 record shape … **changing it only
additively**."* It mandates additive-only change; it does not freeze the per-item field set. The
"no per-item field may ever be added" language is the **code comment's gloss**, and the comment
itself attributes it to a different document ("spec.md §E out-of-scope") — a scope declaration, not
a requirement. Worse, a **completed** sibling SPEC records the contrary reading in writing:
`.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md:51` — *"그 SPEC의 §E는 자기 범위의 out-of-scope 선언이지
영구 동결이 아니고, 같은 REQ가 additively 변경을 허용한다"* (that §E is a scope-local out-of-scope
declaration, not a permanent freeze; the same REQ permits additive change, with `last_seq` as the
precedent). This is an unverified premise dressed as a citation
(`verification-claim-integrity.md` §1.1 surface 4), and it is the premise a human gate will rule on.
In fairness: M2 (`spec.md:310-324`, the `IF NOT EXISTS` / CHECK-rebuild precedent — verified verbatim
at `backlog_sqlite.go:90-97`) and REQ-TLS-016's ordered multi-observation requirement support the
chosen shape **independently**, so the decision may well be right; the stated reason for overriding
the lead is not.
*Required fix*: quote REQ-TODO-013 verbatim; state plainly that the freeze comes from
SPEC-KANBAN-TODO-CLI-001 §E (a scope declaration) as glossed by `backlog_store.go:45`, not from
REQ-TODO-013; reconcile `SPEC-TODO-ANALYSIS-001/spec.md:51` explicitly; and re-rest the deviation on
M2 + REQ-TLS-016, which survive the correction.

**D2. `spec.md:222-224` (Table 2, rows C4 and C5) — requires `todo pr` to name the delivering commit, reversing REQ-1.10 without reconciliation. — Severity: critical — Class: blocking**
Both rows require `todo pr` to render "`landed` **with the observed commit named**".
`SPEC-KANBAN-QUEUE-PR-SYNC-001` (status `in-progress`) `spec.md:251-255` states: *"**REQ-1.10** — The
resolver shall not name, return, or otherwise claim which commit delivered a card. The `landed`
outcome is a boolean fact about `origin/main` and nothing more. (Grounds: a card's first matching
commit may be another card's report commit that merely mentions it, so any 'first match is the
delivering commit' reading attributes wrongly.)"* The ruling is enforced in three places in code —
`prlink_landed.go:64-65` ("returns a boolean and an error, and NOTHING else (REQ-1.10) … no SHA and
no subject escapes this function"), `prlink.go:43`, `prlink.go:76`. Nothing in this SPEC's §A.5
inheritance table, §D exclusions, or §F cross-references acknowledges that REQ-1.10 is being
reversed, and REQ-1.10's stated grounds are not addressed anywhere. Recording the SHA on the
**observation** (§B.3 / REQ-TLS-014) is a different surface and is not in conflict; the `todo pr`
read surface is.
*Required fix*: either drop "with the observed commit named" from Table 2 C4/C5, or add an explicit
reversal clause naming SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.10, state why the misattribution grounds
no longer bind, and add a REQ that carries the reversal — today no REQ does (see D6).

**D3. `spec.md:447-450` (REQ-TLS-014, REQ-TLS-015) — "the observed commit SHA" is singular where the predicate is many-to-one, and no rule selects among matches. — Severity: major — Class: blocking**
The landed predicate is `git log <ref> --grep='\b<id>\b'`, which returns a set. Measured this run:
t293 → 9 matching commits on `origin/develop`; t322 → 24. REQ-TLS-014 records "the observed commit
SHA" and REQ-TLS-015 constrains only that it be *on the ref* — a constraint every one of the 24
satisfies. No requirement, milestone (`plan.md:123-130`, M5), or AC (`acceptance.md:54`) says which
commit is recorded, and the obvious default ("first match") is precisely the reading REQ-1.10
forbids by name. AC-TLS-009's assertion (`merge-base --is-ancestor <sha> <ref>` rc 0) passes for a
wrong-but-on-ref SHA, so the AC does not close the gap.
*Required fix*: add a requirement fixing the selection rule (newest match, or record the whole
matched set), and strengthen AC-TLS-009 to assert the selected SHA against a fixture with more than
one match.

**D4. `spec.md:411-482` + `plan.md:14` — 26 requirements against a Tier M ceiling of 16 (and a Tier L ceiling of 25). — Severity: major — Class: blocking**
`spec-workflow.md` § SPEC Complexity Tier sets the requirement ceiling at S=8 / M=16 / L=25 and
states that exceeding it "is a signal to tier up or to split the SPEC, not to relax the budget",
because the auditor must hold every requirement in view at once. Measured: 26 requirements
(`grep -c '^- \*\*REQ-TLS-'` → 26), 16 ACs (exactly at the M ceiling). 26 exceeds Tier M by 10 and
exceeds even Tier L. `plan.md:14` declares Tier M and justifies it on blast radius alone, never
addressing the REQ budget.
*Required fix*: split — the ref correction (§C.1) + three-valued answer (§C.2) + stdout
distinguishability (§C.3) is one coherent SPEC; the landing-evidence store (§C.4) + recording verb +
live SPEC status (§C.5) is a second that depends on it. Or merge the near-duplicate authority
requirements (REQ-TLS-021 / 022 / 023 / 024 are four statements of one rule) and re-count.

**D5. `acceptance.md:58` (AC-TLS-013), classified release-blocking at `acceptance.md:67` — a criterion that cannot fail, and that a mutant satisfies while violating its requirement. — Severity: major — Class: blocking**
Its own RED cell says "Cannot go red today — no detection path writes", and its GREEN is a source
sweep: "no assignment to `Items[i].State`, no `ArchiveCard`, no `Drop` call reachable from the
landing paths". `verification-completeness.md` §1.1 and §2 are [HARD] here: a verification artifact
is incomplete until its failure has been observed on a known failing input, and adoption takes a
RED-now cell paired with a green path. This criterion has neither. It also fails the §2 **mutant
probe** — a mutant that writes state through a differently-named helper, or issues
`UPDATE items SET state=…` at the SQLite layer (`backlog_sqlite.go` owns that surface), satisfies
the three named greps while violating REQ-TLS-022 outright. The name-based sweep is exactly the
"too shallow to adopt" shape the rule describes.
Contrast the honest handling of the same shape elsewhere: AC-TLS-002 (`:47`) and AC-TLS-007 (`:52`)
also have no RED and are correctly reclassified as regression-guards at `:68`. AC-TLS-013 is not.
*Required fix*: establish RED by mutation — plant a state write in the landing path, observe the
criterion go red, revert, cite both outputs — or reclassify it as a regression-guard and stop
calling it release-blocking.

**D6. `acceptance.md:61` (AC-TLS-016) + `spec.md:543-559` (§E) — an orphaned, release-blocking AC verifying behaviour no requirement mandates. — Severity: major — Class: blocking**
AC-TLS-016's "Verifies" column names `spec.md §A.6 C2 row`, not a REQ, and it is the only AC absent
from the §E traceability table (verified: the table's 15 rows cover REQ-TLS-001..026 and
AC-TLS-001..015). It is nonetheless listed release-blocking at `acceptance.md:67`. The behaviour it
asserts — `todo pr` rendering the queue `state` beside the link outcome, so `queued`+`no-link` and
`picked`+`no-link` are different rows — is required by no REQ in §C. Its RED is real, which makes the
gap worse rather than better: `todo_pr.go:175-176` prints five columns (`CardID, Kind, PRs,
Confidence, text`) and none is the queue state. The same gap covers Table 2's C4/C5 "commit named"
cell (D2).
*Required fix*: add a requirement for the state column (and one for whatever survives D2), then add
the row to §E. An AC that traces to a table row rather than a requirement is untraceable by
construction.

**D7. Four `file:line` citations carry a real quote at a wrong address. — Severity: major — Class: blocking**
This is the class already corrected once in this SPEC (`backlog_store.go:152-157` → `:42-46`); it
recurs. Each re-measured this run:
- `spec.md:311` cites `backlog_sqlite.go:96-100` for the quoted archived-`*` precedent block. The
  quote actually spans **90-97**. Lines 96-97 carry only its tail; 99-101 are a *different*
  paragraph ("archived_items deliberately carries NO state CHECK"). The quote's opening sentence
  lies outside the cited range.
- `spec.md:498` cites `backlog_sqlite.go:96-100` and `:105-112` as where `items.state` is
  constrained with `CHECK (state IN ('queued','picked','dropped'))`. That CHECK is at line **113** —
  outside both ranges.
- `acceptance.md:55` (AC-TLS-010's RED evidence) cites `backlog_sqlite.go:232-235` for "runs the
  whole DDL on every open". Lines 232-235 are the `backlogEngine` struct close and the head of
  `openBacklogEngine`'s doc comment; the actual `ExecContext(ctx, backlogDDL)` is at line **268**.
  A re-runner cannot find the RED evidence at the address given.
- `spec.md:383` cites `todo.md:47` for "The verb writes a record and touches neither card; `absorbs`
  does not absorb". That sentence is at `todo.md:48`; line 47 is the `moai todo analyze` row — a
  different verb.
Two further under-covering ranges, same family, lower cost: `spec.md:93` cites `prlink.go:160-178`
for the landed leg, which is at 175-185 (`if isLanded` at 182, outside); `acceptance.md:50` cites
`prlink.go:170-178` for the error-alongside path, which is at 179-181.
*Required fix*: re-measure and re-cite all six; adopt the discipline that a range citation must
contain the **first** line of the quoted text.

**D8. `spec.md:69-88` + `acceptance.md:13-17` — every `origin/develop` figure is pinned to a tree SHA, which does not determine a moving remote ref. — Severity: major — Class: blocking**
`verification-completeness.md` §4 [HARD] requires measurement claims to pin the tree where the
evidence was collected. The SPEC does pin `3de2f85a2` — but the *measurement subject* is
`origin/develop`, a remote-tracking ref that moves independently of any tree SHA, so pinning the
tree does not pin the number. Demonstrated this run: the divergence read `0 329` in the artifact and
`0 334` now; t322's develop count read `5` in the artifact and `24` now. Both moved in the direction
the SPEC predicted, so no conclusion is affected — the reproducibility is. `acceptance.md:15-17`
compounds it by making the tree SHA the re-measurement contract for every RED cell. The §4
discriminator applies: these are measurements about a moving branch, so the fix is to record the
ref's SHA at measurement time, not to abandon pinning.
*Required fix*: alongside `@3de2f85a2`, record `origin/develop`'s own SHA at measurement time, and
say in `acceptance.md §A.1` that a count against a moving ref is re-measured rather than re-cited.

**D9. `plan.md:54`, `plan.md:61`, `plan.md:65` — three unresolved `[NEEDS CLARIFICATION]` markers at audit time (MP-7). — Severity: critical — Class: blocking (clarification gate)**
`grep -rn '\[NEEDS CLARIFICATION' plan.md` returns three matches: the landing-verdict token
spelling, the exit code on `unknown`, and which verb records the observation. Per MP-7 this is a
score-independent gate: the orchestrator MUST resolve each topic via `AskUserQuestion` before
Implementation Kickoff Approval, and no aggregate score resolves it.
In fairness to the author, these are well-formed, each carries a recommendation and a stated
consequence, and `progress.md:14` declares them — this is a gate to walk through, not rework. But it
is open, and the third (which verb records the observation) is load-bearing for M5 and REQ-TLS-023,
so run-phase cannot begin without it.
*Required fix*: orchestrator runs one `AskUserQuestion` round over the three; the answers replace
the markers in `plan.md §C`.

**D10. `spec.md:268-278` + `plan.md:19` — "Six production sites" contradicts the seven-line table directly beneath it. — Severity: minor — Class: blocking**
§B.1 M3 says "Six production sites and one test", and `plan.md:19` repeats `6`. Its own table lists
seven production line references: `prlink_landed.go:44`, `:78`, `todo_pr.go:75`, `:87`,
`todo.go:357`, `:399`, `:428`. My `grep -rn 'LandedRef' --include='*.go' internal/` returns 11 lines
— 1 comment, 1 declaration, 7 production uses, 2 test lines (`todo_undone_test.go:277`, `:278`). The
count is 7. The conclusion ("the measured radius is small") is unaffected, but the number
contradicts the evidence printed beside it, and it is the number the deviation-from-t330 argument
leans on.
*Required fix*: change 6 → 7 in both places, or state which site is excluded and why.

**D11. `spec.md:326-330` (§B.2 M3) — the "one column cannot hold two landings" argument answers an alternative the dispatch did not propose. — Severity: minor — Class: optional**
The dispatch said "evidence **columns**" (plural). M3 argues only against a single column. Two
columns (`run_landed_sha`, `sync_landed_sha`) would hold both events, so M3 alone does not decide
it. What decides it is REQ-TLS-016 (unbounded, ordered observations) plus M1/M2 — a fixed column
count cannot express "all observations retained in order". Worth restating on that ground so the
argument survives the obvious counter.

**D12. `spec.md:222-228` (Table 2, rows C5 / C6 / C7) — three rows become cell-identical once S3 is retired. — Severity: minor — Class: optional**
Table 2 marks the S3 column "*retired as a decision input*" for every row. With S3 gone, C5, C6 and
C7 carry identical cells in S1 / S2 / S4 and differ only by provenance (which ref; merge vs squash)
— which the table's own row legend (`spec.md:185-187`) says must be "separately observable". After
the change they are not. Collapsing them into one row with a provenance note would make Table 2 say
what it means; nothing in the design changes either way.

**D13. `spec.md:202-213` (Table 1) — a state is missing: `queued` while the integration branch names the card. — Severity: minor — Class: optional**
The ten rows cover `queued`+not-landed (C1) and `picked`+landed (C4-C7), but not `queued`+landed —
the state a freshly-added duplicate card occupies when work for it has already shipped. That is
close to the card's own premise (the lead files new work for something already done), and
`todo add` refuses only byte-identical duplicates, so it is reachable. The corrected S2 answers it
correctly, so the design already handles it; only the enumeration is short.

**D14. `acceptance.md:51` (AC-TLS-006 RED cell) — one of the two cited tests does not assert what it is cited for. — Severity: minor — Class: optional**
The cell says the two PASSing tests "establish the indistinguishability" of stdout.
`TestTodoDone_NoLandingQueryWithoutTheFlag` (`todo_undone_test.go:306-335`) asserts `err == nil` and
a subprocess count — it never inspects stdout. The two tests therefore establish exit-code identity,
not stdout identity. The conclusion is nevertheless **true**, by the code rather than by the tests:
`todoRequireLanded` returns `nil` on both the landed and the error path (`todo.go:423`, `:430`), so
`done` prints identically. Separately, the "verbatim" PASS pair quoted in that cell cannot both have
come from the `-run 'TestTodoDone_RequireLanded'` invocation shown at `spec.md:133` — that regex
excludes the second test. I reproduced all three PASSes in one run this audit.
*Suggested fix*: cite `todo.go:417-431` as the evidence for stdout identity and keep the tests as
evidence for the exit code.

**D15. `plan.md:159` (R6) — the in-repo half of the risk is measurable now, and measures to zero. — Severity: minor — Class: optional**
R6 defers "check `internal/web` for consumers before landing M2". `internal/web` exists, and
`grep -rn 'PRLink' --include='*.go' .` (non-test, outside `internal/kanban/prlink*`) returns matches
in exactly one file: `internal/cli/todo_pr.go:135`, `:141`, `:176`, `:181-182` — and it formats
`o.Kind` with `%s` rather than switching on it. There is no exhaustive in-repo switch to break. The
residual risk (external consumers of `todo pr --json`) is real and cannot be greped, so the row
should stand — with the in-repo half discharged rather than deferred. Note also that the "documented
as a closed set" wording belongs to `PRLinkConfidence` (`prlink.go:50-51`), not to `PRLinkKind`,
whose comment (`prlink.go:30`) says only "one of the four … outcome kinds".

**D16. `plan.md:88` (M1) — the six user-facing text sites are listed, but the doc comments that hardcode `origin/main` are not. — Severity: minor — Class: optional**
After M1 these go stale and silently mislead the next reader: `prlink_landed.go:26-27` ("LandedRef
is the ref…"), `:33-34` ("origin/main is an existing local ref" — a premise REQ-TLS-004 deliberately
stops relying on), `prlink.go:31` ("already in origin/main"), `:42`. Naming them in M1 costs one line
and prevents the comment layer from contradicting the code.

**D17. `plan.md:32-35` (§B.1) — Template-First is carried, the template-neutrality CI guard is not. — Severity: minor — Class: optional**
REQ-TLS-025 mandates editing `internal/template/templates/.claude/skills/moai/workflows/todo.md`.
Changes under that path trigger `.github/workflows/template-neutrality-check.yaml`, which rejects
SPEC IDs, REQ tokens, internal dates and commit SHAs in template content. This SPEC's own vocabulary
is full of all four, so a careless mirror edit trips CI. One constraint line in §B prevents it.

**D18. `spec.md:296` — the `backlog_store.go:42-46` range is off by one at both ends. — Severity: minor — Class: optional**
The quoted `backlogVersion` comment spans lines **43-47**; the cited range is 42-46. Unlike D7 the
range does contain the quoted text, so this is imprecision rather than a wrong address — recorded
for completeness since the same authoring pass produced D7.

---

## Regression Check

Not applicable — iteration 1.

---

## Recommendation

The thesis is right and the document is, in most respects, unusually good: the root cause is
correctly identified and I reproduced every one of its measurements; the exclusions are reasoned
rather than listed; the non-goal is stated in the body (§B.4) **and** restated where a scope reader
meets it (§D, `spec.md:508-509`), which is exactly right; and the artifact is honest about what it
did not observe. The FAIL is driven by MP-7 plus a small number of specific, fixable defects — not
by an accumulation of preferences.

Fix in this order:

1. **D9** — run the `AskUserQuestion` round over the three `[NEEDS CLARIFICATION]` topics
   (`plan.md:54`, `:61`, `:65`) and write the rulings into §C. Nothing else proceeds past the
   kickoff gate while these are open.
2. **D1** — repair the deviation's justification. Quote REQ-TODO-013 verbatim
   (`SPEC-KANBAN-TODO-CLI-001/spec.md:59`), attribute the freeze to the §E scope declaration it
   actually comes from, reconcile `SPEC-TODO-ANALYSIS-001/spec.md:51`, and rest the deviation on M2
   + REQ-TLS-016. The lead is being asked to accept an override of an explicit instruction; the
   reason given for it must be true.
3. **D2 + D3** — decide the `todo pr` commit-naming question against REQ-1.10
   (`SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251-255`). Either drop the "commit named" cells, or
   reverse REQ-1.10 explicitly and add a REQ carrying the reversal plus the multi-match selection
   rule REQ-1.10's grounds demand.
4. **D4** — split the SPEC (§C.1-§C.3 first, §C.4-§C.5 second) or consolidate REQ-TLS-021..024, and
   re-count against the Tier M ceiling of 16.
5. **D5 + D6** — give AC-TLS-013 a mutation-established RED or reclassify it; add the missing
   requirement behind AC-TLS-016 and put it in §E.
6. **D7 + D8 + D10** — re-measure and re-cite the four wrong addresses, record `origin/develop`'s
   SHA beside the tree SHA, and correct 6 → 7.

D11-D18 are optional: surface them and let scope discretion decide. Routing all of them into a
revision would add speculative requirements this SPEC never claimed.

On the author's declared gaps, judged one by one:

- *No end-to-end `moai` binary run for F6* — **honestly scoped.** The two Go tests plus the rc=128
  probe establish the claim, and I confirmed the code path independently (`todo.go:423`, `:430`). A
  binary run would add nothing the code does not already show.
- *t338's "43 untracked artifacts / stale index.lock" unverified beyond the zero-commit half* —
  **honestly scoped, and correctly fenced.** AC-TLS-016 states the limit inside the criterion ("does
  not claim to detect work that produced no commit"). The unverified half is not load-bearing for
  any requirement.
- *No exhaustive check of `PRLinkKind` consumers (R6)* — **half of it is a hole wearing a Gap
  label.** The in-repo half costs one grep and measures to zero (D15); deferring it to run-phase
  deferred a measurement the plan phase could have made. The external-consumer half is genuinely
  unknowable and correctly deferred.
- *`internal/web` unchecked* — **now checked**: it exists and contains zero `PRLink` references
  (D15).
