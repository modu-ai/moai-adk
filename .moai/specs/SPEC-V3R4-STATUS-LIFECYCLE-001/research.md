# research.md — SPEC-V3R4-STATUS-LIFECYCLE-001

Deep codebase analysis grounding this SPEC. All file references include line numbers and direct quotes. Generated 2026-05-13 from `main` HEAD `07dabe011`.

---

## §1. The Drift Pattern (Quantified)

### 1.1 Retrofit PR Cadence

Querying `gh pr list --search "chore(spec): status drift"` shows a clear 2-7 day cadence over the last 30 days:

| Retrofit PR | Merged | Scope (SPECs touched) | Phase signal |
|-------------|--------|-----------------------|--------------|
| #818 | 2026-05-10 | WF-004 retrofit | Sprint 8 closeout |
| #844 | 2026-05-11 | ORC-001, ORC-005, RT-005, BRAIN-001 (4 SPECs) | Sprint 9-10 closeout |
| #856 | 2026-05-12 | RT-007 + 4 others (5 SPECs) | Sprint 11 closeout |
| #866 | 2026-05-12 | 20 SPECs implemented/in_review → completed | metadata sweep |
| **#871** | **OPEN as of 2026-05-13** | 10 Sprint 12 SPECs draft → planned (1 → in-progress) | **Catalyst** |

**5 retrofits in 4 days** is the evidence-base for system-level intervention.

### 1.2 PR #871 Body — verbatim catalyst note

> "본 PR은 PR #866 (20건 일괄 정정) 같은 retrofit 작업입니다. 만성 재발 패턴 [...] 원인: plan PR 머지 시 status 자동 전이가 부재. hook은 `git commit` 시점만 감지하고 PR 머지 (`gh pr merge --squash`)는 누락. `manager-spec` 책임 종료 시점 vs `manager-docs` 시작 시점 사이가 무주공산."

The PR #871 author has already diagnosed the root cause and named the proposed solution: `SPEC-V3R4-STATUS-LIFECYCLE-001`. This SPEC formalizes that proposal.

---

## §2. Existing Asset Map

### 2.1 `internal/spec/status.go` (`ValidStatuses` enum)

Line 13-21:
```go
var ValidStatuses = []string{
    "draft",
    "planned",
    "in-progress",
    "implemented",
    "completed",
    "superseded",
}
```

**Observation**: 6 values. Missing `archived` and `rejected`. The catalyst SPEC body proposes 8 values; this SPEC accepts that and extends the enum.

**Hyphen vs underscore decision**: The codebase uses hyphen (`in-progress`). PR #871 body explicitly notes "하이픈 사용 — 코드 정합". REQ-1 codifies this.

### 2.2 `internal/spec/status.go` (`updateStatusInContent` format detection)

Line 73-101:
```go
func updateStatusInContent(content, newStatus string) (string, error) {
    lines := strings.Split(content, "\n")
    sampleLines := lines
    if len(sampleLines) > 30 {
        sampleLines = sampleLines[:30]
    }
    sample := strings.Join(sampleLines, "\n")

    if strings.Contains(sample, "| 상태 |") || strings.Contains(sample, "| Status |") {
        return updateStatusInTable(lines, newStatus)
    }
    if strings.Contains(sample, "- **Status**:") || strings.Contains(sample, "- **상태**:") {
        return updateStatusInMarkdownList(lines, newStatus)
    }
    if strings.Contains(sample, "---") {
        return updateStatusInYAML(lines, newStatus)
    }
    return addYAMLFrontmatter(content, newStatus)
}
```

**Observation**: Already supports 4 input formats (YAML, table-Korean, table-English, markdown-list) and a fallback (add YAML if none present). REQ-6 backfill REUSES this — we do not write a new format converter.

### 2.3 `internal/hook/spec_status.go` (commit-only trigger)

