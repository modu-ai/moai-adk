# SPEC Review Report: SPEC-GIT-DELIVERY-PROCEDURE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.67** (harmonic mean of the four dimensions; Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation — the dispatch's framing and the lane observations were treated as hypotheses and re-measured below; the author's plan-time measurements in progress.md were not trusted.

## Tree attribution

| Coordinate | Command | Output |
|---|---|---|
| Worktree root | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t622` |
| Branch | `git branch --show-current` | `WT-git-procedure-fixes` |
| Audited HEAD | `git rev-parse HEAD` | `538684c47dde4bfe6208cabf8c8fe48ec9c2b2bc` |
| SPEC base vs HEAD | `git diff --stat b412f8a33 HEAD` | only the 4 SPEC artifacts changed (823 insertions) — every scope file is byte-identical between the SPEC's cited base and the audited HEAD, so line citations were checked against the working tree |
| Lifecycle audit | `mcp__moai__spec_audit(project_root=<worktree>, filter_spec=SPEC-GIT-DELIVERY-PROCEDURE-001)` | `{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[]}` (server build v3.2.0-rc.5, commit 84fa4ece4) |

Scratch mutants live in the session scratchpad (`/private/tmp/claude-501/.../scratchpad/`), not exported; they are reproducible from the commands quoted below.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** — `REQ-GDP-001` … `REQ-GDP-015`, sequential, no gaps/duplicates (spec.md:L123-L176). AC-GDP-001 … 016 sequential (acceptance.md:L19-L34).
- [PASS] **MP-2 GEARS format** (judged on the REQ layer in spec.md only; ACs are Given-When-Then in acceptance.md and were graded under Testability) — every REQ carries a pattern label and matches it: Ubiquitous (001, 002, 004, 007, 009, 013, 014), Unwanted (003, 005, 008, 015), Event-driven (006 "When … 일어날 때", L143), Capability-gate (010-012 "Where OD-1이 … 결정되면", L159/L162/L165). Minor wording notes are in D18.
- [PASS] **MP-3 Frontmatter** — spec.md:L2-L13 carry all 12 canonical fields: `id`, `title`, `version: "0.1.0"` (quoted), `status: draft` (valid enum), `created/updated: 2026-09-10`, `author`, `priority: P1`, `phase: "v3.2.0 target"` (release target, not a stage token), `module`, `lifecycle: spec-anchored`, `tags` (comma string). No rejected aliases. `era`/`related_specs` are extra (see D16).
- [N/A] **MP-4 language neutrality** — SPEC does not cover multi-language tooling; REQ-GDP-015 carries the template-neutrality obligation.
- [PASS] **MP-5 D7** — references extracted with `grep -o -E 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → 8 IDs; statuses via `grep -H -m1 '^status:'`: MERGE-METHOD-CONFIG completed, WORKTREE-BRANCH-GUARD completed, V3R5-LATE-BRANCH completed, WORKTREE-BRANCH-GUARD-OPTIN completed, V3R4-WORKFLOW-SPLIT completed, GITFLOW-DOCTRINE-ALIGN in-progress, WORKTREE-ENTRY-STRATEGY completed. None retired/superseded/archived; all exist.
- [PASS] **MP-6 D8** — `grep -rn 'syscall' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/` → exit 1.
- [PASS] **MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/` → exit 1; no research.md (Tier M). OD-1/OD-2 are surfaced as explicit decision tables, not unresolved markers.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity | REQs are single-reading; AC-GDP-014's exception clause is unreachable (D14); REQ-GDP-001 uses permissive "남을 수 있다" (L124) |
| Completeness | 0.75 | one area sparse | all sections present (HISTORY L21, §A L29, §B L117, §C L180, §D Out of Scope L203 with `### Out of Scope — …` H3 + bullets, §E L232, §F L244); OD tables miss two material facts (D11, D12) |
| Testability | 0.50 | several ACs not binary / defeatable | decision-independent ACs 001, 002, 005, 007, 008, 013, 015 each have a demonstrated mutant, an empty-swept-set path, or a non-functional detector (D1-D10) |
| Traceability | 0.75 | one indirect mapping | every REQ has an AC (§D table L17-L34); AC-GDP-016 traces to "plan §D 제약", not a REQ (D13); REQ-GDP-015's REQ-token/SHA/language-bias clauses have no dedicated AC command (D1) |

