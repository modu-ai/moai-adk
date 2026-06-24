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

### §B.1 D1 — `agentModelMap` stale (19 entries, 3 correct, 16 phantom, 2 missing) — iter-2 corrected count

File: `internal/template/model_policy.go:193-216`.

> **iter-2 count correction**: the iter-1 SPEC and the plan-auditor iter-1 report both said "18 entries". The authoritative count obtained by `awk '/var agentModelMap = map/,/^\}/' internal/template/model_policy.go | grep -cE '^\s+"'` on 2026-06-16 is **19 entries**. The discrepancy: `builder-plugin` IS present in the map (the iter-1 SPEC erroneously treated it as apocryphal in one location and real in another — this iter-2 unifies to "present, must be removed").

**Retained-catalog entries (CORRECT, keep)**: `manager-spec` (L195), `manager-docs` (L198), `manager-git` (L202) — 3 entries.

**Canonical 16 keys to REMOVE** (cross-referenced from spec.md §C.3 canonical table — single source of truth):
- Legacy aliases (3): `manager-ddd` (L196), `manager-tdd` (L197), `builder-agent` (L213).
- Archived core managers (3): `manager-quality` (L199), `manager-project` (L200), `manager-strategy` (L201).
- Archived `expert-*` (8): `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-debug`, `expert-testing`, `expert-refactoring` (L204-211).
- Archived builder siblings (2): `builder-skill` (L214), `builder-plugin` (L215).

**Missing retained entries (ADD)**:
- `manager-develop` — canonical name (replaces the `manager-ddd`/`manager-tdd` aliases); tuple `{sonnet, sonnet, haiku}` per REQ-MPR-009 iter-2 decision (NOT the iter-1 `{opus, sonnet, sonnet}` derived from retired aliases — that would pin the busiest agent to Opus, contradicting the Default-Sonnet thesis; see §F.2 D6 rationale).
- `builder-harness` — canonical name (replaces `builder-agent`); tuple `{sonnet, sonnet, haiku}` per REQ-MPR-009 iter-2 decision.

**Post-cleanup target (5 entries)**: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `builder-harness`. The 3 meta/evaluator agents (`plan-auditor`, `sync-auditor`) and `Explore` are intentionally NOT in the model map (they use `model: inherit` per model-policy.md § Inherit-by-Default — see §B.4 decision).

### §B.2 D2 — `agentEffortMap` stale + map↔file divergence (iter-2 reconciled)

File: `internal/template/model_policy.go:72-79`.

> **iter-2 D5 reconciliation**: the iter-1 plan treated the effort map as a pure "redundancy" question. Ground-truth verification (2026-06-16) shows `ApplyEffortPolicy` has **2 production callers** (`initializer.go:181`, `update.go:2661`) whose job is to INJECT `effort:` into freshly-deployed agents. The map is NOT redundant — retiring it causes a deployment regression. The iter-2 scope is therefore PRUNE + RECONCILE, not retire.

**Current entries (6, verified L72-79)**: `manager-spec` (xhigh), `manager-strategy` (xhigh), `plan-auditor` (high), `sync-auditor` (high), `expert-security` (high), `expert-refactoring` (high).

**Archived phantom entries (REMOVE — REQ-MPR-010/011a)**: `manager-strategy`, `expert-security`, `expert-refactoring`.

**Map↔file divergence (RECONCILE — REQ-MPR-011b)** — verified by `grep '^effort:' .claude/agents/moai/*.md` on 2026-06-16 (the `meta/` and `builder/` subdirectories DO NOT EXIST — agents were re-consolidated to `.claude/agents/moai/` post-AGENT-FOLDER-SPLIT revert; iter-3 D12 path correction):
- Map says `plan-auditor: high`; file `.claude/agents/moai/plan-auditor.md` says `effort: xhigh`. → sync map to `xhigh`.
- Map says `sync-auditor: high`; file `.claude/agents/moai/sync-auditor.md` says `effort: xhigh`. → sync map to `xhigh`.
- `manager-develop` is MISSING from the map; file `.claude/agents/moai/manager-develop.md` says `effort: xhigh`. → add to map as `xhigh`.
- `builder-harness` is MISSING from the map; file `.claude/agents/moai/builder-harness.md` says `effort: high`. → add to map as `high`.

