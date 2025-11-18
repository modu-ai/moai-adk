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

# SPEC-FIX-001: 수용 기준 및 검증 시나리오

## 1. 주요 성공 시나리오 (Happy Path)

### Main Success Scenario: 정상 환경에서의 statusline 복구

```
Given (전제조건):
  - MoAI-ADK 프로젝트 디렉토리 활성화
  - Python 3.13.9+ 설치됨
  - moai-adk 패키지 설치됨
  - .moai/config/config.json 파일 존재
  - + 활성화

When (실행):
  - Claude Code가 새로운 세션을 시작함
  - SessionStart 훅이 자동 실행됨
  - uvx moai-adk statusline 명령어 호출

Then (결과):
  - 명령이 성공적으로 완료됨 (exit code: 0)
  - 출력: "🤖 Haiku 4.5 | 🗿 Ver 0.26.0 | 📊 +0 M26 ?9 | 🔀 release/0.26.0"
  - "Ver unknown" 절대 표시 안 됨
  - Claude Code 하단 statusline에 정확한 버전 정보 표시됨
  - 응답 시간: 2초 이내

Acceptance Criteria:
  [ ] uvx 명령어 정상 실행
  [ ] config.json에서 버전 "0.26.0" 읽음
  [ ] Git 상태 정보 정확 (변경사항, 브랜치)
  [ ] 성능 요구사항 만족 (2초 이내)
```

---

## 2. 대안 시나리오 (Alternative Paths)

### Alternative Scenario 1: uvx 캐시 오염 상황 복구

```
Given (전제조건):
  - uvx 캐시에 이전 버전(0.25.x) 저장된 상태
  - 새 버전(0.26.0) PyPI에 배포됨
  - 사용자가 uvx 캐시 클리어 미실시

When (실행):
  - 캐시 클리어 없이 uvx moai-adk statusline 실행
  - 또는: 사용자가 문제 인식 후 캐시 클리어 지시 받음

Then (기대 결과 - 캐시 클리어 전):
  - 오염된 캐시의 이전 버전 실행됨
  - 출력: "Ver unknown" 또는 잘못된 버전 표시

Then (기대 결과 - 캐시 클리어 후):
  - 캐시 클리어 명령 실행: uvx cache clean moai-adk
  - 다시 실행: uvx moai-adk statusline
  - 출력: 최신 버전 정보 정확하게 표시됨
  - "Ver 0.26.0" 정상 표시

Recovery Steps:
  1. uvx cache clean moai-adk
  2. uv sync --force
  3. uvx moai-adk statusline 다시 실행
  4. 버전 정보 확인

Acceptance Criteria:
  [ ] 캐시 클리어 명령 성공
  [ ] 강제 재설치 완료
  [ ] 최신 버전 정보 표시
  [ ] 성능 저하 없음
```

### Alternative Scenario 2: config.json 파일 누락

```
Given (전제조건):
  - .moai/config/config.json 파일이 삭제된 상태
  - uvx moai-adk statusline 명령어 호출

When (실행):
  - statusline 명령어 config.json 로드 시도

Then (기대 결과 - Graceful Fallback):
  - 파일 미발견 에러 발생하지 않음 (중요!)
  - Fallback 값 사용: "Ver dev" 또는 "Ver N/A"
  - 명령 성공 반환 (exit code: 0)
  - 개발자 로그에만 경고 기록
  - Claude Code 중단되지 않음

Alternative Recovery:
  - 사용자에게 안내: "config.json 재생성 필요"
  - 자동 수복: .moai/config/config.json 템플릿 복원

Acceptance Criteria:
  [ ] 에러 화면 표시 안 됨
  [ ] Fallback 값 명확
  [ ] 로그에 오류 기록
  [ ] Claude Code 정상 작동
```

### Alternative Scenario 3: 패키지 import 실패

```
Given (전제조건):
  - moai_adk 패키지 손상 또는 불완전 설치
  - import 실패 발생 가능성

When (실행):
  - uvx moai-adk statusline 실행
  - 내부 import 오류 발생

Then (기대 결과 - 타임아웃 보호):
  - 재시도 횟수 제한: 최대 3회
  - 타임아웃 적용: 3초
  - 초과 시 Fallback 표시
  - 명령 실패하되 Claude Code는 정상

Recovery Steps:
  1. uv sync --force (강제 재설치)
  2. 패키지 무결성 검사
  3. 필요 시 Python 환경 재구성

Acceptance Criteria:
  [ ] 무한 루프 없음
  [ ] 재시도 제한 적용
  [ ] 타임아웃 보호 작동
  [ ] Fallback 표시
```

