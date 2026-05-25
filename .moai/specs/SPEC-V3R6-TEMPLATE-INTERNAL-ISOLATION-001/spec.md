---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Template Internal-Content Isolation — Permanent Removal of moai-adk Dev-Internal Tokens from internal/template/templates/"
version: "0.1.4"
status: completed
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, internal-content, leak-cleanup, governance"
tier: M
depends_on: []
related_specs: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
---

# SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 — template 내 moai-adk 내부 개발 내용 영구 격리

## §A. 배경 (Background)

### A.1 문제 진술

`internal/template/templates/` 디렉토리는 moai-adk-go CLI가 전 세계 사용자 프로젝트로 deploy하는 템플릿 자산의 단일 소스다. 본 SPEC은 이 디렉토리에 누출되어서는 안 되는 콘텐츠를 **moai-adk dev-internal-content token** (canonical term; 약식: "internal-content" — frontmatter tags 및 약식 prose에서만 사용)으로 정의한다. 누출 금지 대상:

- moai-adk 개발 과정에서 등장한 SPEC ID 문자열 (예: `SPEC-V3R6-AGENT-TEAM-REBUILD-001`, `SPEC-V3R6-WORKFLOW-OPT-001`)
- REQ token 문자열 (예: `REQ-ATR-009`, `REQ-ATR-014`, `REQ-WO-007`)
- moai-adk 메인테이너용 audit 인용 (예: `Audit 3 Findings A1-A6`, `Anthropic 2026 Alignment`)
- moai-adk 개발 history 날짜 (예: `archive-2026-05-25`, `2026-05-25` 자체 인용)
- moai-adk 개발 commit sha (예: `b957a4d04`, `40dc43f5b`)

이러한 토큰은 사용자 프로젝트에서 의미가 없으며, 사용자에게 "moai-adk가 어떻게 만들어졌는가" 의 메타 정보를 강제 노출한다. 사용자는 자신의 프로젝트 도메인에 집중해야 한다.

### A.2 기존 정책 격차

기존 CLAUDE.local.md 정책들은 본 격리 의무를 명시적으로 다루지 않았다:

- **§2 Template-First Rule**: "템플릿 우선 갱신" 의무는 명시하나, 템플릿에 들어가서는 안 되는 내용은 다루지 않음
- **§15 템플릿 언어 중립성**: 16개 언어 동등 취급 의무는 명시하나, 언어 편향이 아닌 "개발 컨텍스트 누출" 차원은 implicit
- **§21 Dev-Only Commands Isolation**: 97/98/99 prefix 커맨드 파일 격리는 명시하나, 일반 템플릿 본문 내 토큰 격리는 다루지 않음
- **§24 Harness Namespace 분리 정책**: namespace 차원의 격리는 명시하나, content 차원의 격리는 다루지 않음

결과적으로 SPEC-V3R6-AGENT-TEAM-REBUILD-001 (Tier L, 2026-05-25 완료) 진행 중 M5/M7 milestones에서 moai-adk dev-internal-content token이 templates 디렉토리에 유입됐고, 2회의 predecessor partial cleanup (pass 1 + pass 2)을 거친 본 SPEC plan-phase 시점의 ground-truth 잔존은 **35 files** (`grep -rln ... | wc -l` 로 검증, §A.3 ground truth 표 참조).

### A.3 Ground Truth (verified at plan-phase start)

| 항목 | 값 | 검증 명령 |
|------|-----|----------|
| HEAD commit | `40dc43f5b` (chore pass 2, 2 files) | `git log --oneline -1` |
| origin/main race | `0 0` (synced) | `git rev-list --count --left-right origin/main...HEAD` |
| 잔여 leak files | **35** | `grep -rln 'SPEC-V3R6-\|REQ-ATR-\|Audit 3\|Finding A[1-6]\|archive-2026-05-25' internal/template/templates/ \| wc -l` |
| CLAUDE.local.md §25 존재 여부 | 부재 | `grep -c 'Template Internal-Content Isolation' CLAUDE.local.md` → `0` |
| 본 SPEC 디렉토리 | 신규 생성 | `ls .moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/` |
| 메모리 feedback 파일 | 존재 (8996 bytes, 2026-05-25 18:48) | `ls -la /Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_template_internal_content_isolation.md` |
| Predecessor partial cleanup pass 1 | `20a66df85` (9 files, -80 lines) | `git show --stat 20a66df85` |
| Predecessor partial cleanup pass 2 | `40dc43f5b` (2 files) | `git show --stat 40dc43f5b` |
| Plan-phase commit anchor (canonical attribution base for AC verification) | `b7d1528c8` | `git log --grep='SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001' --format=%H b7d1528c8..HEAD` |

