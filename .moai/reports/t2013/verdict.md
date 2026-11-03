# t2013 verdict - completion guidance order (decisions row R22)

Card t2013, Class B. Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013`, branch WT-agent-trunk-factory, base cb30e05b3e. Local commits only; nothing pushed.

## Claim

1. Reproduced: before the change, `factoryPrintClearPolicyLine` printed `ExitWorktree(keep)`, then `moai factory next`, then `EnterWorktree(...)`. It never mentioned a default-branch check or /clear, and `factory_m6_test.go` asserted that shape (and asserted `/clear` absent). The hypothesis holds on this tree.
2. After the change the printed line orders: preserve-first checklist, (1) ExitWorktree(keep), (2) default-branch check, (3) "operator: run /clear", (4) next-card request. If anything is incomplete or pending it says do not clear; it never suggests a force-clear.
3. The line is printed only on completed edges (merged-local adoption, merged-local merge step, merged-pr). pr-open, a still-open PR, a closed PR, audit-pending, unpushed (no origin), a conflicting card and a merged-local refusal print no completion wording.
4. Nothing in code or tests claims ExitWorktree or /clear ran. The line says it is guidance only.
5. /clear measured fact (recorded verbatim, nothing beyond it): the Skill tool in this harness states "Built-in CLI commands (`/help`, `/clear`, ...) aren't skills", and no tool in this session's tool list executes /clear. So the guidance addresses the OPERATOR; the session only prints the instruction.

Commits:
- RED `7169d369bedf8f03f7895740d320aebdd6fa4aa0` (tests only)
- GREEN `aa457ae7afba509d9a79a6fcbc89b6b615db307a` (code + rule + template mirror)
- this evidence commit is the final HEAD (a file cannot name its own commit; read `git rev-parse HEAD`).

## Evidence

Reproduction, command `go test -run 'TestSD_AC020_ClearPolicies' -count=1 -v ./internal/cli/` on base cb30e05b3e (parent of RED): `--- PASS: TestSD_AC020_ClearPolicies (9.33s)`, 13 subtests PASS, `ok ... 10.935s`. The old line, verbatim from the RED run output:

```
card transition (policy clear-each): ExitWorktree(keep), then `moai factory next` from the primary checkout, then EnterWorktree(<leased worktree path>) and verify the toplevel; earlier cards' conversation context is non-authoritative
```

RED, command `go test -run 'TestSD_AC020_ClearPolicies|TestT2013_' -count=1 ./internal/cli/` at 7169d369: FAIL. Verbatim excerpt:

```
--- FAIL: TestT2013_CompletionGuidanceWording (0.01s)
    --- FAIL: .../auto_dispatch_prints_the_ordered_guidance
        factory_complete_guidance_t2013_test.go:79: guidance does not name "default branch":
--- FAIL: TestT2013_CompletionGuidanceOnlyOnCompletedEdges (45.27s)
    --- FAIL: .../pr_merged_(merged-pr)_prints_the_ordered_guidance (6.47s)
--- FAIL: TestSD_AC020_ClearPolicies (10.50s)
    (clear-each, absent policy, clear-when-full x2, complete end-to-end: FAIL, same reason)
FAIL	github.com/modu-ai/moai-adk/internal/cli	57.608s
```

In the RED run every negative-edge subtest (pr-open, still-open, closed, audit-pending x2, unpushed, conflicting, merged-local refusal) passed: they are guards that pin existing behavior, not reds. RED was red only for wording/order.

GREEN, same command at aa457ae7: `--- PASS: TestT2013_CompletionGuidanceWording`, `--- PASS: TestT2013_CompletionGuidanceOnlyOnCompletedEdges (51.71s)` with 9 subtests PASS, `--- PASS: TestSD_AC020_ClearPolicies (14.26s)` with 13 subtests PASS, `ok ... 67.681s`.

New line, exactly as printed for clear-each, auto dispatch:

```
  card complete (policy clear-each): before leaving, check and preserve completion evidence, dirty changes, active checks or handoff, and unsubmitted input; if any is incomplete or pending, resolve it first and do not clear (never force-clear). Then, in order: (1) ExitWorktree(keep); (2) confirm the primary checkout is back on the default branch (`git branch --show-current`); (3) operator: run /clear (this session cannot run /clear itself); (4) request the next card with `moai factory next` from the primary checkout. This line is guidance only and does not show that any step ran; earlier cards' conversation context is non-authoritative
