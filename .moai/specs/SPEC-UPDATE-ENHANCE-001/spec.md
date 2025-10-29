---
id: UPDATE-ENHANCE-001
version: 0.0.1
status: draft
created: 2025-10-29
updated: 2025-10-29
author: @goos
priority: high
category: infrastructure,performance
labels: [version-check, caching, offline-support, user-preferences]
scope: SessionStart hook enhancement with 6 improvements
---

# SPEC-UPDATE-ENHANCE-001: MoAI-ADK SessionStart 버전 체크 시스템 강화

@SPEC:UPDATE-ENHANCE-001

## HISTORY

### v0.0.1 (2025-10-29)
- **INITIAL**: SessionStart 버전 체크 시스템 강화 (SPEC-UPDATE-ENHANCE-001)
- **AUTHOR**: @goos
- **SCOPE**: PyPI 버전 체크, 캐싱, 오프라인 감지, 사용자 설정, 릴리스 노트, Major 버전 경고
- **CONTEXT**: 사용자 경험 개선 및 세션 시작 성능 최적화 (95% 향상)

---

## 1. 환경 (Environment)

### 1.1 현재 시스템 환경

**파일 구조**:
```
.claude/hooks/alfred/
├── handlers/
│   └── session.py              # SessionStart 이벤트 핸들러
└── core/
    └── project.py              # 버전 체크 로직 (get_package_version_info)
```

**현재 버전 체크 동작**:
- `get_package_version_info()` 함수가 PyPI API 호출 (1초 타임아웃)
- 매 세션마다 무조건 PyPI 접근 시도
- 네트워크 실패 시 graceful degradation으로 건너뛰기
- `uv pip install --upgrade moai-adk` 명령어 표시

**성능 현황**:
- SessionStart 평균 소요 시간: 1.2초 (PyPI 체크 포함)
- PyPI 타임아웃 비율: 약 5% (네트워크 환경에 따라 변동)
- 사용자 피드백: "매번 체크하는 것이 부담스럽다"

### 1.2 기술 스택

- **언어**: Python 3.11+
- **패키지 관리**: uv (UV tool)
- **버전 체크**: PyPI JSON API (`https://pypi.org/pypi/moai-adk/json`)
- **메타데이터**: importlib.metadata
- **네트워크**: urllib.request (표준 라이브러리)
- **캐싱**: 파일 기반 (.moai/cache/version-check.json)

### 1.3 통합 포인트

- SessionStart Hook → `handle_session_start()` → `get_package_version_info()`
- Config 시스템 → `.moai/config.json` (사용자 설정 추가)
- 캐시 디렉토리 → `.moai/cache/` (신규 생성)

---

## 2. 가정 (Assumptions)

### 2.1 사용자 환경 가정

- ✅ 사용자는 인터넷 연결이 불안정하거나 오프라인 환경에서도 작업 가능
- ✅ 대부분의 사용자는 최신 버전 알림을 원하지만, 매번 체크는 원하지 않음
- ✅ Major 버전 업그레이드(breaking changes)는 사전 경고가 필요
- ✅ 사용자는 명시적인 릴리스 노트를 확인하고 싶어함

### 2.2 기술적 가정

- ✅ PyPI API는 공개 API이며 별도 인증 없이 접근 가능
- ✅ 캐시 파일은 `.moai/cache/` 디렉토리에 안전하게 저장 가능
- ✅ uv 툴체인이 시스템에 설치되어 있음 (MoAI-ADK 설치 전제 조건)
- ✅ 파일 시스템 쓰기 권한이 프로젝트 디렉토리에 있음

### 2.3 제약 조건

- ⚠️ SessionStart 총 지연 시간은 **3초 이내**로 유지 필수
- ⚠️ 캐시 파일 손상 시에도 정상 동작해야 함 (graceful degradation)
- ⚠️ 오프라인 환경에서도 에러 없이 건너뛰어야 함
- ⚠️ 기존 SessionStart 동작을 변경하지 않아야 함 (하위 호환성)

---

## 3. 요구사항 (Requirements)

### 3.1 보편적 요구사항 (Ubiquitous Requirements)