Line 46-89:
```go
if !isGitCommitCommand(data.Command) {
    return &HookOutput{}, nil
}
// [...]
func isGitCommitCommand(command string) bool {
    return strings.Contains(command, "git") && strings.Contains(command, "commit")
}
```

**Observation**: Trigger is exclusively `git commit`. `gh pr merge --squash` is captured by PostToolUse (Bash matcher) but rejected here because the command string contains `gh pr merge`, not `git commit`. **This is the smoking gun for L2.**

**Fix path**: REQ-2 adds `isGhPrMergeCommand()` peer function and a transition map keyed by PR-title prefix.

### 2.4 `internal/spec/lint.go` (`FrontmatterSchemaRule`)

Line 510-566:
```go
required := []struct {
    name  string
    value string
}{
    {"id", fm.ID},
    {"title", fm.Title},
    {"version", fm.Version},
    {"status", fm.Status},
    // [...]
}

for _, field := range required {
    if strings.TrimSpace(field.value) == "" {
        findings = append(findings, Finding{
            Severity: SeverityError,
            Code:     "FrontmatterInvalid",
            Message:  fmt.Sprintf("Frontmatter required field missing: %s", field.name),
        })
    }
}

if fm.ID != "" && !specIDPattern.MatchString(fm.ID) {
    findings = append(findings, Finding{ ... })
}
```

**Observation**: Rule checks **presence** (non-empty) but NOT **value validity**. `status: Planned` passes this rule because the field is non-empty. **This is the L3 root cause.**

**Fix path**: REQ-3 adds three sibling rules: `StatusValueEnumRule`, `StatusCaseNormalizationRule`, `StatusGitConsistencyRule`. Each implements the `lint.Rule` interface (verified at line 503-504 for the pattern).

### 2.5 `internal/cli/spec_status.go` (CLI surface)

Line 19-66 — `newSpecStatusCmd` registers `moai spec status` with three flags: `--dry-run`, `--list`, `--sync-git`.

Line 221-244 — `getSPECIDsFromGitLog()` already does the cross-reference work REQ-7 needs:
```go
out, err := exec.Command("git", "log", branch, "--oneline", "--no-merges").Output()
// [...]
matches := specIDPattern.FindAllString(string(out), -1)
```

**Observation**: We can build `moai spec drift` on this foundation in <100 LOC. The git-log scan is already proven sub-second on the 180+ SPEC repo.

### 2.6 `.claude/skills/.../plan.md:454`

```
- [ ] `status: draft` — enum: draft | approved | completed | superseded | archived
```

**Observation**: 5-value enum stated here. Does NOT match `internal/spec/status.go` (6 values) or REQ-1 proposal (8 values). **L1 documentation drift.**

**Note**: `approved` appears here but not in code. Likely vestigial from an earlier design; not used in actual frontmatter today.

### 2.7 `.claude/skills/.../sync.md`

Mixed enum usage across the file:
- Line 720: `"completed" if all requirements met, or "in-progress" if partial"`
- Line 731: `"If SPEC status set to "in-progress":`
- Line 815: `"SPEC Status: [completed|in-progress]"`
- Line 1154: `"SPEC status updated to "implemented""`
- Line 1162: `"SPEC status updated to "in-progress" (not "implemented")"`

**Observation**: At least 4 distinct values used (`completed`, `in-progress`, `implemented`, `planned`) but no central enum declaration. **L1 documentation drift compounds the L3 lint gap.**

### 2.8 `.github/workflows/` (18 workflows)

Inventory by status-relevance:

| Workflow | Touches SPEC? | Why irrelevant to status |
|----------|---------------|--------------------------|
| ci.yml | No | Go test/lint only |
| codeql.yml | No | Security scan |
| auto-merge.yml | No | Dependabot auto-merge |
| docs-i18n-check.yml | No | docs-site i18n |
| release-drafter.yml | No | CHANGELOG draft |
| release.yml | No | GoReleaser |
| test-install.yml | No | Install script test |
| Other 11 | No | Reviews, labels, community |