```

Manual dispatch variant: header `card complete (policy <p>, manual dispatch):`, same preserve-first and steps 1-3, step 4 is `ask the leader for the next card, do not lease one`. Relaunch line and clear-when-full gating code paths are unchanged (relaunch subtest still asserts "end this session" and no `/clear`).

Edges tested (`internal/cli/factory_complete_guidance_t2013_test.go`, using agenticTrunkPrimaryFixture / ghfNew / newGHDouble / sdMergeFixture):
- prints guidance: merged-pr (gh double MERGED), merged-local end-to-end (existing `complete prints the policy line end-to-end`, assertions updated)
- prints nothing: pr-open; PR still open on re-run (asserts "still open" note); PR CLOSED (error, no merged-pr); audit-pending (no sync-audit record; `sync_status: complete-pending-audit`); unpushed (no origin remote); conflicting card; merged-local refused for missing re-measure record
Call-site check: `grep -rn factoryPrintClearPolicyLine internal` shows exactly three callers (factory_card.go adoption, factory_card.go merge step, factory_card_pr.go after `factoryCloseLaneCard` in merged-pr), all after the merged transition.

Other scoped runs:
- `go test -run 'TestBindingLedgerIntegrity' -count=1 -v ./internal/template/` -> `--- PASS: TestBindingLedgerIntegrity (4.42s)`, `ok`
- `go test -run 'TestAlwaysLoaded|Mirror|Parity' -count=1 -v ./internal/template/` -> ok 105.330s; includes `TestRuleTemplateMirrorDrift` PASS and `TestDeclaredRuleMirrorForks` PASS
- `gofmt -l` on factory_card.go, factory_m6_test.go, factory_complete_guidance_t2013_test.go -> no output
- `go vet ./internal/cli/` -> no output, exit 0
- `env GOOS=windows go build ./internal/cli/` -> no output, exit 0
- `git status --short` after the tests -> only the 3 intended modified files (before the GREEN commit); no unexpected tracked change.

Wording differences, `factory_card.go` completion line vs `factory-dispatch.md` section `## Lane card transition - in-session, no clear` (read at base, line numbers of the local file):
1. Old code and rule agreed: ExitWorktree(keep), lease, EnterWorktree (rule lines 159 and 162). New code diverges: adds default-branch check, operator /clear, preserve-first. The rule's role-core line 159 is left untouched (it is binding-ledger text).
2. Rule line 156: "the leader does not ask the operator to clear the lane session". New printed line asks the operator to run /clear. Conflict. Handled by a new paragraph after line 166 in both rule copies declaring the printed guidance governs. Heading "in-session, no clear" and line 156 were NOT edited (see Gaps).
3. Rule line 162 requires `EnterWorktree` of the leased tree and a `git rev-parse --show-toplevel` check; the new line stops at "request the next card with `moai factory next`" and does not mention EnterWorktree or the toplevel check (those follow the lease and remain in the rule).
4. Rule line 162 "never a session clear" for a refused entry: still stands for that path; the printed line fires only after completion.
5. Rule line 162 relaunch-policy wording matches the unchanged relaunch line. clear-when-full gating appears only in code and its test, not in the rule.
6. Rule line 166 (leader session cleared between cards) is consistent with the new line.
7. Left unchanged and now in tension: `factory-dispatch-core.md` ("The leader never asks the operator to clear a lane"), `factory-dispatch-cards.md` ("asks the operator to clear nothing"), local `.claude/rules/local/gitflow-lane-protocol.md` section 1 ("no operator clear is involved").

Template mirror impact: changed under `internal/template/templates/`: only `.claude/rules/moai/workflow/factory-dispatch.md` (one added paragraph, identical text to the local copy, same commit). The file is role-gated (`paths:` frontmatter), not part of the always-loaded surface. The binary change (printed line) reaches a customer project when its `moai` binary is upgraded; its deployed rule text stays old until `moai update`; no forced sync was run. Until then a customer sees a printed line that contradicts its deployed rule text (rule says "no clear").

SPEC conflict note (for the leader to amend later; `.moai/specs/**` not touched): SPEC-FACTORY-LANE-NO-CLEAR-001 REQ-LNC-007 and REQ-LNC-009, as quoted in the doc comment of `factoryPrintClearPolicyLine` ("No value instructs a session clear"), conflict with R22. Judgment: R22 is a later operator instruction and overrides it for the operator-facing completion guidance; the lane's in-session transition remains the documented route when the operator does not clear. The old assertions that encoded those REQs (`strings.Contains(out, "/clear")` must be false) were replaced, not deleted: they now assert the new contract.

## Baseline-attribution

- Tree: base `cb30e05b3e`; RED `7169d369bedf8f03f7895740d320aebdd6fa4aa0`; GREEN `aa457ae7afba509d9a79a6fcbc89b6b615db307a`. The test commands above were run by me in this tree in this session. Each pass/fail quoted comes from the run named next to it.
- Judging build: the tests compile and run from this tree (`go test` builds the test binary from source), so no installed `moai` binary judged anything. The `moai` MCP server was never used for a measurement.
- The R22 text read here is the `grep -n R22` output of `/Users/goos/MoAI/moai-adk-go/.moai/reports/pr-repair-20261101/decisions.md` (file changed on disk mid-task; row R22 re-read, acceptance sentence unchanged).
- Leader evidence: `/Users/goos/MoAI/moai-adk-go/.moai/reports/adk-workflow-pages-handoff.md` lines 124-128 (section "완료 후 공식 전환 절차") read by absolute path; the header section "지금 적용할 최신 요구" was located at line 5 but I did not read its body. Nothing was copied or committed from that file.

