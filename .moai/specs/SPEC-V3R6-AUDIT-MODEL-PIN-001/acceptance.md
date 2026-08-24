# acceptance.md — SPEC-V3R6-AUDIT-MODEL-PIN-001

Verification layer. Every AC is binary-testable and cites its owning requirement.
GEARS obligations live in spec.md §D; the entries here are Given-When-Then
scenarios. The live-delivery ACs (006/007) are the run-phase evidence gates —
config-read assertions never substitute for them
(`verification-claim-integrity.md` §1.1).

> Revision 1.1.0 — MF3 (real symmetry guard), MF4 (single-reading effort rule),
> MF5 (numeric decision rule), MF6 (SKIP semantics) applied; schema anchors
> relocated to `workflow.audit` per MF1/lead ruling C.

## §D AC Matrix

### AC-AMP-001 — AuditConfig extension lands with a REAL drift guard (REQ-AMP-001) [MUST]
- **Given** the tree with M1 applied, **When** a workflow.yaml fixture containing a
  populated `audit.codex`/`audit.glm` block is loaded AND the dedicated
  `AuditConfig` round-trip unit test runs (N2 mechanism: e.g.
  `TestAuditConfigYAMLRoundTrip` in `internal/config/audit_models_test.go` —
  marshal a fully-populated AuditConfig → unmarshal → field-for-field equality;
  chosen over a symmetryCases extension because `checkSymmetry` navigates one
  level and cannot reach the nested `workflow.audit` block), **Then** the loader
  returns the pinned `{model, effort}` pairs verbatim AND the round-trip test
  passes; removing either new struct field or either new yaml key makes that
  test FAIL (the MF3 requirement: the cited guard must actually exercise the
  new schema).

### AC-AMP-002 — codex pin reaches both transmission seams, audit path only (REQ-AMP-002, REQ-AMP-008) [MUST]
- **Given** a project dir whose workflow.yaml sets
  `audit.codex = {model: gpt-5.6-sol, effort: high}`, **When** the codex AUDIT
  session opens and a review turn starts (seam-captured test:
  `TestCodexSession_ResolvedModelReachesTransmittedParams` pattern extended with
  the audit-pin fixture), **Then** `thread/start` params carry
  `"model": "gpt-5.6-sol"` and `turn/start` params carry both
  `"model": "gpt-5.6-sol"` and `"effort": "high"` — AND the MF2 regression arm:
  a `codex_task` turn under the SAME config carries NO pinned model/effort
  (legacy SSOT-only resolution).

### AC-AMP-003 — empty/unservable pin = legacy behavior (REQ-AMP-004) [MUST]
- **Given** (a) a workflow.yaml with no `audit.codex`/`audit.glm` sub-keys, (b)
  sub-keys present but empty, and (c) `audit.codex.model: opus`, **When** the
  audit resolvers run in each state, **Then** the resolved `config.ModelEffort`
  and the transmitted params are identical to the pre-SPEC tree in all three
  states (state (c) falls back through the SSOT path — the servability filter
  holds).
- Regression anchors: `TestMCPAudit_NoDirectFrontmatterRead`,
  `TestCodexSession_ResolvedModelReachesTransmittedParams` unmodified and green.

### AC-AMP-004 — GLM pin reaches the request body under the single-reading rule (REQ-AMP-003, REQ-AMP-006, REQ-AMP-008) [MUST]
- **Given** workflow.yaml sets `audit.glm = {model: glm-5.3, effort: max}`,
  **When** `glm_audit` (or the `audit_multi` GLM leg) builds its request
  (glmHTTPClient seam capture), **Then** the outbound body carries
  `"model": "glm-5.3"` and the reasoning directive set to `max` on the delivery
  field selected by the live evidence.
- **Given** `audit.glm.effort: medium` (a Claude-only-vocabulary value), **When**
  the request is built, **Then** the body carries the pinned model but NO
  reasoning directive (REQ-AMP-006 single-reading rule: valid set is exactly
  `{low, high, max}`; no collapse runs on this path).
- **Given** no sub-keys / empty sub-keys, **When** the request is built, **Then**
  the body contains no reasoning field and the model resolves exactly as before;
  a test additionally asserts `glm_task` resolution is unaffected by a populated
  `audit.glm` (REQ-AMP-008).

### AC-AMP-005 — template + Go defaults stay empty (REQ-AMP-005) [MUST]
- **Given** the merged tree, **When** `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`
  and `grep -rn "gpt-5.6-sol" internal/template/templates/` run, **Then** both
  pass/return zero matches, and the template workflow.yaml's `audit.codex` /
  `audit.glm` sub-keys carry only empty-string defaults.

### AC-AMP-006 — live GLM reasoning-delivery proof, numeric rule (REQ-AMP-006, REQ-AMP-007) [MUST — closes AC-MTP-032b]
- **Given** a live z.ai credential (`~/.moai/.env.glm`) and ONE fixed reviewable
  diff target in this tree, **When** two live `glm_audit` calls run — one with
  `audit.glm.effort: max`, one with `audit.glm.effort: low` — and both raw z.ai
  response envelopes are captured, **Then** the decision rule is:
  - Define the reasoning signal **S(run)** = the envelope's usage reasoning-token
    count when the field exists; otherwise the total byte count of
    thinking/reasoning content blocks in the captured envelope.
  - Both runs must return a non-`inconclusive` verdict (an inconclusive run is
    evidence-invalid, not a pass).
  - **PASS ⇔ S(max-run) ≥ 1.25 × max(S(low-run), 1).** *(amended 2026-08-24 —
    see the dated record below; original threshold was 2.0)*
  - **Discriminant clause (amendment validity)**: a known-ignored delivery field
    (the null control) must measure **< 1.1** on the same differential — the
    lowered threshold is valid only because the ignored-field control separates
    delivery from noise (measured null: the hypothesis-A thinking-budget run
    measured 1.02).
  - **FAIL ⇔ the ratio is below 1.25** (or the null control measures ≥ 1.1) —
    the current delivery field is ignored by the endpoint; switch the field,
    re-run, and record which field the endpoint honors.
