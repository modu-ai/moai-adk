# SPEC-CH08-001 구현 계획

**SPEC**: Claude Code Plugins & Migration
**작성 기간**: 2025-10-30 ~ 2025-11-05 (1주)
**담당**: Alfred SuperAgent
**목표**: ch08 완벽한 플러그인 가이드 작성

---

## 📋 작성 계획

### Phase 1: 아키텍처 섹션 (Day 1-2)

#### 8-1: Plugin Architecture Overview (1000단어)
**목표**: 플러그인 시스템 전체 이해

**내용**:
- 플러그인이란? (정의, 목적)
- Output Styles vs Plugins 비교
- 플러그인 생명주기 (설치 → 활성화 → 실행)
- 5개 v1.0 플러그인 소개

**실행**:
- 아키텍처 다이어그램 1개
- 비교 테이블 1개
- 예제 screenshot 2개

---

#### 8-2: Plugin.json Schema Deep Dive (1500단어)
**목표**: plugin.json 완벽한 이해

**내용**:
- plugin.json 구조 (metadata, commands, agents, hooks, permissions)
- 각 필드 상세 설명 및 필수/선택 여부
- 스키마 검증 규칙
- 실제 예제 (PM Plugin, UI/UX Plugin)

**실행**:
- JSON 스키마 예제 3개
- 필드별 설명 테이블 1개
- 실패 케이스 + 해결책 2개

---

### Phase 2: 패턴 섹션 (Day 2-3)

#### 8-3: Command Development Patterns (1000단어)
**목표**: 명령어 개발 방법론

**내용**:
- Command 정의 및 구문
- Command 템플릿 작성법
- 인자 처리 (required, optional, flags)
- 실행 흐름 (User → Agent → Output)

**실행**:
- Command 템플릿 예제 3개
- 명령어 실행 흐름도 1개
- 실제 `/init-pm` 예제 상세 설명

---

#### 8-4: Agent Orchestration (1200단어)
**목표**: Agent를 통한 복잡한 워크플로우 구현

**내용**:
- Plugin 내 Agent의 역할
- Agent 정의 (agent.md)
- Tool 권한 제어
- Sub-agent 호출 및 조율

**실행**:
- Agent 정의 예제 2개
- Agent 실행 흐름도 1개
- 여러 Agent 조율 예제 1개

---

### Phase 3: 고급 기능 섹션 (Day 4-5)

#### 8-5: Hook Lifecycle & Events (1200단어)
**목표**: Hook 시스템 완벽 이해

**내용**:
- Hook이란? (이벤트 기반 반응)
- Hook 종류 (SessionStart, PreToolUse, PostToolUse, SessionEnd)
- Hook 개발 패턴 (JavaScript/TypeScript)
- 우선순위 및 타임아웃 관리

**실행**:
- Hook 정의 예제 3개
- Hook 실행 순서도 1개
- 실제 권한 검증 Hook 예제

---

#### 8-6: Skill Integration for Plugins (1000단어)
**목표**: Plugin에서 Skill 활용하기

**내용**:
- Skill이란? (재사용 가능한 지식)
- 플러그인 Skill 구조 (SKILL.md, examples.md, reference.md)
- Skill 로딩 및 Agent 활용
- 플러그인 Skill 예제 (moai-plugin-scaffolding)

**실행**:
- Skill 구조 예제 2개
- Agent에서 Skill 활용 코드 2개
- 실제 plgin Skill README 예제

---

#### 8-7: Permission Model & Security (1000단어)
**목표**: 플러그인 권한 모델 및 보안

**내용**:
- Deny-by-Default 전략
- allowedTools vs deniedTools
- 런타임 권한 검증 (Hook)
- 보안 모범 사례

**실행**:
- 권한 모델 다이어그램 1개
- 안전/위험한 패턴 비교 1개
- 권한 위반 시나리오 + 해결책 2개

---

### Phase 4: 마이그레이션 및 FAQ (Day 5-6)

#### 8-8: Migration Path (Output Styles → Plugins) (800단어)
**목표**: v0.x 사용자를 위한 마이그레이션 가이드

**내용**:
- Output Styles의 역사 및 한계
- Plugin으로 이행하는 이유
- 단계별 마이그레이션 체크리스트
- 호환성 및 대체재

**실행**:
- 마이그레이션 체크리스트 1개
- 변환 예제 (Output Style → Plugin) 2개
- 이슈 해결 가이드 2개

---

#### 8-9: FAQ & Troubleshooting (500단어)
**목표**: 자주 묻는 질문 해결

**내용**:
- 플러그인 설치/제거 FAQ (5개)
- 명령어 개발 FAQ (3개)
- Hook 개발 FAQ (2개)
- 보안/권한 FAQ (3개)

