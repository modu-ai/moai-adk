# Plan — SPEC-CC2178-MODEL-POLICY-REPAIR-001

> **Tier recommendation: M (standard)** — 3-axis alignment + phantom-map cleanup + `[1m]` re-verification + config/Go/doc surfaces. Full plan/run/sync with plan-auditor gate. NOT Tier L: if the cycle_type axis (M3) proves too large, it splits into a follow-up SPEC (see §F.3 Split Trigger).

## §A. Context

This plan implements SPEC-CC2178-MODEL-POLICY-REPAIR-001, which repairs 4 verified defects (D1-D4) in the model-policy stack and wires the CC 2.1.175 Default-model cost lever identified by `.moai/research/cc-update-2.1.163-to-2.1.178.md`. The work spans 3 code surfaces:

1. **`internal/template/templates/.claude/settings.json.tmpl`** — the CC 2.1.175 lever (D3).
2. **`internal/template/model_policy.go`** — the phantom-map cleanup (D1, D2) + effort-map redundancy decision.
3. **`.moai/config/sections/quality.yaml` + harness Complexity Estimator integration** — the cycle_type routing (D4).
4. **`.claude/rules/moai/development/model-policy.md`** (+ template mirror) — the `[1m]` re-verification verdict + Default-model documentation.

The plan is organized as 6 milestones (M1-M6), each with an explicit `cycle_type` per the Tier M standard structure. cycle_type for the IMPLEMENTATION milestones is `tdd` (the project default per `quality.yaml`); M1 (research) and M6 (docs) are non-code milestones.

## §B. Known Issues (Pre-flight Ground-Truth, verified 2026-06-16)

### §B.1 D1 — `agentModelMap` stale (18 entries, 3 correct, 15 phantom, 2 missing)

File: `internal/template/model_policy.go:193-216`.

**Retained-catalog entries (CORRECT, keep)**:
- `manager-spec` (L195), `manager-docs` (L198), `manager-git` (L202) — 3 entries.

**Archived phantom entries (REMOVE, 13 keys / 15 entries counting the legacy name aliases)**:
- `manager-ddd` (L196) — legacy alias for `manager-develop`; the canonical name is `manager-develop`.
- `manager-tdd` (L197) — legacy alias for `manager-develop`.
- `manager-quality` (L199), `manager-project` (L200), `manager-strategy` (L201) — archived core managers.
- `expert-backend` (L204), `expert-frontend` (L205), `expert-security` (L206), `expert-devops` (L207), `expert-performance` (L208), `expert-debug` (L209), `expert-testing` (L210), `expert-refactoring` (L211) — all 8 `expert-*` archived.
- `builder-agent` (L213) — legacy alias for `builder-harness`.
- `builder-skill` (L214), `builder-plugin` (L215) — archived builders.

**Missing retained entries (ADD)**:
- `manager-develop` — canonical name (replaces the `manager-ddd`/`manager-tdd` aliases).
- `builder-harness` — canonical name (replaces `builder-agent`).

**Post-cleanup target (5 entries)**: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `builder-harness`. The 3 meta/evaluator agents (`plan-auditor`, `sync-auditor`) and `Explore` are intentionally NOT in the model map (they use `model: inherit` per model-policy.md § Inherit-by-Default — see §B.4 decision).

### §B.2 D2 — `agentEffortMap` stale (6 entries, 3 correct, 3 phantom)

File: `internal/template/model_policy.go:72-79`.

**Correct entries (keep if map is retained)**: `manager-spec` (xhigh, L73), `plan-auditor` (high, L75), `sync-auditor` (high, L76).

**Archived phantom entries (REMOVE)**: `manager-strategy` (xhigh, L74), `expert-security` (high, L77), `expert-refactoring` (high, L78).

**Missing retained (ADD if map is retained)**: `manager-develop`, `builder-harness` — but see §B.4 redundancy decision first.

### §B.3 D3 — CC 2.1.175 cost lever absent

Verified: `grep -rn "availableModels\|enforceAvailableModels" internal/template/templates/ internal/config/` returns 0 matches. The `settings.json.tmpl` currently has no Default-model constraint field. The research doc identifies these fields as the `[1m]`-safe lever (they constrain Default-model resolution at the settings level, not per-agent).

### §B.4 D4 + Effort-Map Redundancy Decision (research at M1)