### A.4 35개 잔여 leak files 분포 (ground-truth verified)

| 경로 prefix | 파일 수 | 대표 파일 |
|------------|--------|----------|
| `.claude/agents/core/` | 4 | manager-spec.md, manager-develop.md, manager-docs.md, manager-git.md |
| `.claude/agents/meta/` | 1 | plan-auditor.md |
| `.claude/rules/moai/core/` | 3 | agent-common-protocol.md, settings-management.md, askuser-protocol.md |
| `.claude/rules/moai/development/` | 2 | agent-patterns.md, manager-develop-prompt-template.md |
| `.claude/rules/moai/workflow/` | 5 | orchestration-mode-selection.md, worktree-integration.md, archived-agent-rejection.md, spec-workflow.md, session-handoff.md |
| `.claude/skills/moai-foundation-core/` | 2 | SKILL.md, modules/spec-ears-format.md |
| `.claude/skills/moai/workflows/` | 7 | fix.md, sync.md, harness.md, run.md, plan.md, plan/spec-assembly.md, sync/delivery.md |
| `.claude/skills/moai-workflow-*/` 외 SKILL.md | 6 | moai-workflow-spec, moai-workflow-worktree, moai-foundation-thinking, moai-workflow-ci-loop, moai-workflow-design, moai-workflow-testing |
| `.claude/output-styles/moai/moai.md` | 1 | output-styles/moai/moai.md |
| `.claude/hooks/moai/team-ac-verify.sh` | 1 | team-ac-verify.sh |
| `.moai/config/sections/system.yaml.tmpl` | 1 | system.yaml.tmpl |
| `.moai/docs/generic-patterns-guide.md` | 1 | generic-patterns-guide.md |
| `.gitignore` | 1 | .gitignore |
| **합계** | **35** | |

(M1 단계 W1 audit에서 `grep -rln ... \| wc -l`로 재검증; 변동 감지 시 plan.md §F W1 표 갱신)

### A.5 5 Deliverables Summary

1. **(a) 35 잔여 leak files cleanup** — moai-adk dev-internal-content token → generic patterns 변환
2. **(b) CLAUDE.local.md §25 NEW HARD rule** — Template Internal-Content Isolation 영구 anchor
3. **(c) Go lint test `TestTemplateNoInternalContentLeak`** — `internal/template/internal_content_leak_test.go` 신규
4. **(d) CI workflow integration** — `.github/workflows/` 내 본 lint test gating
5. **(e) Maintainer-only file template-removal review** — 97/98/99 등 dev-only 파일이 templates에 잘못 포함되어 있지 않은지 audit

---

## §B. 요구사항 (REQs in GEARS Format)

### B.1 Cleanup 의무 (Deliverable a)

**REQ-TII-001** (Ubiquitous): The `internal/template/templates/` directory **shall** be free of moai-adk dev-internal-content tokens including, but not limited to: actual SPEC ID literals matching `SPEC-V3R6-`, REQ token literals matching `REQ-ATR-`, audit citations matching `Audit 3` or `Finding A[1-6]`, and archive date references matching `archive-2026-05-25`.

**REQ-TII-002** (Event-driven): **When** a template author identifies an existing moai-adk dev-internal-content token within `internal/template/templates/`, the author **shall** apply a generic substitution pattern (예: actual SPEC ID → "선행 SPEC" or "predecessor SPEC"; REQ token → 일반 prose 설명) that preserves the pedagogical intent without naming the specific moai-adk dev-internal-content token.

**REQ-TII-003** (Where-capability): **Where** a template file is a generated mirror of a source-of-truth file located elsewhere in the repository (e.g., `internal/template/templates/.claude/rules/moai/NOTICE.md` mirrors `.claude/rules/moai/NOTICE.md`), the template mirror **shall** be re-generated from the cleaned source-of-truth via `make build`, not edited independently.

