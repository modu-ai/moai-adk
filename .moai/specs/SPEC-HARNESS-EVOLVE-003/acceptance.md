---
id: SPEC-HARNESS-EVOLVE-003
title: "Curator production wiring — Tier-Surface mapping + validation gates + re-proposal suppression"
version: "0.1.0"
status: completed
created: 2026-07-12
updated: 2026-07-13
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness/safety, internal/harness/curator, internal/harness, internal/config"
lifecycle: spec-anchored
tags: "harness-evolve-epic, curator-wiring, tier-surface, l2-canary, l3-contradiction, negative-evidence, frozen-guard, re-proposal-suppression, glm-observe-only"
era: V3R6
tier: L
depends_on: [SPEC-HARNESS-EVOLVE-002]
---

# SPEC-HARNESS-EVOLVE-003 — Acceptance Criteria Matrix

> Counterpart to `spec.md` (requirements SSOT), `plan.md` (implementation plan),
> `design.md` (architecture decisions), `research.md` (codebase investigation).
> This document owns the machine-verifiable + behavior-verifiable AC matrix.
> Every REQ in spec.md §C maps to at least one AC here. Gates (L1 A1, L2, L3,
> A7) each carry a baseline-0 → ≥1 behavior assertion per the
> `feedback_ac_token_presence_not_reachability` discipline.

## §A. AC Identification Convention

- AC IDs: `AC-HEV3-NNN` (sequential, zero-padded to 3 digits).
- Sub-criteria: `AC-HEV3-NNNa` / `AC-HEV3-NNNb` (paired sub-criteria within one
  logical AC — the trailing lowercase alpha is permitted for ACCEPTANCE
  CRITERIA ONLY, never for the SPEC ID).
- Severity: **Must-pass** (gate reachability + safety invariants),
  **Should-pass** (quality + completeness), **Nice-to-have** (polish).
- Verification mode: **grep** (token reachability), **behavior** (inject input
  → assert gate fires), **build** (compile + cross-platform), **coverage**
  (≥90%).

## §B. AC Matrix

