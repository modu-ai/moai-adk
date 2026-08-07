# Plan — SPEC-LSEL-LOCAL-EVOLUTION-001

> **Source-of-truth report:** `.moai/reports/moai-local-self-evolution-design-20260804.html` (v2, ~626 lines). This plan exceeds the rebuild-execution-plan structure: (a) measured ground-truth stats, (b) milestone map M1-M5 with per-milestone scope / audit-gap-closed / reused-assets / critical-path, (c) gate status (which REQs/ACs land in which milestone), (d) continuous-execution protocol, (e) immediate next step. The LLM authors the markdown SPEC; this is a plan, not an explainer.

## §A. Context

### §A.1 Ground-truth stats (report §1, verified 2026-08-04)

| Metric | Value | Evidence |
|--------|-------|----------|
| `usage-log.jsonl` events | 110,403 | report §1 stat grid |
| `lessons-inbox.jsonl` stubs / 25 days | 569 | report §1; writer-only non-test Go: `internal/hook/failure_observer.go:129` |
| drained stubs | ZERO | report §1 |
| `Applier.Apply` runs | 0 | `manifest.jsonl` absent; `enableTriggerInjectionWrites=false` at `internal/harness/applier.go:22` (verified this session) |
| `CuratorDispatch` production callers | 0 | type at `internal/harness/curator_dispatch.go`; report §2 grep |
| `drain`-marking Go code | 0 | `drain-marking` literal only in `moai-constitution.md:147` (verified this session) |
| abandoned drafts | 72 | report §1 |
| evolvable surfaces | 6 | report §5 (listed in spec.md §B.3) |
| inbox noise share | ~63% | `tool_failure:Bash:UnknownFailure` (timeout / sandbox) — report §2 |

Every number above is attributable to either report §1 stat grid or a mechanical grep named in §B.4 below. No carry-over from unrelated prior measurement.

### §A.2 Current branch + state

- Branch: `feat/SPEC-LSEL-LOCAL-EVOLUTION-001` (run-phase; the formerly-referenced `feat/SPEC-RALPH-CONFIG-REDESIGN-001` was merged at `32150f131` before run-phase began).
- HEAD: per session-start status; re-read before any commit per main-checkout-branch-guard staleness rule.
- SPEC dir: `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/` — newly created.

### §A.3 PRESERVE list (this SPEC writes NOTHING here)

- `internal/template/templates/**` — FROZEN distributed doctrine source.
- `.claude/rules/moai/**`, `CLAUDE.md` — FROZEN doctrine.
- `.claude/agents/moai/**` — FROZEN retained agents.
- `moai*` / `moai-*` skills — FROZEN template-managed skills.
- `internal/harness/applier.go`, `internal/harness/curator_dispatch.go` — FROZEN Go applier layer (BYPASS, not unfreeze).
- `.moai/config/sections/**` — merge-managed evolvable in principle, but NO new section file created (§B.4 v2 correction).
- All other SPEC directories (`.moai/specs/SPEC-*/` — including parallel RALPH/LSEL sessions).

### §A.4 EXTEND targets (the only surfaces this SPEC adds to)

