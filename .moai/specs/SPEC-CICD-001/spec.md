# SPEC-CICD-001: CI/CD AI 자동화 재설계

## 메타데이터

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CICD-001 |
| 제목 | CI/CD AI 자동화 재설계 |
| 상태 | Planned |
| 우선순위 | High |
| 생성일 | 2026-03-31 |
| 담당 | expert-devops |
| 관련 파일 | `.github/workflows/*.yml` |

---

## 1. Environment (환경)

### 1.1 현재 시스템 구성

- **CI 플랫폼**: GitHub Actions
- **AI 코드 리뷰**: Claude Code Action (anthropics/claude-code-action@v1.0.82)
- **보안 분석**: CodeQL (github/codeql-action@v4)
- **커뮤니티 자동화**: actions/first-interaction, actions/stale, actions/labeler, pascalgn/automerge-action
- **외부 리뷰 봇**: CodeRabbitAI (자동), GoosLab bot (@claude 멘션)
- **언어/빌드**: Go 1.26, GoReleaser
- **브랜치 보호**: required_approving_review_count: 0, dismiss stale reviews: true

### 1.2 워크플로우 파일 현황

| 파일 | 역할 | 상태 |
|------|------|------|
| `ci.yml` | 테스트 (3 OS), Lint, Build (5 cross-compile) | 유지 |
| `claude-code-review.yml` | PR 오픈 시 AI 자동 리뷰 | 수정 필요 |
| `claude.yml` | @claude 멘션 핸들러 (Issue + PR) | 수정 필요 |
| `review-quality-gate.yml` | Claude 리뷰 심각도 파싱 | 개선 |
| `community.yml` | Welcome, Stale, Labeler, Auto-merge | 분리 필요 |
| `codeql.yml` | 보안 분석 (Go 1.25) | 버전 업데이트 |
| `release.yml` | GoReleaser on tag | 유지 |
| `test-install.yml` | 설치 테스트 | 유지 |

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- A1: Claude Code Action의 code-review 플러그인이 유일한 AI 리뷰 도구로 충분하다
- A2: CodeRabbitAI는 무료이며 Claude와 다른 관점을 제공하므로 유지 가치가 있다
- A3: GoosLab bot의 @claude PR 멘션은 리포지토리 외부 설정이며, 별도 수동 조치가 필요하다
- A4: `review-quality-gate.yml`이 `claude-code-review.yml`의 check run 완료에 의존한다
- A5: CodeQL의 Go 1.25 -> 1.26 업데이트는 breaking change 없이 가능하다

### 2.2 비즈니스 가정

- A6: 현재 3중 AI 리뷰로 인한 비용이 유의미하게 낭비되고 있다
- A7: 모든 PR의 무조건 auto-merge는 보안 및 품질 리스크를 초래한다
- A8: 1인 개발 프로젝트이므로 `required_approving_review_count: 0`은 유지하되, AI 리뷰 품질 게이트는 필요하다

---

## 3. Requirements (요구사항)

### REQ-1: AI 코드 리뷰 단일화 (Critical)

**[Ubiquitous]** 시스템은 **항상** PR당 하나의 AI 코드 리뷰만 실행해야 한다.

- REQ-1.1: `claude-code-review.yml`이 유일한 자동 AI 코드 리뷰 워크플로우여야 한다
- REQ-1.2: `claude.yml`에서 PR 관련 트리거(`issue_comment`의 PR 컨텍스트, `pull_request_review_comment`, `pull_request_review`)를 제거해야 한다
- REQ-1.3: `claude.yml`은 Issue 이벤트(`issues.opened`, `issues.assigned`, Issue 댓글의 `issue_comment`)에서만 동작해야 한다

### REQ-2: PR 생성 시 즉시 Auto-merge 활성화 (Critical)

**WHEN** PR이 생성되면 **THEN** `gh pr merge --auto --squash --delete-branch`로 auto-merge를 즉시 활성화해야 한다. GitHub이 required status checks (CI) 통과를 대기한 후 자동 merge한다.

- REQ-2.1: `community.yml`에서 auto-merge 관련 로직을 완전히 제거해야 한다
- REQ-2.2: 별도 `automerge.yml` 워크플로우 불필요 (GitHub 내장 auto-merge 사용)
- REQ-2.3: `/moai sync --pr` 워크플로우에서 PR 생성 후 `gh pr merge --auto` 자동 실행
- REQ-2.4: merge 조건: GitHub branch protection의 required status checks 통과

### REQ-3: claude.yml Issue 전용 제한 (High)

**WHEN** GitHub Issue에서 @claude가 멘션되면 **THEN** Claude Code Action이 Issue 컨텍스트에서만 실행되어야 한다.

- REQ-3.1: `issue_comment` 트리거는 유지하되, PR 컨텍스트의 댓글은 필터링해야 한다
- REQ-3.2: `pull_request_review_comment` 트리거를 제거해야 한다
- REQ-3.3: `pull_request_review` 트리거를 제거해야 한다
- REQ-3.4: `issues` 트리거(opened, assigned)는 유지해야 한다

### REQ-4: CI-only 변경 시 리뷰 스킵 (Medium)

**WHEN** PR이 `.github/` 디렉토리의 파일만 변경한 경우 **THEN** AI 코드 리뷰를 스킵해야 한다.

- REQ-4.1: `claude-code-review.yml`에 `paths-ignore: [".github/**"]` 추가
- REQ-4.2: 워크플로우 파일 자체의 변경이 불필요한 AI 리뷰를 트리거하지 않아야 한다

### REQ-5: CodeQL Go 버전 동기화 (Medium)

