---
created: 2026-05-22
updated: 2026-05-22
tags: [research, config, audit, dead-code, retracted]
status: retracted-v1-corrected-v2
authors: [Claude Opus 4.7 + manager-spec verification]
related_specs:
  - SPEC-CONFIG-001 (Configuration Management System) status:completed v1.1.0
  - SPEC-CI-MULTI-LLM-001 (Multi-LLM GitHub Actions Self-Hosted Runner) status:in-progress v1.0.0
  - SPEC-V3R2-RT-006 (Hook Handler Completeness) status:completed
  - SPEC-V3R2-MIG-002 (Hook Registration Cleanup) status:completed
---

# Config Dead Code Audit — 2026-05-22 (v2 corrected)

> **NOTICE**: This document supersedes the initial audit summary surfaced in the orchestrator response of 2026-05-22. The initial audit (v1) used grep-only symbol reference signals and **mis-classified 3 active items as dead**. Manager-spec cross-check against `.moai/specs/` SPEC catalog blocked the proposed cleanup SPEC and surfaced this defect. v2 documents the correction and the SPEC catalog cross-check procedure that must precede future dead-code audits.

---

## 1. Retraction Summary

The 2026-05-22 v1 audit proposed three items as "fully dead config":

- **A1**: `PricingConfig` struct (`internal/config/types.go:130-134` + defaults + manager)
- **A2**: `github-actions.yaml` + `.tmpl` (725B yaml + 918B tmpl)
- **A3**: `system.hook.observability_events` field (`internal/config/types.go:53-60`)

All three are **withdrawn**. Each has an active SPEC owner. Removing them without SPEC supersession would partially revert completed SPECs or strand in-progress work.

---

## 2. Item-by-Item Correction

### 2.1 A1 — PricingConfig

**v1 claim**: "yaml file `pricing:` key not found anywhere, self-references only, external consumers 0 → dead struct"

**v2 correction**: PricingConfig is **forward-compat infrastructure registered by SPEC-CONFIG-001 v1.1.0 (status: completed)**. Evidence:

| Location | Content |
|----------|---------|
| `.moai/specs/SPEC-CONFIG-001/spec.md:83` | `pricing.yaml # PricingConfig` |
| `.moai/specs/SPEC-CONFIG-001/spec.md:316` | `Pricing PricingConfig yaml:"pricing"` field declaration |
| `.moai/specs/SPEC-CONFIG-001/spec.md:435` | `type PricingConfig struct {…}` schema |
| `.moai/specs/SPEC-CONFIG-001/plan.md:46` | T1.7 "Define PricingConfig in config types" |
| `.moai/specs/SPEC-CONFIG-001/plan.md:231` | Pricing type decl block |

**Interpretation**: The runtime yaml `pricing.yaml` was never created in v1.1.0 (deferred deliverable). The struct + defaults + manager wiring exist as **explicit forward-compat scaffolding**. The "zero yaml consumer" signal does not equal dead — it equals "yaml deliverable pending in completed SPEC".

**Disposition**: Keep PricingConfig as-is. Audit signal was insufficient — symbol grep does not detect deliverable-pending scaffolding.

### 2.2 A2 — github-actions.yaml

**v1 claim**: "yaml is 725B with rich content (auto_panel/review.llms.{claude,codex,gemini,glm}/runner/sha_pin), code consumer 0, dead"

**v2 correction**: github-actions.yaml is **scaffolding for SPEC-CI-MULTI-LLM-001 v1.0.0 (status: in-progress)**. Evidence:

| Location | Content |
|----------|---------|
| `.moai/specs/SPEC-CI-MULTI-LLM-001/tasks.md` | T-21 status=`pending`, label "github-actions.yaml.tmpl (Config Template)" |
| `.moai/specs/SPEC-CI-MULTI-LLM-001/plan.md:64,86,375-381` | yaml + tmpl declared as REQ-CI-015/CI-020 deliverable |
| `.moai/specs/SPEC-CI-MULTI-LLM-001/spec.md:105,154,191,219,331,366,424` | runtime contract reading this yaml (auto_panel.trigger, review.llms.*, runner.mode, sha_pin.actions_checkout, etc.) |

**Interpretation**: "zero Go consumer" is true *today only because the consumer hasn't been written yet*. T-21 = pending means the implementer task is queued. Deleting the yaml would strand the in-progress SPEC.

**Secondary finding**: yaml is NOT registered in `internal/config/audit_registry.go` `yamlToStructRegistry` nor `yamlAuditExceptions`. This is an **orphan-registration defect** independent of the dead-code claim. Recommended fix: add `github_actions` to `yamlAuditExceptions` with comment "pending SPEC-CI-MULTI-LLM-001 T-21".

**Disposition**: Keep yaml + tmpl. Optionally add registry exception entry as separate hygiene task.

