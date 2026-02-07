# 기여 가이드

MoAI-ADK Go Edition 프로젝트에 기여해 주셔서 감사합니다! 이 문서는 프로젝트에 효과적으로 기여하는 방법을 안내합니다.

## 목차

- [시작하기](#시작하기)
- [개발 환경 설정](#개발-환경-설정)
- [코드 스타일](#코드-스타일)
- [테스트 작성](#테스트-작성)
- [커밋 규칙](#커밋-규칙)
- [풀 리퀘스트](#풀-리퀘스트)

---

## 시작하기

### 1. 저장소 포크 및 클론

```bash
# GitHub에서 저장소 포크
# 로컬에 클론
git clone https://github.com/YOUR_USERNAME/moai-adk-go.git
cd moai-adk-go

# 업스트림 원격 저장소 추가
git remote add upstream https://github.com/modu-ai/moai-adk-go.git
```

### 2. 개발 브랜치 생성

```bash
git checkout -b feature/your-feature-name
```

브랜치 네이밍 규칙:
- `feature/` - 새로운 기능 추가
- `fix/` - 버그 수정
- `docs/` - 문서 업데이트
- `refactor/` - 코드 리팩토링
- `test/` - 테스트 추가 또는 수정

---

## 개발 환경 설정

### 필수 도구

- **Go 1.25+**: [golang.org](https://go.dev/dl/)에서 설치
- **golangci-lint**: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **gofumpt**: `go install mvdan.cc/gofumpt@latest`

### 개발 워크플로우

```bash
# 1. 의존성 설치
go mod download

# 2. 테스트 실행
make test

# 3. 린팅 확인
make lint

# 4. 코드 포맷팅
make fmt

# 5. 빌드
make build
```

---

## 코드 스타일

### Go 규칙

모든 코드는 [Effective Go](https://go.dev/doc/effective_go)와 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)를 따라야 합니다.

### 프로젝트 특정 규칙

1. **패키지 이름**: 소문자, 단일 단어 (예: `template`, `cli`)
2. **내보내기**: PascalCase (예: `TemplateDeployer`, `NewManager`)
3. **비내보내기**: camelCase (예: `deployTemplates`, `localConfig`)
4. **상수**: PascalCase 또는 UPPER_SNAKE_CASE
5. **인터페이스**: `-er`로 끝나는 이름 (예: `Reader`, `Writer`)

### 주석

```go
// Package template provides template deployment and rendering...
package template

// Deployer extracts and deploys templates from an embedded filesystem...
type Deployer interface { ... }

// NewDeployer creates a new Deployer with the given options...
func NewDeployer(opts ...Option) (Deployer, error) { ... }
```

### 에러 처리

```go
// 에러 래핑 - 항상 컨텍스트 포함
if err != nil {
    return fmt.Errorf("deploy templates: %w", err)
}

// 에러 래핑 - 경로 정보 포함
return fmt.Errorf("read %q: %w", path, err)  // GOOD
return fmt.Errorf("read " + path + ": " + err.Error())  // BAD
```

---

## 테스트 작성

### 테스트 커버리지 요구사항

- 모든 패키지는 **최소 85%** 커버리지 유지
- 새로운 코드는 **TDD 방식**으로 작성 (RED-GREEN-REFACTOR)
- 기존 코드 수정은 **DDD 방식**으로 접근 (ANALYZE-PRESERVE-IMPROVE)

### 테이블 기반 테스트 (권장)

```go
func TestBuildRequiredPATH(t *testing.T) {
    tests := []struct {
        name    string
        goBin   string
        goPath  string
        want    string
    }{
        {"default", "", "", wantDefault},
        {"custom bin", "/custom/bin", "", wantCustom},
        {"custom path", "", "/custom/path", wantPath},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 테스트 구현
        })
    }
}
```

### Race Detection

```bash
# 항상 race detection으로 테스트 실행
go test -race ./...

# Makefile 사용
make test
```

---

## 커밋 규칙

### 커밋 메시지 형식

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Type 카테고리

- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서만 변경
- `style`: 코드 포맷팅, 세미콜론 누락 등 (코드 동작에 영향 없음)
- `refactor`: 코드 리팩토링 (새 기능이나 버그 수정 아님)
- `perf`: 성능 개선
- `test`: 테스트 추가 또는 수정
- `chore`: 빌드 프로세스, 도구 업데이트 등

### 예시

```
feat(cli): add interactive wizard for project initialization

The new wizard guides users through:
- Development mode selection (DDD/TDD/Hybrid)
- Language preference configuration
- Quality gates customization

Closes #123
```

```
fix(update): correct version comparison for go-v prefix

The update checker was failing to compare versions starting with
"go-v" prefix. This fix normalizes version strings before comparison.

Fixes #456
```

---

## 풀 리퀘스트

### PR 제출 전 체크리스트

- [ ] 모든 테스트 통과 (`make test`)
- [ ] 0개의 린트 오류 (`make lint`)
- [ ] 코드 포맷팅 완료 (`make fmt`)
- [ ] 85%+ 테스트 커버리지
- [ ] Race detection 통과
- [ ] 커밋 메시지가 컨벤셔널 커밋 형식 따름
- [ ] PR 설명에 변경사항 요약 포함

### PR 템플릿

```markdown
## 설명
이 PR은 [기능/버그 수정]을 구현합니다.

## 변경사항
- 변경사항 1
- 변경사항 2

## 테스트
- 테스트 방법 1
- 테스트 방법 2

## 체크리스트
- [x] 모든 테스트 통과
- [x] 린팅 통과
- [x] 문서 업데이트
```

---

## 행동 강령

- 존중과 포용성을 유지하세요
- 건설적인 피드백을 제공하세요
- 다른 기여자를 돕고 지원하세요
- 행동 강령 위반 시 [report@modu-ai](mailto:report@modu-ai)로 신고하세요

---

## 라이선스

기여하신 코드는 프로젝트의 [Copyleft 3.0 라이선스](LICENSE)에 따라 배포됩니다.

---

질문이 있으시면 [GitHub Issues](https://github.com/modu-ai/moai-adk-go/issues)를 생성해 주세요.
