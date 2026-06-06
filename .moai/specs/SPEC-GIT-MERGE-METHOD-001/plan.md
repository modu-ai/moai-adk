---
spec_id: SPEC-GIT-MERGE-METHOD-001
tier: S
status: draft
created: 2026-06-06
updated: 2026-06-06
---

# Plan — SPEC-GIT-MERGE-METHOD-001

This plan is the implementation companion to `spec.md`. Per the Tier S contract in `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, the acceptance criteria live inline in `spec.md §3`; this file owns the implementation sequencing, file-level change map, PRESERVE list, and the open decision questions surfaced for plan-auditor review.

---

## §A.1 — Milestone Map

The implementation decomposes into five small milestones (M1-M5). Total estimated scope: ~80 LOC across 6 files, well under the Tier S 300 LOC / 5-files-affected guidance. Files-affected count is 6 because the mirror parity invariant (per `internal/template/CLAUDE.md`) requires regenerating `embedded.go`, which is the auto-generated 6th file.

| Milestone | Scope | Affected files | LOC estimate | Bound REQ / AC |
|-----------|-------|----------------|--------------|-----------------|
| M1 | Add `merge_method:` YAML key to all three mode profiles in `git-strategy.yaml.tmpl` + inline comment documenting accepted values | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | ~12 | REQ-GMM-001 / AC-GMM-001 |
| M2 | Add `MergeMethod string` field to per-mode profile struct in `types.go` + add `DefaultMergeMethod = "squash"` constant in `defaults.go` + extend validation/loader if needed; SAME COMMIT as M1 to preserve struct-yaml symmetry CI guard | `internal/config/types.go`, `internal/config/defaults.go` | ~20 | REQ-GMM-002 / AC-GMM-002 |
| M3 | Replace the two literal `gh pr merge --squash --delete-branch` invocations in `sync/delivery.md` Step 3.4 with a config-driven prose substitution. The new prose SHALL instruct the agent to read `git_strategy.<active-mode>.merge_method` and substitute the flag accordingly | `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | ~10 | REQ-GMM-003 / AC-GMM-003 |
| M4 | Replace the three `--squash` literals in `manager-git.md` (lines :140, :176, :204) with the same config-driven shape (per the §1.6 Q1 default recommendation, additionally wire through `internal/github.PRMerger` if the M4 deliberation surfaces it as in-scope; otherwise leave the wire-through as a §B.4 follow-up) | `internal/template/templates/.claude/agents/moai/manager-git.md` | ~12 | REQ-GMM-004 / AC-GMM-004 |
| M5 | Amend the FROZEN-rule sync row cell in `spec-workflow.md:31` to `configurable (default: squash)` + add SPEC cross-reference footnote; add `merge_method` cross-reference to `moai-ref-git-workflow/skill.md` near :133-134 / :171 | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`, `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md` | ~10 | REQ-GMM-005 + REQ-GMM-006 / AC-GMM-005 + AC-GMM-006 |
| M6 (final) | Run `make build`; verify `embedded.go` regenerates cleanly; run the full Go test suite to confirm AC-GMM-002 sub-check 3 + AC-GMM-007 + AC-GMM-008 fixture-test | `internal/template/embedded.go` (auto-generated) | n/a (regenerated) | AC-GMM-007 + AC-GMM-008 |

M1 + M2 SHALL land in a single commit to preserve the struct-yaml symmetry guard invariant. M3-M5 MAY land as separate commits but SHALL all land within the same PR.

## §A.2 — File-level Change Map

| File | Action | Why |
|------|--------|-----|
| `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | EDIT (add 3 keys) | M1 (REQ-GMM-001) |
| `internal/config/types.go` | EDIT (add 1 struct field per profile substruct) | M2 (REQ-GMM-002) |
| `internal/config/defaults.go` | EDIT (add 1 constant + 3 default population sites) | M2 (REQ-GMM-002) |
| `internal/config/loader_test.go` (or the Round 5 git-strategy test file) | EDIT (add 1 fixture + 1 assertion) | AC-GMM-008 backward-compat test |
| `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | EDIT (replace 2 hardcoded lines :313, :325) | M3 (REQ-GMM-003) |
| `internal/template/templates/.claude/agents/moai/manager-git.md` | EDIT (replace 3 hardcoded lines :140, :176, :204) | M4 (REQ-GMM-004) |
| `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | EDIT (one cell + one footnote in lifecycle table) | M5 (REQ-GMM-005) |
| `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md` | EDIT (add cross-reference near :133-134 or :171) | M5 (REQ-GMM-006) |
| `internal/template/embedded.go` | REGENERATE (`make build`) | M6 (mirror parity invariant) |

