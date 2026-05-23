---
id: SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001
title: "Template mirror drift cleanup: 4-file mechanical mirror parity — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "template-mirror, drift-fix, sprint-7-entry, tier-s, mechanical-cleanup"
---

# SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 — Implementation Plan (Tier S minimal Section A-E)

## §A. Context

### §A.1 SPEC artifacts location

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Branch: `main` (Hybrid Trunk 1-person OSS per CLAUDE.local.md §23.7)
- HEAD SHA at plan-phase commit: `38a638d3c` (Sprint 2 P4 trio close, push range will be derived at run-phase commit)
- SPEC artifacts (4 files):
  - `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/spec.md` (canonical SSOT)
  - `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/plan.md` (this file)
  - `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/acceptance.md` (5 ACs canonical matrix)
  - `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/progress.md` (lifecycle + audit-ready signal)
- plan-auditor verdict: TBD at plan-phase write completion (target: PASS ≥ 0.75 Tier S threshold)

### §A.2 Existing infrastructure (PRESERVE vs EXTEND)

**PRESERVE (unchanged across fix commit) — 11 dirty/untracked entries per L45 discipline**:

| # | Path | Type | Source |
|---|------|------|--------|
| 1 | `.claude/output-styles/moai/einstein.md` | M | local edit |
| 2 | `.claude/output-styles/moai/moai.md` | M | local edit |
| 3 | `.moai/config/sections/git-convention.yaml` | M | local edit |
| 4 | `.moai/config/sections/language.yaml` | M | §22 dev settings |
| 5 | `.moai/config/sections/quality.yaml` | M | §22 dev settings |
| 6 | `.moai/harness/usage-log.jsonl` | M | runtime-managed |
| 7 | `internal/template/templates/.claude/output-styles/moai/einstein.md` | M | template mirror emergent |
| 8 | `internal/template/templates/.claude/output-styles/moai/moai.md` | M | template mirror emergent |
| 9 | `.moai/harness/observations.yaml` | ?? | runtime-managed |
| 10 | `.moai/research/v3.0-redesign-2026-05-23.md` | ?? | parallel session research |
| 11 | `i18n-validator` | ?? | parallel session artifact |

Total `git status --porcelain | wc -l` = 11 (verified 2026-05-24 pre-plan).

**EXTEND (modified in this SPEC) — 5 entries exactly**:

| # | Path | Type | Operation |
|---|------|------|-----------|
| 1 | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | mirror file | byte-for-byte cp overwrite from source |
| 2 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | mirror file | byte-for-byte cp overwrite from source |
| 3 | `internal/template/templates/.claude/agents/meta/plan-auditor.md` | mirror file | byte-for-byte cp overwrite from source |
| 4 | `internal/template/templates/.claude/rules/moai/core/hooks-system.md` | mirror file | byte-for-byte cp overwrite from source |
| 5 | `internal/template/rule_template_mirror_test.go` | Go test registry | insert 1 entry + 1 comment line into `workflowOptMirroredPaths` slice |

### §A.3 plan-auditor expectation

- Tier S threshold: PASS ≥ 0.75 (per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier)
- Target: iter-1 PASS ≥ 0.85 (margin ≥ 0.10) for skip-eligibility at Phase 0.5 in next /moai run invocation
- MP-1 REQ consistency / MP-2 EARS / MP-3 frontmatter / MP-4 lang neutrality expected to pass

## §B. Known Issues (B1-B12 filtered for relevance)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability, Section B is filtered for Tier S minimal — only relevant categories enumerated:

### B-relevant.1 — Cross-platform Build Tags (B1)

Not relevant — this SPEC modifies only `.md` mirrors (no syscall package usage) and a Go test file that uses only `os`, `bytes`, `filepath`, `testing` (cross-platform safe).

### B-relevant.2 — Cross-SPEC 정책 충돌 사전 스캔 (B2)

Scope: Sprint 2 P4 trio recently closed (`38a638d3c`); no retired/superseded SPEC conflicts with this scope. Verified via:

```bash
$ grep -rn "Retired\|TestHarnessRetirement\|deprecation-marker" internal/template/ 2>&1 | head -10
# expected: no matches in `internal/template/rule_template_mirror_test.go` scope
```

### B-relevant.3 — C-HRA-008 / Subagent Boundary Discipline (B3)

Not applicable — `internal/template/rule_template_mirror_test.go` is test code in `package template_test`; no `AskUserQuestion` invocation possible. Verification:

```bash
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/template/rule_template_mirror_test.go
# expected: 0 matches
```

### B-relevant.4 — Frontmatter Canonical Schema (B4)

