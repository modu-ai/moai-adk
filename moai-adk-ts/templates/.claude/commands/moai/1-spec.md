---
name: moai:1-spec
description: EARS 명세 작성 + 브랜치/PR 생성
argument-hint: "제목1 제목2 ... | SPEC-ID 수정내용"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# MoAI-ADK 1단계: EARS 명세 작성 + 브랜치/PR 생성

**SPEC 생성 대상**: ${ARGUMENTS:-"SPEC 후보"}

## 🚀 SPEC 작성 및 브랜치 생성

프로젝트 문서를 분석하여 SPEC 후보를 제안하고, 선택된 SPEC을 즉시 생성합니다.

## 핵심 기능

- **스마트 분석**: 프로젝트 문서를 분석하여 SPEC 후보 자동 제안
- **Personal 모드**: `.moai/specs/SPEC-XXX/` 디렉터리 생성 + 로컬 브랜치
- **Team 모드**: GitHub Issue 생성 + 원격 브랜치 + PR 템플릿

## 사용법

```bash
/moai:1-spec                      # 자동 제안 (권장)
/moai:1-spec "JWT 인증 시스템"       # 수동 생성
/moai:1-spec SPEC-001 "보안 강화"   # 기존 SPEC 수정
```

## 워크플로우

### 1. 프로젝트 분석
- product/structure/tech.md 문서 스캔
- 기존 SPEC 목록 및 우선순위 검토
- 핵심 요구사항 추출

### 2. SPEC 후보 제안
- 비즈니스 가치 기반 우선순위 설정
- 기술적 제약사항 반영
- 3-5개 후보 리스트 생성

### 3. SPEC 문서 생성
- **EARS 방법론**: Easy Approach to Requirements Syntax (5가지 구문 형식)
- **3개 파일**: spec.md, plan.md, acceptance.md
- **@TAG**: 명령어가 tag-agent를 호출하여 @REQ → @DESIGN → @TASK → @TEST 체인 생성

### EARS (Easy Approach to Requirements Syntax) 작성법

#### EARS 구문 형식
1. **Ubiquitous Requirements**: 시스템은 [기능]을 제공해야 한다
2. **Event-driven Requirements**: WHEN [조건]이면, 시스템은 [동작]해야 한다
3. **State-driven Requirements**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
4. **Optional Features**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
5. **Constraints**: IF [조건]이면, THEN 시스템은 [제약 동작]해야 한다

#### EARS 작성 예시
```markdown
### Ubiquitous Requirements (언제나 적용)
- 시스템은 사용자 인증 기능을 제공해야 한다
- 시스템은 JWT 토큰 기반 세션 관리를 지원해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 이메일과 패스워드로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 액세스 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 잘못된 자격증명이 제공되면, 시스템은 로그인을 거부해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
- WHILE 토큰이 유효한 상태일 때, 시스템은 API 요청을 처리해야 한다

### Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다
- WHERE 2FA가 활성화되면, 시스템은 추가 인증을 요구할 수 있다

### Constraints (제약사항)
- IF 토큰이 변조된 상태라면, THEN 시스템은 모든 요청을 거부해야 한다
- IF 사용자가 3회 연속 로그인에 실패하면, THEN 계정을 30분간 잠금해야 한다
- IF 액세스 토큰 만료시간이 15분을 초과하면, THEN 시스템은 토큰 생성을 거부해야 한다
- IF 리프레시 토큰이 7일을 초과하면, THEN 시스템은 재인증을 요구해야 한다
```

### 4. Git 작업 자동화
- **Personal 모드**: 로컬 브랜치 + 체크포인트
- **Team 모드**: GitHub Issue + 원격 브랜치 + PR

## 실행 순서

### 1. 순수 SPEC 문서 생성 (TAG 작업 제외)

먼저 spec-builder 에이전트로 순수한 EARS 방법론 SPEC을 생성합니다:

@agent-spec-builder "${ARGUMENTS:-"프로젝트 분석을 통한 SPEC 후보"}를 위한 EARS 방법론 SPEC 생성해주세요. TAG 관련 작업은 제외하고 SPEC 내용과 TAG 요구사항 YAML만 작성해주세요."

- 프로젝트 문서 분석 및 SPEC 후보 제안
- 사용자 선택 후 EARS 5개 구문 형식 SPEC 작성
- 3개 파일 동시 생성 (spec.md, plan.md, acceptance.md)
- **TAG 요구사항 YAML 내에 도메인, 키워드, 의존성 명시**

### 2. TAG 시스템 전체 관리 (독점 처리)

SPEC 생성 후 tag-agent가 생성된 TAG 요구사항 YAML을 바탕으로 전체 TAG 작업을 수행합니다:

@agent-tag-agent "생성된 SPEC 기반으로 완전한 TAG 체인 생성하고 기존 TAG 검색, 중복 방지, 체인 검증, 인덱스 업데이트를 수행해주세요"

