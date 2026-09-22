# t580 — Explicit `required` gate enforcement (GH #1632 item 3) — Verdict

## Claim

1. A gate the project EXPLICITLY configures `required` in `workflow.audit.gates` now fails
   `overall_verdict` when its backend returned no verdict (fail-open `inconclusive`),
   instead of riding the claude-anchor fall-through to `pass`.
2. A satisfied explicit `required` gate (backend produced pass/fail) behaves exactly as
   before: pass passes; a required FAIL blocks on the pre-existing contract.
3. When `required` is NOT set, behavior is preserved byte-for-byte: fail-open to the claude
   anchor, annotate-only — zero change for existing users. The discriminator is the RAW
   workflow.yaml value (the same read `applyGateUnmet` uses at the single-backend surface),
   never the engine's distributed default (codex→required when the key is absent), because
   treating the default as an opt-in would flip every existing project to fail-closed.
4. The enforced verdict reaches the persisted state file (`.moai/state/audit-multi/<session>.json`)
   the multi-review-gate Stop hook reads, so the block survives past the tool call.
5. The enforcement moves ONLY `overall_verdict` + `residual_risk_note`; the backend's own
   `per_backend_verdicts` entry stays `inconclusive` and `fail_open_backends` keeps naming it.

## Evidence

New tests: `internal/cli/required_gate_block_test.go` (8 tests). RED first (4 failures on the
un-enforced tree, including `overall_verdict = "inconclusive"` for an unmet required claude
anchor — a value outside the declared {pass, fail} set), GREEN after implementation.

Final run, all three card branches + guards:

```
$ go test ./internal/cli -count=1 -timeout 9m -run 'TestRunMultiAudit_ExplicitRequiredGateUnmet_FailsOverall|TestRunMultiAudit_ExplicitRequiredGateMet_Passes|TestRunMultiAudit_ExplicitRequiredGateFail_BlockedByExistingContract|TestRunMultiAudit_UnsetGate_KeepsAnnotateOnlyFailOpen|TestRunMultiAudit_ExplicitAdvisoryGateUnmet_StillPasses|TestRunMultiAudit_ExplicitRequiredClaudeAnchorUnmet_FailsOverall|TestRunMultiAudit_ExplicitRequiredGateUnmet_PersistedResultCarriesFail|TestRunMultiAudit_GateIsTreeScoped' -v
--- PASS: TestRunMultiAudit_ExplicitRequiredGateUnmet_FailsOverall (0.00s)
--- PASS: TestRunMultiAudit_ExplicitRequiredGateMet_Passes (0.00s)
--- PASS: TestRunMultiAudit_ExplicitRequiredGateFail_BlockedByExistingContract (0.00s)
--- PASS: TestRunMultiAudit_UnsetGate_KeepsAnnotateOnlyFailOpen (0.00s)
--- PASS: TestRunMultiAudit_ExplicitAdvisoryGateUnmet_StillPasses (0.00s)
--- PASS: TestRunMultiAudit_ExplicitRequiredClaudeAnchorUnmet_FailsOverall (0.00s)
--- PASS: TestRunMultiAudit_ExplicitRequiredGateUnmet_PersistedResultCarriesFail (0.00s)
--- PASS: TestRunMultiAudit_GateIsTreeScoped (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.854s
```

Branch 1 (required: true + unmet → fail): `TestRunMultiAudit_ExplicitRequiredGateUnmet_FailsOverall` —
overall=fail, codex stays `inconclusive` in per_backend_verdicts, fail_open_backends names codex,
note names codex.

Branch 2 (required: true + met → passes): `TestRunMultiAudit_ExplicitRequiredGateMet_Passes` —
overall=pass, no unmet claim. `..._Fail_BlockedByExistingContract` — a real fail blocks via the
pre-existing required-FAIL contract with NO unmet claim added.

Branch 3 (required unset + unmet → annotate-only, still pass):
`TestRunMultiAudit_UnsetGate_KeepsAnnotateOnlyFailOpen` — overall=pass, codex still listed in
fail_open_backends. Explicit advisory + unmet also passes (`..._ExplicitAdvisoryGateUnmet_StillPasses`).