@REQ:UPDATE-ENHANCE-001-U1
**R-U1**: 시스템은 **모든 SessionStart 이벤트**에서 버전 정보를 표시해야 한다.
- **현재**: `🗿 MoAI-ADK Ver: 0.8.1`
- **업데이트 가능 시**: `🗿 MoAI-ADK Ver: 0.8.1 → 0.9.0 available ✨`

@REQ:UPDATE-ENHANCE-001-U2
**R-U2**: 시스템은 **24시간 캐시**를 통해 불필요한 PyPI 요청을 줄여야 한다.
- 캐시 파일 위치: `.moai/cache/version-check.json`
- 캐시 유효 기간: 24시간 (86400초)
- 캐시 만료 시 자동 갱신

@REQ:UPDATE-ENHANCE-001-U3
**R-U3**: 시스템은 **uv tool upgrade** 명령어를 권장해야 한다.
- **변경 전**: `uv pip install --upgrade moai-adk>=0.9.0`
- **변경 후**: `uv tool upgrade moai-adk` (uv 공식 권장 방식)

@REQ:UPDATE-ENHANCE-001-U4
**R-U4**: 시스템은 **오프라인 환경**을 감지하고 네트워크 요청을 건너뛰어야 한다.
- 네트워크 가용성 사전 체크 (DNS 조회 또는 간단한 핑)
- 오프라인 감지 시 캐시 데이터 활용
- 캐시도 없으면 "버전 확인 불가" 표시 없이 건너뛰기

@REQ:UPDATE-ENHANCE-001-U5
**R-U5**: 시스템은 **사용자 설정**을 통해 체크 빈도를 조정할 수 있어야 한다.
- Config 옵션: `version_check.frequency` (never, daily, weekly, always)
- 기본값: `daily` (24시간 캐시 사용)
- `never` 설정 시 버전 체크 완전 비활성화

### 3.2 이벤트 기반 요구사항 (Event-driven Requirements)

@REQ:UPDATE-ENHANCE-001-E1
**R-E1**: **WHEN** SessionStart 이벤트가 발생하고 캐시가 만료되었을 때,
**THEN** 시스템은 PyPI API를 호출하고 새 버전 정보를 캐시에 저장해야 한다.
- 캐시 만료 확인: 파일 수정 시간 + 24시간 < 현재 시간
- API 호출 타임아웃: 1초 (기존 유지)
- 성공 시 캐시 파일 갱신

@REQ:UPDATE-ENHANCE-001-E2
**R-E2**: **WHEN** Major 버전 업데이트가 감지되었을 때 (예: 0.8.x → 1.0.0),
**THEN** 시스템은 **경고 메시지**와 함께 릴리스 노트 링크를 표시해야 한다.
```
⚠️  Major version update available: 0.8.1 → 1.0.0
   Breaking changes detected. Review release notes before upgrading:
   📝 https://github.com/modu-ai/moai-adk/releases/tag/v1.0.0
   ⬆️ Upgrade: uv tool upgrade moai-adk
```

@REQ:UPDATE-ENHANCE-001-E3
**R-E3**: **WHEN** 네트워크가 오프라인 상태이거나 PyPI 접근 실패 시,
**THEN** 시스템은 에러 로그 없이 캐시 데이터를 활용하거나 건너뛰어야 한다.
- 오프라인 감지 로직: `socket.create_connection()` 테스트 (0.1초 타임아웃)
- 캐시 우선 활용 (stale cache 허용)
- 로그 레벨: DEBUG (사용자에게 노출하지 않음)

### 3.3 상태 기반 요구사항 (State-driven Requirements)

@REQ:UPDATE-ENHANCE-001-S1
**R-S1**: **WHILE** 캐시가 유효한 동안 (24시간 이내),
**THEN** 시스템은 PyPI API 호출을 건너뛰고 캐시된 버전 정보를 즉시 표시해야 한다.
- 성능 목표: SessionStart 지연 시간 **50ms 이하** (95% 개선)
- 캐시 히트율 목표: 95% 이상