- **기존 TAG 검색**: ripgrep으로 도메인 관련 TAG 발견 및 재사용 검토
- **TAG 생성**: @REQ → @DESIGN → @TASK → @TEST Primary Chain 생성
- **중복 방지**: TAG 형식 및 고유성 검증
- **체인 검증**: 체인 무결성 및 순환 참조 방지
- **인덱스 관리**: JSONL 기반 분산 인덱스 업데이트

### 3. 브랜치 생성 자동화 (신규 책임)

SPEC과 TAG 작업 완료 후 브랜치 생성 및 초기 설정을 자동화합니다:

@agent-git-manager "SPEC 및 TAG 생성 완료, 브랜치 생성 및 초기 커밋을 수행해주세요"

#### 브랜치 생성 단계

1. **기존 브랜치 확인**:
   ```bash
   # 중복 브랜치 검색
   git branch --list "feature/spec-${SPEC_ID}*"
   git branch --list "*${DOMAIN}*"

   # 중복 발견 시
   → 사용자에게 경고: "유사한 브랜치가 존재합니다. 계속하시겠습니까? (y/n)"
   → 정리 제안: "기존 브랜치를 정리하시겠습니까? (y/n)"
   ```

2. **사용자 확인 요청**:
   ```
   📋 브랜치 생성 계획:
   - 브랜치명: feature/spec-${SPEC_ID}-${DESCRIPTION}
   - 기반 브랜치: main (또는 develop)
   - SPEC 파일: .moai/specs/SPEC-${SPEC_ID}/
   - TAG 체인: @REQ → @DESIGN → @TASK

   생성하시겠습니까? (y/n)
   ```

3. **브랜치 생성 (승인 시)**:
   ```bash
   # Personal 모드
   git checkout -b feature/spec-${SPEC_ID}-${DESCRIPTION}
   git add .moai/specs/SPEC-${SPEC_ID}/
   git commit -m "📝 SPEC: [SPEC-${SPEC_ID}] ${DESCRIPTION} SPEC 생성

   @REQ:${SPEC_ID}, @DESIGN:${SPEC_ID}, @TASK:${SPEC_ID}

   🤖 Generated with Claude Code"

   # Team 모드 (추가)
   git push -u origin feature/spec-${SPEC_ID}-${DESCRIPTION}
   gh pr create --draft \
     --title "🚧 [SPEC-${SPEC_ID}] ${DESCRIPTION}" \
     --body "$(cat <<'EOF'
   ## SPEC 개요
   ${SPEC_SUMMARY}

   ## TAG 체인
   - @REQ:${SPEC_ID}
   - @DESIGN:${SPEC_ID}
   - @TASK:${SPEC_ID}

   ## 다음 단계
   - [ ] /moai:2-build로 TDD 구현
   - [ ] 품질 게이트 통과
   - [ ] /moai:3-sync로 문서 동기화

   🤖 Generated with Claude Code
   EOF
   )"
   ```

4. **브랜치 생성 완료 보고**:
   ```
   ✅ 브랜치 생성 완료
   📂 브랜치: feature/spec-${SPEC_ID}-${DESCRIPTION}
   📝 SPEC 파일: .moai/specs/SPEC-${SPEC_ID}/spec.md
   🏷️ TAG 체인: @REQ → @DESIGN → @TASK

   🚀 다음 단계: /moai:2-build SPEC-${SPEC_ID}
   ```

#### 브랜치 충돌 방지 정책

- **중복 브랜치 감지**: 동일 SPEC ID 또는 유사 도메인 브랜치 확인
- **자동 정리 제안**: 머지된 feature 브랜치 자동 삭제 제안
- **병렬 작업 감지**: 동일 SPEC에 대한 여러 브랜치 생성 방지

## 품질 기준

- **EARS 방법론**: Easy Approach to Requirements Syntax 5가지 구문 형식 완전 준수
- **에이전트 역할 분리**: 각 에이전트 고유 책임 영역 100% 준수
- **Acceptance Criteria**: Given-When-Then 시나리오 최소 2개
- **TAG 위임 완료**:
  - spec-builder: TAG 요구사항 YAML 완성 (tag-agent 위임)
  - tag-agent: TAG 생성, 검증, 체인 관리, 인덱스 업데이트 독점 처리
  - 중복 작업 0건 (각 에이전트 단일 책임)
  - 오케스트레이션 품질: 에이전트 간 데이터 전달 무결성

## 다음 단계

- `/moai:2-build SPEC-XXX`: TDD 구현 및 브랜치 머지
  - Red-Green-Refactor 사이클 실행
  - 품질 게이트 검증 후 main 브랜치로 머지
  - **브랜치 종료는 이 단계에서 수행됨**

- `/moai:3-sync`: 문서 동기화 및 TAG 검증 (브랜치 관리 없음)
  - Living Document 동기화
  - TAG 무결성 검증
  - sync-report.md 생성
