---
id: CLAUDE-CODE-FEATURES-001
version: 0.0.1
status: active
created: 2025-11-02
updated: 2025-11-02
author: @GOOS
priority: high
category: feature
labels:
  - claude-code
  - optimization
  - cost-reduction
  - performance
---

## HISTORY

### v0.0.1 (2025-11-02)
- **INITIAL**: MoAI-ADK v0.9.0을 위해 Claude Code v2.0.30+ 신규 기능 3개 통합
- **AUTHOR**: @GOOS
- **SCOPE**: Haiku Auto SonnetPlan, Background Bash, Enhanced Grep
- **NOTE**: Feature 2 (PreCompact Hook), 5 (Plan Resume), 6 (TodoWrite Auto-Init) - Claude Code 내부 기능 또는 구현 불가로 제외

---

# SPEC: Claude Code v2.0.30+ Features Integration

## @SPEC:CLAUDE-CODE-FEATURES-001

### 1. Environment
- Claude Code v2.0.30 이상
- MoAI-ADK v0.12.0+ (Python 3.13+)
- .moai/config.json 설정 완료

### 2. Assumptions
- 모든 기능은 기존 Alfred 워크플로우와 호환
- Claude Code 최신 API가 안정적으로 지원됨

### 3. Feature Requirements (EARS Format)

#### 3.1 Feature 1: Haiku Auto SonnetPlan Mode

**Ubiquitous**: Plan 에이전트는 Sonnet 모델, 실행 에이전트는 Haiku 모델 사용

**Event-Driven**:
- WHEN spec-builder 호출 → THEN Sonnet 4.5 사용
- WHEN tdd-implementer 호출 → THEN Haiku 4.5 사용

**State-Driven**:
- WHILE 계획 모드 → Sonnet 사용
- WHILE 실행 모드 → Haiku 사용

**Unwanted Behaviors**:
- IF Haiku 실패 → THEN Sonnet으로 폴백

#### 3.2 Feature 3: Background Bash
- Ubiquitous: 테스트/빌드는 백그라운드 실행
- Event: run_in_background=true → 백그라운드 실행
- Performance: 40% 속도 향상

#### 3.3 Feature 4: Enhanced Grep
- Ubiquitous: multiline 패턴 매칭 지원
- Parameter: head_limit으로 결과 개수 제한
- Performance: 4-6배 빠른 검색

### Traceability (@TAG)
- @DOC:CLAUDE-CODE-FEATURES → .moai/memory/claude-code-features-guide.md
- @DOC:CLAUDE-CODE-FEATURES-1 → Agent model selection documentation
- @DOC:CLAUDE-CODE-FEATURES-3 → Background Bash usage guide
- @DOC:CLAUDE-CODE-FEATURES-4 → Enhanced Grep usage guide

### Success Criteria
- Model selection: Sonnet (계획), Haiku (실행) 명시적 선언 완료
- Background Bash: pytest 백그라운드 실행 가능 확인
- Enhanced Grep: multiline + head_limit 파라미터 사용 가능 확인
- Documentation: 사용 가이드 작성 완료
