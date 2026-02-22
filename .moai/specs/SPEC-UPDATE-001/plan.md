# Implementation Plan: SPEC-UPDATE-001

## Overview

MoAI-ADK의 설치 모드를 Global과 Local로 구분하고, moai init 및 moai update 명령어 동작을 재설계합니다.

---

## Milestones

### Milestone 1: Configuration Schema Update

Priority: Primary Goal
Description: 설치 모드 설정을 위한 데이터 구조 및 설정 파일 스키마 업데이트

Tasks:
1. internal/config/types.go에 InstallationMode 필드 추가
2. internal/config/defaults.go에 기본값 설정 (기본값: "global")
3. internal/defs/에 상수 정의 추가
4. 설정 파일 파싱/직렬화 로직 업데이트

Target Files:
- internal/config/types.go
- internal/config/defaults.go
- internal/defs/constants.go (또는 기존 파일)

Acceptance Criteria:
- [ ] InstallationMode 타입 정의 ("global" | "local")
- [ ] 기본값이 "global"로 설정됨
- [ ] YAML 파싱/직렬화 정상 동작
- [ ] 기존 설정 파일과 호환됨 (필드 없으면 기본값 사용)

---

### Milestone 2: Deployer Enhancement for Global/Local Mode

Priority: Primary Goal
Description: 템플릿 배포 로직을 Global/Local 모드에 따라 분기 처리

Tasks:
1. Deployer 인터페이스에 installationMode 파라미터 추가
2. Global 모드 배포 로직 구현 (~/.claude/ 대상)
3. Local 모드 배포 로직 구현 (기존 동작 유지)
4. 파일 배치 전략 구현 (시스템 파일 vs 프로젝트 파일 구분)

Target Files:
- internal/template/deployer.go
- internal/template/deployer_global.go (신규)
- internal/template/deployer_local.go (신규)

Acceptance Criteria:
- [ ] Global 모드에서 ~/.claude/agents/moai/ 생성
- [ ] Global 모드에서 ~/.claude/skills/moai*/ 생성
- [ ] Global 모드에서 ~/.claude/rules/moai/ 생성
- [ ] Global 모드에서 프로젝트 .claude/에 hooks만 배치
- [ ] Local 모드에서 기존 동작 유지

---

### Milestone 3: moai init Command Refactoring

Priority: Primary Goal
Description: moai init 명령어에 설치 모드 선택 기능 추가

Tasks:
1. .moai/ 디렉토리 존재 확인 로직 추가
2. 설치 모드 선택 AskUserQuestion 구현
3. Global/Local 모드에 따른 배포 로직 호출
4. 설정 파일에 installation_mode 저장
5. 초기화 안내 메시지 개선

Target Files:
- internal/cli/init.go
- internal/cli/wizard/run.go

Acceptance Criteria:
- [ ] .moai/ 없으면 "moai init . 으로 초기화하세요" 메시지 표시
- [ ] Global/Local 모드 선택 UI 제공
- [ ] Global 모드 선택 시 ~/.claude/에 시스템 파일 설치
- [ ] Local 모드 선택 시 기존 동작 유지
- [ ] installation_mode가 system.yaml에 저장됨

---

### Milestone 4: moai update Command Redesign

Priority: Primary Goal
Description: moai update 명령어 동작 재설계 (-c 옵션 제거, 설정 재설정 질문 추가)

Tasks:
1. -c 옵션 제거
2. "설정을 재설정하시겠습니까? (y/n)" 질문 추가
3. 'y' 응답 시 설정 wizard 재실행
4. 'n' 응답 시 기존 설정 백업 후 병합
5. Global 모드 변경 질문 추가
6. Global/Local 모드 전환 로직 구현

Target Files:
- internal/cli/update.go
- internal/cli/wizard/run.go

