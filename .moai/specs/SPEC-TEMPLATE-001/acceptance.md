# SPEC-TEMPLATE-001: Acceptance Criteria

---
spec_id: SPEC-TEMPLATE-001
type: acceptance
created: 2026-02-03
status: Planned
format: Given-When-Then (Gherkin)
---

## 1. Template Deployment (deployer.go)

### AC-T-001: 정상 템플릿 배포

```gherkin
Feature: 템플릿 배포
  go:embed 파일시스템에서 프로젝트 루트로 템플릿을 배포한다.

  Scenario: 전체 템플릿 정상 배포
    Given 유효한 프로젝트 루트 디렉토리가 존재하고
    And 매니페스트 매니저가 초기화되어 있을 때
    When Deploy(ctx, projectRoot, manifest)를 호출하면
    Then 모든 임베드 템플릿이 projectRoot에 배포되고
    And 배포된 각 파일이 manifest.Track()으로 등록되고
    And 각 파일의 provenance가 "template_managed"로 설정되고
    And 반환 오류가 nil이다

  Scenario: 이미 존재하는 파일에 대한 재배포
    Given projectRoot에 이전 배포 파일이 존재하고
    And 매니페스트에 해당 파일이 "template_managed"로 등록되어 있을 때
    When Deploy를 다시 호출하면
    Then 파일이 최신 템플릿으로 덮어쓰기되고
    And 매니페스트의 해시가 갱신된다
```

### AC-T-002: 개별 템플릿 추출

```gherkin
  Scenario: 존재하는 템플릿 추출
    Given go:embed 파일시스템에 ".claude/agents/moai/expert-backend.md"가 존재할 때
    When ExtractTemplate(".claude/agents/moai/expert-backend.md")를 호출하면
    Then 해당 파일의 바이트 내용이 반환되고
    And 반환 오류가 nil이다

  Scenario: 존재하지 않는 템플릿 추출
    Given go:embed 파일시스템에 "nonexistent.txt"가 존재하지 않을 때
    When ExtractTemplate("nonexistent.txt")를 호출하면
    Then nil 바이트 배열이 반환되고
    And ErrTemplateNotFound 오류가 반환된다
```

### AC-T-003: 템플릿 목록 조회

```gherkin
  Scenario: 전체 템플릿 목록 반환
    Given go:embed 파일시스템에 템플릿이 번들되어 있을 때
    When ListTemplates()를 호출하면
    Then 임베드된 모든 파일의 상대 경로 목록이 반환되고
    And 목록이 비어있지 않다
```

### AC-T-004: 경로 보안

```gherkin
  Scenario: 경로 순회 공격 차단
    Given 프로젝트 루트가 "/home/user/project"이고
    And 템플릿 경로에 "../etc/passwd"가 포함되어 있을 때
    When Deploy를 호출하면
    Then ErrPathTraversal 오류가 반환되고
    And "/etc/passwd"에 파일이 기록되지 않는다

  Scenario: 절대 경로 차단
    Given 템플릿 경로에 "/tmp/malicious"가 포함되어 있을 때
    When 경로 정규화를 수행하면
    Then ErrPathTraversal 오류가 반환된다

  Scenario: 경로 정규화 적용
    Given 템플릿 경로가 ".claude/./agents/../agents/moai/file.md"일 때
    When 경로를 정규화하면
    Then 결과 경로가 ".claude/agents/moai/file.md"이다
```

### AC-T-005: 디렉토리 자동 생성

```gherkin
  Scenario: 중간 디렉토리 자동 생성
    Given projectRoot 내에 ".claude/agents/moai/" 디렉토리가 존재하지 않을 때
    When ".claude/agents/moai/expert-backend.md" 배포를 실행하면
    Then ".claude/agents/moai/" 디렉토리가 자동으로 생성되고
    And 파일이 정상적으로 배포된다
```

### AC-T-006: 컨텍스트 취소 처리