## Gaps

- The SPEC bodies (REQ-LNC-007/009) were not opened; they are named from the code doc comment only.
- The header bullet "지금 적용할 최신 요구" of the handoff file (line 5 onward) was not read, only the lines 124-128 section. The four preserve-first items come from that section and the leader message.
- Completion via the MCP `factory_complete` surface was not examined; I only checked that all CLI call sites of the line are post-merge.
- Whether a card with dirty changes or an active check reaches `merged-pr` is not detected by code; the dirty/active/handoff/unsubmitted-input check is text for the reader, not a mechanical gate. I did not test a "dirty worktree" edge because `complete` has no such check; that edge does not exist.
- "Unpushed" exists only as the no-origin refusal and the pr-open state; I did not test a push-failure refusal (not constructible with the fixture I used).
- The heading `in-session, no clear`, line 156 of the rule, `factory-dispatch-cards.md` line 52, `factory-dispatch-core.md` and local `gitflow-lane-protocol.md` section 1 were left as is (binding-ledger text pinned by `internal/template/testdata/binding_ledger.json`, or a mirror pair outside the minimal change). Their text now disagrees with the printed line; reconciling is a follow-up.
- Local rule file and template mirror differed before my edit (first difference at line 224, `cmp` exit 1, measured at base); I did not reconcile that. `TestRuleTemplateMirrorDrift` and `TestDeclaredRuleMirrorForks` pass.
- No whole `./internal/cli/...` run and no `make build` or catalog regen (forbidden/not requested). `go vet` compiled only.
- Mechanical refusal record: no Bash call was refused by the worktree guard in this task (single plain commands used). Two grep calls with `--include=...` globs failed under zsh ("no matches found"), a shell error and not a guard refusal; I re-ran them without the glob.
- `/clear` capability: measured only as stated in Claim 5; I did not probe other harnesses.

## Residual-risk

- A reader may follow step 3 without having done step 1/2; the line states it does not prove they ran, and nothing enforces order.
- The tests assert substrings and relative order of the markers; a future edit keeping the markers but changing meaning would still pass.
- Customers on an older rule text see the contradiction until `moai update`; if the binary ships before the template refresh, lanes there may refuse the operator-/clear instruction on the strength of the old rule.
- Other docs and tests outside the scoped runs (anything asserting the old `card transition (policy` string) were found by one grep over `internal` and `cmd` (3 files only); a consumer outside those trees would not have been seen.

## Follow-up: rule wording alignment (2026-11-01)

### Claim

1. The three factory-dispatch rule files (local copy and template mirror, six files) no longer carry "asks the operator to clear nothing", "does not ask the operator to clear the lane session", "The leader never asks the operator to clear a lane", or the heading "in-session, no clear". They now state the completion order ExitWorktree(keep), default-branch check, operator /clear (a session cannot run /clear itself), next-card request, with the preserve-first condition; the no-card-context-carries-over clause is kept.
2. The heading was renamed to "Lane card transition — completion order and in-session move"; no other file cites the old heading.
3. The one binding-ledger row that pinned the old sentence (ALB-0286, after_text) was updated to the new sentence; no other row changed.
4. Local and template copies received identical hunks.

### Evidence

Baseline before editing, `go test -run 'TestBindingLedgerIntegrity|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestAlwaysLoaded' -count=1 -v ./internal/template/` (filtered to result lines): `--- PASS: TestBindingLedgerIntegrity`, `--- PASS: TestDeclaredRuleMirrorForks`, `--- PASS: TestRuleTemplateMirrorDrift`, `ok`. The `TestAlwaysLoaded` pattern matched no test (the real names are `TestDeployedAlwaysLoadedCharBudget` and `TestDeployedAlwaysLoadedSurfaceDerivation`); re-run afterwards, see below.

