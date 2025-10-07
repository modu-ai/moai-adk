# SPEC-INIT-004 구현 계획서

## 목표

Git 초기화 워크플로우를 개선하여 사용자 경험을 향상시키고, 초기화 속도를 67% 개선(3분 → 1분)합니다.

## 구현 범위

### Phase 1: Git 자동 감지 유틸리티 작성 (우선순위: 최고)

**산출물**: `src/utils/git-detector.ts`

**기능**:
1. `.git` 폴더 존재 확인
2. Git 저장소 정보 수집 (커밋 수, 브랜치, remote)
3. GitHub URL 자동 추출

**상세 작업**:
- [ ] `detectGitStatus()` 함수 작성
  - `git rev-parse --is-inside-work-tree` 검증
  - `git rev-list --count HEAD` 커밋 수
  - `git branch --show-current` 브랜치
  - `git remote -v` remote 목록
- [ ] `detectGitHubRemote()` 함수 작성
  - 정규식: `/github\.com[:/]([^/]+)\/([^/.]+)/`
  - HTTPS/SSH 형식 모두 지원
- [ ] `autoInitGit()` 함수 작성
  - `git init` 실행
  - 에러 핸들링 (Git 미설치, 권한 오류)
- [ ] 단위 테스트 작성 (TDD)

---

### Phase 2: 언어 제한 적용 (우선순위: 높음)

**수정 파일**:
- `src/cli/prompts/init/definitions.ts`
- `templates/.moai/config.json`
- `src/utils/i18n.ts`

**상세 작업**:
- [ ] 언어 선택 프롬프트에서 ja, zh 제거
- [ ] locale 타입 제약: `"ko" | "en"`
- [ ] 기존 ja/zh 리소스 제거 또는 주석 처리
- [ ] 폴백 로직: ja/zh → en
- [ ] 통합 테스트

---

### Phase 3: Git 자동 초기화 로직 추가 (우선순위: 높음)

**수정 파일**:
- `src/cli/commands/init/interactive-handler.ts`
- `src/core/installer/orchestrator.ts`

**상세 작업**:
- [ ] `.git` 없으면 자동 `git init` 실행
- [ ] 초기화 메시지 표시: "Git 저장소 초기화 완료"
- [ ] 기존 저장소 유지/삭제 질문 추가
- [ ] 백업 로직: `.git-backup-{timestamp}/`
- [ ] 통합 테스트

---

### Phase 4: GitHub 자동 감지 및 설정 (우선순위: 중간)

**수정 파일**:
- `src/cli/commands/init/interactive-handler.ts`
- `src/cli/config/config-builder.ts`

**상세 작업**:
- [ ] GitHub remote 자동 감지
- [ ] `.moai/config.json`에 자동 저장
- [ ] GitHub URL 변경 옵션 추가 (선택)
- [ ] Team 모드: GitHub 필수 검증
- [ ] Personal 모드: GitHub 선택 사항
- [ ] 통합 테스트

---

### Phase 5: 질문 흐름 단순화 (우선순위: 중간)

**수정 파일**:
- `src/cli/commands/init/interactive-handler.ts`
- `src/cli/prompts/init/definitions.ts`

**상세 작업**:
- [ ] 자동 감지 기반 질문 건너뛰기
- [ ] `.git` 있음 + GitHub 있음: 0~1개 질문
- [ ] `.git` 있음 + GitHub 없음: 1~2개 질문
- [ ] `.git` 없음: 0~1개 질문
- [ ] 질문 순서 최적화
- [ ] 통합 테스트

---

### Phase 6: 비대화형 모드 옵션 (우선순위: 낮음)

**수정 파일**:
- `src/cli/commands/init/index.ts`
- `src/cli/commands/init/non-interactive-handler.ts`

**상세 작업**:
- [ ] `--auto-git` 옵션 추가
- [ ] `--locale <locale>` 옵션 추가 (ko|en 검증)
- [ ] CI/CD 환경 대응
- [ ] 통합 테스트

---

### Phase 7: 테스트 및 문서화 (우선순위: 필수)

