---
id: SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001
title: "Plan — Anthropic Best-Practice Audit Tier 3 (F3+F9+F13)"
version: "0.2.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules + internal/spec"
lifecycle: spec-anchored
tags: "anthropic-best-practice, audit-tier-3, plan, tier-m"
tier: M
---

# Plan — SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001

Tier M standard plan. Sections A-E mandatory per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability.

---

## §A. Context — 위치 + 분기 + SPEC 산출물 경로

### §A.1 작업 위치 + 분기

- **Project root**: `/Users/goos/MoAI/moai-adk-go` (absolute)
- **현재 branch**: `main` (BODP default — late-branch + Hybrid Trunk per `.moai/docs/git-workflow-doctrine.md`, CLAUDE.local.md §23.7)
- **Run-phase branch convention**: `main` 직진 push (Tier M 모두 main 허용 per §23.7 Hybrid Trunk 1-person OSS)
- **Pre-spawn fetch obligation**: REQUIRED for every manager-develop / manager-docs Agent() spawn (CLAUDE.local.md §23.8 + agent-common-protocol.md §Pre-Spawn Sync Check)

### §A.2 SPEC 산출물 경로

- `spec.md` (this SPEC, plan-phase complete after this commit)
- `plan.md` (this file)
- `acceptance.md` (Tier M mandatory)
- `progress.md` (Tier M mandatory + run/sync/Mx audit-ready signal slots)

### §A.3 plan-auditor verdict 의무

Tier M PASS threshold: **0.80**. Skip-eligible: **0.90**. plan-auditor iter-1 후 결정.

### §A.4 PRESERVE 대상 (5 categories, run-phase forbidden modifications)

**P1 — Sibling SPEC active scope (HARD)**:
- `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/*` (active run-phase)
- `.claude/rules/moai/core/agent-common-protocol.md` (COORD-001 EXTEND 대상)
- `internal/governance/*` (COORD-001 new package)

**P2 — Recently closed SPEC scope (HARD)**:
- `.claude/agents/core/manager-spec.md` (ARR-001 sealed)
- `.claude/agents/core/manager-develop.md` (ARR-001 sealed)
- `.claude/agents/core/manager-docs.md` (ARR-001 sealed)
- `internal/template/templates/.claude/agents/core/manager-*.md` (ARR-001 template mirrors, sealed)

**P3 — runtime-managed files (HARD, never touched)**:
- `.moai/harness/usage-log.jsonl` (runtime appended)
- `.moai/harness/learning-history/` (runtime managed)
- `.moai/state/` (runtime managed)
- `.moai/cache/` (runtime managed)

**P4 — Untracked / parallel-session artifacts (HARD)**:
- `.moai/research/anthropic-best-practices-2026-05-24.md` (read-only reference, no modification)
- `.moai/research/v3.0-redesign-2026-05-23.md` (unrelated)
- `i18n-validator` (untracked unrelated artifact)
- `.moai/config/sections/*.yaml` modified files (untouched per §23.7)

**P5 — Out-of-scope module conventions**:
- `pkg/` (외부 export API — 별도 SPEC)
- `cmd/moai/` (entry point — root suffices)
- `internal/governance/CLAUDE.md` (COORD-001 ownership)

### §A.5 EXTEND 대상 (5 categories, run-phase create or extend)

**E1 — 신규 module CLAUDE.md (5 CREATE)**:
- `internal/cli/CLAUDE.md` (NEW)
- `internal/template/CLAUDE.md` (NEW)
- `internal/spec/CLAUDE.md` (NEW)
- `internal/hook/CLAUDE.md` (NEW)
- `internal/config/CLAUDE.md` (NEW)

**E2 — internal/spec/ lint extension**:
- `internal/spec/lint.go` (EXTEND — new `OwnershipTransitionRule` struct + Check method, ~+120-180 LOC)
- `internal/spec/lint_test.go` (EXTEND — table-driven test for the new rule, ~+180-240 LOC)

**E3 — schema doc cross-reference**:
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (EXTEND — short cross-ref subsection, ~+20-30 LOC)

**E4 — template mirror (Template-First Rule [HARD])**:
- `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` (MIRROR — byte-identical to E3)

**E5 — progress.md audit-ready signal slots**:
- `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/progress.md` (EXTEND — §E.2 run / §E.3 audit-ready / §E.4 sync / §E.5 Mx slots populated phase-by-phase)

