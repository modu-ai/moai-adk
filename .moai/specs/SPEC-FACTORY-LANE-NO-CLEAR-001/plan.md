# SPEC-FACTORY-LANE-NO-CLEAR-001 — Implementation Plan

## §A Context

- Tier M (3 tier artifacts: spec.md, plan.md, acceptance.md), Class C (a design change across subsystems: doctrine, SessionStart text, factory completion output, todo lane cycle, agent definitions), Template-First, priority P1.
- Open decisions are recorded in `decision-index.md` (Q1 through Q8). Verdict lines are empty at authoring. Where a decision is pending, this plan states the default the SPEC encodes and the alternate the operator may select.
- Template source of truth: `internal/template/templates/`. Build gate: `make build`. Local parity: the dogfood tree (`.claude/`, `.moai/policies/`) is brought to byte-identity with the template for every changed artifact, except where a decision row says otherwise.

## §B Known Issues (observed in the research step)

1. Local and template copies of `factory-dispatch.md` differ by one appended sentence at line 220 (the local copy carries it; the template does not). Decision Q3 (POLICY-COVERED: carry it into the template).
2. The `.moai/policies/workflow/` copies are not a clean path transform of the rule files. Diff line counts after the transform: `factory-dispatch.md` 198, `factory-dispatch-mechanics.md` 69, `factory-dispatch-detail.md` 4, `factory-dispatch-cards.md` 0. Decision Q5.
3. The root `.codex/agents/moai/manager-lead.toml` differs from the template TOML in 106 lines, and the root directory has no `manager-todo.toml`. The codex golden check pins only the template tree. Decision Q4.
4. The role-core test `TestInjectedCoreCarriesObligationElements` checks the whole rule body for the ALB-0294 phrases (`never before`, `returns to the card worktree`). The `never before` row must be replaced by the new transition phrase.
5. One printer, `factoryPrintClearPolicyLine`, produces the lane completion line and has three callers (two in `factoryCompleteCard`, one in `factoryObservePRMerge`). Its tests are `TestSD_AC020_ClearPolicies` and `TestSD_M6_ClearPolicyCarrierConstantOnly` in `internal/cli/factory_m6_test.go`.
6. The card's stated branch name (`t1793`) differs from the worktree's actual branch (`WT-lane-no-clear`). The factory branch rule forbids the card id in the branch name, so the commit goes to `WT-lane-no-clear`.
7. Session memory records one observation of the in-session switch sequence. It is not committed evidence (decision Q7).

## §C Pre-flight (run phase)

- Confirm the worktree root and branch: `git -C <worktree-absolute-path> rev-parse --show-toplevel` and `git -C <worktree-absolute-path> branch --show-current`. Expected branch: `WT-lane-no-clear`.
- Confirm a clean start before any edit: `git -C <worktree-absolute-path> status --short` prints nothing.
- Use literal absolute paths inside the worktree session. The worktree-session guard refuses shell loops and `sed` invocations whose paths are computed from runtime variables. Expand each path literally.
- Targeted test packages only. Do not run `go test ./...`. Do not start background load.

## §D Milestones (priority order, decision-reversibility first)

Order follows the reversibility of the decision each milestone carries: the highest-change-likelihood items first (the lane flow text and the lane-facing outputs), then the leader and operator text, then the doctrine copies, and the mechanical parity and verification steps last.

