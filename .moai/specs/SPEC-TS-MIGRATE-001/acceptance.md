# SPEC-TS-MIGRATE-001: 인수 기준

---
spec_id: SPEC-TS-MIGRATE-001
version: "1.0.0"
created: "2025-12-05"
updated: "2025-12-05"
author: "MoAI-ADK Team"
---

## 1. 개요

본 문서는 MoAI-ADK TypeScript/Bun 마이그레이션의 인수 기준을 정의합니다. 모든 기준을 충족해야 마이그레이션이 완료된 것으로 간주됩니다.

---

## 2. Given-When-Then 시나리오

### 2.1 프로젝트 초기화 시나리오

**Scenario 1: 새 프로젝트 초기화**

```gherkin
Given 사용자가 빈 디렉토리에 있고
  And TypeScript 버전의 moai-adk가 설치되어 있을 때
When 사용자가 "moai-adk init" 명령을 실행하면
Then 시스템은 100ms 이내에 응답을 시작해야 하고
  And .moai/ 디렉토리 구조가 생성되어야 하고
  And .claude/ 디렉토리 구조가 생성되어야 하고
  And CLAUDE.md 파일이 생성되어야 하고
  And config.yaml 파일이 올바른 기본값으로 생성되어야 한다
```

**Scenario 2: 언어 설정과 함께 초기화**

```gherkin
Given 사용자가 빈 디렉토리에 있고
  And TypeScript 버전의 moai-adk가 설치되어 있을 때
When 사용자가 "moai-adk init --language ko" 명령을 실행하면
Then config.yaml의 conversation_language가 "ko"로 설정되어야 하고
  And conversation_language_name이 "Korean"으로 설정되어야 하고
  And 초기화 성공 메시지가 한국어로 표시되어야 한다
```

**Scenario 3: 기존 프로젝트 업데이트**

```gherkin
Given 사용자가 Python 버전으로 초기화된 MoAI 프로젝트에 있고
  And TypeScript 버전의 moai-adk가 설치되어 있을 때
When 사용자가 "moai-adk update" 명령을 실행하면
Then 기존 설정이 보존되어야 하고
  And 새로운 기능이 추가되어야 하고
  And 마이그레이션 보고서가 생성되어야 한다
```

---

### 2.2 CLI 호환성 시나리오

**Scenario 4: 모든 기존 CLI 명령어 동작**

```gherkin
Given TypeScript 버전의 moai-adk가 설치되어 있을 때
When 사용자가 Python 버전에서 지원하던 모든 CLI 명령어를 실행하면
Then 모든 명령어가 동일한 옵션과 동작으로 작동해야 하고
  And 출력 형식이 동일해야 하고
  And 종료 코드가 동일해야 한다
```

**지원 명령어 목록:**
- `moai-adk init [options]`
- `moai-adk update [options]`
- `moai-adk status`
- `moai-adk worktree [subcommand]`
- `moai-adk config [subcommand]`
- `moai-adk --version`
- `moai-adk --help`

**Scenario 5: 대화형 프롬프트 동작**

```gherkin
Given 사용자가 터미널에서 moai-adk를 실행하고 있을 때
When 대화형 입력이 필요한 명령어를 실행하면
Then 프롬프트가 올바르게 표시되어야 하고
  And 키보드 입력이 정상적으로 처리되어야 하고
  And 선택 옵션이 화살표 키로 탐색 가능해야 한다
```

---

### 2.3 타입 안전성 시나리오

**Scenario 6: 설정 파일 검증**

```gherkin
Given 사용자가 config.yaml 파일을 수정했을 때
When 시스템이 설정 파일을 로드하면
Then 스키마에 맞지 않는 값은 거부되어야 하고
  And 명확한 오류 메시지가 표시되어야 하고
  And 오류 위치가 정확히 표시되어야 한다
```

**Scenario 7: 잘못된 타입 입력 처리**

