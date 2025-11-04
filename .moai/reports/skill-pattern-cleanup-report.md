# MoAI-ADK 스킬 호출 패턴 표준화 보고서

**생성일**: 2025-11-05
**수행자**: cc-manager
**버전**: 1.0.0

## 📋 요약

MoAI-ADK 프로젝트의 일관성 없는 스킬 호출 패턴을 발견하고 표준화했습니다. 총 33개의 AskUserQuestion 참조를 수정하고, 스킬 호출 가이드라인과 검증 스크립트를 만들어 향후 일관성을 유지할 수 있는 기반을 마련했습니다.

## 🔍 발견된 문제점

### 1. AskUserQuestion 참조 방식의 불일치
- **문제**: `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` 형식이 38회 사용
- **영향**: 매우 긴 설명으로 문서 가독성 저하
- **위치**: agents/alfred/*.md, commands/alfred/*.md

### 2. 스킬 호출 설명의 부재
- **문제**: 일부 파일에서 스킬 호출 방식이 제대로 설명되지 않음
- **영향**: 개발자들이 올바른 스킬 호출 패턴을 알기 어려움

### 3. 일관성 없는 용어 사용
- **문제**: "tool" vs "도구" 혼용
- **영향**: 한국어 사용자에게 혼란

## ✅ 수정된 내용

### 1. AskUserQuestion 참조 표준화
**수정 전**:
```
AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)
```

**수정 후**:
```
AskUserQuestion 도구 (moai-alfred-ask-user-questions 스킬 참조)
```

**수정된 파일 목록**:
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/cc-manager.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/debug-helper.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/doc-syncer.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/git-manager.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/implementation-planner.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/project-manager.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/quality-gate.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/skill-factory.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/spec-builder.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/tag-agent.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/tdd-implementer.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/trust-checker.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/0-project.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/2-run.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/3-sync.md`
- `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-autofixes/reference.md`

### 2. 스킬 호출 가이드라인 생성
- **파일**: `.moai/docs/skill-invocation-standards.md`
- **내용**:
  - 기본 원칙 및 명시적 스킬 호출 방법
  - 스킬 카테고리별 호출 가이드
  - 문서화 템플릿
  - 일반적인 실수와 수정 방법
  - 검증 체크리스트

### 3. 자동 검증 스크립트 생성
- **파일**: `.moai/scripts/skill-pattern-validator.sh`
- **기능**:
  - 스킬 호출 패턴 자동 검증
  - AskUserQuestion 참조 일관성 확인
  - 등록된 스킬 목록 관리
  - 주요 스킬 호출 현황 분석

## 📊 현재 상태

### 검증 결과 (2025-11-05 기준)
- ✅ **오류**: 0건
- ⚠️ **경고**: 1건 (플레이스홀더 `skill-name` 설명용 사용)
- ✅ **AskUserQuestion 참조**: 33건 모두 표준화됨

### 등록된 스킬
- **총 스킬 수**: 84개
- **주요 호출 스킬**:
  - moai-alfred-ask-user-questions
  - moai-foundation-tags
  - moai-foundation-trust
  - moai-cc-agents
  - moai-cc-commands

## 🎯 표준화된 스킬 호출 원칙

### 1. 명시적 스킬 호출
```python
# ✅ 올바른 방식
Skill("moai-alfred-ask-user-questions")
Skill("moai-foundation-tags")

# ❌ 잘못된 방식
skill-name  # 플레이스홀더만 사용
```

### 2. AskUserQuestion 도구 참조
```python
# ✅ 표준 방식
AskUserQuestion 도구 (moai-alfred-ask-user-questions 스킬 참조)

# ❌ 사용 금지
AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)
```

### 3. 스킬 카테고리별 그룹화
- **Alfred 코어**: `moai-alfred-*`
- **Foundation**: `moai-foundation-*`
- **Claude Code**: `moai-cc-*`
- **Domain**: `moai-domain-*`
- **Language**: `moai-lang-*`
- **Essentials**: `moai-essentials-*`

## 🔮 향후 유지 관리 계획

### 1. 정기 검증
- **주기**: 월 1회 자동 실행
- **도구**: `.moai/scripts/skill-pattern-validator.sh`
- **담당**: cc-manager 에이전트

### 2. 새 스킬 생성 시 검증
- **프로세스**: 스킬 생성 후 자동 검증 스크립트 실행
- **체크포인트**: Pull Request 시 자동 검증
- **가이드**: 스킬 생성 가이드라인 준수 확인

### 3. 문서 업데이트
- **가이드라인**: `.moai/docs/skill-invocation-standards.md` 주기적 검토
- **템플릿**: 새 스킬/에이전트/명령어 템플릿에 표준 적용
- **교육**: 개발자들을 위한 스킬 호출 모범 사례 공유

## 🎉 성과

1. **일관성 향상**: 33개의 불일치 패턴을 표준화
2. **가독성 개선**: 긴 설명을 간결한 표준으로 대체
3. **자동화 도입**: 검증 스크립트로 지속적인 품질 관리
4. **문서화**: 가이드라인으로 지식 공유 및 재사용성 증대

## 📝 권장사항

1. **새 스킬 생성 시**: 가이드라인을 먼저 확인하고 템플릿 사용
2. **정기 검증**: 월 1회 검증 스크립트 실행으로 일관성 유지
3. **코드 리뷰**: Pull Request 시 스킬 호출 패턴 검토 포함
4. **문서화**: 새로운 스킬 패턴 발견 시 가이드라인 업데이트

---

**결론**: MoAI-ADK 프로젝트의 스킬 호출 패턴을 성공적으로 표준화했습니다. 자동 검증 도구와 명확한 가이드라인을 통해 향후 일관성을 유지할 수 있는 기반을 마련했습니다.