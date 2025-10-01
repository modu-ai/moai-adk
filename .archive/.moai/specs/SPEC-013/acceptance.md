# SPEC-013 수락 기준: Python → TypeScript 완전 포팅

> **@TEST:MIGRATION-VERIFICATION-013** Python → TypeScript 완전 전환 검증 및 품질 보장

---

## 개요

### 수락 기준 범위
SPEC-013은 MoAI-ADK의 Python → TypeScript 완전 포팅을 통해 단일 런타임 환경을 달성하는 것입니다. 이 수락 기준은 기능 완성도, 성능 목표, 품질 보장, 사용자 경험 관점에서 완전한 전환을 검증합니다.

### 핵심 검증 영역
- **기능 동등성**: Python 버전과 100% 동일한 기능 제공
- **성능 향상**: 실행 속도 80% 향상, 메모리 50% 절약
- **타입 안전성**: TypeScript strict 모드 100% 적용
- **플랫폼 호환**: Windows/macOS/Linux 완전 지원
- **마이그레이션**: 기존 사용자 seamless 전환

---

## Given-When-Then 테스트 시나리오

### 시나리오 1: CLI 명령어 기능 동등성 검증

#### TC-013-001: moai init 명령어 동작 검증
**Given**: 
- 새로운 디렉토리 `/tmp/test-project`
- TypeScript 버전 MoAI-ADK 설치 완료
- Node.js 18+ 환경

**When**: 
```bash
cd /tmp && moai init test-project --mode=personal
```

**Then**:
- [ ] `.moai/` 디렉토리가 생성됨
- [ ] `.claude/` 디렉토리가 생성됨
- [ ] `config.json` 파일이 Personal 모드로 설정됨
- [ ] 실행 시간 < 2초
- [ ] 메모리 사용량 < 50MB
- [ ] 성공 메시지 "Project 'test-project' initialized successfully"

#### TC-013-002: moai doctor 명령어 시스템 검증
**Given**:
- TypeScript MoAI-ADK 설치 완료
- 시스템에 Node.js, Git, SQLite3 설치

**When**:
```bash
moai doctor
```

**Then**:
- [ ] Node.js 버전 검증 완료 (>= 18.0.0)
- [ ] Git 버전 검증 완료 (>= 2.0.0)
- [ ] SQLite3 검증 완료 (>= 3.0.0)
- [ ] 전체 검증 시간 < 1초
- [ ] 모든 요구사항 통과 메시지 출력

#### TC-013-003: moai status 명령어 TAG 스캔
**Given**:
- MoAI 프로젝트 초기화 완룼
- 소스 코드에 @TAG 스키마 포함

**When**:
```bash
moai status
```

**Then**:
- [ ] TAG 스캔 완료 및 결과 표시
- [ ] 진행률 추적 및 품질 지표 계산
- [ ] sync-report.md 자동 생성
- [ ] 스캔 시간 < 2초 (1000개 파일 기준)

### 시나리오 2: 성능 벤치마크 검증

#### TC-013-004: 대용량 파일 스캔 성능
**Given**:
- 5000개 파일을 포함한 대용량 프로젝트
- Python 버전 MoAI-ADK 성능 기준치 측정 완료

**When**:
```bash
# Python 버전 성능 측정
time python -m moai_adk.cli.commands status
# TypeScript 버전 성능 측정
time moai status
```

**Then**:
- [ ] TypeScript 버전 실행 시간 ≤ Python 버전 의 80%
- [ ] 메모리 사용량 ≤ Python 버전의 50%
- [ ] CPU 사용률 안정성 (피크 차이 < 20%)
- [ ] 결과 정확성 100% 동일

#### TC-013-005: 비동기 I/O 성능 검증
**Given**:
- 다수의 파일 작업이 필요한 상황
- 동시에 여러 파일 읽기/쓰기 작업

**When**:
```bash
moai install --template=large-project --files=1000
```

**Then**:
- [ ] 동시 작업 수 >= 10개 파일
- [ ] 전체 설치 시간 < 30초
- [ ] 메모리 사용량 안정성 (< 100MB)
- [ ] 모든 파일 무결성 검증 통과

### 시나리오 3: Claude Code 훅 시스템 검증

#### TC-013-006: pre-write-guard 훅 동작 검증
**Given**:
- Claude Code 환경에서 TypeScript 훅 설치
- 파일 쓰기 작업 시도

