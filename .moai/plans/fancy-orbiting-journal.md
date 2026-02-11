# Config 구조 개선 및 Update 병합 로직 개선

## Context

`.moai/config/` 분석 결과, config.yaml과 sections/ 파일 간 중복 필드, 미사용 설정 파일, 그리고 `moai update` 시 병합 로직 결함이 발견됨.

**문제 요약:**
1. `config.yaml`과 `sections/project.yaml`에 동일 필드 중복 (project.name, description, created_at, initialized)
2. `config.yaml`의 `moai.version`이 빈 값이고 `system.yaml`이 실제 값 보유 → 모순
3. `pricing.yaml`, `llm.yaml`은 Go 코드에서도, AI 에이전트에서도 사용하지 않는 완전 미사용 파일
4. `moai update` 병합 시 old 값 항상 우선 → 템플릿 기본값 변경이 사용자에게 반영 안됨

**사용자 결정:**
- config.yaml 완전 제거
- 사용자 변경 감지 3-way 병합 도입

---

## Phase 1: 미사용 설정 파일 제거

### 1.1 템플릿에서 제거할 파일

| 파일 | 제거 사유 |
|------|----------|
| `pricing.yaml` | Go 코드 미사용, AI 에이전트 미참조, GLM 미구현 기능 |
| `llm.yaml` | Go 코드 미사용, AI 에이전트 미참조, GLM 미구현 기능 |

**수정 파일:**
- `internal/template/templates/.moai/config/sections/pricing.yaml` → 삭제
- `internal/template/templates/.moai/config/sections/llm.yaml` → 삭제
- `internal/defs/files.go` → `PricingYAML`, `LLMYAML` 상수 제거 (존재 시)

### 1.2 로컬 프로젝트에서 제거

- `.moai/config/sections/pricing.yaml` → 삭제
- `.moai/config/sections/llm.yaml` → 삭제

---

## Phase 2: config.yaml 제거 및 필드 이전

### 2.1 필드 이전 매핑

| config.yaml 필드 | 이전 대상 | 비고 |
|------------------|----------|------|
| `moai.initialized` | `sections/project.yaml` → `project.initialized` | 이미 project.yaml에 존재 |
| `moai.initialized_at` | 제거 | 빈 값, 미사용 |
| `moai.version` | 제거 | system.yaml이 canonical source |
| `project.name` | `sections/project.yaml` → `project.name` | 이미 존재 |
| `project.description` | `sections/project.yaml` → `project.description` | 이미 존재 |
| `project.mode` | `sections/project.yaml` → `project.mode` | 신규 이전 |
| `project.created_at` | `sections/project.yaml` → `project.created_at` | 이미 존재 |
| `project.optimized` | `sections/project.yaml` → `project.optimized` | 신규 이전 |
| `project.template_version` | `sections/system.yaml` → `moai.template_version` | 핵심 이전 |

### 2.2 project.yaml 템플릿 업데이트

```yaml
# Before (현재)
project:
  name: ""
  description: ""
  type: ""
  created_at: ""
  initialized: "true"
github:
  profile_name: ""

# After (개선)
project:
  name: ""
  description: ""
  mode: "personal"
  created_at: ""
  initialized: true
  optimized: false
  template_version: ""
```

### 2.3 system.yaml 템플릿 업데이트

`template_version` 필드를 system.yaml로 이전:

```yaml
moai:
  version: "2.2.6"
  template_version: ""    # 추가: moai update 시 업데이트됨
  update_check_frequency: daily
  # ... 나머지 유지
```

### 2.4 수정할 Go 파일

1. **`internal/statusline/version.go`** (Lines 24-44, 111-138)
   - `readVersionFromConfig()`: config.yaml 대신 system.yaml에서 `moai.template_version` 읽기
   - `VersionConfig` struct 업데이트
   - `effectiveVersion()`: system.yaml의 template_version 사용

2. **`internal/cli/update.go`** (Lines 825-868)
   - `getProjectConfigVersion()`: system.yaml에서 template_version 읽기
   - `highRiskFiles` 목록에서 config.yaml 제거 (Line 595-615)
   - `updateTemplateVersion()` (있다면): system.yaml에 기록

3. **`internal/core/project/initializer.go`** (Lines 304-404)
   - `generateConfigsFallback()`: config.yaml 생성 제거
   - project.yaml에 mode, optimized 필드 추가

4. **`internal/defs/files.go`** (Line 15)
   - `ConfigYAML` 상수 제거

