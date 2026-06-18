# plan.md — Implementation Plan

> Run-phase implementation guide for `manager-develop`. Each milestone is
> a discrete, committable unit. Go code changes are enumerated per-file
> with line numbers and exact intended edits.

## §A. Context

See `spec.md` §A (retirement rationale) and `research.md` (evidence base).
The retirement resumes the H5 attempt staged in the 2026-06-18
skill-audit-followup session. The orchestrator has already
`git stash pop`-ed the 2-file SKILL.md deletion seed (unstaged in working
tree); run-phase confirms those deletions at commit and applies the
remaining 8 file-category edits.

## §B. Known issues / pre-conditions

1. **Working-tree state**: 2 unstaged deletions present (`.claude/skills/moai-design-system/SKILL.md`, `internal/template/templates/.claude/skills/moai-design-system/SKILL.md`). Run-phase M1 commits these as the seed.
2. **Parallel-session race risk**: the prior H5 attempt was aborted on a parallel-session working-tree race. Run-phase MUST verify `moai session list --json` shows no other active session on this SPEC scope before starting (per CLAUDE.local.md §23.8 multi-session race mitigation).
3. **Pre-existing test failures**: the `internal/statusline` TestCollectMemory / TestCollectMemory_AutoCompactScaling failures are PRE-EXISTING (per `project_skill_audit_followup_wip` memory line 19, caused by the GLM 1M transition commit `192bd5f81`, unrelated to this SPEC). Run-phase must NOT treat these as regression signals but MUST distinguish them from any new failures this SPEC's changes might introduce.

## §C. Pre-flight checks (run before M1)

```bash
# C.1 No active parallel session on this SPEC scope
moai session list --json --filter-spec=SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001
# Expected: [] (empty array)

# C.2 Confirm the 2 stash-pop deletions are present as unstaged
git status --short | grep "moai-design-system/SKILL.md"
# Expected: 2 lines starting with " D "

# C.3 Baseline the current Go test state (to distinguish pre-existing failures)
go test ./internal/cli/... ./internal/design/dtcg/... 2>&1 | tee /tmp/dsr-baseline.txt
# Record: which tests pass/fail BEFORE any change. The statusline failures
# are NOT in these packages — they are in internal/statusline. So this
# baseline should be clean for cli/ + dtcg/.
```

## §D. Constraints

- **No time estimates** (CLAUDE.md §7 + agent-common-protocol Time Estimation). Milestones are ordered by dependency, not duration.
- **Template-neutrality symmetry** preserved at every milestone.
- **Historical-reference immutability** (design.md §D3.4) — no edits to CHANGELOG/.moai archival records.
- **Commit cadence**: each milestone is a discrete commit. M1 (seed) → M2 (Go allowlist + test) → M3 (catalog) → M4 (frozen-guard) → M5 (cross-skill ref) → M6 (docs-site 4-locale) → M7 (verification + empty-dir cleanup). Manager-develop may consolidate M2-M4 into one commit if the edits are tightly coupled, but MUST NOT skip any.

## §E. Self-verification (run-phase §E — owned by manager-develop, placeholder here)

The `progress.md` §E skeleton is emitted alongside this plan. Run-phase
populates §E.1/§E.2/§E.3; sync-phase populates §E.4; Mx-phase populates
§E.5. See the progress.md skeleton file.

## §F. Milestones

### §F.1 M1 — Commit the 2-file SKILL.md deletion seed

**Goal**: lock in the template-symmetry deletion that the stash-pop staged.

**Actions**:
1. Verify pre-flight §C.1-C.3 pass.
2. Confirm both deletions are unstaged: `git status --short`.
3. `rmdir` the now-empty skill directories if git status shows them as empty (or let git track the deletion):
   - `.claude/skills/moai-design-system/`
   - `internal/template/templates/.claude/skills/moai-design-system/`
4. Commit with subject `feat(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M1 design-system SKILL.md deletion (local + template)`.
5. Body trailer: `Authored-By-Agent: manager-develop`.

**REQs satisfied**: REQ-DSR-001 (template-symmetry), REQ-DSR-008 (empty-dir cleanup, partial).

**Verification**: `git show HEAD --stat` shows exactly the 2 SKILL.md deletions (+ 2 directory removals if git tracked them).

### §F.2 M2 — Go allowlist + test cleanup

**Goal**: remove the `moai-design-system` literal from the static core-skill list and its test case.

**File 1: `internal/cli/doctor_skills.go`**

