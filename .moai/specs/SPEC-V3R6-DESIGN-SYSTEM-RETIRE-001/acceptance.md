# acceptance.md — Acceptance Criteria

> Given-When-Then scenarios mapped to the §D requirements (spec.md).
> Each AC is independently verifiable via grep/build/test commands enumerated
> in plan.md §F milestones. Minimum 2 GWT scenarios required; this SPEC
> carries 10 ACs to cover every REQ + the symmetry/parity invariants.

## §D. AC Matrix

| AC ID | REQ | Severity | Description |
|-------|-----|----------|-------------|
| AC-DSR-001 | REQ-DSR-001 | MUST | Template-symmetry of the SKILL.md deletion |
| AC-DSR-002 | REQ-DSR-002 | MUST | Go allowlist literal removed + stale comment removed |
| AC-DSR-003 | REQ-DSR-003 | MUST | Go test case removed in same change as allowlist |
| AC-DSR-004 | REQ-DSR-004 | MUST | Catalog entry block removed |
| AC-DSR-005 | REQ-DSR-005 | MUST | Frozen-guard test fixtures updated (both sites) |
| AC-DSR-006 | REQ-DSR-006 | MUST | Cross-skill reference parenthetical removed (surgical) |
| AC-DSR-007 | REQ-DSR-007 | MUST | docs-site 4-locale parity (symmetric row removal + header treatment + global 32→31 count update) |
| AC-DSR-008 | REQ-DSR-008 | MUST | Empty skill directories removed |
| AC-DSR-009 | REQ-DSR-009 | MUST | Historical references preserved (no CHANGELOG/.moai archival edits) |
| AC-DSR-010 | REQ-DSR-010 | MUST | Build + targeted tests + spec-lint green |

## §D.1 Given-When-Then scenarios

### AC-DSR-001 — Template-symmetry of the deletion

**Given** the retirement commits M1-M7 are applied to the working tree,
**When** the verifier runs `git log --oneline --diff-filter=D -- .claude/skills/moai-design-system/SKILL.md internal/template/templates/.claude/skills/moai-design-system/SKILL.md`,
**Then** BOTH paths appear as deletions in the same commit (M1) OR in commits that together cover both paths symmetrically, AND neither path exists on disk post-retirement.

**Verification command**:
```bash
test ! -f .claude/skills/moai-design-system/SKILL.md \
  && test ! -f internal/template/templates/.claude/skills/moai-design-system/SKILL.md \
  && echo "AC-DSR-001 PASS" || echo "AC-DSR-001 FAIL"
```

### AC-DSR-002 — Go allowlist literal + stale comment removed

**Given** the M2 commit is applied,
**When** the verifier runs `grep -n "moai-design-system\|// design" internal/cli/doctor_skills.go`,
**Then** the command returns zero matches — neither the literal nor the `// design (1)` comment line remains.

### AC-DSR-003 — Go test case removed in same change

**Given** the M2 commit is applied,
**When** the verifier runs `go test ./internal/cli/...`,
**Then** the test suite passes (exit 0), AND `grep -n "moai-design-system" internal/cli/doctor_skills_test.go` returns zero matches.

### AC-DSR-004 — Catalog entry block removed

**Given** the M3 commit is applied,
**When** the verifier runs `grep -n "moai-design-system" internal/template/catalog.yaml`,
**Then** the command returns zero matches, AND the catalog.yaml still parses as valid YAML (smoke: `go run ./cmd/moai` does not error on catalog load).

### AC-DSR-005 — Frozen-guard test fixtures updated (both sites)

**Given** the M4 commit is applied,
**When** the verifier runs `grep -cn "moai-design-system" internal/design/dtcg/frozen_guard_test.go`,
**Then** the count is 0 (both line-50 and line-104 sites removed), AND `go test ./internal/design/dtcg/...` passes.

### AC-DSR-006 — Cross-skill reference removed (surgical)

**Given** the M5 commit is applied,
**When** the verifier runs `grep -n "see moai-design-system\|moai-design-system" internal/template/templates/.claude/skills/moai-workflow-design/SKILL.md`,
**Then** the command returns zero matches, AND the description frontmatter otherwise reads identically to pre-retirement (only the parenthetical removed — verified by `git show` diff showing exactly the parenthetical line changed, no surrounding prose rewritten).

### AC-DSR-007 — docs-site 4-locale parity

**Given** the M6 commit is applied,
**When** the verifier runs `grep -rn "moai-design-system" docs-site/content/`,
**Then** the command returns zero matches across all 4 locales, AND the section-header treatment is consistent — for each locale, EITHER the `### Design (Design System) - 1 skill` section is removed entirely OR rewritten to the locale-equivalent of `- 0 skills`, with the SAME choice applied in all 4 locales.

**4-locale parity verification** (per CLAUDE.local.md §17):
```bash
for loc in en ko ja zh; do
  echo "=== $loc ==="
  grep -c "moai-design-system" docs-site/content/$loc/advanced/skill-guide.md
done
# Expected: 0 0 0 0 (four zeros)
```

**Global skill-count update (32 → 31) — per-locale line-class coverage**:
the count-referencing prose lives in THREE line-classes, but the line-classes
are NOT all present in every locale. `en`/`ko`/`ja` each carry all THREE
classes (total-of-N, umbrella-includes-N, load-all-N-tokens); `zh` carries
only TWO (total-of-N, load-all-N-tokens — zh has no umbrella line). The AC
applies per-locale to whichever of the three line-classes each locale
actually carries; all referenced counts MUST be decremented from 32 to 31
for internal consistency (else the doc claims 32 while enumerating 31):
1. The "total of N skills" line (en:65 / ko:62 / ja:59 / zh:61) — present in
   ALL FOUR locales; both the total count AND the "31 specialized" sub-count
   decrement (`31 → 30`).
