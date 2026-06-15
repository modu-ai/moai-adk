# SPEC-PROGRESS-MARKER-CANON-001 — Implementation Plan

## A. Context

- **Project root**: `/Users/goos/MoAI/moai-adk-go`
- **Branch strategy**: Hybrid Trunk 1-person OSS — all-tier main direct push (per `.moai/docs/git-local-workflow-doctrine.md`). No PR required for Tier M doc-canon SPEC unless `--pr` requested.
- **SPEC artifacts**: `.moai/specs/SPEC-PROGRESS-MARKER-CANON-001/{spec,plan,acceptance,progress}.md`
- **Tier**: M (3 source files edited across 2 namespaces + 1 mirror + comment-only Go change; zero new production code)
- **cycle_type**: `tdd` is the project default, but this SPEC has ZERO new code and ZERO new tests — the verification is grep-based + existing-test-preservation. Recommend `cycle_type=tdd` with the understanding that the "test" is `go test ./internal/spec/...` preservation (REQ-PMC-006), not a new RED test.

### Existing infrastructure (PRESERVE vs EDIT)

- **PRESERVE (byte-identical)**: era.go executable logic (L110-134, L161), H-1..H-6 heuristic table in lifecycle-sync-gate.md (L37-45), JSON audit-output excerpt, all era.go tests.
- **EDIT**: era.go comments only (L33, L92, const-block prose L29-38, ClassifyEra doc-comment L83-94); lifecycle-sync-gate.md worked example (L294-317 region); manager-spec.md body (insert skeleton instruction) + its template mirror.

## B. Known Issues (filtered for this SPEC)

- **B4 (Frontmatter Canonical Schema)**: This SPEC's own spec.md uses `created`/`updated`/`tags` (canonical, not snake_case). Verified at authoring.
- **B6 (spec-lint Heading)**: spec.md §B uses H3 `### Out of Scope — <topic>` sub-headings (4 entries) to satisfy `MissingExclusions`. Verified.
- **B10 (Untouched Paths PRESERVE)**: era.go is the ONLY `.go` file touched, and ONLY its comments. Do NOT touch era_test.go, audit.go, lint.go, or any other internal/spec file.
- **B8 (Working Tree Hygiene)**: The working tree currently has unrelated modified files (settings.json, deployer.go, sync-phase-quality-gate.sh, etc. — a parallel statusline workstream). Do NOT include those in this SPEC's commits. `git add` only this SPEC's 4 target files + the SPEC artifact dir.
- **B2 (Cross-SPEC conflict scan)**: lifecycle-sync-gate.md is owned by SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 (completed). Editing its worked example is a clarification, not a reversal — the heuristic semantics are unchanged. No conflict.

Filtered OUT (not applicable): B1 (no syscall), B3/B11 (no subagent-domain code edited), B5 (no CI-tier-specific code), B7 (no observer path), B9/B12 (manager-develop B9 applies at run-phase; this is plan-phase).

## C. Pre-flight Checks (run-phase, before any edit)

```bash
# 1. baseline branch + HEAD
git branch --show-current
git rev-parse HEAD

# 2. era.go tests pass BEFORE edit (baseline for REQ-PMC-006)
go test ./internal/spec/... 2>&1 | tail -5

# 3. build green before edit
go build ./...

# 4. confirm mirror parity before edit (so post-edit parity is meaningful)
diff -q .claude/agents/moai/manager-spec.md internal/template/templates/.claude/agents/moai/manager-spec.md

# 5. confirm lifecycle-sync-gate.md has NO mirror (edit single copy)
ls internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md 2>/dev/null || echo "NO MIRROR (correct)"
```

## D. Constraints (DO NOT VIOLATE)

- era.go: comments ONLY. Every non-comment line byte-identical. (REQ-PMC-005)
- No new test, no new function, no new lint rule. (CON-PMC-002)
- No retro-edit of any existing SPEC progress.md. (CON-PMC-003)
- lifecycle-sync-gate.md: single copy, no mirror, no `make build`. (REQ-PMC-011)
- manager-spec.md: BOTH copies edited byte-identical, then `make build`. (REQ-PMC-010)
- manager-spec.md body edit uses generic prose — do NOT embed this SPEC's ID in the agent body (template neutrality, CON-PMC-005).
- Exclude unrelated working-tree changes from commits (B8).
- `--no-verify` / `--amend` / force-push prohibited.

## E. Self-Verification (run-phase deliverables — map to acceptance.md)