Total file delta: **10 files**, **~700-1000 LOC** — Tier M envelope.

---

## §B. Known Issues (8 categories from manager-develop-prompt-template.md §1 Section B)

### B1 — Cross-platform Build Tags
- **Risk**: NONE. 본 SPEC scope는 markdown (5 module CLAUDE.md) + `internal/spec/lint.go` extension. Lint extension은 `internal/spec/` 내 기존 패턴 (`FrontmatterSchemaRule`) mirror — no syscall, no platform-specific code.
- **Verification**: `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (run-phase pre-flight).

### B2 — Cross-SPEC 정책 충돌 사전 스캔
- **Risk**: HIGH. COORD-001 (active) + ARR-001 (closed). 본 SPEC §B.3.1/§B.3.2 disjoint scope 명시.
- **Verification (run-phase pre-flight)**:
  ```bash
  grep -r "OwnershipTransitionRule\|OwnershipTransitionInvalid" internal/spec/ 2>/dev/null
  # → 0 matches expected (new rule not yet implemented)
  grep -rn "subdirectory.*CLAUDE.md\|module-level CLAUDE.md" .claude/rules/ 2>/dev/null | head -5
  # → no conflicting policy expected
  ```

### B3 — C-HRA-008 / Subagent Boundary Discipline
- **Risk**: NONE (본 SPEC은 `internal/spec/` 만 다룬다 — harness/hook 도메인 아님)
- **Verification**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/spec/ | grep -v "_test.go" | grep -v "// "` → 0 matches expected.

### B4 — Frontmatter Canonical Schema
- **Risk**: LOW. 본 SPEC artifact 4개 모두 12-canonical-field 적용 (spec.md frontmatter `id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags` + `tier: M` optional).
- **Verification**: `moai spec lint .moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/` exit 0 expected (자기 자신은 valid).

### B5 — CI 3-tier 인지
- **Risk**: MEDIUM. `OwnershipTransitionRule` 신규 도입 시 sibling SPEC들이 historical commit format mismatch로 finding 발생 가능 → 신규 baseline 형성. 대응: REQ-AAT-009 default-subset 적용 + observation period.
- **Verification**: run-phase에서 `go test ./internal/spec/...` 새 rule 테스트 통과 + 기존 spec-lint baseline regression 0건 (NEW vs pre-existing 구분).

### B6 — spec-lint Heading 규약
- **Risk**: LOW. 본 SPEC spec.md §B.3 sub-sections (§B.3.1..§B.3.6) all H4 (`####`) following sibling SPECs pattern. ARR-001 / TMC-001 / PROPOSAL-GEN-001 precedent (`### §X.Y` h3 sub-headings under `## §B Scope`).
- **Verification**: `moai spec lint` no `MissingExclusions` finding.

### B7 — observer.go / capture path resolution
- **Risk**: NONE (본 SPEC은 hook 영역 아님). `internal/hook/CLAUDE.md` 작성 시 B7 issue를 module convention 항목으로 surfacing할 것 (REQ-AAT-003).

### B8 — Working Tree Hygiene
- **Risk**: MEDIUM. plan-phase는 본 SPEC 디렉토리만 추가; run-phase는 §A.5 EXTEND 10 files만. P3 runtime-managed files 절대 변경 금지.
- **Verification (pre-commit)**: `git status --short` 결과가 §A.5 list 범위 내인지 확인.

### B9 — Git Commit + Push 자체 수행 (Hybrid Trunk 1-person OSS)
- **Plan-phase**: orchestrator commit (manager-spec은 NO COMMIT 명시 instructions per parent prompt).
- **Run-phase**: manager-develop이 main 직진 commit + push. Conventional Commits format `feat(SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001): M{N} <subject>`.
- **Sync-phase**: manager-docs가 CHANGELOG 추가 + frontmatter status `in-progress → implemented` 4 artifacts + main 직진 push.

### B10 — Untouched Paths PRESERVE
- **Strict adherence**: §A.4 P1-P5 lists. 특히 COORD-001 / ARR-001 / runtime / research / config dirty files 절대 손대지 말 것.

### B11 — AskUserQuestion 금지 (Subagent Boundary)
- manager-spec / manager-develop / manager-docs 모두 subagent — AskUserQuestion 금지. Blocker 시 structured blocker report 반환.

### B12 — Sync-phase CHANGELOG emission discipline (manager-docs only)
- Sync-phase에서 CHANGELOG entry 작성 전 `grep -c 'SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001' CHANGELOG.md` 가 0인지 확인.
- 모든 implementation file Read 후 entry 작성 (plan.md description alone 의존 금지).

