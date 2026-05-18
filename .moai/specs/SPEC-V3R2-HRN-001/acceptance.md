# SPEC-V3R2-HRN-001 — Acceptance Criteria (acceptance.md)

> Companion to `spec.md` §6 AC summary (10 ACs).
> Companion to `plan.md` REQ↔AC matrix §5.
> Companion to `research.md` §13 test fixture inventory.

This file is the **binary verification contract** for HRN-001 run-phase exit. Each AC is independently testable, observable, and produces a binary PASS/FAIL outcome. All 10 ACs must PASS before `/moai sync` may execute.

---

## HISTORY

| Version | Date       | Author | Description |
|---------|------------|--------|-------------|
| 0.1.0   | 2026-05-18 | manager-spec | Initial plan-phase ACs: 10 binary Given/When/Then scenarios |

---

## Format

Each AC follows Given/When/Then format with an explicit **Verification command** producing binary exit code (0 = PASS, non-zero = FAIL) and an **Evidence artifact** (file path or stdout regex).

---

## AC-HRN-001-01 — Loader struct + test suite GREEN

**REQ coverage**: REQ-HRN-001-001 (struct), REQ-HRN-001-002 (LoadHarnessConfig signature)

**Given**:
- HRN-001 run-phase M2 has completed
- `internal/config/types.go` `HarnessConfig` struct contains all 6 extended fields: `DefaultProfile`, `ModeDefaults`, `AutoDetection`, `Escalation`, `EffortMapping`, `Levels`, `ModelUpgradeReview`, `Evaluator`
- `internal/config/loader.go` `LoadHarnessConfig()` parses the shipping schema without error

**When**:
```bash
go test -count=1 -race ./internal/config/... ./internal/harness/router/...
go test -cover ./internal/harness/router/...
```

**Then**:
- Exit code 0
- Coverage report on stdout shows `>= 85.0%` for `internal/harness/router/...`
- No `FAIL` lines in test output

**Evidence artifact**: stdout containing `ok ... internal/harness/router` and `coverage: NN.N% of statements` where `NN.N >= 85.0`

---

## AC-HRN-001-02 — Standard-level routing for ORC-001

**REQ coverage**: REQ-HRN-001-003 (Route signature), REQ-HRN-001-006 (CLI route subcommand), REQ-HRN-001-007 (priority order)

**Given**:
- `.moai/specs/SPEC-V3R2-ORC-001/spec.md` exists with frontmatter `priority: P1` and tags spanning multiple domains
- `.moai/config/sections/harness.yaml` is the shipping schema (unmodified)

**When**:
```bash
moai harness route --spec SPEC-V3R2-ORC-001 --json
```

**Then**:
- Exit code 0
- stdout is valid JSON parseable via `jq`
- JSON document contains: `.level == "standard"`
- JSON document contains: `.rationale.matched_rule` containing `"file_count > 3"` OR `"multi_domain"` (mode_defaults.solo / `auto_detection.rules.standard`)

**Verification command**:
```bash
moai harness route --spec SPEC-V3R2-ORC-001 --json | jq -e '.level == "standard" and (.rationale.matched_rule | contains("standard") or contains("multi_domain") or contains("file_count"))'
```

**Evidence artifact**: stdout JSON document; jq returns exit 0

---

## AC-HRN-001-03 — Thorough-level routing for security/payment SPEC

**REQ coverage**: REQ-HRN-001-008 (force-thorough on security/payment keywords)

**Given**:
- A SPEC whose body or title contains any keyword from `securityKeywords` (`auth`, `crypto`, `encrypt`, `oauth`, `jwt`, `session`, `password`, `rbac`, `acl`) OR `paymentKeywords` (`payment`, `billing`, `subscription`, `invoice`, `charge`, `stripe`, `paypal`)
- For canary: use `SPEC-V3R2-HRN-002` which mentions `evaluator` and amendment-domain semantics (or fall back to a synthetic fixture `.moai/specs/SPEC-TEST-AUTH-001/spec.md` if HRN-002 does not match)