**Post-reconciliation target (5 entries)**: `manager-spec` (xhigh), `plan-auditor` (xhigh), `sync-auditor` (xhigh), `manager-develop` (xhigh), `builder-harness` (high).

### §B.3 D3 — CC 2.1.175 cost lever absent

Verified: `grep -rn "availableModels\|enforceAvailableModels" internal/template/templates/ internal/config/` returns 0 matches. The `settings.json.tmpl` currently has no Default-model constraint field. The research doc identifies these fields as the `[1m]`-safe lever (they constrain Default-model resolution at the settings level, not per-agent).

### §B.4 D4 + Effort-Map Decision + Dead-Walker Correction (iter-3 re-grounded)

- **D4 (cycle_type)**: `quality.yaml` has `development_mode: tdd` + `enforce_quality: true` **nested under the top-level `constitution:` key** (verified 2026-06-16 — NOT flat globals; the iter-1 "flat globals" wording was imprecise per the plan-auditor prose-precision note). The harness Complexity Estimator (`internal/harness/router/router.go:104-108`) resolves a harness `Level` from SPEC frontmatter `harness_level` but does NOT emit a cycle_type. The cycle_type axis gap is that harness depth does not influence cycle_type. **There is currently no symbol in `internal/` that maps harness level → cycle_type** (verified: `grep -rn "resolveCycleType\|ResolveCycleType" internal/` returns 0 matches). The M3 milestone (§F.3) defines and authors this NEW symbol.
- **Effort-map decision (iter-3 re-grounded per D11)**: `ApplyEffortPolicy` (`model_policy.go:134-180`) HAS **2 production callers** (`initializer.go:181`, `update.go:2661`) AND `ApplyModelPolicy` (`model_policy.go:240+`) HAS **2 production callers** (`initializer.go:176`, `update.go:2656`). The iter-2 premise — "retiring the map loses live effort injection = regression" — was re-grounded in iter-3: the walker is CURRENTLY A NO-OP. Both functions hardcode `domains := []string{"core", "expert", "meta", "harness"}` (L137, L246), but those subdirectories DO NOT EXIST on disk (verified: `find .claude/agents -type d` returns only `moai/` and `local/`). The walker reads 4 non-existent dirs, `os.IsNotExist` → `continue` silently for each, and injects NOTHING today. Therefore REQ-MPR-008/009 (model-map cleanup) + REQ-MPR-011a/011b (effort-map prune + reconcile) currently have zero production effect. **iter-3 ABSORBS the walker-target fix as REQ-MPR-019 (MUST)** — correcting `domains` to `[]string{"moai"}` at both L137 and L246 restores live injection so the map cleanup actually takes effect. Full retirement of `agentEffortMap` + `ApplyEffortPolicy` (migrating the 2 callers to an alternative path) remains DEFERRED to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001`.

### §B.5 `[1m]` re-verification target (M1)

`model-policy.md` L30-50 cites `#45847`, `#51060`, `#36670` as the constraint basis, `last_analyzed=2.1.163` (research doc). The re-verification fetches the current state of these issues + the CC 2.1.178 CHANGELOG to determine whether the `[1m]` entitlement inheritance behavior changed. Verdict recorded in M1 research notes.

### §B.6 Uncommitted working-tree files (OUT OF SCOPE — AG-03) + SPEC target-file enumeration (iter-3 updated)

**Unrelated parallel-workstream files (DO NOT ABSORB)**: the working tree currently carries a parallel workstream (settings-management, hooks, sync-phase-quality-gate, llm.yaml, etc. — verified by `git status --porcelain` on 2026-06-16 iter-3). The exact count shifts across iterations as that workstream progresses; the rule is unchanged from AG-03: this SPEC MUST NOT absorb any of those files. The plan-phase touches ONLY `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/`.

**SPEC OWN run-phase target files (iter-3 enumeration — M2/M3/M4/M5)**:
- **M2 modified**: `internal/template/model_policy.go` (agentModelMap L193-216 + agentEffortMap L72-79 + **walker domains L137, L246 — iter-3 REQ-MPR-019**), `internal/template/model_policy_test.go` (characterization + reconciliation coverage).
- **M2 NEW untracked (iter-3 added)**: `internal/template/model_policy_walker_test.go` — dedicated walker-target test asserting `domains` contains `"moai"` and the walker reads `.claude/agents/moai/` (AC-MPR-015 verification).
- **M3 NEW untracked**: `internal/config/cycle_type.go`, `internal/config/cycle_type_test.go` (the `ResolveCycleType` symbol — D1 remediation).
- **M3 modified**: `.moai/config/sections/quality.yaml` (documentation section), `internal/harness/router/router.go` (1 call-site wire near L104-108).
- **M4 modified**: `internal/template/templates/.claude/settings.json.tmpl` (D3 fields), `internal/template/settings_test.go` (render test), `internal/template/embedded.go` (regenerated by `make build` — DO NOT hand-edit).
- **M5 modified**: `.claude/rules/moai/development/model-policy.md` (+ template mirror — `[1m]` verdict + Default-model doctrine).