```gherkin
Given API가 특정 타입의 입력을 기대하고 있을 때
When 잘못된 타입의 값이 전달되면
Then 컴파일 타임에 타입 오류가 발생해야 하고
  And 런타임에 유효성 검사 오류가 발생해야 하고
  And 오류 메시지에 기대 타입과 실제 타입이 표시되어야 한다
```

---

### 2.4 성능 시나리오

**Scenario 8: CLI 응답 시간**

```gherkin
Given moai-adk가 설치되어 있고
  And 시스템이 유휴 상태일 때
When 사용자가 "moai-adk --version" 명령을 실행하면
Then 응답이 100ms 이내에 완료되어야 한다
```

**Scenario 9: 대규모 프로젝트 처리**

```gherkin
Given 1000개 이상의 파일이 있는 프로젝트에서
  And moai-adk가 설치되어 있을 때
When 사용자가 "moai-adk status" 명령을 실행하면
Then 응답이 3초 이내에 완료되어야 하고
  And 메모리 사용량이 500MB를 초과하지 않아야 한다
```

---

### 2.5 에러 처리 시나리오

**Scenario 10: 네트워크 오류 복구**

```gherkin
Given 사용자가 업데이트를 실행 중이고
  And 네트워크 연결이 중단되었을 때
When 시스템이 네트워크 오류를 감지하면
Then 적절한 오류 메시지가 표시되어야 하고
  And 부분 다운로드 파일이 정리되어야 하고
  And 재시도 옵션이 제공되어야 한다
```

**Scenario 11: 파일 시스템 권한 오류**

```gherkin
Given 사용자가 쓰기 권한이 없는 디렉토리에서
  And "moai-adk init" 명령을 실행할 때
When 시스템이 권한 오류를 감지하면
Then 명확한 권한 오류 메시지가 표시되어야 하고
  And 필요한 권한에 대한 안내가 제공되어야 한다
```

---

### 2.6 마이그레이션 특화 시나리오

**Scenario 12: Python 버전과의 병행 운영**

```gherkin
Given Python 버전의 moai-adk가 설치되어 있고
  And TypeScript 버전도 설치되어 있을 때
When 두 버전이 동일한 프로젝트에서 실행되면
Then 설정 파일이 양쪽 버전과 호환되어야 하고
  And 데이터 손상이 발생하지 않아야 한다
```

**Scenario 13: 템플릿 변수 치환**

```gherkin
Given TypeScript 버전의 moai-adk가 설치되어 있고
  And 템플릿 파일에 {{변수}} 패턴이 있을 때
When 시스템이 템플릿을 처리하면
Then 모든 변수가 올바른 값으로 치환되어야 하고
  And 치환되지 않은 변수가 있으면 경고가 표시되어야 한다
```

---

## 3. 품질 메트릭

### 3.1 테스트 커버리지

| 메트릭 | 최소 기준 | 목표 |
|--------|----------|------|
| Statement Coverage | 90% | 95% |
| Branch Coverage | 90% | 95% |
| Function Coverage | 90% | 95% |
| Line Coverage | 90% | 95% |

### 3.2 코드 품질

| 메트릭 | 기준 |
|--------|------|
| TypeScript strict mode | 필수 |
| `any` 타입 사용 | 0건 (명시적 예외 제외) |
| 순환 의존성 | 0건 |
| 파일당 LOC | 500 이하 |
| 린팅 오류 | 0건 |
| 타입 오류 | 0건 |

### 3.3 성능 기준

| 작업 | 최대 시간 | 최대 메모리 |
|------|----------|-------------|
| CLI 시작 | 100ms | 50MB |
| 프로젝트 초기화 | 2s | 100MB |
| 설정 로드 | 50ms | 20MB |
| 대규모 프로젝트 스캔 | 3s | 500MB |

---

## 4. 기능 동등성 체크리스트

### 4.1 CLI 명령어

