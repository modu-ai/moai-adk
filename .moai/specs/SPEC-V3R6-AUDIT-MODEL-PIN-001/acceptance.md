# acceptance.md — SPEC-V3R6-AUDIT-MODEL-PIN-001

Verification layer. Every AC is binary-testable and cites its owning requirement.
GEARS obligations live in spec.md §D; the entries here are Given-When-Then
scenarios. The live-delivery ACs (006/007) are the run-phase evidence gates —
config-read assertions never substitute for them
(`verification-claim-integrity.md` §1.1).

## §D AC Matrix

### AC-AMP-001 — llm.audit schema lands symmetric (REQ-AMP-001) [MUST]
- **Given** the tree with M1 applied, **When**
  `go test ./internal/config/ -run 'TestStructYAMLSymmetry'` runs and a test loads
  a fixture llm.yaml containing a populated `audit:` block, **Then** the symmetry
  guard passes and `loadLLMSectionOnly` returns `Audit.Codex`/`Audit.GLM` carrying
  the fixture `{model, effort}` values verbatim.

### AC-AMP-002 — codex pin reaches both transmission seams (REQ-AMP-002) [MUST]
- **Given** a project dir whose llm.yaml sets `audit.codex = {model: gpt-5.6-sol, effort: high}`,
  **When** the codex session opens and a review turn starts (seam-captured test:
  `TestCodexSession_ResolvedModelReachesTransmittedParams` pattern extended with
  the audit-section fixture), **Then** `thread/start` params carry
  `"model": "gpt-5.6-sol"` and `turn/start` params carry both
  `"model": "gpt-5.6-sol"` and `"effort": "high"`.

### AC-AMP-003 — empty/unservable section = legacy behavior (REQ-AMP-004) [MUST]
- **Given** (a) no llm.yaml, (b) an llm.yaml with an empty `audit:` block, and
  (c) an llm.yaml with `audit.codex.model: opus`, **When** the codex resolver
  runs in each state, **Then** the resolved `config.ModelEffort` and the
  transmitted params are identical to the pre-SPEC tree in all three states
  (state (c) falls back through the SSOT path — the servability filter holds).
- Regression anchors: `TestMCPAudit_noDirectFrontmatterRead`,
  `TestCodexSession_ResolvedModelReachesTransmittedParams` unmodified and green.

### AC-AMP-004 — GLM pin reaches the request body (REQ-AMP-003, REQ-AMP-006) [MUST]
- **Given** llm.yaml sets `audit.glm = {model: glm-5.3, effort: max}`, **When**
  `glm_audit` (or the `audit_multi` GLM leg) builds its request (glmHTTPClient
  seam capture), **Then** the outbound body carries `"model": "glm-5.3"` and the
  reasoning directive on the delivery field selected by the live evidence, and a
  Claude-vocabulary effort value arriving in `audit.glm.effort` collapses through
  `CollapseClaudeEffortToGLM` rather than being sent raw.
- **Given** no llm.yaml / empty `audit:`, **When** the request is built, **Then**
  the body contains no reasoning field and the model resolves exactly as before
  (legacy path); a test additionally asserts `glm_task` resolution is unaffected
  by a populated `audit.glm` (REQ-AMP-008).

### AC-AMP-005 — template + Go defaults stay empty (REQ-AMP-005) [MUST]
- **Given** the merged tree, **When** `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`
  and `grep -rn "gpt-5.6-sol" internal/template/templates/` run, **Then** both
  pass/return zero matches, and the template llm.yaml's `audit:` block carries
  only empty-string defaults.

### AC-AMP-006 — live GLM reasoning-delivery proof (REQ-AMP-006, REQ-AMP-007) [MUST — closes AC-MTP-032b]
- **Given** a live z.ai credential (`~/.moai/.env.glm`) and one fixed reviewable
  diff target in this tree, **When** two live `glm_audit` calls run — one with
  `audit.glm.effort: max`, one with `audit.glm.effort: low` — and the raw z.ai
  response envelopes are captured (usage fields or equivalent backend-observed
  reasoning signal), **Then** the observed evidence distinguishes the two
  delivery hypotheses:
  - **Delivery proven**: the max run shows a materially higher reasoning signal
    (reasoning-token count or documented equivalent) than the low run on the
    same diff → the delivery field is correct; PASS.
  - **Delivery absent**: no material delta between max and low → the current
    delivery field is ignored by the endpoint; this is a FAIL of the delivery
    mechanism (not of the test) — switch the field (hypothesis A ↔ B), re-run,
    and record which field the endpoint honors.
- Evidence persisted under `.moai/state/verify/<session>/` with command +
  verbatim output. The verdict names which hypothesis the live evidence
  selected.

### AC-AMP-007 — live codex pin confirmation (REQ-AMP-002) [MUST]
- **Given** this project's local llm.yaml pin
  (`audit.codex = {model: gpt-5.6-sol, effort: high}`), **When** one real
  `codex_audit` / `mcp__moai__codex_audit` call runs against this tree, **Then**
  observed evidence (transmitted-params log or the codex session's resolved
  model report) shows `gpt-5.6-sol` + `high` carried by the moai-side request —
  the model choice no longer depends on `~/.codex/config.toml`.

### AC-AMP-008 — web exposure round-trip (REQ-AMP-009) [SHOULD]
- **Given** the web console with M4 applied, **When** the 3rd Party LLM panel
  renders and the four audit fields are saved with values and reloaded, **Then**
  each field shows its persisted value (typed-path round-trip test), all four
  locale dictionaries carry the new keys with no untranslated-value governance
  violation, and the effort fields' help text marks them runtime-applied.

## §D.1 Severity mapping

| AC | Severity | owning REQ | milestone |
|---|---|---|---|
| AC-AMP-001 | MUST | 001 | M1 |
| AC-AMP-002 | MUST | 002 | M2 |
| AC-AMP-003 | MUST | 004 | M2 |
| AC-AMP-004 | MUST | 003, 006, 008 | M3 |
| AC-AMP-005 | MUST | 005 | M1/M5 |
| AC-AMP-006 | MUST | 006, 007 | M5 |
| AC-AMP-007 | MUST | 002 | M5 |
| AC-AMP-008 | SHOULD | 009 | M4 |

## §D.2 Edge cases

- llm.yaml exists but is unparseable → existing error path (load failure →
  codex: empty pair; GLM: default model) unchanged.
- `audit.codex.effort` set with empty model → section treated as absent (model
  is the gate; effort alone pins nothing).
- `audit.glm.model` set to a non-GLM id → sent verbatim; z.ai 4xx degrades via
  the existing fail-open `VerdictInconclusive` (observable in the summary, no
  hard error).
- Web save with empty values → persists empty strings → resolver falls back
  (indistinguishable from absent section).

## §D.3 Indirect verification

Where a direct assertion is impossible, the AC relies on the named seam tests
(glmHTTPClient / codex session param capture) — the same injectable-seam pattern
the M2/M3 backends already use; no new harness is introduced.

## §D.4 Closure gates (Definition of Done)

- All MUST ACs PASS with verbatim command + output recorded in progress.md §E.
- AC-AMP-006's verdict names the honored delivery field explicitly.
- Companion statusline patch files show no diff from this SPEC's commits.
- CI green on the PR head (full-suite verdict per repo discipline).

## §D.5 Forward-looking checks

- If z.ai later honors a native `reasoning_effort` passthrough, the delivery
  field switch is a one-seam change (`callGLMAudit`) — no schema churn.
- The `llm.audit` section is the natural home for future per-backend audit
  knobs (e.g. timeout overrides); extending it must not break the empty-fallback
  contract (REQ-AMP-004).
