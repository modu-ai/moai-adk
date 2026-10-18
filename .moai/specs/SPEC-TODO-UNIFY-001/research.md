# Research: SPEC-TODO-UNIFY-001

Measured at tree `026f48658f` (branch `WT-todo-unify`, clean) on 2026-10-18. Every figure below names the command that produced it. A figure that cannot be re-run on the current tree is marked.

## §R.1 Provenance and refusals

- **Revision 4 provenance (post-ceiling correction, audit iteration 3).** HEAD `f40740ea61` on branch `WT-todo-unify`. The judging build for the revision-4 `moai spec lint` runs is a binary built from this tree by plain `go build -o <scratchpad>/moai-t1641-iter4 ./cmd/moai` (exit 0, no Makefile LDFLAGS, so the tree HEAD above, not the version stamp, is its identity). The measurements new in this revision are ledger rows E-24 to E-26 and §R.16. Refusals this round are recorded in the card's progress record, `.moai/reports/t1641/progress.md`.
- **Tree.** `git rev-parse --short HEAD` printed `026f48658f`; `git branch --show-current` printed `WT-todo-unify`. Revision 2 (audit iteration 1 FAIL) measured at HEAD `23e40eb7c4`, which adds only plan artifacts and the audit report: `git diff --stat 026f48658f 23e40eb7c4 -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing. Ledger entries E-15 to E-18 and the §R.5b counts carry the `23e40eb7c4` pin.
- **Revision 3 provenance (audit iteration 2).** HEAD `3dc1052f6f` on branch `WT-todo-unify`, clean (`git rev-parse HEAD` printed `3dc1052f6faf0883ceebd747a8af5979c1508d93`, `git status --short` printed nothing). `git diff --stat 23e40eb7c4 3dc1052f6f -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing, exit 0, so every measurement pinned `23e40eb7c4` or `026f48658f` outside the SPEC directory and the card's reports holds at `3dc1052f6f`. The judging build for the revision-3 `moai` observations (the help-page verb counts E-19, E-20, and `moai spec lint`) is a binary built from this tree by plain `go build -o <scratchpad>/moai-t1641-iter3 ./cmd/moai` (exit 0); it carries no Makefile LDFLAGS, so its version stamp (`moai-adk v3.1.3`) is the unstamped default and the tree HEAD above, not the stamp, is its identity. `moai decision read` ran on the installed build `ee981a932` (a strict ancestor 516 commits behind at revision 2; the lag was not re-measured this round), which is acceptable because it only reads the board.
- **Revision-3 refusals (Gaps by verification-claim-integrity §3.1).** The worktree guard refused three commands this round, all by the same reading ("too complex to verify that it stays inside the worktree"): a `for` loop over four commit ids calling `git rev-list`, `python3 -I <script> $SP/...` with a shell variable in the path, and `cat <path>/$(git -C ... rev-parse ...).json` (a command substitution holding git). What was done instead: one plain `git -C <own absolute path>` command per commit, literal absolute paths for the Python script, and two separate commands (resolve the tree, then read the file). Every other revision-3 command ran as written.
- **Revision-2 refusals (Gaps by verification-claim-integrity §3.1).** The worktree guard refused two more commands this round: a `python3 -c` program that called `git` after a `cd` (too complex to verify), and a `git grep ... --include=*.go` (a glob pattern that expands at runtime). What was done instead: a Python script file in the scratchpad that takes the tree path as an argument and calls `git -C`, and a plain `git grep -F ... -- internal`. Every other revision-2 command ran as written.
- **Judging build for `moai` observations.** The installed binary reports `moai-adk v3.2.0-rc.36-t1822c` with build id `gee981a932-dirty`, built 2026-10-15T21:56:00Z. `git merge-base --is-ancestor ee981a932 HEAD` exited 0 and `git rev-list --count ee981a932..HEAD` printed `516`: the installed build is a strict ancestor of the tree, 516 commits behind. Observations E-09 and E-10 therefore describe that build, not this tree; each is paired with a source reading of this tree (E-04, E-05) so the claim does not rest on the older build alone.
- **Refused commands (Gaps by verification-claim-integrity §3.1).** The worktree-isolation guard refused six commands: two shell loops over `git grep`, three compound lines that joined `git` with an assignment, a redirect, or `;`, and `go list -f '{{.ImportPath}} {{join .Imports " "}}' ...` (the reverse-import query). What was done instead: plain single commands (the ledger below), and `git grep -l` of each import path for the importer lists in `plan.md` §F. Consequence: the importer lists are non-test importers found by text search and may miss an importer that aliases the import path.
- **Not run by design.** No `go test` and no heavy-slot use (card constraint). `moai todo done` was not run, to avoid the one failure mode in which a lane-guard false negative would archive the live card; the lane refusal of `done` is established by reading `todo.go` (`todoLaneReadOnlyVerbs` lists `list`, `history`, `show`, `why`, `pr`, `triage`, `trace`; `PersistentPreRunE` refuses everything else for a lane session).

## §R.2 Evidence ledger

Single-invocation, read-only commands. Tree `026f48658f` unless the row says otherwise.

