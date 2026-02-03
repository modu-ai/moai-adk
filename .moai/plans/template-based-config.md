# 템플릿 기반 설정 파일 생성 시스템 전환

## 개요

현재 Go 구조체 직렬화 방식을 **템플릿 + 변수 치환** 방식으로 전환합니다.

**목표**: 설정 필드 추가/수정 시 Go 코드 변경 없이 템플릿 파일만 편집

---

## 대상 파일

| 파일 | 현재 방식 | 전환 후 |
|------|-----------|---------|
| `.moai/config/sections/user.yaml` | Go struct → YAML | 템플릿 렌더링 |
| `.moai/config/sections/language.yaml` | Go struct → YAML | 템플릿 렌더링 |
| `.moai/config/sections/quality.yaml` | Go struct → YAML | 템플릿 렌더링 |
| `.moai/config/sections/workflow.yaml` | Go struct → YAML | 템플릿 렌더링 |
| `.claude/settings.json` | SettingsGenerator | **유지** (플랫폼 로직) |
| `.mcp.json` | MCPGenerator | **유지** (플랫폼 로직) |

**settings.json, .mcp.json은 플랫폼별 로직(hooks, MCP 커맨드)이 필요하므로 Go 생성 유지**

---

## TemplateContext 데이터 모델

```go
// internal/template/context.go
type TemplateContext struct {
    // 프로젝트
    ProjectName string
    ProjectRoot string

    // 사용자
    UserName string

    // 언어 설정
    ConversationLanguage     string // "ko", "en"
    ConversationLanguageName string // "Korean (한국어)"
    AgentPromptLanguage      string // "en"
    GitCommitMessages        string // "en"
    CodeComments             string // "en"
    Documentation            string // "en"
    ErrorMessages            string // "en"

    // 개발 설정
    DevelopmentMode    string // "ddd", "tdd", "hybrid"
    EnforceQuality     bool   // true
    TestCoverageTarget int    // 85

    // 워크플로우
    AutoClear  bool // true
    PlanTokens int  // 30000
    RunTokens  int  // 180000
    SyncTokens int  // 40000

    // 메타
    Version  string // MoAI-ADK 버전
    Platform string // "darwin", "linux", "windows"
}
```

---

## 템플릿 파일 구조

**위치**: `internal/template/templates/.moai/config/sections/`

### user.yaml.tmpl
```yaml
user:
  name: "{{.UserName}}"
```

### language.yaml.tmpl
```yaml
language:
  conversation_language: {{.ConversationLanguage}}
  conversation_language_name: {{.ConversationLanguageName}}
  agent_prompt_language: {{.AgentPromptLanguage}}
  git_commit_messages: {{.GitCommitMessages}}
  code_comments: {{.CodeComments}}
  documentation: {{.Documentation}}
  error_messages: {{.ErrorMessages}}
```

### quality.yaml.tmpl
```yaml
constitution:
  development_mode: {{.DevelopmentMode}}
  enforce_quality: {{.EnforceQuality}}
  test_coverage_target: {{.TestCoverageTarget}}
  # ... 나머지 정적 설정
```

### workflow.yaml.tmpl
```yaml
workflow:
  auto_clear: {{.AutoClear}}
  plan_tokens: {{.PlanTokens}}
  run_tokens: {{.RunTokens}}
  sync_tokens: {{.SyncTokens}}
```

---

## 렌더링 흐름

```
moai init
    ↓
Step 1-3: 디렉토리 생성 (기존과 동일)
    ↓
Step 4: deployTemplates()
    ├─ Deployer.Deploy(ctx, root, manifest, templateContext)
    ├─ WalkDir로 모든 파일 순회
    ├─ .tmpl 파일 감지 → Renderer.Render() 호출
    ├─ 결과 파일명에서 .tmpl 제거
    └─ 렌더링된 내용을 디스크에 기록
    ↓
Step 5: generateSettings(), generateMCPConfig() (기존과 동일)
    ↓
Step 6-7: CLAUDE.md, manifest (기존과 동일)
```

---

## 코드 변경 사항

### 1. 새 파일: `internal/template/context.go`
- `TemplateContext` 구조체
- `NewTemplateContext(opts InitOptions) *TemplateContext`
- `resolveLanguageName()` (initializer.go에서 이동)

### 2. 수정: `internal/template/deployer.go`
```go
type deployer struct {
    fsys     fs.FS
    renderer Renderer  // 추가
}

func NewDeployer(fsys fs.FS, renderer Renderer) Deployer

func (d *deployer) Deploy(ctx, root, manifest, tmplCtx *TemplateContext) error {
    // .tmpl 파일 감지 및 렌더링 로직 추가
}
```

### 3. 수정: `internal/core/project/initializer.go`
- 삭제: `generateConfigs()`, YAML 구조체들, `writeYAML()`
- 수정: `deployTemplates()`에서 TemplateContext 생성 후 Deployer에 전달

### 4. 수정: `internal/cli/init.go`
- NewDeployer 호출 시 Renderer 전달

### 5. 템플릿 파일 생성
- `templates/.moai/config/sections/user.yaml.tmpl`
- `templates/.moai/config/sections/language.yaml.tmpl`
- `templates/.moai/config/sections/quality.yaml.tmpl`
- `templates/.moai/config/sections/workflow.yaml.tmpl`

---

## 구현 순서

1. **Phase 1**: context.go 생성 (TemplateContext + NewTemplateContext)
2. **Phase 2**: deployer.go 수정 (Renderer 통합, .tmpl 처리)
3. **Phase 3**: 템플릿 파일 4개 생성
4. **Phase 4**: initializer.go에서 generateConfigs() 제거, deployTemplates() 수정
5. **Phase 5**: init.go 수정 (Renderer 주입)
6. **Phase 6**: 테스트 업데이트 및 검증

---

## 검증 방법

1. `go build ./...` - 빌드 성공
2. `go test ./...` - 모든 테스트 통과
3. `moai init --force` 실행 후:
   - `.moai/config/sections/*.yaml` 파일 생성 확인
   - YAML 파싱 유효성 검증
   - 변수 값 치환 확인 (예: UserName이 올바르게 들어갔는지)
4. `.claude/settings.json`, `.mcp.json` 기존대로 생성되는지 확인

---

## 핵심 파일 경로

- `/Users/goos/MoAI/moai-adk-go/internal/template/deployer.go`
- `/Users/goos/MoAI/moai-adk-go/internal/template/renderer.go`
- `/Users/goos/MoAI/moai-adk-go/internal/template/context.go` (신규)
- `/Users/goos/MoAI/moai-adk-go/internal/core/project/initializer.go`
- `/Users/goos/MoAI/moai-adk-go/internal/cli/init.go`
- `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.moai/config/sections/*.yaml.tmpl` (신규)