### §B.1 Tier↔Surface mapping activation (A6) — REQ-HEV3-001…004

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-001 | Must-pass | grep | `grep -rn 'auto_detection\|AutoDetectionConfig' internal/harness/curator/dispatch.go` — the dispatch layer references the auto_detection surface as a registered Tier-4 Evolvable surface | 0 → ≥1 |
| AC-HEV3-002a | Must-pass | behavior | Inject a Curator proposal editing `auto_detection.rules.minimal.conditions` with `file_count <= 0` (out of range). Assert the writer returns `ErrAutoDetectionOutOfRange` and the file is NOT touched (byte-hash before == byte-hash after). | 0 → ≥1 rejection |
| AC-HEV3-002b | Must-pass | behavior | Inject a Curator proposal editing `auto_detection.rules.thorough.conditions` with `file_count <= 99999` (above upper bound). Assert `ErrAutoDetectionOutOfRange` + file untouched. | 0 → ≥1 rejection |
| AC-HEV3-003a | Must-pass | behavior | Inject a Tier-3 proposal (5 observations). Assert the write lands in `CLAUDE.local.md` (the append-only Learned surface) AND `CLAUDE.md` byte-hash is unchanged. | dispatch correctness |
| AC-HEV3-003b | Must-pass | behavior | Inject a Tier-4 proposal (10 observations). Assert the write lands in `CLAUDE.md` (the digest managed block) AND `CLAUDE.local.md` byte-hash is unchanged. | dispatch correctness |
| AC-HEV3-004 | Must-pass | behavior | Inject a Tier-3 proposal with `content.Tier = 4` forced. Assert `ErrTierNotQualified` (the writer's self-tier-escalation guard fires at the dispatch layer). | cross-surface leak blocked |

### §B.2 L5 orchestrator approval round — REQ-HEV3-005…007

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-005 | Must-pass | grep | Per-token verification across the 3 L5-contract tokens (NOT a single compound count — a single token repeated 3× must NOT mechanically satisfy this AC): `grep -rn 'WriteManagedBlockGated' internal/harness/curator/dispatch.go internal/harness/applier.go` ≥1 AND `grep -rn 'ApprovalDecision' internal/harness/curator/dispatch.go internal/harness/applier.go` ≥1 AND `grep -rn 'RejectionRecorder' internal/harness/curator/dispatch.go internal/harness/applier.go` ≥1 — the production wiring threads all three L5-contract tokens | 0 → ≥1 per token (3 tokens) |
| AC-HEV3-006 | Must-pass | grep | `grep -rn 'WriteManagedBlock\b' internal/harness/curator/dispatch.go internal/harness/applier.go | grep -v WriteManagedBlockGated` — NO direct `WriteManagedBlock` call bypassing the gate | 0 → 0 (must remain 0) |
| AC-HEV3-007a | Must-pass | behavior | Inject an L5 `ApprovalDecision{Approved: false, Rationale: "test"}`. Assert the file is NOT touched AND a `LineageEntry` with outcome `rejected` is appended AND the A7 registry gains an entry. | rejection path |
| AC-HEV3-007b | Must-pass | behavior | Inject an L5 `ApprovalDecision{Approved: true}`. Assert the file IS written AND the lineage outcome is `applied`. | approval path |

### §B.3 PRODUCTION wiring (TierGatedWrite / WriteManagedBlockGated) — REQ-HEV3-008…011

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-008 | Must-pass | grep | `grep -rn 'curator.TierGatedWrite\|TierGatedWrite(' internal/harness/ internal/cli/ internal/cmd/ | grep -v _test.go | grep -v 'internal/harness/curator/'` — production caller count | **0 → ≥1** (the reachability breakthrough) |
| AC-HEV3-009 | Must-pass | grep | `grep -rn 'curator.WriteManagedBlockGated\|WriteManagedBlockGated(' internal/harness/ internal/cli/ internal/cmd/ | grep -v _test.go | grep -v 'internal/harness/curator/'` — production caller count | **0 → ≥1** |
| AC-HEV3-010 | Must-pass | behavior | Inject a 6-observation pattern (Tier 3). Assert the dispatch reads `tier.ClassifyStatus(6) == StatusRule` (Tier 3) and constructs `BlockContent{Tier: 3}`. | tier-read |
| AC-HEV3-011 | Must-pass | grep | `grep -B2 -A2 'TierGatedWrite\|WriteManagedBlockGated' internal/harness/curator/dispatch.go | grep -c '// TODO\|feature_flag\|if false'` — no dead-code / feature-flag guards around the wiring | 0 → 0 (must remain 0) |

### §B.4 L2 Canary activation — REQ-HEV3-012…014, REQ-HEV3-034

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-012 | Must-pass | behavior | Inject a Tier-4 CLAUDE.md proposal. Assert the L2 layer shadow-applies (a temp copy exists during evaluation) — verify via a test hook or the shadow-apply log. | shadow-apply fires |
| AC-HEV3-013 | Must-pass | behavior | Inject a Tier-4 proposal whose shadow-apply regresses a held-out signal (e.g. exceeds the 3K digest budget). Assert `Decision.RejectedBy == 2` AND the real file is NOT touched. | **0 → ≥1 L2 rejection** (the reachability breakthrough) |
| AC-HEV3-014 | Must-pass | behavior | Trigger a Canary veto (post-apply regression detected by `CanaryVeto.VetoAndRollback`). Assert the A7 registry gains an entry with `outcome: rolled-back` for the same pattern key. | veto → registry agreement |
| AC-HEV3-034 | Must-pass | grep | `grep -rn 'EvaluateCanary\|l2CanaryCheck' internal/harness/curator/dispatch.go internal/harness/safety/pipeline.go` — the L2 consult is wired into the Curator path (not just the harness-applier path) | Curator path 0 → ≥1 |

### §B.5 L3 Contradiction activation — REQ-HEV3-015…017, REQ-HEV3-033

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-015a | Must-pass | grep | `sed -n '68,74p' internal/harness/safety/pipeline.go | grep -c 'return harness.ContradictionReport{}'` — the no-op line count. The no-op body MUST be gone (replaced by a real check). | 1 → 0 (no-op removed) |
| AC-HEV3-015b | Must-pass | grep | `grep -rn 'FrozenRules\|frozen_rules\|ConsultFrozen' internal/harness/safety/contradiction.go internal/harness/safety/pipeline.go` — the real Frozen-rules consult is wired | 0 → ≥1 |
| AC-HEV3-016 | Must-pass | behavior | Inject a Curator proposal that contradicts a registered Frozen rule (e.g. proposes editing `.claude/rules/moai/core/`). Assert `Decision.RejectedBy == 3` AND the file is NOT touched. | **0 → ≥1 L3 rejection** (the reachability breakthrough) |
| AC-HEV3-017 | Must-pass | behavior | Inject a contradiction proposal. Assert the rejection `Reason` cites the Frozen-rule identifier (not an opaque "contradiction detected"). | audit-trail |
| AC-HEV3-033 | Must-pass | behavior | Run a full Curator cycle. Assert the L3 evaluator was consulted ≥1 time (via a counter / log / spy), NOT short-circuited by a prior layer. | L3 reachability 0 → ≥1 per cycle |

### §B.6 A7 Re-proposal suppression (negative-evidence registry) — REQ-HEV3-018…022, REQ-HEV3-035

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-018a | Must-pass | grep | `grep -rn 'negative_evidence\|NegativeEvidence\|negative-evidence.jsonl' internal/harness/negative_evidence.go` — the registry data structure exists | 0 → ≥1 file |
| AC-HEV3-018b | Must-pass | build | `test -f internal/harness/negative_evidence.go && go build ./internal/harness/...` exit 0 | compiles |
| AC-HEV3-019 | Must-pass | behavior | Reject a proposal at L3. Assert `.moai/state/negative-evidence.jsonl` gains an entry with `outcome: rejected`, the pattern key, `cooldown_until` set, and `machine_signal_ref` citing the lineage entry. | register-on-reject |
| AC-HEV3-020 | Must-pass | behavior | Rollback a promotion (via `RestoreSnapshot`). Assert the registry gains an entry with `outcome: rolled-back` for the rolled-back pattern key. | register-on-rollback (auto) |
| AC-HEV3-021a | Must-pass | behavior | Reject pattern key `K`. Immediately re-propose `K`. Assert the dispatch returns `ErrReProposalSuppressed` BEFORE reaching L2/L3/L5 (early-block). | **same-key re-proposal blocked** (the A7 breakthrough) |
| AC-HEV3-021b | Must-pass | behavior | Reject pattern key `K`. Add N=3 NEW post-rejection evidences for `K` AND let the cooldown elapse. Re-propose `K`. Assert the proposal proceeds past the A7 block to L2. | re-eligibility after cooldown + N |
| AC-HEV3-022 | Must-pass | grep | `grep -c 'cooldown_until.*null\|cooldown_until.*9999' internal/harness/negative_evidence.go` — no permanent-suppression code path | 0 → 0 (must remain 0) |
| AC-HEV3-035 | Must-pass | behavior | Run a full Curator cycle. Assert the A7 registry was consulted ≥1 time (early-block check) before the L2/L3/L5 chain. | A7 reachability 0 → ≥1 per cycle |

### §B.7 A1 Permission-surface Frozen registration — REQ-HEV3-023…025

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-023a | Must-pass | grep | Per-token verification (design.md §C.2 decides the hooks axis is covered by the `.claude/settings.json` prefix match, so the A1 expansion adds 2 entries, NOT 3): `grep -A8 'var frozenPrefixes' internal/harness/safety/frozen_guard.go \| grep -c 'settings\.json'` ≥1 AND `grep -A8 'var frozenPrefixes' internal/harness/safety/frozen_guard.go \| grep -c 'settings\.local\.json'` ≥1 — both permission-surface tokens present in the expanded list (a single compound `grep -c 'A\|B\|C' ≥3` would mechanically FAIL since only 2 entries are added) | 0 → ≥1 per token (2 tokens) |
| AC-HEV3-023b | Must-pass | grep | `grep -A8 'var frozenPrefixes' internal/harness/safety/frozen_guard.go | grep -c 'frozen_guard'` — the guard source files are self-protected | 0 → ≥1 |
| AC-HEV3-024a | Must-pass | behavior | Inject a Curator proposal targeting `.claude/settings.json`. Assert `Decision.RejectedBy == 1` (L1 Frozen guard fires) AND the file is NOT touched. | **permission-surface L1-blocked** (the A1 breakthrough) |
| AC-HEV3-024b | Must-pass | behavior | Inject a Curator proposal targeting `.claude/settings.local.json`. Assert `RejectedBy == 1`. | permission-surface L1-blocked |
| AC-HEV3-025 | Must-pass | behavior | Inject a Curator proposal targeting `internal/harness/safety/frozen_guard.go` (the guard itself). Assert `RejectedBy == 1`. | guard self-protection |

### §B.8 GLM observe-only guard — REQ-HEV3-026…027

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-026 | Must-pass | behavior | Inject a Tier-3 proposal with `model_class: "glm"` in the session context. Assert the proposal is recorded in the routing ledger (observe-only) AND NO Learned surface is written (no CLAUDE.md / CLAUDE.local.md write). | GLM observe-only |
| AC-HEV3-027 | Must-pass | grep | `grep -rn 'model_class\|ModelClass\|IsGLM' internal/harness/curator/dispatch.go` — the model-class gate is in the dispatch layer (upstream of the writer), NOT in the curator writer package | 0 → ≥1 |

### §B.9 Anti-fabrication + template neutrality — REQ-HEV3-028…029

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-028 | Must-pass | behavior | Construct a negative-evidence registry entry with `pattern_key: "SPEC-HARNESS-EVOLVE-003 was rejected"` (contains an internal SPEC ID). Assert the writer rejects it (anti-fabrication input validation, inherited from EVOLVE-002). | registry anti-fabrication |
| AC-HEV3-029a | Must-pass | grep | `grep -rn 'auto_detection' internal/template/templates/.moai/config/sections/harness.yaml` — the template ships the empty schema (mechanism) | ≥1 (already present, verify unchanged) |
| AC-HEV3-029b | Must-pass | grep | Run the template-neutrality CI guard: `go test ./internal/template/ -run TestInternalContentLeak` exit 0. The auto_detection threshold DATA (a learned correction) is NOT in the template tree. | neutrality guard PASS |

### §B.10 Go quality + subagent boundary — REQ-HEV3-030…032

| AC | Severity | Mode | Verification | Baseline → Target |
|----|----------|------|--------------|-------------------|
| AC-HEV3-030a | Must-pass | coverage | `go test -cover ./internal/harness/safety/... ./internal/harness/curator/... ./internal/harness/... ./internal/config/...` — ≥ 90% on the new/extended packages | ≥ 90% |
| AC-HEV3-030b | Must-pass | build | `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0 | cross-platform |
| AC-HEV3-031 | Must-pass | grep | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/safety/ internal/harness/negative_evidence.go internal/harness/curator/dispatch.go | grep -v _test.go | grep -v '//'` — subagent boundary | 0 → 0 (must remain 0) |
| AC-HEV3-032 | Must-pass | grep | Verify no new hook wrapper: `git diff --name-only main -- .claude/hooks/ .claude/settings.json .claude/settings.json.tmpl` — empty (no hook surface change) | 0 new hook files |

## §C. Severity Gates

- **Must-pass**: ALL Must-pass ACs MUST be PASS for the run-phase to close. Any
  Must-pass FAIL blocks sync.
- **Should-pass**: ≥ 80% of Should-pass ACs PASS.
- **Nice-to-have**: informational; no gate.

## §D. Edge Cases

- **Edge-1 — Concurrent Curator cycle + registry append**: two concurrent
  Curator proposals for DIFFERENT pattern keys. Both should succeed; the
  registry appends atomically (jsonl append-only, per-entry write).
- **Edge-2 — A7 registry file absent on first run**: the registry
  `.moai/state/negative-evidence.jsonl` does not exist on a fresh project. The
  reader MUST handle "file not found" as "empty registry" (not an error).
- **Edge-3 — Cooldown boundary**: a re-proposal at exactly `cooldown_until`
  (not after). Boundary semantics: the cooldown is ELAPSED at the instant
  `now >= cooldown_until` (inclusive).
- **Edge-4 — L1 A1 vs L3 Frozen-rules overlap**: a proposal targeting
  `.claude/rules/moai/` hits BOTH the L1 prefix block AND the L3 Frozen-rules
  registry. L1 fires first (short-circuit); L3 is not consulted. This is
  correct — L1 is the cheaper check.
- **Edge-5 — GLM session + Tier-2 observation**: a GLM session observing a
  1-evidence pattern (Tier 1). Observation IS recorded (Tier 1-2 is always
  accepted regardless of model class); only Tier 3+ promotion is GLM-gated.
- **Edge-6 — Rollback of a multi-surface promotion**: a promotion that wrote
  BOTH CLAUDE.md (Tier 4) and CLAUDE.local.md (Tier 3). Rollback restores
  both snapshot restore units (EVOLVE-002 distinct restore units) AND registers
  TWO negative-evidence entries (one per pattern key per surface).

## §E. Definition of Done

- All Must-pass ACs PASS with observed evidence (verification-claim-integrity
  §1.1 — no unobserved claims).
- Coverage ≥ 90% on new/extended packages.
- Cross-platform build green.
- Subagent boundary grep clean (0 AskUserQuestion references in production code).
- The 4 reachability breakthroughs each verified by behavior ACs:
  - L3: AC-HEV3-016 (inject contradiction → RejectedBy == 3)
  - L2: AC-HEV3-013 (inject regression → RejectedBy == 2)
  - A7: AC-HEV3-021a (reject + re-propose → ErrReProposalSuppressed)
  - A1: AC-HEV3-024a (settings.json target → RejectedBy == 1)
  - Production wiring: AC-HEV3-008/009 (TierGatedWrite + WriteManagedBlockGated production callers 0 → ≥1)
- progress.md §E.2/§E.3 populated by manager-develop with the observed evidence.
