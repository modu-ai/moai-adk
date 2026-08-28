# SPEC Review Report: SPEC-AC-COUNT-DISCRIMINATOR-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.67** (Tier M threshold 0.80)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t338`, branch `WT-ac-count-sweep`, HEAD `da03d9188`

Reasoning context ignored per M1 Context Isolation. Judged from the four artifacts plus the tree.

Cross-model second opinion NOT run: the SPEC directory is untracked (`git status --porcelain` → `?? .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/`), so both `codex_audit` and `glm_audit` collect an empty diff and return `inconclusive` by construction. Recorded as a gap, not a pass.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE '\*\*REQ-ACD-[0-9]+\*\*' spec.md | sort | uniq -c` → REQ-ACD-001…008, each exactly 1, 3-digit zero-padded, no gap, no duplicate.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §4, REQ-ACD-001…008). All 8 match a GEARS pattern: 001/002/005/006/008 Ubiquitous, 003 Event-driven (`When … shall halt`), 004 Unwanted (`shall not`), 007 `Where`-form. The `AC-ACD-*` entries in `acceptance.md` are Given-When-Then; that is the correct verification-layer format and is graded under Group 4, not here. REQ-ACD-007's `Where` is used as an event trigger rather than a capability gate — form-conformant, pattern-semantics off; recorded as D12 (minor), not an MP-2 failure.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present (`id title version status created updated author priority phase module lifecycle tags`) plus optional `tier: M` and `era: V3R6`. `grep -nE '^(created_at|updated_at|labels|spec_id):' spec.md` → rc=1 (no rejected alias).
- **[N/A] MP-4 Section 22 language neutrality** — the template-bound payload is the B12 clause, whose counter is a `grep`/`sort`/`wc` pipeline (language-neutral); the verifier is scoped to this repository's own Go tree. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 8 referenced SPEC IDs extracted; all 8 directories exist; statuses: `completed`×5 (AGENT-EMIT-LINEAGE-001, AGENT-PARALLEL-OPT-001, COMPLETION-MARKER-RETIRE-001, LSPMCP-RETIRE-001, V3R3-RETIRED-AGENT-001), `in-progress` (CONFIG-DEAD-SWEEP-001), `implemented` (V3R2-ORC-001), `draft` (UPDATE-DOC-DRIFT-001). None retired/superseded/archived → no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` on all three artifacts → `0 0 0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md` → rc=1. `research.md` absent (correct for Tier M).

**No must-pass failure.** The FAIL is driven by rubric score, not by the firewall.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.60 | 0.50–0.75 | REQ-ACD-002 (`spec.md:L152`) is unsatisfiable on 123 measured lines and admits two opposite implementations (D1). Remaining 7 REQs are crisp; no `should`/`may` in normative text (grep → no match). |
| Completeness | 0.75 | 0.75 | All sections present; frontmatter complete; 5× `### Out of Scope — …` H3 with `-` bullets (`spec.md:L201,205,209,213,217`). Materially under-specified in three named places: extraction anchor (D6), debt-list path and snapshot path (D10), third over-count axis unmodelled (D2). |
| Testability | 0.60 | 0.50–0.75 | No weasel words (`grep -Ei 'appropriate\|adequate\|reasonable\|proper\|적절\|충분히' acceptance.md` → rc=1); mutation obligations are strong and restore-bounded. But 3 of 8 ACs are unexecutable or self-contradictory as written: AC-ACD-006 (D5), AC-ACD-007 (D3), AC-ACD-008 (D4). |
| Traceability | 0.75 | 0.75 | `spec.md` §4 matrix maps all 8 REQ → AC; all 8 AC headings cite an existing REQ; no orphan, no uncovered REQ. One indirect mapping: REQ-ACD-007's mirror obligation for the non-command clause body is unverified by any AC (D8). |

Harmonic mean = 4 / (1/0.60 + 1/0.75 + 1/0.60 + 1/0.75) = **0.667**. Arithmetic mean = 0.675. Reported: **0.67** — below the 0.80 Tier M threshold.

---

## Defects Found

**D1. Line-granularity collision makes REQ-ACD-002 unsatisfiable, and the prescribed fix silently under-counts** — `spec.md:L152` (REQ-ACD-002), `spec.md:L143-147` (§3.2 table), `spec.md:L134` (§3.1) — **Severity: critical — Class: blocking**