Harmonic mean: 4 / (1/0.75 × 3 + 1/0.50) = 4 / 6 = 0.667.

## Adjudication of the lane observations

| Observation | Verdict | Evidence |
|---|---|---|
| AC-GDP-013 selects 3 tests but accepts `--- PASS` ≥ 1 | **Confirmed** | all three functions exist (`internal/template/rule_template_mirror_test.go:137`, `:197`, `internal/template/internal_content_leak_test.go:1535`); the `-run` alternation silently matches fewer if any is renamed, and `grep -c -e '--- PASS'` ≥ 1 would still pass. See D6 |
| AC-GDP-001 ordering check can be satisfied by "git fetch then issue everything in parallel" | **Confirmed** | mutant run below (D5) passes both judge greps |
| spec.md §E says doc-execution.md:36 was read only in the local copy; template carries the same sentence | **Confirmed** | `sed -n '30,38p'` on both copies prints the identical line 36 `This affects auto-merge behavior: worktree contexts default to auto-merge.`; the §E bullet at spec.md:L240 is stale (D15) |

## Defects Found

### Must-fix (blocking)

**D1. AC-GDP-015 SPEC-ID detector cannot match this repository's SPEC IDs — acceptance.md:L355, L365 — Severity: critical — Class: blocking**

The regex `SPEC-[A-Z][A-Z0-9]*-[0-9]{3}` allows exactly one segment between `SPEC-` and the digits. Every ID in the SPEC is multi-segment. The positive control, which acceptance.md says must return "1 이상", returns 0; the run-phase pre-check (plan.md:L29) would therefore halt, and a template leak of a real ID would pass the judge.

```
$ grep -c -E 'SPEC-[A-Z][A-Z0-9]*-[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md
0
specid-exit=1
$ grep -c -E 'SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md
18
$ grep -c -E '20[0-9]{2}-[0-9]{2}-[0-9]{2}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md
4
$ printf '%s\n' 'SPEC-GIT-DELIVERY-PROCEDURE-001' 'SPEC-FOO-001' 'SPEC-MERGE-METHOD-CONFIG-001' | grep -c -E 'SPEC-[A-Z][A-Z0-9]*-[0-9]{3}'
1
```

progress.md:L23 records "SPEC-ID/date detector on spec.md → 4"; the 4 is the date count alone, so the SPEC-ID control was never observed green.
Required fix: use `SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}` in both the control and the judge; correct progress.md:L23 to report the two counts separately. Also either add judge commands for REQ tokens (`REQ-[A-Z]+-[0-9]{3}`) and commit SHAs, or state explicitly that REQ-GDP-015's remaining clauses rest on `TestTemplateNoInternalContentLeak`.

**D2. AC-GDP-002 judge passes on a mutant that keeps fetch and rev-list as separate batch items — acceptance.md:L89-L100 — Severity: critical — Class: blocking**

The standalone check is anchored `^git fetch origin main 2>&1$`, and the joined-line count is file-wide. A trailing comment on the fetch line defeats the anchor, and a joined form placed in prose restores the count to 2. All five judge commands pass while the Pre-Spawn code block still lists fetch and rev-list as independent parallel commands (REQ-GDP-002 violated).

```
$ sed -e 's/^git fetch origin main 2>&1$/git fetch origin main 2>\&1  # step 1/' \
      -e 's/no false positives\./no false positives. Joined form: `git fetch origin main 2>\&1; git rev-list --count --left-right origin\/main...HEAD`./' \
      .claude/rules/moai/core/agent-common-protocol.md > scratchpad/acp-mutant.md
$ grep -c -E 'git fetch origin main 2>&1(;| &&) git rev-list --count --left-right origin/main[.][.][.]HEAD' scratchpad/acp-mutant.md
2
mutant-standalone-exit=1          # ^git fetch origin main 2>&1$ on the extracted Pre-Spawn section
1                                 # moai session list --json --filter-spec= count
matrix-diff-exit=0
$ grep -n -E '^git (fetch|rev-list)' scratchpad/prespawn-mutant.md
7:git fetch origin main 2>&1  # step 1
10:git rev-list --count --left-right origin/main...HEAD
```

