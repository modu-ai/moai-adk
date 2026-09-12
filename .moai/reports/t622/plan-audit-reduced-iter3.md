# SPEC Review Report: SPEC-GIT-DELIVERY-PROCEDURE-001 (reduced, revision 0.2.3)
Iteration: 3 (limited, operator-authorized override beyond the Tier M ceiling of 2 — scope restricted to iteration-2 D1–D8 resolution and regressions introduced by the 0.2.3 change set)
Verdict: **PASS**
Overall Score: **0.92** (harmonic mean; Tier M PASS threshold 0.80). Iteration 2 was 0.75, so the score went up and no STOP-on-regression applies.

Reasoning context ignored per M1 Context Isolation. The lane reproduction record (`.moai/reports/t622/lane-repro-reduced-iter2-fixes.md`) was treated as a claim; every judge below was re-run by this auditor on its own fixtures.

## Tree attribution

```
$ git rev-parse --show-toplevel   → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t622
$ git branch --show-current       → WT-git-procedure-fixes
$ git log --format='%h %s' -4     → ba62ff218 docs(t622): lane reproduction of reduced audit iteration 2 detector fixes
                                    8cf5aaf06 feat(SPEC-GIT-DELIVERY-PROCEDURE-001): targeted detector fixes for reduced audit iteration 2
                                    0af445525 docs(t622): reduced plan-audit iteration 2 verdict for SPEC-GIT-DELIVERY-PROCEDURE-001
                                    3512b9f75 feat(SPEC-GIT-DELIVERY-PROCEDURE-001): complete revision 0.2.2 consumer scope
$ git status --porcelain          → (empty)
$ git diff --stat 3512b9f75 8cf5aaf06 -- .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/
  acceptance.md 189 ±, plan.md 24 ±, progress.md 19 ±, spec.md 23 ± (4 files, +189 −66)
$ git diff --stat b412f8a33 HEAD -- .claude/agents .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai .claude/commands/moai .agents/skills/moai-sync .claude/rules/moai/core/zone-registry.md internal/template Makefile docs-site/content
  → (no output)
```

Base-fragment measurements below were taken from the template copies on HEAD `ba62ff218`. The diff above shows those copies are byte-identical to `b412f8a33`. Fixtures live in session scratch and are not exported; every fixture line is quoted verbatim below so the runs can be reproduced.

## Must-Pass Results (change-set regression only)

- [PASS] **MP-1** — `grep -o -E '^\*\*REQ-GDP-[0-9]{3}\*\*' spec.md | sort | uniq -c` → `req_lines=26 dup=0`. AC headings `001–007, 013–017, 025–030`, the same set as iteration 2. The table still reads "판정 대상: 16개" (acceptance.md:L39 context).
- [PASS] **MP-2** (requirement layer, spec.md) — the change set touches normative text only in REQ-GDP-005 (L141) and REQ-GDP-024 (L184), and only to replace "열 개" with "아홉 개". Their GEARS forms are unchanged (Unwanted). REQ-GDP-025 (L189) and REQ-GDP-026 (L192) do not appear in the diff.
- [PASS] **MP-3** — spec.md:L1-L14. `version: "0.2.3"` is quoted, and all 12 fields are present (verbatim head read). No rejected aliases.
- [N/A] **MP-4** — not multi-language tooling.
- [PASS] **MP-5 D7** — the change set adds no SPEC ID. The 0.2.3 HISTORY row cites only report paths. `mcp__moai__spec_audit(project_root=<this worktree>, filter_spec=…)` → `{"modern_era_clean":1,"drift_findings":[]}`.
- [PASS] **MP-6 D8** — `grep -c -e syscall -e '\[NEEDS CLARIFICATION'` on the four files → `0 0 0 0`.
- [PASS] **MP-7** — the same command gives 0 on plan.md. There is no research.md (Tier M).
- Hygiene: `perl -CSD` `\p{Cf}` over the four files → `cf=0`.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 1.00 | full | D5 resolved: REQ-GDP-005 L141 and REQ-GDP-024 L184 now say "범위 파일 아홉 개", and §C.4 L235 states the counting rule and 19 (+2 → 21). No remaining "열 개" / "나머지 아홉" / "뺀 아홉" / "위 아홉" hits in the four files (sweep below). |
| Completeness | 1.00 | full | The change set removes no section. §E.2 L303 is refreshed, and a new §E.2 bullet declares that AC-027 is automatic-only. |
| Testability | 0.75 | one AC not precisely binary | D1–D3 closed: the auditor's mutants now fail, and AC-026/028/029 carry mandatory reading records (§D.3 L944). AC-027 still passes four out-of-list wrong phrasings with no reading gate — declared residual (N5). Fail-closed false-fail wordings remain (N4). |
| Traceability | 1.00 | full | AC table rows L45-L48 keep REQ-GDP-025 → 026·027 and REQ-GDP-026 → 028·029; only the description column changed. |

