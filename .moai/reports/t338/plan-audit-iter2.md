# SPEC Review Report: SPEC-AC-COUNT-DISCRIMINATOR-001

Iteration: 2/2 (Tier M ceiling reached)
Verdict: **FAIL**
Overall Score: **0.72** (Tier M threshold 0.80) — monotone improvement from iter-1's 0.67, no score regression, no STOP escalation
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t338`, branch `WT-ac-count-sweep`, HEAD `da03d9188`
SPEC version audited: `0.3.0` (REQ 8 / AC 8, `tier: M`)

Reasoning context ignored per M1 Context Isolation. Scoped to the iter-1 defect delta (D1–D14) plus regression, per the Tier M re-audit contract.

Cross-model second opinion NOT run: the SPEC directory is still untracked (`git status --porcelain` → `?? .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/`), so `codex_audit` / `glm_audit` collect an empty diff and return `inconclusive` by construction. Recorded as a gap, not a pass.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE '\*\*REQ-ACD-[0-9]+\*\*' spec.md | sort | uniq -c` → REQ-ACD-001…008, each exactly 1, zero-padded, no gap, no duplicate.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §4). All 8 conform; REQ-ACD-007 has been corrected from `Where` to `When` (`spec.md:270`), closing iter-1 D12. The `AC-ACD-*` entries are Given-When-Then, the correct verification-layer format, graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — 12 canonical fields present plus `tier: M` / `era: V3R6`; `grep -nE '^(created_at|updated_at|labels|spec_id):' spec.md` → rc=1.
- **[N/A] MP-4 Section 22 language neutrality** — the template-bound payload is the B12 clause, whose counter is a `grep`/`sort`/`wc` pipeline; the verifier is scoped to this repository's own Go tree. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 11 referenced SPEC IDs resolved; statuses `completed`×5, `implemented`×2, `in-progress`×1, `draft`×2. None retired/superseded/archived. `SPEC-SKILL-001` resolves under `.moai/specs/_archive/` (confirmed present) and is cited by the SPEC precisely as the archive-glob example — not a missing reference.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` on all four artifacts → `0 0 0 0`.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' spec.md plan.md acceptance.md progress.md` → rc=1. `research.md` correctly absent for Tier M **as declared** (see N1 — the declared Tier is itself in question).

**No must-pass failure.** The FAIL is driven by rubric score plus one Tier-determining defect.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirement layer is materially crisper. §3.2's exhaustiveness argument is sound over the new unit. Two residual ambiguities in the load-bearing adjacency definition, both measured: code-span placement (N4) and cross-line whitespace (N5). |
| Completeness | 0.70 | 0.50–0.75 | All sections present; frontmatter complete; 5× `### Out of Scope — …` H3 with `-` bullets. Two material gaps: §F omits a third distributed carrier of the B12 clause (N1, Tier-determining) and no artifact states what the corpus run does with a halting file (N3). |
| Testability | 0.65 | 0.50–0.75 | No weasel words (rc=1); the 53-vs-52 discrimination is genuinely binary and verified. But AC-ACD-001's mutant obligation is unsatisfiable (N2), its Given is placement-ambiguous (N4), AC-ACD-006 is indeterminate on the one halting file (N3), AC-ACD-007 permits a vacuous pass (N6). |
| Traceability | 0.80 | 0.75–1.0 | 8/8 both directions; §4 matrix matches the AC headings exactly; no orphan, no uncovered REQ; iter-1 D8 closed by AC-ACD-005 item 3. One requirement is partially unverified: REQ-ACD-007's mirror obligation reaches a mirror surface no AC asserts (N1). |

Harmonic mean = 4 / (1/0.75 + 1/0.70 + 1/0.65 + 1/0.80) = 4 / 5.5504 = **0.7207 → 0.72**. Below the 0.80 Tier M threshold (SSOT: `spec-workflow.md:141`).

---