Current state (lines 20-28):
```go
	// ref (5)
	"moai-ref-api-patterns", "moai-ref-git-workflow", "moai-ref-owasp-checklist",
	"moai-ref-react-patterns", "moai-ref-testing-pyramid",
	// design (1)
	"moai-design-system",
	// FROZEN domain (2)
	"moai-domain-brand-design", "moai-domain-copywriting",
```

Intended edit: remove lines 23-24 (the `// design (1)` comment AND the `"moai-design-system",` literal). Result:
```go
	// ref (5)
	"moai-ref-api-patterns", "moai-ref-git-workflow", "moai-ref-owasp-checklist",
	"moai-ref-react-patterns", "moai-ref-testing-pyramid",
	// FROZEN domain (2)
	"moai-domain-brand-design", "moai-domain-copywriting",
```

**File 2: `internal/cli/doctor_skills_test.go`**

Current state (lines 51-54):
```go
		{
			name:      "valid static core skill moai-design-system returns PASS",
			skillName: "moai-design-system",
			wantClass: "PASS",
		},
```

Intended edit: remove the entire 4-line test-case struct block (lines 51-54 inclusive).

**Commit**: `feat(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M2 remove moai-design-system from doctor_skills allowlist + test`.

**REQs satisfied**: REQ-DSR-002 (allowlist cleanup), REQ-DSR-003 (test consistency).

**Verification**: `go test ./internal/cli/...` passes; `grep -n "moai-design-system" internal/cli/` returns zero matches.

### §F.3 M3 — Catalog entry removal

**Goal**: remove the `moai-design-system` block from catalog.yaml's `design` pack.

**File: `internal/template/catalog.yaml`**

Current state (lines 163-167):
```yaml
                - name: moai-design-system
                  tier: optional-pack:design
                  path: templates/.claude/skills/moai-design-system/
                  hash: dc4a1fc0853c0113277d92fd7e4bbad412f09c414f7c346fc4e581cf3944fd8d
                  version: 1.0.0
```

Intended edit: remove the 5-line block. The `design` pack retains its other entries (e.g. `moai-domain-brand-design`).

**Commit**: `feat(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M3 remove moai-design-system from catalog.yaml design pack`.

**REQs satisfied**: REQ-DSR-004 (catalog entry removal).

**Verification**: `grep -n "moai-design-system" internal/template/catalog.yaml` returns zero matches; YAML still parses (`go run ./cmd/moai` smoke or a yaml-lint).

### §F.4 M4 — Frozen-guard test fixture update

**Goal**: remove the `moai-design-system` path literal from BOTH allowedPaths slices in frozen_guard_test.go.

**File: `internal/design/dtcg/frozen_guard_test.go`**

**Site 1 (line 50)**: in the first `allowedPaths := []string{...}` slice, remove the line:
```go
		".claude/skills/moai-design-system/SKILL.md",
```

**Site 2 (line 104)**: in the second `allowedPaths := []string{...}` slice, remove the line:
```go
		".claude/skills/moai-design-system/SKILL.md",
```

Both sites are mechanical removals of one line each. The slices remain valid Go (trailing commas on the preceding/following entries are preserved).

**Commit**: `test(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M4 remove moai-design-system from frozen_guard_test fixtures`.

**REQs satisfied**: REQ-DSR-005 (frozen-guard fixture update).

**Verification**: `go test ./internal/design/dtcg/...` passes; `grep -n "moai-design-system" internal/design/dtcg/` returns zero matches.

### §F.5 M5 — Cross-skill reference cleanup

**Goal**: remove the `(see moai-design-system)` parenthetical from `moai-workflow-design/SKILL.md`.

**File: `internal/template/templates/.claude/skills/moai-workflow-design/SKILL.md`**

Current state (line 8, in the description frontmatter):
```
  constitutional priority. Use for /moai design workflow — NOT for general design system documentation
  (see moai-design-system).
```

Intended edit: remove the ` (see moai-design-system)` parenthetical. Result:
```
  constitutional priority. Use for /moai design workflow — NOT for general design system documentation.
```

Note: the period after `documentation` is preserved; only the parenthetical is removed. This is a surgical edit — no rewrite of the description prose.

**Commit**: `docs(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M5 remove dangling moai-design-system pointer from workflow-design`.

**REQs satisfied**: REQ-DSR-006 (cross-skill reference cleanup).

**Verification**: `grep -n "moai-design-system" internal/template/templates/.claude/skills/moai-workflow-design/SKILL.md` returns zero matches.

### §F.6 M6 — docs-site 4-locale row removal + header update