### 2.3 A3 — system.hook.observability_events

**v1 claim**: "yaml default `[]`, comment says 'silent no-op', field effectively unused even when set → dead"

**v2 correction**: The audit misread the comment. The "silent no-op" wording describes opt-in OFF behavior. When the list is populated, the field activates a live opt-in path:

| Location | Content |
|----------|---------|
| `internal/hook/observability.go:21-44` | LIVE `observabilityOptIn()` function reads `cfg.Get().System.Hook.ObservabilityEvents` |
| `internal/hook/observability.go` header | `@MX:ANCHOR fan_in=4`. Resolution comment: "KEEP — observability gate helper for RETIRE-OBS-ONLY handlers. Implements SPEC-V3R2-RT-006 REQ-040" |
| Call sites (4) | `notification`, `elicitation`, `elicitationResult`, `taskCreated` hook handlers |
| SPEC-V3R2-RT-006 | status: completed. REQ-040 defines opt-in semantics |
| SPEC-V3R2-MIG-002 | status: completed. Established the retire-with-opt-in pattern |

**Interpretation**: `ObservabilityEvents` is **active opt-in infrastructure for retired-event observability**. The field is "always-no-op-when-empty" by design, not by accident. Populating it activates 4 handlers as taps. Deleting the field would silently break the retired-event observability story without retiring REQ-040 first.

**Disposition**: Keep field. Audit's reading of the comment was incorrect.

---

## 3. Root Cause Analysis

### 3.1 What signal was used in v1

```
For each yaml top-level key K:
  1. grep -rln "yaml:\"K\"" internal pkg
  2. If hit count == 0 outside _test.go and templates/, mark "dead"
  3. For struct fields, check if consumer code outside config package references the field
  4. If no external consumer, mark "dead struct"
```

### 3.2 Why the signal fails (5 Whys)

1. **WHY** did v1 misclassify? — grep symbol reference is a **necessary but insufficient** signal for dead-config detection.
2. **WHY** insufficient? — Three orthogonal "alive" patterns are invisible to grep:
   - **Forward-compat scaffolding**: code wired but yaml deliverable deferred (A1 pattern)
   - **In-progress deliverable**: yaml exists ahead of consumer code (A2 pattern)
   - **Opt-in inactive-by-default**: empty default means inactive but populating activates (A3 pattern)
3. **WHY** wasn't SPEC catalog checked? — v1 audit was built around runtime artifacts (yaml + Go code). SPEC documents at `.moai/specs/<SPEC-ID>/{spec,plan,acceptance,tasks}.md` declare ownership and lifecycle but live outside the runtime tree.
4. **WHY** is `audit_registry.go` insufficient? — Registry tracks yaml→struct symmetry but does not encode SPEC ownership, status, or supersession. github-actions is orphan-registered (neither in registry nor exceptions) which made it look extra-dead.
5. **WHY** wasn't this caught earlier? — `audit_loader_completeness_test.go` and `audit_struct_yaml_symmetry_test.go` test symmetry between Go struct and yaml. They do NOT cross-reference SPEC catalog. This is a test infrastructure gap.

### 3.3 Bias contributor

The v1 audit was framed as "find dead code", which biases toward false positives. A balanced framing ("classify each config item as: live / dormant-by-design / forward-compat / dead") would have produced different candidates.

---

## 4. Required Procedure for v3 Audit

Future config dead-code audits MUST execute these checks per candidate item:

### Step 1 — Symbol reference (necessary)

```bash
# Same as v1
grep -rn "yaml:\"<key>\"" internal pkg
grep -rn "<StructName>" internal pkg | grep -v _test | grep -v templates
```

### Step 2 — SPEC catalog cross-check (mandatory addition)

```bash
# Find owner SPECs
grep -rln "<StructName>\|<yaml_key>" .moai/specs/

# For each owner SPEC, read frontmatter and tasks
for spec in $(grep -rln "<symbol>" .moai/specs/ | sort -u); do
  echo "=== $spec ==="
  head -20 "$spec"  # frontmatter: status, superseded_by, version
  grep -n "$symbol" "$spec"  # context
done
```

### Step 3 — Owner SPEC disposition resolution

Per owner SPEC, classify:
- `status: completed` + symbol in spec.md as deliverable → **forward-compat scaffolding** (NOT dead)
- `status: in-progress` + symbol in tasks.md `pending` → **deliverable pending** (NOT dead)
- `status: archived/superseded` + no successor → **retire candidate** (verify successor SPEC)
- `superseded_by: <newer SPEC>` → check successor SPEC; if newer SPEC defines successor symbol → **safe to retire old**
- No owner SPEC found → **orphan** (possibly genuine dead, but verify with `git log -S <symbol>` for historical context)

### Step 4 — Consumer fan_in measurement (replace `grep -v _test`)