## Regression Check — iter-1 defects D1–D14

All fourteen were re-verified against the tree, not against the repair's own narration.

| # | Status | Establishing evidence |
|---|---|---|
| **D1** | **RESOLVED** | Convention rebound from line to identifier occurrence with an adjacency rule (`spec.md:163-204`, REQ-ACD-002 `:242`). Verified by implementing both semantics and running them on the decisive input — see § D1 verification below. The corpus is additionally **inert** under the new convention: the counter run over all 602 depth-1 files produces zero false exclusions (`files-with-excluded=0`), which measures the §3.1 claim that prose does not produce bracketed tokens. |
| **D2** | **RESOLVED** | Alias axis modeled by widening `[REF]` (`spec.md` §2.3, token table `:170`) with no third token, no new REQ/AC, no counting-rule change. `alias-shape-scan.sh` re-run: **byte-identical** to the committed `.txt`, `grep -c '^=== '` → **85**, stderr 0 bytes. Both hand-checks verified independently: `SPEC-HUMANIZE-001:300` uses `AC-001` for the criterion declared `AC-HUM-001` at `:32` (**true alias**); `SPEC-STATUSLINE-001` `AC-SL-001` `:14` and `AC-SL-NF-001` `:216` are separate declared criteria (**shape coincidence**). The corrected scan is sound: identifiers arrive on stdin, not via `-v`, so the BSD-awk newline rejection that produced the first 0-return cannot recur, and the run demonstrably produces output rather than an empty result. |
| **D3** | **RESOLVED** | AC-ACD-007 now re-derives the target set every run and records per-flag hand adjudication in `progress.md §E.2` (`acceptance.md:121-140`); the 29-flag list is cited nowhere binding. The `SPEC-V3R2-ORC-001` false positive is hand-verified in §8 `:342`; the arithmetic is corrected coherently (29−8 = 21 identifiers, 18−4 = 14 files). Then-2 adds the false-positive regression check (live-adjudicated identifiers must be counted live). Residual: no non-vacuity guard — N6. |
| **D4** | **RESOLVED** | The literal 134 survives only in the HISTORY row describing its deletion (`grep -n '134'` → `spec.md:24` only). REQ-ACD-008 `:272` mandates per-axis sections with bound direction; AC-ACD-008 item 3 makes the no-total rule grep-decidable. |
| **D5** | **RESOLVED** | REQ-ACD-006 `:262` pins the depth-1 glob and excludes `_archive/`; AC-ACD-006 item 1 re-derives corpus size at run time; the frozen 601 is gone. |
| **D6** | **RESOLVED** | REQ-ACD-005 `:256` requires a sentinel comment pair in **both** files, and AC-ACD-005 item 1 asserts each extraction yields **exactly one non-empty** command (0 / ≥2 / empty all fail). The sentinel string itself is deferred to M2, which is a decision milestone — acceptable. |
| **D7** | **RESOLVED-with-recorded-residual** | Judgment asked for: the per-state tallies **do** reduce the rubber-stamp risk rather than merely documenting it — a regeneration diff that only moves a total is unreviewable, whereas a `live → excluded` migration is visible per file and per state. It does not eliminate the risk, and the SPEC says exactly that rather than claiming a fix (`spec.md:268`, `acceptance.md:117`, `progress.md §E.1` residual ②). That honesty is the correct disposition. Minor blemish: AC-ACD-006 item 4's "the regeneration commit states which files moved and why" is a process obligation with no mechanical test. |
| **D8** | **RESOLVED** | AC-ACD-005 item 3 asserts the `manager-docs.md` pair `diff` produces no output, covering the prose half. Baseline re-measured this run: `diff .claude/agents/moai/manager-docs.md internal/template/templates/.claude/agents/moai/manager-docs.md` → no output (**byte-identical**). Item 4's 171-line claim also re-measured: exactly `171c171`. Both ACs rest on measured facts. |
| **D9** | **RESOLVED** | `plan.md:84-88` names the old ground as wrong, disclaims the halt property for the 34-file base case explicitly, and replaces it with three grounds (authorship context, timing of value, locality of the residual). The contradiction with L82 is gone. |
| **D10** | **RESOLVED** | Both paths named: `.moai/reports/t338/debt-list.md` and `.moai/reports/t338/ac-count-baseline.txt` (`plan.md:152-153`, `acceptance.md:108,146`, §F rows 7 and 11). Additionally verified committable — `git check-ignore -v` → rc=1 on both, and 696 files under `.moai/reports/` are already tracked, so AC-ACD-006's "committed snapshot" is achievable at the chosen path. |
| **D11** | RESOLVED | `acceptance.md:46` now cites the three scan outputs, not §3. |
| **D12** | RESOLVED | REQ-ACD-007 `:270` is `When …`. |
| **D13** | RESOLVED | §F row 13 adds `progress.md`. |
| **D14** | RESOLVED | `spec.md:326` extends the neutrality constraint to the revised clause itself; AC-ACD-005 item 5 asserts it mechanically. |

