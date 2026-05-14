# Follow-Up Tasks — SPEC-V3R4-HARNESS-001

This document captures post-merge tasks that are **out of scope of this SPEC's PR** but **MUST be performed** by `manager-git` (or an equivalent commit-only agent) immediately after the foundation SPEC PR is squash-merged into `main`. The tasks are unambiguous and do not require interpretation; `manager-git` is expected to execute them mechanically.

REQ coverage: REQ-HRN-FND-013 (this SPEC's PR MUST NOT modify the three superseded V3R3 SPECs), REQ-HRN-FND-001 (CLI verb path retirement is finalized by this commit chain).

AC coverage: AC-HRN-FND-010 (Verification: `git diff main..HEAD --name-only | grep -E 'SPEC-V3R3-(HARNESS-001|HARNESS-LEARNING-001|PROJECT-HARNESS-001)/'` returns zero matches **inside this SPEC's PR**; the V3R3 SPEC mutation is a separate commit performed AFTER this PR merges).

---

## Owner

`manager-git` subagent. The follow-up is invoked by the orchestrator after this SPEC's PR reaches `MERGED` state.

## Trigger

Precondition: `gh pr view <SPEC-V3R4-HARNESS-001 PR number> --json state -q .state` returns `MERGED`.

## Steps (verbatim — execute in order)

### 1. Branch creation

```bash
git fetch origin
git checkout -b chore/SPEC-V3R3-status-transition origin/main
```

The branch name `chore/SPEC-V3R3-status-transition` is canonical for this follow-up. Do not use a different name.

### 2. Update three V3R3 SPEC frontmatters

For each of the three files below, locate the YAML frontmatter (between the opening `---` and closing `---` markers near the top of the file) and apply the following two changes:

- Change the `status:` field value to `superseded` (replace whatever current value is present — typically `completed`).
- Add a new field `superseded_by: SPEC-V3R4-HARNESS-001` directly after the `status:` line. If `superseded_by:` already exists (rare), update its value to `SPEC-V3R4-HARNESS-001`.

Target files:

1. `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`
2. `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`
3. `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md`

### 3. Add a HISTORY entry to each of the three V3R3 SPECs

After the existing HISTORY table (search for the line beginning with `| Version |` near the top of each file's body), append a new row whose contents are:

```
| (n+1).0.0 | <YYYY-MM-DD of merge commit> | manager-git | Status transition: superseded by SPEC-V3R4-HARNESS-001. The V3R4 foundation SPEC consolidates the three V3R3 harness SPECs into a single architecture; the `moai harness` CLI verb path is retired per BC-V3R4-HARNESS-001-CLI-RETIREMENT, and the 4-tier ladder + 5-Layer Safety architecture are preserved verbatim. See SPEC-V3R4-HARNESS-001 for the current authoritative contract. |
```

Where:

- `(n+1)` is the next major version number after the SPEC's current latest HISTORY row. If the latest row is `| 1.0.0 |`, use `2.0.0`. If `2.0.0`, use `3.0.0`. The HISTORY ordering convention (newest-first or newest-last) MUST be preserved per the file's existing convention.
- `<YYYY-MM-DD>` is the ISO-8601 date of the SPEC-V3R4-HARNESS-001 PR merge commit. Derive via `git log -1 --format=%cs <merge-commit-sha>`.

### 4. Commit and PR

```bash
git add .moai/specs/SPEC-V3R3-HARNESS-001/spec.md \
        .moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md \
        .moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md

git commit -m "chore(spec): transition three V3R3 harness SPECs to superseded per SPEC-V3R4-HARNESS-001

V3R4 foundation SPEC merged at <merge-commit-sha>. The three V3R3 SPECs
(HARNESS-001 meta-skill, HARNESS-LEARNING-001 4-tier ladder + 5-Layer
Safety, PROJECT-HARNESS-001 16Q socratic interview) are consolidated
into the V3R4 family. Their status transitions to 'superseded' and
their HISTORY tables record the V3R4 takeover.

Refs: SPEC-V3R4-HARNESS-001 (BC-V3R4-HARNESS-001-CLI-RETIREMENT)
Follow-up: .moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md

🗿 MoAI <email@mo.ai.kr>"

git push -u origin chore/SPEC-V3R3-status-transition
```

### 5. PR creation

```bash
gh pr create \
  --base main \
  --title "chore(spec): transition three V3R3 harness SPECs to superseded per SPEC-V3R4-HARNESS-001" \
  --body "$(cat <<'EOF'
## Summary

- Transitions three V3R3 harness SPECs to `status: superseded` and adds `superseded_by: SPEC-V3R4-HARNESS-001` per the foundation SPEC's `supersedes:` frontmatter.
- HISTORY tables updated to record the V3R4 takeover with the merge commit SHA.

## Why this is a separate PR

REQ-HRN-FND-013 of SPEC-V3R4-HARNESS-001 explicitly prohibits modifying the three V3R3 SPEC files inside the foundation SPEC's own PR. The status transition is delegated to this follow-up commit to keep the foundation PR's diff scope clean.

## Verification

- Each V3R3 SPEC frontmatter shows \`status: superseded\` and \`superseded_by: SPEC-V3R4-HARNESS-001\`.
- Each V3R3 SPEC HISTORY table records the transition with the V3R4 merge date.
- \`git diff main..HEAD --name-only\` returns exactly three paths: the three V3R3 \`spec.md\` files.

## Test plan

- [x] Frontmatter \`status\` field changed for all three SPECs
- [x] \`superseded_by\` field added for all three SPECs
- [x] HISTORY entry appended for all three SPECs
- [x] No file outside \`.moai/specs/SPEC-V3R3-*/spec.md\` modified

Refs: SPEC-V3R4-HARNESS-001, BC-V3R4-HARNESS-001-CLI-RETIREMENT

🗿 MoAI <email@mo.ai.kr>
EOF
)"
```

### 6. Merge strategy

Squash merge per Enhanced GitHub Flow (`CLAUDE.local.md` §18). After merge, the V3R3 SPEC frontmatter changes are recorded in `main` as a single commit. No further follow-up commits are needed for this SPEC family.

---

## Out of scope for this follow-up

The following are **NOT** part of this follow-up commit; they belong to a future SPEC:

- Physical deletion of `internal/cli/harness.go` or `internal/cli/harness_test.go`. They remain as deprecation markers per SPEC-V3R4-HARNESS-001 §2.1.
- Modification of `.claude/rules/moai/design/constitution.md` (FROZEN file).
- Changes to `.claude/skills/moai-harness-learner/` or `.claude/skills/moai-meta-harness/` skill bodies beyond text annotations (already performed by the SPEC-V3R4-HARNESS-001 PR per §10 exclusion #10).

---

REQ traceability: REQ-HRN-FND-013, REQ-HRN-FND-001
AC traceability: AC-HRN-FND-010