### B.2 영구 정책 anchor (Deliverable b)

**REQ-TII-004** (Ubiquitous): CLAUDE.local.md **shall** contain a `§25` section titled `Template Internal-Content Isolation` that defines (a) the allowed content classes, (b) the forbidden content classes, (c) a 5-item pre-commit self-check checklist, and (d) at least 3 worked anti-pattern examples extracted from the predecessor partial-cleanup history.

**REQ-TII-005** (Event-driven): **When** a contributor prepares a commit that touches files under `internal/template/templates/`, the contributor **shall** execute the §25 pre-commit self-check checklist before invoking `git commit`.

### B.3 Lint 자동화 (Deliverable c)

**REQ-TII-006** (Ubiquitous): A Go test `TestTemplateNoInternalContentLeak` **shall** exist at `internal/template/internal_content_leak_test.go` that walks `internal/template/templates/` and fails when any file matches the canonical leak pattern set defined in REQ-TII-001.

**REQ-TII-007** (Where-capability): **Where** a template file legitimately requires a literal token that would otherwise match the leak pattern set (예: NOTICE.md attribution citations referencing third-party Apache 2.0 import dates), the file **shall** be enumerated in a structured allowlist (proposed: in-test Go literal slice with documented rationale per entry) and the test **shall** consult the allowlist before raising failure.

**REQ-TII-008** (Event-driven): **When** the leak test is executed on a HEAD with the 35 leak files cleaned to zero, the test **shall** return PASS with exit code 0; **when** a synthetic leak (one of the canonical patterns) is reintroduced into any template file, the test **shall** return FAIL with a deterministic, file-path-specific error message.

### B.4 CI gating (Deliverable d)

**REQ-TII-009** (State-driven): **While** a pull request is open against `main`, the GitHub Actions CI pipeline **shall** execute `TestTemplateNoInternalContentLeak` as part of the standard test job, and merge **shall** be blocked when the test fails.

**REQ-TII-010** (Where-capability): **Where** the GitHub Actions workflow file `.github/workflows/test.yml` (or equivalent) does not already invoke `go test ./internal/template/...`, the workflow **shall** be extended to include this invocation, and the workflow file **shall** carry a docstring comment cross-referencing `/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_template_internal_content_isolation.md` as the policy rationale.

### B.5 Maintainer-only audit (Deliverable e)

**REQ-TII-011** (Ubiquitous): The audit pass **shall** verify that no maintainer-only file class (97/98/99 prefix dev-only commands per CLAUDE.local.md §21, `settings.local.json` per §2, dev-only state files in `.moai/state/last-cc-version.json`, dev-only reports in `.moai/research/cc-update-*.md`) exists within `internal/template/templates/` either directly or via mirror.

**REQ-TII-012** (Event-driven): **When** the audit pass discovers a maintainer-only file class inadvertently present in `internal/template/templates/`, the file **shall** be removed via `git rm` and an entry **shall** be added to `.gitignore` (or equivalent guard) preventing future re-inclusion.

### B.6 Anti-pattern enforcement (cross-cutting)

**REQ-TII-013** (Unwanted Behavior): The template author **shall not** introduce new template content that names a specific moai-adk SPEC ID, REQ token, audit citation, archive date, or commit sha, even when the content describes a pattern derived from that historical artifact; instead the author **shall** describe the pattern in generic prose terms.

---

## §C. 제약사항 (Constraints)

### C.1 HARD constraints (must-honor)

- **[HARD-1]** Mirror parity preservation: `make build` 실행 시 `internal/template/embedded.go`가 깨지지 않고 재생성되어야 하며, deploy된 사용자 프로젝트의 `.claude/`, `.moai/` 트리가 cleaned templates와 byte-for-byte 일치
- **[HARD-2]** Predecessor cleanup pattern 일관성: pass 1 (`20a66df85`)과 pass 2 (`40dc43f5b`)가 사용한 변환 dictionary와 일관된 substitution을 적용; deviation 시 design.md §C에 정당화 명시
- **[HARD-3]** No behavioral regression: cleanup이 템플릿 본문의 pedagogical intent (사용자가 학습할 내용)를 훼손하지 않음 — generic substitute가 충분히 informative해야 함
- **[HARD-4]** Allowlist 최소화: lint test의 allowlist는 명시적으로 정당화 가능한 항목만 포함; 일반화 가능한 경우 allowlist 추가가 아닌 본문 수정 선택
- **[HARD-5]** 16-language neutrality (§15) 호환: cleanup 변환이 특정 언어 편향을 새로 도입하지 않음

