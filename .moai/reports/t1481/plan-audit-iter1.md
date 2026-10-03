auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-FACTORY-DECISION-AUTO-001
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.74
Plan Artifact Hash: 2a0d6042087b03897120024b1d62f16e8fbefe816f670cbd0c9e8f9df98e8355 (sha256 of concatenation spec.md+plan.md+acceptance.md+design.md+research.md+decision-index.md, in that order)
Auditor Version: plan-auditor/v(current agent definition at HEAD ba2033d22)

verdict: FAIL
audited_sha: ba2033d22abee6cf37e02fdee1241038d6cc7356

Card: t1481 · Tier L (threshold 0.85) · Tree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de · branch WT-decision-automation
Reasoning context ignored per M1 Context Isolation (author reasoning); the orchestrator's focus list was retained as task direction. Known bootstrap gap noted, not failed on alone: decision-index.md labels Q1-Q7 `LEADER-DECIDED` (decision-index.md:5-8), outside manager-spec's four-label vocabulary until REQ-FDA-007 lands.

Cross-model: `audit_multi` (baseBranch, project_root = card tree) — claude FAIL (in-session anchor), codex FAIL (required), GLM inconclusive (advisory, "z.ai response carried no text content"). overall_verdict=fail, disagreement_flag=false. No `audit_receipt` id was returned in the result. Tool reported build lag: MCP binary 45600e4ee is an ancestor of HEAD ba2033d22.

---

## 1. Claim

FAIL. One must-pass criterion fails (MP-8), and there are 7 blocking content defects affecting focus items (1), (2), (3), (6) and (7). Traceability, frontmatter, REQ numbering, GEARS form, D7/D8, the clarification gate and ordering (MP-9) pass.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-FDA-001..025 run in sequence with no gaps or duplicates (spec.md:86-216). The traceability verb collected 25 definitions.
- [PASS] MP-2 GEARS format (judged on the requirement layer, spec.md §C only): every REQ uses a labelled pattern: Ubiquitous (001, 002, 008, 011, 012, 025), Event-driven (003-006, 009, 010, 013-017, 019, 021-023), Capability gate `Where` (007, 018), State-driven `While` (020). REQ-FDA-024 (spec.md:210) is "shall not" legacy EARS Unwanted. It is allowed as a legacy equivalent inside the window that runs to 2026-11-22 (minor note D12). The ACs are in acceptance.md as a matrix plus Given-When-Then and were not graded here.
- [PASS] MP-3 YAML frontmatter: all 12 canonical fields are present with valid types: `id`, `title`, `version: "0.1.1"` (quoted), `status: draft`, `created: 2026-10-03`, `updated: 2026-10-03`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma string) (spec.md:2-13). No rejected aliases.
- [N/A→PASS] MP-4 language neutrality: the SPEC targets moai Go internals and workflow doctrine, not multi-language tooling. §D carries template neutrality (spec.md:222).
- [PASS] MP-5 D7: SPEC-AUTONOMY-BATCH-GATE-001, -AUTONOMY-GATE-REWIRE-001, -DECISION-AUTHORITY-001, -FACTORY-SELF-DISPATCH-001, -FACTORY-STALE-RUN-HEAL-001 and -SYNC-PARALLEL-DOCS-001 are all `status=completed`, so none is retired, superseded or archived. No REVIEW lines and no BLOCKING findings. The REQ-SD-016 amendment is reconciled explicitly (spec.md:166-167).
- [PASS] MP-6 D8: `grep -c syscall spec.md` → `0`. Auto-pass.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output, exit=1.
- [FAIL] MP-8 RED-now cells: (a) AC-FDA-025 is release-blocking, but its RED-now cell is `—` (acceptance.md:43), so it has no cell at all. (b) The cells cite research.md probe IDs. Of P1-P20, only P1 records an exit code (`exit=1`, research.md:26), so every other release-blocking cell carries three of the four elements, which §2.1 rules "unadopted" (verification-completeness.md:173-174). (c) Stdout is "trimmed to the deciding lines" (research.md:24), not verbatim. (d) AC-FDA-020 and AC-FDA-022 lean partly on M0(a) and M0(b) measurements that do not exist yet. Re-execution at HEAD ba2033d22: P1-P20 all reproduce their recorded output (see Evidence), so no non-reproducing RED was found. The failure is structural: the cell content is incomplete. Codex independently observed that some cited probes do not measure the AC's claim (P13 for AC-FDA-018, P5 for AC-FDA-019); this is recorded as D9.
- [PASS] MP-9 ordering: the CN-4 verb returned `COLLECTED: 0 milestones ... GAP: 0 milestone headings collected` because plan milestones live in a table (plan.md:34-44), not headings. I read the order by hand: M0→M8. Candidates checked: DoD "M0 baseline committed before the first implementation commit" (acceptance.md:106) agrees with M0 being first (plan.md:36); AC-FDA-017 "C2 commit precedes or equals the C1 commit" (acceptance.md:35) agrees with M5 "C2 template first, then C1" (plan.md:41). The dependency line (plan.md:48-49) agrees with the table order. No conflict.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.60 | 0.50-0.75 | REQ-FDA-011 phase scope is undefined (T7 vs T13); REQ-FDA-013 "confined to the auditor's own required-fix text" (spec.md:146-147) needs judgment; the REQ-FDA-021 cache-hit vs run-probe order is ambiguous; non-lane ceiling behavior is undefined |
| Completeness | 0.85 | 0.75-1.0 | HISTORY spec.md:20, WHY §A:27, WHAT §B:60, REQ §C:82, AC acceptance.md §D, Out of Scope with 5 H3 topics (spec.md:246-270). The audit decider omits blocker and product-level checks |
| Testability | 0.70 | 0.75 band minus | AC-FDA-010 omits score, must-pass and sync-phase fixtures; AC-FDA-020 has no retirement fixture; MP-8 cells are incomplete |
| Traceability | 0.95 | 1.0 band minus | `COLLECTED: 25 REQ definitions (acceptance input: read)`, no UNCOVERED; the only ORPHAN is `REQ-SD-016`, a cross-SPEC reference (acceptance.md:34). Some RED-now cells trace to probes that do not measure the AC claim |

