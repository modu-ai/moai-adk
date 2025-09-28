# SPEC-013 구현 계획: Python → TypeScript 완전 포팅

> **@TASK:COMPLETE-MIGRATION-013** Week 3 목표: Python 종속성 완전 제거 및 TypeScript 단일 런타임 달성

---

## 프로젝트 개요

### 목표
MoAI-ADK의 Python → TypeScript 완전 포팅을 통해 단일 런타임 환경 구축 및 성능 최적화 달성

### 범위
- **포팅 대상**: 남은 Python 모듈 완전 TypeScript 전환
- **성능 목표**: 실행 속도 80% 향상, 메모리 50% 절약
- **배포 전환**: PyPI → npm 단독 배포 채널
- **Claude Code 통합**: 7개 Python 훅 → TypeScript 완전 대체

---

## 마일스톤 및 우선순위

### 1차 목표: 핵심 CLI 모듈 완전 포팅
**우선순위**: High
**의존성**: Week 1-2 TypeScript 기반 구축 완료

#### CLI 명령어 시스템 완전 전환
- **status.ts**: Python `moai status` 명령어 완전 대체
  - TAG 스캔 결과 표시
  - 프로젝트 상태 분석
  - 품질 지표 리포팅

- **update.ts**: Python `moai update` 명령어 완전 대체
  - 패키지 업데이트 로직
  - 설정 동기화
  - 버전 호환성 검증

- **restore.ts**: Python `moai restore` 명령어 완전 대체
  - 백업에서 프로젝트 복원
  - 설정 복구
  - 무결성 검증

#### CLI 지원 모듈 포팅
- **wizard.ts**: 대화형 설치 가이드 TypeScript 구현
  - inquirer.js 기반 인터랙티브 UI
  - 프로젝트 유형별 설정 가이드
  - 실시간 검증 및 피드백

- **banner.ts**: UI/UX 요소 TypeScript 구현
  - chalk 기반 터미널 색상
  - 진행률 표시
  - 브랜딩 요소

- **executor.ts**: 명령어 실행 엔진 TypeScript 구현
  - 통합 명령어 실행 로직
  - 에러 처리 및 로깅
  - 성능 모니터링

### 2차 목표: Install 시스템 완전 전환
**우선순위**: High
**의존성**: 1차 목표 완료 후

#### 설치 오케스트레이션 시스템
- **orchestrator.ts**: Python InstallationOrchestrator 완전 대체
  - 전체 설치 프로세스 관리
  - 단계별 진행률 추적
  - 실패 시 롤백 메커니즘

- **resource.ts**: Python ResourceManager 완전 대체
  - 템플릿 및 리소스 관리
  - 파일 복사 최적화
  - 권한 설정 자동화

- **template.ts**: Python TemplateManager 완전 대체
  - Jinja2 → Handlebars 전환
  - 동적 템플릿 렌더링
  - 조건부 콘텐츠 생성

#### 설정 및 검증 시스템
- **config.ts**: Python ConfigManager 완전 대체
  - Personal/Team 모드 전환
  - 설정 검증 및 마이그레이션
  - 환경별 설정 관리

- **validator.ts**: Python ResourceValidator 완전 대체
  - 경로 검증 및 보안 검사
  - 리소스 무결성 확인
  - 크로스 플랫폼 호환성 검증

### 3차 목표: Claude Code 훅 시스템 완전 전환
**우선순위**: High
**의존성**: 2차 목표 완료 후

#### 보안 및 정책 훅
- **pre-write-guard.ts**: `pre_write_guard.py` 완전 대체
  - 파일 쓰기 전 보안 검증
  - 민감 정보 탐지 및 차단
  - 정책 위반 사전 방지

- **policy-block.ts**: `policy_block.py` 완전 대체
  - 명령어 정책 검증
  - 위험 명령어 차단
  - 허용 목록 관리

- **steering-guard.ts**: `steering_guard.py` 완전 대체
  - AI 응답 품질 검증
  - TRUST 원칙 준수 확인
  - 개발 가이드 적용

#### 모니터링 및 실행 훅
- **session-start.ts**: `session_start.py` 완전 대체
  - 세션 초기화 및 환경 설정
  - 프로젝트 컨텍스트 로딩
  - 개발 가이드 메모리 로딩

- **file-monitor.ts**: `file_monitor.py` 완전 대체
  - 실시간 파일 변경 감지
  - 자동 TAG 업데이트
  - 품질 지표 실시간 계산

- **language-detector.ts**: `language_detector.py` 완전 대체
  - 프로그래밍 언어 자동 감지
  - 컨텍스트 기반 도구 설정
  - 언어별 최적화 적용

