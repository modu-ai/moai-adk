---
title: moai init 초기화
weight: 5
draft: false
---

`moai init` 은 현재 디렉터리 또는 새 폴더에 MoAI 프로젝트를 초기화합니다. Claude Code 연동에 필요한 `.claude/`, `.moai/` 구조와 설정을 배치하고, 필요 시 대화형 마법사로 프로젝트 모드·언어·품질 게이트를 설정합니다.

## 사용법

```bash
moai init [project-name]
```

| 패턴 | 동작 |
|------|------|
| `moai init <name>` | `./<name>/` 폴더를 만들고 그 안에 초기화 |
| `moai init .` | 현재 디렉터리에 초기화 |
| `moai init` | 현재 디렉터리에 초기화 (`moai init .` 과 동일) |

인자는 최대 1개까지 받습니다.

## 주요 플래그

### 배치 범위

| 플래그 | 설명 |
|--------|------|
| `--all` | 카탈로그 전체 배치 (core + 선택 팩 + 하네스 생성물). 기본값은 core-only slim 모드 |
| `--force` | 기존 프로젝트 재초기화 (현재 `.moai/` 를 백업) |
| `--no-hooks` | git 훅 설치 생략 |

### 프로젝트 기본값

| 플래그 | 설명 |
|--------|------|
| `--root <dir>` | 프로젝트 루트 (기본: 현재 디렉터리) |
| `--name <name>` | 프로젝트 이름 (기본: 디렉터리명) |
| `--language <lang>` | 주 프로그래밍 언어 |
| `--framework <name>` | 프레임워크 (기본: 자동 감지 또는 `none`) |
| `--mode <ddd\|tdd>` | 개발 방법론 (기본: tdd) |
| `--non-interactive` | 대화형 마법사 생략 — 플래그와 기본값만 사용 |

### 마법사 단계

| 플래그 | 설명 |
|--------|------|
| `--standard` | Phase 1 질문 제시 (프로젝트 모드, 하네스 프로필, LSP, 품질 게이트, 디자인) |
| `--advanced` | Phase 1 + Phase 2 질문 제시 (`--standard` 포함) |
| `--project-mode <personal\|team>` | 프로젝트 모드 (기본: personal) |
| `--harness-profile <name>` | 하네스 평가 프로필: default, strict, lenient, frontend |
| `--enable-lsp` | LSP 통합 활성화 (기본: false) |
| `--enforce-quality` | 품질 게이트 강제 (기본: true) |
| `--enable-design` | 디자인 워크플로우 활성화 (기본: true) |

### Git / 모델 정책

| 플래그 | 설명 |
|--------|------|
| `--git-mode <manual\|personal\|team>` | Git 워크플로우 모드 (기본: manual) |
| `--git-provider <github\|gitlab>` | Git 제공자 |
| `--github-username <name>` | GitHub 사용자명 (personal/team 모드 필수) |
| `--model-policy <max\|medium\|low>` | 성능 티어 — `llm.yaml` 의 `performance_tier` 에 저장 |
| `--plan-type <api\|subscription>` | 청구 플랜 유형 — `llm.yaml` 의 `plan_type` 에 저장 |

## 예시

```bash
# 새 폴더에 초기화
moai init my-app

# 현재 디렉터리에 초기화
moai init .

# 방법론 지정
moai init --mode tdd

# 카탈로그 전체 배치 (slim 모드 우회)
moai init --all

# 비대화형 (CI 등)
moai init . --non-interactive --language go
```

## 관련 명령어

| 명령어 | 설명 |
|--------|------|
| `moai update` | 초기화된 프로젝트의 템플릿 동기화 |
| `moai status` | 초기화 상태 확인 |
| `moai doctor` | 초기화 후 환경 검증 |

## 참고

- [프로젝트 상태](/ko/cli-reference/status)
- [업데이트](/ko/cli-reference/update)
- [CLI 개요](/ko/getting-started/cli)