The implementing agent must report, with actual command output:

1. era.go grep-equality proof: `git diff internal/spec/era.go` shows ONLY comment-line (`//`) changes — zero changes to lines without a leading `//` (excluding const declarations whose VALUES are unchanged). (AC-PMC-004, AC-PMC-005)
2. `go test ./internal/spec/...` PASS, identical to baseline. (AC-PMC-006)
3. lifecycle-sync-gate.md worked example greps: `§E.2 Run-phase Evidence` present, `§E.4 Sync-phase Audit-Ready Signal` present, `§E.2 Sync-phase Audit-Ready Signal` (convention A) ABSENT. (AC-PMC-001, AC-PMC-002)
4. lifecycle-sync-gate.md heuristic table rows unchanged: `git diff` shows no change in the L37-45 region. (AC-PMC-003)
5. manager-spec.md skeleton instruction present + all 5 §E headings enumerated in convention-B order. (AC-PMC-007, AC-PMC-008)
6. manager-spec.md mirror parity restored + embedded.go regenerated. (AC-PMC-010)
7. `go build ./...` green post-edit. (AC-PMC-006 build half)

## F. Milestones (priority-ordered, no time estimates)

- **M1 — era.go comment correction**: Edit L33 const-block comment, L92 doc-comment, and surrounding §E.2-"sync" prose to convention-B semantics. Verify `git diff` is comment-only. Run `go test ./internal/spec/...` to confirm preservation. Commit `fix(SPEC-PROGRESS-MARKER-CANON-001): M1 era.go §E.2/§E.5 comment correction (behavior-preserving)`.

- **M2 — lifecycle-sync-gate.md worked-example alignment**: Edit the worked-example progress.md excerpt (L294-317 region) so §E.2 → §E.4 for the Sync heading, add `§E.2 Run-phase Evidence` framing, keep `§E.5 Mx`. Update the auto-detection trace narration for consistency while preserving H-4 detection validity. Verify heuristic table + JSON excerpt unchanged. Commit `docs(SPEC-PROGRESS-MARKER-CANON-001): M2 lifecycle-sync-gate worked-example §E convention-B alignment`.

- **M3 — manager-spec.md skeleton instruction (both copies + build)**: Insert the progress.md skeleton-generation instruction (5 §E placeholder headings, minimal) into the manager-spec.md body. Apply byte-identical to the template mirror. Run `make build`. Verify mirror parity + embedded regeneration. Commit `feat(SPEC-PROGRESS-MARKER-CANON-001): M3 manager-spec progress.md skeleton-generation instruction + mirror + build`.

- **M-final — verification batch + progress.md §E.2/§E.3 population**: Run the full self-verification batch (E1-E7 above). Populate this SPEC's own progress.md §E.2 (Run Evidence) + §E.3 (Run AR). Transition `draft → in-progress` is owned by manager-develop on M1 commit per the ownership matrix.

> Status transition note: manager-spec (this agent) emits initial `status: draft` (done at plan-phase). `draft → in-progress` is manager-develop's on M1. `in-progress → implemented` + `→ completed` are manager-docs/orchestrator at sync/Mx.

## G. Anti-Patterns to avoid

- Renaming era.go variables (`hasSyncSection`/`hasMxSection`) — OUT of scope, comment-clarify only.
- Editing the H-1..H-6 heuristic table — it is presence-accurate, not semantically wrong.
- Retro-fixing the 16 §F.* SPECs — grandfather-protected.
- Adding a §E-marker lint rule — deferred to follow-up.
- Embedding this SPEC's ID in the manager-spec.md agent body — template neutrality violation.
- Bundling the unrelated statusline working-tree changes into this SPEC's commits.

## H. Cross-References

- `internal/spec/era.go` — H-1..H-6 + `hasProgressMarker` (the behavior to PRESERVE)
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` § Era Classification + worked example (M2 target)
- `.claude/agents/moai/manager-spec.md` § SPEC Artifact Ownership + Forbidden modifications (convention-B SSOT, M3 target)
- `.claude/agents/moai/manager-develop.md` §Artifacts owned (§E.2 Run Evidence / §E.3 Run AR — convention-B reference)
- `.claude/agents/moai/manager-docs.md` §Artifacts owned (§E.4 Sync AR — convention-B reference)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- CLAUDE.local.md §2 (Template-First Rule + mirror discipline), §24 (Harness Namespace), §25 (Template Internal-Content Isolation)