**iter-3 delta vs iter-2**: +1 NEW untracked file (`model_policy_walker_test.go`) for REQ-MPR-019 walker-target verification. The `model_policy.go` modification was already in iter-2's M2 scope; iter-3 adds 2 more modified lines (L137, L246 `domains` slice) to that same file.

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
5. Scope discipline (unrelated parallel-workstream working-tree files out-of-scope per AG-03; see §B.6 for the current enumeration — the count shifts across iterations as the parallel workstream progresses).
6. Language policy (`documentation: ko`; REQ tokens English).
7. Mirror parity (`model-policy.md` + template mirror edited together).

## §E. Self-Verification (§E.1 Plan-Phase Audit-Ready Signal)

### §E.1 Plan-phase Audit-Ready Signal

- **Artifact set**: spec.md (12-field frontmatter, `era: V3R6`, version 0.3.0 iter-3), plan.md (this file), acceptance.md (16 ACs — iter-3 added AC-MPR-015 walker-target + AC-MPR-016 REQ-MPR-018 binding).
- **SPEC ID regex**: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` → PASS (decomposition in §B.7).
- **Frontmatter schema**: all 12 canonical fields present; `status: draft`; `era: V3R6` explicit (H-override suppresses `EraAutoDetected` INFO).
- **GEARS compliance**: REQs use Ubiquitous / Event-driven (When) / State-driven (While) / Capability gate (Where) / Unwanted (shall not) — 0 residual `IF/THEN`.
- **Exclusions**: EX-01..EX-06 present (6 entries, exceeds the 1-entry minimum).
- **Anti-goals**: AG-01..AG-03 present.
- **Spec lint**: `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` — path-prefixed form (D3). Observed output at iter-2 authoring (2026-06-16): `0 error(s), 1 warning(s)` (WARNING `StatusGitConsistency` — frontmatter `status: draft` disagrees with git-implied `implemented`; expected for a plan-phase draft not yet committed with the canonical `feat(SPEC-...): plan-phase artifacts` subject).

### §E.2 Run-phase Evidence

_<pending run-phase>_

### §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

### §E.5 Mx-phase Audit-Ready Signal

_<pending mx-phase>_

## §F. Milestones

### §F.1 M1 — Research: `[1m]` re-verification + Default-model key confirmation + task-triage decision

**cycle_type**: N/A (research milestone, no code).

> **iter-2 change**: the iter-1 M1 included "effort-map redundancy decision" as an open research question. Per D5 remediation, the effort-map decision is now RESOLVED at plan-phase (premise was factually inverted — see §B.4). M1 no longer researches "retire vs. keep"; it only records the deferral rationale (REQ-MPR-012 downgraded). M1 gains a NEW task: confirm the CC 2.1.178 Default-resolution JSON key (D4 remediation for AC-MPR-003).

**Tasks**:
1. Fetch the current state of Claude Code issues `#45847`, `#51060`, `#36670` (upstream GitHub).
2. Fetch the CC 2.1.178 CHANGELOG entries for `availableModels` / `enforceAvailableModels` / `[1m]` entitlement.
3. Record the `[1m]` verdict: still-active / relaxed / partially-relaxed (REQ-MPR-013, AC-MPR-011).
4. **NEW (D4)**: confirm the exact CC 2.1.178 Default-resolution JSON key for the `settings.json` `Default = sonnet` setting. Post-CC-2.1.175 the settings file carries `model`, `availableModels`, and possibly other model-named keys; a naive `grep '"model"'` matches multiple lines. Record the confirmed key + the exact grep command that AC-MPR-003 will use. If M1 cannot confirm the key (e.g., the CC 2.1.178 runtime is not available locally), record the fallback verification approach (render + manual JSON inspection).
5. Record the effort-map deferral rationale (REQ-MPR-012 downgraded): the 2 production callers, the regression risk, the follow-up SPEC name. This is a documentation task — the decision is already made at plan-phase.
6. Decide task-triage in-scope vs. deferred (REQ-MPR-016, AC-MPR-012). Recommendation: **deferred** (the 3-axis alignment is already substantial; triage is a follow-up).
7. Record decisions in a research note (inline in plan.md §B or a sibling `research.md` if Tier M scope warrants).

