# SPEC-INVOCATION-MODEL-002 — Progress

Lifecycle: plan → run → sync (3-phase). Tier S. Status: draft.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-07-01
plan_status: audit-ready
tier: S
artifacts:
  - spec.md          # GEARS REQ-IM2-001..011, §A authoritative record, §E exclusions
  - plan.md          # Tier S, per-mapping Axis-A capability verdicts, Template-First mirror plan, milestones M1-M2
  - acceptance.md    # AC-IM2-001..011 (grep/diff-verifiable), Definition of Done
  - progress.md      # this file
deliverables:
  - divergence_reconciliation:
      authoritative_record: spec.md §A + §C (6 PROGRAMMATIC / 3 HUMAN-ONLY)
      errata_pointer: single line appended to closed SPEC-INVOCATION-MODEL-001/spec.md (run-phase)
      rule_file_edit: none (native-invocation-model.md already correct; SPEC-ID cross-ref scoped out)
  - axis_a_alignment:
      clean_simplify: scoped OUT (capability mismatch — dead-code removal != changed-code quality refactoring)
      review_code_review: COMPOSE (native /code-review via Skill() as one Phase 2 component; Security/@MX/UX/design preserved; sync-auditor fallback)
axis_a_verdicts:
  clean_to_simplify: scoped-out
  review_to_code_review: compose-not-swap
run_phase_edit_surface:
  - .claude/skills/moai/workflows/review.md            # local compose edit
  - internal/template/templates/.claude/skills/moai/workflows/review.md   # template mirror (neutral)
  - .moai/specs/SPEC-INVOCATION-MODEL-001/spec.md       # 1 errata line (closed body exception)
milestones:
  - M1: review↔/code-review compose (local + template mirror) + make build
  - M2: divergence errata-pointer line append
constraints:
  - no hook / no lint rule / no Go runtime (doctrine is codification-only)
  - closed spec.md: single errata line only, §A immutable
  - template review.md neutrality: no SPEC ID / REQ token / date / SHA
  - clean.md + native-invocation-model.md untouched
```

_Plan-phase artifacts complete and audit-ready. Awaiting Implementation Kickoff Approval before run-phase._

## §E.2 Run-phase Evidence

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-IM2-001 | PASS | `grep -c "\| PROGRAMMATIC \|" spec.md` + `grep -c "\| HUMAN-ONLY \|"` | 6 PROGRAMMATIC / 3 HUMAN-ONLY; /security-review + /review both PROGRAMMATIC |
| AC-IM2-002 | PASS | `grep -c "Errata (SPEC-INVOCATION-MODEL-002)" SPEC-001/spec.md` | 1 (single errata line at file tail) |
| AC-IM2-003 | PASS | `git diff SPEC-001/spec.md` | `@@ -140,3 +140,4 @@` — 1 insertion, 0 deletions; §A body byte-unchanged |
| AC-IM2-004 | PASS | `git status native-invocation-model.md (local+template)` | 0 (rule file not in changeset) |
| AC-IM2-005 | PASS | `grep -ic "scoped OUT\|capability mismatch"` + `grep -ic "dead-code"` + clean.md status | 11 / 3 matches; clean.md status 0 (untouched) |
| AC-IM2-006 | PASS | `awk Phase2 \| grep -c 'code-review'`; `grep -c 'Skill("code-review")'`; heading count; 4 composition headings | 4 / 2 / 1; Perspective1=1, Phase3=1, Perspective4=1, Phase4.5=1 (all preserved) |
| AC-IM2-007 | PASS | `awk Phase2 \| grep -Ec 'disable-model-invocation\|disableBundledSkills'`; `grep -c 'not auto-invocable'`; `grep -c 'sync-auditor'` | 1 (caveat) / 1 (fallback) / 4 (sync-auditor co-located) |
| AC-IM2-008 | PASS | `grep -c "Dependency Vulnerability Scan"`; `grep -c "Secrets Scan (Full Git History)"` | 1 / 1 (Security sub-sections preserved; compose uses augment framing) |
| AC-IM2-009 | PASS | `diff review.md(local) review.md(template)` | empty output, exit 0 (byte-identical); `make build` exit 0 |
| AC-IM2-010 | PASS | `git status \| grep -E '.claude/hooks/\|internal/spec/lint.*.go\|.go$'` | NO-RUNTIME-CHANGE (doc/prose only) |
| AC-IM2-011 | PASS | `grep -c "SPEC-INVOCATION" template`; `grep -Ec "REQ-IM2\|2026-07-01" template`; `go test TestTemplateNoInternalContentLeak` | 0 / 0; `--- PASS: TestTemplateNoInternalContentLeak (0.35s)` |

Invariants:
- PRESERVE (clean.md local+template, native-invocation-model.md local+template) — untouched (git status 0). **Status: PASS**
- Perspective 1/2/3/4 + Phase 3 MX + Phase 4.5 design headings in review.md — all preserved (grep ==1 each). **Status: PASS**
- Closed SPEC-001 §A body — byte-unchanged (git diff shows only 1 appended line at tail). **Status: PASS**
- Template neutrality — no SPEC ID / REQ token / date in template review.md (grep 0/0). **Status: PASS**

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-01
run_commit_sha: <backfill — this run commit (worktree branch worktree-agent-ad895cb0c2c0d8063; orchestrator reconciles to main)>
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 4   # clean.md(local+template) + native-invocation-model.md(local+template) all untouched
l44_pre_commit_fetch: "git fetch origin main; git rev-list --count --left-right origin/main...HEAD"
l44_post_push_fetch: <deferred — orchestrator performs push/reconcile to main per B9 worktree-isolation exception>
new_warnings_or_lints_introduced: 0   # doc/prose-only edit; no Go code touched
cross_platform_build:
  note: "N/A — no Go source changed; make build (gen-catalog-hashes + go build) exit 0"
  make_build_exit: 0
  catalog_yaml_delta: 0   # moai skill workflow sub-files not catalog-hashed; catalog.yaml byte-unchanged
total_run_phase_files: 5   # review.md(local) + review.md(template) + SPEC-001/spec.md(errata) + SPEC-002/spec.md(frontmatter) + SPEC-002/progress.md(§E)
m1_to_mN_commit_strategy: "single M1 commit (Tier S; M1 compose+mirror+build and M2 errata bundled — independent files, low-risk append)"
worktree_isolation_note: "manager-develop ran in isolated L1 worktree; SPEC-002 plan artifacts were untracked in parent and copied in; commit lands on worktree branch, orchestrator reconciles the 3 tracked-file diffs + SPEC-002 dir to the (moving) main per B9 exception + shared-main orphan-race defense."
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
