---
id: SPEC-V3R6-CLI-AUDIT-001
title: "Implementation Plan — moai CLI inventory + dead command + integration analysis (research-only baseline)"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cli, audit, research, sprint-2, sprint-7-baseline"
related_specs: [SPEC-V3R6-LEGACY-CLEANUP-001, SPEC-V3R6-LEGACY-CLEANUP-002]
---

# Implementation Plan — SPEC-V3R6-CLI-AUDIT-001

## Section A — Context

### A.1 Prior Sprint 2 SPECs (completed lifecycle)

| SPEC | Tier | Status | Deliverable |
|------|------|--------|-------------|
| SPEC-V3R6-CHANGELOG-CLEANUP-001 | S | implemented | CHANGELOG line 65 hallucination removal + B12 standing-rule guard |
| SPEC-V3R6-SESSION-HANDOFF-AUTO-001 | S | implemented | `internal/hook/handoff/persist.go` SessionEnd auto-handoff |
| SPEC-V3R6-LEGACY-CLEANUP-001 | L | implemented | 31-file agency keyword cleanup + per-canonical-locale ≤5 |
| SPEC-V3R6-LEGACY-CLEANUP-002 | S | implemented | 7-file template mirror cascade (Template-First Rule restoration) |

This SPEC continues Sprint 2 P2 sequencing per v3.0 roadmap Scenario B v2 Round 2 re-prioritization (memory `project_v3r6_sprint2_amr_run_ready`).

### A.2 Research scope (4 areas — recapped from spec.md §1)

1. **Area 1**: Subcommand inventory (40-55 estimated subcommands across root + sub-subcommand + flags)
2. **Area 2**: Dead-command identification (cross-reference 6 source classes: hooks-local, hooks-template, skills/workflows, cmd/main.go, test files, docs-site)
3. **Area 3**: `moai init` / `moai update -c` / profile system integration map (10 `moai update` flags + profile triad)
4. **Area 4**: Sprint 7 FINAL CLI-INTEGRATION-001 baseline scope (directly consumable scope section)

### A.3 Deferred consumer (Sprint 7 FINAL)

**SPEC-V3R6-CLI-INTEGRATION-001** (Sprint 7 FINAL STAGE, not yet authored). Will consume `## §4 Sprint 7 Baseline` directly as input. This SPEC's audit MUST be written to enable that consumption (not aspirational; concrete + actionable).

### A.4 Execution mode

**Hybrid Trunk Tier M**. Default: feat-branch + auto PR with 4 CI status checks. Code diff is .md-only (research deliverable + SPEC artifacts), so CI gates pass trivially. **Per-SPEC main-direct override** available at run-phase via AskUserQuestion (L40 pattern), analogous to LEGACY-CLEANUP-001's override decision.

### A.5 No `/moai mx` step

Research-only SPECs produce no Go code modification → no `@MX:NOTE`/`@MX:ANCHOR` annotation. SPEC lifecycle ends at run-phase completion (+ optional `/moai sync` for CHANGELOG entry; manager-docs may elect minimal append). No M5 milestone.

---

## Section B — Milestones

This SPEC has **4 milestones (M1-M4)**. Each milestone produces a discrete section of the audit report. Files-touched-per-milestone: 1 (the growing audit report) + progress.md updates. Sequential execution (M1 → M2 → M3 → M4) — each section depends on prior section's inventory output.

### M1 — Subcommand Inventory (Area 1)

**Scope**: Walk the cobra command tree starting from `cmd/moai/main.go` root command and enumerate every subcommand (root, sub, sub-sub) + per-command flags. Cross-reference with `grep "AddCommand"` + `grep "cmd.AddCommand"` to validate completeness.

