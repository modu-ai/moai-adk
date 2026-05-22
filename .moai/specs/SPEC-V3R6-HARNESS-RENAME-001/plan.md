---
spec_id: SPEC-V3R6-HARNESS-RENAME-001
version: "0.1.0"
created_at: 2026-05-22
updated_at: 2026-05-22
---

# Implementation Plan — SPEC-V3R6-HARNESS-RENAME-001

## 0. Tier Classification

**Tier S** (LEAN minimal — per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability).

- **LOC estimate**: ~150-250 LOC of string-literal edits across 35-50 files
- **Files affected**: ~35-50 (8 in-place agent+skill bodies + 8 new template mirror files + ~20-30 cross-file references + 10 test files + 2-3 internal/cli/Go files)
- **Tier S deviation from <5 files**: This SPEC has wide reference surface but **shallow per-file edits** (mostly single-string-substitution per file). Per `manager-develop-prompt-template.md` Tier S rationale, complexity is the trigger — not file count. Each file edit is mechanical regex substitution with low cognitive load.

## 1. Approach

### 1.1 Strategy: Mechanical Rename + Template Mirror Creation + Catalog Regen

3-step strategy:
1. **Phase 1 — Source rename** (local `.claude/` + cross-file references + Go test/source literals)
2. **Phase 2 — Template-First mirror** (create `internal/template/templates/.claude/agents/harness/` + 4 skill dirs)
3. **Phase 3 — Catalog hash regen** (`gen-catalog-hashes.go --all` self-healing per SPEC-V3R6-CATALOG-SSOT-001)

### 1.2 Atomic vs Incremental

**Single atomic commit** (no intermediate broken states). Cross-platform build PASS gate ensures no partial rename leaves the tree broken.

### 1.3 Agent Name Convention Decision Matrix

Three options for agent frontmatter `name:` field:

| Option | Pattern | Pro | Con |
|--------|---------|-----|-----|
| A (recommended) | `moai-harness-{X}-specialist` (e.g., `moai-harness-workflow-specialist`) | 1:1 alignment with skill names (`moai-harness-workflow`) | Verbose (28-32 chars) |
| B | `harness-{X}-specialist` (e.g., `harness-workflow-specialist`) | Matches folder name (`harness/`), shorter (22-26 chars) | Skill/agent prefix mismatch |
| C | `harness-{X}` (drop `-specialist` suffix) | Shortest (14-18 chars) | Loses semantic role indicator |

**Default = Option A** (skill-agent parity is the strongest signal per blueprint §0 Decision row 4 `skills/moai-harness-*` series).

If user prefers Option B or C at run-phase, manager-develop applies uniformly to all 4 agents (AC-HRN-006 consistency check).

### 1.4 Backward Compatibility

**No aliases. No symlinks. No deprecation warnings.**

Rationale: All current users of `my-harness/` are MoAI-ADK-internal code paths (no external consumers). Completed/implemented SPECs in `.moai/specs/SPEC-V3R*-HARNESS*/` are historical read-only artifacts. Live agent invocation references in CLAUDE.md/CLAUDE.local.md/skills/rules are updated in this commit. Test files update string literals only — test functions and counts preserved.

## 2. Milestones

### M1 — Source Rename (Local `.claude/` + Cross-File References)

**Goal**: Rename all in-scope local sources to new path/identifier scheme.

