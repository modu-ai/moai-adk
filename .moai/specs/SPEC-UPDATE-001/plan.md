# SPEC-UPDATE-001: 구현 계획

---
tags: SPEC-UPDATE-001
---

## 1. 구현 전략 개요

### 1.1 핵심 원칙

- **Merge 엔진 우선 구현**: 바이너리 의존성 없이 독립적으로 테스트 가능한 `internal/merge/`를 먼저 구현
- **인터페이스 기반 DDD**: 모든 컴포넌트를 Go 인터페이스로 정의하여 독립 테스트 및 모킹 가능
- **원자적 연산 보장**: 바이너리 교체와 파일 머지 모두 실패 시 원래 상태 복원 가능
- **점진적 통합**: 각 컴포넌트를 독립 단위로 개발/테스트 후 Orchestrator에서 조합

### 1.2 구현 순서 근거

```
Phase A: Merge Engine (바이너리/네트워크 의존성 없음, 순수 알고리즘)
    |
Phase B: Checker + Rollback (네트워크 읽기 전용 + 파일시스템)
    |
Phase C: Updater (네트워크 다운로드 + 파일시스템 쓰기)
    |
Phase D: Orchestrator + CLI 통합 (전체 파이프라인 조합)
```

Merge 엔진을 가장 먼저 구현하는 이유:
- 외부 의존성 없이 순수 Go 알고리즘으로 완전한 단위 테스트 가능
- `internal/manifest/`(SPEC-TEMPLATE-001)와의 유일한 의존성은 인터페이스 레벨
- 6가지 전략 각각이 독립적이므로 병렬 개발 가능
- 가장 복잡한 로직(diff, conflict detection)을 먼저 안정화

---

## 2. 마일스톤

### Primary Goal: Merge Engine (`internal/merge/`)

**범위**: differ.go, three_way.go, strategies.go, conflict.go

**구현 순서:**

1. **differ.go** - Diff 생성기
   - 라인 단위 diff 알고리즘 구현 (Myers diff 또는 patience diff)
   - unified diff 형식 출력
   - base, current, updated 세 버전 간의 변경 감지
   - 테스트: 동일 파일, 추가만, 삭제만, 변경, 복합 변경 케이스

2. **three_way.go** - 3-Way Merge 코어
   - `Engine` 인터페이스 구현체
   - `ThreeWayMerge`: 바이트 슬라이스 기반 범용 머지
   - `MergeFile`: 파일 경로 기반 전략 자동 선택 머지
   - 테스트: 충돌 없는 머지, 한쪽만 변경, 양쪽 변경(비충돌), 양쪽 변경(충돌)

3. **strategies.go** - 6가지 머지 전략
   - `StrategySelector` 구현 (파일 확장자/이름 매핑)
   - `LineMerge`: 기본 라인 단위 머지
   - `YAMLDeep`: `gopkg.in/yaml.v3` 기반 deep merge
   - `JSONMerge`: `encoding/json` 기반 객체 머지
   - `SectionMerge`: Markdown 헤딩 기반 섹션 머지
   - `EntryMerge`: 라인 엔트리 합집합 머지
   - `Overwrite`: 전체 교체 + 백업
   - 테스트: 각 전략별 독립 테스트 스위트

4. **conflict.go** - 충돌 처리
   - `Conflict` 구조체 생성 및 관리
   - `.conflict` 파일 생성 (Git 스타일 충돌 마커)
   - 충돌 보고서 포맷팅
   - 테스트: 충돌 파일 생성, 마커 형식 검증, 다중 충돌 처리

**품질 기준:**
- 단위 테스트 커버리지 >= 90%
- 벤치마크 테스트: 200개 파일 머지 < 5초
- table-driven 테스트 패턴 사용
- testdata/ 디렉토리에 실제 파일 타입별 테스트 픽스처

---

### Secondary Goal: Checker + Rollback (`internal/update/`)

**범위**: checker.go, rollback.go

**구현 순서:**

