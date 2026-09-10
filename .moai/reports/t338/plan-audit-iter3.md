# SPEC Review Report: SPEC-AC-COUNT-DISCRIMINATOR-001

Iteration: 3/3 (Tier L ceiling — `harness.yaml:78` `L: 3`)
Verdict: **PASS** (with one recorded blocking-class debt item, P1)
Overall Score: **0.91** (Tier L threshold 0.85)

Measurement tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t338`, branch `WT-ac-count-sweep`, HEAD `da03d9188`.

**Reasoning context ignored per M1 Context Isolation.** The dispatch's account of what changed since iter-2 was treated as a set of claims to verify against the tree, not as findings to accept. Every claim below was re-established by running the command shown. The audit read all six artifacts per the Tier L input contract (`spec.md` + `plan.md` + `acceptance.md` + `design.md` + `research.md` + `progress.md`).

**Independence note.** The N2 reproduction did not reuse the repair's fixtures. `repair-scratch/{inside,outside,eol}.md` are the repair's inputs; using them with an audit-side counter makes only the implementation independent, not the input. iter-3 therefore rebuilt the three fixtures from the original file and ran the audit-side counter (`iter2-scratch/counter.py`) over them. Assets: `.moai/reports/t338/iter3-scratch/`.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 8 requirements, sequential, no gap, no duplicate, uniform zero-padding.
  ```
  $ grep -oE '^\*\*REQ-ACD-[0-9]{3}\*\*' spec.md | sort | uniq -c
     1 **REQ-ACD-001**  …  1 **REQ-ACD-008**      (each exactly 1)
  ```

- **[PASS] MP-2 GEARS format compliance** — **judged against the requirement layer only** (`spec.md` §4 `REQ-XXX` entries). The `AC-XXX` Given-When-Then entries in `acceptance.md` are the verification layer and were graded under Group 4, never here. All eight match a GEARS pattern: REQ-001/002/005/006/008 ubiquitous (`The <subject> shall …`), REQ-003/007 event-driven (`When …, the … shall …`), REQ-004 unwanted (`shall not decide … shall not anchor`). No informal language, no Given-When-Then presented as a REQ.

- **[PASS] MP-3 YAML frontmatter validity** — `spec.md:1-16` carries all 12 canonical fields with correct types: `id`, `title`, `version: "0.4.0"` (quoted semver), `status: draft`, `created`/`updated` `2026-08-28` (ISO), `author`, `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) present. Optional `tier: L` (`spec.md:14`) and `era: V3R6` additionally present.

- **[N/A] MP-4 Section 22 language neutrality** — the SPEC is scoped to this repository's own Go tooling and to language-neutral markdown clause text. It names no per-language tool and enumerates no partial language set. The template-bound surface was verified neutral: `grep -coE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' internal/template/templates/.claude/agents/moai/manager-docs.md` → `0`, and `spec.md` §7 requires the revised clause to keep that property using synthetic ids (`AC-SYN-00N`). N/A auto-passes.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — verb executed. 11 external SPEC ids referenced; all 11 resolve to an existing `spec.md`; **none** carries a status in {retired, superseded, archived}:
  ```
  SPEC-AGENT-EMIT-LINEAGE-001        completed     SPEC-LSPMCP-RETIRE-001        completed
  SPEC-AGENT-PARALLEL-OPT-001        completed     SPEC-STATUSLINE-001           implemented
  SPEC-COMPLETION-MARKER-RETIRE-001  completed     SPEC-UPDATE-DOC-DRIFT-001     draft
  SPEC-CONFIG-DEAD-SWEEP-001         in-progress   SPEC-V3R2-ORC-001             implemented
  SPEC-HUMANIZE-001                  completed     SPEC-V3R3-RETIRED-AGENT-001   completed
  SPEC-SKILL-001                     completed (_archive/ — referenced only as the recursive-glob difference, reconciled by the depth-1 glob pin in REQ-ACD-006)
  ```
  No BLOCKING finding. Collateral confirmation: `§3.3`'s claim that the two retirement-axis normalization targets are pre-terminal holds (`in-progress`, `draft`), and `§2.3`'s alias-axis target `SPEC-V3R2-ORC-001` is `implemented`, which is likewise outside the terminal set AC-ACD-007 names.

- **[PASS] MP-6 D8 cross-platform discipline** — auto-PASS per D8-4. `grep -c 'syscall'` returns `0` on every one of the six artifacts.

- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/` → `rc=1`, no match, across all six artifacts. `progress.md` §E.1 independently records `미결 결정: 없음`.

