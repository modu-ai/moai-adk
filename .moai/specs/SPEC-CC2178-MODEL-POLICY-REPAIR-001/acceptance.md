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

### AC-MPR-003 (MUST) — Deployed Default model is `sonnet`

**Given** the deployed `settings.json` rendered from the template,
**When** a reviewer reads the Default-model field (top-level `model:` or CC 2.1.175 Default-resolution key — exact key confirmed at M1),
**Then** the value is `sonnet`.

**Verification command**:
```bash
# Render the template to a temp dir and inspect
moai init /tmp/mpr-verify-001 --force 2>/dev/null
grep '"model"' /tmp/mpr-verify-001/.claude/settings.json
# Expected: "model": "sonnet" (or the CC 2.1.175 Default-resolution equivalent)
```

**REQ binding**: REQ-MPR-003.

---

### AC-MPR-004 (MUST) — Harness `minimal` resolves to lightweight cycle_type

**Given** a SPEC classified as harness `minimal` (e.g., a typo-fix Tier S),
**When** the harness routing resolves the cycle_type,
**Then** the resolved cycle_type is NOT `tdd` (it is a skip_phases-driven DDD-lite or equivalent lightweight cycle).

**Verification**: go test in `internal/harness/` (or `internal/config/`) asserting `resolveCycleType("minimal") != "tdd"`.

**REQ binding**: REQ-MPR-004.

---

### AC-MPR-005 (MUST) — Harness `thorough` resolves to full TDD

**Given** a SPEC classified as harness `thorough` (e.g., a critical-feature Tier L),
**When** the harness routing resolves the cycle_type,
**Then** the resolved cycle_type is `tdd`.

**Verification**: go test asserting `resolveCycleType("thorough") == "tdd"`.

**REQ binding**: REQ-MPR-005.

---

### AC-MPR-006 (MUST) — Explicit `quality.yaml development_mode: tdd` pin preserved (backward-compat)

**Given** a project whose `quality.yaml` explicitly sets `development_mode: tdd`,
**When** the harness routing resolves the cycle_type for a `minimal`-classified SPEC,
**Then** the explicit pin is preserved (the resolved cycle_type is `tdd`, NOT overridden by the harness level).

**Verification**: go test asserting that when `quality.yaml development_mode` is non-empty, it takes precedence over the harness-derived cycle_type.

**REQ binding**: REQ-MPR-006. **Anti-goal binding**: AG-01.

---

### AC-MPR-007 (MUST) — `agentModelMap` has 0 archived-phantom keys

**Given** the file `internal/template/model_policy.go` after M2,
**When** a reviewer greps the `agentModelMap` (L193-216) for the 13 archived-phantom keys + 2 legacy aliases,
**Then** the grep returns 0 matches.

**Verification command**:
```bash
grep -cE '"(manager-ddd|manager-tdd|manager-quality|manager-project|manager-strategy|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-debug|expert-testing|expert-refactoring|builder-agent|builder-skill|builder-plugin)"\s*:' internal/template/model_policy.go
# Expected: 0
```

**REQ binding**: REQ-MPR-008.

---

### AC-MPR-008 (MUST) — `manager-develop` and `builder-harness` covered in `agentModelMap`

**Given** the file `internal/template/model_policy.go` after M2,
**When** `GetAgentModel(ModelPolicyHigh, "manager-develop")` and `GetAgentModel(ModelPolicyHigh, "builder-harness")` are called,
**Then** both return a non-empty model string (not `""`).

**Verification**: go test `TestGetAgentModel_RetainedAgents` asserting non-empty returns for both agents across all 3 policies (High/Medium/Low).

**REQ binding**: REQ-MPR-009.

---

### AC-MPR-009 (MUST) — `agentEffortMap` has 0 archived-phantom keys (or is fully retired)

**Given** the file `internal/template/model_policy.go` after M2,
**When** a reviewer greps for the 3 archived-phantom effort keys,
**Then** EITHER the grep returns 0 matches (prune-and-keep decision) OR the `agentEffortMap` variable no longer exists (retire decision per M1).