**Goal**: remove the `moai-design-system` table row across all 4 locales and update the section header consistently.

**Files (4 locales)**:
- `docs-site/content/en/advanced/skill-guide.md` (line 126)
- `docs-site/content/ko/advanced/skill-guide.md` (line 123)
- `docs-site/content/ja/advanced/skill-guide.md` (line 120)
- `docs-site/content/zh/advanced/skill-guide.md` (line 122)

**Per-locale edit**:
1. Remove the table row `| \`moai-design-system\` | <locale description> |`.
2. Update the section header `### Design (Design System) - 1 skill` (EN; locale-equivalent in others) to reflect the removal. Preferred treatment: rewrite the count to `- 0 skills` OR remove the section entirely if it becomes an empty header with no rows.

**Decision on header treatment**: prefer **removing the section entirely** (header + table + the now-empty body) across all 4 locales, because an empty "Design - 0 skills" section adds noise. BUT apply the SAME treatment in all 4 locales — do NOT remove in EN and keep-as-empty-header in KO. Consistency is the invariant (REQ-DSR-007).

If removing the section breaks a downstream doc-site build (e.g. a TOC anchor), fall back to `- 0 skills` with a one-line note. Verify the docs-site build after the edit.

**Commit**: `docs(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M6 remove moai-design-system from docs-site skill-guide (4-locale)`.

**REQs satisfied**: REQ-DSR-007 (4-locale parity).

**Verification**: `grep -rn "moai-design-system" docs-site/content/` returns zero matches; docs-site 4-locale parity grep (per the docs-site i18n rules) shows symmetric row removal.

### §F.7 M7 — Final verification + empty-directory cleanup confirmation

**Goal**: prove the retirement is complete via a read-only verification batch, then confirm no orphan empty directories remain.

**Verification batch** (single-turn multi-Bash per agent-common-protocol §Parallel Execution):
```bash
# 1. Build
go build ./...

# 2. Targeted tests (cli + dtcg — the two packages this SPEC touches)
go test ./internal/cli/... ./internal/design/dtcg/...

# 3. Full test suite (catch cascading failures; distinguish pre-existing statusline failures)
go test ./... 2>&1 | tee /tmp/dsr-final.txt

# 4. Active-code zero-tolerance grep (historical references in CHANGELOG/.moai excluded by design)
grep -rn "moai-design-system" \
  .claude/ \
  internal/template/templates/ \
  internal/cli/ \
  internal/template/catalog.yaml \
  internal/design/dtcg/ \
  docs-site/content/ \
  2>/dev/null
# Expected: zero matches

# 5. spec-lint on this SPEC
go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001/

# 6. Empty-directory confirmation
ls .claude/skills/moai-design-system/ 2>&1
ls internal/template/templates/.claude/skills/moai-design-system/ 2>&1
# Expected: "No such file or directory" for both (M1 removed them)

# 7. CLI smoke
go run ./cmd/moai doctor --skills 2>&1 | head -20
# Expected: no "moai-design-system" in the skill list; no doctor error
```

**Commit (if any cleanup residue)**: `chore(SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001): M7 verification + cleanup residue`.

**REQs satisfied**: REQ-DSR-008 (empty-dir cleanup, final), REQ-DSR-009 (historical preservation — verified by the grep excluding CHANGELOG/.moai), REQ-DSR-010 (build/test/lint green).

**Verification**: all 7 commands produce expected output. The final test suite may show the pre-existing `internal/statusline` TestCollectMemory failures — those are NOT this SPEC's regression (§B.3).

## §G. Anti-patterns (recap from design.md §D4)

- AP-DSR-001 migration-instead-of-deletion
- AP-DSR-002 asymmetric local/template edit
- AP-DSR-003 over-rewriting workflow-design description
- AP-DSR-004 editing CHANGELOG/RETIRED.md historical records
- AP-DSR-005 inconsistent 4-locale docs-site treatment
- AP-DSR-006 leaving stale `// design (1)` comment

## §H. Cross-references

- spec.md §B.1 (file-surface table — authoritative source for the edits above)
- design.md §D2 (per-file treatment decision table)
- research.md §R1-R3 (rationale evidence + residual risks)
- CLAUDE.local.md §15 (template neutrality), §17 (docs-site i18n), §23 (git workflow), §24 (harness namespace), §25 (template internal-content isolation)
- `SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001` (OutOfScopeRule h3 convention)
- `SPEC-V3R6-HARNESS-NAMESPACE-V2-001/progress.md` line 87 (scope-out provenance)
