# Changelog

All notable changes to MoAI-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

2. **16-Core TAG 시스템 v0.0.1**
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