Acceptance Criteria:
- [ ] -c 플래그가 제거됨
- [ ] 설정 재설정 질문이 표시됨
- [ ] 'y' 응답 시 wizard 재실행
- [ ] 'n' 응답 시 백업 후 병합
- [ ] Global 모드 변경 질문이 표시됨
- [ ] Local to Global 전환 시 불필요한 로컬 파일 삭제
- [ ] Global to Local 전환 시 시스템 파일 로컬 복사

---

### Milestone 5: moai glm Command Update

Priority: Secondary Goal
Description: moai glm이 settings.local.json만 수정하도록 변경

Tasks:
1. glm.go에서 settings.json 수정 로직 제거
2. settings.local.json 생성/수정 로직만 유지
3. 기존 settings.json의 GLM 관련 항목 정리 (마이그레이션)

Target Files:
- internal/cli/glm.go

Acceptance Criteria:
- [ ] moai glm 실행 시 settings.local.json만 수정
- [ ] settings.json은 수정하지 않음
- [ ] 기존 settings.json의 GLM 설정이 settings.local.json으로 마이그레이션됨

---

### Milestone 6: Mode Transition Logic

Priority: Secondary Goal
Description: Global to Local 모드 전환 기능 구현

Tasks:
1. Local to Global 전환 로직 구현
   - .claude/agents/moai/ 삭제
   - .claude/skills/moai*/ 삭제
   - .claude/rules/moai/ 삭제
   - hooks, settings.json, settings.local.json, .moai 유지
2. Global to Local 전환 로직 구현
   - ~/.claude/에서 로컬로 시스템 파일 복사
3. 전환 확인 UI 구현

Target Files:
- internal/cli/update.go
- internal/template/deployer.go

Acceptance Criteria:
- [ ] Local to Global 전환 시 불필요한 로컬 파일 삭제
- [ ] Global to Local 전환 시 시스템 파일 로컬 복사
- [ ] 프로젝트 설정 파일 (hooks, settings, .moai) 보존
- [ ] 전환 전 사용자 확인

---

### Milestone 7: Testing and Documentation

Priority: Final Goal
Description: 단위 테스트, 통합 테스트 작성 및 문서 업데이트

Tasks:
1. 각 마일스톤별 단위 테스트 작성
2. 통합 테스트 시나리오 작성
3. README.md 업데이트
4. CLAUDE.md 업데이트 (필요 시)
5. CHANGELOG.md 업데이트

Target Files:
- internal/cli/*_test.go
- internal/template/*_test.go
- README.md
- CHANGELOG.md

Acceptance Criteria:
- [ ] 모든 새 기능에 대한 단위 테스트 작성
- [ ] 통합 테스트로 E2E 시나리오 검증
- [ ] README에 설치 모드 설명 추가
- [ ] CHANGELOG에 변경 사항 기록

---

## Technical Approach

### Architecture Pattern

- Strategy Pattern: Global/Local 배포 전략을 별도 타입으로 분리
- Factory Pattern: 설치 모드에 따른 Deployer 생성

### Key Design Decisions

1. 기본값: Global 모드를 기본값으로 설정 (대부분의 사용 사례에 적합)
2. 마이그레이션: 자동 마이그레이션보다 명시적 선택 유도
3. 호환성: 기존 프로젝트는 Local 모드로 자동 감지

### Dependencies

- 기존 internal/template/deployer.go
- 기존 internal/cli/init.go, internal/cli/update.go
- github.com/spf13/cobra (CLI 프레임워크)

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| 기존 사용자 혼란 | Medium | 명확한 마이그레이션 가이드 제공 |
| Global 설치 권한 문제 | Low | ~/.claude/ 사용으로 권한 문제 최소화 |
| 멀티 프로젝트 버전 충돌 | Medium | 프로젝트별 Local 모드 옵션 유지 |

---

## Definition of Done

- [ ] 모든 EARS 요구사항 구현 완료
- [ ] 단위 테스트 커버리지 85% 이상
- [ ] 통합 테스트 통과
- [ ] 코드 리뷰 완료
- [ ] 문서 업데이트 완료
- [ ] TRUST 5 품질 게이트 통과

---

Version: 1.0.0
Last Updated: 2026-02-21