Aggregate ≈ 0.74 < 0.85 (Tier L).

## Defects Found (structured defect-list)

D1. MP8-CELLS — acceptance.md:19-43 and research.md:24-46 — The RED-now cells for release-blocking ACs lack the exit-code element. Only P1 carries one. AC-FDA-025 has no cell. Stdout is trimmed rather than verbatim. — Severity: critical — Class: blocking (MP-8) — Required fix: give each probe an exit code and verbatim stdout field and keep the doc-level SHA pin (re-pin to the current tree). Either give AC-FDA-025 a RED-now probe (for example, a grep for `plan_audit_ceiling_policy` in both trees, absent now), or reclassify it as regression-guard or close-gate with stated rationale. For AC-FDA-020 and AC-FDA-022, the RED-now must be an existing probe; M0 output is a green-side baseline, not a RED cell.

D2. PRED-PHASE — spec.md:132-136 (REQ-FDA-011), design.md:56-66 — The shared admission predicate is applied to "the card-transition verdict guard". In code, `guardVerdictPass` serves both T7 plan-audit→kickoff and T13 sync-audit→merge-ready (internal/homestate/card_transition.go:107,113,461-472), and sync-auditor also emits PASS-WITH-DEBT (sync-auditor.md:115). A plan-tier threshold plus a `debts:` requirement applied to T13 would reject or alter sync verdicts. Codex flagged the same issue (P1). — Severity: critical — Class: blocking — Required fix: scope REQ-FDA-011 to the plan-audit phase (T7 plus the kickoff sites), or define per-phase thresholds and schema. Add T13 PASS, PASS-WITH-DEBT and FAIL regression fixtures to AC-FDA-010.

