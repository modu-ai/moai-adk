# SPEC-LANDING-SURFACE-RULES-001 — Acceptance (card t1846)

Verification layer for spec.md §B. Every `AC-LSR-NNN` is Given-When-Then and binary. The requirement layer (GEARS) lives in spec.md; nothing here restates it. Figures and command outputs are in the evidence ledger (§E); the tree pin for the whole document is ledger G0 (`7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`), which binds every criterion that carries no pin of its own.

## §A — How to read a criterion

Each criterion carries a class:

- **flip** — red at the pinned tree, expected green after the named milestone. It names the ledger cell that holds its RED-now observation (command, verbatim stdout, exit code) and says why it is red.
- **guard** — green at the pinned tree by design (a regression guard). It is never counted as a flip and claims no RED-now. Where a positive control or a mutant exists, it is named.
- **chain** — needs a heavy run (`make`, `go test`) that belongs to run-phase slots. It claims no RED-now; its RED is observed at the named run milestone (M4 for a stale generated artifact).

Milestones are in plan.md §C: M1 re-measure, M2 repo-local guidance, M3 template sources, M4 mirrors and generated artifacts, M5 chain and guards. Greps use fixed strings (`-F`) anchored on the sentences this work changes or adds, not on words that pre-exist (the bare phrase `does not check out` already occurs in protocol line 35 and is therefore never used as an anchor).

---

## §B — Criteria

**AC-LSR-001** (maps REQ-LSR-001; flip, RED-now L3). Given the protocol file, When `git grep -noF -e 'landing and resync only' -e 'is the human and leader workspace and does not check out `main`' -- .claude/rules/local/gitflow-lane-protocol.md` runs after M2, Then it exits 0 with at least two lines, and the first line of each phrase lies between the `## 2.` heading and the `## 3.` heading (`git grep -nE '^## [23]\. ' -- .claude/rules/local/gitflow-lane-protocol.md` gives the bounds; at the pinned tree they are lines 33 and 51). Red now because L3 returns nothing (exit 1) for both phrases. Green path: M2.

**AC-LSR-002** (maps REQ-LSR-001; flip, RED-now L1). Given the five primary-as-surface sentences of the protocol file (the `## 2.` heading and four body lines), When the L1 command runs after M2, Then it exits 1 with empty stdout. Red now because L1 returns five lines (33, 35, 43, 137, 178) with exit 0, which is also why the criterion is not vacuous. A run that renames the body but leaves the heading, or the reverse, stays red. Paired with AC-LSR-001: deleting the sentences without adding the rule fails 001, adding the rule while keeping a sentence fails 002. Green path: M2.

**AC-LSR-003** (maps REQ-LSR-002; flip, RED-now L3). Given `AGENTS.local.md`, When `git grep -noF -e 'landing and resync only' -e 'is the human and leader workspace and does not check out `main`' -- AGENTS.local.md` runs after M2, Then it exits 0 with at least two lines, the first line of each phrase lying between the `### §4.0` heading and the `### §4.1` heading (lines 182 and 239 at the pinned tree). Red now because L3 returns nothing (exit 1). Green path: M2.

**AC-LSR-004** (maps REQ-LSR-002; flip, RED-now L2). Given the five primary-as-surface sentences of `AGENTS.local.md`, When the L2 command runs after M2, Then it exits 1 with empty stdout. Red now because L2 returns five lines (184, 192, 194, 198, 204) with exit 0; line 194 is the tool-step description "(3) confirms that the primary checkout is clean (item 3)", which for a separate integration worktree is replaced by the overlap decision (`integration_merge_step.go:309-310`, ledger G8). Green path: M2.

**AC-LSR-005** (maps REQ-LSR-003; flip, RED-now L3). Given the five template sources, When `git grep -c -F -e 'landing and resync only' -- internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` runs after M3, Then it prints five lines, each with a count of at least 1; and the same command with `-e 'is the human and leader workspace and does not check out the integration branch'` prints five lines likewise, and with `-e 'Where the integration branch is held by a dedicated worktree'` (the scope clause, REQ-LSR-003) prints five lines likewise. The placement is read, not grepped: each rule sentence sits in the git-flow `WT-*` passage named in plan.md §B.1, so `spec-workflow.md` Route A and `sync/delivery.md:324` are not contradicted. Red now because L3 and L3c cover these paths and return nothing (exit 1). Green path: M3.

**AC-LSR-006** (maps REQ-LSR-003; guard, positive control L4b). Given the template edits, When `git diff --name-only '-G(SPEC-[A-Z]|[^a-z]t[0-9]{3,4}[^0-9]|20[0-9]{2}-[0-9]{2}-[0-9]{2}|planned in)' 7faf75fdf HEAD -- internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/sync` runs after M3, Then stdout is empty (exit 0): no changed line carries a SPEC id, card id, date or "planned in". Green now by construction (no change yet); the positive control L4b shows the same regex fires on a real difference (it lists the sanitized ops pair, whose differing lines carry card ids). The authoritative check is `TestTemplateNoInternalContentLeak` (AC-LSR-008); this guard is the early warning. Mutant: a template line saying "planned in t1847" turns it non-empty.