All 4 SPEC artifacts use canonical 12 fields per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT: `id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags`. `created:`/`updated:`/`tags:` used (NOT `created_at`/`updated_at`/`labels`). `tags:` CSV string (NOT YAML array) per `internal/spec/lint.go` `SPECFrontmatter.Tags string yaml:"tags"` binding.

### B-relevant.5 — CI 3-tier 인지 (B5)

- spec-lint: spec.md emits `✓ No findings`; plan/acceptance/progress MAY emit `MissingExclusions` ERROR (accepted derived-artifact lint surface per §D.1 in spec.md)
- golangci-lint: 0 new issues introduced (test registry edit is single-line slice insertion)
- Test (per OS): `go test ./internal/template/` PASS on linux/darwin/windows (cross-platform safe)

### B-relevant.6 — spec-lint Heading 규약 (B6)

spec.md uses `## §B. What scope (and what is explicitly out-of-scope)` heading + `### §B.2 Out of Scope (deferred to Sprint 8 or follow-up SPECs)` h3 sub-section — `MissingExclusions` requirement satisfied.

### B-relevant.7 — observer.go / capture path resolution (B7)

Not applicable — no `observer.go` modification in scope.

### B-relevant.8 — Working Tree Hygiene (B8)

PRESERVE list 11 entries above §A.2. Path-specific `git add <5 paths>` discipline only. `--no-verify` absolutely forbidden. Runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/`) untouched.

### B-relevant.9 — Git Commit + Push 자체 수행 (B9)

manager-develop performs commit + push for this SPEC scope. Conventional Commits format obligation: `fix(SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001): M1 — 4-file mirror parity cleanup (A1-A4) + rule_template_mirror_test.go registry add`. Korean commit body per `git_commit_messages: ko`. `🗿 MoAI <email@mo.ai.kr>` trailer obligatory.

### B-relevant.10 — Untouched Paths PRESERVE — Scope Discipline (B10)

Per L40 + L45 strict enforcement. PRESERVE list 11 entries above §A.2 unchanged across all M1 commits.

### B-relevant.11 — AskUserQuestion 금지 — Subagent Boundary (B11)

manager-develop is a subagent — AskUserQuestion invocation prohibited per CLAUDE.md §8 + askuser-protocol.md §Orchestrator-Subagent Boundary. Blocker발견 시 structured blocker report 반환 (orchestrator-direct fix-forward 회피). Blocker report 4-option format with change/impact/risk/ETA per agent-common-protocol.md.

### B-relevant.12 — Sync-phase CHANGELOG emission discipline (B12)

Applies to manager-docs at /moai sync phase (NOT this M1 run-phase). manager-docs MUST: (a) Read every implementation file referenced in plan.md before drafting; (b) `grep -c '<SPEC-ID>' CHANGELOG.md` pre-emission verify = 0; (c) AC count match `acceptance.md` SSOT (5 ACs). For B12 9th self-test PASS streak preservation. Origin: SPEC-V3R6-CHANGELOG-CLEANUP-001 §A.4.

## §C. Pre-flight Check List (M1 시작 전 manager-develop 의무 실행)

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current             # expect: main
git rev-parse HEAD                    # expect: 38a638d3c (or descendant post-Sprint-2-close)

# 2. L44 HARD pre-spawn fetch
git fetch origin main 2>&1
git rev-list --count --left-right origin/main...HEAD  # expect: 0 0

# 3. Cross-platform build 사전 확인 (test file edit affects all OS)
go build ./...                        # expect: exit 0
GOOS=windows GOARCH=amd64 go build ./...  # expect: exit 0

# 4. Lint baseline 측정
golangci-lint run --timeout=2m 2>&1 | tail -5  # capture baseline; expect: 0 issues.

# 5. PRESERVE list snapshot
git status --porcelain | tee /tmp/preserve-list-pre.txt
# expect: 11 entries (5 M base + 2 M template-mirror emergent + 3 ??)

# 6. Pre-fix test FAIL evidence (for progress.md)
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -v 2>&1 | tail -20
# expect: FAIL spec-workflow.md, agent-common-protocol.md, plan-auditor.md
# (hooks-system.md not yet registered — will be added by M1)

# 7. Drift evidence enumeration
for src in \
    .claude/rules/moai/workflow/spec-workflow.md \
    .claude/rules/moai/core/agent-common-protocol.md \
    .claude/agents/meta/plan-auditor.md \
    .claude/rules/moai/core/hooks-system.md ; do
  mirror="internal/template/templates/${src}"
  echo "=== $src vs $mirror ==="
  wc -c -l "$src" "$mirror"
  echo "diff line count: $(diff "$src" "$mirror" | wc -l)"
done
```

