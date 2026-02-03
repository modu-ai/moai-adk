# SPEC-TEMPLATE-001: Implementation Plan

---
spec_id: SPEC-TEMPLATE-001
type: plan
created: 2026-02-03
status: Planned
---

## 1. 구현 전략

### 1.1 접근 방식

Bottom-up 순서로 구현한다. 의존성이 없는 기반 타입부터 시작하여, 해시 유틸리티, 매니페스트 관리자, 템플릿 배포자, 설정 생성기, 검증기 순서로 진행한다. 각 단계는 이전 단계의 인터페이스에 의존하므로 순차적 구현이 필수적이다.

### 1.2 DDD 사이클 적용

ANALYZE-PRESERVE-IMPROVE 사이클에 따라:

- **ANALYZE**: Python 전작의 template/, manifest 관련 코드 분석 및 12개 이슈의 근본 원인 이해
- **PRESERVE**: Go 인터페이스 기반 계약을 먼저 정의하고, 각 인터페이스에 대한 특성화 테스트 작성
- **IMPROVE**: ADR-007/ADR-011 원칙에 따른 구현체 개발

---

## 2. 태스크 분해

### Milestone 1: Manifest Types & Hasher (Primary Goal)

기반 타입과 해시 유틸리티를 먼저 구현한다. 다른 모든 컴포넌트의 의존성이다.

**Task 1.1: Provenance 열거형 및 데이터 구조체 정의**
- 파일: `internal/manifest/types.go`
- 내용:
  - `Provenance` 타입 (string 기반) 및 4개 상수 정의
  - `Manifest` 구조체 (Version, DeployedAt, Files)
  - `FileEntry` 구조체 (Provenance, TemplateHash, DeployedHash, CurrentHash)
  - `ChangedFile` 구조체 (Path, OldHash, NewHash, Provenance)
  - JSON 태그 설정 및 godoc 주석
- 요구사항: REQ-M-020, REQ-M-021, REQ-M-022, REQ-M-023
- 테스트: JSON 직렬화/역직렬화 라운드트립 테스트, Provenance 문자열 값 검증

**Task 1.2: SHA-256 해시 유틸리티 구현**
- 파일: `internal/manifest/hasher.go`
- 내용:
  - `HashFile(path string) (string, error)` 함수
  - `HashBytes(data []byte) string` 함수
  - 스트리밍 방식 해시 (`io.Copy` + `crypto/sha256`)
  - `sha256:` 접두사 포맷
- 요구사항: REQ-M-010, REQ-M-011, REQ-M-012
- 테스트: 알려진 해시 값 검증, 빈 파일, 대용량 파일(10MB+), 존재하지 않는 파일

**Task 1.3: 오류 타입 정의**
- 파일: `internal/manifest/types.go` (추가)
- 내용:
  - `ErrManifestNotFound`, `ErrManifestCorrupt`, `ErrEntryNotFound`, `ErrHashMismatch`
  - 에러 메시지에 컨텍스트 포함 (`fmt.Errorf` wrapping)
- 테스트: `errors.Is()` 체인 검증

---

### Milestone 2: Manifest Manager (Primary Goal)

매니페스트 CRUD 및 변경 감지를 구현한다.

**Task 2.1: Manifest Manager 인터페이스 구현**
- 파일: `internal/manifest/manifest.go`
- 내용:
  - `Manager` 인터페이스 구현체 (`manifestManager` 구조체)
  - `NewManager() Manager` 생성자
  - 내부 상태: `projectRoot string`, `manifest *Manifest`, `hasher` 참조
- 요구사항: REQ-M-001 ~ REQ-M-008

**Task 2.2: Load/Save 구현**
- 파일: `internal/manifest/manifest.go`
- 내용:
  - `Load()`: JSON 파싱, 파일 미존재 시 빈 매니페스트 생성
  - `Save()`: `json.MarshalIndent()` + 원자적 쓰기 (임시파일 + rename)
  - 손상된 JSON 처리: `.corrupt` 백업 후 새 매니페스트 생성
