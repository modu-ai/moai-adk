# SPEC-INIT-WIZARD-REPAIR-001 — Acceptance Criteria

Verification layer. Requirements (GEARS obligations) live in spec.md §3; every AC below is a binary Given-When-Then scenario mapped to one REQ. All file-path assertions run against a `t.TempDir()` project; "byte-identical" means the deployed file equals the same init run with the relevant selection absent.

## §D AC Matrix

- **AC-001** (REQ-001, chain ① wizard link) **Given** an interactive init whose injected wizard result (`runWizardFn` seam) carries `AutonomyTier: "automatic"` and no explicit `--autonomy-tier` flag, **When** init completes, **Then** `opts.AutonomyTier` resolved to `automatic` AND the deployed settings reflect the tier bundle (USER-scope `defaultMode` written to the injected temp path; PROJECT deny/ask regenerated when a tool-policy doc exists).
- **AC-002** (REQ-001 gates) **Given** a wizard result of `fully-autonomous` with no sandbox proof present, **When** the tier is applied, **Then** the effective tier is downgraded per `EffectiveTierWithGates` and the downgrade advisory is appended to the project's autonomy-downgrade log path.
- **AC-003** (REQ-002, chain ① flag link) **Given** non-interactive init with `--autonomy-tier=automatic` and no TTY, **When** init completes, **Then** the tier reached the bundle apply (flag value not discarded) AND no wizard dependency was required on this path.
- **AC-004** (REQ-003 zero-delta) **Given** an interactive init whose wizard answer is empty/`semi-auto` and no explicit flag, **When** init completes, **Then** neither USER-scope nor PROJECT-scope settings files differ from a run with no autonomy selection at all (zero file delta).
- **AC-005** (REQ-004 + REQ-006, chain ② persist) **Given** init with `--branch-guard` explicitly passed, **When** init completes, **Then** the deployed `workflow.yaml` carries `workflow.branch_guard.enabled: true` AND the same holds for each of the three `--worktree-auto-*` flags against its key.
- **AC-006** (REQ-006 byte-identity) **Given** init with none of the four toggle flags passed (interactive or not), **When** init completes, **Then** the deployed `workflow.yaml` is byte-identical to the template-deployed baseline (no key synthesized, no comment disturbed).
- **AC-007** (REQ-005 precedence) **Given** `--worktree-auto-create=true` explicitly passed and an injected wizard worktree-advisory answer of `false`, **When** init completes, **Then** the persisted `workflow.worktree.auto_create` is `true` (flag wins); **and given** the flag absent with the same wizard answer, the persisted value is the wizard answer.
- **AC-008** (REQ-007 update wizard) **Given** `moai update -c` on a project with a workflow.yaml and a TTY stdin, **When** the wizard finishes applying config, **Then** `runWorkflowConfigStep` ran and only answered-delta keys changed; **given** stdin is not a TTY or workflow.yaml is absent, **then** the step was a no-op (file untouched).
- **AC-009** (REQ-008, chain ③) **Given** an interactive init whose wizard collected an audit selection (`AuditConfigSet=true`, `AuditModel`/gates set, `CodexAuditEnabled=true`), **When** init completes, **Then** the deployed `workflow.yaml` carries the `workflow.audit` block with the selected model/gates and `workflow.codex.review_gate.enabled: true`; **given** `AuditConfigSet=false` (non-interactive), **then** workflow.yaml is byte-identical to the template baseline.
- **AC-010** (REQ-009 comment truth) **Given** the repaired tree, **When** the contract comments in `autonomy_bundle.go`, `initializer.go:79-82`, `init.go:213-216`, and `internal/config/types.go` reader-status are read against the wired call graph, **Then** every claim is true (grep-verifiable: each named function has ≥1 production call site; reader-status lists the two live AutoCleanup readers; AutoMerge remains declared-not-read).

## §D.1 Severity

| AC | Severity | Rationale |
|---|---|---|
| AC-001, AC-003, AC-005, AC-009 | MUST | the core user-facing contract repairs — a dropped selection is the defect |
| AC-004, AC-006 | MUST | regression guards: zero-delta / byte-identity invariants (C6); failure means the repair broke defaults |
| AC-002, AC-007, AC-008 | MUST | safety invariants: gate enforcement, flag precedence, CI no-op — each protects a documented contract |
| AC-010 | SHOULD | comment truth; wrong comments mislead future repairs but do not change runtime behavior |

## §D.2 Traceability

REQ-001→AC-001/002 · REQ-002→AC-003 · REQ-003→AC-001/004 · REQ-004→AC-005 · REQ-005→AC-007 · REQ-006→AC-005/006 · REQ-007→AC-008 · REQ-008→AC-009 · REQ-009→AC-010. Coverage: every REQ has ≥1 MUST/SHOULD AC; no AC orphaned.

## §D.3 Edge cases

- Wizard cancelled (`ErrCancelled`) ⇒ no tier/toggle/audit side effects (existing cancel path unchanged).
- `--autonomy-tier` invalid value ⇒ existing fail-loud validation (init.go:322) unchanged.
- Explicit `--worktree-auto-create=false` ⇒ tracker true, `false` persisted (explicit-off is a selection, not an absence).
- Fallback path (no deployer — tests/minimal init) ⇒ toggles writer and audit writer both behave (fresh-file branches).
- Corrupted project for update wizard (workflow.yaml absent) ⇒ `runWorkflowConfigStep` no-op, never synthesizes the file.

## §D.4 Indirect verification (quality gates)

- `go vet ./internal/cli/... ./internal/core/project/...` and `golangci-lint run` clean on touched packages.
- `GOOS=windows GOARCH=amd64 go build ./...` and `GOOS=linux go build ./...` pass (cross-platform config writes).
- Coverage ≥ 85% on the two touched packages' new/modified paths.
- Full-suite verdict read from CI (no local `go test ./...`).

## §D.5 Closure gates (Definition of Done)

1. All MUST ACs PASS with observed command output + verbatim evidence recorded in progress.md §E.2.
2. The four pre-existing characterization suites (`init_autonomy_wizard_test.go`, `init_workflow_flags_test.go`, `initializer_audit_test.go`, `autonomy_bundle_test.go`) green with unmodified assertions.
3. Zero `[NEEDS CLARIFICATION]` markers remaining in plan.md (both resolved before run-phase entry).
4. No file outside the milestone scope touched (`git diff --stat` review).

## §D.6 Forward-looking checks

- `workflow.worktree.auto_merge` remains declared-without-reader — flagged for the key-honesty backlog, not fixed here.
- The USER-scope write activation (if approved) should be revisited if a future per-project default-mode mechanism lands.

## §D.7 Not tested (out of layer)

- Real `~/.claude/settings.json` mutation (always temp-path injected).
- Wizard TUI rendering/translations (unchanged surfaces).
- Hook-side consumption of `branch_guard.enabled` (already covered by hook tests; here only the persisted value is asserted).
