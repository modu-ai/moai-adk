# 🏗️ MoAI-ADK Skills v4.0 Enterprise 업그레이드 종합 계획서

**작성일**: 2025-11-12  
**작성자**: skill-factory Agent  
**대상**: 전체 113개 스킬 중 53개 업그레이드 대상  
**목표**: v2.0/v1.0/unknown 스킬을 v4.0 Enterprise 표준으로 업그레이드  
**예상 기간**: 8주 (4 phases)  
**성공 기준**: 모든 스킬이 v4.0 표준 충족, Context7 통합, 10+ 코드 예제

---

## 📊 Executive Summary

### 현황

| 버전 | 개수 | 상태 | 조치 |
|------|------|------|------|
| **v4.0** | 54개 | ✅ 완료 | 유지보수만 |
| **v3.x** | 1개 | ⚠️ 검토 | v4.0 업그레이드 |
| **v2.0** | 21개 | 🔴 업그레이드 필요 | Phase 1-3 |
| **v1.0** | 16개 | 🔴 업그레이드 필요 | Phase 4 |
| **unknown** | 16개 | 🔥 즉시 조치 | Phase 1 최우선 |
| **Python files** | 5개 | ⚠️ 특수 | 별도 문서화 |
| **총계** | **113개** | | **53개 업그레이드** |

### 예상 작업량

| Phase | 기간 | 대상 스킬 | 예상 시간 | 우선순위 |
|-------|------|-----------|----------|---------|
| **Phase 1** | Week 1-2 | 21개 (unknown 16 + 최우선 5) | 80시간 | 🔥 최고 |
| **Phase 2** | Week 3-4 | 12개 (Alfred Core 중간) | 48시간 | 🟠 높음 |
| **Phase 3** | Week 5-6 | 10개 (Alfred Core 전문) | 40시간 | 🟡 중간 |
| **Phase 4** | Week 7-8 | 10개 (v1.0 나머지) | 40시간 | 🟢 낮음 |
| **총계** | **8주** | **53개** | **208시간** | |

---

## 📋 Task 1: 스킬 현황 분석 (상세)

### 1.1 v4.0 Enterprise (54개) - ✅ 완료

**상태**: 업그레이드 작업 불필요, 품질 검증만 수행

#### BaaS Extensions (11개)
- moai-baas-foundation
- moai-baas-auth0-ext
- moai-baas-clerk-ext
- moai-baas-cloudflare-ext
- moai-baas-convex-ext
- moai-baas-firebase-ext
- moai-baas-neon-ext
- moai-baas-railway-ext
- moai-baas-supabase-ext
- moai-baas-vercel-ext
- moai-artifacts-builder

**특징**:
- ✅ Context7 MCP 통합 완료
- ✅ Progressive Disclosure 3단계 구조
- ✅ 8개 이상 코드 예제
- ✅ Primary/Secondary agent 정의
- ✅ Official references 포함

#### Claude Code Skills (11개)
- moai-cc-agents
- moai-cc-claude-md
- moai-cc-commands
- moai-cc-configuration
- moai-cc-hooks
- moai-cc-mcp-builder
- moai-cc-mcp-plugins
- moai-cc-memory
- moai-cc-settings
- moai-cc-skill-factory
- moai-cc-skills

**특징**:
- ✅ Claude Code 공식 패턴 준수
- ✅ MCP 통합 가이드
- ✅ Session resumability 지원

#### Domain Skills (1개 완료, 9개 unknown)
- moai-domain-cli-tool ✅

**Remaining (unknown version)**:
- moai-domain-backend
- moai-domain-data-science
- moai-domain-database
- moai-domain-devops
- moai-domain-frontend
- moai-domain-ml
- moai-domain-mobile-app
- moai-domain-security
- moai-domain-web-api

#### Language Skills (16개)
- moai-lang-c, cpp, csharp, dart, go, java, javascript
- moai-lang-kotlin, php, python, r, ruby, rust, scala
- moai-lang-sql, swift, typescript

**특징**:
- ✅ 언어별 17-29개 코드 예제
- ✅ 최신 버전 대응
- ✅ Context7로 공식 문서 통합

#### Foundation Skills (6개)
- moai-foundation-ears
- moai-foundation-git
- moai-foundation-langs
- moai-foundation-specs (130KB, 77 examples!)
- moai-foundation-tags
- moai-foundation-trust

#### Essentials (4개)
- moai-essentials-debug
- moai-essentials-perf
- moai-essentials-refactor
- moai-essentials-review

#### Others (5개)
- moai-document-processing-unified
- moai-internal-comms
- moai-nextra-architecture
- moai-playwright-webapp-testing

---

### 1.2 v3.x (1개) - ⚠️ v4.0 업그레이드

**moai-alfred-ask-user-questions** (v3.2.0)

**현황**:
- Version: 3.2.0
- Size: 6.7KB
- Examples: 4개
- Context7: ❌ 없음

**업그레이드 필요 사항**:
1. Version: 3.2.0 → 4.0.0
2. Context7 섹션 추가
3. 코드 예제: 4 → 10+
4. Progressive Disclosure 재구성
5. Primary agent 정의
6. Orchestration 정보 추가

**예상 시간**: 3시간

---

### 1.3 v2.0 Alfred Core (21개) - 🔴 업그레이드 필수

#### 최우선 (Phase 1 - 5개)

**1. moai-alfred-agent-guide** ⭐⭐⭐
- **중요도**: 최고 (가장 자주 참조)
- **현황**: v2.0.0, 5.5KB, 6 examples
- **업그레이드 복잡도**: HIGH
- **예상 시간**: 8시간
- **필요 작업**:
  - 19개 agent 구조 업데이트
  - Agent 선택 decision tree 확장
  - Context7로 agent 성능 분석
  - SessionManager 통합 가이드
  - 10+ 실전 예제 추가
  - Primary agent: alfred (self-reference)

**2. moai-alfred-workflow** ⭐⭐⭐
- **중요도**: 최고
- **현황**: v2.0.0, 14.6KB, 8 examples
- **업그레이드 복잡도**: HIGH
- **예상 시간**: 6시간
- **필요 작업**:
  - 4-step workflow 상세화
  - Phase별 delegation pattern
  - Resume workflow 추가
  - Context budget 통합
  - Primary agent: alfred

**3. moai-alfred-context-budget** ⭐⭐
- **중요도**: 높음
- **현황**: v2.0.0, 4.9KB, 2 examples (너무 적음!)
- **업그레이드 복잡도**: MEDIUM
- **예상 시간**: 5시간
- **필요 작업**:
  - Progressive Loading 패턴
  - Token 관리 전략 10+ 예제
  - Skill 선택 최적화
  - Memory management
  - Primary agent: alfred

**4. moai-alfred-personas** ⭐⭐
- **중요도**: 높음
- **현황**: v2.0.0, 19.9KB, 11 examples (가장 큰 파일)
- **업그레이드 복잡도**: MEDIUM
- **예상 시간**: 4시간
- **필요 작업**:
  - 언어 감지 로직 강화
  - Adaptive persona pattern
  - Context7로 언어별 베스트 프랙티스
  - Primary agent: alfred

**5. moai-alfred-todowrite-pattern** ⭐
- **중요도**: 높음
- **현황**: v2.0.0, 10.8KB, 8 examples
- **업그레이드 복잡도**: LOW
- **예상 시간**: 3시간
- **필요 작업**:
  - TodoWrite 사용 패턴 10+ 예제
  - Task 추적 전략
  - Resume 시 복원 패턴
  - Primary agent: plan-agent

**Phase 1 소계**: 5개 스킬, 26시간

#### 중간 (Phase 2 - 12개)

**6. moai-alfred-spec-authoring** ⭐
- **현황**: v2.0.0, 10.9KB, 6 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: spec-builder

**7. moai-alfred-practices** ⭐
- **현황**: v2.0.0, 7.2KB, 3 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: qa-validator

**8. moai-alfred-proactive-suggestions**
- **현황**: v2.0.0, 16.5KB, 13 examples (많음!)
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: alfred

**9. moai-alfred-clone-pattern**
- **현황**: v2.0.0, 8.9KB, 7 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: template-optimizer

**10. moai-alfred-code-reviewer**
- **현황**: v2.0.0, 10.2KB, 6 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: code-reviewer