**Verification command** (prune-and-keep path):
```bash
grep -cE '"(manager-strategy|expert-security|expert-refactoring)"\s*:' internal/template/model_policy.go
# Expected: 0
```

**Verification command** (retire path):
```bash
grep -c 'agentEffortMap' internal/template/model_policy.go
# Expected: 0 (the variable, GetAgentEffort, and ApplyEffortPolicy all removed)
```

**REQ binding**: REQ-MPR-010, REQ-MPR-011.

---

### AC-MPR-010 (MUST) — Effort-map redundancy decision recorded with test consequences

**Given** the M1 research note,
**When** a reviewer reads the redundancy decision,
**Then** the decision (retire vs. prune-and-keep) is explicitly recorded WITH (a) the grep evidence (which retained agents have hand-authored `effort:`), (b) the caller-grep evidence (whether `GetAgentEffort`/`ApplyEffortPolicy` have non-test callers), and (c) the test consequences (which tests are removed/updated if retired).

**Verification**: the research note (inline in plan.md §B.4 or sibling `research.md`) contains all 3 evidence items.

**REQ binding**: REQ-MPR-011, REQ-MPR-012.

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

### AC-MPR-013 (SHOULD) — `moai spec lint` exits 0

**Given** the SPEC directory `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/` with all 3 artifacts,
**When** a reviewer runs `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001`,
**Then** the command exits 0 with no MUST-FIX findings (warnings acceptable with rationale).

**Verification command**:
```bash
moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001
# Expected: exit 0
```

**REQ binding**: spec.md §D summary. **Non-blocking**: lint warnings with rationale do not block close.

---

## §D.1 Severity Breakdown

| AC | Severity | Blocking? |
|----|----------|-----------|
| AC-MPR-001 | MUST | Yes |
| AC-MPR-002 | MUST | Yes |
| AC-MPR-003 | MUST | Yes |
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

**Total**: 11 MUST + 2 SHOULD = 13 ACs.

## §D.2 Traceability Matrix

| REQ | AC | Milestone |
|-----|----|-----------|
| REQ-MPR-001 (availableModels) | AC-MPR-001 | M4 |
| REQ-MPR-002 (enforceAvailableModels) | AC-MPR-002 | M4 |
| REQ-MPR-003 (Default=sonnet) | AC-MPR-003 | M4 |
| REQ-MPR-004 (minimal→lightweight) | AC-MPR-004 | M3 |
| REQ-MPR-005 (thorough→tdd) | AC-MPR-005 | M3 |
| REQ-MPR-006 (explicit-pin preserved) | AC-MPR-006 | M3 |
| REQ-MPR-008 (agentModelMap no phantoms) | AC-MPR-007 | M2 |
| REQ-MPR-009 (manager-develop/builder-harness) | AC-MPR-008 | M2 |
| REQ-MPR-010 (agentEffortMap no phantoms) | AC-MPR-009 | M2 |
| REQ-MPR-011/012 (effort-map redundancy decision) | AC-MPR-010 | M1 |
| REQ-MPR-013/014/015 ([1m] verdict) | AC-MPR-011 | M1, M5 |
| REQ-MPR-016/017 (task-triage decision) | AC-MPR-012 | M1 |
| (spec lint) | AC-MPR-013 | M6 |

## §D.3 Indirect Verification (Trust-but-verify batch, run at M6)

The following read-only verifications run in parallel at M6 to independently confirm defect resolution:

1. **D1 resolved**: `grep -cE '"(manager-ddd|manager-tdd|manager-quality|manager-project|manager-strategy|expert-.*|builder-agent|builder-skill|builder-plugin)"\s*:' internal/template/model_policy.go` → 0.
2. **D2 resolved**: `grep -c '"manager-strategy"\|"expert-security"\|"expert-refactoring"' internal/template/model_policy.go` → 0 (or variable absent).
3. **D3 resolved**: `grep -c 'availableModels\|enforceAvailableModels' internal/template/templates/.claude/settings.json.tmpl` → ≥2.
4. **D4 resolved**: `go test ./internal/config/... ./internal/harness/... -run CycleType` → PASS.
5. **Template neutrality**: `go test ./internal/template/... -run TestTemplateNeutralityAudit` → PASS (no internal-content leak).
6. **Mirror parity**: `go test ./internal/template/... -run TestEmbeddedMirror` → PASS.
7. **Full suite**: `go test ./...` → PASS (0 regressions).

## §D.4 Closure Gates (Definition of Done)

A SPEC is close-eligible when ALL of the following hold:

1. All 11 MUST ACs PASS with observed evidence (not assumed — per verification-claim-integrity.md §1.1).
2. The 2 SHOULD ACs either PASS or have documented deferral rationale.
3. `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001` exits 0.
4. `go test ./...` exits 0 (0 regressions).
5. The plan-auditor verdict is ≥ 0.80 (Tier M threshold) AND the Implementation Kickoff Approval was obtained before run-phase (§I of plan.md).
6. The sync-auditor verdict is PASS or PASS-WITH-DEBT (if sync-phase runs).
7. No uncommitted working-tree files from the unrelated parallel workstream (AG-03) were absorbed into this SPEC's commits.

## §D.5 Forward-Looking Checks (post-close)

- **FL-01**: If REQ-MPR-015 fired (relaxation verdict), a follow-up SPEC `SPEC-CC2178-PER-AGENT-PIN-RELAXATION-001` is recorded in `.moai/specs/` as a `draft` placeholder.
- **FL-02**: If AC-MPR-012 deferred task-triage, a follow-up SPEC `SPEC-CC2178-TASK-TRIAGE-001` is recorded as a `draft` placeholder.
- **FL-03**: The cost-savings measurement SPEC (`SPEC-CC2178-COST-TELEMETRY-001`, EX-04) is recorded as a `draft` placeholder regardless (it is always deferred from this SPEC).
- **FL-04**: The vff prose-discipline SPEC (`SPEC-CC2178-VFF-PROSE-DISCIPLINE-001`, EX-02) is recorded as a `draft` placeholder regardless.

## §D.6 Edge Cases

- **EC-01**: What if M1 `[1m]` re-verification cannot fetch upstream issues (network-restricted environment)? → Record the fetch failure in the research note, default to "still-active" (conservative — preserves EX-01), and note that the re-verification is incomplete. AC-MPR-011 PASSES with the documented limitation.
- **EC-02**: What if the CC 2.1.178 Default-model field is NOT `model:` but a new key (e.g., `defaultModel`)? → M1 confirms the exact key; M4 uses the confirmed key; AC-MPR-003 verification command is updated to grep the confirmed key.
- **EC-03**: What if M3 reveals the cycle_type routing requires touching >3 files (Split Trigger fires)? → The cycle_type axis splits into `SPEC-CC2178-CYCLE-TYPE-ROUTING-001`; THIS SPEC's AC-MPR-004/005/006 transfer to the follow-up; this SPEC closes on the model + cleanup axes only, with a documented scope-reduction note in §H.
- **EC-04**: What if `ApplyEffortPolicy` has a non-test caller that blocks retirement? → M1 records the caller; the decision becomes "prune-and-keep" (remove phantom keys, keep the mechanism); AC-MPR-009 follows the prune-and-keep verification path.

## §D.7 Quality Gate Criteria

- **Lint**: `golangci-lint run` → 0 errors.
- **Test**: `go test ./...` → 0 failures, 0 regressions.
- **Coverage**: `internal/template/` package coverage ≥ baseline (no drop from phantom-map cleanup).
- **Spec lint**: `moai spec lint SPEC-CC2178-MODEL-POLICY-REPAIR-001` → exit 0.
- **Template neutrality**: `TestTemplateNeutralityAudit` → PASS.
- **Mirror parity**: `TestEmbeddedMirror` → PASS.