All seven must-pass criteria clear. No must-pass-equivalent BLOCKING finding from D7 or D8.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75–1.0 | The two iter-2 ambiguities in the load-bearing `marked` predicate are closed to the character class and independently reproduced: same-line only (`REQ-ACD-002`, `spec.md:173-178`, `design.md` §B.2) and any non-space/tab character breaks adjacency including a closing backtick (measured 53 vs 54, below). §3.2's exhaustive/exclusive argument holds over the identifier unit. Not 1.0: §3.4 states its rule for "예시 문서" universally while the SPEC honors it for `acceptance.md` only (P1), and `REQ-ACD-002`/`REQ-ACD-006` are single ~70/~90-word sentences that need re-reading — unambiguous, but strained. |
| Completeness | 0.90 | 0.75–1.0 | All required sections present; 12/12 frontmatter fields; five `### Out of Scope — <topic>` H3 sub-headings each with specific `-` bullets (`spec.md:367,371,375,379,383`). Tier L artifact set complete and genuine — `research.md` carries measurement plus an explicit five-item Gaps section, `design.md` carries at least three artifacts found nowhere else (§C.1 snapshot line schema, §D carrier topology, §E consolidated six-row rejection table). Both iter-2 material gaps closed: §F row 14 (third carrier) and §3.5 + AC-ACD-006 item 5 (corpus halt handling). Not 1.0: `design.md` restates roughly 40% of `spec.md` §3.2/§3.5 despite its own disclaimer, and `progress.md`'s `module:` field did not migrate with the tier (P2). |
| Testability | 0.90 | 0.75–1.0 | Zero weasel words (`grep -niE '적절\|합리적\|충분히\|적당\|reasonable\|appropriate\|adequate\|proper' acceptance.md` → `rc=1`). Every AC names a command and an expected value; four carry mutant obligations. All four iter-2 testability defects verified closed — AC-ACD-001 is now satisfiable by a single implementation (independently reproduced below), its Given pins placement with distinct synthetic ids, AC-ACD-006 item 5 is determinate on halting files, AC-ACD-007 item 5 removes the vacuous pass. Not 1.0: AC-ACD-006 item 5's enumeration (b/c/d) does not name the HALT→HALT-with-different-identifiers transition, which is caught only by REQ-ACD-006's general "fail on any difference" (P4); the three-value reading table is exclusive but not exhaustive. |
| Traceability | 1.00 | 1.0 | 8 REQ / 8 AC, verified both directions. `spec.md` §4 matrix maps every REQ to ≥1 AC and every one of AC-ACD-001…008 appears at least once; each AC heading's parenthetical REQ list matches the matrix row-for-row (checked all eight). No orphan AC, no uncovered REQ, no reference to a non-existent identifier. `grep -coE '^### AC-ACD-[0-9]{3}' acceptance.md` → `8`. The iter-2 traceability gap (REQ-ACD-007 reaching a surface no AC asserted) is closed by AC-ACD-005 item 6(a)(b)(c). |

Aggregate: harmonic mean of the four dimensions = **0.9094 → 0.91** (per `agent-common-protocol.md` § Skeptical Evaluation Stance). Arithmetic mean is 0.9125 — the two agree, so the verdict does not turn on the choice of aggregation. **0.91 ≥ 0.85 (Tier L).**

---

## Verification of the dispatch's claims

### Tier migration — the SSOTs, verified

`spec-workflow.md` § SPEC Complexity Tier (read at `:141-142` and the two tables following):

```
| L (Large) | > 1000 LOC or constitutional | > 15 files | 5 files: spec+plan+acceptance+design+research | 0.85 |
| L | 25 | 25 |     (REQ ceiling | AC ceiling)
```
`harness.yaml:78` → `L: 3   # Tier L … up to three spawns`.

Every one of the four values the dispatch asked about moved together, and no stale current-tier statement survives:

```
$ grep -rn 'Tier M|0\.80|tier: M' .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/
progress.md:49   … Tier M → L                    (historical, iter-2 repair record)
design.md:18     … Tier M 상한 15 를 넘어 Tier L   (historical rationale)
plan.md:225,228  | 축 | Tier M (종전) | Tier L (현재) | / | 0.80 | 0.85 |   (explicit before/after table)
spec.md:24       … Tier M 상한 15 초과 → Tier L    (HISTORY row)
```
All five are historical or comparative. `tier: L` appears exactly once, in `spec.md:14`, which is the correct carrier. Threshold 0.85, ceiling 3, and the 25/25 REQ-AC ceiling are stated consistently in `progress.md` §E.1, `plan.md` §F, and `acceptance.md` Definition of Done. **The migration is complete on the tier/threshold/ceiling axis.** One field did not move — see P2.