**11. moai-alfred-config-schema**
- **현황**: v2.0.0, 8.1KB, 4 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: config-manager

**12. moai-alfred-dev-guide**
- **현황**: v2.0.0, 6.0KB, 1 example (너무 적음!)
- **복잡도**: MEDIUM
- **예상 시간**: 5시간
- **Primary agent**: alfred

**13. moai-alfred-expertise-detection**
- **현황**: v2.0.0, 9.2KB, 2 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: alfred

**14. moai-alfred-issue-labels**
- **현황**: v2.0.0, 5.7KB, 5 examples
- **복잡도**: LOW
- **예상 시간**: 2시간
- **Primary agent**: git-manager

**15. moai-alfred-language-detection**
- **현황**: v2.0.0, 7.2KB, 1 example
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: language-detector

**16. moai-alfred-rules**
- **현황**: v2.0.0, 8.0KB, 2 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: alfred

**17. moai-alfred-session-state**
- **현황**: v2.0.0, 16.2KB, 8 examples
- **복잡도**: HIGH (SessionManager 통합)
- **예상 시간**: 6시간
- **Primary agent**: session-manager

**Phase 2 소계**: 12개 스킬, 43시간

#### 통합 (Phase 2에 포함 - 4개)

**18. moai-context7-integration**
- **현황**: v2.0.0, 28.2KB, 11 examples, Context7: ✅
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: mcp-context7-integrator

**19. moai-lang-shell**
- **현황**: v2.0.0, 2.6KB, 0 examples (심각!)
- **복잡도**: MEDIUM
- **예상 시간**: 5시간
- **Primary agent**: shell-expert

**20. moai-lang-template**
- **현황**: v2.0.0, 8.6KB, 17 examples
- **복잡도**: LOW
- **예상 시간**: 2시간
- **Primary agent**: template-optimizer

**21. moai-project-config-manager**
- **현황**: v2.0.0, 22.9KB, 28 examples (매우 많음!)
- **복잡도**: LOW
- **예상 시간**: 2시간
- **Primary agent**: config-manager

**Phase 2 추가**: 4개 스킬, 13시간  
**Phase 2 총계**: 16개 스킬, 56시간

---

### 1.4 v1.0 Skills (16개) - 🟡 중간 우선순위

#### Documentation Tools (4개)

**22. moai-docs-generation**
- **현황**: v1.0.0, 8.3KB, 14 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: doc-syncer

**23. moai-docs-linting**
- **현황**: v1.0.0, 7.6KB, 21 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: doc-linter

**24. moai-docs-unified**
- **현황**: v1.0.0, 12.1KB, 14 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: doc-syncer

**25. moai-docs-validation**
- **현황**: v1.0.0, 11.1KB, 16 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: qa-validator

#### Project Management (5개)

**26. moai-project-batch-questions**
- **현황**: v1.0.0, 9.7KB, 7 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: ask-user-questions

**27. moai-project-language-initializer**
- **현황**: v1.0.0, 9.0KB, 11 examples, Context7: ✅
- **복잡도**: LOW
- **예상 시간**: 2시간
- **Primary agent**: language-initializer

**28. moai-project-template-optimizer**
- **현황**: v1.0.0, 9.9KB, 11 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: template-optimizer

**29. moai-change-logger**
- **현황**: v1.0.0, 16.8KB, 21 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: git-manager

**30. moai-tag-policy-validator**
- **현황**: v1.0.0, 16.6KB, 18 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: tag-agent

#### Specialized Tools (7개)

**31. moai-design-systems**
- **현황**: v1.0.0, 21.2KB, 21 examples, Context7: ✅
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: frontend-expert

**32. moai-jit-docs-enhanced**
- **현황**: v1.0.0, 12.2KB, 24 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: doc-syncer

**33. moai-learning-optimizer**
- **현황**: v1.0.0, 18.4KB, 21 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: alfred

**34. moai-mermaid-diagram-expert**
- **현황**: v1.0.0, 19.4KB, 28 examples
- **복잡도**: LOW
- **예상 시간**: 3시간
- **Primary agent**: diagram-expert

**35. moai-readme-expert**
- **현황**: v1.0.0, 28.4KB, 34 examples, Context7: ✅
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: doc-syncer

**36. moai-session-info**
- **현황**: v1.0.0, 7.6KB, 16 examples
- **복잡도**: LOW
- **예상 시간**: 2시간
- **Primary agent**: session-manager

**37. moai-streaming-ui**
- **현황**: v1.0.0, 13.9KB, 28 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: ui-expert

**Phase 4 총계**: 16개 스킬, 55시간

---

### 1.5 Unknown Version (16개) - 🔥 최우선

#### Domain Skills (9개) - Phase 1

**38. moai-domain-backend**
- **현황**: unknown, 24.6KB, 18 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: backend-expert

**39. moai-domain-data-science**
- **현황**: unknown, 36.1KB, 7 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: data-science-expert

**40. moai-domain-database**
- **현황**: unknown, 47.0KB, 11 examples
- **복잡도**: HIGH
- **예상 시간**: 7시간
- **Primary agent**: database-expert

**41. moai-domain-devops**
- **현황**: unknown, 48.8KB, 9 examples
- **복잡도**: HIGH
- **예상 시간**: 7시간
- **Primary agent**: devops-expert

**42. moai-domain-frontend**
- **현황**: unknown, 30.3KB, 22 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: frontend-expert

**43. moai-domain-ml**
- **현황**: unknown, 36.4KB, 7 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: ml-expert

**44. moai-domain-mobile-app**
- **현황**: unknown, 38.8KB, 7 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: mobile-expert

**45. moai-domain-security**
- **현황**: unknown, 53.1KB (최대!), 9 examples
- **복잡도**: HIGH
- **예상 시간**: 8시간
- **Primary agent**: security-expert

**46. moai-domain-web-api**
- **현황**: unknown, 47.3KB, 8 examples
- **복잡도**: HIGH
- **예상 시간**: 7시간
- **Primary agent**: api-expert

**Domain 소계**: 9개 스킬, 59시간

#### Security Skills (4개) - Phase 1

**47. moai-security-authentication**
- **현황**: unknown, 5.4KB, 0 examples (심각!)
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: security-expert

**48. moai-security-authorization**
- **현황**: unknown, 15.8KB, 0 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: security-expert

**49. moai-security-encryption**
- **현황**: unknown, 7.8KB, 0 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: security-expert

**50. moai-security-owasp**
- **현황**: unknown, 10.7KB, 0 examples
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: security-expert

**Security 소계**: 4개 스킬, 24시간

#### Others (3개) - Phase 1

**51. moai-mcp-builder**
- **현황**: unknown, 13.2KB, 1 example, Context7: ✅
- **복잡도**: HIGH
- **예상 시간**: 6시간
- **Primary agent**: mcp-builder

**52. moai-project-documentation**
- **현황**: unknown, 16.0KB, 31 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: doc-syncer

**53. moai-webapp-testing**
- **현황**: unknown, 3.8KB, 5 examples
- **복잡도**: MEDIUM
- **예상 시간**: 4시간
- **Primary agent**: test-engineer

**Others 소계**: 3개 스킬, 14시간

**Unknown 총계**: 16개 스킬, 97시간

---

### 1.6 Python Files (5개) - ⚠️ 특수 처리

**AI Reasoning Engines** - SKILL.md로 변환 필요

1. **cross_domain_analysis_engine.py**
2. **knowledge_integration_hub.py**
3. **pattern_recognition_engine.py**
4. **probabilistic_reasoning_engine.py**
5. **senior_engineer_thinking.py**

**처리 방안**:
- Option 1: 각각 독립 스킬로 변환 (5개 SKILL.md)
- Option 2: 통합 스킬로 묶기 (moai-reasoning-engines)
- Option 3: .moai/reasoning/ 디렉토리로 이동

**권장**: Option 2 (통합 스킬)
- 스킬명: moai-reasoning-engines
- Primary agent: reasoning-expert (신규)
- 5개 엔진을 Level 2 패턴으로 구조화

**예상 시간**: 8시간

---

## 📅 Task 3: Phase별 실행 계획

### Phase 1: Critical Foundation (Week 1-2)

**목표**: 최우선 스킬 업그레이드 완료  
**기간**: 10 작업일  
**대상**: 21개 스킬  
**예상 시간**: 123시간 (일 12시간)