@REQ:UPDATE-ENHANCE-001-S2
**R-S2**: **WHILE** 사용자가 `version_check.frequency = "never"` 설정을 유지하는 동안,
**THEN** 시스템은 버전 체크를 완전히 비활성화하고 현재 버전만 표시해야 한다.
```
🗿 MoAI-ADK Ver: 0.8.1
```
(업데이트 알림 및 upgrade 명령어 미표시)

### 3.4 선택적 요구사항 (Optional Requirements)

@REQ:UPDATE-ENHANCE-001-O1
**R-O1**: 시스템은 **릴리스 노트 URL**을 자동으로 생성하여 표시할 수 있다.
- GitHub Releases 링크: `https://github.com/modu-ai/moai-adk/releases/tag/v{latest_version}`
- PyPI API에서 `project_urls` 필드 활용 (가능한 경우)

@REQ:UPDATE-ENHANCE-001-O2
**R-O2**: 시스템은 **버전 비교 로직**을 강화하여 pre-release 버전을 구분할 수 있다.
- Stable 버전 우선 (0.9.0 > 0.9.0rc1)
- Pre-release 버전은 별도 라벨 표시 (`0.9.0-rc1 available (pre-release)`)

---

## 4. 제약사항 (Constraints)

@CONSTRAINT:UPDATE-ENHANCE-001-C1
**C-1 성능 제약**: SessionStart 총 실행 시간은 **3초를 초과할 수 없다**.
- 현재: ~1.2초 (PyPI 체크 포함)
- 목표: ~0.05초 (캐시 히트 시), ~1.2초 (캐시 미스 시)
- 측정: 통합 테스트에서 타이밍 검증

@CONSTRAINT:UPDATE-ENHANCE-001-C2
**C-2 하위 호환성**: 기존 SessionStart 출력 형식을 유지해야 한다.
- 버전 정보 라인은 기존 위치 유지 (Language 정보 위)
- 추가 정보는 별도 라인으로 표시 (기존 출력 변경 금지)

@CONSTRAINT:UPDATE-ENHANCE-001-C3
**C-3 보안 제약**: 캐시 파일은 민감 정보를 포함하지 않아야 한다.
- 저장 데이터: 버전 번호, 체크 시간, 릴리스 노트 URL만 포함
- 사용자 정보, 토큰, API 키 등 저장 금지
- 파일 권한: 644 (읽기 전용)

@CONSTRAINT:UPDATE-ENHANCE-001-C4
**C-4 에러 처리**: 모든 실패 시나리오에서 graceful degradation 필수.
- 캐시 파일 손상 → 무시하고 재생성
- PyPI API 실패 → 캐시 활용 또는 건너뛰기
- 설정 파일 오류 → 기본값(daily) 사용
- **절대 SessionStart 실패로 이어지지 않아야 함**

---

## 5. 상세 명세 (Specifications)

### 5.1 캐시 파일 구조

**파일 경로**: `.moai/cache/version-check.json`

```json
{
  "current_version": "0.8.1",
  "latest_version": "0.9.0",
  "checked_at": "2025-10-29T15:30:00+09:00",
  "is_major_update": false,
  "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v0.9.0",
  "upgrade_command": "uv tool upgrade moai-adk"
}
```

**필드 설명**:
- `current_version`: 설치된 버전 (importlib.metadata 기준)
- `latest_version`: PyPI 최신 버전
- `checked_at`: ISO 8601 형식 타임스탬프
- `is_major_update`: Major 버전 변경 여부 (boolean)
- `release_notes_url`: GitHub Releases 링크
- `upgrade_command`: 권장 업그레이드 명령어

### 5.2 Config 확장

**`.moai/config.json` 추가 필드**:

```json
{
  "version_check": {
    "enabled": true,
    "frequency": "daily",
    "cache_ttl_hours": 24,
    "show_release_notes": true,
    "warn_major_updates": true
  }
}
```

**필드 설명**:
- `enabled`: 버전 체크 활성화 여부 (기본: true)
- `frequency`: 체크 빈도 (`never`, `daily`, `weekly`, `always`)
- `cache_ttl_hours`: 캐시 유효 시간 (기본: 24)
- `show_release_notes`: 릴리스 노트 URL 표시 여부 (기본: true)
- `warn_major_updates`: Major 업데이트 경고 표시 (기본: true)

### 5.3 오프라인 감지 로직