### §F honesty at 16 rows / 18 files

The arithmetic is reproducible from the table: rows 1-5 → 5, row 6 (fixtures, 3 files) → 8, row 7 → 9, rows 8-10 → 12, row 11 → 13, row 12 → 14, row 13 → 15, **row 14 → 16**; rows 15-16 (`design.md`, `research.md`) bring the file total to 18. The 16-vs-18 distinction is stated rather than elided. Every row that names an existing path resolves:

```
$ ls -la  (all present)
.claude/agents/moai/manager-docs.md                                              11945
internal/template/templates/.claude/agents/moai/manager-docs.md                  11945
.claude/rules/moai/development/manager-develop-prompt-template.md                16144
internal/template/templates/.claude/rules/moai/development/…prompt-template.md   16149
internal/template/templates/.codex/agents/moai/manager-docs.toml                 11579
```
Row 3's target is real and matches plan.md M2's quotation of it verbatim — `manager-develop-prompt-template.md:131`: `- Verify AC count in CHANGELOG matches \`acceptance.md\` (SSOT) …`. Row 12's mechanism is real: `catalog.yaml:125` carries `hash: 27d6252a…3b7954`, and `shasum -a 256` of the mirror returns the identical digest, so a mirror edit does move that tracked file. Rows 5, 6, 7, 11 are files the run-phase creates. Rows 8-10 are contingent — see P3.

### The third carrier is genuinely a separate file, and the gate is where the SPEC says

```
$ grep -c 'codex' internal/template/catalog.yaml        → 0
$ grep -c '\.codex' internal/template/catalog.yaml      → 0
$ sed -n '23p' Makefile
build: agents-emit-check templ-generate ## Build the binary
$ sed -n '31,38p' Makefile
agents-emit-check: ## Verify the committed .codex TOMLs match the .md source layer (read-only; never regenerates)
	@AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run TestGoldenCommitted… -count=1 \
		|| { printf 'agent-emit drift: … run `make agents-emit`\n' >&2; exit 1; }
```
The catalog hash path and the emission path are disjoint, so row 14 is not double-counted, and the read-only gate genuinely precedes compilation — M4's ordering claim (mirror edit → `make agents-emit` → `make build`) is grounded. The TOML does carry the counter command in full (`manager-docs.toml:68-74`, including the `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' …` fenced block), so REQ-ACD-007's "every distributed carrier" is not rhetorical.

### `design.md` and `research.md` are Tier L artifacts, not padding

`research.md` is a measurement document: §A direction table (why the three scans cannot be summed), §B the three simultaneously-true populations, §C the per-axis measurements with console output, §D the four iter-2 measurements, §E the mirror baseline with sha256, §F five explicit non-measurements, §G a reproduction-asset table. Its §F Gaps section names what was *not* observed — including "`make build` 를 끝까지 돌리지 않았다" — which is the section most documents omit. Real artifact.

`design.md` is thinner. Its §B.2 state function and §C.2 transition table restate `spec.md` §3.2/§3.5, and §E's first two rows restate §3's parser rejection. But it carries three things found nowhere else: the snapshot line schema (`<path> COUNT <n> live=… excluded=… ambiguous=0` / `<path> HALT <id>… owner=… reason=…`), the carrier topology diagram distinguishing the hash path from the emission path, and the consolidated six-row rejection table. Not padding, but the weakest of the six artifacts.

### Route B and the repo-local override — the citation is real and correctly used

`.claude/rules/local/repo-local-pr-policy.md` exists and states, verbatim: "Completed cards integrate into LOCAL `develop` via `git merge --no-ff` … There are NO card-level PRs" and "PR-based ceremony (spec-workflow Route B tier routing) applies ONLY to the release path above." `progress.md` §E.1 and `acceptance.md`'s Definition of Done both cite it correctly ("리포 고유 규정이 우선한다 — 카드는 `develop` 로 통합하고 카드 PR 을 내지 않는다"). The override is real, current (the file's prior all-tier-Route-B policy is explicitly marked `[RETIRED 2026-08-27]`), and correctly cited. `plan.md` §F's table omits the caveat — see P5.

### N3 half 1 — the halt/COUNT distinction is coherent

The separation the repair argues (B12 asks whether *this SPEC* produces the right count, where a halt is not a pass; the corpus verifier asks whether *the counter* classifies stably tree-wide, where a halt is a legitimate observation and the judgeable thing is whether it moved) holds under probing. I looked for a point at which a halt reads as a pass and found none:

- In B12 the halt emits no integer and a non-zero exit code (`design.md` §B.1) — unchanged from the SPEC's founding premise.
- In the corpus a halt is a first-class `HALT` entry; `COUNT→HALT` fails, `HALT→COUNT` fails, an unrecorded halt always fails, and skipping or zero-scoring is prohibited by name (`spec.md` §3.5 rules 1-3, AC-ACD-006 item 5(a)-(d), `design.md` §C.2).
- The regeneration escape (a bad move absorbed by one snapshot regeneration) is real and is disclosed three separate times as residual risk, each time explicitly as visibility rather than prevention. That is the correct handling of a risk the design does not eliminate, not a gap.

One drafting seam survives (P4) and one operational consequence is unnamed (P4's second half), neither of them a loophole through which a halt becomes a pass.

### N3 half 2 — the relocation is a hiding place, and it is broken today

This is the finding of the iteration. Verified in three steps.

**Step 1 — `acceptance.md` is genuinely clean.** Two independent implementations agree, and the whole corpus is halt-free:

```
$ python3 .moai/reports/t338/iter2-scratch/counter.py …/acceptance.md adj      # audit-side
COUNT 24   (live=24 excluded=3 ambiguous=0)     rc=0
$ python3 .moai/reports/t338/repair-scratch/counter.py …/acceptance.md sameline # repair-side
COUNT 24 (live=24 excluded=3 ambiguous=0)       rc=0