No new files are created. No files are deleted. No template files outside `internal/template/templates/` are touched (per CLAUDE.local.md §2 Template-First Rule).

## §A.3 — Sequencing Rationale

The order M1+M2 → M3 → M4 → M5 → M6 is dictated by three constraints:

1. **Struct-yaml symmetry CI guard**: M1 (YAML key) + M2 (Go struct field + default constant) MUST land in the same commit, otherwise `audit_struct_yaml_symmetry_test.go:TestStructYAMLSymmetry_*` fails between the two commits. This is the dominant ordering constraint.
2. **Workflow body depends on config existing**: M3 and M4 reference `git_strategy.<mode>.merge_method`; they cannot land before M1+M2 register the key. (Reverse-ordering would not technically break tests since markdown prose is not parsed by the Go suite, but it would create a transient state where the workflow references a non-existent key.)
3. **FROZEN-rule amendment is the contract-public statement**: M5 amends the table cell that pins `squash` for sync; it MUST follow M3+M4 since the FROZEN rule is what blesses the new configurability. Landing M5 before M3+M4 would relax the contract before any implementation honors the new freedom.
4. **`make build` is the final gate**: M6 regenerates `embedded.go` and runs the full template-mirror test; this is always last in MoAI template-edit workflows per `internal/template/CLAUDE.md`.

## §A.4 — Pre-flight check list (manager-develop entry)

Before M1, the implementer SHALL run:

```bash
# 1. Confirm clean baseline
git status --porcelain
# Expected: empty

git rev-parse --abbrev-ref HEAD
git rev-parse HEAD

# 2. Confirm Go test baseline passes
go test ./internal/config/... ./internal/template/...
# Expected: PASS — this is the pre-SPEC baseline

# 3. Confirm grep counts match §1.2 evidence (no surprises)
grep -c '^\s*gh pr merge .*--squash --delete-branch\b' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
# Expected: 2

grep -c 'gh pr merge.*--squash --delete-branch' internal/template/templates/.claude/agents/moai/manager-git.md
# Expected: 3

grep -c 'merge_method' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
# Expected: 0

# 4. Confirm the unused abstraction exists (so §1.6 Q1 has a target)
grep -nE 'func PRMerge|func.*PRMerger' internal/github/gh.go internal/github/pr_merger.go
# Expected: ≥2 matches

# 5. Confirm the cross-platform build still works (per coding-standards.md §2.1 B1)
GOOS=windows GOARCH=amd64 go build ./...
# Expected: exit 0
```

## §A.5 — PRESERVE list (do NOT modify)

Per CLAUDE.local.md §16 scope discipline and the §B10 Untouched Paths convention from `.claude/rules/moai/development/manager-develop-prompt-template.md`:

- `internal/cli/worktree/sync.go` — its `--strategy merge|rebase` flag is for worktree sync, not PR merge. §1.4 already disambiguates; touching this file is OUT OF SCOPE.
- `internal/github/pr_merger.go` runtime semantics — the default-method (`"merge"`) in `pr_merger.go:134` SHALL NOT be changed by this SPEC. If §1.6 Q1 resolves as "wire through `PRMerger`", the wire-through SHALL pass the configured method explicitly per call; the package-level default stays `"merge"` as it is today.
- All `_test.go` files in `internal/github/` — extending the existing test coverage is fine; rewriting or removing tests is OUT OF SCOPE.
- All other `.moai/specs/SPEC-*/` directories — only this SPEC's directory is touched.
- All other `.claude/rules/`, `.claude/skills/`, `.claude/agents/` files NOT listed in §A.2 — scope discipline.
- All `.moai/state/`, `.moai/cache/`, `.moai/logs/`, `.moai/harness/` runtime-managed files.
- `CLAUDE.md`, `CLAUDE.local.md`, README files, CHANGELOG.md (CHANGELOG gets a sync-phase entry, but the run-phase SPEC implementation itself does not touch it).

## §B.1 — Known Issues / Risk-Mitigation Hooks (manager-develop Section B mapping)

Mapped against `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B B1-B12 categories:

- **B1 Cross-platform build tags**: not applicable — no syscall usage in this SPEC.
- **B2 Cross-SPEC policy conflict**: dependency D1 (`SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` completed) is the only cross-SPEC reference. No retired/superseded SPECs touch this scope. Pre-flight grep at §A.4 step 4 confirms the `internal/github` abstraction has no `Retired` markers.
- **B3 Subagent boundary discipline (C-HRA-008)**: not applicable — no harness/hook code touched.
- **B4 Frontmatter canonical schema**: spec.md frontmatter uses the canonical 12 fields per `.claude/rules/moai/development/spec-frontmatter-schema.md`; the `issue_number: 1061` optional field is correctly placed.
- **B5 CI 3-tier awareness**: the symmetry CI guard (`audit_struct_yaml_symmetry_test.go`) + mirror parity test (`embedded_mirror_test.go`) + `go test ./...` are the three regression layers. All three SHALL be exercised by M6.
- **B6 spec-lint heading regime**: spec.md uses `## 1. 개요`, `## 2. EARS Requirements`, `## 3. Acceptance Criteria`, `## 4. Risk Register`, `## 5. Edge Cases`, `## 6. Dependencies & Sequencing`, `## 7. References` — standard Tier S structure with AC inline at §3.
- **B7 observer.go / capture path**: not applicable.
- **B8 working tree hygiene**: PRESERVE list at §A.5 enumerates every path that must NOT be touched.
- **B9 Git commit + push self-execution (Hybrid Trunk 1-person OSS)**: per CLAUDE.local.md §23 Tier S = main-direct push; manager-develop self-executes commits + push for this SPEC. Conventional Commits format `feat(SPEC-GIT-MERGE-METHOD-001): M{N} <subject>`.
- **B10 untouched paths PRESERVE**: see §A.5 above.
- **B11 AskUserQuestion ban for subagents**: any blocker encountered during M1-M6 SHALL be returned as a structured blocker report; the orchestrator owns the user dialog.
- **B12 sync-phase CHANGELOG discipline (manager-docs)**: not applicable in run-phase; manager-docs handles sync-phase per the canonical Status Transition Ownership Matrix.

## §B.2 — Open Decision Questions (replicated from spec.md §1.6 for plan-auditor visibility)

These three questions affect the FROZEN-rule relaxation contract and the implementation scope. plan-auditor should evaluate whether the recommended defaults are defensible:

- **Q1 — Wire-through to `internal/github.PRMerger`**: Recommended = **wire through**. Rationale: eliminates the "complete but unused abstraction" defect that the issue specifically flags as a code smell; routes through tested code paths (`pr_merger_test.go` has 30 refs). Counter-argument: keeps M4 strictly markdown-prose (no Go code changes in M4); strictly minimal scope. plan-auditor should weigh "minimum viable" against "leave the abstraction unused for a third time".
- **Q2 — Per-branch-type overrides (gitflow `release/*` → `merge`)**: Recommended = **defer**. Rationale: keeps this SPEC Tier S; the extension shape is additive (`merge_method: squash` → `merge_method: {default: squash, overrides: {release: merge, hotfix: merge}}`) so a follow-up can land it without breaking change. Counter-argument: the contradiction between `delivery.md:252` and `moai-ref-git-workflow/skill.md:133-134` is closed by this SPEC's REQ-GMM-006 cross-reference but not by behavior. If plan-auditor judges the contradiction must be closed in implementation too, this SPEC tier-ups to M.
- **Q3 — Scope of FROZEN-rule relaxation**: Recommended = **sync row only**. Rationale: matches issue scope (the issue body cites the sync auto-merge); leaves plan/run rows untouched until a future SPEC demands them. Counter-argument: relaxing all three rows symmetrically is cleaner contract-wise. plan-auditor should consider whether asymmetric relaxation (sync configurable, plan/run pinned) is a future debt source.