- **D4**: `quality.yaml` L1-2 has `development_mode: tdd` + `enforce_quality: true` as flat globals. The harness `skip_phases` array (`harness.yaml:37-45`) encodes phase-skipping per level but does NOT feed `development_mode`. The cycle_type axis gap is that harness depth does not influence cycle_type.
- **Effort-map redundancy**: modern agents declare `effort:` directly in YAML frontmatter (e.g., `manager-spec.md` has `effort:` injected by `ApplyEffortPolicy` OR authored manually). The redundancy question: does any agent rely on `ApplyEffortPolicy` injection, or do all retained agents have hand-authored `effort:`? If the latter, the map is dead code. M1 research resolves this.

### §B.5 `[1m]` re-verification target (M1)

`model-policy.md` L30-50 cites `#45847`, `#51060`, `#36670` as the constraint basis, `last_analyzed=2.1.163` (research doc). The re-verification fetches the current state of these issues + the CC 2.1.178 CHANGELOG to determine whether the `[1m]` entitlement inheritance behavior changed. Verdict recorded in M1 research notes.

### §B.6 Uncommitted working-tree files (OUT OF SCOPE — AG-03)

The working tree has 10 modified files (internal/config/defaults.go, internal/template/deployer.go, .claude/settings.json, etc.) belonging to an unrelated parallel workstream. This SPEC MUST NOT absorb them. The plan-phase touches ONLY `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/`.

### §B.7 SPEC ID canonicalization trace

Orchestrator-proposed ID: `SPEC-CC2178-MODEL-POLICY-REPAIR` (no numeric suffix).

```
decomposition (proposed): SPEC ✓ | CC2178 ✓ | MODEL ✓ | POLICY ✓ | REPAIR ✓ | (no \d{3}) → FAIL
decomposition (corrected): SPEC ✓ | CC2178 ✓ | MODEL ✓ | POLICY ✓ | REPAIR ✓ | 001 ✓ → PASS
```

Canonicalized to `SPEC-CC2178-MODEL-POLICY-REPAIR-001` (appended `-001`). Regex: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`. All domain segments valid (`CC2178` matches `[A-Z][A-Z0-9]*`; `MODEL`, `POLICY`, `REPAIR` likewise); only the missing `\d{3}` suffix failed. See spec.md §H.

## §C. Pre-flight Checks

- [x] SPEC ID uniqueness verified (`grep -rl "SPEC-CC2178-MODEL-POLICY-REPAIR" .moai/specs/` = 0 matches pre-authoring).
- [x] 8-agent retained catalog verified (CLAUDE.md §4).
- [x] D1-D4 verified against current tree (2026-06-16).
- [x] `moai spec lint` subcommand confirmed available.
- [x] research doc exists (`.moai/research/cc-update-2.1.163-to-2.1.178.md`, 19871 bytes).
- [ ] plan-auditor gate (run after plan-phase commit).

## §D. Constraints (from spec.md §E)

1. Template-First Rule.
2. `[1m]` entitlement boundary (no per-agent pins).
3. Backward compatibility (existing `tdd` projects).
4. 8-agent catalog alignment.
5. Scope discipline (10 uncommitted files out-of-scope).
6. Language policy (`documentation: ko`; REQ tokens English).
7. Mirror parity (`model-policy.md` + template mirror edited together).

## §E. Self-Verification (§E.1 Plan-Phase Audit-Ready Signal)

### §E.1 Plan-phase Audit-Ready Signal

- **Artifact set**: spec.md (12-field frontmatter, `era: V3R6`), plan.md (this file), acceptance.md (13 ACs).
- **SPEC ID regex**: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` → PASS (decomposition in §B.7).
- **Frontmatter schema**: all 12 canonical fields present; `status: draft`; `era: V3R6` explicit (H-override suppresses `EraAutoDetected` INFO).
- **GEARS compliance**: REQs use Ubiquitous / Event-driven (When) / State-driven (While) / Capability gate (Where) / Unwanted (shall not) — 0 residual `IF/THEN`.
- **Exclusions**: EX-01..EX-06 present (6 entries, exceeds the 1-entry minimum).
- **Anti-goals**: AG-01..AG-03 present.
- **Spec lint**: `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001` — run pre-commit (AC-MPR-013).

### §E.2 Run-phase Evidence

_<pending run-phase>_

### §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

### §E.5 Mx-phase Audit-Ready Signal

_<pending mx-phase>_

## §F. Milestones

### §F.1 M1 — Research: `[1m]` re-verification + effort-map redundancy + task-triage decision

**cycle_type**: N/A (research milestone, no code).

