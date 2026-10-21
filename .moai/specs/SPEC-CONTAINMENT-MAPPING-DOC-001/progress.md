# Progress: SPEC-CONTAINMENT-MAPPING-DOC-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true

Basis for the explicit signal: independent plan-audit round 6 PASS 0.96 at audited_sha 5ea162734398a6e4d537e29e79101a99de9a1293, verdict file .moai/reports/t1760/plan-audit-r6.md.

plan_complete_at: 2026-10-16
tier: M (spec.md, plan.md, acceptance.md)

### Plan fix round 2 record (card t1760)

Scope: the two blocking findings N1 and N2 of the round-2 plan audit (verdict FAIL, 0.875), plus optional O1, O2, O3. Dispositions: `plan.md` section J. Round-1 blockers D1 and D2 were not touched.

Tree: HEAD `1b2319571070c4d3b60d270037d2f90f7d7ffa8d`, branch `WT-10-15-release-3`; the working tree was that HEAD plus uncommitted edits confined to this SPEC directory. Tools used: `git`, `grep` (in this shell a function over ugrep; its version string was not observed), `awk`, and the CN-4 awk program extracted from `.claude/agents/moai/plan-auditor.md` lines 477-545. The only project tooling used is the `mcp__moai__spec_audit` call at the end; its judging build is the MCP server `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner), and `git merge-base --is-ancestor ee981a932 HEAD` printed exit 1, so that commit is not an ancestor of this tree's HEAD `1b2319571` and the build is not a lagging ancestor.

Claim 1 (N1): E10a no longer fails on tracked card evidence.

```text
$ git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001'
E10a-current exit=1
$ git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10a-with-reports-excluded exit=0
$ git status --porcelain --untracked-files=all -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10b-new exit=0        (stdout empty)
$ git diff --quiet 46960ac7d8043f1bee3f352688cceba8fb405738 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10a-new vs older base 46960ac7d exit=1
$ git ls-files .moai/reports/t1760
.moai/reports/t1760/plan-audit.md
.moai/reports/t1760/progress.md
```

E10b probe: with an untracked `.moai/docs/zz-e10-probe.md` present, the new E10b printed `?? .moai/docs/zz-e10-probe.md` (exit 0); the file was then removed with `rm` and `git status --porcelain --untracked-files=all` printed nothing (exit 0). This created and deleted one file outside the SPEC directory; no tracked file was touched.

Claim 2 (N2): E08 now counts the date shapes AC-CMM-008 promises. The exact fix-round-2 command is in `acceptance.md` (ledger entry E08); each run below used it with only the file argument changed.

```text
24   NEW on e08_hit.txt (24-line must-count) exit=0
0    NEW on e08_zero.txt (26-line must-stay-zero) exit=1
18   NEW on r2/e08_hit_new.txt (18-line new date forms) exit=0
0    NEW on r2/e08_zero_new.txt (21-line extra must-stay-zero) exit=1
3    NEW on r2/e08_three.txt (october 16 / OCTOBER 16 / 16 October) exit=0
0    NEW on f_pass.md (Korean passing fixture) exit=1
1    NEW on r2/f_mut_dayfirst.md (passing fixture + one 16 October line) exit=0
0    NEW on r2/e08_limits.txt (7 shapes expected uncounted) exit=1
4    OLD (fix-round-1 command) on r2/e08_hit_new.txt exit=0
0    OLD (fix-round-1 command) on r2/e08_three.txt exit=1
2    E08-new on the absent document exit=2   (stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory)
```

The 2 on the absent-document line is the command's exit code, not a count. `e08_zero.txt` holds 26 lines by `wc -l` (the audit text said 25).

Claim 3 (O3): the CN-4 verb now reads the plan. `awk -f cn4.awk plan.md acceptance.md` printed `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`, exit 0, no `CONFLICT:` line. The eight candidates were not each re-read in this round; the audit's earlier hand-read of the same clause family found none ordering work across milestones.

Claim 4: frontmatter and era unchanged in kind. `mcp__moai__spec_audit` (project_root = this worktree, filter SPEC-CONTAINMENT-MAPPING-DOC-001) returned `modern_era_clean: 1`, one INFO `EraAutoDetected` (H-5). ID check: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` printed `PASS`.

Gaps: the final document does not exist, so E01 to E09 were observed red on the absent document and on constructed fixtures only (fixtures live in session scratch, not committed); the post-commit run of E10a and E10b is reported in the hand-off message because a commit cannot cite its own hash here; no `audit_multi` or `spec lint` run was made in this round (the independent re-audit is the next step); the fix-round-1 tracked-file probe for E10a and E10b was not re-run; the 8 CN-4 candidates were not re-read one by one.

