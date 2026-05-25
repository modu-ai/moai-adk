---
id: SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001
title: "Harness Namespace 누출 검증 및 정리 (§24 SSOT 자기 일관성)"
version: "0.1.1"
status: completed
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: Low
phase: "v3.0.0 cleanup"
module: ".claude/agents/, .claude/skills/, internal/cli/update_namespace_protect.go"
lifecycle: spec-anchored
tags: "harness, namespace, cleanup, verification, ssot, dev-only"
issue_number: null
depends_on: [SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001]
related_specs: [SPEC-V3R4-HARNESS-NAMESPACE-001, SPEC-V3R6-HARNESS-RENAME-001]
tier: S
---

# SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001

## HISTORY

| 날짜 | 버전 | 변경 내용 | 작성자 |
|------|------|----------|--------|
| 2026-05-25 | 0.1.0 | 초안 작성 — §24 SSOT 자기 일관성 검증 + 로컬 dev 프로젝트 누출 정리 | manager-spec |
| 2026-05-25 | 0.1.1 | L60 spec.md status backfill — `implemented` → `completed`. progress.md 4-phase 완전 종료 (sync `378ef732d` + mx `4485c772f` + L60 atomic backfill `43e65bd03`) 상태였으나 spec.md frontmatter status drift 잔존 (L67 manager-docs scope-creep 패턴). Cross-file 일관성 복원. Body 변경 없음; frontmatter status + version만. | orchestrator |

## 1. 배경 (Why)

**CLAUDE.local.md §24 "Harness Namespace 분리 정책"**(2026-05-23 chore commit `4f1135684`로 명문화)은 다음 분리 contract를 정의했다:

| 범위 | Skills 접두사 | Agents 경로 | `moai update` 동작 |
|------|---------------|------------|---------------------|
| 범용 배포 (template-managed) | `moai-*` (incl. `moai-harness-learner`, `moai-meta-harness`) | `.claude/agents/{core,expert,meta}/` | sync (overwrite) |
| **사용자 생성** | **`my-harness-*`** | **`.claude/agents/harness/`** | **NOT synced (보호)** |