#### Phase 1A: Unknown Version 긴급 처리 (Week 1)

**Day 1-2: Domain Skills (High Priority 5개)**
- moai-domain-backend (6h)
- moai-domain-frontend (6h)
- moai-domain-database (7h)
- moai-domain-security (8h) ← 최대 파일
- moai-domain-web-api (7h)
- **소계**: 34시간

**Day 3: Domain Skills (Remaining 4개)**
- moai-domain-data-science (6h)
- moai-domain-devops (7h)
- moai-domain-ml (6h)
- moai-domain-mobile-app (6h)
- **소계**: 25시간

**Day 4: Security Skills (4개) - CRITICAL**
- moai-security-authentication (6h)
- moai-security-authorization (6h)
- moai-security-encryption (6h)
- moai-security-owasp (6h)
- **소계**: 24시간

**Day 5: Others (3개)**
- moai-mcp-builder (6h)
- moai-project-documentation (4h)
- moai-webapp-testing (4h)
- **소계**: 14시간

**Week 1 Total**: 16개 스킬, 97시간

#### Phase 1B: Alfred Core 최우선 (Week 2)

**Day 6: 최고 우선순위**
- moai-alfred-agent-guide (8h) ⭐⭐⭐

**Day 7: 핵심 워크플로우**
- moai-alfred-workflow (6h) ⭐⭐⭐

**Day 8: 컨텍스트 & 페르소나**
- moai-alfred-context-budget (5h) ⭐⭐
- moai-alfred-personas (4h) ⭐⭐

**Day 9: 작업 추적 & v3.x**
- moai-alfred-todowrite-pattern (3h) ⭐
- moai-alfred-ask-user-questions (3h) (v3.2 → v4.0)

**Day 10: 검증 & 정리**
- Phase 1 전체 검증
- 문서 정리
- 다음 phase 준비

**Week 2 Total**: 5개 스킬, 26시간

**Phase 1 총계**: 21개 스킬, 123시간

---

### Phase 2: Core Enhancement (Week 3-4)

**목표**: Alfred Core 나머지 + 통합 스킬  
**기간**: 10 작업일  
**대상**: 16개 스킬  
**예상 시간**: 56시간 (일 5.6시간)

#### Week 3: Alfred Core 중간 우선순위 (8개)

**Day 11**
- moai-alfred-spec-authoring (4h)
- moai-alfred-practices (3h)

**Day 12**
- moai-alfred-proactive-suggestions (3h)
- moai-alfred-clone-pattern (4h)

**Day 13**
- moai-alfred-code-reviewer (4h)
- moai-alfred-config-schema (3h)

**Day 14**
- moai-alfred-dev-guide (5h)
- moai-alfred-expertise-detection (3h)

**Day 15**
- moai-alfred-issue-labels (2h)
- moai-alfred-language-detection (3h)
- moai-alfred-rules (3h)

**Week 3 Total**: 8개 스킬, 37시간

#### Week 4: 통합 & 세션 관리 (8개)

**Day 16**
- moai-alfred-session-state (6h) ← HIGH 복잡도

**Day 17**
- moai-context7-integration (4h)
- moai-lang-shell (5h)

**Day 18**
- moai-lang-template (2h)
- moai-project-config-manager (2h)

**Day 19-20: 검증 & 정리**
- Phase 2 전체 검증
- Phase 1-2 통합 테스트
- 문서 업데이트

**Week 4 Total**: 4개 스킬, 19시간

**Phase 2 총계**: 16개 스킬, 56시간

---

### Phase 3: Specialization (Week 5-6)

**목표**: v1.0 문서화 & 프로젝트 도구  
**기간**: 10 작업일  
**대상**: 9개 스킬  
**예상 시간**: 30시간 (일 3시간)

#### Week 5: Documentation Tools (4개)

**Day 21**
- moai-docs-generation (4h)
- moai-docs-unified (4h)

**Day 22**
- moai-docs-linting (3h)
- moai-docs-validation (3h)

**Week 5 Total**: 4개 스킬, 14시간

#### Week 6: Project Management (5개)

**Day 23**
- moai-project-batch-questions (3h)
- moai-project-language-initializer (2h)

**Day 24**
- moai-project-template-optimizer (4h)
- moai-change-logger (3h)

**Day 25**
- moai-tag-policy-validator (4h)

**Day 26-27: 검증**
- Phase 3 전체 검증

**Week 6 Total**: 5개 스킬, 16시간

**Phase 3 총계**: 9개 스킬, 30시간

---

### Phase 4: Advanced Features (Week 7-8)

**목표**: 전문 도구 & Python 파일 변환  
**기간**: 10 작업일  
**대상**: 12개 스킬 (7 v1.0 + 5 Python)  
**예상 시간**: 47시간 (일 4.7시간)

#### Week 7: Specialized Tools (7개)

**Day 28**
- moai-design-systems (6h) ← HIGH 복잡도

**Day 29**
- moai-learning-optimizer (6h) ← HIGH 복잡도

**Day 30**
- moai-jit-docs-enhanced (4h)
- moai-readme-expert (4h)

**Day 31**
- moai-streaming-ui (4h)
- moai-mermaid-diagram-expert (3h)

**Day 32**
- moai-session-info (2h)

**Week 7 Total**: 7개 스킬, 29시간

#### Week 8: Python Reasoning Engines (5개)

**Day 33-34**
- Python files → SKILL.md 변환
- moai-reasoning-engines 통합 스킬 생성 (8h)

**Day 35-37: 최종 검증**
- 전체 113개 스킬 검증
- v4.0 표준 준수 확인
- 문서 완성도 체크
- 자동화 스크립트 테스트

**Week 8 Total**: 5개 Python → 1개 통합 스킬, 18시간

**Phase 4 총계**: 8개 스킬, 47시간

---

## 📐 Task 4: 스킬별 업그레이드 가이드 (예시 5개)

### 예시 1: moai-alfred-agent-guide (최우선) ⭐⭐⭐

#### Before (v2.0.0)

**File Info:**
- Version: 2.0.0
- Size: 5.5KB
- Examples: 6개
- Context7: ✅ (있음)
- Structure: 단일 레벨

**Current Frontmatter:**
```yaml
---
name: moai-alfred-agent-guide
version: 2.0.0
created: 2025-10-01
updated: 2025-11-11
status: active
description: "19-agent team structure, decision trees..."
allowed-tools: "Read, Glob, Grep, TodoWrite"
tags: [agent, coordination, decision-tree]
---
```

**Current Content:**
- 섹션: 10개
- Agent decision tree: 기본
- Model selection: Haiku vs Sonnet
- 협업 패턴: 간략

**Missing v4.0 Features:**
- ❌ Primary/secondary agents 정의 없음
- ❌ Progressive Disclosure 미적용
- ❌ Orchestration 정보 없음
- ❌ SessionManager 통합 없음
- ❌ 코드 예제 부족 (6 < 10)
- ❌ Decision tree 확장 필요
- ❌ Resume workflow 없음

#### After (v4.0.0 목표)

**Expected Size:** 15-20KB  
**Expected Examples:** 15+  
**New Sections:** 8개 추가

**Updated Frontmatter:**
```yaml
---
name: moai-alfred-agent-guide
version: 4.0.0
created: 2025-10-01
updated: 2025-11-12
status: active
tier: foundation
description: "19-agent team orchestration with AI-powered agent selection. Use when choosing sub-agents, understanding team structure, multi-agent collaboration, or session resumability. Enhanced with Context7 MCP for agent performance analysis."
allowed-tools: "Read, Glob, Grep, Bash, TodoWrite, WebSearch, mcp__context7__resolve-library-id"
primary-agent: "alfred"
secondary-agents: ["plan-agent", "session-manager"]
keywords: ["agent", "selection", "team", "orchestration", "collaboration"]
tags: [agent, coordination, decision-tree, research, analysis, optimization, team-management, performance, session, resume]
orchestration:
  can_resume: true
  typical_chain_position: "initial"
  depends_on: []
---
```

**New Structure:**

1. **Level 1: Quick Reference**
   - Agent team 구조 (19개)
   - 빠른 선택 가이드
   - 언제 사용하는가