**Tasks**:
1. Fetch the current state of Claude Code issues `#45847`, `#51060`, `#36670` (upstream GitHub).
2. Fetch the CC 2.1.178 CHANGELOG entries for `availableModels` / `enforceAvailableModels` / `[1m]` entitlement.
3. Record the `[1m]` verdict: still-active / relaxed / partially-relaxed (REQ-MPR-013, AC-MPR-011).
4. Grep all retained agent files (`.claude/agents/{core,meta,builder}/*.md`) for hand-authored `effort:` fields. If ALL retained agents with effort needs have hand-authored `effort:`, the `agentEffortMap` is redundant (REQ-MPR-011, AC-MPR-010).
5. Decide task-triage in-scope vs. deferred (REQ-MPR-016, AC-MPR-012). Recommendation: **deferred** (the 3-axis alignment is already substantial; triage is a follow-up).
6. Record decisions in a research note (inline in plan.md §B or a sibling `research.md` if Tier M scope warrants).

**AC bindings**: AC-MPR-010 (effort-map redundancy), AC-MPR-011 (`[1m]` verdict), AC-MPR-012 (triage decision).

**File targets**: research note (inline or `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/research.md`).

### §F.2 M2 — Phantom-map cleanup (D1, D2)

**cycle_type**: `tdd` (Go code change in `internal/template/`).

**RED first**: write characterization tests asserting (a) the 15 phantom keys are absent from `agentModelMap`, (b) `GetAgentModel("manager-develop")` and `GetAgentModel("builder-harness")` return non-empty, (c) the 3 phantom keys are absent from `agentEffortMap`.

**GREEN**: edit `internal/template/model_policy.go:193-216` — remove the 13 phantom keys + 2 legacy aliases, add `manager-develop` and `builder-harness` with appropriate model tuples (derive from the retired `manager-ddd`/`manager-tdd` = `{opus, sonnet, sonnet}` and `builder-agent` = `{opus, sonnet, haiku}`). Edit L72-79 to remove the 3 phantom keys from `agentEffortMap` (or retire the map entirely if M1 decided redundant).

**REFACTOR**: update the docstring at L66-71 and L191-192 to reflect the retained catalog.

**File:line targets**:
- `internal/template/model_policy.go:193-216` (agentModelMap).
- `internal/template/model_policy.go:72-79` (agentEffortMap — or full removal per M1).
- `internal/template/model_policy_test.go` (characterization + new coverage).

**AC bindings**: AC-MPR-007, AC-MPR-008, AC-MPR-009.

### §F.3 M3 — Cycle_type harness routing (D4)

**cycle_type**: `tdd` (config + Go integration).

> **Split Trigger**: if M3 scope grows beyond (a) reading harness level in the quality-gate path, (b) mapping to cycle_type, and (c) preserving explicit `quality.yaml` pins — split into `SPEC-CC2178-CYCLE-TYPE-ROUTING-001` and reduce THIS SPEC to the model + cleanup axes. The trigger fires if the integration requires touching >3 files outside `quality.yaml` + the Complexity Estimator.

**RED first**: write a test that asserts (a) harness `minimal` resolves to a non-tdd cycle_type, (b) harness `thorough` resolves to `tdd`, (c) an explicit `quality.yaml development_mode: tdd` pin is preserved even when harness is `minimal` (backward-compat, AG-01).

**GREEN**: implement the routing. The lightest approach: add a `cycle_type_resolution` field/section to `quality.yaml` that maps harness level → cycle_type, with an explicit-pin override. The harness Complexity Estimator's resolved level feeds this map. Keep `enforce_quality: true` semantics unchanged.

**File:line targets**:
- `.moai/config/sections/quality.yaml` (add cycle_type routing section).
- `internal/config/` (the config struct that reads `quality.yaml` — exact file TBD at M3 entry; likely `defaults.go` or a quality-config reader).
- `internal/harness/` or `internal/cli/` (the Complexity Estimator integration point — exact file TBD).

**AC bindings**: AC-MPR-004, AC-MPR-005, AC-MPR-006.

### §F.4 M4 — Default-model cost lever (D3)

**cycle_type**: `tdd` (template + Go render test).

**RED first**: write a render test that asserts the deployed `settings.json` contains `availableModels` with the 3 aliases, `enforceAvailableModels: true`, and Default model `sonnet`.

**GREEN**: edit `internal/template/templates/.claude/settings.json.tmpl` — add the fields. Run `make build` to regenerate `embedded.go`. The Default-model field is the top-level `model:` key (or the CC 2.1.175 Default-resolution mechanism — verify at M1 which key CC 2.1.178 uses).

**Template-First**: edit `templates/` FIRST, then `make build`. Do NOT edit `embedded.go` directly.

**File:line targets**:
- `internal/template/templates/.claude/settings.json.tmpl` (add fields).
- `internal/template/embedded.go` (regenerated by `make build` — DO NOT edit).
- `internal/template/settings_test.go` (render test).