Regression sweep of the affected families (convergence engine, fan-out, single-backend codex
audit incl. the 4 pre-existing gate-unmet annotation tests, multi-review-gate Stop hook,
audit_multi handler, blank-review pins, state-file decoding):

```
$ go test ./internal/cli -count=1 -timeout 9m -run 'TestConverge_|TestRunMultiAudit_|TestCodexAudit_|TestMultiReviewGate_|TestAuditMulti_|TestCodexBlankReview_|TestLoadConvergenceResult' -v   (PASS count)
82
$ go test ./internal/cli -count=1 -timeout 9m -run '...same pattern...'
ok  	github.com/modu-ai/moai-adk/internal/cli	5.371s
```

Static checks:

```
$ go vet ./internal/cli/
VET-CLEAN
$ golangci-lint run ./internal/cli/...   (issues counted from saved output /tmp/t580_lint.log)
0 issues in mcp_convergence.go / multi_review_gate.go / required_gate_block_test.go;
28 pre-existing issues in 8 untouched files (factory_handoff_recover.go, handoff_recover_test.go,
home_state_coverage.go, migrate_home_state.go, migrate_home_state_test.go,
profile_continue_lease_test.go, profile_lease_integration_test.go, profile_lease_launch.go)
```

Template edits (SKILL.md docs + catalog regeneration) — `make build` run, drift checks passed,
binary rebuilt:

```
$ make build   (tail)
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "..." -o bin/moai ./cmd/moai
```

## Baseline-attribution

- Branch `WT-required-gate-block`, worktree `agent-a436b155a2dbc17b5`, base HEAD `04de513e4`
  (develop lineage, clean tree at start — `git status --porcelain` empty before edits).
- All commands above run in THIS worktree against THIS tree, `-count=1` (no cached results).
- RED state measured before implementation (4 failing tests, same tree, same commands).
- The full-package `go test ./internal/cli -count=1` run was killed by go test's DEFAULT 10m
  timeout (667: `panic: test timed out after 10m0s` in the log) with 0 test failures before the
  kill — the package's known >600s floor. The targeted-family runs above are the verdict
  evidence per the card's load discipline; the full-suite judgment stays with CI.

## Gaps

- The full internal/cli suite was NOT observed to completion in one run (default-timeout kill
  described above). Families touched by this change (82 tests) were observed green; CI owns the
  full-suite verdict.
- Other packages were not re-run: no file outside internal/cli/ Go code + 2 skill docs +
  catalog.yaml changed; `go build` of the whole module succeeded inside `make build`, which
  compile-checks every package.
- `codex_review_gate.go` (the sibling Stop-hook gate) was read but not modified: it gates on its
  own `workflow.codex.review_gate.enabled` toggle, not on `audit.gates.*`.
- The wizard's "Required: Fail blocks convergence" description was read and left as-is: it is now
  accurate for wizard-configured (i.e. explicitly written) gates; no wording change needed.
- docs-site content was scanned for `audit.gates` (no hits; only LSP `lsp_quality_gates` on
  unrelated pages) — no docs-site change made.
- Single-backend `codex_audit` verdict semantics unchanged (still returns `inconclusive` +
  `gate_unmet` annotation): the operator decision is implemented at the layer that computes an
  OVERALL verdict (convergence), which is where the card locates the defect.

## Residual-risk

- The enforcement keys on the raw `required` token comparison; a typo'd token (e.g. `Required`)
  silently stays fail-open — the same failure shape `applyGateUnmet` already has at the
  single-backend surface. Consistency chosen over new validation.
- A project that explicitly sets `gates.codex: required` in a tree it does NOT pass as
  `project_root` and whose server cwd resolves elsewhere gets the OTHER tree's gates — the same
  tree-scoped rule the whole audit layer follows; the `gate is tree-scoped` test pins it.
- The persisted-state flip means a Stop-hook session with an explicit required gate and an
  unavailable codex will now BLOCK turn-end where it previously allowed — that is the contract
  change, but it is a behavior change a user with the explicit config will first meet at their
  Stop hook, not in the audit report.
- Lint issues in the 8 untouched files predate this card; if CI lints strictly per-package, those
  fail independent of this change (verified absent from my diff, not repaired here — scope).