Harmonic mean: 4 / (1 + 1 + 1/0.75 + 1) = 4 / 4.333 = **0.92**.

## D1–D8 resolution

### D1 — AC-026 presence and direction — **RESOLVED**

Change: (ii) is directional (L655), (iii) is a presence guard (L658), and a reading record is added (L660-L663), which §D.3 L944 lists.

Judge (ii) copied verbatim from L655, run with `E=<scratch>/e026` over the nine fragment files. The eight other fragment files are empty, and `ac026-skill.md` holds this 20-line fixture:

```
 1 - `--merge`: formerly a deprecated alias of `--auto-merge`, now the primary auto-merge flag      (wrong: un-deprecated)
 2 - `--merge`: deprecated alias of `--auto-merge`; unlike `--auto-merge` it also skips CI checks   (wrong: own behavior)
 3 - `--auto-merge`: deprecated alias for `--merge`                                                  (wrong: reversed)
 4 - `--merge` replaces `--auto-merge`, which is a deprecated alias of `--merge`                     (wrong: reversed)
 5 - `--merge`: deprecated alias of `--auto-merge`, but it is no longer deprecated                   (wrong)
 6 - `--merge`: Deprecated alias of `--auto-merge`.                                                  (correct)
 7 - `--merge` (deprecated alias for `--auto-merge`; logs a deprecation warning)                     (correct)
 8 - `--merge`: deprecated; alias of `--auto-merge`                                                  (correct meaning, off-shape)
 9 - `--auto-merge`: merge the PR after sync. `--merge` is a deprecated alias of `--auto-merge`.     (correct)
10 - `--merge`: "deprecated alias of --auto-merge"                                                   (correct)
11 - `--merge`: deprecated alias of `--auto-merge-now`                                               (wrong target)
12 - `--merge`: a deprecated alias of `--auto-merge` that is not deprecated in team mode             (wrong)
13 /moai sync [mode] [--pr] [--auto-merge] [--merge: deprecated alias of --auto-merge] [--skip-mx]  (correct)
14 - `--merge`: deprecated alias of `--auto-merge` (still supported, never removed)                  (correct)
15 - `--merge`: undeprecated alias of `--auto-merge`                                                 (wrong)
16 - `--merge`: auto-merge the PR; the deprecated alias of `--auto-merge` warning is gone            (wrong)
17 - `--auto-merge`: deprecated alias of `--merge`; use `--merge` to auto-merge the PR               (iter-2 auditor line)
18 - `--merge`: auto-merge the PR after sync (the older `--auto-merge` spelling is deprecated)       (iter-2 auditor line)
19 - `--merge`: alias of `--auto-merge`                                                              (iter-2 auditor line)
20 - `--merge`: deprecated alias of `--auto-merge` (logs a warning)                                  (iter-2 auditor line, correct)
```

```
test_s=0
(ii) hit lines: 3 4 5 8 11 12 15 17 18 19
merge-lines (L660 selector) lines: 1 2 … 20   (all 20)
```

- Iteration-2 reversed and un-deprecated lines 17 and 18, and the undeprecated line 19, all hit. The correct line 20 does not. This confirms the lane's claim.
- Lexical escapes 1, 2, and 16 are wrong but pass (ii). All three sit in `ac026-merge-lines.txt`, so the mandatory reading question at L154-L156 ("뒤집힌 관계, 별개 플래그, 폐기되지 않은 플래그, 자체 동작을 가진 플래그가 아닌가") covers them. Iteration 2 explicitly allowed reading-gating as a fix.
- Correct lines 6, 7, 9, 10, 13, and 14 pass. Line 8 false-fails (fail-closed; see N4).

Judge (iii), L658:

```
iteration-2 full-removal line "Modes: auto, force, status, project. Flags: --auto-merge, --skip-mx"  → 0 (exit=1) → FAIL
partial removal (skill keeps alias; sync "**Flags**: --pr, --auto-merge, --skip-mx"; qgc-flags keeps alias; dl Step 3.4 only `--auto-merge` and `--merged-only`)
  → skill.md:1  sync.md:0  qgc-flags.md:1  dl.md:0  (grep exit=0; judged by counts as L659 says) → FAIL
base (template):  skill 1, sync 1, qgc-flags 1, dl 2; merge-lines 10 lines (test_s=0)
```

Base (ii) with the 0.2.3 judge on the template fragments: `test_s=0`. The same 10 lines as iteration 2 hit: skill:1, ref:2, qgc-args:4, qgc-flags:4, sync:1, dl:9, dl:20, sync-usage:1, hint:1, dl-next:9. Base (i): all seven counted files 0. Both match acceptance.md:L632-L636.

### D2 — AC-028 all-approvals — **RESOLVED**

Change: (a) requires "all" within two words before "approv" (L777). (b) is widened (L780). A reading record is added over `ac028-mode-lines.txt` (L783-L786) and listed in §D.3.

Fixture `e028/ac028-mg.md`, 24 lines, with an empty `ac026-dl.md`. The (a) predicate from L777 was printed per line; (b), the mode-lines selector, and 029 (b) were run as written.

```
 1 In team mode, `--auto-merge` merges once most approvals are in.
 2 In team mode, `--auto-merge` merges after approval from a maintainer.
 3 In team mode, `--auto-merge` merges before all approvals are obtained.
 4 In team mode, `--auto-merge` does not wait for all approvals.
 5 In team mode, `--auto-merge` merges when not all approvals are in.
 6 In team mode, `--auto-merge` merges only after all required approvals; a single approval is not enough.   (correct)
 7 In team mode, `--auto-merge` requires approval from all reviewers.                                          (correct)
 8 In team mode, `--auto-merge` merges once all reviewers approve.                                             (correct)
 9 In team mode, `--auto-merge` merges after all approvals are obtained; in personal and manual modes it merges without an approval condition.  (correct, combined)
10 `--auto-merge` merges as soon as CI checks pass.                                                            (wrong, mode-less)
11 In team mode, `--auto-merge` merges only after all approvals; maintainers can override approvals.
12 In personal and manual modes, `--auto-merge` merges without waiting for review.                             (correct)
13 In personal and manual modes, `--auto-merge` merges after the owner signs off.                              (wrong)
14 In personal and manual modes, `--auto-merge` merges immediately; no review or approval is needed.           (correct)
15 In personal and manual modes, `--auto-merge` merges without requiring approval unless a teammate requests review. (wrong)
16 In team mode, `--auto-merge` merges without all approvals.
17 In team mode, `--auto-merge` merges with most approvals.
18 In team mode, `--auto-merge` merges only after all approvals are obtained.                                  (correct)
19 In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve). (correct)
20 In team mode, `--auto-merge` merges once CI passes; approvals are optional.                                 (iter-2 auditor line)
21 In team mode, `--auto-merge` merges after at least one approval.                                            (iter-2 auditor line)
22 In personal and manual modes, `--auto-merge` merges after a code review.                                    (iter-2 auditor line)
23 In personal and manual modes, `--auto-merge` merges without requiring approval.                             (iter-2 auditor line, correct)
24 In personal and manual modes, `--auto-merge` merges; approval is not needed.                                (iter-2 auditor line, correct)
```

```
028 (a) per line : 3 4 5 6 8 9 11 16 18
028 (b)          : 6 9 20 21                  (028b_test_s=0 on the full fixture)
mode-lines       : 1–9, 11–24                 (line 10 not selected)
029 (a) per line : 9 12 13 14 15 19 22 23 24
029 (b)          : 9 12 22
```