### C.2 SHOULD constraints

- **[SHOULD-1]** Sub-batch 분할: 35 files를 한 atomic commit 대신 3-5개 sub-batch (agents / rules / skills / single-file)로 처리하여 reviewability 확보
- **[SHOULD-2]** §25 HARD rule 작성 시 §15, §21, §24와의 cross-reference 명시
- **[SHOULD-3]** Lint test red-green proof: cleaned HEAD에서 PASS + synthetic leak 주입 시 FAIL을 명시적으로 demo (run-phase 산출물)

---

## §D. Out-of-Scope (Exclusions)

본 SPEC은 다음을 명시적으로 다루지 않는다:

1. **신규 템플릿 자산 추가**: 본 SPEC은 기존 35 leak files cleanup에만 집중; 새로운 agent/skill/rule 파일을 templates에 추가하는 작업은 별도 SPEC
2. **Agent re-architecture**: `manager-spec`, `manager-develop` 등 retained agent의 동작/책임 재정의는 별도 SPEC (predecessor SPEC-V3R6-AGENT-TEAM-REBUILD-001에서 이미 처리됨)
3. **SPEC-V3R6-TEST-REFACTOR-001 scope**: predecessor의 PROCEED-WITH-DEBT 8 test failure 해소는 별개 follow-up SPEC; 본 SPEC의 lint test 추가는 그와 무관
4. **SPEC-V3R6-CATALOG-HASH 통합**: catalog hash regression cleanup 관련 작업과 본 SPEC은 독립; 두 SPEC이 같은 파일을 수정할 경우 시간상 본 SPEC이 후행으로 가정 (predecessor 결과를 입력으로 받음)

---

## §E. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| v0.1.0 | 2026-05-25 | manager-spec | Initial plan-phase 5-artifact set authored (Tier M); 13 REQ-TII / 12 AC-TII / 35 leak files ground-truth verified / 5 deliverables anchored |
| v0.1.1 | 2026-05-25 | manager-spec | iter-1 amendment — 4 SHOULD-FIX 해소: (D-001) AC verification command coverage 회복 (REQ-TII-013 5-class 중 commit sha class 누락 보강 — acceptance.md AC-TII-001 grep 패턴에 40-char/7-8-char-space sha 추가); (D-002) AC-TII-011 §25-scoped awk range (기존 전체 파일 grep은 §17/18/21/23/24 정당 인용 5건 때문에 deterministic FAIL); (D-003) `HEAD~N..HEAD` 하드코딩 → SPEC-scoped attribution range (plan-phase anchor `b7d1528c8` added to §A.3 ground truth table; AC-TII-002/008/010 verification commands rewritten with `git log --grep=<SPEC-ID> b7d1528c8..HEAD`); (D-005) terminology canonicalization — "moai-adk dev-internal-content token" 단일 표현 통일 (§A.1 도입부 + REQ-TII-001/002 + §A.5 deliverable 라인); (D-006) §A.2 산수 phrasing 단순화 ("약 44개 leak 유입" 추정치 제거 → "35 files ground-truth 잔존" verified 수치). REQ count 불변 (13). AC count 불변 (12). frontmatter `version: 0.1.0` → `0.1.1`. |
| v0.1.2 | 2026-05-25 | manager-develop | run-phase M1 — frontmatter status transition `draft` → `in-progress` (Status Transition Ownership Matrix per `.claude/rules/moai/development/spec-frontmatter-schema.md`); 4-artifact frontmatter (spec/plan/acceptance/design/research) 동기화. progress.md initial §A Lifecycle Sync 작성 (plan_commit_sha `b7d1528c8`, iter1_amend_commit_sha `5ff9da7d2`). REQ/AC count 불변 (13 REQ / 12 AC). 본 SPEC body 무변경. |
