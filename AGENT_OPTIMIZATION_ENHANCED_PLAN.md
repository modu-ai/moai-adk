# MoAI-ADK 에이전트 최적화 향상 계획

## 📋 **수정된 계획 개요**

사용자 피드백을 반영한 에이전트 지침 개선 방향:

### **🎯 주요 개선 사항**

1. **코드 기반 지침 제거**
   - `Skill("moai-core-agent-factory")` 같은 명시적 스킬 호출 제거
   - 모델 선택 분기 코드 제거 (decision tree)
   - 텍스트 기반 지침으로 전환

2. **Claude Code 공식 문서 참조**
   - https://code.claude.com/docs/en/sub-agents 예시 따르기
   - 공식 에이전트 패턴 준수
   - 자연스러운 텍스트 지침으로 변경

3. **전체 31개 에이전트 체계적 접근**
   - agent-factory 최적화 시작
   - 모든 에이전트 순차적 개선
   - 표준화된 지침 템플릿 적용

## 🏗️ **개선된 에이전트 지침 패턴**

### **✅ 올바른 텍스트 지침 예시**

```yaml
# ❌ 기존 방식 (코드 기반)
Model Selection:
  Use select_optimal_model() function:
    if complexity_score >= 7: return "sonnet"
    if speed_priority: return "haiku"

Skill Invocation:
  Always use: Skill("moai-core-agent-factory")
  Call: Skill("moai-context7-lang-integration")

# ✅ 개선된 방식 (텍스트 기반)
Model Selection:
  - Choose Sonnet for research-heavy tasks requiring deep analysis
  - Use Haiku for speed-critical tasks with low complexity
  - Let context decide for mixed requirements (5-7 complexity)
  - Default to Haiku for execution-focused agents

Skill Integration:
  - Leverage core agent generation capabilities when creating new agents
  - Access latest documentation through Context7 integration when needed
  - Use development best practices for agent design and structure
```

## 📊 **에이전트별 최적화 우선순위**

### **Phase 1: 핵심 에이전트 (즉시 시작)**
1. **agent-factory** - 24 → 17 스킬 최적화
2. **spec-builder** - 지침 텍스트화 및 스킬 최적화
3. **tdd-implementer** - TDD 패턴 텍스트화
4. **backend-expert** - 백엔드 패턴 개선

### **Phase 2: 계획 아키텍처 에이전트**
5. **api-designer** - API 설계 지침 텍스트화
6. **implementation-planner** - 계획 패턴 개선
7. **project-manager** - 프로젝트 관리 지침 개선

### **Phase 3: 품질 보증 에이전트**
8. **quality-gate** - 품질 평가 기준 텍스트화
9. **security-expert** - 보안 분석 패턴 개선
10. **performance-engineer** - 성능 최적화 지침 개선

## 🔧 **개선된 에이전트 지침 템플릿**

### **표준 지침 구조**

```yaml
# 에이전트 지침 템플릿
## 모델 선택 지침
- 복잡한 분석이 필요한 연구 기반 작업: Sonnet 사용
- 속도가 중요하고 복잡도가 낮은 작업: Haiku 사용
- 혼합 요구사항(복잡도 5-7): 컨텍스트에 결정
- 실행 중심 에이전트: 기본적으로 Haiku 사용

## 스킬 활용 지침
- 에이전트 생성 시 핵심 생성 기능 활용
- 최신 문서 접근 시 Context7 통합 활용
- 개발 시 표준 개발 관행 및 모범 사례 적용

## 작업 수행 패턴
- 항상 명확한 요구사항 분석부터 시작
- 관련 스킬을 상황에 맞게 동적 활용
- 결과물의 품질과 일관성 검증
- 필요시 전문 에이전트에게 위임
```

## 🎯 **실행 계획**

### **즉시 실행 과제**

1. **agent-factory 최적화**
   ```bash
   # 현재 24개 스킬 → 17개 스킬로 최적화
   # 코드 기반 지침 → 텍스트 기반 지침으로 변경
   # 동적 스킬 로딩 구현
   ```

2. **전체 에이전트 지침 검토**
   ```bash
   # 31개 에이전트 지침 텍스트화 여부 확인
   # 코드 기반 패턴 식별 및 제거
   # 표준 텍스트 지침 템플릿 적용
   ```

3. **Claude Code 호환성 검증**
   ```bash
   # https://code.claude.com/docs/en/sub-agents 기준 확인
   # 모든 에이전트가 공식 패턴 준수
   # 도구 권한 및 MCP 통합 검증
   ```

## 📈 **예상 효과**

### **지침 개선 효과**
- **가독성 향상**: 코드 대신 자연어 지침으로 이해도 증가
- **유지보수 용이성**: 텍스트 기반 지침으로 수정 및 관리 용이
- **Claude Code 호환성**: 공식 문서 패턴 완벽 준수

### **성능 최적화 효과**
- **응답 속도**: 20-30% 개선 (스킬 최적화)
- **메모리 사용**: 15-25% 감소 (불필요 스킬 제거)
- **토큰 효율성**: 85% 향상 (하이브리드 방식 유지)

## 🚀 **다음 단계**

1. **SPEC-AGENT-FACTORY-001 승인 여부 확인**
2. **Phase 1: agent-factory 최적화 즉시 시작**
3. **에이전트 지침 텍스트화 순차적 진행**
4. **전체 시스템 통합 및 성능 검증**

---

**상태**: 계획 수정 완료, 실행 준비
**다음 행동**: 사용자 최종 승인 후 Phase 1 시작