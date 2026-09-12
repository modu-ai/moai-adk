# SPEC Review Report: SPEC-GIT-DELIVERY-PROCEDURE-001 (revision 0.1.1)

Iteration: 2/2 (Tier M ceiling reached)
Verdict: **FAIL**
Overall Score: **0.75** (harmonic mean; Tier M PASS threshold 0.80; iteration 1 was 0.67, so no score regression and no STOP signal)

Reasoning context ignored per M1 Context Isolation. The author's recorded re-runs in acceptance.md and progress.md were not relied on; every result below comes from commands run in this audit.

## Tree attribution

| Coordinate | Command | Output |
|---|---|---|
| Root | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t622` |
| Branch | `git branch --show-current` | `WT-git-procedure-fixes` |
| HEAD | `git log --format='%h %s' -3` | `02aca7afe feat(SPEC-GIT-DELIVERY-PROCEDURE-001): revise plan artifacts for plan-audit iteration 1` / `647460835 docs(t622): plan-audit iteration 1 verdict …` / `538684c47 …` |
| Full SHA | `git rev-parse HEAD` | `02aca7afe44332d2b57542205bdff62ef0342dc4` |
| Clean tree | `git status --porcelain` | (empty) |
| Change set | `git diff --stat 647460835 HEAD` | 4 files: acceptance.md 311, plan.md 32, progress.md 33, spec.md 67 (283 insertions, 160 deletions) — only the SPEC directory |
| Scope files vs base | `git diff --stat b412f8a33 HEAD -- <5 local scope files> internal/template/templates/.claude AGENTS.md internal/template/templates/AGENTS.md internal/template/templates/.codex` | (empty) — base citations still resolve against the working tree |
| Lifecycle audit | `mcp__moai__spec_audit(project_root=<worktree>, filter_spec=…)` | `{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[]}` |

Mutant fixtures were built in the session scratchpad (`/private/tmp/claude-501/…/scratchpad/i2-*`) and not exported; every one is reproducible from the command quoted with it. The worktree guard rejects the literal word `git` inside fixture text, so those lines were written with `printf '%b' '\0147it …'`. `cat -n` of the fixture shows the rendered `git`.

## Must-Pass Results

- [PASS] **MP-1** — `/usr/bin/grep -o -E '^\*\*REQ-GDP-[0-9]{3}\*\*' spec.md` → `REQ-GDP-001 … REQ-GDP-015`: 15 entries, sequential. `/usr/bin/grep -c -E '^### AC-GDP-[0-9]{3}' acceptance.md` → `16`.
- [PASS] **MP-2** (requirement layer only) — the REQ edits keep their pattern labels. REQ-GDP-001 lost its permissive clause and now says "…이 요구사항이 제한하지 않는다". REQ-GDP-002 is Ubiquitous. REQ-GDP-003 and REQ-GDP-005 are Unwanted.
- [PASS] **MP-3** — only `version: "0.1.0"` → `"0.1.1"` changed in frontmatter. All 12 fields are still present.
- [N/A] **MP-4** — unchanged from iteration 1.
- [PASS] **MP-5 D7** — no new SPEC references were added. The referenced set and statuses were verified in iteration 1 (none retired).
- [PASS] **MP-6 D8** — no `syscall` introduced.
- [PASS] **MP-7** — no `[NEEDS CLARIFICATION` markers; no research.md.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor | REQ-GDP-005/008 still mix an Unwanted clause with positive duties (accepted, see D18). The widened AC-GDP-001 absent detector also reds a correct sentence (N5), which pushes authors toward awkward line splits |
| Completeness | 1.00 | full | OD tables now carry the D11 (`class: D`, `deprecate_after: "v3.1.0"`) and D12 (REQ-WBG-011 rationale, `branch_guard.go:455`) facts across all three OD-1 options. §A.2 is extended, §E is updated, and the export list is complete |
| Testability | 0.50 | several ACs defeatable | new mutants pass AC-GDP-015 (SHA), AC-GDP-008, AC-GDP-007, AC-GDP-002 and AC-GDP-001 (N1-N5) |
| Traceability | 1.00 | full | AC-GDP-016 is explicitly labelled a process check outside the REQ trace (acceptance.md:L37-L39). Every REQ-GDP-015 clause now has a judge line or a stated reading step (L437, L453-L465) |

Harmonic mean: 4 / (1/0.75 + 1/1 + 1/0.5 + 1/1) = 4 / 5.333 = **0.75**.

## Regression check — iteration-1 defects

| Defect | Status | Evidence (command → verbatim output) |
|---|---|---|
| **D1** AC-GDP-015 single-segment SPEC-ID regex (blocking) | **PARTIALLY RESOLVED — SHA clause introduces N1** | SPEC-ID control: `/usr/bin/grep -c -E 'SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' spec.md` → `18`. REQ `27`, date `5`, SHA-line `7` (all match progress.md). Judges on a mutant diff: specid → `2:+see SPEC-MERGE-METHOD-CONFIG-001 here`; req → `3:+per REQ-GDP-001 rule`; date → `4:+landed 2026-09-10 today`; local → `11:+see CLAUDE.local.md`; header `+++` line correctly excluded. The SHA judge is defective — see N1 |
| **D2** AC-GDP-002 mutant (blocking) | **RESOLVED for the iteration-1 mutant**; residual N4 (optional) | iteration-1 mutant rebuilt from the current file: `m1 a-lines: git fetch origin main 2>&1  # step 1` · `m1 b: 5:git rev-list --count --left-right origin/main...HEAD` `b-exit=0` · `m1 c: 0` → FAIL on all three checks, as required. Fixed fixture: `fixed a-lines:` (none) · `b-exit=1` · `fixed c: 1` → PASS. New reversed-order background mutant (`rev-list … & git fetch …`): `m3 b: 2:git rev-list … & git fetch origin main 2>&1` `b-exit=0`, `c: 0` → caught |
| **D3** AC-GDP-007 narrow swept set (blocking) | **RESOLVED for the iteration-1 mutants and base count**; **new blocking gap N3** | Base re-measure `git grep -n -E '<widened>' b412f8a33 -- <4 template files>` → 20 lines: manager-git 42·96·110·119·121·122·125·127·137·160, spec-workflow 50·56·58·59·62, spec-assembly 336, delivery 276·323·324·328, exactly as progress.md and acceptance.md:L259-L260 state. The new §A.2 rows (122, 160, 59, 328) plus the two non-instruction rows (62 narrative, 276 worktree-internal) account for all 20. Iteration-1 mutants: `1:git pull --rebase origin main`, `2:git branch -f main origin/main`, `3:git merge --ff-only origin/main` → all hit |
| **D4** AC-GDP-008 vacuous path (blocking) | **PARTIALLY RESOLVED — new blocking gap N2** | Current tree strict = loose = `1 / 1 / 0`. Lookup pipeline: headings `### Late-Branch Invocation Pattern`, heading-count `1`, lookup `1`. The iteration-1 mutant is caught (strict 0 vs loose 1), and the all-zero route is now conditioned on option 3. A per-file vacuous route remains — see N2 |
| D5 AC-GDP-001 "then … parallel" mutant | Resolved for that mutant; residual N5 | `--- mutants absent` → `1:Issue these as ONE single-turn multi-Bash batch: … all in paral…el.` (hit → FAIL, as intended). Base section: absent `1`, order `0` (red on base, control recorded at acceptance.md:L58-L59) |
| D6 AC-GDP-013 PASS count | Resolved | acceptance.md:L403-L404 `^--- PASS: (A\|B\|C) ` expected exactly 3; the three functions exist (iteration 1: `rule_template_mirror_test.go:137`, `:197`, `internal_content_leak_test.go:1535`) |
| D7 AC-GDP-005 literal detector | Resolved | `gh pr merge[^|]*--squash` → manager-git `32`, `114`; delivery `343`, `355`; toml `26`, `108`; spec-workflow/spec-assembly/agent-common-protocol `0` — matches the 6-line control at acceptance.md:L209 |
| D8 AC-GDP-004 source control | Resolved | `/usr/bin/grep -n 'merge_method' delivery.md` → `exit=1`; recorded at acceptance.md:L179-L180 |
| D9 matrix row count | Resolved | progress.md now says 8, matching iteration-1 measurement `8` |
| D10 export list | Resolved | plan.md §C step 2 lists `base-AGENTS.md`, `base-AGENTS-template.md`, `base-SKILL.md`, `base-delivery-template.md`, `manager-git.toml` |
| D11 OD-2 C deprecation fact | Resolved, neutral | spec.md §A.5 row and §C.2 option C cell cite `shipped_key_inventory.yaml:2940-2943` `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` against target `v3.2.0` — line range matches iteration-1 `sed -n '2936,2944p'` output. Stated as fact without judgement; §E adds the unverified follow-on as a Gap |
| D12 REQ-WBG-011 premise | Resolved, neutral | §A.6 row quotes GUARD-001 spec.md line 210 (iteration-1 `grep -n` → `210:`manager-git.md` Phase D is the ONLY place …`). Options 1 and 3 each state that the rationale disappears and `branch_guard.go:455` stays without it. Option 2 states the rationale survives because Phase D remains — correct, since option 2 keeps Phase D. §D Out of Scope records the hook gap as a post-decision item |
| D13 AC-GDP-016 untraced | Resolved | acceptance.md:L37 REQ column `없음 — 절차 점검`, explanation L39 |
| D14 AC-GDP-014 unreachable clause | Resolved | L427 replaces it with "M1 … `:156`, M2 … `:114` 를 항상 고치므로 첫 점검 RED는 어느 결정 조합에서도 나온다" — true per plan.md M1/M2 |
| D15 stale §E gap | Resolved | diff shows the doc-execution bullet removed; §D sibling bullet now says "(로컬·템플릿 같은 문장)" |
| D16 `era`/`related_specs` | Resolved | progress.md now names them as undocumented-but-precedented; precedent verified: `sed -n '1,20p' SPEC-WORKTREE-BRANCH-GUARD-001/spec.md \| grep` → `13:era: V3R6`, `14:tier: M`, `16:related_specs:` |
| D17 `main_late_branch` search wording | Resolved | spec.md §A.5 now reads "`.moai/` 밖 검색에서 로컬·템플릿 `manager-git.md:137` 과 … `.toml:131`", matching iteration-1 `grep -rl` output (3 files) |
| D18 GEARS mixing | Partial, accepted | REQ-GDP-005 and REQ-GDP-008 are still mixed. The ceiling reason is arithmetically true: 15 REQs + 2 splits = 17 > Tier M ceiling 16 (`spec-workflow.md:146-152`). REQ-GDP-001's permissive clause is fixed. Remains a note, not a defect |
| D19 REQ-003 vs AC-003 unit | Resolved | REQ-GDP-003 now binds "328행 제목부터 `#### The sweep prohibition` 소절 앞까지", the same unit AC-GDP-003 extracts |

Lane facts re-measured: widened detector 10/5/1/4 = 20 (confirmed above). REQ 15 and AC 16 (confirmed). Cf characters: `perl -CSD -ne '$n++ while /\p{Cf}/g …'` → `spec Cf=0`, `plan Cf=0`, `acc Cf=0`, `prog Cf=0`, planted control `control Cf=1` (confirmed). Only the four SPEC files changed (confirmed).

## New findings

### Blocking

**N1. AC-GDP-015 SHA detector misses a large class of real short SHAs, and its stated limitation is wrong — acceptance.md:L445, L461, L467 — Severity: critical — Class: blocking (residue of D1)**

The pattern `(^|[^0-9A-Za-z])[0-9]*[a-f][0-9a-f]{6,39}([^0-9A-Za-z]|$)` needs at least six hex characters AFTER the first a-f letter. A 9-character short SHA whose first letter falls in its last six positions does not match. acceptance.md:L467 says only all-digit SHAs are missed.

```
$ printf '%s\n' '+++ b/…' … '+sha 538684c47 here' '+sha 7374b183e here' '+sha 2213871af here' '+sha 02aca7afe here' '+sha b412f8a33 here' '+sha 647460835 here' … > scratchpad/i2-ac015-mutant.diff
$ /usr/bin/grep -n -E '^[+][^+].*(^|[^0-9A-Za-z])[0-9]*[a-f][0-9a-f]{6,39}([^0-9A-Za-z]|$)' scratchpad/i2-ac015-mutant.diff
8:+sha 02aca7afe here
9:+sha b412f8a33 here
```

`538684c47`, `7374b183e`, and `2213871af` are all commits named in this worktree's own log and conversation, and all contain letters, yet the judge misses them. The positive control on spec.md hides this, because the missed SHA `980ccdc56` shares line 123 with the matched `b7447cb90`:

```
$ /usr/bin/grep -n -o -E '(^|[^0-9A-Za-z])[0-9]*[a-f][0-9a-f]{6,39}([^0-9A-Za-z]|$)' spec.md   → … 123:`b7447cb90` …   (no 980ccdc56)
$ /usr/bin/grep -n -o -E '[0-9a-f]{9,40}' spec.md                                               → … 123:b7447cb90  123:980ccdc56 …
```

Required fix: match `(^|[^0-9A-Za-z])[0-9a-f]{7,40}([^0-9A-Za-z]|$)` on added lines. Remove all-digit tokens with a second filter or a reading step, not in the regex. Rewrite L467 to state what is actually excluded. Make the control assert that `980ccdc56` is hit (for example `grep -o` count of hex tokens = 9 on spec.md), not only a line count.

**N2. AC-GDP-008 still passes when one file's reference is reworded below both detectors — acceptance.md:L285-L310 — Severity: major — Class: blocking (residue of D4; contradicts the AC's own summary "문구가 바뀐 참조도 놓치지 않음", L29)**

