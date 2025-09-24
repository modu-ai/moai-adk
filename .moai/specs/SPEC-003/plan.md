# SPEC-003 구현 계획: cc-manager 중심 Claude Code 최적화

## 구현 전략

### 핵심 전략: cc-manager First Approach

**cc-manager를 Claude Code 표준화의 중앙 관제탑으로 만들어 모든 커맨드/에이전트 생성과 표준화를 담당하도록 설계**

### 1단계: cc-manager 강화 (우선순위: Critical)

**목표**: cc-manager를 단순 설정 관리자에서 표준화 중앙 관제탑으로 전환

**구현 방법**:

1. **템플릿 지침 내장**
   - cc-manager.md에 커맨드/에이전트 표준 템플릿 추가
   - 파일 생성/수정 가이드라인 포함
   - Claude Code 공식 문서 준수 체크리스트

2. **검증 로직 통합**
   - YAML frontmatter 유효성 검사 함수
   - 필수 필드 존재 확인 로직
   - 표준 구조 검증 체크리스트

3. **파일 관리 기능**
   - 새로운 커맨드/에이전트 생성 기능
   - 기존 파일 표준화 업데이트 기능
   - 일괄 검증 및 수정 제안

### 2단계: 기존 파일 표준화 (우선순위: High)

**목표**: 모든 Claude Code 파일을 공식 구조에 맞게 변경

**구현 방법**:

1. **커맨드 파일 표준화** (.claude/commands/moai/)

   ```yaml
   # 표준 YAML frontmatter
   ---
   name: moai:command-name
   description: Clear one-line description
   argument-hint: [param1] [param2]
   allowed-tools: Tool1, Tool2, Task, Bash(cmd:*)
   model: sonnet
   ---
   ```

2. **에이전트 파일 표준화** (.claude/agents/moai/)

   ```yaml
   # 표준 YAML frontmatter
   ---
   name: agent-name
   description: Use PROACTIVELY for [specific task trigger]
   tools: Read, Write, Edit, MultiEdit, Bash, Glob, Grep
   model: sonnet
   ---
   ```

3. **프로액티브 트리거 명확화**
   - 각 에이전트의 자동 호출 조건 명시
   - "Use PROACTIVELY" 패턴 표준화
   - 에이전트 간 호출 규칙 정의

### 3단계: 핵심 문서 최적화 (우선순위: Medium)

**목표**: MoAI-ADK 핵심 문서들을 cc-manager 중심으로 재구성

**구현 방법**:

1. **CLAUDE.md 최적화**
   - cc-manager 역할 강조 섹션 추가
   - Claude Code 공식 문서 참조 링크
   - 워크플로우에서 cc-manager 중심성 명시

2. **development-guide.md 업데이트**
   - TRUST 원칙과 Claude Code 표준 통합
   - cc-manager 사용 가이드라인 추가
   - 표준 준수 체크리스트 포함

3. **settings.json 권한 최적화**
   ```json
   {
     "permissions": {
       "allow": [
         // 기존 도구 + 추가
         "WebSearch",
         "BashOutput",
         "KillShell",
         "Bash(gemini:*)",
         "Bash(codex:*)"
       ]
     }
   }
   ```

### 4단계: 검증 도구 개발 (우선순위: Medium)

**목표**: 표준 준수를 자동으로 확인하고 개선사항을 제안하는 도구

**구현 방법**:

1. **validate_claude_standards.py 개발**
   - YAML frontmatter 파싱 및 검증
   - 필수 필드 존재 확인
   - 표준 구조 준수 여부 체크

2. **통합 테스트 프레임워크**
   - 각 파일별 표준 준수 테스트
   - cc-manager 기능 테스트
   - 전체 워크플로우 통합 테스트

## 기술적 접근 방법

### TDD 구현 사이클

1. **RED Phase**: 실패하는 테스트 먼저 작성
   - cc-manager 템플릿 검증 테스트
   - 기존 파일 표준 준수 테스트
   - 검증 도구 동작 테스트

2. **GREEN Phase**: 최소한의 구현으로 테스트 통과
   - cc-manager.md 템플릿 지침 추가
   - 기존 파일 YAML frontmatter 수정
   - 검증 스크립트 기본 구현

