# Tasks — SPEC-V3R2-SPC-001 EARS + Hierarchical Acceptance Criteria

> Run-phase task breakdown derived from plan.md milestones. Naming convention: `T-SPC001-NN`. Owner roles use the agent catalog (manager-spec / expert-backend / manager-quality / manager-docs).

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-09 | manager-spec | Initial T-SPC001-01 through T-SPC001-12 task list. |

---

## 1. Task Index

| Task ID | Milestone | Owner | Priority | Depends on |
|---------|-----------|-------|----------|------------|
| T-SPC001-01 | M1 | manager-spec | Critical | — |
| T-SPC001-02 | M1 | manager-spec | Critical | T-SPC001-01 |
| T-SPC001-03 | M2 | expert-backend | High | T-SPC001-02 |
| T-SPC001-04 | M2 | expert-backend | High | T-SPC001-03 |
| T-SPC001-05 | M3 | expert-backend | High | T-SPC001-03 |
| T-SPC001-06 | M3 | manager-docs | High | T-SPC001-05 |
| T-SPC001-07 | M3 | manager-docs | High | T-SPC001-06 |
| T-SPC001-08 | M3 | manager-docs | Medium | T-SPC001-06 |
| T-SPC001-09 | M4 | manager-spec | Medium | T-SPC001-06 |
| T-SPC001-10 | M5 | manager-quality | Critical | T-SPC001-04, T-SPC001-05, T-SPC001-06 |
| T-SPC001-11 | M5 | manager-quality | Critical | T-SPC001-10 |
| T-SPC001-12 | M5 | manager-spec | Critical | T-SPC001-11 |

---

## 2. Task Detail

### M1 — Schema design + research delivery

#### T-SPC001-01 — Author research.md with codebase audit

Owner: manager-spec
Priority: Critical
Description: Produce research.md that captures the as-is state of EARS/AC handling in `.claude/rules/moai/workflow/spec-workflow.md`, the Go subsystem `internal/spec/`, and the SPEC corpus (185 SPECs). Compare against SPC-001 REQs and produce a gap analysis.
Inputs: spec.md, `.claude/rules/moai/workflow/spec-workflow.md`, `internal/spec/*.go`, `.moai/specs/` directory listing.
Outputs: `.moai/specs/SPEC-V3R2-SPC-001/research.md` with ≥25 file:line anchors and §11 decisions.
Acceptance: research.md committed; ≥25 anchors verified by plan-auditor.
Status: DONE (this PR).

#### T-SPC001-02 — Author plan.md, acceptance.md, tasks.md, progress.md, spec-compact.md

Owner: manager-spec
Priority: Critical
Description: Produce remaining 5 plan-phase artifacts. acceptance.md self-demonstrates hierarchical schema on ≥3 ACs. progress.md frontmatter `plan_status: audit-ready`.
Inputs: research.md (T-SPC001-01).
Outputs: 5 markdown files in `.moai/specs/SPEC-V3R2-SPC-001/`.
Acceptance: All 18 REQs mapped; tasks.md uses T-SPC001-NN naming; plan-auditor PASS at iteration ≤2.
Status: DONE (this PR).

### M2 — Performance benchmark + parser hardening

#### T-SPC001-03 — Add 365-leaf perf benchmark

Owner: expert-backend
Priority: High
Description: Add `BenchmarkParse365Leaves` to `internal/spec/parser_test.go`. Generate fixture with 55 top-level parents × ~6.6 average children = 365 leaves matching DevAI shape. Assert wall-clock parse <500ms.
Inputs: AC-SPC-001-14 in acceptance.md.
Files: `internal/spec/parser_test.go`, new `internal/spec/testdata/perf-365-leaves/spec.md`.
Outputs: Benchmark function + fixture + CI pass.
Acceptance: `go test -bench BenchmarkParse365Leaves -benchtime 1x ./internal/spec/...` reports `<500_000_000 ns/op`. Add `@MX:WARN reason="365-leaf parse perf budget"` to benchmark function.

