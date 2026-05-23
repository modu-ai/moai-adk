---
id: SPEC-V3R6-LEGACY-CLEANUP-001
title: "Acceptance Criteria — v2.x agency keyword residual cleanup"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: Medium
tags: "cleanup, legacy, v3-roadmap, sprint-2, docs, brand-design"
phase: "v3.0.0"
module: "docs"
lifecycle: spec-anchored
tier: L
---

# Acceptance Criteria — SPEC-V3R6-LEGACY-CLEANUP-001

All criteria below are **binary-verifiable** with concrete commands. Each AC links back to one or more REQ-LCL-XXX from spec.md §B.

## AC-LCL-001 — Backup directory created with all 31 in-scope files

**Linked REQ**: REQ-LCL-001, REQ-LCL-002

**Given**: The cleanup process is invoked via `/moai run SPEC-V3R6-LEGACY-CLEANUP-001`
**When**: Milestone M1 completes
**Then**: A backup directory matching `.moai/backups/legacy-cleanup-*/` exists, containing a `manifest.json` listing exactly 31 entries with `path`, `sha256`, and `bytes` fields per entry.

**Verification**:
```bash
ls -d .moai/backups/legacy-cleanup-*/ | wc -l   # = 1
ls .moai/backups/legacy-cleanup-*/manifest.json | wc -l   # = 1
jq 'length' .moai/backups/legacy-cleanup-*/manifest.json   # = 31
jq '.[0] | keys' .moai/backups/legacy-cleanup-*/manifest.json   # ["bytes", "path", "sha256"]
```

---

## AC-LCL-002 — PRESERVE paths byte-identical after cleanup

**Linked REQ**: REQ-LCL-004

**Given**: Cleanup completes (all M1-M4 milestones)
**When**: Running SHA256 comparison on PRESERVE path sample (10 files chosen at random across the PRESERVE set)
**Then**: All 10 sample SHA256 hashes match their pre-edit baseline (captured before M1 commit).

**Verification**:
```bash
# Pre-cleanup baseline capture (run BEFORE M1) — 10 real PRESERVE paths verified via filesystem at iter-2 plan-auditor B3 fix-forward:
shasum -a 256 \
  .moai/design/v3-legacy/research/findings-wave1-ui-ux.md \
  .moai/design/v3-legacy/research/audit-report-v3.md \
  .moai/design/v3-research/findings-wave1-ui-ux.md \
  .moai/brain/IDEA-001/ideation.md \
  .moai/marketing/launch-kit/show-hn.md \
  .moai/specs/SPEC-AGENCY-ABSORB-001/audit-report-v2.md \
  .moai/archive/skills/v2.16/moai-domain-database/SKILL.md \
  .moai/plans/staged-forging-manatee.md \
  .moai/specs/_archive/README.md \
  internal/cli/migrate_agency_test.go \
  > /tmp/preserve-baseline.txt

# Post-cleanup verification:
shasum -a 256 -c /tmp/preserve-baseline.txt
# Expected: all 10 OK
```

---

## AC-LCL-003 — In-scope keyword count ≤5 after cleanup

**Linked REQ**: REQ-LCL-013

**Given**: Cleanup completes (M1-M4)
**When**: Running grep over the 31 in-scope files
**Then**: Exit count is ≤5 (the residual occurrences are legitimate "v2.x agency (retired, see SPEC-AGENCY-ABSORB-001)" references introduced by REQ-LCL-005 category 1 strategy).