2. The "umbrella skill is included in the N total" line (en:125-equivalent /
   ko:125 / ja:122) — present in `en`/`ko`/`ja` ONLY (zh has no umbrella
   line); the `32`/`N` total decrements to `31`.
3. The "load all N skills = ~160,000 tokens" line (en:168 / ko:165 / ja:160 /
   zh:161) — present in ALL FOUR locales; the `N` in "all N skills" decrements
   from 32 to 31 (the token estimate stays approximate; the count reference
   is the load-bearing fix).

Per-locale verification:
```bash
for loc in en ko ja zh; do
  echo "=== $loc ==="
  grep -nE "(32|31) (skills|스킬|スキル|个技能)" docs-site/content/$loc/advanced/skill-guide.md
done
# Expected: NO line references "32" as the current total — only "31"
```

### AC-DSR-008 — Empty skill directories removed

**Given** the M1 + M7 commits are applied,
**When** the verifier runs `ls .claude/skills/moai-design-system/ 2>&1; ls internal/template/templates/.claude/skills/moai-design-system/ 2>&1`,
**Then** both commands return "No such file or directory" — no orphan empty skill directories remain.

### AC-DSR-009 — Historical references preserved

**Given** the full retirement (M1-M7) is applied,
**When** the verifier runs `grep -rln "moai-design-system" CHANGELOG.md .moai/specs/SPEC-V3R2-WF-001/ .moai/research/ .moai/archive/ .moai/release/ .moai/decisions/ .moai/brain/ .moai/design/v3-redesign/ .moai/state/ .moai/backups/ .claude/worktrees/`,
**Then** matches ARE expected and acceptable in these historical paths (the files are records-of-the-past), AND `git diff main --stat -- CHANGELOG.md .moai/specs/SPEC-V3R[2-5]-* .moai/archive/ .moai/research/` shows ZERO lines changed in any historical file.

**Rationale**: this AC is a negative-space check — it verifies the retirement did NOT accidentally edit archival records. The grep finding matches is correct behavior; the `git diff --stat` showing zero changes is the load-bearing assertion.

### AC-DSR-010 — Build + targeted tests + spec-lint green

**Given** the full retirement (M1-M7) is applied,
**When** the verifier runs the M7 verification batch (plan.md §F.7),
**Then**:
- `go build ./...` exits 0.
- `go test ./internal/cli/... ./internal/design/dtcg/...` exits 0 (the two packages this SPEC touches).
- `go test ./...` exits 0 for all packages EXCEPT the pre-existing `internal/statusline` TestCollectMemory / TestCollectMemory_AutoCompactScaling failures (plan.md §B.3 — these are NOT this SPEC's regression; they predate the retirement).
- `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001/` returns 0 findings.
- `go run ./cmd/moai doctor --skills` does not list `moai-design-system` and does not error.

## §D.2 Severity and traceability

All 10 ACs are **MUST-pass** (severity MUST). The retirement is atomic —
partial completion (e.g. removing the literal but leaving the test case,
or editing EN but not KO) leaves the codebase in an inconsistent state.
If any MUST AC fails, run-phase MUST NOT mark the SPEC complete.

## §D.3 Indirect verification (forward-looking)

- **AC-DSR-011 (indirect)**: after the next `moai update` cycle on a downstream user project, the user's `.claude/skills/moai-design-system/` directory should not be re-created from the template (because the catalog entry is gone). This is a downstream verification, out of scope for this SPEC's run-phase, but recorded for the sync-phase / Mx-phase awareness.
- **AC-DSR-012 (indirect)**: the `doctor --skills` command's static core-skill list count decreases by exactly 1 (the `design` category goes from 1 to 0 entries). Verifiable via `moai doctor --skills --json | jq '.core_skills | length'` before/after.

## §D.4 Closure gates

The SPEC is close-eligible when:
1. All 10 MUST ACs pass (§D.1).
2. The active-code zero-tolerance grep returns zero matches (plan.md §F.7 command 4).
3. spec-lint on this SPEC returns 0 findings.
4. The pre-existing statusline test failures are acknowledged as out-of-scope (not regression).

## §D.5 Quality gate criteria

- **TRUST 5 — Tested**: targeted Go tests pass (AC-DSR-003, AC-DSR-005); full suite distinguished from pre-existing failures (AC-DSR-010).
- **TRUST 5 — Readable**: surgical edits (AC-DSR-006 enforces no over-rewrite); stale comments removed (AC-DSR-002).
- **TRUST 5 — Unified**: template-neutrality symmetry (AC-DSR-001); 4-locale docs-site parity (AC-DSR-007).
- **TRUST 5 — Secured**: no new attack surface (pure deletion + fixture cleanup).
- **TRUST 5 — Trackable**: per-milestone commits with `Authored-By-Agent: manager-develop` trailer; spec-lint clean.

## §D.6 Definition of Done

The SPEC reaches `implemented` status (manager-docs sync-phase) when §D.4
closure gates 1-3 are met. It reaches `completed` status (Mx-phase) when
the 4-phase close (sync_commit_sha + mx_commit_sha) is recorded in
progress.md §E.4/§E.5.
