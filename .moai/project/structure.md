# MoAI-ADK 프로젝트 구조

> **최종 업데이트**: 2026-01-24
> **버전**: 4.1.0

---

## 저장소 개요

```
MoAI-ADK/
├── src/moai_adk/             # Python 패키지 소스 (~71,800 LOC)
├── tests/                    # 테스트 스위트 (150+ 테스트 파일)
├── docs/                     # 문서 사이트 (Nextra)
├── assets/                   # 이미지 및 리소스
├── .claude/                  # Claude Code 설정 (templates에서 동기화)
├── .moai/                    # MoAI-ADK 설정 (templates에서 동기화)
├── .github/                  # GitHub 워크플로우 및 스크립트
├── CLAUDE.md                 # Alfred 실행 지시문
├── CLAUDE.local.md           # 로컬 개발 가이드 (git-ignored)
├── pyproject.toml            # 패키지 설정 (버전의 단일 출처)
└── README.md                 # 프로젝트 문서화
```

---

## 패키지 구조 (src/moai_adk/)

```
src/moai_adk/
├── __init__.py               # 패키지 초기화
├── __main__.py               # CLI 진입점
├── version.py                # 버전 관리 (pyproject.toml 읽기)
│
├── astgrep/                  # AST-grep 통합
│   ├── checker.py            # 구문 검사기
│   ├── patterns.py           # 패턴 정의
│   └── scanner.py            # 보안 스캐너
│
├── cli/                      # CLI 명령 (Click 프레임워크)
│   ├── init.py               # 프로젝트 초기화
│   ├── worktree/             # Git worktree 관리
│   └── commands/             # CLI 하위명령
│
├── core/                     # 핵심 모듈 (50개 이상 파일)
│   ├── config.py             # 설정 관리
│   ├── spec_manager.py       # SPEC 문서 처리
│   ├── ddd_engine.py         # DDD 워크플로우
│   ├── analysis/             # 코드 분석 모듈
│   ├── integration/          # 통합 모듈
│   ├── migration/            # 마이그레이션 모듈
│   ├── project/              # 프로젝트 관리
│   └── statusline/           # 상태줄 모듈
│
├── foundation/               # 파운데이션 구성요소
│   ├── __init__.py           # 패키지 초기화
│   ├── claude.py             # Claude Code 통합
│   ├── core.py               # 핵심 원칙
│   ├── git/                  # Git 작업 모듈 (디렉토리 구조)
│   ├── quality.py            # 품질 프레임워크
│   └── testing.py            # DDD 프레임워크 유틸리티
│
├── loop/                     # 피드백 루프 시스템
│   ├── controller.py         # 루프 컨트롤러
│   ├── processor.py          # 오류 프로세서
│   └── hooks.py              # 후크 통합
│
├── lsp/                      # LSP (Language Server Protocol) 통합
│   ├── client.py             # LSP 클라이언트
│   ├── diagnostics.py        # 진단 핸들러
│   └── provider.py           # LSP 공급자
│
├── project/                  # 프로젝트 관리
│   ├── manager.py            # 프로젝트 매니저
│   ├── template.py           # 템플릿 처리
│   └── config.py             # 프로젝트 설정
│
├── ralph/                    # Ralph 엔진 (품질 자동화)
│   ├── engine.py             # 메인 엔진
│   ├── analyzer.py           # 오류 분석기
│   └── fixer.py              # 자동 수정기
│
├── statusline/               # Claude Code 상태줄
│   ├── display.py            # 상태 표시
│   ├── formatter.py          # 출력 포맷터
│   └── config.py             # 상태줄 설정
│
├── templates/                # 배포 템플릿
│   ├── .claude/              # Claude Code 템플릿
│   ├── .moai/                # MoAI 설정 템플릿
│   └── CLAUDE.md             # Alfred 지시문 템플릿
│
├── utils/                    # 유틸리티 모듈
│   ├── file_ops.py           # 파일 작업
│   ├── git_ops.py            # Git 작업
│   └── yaml_ops.py           # YAML 처리
```

---

## 설정 구조 (.claude/)

```
.claude/
├── settings.json             # Claude Code 설정
├── settings.local.json       # 로컬 설정 (git-ignored)
│
├── agents/                   # 하위 에이전트 정의
│   └── moai/                 # MoAI 에이전트 (20개)
│       ├── expert-*.md       # 전문가 에이전트 (9개)
│       ├── manager-*.md      # 매니저 에이전트 (7개)
│       └── builder-*.md      # 빌더 에이전트 (4개)
│
├── commands/                 # 슬래시 명령
│   └── moai/                 # MoAI 명령
│       ├── 0-project.md      # 프로젝트 초기화
│       ├── 1-plan.md         # SPEC 계획
│       ├── 2-run.md          # DDD 실행
│       ├── 3-sync.md         # 문서화 동기화
│       ├── 9-feedback.md     # 피드백 제출
│       ├── alfred.md         # Alfred 자동화
│       ├── fix.md            # 자동 수정
│       └── loop.md           # 피드백 루프
│
├── hooks/                    # 자동화 후크
│   └── moai/                 # MoAI 후크
│       ├── session_start_*.py
│       ├── quality_gate_with_lsp.py
│       └── lib/              # 후크 라이브러리
│
├── output-styles/            # 출력 포맷팅
│   └── moai/
│       ├── r2d2.md           # R2-D2 스타일
│       └── yoda.md           # Yoda 스타일
│
└── skills/                   # 도메인 지식
    └── moai/                 # MoAI 스킬 (48개)
        ├── moai-foundation-*.md   # 파운데이션 스킬
        ├── moai-domain-*.md       # 도메인 스킬
        ├── moai-lang-*.md         # 언어 스킬
        ├── moai-platform-*.md     # 플랫폼 스킬
        ├── moai-workflow-*.md     # 워크플로우 스킬
        └── moai-library-*.md      # 라이브러리 스킬
```

