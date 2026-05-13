---
id: SPEC-V3R4-STATUS-LIFECYCLE-001
version: "0.1.0"
status: in-progress
created_at: 2026-05-13
updated_at: 2026-05-13
author: manager-spec
priority: High
labels: [spec-status, lifecycle, lint, hook, github-actions, automation, drift-prevention]
issue_number: null
depends_on: [SPEC-STATUS-AUTO-001]
related_specs: [SPEC-V3R4-CATALOG-001, SPEC-V3R4-CATALOG-002, SPEC-V3R2-WF-003, SPEC-V3R2-WF-004]
---

# SPEC-V3R4-STATUS-LIFECYCLE-001: SPEC Status Lifecycle Automation — 7-Layer Defense in Depth

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec | Initial plan-in-main draft (Step 1a per spec-workflow.md § SPEC Phase Discipline). Catalyst: PR #871 retrofit (10 Sprint 12 SPECs). 7-Layer Defense in Depth proposed. |

---

## §0. SPEC ID Justification

**Chosen ID**: `SPEC-V3R4-STATUS-LIFECYCLE-001`

**Rationale**:
- PR #871 (OPEN, 2026-05-13) explicitly references `SPEC-V3R4-STATUS-LIFECYCLE-001` as the planned systemic resolution. Aligning with the catalyst PR's external reference preserves cross-document traceability.
- V3R4 series is the active v3.0 release-prep wave (CATALOG-001 PR #864 merged 2026-05-12, CATALOG-002 PR #867+#869 merged 2026-05-12). This SPEC continues that wave.
- V3R3 series was largely closed out (CI-AUTONOMY-001 last merged 2026-05-09; STATUSLINE-FALLBACK-001 closed 2026-05-10). Placing this SPEC in V3R3 would dilute that closure boundary.
- The drift pattern this SPEC addresses is intrinsically tied to the v3.0 release-prep sprint cadence (Sprint 11 → Sprint 12 → ...), and V3R4 is that series.

**Alternative rejected**: `SPEC-V3R3-STATUS-LIFECYCLE-001` — would create cross-reference drift with PR #871, and V3R3 closure intent (per CIAUT-001 wrap-up) would be muddied.

---

## §1. Environment

### 1.1 Current System State

MoAI-ADK manages SPEC documents in `.moai/specs/SPEC-<ID>/` directories. Each SPEC carries a `status` field whose canonical lifecycle is:

```
draft → planned → in-progress → implemented → completed
                                            ↓
                                       superseded | archived | rejected
```

Status transitions occur at four organic milestones:
1. **Plan PR merge** → `draft` to `planned`
2. **Run PR merge** (AC partial) → `planned` to `in-progress`
3. **Run PR merge** (AC complete) → `planned`/`in-progress` to `implemented`
4. **Sync PR merge** → `implemented` to `completed`

### 1.2 Observable Drift Pattern

Status drift retrofit PRs occur on a recurring cadence. Catalogued evidence:

| Retrofit PR | Date | Scope | Trigger |
|-------------|------|-------|---------|
| #818 | 2026-05-10 | WF-004 status retrofit | Sprint 8 closeout |
| #844 | 2026-05-11 | ORC-001/005, RT-005, BRAIN-001 | Sprint 9-10 closeout |
| #856 | 2026-05-12 | RT-007 + 4 others | Sprint 11 closeout |
| #866 | 2026-05-12 | 20 SPECs (implemented/in_review → completed) | metadata-only sweep |
| **#871** | **2026-05-13** | **10 Sprint 12 SPECs (draft → planned/in-progress)** | **Catalyst for this SPEC** |

Pattern frequency: Every 2-7 days a retrofit PR is required. Manual cost per retrofit: ~15 minutes review + merge. Aggregate cost over 30 days: ~3 hours pure rework, plus cognitive load and audit-trail noise.

### 1.3 Root-Cause Topology

Seven independent system failures contribute to drift:

| Layer | Failure | Asset |
|-------|---------|-------|
| L1 — Standard | No single source of truth for status enum; 6 vs 5 vs 8 values used in different files | `internal/spec/status.go:13` (6 values), `.claude/skills/.../plan.md:454` (5 values), this SPEC (8 proposed) |
| L2 — Trigger | Hook only fires on local `git commit`, not on PR merge | `internal/hook/spec_status.go:46` (`isGitCommitCommand`) |
| L3 — Lint | `FrontmatterSchemaRule` checks presence only, not value validity | `internal/spec/lint.go:510` (no enum check) |
| L4 — CI | Zero GitHub Actions enforce SPEC status correctness | `.github/workflows/` (18 workflows, none touch SPEC frontmatter) |
| L5 — Format | 21 legacy SPECs use markdown H2/table instead of YAML frontmatter | `SPEC-MX-001` (markdown table), `SPEC-CC2122-HOOK-001` (`## Status:` H2), etc. |
| L6 — Visibility | No drift report; user discovers drift only via retrofit PR review | (none) |
| L7 — Ownership | Agent definitions don't specify status-transition responsibility | `.claude/agents/manager-spec.md`, `manager-docs.md`, `manager-develop.md` |

### 1.4 Existing Assets (Build-On vs Replace)

| Asset | Status | Action |
|-------|--------|--------|
| `internal/spec/status.go` (`ValidStatuses`, `UpdateStatus`, format detection) | Working, well-tested | **Extend** — add 2 enum values, preserve format-detection logic |
| `internal/hook/spec_status.go` (commit-message detection) | Working but narrow trigger | **Extend** — add PR-title prefix recognition + transition map |
| `internal/cli/spec_status.go` (`update`, `--list`, `--sync-git`) | Working CLI surface | **Extend** — add `moai spec drift` subcommand |
| `internal/spec/lint.go` (`FrontmatterSchemaRule` presence-only) | Working | **Extend** — add 3 new lint rules (Value/Case/GitConsistency) |
| `SPEC-STATUS-AUTO-001` (completed 2026-04-27) | Foundation for this SPEC | **Depend on** — this SPEC closes the gap STATUS-AUTO-001 deferred to L3 future work |
| 21 legacy SPECs with non-YAML frontmatter | Drift source | **Backfill** — one-time conversion via `moai spec status` CLI |

---

## §2. Assumptions

### A-1: PR-title convention is stable
- **Statement**: Conventional Commits format (`plan(SPEC-X-NNN):`, `feat(SPEC-X-NNN):`, `docs(sync):`) is consistently applied across the project.
- **Confidence**: High. Validated by 30+ recent merged PRs (PR #864 `feat(catalog):`, PR #869 `docs(sync):`, PR #870 `feat(spec):`, etc.).
- **Validation**: Lint Wave 1 includes commit-title regex check; CI fails on mismatch.

### A-2: GitHub Actions `pull_request.closed` event reliability is sufficient
- **Statement**: GitHub fires `pull_request.closed` with `merged == true` reliably; we do not need to poll.
- **Confidence**: High. Industry standard; used by Release Drafter, auto-merge.yml already in repo.
- **Validation**: Wave 3 W3-T1 test triggers on real PR merge in dry-run mode first.

### A-3: 21 legacy SPECs are functionally complete
- **Statement**: All 21 legacy SPECs in REQ-6 backfill list are `completed`, `archived`, or `superseded` — none are active. Backfill is format-only, not state-changing.
- **Confidence**: Medium-High. Spot-checked SPEC-MX-001 (Planned in markdown table, but implementation is on main), SPEC-I18N-001-ARCHIVED (explicit archive banner), SPEC-CC2122-HOOK-001 (`## Status: COMPLETED`). Caveat: a few may need lifecycle audit during W1-T3.
- **Validation**: W1-T3 dry-run lists current status before write; user reviews before commit.

### A-4: Hook PostToolUse can detect PR merges
- **Statement**: PR merges via `gh pr merge` are detected by PostToolUse (Bash matcher) just as `git commit` is.
- **Confidence**: Medium. `gh pr merge --squash` does invoke an underlying git operation, but the PostToolUse event captures the `gh` Bash invocation, not the resulting commit on origin/main.
- **Risk**: Hook may miss PR merges performed via the GitHub web UI or merge queue.
- **Mitigation**: REQ-5 GitHub Actions workflow is the **primary** trigger; PostToolUse hook is a **secondary** trigger for local-merge workflows. Both must converge on the same outcome.

### A-5: Drift detection cost is bounded
- **Statement**: Cross-referencing every SPEC against `git log main --no-merges` for SPEC-XXX patterns completes in <2 seconds for the current 180+ SPEC count.
- **Confidence**: High. `getSPECIDsFromGitLog()` in `internal/cli/spec_status.go:221` already performs this on `--sync-git` and is sub-second.
- **Validation**: W3-T2 benchmarks on actual repo; if >5s, fall back to incremental scan.

---

## §3. Requirements (EARS Format)

### REQ-1: Status Enum Single Source of Truth (Ubiquitous)

**The system SHALL** define a single canonical status enum containing exactly 8 values: `draft`, `planned`, `in-progress`, `implemented`, `completed`, `superseded`, `archived`, `rejected`. **The system SHALL** use hyphen `in-progress` (not underscore `in_progress`) per codebase precedent (`internal/spec/status.go:16`). **The system SHALL** document this enum in `.claude/rules/moai/workflow/spec-workflow.md` under a new `§ Status Lifecycle` section, treating that document as the canonical reference.

### REQ-2: PR-Merge Trigger Recognition (Event-Driven)

**WHEN** a PR is merged with title prefix matching one of the patterns below, **THE system SHALL** classify it and update affected SPECs accordingly:

| Prefix Pattern | Classification | Status Transition |
|----------------|----------------|-------------------|
| `plan(`, `plan(SPEC-X):`, `plan(spec):` | plan-merge | `draft` → `planned` |
| `feat(SPEC-X)`, `fix(SPEC-X)` (AC incomplete) | run-partial | `planned`/`draft` → `in-progress` |
| `feat(SPEC-X)`, `fix(SPEC-X)` (AC complete) | run-complete | `planned`/`draft`/`in-progress` → `implemented` |
| `docs(sync)`, `chore(sync)` | sync-merge | `implemented` → `completed` |

**The system SHALL** preserve the existing `git commit` trigger in `internal/hook/spec_status.go` for backward compatibility with local-development workflows.

### REQ-3: Lint Rule Trio (Ubiquitous)

**The system SHALL** add three lint rules to `internal/spec/lint.go`:

- **REQ-3.1 `StatusValueEnumRule`**: **IF** a SPEC's `status` field value is NOT in the 8-value canonical enum, **THEN the system SHALL** emit a `SeverityError` finding with code `StatusValueInvalid`.
- **REQ-3.2 `StatusCaseNormalizationRule`**: **IF** a SPEC's `status` field value contains uppercase letters (e.g., `Planned`, `COMPLETED`), **THEN the system SHALL** emit a `SeverityError` finding with code `StatusCaseInvalid` and recommend the lowercase equivalent.
- **REQ-3.3 `StatusGitConsistencyRule`**: **IF** a SPEC's current status is `draft` but `git log main --no-merges` contains a `plan(<SPEC-ID>)` commit, **THEN the system SHALL** emit a `SeverityWarning` finding with code `StatusGitDrift`. In `--strict` mode, the finding severity SHALL be elevated to `SeverityError`.

### REQ-4: SPEC Lint CI Gate (Event-Driven)

**WHEN** a pull request is opened or updated, **THE system SHALL** execute `moai spec lint --strict` via a new GitHub Actions workflow `.github/workflows/spec-lint.yml`. **THE system SHALL** block PR merge if lint exits non-zero.

### REQ-5: PR-Merge Auto-Sync CI Gate (Event-Driven)

**WHEN** a pull request closes with `merged == true`, **THE system SHALL** execute a new GitHub Actions workflow `.github/workflows/spec-status-auto-sync.yml` that:
1. Parses the PR title via REQ-2 classification.
2. Identifies all SPEC-XXX references in the title and body.
3. Runs `moai spec status <SPEC-ID> <new-status>` for each match.
4. Commits the changes back to `main` via a follow-up commit with title `chore(spec): auto-sync status for #<PR>` and author `moai-bot[bot]`.
5. Skips when the merged PR is itself a `chore(spec): auto-sync` follow-up (loop prevention).

### REQ-6: Legacy Frontmatter Backfill (Ubiquitous, one-shot)

**The system SHALL** backfill the following 21 SPECs to YAML frontmatter format using `moai spec status` CLI (which preserves `updateStatusInContent`'s format-detection logic):

```
SPEC-CC2122-HOOK-001, SPEC-CC2122-HOOK-002, SPEC-CICD-001, SPEC-CORE-BEHAV-001,
SPEC-DESIGN-001, SPEC-EVO-001, SPEC-GLM-001, SPEC-HOOK-008,
SPEC-I18N-001-ARCHIVED, SPEC-LSP-FLAKY-001, SPEC-LSP-FLAKY-002, SPEC-MX-001,
SPEC-PSR-001, SPEC-REFLECT-001, SPEC-SKILL-ENHANCE-001, SPEC-SLE-001,
SPEC-SLQG-001, SPEC-SLV3-001, SPEC-STATUSLINE-001, SPEC-TEAM-001,
SPEC-TELEMETRY-001
```

**The system SHALL** preserve the original status value (lowercased) and SHALL NOT change SPEC semantics. **The system SHALL** detect existing format (`## Status:`, `| Status |`, `| 상태 |`, archive banner) and emit YAML frontmatter equivalent.

### REQ-7: Drift Detection CLI (Ubiquitous)

**WHEN** `moai spec drift` is invoked, **THE system SHALL**:
1. Scan all SPECs in `.moai/specs/`.
2. Cross-reference with `git log main --no-merges` for SPEC-XXX commit patterns.
3. Classify each SPEC as: `aligned`, `drift-plan-not-recorded`, `drift-implementation-not-recorded`, `drift-completion-not-recorded`, `unknown`.
4. Print a tabular report with `--json` and `--exit-code-on-drift` flags.

### REQ-8: SessionStart Drift Warning (Event-Driven)

**WHEN** a session starts AND `moai spec drift --count` returns >= 5, **THE system SHALL** emit a one-line warning to stderr: `[drift] N SPECs out of sync — run 'moai spec drift' for details`. **The system SHALL NOT** emit a full report at session start (token-budget conscious; report must be user-initiated).

### REQ-9: Agent Responsibility Matrix (Ubiquitous)

**The system SHALL** update three agent definitions with explicit status-transition responsibility entries:

| Agent | Owned Transitions |
|-------|-------------------|
| `manager-spec` | `draft` (initial creation), `draft` → `planned` (after plan PR merge — handover) |
| `manager-develop` | `planned` → `in-progress` → `implemented` (during run phase) |
| `manager-docs` | `implemented` → `completed` (after sync PR merge) |

### REQ-10: Out-of-Scope Documentation (Ubiquitous)

**The system SHALL** include in this SPEC's §6 Out of Scope an explicit statement that 2026-04-23-era `Planned` (capitalized) values are addressed via REQ-6 one-shot backfill, NOT via a runtime migration shim. **The system SHALL NOT** add runtime case-folding to `IsValidStatus()`.

---

## §4. Specifications

### 4.1 Status Enum Canonical Definition

Code site: `internal/spec/status.go:13`

```go
// ValidStatuses defines all allowed status values per SPEC-V3R4-STATUS-LIFECYCLE-001.
var ValidStatuses = []string{
    "draft",        // initial, not yet planned
    "planned",      // plan PR merged, ready for run phase
    "in-progress",  // run started, AC partially met
    "implemented",  // run complete, all AC GREEN, sync not yet done
    "completed",    // sync PR merged, lifecycle terminal-success
    "superseded",   // replaced by another SPEC (with superseded_by field)
    "archived",     // historical record, not active
    "rejected",     // never implemented, lifecycle terminal-decline
}
```

### 4.2 Transition Map (REQ-2)

PR-title prefix → status mapping is encoded in a single transition table consumed by both `internal/hook/spec_status.go` (PostToolUse local trigger) and `.github/workflows/spec-status-auto-sync.yml` (CI authoritative trigger). The table lives in `internal/spec/transitions.go` (new file) to keep behaviour parity between the two triggers.

### 4.3 Lint Rule Registration

`internal/spec/lint.go:118` already wires `&FrontmatterSchemaRule{}`. REQ-3 adds three more rules to the registry. Existing `FrontmatterSchemaRule` is unchanged.

### 4.4 CI Workflow Skeleton

Two new workflows:
- `.github/workflows/spec-lint.yml` — runs on `pull_request` events, executes `moai spec lint --strict`.
- `.github/workflows/spec-status-auto-sync.yml` — runs on `pull_request.closed` with `merged == true`, calls REQ-5 logic.

Both workflows use the existing `ci.yml` Go setup as the dependency baseline.

---

## §5. In Scope

- New canonical 8-value status enum.
- Three new lint rules + CI gate.
- Hook extension for PR-title prefix recognition (additive, backward compatible).
- Two new GitHub Actions workflows (lint + auto-sync).
- One-shot backfill of 21 legacy SPECs to YAML frontmatter.
- New `moai spec drift` CLI subcommand.
- SessionStart drift warning (one-line, token-conscious).
- Agent definition updates: manager-spec, manager-docs, manager-develop.
- 30-day post-merge monitoring period (§7 Verification).

---

## §6. Out of Scope

Per Section 17 of `.claude/skills/moai-workflow-spec/SKILL.md` (SPEC Scope Classification), this section is **mandatory** and contains at least one explicit exclusion.

- **Runtime `Planned` → `planned` shim**: `IsValidStatus()` will continue to reject `Planned`. REQ-6 backfill converts the 21 legacy SPECs once; no ongoing runtime case-folding. **Rationale**: A shim would mask the lint signal that drift has reappeared.
- **Cross-repo SPEC standardization**: This SPEC is internal to `moai-adk-go`. Downstream projects using moai-adk as a template are out of scope; they may adopt the standard independently.
- **Retroactive PR-history scanning beyond `git log main --no-merges`**: We don't crawl closed PRs from before the SPEC's merge date. Drift older than the catalyst window (2026-04-23) is addressed only via REQ-6 backfill.
- **Web UI for status management**: CLI + CI only.
- **SPEC lifecycle state machine enforcement** (e.g., reject `completed` → `draft`): Lint warns on inconsistency, but doesn't reject arbitrary transitions. Future SPEC may add this once the standard is stable.
- **Automatic archive of `completed` SPECs after N days**: Lifecycle-event scope only, not time-based GC.
- **Loop prevention beyond title prefix**: REQ-5 skips its own `chore(spec): auto-sync` follow-ups by title. Other meta-loops are not handled.

---

## §7. Verification

### Phase-Gate Acceptance

Each Wave has independent acceptance (see `acceptance.md`). Wave 1 must pass before Wave 2 starts. Wave 2 must pass before Wave 3 starts.

### 30-Day Monitoring Period

**WHEN** Wave 3 merges AND 30 days elapse, **THE system SHALL** be evaluated against:
- **M-1 (primary)**: Zero retrofit PRs of pattern `chore(spec): status drift sync` merged in the 30-day window. Catalyst PR #871's pattern must not recur.
- **M-2 (secondary)**: `moai spec drift --count` returns <= 2 at any point in the window.
- **M-3 (tertiary)**: Zero SPEC frontmatter format violations in `moai spec lint --strict` output.

If M-1 fails, this SPEC's status MUST be re-evaluated and a follow-up SPEC opened to address the gap.

---

## §8. Constraints

- **C-1 Enum naming precedent**: Hyphen `in-progress`, NOT underscore. Maintain consistency with `internal/spec/status.go:16`.
- **C-2 Legacy preservation**: 21 SPECs in REQ-6 are `completed`/`archived` — backfill MUST NOT change behaviour, only frontmatter format.
- **C-3 CI grace window**: New lint rules ship with `--strict` opt-in initially. Strict-mode becomes the default only after Wave 3 lands and the 30-day window observes zero unexpected blockers.
- **C-4 Wave 1 reversibility**: All Wave 1 changes (lint rules, backfill) are markdown + Go-internal additions. No breaking changes to existing CLI or workflow tooling. Reversible via `git revert` without state surgery.
- **C-5 Hook-CI dual-trigger consistency**: REQ-2 and REQ-5 use the same transition table (`internal/spec/transitions.go`); divergence between local and CI behaviour is a regression.

---

## §9. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| R-1: PR-title classification missclassifies edge cases (e.g., revert PRs) | Medium | Medium | Add `revert(` prefix to "no-op" category in transition map; CI logs all classifications for audit. |
| R-2: Auto-sync follow-up commit creates merge conflict with concurrent PR | Low | Medium | Auto-sync commit retries on conflict with rebase; on second failure, opens an issue tagged `spec-drift`. |
| R-3: REQ-6 backfill changes a SPEC's effective semantics (e.g., active → archived misread) | Low | High | W1-T3 produces a dry-run diff for human review before write. User explicitly approves each line. |
| R-4: Hook PostToolUse misses PR merges via web UI | Medium | Low | REQ-5 CI workflow is authoritative; hook is best-effort secondary. |
| R-5: 30-day monitoring not enforceable mechanically | High | Low | Document the M-1/M-2/M-3 criteria in §7; future audit SPEC may automate. |

---

## §10. Dependencies

- **Hard**: `SPEC-STATUS-AUTO-001` (completed 2026-04-27) — foundation for `internal/spec/status.go` and `internal/hook/spec_status.go`. This SPEC extends both.
- **Soft**: `SPEC-V3R4-CATALOG-001/002` (recently merged) — establishes V3R4 series precedent.
- **Soft**: `SPEC-V3R2-WF-003`/`WF-004` (mode dispatch, classification) — provides the conventional-commits PR-title contract this SPEC parses.

---

## §11. References

- Catalyst: PR #871 (OPEN, 2026-05-13) — "chore(spec): status drift sync — Sprint 12 plan-merged 10건"
- Foundation SPEC: `.moai/specs/SPEC-STATUS-AUTO-001/spec.md`
- Phase discipline: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline
- Lint framework: `internal/spec/lint.go`
- Status update logic: `internal/spec/status.go`
- Hook handler: `internal/hook/spec_status.go`
- CLI surface: `internal/cli/spec_status.go`
- Existing CI: `.github/workflows/` (18 workflows)