---

## §C. Pre-flight Check List (run-phase 착수 전 의무 검증)

run-phase 위임 prompt에 다음을 의무 포함:

```bash
# 1. 현재 branch + baseline
git branch --show-current   # → main
git rev-parse HEAD          # → record baseline SHA

# 2. Pre-spawn fetch (race detection per §23.8)
git fetch origin main 2>&1
git rev-list --count --left-right origin/main...HEAD
# 0 0 or 0 N expected; N 0 or N M → STOP + blocker report

# 3. Cross-platform build baseline
go build ./...                              # → exit 0 expected
GOOS=windows GOARCH=amd64 go build ./...    # → exit 0 expected

# 4. Existing lint baseline (NEW vs pre-existing 구분)
golangci-lint run --timeout=2m 2>&1 | tail -10

# 5. Sibling SPEC scope verification
ls .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/  # active — DO NOT TOUCH
ls .claude/agents/core/manager-spec.md .claude/agents/core/manager-develop.md .claude/agents/core/manager-docs.md
# 위 3 manager-*.md PRESERVE only (ARR-001 sealed)

# 6. PRESERVE 대상 확인
ls -1 .claude/rules/moai/core/agent-common-protocol.md \
      internal/governance/ 2>/dev/null
# COORD-001 active scope — DO NOT MODIFY

# 7. 영향 패키지 retired/superseded SPEC 확인
grep -r "Retired\|TestHarnessRetirement\|superseded" internal/spec/ || echo "no conflicts"
```

---

## §D. Constraints (DO NOT VIOLATE)

[HARD] Run-phase 위임 시 explicit list:

### D1 — File-level forbiddens (HARD)
- `.claude/rules/moai/core/agent-common-protocol.md` — COORD-001 active
- `internal/governance/*` — COORD-001 new package
- `.claude/agents/core/manager-spec.md` / `manager-develop.md` / `manager-docs.md` — ARR-001 sealed
- `internal/template/templates/.claude/agents/core/manager-*.md` — ARR-001 mirrors sealed
- `.moai/research/*.md` — read-only references
- `.moai/harness/*` / `.moai/state/*` / `.moai/cache/*` — runtime managed
- `i18n-validator` (untracked unrelated)
- `.moai/config/sections/*.yaml` modified status — leave untouched

### D2 — Command-level forbiddens (HARD)
- `--no-verify` (pre-commit hook bypass)
- `--amend` (history rewrite — Hybrid Trunk Tier S/M policy)
- `git push --force` / `--force-with-lease` to main
- `git rebase -i` (interactive disallowed in sandbox)
- `git reset --hard` (sandbox auto-blocks per CLAUDE.local.md §23.5; use `--keep` if reset needed)

### D3 — Required command patterns
- Conventional Commits: `feat(SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001): M{N} <subject>` (M1/M2/M3)
- Trailer line: `🗿 MoAI <email@mo.ai.kr>`
- Coauthor tag (manager-develop): `Co-Authored-By: Claude <noreply@anthropic.com>`

### D4 — Run-phase max-3-commits cap (§B.3.6)
- M1 commit = 5 subdirectory CLAUDE.md CREATE
- M2 commit = `OwnershipTransitionRule` lint extension + tests
- M3 commit = schema doc cross-ref + template mirror + progress.md run-evidence
- 추가 commit 시 blocker report → manager-spec re-delegate (D-NEW-1 inline-fix pattern)

### D5 — Binary verifications (REQ-AAT cross-ref)
- `OwnershipTransitionRule` lint MUST be observation-only (REQ-AAT-011) — NO file mutation, NO auto-fix
- Schema doc + template mirror byte-identical (REQ-AAT-013)
- Subdirectory CLAUDE.md 60-200 LOC each (REQ-AAT-004)
- Root vs subdirectory diff <50% overlap (REQ-AAT-005)

---

## §E. Self-Verification Deliverables (run-phase 완료 시 manager-develop 보고 의무)

### E1 — AC Binary PASS/FAIL Matrix

manager-develop은 acceptance.md §D AC 10건에 대해 자체 검증한 PASS/FAIL matrix 보고:

| AC | Status | Verification Command | Expected Output |
|----|--------|---------------------|-----------------|
| AC-AAT-001 | PASS/FAIL | `ls internal/{cli,template,spec,hook,config}/CLAUDE.md` | 5 files present |
| AC-AAT-002 | PASS/FAIL | `wc -l internal/{cli,template,spec,hook,config}/CLAUDE.md` | each 60-200 lines |
| AC-AAT-003 | PASS/FAIL | (custom diff script) | <50% overlap each pair |
| AC-AAT-004 | PASS/FAIL | `grep -n "type OwnershipTransitionRule" internal/spec/lint.go` | 1 match expected |
| AC-AAT-005 | PASS/FAIL | `go test -run TestOwnershipTransitionRule_Pass -v ./internal/spec/` | 7 PASS scenarios |
| AC-AAT-006 | PASS/FAIL | `go test -run TestOwnershipTransitionRule_Fail -v ./internal/spec/` | 5 FAIL detections |
| AC-AAT-007 | PASS/FAIL | `go test -run TestOwnershipTransitionRule_UnreachableGit -v ./internal/spec/` | graceful Info |
| AC-AAT-008 | PASS/FAIL | `diff .claude/rules/.../schema.md internal/template/templates/.claude/rules/.../schema.md` | empty diff |
| AC-AAT-009 | PASS/FAIL | 4-platform cross-build + go vet + golangci-lint | all exit 0 |
| AC-AAT-010 | PASS/FAIL | `git diff --name-only HEAD~3..HEAD` | only §A.5 EXTEND files |

### E2 — Cross-Platform Build 결과
```
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
$ GOOS=linux GOARCH=amd64 go build ./...    → exit 0
$ GOOS=darwin GOARCH=arm64 go build ./...   → exit 0
```

### E3 — Coverage 측정 (≥85% threshold for internal/spec/)
```
$ go test -cover ./internal/spec/...
# coverage report; OwnershipTransitionRule 신규 코드 ≥85% expected
```

### E4 — Subagent Boundary Grep
```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/spec/ | grep -v "_test.go" | grep -v "// "
(no output expected)
```

### E5 — Lint Status (NEW vs baseline 구분)
```
$ golangci-lint run --timeout=2m ./internal/spec/...
$ moai spec lint .moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/
$ moai spec lint .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/  # MUST still pass — verify OwnershipTransitionRule does NOT false-flag closed SPECs
# NEW issues 발견 시 explicit report; pre-existing baseline 별도 mark
```

### E6 — Branch HEAD + Push 상태
- M1/M2/M3 commit SHA 리스트
- `git push origin main` 결과 (Hybrid Trunk Tier M)

### E7 — Disjoint Scope Verification (AC-AAT-010)
```
$ git diff --name-only HEAD~3..HEAD | sort | uniq
# Expected: 10 files only, all listed in §A.5 EXTEND.
# MUST NOT contain:
#   - .claude/rules/moai/core/agent-common-protocol.md (COORD-001)
#   - internal/governance/* (COORD-001)
#   - .claude/agents/core/manager-{spec,develop,docs}.md (ARR-001)
#   - .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/* (active sibling)
```

### E8 — Blocker Report (있을 시)
- §B.3.6 max-3-commits cap 위반 가능성 발견 시 (e.g., subdirectory CLAUDE.md가 5개 모두 200 LOC 초과 → REQ-AAT-004 위반 risk → 본문 압축 필요): manager-develop 임의 판단 금지, structured blocker report 반환 후 orchestrator → manager-spec re-delegate 패턴 (D-NEW-1 inline-fix).

---

## §3. Trade-off Analysis (Tier M Section C — 3 sub-axes)

본 절은 plan-auditor §C 평가 항목 (alternative consideration) 충족.

### §3.1 F9 subdirectory CLAUDE.md 위치 분기 (3 options)

| Option | Description | Pros | Cons | 채택 |
|--------|-------------|------|------|------|
| A — 5 module (internal/cli + template + spec + hook + config) | 본 SPEC 채택 | Anthropic 권장 정석 / agent automatic loading / 도메인 분리 명확 / scope manageable | 5 files create overhead | ✓ |
| B — 모든 internal/<pkg>/ (~15 module) | full coverage | comprehensive | scope drift / Tier L envelope 진입 / 일부 module은 trivial | ✗ |
| C — root CLAUDE.md에 module-specific sections 추가 | minimal create | only 1 file edit | bloated root (Anthropic anti-pattern) / loading inefficiency | ✗ |

### §3.2 F13 verification mechanism 분기 (3 options)

