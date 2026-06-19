# Progress — SPEC-V3R6-RULES-HOTFIX-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored: `spec.md` (Tier S LEAN, inline acceptance criteria), `progress.md` (this file).
- SPEC ID self-check: `SPEC ✓ | V3R6 ✓ | RULES ✓ | HOTFIX ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- 6 in-scope defects (D1–D6) verified against deployed file + template mirror with cited file:line.
- 9 acceptance criteria (AC-RHF-001 … AC-RHF-009), each grep-verifiable in both deployed + template paths.
- Borderline glossary item DEFERRED as a forward-link (see spec.md §F.1) — new authoring, out of hotfix charter.
- v0.2.0 revision (plan-auditor 0.82 PASS-WITH-DEBT, independently re-verified by orchestrator): BLOCKING `embedded.go` premise PURGED (verified absent via `find`; real mechanism `//go:embed all:templates` at `embed.go:28`; parity via `TestRuleTemplateMirrorDrift`); D1 live-runtime citation + tool-policy.yaml:776 forward-note; D2-file intentional neutrality divergence (no byte-identity reconcile); D4/D5 prose rewrites with correct labels; AC-RHF-006 `:0` predicate tightened.
- Status: `draft`. Awaiting plan-audit gate, then Implementation Kickoff Approval before run-phase.

## §E.2 Run-phase Evidence

Run-phase executed by `manager-develop` (worktree `agent-a489207659f3fb2ac`, baseline `a964772fa`); edits propagated to shared main checkout via `git apply` (orchestrator reconciliation — worktree isolation prevented direct shared-tree commit, see § Notes below).

**D1 — zai vision tool names (AC-RHF-001)** — `grep 'image_analysis\|video_analysis' glm-web-tooling.md` → 0; `grep -c 'analyze_image\|analyze_video'` → 5. Live-runtime-confirmed tool names (`mcp__zai-mcp-server__analyze_image`, `analyze_video`) replace inverted `image_analysis`/`video_analysis` in the HARD routing table + 8-vision-tool table + ToolSearch preload + AP-GWT-003.

**D2 — inverted-logic clause (AC-RHF-002)** — `grep -c 'An actor MUST NOT assert' verification-claim-integrity.md` → 1; `grep -c 'No actor MUST assert'` → 0. The §1.1 binding-scope HARD clause now reads `[HARD] An actor MUST NOT assert a verification, a completion, OR a defect / debt / drift it did not actually verify` (direct-logic, not double-negative).

**D3 — cpp flag typo (AC-RHF-003)** — `grep 'adddress\|adddess' cpp.md` → 0; `grep -c 'fsanitize=address'` → 1.

**D4 — broken anti-patterns ref (AC-RHF-004)** — `grep -c 'moai-reference-anti-patterns' karpathy-quickref.md` → 0; corrected to the real `.claude/skills/moai/references/anti-patterns.md` path; target file existence verified.

**D5 — broken handoff ref (AC-RHF-005)** — `grep -c 'modules/trigger-handoff.md' ci-watch-protocol.md` → 0; the `modules/` directory is absent — ref corrected to the actual SKILL.md path; target verified.

**D6 — two path errors (AC-RHF-006)** — worktree-integration.md §Terminology Glossary + branch-origin / git-workflow-doctrine path errors resolved; `grep` for wrong paths → 0.

**Mirror parity (AC-RHF-007)** — 7 deployed files + 7 template mirrors edited identically (22 insertions / 22 deletions, token-only). `find internal/template -name embedded.go` → no output (premise absent; real mechanism `//go:embed all:templates` at `embed.go:28`); parity enforced via `TestRuleTemplateMirrorDrift`.

**Neutrality (AC-RHF-008)** — SPEC-ID / REQ-token / commit-SHA leak grep → 0; `TestTemplateNeutralityAudit` PASS (no internal-content leakage into template mirrors).

**Build (AC-RHF-009)** — `go vet ./...` exit 0; `go build ./...` exit 0; `TestRuleTemplateMirrorDrift` PASS. Known pre-existing failure `TestOutputStylesTemplateLiveParity` (moai.md output-styles drift, 3-line `Interrupt Closure`) is OUT OF SCOPE per SPEC §F — not in this edit set.

## §E.3 Run-phase Audit-Ready Signal

- Run-phase complete: 6 defect groups (D1–D6) resolved across 14 files (7 deployed + 7 template mirror).
- 9 ACs verified: AC-RHF-001..008 PASS, AC-RHF-009 PASS-WITH-DEBT (pre-existing out-of-scope failure, documented).
- `moai spec lint` clean (0 findings).
- run_commit_sha: d7d48e648aa5d2d0572d03e4527480f6b8787016
- Status: `in-progress`. Ready for sync-phase.

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: <backfill pending — close commit SHA>

Sync-phase orchestrator-direct (GLM manager-docs spawn context-limit fallback per `feedback_glm_orchestrator_direct_sync_mx`). frontmatter status in-progress → completed rides this sync commit (3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-008 — `completed` transition merged into the sync commit, no separate Mx chore). §E.5 Mx-phase retired (folded into §E.4 per LIFECYCLE-REDESIGN). CHANGELOG entry added (Sprint 16 RULES cohort).

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_