- 요구사항: REQ-M-001, REQ-M-002, REQ-M-003, REQ-M-008
- 테스트:
  - 정상 로드/저장 라운드트립
  - 파일 미존재 시 빈 매니페스트
  - 손상된 JSON 처리 (`.corrupt` 백업 검증)
  - 원자적 쓰기 검증 (쓰기 중 비정상 종료 시뮬레이션)

**Task 2.3: Track/GetEntry/Remove 구현**
- 파일: `internal/manifest/manifest.go`
- 내용:
  - `Track()`: 현재 해시 계산 + FileEntry 생성/갱신
  - `GetEntry()`: 맵 조회 + 존재 여부 반환
  - `Remove()`: 맵에서 항목 삭제
- 요구사항: REQ-M-004, REQ-M-005, REQ-M-007
- 테스트: CRUD 시나리오, 중복 Track, 미존재 항목 Remove

**Task 2.4: DetectChanges 구현**
- 파일: `internal/manifest/manifest.go`
- 내용:
  - 매니페스트 전체 파일 순회
  - 현재 해시 계산 후 기록 해시와 비교
  - ChangedFile 목록 반환
  - 삭제된 파일 처리 (해시 계산 불가)
- 요구사항: REQ-M-006
- 테스트: 변경/미변경/삭제된 파일 혼합 시나리오

---

### Milestone 3: Template Deployer (Secondary Goal)

go:embed 기반 템플릿 추출 및 배포를 구현한다.

**Task 3.1: go:embed 파일시스템 설정**
- 파일: `internal/template/deployer.go`
- 내용:
  - `//go:embed templates/*` 디렉티브
  - `embed.FS` 변수 선언
  - `ListTemplates()`: `fs.WalkDir`로 전체 목록 반환
  - `ExtractTemplate()`: 단일 파일 읽기
- 요구사항: REQ-T-002, REQ-T-003

**Task 3.2: Deploy 로직 구현**
- 파일: `internal/template/deployer.go`
- 내용:
  - `Deploy()`: 전체 템플릿 순회 및 배포
  - `filepath.Clean()` 경로 정규화
  - containment check (프로젝트 루트 내부 확인)
  - 상위 디렉토리 자동 생성 (`os.MkdirAll`)
  - 각 파일 배포 후 `manifest.Track()` 호출
  - `context.Context` 취소 지원
- 요구사항: REQ-T-001, REQ-T-004, REQ-T-005, REQ-T-006, REQ-T-007
- 테스트:
  - 정상 배포 (파일 존재 + 매니페스트 등록 검증)
  - 경로 순회 시도 (`../etc/passwd` 등)
  - 컨텍스트 취소 시 부분 배포 상태 검증
  - 디렉토리 자동 생성 검증

**Task 3.3: 경로 보안 유틸리티**
- 파일: `internal/template/deployer.go` (내부 함수)
- 내용:
  - `sanitizePath()`: `filepath.Clean()` + `..` 검출
  - `isContained()`: 정규화된 경로가 프로젝트 루트 하위인지 검증
