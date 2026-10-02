# SPEC-JEV-AUTO-EXCEPTION-001 — Research

Evidence for `spec.md`, kept here because the verbatim output exceeds 50 lines.
All measurements: pinned tree `c50da9c2f`, branch `WT-jev-auto-exception`, taken in the
plan run; the revision (plan-audit iteration 1) re-ran the ledger and the third sweep at
tree `1eef55dd9` (§R5) — `git diff --name-only c50da9c2f 1eef55dd9` lists only this
SPEC's directory. Go commands ran under the eleven-variable scrub in one compound
invocation.

## §R1 Inventory — every restatement found, classified

Sweeps (all three over tracked files, excluding `.moai/specs/`, `.moai/reports/`,
`CHANGELOG.md`):

- **Primary**: `git grep -n -i -E 'display-only|display only'` → **69 hits in 43
  files** (raw output in §R2).
- **Synonym**: `git grep -n -E 'never reorder by inferred priority|판단 자료|모델 답을 입력으로도'`
  → **5 hits** (§R2); it added `SKILL.md:180` and the doctrine guide, neither of which
  contains the literal "display-only".
- **Closed-target phrases** (added in the revision after plan-audit iteration 1 found
  the reference skill that neither sweep above matched):
  `git grep -n -i -E 'a person reads|labelled model|queue mutation|hard to undo'` →
  **84 hits in 47 files** (re-measured: `| wc -l` → 84, `-l | wc -l` → 47; raw output
  in §R2b, classification in §R1.5). It added the reference-skill pair (X5) as the
  only new class-(i) surface.
- **Specs and changelog** (excluded above): `git grep -c -i -E 'display-only|display only' -- .moai/specs`
  → dozens of files; the ones that bear on the principle are read in §R1.3.

### R1.1 Class (i) — states the principle, contradicts the exception — in scope

| ID | File:line (live) | Mirror | Evidence (verbatim fragment) | Origin |
|---|---|---|---|---|
| S1 | `internal/jev/jev.go:24-28` | — | `//   - **The capability is display-only.** A Jev answer is a labelled, … (see doc_display_only_test.go, which asserts the`; `:27` names a test file that does not exist (`internal/jev/` holds `display_only_test.go`; `git grep -n 'doc_display_only_test'` → `jev.go:27` only) | card |
| S2 | `.moai/config/sections/workflow.yaml:227-232` | `internal/template/templates/.moai/config/sections/workflow.yaml:229-234` | `# The capability is display-only: an answer is a labelled model-produced signal a person reads, never an input to a completion verdict, a merge approval, a queue mutation, …` then `# One exception: the contract-mode Kickoff \`llm+jev\` cross-check …` | card |
| S3 | `internal/cli/mcp_jev.go:8-10` | — | `// gated-unavailable. The capability is display-only: an answer is a labelled model signal a person reads, never a completion verdict, a merge approval, a queue mutation, or any other decision that is hard to undo.` | card |
| S4 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md:139`, `:233` | same lines in the template mirror (byte-identical copies) | `display-only — a labelled model signal a person reads, never a completion predicate, merge approval, queue mutation, or gate input, except as the second signal of the contract-mode Kickoff \`llm+jev\` cross-check …` | card |
| S5 | `.moai/specs/SPEC-JEV-CORE-001/spec.md:81`, `:99`, `:160-161`, HISTORY `:21-25`, `:4`, `:7` | — | REQ-JEVC-011 "The capability is display-only, with exactly one exception"; REQ-JEVC-012 "shall not consult Jev … as an input — for any of: … a \`moai todo\` or \`moai gtd\` mutation …. The single exception is the operator gate at Kickoff" | card |
| S6 | `.moai/specs/SPEC-MANAGER-TODO-001/spec.md:69`, `:71` | — | REQ-MT-014 "a permitted display-only signal … never as authority"; REQ-MT-015 "Jev output shall never be the basis of a queue mutation, a completion verdict, a merge approval, or any third-grade decision" | card |
| X1 | `.moai/docs/jev-local-operations.md:28-32` | — | `예외는 한 곳뿐이다 — contract 모드 Kickoff 의 \`llm+jev\` 교차 확인에서는 …` (`:32`), under `### [HARD] 되돌릴 수 없는 판정은 모델 답을 입력으로도 쓰지 않는다` (`:28`) | extension — the guide `TestJevAmendmentLinkage` names |
| X2 | `.claude/rules/moai/development/agent-authoring.md:147` | template `:147` | `- manager-todo: … consults Jev as a display-only signal; carries the read-only sealed-snapshot judgment …` | extension — observed by the predecessor (plan §Findings (e)) |
| X3 | `.claude/skills/moai/SKILL.md:180` | template `:180` | `The pick is the operator's: never preselect, never reorder by inferred priority, never auto-populate from TODO comments or issues.` (the identical sentence in `kanban-dispatch.md:29` was given a qualifier by the predecessor) | extension — observed by the predecessor |
| X4 | `CLAUDE.md:63` | template `:63` | `… manager-todo … (queue lifecycle, \`/moai:todo --auto\` serial processing, dispatch guidance, Jev display-only consultation) …` | extension, weakest |
| X5 | `.claude/skills/moai-ref-jev-question-design/SKILL.md:25-29` | template, same lines (byte-identical: `diff` silent, exit 0) | `decides nothing: the answer is a labelled model signal a person reads, never a completion predicate, a merge approval, a queue mutation, or any other decision that is hard to undo.` — the sentence S3 and S4 carry, stated at capability level | extension — found by the closed-target-phrase sweep; operator-confirmed (OD-3) |

### R1.2 Class (iii) — already compatible, or about a different object — left alone