**Tasks**:
- M1.1: `git mv .claude/agents/my-harness/ .claude/agents/harness/` (1 directory, 4 files)
- M1.2: `git mv .claude/skills/my-harness-cli-template/ .claude/skills/moai-harness-cli-template/` (1 dir)
- M1.3: `git mv .claude/skills/my-harness-hook-ci/ .claude/skills/moai-harness-hook-ci/` (1 dir)
- M1.4: `git mv .claude/skills/my-harness-quality/ .claude/skills/moai-harness-quality/` (1 dir)
- M1.5: `git mv .claude/skills/my-harness-workflow/ .claude/skills/moai-harness-workflow/` (1 dir)
- M1.6: Update agent frontmatter `name:` field in 4 .md files (per §1.3 decision)
- M1.7: Update `skills:` array in agent frontmatter (e.g., `skills: [my-harness-workflow]` → `skills: [moai-harness-workflow]`)
- M1.8: Update cross-file references in `.claude/rules/`, `.claude/skills/moai/`, `.claude/skills/moai-workflow-design-import/`, `.claude/skills/moai-meta-harness/SKILL.md`, `.claude/skills/moai-harness-learner/SKILL.md`
- M1.9: Update `CLAUDE.md`, `CLAUDE.local.md` if any my-harness references exist
- M1.10: Update Go test files (10 files) — string literal substitution `my-harness` → `harness` for paths, `my-harness-X` → `moai-harness-X` for skill identifiers
- M1.11: Update Go source files (`internal/cli/harness.go`, `internal/cli/update.go`, `internal/cli/doctor_skills.go`, `internal/cli/doctor_harness.go`, `internal/harness/*.go`) — string literal only

**Exit criteria**: `grep -rln "my-harness" .claude/ CLAUDE.md CLAUDE.local.md` returns 0 (research/ excluded per AC-HRN-001).

### M2 — Template-First Mirror Creation

**Goal**: Mirror new agent/skill paths under `internal/template/templates/.claude/`.

**Tasks**:
- M2.1: Create `internal/template/templates/.claude/agents/harness/` directory
- M2.2: Copy 4 agent .md files from `.claude/agents/harness/` → `internal/template/templates/.claude/agents/harness/` (verbatim mirror)
- M2.3: Create 4 skill directories: `internal/template/templates/.claude/skills/moai-harness-{cli-template,hook-ci,quality,workflow}/`
- M2.4: Copy each skill's SKILL.md (+ modules/, examples.md, reference.md if present) verbatim from `.claude/skills/moai-harness-X/` to template path
- M2.5: Run `make build` to regenerate `internal/template/embedded.go`

**Exit criteria**: AC-HRN-002 PASS (4 agent files + 4 skill dirs present under template).

### M3 — Catalog SSoT Regen + Build Verification

**Goal**: Refresh catalog.yaml hashes and verify build + test sanity.

**Tasks**:
- M3.1: `go run gen-catalog-hashes.go --all` (Makefile `build` recipe self-healing per SPEC-V3R6-CATALOG-SSOT-001)
- M3.2: Verify `internal/template/catalog.yaml` contains entries for `moai-harness-cli-template`, `moai-harness-hook-ci`, `moai-harness-quality`, `moai-harness-workflow` (NOT `my-harness-*`). Manual entry add if `gen-catalog-hashes.go` does not auto-discover.
- M3.3: `go test -run TestManifestHashFormat ./internal/template/...` (AC-HRN-005)
- M3.4: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` (AC-HRN-003)
- M3.5: Run full in-scope test suite: `go test -count=1 ./internal/cli/... ./internal/template/... ./internal/design/pipeline/... ./internal/harness/...` (AC-HRN-004 baseline comparison)
- M3.6: Commit (Conventional Commits + `🗿 MoAI` trailer + Late-Branch main 직진)
- M3.7: Update SPEC frontmatter `status: draft → implemented` + `version: 0.1.0 → 0.2.0`

**Exit criteria**: All 7 ACs PASS.

## 3. Technical Approach

### 3.1 Tooling

- **Grep + Edit (multi-occurrence)**: Mechanical string substitution per file. Use `Edit` tool with `replace_all: true` flag when literal is unambiguous within a file.
- **git mv**: Preserves Git history per directory rename (5 invocations: 1 agent dir + 4 skill dirs).
- **gen-catalog-hashes.go**: Self-healing manifest hash regen.
- **make build**: Embedded template regeneration.

### 3.2 Edit Discipline

For each cross-reference file:
1. Read full file to identify context
2. Apply `Edit` with `replace_all: false` for unique strings, `replace_all: true` for uniform-rename literals
3. Verify with `grep -n` after edit

**No MultiEdit** for this SPEC — atomic per-file edits reduce risk of partial state on partial failure.

### 3.3 Catalog Entry Add

If `gen-catalog-hashes.go --all` does not auto-discover new skill paths, manually add 4 entries to `internal/template/catalog.yaml` under `skills:` section with structure:
```yaml
- name: moai-harness-X
  path: .claude/skills/moai-harness-X
  hash: <generated by gen-catalog-hashes.go>
  tier: ...