**When**:
- 민감한 파일에 대한 쓰기 시도 (API 키, 비밀번호 등)

**Then**:
- [ ] 민감 정보 탐지 및 차단
- [ ] 사용자에게 경고 메시지 표시
- [ ] 훅 실행 시간 < 100ms
- [ ] Python 훅과 100% 동일한 검증 결과

#### TC-013-007: policy-block 훅 명령어 차단
**Given**:
- 위험한 명령어 입력 상황
- TypeScript policy-block 훅 활성화

**When**:
- 위험 명령어 입력: `rm -rf /`, `sudo chmod 777`

**Then**:
- [ ] 명령어 실행 사전 차단
- [ ] 정책 위반 사유 로깅
- [ ] 대체 안전 명령어 제안
- [ ] 차단 시간 < 50ms

#### TC-013-008: tag-system 데이터베이스 전환
**Given**:
- 기존 Python SQLite TAG 데이터베이스
- better-sqlite3 기반 TypeScript 시스템

**When**:
```bash
# 기존 데이터베이스 마이그레이션
moai migrate --from=python --to=typescript
```

**Then**:
- [ ] 기존 TAG 데이터 100% 보존
- [ ] 스키마 호환성 유지
- [ ] 쿼리 성능 향상 (>= 20%)
- [ ] 데이터 무결성 검증 통과

### 시나리오 4: 설치 및 마이그레이션 검증

#### TC-013-009: npm 전역 설치
**Given**:
- 깨끗한 시스템 (Node.js 18+ 만 설치)
- Python MoAI-ADK 미설치 상태

**When**:
```bash
npm install -g moai-adk
```

**Then**:
- [ ] 설치 시간 < 30초
- [ ] 전역 명령어 `moai` 사용 가능
- [ ] 의존성 충돌 없음
- [ ] 모든 플랫폼에서 동일 결과

#### TC-013-010: Python → TypeScript 마이그레이션
**Given**:
- 기존 Python MoAI-ADK 설치 환경
- Python 버전으로 생성된 .moai/ 프로젝트

**When**:
```bash
# Python 버전 제거
pip uninstall moai-adk
# TypeScript 버전 설치
npm install -g moai-adk
# 마이그레이션 실행
moai migrate
```

**Then**:
- [ ] 기존 설정 100% 보존
- [ ] .moai/ 구조 완전 호환
- [ ] .claude/ 구조 완전 호환
- [ ] TAG 데이터베이스 완전 마이그레이션
- [ ] 마이그레이션 시간 < 10초

### 시나리오 5: 크로스 플랫폼 호환성

#### TC-013-011: Windows 환경 동작 검증
**Given**:
- Windows 10/11 환경
- PowerShell 5.0+ 설치

**When**:
```powershell
npm install -g moai-adk
moai init test-project
```

**Then**:
- [ ] 정상 설치 및 실행
- [ ] 윈도우 파일 시스템 호환
- [ ] PowerShell 환경 완전 지원
- [ ] 한글 경로 지원

#### TC-013-012: macOS 환경 동작 검증
**Given**:
- macOS 12+ 환경
- Zsh/Bash 쉴 환경

**When**:
```bash
npm install -g moai-adk
moai init test-project
```

**Then**:
- [ ] 정상 설치 및 실행
- [ ] Apple Silicon/Intel 모두 지원
- [ ] macOS 보안 정책 호환
- [ ] 터미널 출력 최적화

#### TC-013-013: Linux 배포판 호환성
**Given**:
- Ubuntu 20.04+, CentOS 8+, Debian 11+
- Bash 쉴 환경

**When**:
```bash
npm install -g moai-adk
moai init test-project
```

**Then**:
- [ ] 주요 Linux 배포판 지원
- [ ] 파일 권한 자동 설정
- [ ] systemd 서비스 호환
- [ ] 컴파일된 바이너리 지원

---

## 품질 게이트 및 검증 기준

### 자동화된 품질 검사

#### QG-013-001: 코드 품질 검사
**기준**:
- [ ] ESLint 규칙 100% 통과 (0개 오류)
- [ ] Prettier 코드 포맷 100% 적용
- [ ] TypeScript strict 모드 100% 적용
- [ ] 컴플렉시티 점수 ≤ 10 (함수별)
- [ ] 함수 크기 ≤ 50 LOC

**검증 명령어**:
```bash
npm run lint
npm run type-check
npm run format:check
npm run complexity-check
```

