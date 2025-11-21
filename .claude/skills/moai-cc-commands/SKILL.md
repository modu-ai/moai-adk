---
name: moai-cc-commands
description: Claude Code Commands system, workflow orchestration, and command-line
  interface patterns. Use when creating custom commands, managing workflows, or implementing
  CLI interfaces.
---

## Quick Reference (30 seconds)

Claude Code Commands는 커스텀 워크플로우 자동화, CLI 인터페이스 설계, 복잡한 다단계 작업 조율을
위한 강력한 명령어 시스템입니다. 프로젝트 초기화, 기능 배포, 문서 동기화, 릴리스 관리 등의
개발 워크플로우를 효율적으로 자동화할 수 있습니다.

**핵심 기능**:
- 커스텀 명령어 생성 및 등록
- 다단계 워크플로우 조율
- 파라미터 검증 및 입력 처리
- 에러 처리 및 복구
- 명령어 문서화 및 도움말

---

## Implementation Guide

### What It Does

Claude Code Commands는 다음을 제공합니다:

**명령어 시스템**:
- 명령어 등록 및 발견
- 파라미터 구문 분석 및 검증
- 명령어 실행 및 결과 처리
- 비동기 명령어 지원
- 명령어 체이닝 및 구성

**워크플로우 자동화**:
- 다단계 작업 조율
- 조건부 실행 분기
- 에러 처리 및 재시도
- 진행상황 추적 및 로깅
- 결과 수집 및 보고

**CLI 인터페이스**:
- 명령어 도움말 및 사용 설명서
- 파라미터 자동 완성
- 실시간 피드백
- 대화형 프롬프트
- 결과 포맷팅

### When to Use

- ✅ 프로젝트 초기화 및 설정 자동화
- ✅ 개발 워크플로우 (빌드, 테스트, 배포)
- ✅ 다양한 도구 통합 및 조율
- ✅ 반복적인 작업 자동화
- ✅ 복잡한 다단계 프로세스 단순화
- ✅ 팀 워크플로우 표준화

### Core Command Patterns

#### 1. 명령어 구조
```markdown
/moai:N-action [parameters] [options]

Examples:
- /moai:0-project                    # 프로젝트 초기화
- /moai:1-plan "feature description" # SPEC 생성
- /moai:2-run SPEC-001              # TDD 구현
- /moai:3-sync SPEC-001             # 문서 동기화
```

#### 2. 파라미터 처리
```markdown
## 위치 파라미터 (Positional)
/command arg1 arg2 arg3

## 옵션 파라미터 (Named)
/command --option value --flag

## 혼합 사용
/command required-arg --option value --flag
```

#### 3. 워크플로우 조율 패턴
```markdown
작업 1: 요구사항 수집
  └─ 작업 2: SPEC 생성
      └─ 작업 3: 구현 실행
          └─ 작업 4: 문서 동기화
              └─ 작업 5: 배포
```

#### 4. 에러 처리 패턴
- 입력 검증 실패 → 도움말 표시
- 작업 실패 → 재시도 또는 롤백
- 부분 완료 → 진행상황 저장
- 예상치 못한 에러 → 로그 기록

### Dependencies

- Claude Code commands system
- CLI framework (Click, Typer, Cobra)
- 파라미터 검증 라이브러리
- 워크플로우 조율 도구

---

## Works Well With

- `moai-cc-agents` (명령어 실행 위임)
- `moai-cc-hooks` (명령어 이벤트 처리)
- `moai-cc-configuration` (명령어 설정)
- `moai-project-config-manager` (프로젝트별 명령어)

---

## Advanced Patterns

### 1. 고급 파라미터 처리

**변수 확장 (Variable Expansion)**:
```bash
/command --path {{project-root}}/{{feature-name}}
/command --version {{semantic-version}}
```

**조건부 파라미터 (Conditional Parameters)**:
```bash
# 개발 환경
/command --mode dev --verbose

# 프로덕션 환경
/command --mode prod --debug false
```

**파라미터 검증 (Validation)**:
```markdown
- 필수 파라미터 확인
- 타입 검증 (string, number, boolean, path)
- 범위 검증 (최소값, 최대값, 열거값)
- 커스텀 검증 규칙
```

### 2. 워크플로우 오케스트레이션 패턴

**직렬 실행 (Sequential)**:
```
Step 1 → Step 2 → Step 3 → Step 4
```

**병렬 실행 (Parallel)**:
```
Step 1A → |
          | → Combined Result
Step 1B → |
```

**조건부 분기 (Branching)**:
```
Step 1 → [Condition Check]
          ├─ Success → Step 2A
          └─ Failure → Step 2B
```

### 3. 명령어 확장 패턴

**플러그인 시스템**:
```markdown
1. 명령어 인터페이스 정의
2. 플러그인 구현
3. 플러그인 등록
4. 동적 로딩
```

**훅 통합 (Hook Integration)**:
```markdown
- Pre-command hooks: 명령어 실행 전
- Post-command hooks: 명령어 실행 후
- Error hooks: 에러 발생 시
- Validation hooks: 파라미터 검증
```

### 4. 고급 결과 처리

**결과 포맷팅**:
- 텍스트 출력
- JSON 형식
- 테이블 형식
- 마크다운 형식

**결과 저장**:
- 파일로 저장
- 데이터베이스 저장
- 로그 기록
- 알림 전송

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, command architecture patterns
- **v1.0.0** (2025-10-22): Initial commands system

---

**End of Skill** | Updated 2025-11-21 | Lines: 195