| Id | Command | Exit | Verbatim stdout |
|----|---------|------|-----------------|
| E-01 | `git grep -I -l 'moai gtd' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` | 0 | the 73 paths of §R.4 |
| E-02 | `git grep -n -E '"T18"\|isReservedEdge\|ErrReservedEdge' -- internal/homestate/card_transition.go` | 0 | see below (6 lines) |
| E-03 | `git grep -n -E 'moai todo (capture\|clarify\|organize\|reflect\|engage\|answer)' -- .claude/skills/moai/workflows/gtd.md` | 0 | see below (6 lines) |
| E-04 | `git grep -n -c 'newGTDReflectCmd' -- internal/cli/todo.go` | 1 | (empty) |
| E-05 | `git grep -n -i 'deprecat' -- internal/cli/gtd.go` | 1 | (empty) |
| E-06 | `git grep -n -E 'Use: +"spec ' -- internal/cli` | 1 | (empty) |
| E-07 | `git grep -n -w '"done":' -- internal/cli/todo.go` | 1 | (empty) |
| E-08 | `git grep -n 'gtdErr != todoErr' -- internal/cli/gtd_compat_test.go` | 0 | `internal/cli/gtd_compat_test.go:83:		if gtdExit != nil \|\| todoExit != nil \|\| gtdErr != todoErr \|\| gtdOut != todoOut {` |
| E-09 | `moai todo reflect` (installed build `ee981a932-dirty`) | 1 | `ERROR` block: `Unknown command "reflect" for "moai todo".` (the stream it was written to was not recorded) |
| E-10 | `moai gtd list --limit 1` (installed build) | 0 | 8 stdout lines; stderr 2 lines (`todo: answered by the SQLite backlog store; ...` and `list: 160 rows withheld — showing 1 of 161 ...`); no deprecation text |
| E-11 | `wc -c AGENTS.md` | 0 | `   23780 AGENTS.md` |
| E-12 | `git grep -n -E 'len\(accepted\) != 71' -- internal/homestate/fr_transition_test.go` | 0 | `internal/homestate/fr_transition_test.go:166:	if len(accepted) != 71 \|\| refused != 370 {` |
| E-13 | `git grep -n -E 'TestGTDCanonicalSurfaceGolden\|arguments: `gtd` \$ARGUMENTS' -- internal/template/gtd_canonical_surface_test.go` | 0 | `...:10:func TestGTDCanonicalSurfaceGolden(t *testing.T) {` and `...:23:		if strings.Contains(path, "todo") && !strings.Contains(string(data), "arguments: `gtd` $ARGUMENTS") {` |
| E-14 | `git grep -n -E "'gtd'\|moai gtd" -- mods/moai-board/hooks/data.ts scripts/cutover-precheck.sh internal/cli/factory_card.go` | 0 | see below (8 lines) |
| E-15 | `git grep -n -E 'renamed to moai todo' -- internal/cli` (HEAD `3dc1052f6f`; revision 3 widened the pathspec from `internal/cli/gtd.go` to the whole package, audit D24, so the proxy flips wherever the phrase lands; the narrow form was re-run too and printed exit 1) | 1 (observed) | (empty) |
| E-16 | `git grep -n -c 'fully supported compatibility alias' -- AGENTS.md internal/template/templates/AGENTS.md.tmpl` (HEAD `3dc1052f6f`) | 0 (observed) | `AGENTS.md:1` and `internal/template/templates/AGENTS.md.tmpl:1` (lines `AGENTS.md:344` and `AGENTS.md.tmpl:343`) |
| E-17 | `git grep -n -E '"T29"' -- internal/homestate/card_transition.go` (HEAD `3dc1052f6f`) | 1 (observed) | (empty) |
| E-18 | `git grep -n -E 'ApprovalIssuerAuto\|IssuerAuto\|RevokeLeaderApproval' -- internal/homestate internal/cli` (HEAD `3dc1052f6f`) | 1 (observed) | (empty) |
| E-19 | `moai todo --help` from the binary built from this tree (HEAD `3dc1052f6f`), COMMANDS block counted by a script (`python3 -I count_verbs.py <help file>`); a measurement, not a single-invocation release-blocking cell | 0 (observed for the help command) | `26 add list done undone next claim unpick edit move drop undrop merge hold unhold analyze relate unrelate why trace pr landed auto-done export-json history show triage` |
| E-20 | `moai gtd --help`, same binary and script | 0 (observed) | `32 add list done undone next claim unpick edit move drop undrop merge hold unhold analyze relate unrelate why trace pr landed auto-done export-json history show triage capture clarify organize reflect engage answer` |
| E-21 | `git grep -n -E '"T29"\|AutoIssue\|ApprovalIssuerAuto\|IssuerAuto\|RevokeLeaderApproval\|VerifyRemeasureBinding' -- internal/homestate internal/cli internal/factory` (HEAD `3dc1052f6f`): none of the revision-3 design identifiers exists yet | 1 (observed) | (empty) |
| E-22 | `git grep -n -w -F 'Run:' -- internal/cli/todo.go internal/cli/todo_*.go internal/cli/gtd.go internal/cli/gtd_answer.go ':!*_test.go'` (HEAD `3dc1052f6f`): no cobra command of the two trees sets `Run`, so wrapping `RunE` covers every handler (`design.md` §D.2). A first version of this check, `-E '^\s+Run:\s'`, only finds a line-leading `Run:` and would miss an inline `Run:` in the one-line composite literals of `gtd.go`; it was replaced by the whole-word form | 0 (observed) | `internal/cli/todo.go:1534:	q := factory.GitLandedQuerier{Run: todoRunCommand, Ref: ref}` and `internal/cli/todo_pr.go:237:	landed := factory.GitLandedQuerier{Run: todoRunCommand, Ref: landedRef}` (a struct field of another type, not a cobra command); the companion `git grep -n -E '\.Run\s*=' -- <same files>` printed nothing, exit 1; the positive control `git grep -c -w -F 'RunE:' -- internal/cli/todo.go internal/cli/gtd.go internal/cli/gtd_answer.go` printed `gtd.go:5`, `gtd_answer.go:1`, `todo.go:6`, exit 0 |
| E-23 | `git grep -n -E 'MarkFlagRequired\|MarkFlagsMutuallyExclusive\|MarkFlagsRequiredTogether\|MarkFlagsOneRequired' -- internal/cli/todo.go internal/cli/todo_*.go internal/cli/gtd.go internal/cli/gtd_answer.go ':!*_test.go'` (HEAD `3dc1052f6f`) | 0 (observed) | `internal/cli/gtd.go:98: _ = cmd.MarkFlagRequired("event")`, `:121 ("disposition")`, `:158 ("class")`, and `internal/cli/todo.go:356: cmd.MarkFlagsMutuallyExclusive("auto", "auto-lane", "auto-leader")` |
| E-24 | `git grep -n -F 'config.LoadGitFlowIntegrationConfig(root).IntegrationTarget' -- internal/cli/factory_card.go` (HEAD `f40740ea61`): the integration-branch sources in the push gate file | 0 (observed, exit captured by script) | `internal/cli/factory_card.go:2553:` (inside `factoryResolveIntegrationBranch`) and `internal/cli/factory_card.go:3238:	integration := config.LoadGitFlowIntegrationConfig(root).IntegrationTarget` (`factoryDecideCards`, the push gate path, the only value it forwards) |
| E-25 | `git grep -n -E 'INTO leader_approvals\|UPDATE leader_approvals' -- internal cmd ':!*_test.go'`; `git grep -n -F 'IssueLeaderApproval(' -- internal cmd ':!*_test.go'`; `git grep -n -E 'Issuer: *"auto"' -- internal cmd ':!*_test.go'` (HEAD `f40740ea61`): writers of the receipt table and any tracked writer of issuer `auto` | 0, 0, 1 (observed, exit captured by script) | `internal/homestate/card_transition.go:849` (the version re-stamp, `SET factory_version=?`) and `internal/homestate/leader_approval.go:173` (the insert); `internal/cli/leader_approval_gate.go:350` (the one caller) and `internal/homestate/leader_approval.go:149` (the definition); the third printed nothing |
| E-26 | `git grep -n -F 'TrimSpace(string(out))' -- internal/homestate/card_evidence.go`; `git grep -n -E -e '--no-replace-objects\|GIT_NO_REPLACE_OBJECTS' -- internal cmd ':!*_test.go'`; `git grep -c -F gitReadRaw -- internal` (HEAD `f40740ea61`) | 0, 0, 1 (observed, exit captured by script) | `internal/homestate/card_evidence.go:33:	return strings.TrimSpace(string(out)), nil`; `internal/hook/commit_identity_guard.go:456:` (a skip-list of git global options in a command parser, not a use of the option); nothing, so the raw reader of `design.md` §D.5 does not exist yet |

E-14 stdout, verbatim:

```
internal/cli/factory_card.go:1320:		return factoryRefusal(factoryRefuseHeld, "the card is held (moai gtd hold); only the operator releases it")
mods/moai-board/hooks/data.ts:36:  queueJson: ['moai', 'gtd', 'list', '--json'],
mods/moai-board/hooks/data.ts:37:  queueText: ['moai', 'gtd', 'list', '--limit', '0'],
mods/moai-board/hooks/data.ts:57:  isCardId(id) && isPrefixOk(prefix) ? ['moai', 'gtd', 'next', id, '--expect', prefix] : undefined
mods/moai-board/hooks/data.ts:125:export const SRC_QUEUE: Source = { cmd: 'moai gtd list', what: 'the queue' }
mods/moai-board/hooks/data.ts:202:/** The `moai gtd list` text form: `id<TAB>state<TAB>[by=… <TAB>lease=… <TAB>]text`; continuation lines and the footer are not rows. */
scripts/cutover-precheck.sh:45:#   CUTOVER_GTD_LIST_CMD            [moai gtd list]
scripts/cutover-precheck.sh:83:GTD_CMD="${CUTOVER_GTD_LIST_CMD:-moai gtd list}"
```

E-02 stdout, verbatim:

```
internal/homestate/card_transition.go:19:	ErrReservedEdge        = errors.New("reserved card transition")
internal/homestate/card_transition.go:157:		{"T18", CardMergedLocal, CardDone, guardNoRemote},
internal/homestate/card_transition.go:203:// isReservedEdge — pushed → ci-green stays reserved: the CI verdict reader
internal/homestate/card_transition.go:207:func isReservedEdge(from, to string) bool {
internal/homestate/card_transition.go:333:	if isReservedEdge(cur.State, req.To) {
internal/homestate/card_transition.go:334:		return Card{}, fmt.Errorf("%w: %s → %s is reserved; the CI verdict reader that admits it is owned by F3", ErrReservedEdge, cur.State, req.To)
```

E-03 stdout, verbatim:

```
.claude/skills/moai/workflows/gtd.md:150:`moai todo capture <text>` captures an inbox item without publishing a card.
.claude/skills/moai/workflows/gtd.md:161:`moai todo clarify <gtd-id>` clarifies outcome, evidence, trust, and authority.
.claude/skills/moai/workflows/gtd.md:173:`moai todo organize <gtd-id>` organizes a clarified item and its relationships.
.claude/skills/moai/workflows/gtd.md:185:`moai todo reflect` reviews blockers, stale evidence, and missing next actions.
.claude/skills/moai/workflows/gtd.md:194:`moai todo engage <gtd-id>` publishes, picks, and dispatches an approved
.claude/skills/moai/workflows/gtd.md:211:- `moai todo answer <t-id> <text>` answers a gate-blocked card; it is not a stage.
```

Reading of E-15 to E-18, E-21, E-22: revision 2 recorded the exit field of E-15, E-17, E-18 as git's contract because the tool then printed no status for an empty `git grep` (a Gap). Revision 3 re-ran each with `echo "exit=$?"` appended in the same call and observed exit 1 for E-15, E-17, E-18, E-21 and exit 0 for E-16, E-22, E-23, so that Gap is closed for these entries. They are absence-of-code observations of the same kind as E-04 to E-07; E-15 is the proxy that flips when the warning's lead phrase (REQ-TU-002) lands anywhere under `internal/cli`, and E-16 flips when the wording test's phrase replaces the old one. E-19 and E-20 baseline the verb count of AC-TU-004 (26 and 32; the expected 33 and 33 add `spec` and the six stage verbs, `design.md` §D.4).

Reading of E-04 to E-07: each is an absence-of-code observation. It characterizes the pre-implementation tree; it is not the criterion. The criteria are the tests named in `acceptance.md`, whose RED is observed at the first step of the owning milestone.

### Observation made in a lane session (not a ledger command)

`moai gtd reflect --json` in this session exited 1 with `Moai reflect: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage)`. `reflect` is a read-only review verb, yet the lane guard refuses it because it is on neither read-only list. This is the measured input behind REQ-TU-006.

## §R.3 The card text and the absorbed cards (verbatim quotes)

Card t1641 (queue record; the text is Korean and is quoted unmodified):

> [운영자 지시 10-17: 3.3 이월 해제·지금 진행 — gtd→todo 이름 전면 통일] [3.2-2-4·P1·Class C·운영자 승인 10-10] todo 통일(D3) — 범위: gtd 하위 6개 정리·경고 별칭(한 릴리스 유예)·문서 치환·슬래시 정본, 레인 자기 카드 done 예외, T18/T19 제거. 흡수: t1602(gtd→todo 개명 완결)·t1590(gtd 6동사 안내-실행 불일치). 리더 후보(10-10): gtd_operation.go:583(P2). 계획: .moai/reports/moai-32-master-upgrade-plan-20261010.html §2 2단계 · 재설계 결정 3·4. 완료 판정: gtd 경고 후 위임 · 최근 10 착지 카드 done 자동. 추가(영실이 10-10): pick 이후 카드에 spec_id를 기록하는 동사 부재 — contract_report.go cardToSpecID가 큐 spec_id를 읽으나 설정 경로는 next --spec뿐(t1631은 직접 UPDATE로 임시 처리).

Card t1590 (dropped into t1641 on 2026-10-10) fixes direction A: register the six factories on the canonical tree; it records that direction B (restore the `gtd` spelling) is rejected because it reverses the card's purpose, and it asks that the repair "execute every guidance sub-verb" as a clause.

Card t1602 (dropped into t1641 on 2026-10-10) says: queue verbs become `moai todo`; "GTD-stage verbs (capture, clarify, organize, reflect, engage, answer) stay under `moai gtd` as a separate feature"; flip the alias direction in the skill header; `/moai:gtd` becomes `/moai:todo`; update the body-check script; refresh the neutrality guard and pinned tests.

Redesign decision D3 (report `moai-adk-32-cross-harness-redesign-20261010.html`, §10.3, untracked): `todo` is the only canonical name; of the six gtd sub-verbs, absorb the ones in use into `todo` and delete the rest; `moai gtd` and `/moai:gtd` stay for one release behind a "renamed to todo" warning alias, then go; documents, skills, and rules are substituted in bulk. Decision D4 (§7.2): landing to done becomes a machine transition by an observed remote landing (three observers); receipts are issued from evidence; a lane may close its own card; the git-flow path T18/T19 is removed together with the completion of the GitHub Flow transition.

The three sources disagree about the six verbs (card t1590 versus card t1602 versus decision D3); that disagreement is decision-index Q1, decided in revision 2.

## §R.4 The 73 paths of E-01 (verbatim stdout)