**AC bindings**: AC-MPR-010 (effort-map prune/reconcile applied — no longer "redundancy decision"), AC-MPR-011 (`[1m]` verdict), AC-MPR-012 (triage decision), AC-MPR-003 (Default-model key confirmed).

**File targets**: research note (inline or `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/research.md`).

### §F.2 M2 — Phantom-map cleanup (D1) + effort-map prune-and-reconcile (D2) + dead-walker fix (D11)

**cycle_type**: `tdd` (Go code change in `internal/template/`).

> **iter-3 addition (D11 / REQ-MPR-019)**: M2 now ALSO fixes the dead domain-walker in `model_policy.go`. Both `ApplyEffortPolicy` (L137) and `ApplyModelPolicy` (L246) currently hardcode `domains := []string{"core", "expert", "meta", "harness"}` — subdirectories that do not exist on disk (agents live in `.claude/agents/moai/`). The walker is a no-op today, defeating the map cleanup. M2 corrects `domains` to `[]string{"moai"}` at BOTH sites so the map cleanup takes live effect.

**RED first**: write characterization tests asserting (a) the **16** canonical phantom keys are absent from `agentModelMap` (see spec.md §C.3 table — iter-2 corrected from the iter-1 "15"), (b) `GetAgentModel("manager-develop")` and `GetAgentModel("builder-harness")` return `{sonnet, sonnet, haiku}` (iter-2 tuple decision per D6), (c) the 3 phantom keys are absent from `agentEffortMap`, (d) `agentEffortMap` values for `plan-auditor`/`sync-auditor` are `xhigh` (map↔file reconciliation per REQ-MPR-011b), **(e) iter-3 NEW**: the `domains` slice in both `ApplyEffortPolicy` and `ApplyModelPolicy` contains `"moai"` and does NOT contain any of `"core"`/`"expert"`/`"meta"`/`"harness"` (AC-MPR-015).

**GREEN**:
- `agentModelMap` (L193-216): remove the 16 canonical phantom keys; add `manager-develop: {sonnet, sonnet, haiku}` and `builder-harness: {sonnet, sonnet, haiku}` (iter-2 tuples — NOT the iter-1 `{opus, sonnet, sonnet}` derived from retired aliases; D6 rationale: the busiest run-phase agents follow Default-Sonnet, not Opus).
- `agentEffortMap` (L72-79): remove the 3 phantom keys (`manager-strategy`, `expert-security`, `expert-refactoring`); sync `plan-auditor` and `sync-auditor` from `high` to `xhigh`; add `manager-develop: xhigh` and `builder-harness: high`. Do NOT retire the map (D5 inversion — `ApplyEffortPolicy` has 2 production callers; retirement is deferred).
- **Walker fix (iter-3 REQ-MPR-019)**: change `domains := []string{"core", "expert", "meta", "harness"}` → `domains := []string{"moai"}` at BOTH L137 (ApplyEffortPolicy) AND L246 (ApplyModelPolicy). Both share the same bug; both MUST be fixed together so the map cleanup reaches the real `.claude/agents/moai/` layout.

**REFACTOR**: update the docstrings at L66-71 and L191-192 to reflect the retained 5-entry catalog; update the docstrings/comments near L137 and L246 to document that the walker targets `.claude/agents/moai/` (the consolidated layout post-AGENT-FOLDER-SPLIT revert).

**File:line targets**:
- `internal/template/model_policy.go:193-216` (agentModelMap).
- `internal/template/model_policy.go:72-79` (agentEffortMap — prune + reconcile only, NO retirement).
- `internal/template/model_policy.go:137` (ApplyEffortPolicy `domains` — iter-3 REQ-MPR-019).
- `internal/template/model_policy.go:246` (ApplyModelPolicy `domains` — iter-3 REQ-MPR-019).
- `internal/template/model_policy_test.go` (characterization + reconciliation coverage).
- `internal/template/model_policy_walker_test.go` (NEW iter-3 — dedicated walker-target test for AC-MPR-015).