---

## 3. 에러 처리 시나리오 (Error Paths)

### Error Scenario 1: Python 버전 불일치

```
Given (전제조건):
  - Python 3.11 이전 버전 설치
  - moai-adk 0.26.0+ (3.13.9+ 요구)

When (실행):
  - uvx moai-adk statusline 호출

Then (예상 결과):
  - 명확한 에러 메시지:
    "Error: Python 3.13.9+ required. Current: 3.11.x"
  - Claude Code statusline: "Ver error"
  - 사용자 가이드 제공

Acceptance Criteria:
  [ ] 명확한 에러 메시지
  [ ] 대안 제시 (업그레이드 방법)
  [ ] 로그 기록
```

### Error Scenario 2: 네트워크 연결 실패

```
Given (전제조건):
  - 오프라인 환경 또는 PyPI 접근 불가
  - 로컬 캐시 비어있음

When (실행):
  - uvx moai-adk statusline 호출
  - PyPI 다운로드 불가

Then (예상 결과):
  - 캐시된 버전 사용 (있다면)
  - 캐시 없으면 Fallback: "Ver offline"
  - 재연결 시 자동 업데이트

Acceptance Criteria:
  [ ] 오프라인 안내
  [ ] 캐시 활용
  [ ] Fallback 안내
  [ ] 재연결 감지
```

### Error Scenario 3: 권한 부족

```
Given (전제조건):
  - .moai/config/config.json 읽기 권한 없음 (e.g., 444)
  - 사용자가 읽기 권한 부족

When (실행):
  - uvx moai-adk statusline 실행

Then (예상 결과):
  - 권한 에러 감지
  - 대안: 환경 변수로 버전 제공 (MOAI_VERSION)
  - 또는 Fallback: "Ver restricted"

Acceptance Criteria:
  [ ] 권한 에러 처리
  [ ] 대안 메커니즘
  [ ] 명확한 안내
```

---

## 4. 성능 검증 (Performance Validation)

### Test 4-1: 응답 시간

```gherkin
Given:
  - statusline.py 마이그레이션 완료
  - uvx 캐시 준비됨

When:
  - 첫 번째 실행 (캐시 미스):
    $ time uvx moai-adk statusline
  - 두 번째 실행 (캐시 히트):
    $ time uvx moai-adk statusline

Then:
  - 첫 실행: 2초 이내
  - 두 번째: 1초 이내
  - 메모리 사용: < 50MB

Metrics:
  [ ] real: < 2.0s (캐시 미스)
  [ ] real: < 1.0s (캐시 히트)
  [ ] user: < 1.5s
  [ ] sys: < 0.5s
```

### Test 4-2: 메모리 사용

```bash
# 메모리 프로파일링
$ /usr/bin/time -v uvx moai-adk statusline 2>&1 | grep -E "Maximum|Elapsed"

# 기대:
# Maximum resident set size: < 50000 KB
# Elapsed (wall clock) time: < 2 seconds
```

### Test 4-3: CPU 효율성

```bash
# CPU 사용률 확인
$ python -m cProfile -s cumulative -c "uvx moai-adk statusline" 2>&1

# 기대:
# 상위 함수: config_loader, version_formatter
# CPU time < 0.5초
```

---

## 5. 기능 검증 Checklist

### 5-1: 정상 출력 검증

```
출력 형식: 🤖 {MODEL} | 🗿 Ver {VERSION} | 📊 {GIT_STATUS} | 🔀 {BRANCH}

예시:
✅ 🤖 Haiku 4.5 | 🗿 Ver 0.26.0 | 📊 +0 M26 ?9 | 🔀 release/0.26.0
❌ 🤖 Haiku 4.5 | 🗿 Ver unknown | 📊 +0 M26 ?9 | 🔀 release/0.26.0
❌ 🤖 Haiku 4.5 | 🗿 error | 📊 +0 M26 ?9 | 🔀 release/0.26.0
```

**검증 항목**:

- [ ] 정확한 emojis (🤖, 🗿, 📊, 🔀)
- [ ] 올바른 모델명 (Haiku, Sonnet, Opus)
- [ ] 정확한 버전 (0.26.0)
- [ ] Git 상태 동기화 (+0 변경, M26 수정, ?9 미추적)
- [ ] 정확한 브랜치명 (release/0.26.0)

### 5-2: Git 상태 동기화

```
Given:
  - 파일 26개 수정됨
  - 파일 9개 미추적
  - 스테이징 영역 비어있음

When:
  - uvx moai-adk statusline 실행

Then:
  - 📊 +0 M26 ?9 정확하게 표시됨
  - real-time 업데이트 (git status 동기화)

[ ] Git 상태 변경 감지
[ ] 즉각적인 업데이트 (< 100ms)
[ ] 멀티 세션 일관성
```

