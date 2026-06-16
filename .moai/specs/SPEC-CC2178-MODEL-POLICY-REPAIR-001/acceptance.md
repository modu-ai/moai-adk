# Acceptance Criteria — SPEC-CC2178-MODEL-POLICY-REPAIR-001

> **Severity model**: MUST = blocking (SPEC cannot close without PASS). SHOULD = non-blocking (deferred debt allowed with rationale).
>
> **Falsifiability**: every AC below is observable via a concrete command (grep, go test, file read, moai spec lint). No subjective or unfalsifiable ACs (per L_plan_auditor_value_realization_unfalsifiable).

## §D. AC Matrix

### AC-MPR-001 (MUST) — `availableModels` field present in settings template

**Given** the template file `internal/template/templates/.claude/settings.json.tmpl` after M4,
**When** a reviewer greps for `availableModels`,
**Then** the field is present and lists exactly the 3 aliases `sonnet`, `opus`, `haiku`.

**Verification command**:
```bash
grep -A5 '"availableModels"' internal/template/templates/.claude/settings.json.tmpl
# Expected: a JSON array containing "sonnet", "opus", "haiku"
```

**REQ binding**: REQ-MPR-001.

---

### AC-MPR-002 (MUST) — `enforceAvailableModels: true` present

**Given** the template file `internal/template/templates/.claude/settings.json.tmpl` after M4,
**When** a reviewer greps for `enforceAvailableModels`,
**Then** the field is present and set to `true`.

**Verification command**:
```bash
grep '"enforceAvailableModels"' internal/template/templates/.claude/settings.json.tmpl
# Expected: "enforceAvailableModels": true
```

**REQ binding**: REQ-MPR-002.

---

### AC-MPR-003 (MUST, CONDITIONAL verification command) — Deployed Default model is `sonnet`

**Given** the deployed `settings.json` rendered from the template after M4,
**When** a reviewer reads the Default-model field at the key confirmed by M1 task 4,
**Then** the value is `sonnet`.

> **D4 conditional (iter-2)**: post-CC-2.1.175 the settings file carries `model`, `availableModels`, and possibly other model-named keys, so a naive `grep '"model"'` matches multiple lines. The EXACT verification command is **pinned at M1** once REQ-MPR-013 re-verification confirms the CC 2.1.178 Default-resolution JSON key. M1 MUST record (a) the confirmed key name and (b) the exact grep/JSON-inspection command. Do NOT pretend the verification is ready now — this AC is a PLACEHOLDER until M1 records the command. If M1 cannot confirm the key (e.g., CC 2.1.178 runtime not locally available), the fallback is render + manual JSON inspection of the rendered file.

**Verification command (TENTATIVE — to be pinned at M1)**:
```bash
# PLACEHOLDER — exact key confirmed at M1 task 4
moai init /tmp/mpr-verify-001 --force 2>/dev/null
# M1 records the confirmed key, e.g.:
#   grep '"<confirmed-key>"' /tmp/mpr-verify-001/.claude/settings.json
# Expected: the confirmed Default-model key resolves to "sonnet"
```

**REQ binding**: REQ-MPR-003. **M1 binding**: M1 task 4 pins the verification command before M4 GREEN.

---

### AC-MPR-004 (MUST) — Harness `minimal` resolves to lightweight cycle_type

**Given** the `ResolveCycleType` symbol authored at M3 (plan.md §F.3.1 contract — package `internal/config`, signature `func ResolveCycleType(harnessLevel string, explicitPin string) string`),
**When** a test calls `ResolveCycleType("minimal", "")`,
**Then** the return value is `"ddd"` (the lightweight DDD-lite cycle, NOT `"tdd"`).

> **D1 remediation (iter-2)**: the iter-1 AC asserted `resolveCycleType("minimal") != "tdd"` — a non-existent symbol (`grep -rn "resolveCycleType\|ResolveCycleType" internal/` returned 0 matches on 2026-06-16). This iter-2 AC references the agreed symbol `ResolveCycleType` (exported, `internal/config` package) with the exact 2-argument signature from plan.md §F.3.1, and asserts the concrete return value `"ddd"` per the dispatch table.

**Verification command**:
```bash
go test ./internal/config/... -run TestResolveCycleType -v
# Expected: a sub-test asserting ResolveCycleType("minimal", "") == "ddd" PASSES
```

**REQ binding**: REQ-MPR-004.

---

### AC-MPR-005 (MUST) — Harness `thorough` resolves to full TDD