- NEW: `.claude/skills/hns-lsel-curator/SKILL.md`, `.claude/skills/hns-lsel-applier/SKILL.md` (mechanism skills, hns namespace).
- NEW: `.claude/agents/harness/hns-lsel-curator-specialist.md` (OPTIONAL — large drain/reflection passes).
- NEW: `.moai/hooks/lsel-apply.sh`, `.moai/hooks/lsel-signal-capture.sh` (user-owned hook scripts, NOT template `.claude/hooks/moai/`).
- NEW: `.moai/state/lsel/` directory (`drain-offset.json`, `signals.jsonl`, `clusters.json`, `proposals/<id>/...`, `apply-ledger.jsonl`).
- NEW: a frozen allowlist meta file OUTSIDE the 6 evolvable surfaces (location [DECISION] — see §B.1 below).
- NEW: `.moai/logs/lsel-reject.log` (rejection audit).
- NEW: `CLAUDE.local.md §28` + INVARANTS kernel block at file top (the only `CLAUDE.local.md` addition — ~40 local lines, zero shipped-doctrine change).
- EDIT: `.claude/settings.local.json` (hook-registration subblock — runtime-managed, never templated; protected by REQ-LSEL-002 CSA forced gate).
- EDIT: `internal/template/internal_content_leak_test.go` (EXTEND — add `CLAUDE.local.md` internal-marker + `lsel`/`hns-lsel` leak guards; NOT a distributed-doctrine edit, this is a CI guard extension).
- EDIT: `.github/workflows/` — add named workflow files for the lsel-leak + CLAUDE.local.md internal-marker guards (report §11 mustFix B#7 — must be CI, not "local test").

## §B. Known Issues + [DECISION] markers

### §B.1 [DECISION] Frozen allowlist meta file location

The report §7 / §11 mustFix A#1 mandates the allowlist live "outside the 6 evolvable surfaces" but does NOT name the exact path. Candidates:

- (a) `.moai/state/lsel/frozen-allowlist.json` — but `.moai/state/lsel/` IS one of the 6 evolvable surfaces, so this VIOLATES the invariant.
- (b) `.claude/lsel/frozen-allowlist.json` — a new directory under `.claude/` that is NOT `.claude/rules/moai/`, NOT a skill, NOT an agent, NOT `.claude/hooks/moai/`. Closer to compliant.
- (c) inline as a `const` in `internal/template/split_namespace_test.go` or a new `internal/lsel/allowlist.go` — the allowlist becomes Go-source-tracked, edited only via Template-First PR. SAFEST but couples the user-owned loop to Go-source.

**[DECISION]: (b) `.claude/lsel/frozen-allowlist.json`** — outside the 6 evolvable surfaces, outside template-managed doctrine, git-tracked for audit. The applier loads it read-only; REQ-LSEL-002 adds ANY edit to it as a CSA forced-gate target. Flagged for plan-auditor challenge: option (c) is more defensible "the loop cannot amend its own handcuffs" but costs a Go-source coupling this SPEC otherwise avoids.

### §B.2 [DECISION] Simulation harness for M5 override-rate proof

Report §13 caveat 4 + REQ-LSEL-015 require a NAMED simulation harness proving override-rate reduction vs static default BEFORE M5 ships. The report does NOT specify the harness. **[DECISION]: defer harness specification to M5 entry** — M5 is explicitly conditional and may never ship; specifying the harness now is premature. M5 entry pre-condition: a harness SPEC (replay last N sessions' decisions under both static-default and Thompson-sampled policies, count overrides) is authored and approved.

### §B.3 [DECISION RESOLVED] — Distributed dead-verb retirement is OUT OF SCOPE (sequential companion SPEC)

Report §11 mustFix B#4 demands the SPEC "drive the existing opt-in `execute` path and discard the parallel applier, OR explicitly retire the distributed dead verbs in the same SPEC". **Decision (orchestrator AskUserQuestion round, 2026-08-04): SEQUENTIAL companion SPEC.** The distributed dead-verb retirement (`moai harness apply` / `execute` / `propose` CLI shipped to all 16-language users) is OUT OF SCOPE for THIS SPEC and deferred to a SEQUENTIAL follow-up companion SPEC, placeholder ID `SPEC-LSEL-DEAD-VERB-RETIRE-001` (explicitly NOT yet authored — a future follow-up). Boundary rationale (per design-report §13 caveat 1): the local bypass has independent value and ships first (it unblocks the 569-stub backlog drain in user-owned surfaces without waiting on the distributed-CLI retirement decision); distributed retirement is a separate concern whose scope (drive the existing opt-in `execute` path vs. explicitly retire the verbs) warrants its own design pass. This SPEC does NOT claim to close the distributed surface (REQ-LSEL-016).

### §B.4 Verification-claim integrity grounding

Every stat in §A.1 is re-verified at run-phase start by re-running the named greps:
- `grep -n "enableTriggerInjectionWrites" internal/harness/applier.go` → line 22, value `false`.
- `grep -rn "drain-marking\|drainMarking" internal/ .claude/` → only `moai-constitution.md:147`.
- `grep -rn "CuratorDispatch" internal/ | grep -v _test.go` → type definitions only (no production caller).
- `wc -l .moai/lessons-inbox.jsonl` → baseline ~569 (re-measured at M1 start; the figure is a moving count).

## §C. Pre-flight (before M1)

- Re-verify §A.1 stats have not drifted (the above greps; if `CuratorDispatch` gained a production caller since 2026-08-04, M2's "verify-firing" AC adapts).
- `make build` baseline green.
- `go test ./...` baseline green.
- `golangci-lint run --timeout=2m` baseline recorded (to distinguish NEW findings from pre-existing).
- `git status --porcelain` clean (no in-flight untracked under the EXTEND targets).
- Re-read `git rev-parse HEAD` + `git branch --show-current` immediately before the plan-phase commit (main-checkout-branch-guard staleness rule).

## §D. Constraints

- Template-First: zero edits under `internal/template/templates/**`. The EXTEND to `internal/template/internal_content_leak_test.go` is a CI-guard extension, NOT a distributed-template edit (the test asserts leak prevention; it does not appear in distributed user projects).
- Namespace: all new artifacts under `hns-*` / `.claude/agents/harness/` / `.moai/state/lsel/` / `.claude/lsel/` / `.moai/hooks/`.
- PR-mandatory: every `lsel-*` apply commit lands via feature branch + PR (Tier S self-merge).
- Subagent boundary: LSEL mechanism skills NEVER call `AskUserQuestion`; they return blocker reports. The orchestrator owns the user gate.
- `lsel-apply.sh` is playback-only (REQ-LSEL-008).
- No new `.moai/config/sections/*.yaml` file (§B.4 v2 correction).
- Settings.local.json hook-registration subblock is a CSA forced-gate target (REQ-LSEL-002).

## §E. Self-Verification (plan-phase)

- [x] SPEC ID regex PASS (executed: `SPEC-LSEL-LOCAL-EVOLUTION-001` matched `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → PASS).
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags). `phase: "v3.x target"` — release target, not a lifecycle token.
- [x] Out of Scope section carries 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule` lint).
- [x] Every REQ uses a GEARS pattern (Ubiquitous / When / While / Where / Event-detected). No `IF/THEN`.
- [x] Every AC in acceptance.md is Given-When-Then and binary-testable.
- [x] No implementation detail in spec.md (function names are evidence, not design; the frozen-allowlist location decision lives HERE in plan.md §B.1, not in spec.md).
- [x] 16 REQs / 16 ACs — at the Tier M ceiling but within it.

## §F. Milestones (M1-M5 mapped from report §10 P1-P5)

Critical-path ordering: **M1 → M2 → M3** (the seam closure; M3 is the load-bearing milestone). **M4** hardens the closed loop against storage rot. **M5** is a conditional generalization layer. M1-M2 are the cheapest gaps; M3 is the heaviest (introduces the live write path).

### §F.1 M1 — Drain closure (report P1)

**Scope:**
- `hns-lsel-curator/SKILL.md` with drain + cluster + importance-gate logic.
- `drain-offset.json` companion-offset semantics (append-only inbox preserved; offset marks consumed stubs).
- Write-time coarse severity hint (in the `failure_observer.go` inbox writer OR as a drain-side filter — **[DECISION]: drain-side filter**, NOT a `failure_observer.go` edit, because `internal/hook/failure_observer.go` is outside the six loop-writable surfaces enumerated in spec.md §B.3, so the LSEL loop cannot write to it). Report §11 mustFix B#3.
- First real drain of the 569-stub backlog → candidate `feedback_*.md` topic files staged in `.moai/state/lsel/clusters.json` (NOT yet written to `memory/`).

**Audit gap closed:** "drain-marking mechanism = Go code 0 / constitution paragraph only" (report §2). Model-mediated skill over JSONL → 0 Go changes, user-owned surface only.

**Reused assets:** existing inbox schema, existing auto-memory, SPEC-HARNESS-RATCHET-REWIRE-001 D3 companion-offset recommendation (cited in report §10 P1), existing Lessons Protocol.

**REQs/ACs landing here:** REQ-LSEL-009, REQ-LSEL-010 (drain behavior + noise filter). AC-LSEL-009, AC-LSEL-010.

**Critical-path note:** M1 is the cheapest gap and the evidence-generator for M2-M4. Shipping M1 alone is a valid stopping point (it converts 569 stagnant stubs into candidate topics even without APPLY).

### §F.2 M2 — PROPOSE shadow + INVARANTS kernel (report P2)

**Scope:**
- Extend `hns-lsel-curator` to emit shadow proposals at `.moai/state/lsel/proposals/<id>/{proposal.md, diff.patch, self-critique.md}`.
- Retrieval-before-propose (read relevant `feedback_*.md` BEFORE drafting).
- **VERIFY the existing `moai-harness-learner` Tier-4 AskUserQuestion flow actually fires** before wiring depends on it (grep + behavioral test; the audit's "0 callers" pattern is the cautionary precedent — do NOT assume the Tier-4 path is live).
- Add `CLAUDE.local.md §28` (LSEL operating instructions, ~40 local lines) + INVARANTS kernel block at the top of the file (read-only goal kernel per report §7 last bullet).
- Mechanical trigger wiring (REQ-LSEL-007): SessionStart backlog check + default-on `/loop` recipe. This is the M2 anti-feature-decay step (report §11 mustFix B#1).

**Audit gap closed:** "72 drafts go nowhere / CuratorDispatch caller 0" — the PROPOSE half finally has a consumer (report §10 P2).

**Reused assets:** `moai-harness-learner` skill (existing Tier-4 flow — IF verified-firing), `CuratorDispatch.Prepare()` scaffold as read-only reference (NEVER called), existing `ToolSearch` + AskUserQuestion channel monopoly.

**REQs/ACs landing here:** REQ-LSEL-005 (CSA forced gate), REQ-LSEL-006 (namespace CI guard), REQ-LSEL-007 (mechanical trigger), REQ-LSEL-010 (propose shadow). AC-LSEL-005, AC-LSEL-006, AC-LSEL-007, AC-LSEL-011, AC-LSEL-012.

### §F.3 M3 — APPLY closure via bypass (report P3) — CRITICAL PATH

**Scope:**
- `hns-lsel-applier/SKILL.md` — APPLY engine.
- Frozen allowlist at `.claude/lsel/frozen-allowlist.json` (decision §B.1).
- `lsel-apply.sh` hook script (registered in `settings.local.json`) — playback-only consumer of approved `decision.json` queue (REQ-LSEL-008).
- `apply-ledger.jsonl` — the manifest the frozen applier never produced, finally real in user-owned space.
- First `lsel-<proposal-id>`-tagged git commit on a feature branch (PR-mandatory).
- **Rollback-rehearsal AC** (REQ-LSEL-012): `git revert <lsel-tag>` lands clean on a mixed `CLAUDE.local.md` history (lsel + manual edits).

**Audit gap closed:** "`applier.go enableTriggerInjectionWrites=false` + manifest absent" — the audit's central finding. Closed BY BYPASS (report §10 P3): the frozen Go applier stays frozen (would write to distributed surfaces), the parallel user-owned applier writes only to the 6 evolvable surfaces.

**Reused assets:** `branch_guard.go` path-allowlist + exception-pattern shape (ARCHITECTURE reuse only, not code reuse — report §6 stage 5), existing verification-batch pattern, existing `/moai gate`, git-revert immutable versioning (LangWatch).

**REQs/ACs landing here:** REQ-LSEL-001 (allowlist hard-reject), REQ-LSEL-002 (allowlist location + execution-meta forced gate), REQ-LSEL-003 (frozen applier stays frozen), REQ-LSEL-004 (lsel-tagged commit), REQ-LSEL-008 (queue-only apply), REQ-LSEL-011 (APPLY behavior), REQ-LSEL-012 (rollback rehearsal). AC-LSEL-001, AC-LSEL-002, AC-LSEL-003, AC-LSEL-004, AC-LSEL-008, AC-LSEL-013, AC-LSEL-014.

**Critical-path note:** M3 is the load-bearing milestone — it introduces the live write path. The CSA forced-gate on execution-meta files (REQ-LSEL-002) is the heaviest fix from critique (report §11 mustFix A#1). If M3's rollback-rehearsal FAILS, M3 does NOT ship.

### §F.4 M4 — VERIFY + reflection (report P4)

**Scope:**
- `verify_command` wiring + MANDATORY `/moai gate` superset (proposer-authored verify alone is circular — report §11 mustFix B#6).
- Auto-revert on second non-timeout failure; timeout-class failures retry exactly once.
- Periodic reflection pass (monthly `/loop` recipe or `/moai:harness lsel reflect`) — synthesize abstract-principle `feedback_*.md` from several concrete topic files; originals move to `memory/_archive/` (cold tier, NOT deleted).
- `memory_type` labeling (CoALA): procedural → `hns-*` skills, semantic → `feedback`.
- Decay-weighted retrieval (older lessons fall out of top-5).

**Audit gap closed:** "no consolidation / decay / pruning → un-refined accumulation → wrong-lesson retrieval" — the dominant failure mode every research lens attacks differently (report §10 P4).

**Reused assets:** existing `/moai gate`, existing verification-batch single-turn pattern, `go test -run <pattern>`, MemGPT promote-on-retrieval-hit paging.

**REQs/ACs landing here:** REQ-LSEL-013 (verify independence + flaky policy), REQ-LSEL-014 (reflection). AC-LSEL-015, AC-LSEL-016.

### §F.5 M5 — Personalization + signal capture (report P5) — CONDITIONAL

**Scope:**
- `signals.jsonl` intentional-hint capture (PostToolUse hook): Ctrl+G plan edits, 'Other' free-form overrides, recommendation rejections — 5-10x posterior weight vs passive approvals.
- Thompson-sampling posterior auto-tune of evolvable config (effort / model / harness level); template default = safe fallback arm.
- Hysteresis margin (anti-thrash).
- Bother-cost gate (integrate with cache-aware-execution directive 1).

**CONDITIONAL deploy gate (REQ-LSEL-015):** MUST reduce override-rate vs static default in a NAMED simulation harness OR do not ship. Harness is [DECISION]-deferred to M5 entry (§B.2).

**Audit gap closed:** "recommendation = static template default, observation runs at scale but no personalization" (report §10 P5).

**REQs/ACs landing here:** REQ-LSEL-015 (capability gate). No MUST AC — M5 may never ship.

## §G. Gate Status (which gates fire when)

| Gate | Fires at | REQ/AC |
|------|----------|--------|
| plan-auditor (Phase 0.5) | after plan-phase commit | Tier M threshold 0.80 |
| Implementation Kickoff Approval | plan→run boundary, M1 entry | mandatory, score-independent (§19.1) |
| M2 Tier-4-firing verification | M2 entry | AC-LSEL-012 (grep + behavioral test) |
| M3 rollback rehearsal | M3 ship | AC-LSEL-014 (must pass or M3 holds) |
| M3 CSA forced gate (any execution-meta proposal) | any M1-M5 proposal touching REQ-LSEL-002 surfaces | REQ-LSEL-005 — synchronous AskUserQuestion |
| M4 reflection threshold | reflection pass | REQ-LSEL-014 — accumulation, NOT wall clock |
| M5 conditional deploy | M5 ship | REQ-LSEL-015 — simulation-harness proof |
| Per-apply VERIFY | every apply | REQ-LSEL-013 — `/moai gate` superset + timeout-retry-once |

## §H. Continuous-Execution Protocol

This SPEC spans 5 milestones. Within a milestone, the manager-develop delegation template (Section A-E for Tier M) applies. Across milestones:

- **After M1:** `/clear` + paste-ready resume (the 30K→180K context reset). Memory: `project_lsel_m1_drain_complete.md`.
- **After M2:** `/clear` + paste-ready resume. Memory: `project_lsel_m2_propose_complete.md`.
- **After M3 (critical path):** `/clear` + paste-ready resume. The rollback-rehearsal AC MUST have passed. Memory: `project_lsel_m3_apply_complete.md`.
- **After M4:** `/clear` + paste-ready resume. Memory: `project_lsel_m4_verify_complete.md`.
- **M5 entry pre-condition:** harness SPEC authored + approved (§B.2 [DECISION]). M5 may be deferred indefinitely.

Goal-arming (`/moai goal`): for M1-M4 a per-milestone `ac_converge`-style condition is armable after Implementation Kickoff Approval, scoped to that milestone's AC set. M5 is NOT armable until its harness SPEC passes.

## §I. Anti-Patterns (specific to this SPEC)

- **AP-LSEL-001 — Self-amending allowlist.** Allowlist stored inside any of the 6 evolvable surfaces (e.g. inside `hns-lsel-applier/SKILL.md`). FORBIDDEN by REQ-LSEL-002; the v1 design was caught here (report §11).
- **AP-LSEL-002 — Unfreezing the Go applier.** Setting `enableTriggerInjectionWrites=true` to "make the existing path work". FORBIDDEN by REQ-LSEL-003; unfreezing injects `@MX` into distributed agent surfaces (doctrine corruption).
- **AP-LSEL-003 — "Doctrine says drain" without a mechanism.** Depending on the orchestrator remembering to call drain. FORBIDDEN by REQ-LSEL-007; the audit's exact failure mode.
- **AP-LSEL-004 — Proposer-authored verify alone.** Treating the proposal's `verify_command` as sufficient. FORBIDDEN by REQ-LSEL-013; circular logic.
- **AP-LSEL-005 — Creating `.moai/config/sections/lsel.yaml`.** Wiped on `moai update` (`deploy.go:113`). Loop state lives under `.moai/state/lsel/` (§B.4 v2 correction).
- **AP-LSEL-006 — Treating M1 as "the loop is closed".** M1 closes DRAIN only, not APPLY. Claiming the loop is closed after M1 is an unobserved completion claim.
- **AP-LSEL-007 — Shipping M5 without the simulation-harness proof.** Violates REQ-LSEL-015.
- **AP-LSEL-008 — `lsel-apply.sh` creating new apply operations.** Violates REQ-LSEL-008 (playback-only).

## §J. Immediate Next Step

1. plan-auditor verdict on this plan-phase artifact set (Tier M threshold 0.80).
2. On PASS: Implementation Kickoff Approval (mandatory human gate, §19.1).
3. On approval: `/clear` + paste-ready resume → `/moai run SPEC-LSEL-LOCAL-EVOLUTION-001` enters M1 (drain closure).
4. M1 deliverable: 569-stub backlog drained for the first time; `drain-offset.json` advances; candidate topic files staged under `.moai/state/lsel/clusters.json`; ZERO writes to `memory/` from M1.

## §K. Cross-References

- spec.md §C (16 REQs, GEARS) / §D (16 ACs summary) / §G (Out of Scope — 5 topics).
- acceptance.md §D (16 Given-When-Then AC definitions).
- progress.md §E.1 (plan-phase audit-ready signal — awaiting plan-auditor verdict).
- Design report (SSOT): `.moai/reports/moai-local-self-evolution-design-20260804.html`.