**Files touched**:
1. `.moai/reports/cli-audit/audit-{ISO-DATE}.md` (NEW — initial creation with `## §1 Subcommand Inventory` section)
2. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` (NEW — M1 row added)

**Tasks**:
- T1.1: Initialize audit report file `.moai/reports/cli-audit/audit-$(date -u +%Y-%m-%d).md` with template header + 5 section placeholders (REQ-CLA-001, REQ-CLA-002, REQ-CLA-003, REQ-CLA-004 + §5 Methodology)
- T1.2: Read `cmd/moai/main.go` to identify root command and direct children registrations
- T1.3: For each `internal/cli/*.go` file containing `AddCommand`, extract: command name, parent command, file:line, flag definitions
- T1.4: Build hierarchical inventory table in `## §1 Subcommand Inventory`:
  - Column 1: Full command path (e.g., `moai harness status`)
  - Column 2: Parent (e.g., `harness`)
  - Column 3: File:line registration (e.g., `internal/cli/harness.go:42`)
  - Column 4: Flags (CSV or YAML list)
  - Column 5: Brief Use description (1-line from `Short` cobra field)
- T1.5: Total count verification: bottom of `## §1` reports `Total subcommands: N` and `Total flag definitions: M`
- T1.6: Commit `plan(SPEC-V3R6-CLI-AUDIT-001): M1 — subcommand inventory`

**Verification (binary)**:
- `ls .moai/reports/cli-audit/audit-*.md | wc -l` returns ≥1
- `grep -c "^| \`moai " .moai/reports/cli-audit/audit-*.md` returns ≥40 (one table row per subcommand)
- `grep "Total subcommands:" .moai/reports/cli-audit/audit-*.md` returns 1 line with positive count

**Satisfies AC**: AC-CLA-001 (subcommand inventory ≥40 entries with structured fields)

**Risk**: LOW. Pure read + table writing. No code touched.

---

### M2 — Dead-Command Classification (Area 2)

**Scope**: For each subcommand inventoried in M1, perform 6-class cross-reference grep to classify as **active** (user-facing, externally referenced), **internal-only** (referenced only by hooks/templates), or **dead-suspect** (zero non-self references). Dead-suspect list requires ≥2 negative evidence sources before classification.

**Files touched**:
1. `.moai/reports/cli-audit/audit-{ISO-DATE}.md` (APPEND `## §2 Dead-Command Classification`)
2. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` (M2 row appended)

**Tasks**:
- T2.1: For each subcommand from M1 inventory, run 6 grep classes (parallel batch):
  - Class A (hooks-local): `grep -rn "moai <cmd>" .claude/hooks/moai/`
  - Class B (hooks-template): `grep -rn "moai <cmd>" internal/template/templates/.claude/hooks/`
  - Class C (skills/workflows): `grep -rn "moai <cmd>" .claude/skills/moai/workflows/ .claude/skills/moai-*/`
  - Class D (cmd/main.go): `grep -n "<cmd>" cmd/moai/main.go`
  - Class E (test invocations): `grep -rn "\"<cmd>\"" internal/cli/*_test.go`
  - Class F (docs-site): `grep -rn "moai <cmd>" docs-site/content/`
- T2.2: Classification rule:
  - **active**: ≥1 reference in F (docs-site) OR ≥2 references across A/B/C/D
  - **internal-only**: References only in A/B (hooks-local + hooks-template) with zero in C/D/F
  - **dead-suspect**: Zero references in A/B/C/D/F (self-references in command file excluded)
  - **Sub-subcommand handling** (e.g., `moai hook X`): classify the *root* + *sub* full path (`moai hook X`), NOT the bare sub-subcommand name. References to the bare sub-subcommand name (e.g., `db-schema-sync` without `moai hook` prefix) MAY indicate internal hook routing, not user-facing dead command. Cross-check with corresponding SPEC-{ID} status when sub-subcommand has an associated SPEC (e.g., `db-schema-sync` → SPEC-DB-SYNC-001 status check before dead-suspect classification).
- T2.3: Build classification table at `## §2`:
  - Column 1: Full command path
  - Column 2: Classification (active/internal-only/dead-suspect)
  - Column 3: Evidence count per class (A/B/C/D/E/F → e.g., `0/0/0/0/0/0`)
  - Column 4: Notes (rationale for borderline cases)
- T2.4: Dead-suspect summary subsection `### §2.1 Dead-Suspect Candidates`:
  - Enumerate full dead-suspect list with grep commands for reproducibility
  - Required: confirm preliminary `harness-observe*` family classification (per spec.md §1 dead-command candidates)
  - Document `moai db-schema-sync` actual presence/absence in code
- T2.5: Commit `plan(SPEC-V3R6-CLI-AUDIT-001): M2 — dead-command classification`

**Verification (binary)**:
- `grep -c "| dead-suspect |" .moai/reports/cli-audit/audit-*.md` returns ≥1
- `grep "### §2.1 Dead-Suspect Candidates" .moai/reports/cli-audit/audit-*.md` returns 1 match
- Every M1 subcommand appears as a classification row (count parity: M1 subcommands == M2 classified entries)

**Satisfies AC**: AC-CLA-002 (every inventoried subcommand classified with grep evidence + harness-observe family confirmed/refuted)

**Risk**: MEDIUM. False-positive dead classification = high cost (Sprint 7 mistakenly retires a used command). Mitigation: require ≥2 negative evidence sources + user review at run-phase completion.

---

### M3 — Integration Map (Area 3)

**Scope**: Produce dataflow + flag matrix for `moai init` + `moai update` (10 flags) + `moai cc -p` (profile system). Map config files read/written, template variables, environment exports, cross-command interactions.

**Files touched**:
1. `.moai/reports/cli-audit/audit-{ISO-DATE}.md` (APPEND `## §3 Integration Map`)
2. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` (M3 row appended)

**Tasks**:
- T3.1: `## §3.1 moai init flow` — Read `internal/cli/init*.go` to map:
  - Input: command args (project name, flags)
  - Template deployment path (`internal/template/embedded.go` → target dir)
  - Profile selection mechanism (if any during init)
  - Files written (`.moai/config/sections/*.yaml`, `CLAUDE.md`, `.claude/`, `.moai/`, etc.)
  - Environment variables read/exported
- T3.2: `## §3.2 moai update 10-flag matrix` — Read `internal/cli/update*.go` to build NxN flag interaction table:
  - 10 rows × 10 cols (flags: check/shell-env/config -c/force/yes/templates-only/binary/dry-run/no-hooks/verbose)
  - Cell = compatible (✓) / incompatible (✗) / no-op (-) / unique behavior (text)
  - Highlight `-c` (config-only) interaction with other flags (key Sprint 7 input)
- T3.3: `## §3.3 moai cc -p profile system` — Read `internal/cli/cc.go` + `internal/cli/cg.go` + `internal/cli/glm.go`:
  - Profile storage location (e.g., `~/.moai/profiles/<name>/` or `.claude/settings.local.json` derivation)
  - Profile naming convention + default
  - Profile switching mechanism + persistence
  - Interaction with `moai cg` (tmux + Claude leader + GLM teammates) + `moai glm` (GLM standalone)
- T3.4: `## §3.4 Cross-cutting concerns` — Document how:
  - `moai init` template selection interacts with `moai cc -p` profile choice
  - `moai update -c` (config-only) interacts with profile-managed config sections
  - Sprint 7 unification opportunities surface
- T3.5: Mermaid diagram (TD or TB direction recommended, not enforced — local research report scope, distinct from docs-site Mermaid TD-only rule which applies only to `docs-site/` per CLAUDE.local.md §17 — see `.moai/docs/docs-site-i18n-rules.md`) showing the 3-way relationship between init/update/profile commands
- T3.6: Commit `plan(SPEC-V3R6-CLI-AUDIT-001): M3 — integration map`

**Verification (binary)**:
- `grep "## §3.1 moai init flow" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §3.2 moai update 10-flag matrix" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §3.3 moai cc -p profile system" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §3.4 Cross-cutting" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep '```mermaid' .moai/reports/cli-audit/audit-*.md` returns ≥1

**Satisfies AC**: AC-CLA-003 (integration map 4 sub-sections + 10×10 flag matrix + mermaid diagram)

**Risk**: MEDIUM. Integration map complexity grows with command count. Mitigation: 4 sub-sections (§3.1-§3.4) decompose work; mermaid keeps cross-reference visual.

---

### M4 — Sprint 7 Baseline + Methodology Appendix (Area 4 + closure)

**Scope**: Synthesize §1+§2+§3 findings into a Sprint 7 directly-consumable scope section (§4). Append methodology appendix (§5) documenting grep commands used + files scanned + exclusions applied. Final SPEC frontmatter status transition draft → implemented.

**Files touched**:
1. `.moai/reports/cli-audit/audit-{ISO-DATE}.md` (APPEND `## §4 Sprint 7 Baseline Scope` + `## §5 Methodology Appendix`)
2. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md` (frontmatter `status: draft → implemented`, `version: 0.1.0 → 0.2.0`, `updated: <ISO-DATE>`)
3. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/plan.md` (same frontmatter sync)
4. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/acceptance.md` (same frontmatter sync)
5. `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` (M4 row + final summary + AC verification table)

**Tasks**:
- T4.1: `## §4.1 Recommended unifications` — Subcommands to consolidate (e.g., if `moai init` + `moai update -c` share template logic, recommend unified `moai project init|update`)
- T4.2: `## §4.2 Recommended retirements` — Dead-suspect list with deprecation path proposal (e.g., `moai harness-observe*` → mark deprecated in v3.1.0 release → remove in v3.2.0)
- T4.3: `## §4.3 Integration bridging gaps` — Code-level gaps in profile triad that Sprint 7 must address
- T4.4: `## §4.4 Sprint 7 SPEC scope draft` — A 5-section outline that Sprint 7 SPEC-V3R6-CLI-INTEGRATION-001 manager-spec can directly consume:
  - Section 1: Subcommand unification scope (consumes §4.1)
  - Section 2: Retirement scope (consumes §4.2)
  - Section 3: Integration gap bridging scope (consumes §4.3)
  - Section 4: Backward compatibility / deprecation policy
  - Section 5: Out-of-scope (what Sprint 7 will NOT do)
- T4.5: `## §5 Methodology Appendix`:
  - List all grep commands used in M1+M2+M3 (reproducibility)
  - List files scanned per class
  - List exclusions applied (e.g., comments, generated code)
  - Audit report generation date + git SHA + moai version
- T4.6: Frontmatter sync 4 artifacts (spec/plan/acceptance/progress) to status `implemented` + version `0.2.0`
- T4.7: Commit `plan(SPEC-V3R6-CLI-AUDIT-001): M4 — Sprint 7 baseline + methodology + frontmatter sync`

**Verification (binary)**:
- `grep "## §4.1 Recommended unifications" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §4.2 Recommended retirements" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §4.3 Integration bridging gaps" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §4.4 Sprint 7 SPEC scope draft" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "## §5 Methodology Appendix" .moai/reports/cli-audit/audit-*.md` returns 1
- `grep "^status: implemented" .moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md` returns 1
- `grep "^version: \"0.2.0\"" .moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md` returns 1

**Satisfies AC**: AC-CLA-004 (§4 Sprint 7 baseline 4 sub-sections + 5-section outline directly consumable) + AC-CLA-005 (spans all M1-M4 — zero protected-path diff verified at completion) + AC-CLA-006 (§5 methodology appendix reproducibility — grep commands + files scanned + metadata)

**Risk**: LOW. M4 is consolidation + frontmatter sync, well-understood pattern from prior Sprint 2 SPECs.

---

## Section C — Pre-flight Verification (5-command batch)

[HARD] Execute these 5 commands in a single parallel Bash batch BEFORE manager-develop spawn. All MUST return expected values; halt on any deviation.

```bash
# C-1: Non-test Go file count in internal/cli/ (baseline: 106)
find internal/cli -name "*.go" -not -name "*_test.go" | wc -l
# Expected: ≥100 (drift tolerance ±10 vs plan-phase baseline 106)

# C-2: AddCommand invocation count across CLI sources (baseline: 110)
grep -rn "AddCommand" internal/cli/ cmd/moai/ 2>/dev/null | wc -l
# Expected: ≥100 (drift tolerance ±10 vs plan-phase baseline 110)

# C-3: Hook script count (baseline: 32)
ls .claude/hooks/moai/ | wc -l
# Expected: ≥30 (drift tolerance ±5 vs plan-phase baseline 32)

# C-4: moai invocation reference count in hooks (baseline: 310)
grep -rn "moai " .claude/hooks/ internal/template/templates/.claude/hooks/ 2>/dev/null | wc -l
# Expected: ≥280 (drift tolerance ±30 vs plan-phase baseline 310)

# C-5: Current HEAD verification
git log --oneline -1
# Expected: matches main HEAD at plan-phase entry (136f58432 LCL-002 mx commit OR newer commit if user advanced main)
```

If C-1 through C-4 drift beyond tolerance, halt and report — substantial CLI surface change since plan-phase invalidates the M1 scope estimate.

---

## Section D — Risk Assessment

(See spec.md §D for full risk table. Plan-phase risks:)

| Risk | Severity | Mitigation |
|------|----------|------------|
| Audit incompleteness | M | M1 cross-reference 3 grep classes |
| Dead-command false positive | H | M2 requires ≥2 negative evidence sources + user review |
| Sprint 7 scope drift | M | REQ-CLA-004 constrains §4 to directly-consumable scope |
| Integration map complexity | M | 4 sub-sections decompose §3 |
| Parallel session race (L9) | L | Pre-spawn `git fetch origin` already PASS (0 0) |

---

## Section E — Deliverables Summary

| Artifact | Path | Lifecycle |
|----------|------|-----------|
| Audit report (research deliverable) | `.moai/reports/cli-audit/audit-{ISO-DATE}.md` | NEW |
| SPEC document | `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md` | MODIFIED (frontmatter sync at M4) |
| Plan document | `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/plan.md` | MODIFIED (frontmatter sync at M4) |
| Acceptance document | `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/acceptance.md` | MODIFIED (frontmatter sync at M4) |
| Progress tracker | `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` | NEW (created at M1, updated M2-M4) |

**No code, hook script, template, docs-site, or config file modification.** [HARD] REQ-CLA-005 [Unwanted].

---

Version: 0.1.0 (plan-phase draft)