```gherkin
  Scenario: 배포 중 컨텍스트 취소
    Given 150개 템플릿 중 50개가 배포된 상태에서
    When context가 취소되면
    Then 배포 작업이 중단되고
    And 이미 배포된 50개 파일은 디스크에 유지되고
    And 매니페스트는 50개 파일만 추적하고
    And context.Canceled 오류가 반환된다
```

---

## 2. Settings Generator (settings.go)

### AC-T-020: settings.json 생성 (ADR-011)

```gherkin
Feature: settings.json 프로그래밍 방식 생성
  Go 구조체 직렬화를 통해 유효한 settings.json을 생성한다.

  Scenario: darwin 플랫폼 settings.json 생성
    Given 유효한 Config 구조체가 존재하고
    And platform이 "darwin"일 때
    When Generate(cfg, "darwin")를 호출하면
    Then 유효한 JSON 바이트 배열이 반환되고
    And json.Valid(result)가 true를 반환하고
    And JSON을 Settings 구조체로 역직렬화할 수 있고
    And 훅 명령이 "moai hook session-start" 형식이다

  Scenario: linux 플랫폼 settings.json 생성
    Given platform이 "linux"일 때
    When Generate(cfg, "linux")를 호출하면
    Then 훅 명령이 "moai hook session-start" 형식이고
    And JSON이 유효하다

  Scenario: windows 플랫폼 settings.json 생성
    Given platform이 "windows"일 때
    When Generate(cfg, "windows")를 호출하면
    Then 훅 명령이 Windows 호환 형식이고
    And JSON이 유효하다
```

### AC-T-021: JSON 라운드트립 검증 (계약 테스트)

```gherkin
  Scenario: JSON 라운드트립 무손실
    Given Generate()로 settings.json을 생성했을 때
    When 결과를 json.Unmarshal로 파싱하고
    And 다시 json.MarshalIndent로 직렬화하면
    Then 원본과 재직렬화 결과가 동일하다
```

### AC-T-022: 미확장 토큰 부재 검증

```gherkin
  Scenario: 동적 토큰 부재 확인
    Given Generate()로 settings.json을 생성했을 때
    When 결과 바이트를 문자열로 검사하면
    Then "${" 문자열이 포함되어 있지 않고
    And "{{" 문자열이 포함되어 있지 않고
    And "$VAR" 패턴이 포함되어 있지 않다
```

### AC-T-023: 필수 훅 이벤트 포함

```gherkin
  Scenario: 모든 필수 훅 이벤트 포함
    Given Generate()로 settings.json을 생성했을 때
    When 결과를 Settings 구조체로 파싱하면
    Then Hooks 맵에 "SessionStart" 키가 존재하고
    And Hooks 맵에 "PreToolUse" 키가 존재하고
    And Hooks 맵에 "PostToolUse" 키가 존재하고
    And Hooks 맵에 "SessionEnd" 키가 존재하고
    And Hooks 맵에 "Stop" 키가 존재하고
    And Hooks 맵에 "PreCompact" 키가 존재한다
```

### AC-T-024: 문자열 연결 미사용 검증

```gherkin
  Scenario: 코드 내 문자열 연결 부재 (정적 분석)
    Given internal/template/settings.go 소스 코드를 분석할 때
    When fmt.Sprintf 또는 strings.Builder를 JSON 구성에 사용한 부분을 검색하면
    Then 해당 패턴이 발견되지 않는다
```

---

## 3. Renderer (renderer.go)

### AC-T-030: 템플릿 렌더링

```gherkin
Feature: Go text/template 기반 렌더링
  Strict mode로 템플릿을 렌더링한다.

  Scenario: 정상 렌더링
    Given "CLAUDE.md.tmpl" 템플릿이 존재하고
    And data에 모든 필수 키가 포함되어 있을 때
    When Render("CLAUDE.md.tmpl", data)를 호출하면
    Then 렌더링된 바이트 배열이 반환되고
    And 모든 템플릿 변수가 실제 값으로 대체되고
    And 반환 오류가 nil이다

  Scenario: 누락 키 오류 (Strict Mode)
    Given 템플릿에 "{{.ProjectName}}" 변수가 있고
    And data에 "ProjectName" 키가 없을 때
    When Render를 호출하면
    Then ErrMissingTemplateKey 오류가 반환되고
    And 오류 메시지에 누락된 키 이름이 포함된다

  Scenario: 렌더링 결과 미확장 토큰 검증
    Given 정상적으로 렌더링을 완료했을 때
    When 결과를 문자열로 검사하면
    Then "{{." 패턴이 포함되어 있지 않다
```