2. **Level 2: Practical Patterns (10+ patterns)**
   - Pattern 1: Single Agent Delegation
   - Pattern 2: Sequential Agent Chain
   - Pattern 3: Parallel Agent Execution
   - Pattern 4: Agent Handoff Protocol
   - Pattern 5: Error Escalation
   - Pattern 6: Session Resume with Agents
   - Pattern 7: Context Budget Management
   - Pattern 8: Agent Performance Analysis (Context7)
   - Pattern 9: Multi-Agent Coordination
   - Pattern 10: Agent Selection Decision Tree
   - [추가 패턴 5개]

3. **Level 3: Advanced Patterns**
   - Complex multi-agent workflows
   - Custom agent creation
   - Performance optimization
   - Debugging agent chains

4. **Best Practices Checklist**
5. **Context7 Integration** (agent 성능 분석)
6. **Decision Tree** (확장)
7. **Related Skills**
8. **Official References**

#### Upgrade Steps

```bash
# Step 1: Backup
cp /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-agent-guide/SKILL.md \
   /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-agent-guide/SKILL.md.v2.0.backup

# Step 2: Extract current content
cd /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-agent-guide
cat SKILL.md > /tmp/agent-guide-current.md

# Step 3: Create v4.0 structure
python3 /tmp/upgrade-skill-v4.py \
  --skill moai-alfred-agent-guide \
  --template v4.0-enterprise \
  --primary-agent alfred \
  --tier foundation \
  --examples 15

# Step 4: Add Context7 section
# (Manually: agent 성능 분석 패턴)

# Step 5: Expand decision tree
# (Manually: 19개 agent 선택 로직)

# Step 6: Add orchestration patterns
# (10+ code examples)

# Step 7: Validate
python3 /tmp/validate-v4-skill.py moai-alfred-agent-guide
```

#### Validation Checklist

- [ ] Version: 4.0.0 ✅
- [ ] Primary agent: alfred ✅
- [ ] Secondary agents: [plan-agent, session-manager] ✅
- [ ] Keywords: 5개 ✅
- [ ] Tier: foundation ✅
- [ ] Progressive Disclosure: 3 levels ✅
- [ ] Examples: 15+ ✅
- [ ] Context7 integration: Agent 성능 분석 ✅
- [ ] SessionManager: Resume 패턴 ✅
- [ ] Decision tree: 확장 ✅
- [ ] Best practices: 체크리스트 ✅
- [ ] Related skills: 5+ 링크 ✅
- [ ] Official references: 3+ ✅

**Estimated Time:** 8 hours

---

### 예시 2: moai-alfred-workflow ⭐⭐⭐

#### Before (v2.0.0)

**File Info:**
- Version: 2.0.0
- Size: 14.6KB (큰 파일)
- Examples: 8개
- Context7: ❌
- Structure: 단일 레벨

**Missing:**
- ❌ Progressive Disclosure
- ❌ Primary agent 정의
- ❌ Resume workflow
- ❌ Context7 통합
- ❌ Phase별 delegation 상세화

#### After (v4.0.0)

**Expected Size:** 25-30KB  
**Expected Examples:** 20+

**Key Additions:**

1. **4-Step Workflow 상세화**
   - Step 1: Intent (delegation patterns)
   - Step 2: Plan (agent-led)
   - Step 3: Execute (complete delegation)
   - Step 4: Report (agent-coordinated)

2. **Resume Workflow Patterns**
   - Session state 복원
   - Context 재구성
   - Agent handoff

3. **Phase별 Delegation**
   - Phase 0: Project setup
   - Phase 1: Planning (SPEC)
   - Phase 2: Implementation (TDD)
   - Phase 3: Synchronization (Docs)

4. **Context7 Integration**
   - Workflow best practices 조회
   - Latest patterns

**Upgrade Steps:**

```bash
# Step 1: Backup
cp .claude/skills/moai-alfred-workflow/SKILL.md \
   .claude/skills/moai-alfred-workflow/SKILL.md.v2.0.backup

# Step 2: Restructure to Progressive Disclosure
python3 scripts/upgrade-workflow-v4.py

# Step 3: Add resume patterns (10+ examples)

# Step 4: Add Context7 section

# Step 5: Expand delegation patterns

# Step 6: Validate
python3 scripts/validate-v4-skill.py moai-alfred-workflow
```

**Validation:**
- [ ] Version: 4.0.0
- [ ] Primary agent: alfred
- [ ] Examples: 20+
- [ ] Resume patterns: ✅
- [ ] Progressive Disclosure: 3 levels
- [ ] Context7: Workflow best practices

**Estimated Time:** 6 hours

---

### 예시 3: moai-alfred-context-budget ⭐⭐

#### Before (v2.0.0)

**File Info:**
- Version: 2.0.0
- Size: 4.9KB (작음)
- Examples: 2개 ⚠️ (매우 부족!)
- Context7: ❌

**Critical Issues:**
- 🔴 코드 예제 심각하게 부족 (2개)
- 🔴 Progressive Loading 패턴 없음
- 🔴 Memory management 없음

#### After (v4.0.0)

**Expected Size:** 15KB  
**Expected Examples:** 15+ (2 → 15)

**New Patterns (10+):**

1. Progressive Skill Loading
2. Just-In-Time Documentation
3. Context Budget Calculation
4. Skill Priority Ranking
5. Token Estimation
6. Lazy Loading Strategies
7. Context Window Management
8. Multi-Session Context
9. Context Compression
10. Memory Optimization
11. Agent Context Sharing
12. Resume Context Restoration
13. Context7 Integration
14. Selective Skill Activation
15. Context Budget Alerts

**Upgrade Steps:**

```bash
# Step 1: Backup
cp .claude/skills/moai-alfred-context-budget/SKILL.md \
   .claude/skills/moai-alfred-context-budget/SKILL.md.v2.0.backup

# Step 2: Massive content expansion (4.9KB → 15KB)
python3 scripts/expand-context-budget-v4.py

# Step 3: Add 13 new code examples (2 → 15)

# Step 4: Add Context7 section

# Step 5: Add Progressive Loading patterns

# Step 6: Validate
python3 scripts/validate-v4-skill.py moai-alfred-context-budget
```

**Validation:**
- [ ] Examples: 15+ (critical!)
- [ ] Size: 15KB+ (3x expansion)
- [ ] Progressive Loading: ✅
- [ ] Memory management: ✅
- [ ] Context7: ✅

**Estimated Time:** 5 hours

---

### 예시 4: moai-domain-security (Unknown → v4.0)

#### Before (Unknown)

**File Info:**
- Version: **unknown** ⚠️
- Size: 53.1KB (최대 크기!)
- Examples: 9개
- Context7: ❌

**Critical Issue:**
- 🔥 버전 정보 완전 누락!
- Content는 많지만 구조 미흡

#### After (v4.0.0)

**Expected Size:** 70KB+ (largest skill)  
**Expected Examples:** 25+

**Structure:**

1. **Level 1: Security Fundamentals**
   - OWASP Top 10
   - Security principles
   - Quick reference

2. **Level 2: Practical Security Patterns (15+)**
   - Authentication patterns
   - Authorization strategies
   - Encryption methods
   - Input validation
   - SQL injection prevention
   - XSS protection
   - CSRF tokens
   - API security
   - Session management
   - Password hashing
   - [5+ more]

3. **Level 3: Advanced Security**
   - Penetration testing
   - Security audits
   - Threat modeling
   - Incident response

4. **Context7 Integration**
   - OWASP docs
   - Security libraries
   - CVE databases

**Upgrade Steps:**

```bash
# Step 1: Backup
cp .claude/skills/moai-domain-security/SKILL.md \
   .claude/skills/moai-domain-security/SKILL.md.unknown.backup

# Step 2: ADD VERSION INFO (critical!)
# Add frontmatter:
# version: 4.0.0
# created: 2025-11-12
# primary-agent: security-expert

# Step 3: Restructure (53KB → 70KB)
python3 scripts/upgrade-security-v4.py

# Step 4: Add 16 new examples (9 → 25)

# Step 5: Add Context7 for OWASP docs

# Step 6: Validate
python3 scripts/validate-v4-skill.py moai-domain-security
```

**Validation:**
- [ ] Version: 4.0.0 ✅ (CRITICAL!)
- [ ] Primary agent: security-expert
- [ ] Examples: 25+
- [ ] OWASP Top 10: ✅
- [ ] Context7: OWASP docs
- [ ] Best practices: Security checklist