**When**:
```bash
moai harness route --spec SPEC-V3R2-HRN-002 --json
```

**Then**:
- Exit code 0
- JSON `.level == "thorough"`
- JSON `.rationale.matched_rule` contains `"force_thorough"` or `"security_keywords"` or `"payment_keywords"`
- JSON `.rationale.keywords` is a non-empty array containing at least one matched keyword

**Verification command**:
```bash
moai harness route --spec SPEC-V3R2-HRN-002 --json | jq -e '.level == "thorough" and (.rationale.keywords | length) > 0'
```

**Evidence artifact**: stdout JSON document

---

## AC-HRN-001-04 — Shipping harness.yaml validates clean

**REQ coverage**: REQ-HRN-001-002 (loader), REQ-HRN-001-006 (CLI validate subcommand)

**Given**:
- Unmodified `.moai/config/sections/harness.yaml`
- `.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md` all present with valid rubric markdown

**When**:
```bash
moai harness validate
```

**Then**:
- Exit code 0
- stderr empty (no `slog.Warn` warnings)
- stdout contains `harness.yaml: OK` or equivalent success marker

**Verification command**:
```bash
moai harness validate && echo PASS
```

**Evidence artifact**: stdout containing `OK` or `PASS`

---

## AC-HRN-001-05 — FROZEN pass_threshold floor enforcement

**REQ coverage**: REQ-HRN-001-010 (validation error wrapping), REQ-HRN-001-012 (FROZEN 0.60 floor)

**Given**:
- A copy of `.moai/config/sections/harness.yaml` and `.moai/config/evaluator-profiles/lenient.md` is placed in `/tmp/test-floor/`
- The `lenient.md` copy is edited to introduce an anchor at `0.5` for `Functionality` dimension
- The harness.yaml copy points `levels.standard.evaluator_profile: lenient`

**When**:
```bash
moai harness validate --path /tmp/test-floor/.moai/config/sections/harness.yaml
```

**Then**:
- Exit code 1
- stderr contains the sentinel error code `HRN_PASS_THRESHOLD_FLOOR` (or wrapped via `ValidationError.Field == "levels.standard.evaluator_profile"`)
- stderr names the offending field path AND value (`0.5`)

**Verification command**:
```bash
moai harness validate --path /tmp/test-floor/.moai/config/sections/harness.yaml 2>&1 | grep -q "HRN_PASS_THRESHOLD_FLOOR"; echo "exit=$?"
```

**Evidence artifact**: stderr line matching `/HRN_PASS_THRESHOLD_FLOOR.*0\.5/`

---

## AC-HRN-001-06 — JSON output schema conformance

**REQ coverage**: REQ-HRN-001-011 (JSON output)

**Given**:
- Any valid SPEC ID

**When**:
```bash
moai harness route --spec SPEC-V3R2-ORC-001 --json
```

**Then**:
- Exit code 0
- stdout is exactly one well-formed JSON document (no preamble, no trailing prose)
- JSON document conforms to schema:
  ```
  {
    "level": "minimal" | "standard" | "thorough",
    "rationale": {
      "matched_rule": <string>,
      "file_count": <int>,
      "domain_count": <int>,
      "spec_type": <string>,
      "spec_priority": <string>,
      "keywords": [<string>, ...]
    },
    "effort": "medium" | "high" | "xhigh",
    "evaluator_profile": <string>,
    "sprint_contract": <bool>,
    "plan_audit": <bool>
  }
  ```

**Verification command**:
```bash
moai harness route --spec SPEC-V3R2-ORC-001 --json | \
  jq -e 'has("level") and has("rationale") and has("effort") and has("evaluator_profile") and has("sprint_contract") and has("plan_audit") and (.rationale | has("matched_rule") and has("file_count") and has("domain_count") and has("spec_type") and has("spec_priority") and has("keywords"))'
```

**Evidence artifact**: jq exit 0; full JSON document captured to stdout

---

## AC-HRN-001-07 — Schema drift detection in strict mode

**REQ coverage**: REQ-HRN-001-019 (MOAI_CONFIG_STRICT)