REQ-ACD-002 states both halves on one line: a non-live identifier "shall carry a reserved literal token **on every line on which that identifier occurs**", and "**no reserved token shall occur on a line bearing a live criterion's identifier**". On a line carrying both a live and a non-live identifier the two halves contradict; the line is unmarkable.

Not hypothetical — measured corpus-wide:

```
$ awk '... count lines with >=2 distinct AC prefixes ...' .moai/specs/*/acceptance.md
lines carrying >=2 distinct AC PREFIXES: 123
files containing such a line: 56
```

Concrete instance — `.moai/specs/SPEC-AGENT-PARALLEL-OPT-001/acceptance.md:248` carries the live `AC-APO-070` (its own row ID, occurring on that line and no other) together with the foreign, superseded `AC-DCP-010` cited from `SPEC-DWF-CODEMAPS-PILOT-001`.

Executed the §3.2 counter faithfully on a scratchpad copy (tree not mutated):

```
BEFORE normalization   sweep: 54   counter: 54   rc=0
AFTER  normalization   sweep: 54   counter: 52   rc=0        # line 248 tagged [REF]
```

Intent was 54 → 53 (drop one citation). Actual 54 → **52**, exit code **0**: the live `AC-APO-070` was swallowed because all of its occurrence lines now carry the token → state `retired` → silently excluded. **No halt, integer emitted, wrong number** — the exact failure mode `spec.md` §1 declares the SPEC exists to prevent ("근거를 못 세우면 통과가 아니라 정지여야 한다").

This answers audit point 1 negatively: the three states are **not** exhaustive over what a real `acceptance.md` contains, and an input **can** silently reach a number the SPEC intends to forbid. It is also distinct from the accepted residual (`plan.md` M0 "남는 것"), which covers *unnormalized* pure citations; this defect fires *after* the author performs the prescribed normalization.

**Required fix**: bind the token to the identifier rather than to the line — e.g. the token must be adjacent to (immediately following) the identifier it marks, and an identifier is `retired`/`ref` iff every occurrence carries an adjacent token. Restate §3.2's three states over *identifier occurrences*, not *lines*. Then add a fourth halt condition, or state explicitly why co-occurrence cannot arise.

---

**D2. A third over-count axis — same-file identifier aliasing — is unmodelled, and it dominates one of the SPEC's own named normalization targets** — `spec.md:L96` (§2.2 "두 축"), `plan.md:L135` (M5 target list) — **Severity: critical — Class: blocking**

`spec.md` §2.2 asserts over-counting has exactly **two** axes (retirement, citation). A third exists: one criterion referenced under two spellings in the same file.

```
$ grep -c '^## AC-ORC-001-' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md
17
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md | sort -u | wc -l
34
```

`AC-01`…`AC-17` are short-form aliases of `AC-ORC-001-01`…`-17`; the file's own traceability matrix uses the short form (`:423 | REQ-ORC-001-001 | AC-01 |`, `:445 | AC-01 | REQ-001 |`) and `:472` reads "All 17 ACs (AC-01 through AC-17)". True count **17**, sweep **34** — a 100% over-count.

The convention cannot express it: neither spelling is retired (`[RETIRED]` false) nor foreign (`[REF]` false — it belongs to this SPEC). The counter emits `34`, rc=0 — silent. And this file is one of only three files `plan.md` M5 selects for normalization and AC-ACD-007 binds as a Given.

**Required fix**: either model the alias axis (a third reserved token, or a rule that an identifier which is a proper suffix-form of another in the same file is not counted separately), or state it out of scope explicitly AND remove `SPEC-V3R2-ORC-001` from the M5 target list, replacing it with a pre-terminal file whose flag has been verified.

---

**D3. AC-ACD-007's Given names a false positive of the very class §8 catalogues, making the AC unsatisfiable** — `acceptance.md:L102-104`, `spec.md:L184` (§3.3), `spec.md:L228-236` (§8) — **Severity: major — Class: blocking**

AC-ACD-007 requires that for **all three** pre-terminal files the counter come out **strictly below** the sweep after normalization. The scan flags exactly one identifier in `SPEC-V3R2-ORC-001`:

```
$ grep 'SPEC-V3R2-ORC-001' .moai/reports/t338/overcount-scan.txt
OVERCOUNT .moai/specs/SPEC-V3R2-ORC-001/acceptance.md  AC-ORC-001-05
$ sed -n '122,128p' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md
## AC-ORC-001-05 — All 7 retired stubs exist with status=retired (REQ-006)
**Given** M2 has run for all 5 new retirements
**When** I list retired stubs
**Then** all 7 retired stubs exist in template tree:
```

A full Given/When/Then — a **live** criterion whose *subject* is retired stubs. Identical in kind to `AC-RA-02`/`AC-RA-07`/… in §8, which the SPEC correctly excluded. §2.1 states plainly that the 29 flags were **not** exhaustively compared ("플래그된 29개를 전수 대조하지는 않았으나"), and `measurement.md` § Gaps repeats it. The SPEC nonetheless promoted the unverified list into a binding Given. With nothing to normalize, counter == sweep and AC-ACD-007 fails.

This is a `verification-claim-integrity.md` §1.1 surface-3 violation: a defect claim (this file over-counts) asserted without verifying it with the domain's tool, then bound as an acceptance precondition.

**Required fix**: verify each of the three pre-terminal flags by hand before binding them; drop `SPEC-V3R2-ORC-001`; restate the arithmetic (15 flagged files → 14: 12 `completed` + 2 pre-terminal) and AC-ACD-007's Given accordingly. Related evidence that the scan is also a *lower* bound: `SPEC-UPDATE-DOC-DRIFT-001:778` reads "AC-UDD-002, AC-UDD-003, AC-UDD-006 are retired" but only `AC-UDD-006` was flagged.

---

**D4. AC-ACD-008 contradicts itself, and its 134 conflates a lower bound with an upper bound** — `acceptance.md:L110-116`, `spec.md:L196` (REQ-ACD-008) — **Severity: major — Class: blocking**

Item 2 hardcodes "부채 목록 파일이 **134건**을 담고"; item 3 requires the citation-axis count be "매 실행 재유도한다 — 목록을 손으로 단언하지 않고". Item 2 is exactly the hand-asserted count item 3 forbids.

Worse, the two summands are bounds in opposite directions. The retirement axis's 12 comes from `overcount-detector.sh`, which `spec.md` §2 (제약 2) declares a **lower** bound. The citation axis's 122 comes from `pre-terminal-scan.sh`, which `spec.md` §2.2 declares an **upper** bound ("156/602 는 **상한**이다") because a legitimately multi-family SPEC reads the same way — e.g. `SPEC-AGENCY-ABSORB-001/acceptance.md:546 | REQ-ROUTE-008 | AC-ROUTE-003, AC-SKILL-007 |`, two native families, not a citation. Recording those as "상시 부채" asserts a defect that was never verified.

**Required fix**: drop the literal 134 from item 2 and let item 3's re-derivation be the sole source; and either label the citation-axis 122 an unverified upper-bound candidate list (not "debt"), or verify it.

---

**D5. AC-ACD-006's corpus is both mis-scoped and off by one against this tree** — `acceptance.md:L94`, `spec.md:L179` (REQ-ACD-006), `spec.md:L30` — **Severity: major — Class: blocking**

```
$ ls -d .moai/specs/*/acceptance.md | wc -l          # depth-1 glob (what the scans used)
602
$ find .moai/specs -name acceptance.md | wc -l       # recursive (what "**" / "under" means)
603
$ find .moai/specs -mindepth 3 -name acceptance.md
.moai/specs/_archive/SPEC-SKILL-001/acceptance.md
```