---

## 4. Validator (validator.go)

### AC-T-040: JSON 유효성 검증

```gherkin
Feature: 배포 후 검증
  배포된 파일의 무결성을 검증한다.

  Scenario: 유효한 JSON 검증 통과
    Given 올바른 JSON 바이트가 주어졌을 때
    When ValidateJSON(data)를 호출하면
    Then 반환 오류가 nil이다

  Scenario: 무효한 JSON 검증 실패
    Given 손상된 JSON 바이트 (예: 닫히지 않은 중괄호)가 주어졌을 때
    When ValidateJSON(data)를 호출하면
    Then ErrInvalidJSON 오류가 반환된다

  Scenario: 빈 입력 검증
    Given 빈 바이트 배열이 주어졌을 때
    When ValidateJSON(data)를 호출하면
    Then ErrInvalidJSON 오류가 반환된다
```

### AC-T-041: 경로 유효성 검증

```gherkin
  Scenario: 정상 경로 검증 통과
    Given projectRoot가 "/home/user/project"이고
    And files에 [".claude/settings.json", "CLAUDE.md"]가 있을 때
    When ValidatePaths(projectRoot, files)를 호출하면
    Then 빈 PathError 슬라이스가 반환된다

  Scenario: 경로 순회 검출
    Given files에 ["../../../etc/passwd"]가 있을 때
    When ValidatePaths를 호출하면
    Then PathError 슬라이스에 해당 경로의 오류가 포함된다
```

### AC-T-042: 배포 통합 검증

```gherkin
  Scenario: 정상 배포 검증 통과
    Given 모든 템플릿이 정상 배포된 프로젝트 루트가 있을 때
    When ValidateDeployment(projectRoot)를 호출하면
    Then report.Valid가 true이고
    And report.Errors가 비어있고
    And report.FilesChecked가 배포된 파일 수와 일치한다

  Scenario: 누락 파일 감지
    Given 배포 후 ".claude/settings.json"을 삭제했을 때
    When ValidateDeployment를 호출하면
    Then report.Valid가 false이고
    And report.Errors에 해당 파일 경로가 포함된다

  Scenario: 손상된 JSON 파일 감지
    Given ".claude/settings.json" 내용을 손상시켰을 때
    When ValidateDeployment를 호출하면
    Then report.Valid가 false이고
    And report.Errors에 JSON 유효성 오류가 포함된다
```

---

## 5. Manifest Manager (manifest.go)

### AC-M-001: 매니페스트 로드

```gherkin
Feature: 매니페스트 관리
  파일 출처와 해시를 추적하는 매니페스트를 관리한다.

  Scenario: 정상 매니페스트 로드
    Given ".moai/manifest.json"에 유효한 매니페스트가 존재할 때
    When Load(projectRoot)를 호출하면
    Then Manifest 구조체가 반환되고
    And 모든 FileEntry가 올바르게 파싱되고
    And 반환 오류가 nil이다

  Scenario: 매니페스트 파일 미존재 시 자동 생성
    Given ".moai/manifest.json"이 존재하지 않을 때
    When Load(projectRoot)를 호출하면
    Then 빈 Files 맵을 가진 Manifest가 반환되고
    And 반환 오류가 nil이다

  Scenario: 손상된 매니페스트 복구
    Given ".moai/manifest.json"에 손상된 JSON이 있을 때
    When Load(projectRoot)를 호출하면
    Then ErrManifestCorrupt 오류가 반환되고
    And 손상된 파일이 ".moai/manifest.json.corrupt"로 백업되고
    And 새 빈 매니페스트가 생성된다
```

### AC-M-002: 매니페스트 저장