- **test-runner.ts**: `test_runner.py` 완전 대체
  - 자동 테스트 실행
  - 다중 테스트 프레임워크 지원
  - 커버리지 분석 및 리포팅

### 4차 목표: Core 시스템 모듈 완전 전환
**우선순위**: Medium
**의존성**: 3차 목표 완료 후

#### Git 관리 시스템
- **manager.ts**: Python GitManager 완전 대체
  - simple-git 라이브러리 기반 Git 작업
  - Personal/Team 전략 지원
  - 자동 브랜치 및 커밋 관리

- **operations.ts**: Git 작업 로직 TypeScript 구현
  - 브랜치 생성 및 관리
  - 커밋 메시지 자동 생성
  - 태그 및 릴리스 관리

#### TAG 시스템 완전 전환
- **database.ts**: Python sqlite3 → better-sqlite3 완전 전환
  - 동기 SQLite 바인딩 활용
  - 쿼리 성능 최적화
  - 트랜잭션 처리 개선

- **parser.ts**: Python TagParser 완전 대체
  - 16-Core TAG 체계 파싱
  - 정규식 최적화
  - 다중 파일 병렬 처리

- **reporter.ts**: Python SyncReporter 완전 대체
  - sync-report.md 자동 생성
  - 품질 지표 계산
  - 진행률 추적 및 시각화

#### 보안 및 유틸리티 시스템
- **security/validator.ts**: 보안 검증 TypeScript 구현
  - 입력 검증 및 정규화
  - 파일 권한 검사
  - 보안 정책 적용

- **utils/file-ops.ts**: 파일 작업 TypeScript 구현
  - 비동기 파일 I/O 최적화
  - 크로스 플랫폼 호환성
  - 에러 처리 및 복구

### 5차 목표: 배포 시스템 완전 전환
**우선순위**: Medium
**의존성**: 모든 포팅 완료 후

#### npm 단독 배포 채널
- **package.json 최적화**: 단일 패키지 설정
- **바이너리 배포**: pkg/nexe 활용 실행 파일 생성
- **GitHub Actions**: 자동 배포 파이프라인
- **버전 관리**: semantic versioning 적용

#### Python 의존성 완전 제거
- **PyPI 패키지 deprecated**: 폐기 예고 및 마이그레이션 가이드
- **문서 업데이트**: 설치 방법 변경 안내
- **사용자 지원**: 전환 과정 지원 시스템

---

## 기술적 접근 방법

### 아키텍처 설계

#### 모듈 간 의존성 관리
```typescript
// 핵심 의존성 구조
interface ModuleDependencies {
  cli: ['core/system-checker', 'utils/logger'];
  installer: ['core/git', 'core/security', 'utils/file-ops'];
  hooks: ['core/tag-system', 'core/security'];
  'tag-system': ['utils/config'];
}
```

#### 성능 최적화 전략
- **비동기 I/O**: Node.js의 이벤트 루프 활용
- **병렬 처리**: Promise.all을 통한 동시 작업
- **스트림 처리**: 대용량 파일 메모리 효율적 처리
- **캐싱**: 자주 사용되는 데이터 메모리 캐시

#### 타입 안전성 보장
```typescript
// 엄격한 타입 정의
interface InstallOptions {
  readonly projectName: string;
  readonly mode: 'personal' | 'team';
  readonly features: readonly Feature[];
  readonly dryRun?: boolean;
}

// 런타임 검증
import { z } from 'zod';
const InstallOptionsSchema = z.object({
  projectName: z.string().min(1),
  mode: z.enum(['personal', 'team']),
  features: z.array(z.enum(['git', 'claude', 'docs'])),
  dryRun: z.boolean().optional()
});
```

### 마이그레이션 전략

#### 점진적 포팅 방법
1. **인터페이스 우선**: Python API와 동일한 TypeScript 인터페이스 정의
2. **기능 단위 포팅**: 모듈별로 완전 포팅 후 테스트
3. **통합 검증**: 포팅된 모듈 간 통합 테스트
4. **성능 검증**: Python 버전과 성능 비교

#### 호환성 유지
```typescript
// 기존 설정 파일 호환성
interface LegacyConfig {
  version: string;
  mode: 'personal' | 'team';
  // Python 버전 설정 구조 유지
}

// 마이그레이션 유틸리티
class ConfigMigrator {
  migrateFromPython(pythonConfig: LegacyConfig): TypeScriptConfig {
    // Python 설정을 TypeScript 형식으로 변환
  }
}
```

#### 데이터 마이그레이션
- **SQLite 스키마**: 기존 데이터베이스 구조 유지
- **설정 파일**: .moai/config.json 형식 호환
- **TAG 데이터**: 기존 TAG 인덱스 완전 호환

### 품질 보장 전략

