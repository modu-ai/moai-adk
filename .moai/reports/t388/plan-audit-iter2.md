# SPEC Review Report: SPEC-VERSION-STAMP-GUARD-001

- Iteration: 2/3
- Verdict: **FAIL**
- Overall Score: **0.80** (Tier S PASS threshold 0.75 — the aggregate alone would pass; the FAIL is driven by D1, a major blocking defect in the SPEC's central evidence mechanism)
- Score movement: iter-1 0.77 → iter-2 0.80. Improvement, so no STOP escalation under the score-regression clause.
- Audited tree: `9328a52422baa13fdb7d7fd0c8409151da3ba3c1`, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t388`, branch `WT-version-sync-list`
- SPEC version audited: `0.3.0`, 6 REQ / 6 AC, Tier S, status `draft`
- Reasoning context ignored per M1 Context Isolation. The dispatch carried the lead's own measurements; every one was re-run in this tree before use and the re-runs are quoted below. One of the lead's statements did not survive re-measurement as stated — see D3.

---

## Scope of this iteration

Per the Retry Loop Contract, iteration 2 is scoped to the previous iteration's enumerated defect
delta plus a regression check over it, not a from-scratch re-audit. The card was **split** by
operator ruling between iterations, so the delta here is unusually large: three requirements, one
acceptance section, the §3 deny-list design, the `internal/versionstamp/` package, and the
`eba919e44` RED fixture were deleted rather than repaired. The regression check below therefore
covers both the iter-1 defect list and the integrity of the transfer itself.

---

## Regression Check — iter-1 defects

| ID | iter-1 severity | Status | Evidence |
|---|---|---|---|
| D1 (§3 survivor set 10 vs 8) | major | **RESOLVED (by deletion)** | §3 is now the GEARS requirements section (`grep -n '^## '` → `174:## §3 요구 (GEARS)`). The survivor-set enumeration no longer exists. Left one dangling pointer — see new D2. |
| D2 (measured area ≠ stated predicate) | critical | **RESOLVED (by scope transfer)** | The token predicate moved to card t392. Re-measured: the area figure is preserved verbatim and correctly in §2 (`2225줄 / 592파일`, reproduced exactly). The transfer is clean — see § Transfer integrity. |
| D3 (testdata fixture is itself a finding) | major | **RESOLVED (by deletion)** | `grep -rn testdata` over all four artifacts → no match. The `eba919e44` RED fixture is gone; the three surviving `eba919e44` mentions are HISTORY, an explicit Out-of-Scope refusal, and a §8 provenance line. |
| D4 (AC-VSG-001 judged by an unreachable object) | major | **RESOLVED** | `acceptance.md:33-42` now carries the literal 7-path set as judge, with `[HARD] 판정자는 이 리터럴 집합이며, 저장소 객체를 조회하지 않는다` at :44. `61921f1ba` demoted to provenance at :46-47, and §7 R-1 records the reachability fact. Verified independently: `git merge-base --is-ancestor 61921f1ba HEAD` → rc=1; `git branch -a --contains 61921f1ba` → `release/v3.1.4` only; `git show --numstat --format= 61921f1ba` → exactly the same 7 paths, 9 changed lines. |
| D5 (`Version` called a constant) | minor | **RESOLVED** | REQ-VSG-003 (`spec.md:180`) now reads "`Version` is a package-level `var`, not a constant, and the documentation shall not call it one." AC-VSG-003 item 3 (`acceptance.md:92-93`) mirrors it. Source confirms: `pkg/version/version.go:7-8` is `var (` / `Version = "v3.1.3"`. |
| D6 (counting-unit ambiguity) | minor | **RESOLVED, then partly re-broken** | §2 now carries a `[HARD] 계수 단위` clause (`spec.md:121-124`) fixing line vs occurrence, and both figures reproduce (2225 lines / 2494 occurrences). But that clause cites a `§3의 단위 고정 조항` that no longer exists — new D2. |
| D7 (AC-VSG-006 synthetic input unpinned) | minor/optional | **RESOLVED** | Re-applied to the surviving check: `acceptance.md:110` `[HARD] 심는 경로는 **고정한다**: docs-site/nonexistent-stamp.toml`, with the two-property justification. Verified absent: `test -e docs-site/nonexistent-stamp.toml` → rc=1. |
| D8 (`Where` is a data condition, not a capability gate) | minor/optional | **CARRIED FORWARD** | The clause moved verbatim in shape to REQ-VSG-004 (`spec.md:182`): "**Where** a path is named under the release-artifact heading". Still a data condition. Re-raised as D9 below at the same optional severity. |
| D9 (modal "may" in normative text) | minor/optional | **RESOLVED (by deletion)** | REQ-VSG-007 no longer exists. |
| D10 (§1.1 cited only the Makefile) | minor/optional | **RESOLVED** | `spec.md:69-77` now cites `Makefile:20,36,72` **and** `.goreleaser.yml:22`, and narrows the exposure to the hand-build path explicitly ("노출 면적은 릴리스 바이너리가 아니라 손빌드 경로에 한정된다"). |

No iter-1 defect is unresolved-and-unaccounted. No stagnation.

---

## Transfer integrity (dispatch item 1) — clean

Checked for orphans, dangling references, subjectless requirements, and residue of the deleted
guard machinery.

- **Renumbering is contiguous.** `grep -oE 'REQ-VSG-[0-9]+'` over all four artifacts → exactly
  `REQ-VSG-001..006`, no gap, no duplicate, uniform 3-digit padding. Same for
  `AC-VSG-001..006`. AC headings at `acceptance.md:30,64,82,103,137,157`.
- **Traceability is bidirectional and complete.** `spec.md:190-197` and the `acceptance.md:19-26`
  matrix agree in both directions. Every REQ has ≥1 AC; every AC names an existing REQ. No orphan,
  no uncovered requirement. Corroborated mechanically: `moai spec lint` (binary built from this
  tree) emitted **0 findings for this SPEC** — `grep -c 'SPEC-VERSION-STAMP-GUARD-001'` over the
  full lint output returned `0`, against 1096 repo-wide warnings on other SPECs.
- **No residue that reads as live scope.** `grep -rn` over the four artifacts for `versionstamp`
  returns 5 hits, and every one of them is a refusal or a deletion record — `plan.md:62` and
  `plan.md:125` ("t392 소관이다"), `spec.md:258` (Out of Scope), `spec.md:228` ("0.2.0의
  `internal/versionstamp/`는 t392와 함께 나갔고"), `progress.md:27` ("함께 삭제(이월 아님)").
  Nothing states or implies the package is to be built here. `testdata`, `D2-OPEN`, `D4-NOTE`,
  `REQ-VSG-007`, `AC-VSG-007` all return no match.
- **No dangling cross-SPEC reference.** SPEC-ID extraction over all four artifacts returns only the
  self-reference `SPEC-VERSION-STAMP-GUARD-001`. `mcp__moai__spec_audit` with `project_root` set to
  this worktree returns one `INFO / EraAutoDetected` finding and no drift.
- **The measurement that justified the split is preserved and correct.** Independently re-run in
  this tree, both commands exactly as printed in §2:

      git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . ':!.moai/reports' ':!.moai/specs' \
        ':!.moai/release-notes' ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*'
      → lines=2225  files=592

      (same, with -hoE) | sort | uniq -c | sort -rn | head -6
      → 270 v3.0.0 / 83 v2.12.0 / 80 v3.1.1 / 80 v2.1.219 / 72 v2.14.0 / 68 v2.1.198
      total occurrences = 2494

  Every figure in §2 reproduces to the digit, including the tie order of the two 80s.

**One dangling pointer survives the reduction** — `spec.md:123` cites "§3의 단위 고정 조항", and §3
is now the GEARS requirements section. See D2.

---

## The `-n` inflation correction (dispatch context) — mechanism right, enumeration wrong

The lead reported the histogram correction as verified. The **mechanism** reproduces exactly:

    git grep -nE '<token>' -- <deny-list> | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort | uniq -c
    → 112 v2.14.0 ... 90 v2.12.0

against the `-h` form's 72 and 83. The deltas are 40 and 7, and the two named files carry exactly
those match counts (`git grep -cE` → `docs/design/v2.14.0-release-plan.md:40`,
`.moai/marketing/awesome-lists/github-release-v2.12.0-enhanced.md:7`). The `[HARD]` warning §2
raises for t392 is correct and worth keeping.

The **enumeration attached to it is not**. `spec.md:151` states "이 트리에는 그런 파일이 둘 있다".
There are **eight** in the deny-listed scope. See D3.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-VSG-001..006` at `spec.md:176,178,180,182,184,186`.
  Sequential, no gap, no duplicate, uniform padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer**
  (`REQ-VSG-*` in `spec.md` §3); the Given-When-Then `AC-VSG-*` entries in `acceptance.md` are the
  verification layer and are graded under Group 4, not here. 001/002/003/005/006 are Ubiquitous
  (`The <subject> shall …`). 004 is Ubiquitous with a trailing `**Where**` clause — syntactically a
  GEARS pattern. Semantic note carried forward as D9 (optional).
- **[PASS] MP-3 YAML frontmatter validity** — `spec.md:1-15`. All 12 canonical fields present with
  correct types: `id`, `title`, `version: "0.3.0"` (quoted), `status: draft`, `created`/`updated`
  ISO dates, `author`, `priority: Medium`, `phase: "v3.1.5 target"`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma string). No rejected snake_case alias. `tier: S`
  additionally present.
- **[N/A] MP-4 language neutrality** — single-language (Go) repo-internal SPEC.
  `find internal/template/templates -name 'version-management*'` → empty; the target doc has no
  template mirror, and the sole code artifact lands in `internal/cli/`. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — SPEC-ID extraction returns the self-reference only.
  No retired/superseded reconciliation obligation. No BLOCKING finding. (`t388`/`t392` are card ids,
  not SPEC ids; t392's SPEC is unpublished and the SPEC says so.)
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn syscall` over all four artifacts →
  no match. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' plan.md` → rc=1.
  `research.md` absent (Tier S).

The must-pass firewall is clean. The FAIL below comes from the rubric side, not the firewall.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.70 | between 0.50 and 0.75 | §4 is the clearest thing in the document — the two-direction table at `spec.md:210-213` plus the sentence naming the uncloseable direction as the one that caused the card. Every external citation I checked is exact: `Makefile:20,36,72`, `.goreleaser.yml:22`, doc heading `:66` and assertions `:8`/`:12`, `hook_flush_test.go:22`, `ci.yml:129`/`:208`, `system.yaml.tmpl:6,9`. Deducted for D1 — the SPEC asserts a cause for the M1 RED that its own [HARD] parser scope forbids, so a reader implementing `plan.md` M1 as written builds a check that cannot produce the failure `plan.md:90` promises — and for D2, a pointer to a section the reduction deleted. |
| Completeness | 0.85 | between 0.75 and 1.0 | All required sections present. Five `### Out of Scope — <topic>` sub-headings (`spec.md:256,264,269,274,280`), each with specific bullets, including an explicit refusal to reopen t392's axis and the operator-rejected mechanical-derivation option. Frontmatter complete. The split's justifying measurement is preserved rather than dropped. Deducted only for D3 — a false enumeration inside a `[HARD]` block written expressly for a downstream card to inherit. |
| Testability | 0.65 | between 0.50 and 0.75 | Zero weasel words (`적절\|합리적\|충분히\|적당\|reasonable\|appropriate\|adequate\|proper` over all three artifacts → rc=1). Every AC carries a RED-now cell and a green path per `verification-completeness.md` §2, and AC-VSG-004/005 carry explicit mutant probes. But the AC that the SPEC itself names as its evidence (`acceptance.md:129`, "그 관측이 이 AC의 증거다") cannot be observed as specified (D1); AC-VSG-006's third condition is a judgement call in the shape of a count (D4); and REQ-VSG-005's own failure predicate is circular, repaired only downstream in the AC (D5). |
| Traceability | 1.00 | 1.0 | 6 REQ ↔ 6 AC, bidirectional, verified by extraction and corroborated by a zero-finding `moai spec lint` on this SPEC. No orphan AC, no uncovered REQ. Renumbering after the split left no hole. |

Aggregate: arithmetic mean (0.70 + 0.85 + 0.65 + 1.00) / 4 = **0.80**.

---

## Defects Found

**D1** — `plan.md:83-84` vs `plan.md:90` / `plan.md:70-71` / `acceptance.md:120-129` — **the M1 RED
the SPEC records as its evidence cannot be caused by what the SPEC says causes it.** `plan.md`
M1.1 fixes the parser's anchor to the *stamp subheading*, and states the anchor string is held as a
constant "M2의 문서 수정과 짝을 맞춘다" — i.e. it targets the heading M2 will create. That scope
restriction is `[HARD]` in three places: REQ-VSG-004 (`spec.md:182`), `spec.md:243`, `plan.md:54-55`.
At M1 the document has no such heading — measured: `grep -n 'Documentation Files\|Configuration
Files' .moai/docs/version-management.md` → `70`, `76`, and `grep -n 'Files Requiring Version Sync'`
→ `66`; there is no stamp/artifact axis anywhere in the file.

So at M1 the parse yields zero paths and the check goes red on the **count** assertion
(AC-VSG-005), not on the ghost. But `plan.md:90` states the cause as "유령이 아직 목록에 있기
때문이다", `plan.md:70-71` (E3) as "검사 실패(유령 때문)", and `acceptance.md:126-127` as "오늘의
트리가 이 검사를 실패시킬 입력 그 자체". The ghost sits under `**Configuration Files:**`, which the
`[HARD]` scope restriction excludes.

The alternative reading — that M1 anchors on the current `### Files Requiring Version Sync` and
takes all six bullets — does not rescue it either: that check would also flag
`.moai/release-notes/vX.Y.Z.ko.md`, which is the placeholder `plan.md:121` names as an anti-pattern,
and the count would be 6 against an expected 7. Under either reading the M1 RED is a mixture the
SPEC never pins.

This matters more here than it would elsewhere, because the SPEC's own cited rule rules the
resulting evidence out: `verification-completeness.md` §1.1 [HARD] — a check is complete only when
its failure "has been executed and observed on a known failing input", and a green or red produced
by a check "matching nothing" is "uninterpreted output". A zero-parse RED is precisely the case
§1.1 excludes. The card would record, as its headline completion evidence, the one signal its own
authority says means nothing. — Severity: **major** — Class: **blocking** — Required fix: decide
which tree the ghost RED is observed on and pin the expected RED output. The cheap route is to
state that M1's check anchors on the post-M2 heading, accept that the M1 red is a zero-parse red,
and move the ghost observation into M2: land the doc restructure *without* removing the ghost,
observe the check name `internal/template/templates/.moai/config/config.yaml`, quote that output,
then remove the ghost and observe green. Whichever route is chosen, `acceptance.md` AC-VSG-004's
green path and `plan.md` M1.5 / E3 must state the same cause, and the expected RED text must be
written down before it is measured.

**D2** — `spec.md:123` — dangling cross-reference created by the reduction. The `[HARD] 계수 단위`
clause says "§3의 단위 고정 조항과 같은 이유로". §3 is now `## §3 요구 (GEARS)`
(`grep -n '^## '` → `174`), containing REQ-VSG-001..006 and the traceability table; it carries no
unit clause. `grep -n '단위' spec.md` returns only lines 121-124 (the citing clause itself) and 307
(unrelated, "단위 테스트"). The clause it points at was §3 in 0.2.0 and was deleted with the scope
section. — Severity: **minor** — Class: **blocking** — Required fix: drop the subordinate clause,
or restate the reason inline.

**D3** — `spec.md:150-156` — a false enumeration inside the `[HARD]` block written for t392 to
inherit. The block states "이 트리에는 그런 파일이 둘 있다" (files whose *name* carries a version
token, inflating the `-n` count). Measured in the deny-listed scope:

    cut -d: -f1 <the §2 -n output> | sort -u | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+'
    → 8 files:
      .moai/marketing/awesome-lists/github-release-v2.12.0-enhanced.md
      .moai/release/MIGRATION-v2.17.0.md
      .moai/release/RELEASE-NOTES-v2.15.0.md
      .moai/release/RELEASE-NOTES-v2.16.0.md
      .moai/release/RELEASE-NOTES-v2.17.0.md
      .moai/release/RELEASE-NOTES-v2.20.0.md
      .moai/release/v2.15.0-draft.md
      docs/design/v2.14.0-release-plan.md

The uncovered inflation is visible in the same histogram the SPEC prints: `v2.17.0` reads **65**
under `-n` and **25** under `-h` (`grep -c 'v2\.17\.0'` over the `-h` output), a 40-count gap the
"둘" claim does not account for. The two named files do explain the two deltas §2 quotes, so the
correction itself is sound; the enumeration around it is not, and it is the sentence a t392 designer
will read as the scope of the hazard. — Severity: **minor** — Class: **blocking** — Required fix:
replace "둘" with the measured 8-file set (or state the two as examples and give the count), and
note that `.moai/release/` is inside the deny-listed scope while `.moai/release-notes/` is not.

**D4** — `acceptance.md:166-167` — AC-VSG-006's third condition is not mechanically decidable.
"「목록이 더는 썩지 않는다」에 해당하는 서술이 **0건**" asks for a count of statements *amounting
to* a claim, with no literal named. Compare the parallel condition in AC-VSG-003 item 1
(`acceptance.md:88-89`), which names its two literals (`reads from git tags at build time`,
`via git describe`) and is therefore greppable. As written, M2 can satisfy AC-VSG-006 item 3 by
assertion — and this is the one condition standing between the card and the over-claim that the
card exists to correct. — Severity: **minor** — Class: **blocking** — Required fix: convert to a
decidable form — either a literal deny-list of phrases (e.g. "더는 썩지 않는다", "완전히 방지",
"보장한다" applied to the list) checked by grep, or an explicit positive requirement that the
paragraph carry the word "절반"/"partial" and name the uncaught direction, which is what §4 itself
already does and can be checked by presence rather than by absence.

**D5** — `spec.md:184` (REQ-VSG-005) — the requirement's second failure clause is circular as
written: "shall fail when that count is zero **or differs from the number of entries under the
version-stamp heading**". The count it reports *is* the number of entries it parsed under that
heading, so under the natural reading the comparison compares a value with itself and can never
fire. `acceptance.md:141-142` repairs it by pinning the expected value ("수정 후 **7**"), but the
requirement layer never says the check holds an independent expected count. Given that this is the
card's designated non-vacuity requirement (`spec.md:325-326`), the requirement should say what the
AC does. — Severity: **minor** — Class: **blocking** — Required fix: restate as "differs from the
expected entry count recorded in the check (7 after the M2 correction)".

**D6** — `acceptance.md:106` vs `acceptance.md:141-142` — the two ACs collide procedurally and no
ordering is specified. AC-VSG-004 seeds one extra path line under the stamp heading, which makes
the parsed count 8; AC-VSG-005 requires the check to fail when the count differs from 7. An
implementation whose count assertion is fatal (`require.Equal`) aborts before the existence checks
run, and then the `[HARD]` obligation at `acceptance.md:107` — "그 경로를 **이름으로** 지목한다" —
is not met even though the check correctly went red. The SPEC is meticulous about mutant-killing
elsewhere and this is the one place a passing-looking implementation can skip a `[HARD]` clause.
— Severity: **minor** — Class: **optional** — Required fix: state that the existence report is
non-fatal (`t.Errorf`, not `t.Fatalf`) so both assertions are observed in one run, or have
AC-VSG-004 replace rather than add a line.

**D7** — `acceptance.md:159` vs `:169-171` — AC-VSG-006's **Given** spans both the document and this
SPEC, but its RED-now measures only the document, and its green path concedes "SPEC 쪽은
`spec.md` §4가 이미 만족한다". Half the AC's subject is green before run-phase begins, so the
RED-now cell does not establish a failing input for the whole criterion. — Severity: **minor** —
Class: **optional** — Required fix: scope AC-VSG-006's Given to the document and note the SPEC half
as already satisfied, or split it.

**D8** — `spec.md:29` — the list items are cited as "항목 71-78행". Measured, the item lines are
71-74 and 77-78; 75 is blank and 76 is the `**Configuration Files:**` label. Read as a span it is
accurate; read as an item enumeration it is not, and the SPEC's own citation standard elsewhere is
exact to the line. — Severity: **minor** — Class: **optional** — Required fix: cite `71-74, 77-78`,
or say "구간 71-78행".

**D9** — `spec.md:182` (REQ-VSG-004) — carried forward from iter-1 D8 at the same severity. The
`**Where** a path is named under the release-artifact heading …` clause states a data condition, not
a capability gate / feature flag / static config, which is what GEARS `Where` denotes; `While` fits
the intent. Syntactically it still matches a GEARS pattern, so MP-2 passes. — Severity: **minor** —
Class: **optional** — Required fix: recast as `While`, or fold into the main clause.

---

## Rulings on the dispatch's questions

**1. Is the transfer to t392 clean?** — **Yes**, with one dangling pointer. Renumbering is
contiguous and verified by extraction; traceability is complete in both directions and corroborated
by a zero-finding lint on this SPEC; every surviving mention of the deleted machinery is a refusal
or a deletion record, not live scope; the justifying measurement is preserved and reproduces to the
digit. The one artifact of the reduction is D2 — a `[HARD]` clause pointing at a section that was
deleted with the scope it belonged to.

**2. The partial-guarantee statement (§4, REQ-VSG-006 / AC-VSG-006).** — **§4 states it plainly;
AC-VSG-006 does not make all of it checkable.** §4 is the strongest section in the document and I
would not ask for a word of it: it opens `[HARD]` with "이 카드가 착지해도 목록은 여전히 썩을 수
있다", tables the two directions against what the card does with each, and then states outright that
the direction it cannot close is the one that caused the card — naming the `hugo.toml` omission as
the actual incident. `spec.md:220-222` closes the loop by refusing the over-claim on the ground that
over-claiming is why the card exists. No reader finishes §4 concluding the list can no longer rot.
The defect is in the verification layer, not the honesty: AC-VSG-006's conditions 1 and 2 are
checkable (both directions named; t392 named — greppable), but condition 3 — the one that actually
prevents the over-claim — is an absence-of-equivalent-statements judgement with no literal named
(D4). The honesty requirement is stated well and then not given an instrument.

**3. Non-vacuity (REQ-VSG-005 / AC-VSG-005), and AC-VSG-004's RED-now.** — **The zero-parse hole is
closed at the AC layer and left open at the REQ layer; the RED→GREEN ordering is NOT specified
firmly enough.**

On the hole: AC-VSG-005 closes it. The expected count is pinned to a literal 7 rather than derived
from the same parse, so it is non-circular, and it kills the mutant AC-VSG-004 alone would leave
alive — `acceptance.md:152-153` names that mutant explicitly. The §E boundary cases
(`acceptance.md:177-183`) are consistent with it: heading present but empty → 0 → fail; backticked
paths → either stripped or a count mismatch, never a silent skip; the `system.yaml.tmpl` comment →
8 if wrongly counted, which the assertion catches. R-5 (`spec.md:330-333`) honestly names the
residual: a renamed heading with an unchanged item count still needs a human to move the constant.
The REQ layer is where it leaks — REQ-VSG-005's own comparison is circular (D5).

On AC-VSG-004's claimed RED-now property: the two component measurements are true and I reproduced
both — `grep -c` of the ghost path in the document returns `1` (rc=0) and `test -e` of that path
returns rc=1. What does not follow is the conclusion drawn from them. Today's tree is a failing
input only for a check that reads the ghost, and the `[HARD]` scope restriction stops the check from
reading it until M2 creates the stamp heading. That is D1, and it is the direct negative answer to
the question as asked: the transition is currently **asserted, not specified to be observed**.

**4. Scope discipline.** — **Held. Nothing grew back.** 6 REQ / 6 AC against a stated 8/8 ceiling.
The requirement set decomposes cleanly: 001/002/003 are the document fix, 004/005 are the one check
plus its non-vacuity guard, 006 is the honesty clause. No requirement reintroduces a token
predicate, a discriminator, a deny list, an exemption table, or a new package — verified by the
residue sweep. The single Go artifact is one file in an existing package with an existing precedent
(`internal/cli/deprecated_paths_text_reference_test.go`, confirmed present) and an existing root
helper (`repoRootFromCLITest` at `hook_flush_test.go:22`, confirmed). No new CI job — `ci.yml:208`
runs `go test ./...` inside the `test:` job that begins at `ci.yml:114`, so the new file is picked
up. `internal/cli/version_sync_list_test.go` does not exist (rc=1), so no name collision.

**5. AC mechanical checkability.** — **Four of six are fully decidable by command; two are not.**
AC-VSG-001 (set comparison against a literal), AC-VSG-003 (two named literals plus a presence
condition), AC-VSG-004 (seeded path, named output, revert), AC-VSG-005 (reported count against a
pinned 7) are all decidable, and the tree is pinned both by a document-level pin
(`acceptance.md:6-9`) and per-item RED-now cells. AC-VSG-002's Then-clause is a set-intersection
condition and decidable once the headings exist, though its judging command is not written out.
AC-VSG-006 item 3 is the one judgement call dressed as a check (D4). Zero weasel words across all
three artifacts.

**6. Anything the reduction broke.** — **Two things, both listed: D2 (dangling `§3` pointer) and D3
(a figure whose subject was widened without re-measuring the enumeration).** Everything else I
checked survived intact: §5's reasoning about the check's home still rests on premises that hold
(precedent file present, helper present, CI runner present, Template-First not triggered —
`find internal/template/templates -name 'version-management*'` → empty); §7 R-1's rewritten grounds
are correct in every particular (`actions/checkout` appears 7 times in `ci.yml`, 6 with
`fetch-depth: 0`, the `test:` job's checkout at `ci.yml:129` among them, `spec-lint.yml` carries no
`fetch-depth` → rc=1); R-2 and R-3 now correctly point at t392 rather than at deleted machinery; and
`verification-completeness.md` §1.1 / §1.3 / §2 / §2.1 all exist and say what the SPEC cites them as
saying — I read §1.1 and §1.3 to confirm the attributions rather than the section numbers alone.

---

## Recommendation

The score rose and the must-pass firewall is clean; this revision is materially better than 0.2.0.
The split was the right call and it was executed cleanly — the deleted axis left no live residue,
the measurement that justified deleting it was preserved rather than discarded, and §4 is a model of
the honesty this card is about. The FAIL is one defect deep.

That defect is the card's own subject turned on itself. This SPEC exists because a document claimed
more than it knew; D1 is the SPEC arranging to record, as proof its check works, a red signal that
its own cited rule classifies as uninterpreted output. Entering run-phase as written puts the
implementer at M1 in front of a check that parses nothing, with three documents telling them the
red they are looking at was caused by the ghost. The cheapest move at that moment is to write it
down as the ghost red.

Fix route, in order:

1. **D1 (major)** — pin where and how the ghost RED is observed, and write the expected RED output
   down before measuring it. Make `plan.md` M1.5, `plan.md` E3, and AC-VSG-004's green path state
   one cause, not two.
2. **D5, D4 (minor, blocking)** — give REQ-VSG-005 the expected count the AC already has; give
   AC-VSG-006 item 3 an instrument instead of a judgement.
3. **D2, D3 (minor, blocking)** — drop the pointer to the deleted §3 clause; correct the two-file
   claim to the measured eight.
4. **D6, D7, D8, D9 (minor, optional)** — orchestrator's discretion. D6 is the one worth taking: one
   sentence about non-fatal reporting keeps a `[HARD]` clause from being silently skippable.

This is iteration 2 of at most 3. The remaining work is specification wording confined to `plan.md`
M1, `acceptance.md` AC-VSG-004/005/006, and three lines of `spec.md`; no scope decision is pending
and no measurement needs re-taking.

---

## Appendix — verification commands

All executed in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t388` at HEAD
`9328a52422baa13fdb7d7fd0c8409151da3ba3c1`, branch `WT-version-sync-list`.

| Purpose | Command (abbreviated) | Observed |
|---|---|---|
| Tree identity | `git rev-parse --show-toplevel` / `HEAD` / `--show-current` | worktree t388, `9328a5242`, `WT-version-sync-list` |
| REQ enumeration | `grep -n 'REQ-VSG-' spec.md` / `grep -oE` over 4 artifacts | 001..006 at :176,178,180,182,184,186 |
| AC enumeration | `grep -n '^### AC-' acceptance.md` | 6 AC at :30,64,82,103,137,157 |
| Frontmatter | `sed -n '1,15p' spec.md` | 12 canonical fields + `tier: S`, `version: "0.3.0"` |
| Out of Scope | `grep -n '^### Out of Scope' spec.md` | 5 sub-headings at :256,264,269,274,280 |
| Residue sweep | `grep -rn -e versionstamp -e testdata -e D2-OPEN -e D4-NOTE -e 'NEEDS CLARIFICATION' -e syscall -e REQ-VSG-007 -e AC-VSG-007` | only refusal/deletion-record hits for `versionstamp`; all others no match |
| `eba919e44` residue | `grep -rn 'eba919e44' spec.md` | :23 HISTORY, :51 incident, :259 Out of Scope, :348 §8 — no live fixture |
| D7 cross-SPEC | SPEC-ID extraction over 4 artifacts | self-reference only |
| MP-7 | `grep -rn 'NEEDS CLARIFICATION' plan.md` | rc=1 |
| Weasel scan | `적절\|합리적\|충분히\|적당\|reasonable\|appropriate\|adequate\|proper` over 3 artifacts | rc=1 |
| §2 area | `git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+'` + §2 deny-list | **2225 lines / 592 files** |
| §2 histogram | same with `-hoE`, `sort\|uniq -c\|sort -rn` | 270/83/80/80/72/68, total **2494** — exact match |
| `-n` inflation | same with `-nE` piped to `grep -oE` | 112 v2.14.0 / 90 v2.12.0 — mechanism reproduced |
| D3 enumeration | in-scope files with a version token in the path | **8 files**, not 2 |
| D3 uncovered case | `grep -c 'v2\.17\.0'` over `-h` output vs `-n` count | 25 vs 65 |
| Path-token match counts | `git grep -cE` on the two named files | `v2.14.0-release-plan.md:40`, `github-release-v2.12.0-enhanced.md:7` |
| Authoritative set | `git show --numstat --format= 61921f1ba` | 7 paths, 9 lines — matches the AC's literal set |
| D4 closure | `git merge-base --is-ancestor 61921f1ba HEAD` / `git branch -a --contains` | rc=1 / `release/v3.1.4` only |
| Ghost in doc | `grep -c '<ghost path>' .moai/docs/version-management.md` | `1`, rc=0 |
| Ghost on disk | `test -e internal/template/templates/.moai/config/config.yaml` | rc=1 |
| Placeholder | `test -e .moai/release-notes/vX.Y.Z.ko.md` | rc=1 (real files: `v3.1.0.ko.md`, `v3.1.3.ko.md`) |
| Synthetic input | `test -e docs-site/nonexistent-stamp.toml` | rc=1 |
| D1 heading absence | `grep -n 'Files Requiring Version Sync\|Documentation Files\|Configuration Files'` | :66, :70, :76 — no stamp/artifact axis |
| Four omissions | `grep -n` for the four paths in the doc | only `version.go` at :8/:12 (the SSOT assertions), none in the list |
| D2 dangling ref | `grep -n '단위' spec.md` + `grep -n '^## ' spec.md` | :123 cites `§3`; §3 at :174 is `요구 (GEARS)` |
| §1.1 citations | `sed -n '1,10p' pkg/version/version.go`, `Makefile`, `.goreleaser.yml:22` | `var ( Version = "v3.1.3"`; all injection points exact |
| §1.2 render claim | `sed -n '1,10p' …/system.yaml.tmpl` | `{{.Version}}` at :6 and :9 |
| R-1 grounds | `grep -c 'actions/checkout'` / `grep -c 'fetch-depth: 0'` on `ci.yml`; `sed -n '125,132p'`; `sed -n '205,210p'` | 7 / 6; `:129` fetch-depth in `test:` job; `:208` `go test ./...` |
| R-1 spec-lint | `grep -n 'fetch-depth' .github/workflows/spec-lint.yml` | rc=1 |
| §5 precedents | `test -e …/deprecated_paths_text_reference_test.go`; `sed -n '20,24p' hook_flush_test.go` | rc=0; `repoRootFromCLITest` at :22 |
| Path collision | `test -e internal/cli/version_sync_list_test.go` | rc=1 |
| Template-First | `find internal/template/templates -name 'version-management*'` | empty |
| Cited rule sections | `grep -n '^## \|^### ' verification-completeness.md`; read §1.1, §1.3 | §1.1/§1.3/§2/§2.1 exist and match the SPEC's use verbatim |
| SPEC lint | `moai spec lint` (binary built from this tree) → `grep -c 'SPEC-VERSION-STAMP-GUARD-001'` | **0 findings for this SPEC** (1096 repo-wide warnings, none here) |
| SPEC lifecycle | `mcp__moai__spec_audit` (`project_root` = this worktree) | 1 INFO `EraAutoDetected`, no drift |