```gherkin
  Scenario: 정상 저장
    Given 메모리에 매니페스트가 로드되어 있을 때
    When Save()를 호출하면
    Then ".moai/manifest.json"에 JSON이 기록되고
    And json.Valid()로 검증 시 유효하고
    And 반환 오류가 nil이다

  Scenario: 원자적 쓰기 보장
    Given Save() 실행 중 프로세스가 중단될 때
    When 파일 시스템을 확인하면
    Then 기존 manifest.json이 손상되지 않았거나
    Or 완전히 새로운 내용으로 교체되어 있다
    And 부분적으로 기록된 상태는 존재하지 않는다
```

### AC-M-003: 파일 추적

```gherkin
  Scenario: 새 파일 추적 등록
    Given 매니페스트에 ".claude/settings.json"이 등록되어 있지 않고
    And 해당 파일이 디스크에 존재할 때
    When Track(".claude/settings.json", TemplateManaged, "sha256:abc...")를 호출하면
    Then 매니페스트에 새 FileEntry가 추가되고
    And provenance가 "template_managed"이고
    And template_hash가 "sha256:abc..."이고
    And deployed_hash와 current_hash가 현재 파일 해시와 일치한다

  Scenario: 기존 파일 추적 갱신
    Given 매니페스트에 ".claude/settings.json"이 이미 등록되어 있을 때
    When Track()을 동일 경로로 다시 호출하면
    Then 기존 항목이 갱신되고
    And 새 항목이 추가되지 않는다
```

### AC-M-004: 변경 감지

```gherkin
  Scenario: 변경된 파일 감지
    Given 매니페스트에 3개 파일이 등록되어 있고
    And 1개 파일의 내용이 변경되었을 때
    When DetectChanges()를 호출하면
    Then ChangedFile 슬라이스에 1개 항목이 반환되고
    And OldHash와 NewHash가 다르고
    And Path가 변경된 파일을 가리킨다

  Scenario: 변경 없음
    Given 매니페스트에 등록된 모든 파일이 변경되지 않았을 때
    When DetectChanges()를 호출하면
    Then 빈 ChangedFile 슬라이스가 반환된다

  Scenario: 삭제된 파일 감지
    Given 매니페스트에 등록된 파일이 디스크에서 삭제되었을 때
    When DetectChanges()를 호출하면
    Then ChangedFile에 해당 파일이 포함되고
    And NewHash가 빈 문자열이다
```

### AC-M-005: 항목 조회 및 제거

```gherkin
  Scenario: 존재하는 항목 조회
    Given 매니페스트에 ".claude/settings.json"이 등록되어 있을 때
    When GetEntry(".claude/settings.json")를 호출하면
    Then FileEntry 포인터가 반환되고
    And exists가 true이다

  Scenario: 미존재 항목 조회
    Given 매니페스트에 "nonexistent.txt"가 등록되어 있지 않을 때
    When GetEntry("nonexistent.txt")를 호출하면
    Then nil이 반환되고
    And exists가 false이다

  Scenario: 항목 제거
    Given 매니페스트에 ".claude/settings.json"이 등록되어 있을 때
    When Remove(".claude/settings.json")를 호출하면
    Then 매니페스트에서 해당 항목이 삭제되고
    And GetEntry(".claude/settings.json") 결과의 exists가 false이다
```

---

## 6. Hasher (hasher.go)

### AC-M-010: SHA-256 해시 계산

