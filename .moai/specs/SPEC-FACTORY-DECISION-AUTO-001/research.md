# research.md — SPEC-FACTORY-DECISION-AUTO-001

All probes below were run in this plan session against tree `d7112d005` (branch
`WT-decision-automation`, local develop tip at authoring). The leader's original investigation ran
at `42d8474de`; every site it named was re-located here. Line numbers are locating aids at
`d7112d005` only.

## 1. Observed waits (leader investigation, card t1481 evidence)

Reported by the leader, not re-measured here (they are session observations, not tree state):
ceiling hits on t1356 (iter3/iter4), t1404 (iter2, score regression), t1409 (×3), t1458 (Tier M
cap); PASS-WITH-DEBT Kickoff waits on t1377, t1409 F3, t1438; audit-debt disposition on t1409 D1
and t1438 OD-1..OD-8 (all FOUNDER, verdict blank); merge-window nomination on every lane (card
t1479); stale lane MCP (rc.23 lacked `codex_review`); repeated
`factory messaging degraded: context deadline exceeded`. Gap: these are leader-reported, not
re-observed in this session.

Ceiling-ruling evidence (leader-reported 2026-10-03, observed in the mission session): the leader
applied the same "one delta round" ruling four times (t1399, t1454, t1469, t1458) and ruled "hold"
twice (t1356, and one earlier card). REQ-FDA-013/014 codify this rule as policy.

## 2. Probe ledger

| # | Command (run from the worktree root) | Observed output (verbatim, trimmed to the deciding lines) |
|---|---|---|
| P1 | `grep -rn 'Use: *"decision' internal/cli/` | no output, `exit=1` |
| P1+ | positive control: `grep -rn 'Use: *"contract\|Use: *"decide' internal/cli/*.go` | `internal/cli/contract_decide.go:28: Use: "decide <card> ..."`, `internal/cli/contract.go:449: Use: "contract"`, `internal/cli/factory_card.go:1972: Use: "decide <card>..."` |
| P2 | `grep -rln "decision-board\|decisionboard\|decision_board" internal/` | no output |
| P3 | `grep -n "PASS-WITH-DEBT" internal/contract/rules.go internal/contract/kickoff/decide.go internal/homestate/card_transition.go` | `rules.go:25: passingVerdicts = []string{"PASS", "PASS-WITH-DEBT"}`; `card_transition.go:470: ... v.Verdict != "PASS" && v.Verdict != "PASS-WITH-DEBT"`; `decide.go:393: if verdict != "PASS" && verdict != "PASS-WITH-DEBT"` |
| P4 | `grep -n "PASS-WITH-DEBT" .claude/rules/moai/workflow/auto-semantics.md` | `:244: - Blocked states are: PASS-WITH-DEBT, BYPASSED, FAIL, INCONCLUSIVE, ...` |
| P5 | `grep -n "PASS-WITH-DEBT" .claude/agents/moai/plan-auditor.md` | `:203: verdict: <PASS\|PASS-WITH-DEBT\|FAIL>`; `:243: AUDIT-VERDICT: <PASS\|PASS-WITH-DEBT\|FAIL> ...` — no emission rule |
| P6 | `grep -n "DeciderHuman" internal/homestate/card_record.go internal/homestate/card_transition.go` | `card_record.go:46: const DeciderHuman = "human"`; `card_transition.go:480-481: if req.Decider != DeciderHuman { ... F1 accepts only decider ...` |
| P7 | `sed -n 100,115p internal/homestate/card_transition.go` | `{"T8", CardKickoff, CardAssigned, guardKickoffDecision},` |
| P8 | `sed -n 1975,1990p internal/cli/factory_card.go` | `if factoryLaneRefusal() { return factoryDecideLaneRefusal() }` then `if decider != homestate.DeciderHuman { ... F1 records only ...` |
| P9 | `grep -n "ceiling\|max_iter" .moai/config/sections/harness.yaml` | `:75: plan_audit_tier_ceilings:` `:76: S: 1` (M: 2, L: 3 follow); `:102`, `:115: max_iterations: 3` |
| P10 | `grep -n "Max 3 iterations cap\|STOP escalation" .claude/agents/moai/plan-auditor.md` | `:693` STOP → "present the user with three options via the orchestrator's user-question channel"; `:703` "After iter3 ... escalates to the user" |
| P11 | `sed -n 158p .claude/rules/moai/workflow/spec-workflow.md` | "Maximum 3 plan-auditor iterations per SPEC plan-phase; after iter3, escalate via PASS-with-debt OR scope-reduction OR explicit user override." |
| P12 | `sed -n 108,116p .claude/agents/moai/manager-spec.md` | "The **authority register** is committed artifacts only: `.moai/project/product.md`, prior completed SPECs' HISTORY and `## Amendments` rows, `.moai/config/sections/*.yaml` operator settings, and the project constitution." |
| P13 | `grep -n "decision_gate" .moai/config/sections/interview.yaml` | `6: decision_gate: on` |
| P14 | `sed -n 145,153p internal/hook/factory_messages.go` | `:145 s, err := factorymsg.Open(root, runID)` ... `s.Peer(ctx, input.SessionID)` ... `:153 return "", ""` (already-bound early return after open + peer query) |
| P15 | `sed -n 205,235p internal/hook/factory_messages.go` | `:210 OpenExistingWithDeadline(root, runID, factoryHookInspectionDeadline)` (200ms); comment: "Both call sites still discard the state ... nothing on this path emits one yet." |
| P16 | `sed -n 150,170p internal/hook/user_prompt_submit.go` | bind under `factoryBindBudget` (`:62` = 2s); `:168 if factoryCtx, _, _ := factoryHookBatchForRun(...)` (state discarded) |
| P17 | `sed -n 80,95p internal/hook/stop.go` | `factoryHookBatch(ctx, input, EventStop)` — messages claimed at turn end |
| P18 | `grep -n "7,27,47" .claude/rules/moai/workflow/auto-semantics.md` | `:108: CronCreate with cron: "7,27,47 * * * *" (off-minute, every 20` |
| P19 | `sed -n 198,206p .claude/rules/moai/workflow/kanban-dispatch-detail.md` | "**Fallback.** Where the running MCP server predates the tools ... the codex leg runs as `moai verify codex-review --project-root <tree>`" |
| P20 | `sed -n 190,196p .claude/rules/moai/workflow/auto-semantics.md` | `factory decide: kickoff approve/reject \| AUTONOMOUS — the independent audit cross is the entry evidence (§9.1)` |