**테스트**:
- [ ] 단위 테스트: `tests/utils/git-detector.test.ts`
- [ ] 통합 테스트: `tests/cli/commands/init/git-workflow.test.ts`
- [ ] E2E 테스트: 실제 프로젝트 초기화 시나리오
- [ ] 테스트 커버리지 ≥85%

**문서**:
- [ ] `docs/guides/git-workflow-improvement.md` 작성
- [ ] CHANGELOG 업데이트 (Breaking Changes)
- [ ] README 업데이트

---

## TDD 전략

**RED 단계**:
1. `.git` 감지 테스트 작성 (실패 확인)
2. GitHub URL 추출 테스트 작성 (실패 확인)
3. 자동 초기화 테스트 작성 (실패 확인)

**GREEN 단계**:
1. 최소 구현으로 테스트 통과
2. 모든 시나리오 커버

**REFACTOR 단계**:
1. 코드 품질 개선 (함수 분리, 타입 안전성)
2. 에러 핸들링 추가
3. TDD 이력 주석 추가

---

## 리스크 관리

### 리스크 1: Git 미설치 환경

**확률**: 낮음 (5%)
**영향**: 높음 (초기화 실패)

**대응**:
- `moai doctor` 명령어에서 Git 검증
- Git 미설치 시 명확한 에러 메시지 + 설치 가이드

---

### 리스크 2: GitHub 접근 권한 부족

**확률**: 중간 (20%)
**영향**: 중간 (첫 push 실패)

**대응**:
- URL 형식 검증만 수행 (실제 접근은 검증하지 않음)
- 에러 메시지에 권한 확인 안내 포함

---

### 리스크 3: 기존 사용자 혼란 (ja, zh 제거)

**확률**: 중간 (10%)
**영향**: 낮음 (en으로 폴백 가능)

**대응**:
- CHANGELOG에 Breaking Changes 명시
- 마이그레이션 가이드 제공
- 폴백 로직: ja/zh → en

---

### 리스크 4: 복잡한 Git 상태 (충돌, detached HEAD 등)

**확률**: 낮음 (5%)
**영향**: 중간 (정보 수집 실패)

**대응**:
- 기본적인 감지만 수행 (커밋 수, 브랜치, remote)
- 복잡한 상태는 경고 메시지만 표시하고 계속 진행
- 사용자에게 수동 확인 권장

---

## 의존성

**외부 라이브러리**:
- `simple-git` v3.28.0 (기존 사용 중)
- `inquirer` v12.9.6 (기존 사용 중)

**내부 의존성**:
- `src/utils/tty-detector.ts` (SPEC-INIT-001)
- `src/cli/config/config-builder.ts`
- `src/core/installer/orchestrator.ts`

**Breaking Changes**:
- 언어 지원 제거 (ja, zh)

---

## 마일스톤

**Milestone 1: Git 자동 감지 (Phase 1)**
- GitDetector 유틸리티 완성
- 단위 테스트 통과

**Milestone 2: 언어 제한 (Phase 2)**
- ko, en만 지원
- 폴백 로직 적용

**Milestone 3: 자동 초기화 (Phase 3)**
- `.git` 없으면 자동 `git init`
- 기존 저장소 보호

**Milestone 4: GitHub 자동 감지 (Phase 4)**
- GitHub URL 자동 추출
- config.json 자동 저장

**Milestone 5: 통합 완료 (Phase 5-7)**
- 질문 흐름 최적화
- 테스트 커버리지 ≥85%
- 문서화 완료

---

## 롤백 계획

**롤백 트리거**:
- 테스트 커버리지 < 80%
- 크리티컬 버그 발생 (Git 저장소 손실 등)
- 사용자 피드백 부정적 (60% 이상)

**롤백 방법**:
1. Git 브랜치 revert
2. 이전 버전 (v0.2.11) 배포
3. 사용자 공지

---

## 완료 조건 (Definition of Done)

- [ ] Phase 1-7 모두 완료
- [ ] 단위 테스트 통과 (커버리지 ≥85%)
- [ ] 통합 테스트 통과
- [ ] TRUST 5원칙 검증 완료
- [ ] @TAG 체인 무결성 확인
- [ ] 문서화 완료 (CHANGELOG, README, 가이드)
- [ ] PR 리뷰 승인 (Team 모드)
- [ ] develop 브랜치 병합 (Team 모드)