```gherkin
Feature: SHA-256 파일 해시
  파일 내용의 SHA-256 해시를 계산한다.

  Scenario: 정상 파일 해시 계산
    Given "hello world" 내용의 파일이 존재할 때
    When HashFile(path)를 호출하면
    Then "sha256:" 접두사로 시작하는 hex 인코딩 문자열이 반환되고
    And 해시 값이 알려진 SHA-256 값과 일치한다

  Scenario: 빈 파일 해시 계산
    Given 빈 파일이 존재할 때
    When HashFile(path)를 호출하면
    Then SHA-256 빈 문자열 해시가 반환되고
    And 반환 오류가 nil이다

  Scenario: 대용량 파일 스트리밍 해시
    Given 10MB 이상의 파일이 존재할 때
    When HashFile(path)를 호출하면
    Then 해시가 정상적으로 반환되고
    And 메모리 사용량이 파일 크기보다 현저히 적다

  Scenario: 미존재 파일 해시 시도
    Given 존재하지 않는 파일 경로가 주어졌을 때
    When HashFile(path)를 호출하면
    Then 빈 문자열이 반환되고
    And os.ErrNotExist를 래핑한 오류가 반환된다

  Scenario: 바이트 배열 해시 계산
    Given "hello world" 바이트 배열이 주어졌을 때
    When HashBytes(data)를 호출하면
    Then "sha256:" 접두사로 시작하는 hex 인코딩 문자열이 반환되고
    And HashFile()로 동일 내용의 파일을 해시한 결과와 일치한다
```

---

## 7. Provenance Types (types.go)

### AC-M-020: Provenance 열거형

```gherkin
Feature: 파일 출처 분류
  배포된 파일의 출처를 4가지 타입으로 분류한다.

  Scenario: Provenance 상수 값 검증
    Given Provenance 타입이 정의되어 있을 때
    When 각 상수를 문자열로 변환하면
    Then TemplateManaged는 "template_managed"이고
    And UserModified는 "user_modified"이고
    And UserCreated는 "user_created"이고
    And Deprecated는 "deprecated"이다

  Scenario: Provenance JSON 직렬화
    Given TemplateManaged Provenance가 있을 때
    When JSON으로 직렬화하면
    Then "template_managed" 문자열로 인코딩된다

  Scenario: Provenance JSON 역직렬화
    Given JSON에 "user_modified" 문자열이 있을 때
    When Provenance로 역직렬화하면
    Then UserModified 값이 된다
```

### AC-M-021: Manifest/FileEntry 구조체

```gherkin
  Scenario: Manifest JSON 라운드트립
    Given 3개 FileEntry를 포함하는 Manifest 구조체가 있을 때
    When json.MarshalIndent로 직렬화하고
    And json.Unmarshal로 역직렬화하면
    Then 원본과 역직렬화 결과가 동일하다

  Scenario: FileEntry 필드 완전성
    Given FileEntry가 생성될 때
    When 모든 필드를 검사하면
    Then Provenance, TemplateHash, DeployedHash, CurrentHash가 모두 존재한다
```

---

## 8. Edge Cases & Error Scenarios

### AC-E-001: 동시성 안전

```gherkin
  Scenario: 매니페스트 단일 접근 보장
    Given CLI 컨텍스트에서 단일 goroutine이 실행 중일 때
    When manifest 작업을 수행하면
    Then 데이터 경쟁이 발생하지 않는다
    And go test -race 플래그로 검증된다
```

### AC-E-002: 디스크 공간 부족

```gherkin
  Scenario: 배포 중 디스크 공간 부족
    Given 디스크 공간이 부족한 환경에서
    When Deploy()를 호출하면
    Then 적절한 오류가 반환되고
    And 매니페스트는 실제로 배포된 파일만 반영한다
```

### AC-E-003: 읽기 전용 디렉토리

```gherkin
  Scenario: 읽기 전용 projectRoot에 배포 시도
    Given projectRoot에 쓰기 권한이 없을 때
    When Deploy()를 호출하면
    Then permission denied 오류가 반환된다
```

### AC-E-004: 특수 문자 경로

```gherkin
  Scenario: 공백 포함 경로 처리
    Given 프로젝트 루트에 "My Project" 같은 공백이 포함될 때
    When Deploy()를 호출하면
    Then 정상적으로 배포가 완료된다

  Scenario: 유니코드 경로 처리
    Given 프로젝트 루트에 한글/일본어 문자가 포함될 때
    When Deploy()를 호출하면
    Then 정상적으로 배포가 완료된다
```

### AC-E-005: 빈 Config 처리

```gherkin
  Scenario: 최소 Config로 settings.json 생성
    Given Config 구조체에 기본값만 설정되어 있을 때
    When Generate(cfg, "darwin")를 호출하면
    Then 유효한 JSON이 생성되고
    And 필수 훅 이벤트가 모두 포함된다
```