#### T-SPC001-04 — Cover `acceptance_format: flat` malformed-frontmatter cases

Owner: expert-backend
Priority: High
Description: Add fixture cases for malformed frontmatter (missing colon, multiline value, unknown enum). Confirm parser refuses gracefully per research.md R6.
Files: `internal/spec/parser_test.go`, `internal/spec/testdata/flat-format-malformed-*/spec.md`.
Outputs: 3 new fixtures + 3 test cases.
Acceptance: `go test ./internal/spec/...` green; no panic on malformed frontmatter.

### M3 — Documentation + spec-workflow.md amendment

#### T-SPC001-05 — Audit `--shape-trace` field emission

Owner: expert-backend
Priority: High
Description: Confirm `internal/cli/spec_view.go:41-158` emits `depth` and `parent_id` for each node when `--shape-trace` is passed. Add CLI test if absent.
Files: `internal/cli/spec_view.go`, `internal/cli/spec_view_test.go` (create if missing).
Outputs: Test asserting trace output contains `depth=0`, `depth=1`, `parent=AC-XXX-NN` lines.
Acceptance: AC-SPC-001-08 test green.

#### T-SPC001-06 — Amend spec-workflow.md with hierarchical schema block

Owner: manager-docs (with manager-spec author oversight)
Priority: High
Description: Insert hierarchical-AC schema example block after Plan Phase output bullet list (around `.claude/rules/moai/workflow/spec-workflow.md:104`). Document Given inheritance and leaf-only REQ tail rule. Add `@MX:NOTE` cross-link.
Files: `.claude/rules/moai/workflow/spec-workflow.md`.
Outputs: ~30-line schema block citing research.md §6 canonical form.
Acceptance: Block renders correctly; mx-tag scan confirms NOTE present.

#### T-SPC001-07 — Update SKILL.md hierarchical AC subsection

Owner: manager-docs
Priority: High
Description: Add §"Hierarchical Acceptance" subsection to `.claude/skills/moai-workflow-spec/SKILL.md` under "EARS Format Deep Dive". Concise schema reference (Quick Reference + Implementation Guide).
Files: `.claude/skills/moai-workflow-spec/SKILL.md`.
Outputs: New subsection ~40 lines.
Acceptance: SKILL.md preserves existing 6-pillar structure; line count remains ≤500.

#### T-SPC001-08 — Cross-link zone-registry CONST-V3R2-001

Owner: manager-docs
Priority: Medium
Description: Update `.claude/rules/moai/core/zone-registry.md:16-19` CONST-V3R2-001 entry to reference SPC-001 as the amendment vehicle. No change to canary_gate value.
Files: `.claude/rules/moai/core/zone-registry.md`.
Outputs: Updated comment / cross-reference.
Acceptance: Entry references SPC-001 verbatim.

### M4 — Migration handoff to MIG-001

#### T-SPC001-09 — Author MIG-001 handoff note

Owner: manager-spec
Priority: Medium
Description: Add cross-reference to `.moai/specs/SPEC-V3R2-MIG-001/spec.md` confirming SPC-001 wrap-synthesis (`internal/spec/parser.go:200-227`) covers runtime compatibility and MIG-001 owns optional cosmetic rewrite. Add `@MX:TODO resolved="MIG-001 run phase"` on SPC-001 §2.1 migration bullet during run-phase commit.
Files: `.moai/specs/SPEC-V3R2-MIG-001/spec.md` (or its plan.md when filed).
Outputs: 5-10 line note.
Acceptance: MIG-001 confirms cosmetic rewrite ownership; no behavioural conflict.

### M5 — REFACTOR + MX tags + completion gate (CON-002 paperwork)

#### T-SPC001-10 — Run Canary re-parse against last 10 v2 SPECs