- [ ] `moai-adk init` - 프로젝트 초기화
- [ ] `moai-adk init --language <lang>` - 언어 설정 초기화
- [ ] `moai-adk update` - 버전 업데이트
- [ ] `moai-adk update --check` - 업데이트 확인
- [ ] `moai-adk status` - 상태 표시
- [ ] `moai-adk config get <key>` - 설정 조회
- [ ] `moai-adk config set <key> <value>` - 설정 변경
- [ ] `moai-adk worktree list` - Worktree 목록
- [ ] `moai-adk worktree create <name>` - Worktree 생성
- [ ] `moai-adk worktree delete <name>` - Worktree 삭제
- [ ] `moai-adk --version` - 버전 표시
- [ ] `moai-adk --help` - 도움말 표시

### 4.2 핵심 기능

- [ ] 설정 파일 로드/저장
- [ ] 템플릿 변수 치환
- [ ] Git 연동 (simple-git)
- [ ] 파일 시스템 작업
- [ ] 네트워크 요청
- [ ] 캐싱 시스템
- [ ] 로깅 시스템
- [ ] 에러 복구 시스템

### 4.3 고급 기능

- [ ] 훅 시스템
- [ ] 이벤트 시스템
- [ ] 모니터링 대시보드
- [ ] 엔터프라이즈 기능
- [ ] JIT 컴파일러

---

## 5. 마이그레이션 성공 기준

### 5.1 필수 기준 (Must Have)

1. **기능 완전성:** Python 버전의 모든 기능이 TypeScript 버전에서 동작
2. **테스트 통과:** 단위/통합/E2E 테스트 100% 통과
3. **타입 안전성:** strict mode 활성화, `any` 0건
4. **성능 동등성:** Python 버전 대비 성능 저하 없음
5. **문서화:** API 문서, 마이그레이션 가이드 완료

### 5.2 권장 기준 (Should Have)

1. **성능 개선:** CLI 응답 시간 20% 개선
2. **코드 품질:** 모든 파일 400 LOC 이하
3. **테스트 커버리지:** 95% 이상

### 5.3 선택 기준 (Nice to Have)

1. **WebAssembly 통합:** 성능 크리티컬 모듈
2. **플러그인 시스템:** 확장 가능한 아키텍처
3. **Deno 호환성:** 대체 런타임 지원

---

## 6. Definition of Done (DoD)

마이그레이션이 완료되었다고 판단하기 위한 기준:

### 6.1 코드 완료

- [ ] 모든 Python 모듈이 TypeScript로 전환됨
- [ ] 모든 God Object가 분리됨
- [ ] 순환 의존성이 제거됨
- [ ] TypeScript strict mode 활성화됨
- [ ] 린팅 오류 0건

### 6.2 테스트 완료

- [ ] 단위 테스트 커버리지 90% 이상
- [ ] 통합 테스트 100% 통과
- [ ] E2E 테스트 100% 통과
- [ ] 성능 벤치마크 통과

### 6.3 문서화 완료

- [ ] API 문서 생성
- [ ] 마이그레이션 가이드 작성
- [ ] CHANGELOG 업데이트
- [ ] README 업데이트

### 6.4 배포 준비

- [ ] npm 패키지 빌드 성공
- [ ] CI/CD 파이프라인 통과
- [ ] 릴리스 노트 작성
- [ ] Python 버전 deprecation 공지 준비

---

## 7. Phase별 완료 기준

### Phase 0: God Object 리팩토링

- [ ] 모든 파일 500 LOC 이하
- [ ] 기존 테스트 100% 통과
- [ ] 순환 의존성 0건

### Phase 1: 인프라 구축

- [ ] `bun build` 성공
- [ ] `bun test` 실행 가능
- [ ] `biome check` 통과
- [ ] `tsc --noEmit` 통과

### Phase 2: Foundation 마이그레이션