**AC bindings**: AC-MPR-001, AC-MPR-002, AC-MPR-003.

### §F.5 M5 — model-policy.md doctrine update (+ mirror parity)

**cycle_type**: N/A (documentation milestone).

**Tasks**:
1. Edit `.claude/rules/moai/development/model-policy.md` — add the `[1m]` re-verification verdict from M1, document the Default-model lever (REQ-MPR-001..003), note that per-agent pins remain forbidden (EX-01) regardless of verdict (REQ-MPR-014).
2. Apply the mirror edit to `internal/template/templates/.claude/rules/moai/development/model-policy.md` (byte-parity per internal/template/CLAUDE.md § Mirror parity checks).
3. Run `internal/template/embedded_mirror_test.go` to verify byte-identity.

**File:line targets**:
- `.claude/rules/moai/development/model-policy.md` (L30-50 Inherit-by-Default section + new Default-model section).
- `internal/template/templates/.claude/rules/moai/development/model-policy.md` (mirror).

**AC bindings**: AC-MPR-011 (verdict documented), AC-MPR-013 (mirror parity).

### §F.6 M6 — Verification + lint + commit

**cycle_type**: N/A (verification milestone).

**Tasks**:
1. `go test ./internal/template/...` — confirm M2/M4 tests pass.
2. `go test ./internal/config/... ./internal/harness/...` — confirm M3 routing tests pass.
3. `golangci-lint run` — 0 errors.
4. `make build` — regenerate `embedded.go` cleanly.
5. `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001` — exit 0 (AC-MPR-013).
6. `internal/template/internal_content_leak_test.go` + `TestTemplateNeutralityAudit` — confirm no internal-content leak (the `settings.json.tmpl` change references only CC model aliases, no SPEC IDs).
7. Trust-but-verify: independently re-run the D1-D4 greps to confirm the defects are resolved.

**AC bindings**: all ACs.

## §G. Anti-Patterns (to avoid)

- **AP-MPR-001**: Expanding scope into per-agent pins when the `[1m]` re-verification returns "partially-relaxed". Mitigation: REQ-MPR-015 routes relaxation to a follow-up SPEC.
- **AP-MPR-002**: Breaking backward-compat by overriding an explicit `quality.yaml development_mode: tdd` pin with a harness-derived cycle_type. Mitigation: REQ-MPR-006 + AC-MPR-006.
- **AP-MPR-003**: Editing `embedded.go` directly instead of `templates/` + `make build`. Mitigation: M4 Template-First discipline.
- **AP-MPR-004**: Absorbing the 10 uncommitted working-tree files into this SPEC's commits. Mitigation: AG-03 scope discipline; commit ONLY the SPEC directory + the explicitly-targeted source files.
- **AP-MPR-005**: Retiring `agentEffortMap` without grepping for non-test callers of `GetAgentEffort`/`ApplyEffortPolicy`. Mitigation: M1 redundancy decision requires caller evidence.
- **AP-MPR-006**: Using `IF/THEN` in new REQs (GEARS lint warning). Mitigation: all REQs use `When`/`While`/`Where`/`shall not`.

## §H. Cross-References

- `internal/template/model_policy.go:72-79,193-216` — D1/D2 defect sites.
- `internal/template/templates/.claude/settings.json.tmpl` — D3 target.
- `.moai/config/sections/quality.yaml:1-2` — D4 defect site.
- `.moai/config/sections/harness.yaml:32-75` — harness levels + skip_phases (M3 integration point).
- `.claude/rules/moai/development/model-policy.md:30-50` — `[1m]` Inherit-by-Default doctrine (M5 update target).
- `.moai/research/cc-update-2.1.163-to-2.1.178.md` — research doc (M1 input).
- CLAUDE.md §4 — 8-agent retained catalog (D1/D2 reconciliation source).
- SPEC-V3R6-AGENT-TEAM-REBUILD-001 — archived-agent consolidation precedent.

## §I. Implementation Kickoff Approval (plan-to-implement human gate)

Per CLAUDE.local.md §19.1, this SPEC requires explicit user approval before run-phase entry (the plan-to-implement HUMAN GATE is NOT skip-eligible regardless of plan-auditor score). After plan-auditor returns its verdict and this plan is marked audit-ready, the orchestrator MUST:

1. Present the plan-auditor verdict + this plan summary to the user via `AskUserQuestion`.
2. Offer 3 options: (a) proceed to `/moai run` (Recommended), (b) request plan revision, (c) abort.
3. On approval, delegate to `manager-develop` with `cycle_type=tdd` for M2-M4 and orchestrator-direct for M5-M6.