#### QG-013-002: 테스트 커버리지
**기준**:
- [ ] 단위 테스트 커버리지 >= 85%
- [ ] 브랜치 커버리지 >= 80%
- [ ] 함수 커버리지 >= 90%
- [ ] 모든 중요 경로 커버리지 100%

**검증 명령어**:
```bash
npm run test:coverage
npm run test:critical-path
```

#### QG-013-003: 빌드 및 배포 검사
**기준**:
- [ ] 빌드 시간 < 5초
- [ ] 번들 크기 < 50MB
- [ ] Tree shaking 효과 >= 30%
- [ ] 런타임 에러 0개

**검증 명령어**:
```bash
npm run build
npm run bundle-analyzer
npm run runtime-test
```

### 성능 벤치마크

#### PB-013-001: 실행 성능 비교
**비교 대상**: Python MoAI-ADK v0.1.28
**측정 환경**: 동일 하드웨어, 동일 데이터셋

**측정 지표**:
- [ ] CLI 시작 시간: Python 대비 >= 70% 향상
- [ ] TAG 스캔: Python 대비 >= 80% 향상
- [ ] 파일 I/O: Python 대비 >= 60% 향상
- [ ] 메모리 사용: Python 대비 >= 50% 절약

**벤치마크 명령어**:
```bash
# TypeScript 버전 측정
npm run benchmark:cli
npm run benchmark:scan
npm run benchmark:io
npm run benchmark:memory
```

#### PB-013-002: 동시성 성능 검증
**시나리오**: 다중 사용자 동시 접근
**부하**: 10개 동시 세션

**성능 기준**:
- [ ] 응답 시간 증가율 < 20%
- [ ] 에러 률 < 1%
- [ ] 메모리 누수 없음
- [ ] CPU 사용률 < 80%

### 보안 검사

#### SC-013-001: 정적 보안 검사
**도구**: ESLint security plugin, CodeQL
**대상**: 전체 TypeScript 코드베이스

**검사 항목**:
- [ ] 입력 검증 빠짐 0개
- [ ] SQL 인젝션 취약점 0개
- [ ] XSS 취약점 0개
- [ ] 경로 순환 취약점 0개
- [ ] 민감 정보 노출 0개

#### SC-013-002: 동적 보안 검사
**도구**: OWASP ZAP, custom security tests
**대상**: 실행 중인 MoAI-ADK 인스턴스

**검사 항목**:
- [ ] 무단 파일 접근 차단
- [ ] 권한 상승 방지
- [ ] 네트워크 더스 보호
- [ ] 암호화 데이터 보호

---

## 사용자 수락 테스트

### 초보 사용자 시나리오

#### UAT-013-001: 처음 설치 및 사용
**사용자 프로필**: 조직에서 정해진 개발 도구를 처음 사용하는 신입 개발자
**전제 조건**: Node.js 기본 지식, npm 사용 경험

**사용자 여정**:
1. **설치**: `npm install -g moai-adk`
   - 예상 결과: 30초 이내 설치 완료
   - 성공 기준: 오류 없이 설치 완룼

2. **환경 검사**: `moai doctor`
   - 예상 결과: 시스템 요구사항 자동 검증
   - 성공 기준: 요구사항 누락 시 설치 가이드 제공

3. **프로젝트 초기화**: `moai init my-first-project`
   - 예상 결과: 대화형 설정 가이드 실행
   - 성공 기준: .moai/, .claude/ 구조 완전 생성

**수락 기준**:
- [ ] 전체 과정 30분 이내 완료
- [ ] 에러 또는 경고 없이 진행
- [ ] 모든 단계에서 명확한 가이드 메시지 제공
- [ ] 사용자 만족도 >= 4.5/5.0

### 기존 사용자 마이그레이션

#### UAT-013-002: Python → TypeScript 전환
**사용자 프로필**: 6개월 이상 Python MoAI-ADK 사용 경험
**전제 조건**: 기존 Python 프로젝트 3개 이상 운영 중

**마이그레이션 여정**:
1. **백업**: 기존 프로젝트 백업
   - 대상: .moai/, .claude/, TAG 데이터베이스
   - 성공 기준: 100% 데이터 보존

2. **전환**: Python 제거 → TypeScript 설치
   - 명령어: `pip uninstall moai-adk && npm install -g moai-adk`
   - 성공 기준: 단순 명령어로 전환 완료