$ bash .moai/reports/t338/iter2-scratch/corpus.sh | tail -2
EXCL .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/acceptance.md :: COUNT 24 …
files=602  halting=0  files-with-excluded=1
```
602 files, `halting=0`. The single file that halted at iter-2 no longer does, and no other file took its place. §3.4's fix to the AC-ACD-004 table works.

**Step 2 — the destination halts.** Running the same counter over the sibling artifacts:

```
$ python3 …/iter2-scratch/counter.py …/research.md adj   →  AMBIGUOUS AC-DCP-010 AC-SYN-010   rc=3
$ python3 …/iter2-scratch/counter.py …/spec.md     adj   →  AMBIGUOUS AC-DCP-010 AC-SYN-010   rc=3
$ python3 …/iter2-scratch/counter.py …/plan.md     adj   →  COUNT 11   rc=0
$ python3 …/iter2-scratch/counter.py …/design.md   adj   →  COUNT 2    rc=0
```

**Step 3 — the two causes are exactly the two the repair reported closing.**

```
$ grep -n 'AC-SYN-010' …/research.md
141:AMBIGUOUS AC-SYN-010 AC-SYN-012 (live=22 excluded=0)          ← unmarked (the relocated diagnostic)
144:… fixture 열에서 `AC-SYN-010 [REF]` 를 표시된 채로 보이고 …      ← marked
$ grep -n 'AC-DCP-010' …/research.md
109:$ python3 counter.py inside.md  sameline     # `AC-DCP-010 [REF]`     ← marked
111:$ python3 counter.py outside.md sameline     # `AC-DCP-010` [REF]     ← unmarked (backtick intervenes)
```
The `AC-SYN-010` split is the quoting recursion, reappearing verbatim at the file the diagnostic was moved into. The `AC-DCP-010` split is the situation `spec.md` §3.4 rule 2 governs — one identifier shown in both a marked and an unmarked shape — which the repair applied in `acceptance.md` (using `AC-SYN-020` / `AC-SYN-021`) but not in `spec.md:178` or `research.md:109-111`, where `AC-DCP-010` carries both shapes.

Judgment on the question asked: **relocation is a hiding place, not a fix — and the break has already returned, it is not hypothetical.** `research.md` does not need to "ever enter a corpus" for the claim to be false; the claim is about the document, and the document fails it now. Full finding at P1.

### N2 — AC-ACD-001 is satisfiable, and the three readings are mutually exclusive

Reproduced from fixtures I rebuilt myself (`.moai/reports/t338/iter3-scratch/`), so the input as well as the counter is independent of the repair:

```
$ sed -n '248p' orig.md
| AC-APO-070 | 070 | MUST | … (b) `spec.md`가 superseded AC를 ID로 인용 — `AC-DCP-010`(`acceptance.md:79` …

$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' orig.md | sort -u | wc -l          →  54
$ counter.py inside.md  adj    COUNT 53   (live=53 excluded=1 ambiguous=0)
$ counter.py outside.md adj    COUNT 54   (live=54 excluded=0 ambiguous=0)
$ counter.py eol.md     adj    COUNT 54   (live=54 excluded=0 ambiguous=0)
$ counter.py eol.md     line   COUNT 52   (live=52 excluded=2 ambiguous=0)
$ counter.py inside.md  line   COUNT 52   (live=52 excluded=2 ambiguous=0)
```

- **Satisfiable.** One implementation — same-line, whitespace-only adjacency on the identifier unit — simultaneously yields item 1's `54` on the original, item 2's `53` on the copy, satisfies item 3 by not yielding `52`, and yields `54` on the end-of-line mutant so the mutant dies. The iter-2 impossibility (`52` demanded of both the mutant and the regression signature) is genuinely closed, and the corrected value `54` appears in both places the dispatch named — `acceptance.md` AC-ACD-001 and `plan.md:134` onward.
- **Mutually exclusive.** On `eol.md` the adjacency semantics yields 54 and the line semantics yields 52; 53 would require a counter that ignores token placement entirely. Three distinct integers from three distinct semantics — no input produces two of them.
- **Bonus discrimination, unclaimed by the SPEC.** The line implementation also returns `52` on `inside.md`, the AC's *primary* input, not only on the mutant. So item 3's D1 signature fires on the main path too, which makes the AC stronger than it claims to be.

### N4/N5 — no AC rests on an example contradicting the pin

Checked every fixture row of AC-ACD-004 against the pinned predicate. `AC-SYN-010` marked in both the fixture and expectation columns (excluded); `AC-SYN-011` unmarked in both (live); `AC-SYN-012` marked in both (excluded); `AC-SYN-013` token-before (live); `AC-SYN-014` token outside the span (live). AC-ACD-001's comparison table uses `AC-SYN-020` (inside, → 53) and `AC-SYN-021` (outside, → 54) as distinct identifiers, which is §3.4 rule 2 applied correctly. Mechanical confirmation: the file's own count is `ambiguous=0` with `excluded=3`, and the three excluded are exactly `AC-SYN-010`, `AC-SYN-012`, `AC-SYN-020` — the set the tables intend. **No AC depends on a contradicting example.** The contradiction lives only in `spec.md`/`research.md`, where no AC reads it (P1).

---

## Regression Check (iter-2 defects N1-N11)

| # | Status | Evidence |
|---|---|---|
| N1 third carrier / Tier ceiling | **RESOLVED** | §F row 14; Tier L propagated to five surfaces; M4 orders `make agents-emit` before `make build`; REQ-ACD-007 broadened to "every distributed carrier … each machine-generated artifact"; AC-ACD-005 item 6(a)(b)(c). Carrier verified present and command-bearing; `grep -c '\.codex' catalog.yaml` → 0 confirms it is a separate file. |
| N2 inverted mutant | **RESOLVED** | `54` in `acceptance.md` AC-ACD-001 and `plan.md:134`; independently reproduced 54/53/52; satisfiability established above. |
| N3 own file halts | **RESOLVED for `acceptance.md`**; the corpus-handling half is fully specified (§3.5, REQ-ACD-006, AC-ACD-006 item 5, `design.md` §C). Corpus `halting=0` over 602 files. The relocation consequence is a **new** finding (P1), not an N3 regression. |
| N4 code-span placement | **RESOLVED** | `spec.md` §3.1 clause 2; REQ-ACD-002; AC-ACD-001 Given table; AC-ACD-004 backtick row. Branch reproduced (53 vs 54). |
| N5 cross-line adjacency | **RESOLVED** | `spec.md` §3.1 clause 1; REQ-ACD-002 "on the same line … a newline … breaks the adjacency"; `design.md` §B.2; AC-ACD-004 newline row. |
| N6 vacuous pass | **RESOLVED** | AC-ACD-007 item 5(a)/(b) — 1+ targets, or a recorded per-identifier account showing 0 is a measurement. |
| N7 frozen 123/56 | **RESOLVED** | Re-derived this run: `lines carrying >=2 distinct AC prefixes: 125 / files: 56` — matching the SPEC's stated 125. No literal frozen in the body; the `123` survives only inside the 0.3.0 HISTORY row, correctly scoped to that version. Direction label ("neither a clean upper nor lower bound") present in §3.1 and `research.md` §D.4. |
| N8 13-15 range | **RESOLVED** | Single value plus the eight-row arithmetic table; `progress.md` §E.1 carries the same arithmetic. |
| N9 `testdata/account/` | **RESOLVED** | §F row 6 now `internal/spec/testdata/ac_count/`, aligned with the verifier filename. |
| N10 retirement-section re-derivation | **RESOLVED** | AC-ACD-008 item 4 exempts that section by name and designates `progress.md` §E.2's per-identifier hand-adjudication record as its source. |
| N11 exclusion scope overclaim | **RESOLVED** | AC-ACD-002's note now states the limit and cites it. Re-derived: `grep -c 'acceptance.md' pre-terminal-scan.txt` → `34` of `39` total lines — exactly the claim. `grep -c '^=== ' alias-shape-scan.txt` → `85`, also exact. |

**iter-1 D1-D14: no regression.** Spot-checked the load-bearing ones — D1's identifier-unit binding is intact and now measured from two directions; D4's forbidden total is absent (`grep` for a summed figure finds none; REQ-ACD-008 and AC-ACD-008 item 3 both forbid it); D5's glob is pinned in REQ-ACD-006 with `_archive/` excluded; D6's sentinel anchor plus the exactly-one-non-empty assertion is in REQ-ACD-005 and AC-ACD-005 item 1; D7's regeneration risk is disclosed in three places; D8's mirror-pair `diff` assertion is AC-ACD-005 item 3.

**Named regression items from the dispatch, each verified:**

```
$ diff …/templates/.claude/agents/moai/manager-docs.md .claude/agents/moai/manager-docs.md
(no output)                                                              rc=0   ← pair byte-identical
$ diff .claude/rules/…/manager-develop-prompt-template.md …/templates/…/manager-develop-prompt-template.md
171c171   < …(SPEC-SYNC-PARALLEL-DOCS-001 A9).  > …(the attributable diff-check pattern).
(exactly one hunk, line 171, the neutralization)                         rc=1
$ AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.667s      ← golden gate green
$ git status --short
?? .moai/reports/t338/
?? .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/                          ← no tracked file modified
```
The three B12 carriers are untouched, the iter-2 audit's own mirror mutant was restored (the pair is byte-identical again), and the golden gate is green. The 34-file citation-axis residual is still recorded honestly as an accepted cost in `plan.md` §E M0, `spec.md` §2.2, `progress.md` §E.1, and `acceptance.md`'s Definition of Done, with the `completed` populations labelled **upper bounds** and never asserted as debt. No natural-language vocabulary matching has been reintroduced — REQ-ACD-004 forbids it and `design.md` §E records it as a rejected alternative.

### The HISTORY gap the iter-2 stall left

Checked each repair the 0.4.0 HISTORY row claims against the body, rather than taking the row's word for it. **Every claimed repair exists**: §3.4 (`spec.md:268` onward) and §3.5 (`spec.md:281` onward) are both present and substantive; the adjacency pin is in §3.1 clauses 1-2 and in REQ-ACD-002's normative text; the `52 → 54` correction is in `acceptance.md` AC-ACD-001 and `plan.md` M3; the N6 guard is AC-ACD-007 item 5; N7's re-derivation replaced the frozen literal; N8-N11 are each traceable to a specific edit. The `요구 8 / 수락 8 불변` claim is mechanically true (8 and 8, verified). **The overstating-HISTORY defect is closed.** One pointer inside the row is wrong — see P6.

---

## Defects Found (structured defect-list)

**P1. `spec.md` and `research.md` halt under the SPEC's own counter, and `acceptance.md`'s stated reason for relocating the diagnostic is false at its destination** — `acceptance.md:172-174` (the AC-ACD-006 note), `spec.md:268-280` (§3.4), `research.md:141`+`:144`, `research.md:109`+`:111`, `spec.md:173`+`:178` — **Severity: major — Class: blocking**

`acceptance.md`'s AC-ACD-006 note justifies moving the halt diagnostic into `research.md` with a single sentence whose two halves contradict each other: *"corpus 는 `acceptance.md` 만 판정하므로 축차 출력은 `research.md` 가 담는 것이 옳은 자리다 — 규약을 예시하는 문서는 **자기가 인용하는 진단까지** 규약 아래 있다."* The first half scopes by what the corpus reads; the second half asserts a universal property of illustrating documents. Measurement shows the second half is false at the named destination — `research.md` halts (`AMBIGUOUS AC-DCP-010 AC-SYN-010`, rc=3), for the same two reasons the repair reported closing: the quoted diagnostic reintroduces `AC-SYN-010` unmarked beside a marked occurrence, and `AC-DCP-010` is shown in both a marked and an unmarked shape without the distinct synthetic identifiers §3.4 rule 2 requires. `spec.md` halts identically. `spec.md` §3.4's own reasoning forecloses the move: it rejects excluding the illustrating document from the corpus on the ground that doing so would make the SPEC assert "규약은 모두에게 적용되지만 규약을 적는 문서에는 적용되지 않는다" — and relocating the offending text into a document the corpus never reads is that same move by another route. This is a truth defect in the artifact (`verification-claim-integrity.md` §1.1 surface 1), not an implementation blocker: `AC-ACD-006`'s corpus is `acceptance.md`-scoped and green at `halting=0` over 602 files, so no AC is unsatisfiable and no run-phase step fails because of it.
**Required fix (choose one, both one edit):** (a) narrow the claim — replace the universal second half with an explicit statement that the convention is enforced on the corpus surface (`acceptance.md`) and that `spec.md`/`research.md` carry the diagnostic and the two-placement comparison outside that surface by design, recording the fact that they would halt; **or** (b) apply §3.4 rule 2 to the two sites — give `spec.md:178` and `research.md:109-111` distinct synthetic identifiers for the inside-span and outside-span shapes (as `acceptance.md` already does with `AC-SYN-020`/`AC-SYN-021`), and mark or re-render the quoted `AMBIGUOUS …` line at `research.md:141` and `spec.md:263` so its identifiers are uniform. (b) makes the universal claim true; (a) makes the SPEC honest about a scope it already relies on.

**P2. `progress.md`'s `module:` field did not migrate with the tier** — `progress.md:11` — **Severity: minor — Class: blocking**

Five of six artifacts list `internal/template/templates/.codex` in `module:`; `progress.md` alone omits it:
```
$ grep -n '^module:' .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/*.md
spec/plan/acceptance/design/research  … internal/template/templates/.claude, internal/template/templates/.codex, internal/spec
progress.md:11                        … internal/template/templates/.claude, internal/spec
```
This is the one place the Tier M→L migration is half-done, and it is the file whose §E.1 carries the Tier-L arithmetic naming the `.codex` TOML as the file that crossed the ceiling. **Required fix:** add `internal/template/templates/.codex` to `progress.md:11`, matching the other five artifacts verbatim.

**P3. The Tier L determination rests in part on §F rows the SPEC itself declares non-binding** — `plan.md` §F rows 8-10 and the arithmetic table, against `acceptance.md` AC-ACD-007 item 5(b) — **Severity: minor — Class: optional**

§F counts three pre-terminal `acceptance.md` normalizations toward the 16-file total, while AC-ACD-007's Given requires the target set to be re-derived at run time and item 5(b) explicitly contemplates a result of zero. Remove those three and §F's own arithmetic yields 13 + 2 Tier L artifacts = 15, which is not `> 15` — Tier M. The tier therefore turns on an expectation the SPEC correctly refuses to freeze. This is not a defect to fix by changing the tier: L is the conservative direction (stricter threshold, more artifacts, higher audit ceiling), and a plan-phase tier is an estimate. **Required fix (optional):** add one sentence to §F noting that the count includes the expected normalization targets, that a re-derived count of zero would place the SPEC at the M/L boundary, and that L is retained deliberately as the safe side.

**P4. AC-ACD-006 item 5's transition enumeration omits `HALT → HALT` with a changed identifier set, and no artifact says who regenerates the snapshot when an unrelated SPEC's sync normalizes its file** — `acceptance.md:167-171`, `spec.md` §3.5 rules 1-3, `plan.md` M5 — **Severity: minor — Class: optional**

Item 5(b)(c)(d) name `COUNT→HALT`, `HALT→COUNT`, and an unrecorded halt. A file recorded `HALT AC-Y` that begins halting on `AC-Z` instead changes neither its recorded state nor its presence in the snapshot, so it falls through the enumeration; it is caught only by REQ-ACD-006's more general "fail on any difference" and AC-ACD-006 item 2. An implementer reading item 5 as the specification of halt handling could implement three rules and miss the fourth. Separately, §3.5 rule 1 makes `HALT→COUNT` a failure by design — which means every legitimate downstream normalization, performed by another card's author at their own sync, breaks this corpus test until someone regenerates the snapshot. `plan.md` M5 makes regeneration normal procedure but assigns it only within this card. **Required fix (optional):** add a fifth sub-item to AC-ACD-006 item 5 making a change in a `HALT` entry's identifier set an explicit failure, and one line to M5 or §3.5 naming who owns snapshot regeneration when an unrelated SPEC's normalization moves a file.

**P5. `plan.md` §F's Tier table states the git path without the repo-local override the other two artifacts carry, and its Tier M cell describes a route this repository prohibits** — `plan.md:229` — **Severity: minor — Class: optional**

The table reads `| git 경로 | Route A(main 직행) | **Route B**(phase 별 PR — manager-git) |` with no caveat, while `progress.md` §E.1 and `acceptance.md`'s Definition of Done both correctly subordinate Route B to `.claude/rules/local/repo-local-pr-policy.md`. That file states card work integrates into local `develop` with no card-level PR, and that direct push to `main` is disabled with `enforce_admins: true` — so the "Route A(main 직행)" cell describes something unavailable here even hypothetically. **Required fix (optional):** append the same one-clause caveat used in `progress.md` §E.1 to the `git 경로` row.

**P6. The 0.4.0 HISTORY row cites §3.1 for the N2 measurement, which §3.1 does not contain** — `spec.md:24` — **Severity: minor — Class: optional**

The row reads "N2 … `52` → `54` 로 정정 … (실측 표 §3.1)". §3.1's console block measures the inside-span vs outside-span branch (53 vs 54) — the N4 measurement. The end-of-line mutant measurement that establishes N2 (`counter.py eol.md sameline` → 54, `… line` → 52) lives in `research.md` §D.1, `plan.md` M3, and `acceptance.md` AC-ACD-001, not in §3.1. The repair described is real and present; only the pointer is wrong. **Required fix (optional):** change the citation to `research.md` §D.1.

---

## Recommendation

**PASS at 0.91 against the Tier L threshold of 0.85.** Monotonicity: **0.67 → 0.72 → 0.91**, strictly increasing across all three iterations and improving in every dimension (Clarity 0.75→0.85, Completeness 0.70→0.90, Testability 0.65→0.90, Traceability 0.80→1.00). No score regression, so no STOP escalation and no scope-reduction proposal is triggered. All seven must-pass criteria clear, with no D7 or D8 BLOCKING finding.

The iteration-2 verdict turned on four things, and all four are closed by measurement rather than by assertion: the third distributed carrier is enumerated and its build gate reproduced; the inverted mutant is corrected and the AC is now satisfiable by a single implementation I reproduced from independently rebuilt fixtures; the SPEC's own `acceptance.md` no longer halts and the whole 602-file corpus is clean; and the adjacency predicate is pinned to the character class in the requirement layer, the design layer, and the fixture table alike. The Tier migration is complete on every axis the dispatch named — artifact set, threshold, audit ceiling, REQ/AC ceiling — and the repo-local PR override is real and correctly cited.

**Does any surviving defect block implementation?** No. Stated plainly: **P1 is debt the run-phase can carry, not a blocker.** It is classified blocking under M6 because it is a consistency defect rather than a preference — one sentence in `acceptance.md` asserts something the tree contradicts — but nothing downstream depends on it. `AC-ACD-006`'s corpus is `acceptance.md`-scoped and measured green; no acceptance criterion is unsatisfiable; no milestone stalls. The correct handling is a single pre-kickoff edit, not another audit round: take option (a) or (b) under P1, and fold in P2 (one field, one line) at the same time. Both are text edits with no measurement dependency, so they need no re-audit — verify them by re-running `iter2-scratch/corpus.sh` and confirming `halting=0` is unchanged, plus `grep -n '^module:'` across the six artifacts.

P3 through P6 are optional. Routing all four into a revision would produce more prose than it removes risk, which is the over-engineering brake M6 exists to apply; surface them and let the orchestrator decide. P4's first half is the most worth taking, since it costs one sub-item and closes a real enumeration gap an implementer could fall through.

**Residual risk.** Two items the SPEC discloses and this audit could not eliminate. First, the baseline snapshot is regenerable and `plan.md` M5 makes regeneration routine, so a regression can be absorbed by one regeneration; the per-state tallies make it visible in the diff but do not prevent it, and the SPEC says so in three places rather than claiming otherwise. Second, the counters used throughout this audit — mine and the repair's — are reconstructions from the §3.1/§3.2 clause text, not the implementation the run-phase will build; they agree with each other on every input tested, which is evidence the clause text is unambiguous enough to reimplement twice, but it is not evidence about the eventual implementation. `research.md` §F records both.

**Gaps — what this audit did not observe.** `make build` was not run to completion (only the `agents-emit-check` gate, which passed). The 122 citation-axis and 85 alias-axis candidate populations were not adjudicated file by file; I verified only that the SPEC labels them upper bounds and never asserts them as debt. Coverage, lint, and the wider Go test suite were not run — nothing in this SPEC's plan-phase touches compiled code yet. The D7 verb was executed against the eleven referenced ids only, not transitively.
