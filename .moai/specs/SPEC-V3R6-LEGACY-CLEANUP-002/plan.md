# SPEC-V3R6-LEGACY-CLEANUP-002 — Plan

## §A Context

### §A.1 Predecessor SPEC

This SPEC is a pure follow-up cascade from **SPEC-V3R6-LEGACY-CLEANUP-001** (status `implemented`, v0.2.0, closed at HEAD `2509de913`).

LEGACY-CLEANUP-001 cleaned v2.x agency-keyword residuals from 31 user-facing paths (`.claude/`, `docs-site/`, root markdown) but **did NOT touch the embedded template tree** under `internal/template/templates/`. CLAUDE.local.md §2 [HARD] Template-First Rule is therefore currently violated: any new user running `moai init` receives outdated v2.x agency-keyword content despite the orchestrator-side cleanup.

This SPEC restores conformance by mirroring exactly the 5 source files in the LCL-001 commit range `ffa65ab15..19bc873ff` that touched embedded-mirrored paths.

### §A.2 Why Tier S Minimal

- Scope: 5 files only (well below Tier M threshold of 15)
- Operation: verbatim file-copy (no manual edits, no design decisions)
- Risk: ZERO behavioral change (template tree is data-only, consumed by `internal/template/embedded.go` at build time)
- Verification: 5 binary SHA-256 comparisons (mechanical)
- No expected ambiguity, no design.md needed, no research.md needed, no separate acceptance.md (ACs inlined in spec.md §3)

### §A.3 Source → Target Mirror Table

| # | REQ / AC | Source path | Target path |
|---|----------|-------------|-------------|
| 1 | REQ-LCL2-001 / AC-LCL2-001 | `.claude/rules/moai/design/constitution.md` | `internal/template/templates/.claude/rules/moai/design/constitution.md` |
| 2 | REQ-LCL2-002 / AC-LCL2-002 | `.claude/skills/moai-domain-brand-design/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md` |
| 3 | REQ-LCL2-003 / AC-LCL2-003 | `.claude/skills/moai-domain-copywriting/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md` |
| 4 | REQ-LCL2-004 / AC-LCL2-004 | `.claude/skills/moai-workflow-gan-loop/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` |
| 5 | REQ-LCL2-005 / AC-LCL2-005 | `.claude/skills/moai/workflows/design.md` | `internal/template/templates/.claude/skills/moai/workflows/design.md` |

### §A.4 Constraints

- Source files MUST NOT be modified (LCL-001 already cleaned them; no further edits planned).
- Target files MUST be overwritten verbatim from source (no manual reformatting, no whitespace normalization).
- Working tree ambient files (`.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml`, `.moai/research/v3.0-redesign-2026-05-23.md`) MUST NOT be staged or modified.
- Hybrid Trunk Tier S direct push to `main` per CLAUDE.local.md §23.7 (P1 priority + small surgical change + binary verifiable).
- Conventional Commit type: `plan` (per orchestrator spawn directive).

## §B Milestones

Single milestone (Tier S minimal).

### §B.1 M1 — Apply Mirror Diffs

**Goal**: Overwrite the 5 target files with byte-identical copies of their source counterparts, then commit.

**Steps**:

1. Verify pre-flight (see §C below) — all 5 source files exist and 5 target files exist at HEAD.
2. For each of the 5 source/target pairs: `cp -p <source> <target>` (preserves mode, timestamps not critical since git tracks content only).
3. Run trust-but-verify: `shasum -a 256 <source> <target>` for each pair → both digests identical.
4. Stage only the 5 target files: `git add internal/template/templates/.claude/rules/moai/design/constitution.md internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md internal/template/templates/.claude/skills/moai/workflows/design.md` (NO `git add -A` per L37 ambient-mutation discipline).
5. `git status --short` to confirm: only the 5 target files PLUS the 2 SPEC artifacts (spec.md, plan.md from plan-phase) are staged; ambient `.moai/harness/usage-log.jsonl` (M) and untracked `.moai/harness/observations.yaml`, `.moai/research/v3.0-redesign-2026-05-23.md` are NOT staged.
6. Commit: `run(SPEC-V3R6-LEGACY-CLEANUP-002): M1 — template mirror cascade 5 files` + `🗿 MoAI <email@mo.ai.kr>` trailer.
7. Push to `origin/main` (Hybrid Trunk Tier S direct push).
8. Re-run trust-but-verify post-commit (5 SHA-256 pairs) as final AC validation.

**Deliverables**:
- 5 target files updated (SHA-256 byte-identical to source)
- 1 commit on `main`
- Push success to `origin/main`

**Exit criteria**: All 5 ACs PASS via binary SHA-256 comparison.

## §C Pre-flight Checklist (5-cmd batch)

Execute BEFORE M1 step 2 to confirm baseline state. All commands SHOULD be run as a single parallel batch per `.claude/rules/moai/workflow/verification-batch-pattern.md`.

```bash
# C-1. All 5 target files exist (sanity check baseline)
ls -la internal/template/templates/.claude/rules/moai/design/constitution.md \
       internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md \
       internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md \
       internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md \
       internal/template/templates/.claude/skills/moai/workflows/design.md

# C-2. Git working tree clean baseline (only ambient — usage-log.jsonl M, observations.yaml + research/ ??)
git diff --name-only HEAD -- internal/template/templates/

# C-3. golangci-lint baseline (data-only template tree change MUST NOT affect Go lint)
golangci-lint run --timeout=2m ./... 2>&1 | tail -3

# C-4. go test baseline (data-only template tree change MUST NOT affect Go tests — embedded.go is auto-regen)
go test ./internal/template/... 2>&1 | tail -5

# C-5. Agency keyword residual baseline in template tree (informational — verifies the desync hypothesis)
grep -rln -i "agency" internal/template/templates/.claude/rules/moai/design/ internal/template/templates/.claude/skills/moai-domain-brand-design/ internal/template/templates/.claude/skills/moai-domain-copywriting/ internal/template/templates/.claude/skills/moai-workflow-gan-loop/ internal/template/templates/.claude/skills/moai/workflows/ 2>&1 | wc -l
```

Expected outcomes:
- C-1: 5 file paths listed (all exist)
- C-2: 0 lines (no in-flight template modifications)
- C-3: Lint clean (or unchanged baseline)
- C-4: Tests PASS (or unchanged baseline)
- C-5: Some non-zero count of agency keyword occurrences in the 5 files PRE-M1 (will drop after M1 since source is cleaned)

If any pre-flight step fails unexpectedly (e.g., file missing, dirty working tree containing other-session edits in the template tree), the agent MUST return a blocker report and NOT proceed with M1.

---

## §D Out of Scope (explicit non-goals)

- Editing source files in `.claude/` (LCL-001 already cleaned them).
- Editing other template-mirrored files (only the 5 LCL-001 source-mirror pairs).
- Editing docs-site, root markdown, CHANGELOG (LCL-001 sync already handled these in the user-facing tree; no template mirrors exist for `docs-site/` or root markdown).
- Re-running `make build` to regenerate `internal/template/embedded.go` (the embed directive picks up file content at build time; commit + push is sufficient for the SPEC scope. If user later runs `make build`, embedded.go will reflect the new content automatically).
- Verification on `moai init` deployment behavior (deferred to integration testing / next release).