- Iteration-2 lines 20 ("approvals are optional") and 21 ("at least one approval") miss (a) and hit (b), so a section holding only them FAILs. This confirms the lane's claim.
- Wrong team lines 1, 2, and 17 miss (a) and (b) alike, so a section holding only them FAILs on (a).
- Wrong lines 3, 4, 5, 11, and 16 contain "all … approv" and pass both lexical judges. All of them are in `ac028-mode-lines.txt` and fall under the mandatory reading question at L264 ("전원 승인을 병합 조건으로 두고 그 조건을 낮추거나 빼는 말이 없는가").
- Correct lines 8 and 18 pass. Correct lines 6, 7, and 9 false-fail (N4).
- Base template sections: (a) `0 0`, (b) `b_test_s=1`, mode-lines `test_s=0` (mg:3, dl:33, dl:34, dl:50, dl:51). This matches acceptance.md "0.2.3 검출식으로 다시 재도 (a) 0·0, (b) 0·0".

### D3 — AC-029 (b) false fail and review — **RESOLVED**

On the same fixture:
- Iteration-2 correct lines 23 and 24 no longer hit (b), and the review-gated line 22 hits. This confirms the lane's claim.
- Correct lines 14 and 19 pass. Correct lines 9 (combined sentence) and 12 ("without waiting for review", which L359 anticipates as FAIL-leaning) false-fail (N4).
- Wrong lines 13 and 15 escape the lexical judge but are in the mandatory reading list.

Base: (a) `0 0`, (b) empty (`b_test_s=1`).

### D4 — AC-027 effect verbs — **RESOLVED as specified; residual declared**

Judges (i) and (ii) copied from L727/L730, fixture `e027/ac026-dl.md`:

```
 1 - `--no-merge`: Deprecated no-op; stops auto-merge for this run.                          (wrong)
 2 - `--no-merge`: Deprecated no-op; blocks merging when `--auto-merge` is given.            (wrong)
 3 - `--no-merge`: Deprecated no-op; opts out of auto-merge.                                 (wrong)
 4 - `--no-merge`: Deprecated no-op; auto-merge is disabled by default.                      (correct)
 5 - `--no-merge`: Deprecated no-op kept for compatibility; changes nothing.                 (correct)
 6 - `--no-merge`: Deprecated no-op; takes precedence over `--auto-merge`.                   (wrong)
 7 - `--no-merge`: Deprecated no-op; cancels auto-merge.                                     (wrong)
 8 - `--no-merge`: Deprecated no-op; disables auto-merge for this run.                       (iter-2 auditor line)
 9 - `--no-merge`: Deprecated no-op that overrides `--auto-merge`.                           (iter-2 auditor line)
10 - `--no-merge`: Deprecated no-op; turns off merging even when `--auto-merge` is given.   (iter-2 auditor line)
i_test_s=1    ii_test_s=0    (ii) hits: 4 7 8 9 10
base dl: (i) 8, 19   (ii) 8, 19
```

- All three iteration-2 lines hit, which is what the iteration-2 fix asked for.
- Wrong lines 1, 2, 3, and 6 escape. AC-027 has no reading gate, and this is now declared at spec.md §E.2 and acceptance.md L379 (N5).
- Correct line 4 false-fails (N4).

### D5 — scope-file count — **RESOLVED**

```
$ /usr/bin/grep -n -e '열 개' -e '열 파일' -e '나머지 아홉' -e '위 아홉' -e '뺀 아홉' -e '아홉 파일' -e '여덟' spec.md plan.md acceptance.md progress.md
```

- Zero "열 개" / "열 파일" / "나머지 아홉" / "위 아홉" / "뺀 아홉" / "아홉 파일" hits remain.
- The "여덟" hits are all consistent:
  - spec L90 (8 others besides agent-common-protocol)
  - spec L99 "위 여덟" (the §A.5 list names 8)
  - spec L290 "뺀 여덟" with 8 names
  - plan L24/L30
  - acceptance L237 (AC-005: 9 − manager-git)
  - acceptance L551/L563 (registry: 9 − agent-common-protocol)
  - acceptance L720 "나머지 여덟 조각" (9 fragments − dl)
  - "테스트 파일 여덟 개" at spec L48 and acceptance L19 (unrelated)
  - progress L13

Affected-file figures (sweep `(19|20|21)`): spec L235 "19 … 21", plan L7 "19개(… 21개)", progress L24 "19 … (21)". No stale "21 (… 20 + 1)" remains. AC-013's split (4 diff-exit-0, of which 3 are scope files plus the published copy, and 6 intended-difference files) sums to nine.

### D6 — stale HEAD note — **RESOLVED**