**No unresolved iter-1 defect. No stagnation.**

### The audit claim the repair overturned — CONFIRMED, independently

iter-1 D13 excluded `internal/template/catalog.yaml` from §F on the grounds that the `manager-docs` entry already exists. **That judgment was wrong and the repair is right.** Reproduced in this run:

```console
$ shasum -a 256 internal/template/templates/.claude/agents/moai/manager-docs.md
27d6252a33131be637294ddced274213bb817012747731d08844becd7a3b7954  …
$ sed -n '122,125p' internal/template/catalog.yaml
            - name: manager-docs
              tier: core
              path: templates/.claude/agents/moai/manager-docs.md
              hash: 27d6252a33131be637294ddced274213bb817012747731d08844becd7a3b7954
$ grep -n 'gen-catalog-hashes' Makefile
24:	@go run ./internal/template/scripts/gen-catalog-hashes.go --all
$ grep -c 'manager-develop-prompt-template' internal/template/catalog.yaml
0
```

The entry carries a **content hash**, `make build` regenerates it, so editing the mirror necessarily changes a tracked file. §F row 12 is correct, and the prompt-template pair correctly has no such row. The consequence the prompt flagged is real and is why the enumeration was audited hard — see N1.

### D1 verification (the critical repair, tested rather than read)

Both semantics were implemented from `spec.md` §3.1/§3.2 verbatim and run on the decisive input. Normalization was applied to an out-of-tree copy; the original `completed` file was never touched.

```console
sweep original                          :  54
adjacency counter, original             :  COUNT 54  (live=54 excluded=0 ambiguous=0)
adjacency counter, normalized copy      :  COUNT 53  (live=53 excluded=1 ambiguous=0)   <- AC-ACD-001 item 2 ✓
LINE counter (D1 regression), normalized:  COUNT 52  (live=52 excluded=2 ambiguous=0)   <- AC-ACD-001 item 3 ✓
adjacency counter, END-OF-LINE MUTANT   :  COUNT 54  (live=54 excluded=0 ambiguous=0)   <- AC-ACD-001 note says 52  ✗
LINE counter, END-OF-LINE MUTANT        :  COUNT 52
```

Reading, in order:

1. **The convention repair is genuine.** 53 and 52 are mechanically distinguishable on the exact input that broke the v0.2.0 rule, and a line-semantics implementation is caught by item 2 (it returns 52 where 53 is required). The collision is dissolved rather than papered over.
2. **The exhaustiveness claim survives adversarial probing.** Constructed inputs: a floating token adjacent to nothing (no effect — the SPEC states this residual at `:204`, and stating it is sufficient because it provably cannot change any identifier's state); a token between two identifiers (marks the preceding one, deterministic, and AC-ACD-004's fixture row 2 covers it); repeated tokens (`[REF] [RETIRED]` — same effect, §3.2 says the tokens are interchangeable for counting); same-prefix collision on one line (dissolved by identifier granularity). {none, all, some} does partition the space **given** that "marked" is a total predicate on occurrences — which is where N5 bites.
3. **The mutant obligation is inverted** — N2 below.