- 요구사항: REQ-T-004, REQ-T-005
- 테스트: 다양한 경로 순회 벡터 (symlink, `..\`, `//`, 절대 경로)

---

### Milestone 4: Settings Generator (Secondary Goal)

ADR-011을 준수하는 settings.json 생성기를 구현한다.

**Task 4.1: Settings 구조체 정의**
- 파일: `internal/template/settings.go`
- 내용:
  - `Settings`, `HookGroup`, `HookEntry` 구조체
  - JSON 태그, omitempty 설정
  - `OutputStyle` 관련 구조체
- 요구사항: REQ-T-024

**Task 4.2: Generate 함수 구현**
- 파일: `internal/template/settings.go`
- 내용:
  - `Generate(cfg, platform)`: Config에서 Settings 구조체 조립
  - 플랫폼별 훅 명령 분기 (darwin/linux/windows)
  - `json.MarshalIndent()` 직렬화
  - `json.Valid()` 후처리 검증
- 요구사항: REQ-T-020, REQ-T-021, REQ-T-022, REQ-T-023
- 테스트:
  - JSON 라운드트립 (Marshal -> Valid -> Unmarshal -> 비교)
  - 플랫폼별 출력 차이 검증 (darwin, linux, windows)
  - 모든 훅 이벤트 포함 검증
  - 문자열 연결 미사용 검증 (코드 리뷰 + AST 분석)

**Task 4.3: 플랫폼별 훅 설정 조립**
- 파일: `internal/template/settings.go`
- 내용:
  - `buildHooks(platform)`: 플랫폼에 따른 HookGroup 구성
  - SessionStart, PreToolUse, PostToolUse, SessionEnd, Stop, PreCompact 이벤트
  - darwin/linux: `moai hook <event>` 직접 실행
  - windows: `cmd.exe /c moai hook <event>` 래핑
- 요구사항: REQ-T-022, REQ-T-023
- 테스트: 각 플랫폼별 명령 문자열 정확성 검증

---

### Milestone 5: Renderer (Secondary Goal)

Go text/template 기반 strict 렌더러를 구현한다.

**Task 5.1: Renderer 구현**
- 파일: `internal/template/renderer.go`
- 내용:
  - `NewRenderer() Renderer` 생성자
  - `Render()`: go:embed 템플릿 파싱 + 데이터 주입
  - `template.Option("missingkey=error")` strict mode
  - 커스텀 함수 맵 (필요 시)
- 요구사항: REQ-T-010, REQ-T-011, REQ-T-012
- 테스트:
  - 정상 렌더링 (데이터 주입 검증)
  - 누락 키 오류 발생 검증
  - 결과물 내 미확장 토큰 부재 검증

---

### Milestone 6: Validator (Final Goal)

배포 후 무결성 검증기를 구현한다.

**Task 6.1: JSON 유효성 검증**
- 파일: `internal/template/validator.go`
- 내용:
  - `ValidateJSON()`: `json.Valid()` 래핑
  - 상세 오류 메시지 (바이트 위치 포함)
- 요구사항: REQ-T-030
- 테스트: 유효/무효 JSON, 빈 입력, UTF-8 BOM

**Task 6.2: 경로 유효성 검증**
- 파일: `internal/template/validator.go`
- 내용:
  - `ValidatePaths()`: 경로별 정규화, containment, 특수문자 검증
  - PathError 목록 반환
- 요구사항: REQ-T-031
- 테스트: 정상 경로, `..` 포함, 절대 경로, 특수문자

**Task 6.3: 배포 통합 검증**
- 파일: `internal/template/validator.go`
- 내용:
  - `ValidateDeployment()`: 종합 검증 오케스트레이터
  - 파일 존재 여부, JSON 유효성, 경로 정규화, 에이전트 파일 무결성
  - ValidationReport 생성
- 요구사항: REQ-T-032, REQ-T-033
- 테스트: 완전한 배포 후 검증, 누락 파일, 손상 JSON

---

### Milestone 7: 통합 및 벤치마크 (Optional Goal)

**Task 7.1: 통합 테스트**
- Deploy -> Track -> DetectChanges -> Validate 전체 흐름 테스트
- 레거시 프로젝트(manifest.json 미존재) 마이그레이션 테스트
- 대규모 템플릿 세트(150+ 파일) 배포 성능 테스트

**Task 7.2: 벤치마크 테스트**
- 전체 배포 시간, 해시 계산 시간, 매니페스트 로드/저장 시간
- 메모리 할당 프로파일링 (`b.ReportAllocs()`)
- 성능 목표 충족 검증

**Task 7.3: mockery 인터페이스 Mock 생성**
- `Deployer`, `Renderer`, `SettingsGenerator`, `Validator`, `Manager` mock 생성
- 상위 모듈(`internal/update/`, `internal/core/project/`)에서 사용

---

## 3. 아키텍처 설계 방향

### 3.1 패키지 구조

```
internal/
  manifest/
    manifest.go       # Manager 인터페이스 + 구현체
    hasher.go         # SHA-256 해시 유틸리티
    types.go          # Provenance, Manifest, FileEntry, ChangedFile, 오류 타입
    manifest_test.go  # 단위 테스트
    hasher_test.go    # 단위 테스트
    testdata/         # 테스트 픽스처

  template/
    deployer.go       # Deployer 인터페이스 + 구현체 + go:embed
    renderer.go       # Renderer 인터페이스 + 구현체
    settings.go       # SettingsGenerator + Settings 구조체
    validator.go      # Validator 인터페이스 + 구현체
    deployer_test.go  # 단위 테스트
    renderer_test.go  # 단위 테스트
    settings_test.go  # 단위/계약 테스트
    validator_test.go # 단위 테스트
    testdata/         # 테스트 픽스처
```

### 3.2 핵심 설계 패턴

**Interface-Based DDD (ADR-004):**
- 모든 공개 기능은 인터페이스로 정의
- 구현체는 비공개(unexported) 구조체
- 생성자 함수(`New*`)로만 인스턴스화

**Constructor Injection:**
- `NewDeployer(manifest.Manager) Deployer` -- manifest 의존성 주입
- `NewSettingsGenerator() SettingsGenerator` -- 독립 생성
- `NewValidator() Validator` -- 독립 생성

**Atomic File Operations:**
- 모든 파일 쓰기는 임시 파일 + rename 패턴
- `pkg/utils/file.go`의 `SafeWrite()` 활용

### 3.3 go:embed 전략

```go
//go:embed templates/*
var templateFS embed.FS
```

- 빌드 시점에 `templates/` 디렉토리 전체가 바이너리에 포함
- `fs.WalkDir(templateFS, ...)` 로 순회
- 읽기 전용: 런타임 수정 불가
- 바이너리 버전과 템플릿 버전 자동 일치

### 3.4 동시성 고려사항

- Manifest Manager: 단일 goroutine 내에서 사용 (CLI 컨텍스트)
- DetectChanges: 150+ 파일 해시 계산 시 `errgroup` + goroutine pool로 병렬화 가능
- Deploy: 순차 배포 (파일 시스템 순서 보장 필요)
- context.Context: 모든 장기 실행 작업에 전파

---

## 4. 리스크 분석

### 4.1 기술적 리스크

| 리스크 | 확률 | 영향 | 대응 전략 |
|--------|------|------|----------|
| go:embed 바이너리 크기 초과 (30MB 목표) | 낮음 | 중간 | 템플릿 최소화, 불필요 파일 제외, `-ldflags "-s -w"` 적용 |
| 경로 보안 우회 (symlink 공격) | 낮음 | 높음 | `filepath.EvalSymlinks()` 추가, containment check 후 재검증 |
| 플랫폼별 경로 구분자 불일치 | 중간 | 중간 | `filepath.Clean()` + `filepath.ToSlash()` 일관 적용, 크로스 플랫폼 테스트 |
| 대규모 프로젝트 매니페스트 성능 | 낮음 | 낮음 | 150 파일 기준 벤치마크, 필요 시 병렬 해시 계산 |
| 레거시 manifest.json 미존재 | 중간 | 낮음 | 자동 생성 로직 포함, 기존 파일은 `user_created`로 분류 |
| SPEC-CONFIG-001 완료 지연 | 중간 | 높음 | Config 인터페이스 mock으로 개발 진행, 통합은 후행 |

### 4.2 설계 리스크

| 리스크 | 확률 | 영향 | 대응 전략 |
|--------|------|------|----------|
| Settings 구조체와 Claude Code 스키마 불일치 | 중간 | 높음 | Claude Code settings.json 문서 확인, 실제 파일로 역직렬화 테스트 |
| 매니페스트 동시 접근 충돌 (향후 병렬 worktree) | 낮음 | 중간 | 현재는 단일 접근 설계, 향후 파일 잠금 추가 가능 |
| 3-way merge 모듈과의 인터페이스 불일치 | 낮음 | 중간 | Manager 인터페이스를 merge 모듈과 사전 합의 |

---

## 5. 테스트 전략

### 5.1 테스트 유형

| 유형 | 대상 | 도구 | 커버리지 목표 |
|------|------|------|-------------|
| 단위 테스트 | 각 함수/메서드 | `testing` + `testify` | 90%+ |
| 테이블 기반 테스트 | 경로 검증, 해시, Provenance | `testing` | 100% 경계값 |
| 퍼즈 테스트 | JSON 파싱, 경로 정규화 | `testing.F` | 경계 발견 |
| 벤치마크 테스트 | 배포, 해시, 매니페스트 I/O | `testing.B` | 성능 목표 충족 |
| 통합 테스트 | Deploy -> Track -> Validate | `testing` | 핵심 경로 |
| 계약 테스트 | settings.json JSON 라운드트립 | `testing` | 100% |

### 5.2 테스트 픽스처

```
internal/manifest/testdata/
  valid-manifest.json           # 정상 매니페스트
  corrupt-manifest.json         # 손상된 JSON
  empty-manifest.json           # 빈 매니페스트
  large-manifest.json           # 1000+ 항목

internal/template/testdata/
  sample-templates/             # 테스트용 임베드 템플릿
  expected-settings-darwin.json # darwin 기대 출력
  expected-settings-linux.json  # linux 기대 출력
  expected-settings-windows.json # windows 기대 출력
  invalid-json/                 # 무효 JSON 샘플
```

### 5.3 Mock 인터페이스

mockery를 사용하여 다음 인터페이스의 Mock을 자동 생성:
- `manifest.Manager` -- deployer, update 모듈에서 사용
- `template.Deployer` -- project initializer, update 모듈에서 사용
- `template.Renderer` -- deployer 내부에서 사용
- `template.SettingsGenerator` -- deployer, hook 모듈에서 사용
- `template.Validator` -- deployer, project initializer에서 사용

---

## 6. 구현 순서 요약

```
[Milestone 1] Manifest Types & Hasher
    types.go -> hasher.go -> 오류 타입
         |
         v
[Milestone 2] Manifest Manager
    manifest.go (Load/Save -> Track/Get/Remove -> DetectChanges)
         |
         v
[Milestone 3] Template Deployer
    deployer.go (embed -> Extract -> Deploy + 경로 보안)
         |
         v
[Milestone 4] Settings Generator
    settings.go (구조체 -> Generate -> 플랫폼별 훅)
         |
         v
[Milestone 5] Renderer
    renderer.go (text/template + strict mode)
         |
         v
[Milestone 6] Validator
    validator.go (JSON -> 경로 -> 통합 검증)
         |
         v
[Milestone 7] 통합 및 벤치마크
    통합 테스트 -> 벤치마크 -> Mock 생성
```

---

## 7. 정의 (Definitions)

| 용어 | 정의 |
|------|------|
| Provenance | 파일의 출처와 소유권을 분류하는 메타데이터 |
| Template Hash | go:embed에 내장된 원본 템플릿의 SHA-256 해시 |
| Deployed Hash | 파일이 프로젝트에 최초 배포된 시점의 SHA-256 해시 |
| Current Hash | 파일의 현재 디스크 상태 SHA-256 해시 |
| Containment Check | 경로가 지정된 루트 디렉토리 내부에 존재하는지 검증 |
| Atomic Write | 임시 파일 작성 후 rename으로 대상 파일을 교체하는 안전한 쓰기 방식 |
| Strict Mode | Go text/template의 missingkey=error 옵션으로 누락 키 시 오류 발생 |
| Zero Runtime Expansion | 생성된 파일에 미확장 동적 토큰이 존재하지 않는 원칙 (ADR-011) |