**Given** the `ResolveCycleType` symbol authored at M3,
**When** a test calls `ResolveCycleType("thorough", "")`,
**Then** the return value is `"tdd"`.

**Verification command**:
```bash
go test ./internal/config/... -run TestResolveCycleType -v
# Expected: sub-test asserting ResolveCycleType("thorough", "") == "tdd" PASSES
```

**REQ binding**: REQ-MPR-005.

---

### AC-MPR-006 (MUST) — Explicit `quality.yaml constitution.development_mode: tdd` pin preserved (backward-compat)

**Given** the `ResolveCycleType` symbol authored at M3,
**When** a test calls `ResolveCycleType("minimal", "tdd")` (simulating an explicit `constitution.development_mode: tdd` pin in `quality.yaml`),
**Then** the return value is `"tdd"` (the explicit pin wins over the harness-derived `"ddd"`).

> **iter-2 prose correction**: the iter-1 AC said "`quality.yaml explicitly sets `development_mode: tdd`" without locating the key. Verified 2026-06-16: the key is nested under `constitution:` (`quality.yaml` L1-2: `constitution: development_mode: tdd`). The Go reader is `internal/config/manager.go:394` (`MOAI_DEVELOPMENT_MODE` env) + the parsed `cfg.Quality.DevelopmentMode` struct field.

**Verification command**:
```bash
go test ./internal/config/... -run TestResolveCycleType -v
# Expected: sub-test asserting ResolveCycleType("minimal", "tdd") == "tdd" PASSES
```

**REQ binding**: REQ-MPR-006. **Anti-goal binding**: AG-01.

---

### AC-MPR-007 (MUST) — `agentModelMap` has 0 of the 16 canonical phantom keys

**Given** the file `internal/template/model_policy.go` after M2,
**When** a reviewer greps for the 16 canonical phantom keys enumerated in spec.md §C.3 (13 archived phantoms + `manager-ddd` + `manager-tdd` + `builder-agent` + `builder-skill` + `builder-plugin`... precisely: the 16 keys from the §C.3 table),
**Then** the grep returns 0 matches.