**Verification**:
```bash
COUNT=$(grep -rln -E '\bagency\b|\.agency/|/agency/' \
  CHANGELOG.md CLAUDE.md README.md README.ko.md README.ja.md README.zh.md \
  .claude/skills/moai-domain-brand-design/SKILL.md \
  .claude/skills/moai-domain-copywriting/SKILL.md \
  .claude/skills/moai-workflow-gan-loop/SKILL.md \
  .claude/skills/moai/workflows/design.md \
  .claude/rules/moai/design/constitution.md \
  docs-site/content/ko/design/_index.md docs-site/content/ko/design/gan-loop.md \
  docs-site/content/ko/design/getting-started.md docs-site/content/ko/design/migration-guide.md \
  docs-site/content/ko/workflow-commands/moai-design.md \
  docs-site/content/en/design/_index.md docs-site/content/en/design/gan-loop.md \
  docs-site/content/en/design/getting-started.md docs-site/content/en/design/migration-guide.md \
  docs-site/content/en/workflow-commands/moai-design.md \
  docs-site/content/ja/design/_index.md docs-site/content/ja/design/gan-loop.md \
  docs-site/content/ja/design/getting-started.md docs-site/content/ja/design/migration-guide.md \
  docs-site/content/ja/workflow-commands/moai-design.md \
  docs-site/content/zh/design/_index.md docs-site/content/zh/design/gan-loop.md \
  docs-site/content/zh/design/getting-started.md docs-site/content/zh/design/migration-guide.md \
  docs-site/content/zh/workflow-commands/moai-design.md \
  2>/dev/null | wc -l)
[ "$COUNT" -le 5 ] && echo PASS || echo "FAIL (count=$COUNT)"
```

---

## AC-LCL-004 — Hugo build exit 0

**Linked REQ**: REQ-LCL-011

**Given**: Cleanup completes (M1-M4)
**When**: Running `hugo --source docs-site --quiet`
**Then**: Exit status is 0 (docs-site Hugo build remains green).

**Verification**:
```bash
hugo --source docs-site --quiet
echo $?   # Expected: 0
```

---

## AC-LCL-005 — Go test PASS rate maintained (zero regression)

**Linked REQ**: REQ-LCL-012

**Given**: Cleanup completes (M1-M4). This SPEC modifies no `.go` files.
**When**: Running `go test ./...`
**Then**: All tests PASS; the post-cleanup PASS count equals the pre-cleanup PASS count (delta = 0).

**Verification**:
```bash
# Pre-cleanup baseline (run BEFORE M1):
go test -count=1 ./... 2>&1 | grep -E '^(ok|FAIL|---)' | tee /tmp/pre-test-baseline.txt
PRE_PASS=$(grep -c '^ok' /tmp/pre-test-baseline.txt)

# Post-cleanup verification:
go test -count=1 ./... 2>&1 | grep -E '^(ok|FAIL|---)' | tee /tmp/post-test-results.txt
POST_PASS=$(grep -c '^ok' /tmp/post-test-results.txt)

[ "$PRE_PASS" = "$POST_PASS" ] && echo "PASS (count=$POST_PASS)" || echo "FAIL (pre=$PRE_PASS post=$POST_PASS)"
```

---

## AC-LCL-006 — 4-locale parity preserved (5 + 5 + 5 + 5 = 20 docs-site files)

**Linked REQ**: REQ-LCL-009, REQ-LCL-010

**Given**: Cleanup completes (M1-M4)
**When**: Counting `agency`-keyword-containing files per locale directory
**Then**: The per-locale grep count for ko, en, ja, zh under `docs-site/content/<loc>/design/` + `docs-site/content/<loc>/workflow-commands/` is **equal** across all 4 locales (symmetric residual, REQ-LCL-010).

**Verification**:
```bash
for loc in ko en ja zh; do
  COUNT=$(grep -rln agency \
    "docs-site/content/$loc/design" \
    "docs-site/content/$loc/workflow-commands" 2>/dev/null | wc -l)
  echo "$loc: $COUNT"
done
# Expected: all 4 lines show the same count (e.g., "0" each, or "1" each if 1 legitimate retired-reference per locale)
```

---

## AC-LCL-007 — CHANGELOG pre-v3.0 entries untouched

**Linked REQ**: REQ-LCL-007

**Given**: Cleanup completes (M4)
**When**: Comparing CHANGELOG.md pre-v3.0 sections pre/post cleanup
**Then**: All entries dated before v3.0.0 are byte-identical to baseline (SHA256 match on the pre-v3.0 line range).

**Verification**:
```bash
# Identify pre-v3.0 line range — the line BEFORE the first occurrence of "## [v3" or "## [3.0"
PRE_V3_END=$(grep -n -E '^## \[(v?3\.0|Unreleased)' CHANGELOG.md | head -1 | cut -d: -f1)
PRE_V3_END=$((PRE_V3_END - 1))

# Extract pre-v3.0 section from current CHANGELOG.md (S1 fix: sed → head)
head -n ${PRE_V3_END} CHANGELOG.md > /tmp/pre-v3.0-current.txt

# Compare to backup
head -n ${PRE_V3_END} .moai/backups/legacy-cleanup-*/CHANGELOG.md > /tmp/pre-v3.0-backup.txt
diff /tmp/pre-v3.0-current.txt /tmp/pre-v3.0-backup.txt && echo PASS || echo FAIL
```

