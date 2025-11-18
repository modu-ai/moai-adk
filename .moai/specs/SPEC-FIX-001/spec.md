---
id: SPEC-FIX-001
version: "1.0.0"
status: completed
created: "2025-11-18"
updated: "2025-11-18"
completed: "2025-11-18"
author: Claude Code
priority: high
implementation_status: done
---

# SPEC-FIX-001: Statusline 복구 - Ver unknown 이슈 해결

## HISTORY

| 버전 | 날짜 | 작성자 | 변경사항 |
|------|------|--------|---------|
| 1.0.0 | 2025-11-18 | Claude Code | 초안 작성 - 상태표시줄 마이그레이션 이슈 분석 및 복구 방안 수립 |

---

## 문제 정의

### 증상
- Claude Code 하단 상태표시줄에 "Ver unknown" 표시됨
- 예상 출력: `🤖 Haiku 4.5 | 🗿 Ver 0.26.0 | 📊 +0 M26 ?9 | 🔀 release/0.26.0`
- 실제 출력: `🤖 Haiku 4.5 | 🗿 Ver unknown | 📊 +0 M26 ?9 | 🔀 release/0.26.0`

### 근본 원인
1. statusline.py 스크립트가 `.moai/scripts/statusline.py`에서 삭제됨 (의도된 변경)
2. 패키지 CLI 명령어 `uvx moai-adk statusline`으로 마이그레이션 완료 (Commit 05b98e56)
3. 환경 문제:
   - uvx 캐시에 이전 버전 보유 가능성
   - config.json 파일 경로 해석 오류
   - 패키지 import 실패

### 영향 범위
- 개발자 경험 저하 (프로젝트 버전 정보 미표시)
- 작업 진행 중 혼동 가능성 (현재 버전 불명확)
- Claude Code 상태표시줄 신뢰도 감소

### 긴급도
**HIGH** - 개발 생산성에 직접적 영향

---

## EARS 요구사항

### Ubiquitous (항상 참이어야 하는 요구사항)

#### U1: uvx 환경 설정
**The system SHALL** 설치된 uvx 환경을 올바르게 인식하고 모든 의존성이 정상 로드되어야 함.
- uvx 버전: 0.5.0 이상
- 활성 Python: 3.13.9 이상
- moai-adk 패키지: 설치 완료 상태

#### U2: config.json 버전 관리
**The system SHALL** `.moai/config/config.json` 파일에서 `moai.version` 필드를 정확하게 읽어 최신 버전 정보를 유지해야 함.
- 파일 위치: `.moai/config/config.json`
- 필수 필드: `moai.version`
- 현재 버전: "0.26.0"
- 인코딩: UTF-8

#### U3: CLI 명령어 실행 가능
**The system SHALL** `uvx moai-adk statusline` 명령어를 언제든지 실행할 수 있어야 하며, 패키지 버전과 무관하게 최신 설치 버전을 사용해야 함.
- 명령어: `uvx moai-adk statusline`
- 캐시 정책: 기본 설정 준수
- 오류 발생 시: graceful fallback 제공

---

### Event-Driven (이벤트 기반 요구사항)

#### ED1: Claude Code 세션 초기화 시 statusline 업데이트
**WHEN** Claude Code가 시작되고 SessionStart 훅이 실행될 때
**THEN** `uvx moai-adk statusline` 명령어를 실행하여 버전 정보를 구성하고
**AND** 상태표시줄에 정확한 버전 번호(예: Ver 0.26.0)를 표시함.

```
Given: Claude Code가 새로운 세션을 시작한 상태
When: SessionStart 훅이 실행됨
Then: 버전 정보가 정확하게 표시됨
  And: "Ver unknown"이 절대 표시되지 않음
```

#### ED2: 패키지 버전 변경 감지
**WHEN** moai-adk 패키지 버전이 업데이트될 때
**THEN** config.json의 버전 정보도 함께 업데이트되어야 하고
**AND** statusline이 새로운 버전을 반영함.

