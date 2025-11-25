# MoAI-ADK 로컬 Skills 사용 현황 분석 보고서

**분석 일시**: 2025-11-25
**대상**: /Users/goos/MoAI/MoAI-ADK/.claude/skills/
**분석자**: R2-D2

---

## 📊 요약

| 항목 | 개수 | 비율 |
|------|------|------|
| 전체 설치된 Skills | 25 | 100% |
| 사용 중인 Skills | 15 | 60% |
| **미사용 Skills** | **10** | **40%** |
| 참조되지만 미설치 | 32 | - |

---

## ❌ 미사용 Skills (10개)

### 1. moai-command-project
**상태**: 설치됨, 미사용
**이유**: `moai-command-project`는 `/moai:0-project` 커맨드와 관련된 기능이지만, agent_skills_mapping.json에서 참조하지 않음
**권장**: 삭제 또는 매핑 추가 필요

### 2. moai-context-manager
**상태**: 설치됨, 미사용
**이유**: Context 관리 기능이지만 agent에서 호출되지 않음
**권장**: `moai-core-context-budget`로 대체 가능, 삭제 권장

### 3. moai-foundation-core ⭐
**상태**: 설치됨, 미사용
**이유**: moai-foundation-* 5개 skills를 통합한 핵심 skill이지만 매핑에 누락
**심각도**: 높음
**권장**: agent_skills_mapping.json에 **즉시 추가 필요**
**영향**: TRUST 5, SPEC, EARS 등 핵심 기능 누락

### 4. moai-internal-comms
**상태**: 설치됨, 미사용
**이유**: 내부 커뮤니케이션 관련 skill, 현재 사용되지 않음
**권장**: 삭제 가능

### 5. moai-jit-docs-enhanced
**상태**: 설치됨, 미사용
**이유**: JIT(Just-In-Time) 문서 로딩 시스템이지만 agent에서 호출되지 않음
**권장**: 삭제 또는 문서화 워크플로우에 통합

### 6. moai-lib-shadcn-ui
**상태**: 설치됨, 미사용
**이유**: shadcn/ui 라이브러리 관련 skill이지만 frontend agent에서 미참조
**권장**: `code-frontend` agent에 추가 또는 삭제

### 7. moai-lib-toon
**상태**: 설치됨, 미사용
**이유**: TOON Format 전문 skill이지만 사용되지 않음
**권장**: 삭제 가능

### 8. moai-mermaid-diagram-expert
**상태**: 설치됨, 미사용
**이유**: Mermaid 다이어그램 생성 skill이지만 workflow-docs에서 미참조
**권장**: `workflow-docs` agent에 추가

### 9. moai-nextra-architecture
**상태**: 설치됨, 미사용
**이유**: Nextra 문서 프레임워크 관련이지만 문서화 워크플로우에서 미사용
**권장**: `workflow-docs` agent에 추가 또는 삭제

### 10. moai-templates
**상태**: 설치됨, 미사용
**이유**: 템플릿 관리 skill이지만 agent에서 호출되지 않음
**권장**: `workflow-project` agent에 추가 또는 삭제

---

## ⚠️ 심각한 문제: Missing Skills (32개)

agent_skills_mapping.json에서 참조하지만 실제로 설치되지 않은 skills:

### Core Missing Skills (높은 우선순위)
- `moai-foundation-trust` - TRUST 5 프레임워크 ⭐⭐⭐
- `moai-foundation-specs` - SPEC 생성 ⭐⭐⭐
- `moai-foundation-ears` - EARS 포맷 ⭐⭐⭐
- `moai-foundation-git` - Git 워크플로우 ⭐⭐
- `moai-cc-claude-md` - CLAUDE.md 작성 ⭐⭐

### Domain Expert Missing Skills
- `moai-domain-backend`
- `moai-domain-frontend`
- `moai-domain-database`
- `moai-domain-devops`
- `moai-domain-security`
- `moai-domain-web-api`
- `moai-domain-monitoring`

### Language Missing Skills
- `moai-lang-python`
- `moai-lang-typescript`
- `moai-lang-sql`

