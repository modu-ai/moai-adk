# MoAI-ADK (Modu-AI's Agentic Development Kit)

**🏆 Claude Code 환경에서 가장 완전한 Spec-First TDD 개발 프레임워크**

**🎯 0.1.9+ Major Modernization: TRUST Compliance + Next-Gen Toolchain (10-100x Faster) ⚡**

[![Version](https://img.shields.io/github/v/release/modu-ai/moai-adk?label=release)](https://github.com/modu-ai/moai-adk/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.11%2B-blue)](https://www.python.org/)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-compatible-purple)](https://docs.anthropic.com/claude-code)
[![Modern Toolchain](https://img.shields.io/badge/toolchain-uv%20%2B%20ruff-orange)](https://github.com/astral-sh/uv)
[![Tests](https://img.shields.io/badge/tests-100%25%20Git%20+%2091.7%25%20cc--manager-brightgreen)](https://github.com/modu-ai/moai-adk)
[![TAG System](https://img.shields.io/badge/16--Core%20TAG-103%20occurrences-blue)](https://github.com/modu-ai/moai-adk)

---

## 🚀 **0.1.9+ Complete Modernization Achievements**

### ⚡ **Language-Neutral Development Framework**

MoAI-ADK는 **모든 프로그래밍 언어**를 지원하는 Claude Code 통합 개발 프레임워크입니다.
프로젝트의 기술 스택을 자동 감지하고 각 언어별 최적의 도구를 제공합니다.

**지원 언어 및 자동 도구 매핑**:
- 🐍 **Python**: uv (10-100x faster) + ruff (100x faster) + pytest
- 📜 **JavaScript/TypeScript**: pnpm + biome + vitest
- 🦫 **Go**: go mod + golangci-lint + go test
- 🦀 **Rust**: cargo + clippy + cargo test
- ☕ **Java/Kotlin**: gradle + ktlint + junit
- 🔷 **.NET**: dotnet + analyzers + xunit
- 🎯 **Swift**: swift-format + SwiftLint + XCTest
- 🎯 **Dart/Flutter**: dart format + flutter analyze + flutter test

### 🛠️ **Auto-Detection & Optimization**

- **Project Type Detection**: 파일 구조 분석으로 언어/프레임워크 자동 감지
- **Optimal Toolchain**: 각 언어별 가장 빠르고 현대적인 도구 자동 설정
- **Performance Benchmarks**: 269 issues in 0.77s, formatting in 0.019s (Python 기준)
- **Cross-Platform**: Windows/macOS/Linux 완전 지원

### 🛠️ **TRUST Principles Compliance (87.6% Issue Reduction)**

- **Code Quality Improvement**: 1,904 → 236 issues (87.6% reduction)
- **Module Decomposition**: 70%+ LOC reduction in monolithic files
- **Complete Internationalization**: All Korean comments → English
- **4 Legacy Files Removed**: 2,606 lines of deprecated code cleaned
- **Modern Standards**: All new modules follow TRUST 5 principles

### 📊 **Enhanced Documentation & TAG System**

- **103 @TAG Occurrences**: Complete 16-Core TAG implementation across 20 files
- **Living Document Sync**: Real-time code-documentation synchronization
- **MkDocs + Material**: 85 API modules auto-generated (0.54s build time)
- **Traceability Matrix**: Complete @REQ → @DESIGN → @TASK → @TEST chains
- **SPEC Completion**: 9/9 specifications fully implemented and tested

### 🔧 **Development Workflow Revolution**

- **Ultra-Fast Development**: uv + ruff toolchain for instant feedback
- **TRUST-Compliant Architecture**: T-est first, R-eadable, U-nified, S-ecured, T-rackable
- **Modern Makefile**: Parallel execution, colored output, performance metrics
- **Global Ready**: Complete English internationalization for worldwide adoption
- **Zero-Config Setup**: Automatic toolchain detection and optimization

### 🔄 **이전 버전과의 연계**

- **SPEC-009 SQLite 기반**: 83배 성능 향상의 TAG 시스템을 문서에 활용
- **기존 문서 활용**: README.md, CHANGELOG.md를 온라인 사이트에 완전 통합
- **백워드 호환성**: 기존 워크플로우는 그대로 유지하면서 문서만 자동화

---

## 🚀 Executive Summary

MoAI-ADK는 Claude Code 환경에서 **/moai:0-project → /moai:3-sync** 4단계 파이프라인과 **/moai:git:\*** 명령군을 제공하여, Git을 몰라도 Spec-First TDD 개발을 수행할 수 있도록 설계된 Agentic Development Kit입니다.

| 핵심 역량   | Personal Mode                                             | Team Mode                                     |
| ----------- | --------------------------------------------------------- | --------------------------------------------- |
| 작업 보호   | Annotated Tag 기반 자동 체크포인트 (파일 변경 / 5분 주기) | GitFlow + Draft PR + 7단계 커밋 템플릿        |
| 명세/브랜치 | `/moai:1-spec` → 로컬 SPEC 생성                           | `/moai:1-spec` → GitHub Issue + 브랜치 템플릿 |
| TDD 지원    | `/moai:2-build` → 체크포인트 + RED/GREEN/REFACTOR         | `/moai:2-build` → 7단계 자동 커밋 + CI 게이트 |
| 동기화      | `/moai:3-sync` → 문서 동기화 + TAG 인덱스 갱신            | `/moai:3-sync` → PR Ready, 리뷰어/라벨 자동화 |

**Git 명령어 시스템** (`/moai:git:*`)

- `checkpoint`, `rollback`, `branch`, `commit`, `sync` 5종으로 Git 자동화를 완성합니다.
- 모든 명령은 TRUST 5원칙과 16-Core TAG 추적성을 준수하도록 설계되었습니다.

---

## ⚙️ 설치 & 초기화

### 🚀 Any Language, Optimal Setup

```bash
# MoAI-ADK 설치 (Python required for CLI)
pip install moai-adk         # 기존 pip 사용자
uv pip install moai-adk      # 현대적 uv 사용자 (권장)

# 새 프로젝트 초기화
moai init my-project
cd my-project

# Claude Code에서 자동 설정 (Magic happens here! ✨)
# /moai:0-project 실행 시:
# → 프로젝트 언어/프레임워크 자동 감지
# → 최적 도구체인 자동 설정
# → 언어별 품질 게이트 자동 적용
```

### 📋 사용 예시

```bash
# Python 프로젝트
moai init fastapi-project
# → Python 감지 → uv + ruff + pytest 설정

# TypeScript 프로젝트
moai init next-app
# → TypeScript 감지 → pnpm + biome + vitest 설정

# Go 프로젝트
moai init gin-api
# → Go 감지 → go mod + golangci-lint 설정

# 풀스택 프로젝트
moai init fullstack-app
# → backend/ (Python) + frontend/ (React) 자동 감지
```

### 🔧 선택적 도구

- **Python**: `uv` 설치로 10-100배 빠른 패키지 관리
- **Node.js**: `pnpm` 설치로 빠른 의존성 관리
- **팀 협업**: GitHub CLI (`gh`) + Claude Code GitHub App

---

## 🧭 4단계 파이프라인

```mermaid
flowchart LR
    A[/moai:0-project] --> B[/moai:1-spec]
    B --> C[/moai:2-build]
    C --> D[/moai:3-sync]
```

| 단계 | 명령어            | 담당 에이전트   | 산출물                                                     |
| ---- | ----------------- | --------------- | ---------------------------------------------------------- |
| 0    | `/moai:0-project` | project-manager | `.moai/project/{product,structure,tech}.md`, CLAUDE 메모리 |
| 1    | `/moai:1-spec`    | spec-builder    | Personal: 로컬 SPEC, Team: GitHub Issue + 브랜치 템플릿    |
| 2    | `/moai:2-build`   | code-builder    | TDD 구현, 체크포인트 or 7단계 커밋                         |
| 3    | `/moai:3-sync`    | doc-syncer      | Living Document 동기화, TAG 인덱스, PR Ready               |

보조 명령어: `/moai:git:checkpoint`, `/moai:git:rollback`, `/moai:git:branch`, `/moai:git:commit`, `/moai:git:sync`.

---

## 🤖 핵심 에이전트 생태계

| 에이전트            | 역할                                         |
| ------------------- | -------------------------------------------- |
| **project-manager** | `/moai:0-project` 인터뷰, 프로젝트 문서 생성 |
| **cc-manager**      | Claude Code 권한/훅/환경 최적화              |
| **spec-builder**    | 프로젝트 문서 기반 SPEC 자동 제안/작성       |
| **code-builder**    | TDD RED→GREEN→REFACTOR 실행                  |
| **doc-syncer**      | 문서/TAG/PR 동기화 및 보고                   |
| **git-manager**     | 체크포인트/브랜치/커밋/동기화 전담           |

필요 시 사용자 정의 에이전트를 `.claude/agents/` 아래 추가해 특정 도메인 업무를 확장할 수 있습니다.

---

## 🧭 TRUST 원칙 & 개발 가이드

- `.moai/memory/development-guide.md`: MoAI 개발 가이드 (TRUST 원칙, Waiver 제도 포함)
- `.claude/settings.json`: `defaultMode = acceptEdits`, 고위험 작업은 ask/deny로 분리
- `.moai/config.json`: Personal/Team Git 전략, 체크포인트 정책, TRUST 원칙 설정

**TRUST 5원칙 요약**

- **T** - **Test First** (테스트 우선): 코드 전에 테스트를 작성하라
- **R** - **Readable** (읽기 쉽게): 미래의 나를 위해 명확하게 작성하라
- **U** - **Unified** (통합 설계): 계층을 나누고 책임을 분리하라
- **S** - **Secured** (안전하게): 로그를 남기고 검증하라
- **T** - **Trackable** (추적 가능): 버전과 태그로 히스토리를 관리하라

**✨ 새로운 품질 개선 시스템 (SPEC-002 완료)**

- **GuidelineChecker**: Python 코드 TRUST 원칙 자동 검증 엔진
- **실시간 품질 게이트**: 함수 길이, 파일 크기, 매개변수, 복잡도 자동 검사
- **TDD 지원**: Red-Green-Refactor 사이클 자동화
- **성능 최적화**: AST 캐싱, 병렬 처리, 66.7% 캐시 히트율 달성

검증 도구: `python .moai/scripts/check_constitution.py`, `python .moai/scripts/check-traceability.py --update`

---

## 🏷️ 16-Core @TAG 시스템

| 체인               | 태그                               |
| ------------------ | ---------------------------------- |
| **Primary**        | `@REQ → @DESIGN → @TASK → @TEST`   |
| **Steering**       | `@VISION → @STRUCT → @TECH → @ADR` |
| **Implementation** | `@FEATURE → @API → @UI → @DATA`    |
| **Quality**        | `@PERF → @SEC → @DOCS → @TAG`      |

`/moai:3-sync`는 `.moai/reports/sync-report.md`와 `.moai/indexes/tags.json`을 갱신하여 추적성을 유지합니다.

---

## 📂 프로젝트 구조 (요약)

```
MoAI-ADK/
├── src/moai_adk/                # Python 패키지
│   ├── core/
│   │   ├── docs/                # 📖 새로운 온라인 문서 시스템 (SPEC-010)
│   │   │   ├── documentation_builder.py  # MkDocs 빌드 관리
│   │   │   ├── api_generator.py          # API 문서 자동 생성
│   │   │   └── release_notes_converter.py # sync-report → Release Notes
│   │   └── quality/             # ✨ 품질 개선 시스템 (SPEC-002)
│   │       └── guideline_checker.py # TRUST 원칙 자동 검증 엔진
│   ├── cli/, install/           # CLI & 설치 시스템
│   └── utils/                   # 공통 유틸리티
├── docs/                        # 📖 온라인 문서 사이트 (MkDocs 기반)
│   ├── getting-started/         # 시작 가이드
│   ├── guide/                   # 사용자 가이드
│   ├── development/            # 개발자 가이드
│   ├── examples/               # 예제
│   ├── reference/              # API 문서 (자동 생성)
│   ├── releases/               # 릴리스 노트 (자동 생성)
│   └── gen_ref_pages.py        # 자동 생성 스크립트
├── mkdocs.yml                   # 📖 MkDocs 설정
├── .github/workflows/docs.yml   # 📖 문서 자동 배포
├── .claude/                     # Claude Code 설정/에이전트/명령어
├── .moai/                       # MoAI 설정, 스크립트, 메모리, TAG 인덱스
├── scripts/, tests/             # 유틸리티 스크립트 및 테스트
├── CLAUDE.md                    # 프로젝트 메모리
└── README.md                    # 이 문서
```

---

## 🔧 개발 & 기여 워크플로우

### 🛠️ MoAI-ADK 개발 참여

```bash
# 개발 환경 구성
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# Python 개발 환경 (uv 권장)
uv pip install -e .         # uv 사용자
pip install -e .             # 기존 pip 사용자

# 현대적 Python 도구 사용
python scripts/build.py              # 통합 빌드 시스템
python scripts/test_runner.py        # 크로스플랫폼 테스트 실행
python scripts/version_manager.py status  # 버전 상태 확인

# Makefile 단축 명령어 (내부적으로 Python 도구 사용)
make build                   # Python 빌드 시스템 실행
make test                    # Python 테스트 러너 실행
make version-check           # 통합 버전 관리 시스템
```

### 🌍 다국어 프로젝트에서 사용하기

MoAI-ADK는 Python으로 작성되었지만, **모든 언어 프로젝트**에서 사용할 수 있습니다:

```bash
# 어떤 언어든 상관없이
moai init my-awesome-project
cd my-awesome-project

# /moai:0-project 실행하면
# → 자동으로 언어 감지 및 도구 설정
# → Go면 go.mod, TypeScript면 tsconfig.json 등을 감지
# → 해당 언어의 최적 도구체인 자동 구성
```

### 🔄 통합 개발 도구

**새로운 크로스플랫폼 Python 도구**:
- `python scripts/build.py` → 통합 빌드 자동화 (build.sh 대체)
- `python scripts/test_runner.py` → 종합 테스트 실행 (run-tests.sh 대체)
- `python scripts/version_manager.py` → 통합 버전 관리 시스템

**MoAI 자동화 스크립트**:
- `python .moai/scripts/detect_project_type.py` → 언어/프레임워크 자동 감지
- `python .moai/scripts/doc_sync.py` → 최신 문서/상태 리포트 생성
- `moai update --check` → 템플릿/스크립트 최신 상태 확인

> **📢 Shell 스크립트 제거**: 크로스플랫폼 지원을 위해 모든 shell 스크립트를 Python으로 대체했습니다.

---

## 📚 문서 & 참고 자료

### 📖 온라인 문서 사이트 (SPEC-010 테스트 완료 ✅)
- **[MoAI-ADK Documentation](https://moai-adk.github.io)** - 완전 자동화된 온라인 문서
- **로컬 테스트**: `mkdocs serve` → http://127.0.0.1:8000/ 성공 (0.54초 빌드)
- **85개 API 모듈**: CLI(7개), Core(33개), Install(5개), Utils(3개), Resources(37개) 자동 생성
- **Getting Started**: 설치부터 첫 프로젝트까지 단계별 가이드
- **User Guide**: 4단계 워크플로우 상세 설명
- **API Reference**: 소스코드에서 자동 생성되는 완전한 API 문서 ✅
- **Development**: 기여 방법 및 아키텍처 가이드
- **Examples**: 실제 사용 예제 및 템플릿
- **Release Notes**: sync-report 기반 자동 생성 릴리스 노트

**🎯 테스트 성과**: MkDocs Material 전문 사이트, HTTP 200 OK 정상 서비스, 25,842 bytes 홈페이지 생성

### 📄 로컬 문서 & 예제
- [종합 개발 가이드](docs/MOAI-ADK-GUIDE.md)
- [SPEC 작성 예제](examples/specs/README.md) - EARS 형식 학습용
- [언어별 설정 가이드](docs/languages/) - Python, JS/TS, Go, Rust 등
- [Troubleshooting Guide](docs/MOAI-ADK-GUIDE.md#️-troubleshooting-guide)
- [System Verification](docs/MOAI-ADK-GUIDE.md#-system-verification)

---

## 🤝 기여

1. [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)로 버그/아이디어 제안
2. Fork 후 Pull Request 제출 (테스트/문서 동반 권장)
3. 문서 개선 및 예제 추가 환영

자세한 내용은 `docs/CONTRIBUTING.md`를 참고하세요.

---

## 📝 라이선스 & 지원

- **License**: [MIT](LICENSE)
- **이슈/디스커션**: [Issues](https://github.com/modu-ai/moai-adk/issues) · [Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **공식 문서**: [docs/](docs/)

---

---

**🗿 "명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."**

**MoAI-ADK** | **Universal Claude Code Development Framework for All Languages**

---

### 🌟 차별화 포인트

- 🌍 **언어 중립**: Python부터 Rust까지 모든 언어 지원
- ⚡ **최적 도구**: 각 언어별 가장 빠른 현대적 도구 자동 선택
- 🤖 **Claude Code 완전 통합**: 전용 에이전트와 워크플로우 제공
- 📊 **16-Core TAG 추적**: 요구사항부터 구현까지 완전한 추적성
- 🔄 **Living Document**: 코드와 문서 자동 동기화
- 🎯 **Spec-First TDD**: 체계적인 개발 프로세스 자동화

**Made with ❤️ for Global Developers using Claude Code**