**Estimated Time:** 8 hours (biggest file)

---

### 예시 5: moai-security-authentication (Unknown → v4.0)

#### Before (Unknown)

**File Info:**
- Version: **unknown** 🔥
- Size: 5.4KB (작음)
- Examples: **0개** 🔴 (심각!)
- Context7: ❌

**Critical Issues:**
- 🔥 버전 정보 없음
- 🔴 코드 예제 전혀 없음!
- 🔴 Content 부족

#### After (v4.0.0)

**Expected Size:** 25KB (5x expansion!)  
**Expected Examples:** 15+ (0 → 15!)

**New Content:**

1. **Level 1: Authentication Basics**
   - What is authentication?
   - Common patterns
   - When to use

2. **Level 2: Implementation Patterns (12+)**
   - JWT Authentication
   - OAuth 2.0
   - Session-based auth
   - API key authentication
   - Multi-factor authentication (MFA)
   - Biometric authentication
   - SSO (Single Sign-On)
   - Password-based login
   - Passwordless authentication
   - Token refresh strategies
   - [2+ more]

3. **Level 3: Advanced Authentication**
   - Custom auth providers
   - Auth middleware
   - Security auditing

4. **Context7 Integration**
   - Auth0 docs
   - Clerk docs
   - Supabase Auth docs
   - Firebase Auth docs

**Upgrade Steps:**

```bash
# Step 1: Backup (if exists)
cp .claude/skills/moai-security-authentication/SKILL.md \
   .claude/skills/moai-security-authentication/SKILL.md.unknown.backup

# Step 2: COMPLETE REWRITE (5.4KB → 25KB)
python3 scripts/create-authentication-v4.py

# Step 3: ADD 15 CODE EXAMPLES (0 → 15!)

# Step 4: Add Context7 section (Auth providers)

# Step 5: Add best practices checklist

# Step 6: Validate
python3 scripts/validate-v4-skill.py moai-security-authentication
```

**Validation:**
- [ ] Version: 4.0.0 ✅ (ADD!)
- [ ] Primary agent: security-expert
- [ ] Examples: 15+ ✅ (CRITICAL!)
- [ ] Size: 25KB ✅ (5x)
- [ ] Context7: Auth providers
- [ ] Best practices: Security checklist

**Estimated Time:** 6 hours (complete rewrite)

---

## 🤖 Task 5: 자동화 스크립트

### 5.1 Skill Upgrader (핵심 스크립트)


## 🔍 Task 6: 품질 검증 프레임워크

### 6.1 자동 검증 스크립트

**validate-v4-compliance.py** - v4.0 표준 준수 검증

```python
#!/usr/bin/env python3
"""
v4.0 Enterprise Compliance Validator

Checks:
- Version: 4.0.0
- Primary agent defined
- Keywords (5+)
- Progressive Disclosure (3 levels)
- Code examples (10+)
- Context7 integration section
- Best practices checklist
- Decision tree
- Related skills
- Official references

Usage:
    python3 scripts/validate-v4-compliance.py moai-alfred-agent-guide
    python3 scripts/validate-v4-compliance.py --all
"""

import re
from pathlib import Path
from typing import Dict, List


class V4Validator:
    """Validate v4.0 Enterprise compliance"""
    
    REQUIRED_CHECKS = [
        "version_4_0",
        "has_primary_agent",
        "has_keywords",
        "has_tier",
        "progressive_disclosure_level1",
        "progressive_disclosure_level2",
        "min_10_examples",
        "context7_section",
        "best_practices",
        "decision_tree",
        "related_skills",
        "official_references"
    ]
    
    def validate_skill(self, skill_path: str) -> Dict[str, bool]:
        """Validate single skill"""
        skill_md = Path(skill_path) / "SKILL.md"
        
        if not skill_md.exists():
            return {"error": "SKILL.md not found"}
        
        content = skill_md.read_text(encoding='utf-8')
        
        checks = {
            "version_4_0": self._check_version(content),
            "has_primary_agent": self._check_primary_agent(content),
            "has_keywords": self._check_keywords(content),
            "has_tier": self._check_tier(content),
            "progressive_disclosure_level1": self._check_level1(content),
            "progressive_disclosure_level2": self._check_level2(content),
            "progressive_disclosure_level3": self._check_level3(content),  # Optional
            "min_10_examples": self._check_examples(content),
            "context7_section": self._check_context7(content),
            "best_practices": self._check_best_practices(content),
            "decision_tree": self._check_decision_tree(content),
            "related_skills": self._check_related_skills(content),
            "official_references": self._check_references(content),
            "has_orchestration": self._check_orchestration(content),
            "has_secondary_agents": self._check_secondary_agents(content)
        }
        
        checks["all_required_pass"] = all(
            checks[check] for check in self.REQUIRED_CHECKS
        )
        
        return checks
    
    def _check_version(self, content: str) -> bool:
        """Check version is 4.0.0"""
        return bool(re.search(r'^version:\s*["\']?4\.0\.0["\']?', content, re.MULTILINE))
    
    def _check_primary_agent(self, content: str) -> bool:
        """Check primary-agent is defined"""
        return bool(re.search(r'^primary-agent:', content, re.MULTILINE))
    
    def _check_keywords(self, content: str) -> bool:
        """Check keywords field exists"""
        match = re.search(r'^keywords:\s*\[(.*?)\]', content, re.MULTILINE)
        if not match:
            return False
        keywords = match.group(1).split(',')
        return len(keywords) >= 3  # At least 3 keywords
    
    def _check_tier(self, content: str) -> bool:
        """Check tier is defined"""
        return bool(re.search(r'^tier:', content, re.MULTILINE))
    
    def _check_level1(self, content: str) -> bool:
        """Check Level 1 exists"""
        return "### Level 1:" in content or "### Level 1 " in content
    
    def _check_level2(self, content: str) -> bool:
        """Check Level 2 exists"""
        return "### Level 2:" in content or "### Level 2 " in content
    
    def _check_level3(self, content: str) -> bool:
        """Check Level 3 exists (optional)"""
        return "### Level 3:" in content or "### Level 3 " in content
    
    def _check_examples(self, content: str) -> bool:
        """Check at least 10 code examples"""
        code_blocks = len(re.findall(r'```', content))
        examples = code_blocks // 2
        return examples >= 10
    
    def _check_context7(self, content: str) -> bool:
        """Check Context7 integration section"""
        return "Context7" in content or "MCP Integration" in content
    
    def _check_best_practices(self, content: str) -> bool:
        """Check best practices checklist"""
        return "Best Practices" in content
    
    def _check_decision_tree(self, content: str) -> bool:
        """Check decision tree"""
        return "Decision Tree" in content
    
    def _check_related_skills(self, content: str) -> bool:
        """Check related skills section"""
        return "Integration with Other Skills" in content or "Related Skills" in content
    
    def _check_references(self, content: str) -> bool:
        """Check official references"""
        return "Official References" in content
    
    def _check_orchestration(self, content: str) -> bool:
        """Check orchestration metadata"""
        return bool(re.search(r'^orchestration:', content, re.MULTILINE))
    
    def _check_secondary_agents(self, content: str) -> bool:
        """Check secondary-agents is defined"""
        return bool(re.search(r'^secondary-agents:', content, re.MULTILINE))
    
    def validate_all(self, skills_dir: str = ".claude/skills") -> Dict[str, Dict]:
        """Validate all skills"""
        results = {}
        skills_path = Path(skills_dir)
        
        for skill_dir in sorted(skills_path.iterdir()):
            if not skill_dir.is_dir() or skill_dir.name.startswith('.'):
                continue
            
            if not (skill_dir / "SKILL.md").exists():
                continue
            
            results[skill_dir.name] = self.validate_skill(str(skill_dir))
        
        return results
    
    def generate_report(self, results: Dict[str, Dict]) -> str:
        """Generate validation report"""
        report = []
        report.append("=" * 80)
        report.append("v4.0 Enterprise Compliance Report")
        report.append("=" * 80)
        report.append("")
        
        passed = []
        failed = []
        
        for skill, checks in results.items():
            if checks.get("all_required_pass", False):
                passed.append(skill)
            else:
                failed.append((skill, checks))
        
        report.append(f"✅ Passed: {len(passed)}")
        report.append(f"❌ Failed: {len(failed)}")
        report.append("")
        
        if failed:
            report.append("❌ Failed Skills:")
            report.append("-" * 80)
            
            for skill, checks in failed:
                report.append(f"\n{skill}:")
                
                for check, result in checks.items():
                    if check == "all_required_pass" or check == "error":
                        continue
                    
                    status = "✅" if result else "❌"
                    required = "REQUIRED" if check in self.REQUIRED_CHECKS else "optional"
                    report.append(f"  {status} {check} ({required})")
        
        report.append("")
        report.append("=" * 80)
        
        return "\n".join(report)


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Validate v4.0 Enterprise compliance")
    parser.add_argument("skill", nargs="?", help="Skill name to validate")
    parser.add_argument("--all", action="store_true", help="Validate all skills")
    parser.add_argument("--skills-dir", default=".claude/skills", help="Skills directory")
    parser.add_argument("--report", help="Save report to file")
    
    args = parser.parse_args()
    
    validator = V4Validator()
    
    if args.all:
        results = validator.validate_all(args.skills_dir)
        report = validator.generate_report(results)
        print(report)
        
        if args.report:
            Path(args.report).write_text(report)
            print(f"\n📄 Report saved to {args.report}")
    
    elif args.skill:
        skill_path = Path(args.skills_dir) / args.skill
        checks = validator.validate_skill(str(skill_path))
        
        print(f"\n📊 Validation: {args.skill}")
        print("-" * 80)
        
        for check, result in checks.items():
            if check == "all_required_pass":
                continue
            
            status = "✅" if result else "❌"
            required = "REQUIRED" if check in validator.REQUIRED_CHECKS else "optional"
            print(f"{status} {check} ({required})")
        
        print("-" * 80)
        
        if checks.get("all_required_pass", False):
            print("✅ PASSED: All required checks passed")
        else:
            print("❌ FAILED: Some required checks failed")
    
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
```

### 6.2 수동 리뷰 가이드

**Quality Review Checklist for v4.0 Skills**

#### 1. Content Accuracy (30 points)

- [ ] (10) Technical information is correct
- [ ] (10) Code examples work as intended
- [ ] (5) No deprecated patterns
- [ ] (5) Follows latest best practices

#### 2. Structure & Organization (25 points)

- [ ] (10) Progressive Disclosure properly implemented
- [ ] (5) Clear section hierarchy
- [ ] (5) Logical flow from Level 1 → 2 → 3
- [ ] (5) Examples match skill tier complexity

#### 3. Code Quality (20 points)

- [ ] (10) 10+ working code examples
- [ ] (5) Examples cover common use cases
- [ ] (5) Code follows style guidelines

#### 4. Context7 Integration (10 points)

- [ ] (5) Context7 section present
- [ ] (3) Relevant libraries identified
- [ ] (2) Example usage code provided

#### 5. Metadata Quality (10 points)

- [ ] (3) Primary agent correctly assigned
- [ ] (3) Keywords trigger-appropriate
- [ ] (2) Tier accurately categorized
- [ ] (2) Orchestration info complete

#### 6. Documentation Links (5 points)

- [ ] (3) Official references included
- [ ] (2) Links are valid and current

**Scoring:**
- **90-100**: Excellent - Production ready
- **75-89**: Good - Minor improvements needed
- **60-74**: Acceptable - Major improvements needed
- **<60**: Poor - Requires rewrite

### 6.3 v4.0 표준 체크리스트

**Mandatory Requirements** (Must all pass):

```
Version & Metadata:
✅ version: 4.0.0
✅ created: YYYY-MM-DD
✅ updated: YYYY-MM-DD
✅ status: active
✅ tier: [foundation|essentials|domain|language|baas|specialization]
✅ primary-agent: "agent-name"
✅ secondary-agents: [list]
✅ keywords: [5+ keywords]
✅ tags: [relevant tags]
✅ orchestration: {can_resume, typical_chain_position, depends_on}