---

## MoAI 설정 구조 (.moai/)

```
.moai/
├── config/                   # 설정 파일
│   ├── config.yaml           # 메인 설정
│   ├── statusline-config.yaml
│   ├── multilingual-triggers.yaml
│   ├── sections/             # 모듈형 섹션
│   │   ├── user.yaml
│   │   ├── language.yaml
│   │   ├── project.yaml
│   │   ├── git-strategy.yaml
│   │   ├── quality.yaml
│   │   ├── system.yaml
│   │   └── ralph.yaml
│   └── questions/            # 설정 UI
│       ├── _schema.yaml
│       ├── tab0-init.yaml
│       ├── tab1-user.yaml
│       ├── tab2-project.yaml
│       ├── tab3-git.yaml
│       ├── tab4-quality.yaml
│       └── tab5-system.yaml
│
├── specs/                    # SPEC 문서
│   └── SPEC-*.md             # 개별 명세서
│
├── project/                  # 프로젝트 문서
│   ├── product.md            # 제품 설명
│   ├── structure.md          # 프로젝트 구조 (이 파일)
│   └── tech.md               # 기술 스택
│
├── memory/                   # 세션 메모리 (런타임)
│   └── checkpoints/          # 체크포인트 데이터
├── cache/                    # 캐시 데이터 (런타임)
├── logs/                     # 로그 파일 (런타임)
├── error_logs/               # 오류 로그 (런타임)
├── rollbacks/                # 롤백 데이터 (런타임)
├── analytics/                # 사용 분석 (런타임)
└── web/                      # 웹 UI 데이터 (런타임)
```

---

## 테스트 구조 (tests/)

```
tests/
├── unit/                     # 단위 테스트
│   ├── test_core/            # 핵심 모듈 테스트
│   ├── test_cli/             # CLI 명령 테스트
│   ├── test_utils/           # 유틸리티 테스트
│   ├── test_statusline/      # 상태줄 테스트
│   └── test_hooks/           # 후크 테스트
│
├── astgrep/                  # AST-grep 통합 테스트
│   ├── test_analyzer_edge_cases.py
│   ├── test_models_edge_cases.py
│   ├── test_rules_advanced.py
│   └── test_*.py
│
├── cli/                      # CLI 테스트
│   ├── commands/             # CLI 명령 테스트
│   ├── prompts/              # 프롬프트 테스트
│   ├── ui/                   # UI 테스트
│   ├── worktree/             # 워크트리 테스트
│   └── test_*.py
│
├── foundation/               # 파운데이션 테스트
│   ├── test_backend.py
│   ├── test_commit_templates.py
│   ├── test_database.py
│   ├── test_devops.py
│   ├── test_frontend.py
│   ├── test_git.py
│   ├── test_ears_tdd.py
│   ├── test_langs_tdd.py
│   ├── test_testing_tdd.py
│   └── trust/                # TRUST 5 프레임워크 테스트
│
├── integration/              # 통합 테스트
│   ├── cli/                  # CLI 통합 테스트
│   ├── test_core/            # 핵심 통합 테스트
│   └── test_workflow/        # 워크플로우 테스트
│
├── conftest.py               # Pytest 설정
└── fixtures/                 # 테스트 픽스처
```

---

## 문서화 구조 (docs/)

```
docs/
├── app/                      # Nextra 문서화
│   ├── agents/               # 에이전트 문서
│   ├── commands/             # 명령 문서
│   ├── skills/               # 스킬 문서
│   └── workflows/            # 워크플로우 가이드
│
├── public/                   # 정적 자산
└── next.config.mjs           # Next.js 설정
```

---

## 핵심 파일

| 파일 | 목적 |
|------|---------|
| `pyproject.toml` | 패키지 버전 및 의존성의 단일 출처 |
| `CLAUDE.md` | Alfred 실행 지시문 (Mr. Alfred 오케스트레이션 규칙) |
| `CLAUDE.local.md` | 로컬 개발 가이드 (git-ignored) |
| `.moai/config/config.yaml` | 메인 MoAI 설정 |
| `.moai/config/sections/quality.yaml` | TRUST 5 및 LSP 품질 게이트 설정 |
| `src/moai_adk/version.py` | 버전 리더 (pyproject.toml에서 읽기) |

---

## 동기화 메커니즘

템플릿은 `src/moai_adk/templates/`에서 루트 디렉토리로 흐릅니다:

```
src/moai_adk/templates/.claude/  →  .claude/
src/moai_adk/templates/.moai/    →  .moai/
src/moai_adk/templates/CLAUDE.md →  ./CLAUDE.md
```

로컬 전용 파일 (동기화되지 않음):
- `.claude/settings.local.json`
- `CLAUDE.local.md`
- `.moai/cache/`, `.moai/memory/`, `.moai/logs/`
- `.moai/project/`, `.moai/specs/` (사용자 데이터 보호)
