# Gateway SPEC session progress — 2026-09-11 10:55 KST

SPEC-MOAI-GATEWAY-001 · session 25b43a41 · worktree WT-unified-gateway @ ed71054d3

## State
- SPEC 0.6.0: 24 live REQ / 24 AC (tombstones REQ-MG-007, REQ-MG-020, AC-MG-002), 286,400 B, lint clean.
- Plan audits: iter1 FAIL 0.70, iter2 FAIL 0.80, iter3 FAIL 0.84 (3-iteration cap, operator exception), iter4 FAIL 0.80 (STOP), iter5 FAIL 0.83 (no STOP; last authorized iteration consumed).
- Split: picker surface → SPEC-MOAI-GATEWAY-PICKER-001 (proposed); tmux teammate surface → SPEC-MOAI-GATEWAY-TEAMMATE-001 (proposed). Decision 9 kept in core with M1 raw-capture gate. Gateway launches: in-process teammates only, no tmux env writes.
- origin/develop +65 commits past ed71054d3; cited files changed: internal/hook/session_end.go (citations shift +1/+5), .github/workflows/ci.yml (+2). No semantic impact.

## iter5 blockers (report: .moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md)
| ID | Defect | Orchestrator check | Operator decision |
|---|---|---|---|
| G5-B1 critical | design.md:618-619 claims stale tmux GLM keys never reach the Claude child; nothing strips them; gateway `moai cc` could route opus/sonnet/haiku to Z.AI (REQ-MG-022) | launcher.go:837 builds child env from os.Environ(); buildEnvForLaunch (:1180-1200) only replaces the effort key — confirmed. tmux-to-pane inheritance is auditor inference, not run | Not needed for fix (i): strip GLM cleanup key set + Z_AI_API_KEY from child env, add pre-seeded env variant to AC-MG-018 (a). Fix (ii) restore tmux clear = amend decision 12 |
| G5-B2 | Core-only release behaviour unstated: persisted `/model` default makes next `moai cc` start on GPT while provider recorded as claude; release coupled to GPT-AUTH but not PICKER | Consistent with probe 1 (persisted default) and decision 10 moving to sibling | Needed: couple core release to PICKER sibling, or accept written residual |
| G5-B3 | REQ-MG-022 says in-process limit "blocks" tmux-pane bypass, but teammateMode is in a shared file rewritable to tmux (design.md:611-615) | Same as writer residual risk 3 | Needed only if per-session `--teammate-mode in-process` flag chosen; else reword |
Also: stale pointer research.md:917; advisories (outside-tmux teammateMode judgment, recognizer variants stream:null / max_tokens:0 / single system message, fixture path). Audit was Claude-only. Auditor recommendation: explicit override for one scoped fix + re-audit limited to B1–B3; PASS-with-debt not recommended for B1; further split not recommended.

## Timeline (KST)
| Time | Event |
|---|---|
| 09-10 14:18 | Session start |
| 14:21–14:27 | FO-PLAN-1 research fan-out workflow |
| 14:35–15:02 | Author → iter1 FAIL 0.70 |
| 15:49–16:09 | iter2 FAIL 0.80 |
| 16:22–18:17 | 6 mid-pass corrections not delivered → TaskStop + re-issue |
| 18:23–18:39 | iter3 FAIL 0.84; operator one-time exception |
| 20:35–20:37 | 0.4.0 writer died on usage limit (wait to 22:00) |
| 22:06–22:34 | 0.4.0 done |
| 22:31–22:57 | Probes 1–2: client-level same-session /model switching works |
| 23:02–23:35 | Budget blocker (25/25) → fold → 0.5.0 written, agent died on limit (wait to 03:03) |
| 09-11 03:06–03:32 | iter4 FAIL 0.80 STOP |
| 03:33–03:50 | Decisions: split, D9 core, in-process teammates; ff develop (+261), stale index.lock removed |
| 03:51–04:21 | 0.6.0 split written |
| 04:22–04:36 | iter5 died on limit, no report (wait to 10:35) |
| 10:35–10:47 | iter5 re-run: FAIL 0.83 |

Time split: ~9.7h work (incl. user-answer waits) vs ~10.9h usage-limit waits (53%).
Measured agent cost (post-compaction): split writer 611,177 tokens / 29.3 min; iter4 audit 440,923 / 25.9 min; iter5 re-run 368,698 / 12.1 min; budget-blocked writer 318,731 / 2.9 min.
User question rounds: 12.

## Remaining in this session
1. Operator decision: scoped fix + re-audit (B1–B3) / PASS-with-debt / stop and hand off.
2. If fix: decisions for B2 and B3; ff worktree to origin/develop + re-anchor citations; SPEC 0.7.0 (B1–B3, research.md:917); iter6 audit limited to B1–B3; PASS → Implementation Kickoff Approval question and stop.
3. Propose backlog cards for PICKER-001 and TEAMMATE-001 (add only on approval).
4. Memory lessons: raw-body capture in probes; stale index.lock handling; propose scope reduction at audit cap; brief must respect REQ/AC budget; check what a moved decision was supporting in core before splitting.
5. Handoff save + render paste block.
Out of scope without instruction: /moai run, push, PR, merge, worktree removal.

## Why slow
1. Usage-limit waits ~10.9h (3 agent terminations).
2. 5 audit iterations on a Tier L SPEC at the 25/25 ceiling; scope reduction after iter4 instead of at the iter3 cap (spec-workflow.md:160); split introduced G5-B2.
3. Orchestrator errors causing rework: undelivered mid-pass messages; wrong premises in briefs; brief conflicting with budget; probe logged `stream` as bool with suppression flags on (G4-B2); stale base (261 commits) found late; split brief missed decision 10's core dependency (G5-B2).
4. 12 serial decision rounds (2 caused by orchestrator errors).

## Fan-out check against guidelines
| Rule | Session | Verdict |
|---|---|---|
| plan.md FO-PLAN-1 research fan-out | used 14:21 | complied |
| agent-common-protocol: no concurrent write agents; one writer during audit | serial writer/auditor | complied; not parallelizable |
| spec-workflow.md:160 max 3 audits + scope reduction on regression | exception at cap; reduction after iter4 | applied late |
| orchestration-mode-selection §B: edit-heavy serial, research fanout | serial edits | complied |
| Workflow tool only on user/skill opt-in | only FO-PLAN-1 | complied |

Missed parallelism: probes 1–2 concurrently or one combined probe with raw capture; read-only citation/divergence checks while writers ran; lightweight Explore for citation checks. Fan-out does not fix the dominant cost: concurrent agents burn the usage limit faster, and writer/audit are sequential gates. Highest-leverage fixes: split scope at the audit cap (or design smaller SPECs up front); design probes with sufficient evidence.

## Gaps
- Pre-compaction agent token usage not recovered.
- User-answer wait time not separated.
- G5-B1 tmux-to-new-pane env inheritance not executed.
- iter5 audit Claude-only; earlier HISTORY byte-identity unverifiable (untracked SPEC dir).