5. **`internal/template/templates/.moai/config/config.yaml`** → 삭제 (존재 시)

### 2.5 project.yaml 불필요 필드 정리

| 필드 | 조치 |
|------|------|
| `project.type` | 제거 (미사용) |
| `github.profile_name` | 제거 (미사용) |

---

## Phase 3: 3-Way 병합 로직 개선

### 3.1 현재 문제

```
현재: mergeYAMLDeep(newTemplate, oldUserConfig)
      → old 값 항상 우선 → 템플릿 기본값 변경이 반영 안됨

예시:
  Old template default:  development_mode: "ddd"
  User value:            development_mode: "ddd"  (변경 안함)
  New template default:  development_mode: "hybrid"

  현재 결과: "ddd" (old 우선 → 새 기본값 무시)
  기대 결과: "hybrid" (사용자가 변경하지 않았으므로 새 값 적용)
```

### 3.2 개선 방안: 3-Way 병합

**핵심 아이디어:** 백업 시 원본 템플릿 기본값도 함께 저장

```
백업 시:
  1. user config (사용자 값) → .moai-backups/{ts}/sections/*.yaml
  2. template defaults (원본 기본값) → .moai-backups/{ts}/.template-defaults/sections/*.yaml
     (embedded FS에서 추출)

복원 시:
  3-way 비교: base(old template) vs old(user) vs new(new template)

  if user_value == old_template_default:
    → 사용자 미변경 → 새 템플릿 값 적용
  elif user_value != old_template_default:
    → 사용자 커스터마이징 → 사용자 값 보존
```

### 3.3 수정할 함수

1. **`backupMoaiConfig()`** (update.go:870-982)
   - embedded FS에서 원본 템플릿 추출하여 `.template-defaults/` 서브디렉토리에 저장
   - `BackupMetadata`에 `template_defaults_dir` 필드 추가

2. **`restoreMoaiConfig()`** (update.go:1157-1215)
   - 3-way 비교 로직 추가: base + old + new
   - `mergeYAMLDeep()` → `mergeYAML3Way()` 로 교체

3. **`mergeYAML3Way()`** (신규 함수)
   ```go
   func mergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error)
   ```
   - base(원본 템플릿) vs old(사용자 값) 비교로 사용자 변경 감지
   - 사용자 변경 필드만 보존, 나머지는 new 값 적용
   - system fields (template_version 등)는 항상 new 값

4. **`deepMerge3Way()`** (신규 함수)
   ```go
   func deepMerge3Way(newMap, oldMap, baseMap map[string]interface{}) map[string]interface{}
   ```
   - 재귀적 3-way 병합
   - 결정 로직:
     - old == base → 사용자 미변경 → new 사용
     - old != base → 사용자 변경 → old 보존
     - key가 new에만 존재 → new 추가 (새 필드)
     - key가 old에만 존재 → 제거 (템플릿에서 삭제된 필드)

### 3.4 Fallback

base(원본 템플릿)를 찾을 수 없는 경우 (이전 버전 백업):
- 기존 2-way 병합으로 fallback
- 로그 경고 출력

---

## Phase 4: 테스트 업데이트

### 수정할 테스트 파일

1. **`internal/cli/update_test.go`**
   - config.yaml 관련 테스트 케이스 제거/수정 (100+ 케이스)
   - 3-way 병합 테스트 추가
   - system.yaml template_version 읽기 테스트

2. **`internal/statusline/version_test.go`**
   - system.yaml에서 버전 읽기 테스트

3. **`internal/core/project/initializer_test.go`** (존재 시)
   - config.yaml 미생성 확인 테스트

---

## 실행 순서

1. Phase 1: pricing.yaml, llm.yaml 제거 (템플릿 + 로컬 + defs)
2. Phase 2: config.yaml 제거, 필드 이전, Go 코드 수정
3. Phase 3: 3-way 병합 로직 구현
4. Phase 4: 테스트 업데이트 및 검증

## 검증 방법

```bash
# 1. 빌드 확인
make build

# 2. 전체 테스트
go test -race ./... -count=1

# 3. config.yaml 참조 없음 확인
grep -r "config\.yaml" internal/ --include="*.go" | grep -v "_test.go" | grep -v "config/sections"

# 4. pricing/llm 참조 없음 확인
grep -r "pricing\|llm\.yaml" internal/ --include="*.go"

# 5. moai init 테스트 (새 프로젝트)
cd /tmp && mkdir test-init && cd test-init && moai init

# 6. moai update 테스트 (기존 프로젝트)
cd /path/to/project && moai update
```