### AC-E-006: 해시 불일치 시나리오

```gherkin
  Scenario: deployed_hash와 current_hash 불일치 감지
    Given 매니페스트에 deployed_hash가 "sha256:aaa"로 기록되어 있고
    And 사용자가 해당 파일을 수정하여 current_hash가 "sha256:bbb"가 되었을 때
    When DetectChanges()를 호출하면
    Then ChangedFile에 해당 파일이 포함되고
    And OldHash가 "sha256:aaa"이고 NewHash가 "sha256:bbb"이다
```

### AC-E-007: 레거시 프로젝트 마이그레이션

```gherkin
  Scenario: manifest.json 없는 레거시 프로젝트 배포
    Given ".moai/manifest.json"이 존재하지 않는 레거시 프로젝트에서
    When Load()를 호출하면
    Then 빈 매니페스트가 생성되고
    And 이후 Deploy()를 호출하면 정상적으로 배포되고
    And 모든 파일이 매니페스트에 등록된다
```

---

## 9. 성능 수용 기준

### AC-P-001: 배포 성능

```gherkin
  Scenario: 150개 템플릿 배포 시간
    Given ~150개, ~1.1MB 총량의 임베드 템플릿이 있을 때
    When Deploy()를 실행하면
    Then 완료 시간이 500ms 이내이다

  Scenario: 매니페스트 로드 시간
    Given 150개 항목의 manifest.json이 있을 때
    When Load()를 실행하면
    Then 완료 시간이 10ms 이내이다

  Scenario: settings.json 생성 시간
    Given 유효한 Config가 있을 때
    When Generate()를 실행하면
    Then 완료 시간이 5ms 이내이다

  Scenario: 해시 계산 시간 (단일 파일)
    Given 일반적인 크기(~10KB)의 파일이 있을 때
    When HashFile()를 실행하면
    Then 완료 시간이 5ms 이내이다
```

---

## 10. Quality Gate (Definition of Done)

### 코드 품질 기준

| 기준 | 목표 | 검증 방법 |
|------|------|----------|
| 단위 테스트 커버리지 | >= 90% | `go test -coverprofile` |
| 경쟁 조건 | 0 | `go test -race` |
| 린트 오류 | 0 | `golangci-lint run` |
| 보안 취약점 | 0 | `gosec` (golangci-lint 내장) |
| JSON 라운드트립 | 100% 통과 | 계약 테스트 |
| 벤치마크 | 모든 성능 목표 충족 | `go test -bench` |
| godoc 주석 | 모든 exported 심볼 | `go vet` |
| 퍼즈 테스트 | 최소 10,000 반복 | `go test -fuzz` |

### TRUST 5 검증

| 원칙 | 적용 |
|------|------|
| **Tested** | 단위/통합/퍼즈/벤치마크 테스트 90%+ 커버리지 |
| **Readable** | godoc 주석, 명확한 명명, 테이블 기반 테스트 |
| **Unified** | gofumpt 포맷, golangci-lint 통과, 일관된 오류 처리 |
| **Secured** | 경로 순회 방지, JSON 주입 방지, 원자적 쓰기, 해시 무결성 |
| **Trackable** | SPEC 참조 커밋, 이슈 번호 연결, 변경 이력 |

### 수용 완료 조건

- [ ] 모든 AC 시나리오가 자동화된 테스트로 구현됨
- [ ] `go test -race -coverprofile=coverage.out ./internal/template/... ./internal/manifest/...` 90%+ 커버리지
- [ ] `golangci-lint run ./internal/template/... ./internal/manifest/...` 오류 0건
- [ ] JSON 라운드트립 계약 테스트 통과 (settings.json, manifest.json)
- [ ] 벤치마크 테스트가 성능 목표 충족
- [ ] mockery Mock이 생성되어 상위 모듈 테스트 가능
- [ ] 크로스 플랫폼 경로 테스트 통과 (darwin, linux, windows)
- [ ] ADR-007, ADR-011 준수 검증 완료
