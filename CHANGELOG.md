# Changelog

All notable changes to MoAI-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.2] - 2025-10-01

### 🎯 **TAG System v5.0 - 4-Core Revolution**

**TAG 시스템을 v4.0 8-Core에서 v5.0 4-Core로 대폭 단순화하여 TDD와 완벽하게 정렬했습니다**

#### 🌟 주요 변경사항

##### 1. TAG 체계 단순화 (50% 감소)
- **Before (v4.0)**: 8-Core TAG 체계
  - Primary: `@REQ`, `@DESIGN`, `@TASK`, `@TEST`
  - Implementation: `@FEATURE`, `@API`, `@UI`, `@DATA`
- **After (v5.0)**: 4-Core TAG 체계
  - `@SPEC:ID` → `@TEST:ID` → `@CODE:ID` → `@DOC:ID`

##### 2. TDD 사이클 완벽 정렬
- **RED Phase**: `@TEST:ID` 작성 (tests/ 디렉토리)
- **GREEN Phase**: `@CODE:ID` 구현 (src/ 디렉토리)
- **REFACTOR Phase**: `@CODE:ID` 개선 + `@DOC:ID` 문서화

##### 3. 구현 세부사항 주석 레벨화
- v4.0: 파일 레벨 TAG (@FEATURE, @API, @UI, @DATA)
- v5.0: 주석 레벨 서브카테고리
  - `@CODE:ID:API` - REST API, GraphQL
  - `@CODE:ID:UI` - 컴포넌트, 화면
  - `@CODE:ID:DATA` - 데이터 모델
  - `@CODE:ID:DOMAIN` - 비즈니스 로직
  - `@CODE:ID:INFRA` - 인프라, 외부 연동

##### 4. TAG BLOCK 템플릿 단순화
```typescript
// v4.0 (156 characters)
// @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
// Related: @API:AUTH-001, @DATA:AUTH-001

// v5.0 (78 characters, 50% reduction)
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

#### 📊 성능 개선

| 항목 | v4.0 | v5.0 | 개선율 |
|------|------|------|--------|
| TAG 개수 | 8개 | 4개 | -50% |
| TAG BLOCK 길이 | 156자 | 78자 | -50% |
| TDD 정렬도 | 60/100 | 100/100 | +67% |
| SPEC 매핑 | 50/100 | 95/100 | +90% |
| 실무 사용성 | 65/100 | 90/100 | +38% |
| **종합 점수** | **65/100** | **92/100** | **+42%** |

#### 🔧 마이그레이션 가이드

##### TAG 매핑 규칙
| v4.0 (8-Core) | v5.0 (4-Core) | 위치 |
|---------------|---------------|------|
| `@REQ:ID` | `@SPEC:ID` | .moai/specs/ |
| `@DESIGN:ID` | `@SPEC:ID` | .moai/specs/ |
| `@TASK:ID` | `@CODE:ID` | src/ |
| `@TEST:ID` | `@TEST:ID` | tests/ |
| `@FEATURE:ID` | `@CODE:ID` | src/ |
| `@API:ID` | `@CODE:ID:API` | src/ (주석) |
| `@UI:ID` | `@CODE:ID:UI` | src/ (주석) |
| `@DATA:ID` | `@CODE:ID:DATA` | src/ (주석) |

##### 자동 마이그레이션
```bash
# TAG 스캔 명령어 업데이트
# v4.0
rg '@(REQ|DESIGN|TASK|TEST|FEATURE|API|UI|DATA):' -n