- **M1 — Lane transition flow and the open decisions (Priority High).** Fix the transition sequence in the doctrine text (lease, exit keeping the tree, enter the leased tree, verify the toplevel, proceed; relaunch only as the fallback of REQ-LNC-004). Decisions that bind this text (Q1, Q2, Q5, Q6, Q7) are recorded and their defaults applied.
- **M2 — Lane-facing outputs and SessionStart text (Priority High).** The clear-policy printer, the next-card and stall-recheck texts in four locales, the code comments that describe the old loop, and the tests that assert the old output.
- **M3 — Leader, manager, and operator-session text (Priority High).** The manager-lead description and return template, the manager-todo and gtd scoping sentences, and the regenerated codex TOMLs.
- **M4 — Doctrine family in the template tree (Priority Medium).** The stub, the full body, and the companions, plus the template carry of the line-220 sentence (Q3).
- **M5 — Local parity and policy copies (Priority Medium).** Build gate, local sync of changed artifacts, the surgical policy-copy edits (Q5 default), and the local-only lane protocol sentence.
- **M6 — Verification and ledger reconciliation (Priority Low, mechanical).** Acceptance criteria AC-LNC-001 through AC-LNC-016, the runtime observation AC-LNC-003, and the reconciliation of the remaining `/clear` references against the ledger.

Execution order (dependency-driven, distinct from the priority order above): M1 and M2 and M3 and M4 edit template sources first; then the `make build` gate; then M5 local sync; then M6 verification.

## §E Technical Approach by Surface

- **Transition doctrine (template-first).** Edit the factory rule family under `internal/template/templates/.claude/rules/moai/workflow/`. Replace the section "The `/clear` handoff between cards" with the lane transition passage. Replace the isolation clause that forbids moving between card worktrees and names relaunch, with the exit-then-enter passage. Keep the leader's own session-clear sentences. Keep the new-card-worktree clause that already requires `ExitWorktree` first. Put the new [HARD] transition clause inside the role-core region of the full body and of the stub, so the SessionStart injection carries it to lanes.
- **Completion output (Go).** In `internal/cli/factory_card.go`, change the one printer so that clear-each and clear-when-full print the in-session transition line instead of a clear request, and relaunch keeps its end-session line. Flag parsing, the environment key, and the record reads stay unchanged (Q2 default). The three callers are unchanged.
- **Completion merge observation (Go).** The merge-observation path calls the same printer, so no separate change is expected there. Confirm with the printer test that the merge-observation path prints no clear instruction.
- **SessionStart lane join (Go).** In `internal/hook/session_start_factory_i18n.go`, change the English next-card text so it names the in-session transition instead of "follow the clear policy it prints", and change the recheck texts in English, Korean, Japanese, and Chinese so that card takeover is the trigger and the post-clear case is a secondary trigger. Update the comment in `internal/hook/session_start_factory.go` that describes the clear-each loop. The rule still fires on source startup and source clear.
- **Todo lane cycle (Go).** `internal/cli/todo_auto_lane.go` carries no `/clear` text. Add an assertion to the lane cycle tests that the completion path output and the exit guidance contain no `/clear`. Change code only if that assertion fails.
- **Operator-session serial cycle (Go, unchanged).** `internal/cli/todo_auto.go` keeps `writeAutoClearGuidance` and its caller (Q1 default). The guidance test `TestTodoAutoClearGuidancePerCard` and the golden file stay unchanged.
- **Leader and manager text (template-first).** In `manager-lead.md`, change the frontmatter description phrase "`/clear` between cards" and the Role-B return template line that requests an operator `/clear {session-name}`. The return template keeps the action field, which may name only operator actions other than clearing a lane. In `manager-todo.md`, add a scope sentence to the per-card guidance: the guidance applies to the operator session only, and lanes follow the lane transition rule. In `gtd.md` (`--auto` section), add the same scope sentence to the per-card guidance phrase.
- **Codex TOML (generated).** `make agents-emit` regenerates `internal/template/templates/.codex/agents/moai/manager-lead.toml` and `manager-todo.toml` from the neutral `.md` layer. `make agents-emit-check` is the read-only gate. Do not hand-edit the TOML files.
- **Policy copies (surgical, Q5 default).** In `.moai/policies/workflow/factory-dispatch.md`, edit the migrated description line, the section "The `/clear` handoff between cards" (its lane passages), and the isolation clause. In `factory-dispatch-detail.md` and `factory-dispatch-cards.md`, edit the same passages as their rule counterparts. `factory-dispatch-mechanics.md` in the policy tree carries no `/clear` line and needs no edit. The policy copies are edited to the same transition text as the rule files; no other passage is touched.
- **Local-only lane protocol.** `.claude/rules/local/gitflow-lane-protocol.md` has no template mirror. Its § 1 sentence that runs `/clear` once after a move is aligned to the in-session sequence.
- **Local parity.** After the template edits and `make build`, copy each changed template artifact to its local counterpart so the two are byte-identical. Exceptions: the root dogfood codex TOML (Q4 default: untouched), the policy copies (surgical edit, not a copy), and the kanban policy-only copies (Q8 default: untouched).
- **Tests to update (targeted).** `factory_m6_test.go`: clear-each and absent policy print no `/clear`; clear-when-full prints nothing at or above the threshold; the complete output carries zero `/clear`; relaunch keeps its end-session line and names no clear. `role_core_elements_test.go` (`TestInjectedCoreCarriesObligationElements`): replace the `never before` row with the new transition phrases asserted in the injected core. `session_start_factory_rule_test.go`: confirm the rule still fires on source clear and startup.