| File:line | Why it stays |
|---|---|
| `.claude/agents/moai/manager-todo.md:7`, `:58`; template; `.codex/agents/moai/manager-todo.toml:9`, `:57` | amended by the predecessor: "except for the \`--auto\` cycle's own selection order"; pinned by `todo_auto_doc_test.go:351-359` and the manager-todo boundary test |
| `.claude/skills/moai/workflows/gtd.md:354`; template | predecessor text: "Jev's ranking answer is used for the selection order only" |
| `.claude/rules/moai/workflow/kanban-dispatch.md:31`; template | predecessor text (no "display-only" literal; carries the exception) |
| `.claude/rules/moai/workflow/auto-semantics.md:217`; template | §12 "The Jev boundary": a statement about `jev_ask` as the lane watchdog uses it; the tool stays display-only (spec §B.4) |
| `.claude/skills/moai-jev-skill-suggestion/SKILL.md:4,14,24,66`; template | the contract of a different consumer (skill suggestion), unaffected |
| `internal/mcp/catalog.go:97`; `internal/cli/mcp_jev.go:48` (description string) | tool-level; the tool stays display-only; deliberately unchanged (REQ-JAE-003) |
| `docs-site/content/en/guides/mcp-server.md:208`; `ko/ja/zh` `:212` rows | user docs of the tool; the first exception left them unchanged (spec §D) |
| `internal/cli/todo_auto.go:201,231,367`; `todo_auto_test.go:483,512,524`; `todo_auto_doc_test.go:351,352,359`; `todo_jev_finding_test.go:3` | the display-only *script* line (REQ-MT-014/015 semantics) and its tests; contract unchanged |
| `internal/jev/display_only_test.go:73` | AC-JEVC-001 test comment; the package still cannot mutate anything |
| `internal/contract/kickoff/kickoff.go:21` | describes the first amendment accurately |
| `internal/template/contract_mode_blocks_test.go:819,982,984,986`; `:1047` | the first amendment's tests (and its pinned §29 markers) |
| `AGENTS.local.md:448` (§29) | the leader's use of the local scripts; markers pinned by `contract_mode_blocks_test.go:1047` |
| `internal/mission/governance_receipt.go:34`, `governor_test.go:78`, `internal/contract/testdata/mission_surface_baseline.txt:77` | receipt content is display-only: a different object |
| `internal/closure/model.go:88`, `readiness.go:77`, `receipt_view.go:29` | closure receipt display, not Jev |
| `.claude/rules/moai/core/hooks-system.md:85`; template; `workflows/loop.md:116,221`; template; `internal/mx/provenance.go:25,73`; `internal/web/fieldsets_codex.templ:5`, `fieldsets_codex_templ.go:13`, `todo_view.go:52` | "display-only" in an unrelated sense (hook output, loop sentence, provenance timestamp, web panel) |
| `.claude/skills/moai-kanban-foreman/SKILL.md:69`; template | "serial consumption in queue order is authorized": what the batch approval grants; not a Jev statement; observed, not changed |

### R1.3 Class (ii) — historical record — left alone

| Where | Why |
|---|---|
| `CHANGELOG.md` `:24,35,52,72,124,132,205` | released-entry text |
| `.moai/project/codemaps/docs-truth.md:43` | generated map; follows `CLAUDE.md` on regeneration |
| `.moai/specs/SPEC-TODO-AUTO-PRIORITY-001/{spec,plan,progress}.md` | records the interim state as of its own time (§B.5, §D, §G R-6); superseded for the present by this SPEC's HISTORY and §A.1 |
| `.moai/specs/SPEC-AUTONOMY-CONTRACT-001/spec.md:27,92,232,239`, `SPEC-AUTONOMY-GATE-REWIRE-001/design.md` | the "until A3 reconciles the display-only doctrine" lines — A3 happened in commit `185569ef3` |
| `.moai/specs/SPEC-JEV-{CONSUMERS,GOAL-DIST,OPTIN-MEASURE,GUARD,SKILL-SUGGESTION}-001/*` | each states its own contract by reference to `SPEC-JEV-CORE-001`; no live requirement of theirs restates the principle against the exception |
| `.moai/specs/SPEC-TODO-ANALYZER-CONFORMANCE-001/spec.md:142`, `SPEC-RELATION-PICKUP-FILTER-001/spec.md:264`, `SPEC-LOOP-VERDICT-CONTRACT-001/acceptance.md:26,87` | prose citing the script line or the loop sentence |

### R1.4 Coverage check

The 43 files of the primary sweep are each assigned: class (i) — `jev.go`,
`mcp_jev.go` (also (iii) for `:48`), both `workflow.yaml`, both catalogue copies,
`agent-authoring.md` ×2, `CLAUDE.md` ×2 = 10 files; class (ii) — `docs-truth.md` = 1;
the remaining 32 are class (iii) rows above. The synonym sweep adds `SKILL.md` ×2 and
the guide (class (i)) and `AGENTS.local.md`, `contract_mode_blocks_test.go` (class
(iii), already listed). The closed-target-phrase sweep's 47 files are assigned in §R1.5
(8 + 5 + 34 = 47).

### R1.5 The closed-target-phrase sweep — 47 files, each assigned

Command (re-measured at tree `1eef55dd9`): `git grep -n -i -E 'a person reads|labelled model|queue mutation|hard to undo' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'`
→ 84 hit lines, 47 files (§R2b lists the lines).

| Class | Files | Count | Reason |
|---|---|---|---|
| (i) — already inventoried | `internal/jev/jev.go`, `internal/cli/mcp_jev.go`; `workflow.yaml` ×2; catalogue ×2 | 6 | S1-S4: the principle's own statement |
| (i) — **new, X5** | `.claude/skills/moai-ref-jev-question-design/SKILL.md` and its template mirror | 2 | the capability-level sentence; no `display-only` literal, no synonym phrase, so the first two sweeps could not see it |
| (ii) — frozen test input | the five files under `internal/cli/testdata/codex-rollouts-t1171/` (`real/rollout-….jsonl`; `roles/` and `roles-other-version/` `manager-lead.toml` and `mission-governor.toml`) | 5 | captured rollout and role definitions used as data by a different test; a snapshot of older text, not live doctrine |
| (iii-a) — predecessor-amended `--auto` text | `manager-todo.md` ×2, `.codex/agents/moai/manager-todo.toml`, `kanban-dispatch.md` ×2, `workflows/gtd.md` ×2 | 7 | amended by the predecessor card and pinned by the P8 guards |
| (iii-b) — deputy's retained powers | `manager-lead.md` ×2, `.codex/agents/moai/manager-lead.toml` | 3 | "Queue mutations — any `moai gtd` add / pick / done / edit / drop" lists what the deputy never does; not a Jev statement |
| (iii-c) — "queue mutation" about the queue store itself | `todo.go`, `todo_claim.go`, `todo_history.go`, `todo_pr.go`, `todo_show.go`, `todo_triage.go`, `todo_surface_test.go`, `factory_card.go`, `foreman_queue_watch_test.go`, `foreman_queue_watch_wal_test.go`, `moai-lane-watchdog/SKILL.md` ×2 | 12 | the phrase names writes to the todo queue (lock, claim, lane-queue write, foreman watch), not a Jev input |
| (iii-d) — other consumers' contracts and the script line | `moai-jev-skill-suggestion/SKILL.md` ×2, `todo_jev_finding.go`, `todo_auto.go`, `todo_auto_doc_test.go` | 5 | skill-suggestion and finding-mark contracts; `todo_auto.go:231-236,350-357` are the display-only *script* line, which the predecessor's own comment separates from the ranking stage ("The ranking stage's Jev consumer is a separate seam … and does not read this line") |
| (iii-e) — guards and their messages | `doctor_jev_test.go`, `contract_mode_blocks_test.go`, `internal/mission/governor_test.go` | 3 | `doctor_jev_test.go:247,310` are failure messages for files *outside* the declared consumer set (the ranking consumer is inside it, `:221`); the other two are the first amendment's tests and the manager-todo boundary pin |
| (iii-f) — tool-level docs | `docs-site/content/en/guides/mcp-server.md` | 1 | describes the tool, which stays display-only (spec §B.4, §D) |
| (iii-g) — unrelated | `moai-easy.md` ×2 ("if something's hard to undo, I'll check in with you first"), `.moai/archive/skills/v3.0/moai-platform-database-cloud/reference/firestore.md` ("Queue mutations and sync when connection re…") | 3 | the phrase in an unrelated sense |
| | | **47** | 8 + 5 + 34 (7 + 3 + 12 + 5 + 3 + 1 + 3 = 34) |

