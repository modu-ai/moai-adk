# Changelog

All notable changes to MoAI-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added (2025-10-02)

- **AlfredUpdateBridge 클래스**: /alfred:9-update Phase 4를 Claude Code 도구로 처리
- **프로젝트 문서 보호**: {{PROJECT_NAME}} 패턴 기반 조건부 복사
- **훅 파일 권한 처리**: chmod +x 자동 적용 (Unix 계열)
- **Output Styles 복사**: .claude/output-styles/alfred/ 디렉토리 추가
- **--check-quality 옵션**: TRUST 5원칙 검증 기능 준비
- **Output Styles 재구축**: Alfred SuperAgent 통합, 9개 전문 에이전트 소개, 다중 언어 예제 (TypeScript, Python, Go, Rust, Flutter)

### Changed (2025-10-02)

- **UpdateOrchestrator**: Phase 4를 AlfredUpdateBridge에 위임
- **TemplateCopier**: output-styles/alfred 복사 대상 추가
- **UpdateVerifier**: output-styles/alfred 검증 추가
- **/alfred:9-update.md**: v2.0.0 업데이트 (Phase 4 전면 개편)
- **alfred-pro.md**: 914줄 → 405줄 압축 (55% 감소), Alfred 오케스트레이션 다이어그램, 다중 언어 TDD 예제
- **pair-collab.md**: 433줄 → 399줄 개선, 다중 언어 코드 리뷰 예제 (TypeScript, Python, Go)
- **study-deep.md**: 444줄 → 399줄 개선, 프레임워크별 학습 경로 (Express, FastAPI, Gin, Axum)
- **beginner-learning.md**: 224줄 → 324줄 보강, Alfred 9개 에이전트 소개, TRUST 5원칙 비유, Python/Flutter 예제

### Fixed (2025-10-02)

- 문서-구현 불일치 해소 (5개 Critical, 3개 Medium 이슈)
- 프로젝트 문서 무손실 업데이트
- 훅 파일 실행 권한 누락 문제
- output-styles 디렉토리 복사 누락 문제

### Technical Details

- **SPEC**: SPEC-UPDATE-REFACTOR-001
- **테스트**: 7개 (100% 통과)
- **커버리지**: 79-96%
- **TAG**: @CODE:UPDATE-REFACTOR-001

## [0.1.0] - 2025-10-02

### 🎉 **첫 공식 릴리스 - moai-adk**

**0.0.x 테스트 버전을 완전히 정리하고 0.1.0으로 공식 출시합니다**

#### 🎯 주요 변경사항

##### 1. 버전 정책 확립
- **패키지명**: `moai-adk` (단순명료)
- **버전 전략**: 0.1.0부터 공식 시작
- **이전 버전**: 0.0.1-0.0.2 테스트 버전 완전 삭제
- **배포 준비**: GitHub 공개, npm 퍼블릭 배포 준비 완료

##### 2. 공식 배포 준비
- ✅ GitHub Repository: `modu-ai/moai-adk`
- ✅ npm 배포 설정 완료
- ✅ 문서 정비: README.md 전면 개편
- ✅ 키워드 최적화: `spec-first`, `claude-code`, `ai-agent`, `tdd`

##### 3. 문서 개선
- **README.md**: npm 패키지용으로 완전히 재작성
  - 설치 명령어: `npm install -g moai-adk` 또는 `bun add -g moai-adk`
  - 9개 전문 에이전트 시스템 상세 설명
  - 4-Core TAG 시스템 문서화
  - 프로그래매틱 API 문서 추가
  - 실전 시나리오 예제 보강
- **CHANGELOG.md**: 체계적 변경 이력 관리
- **GitHub URL**: `https://github.com/modu-ai/moai-adk` 통일

##### 4. 배포 시스템 개선
- **.npmignore**: 불필요한 파일 제외 설정 (47개 파일, 601.5 KB)
- **빌드 검증**: `prepublishOnly` 스크립트로 CI 자동 실행
- **크로스 플랫폼 지원**: Windows/macOS/Linux 검증 완료
- **프로젝트 정리**: 150MB+ 개발 아티팩트 삭제, 깔끔한 구조 확립

#### ✨ 핵심 기능 (유지)

모든 MoAI-ADK 핵심 기능은 그대로 유지됩니다:

- 🎯 **SPEC-First TDD Workflow**: 3단계 개발 프로세스
- 🤖 **9개 전문 에이전트**: Alfred SuperAgent 오케스트레이션
- 🏷️ **4-Core @TAG System**: 완전한 추적성
- 🌍 **Universal Language Support**: TypeScript, Python, Java, Go, Rust
- 🔒 **TRUST 5원칙**: Test, Readable, Unified, Secured, Trackable

#### 📦 설치 방법

```bash
# npm
npm install -g moai-adk

# Bun (권장, 5배 빠름)
bun add -g moai-adk

# 설치 확인
moai --version  # v0.1.0
moai doctor     # 시스템 환경 체크
```

#### 🚀 빠른 시작

```bash
# 1. 새 프로젝트 초기화
moai init my-awesome-project

# 2. 프로젝트 문서 생성 (선택)
cd my-awesome-project
/alfred:8-project

# 3. SPEC-First TDD 워크플로우 시작
/alfred:1-spec "사용자 인증 기능"
/alfred:2-build
/alfred:3-sync
```