```
Given: 새로운 버전의 moai-adk가 설치된 상태
When: uvx moai-adk statusline 실행
Then: 최신 버전 정보가 표시됨
```

#### ED3: 캐시 오염 시 자동 복구
**WHEN** uvx 캐시에 이전 버전의 패키지가 남아있을 때
**THEN** 캐시를 자동으로 클리어하거나 사용자에게 안내하고
**AND** 최신 버전으로 재실행함.

```
Given: uvx 캐시에 이전 버전 보유 상태
When: uvx moai-adk statusline 실행
Then: 캐시 클리어 후 최신 버전 사용
  Or: 사용자 가이드 제공
```

---

### Unwanted (피해야 할 부정적 시나리오)

#### UW1: 버전 정보 누락 방지
**IF** config.json 파일을 찾을 수 없거나 버전 필드가 없을 때
**THEN** "Ver unknown" 또는 에러 메시지를 표시하지 말고 대신
**AND** graceful fallback으로 기본값(예: "Ver dev")을 표시하거나
**AND** 개발자 로그에만 에러를 기록함.

```
Given: config.json 파일이 없는 상태
When: statusline 명령어 실행
Then: 에러 메시지 표시 금지
  And: 대체값 또는 로깅만 수행
  And: Claude Code 기능 중단 없음
```

#### UW2: 순환 의존성 방지
**IF** statusline 명령어 실행 중 패키지 import 실패가 발생할 때
**THEN** 무한 루프에 빠지지 말고
**AND** 재시도 횟수를 제한하며
**AND** 최대 3초 타임아웃을 적용함.

```
Given: 패키지 import 실패 상황
When: statusline 명령어 실행
Then: 최대 3번 재시도
  And: 3초 타임아웃 적용
  And: fallback 표시
```

#### UW3: 성능 저하 방지
**IF** statusline 초기화로 인해 Claude Code 시작 시간이 5초 이상 증가할 때
**THEN** 백그라운드 작업으로 변경하거나
**AND** 캐시를 활용하여 응답 시간을 2초 이내로 유지함.

```
Given: statusline 성능 오버헤드 발생
When: Claude Code 시작
Then: 응답 시간 2초 이내 유지
  And: 백그라운드 업데이트 고려
```

---

### State-Driven (상태 유지 요구사항)

#### SD1: 세션 중 상태 일관성
**WHILE** Claude Code 세션이 활성화되어 있는 동안
**THEN** statusline은 항상 정상 상태를 유지해야 하고
**AND** Git 상태 정보(변경사항, 브랜치)와 version 정보가 일관성 있게 표시됨.

```
Given: Claude Code 세션 진행 중
When: 여러 번의 파일 변경 작업 수행
Then: 매번 statusline 업데이트
  And: 버전 정보는 변하지 않음
  And: Git 상태만 갱신됨
```

#### SD2: 멀티 세션 동시성
**WHILE** 여러 Claude Code 세션이 동시에 실행 중일 때
**THEN** 각 세션의 statusline이 독립적으로 정상 작동하고
**AND** 캐시 충돌로 인한 버전 표시 오류가 없어야 함.

```
Given: 2개 이상의 Claude Code 세션 활성화
When: 동시에 여러 statusline 조회
Then: 각 세션이 독립적으로 동작
  And: 캐시 충돌 없음
```

#### SD3: 버전 변경 추적
**WHILE** 프로젝트 버전이 유지되는 동안
**THEN** config.json의 버전 변경이 감지되어야 하고
**AND** 다음 statusline 실행 시 새로운 버전을 반영함.

```
Given: config.json의 버전이 0.26.0 → 0.27.0으로 변경
When: statusline 다시 실행
Then: 즉시 새 버전(0.27.0) 표시
```

---

### Optional (사용자 선택 요구사항)

#### OP1: 상태표시줄 표시 활성화
**WHERE** 사용자가 Claude Code 상태표시줄 표시를 활성화했을 때
**THEN** 모든 정보(모델, 버전, Git 상태, 브랜치)를 표시하고
**AND** statusline이 정상 작동하지 않으면 명확한 대체값을 표시함.