Owner: manager-quality
Priority: Critical
Description: Run the upgraded parser against the last 10 landed v2 SPECs (SPEC-V3R2-WF-005, SPEC-V3R3-CLI-TUI-001 M1/M2/M3, SPEC-V3R3-CIAUT-001, SPEC-V3R2-WF-002/003/004, SPEC-V3R2-CON-002, SPEC-V3R2-MIG-001). Capture script output to `.moai/specs/SPEC-V3R2-SPC-001/canary-v2-reparse.txt`.
Inputs: `internal/spec/parser.go`, `.moai/specs/SPEC-V3R2-*/spec.md`.
Outputs: Canary log file showing zero warnings + identical REQ-coverage set vs flat baseline.
Acceptance: Canary log committed; all 10 SPECs parse green.

#### T-SPC001-11 — Add `@MX:ANCHOR` and `@MX:NOTE` tags

Owner: manager-quality
Priority: Critical
Description: Annotate canonical sources with MX tags per plan.md §3 M5 mx_plan.
Files:
- `.moai/specs/SPEC-V3R2-SPC-001/acceptance.md` — `@MX:ANCHOR fan_in=4` on the self-demonstrating example block (cross-referenced by SPC-002, SPC-003, HRN-002, HRN-003).
- `internal/spec/ears.go:11` — `@MX:NOTE` on Acceptance struct: "Hierarchical AC schema per SPC-001 §11.2".
- `.claude/rules/moai/workflow/spec-workflow.md` (insertion point from T-SPC001-06) — `@MX:WARN reason="FROZEN-zone amendment"`.
Outputs: 3 MX-tag annotations.
Acceptance: `moai mx --validate` passes; no orphan tags.

#### T-SPC001-12 — File CON-002 amendment evidence + HumanOversight approval

Owner: manager-spec
Priority: Critical
Description: Compile FrozenGuard / Canary / ContradictionDetector / RateLimiter / HumanOversight evidence per spec.md §11.3. Record maintainer approval timestamp + reviewer in landing PR description. Update progress.md `plan_status: complete` after run-phase merges.
Inputs: T-SPC001-10 Canary log.
Files: `.moai/specs/SPEC-V3R2-SPC-001/progress.md`, landing PR description.
Outputs: PR description with 5-section evidence block; progress.md frontmatter updated.
Acceptance: PR description block complete; HumanOversight approval received.

---

## 3. Task Dependency Graph

```
T-SPC001-01 (research.md)
    └─ T-SPC001-02 (plan/AC/tasks/progress/compact)
            ├─ T-SPC001-03 (perf benchmark)
            │      └─ T-SPC001-04 (frontmatter edge cases)
            ├─ T-SPC001-05 (--shape-trace audit)
            │      ├─ T-SPC001-06 (spec-workflow.md amendment)
            │      │      ├─ T-SPC001-07 (SKILL.md)
            │      │      ├─ T-SPC001-08 (zone-registry)
            │      │      └─ T-SPC001-09 (MIG-001 handoff)
            │      └─ T-SPC001-10 (Canary re-parse)
            │             └─ T-SPC001-11 (MX tags)
            │                    └─ T-SPC001-12 (CON-002 evidence + approval)
```

---

## 4. Owner Role Reference

| Owner | Responsibilities |
|-------|------------------|
| manager-spec | Plan-phase artifact authoring (T-01, T-02), MIG-001 handoff (T-09), CON-002 evidence (T-12). |
| expert-backend | Go-level implementation (perf benchmark T-03, frontmatter edge cases T-04, --shape-trace audit T-05). |
| manager-docs | Documentation amendments (T-06 spec-workflow.md, T-07 SKILL.md, T-08 zone-registry). |
| manager-quality | Run-phase quality gate (T-10 Canary, T-11 MX tags). |

---

## 5. Run-Phase Entry Criteria

This tasks.md becomes actionable when:
- [ ] Plan PR (this) merges to main.
- [ ] plan-auditor PASS recorded.
- [ ] manager-tdd or expert-backend assigned to lead M2-M5 execution.

End of tasks.