```

Then re-invoke `gen-catalog-hashes.go --all` to populate hash fields.

## 4. Risks and Mitigations

| Risk | Mitigation Approach |
|------|---------------------|
| R-HRN-001 (hidden literal) | Exhaustive grep across `.claude/ .moai/research/ CLAUDE.md CLAUDE.local.md internal/` with whitelist for historical research docs. AC-HRN-001 binary gate. |
| R-HRN-002 (name convention) | §1.3 decision matrix in this plan; default to Option A unless user explicitly chooses B/C at run-phase. AC-HRN-006 consistency check. |
| R-HRN-003 (Template-First miss) | M2 phase is explicit and mandatory; AC-HRN-002 binary gate. |
| R-HRN-004 (catalog hash drift) | M3.1 + M3.3 sequential — `gen-catalog-hashes.go --all` runs before `TestManifestHashFormat`. Manual catalog.yaml entry add if needed. |
| R-HRN-005 (filename drift in test file) | Default: keep `update_preserve_my_harness_test.go` filename; update string literals inside. Document in progress.md. |

## 5. Out of Scope (re-statement for run-phase clarity)

- `.moai/harness/` runtime directory (REQ-HRN-008)
- Backward-compat aliases or symlinks (REQ-HRN-009)
- AGENT-FOLDER-SPLIT (`agents/moai/` → `agents/{core,expert,meta}/`) — separate SPEC
- Skill content body restructure — Wave 3 SPEC
- Test filename `update_preserve_my_harness_test.go` rename (default = keep)
- Symbol-level Go API renames (only string literals change)
- docs-site/ updates (0 refs verified)
- Historical SPEC frontmatter `id:` field rewrites (out-of-scope, historical artifact integrity)
- Memory file `~/.claude/projects/.../memory/` references — those are user-local memory persisted across sessions, not part of project source

## 6. Verification Approach

**During M1-M3**: After each milestone, run incremental `go build ./...` to detect early breakage.

**After M3 completion**: Execute all 7 AC verification commands in single orchestrator turn (parallel multi-Bash per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution).

## 7. Estimated Effort

- Tier S minimal: 60-90 minutes orchestrator-direct execution OR 30-45 minutes manager-develop delegation (1-pass expected per § B5 baseline and § B2 cross-SPEC scan completion at plan-phase)
- Rollback complexity: Low (single atomic commit; `git revert <sha>` restores prior state)

## 8. Late-Branch Workflow

Per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005:
- plan-phase commit: `main` 직진 (this artifact)
- run-phase commit: `main` 직진 (after M3 completion)
- sync-phase: stash → branch → cherry-pick → PR (deferred until user decision)

## 9. Open Questions for Run-Phase

1. **Agent name convention finalization**: A (default) / B (folder-aligned shorter) / C (shortest)? — Default Option A unless user override.
2. **Test filename rename**: `update_preserve_my_harness_test.go` → keep (default) or `git mv` to `update_preserve_harness_test.go`? — Default = keep, string-literal-only.
3. **Manual catalog entry add**: If `gen-catalog-hashes.go --all` does not auto-discover the 4 new skill paths, manager-develop manually adds entries with structure shown in §3.3. — Auto-handled.