```
Given: statusLine.enabled = true in settings.json
When: Claude Code 시작
Then: 모든 상태 정보 표시
  And: "Ver unknown" 절대 표시 금지
```

#### OP2: 캐시 관리 옵션
**WHERE** 개발자가 uvx 캐시 관리를 명시적으로 수행하려고 할 때
**THEN** `uvx cache clear moai-adk` 또는 `uv sync --force` 등의 명령어가 정상 작동하고
**AND** 이후 statusline이 최신 버전을 반영함.

```
Given: 개발자가 캐시 클리어 실행
When: uvx cache clear moai-adk 실행
Then: 캐시 완전 삭제
  And: 다음 statusline 실행 시 최신 버전
```

#### OP3: 성능 최적화 옵션
**WHERE** 사용자가 성능을 우선시할 때
**THEN** 캐시된 버전 정보를 사용하거나
**AND** 비동기 방식으로 업데이트하여 응답 시간을 최소화할 수 있음.

```
Given: performance_priority = high in config
When: statusline 실행
Then: 캐시된 정보 우선 사용
  And: 백그라운드에서 업데이트
```

---

## 기술 스택

| 항목 | 버전 | 비고 |
|------|------|------|
| Python | 3.13.9+ | 패키지 구현 언어 |
| moai-adk | 0.26.0+ | 패키지 CLI 호스트 |
| uvx | 0.5.0+ | 패키지 실행 도구 |
| uv | 최신 | 의존성 관리 |
| Claude Code | v4.0+ | 실행 환경 |

---

## 성능 요구사항

| 지표 | 목표 | 비고 |
|------|------|------|
| 초기 실행 | 2초 이내 | uvx 부팅 포함 |
| 캐시 히트 | 1초 이내 | 2차 이상 실행 |
| 타임아웃 | 3초 | 무한 루프 방지 |
| 메모리 | < 50MB | 오버헤드 최소화 |

---

## 추적 가능성 (Traceability)

```
SPEC-FIX-001
  ├─ 요구사항 (EARS)
  │  ├─ Ubiquitous: U1, U2, U3
  │  ├─ Event-Driven: ED1, ED2, ED3
  │  ├─ Unwanted: UW1, UW2, UW3
  │  ├─ State-Driven: SD1, SD2, SD3
  │  └─ Optional: OP1, OP2, OP3
  ├─ 구현 (TDD Red-Green-Refactor)
  │  └─ src/moai_adk/statusline/main.py
  ├─ 테스트
  │  ├─ tests/test_statusline_recovery.py
  │  └─ tests/test_config_loading.py
  └─ 문서
     ├─ docs/statusline-guide.md
     └─ docs/troubleshooting-statusline.md
```

---

## 수용 기준 (Acceptance Criteria)

- [ ] U1: uvx 환경이 정상 인식됨
- [ ] U2: config.json에서 버전 읽기 성공
- [ ] U3: CLI 명령어 실행 가능
- [ ] ED1: SessionStart 훅에서 정확한 버전 표시
- [ ] ED2: 패키지 업데이트 시 버전 반영
- [ ] ED3: 캐시 오염 시 자동 또는 수동 복구
- [ ] UW1: "Ver unknown" 표시 금지
- [ ] UW2: 순환 의존성 및 타임아웃 제한
- [ ] UW3: 성능 저하 없음 (2초 이내)
- [ ] SD1: 세션 중 상태 일관성
- [ ] SD2: 멀티 세션 동시성 보장
- [ ] SD3: 버전 변경 추적
- [ ] OP1: 상태표시줄 정상 표시
- [ ] OP2: 캐시 관리 옵션 작동
- [ ] OP3: 성능 최적화 옵션 작동

---

**문서 작성일**: 2025-11-18
**최종 검토**: 예정
**상태**: Draft (review pending)