---

## Defects Found (structured defect-list)

**N1. §F omits a third distributed carrier of the B12 clause; `make build` fails as M4 is written; the file count crosses the Tier M ceiling** — `plan.md:159-179` (§F), `plan.md:138-145` (M4), `spec.md:325` (§7), `acceptance.md:92-102` (AC-ACD-005) — **Severity: critical — Class: blocking**

`internal/template/templates/.codex/agents/moai/manager-docs.toml` carries the **whole B12 clause**, counter command included:

```console
$ sed -n '65,76p' internal/template/templates/.codex/agents/moai/manager-docs.toml
Before appending to `CHANGELOG.md` … MUST run 3 self-tests …
2. **AC count match**: …
   grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/<SPEC-ID>/acceptance.md | sort -u | wc -l
   **A count of 0 is a RED flag, not a pass.** …
```

It is a tracked artifact generated from the `.md` layer, and the coupling is enforced by the build. Proven by mutation (one character in the mirror, restored immediately after):

```console
$ # baseline
$ AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.758s

$ # one-character edit to internal/template/templates/.claude/agents/moai/manager-docs.md
$ AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.01s)
    golden_test.go:109: .codex/agents/moai/manager-docs.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
FAIL

$ git restore internal/template/templates/.claude/agents/moai/manager-docs.md
$ shasum -a 256 internal/template/templates/.claude/agents/moai/manager-docs.md
27d6252a33131be637294ddced274213bb817012747731d08844becd7a3b7954  …          # back to baseline
$ AGENTEMIT_UPDATE= go test … -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.366s
```

Three consequences, in ascending order of cost:

1. **M4 does not complete as written.** `Makefile:23` reads `build: agents-emit-check templ-generate`, and `agents-emit-check` is deliberately read-only (`Makefile:31-38`: "regeneration stays behind the explicit `agents-emit` verb"). So M4's `make build` **hard-fails** after the mirror edit until `make agents-emit` is run. M4 lists only `make build`.
2. **The distributed codex agent keeps the old clause**, and no AC notices. `spec.md` §7 asserts "두 정의 지점 모두 템플릿 미러를 가진다" — there are **three** carriers of this clause, not two, and AC-ACD-005 asserts identity for the `manager-docs.md` pair only. This is the twin-drift failure `plan.md` §G names as its first anti-pattern, reproduced one surface over.
3. **The Tier is wrong.** §F's own counting rule (fixture = 3 files) enumerates rows 1-13 as **15 files**; this omission makes it **16**. `plan.md:175` and `progress.md §E.1` both state "상한 15 에 닿아 있다 — 여기서 파일이 더 붙으면 L 이다". Per `spec-workflow.md:142`, Tier L means: `> 15 files`, a **5-artifact** set (`design.md` + `research.md`, both absent), a **0.85** PASS threshold, a plan-audit ceiling of 3 rather than 2, and Route B (per-phase PR) rather than Route A main-direct. The Tier is not cosmetic here — it changes the artifact set, the threshold, the iteration ceiling, and the git route.

**Required fix**: add the `.codex` TOML row to §F; add `make agents-emit` to M4 ahead of `make build`; correct §7 to name three carriers; extend AC-ACD-005 with an assertion covering the codex emission surface (the `agents-emit-check` test is the ready-made mechanism); then re-derive the Tier from the corrected enumeration and, if it lands at L, produce the Tier L artifact set or shrink the card's scope.

---

**N2. AC-ACD-001's mutant obligation is inverted — no correct implementation can satisfy the AC** — `acceptance.md:40`, `plan.md:134` — **Severity: major — Class: blocking**