**AC bindings**: AC-MPR-007, AC-MPR-008, AC-MPR-009, AC-MPR-010, **AC-MPR-015 (NEW iter-3)**.

### §F.3 M3 — Cycle_type harness routing (D4) + `ResolveCycleType` function contract (D1 remediation)

**cycle_type**: `tdd` (config + Go integration).

> **Split Trigger**: if M3 scope grows beyond (a) reading harness level in the quality-gate path, (b) mapping to cycle_type via the `ResolveCycleType` symbol, and (c) preserving explicit `quality.yaml` pins — split into `SPEC-CC2178-CYCLE-TYPE-ROUTING-001` and reduce THIS SPEC to the model + cleanup axes. The trigger fires if the integration requires touching >3 files outside `quality.yaml` + the harness Complexity Estimator at `internal/harness/router/router.go` + the new `ResolveCycleType` resolver. (iter-2 names the estimator file explicitly per D9 — the iter-1 "the Complexity Estimator" phrasing left the denominator ambiguous.)

#### §F.3.1 `ResolveCycleType` function contract (NEW code authored at M3 — D1 remediation)

The iter-1 acceptance ACs (AC-MPR-004/005/006) asserted a symbol `resolveCycleType` that does NOT exist anywhere in `internal/` (verified 2026-06-16). This M3 milestone AUTHORS that symbol. The contract below is the authoritative definition; the acceptance ACs reference this contract.

- **Package**: `internal/config` (co-located with the existing `MOAI_DEVELOPMENT_MODE` reader at `internal/config/manager.go:394` which reads env-var overrides into `cfg.Quality.DevelopmentMode`). Placing the resolver in `internal/config` keeps the cycle_type decision adjacent to the config struct it reads.
- **Symbol name**: `ResolveCycleType` (exported, PascalCase — the acceptance ACs use this exact token).
- **Signature**:
  ```go
  // ResolveCycleType determines the run-phase cycle_type (ddd | tdd | autofix)
  // from the harness level, honoring an explicit quality.yaml constitution.development_mode pin.
  //
  // Precedence (highest first):
  //   1. explicitPinvPin (non-empty) — AG-01 backward-compat
  //   2. harnessLevel dispatch table
  //   3. fallback "tdd" (current global default)
  func ResolveCycleType(harnessLevel string, explicitPin string) string
  ```
- **Return type**: `string` — one of `"tdd"`, `"ddd"`, `"autofix"`. (String-typed, not a custom enum, to match the existing `DevelopmentMode` type in `internal/models/` which is already `type DevelopmentMode string`.)
- **Dispatch table** (the authoritative harness-level → cycle_type mapping):
  | `harnessLevel` | Resolved cycle_type | Rationale |
  |----------------|---------------------|-----------|
  | `"minimal"` | `"ddd"` | Lightweight: skip_phases-driven DDD-lite for simple changes (REQ-MPR-004). `minimal` does NOT mean "no methodology" — it means characterization-test-first DDD without the full TDD RED-GREEN overhead. |
  | `"standard"` | `"tdd"` | Default unchanged — the current global behavior (REQ-MPR-007, AC-MPR-014). |
  | `"thorough"` | `"tdd"` | Critical features retain full TDD discipline (REQ-MPR-005). |
  | `""` / unknown | `"tdd"` | Safe fallback — never returns empty; unknown levels get the current default. |
- **Explicit-pin override (AG-01 backward-compat)**: when `explicitPin` is non-empty (i.e., `quality.yaml constitution.development_mode` is set to `tdd` or `ddd`), `ResolveCycleType` returns `explicitPin` verbatim regardless of `harnessLevel`. This preserves existing pinned projects.

**RED first**: write table-driven tests in `internal/config/cycle_type_test.go` (NEW file) asserting:
- `ResolveCycleType("minimal", "") == "ddd"` (AC-MPR-004).
- `ResolveCycleType("thorough", "") == "tdd"` (AC-MPR-005).
- `ResolveCycleType("standard", "") == "tdd"` (AC-MPR-014 — NEW in iter-2, D7).
- `ResolveCycleType("minimal", "tdd") == "tdd"` (AC-MPR-006 — explicit pin wins over harness level).
- `ResolveCycleType("", "") == "tdd"` (unknown-level fallback).

