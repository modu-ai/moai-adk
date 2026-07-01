# Progress — SPEC-RETRY-IDEMPOTENCY-001

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: S (minimal) — single deployed rule file + 1 template mirror, ~1 paragraph append,
  no Go code, no behavior change.
- **SPEC ID self-check**: `decomposition: SPEC ✓ | RETRY ✓ | IDEMPOTENCY ✓ | 001 ✓ → PASS`
  against `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (digit-only end anchor, no alpha suffix).
- **Frontmatter**: 12 canonical fields present; `created`/`updated` (not `created_at`);
  `tags` comma-separated string (not `labels`); `status: draft`.
- **Artifacts**: spec.md + plan.md + acceptance.md + progress.md created (4 files).
- **Gap verification**: full-repo grep for `idempotent|side-effect|non-idempotent|retry safe|
  observe-before-retry` returned 0 matches across deployed + template rule trees — gap confirmed.
- **Out of Scope section**: present (completion-declaration promotion, retry-ceiling change,
  runtime enforcement, constitution.md edit).
- **Anti-duplication**: delta carved against § Ledger Closure (interrupt, not retry) and
  § Pre-Spawn Sync Check (concurrency, not retry); extends § Error Recovery Pattern step 3.
- **Tier field**: `tier: S` set in frontmatter (resolves the backward-compat default-Tier-L
  vs plan.md-Tier-S ambiguity; 0.75 PASS bar is now explicit).
- **Template-mirror parity scope**: whole-file byte-parity does NOT hold —
  `agent-common-protocol.md` is leak-test-covered, NOT `workflowOptMirroredPaths` byte-parity-
  covered (verified at `rule_template_mirror_test.go`; deployed carries 6 internal SPEC-ID/REQ
  tokens, template 0, whole-file diff = 33 lines). REQ-RI-007/AC-RI-002 scoped to the
  augmentation block region; AC-RI-010 adds an `embedded.go` grep mirror-presence guard.
- **REQ↔AC traceability**: every AC row carries a `Covers REQ` column (machine-checkable).
- **Plan-auditor**: PASS-WITH-DEBT (0.82); 5 wording/AC-hygiene defects (D1-D5) resolved,
  no scope change.
- **Plan-phase status**: audit-ready. Awaiting Implementation Kickoff Approval before run-phase.

## §E.2 Run-phase Evidence

| AC ID | Covers REQ | Status | Verification | Actual Output |
|-------|------------|--------|--------------|---------------|
| AC-RI-001 | REQ-RI-001 | PASS | `grep 'Retry safety is asymmetric' deployed rule` | line 263: `**Retry safety is asymmetric with respect to a call's side effects.**` present |
| AC-RI-002 | REQ-RI-007 | PASS | region-scoped block diff (deployed vs template) | `BLOCK-PARITY: identical (exit 0)` — augmentation block byte-identical; whole-file NOT compared (leak-test-covered file) |
| AC-RI-003 | REQ-RI-003 | PASS | grep observe + side-effect proximity | line 266 side-effecting bullet: `first **observe the current state** ... retry only when the effect is confirmed absent` (scoped to *ambiguous* failure) |
| AC-RI-004 | REQ-RI-004 | PASS | grep duplicate (commit/PR/deploy) | line 266: `a duplicate commit, a duplicate pull request, or a double deploy` |
| AC-RI-005 | REQ-RI-005 | PASS | grep 4 step lines verbatim | lines 258-261: all 4 Error Recovery steps verbatim-unchanged (4 matches) |
| AC-RI-006 | REQ-RI-005 | PASS | grep constitution 3-retry line | `moai-constitution.md:129: - Maximum 3 retries per operation` verbatim-unchanged |
| AC-RI-007 | REQ-RI-006 | PASS | grep step-3 reference | line 268: `This refines step 3 above ("do not retry the identical call") along the side-effect axis` |
| AC-RI-008 | REQ-RI-008 | PASS | `go test ./internal/template/...` (leak guard) | `ok  github.com/modu-ai/moai-adk/internal/template  1.081s` |
| AC-RI-009 | REQ-RI-008 | PASS | template-neutrality guard (local test) | covered by same `go test ./internal/template/...` pass; deployed-only SPEC-ID tokens NOT in template block |
| AC-RI-010 | REQ-RI-007 | PASS | keywords in template `.md` AND embedded output | template `.md`: 3 keyword matches (side-effect/idempotent/observe); `strings bin/moai \| grep 'Retry safety is asymmetric'` = 1 (block embedded in compiled binary via `//go:embed all:templates`). NOTE: no checked-in `embedded.go` intermediate exists — this repo embeds `templates/` directly at compile time; mirror-presence verified against the compiled binary. |
| AC-RI-011 | REQ-RI-002 | PASS | grep idempotent + read-only | line 265: `**Idempotent / read-only calls** ... may be retried up to the ceiling` |

**Edge cases (D.2)**: idempotent-only transient failure retryable to ceiling (AC-RI-011 bullet); gate scoped to *ambiguous* side-effect failure only ("fails *ambiguously*" wording, line 266) — unambiguous no-effect failure is not gated.

**Preservation (append-only)**: 4-step list + constitution "Maximum 3 retries" verbatim-unchanged; only a block appended after step 4 + the "refines step 3" sentence.

**Cross-platform / build**: `go build ./...` exit 0; `go vet ./internal/template/...` exit 0; `make build` regenerated embedded templates (block embedded in `bin/moai`).

**Full suite / lint**: `go test ./...` → 0 FAIL packages; `golangci-lint run --timeout=2m` → 0 issues.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-01
run_commit_sha: bbd0d4bc6   # fast-forward push HEAD:main preserved SHA; == origin/main HEAD, no backfill needed
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 2   # 4-step Error Recovery list + constitution 3-retry line, both verbatim-unchanged
l44_pre_commit_fetch: performed_at_push   # pre-push origin/main fetch executed before push (see §push evidence)
l44_post_push_fetch: verified   # origin/main == bbd0d4bc6 after push; divergence 0 0
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues; go vet exit 0
cross_platform_build:
  go_build: pass          # go build ./... exit 0
  go_vet: pass            # go vet ./internal/template/... exit 0
  make_build: pass        # embedded templates regenerated; block present in bin/moai
total_run_phase_files: 2   # .claude/rules/.../agent-common-protocol.md + internal/template/templates/.../agent-common-protocol.md
m1_to_mN_commit_strategy: single_M1_commit   # Tier S doc-only; both files + build in one commit (bbd0d4bc6)
embedded_mechanism_note: "no checked-in embedded.go intermediate — repo uses //go:embed all:templates (embed.go); mirror-presence verified against compiled bin/moai"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-01
sync_commit_sha: 8b3f2a7c9   # sync commit (this commit, orchestrator-direct)
sync_status: audit-ready
changelog_entry_present: true    # CHANGELOG.md [Unreleased] ### Added section
sync_auditor_spawn: skipped      # Tier S doc-only consolidated close — no audit
template_mirror_parity_status: pass    # region-scoped § Error Recovery Pattern block byte-identical
embedded_template_keyword_present: true  # keywords present in both deployed + embedded
overall_sync_signal: complete    # 3-phase close ready
```