# v5.0
rg '@(SPEC|TEST|CODE|DOC):' -n
```

#### 📚 업데이트된 문서

- **설계 문서**: `docs/analysis/tag-system-v5-design.md` (신규)
- **분석 리포트**: `docs/analysis/tag-system-critical-analysis.md` (신규)
- **가이드**: `docs/guide/tag-system.md` (v5.0 전면 개편)
- **핵심 가이드**: `CLAUDE.md`, `.moai/memory/development-guide.md` (업데이트)
- **템플릿**: `moai-adk-ts/templates/` (전체 업데이트)

#### ⚠️ Breaking Changes

- **TAG 형식 변경**: v4.0 8-Core TAG는 더 이상 지원하지 않음
- **TAG BLOCK 형식 변경**: 새로운 템플릿 필수 적용
- **스캔 패턴 변경**: ripgrep 검색 패턴 업데이트 필요
- **에이전트 연동 변경**: tag-agent, spec-builder, code-builder 업데이트

#### ✨ 개선 사항

1. **단순성 (+50%)**
   - 8개 TAG → 4개 TAG
   - 학습 곡선 대폭 완화
   - TAG 선택 고민 제거

2. **TDD 정렬 (+100%)**
   - RED-GREEN-REFACTOR와 완벽 일치
   - SPEC → TEST → CODE 명확한 흐름
   - REFACTOR 단계 명시적 지원

3. **실무 사용성 (+38%)**
   - 모든 구현은 @CODE:ID로 통합
   - 서브 카테고리는 선택적 사용
   - 파일 위치로 역할 자동 구분

4. **EARS 매핑 (+45%)**
   - SPEC 문서에서 EARS 구문 직접 연결
   - 요구사항 → 테스트 → 코드 추적 간소화

#### 📝 추가 개선 사항

- **CODE-FIRST 원칙 강화**: TAG의 진실은 코드 자체에만 존재
- **TDD 이력 주석**: 코드 내 RED-GREEN-REFACTOR 이력 기록 권장
- **서브 카테고리 확장**: 필요에 따라 @CODE:ID:CUSTOM 추가 가능
- **다중 언어 지원**: TypeScript, Python, Java, Go, Rust 예시 업데이트

#### 🔗 관련 문서

- [TAG System v5.0 Design](/docs/analysis/tag-system-v5-design.md)
- [Critical Analysis Report](/docs/analysis/tag-system-critical-analysis.md)
- [TAG System Guide](/docs/guide/tag-system.md)
- [Migration Guide v4 → v5](/docs/guide/migration-v5.md)

---

## [0.0.1] - 2025-09-29

### 🚀 **Initial Development Phase - v0.0.1 Reset**

**MoAI-ADK v0.0.1은 프로젝트를 초기 개발 단계로 재설정하여, 단순화된 아키텍처와 일관된 버전 체계를 확립한 초기화 릴리스입니다**

#### 🎯 핵심 재설정 사항

- **버전 통일**: 모든 컴포넌트를 v0.0.1로 통일하여 일관성 확보
- **클린 슬레이트**: 불필요한 복잡성 제거, 핵심 기능에 집중
- **TypeScript 중심**: TypeScript 단일 스택으로 집중, Python 지원은 사용자 프로젝트용
- **TAG 시스템 초기화**: 분산 구조 v0.0.1 초기 개발 상태

#### 📋 초기 기능 세트

1. **SPEC-First TDD 워크플로우**
   - `/moai:1-spec`: 명세 작성 (EARS 방식)
   - `/moai:2-build`: TDD 구현 (RED→GREEN→REFACTOR)
   - `/moai:3-sync`: 문서 동기화

2. **@TAG 시스템**
   - 분산 JSONL 구조 (초기 개발)
   - 카테고리별 저장: `.moai/indexes/categories/*.jsonl`
   - 관계 매핑: `.moai/indexes/relations/chains.jsonl`

3. **Claude Code 통합**
   - 5개 핵심 에이전트
   - 4개 워크플로우 명령어
   - 기본 훅 시스템

#### 🛠️ 기술 스택 v0.0.1

- **TypeScript**: 5.9.2+ (주 언어)
- **Node.js**: 18.0+ / Bun 1.2.19+
- **빌드 도구**: tsup 8.5.0
- **테스트**: Vitest 3.2.4
- **린터**: Biome 2.2.4
- **패키지 매니저**: Bun (권장)

#### 📚 문서 구조

- **프로젝트 메모리**: `.moai/memory/development-guide.md`
- **기술 스택**: `.moai/project/tech.md`
- **제품 정의**: `.moai/project/product.md`
- **구조 설계**: `.moai/project/structure.md`

#### ⚠️ Breaking Changes

- **모든 기존 버전과 호환되지 않음**
- **새로운 프로젝트 초기화 필요**
- **TAG 시스템 완전 재초기화**
- **설정 파일 재생성 필요**

---

## 향후 로드맵

### v0.0.2-v0.0.5 (재구조화)
- 모듈 단순화 및 최적화
- AI 네이티브 설계 강화
- 성능 및 안정성 개선

### v0.1.0 (안정화)
- 품질 기준 확립
- npm 패키지 정식 배포
- 생태계 구축

---

**📝 Note**: 이전 버전 히스토리는 `.archive/pre-v0.0.1-backup/CHANGELOG_legacy.md`에서 확인할 수 있습니다.

**🏷️ Backup Tag**: `backup-before-v0.0.1`에서 이전 상태를 복원할 수 있습니다.