```
.claude/agents/moai/manager-lead.md
.claude/rules/moai/workflow/factory-dispatch-cards.md
.claude/rules/moai/workflow/factory-dispatch-detail.md
.claude/rules/moai/workflow/factory-dispatch.md
.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
.claude/skills/moai-factory-foreman/SKILL.md
.claude/skills/moai/SKILL.md
.claude/skills/moai/workflows/gtd.md
.claude/skills/moai/workflows/project/doc-generation.md
.codex/agents/moai/manager-lead.toml
.moai/docs/todo-queue-storage.md
.moai/policies/workflow/factory-dispatch-cards.md
.moai/policies/workflow/factory-dispatch-detail.md
.moai/policies/workflow/factory-dispatch.md
.moai/policies/workflow/kanban-dispatch-detail.md
.moai/policies/workflow/kanban-dispatch.md
.moai/policies/workflow/main-checkout-branch-guard-detail.md
.moai/project/codemaps/data-flow.md
.moai/workflows/gtd.md
.moai/workflows/project/doc-generation.md
AGENTS.md
README.ja.md
README.ko.md
README.md
README.zh.md
docs-site/content/en/utility-commands/_meta.yaml
docs-site/content/en/utility-commands/moai-goal.md
docs-site/content/en/utility-commands/moai-gtd.md
docs-site/content/en/utility-commands/moai-todo.md
docs-site/content/ja/utility-commands/_meta.yaml
docs-site/content/ja/utility-commands/moai-goal.md
docs-site/content/ja/utility-commands/moai-gtd.md
docs-site/content/ja/utility-commands/moai-todo.md
docs-site/content/ko/utility-commands/_meta.yaml
docs-site/content/ko/utility-commands/moai-goal.md
docs-site/content/ko/utility-commands/moai-gtd.md
docs-site/content/ko/utility-commands/moai-todo.md
docs-site/content/zh/utility-commands/_meta.yaml
docs-site/content/zh/utility-commands/moai-goal.md
docs-site/content/zh/utility-commands/moai-gtd.md
docs-site/content/zh/utility-commands/moai-todo.md
docs-site/data/menu/main.yaml
internal/cli/factory_card.go
internal/cli/factory_dispatch_test.go
internal/cli/gtd_answer.go
internal/cli/gtd_answer_test.go
internal/cli/gtd_compat_test.go
internal/cli/testdata/codex-rollouts-t1171/real/rollout-2026-09-24T18-41-35-01a0d2ca-b7e2-7e50-8f7f-0ab053beb07f.jsonl
internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-lead.toml
internal/cli/testdata/codex-rollouts-t1171/roles/manager-lead.toml
internal/cli/todo.go
internal/cli/todo_test.go
internal/cli/todo_transition_stamps_test.go
internal/template/templates/.claude/agents/moai/manager-lead.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
internal/template/templates/.claude/skills/moai/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/gtd.md
internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md
internal/template/templates/.codex/agents/moai/manager-lead.toml
internal/template/templates/.moai/docs/todo-queue-storage.md
internal/template/templates/AGENTS.md.tmpl
internal/template/testdata/binding_ledger.json
mods/moai-board/README.md
mods/moai-board/hooks/data.ts
mods/moai-board/tests/board.test.tsx
mods/moai-board/tests/engine.test.ts
mods/moai-board/tests/pure/data.spec.ts
mods/moai-board/tests/support.ts
scripts/cutover-precheck.sh
```

## §R.5 How 385 became 73 (counts, each with its command)

The card's figure of 385 files comes from `git grep -il gtd -- . ':!.moai/reports' ':!CHANGELOG.md'`. Re-measured, one command per row:

| Step | Command | Count |
|------|---------|-------|
| All paths containing "gtd" (case-insensitive), reports and changelog excluded | `git grep -il gtd -- . ':!.moai/reports' ':!CHANGELOG.md'` | 385 |
| Text files only (binary excluded) | the same with `-I` | 371 (14 are binary images that match by bytes) |
| Outside the historical SPEC directories | the same with `-I` and `':!.moai/specs'` added | 204 (so 167 are historical SPEC files) |
| Carrying the command spelling `moai gtd` | `git grep -I -l 'moai gtd' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` | 73 files, 255 lines |
| Carrying the slash spelling `/moai gtd` | the same pattern `'/moai gtd'` | 30 files |

The counts were taken with `wc -l` over the command output (a pipe), so they are measurements for planning, not release-blocking evidence cells; E-01 is the single-invocation form of the 73.

Area breakdown of the 218 non-SPEC paths from the same query without `-I` (so it includes 14 binary matches, among them `docs-site/static` and `assets` images): `internal/cli` 43, `internal/template` 35, `docs-site/content` 35, `internal/factory` 23, `.claude` 18, `.moai/policies` 11, `mods` 6, `.moai/project` 6, `internal/mission` 5, `docs-site/static` 5, `docs-site/themes` 4, `internal/graph` 3, `.moai/workflows` 3, `.agents` 2, and one each in `AGENTS.md`, the four READMEs, `internal/hook`, `docs-site/data`, `.moai/manifest`, `.moai/docs`, `.gitignore`, `.codex`, `scripts`; the remainder are small groups.

### §R.5b Floors measured for the class A criteria (revision 2, tree `23e40eb7c4`)

- Guidance-corpus token floor (AC-TU-005): `git grep -c -E 'moai (todo|gtd) [a-z]' -- <17 files>` printed per-file matching-line counts: `.claude/skills/moai/SKILL.md` 2, `internal/template/templates/.claude/skills/moai/SKILL.md` 2, `.claude/skills/moai/workflows/gtd.md` 42, its template copy 42, four READMEs 2 each (8), `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-gtd.md` 9 each (36), `moai-todo.md` en 50 and ko, ja, zh 48 each (194). Sum 326 lines over 17 files. `-c` counts lines, so tokens are at least 326. The criterion's floors are 300 tokens and 17 files.
- Residual-sweep file floor (AC-TU-010): `python3 -I <scratchpad>/count_scope.py <tree>` (it calls `git -C <tree> ls-files -z`) printed `tracked 16151`, `E1 history 7008`, `E3 fixtures 53`, and `in scope (E1,E3 removed; binary not yet filtered) 9090`. The binary share was not measured, so 9,090 is an upper bound; the criterion's floor is 5,000.

## §R.6 Persisted and structural names that must not move (Q3)

- Tables `gtd_meta`, `gtd_items`, `gtd_relations`, `gtd_contracts`, `gtd_missions`, `gtd_events`, `gtd_operations` (`internal/factory/backlog_gtd_schema.go`).
- The item id pattern `^gtd-[0-9a-f]{16}$` (`internal/cli/gtd.go`, `gtdIDPattern`).
- The answers directory `.moai/gtd/answers` (`internal/cli/gtd_answer.go`); no tracked code reads it.
- 47 files under `internal/` carry GTD identifier-shaped tokens (`git grep -I -l -E 'GTD[A-Z][a-z]+|gtd[A-Z_]' -- internal`); separately, 77 Go files under `internal` contain "gtd" in some form, 26 of them non-test and 51 test files.
- `.gitignore:169` un-ignores `.agents/skills/moai-gtd/`; `internal/template/published_skills.go` lists `moai-gtd` and `moai-todo` in the exact published set (17 names).

## §R.7 Executable consumers of the `gtd` spelling