spec.md:L303 now cites `0af445525`. Re-measured at `ba62ff218`, the broader `git diff --stat b412f8a33 HEAD -- …` (tree attribution above) gives no output, so the premise still holds.

### D7 — variable scope convention — **RESOLVED in intent; the prescribed form regresses (N1)**

The sentence is present at acceptance.md:L14. Measured guard behaviour, one plain invocation each:

| Test | Command shape | Result |
|---|---|---|
| T1 | `BASE=…; git log --format=%H $BASE..HEAD -- Makefile` | ran |
| T2b | `BASE=…; E=…; T=…; git log … -- Makefile` (no redirect) | ran |
| T2 ×2 | `E=<abs>; git log … -- Makefile > $E/t2.txt` | **refused** both times ("names git in a form too complex") |
| T7 | `BASE=…; E=<abs>; T=internal/template/templates; git diff --name-only $BASE -- $T/.codex/agents/moai/ > $E/ac014-changed.txt` (AC-014 L436 in the L14 form) | **refused** |
| T8 | same judge with literal values | ran |
| T3 | AC-016 L521 anchor with literal values | ran; output file 0 bytes, `test -s` exit 1 (expected at base) |
| T4 ×3 | `E=<abs>; awk '/^## PR Auto-Merge/…' .claude/agents/moai/manager-git.md > $E/…` (AC-028 L774 shape) | 1st: reported "completed with no output" but **no file was written**; 2nd and 3rd: **refused** |
| L2 | `E=<abs>; /usr/bin/grep -c -e '--auto-merge' .claude/agents/moai/manager-git.md > $E/…` (AC-006 L323 shape) | **refused** |
| L1 | AC-028 L774 extraction with a literal output path | ran; file 284 bytes, 9 lines (matches the base "mg 절 9줄") |
| T6 ×2 | AC-016 L527 after-list (16 paths) with literal values and redirect | 1st **refused**, 2nd ran (0-byte file) |
| V1–V5 | after-list without redirect; single or double `manager-git.md` path with literal redirect; 4 paths without `manager-git`; `manager-[g]it.md` pathspec | all ran |

`/usr/bin/grep -n -E '^git .*> \$E/' acceptance.md` finds 13 git judges that redirect into `$E` (L436, 487, 519, 521, 527, 568, 879, 882, 895, 902, 904, 905, 909). At least 9 more lines read `manager-git.md` into `$E` (L97, 252, 254, 256, 323, 369, 440, 444, 774). The L14 alternative, "값을 글자 그대로 넣는다", is the form that ran.

### D8 — AC-016 template anchor — **RESOLVED**

```
$ git log --format=%h --reverse b412f8a33..HEAD -- .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md
  538684c47 02aca7afe 880c0c702 87988e946 24e20ec50 caa601d7c ea09ca650 3512b9f75 8cf5aaf06   (first line = oldest)
$ git log -1 --reverse --format=%h b412f8a33..HEAD -- …/spec.md   → 8cf5aaf06   (newest — confirms the L525 warning)
$ git log --format=%h 0af445525..HEAD -- .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/   → 8cf5aaf06   (<ACP>..HEAD excludes the anchor)
AC-016 anchor on this tree (literal form): 0-byte file, test -s exit 1 → "판정 불가(PASS 아님)" as L528-L533 says
```

- The anchor is the oldest commit touching either `agent-common-protocol.md` copy (L521 + L524).
- The after-list covers 16 paths: 8 local and 8 template, command source included.
- The L528-L533 ordering mutants hold by the mechanics measured above.
- No fixture commits were created (Gap).

## New findings (all optional; none must-fix)

N1. D7-PREFIX-FORM-REFUSED — acceptance.md:L14 — Severity: minor — Class: optional
- **Problem.** The first form L14 prescribes, "`BASE=…; E=…; T=…; <판정 명령>`", is refused by the session guard for git judges that redirect into `$E`, and for any command reading `manager-git.md` into `$E`. Evidence: T2 ×2, T7, T4 ×2, L2. The literal-value alternative in the same sentence ran (T3, T8, L1, V2–V5).
- **Why optional.** The refusal is loud, and the working fallback is in the same sentence.
- **Fix.** One clause: "git 명령과 `manager-git.md` 를 읽는 명령은 변수 없이 값을 글자 그대로 넣는다(가드가 `E=…; … > $E/…` 형태를 거부한다)".

