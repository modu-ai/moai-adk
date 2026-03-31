# SPEC-CICD-001: 구현 계획

## 메타데이터

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CICD-001 |
| 제목 | CI/CD AI 자동화 재설계 |
| 관련 SPEC | spec.md, acceptance.md |

---

## 1. 마일스톤

### Primary Goal: AI 리뷰 단일화 및 auto-merge 분리

**우선순위: High**

| 단계 | 작업 | 변경 파일 | 의존성 |
|------|------|----------|--------|
| 1-1 | `claude.yml` Issue 전용 제한 | `claude.yml` | 없음 |
| 1-2 | `claude-code-review.yml` paths-ignore 추가 | `claude-code-review.yml` | 없음 |
| 1-3 | `community.yml`에서 auto-merge 제거 | `community.yml` | 없음 |
| 1-4 | `automerge.yml` 신규 생성 | `automerge.yml` (NEW) | 1-3 완료 |

단계 1-1, 1-2, 1-3은 독립적이므로 **병렬 실행 가능**.
단계 1-4는 1-3 완료 후 실행 (auto-merge 중복 방지).

### Secondary Goal: 보안 및 일관성

**우선순위: Medium**

| 단계 | 작업 | 변경 파일 | 의존성 |
|------|------|----------|--------|
| 2-1 | `codeql.yml` Go 버전 업데이트 | `codeql.yml` | 없음 |
| 2-2 | GoosLab bot 재설정 문서화 | SPEC 문서 | 없음 |

### Optional Goal: 추가 개선

**우선순위: Low**

| 단계 | 작업 | 변경 파일 | 의존성 |
|------|------|----------|--------|
| 3-1 | Branch protection에 Review Quality Gate 추가 고려 | GitHub Settings | Primary Goal 완료 |
| 3-2 | CodeRabbit 설정 최적화 (중복 카테고리 제거) | `.coderabbitai.yaml` | 없음 |

---

## 2. 기술적 접근

### 2.1 `claude.yml` 수정 전략

**현재 문제**: `issue_comment` 이벤트가 Issue와 PR 댓글 모두에서 발생하여 PR에서도 @claude가 동작함.

**해결 방법**: `issue_comment` 조건에 `github.event.issue.pull_request` 존재 여부를 체크하여 PR 댓글을 필터링.

```yaml
# 변경 전
if: |
  (github.event_name == 'issue_comment' && contains(github.event.comment.body, '@claude')) ||
  (github.event_name == 'pull_request_review_comment' && ...) ||
  (github.event_name == 'pull_request_review' && ...) ||
  (github.event_name == 'issues' && ...)

# 변경 후
if: |
  (github.event_name == 'issue_comment' && 
   !github.event.issue.pull_request && 
   contains(github.event.comment.body, '@claude')) ||
  (github.event_name == 'issues' && 
   (contains(github.event.issue.body, '@claude') || 
    contains(github.event.issue.title, '@claude')))
```

추가 변경:
- `on` 트리거에서 `pull_request_review_comment`, `pull_request_review` 제거
- `permissions`에서 `actions: read` 제거 (PR CI 결과 읽기 불필요)
- `pull-requests: read` 권한 제거 (Issue 전용이므로)
- Setup Go 단계 제거 고려 (Issue 처리에 Go 빌드 불필요한 경우)

### 2.2 `claude-code-review.yml` 수정 전략

**변경 내용**: `paths-ignore` 추가로 CI-only 변경 시 불필요한 리뷰 방지.

```yaml
on:
  pull_request:
    types: [opened, ready_for_review, reopened]
    paths-ignore:
      - ".github/**"
      - "*.md"
      - "LICENSE"
```

**주의**: `*.md` 및 `LICENSE` 파일도 코드 리뷰 불필요하므로 함께 제외 고려.

### 2.3 `community.yml` 수정 전략

**제거 대상**:
- `auto-merge` job 전체 (line 102-133)
- `check_suite` 트리거 (line 10-11)
- `status` 트리거 (line 12)
- `permissions`의 `contents: write` (auto-merge 전용이었음)

**유지 대상**:
- `welcome` job
- `stale` job
- `labeler` job

### 2.4 `automerge.yml` 신규 생성 전략

**핵심 설계 원칙**:
- `automerge` 레이블은 수동으로만 추가 (자동 추가 금지)
- merge 전제 조건: CI 통과 + Review Quality Gate 통과
- squash merge + 브랜치 자동 삭제

