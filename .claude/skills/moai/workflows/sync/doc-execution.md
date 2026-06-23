---
description: "Sync Phase 1~2 — Analysis and Planning (with HUMAN GATE 2: Documentation Scope) and Execute Document Synchronization."
user-invocable: false
metadata:
  parent: moai-workflow-sync
  phase: "Phase 1~2: Analysis, Planning, and Document Synchronization"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

### Phase 1: Analysis and Planning

#### Step 1.1: Verify Prerequisites

- .moai/ directory must exist
- .claude/ directory must exist
- Project must be inside a Git repository


#### Step 1.2: Analyze Project Status

- Analyze Git changes: git status, git diff, categorize changed files
- Read project configuration: git_strategy.mode, conversation_language, spec_git_workflow
- Determine synchronization mode from $ARGUMENTS
- Detect branch context: Check current branch name

##### Worktree Context Detection

Detect if the current session is running within a MoAI worktree:
- Check if current git directory path contains `/.moai/worktrees/` component
- OR check if `.moai/worktrees/registry.json` has an active entry for current SPEC-ID
- Store result as `is_worktree_context` boolean for use in Phase 3.4

This affects auto-merge behavior: worktree contexts default to auto-merge.

#### Step 1.3: Project Status Verification

Scan ALL source files (not just changed files) for:

- Broken references and inconsistencies
- Issues with precise locations
- Severity classification (Critical, High, Medium, Low)

#### Step 1.4: Synchronization Plan

Agent: manager-docs subagent

Create synchronization strategy based on Git changes, mode, project verification results, and deployment readiness report from Phase 0. Output: documents to update, SPECs requiring sync, project improvements needed, estimated scope, deployment notes to include in PR.

#### Step 1.5: SPEC-Implementation Divergence Analysis

Purpose: Detect differences between the original SPEC plan and actual implementation to ensure documentation accuracy.

For each SPEC associated with the current sync:

- Step 1.5.1: Load SPEC Documents
  - Read spec.md (requirements), plan.md (implementation plan), acceptance.md (criteria)
  - Extract planned files, planned features, and planned scope

- Step 1.5.2: Analyze Actual Implementation
  - Use git diff and git log to identify all files created, modified, or deleted during the run phase
  - Categorize changes by domain (backend, frontend, tests, config, docs)

- Step 1.5.3: Compare Plan vs Reality
  - Identify files created that were NOT in the original plan.md
  - Identify features or endpoints implemented beyond original spec.md scope
  - Identify planned items that were NOT implemented (deferred or dropped)
  - Identify unplanned refactoring or dependency changes

- Step 1.5.4: Generate Divergence Report
  - Categorize divergences: scope_expansion, unplanned_additions, deferred_items, structural_changes
  - Include: new_directories_created, new_dependencies_added, new_features_implemented
  - This report feeds into Phase 2.2 (SPEC updates) and Phase 2.2.5 (project doc updates)

- Step 1.5.5: Check SPEC Lifecycle Level
  - Read SPEC metadata for lifecycle level (default: spec-first if not specified)
  - Level 1 (spec-first): SPEC will be marked completed with implementation summary appended
  - Level 2 (spec-anchored): SPEC content will be updated to reflect actual implementation
  - Level 3 (spec-as-source): Flag discrepancies as warnings (implementation should match SPEC exactly)

#### Step 1.6: User Approval

<!-- moai:evolvable-start id="gate-sync-2" -->
### HUMAN GATE: Documentation Scope

**Previous phase output:** Divergence analysis showing doc/code drift
**Approval question:** Which documents should be regenerated?
**Cannot proceed until:**
- [ ] User has reviewed divergence report
- [ ] User has approved document regeneration scope
- [ ] User has confirmed PR description draft
<!-- moai:evolvable-end -->

Tool: AskUserQuestion

Display sync plan report and present options:

- Proceed with Sync
- Request Modifications (re-run Phase 1)
- Review Details (show full project results, re-ask)
- Abort (exit with no changes)

### Phase 2: Execute Document Synchronization

#### Step 2.1: Create Safety Backup

Before any modifications:

- Generate timestamp identifier
- Create backup directory: .moai/backups/sync-{timestamp}/
- Copy critical files: README.md, docs/, .moai/specs/
- Verify backup integrity (non-empty directory check)

#### Step 2.2: Document Synchronization

Agent: manager-docs subagent

Input: Approved sync plan, project verification results, changed files list, divergence report from Phase 1.5.

Tasks for manager-docs:

- Reflect changed code in Living Documents
- Auto-generate and update API documentation
- Update README if needed
- Synchronize architecture documents
- Fix project issues and restore broken references
- Update SPEC documents based on divergence analysis and lifecycle level (see Step 2.2.1)
- Detect changed domains and generate domain-specific updates
- Generate sync report: .moai/reports/sync-report-{timestamp}.md

All document updates use conversation_language setting.

##### Step 2.2.1: SPEC Document Update (Based on Divergence Report)

Apply updates based on SPEC lifecycle level detected in Phase 1.5.5:

Level 1 (spec-first):
- Append "Implementation Notes" section to spec.md summarizing actual implementation
- Record scope changes: features added beyond plan, items deferred
- Mark SPEC as completed (no ongoing maintenance expected)

Level 2 (spec-anchored):
- Update spec.md requirements to reflect actual implementation
- Add new EARS-format requirements for features implemented beyond original scope
- Update plan.md with actual implementation steps taken
- Update acceptance.md with new acceptance criteria for added features
- Preserve original requirements with "as-implemented" annotations where changed

Level 3 (spec-as-source):
- Do NOT modify SPEC content
- Generate discrepancy report listing implementation deviations from SPEC
- Flag as warnings in sync report for manual review
- Recommend either updating SPEC or adjusting implementation

#### Step 2.2.5: Project Document Update (Conditional)

Purpose: Update .moai/project/ documents when significant structural changes are detected.

Condition: Execute this step ONLY when the divergence report from Phase 1.5 indicates:
- New directories were created in the project
- New dependencies or technologies were added
- New major features or capabilities were implemented
- Significant architectural changes occurred

Skip condition: If .moai/project/ directory does not exist or contains no files, skip this step entirely.

Agent: manager-docs subagent

Tasks for manager-docs:

- If new directories created: Update structure.md with new directory descriptions and purposes
- If new dependencies added: Update tech.md with new technology stack entries and rationale
- If new features implemented: Update product.md with new feature descriptions and use cases
- If architectural changes: Update structure.md with revised architecture patterns
- If architectural changes: Regenerate .moai/project/codemaps/ via codemaps workflow (workflows/codemaps.md) when significant structural changes (new directories, dependency graph changes, or module reorganization) are detected

Constraints:
- Only update sections relevant to detected changes (do not regenerate entire files)
- Preserve existing content and append or modify incrementally
- Use conversation_language setting for all updates

#### Step 2.3: Post-Sync Quality Verification

Agent: sync-auditor subagent (independent quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2; OR orchestrator verification batch — lint + test + coverage)

Verify synchronization quality against TRUST 5:

- All project links complete
- Documents well-formatted
- All documents consistent
- No credentials exposed
- All SPECs properly linked

#### Step 2.4: Update SPEC Status

Update SPEC status based on lifecycle level and implementation completeness:

- Level 1 (spec-first): Set status to "completed". No further maintenance required.
- Level 2 (spec-anchored): Set status to "completed" if all requirements met, or "in-progress" if partial. Schedule next review based on quarterly maintenance policy.
- Level 3 (spec-as-source): Set status based on implementation-SPEC alignment. Flag discrepancies for resolution.

Record version changes, status transitions, and divergence summary. Include in sync report.

#### Step 2.4.1: GitHub Issue Status Sync

When SPEC metadata contains `issue_number` (non-zero):

- If SPEC status set to "completed":
  - Post completion comment on Issue: `gh issue comment {issue_number} --body "Implementation complete. SPEC-{ID} marked as completed. PR with Fixes #{issue_number} will auto-close this issue on merge."`
- If SPEC status set to "in-progress":
  - Post progress comment: `gh issue comment {issue_number} --body "Partial implementation synced. SPEC-{ID} status: in-progress."`

This step is informational only. Actual Issue closure happens automatically via GitHub's `Fixes #N` mechanism when the PR is merged.