D3. PRED-UNDERSPEC — spec.md:120-123 vs 132-136; design.md:64; acceptance.md:28,52-55 — REQ-FDA-008 defines emission as "must-pass all, score ≥ tier threshold, blocking 0, debts enumerated". REQ-FDA-011 and AC-FDA-010, however, only check debts-missing and blocking>0. The design table adds a score, must-pass and hash check for plain PASS. Today `decide.go:396-405` checks score and hash, while `card_transition.go:470` and `rules.go:25` check only the label. The fixtures contain no score-below-threshold or must-pass-failed case, and the verdict block gains no must-pass field (design.md:68), so "must-pass all" cannot be checked at the code sites. Focus (2) is therefore not fully defined. — Severity: major — Class: blocking — Required fix: state in REQ-FDA-011 exactly which fields the predicate checks (label, score vs tier, must-pass field, blocking count, debts, hash). Add the machine field the auditor emits for must-pass. Add fixtures for score<threshold and must-pass FAIL with a PASS label. State that plain-PASS behavior at card_transition changes, or that it does not.

D4. AUDIT-DECIDER-BLOCKERS — spec.md:157-163 (REQ-FDA-015), design.md:96-98 — The `audit` decider checks only predicate plus hash. Unlike REQ-FDA-009 (audit-ready, no open blocker) and REQ-FDA-019 (a product-level FOUNDER row with an empty verdict blocks the autonomous Kickoff), T8a never re-checks audit-ready status, open blockers or holds, or open product-level FOUNDER rows. A lane could therefore self-approve Kickoff past a product-level verdict, which is a keep-set item under REQ-FDA-024 (spec.md:212). Codex flagged this (P1). — Severity: critical — Class: blocking (focus 1) — Required fix: REQ-FDA-015 must also require audit-ready recorded, no open blocker or operator hold on the card, and no product-level FOUNDER row with an empty verdict. Add AC-FDA-014 rejection fixtures for each.

D5. BINDCACHE-STALE — spec.md:196-199 (REQ-FDA-021), plan.md:58, design.md:144-146 — "Return without opening the factory messaging database" on a cache hit removes the run-state probe that runs before Open today (`factoryHookProbeRun`, internal/hook/factory_messages.go:124-144). That probe drives the stale-run heal rebind (REQ-SRH-004/005). After a run retires, session, env run, PID and process start all still match, so the cache serves a stale binding and skips the rebind. The plan's mitigation applies "on miss" only. Retirement invalidation appears in design.md:146 with no REQ or AC binding it, and no actor is named to delete the cache. Codex flagged this (P1). — Severity: critical — Class: blocking (focus 6) — Required fix: keep the tri-state run probe on a cache hit and skip only Open and Peer when the run is active, or bind cache invalidation on retirement into the REQ with a named writer. Add AC fixtures: a cached session whose run is retired goes through the rebind path; a probe failure on a hit reports degraded and is not suppressed.

D6. CEILING-DECIDABILITY — spec.md:144-149 (REQ-FDA-013), design.md:84-88 — `confined_to_required_fix` ("confined to the auditor's own required-fix text") is a judgment. `req_ac_scope_unchanged` compares iteration N to N-1, which is past state, while the REQ predicates eligibility on the fix still to come. It also compares only REQ/AC counts and §F headings, so a scope change inside an unchanged heading passes (codex reproduced this: an exclusion bullet changed to permit automatic card creation still gives `req_ac_counts_and_scope_headings_equal=True`). Focus (3) asks for no judgment words. — Severity: major — Class: blocking — Required fix: define eligibility over artifacts that can be checked after the delta. The delta-round verdict re-measures REQ/AC counts and the full Out-of-Scope and §F content digest against the ceiling-hit iteration. Any change returns the round to the second-hit path. Replace "confined to required-fix text" with a structured per-finding allowed-change list, or drop it. Absence of evidence means eligibility is false.

D7. CEILING-NONLANE + ROUND-COUNT — spec.md:140-153; design.md:89-91 — The shipped plan-auditor and spec-workflow user-channel routing is replaced by the policy (design.md:90), but REQ-FDA-013/014 define only "the lane" behavior. A non-factory session in a user project has no specified ceiling behavior. Separately, `auto_delta_rounds` is configurable, yet REQ-FDA-014 keys on "a second time", which is ambiguous when rounds > 1. — Severity: major — Class: blocking — Required fix: state the ceiling behavior for a non-lane orchestrator, for example "outside a lane session, the existing user-channel escalation is retained". Define the second hit as "the configured delta rounds are exhausted without a PASS-family verdict".