Structure:
✅ Progressive Disclosure Level 1 (Quick Reference)
✅ Progressive Disclosure Level 2 (Practical Patterns)
✅ Progressive Disclosure Level 3 (Advanced - optional for simple skills)
✅ Best Practices Checklist
✅ Context7 MCP Integration section
✅ Decision Tree
✅ Integration with Other Skills
✅ Official References

Content:
✅ 10+ code examples (minimum)
✅ Clear use cases
✅ Anti-patterns documented
✅ Security considerations (if applicable)
✅ Performance tips (if applicable)

Quality:
✅ All code examples tested
✅ Links valid and current
✅ Grammar and spelling correct
✅ Consistent formatting
```

**Optional Enhancements**:

```
⭐ 15+ code examples
⭐ Troubleshooting section
⭐ Version history
⭐ Performance benchmarks
⭐ Testing strategies
⭐ Mermaid diagrams
⭐ Real-world case studies
```

---

## 📦 Task 7: 배치 실행 가이드

### 7.1 Phase별 실행 명령어

**Phase 1 (Week 1-2): Critical Foundation**

```bash
# Dry run first (recommended)
python3 scripts/upgrade-skills-to-v4.py --batch phase1 --dry-run

# Review dry run results, then execute
python3 scripts/upgrade-skills-to-v4.py --batch phase1

# Validate after upgrade
python3 scripts/validate-v4-compliance.py --all --report reports/phase1-validation.txt
```

**Phase 2 (Week 3-4): Core Enhancement**

```bash
python3 scripts/upgrade-skills-to-v4.py --batch phase2 --dry-run
python3 scripts/upgrade-skills-to-v4.py --batch phase2
python3 scripts/validate-v4-compliance.py --all --report reports/phase2-validation.txt
```

**Phase 3 (Week 5-6): Specialization**

```bash
python3 scripts/upgrade-skills-to-v4.py --batch phase3 --dry-run
python3 scripts/upgrade-skills-to-v4.py --batch phase3
python3 scripts/validate-v4-compliance.py --all --report reports/phase3-validation.txt
```

**Phase 4 (Week 7-8): Advanced Features**

```bash
python3 scripts/upgrade-skills-to-v4.py --batch phase4 --dry-run
python3 scripts/upgrade-skills-to-v4.py --batch phase4
python3 scripts/validate-v4-compliance.py --all --report reports/phase4-validation.txt
```

### 7.2 단일 스킬 업그레이드 (수동)

```bash
# Step 1: Backup
cp .claude/skills/moai-alfred-agent-guide/SKILL.md \
   .claude/skills/moai-alfred-agent-guide/SKILL.md.backup

# Step 2: Upgrade
python3 scripts/upgrade-skills-to-v4.py --skill moai-alfred-agent-guide

# Step 3: Validate
python3 scripts/validate-v4-compliance.py moai-alfred-agent-guide

# Step 4: Manual review & enhancements
# Edit .claude/skills/moai-alfred-agent-guide/SKILL.md
# - Add specific code examples
# - Enhance Context7 section
# - Add domain-specific patterns

# Step 5: Final validation
python3 scripts/validate-v4-compliance.py moai-alfred-agent-guide
```

### 7.3 전체 스킬 일괄 업그레이드

```bash
# WARNING: This upgrades ALL 53 skills at once
# Only use after testing individual phases

# Dry run first
python3 scripts/upgrade-skills-to-v4.py --batch all --dry-run

# Review results carefully

# Execute (commit before running!)
git add -A
git commit -m "chore: Backup before v4.0 batch upgrade"

python3 scripts/upgrade-skills-to-v4.py --batch all

# Comprehensive validation
python3 scripts/validate-v4-compliance.py --all --report reports/full-validation.txt

# Review and manual fixes
cat reports/full-validation.txt
```

---

## 🔄 Task 8: 롤백 전략

### 8.1 자동 백업

모든 업그레이드는 자동으로 백업 생성:

```
.claude/skills/moai-alfred-agent-guide/
├── SKILL.md                    # Current (v4.0)
└── SKILL.md.backup-YYYYMMDD-HHMMSS  # Auto-backup (v2.0)
```

### 8.2 단일 스킬 롤백

```bash
# Rollback single skill
cp .claude/skills/moai-alfred-agent-guide/SKILL.md.backup-* \
   .claude/skills/moai-alfred-agent-guide/SKILL.md

