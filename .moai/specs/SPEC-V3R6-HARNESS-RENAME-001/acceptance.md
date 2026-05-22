---
spec_id: SPEC-V3R6-HARNESS-RENAME-001
version: "0.1.0"
created_at: 2026-05-22
updated_at: 2026-05-22
---

# Acceptance Criteria — SPEC-V3R6-HARNESS-RENAME-001

7 binary ACs. All must PASS for SPEC to be marked `implemented`.

## AC-HRN-001 — Source File Rename Completeness

**Given** the local working tree state after run-phase completion,
**When** the orchestrator executes the verification command:
```bash
grep -rln "my-harness" .claude/ .moai/research/ CLAUDE.md CLAUDE.local.md 2>/dev/null \
  | grep -v "v3-redesign-blueprint-2026-05-22.md" \
  | grep -v "harness-system-audit-2026-05-14.md" \
  | grep -v "harness-autonomy-vision-2026-05-18.md" \
  | grep -v "core-slimming-audit-2026-05-20.md" \
  | grep -v "architecture-audit-2026-05-18.md" \
  | grep -v "workflow-audit-2026-05-16.md" \
  | grep -v "docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md"
```
**Then** the output shall be empty (exit 0 with 0 lines).

**Rationale**: Historical research documents (`.moai/research/*.md` listed above) and v3 blueprint preserve `my-harness` as historical evidence per Out-of-Scope. All live agent/skill/rule/CLAUDE/CLAUDE.local references must be renamed.

**Pass**: 0 lines output
**Fail**: ≥1 line output (live reference still uses `my-harness`)

---

## AC-HRN-002 — Template-First Mirror Presence

**Given** the local working tree state after run-phase completion,
**When** the orchestrator executes:
```bash
test -d internal/template/templates/.claude/agents/harness && \
  ls internal/template/templates/.claude/agents/harness/ | wc -l && \
test -d internal/template/templates/.claude/skills/moai-harness-cli-template && \
test -d internal/template/templates/.claude/skills/moai-harness-hook-ci && \
test -d internal/template/templates/.claude/skills/moai-harness-quality && \
test -d internal/template/templates/.claude/skills/moai-harness-workflow && \
echo "MIRROR_OK"
```
**Then** the output shall contain `4` (4 agent files) followed by `MIRROR_OK`.

**Rationale**: Template-First Rule (CLAUDE.local.md §2) — all `.claude/` artifacts must have a corresponding `internal/template/templates/.claude/` mirror.

**Pass**: Output contains `4` + `MIRROR_OK`
**Fail**: Any of the 4 skill dirs or the agent dir absent in template

---

## AC-HRN-003 — Cross-Platform Build Sanity

**Given** the source tree after run-phase completion,
**When** the orchestrator executes:
```bash
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
```
**Then** both commands shall exit with code 0 (no compilation errors from broken string references).

**Pass**: Both exit 0
**Fail**: Either exits non-zero

---

## AC-HRN-004 — Test Count Preservation

**Given** the test suite state before and after the rename,
**When** the orchestrator executes:
```bash
go test -count=1 ./internal/cli/... ./internal/template/... ./internal/design/pipeline/... ./internal/harness/... 2>&1 \
  | grep -E "^(ok|FAIL|---)" | wc -l
```
**Then** the test outcome line count after rename shall be ≥ baseline measured before rename.

**Baseline measurement** (executed at plan-phase Pre-flight Step 4 result, captured as variable):
```bash
go test -count=1 ./internal/cli/... ./internal/template/... ./internal/design/pipeline/... ./internal/harness/... 2>&1 \
  | grep -E "^(ok|FAIL|---)" | wc -l
```

**Rationale**: This SPEC must not delete any test. String literals inside test files are updated, but test function signatures and counts are preserved.

**Pass**: Post-rename count ≥ baseline count
**Fail**: Post-rename count < baseline count (indicates accidental test deletion)

---

## AC-HRN-005 — Catalog SSoT Consistency

