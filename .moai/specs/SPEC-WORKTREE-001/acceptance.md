# SPEC-WORKTREE-001: 인수 기준

---
spec_id: SPEC-WORKTREE-001
type: acceptance
---

## 시나리오 1: SPEC-ID로 worktree 생성 시 글로벌 경로 사용

```gherkin
Feature: SPEC-ID 기반 글로벌 worktree 생성

  Scenario: SPEC-ID 입력 시 ~/.moai/worktrees/{Project}/ 하위에 생성
    Given 프로젝트 루트에 go.mod 파일이 존재하고 module path가 "github.com/modu-ai/moai-adk"
    And ~/.moai/worktrees/ 디렉토리가 존재하지 않음
    And --path 플래그가 지정되지 않음
    When 사용자가 "moai worktree new SPEC-AUTH-001"을 실행
    Then git worktree가 "~/.moai/worktrees/moai-adk/SPEC-AUTH-001/" 경로에 생성됨
    And 브랜치명은 "feature/SPEC-AUTH-001"
    And "~/.moai/worktrees/moai-adk/" 디렉토리가 자동 생성됨
    And 성공 메시지에 생성된 경로가 포함됨
```

## 시나리오 2: 일반 브랜치명으로 worktree 생성

```gherkin
  Scenario: 비-SPEC-ID 브랜치도 글로벌 경로에 생성
    Given 프로젝트 루트에 go.mod 파일이 존재하고 module path가 "github.com/modu-ai/moai-adk"
    And --path 플래그가 지정되지 않음
    When 사용자가 "moai worktree new feature-x"을 실행
    Then git worktree가 "~/.moai/worktrees/moai-adk/feature-x/" 경로에 생성됨
    And 브랜치명은 "feature-x" (변환 없음)
```

## 시나리오 3: --path 플래그 우선

```gherkin
  Scenario: --path 플래그가 글로벌 경로 로직을 override
    Given --path 플래그가 "/tmp/my-worktree"로 지정됨
    When 사용자가 "moai worktree new SPEC-AUTH-001 --path /tmp/my-worktree"을 실행
    Then git worktree가 "/tmp/my-worktree" 경로에 생성됨
    And 글로벌 경로 로직은 실행되지 않음
```

## 시나리오 4: 프로젝트명 감지 -- go.mod 우선

```gherkin
Feature: 프로젝트명 자동 감지

  Scenario: go.mod module path에서 프로젝트명 추출
    Given go.mod 파일에 "module github.com/modu-ai/moai-adk" 라인이 존재
    When detectProjectName() 함수가 호출됨
    Then 반환값은 "moai-adk"
```

## 시나리오 5: 프로젝트명 감지 -- git remote fallback

```gherkin
  Scenario: go.mod 없을 때 git remote에서 프로젝트명 추출
    Given go.mod 파일이 존재하지 않음
    And git remote origin이 "git@github.com:modu-ai/moai-adk-go.git"
    When detectProjectName() 함수가 호출됨
    Then 반환값은 "moai-adk-go"
```

## 시나리오 6: 프로젝트명 감지 -- 디렉토리명 fallback

```gherkin
  Scenario: go.mod와 git remote 모두 없을 때 디렉토리명 사용
    Given go.mod 파일이 존재하지 않음
    And git remote가 설정되지 않음
    And 현재 디렉토리가 "/Users/goos/projects/my-project"
    When detectProjectName() 함수가 호출됨
    Then 반환값은 "my-project"
```

## 시나리오 7: 하위 호환성 경고

```gherkin
Feature: 레거시 worktree 감지 및 경고

  Scenario: 기존 .moai/worktrees/ 디렉토리에 worktree가 있을 때 경고
    Given 프로젝트 루트에 ".moai/worktrees/SPEC-OLD-001/" 디렉토리가 존재
    When 사용자가 "moai worktree new SPEC-NEW-001"을 실행
    Then stderr에 마이그레이션 안내 메시지가 출력됨
    And 새 worktree는 정상적으로 글로벌 경로에 생성됨
    And 기존 worktree는 영향받지 않음
```

