# Acceptance: SPEC-ADVISOR-001 (Advisor Strategy Adoption)

> Companion to `spec.md`. Detailed Given-When-Then scenarios, edge cases, quality gates.

---

## Scenario 1 — Verification Probe Resolves Native Support (REQ-ADV-010, AC-1)

**Given** the Claude Code runtime is at v2.1.111 or later
**And** a probe agent at `.claude/agents/moai/.scratch/advisor-probe.md` declares `tools: Read, advisor_20260301`
**When** the orchestrator spawns the probe agent via `Agent(subagent_type: "advisor-probe")`
**Then** the runtime SHALL either accept the tool declaration and return a measurable advisor response, OR reject it with a tool-not-found error.
**And** the outcome SHALL be recorded in `.moai/reports/advisor-probe-{DATE}/result.md` with a verdict field of one of `{native_supported, native_rejected, ambiguous}`.

**Edge case 1a**: Runtime accepts the tool declaration silently but advisor invocation returns empty content. The probe SHALL classify this as `ambiguous` and select the fallback path conservatively.

**Edge case 1b**: Runtime version is below v2.1.111. The probe SHALL skip and record `version_below_minimum`. SPEC implementation halts pending Claude Code upgrade.

---

## Scenario 2 — manager-spec Pilot, Native Path Success (REQ-ADV-003, REQ-ADV-007, AC-2, AC-4)

**Given** the verification probe (Scenario 1) recorded `native_supported`
**And** `manager-spec.md` frontmatter declares `model: sonnet, effort: high, advisor: {model: opus, max_uses: 3, trigger_class: spec_decisions}`
**And** the user requests `/moai plan "user authentication system with OAuth2 and session management"`
**When** the orchestrator delegates to manager-spec
**Then** the agent SHALL produce a SPEC directory `.moai/specs/SPEC-AUTH-NNN/` containing valid spec.md, plan.md, acceptance.md.
**And** the spec.md SHALL pass plan-auditor frontmatter schema validation (9 required fields, valid types).
**And** the telemetry log entry for this invocation SHALL contain `executor_model: sonnet`, `advisor_model: opus`, `advisor_invocations >= 0`, `advisor_invocations <= 3`.
**And** when the same prompt is replayed 5 times against this configuration, **all 5 attempts** SHALL produce schema-valid SPECs.

**Edge case 2a**: Advisor returns malformed guidance (truncated JSON, error response). The executor SHALL log `advisor_response: malformed`, fall back to its own primary reasoning, and still produce a valid SPEC.

**Edge case 2b**: User prompt contains genuinely ambiguous EARS pattern (e.g., a requirement that could be Event-Driven OR State-Driven). Advisor SHALL be consulted exactly once for this decision, and the advisor's choice SHALL be reflected in the final spec.md (verifiable by EARS keyword inspection).

**Edge case 2c**: max_uses cap reached (advisor consulted 3 times). The 4th consultation attempt SHALL be rejected by the runtime; agent SHALL emit `advisor_cap_reached: true` in completion summary; agent SHALL still complete successfully.

---

## Scenario 3 — manager-spec Pilot, Fallback Path Success (REQ-ADV-013, AC-6)

**Given** the verification probe recorded `native_rejected` or `ambiguous`
**And** `manager-spec.md` body documents the fallback Agent() consultation pattern
**When** the executor encounters an EARS-pattern ambiguity decision
**Then** the executor SHALL invoke `Agent(subagent_type: "researcher", model: "opus", mode: "plan", description: "advisor consultation: EARS pattern selection for requirement <id>")`.
**And** the spawned researcher SHALL return structured guidance (recommendation + rationale) without writing any files.
**And** the executor SHALL incorporate the recommendation into the spec.md.
**And** the telemetry log SHALL record `advisor_pattern: fallback_agent`, `advisor_invocations: 1`, with `advisor_tokens` populated from the spawned-agent's token usage.