1. **checker.go** - 버전 확인
   - `Checker` 인터페이스 구현체
   - GitHub Releases API 클라이언트 (`net/http`)
   - `VersionInfo` 파싱 (JSON 응답)
   - 시맨틱 버전 비교 로직
   - HTTP 클라이언트를 인터페이스로 추상화 (테스트 모킹용)
   - 테스트: httptest.Server를 이용한 API 응답 모킹, 버전 비교 로직, 네트워크 오류 처리

2. **rollback.go** - 롤백 메커니즘
   - `Rollback` 인터페이스 구현체
   - `CreateBackup`: 현재 바이너리 복사 (`{path}.backup.{timestamp}`)
   - `Restore`: 백업에서 원래 위치로 복원
   - 파일 권한(실행 비트) 보존
   - 테스트: 백업 생성/복원, 파일 권한 보존, 존재하지 않는 백업 복원 시도

**품질 기준:**
- 단위 테스트 커버리지 >= 85%
- checker: HTTP 모킹으로 네트워크 의존성 제거
- rollback: 임시 디렉토리 기반 파일시스템 테스트

---

### Tertiary Goal: Updater (`internal/update/`)

**범위**: updater.go

**구현 순서:**

1. **updater.go** - 바이너리 다운로드 및 교체
   - `Updater` 인터페이스 구현체
   - `Download`: HTTP GET으로 플랫폼 바이너리 다운로드
     - 진행률 콜백 지원 (`io.TeeReader` + 카운터)
     - 임시 파일에 기록 후 체크섬 검증
     - 실패 시 임시 파일 정리 (`defer`)
   - `Replace`: 원자적 바이너리 교체
     - 임시 위치에 새 바이너리 기록
     - 실행 권한 설정 (`os.Chmod 0755`)
     - 새 바이너리 검증 (`exec.Command("new_binary", "version")`)
     - `os.Rename`으로 원자적 교체
   - 플랫폼별 처리:
     - Unix: `os.Rename` 직접 사용
     - Windows: 프로세스 잠금 우회 (실행 중인 바이너리 rename 후 새 바이너리 배치)
   - 테스트: httptest.Server로 다운로드 모킹, 체크섬 불일치 검증, 파일 교체 원자성

**품질 기준:**
- 단위 테스트 커버리지 >= 85%
- 플랫폼별 빌드 태그 테스트 (`//go:build !windows`, `//go:build windows`)
- 다운로드 중단/재시도 시나리오

---

### Final Goal: Orchestrator + CLI 통합

**범위**: orchestrator.go, internal/cli/update.go 연동

**구현 순서:**

1. **orchestrator.go** - 전체 파이프라인 조율
   - `Orchestrator` 인터페이스 구현체
   - 의존성 주입: Checker, Updater, Rollback, Manifest, Template, Merge Engine
   - `Update` 메서드: 전체 9단계 파이프라인 실행
   - 매니페스트 기반 파일별 처리 결정 (REQ-INT-001)
   - `UpdateResult` 집계 및 반환
   - 오류 발생 시 자동 롤백 트리거
   - 테스트: 모든 의존성을 모킹한 통합 테스트, 실패 시나리오별 롤백 검증

2. **CLI 통합** (`internal/cli/update.go`)
   - Cobra 커맨드 등록 (`moai update`)
   - Orchestrator 의존성 와이어링
   - 진행률 UI (`internal/ui/progress.go`)
   - 결과 요약 출력
   - `--check` 플래그: 업데이트 확인만 (다운로드/교체 없음)
   - `--force` 플래그: 확인 없이 업데이트
   - 테스트: CLI 플래그 파싱, 출력 형식 검증

**품질 기준:**
- 통합 테스트 커버리지 >= 80%
- 엔드투엔드 시나리오: 정상 업데이트, 네트워크 실패, 체크섬 불일치, 롤백 실행
- Orchestrator의 각 단계별 실패 주입 테스트

---

## 3. 기술적 접근

### 3.1 Diff 알고리즘 선택

**권장**: Myers diff 알고리즘

