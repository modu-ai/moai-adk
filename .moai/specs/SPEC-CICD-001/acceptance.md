# SPEC-CICD-001: 인수 기준

## 메타데이터

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CICD-001 |
| 제목 | CI/CD AI 자동화 재설계 |
| 관련 SPEC | spec.md, plan.md |

---

## 1. 인수 시나리오

### AC-1: AI 코드 리뷰 단일화 (REQ-1)

**Scenario 1.1: PR 오픈 시 단일 AI 리뷰 실행**

```gherkin
Given PR이 main 브랜치를 대상으로 열려 있다
  And PR에 Go 소스 코드 변경이 포함되어 있다
When PR이 opened 상태가 된다
Then claude-code-review.yml만 실행되어야 한다
  And claude.yml은 실행되지 않아야 한다
  And AI 코드 리뷰 코멘트가 1세트만 생성되어야 한다
```

**Scenario 1.2: PR 댓글에서 @claude 멘션 시 무시**

```gherkin
Given PR이 열려 있다
When PR 댓글에 "@claude 이 코드 리뷰해줘"라고 작성한다
Then claude.yml이 트리거되지 않아야 한다
  And PR 리뷰는 claude-code-review.yml로만 수행되어야 한다
```

**Scenario 1.3: PR review comment에서 @claude 멘션 시 무시**

```gherkin
Given PR이 열려 있고 코드 리뷰가 진행 중이다
When PR review comment에 "@claude"를 포함한 코멘트가 작성된다
Then claude.yml이 트리거되지 않아야 한다
```

### AC-2: 조건부 Auto-merge (REQ-2)

**Scenario 2.1: automerge 레이블 + 모든 조건 통과 시 자동 merge**

```gherkin
Given PR이 main 브랜치를 대상으로 열려 있다
  And CI (Test, Lint, Build)가 모두 통과했다
  And Claude Code Review Quality Gate에서 Important findings가 0이다
When "automerge" 레이블이 수동으로 추가된다
Then PR이 squash merge되어야 한다
  And 소스 브랜치가 자동 삭제되어야 한다
```

**Scenario 2.2: automerge 레이블 없이 merge 안 됨**

```gherkin
Given PR이 main 브랜치를 대상으로 열려 있다
  And CI가 모두 통과했다
  And Quality Gate가 통과했다
When "automerge" 레이블이 없다
Then automerge.yml이 merge를 수행하지 않아야 한다
```

**Scenario 2.3: CI 실패 시 automerge 레이블이 있어도 merge 안 됨**

```gherkin
Given PR에 "automerge" 레이블이 추가되어 있다
  And CI Test job이 실패했다
When automerge.yml이 실행된다
Then merge가 수행되지 않아야 한다
```

**Scenario 2.4: Quality Gate 실패 시 merge 안 됨**

```gherkin
Given PR에 "automerge" 레이블이 추가되어 있다
  And CI가 통과했다
  And Claude Code Review에서 Important findings가 1 이상이다
When Review Quality Gate가 fail로 완료된다
Then automerge.yml이 merge를 수행하지 않아야 한다
```

**Scenario 2.5: automerge 레이블이 자동 추가되지 않음**

```gherkin
Given 새로운 PR이 opened 상태가 된다
When community.yml이 실행된다
Then "automerge" 레이블이 자동으로 추가되지 않아야 한다
```

### AC-3: Issue 전용 @claude (REQ-3)

**Scenario 3.1: Issue에서 @claude 멘션 시 정상 동작**

```gherkin
Given GitHub Issue가 열려 있다
When Issue 본문 또는 제목에 "@claude"가 포함되어 있다
Then claude.yml이 실행되어야 한다
  And Claude Code Action이 Issue 컨텍스트에서 응답해야 한다
```

**Scenario 3.2: Issue 댓글에서 @claude 멘션 시 정상 동작**

```gherkin
Given GitHub Issue가 열려 있다
When Issue 댓글에 "@claude"가 포함된 코멘트가 작성된다
Then claude.yml이 실행되어야 한다
  And Claude Code Action이 Issue 컨텍스트에서 응답해야 한다
```

### AC-4: CI-only 변경 시 리뷰 스킵 (REQ-4)

**Scenario 4.1: .github/ 파일만 변경한 PR**