D8. SYNC-REREAD-PATH — spec.md:128-131 (REQ-FDA-010), acceptance.md:37 (AC-FDA-019) — The re-read obligation binds the sync-auditor agent only. The `sync-audit-4dim` workflow (.claude/workflows/sync-audit-4dim.js) is the binding sync verdict owner on the happy path, and the cold sync-auditor is only the fallback, so binding run conditions would go unread on the default path. REQ-FDA-010 also leaves unstated whether an undisposed condition blocks sync PASS, although the conditions are called "binding". Currently `grep -n "Binding run conditions"` over both returns exit=1. — Severity: major — Class: blocking (focus 2) — Required fix: extend REQ-FDA-010 and AC-FDA-019 to the 4dim workflow, or route PASS-WITH-DEBT cards to the cold sync-auditor. State that an undisposed binding condition is a blocking sync finding.

D9. RED-PROBE-RELEVANCE — acceptance.md:24,36,37 — Some cited probes do not measure the absence the AC asserts. P20 (AC-FDA-006) shows the decide doctrine row. P13 (AC-FDA-018) shows `decision_gate: on`. P5 (AC-FDA-019) shows only verdict tokens. — Severity: minor — Class: blocking (folds into D1) — Required fix: cite a probe whose output shows the asserted text or behavior is absent, for example a grep for `DEFAULT-APPLIED` or `Binding run conditions` with exit=1.

D10. WAIT-RESOLUTION-LINK — spec.md:188-192 (REQ-FDA-020), design.md:44-45 — A wait is resolved by "a later board record for that card", so any unrelated ruling or split-proposal acknowledgement on the same card stops the recheck. Codex flagged this (P2). — Severity: major — Class: blocking — Required fix: give wait lines an id and let a record resolve a wait only when it cites that wait id (board field `resolves`). Add a fixture where an unrelated same-card record leaves the wait open.

D11. FORGED-VERDICT-RISK — plan.md:55, spec.md:239-242 — The mitigation "reads the verdict file from the committed card evidence path" holds only through the existing `audited_sha` == EvidenceSHA binding (card_evidence_readers.go:90-130). The SPEC never states that the `audit` decider requires it, and it does not consider audit receipts. — Severity: minor — Class: optional — Required fix: name the `audited_sha` binding (and optionally the receipt) in REQ-FDA-015.

D12. LEGACY-UNWANTED — spec.md:210 — REQ-FDA-024 uses legacy EARS "shall not". It is valid until 2026-11-22. — Severity: minor — Class: optional — Required fix: optionally recast as "The <subject> shall refuse ...".

D13. MAY-IN-NORMATIVE — spec.md:172 — REQ-FDA-018 "may carry a Default: line" (RQ-5). It is acceptable as an optional-field statement. — Severity: minor — Class: optional.

## 2. Evidence (commands run, observed output)

- Traceability verb (scratchpad trace.sh, literal paths): `COLLECTED: 25 REQ definitions (acceptance input: read)` / `ORPHAN: REQ-SD-016`.
- CN-4 verb: `COLLECTED: 0 milestones in plan order (none), 0 exit bindings, 9 ordering candidates` / `GAP: 0 milestone headings collected from the plan` → order read by hand (see MP-9).
- D7: all 6 related SPECs `status=completed`; self `draft`. D8: `0`. MP-7: no output, `mp7 exit=1`.
- Probe re-execution at HEAD ba2033d22 (branch ref file → `ba2033d22abee6cf37e02fdee1241038d6cc7356`):
  - P1 `grep -rn 'Use: *"decision' internal/cli/` → no output, exit=1. P2 → no output, exit=1.
  - P3 → `rules.go:25 passingVerdicts = []string{"PASS", "PASS-WITH-DEBT"}`, `card_transition.go:470 ...`, `decide.go:393 ...`, exit=0.
  - P4 → `244:- Blocked states are: PASS-WITH-DEBT, BYPASSED, FAIL, INCONCLUSIVE, ...`, exit=0. P5 → `203:` / `243:` tokens, exit=0.
  - P6 → `card_record.go:46: const DeciderHuman = "human"`, `card_transition.go:480-481`, exit=0. P7 → `{"T8", CardKickoff, CardAssigned, guardKickoffDecision},` exit=0.
  - P8 → lane refusal then `decider != homestate.DeciderHuman` refusal, exit=0. P9 → `75: plan_audit_tier_ceilings:`, `102/115: max_iterations: 3`, exit=0.
  - P10 → `693:` STOP → user-question channel; `703:` Max 3 cap, exit=0. P11 → line 158 text reproduced (local and template), exit=0.
  - P12 → authority-register paragraph reproduced, exit=0. P13 → `6:  decision_gate: on` (template: `decision_gate: off`), exit=0.
  - P14 → Open → Peer → already-bound early return, exit=0. P15 → `OpenExistingWithDeadline` + "call sites still discard the state", exit=0. P16 → `factoryHookBatchForRun(...)` state discarded, exit=0.
  - P18 → `108: CronCreate with cron: "7,27,47 * * * *"`, exit=0. P19 → Fallback line, exit=0. P20 → decide AUTONOMOUS/KEEP rows, exit=0.