**Edge case 3a**: Spawned researcher fails (rate limit, network error). Executor SHALL log `advisor_fallback: failed`, proceed with its own reasoning, and still complete the SPEC.

**Edge case 3b**: Spawned researcher writes a file despite mode: plan. This SHALL trigger a hard violation report — the researcher's output is discarded and a feedback note is filed. Executor proceeds with primary reasoning.

---

## Scenario 4 — evaluator-active Dimension Score Variance (REQ-ADV-017, AC-3)

**Given** evaluator-active baseline runs (5 evaluations, no advisor) on SPEC artifacts S1..S5 produce dimension scores `B[i][d]` for SPEC i and dimension d
**And** evaluator-active pilot runs (5 evaluations, with advisor `max_uses: 2`) on the same artifacts produce dimension scores `P[i][d]`
**When** variance per dimension is computed as `var(d) = max_i |B[i][d] - P[i][d]|`
**Then** `var(d) <= 0.10` SHALL hold for **every** dimension d in {Functionality, Security, Craft, Consistency}.
**And** if `var(d) > 0.10` for any dimension, the pilot SHALL be flagged for rollback per REQ-ADV-017.
**And** the variance computation SHALL be recorded in `.moai/reports/advisor-pilot-results-{DATE}/variance.md`.

**Edge case 4a**: Sample SPECs do not stress all four dimensions (e.g., all 5 are documentation-only with trivial Security scores). The variance is computed only on stressed dimensions; unstressed dimensions are reported as "insufficient signal." Final pilot acceptance requires variance compliance on all stressed dimensions.

**Edge case 4b**: Advisor consultation in pilot leads to a SCORE INCREASE in some dimensions (e.g., Security goes from 0.50 to 0.75 because advisor catches a subtlety baseline missed). This is the Anthropic-claimed quality improvement; var(Security) > 0.10 in this direction is a SIGNAL OF SUCCESS, not failure. The variance threshold applies in the regression direction (P < B). Document this asymmetry in `.moai/reports/advisor-pilot-results-{DATE}/variance.md`.

---

## Scenario 5 — Cost Reduction Validation (REQ-ADV-014, AC-4)

**Given** 10 baseline manager-spec invocations on representative SPEC requests R1..R10 with `model: opus, effort: xhigh` (pre-pilot configuration), produce total Opus token consumption `T_baseline`
**And** 10 pilot manager-spec invocations on the same requests with the new configuration, produce total Opus token consumption `T_pilot` = `executor_tokens (Sonnet, not Opus) + advisor_tokens (Opus)`
**When** the cost ratio is computed
**Then** `1 - (advisor_tokens / T_baseline) >= 0.30` SHALL hold (i.e., Opus token consumption reduced by at least 30%).
**And** the result SHALL be recorded in `.moai/reports/advisor-pilot-results-{DATE}/cost-reduction.md`.

**Edge case 5a**: A specific request requires unusually many advisor consultations (e.g., complex multi-phase SPEC) and `advisor_tokens` for that request alone exceeds 50% of `T_baseline / 10`. The aggregate threshold may still hold; document outliers. If aggregate fails, escalate to user — do not auto-rollback.

**Edge case 5b**: KV-cache hits cause `advisor_tokens` to undercount actual cost. The pilot uses the API-reported `usage` field; if Anthropic API splits this into `cache_read_tokens` vs `cache_creation_tokens`, both contribute to the cost figure. Document methodology in cost-reduction.md.

---

## Scenario 6 — Telemetry Schema and Privacy (REQ-ADV-014, REQ-ADV-015, AC-7)

**Given** telemetry is enabled and an agent invocation completes with advisor consultation
**When** the post-tool hook fires
**Then** the resulting `.moai/observability/hook-metrics.jsonl` line SHALL be valid JSON containing the fields: `executor_model`, `executor_tokens`, `advisor_model`, `advisor_invocations`, `advisor_tokens`, `advisor_latency_ms_total`, `advisor_cap_reached`.
**And** the line SHALL NOT contain any of the following fields: `advisor_prompt`, `advisor_response`, `api_key`, `Authorization`, `prompt_text`, `guidance_text`.
**And** a unit test in `internal/hook/post_tool_advisor_test.go` SHALL parse the line, assert presence of required fields, and assert absence of forbidden fields.

