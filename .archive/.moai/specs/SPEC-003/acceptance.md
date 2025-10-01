# SPEC-003 수락 기준: cc-manager 중심 Claude Code 최적화

## Acceptance Criteria

### AC1. cc-manager 강화 및 템플릿 지침 내장

**Given** cc-manager.md 파일이 업데이트될 때
**When** 파일을 열어 내용을 확인하면
**Then** 다음 요소들이 모두 포함되어 있어야 함:

#### 검증 기준:

- [ ] **커맨드 표준 템플릿** 섹션이 존재하고 YAML frontmatter 예시 포함
- [ ] **에이전트 표준 템플릿** 섹션이 존재하고 YAML frontmatter 예시 포함
- [ ] **Claude Code 공식 문서 핵심 내용** 완전 통합 (sub-agents, slash-commands 가이드)
- [ ] **파일 생성/수정 가이드라인** 섹션 존재
- [ ] **표준 검증 체크리스트** 포함
- [ ] cc-manager의 "중앙 관제탑" 역할이 명시됨
- [ ] **외부 참조 없이 완전한 지침 제공** (docs/cc-docs 참조 불필요)

**Given** cc-manager를 통해 새로운 파일을 생성할 때
**When** 커맨드 또는 에이전트 파일 생성을 요청하면
**Then** 표준 템플릿에 따라 올바른 YAML frontmatter 구조로 생성되어야 함

### AC2. 기존 파일 Claude Code 표준 준수

#### 커맨드 파일 표준 준수

**Given** .claude/commands/moai/ 디렉토리의 5개 파일이 존재할 때
**When** 각 파일의 YAML frontmatter를 검사하면
**Then** 모든 파일이 다음 구조를 가져야 함:

```yaml
---
name: moai:command-name
description: Clear one-line description
argument-hint: [param1] [param2]
allowed-tools: Tool1, Tool2, Task, Bash(cmd:*)
model: sonnet
---
```

#### 검증 기준:

- [ ] `name` 필드: "moai:" 접두사 포함, kebab-case 형식
- [ ] `description` 필드: 한 줄로 명확한 설명
- [ ] `argument-hint` 필드: 파라미터 힌트 배열
- [ ] `allowed-tools` 필드: 허용 도구 목록
- [ ] `model` 필드: "sonnet" 또는 "opus"

#### 에이전트 파일 표준 준수

**Given** .claude/agents/moai/ 디렉토리의 7개 파일이 존재할 때
**When** 각 파일의 YAML frontmatter를 검사하면
**Then** 모든 파일이 다음 구조를 가져야 함:

```yaml
---
name: agent-name
description: Use PROACTIVELY for [specific task trigger]
tools: Read, Write, Edit, MultiEdit, Bash, Glob, Grep
model: sonnet
---
```

#### 검증 기준:

- [ ] `name` 필드: kebab-case 형식의 에이전트 이름
- [ ] `description` 필드: "Use PROACTIVELY for" 패턴 포함
- [ ] `tools` 필드: 최소 권한 원칙에 따른 도구 목록
- [ ] `model` 필드: "sonnet" 또는 "opus"

### AC3. 프로액티브 트리거 조건 명확화

**Given** 각 에이전트의 description 필드에 프로액티브 조건이 명시될 때
**When** 에이전트 파일들을 검사하면
**Then** 다음 조건들이 만족되어야 함:

#### 검증 기준:

- [ ] **spec-builder**: "Use PROACTIVELY for SPEC creation based on project docs"
- [ ] **code-builder**: "Use PROACTIVELY for TDD implementation with TRUST principles validation"
- [ ] **doc-syncer**: "Use PROACTIVELY for documentation sync and PR status management"
- [ ] **git-manager**: "Use PROACTIVELY for all Git operations and workflow automation"
- [ ] **cc-manager**: "Use PROACTIVELY for Claude Code optimization and settings management"
- [ ] **debug-helper**: "Use PROACTIVELY for error diagnosis and development guide compliance"
- [ ] **project-manager**: "Use PROACTIVELY for project kickoff guidance and structure setup"

### AC4. 검증 도구 동작

**Given** validate_claude_standards.py 스크립트가 존재할 때
**When** 스크립트를 실행하면
**Then** 다음 기능들이 모두 동작해야 함:

#### 검증 기준:

- [ ] YAML frontmatter 파싱 성공
- [ ] 필수 필드 존재 확인 (name, description, tools/allowed-tools, model)
- [ ] 표준 구조 준수 여부 체크
- [ ] 프로액티브 패턴 검증 ("Use PROACTIVELY" 포함)
- [ ] 문제점 발견 시 구체적인 에러 메시지 출력
- [ ] 전체 검증 결과 요약 리포트 생성

