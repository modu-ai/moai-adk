# t2070 progress (lane-19, run to0ck1, Class B)

Operator asks (via leader): `moai update`'s worktree migration looks stuck and is very slow; announce first that it can take long; show progress in the plan and the move; and make it faster (parallel migration considered). Safety checks are not reduced.

## Setup
- `moai factory next --card t2070 --run to0ck1` -> `t2070 stage=- worktree=t2070`, tree .../worktrees/projects/moai-adk-go/t2070, branch WT-moai-update-mo.
- Base was local main 69b10035f3; t2064 (c7c8cc05c2) merged first by the leader's instruction (`git merge --no-ff`), so the fixture stubs the lsof probe.
- Everything below was measured on an isolated fixture (temp repo, temp MOAI_HOME, `m6NewWorld` + N extra legacy trees). Nothing was written to mo.ai.kr or any user tree. The machine was also running the leader's regression, so absolute numbers carry load noise (lsof took 1.7-3.8s here against 1.1s the operator saw).

## 1. Measurement before any change (probe: `probe-stage-timing.go.txt`)

Plan, per call, summed over the registered trees (N=120 extra, 129 registered, apply plan):

| stage | measured | per tree |
|---|---|---|
| `processUsesWorktree` against 1,513 process cwds (69 trees, real lsof list) | 13.88 s | **201 ms** |
| `worktreeroot.PathFor` (one `git rev-parse` each) | 7.29 s | **56 ms** |
| `Classify` | 256 ms | 2 ms |
| `resolveTreePath` | 59 ms | 0.5 ms |
| session registry probe | 8.8 ms | 0.07 ms |
| live-anchored probe | 7.0 ms | 0.05 ms |
| `os.Lstat` | 0.9 ms | ~0 |
| `git worktree list` (once) | 442 ms | - |
| `loadMoveList` (once) | 9 us | - |
| lsof scan (once) | 1.7-3.8 s | - |

One tree's move (N=129 / N=69 / N=69 second run):

| step | 129 | 69 | 69 (2nd) |
|---|---|---|---|
| manifest before | 442 ms | 242 ms | 514 ms |
| instance token | 0.7 ms | 1.6 ms | 0.4 ms |
| `git worktree move` | 536 ms | 97 ms | 150 ms |
| re-key | 4 ms | 0.6 ms | 0.7 ms |
| card rebind | 72 ms | 82 ms | 49 ms |
| manifest after | 353 ms | 228 ms | 258 ms |

Manifest (before + after) is 55-70% of one tree's move; each capture was eight git reads one after the other.

`git worktree move` of 8 trees, same repository, serial vs concurrent: 724 ms vs 300 ms (first run), 470 ms vs 233 ms (second run); 0 failed and every destination listed afterwards.