- `mods/moai-board/hooks/data.ts`: argument vectors `['moai', 'gtd', 'list', '--json']`, `['moai', 'gtd', 'list', '--limit', '0']`, `['moai', 'gtd', 'next', id, '--expect', prefix]`, and the label `moai gtd list`. A text pattern `moai gtd` does not match the vector form.
- `scripts/cutover-precheck.sh`: `GTD_CMD="${CUTOVER_GTD_LIST_CMD:-moai gtd list}"` (the environment variable name is a second token form).
- `internal/cli/factory_card.go:1320`: the refusal text points the operator at `moai gtd hold`.
- `internal/cli/todo.go:226`: the command help states that `moai gtd` "remains a fully supported compatibility alias".
- Recorded fixtures that must stay unchanged: `internal/cli/testdata/codex-rollouts-t1171/**` (a real recorded rollout and two role files), `internal/template/testdata/binding_ledger.json`.

## §R.8 Siblings that touch the same ground

| Item | State | Overlap |
|------|-------|---------|
| Card t1639 (3.2-2-2 factory v2) | hold | owns the landing-to-done machine transition with three observers, receipts issued at observation time, and the `pushed → done` decider path (Q4: this SPEC issues a receipt only inside the two close paths and declares no observer and no `pushed → done` edge) |
| Card t1621 / `SPEC-FACTORY-RECORD-REPAIR-001` | draft | moved its `pushed → done` item to t1639 (spec.md:47, :100-101, :134-136); same transition-table rows as M1. **spec_id surface (revision 2):** board record `d-20261010T001732Z-d4e8` (read this run by `moai decision read --scope card:t1641`, 108 records) says "The missing lane-side surface is repair card t1621 scope"; that SPEC's 0.2.0/0.4.0 narrowing lists "changing the SPEC id while a lease is held (the SPEC id is immutable once picked)" as an out-of-scope retired item (spec.md:115, also :99, :113), and none of its ten active requirements carries a spec_id surface. So the pointer is unmet: this SPEC delivers only the leader-side set-once verb (Q9), and the lane-side surface has no owner. Interaction: a lane-side surface that later fills the factory carrier only leaves the queue carrier empty; this verb then fills it (per-carrier set-once, equal values are a no-op), so the two compose. |
| `SPEC-FACTORY-COMPLETION-RECOVERY-001` | draft (`status: draft`, `version 0.1.0`, `updated 2026-10-06`, read at HEAD `3dc1052f6f`) | REQ-FCR-010 would implement T19, opposite of the removal (Q6); its REQ-FCR-002 receipt gate is the gate this SPEC keeps unchanged. **REQ-FCR-014 (revision 3, audit D23):** `spec.md:102` reads "No lane surface shall mint a receipt, invoke the reaper, or reassign a card"; the code carries it in the `@MX:REASON` at `internal/cli/leader_approval_gate.go:282` ("any second minting path would break the lane boundary this verb enforces") and in two doc comments of `internal/homestate/leader_approval.go` (`:9-13`, `:144-148`). REQ-TU-017 reverses it for one path: the lane's own-card `done` now causes the closing edge to mint a receipt when three machine facts hold. **Which wins:** this SPEC, for that one path (leader-confirmed 10-18, conditional; the draft is unrun); REQ-FCR-014 stands everywhere else and the draft is not edited here. M2a updates the three comments and re-reads the draft's status at its pre-flight, blocking if it moved. REQ-FCR-016 ("no recovery path completes a card", `spec.md:104`) is not reversed: the lane's `done` and the auto-done scan are evidence-gated completion paths, not the recovery paths (reaper, watchdog, reassign) it names |
| `SPEC-LOCAL-MAIN-FLOW-001` | completed | assigns "the T18 admission for the local-main surface" to card t1621; current code still refuses T18 while a remote exists |
| `SPEC-GITHUB-FLOW-DEFAULT-001` | in-progress | the "GitHub Flow transition complete" premise of D4 |
| `SPEC-TODO-LAND-AUTO-DONE-001` | completed | the closed evidence set and three misfire guards the lane path reuses |
| `SPEC-GTD-CANON-BODY-001` | completed | its body-check script pins the current workflow body; card t1602 item 4 updates it |
| `SPEC-FACTORY-LEASE-RENEW-001` (card t1663) | plan artifacts committed (seen in `git log`) | overlap with the transition-table files was not measured |

## §R.9 Sources read

`internal/cli/gtd.go`, `gtd_answer.go`, `todo.go` (tree, guard, `next`, `unpick`, `done`), `todo_autodone.go` (header), `leader_approval_gate.go` (header), `contract_report.go` (`cardToSpecID`), `internal/homestate/card_transition.go`, `card_evidence.go`, `card_pr_states.go`, `card_picked.go`, `card_audit_kickoff.go`, `internal/template/gtd_canonical_surface_test.go`, `published_skills.go`, `internal/cli/gtd_compat_test.go`, `todo_skill_doc_parity_test.go`, the three command files and the workflow skill header, `.moai/config/sections/git-strategy.yaml`, the 3.2 master plan (`.md`) and the redesign report (text extracted), the cited SPEC headers, and `moai todo show` for t1641, t1602, t1590, t1639, t1640. Revision 2 added: `internal/homestate/card_transition.go:100-345, 690-809`, `card_evidence.go:30-73`, `card_evidence_readers.go:20-21, 290-384`, `leader_approval.go` (whole), `internal/cli/leader_approval_gate.go` (whole), `internal/cli/factory_card.go:3299-3397`, `internal/factory/backlog_store.go:960-1030`, `internal/homestate/card_picked.go:120-240`, `internal/cli/gtd.go:1-260`, `gtd_compat_test.go:1-176`, `internal/graph/gtd_private.go:1-70`, `internal/factory/prlink_landedref.go` (grep), `spf13/cobra@v1.10.2/command.go:900-1010`, the audit report `.moai/reports/t1641/plan-audit.md`, and the decision board (`moai decision read --scope card:t1641`, `board=ok unparseable=0 records=108`). Revision 3 added: `.moai/reports/t1641/plan-audit-iter2.md` (whole), `spf13/cobra@v1.10.2/command.go:905-1045`, `internal/factory/integration_remeasure.go` (whole), `internal/homestate/card_evidence_readers.go` (whole), `leader_approval.go` (whole), `card_transition.go:90-420, 560-908`, `card_record.go` (whole), `card_picked.go:100-299`, `card_audit_kickoff.go:1-60`, `internal/cli/todo.go:200-460, 1640-1700`, `internal/cli/gtd.go:1-260`, `internal/cli/factory_card.go:100-140, 2340-2400`, `internal/cli/leader_approval_gate.go:270-300`, `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md:1-20, 88-117`, the committed objects of the four merges of §R.14 (`git show`, `git ls-tree`, `git diff --name-only`), and the re-measure records of two of them.

## §R.10 The workflow document path is an identifier (revision 2)

`git grep -n -F 'workflows/gtd.md' -- internal` printed, outside prose comments: the `paths:` frontmatter key of five template rule files (`cross-session-messaging.md`, `factory-dispatch.md`, `factory-dispatch-cards.md`, `factory-dispatch-detail.md`, `factory-dispatch-gates.md`, line 3 of each; `git grep -c -F` printed 1 for each), the pinned-paths test (`internal/template/workflow_rule_paths_pinned_test.go:31`), the generated embed line (`internal/template/embed_manifest_gen.go:516`), the recorded fixture `internal/template/testdata/binding_ledger.json` (the `entry_point` strings; class E3), and tests in `internal/cli/doc_json_shape_test.go:98`, `internal/template/backlog_json_disclosure_mirror_test.go:25`, `card_id_leak_test.go:55,101`, `db_paths_surfaces_test.go:47-54`, `gtd_canonical_surface_test.go:13`, `jev_auto_exception_test.go:76`. The router skill's last line (`.claude/skills/moai/SKILL.md:188`) says "Read workflows/gtd.md". Renaming the file would therefore be an identifier rename across these surfaces, which Q3 excludes, and the card (scope list "문서 치환", "슬래시 정본") does not ask for it. Decision: the document keeps its path; its header and body wording change.

