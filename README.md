# MoAI-ADK (Modu-AI's Agentic Development Kit)

**🏆 Claude Code 환경에서 가장 완전한 Spec-First TDD 개발 프레임워크**

**🎯 0.2.0 Latest Update: SPEC-010 온라인 문서 사이트 제작 - 완전 자동화된 Living Document 시스템**

[![Version](https://img.shields.io/github/v/release/modu-ai/moai-adk?label=release)](https://github.com/modu-ai/moai-adk/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.11%2B-blue)](https://www.python.org/)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-compatible-purple)](https://docs.anthropic.com/claude-code)
[![Tests](https://img.shields.io/badge/tests-100%25%20Git%20+%2091.7%25%20cc--manager-brightgreen)](https://github.com/modu-ai/moai-adk)
[![TAG System](https://img.shields.io/badge/16--Core%20TAG-69%20total%2C%2040%20completed-blue)](https://github.com/modu-ai/moai-adk)

---

## 🎉 **0.2.0 혁신적 성과 하이라이트**

### 📖 **SPEC-010 온라인 문서 사이트 제작**

- **전문적인 문서 사이트**: MkDocs Material 기반 완전 자동화된 온라인 문서
- **Living Document 원칙**: 코드 변경 시 문서 자동 동기화 및 GitHub Pages 배포
- **API 문서 자동화**: Python 소스코드에서 완전 자동 생성되는 API 레퍼런스
- **Release Notes 통합**: sync-report 구조를 활용한 체계적인 릴리스 노트

### 🏗️ **문서 시스템 아키텍처**

- **MkDocs + Material 테마**: 반응형 디자인, 다크/라이트 테마, 검색 기능
- **자동 생성 스크립트**: 소스코드 → API 문서, sync-report → Release Notes
- **CI/CD 완전 자동화**: GitHub Actions를 통한 무중단 배포
- **추적성 통합**: 16-Core TAG 시스템과 문서의 완벽한 연동

### 💎 **개발 경험 혁신**

- **체계적인 문서 구조**: Getting Started, User Guide, API Reference, Development, Examples
- **실시간 문서 동기화**: 코드 변경 즉시 문서 사이트 자동 갱신
- **커뮤니티 친화적**: 기여 방법, 예제, 개발 가이드 완전 통합
- **SEO 최적화**: 구조화된 메타데이터, 사이트맵, 검색 엔진 친화적 설계

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

```bash
# 패키지 설치
pip install moai-adk

# 새 프로젝트 (기본: personal)
moai init my-personal-project

# 팀 프로젝트
mkdir team-project && cd team-project
moai init --team

# 모드 전환 / 확인
moai config --mode team      # personal → team
moai config --mode personal  # team → personal
moai config --show
```

선택 의존성

- 개인: `pip install watchdog` (자동 체크포인트 감시)
- 팀: GitHub CLI(`gh`), Anthropic GitHub App (PR 자동화)

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

## 🔧 개발 & 테스트 워크플로우

```bash
# 개발 모드 설치
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
pip install -e .

# 품질 도구
ython -m pytest             # 테스트 실행
make lint && make test      # 린트 + 테스트
make build                  # 패키지 빌드
```

권장 자동화

- `python .moai/scripts/doc_sync.py` → 최신 문서/상태 리포트 생성
- `python .moai/scripts/checkpoint_watcher.py start` → 개인 모드 자동 체크포인트
- `moai update --check` → 템플릿/스크립트 최신 상태 확인

---

## 📚 문서 & 참고 자료

### 📖 온라인 문서 사이트 (SPEC-010 완료)
- **[MoAI-ADK Documentation](https://moai-adk.github.io)** - 완전 자동화된 온라인 문서
- **Getting Started**: 설치부터 첫 프로젝트까지 단계별 가이드
- **User Guide**: 4단계 워크플로우 상세 설명
- **API Reference**: 소스코드에서 자동 생성되는 완전한 API 문서
- **Development**: 기여 방법 및 아키텍처 가이드
- **Examples**: 실제 사용 예제 및 템플릿
- **Release Notes**: sync-report 기반 자동 생성 릴리스 노트

### 📄 로컬 문서
- [종합 개발 가이드](docs/MOAI-ADK-GUIDE.md)
- [Documentation Index](docs/sections/index.md)
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

**🗿 "명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."**

**MoAI-ADK** | **Made with ❤️ for Claude Code Community**