The note reads: "사본의 `[REF]` 를 … **행 끝**으로 옮기면 …, 계수기는 `52` 를 내야 하고 이 AC 는 실패해야 한다." Measured (§ D1 verification above):

| implementation | normalized copy | end-of-line mutant |
|---|---|---|
| adjacency (correct, §3.1) | **53** ✓ item 2 | **54** ✗ note demands 52 |
| line semantics (the D1 regression) | **52** ✗ item 2 | 52 ✓ note |

52 on the mutant is produced **only** by the implementation the SPEC exists to forbid. So the AC's items 2/3 and its mutant note are mutually unsatisfiable: a correct counter fails the mutant clause, and the counter that satisfies the mutant clause fails item 2. The same wrong literal is repeated in `plan.md:134` ("계수기가 53 대신 52 를 내야 한다"), so a run-phase author reading either artifact reaches the same dead end. AC-ACD-001 is the sole mechanical guard of the critical D1 repair, and the DoD gates it on observing the mutant die.

**Required fix**: the end-of-line mutant's expected value is **54**, not 52 — nothing is adjacent, so nothing is excluded. Correct both `acceptance.md:40` and `plan.md:134`. The regression signature `52` belongs to item 3, where it is already correct.

---

**N3. This SPEC's own `acceptance.md` halts the counter, and it is inside AC-ACD-006's corpus** — `acceptance.md:77-78`, `acceptance.md:106-117` (AC-ACD-006), `acceptance.md:173` (DoD) — **Severity: major — Class: blocking**

The adjacency counter was run over the whole depth-1 corpus. Exactly one file halts, and it is this SPEC's own:

```console
$ bash .moai/reports/t338/iter2-scratch/corpus.sh | tail -2
HALT .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/acceptance.md :: AMBIGUOUS AC-SYN-010 AC-SYN-012
files=602  halting=1  files-with-excluded=0
```

Cause is the convention's own documentation table. `AC-SYN-010` occurs twice on `:77` — once marked (`AC-SYN-010 [REF]`, the fixture column) and once unmarked (`AC-SYN-010` excluded, the expectation column); `AC-SYN-012` likewise on `:78`. Partial marking is exactly the `ambiguous` state, so the counter returns no integer and a non-zero exit code.

AC-ACD-006 runs the counter across `.moai/specs/*/acceptance.md`, which includes this file — the SPEC says so itself (`spec.md:105`: "602 는 이 SPEC 의 `acceptance.md` 가 포함된 값이다"). Nothing in REQ-ACD-006, AC-ACD-006, or the snapshot format states what a corpus run does with a halting file: either the corpus run fails (AC-ACD-006 unsatisfiable) or the shipped baseline permanently records a halt. The DoD's deferral (`:173` — this file is a 선택 A subject, normalized at its own sync) assumes the residual is a **silent over-count**; it is not, it is a halt, and the halt fires at M3-c rather than at some future sync.

Note the cheap fix is closed off: skipping code spans or fenced blocks is markup anchoring, which REQ-ACD-004 forbids. The available fixes are distinct synthetic identifiers per column (`AC-SYN-010` in the fixture column, `AC-SYN-110` in the expectation column), or an explicit corpus exclusion for the file that documents the convention — the second of which needs stating in REQ-ACD-006, not left to the implementer.

---

**N4. AC-ACD-001's Given is placement-ambiguous where the marked occurrence sits inside a code span** — `acceptance.md:30`, `spec.md:172` — **Severity: major — Class: blocking**

