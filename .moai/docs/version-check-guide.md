# 버전 체크 시스템 가이드

MoAI-ADK의 자동 버전 체크 시스템에 대한 사용자 가이드입니다.

## 개요

MoAI-ADK는 세션 시작 시 자동으로 최신 버전을 확인하여 업데이트 권장 사항을 제공합니다.

### 주요 특징

- **자동 버전 체크**: Claude 세션 시작 시 자동으로 PyPI에서 최신 버전 확인
- **스마트 캐싱**: 24시간 캐시로 95% 성능 향상 (20ms vs 1초)
- **오프라인 모드 지원**: 네트워크 연결이 없어도 정상 작동
- **사용자 설정 가능**: 체크 빈도 및 활성화 여부 조정 가능

### 실행 시점

- **SessionStart 훅**: Claude 세션이 시작될 때 자동으로 실행
- **비차단 방식**: 버전 체크가 실패해도 세션은 정상 작동
- **빠른 응답**: 캐시 히트 시 ~20ms, 네트워크 쿼리 시 최대 1초

### 성능 영향

| 상황 | 응답 시간 | 네트워크 액세스 |
|-----|---------|--------------|
| 캐시 히트 (24시간 이내) | ~20ms | 없음 |
| 캐시 미스 + 온라인 | ~1초 | PyPI 쿼리 1회 |
| 캐시 미스 + 오프라인 | ~100ms | 없음 |
| 비활성화 | ~10ms | 없음 |

---

## 설정 옵션

### 기본 설정

`.moai/config.json` 파일에서 버전 체크 동작을 제어할 수 있습니다:

```json
{
  "moai": {
    "update_check_frequency": "daily",
    "version_check": {
      "enabled": true,
      "cache_ttl_hours": 24
    }
  }
}
```

### 빈도 설정 (`update_check_frequency`)

| 값 | 설명 | 캐시 TTL | 권장 용도 |
|----|------|---------|----------|
| `"always"` | 매 세션마다 체크 | 0시간 (캐시 없음) | 개발/테스트 환경 |
| `"daily"` | 하루에 한 번 체크 | 24시간 | **기본값 (권장)** |
| `"weekly"` | 일주일에 한 번 체크 | 168시간 | 안정된 프로덕션 환경 |
| `"never"` | 체크 안 함 | 무한 | 완전히 비활성화 |

### 활성화/비활성화 (`version_check.enabled`)

```json
{
  "moai": {
    "version_check": {
      "enabled": false
    }
  }
}
```

- `true`: 버전 체크 활성화 (기본값)
- `false`: 버전 체크 완전히 비활성화

---

## 사용 예시

### 예시 1: 오프라인 작업을 위해 비활성화

네트워크 연결이 없는 환경에서 작업할 때:

```json
{
  "moai": {
    "version_check": {
      "enabled": false
    }
  }
}
```

**효과**: 버전 체크를 건너뛰고 세션을 즉시 시작합니다.

### 예시 2: 주간 체크로 변경

안정된 프로젝트에서 업데이트 알림을 줄이고 싶을 때:

```json
{
  "moai": {
    "update_check_frequency": "weekly"
  }
}
```

**효과**: 7일마다 한 번만 PyPI를 쿼리합니다.

### 예시 3: 개발 환경에서 항상 최신 확인

새로운 버전을 즉시 확인하고 싶을 때:

```json
{
  "moai": {
    "update_check_frequency": "always"
  }
}
```

**효과**: 매 세션마다 PyPI를 쿼리합니다 (캐시 사용 안 함).

---

## 출력 형식

### 일반 업데이트 (패치/마이너)

```
📦 MoAI-ADK v0.8.1 installed
✨ Update available: v0.8.2
   Run: uv pip install --upgrade moai-adk>=0.8.2
```

### 메이저 버전 경고

```
📦 MoAI-ADK v0.8.1 installed
⚠️  MAJOR UPDATE AVAILABLE: v1.0.0
   Breaking changes may exist. Check release notes:
   https://github.com/modu-ai/moai-adk/releases

   Run: uv pip install --upgrade moai-adk>=1.0.0
```

### 오프라인 모드

네트워크 연결이 없을 때는 업데이트 확인을 건너뛰며, 경고 없이 현재 버전만 표시됩니다.

---

## 캐시 관리

### 캐시 위치

```
.moai/cache/version-check.json
```

### 수동 캐시 삭제

캐시를 강제로 삭제하고 즉시 새로운 버전을 확인하려면:

```bash
rm .moai/cache/version-check.json
```

다음 세션 시작 시 PyPI에서 최신 버전을 다시 쿼리합니다.

### 캐시 내용

캐시 파일은 다음 정보를 저장합니다:

- `latest`: 최신 버전 번호
- `timestamp`: 마지막 체크 시각
- `update_available`: 업데이트 가능 여부
- `is_major_update`: 메이저 버전 변경 여부
- `release_notes_url`: 릴리스 노트 URL

**보안**: 캐시는 공개 정보만 저장하며, API 키나 인증 정보는 포함되지 않습니다.

---

## 문제 해결

### Q: 버전 체크가 너무 자주 실행됩니다

**A**: `update_check_frequency`를 `"weekly"` 또는 `"never"`로 변경하세요.

### Q: 네트워크 오류가 발생합니다

**A**: 오프라인 환경이라면 `enabled: false`로 설정하여 비활성화하세요.

### Q: 캐시가 만료되지 않습니다

**A**: 수동으로 `.moai/cache/version-check.json` 파일을 삭제하세요.

### Q: 최신 버전인데 업데이트 알림이 나옵니다

**A**: 캐시가 오래된 정보를 담고 있을 수 있습니다. 캐시를 삭제하고 다시 시도하세요.

---

## 기술 세부사항

### 버전 비교 로직

- **Semantic Versioning** 사용: `major.minor.patch` 형식
- **메이저 버전 변경**: 첫 번째 숫자 증가 (0.x → 1.x, 1.x → 2.x)
- **마이너 버전 변경**: 두 번째 숫자 증가 (0.8.x → 0.9.x)
- **패치 버전 변경**: 세 번째 숫자 증가 (0.8.1 → 0.8.2)

### 타임아웃 설정

- **네트워크 확인**: 100ms 타임아웃 (Google DNS 8.8.8.8:53)
- **PyPI 쿼리**: 1초 타임아웃
- **캐시 읽기**: 타임아웃 없음 (로컬 파일)

### 오류 처리

모든 오류는 자동으로 처리되며 세션을 차단하지 않습니다:

- **네트워크 오류**: 현재 버전만 반환
- **PyPI 오류**: 캐시 사용 또는 현재 버전만 반환
- **캐시 손상**: 새로운 캐시 생성
- **설정 오류**: 기본값 사용

---

## 기본값 요약

```json
{
  "moai": {
    "update_check_frequency": "daily",
    "version_check": {
      "enabled": true,
      "cache_ttl_hours": 24
    }
  }
}
```

대부분의 사용자에게 권장되는 설정입니다. 필요에 따라 조정하세요.

---

## 관련 문서

- **SPEC**: `.moai/specs/SPEC-UPDATE-ENHANCE-001/spec.md`
- **구현**: `src/moai_adk/templates/.claude/hooks/alfred/core/project.py`
- **테스트**: `tests/unit/test_version_check_config.py`
- **설정 스키마**: `.moai/memory/CONFIG-SCHEMA.md`

---

**마지막 업데이트**: Phase 4 (Config Integration) - 2025-10-29