3. **REFACTOR Phase**: 코드 품질 및 성능 최적화
   - 문서 가독성 개선
   - 검증 로직 최적화
   - 중복 코드 제거

### 아키텍처 설계 원칙

1. **단일 진실의 원천**: cc-manager가 모든 표준의 중심
2. **점진적 마이그레이션**: 기존 기능을 손상시키지 않고 단계적 개선
3. **실용적 접근**: 과도한 추상화 없이 실제 사용 패턴 기반
4. **표준 강제**: 자동화된 검증으로 표준 준수 보장

## 구현 마일스톤

### Week 1: cc-manager 강화

- [ ] cc-manager.md에 템플릿 지침 추가
- [ ] 표준 검증 로직 구현
- [ ] 기본 파일 생성/수정 기능 추가
- [ ] 단위 테스트 작성 및 통과

### Week 2: 파일 표준화

- [ ] 5개 커맨드 파일 YAML frontmatter 표준화
- [ ] 7개 에이전트 파일 YAML frontmatter 표준화
- [ ] 프로액티브 트리거 조건 명확화
- [ ] 통합 테스트 작성 및 통과

### Week 3: 문서 최적화

- [ ] CLAUDE.md cc-manager 섹션 강화
- [ ] development-guide.md 업데이트
- [ ] settings.json 권한 최적화
- [ ] 검증 스크립트 완성

### Week 4: 품질 보증

- [ ] 전체 시스템 통합 테스트
- [ ] 문서 가독성 최종 점검
- [ ] 성능 최적화 및 리팩토링
- [ ] 사용자 가이드 작성

## 리스크 및 대응 방안

### 주요 리스크

1. **기존 워크플로우 영향**
   - 리스크: 파일 구조 변경으로 기존 기능 손상
   - 대응: 점진적 마이그레이션, 하위 호환성 유지

2. **표준 준수 복잡성**
   - 리스크: Claude Code 공식 표준이 복잡하여 구현 어려움
   - 대응: 단계적 적용, 핵심 필드 우선 적용

3. **사용자 학습 곡선**
   - 리스크: cc-manager 중심 워크플로우 적응 시간 필요
   - 대응: 명확한 가이드 문서, 단계적 도입

### 대안 계획

- **Plan B**: 기존 구조 유지하며 새로운 표준 병행 적용
- **Plan C**: 핵심 에이전트만 우선 표준화, 점진적 확장

## 성공 지표

### 정량적 지표

- **표준 준수율**: 100% (모든 커맨드/에이전트 파일)
- **검증 통과율**: ≥95% (validate_claude_standards.py)
- **테스트 커버리지**: ≥85% (TDD 구현 영역)
- **성능 개선**: 파일 생성 시간 50% 단축

### 정성적 지표

- cc-manager 중심 워크플로우 확립
- Claude Code 공식 문서 100% 호환성
- 일관된 파일 구조 및 명명 규칙
- 개발자 경험 개선 (명확한 템플릿, 자동 검증)

## 다음 단계

### 즉시 시작 (이번 세션)

1. **TDD RED Phase**: 실패 테스트 작성
2. **cc-manager.md 템플릿 추가**: 내부 지침 통합
3. **기본 검증 스크립트 구현**: validate_claude_standards.py

### 다음 세션

1. **기존 파일 표준화**: 커맨드/에이전트 YAML frontmatter 수정
2. **핵심 문서 업데이트**: CLAUDE.md, development-guide.md
3. **통합 테스트**: 전체 워크플로우 검증

### 장기 계획

1. **커뮤니티 피드백**: 표준화 결과에 대한 사용자 의견 수렴
2. **지속적 개선**: Claude Code 업데이트에 따른 표준 조정
3. **확장성 검토**: 새로운 에이전트 추가 시 표준 적용

---

**연결된 SPEC**:

- SPEC-003 명세서: [spec.md](./spec.md)
- 수락 기준: [acceptance.md](./acceptance.md)

**관련 TAG**: @DESIGN:CC-MANAGER-ARCH-003, @TASK:IMPLEMENT-003