**Edge case 6a**: A bug introduces inadvertent serialization of advisor response into a misnamed field (e.g., `details: "<full advisor response>"`). The unit test SHALL include a "long string field heuristic": any string field longer than 500 chars in the telemetry line is flagged for manual review.

**Edge case 6b**: Telemetry write fails (disk full, permissions). The agent SHALL still complete the user task; telemetry write failure is logged to stderr but does not block the agent.

---

## Scenario 7 — Cap Enforcement under Stress (REQ-ADV-005, AC-5)

**Given** manager-spec is configured with `advisor.max_uses: 3`
**And** the executor is running on a synthetically pathological SPEC request that would naturally trigger 5 advisor consultations
**When** the agent runs to completion
**Then** the agent SHALL emit exactly 3 advisor consultations (the first 3 trigger conditions).
**And** the 4th and 5th attempted consultations SHALL be rejected by the runtime guard.
**And** the agent's completion summary SHALL include `advisor_cap_reached: true`.
**And** the produced SPEC SHALL still be schema-valid (executor falls through to primary reasoning for the 4th and 5th decisions).

---

## Scenario 8 — Rollback Procedure (REQ-ADV-017)

**Given** the variance threshold (Scenario 4) is violated AND a rollback decision is made
**When** the user confirms rollback via AskUserQuestion
**Then** `manager-spec.md` and/or `evaluator-active.md` frontmatter SHALL be reverted to its pre-pilot state.
**And** a regression report SHALL be filed at `.moai/reports/advisor-regression-{DATE}/` containing: which agent was rolled back, the failing variance dimension, the sample data, and a hypothesis for follow-up.
**And** the SPEC status in `.moai/specs/SPEC-ADVISOR-001/spec.md` frontmatter SHALL be updated from `approved` to one of `{partially_approved, superseded}` per the rollback scope.

---

## Quality Gates (TRUST 5 Mapping)

- **Tested**: Scenarios 1, 2, 3, 5, 6, 7 each have automated or semi-automated verification (probe artifact, schema test, fault injection, telemetry parser test, cap test). Scenario 4 has a defined statistical procedure but acknowledged small-sample limitations (Risk R4).
- **Readable**: Frontmatter `advisor:` block schema is documented in `.claude/rules/moai/development/model-policy.md`; agent-authoring guide cross-references it.
- **Unified**: Both pilot agents adopt the same frontmatter schema and same telemetry fields; protocol is shared.
- **Secured**: REQ-ADV-015 (no raw response logging) is hard-tested in Scenario 6. REQ-ADV-016 (advisor cannot override acceptance criteria) is structural and audited via Scenario 4 (advisor influence is bounded by variance threshold).
- **Trackable**: Per-invocation telemetry (Scenario 6) provides full audit trail. Probe artifact (Scenario 1) and pilot results (Scenarios 4 + 5) are stored under `.moai/reports/`.

---

## Definition of Done

- [ ] M0 verification probe executed; result.md filed in reports
- [ ] M1 protocol documentation merged
- [ ] M2 telemetry schema extended; unit test passes
- [ ] M3 manager-spec migrated (native or fallback path) and 5 sample SPECs produced
- [ ] M4 evaluator-active augmented; 5 baseline + 5 pilot evaluations completed
- [ ] M5 cap and fallback fault-injection tests pass
- [ ] M6 cost-reduction report shows >= 30% Opus token reduction
- [ ] All scenarios 1-8 satisfied (or rollback per Scenario 8 documented)
- [ ] plan-auditor verdict on this SPEC: PASS
- [ ] User approval (AP-4) on pilot outcomes

---

**Status**: draft — pending plan-auditor review and pilot execution