**[Ubiquitous]** 시스템은 **항상** 모든 워크플로우에서 동일한 Go 버전(1.26)을 사용해야 한다.

- REQ-5.1: `codeql.yml`의 Go 버전을 1.25에서 1.26으로 업데이트해야 한다

### REQ-6: community.yml 책임 분리 (Medium)

**[Ubiquitous]** 각 워크플로우 파일은 **항상** 단일 책임 원칙을 따라야 한다.

- REQ-6.1: `community.yml`은 Welcome, Stale, Labeler 기능만 유지해야 한다
- REQ-6.2: auto-merge 관련 job과 트리거를 `community.yml`에서 제거해야 한다
- REQ-6.3: `community.yml`에서 `check_suite`, `status` 트리거를 제거해야 한다 (auto-merge 전용이었음)
- REQ-6.4: `community.yml`의 `permissions`에서 `contents: write` 불필요 (auto-merge 제거 시)

### REQ-7: GoosLab Bot 재설정 문서화 (Low)

**[Optional]** 가능하면 GoosLab bot의 PR @claude 멘션 동작을 비활성화하는 방법을 문서화해야 한다.

- REQ-7.1: GoosLab bot 설정이 리포지토리 외부에 있으므로 수동 조치 항목으로 문서화
- REQ-7.2: bot이 PR에 @claude를 호출하지 않도록 재설정 권고

### REQ-8: CodeRabbit 공존 보장 (Low)

시스템은 CodeRabbitAI와 Claude Code Review가 **충돌하지 않아야 한다**.

- REQ-8.1: 두 리뷰 봇이 동일 PR에서 동시 실행 가능해야 한다
- REQ-8.2: CodeRabbit은 별도 관점(보안, 스타일)을 제공하므로 유지

---

## 4. Specifications (명세)

### 4.1 워크플로우 아키텍처 (목표 상태)

```
Tier 1: Core Quality (Required, blocks merge)
  ci.yml                    -- 테스트, Lint, Build (변경 없음)

Tier 2: AI Code Review (Single, unified)
  claude-code-review.yml    -- 유일한 자동 AI 리뷰
  review-quality-gate.yml   -- 리뷰 심각도 파싱 (변경 없음)

Tier 3: On-Demand AI (Issues only)
  claude.yml                -- Issue 전용 @claude 핸들러

Tier 4: Security
  codeql.yml                -- CodeQL (Go 1.26으로 업데이트)

Tier 5: Community
  community.yml             -- Welcome + Stale + Labeler (auto-merge 제거)

Tier 6: Conditional Merge
  automerge.yml             -- 조건부 auto-merge (NEW)

Release:
  release.yml               -- GoReleaser (변경 없음)
  test-install.yml          -- 설치 테스트 (변경 없음)
```

### 4.2 파일별 변경 명세

#### `claude.yml` 수정

변경 전 트리거:
- `issue_comment` (Issue + PR 댓글 모두)
- `pull_request_review_comment`
- `pull_request_review`
- `issues` (opened, assigned)

변경 후 트리거:
- `issue_comment` (Issue 댓글만, PR 댓글 필터링)
- `issues` (opened, assigned)

조건 변경:
- `github.event_name == 'pull_request_review_comment'` 제거
- `github.event_name == 'pull_request_review'` 제거
- `issue_comment`에서 `github.event.issue.pull_request`가 없는 경우만 실행
- PR 전용 권한(`actions: read`) 제거

#### `claude-code-review.yml` 수정

추가:
- `paths-ignore: [".github/**"]`로 CI-only 변경 스킵

#### `community.yml` 수정

제거:
- `auto-merge` job 전체
- `check_suite` 트리거
- `status` 트리거
- `contents: write` 권한

유지:
- `welcome` job
- `stale` job
- `labeler` job

#### `automerge.yml` 신규 생성

트리거:
- `pull_request` (labeled)
- `check_suite` (completed)
- `workflow_run` (completed, workflows: ["CI", "Review Quality Gate"])

조건:
- `automerge` 레이블이 존재할 것
- CI 워크플로우 통과
- Review Quality Gate 통과 (Important == 0)

동작:
- squash merge
- 브랜치 삭제

#### `codeql.yml` 수정

변경:
- `go-version: "1.25"` -> `go-version: "1.26"`

### 4.3 수동 조치 항목

| 항목 | 설명 | 담당 |
|------|------|------|
| GoosLab bot 재설정 | PR에서 @claude 멘션 비활성화 | GOOS (수동) |
| Branch protection 확인 | required status checks에 `Review Quality Gate / parse-severity` 추가 고려 | GOOS (수동) |

### 4.4 Traceability (추적성)

| 요구사항 | 변경 파일 | 테스트 방법 |
|----------|----------|------------|
| REQ-1 | `claude-code-review.yml`, `claude.yml` | PR 생성 시 AI 리뷰 1회만 실행 확인 |
| REQ-2 | `community.yml`, `automerge.yml` (NEW) | automerge 레이블 수동 추가 후 조건부 merge 확인 |
| REQ-3 | `claude.yml` | Issue에서 @claude 멘션 시 동작, PR에서는 무시 확인 |
| REQ-4 | `claude-code-review.yml` | `.github/` 파일만 변경한 PR에서 리뷰 스킵 확인 |
| REQ-5 | `codeql.yml` | CodeQL 분석 정상 실행 확인 |
| REQ-6 | `community.yml` | Welcome, Stale, Labeler만 동작 확인 |
| REQ-7 | 문서 | GoosLab bot 설정 변경 가이드 확인 |
| REQ-8 | N/A | CodeRabbit + Claude Review 동시 실행 확인 |