3. **검증**: 기존 프로젝트에서 동일 기능 확인
   - 테스트: `moai status`, TAG 스캔, Claude Code 훅
   - 성공 기준: 100% 동일한 결과

**수락 기준**:
- [ ] 전환 시간 < 15분
- [ ] 데이터 손실 0%
- [ ] 기능 호환성 100%
- [ ] 사용자 만족도 >= 4.0/5.0

### 고급 사용자 시나리오

#### UAT-013-003: 대규모 프로젝트 운영
**사용자 프로필**: 10,000개 이상 파일, 50명 이상 개발팀
**전제 조건**: 복잡한 Git 워크플로우, 다중 CI/CD 파이프라인

**성능 테스트**:
1. **대용량 스캔**: 10,000개 파일 TAG 스캔
   - 성능 목표: < 10초
   - 메모리 목표: < 200MB

2. **동시 작업**: 20명 동시 Claude Code 사용
   - 성능 목표: 지연 시간 < 500ms
   - 안정성: 쳚돌 및 데이터 손상 0건

**수락 기준**:
- [ ] 대용량 프로젝트 완전 지원
- [ ] 동시 사용자 안정성 100%
- [ ] 엔터프라이즈 요구사항 충족
- [ ] 성능 목표 100% 달성

---

## 최종 검증 체크리스트

### 전 영역 검증

#### 기능 완성도 (100% 필수)
- [ ] **Python 코드 완전 제거**: `find . -name "*.py" -type f | wc -l` = 0
- [ ] **TypeScript 구현 완성**: 모든 모듈 TypeScript 구현
- [ ] **CLI 명령어 동등성**: init, doctor, status, update, restore 완전 동작
- [ ] **Claude Code 훅 완전 전환**: 7개 훅 TypeScript 대체
- [ ] **설치 방식 전환**: npm 단독 설치 가능

#### 성능 목표 (100% 달성 필수)
- [ ] **실행 속도**: Python 대비 >= 80% 향상
- [ ] **메모리 사용**: Python 대비 >= 50% 절약
- [ ] **설치 시간**: npm install -g < 30초
- [ ] **빌드 시간**: tsup 빌드 < 5초
- [ ] **TAG 스캔**: 1000개 파일 < 2초

#### 품질 보장 (100% 통과 필수)
- [ ] **TypeScript strict**: 0개 타입 오류
- [ ] **테스트 커버리지**: >= 85%
- [ ] **ESLint 통과**: 0개 린트 오류
- [ ] **보안 검사**: 0개 취약점
- [ ] **크로스 플랫폼**: Windows/macOS/Linux 100% 지원

#### 사용자 수락 (>= 95% 만족)
- [ ] **신입 사용자**: 30분 이내 프로젝트 생성 완료
- [ ] **기존 사용자**: 15분 이내 마이그레이션 완료
- [ ] **고급 사용자**: 대규모 프로젝트 완전 지원
- [ ] **전체 만족도**: >= 4.0/5.0

### 마이그레이션 안전성
- [ ] **데이터 보존**: 기존 프로젝트 100% 호환
- [ ] **설정 호환**: .moai/, .claude/ 구조 완전 유지
- [ ] **TAG 데이터**: SQLite → better-sqlite3 완전 마이그레이션
- [ ] **롤백 계획**: 전환 실패 시 원래 상태 복구 가능

### 생태계 통합
- [ ] **npm 배포**: 정식 팩지 배포 완료
- [ ] **GitHub Actions**: CI/CD 파이프라인 100% 자동화
- [ ] **문서화**: 마이그레이션 가이드, API 문서 완성
- [ ] **커뮤니티**: 오픈소스 기여 준비 완료

---

## 완료 조건 (Definition of Done)

**SPEC-013이 완료되려면 위의 모든 체크리스트 항목이 100% 충족되어야 합니다.**

### 최종 인수 기준
1. **기능적 완성**: Python 의존성 0%, TypeScript 100% 구현
2. **성능 목표**: 모든 성능 지표 100% 달성
3. **품질 보장**: 모든 품질 게이트 100% 통과
4. **사용자 만족**: 모든 사용자 시나리오 >= 95% 만족
5. **운영 준비**: 마이그래이션 가이드 및 지원 체계 완성

**최종 검증**: TypeScript 버전만으로 모든 MoAI-ADK 기능을 Python 버전과 동일하게 사용할 수 있어야 하며, Python 환경 없이도 완전히 동작해야 합니다.