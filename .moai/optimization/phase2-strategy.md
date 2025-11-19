# Phase 2: Priority 3 스킬 최적화 전략

## 📊 대상 스킬 현황

**Priority 3 (500-800 lines)**: 38개 스킬, 총 22,479 lines

### 1. 코어 그룹 (6개, 3,277 lines)
- moai-cc-skill-factory (501 lines)
- moai-core-personas (505 lines)
- moai-core-spec-authoring (507 lines)
- moai-core-dev-guide (603 lines)
- moai-core-proactive-suggestions (712 lines)
- moai-core-rules (765 lines)

### 2. 보안 그룹 (5개, 2,877 lines)
- moai-security-compliance (505 lines)
- moai-security-zero-trust (522 lines)
- moai-security-api (771 lines)

### 3. 언어/도메인 그룹 (3개, 1,525 lines)
- moai-lang-kotlin (508 lines)
- moai-lang-go (509 lines)
- moai-domain-monitoring (668 lines)

### 4. 클라우드/BaaS 그룹 (4개, 2,257 lines)
- moai-baas-railway-ext (530 lines)
- moai-baas-clerk-ext (590 lines)
- moai-baas-convex-ext (622 lines)
- moai-baas-cloudflare-ext (655 lines)

### 5. MCP/설정 그룹 (4개, 2,204 lines)
- moai-cc-mcp-builder (523 lines)
- moai-cc-mcp-plugins (563 lines)
- moai-cc-hooks (572 lines)

### 6. 프론트엔드/디자인 그룹 (2개, 1,150 lines)
- moai-domain-figma (599 lines)
- moai-streaming-ui (551 lines)

### 7. 파운데이션 그룹 (5개, 3,247 lines)
- moai-foundation-git (531 lines)
- moai-foundation-ears (627 lines)
- moai-foundation-langs (676 lines)
- moai-foundation-specs (775 lines)
- moai-foundation-trust (776 lines)

### 8. 유틸리티 그룹 (9개, 5,942 lines)
- moai-change-logger (504 lines)
- moai-essentials-debug (509 lines)
- moai-core-session-state (572 lines)
- moai-learning-optimizer (574 lines)
- moai-mcp-builder (546 lines)
- moai-security-identity (564 lines)
- moai-document-processing (523 lines)
- moai-core-context-budget (768 lines)
- moai-mermaid-diagram-expert (783 lines)

## 🎯 그룹별 최적화 전략

### 코어 그룹 (예상 20% 감소)
- 중복 제거 및 템플릿화
- 공통 패턴 추출

### 보안 그룹 (예상 25% 감소)
- 보안 패턴 통합
- 레퍼런스 아키텍처 활용

### 언어/도메인 그룹 (예상 15% 감소)
- 템플릿 확장
- 예제 표준화

### 클라우드/BaaS 그룹 (예상 30% 감소)
- BaaS 패턴 통합
- 제공업체별 차이점 문서화

### MCP/설정 그룹 (예상 20% 감소)
- MCP 패턴 표준화
- 설정 템플릿 활용

### 프론트엔드/디자인 그룹 (예상 15% 감소)
- 컴포넌트 패턴 통합
- 디자인 시스템 연동

### 파운데이션 그룹 (예상 18% 감소)
- 파운데이션 패턴 정리
- TRUST/의존성 통합

### 유틸리티 그룹 (예상 22% 감소)
- 유틸리티 패턴 통합
- 컨텍스트 관리 표준화

## ⏱️ 실행 계획

### 일일 목표: 8-10개 스킬 (90분)
- 그룹별로 2-3일 배정
- 총 5-6일 소요 예상

### 최적화 목표
- **감축 목표**: 22,479 → 16,860 lines (25% 감소)
- **토큰 효율**: 85% → 90% 향상
- **처리 시간**: 5-6일 (일일 90분)

## 🔧 에이전트 활용 계획

### 주요 에이전트
1. **skill-optimizer**: 메인 최적화
2. **domain-expert**: 도메인별 전문성
3. **refactor-specialist**: 구조 개선
4. **quality-gate**: 최종 품질 검증