AC-ACD-006 says "전수(현 트리 **601**건)" and REQ-ACD-006 says "across **every** `acceptance.md` **under** `.moai/specs/`" with the glob written `**`. Three different populations are in play: 601 (depth-1, excluding this SPEC — the figure the measurement produced), 602 (depth-1 as the tree stands now, this SPEC's file included), 603 (recursive, `_archive/` included). The snapshot baseline is defined by "fail on **any** difference", so a scope disagreement between the test author and a later reader silently changes what the test asserts.

**Required fix**: name the exact glob in REQ-ACD-006 (depth-1 `.moai/specs/*/acceptance.md`, `_archive/` excluded or included — decide), and drop the frozen count from the Given or make it re-derived at run time.

---

**D6. The extraction anchor — the load-bearing part of REQ-ACD-005 — is unspecified, while M1/M2 rewrite the extraction target in the same card** — `spec.md:L171` (REQ-ACD-005), `plan.md:L95,108,115` — **Severity: major — Class: blocking**

REQ-ACD-005 requires extracting the counter "verbatim from the B12 clause" and calls extraction load-bearing. Nothing defines what anchors it. Today the clause happens to contain exactly one fenced bash block (`.claude/agents/moai/manager-docs.md:81-97`), so "the fenced block after the `### B12` heading" would work — but M1 and M2 rewrite that clause to add the three-state table, the halt obligation and the resolution message, and M2 explicitly contemplates additional output examples. After the rewrite "the counter command" may no longer be uniquely identified, and the clause also carries inline-code commands (`grep -c '<SPEC-ID>' CHANGELOG.md`, `ls <path>`) that a loose anchor would capture.

Credit where due: a vacuous-extraction pass (both sides extract empty → byte-identity holds trivially — the very `0 == 0` hazard the B12 clause itself warns about) **is** killed by AC-ACD-005 item 2's one-character mirror mutant and by AC-ACD-003's hand-computed fixture counts. The residual defect is the anchor's definition, not vacuity.

**Required fix**: specify a stable machine-readable anchor in the clause itself (e.g. a sentinel comment line delimiting the extracted block) and add an AC asserting the extraction yields exactly one non-empty command.

---

**D7. The baseline snapshot can absorb a regression instead of catching it** — `spec.md:L179` (REQ-ACD-006), `acceptance.md:L96`, `plan.md:L138` — **Severity: major — Class: blocking**

AC-ACD-006's mutant proves the snapshot detects a change *at authoring time*. Nothing constrains regeneration afterwards, and `plan.md` M5 makes regeneration a normal step ("baseline 스냅샷(M3-c)은 M5 의 변경을 반영해 재생성한다"). A future run whose counts shift for a bad reason is cleared by re-running the generator — the golden file becomes a rubber stamp, and REQ-ACD-006's stated purpose ("계수기의 정확성을 주장하지 않고 매 실행 재유도한다") is defeated by the same mechanism that implements it. The SPEC does not acknowledge this anywhere.

**Required fix**: require the snapshot to carry per-file *state* (live/retired/ambiguous counts), not just a total, so a regeneration diff is reviewable; and require any regeneration commit to state which files changed and why. At minimum, record the limitation honestly.

---

**D8. REQ-ACD-007's mirror obligation for `manager-docs.md` is unverified by any AC** — `spec.md:L193` (REQ-ACD-007), `acceptance.md:L74-88` (AC-ACD-005) — **Severity: major — Class: blocking**

Measured today:

```
$ diff .claude/agents/moai/manager-docs.md internal/template/templates/.claude/agents/moai/manager-docs.md
(no output — byte-identical)
```

AC-ACD-005 asserts byte-identity of the **extracted counter command** (item 1) and the prompt-template pair's single-line divergence (item 3). Nothing asserts that the *rest* of the revised B12 clause — the three-state table, the halt obligation, the resolution-message wording — lands identically in the mirror. M4 edits both by hand; a drift in the prose half is undetected by every AC.

**Required fix**: add an AC item asserting `diff` on the `manager-docs.md` pair produces no output (the measured current state), alongside the existing 171-line assertion for the other pair.

---

**D9. `plan.md` M0's justification for the accepted cost is contradicted two paragraphs later** — `plan.md:L80` vs `plan.md:L82` — **Severity: minor — Class: blocking**

L80: "왜 이 비용이 견딜 만한가 … 실패가 **틀린 수**가 아니라 **멈춤**이기 때문이다." L82: for the base case of those same 34 files ("토큰이 하나도 없는 순수 인용, 현재 34건의 기본형") failure is **not** a halt — it stays a silent over-count. The stated reason the cost is bearable does not apply to the population it was invoked for.

**Required fix**: qualify L80 — the halt property covers partial marking only; the 34 files' base case is a silent over-count, and the cost is accepted on different grounds (per-file authorship context, deferred prefix mechanism).

---

**D10. The debt-list file and the baseline snapshot are counted in the Tier arithmetic but never given a path** — `plan.md:L154,156` (§F rows 7 and 11), `acceptance.md:L115` — **Severity: minor — Class: blocking**

AC-ACD-008 item 2 requires counting entries in "부채 목록 파일" — a file no artifact names. Same for "baseline 스냅샷 1건". Both are binding test inputs; without a path the AC is not binary-testable and the run-phase author picks locations unreviewed.

**Required fix**: name both paths in `plan.md` §F and in the ACs that read them.

---

**D11. AC-ACD-002 cites the wrong section for the control-group scan** — `acceptance.md:L42` — **Severity: minor — Class: optional**
"§3 스캔이 플래그하지 않은 대조군" — the scan and its results live in `spec.md` §2 (§2.1/§2.2); §3 is the discriminator decision. **Fix**: cite §2.

**D12. REQ-ACD-007 uses the GEARS `Where` pattern for an event trigger** — `spec.md:L193` — **Severity: minor — Class: optional**
GEARS reframes `Where` as capability gate / feature flag / static config. "Where the B12 clause … changes" is a trigger, i.e. `When`. Form-conformant, so MP-2 stands; pattern selection is wrong. **Fix**: `When the B12 clause or the prompt-template B12 bullet changes, …`.

**D13. `plan.md` §F omits `progress.md`** — `plan.md:L146-156` — **Severity: minor — Class: optional**
Run-phase writes `§E.2` / `§E.3` (and AC-ACD-002 / AC-ACD-007 both require recording values there), so the file is touched. 11-13 → 12-14; Tier M is unaffected. Verified that two plausible omissions are **correctly** excluded: `internal/template/catalog.yaml` already carries the `manager-docs` entry (`:122-124`), and `make build` generates no tracked file (`internal/template/embed.go:28` `//go:embed all:templates`). **Fix**: add the row.

**D14. Mirror neutrality is obliged for fixtures but not for the clause revision itself** — `spec.md:L222` (§7), `plan.md:L34` (§B-5) — **Severity: minor — Class: optional**

```
$ grep -coE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' internal/template/templates/.claude/agents/moai/manager-docs.md
0
```

The mirror carries zero SPEC IDs today. The revised B12 clause will illustrate `[RETIRED]`/`[REF]`; an example carrying a real SPEC ID would pollute the distributed template. The whole-tree CI class is narrow (`SPEC-(V3R[2-6]|AGENCY|WORKTREE)-`, `internal_content_leak_test.go:171`), so CI would not catch e.g. `SPEC-AGENT-EMIT-LINEAGE-001`. **Fix**: extend §7's neutrality constraint to the clause revision, not only fixtures.

---

## What the audit confirmed as sound

Recorded so the fix pass does not disturb these.

- **The accepted residual is honestly recorded** (audit point 2 — PASS). Present in all three claimed places and out-of-scope: `spec.md:L104-108` (§2.2 closing), `plan.md:L82` (M0 "남는 것"), `progress.md:L32` (§E.1), `spec.md:L217-222` (§6, the deferred prefix-level trigger + native-prefix declaration). Nothing overclaims that legacy files are covered; `acceptance.md:L132` (DoD) restates that the 34 are not normalization targets. Only caveat: D9.
- **The false-positive lesson did shape the requirement layer** (audit point 4 — PASS). REQ-ACD-004 forbids vocabulary matching and markup anchoring; §3.1 chooses a bracketed literal for exactly the stated reason; AC-ACD-004 binds the three known false-positive files as a regression test with real discriminating power. Caveat: vocabulary matching survives in the *inputs* — the normalization target list and the debt list are both derived from `overcount-detector.sh`'s `MARK` regex — and D3 is the direct consequence.
- **The 171-line neutralization claim is true, verified**: `diff` on the `manager-develop-prompt-template.md` pair emits exactly `171c171` (SPEC-SYNC-PARALLEL-DOCS-001 A9 → "the attributable diff-check pattern"). `spec.md:L223` and AC-ACD-005 item 3 are accurate.
- **The first test bed's numbers are true**: sweep on `SPEC-AGENT-EMIT-LINEAGE-001/acceptance.md` = 8 (`AC-AEL-001…008`); `AC-AEL-008` occurs on exactly one line (`:165`, the retirement note) carrying no live identifier, so live = 7 and the file is cleanly normalizable. AC-ACD-001's 8/7 holds.
- **The out-of-scope claim about the Go lint engine is true**: `internal/spec/parser.go` locates the AC section inside the passed document (`findACSectionStart`, `:64-70`) and never loads `acceptance.md`.
- **Tier M is correct**: 11-13 files sits in the SSOT's `5 - 15 files` band (`spec-workflow.md:141`); REQ 8 / AC 8 are within the Tier M ceiling of 16/16; the LOC-vs-file tension is resolved by the SSOT's own "LOC thresholds are guidance, not enforcement" (`:157`).
- Requirement-layer prose hygiene: no `should`/`may` in `spec.md`; no weasel words in `acceptance.md`.

---

## Verification integrity

**Claim** — the scores and every defect above.

**Evidence** — commands and verbatim output are inline per defect.

**Baseline-attribution** — all measurements were run in this run, in this worktree (`git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t338`), at HEAD `da03d9188`, with the tree unmodified (`git status --porcelain` before and after → `?? .moai/reports/t338/`, `?? .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/` only). The D1 counter experiment ran on a scratchpad **copy**; no tracked or untracked file in the worktree was mutated, so no restore was required.

**Gaps** — (a) the cross-model backends were not consulted (untracked SPEC ⇒ empty diff ⇒ `inconclusive` by construction); (b) the remaining 26 of the 29 detector flags were not individually adjudicated — D3 is one verified false positive, not a claim about the total; (c) the citation-axis 122 was not verified file-by-file, which is itself part of D4; (d) no claim is made about whether `spec.md`'s REQ sweep suffers the same defect (the SPEC scopes it out).

**Residual-risk** — the alias axis (D2) was established on one file; other spellings of the same phenomenon may exist and would widen D2's scope. The 123 collision lines (D1) were counted by AC-prefix distinctness, an approximation of live-vs-non-live co-occurrence; the decisive instance was confirmed by hand.

---

## Recommendation

**FAIL at 0.67 against the 0.80 Tier M threshold. Run-phase entry is blocked.**

Eight defects are blocking (D1–D8), two are minor-but-blocking (D9, D10), four are optional (D11–D14). The two that make run-phase entry unsafe rather than merely imperfect:

1. **D1 must be resolved before M1.** M1 is the SPEC's own most-irreversible milestone ("토큰 형태가 바뀌면 이미 정규화한 파일과 배포된 절이 모두 갈린다"). Fixing the line-vs-identifier granularity *after* the convention ships means re-normalizing every file and re-editing the distributed clause. The correction is small — bind the token to the identifier — and it belongs in M1's wording, not in a follow-up card.
2. **D2 + D3 must be resolved before M5.** One of the three normalization targets is a false positive whose real over-count the convention cannot express. Proceeding produces either a failed AC or a false `[REF]` marking that corrupts a live criterion.

Ordered fix route for `manager-spec`:

1. Restate `spec.md` §3.1/§3.2 and REQ-ACD-002 over **identifier occurrences** with an adjacency rule; re-check the three states are exhaustive under the new unit (D1).
2. Decide the alias axis — model it or scope it out explicitly — and remove `SPEC-V3R2-ORC-001` from the M5 target list either way (D2, D3). Re-verify the two remaining pre-terminal flags by hand and restate the 15/12/3 arithmetic.
3. Rewrite AC-ACD-008: delete the literal 134, keep re-derivation, and relabel the citation-axis 122 as an unverified candidate list (D4).
4. Fix AC-ACD-006's corpus definition — exact glob, archive in or out, no frozen count (D5).
5. Add the extraction anchor to the M1/M2 clause wording plus an AC that the extraction is unique and non-empty (D6).
6. Add snapshot-regeneration discipline, or record the rubber-stamp limitation honestly (D7).
7. Add the `manager-docs.md` pair `diff` assertion to AC-ACD-005 (D8).
8. Qualify `plan.md:L80` (D9); name the debt-list and snapshot paths (D10); sweep D11–D14 at the author's discretion.

Iteration 2 will be scoped to this enumerated delta plus a regression check over D1–D10, per the Tier M ceiling of 2.