**트리거 설계**:
```yaml
on:
  # automerge 레이블이 추가될 때
  pull_request:
    types: [labeled]
  # CI 또는 Review Quality Gate 완료 시 재확인
  check_suite:
    types: [completed]
  workflow_run:
    workflows: ["CI", "Review Quality Gate"]
    types: [completed]
```

**조건부 실행**:
- `automerge` 레이블 존재 확인
- 모든 required status check 통과 확인
- `pascalgn/automerge-action` 사용 (기존과 동일한 action)
- `MERGE_LABELS: "automerge"` 설정으로 레이블 기반 merge

### 2.5 `codeql.yml` 수정 전략

단순 버전 업데이트:
```yaml
# 변경 전
go-version: "1.25"

# 변경 후
go-version: "1.26"
```

---

## 3. 아키텍처 설계 방향

### 3.1 워크플로우 계층 구조

```
[Tier 1] ci.yml ─────────────────────── Required Status Check
                                            |
[Tier 2] claude-code-review.yml ────── AI Review (single)
              |                             |
         review-quality-gate.yml ───── Quality Gate
                                            |
[Tier 3] claude.yml ──────────────── Issue-only @claude
                                            |
[Tier 4] codeql.yml ──────────────── Security Analysis
                                            |
[Tier 5] community.yml ──────────── Welcome + Stale + Label
                                            |
[Tier 6] automerge.yml ──────────── Conditional Merge
              |                        (label + CI + QG)
         release.yml ─────────────── Tag-triggered Release
```

### 3.2 이벤트 흐름 (PR 라이프사이클)

```
PR Opened
  |
  +-- ci.yml 실행 (필수)
  +-- claude-code-review.yml 실행 (paths-ignore 적용)
  +-- codeql.yml 실행
  +-- community.yml/labeler 실행
  +-- CodeRabbitAI 실행 (외부)
  |
Claude Code Review 완료
  |
  +-- review-quality-gate.yml 실행
  |
  [수동] automerge 레이블 추가
  |
  +-- automerge.yml 실행
       |-- CI 통과 확인
       |-- Quality Gate 통과 확인
       +-- Squash Merge + 브랜치 삭제
```

### 3.3 이벤트 흐름 (Issue 라이프사이클)

```
Issue Opened (with @claude)
  |
  +-- claude.yml 실행
       |-- Claude Code Action (Issue 컨텍스트)
       +-- 응답 작성

Issue Comment (with @claude, NOT on PR)
  |
  +-- claude.yml 실행
       |-- !github.event.issue.pull_request 확인
       +-- Claude Code Action (Issue 컨텍스트)
```

---

## 4. 리스크 및 대응

| 리스크 | 영향 | 대응 |
|--------|------|------|
| `paths-ignore`가 Go 파일 변경과 `.github` 파일을 동시에 포함하는 PR에서 리뷰 스킵 | Low (GitHub는 paths-ignore + paths 조합을 OR로 처리) | `.github/**`만 변경한 PR에서만 스킵됨 (다른 파일 변경 시 정상 실행) |
| `automerge.yml`의 `workflow_run` 트리거가 복잡한 이벤트 체인 생성 | Medium | 대안으로 `check_suite` + label 조건만 사용 가능 |
| GoosLab bot이 여전히 PR에 @claude 멘션 | Low (claude.yml에서 PR 필터링 추가) | claude.yml 필터링이 안전장치 역할 |
| auto-merge 분리 과도기에 레이블 없는 PR이 merge 안 됨 | Low | 전환 전 기존 PR 정리, 필요 시 수동 merge |

---

## 5. 변경 범위 요약

| 파일 | 변경 유형 | 변경 규모 |
|------|----------|----------|
| `claude.yml` | 수정 (트리거/조건 축소) | Medium |
| `claude-code-review.yml` | 수정 (paths-ignore 추가) | Small |
| `community.yml` | 수정 (auto-merge job 제거) | Medium |
| `automerge.yml` | 신규 생성 | Medium |
| `codeql.yml` | 수정 (Go 버전) | Trivial |
| `review-quality-gate.yml` | 변경 없음 | - |
| `ci.yml` | 변경 없음 | - |
| `release.yml` | 변경 없음 | - |
| `test-install.yml` | 변경 없음 | - |

**총 변경 파일 수**: 5개 (신규 1개 포함)