**Verdict**: Zero workflows touch SPEC frontmatter. **L4 root cause confirmed.** REQ-4 and REQ-5 fill this gap with two minimal workflows.

---

## §3. Legacy Format Inventory (REQ-6 evidence)

Sampled 5 of the 21 SPECs to confirm format diversity:

### 3.1 `SPEC-CC2122-HOOK-001/spec.md` — markdown H2

```markdown
# SPEC-CC2122-HOOK-001: Claude Code v2.1.119-121 Hook Feature Integration

## Status: COMPLETED

## Overview
[...]
```

Format: `## Status: <uppercase>`. No frontmatter. Detected by `updateStatusInContent` as "no `---` found, no `| Status |`, no `- **Status**:`" → falls through to `addYAMLFrontmatter`. **REQ-6 backfill compatible.** Backfill writes `status: completed` to a new YAML block.

### 3.2 `SPEC-CICD-001/spec.md` — Korean markdown table

```markdown
| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CICD-001 |
| 상태 | Completed |
```

Format: Korean table (`| 상태 |`). Detected by `updateStatusInTable`. **REQ-6 backfill works.** Backfill rewrites table cell to lowercase `completed` AND adds YAML frontmatter (current behaviour: backfill writes both? Verify in W1-T3 dry-run).

### 3.3 `SPEC-MX-001/spec.md` — English markdown table

```markdown
| Field       | Value         |
|-------------|---------------|
| SPEC ID     | SPEC-MX-001   |
| Status      | Planned       |
```

Format: English table. Detected by `updateStatusInTable`. **Watch-out**: status value is `Planned` (capitalized). After backfill, table cell becomes `planned`. **REQ-6 W1-T3 dry-run must show this transformation explicitly so the user can sanity-check.**

### 3.4 `SPEC-I18N-001-ARCHIVED/spec.md` — archive banner + table

Has both an archive-prefix banner AND an internal table with `Status: Planned`. Backfill must produce `status: archived` (not `status: planned`) per the archive banner intent. **REQ-6 W1-T3 dry-run flags this case for explicit user confirmation.**

### 3.5 `SPEC-STATUSLINE-001/spec.md` — English markdown table

```markdown
| Field    | Value                                                              |
| -------- | ------------------------------------------------------------------ |
| SPEC ID  | SPEC-STATUSLINE-001                                                |
| Title    | Statusline Segment Configuration via Wizard and YAML Config        |
| Status   | Completed                                                          |
```

Format: English table, value `Completed`. Simple case. Backfill produces `status: completed`.

### 3.6 Risk Summary

Of the 21 SPECs, an estimated:
- ~14 are simple (single status value, completed → completed)
- ~5 have edge cases (archive banners, dual status references)
- ~2 need lifecycle audit (SPEC-MX-001 listed Status=Planned but implementation is on main since 2026-02)

Risk R-3 in spec.md §9 covers this with mandatory dry-run review.

---

## §4. The Trigger Gap (Why Hooks Alone Are Insufficient)

### 4.1 PostToolUse Coverage

`internal/hook/spec_status.go` registers as a `PostToolUse` handler. The Claude Code matcher is `Bash`. When the user runs `gh pr merge 871 --squash --admin`, this fires correctly with `tool_name: "Bash"` and `command: "gh pr merge 871 --squash --admin"`.

`isGitCommitCommand` returns false (no "commit" in the command string). Handler returns early. No status update.

### 4.2 What we cannot fix in the hook alone

Even if we extend `isGhPrMergeCommand`, the hook only fires when the user runs `gh pr merge` locally. PR merges via:
- GitHub web UI ("Squash and merge" button)
- GitHub merge queue
- Branch protection auto-merge (auto-merge.yml workflow)
- Dependabot auto-merge

...do NOT fire the local PostToolUse hook. **This is why REQ-5 GitHub Actions is authoritative and the hook is best-effort secondary.**

### 4.3 Convergence guarantee