**실행**:
- Q&A 형식 (질문 → 답변 → 예제)
- 실패 사례 + 해결책

---

### Phase 5: Hands-on Labs (Day 7)

#### Lab 8A: 간단한 Command-Only Plugin 만들기 (500단어)
**목표**: 가장 단순한 플러그인 경험

**내용**:
1. Plugin 디렉토리 구조 생성
2. plugin.json 작성 (command 1개, agent 없음)
3. Command 템플릿 작성
4. Simple Agent 구현
5. 플러그인 설치 및 테스트

**예제**: Simple "hello" command

---

#### Lab 8B: Agent가 포함된 Plugin 만들기 (700단어)
**목표**: 복잡한 워크플로우 플러그인

**내용**:
1. 다중 Agent 구조 설계
2. plugin.json에 여러 Agent 등록
3. Agent 간 조율 로직
4. Skill 통합
5. 통합 테스트

**예제**: SPEC 생성 플러그인 (PM Plugin 유사)

---

#### Lab 8C: Hook 등록 및 테스트 (600단어)
**목표**: 이벤트 기반 반응 시스템

**내용**:
1. hooks.json 정의 (SessionStart, PreToolUse)
2. Hook 핸들러 구현
3. 권한 검증 로직 추가
4. Hook 실행 순서 확인
5. 타임아웃 처리

**예제**: 권한 검증 Hook

---

#### Lab 8D: 권한 모델 구현 (600단어)
**목표**: 보안 강화된 플러그인

**내용**:
1. allowedTools 정의 (최소 권한)
2. deniedTools 명시
3. PreToolUse Hook에서 검증
4. 위반 사항 처리
5. 보안 감사

**예제**: Read-Only 플러그인 + 일부 Write 권한

---

### 참고 자료

| 섹션 | 분량 | 실행 요소 | 완료 기준 |
|------|------|---------|----------|
| 8-1 | 1000 단어 | 다이어그램 1, 스크린샷 2 | ✅ |
| 8-2 | 1500 단어 | JSON 3, 테이블 1, 케이스 2 | ✅ |
| 8-3 | 1000 단어 | 템플릿 3, 흐름도 1 | ✅ |
| 8-4 | 1200 단어 | 예제 2, 흐름도 1, 조율 1 | ✅ |
| 8-5 | 1200 단어 | 예제 3, 순서도 1, Hook 1 | ✅ |
| 8-6 | 1000 단어 | 구조 2, 코드 2, README 1 | ✅ |
| 8-7 | 1000 단어 | 다이어그램 1, 비교 1, 시나리오 2 | ✅ |
| 8-8 | 800 단어 | 체크리스트 1, 변환 2, 가이드 2 | ✅ |
| 8-9 | 500 단어 | Q&A 13개 | ✅ |
| **Lab 8A-D** | **2400 단어** | **4개 랩** | ✅ |
| **총합** | **9,800 단어** | **20+ 실행 요소** | |

---

## 📊 일정

| 일자 | 작업 | 산출물 | 상태 |
|------|------|--------|------|
| Day 1 | SPEC 작성 완료 | spec.md, plan.md, acceptance.md | ✅ |
| Day 2-3 | 8-1, 8-2 섹션 작성 | 2000 단어 + 자료 3개 | 🎯 |
| Day 4-5 | 8-3, 8-4, 8-5, 8-6 섹션 | 4400 단어 + 자료 8개 | 🎯 |
| Day 6-7 | 8-7, 8-8, 8-9, Lab 작성 | 2400 단어 + Lab 4개 | 🎯 |
| Day 7 | 최종 검수 및 QA | 오타, 링크, 실행 확인 | 🎯 |

---

## ✅ 완료 기준

### 문서 완성도
- [ ] 8-1 ~ 8-9 섹션 모두 작성 (총 9,800단어)
- [ ] Lab 8A-D 모두 실행 가능
- [ ] 모든 코드 예제 테스트됨
- [ ] 모든 다이어그램 명확함
- [ ] 모든 링크 작동함

### 품질 기준
- [ ] 오타/문법 0개
- [ ] 일관된 용어 사용 (moai-alfred)
- [ ] FAQ 13개 이상
- [ ] 실제 예제 15개 이상
- [ ] 다이어그램/스크린샷 8개 이상

### 실행 가능성
- [ ] Lab 8A: 누구나 5분 내 완료 가능
- [ ] Lab 8B: 누구나 10분 내 완료 가능
- [ ] Lab 8C: 누구나 8분 내 완료 가능
- [ ] Lab 8D: 누구나 10분 내 완료 가능

---

**작성자**: Alfred SuperAgent
**시작 일**: 2025-10-30
**목표 완료 일**: 2025-11-05