Residual risk: E08 counts lines and enumerates date shapes, so spellings outside the list in `acceptance.md` (Known limit) pass unseen; a change under `.moai/reports/` is invisible to E10 by design.

### Plan fix round 3 record (card t1760)

Scope: the two blocking findings N3 and N4 of the round-3 plan audit (verdict FAIL, 0.90; the required codex leg failed with two P2 findings the auditor reproduced), plus a disposition for the two optional findings. Dispositions: `plan.md` section J.1. N1, N2, and the round-1 blockers D1 and D2 were not touched, and the E08 command text was not changed.

Tree: HEAD `04890fc1336a8d4b4843be7208afbdbdb3ee8150`, branch `WT-10-15-release-3`, working tree clean at the start; the edits are confined to this SPEC directory. Tools used: `git`, `grep` (a shell function over ugrep here; its version string was not observed), `sh`/`eval`, `awk` (CN-4 program from `.claude/agents/moai/plan-auditor.md` lines 477-545). The only project tooling is the `mcp__moai__spec_audit` call; its judging build is the MCP server `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner), and `git merge-base --is-ancestor ee981a932 HEAD` printed `is-ancestor exit=1`, so that commit is not an ancestor of this tree's HEAD.

Claim 1 (N3): the scope guard now sees another card's tracked evidence. Real tree, narrowed pathspec:

```text
E10a-narrow exit=0
E10b-narrow exit=0        (stdout empty)
E10a-broad (round-2 form ':!.moai/reports') exit=0
E10a-narrow vs older base 46960ac7d exit=1
E10a-no-reports-exclusion control exit=1
git ls-files .moai/reports | wc -l    -> 1789
```

Throwaway repository in session scratch (base commit `1203114fe5bd16b10fca5abd31150f723a3dcbc8`; tracked `.moai/reports/t9999/verdict.md`, `.moai/reports/t1760/plan-audit.md`, `src/keep.txt`), the real tree untouched:

```text
clean tree, narrow E10a                                   exit=0
delete .moai/reports/t9999/verdict.md: narrow E10a        exit=1
delete .moai/reports/t9999/verdict.md: broad E10a         exit=0   (the round-2 mutant)
delete .moai/reports/t9999/verdict.md: narrow E10b        stdout " D .moai/reports/t9999/verdict.md", exit=0
delete .moai/reports/t9999/verdict.md: broad E10b         stdout empty, exit=0
delete .moai/reports/t1760/plan-audit.md: narrow E10a     exit=0   (own evidence, invisible by design)
add untracked .moai/reports/t9999/new.md: narrow E10b     stdout "?? .moai/reports/t9999/new.md", exit=0
```

The probe files were restored with `git checkout -- <path>` inside the throwaway repository only. Post-edit, on the real tree: E10a exit 0, E10b stdout empty exit 0, and `git status --short` listed only the three SPEC files modified by this round before the progress edit (`acceptance.md`, `plan.md`, `spec.md`).

Claim 2 (N4): AC-CMM-008 now says what E08 matches. The command was run unchanged (taken verbatim from the ledger by a scratch script). Fixtures: the 24-line and 26-line sets were rebuilt from the lists in `acceptance.md` (the earlier scratch copies had been overwritten); the others are the round-2 files.

```text
count=24 exit=0 file=e08_hit24.txt        (24-line must-count)
count=0  exit=1 file=e08_zero26.txt       (26-line must-stay-zero)
count=18 exit=0 file=e08_hit_new.txt      (18 new date lines)
count=0  exit=1 file=e08_zero_new.txt     (21 further must-stay-zero lines)
count=3  exit=0 file=e08_three.txt        (october 16 / OCTOBER 16 / 16 October)
count=2  exit=0 file=e08_may.txt          (8-line May fixture; May 4 and MAY 4 are the counted lines per the audit, not re-derived line by line)
count=0  exit=1 file=e08_mixedmay.txt     (mAy 4, maY 4, 4 MAy, MaY 4, mAY 4, 4 mAy)
count=1  exit=0 file=e08_range.txt        ("sections 10-12 of the guide")
```

Claim 3: `awk -f cn4.awk plan.md acceptance.md` after the edits printed `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`, exit 0, no `CONFLICT:` line. The 8 candidates were listed but not each re-read in this round.

Claim 4: `mcp__moai__spec_audit` (project_root = this worktree, filter this SPEC) returned `modern_era_clean: 1`, `total_specs: 1`, one INFO `EraAutoDetected` (H-5). ID check on `SPEC-CONTAINMENT-MAPPING-DOC-001` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` printed `PASS`.