Copies: `diff -q` printed nothing for the template copy against the deployed `.claude/skills/moai/workflows/gtd.md` (identical, 523 lines); `diff -q` of `.moai/workflows/gtd.md` against the deployed `.claude` copy printed `differ` (452 lines). `git ls-files internal/template/templates/.moai/workflows` listed only `README.md`, so `.moai/workflows/gtd.md` has no template source: it is a deployed-only file, edited by the script directly and named as such in the M8b batch record. Which of the `.moai/policies/**`, `.moai/project/**`, and `.moai/docs/**` paths among the 67 have a template source was not individually checked (pre-flight item 7 does it).

The four rule files named in the conversion set (`factory-dispatch.md`, `-cards`, `-detail`, `main-checkout-branch-guard-detail.md`) all carry a `paths:` key (`git grep -n -m1 '^paths:'` printed one line for each), so the only always-loaded files among the 73 are `AGENTS.md` and its template.

## §R.11 Cobra's execution order (revision 2, D7; completed in revision 3, D19)

`spf13/cobra@v1.10.2/command.go`, `func (c *Command) execute`. Revision 3 re-read the whole function (`sed -n 905,1045p`) and located its anchors with `grep -n`, which printed: `905:func (c *Command) execute(a []string) (err error) {`, `919: err = c.ParseFlags(a)`, `935: return flag.ErrHelp` (the help flag), `956: return flag.ErrHelp` (a command that is not runnable), `968: if err := c.ValidateArgs(argWoFlags); err != nil {`, `985: if p.PersistentPreRunE != nil {`, `999: if c.PreRunE != nil {`, `1007: if err := c.ValidateRequiredFlags(); err != nil {`, `1010: if err := c.ValidateFlagGroups(); err != nil {`, `1014: if c.RunE != nil {`; `go.mod:23` pins `github.com/spf13/cobra v1.10.2`. So the order is: parse flags, help, `ValidateArgs`, persistent pre-run, local pre-run, **required-flag validation, flag-group validation**, run. (The loop from `:985` runs only the nearest persistent hook unless `EnableTraverseRunHooks` is set, `:974-994`; the default is off, so on the `gtd` tree the one persistent hook is the lane guard defined on the parent.) Revision 2 stated only the first half of this order and concluded that a warning in the pre-run hook never precedes a refusal; the audit's reading (`:985` before `:1007`) is confirmed: a missing required flag and a conflicting flag group are refused **after** the hook. Three stage verbs carry required flags and the three `--auto*` flags are mutually exclusive (E-23). `internal/cli/todo.go:260-279` puts the mistyped-verb guard in the parent's `Args` function (before the hook) and `:333` puts the lane guard in `PersistentPreRunE`. Consequence for the design: a warning printed from the pre-run hook precedes the required-flag and flag-group refusals; printing it from the run step (`:1014`) precedes none of them, and no command of either tree uses `Run` (E-22), so wrapping `RunE` reaches every handler. Established by reading the source and grepping its anchors; no throwaway program was run (a Gap: the first execution of the wrapper's rows is M5's RED step).

## §R.12 `reflect --rebuild-projection` writes (revision 2)

`internal/cli/gtd.go:165-176`: `reflect` calls `ReflectGTDStore` over `newTodoReadStore()`; with `--rebuild-projection` it also calls `graph.BuildPrivateGTDProjectionFromStore`, which calls `BuildPrivateGTDProjection(ctx, store.EnginePath(), ...)`; `git grep -n -E 'WriteFile|os.Rename|os.Create|OpenFile|MkdirAll' -- internal/graph/gtd_private.go` printed `:179 os.CreateTemp`, `:264 os.MkdirAll`, `:273 os.Rename`. So the flag form writes private projection files and the bare form does not (read, not executed).

## §R.13 What the approval gate checks today (revision 2, Q5)

`internal/homestate/leader_approval.go:116-142` `VerifyBinding`: issuer role must equal `leader`; the canonical issuer label must differ from the card's performing owner label; the card uuid must match when supplied; the run id must match; `FactoryVersion` must equal the card's version; the evidence hash must be non-empty and equal the card's `EvidenceSHA`. `IssueLeaderApproval` (`:149-182`) refuses a role other than `leader`, an empty issuer, an empty hash. `verifyTransitionApproval` (`:530-543`) is the done transitions' gate inside `Transition`'s transaction; `ApprovalGate.Verify` (`:493-502`) is the archive-moment gate. Issuance is reachable only through `moai factory approve` (`leader_approval_gate.go:275-375`, lane sessions refused at `:290`). The automatic issuance of REQ-TU-017 adds a second issuance path whose preconditions are the three facts and which runs inside the closing edge's transaction (revision 3, `design.md` §D.5); it changes no verification, and `IssueLeaderApproval` itself keeps its leader-only role check (`:157-159`).

## §R.14 What the legitimate landing flow looks like, measured (revision 3, D16, D17)

Purpose: choose the narrowest binding rule for facts 2 and 3 that still admits the flows lanes run today, and say what the rule lets through. All commands are plain `git -C <own absolute path>` reads at HEAD `3dc1052f6f` unless noted. The eight merges are integration merges named in the memory index: the four most recent (t1857, t1852, t1856, t1837) and four earlier ones (t1776, t1793, t1732, t1812). The first version of this section sampled only the first four and a rule that fitted them (reports-only) turned out too narrow once t1812 was read; both are recorded.