Required fix: extract only the fenced ```bash block of the Pre-Spawn section; assert (a) no line matching `^[[:space:]]*git fetch` lacks `git rev-list` on the same line, (b) no line matches `^[[:space:]]*git rev-list` on its own, (c) exactly one joined line inside that block. Keep the file-wide count only as a secondary check.

**D3. AC-GDP-007 swept set is narrower than the prohibition it enforces — acceptance.md:L200, L208; spec.md §A.2 L51-L65 — Severity: major — Class: blocking**

The detector is `git checkout |git switch |git reset --hard`. `AGENTS.md:64-66` and `main-checkout-branch-guard.md:20-24` also forbid `git stash`, mutating `git branch`, and `git merge`/`git rebase` onto the checked-out branch. Lines that instruct exactly that in the scope files never enter the ledger, and an OD-1 rewrite using `git pull --rebase`, `git merge --ff-only`, or `git branch -f` passes AC-GDP-007 with zero `unconditioned-primary-instruction` rows.

```
$ grep -n -E 'git pull|git merge|git stash|git rebase|git branch ' <4 local scope files>
.claude/agents/moai/manager-git.md:96:git checkout main && git pull origin main
.claude/agents/moai/manager-git.md:122:git pull origin main           # verify (no-op if reset succeeded)
.claude/agents/moai/manager-git.md:160:- `git fetch origin` → `git pull origin [branch]`
.claude/skills/moai/workflows/sync/delivery.md:276:4. Merge the branch there: `git merge --no-ff <branch>`
.claude/skills/moai/workflows/sync/delivery.md:323:**github-flow**: `git checkout {main_branch} && git pull origin {main_branch}`
.claude/rules/moai/workflow/spec-workflow.md:59:  git pull origin main   # verify
$ printf '%s\n' 'git fetch origin' 'git pull --rebase origin main' 'git branch -f main origin/main' 'git merge --ff-only origin/main' > scratchpad/ac007-mutant.md
$ grep -n -E 'git checkout |git switch |git reset --hard' scratchpad/ac007-mutant.md
ac007-mutant-exit=1
```

Base count itself was confirmed (`git grep -c -E … b412f8a33 -- <4 template files>` → 8 / 3 / 1 / 2).
Required fix: widen the detector to the AGENTS.md §2 set (add `git pull`, `git merge`, `git rebase`, `git stash`, `git branch -[fDmM]|git branch [^-]`), re-measure the base count, extend §A.2 with the new rows (at least manager-git.md:122 and spec-workflow.md:59), and let the `worktree-internal` class absorb legitimate hits such as delivery.md:276.

**D4. AC-GDP-008 has a vacuous-pass path — acceptance.md:L233-L239 — Severity: major — Class: blocking**

The reference detector keys on the literal `manager-git.md` + backtick + ` § `. A reworded reference produces zero hits, and zero hits routes to AC-GDP-012, which is N/A under OD-1 options 1 and 2 — so REQ-GDP-008 (a MUST, decision-independent) is then checked by nothing. Option 1 is exactly the edit most likely to reword these sentences.

```
$ grep -n -o -E 'manager-git.md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' spec-assembly.md spec-workflow.md   # current tree
spec-assembly.md:340:manager-git.md` § Late-Branch Invocation Pattern
spec-workflow.md:62:manager-git.md` § Late-Branch Invocation Pattern
refs-exit=0
$ printf '%s\n' 'Reference: see the manager-git agent § Late-Branch Invocation Pattern for the procedure.' > scratchpad/ac008-mutant.md
$ grep -n -o -E 'manager-git.md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' scratchpad/ac008-mutant.md
ac008-mutant-exit=1
```

The control (acceptance.md:L225-L227) only proves `grep -v` removed a line; it does not exercise the ref→target lookup.
Required fix: zero refs is PASS only when OD-1 = option 3 and AC-GDP-012 PASSes; otherwise FAIL. Add a second sweep over the four files for any `§ <Title>` or `Late-Branch`/`Late-branch` mention that names a manager-git/spec-workflow procedure, and make the control run the full lookup against the fixture (expected: target count 0 → reported broken).

### Should-fix (optional, but each weakens a MUST-PASS AC)

**D5. AC-GDP-001 greps pass on a sentence that keeps `git fetch` inside the parallel batch — acceptance.md:L59-L62 — Severity: major — Class: optional**

```
$ printf '%s\n' 'Issue these as ONE single-turn multi-Bash batch: `git fetch` then `git status`, `git rev-list --count --left-right`, `gh pr checks --json`, all in parallel.' > scratchpad/ac001-mutant.md
mutant-absent-exit=1              # git fetch.*independent|independent.*git fetch
1:Issue these as ONE single-turn …  # order grep hit
mutant-order-exit=0
$ grep -n -E 'git fetch[^.]*(first|before|then|completes)' scratchpad/sync.md   # base Synchronization section
order-base-exit=1
```

The order detector is red on base (good, but not recorded as a control). The human-read clause is the only thing standing between this mutant and PASS, and it is not binary.
Required fix: widen the absent check to `git fetch.*(independent|parallel|batch)|(independent|parallel|batch).*git fetch` on the same line, and record the order detector's base exit 1 as its control.

**D6. AC-GDP-013 PASS-count threshold admits a run where only one selected test ran — acceptance.md:L316-L319 — Severity: major — Class: optional**

Required fix: `grep -c -E '^--- PASS: (TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak) ' $E/ac013-gotest.txt` expected exactly 3. Also note in plan §B that `agent-common-protocol.md` is, like manager-git.md, excluded from byte-parity (`rule_template_mirror_test.go` comment: `.claude/rules/moai/core/agent-common-protocol.md (17 tokens)` under "REMOVED from the byte-parity allowlist"), so the `diff` in AC-GDP-013 is its only mirror guard.

**D7. AC-GDP-005 absence detector is a single literal — acceptance.md:L171 — Severity: minor — Class: optional**

```
$ printf '%s\n' 'gh pr merge 42 --squash --delete-branch' 'gh pr merge --squash <PR> --delete-branch' > scratchpad/ac005-mutant.md
$ grep -n 'gh pr merge <PR> --squash' scratchpad/ac005-mutant.md
ac005-mutant-exit=1
```

REQ-GDP-005 binds "범위 파일의 어떤 실행 예시", but the AC checks only manager-git.md/.toml.
Required fix: `grep -n -E 'gh pr merge[^|]*--squash'` across all five scope files, with the expected hit set being exactly the manager-git.md:32 default sentence.

**D8. AC-GDP-004 source check has no recorded base control — acceptance.md:L146 — Severity: minor — Class: optional.** `grep -n 'merge_method' .claude/skills/moai/workflows/sync/delivery.md` → exit 1 on the current tree, so the detector is red on base; record that as the control.

**D9. AC-GDP-002 progress record miscounts the matrix rows — progress.md:L18 — Severity: minor — Class: optional.** `grep -c -E '^[|] ' <extracted Pre-Spawn section>` → `8`, not 9 (separator rows start `|-`).

**D10. AC-GDP-011/012 controls read `$E/base-AGENTS.md` and `$E/base-SKILL.md`, which plan.md §C step 2 does not export (it exports scope files only) — acceptance.md:L272, L289 — Severity: minor — Class: optional.** Deferred ACs; add the exports when the decisions land.

**D11. OD-2 option C omits that the key is already scheduled for deprecation — spec.md:L199 — Severity: major — Class: optional (completeness of a decision table)**

```
$ sed -n '2936,2944p' internal/config/testdata/shipped_key_inventory.yaml
- path: "workflow.worktree.auto_merge"
  class: D
  evidence: none
  deprecate_after: "v3.1.0"