#### ⚠️ Breaking Changes

**이전 0.0.x 테스트 버전을 사용하던 경우**:

```bash
# 기존 버전 제거
npm uninstall -g moai-adk

# 0.1.0 설치
npm install -g moai-adk@0.1.0

# 확인
moai --version  # v0.1.0
```

**주요 변경사항**:
- 버전 0.0.1-0.0.2는 npm에서 삭제됨
- 0.1.0이 첫 공식 릴리스
- CLI 명령어와 기능은 동일하게 유지

#### 🎉 향후 계획

**v0.1.x** (안정화 및 개선):
- 커뮤니티 피드백 반영
- 버그 수정 및 성능 최적화
- 다국어 문서 확장
- TypeScript 타입 에러 완전 수정

**v0.2.0** (기능 확장):
- 추가 프로그래밍 언어 지원 (C#, Ruby, Kotlin)
- TAG 시스템 고급 쿼리 기능
- SPEC 템플릿 라이브러리

**v1.0.0** (프로덕션 릴리스):
- Web UI 대시보드
- 고급 AI 페어 프로그래밍 기능
- 엔터프라이즈 지원

#### 📝 참고 링크

- **공식 문서**: https://moai-adk.vercel.app
- **GitHub**: https://github.com/modu-ai/moai-adk
- **npm**: https://www.npmjs.com/package/moai-adk

---

## [0.0.2] - 2025-10-01 (레거시)

### 🎯 **TAG System - 체계 개선**

**TAG 시스템을 대폭 단순화하여 TDD와 완벽하게 정렬했습니다**

#### 🌟 주요 변경사항

##### 1. TAG 체계 단순화 (50% 감소)
- **Before (이전 버전)**: 8개 TAG 체계
  - Primary: `@REQ`, `@DESIGN`, `@TASK`, `@TEST`
  - Implementation: `@FEATURE`, `@API`, `@UI`, `@DATA`
- **After (현재 버전)**: 4개 TAG 체계
  - `@SPEC:ID` → `@TEST:ID` → `@CODE:ID` → `@DOC:ID`

##### 2. TDD 사이클 완벽 정렬
- **RED Phase**: `@TEST:ID` 작성 (tests/ 디렉토리)
- **GREEN Phase**: `@CODE:ID` 구현 (src/ 디렉토리)
- **REFACTOR Phase**: `@CODE:ID` 개선 + `@DOC:ID` 문서화

##### 3. 구현 세부사항 주석 레벨화
- **이전**: 파일 레벨 TAG (@FEATURE, @API, @UI, @DATA)
- **현재**: 주석 레벨 서브카테고리
  - `@CODE:ID:API` - REST API, GraphQL
  - `@CODE:ID:UI` - 컴포넌트, 화면
  - `@CODE:ID:DATA` - 데이터 모델
  - `@CODE:ID:DOMAIN` - 비즈니스 로직
  - `@CODE:ID:INFRA` - 인프라, 외부 연동

##### 4. TAG BLOCK 템플릿 단순화
```typescript
// 이전 버전 (156 characters)
// @TASK:AUTH-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
// Related: @FEATURE:AUTH-001, @API:AUTH-001

// 현재 버전 (78 characters, 50% reduction)
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

#### 📊 성능 개선

| 항목 | 이전 | 현재 | 개선율 |
|------|------|------|--------|
| TAG 개수 | 8개 | 4개 | -50% |
| TAG BLOCK 길이 | 156자 | 78자 | -50% |
| TDD 정렬도 | 60/100 | 100/100 | +67% |
| SPEC 매핑 | 50/100 | 95/100 | +90% |
| 실무 사용성 | 65/100 | 90/100 | +38% |
| **종합 점수** | **65/100** | **92/100** | **+42%** |

#### 🔧 마이그레이션 가이드

##### TAG 매핑 규칙
| 이전 버전 | 현재 버전 | 위치 |
|-----------|-----------|------|
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
# 이전 버전
rg '@(REQ|DESIGN|TASK|TEST|FEATURE|API|UI|DATA):' -n

# 현재 버전
rg '@(SPEC|TEST|CODE|DOC):' -n
```

#### 📚 업데이트된 문서

- **설계 문서**: `docs/analysis/tag-system-design.md` (신규)
- **분석 리포트**: `docs/analysis/tag-system-critical-analysis.md` (신규)
- **가이드**: `docs/guide/tag-system.md` (전면 개편)
- **핵심 가이드**: `CLAUDE.md`, `.moai/memory/development-guide.md` (업데이트)
- **템플릿**: `moai-adk-ts/templates/` (전체 업데이트)

#### ⚠️ Breaking Changes

- **TAG 형식 변경**: 이전 TAG 체계(@REQ, @DESIGN, @TASK 등)는 더 이상 지원하지 않음
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

- [TAG System Design](/docs/analysis/tag-system-design.md)
- [Critical Analysis Report](/docs/analysis/tag-system-critical-analysis.md)
- [TAG System Guide](/docs/guide/tag-system.md)
- [Development Guide](/.moai/memory/development-guide.md)

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
   - `/alfred:1-spec`: 명세 작성 (EARS 방식)
   - `/alfred:2-build`: TDD 구현 (RED→GREEN→REFACTOR)
   - `/alfred:3-sync`: 문서 동기화

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