Both triggers (hook + Actions) consume the same `internal/spec/transitions.go` map (REQ-2, §4.2 in spec.md). When both run on the same PR, the second invocation is a no-op (idempotent: status already at target value, `updateStatusInContent` is content-stable for unchanged values).

---

## §5. Wave Decomposition Validation

Cross-checking the proposed 3 waves against asset dependencies:

### Wave 1 (Policy + Lint)
- W1-T1 (enum in spec-workflow.md): zero code dependencies. **Independent.**
- W1-T2 (StatusValueEnumRule + StatusCaseNormalizationRule): depends on W1-T1 enum being canonical. **Linear.**
- W1-T3 (21 SPEC backfill): depends on W1-T2 NOT being strict-mandatory yet, otherwise old format triggers lint fail. **Linear after W1-T2.**
- W1-T4 (lint CI workflow): depends on W1-T2 + W1-T3 to avoid CI red on day 1. **Linear after W1-T3.**

**Wave 1 is the lowest-risk first wave.** All changes are additive markdown/YAML/Go-internal. Reversible via `git revert`.

### Wave 2 (Hook + Transitions)
- W2-T1 (extend hook): depends on `internal/spec/transitions.go` (new). **Builds on Wave 1.**
- W2-T2 (unit tests): depends on W2-T1.
- W2-T3 (update plan.md / sync.md docs): zero code dependencies. **Independent within Wave 2.**

Wave 2 is moderate risk. Hook behaviour change affects local-dev only; CI is not yet flipped to strict.

### Wave 3 (Automation + Visibility)
- W3-T1 (Actions workflow): depends on Wave 2 transition table. **Linear after Wave 2.**
- W3-T2 (`moai spec drift` CLI): depends on `getSPECIDsFromGitLog()` (already exists). **Independent.**
- W3-T3 (SessionStart integration): depends on W3-T2 CLI being available.
- W3-T4 (StatusGitConsistencyRule): depends on W3-T2 logic (shared git scan).
- W3-T5 (agent definition updates): zero code dependencies. **Independent.**
- W3-T6 (30-day monitoring documentation): zero code dependencies.

Wave 3 is highest risk (CI auto-commits to main) but isolated to a single workflow file. Loop prevention via title check is straightforward.

---

## §6. Open Questions for User (Pre-Run)

None at plan time. All assumptions in spec.md §2 are validated by file evidence above. The plan-auditor invocation (if harness demands it) will surface any gaps before Wave 1 starts.

One **forward-looking** question for the user, NOT blocking plan PR merge:

- **OQ-1**: Should W3-T1 (Actions workflow) commit auto-sync changes directly to `main`, or open a PR? Direct commit is simpler but circumvents branch protection. PR-based is safer but adds churn. **Recommendation**: Direct commit with `[skip ci]` marker; rationale = auto-sync changes are metadata-only, low-blast-radius. User decides at start of Wave 3.

---

## §7. References (file evidence)

- `internal/spec/status.go:13-21` — ValidStatuses enum (6 values)
- `internal/spec/status.go:73-101` — updateStatusInContent format detection
- `internal/hook/spec_status.go:46-89` — commit-only trigger
- `internal/spec/lint.go:118` — FrontmatterSchemaRule registration
- `internal/spec/lint.go:510-566` — FrontmatterSchemaRule.Check (presence-only)
- `internal/cli/spec_status.go:19-66` — newSpecStatusCmd
- `internal/cli/spec_status.go:221-244` — getSPECIDsFromGitLog
- `.claude/skills/moai/workflows/plan.md:454` — 5-value enum statement
- `.claude/skills/moai/workflows/sync.md:720,731,815,1154,1162` — mixed enum usage
- `.github/workflows/` — 18 workflows, zero SPEC-touching
- PR #871 body — catalyst diagnosis (verbatim quote in §1.2)
- `.moai/specs/SPEC-STATUS-AUTO-001/spec.md` — foundation SPEC