- **SKIP semantics (MF6)**: `GLM_API_KEY` absent ⇒ this AC reports SKIP — an
  explicit `SKIP: GLM credential absent` marker line in the evidence file under
  `.moai/state/verify/<session>/`; a SKIP blocks the AC from being counted PASS
  and is surfaced as unresolved in the §E report (never FAIL, never silent pass).
- Evidence persisted with command + verbatim output; the verdict names which
  hypothesis the live evidence selected.

**Amendment 2026-08-24 (lead-approved)** — threshold 2.0 → 1.25 with
discriminant. The original 2.0 was an a-priori guess at the backend's dynamic
range. Measured range across 3 calibrated runs: 1.34 / 1.85 / 1.48
(output-token ratio 1.40, consistent). Delivery was PROVEN by hypothesis B
(top-level `reasoning_effort`) against the hypothesis-A null (1.02) — the
measured reversal: the thinking-budget object is IGNORED by z.ai, and the
top-level `reasoning_effort` field is the effective delivery field (the
template llm.yaml overlay doc correction is a follow-up card's scope —
reference only). Full 4-run evidence:
`.moai/state/verify/t225/ac-amp-006-glm-differential-attempt{1..4}.md`.

### AC-AMP-007 — live codex pin confirmation (REQ-AMP-002) [MUST]
- **Given** this project's tracked workflow.yaml pin
  (`audit.codex = {model: gpt-5.6-sol, effort: high}`), **When** one real
  `codex_audit` / `mcp__moai__codex_audit` call runs against this tree, **Then**
  observed evidence (transmitted-params log or the codex session's resolved
  model report) shows `gpt-5.6-sol` + `high` carried by the moai-side request —
  the model choice no longer depends on `~/.codex/config.toml`.
- **SKIP semantics (MF6)**: codex binary or auth absent (per the existing
  `codex_setup` probe) ⇒ SKIP with the same marker protocol; never counts as
  PASS.

### AC-AMP-008 — web exposure round-trip in the existing Audit panel (REQ-AMP-009) [SHOULD]
- **Given** the web console with M4 applied, **When** the existing Audit panel
  (the surface already rendering `workflow.audit.*`, `schemaform.go:52`) renders
  and the four new fields are saved with values and reloaded, **Then** each field
  shows its persisted value (typed-path round-trip test), the glm-effort select
  offers exactly `{low, high, max}`, all four locale dictionaries carry the new
  keys with no untranslated-value governance violation, and the effort fields'
  help text marks them runtime-applied (contrast with the stored-only tier
  efforts, REQ-WCR-033 labeling discipline).

## §D.1 Severity mapping

| AC | Severity | owning REQ | milestone |
|---|---|---|---|
| AC-AMP-001 | MUST | 001 | M1 |
| AC-AMP-002 | MUST | 002, 008 | M2 |
| AC-AMP-003 | MUST | 004 | M2 |
| AC-AMP-004 | MUST | 003, 006, 008 | M3 |
| AC-AMP-005 | MUST | 005 | M1/M5 |
| AC-AMP-006 | MUST | 006, 007 | M5 |
| AC-AMP-007 | MUST | 002 | M5 |
| AC-AMP-008 | SHOULD | 009 | M4 |

## §D.2 Edge cases

- workflow.yaml exists but fails to load (unparseable / unreadable) → the
  resolver treats the load error as an ABSENT pin and fails open to the legacy
  SSOT path — never a hard error (N3); the helper returns the zero AuditConfig
  on an absent file by the same rule.
- `audit.codex.effort` set with empty model → the pin is treated as absent
  (model is the gate; effort alone pins nothing).
- `audit.glm.model` set to a non-GLM id → sent verbatim; z.ai 4xx degrades via
  the existing fail-open `VerdictInconclusive` (observable in the summary, no
  hard error).
- `audit.glm.effort` outside `{low, high, max}` → reasoning directive omitted,
  model pin still applies (REQ-AMP-006; the web select makes this reachable only
  via hand-edited yaml).
- Web save with empty values → persists empty strings → resolver falls back
  (indistinguishable from absent sub-keys).

## §D.3 Indirect verification

Where a direct assertion is impossible, the AC relies on the named seam tests
(glmHTTPClient / codex session param capture) — the same injectable-seam pattern
the M2/M3 backends already use; no new harness is introduced. The MF2 isolation
is verified negatively: the codex_task seam test asserts the pin's ABSENCE from
transmitted params under a populated pin.

## §D.4 Closure gates (Definition of Done)

- All MUST ACs PASS with verbatim command + output recorded in progress.md §E;
  any AC in SKIP state (MF6 marker present) keeps the SPEC from closing as
  completed until re-run with the backend available or explicitly waived by the
  lead.
- AC-AMP-006's verdict names the honored delivery field explicitly and shows the
  computed S(max), S(low), and ratio.
- Companion statusline patch files show no diff from this SPEC's commits.
- CI green on the PR head (full-suite verdict per repo discipline).

## §D.5 Forward-looking checks

- If z.ai later honors a native `reasoning_effort` passthrough, the delivery
  field switch is a one-seam change (`callGLMAudit`) — no schema churn.
- The `AuditConfig` extension is the natural home for future per-backend audit
  knobs (e.g. timeout overrides); extending it must not break the empty-fallback
  contract (REQ-AMP-004) or the symmetry guard added by AC-AMP-001.
