# SPEC-WORKTREE-001: 구현 계획

---
spec_id: SPEC-WORKTREE-001
type: plan
---

## 개요

`moai worktree new` 명령의 기본 worktree 생성 경로를 프로젝트 내부(`.moai/worktrees/`)에서 글로벌(`~/.moai/worktrees/{ProjectName}/`)로 마이그레이션한다. 이를 통해 git worktree 설계 원칙에 부합하고, IDE 간섭과 중첩 구조 문제를 해소한다.

---

## 마일스톤

### Primary Goal: 경로 결정 로직 변경

핵심 Go 코드 변경으로 worktree 생성 경로를 글로벌로 전환한다.

**태스크 분해:**

1. **`detectProjectName()` 함수 구현**
   - 파일: `internal/cli/worktree/new.go` (또는 `internal/cli/worktree/project.go` 신규)
   - `go.mod` module path 파싱 -> git remote origin 파싱 -> `filepath.Base(cwd)` fallback
   - 단위 테스트 작성 (table-driven, 3가지 fallback 시나리오)

2. **`runNew()` 경로 로직 수정**
   - 파일: `internal/cli/worktree/new.go:41-47`
   - `os.UserHomeDir()` + `detectProjectName()` 조합으로 글로벌 경로 생성
   - `os.MkdirAll`로 디렉토리 자동 생성
   - SPEC-ID와 일반 브랜치명 모두 동일 패턴 적용

3. **하위 호환성 경고 추가**
   - 기존 `.moai/worktrees/` 디렉토리에 항목이 있으면 stderr 경고
   - 기존 worktree 동작은 방해하지 않음

4. **기존 테스트 수정**
   - `subcommands_test.go`: 기댓값을 글로벌 경로로 변경
   - 테스트에서 `os.UserHomeDir()`를 `t.TempDir()`로 mocking 필요

### Secondary Goal: 관련 모듈 정합성

worktree 경로를 참조하는 다른 모듈의 정합성을 확보한다.

**태스크 분해:**

5. **`worktree_orchestrator.go` 검증**
   - `filepath.Base(wt.Path)` 동작 검증 (글로벌 경로에서도 SPEC-ID 추출 확인)
   - 필요 시 수정, 불필요하면 테스트 추가만

6. **`launcher.go` cleanup 경로 수정**
   - `cleanupMoaiWorktrees` 함수의 탐색 경로 확장
   - `.claude/worktrees` + `~/.moai/worktrees/{ProjectName}/` 양쪽 모두 정리

7. **hook 테스트 데이터 수정**
   - `worktree_create_test.go`: 경로 데이터 업데이트
   - `worktree_remove_test.go`: 경로 데이터 업데이트

### Final Goal: 문서 및 템플릿 동기화

문서와 템플릿의 경로 표기를 새 경로로 통일한다.

**태스크 분해:**

8. **규칙/스킬 문서 수정**
   - `worktree-integration.md`: 경로 표기 전체 수정
   - `SKILL.md`: 경로 예시 수정
   - `registry-architecture.md`: 레지스트리 경로 수정
   - `CLAUDE.local.md`: Section 16 worktree 설명 수정

9. **템플릿 동기화**
   - `internal/template/templates/` 하위 동일 파일 수정
   - `make build` 실행하여 `embedded.go` 재생성

10. **전체 테스트 스위트 실행**
    - `go test -race ./...` 실행
    - `golangci-lint run` 실행

### Optional Goal: 마이그레이션 서브커맨드

기존 프로젝트 내부 worktree를 글로벌 경로로 자동 이동하는 유틸리티.

**태스크 분해:**

11. **`moai worktree migrate` 서브커맨드** (선택적)
    - `.moai/worktrees/` 스캔 -> `~/.moai/worktrees/{ProjectName}/`으로 `git worktree move`
    - 레지스트리 경로 업데이트
    - dry-run 모드 지원

---

## 기술적 접근

### 프로젝트명 감지 전략

```
우선순위:
1. go.mod module path -> filepath.Base(modulePath)
   예: "github.com/modu-ai/moai-adk" -> "moai-adk"

2. git remote get-url origin -> 리포지토리명 추출
   예: "git@github.com:modu-ai/moai-adk-go.git" -> "moai-adk-go"

3. filepath.Base(cwd) -> 현재 디렉토리명
   예: "/Users/goos/MoAI/moai-adk-go" -> "moai-adk-go"
```

### 경로 생성 흐름

```
사용자 입력: moai worktree new SPEC-AUTH-001

1. isSpecID("SPEC-AUTH-001") -> true
2. branchName = "feature/SPEC-AUTH-001"
3. homeDir = os.UserHomeDir() -> "/Users/goos"
4. projectName = detectProjectName() -> "moai-adk"
5. wtPath = filepath.Join(homeDir, ".moai", "worktrees", projectName, "SPEC-AUTH-001")
   결과: "/Users/goos/.moai/worktrees/moai-adk/SPEC-AUTH-001"
6. os.MkdirAll(filepath.Dir(wtPath), 0o755)
7. WorktreeProvider.Add(wtPath, branchName)
```

### 테스트 전략

- **단위 테스트**: `detectProjectName()` 함수의 3가지 fallback 시나리오
- **통합 테스트**: `runNew()` 함수의 경로 생성 결과 검증 (`t.TempDir()` 기반)
- **기존 테스트 수정**: 경로 기댓값을 글로벌 경로로 변경, `UserHomeDir` mocking
- **table-driven**: 모든 새 테스트는 Go table-driven 패턴 사용

### 위험 및 대응

| 위험 | 확률 | 영향 | 대응 |
|------|------|------|------|
| `os.UserHomeDir()` 실패 (컨테이너 환경) | 낮음 | 높음 | fallback으로 프로젝트 내부 경로 사용 + 경고 |
| `go.mod` 없는 프로젝트에서 실행 | 중간 | 낮음 | git remote -> cwd basename fallback 체인 |
| 기존 `.moai/worktrees/` 사용자의 혼란 | 중간 | 중간 | stderr 경고 메시지 + 문서 안내 |
| Windows 환경에서 경로 호환성 | 낮음 | 중간 | `filepath.Join` 사용으로 자동 처리 |
| 글로벌 디렉토리 권한 문제 | 낮음 | 높음 | `os.MkdirAll` 에러 시 명확한 메시지 반환 |

### 아키텍처 설계 방향

- **단일 책임**: 경로 결정 로직을 `resolveWorktreePath()` 함수로 분리
- **테스트 용이성**: `os.UserHomeDir()`를 함수 변수로 추출하여 테스트 시 override 가능
- **점진적 변경**: 코드 변경 -> 테스트 수정 -> 문서 수정 -> 템플릿 동기화 순서
- **하위 호환성**: 기존 worktree는 동작을 방해하지 않고 경고만 출력

---

## 의존성

- 선행 작업 없음 (독립적 변경)
- 후행 작업: Optional Goal의 `migrate` 서브커맨드는 별도 SPEC으로 분리 가능
- 빌드 의존성: 문서/템플릿 변경 후 `make build` 필수

---

## 추적성

- **SPEC**: SPEC-WORKTREE-001
- **관련 파일**: spec.md, acceptance.md