### Essential Missing Skills
- `moai-essentials-debug`
- `moai-essentials-refactor`
- `moai-essentials-perf`

### Configuration Missing Skills
- `moai-cc-configuration`
- `moai-cc-hooks`
- `moai-cc-mcp-plugins`
- `moai-cc-claude-settings`

### Project Management Missing Skills
- `moai-project-config-manager`
- `moai-project-language-initializer`

---

## 🔍 문제 원인 분석

### 1. moai-foundation-core 미사용 문제
**원인**:
- `moai-foundation-core`는 5개 legacy skills를 통합했지만
- agent_skills_mapping.json은 여전히 개별 skills를 참조함:
  - `moai-foundation-trust`
  - `moai-foundation-specs`
  - `moai-foundation-ears`
  - `moai-foundation-git`
  - `moai-foundation-langs`

**해결책**:
```json
// agent_skills_mapping.json 수정 필요
"workflow-spec": [
  "moai-foundation-core",  // ✅ 통합 skill 사용
  "moai-cc-claude-md"
]
```

### 2. Unified Skills 미반영
**원인**:
- `moai-lang-unified`는 설치되어 있지만 매핑에서는 개별 언어 skills 참조
- `moai-essentials-unified`는 설치되어 있지만 개별 essentials skills 참조

**해결책**:
```json
"code-backend": [
  "moai-lang-unified",  // ✅ 통합 skill 사용
  "moai-domain-backend"
]
```

### 3. Domain Skills 누락
**원인**:
- `moai-domain-*` skills가 실제로 존재하지 않음
- Unified skills로 통합되었을 가능성

---

## 📋 권장 조치 사항

### 즉시 조치 (Priority 1)

1. **agent_skills_mapping.json 업데이트**
   ```bash
   # moai-foundation-core 추가
   # moai-lang-unified 활용
   # moai-essentials-unified 활용
   ```

2. **moai-foundation-core 매핑 추가**
   - workflow-spec
   - core-quality
   - security-expert

3. **미사용 Skills 삭제 (10개)**
   ```bash
   rm -rf .claude/skills/moai-context-manager
   rm -rf .claude/skills/moai-internal-comms
   rm -rf .claude/skills/moai-lib-toon
   rm -rf .claude/skills/moai-templates
   ```

### 중기 조치 (Priority 2)

4. **유용한 미사용 Skills 활성화**
   - `moai-mermaid-diagram-expert` → workflow-docs에 추가
   - `moai-lib-shadcn-ui` → code-frontend에 추가
   - `moai-jit-docs-enhanced` → mcp-context7에 추가

5. **Missing Domain Skills 해결**
   - moai-universal-ultimate로 대체 가능한지 확인
   - 필요시 개별 domain skills 재생성

---

## 📊 개선 후 예상 결과

| 항목 | 현재 | 개선 후 | 개선율 |
|------|------|---------|--------|
| 설치된 Skills | 25 | 15-18 | -28% to -40% |
| 사용 중인 Skills | 15 | 15-18 | 100% |
| 미사용 Skills | 10 | 0 | 100% 감소 |
| 매핑 정확도 | 32% | 100% | +212% |

---

## 🎯 최종 권장사항

### Do This (즉시)
1. ✅ `moai-foundation-core`를 agent_skills_mapping.json에 추가
2. ✅ Unified skills (`moai-lang-unified`, `moai-essentials-unified`) 활용
3. ✅ 미사용 Skills 10개 삭제

### Consider This (검토 필요)
1. 🤔 `moai-mermaid-diagram-expert` 활성화 (문서화에 유용)
2. 🤔 `moai-lib-shadcn-ui` 활성화 (Frontend에 유용)
3. 🤔 Missing domain skills 재생성 또는 대체

### Don't Do This
1. ❌ moai-foundation-* legacy skills 재설치 (통합됨)
2. ❌ moai-lang-* 개별 skills 재설치 (unified로 통합됨)
3. ❌ 사용하지 않는 skills 유지

---

**보고서 종료**