Equality is checked per file, and the all-zero route is guarded only on the total. A file whose only reference is rewritten in lowercase, or as prose without `§ <Title>` or "Invocation Pattern", scores 0 = 0 while the other file keeps the total non-zero.

```
$ sed -e 's/… § Late-Branch Invocation Pattern\./… § late-branch invocation pattern./' spec-workflow.md > scratchpad/i2-sw-lower.md
lower strict:  spec-assembly.md:1  i2-sw-lower.md:0  delivery.md:0
lower loose:   spec-assembly.md:1  i2-sw-lower.md:0  delivery.md:0          → diff exit 0 → PASS
lower line62: .claude/agents/moai/manager-git.md` § late-branch invocation pattern.
$ sed -e 's/see `…manager-git.md` § Late-Branch Invocation Pattern\./see the git agent definition, Late-branch section./' spec-workflow.md > scratchpad/i2-sw-reword.md
reword strict: spec-assembly.md:1  i2-sw-reword.md:0  delivery.md:0
reword loose:  spec-assembly.md:1  i2-sw-reword.md:0  delivery.md:0         → PASS
```

Both fixtures leave spec-workflow.md pointing at a procedure whose existence nothing checks. The iteration-1 required fix asked the second sweep to cover "`Late-Branch`/`Late-branch` mention"; the implemented loose detector does not.
Required fix (smallest): add a per-file floor. Unless OD-1 = option 3 with AC-GDP-012 PASS, the strict count in spec-assembly.md and spec-workflow.md must each be ≥ its base value (1). Optionally make the loose detector case-insensitive on the `§` form (`-i 'manager-[g]it[^§]*§'`).

**N3. AC-GDP-007 detector misses long-option forms of the "mutating branch" class it declares, and `git -C` forms — acceptance.md:L249-L253 — Severity: major — Class: blocking (contradicts the AC's WHEN clause "AGENTS.md §2 금지 집합 전체를 잡는 검출식" and plan.md §G anti-pattern "금지 집합의 일부만 검출식에 넣는 것")**

```
$ cat -n scratchpad/i2-ac007-mutants.md
     1	git pull --rebase origin main
     2	git branch -f main origin/main
     3	git merge --ff-only origin/main
     4	git branch --force main origin/main
     5	git branch --move feat/x
     6	git branch --delete feat/x
     7	git -C . switch main
     8	git reset --keep origin/main