The AC says to attach `[REF]` "바로 뒤" of the `AC-DCP-010` occurrence. In the real line that occurrence is inside a code span immediately followed by `` ` `` and `(`:

```
… `spec.md`가 superseded AC를 **ID로 인용** — `AC-DCP-010`(`acceptance.md:79` …
```

Two plausible readings, measured:

```console
inside the span   `AC-DCP-010 [REF]`(`acceptance.md:79` …   →  COUNT 53   (AC passes)
outside the span  `AC-DCP-010` [REF](`acceptance.md:79` …   →  COUNT 54   (AC fails)
```

The outside placement is what a reader of the rendered markdown would naturally write, and it is not "spaces only" between the identifier and the token — a backtick intervenes — so it marks nothing. The SPEC's single worked example of its own convention therefore depends on an unstated authoring rule, and the AC's verdict flips on it.

**Required fix**: state in §3.1 that intervening non-space characters — a closing backtick included — break adjacency, and pin the intended placement in AC-ACD-001's Given verbatim.

---

**N5. Cross-line adjacency is undefined in the load-bearing clause** — `spec.md:172`, `spec.md:192` — **Severity: minor — Class: blocking**

"사이에는 공백만 둔다" / "그 등장 바로 뒤에 공백만 두고" does not say whether a newline counts as one of those spaces. Two conformant implementations therefore disagree, measured on a 6-line fixture where one identifier ends a line and the token opens the next: same-line-only reading → 4 live; whitespace-including reading → 3 live. The count differs, and no AC pins it — AC-ACD-004's fixture covers four cases (adjacent, same-line non-adjacent, token-first, floating) and cross-line is not among them.

This belongs in M1 by the SPEC's own reasoning: M1 is the irreversible milestone ("토큰 형태가 바뀌면 이미 정규화한 파일과 배포된 절이 모두 갈린다"), and the fix is a five-word qualification ("same line, spaces and tabs only").

---

**N6. AC-ACD-007 permits a vacuous pass** — `acceptance.md:121-138` — **Severity: minor — Class: blocking**

The Given re-derives the target set, filters terminal status, and drops hand-adjudicated false positives. If that leaves zero files, Then-1 ("정규화한 **모든** 파일에서 …") is vacuously true and the AC passes having asserted nothing. The SPEC applies precisely this guard elsewhere — AC-ACD-005 item 1 asserts the extraction is "정확히 1건이며 비어 있지 않음" for exactly this reason, and §1 makes `0 == 0` vacuity the clause's founding observation. The asymmetry is internal inconsistency, not a hypothetical.

**Required fix**: require at least one normalized file, or require that a zero-target outcome be recorded with its per-flag adjudication rather than passing silently.

---

**N7. The 123/56 collision figure is frozen, is now stale in this same tree, and carries no bound direction** — `spec.md:182-183`, `spec.md:244`, `progress.md` §E.1 repair record — **Severity: minor — Class: optional**

```console
$ bash .moai/reports/t338/collision-scan.sh
lines carrying >=2 distinct AC prefixes: 125
files containing such a line          : 56
```

125, not the quoted 123, at this same HEAD. The corpus includes this SPEC's own `acceptance.md`, whose collision-line count grew from 1 to 3 in the v0.3.0 rewrite — the figure moved because the SPEC that quotes it was edited. This is the frozen-literal hazard the SPEC itself legislates against (AC-ACD-006 item 1; REQ-ACD-008's "재유도"), applied everywhere except to its own evidence.

Separately, the figure is unlabeled for direction while REQ-ACD-008 makes direction labelling mandatory for every other figure — and it is neither a clean upper nor a clean lower bound: it over-counts (a traceability row citing two legitimately-native families is not a live/non-live collision) and under-counts (a collision between two same-prefix identifiers on one line is invisible to a distinct-prefix scan). D1's repair does not depend on it — the decisive instance is hand-verified — so this is an evidence-hygiene defect, not a structural one.

---

**N8. §F's "총 13-15 파일" conflates rows with files** — `plan.md:175`, `progress.md` §E.1 — **Severity: minor — Class: optional**
By §F's own stated rule ("fixture 를 3파일로 셈"), rows 1-13 enumerate 5 + 3 + 1 + 3 + 1 + 1 + 1 = **15 files** — a single value, not a 13-15 range; 13 is the row count. The range reads as headroom that does not exist, which is what made N1's single omission decisive. **Fix**: state the file count as one number derived from the rows.

**N9. §F row 6 names `internal/spec/testdata/account/`** — `plan.md:168` — **Severity: minor — Class: optional**
Reads as a typo for `ac_count` / `acceptance`; the verifier file in row 5 is `ac_count_clause_test.go`. **Fix**: correct the directory name.

**N10. AC-ACD-008 item 4 gives no re-derivation source for the retirement section** — `acceptance.md:160` — **Severity: minor — Class: optional**
Item 4 requires each section's item count to match "그 실행의 스캔 재유도 값" but names sources only for the citation axis (`pre-terminal-scan.sh` `status=completed`) and the alias axis (`grep -c '^=== '`). The retirement section's figure is hand-adjudicated from a lower-bound detector, so it has no re-derivable machine value and item 4 is undecidable for that section. **Fix**: exempt the retirement section explicitly, or name the adjudication record as its source.

**N11. AC-ACD-002's three-scan exclusion adds no discriminating power for the citation axis** — `acceptance.md:46` — **Severity: minor — Class: optional**
`pre-terminal-scan.txt` enumerates only the 34 pre-terminal files, not the 122 `completed` multi-domain ones (`grep -c 'acceptance.md'` → 34 of 39 lines), so a citation-axis over-count file can pass the exclusion filter. Harmless in outcome — an unmarked file yields sweep == counter either way, so the AC cannot false-fail — but the Given claims a guarantee it does not provide. **Fix**: either scope the exclusion to what the outputs actually enumerate, or state that the citation axis is filtered only for pre-terminal files.

---

## What the audit re-confirmed as sound (do not disturb)

- The three-state partition over identifier occurrences, tested against constructed adversarial inputs (§ D1 verification, reading 2).
- The reserved tokens are **inert across the 601 legacy files** — `files-with-excluded=0` over the corpus run. The §3.1 claim that prose does not produce bracketed literals is now measured rather than argued, and it is the strongest single piece of evidence for the design.
- `[REF]` widening absorbs the alias axis with no token, REQ, AC, or counting-rule growth — the minimal available repair for D2.
- Bound-direction discipline in REQ-ACD-008 and AC-ACD-008 (no totals; upper-bound axes labeled candidates, not debt) — the correct disposition of iter-1 D4, and grep-decidable.
- Both mirror baselines are measured facts, re-verified this run: `manager-docs.md` pair byte-identical; prompt-template pair exactly `171c171`.
- The `catalog.yaml` hash coupling — the repair's correction of the audit stands (see above).
- Requirement/verification-layer hygiene: no `should`/`may` in REQs, no weasel words in ACs, 8/8 traceability both directions, no `[NEEDS CLARIFICATION]`, `syscall` absent.
- The accepted 34-file citation residual remains honestly recorded in all its declared places (`spec.md:121`, `plan.md` M0 "남는 것", `progress.md` §E.1 residual ①, DoD `:172-173`); nothing overclaims legacy coverage.

---

## Verification integrity

**Claim** — the verdict, the four dimension scores, the D1–D14 regression dispositions, and defects N1–N11.

**Evidence** — commands and verbatim output are inline per row and per defect.

**Baseline-attribution** — every measurement was executed in this run, in this worktree (`git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t338`), at HEAD `da03d9188`. Tracked-tree state before and after: `git status --porcelain` → `?? .moai/reports/t338/`, `?? .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/` only.

**Mutations and restores** — one tracked file was mutated, for N1: `internal/template/templates/.claude/agents/moai/manager-docs.md` (one character). Restored with `git restore`; sha256 verified back to `27d6252a…3b7954` and the drift check verified `ok` afterwards, both shown in N1. The D1 counter experiments ran on out-of-tree copies of `SPEC-AGENT-PARALLEL-OPT-001/acceptance.md`, which was never touched; those copies were deleted after use. The audit's own scripts remain at `.moai/reports/t338/iter2-scratch/` (`counter.py`, `corpus.sh`, `perfile.sh`, `d7.sh`) so every cited command resolves at audit time.

**Gaps** — (a) cross-model backends not consulted (untracked SPEC ⇒ empty diff ⇒ `inconclusive` by construction); (b) the alias axis's true value was not measured — only the 85 upper bound reproduced and two directions hand-checked, which is what the SPEC itself claims; (c) the citation axis's 122 was not verified file-by-file, consistent with the SPEC labeling it a candidate list; (d) `make build` was not run end-to-end — N1's build failure is established from `Makefile:23`'s dependency plus the observed `agents-emit-check` failure, not from a full build; (e) the local `.claude/agents/moai/manager-docs.md` was not separately tested as an emission source, since the mirror alone establishes the coupling.

**Residual-risk** — the corpus counter used for N3 is my reimplementation of §3.2, not the SPEC's future implementation; a different-but-conformant implementation would still halt on partial marking, since that is the defining rule, but the specific identifier set could differ if the adjacency reading differs (which is N5). N1's file count assumes the `.codex` TOML is the only unenumerated carrier; a further generated surface would push the count higher still, not lower.

---

## Recommendation

**FAIL at 0.72 against the 0.80 Tier M threshold. Run-phase entry is blocked.**

The repair is substantial and honest: all fourteen iter-1 defects are resolved, the two criticals are resolved by measurement rather than by assertion, the score moved 0.67 → 0.72 with no regression, and the SPEC corrected the audit where the audit was wrong. What blocks entry is not the convention — that now holds up under adversarial testing — but three things around it.

Two of them make run-phase entry unsafe rather than merely imperfect:

1. **N1 must be resolved before M4, and re-scoped before kickoff.** `make build` fails as M4 is written, the distributed codex agent would silently keep the superseded clause, and the corrected enumeration crosses the Tier M ceiling — which changes the required artifact set, the PASS threshold, the audit ceiling, and the git route. This is a Tier decision, not a wording fix, and it belongs to the operator.
2. **N2 must be resolved before M3.** AC-ACD-001 is the sole mechanical guard of the critical D1 repair, its mutant obligation is unsatisfiable by any correct implementation, and the same wrong literal sits in `plan.md` M3. Left as is, the run-phase author either fails the DoD or silently rewrites the AC to fit the code — the failure mode this SPEC exists to prevent.

Ordered fix route for `manager-spec`:

1. Add the `.codex` TOML to §F; add `make agents-emit` to M4; correct §7's "two definition points"; extend AC-ACD-005 to the codex surface; re-derive the file count and settle the Tier with the operator (N1, N8).
2. Correct the end-of-line mutant's expected value to **54** in `acceptance.md:40` and `plan.md:134` (N2).
3. Decide how the corpus run treats this SPEC's own halting file — distinct synthetic ids per column, or an exclusion stated in REQ-ACD-006 (N3).
4. Pin adjacency precisely: intervening non-space characters break it, and it is same-line only (N4, N5) — both in M1's wording, the irreversible milestone.
5. Add a non-vacuity guard to AC-ACD-007 (N6).
6. Re-derive or drop the frozen 123/56 and label its direction (N7); sweep N9–N11 at the author's discretion.

**Iteration ceiling.** This is iteration 2 of the Tier M ceiling of 2 (`harness.plan_audit_tier_ceilings`). The verdict is FAIL, so per the Retry Loop Contract the orchestrator escalates to the operator with the three options — PASS-with-debt, scope reduction, or an explicit override extending the ceiling. Note that N1 bears on the ceiling itself: if the SPEC is Tier L, the ceiling is 3 and a third iteration is available without an override. Verdict authority stays with this agent either way; the ceiling bounds iteration count, never the verdict.