후속으로 **SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001**(2026-05-23 PR #1048 머지, commit `767bc04a4`)이 이 contract를 Go 코드 (`internal/cli/update_namespace_protect.go` + `update.go:1156-1229` `isUserOwnedNamespace`)로 구현하고 14건의 `update_namespace_protect_test.go` + `update_preserve_my_harness_test.go` 단위 테스트로 강제했다.

그러나 **현재 (2026-05-25) 로컬 moai-adk-go dev 프로젝트**에는 §24 정책 도입 이전(2026-05-22~23 무렵)에 잘못된 prefix로 생성된 다음 잔여물이 남아 있다:

```
.claude/agents/harness/cli-template-specialist.md      (2026-05-22 작성, 2,111 bytes)
.claude/agents/harness/hook-ci-specialist.md           (2026-05-23 작성, 2,755 bytes)
.claude/agents/harness/quality-specialist.md           (2026-05-23 작성, 2,151 bytes)
.claude/agents/harness/workflow-specialist.md          (2026-05-23 작성, 2,250 bytes)
.claude/skills/moai-harness-cli-template/SKILL.md      (2026-05-22 작성, 4,683 bytes)
.claude/skills/moai-harness-patterns/SKILL.md          (2026-05-23 작성, 10,052 bytes)
```

§24.1에 따르면 `moai-harness-*` 접두사는 **moai-adk가 제공하는 builder/lifecycle** (현재 `moai-meta-harness` + `moai-harness-learner`만 valid)에 한정되며, 사용자 도메인 specialist는 **`my-harness-*`** prefix만 허용된다. 위 6개 파일은 이 contract를 위반한다 — 단, **template (`internal/template/templates/`)** 에는 이미 깨끗하게 정리되어 있으므로 (chore `4f1135684`에서 제거 완료, 본 SPEC plan-phase audit으로 검증 — `find` 결과 `core/expert/meta`만 존재) 외부 사용자에게 배포된 흔적은 없다.

**WHY**: §24 SSOT는 "정책 명문화 + 코드 구현 (UPDATE-NAMESPACE-PROTECT-001) + template 정리 (chore `4f1135684`)" 3가지를 모두 갖추었으나, **dev 프로젝트 자기 자신**이 정책 위반 상태로 남아 있다. SSOT 자기 일관성 (self-consistency)을 회복하지 않으면:

1. `moai update` Contract 회귀 검증 시 dev 프로젝트가 "보호되어야 할 user-owned namespace"의 부정 사례 (잘못된 prefix)로 작동하여 회귀 테스트의 발견력이 떨어진다
2. §24 정책을 인용하는 신규 SPEC들이 "dev 프로젝트도 잔여물이 있으니 이 정책은 약하게 강제된다"고 오해할 여지가 있다
3. `moai-harness-cli-template` / `moai-harness-patterns` 스킬은 진행 중인 `/moai project` 인터뷰 + `moai-meta-harness` 빌더 호출 시 잘못된 prefix 예시로 학습 데이터에 섞일 수 있다

**WHAT**: 본 SPEC은 (a) §24 SSOT 자기 일관성 검증 audit, (b) 로컬 dev 프로젝트 잔여물 정리 의무 명세, (c) audit를 회귀 방지 Go integration test로 codify하는 범위 정의이다. 실제 코드/파일 변경은 run-phase에서 수행한다 (본 plan-phase는 명세만).

## 2. 요구사항 (EARS)

### REQ-HNC-001: Template Cleanliness Invariant (Ubiquitous)

The system **shall** maintain the following invariants in `internal/template/templates/.claude/`:

- `.claude/agents/` SHALL contain ONLY the three system subdirectories `{core, expert, meta}/` — NO `harness/` subdirectory
- `.claude/skills/` SHALL contain the `moai-harness-learner/` skill but NO other `moai-harness-*` prefixed skill directory (즉 `moai-meta-harness/` + `moai-harness-learner/` 두 개만 valid `moai-harness*` 패턴)

### REQ-HNC-002: Local Dev Project Cleanup (Event-Driven)

**When** the run-phase of this SPEC executes, the system **shall** remove the following 6 leaked artifacts from the local moai-adk-go dev project (NOT the template):

```
.claude/agents/harness/cli-template-specialist.md
.claude/agents/harness/hook-ci-specialist.md
.claude/agents/harness/quality-specialist.md
.claude/agents/harness/workflow-specialist.md
.claude/skills/moai-harness-cli-template/
.claude/skills/moai-harness-patterns/
```

추가로 비어진 부모 디렉토리 `.claude/agents/harness/` 자체도 제거한다.

### REQ-HNC-003: §24 SSOT Cross-Reference Self-Consistency (Ubiquitous)

The CLAUDE.local.md §24 textual policy **shall** be cross-referenced from at least the following sources (각 cross-ref가 §24 contract를 약화시키지 않아야 함):

- `internal/cli/update.go:1166` (isUserOwnedNamespace REQ-UNP-002 comment)
- `internal/cli/update.go:1240-1244` (isMoaiManaged exclusion note)
- `internal/cli/update_namespace_protect.go:7` (package-level docstring)
- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention
- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation (Storage Roots table)

### REQ-HNC-004: Audit Verification Mechanism (Event-Driven)

**When** the run-phase executes the cleanup, the system **shall** verify completion with the following 3 grep assertions (all MUST return zero matches AFTER cleanup):

1. `find internal/template/templates/.claude/agents -type d -name harness` → 0 results (template invariant)
2. `find .claude/agents/harness -type f` → 0 results (local dev cleanup)
3. `ls -d .claude/skills/moai-harness-cli-template .claude/skills/moai-harness-patterns 2>/dev/null` → 0 results (local skill cleanup)

### REQ-HNC-005: Regression Prevention via Go Integration Test (Event-Driven)

**When** the run-phase Go integration test scope is implemented, the system **shall** add a regression test under `internal/template/` (구체적 파일명은 run-phase 결정) that asserts:

- `internal/template/templates/.claude/agents/` only contains `core`, `expert`, `meta` subdirectories
- `internal/template/templates/.claude/skills/` containing any `moai-harness-*` prefixed directory matches the canonical allow-list `{moai-harness-learner}` (with `moai-meta-harness` being a separate `moai-meta-*` prefix; the strict `moai-harness-*` allowlist is `{moai-harness-learner}`)

이 테스트는 template embedded.go 재생성 시 `make build` 산출물에 누출이 재발하는 경우를 컴파일/CI 시점에 차단한다.

### REQ-HNC-006: No Template Modification in This SPEC (Ubiquitous)

The cleanup actions defined by REQ-HNC-002 **shall not** modify any file under `internal/template/templates/`. Template is already clean (verified by REQ-HNC-001 invariant during plan-phase audit). 본 SPEC 범위는 로컬 dev 프로젝트 + Go test 추가 + cross-reference doc audit만 포함한다.

### REQ-HNC-007: Backup Before Cleanup (Event-Driven)

**When** the run-phase removes the 6 leaked artifacts, the system **shall** create a backup at `.moai/backups/harness-namespace-cleanup-{ISO}/` containing the 6 files in their pre-delete state, before any `rm` operation. 이 backup은 `moai update`의 `.moai/backups/update-*/` pattern (UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-010)과 같은 ISO-8601 hyphenated 형식을 사용한다.

## 3. 비요구사항 (Non-requirements)

- `moai update` Go 코드 변경 (UPDATE-NAMESPACE-PROTECT-001에서 이미 구현 완료, 본 SPEC은 그 contract의 적용 검증만)
- `moai-meta-harness` skill 본체 수정 (정상 동작 중, namespace generator 역할 유지)
- `moai-harness-learner` skill 수정 (template에 정상 존재, valid `moai-harness-*` 사용 사례)
- 새로운 `my-harness-*` 사용자 도메인 specialist 생성 (`/moai project` interview 후 `moai-meta-harness`가 수행할 작업, 본 SPEC 범위 외)
- §24 policy doctrine 수정 (현재 contract 그대로 유지)
- `.moai/harness/` 디렉토리 정리 (UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-003에서 user-owned로 보호 — 건드리지 않음)
- 사용자 프로젝트 마이그레이션 가이드 (외부 배포된 적 없음, 영향 없음)

## 4. 제약사항 (Constraints)

| 제약 | 근거 |
|------|------|
| Plan-phase 산출물만 본 SPEC 범위 (코드/파일 변경 X) | manager-spec scope; 실제 cleanup은 run-phase manager-develop |
| Multi-session race 활성 (`MULTI-SESSION-COORD-001` run-phase 동시 진행) | spawn 프롬프트 CRITICAL CONSTRAINTS |
| `internal/template/templates/` 수정 금지 | REQ-HNC-006 + 16-language neutrality (`CLAUDE.local.md` §15) |
| `internal/governance/`, `internal/session/registry*` 접근 금지 | COORD-001과 disjoint scope 유지 |
| Backup 형식은 ISO-8601 hyphenated (`2026-05-25T10-30-45Z`) | REQ-HNC-007 + UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-010 일관성 |
| Go test는 `internal/template/` 패키지 하위 (구체적 파일명 run-phase 결정) | REQ-HNC-005 |

## 5. 의존성

- **Strict dependency**: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (이미 머지됨 `767bc04a4`). 본 SPEC은 그 SPEC이 정의한 contract의 self-consistency 검증.
- **Related**: SPEC-V3R6-HARNESS-RENAME-001 (PR #1043, 2026-05-22 — `my-harness-*` → `moai-harness-*` 통합 시도, 그 후 §24 namespace 분리 정책 도입으로 부분 supersede됨). 본 SPEC은 §24 정책의 적용 검증으로 그 supersede를 완결.
- **Related**: SPEC-V3R4-HARNESS-NAMESPACE-001 (legacy `my-harness-*` 정책 첫 도입). 본 SPEC은 그 후속 cleanup.

## 6. 영향 분석

| 범주 | 영향 |
|------|------|
| 사용자 (external) | 없음 — template은 이미 깨끗, 외부 배포된 적 없는 dev-only 잔여물 정리 |
| 메인테이너 (local dev) | 6개 파일 제거 + 1 비어진 디렉토리 제거 |
| `moai update` 동작 | 변경 없음 (UPDATE-NAMESPACE-PROTECT-001에서 이미 구현, 본 SPEC은 적용 검증) |
| Go test surface | `internal/template/` 패키지에 회귀 방지 test 1-2개 추가 |
| Cross-reference doc | 검증만 수행, 변경 없음 (현재 6개 cross-ref 모두 일관) |
| Risk level | **Low** — 6개 파일이 모두 dev-only, runtime 동작에 무관 (workflow 실행 경로에서 참조되지 않음) |

## 7. Exclusions (What NOT to Build)

본 SPEC의 plan-phase 산출물은 **범위 명세만** 다루며, 다음 항목은 명시적으로 제외한다.

### 7.1 Template Modification 제외

`internal/template/templates/` 하위 어떤 파일도 본 SPEC의 run-phase에서 수정하지 않는다. Template은 plan-phase audit (REQ-HNC-001)에서 이미 invariant compliant로 확인됨. 만약 template 잔여물이 발견되면 본 SPEC을 confidence 0.0 BLOCK 처리하고 별도 hotfix SPEC을 생성한다 (run-phase blocker).

### 7.2 `moai update` Go 코드 수정 제외

`internal/cli/update.go`, `internal/cli/update_namespace_protect.go`, `internal/cli/update_archive.go` 의 Go 코드 본체는 본 SPEC 범위가 아니다 — UPDATE-NAMESPACE-PROTECT-001에서 이미 구현 완료. 본 SPEC은 그 contract의 적용 검증 (`isUserOwnedNamespace` 함수가 `my-harness-*` + `.claude/agents/harness/` 패턴을 보호함을 확인)에 한정한다.

### 7.3 새로운 사용자 도메인 Specialist 생성 제외

`my-harness-cli-template-specialist`, `my-harness-hook-ci-specialist` 등 사용자 도메인 specialist의 생성은 본 SPEC 범위가 아니다. 사용자가 명시적으로 `/moai project` Phase 5+ 인터뷰를 통해 `moai-meta-harness` builder를 호출하여 generate할 작업이며, 메인테이너가 dev 프로젝트에서 selfservice로 생성할 항목이 아니다 (CLAUDE.local.md §24.1 contract).

### 7.4 `.moai/harness/` 디렉토리 수정 제외

`.moai/harness/main.md`, `.moai/harness/interview-results.md`, `.moai/harness/learning-history/`, `.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml` 등 모든 `.moai/harness/` 하위는 user-owned (UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-003) 이며 본 SPEC에서 건드리지 않는다. backup 대상도 아니다.

### 7.5 `internal/governance/`, `internal/session/registry*` 접근 제외

병렬 세션 (SPEC-V3R6-MULTI-SESSION-COORD-001 run-phase)이 작업 중인 영역과 disjoint scope를 유지하기 위해 본 SPEC은 그 영역에 read도 write도 수행하지 않는다 (CRITICAL CONSTRAINTS).

### 7.6 §24 Policy Doctrine 수정 제외

CLAUDE.local.md §24 본문 텍스트 자체는 본 SPEC에서 변경하지 않는다. 본 SPEC은 §24 contract의 self-consistency 검증과 그 적용 완결만을 목적으로 한다 — 정책 doctrine은 이미 4f1135684에서 완성된 상태로 보존.

### 7.7 PR/Commit 작업 제외 (Plan-phase)

본 plan-phase 산출물 (이 4 파일)은 manager-spec 단독으로 작성하며, `git add` / `git commit` / `git push` / `gh pr create` 작업은 일체 수행하지 않는다. 그 작업은 orchestrator의 후속 단계에서 일괄 처리된다 (spawn 프롬프트 CRITICAL CONSTRAINT 2).