N2. EXPECTED-EMPTY-WITHOUT-EXISTENCE — acceptance.md:L527-L528 (AC-016 after-list, "기대: 빈 파일") and the same pattern at L879/L895 — Severity: minor — Class: optional
- **Problem.** A judge whose PASS is an empty output file cannot tell "ran and printed nothing" from "never ran".
- **Measured.** One `E=…; awk … > $E/…` invocation reported completion without creating its file (T4 first run), and the identical T6 command was refused once and accepted once.
- **Mitigation already present.** The awk/grep judges downstream of fragment extraction are protected, because each fragment carries its own `test -s`.
- **Fix.** Add `test -e <file>` (or `wc -c <file>` recorded as `0`) before reading an expected-empty git output.

N3. MODE-LINES-SELECTOR-GAP — acceptance.md:L783 — Severity: minor — Class: optional
- **Problem.** The reading list selects lines with approv / review / team mode / personal / manual. A mode-less merge-condition line, such as fixture line 10 "`--auto-merge` merges as soon as CI checks pass.", is not selected. It also passes 028/029 (a)/(b).
- **Why this matters.** The base mg rule itself is mode-less (mg:3 "Execute only with `--auto-merge` flag AND all approvals obtained:"), so this line shape is realistic.
- **Fix.** Add `|| /--auto-merge/` to the L783 selector.

N4. FAIL-CLOSED-FALSE-FAILS — plan.md:L31-L37 (§B.14), acceptance.md:L655/L730/L777-L780/L834 — Severity: minor — Class: optional
- **Problem.** Correct wordings that still fail:
  - 026 line 8 "deprecated; alias of"
  - 027 line 4 "auto-merge is disabled by default"
  - 028 line 6 "a single approval is not enough", line 7 "approval from all reviewers", line 9 (team and personal/manual in one sentence; hits 028 (b) "without an approv" and 029 (b) "after all approv")
  - 029 line 12 "without waiting for review"
- **Why optional.** All of them fail closed, and §B.14 gives accepted example shapes.
- **Fix.** Add one bullet to §B.14: "모드별 조건은 한 줄에 한 모드만 적는다", plus "`disabled by default` 같은 부정·기본값 서술에 효과 낱말을 쓰지 않는다".

N5. AC027-LEXICAL-RESIDUAL — acceptance.md:L730 — Severity: minor — Class: optional
- **Problem.** AC-027 still passes wrong lines 1 ("stops"), 2 ("blocks merging"), 3 ("opts out of"), and 6 ("takes precedence over"). It has no reading gate.
- **Status.** Declared at spec.md §E.2 and acceptance.md L379.
- **Fix, if wanted.** Reuse the AC-026 reading record for `--no-merge` lines, or add `stop|block|opt(s)? out|precedence`.

N6. LT-EVIDENCE-NAME-COLLISION — acceptance.md:L660, L783 with §D.3 L944 — Severity: minor — Class: optional
- **Problem.** The judges run "로컬·템플릿 각각" (L639, L771) but write fixed names (`$E/ac026-merge-lines.txt`, `$E/ac028-mode-lines.txt`). The template run overwrites the local list, while §D.3 requires the reading record to cover lines from both copies.
- **Status.** The convention predates 0.2.3 (all `ac026-*` outputs share it). The new reading-record wording makes the overwrite load-bearing.
- **Fix.** Suffix `-L` / `-T`, or state that the local list is copied before the template run.

## Regression checks

- **Counts and requirements unchanged.** Judged REQ 12 and AC 16. The requirement-layer diff is limited to D5 wording in REQ-005/024. AC table REQ mappings are unchanged (L45-L48).
- **Reading-record preconditions match the AC bodies:**
  - §D.3 L944 lists AC-001/006/026/028/029.
  - AC-026 L154-L156 and L663 ✓; AC-028 L263-L265 and L786 ✓; AC-029 L338 and L837 (shared record) ✓.
  - AC-027 carries no record, stated at L379 and spec §E.2 ✓.
  - plan §E E7 lists four records ✓; plan §F step 4 "읽기 기록 네 개" ✓; plan §G antipattern lists 001·006·026·028·029 ✓.