```bash
# Count NON-TEST consumers
grep -rn "<symbol>" internal pkg --include="*.go" | grep -v "_test.go" | grep -v "/templates/" | wc -l

# Count @MX:ANCHOR fan_in declarations
grep -rn "@MX:ANCHOR.*<symbol>\|fan_in.*<symbol>" internal pkg
```

A consumer count of 0 but `@MX:ANCHOR fan_in=N` declaration means audit must investigate why annotation and reality disagree (likely audit reading wrong symbol).

### Step 5 — Opt-in pattern detection

For struct fields with empty/zero default in yaml, check for opt-in pattern:
```bash
# Look for "If empty, ... otherwise ..." patterns
grep -B2 -A5 "ObservabilityEvents\|<field>" internal/hook/
grep -rn "len(<field>) == 0\|<field> != nil\|<field> != \"\"" internal pkg
```

If field gates a conditional path, the empty-default does NOT mean dead — it means **inactive by default**.

### Step 6 — Cross-SPEC dependency graph

For each candidate, list:
- Defining SPEC(s)
- Consuming SPEC(s)
- Superseding SPEC(s) (if any)
- Owner SPEC's current status

Only items where ALL owner SPECs are `superseded` or `archived` AND no consuming SPEC depends on them are eligible for dead classification.

---

## 5. v1 Audit's Other Categories (B/C/D) — Re-verification Required

The v1 audit also categorized:

- **B1**: `git-strategy.yaml` nested mode-based structure ~80% dead
- **B2**: `workflow.yaml` completion/loop_prevention/memory/default_mode/execution_mode ~60% dead
- **B3**: `lsp.yaml` aggregator/circuit_breaker/delegate_to_astgrep/discovery/enabled ~80% dead
- **B4**: `llm.yaml` default_model/quality_model/speed_model legacy fields
- **B5**: `constitution.yaml.performance.{max_memory_mb,max_response_time_ms}`
- **C1-C5**: Various dormant items (SunsetConfig, trust5_integration, skip_conditions, simplicity.max_parallel_tasks, report_generation.user_choice)
- **D1-D3**: Struct overhang items (design.gan_loop.strict_mode, state.retention_days, session.stale_seconds)

**Status**: These were NOT verified against SPEC catalog in v1. They remain candidates but **MUST be re-verified via Step 1-6 above before any cleanup SPEC is filed**. The categorization may shift significantly once SPEC ownership is mapped.

---

## 6. Lessons for Future Audits

1. **Grep symbol reference is necessary but insufficient** for dead-config detection.
2. **SPEC catalog cross-check is mandatory** — read `.moai/specs/<*>/{spec,plan,tasks}.md` frontmatter (status / superseded_by) and content (deliverables, REQ-X-NNN mappings) for every candidate.
3. **Three "alive but invisible to grep" patterns**: forward-compat scaffolding, in-progress deliverable, opt-in inactive-by-default.
4. **`audit_registry.go` does not encode SPEC lifecycle** — orphan entries (not in registry AND not in exceptions) require independent SPEC catalog investigation.
5. **Audit framing matters** — "find dead code" biases toward false positives. Use neutral classification ("live / dormant / forward-compat / dead") instead.
6. **`@MX:ANCHOR fan_in=N`** disagreement with grep results is a red flag: either the annotation is stale or the audit is reading the wrong symbol.

---

## 7. Disposition

- **No code change**: 0 files modified.
- **No SPEC created**: `SPEC-V3R5-CONFIG-DEAD-CLEANUP-001` abandoned.
- **Working tree**: clean (manager-spec made no writes; this document is the only new file).
- **Follow-up SPEC candidates**: Deferred to a v2 audit re-run that applies the Step 1-6 procedure to CATEGORY B/C/D candidates. Each surviving candidate needs its own scoped SPEC with explicit owner-SPEC supersession trail.

---

## 8. References

- v1 audit (initial misclassification): turn-of-2026-05-22 orchestrator response, not persisted to disk.
- `internal/config/types.go` — Config aggregate root + section structs.
- `internal/config/audit_registry.go` — yaml↔struct symmetry registry.
- `internal/config/audit_loader_completeness_test.go` — current audit coverage (missing SPEC catalog cross-check).
- `internal/config/audit_struct_yaml_symmetry_test.go` — symmetry test (also missing SPEC cross-check).
- `internal/hook/observability.go` — A3 active consumer (proves A3 is live).
- `.moai/specs/SPEC-CONFIG-001/` — A1 owner.
- `.moai/specs/SPEC-CI-MULTI-LLM-001/` — A2 owner.
- `.moai/specs/SPEC-V3R2-RT-006/` — A3 owner.
- `.moai/specs/SPEC-V3R2-MIG-002/` — A3 retire-with-opt-in pattern origin.