## §D. Constraints (DO NOT VIOLATE)

### D.1 PRESERVE 대상 enumeration

11 entries per §A.2. Path-specific `git add` only. `git status --porcelain` post-M1-commit MUST show:
- Pre-existing 11 PRESERVE entries unchanged status (still M or ?? as before)
- 0 staged paths after the M1 commit pushes
- No new untracked files introduced

### D.2 금지 명령

- `--no-verify` (any git command)
- `--amend` (creates new commit per CLAUDE.md anti-amend policy)
- `git reset --hard` (per CLAUDE.local.md §23.5 — use `--keep` if needed)
- Force-push to main (per CLAUDE.local.md §23.7 Hybrid Trunk discipline)
- `git add -A` or `git add .` (path-specific only)

### D.3 사용 의무 명령

- Conventional Commits format: `fix(SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001): M1 — 4-file mirror parity cleanup (A1-A4) + rule_template_mirror_test.go registry add`
- Korean commit body per `git_commit_messages: ko`
- `🗿 MoAI <email@mo.ai.kr>` trailer
- HEREDOC for commit message: `git commit -m "$(cat <<'EOF' ... EOF)"`

### D.4 Binary constraints (grep / diff 0 매치)

- `git diff HEAD~1..HEAD -- .claude/rules/moai/workflow/spec-workflow.md` = 0 lines (source unchanged per REQ-TMD-006)
- `git diff HEAD~1..HEAD -- .claude/rules/moai/core/agent-common-protocol.md` = 0 lines (source unchanged)
- `git diff HEAD~1..HEAD -- .claude/agents/meta/plan-auditor.md` = 0 lines (source unchanged)
- `git diff HEAD~1..HEAD -- .claude/rules/moai/core/hooks-system.md` = 0 lines (source unchanged)
- `diff <src> <mirror>` post-fix = 0 lines for all 4 pairs
- `git status --porcelain` PRESERVE 11 entries verbatim status preservation

## §E. Self-Verification Deliverables (M1 완료 보고 시 manager-develop 필수)

### E.1 AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Expected Output |
|----|--------|---------------------|-----------------|
| AC-TMD-001 | TBD | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/spec-workflow.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/spec-workflow.md` |
| AC-TMD-002 | TBD | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/agent-common-protocol.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/agent-common-protocol.md` |
| AC-TMD-003 | TBD | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/plan-auditor.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/plan-auditor.md` |
| AC-TMD-004 | TBD | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/hooks-system.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/hooks-system.md` (NEW subtest after registry add) |
| AC-TMD-005 | TBD | `go vet ./...` + `golangci-lint run --timeout=2m` | vet exit 0 + lint `0 issues.` |

### E.2 Cross-Platform Build 결과

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E.3 Coverage 측정 (test registry add scope)

```
$ go test -cover ./internal/template/
ok  github.com/modu-ai/moai-adk/internal/template  X.XXXs  coverage: NN.N% of statements
```

Expected delta: minimal (single new subtest activation; statement coverage of `rule_template_mirror_test.go` unchanged or slightly increased due to additional subtest path execution).

### E.4 Subagent Boundary Grep (C-HRA-008)

```
$ grep -rn 'AskUserQuestion' internal/template/rule_template_mirror_test.go
(no output expected — file has 0 AskUserQuestion references)
```

### E.5 Lint Status — NEW vs baseline 구분

```
$ golangci-lint run --timeout=2m
0 issues.
```

Pre/post delta MUST be 0 issues introduced.

### E.6 Branch HEAD + Push 상태

- 새 commits SHA: 1 commit (M1 fix + test registry add bundled) OR 2 commits (M1 split into A1-A4 mirror cp + A4b test registry add) — manager-develop discretion
- `git push origin main` 결과: post-push HEAD verified

### E.7 Blocker Report (if any)

If blocker found (e.g., source file modified between pre-flight and M1 commit, parallel session race detected via L44 pre-commit fetch returning `N 0`), structured 4-option blocker report returned (orchestrator-direct fix-forward avoided). No `AskUserQuestion` invocation.

## §F. Implementation Milestone (M1 only — Tier S single milestone)

### M1 — 4-mirror cp + test registry add

**Action sequence**:

1. Pre-flight checks per §C above
2. 4 mechanical mirror cp:
   ```bash
   cp .claude/rules/moai/workflow/spec-workflow.md \
      internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
   cp .claude/rules/moai/core/agent-common-protocol.md \
      internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
   cp .claude/agents/meta/plan-auditor.md \
      internal/template/templates/.claude/agents/meta/plan-auditor.md
   cp .claude/rules/moai/core/hooks-system.md \
      internal/template/templates/.claude/rules/moai/core/hooks-system.md
   ```
3. Edit `internal/template/rule_template_mirror_test.go` to insert new entry:
   - Locate lines 47-50 (current `core/` group: `agent-common-protocol.md` at line 48, `spec-workflow.md` at line 50)
   - Insert new comment + entry between line 48 and line 50:
     ```go
     // (new entry — REQ-TMD-005 — hooks-system.md mirror parity)
     ".claude/rules/moai/core/hooks-system.md",
     ```
   - Position: after `.claude/rules/moai/core/agent-common-protocol.md` entry, before the Layer E `spec-workflow.md` entry, keeping the `core/` group lexically clustered
4. Post-edit verify:
   ```bash
   for src in <4 sources>; do diff "$src" "internal/template/templates/${src}" | wc -l; done
   # expect: 0 for all 4
   go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -v
   # expect: all 4 source-modified subtests PASS
   go vet ./...                          # expect: exit 0
   golangci-lint run --timeout=2m        # expect: 0 issues.
   ```
5. Commit + push per §D.3:
   ```bash
   git add internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md \
           internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
           internal/template/templates/.claude/agents/meta/plan-auditor.md \
           internal/template/templates/.claude/rules/moai/core/hooks-system.md \
           internal/template/rule_template_mirror_test.go
   git status  # verify: ONLY 5 paths staged, 11 PRESERVE unchanged
   # Pre-commit L44 HARD fetch
   git fetch origin main && git rev-list --count --left-right origin/main...HEAD  # expect: 0 0
   git commit -m "$(cat <<'EOF'
   fix(SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001): M1 — 4-file mirror parity cleanup (A1-A4) + rule_template_mirror_test.go registry add

   Sprint 7 entry SPEC. Sprint 2 P4 trio (IVB-001/SARM-001/TMC-001) 후속.
   TEMPLATE-MIRROR-DRIFT-001 family mechanical cleanup 첫 단계.

   Scope (5 files):
   - A1 spec-workflow.md mirror cp (+2654 bytes)
   - A2 agent-common-protocol.md mirror cp (+2269 bytes)
   - A3 plan-auditor.md mirror cp (+2264 bytes)
   - A4 hooks-system.md mirror cp (+745 bytes) + rule_template_mirror_test.go registry add

   5/5 ACs PASS: AC-TMD-001..004 TestRuleTemplateMirrorDrift PASS + AC-TMD-005 vet/lint 0 issues.
   Sources untouched per REQ-TMD-006. PRESERVE 11 files verbatim.

   🗿 MoAI <email@mo.ai.kr>
   EOF
   )"
   git push origin main
   git rev-list --count --left-right origin/main...HEAD  # expect: 0 0 post-push
   ```

### M1 acceptance gate

All 5 ACs PASS per §E.1 matrix. Manager-develop self-verifies and reports per §E. If any AC FAIL, return blocker report to orchestrator (no orchestrator-direct fix-forward in run-phase).

## §G. Anti-Patterns to avoid

- Modifying any of the 4 `.claude/` source files (REQ-TMD-006 violation)
- Using `git add .` or `git add -A` (D.2 violation, PRESERVE list contamination risk)
- Adding new `.go` file (scope envelope violation; only test registry edit allowed)
- Adding new `.md` mirror beyond the 4 listed (scope envelope violation; defer to Sprint 8)
- Modifying `internal/template/lateBranchMirroredPaths` slice (out-of-scope per spec.md §B.2; spec-assembly.md already cleared by TMC-001)
- Removing existing entries from `workflowOptMirroredPaths` slice (out-of-scope; only addition allowed)
- Bundling unrelated cleanup into the same commit (scope discipline violation; one M1 commit covers exactly the 5 files in §A.2)

## §H. Cross-References

- spec.md §F — REQ-TMD-001..011 canonical SSOT
- acceptance.md §D — AC-TMD-001..005 binary PASS/FAIL matrix
- progress.md §E — 4-phase Lifecycle Status + Audit-Ready Signal
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S definition
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal form
- `.claude/rules/moai/workflow/mx-tag-protocol.md` §a — Mx Step C SKIP/EVALUATE rules
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical SSOT
- `internal/template/rule_template_mirror_test.go` — test registry SSOT (modified by REQ-TMD-005)
- TMC-001 precedent `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001/spec.md` — Tier S 1-pass success pattern
