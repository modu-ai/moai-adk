# SPEC-TEMPLATE-001: Template Deployment & File Manifest System

---
spec_id: SPEC-TEMPLATE-001
title: Template Deployment & File Manifest System
status: Completed
priority: High
phase: "Phase 1 - Foundation (P0 Critical)"
created: 2026-02-03
depends_on:
  - SPEC-CONFIG-001
modules:
  - internal/template/
  - internal/manifest/
estimated_loc: ~1,300
adr_references:
  - ADR-007 (File Manifest Provenance)
  - ADR-011 (Zero Runtime Template Expansion)
resolves_issues:
  - "#304, #308, #309 (template)"
  - "#162, #187, #236, #246, #318, #319 (manifest)"
lifecycle: spec-anchored
tags: go-embed, template, manifest, provenance, settings-json, sha256
---

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|----------|
| 2026-02-03 | 1.0.0 | 초기 SPEC 생성 |

---

## 1. 개요

### 1.1 배경

MoAI-ADK Go Edition은 Python 전작(~73,000 LOC)에서 발생한 템플릿 배포 및 파일 관리 문제를 근본적으로 해결하기 위해 설계되었다. Python 전작에서는 다음과 같은 문제가 반복되었다:

- **템플릿 변수 치환 실패**: `{{HOOK_SHELL_PREFIX}}`, `${SHELL:-/bin/bash}` 등의 동적 토큰이 JSON 파싱 오류를 유발 (4회 회귀 사이클, 41+ 커밋)
- **파괴적 파일 덮어쓰기**: 사용자 수정 파일이 업데이트 시 무경고 삭제 (update.py 38회 수정에도 해결 불가)
- **경로 정규화 누락**: 플랫폼별 경로 구분자 불일치로 슬래시 누락 버그 발생

본 SPEC은 `internal/template/`과 `internal/manifest/` 두 모듈을 하나의 SPEC으로 통합한다. 두 모듈은 긴밀하게 결합되어 있으며, 템플릿 배포 시 매니페스트 추적이 반드시 동반되어야 하기 때문이다.

### 1.2 목표