```gherkin
Given PR이 .github/workflows/ci.yml만 변경했다
When PR이 opened 상태가 된다
Then claude-code-review.yml이 실행되지 않아야 한다
  And ci.yml은 정상 실행되어야 한다
```

**Scenario 4.2: .github/ 파일과 Go 파일을 함께 변경한 PR**

```gherkin
Given PR이 .github/workflows/ci.yml과 internal/cli/init.go를 변경했다
When PR이 opened 상태가 된다
Then claude-code-review.yml이 정상 실행되어야 한다
  And Go 파일 변경에 대해 AI 리뷰가 수행되어야 한다
```

### AC-5: CodeQL Go 버전 동기화 (REQ-5)

**Scenario 5.1: CodeQL이 Go 1.26으로 실행**

```gherkin
Given codeql.yml의 go-version이 "1.26"으로 설정되어 있다
When main 브랜치에 push가 발생한다
Then CodeQL 분석이 Go 1.26으로 정상 실행되어야 한다
  And 보안 분석 결과가 생성되어야 한다
```

### AC-6: community.yml 책임 분리 (REQ-6)

**Scenario 6.1: community.yml에 auto-merge 없음**

```gherkin
Given community.yml이 업데이트되었다
When PR이 opened 상태가 된다
Then community.yml의 labeler job만 실행되어야 한다
  And auto-merge 관련 동작이 발생하지 않아야 한다
  And "automerge" 레이블이 자동 추가되지 않아야 한다
```

**Scenario 6.2: stale 관리 정상 동작**

```gherkin
Given community.yml에서 auto-merge가 제거되었다
When 스케줄에 의해 stale job이 실행된다
Then 30일간 활동 없는 Issue/PR에 stale 레이블이 추가되어야 한다
  And 기존 stale 기능이 영향받지 않아야 한다
```

---

## 2. Quality Gate 기준

### 워크플로우 문법 검증

- [ ] 모든 변경된 `.yml` 파일이 유효한 GitHub Actions 문법을 따른다
- [ ] `act` 또는 유사 도구로 로컬 워크플로우 문법 검증 (선택)

### 기능 검증

- [ ] 테스트 PR 생성 시 `claude-code-review.yml`만 AI 리뷰를 실행한다
- [ ] `claude.yml`이 PR 이벤트에 반응하지 않는다
- [ ] `automerge` 레이블 수동 추가 시에만 auto-merge가 동작한다
- [ ] `.github/` 파일만 변경한 PR에서 AI 리뷰가 스킵된다
- [ ] CodeQL이 Go 1.26으로 정상 실행된다
- [ ] `community.yml`에서 auto-merge가 완전히 제거되었다

### 비용 검증

- [ ] PR당 Claude Code Action 호출 횟수가 1회로 제한된다
- [ ] 불필요한 워크플로우 실행이 감소한다

---

## 3. 검증 방법

| 검증 항목 | 방법 | 도구 |
|----------|------|------|
| YAML 문법 | `yamllint` 또는 GitHub Actions 자체 검증 | GitHub UI |
| 트리거 동작 | 테스트 PR/Issue 생성 후 Actions 탭 확인 | GitHub Actions |
| auto-merge 조건 | 레이블 추가/제거 후 merge 동작 확인 | GitHub UI |
| 리뷰 횟수 | PR 코멘트에서 AI 리뷰 봇 코멘트 수 확인 | GitHub UI |
| Go 버전 | CodeQL 로그에서 Go 버전 확인 | GitHub Actions 로그 |

---

## 4. Definition of Done

- [ ] `claude.yml`에서 PR 관련 트리거가 모두 제거되었다
- [ ] `claude-code-review.yml`에 `paths-ignore` 설정이 추가되었다
- [ ] `community.yml`에서 `auto-merge` job이 제거되었다
- [ ] `automerge.yml`이 조건부 merge 로직으로 생성되었다
- [ ] `codeql.yml`의 Go 버전이 1.26으로 업데이트되었다
- [ ] 모든 워크플로우 파일이 GitHub Actions에서 정상 파싱된다
- [ ] 테스트 PR로 단일 AI 리뷰 동작을 확인했다
- [ ] GoosLab bot 재설정이 수동 조치 항목으로 문서화되었다