## 3. Findings

### F1 — the board is doctrine without a carrier (P1, P2)

Ladder step ② (§6), the §7 transition row, the §11 SSOT, and the watchdog skill's step 2 all name
the board; nothing writes or reads it. Step ② always falls through. Rulings live in chat and are
copied by hand.

### F2 — PASS-WITH-DEBT: three code sites admit, one doctrine line blocks, the auditor never defines (P3-P5, P11)

Code admits it unconditionally on the label; §9.2 lists it as blocked; §9.1 demands plain PASS;
spec-workflow names "PASS-with-debt" as an iter3 escalation outcome; the auditor's verdict schema
lists the token with no emission rule. Auditors therefore disagree on when to emit it, and lanes
cannot predict whether Kickoff opens.

### F3 — ceiling routing contradicts the leader-decision policy (P9-P11)

The auditor and spec-workflow route ceiling hits and score-regression STOPs to the user channel;
the local kickoff-autonomy policy makes them a leader decision. Caps are stated three ways:
`plan_audit_tier_ceilings` (S=1, M=2, L=3), the auditor's "Max 3" text, and `max_iterations: 3`.
The leader's rulings were formulaic: one delta round when the remaining fixes were mechanical, then
hold or split.

### F4 — factory Kickoff is human-only in code, autonomous in doctrine (P6-P8, P20)

`guardKickoffDecision` and `factory decide` accept only `human`; a lane is refused outright; T8
sends an approved card to `assigned`, forcing a re-lease of a card the lane already holds.

### F5 — FOUNDER is the only reachable label for mission-session decisions (P12, P13)

With `decision_gate: on`, rows whose authority is a mission contract or a leader ruling cannot be
routed DECIDED/POLICY-COVERED, so every row is FOUNDER and blank verdicts block.

### F6 — wake latency and messaging degradation (P14-P18)

Messages are claimed at turn boundaries only; an idle lane wakes on the 20-minute cron. The
prompt-submit bind opens the DB and queries the peer before discovering it is already bound,
under a 2s budget; under load (comment at `factory_messages.go:29-35`) this misses. The inbox
claim's degraded state is discarded at both call sites.

### F7 — stale MCP is detectable but not checked (P19)

The CLI fallback exists; intake never decides to use it.

## 4. Gaps

- The leader-reported wait instances (§1) were not re-observed.
- Bind-miss frequency under load was not measured here; the cache's benefit is argued from code
  order, not from a measured hit rate (EVIDENCE-NEEDED, decision-index Q5).
- The cache cost of a 5-minute recheck relative to the prompt-cache window was not measured.