**GREEN**: author `ResolveCycleType` in `internal/config/cycle_type.go` (NEW file) per the contract above. Wire the harness router (`internal/harness/router/router.go`) to call `ResolveCycleType(resolvedLevel, cfg.Quality.DevelopmentMode)` in the run-phase entry path (exact call site determined at M3 entry; the router currently resolves `Level` at L104-108 but does not emit a cycle_type).

**File:line targets**:
- `internal/config/cycle_type.go` (NEW — the resolver).
- `internal/config/cycle_type_test.go` (NEW — the RED tests).
- `.moai/config/sections/quality.yaml` (add a `cycle_type_routing` section documenting the dispatch table for human readers; the Go resolver is the SSOT, this section is documentation).
- `internal/harness/router/router.go` (wire the call — 1 call site near L104-108).

**AC bindings**: AC-MPR-004, AC-MPR-005, AC-MPR-006, AC-MPR-014 (NEW).

### §F.4 M4 — Default-model cost lever (D3)

**cycle_type**: `tdd` (template + Go render test).

**RED first**: write a render test that asserts the deployed `settings.json` contains `availableModels` with the 3 aliases, `enforceAvailableModels: true`, and Default model `sonnet` at the key confirmed by M1 (AC-MPR-003 conditional verification — D4).

**GREEN**: edit `internal/template/templates/.claude/settings.json.tmpl` — add the fields. Run `make build` to regenerate `embedded.go`. The Default-model field uses the key confirmed at M1 task 4 (currently NO top-level `model:` key exists in the template — verified 2026-06-16). M1 determines whether CC 2.1.178 reads `model:`, `defaultModel:`, or relies solely on `availableModels` + runtime Default resolution.

**Template-First**: edit `templates/` FIRST, then `make build`. Do NOT edit `embedded.go` directly.

**File:line targets**:
- `internal/template/templates/.claude/settings.json.tmpl` (add fields).
- `internal/template/embedded.go` (regenerated by `make build` — DO NOT edit).
- `internal/template/settings_test.go` (render test).

**AC bindings**: AC-MPR-001, AC-MPR-002, AC-MPR-003 (conditional — verification command pinned at M1).

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
5. `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` — exit 0 (AC-MPR-013; path-prefixed form per D3 — the bare-ID form ParseFails).
6. `internal/template/internal_content_leak_test.go` + `TestTemplateNeutralityAudit` — confirm no internal-content leak (the `settings.json.tmpl` change references only CC model aliases, no SPEC IDs).
7. Trust-but-verify: independently re-run the D1-D4 greps to confirm the defects are resolved.

**AC bindings**: all ACs.

## §F.7 D6 Design Rationale — `manager-develop` model tuple decision

The iter-1 plan (§F.2) derived `manager-develop` from retired `manager-ddd`/`manager-tdd` = `{opus, sonnet, sonnet}` and `builder-harness` from retired `builder-agent` = `{opus, sonnet, haiku}`. The plan-auditor iter-1 flagged this as a DESIGN TENSION with the SPEC's central thesis (route cost to Default=Sonnet): `manager-develop` is the most-frequently-spawned run-phase agent, so pinning its High policy to Opus undercuts the 3-axis cost-routing goal.

**iter-2 decision**: `manager-develop = {sonnet, sonnet, haiku}` and `builder-harness = {sonnet, sonnet, haiku}`.