**네트워크 가용성 체크** (0.1초 타임아웃):

```python
def is_network_available() -> bool:
    """Check if network is available by testing PyPI connectivity"""
    try:
        import socket
        with timeout_handler(0.1):
            socket.create_connection(("pypi.org", 443), timeout=0.1)
        return True
    except (OSError, TimeoutError):
        return False
```

### 5.4 Major 버전 감지 로직

```python
def is_major_update(current: str, latest: str) -> bool:
    """Detect major version changes (e.g., 0.8.1 → 1.0.0)"""
    try:
        current_major = int(current.split(".")[0])
        latest_major = int(latest.split(".")[0])
        return latest_major > current_major
    except (ValueError, IndexError):
        return False
```

### 5.5 업데이트된 SessionStart 출력 예시

**케이스 1: 캐시 히트 + Minor 업데이트**
```
🚀 MoAI-ADK Session Started

   🗿 MoAI-ADK Ver: 0.8.1 → 0.8.2 available ✨
   📝 Release Notes: https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2
   ⬆️ Upgrade: uv tool upgrade moai-adk
   🐍 Language: python
   ...
```

**케이스 2: Major 업데이트 경고**
```
🚀 MoAI-ADK Session Started

   ⚠️  Major version update available: 0.8.1 → 1.0.0
   Breaking changes detected. Review release notes before upgrading:
   📝 https://github.com/modu-ai/moai-adk/releases/tag/v1.0.0
   ⬆️ Upgrade: uv tool upgrade moai-adk
   🐍 Language: python
   ...
```

**케이스 3: 오프라인 (캐시 활용)**
```
🚀 MoAI-ADK Session Started

   🗿 MoAI-ADK Ver: 0.8.1 (cached)
   🐍 Language: python
   ...
```

**케이스 4: 버전 체크 비활성화**
```
🚀 MoAI-ADK Session Started

   🗿 MoAI-ADK Ver: 0.8.1
   🐍 Language: python
   ...
```

---

## 6. 추적성 (Traceability)

### 6.1 TAG 체인

```
@SPEC:UPDATE-ENHANCE-001
  ├─ @REQ:UPDATE-ENHANCE-001-U1 (버전 정보 표시)
  ├─ @REQ:UPDATE-ENHANCE-001-U2 (24시간 캐싱)
  ├─ @REQ:UPDATE-ENHANCE-001-U3 (uv tool upgrade 명령어)
  ├─ @REQ:UPDATE-ENHANCE-001-U4 (오프라인 감지)
  ├─ @REQ:UPDATE-ENHANCE-001-U5 (사용자 설정)
  ├─ @REQ:UPDATE-ENHANCE-001-E1 (캐시 갱신 이벤트)
  ├─ @REQ:UPDATE-ENHANCE-001-E2 (Major 업데이트 경고)
  ├─ @REQ:UPDATE-ENHANCE-001-E3 (네트워크 실패 처리)
  ├─ @REQ:UPDATE-ENHANCE-001-S1 (캐시 유효 기간)
  ├─ @REQ:UPDATE-ENHANCE-001-S2 (버전 체크 비활성화)
  ├─ @REQ:UPDATE-ENHANCE-001-O1 (릴리스 노트 URL)
  └─ @REQ:UPDATE-ENHANCE-001-O2 (Pre-release 구분)
```

### 6.2 구현 예정 TAG

```
@CODE:VERSION-CACHE-001        # 캐시 관리 모듈
@CODE:NETWORK-DETECT-001       # 네트워크 감지 로직
@CODE:VERSION-COMPARE-001      # 버전 비교 로직
@TEST:VERSION-CACHE-001        # 캐싱 단위 테스트
@TEST:OFFLINE-MODE-001         # 오프라인 시나리오 테스트
@TEST:MAJOR-UPDATE-WARN-001    # Major 업데이트 경고 테스트
@DOC:VERSION-CHECK-CONFIG-001  # 사용자 설정 가이드
```

### 6.3 기존 TAG 참조

- `@TAG:HOOKS-TIMEOUT-001`: 타임아웃 처리 패턴 재사용
- `@CODE:GITHUB-CONFIG-001`: Config 확장 패턴 참조
- `@SPEC:MOAI-CONFIG-001`: Config 스키마 업데이트