**AC-LSR-007** (maps REQ-LSR-004; flip plus guard, RED-now L3 for the flip half). Given the deployed copies, When the five-path command of AC-LSR-005 runs against the deployed paths (`.claude/agents/moai/manager-git.md`, `.claude/rules/moai/workflow/{delivery-policy,spec-workflow,worktree-integration}.md`, `.claude/skills/moai/workflows/sync/delivery.md`) after M4, Then it prints five lines each with a count of at least 1 (flip half, red now per L3). And When the five G5 identity commands run after M4, Then `diff -q` on the manager-git.md pair exits 0, `diff -rq` on the sync directory pair exits 0 with empty stdout, `diff -rq` on the rules workflow directory pair prints exactly the six lines recorded in G5 (exit 1) and no seventh, and `git diff --no-index --numstat` on the ops pair prints `9<TAB>9` (guard half; green now). Mutant: editing a template source and not its deployed copy adds a line to the six-line list or fails `diff -q`.

**AC-LSR-008** (maps REQ-LSR-005; chain, RED observed at run M4, not RED-now). Given the manager-git.md edit, When `make agents-emit` runs at M4, Then it rewrites the template copy only (the update mode writes `internal/template/templates/.codex/agents/moai/*.toml`; ledger G10), so `git grep -c -F 'landing and resync only' -- internal/template/templates/.codex/agents/moai/manager-git.toml` prints a count of at least 1 while `diff -q .codex/agents/moai/manager-git.toml internal/template/templates/.codex/agents/moai/manager-git.toml` still exits 1 (deployed copy stale: this is the step-order check, observed before the copy). The M4 copy step is then `cp internal/template/templates/.codex/agents/moai/manager-git.toml .codex/agents/moai/manager-git.toml` (a byte copy; the two files are byte-identical today, G5, so the copy reproduces the deployed form), after which the same `diff -q` exits 0, `make agents-emit-check` exits 0, and `git grep -c -F 'landing and resync only' -- .codex/agents/moai/manager-git.toml` prints a count of at least 1; at M5 `make embed-manifest-check` exits 0, `make build VERSION=<current rc>` exits 0 under a resource-slot lease and regenerates `internal/template/catalog.yaml` (`git diff --numstat -- internal/template/catalog.yaml` prints `2<TAB>2<TAB>internal/template/catalog.yaml`: the `hash:` lines of the `moai` skill entry at catalog line 9 and the `manager-git` agent entry at line 387, ledger G11; any other count must be explained by pre-existing drift measured at M1 with `gen-catalog-hashes.go --all --dry-run`), and `go test ./internal/template/ -count=1 -v -run '^(TestRuleTemplateMirrorDrift|TestSanitizedPairParity|TestTemplateNoInternalContentLeak|TestRouteBPrecheckCoversAllEntryPoints|TestWorkflowRulePathsPinned|TestDeclaredRuleMirrorForks)$'` prints six `--- PASS` lines (the swept count must be six, never fewer) and `ok`. RED to observe at M4: after editing manager-git.md and before `make agents-emit`, `make agents-emit-check` exits non-zero. These runs are heavy and are not executed in plan phase (ledger G9 records only that the names exist).

**AC-LSR-009** (maps REQ-LSR-006; flip, RED-now L3). Given the two repo-local files, When `git grep -noF -e 'git show main:' -e 'smudge' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md` runs after M2, Then each of the two files contributes at least one line for each phrase (four or more lines in total, exit 0). Red now because L3 returns nothing (exit 1). Green path: M2.

**AC-LSR-010** (maps REQ-LSR-007; flip, RED-now L3). Given the two repo-local files, When `git grep -noF -e 'each untracked or modified file' -e 'no moai verb moves the primary checkout to the current local' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md` runs after M2, Then it exits 0 and each file contributes at least one line (both phrases appear in at least one of the two files; the second phrase sits next to a `planned in t1847` marker, checked by reading the line). Red now because L3 returns nothing (exit 1). Green path: M2. Accepted residual (plan.md R5): a mutant could add the phrase in a meaningless place; the placement is read in the plan audit and the sync audit.

**AC-LSR-011** (maps REQ-LSR-008; flip plus guard, RED-now L3 for the flip half). Given the two repo-local files, When `git grep -noF -e 'branches from the named ref' -e 'falls back to the invoking tree' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md` runs after M2, Then each file contributes a line for each phrase (four or more lines, exit 0): the sentence that says the tree branches from the named ref also names the fallback limit (spec.md §A.3), and the HEAD-equals-`main` confirmation is stated as the check. Red now per L3 and L3c. And When `git grep -noF -e 'point to the same commit' -- .claude/rules/local/gitflow-lane-protocol.md` runs, Then it still prints one line (ledger L5b: line 27 at the pinned tree; guard half, green now) so the existing HEAD-equals-`main` confirmation is not lost. Green path: M2.

**AC-LSR-012** (maps REQ-LSR-009; flip plus guard, RED-now L3 for the flip half). Given the four files `worktree-integration-ops.md` and `worktree-integration.md` (deployed and template), When `git grep -c -F -e 'named local ref' -- .claude/rules/moai/workflow/worktree-integration-ops.md .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` runs after M3 and M4, Then it prints four lines each with a count of at least 1. Red now per L3. And When `git grep -noF -e 'Branches from default remote branch' -- .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` runs, Then it still prints both lines (ledger L5a; guard half, green now). Green path: M3 and M4.