> **iter-2 D2 unification**: this AC now uses the SAME 16-key enumeration as REQ-MPR-008 and §D.3 — the single source of truth is spec.md §C.3 table. The iter-1 AC used a 16-key regex but REQ-MPR-008 listed 13; iter-2 unifies to 16 everywhere. Note the count is 16 (not the iter-1 "13" or the audit's "15"): the map has 19 entries total, 3 retained, 16 to remove.

**Verification command** (canonical — matches spec.md §C.3 + §D.3):
```bash
grep -cE '"(manager-ddd|manager-tdd|manager-quality|manager-project|manager-strategy|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-debug|expert-testing|expert-refactoring|builder-agent|builder-skill|builder-plugin)"\s*:' internal/template/model_policy.go
# Expected: 0
```

**REQ binding**: REQ-MPR-008.

---

### AC-MPR-008 (MUST) — `manager-develop` and `builder-harness` covered with iter-2 tuples

**Given** the file `internal/template/model_policy.go` after M2,
**When** `GetAgentModel(ModelPolicyHigh, "manager-develop")` and `GetAgentModel(ModelPolicyHigh, "builder-harness")` are called,
**Then** both return `"sonnet"` (the iter-2 tuple `{sonnet, sonnet, haiku}` — High policy = sonnet; NOT the iter-1 `{opus, sonnet, sonnet}` derived from retired aliases per D6).

**Verification**: go test `TestGetAgentModel_RetainedAgents` asserting `GetAgentModel(ModelPolicyHigh, "manager-develop") == "sonnet"` AND `GetAgentModel(ModelPolicyHigh, "builder-harness") == "sonnet"` across all 3 policies (High=sonnet, Medium=sonnet, Low=haiku).

**REQ binding**: REQ-MPR-009 (iter-2 tuple decision).

---

### AC-MPR-009 (MUST) — `agentEffortMap` has 0 archived-phantom keys (prune path ONLY, iter-2)

**Given** the file `internal/template/model_policy.go` after M2,
**When** a reviewer greps for the 3 archived-phantom effort keys,
**Then** the grep returns 0 matches.

> **iter-2 D5 correction**: the iter-1 AC offered a "retire path" alternative (the `agentEffortMap` variable removed entirely). Per D5 remediation, the retirement path is **DEFERRED** to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001` because `ApplyEffortPolicy` has 2 production callers. This SPEC takes the PRUNE path only. The variable, `GetAgentEffort`, and `ApplyEffortPolicy` all REMAIN.

**Verification command** (prune path — the ONLY path for this SPEC):
```bash
grep -cE '"(manager-strategy|expert-security|expert-refactoring)"\s*:' internal/template/model_policy.go
# Expected: 0
```

**REQ binding**: REQ-MPR-010, REQ-MPR-011a.

---

### AC-MPR-010 (MUST) — Effort-map phantom pruning + map↔file reconciliation applied (iter-2)

**Given** the file `internal/template/model_policy.go` after M2,
**When** a reviewer reads the `agentEffortMap` (L72-79),
**Then** (a) the 3 archived phantom keys are absent (AC-MPR-009), (b) `plan-auditor` and `sync-auditor` values are `xhigh` (synced from `high` to match the agent files per REQ-MPR-011b), (c) `manager-develop` (value `xhigh`) and `builder-harness` (value `high`) are present, AND (d) the M1 research note records the deferral rationale for full retirement (REQ-MPR-012 downgraded: the 2 production callers, the regression risk, the follow-up SPEC name `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001`).

**Verification command**:
```bash
# (a) phantoms absent — see AC-MPR-009
# (b)+(c) reconciliation values present:
grep -A8 'var agentEffortMap' internal/template/model_policy.go | grep -E '"(plan-auditor|sync-auditor|manager-develop|builder-harness)"'
# Expected: plan-auditor → xhigh, sync-auditor → xhigh, manager-develop → xhigh, builder-harness → high
```

**REQ binding**: REQ-MPR-011a, REQ-MPR-011b, REQ-MPR-012.

---

### AC-MPR-011 (MUST) — `[1m]` re-verification verdict recorded with upstream evidence

**Given** the M1 research note,
**When** a reviewer reads the `[1m]` verdict,
**Then** the verdict (still-active / relaxed / partially-relaxed) is explicitly recorded WITH (a) the fetched state of issues `#45847`, `#51060`, `#36670` at 2.1.178, (b) the relevant CC 2.1.178 CHANGELOG excerpt, and (c) the implication for per-agent pinning (EX-01 remains if still-active).

**Verification**: the research note contains all 3 evidence items + the verdict.

**REQ binding**: REQ-MPR-013, REQ-MPR-014, REQ-MPR-015.

---

### AC-MPR-012 (SHOULD) — Task-triage in-scope/deferred decision recorded

**Given** the M1 research note,
**When** a reviewer reads the task-triage decision,
**Then** the decision (in-scope vs. deferred) is explicitly recorded WITH rationale. If in-scope, the triage signal (failure-cost × visual-verifiability dimensions + mapping function) is defined concretely; if deferred, a follow-up SPEC candidate name is recorded.

**Verification**: the research note contains the decision + rationale (+ definition if in-scope / + follow-up name if deferred).

**REQ binding**: REQ-MPR-016, REQ-MPR-017. **Non-blocking**: may defer with rationale.

---

### AC-MPR-013 (SHOULD) — `moai spec lint` exits 0 (path-prefixed invocation, iter-2 D3)

**Given** the SPEC file `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md`,
**When** a reviewer runs `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md`,
**Then** the command exits 0 with no MUST-FIX (ERROR) findings (warnings acceptable with rationale).

> **D3 remediation (iter-2)**: the iter-1 AC used the bare SPEC ID `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001`, which the CLI treats as a filename and ParseFails (`open SPEC-CC2178-MODEL-POLICY-REPAIR-001: no such file or directory`). The PATH-PREFIXED form works. Per L_config_theater_removal_ac_pitfalls, an AC whose verification command does not reproduce is a defect even if the underlying claim is true. The observed output at iter-2 authoring (2026-06-16):
> ```
> SEVERITY  CODE                  FILE                                                     LINE  MESSAGE
> --------  ----                  ----                                                     ----  -------
> WARNING   StatusGitConsistency  .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md  1     SPEC SPEC-CC2178-MODEL-POLICY-REPAIR-001 frontmatter status 'draft' disagrees with git-implied status 'implemented'
>
> 0 error(s), 1 warning(s)
> --- exit code: 0 ---
> ```
> The single WARNING (`StatusGitConsistency`) is expected for an uncommitted plan-phase draft and is acceptable with rationale.

**Verification command** (canonical — path-prefixed):
```bash
moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md
# Expected: exit 0; 0 errors; warnings acceptable with rationale
```

**REQ binding**: spec.md §D summary. **Non-blocking**: lint warnings with rationale do not block close.

---

### AC-MPR-014 (MUST, NEW in iter-2 — D7 remediation) — Harness `standard` resolves to `tdd`

**Given** the `ResolveCycleType` symbol authored at M3 (plan.md §F.3.1 contract),
**When** a test calls `ResolveCycleType("standard", "")`,
**Then** the return value is `"tdd"` (the current global default is unchanged).

> **D7 remediation (iter-2)**: the iter-1 §D.2 traceability matrix omitted REQ-MPR-007 (`standard → tdd`), leaving it orphaned with no AC. This AC binds REQ-MPR-007 to M3 and references the same `ResolveCycleType` symbol contract from AC-MPR-004/005/006.

**Verification command**:
```bash
go test ./internal/config/... -run TestResolveCycleType -v
# Expected: sub-test asserting ResolveCycleType("standard", "") == "tdd" PASSES
```

**REQ binding**: REQ-MPR-007.

---

## §D.1 Severity Breakdown

| AC | Severity | Blocking? |
|----|----------|-----------|
| AC-MPR-001 | MUST | Yes |
| AC-MPR-002 | MUST | Yes |
| AC-MPR-003 | MUST (conditional cmd) | Yes |
| AC-MPR-004 | MUST | Yes |
| AC-MPR-005 | MUST | Yes |
| AC-MPR-006 | MUST | Yes (backward-compat) |
| AC-MPR-007 | MUST | Yes |
| AC-MPR-008 | MUST | Yes |
| AC-MPR-009 | MUST | Yes |
| AC-MPR-010 | MUST | Yes |
| AC-MPR-011 | MUST | Yes |
| AC-MPR-012 | SHOULD | No (defer with rationale) |
| AC-MPR-013 | SHOULD | No (warnings acceptable) |
| AC-MPR-014 | MUST (NEW iter-2) | Yes (binds orphaned REQ-MPR-007) |

**Total (iter-2)**: 12 MUST + 2 SHOULD = 14 ACs (iter-1 had 13; AC-MPR-014 added).

## §D.2 Traceability Matrix

| REQ | AC | Milestone |
|-----|----|-----------|
| REQ-MPR-001 (availableModels) | AC-MPR-001 | M4 |
| REQ-MPR-002 (enforceAvailableModels) | AC-MPR-002 | M4 |
| REQ-MPR-003 (Default=sonnet) | AC-MPR-003 (conditional) | M1 (key), M4 |
| REQ-MPR-004 (minimal→lightweight) | AC-MPR-004 | M3 |
| REQ-MPR-005 (thorough→tdd) | AC-MPR-005 | M3 |
| REQ-MPR-006 (explicit-pin preserved) | AC-MPR-006 | M3 |
| **REQ-MPR-007 (standard→tdd)** | **AC-MPR-014 (NEW iter-2)** | **M3** |
| REQ-MPR-008 (agentModelMap no phantoms) | AC-MPR-007 | M2 |
| REQ-MPR-009 (manager-develop/builder-harness, iter-2 tuples) | AC-MPR-008 | M2 |
| REQ-MPR-010 (agentEffortMap no phantoms) | AC-MPR-009 | M2 |
| REQ-MPR-011a/011b (prune + reconcile, iter-2) | AC-MPR-010 | M2 |
| REQ-MPR-012 (retirement deferral, iter-2 downgraded) | AC-MPR-010 | M1 |
| REQ-MPR-013/014/015 ([1m] verdict + relaxation conditions) | AC-MPR-011 | M1, M5 |
| REQ-MPR-016/017 (task-triage decision) | AC-MPR-012 | M1 |
| (spec lint, path-prefixed iter-2) | AC-MPR-013 | M6 |

## §D.3 Indirect Verification (Trust-but-verify batch, run at M6)

The following read-only verifications run in parallel at M6 to independently confirm defect resolution:

1. **D1 resolved (16-key canonical — matches AC-MPR-007 + spec.md §C.3)**: `grep -cE '"(manager-ddd|manager-tdd|manager-quality|manager-project|manager-strategy|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-debug|expert-testing|expert-refactoring|builder-agent|builder-skill|builder-plugin)"\s*:' internal/template/model_policy.go` → 0.
2. **D2 resolved (prune path, not retire)**: `grep -cE '"(manager-strategy|expert-security|expert-refactoring)"\s*:' internal/template/model_policy.go` → 0 (the `agentEffortMap` variable REMAINS per D5; only the 3 phantom keys are absent).
3. **D3 resolved**: `grep -c 'availableModels\|enforceAvailableModels' internal/template/templates/.claude/settings.json.tmpl` → ≥2.
4. **D4 resolved (ResolveCycleType authored)**: `grep -rn 'func ResolveCycleType' internal/config/cycle_type.go` → 1 match (the symbol exists, D1 remediation), AND `go test ./internal/config/... -run TestResolveCycleType` → PASS (covers AC-MPR-004/005/006/014).
5. **Template neutrality**: `go test ./internal/template/... -run TestTemplateNeutralityAudit` → PASS (no internal-content leak).
6. **Mirror parity**: `go test ./internal/template/... -run TestEmbeddedMirror` → PASS.
7. **Full suite**: `go test ./...` → PASS (0 regressions).
8. **Spec lint (path-prefixed, D3)**: `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` → exit 0, 0 errors.

## §D.4 Closure Gates (Definition of Done)

A SPEC is close-eligible when ALL of the following hold:

1. All 12 MUST ACs PASS with observed evidence (iter-2: 12 MUST, up from iter-1's 11 — AC-MPR-014 added; per verification-claim-integrity.md §1.1).
2. The 2 SHOULD ACs either PASS or have documented deferral rationale.
3. `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` exits 0 (path-prefixed — D3).
4. `go test ./...` exits 0 (0 regressions).
5. The plan-auditor verdict is ≥ 0.80 (Tier M threshold) AND the Implementation Kickoff Approval was obtained before run-phase (§I of plan.md).
6. The sync-auditor verdict is PASS or PASS-WITH-DEBT (if sync-phase runs).
7. No uncommitted working-tree files from the unrelated parallel workstream (AG-03 — 15 files: 10 modified + 5 untracked) were absorbed into this SPEC's commits.

## §D.5 Forward-Looking Checks (post-close)

- **FL-01**: If REQ-MPR-015 fired (relaxation verdict), a follow-up SPEC `SPEC-CC2178-PER-AGENT-PIN-RELAXATION-001` is recorded in `.moai/specs/` as a `draft` placeholder.
- **FL-02**: If AC-MPR-012 deferred task-triage, a follow-up SPEC `SPEC-CC2178-TASK-TRIAGE-001` is recorded as a `draft` placeholder.
- **FL-03**: The cost-savings measurement SPEC (`SPEC-CC2178-COST-TELEMETRY-001`, EX-04) is recorded as a `draft` placeholder regardless (it is always deferred from this SPEC).
- **FL-04**: The vff prose-discipline SPEC (`SPEC-CC2178-VFF-PROSE-DISCIPLINE-001`, EX-02) is recorded as a `draft` placeholder regardless.

## §D.6 Edge Cases

- **EC-01**: What if M1 `[1m]` re-verification cannot fetch upstream issues (network-restricted environment)? → Record the fetch failure in the research note, default to "still-active" (conservative — preserves EX-01), and note that the re-verification is incomplete. AC-MPR-011 PASSES with the documented limitation.
- **EC-02**: What if the CC 2.1.178 Default-model field is NOT `model:` but a new key (e.g., `defaultModel`)? → M1 confirms the exact key; M4 uses the confirmed key; AC-MPR-003 verification command is updated to grep the confirmed key.
- **EC-03**: What if M3 reveals the cycle_type routing requires touching >3 files (Split Trigger fires)? → The cycle_type axis splits into `SPEC-CC2178-CYCLE-TYPE-ROUTING-001`; THIS SPEC's AC-MPR-004/005/006 transfer to the follow-up; this SPEC closes on the model + cleanup axes only, with a documented scope-reduction note in §H.
- **EC-04 (iter-2 D5 RESOLVED)**: `ApplyEffortPolicy` HAS 2 non-test production callers (`initializer.go:181`, `update.go:2661`) — verified at plan-phase, NOT a hypothetical. The decision is therefore PRUNE-AND-KEEP (remove phantom keys + reconcile map↔file divergence, keep the mechanism). Full retirement is DEFERRED to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001` (caller migration required first). AC-MPR-009/010 follow the prune path only.

## §D.7 Quality Gate Criteria

- **Lint**: `golangci-lint run` → 0 errors.
- **Test**: `go test ./...` → 0 failures, 0 regressions.
- **Coverage**: `internal/template/` package coverage ≥ baseline (no drop from phantom-map cleanup).
- **Spec lint**: `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` → exit 0 (path-prefixed — D3).
- **Template neutrality**: `TestTemplateNeutralityAudit` → PASS.
- **Mirror parity**: `TestEmbeddedMirror` → PASS.