**Options evaluated** (per the audit's D6 directive, cross-referenced against `.moai/research/cc-update-2.1.163-to-2.1.178.md` which found Sonnet+v2 ≈ Opus on diagnostics, +5-7% gap only on deep reasoning):

| Option | manager-develop tuple | Pros | Cons | Verdict |
|--------|----------------------|------|------|---------|
| **A** | `inherit` (no entry; inherits Default=Sonnet) | Maximum cost consistency; simplest | `GetAgentModel` returns `""` which some callers treat as "skip" — loses explicit documentation | Rejected — loses the explicit-map documentation benefit |
| **B** | `{sonnet, sonnet, haiku}` explicit | Same cost effect as A; explicit in map; `GetAgentModel` returns non-empty | Sonnet at Low policy for coding may underperform on complex run-phase work | **CHOSEN** — explicit + cost-aligned; Low-policy coding-agent spawn is rare (most run-phase work is High/Medium) |
| **C** | `{opus, sonnet, haiku}` (deep-reasoning exception) | Opus at High for complex coding | Directly contradicts the SPEC thesis; the busiest agent on Opus = no cost savings | Rejected — would make the SPEC self-defeating |

**Reconciliation with Default-model routing (REQ-MPR-001..003)**: the Default model is Sonnet (set via `availableModels`/`enforceAvailableModels`); per-agent `model:` pins are FORBIDDEN by the `[1m]` constraint (EX-01). The `agentModelMap` entries are the POLICY-layer intent (what `GetAgentModel` returns when a caller queries the desired model for a given agent × policy) — they do NOT directly set a per-agent `model:` pin in deployed settings. The map's High=Sonnet for `manager-develop` is therefore consistent with Default=Sonnet: the policy says "when you would spawn manager-develop at High cost, use Sonnet", which is the same model the Default already resolves to. No tension.

**Same scrutiny for `builder-harness`**: builder-harness spawns rarely (only for dynamic-harness generation) and its files say `effort: high` (less reasoning than `manager-develop`'s `xhigh`). `{sonnet, sonnet, haiku}` is consistent.

## §G. Anti-Patterns (to avoid)

- **AP-MPR-001**: Expanding scope into per-agent pins when the `[1m]` re-verification returns "partially-relaxed". Mitigation: REQ-MPR-015 routes relaxation to a follow-up SPEC.
- **AP-MPR-002**: Breaking backward-compat by overriding an explicit `quality.yaml development_mode: tdd` pin with a harness-derived cycle_type. Mitigation: REQ-MPR-006 + AC-MPR-006.
- **AP-MPR-003**: Editing `embedded.go` directly instead of `templates/` + `make build`. Mitigation: M4 Template-First discipline.
- **AP-MPR-004**: Absorbing the 10 uncommitted working-tree files into this SPEC's commits. Mitigation: AG-03 scope discipline; commit ONLY the SPEC directory + the explicitly-targeted source files.
- **AP-MPR-005**: Retiring `agentEffortMap` + `ApplyEffortPolicy` without migrating the 2 production callers (`initializer.go:181`, `update.go:2661`). **iter-3 correction**: the iter-1 plan framed this as "grep for callers first"; iter-2 found BOTH callers exist (DEFERRED, not pending-grep); iter-3 further found the walker is a no-op (D11), so the deferral rationale is caller-migration effort, NOT preserving live injection. This SPEC prunes phantoms + reconciles map↔file divergence + FIXES THE WALKER (REQ-MPR-019) only. Full retirement belongs to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001`.
- **AP-MPR-009** (NEW iter-3, D11): Forgetting that the walker-target fix (REQ-MPR-019) MUST land alongside the map cleanup — without the walker fix, REQ-MPR-008/009/011a/011b have zero production effect because the walker never reads `.claude/agents/moai/`. Mitigation: AC-MPR-015 verifies the `domains` slice targets `moai/` at both L137 and L246.
- **AP-MPR-006**: Using `IF/THEN` in new REQs (GEARS lint warning). Mitigation: all REQs use `When`/`While`/`Where`/`shall not`.
- **AP-MPR-007** (NEW iter-2): Asserting `resolveCycleType` (lowercase) in ACs when the agreed symbol is `ResolveCycleType` (exported, `internal/config` package) per §F.3.1 contract. Mitigation: AC-MPR-004/005/006/014 reference `ResolveCycleType` with the exact signature from §F.3.1.
- **AP-MPR-008** (NEW iter-2, D6): Deriving `manager-develop`'s model tuple from the retired `manager-ddd`/`manager-tdd` aliases (`{opus, sonnet, sonnet}`). This pins the most-frequently-spawned agent to Opus at High policy, contradicting the SPEC's Default-Sonnet thesis. Mitigation: REQ-MPR-009 iter-2 sets `{sonnet, sonnet, haiku}` with explicit rationale.

## §H. Cross-References

- `internal/template/model_policy.go:72-79,193-216` — D1/D2 defect sites.
- `internal/template/model_policy.go:137` (ApplyEffortPolicy `domains`), `:246` (ApplyModelPolicy `domains`) — **D11 dead-walker defect sites (iter-3 REQ-MPR-019)**. Both hardcode the dead `{core,expert,meta,harness}` layout; fixed to `[]string{"moai"}` at M2.
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
