# MoAI-ADK 스킬 호출 표준 가이드라인

## 🎯 기본 원칙

### 1. 명시적 스킬 호출 (Explicit Skill Invocation)

**항상 정확한 스킬 이름을 사용하세요:**

```python
# ✅ 올바른 방식
Skill("moai-alfred-ask-user-questions")
Skill("moai-foundation-tags")
Skill("moai-cc-agents")

# ❌ 잘못된 방식
skill-name  # 플레이스홀더만 사용
ask-user-questions  # moai- 접두사 없음
```

### 2. AskUserQuestion 도구 참조 표준

**일관된 참조 방식을 사용하세요:**

```python
# ✅ 표준 방식
AskUserQuestion 도구 (moai-alfred-ask-user-questions 스킬 참조)

# ❌ 사용 금지
AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)  # 너무 김
AskUserQuestion tool  # 스킬 참조 없음
```

### 3. 스킬 호출 문서화 패턴

**설명서에서의 일관된 형식:**

```markdown
## 필수 스킬

- `Skill("moai-alfred-ask-user-questions")` - 사용자 상호작용이 필요할 때
- `Skill("moai-foundation-tags")` - TAG 체인 검증 시
- `Skill("moai-cc-agents")` - 새 에이전트 생성 시

## 스킬 호출 규칙

1. 항상 `Skill("스킬이름")` 형식 사용
2. 스킬 이름은 정확하게 (moai- 접두사 포함)
3. 자동 로드는 명시적 호출로 대체
```

## 🔧 스킬 카테고리별 호출 가이드

### Alfred 코어 스킬
```python
Skill("moai-alfred-ask-user-questions")     # 사용자 질문/선택
Skill("moai-alfred-language-detection")     # 언어 감지
Skill("moai-alfred-workflow")              # 워크플로우 결정
Skill("moai-alfred-git-workflow")          # Git 전략
```

### Foundation 스킬
```python
Skill("moai-foundation-specs")              # SPEC 구조 검증
Skill("moai-foundation-tags")               # TAG 관리
Skill("moai-foundation-trust")              # TRUST 5 원칙
```

### CC (Claude Code) 스킬
```python
Skill("moai-cc-agents")                     # 에이전트 생성
Skill("moai-cc-commands")                   # 명령어 생성
Skill("moai-cc-skills")                     # 스킬 생성
Skill("moai-cc-settings")                   # 설정 관리
Skill("moai-cc-hooks")                      # Hook 설정
```

### Domain 스킬
```python
Skill("moai-domain-backend")                # 백엔드 전문
Skill("moai-domain-database")               # 데이터베이스
Skill("moai-domain-security")               # 보안
```

### Language 스킬
```python
Skill("moai-lang-python")                   # Python 패턴
Skill("moai-lang-typescript")               # TypeScript
Skill("moai-lang-go")                       # Go
```

### Essentials 스킬
```python
Skill("moai-essentials-debug")              # 디버깅
Skill("moai-essentials-perf")               # 성능
Skill("moai-essentials-refactor")           # 리팩토링
Skill("moai-essentials-security")           # 보안
```

## 📝 문서화 템플릿

### Agent/Command/Skill 파일 템플릿

```yaml
---
name: my-agent
description: "Use PROACTIVELY for [trigger conditions]"
tools: [tool list]
model: sonnet
---

# 에이전트 이름

## 스킬 활성화

**자동 로드**:
- `Skill("moai-foundation-specs")` - SPEC 구조 검증
- `Skill("moai-alfred-workflow")` - 워크플로우 결정

**조건부 로드**:
- `Skill("moai-alfred-language-detection")` - 언어 감지 필요 시
- `Skill("moai-alfred-ask-user-questions")` - 사용자 상호작용 필요 시

## 스킬 호출 규칙

1. 항상 `Skill("정확한-스킬이름")` 형식 사용
2. 스킬 이름은 moai- 접두사 포함
3. 플레이스홀더(`skill-name`)는 설명용으로만 사용

## 주요 스킬 매핑

| 작업 | 스킬 호출 | 설명 |
|------|-----------|------|
| 사용자 질문 | `Skill("moai-alfred-ask-user-questions")` | TUI 메뉴 제공 |
| TAG 검증 | `Skill("moai-foundation-tags")` | TAG 체인 분석 |
| 에이전트 생성 | `Skill("moai-cc-agents")` | 에이전트 템플릿 |
```

## 🚨 일반적인 실수와 수정

### 실수 1: 부정확한 AskUserQuestion 참조

```python
# ❌ 잘못된 방식
AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)

# ✅ 올바른 방식
AskUserQuestion 도구 (moai-alfred-ask-user-questions 스킬 참조)
```

### 실수 2: 플레이스홀더와 실제 이름 혼용

```python
# ❌ 잘못된 방식
Use Skill("skill-name") for loading skills

# ✅ 올바른 방식
Use Skill("moai-cc-agents") for agent creation
Use Skill("moai-foundation-tags") for TAG validation
```

### 실수 3: 불필이한 긴 설명

```python
# ❌ 너무 긴 방식
Interactive prompts use AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill) for TUI selection menus

# ✅ 간결한 방식
AskUserQuestion 도구 (moai-alfred-ask-user-questions 스킬 참조)로 TUI 메뉴 제공
```

## 🔍 검증 체크리스트

### 파일 검증 시 확인 항목

- [ ] 모든 `Skill("...")` 호출이 정확한 스킬 이름을 사용하는가?
- [ ] AskUserQuestion 참조가 표준 형식을 따르는가?
- [ ] 플레이스홀더(`skill-name`)가 실제 코드에 사용되지 않는가?
- [ ] 스킬 이름에 moai- 접두사가 포함되어 있는가?
- [ ] 문서화가 일관된 형식을 따르는가?

### 자동 검증 스크립트 (권장)

```bash
# .claude 디렉토리에서 스킬 호출 패턴 검증
grep -r "Skill(" .claude --include="*.md" | grep -v "moai-"  # 잘못된 스킬 호출 찾기
grep -r "AskUserQuestion tool" .claude --include="*.md"     # 비표준 참조 찾기
grep -r "skill-name" .claude --include="*.md" | grep -v "정확한 스킬 이름"  # 플레이스홀더 오용 찾기
```

## 📚 참고 자료

- [Claude Code Skills 문서](../skills/)
- [Agent 생성 가이드](../agents/)
- [Command 생성 가이드](../commands/)

---

**버전**: 1.0.0
**작성일**: 2025-11-05
**유지보수**: cc-manager 에이전트