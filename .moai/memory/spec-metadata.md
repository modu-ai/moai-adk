# SPEC 메타데이터 구조 가이드

> **MoAI-ADK SPEC 메타데이터 표준**
>
> 모든 SPEC 문서는 이 메타데이터 구조를 따라야 합니다.

---

## 📋 메타데이터 구조 개요

SPEC 메타데이터는 **필수 필드 7개**와 **선택 필드 9개**로 구성됩니다.

### 전체 구조 예시

```yaml
---
# 필수 필드 (7개)
id: AUTH-001                    # SPEC 고유 ID
version: 0.0.1                  # Semantic Version (v0.0.1 = INITIAL, draft 시작)
status: draft                   # draft|active|completed|deprecated
created: 2025-09-15            # 생성일 (YYYY-MM-DD)
updated: 2025-09-15            # 최종 수정일 (YYYY-MM-DD, 최초에는 created와 동일)
author: @Goos                   # 작성자 (GitHub ID, 단수형)
priority: high                  # low|medium|high|critical

# 선택 필드 - 분류/메타
category: security              # feature|bugfix|refactor|security|docs|perf
labels:                         # 분류 태그 (검색용)
  - authentication
  - jwt

# 선택 필드 - 관계 (의존성 그래프)
depends_on:                     # 의존하는 SPEC (선택)
  - USER-001
blocks:                         # 차단하는 SPEC (선택)
  - AUTH-002
related_specs:                  # 관련 SPEC (선택)
  - TOKEN-002
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"

# 선택 필드 - 범위 (영향 분석)
scope:
  packages:                     # 영향받는 패키지
    - src/core/auth
  files:                        # 핵심 파일 (선택)
    - auth-service.ts
    - jwt-manager.ts
---
```

---

## 필수 필드 (Required Fields)

### 1. `id` - SPEC 고유 ID
- **타입**: string
- **형식**: `<DOMAIN>-<NUMBER>`
- **예시**: `AUTH-001`, `INSTALLER-SEC-001`
- **규칙**:
  - 영구 불변 (한 번 부여하면 변경 불가)
  - 3자리 숫자 사용 (001~999)
  - 도메인은 대문자, 하이픈 사용 가능
  - 디렉토리명: `.moai/specs/SPEC-{ID}/` (예: `.moai/specs/SPEC-AUTH-001/`)

### 2. `version` - 버전
- **타입**: string (Semantic Version)
- **형식**: `MAJOR.MINOR.PATCH`
- **기본값**: `0.0.1` (모든 SPEC 시작 버전, status: draft)
- **버전 체계**:
  - **v0.0.1**: INITIAL - SPEC 최초 작성 (status: draft)
  - **v0.0.x**: Draft 수정/개선 (SPEC 문서 수정 시 패치 버전 증가)
  - **v0.1.0**: TDD 구현 완료 (status: completed, /alfred:3-sync 자동 업데이트)
  - **v0.1.x**: 버그 수정, 문서 개선 (패치 버전)
  - **v0.x.0**: 기능 추가, 주요 개선 (마이너 버전)
  - **v1.0.0**: 정식 안정화 버전 (프로덕션 준비, 사용자 명시적 승인 필수)

### 3. `status` - 진행 상태
- **타입**: enum
- **가능한 값**:
  - `draft`: 초안 작성 중
  - `active`: 구현 진행 중
  - `completed`: 구현 완료
  - `deprecated`: 사용 중지 예정

### 4. `created` - 생성일
- **타입**: date (string)
- **형식**: `YYYY-MM-DD`
- **예시**: `2025-10-06`

### 5. `updated` - 최종 수정일
- **타입**: date (string)
- **형식**: `YYYY-MM-DD`
- **규칙**: SPEC 내용 수정 시마다 업데이트

### 6. `author` - 작성자
- **타입**: string
- **형식**: `@{GitHub ID}`
- **예시**: `@Goos`
- **규칙**:
  - 단수형 사용 (~~authors 배열 사용하지 않음~~)
  - GitHub ID 앞에 @ 접두사 필수
  - 복수 작성자는 HISTORY 섹션에 기록

### 7. `priority` - 우선순위
- **타입**: enum
- **가능한 값**:
  - `critical`: 즉시 처리 필요 (보안, 중대 버그)
  - `high`: 높은 우선순위 (주요 기능)
  - `medium`: 중간 우선순위 (개선사항)
  - `low`: 낮은 우선순위 (최적화, 문서)

