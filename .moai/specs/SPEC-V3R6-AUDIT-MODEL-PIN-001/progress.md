# progress.md — SPEC-V3R6-AUDIT-MODEL-PIN-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-24
- plan-audit verdict: PASS, score 1.0 (iter 2/2; N1-N3 findings folded into
  spec/plan/acceptance revision 1.1.0 — MF1 schema relocation to workflow.audit,
  MF2 codex_task isolation, MF3 real symmetry guard, MF4 single-reading effort
  vocabulary, MF5 numeric live-gate rule, MF6 SKIP semantics)
- Implementation Kickoff Approval: granted by lead kickoff 2026-08-24 (run-phase
  entry operator-approved)

## §E.2 Run-phase Evidence

### M1 — Config schema: AuditConfig extension + load helper + template block

Branch `WT-audit-model-pin`, base HEAD `63e10bc1b`. TDD cycle (RED → GREEN).

**RED evidence (E8)** — `TestAuditConfigYAMLRoundTrip` authored BEFORE the
struct extension; verbatim failing output:

```
$ go test ./internal/config/ -run TestAuditConfigYAML -count=1
internal/config/audit_models_test.go:48:66: populated.Codex undefined (type AuditConfig has no field or method Codex)
internal/config/audit_models_test.go:50:10: back.GLM undefined (type AuditConfig has no field or method GLM)
...
too many errors
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
```

`TestLoadWorkflowAuditSection` (internal/cli/audit_pin_test.go) likewise RED
(`undefined: loadWorkflowAuditSection`, build failed) before audit_pin.go.

**GREEN + verification:**

```
$ go test ./internal/config/ -run TestAuditConfigYAML -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	0.460s
$ go test ./internal/cli/ -run TestLoadWorkflowAuditSection -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.131s
$ go test ./internal/config/ -count=1            (full package — incl. TestShippedConfigKeysHaveReaders)
ok  	github.com/modu-ai/moai-adk/internal/config	5.837s
$ go vet ./internal/config/... ./internal/cli/...
exit 0 (no findings)
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	99.814s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	1.310s
?  	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
$ make build
catalog.yaml updated successfully (12899 bytes) — binary rebuilt
exit 0
```

**Changed files (M1):**

- `internal/config/audit_models.go` — `AuditConfig` + `Codex`/`GLM`
  `config.ModelEffort` pin fields (REQ-AMP-001)
- `internal/config/audit_models_test.go` — NEW:
  `TestAuditConfigYAMLRoundTrip` (the AC-AMP-001 dedicated drift guard:
  marshal→unmarshal field-for-field equality + codex:/glm: key-drop arms) +
  `TestAuditConfigYAMLWorkflowWrapperLoad` (workflow: wrapper load contract)
- `internal/cli/audit_pin.go` — NEW: `loadWorkflowAuditSection(projectDir)`
  section-only loader (absent file → zero value, no error) +
  `workflowAuditPins` fail-open wrapper (N3)
- `internal/cli/audit_pin_test.go` — NEW: 4 subtests (populated verbatim /
  absent file / audit-less file / unparseable → error the caller fails open on)
- `internal/template/templates/.moai/config/sections/workflow.yaml` — NEW
  `audit:` block, EMPTY codex/glm sub-keys + precedence/vocabulary comment
  (REQ-AMP-005; `grep -rn "gpt-5.6-sol" internal/template/templates/` = 0
  matches)
- `internal/config/testdata/shipped_key_inventory.yaml` — 4 new
  `workflow.audit.{codex,glm}.{model,effort}` entries (class W, evidence
  reader), header count corrected to the measured 963

**AC trace:** AC-AMP-001 (round-trip guard + loader verbatim pins) — PASS at
M1 scope; AC-AMP-005 template half — PASS (empty defaults + leak-strict green).

### M2 — Codex audit-scoped resolver precedence + task isolation

**RED evidence (E8)** — `internal/cli/mcp_codex_audit_pin_test.go` authored
BEFORE the resolver/seam change; verbatim failing output:

```
$ go test ./internal/cli/ -run TestCodexAuditPin -count=1
internal/cli/mcp_codex_audit_pin_test.go:135:15: undefined: runCodexAuditReviewRPC
internal/cli/mcp_codex_audit_pin_test.go:163:15: undefined: runCodexAuditReviewRPC
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**GREEN + verification:**

```
$ go test ./internal/cli/ -run "TestCodexAuditPin|TestCodexSession|TestMCPAudit|TestLoadWorkflowAuditSection" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.569s
$ go vet ./internal/cli/
exit 0
$ go test ./internal/cli/ -run "TestCodex|TestMCP|TestGLM|TestMultiAudit|TestConvergence|TestAudit" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	9.770s
```

**Mechanism (design decision D2 — injected resolver, not shared-body read):**

- `resolveCodexAuditModelEffort` (mcp_codex.go) — pin-first: `workflowAuditPins`
  .Codex applies only when Model non-empty AND `codexServableModel` holds;
  explicit caller `model` still outranks the pinned model; everything else
  falls through to the untouched legacy `resolveCodexModelEffort`.
- `openCodexSessionResolved` (new core) accepts the injected resolver;
  `openCodexSessionOn` keeps its signature and passes nil — **codex_task.go is
  UNTOUCHED by M2** (isolation is structural: the task path cannot see the pin).
- `codexSessionHandle.resolveME` carries the resolver to
  `buildCodexReviewParams` (4th param) so BOTH seams resolve through it:
  `thread/start` model + `turn/start` model+effort.
- `runCodexReviewRPCResolved` core; `runCodexReviewRPC` (legacy, gate +
  existing tests) and `runCodexAuditReviewRPC` (pin-aware) variants;
  `var codexReviewRPC = runCodexAuditReviewRPC` — both audit callers
  (handleCodexAudit mcp_codex.go:1227, performCodexAudit mcp_convergence.go:425)
  get the pin; the Stop-hook review gate keeps the legacy resolution.

**Changed files (M2):** `internal/cli/mcp_codex.go` (resolver + seams),
`internal/cli/mcp_codex_audit_pin_test.go` (NEW — 6 tests).

**Test inventory (mcp_codex_audit_pin_test.go):**
`TestCodexAuditPin_ReachesTransmittedParams` (AC-AMP-002 positive: both seams
carry gpt-5.6-sol/high), `TestCodexAuditPin_TaskTurnCarriesNoPin` (MF2/REQ-AMP-008
negative: pin present + codex_task turn → SSOT pair, no pin),
`TestCodexAuditPin_UnservableFallsBackToSSOT` (AC-AMP-003 state c),
`TestCodexAuditPin_EffortAlonePinsNothing` (§D.2 edge),
`TestCodexAuditPin_ExplicitModelOverridesPin`, `TestCodexAuditPin_AbsentPinEqualsLegacyResolution`
(AC-AMP-003 states a/b, C7 anchor; the unmodified TestCodexSession_* /
TestMCPAudit_NoDirectFrontmatterRead suite stayed green in the same run).

**AC trace:** AC-AMP-002 — PASS (both arms incl. MF2 regression);
AC-AMP-003 — PASS (states a/b/c + anchors green).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase completion>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