**Given**:
- A copy of `.moai/config/sections/harness.yaml` in `/tmp/test-drift/` with an added unknown key `unknown_top_field: 1` at the top of the `harness:` section

**When (Test A — default mode)**:
```bash
unset MOAI_CONFIG_STRICT
moai harness validate --path /tmp/test-drift/.moai/config/sections/harness.yaml
```

**Then (Test A)**:
- Exit code 0
- stderr contains a `HRN_SCHEMA_DRIFT` warning naming `unknown_top_field`
- Validation does NOT fail (warning-only mode per REQ-019)

**When (Test B — strict mode)**:
```bash
MOAI_CONFIG_STRICT=1 moai harness validate --path /tmp/test-drift/.moai/config/sections/harness.yaml
```

**Then (Test B)**:
- Exit code 1
- stderr contains `HRN_SCHEMA_DRIFT` as an error (not warning)

**Verification command**:
```bash
# Test A (warning, exit 0)
unset MOAI_CONFIG_STRICT
moai harness validate --path /tmp/test-drift/.moai/config/sections/harness.yaml 2>&1 | grep -q "HRN_SCHEMA_DRIFT.*unknown_top_field" && \
  moai harness validate --path /tmp/test-drift/.moai/config/sections/harness.yaml > /dev/null 2>&1
# Test B (error, exit 1)
MOAI_CONFIG_STRICT=1 moai harness validate --path /tmp/test-drift/.moai/config/sections/harness.yaml; [ $? -eq 1 ]
```

**Evidence artifact**: stderr lines matching both test scenarios

---

## AC-HRN-001-08 — Escalation cap enforcement

**REQ coverage**: REQ-HRN-001-004 (EscalationManager), REQ-HRN-001-009 (triggers), REQ-HRN-001-013 (max_escalations cap), REQ-HRN-001-018 (cap-reached log)

**Given**:
- A test harness fixture with `escalation.max_escalations: 2`
- A test runner that simulates 3 consecutive `quality_gate_fail` events starting from `LevelMinimal`

**When**:
```bash
go test -run TestEscalationCapEnforcement ./internal/harness/router/...
```

**Then**:
- Exit code 0
- First escalation: `LevelMinimal` → `LevelStandard`, `escalated: true`
- Second escalation: `LevelStandard` → `LevelThorough`, `escalated: true`
- Third escalation attempt: stays at `LevelThorough`, `escalated: false`, log emits `HRN_ESCALATION_CAP_REACHED`

**Verification command**:
```bash
go test -run TestEscalationCapEnforcement -v ./internal/harness/router/... | grep -q "PASS"
```

**Evidence artifact**: `internal/harness/router/escalation_test.go` test PASS line; log output containing `HRN_ESCALATION_CAP_REACHED`

---

## AC-HRN-001-09 — SPEC frontmatter override (spec_override matched_rule)

**REQ coverage**: REQ-HRN-001-015 (SPEC harness_level: override)

**Given**:
- Test fixture `.moai/specs/SPEC-TEST-OVERRIDE-001/spec.md` with frontmatter containing `harness_level: thorough` (NEW optional field)
- SPEC body intentionally lacks security/payment keywords (no force-thorough trigger)
- SPEC file_count + domain_count would naturally route to `minimal`

**When**:
```bash
moai harness route --spec SPEC-TEST-OVERRIDE-001 --json
```

**Then**:
- Exit code 0
- JSON `.level == "thorough"` (matching the override, NOT the auto-detection result)
- JSON `.rationale.matched_rule == "spec_override"`

**Verification command**:
```bash
moai harness route --spec SPEC-TEST-OVERRIDE-001 --json | \
  jq -e '.level == "thorough" and .rationale.matched_rule == "spec_override"'
```

**Evidence artifact**: stdout JSON; jq exit 0

---

## AC-HRN-001-10 — Effort mapping correctness

**REQ coverage**: REQ-HRN-001-005 (EffortForLevel)

**Given**:
- Shipping `.moai/config/sections/harness.yaml` with `effort_mapping.minimal: medium`, `standard: high`, `thorough: xhigh`