$ /usr/bin/grep -n -E '[g]it (checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it branch (-[a-zA-Z]*[fDmMcCdtu]|[^-])' scratchpad/i2-ac007-mutants.md
1:git pull --rebase origin main
2:git branch -f main origin/main
3:git merge --ff-only origin/main
exit=0
```

Lines 4-6 (`--force`, `--move`, `--delete`) are mutating `git branch` forms inside the declared set. Line 7 (`git -C <path> switch`) is the idiom this repository's own rules prescribe, so a rewrite under OD-1 option 1 is likely to use it. Line 8 (`reset --keep`) is outside the AGENTS.md literal list and is noted only.
Required fix: add `--(force|move|copy|delete|set-upstream-to|unset-upstream)` to the branch alternation, and allow an optional `-C <path> ` between `git` and the subcommand (e.g. `[g]it (-C [^ ]+ )?(checkout|switch|…)`). Re-measure the base count; the current base has no `git -C` hits in the four files, so 20 should hold.

### Optional (should-fix)

**N4. AC-GDP-002 passes with a second, unordered rev-list — acceptance.md:L113-L121 — Severity: minor — Class: optional**

```
$ cat -n scratchpad/i2-m2-block.md
     2	git fetch origin main 2>&1; git rev-list --count --left-right origin/main...HEAD
     5	git -C . rev-list --count --left-right origin/main...HEAD