- 최소 편집 거리를 보장하여 가장 직관적인 diff 생성
- O(ND) 시간 복잡도 (D = 편집 거리)
- Git이 사용하는 알고리즘과 동일하여 사용자에게 익숙한 결과
- Go 표준 라이브러리에는 미포함이므로 자체 구현 또는 `github.com/sergi/go-diff` 사용 검토

### 3.2 원자적 파일 교체

```go
// Unix: os.Rename은 원자적 (같은 파일시스템 내)
func atomicReplace(src, dst string) error {
    // 1. 임시 파일을 dst와 같은 디렉토리에 생성 (같은 파일시스템 보장)
    // 2. 내용 기록
    // 3. os.Rename(tmp, dst) -- 원자적
    return os.Rename(src, dst)
}
```

### 3.3 GitHub Releases API 활용

```
GET /repos/modu-ai/moai-adk-go/releases/latest
Response:
{
  "tag_name": "v1.2.0",
  "assets": [
    {
      "name": "moai-darwin-arm64",
      "browser_download_url": "https://..."
    },
    {
      "name": "checksums.txt",
      "browser_download_url": "https://..."
    }
  ]
}
```

- 플랫폼 바이너리명 패턴: `moai-{GOOS}-{GOARCH}` (Windows: `.exe` 접미사)
- 체크섬 파일: goreleaser가 생성하는 `checksums.txt` (SHA-256)

### 3.4 YAML Deep Merge 전략

```go
// YAML deep merge 핵심 로직
func yamlDeepMerge(base, current, updated map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    // 1. base의 모든 키를 result에 복사
    // 2. current의 변경(base 대비)을 result에 적용
    // 3. updated의 변경(base 대비)을 result에 적용
    // 4. current와 updated 양측 변경이 동일 키에서 발생하면 충돌 기록
    return result
}
```

### 3.5 SectionMerge 파싱

CLAUDE.md 섹션 분리 방식:

```
## Section A        <-- 섹션 경계
content a-1
content a-2

## Section B        <-- 섹션 경계
content b-1
```

- `## ` 또는 `### ` 패턴으로 섹션 경계 식별
- 각 섹션을 독립 머지 단위로 취급
- 섹션 순서는 updated(새 템플릿)의 순서를 따르되, 사용자 추가 섹션은 끝에 배치

### 3.6 모킹 전략

```go
// mockery로 모든 인터페이스에 대한 mock 자동 생성
//go:generate mockery --name=Checker --dir=. --output=./mocks --outpkg=mocks
//go:generate mockery --name=Updater --dir=. --output=./mocks --outpkg=mocks
//go:generate mockery --name=Rollback --dir=. --output=./mocks --outpkg=mocks
//go:generate mockery --name=Engine --dir=. --output=./mocks --outpkg=mocks
```

---

## 4. 아키텍처 설계

### 4.1 패키지 구조

```
internal/
├── merge/
│   ├── engine.go           # Engine 인터페이스 + 구현체
│   ├── three_way.go        # 3-Way Merge 코어 알고리즘
│   ├── strategies.go       # StrategySelector + 6가지 전략
│   ├── conflict.go         # Conflict 구조체 + .conflict 파일 생성
│   ├── differ.go           # Diff 생성기 (Myers algorithm)
│   ├── types.go            # MergeStrategy, MergeResult, Conflict 타입
│   ├── mocks/              # mockery 생성 mocks
│   └── testdata/           # 테스트 픽스처
│       ├── yaml/
│       ├── json/
│       ├── markdown/
│       └── gitignore/
├── update/
│   ├── checker.go          # GitHub Releases API 클라이언트
│   ├── updater.go          # 다운로드 + 원자적 교체
│   ├── rollback.go         # 백업 + 복원
│   ├── orchestrator.go     # 전체 파이프라인 조율
│   ├── types.go            # VersionInfo, UpdateResult 타입
│   ├── mocks/              # mockery 생성 mocks
│   └── testdata/           # 테스트 바이너리 픽스처
```

### 4.2 의존성 주입 패턴