| Card | Merge commit | First parent | Second parent (card tip) | `.moai/reports/<card>/` in the second parent's tree |
|------|--------------|--------------|--------------------------|------------------------------------------------------|
| t1857 | `f4bcd4f2cc65342559b505f1e5a29df80f992894` | `dc9cab9843f0ebab04a6f317edcb3b165bf282a3` | `38c550f414f256622b8a16787df9d4e03c3a5ce2` | no files (`git ls-tree -r --name-only` printed nothing); the merge brought only `internal/harness/rosterguard/registry.go` (`git diff --name-only <first parent> <merge>`) |
| t1852 | `0e083d15239aede9d605e7a06da878a159cb104d` | `7f0dd62af20d75297e895bfa5b606e16f5f8563c` | `476c33d3b2b998d38e482d996644e42652fb8b7d` | no files (printed nothing) |
| t1856 | `7f0dd62af20d75297e895bfa5b606e16f5f8563c` | `f4bcd4f2cc65342559b505f1e5a29df80f992894` | `9668c74ded6e01049b62bec2d24e1abaf90f38f0` | `progress.md`, `sync-audit.md`, `verdict.md` |
| t1837 | `667e4bc83ff372a669d7b402116654cd63e12666` | `3642a3a9b9cf101863656153d2d99b69fcb66a05` | `41ec90c81e31cdeaada93d9ff956af4cca956489` | `card-review.md`, `sync-audit.md`, `verdict.md` |
| t1812 | `d2d9ac08fb1899440da911121ddf13eecc806a3f` | `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6` | `a49f31013af391c00f42162c67642343319cc66d` | `card-review.md`, `progress.md`, `sync-audit.md` |
| t1793 | `2534cf400a6c32adfca25bc83abbe8e9976f7d5f` | `f088cdbde94f64f5464ab4db6545a8d9450d557b` | `0a1a8f5e426cc38744221427f1329e4081e25b0f` | no files (printed nothing) |
| t1776 | `f088cdbde94f64f5464ab4db6545a8d9450d557b` | `fa6c425e2ea3ea91ed56580568614fb8414cf1b0` | `a89eacdc7142fbae1e5d9601a8f86b0fb1d522c5` | no files (printed nothing) |
| t1732 | `a9a4d6ea0521ec360aa9ab81094407fec246c4bc` | `e04ebd5ca6bc441b380c1072d5a1dff8431459ba` | `ed07c13602120403ec6cccf9356f6407886a3b66` | no files (printed nothing) |

For the three cards that commit a verdict (t1856, t1837, t1812):