**AC-LSR-013** (maps REQ-LSR-010; flip plus guard, RED-now L3 for the flip half). Given the two repo-local files, When `git grep -c -F -e 'planned in t1847' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md` runs after M2, Then it prints two lines each with a count of at least 1 (flip half, red now per L3). And When `git grep -noF -e 'planned in t1847' -- internal/template/templates .claude/agents .claude/rules/moai .claude/skills .codex` runs, Then it exits 1 with empty stdout (guard half; ledger L6 records the same at the pinned tree). Mutant: a "planned in t1847" sentence leaking into a template source turns the guard half non-empty.

**AC-LSR-014** (maps REQ-LSR-011; guard, positive control L4b-style). Given all edited files, When `git diff --name-only -G'race-safe|guaranteed|always safe' 7faf75fdf HEAD -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/moai/manager-git.md .claude/rules/moai/workflow .claude/skills/moai/workflows/sync internal/template/templates/.claude/agents/moai internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/sync .codex` runs at M5, Then stdout is empty (exit 0). The `-G` form inspects changed lines only, so the two pre-existing negated uses (`worktree-integration-ops.md:248` "Refusal is not guaranteed", protocol line 108 "routing is not guaranteed") never trip it (ledger L4). Positive control: ledger L4b shows `-G` listing a file whose changed lines match. The second half of REQ-LSR-011 (each protective sentence names a limit) is read, not grepped: the plan audit and sync audit read AC-LSR-009's `git show main:` sentence and any sentence about a checkout or a move, and fail the criterion when one claims protection without a limit.

**AC-LSR-015** (maps REQ-LSR-012; guard, allow-list form). Given the run-phase diff, When the allow-list command of ledger L7 (`git diff --name-only 7faf75fdf HEAD -- .` followed by one `':(exclude)<path>'` pathspec for each of the 17 files of plan.md §B.7, for `.moai/specs/SPEC-LANDING-SURFACE-RULES-001` and for `.moai/reports/t1846`) runs at M5, Then stdout is empty: every changed file is on the list, so any other file — a Go file, `cmd/moai/main.go`, `.codex/hooks.json`, a hook, a settings or statusline file, a `.moai/policies` projection, an unlisted sibling rule — prints and fails. The exit code is 0 either way, so the verdict is the stdout. Controls (ledger L7c, run in a scratch repository under the session scratchpad): the legitimate edit set prints nothing (negative control), and a mutant editing `cmd/moai/main.go` and `.codex/hooks.json` prints exactly those two paths, a second mutant adding a Go file, a projection and an unlisted sibling rule prints all three (positive controls; the two mutants are the ones the plan audit showed the earlier path-list form missed). The second half: `git grep -c -F -e 'agentic-trunk' -- <the edited repo-local, template and deployed paths>` exits 1 with empty stdout (ledger L7b). Pin note: the pin is the plan base `7faf75fdf`; if the card tree is rebased onto a later `main`, M1 re-pins and re-runs L7.

---

## §C — Requirement coverage

| Requirement | Criteria | Class mix |
|---|---|---|
| REQ-LSR-001 | AC-LSR-001, AC-LSR-002 | 2 flip |
| REQ-LSR-002 | AC-LSR-003, AC-LSR-004 | 2 flip |
| REQ-LSR-003 | AC-LSR-005, AC-LSR-006 | 1 flip, 1 guard |
| REQ-LSR-004 | AC-LSR-007 | flip plus guard |
| REQ-LSR-005 | AC-LSR-008 | chain |
| REQ-LSR-006 | AC-LSR-009 | flip |
| REQ-LSR-007 | AC-LSR-010 | flip |
| REQ-LSR-008 | AC-LSR-011 | flip plus guard |
| REQ-LSR-009 | AC-LSR-012 | flip plus guard |
| REQ-LSR-010 | AC-LSR-013 | flip plus guard |
| REQ-LSR-011 | AC-LSR-014 | guard |
| REQ-LSR-012 | AC-LSR-015 | guard |

15 criteria cover 12 requirements (Tier M ceilings: 16 and 16).

## §D — Edge cases, quality gates, Definition of Done

Edge cases:

- **Main moves during run.** Ledger G3 shows 12 commits on local `main` after the pinned tree and none touching an in-scope file; if `main` moves again, M1 re-runs G3 and re-pins G0 before any edit, and no AC is re-cited from memory.
- **t1831 lands first.** A.5 of spec.md orders t1831 after this card; if the order inverts, this card's run rebases onto t1831's text and re-runs L1 and L2 to re-anchor line numbers (the anchors are sentences, not line numbers).
- **The ops pair is a sanitized pair.** The paragraph added to `worktree-integration-ops.md` must be byte-identical in both copies and carry no card id or date, so the recorded `9 9` difference does not change (AC-LSR-007).
- **Generated artifact.** Neither `.codex/agents/moai/manager-git.toml` copy is edited by hand. `make agents-emit` rewrites the template copy only (`internal/template/templates/.codex/agents/moai/manager-git.toml`, ledger G10); the deployed copy is then brought in line by a `cp` of that regenerated template toml, confirmed by `diff -q` (AC-LSR-008).

Quality gates: each flip criterion is paired with a presence clause and a removal clause so neither one-sided mutant passes; each guard names its mutant; the chain criterion names its six tests and requires the swept count of six.