---

## AC-LCL-008 — Backup manifest SHA256 verifies original content

**Linked REQ**: REQ-LCL-002

**Given**: M1 completes (backup manifest generated)
**When**: Validating manifest.json entries against the backed-up files
**Then**: Every `sha256` field in `manifest.json` matches the SHA256 of the corresponding backup file.

**Verification**:
```bash
MANIFEST=$(ls .moai/backups/legacy-cleanup-*/manifest.json)
BACKUP_ROOT=$(dirname "$MANIFEST")

jq -r '.[] | "\(.sha256)  \(.path)"' "$MANIFEST" | while read sha path; do
  ACTUAL=$(shasum -a 256 "$BACKUP_ROOT/$path" 2>/dev/null | cut -d' ' -f1)
  [ "$sha" = "$ACTUAL" ] || echo "MISMATCH: $path (expected=$sha actual=$ACTUAL)"
done
# Expected: no MISMATCH lines printed
```

---

## AC-LCL-009 — No production Go code modified

**Linked REQ**: REQ-LCL-014 (iter-2 fix-forward, plan-auditor S5 traceability — §C exclusion #1 promoted to REQ-LCL-014 [Unwanted])

**Given**: Cleanup completes (M1-M4)
**When**: Listing modified Go files since this SPEC's plan-phase entry commit
**Then**: Zero `.go` files appear in the modification list.

**Verification**:
```bash
# Compare HEAD to plan-phase entry (the spec.md commit)
PLAN_COMMIT=$(git log --pretty=format:'%H' -1 -- .moai/specs/SPEC-V3R6-LEGACY-CLEANUP-001/spec.md)
git diff --name-only "$PLAN_COMMIT" HEAD -- '*.go' | wc -l
# Expected: 0
```

---

## AC-LCL-010 — No template mirror files modified (deferred to SPEC-V3R6-LEGACY-CLEANUP-002)

**Linked REQ**: REQ-LCL-015 (iter-2 fix-forward, plan-auditor S5 traceability — §C exclusion #2 promoted to REQ-LCL-015 [Unwanted])

**Given**: Cleanup completes (M1-M4)
**When**: Listing modified files under `internal/template/templates/`
**Then**: Zero files under `internal/template/templates/` appear in the modification list (this SPEC scope is source-only; mirror cascade is a separate SPEC).

**Verification**:
```bash
PLAN_COMMIT=$(git log --pretty=format:'%H' -1 -- .moai/specs/SPEC-V3R6-LEGACY-CLEANUP-001/spec.md)
git diff --name-only "$PLAN_COMMIT" HEAD -- 'internal/template/templates/' | wc -l
# Expected: 0
```

---

## AC-LCL-011 — Locale file count unchanged (no add/remove) — iter-2 git-diff approach

**Linked REQ**: REQ-LCL-009

**Given**: Cleanup completes (M1-M4)
**When**: Listing locale-directory files added or deleted between plan-commit and HEAD
**Then**: Zero files added AND zero files deleted under any `docs-site/content/<locale>/` directory.

**iter-2 fix-forward rationale (plan-auditor B4)**: Original iter-1 verification compared `find .moai/backups/.../docs-site/content/$loc -name "*.md" | wc -l` (5 in-scope files per locale) vs `find docs-site/content/$loc -name "*.md" | wc -l` (30+ files per locale = full directory). Scope mismatch made PRE != POST always FAIL. Option (b) — `git diff --diff-filter=AD` between plan-commit and HEAD — is backup-free, scope-precise, and binary-verifiable.

**Verification**:
```bash
PLAN_COMMIT=$(git log --pretty=format:'%H' -1 -- .moai/specs/SPEC-V3R6-LEGACY-CLEANUP-001/spec.md)
for loc in ko en ja zh; do
  ADDED=$(git diff --diff-filter=A --name-only "$PLAN_COMMIT" HEAD -- "docs-site/content/$loc/" | wc -l | tr -d ' ')
  DELETED=$(git diff --diff-filter=D --name-only "$PLAN_COMMIT" HEAD -- "docs-site/content/$loc/" | wc -l | tr -d ' ')
  if [ "$ADDED" = "0" ] && [ "$DELETED" = "0" ]; then
    echo "$loc: PASS (no add/remove)"
  else
    echo "$loc: FAIL (+$ADDED added, -$DELETED deleted)"
  fi
done
# Expected: all 4 lines show PASS
```

---

## AC-LCL-012 — Definition of Done

This SPEC is COMPLETE when ALL of the following hold:

1. AC-LCL-001 through AC-LCL-011 all return PASS.
2. Git log shows 5 commits on `main` referencing SPEC-V3R6-LEGACY-CLEANUP-001 (1 plan + 4 milestone commits).
3. SPEC frontmatter `status` updated from `draft` to `implemented`.
4. CHANGELOG.md `[Unreleased]` section adds a single line: `- SPEC-V3R6-LEGACY-CLEANUP-001: v2.x agency keyword residual cleanup (scope C, 31 files)` (CHANGELOG entry is added by SYNC phase, not by run-phase — per SPEC-V3R6-CHANGELOG-CLEANUP-001 M3 B12 standing rule).
5. 4 follow-up SPEC candidates documented in spec.md §A.6 (LEGACY-CLEANUP-002 through 005) for the orchestrator to triage post-merge.

**Verification (composite)**:
```bash
# Composite checklist — should print "ALL PASS" if all 11 ACs PASS
# (script left to implementer; sample skeleton below)

for ac in AC-LCL-001 AC-LCL-002 AC-LCL-003 AC-LCL-004 AC-LCL-005 \
         AC-LCL-006 AC-LCL-007 AC-LCL-008 AC-LCL-009 AC-LCL-010 AC-LCL-011; do
  echo "Verifying $ac..."
  # Run the AC-specific verification block (see each AC above)
done
```

---

## Edge Cases

### Edge Case 1: A skill description rewrite changes the skill's intent

**Scenario**: REQ-LCL-006 category 3 (skill frontmatter `description:` rewrite) might unintentionally change the skill's purpose.
**Mitigation**: Implementer SHOULD review the rewritten description against the skill's body content to ensure semantic alignment. If misaligned, escalate to user via blocker report; do NOT proceed.

### Edge Case 2: A docs-site locale has an orphaned reference

**Scenario**: ko file has 3 `agency` mentions; en/ja/zh siblings have 2 each. Cleanup mirrors edits 1:1 but ko still has 1 residual.
**Mitigation**: Per REQ-LCL-010, the residual ko reference MUST be cleaned to restore parity, OR an equivalent reference added to en/ja/zh to maintain symmetric residual (typically the cleaner approach is to remove the ko residual).

### Edge Case 3: Hugo build breaks after a docs-site edit

**Scenario**: Markdown front-matter parsing fails after an edit (e.g., quotes not escaped).
**Mitigation**: Run `hugo --source docs-site --quiet` after each locale group; if exit ≠ 0, revert the most recent edit and retry with stricter front-matter handling.

### Edge Case 4: Pre-v3.0 CHANGELOG entry mistakenly edited

**Scenario**: M4 T4.1 classification fails; a v2.5 entry is edited.
**Mitigation**: AC-LCL-007 verification (SHA256 diff of pre-v3.0 line range) catches this. If failed, revert M4 and re-classify.

### Edge Case 5: Backup directory creation fails (disk full)

**Scenario**: M1 T1.1 backup creation fails before manifest generation.
**Mitigation**: REQ-LCL-003 mandates halt + error report. Implementer fixes disk space and retries M1 from scratch.

---

## Quality Gate Criteria

- TRUST 5 Tested: All 11 AC verification commands documented above ✓
- TRUST 5 Readable: Each AC follows Given/When/Then format ✓
- TRUST 5 Unified: All verification commands use Bash + grep + shasum + git diff (no `sed`/`awk`/`find` -exec) ✓
- TRUST 5 Secured: Backup manifest SHA256 (AC-LCL-008) provides cryptographic integrity proof ✓
- TRUST 5 Trackable: Each AC links to one or more REQ-LCL-XXX from spec.md §B ✓