- [ ] 테스트 커버리지 90% 이상
- [ ] Python 버전과 기능 동등성 검증
- [ ] 타입 오류 0건

### Phase 3: Utils/Project 마이그레이션

- [ ] 테스트 커버리지 90% 이상
- [ ] 통합 캐시 시스템 동작
- [ ] 설정 로드/저장 동작

### Phase 4: Statusline/Core 마이그레이션

- [ ] 테스트 커버리지 90% 이상
- [ ] DI 컨테이너 동작
- [ ] 순환 의존성 0건

### Phase 5: CLI 마이그레이션

- [ ] 모든 CLI 명령어 동작
- [ ] 대화형 프롬프트 동작
- [ ] 응답 시간 100ms 이내

### Phase 6: 통합 완료

- [ ] E2E 테스트 100% 통과
- [ ] 기능 동등성 100%
- [ ] 문서화 완료
- [ ] 배포 준비 완료

---

## 8. 검증 도구

### 8.1 자동화 검증

```bash
# 테스트 실행
bun test --coverage

# 타입 체크
tsc --noEmit

# 린팅
biome check .

# 순환 의존성 체크
npx madge --circular src/

# 번들 크기 분석
npx bundlesize
```

### 8.2 성능 벤치마크

```typescript
// benchmarks/cli-startup.ts
import { bench, run } from 'mitata';

bench('CLI startup', async () => {
  await $`moai-adk --version`;
});

await run();
```

---

## 9. 추적성

### 9.1 요구사항 검증 매핑

| 요구사항 ID | 시나리오 ID | 검증 방법 |
|-------------|-------------|-----------|
| REQ-U001 | Scenario 6, 7 | 타입 체크, 테스트 |
| REQ-U002 | - | 커버리지 리포트 |
| REQ-U003 | Scenario 4, 5 | E2E 테스트 |
| REQ-U004 | - | 타입 정의 검증 |
| REQ-E001 | Scenario 8 | 벤치마크 |
| REQ-E002 | - | 코드 리뷰 |
| REQ-W001 | - | 린팅, 타입 체크 |
| REQ-W002 | - | madge 도구 |
| REQ-W003 | - | LOC 분석 |

### 9.2 관련 문서

- spec.md: SPEC-TS-MIGRATE-001 명세서
- plan.md: 구현 계획서

---

## 10. 롤백 기준

### 10.1 Phase별 롤백 트리거

마이그레이션 중 다음 상황 발생 시 해당 Phase를 롤백하고 이전 상태로 복구합니다.

**즉시 롤백 조건:**

1. **테스트 커버리지 80% 미만**: 목표 90%에서 10% 이상 하락 시
2. **프로덕션 크리티컬 버그**: 데이터 손실, 보안 취약점 발견 시
3. **성능 저하 50% 이상**: Python 버전 대비 응답 시간 2배 이상 증가 시
4. **순환 의존성 발생**: 새로운 순환 의존성 도입 시

**경고 후 롤백 조건:**

1. **`any` 타입 10건 이상**: 타입 안전성 목표 미달
2. **린팅 오류 50건 이상**: 코드 품질 기준 미달
3. **빌드 실패 3회 연속**: CI/CD 파이프라인 불안정

### 10.2 롤백 절차

```
1. 롤백 결정 → 팀 합의
2. 현재 상태 백업
3. Git revert 또는 branch 전환
4. Python 버전 활성화
5. 롤백 원인 분석 및 문서화
6. 수정 계획 수립 후 재시도
```

### 10.3 롤백 방지 전략

**사전 예방:**
- 각 Phase 완료 전 품질 게이트 검증
- 증분 배포로 위험 최소화
- Python 버전 병행 운영 유지

**모니터링:**
- 테스트 커버리지 일일 추적
- 성능 벤치마크 주간 실행
- 코드 품질 메트릭 자동 리포팅

---

*본 인수 기준은 SPEC-TS-MIGRATE-001의 완료 조건을 정의합니다.*