Definition of Done: all 15 criteria hold at the run-phase tree; the verdict file under `.moai/reports/t1846/` carries the commands and outputs that decided each criterion (not a pointer to machine-local scratch); the Gaps of plan.md §F are re-stated with their run-phase status.

---

## §E — Evidence ledger

Every cell was run in the card tree (working directory = tree root `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1846`), tree pin G0, by the installed `moai`/`git` named in G4. Stdout is verbatim; where an output line exceeds the ledger width it was produced with `-o` (only-matching) so it is complete, never truncated. Cells marked "(CWD tree)" use paths relative to the tree root.

### G0 — tree pin
```
$ git rev-parse HEAD
7faf75fdfbe27e442ea2c17a1a90caf5792d09e6
exit=0
```

### G1 — the card grep (scope classification source; 47 files)
```
$ git grep -lE 'worktrees/main|통합 워크트리|integration worktree|main을 잡은' -- . ':(exclude)*_test.go' ':(exclude).moai/specs' ':(exclude).moai/reports' ':(exclude)docs-site'
.claude/agents/harness/hns-release-specialist.md
.claude/agents/moai/manager-git.md
.claude/rules/local/gitflow-lane-protocol.md
.claude/rules/moai/workflow/delivery-policy.md
.claude/rules/moai/workflow/spec-workflow.md
.claude/rules/moai/workflow/worktree-integration.md
.claude/skills/moai-workflow-worktree/modules/parallel-advanced.md
.claude/skills/moai-workflow-worktree/modules/parallel-workflows.md
.claude/skills/moai/workflows/sync/delivery.md
.codex/agents/moai/manager-git.toml
.moai/docs/git-local-workflow-doctrine.md
.moai/docs/git-workflow-doctrine.md
.moai/docs/gitflow-integration-chain.md
.moai/policies/workflow/delivery-policy.md
.moai/policies/workflow/factory-dispatch.md
.moai/policies/workflow/kanban-dispatch.md
.moai/policies/workflow/spec-workflow.md
.moai/policies/workflow/worktree-integration.md
.moai/workflows/sync/delivery.md
AGENTS.local.md
CHANGELOG.md
internal/cli/cc.go
internal/cli/factory_card.go
internal/cli/factory_merge.go
internal/cli/integration.go
internal/cli/integration_merge.go
internal/cli/session_worktree_automerge.go
internal/cli/testdata/codex-rollouts-t1171/real/rollout-2026-09-24T18-41-16-01a0d2ca-6afa-7553-ad97-70d3de5e0931.jsonl
internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-git.toml
internal/cli/testdata/codex-rollouts-t1171/roles/manager-git.toml
internal/cli/testdata/codex-rollouts-t1171/synthetic/n4-crossed-roles.jsonl
internal/cli/worktree_branch_flag.go
internal/factory/candidate_landing_check.go
internal/factory/integration_merge_collision.go
internal/factory/integration_merge_step.go
internal/mission/git_owner.go
internal/template/templates/.claude/agents/moai/manager-git.md
internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
internal/template/templates/.claude/skills/moai-workflow-worktree/modules/parallel-advanced.md
internal/template/templates/.claude/skills/moai-workflow-worktree/modules/parallel-workflows.md
internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
internal/template/templates/.codex/agents/moai/manager-git.toml
internal/template/testdata/binding_ledger.json
reports/hooks-audit-20260911-01a08e35/cards-after.json
reports/hooks-audit-20260911-01a08e35/cards-before.json
exit=0
```
(The 47-line count was taken with `wc -l` on the redirected output: 47. Classification of every line: plan.md §B.5.)

### G2 — which tree holds `main`
```
$ git for-each-ref refs/heads/main '--format=%(refname) %(objectname) %(worktreepath)'
refs/heads/main d2d9ac08fb1899440da911121ddf13eecc806a3f /Users/goos/MoAI/moai-adk-go/.claude/worktrees/main
exit=0
```
The primary checkout (`/Users/goos/MoAI/moai-adk-go`) was observed detached at `478285ab4c1fc893b020bb464d14cb69c8422523` in the first entry of `git worktree list --porcelain` (read once at plan time; the rule's layout is the operator's, not a repository setting).

### G3 — local `main` moved after the tree pin; no in-scope file moved with it
```
$ git rev-list --count 7faf75fdf..main
12
exit=0
$ git diff --stat 7faf75fdf main -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/delivery-policy.md .claude/rules/moai/workflow/spec-workflow.md .claude/rules/moai/workflow/worktree-integration.md .claude/rules/moai/workflow/worktree-integration-ops.md .claude/skills/moai/workflows/sync/delivery.md .codex/agents/moai/manager-git.toml
(empty)
exit=0
```

### G4 — tool provenance (which build judged the tree)
The installed `moai` prints version `v3.2.0-rc.36-t1822c` (`moai version`); the `moai` MCP server banner of this session names the same build with commit `ee981a932`. The commit of the CLI binary itself was not read separately (a Gap, plan.md §F). The three commands below show that commit is a strict ancestor of the tree HEAD and that `internal/spec` did not change in that range, so the SPEC lint rules that build applies are those of HEAD.
```
$ git merge-base --is-ancestor ee981a932 HEAD
exit=0
$ git rev-list --count ee981a932..HEAD
243
exit=0
$ git log --oneline ee981a932..HEAD -- internal/spec
(empty)
exit=0
```