## §F Location Ledger

Each row names a location, its current text in short form, its disposition, its kind, and the requirement it serves. Template and local rows in the rule family share line numbers; the local copy differs from the template only at line 220 (Q3). Kinds: EDIT (change the passage), ADD (new passage), KEEP (no change, with the kind named), REGEN (regenerated from the source), PENDING (no change until the named decision), LOCAL (local only, no template).

| ID | Path | Location | Current text (short) | Disposition | Kind | REQ |
|----|------|----------|----------------------|-------------|------|-----|
| L01 | `.claude/rules/moai/workflow/factory-dispatch.md` (template mirrored) | L31 | "After a `/clear`, the leader presents the queued cards" | KEEP | leader session's own clear | 011, 016 |
| L02 | same | L40 | "after `/clear` present the queue" | KEEP | leader session's own clear | 011, 016 |
| L03 | same | L152 | heading "The `/clear` handoff between cards" | EDIT | lane transition | 013 |
| L04 | same | L154 | "No card context crosses ... asks the operator to `/clear` that lane session" | EDIT | lane transition | 001, 006 |
| L05 | same | L157 | role-core "`/clear` exactly once per transition ... never before" | EDIT (role-core region) | lane transition | 013, 015 |
| L06 | same | L160 | "Where the next card reuses a just-cleared lane, the leader re-sends the full pointer" | EDIT | lane transition | 005 |
| L07 | same | L162 | "The leader's own session is cleared the same way" | KEEP | leader session's own clear | 011, 016 |
| L08 | same | L166 | "moving a running card session into another card worktree is prohibited; end the session and launch inside the new tree" | EDIT (role-core region) | lane transition | 013 |
| L09 | same | L171 | "moving into another card worktree is prohibited (exit-and-relaunch ...); an unavoidable move takes `/clear` once" | EDIT | lane transition | 013 |
| L10 | same | L184 | "A new card gets a new worktree ... `ExitWorktree` the old tree first" | KEEP | no clear; consistent with the transition | 013 |
| L11 | same | L220 (local only) | appended sentence on disposal of the card worktree after the remote landing | EDIT (carry into the template, then local equals template) | POLICY-COVERED (Q3) | 014 |
| L12 | `.claude/rules/moai/workflow/factory-dispatch-core.md` (stub) | L52 | heading "The `/clear` handoff between cards" | EDIT | lane transition | 013 |
| L13 | same | L54 | "Where the next card reuses a just-cleared lane" | EDIT | lane transition | 005 |
| L14 | same | L56 | "The leader's own session is cleared the same way" | KEEP | leader session's own clear | 011, 016 |
| L15 | same | L60 | "exit-and-relaunch satisfies the new-card rule; ... takes `/clear` once, after" | EDIT | lane transition | 013 |
| L16 | same | new passage under the section | the lane transition rule (role-core region) | ADD | lane transition | 013 |
| L17 | `.claude/rules/moai/workflow/factory-dispatch-detail.md` | L23 | glossary: leader "asks the operator to `/clear` a lane between cards" | EDIT | lane transition | 011 |
| L18 | same | L73 | "the leader's reports to the operator (progress notes, `/clear` requests)" | EDIT | lane transition | 011 |
| L19 | same | L79 | manager-lead dialogue "card selection, `/clear` prompts, blocker surfacing" | EDIT | lane transition | 011 |
| L20 | same | L137 | "The `/clear` boundary is between cards" | EDIT | lane transition | 013 |
| L21 | same | L147 | "[report] ... then the `/clear` the stub requires" | EDIT | lane transition | 007, 013 |
| L22 | same | L166 | cross-reference "the `/clear` handoff message structure" | EDIT | lane transition | 013 |
| L23 | `.claude/rules/moai/workflow/factory-dispatch-cards.md` | L2 | companion description "the /clear handoff message structure" | EDIT | lane transition | 013 |
| L24 | same | L8 | companion description "the `/clear` message structure" | EDIT | lane transition | 013 |
| L25 | same | L36 and L41 | section "The `/clear` handoff between cards — message structure"; item "Which session to `/clear`" | EDIT (retire the operator message structure; replace with the transition statement) | lane transition | 011, 013 |
| L26 | `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` | L202 | migrated description "when the operator is asked to `/clear`" | EDIT | lane transition | 011 |
| L27 | same | L247 | "the question channel the leader uses for card selection and `/clear` prompts" | EDIT | lane transition | 011 |
| L28 | `.claude/rules/moai/workflow/auto-semantics.md` | L130 | "`CronList` at card intake and after every `/clear`" | EDIT (card takeover named, including the in-session transition) | lane stall-recheck | 010 |
| L29 | `.claude/rules/local/gitflow-lane-protocol.md` | § 1, the sentence that runs `/clear` once after a move | "if a move is unavoidable, run `/clear` exactly once after the move" | EDIT | LOCAL | 013 |
| L30 | `.claude/agents/moai/manager-lead.md` | L6 | frontmatter description "`/clear` between cards" | EDIT | leader text | 011 |
| L31 | same | L107 | "re-derives it from SPEC status after a `/clear`" | KEEP | leader session's own clear | 011, 016 |
| L32 | same | L115 | return template "operator action requested: /clear {session-name} (or: none)" | EDIT (the action field names only operator actions other than clearing a lane) | leader text | 011 |
| L33 | same | L171 | "/compact ... fall back to `/clear` + paste-ready resume" | KEEP | leader / manager-lead own context | 016 |
| L34 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | L7, L102, L110, L166 | generated from L30 to L33 | REGEN (`make agents-emit`) | generated | 014 |
| L35 | `.claude/agents/moai/manager-todo.md` (template mirrored) | L55 | "After each completed card, emit the /clear guidance naming the completed card" | EDIT (scope sentence: operator session only) | operator-session serial cycle | 012 |
| L36 | `internal/template/templates/.codex/agents/moai/manager-todo.toml` | L54 | generated from L35 | REGEN (`make agents-emit`) | generated | 012, 014 |
| L37 | `.claude/skills/moai/workflows/gtd.md` (template mirrored) | ~L294 | "the leader's own post-`/clear` opening move" | KEEP | leader session's own clear | 011, 016 |
| L38 | same | ~L336 | "emit the /clear guidance for the completed card, and only then accept the next" | EDIT (scope sentence: operator session; lanes follow the lane transition rule) | operator-session serial cycle | 012 |
| L39 | `internal/cli/factory_card.go` | L2433 (comment) and L2437-L2447 (printer text) | "ask the operator to /clear, then take the next card on the fresh session" | EDIT | lane transition | 007, 009 |
| L40 | same | L2167, L2257 (callers) | `factoryPrintClearPolicyLine` calls in `factoryCompleteCard` | KEEP (callers unchanged) | not text | 007 |
| L41 | `internal/cli/factory_card_pr.go` | L515 (caller) | `factoryPrintClearPolicyLine` call in `factoryObservePRMerge` | KEEP (caller unchanged) | not text | 007 |
| L42 | `internal/cli/factory_relaunch.go` | L26 (comment) | "is substituted for the clear-policy loop" | EDIT (comment) | lane transition | 009 |
| L43 | `internal/config/envkeys.go` | constant comments around the lines 394 to 434 (clear-each, clear-when-full, relaunch) | comments describing a clear request | EDIT (comments only; keys and values unchanged) | lane transition | 009 |
| L44 | `internal/cli/factory_m6_test.go` | functions `TestSD_AC020_ClearPolicies` (from L83) and `TestSD_M6_ClearPolicyCarrierConstantOnly` (L291); assertions at the clear-each, absent, clear-when-full at threshold, complete-output, and relaunch cases | assertions for one `/clear` request | EDIT (assertions flip to zero clear instructions; relaunch case keeps its no-clear assertion) | lane transition test | 007, 009, 015 |
| L45 | `internal/cli/todo_auto_lane.go` | no `/clear` text | none | KEEP; a new assertion in the lane cycle tests | not text | 008 |
| L46 | `internal/cli/todo_auto_lane_test.go` | lane cycle tests (`TestAutoLaneCycle*`) | none for clear text | EDIT (add the no-`/clear` assertion to completion output and exit guidance) | lane cycle test | 008 |
| L47 | `internal/cli/todo_auto.go` | L23 (comment), L428 (caller), L487-L493 (`writeAutoClearGuidance`) | operator-session per-card guidance | KEEP | operator-session serial cycle | 012, 016 |
| L48 | `internal/cli/todo_auto_test.go` | `TestTodoAutoClearGuidancePerCard` and its literal assertion | operator-session guidance assertion | KEEP | operator-session serial cycle | 012, 016 |
| L49 | `internal/cli/testdata/qas_baseline_todo_auto.golden.txt` | the two guidance blocks | operator-session guidance | KEEP | operator-session serial cycle | 012, 016 |
| L50 | `internal/hook/session_start_factory_i18n.go` | English next-card text (around L135-L150): "then follow the clear policy it prints" | lane join next-card text | EDIT | lane transition | 010 |
| L51 | same | English recheck text (L156): "after every clear" | stall-recheck text | EDIT (card takeover named as trigger) | lane stall-recheck | 010 |
| L52 | same | Korean recheck text (L216), Japanese (L276), Chinese (L335) | stall-recheck text, each with `/clear` after takeover | EDIT (card takeover named as trigger; four locales carry the same meaning) | lane stall-recheck | 010 |
| L53 | `internal/hook/session_start_factory.go` | comment near L108-L110 ("keeps the clear-each loop going") | comment | EDIT (comment) | lane transition | 010 |
| L54 | `internal/hook/role_core_elements_test.go` | row "ALB-0294 (/clear after, never before)" at L53 | asserts `never before` in the rule body | EDIT (row replaced by the new transition phrases; the injected-core check asserts them in the role-core region) | lane transition test | 015 |
| L55 | `internal/hook/session_start_factory_rule_test.go` | L13-L14 (comments), L59-L75 (rule matrix), L122 (source clear) | rule injection matrix | KEEP (review; assertions still hold) | not text | 015 |
| L56 | `.moai/policies/workflow/factory-dispatch.md` (policy copy, Q5 default) | L3 (migrated description) | "when the operator is asked to `/clear`" | EDIT (surgical) | lane transition | 014 |
| L57 | same | L31 | leader "After a `/clear`" | KEEP | leader session's own clear | 011, 016 |
| L58 | same | L137-L145 (section "The `/clear` handoff between cards") | the section's lane passages (L139, L141) | EDIT (surgical: the same transition passage as the rule file) | lane transition | 013, 014 |
| L59 | same | L145 | "The leader's own session is cleared the same way" | KEEP | leader session's own clear | 011, 016 |
| L60 | same | L157 | isolation clause "A card session starts inside its worktree and stays there" | EDIT (surgical: same passage as the rule file) | lane transition | 013, 014 |
| L61 | same | L202 | "question channel ... and `/clear` prompts" | EDIT (surgical) | lane transition | 011, 014 |
| L62 | `.moai/policies/workflow/factory-dispatch-detail.md` | L23, L73, L79, L137, L147, L166 | same passages as L17 to L22 | EDIT (surgical) | lane transition | 011, 013, 014 |
| L63 | `.moai/policies/workflow/factory-dispatch-cards.md` | L2, L8, L36, L41 | same passages as L23 to L25 | EDIT (surgical) | lane transition | 013, 014 |
| L64 | `.moai/policies/workflow/factory-dispatch-mechanics.md` | no `/clear` line | none | KEEP (no edit; parity checked by AC-LNC-013) | not text | 014 |
| L65 | `.moai/policies/workflow/kanban-dispatch.md` (8 lines) and `kanban-dispatch-detail.md` (7 lines) | policy-only copies with no rule-tree source | predecessor kanban lane text | PENDING (Q8 default: untouched) | PENDING | 016 |
| L66 | `internal/template/templates/.claude/rules/moai/workflow/*` and `.claude/rules/moai/workflow/*` (template mirrors of L01 to L29) | template and local copies | as above | EDIT (template first), then local parity | template-first | 014 |
| L67 | `.claude/rules/moai/workflow/worktree-integration.md` | L126, L127, L156, L157 | generic re-entry and new-session launch | KEEP | generic session-lifecycle note | 016 |
| L68 | `.claude/rules/moai/workflow/spec-workflow.md` (3 lines; template mirrored), `context-window-management.md` (6), `context-window-management-detail.md` (15), `context-clear-policy.md` (1), `session-handoff.md` (2), `session-handoff-format.md` (6), `session-handoff-examples.md` (8), `runtime-recovery-doctrine.md` (4), `cache-aware-execution.md` (3), `cache-aware-execution-reference.md` (3), `goal-directive-detail.md` (4), `native-invocation-model.md` (3), `settings-management.md` (3), `rule-authoring.md` (2), `askuser-protocol-reference.md` (3), `agent-common-protocol-reference.md` (2), `zone-registry.md` (1), `moai-memory.md` (1), `contract-autonomy.md` (1) | SPEC-phase, context, handoff, and goal mechanisms | outside the lane card transition | KEEP | generic session-lifecycle or context-threshold note | 016 |
| L69 | `.claude/skills/moai/workflows/run.md`, `goal.md`, `sync/delivery.md` (L393, L398 "Start New Session (/clear)") | next-step menus and goal state | generic | KEEP | generic session-lifecycle note | 016 |
| L70 | `.claude/output-styles/moai/moai.md` (L677, L730); `.moai/docs/*.md` (four design documents); `.moai/config/sections/workflow.yaml` (L56: Tier L folding fallback); `.moai/config/sections/handoff.yaml` (and template mirror); `.claude/hooks/tests/test-context-clear-policy.sh` | session handoff, design notes, Tier L folding, hook test | generic | KEEP | generic session-lifecycle or context-threshold note | 016 |
| L71 | `internal/cli/{chain,handoff,codex_review_scope,goal,integration,integration_codemaps_card,factory_launch_helpers}.go`; `internal/hook/{session_start,chain_banner,handoff_inject,user_prompt_submit,factory_rebind,role_rules}.go`; `internal/hook/handoff/pending.go`; `internal/factory/integration_lock.go`; `internal/statusline/*`; `internal/runtime/*`; `internal/settings/llmoverrides.go`; `internal/config/types.go`; `internal/tokenusage/tokenusage.go`; `internal/chain/*`; `internal/goal/dashboard.go`; `internal/template/{agent_overrides_consume,profile_matrix}.go`; `internal/template/templates/.claude/skills/moai-foundation-core/**`; `internal/template/templates/.claude/skills/moai-workflow-spec/references/reference.md` | comments and text about session identity, goal re-arm, handoff, statusline, cache | generic | KEEP | generic session-lifecycle note | 016 |
| L72 | `.claude/agents/harness/hns-release-update-specialist.md` L147 | keyword list | unrelated | KEEP | not a lane instruction | 016 |

