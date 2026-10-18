# Research: SPEC-TODO-UNIFY-001

Measured at tree `026f48658f` (branch `WT-todo-unify`, clean) on 2026-10-18. Every figure below names the command that produced it. A figure that cannot be re-run on the current tree is marked.

## §R.1 Provenance and refusals

- **Tree.** `git rev-parse --short HEAD` printed `026f48658f`; `git branch --show-current` printed `WT-todo-unify`. Revision 2 (audit iteration 1 FAIL) measured at HEAD `23e40eb7c4`, which adds only plan artifacts and the audit report: `git diff --stat 026f48658f 23e40eb7c4 -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing. Ledger entries E-15 to E-18 and the §R.5b counts carry the `23e40eb7c4` pin.
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
| E-15 | `git grep -n -E 'renamed to moai todo' -- internal/cli/gtd.go` (tree `23e40eb7c4`) | not displayed; git's contract for no match is 1 | (empty) |
| E-16 | `git grep -n -c 'fully supported compatibility alias' -- AGENTS.md internal/template/templates/AGENTS.md.tmpl` (tree `23e40eb7c4`) | not displayed; a match prints, so 0 | `AGENTS.md:1` and `internal/template/templates/AGENTS.md.tmpl:1` (lines `AGENTS.md:344` and `AGENTS.md.tmpl:343`) |
| E-17 | `git grep -n -E '"T29"' -- internal/homestate/card_transition.go` (tree `23e40eb7c4`) | not displayed; git's contract for no match is 1 | (empty) |
| E-18 | `git grep -n -E 'ApprovalIssuerAuto\|IssuerAuto\|RevokeLeaderApproval' -- internal/homestate internal/cli` (tree `23e40eb7c4`) | not displayed; git's contract for no match is 1 | (empty) |

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

Reading of E-15, E-17, E-18: the tool printed no exit status for an empty `git grep`, so the exit field is the git contract and not an observation of this run (a Gap); stdout was observed empty. E-15 and E-17 and E-18 are absence-of-code observations of the same kind as E-04 to E-07; E-15 is the proxy that flips when the warning's lead phrase (REQ-TU-002) lands in `gtd.go`, and E-16 flips when the wording test's phrase replaces the old one.

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
| `SPEC-FACTORY-COMPLETION-RECOVERY-001` | draft | REQ-FCR-010 would implement T19, opposite of the removal (Q6); its REQ-FCR-002 receipt gate is the gate this SPEC keeps unchanged |
| `SPEC-LOCAL-MAIN-FLOW-001` | completed | assigns "the T18 admission for the local-main surface" to card t1621; current code still refuses T18 while a remote exists |
| `SPEC-GITHUB-FLOW-DEFAULT-001` | in-progress | the "GitHub Flow transition complete" premise of D4 |
| `SPEC-TODO-LAND-AUTO-DONE-001` | completed | the closed evidence set and three misfire guards the lane path reuses |
| `SPEC-GTD-CANON-BODY-001` | completed | its body-check script pins the current workflow body; card t1602 item 4 updates it |
| `SPEC-FACTORY-LEASE-RENEW-001` (card t1663) | plan artifacts committed (seen in `git log`) | overlap with the transition-table files was not measured |

## §R.9 Sources read

`internal/cli/gtd.go`, `gtd_answer.go`, `todo.go` (tree, guard, `next`, `unpick`, `done`), `todo_autodone.go` (header), `leader_approval_gate.go` (header), `contract_report.go` (`cardToSpecID`), `internal/homestate/card_transition.go`, `card_evidence.go`, `card_pr_states.go`, `card_picked.go`, `card_audit_kickoff.go`, `internal/template/gtd_canonical_surface_test.go`, `published_skills.go`, `internal/cli/gtd_compat_test.go`, `todo_skill_doc_parity_test.go`, the three command files and the workflow skill header, `.moai/config/sections/git-strategy.yaml`, the 3.2 master plan (`.md`) and the redesign report (text extracted), the cited SPEC headers, and `moai todo show` for t1641, t1602, t1590, t1639, t1640. Revision 2 added: `internal/homestate/card_transition.go:100-345, 690-809`, `card_evidence.go:30-73`, `card_evidence_readers.go:20-21, 290-384`, `leader_approval.go` (whole), `internal/cli/leader_approval_gate.go` (whole), `internal/cli/factory_card.go:3299-3397`, `internal/factory/backlog_store.go:960-1030`, `internal/homestate/card_picked.go:120-240`, `internal/cli/gtd.go:1-260`, `gtd_compat_test.go:1-176`, `internal/graph/gtd_private.go:1-70`, `internal/factory/prlink_landedref.go` (grep), `spf13/cobra@v1.10.2/command.go:900-1010`, the audit report `.moai/reports/t1641/plan-audit.md`, and the decision board (`moai decision read --scope card:t1641`, `board=ok unparseable=0 records=108`).

## §R.10 The workflow document path is an identifier (revision 2)

`git grep -n -F 'workflows/gtd.md' -- internal` printed, outside prose comments: the `paths:` frontmatter key of five template rule files (`cross-session-messaging.md`, `factory-dispatch.md`, `factory-dispatch-cards.md`, `factory-dispatch-detail.md`, `factory-dispatch-gates.md`, line 3 of each; `git grep -c -F` printed 1 for each), the pinned-paths test (`internal/template/workflow_rule_paths_pinned_test.go:31`), the generated embed line (`internal/template/embed_manifest_gen.go:516`), the recorded fixture `internal/template/testdata/binding_ledger.json` (the `entry_point` strings; class E3), and tests in `internal/cli/doc_json_shape_test.go:98`, `internal/template/backlog_json_disclosure_mirror_test.go:25`, `card_id_leak_test.go:55,101`, `db_paths_surfaces_test.go:47-54`, `gtd_canonical_surface_test.go:13`, `jev_auto_exception_test.go:76`. The router skill's last line (`.claude/skills/moai/SKILL.md:188`) says "Read workflows/gtd.md". Renaming the file would therefore be an identifier rename across these surfaces, which Q3 excludes, and the card (scope list "문서 치환", "슬래시 정본") does not ask for it. Decision: the document keeps its path; its header and body wording change.

Copies: `diff -q` printed nothing for the template copy against the deployed `.claude/skills/moai/workflows/gtd.md` (identical, 523 lines); `diff -q` of `.moai/workflows/gtd.md` against the deployed `.claude` copy printed `differ` (452 lines). `git ls-files internal/template/templates/.moai/workflows` listed only `README.md`, so `.moai/workflows/gtd.md` has no template source: it is a deployed-only file, edited by the script directly and named as such in the M8b batch record. Which of the `.moai/policies/**`, `.moai/project/**`, and `.moai/docs/**` paths among the 67 have a template source was not individually checked (pre-flight item 7 does it).

The four rule files named in the conversion set (`factory-dispatch.md`, `-cards`, `-detail`, `main-checkout-branch-guard-detail.md`) all carry a `paths:` key (`git grep -n -m1 '^paths:'` printed one line for each), so the only always-loaded files among the 73 are `AGENTS.md` and its template.

## §R.11 Cobra's execution order (revision 2, D7)

`spf13/cobra@v1.10.2/command.go`, `func (c *Command) execute`: `ParseFlags` returns on a flag error (`:919-922`); a help flag returns `flag.ErrHelp` (`:926-935`); `c.preRun()` (`:959`); `ValidateArgs(argWoFlags)` returns its error (`:968-970`); only then the loop over `PersistentPreRunE` (`:984-998`) and `PreRunE` (`:999-1005`). `go.mod:23` pins `v1.10.2`. `internal/cli/todo.go:260-279` puts the mistyped-verb guard in the parent's `Args` function and `:333` puts the lane guard in `PersistentPreRunE`. Consequence: a mistyped verb, a missing argument, an unknown flag, and a help request all return before any pre-run hook, so a warning placed in the pre-run hook never fires for them. This was established by reading the source; no throwaway program was run (a Gap).

## §R.12 `reflect --rebuild-projection` writes (revision 2)

`internal/cli/gtd.go:165-176`: `reflect` calls `ReflectGTDStore` over `newTodoReadStore()`; with `--rebuild-projection` it also calls `graph.BuildPrivateGTDProjectionFromStore`, which calls `BuildPrivateGTDProjection(ctx, store.EnginePath(), ...)`; `git grep -n -E 'WriteFile|os.Rename|os.Create|OpenFile|MkdirAll' -- internal/graph/gtd_private.go` printed `:179 os.CreateTemp`, `:264 os.MkdirAll`, `:273 os.Rename`. So the flag form writes private projection files and the bare form does not (read, not executed).

## §R.13 What the approval gate checks today (revision 2, Q5)

`internal/homestate/leader_approval.go:116-142` `VerifyBinding`: issuer role must equal `leader`; the canonical issuer label must differ from the card's performing owner label; the card uuid must match when supplied; the run id must match; `FactoryVersion` must equal the card's version; the evidence hash must be non-empty and equal the card's `EvidenceSHA`. `IssueLeaderApproval` (`:149-182`) refuses a role other than `leader`, an empty issuer, an empty hash. `verifyTransitionApproval` (`:530-543`) is the done transitions' gate inside `Transition`'s transaction; `ApprovalGate.Verify` (`:493-502`) is the archive-moment gate. Issuance is reachable only through `moai factory approve` (`leader_approval_gate.go:275-375`, lane sessions refused at `:290`). The auto-issuing function of REQ-TU-017 adds a second issuance path whose preconditions are the three facts; it changes no verification.