### G5 — mirror baselines (template source vs deployed copy)
```
$ diff -q .claude/agents/moai/manager-git.md internal/template/templates/.claude/agents/moai/manager-git.md
(empty)  exit=0
$ diff -q .codex/agents/moai/manager-git.toml internal/template/templates/.codex/agents/moai/manager-git.toml
(empty)  exit=0
$ diff -rq .claude/skills/moai/workflows/sync internal/template/templates/.claude/skills/moai/workflows/sync
(empty)  exit=0
$ diff -rq .claude/rules/moai/workflow internal/template/templates/.claude/rules/moai/workflow
Files .claude/rules/moai/workflow/contract-sign-guard.md and internal/template/templates/.claude/rules/moai/workflow/contract-sign-guard.md differ
Files .claude/rules/moai/workflow/cross-session-messaging-detail.md and internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md differ
Files .claude/rules/moai/workflow/factory-dispatch.md and internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md differ
Files .claude/rules/moai/workflow/moai-memory.md and internal/template/templates/.claude/rules/moai/workflow/moai-memory.md differ
Files .claude/rules/moai/workflow/orchestration-mode-selection.md and internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md differ
Files .claude/rules/moai/workflow/worktree-integration-ops.md and internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md differ
exit=1
$ git diff --no-index --numstat .claude/rules/moai/workflow/worktree-integration-ops.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md
9	9	{.claude => internal/template/templates/.claude}/rules/moai/workflow/worktree-integration-ops.md
exit=1
```
Consequence: five of the six source/deployed pairs this work edits (`manager-git.md`, `delivery-policy.md`, `spec-workflow.md`, `worktree-integration.md`, `sync/delivery.md`) and the generated toml are byte-identical today; the ops pair is a sanitized pair differing in nine lines (card ids and dates scrubbed in the template copy, outside the section this work edits). The agents and codex directory pairs differ in other files (seven agents; eight tomls plus one toml present only in the template); this work touches none of those.

### G6 — the one verb that resyncs a main
```
$ git grep -noF -e "Short: \"Fast-forward the primary checkout's local main to origin/main" -- internal/cli/local_main_resync.go
internal/cli/local_main_resync.go:91:Short: "Fast-forward the primary checkout's local main to origin/main
exit=0
```
`moai integration --help` lists the verbs `status`, `acquire`, `release`, `policy`, `remeasure`, `merge`, `preflight`, `resync`, `candidate`; none moves a stale primary checkout to the local `main` tip.

### G7 — `.moai/policies` and `.moai/workflows` are deployer output
```
$ git grep -nF -e '.moai/policies ← .claude/rules/moai' -- internal/cli/update_reconcile.go internal/cli/migrate_render.go
internal/cli/migrate_render.go:36:// (.moai/policies ← .claude/rules/moai, .moai/workflows ← .claude/skills/
internal/cli/update_reconcile.go:71:	// The harness-neutral shared surfaces (.moai/policies ← .claude/rules/moai,
exit=0
```
`.moai/policies/README.md` (first lines, read) says the deployer projects the Claude rule sources into that path. The four candidates differ from their deployed sources today:
```
$ git diff --no-index --numstat .claude/rules/moai/workflow/delivery-policy.md .moai/policies/workflow/delivery-policy.md
(empty)  exit=0
$ git diff --no-index --numstat .claude/rules/moai/workflow/spec-workflow.md .moai/policies/workflow/spec-workflow.md
87	11	{.claude/rules/moai => .moai/policies}/workflow/spec-workflow.md
exit=1
$ git diff --no-index --numstat .claude/rules/moai/workflow/worktree-integration.md .moai/policies/workflow/worktree-integration.md
73	8	{.claude/rules/moai => .moai/policies}/workflow/worktree-integration.md
exit=1
$ git diff --no-index --numstat .claude/skills/moai/workflows/sync/delivery.md .moai/workflows/sync/delivery.md
34	34	{.claude/skills/moai => .moai}/workflows/sync/delivery.md
exit=1
```

### G8 — Go and configuration citations used in spec.md §A.1
```
$ git grep -noF -e 'func integrationMergeWorktree' -e 'the primary never changes branch' -- internal/cli/integration_merge.go
internal/cli/integration_merge.go:145:func integrationMergeWorktree
internal/cli/integration_merge.go:155:the primary never changes branch
exit=0
$ git grep -noF -e 'The re-sync applies to the primary checkout only' -- internal/cli/local_main_resync.go
internal/cli/local_main_resync.go:361:The re-sync applies to the primary checkout only
exit=0
$ git grep -nF -e 'if !primary {' -e 'integrationOverlapRefusal(git, pinned)' -- internal/factory/integration_merge_step.go
internal/factory/integration_merge_step.go:309:	if !primary {
internal/factory/integration_merge_step.go:310:		reason, err := integrationOverlapRefusal(git, pinned)
internal/factory/integration_merge_step.go:536:			if !primary {
internal/factory/integration_merge_step.go:776:		if !primary {
exit=0
$ git grep -nF -e 'base := config.LoadWorktreeBaseBranch(projectRoot)' -e 'if base != "" && !sessionWorktreeBaseResolvable(base) {' -- internal/cli/session_worktree.go
internal/cli/session_worktree.go:236:	base := config.LoadWorktreeBaseBranch(projectRoot)
internal/cli/session_worktree.go:237:	if base != "" && !sessionWorktreeBaseResolvable(base) {
exit=0
$ git grep -nF -A1 -e '    local_main_integration:' -- .moai/config/sections/workflow.yaml
.moai/config/sections/workflow.yaml:194:    local_main_integration:
.moai/config/sections/workflow.yaml-195-        enabled: true
exit=0
$ git grep -noF -e 'worktree_base_branch: main' -- .moai/config/sections/git-strategy.yaml
.moai/config/sections/git-strategy.yaml:5:worktree_base_branch: main
exit=0
```
Not a ledger command: the Claude Code `--worktree` base is documented at `.claude/rules/moai/workflow/worktree-integration-ops.md:277` ("Native worktrees ... branch from the repository's default branch (`origin/HEAD`) by default"), read at plan time.

