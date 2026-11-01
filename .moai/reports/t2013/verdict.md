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