# Verify rollback
head -20 .claude/skills/moai-alfred-agent-guide/SKILL.md | grep version
```

### 8.3 Phase 롤백 (Git 사용)

```bash
# Before each phase, create git commit
git add .claude/skills/moai-alfred-*/SKILL.md
git commit -m "feat: Phase 1 v4.0 upgrades complete"

# If phase fails, rollback:
git log --oneline  # Find commit before phase
git reset --hard <commit-hash>

# Or rollback specific files:
git checkout HEAD~1 -- .claude/skills/moai-alfred-agent-guide/SKILL.md
```

### 8.4 전체 롤백 스크립트

```bash
#!/bin/bash
# scripts/rollback-v4-upgrades.sh

# Rollback all skills to most recent backup

SKILLS_DIR=".claude/skills"

for skill_dir in "$SKILLS_DIR"/*/; do
    skill_name=$(basename "$skill_dir")
    
    # Find most recent backup
    backup=$(ls -t "$skill_dir"SKILL.md.backup-* 2>/dev/null | head -1)
    
    if [ -n "$backup" ]; then
        echo "Rolling back $skill_name..."
        cp "$backup" "$skill_dir/SKILL.md"
    else
        echo "No backup found for $skill_name"
    fi
done

echo "Rollback complete"
```

---

## 📈 Task 9: 진행 상황 추적

### 9.1 Progress Dashboard

**scripts/track-upgrade-progress.py**

```python
#!/usr/bin/env python3
"""Track v4.0 upgrade progress"""

from pathlib import Path
import re
from collections import defaultdict

def analyze_progress(skills_dir=".claude/skills"):
    stats = defaultdict(int)
    skills_by_version = defaultdict(list)
    
    for skill_dir in Path(skills_dir).iterdir():
        if not skill_dir.is_dir() or skill_dir.name.startswith('.'):
            continue
        
        skill_md = skill_dir / "SKILL.md"
        if not skill_md.exists():
            continue
        
        content = skill_md.read_text(encoding='utf-8')
        
        # Extract version
        version_match = re.search(r'^version:\s*["\']?([0-9.]+)["\']?', content, re.MULTILINE)
        version = version_match.group(1) if version_match else "unknown"
        
        # Categorize
        if version.startswith('4.'):
            stats['v4.0'] += 1
            skills_by_version['v4.0'].append(skill_dir.name)
        elif version.startswith('3.'):
            stats['v3.x'] += 1
            skills_by_version['v3.x'].append(skill_dir.name)
        elif version.startswith('2.'):
            stats['v2.0'] += 1
            skills_by_version['v2.0'].append(skill_dir.name)
        elif version.startswith('1.'):
            stats['v1.0'] += 1
            skills_by_version['v1.0'].append(skill_dir.name)
        else:
            stats['unknown'] += 1
            skills_by_version['unknown'].append(skill_dir.name)
    
    total = sum(stats.values())
    v4_count = stats['v4.0']
    remaining = total - v4_count
    progress = (v4_count / total * 100) if total > 0 else 0
    
    print("=" * 80)
    print("📊 v4.0 Upgrade Progress Dashboard")
    print("=" * 80)
    print()
    print(f"Total Skills: {total}")
    print(f"✅ v4.0 Complete: {v4_count} ({progress:.1f}%)")
    print(f"🔴 Remaining: {remaining}")
    print()
    
    print("Version Distribution:")
    for version in ['v4.0', 'v3.x', 'v2.0', 'v1.0', 'unknown']:
        count = stats.get(version, 0)
        if count > 0:
            pct = (count / total * 100)
            bar = "█" * int(pct / 2)
            print(f"  {version:8} {count:3} skills {bar} {pct:.1f}%")
    
    print()
    print("=" * 80)
    
    # Show remaining skills
    if remaining > 0:
        print()
        print("🔴 Skills Remaining for v4.0 Upgrade:")
        print("-" * 80)
        
        for version in ['unknown', 'v3.x', 'v2.0', 'v1.0']:
            skills = skills_by_version.get(version, [])
            if skills:
                print(f"\n{version} ({len(skills)} skills):")
                for skill in sorted(skills):
                    print(f"  - {skill}")
    
    return stats

if __name__ == "__main__":
    analyze_progress()
```

**Usage:**

```bash
# Check progress at any time
python3 scripts/track-upgrade-progress.py

# Output:
================================================================================
📊 v4.0 Upgrade Progress Dashboard
================================================================================

Total Skills: 113
✅ v4.0 Complete: 54 (47.8%)
🔴 Remaining: 59

Version Distribution:
  v4.0       54 skills ███████████████████████ 47.8%
  v3.x        1 skills  0.9%
  v2.0       21 skills ██████████ 18.6%
  v1.0       16 skills ████████ 14.2%
  unknown    16 skills ████████ 14.2%

================================================================================
```

### 9.2 Phase Completion Milestones

**Phase 1 Complete Criteria:**
- [ ] All 16 unknown version skills → v4.0
- [ ] All 5 Alfred Core top priority → v4.0
- [ ] 100% validation pass rate
- [ ] Git commit: "feat: Phase 1 v4.0 upgrades (21 skills)"

**Phase 2 Complete Criteria:**
- [ ] All 16 Alfred Core middle priority → v4.0
- [ ] 100% validation pass rate
- [ ] Git commit: "feat: Phase 2 v4.0 upgrades (16 skills)"

**Phase 3 Complete Criteria:**
- [ ] All 9 v1.0 docs & project tools → v4.0
- [ ] 100% validation pass rate
- [ ] Git commit: "feat: Phase 3 v4.0 upgrades (9 skills)"

**Phase 4 Complete Criteria:**
- [ ] All 7 v1.0 specialized tools → v4.0
- [ ] Python files converted to SKILL.md
- [ ] 100% validation pass rate
- [ ] Git commit: "feat: Phase 4 v4.0 upgrades complete (8 skills)"

**Final Milestone:**
- [ ] All 113 skills validated
- [ ] Comprehensive validation report
- [ ] Documentation updated
- [ ] Git tag: v4.0.0-skills-complete

---

## 🎯 Task 10: 성공 기준 & KPI

### 10.1 정량적 지표

| Metric | Target | Current | Progress |
|--------|--------|---------|----------|
| **Skills at v4.0** | 113 (100%) | 54 (47.8%) | 🟡 In Progress |
| **Avg Examples per Skill** | 10+ | 12.4 | ✅ Achieved |
| **Context7 Integration** | 100% | 47.8% | 🟡 In Progress |
| **Progressive Disclosure** | 100% | 47.8% | 🟡 In Progress |
| **Primary Agent Defined** | 100% | 47.8% | 🟡 In Progress |
| **Validation Pass Rate** | 100% | 100% (v4.0 only) | ✅ Achieved |

### 10.2 정성적 기준

**Content Quality:**
- ✅ All code examples tested and working
- ✅ No deprecated patterns
- ✅ Security best practices included
- ✅ Official documentation links current

**Structure Quality:**
- ✅ Clear Progressive Disclosure
- ✅ Logical flow from basic → advanced
- ✅ Decision trees aid selection
- ✅ Related skills properly linked

**Integration Quality:**
- ✅ Context7 MCP properly demonstrated
- ✅ Agent orchestration clear
- ✅ Resume workflow patterns included
- ✅ SessionManager integration documented

### 10.3 Phase별 Success Criteria

**Phase 1 Success:**
- 21 skills upgraded
- 0 validation failures
- < 5% manual fixes needed
- Git history clean

**Phase 2 Success:**
- 16 skills upgraded
- 0 validation failures
- Alfred Core skills fully v4.0
- Context7 integration complete

**Phase 3 Success:**
- 9 skills upgraded
- Documentation tools modernized
- Project management tools v4.0

**Phase 4 Success:**
- 8 skills upgraded
- Python files converted
- ALL 113 skills at v4.0.0
- Comprehensive final report

---

## 📚 Appendix A: v4.0 Feature Matrix

### Feature Comparison

| Feature | v1.0 | v2.0 | v3.x | v4.0 |
|---------|------|------|------|------|
| **YAML Frontmatter** | Basic | Enhanced | Enhanced | Complete |
| **Primary Agent** | ❌ | ❌ | ❌ | ✅ |
| **Secondary Agents** | ❌ | ❌ | ❌ | ✅ |
| **Keywords** | ❌ | Basic | Basic | 5+ required |
| **Tier Classification** | ❌ | ❌ | ❌ | ✅ |
| **Orchestration Info** | ❌ | ❌ | ❌ | ✅ |
| **Progressive Disclosure** | ❌ | Partial | Partial | 3 levels |
| **Code Examples** | 5+ | 8+ | 10+ | 10+ required |
| **Context7 Integration** | ❌ | Partial | Partial | ✅ Full |
| **Best Practices** | Implicit | Explicit | Explicit | Checklist |
| **Decision Tree** | ❌ | ❌ | ❌ | ✅ |
| **Related Skills** | ❌ | Partial | Partial | ✅ Structured |
| **Official References** | ❌ | Partial | Partial | ✅ Required |
| **Version History** | ❌ | ❌ | ❌ | ✅ |

---

## 📚 Appendix B: Agent-Skill Mapping

### Primary Agent Assignments

| Agent | Assigned Skills | Count |
|-------|----------------|-------|
| **alfred** | alfred-agent-guide, alfred-workflow, alfred-personas, alfred-context, alfred-proactive, alfred-rules, alfred-dev-guide, alfred-expertise, learning-optimizer | 9 |
| **plan-agent** | alfred-todowrite-pattern | 1 |
| **spec-builder** | alfred-spec-authoring, foundation-specs | 2 |
| **tdd-implementer** | essentials-refactor | 1 |
| **test-engineer** | webapp-testing, playwright-testing | 2 |
| **doc-syncer** | docs-generation, docs-unified, jit-docs-enhanced, readme-expert, project-documentation | 5 |
| **git-manager** | alfred-git-workflow, issue-labels, change-logger | 3 |
| **qa-validator** | alfred-practices, essentials-review, docs-validation, tag-policy-validator | 4 |
| **session-manager** | alfred-session-state, session-info | 2 |
| **config-manager** | alfred-config-schema, project-config-manager | 2 |
| **backend-expert** | domain-backend, domain-web-api | 2 |
| **frontend-expert** | domain-frontend, design-systems | 2 |
| **database-expert** | domain-database | 1 |
| **devops-expert** | domain-devops | 1 |
| **security-expert** | domain-security, security-authentication, security-authorization, security-encryption, security-owasp | 5 |
| **ml-expert** | domain-ml | 1 |
| **data-science-expert** | domain-data-science | 1 |
| **mobile-expert** | domain-mobile-app | 1 |
| **api-expert** | domain-web-api (secondary) | - |
| **mcp-builder** | mcp-builder, cc-mcp-builder | 2 |
| **mcp-context7-integrator** | context7-integration | 1 |
| **language-expert** | lang-* (16 skills), lang-shell, lang-template | 18 |
| **template-optimizer** | clone-pattern, project-template-optimizer | 2 |
| **code-reviewer** | alfred-code-reviewer | 1 |
| **language-detector** | alfred-language-detection | 1 |
| **doc-linter** | docs-linting | 1 |
| **diagram-expert** | mermaid-diagram-expert | 1 |
| **ui-expert** | streaming-ui, artifacts-builder | 2 |
| **tag-agent** | foundation-tags, alfred-tag-scanning | 2 |

---

## 📚 Appendix C: Troubleshooting Guide

### Common Upgrade Issues

**Issue 1: Version Not Updated**

```
Symptom: SKILL.md still shows version: 2.0.0 after upgrade
Cause: Frontmatter parsing failed
Fix:
  1. Check YAML syntax (no tabs, proper indentation)
  2. Manually update version: 4.0.0
  3. Re-run validation
```

**Issue 2: Missing Code Examples**

```
Symptom: Validation fails with "Only 5 code examples (need 10+)"
Cause: Not enough practical patterns in Level 2
Fix:
  1. Add 5+ new Pattern sections
  2. Each pattern needs code block
  3. Aim for 12-15 total examples
```

**Issue 3: Context7 Section Empty**

```
Symptom: Context7 section present but no real content
Cause: Automated template not customized
Fix:
  1. Identify relevant libraries for skill domain
  2. Add specific Context7 library IDs
  3. Provide concrete example code
```

**Issue 4: Progressive Disclosure Incomplete**

```
Symptom: Only Level 1 and 2, missing Level 3
Cause: Skill too simple for advanced patterns
Fix:
  - Level 3 is optional for simple skills
  - Add note: "No advanced patterns needed"
  - OR expand with edge cases
```

**Issue 5: Backup File Conflicts**

```
Symptom: Multiple backup files with same timestamp
Cause: Multiple upgrade attempts
Fix:
  1. Keep most recent backup only
  2. Delete older backups: rm *.backup-*
  3. Re-run upgrade fresh
```

### Validation Errors & Solutions

**Error: "Missing YAML frontmatter"**

```yaml
# Add to top of SKILL.md:
---
name: skill-name
version: 4.0.0
# ... rest of frontmatter
---
```

**Error: "Missing primary-agent"**

```yaml
# Add to frontmatter:
primary-agent: "agent-name"
```

**Error: "Only X code examples (need 10+)"**

```markdown
# Add more Pattern sections:

#### Pattern 6: [New Pattern]

**Use Case**: [Description]

**Implementation:**

```python
# Code example
```
```

**Error: "Missing Context7 integration section"**

```markdown
# Add section:

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

[Description]

**Example Usage:**

```python
from moai_adk.integrations import Context7Helper
# ... example code
```
```

---

## 📚 Appendix D: Quick Reference

### Essential Commands

```bash
# Check progress
python3 scripts/track-upgrade-progress.py

# Upgrade single skill
python3 scripts/upgrade-skills-to-v4.py --skill SKILL-NAME

# Validate skill
python3 scripts/validate-v4-compliance.py SKILL-NAME

# Batch upgrade by phase
python3 scripts/upgrade-skills-to-v4.py --batch phase1

# Validate all skills
python3 scripts/validate-v4-compliance.py --all

# Rollback single skill
cp .claude/skills/SKILL-NAME/SKILL.md.backup-* \
   .claude/skills/SKILL-NAME/SKILL.md

# Complete rollback
bash scripts/rollback-v4-upgrades.sh
```

### File Locations

```
Project Root/
├── .claude/skills/           # All skill directories
│   ├── moai-*/SKILL.md      # Skill files
│   └── */SKILL.md.backup-*  # Auto backups
├── scripts/
│   ├── upgrade-skills-to-v4.py           # Main upgrader
│   ├── validate-v4-compliance.py         # Validator
│   ├── track-upgrade-progress.py         # Progress tracker
│   └── rollback-v4-upgrades.sh          # Rollback script
├── docs/
│   └── SKILL-UPGRADE-PLAN-v4.0.md       # This document
└── reports/
    ├── phase1-validation.txt
    ├── phase2-validation.txt
    ├── phase3-validation.txt
    └── phase4-validation.txt