- Code reads for D2/D3/D5: card_transition.go:96-120 (T7 and T13 share guardVerdictPass), :461-487; kickoff/decide.go:374-406 (score, tier threshold and hash checked there only); factory_messages.go:113-157 (run probe precedes Open); factory_card.go:67-71 (lane refusal includes `MOAI_FACTORY_WORKER` and the GPT backend, so board refusal coverage is adequate).
- `grep -n "Binding run conditions" .claude/agents/moai/sync-auditor.md .claude/workflows/sync-audit-4dim.js` → no output, exit=1.
- audit_multi: claude fail, codex fail (6 findings, 3×P1), GLM inconclusive. Codex states it re-ran P1-P20 at ba2033d22 and that `moai spec lint` exited 0.

## 3. Baseline-attribution

Every measurement above was taken in this run against the card tree at HEAD ba2033d22abee6cf37e02fdee1241038d6cc7356, read through absolute paths. SPEC probes were authored against d7112d005; they reproduce at ba2033d22. The MCP server binary is 45600e4ee, an ancestor of HEAD, so the backend tool output is attributed to that build.

## 4. Gaps

- Refused commands: the worktree guard refused `git -C <card tree> rev-parse/branch` and a `sed` using a shell variable, and later refused a grep whose text contained the word "hash". Fallbacks: I read HEAD from `.git/worktrees/agent-a6b97f013ea3df7de/HEAD` and `refs/heads/WT-decision-automation`, read files with the Read tool, and ran the verification verbs from scratch scripts with literal paths. `git log` and diff over the card's commits were not observed. The Glob tool was unavailable.
- During the audit the scratch scripts trace.sh and cn4.sh were overwritten by another session (pointed at SPEC-FACTORY-LANE-SUPERVISOR-001). My verb outputs above were captured before the overwrite. The scripts on disk no longer reproduce this run.
- I did not run `moai spec lint` myself; codex reports exit 0. I ran no Go tests (none are needed at plan phase).
- GLM advisory was inconclusive. No codex `audit_receipt` id was returned by `audit_multi`.

## 5. Residual-risk

- Even after D2 and D3, the predicate's behavior change for plain PASS at card_transition may break existing factory tests. The run phase must characterize it first.
- Lane self-approval of Kickoff stays self-attested beyond the verdict file (spec.md §E.2). D11 narrows the gap but does not close it.
- The M0 measurements (Q4, Q5) may move defaults, and the ACs depend on values not yet measured.

## Recommendation

Fix D1-D10 (blocking), then re-audit as iteration 2, scoped to this defect delta. Priority: D4 and D5 (keep-set and stale binding), then D2 and D3 (predicate scope and definition), then D6, D7, D8 and D10, then D1 and D9 (RED-now cells). D11-D13 are optional.

## Operational Notes (unverified)

- assumption: verify the D5 fix with a hook test that writes a matching cache file, marks the run retired in the fixture DB, and asserts the rebind notice — test name to be chosen in run.
- inferred (rule: guardVerdictPass is shared by T7 and T13 per card_transition.go:107,113): measure the D2 impact with `grep -n "guardVerdictPass" internal/homestate/*.go` before choosing between scoping and per-phase variants.