### G9 — the Template-First chain names exist (not run in plan phase)
```
$ git grep -nE '^func (TestRuleTemplateMirrorDrift|TestSanitizedPairParity|TestTemplateNoInternalContentLeak|TestRouteBPrecheckCoversAllEntryPoints|TestWorkflowRulePathsPinned|TestDeclaredRuleMirrorForks)\(' -- 'internal/template/*_test.go'
internal/template/internal_content_leak_test.go:1567:func TestTemplateNoInternalContentLeak(t *testing.T) {
internal/template/route_b_precheck_test.go:26:func TestRouteBPrecheckCoversAllEntryPoints(t *testing.T) {
internal/template/rule_template_mirror_test.go:99:func TestDeclaredRuleMirrorForks(t *testing.T) {
internal/template/rule_template_mirror_test.go:378:func TestRuleTemplateMirrorDrift(t *testing.T) {
internal/template/sanitized_pair_parity_test.go:189:func TestSanitizedPairParity(t *testing.T) {
internal/template/workflow_rule_paths_pinned_test.go:24:func TestWorkflowRulePathsPinned(t *testing.T) {
exit=0
$ git grep -nE '^(agents-emit|agents-emit-check|embed-manifest-check|build):' -- Makefile
Makefile:34:build: agents-emit-check commands-emit-check tool-policy-drift-check embed-manifest-check templ-generate ## Build the binary
Makefile:38:agents-emit: ## Regenerate the .codex/agents/moai TOMLs from the neutral .md layer
Makefile:47:agents-emit-check: ## Verify the committed .codex TOMLs match the .md source layer (read-only; never regenerates)
Makefile:72:embed-manifest-check: ## Verify the committed embed manifest matches the tracked template set and the templates/ tree (read-only; never regenerates)
exit=0
```

### L1 — RED-now for AC-LSR-001 and AC-LSR-002 (protocol removal anchors)
```
$ git grep -noF -e '## 2. Integration surface — the primary checkout' -e 'The landing surface is the **primary checkout**' -e 'When no separate integration worktree holds `main`, the surface is the primary' -e '# Inside the tree whose HEAD is local main (the primary checkout)' -e 'The tool takes the integration window, confirms that HEAD points at `main` and that the primary checkout is clean' -- .claude/rules/local/gitflow-lane-protocol.md
.claude/rules/local/gitflow-lane-protocol.md:33:## 2. Integration surface — the primary checkout
.claude/rules/local/gitflow-lane-protocol.md:35:The landing surface is the **primary checkout**
.claude/rules/local/gitflow-lane-protocol.md:43:When no separate integration worktree holds `main`, the surface is the primary
.claude/rules/local/gitflow-lane-protocol.md:137:# Inside the tree whose HEAD is local main (the primary checkout)
.claude/rules/local/gitflow-lane-protocol.md:178:The tool takes the integration window, confirms that HEAD points at `main` and that the primary checkout is clean
exit=0
```
Red because the five sentences (heading plus four body lines) that make the primary checkout the landing, build or update tree are all present; green means exit 1 with empty stdout.

### L2 — RED-now for AC-LSR-003 and AC-LSR-004 (AGENTS.local.md removal anchors)
```
$ git grep -noF -e 'Card work lands on local `main` in the primary checkout' -e 'the HEAD of the primary checkout points at `main`, and no separate integration worktree holds `main`' -e '(3) confirms that the primary checkout is clean (item 3)' -e 'To land from the primary checkout' -e '(1) Take the window, and confirm that HEAD points at `main` and that the primary checkout is clean.' -- AGENTS.local.md
AGENTS.local.md:184:Card work lands on local `main` in the primary checkout
AGENTS.local.md:192:the HEAD of the primary checkout points at `main`, and no separate integration worktree holds `main`
AGENTS.local.md:194:(3) confirms that the primary checkout is clean (item 3)
AGENTS.local.md:198:To land from the primary checkout
AGENTS.local.md:204:(1) Take the window, and confirm that HEAD points at `main` and that the primary checkout is clean.
exit=0
```
Red for the same reason as L1: five sentences present. (The first draft of this cell had four; the plan audit found lines 33 and 194 missing, and this re-recording at HEAD `02a4b613a` returns five per file, ten in all.)