Cause of the operator's two stalls:
1. The ~2 minutes of silence after "Leftover state checked": the plan checked every tree against every process cwd with symlink resolution repeated for each pair (201 ms/tree here; 648 trees -> the observed two minutes), plus one `git rev-parse` per tree for the destination, and nothing was printed until the plan returned.
2. The ~10 s per tree: manifest captured twice with eight sequential git subprocesses each, plus the move itself (the mo.ai.kr trees are much larger than the fixture's, so git reads dominate more there; not measured here).

## 2. Decisions (a) and (b)

- (a) Cost per tree - adopted. Home root resolved once per plan (not per tree); process cwd index built once (symlink/ancestor work per process directory, not per pair); the eight manifest reads of one capture run together and are settled in the old order.
- (b) Parallel - adopted only for the read-only part: the before-the-move manifest of up to 3 trees ahead is read while the current tree moves (`config.DefaultWorktreeMigrateManifestLookahead = 3`). The move, the journal `pending` write, the move-list lines, the re-key/rebind and the after-the-move comparison stay one tree at a time, in plan order. Reason: concurrent `git worktree move` looked safe in two small samples, but two samples do not prove git's registration locking, and the journal and move-list order are what `--rollback` recovers from; the gain left on the table is the move call itself (60-150 ms/tree in the fixture). Reported below as an option, not implemented.
- Safety checks unchanged: before/after manifest comparison, journal `pending` before the move, the reverse order, fail-safe skips (a probe error is still a skip).

## 3. RED -> GREEN (file `internal/cli/worktree/migrate_speed_t2070_test.go`)

Seams were added first with the old behaviour, so the tests failed for the intended reason:
- `TestPlanMigration_ResolvesTheDestinationRootOnce`, `..._DryRun...`: "resolved 6 times for 6 trees; want exactly 1" -> GREEN.
- `TestCaptureTreeManifest_ReadsRunTogether`: peak concurrency 1, 835 ms for eight 80 ms reads -> GREEN.
- `TestProcessCWDIndex_ManyTreesAgainstManyProcessesIsQuick`: 120 trees x 901 cwds took 15.85 s -> GREEN 0.35 s.
- `TestTreeProgress_*` (count + "can take a long time", throttled scan, elapsed/remaining) and `TestMigrationEngine_AnnouncesBeforeThePlanRuns` (notice already on screen when the plan first resolves a destination, before the "to move" line) -> GREEN.
- `TestMigrationEngine_ReadsManifestsAheadButBoundedAndMovesInPlanOrder`: peak 1 -> GREEN (peak between 2 and 3; rows and move-list in plan order).
- Characterization tests that pass before and after: `TestCaptureTreeManifest_EqualsTheSequentialCapture` (digest identical to the sequential algorithm), `..._FailureReadsTheSameAsTheSequentialCapture`, `..._TimeoutStillFailsTheCapture`, `TestProcessCWDIndex_AnswersLikeTheDirectComparison`, `TestMigrationEngine_AManifestReadAheadStillMustMatchAfterTheMove` (a stale read-ahead digest still ends `reversed`, tree back at source), `TestClassify_AgentNamedTreeFollowsItsRoot`.

Mutants (each reverted, `git diff --stat` showed only the intended files afterwards):
- ancestor comparison of the process index disabled: first SURVIVED (no respelled-path case) -> test case added (same directory spelled with another case; macOS) -> then RED (`index false, direct comparison true`).
- read-ahead bound removed (slots 1000): RED (`peak of 10 manifests in flight exceeds the lookahead bound 3`).

## 4. Before / after, same fixture (probe: `probe-end-to-end.go.txt`, real lsof, 100 extra trees)

| | before (HEAD of the base) | after run 1 | after run 2 |
|---|---|---|---|
| registered / eligible / skipped | 106 / 104 / 2 | 106 / 104 / 2 | 106 / 104 / 2 |
| process cwds seen | 1,445 | 1,426 | 1,448 |
| plan | 9.82 s (92.7 ms/tree) | 0.56 s (5.3 ms/tree) | 0.66 s (6.3 ms/tree) |
| whole apply | 51.85 s | 15.26 s | 13.25 s |
| move phase (apply - plan) | 42.03 s | 14.70 s | 12.58 s |
| per moved tree | 404 ms | 141 ms | 121 ms |
| moved / failed / reversed | 104 / 0 / 0 | 104 / 0 / 0 | 104 / 0 / 0 |

The "before" tree is `git archive HEAD` of this branch before the change, extracted under the scratchpad and run with the same probe file.

## 5. The announcements

- Plan start (before the first destination is resolved), apply only: `Worktree migration: checking N registered tree(s). With many trees this can take a long time; nothing is moved until the checks finish.`
- Plan scan line: `Checking worktrees (i/N)`, first and last always shown, at most one per second between (`scanRefreshInterval`), then `Worktree migration checked N tree(s)`.
- Move line: `Migrating worktrees (i/N) <name> (elapsed 1m02s, about 3m10s left)` from the second tree once a whole second has passed; estimate = average of finished trees x trees left. The first tree and sub-second runs keep the old text.
- Dry run and `--json` stay silent, as before.

## 6. Item 4 of the card (agent-* tree in the "1/550")

`Classify`: `agent-*` under `<repo>/.claude/worktrees` is `ClassClaudeNative` and is never moved; the same name under `<repo>/.moai/worktrees` is `ClassMoaiMade` by the legacy-root witness and is moved. So an `agent-*` tree counted for the move must sit under `.moai/worktrees` (or its path spelling put it outside the `.claude` match). Which of the two it is on mo.ai.kr was not observed (no read of the real repo). Both answers are pinned by `TestClassify_AgentNamedTreeFollowsItsRoot`; whether an `agent-*` tree under `.moai/worktrees` should be moved is a rule question for the operator, not changed here.

## Verification
- `gofmt -l internal/cli/worktree internal/config`: empty. `GOOS=windows go vet ./internal/cli/worktree ./internal/config`: no output. `GOOS=windows go test -c -o /dev/null ./internal/cli/worktree`: no output.
- New and directly affected tests (listed above), plus a targeted sweep of the migration surface: `go test ./internal/cli/worktree -run "M6|Migrat|PlanMigration|Journal|Rollback|Interrupted|Manifest|Progress|Classify|Verify_|Plan_|Apply_|MovedTo_|UpdateConsolidation" -count=1` -> ran 316 s and printed exactly one `--- FAIL` (below) and no other failure. It ran without `-v`, so the number of tests that passed was not printed; `go test -list` for the narrower pattern `M6|Migrat|PlanMigration|Journal|Rollback|Interrupted|Manifest|Progress|Classify` named 95 tests. Not the full package.

## Gaps
- Windows was not executed (vet and test-compile only). On Windows lsof is absent and the plan skips every tree by design (t2064); the index and the pipeline are platform-neutral code that compiles there.
- No run against mo.ai.kr or any real 648-tree project; the fixture's trees are tiny, so the manifest share on real trees is expected to be larger but was not measured. Numbers are from a loaded macOS machine.
- `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved` failed in the sweep (`move-list results = pending,reversed,moved, want pending,reversed,pending,moved`). It fails on the pre-change tree too: 4 of 10 runs there (`-count=10`), 7 of 10 on the changed tree, and repeated failures in a 20-run `-json` sample on the pre-change tree whose output was truncated before I could count them; the samples cannot separate the two rates. Cause not diagnosed (a guess, unverified: the move-list append drops a line identical to the latest one). Not changed, not skipped; it needs its own card.
- Terminal (TTY) rendering of the new lines was not observed on a real terminal; the non-TTY path is what the tests pin, and the TTY path is `tui.ProgressLine`'s existing one.
- A read-ahead manifest is taken up to 3 trees before the tree moves; a tree edited in that gap is caught by the after-the-move comparison and reversed (fail-safe, pinned by test), which is a visible failure the old order could not produce for edits made in the gap.
- Guard refusal: my first `go test` command used a shell redirect, `;` and a pipe, and the worktree-isolation guard answered: "This session is isolated in the worktree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2070, but this command runs go with the text ^TestT2070TimingProbe$ inside a construct too complex to verify, so what it runs cannot be shown not to be git. Refusing to run it ... Split it into plain, separate commands". I re-ran the same test as one plain command. A second notice (`WorktreeGuardRefusal ... Re-issue it from inside this session's own worktree, without the cross-tree redirect`) arrived later without a visible refusal I could match to a command; every later command returned output, so I took it as the first refusal's echo. The baseline runs used `go -C <scratchpad copy> test`, which the guard allowed.
- `/verify` (the `/moai gate`) was not invoked before the commit: it runs tests beyond the targeted set and the leader's constraint for this card forbids the full suite while the RC41 regression runs. The scoped equivalents above (gofmt, vet incl. GOOS=windows, test-compile, targeted sweep) stand in; the pre-commit hook may re-run the gate at commit time.
- Lease: `moai factory renew t2070 --run to0ck1` after the commit answered `card lease expired at 2026-11-06T04:05:07.014936Z`; I did not renew during the long measurement and did not re-lease. The commit is 2ac39de91c on WT-moai-update-mo (parent merge 98ac41c7ca of t2064's c7c8cc05c2).
- Landing attempt (operator-approved in this session, git-strategy manual workflow reads `af-flow`): `moai factory next --card t2070 --run to0ck1` from the parent checkout re-leased the card (output: `t2070 stage=- worktree=t2070`, note `open pull requests unavailable (no open-PR query was taken: no lease was reachable at the attempt)`). Then `moai factory complete t2070 --run to0ck1` was REFUSED, verbatim: `Factory complete: refused — card t2070 is "", not merge-ready; the integration branch and the card state are unchanged (REQ-MWQ-019 step 1).` The ledger holds the card at an empty state on the dead run `to0ck1`; I did not advance its stage by hand or bypass the refusal. Nothing was merged; HEAD of the branch stays 2ac39de91c.
- Stage walk (leader's instruction, from the parent checkout, `--run to0ck1`), original outputs: `moai factory stage t2070 run .moai/reports/t2070/progress.md` -> `t2070 run v7 lease renewed`. `moai factory stage t2070 sync .moai/reports/t2070/progress.md` -> `Factory stage: card evidence refused: ".moai/reports/t2070/progress.md" is not a commit SHA.` Then `moai factory stage t2070 sync 2ac39de91c` -> `t2070 sync v8 lease renewed`. `moai factory stage t2070 merge-ready 2ac39de91c` -> `Factory stage: illegal card transition: sync → merge-ready.` `moai factory stage t2070 sync-audit 2ac39de91c` -> `Factory stage: card evidence refused: artifact path "" is not usable.`
- Why I stopped at `sync` (read from internal/homestate/card_transition.go and card_evidence_readers.go, not guessed): the table has sync -> sync-audit (T11, evidence `<sha>:<artifact>` where the artifact must exist AT that commit) and sync-audit -> merge-ready (T13, needs a PASS verdict file `.moai/reports/t2070/sync-audit*.md` with `audited_sha`). `.moai/reports/t2070/progress.md` is not tracked, so it is absent at 2ac39de91c; naming some other tracked file as the artifact would only satisfy the format. The T13 verdict is an independent audit's output; a paid audit is outside this lane's constraints and I did not write one. Card t2070 therefore sits at stage `sync` (v8) on run to0ck1 and `moai factory complete` still cannot take it.
- Independent sync audit (Claude `sync-auditor`, no Codex/GLM, one round, audited commit 30392ed445): report `.moai/reports/t2070/sync-audit.md`, machine lines `verdict: FAIL` / `audited_sha: 30392ed445f084485a55d277fbf5fd76becd4aff`. The FAIL is not a code finding: `.moai/config/sections/workflow.yaml` sets `audit.gates.codex: required`, no Codex review was run (the brief forbade external backends), so no receipt is cited and the guard refuses a pass. The code review alone is PASS-WITH-DEBT 0.89 with no BLOCKING defect. The operator was asked (this session's user, AskUserQuestion) whether to authorize one `codex_audit` and answered: do not run it, hold. So the card is NOT merge-ready: no `sync-audit` entry, no `merge-ready`, no `complete`.
- Audit finding F1 (`migrateOneTree` had no caller left, a probable `unused` lint hit; confirmed by grep: only its definition remained): fixed in this branch by deleting the wrapper and moving its doc comment onto `migrateOneTreeWith`; gofmt empty, vet (darwin and GOOS=windows) no output, engine tests ok (82.8 s). The operator chose to do this repair. golangci-lint was not run, so the lint claim stays unconfirmed. Audit finding F2 (the auditor reads `revLines` as returning lines in file order, not last first, which would explain the long-standing flake of `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved`) is a reading of the code, not a measurement: recorded as a follow-up card, not touched here.
- Not changed: concurrent `git worktree move` (option below) and a parallel after-the-move comparison.

## Options for the leader/operator (not implemented)
1. Parallel `git worktree move` across trees of one repository: 8 concurrent moves took 300 ms vs 724 ms serial with 0 failures in two samples; needs a proof of git's registration locking plus a journal/move-list design that stays recoverable per tree before it is safe to adopt.
2. Read the after-the-move manifest of a tree while the next tree moves: saves the second capture's latency; a mismatch would then be found after a later tree has already moved, which changes the reverse-order guarantee.
3. Raise `DefaultWorktreeMigrateManifestLookahead` (a chosen value, 3): trades a wider capture-to-move gap for more overlap.