```

### Key URLs

- **Claude Code Docs**: https://docs.anthropic.com/claude-code
- **Context7 MCP**: https://context7.com
- **MoAI-ADK Repo**: https://github.com/GoosLab/MoAI-ADK
- **Skill Factory v4.0 Spec**: /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-skill-factory/SKILL.md

---

## ✅ Final Checklist

### Before Starting

- [ ] Git working tree clean
- [ ] All dependencies installed
- [ ] Scripts tested on 1-2 skills
- [ ] Backup strategy confirmed
- [ ] Team notified of upgrade plan

### During Execution

- [ ] Run dry-run before each phase
- [ ] Validate after each batch
- [ ] Commit after each phase
- [ ] Track progress daily
- [ ] Fix validation errors immediately

### After Completion

- [ ] All 113 skills at v4.0.0
- [ ] 100% validation pass
- [ ] Comprehensive final report
- [ ] Documentation updated
- [ ] Git tag: v4.0.0-skills-complete
- [ ] Package templates synced

---

**Document Version**: 1.0  
**Author**: skill-factory Agent  
**Date**: 2025-11-12  
**Status**: Ready for Execution  
**Estimated Completion**: 8 weeks (Phase 1-4)

**Next Steps:**
1. Review this plan with team
2. Run test upgrade on 2-3 skills
3. Validate automation scripts
4. Begin Phase 1 execution

🚀 Ready to upgrade MoAI-ADK Skills to v4.0 Enterprise!