**When**:
```bash
go test -run TestEffortForLevel ./internal/harness/router/...
```

**Then**:
- Exit code 0
- Table-driven assertions:
  - `EffortForLevel(LevelMinimal, cfg) == "medium"`
  - `EffortForLevel(LevelStandard, cfg) == "high"`
  - `EffortForLevel(LevelThorough, cfg) == "xhigh"`
- Additionally verified via CLI: `moai harness route --spec X --json | jq -r .effort` returns the correct value for X's resolved level

**Verification command**:
```bash
go test -run TestEffortForLevel -v ./internal/harness/router/... | grep -q "PASS"
```

**Evidence artifact**: test PASS line; CLI JSON `.effort` field matches the level's mapping

---

## Definition of Done (DoD)

All 10 ACs above MUST be PASS. Additionally:

- [ ] `go test -race -count=1 ./internal/config/... ./internal/harness/router/... ./internal/cli/cmd/...` exits 0
- [ ] `go test -cover ./internal/harness/router/...` reports ≥ 85.0%
- [ ] `golangci-lint run ./internal/config/... ./internal/harness/router/... ./internal/cli/cmd/...` reports 0 ERROR
- [ ] `moai spec lint --strict` reports 0 ERROR (frontmatter `harness_level:` optional field tolerated)
- [ ] `moai harness validate` on the shipping harness.yaml exits 0
- [ ] CI guard `internal/cli/harness_retirement_test.go` `TestHarnessRetirement` PASS (retired factory remains unregistered)
- [ ] CHANGELOG.md `[Unreleased]` section contains a SPEC-V3R2-HRN-001 entry
- [ ] All new exported symbols have `@MX:NOTE` or `@MX:ANCHOR` tags per `mx-tag-protocol.md`
- [ ] `make preflight` succeeds locally (lefthook pre-push hook will block otherwise)
- [ ] `progress.md` updated with `acceptance_phase_status: all-pass` marker

---

## Edge Cases & Negative Tests

The following edge cases are verified via the test suite but are NOT primary ACs:

1. **Missing harness.yaml**: `LoadHarnessConfig("/nonexistent")` returns `ErrConfigNotFound` wrapped in `%w`.
2. **Empty SPEC frontmatter**: Router returns `LevelStandard` (fallthrough default per REQ-007) with `rationale.matched_rule == "fallthrough_default"`.
3. **mode_defaults.cg = thorough regression**: When `--mode cg` is detected, force `LevelThorough` regardless of auto-detection.
4. **Model upgrade reminder (REQ-016)**: When `cfg.ModelUpgradeReview.Trigger.OnModelChange == true` and `os.Getenv("CLAUDE_MODEL_PREVIOUS")` differs from current, CLI emits reminder pointing to `cfg.ModelUpgradeReview.Output.ReportPath`.
5. **Concurrent Reload safety**: `ConfigManager.Reload()` calling `LoadHarnessConfig()` is safe; verified via `go test -race`.
6. **Lenient profile floor check**: At HEAD, `.moai/config/evaluator-profiles/lenient.md` is verified to have all anchor levels ≥ 0.60; M1 pre-flight check.
7. **harness_level: invalid value rejection**: SPEC frontmatter with `harness_level: extreme` produces `ErrUnknownLevel` from spec parser side (out-of-scope for HRN-001 if `harness_level:` validation lives in SPC-001 lint rule; covered as regression).

---

## Bypassed / Deferred Scenarios

Per `acceptance.md` contract with `spec-workflow.md` § Phase 0.5 Plan Audit Gate, the following scenarios are explicitly deferred:

- Sprint Contract artifact emission — DEFERRED to HRN-002
- Per-iteration evaluator memory reset — DEFERRED to HRN-002
- Hierarchical 4-dim × sub-criteria scoring — DEFERRED to HRN-003
- Telemetry export of routing decisions — DEFERRED to beta.1 follow-up SPEC

These deferrals are recorded in `spec.md` §1.2 Non-Goals and §9.2 Blocks.

---

End of acceptance.md.