---

## 7. 성공 지표 (Success Metrics)

### 7.1 성능 지표

| 항목 | 현재 | 목표 | 측정 방법 |
|------|------|------|-----------|
| 캐시 히트 시 지연 | N/A | < 50ms | 단위 테스트 타이밍 |
| 캐시 미스 시 지연 | ~1.2s | < 1.5s | PyPI 체크 포함 |
| 캐시 히트율 | 0% | > 95% | 7일간 사용자 로그 분석 |
| SessionStart 실패율 | < 0.1% | < 0.01% | 에러 로그 모니터링 |

### 7.2 사용자 경험 지표

- **빠른 세션 시작**: 95% 사용자가 50ms 이내 버전 정보 확인
- **명확한 업그레이드 경로**: 릴리스 노트 클릭률 > 20%
- **오프라인 환경 지원**: 네트워크 오류로 인한 불만 제로
- **사용자 만족도**: GitHub 이슈에서 "버전 체크 개선" 피드백 증가

### 7.3 코드 품질 지표

- **테스트 커버리지**: 95% 이상 (캐싱, 오프라인, Major 업데이트 시나리오)
- **에러 핸들링**: 모든 실패 케이스에 대한 단위 테스트 존재
- **문서화**: 사용자 가이드 및 개발자 문서 완비

---

## 8. 리스크 및 대응 전략

### 8.1 기술적 리스크

**R-1: 캐시 파일 손상**
- **확률**: Medium
- **영향**: Low (graceful degradation으로 복구)
- **대응**: JSON 파싱 실패 시 캐시 무시 + 재생성

**R-2: PyPI API 변경**
- **확률**: Low
- **영향**: Medium (버전 정보 표시 불가)
- **대응**: API 응답 스키마 검증 + fallback 로직

**R-3: 네트워크 타임아웃 증가**
- **확률**: Medium (일부 환경)
- **영향**: Low (타임아웃 설정으로 제한)
- **대응**: 0.1초 타임아웃 유지 + 오프라인 모드로 전환

### 8.2 운영 리스크

**R-4: 사용자 설정 오류**
- **확률**: Medium
- **영향**: Low (기본값으로 복구)
- **대응**: Config 스키마 검증 + 기본값 fallback

**R-5: 캐시 디렉토리 쓰기 권한 없음**
- **확률**: Low
- **영향**: Medium (캐싱 불가)
- **대응**: 권한 오류 시 메모리 캐시로 대체 (세션 내 유효)

---

## 9. 롤백 계획

### 9.1 단계별 롤백

**Phase 1: 기능 비활성화**
```json
{
  "version_check": {
    "enabled": false
  }
}
```
→ 기존 동작으로 복귀 (매번 PyPI 체크)

**Phase 2: 코드 롤백**
- `get_package_version_info()` 함수를 이전 버전으로 복원
- 캐시 관련 코드 제거

**Phase 3: 캐시 데이터 정리**
```bash
rm -rf .moai/cache/version-check.json
```

### 9.2 긴급 대응 시나리오

**시나리오 1: SessionStart 실패율 급증**
- 즉시 `version_check.enabled = false` 배포
- 로그 분석 후 원인 파악

**시나리오 2: 사용자 불만 급증**
- 캐시 TTL 조정 (24시간 → 7일)
- 또는 기본값을 `never`로 변경

---

## 10. 다음 단계

### 10.1 구현 우선순위

1. **P0 (필수)**: 24시간 캐싱 + `uv tool upgrade` 명령어 변경
2. **P1 (중요)**: 오프라인 감지 + 사용자 설정
3. **P2 (권장)**: Major 버전 경고 + 릴리스 노트 URL

### 10.2 후속 작업

- [ ] `.moai/memory/config-schema.md` 업데이트 (version_check 섹션 추가)
- [ ] 사용자 가이드 작성 (버전 체크 설정 방법)
- [ ] 통합 테스트 시나리오 작성 (E2E)
- [ ] 성능 벤치마크 스크립트 작성

---

**END OF SPEC**

@SPEC:UPDATE-ENHANCE-001