Gaps: the deliverable document does not exist, so E01 to E09 remain observed red on the absent document and green only on constructed fixtures; the fix-round-1 and fix-round-2 E10 probes carried in the ledger were not re-run; the `.moai/lessons-inbox.jsonl` tool-failure rows were not inspected, though no command was refused in this round; no `audit_multi` or `spec lint` run was made (the independent re-audit is the next step); the post-commit E10a and E10b run is reported in the hand-off message because a commit cannot cite its own hash here; optional finding O5 (sync-phase template-mirror check) was not taken.

Residual risk: E08 still counts lines and enumerates date shapes, so mixed-case May, a `10-12` section range (false red), and the shapes listed under Known limit are outside what the command decides; a change under `.moai/reports/t1760/` is invisible to E10 by design; if the run phase merges `main` into the card branch, E10a (fixed base) would read red at the run-phase tip (inferred, not tested).

### Plan fix round 4 record (card t1760)

Scope: the one blocking finding N5 of the round-4 plan audit (verdict FAIL, 0.92; the required codex leg failed with one P2 finding the auditor reproduced), plus the two optional findings, labelled O-a and O-b in `plan.md` section J.2. Dispositions: `plan.md` section J.2. N3, N4, N1, N2, and the round-1 blockers D1 and D2 were not touched; the E08, E10a, and E10b command texts were not changed (`git diff -U0` of `acceptance.md` has no changed `command:` line for them: three `grep -c -F` counts, each 0).