| Option | Description | Pros | Cons | 채택 |
|--------|-------------|------|------|------|
| A — spec-lint rule extension | 본 SPEC 채택 (B1 in §A.4) | shift-left detection / 기존 인프라 활용 / 관찰 (no mutation) | git log dependency (가벼운) | ✓ |
| B — PostToolUse hook enforcement | execution-time block | true prevention | ARR-001 REQ-009 deferred / agent attribution 복잡 | ✗ (deferred) |
| C — GitHub Actions CI workflow | post-push check | comprehensive history scan | too-late detection / push 후 발견 | ✗ |

### §3.3 module CLAUDE.md depth-of-prose 분기 (3 options)

| Option | Description | Pros | Cons | 채택 |
|--------|-------------|------|------|------|
| A — 60-90 LOC concise per module | 본 SPEC 채택 | Anthropic lean principle / token efficient / surveyable | 일부 module은 더 풍부한 contract description 필요 가능 | ✓ |
| B — 150-200 LOC comprehensive per module | exhaustive | full context | bloat risk / maintenance overhead | ✗ |
| C — module별 가변 (cli 150 / template 80 / spec 120 / hook 80 / config 60) | adaptive | scope-appropriate | inconsistent template / plan-auditor 복잡 | ✗ |

---

## §4. Milestone Decomposition (M1-M3, max-3-commits cap)

### M1 — Subdirectory CLAUDE.md × 5 (REQ-AAT-001..006)

**Scope**: 5 NEW files (`internal/cli/CLAUDE.md` + `internal/template/CLAUDE.md` + `internal/spec/CLAUDE.md` + `internal/hook/CLAUDE.md` + `internal/config/CLAUDE.md`).

**Steps**:
1. Read root CLAUDE.md to identify what NOT to duplicate (REQ-AAT-005)
2. For each module:
   - Read 5-10 representative files in the module to extract conventions
   - Draft CLAUDE.md (60-90 LOC) with 4 required sections (REQ-AAT-003): purpose / key files / module-specific conventions / cross-references
   - Verify <50% diff overlap with root CLAUDE.md (REQ-AAT-005)
3. Commit: `feat(SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001): M1 subdirectory CLAUDE.md × 5 (F9)`
4. Push to main.

**Delegation strategy**: manager-develop with Section A-E prompt. Single milestone, single commit. No sub-delegation.

### M2 — OwnershipTransitionRule lint extension (REQ-AAT-007..012)

**Scope**: 2 files modified (`internal/spec/lint.go` + `internal/spec/lint_test.go`).

**Steps**:
1. Read existing `FrontmatterSchemaRule` in `internal/spec/lint.go` as pattern reference
2. Read `spec-frontmatter-schema.md § Status Transition Ownership Matrix` for the 7-row truth table
3. Implement TDD per quality.yaml `development_mode: tdd`:
   - RED: 7 PASS + 5 FAIL + 1 unreachable-git table-driven test cases (lint_test.go)
   - GREEN: `OwnershipTransitionRule` struct + Check() method (lint.go)
   - REFACTOR: extract helper for `git log --follow` parsing
4. Verify: `go test ./internal/spec/...`, coverage ≥85%, `golangci-lint run` 0 issues
5. Verify ARR-001 SPEC itself does NOT false-positive (REQ-AAT-009 default subset behavior)
6. Commit: `feat(SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001): M2 OwnershipTransitionRule lint (F13)`
7. Push to main.

**Delegation strategy**: manager-develop with full Section A-E prompt including B5 (CI 3-tier) + B6 (heading) + B12 (sync discipline pre-warning).

### M3 — Schema doc cross-ref + template mirror + progress.md (REQ-AAT-013)

**Scope**: 3 files modified (`.claude/rules/moai/development/spec-frontmatter-schema.md` + template mirror + progress.md run-evidence).

**Steps**:
1. Append `## OwnershipTransitionRule Cross-Reference` subsection (~20-30 LOC) to schema doc — file/line/finding-code documentation pointing to `internal/spec/lint.go`
2. Copy byte-identical content to template mirror (`internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md`)
3. Verify `diff` between two files yields empty output (REQ-AAT-013 byte-identical)
4. Populate `progress.md §E.2 Run-phase Evidence` with M1/M2/M3 commit SHAs + verification command outputs
5. Populate `progress.md §E.3 Run-phase Audit-Ready Signal` block
6. Commit: `feat(SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001): M3 schema cross-ref + template mirror + run-evidence`
7. Push to main.