```

The table cites this file and line but not its content; the SPEC targets v3.2.0. An operator choosing C would give a first consumer to a key the inventory marks for deprecation after v3.1.0. Required fix: state `class: D`, `deprecate_after: "v3.1.0"` in the option C row's default-impact column (fact only, no recommendation).

**D12. OD-1 options 1 and 3 omit the guard-exemption premise — spec.md:L188, L190 — Severity: major — Class: optional (completeness of a decision table)**

```
$ sed -n '175,215p' .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/spec.md   (excerpt)
### REQ-WBG-011 — manager-git Exemption Boundary (Ubiquitous, design-decision)
... The exemption is identity-based because
`manager-git.md` Phase D is the ONLY place the project authorizes branch-state
changes in the primary checkout.
$ grep -rn 'manager-git' internal/hook --include='*.go' | grep -v _test.go
internal/hook/branch_guard.go:455:	return input.AgentType == "manager-git"
```

Option 1 cites only "§A가 Phase D를 … 예외로 인용한 전제"; option 3, which removes Phase D outright, lists no GUARD-001 impact at all. Both options remove the stated justification for the `manager-git` identity exemption in `internal/hook/branch_guard.go:455`. §D already excludes hook changes, which is fine — but the option table should record that REQ-WBG-011's rationale is withdrawn and the exemption would be left without its stated reason.

**D13. AC-GDP-016 is untraced — acceptance.md:L34 — Severity: minor — Class: optional.** It maps to "plan §D 제약", not a REQ. Either add a REQ (ordering constraint) or label it explicitly as a process check outside the REQ trace.

**D14. AC-GDP-014 exception clause is unreachable — acceptance.md:L342 — Severity: minor — Class: optional.** M1 always edits template `manager-git.md:156` (plan.md:L70) and M2 edits `:114`, so "manager-git.md 템플릿이 전혀 바뀌지 않는 조합" cannot occur under any OD-2 choice. Delete or restate the clause.

### Notes

**D15. Stale Gap — spec.md:L240.** Template `sync/doc-execution.md:36` was read and is identical to the local copy (`diff` of the two files was not separately run, but `sed -n '30,38p'` output is identical). Remove the bullet.

**D16. `era`, `related_specs` are not in the schema's Optional Fields table** (`spec-frontmatter-schema.md:156-167` lists `issue_number`, `depends_on`, `lint.skip`, `bc_id`, `amendment_of`, `tier`). Not an MP-3 failure; progress.md:L13 calls them "optional", which the SSOT does not document.

**D17. spec.md:L98 "저장소 전체 검색"** — outside `.moai/`, `grep -rl 'main_late_branch'` returns three files: local and template `manager-git.md` and the template `.toml`. The claim's substance (not a config key) holds; the local copy is omitted from the sentence.

**D18. GEARS wording.** REQ-GDP-001 ends with a permissive "남을 수 있다" (L124); REQ-GDP-005 and REQ-GDP-008 mix an Unwanted clause with positive obligations in one requirement (L138, L151). Not pattern failures; splitting would sharpen them.

**D19. REQ-GDP-003 scope vs AC-GDP-003 extraction.** REQ binds the 342-348 check block; AC compares the whole 328-362 section — stricter than the REQ, which is acceptable, but the two should name the same unit.

## Facts verified accurate (sampled against the tree)

| Claim | Command | Result |
|---|---|---|
| Unconditional prohibition at `AGENTS.md:64` (root and template) | `sed -n '62,67p'` both | identical "Never change branch state in the primary checkout." |
| Guard rule L16 unconditional; condition only on hook default L85-86 | `sed -n '12,42p;80,90p'` template, `sed -n '16p;86p'` local | confirmed |
| manager-git.md cited lines 32, 42, 88, 96, 110, 114, 119, 121, 125, 127, 129, 137, 148, 156, 166, 170 | `cat -n … sed -n '25,175p'` | all match |
| .toml lines 131, 142, 150, 160 (and 108 hardcoded squash) | `grep -n` on the .toml | confirmed |
| agent-common-protocol.md 292/296/299 Pre-Spawn; 347 Pre-Edit joined form | `cat -n … sed -n '285,365p'` | confirmed; joined-form count 1 |
| spec-workflow.md 49 Frozen, 50, 56, 58, 62, 65 | `sed -n '44,68p'` | confirmed |
| spec-assembly.md 258, 332-340 | `sed -n` | confirmed |
| delivery.md 323-324, 335-338, 343, 355, 348-349, 363; copy differences only at 275/278/479-480 | `sed -n`, `diff` local vs template | confirmed (`dl=1`, three hunks) |
| Four other scope files copy-identical | `diff` each | `exit=0`, `sw=0`, `sa=0`, `acp=0` |
| git-strategy tmpl path; `auto_enabled: false` at 28/61/100; `merge_method: squash` at 23/56/95; comment 82-84; no `.yaml` sibling | `grep -n`, `ls` | confirmed (`No such file or directory`) |
| defaults.go 738-739/751-752/767-768, 867-872, 893-895 | `grep -n`, `cat -n` | confirmed |
| types.go 69-75, 142, 623; validation.go 254-305 | `sed -n`, `grep -n` | confirmed |
| workflow.yaml:58 `auto_merge: false`; no `branch_guard` key | `sed -n`, `grep -n` → `bg-exit=1` | confirmed |
| moai-adk-integration.md:195 "Opt-in; off by default"; dead_config_guard_test.go:72 | `sed -n` | confirmed |
| Template citations CLAUDE.md:75, agent-authoring.md:138, delegation.yaml:74, generic-patterns-guide 195 (T)/207 (L), SKILL.md 124/227, workflows/moai.md 45/233 | `sed -n`, `grep -n` | confirmed |
| Prior-SPEC clauses EXCL-LB-004, REQ-LB-001..006/009, OPTIN quote, REQ-MMC-007/008/009, REQ-GDA-006 | `grep -n` on each spec.md | confirmed |
| Worktree default-merge phrase origin `b7447cb90`, moved by `980ccdc56` | `git log -S 'Auto-merge is now the default for worktree contexts'` | `980ccdc56 … sync.md 분할`, `b7447cb90 … AI Agency v3.2` |
| AGENTS.md sizes 15,277 / 16,936 bytes | `wc -c` | confirmed |
| Mirror test coverage: spec-workflow in `workflowOptMirroredPaths`, spec-assembly in `lateBranchMirroredPaths`, manager-git excluded | `sed -n '40,130p' rule_template_mirror_test.go` | confirmed (plus agent-common-protocol also excluded — D6) |
| Repro evidence files tracked | `git ls-files .moai/reports/t622` | 14 files listed |
| Excluded option "예시 명령만 삭제" | spec-assembly.md:336/340, spec-workflow.md:62, `auto_enabled: false` in all three modes | exclusion reasoning is factually correct and consistent with REQ-GDP-008 |
| Neutrality of §C | `grep -n -E '권고\|권장\|추천\|바람직\|recommend' spec.md` | only L182 "권고하지 않는다"; no option phrased favourably. Option costs are stated for all three (Frozen edit for 1, AGENTS.md budget for 2, feature withdrawal for 3) |
| Tier M ceilings | `spec-workflow.md:146-152` | REQ 15 ≤ 16, AC 16 ≤ 16, counted independently |

## Recommendation (fix route for manager-spec)

1. D1 — replace the SPEC-ID regex in acceptance.md:L355 and L365 with `SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}`; add REQ-token and SHA judge lines (or name the leak test as the sole guard); correct progress.md:L23.
2. D2 — scope AC-GDP-002's standalone and joined checks to the Pre-Spawn fenced code block, tolerant of leading whitespace and trailing comments, asserting no standalone `git fetch` or `git rev-list` line remains.
3. D3 — widen AC-GDP-007's detector to the full AGENTS.md §2 forbidden set, re-measure the base count, add the new rows to spec.md §A.2.
4. D4 — make AC-GDP-008's zero-hit path FAIL unless OD-1 = option 3; add a reworded-reference sweep; make the control exercise the lookup.
5. Optional but recommended in the same pass: D5, D6, D11, D12 (each is a one-line or one-cell change); D7-D10, D13-D19 at discretion.

The re-audit (iteration 2, the Tier M ceiling) will be scoped to D1-D4 plus a regression check of any optional items the author chose to fix.

## Gaps (not observed)

- No `go test`, `make agents-emit-check`, or any make target was run (dispatch prohibition); AC-GDP-013/014 behaviour is judged from the command text and the test/Makefile sources.
- Cross-model second opinion not run: `grep -rn 'audit_model' .moai/config/sections/` returned no key, so no project config requested it; the review is Claude-only.
- `diff` of the two `sync/doc-execution.md` copies was not run as a whole-file comparison; only lines 30-38 were compared.
- The `git log -S` attribution was confirmed for the phrase "Auto-merge is now the default for worktree contexts" (delivery.md:349), not for the trigger wording at delivery.md:337 or doc-execution.md:36.
- No actual fetch/rev-list race and no destructive git command was executed (same limits as the SPEC's own §E).
- Whether `internal/spec/lint.go` accepts `era`/`related_specs` silently was inferred from the clean `spec_audit` result, not from reading the decoder.

## Residual risk

- Even with D1-D4 fixed, REQ-GDP-007's decision-independent verification is a human-classified ledger; a classification error remains possible and is visible only to a reader of `.moai/reports/t622/ac01-classification.md`.
- The run-phase's positive controls re-measure against `$BASE`; if develop moves and the lane merges it before run-phase, cited line numbers in §A/§C will drift (the controls use patterns, the tables use line numbers).
- OD-2 option C's interaction with the dead-config inventory test (`TestDeadConfig…` / `shipped_key_inventory.yaml`) remains unverified by the SPEC's own admission (§E L237).