### L3 — RED-now for every presence clause (AC-LSR-001, 003, 005, 007, 009, 010, 011, 012, 013)
```
$ git grep -noF -e 'landing and resync only' -e 'is the human and leader workspace and does not check out' -e 'git show main:' -e 'smudge' -e 'planned in t1847' -e 'branches from the named ref' -e 'named local ref' -e 'no moai verb moves the primary checkout to the current local' -e 'each untracked or modified file' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/moai/manager-git.md .claude/rules/moai/workflow .claude/skills/moai/workflows/sync internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/sync .codex/agents/moai/manager-git.toml internal/template/templates/.codex/agents/moai/manager-git.toml
(empty)
exit=1
```
Red because none of the nine anchor phrases exists in any file the work will edit (the swept path set is the edit set, with the generated tomls). Why these anchors and not shorter ones: the bare phrase `does not check out` already occurs in protocol line 35 ("A lane does not check out `main` in its own card tree"), which would have made a presence clause green at arrival; `is the human and leader workspace and does not check out` is unique.

### L4 — guarantee-word guard baseline and positive control (AC-LSR-014)
```
$ git diff --name-only -G'race-safe|guaranteed|always safe' 7faf75fdf HEAD -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/moai/manager-git.md .claude/rules/moai/workflow .claude/skills/moai/workflows/sync internal/template/templates/.claude/agents/moai internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/sync .codex
(empty)
exit=0
```
Pre-existing uses are not changed lines and so do not appear: `git grep -cE 'race-safe|guaranteed|always safe'` over the edit set finds `.claude/rules/local/gitflow-lane-protocol.md:1` (line 108, "routing is not guaranteed either") and `.claude/rules/moai/workflow/worktree-integration-ops.md:1` (line 248, "Refusal is not guaranteed"), both negations.

### L4b — positive control for the `-G` form
```
$ git diff --no-index --name-only '-G(SPEC-[A-Z]|[^a-z]t[0-9]{3,4}[^0-9]|20[0-9]{2}-[0-9]{2}-[0-9]{2}|planned in)' .claude/rules/moai/workflow/worktree-integration-ops.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md
internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md
exit=1
$ git diff --name-only '-G(SPEC-[A-Z]|[^a-z]t[0-9]{3,4}[^0-9]|20[0-9]{2}-[0-9]{2}-[0-9]{2}|planned in)' 7faf75fdf HEAD -- internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/sync
(empty)
exit=0
```
The first command (exit 1 is `git diff --no-index` reporting a difference) proves the regex fires on lines that carry card ids and dates; the second is the AC-LSR-006 baseline.

### L5 — existing statements the work must keep (guard halves)
```
$ git grep -noF -e 'Branches from default remote branch' -- .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
.claude/rules/moai/workflow/worktree-integration.md:92:Branches from default remote branch
internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md:92:Branches from default remote branch
exit=0
```
(L5a.) L5b:
```
$ git grep -noF -e 'point to the same commit' -- .claude/rules/local/gitflow-lane-protocol.md
.claude/rules/local/gitflow-lane-protocol.md:27:point to the same commit
exit=0
```

### L6 — "planned in t1847" absent from distributed and deployed surfaces (AC-LSR-013 guard half)
```
$ git grep -noF -e 'planned in t1847' -- internal/template/templates .claude/agents .claude/rules/moai .claude/skills .codex
(empty)
exit=1
```

### L7 — scope fence, allow-list form (AC-LSR-015; replaces the first draft's path-list form, which let `cmd/moai/main.go` and `.codex/hooks.json` edits through)
```
$ git diff --name-only 7faf75fdf HEAD -- . ':(exclude)AGENTS.local.md' ':(exclude).claude/rules/local/gitflow-lane-protocol.md' ':(exclude).claude/agents/moai/manager-git.md' ':(exclude).claude/rules/moai/workflow/delivery-policy.md' ':(exclude).claude/rules/moai/workflow/spec-workflow.md' ':(exclude).claude/rules/moai/workflow/worktree-integration.md' ':(exclude).claude/rules/moai/workflow/worktree-integration-ops.md' ':(exclude).claude/skills/moai/workflows/sync/delivery.md' ':(exclude).codex/agents/moai/manager-git.toml' ':(exclude)internal/template/templates/.claude/agents/moai/manager-git.md' ':(exclude)internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md' ':(exclude)internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md' ':(exclude)internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md' ':(exclude)internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md' ':(exclude)internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md' ':(exclude)internal/template/templates/.codex/agents/moai/manager-git.toml' ':(exclude)internal/template/catalog.yaml' ':(exclude).moai/specs/SPEC-LANDING-SURFACE-RULES-001' ':(exclude).moai/reports/t1846'
(empty)
exit=0
```
At the audited HEAD `02a4b613a` the only changes since the pin are the four SPEC files (`git diff --name-only 7faf75fdf HEAD` printed exactly `acceptance.md`, `plan.md`, `progress.md`, `spec.md` under the SPEC directory), which the list excludes, hence the empty stdout.