## 시나리오 8: os.UserHomeDir 실패 시 fallback

```gherkin
Feature: 홈 디렉토리 해석 실패 대응

  Scenario: 홈 디렉토리를 얻을 수 없을 때 에러 반환
    Given os.UserHomeDir()가 에러를 반환하는 환경
    When 사용자가 "moai worktree new SPEC-AUTH-001"을 실행
    Then 명확한 에러 메시지가 반환됨
    And 에러 메시지에 "home directory" 관련 안내가 포함됨
```

## 시나리오 9: worktree_orchestrator SPEC-ID 매칭

```gherkin
Feature: 글로벌 경로에서 SPEC-ID 매칭

  Scenario: 글로벌 경로의 worktree에서 SPEC-ID로 검색
    Given SPEC-AUTH-001에 대한 worktree가 "~/.moai/worktrees/moai-adk/SPEC-AUTH-001/"에 존재
    When findWorktreeForSpec("SPEC-AUTH-001")이 호출됨
    Then WorktreeContext가 정상적으로 반환됨
    And WorktreeContext.SpecID는 "SPEC-AUTH-001"
    And WorktreeContext.WorktreeDir는 글로벌 경로
```

## 시나리오 10: launcher cleanup 확장

```gherkin
Feature: 세션 정리 시 글로벌 worktree 포함

  Scenario: cleanupMoaiWorktrees가 글로벌 경로도 탐색
    Given "~/.moai/worktrees/moai-adk/" 하위에 "worker-SPEC-XXX" worktree가 존재
    When cleanupMoaiWorktrees 함수가 실행됨
    Then 글로벌 경로의 worker worktree도 정리 대상에 포함됨
```

## 시나리오 11: 금지 동작 검증

```gherkin
Feature: 프로젝트 내부 worktree 생성 금지

  Scenario: --path 없이 실행 시 프로젝트 내부에 worktree 생성하지 않음
    Given --path 플래그가 지정되지 않음
    When 사용자가 "moai worktree new SPEC-AUTH-001"을 실행
    Then ".moai/worktrees/SPEC-AUTH-001/" 경로에 worktree가 생성되지 않음
    And 프로젝트 디렉토리 내부에 새 디렉토리가 생성되지 않음
```

---

## 품질 게이트

### 테스트 커버리지

- [ ] `detectProjectName()` 함수: 3가지 fallback 시나리오 (go.mod, git remote, cwd)
- [ ] `runNew()` 수정: SPEC-ID 경로, 브랜치명 경로, --path override
- [ ] 하위 호환성: 기존 `.moai/worktrees/` 감지 및 경고
- [ ] 에러 처리: `os.UserHomeDir()` 실패, `os.MkdirAll` 실패
- [ ] `worktree_orchestrator` 매칭: 글로벌 경로에서 `filepath.Base` 동작 검증
- [ ] 기존 테스트 수정: `subcommands_test.go`, hook 테스트 경로 업데이트

### 검증 명령

```bash
# 전체 테스트 실행 (race detector 포함)
go test -race ./...

# worktree 패키지 테스트
go test -v ./internal/cli/worktree/...

# hook 테스트
go test -v ./internal/hook/...

# launcher 테스트
go test -v ./internal/cli/ -run TestCleanup

# 린터 실행
golangci-lint run

# 템플릿 빌드
make build
```

### Definition of Done

- [ ] 모든 기존 테스트 통과 (`go test -race ./...`)
- [ ] 신규 테스트 추가 (`detectProjectName` + 글로벌 경로 통합 테스트)
- [ ] `golangci-lint run` 경고 없음
- [ ] 문서 경로 표기 통일 완료
- [ ] 템플릿 동기화 및 `make build` 성공
- [ ] `--path` 플래그 동작 변경 없음 확인
- [ ] 하위 호환성 경고 메시지 동작 확인

---

## 추적성

- **SPEC**: SPEC-WORKTREE-001
- **관련 파일**: spec.md, plan.md