RED on the unedited rule files, `go test -run TestT2013_RuleWordingMatchesCompletionOrder -count=1 ./internal/template/` (new test `internal/template/factory_completion_order_t2013_test.go`), excerpt:
```
--- FAIL: TestT2013_RuleWordingMatchesCompletionOrder (0.00s)
    --- FAIL: .../template/factory-dispatch.md
        stale phrase still present: "does not ask the operator to clear the lane session"
        stale phrase still present: "in-session, no clear"
        order marker "next card" missing or out of order after offset 19998
    --- FAIL: .../template/factory-dispatch-core.md
        stale phrase still present: "The leader never asks the operator to clear a lane"
    --- FAIL: .../template/factory-dispatch-cards.md
        stale phrase still present: "asks the operator to clear nothing"
    (the same three failures for the local copies)
FAIL	github.com/modu-ai/moai-adk/internal/template
```
One intermediate run after the edits failed on my own test (marker "default branch" vs the text's "default-branch"); I corrected the test marker, not the rule text.

GREEN after the edits: `--- PASS: TestBindingLedgerIntegrity`, `--- PASS: TestT2013_RuleWordingMatchesCompletionOrder`, `--- PASS: TestDeclaredRuleMirrorForks`, `--- PASS: TestRuleTemplateMirrorDrift`, `ok ... internal/template`. `go test -run TestDeployedAlwaysLoaded -count=1 ./internal/template/` -> `ok`. `go test -run 'TestSD_AC020_ClearPolicies|TestT2013_' -count=1 ./internal/cli/` -> `ok ... internal/cli 33.212s`. `gofmt -l internal/template/factory_completion_order_t2013_test.go` -> no output. `go vet ./internal/template/` -> no output, exit 0.

Heading citation sweep, `grep -rn "in-session, no clear"` over internal, .claude, .moai/docs, cmd: only the new test (as a forbidden string). `git diff --stat`: 7 files, 13 insertions, 13 deletions.

Byte sizes (local copy, HEAD -> now): factory-dispatch-core.md 9969 -> 10164 (+195, the always-loaded stub); factory-dispatch-cards.md 11671 -> 11760 (+89); factory-dispatch.md 28279 -> 28116 (-163). Net across the three: +121 bytes; always-loaded stub growth +195 < 1,000.

### Baseline-attribution

Tree: HEAD 00c3162cb897cf846f48f4a87842ddfbaf05e878 plus the uncommitted edits listed above, measured in this session by `go test`, which builds from the tree; no installed `moai` binary judged anything. Byte sizes: `wc -c` on the edited files and `git cat-file -s HEAD:<path>` on the originals.

### Gaps

- `.claude/rules/local/gitflow-lane-protocol.md` section 1 ("no operator clear is involved") and `.moai/specs/SPEC-FACTORY-LANE-NO-CLEAR-001` were deliberately not touched (forbidden by the task); they still disagree with the new order.
- `.claude/agents/moai/manager-todo.md:68` and `.claude/skills/moai/workflows/gtd.md:349` ("receive no clear guidance") were not edited; they describe non-completed edges (no guidance), which is still true, so I judged them consistent but did not change them.
- The role-core line `[HARD] ExitWorktree(keep), lease, EnterWorktree(path); conversation context never carries.` (ledger-pinned, injected at session start) was left as is; it omits the default-branch check and /clear and is therefore incomplete relative to the new order, not contradictory.
- I did not run the whole `./internal/template/` or `./internal/cli/...` packages (forbidden/scoped only), no `make build`, no `moai update`.
- No Bash call was refused by the worktree guard in this follow-up.

### Residual-risk

- Substring-order test: a future edit keeping the markers but changing meaning would still pass.
- Customer projects keep the old deployed rule text ("no clear") until `moai update`; the binary's printed line and their deployed rules disagree until then. Template files changed: factory-dispatch.md, factory-dispatch-core.md, factory-dispatch-cards.md, plus testdata/binding_ledger.json (test fixture, not deployed). No forced sync was run.

## Follow-up 2: verdict-file gap wording

### Claim

The rule sentence "an absent, unreadable, or uncommitted verdict file is a gap" contradicted `.gitignore` (`.moai/reports/*`): verdict files cannot be committed, and the consumers (`internal/homestate/card_evidence_readers.go`, `internal/factorylane/merge_specless.go`) read them from disk and bind `audited_sha:` to the card's HEAD ancestry. The stub sentence (factory-dispatch-core.md, live and template) now reads: "an absent or unreadable verdict file, or one whose `audited_sha` is not bound to the card's HEAD (not an ancestor of it, or stale), is a gap exactly like a missing progress record." The body copy (factory-dispatch.md, live and template) carries the same sentence plus a rationale tail: the file is read from disk, not git; verdict files are gitignored and never committed, so commit state is not part of the test. The requirement is not lowered.

### Evidence

- RED, commit 66ce5c8ba9: `go test ./internal/template/ -run TestT2013_VerdictGapWording` FAIL on all four file subtests (stale phrase present, `audited_sha` and `HEAD` markers missing).
- GREEN, `go test -run 'TestBindingLedgerIntegrity|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestT2013|TestDeployedAlwaysLoaded' -count=1 -v ./internal/template/`: PASS for TestDeployedAlwaysLoadedCharBudget, TestDeployedAlwaysLoadedSurfaceDerivation, TestBindingLedgerIntegrity, TestT2013_RuleWordingMatchesCompletionOrder, TestT2013_VerdictGapWordingDropsUncommitted, TestDeclaredRuleMirrorForks, TestRuleTemplateMirrorDrift; `ok github.com/modu-ai/moai-adk/internal/template 3.003s`.
- `gofmt -l` on the new test file: no output. `go vet ./internal/template/`: no output, exit 0.
- Ledger: only ALB-0281 `after_text` changed (the stub sentence); `git diff --numstat` shows 1/1 on the ledger. Stub growth is about 76 bytes (core file 10245 bytes after edit), below the 1,000-byte threshold.

### Baseline-attribution

Tree: HEAD 66ce5c8ba9 plus the uncommitted edits listed above, measured in this session by `go test`, which builds from the tree; no installed `moai` binary judged anything.

### Gaps

- No "clone environment" procedure for TestBindingLedgerIntegrity was found in verdict.md or progress.md; the earlier runs there used a plain scoped `go test -run ... ./internal/template/`, and this follow-up did the same.
- AC-ALB-021(2) failed once on the stub-only edit: ALB-0281 (always: stub row) matched the changed full-body line in the on-demand file, and the anchor-copy exemption no longer applied. Verbatim: `ledger integrity: AC-ALB-021(2): binding fragment of row ALB-0281 appears in on-demand file .claude/rules/moai/workflow/factory-dispatch.md`.
- A new ledger row ALB-0364 (rule-body copy) was tried and rejected, verbatim: `row coverage: row ALB-0364 duplicates a claim on unit workflow/factory-dispatch.md:116 — exactly one row per unit`. It was removed; the removal used `sed -i` plus a `printf` re-append of the file's closing braces (confirmed by `git diff --numstat` showing only the ALB-0281 line changed). One stray empty `python3` heredoc was run during that step and did nothing.
- Resolution: the coordinator's third route. The stub sentence stays exactly as in the ledger; the body copy is made a superset line (rationale tail appended), so the whole-line sweep no longer matches. No ledger row, test, or anchor change beyond ALB-0281 `after_text`.
- Not run: the full `./internal/template/` package, `./internal/cli/...`, `make build`, `moai update`. `.claude/rules/moai/workflow/gitflow-lane-protocol.md` and SPEC-FACTORY-LANE-NO-CLEAR-001 were not touched.

### Residual-risk

- The sweep exemption is whole-line equality: a future edit that makes the body line equal to a ledger line again would re-trigger AC-ALB-021(2).
- The wording test checks markers (`audited_sha`, `HEAD`, absence of the old phrase), not meaning.
- Customer projects keep the old deployed sentence until `moai update`.


## Follow-up 3: lane wait record and wake (2026-11-01)

### Claim

1. A lane's wait is recorded mechanically (additive `card_waits` table, `moai factory wait` verb) with waker and condition. 2. The `moai factory complete` landing edges (merged-local via completeTransitions, merged-pr via factoryCompleteGitHubFlow) print, once, `wake: lane=<waker> card=<id> reason=<...>` for each wait the landing satisfies; re-run prints nothing; operator waker prints a leader note only; unmet condition or lease expiry prints nothing. 3. The wake output points the reader at `moai factory status` lease holder and the card progress record. 4. Base-branch question investigated, no code changed.

### Evidence

- RED commit 54be323517, `go test ./internal/cli -run TestFactoryWake -count=1`, verbatim tail: `--- FAIL: TestFactoryWakeOnLanding/predecessor_landed_local,_lane_waker:_one_wake,_rerun_silent` ... `factory_wait_test.go:96: first print: want 1 wake line(s), got 0:` ... `--- FAIL: TestFactoryWakeTwoWaitsOnePredecessor` ... `want 2 wake lines, got 0:` ... `FAIL github.com/modu-ai/moai-adk/internal/cli 1.296s`.
- GREEN commit 61c9c1e9f0: same command, 7 subtests PASS (`ok github.com/modu-ai/moai-adk/internal/cli 1.034s`).
- `go test ./internal/homestate -count=1`: `ok github.com/modu-ai/moai-adk/internal/homestate 179.903s`.
- `go test ./internal/cli -run 'TestFactoryComplete|TestMWQ|TestFactoryGitHubFlow|TestFactoryWake|TestComplete|TestFactoryPR|TestS03' -count=1`: `ok ... 157.388s`.
- `go vet ./internal/cli ./internal/homestate`: no output. `gofmt -l internal/cli internal/homestate`: no output (after removing a trailing comment that gofmt realigned).
- Mutants via `go test -overlay` copies in the scratchpad (tree untouched), all `-run TestFactoryWake`:
  - M1 drop the woken_at guard (SELECT and UPDATE): FAIL `rerun 1 must print nothing, got: wake: lane=lane-3 card=t2020 ...` (3 subtests). A first attempt dropping only the SELECT guard SURVIVED (`ok`), because the UPDATE `WHERE woken_at=''` re-guards; that is double protection, not a test hole.
  - M2 ignore waker=operator: FAIL `operator_waker_on_a_predecessor_wait ... want 0 wake line(s), got 1: wake: lane=operator card=t2020`.
  - M3 remove the landed-state guard (print on a lease-expired, unlanded predecessor): FAIL `predecessor_only_has_an_expired_lease,_not_landed ... want 0 wake line(s), got 1`.
  - M4 ignore the predecessor ref: FAIL `a_different_card_landed ... want 0 wake line(s), got 1: wake: lane=lane-3 card=t2020 reason=r (predecessor t2011 landed)`.

### Design

Reused: the existing `after` hint is the default predecessor of `moai factory wait` (no restated dependency); the additive-table precedent of `card_metadata_events`. New: `card_waits` (card_id, cond_kind, cond_ref, waker, reason, recorded_by, recorded_at, woken_at), created by `CREATE TABLE IF NOT EXISTS` in factoryDDL, so no schema version bump and no existing row touched. Not reused: `decision_*` columns hold one gate per card and are a card state (needs-decision), not a per-condition wake target.

### (4) Base branch: investigation only, no code changed

- Chosen at `internal/cli/session_worktree.go:241` `base := config.LoadWorktreeBaseBranch(projectRoot)` (`git_strategy.worktree_base_branch: main`), reached from `factoryCreateAndRenameWorktree` (factory_card.go:997) via `worktree.WorktreeCreator` = `materializeSessionWorktree`.
- Operand: `gitWorktreeAddArgs` appends the BARE name (`main`) so git resolves LOCAL `refs/heads/main`.
- Fetch: none on this path (grep of `"fetch"` in session_worktree.go and the factory next path finds no call; fetches exist only in factory_card_pr.go:236, 337, 566 for the PR edge). Local main is therefore as stale as the last manual resync, matching the leader's 14-behind observation.
- Resolve check: `hook.WorktreeBaseBranchResolvable` = `git show-ref --verify refs/remotes/origin/<base>` (internal/hook/worktree_base_branch.go:194). It tests the REMOTE-tracking ref, while the operand is the LOCAL ref: the two can disagree. If `origin/main` is absent (no remote, never fetched, shallow), `base` is silently set to `""` (session_worktree.go:242-244) and the tree is cut from the invoking tree's HEAD, with no warning (the planned warning is t1847 per gitflow-lane-protocol.md section 1).
- Minimal change (a) refuse on unresolvable base: at session_worktree.go:242-244 return an error instead of `base = ""` when `base != "" && !resolvable`; touches the plan D4 decision (SPEC-WORKTREE-BASEREF-001 chose fall back, "a tree on the wrong base is recoverable"), and existing tests of that fallback would flip, so it is not "clearly bounded"; not implemented.
- Minimal change (b) follow origin: a `git fetch origin <base>` before the add and operand `origin/<base>`. This CONTRADICTS the local rule `.claude/rules/local/gitflow-lane-protocol.md` section 1 (local main is deliberately both base and landing surface; origin/main advances only by release batches). Leader's decision; not implemented.

### Baseline-attribution

Tree: HEAD 61c9c1e9f0 plus this evidence commit; all results from `go test`/`go vet` built from this tree in this session; no installed `moai` binary judged anything. Interaction with t2009: commit 2d4a257875 (branch-local, not in this HEAD) edits `completePostMergeConflict` at about factory_card.go:3341; this card edits `completeTransitions` (one line after the `Fprintf` at about line 3295) and factory_card_pr.go `factoryCompleteGitHubFlow` (one line before `factoryPrintClearPolicyLine`, about line 584), plus factory_handoff_recover.go (one `AddCommand` line) and factory.go DDL. Adjacent but non-overlapping hunks in factory_card.go (about 45 lines apart).

### Gaps

- NOT executed: any real cross-session wake or send, any live session, `moai factory wait` through the real cobra tree against a live run (the verb's `--on` default via the card `after` hint is untested), the MCP surface (no MCP tool added for `wait`).
- Not run: full `./internal/cli/...`, `moai update`, `make build`.
- Refused command: one compound Bash (variable assignments plus `cp` and `sed -i` chain with `$S`/`$T`) was refused by the worktree guard ("runs sed with a value computed at runtime"); I redid the copies and edits as plain single commands and Edit calls. No result depends on the refused command.
- No wake path exists for lease expiry (no code calls the printer there); the lease-expiry case tests that a not-landed, lease-expired predecessor prints nothing, i.e. the guard, not a lease-expiry hook.
- Wake recording at `moai factory next`/lease-expiry recovery (8859) is not wired; a wait on a lease-expired card is not woken.

### Residual-risk

- Printed wake can be missed if the completing session's output is not read; the record (`woken_at`) marks it handed out, so a lost wake is not re-printed. The leader reads `card_waits` (via future status surface) or the progress record.
- `woken_at` is stamped before the session sends; a crash between print and send loses the nudge (queue-and-evidence reads still cover it, per factory-dispatch).
- Always-loaded rule text growth for this follow-up: 0 bytes (no rule file touched).


## Follow-up 4: wait verb reaches lanes through the rule

### Claim

The lane-wait section of the factory-dispatch rule body now tells a lane to record a wait with `moai factory wait <card> [--on <predecessor card>|--operator] --waker <lane|leader|operator> --reason "<why>"`, and tells the session whose `moai factory complete` printed a `wake:` line to send ONE message to the named lane; the verb sends nothing, and a lane confirms start from the lease holder (`moai factory status`) and the progress record, never from a reply. Placement is body-only: factory-dispatch.md (template source plus live mirror), one [HARD] paragraph after the long lane-wait paragraph, outside the role-core markers. The always-loaded stub factory-dispatch-core.md is unchanged.

### Evidence

- RED commit 68d8078a35, `go test -run TestT2013_WaitVerbReachesLanesThroughRule -count=1 ./internal/template/`, verbatim tail: `lane-wait section lacks marker "moai factory wait"` / `"`wake:` line"` / `"--waker"` / `"sends nothing"` for template and local, `FAIL github.com/modu-ai/moai-adk/internal/template 0.352s`. (A first RED run also showed the bare marker `wake:` already matched the existing words "on wake:"; the marker was tightened to the backticked `wake:` line before the RED commit.)
- GREEN commit 76b3df294c, `go test -run 'TestBindingLedgerIntegrity|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestT2013|TestDeployedAlwaysLoaded' -count=1 ./internal/template/`: `ok  	github.com/modu-ai/moai-adk/internal/template	2.971s`. One earlier run failed only on my own marker (`wake:` line) because my first wording said `a `wake: lane=...` line`; the wording was adjusted, not the test.
- `make build`: exit 0, ended with `go build ... -o bin/moai ./cmd/moai`; `git status --short` afterwards listed only the two intended rule files, so no regenerated tracked file needs inclusion.
- Byte growth: template factory-dispatch.md 28035 to 28521 (+486 bytes, via `git cat-file -s HEAD:...` and `wc -c`); live mirror +486 (28357 at HEAD to 28843 by the same added paragraph); factory-dispatch-core.md 10245 unchanged. Below the 1,000-byte rule-authoring threshold, and none of it in the always-loaded surface. No ledger row, anchor SHA or test of the ledger touched.
- Why body-only: lanes read factory-dispatch.md in full by the [HARD] read directive (the role-core injection ends with "read factory-dispatch.md to the end"), so the paragraph reaches them; a stub edit would add bytes to every session and need an ALB row after_text update.

### Baseline-attribution

Tree: HEAD 68d8078a35 plus the GREEN edit (committed as 76b3df294c); results from `go test` and `make build` built from this tree in this session; no installed `moai` binary judged anything.

### Gaps

- Base-branch finding carried verbatim from Follow-up 3 item 4: (a) refuse on an unresolvable base (return an error at session_worktree.go:242-244 instead of `base = ""`) and (b) follow origin default (a `git fetch origin <base>` before the add and operand `origin/<base>`). Option (b) CONTRADICTS `.claude/rules/local/gitflow-lane-protocol.md` section 1 (local main is deliberately both base and landing surface; origin/main advances only by release batches), and option (a) touches the SPEC-WORKTREE-BASEREF-001 D4 fallback decision and would flip its existing tests. NEITHER (a) nor (b) is implemented. The leader decided not to implement it and not to create a card.
- Not run: full ./internal/cli/..., `moai update`, any live-session wake or send, `moai factory wait` through the real cobra tree against a live run.
- The live and template factory-dispatch.md are byte-identical in my added paragraph but not as whole files: `diff` shows a pre-existing difference at line 228 (the live copy keeps a longer sentence tail on the git-flow integration paragraph). TestRuleTemplateMirrorDrift and TestDeclaredRuleMirrorForks pass, so it is a declared fork; I did not touch it.
- Refused commands: none in this follow-up. One `git show HEAD:<path> --no-patch` printed the whole file instead of a size (the flag is ignored for blobs); it was a wasted read, not a refusal.

### Residual-risk

- The wording test checks markers, not meaning.
- Customer projects keep the old deployed rule until `moai update`.
- A lane that never reads the full factory-dispatch.md body does not see the paragraph; the stub still says only "records a disk wait".

## Follow-up 5: merge of origin/main and pre-push verification (2026-11-02)

### Claim
Merging origin/main (merge commit, no rebase/reset/amend/force) into WT-agent-trunk-factory produced no conflict, and the card's affected tests and lint pass on the merged tree.

### Evidence
- Lease: `moai factory renew t2013 --json` -> `refused-expired` (expired 2026-11-01T11:31:50Z). 8859 recovery: `moai factory stage t2013 run --run tnsei4` (refused, expired), `moai factory next --card t2013 --run tnsei4` from the primary checkout, then `moai factory renew t2013 --json` -> `renewed`. Record version 4 -> 6, stage field empty before and after.
- `git rev-list --count --left-right origin/main...HEAD` before merge: `14 13` (behind 14, ahead 13).
- `git merge --no-ff origin/main` -> merge commit 68ea579390, no conflict output, `git status --short` empty afterwards. Conflict content to preserve: none (t2009/AF overlaps in factory_card.go, factory-dispatch rules and binding_ledger.json merged cleanly at the textual level).
- `go test ./internal/cli/ -run TestFactoryWake -count=1` -> ok 0.935s
- `go test ./internal/homestate/ -run Wait -count=1` -> ok 4.791s
- `go test ./internal/template/ -run 'TestT2013|TestBindingLedgerIntegrity|Mirror|Parity' -count=1` -> ok 87.193s
- `golangci-lint run --allow-parallel-runners ./internal/homestate/... ./internal/cli/... ./internal/template/...` -> `0 issues.`, exit 0.

### Baseline-attribution
All commands ran in this run against the merged tree at 68ea579390. The moai binary used for factory verbs is the installed one (build not compared to tree HEAD); the Go tests and lint do not use it.

### Gaps
- `TestT2013` pattern was matched by `-run` alone; I did not list which tests it selected.
- Textually clean merge does not prove semantic compatibility with t2009's factory_card.go or AF's rule/ledger changes; only the scoped tests above were run. Full ./internal/cli/... not run.
- First `golangci-lint run` was refused by golangci-lint itself ("parallel golangci-lint is running", exit 3, another session's run); the passing run used `--allow-parallel-runners`.
- One `rev-parse --short origin/main FETCH_HEAD` returned exit 128 (two-arg form); the divergence count was read separately instead.
- One compound git command was refused by the worktree guard; rerun as plain separate commands.

### Residual-risk
- CI on the PR may fail in jobs not covered by the scoped tests (rules budget, mirror, full suite).

## Follow-up 6: PR #1960 first CI run, failures read and repaired (2026-11-02)

### Claim
The first CI run of PR #1960 (run 41333008569 and sibling runs, head 5fea52b4c7) failed on six jobs; all failures except one calendar test were caused by this card's own commits and are repaired in local commits 22c3381085, 4a80617a90, a32d3a9b57, a6c9084d8d (not yet pushed at the time of this entry).

### Evidence (job logs read via `gh api repos/modu-ai/moai-adk/actions/jobs/<id>/logs`)
- Template Neutrality Audit (job 124926876735) and lsel-leak-guard (124926877014): TestTemplateNoInternalContentLeak, `factory-dispatch.md` leaked `SPEC-FACTORY-LANE-NO-CLEAR-001` (C1) and `t2013` (C9) from this card's completion-guidance line; origin/main has 0 occurrences. Fixed by 22c3381085.
- Race Test (internal/cli 1 of 4) (124926965078): `TestCleanReportsArchive_MovesOnlyEligibleCandidates` ("no new YYYY-MM shard created under archive/") = calendar defect, assigned to PR #1956 by the leader, NOT repaired here; and `TestFactoryClearPolicyManualDispatch` (two subtests) still asserting the pre-t2013 shape. Fixed by 4a80617a90.
- Race Test (internal/cli 4 of 4) (124926965178): `TestSD_AC022_SchemaStatementsFrozen`, pin not moved for the additive `card_waits` table. Fixed by a32d3a9b57 (same path as t1538/t1840).
- Race Test (remainder packages) (124926965229) and Test (ubuntu-latest) (124926965126): same failures above plus `TestRoleCoreStubGate`: `factory-dispatch-core.md` 10167 UTF-16 vs budget 10000 (origin/main 9893). Fixed by a6c9084d8d (stub 9988; ledger rows ALB-0246/0257/0281 after_text only).
- Re-run by me on HEAD a6c9084d8d: `go test ./internal/template/ -run 'TestRoleCoreStubGate|TestBindingLedgerIntegrity|TestTemplateNoInternalContentLeak|TestSplitHarnessNamespaceNoLeak|TestLSELLeakPositiveControl|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestT2013|TestDeployedAlwaysLoaded' -count=1` -> ok 5.108s; `go test ./internal/cli/ -run 'TestT2013|TestFactoryClearPolicy|TestSD_AC020|TestSD_AC022|TestFactoryWake' -count=1` -> ok 28.641s. Agent-reported (not re-run by me): whole `./internal/template/` package ok 317.082s, go vet, golangci-lint 0 issues.

### Baseline-attribution
Failures attributed from the job logs above; origin/main baseline for the rule files read with `git show origin/main:<path>` and `git grep` (0 occurrences; stub 9893 UTF-16).

### Gaps
- Full `./internal/cli/...` and the CI jobs themselves were not re-run locally; the repaired state is judged by the scoped tests only until the next CI run is read.
- My first scoped runs omitted the whole-package template test, which is why TestRoleCoreStubGate was missed before the first push.
- Factory structure stages 1-5 are NOT started (leader withdrew the approval relay).

### Residual-risk
- Stub margin under the role-core budget is 12 UTF-16 units.
- The calendar test stays red on this PR until #1956 lands.