## §B.3 — Implementation hint for M4 §1.6 Q1 (if "wire through" wins)

If Q1 resolves as "wire through `PRMerger`", M4 also requires editing `internal/github/pr_merger.go` to accept the configured method as an explicit parameter from the calling code path. The actual Go wiring site (i.e. where `/moai sync` ultimately invokes the merge in production Go code, if any) is presently absent — per §1.2 evidence #5, every call to `PRMerge` / `PRMerger` is test-only. So the M4 "wire through" decision implies also adding the first production call site, which expands scope from "thin template edits" to "one new Go production path".

If plan-auditor judges this expansion makes the SPEC tier-up to M, the §A.1 milestone table should split M4 into M4a (markdown edits) + M4b (new Go production call site) — but this is a Q1 outcome, not a pre-decided expansion.

## §B.4 — Out-of-scope follow-up extension shapes (documented for forward audit)

These are NOT implemented by this SPEC but are documented here so a future SPEC has a clean attachment point:

- **F1 — Per-branch-type override (deferred from Q2)**: extend the `merge_method` value from a string enum to a `{default: <method>, overrides: {<prefix>: <method>}}` shape. Backward compat: when value is a string, treat as `{default: <string>, overrides: {}}`. Implementation site: same per-mode profile substruct; loader gains a custom `UnmarshalYAML` for the union type. Affects: REQ-GMM-002 only (REQ-GMM-001 needs a one-line schema doc update; the rest are unchanged because they still read a single resolved method per branch).
- **F2 — Wire through `internal/github.PRMerger` (deferred from Q1)**: replace the agent-level `gh pr merge` shell-out at `manager-git.md` with a call into the Go production code path that uses `PRMerger`. Requires adding the first production caller for `PRMerger`. This eliminates the unused-abstraction defect for good.
- **F3 — CI guard against future hardcoded `--squash` reintroduction**: add a `audit_no_hardcoded_merge_method_test.go` in `internal/template/` that greps the workflow template files for the literal `--squash --delete-branch` and fails the build if any match exists. This guard would belong in a separate "doctrine enforcement" SPEC rather than here.
- **F4 — Relax plan/run rows of the FROZEN lifecycle table (deferred from Q3)**: if a future SPEC needs configurable plan-PR or run-PR merge methods, it amends the same table cells the same way. The §A.5 mirror-parity invariant and §A.4 step 5 cross-platform build still apply.

## §B.5 — Plan-audit checklist (verbatim signals plan-auditor should look for)

- [ ] §1.2 evidence is verifiable (re-run all 5 greps against working tree commit).
- [ ] All 6 REQs map to at least one AC; every AC binds to exactly one REQ.
- [ ] AC-GMM-007 and AC-GMM-008 cover the regression-hazard surface (mirror parity + backward compat).
- [ ] PRESERVE list at §A.5 explicitly excludes `internal/cli/worktree/sync.go` (the §1.4 disambiguation).
- [ ] §B.2 open decisions surface §1.6 Q1/Q2/Q3 to the audit gate rather than burying them.
- [ ] Tier S 2-artifact rule honored (spec.md + plan.md; no acceptance.md needed since AC inline §3).
- [ ] FROZEN-rule amendment scope (REQ-GMM-005) is exactly one cell + one footnote; the `[ZONE:Frozen] [HARD]` zone marker stays.