### L7c — controls for the allow-list form (scratch repository)
Script `scratchpad/allow/run.sh` (session scratchpad `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/f953d45c-4d83-4b73-ae50-3e27847f6232/scratchpad/allow/`) builds a throw-away repository holding the 17 listed files, the SPEC and report paths, and the forbidden paths `cmd/moai/main.go`, `.codex/hooks.json`, `internal/cli/p.go`, `.moai/policies/workflow/delivery-policy.md` and an unlisted sibling `.claude/rules/moai/workflow/other-rule.md`; it then commits edit sets and runs the same `git diff --name-only <pin> HEAD -- . ':(exclude)...'` form (the identical 19 exclusions as L7) after each:
```
== base == pin (nothing changed)
exit=0
== negative control: only the 17 allowed files + SPEC dir + report dir edited
exit=0
== positive control 1: add edits to cmd/moai/main.go and .codex/hooks.json
.codex/hooks.json
cmd/moai/main.go
exit=0
== positive control 2: add an Go file edit, a projection edit, an unlisted sibling rule
.claude/rules/moai/workflow/other-rule.md
.codex/hooks.json
.moai/policies/workflow/delivery-policy.md
cmd/moai/main.go
internal/cli/p.go
exit=0
```
The negative control (every legitimate file edited) prints nothing; the two mutants print exactly the forbidden paths. Exit is 0 in every case (`git diff --name-only` without `--exit-code`), so the stdout is the verdict.

L7b:
```
$ git grep -c -F -e 'agentic-trunk' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/moai/manager-git.md .claude/rules/moai/workflow .claude/skills/moai/workflows/sync internal/template/templates/.claude/agents/moai internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/sync .codex
(empty)
exit=1
```
Both L7 and L7b are green at arrival by design (guards): L7 is empty because the tree differs from the pin only in listed paths, L7b because the name does not occur in the edit set today. Their failing direction is the named mutant: for L7 the controls of L7c are recorded verbatim; for L7b an inserted flow name is observed at M5.

### G10 — what `make agents-emit` writes (D1)
```
$ git grep -nE 'templatesDir = |func committedTOMLPath|filepath.Join\(templatesDir, filepath.FromSlash\(emitted|os.WriteFile\(committedTOMLPath|update := os.Getenv' -- internal/template/agentemit/golden_test.go
internal/template/agentemit/golden_test.go:32:const templatesDir = "../templates"
internal/template/agentemit/golden_test.go:64:func committedTOMLPath(emitted string) string {
internal/template/agentemit/golden_test.go:65:	return filepath.Join(templatesDir, filepath.FromSlash(emitted))
internal/template/agentemit/golden_test.go:75:	update := os.Getenv("AGENTEMIT_UPDATE") == "1"
internal/template/agentemit/golden_test.go:90:			if err := os.WriteFile(committedTOMLPath(p), data, 0o644); err != nil {
exit=0
```
Reading (golden_test.go:60-140, read in full this session): with `AGENTEMIT_UPDATE=1` the test writes only the `.toml` artifacts, to `internal/template/templates/<emitted path>` (line 90 via 64-65); the neutral `.md` layer is read from the same template tree (`agentMDRoot = ".claude/agents/moai"`, `emitRealSet` over `os.DirFS(templatesDir)`) and is never written (the `.md` loop at the end only compares). Without the variable the test compares the emission with the committed template toml. Nothing under the repository-root `.codex/` is read or written by this test. `make agents-emit` is that command (`Makefile:38-39`); `make agents-emit-check` is the read-only comparison with `AGENTEMIT_UPDATE=` scrubbed (`Makefile:47-49`). The deployed `.codex/agents/moai/manager-git.toml` is therefore not touched by either target. The two toml copies are byte-identical today (G5), and `moai update` (`.moai/docs/local-dev-guide.md:44`, "Sync templates to current project") rewrites the whole managed template set rather than one file, so the narrow step is a plain `cp` of the regenerated template toml onto the deployed path — the same step the deployed `.md` copies need.

### G11 — which `catalog.yaml` entries the edit set hits (D2)
```
$ git grep -nE 'path: templates/.claude/(rules|skills/moai/|agents/moai/manager-git)' -- internal/template/catalog.yaml
internal/template/catalog.yaml:8:              path: templates/.claude/skills/moai/
internal/template/catalog.yaml:386:                  path: templates/.claude/agents/moai/manager-git.md
exit=0
$ git grep -c -E 'worktree-integration|delivery-policy|spec-workflow' -- internal/template/catalog.yaml
(empty)
exit=1
```
The `hash:` lines of those two entries are `catalog.yaml:9` (`moai` skill, whole directory tree, so `workflows/sync/delivery.md` is covered) and `catalog.yaml:387` (`manager-git` agent). The rule files (`delivery-policy`, `spec-workflow`, `worktree-integration`, `worktree-integration-ops`) and the `.codex` tomls have no catalog entry. `Makefile:34-35`: `build` depends on the check targets and then runs `go run ./internal/template/scripts/gen-catalog-hashes.go --all`, whose header says it updates every entry's hash in place ("Update every entry's hash field in catalog.yaml in-place"; skill entries hash the directory tree, agent entries the `*.md` file). Not executed in plan phase; whether `--all` also changes other entries (pre-existing drift) is a run-phase M1 measurement through its `--dry-run` flag.

### L3c — RED-now for the added anchors (D5, D9)
```
$ git grep -noF -e "falls back to the invoking tree" -e 'Where the integration branch is held by a dedicated worktree' -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/delivery-policy.md .claude/rules/moai/workflow/spec-workflow.md .claude/rules/moai/workflow/worktree-integration.md .claude/skills/moai/workflows/sync/delivery.md internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
(empty)
exit=1
```
Red because neither the fallback-limit phrase nor the template scope clause exists in any file the work edits; both flip at M2 (the limit, local files) and M3 (the clause, template sources).