m2 a-lines: (none)   m2 b: b-exit=1   m2 c: 1          → PASS
```

Fix: add `/usr/bin/grep -c 'rev-list' <block>` expected exactly 1.

**N5. AC-GDP-001 absent detector is both evadable and over-strict — acceptance.md:L66-L69 — Severity: minor — Class: optional**

A multi-line list evades it: lines 2-5 of `scratchpad/i2-ac001-mutants.md` put `` `git fetch` (first) `` on its own list item under "Issue the following as ONE single-turn multi-Bash call". The absent check has no hit on those lines, and the order check hits `3:- `git fetch` (first)`. A correct single sentence is also red:

```
$ /usr/bin/grep -n -E 'fetch.*(independent|para[l]lel|batch)|…' scratchpad/i2-ac001-correct.md
1:Run `git fetch` first; once it completes, issue the independent reads (`git rev-list --count --left-right`, `git status`) as one batch.
exit=0
```

The judge therefore forces authors to split lines, which is exactly the shape that evades it. Fix: extract the Synchronization section, then judge on the paragraph or bullet group rather than the line. At minimum, record that the human-read order check must confirm no list or batch groups `git fetch` with `git rev-list`.

**N6. AC-GDP-008 has no runnable control command — acceptance.md:L314-L320 — Severity: minor — Class: optional.** The section records control results but provides no command block, while plan.md §C step 3 tells run-phase to execute each AC's "대조" commands. Add the fixture command (heading removed → lookup 0).

## Recommendation

The three blocking items are each a one-line regex or one extra count:
1. N1 — replace the SHA pattern with `[0-9a-f]{7,40}` bounded by non-alphanumerics, filter all-digit tokens separately, correct L467, strengthen the control.
2. N2 — per-file floor for strict reference counts (spec-assembly.md ≥ 1, spec-workflow.md ≥ 1) unless OD-1 = 3 with AC-GDP-012 PASS.
3. N3 — add long `git branch` options and an optional `-C <path>` to the AC-GDP-007 detector, then re-measure the base count.

Iteration 2 is the Tier M ceiling. Per the retry-loop contract the orchestrator escalates to the user with three options: PASS-with-debt, scope reduction, or an explicit override to iterate again. The OD tables, requirements, and traceability are in good shape. The remaining debt sits entirely in the verification layer (acceptance.md detector strings) and does not affect REQ content or the decision tables.

## Gaps (not observed)

- No `go test`, `make`, or `agents-emit` target was run (dispatch prohibition). The AC-GDP-013 top-level `--- PASS:` line format is taken from Go's standard `-v` output, not observed in this run.
- AC-GDP-009 to 012 (deferred) judge commands were not exercised beyond the recorded controls.
- Fixtures under the session scratchpad were not exported; they are reproducible from the quoted commands.
- The `printf` in the first N1 fixture command wrote literal lines, so no escape interpretation was involved. The N3 fixture used `%b` octal escapes, and the rendered text was checked with `cat -n`.
- No cross-model audit (no `audit_model` key in project config — iteration 1 measurement, not repeated).

## Residual risk

- Even with N1-N3 fixed, AC-GDP-001 and AC-GDP-007 finish with human reading (order sentence, ledger classification); a misreading remains possible.
- The detectors are line-based; any requirement violation spread across lines, or phrased in a way no pattern anticipates, can still pass. The per-file floor in N2 and the rev-list count in N4 narrow but do not remove this.
- Base line citations rely on scope files staying byte-identical to `b412f8a33`. The lane memory notes develop has moved (t610 at `296ba7aa5`); absorbing it before run-phase requires re-checking that the scope files are unchanged.
