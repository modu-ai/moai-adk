# SPEC Review Report: SPEC-VERSION-STAMP-PREDICATE-001

Iteration: 3/3 (operator override of the Tier M two-iteration ceiling, this round only)
Tree: worktree `.claude/worktrees/t392`, HEAD `9a3e2dabe`, branch `WT-version-stamp-predicate`
Artifact version: `0.3.0` — 15 requirements / 15 acceptance criteria

Verdict: **PASS-WITH-DEBT**
Overall Score: **0.90** (Tier M threshold 0.80 — `spec-workflow.md:141`)
Delta from iter-2: **+0.06** (0.84 → 0.90). Monotone increase; no regression, so the STOP
escalation clause does not fire.

Reasoning context ignored per M1 Context Isolation. Everything below is measured against the four
artifacts and the tree, not against the author's account of them. Where a number is taken on
attribution rather than re-executed, the attribution is named.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-VSP-[0-9]*' spec.md | sort -u` returns
  exactly `REQ-VSP-001 … REQ-VSP-015`, 15 ids, zero-padded, no gap, no duplicate.
  `grep -c '^\*\*REQ-VSP-' spec.md` → `15`, so every id is also a definition, not just a
  reference. Acceptance ids: `grep -o '^### AC-VSP-[0-9]*' acceptance.md | sort | uniq -c` →
  15 headings, each count 1.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer only**
  (`spec.md` §4, lines 317-390). All 15 are ubiquitous (`The check shall …`, REQ-VSP-001/002/
  005/006/007/008/009/010/011/012/014/015) or event-driven (`When the sweep finds …`,
  REQ-VSP-003; `When a registry entry classified 'stamp' does not carry …`, REQ-VSP-004; `When
  an entry does not resolve …`, REQ-VSP-013). No IF/THEN. No informal modal in normative text.
  The Given-When-Then entries in `acceptance.md` are the verification layer and are graded under
  Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — `spec.md:1-15` carries all 12 canonical fields
  (`id`, `title`, `version: "0.3.0"` quoted, `status: draft`, `created`/`updated` ISO,
  `author`, `priority: Medium`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` as a
  comma-separated string) plus `tier: M`. No rejected snake_case alias
  (`created_at`/`updated_at`/`labels`/`spec_id`) present.
- **[N/A] MP-4 Section 22 language neutrality** — the SPEC is scoped to this repository's own Go
  test surface and its own docs; it makes no multi-language tooling claim. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'`
  over all three artifacts yields exactly two ids: itself and `SPEC-VERSION-STAMP-GUARD-001`.
  `grep '^status:' .moai/specs/SPEC-VERSION-STAMP-GUARD-001/spec.md` → `status: completed` —
  not in {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md plan.md acceptance.md`
  → `0 0 0`. D8 auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' plan.md spec.md
  acceptance.md progress.md` → rc=1, no match.

---

## What I re-ran independently this round

Every number below was produced by a command I issued in this session, in this worktree, at HEAD
`9a3e2dabe`. Attributed-not-re-executed items are marked.

### The eight-mutant battery (N2) — re-run, confirms the author's claim

```
$ sed -n '265,290p' plan.md | sed 's/^> \{0,1\}//' > pinned.txt
$ grep -cEif .moai/reports/t392/scratch/re2.txt pinned.txt
0
rc=1

$ grep -nEif .moai/reports/t392/scratch/re2.txt .moai/reports/t392/scratch/mut8.txt
1:It does not catch three cases.
2:Three cases remain uncaught.
3:The following three cases remain.
4:There are three things it still does not catch.
5:Still uncaught: three.
6:The list below is exhaustive.
7:It fails to detect 3 kinds of site.
8:여전히 잡지 못하는 것이 셋이다.

$ grep -cEif .moai/reports/t392/scratch/re2.txt .moai/docs/version-management.md
0
rc=1
```

8/8 mutants matched — including mutants **1** and **5**, the two English forms the iter-2 regex
was blind to. The real pinned text matched 0. The current document matched 0, so the regex is not
a false-positive generator on prose it was not written against. Both directions observed in the
same session, so "the regex sees nothing" is excluded.

This is a genuine repair of the iter-2 N2 defect, not a re-fit to one mutant: the mutant set is
declared **in `acceptance.md:295-306` before the measurement**, and `acceptance.md` §D.2 makes
running the whole set a Definition-of-Done item ("하나만 걸리는 것으로는 통과로 세지 않는다").

### The eleven pinned phrases (N3 / grep-fragility) — re-run

```
$ bash .moai/reports/t392/scratch/check.sh pinned.txt
1  aged-out token
1  registered as `prose`
1  inlined inside a file the exclusion set hides
1  not a version token at all
1  renders the version rather than carrying it
1  the repository does not track
1  this list is not exhaustive
1  version_stamp_registry_test.go
1  edits the registry in that same commit
1  the check fails naming the path
1  A version bump does not touch the registry

$ bash .moai/reports/t392/scratch/check.sh .moai/docs/version-management.md
0  (all eleven)
```

Two things are established at once. Each of the eleven phrases occurs exactly once and on a single
line inside `plan.md` §E — the one-line measurement is real, not asserted. And all eleven are 0 in
the current document, so the pre-pinned RED covers AC-VSP-014's four phrases as well as
AC-VSP-011's seven, which the earlier rounds could not say.

The six judgment greps of AC-VSP-011(b) each match a phrase that exists in the pinned text. I
checked the mapping item by item against `plan.md:271-282`: item 1 → `aged-out token`, item 2 →
`registered as \`prose\``, item 3 → `inlined inside a file the exclusion set hides`, item 4 →
`not a version token at all`, item 5 → `renders the version rather than carrying it`, item 6 →
`A file the repository does not track`. No orphan grep.

### The open-form marker — load-bearing, not decorative

`grep -cF 'this list is not exhaustive'` is a **positive** assertion, and it is doing work the
negative regex cannot: the regex alone returns 0 on an empty file, a deleted paragraph, and a
missing file identically. AC-VSP-011 states this explicitly ("닫힌 수의 부재만 보면 파일이 없거나
문단이 통째로 사라진 경우와 구별되지 않는다") and pairs the two. The pinned text carries the
marker at `plan.md:273` inside the sentence `**At least the following remain, and this list is not
exhaustive.**`, and it keeps t388's refusal sentence at `plan.md:285` (`None of this means the
list can no longer rot.`). Verified: `check.sh` row 7 → 1 in the pinned text, 0 in the current
document. **Load-bearing.**

### The document baseline (N3 premise)

```
$ grep -cP '[가-힣]' .moai/docs/version-management.md
0        rc=1
$ wc -l .moai/docs/version-management.md
104
$ grep -n 'does not detect' .moai/docs/version-management.md
90:The guarantee it establishes is **partial**. …
$ grep -c 'docs-site/content' .moai/docs/version-management.md
0        rc=1
```

English-only confirmed. AC-VSP-014's fifth command (prose-path leak → 0) holds at baseline, so
its bidirectional mutant is meaningful rather than trivially satisfied.

### The D3-residual measurement — reproduced exactly

```
$ git grep -lF version-management.md -- '*.go'
internal/cli/version_sync_list_test.go
$ git grep -lF 'docs-site/content' -- '*.go'
scripts/convert-nextra-to-hextra/main.go
scripts/docs-version-snapshot/main.go
```

Byte-identical to `spec.md` §3.1's recorded output. The inference the SPEC draws from it —
that zero of the 21 prose entries has a consumer test that a bump miss would break, so today's
silence on a misclassified prose entry is maximal — follows from the measurement.

### Registry composition (used below for NEW-3)

Taken from `spec.md` §2.2 and cross-read against `.moai/docs/version-management.md:72-80`. The 7
stamps are `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`,
`.moai/config/sections/system.yaml`, `docs-site/hugo.toml`, `pkg/version/version.go`; the 21
prose entries are 20 under `docs-site/content/` plus `.moai/docs/version-management.md`. The
registry's top-level footprint is therefore `README*` (root), `.moai/config/`, `.moai/docs/`,
`docs-site/`, `pkg/version/`. **No registry path lies under `internal/`, `cmd/`, `scripts/`,
`.github/`, or `.claude/`.** This is a read of the two lists, not a re-derivation of the 28.

### Attributed, not re-executed

- The disjointness of the six exclusion groups (121 measured as a single pathspec) is attributed
  by `spec.md:198-201` to the iter-1 plan-audit and explicitly marked as not re-run this session.
  The SPEC labels it correctly; I did not re-run it either.
- `moai spec lint` rc=0 with zero findings on this SPEC ID: the operator reports independently
  re-verifying it. My own run of `moai spec lint` exited 0 for the whole catalogue
  (`0 error(s), 1096 warning(s)`), and `mcp__moai__spec_audit` filtered to this SPEC ID with
  `project_root` set to this worktree returned a single `INFO`/`EraAutoDetected` row and no
  drift or error finding. I did not isolate this SPEC's own warning rows out of the 1096; that is
  a gap in my measurement, not a claim of cleanliness.

---

## N1 — is the residual honestly bounded, or is it my N1 under a new name?

**Honestly bounded. It is strictly narrower than N1, and the narrowing is real.**

N1 was: the sweep could lose everything except the 7 stamps and stay green. Trace the new pair
against exactly that state. Population = the 7 stamp paths. Assertion (가) is `registry(28) ⊆
population`; the 21 prose entries are registry members and are not in that population, so (가)
fails and names all 21 paths. **The case I raised is caught, by path name, not by count.**

The residual named at `spec.md:557-563` (§6.2) and `spec.md` §8 R-9 is a different set:
population = *exactly* the 28 registry paths. There (가) passes because the registry is contained,
and (나) passes because judged and handed are both 28. That state is a strict superset of N1's
state and a strict subset of the true population — a genuine remainder, not a relabelling.

Two further things make the bound honest rather than convenient:

- The SPEC does not claim the pair closes it. `acceptance.md:449-451` states outright that
  AC-VSP-015 "그것을 잡는다고 주장하지 않는다", and R-9's `관측하지 않음` line concedes the
  residual was never turned into an actual mutant — "실행으로 보이지 않았다". That is the correct
  form of an unobserved claim under the evidence rule.
- The rejection of my suggested `git ls-files → 10048` literal is right, and I withdraw the
  suggestion. `spec.md:539-545` distinguishes bump-invariance from development-invariance: the
  number is invariant across the two measured bump commits but moves on every ordinary
  file-adding commit, so pinning it produces a bare-integer failure on an unrelated commit and a
  cheap repair path (edit the number). That is the same defect class D2 removed, in a worse
  place, and `plan.md` AP-11 forbids the return. The author's reasoning is better than my
  suggestion was.

What I do flag is an asymmetry in *which* instance is named — see NEW-3.

---

## Self-reported defect (the §6.1.1 → §6.2 relocation) — checked, not taken on report

The fix holds.

- `grep -n '^#\{1,4\} ' spec.md` shows `### §6.1` at line 481 and `### §6.2` at line 527. The
  maintenance contract block (`**계약.**` and its four bullets, lines 508-521) plus the two
  bump-invariant constants table (lines 504-507) both fall in `[481, 526]` — **inside §6.1's
  scope**, ahead of the §6.2 heading. The reach material begins at 527 and does not straddle.
- REQ-VSP-014 (`spec.md:381-383`) requires the documentation to state who edits the registry, the
  obliging event, what the check reports in between, and that a bump obliges no edit. All four
  are present in §6.1's contract bullets (누가 / 언제 / 그 사이 검사는 무엇을 말하는가, and the
  §6-사유-1 bullet), and AC-VSP-014's four judgment greps map onto the four pinned English
  phrases I measured above. **REQ-VSP-014 points at live text.**
- Nothing else was displaced. `grep -n '§6' spec.md plan.md acceptance.md` returns 22 references;
  I read each. Every citation of §6.1 refers to the constant-removal argument or the maintenance
  contract (`plan.md:81, 186, 239, 332, 373`; `spec.md:529, 544, 551, 675, 687`), and every
  citation of §6.2 refers to the reach argument (`plan.md:336`; `spec.md:589, 706, 710`;
  `acceptance.md:440, 445`). No reference resolves to the wrong section.
- **Zero `§6.1.1` strings remain** in any of the three files — I re-measured rather than
  accepting the operator's statement: the heading grep above enumerates every heading and no
  `§6.1.1` appears, and no `§6.1.1` occurs in the 22-line cross-reference sweep.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.88 | 0.75-1.0 | The two iter-2 ambiguities are gone: the pinned §E text is English against an English-only target (`plan.md:255-258` + measured `grep -cP '[가-힣]' → 0`), and §3 row 4 is qualified with a two-row entry-state table (`spec.md:268-278`) that routes ghost entries to REQ-VSP-013 and exclusion-group registration to REQ-VSP-015. Held below 1.0 by NEW-1 (two conflicting pinned strings for one release-blocking assertion), NEW-4 and NEW-5. |
| Completeness | 0.94 | 1.0 band, one notch down | All sections present. Out of Scope carries 9 `### Out of Scope — <topic>` H3 sub-headings (`spec.md:568-638`), each with specific `-` bullets, including the new `모집단 크기의 하한` entry at 583. §8 now carries R-9 with an explicit `관측 / 관측하지 않음` split (`spec.md:704-710`). The one structural hole of iter-2 (no floor on sweep reach) is closed by REQ/AC-VSP-015. Held below 1.0 only by NEW-1's missing §D.2 row. |
| Testability | 0.90 | 0.75-1.0, upper | Three declared mutant sets, all bidirectional and all Definition-of-Done items (`acceptance.md:481-484`): AC-VSP-010 four, AC-VSP-011 eight, AC-VSP-015 three. The eight-mutant battery reproduced 8/8 here. Every release-blocking AC pins a RED expectation. Held below higher by NEW-1: the newest release-blocking assertion is the one whose expected failure string is doubly pinned and absent from the message contract. |
| Traceability | 0.97 | 1.0 band | `spec.md` §4.1 maps all 15 REQ → 15 AC, one to one. No orphan AC, no uncovered REQ (verified by id extraction above). `moai spec lint` returned 0 errors catalogue-wide; the per-SPEC warning isolation is the gap noted under Attributed. |

Aggregate **0.90**, above the Tier M threshold of 0.80.

---

## Defects Found (structured defect-list)

**D1. NEW-THIS-ROUND — conflicting pinned failure string for REQ-VSP-015, and no §D.2 row**
`plan.md:100` / `acceptance.md:432` / `plan.md:206-219` — Severity: **major** — Class: **blocking**
`plan.md:100` (M2 synthetic-RED table) pins the 015-나 expected failure as
`` `judged=N examined_of=M` ``. `acceptance.md:432` pins the same assertion's RED as
`` `judged=27 handed=28` ``. The second key differs: `examined_of` vs `handed`.
`plan.md` §D.2 is the artifact's own [HARD] message contract — "RED을 관측하기 전에 기대 실패
문자열을 고정한다. 예측이 틀려도 고정되어 있으면 조사가 강제된다" — and its table carries six
rows (003, 004, 005, 006, 009, 013) with **no row for 015**, although AC-VSP-015 is classified
**릴리스 차단** at `acceptance.md:462`. So the newest release-blocking assertion is the one
assertion whose string is pinned twice, inconsistently, and nowhere in the contract that exists to
pin it. A run-phase agent must choose between two strings with no authority to break the tie, and
whichever it picks, one artifact's pinned expectation is falsified by its own implementation —
which is precisely the "fix the expected signal before measuring" discipline this SPEC invokes.
This is the one finding I consider blocking.
Required fix: add a `도달 범위(015)` row to the `plan.md` §D.2 table carrying both strings —
`registry path missing from population: <path>` and one canonical count form — and edit whichever
of `plan.md:100` / `acceptance.md:432` disagrees so a single key name (`handed` or `examined_of`,
the choice is free) appears in all three places.

**D2. NEW-THIS-ROUND — the `현 L83 · L90` citation does not resolve**
`acceptance.md:257` and `plan.md:292` — Severity: minor — Class: optional
Both places justify the choice of the six judgment phrases by noting that `_test.go` and
`system.yaml.tmpl` already exist elsewhere in the target document, citing "현 L83 · L90".
Measured: `grep -n 'system\.yaml\.tmpl' .moai/docs/version-management.md` → **82**;
`grep -n '_test\.go' .moai/docs/version-management.md` → **88**. Neither cited line number is
correct. The *inference* is sound and load-bearing — I confirmed both tokens do occur in other
sections, which is exactly why they are unusable as judgment phrases — so only the coordinates
are wrong. But a file:line citation that does not resolve is a defect on this repository's own
terms, and these two coordinates are the sole evidence offered for a decision that shaped six
judgment commands.
Required fix: replace `현 L83 · L90` with `현 L82 · L88` in both files, or drop the line numbers
and cite the section names instead.

**D3. NEW-THIS-ROUND — R-9 names the hard-to-reach instance and not the easy one**
`spec.md:557-563` (§6.2) and `spec.md:704-710` (R-9) — Severity: minor — Class: optional
The residual is stated as "등록부 경로만 담고 나머지를 버리는 드라이버", with the mitigating
argument "그런 드라이버는 손으로 그렇게 써야만 나오고". That is true of that instance. But the
same pair is also passed by a far cheaper mutation of the kind §6.2 itself identifies as the
motivating hazard: **an exclusion-group literal covering a directory that contains no registry
path.** Measured composition (above): the registry's footprint is `README*`, `.moai/config/`,
`.moai/docs/`, `docs-site/`, `pkg/version/`. Adding one line — `internal/`, or `cmd/`, or
`scripts/` — to the exclusion enumeration drops that entire subtree from the population while
(가) still holds (no registry path is under it) and (나) still holds (both sides move together).
REQ-VSP-003 then goes silently blind over the whole Go tree. This is one literal edit, not a
hand-written driver.
The SPEC asserts nothing false about it — `spec.md:562-563` says "REQ-VSP-015가 그 반경을 등록부가
덮는 만큼 줄이지만 전부는 아니다", which is exactly right, and the `docs-site/` instance is named
concretely because the registry *does* cover it. So this is an asymmetry of sharpness, not an
error. It is optional for that reason.
Required fix (optional): in R-9's `관측` paragraph, add one sentence naming the reachable form —
"제외 그룹이 등록부 항목을 하나도 담지 않은 서브트리(`internal/` · `cmd/` · `scripts/`)를
덮으면 두 단언 모두 통과한다" — so the residual's cheap instance is named alongside its expensive
one.

**D4. NEW-THIS-ROUND — the exclusion set's matching semantics is undecided**
`spec.md:203-215` (§2.3 table) and `plan.md:124` (M3) — Severity: minor — Class: optional
The six groups have four different shapes: directory prefix (`.moai/reports/`, `.moai/specs/`,
`.moai/release-notes/`), an exact root path (`CHANGELOG.md`), a basename suffix (`*_test.go`), and
a path glob with a wildcard segment (`docs-site/content/*/changelog*`). All six were **measured**
as git pathspecs, where `:!*_test.go` matches at any depth. `plan.md` M3 says only "제외 그룹
여섯을 리터럴로 열거한다" and decides nothing about how a Go implementation matches them. A naive
`filepath.Match` matches no `*_test.go` path (Go's `*` does not cross `/`); a naive
`strings.HasPrefix` matches no glob group. The severity is minor rather than major *because of*
the iter-3 repair: an under-exclusion surfaces loudly through REQ-VSP-003 naming the extra paths,
and an over-exclusion of a registry-covered region surfaces loudly through REQ-VSP-015(가) naming
the dropped registry paths. Only the D3 case stays silent.
Required fix (optional): add one line to `plan.md` M3 stating the matcher — e.g. "각 그룹은 git
pathspec 의미로 대조한다: 접두사 그룹은 경로 접두사, `*_test.go` 는 basename 접미사,
`docs-site/content/*/changelog*` 는 세그먼트 단위 glob".

**D5. NEW-THIS-ROUND — AC-VSP-010(a)'s judgment command is a placeholder, not a command**
`acceptance.md:221` — Severity: minor — Class: optional
`grep -nE '\bexec\.|os/exec' <검사 파일의 순수 코어 블록>` is the only judgment command in the 15
ACs whose target is prose rather than a literal path or an awk-boundable range. `plan.md` M1
commits to a pure function and names its output as "함수 시그니처" but pins no function name, so
the block boundary does not exist until run-phase invents it. Everything else in AC-VSP-010,
including the four declared mutants, is runnable as written.
Required fix (optional): pin the core function's name in `plan.md` M1 and express (a) as an
awk-bounded extraction over that name, so the command runs without a run-phase decision.

**Accepted debt, not a defect — the one-line-per-phrase [HARD] rule.**
`plan.md:296-302` and `acceptance.md:325-329` impose a discipline on how a human or agent pastes
text. It is a wish, not a mechanism: nothing in the check enforces it, and a later editor
re-wrapping the paragraph violates it silently. I do not score it as a defect for one reason,
which the SPEC states and I verified is the correct reason — the failure mode is a **false RED**,
not a false GREEN. A re-wrap makes `grep -cF` return 0 on text that is actually present, the M5
verification fails loudly, and a human looks. A control whose only failure is self-announcing is
acceptable as a discipline. The stronger form (normalise newlines before matching, e.g.
`tr '\n' ' '` into the grep, or `grep -z`) was available and not taken; that is a legitimate cost
choice, and the SPEC discloses the fragility rather than hiding it. If a future round wants to
close it, that is the mechanism.

---

## Regression Check (defects from iter-2)

| iter-2 finding | Status | Evidence |
|---|---|---|
| **N1** — the D2 subtraction removed the only floor on sweep reach | **RESOLVED** | REQ/AC-VSP-015 (`spec.md:385-390`, `acceptance.md:422-455`). My exact N1 state (population = 7 stamps) is caught by (가), naming the 21 missing registry paths. The remainder is strictly narrower and is named at §6.2 and R-9. The suggested `10048` literal is correctly rejected on development-variance grounds; I withdraw it. |
| **N2** — the closed-count regex was fitted to its own author's single mutant; 6 of 7 fed mutants escaped | **RESOLVED** | Eight-mutant set declared at `acceptance.md:295-306` **before** the measurement, and re-run by me here: 8/8 matched, including the two English forms (mutants 1 and 5) the iter-2 regex missed. Real text 0, current doc 0. Whole-set execution is a DoD item (`acceptance.md:483-484`). |
| **N3** — Korean replacement text pinned into an English-only document | **RESOLVED** | §E rewritten in English (`plan.md:265-290`), premise re-measured (`grep -cP '[가-힣]' → 0`, 104 lines). t388's refusal sentence retained verbatim at `plan.md:285`. Six-item open form, marker positive-asserted. Judgment phrases re-pinned to English strings that exist only in the new text; all eleven measured at 1 in §E and 0 in the current document. |
| **N4** — §3 row 4 conflated file-state with entry-state | **RESOLVED** | Row 4 qualified "파일이 실재하는 항목에 한한다" (`spec.md:249`); two-row entry-state table added at `spec.md:274-277` routing ghost entries to REQ-VSP-013 and exclusion-group registration to REQ-VSP-015, with the reason the second matters (`스윕 ⊇ 스탬프` would fail permanently). |
| **N5** — the denylist of forbidden git subcommands omitted six live ones | **RESOLVED** | AC-VSP-010(b) inverted to an allowlist (`acceptance.md:217-238`): exactly one `"git"` literal, exactly `[]string{"git", "ls-files"}`, backed by `plan.md` M3 [HARD] and AP-12. The discriminating mutant (`ls-tree -r HEAD`, which passed the iter-2 denylist) is declared as mutant 1. The argument that an allowlist has nothing to omit is correct. |
| **N6** — bidirectional check required touching the real git index | **RESOLVED** | AC-VSP-012 now uses two synthetic populations against the pure core (`acceptance.md:341-352`), with AP-13 forbidding the return. Consistent with §6's pure-core separation and with this repository's shared-checkout doctrine. |
| **D6** — the §2.3 delta pinned integers that had already rotted three times | **RESOLVED** | Integers removed; replaced by a rule (`spec.md:207-215`, `plan.md:56-62`) that is invariant under the act of auditing the card. The reason is written in: auditing this card enlarges the population, so a pre-committed integer is structurally guaranteed to rot. This is the round's best move — repair by subtraction with the reason recorded. |
| **D3-residual** — adjacent measurement not taken | **RESOLVED** | Taken and recorded at `spec.md:296-310`; reproduced byte-identically here. Zero of the 21 prose entries has a consumer test. |

No iter-2 defect is carried over. No defect has appeared in all three iterations, so the
stagnation flag does not fire.

---

## The three specific judgments asked for

**Monotonicity.** 0.84 → **0.90**, +0.06. The rise is concentrated in Testability and
Completeness and is attributable to three closures I verified by execution rather than by report
(N1's reach pair, N2's battery, N3's phrase set). Clarity rose less than it otherwise would have
because the iter-3 addition introduced D1. No dimension fell. The STOP-on-regression clause does
not apply.

**The 15/16 ceiling — is 15 honest?** Yes, with one disclosed edge. Requirement granularity here
is coarser than one-assertion-per-requirement in four places: REQ-VSP-002 (registry shape +
classification rule + superset property), REQ-VSP-005 (two constants + sweep⊇stamp + the negative
clause), REQ-VSP-011 (two documents + five enumerated cases), REQ-VSP-015 (two assertions + the
negative clause). Three of those four predate any ceiling pressure — they existed at counts 11 and
14, well clear of 16 — so they are the author's habitual granularity, not cap-driven merging. Only
REQ-VSP-015 was authored at 14→15 with the cap one away, and splitting it would have hit 16
exactly. But its two assertions share one purpose (sweep reach), one acceptance criterion, one
milestone (`plan.md` M3), and one test function, which makes them a defensible single unit on the
merits. Decisive for me: the author **disclosed** the position rather than hiding it — `spec.md`
§5 states "요구·수락 15/15는 상한 16 바로 아래다" and adds the [HARD] tier-up note. A cap-driven
merge conceals; this one announces. I read 15 as honest.

**Is the SPEC implementable as written?** Nearly. A run-phase agent handed these four documents
can execute M1 through M5 without inventing design, with **one decision the SPEC did not make and
three it left loose**:

1. **The decision it did not make — D1.** Which failure string does REQ-VSP-015's count assertion
   emit? `plan.md:100` and `acceptance.md:432` disagree, and the artifact's own message contract
   (`plan.md` §D.2) has no row for it. The agent cannot resolve this from the artifacts; it must
   pick, and picking falsifies one pinned expectation. This is why the verdict carries debt.
2. **Loose — D4.** How the six exclusion-group literals are matched against `git ls-files` output.
   Four different pathspec shapes, no matcher stated. Recoverable at run-phase, and most
   mis-implementations now fail loudly thanks to REQ-VSP-015.
3. **Loose — D5.** What text AC-VSP-010(a)'s grep is run over. Needs a function name that M1
   promises but does not pin.
4. **Loose — the golden regeneration diff.** `plan.md` M4 says the diff must be read and that
   "버전 줄 하나만 바뀌어야 한다", with the per-golden occurrence measured at 1. That is a human
   judgment step with no pinned command; it is correctly a judgment and I do not count it as a
   gap, only note it so the run-phase does not treat it as mechanical.

Everything else — the registry contents, the two constants, the seven synthetic REDs, the three
mutant sets, the pinned English replacement text, the eleven judgment phrases, the DoD — is
specified to the level where an implementer types rather than decides.

---

## Recommendation

**Verdict: PASS-WITH-DEBT at 0.90, above the Tier M threshold of 0.80.**

This SPEC is ready for the Implementation Kickoff Approval gate **once D1 is closed**. D1 is a
one-edit fix (add the `015` row to `plan.md` §D.2 and unify one key name across three lines) and
does not require another audit round: it is verifiable by a single grep that the three sites agree,
and the orchestrator can confirm that itself without re-opening the verdict.

D2 through D5 are optional. Routing all four into a revision would produce a fourth iteration on a
SPEC whose defect count has fallen from eight to five and whose remaining five are, in aggregate,
worth less than the cost of another round — and this document's own over-engineering brake exists
for exactly that arithmetic. D2 is a two-character fix worth taking opportunistically. D3 sharpens
an already-correct statement. D4 and D5 are answered by the first hour of run-phase and their
failure modes are loud.

Three things I want on the record for whoever runs this:

- The iter-3 repairs are **verified, not reported**. I re-ran the eight-mutant battery, the
  eleven-phrase one-line measurement, the pre-pinned RED, the Korean-character baseline, and the
  consumer-test measurement in this session at HEAD `9a3e2dabe`. All reproduced.
- The one thing I did **not** verify and no one has: R-9's residual has never been made into a
  mutant. The SPEC says so itself. That gap is correctly labelled, and D3 argues it is wider in
  practice than R-9's wording suggests.
- The per-SPEC isolation of `moai spec lint` warnings for this SPEC ID is attributed to the
  operator, not measured by me. My run exited 0 catalogue-wide; I did not extract this SPEC's own
  rows from the 1096 warnings.

Iteration 3 of 3 under the operator override closes here. Whether the cap extends further is the
operator's decision, and I see no defect that would justify asking for it.