---

## 선택 필드 (Optional Fields)

### 분류/메타 필드

#### 8. `category` - 변경 유형
- **타입**: enum
- **가능한 값**:
  - `feature`: 새 기능 추가
  - `bugfix`: 버그 수정
  - `refactor`: 리팩토링
  - `security`: 보안 개선
  - `docs`: 문서화
  - `perf`: 성능 최적화

#### 9. `labels` - 분류 태그
- **타입**: array of string
- **용도**: 검색, 필터링, 그루핑
- **예시**:
  ```yaml
  labels:
    - installer
    - template
    - security
  ```

### 관계 필드 (Dependency Graph)

#### 10. `depends_on` - 의존 SPEC
- **타입**: array of string
- **의미**: 이 SPEC이 완료되려면 먼저 완료되어야 하는 SPEC 목록
- **예시**:
  ```yaml
  depends_on:
    - USER-001
    - AUTH-001
  ```
- **활용**: 작업 순서 결정, 병렬 작업 가능 여부 판단

#### 11. `blocks` - 차단 SPEC
- **타입**: array of string
- **의미**: 이 SPEC으로 인해 차단된 SPEC 목록
- **예시**:
  ```yaml
  blocks:
    - PAYMENT-003
  ```

#### 12. `related_specs` - 관련 SPEC
- **타입**: array of string
- **의미**: 직접적 의존성은 없지만 관련된 SPEC 목록
- **예시**:
  ```yaml
  related_specs:
    - TOKEN-002
    - SESSION-001
  ```

#### 13. `related_issue` - 관련 GitHub Issue
- **타입**: string (URL)
- **형식**: GitHub Issue 전체 URL
- **예시**:
  ```yaml
  related_issue: "https://github.com/modu-ai/moai-adk/issues/123"
  ```

### 범위 필드 (Scope/Impact)

#### 14. `scope.packages` - 영향받는 패키지
- **타입**: array of string
- **의미**: 이 SPEC이 영향을 주는 패키지/모듈 경로
- **예시**:
  ```yaml
  scope:
    packages:
      - moai-adk-ts/src/core/installer
      - moai-adk-ts/src/core/git
  ```

#### 15. `scope.files` - 핵심 파일
- **타입**: array of string
- **의미**: 주요 변경 대상 파일 (참고용)
- **예시**:
  ```yaml
  scope:
    files:
      - template-processor.ts
      - template-security.ts
  ```

---

## 메타데이터 검증

### 필수 필드 검증
```bash
# 모든 SPEC 파일에 필수 필드가 있는지 확인
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# priority 필드 누락 확인
rg -L "^priority:" .moai/specs/SPEC-*/spec.md
```

### 형식 검증
```bash
# author 필드 형식 확인 (@Goos 형식)
rg "^author: @[A-Z]" .moai/specs/SPEC-*/spec.md

# version 필드 형식 확인 (0.x.y)
rg "^version: 0\.\d+\.\d+" .moai/specs/SPEC-*/spec.md
```

---

## 마이그레이션 가이드

### 기존 SPEC 업데이트

#### 1. priority 필드 추가
기존 SPEC에 priority 필드가 없다면 추가:
```yaml
priority: medium  # 또는 low|high|critical
```

#### 2. author 필드 표준화
- `authors: ["@goos"]` → `author: @Goos`
- 소문자 → 대문자로 변경

#### 3. 선택 필드 추가 (권장)
```yaml
category: refactor
labels:
  - code-quality
  - maintenance
```

---

## 설계 원칙

### 1. DRY (Don't Repeat Yourself)
- ❌ **제거**: `reference` 필드 (모든 SPEC이 같은 masterplan 참조 → 중복)
- ✅ **대안**: README.md에 프로젝트 레벨 문서 명시

### 2. Context-Aware
- 필요한 컨텍스트만 포함
- 선택 필드는 실제 필요할 때만 사용

### 3. Traceable
- `depends_on`, `blocks`, `related_specs`로 SPEC 간 의존성 명시
- 자동화 도구로 순환 의존성 검증 가능

### 4. Maintainable
- 모든 필드는 자동화 도구로 검증 가능
- 일관된 형식으로 파싱 용이

### 5. Simple First
- 복잡도 최소화
- 필수 7개 + 선택 9개로 제한
- 점진적 확장 가능

---

**최종 업데이트**: 2025-10-06
**작성자**: @Alfred