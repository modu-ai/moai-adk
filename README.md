# MoAI-ADK (MoAI Agentic Development Kit) v0.1.17

**Claude Code 표준 기반 Spec-First TDD 완전 자동화 개발 시스템**

[![Version](https://img.shields.io/badge/version-0.1.17-blue)](https://github.com/modu-ai/moai-adk/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.11%2B-blue)](https://www.python.org/)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-compatible-purple)](https://docs.anthropic.com/claude-code)

## 🗿 개요

MoAI-ADK는 Claude Code의 2025년 최신 기능을 활용하여 **SPECIFY → PLAN → TASKS → IMPLEMENT** 4단계 파이프라인을 통한 완전 자동화 개발 환경을 제공하는 혁신적인 개발 프레임워크입니다.

### 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **Living Document**: 문서와 코드는 항상 동기화
- **Full Traceability**: 16-Core TAG 시스템으로 완전 추적

## 🚀 빠른 시작

### 1. 설치

```bash
# PyPI에서 설치
pip install moai-adk

# 새 프로젝트 생성
moai init myproject && cd myproject
```

### 2. Claude Code에서 사용

```bash
# Claude Code 실행
claude

# 프로젝트 초기화
/moai:1-project init

# 첫 번째 기능 개발
/moai:2-spec user-auth "JWT 기반 사용자 인증 시스템"
/moai:3-plan SPEC-001
/moai:4-tasks PLAN-001
/moai:5-dev T001
```

## 🤖 주요 기능

### 4단계 자동화 파이프라인
1. **SPECIFY**: EARS 형식 명세 자동 작성
2. **PLAN**: Constitution Check & 아키텍처 설계
3. **TASKS**: TDD 기반 작업 분해
4. **IMPLEMENT**: Red-Green-Refactor 자동 구현

### 10개 전문 AI 에이전트
- **steering-architect**: 프로젝트 비전 설계
- **spec-manager**: EARS 형식 명세 작성
- **plan-architect**: Constitution Check 수행
- **task-decomposer**: TDD 작업 분해
- **code-generator**: 자동 구현
- **test-automator**: 품질 검증
- 그 외 6개 에이전트...

### Constitution 5원칙 자동 검증
1. **Simplicity**: 프로젝트 복잡도 ≤ 3개
2. **Architecture**: 모든 기능은 라이브러리로
3. **Testing**: RED-GREEN-REFACTOR 강제 (80% 커버리지)
4. **Observability**: 구조화된 로깅 필수
5. **Versioning**: MAJOR.MINOR.BUILD 체계

## 📂 프로젝트 구조

```
MoAI-ADK/                       # 프로젝트 루트
├── src/moai_adk/               # Python 패키지 소스
├── tests/                      # 테스트 코드
├── docs/                       # 문서
├── scripts/                    # 유틸리티 스크립트
├── .claude/                    # Claude Code 설정
│   ├── agents/moai/           # 10개 전문 에이전트
│   ├── commands/moai/         # 6개 슬래시 명령어
│   ├── hooks/moai/           # Python Hook 스크립트
│   ├── memory/               # 프로젝트 메모리 시스템
│   └── settings.json         # 권한 및 Hook 설정
├── .moai/                     # MoAI 프레임워크 설정
│   ├── config.json           # MoAI 설정
│   ├── indexes/              # 16-Core TAG 인덱스
│   └── memory/               # Constitution 메모리
├── .github/                   # GitHub CI/CD 워크플로우
├── pyproject.toml            # 패키지 설정
├── CLAUDE.md                 # Claude 프로젝트 메모리
└── README.md                 # 이 파일
```

## 🏷️ 16-Core TAG 시스템

완전한 추적성을 보장하는 태그 시스템:

### SPEC 카테고리 (문서 추적 - 필수)
- **@REQ**: 요구사항 정의
- **@SPEC**: 명세 문서(요약/식별자)
- **@DESIGN**: 설계 문서
- **@TASK**: 구현 작업

### STEERING 카테고리 (원칙 추적 - 필수)
- **@VISION**: 프로젝트 비전
- **@STRUCT**: 구조 설계
- **@TECH**: 기술 선택
- **@ADR**: 아키텍처 결정 기록

### IMPLEMENTATION 카테고리 (코드 추적 - 필수)
- **@FEATURE**: 기능 개발
- **@API**: API 설계 및 구현
- **@TEST**: 테스트 케이스
- **@DATA**: 데이터 모델링

### QUALITY 카테고리 (품질 추적 - 선택)
- **@PERF**: 성능 최적화
- **@SEC**: 보안 검토
- **@DEBT**: 기술 부채
- **@TODO**: 할 일 추적

## 🛠️ 명령어 참조

| 순서 | 명령어 | 담당 에이전트 | 기능 |
|------|--------|---------------|------|
| **1** | `/moai:1-project` | steering-architect | 프로젝트 설정 |
| **2** | `/moai:2-spec` | spec-manager | EARS 형식 명세 작성 |
| **3** | `/moai:3-plan` | plan-architect | Constitution Check |
| **4** | `/moai:4-tasks` | task-decomposer | TDD 작업 분해 |
| **5** | `/moai:5-dev` | code-generator + test-automator | 자동 구현 |
| **6** | `/moai:6-sync` | doc-syncer + tag-indexer | 문서 동기화 |

## 🔧 개발 및 빌드

### 개발 환경 설정
```bash
# 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 개발 모드 설치
pip install -e .

# 테스트 실행
python -m pytest tests/

# 코드 품질 검사
make lint
make test
make build
```

### 빌드 명령어
```bash
make build      # 패키지 빌드
make test       # 테스트 실행
make lint       # 코드 품질 검사
make clean      # 정리
make dev        # 개발 모드
```

## 🛡️ 품질 보장

### 자동화된 검증
- **Hook 시스템**: PreToolUse, PostToolUse, SessionStart
- **품질 게이트**: 각 단계별 자동 검증
- **Constitution 검증**: 5원칙 자동 확인
- **TAG 일관성**: 실시간 추적성 검증
- **GitHub CI/CD**: Constitution 5원칙 자동 검증 파이프라인

### 성과 지표
- **개발 생산성**: 명세 완성도 ≥90%, 구현 속도 50% 향상
- **품질 지표**: 테스트 커버리지 ≥80%, 버그 감소 70%
- **추적성**: TAG 정확성 ≥95%, 체인 완성도 ≥90%

## 📚 문서

- [📖 설치 가이드](docs/INSTALLATION.md)
- [🏗️ 아키텍처 문서](docs/ARCHITECTURE.md)
- [📋 API 문서](docs/API.md)
- [🔨 빌드 가이드](BUILD.md)
- [🤝 기여 가이드](docs/CONTRIBUTING.md)
- [🐛 트러블슈팅](docs/TROUBLESHOOTING.md)

## 🤝 기여하기

MoAI-ADK에 기여해주셔서 감사합니다!

1. **이슈 리포트**: 버그 발견 시 [이슈](https://github.com/modu-ai/moai-adk/issues) 등록
2. **기능 제안**: 새로운 기능 아이디어 제안
3. **코드 기여**: Fork → 개발 → Pull Request
4. **문서 개선**: 문서 오류 수정 및 내용 보완

자세한 내용은 [CONTRIBUTING.md](docs/CONTRIBUTING.md)를 참조하세요.

## 📝 라이선스

이 프로젝트는 [MIT 라이선스](LICENSE) 하에 배포됩니다.

## 📞 지원

- **공식 문서**: [docs/](docs/)
- **이슈 트래커**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **디스커션**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)

---

**🗿 "명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."**

**MoAI-ADK v0.1.17** | **Made with ❤️ for Claude Code Community**