- The verdict file read from the second parent's tree (`git show <second parent>:.moai/reports/<card>/sync-audit.md`) carries `verdict: PASS` and `audited_sha: b234210fcd1f3e912a974032b8a5aa4b8e7bd3d4` (t1856), `audited_sha: 1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365` (t1837), and `audited_sha: 12ac4803d8166c24ab4db5fd7a1e62892c2f0a59` (t1812).
- `git diff --name-only <audited_sha> <second parent>`: for **t1856** it printed exactly `.moai/reports/t1856/sync-audit.md` and `.moai/reports/t1856/verdict.md` (exit 0), for **t1837** exactly `.moai/reports/t1837/sync-audit.md` and `.moai/reports/t1837/verdict.md` (exit 0), and for **t1812** it printed **254 paths** (exit 0): the card branch absorbed main after the audit, so the absorbed files of other cards, among them whole SPEC directories, `internal/cli/update/**`, and other cards' reports, differ from the audited commit. The audited commit is an ancestor of the card tip in all three (`git merge-base --is-ancestor`, exit 0). For t1856 and t1837 the card tip differs from the audited commit by the card's own report files only, because the verdict file is committed after the commit it names (the t1856 verdict file records that its audited commit `b234210fc` was the card branch after main was absorbed, a re-audit, "round 5"); t1812 is the other flow: audit, then absorb main, then re-measure and land, the order the landing train prescribes (board record `d-20261018T012631Z-2918`, "absorb main, remeasure ..., complete").
- Both candidate rules, run by `ad_rule.py` (a throwaway script in the session scratchpad that calls `git -C <own path> diff --name-only --no-renames -z` for the sets `A` = audited..tip and `D` = first parent..tip): t1856 `A` 2, `D` 7, `A`∩`D` 2, R1 PASS, R2 PASS; t1837 `A` 2, `D` 7, `A`∩`D` 2, R1 PASS, R2 PASS; t1812 `A` 254, `D` 5, `A`∩`D` 2, **R1 FAIL (252 violations), R2 PASS (0)**. R1 is "every path of `A` lies under the card's reports directory"; R2 is "every path of `A`∩`D` lies under it". R2 is the rule REQ-TU-017 states. It admits an absorbed main (those paths equal the first parent's, so they are not in `D`) and refuses any post-audit change the card authored (it is in `A` and, differing from main, in `D`).
- The merge tree equals the second parent's tree in all three (`git rev-parse <merge>^{tree} <second parent>^{tree}` printed one identical SHA: `b5f972f944697a9304066b3284f1c3954b46b5ea` for t1856, `c89464065eeb3536f32e7a49150888c1674710fe` for t1837, `7364bdf82fdfbf4461950c4b14cc178d0c427ff7` for t1812).
- The re-measure record `/Users/goos/MoAI/moai-adk-go/.moai/state/remeasure/<merge tree>.json` (read with `cat`) names `tree` equal to the merge tree in all three and `base` equal to the merge's **first parent** in all three (t1856: `f4bcd4f2cc65342559b505f1e5a29df80f992894`; t1837: `3642a3a9b9cf101863656153d2d99b69fcb66a05`; t1812: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`), with `exit_code: 0`, `structured_count: true`, `test_count` 4595, 8082, and 11228. So fact 3's comparison (tree equals the merged tree, base equals the first parent) holds for the live flows, including the one whose audit predates the absorption.

Findings the design uses: (a) the narrowest admissible rule for fact 2 is R2, not equality with the audited commit, not ancestry, and not R1; (b) fact 3's base rule can be equality with the first parent; (c) **five of the eight merges (t1857, t1852, t1793, t1776, t1732) carry no committed report files at all in the card tip** (the three that do are the most recent audited ones, t1856, t1837, and t1812), so the automatic path would not issue for the five (fact 2 fails, fail-closed, the leader approves by hand) and the ten-card criterion's live rate depends on lanes committing their evidence; the sample is small and its recency pattern is a reading, not a measured cause; (d) the verdict file does not carry a machine-readable audit receipt line (t1856's file says the `AUDIT-VERDICT:` line "is carried in the chat message, not here"), and the audit-receipt store exists (`internal/auditreceipt/store.go:15`, "written by the MCP server, read by the SubagentStop and PreToolUse guards"), so a lane-uncorroborated verdict is a residual the SPEC names and does not close.

The adjacent finding left out of scope: the T16 callback (`internal/cli/factory_card.go:2372-2378`) validates the record but never compares its `tree` or `base`, and `ReadRemeasureRecord` returns them "as written" on purpose (`integration_remeasure.go:98-103`, "forcing the field here would turn that check into dead code"; the comparison belongs to the caller, and the T16 caller does not make it). A record stored under a tree's file name that names another tree therefore satisfies T16 today. REQ-TU-017 makes the automatic path re-check both fields itself; tightening T16 is its owner's (spec.md Out of Scope). Not measured: whether the merge verb's own step (REQ-MWQ-018 cause 4) makes the comparison before T16 is reached; the comment says so, the call site was not read.

## §R.15 The wedge the setter exists for (revision 3, D20, D21)

- The board record `d-20261010T001732Z-d4e8` (`moai decision read --scope card:t1641`, saved to the scratchpad and located with `grep -n`; `board=ok unparseable=0 records=108`) says: predicate "a card of run tmnboq leased without a spec_id whose SPEC was authored after the lease"; body "The leader stamps spec_id on a leased card row once the lane names the SPEC directory it authored: a conditional UPDATE on an empty spec_id, database backup first, nothing else changed. Stamped now: t1616 = SPEC-LOCAL-MAIN-FLOW-001, t1619 = SPEC-UPDATE-REPAIR-RC30-001, t1620 = SPEC-REVIEW-GATE-REPAIR-001"; evidence "spec_id cannot be set while leased ... card_picked.go:184-186". The wedge is therefore a **leased** card, not a `picked` or `assigned` one.
- Who reads the identifier: the plan-audit verdict admission requires a valid `SpecID` on the factory row (`internal/homestate/card_evidence_readers.go:185-188`, "card carries no valid SPEC id") and the audit kickoff builds the SPEC directory from it (`card_audit_kickoff.go:31`). Both sit between lease and `run`, so a first fill is needed from `picked` through `kickoff` and from nowhere after.
- Lease acquisition needs no spec_id: `case guardLeaseAcquire` (`card_transition.go:582-607`) checks the registered worker and the owner label only, and `stageAdmits` lets a card with an empty stage resume into `plan` or `run` (`:813-818`). So a half-written pair can be overtaken by the lease edge and by T4a/T4b before the leader repeats.
- A half-written card cannot pass `plan-audit → kickoff`: the verdict admission refuses an empty factory `SpecID`, so the card it leaves behind is at most `plan-audit` for a Class C card.
- The queue item leaves `picked` only through `unpick` or `done` (the help text of `hold`, `drop`, `unpick`, `done`; read, not exercised); its `SpecID` is a pointer field set at pick (`internal/cli/todo.go` `rec.Items[i].SpecID = &spec`, and cleared by unpick).
- The version bump of a field write stales a leader receipt bound to the old version (`VerifyBinding` compares `FactoryVersion`, `leader_approval.go:129`); that is why the first-fill set stops before `run`.

## §R.16 Post-ceiling correction: what was read and what was run (revision 4, D29 to D31)

**D29, the push gate caller (read, E-24).** `factoryDecideCards` takes one integration-branch value, `config.LoadGitFlowIntegrationConfig(root).IntegrationTarget` (`internal/cli/factory_card.go:3238`), and `decideOne` forwards it as `TransitionRequest.IntegrationBranch` (`:3394`); it tests it only when the request is `pushed` (`:3334-3336`). `LoadGitFlowIntegrationConfig` returns the zero value when `git-strategy.yaml` is absent or unreadable (`internal/config/loader_integration_branch.go:94-97`) and fills the target only inside `if ok` (`:118-127`); `WorkflowIntegrationTarget` answers `""` for a workflow it does not name (`internal/config/loader_workflow_disposition.go:100-101`) and for a named one whose key is empty (`:95-99`). Removed T18's no-remote branch takes no branch name (`internal/homestate/card_transition.go:730-746`); `verifyMerge` and a T29 that named `refs/heads/<IntegrationBranch>` both need a usable name (`card_evidence_readers.go:273-276`, `:291`, `:316`). The audit's first alternative, `factoryResolveIntegrationBranch` (`factory_card.go:2552-2557`), answers `factoryBranchOfWorktree(card.WorktreePath)` with source `BranchSourceCaller` when nothing is configured, and `factory complete` refuses exactly that source and the card's own branch as an integration branch (`:2146-2162`). The merge verb falls back to the literal `develop` and the window record's branch (`internal/cli/integration_merge.go:54-63`), neither of which survives the window's release as a record of where a merge landed. Not run: a CLI invocation of `decide --gate push` in a configuration-less temporary repository (no `go test`, and a built `moai` cannot reach a `merged-local` row without the whole factory chain); the claim rests on the read above and on the audit's own reading (P9), and AC-TU-018's CLI-level row is its test.

**D30, the receipt writers (read and grepped, E-25).** The table's issuer column has two writers outside tests: the insert in `IssueLeaderApproval` (`internal/homestate/leader_approval.go:173`, `ON CONFLICT(card_uuid,run_id) DO UPDATE SET ... issuer=excluded.issuer`, `:173-175`) and the version re-stamp (`card_transition.go:849`, which sets `factory_version` only). `IssueLeaderApproval` has one caller outside tests, `internal/cli/leader_approval_gate.go:350`, which passes the `--issuer` string (`:344-349`, flag at `:373`); the function itself refuses only a role other than `leader`, an empty issuer, and an empty hash (`:157-168`). No tracked non-test code writes the issuer `auto` (`git grep ... 'Issuer: *"auto"'` exit 1). Not measured: whether any row in a live factory store already carries `auto` (the store lives under the operator's home, was not opened, and M2a's pre-flight item 8 counts it).

**D31, git's behaviour (run in throwaway repositories, E-26 for the code side).** Two scratchpad scripts built the repositories and called `git -C <repo>`; git was `git version 2.54.0 (Apple Git-157)`. Verbatim results:

- *Leading-space path.* A commit added ` .moai/reports/c1/x.txt` (a top-level directory named with a leading space) and changed `code.txt`. `git diff --name-only --no-renames -z E T` printed ` .moai/reports/c1/x.txt\0code.txt\0`; the first entry split on NUL has the prefix `.moai/reports/c1/` = False, and after a whole-output strip = True. A ten-line Go program applying `strings.TrimSpace` to those bytes (the operation at `card_evidence.go:33`) printed `trimmed first path = ".moai/reports/c1/x.txt", has prefix .moai/reports/c1/ = true` and `untrimmed first path = " .moai/reports/c1/x.txt", has prefix .moai/reports/c1/ = false`.
- *Replacement ref.* `git replace E E3`, where `E3` was made by `git commit-tree <T's tree> -p E`: `git diff --name-only --no-renames -z E T` printed nothing (empty); the same command with `--no-replace-objects` printed ` .moai/reports/c1/x.txt\0code.txt\0`, and with `GIT_NO_REPLACE_OBJECTS=1` the same; `git rev-parse E^{tree}` printed `848b4e91c0f7` by default and `c702759972b9` with the option; `git rev-parse E` printed `E` itself in both, so the name a row records shows nothing.
- *Graft and shallow files.* From a linked worktree, `git rev-parse --git-path info/grafts` and `--git-path shallow` printed paths under the main checkout's `.git`. With the tip listed in `info/grafts`, `git rev-list --parents -n 1 <tip>` printed the tip alone (no parent) by default and with `--no-replace-objects`; with the file removed it printed the tip and its parent. With the tip written into the `shallow` file the result was the tip alone in both forms and `git rev-parse --is-shallow-repository` printed `true` (it printed `false` before).
- *Containing branches.* After a `--no-ff` merge M of the card into `main` and a branch `other-lane` cut from `main`, `git for-each-ref --format=%(refname) --contains M refs/heads/` printed `refs/heads/main` and `refs/heads/other-lane` (not the card's branch); after `main` was reset below M it printed `refs/heads/other-lane` only.
- *Documentation.* `man git` read locally: "--no-replace-objects: Do not use replacement refs to replace Git objects. This is equivalent to exporting the GIT_NO_REPLACE_OBJECTS environment variable with any value."

Gaps of this section: one git version only; the Go program reproduces the `TrimSpace` line and does not call the unexported `gitRead`; the graft entry used lists the tip alone (a parentless rewrite), not the first-parent rewrite of AC-TU-016 row (k1), whose form the run-phase fixture must pick and record; none of the AC-TU-016 rows (i) to (k) exists as a Go test yet.