No class-(i) surface beyond X1-X5 was found, so nothing is added to the orchestrator's
decision list on this account.

## §R2 Raw sweep output (primary pattern, 69 lines, each cut at 200 characters)

```
.claude/agents/moai/manager-todo.md:7:  Jev as a display-only signal for dispatch order and priority, except for
.claude/agents/moai/manager-todo.md:58:is a permitted display-only signal for dispatch order and priority judgment.
.claude/rules/moai/core/hooks-system.md:85:| MessageDisplay | No | No | Runs while assistant message text is displayed (v2.1.152+). Returns `hookSpecificOutput.displayContent` to replace the on-screen
.claude/rules/moai/core/moai-mcp-tools-catalogue.md:135:### Judgment (gated, display-only)
.claude/rules/moai/core/moai-mcp-tools-catalogue.md:139:| `mcp__moai__jev_ask` | Ask the gated judgment capability typed questions over one supplied state; typed answers with probability | gated-unava
.claude/rules/moai/core/moai-mcp-tools-catalogue.md:233:| Judgment (gated) | `jev_ask` | gated-unavailable at the shipped default (`workflow.jev.enabled: false`) — no request constructed, no network c
.claude/rules/moai/development/agent-authoring.md:147:- manager-todo: todo-queue management agent — owns queue lifecycle, the `/moai:todo --auto` serial cycle, and dispatch guidance; consults Jev as a
.claude/rules/moai/workflow/auto-semantics.md:217:operator's act). jev_ask is a **display-only** judgment query: it observes
.claude/skills/moai-jev-skill-suggestion/SKILL.md:4:  Guidance for a display-only skill-suggestion capability: how a caller
.claude/skills/moai-jev-skill-suggestion/SKILL.md:14:  display-only contracts, or anyone diagnosing a suggestion that appeared to
.claude/skills/moai-jev-skill-suggestion/SKILL.md:24:# Skill Suggestion, Display-Only
.claude/skills/moai-jev-skill-suggestion/SKILL.md:66:  breaks the display-only chain: a skill that never appeared in context has
.claude/skills/moai/workflows/gtd.md:354:display-only signal, and Jev's ranking answer is used for the selection order
.claude/skills/moai/workflows/loop.md:116:- The completion sentence "All loop completion conditions satisfied; exiting loop." is DISPLAY-ONLY (emitted by Step 4 as a report string) — it carries no exi
.claude/skills/moai/workflows/loop.md:221:- Mechanical predicate confirmed twice: the previous iteration's parsed diagnostics satisfy zero errors + tests passing + coverage threshold (Step 1) AND the 
.moai/config/sections/workflow.yaml:227:    # The capability is display-only: an answer is a labelled model-produced
.moai/project/codemaps/docs-truth.md:43:| 12 | `manager-todo` | MoAI-custom — no Selection Decision Tree row | Todo-queue management (queue lifecycle, `/moai:todo --auto` serial cycle, dispatch guidan
CLAUDE.md:63:**Retained agents (13)**: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`, `super-advisor`, `manager-design`, `e2e-test
docs-site/content/en/guides/mcp-server.md:208:### Judgment (gated, display-only)
internal/cli/mcp_jev.go:8:// gated-unavailable. The capability is display-only: an answer is a labelled
internal/cli/mcp_jev.go:48:		mcp.WithDescription("Ask the gated TypeSafe System One judgment capability one typed question set over one supplied state; returns typed answers with the model's probabili
internal/cli/todo_auto.go:201:	jev       func(root string) string // display-only consultation seam (tests stub it)
internal/cli/todo_auto.go:231:	// The script consultation is display-only (REQ-MT-014/015): its signal is
internal/cli/todo_auto.go:367:	return "jev signal (display-only): " + strings.TrimSpace(string(out))
internal/cli/todo_auto_doc_test.go:351:	// display-only signal for dispatch order and priority — never as authority".
internal/cli/todo_auto_doc_test.go:352:	"consults the Jev judgment scripts as a display-only signal for dispatch order and priority",
internal/cli/todo_auto_doc_test.go:359:	"is a permitted display-only signal for dispatch order and priority judgment.",
internal/cli/todo_auto_test.go:483:			return "jev signal (display-only): MUTATE " + rec.Items[0].ID
internal/cli/todo_auto_test.go:512:// display-only signal.
internal/cli/todo_auto_test.go:524:	if !strings.Contains(got, "display-only") || !strings.Contains(got, "LOCAL-SIGNAL") {
internal/cli/todo_jev_finding_test.go:3:// write-path `agent` prohibition, and the display-only queue-hash guard.
internal/closure/model.go:88:// the receipt (REQ-CLOSURE-010). Receipt content is display-only: no readiness
internal/closure/readiness.go:77:	// ContractSecondModel is contract review.second_model (display only).
internal/closure/receipt_view.go:29:// hard-codes no set. Receipt content is display-only — no readiness decision
internal/contract/kickoff/kickoff.go:21:// JevDoctrineAmended is true once the Jev display-only principle is amended
internal/contract/testdata/mission_surface_baseline.txt:77:    the receipt records what was displayed, and a display-only signal never
internal/jev/display_only_test.go:73:// AC-JEVC-001 — the display-only invariant, measured the way the SPEC names:
internal/jev/jev.go:24://   - **The capability is display-only.** A Jev answer is a labelled,
internal/mcp/catalog.go:97:	// Gated judgment wrapper (display-only): registered unconditionally so its
internal/mission/governance_receipt.go:34:// records what was displayed, and a display-only signal never becomes a
internal/mission/governor_test.go:78:		"display-only",
internal/mx/provenance.go:25:// @MX:NOTE: [AUTO] Provenance — GeneratedAt is display-only; freshness never reads wall-clock time (mtime and timestamps are banned staleness signals per REQ-GF-002)
internal/mx/provenance.go:73:	// GeneratedAt is RFC3339 display-only metadata — never a freshness input.
internal/template/contract_mode_blocks_test.go:819:// The amendment opens exactly one exception to the Jev display-only
internal/template/contract_mode_blocks_test.go:982:	t.Run("falsifier/req-011-display-only", func(t *testing.T) {
internal/template/contract_mode_blocks_test.go:984:		bad := strings.Replace(spec, body, " (Ubiquitous) "+grJevAmended+" — v0.3.0] A Jev answer shall not mutate anything. The capability is display-only
internal/template/contract_mode_blocks_test.go:986:			t.Fatal("checker accepted REQ-JEVC-011 left display-only with no exception")
internal/template/templates/.claude/agents/moai/manager-todo.md:7:  Jev as a display-only signal for dispatch order and priority, except for
internal/template/templates/.claude/agents/moai/manager-todo.md:58:is a permitted display-only signal for dispatch order and priority judgment.
internal/template/templates/.claude/rules/moai/core/hooks-system.md:85:| MessageDisplay | No | No | Runs while assistant message text is displayed (v2.1.152+). Returns `hookSpecificOutput.displayConte
internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md:135:### Judgment (gated, display-only)
internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md:139:| `mcp__moai__jev_ask` | Ask the gated judgment capability typed questions over one supplied state; typed answers wi
internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md:233:| Judgment (gated) | `jev_ask` | gated-unavailable at the shipped default (`workflow.jev.enabled: false`) — no reque
internal/template/templates/.claude/rules/moai/development/agent-authoring.md:147:- manager-todo: todo-queue management agent — owns queue lifecycle, the `/moai:todo --auto` serial cycle, and dispatch
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:217:operator's act). jev_ask is a **display-only** judgment query: it observes
internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md:4:  Guidance for a display-only skill-suggestion capability: how a caller
internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md:14:  display-only contracts, or anyone diagnosing a suggestion that appeared to
internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md:24:# Skill Suggestion, Display-Only
internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md:66:  breaks the display-only chain: a skill that never appeared in context has
internal/template/templates/.claude/skills/moai/workflows/gtd.md:354:display-only signal, and Jev's ranking answer is used for the selection order
internal/template/templates/.claude/skills/moai/workflows/loop.md:116:- The completion sentence "All loop completion conditions satisfied; exiting loop." is DISPLAY-ONLY (emitted by Step 4 as a report
internal/template/templates/.claude/skills/moai/workflows/loop.md:221:- Mechanical predicate confirmed twice: the previous iteration's parsed diagnostics satisfy zero errors + tests passing + coverage threshold (Step 1) AND the 
internal/template/templates/.codex/agents/moai/manager-todo.toml:9:Jev as a display-only signal for dispatch order and priority, except for
internal/template/templates/.codex/agents/moai/manager-todo.toml:57:is a permitted display-only signal for dispatch order and priority judgment.
internal/template/templates/.moai/config/sections/workflow.yaml:229:    # The capability is display-only: an answer is a labelled model-produced
internal/template/templates/CLAUDE.md:63:**Retained agents (13)**: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`, `super-advisor`,
internal/web/fieldsets_codex.templ:5:// Every element here is display-only. The panel renders NO form element
internal/web/fieldsets_codex_templ.go:13:// Every element here is display-only. The panel renders NO form element
internal/web/todo_view.go:52:// display-only rendering of the recorded findings that name this card
```

Synonym sweep, 5 lines (cut at 170 characters):

```
.claude/skills/moai/SKILL.md:180:The pick is the operator's: never preselect, never reorder by inferred priority, never auto-populate from TODO comments or issues.
.moai/docs/jev-local-operations.md:28:### [HARD] 되돌릴 수 없는 판정은 모델 답을 입력으로도 쓰지 않는다
AGENTS.local.md:448:[HARD] `scripts/jev/`는 이 저장소의 로컬 도구이며 제품에 배선되지 않았다. `-k`/`-f` 리더는 묵은 카드 배차 전에 `scripts/jev/triage.sh <id>`, 레인 질문으로 멈췄을 때 `scripts/jev/route.sh < 질문`을
internal/template/contract_mode_blocks_test.go:1047:		for _, marker := range []string{"판단 자료일 뿐", "판정 근거로 쓰지 않는다"} {
internal/template/templates/.claude/skills/moai/SKILL.md:180:The pick is the operator's: never preselect, never reorder by inferred priority, never auto-populate from TOD
```

## §R2b Raw output of the closed-target-phrase sweep (84 lines, each cut at 200 characters)

Command: `git grep -n -i -E 'a person reads|labelled model|queue mutation|hard to undo' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'`
(captured at tree `c50da9c2f` in the revision session; re-counted at `1eef55dd9`: 84 lines, 47 files).

```
.claude/agents/moai/manager-lead.md:231:4. **Queue mutations** — any `moai gtd` add / pick / done / edit / drop. Deputy dispatch is limited to ALREADY-PICKED cards; admission and closure are
.claude/agents/moai/manager-todo.md:63:basis of a queue mutation, a completion verdict, a merge approval, or any
.claude/output-styles/moai/moai-easy.md:40:4. **Confirm before big moves** — if something's hard to undo, I'll check in with you first
.claude/rules/moai/core/moai-mcp-tools-catalogue.md:139:| `mcp__moai__jev_ask` | Ask the gated judgment capability typed questions over one supplied state; typed answers with probability | g
.claude/rules/moai/core/moai-mcp-tools-catalogue.md:233:| Judgment (gated) | `jev_ask` | gated-unavailable at the shipped default (`workflow.jev.enabled: false`) — no request constructed, no
.claude/rules/moai/workflow/kanban-dispatch.md:25:[HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns the request into a card with `moai gtd add "<descrip
.claude/rules/moai/workflow/kanban-dispatch.md:33:[HARD] **The self-dispatch lane exception.** In a self-dispatch factory run, a lane session may lease the next queued card — the one promoti
.claude/skills/moai-jev-skill-suggestion/SKILL.md:47:A suggestion is a signal a person reads: a labelled model answer that names
.claude/skills/moai-jev-skill-suggestion/SKILL.md:56:  queue mutation — anything hard to undo stays with people.
.claude/skills/moai-lane-watchdog/SKILL.md:116:lane does NOT run it (queue mutation is prohibited). The queue-level pickup
.claude/skills/moai-lane-watchdog/SKILL.md:170:zero queue mutation, and never picks, drops, edits, or unrelates anything.
.claude/skills/moai-ref-jev-question-design/SKILL.md:27:nothing: the answer is a labelled model signal a person reads, never a
.claude/skills/moai-ref-jev-question-design/SKILL.md:28:completion predicate, a merge approval, a queue mutation, or any other
.claude/skills/moai-ref-jev-question-design/SKILL.md:29:decision that is hard to undo. Good questions keep that contract; bad
.claude/skills/moai/workflows/gtd.md:271:  the `machine-only` mark, and it is a record a person reads — never an input
.claude/skills/moai/workflows/gtd.md:272:  to a queue mutation or a completion verdict. A finding leaves the file when
.claude/skills/moai/workflows/gtd.md:355:only; it is never the basis of a queue mutation or a completion verdict. See
.moai/archive/skills/v3.0/moai-platform-database-cloud/reference/firestore.md:65:Cache Throttling implements custom throttling for offline writes. Queue mutations and sync when connection re
.moai/config/sections/workflow.yaml:227:    # The capability is display-only: an answer is a labelled model-produced
.moai/config/sections/workflow.yaml:228:    # signal a person reads, never an input to a completion verdict, a merge
.moai/config/sections/workflow.yaml:229:    # approval, a queue mutation, or any other decision that is hard to undo.
docs-site/content/en/guides/mcp-server.md:214:The tool is always registered, but with the gate off it builds no request and makes no network call. Its answer is a signal for a person to read
internal/cli/doctor_jev_test.go:247:		t.Errorf("internal/cli files outside the declared consumer set import internal/jev: %v — a completion verdict, a merge approval, a queue mutation, an op
internal/cli/doctor_jev_test.go:310:				t.Errorf("%s imports internal/jev — a queue mutation, verdict, or integration-window surface MUST NOT reach the call path (REQ-JEVC-012)", name)
internal/cli/factory_card.go:58:// @MX:REASON: 4 call sites (factory_card.go, mcp_factory_card.go, mcp_todo.go, todo.go); widening one path alone would let a lane queue mutation slip past it
internal/cli/mcp_jev.go:9:// model signal a person reads, never a completion verdict, a merge approval,
internal/cli/mcp_jev.go:10:// a queue mutation, or any other decision that is hard to undo.
internal/cli/mcp_jev.go:48:		mcp.WithDescription("Ask the gated TypeSafe System One judgment capability one typed question set over one supplied state; returns typed answers with the model's
internal/cli/testdata/codex-rollouts-t1171/real/rollout-2026-09-24T18-41-35-01a0d2ca-b7e2-7e50-8f7f-0ab053beb07f.jsonl:9:{"timestamp":"2026-09-24T09:41:35.997Z","ordinal":8,"type":"response_
internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-lead.toml:227:4. **Queue mutations** — any `moai gtd` add / pick / done / edit / drop. Deputy dispatch is limited to AL
internal/cli/testdata/codex-rollouts-t1171/roles-other-version/mission-governor.toml:10:NOT for: writing files, shell or Git execution, queue mutation, dispatch, merge, approval, or PASS/FAI
internal/cli/testdata/codex-rollouts-t1171/roles/manager-lead.toml:226:4. **Queue mutations** — any `moai gtd` add / pick / done / edit / drop. Deputy dispatch is limited to ALREADY-PICKED c
internal/cli/testdata/codex-rollouts-t1171/roles/mission-governor.toml:10:NOT for: writing files, shell or Git execution, queue mutation, dispatch, merge, approval, or PASS/FAIL audit verdic
internal/cli/todo.go:1378:	// Queue mutations resolve through the primary checkout, but provenance must
internal/cli/todo_auto.go:233:	// never a queue mutation, a completion verdict, a merge approval, or an
internal/cli/todo_auto.go:353:// a queue mutation, a completion verdict, a merge approval, or any
internal/cli/todo_auto_doc_test.go:153:	"it is never the basis of a queue mutation or a completion verdict.",
internal/cli/todo_auto_doc_test.go:357:// the basis of a queue mutation or a completion verdict.
internal/cli/todo_auto_doc_test.go:361:	"It is never the basis of a queue mutation, a completion verdict, a merge approval, or any operator-gate decision.",
internal/cli/todo_claim.go:7:// explicit root. The verb is a queue MUTATION — it is deliberately absent
internal/cli/todo_history.go:86:The verb changes no card or schema and takes no queue mutation lock. SQLite
internal/cli/todo_jev_finding.go:5:// a RECORD a person reads, never a decision anything acts on (REQ-JEVC-011 /
internal/cli/todo_jev_finding.go:8:// approval, a queue mutation, or an operator gate. The only write is one
internal/cli/todo_pr.go:5:// migration and takes no queue mutation lock. SQLite may use transient
internal/cli/todo_pr.go:144:and takes no queue mutation lock. SQLite may use transient coordination files
internal/cli/todo_show.go:43:The verb is read-only: it changes no card, takes no queue mutation lock,
internal/cli/todo_surface_test.go:84://     schema and takes no queue mutation lock, which
internal/cli/todo_triage.go:14:// schema, performs no migration, and takes no queue mutation lock. It reads
internal/cli/todo_triage.go:123:and takes no queue mutation lock.
internal/jev/jev.go:25://     model-produced signal a person reads (REQ-JEVC-013). Nothing in this
internal/jev/jev.go:205:// Label renders the answer as a labelled model signal (REQ-JEVC-013). It names
internal/kanban/foreman_queue_watch_test.go:5:// producing zero events across a real queue mutation on a migrated project:
internal/kanban/foreman_queue_watch_test.go:205:				if _, _, err := store.Add("fixture card two — queue mutation under watch"); err != nil {
internal/kanban/foreman_queue_watch_test.go:210:				t.Errorf("no change event within %s across a real queue mutation — the %s watch does not observe the authoritative store", watchWindow, na
internal/kanban/foreman_queue_watch_test.go:228:		if _, _, err := store.Add("fixture card two — queue mutation under watch"); err != nil {
internal/kanban/foreman_queue_watch_test.go:250:		if _, _, err := store.Add("fixture card two — queue mutation under watch"); err != nil {
internal/kanban/foreman_queue_watch_wal_test.go:3:// to backlog.db-wal must not hide a queue mutation from the foreman watch.
internal/mission/governor_test.go:52:// agent body: the grade-3 authority list (queue mutation, completion verdict,
internal/mission/governor_test.go:68:		"queue mutation",
internal/template/contract_mode_blocks_test.go:993:		if f := grJevNoteFindings("fixture", passage, []string{"merge approval", "queue mutation"}); len(f) == 0 {
internal/template/contract_mode_blocks_test.go:1008:			closed := []string{"completion verdict", "merge", "queue mutation"}
internal/template/contract_mode_blocks_test.go:1017:			for _, f := range grJevNoteFindings(label, grJevYAMLComment(wf), []string{"completion verdict", "merge", "queue mutation"}) {
internal/template/templates/… (the template mirrors of the rows above, same lines: manager-lead.md:233, manager-todo.md:63, moai-easy.md:40, the catalogue :139/:233, kanban-dispatch.md:25/:33, moai-jev-skill-suggestion SKILL.md:47/:56, moai-lane-watchdog SKILL.md:116/:170, moai-ref-jev-question-design SKILL.md:27-29, workflows/gtd.md:271/:272/:355, .codex/agents/moai/manager-lead.toml:228, manager-todo.toml:62, workflow.yaml:229-231)
```

The last line of the block is an abridgement, not command output: the 22 template-mirror
lines (and the 84-line total) are in the command's output; the abridgement lists them
file by file so this block stays readable.

## §R3 Baselines at `c50da9c2f` (all PASS)

Each command was `unset <11 variables> && go test -count=1 -v -run '<anchored>' ./internal/<pkg>/`.

```
./internal/contract/kickoff/   -run '^TestJevAmendmentLinkage$'
--- PASS: TestJevAmendmentLinkage (1.19s)
    --- PASS: TestJevAmendmentLinkage/falsifier/markers-only (0.18s)
    --- PASS: TestJevAmendmentLinkage/falsifier/constant-only (0.14s)
    --- PASS: TestJevAmendmentLinkage/falsifier/all-but-sec-29 (0.14s)
    --- PASS: TestJevAmendmentLinkage/all-in-one-commit (0.15s)
    --- PASS: TestJevAmendmentLinkage/migrated-guide (0.25s)
    --- PASS: TestJevAmendmentLinkage/tree (0.34s)      [log: JevDoctrineAmended = true]
ok  	github.com/modu-ai/moai-adk/internal/contract/kickoff	1.660s

./internal/template/           -run '^TestJevDoctrineAmendment$'
--- PASS: TestJevDoctrineAmendment (0.00s)
    --- PASS: TestJevDoctrineAmendment/falsifier/authority-items-unamended (0.00s)
    --- PASS: TestJevDoctrineAmendment/falsifier/req-011-display-only (0.00s)
    --- PASS: TestJevDoctrineAmendment/falsifier/note-widened (0.00s)
    --- PASS: TestJevDoctrineAmendment/spec (0.00s)
    --- PASS: TestJevDoctrineAmendment/rules-and-config (0.00s)
    --- PASS: TestJevDoctrineAmendment/local-guide (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.378s

./internal/cli/                -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure|TestAutoHelpAndRefusalDoNotAssertPickOrder|TestAutoRankAgentDoctrine|TestMCPToolCatalogueDocsStayMirrorIdentical)$'
--- PASS: TestMCPToolCatalogueDocsStayMirrorIdentical (0.00s)
--- PASS: TestAutoRankDoctrineAmendment (0.00s)
--- PASS: TestAutoRankMirrorParity (0.00s)
--- PASS: TestAutoRankMarkerDisclosure (0.01s)
--- PASS: TestAutoHelpAndRefusalDoNotAssertPickOrder (0.31s)
--- PASS: TestAutoRankAgentDoctrine (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.947s

./internal/cli/                -run '^(TestAutoRankJevOrdering|TestAutoRankBlockedExcluded|TestAutoRankJevMalformedAnswer|TestAutoRankQueueUnchanged|TestAutoRankNoQueueWriteGuard|TestAutoRankRescueFirst|TestAutoLiveJevRanker|TestAutoDefaultSeamsAreInert)$'
--- PASS: TestAutoRankBlockedExcluded (0.53s)
--- PASS: TestAutoRankJevOrdering (0.78s)
--- PASS: TestAutoRankJevMalformedAnswer (2.50s)
--- PASS: TestAutoRankRescueFirst (0.42s)
--- PASS: TestAutoRankQueueUnchanged (0.37s)
--- PASS: TestAutoRankNoQueueWriteGuard (0.00s)
--- PASS: TestAutoLiveJevRanker (0.03s)
--- PASS: TestAutoDefaultSeamsAreInert (0.27s)
ok  	github.com/modu-ai/moai-adk/internal/cli	5.565s

./internal/jevmeasure/         -run '^TestNoConsumerCallPathShips$'
=== RUN   TestNoConsumerCallPathShips
--- PASS: TestNoConsumerCallPathShips (0.19s)
ok  	github.com/modu-ai/moai-adk/internal/jevmeasure	0.461s

./internal/jev/                -run '^(TestPackageImports_AreStandardLibraryOnly|TestImportClassifierPositiveControl)$'
--- PASS: TestPackageImports_AreStandardLibraryOnly (0.00s)
--- PASS: TestImportClassifierPositiveControl (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/jev	0.295s
```

`moai spec lint`, run with a tree-local binary (`go build -o <scratchpad>/moai ./cmd/moai`
at HEAD `c50da9c2f`; the binary reports `moai-adk v3.1.3  none built unknown` because no
ldflags were passed; its output was piped through `tail`, so the quoted line is the
evidence and the exit code is `tail`'s):

```
$ <scratchpad>/moai spec lint SPEC-JEV-CORE-001      → ✓ No findings — all SPEC documents are valid
$ <scratchpad>/moai spec lint SPEC-MANAGER-TODO-001  → ✓ No findings — all SPEC documents are valid
$ <scratchpad>/moai spec lint SPEC-TODO-AUTO-PRIORITY-001 → ✓ No findings — all SPEC documents are valid
```

## §R4 Quoted source evidence

**The consumer calls the client, not the tool** (`internal/cli/todo_auto_rank.go`):
`:467 enabled, err := jevEnabled(resolveProjectDir())`, `:474 client := jev.New(true)`,
`:475 client.LoadCredential = jevcred.Load`, `:481 return client.Ask(ctx, req)`. The tool
handler is `handleJevAsk` (`internal/cli/mcp_jev.go:67`); `git grep -n 'handleJevAsk' --
internal/cli/mcp_jev.go internal/cli/mcp_server.go` lists `mcp_jev.go` lines `62, 65, 67`
only (the registration, an MX comment, the definition) and nothing in `mcp_server.go`.

**The consumer-set guard already declares the consumer** (`internal/cli/doctor_jev_test.go:221`,
the allow-list of `TestJevCallPath_HasExactlyTheDeclaredConsumers`):
`"todo_auto_rank.go": "SPEC-TODO-AUTO-PRIORITY-001 M2 — the \`todo --auto\` selection-order
consumer (inert behind workflow.jev.enabled; a Jev answer orders candidates and nothing
else; no ordering accuracy is claimed)"`. The sibling guard
`TestJevCallPath_UnreachableFromDecisionSurfaces` names `gtd.go`, `todo_analysis.go`,
`todo_autodone.go`, `integration.go`, `integration_settings_drift.go` — not
`todo_auto.go` or `todo_auto_rank.go`.

**Why REQ-JEVC-012 needs the exception.** `spec.md:99`: "The system shall not consult
Jev — not as a decision, and **not as an input** — for any of: a completion verdict, a
merge approval, a `moai todo` or `moai gtd` mutation, an operator gate, …"; `:103-108`
define "for" by the role the answer plays, with a mechanical discriminator
"(1) Determined first. (2) Unread after." The ranking key fails limb (2): the order
derived from the answer is read on the pick path. That is why the exception is *named*,
as the Kickoff cross-check was ("outside the … discriminator below by design — there the
answer does reach the gate, which is why the exception is named rather than derived").
`spec.md:160-161` (the second bullet, "Still excluded after the v0.2.0 and v0.3.0
amendments") also bars "a Jev answer that a queue mutation … **selects on**".

**The pinned guide paragraph** (`SPEC-AUTONOMY-GATE-REWIRE-001/design.md:296`, between
`<!-- §29-amendment-text-start -->` and `-end -->`, required verbatim in
`.moai/docs/jev-local-operations.md` by `TestJevDoctrineAmendment/local-guide`):
`예외는 한 곳뿐이다 — contract 모드 Kickoff 의 \`llm+jev\` 교차 확인에서는 \`moai contract
decide\` 가 Jev 를 두 번째 신호로 직접 부를 수 있다. Jev 답은 LLM 의 승인을 확인하거나
사람에게 보낼 뿐 혼자서 시작시키지 못하며, 그 밖의 3등급 항목에 대한 금지는 그대로다.`
(one line in the source).

**The first guard's marker list** (`internal/contract/kickoff/activation_test.go:155-163`):
`SPEC-JEV-CORE-001/spec.md` ← `[AMENDED 2026-09-26`; both catalogue copies and both
`workflow.yaml` copies ← `contract-mode Kickoff`; the guide (`CLAUDE.local.md`, or
`.moai/docs/jev-local-operations.md` once `AGENTS.local.md` exists) ← `` `moai contract
decide` 가 Jev 를 두 번째 신호로 ``; `internal/contract/kickoff/kickoff.go` ←
`JevDoctrineAmended = true`. `firstCommit` = the oldest commit in
`git log --reverse --format=%H -S<token> -- <path>`.

**The first amendment's assembly** (`SPEC-AUTONOMY-GATE-REWIRE-001/design.md:300-305`):
manager-spec writes the SPEC body and returns; manager-develop then writes code, rules,
config and the guide line; the lane orchestrator stages both by explicit path into one
commit; verification scope `go test ./internal/contract/... ./internal/template/
./internal/spec/` and `moai spec lint SPEC-JEV-CORE-001`.

**Gate default**: `internal/config/defaults.go:1180-1182` — `Jev: WorkflowJevConfig{
Enabled: false, },` under the comment "Template neutrality: no `enabled: true` under
internal/template/templates/".

**Catalog hashing** (`internal/template/catalog.yaml`): 49 `hash:` entries, the first for
skill `moai` (`path: templates/.claude/skills/moai/`). `Makefile:35`: `build:` runs
`@go run ./internal/template/scripts/gen-catalog-hashes.go --all`.

**Zone/constitution check**: `git grep -c -i -E 'jev|display-only' -- .claude/rules/moai/core/zone-registry.md .claude/rules/moai/core/moai-constitution.md`
printed nothing (exit 1): no Frozen or constitutional clause names the principle.

**Reference-skill call-path guard** (`internal/cli/jev_question_design_skill_test.go:17-25`):
the forbidden tokens `internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev`, scanned in
both copies of the skill, plus `TestJevQuestionDesignSkillCopiesStayIdentical`
(`bytes.Equal` of the two copies, `:77-90`).

**Base-ref guards** (`internal/template/contract_mode_guided_test.go`): `grBaseEnv =
"MOAI_GR_BASE"` (`:30`); `t.Skipf` when unset (`:34-36`); the always-loaded growth cap
1,500 characters (`:617`) and the catalogue-file cap 600 characters (`:633`); the
change-set allowlist naming `SPEC-JEV-CORE-001/spec.md` (`:260`).

## §R3b Baselines added in the revision (tree `1eef55dd9`, no Go or template file differs from `c50da9c2f`)

```
./internal/cli/   -run '^(TestJevQuestionDesignSkillCarriesNoCallPath|TestJevQuestionDesignSkillCopiesStayIdentical)$'
--- PASS: TestJevQuestionDesignSkillCarriesNoCallPath (0.00s)
--- PASS: TestJevQuestionDesignSkillCopiesStayIdentical (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.983s

./internal/spec/  -run '^TestCatalogHashParity$'
    catalog_hash_test.go:200: verified 49 catalog entries against normalized source bodies — 0 drift
--- PASS: TestCatalogHashParity (0.05s)
ok  	github.com/modu-ai/moai-adk/internal/spec	0.482s
```

## §R5 Ledger re-execution (revision, tree `1eef55dd9`; working tree = HEAD plus this revision's uncommitted SPEC edits)

Each row is the `acceptance.md` ledger command, executed once as a plain command, with
its stdout and exit code. Results are identical to the pinned tree's.

| Row | Observed stdout | Exit | Matches the pinned-tree result |
|---|---|---|---|
| L1, L2, L3, L4, L5, L6, L6b | (empty) | 1 each | yes |
| L7 (`go test -list`, scrubbed) | `ok  	github.com/modu-ai/moai-adk/internal/template	0.413s` — no test name; the elapsed figure differs per run (0.292s, 0.169s, 0.413s observed), which is why L7 is informational | 0 | yes, modulo timing |
| L8 | `.claude/agents/moai/manager-todo.md:2` · `.claude/rules/moai/workflow/kanban-dispatch.md:1` · `.claude/skills/moai/workflows/gtd.md:1` | 0 | yes |
| L9 | `.moai/config/sections/workflow.yaml:227:    # The capability is display-only: an answer is a labelled model-produced` · `internal/template/templates/.moai/config/sections/workflow.yaml:229:    # The capability is display-only: an answer is a labelled model-produced` | 0 | yes |
| L10 | `.claude/agents/moai/manager-todo.md:7:  Jev as a display-only signal for dispatch order and priority, except for` | 0 | yes |
| L11 | `internal/cli/mcp_jev.go:1` | 0 | yes |
| L12, L13 | (empty) | 1 each | yes |
| L14 | `.moai/specs/SPEC-JEV-CORE-001/spec.md:4:version: "0.3.0"` | 0 | yes |
| L15 | `internal/jev/jev.go:1` | 0 | yes |
| L16, L17, L19 | (empty) | 1 each | new rows |
| L18 | `internal/template/contract_mode_blocks_test.go:2` | 0 | new row |
| C1 | `internal/cli/mcp_jev.go:1` · `internal/jev/jev.go:1` | 0 | new row |
| C2 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md:2` · `.moai/specs/SPEC-JEV-CORE-001/spec.md:4` · `.moai/specs/SPEC-MANAGER-TODO-001/spec.md:1` · `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md:2` | 0 | new row |
| C3 | `.moai/docs/jev-local-operations.md:1` | 0 | new row |
| C4 | six lines, `:1` each (the three live and three template extension paths) | 0 | new row |
| C5 | `.claude/skills/moai-ref-jev-question-design/SKILL.md:1` · `internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md:1` | 0 | new row |

`git diff --name-only c50da9c2f 1eef55dd9` → the five files of
`.moai/specs/SPEC-JEV-AUTO-EXCEPTION-001/` (acceptance, plan, progress, research, spec);
no ledger pathspec lies there, so no row's result could move between the two trees.

## §R6 The D1 draft — arming token built by parts versus spelled as a literal

A throwaway Go program in the session scratchpad (outside the repository, not committed)
re-implements the predecessor's mechanics: presence = `strings.Contains` over the file
bytes (`activation_test.go:179`); first-commit = `git log --reverse --format=%H -S<token> -- <path>`
(`:44-50`); the all-or-none and one-commit checks of `linkageFindings`. It builds a
temporary repository with a registry of three rows — the arming token in a stand-in for
the test file, and the universal token in two stand-in documents — and runs each
scenario. Verbatim output of `go run` on it (the scrub prefix applied):

```
design (a): token assembled from parts
  at G: findings=[]
  at G: git grep -c -F <armTok> <G> -- test file -> stdout="" err=exit status 1
  at K: findings=[]
  at K: git grep -c -F <armTok> <K> -- test file -> stdout="7fbdf0c3515be3fdd8e671e12372a59c6a8a9a9a:internal/template/jev_auto_exception_test.go:1" err=<nil>
mutant: registry literal written in the file under test (self-match)
  at G: findings=[partial amendment: present in internal/template/jev_auto_exception_test.go; absent in a/doc1.md, b/doc2.md]
  at G: git grep -c -F <armTok> <G> -- test file -> stdout="a92eada8a58a3f13fecb809b78fb91152dcf3e94:internal/template/jev_auto_exception_test.go:1" err=<nil>
  at K: findings=[marker in a/doc1.md first appears in 78b6238, not a92eada marker in b/doc2.md first appears in 78b6238, not a92eada]
  at K: git grep -c -F <armTok> <K> -- test file -> stdout="78b623860c3ddb8f5f1a6c8fc35e2cb52f84019f:internal/template/jev_auto_exception_test.go:2" err=<nil>
partial: const flips, docs do not
  findings=[partial amendment: present in internal/template/jev_auto_exception_test.go; absent in a/doc1.md, b/doc2.md]
split: docs land in a later commit than the const
  findings=[marker in a/doc1.md first appears in e328fc3, not 5677dc7 marker in b/doc2.md first appears in e328fc3, not 5677dc7]
claim/over-reach regexp:
  ok       flagged=true  want=true   "which sets its selection order only, and the ordering is accurate"
  ok       flagged=true  want=true   "selection order only; the ordering beats the fallback"
  ok       flagged=true  want=true   "with a measured accuracy of 80%"
  ok       flagged=false want=false  "no ordering accuracy is claimed"
  ok       flagged=false want=false  "and is not claimed to be accurate"
  ok       flagged=false want=false  "the answer is never the basis of a completion verdict, a merge approval or any other decision that is hard to undo"
  ok       flagged=true  want=true   "it applies to every moai todo pick"
  ok       flagged=false want=false  "any other decision that is hard to undo"
```

(The SHAs are those of the throwaway repositories; they differ per run.) Reading: by
parts, the checker returns no finding at `G` or at `K`, the literal is absent at `G`
(`git grep` exit 1) and occurs once at `K`; spelled as a literal, the checker reports a
partial amendment at `G`, and at `K` the arming marker first appears in `G`'s commit and
occurs twice. The draft is a re-implementation, not the real guard — see `plan.md` §I
G-7. The claim and over-reach expressions are the two reference expressions of `plan.md`
§E; the eight cases are the full set they were tried on.