**Given** 표준을 위반한 파일이 존재할 때
**When** 검증 스크립트를 실행하면
**Then** 위반 사항이 명확하게 식별되고 수정 방법이 제안되어야 함

### AC5. 핵심 문서 최적화

#### CLAUDE.md 최적화

**Given** CLAUDE.md 파일이 업데이트될 때
**When** 파일 내용을 확인하면
**Then** 다음 요소들이 포함되어야 함:

#### 검증 기준:

- [ ] cc-manager 역할이 강조된 별도 섹션 존재
- [ ] Claude Code 공식 문서 참조 링크 추가
- [ ] 4단계 워크플로우에서 cc-manager의 중심 역할 명시
- [ ] 표준 준수 중요성 언급

#### settings.json 권한 최적화

**Given** .claude/settings.json 파일이 업데이트될 때
**When** permissions 섹션을 확인하면
**Then** 다음 도구들이 허용 목록에 추가되어야 함:

#### 검증 기준:

- [ ] "WebSearch" - 웹 검색 도구
- [ ] "BashOutput" - 백그라운드 쉘 출력 확인
- [ ] "KillShell" - 백그라운드 쉘 종료
- [ ] "Bash(gemini:\*)" - Gemini 브릿지 명령어
- [ ] "Bash(codex:\*)" - Codex 브릿지 명령어

## 품질 게이트 통과 기준

### 필수 통과 조건

1. **표준 준수율**: 100% (모든 커맨드/에이전트 파일)
2. **검증 도구 통과율**: ≥95% (validate_claude_standards.py)
3. **프로액티브 조건 명시**: 100% (모든 에이전트)
4. **템플릿 지침 완성도**: 100% (cc-manager.md)

### 검증 방법

#### 자동화 검증

```bash
# 표준 준수 검증
python3 .moai/scripts/validate_claude_standards.py

# 파일 구조 검증
python3 -m pytest tests/integration/test_claude_standards.py

# 통합 워크플로우 테스트
python3 -m pytest tests/integration/test_cc_manager_workflow.py
```

#### 수동 검증

- **문서 가독성**: 모든 변경된 파일의 내용이 명확하고 일관성 있음
- **사용성**: cc-manager를 통한 파일 생성/수정이 직관적임
- **호환성**: 기존 MoAI-ADK 워크플로우와 완전 호환됨

### 성능 기준

- **파일 생성 시간**: 기존 대비 50% 단축 (템플릿 적용)
- **검증 실행 시간**: 전체 파일 검증이 10초 이내
- **메모리 사용량**: 검증 도구가 100MB 이하 사용

## 회귀 테스트 (Regression Testing)

### 기존 기능 보존

**Given** 기존 MoAI-ADK 워크플로우가 동작할 때
**When** SPEC-003 변경사항을 적용한 후 테스트하면
**Then** 모든 기존 기능이 정상 동작해야 함:

#### 검증 기준:

- [ ] `/moai:8-project` 명령어 정상 동작
- [ ] `/moai:1-spec` 명령어 정상 동작
- [ ] `/moai:2-build` 명령어 정상 동작
- [ ] `/moai:3-sync` 명령어 정상 동작
- [ ] `/moai:4-debug` 명령어 정상 동작
- [ ] 모든 에이전트 독립 실행 가능
- [ ] Git 워크플로우 정상 동작

### 하위 호환성

**Given** 기존 프로젝트가 MoAI-ADK를 사용하고 있을 때
**When** 업데이트된 버전을 적용하면
**Then** 기존 설정과 파일들이 계속 작동해야 함

## 완료 조건 (Definition of Done)

### 기능 완료 기준

- [ ] cc-manager.md 템플릿 지침 통합 완료
- [ ] 12개 파일(5개 커맨드 + 7개 에이전트) 표준화 완료
- [ ] validate_claude_standards.py 검증 도구 구현 완료
- [ ] 핵심 문서(CLAUDE.md, settings.json) 최적화 완료
- [ ] 모든 자동화 테스트 통과

### 문서화 완료 기준

- [ ] SPEC-003 문서 세트(spec.md, plan.md, acceptance.md) 작성
- [ ] cc-manager 사용 가이드 포함
- [ ] 표준화 체크리스트 제공
- [ ] 트러블슈팅 가이드 작성

### 품질 보증 기준

- [ ] TDD 사이클(Red-Green-Refactor) 완료
- [ ] 코드 리뷰 완료 (자기 검증 포함)
- [ ] 통합 테스트 100% 통과
- [ ] 문서 품질 검증 완료

---

**연결된 SPEC**:

- SPEC-003 명세서: [spec.md](./spec.md)
- 구현 계획: [plan.md](./plan.md)

**관련 TAG**: @TEST:ACCEPTANCE-003, @CODE:QA-003, @TEST:INTEGRATION-003