#### 테스트 커버리지
- **단위 테스트**: 각 모듈 85% 이상 커버리지
- **통합 테스트**: 모듈 간 상호작용 검증
- **E2E 테스트**: 전체 워크플로우 검증
- **성능 테스트**: Python 버전과 벤치마크 비교

#### 코드 품질
```typescript
// ESLint 규칙 강화
rules: {
  '@typescript-eslint/strict-boolean-expressions': 'error',
  '@typescript-eslint/prefer-readonly': 'error',
  'complexity': ['error', 10],
  'max-lines-per-function': ['error', 50]
}
```

#### 문서화
- **API 문서**: TypeDoc 자동 생성
- **마이그레이션 가이드**: Python → TypeScript 전환 가이드
- **아키텍처 문서**: 시스템 설계 및 의존성 문서

---

## 리스크 및 대응 방안

### 기술적 리스크

#### 성능 목표 미달성
**리스크**: TypeScript 포팅 후 Python 대비 성능 향상 목표 미달
**확률**: Medium
**영향도**: High
**대응 방안**:
- 단계별 성능 벤치마크 실시
- V8 프로파일링 도구 활용
- 병목 지점 식별 및 최적화
- 필요시 Rust/C++ 네이티브 모듈 활용

#### Claude Code 훅 호환성
**리스크**: TypeScript 훅이 Claude Code API와 완전 호환되지 않음
**확률**: Medium
**영향도**: High
**대응 방안**:
- Claude Code API 명세 철저 분석
- Python 훅 동작 완전 복제
- 단계적 테스트 및 검증
- Claude 팀과 호환성 논의

### 프로젝트 리스크

#### 사용자 마이그레이션 저항
**리스크**: 기존 Python 사용자의 TypeScript 전환 거부
**확률**: Medium
**영향도**: Medium
**대응 방안**:
- 명확한 마이그레이션 가이드 제공
- 단계적 전환 지원 도구
- 기존 설정 자동 마이그레이션
- 충분한 전환 기간 제공

#### 개발 일정 지연
**리스크**: 복잡한 포팅 작업으로 인한 일정 지연
**확률**: High
**영향도**: Medium
**대응 방안**:
- 마일스톤별 중간 점검
- 핵심 기능 우선 포팅
- 병렬 개발 가능한 작업 식별
- 필요시 범위 조정

### 운영 리스크

#### 배포 시스템 장애
**리스크**: npm 배포 과정에서 패키지 손상 또는 배포 실패
**확률**: Low
**영향도**: High
**대응 방안**:
- 철저한 배포 테스트
- 스테이징 환경 구축
- 롤백 계획 수립
- 모니터링 시스템 구축

---

## 성공 지표

### 기능 완성도 지표
- [ ] **Python 코드 제거율**: 100% (모든 .py 파일 제거)
- [ ] **TypeScript 구현율**: 100% (모든 기능 TypeScript 구현)
- [ ] **기능 동등성**: 100% (Python 버전과 동일 기능)
- [ ] **테스트 통과율**: 100% (모든 테스트 통과)

### 성능 개선 지표
- [ ] **실행 속도**: 80% 향상 (Python 대비)
- [ ] **메모리 사용량**: 50% 절약 (Python 대비)
- [ ] **설치 시간**: 30초 이내 (npm install -g)
- [ ] **파일 스캔**: 30% 성능 개선 (비동기 I/O)

### 품질 지표
- [ ] **타입 커버리지**: 100% (모든 함수 타입 정의)
- [ ] **테스트 커버리지**: ≥ 85%
- [ ] **ESLint 통과**: 0개 오류
- [ ] **빌드 성공**: < 1초 빌드 시간

### 사용자 경험 지표
- [ ] **설치 성공률**: ≥ 98%
- [ ] **마이그레이션 성공률**: ≥ 95%
- [ ] **크로스 플랫폼 호환**: 100% (Windows/macOS/Linux)
- [ ] **기존 프로젝트 호환**: 100% (.moai/, .claude/ 구조)

---

## 다음 단계 연결

### SPEC-014 준비
**다음 주제**: TypeScript 고성능 최적화 및 Rust 모듈 통합
**연결점**: Week 3 완전 포팅 완료 후 극한 성능 최적화

### SPEC-015 준비
**다음 주제**: MoAI-ADK 2.0 - 차세대 AI 개발 도구 생태계
**연결점**: TypeScript 단일 런타임 기반 확장 가능한 플러그인 시스템

### 지속적 개선
- **성능 모니터링**: 실사용 환경에서 성능 지표 수집
- **사용자 피드백**: TypeScript 전환 경험 분석
- **생태계 확장**: npm 생태계 통합 및 플러그인 개발
- **커뮤니티 기여**: 오픈소스 기여자 확대