Ledger totals used by AC-LNC-016: the scoped surfaces are the files in L01 to L64 and L66 (template mirrors included). Their remaining `/clear` lines must equal the KEEP and PENDING rows of the ledger for each file.

## §G Risks

- **R1 — In-session entry refused (Q7).** The sequence may be refused on the deployed runtime. Mitigation: REQ-LNC-004 sends the lane to the blocker path with the verbatim refusal. No clear substitute is allowed.
- **R2 — Context growth without a lane prompt.** The card's acceptance removes the threshold request from lane output. Mitigation: recorded residual risk; the orchestrator threshold rule still applies; the operator may clear at any time.
- **R3 — Stale policy copies (Q5).** The surgical default leaves pre-existing drift. Mitigation: AC-LNC-013 checks only the transition phrases; the operator may choose a wholesale regeneration.
- **R4 — Template change from the local-only sentence (Q3).** Carrying the sentence into the template changes the distributed text by one sentence. Mitigation: the policy direction is stated in the decision row.
- **R5 — Codex lanes keep the old surfaces (Q6).** Codex lanes end their session and do not use the Claude tools. Mitigation: REQ-LNC-013 scopes the rule to Claude Code lanes; Q6 records the alternative.
- **R6 — Orphan kanban copies (Q8).** Predecessor copies still carry lane clear text. Mitigation: recorded as PENDING; the operator decides.
- **R7 — Test churn.** The clear-policy and role-core tests change expectation by design. Mitigation: the acceptance criteria name each changed test and its new expectation.