- **No judged AC passes on an empty swept set:**
  - AC-026: (i) ≥1 per fragment, (iii) ≥1 in four files, `test -s ac026-merge-lines.txt` (base 10 lines).
  - AC-028: `test -s ac028-mg.md`, (a) ≥1, `test -s ac028-mode-lines.txt` (base 5 lines).
  - AC-029: (a) ≥1 each.
  - AC-027: per-fragment `--no-merge` count guard, unchanged.
  - AC-016: `test -s ac016-acp-commits.txt` (base exit 1 → 판정 불가).
  - AC-030: published-flags regex swapped to the directional form. The published file has no `merge`, so it stays empty, and `test -s` on the published file is unchanged.
  - The one weakness is N2, a missing output file.
- **Runnability.** The new judge strings (L655, L658, L660, L730, L777, L780, L783, L834, L521, L524) all ran under the guard in the `E=<abs>; awk|grep … $E/…` form over fragment files, or in literal form for git. The L14 prefix form fails on git and `manager-git.md` commands (N1).
- **Cross-document consistency** of the 0.2.3 wording (spec HISTORY row L32, plan §A L7, progress L6-L16/L24-L26, acceptance claim table L57-L61): consistent. The claim table still names AC-028 (a)·(b) without mentioning the reading record for "team 모드가 승인 없이 병합"; the AC body carries it. Cosmetic; not listed.

## Regression Check (Iteration 3)

Defects from reduced iteration 2:
- D1: [RESOLVED] — iteration-2 lines 17/18/19 hit (ii), and the removal line gives presence 0. Own lexical escapes 1/2/16 are covered by the mandatory reading record.
- D2: [RESOLVED] — iteration-2 lines 20/21 fail (a) and hit (b). Own escapes 3/4/5/11/16 are covered by the reading record.
- D3: [RESOLVED] — lines 23/24 pass, and line 22 hits.
- D4: [RESOLVED as specified] — lines 8/9/10 hit. Residual escapes are declared (N5).
- D5: [RESOLVED] — sweep clean.
- D6: [RESOLVED] — citation updated; premise re-measured empty.
- D7: [RESOLVED in intent] — sentence present. The first prescribed form is refused for git and `manager-git.md` judges (N1).
- D8: [RESOLVED] — anchor mechanics measured.

No defect persisted across all three iterations.

## Recommendation

PASS. Every must-pass criterion holds on the change set, the two iteration-2 blocking defects (D1, D2) are closed, and the aggregate score is 0.92 ≥ 0.80. No must-fix remains.

N1–N6 are optional. The cheapest high-value ones, if the lead wants a final touch before kickoff:
1. N1 — one clause at L14 (literal values for git and `manager-git.md` judges).
2. N3 — add `/--auto-merge/` to the L783 selector.
3. N2 — `test -e` before expected-empty git outputs.

Implementation Kickoff Approval remains a separate, mandatory human gate. This verdict does not substitute for it.

## Gaps (not observed)

- `make`, `go test`, and real run-phase edits were not executed (dispatch prohibition).
- The AC-016 ordering mutants (C1/C2/C3 histories) were not executed with real commits. The range and ordering mechanics were measured on this tree's history instead.
- The cause of the T4 first-run no-op (completion reported, no file) is not established. The two re-runs were refused, and parallel-batch interaction is possible.
- Reading-record quality cannot be judged at plan time. AC-026/028/029 PASS in run-phase depends on a human or agent answering the per-line questions correctly.
- The local-copy fragments of AC-026–029 were not re-extracted in this iteration; only template copies were. The local copies are byte-identical to base per the tree diff, and iteration 2 measured L = T.
- No codex or GLM second opinion was requested (Claude-only verdict).
- The lane's own fixture files (`m026.md`, `m027.md`, `m028.md`, `rm026/`, `pos026/` in the shared scratch) were not used. Their recorded lines were re-created from the report text where iteration-2 lines are verbatim.

## Residual risk

- Lexical judges stay line-level. A wrong implementation phrased outside both the regex vocabulary and the reading-list selector passes, especially on AC-027 (no reading gate) and mode-less merge lines (N3).
- Guard behaviour was non-deterministic for identical commands (T6 refused then accepted; T4 no-op then refused). An evidence file absent because of that can masquerade as an empty PASS where no existence check exists (N2).
- The line citations still assume scope files are byte-identical to `b412f8a33`. That must be re-measured if run-phase absorbs develop first (spec.md §E.2 L303).