### 5-3: 브랜치 정보

```
Given:
  - 현재 브랜치: release/0.26.0

When:
  - uvx moai-adk statusline 실행
  - 브랜치 변경 후 다시 실행

Then:
  - 🔀 release/0.26.0 표시
  - 브랜치 변경 시 업데이트됨

[ ] 현재 브랜치 정확 표시
[ ] 브랜치 변경 감지
[ ] 명확한 표시 형식
```

---

## 6. 멀티 세션 검증

### Test 6-1: 동시 세션

```
Scenario:
  1. Claude Code 세션 A 시작
     → statusline 표시: Ver 0.26.0
  2. Claude Code 세션 B 시작
     → statusline 표시: Ver 0.26.0
  3. 세션 A 파일 수정
     → 세션 A Git 상태: M1 (1 파일 수정)
     → 세션 B Git 상태: M1 (동기화)
  4. 세션 B 파일 추가
     → 세션 B Git 상태: M1 ?1
     → 세션 A Git 상태: M1 ?1 (동기화)

Expected:
  [ ] 각 세션 독립적 statusline 표시
  [ ] 캐시 충돌 없음
  [ ] Git 상태 일관성 유지
  [ ] 성능 저하 없음 (< 2초)
```

### Test 6-2: 빠른 연속 실행

```bash
# 10번 연속 실행
for i in {1..10}; do
  echo "=== Run $i ==="
  time uvx moai-adk statusline
  sleep 0.1
done

Expected:
  - 모든 실행 성공 (exit code: 0)
  - 버전 정보 일관성
  - 성능 저하 없음
  - 캐시 메커니즘 효과 (6-10번은 < 1초)
```

---

## 7. Definition of Done (DoD)

### 코드 완성

- [ ] SPEC-FIX-001 모든 요구사항 구현
- [ ] src/moai_adk/statusline/main.py 업데이트
- [ ] 타입 힌팅 완전 (mypy: pass)
- [ ] 린팅 통과 (ruff: pass)
- [ ] 테스트 커버리지 ≥ 90%

### 테스트 완성

- [ ] 단위 테스트 작성 및 통과
  - [ ] test_config_loading.py
  - [ ] test_version_formatting.py
  - [ ] test_error_handling.py
- [ ] 통합 테스트 작성 및 통과
  - [ ] test_statusline_recovery.py
  - [ ] test_cache_mechanism.py
  - [ ] test_multi_session.py
- [ ] 성능 테스트 통과
  - [ ] 초기 실행: 2초 이내
  - [ ] 캐시 히트: 1초 이내
  - [ ] 메모리: < 50MB

### 문서 완성

- [ ] 개발자 가이드 업데이트
  - [ ] docs/statusline-guide.md 작성
  - [ ] 설정 방법 설명
  - [ ] 캐시 관리 가이드
- [ ] 트러블슈팅 가이드
  - [ ] docs/troubleshooting-statusline.md 작성
  - [ ] 일반적인 문제 및 해결책
  - [ ] 진단 도구 사용법
- [ ] CHANGELOG 업데이트
  - [ ] SPEC-FIX-001 적용 내용 기록
  - [ ] 버전 0.26.1 또는 0.26.0 패치 기록

### 품질 게이트 (TRUST 5)

- [ ] **Test**: 커버리지 ≥ 90%, 모든 테스트 통과
- [ ] **Readable**: mypy, ruff, pylint 모두 통과
- [ ] **Unified**: 일관된 에러 처리 및 로깅
- [ ] **Secured**: 보안 취약점 검사 완료
- [ ] **Trackable**: SPEC → 코드 → 테스트 추적 가능

### 배포 완료

- [ ] Git 커밋 및 PR 병합
- [ ] PyPI 배포 (새 버전 릴리스)
- [ ] 릴리스 노트 작성
- [ ] 사용자 공지 (필요시)

---

## 8. 롤백 계획

### 긴급 롤백 절차

```bash
# Scenario: 새 버전에서 버그 발생
# Action: 이전 버전으로 즉시 복구

# Step 1: 캐시 클리어
uvx cache clean moai-adk

# Step 2: 이전 버전 설치 명시
uvx moai-adk==0.26.0 statusline

# Step 3: 검증
uvx moai-adk statusline

# Step 4: 보고 (bug report)
```

---

**검증 계획 작성일**: 2025-11-18
**최종 검토**: 예정
**상태**: Draft
