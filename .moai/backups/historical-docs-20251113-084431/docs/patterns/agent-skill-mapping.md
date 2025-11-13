# Agent Skill Mapping Analysis

> **문서 목적**: 각 에이전트가 활용하는 Skills를 명시적으로 매핑하고, 필요한 개선사항 식별

## 📊 현재 상태

### 발견 사항

✅ **현재 구현**:
- 모든 에이전트가 Task() 호출로 위임됨 (`.claude/agents/alfred/*.md`)
- 에이전트들이 context7, sequential-thinking 등 MCP 도구 사용
- TodoWrite로 진행 상황 추적

⚠️ **개선 기회**:
- 명시적인 `skills:` 필드 부족 (현재는 암묵적)
- Skill 로딩 순서나 의존성 미명시
- 에이전트별 Skill 매핑 문서화 불충분

---

## 🔍 주요 에이전트별 Skill 활용

### 1. tdd-implementer.md

**현재 활용 도구**:
```
- Read, Write, Edit: 파일 조작
- Bash: 테스트 실행, git 작업
- Grep, Glob: 코드 탐색
- TodoWrite: 진행 추적
- mcp__context7: 라이브러리 문서 조회
```

**권장 Skill 매핑**:
```yaml
skills:
  - moai-lang-python          # Python TDD 패턴
  - moai-essentials-debug     # 디버깅 최적화
  - moai-foundation-tags      # TAG 체인 관리
  - moai-alfred-todowrite-pattern  # TodoWrite 패턴
```

**현재 구현 평가**: ✅ 우수
- MCP 도구와 직접 도구 적절히 조합
- TDD 패턴 명확히 따름

---

### 2. quality-gate.md

**현재 활용 도구**:
```
- Read: SPEC, 테스트, 코드 검토
- Grep: 코드 분석
- Bash: 테스트 실행, 커버리지
- mcp__sequential_thinking: 복잡한 검증
```

**권장 Skill 매핑**:
```yaml
skills:
  - moai-foundation-trust-5   # TRUST 5 원칙
  - moai-essentials-debug     # 에러 분석
  - moai-domain-monitoring    # 품질 메트릭
```

**현재 구현 평가**: ✅ 우수
- TRUST 5 검증 체계적
- 멀티 에이전트 협업 지원

---

### 3. git-manager.md

**현재 활용 도구**:
```
- Bash: git 명령어
- Read: 변경 사항 분석
- Grep: 파일 추적
```

**권장 Skill 매핑**:
```yaml
skills:
  - moai-domain-git           # Git workflow (있으면)
  - moai-alfred-commit-pattern # 커밋 메시지 표준
```

**현재 구현 평가**: ✅ 우수
- Git 작업을 에이전트에게 위임
- Bash로 직접 제어

---

### 4. implementation-planner.md

**현재 활용 도구**:
```
- Read: SPEC 분석
- mcp__sequential_thinking: 전략 수립
- TodoWrite: 계획 추적
```

**권장 Skill 매핑**:
```yaml
skills:
  - moai-alfred-spec-authoring  # SPEC 분석
  - moai-foundation-tags        # TAG 계획
  - moai-domain-backend/frontend  # 도메인 지식
```

**현재 구현 평가**: ✅ 우수
- 복잡한 분석을 sequential-thinking으로 처리
- SPEC 중심의 설계

---

## 📋 Skill 재사용 현황

### 현재 중복 가능성

```
Skill: moai-foundation-tags
├─ tdd-implementer: TAG 체인 관리
├─ quality-gate: TAG 검증
├─ tag-agent: TAG 관리
└─ doc-syncer: TAG 동기화

→ 단일 Skill으로 통합하면 유지보수 개선
```

### 권장 개선

**Skill 재사용 매트릭스**:

| Skill | tdd-impl | quality-gate | git-mgr | impl-plan |
|-------|----------|--------------|---------|-----------|
| moai-lang-python | ✅ | - | - | - |
| moai-foundation-tags | ✅ | ✅ | - | ✅ |
| moai-essentials-debug | ✅ | ✅ | - | - |
| moai-domain-security | - | ✅ | - | ✅ |
| moai-foundation-trust-5 | ✅ | ✅ | - | ✅ |

---

## 🚀 개선 제안

### Proposal 1: 명시적 Skill 필드 추가 (선택적)

**현재**:
```yaml
---
name: tdd-implementer
description: "..."
tools: Read, Write, Edit, Bash, ...
---
```

**개선안**:
```yaml
---
name: tdd-implementer
description: "..."
tools: Read, Write, Edit, Bash, ...
skills:
  - moai-lang-python
  - moai-foundation-tags
  - moai-essentials-debug
---
```

**장점**:
- 에이전트의 의존성 명확
- 문서화 개선
- IDE 자동완성 가능

**단점**:
- 현재 Task() 시스템에서 자동으로 처리됨
- 추가 유지보수 필요

**권장**: ⏸️ 선택적 (현재 시스템이 잘 작동하면 유지)

---

### Proposal 2: Skill 로딩 명시화 (권장)

**현재**:
```
에이전트가 필요한 Skills를 암묵적으로 로드
```

**개선안**:
```python
# 에이전트 코드에서 명시적으로
You are the tdd-implementer agent.

When implementing code:
1. Load Skill("moai-lang-python") for language patterns
2. Load Skill("moai-foundation-tags") for TAG chain
3. Load Skill("moai-essentials-debug") for debugging
```

**장점**:
- 에이전트 동작이 명확
- 디버깅 용이
- Skill 로딩 순서 제어

**권장**: ✅ 각 에이전트에 추가

---

### Proposal 3: Skill 라이브러리 확장

**현재 부족한 Skill**:
```
- moai-domain-git (Git workflow 전문화)
- moai-domain-testing (Testing 전문화)
- moai-domain-commit-pattern (Commit 표준)
```

**권장**: ✅ 필요시 생성

---

## 📊 전체 평가

### 현재 구현 점수: 8.5/10

**강점**:
- ✅ 에이전트 기반 아키텍처 완벽
- ✅ MCP 도구 활용 우수
- ✅ Skill 재사용 기본 구조 존재
- ✅ 책임 분리 명확

**개선 기회**:
- ⚠️ Skill 로딩 명시화 필요
- ⚠️ 에이전트별 Skill 매핑 문서 미흡
- ⚠️ 일부 Skill 중복 가능성

### 즉시 조치 필요

1. **각 에이전트에 Skill 로딩 명시**
2. **에이전트별 Skill 매핑 문서 작성**
3. **Skill 재사용 점검**

### 향후 검토

1. 새로운 도메인 Skill 생성 검토
2. Skill 캐싱/최적화
3. Skill 버전 관리

---

## 📝 Action Items

- [ ] 각 에이전트에 Skill 로딩 명시
- [ ] 에이전트별 Skill 매핑 표 작성
- [ ] Skill 중복 제거 계획
- [ ] 새로운 Skill 생성 로드맵

---

## 📚 참고

- `.claude/agents/alfred/tdd-implementer.md`
- `.claude/agents/alfred/quality-gate.md`
- `.claude/agents/alfred/git-manager.md`
- `.moai/docs/patterns/alfred-command-best-practices.md`