Tree: HEAD `5a8e8dfd5731edfb767c2e468eb82a5b7c268f1d`, branch `WT-10-15-release-3`, working tree clean at the start; the edits are confined to this SPEC directory. Tools used: `git`, `awk` (CN-4 and traceability programs from the earlier scratch copies `r3_cn4.awk` and `r3_trace.awk`, extracted in round 3 from `.claude/agents/moai/plan-auditor.md`; not re-extracted this round), `cp`, `sed`. The only project tooling is the `mcp__moai__spec_audit` call; its judging build is the MCP server `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner), and `git merge-base --is-ancestor ee981a932 HEAD` printed `is-ancestor exit=1`, so that commit is not an ancestor of this tree's HEAD.

Claim 1 (N5): a new file under another card's report directory is not seen by the scope guard in this repository, and the criterion now says so. Real tree:

```text
$ git check-ignore -v .moai/reports/t9999/new.md
.gitignore:235:.moai/reports/*	.moai/reports/t9999/new.md
check-ignore exit=0
$ git ls-files .moai/reports/t1760 | wc -l          -> 5
$ git ls-files .moai/reports | wc -l                -> 1792
real-tree E10a exit=0
real-tree E10b exit=0        (stdout empty)
```

Throwaway repository in session scratch with a copy of the real `.gitignore` (base commit `bbbb14105fee0500d89ad6d7a7e971b980071e0e`; tracked, force-added `.moai/reports/t9999/verdict.md`, `.moai/reports/t1760/plan-audit.md`, plus `src/keep.txt`), the real tree untouched; the narrowed E10a and E10b were run unchanged:

```text
clean tree                                       E10a exit=0   E10b stdout empty
new untracked t9999/new.md (ignored)             E10a exit=0   E10b stdout empty   (check-ignore in that repo: .gitignore:235:.moai/reports/*  .moai/reports/t9999/new.md, exit 0)
t9999/verdict.md modified (appended line)        E10a exit=1   E10b " M .moai/reports/t9999/verdict.md"
t9999/verdict.md deleted                         E10a exit=1   E10b " D .moai/reports/t9999/verdict.md"
t9999/forced.md created and `git add -f`         E10a exit=1   E10b "A  .moai/reports/t9999/forced.md"
```

The fix-round-3 throwaway repository (base `1203114`) was re-inspected: no `.gitignore` (`test -f .gitignore` exit 1) and `git check-ignore -v .moai/reports/t9999/new.md` exit 1, which is why its `??` line did not transfer. Two probe pairs were issued in the same turn against the same scratch repository, so the first clean-tree line was re-taken on its own after removing the new file (exit 0, stdout empty).

Claim 2 (O-a): the two stale counts are replaced by statements that name their command and tree; the measured values at the tree above are 5 and 1792.

Claim 3: `awk -f r3_cn4.awk plan.md acceptance.md` after the edits printed `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 10 ordering candidates`, exit 0, no `CONFLICT:` line. The candidate count rose from 8 to 10 because the new prose carries ordering words; the ten printed lines were read on screen, and none orders work across milestones except the Definition of Done line for AC-CMM-010 (run-phase tip, before sync), which agrees with M4. The traceability verb with `-v accstate=read` printed `COLLECTED: 11 REQ definitions (acceptance input: read)`, exit 0, no `UNCOVERED:` or `ORPHAN:` line (a first invocation without `-v accstate` printed an empty input state and was discarded).

Claim 4: `mcp__moai__spec_audit` (project_root = this worktree, filter this SPEC) returned `modern_era_clean: 1`, `total_specs: 1`, one INFO `EraAutoDetected` (H-5). ID check on `SPEC-CONTAINMENT-MAPPING-DOC-001` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` printed `PASS`.

Gaps: the deliverable document does not exist, so E01 to E09 were not re-run in this round (their commands and recorded red are unchanged); the fix-round-1 and fix-round-2 E10 probes carried in the ledger were not re-run; only the `.moai/reports/*` ignore rule was probed, so the claim that other ignored paths behave the same is a git-behaviour expectation, not an observation; the `.moai/lessons-inbox.jsonl` tool-failure rows were not inspected, though no command was refused in this round; no `audit_multi` or `spec lint` run was made (the independent re-audit is the next step); optional finding O-b (merge of `main` into the card branch) was not measured; the post-commit E10a and E10b run is reported in the hand-off message because a commit cannot cite its own hash here.

Residual risk: a new local file under `.moai/reports/` (any card's directory) or under any other ignored path passes E10a and E10b by design (Known limit 2 in AC-CMM-010); a change under `.moai/reports/t1760/` is invisible by design; if the run phase merges `main` into the card branch, E10a (fixed base) would read red at the run-phase tip (inferred, not tested, carried in `plan.md` section J.1 and J.2); E08 still counts lines and enumerates date shapes, so the shapes listed under its Known limit are outside what the command decides.

## §E.2 Run-phase Evidence

Card t1760, run phase. Tree: branch `WT-10-15-release-3`, HEAD `1ed32946c` at start, document commit `5e3902e27`. Every command below was run in this run against this tree, from the worktree root. Tools used: `curl`, `grep` (a shell function over ugrep here; its version string was not observed), `awk`, `python3 -I`, `git`. No project-built tool judged anything, so no tool-provenance coordinate (judging build) applies. Scratch directory for downloads: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/39eb45b2-9233-459e-a4d1-94f379f1058c/scratchpad` (machine-local; not a citation target; the deciding output is reproduced here).

### E.2.1 M1 — source re-verification (REQ-CMM-004)

Fetch, one URL per scratch directory, raw bytes (`curl -sS -L -o page.bin -w 'HTTP=%{http_code} size_download=%{size_download} url_effective=%{url_effective}\n' <URL>`):

```text
BP-5 HTTP=200 size_download=165292 url_effective=https://www.anthropic.com/engineering/redesigning-source-control   (exit=0)
BP-6 HTTP=200 size_download=193084 url_effective=https://www.anthropic.com/engineering/how-we-contain-claude          (exit=0)
BP-1 HTTP=200 size_download=144992 url_effective=https://www.anthropic.com/engineering/managed-agents                  (exit=0)
BP-2 HTTP=200 size_download=453103 url_effective=https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5-5   (exit=0)
```

Fragment match, `grep -c -F -- "$fragment" page.bin` per fragment (counts are matching lines, not occurrences):

```text
BP-5 count=1 :: Redesigning source control at Anthropic
BP-5 count=1 :: source control became our biggest development bottleneck
BP-5 count=1 :: around 2,500 code changes merged per day
BP-5 count=1 :: we merged over 30,000 per day
BP-5 count=1 :: A proper merge queue, for example
BP-5 count=1 :: much sense at agent scale
BP-5 count=1 :: global repository lock on the merge queue
BP-5 count=1 :: post-merge and pre-deploy automated testing
BP-5 count=1 :: concurrent changes can conflict
BP-5 count=1 :: Claude can typically fix these conflicts faster than humans
BP-6 count=2 :: How we contain Claude across products
BP-6 count=1 :: users approved roughly 93% of permission prompts
BP-6 count=1 :: Design for containment at the environment layer first, then steer behavior at the model layer
BP-6 count=1 :: The goal is to set a hard boundary on what an agent can reach
BP-1 count=1 :: Scaling Managed Agents: Decoupling the brain from the hands
BP-1 count=1 :: Decoupling the brain from the hands meant the harness no longer lived inside the container
BP-1 count=1 :: each could fail or be replaced independently
BP-1 count=1 :: the tokens are never reachable from the sandbox where Claude
BP-1 count=1 :: nothing in the harness needs to survive a crash
BP-2 count=2 :: Prompting Claude Opus 5.5
BP-2 count=2 :: Treat a text-only end of turn as a report rather than as proof the task is done
BP-2 count=2 :: stop after two or three automatic continuations on the same task
```

Two further context sentences relied on by the rows were matched the same way (each count=1): BP-6 `A tight perimeter also means you can relax oversight`; BP-1 `Auth can be bundled with a resource or held in a vault outside the sandbox`. The BP-5 bottleneck sentence also occurs in the article body, not only in the page description (python `-I` offset scan on page.bin: ten offsets, first body offset 15352). Surrounding text of each quoted fragment was read from the fetched bytes (python `-I` context extraction) before the source-claim cells were paraphrased.

Flags from this run: BP-1 VERIFIED, BP-2 VERIFIED, BP-5 VERIFIED, BP-6 VERIFIED. The register fragments in the document are: BP-1 `each could fail or be replaced independently`, BP-2 `Treat a text-only end of turn as a report rather than as proof the task is done`, BP-5 `post-merge and pre-deploy automated testing`, BP-6 `Design for containment at the environment layer first, then steer behavior at the model layer`; each is in the table above with count 1 or more.

### E.2.2 M3 — row basis (REQ-CMM-007, REQ-CMM-008)

Sections read before each row was written, with the Read tool, by line range, at this tree: `factory-dispatch.md` lines 86-239; `factory-dispatch-mechanics.md` lines 10-30 and 100-128; `factory-dispatch-gates.md` lines 52-72; `auto-semantics.md` lines 113-149 and 206-251; `worktree-integration.md` lines 161-216; `gitflow-lane-protocol.md` lines 51-97; `cross-session-messaging-core.md` lines 44-63; `moai-constitution.md` lines 132-142; `ci-autofix-protocol.md` lines 120-141; `glm-web-tooling.md` lines 46-53; `goal-directive-detail.md` lines 120-131. `main-checkout-branch-guard.md` was read from the full copy the harness attached from this worktree path (a Read call answered that the file was already in context), so that one was not opened by a Read call of mine. The harness also attached primary-checkout copies of several rule files that differ from the worktree copies; every citation follows the worktree copy.

Fit calls, seed (plan.md § D) against final:

| Row | Theme / Source | Seed | Final | Basis |
|---|---|---|---|---|
| 1 | correlated-failure / BP-1 crash | aligned | aligned | queue on disk, evidence read, standing awaken carrier |
| 2 | correlated-failure / BP-1 independent | differs | differs | shared machine, incident record, opt-in slot-lease guard |
| 3 | correlated-failure / BP-2 text-only | aligned | aligned | idle notice is a scheduling hint; scope note on layer difference |
| 11 (candidate) | correlated-failure / BP-2 continuations | not seeded | differs | goal ceiling, runtime block cap, stagnation guard; magnitude differs |
| 4 | isolation-first / BP-6 environment layer | differs | differs | doctrine always loaded; mechanical layer opt-in, fail-open, identity exemption |
| 5 | isolation-first / BP-6 hard boundary | aligned | **differs (flipped)** | read text justifies worktree isolation as write-conflict and branch-state separation, and read-only agents by tool list; no file-system, network, or credential bound was read, so `aligned` would overstate |
| 6 | isolation-first / BP-6 93% | aligned | aligned | gate inventory; order of operations (containment before relaxing oversight) stated as NOT matched |
| 7 | isolation-first / BP-1 tokens | gap | gap | three files read besides the constitution; none describes structural separation |
| 8 | merge-queue-alternative / BP-5 post-merge testing | differs | differs | FIFO window, re-measure before queue, no tests inside window; also read § 4 of the lane protocol: origin CI is the verdict surface and a red batch holds the next, so the contrast is partial (stated in the cell) |
| 9 | merge-queue-alternative / BP-5 conflicts | aligned | aligned | conflicts bullet; speed claim explicitly not mapped |
| 10 | merge-queue-alternative / BP-5 bottleneck | gap | gap | no merge-volume or bottleneck measurement opened |

Seed gap claims still without a located defense (G in AC-CMM-006): **2** of 2. Seed gap claims carried as a non-gap row: **0**. Candidate rows of plan.md § B.4: the bounded-continuation candidate was added as row 11 with its source (BP-2 fragment); the independent-review-evidence candidate was dropped because no source row backs it.

Counts on the final document: 11 mapping rows (4 aligned, 5 differs, 2 gap); 12 distinct cited `.md` paths (`grep -o -E '\.(claude|moai)/[A-Za-z0-9_./-]+\.md' | sort -u | wc -l` printed `12`, an extra measurement outside the battery).

### E.2.3 M2/M4 — humanize pass (plan.md § B.2, native-idiom rule)

The `moai-domain-humanize` skill is installed in this session, so it was loaded and its Korean module (`modules/korean.md`, prose layer A-J) read before the pass. Invariant ledger held fixed: the literal tokens, the four fragments, URLs, flags, paths, section names, Fit values, the two source figures, the 93% statistic, and every scope caveat. Seven surgical edits: four `~에 대한/대해` calque spans (category A), one `~하려는 것이다` frame (category I), the calque-like noun `판정 면`, and one clarity rewording of `측정값을 열어 보지 않았다` that is not a catalogued tell (it removed an ambiguity about what was left unopened; the module's principle 2 would count it as an unmatched edit). Estimated change under 2% of the text (an estimate, not a computed metric); no residual S1 identified. The battery below was re-run after the last wording edit.

### E.2.4 M4 — acceptance battery on the final document

Commands are those in `acceptance.md` § Evidence ledger (E01 to E09), copied verbatim, run with the worktree root as cwd, each followed by `echo "Exxx exit=$?"`. First pass was on the pre-humanize draft; the pass below is the re-run on the final text (document commit `5e3902e27`). Both passes gave the same results.

| Id | stdout | exit | Reads as |
|---|---|---|---|
| E01 | (empty) | 0 | document exists |
| E02 | `1` | 0 | `reference-only` present |
| E03 | (empty) | 0 | four register rows, one per id, fragment of 20+ chars, flag cell last |
| E04a | `4` | 0 | four URL lines |
| E04b | `4` | 0 | all four end with a flag cell (equal to E04a) |
| E05 | (empty) | 0 | header once, Fit tokens valid, every non-gap defense cell has path then section marker, three themes present |
| E06 | `2` | 0 | two gap rows, G = 2 |
| E07 | (empty) | 0 | no bad suffix, no missing path, at least eight distinct paths |
| E08 | `0` | 1 | no identifier, year, date, or `develop` (exit 1 is the conforming grep exit for a zero count) |
| E09 | `0` | 1 | no rule-marker token (exit 1 conforming for a zero count) |

Scope guard, run on this tree after the document commit, before this evidence edit (the two commands exclude this SPEC directory, so the evidence edit does not change them):

```text
E10a: git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports/t1760'
E10a exit=0
E10b: git status --porcelain --untracked-files=all -- . (same three exclusions)
(stdout empty)  E10b exit=0
git merge-base --is-ancestor 629bb51cda73fc70422769988b8c28182c9bd2d5 HEAD  -> exit=0
git diff --name-only 629bb51cda73fc70422769988b8c28182c9bd2d5 HEAD, collapsed -> .moai/docs/multi-agent-containment-mapping.md ; .moai/reports/t1760/* ; four SPEC files (acceptance.md, plan.md, progress.md, spec.md)
ls internal/template/templates/.moai/docs/ | grep -c -i 'containment'  -> 0
```

E10 is re-run after the evidence commit and reported in the hand-off message, because a commit cannot cite its own hash here.

### E.2.5 Reviewer-checked criteria (AC-CMM-011 to 014) — self-check, not a mechanical pass

- AC-CMM-011: each VERIFIED flag has its fetch and fragment-match output in E.2.1 above. The check that a fragment is verbatim against the live page was done by the fixed-string match; a reviewer still owns whether the URL is the page intended. Needs the sync audit's read.
- AC-CMM-012: dormant, all four flags are VERIFIED.
- AC-CMM-013: the volume figures and the post-merge-testing conclusion are attributed to the source in the scale paragraph and in rows 8 and 10; the document states that no adoption is recommended. Self-read only. Needs the sync audit's read.
- AC-CMM-014: each defense cell was written from the section named in it and paraphrases; a reviewer compares the cells with the cited sections. The marker-token check cannot see a paraphrased restatement. Needs the sync audit's read.

### E.2.6 Status transition

The `draft -> in-progress` edit to `spec.md` frontmatter was performed after this evidence was first written, in commit `1fdf0c796` (`status: in-progress`, read back with `git show 1fdf0c796:.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/spec.md`). The sync phase then moved `status` to `completed` in commit `d8f8b1196`, and its `sync_commit_sha` was backfilled in `f328a444d`. `internal/runtime/audit_cache.go` `planArtifactNames` includes `spec.md` and `ComputeHash` reads each file as exact bytes (`os.ReadFile`, line 126 at the tree this evidence was first written on), so any `spec.md` status edit changes the plan-artifact hash that the passed plan audit is bound to; that is inherent to the transition. When the run-phase evidence commit `d2f90c152` was made, `spec.md`, `plan.md`, and `acceptance.md` were still untouched and the status was `draft`.

### E.2.7 Gaps and residual risk

- `grep -o -E` with a `.{0,420}` window around a fragment was refused by ugrep (`exceeds complexity limits`, twelve fragment invocations across the four sources: 4 + 3 + 3 + 2). That was a tool error, not a worktree-guard refusal. Fallback: a short Python script (`python3 -I`) printing the context window from the same raw bytes; it is the source of the context sentences used for paraphrase. The fixed-string counts above are unaffected.
- No worktree-guard refusal was seen in this run. `.moai/lessons-inbox.jsonl` was not inspected.
- The `grep -c` counts are matching lines; the fetched pages are large single-line HTML, so a count of 1 does not mean one occurrence.
- Not opened: this repository's own `branch_guard` setting values and any file measuring merge volume; rows 4, 8, and the two gap rows say so.
- Paraphrase drift in the defense cells is not detectable by any command (E09 only sees marker tokens). E08 counts lines and covers only the date shapes listed in its Known limit.
- A third-party page can change after this run; the flags describe this run's fetch only.
- Row 5 was flipped from the seed's `aligned`; a reviewer may weigh the tool-list restriction of read-only agents (named in the cell) differently.
- If the landing sequence merges `main` into the card branch, E10a (fixed base) would read red at the tip (inferred from the fixed-base design, not tested; carried from plan.md § J.1).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-17T21:58:12+09:00 (commit time of the last run-phase commit, `1fdf0c796`, from `git log -1 --format=%cI 1fdf0c796`)
run_commit_sha: 5e3902e27 (document commit; the evidence commit carrying this section is reported in the hand-off message)
run_status: audit-ready
ac_pass_count: 9 of 9 mechanical criteria (AC-CMM-001 to 009) plus AC-CMM-010 at the tip (E10a exit 0, E10b empty)
ac_fail_count: 0
reviewer_checked_pending: AC-CMM-011, 013, 014 (AC-CMM-012 dormant)
preserve_list_post_run_count: 0 files outside the document, this SPEC directory, and `.moai/reports/t1760/` changed since the pinned base
l44_pre_commit_fetch: not run (no push in this run; the pre-commit HEAD and branch were re-read: `1ed32946c`, `WT-10-15-release-3`)
l44_post_push_fetch: not applicable (no push, by instruction)
new_warnings_or_lints_introduced: not measured (no lint or build run; documentation-only card)
cross_platform_build: not applicable (no Go code changed)
total_run_phase_files: 1 deliverable (`.moai/docs/multi-agent-containment-mapping.md`) plus this progress file
m1_to_mN_commit_strategy: document first (one commit, `5e3902e27`), then this evidence in a second commit
spec_status_transition: draft -> in-progress performed in `1fdf0c796`; in-progress -> completed performed by the sync phase in `d8f8b1196` (see E.2.6)

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-10-17T22:00:56+09:00 (authoring time of this record, from `date +%FT%T%z`; the commit follows it)
sync_commit_sha: d8f8b1196
sync_status: audit-ready
sync_agent: manager-docs
sync_scope: SPEC-directory-only close; no CHANGELOG, README, docs-site, or codemaps change (decision and reasons below)
run_commits_cited: 5e3902e27 (document), d2f90c152 (run-phase evidence, §E.2 and §E.3), 1fdf0c796 (spec.md status draft to in-progress)
changelog_entry_position: none — no entry emitted
b12_self_test_a: not applicable (no CHANGELOG emission); observed anyway: `grep -c -F 'SPEC-CONTAINMENT-MAPPING-DOC-001' CHANGELOG.md` printed 0, and `grep -c -F 't1760' CHANGELOG.md` printed 0, this run, this tree
b12_self_test_b: not applicable (no CHANGELOG entry, so no acceptance-criterion count is stated anywhere in a sync artifact); the AC count in §E.3 is manager-develop's and was not re-derived here
b12_self_test_c: not applicable (no CHANGELOG entry claims a file path); the paths named in this section were checked with `ls` before the commit (see the report)
frontmatter_status_transitions.in-progress→implemented→completed: single sync commit; `spec.md` frontmatter `status: in-progress` to `status: completed` only. The intermediate `implemented` value is not written as a separate commit, per the schema's Status Transition Ownership Matrix (the `completed` transition rides the sync commit). `updated:` was already 2026-10-17, the sync commit date, so it is unchanged. No body edit in spec.md, plan.md, or acceptance.md; plan.md and acceptance.md carry no frontmatter, so no `updated:` refresh is owed.
mx_tag_validation: sync diff is documentation only (spec.md frontmatter line plus this section); zero @MX tag changes, no source file touched
canary_compliance_check: not applicable (this SPEC defines no forward-looking policy that its own sync tests)
commit_plan: one sync commit `docs(SPEC-CONTAINMENT-MAPPING-DOC-001): sync-phase artifacts (card t1760)` carrying the placeholder above, then one backfill commit replacing `pending-backfill` with the real sync-commit SHA (the D3 backfill window of `spec-frontmatter-schema.md`)
backfill: done — the sync commit is d8f8b1196 (full SHA d8f8b11967d30d0213468655b0c8f2d72b52660b, read back with `git rev-parse HEAD` right after it landed) and is now written into `sync_commit_sha` above; this backfill commit follows it and touches only that field and this line

### Sync-artifacts decision (what the sync phase did not produce, and why)

- CHANGELOG.md — omitted. (1) `CHANGELOG.md` opens with Keep a Changelog framing ("All notable changes to this project"), and the deliverable changes no shipped behavior, rule, configuration, code, or template. (2) The document is local-development material that is deliberately not mirrored to `internal/template/templates/` (`plan.md` § E; spec.md § 1.3; `ls internal/template/templates/.moai/docs/ | grep -c -i containment` is checked below), so no user project receives it. (3) A CHANGELOG.md edit would turn the AC-CMM-010 scope guard E10a red when it is re-run at the sync tip: E10a compares the pinned base with the working tree and excludes only the document, this SPEC directory, and `.moai/reports/t1760/`. The Definition of Done in `acceptance.md` says the sync phase "may" add a changelog entry; it does not require one. Precedent is mixed and not decisive: two other local `.moai/docs` files sampled (`stale-card-premise-check.md`, `memory-dir-resolution-doctrine.md`) have zero CHANGELOG mentions, while others have one or two (not checked whether those mentions came from code-bearing SPECs).
- README.md and the docs-site (4 locales) — omitted. The SPEC touches no user-facing documentation and `.claude/agents/moai/manager-docs.md` limits README and docs-site synchronization to what the SPEC dictates; this SPEC dictates none (REQ-CMM-001, REQ-CMM-011).
- Codemaps (`.moai/project/codemaps/`) — omitted. No directory, dependency graph, or module changed; the deliverable is one markdown file under `.moai/docs/`, and a content search for the string `multi-agent-containment` over `*.md`, `*.go`, `*.yaml`, and `*.json` (excluding `.moai/reports`) matched only the four files of this SPEC directory (observed this run, this tree), so no codemap, index, or other document references it.
- Template-First mirror — none, by REQ-CMM-011; `internal/template/templates/` is unchanged by every commit of this card.

### Carried forward (stated, not fixed here)

- §E.2.6 and the `spec_status_transition: not performed` line of §E.3 were true when written and are superseded by commit 1fdf0c796 (status draft to in-progress) and by this sync commit. Those two sections belong to manager-develop and were not edited.
- §E.3 `run_complete_at: pending-backfill` is still a placeholder; backfilling it belongs to the run-phase owner.
- AC-CMM-011, 013, and 014 are reviewer-checked criteria that §E.3 lists as pending; this sync phase did not perform that review (the independent sync audit owns the read).
- Replacing the status `in-progress` with `completed` changes the bytes of `spec.md`, a subject of the plan-artifact hash that the passed plan audit is bound to (§E.2.6). That is inherent to the sync-phase transition and was not worked around.