**Given** `internal/template/catalog.yaml` after run-phase completion,
**When** the orchestrator executes:
```bash
go test -run TestManifestHashFormat ./internal/template/...
```
**Then** the test shall PASS (exit 0).

**Rationale**: Catalog manifest hashes (SHA256 of skill/agent files) must be regenerated after rename. SPEC-V3R6-CATALOG-SSOT-001 introduced `gen-catalog-hashes.go --all` self-healing — this AC verifies the gate is satisfied.

**Pass**: Exit 0
**Fail**: Exit non-zero (hash drift detected)

---

## AC-HRN-006 — Agent Name Self-Reference Consistency

**Given** the 4 renamed agent .md files under `.claude/agents/harness/`,
**When** the orchestrator executes:
```bash
for f in .claude/agents/harness/*.md; do
  name_field=$(awk '/^name:/ {print $2; exit}' "$f")
  filename_base=$(basename "$f" .md)
  echo "FILE: $filename_base | NAME_FIELD: $name_field"
done
```
**Then** for each file, the agent frontmatter `name:` field shall match the chosen naming convention (Option A `moai-harness-X-specialist`, Option B `harness-X-specialist`, or Option C — final decision documented in plan.md §1.3 and applied uniformly to all 4 agents).

**Pass**: All 4 agents use the same prefix convention (consistency check)
**Fail**: Mixed prefix conventions across the 4 agents

---

## AC-HRN-007 — Exclusion Compliance

**Given** the working tree state after run-phase completion,
**When** the orchestrator executes:
```bash
test -d .moai/harness && \
test -f .moai/harness/README.md && \
test -f .moai/harness/main.md && \
test -d .moai/harness/seeds && \
test -f .moai/harness/usage-log.jsonl && \
echo "EXCLUSION_OK"
```
**Then** the output shall be `EXCLUSION_OK` (REQ-HRN-008: `.moai/harness/` runtime config preserved untouched).

**Pass**: Output is `EXCLUSION_OK`
**Fail**: Any of the 4 paths missing (indicates accidental modification of excluded runtime config)

---

## Edge Cases

### EC-HRN-001: Filename Drift in Test Files

Test file `internal/cli/update_preserve_my_harness_test.go` contains `my_harness` in its filename. The string-literal-only policy (Out-of-Scope §3.7) means the filename stays as-is unless `git mv` is explicitly chosen at run-phase. If renamed, ensure all import paths and references update atomically.

**Resolution**: Default = keep filename. Document any rename in run-phase progress.md.

### EC-HRN-002: SPEC `id:` Field Self-Reference

Active SPECs (e.g., `SPEC-V3R6-HARNESS-LEARNER-FIX-001`) contain `my-harness`/`moai-harness` in body text. Frontmatter `id:` field is historical artifact and must NOT be renamed (per Exclusions §5.4).

**Resolution**: Only update body text inside completed SPEC files if it points to a now-moved path (e.g., link rewrites in plan.md body). `id:` field stays untouched.

### EC-HRN-003: Catalog Hash Cascade

After rename, `gen-catalog-hashes.go --all` must run to refresh the 2 known stale entries (`moai-harness-learner`) PLUS the 4 newly-named entries (`moai-harness-cli-template`, etc.). If the script does not detect the 4 new entries, manual regen of all harness-category entries is required.

**Resolution**: Verify Makefile `build` recipe invokes `gen-catalog-hashes.go --all` (SPEC-V3R6-CATALOG-SSOT-001 introduced this gate); manually invoke if needed.

---

## Definition of Done

- [ ] All 7 ACs PASS
- [ ] SPEC frontmatter status: `draft → implemented` (version 0.1.0 → 0.2.0)
- [ ] progress.md created with final commit SHA and AC matrix
- [ ] Commit message follows Conventional Commits + `🗿 MoAI` trailer
- [ ] No accidental modification of working tree untracked files (9 preserved)
- [ ] Late-Branch workflow: commit on main, push deferred to sync-phase