**Delegation strategy**: manager-develop. Final run-phase commit.

### Optional follow-up — Sync-phase

After M3 push, orchestrator runs pre-spawn fetch then delegates to manager-docs for `/moai sync SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001`:
- CHANGELOG.md `[Unreleased]` entry (manager-docs scope per ARR-001 ownership matrix)
- 4 SPEC artifact frontmatter `status: in-progress → implemented` transition
- `progress.md §E.4 Sync-phase Audit-Ready Signal`

manager-docs MUST NOT modify spec.md / plan.md / acceptance.md body (ARR-001 forbidden ownership crossing). Frontmatter `status:` + `updated:` updates only.

---

## §5. Verification strategy (run-phase Self-Verification deliverables 대응)

| Verification | Phase | Owner | Outcome |
|--------------|-------|-------|---------|
| AC PASS/FAIL matrix (10 items) | run-phase final | manager-develop | E1 |
| Cross-platform build (4 platforms) | M1 + M2 + M3 each | manager-develop | E2 |
| Coverage ≥85% for internal/spec/ | M2 final | manager-develop | E3 |
| Subagent boundary grep | run-phase final | manager-develop | E4 |
| Lint (golangci-lint + moai spec lint) | run-phase final | manager-develop | E5 |
| Branch HEAD + push | each M commit | manager-develop | E6 |
| Disjoint scope verification | run-phase final | manager-develop | E7 (AC-AAT-010) |
| Blocker report if any | as-needed | manager-develop | E8 |

---

## §6. Risk Register (also see spec.md §F)

추가 plan-level risk:

| Risk | Mitigation |
|------|-----------|
| M2 `OwnershipTransitionRule` git log parsing이 부분 false-positive 빈발 | REQ-AAT-009 default-subset (4 most-common transitions) + warning severity (not error) — observation period 후 tighten |
| M1 subdirectory CLAUDE.md prose가 root와 50%+ overlap | M1 작성 중 incremental `diff` 자가 점검 — `wc -l && diff` 매 module 작성 후 |
| M3 template mirror drift | `diff` 자가 검증 명시 (D5) + M3 commit message에 verification 출력 포함 |
| Sibling SPEC (COORD-001) 동시 main push race | pre-spawn fetch `git rev-list --count --left-right origin/main...HEAD` 매 M commit 직전 + `0 0` or `0 N` 확인 |
| `OwnershipTransitionRule`이 ARR-001 SPEC 자체에 false-flag 발생 (자기 자신이 첫 번째 사례) | M2 verification step 5 명시: ARR-001 SPEC 통과 확인 |

---

## §7. Open Questions (plan-auditor § Open Questions 평가 항목)

본 plan은 아래 OQ를 명시한다 — orchestrator/user 결정 사항:

**OQ1 — F3 unused skills reconnect 처리 위치**: 본 SPEC §B.2 항목 #5에서 Tier 4 backlog로 분리. 대안: 본 SPEC scope에 포함 (3 trivial frontmatter edits). 결정: §B.2 적용 (Tier M envelope 유지 위해).

**OQ2 — `OwnershipTransitionRule` lint config opt-out 메커니즘**: REQ-AAT-009은 default-enabled-for-all-7-transitions로 결정. 대안: 시작은 4-transition default, opt-in으로 나머지 3. 결정: default-enabled-all-7 + lint config 통한 opt-out (운영 중 false-positive 빈발 시).

**OQ3 — Subdirectory CLAUDE.md 추가 module (governance / pkg / cmd) 확장 시점**: §B.2 #6 + §B.3.4 deferred. 대안: 본 SPEC 즉시 8 module. 결정: 5 module 우선 — Tier M envelope.

---

## §8. Cross-references (related canonical artifacts)

- `.moai/research/anthropic-best-practices-2026-05-24.md` §3 F3+F9+F13 (audit findings origin)
- `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` (Tier 2 antecedent, ownership matrix introduction)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (SSOT being verified by REQ-AAT-007..012)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier M classification)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section A-E (Tier M REQUIRED template)
- `CLAUDE.local.md` §23.7 Hybrid Trunk 1-person OSS + §23.8 Multi-Session Race Mitigation
- `.claude/agents/core/manager-spec.md` § SPEC Artifact Ownership (ARR-001 boundaries — read-only)
- `internal/spec/lint.go` `FrontmatterSchemaRule` (pattern reference for M2)