```go
// Orchestrator 생성자 -- 모든 의존성을 인터페이스로 주입
func NewOrchestrator(
    checker   Checker,
    updater   Updater,
    rollback  Rollback,
    manifest  manifest.Manager,
    deployer  template.Deployer,
    merger    merge.Engine,
    ui        ui.Progress,
    cfg       *config.Config,
) *orchestratorImpl { ... }
```

### 4.3 오류 처리 전략

- 모든 오류는 `fmt.Errorf("context: %w", err)`로 래핑
- 센티널 오류 (`ErrChecksumMismatch` 등)로 타입 체크 가능
- Orchestrator 수준에서 오류 발생 시 자동 롤백 트리거
- 롤백 실패 시 사용자에게 수동 복구 안내

---

## 5. 리스크 및 대응

### 5.1 기술적 리스크

| 리스크 | 영향 | 확률 | 대응 방안 |
|--------|------|------|-----------|
| Windows 바이너리 자기 교체 제한 | 높음 | 중간 | 실행 중 바이너리 rename 후 새 바이너리 배치 패턴 적용 |
| YAML deep merge 시 주석 유실 | 중간 | 높음 | `gopkg.in/yaml.v3`의 Node API 사용으로 주석 보존 |
| 대용량 파일 머지 시 메모리 초과 | 중간 | 낮음 | 파일 크기 임계값(10MB) 초과 시 Overwrite 전략으로 폴백 |
| GitHub API rate limit | 중간 | 중간 | 조건부 요청(If-None-Match), 캐시된 응답 활용 |
| 크로스 파일시스템 rename 실패 | 높음 | 낮음 | 임시 파일을 대상과 같은 디렉토리에 생성 |

### 5.2 의존성 리스크

| 의존성 | 리스크 | 대응 방안 |
|--------|--------|-----------|
| SPEC-TEMPLATE-001 (manifest) | 선행 구현 미완료 | Merge 엔진은 manifest 없이 독립 테스트 가능, 인터페이스 기반 모킹 |
| SPEC-CONFIG-001 | 설정 스키마 미확정 | 업데이트 관련 설정은 기본값으로 동작하도록 설계 |
| GitHub Releases | API 변경 가능 | API 응답 파싱을 별도 함수로 격리, 버전 고정 |

---

## 6. 테스트 전략

### 6.1 단위 테스트

- **table-driven 테스트**: 모든 머지 전략에 대해 `[]struct{name, base, current, updated, expected}` 패턴
- **testdata/**: 실제 파일 형식별 테스트 픽스처 (YAML, JSON, Markdown, .gitignore)
- **t.Parallel()**: 독립 테스트는 병렬 실행
- **fuzz 테스트**: differ.go의 diff 알고리즘에 대한 fuzz 테스트

### 6.2 통합 테스트

- **httptest.Server**: GitHub API 응답 모킹
- **임시 디렉토리**: 파일시스템 작업용 `t.TempDir()`
- **Orchestrator 통합**: 모든 컴포넌트 결합 테스트 (실제 파일 I/O, 모킹 네트워크)

### 6.3 벤치마크 테스트

- `BenchmarkThreeWayMerge`: 다양한 파일 크기별
- `BenchmarkYAMLDeepMerge`: 복잡한 YAML 구조
- `BenchmarkFullUpdate`: 200개 파일 프로젝트 머지

---

## 7. 추적성

| 마일스톤 | 요구사항 | 파일 |
|----------|----------|------|
| Primary Goal | REQ-MRG-001 ~ REQ-MRG-010 | three_way.go, strategies.go, conflict.go, differ.go |
| Secondary Goal | REQ-UPD-001, REQ-UPD-002, REQ-UPD-006, REQ-UPD-007 | checker.go, rollback.go |
| Tertiary Goal | REQ-UPD-003, REQ-UPD-004, REQ-UPD-005, REQ-UPD-009, REQ-UPD-010 | updater.go |
| Final Goal | REQ-UPD-008, REQ-INT-001, REQ-INT-002 | orchestrator.go |