- go:embed를 통한 ~150개 템플릿 파일(~1.1 MB)의 바이너리 내장 배포
- ADR-011(Zero Runtime Template Expansion) 준수: 모든 JSON/YAML은 Go 구조체 직렬화로 생성
- ADR-007(File Manifest Provenance) 준수: 배포된 모든 파일의 출처, 해시, 변경 상태 추적
- 12개 GitHub 이슈 해결 (#304, #308, #309, #162, #187, #236, #246, #318, #319 등)

### 1.3 범위

**포함 범위:**
- `internal/template/`: deployer.go, renderer.go, settings.go, validator.go
- `internal/manifest/`: manifest.go, hasher.go, types.go
- `.moai/manifest.json` 파일 포맷 정의
- Claude Code settings.json 프로그래밍 방식 생성

**제외 범위:**
- `internal/merge/` (3-way merge engine) -- 별도 SPEC으로 관리
- `internal/update/` (self-update system) -- 별도 SPEC으로 관리
- `internal/config/` -- SPEC-CONFIG-001에서 관리

---

## 2. 환경 (Environment)

### 2.1 시스템 환경

| 항목 | 사양 |
|------|------|
| 언어 | Go 1.22+ |
| 모듈 경로 | `github.com/modu-ai/moai-adk-go` |
| 대상 플랫폼 | darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64, windows/arm64 |
| CGO | CGO_ENABLED=0 (순수 Go) |
| 바이너리 크기 목표 | < 30MB (임베드 템플릿 포함) |

### 2.2 의존성

| 의존 모듈 | 용도 | 인터페이스 |
|-----------|------|-----------|
| `internal/config/` | 설정 로드 (Config 구조체) | `config.ConfigManager` |
| `pkg/utils/` | 파일 I/O, 경로 해석, 로거 | `utils.Logger`, `utils.File` |
| `pkg/models/` | 공유 데이터 구조체 | `models.ProjectConfig` |

### 2.3 템플릿 파일 구조

```
templates/                      # go:embed 소스
  .claude/
    settings.json.tmpl          # 플랫폼별 설정 템플릿
    agents/moai/                # 에이전트 정의
    skills/                     # 스킬 정의
    commands/moai/              # 슬래시 명령
    rules/moai/                 # 규칙 파일
    output-styles/              # 출력 스타일
  .moai/
    config/sections/            # 설정 섹션 템플릿
  CLAUDE.md.tmpl                # CLAUDE.md 템플릿
  .gitignore.tmpl               # .gitignore 템플릿
```

---

## 3. 가정 (Assumptions)

### 3.1 기술적 가정

- **A-01**: SPEC-CONFIG-001이 완료되어 `config.Config` 구조체 및 `ConfigManager` 인터페이스가 사용 가능하다.
- **A-02**: go:embed로 번들된 템플릿은 바이너리 내에서 읽기 전용이며, 런타임 수정이 불가능하다.
- **A-03**: 임베드 템플릿의 총 크기는 ~1.1 MB이며, 바이너리 크기 목표(30MB) 내에서 수용 가능하다.
- **A-04**: 대상 프로젝트 루트에 쓰기 권한이 존재한다.
- **A-05**: SHA-256 해시 충돌 확률은 무시 가능한 수준이다.

### 3.2 비즈니스 가정

- **A-06**: Python 전작의 `.moai/` 디렉토리 구조와 하위 호환성을 유지해야 한다.
- **A-07**: 기존 사용자가 수정한 파일은 절대 무경고 덮어쓰기하지 않는다.
- **A-08**: manifest.json이 없는 레거시 프로젝트는 최초 배포 시 자동으로 매니페스트를 생성한다.

---

## 4. 요구사항 (Requirements)

### 4.1 Template Module -- `internal/template/`

#### 4.1.1 Deployer (deployer.go)

**[REQ-T-001] 템플릿 추출 및 배포 (Event-Driven)**

WHEN `Deploy(ctx, projectRoot, manifest)` 호출 시 THEN go:embed 파일시스템에서 모든 템플릿을 추출하여 projectRoot에 배포하고, 배포된 각 파일을 manifest에 Track()으로 등록해야 한다.

**[REQ-T-002] 개별 템플릿 추출 (Event-Driven)**

WHEN `ExtractTemplate(name)` 호출 시 THEN 지정된 이름의 템플릿 파일 내용을 바이트 배열로 반환해야 한다. IF 해당 이름의 템플릿이 존재하지 않을 경우 THEN 적절한 오류를 반환해야 한다.

**[REQ-T-003] 템플릿 목록 조회 (Ubiquitous)**

시스템은 항상 `ListTemplates()`를 통해 임베드된 전체 템플릿 목록을 반환할 수 있어야 한다.

**[REQ-T-004] 경로 정규화 (Ubiquitous)**

시스템은 항상 배포 전 모든 파일 경로에 `filepath.Clean()`을 적용하고, 프로젝트 루트 디렉토리 내부로의 containment check를 수행해야 한다.

**[REQ-T-005] 경로 순회 방지 (Unwanted)**

시스템은 `../` 또는 절대 경로를 통한 프로젝트 루트 외부 경로 접근을 허용하지 않아야 한다.

**[REQ-T-006] 디렉토리 자동 생성 (Event-Driven)**

WHEN 배포 대상 경로의 상위 디렉토리가 존재하지 않을 경우 THEN 필요한 디렉토리를 자동으로 생성해야 한다.

**[REQ-T-007] 컨텍스트 취소 지원 (Event-Driven)**

WHEN context가 취소 또는 타임아웃될 경우 THEN 배포 작업을 즉시 중단하고 이미 배포된 파일은 롤백하지 않으나 매니페스트는 부분 배포 상태를 정확히 반영해야 한다.

#### 4.1.2 Renderer (renderer.go)

**[REQ-T-010] 템플릿 렌더링 (Event-Driven)**

WHEN `Render(templateName, data)` 호출 시 THEN Go `text/template`을 사용하여 데이터를 주입한 결과 바이트 배열을 반환해야 한다.

**[REQ-T-011] Strict Mode (Ubiquitous)**

시스템은 항상 `template.Option("missingkey=error")`를 적용하여 누락된 키에 대해 오류를 발생시켜야 한다.

**[REQ-T-012] 미확장 토큰 방지 (Unwanted)**

렌더링 결과에 `${VAR}`, `{{VAR}}`, `$VAR` 형태의 미확장 동적 토큰이 포함되어서는 안 된다.

#### 4.1.3 SettingsGenerator (settings.go)

**[REQ-T-020] settings.json 생성 (Event-Driven, ADR-011)**

WHEN `Generate(cfg, platform)` 호출 시 THEN Go 구조체(`Settings`)를 `json.MarshalIndent()`로 직렬화하여 유효한 JSON 바이트 배열을 생성해야 한다.

**[REQ-T-021] 문자열 연결 금지 (Unwanted, ADR-011)**

settings.json 생성 시 문자열 연결(`fmt.Sprintf`, `+` 연산자, `strings.Builder`)을 사용하지 않아야 한다. 반드시 `json.MarshalIndent()`를 통한 구조체 직렬화만 허용된다.

**[REQ-T-022] 플랫폼별 분기 (State-Driven)**

IF platform이 "darwin"인 경우 THEN macOS 전용 설정(예: 훅 명령 경로)을 적용해야 한다. IF platform이 "windows"인 경우 THEN Windows 전용 설정(예: cmd.exe 기반 실행)을 적용해야 한다.

**[REQ-T-023] 훅 구성 포함 (Ubiquitous)**

생성된 settings.json은 항상 SessionStart, PreToolUse, PostToolUse, SessionEnd, Stop, PreCompact 이벤트에 대한 `moai hook <event>` 명령을 포함해야 한다.

**[REQ-T-024] Settings 구조체 스키마 (Ubiquitous)**

Settings 구조체는 다음 필드를 포함해야 한다:
- `Hooks map[string][]HookGroup` -- 이벤트별 훅 그룹
- `OutputStyle string` -- 출력 스타일

HookGroup 구조체는 다음 필드를 포함해야 한다:
- `Matcher string` -- 도구 매처 패턴
- `Hooks []HookEntry` -- 훅 항목 목록

HookEntry 구조체는 다음 필드를 포함해야 한다:
- `Type string` -- 훅 타입 ("command")
- `Command string` -- 실행할 명령

#### 4.1.4 Validator (validator.go)

**[REQ-T-030] JSON 유효성 검증 (Event-Driven)**

WHEN `ValidateJSON(data)` 호출 시 THEN `json.Valid()`를 사용하여 JSON 문법 유효성을 검증하고, 무효할 경우 오류를 반환해야 한다.

**[REQ-T-031] 경로 유효성 검증 (Event-Driven)**

WHEN `ValidatePaths(projectRoot, files)` 호출 시 THEN 각 파일 경로에 대해 정규화, containment check, 특수문자 검증을 수행하고 오류 목록을 반환해야 한다.

**[REQ-T-032] 배포 후 통합 검증 (Event-Driven)**

WHEN `ValidateDeployment(projectRoot)` 호출 시 THEN 배포된 모든 파일의 존재 여부, JSON 파일의 유효성, 경로 정규화 상태, 에이전트 파일 무결성을 종합 검증하여 ValidationReport를 반환해야 한다.

**[REQ-T-033] ValidationReport 구조 (Ubiquitous)**

ValidationReport는 다음을 포함해야 한다:
- `Valid bool` -- 전체 검증 통과 여부
- `Errors []DeploymentError` -- 발견된 오류 목록
- `Warnings []string` -- 경고 목록
- `FilesChecked int` -- 검증된 파일 수

### 4.2 Manifest Module -- `internal/manifest/`

#### 4.2.1 Manager (manifest.go)

**[REQ-M-001] 매니페스트 로드 (Event-Driven)**

WHEN `Load(projectRoot)` 호출 시 THEN `{projectRoot}/.moai/manifest.json`을 읽어 Manifest 구조체로 파싱해야 한다. IF 파일이 존재하지 않을 경우 THEN 빈 매니페스트를 생성하여 반환해야 한다.

**[REQ-M-002] 매니페스트 저장 (Event-Driven)**

WHEN `Save()` 호출 시 THEN 현재 메모리 상태의 매니페스트를 `json.MarshalIndent()`로 직렬화하여 `.moai/manifest.json`에 원자적으로 기록해야 한다.

**[REQ-M-003] 원자적 쓰기 (Ubiquitous)**

매니페스트 저장 시 항상 임시 파일 작성 후 rename 방식의 원자적 쓰기를 수행하여, 쓰기 중 비정상 종료 시 기존 파일이 손상되지 않도록 해야 한다.

**[REQ-M-004] 파일 추적 등록 (Event-Driven)**

WHEN `Track(path, provenance, templateHash)` 호출 시 THEN 지정된 경로의 현재 파일 해시를 계산하고, FileEntry를 생성/갱신하여 매니페스트에 등록해야 한다.

**[REQ-M-005] 항목 조회 (Event-Driven)**

WHEN `GetEntry(path)` 호출 시 THEN 해당 경로의 FileEntry와 존재 여부를 반환해야 한다.

**[REQ-M-006] 변경 감지 (Event-Driven)**

WHEN `DetectChanges()` 호출 시 THEN 매니페스트에 등록된 모든 파일의 현재 해시를 계산하여, 기록된 해시와 다른 파일 목록(ChangedFile)을 반환해야 한다.

**[REQ-M-007] 항목 제거 (Event-Driven)**

WHEN `Remove(path)` 호출 시 THEN 매니페스트에서 해당 경로의 항목을 삭제해야 한다.

**[REQ-M-008] 손상된 매니페스트 처리 (Unwanted)**

IF 매니페스트 JSON이 파싱 불가능할 경우 THEN `ErrManifestCorrupt` 오류를 반환하고, 손상된 파일을 `.moai/manifest.json.corrupt` 백업 후 새 매니페스트를 생성해야 한다.

#### 4.2.2 Hasher (hasher.go)

**[REQ-M-010] SHA-256 해시 계산 (Event-Driven)**

WHEN 파일 해시 계산 요청 시 THEN `crypto/sha256`을 사용하여 파일 내용의 SHA-256 해시를 `sha256:` 접두사와 함께 hex 인코딩 문자열로 반환해야 한다.

**[REQ-M-011] 스트리밍 해시 (Ubiquitous)**

대용량 파일에 대해 항상 스트리밍 방식(`io.Copy` + `hash.Hash`)으로 해시를 계산하여, 전체 파일을 메모리에 로드하지 않아야 한다.

**[REQ-M-012] 파일 미존재 처리 (Event-Driven)**

WHEN 해시 대상 파일이 존재하지 않을 경우 THEN 적절한 오류를 반환해야 한다.

#### 4.2.3 Types (types.go)

**[REQ-M-020] Provenance 열거형 (Ubiquitous)**

시스템은 4개의 Provenance 타입을 정의해야 한다:
- `template_managed`: 템플릿에서 배포됨, 사용자 변경 없음. 안전하게 덮어쓰기 가능.
- `user_modified`: 템플릿 기반 + 사용자 편집 감지. 3-way merge 필요.
- `user_created`: 사용자 고유 파일, 템플릿과 무관. 절대 수정 금지.
- `deprecated`: 새 템플릿 버전에서 제거됨. 사용자에게 알림, 파일 유지.

**[REQ-M-021] Manifest 구조체 (Ubiquitous)**

Manifest 구조체는 다음 필드를 포함해야 한다:
- `Version string` -- ADK 버전
- `DeployedAt string` -- 배포 시간 (ISO 8601)
- `Files map[string]FileEntry` -- 경로별 파일 항목

**[REQ-M-022] FileEntry 구조체 (Ubiquitous)**

FileEntry 구조체는 다음 필드를 포함해야 한다:
- `Provenance Provenance` -- 파일 출처 분류
- `TemplateHash string` -- 원본 템플릿의 SHA-256 해시
- `DeployedHash string` -- 배포 시점의 SHA-256 해시
- `CurrentHash string` -- 마지막 확인 시점의 SHA-256 해시

**[REQ-M-023] ChangedFile 구조체 (Ubiquitous)**

ChangedFile 구조체는 다음 필드를 포함해야 한다:
- `Path string` -- 변경된 파일 경로
- `OldHash string` -- 매니페스트 기록 해시
- `NewHash string` -- 현재 파일 해시
- `Provenance Provenance` -- 파일 출처 분류

---

## 5. 사양 (Specifications)

### 5.1 인터페이스 정의

#### Template Module

```go
// internal/template/deployer.go
type Deployer interface {
    Deploy(ctx context.Context, projectRoot string, manifest manifest.Manager) error
    ExtractTemplate(name string) ([]byte, error)
    ListTemplates() []string
}

// internal/template/renderer.go
type Renderer interface {
    Render(templateName string, data interface{}) ([]byte, error)
}

// internal/template/settings.go
type SettingsGenerator interface {
    Generate(cfg *config.Config, platform string) ([]byte, error)
}

// internal/template/validator.go
type Validator interface {
    ValidateJSON(data []byte) error
    ValidatePaths(projectRoot string, files []string) []PathError
    ValidateDeployment(projectRoot string) *ValidationReport
}
```

#### Manifest Module

```go
// internal/manifest/manifest.go
type Manager interface {
    Load(projectRoot string) (*Manifest, error)
    Save() error
    Track(path string, provenance Provenance, templateHash string) error
    GetEntry(path string) (*FileEntry, bool)
    DetectChanges() ([]ChangedFile, error)
    Remove(path string) error
}

// internal/manifest/types.go
type Provenance string

const (
    TemplateManaged Provenance = "template_managed"
    UserModified    Provenance = "user_modified"
    UserCreated     Provenance = "user_created"
    Deprecated      Provenance = "deprecated"
)
```

#### Settings 구조체 (ADR-011)

```go
// internal/template/settings.go
type Settings struct {
    Hooks       map[string][]HookGroup `json:"hooks,omitempty"`
    OutputStyle string                 `json:"output_style,omitempty"`
}

type HookGroup struct {
    Matcher string      `json:"matcher,omitempty"`
    Hooks   []HookEntry `json:"hooks"`
}

type HookEntry struct {
    Type    string `json:"type"`
    Command string `json:"command"`
}
```

### 5.2 데이터 모델 -- manifest.json

```json
{
  "version": "1.14.0",
  "deployed_at": "2026-02-03T10:30:00Z",
  "files": {
    ".claude/agents/moai/expert-backend.md": {
      "provenance": "template_managed",
      "template_hash": "sha256:a1b2c3d4...",
      "deployed_hash": "sha256:a1b2c3d4...",
      "current_hash": "sha256:a1b2c3d4..."
    },
    "CLAUDE.md": {
      "provenance": "user_modified",
      "template_hash": "sha256:e5f6g7h8...",
      "deployed_hash": "sha256:e5f6g7h8...",
      "current_hash": "sha256:i9j0k1l2..."
    }
  }
}
```

### 5.3 오류 타입

```go
// internal/manifest/ 오류
var (
    ErrManifestNotFound = errors.New("manifest: file not found")
    ErrManifestCorrupt  = errors.New("manifest: JSON parse error")
    ErrEntryNotFound    = errors.New("manifest: entry not found")
    ErrHashMismatch     = errors.New("manifest: hash verification failed")
)

// internal/template/ 오류
var (
    ErrTemplateNotFound    = errors.New("template: not found in embedded filesystem")
    ErrPathTraversal       = errors.New("template: path traversal detected")
    ErrInvalidJSON         = errors.New("template: generated JSON is invalid")
    ErrUnexpandedToken     = errors.New("template: unexpanded dynamic token detected")
    ErrMissingTemplateKey  = errors.New("template: missing key in template data")
)

// internal/template/validator.go
type PathError struct {
    Path   string `json:"path"`
    Reason string `json:"reason"`
}

type DeploymentError struct {
    Path   string `json:"path"`
    Reason string `json:"reason"`
}

type ValidationReport struct {
    Valid        bool              `json:"valid"`
    Errors       []DeploymentError `json:"errors,omitempty"`
    Warnings     []string          `json:"warnings,omitempty"`
    FilesChecked int               `json:"files_checked"`
}
```

### 5.4 의존성 그래프

```
internal/template/ --> internal/manifest/   (Track 호출)
                   --> internal/config/     (Config 읽기)
                   --> pkg/utils/           (파일 I/O, 경로)

internal/manifest/ --> pkg/utils/           (파일 I/O)

외부 소비자:
internal/update/   --> internal/template/   (새 템플릿 배포)
                   --> internal/manifest/   (매니페스트 갱신)
internal/core/project/ --> internal/template/ (초기화 시 배포)
```

### 5.5 성능 요구사항

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| 전체 템플릿 배포 (~150 파일) | < 500ms | 벤치마크 테스트 |
| 단일 파일 해시 계산 | < 5ms | 벤치마크 테스트 |
| 매니페스트 로드/파싱 | < 10ms | 벤치마크 테스트 |
| 매니페스트 저장 | < 10ms | 벤치마크 테스트 |
| settings.json 생성 | < 5ms | 벤치마크 테스트 |
| 변경 감지 (150 파일) | < 1s | 벤치마크 테스트 |
| 메모리 사용량 (배포 중) | < 50MB | 프로파일링 |

### 5.6 보안 요구사항

| 항목 | 구현 방식 |
|------|----------|
| 경로 순회 방지 | `filepath.Clean()` + 프로젝트 루트 containment check |
| JSON 주입 방지 | `json.MarshalIndent()` 구조체 직렬화만 허용 |
| 템플릿 확장 금지 | 생성된 파일에 `${VAR}`, `{{VAR}}` 토큰 부재 검증 |
| 경로 정규화 | 모든 경로에 `filepath.Clean()` 적용 후 배포 |
| 원자적 파일 쓰기 | 임시 파일 + rename 패턴 |
| 해시 알고리즘 | SHA-256 (crypto/sha256) |

### 5.7 ADR 준수 사항

**ADR-007 (File Manifest Provenance):**
- 배포된 모든 파일은 매니페스트에 등록되어야 한다
- 4개 Provenance 타입에 따른 업데이트 동작 차별화
- 3-way merge를 위한 3중 해시 (template_hash, deployed_hash, current_hash) 기록

**ADR-011 (Zero Runtime Template Expansion):**
- settings.json은 `json.MarshalIndent()`로만 생성
- 어떤 설정 파일에도 미확장 동적 토큰($VAR, {{VAR}}, ${SHELL})이 존재해서는 안 됨
- CLAUDE.md는 Go `text/template` strict mode (`missingkey=error`)로 렌더링
- 생성 후 `json.Valid()`로 JSON 유효성 검증

---

## 6. 크로스 레퍼런스

| 참조 | 위치 | 관련 내용 |
|------|------|----------|
| ADR-007 | structure.md, design.md | File Manifest Provenance |
| ADR-011 | structure.md, design.md | Zero Runtime Template Expansion |
| ADR-003 | tech.md | go:embed for Template Distribution |
| ADR-008 | structure.md, design.md | 3-Way Merge (manifest 활용) |
| SPEC-CONFIG-001 | .moai/specs/ | Config 모듈 의존성 |
| product.md | .moai/project/ | 성공 지표, 경쟁 우위 |

---

## 7. 전문가 컨설팅 권장

본 SPEC은 다음 도메인 전문가 컨설팅을 권장한다:

- **expert-backend**: Go 인터페이스 설계, go:embed 최적화, 동시성 안전성 검토
- **expert-security**: 경로 순회 방지, 원자적 파일 쓰기, SHA-256 해시 무결성 검증
- **expert-testing**: 테이블 기반 테스트, 퍼즈 테스트 전략, 계약 테스트 설계

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 85.7%

### Summary

Template deployment system implemented with go:embed-based template bundling, Go template rendering with custom functions, atomic file writing with temp+rename pattern, and SHA-256 hash-based integrity verification. Manifest module tracks file provenance (template_managed, user_modified, user_created, deprecated) and detects changes for safe re-deployment. Settings generator produces Claude Code settings.json with hook configurations.

### Files Created

- `internal/template/deployer.go`
- `internal/template/deployer_test.go`
- `internal/template/errors.go`
- `internal/template/renderer.go`
- `internal/template/renderer_test.go`
- `internal/template/settings.go`
- `internal/template/settings_test.go`
- `internal/template/validator.go`
- `internal/template/validator_test.go`
- `internal/manifest/hasher.go`
- `internal/manifest/hasher_test.go`
- `internal/manifest/manifest.go`
- `internal/manifest/manifest_test.go`
- `internal/manifest/types.go`
- `internal/manifest/types_test.go`