## §H Procedure Notes for the Run Phase

- Template first: edit `internal/template/templates/` for every changed artifact, then run `make build`, then sync the local copies, then run the targeted tests.
- Generated TOML: run `make agents-emit` after the agent template edits; run `make agents-emit-check` to confirm.
- Verification evidence: write the command and the observed output into the verdict file under `.moai/reports/t1793/verdict.md` for each acceptance criterion. Scratch output stays in the session scratchpad and is named in Residual-risk when not carried.
- Do not fill any decision verdict line. Do not edit `.moai/spec-lint-baseline.json`.

## §I Cross-References

- `.claude/rules/moai/workflow/factory-dispatch.md`, `factory-dispatch-core.md`, `factory-dispatch-detail.md`, `factory-dispatch-cards.md`, `factory-dispatch-mechanics.md` — the factory doctrine family this SPEC edits.
- `.claude/rules/moai/workflow/auto-semantics.md` § 13 and § 6 — the lane judgment ladder that the card takeover trigger uses.
- `.claude/rules/moai/workflow/worktree-integration.md` — worktree tiers and disposal; not edited.
- `.claude/rules/moai/development/coding-standards.md` § Duplicate Prevention — the policy that governs Q3.
- `decision-index.md` — the open decision rows Q1 through Q8.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — the frontmatter and phase rules this SPEC follows.
- Predecessor SPECs carried in `related_specs`: SPEC-FACTORY-SELF-DISPATCH-001 (the lane rule and clear-policy requirements), SPEC-FACTORY-LANE-AUTONOMY-001 (lane autonomy), SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 (lane worktree handoff), SPEC-TODO-CLASSIFY-DISPATCH-001 (lane dispatch selection), SPEC-WORKTREE-EXIT-RETURN-001 (worktree exit return).
