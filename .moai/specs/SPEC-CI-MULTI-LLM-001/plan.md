---
spec_id: SPEC-CI-MULTI-LLM-001
version: 1.0.0
status: Planned
created: 2026-04-27
updated: 2026-04-27
author: manager-strategy
development_mode: tdd
harness_level: standard
---

# SPEC-CI-MULTI-LLM-001 — Implementation Plan

## 1. Architecture Overview

### 1.1 Package Structure (Alternative B — spec.md §4.5 alignment)

```
internal/cli/
├── github.go                  # 기존: root cobra command (parse-issue, link-spec)
├── github_init.go             # REQ-CI-002: 통합 부트스트랩
├── github_runner.go           # REQ-CI-003~005: runner 서브커맨드
├── github_auth.go             # REQ-CI-008~011: auth 서브커맨드
├── github_workflow.go         # REQ-CI-016: workflow 서브커맨드
├── github_status.go           # REQ-CI-017: 통합 상태 조회
├── github_test.go             # 모든 github_* 테스트 (table-driven)
├── doctor.go                  # 기존: runner version check 통합 (REQ-CI-005)
└── hook.go                    # 기존: SessionStart hook 확장

internal/github/               # 기존 패키지 (gh.go, issue_parser 등 보존)
├── gh.go                      # 기존: RunGH / RunGHOutput 재사용
├── secret.go                  # 신규: gh secret CRUD 래퍼 (REQ-SEC-002)
├── runner/
│   ├── installer.go           # REQ-CI-003: 다운로드 + SHA256 검증
│   ├── registrar.go           # REQ-CI-004: GitHub 등록 (gh API)
│   ├── service.go             # REQ-CI-003.4: launchd 관리 (systemd stub)
│   └── version.go             # REQ-CI-005: 30일 폐기 윈도우 검사
├── auth/
│   ├── claude.go              # REQ-CI-008
│   ├── codex.go               # REQ-CI-009 + REQ-CI-007 private guard
│   ├── gemini.go              # REQ-CI-010
│   └── glm.go                 # REQ-CI-011 + SPEC-GLM-001 정합
└── workflow/
    ├── deployer.go            # REQ-CI-016, CI-019: 템플릿 동기화
    └── validator.go           # REQ-CI-018, CI-021, SEC-003, SEC-005: 템플릿 검증

internal/template/templates/
├── .github/
│   ├── workflows/
│   │   ├── claude.yml.tmpl           # REQ-CI-014 (Issue trigger)
│   │   ├── claude-code-review.yml.tmpl # REQ-CI-014 (PR trigger)
│   │   ├── codex-review.yml.tmpl     # REQ-CI-007, SEC-001, SEC-006
│   │   ├── gemini-review.yml.tmpl    # REQ-CI-006, CI-014
│   │   ├── glm-review.yml.tmpl       # REQ-CI-014
│   │   └── llm-panel.yml.tmpl        # REQ-CI-013
│   └── actions/
│       ├── detect-language/
│       │   └── action.yml.tmpl       # REQ-CI-012
│       ├── setup-glm-env/
│       │   └── action.yml.tmpl       # REQ-CI-011
│       └── codex-bootstrap/
│           └── action.yml.tmpl       # REQ-CI-009, CI-021, SEC-006
└── .moai/config/sections/
    └── github-actions.yaml.tmpl      # REQ-CI-015, CI-020
```

### 1.2 Key Architecture Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | `internal/github/` 하위 sub-package 분리 (runner/auth/workflow) | 단일 패키지 과부하 방지, spec.md §4.5 정합 |
| D2 | Go template (`.yml.tmpl`)으로 워크플로우 렌더링 | SHA 주입, runner_mode 분기, LLM 선택 동적 지원 |
| D3 | Thin cobra wrapper (CLI → domain) 패턴 | 기존 update.go → internal/template/ 패턴 일치 |
| D4 | 기존 `internal/github/gh.go` RunGH/RunGHOutput 재사용 | 중복 방지, 테스트 인프라 재활용 |
| D5 | Auth 토큰 stdin pipe (디스크 기록 금지) | REQ-SEC-002 준수, `exec.Command.StdinPipe()` |
| D6 | macOS arm64 only v1.0 (OQ1) | systemd stub은 T-04에서 interface만 정의 |

### 1.3 Existing Code Reuse

| Component | Location | Reuse Pattern |
|-----------|----------|---------------|
| GH client | `internal/github/gh.go` | `RunGH()`, `RunGHOutput()` for all `gh` CLI calls |
| Template deploy | `internal/template/` | `Deploy()` API for workflow sync |
| Doctor framework | `internal/cli/doctor.go` | `DiagnosticCheck` struct for runner version |
| Cobra root | `internal/cli/root.go` | Command groups, `rootCmd.AddCommand()` |
| Config system | `.moai/config/sections/` | YAML section pattern for github-actions.yaml |

---

## 2. Task Decomposition (30 Tasks, 4 Waves)

### Wave 1 — CLI Skeleton + Runner Domain (M1)

#### T-01: GitHub Cobra Command Group Extension

- **Description**: 기존 `internal/cli/github.go`의 `githubCmd`에 새 하위 명령 그룹(init, runner, auth, workflow, status)을 등록하는 cobra 트리 골격 생성
- **REQ**: REQ-CI-001, CI-001.1, CI-001.2, CI-001.3
- **Dependencies**: None (Wave 1 선행 작업)
- **Planned Files**:
  - `internal/cli/github.go` (수정: init()에 새 서브커맨드 그룹 추가)
  - `internal/cli/github_test.go` (신규: 명령 트리 구조 테스트)
- **Acceptance Criteria**:
  - `moai github --help`에 init, runner, auth, workflow, status 하위 명령 표시
  - `moai github init --help` 출력 확인
  - `--dry-run` 플래그가 루트 명령에 존재
  - 기존 parse-issue, link-spec 명령 동작 보존

#### T-02: Runner Installer Domain

- **Description**: `internal/github/runner/installer.go` — GitHub Release API에서 최신 runner 다운로드, SHA256 검증, 압축 해제 로직
- **REQ**: REQ-CI-003, CI-003.1 (3회 재시도), CI-003.2 (덮어쓰기 확인), CI-003.3 (SHA256 불일치 중단)
- **Dependencies**: T-01 (cobra 트리 필요)
- **Planned Files**:
  - `internal/github/runner/installer.go`
  - `internal/github/runner/installer_test.go`
- **Acceptance Criteria**:
  - `DownloadRunner(ctx, os, arch)` 성공 시 `~/actions-runner/`에 바이너리 존재
  - SHA256 불일치 시 다운로드 파일 삭제 + error 반환
  - 3회 재시도 후 최종 실패 시 명확한 error message
  - `~/actions-runner/` 존재 시 사용자 확인 로직 (interface로 mock 가능)

#### T-03: Runner Registrar Domain

- **Description**: `internal/github/runner/registrar.go` — `gh api`로 등록 토큰 발급, `config.sh --replace` 실행, labels 기록
- **REQ**: REQ-CI-004, CI-004.1 (--replace 기본), CI-004.2 (토큰 메모리만), CI-004.3 (Settings 링크 출력)
- **Dependencies**: T-02 (runner 설치 선행)
- **Planned Files**:
  - `internal/github/runner/registrar.go`
  - `internal/github/runner/registrar_test.go`
- **Acceptance Criteria**:
  - 등록 토큰이 디스크에 기록되지 않음 (테스트로 검증)
  - `--replace` 플래그가 config.sh 호출에 포함
  - 등록 성공 후 GitHub Settings URL 출력

#### T-04: Runner Service Manager (launchd + systemd stub)

- **Description**: `internal/github/runner/service.go` — macOS launchd 서비스 install/start/stop, Linux systemd는 stub interface만 정의 (v1.1 구현)
- **REQ**: REQ-CI-003.4, CI-005.3
- **Dependencies**: T-03
- **Planned Files**:
  - `internal/github/runner/service.go`
  - `internal/github/runner/service_test.go`
- **Acceptance Criteria**:
  - macOS에서 `svc.sh install/start/stop` 실행
  - launchd plist 경로 출력
  - Linux에서 "not yet supported" error 반환
  - ServiceManager interface 정의로 v1.1 확장 가능

#### T-05: Runner Version Checker

- **Description**: `internal/github/runner/version.go` — 설치된 runner 버전 조회, GitHub Release API latest 비교, 25일/30일 경과 판정
- **REQ**: REQ-CI-005, CI-005.1 (25일 WARN / 30일 ERROR), CI-005.2
- **Dependencies**: T-02
- **Planned Files**:
  - `internal/github/runner/version.go`
  - `internal/github/runner/version_test.go`
- **Acceptance Criteria**:
  - `CheckVersion(ctx)` → (daysOld int, status CheckStatus, message string)
  - 25일 초과: WARN, 30일 초과: FAIL
  - runner 미설치 시 SKIP 상태 반환

#### T-06: Runner CLI Subcommands (Thin Cobra Wiring)

- **Description**: `internal/cli/github_runner.go` — install/register/start/stop/status/upgrade 서브커맨드 cobra 등록. 도메인 로직은 T-02~T-05의 패키지 호출만
- **REQ**: REQ-CI-001 (runner subtree)
- **Dependencies**: T-01, T-02, T-03, T-04, T-05
- **Planned Files**:
  - `internal/cli/github_runner.go`
  - `internal/cli/github_runner_test.go`
- **Acceptance Criteria**:
  - `moai github runner --help`에 6개 하위 명령 표시
  - `moai github runner install --dry-run` 실행 계획 출력
  - `moai github runner status`가 T-05 version check 통합

---

### Wave 2 — LLM Auth Bootstrap (M2)

#### T-07: Secret Manager (gh secret Wrapper)

- **Description**: `internal/github/secret.go` — `gh secret set/get/list` 래퍼. 토큰은 stdin pipe로 전달, 디스크 기록 없음
- **REQ**: REQ-SEC-002
- **Dependencies**: T-01
- **Planned Files**:
  - `internal/github/secret.go`
  - `internal/github/secret_test.go`
- **Acceptance Criteria**:
  - `SetSecret(ctx, name, value)` → `gh secret set NAME` stdin pipe
  - 토큰이 임시 파일이나 환경변수에 노출되지 않음
  - 디버그 로그에 `***` 마스킹
  - `ListSecrets(ctx)` → 등록된 secret 이름 목록 반환

#### T-08: Claude Auth Handler

- **Description**: `internal/github/auth/claude.go` — Claude CLI 설치 확인, setup-token 안내, `gh secret set CLAUDE_CODE_OAUTH_TOKEN`
- **REQ**: REQ-CI-008, CI-008.1 (메모리만), CI-008.2 (Max plan 안내)
- **Dependencies**: T-07
- **Planned Files**:
  - `internal/github/auth/claude.go`
  - `internal/github/auth/claude_test.go`
- **Acceptance Criteria**:
  - 토큰이 파일에 기록되지 않음
  - claude CLI 미설치 시 설치 안내 메시지
  - 등록 후 Max plan 구독 필수 안내 출력

#### T-09: Codex Auth Handler (Private Guard)

- **Description**: `internal/github/auth/codex.go` — private repo 검증, codex CLI 설치, auth.json → gh secret set
- **REQ**: REQ-CI-009, CI-009.1 (auth.json 600 권한), CI-009.2 (조건부 시드), REQ-CI-007, CI-007.1, CI-007.2, REQ-SEC-001
- **Dependencies**: T-07
- **Planned Files**:
  - `internal/github/auth/codex.go`
  - `internal/github/auth/codex_test.go`
- **Acceptance Criteria**:
  - public repo에서 실행 시 즉시 중단 + OpenAI 정책 링크 출력
  - `--force-public` 플래그로도 Codex 차단 (HARD)
  - auth.json 권한 600 검증
  - 조건부 시드 패턴 템플릿 메타데이터 생성

#### T-10: Gemini Auth Handler

- **Description**: `internal/github/auth/gemini.go` — API key 입력 (masked), 형식 검증, gh secret set
- **REQ**: REQ-CI-010, CI-010.1 (마스킹), CI-010.2 (무료 tier 안내)
- **Dependencies**: T-07
- **Planned Files**:
  - `internal/github/auth/gemini.go`
  - `internal/github/auth/gemini_test.go`
- **Acceptance Criteria**:
  - 입력 키 마스킹 (첫 글자 + 마지막 4글자)
  - 형식 검증 (영숫자 + `-_`, 길이 39자)
  - 무료 tier 한도 안내 출력

#### T-11: GLM Auth Handler

- **Description**: `internal/github/auth/glm.go` — GLM auth token 입력, gh secret set, SPEC-GLM-001 환경변수 4종 주입 메타데이터
- **REQ**: REQ-CI-011, CI-011.1, CI-011.2 (DISABLE_BETAS 등)
- **Dependencies**: T-07
- **Planned Files**:
  - `internal/github/auth/glm.go`
  - `internal/github/auth/glm_test.go`
- **Acceptance Criteria**:
  - GLM_AUTH_TOKEN stdin 입력 (masked)
  - 환경변수 4종 (`ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, `DISABLE_BETAS`, `DISABLE_PROMPT_CACHING`) 메타데이터 생성
  - SPEC-GLM-001 정합성 (값/키 일치)

---

### Wave 3 — Workflow Templates + Composite Actions (M3)

#### T-12: detect-language Composite Action

- **Description**: `internal/template/templates/.github/actions/detect-language/action.yml.tmpl` — 16개 언어 project_markers 자동 감지
- **REQ**: REQ-CI-012, CI-012.2, CI-012.3
- **Dependencies**: T-01 (cobra 트리 필요하지만 독립 템플릿이므로 사실상 무의존)
- **Planned Files**:
  - `internal/template/templates/.github/actions/detect-language/action.yml.tmpl`
  - `internal/github/workflow/detect_language_test.go` (감지 로직 검증)
- **Acceptance Criteria**:
  - 16개 project_markers (go.mod, package.json, pyproject.toml, Cargo.toml, pom.xml, build.gradle, *.sln, Gemfile, composer.json, mix.exs, CMakeLists.txt, build.sbt, *.Rproj, pubspec.yaml, Package.swift, meson.build) 모두 인식
  - 감지 결과를 `outputs.languages` JSON으로 출력
  - 하나도 감지 안될 시 graceful skip
  - 특정 언어를 PRIMARY로 지정하지 않음

#### T-13: setup-glm-env Composite Action

- **Description**: `internal/template/templates/.github/actions/setup-glm-env/action.yml.tmpl` — GLM 환경변수 4종 조건부 주입
- **REQ**: REQ-CI-011.1, CI-011.2
- **Dependencies**: T-11
- **Planned Files**:
  - `internal/template/templates/.github/actions/setup-glm-env/action.yml.tmpl`
- **Acceptance Criteria**:
  - `GLM_AUTH_TOKEN` secret에서 읽어 `~/.moai/.env.glm` 조건부 생성
  - 환경변수 4종 주입 (SPEC-GLM-001 정합)
  - 매 실행 시 재생성 (REQ-SEC-005.2)

#### T-14: codex-bootstrap Composite Action

- **Description**: `internal/template/templates/.github/actions/codex-bootstrap/action.yml.tmpl` — auth.json 조건부 시드 + drop-sudo
- **REQ**: REQ-CI-009.1, CI-021 (조건부 시드), REQ-SEC-006 (drop-sudo)
- **Dependencies**: T-09
- **Planned Files**:
  - `internal/template/templates/.github/actions/codex-bootstrap/action.yml.tmpl`
- **Acceptance Criteria**:
  - `if [ ! -f "$CODEX_HOME/auth.json" ]; then ... fi` 가드 포함
  - `--safety-strategy drop-sudo` 플래그 포함
  - 무조건 덮어쓰기 패턴 미포함 (validator로 검증)

#### T-15: claude.yml.tmpl (Issue Trigger)

- **Description**: Issue 이벤트 기반 Claude 워크플로우 템플릿
- **REQ**: REQ-CI-014 (@claude Issue trigger)
- **Dependencies**: T-12 (detect-language action)
- **Planned Files**:
  - `internal/template/templates/.github/workflows/claude.yml.tmpl`
- **Acceptance Criteria**:
  - `on: issues: types: [opened, edited]` + comment trigger
  - `author_association` guard (REQ-SEC-010)
  - 코멘트 명령 인자 prompt 전달 (CI-014.3)
  - 401 응답 시 명확한 error + rotate 안내 (REQ-SEC-009)
  - workspace cleanup step (REQ-SEC-005)

#### T-16: claude-code-review.yml.tmpl (PR Trigger)

- **Description**: PR 코멘트 + PR open 트리거 Claude 코드 리뷰 워크플로우
- **REQ**: REQ-CI-014 (@claude PR trigger), CI-018 (SHA-pin), CI-018.1 (broken range 회피)
- **Dependencies**: T-12
- **Planned Files**:
  - `internal/template/templates/.github/workflows/claude-code-review.yml.tmpl`
- **Acceptance Criteria**:
  - Claude Action SHA-pin (v1.0.72 또는 v1.0.85+)
  - v1.0.79-84 범위 미사용 (validator로 검증)
  - `anthropics/claude-code-action` SHA 주입 가능 (`workflows.claude.action_sha`)
  - `author_association` guard
  - workspace cleanup step

#### T-17: codex-review.yml.tmpl

- **Description**: Codex 코드 리뷰 워크플로우 (self-hosted 필수, private repo guard)
- **REQ**: REQ-CI-007, REQ-SEC-001, REQ-SEC-006 (drop-sudo), REQ-SEC-010
- **Dependencies**: T-12, T-14
- **Planned Files**:
  - `internal/template/templates/.github/workflows/codex-review.yml.tmpl`
- **Acceptance Criteria**:
  - `if: github.event.repository.private == true` job-level 가드
  - self-hosted runner 전용 (`runs-on: [self-hosted, ...]`)
  - codex-bootstrap composite action 사용
  - `--safety-strategy drop-sudo` 포함
  - `author_association` guard
  - workspace cleanup step

#### T-18: gemini-review.yml.tmpl

- **Description**: Gemini 코드 리뷰 워크플로우 (github-hosted 가능, runner_mode fallback)
- **REQ**: REQ-CI-014, CI-006 (github-hosted fallback), CI-006.1 (비용 경고)
- **Dependencies**: T-12
- **Planned Files**:
  - `internal/template/templates/.github/workflows/gemini-review.yml.tmpl`
- **Acceptance Criteria**:
  - `google-github-actions/run-gemini-cli` SHA-pin
  - runner_mode에 따라 `runs-on` 분기 (template variable)
  - github-hosted 모드 시 macOS 비용 경고
  - `author_association` guard
  - workspace cleanup step

#### T-19: glm-review.yml.tmpl

- **Description**: GLM 코드 리뷰 워크플로우 (self-hosted, Anthropic-compat shim)
- **REQ**: REQ-CI-014
- **Dependencies**: T-12, T-13
- **Planned Files**:
  - `internal/template/templates/.github/workflows/glm-review.yml.tmpl`
- **Acceptance Criteria**:
  - `moai glm -- claude -p ...` 호출 패턴
  - setup-glm-env composite action 사용
  - self-hosted runner 전용
  - `author_association` guard
  - `continue-on-error: true` (REQ R10: GLM 장애 시 PR 블로킹 방지)
  - workspace cleanup step

#### T-20: llm-panel.yml.tmpl (PR Open Auto Panel)

- **Description**: PR open 시 다중 LLM 병렬 리뷰 + 단일 코멘트 aggregator
- **REQ**: REQ-CI-013, CI-013.1~013.4, REQ-CI-015 (사용자 정의 세트)
- **Dependencies**: T-15, T-16, T-17, T-18, T-19 (모든 단일 LLM 템플릿 완료 후)
- **Planned Files**:
  - `internal/template/templates/.github/workflows/llm-panel.yml.tmpl`
- **Acceptance Criteria**:
  - `on: pull_request: types: [opened, ready_for_review, reopened]`
  - 기본 패널 `[claude, gemini]`
  - 독립 concurrency group per LLM (`cancel-in-progress: true`)
  - `paths-ignore: [.github/**]`
  - 단일 코멘트 LLM별 섹션 분리 (`### Claude Review`, `### Gemini Review`)
  - `author_association` guard (all jobs)

#### T-21: github-actions.yaml.tmpl (Config Template)

- **Description**: `.moai/config/sections/github-actions.yaml.tmpl` — runner_mode, LLM 설정, auto_panel, comment_router
- **REQ**: REQ-CI-015, CI-020 (runner_mode state-driven rendering), CI-020.1 (hybrid per-LLM)
- **Dependencies**: T-01
- **Planned Files**:
  - `internal/template/templates/.moai/config/sections/github-actions.yaml.tmpl`
- **Acceptance Criteria**:
  - spec.md §4.4 config 구조와 완전 일치
  - `runner_mode: self-hosted | github-hosted | hybrid` 지원
  - per-LLM `runner_mode` override (hybrid)
  - `auto_panel.llms` 배열로 패널 세트 정의

#### T-22: Template Validator

- **Description**: `internal/github/workflow/validator.go` — SHA-pin 검증, Codex private guard 탐지, workspace cleanup 포함 확인, 무조건 덮어쓰기 패턴 탐지
- **REQ**: REQ-CI-018, CI-018.1, REQ-CI-021, REQ-SEC-003, REQ-SEC-005
- **Dependencies**: T-15~T-19 (검증 대상 템플릿)
- **Planned Files**:
  - `internal/github/workflow/validator.go`
  - `internal/github/workflow/validator_test.go`
- **Acceptance Criteria**:
  - floating tag (`@v1`, `@latest`) 탐지 시 error
  - Claude Action v1.0.79-84 SHA 탐지 시 error
  - Codex 워크플로우에 `private == true` 가드 미포함 시 error
  - 무조건 auth.json 덮어쓰기 패턴 탐지 시 error
  - workspace cleanup step 미포함 시 warn
  - `ValidateTemplate(ctx, filePath)` → (errors []string, warnings []string)

---

### Wave 4 — Integration + Docs (M4)

#### T-23: GitHub Init Command (Integrated Bootstrap)

- **Description**: `internal/cli/github_init.go` — 대화형 LLM 선택 → auth 자동 호출 → runner 설치 → workflow 동기화 → config 생성
- **REQ**: REQ-CI-002, CI-002.1~002.3
- **Dependencies**: T-06, T-07~T-11, T-21
- **Planned Files**:
  - `internal/cli/github_init.go`
  - `internal/cli/github_init_test.go`
- **Acceptance Criteria**:
  - 대화형 LLM 선택 (4종 중 다중 선택)
  - `--llm claude,gemini` 비대화형 플래그 지원
  - 선택하지 않은 LLM 워크플로우 미배포
  - 동기화 결과 요약 출력 (배포 파일 + secret 이름)
  - 기존 파일 충돌 시 덮어쓰기/merge/skip 선택

#### T-24: GitHub Auth Command (Thin Wrapper)

- **Description**: `internal/cli/github_auth.go` — auth 서브커맨드 cobra 등록, 각 LLM auth handler 호출
- **REQ**: REQ-CI-008~011
- **Dependencies**: T-08~T-11
- **Planned Files**:
  - `internal/cli/github_auth.go`
  - `internal/cli/github_auth_test.go`
- **Acceptance Criteria**:
  - `moai github auth claude/codex/gemini/glm` 4개 하위 명령
  - `--dry-run` 지원
  - 각 명령이 해당 auth handler 호출

#### T-25: GitHub Workflow Command

- **Description**: `internal/cli/github_workflow.go` — workflow add/remove/list 서브커맨드
- **REQ**: REQ-CI-016
- **Dependencies**: T-22 (validator), T-21 (config)
- **Planned Files**:
  - `internal/cli/github_workflow.go`
  - `internal/cli/github_workflow_test.go`
- **Acceptance Criteria**:
  - `moai github workflow add <llm>` → 템플릿 동기화
  - `moai github workflow remove <llm>` → 워크플로우 삭제
  - `moai github workflow list` → 배포된 워크플로우 목록

#### T-26: GitHub Status Command

- **Description**: `internal/cli/github_status.go` — runner 목록, 워크플로우 상태, secret 목록, 회전 주기 안내
- **REQ**: REQ-CI-017, REQ-SEC-004
- **Dependencies**: T-05, T-07, T-22
- **Planned Files**:
  - `internal/cli/github_status.go`
  - `internal/cli/github_status_test.go`
- **Acceptance Criteria**:
  - 등록된 runner 목록 (이름, 상태, 버전, 마지막 실행)
  - 활성화된 워크플로우 목록 (파일명, LLM, 트리거 모드)
  - 등록된 GitHub secret 목록 (이름만, 값 마스킹)
  - runner 30일 폐기 잔여일 표시
  - 각 secret 권장 회전 주기 표시

#### T-27: Doctor Integration (Runner Version Check)

- **Description**: 기존 `internal/cli/doctor.go`에 runner version check 진단 항목 추가
- **REQ**: REQ-CI-005 (doctor 경고)
- **Dependencies**: T-05
- **Planned Files**:
  - `internal/cli/doctor.go` (수정: `runDiagnosticChecks`에 runner check 추가)
  - `internal/cli/doctor_test.go` (수정: runner check 테스트)
- **Acceptance Criteria**:
  - `moai doctor` 실행 시 runner 버전 항목 표시
  - 25일 초과: WARN, 30일 초과: FAIL
  - runner 미설치: SKIP

#### T-28: SessionStart Hook Integration

- **Description**: 기존 SessionStart 훅에 runner version check 통합 (1일 1회 알림)
- **REQ**: REQ-CI-005 (SessionStart 알림)
- **Dependencies**: T-05
- **Planned Files**:
  - `internal/cli/hook.go` (수정: SessionStart handler에 runner check 추가)
  - `internal/cli/hook_test.go` (수정)
- **Acceptance Criteria**:
  - SessionStart 시 runner 버전 check 실행
  - 1일 1회 캐시 (중복 알림 방지)
  - 알림 내용: 버전, 경과일, 업그레이드 명령어 안내

#### T-29: Network Egress Validation + Audit Log Setup

- **Description**: Runner 설치 후 outbound HTTPS 검증 + `.moai/logs/` audit log 보존 설정
- **REQ**: REQ-SEC-007, CI-007.1 (방화벽 안내), CI-007.2 (미사용 LLM 생략), REQ-SEC-008
- **Dependencies**: T-02, T-11
- **Planned Files**:
  - `internal/github/runner/egress.go`
  - `internal/github/runner/egress_test.go`
- **Acceptance Criteria**:
  - 6개 도메인 outbound HTTPS 검증 (api.github.com, api.anthropic.com, api.openai.com, generativelanguage.googleapis.com, api.z.ai, codecov.io)
  - 미사용 LLM 도메인 생략
  - 실패 시 방화벽 설정 안내 메시지
  - `.moai/logs/github-cli-{date}.log` audit log 기록 (30일 retention)

#### T-30: docs-site 4-Locale Guides

- **Description**: docs-site에 multi-LLM CI 설정 가이드 4개국어 (ko/en/ja/zh) 작성
- **REQ**: REQ-CI-012 (16-lang docs), CLAUDE.local.md §17
- **Dependencies**: T-23 (init command 완료 후 문서화)
- **Planned Files**:
  - `docs-site/content/ko/guides/multi-llm-ci.md`
  - `docs-site/content/en/guides/multi-llm-ci.md`
  - `docs-site/content/ja/guides/multi-llm-ci.md`
  - `docs-site/content/zh/guides/multi-llm-ci.md`
- **Acceptance Criteria**:
  - 4개 locale 동시 업데이트 (§17.3 HARD)
  - Mermaid 방향 TD only (§17.2)
  - 금지 URL 미포함 (§17.1)
  - ko canonical → en → zh/ja 병렬 번역 체인

---

## 3. Risk Register Mapping

| Risk ID | Risk | Mapped Tasks | Mitigation |
|---------|------|-------------|------------|
| R1 | Codex auth.json 손상 | T-14, T-17, T-22 | 조건부 시드 + concurrency group |
| R2 | Claude Action SDK regression | T-16, T-22 | SHA-pin + broken range validator |
| R3 | Runner 30일 폐기 | T-05, T-27, T-28 | 3-point detection (status/doctor/hook) |
| R4 | 단일 머신 SPOF | T-18, T-21 | github-hosted fallback + hybrid mode |
| R5 | Network egress 차단 | T-29 | 자동 검증 + 방화벽 안내 |
| R6 | 구독 계정 정지 | T-26 (status 안내) | REQ-SEC-009 error handling in templates |
| R7 | Public 기여자 abuse | T-15~T-20 | author_association guard (all templates) |
| R8 | Codex auth.json 동시 실행 | T-17 | concurrency group per LLM job |
| R9 | Gemini 무료 tier 한도 | T-10 (안내) | CLI 안내 메시지 |
| R10 | GLM 엔드포인트 장애 | T-19 | continue-on-error: true |
| R11 | OAuth 토큰 만료 | T-15~T-19 | 401 error → rotate 안내 step |

---

## 4. HARD Constraints Compliance

| # | Constraint | Source | Task Coverage |
|---|-----------|--------|---------------|
| C1 | Template-First (모든 workflow → templates/) | REQ-CI-019 | T-12~T-22 |
| C2 | 16-language neutrality | REQ-CI-012 | T-12, T-30 |
| C3 | macOS arm64 only v1.0 | OQ1 | T-04 |
| C4 | SHA-pin enforcement | REQ-CI-018 | T-16, T-22 |
| C5 | Codex private repo guard | REQ-CI-007, SEC-001 | T-09, T-17 |
| C6 | Claude Action broken range (v1.0.79-84) | REQ-CI-018.1 | T-16, T-22 |
| C7 | No disk writes for tokens | REQ-SEC-002 | T-07 |
| C8 | drop-sudo for Codex | REQ-SEC-006 | T-14, T-17 |
| C9 | Workspace cleanup | REQ-SEC-005 | T-15~T-19, T-22 |
| C10 | Audit log preservation | REQ-SEC-008 | T-29 |
| C11 | External contributor guard | REQ-SEC-010 | T-15~T-20 |
| C12 | Codex auth.json conditional seed | REQ-CI-021 | T-14, T-22 |
| C13 | moai init 자동 호출 안 함 | OQ2 | T-23 (별도 명령) |

---

## 5. EARS Pattern Count (Corrected)

spec.md §3.0의 패턴 분포 표와 실제 REQ 헤더의 패턴 불일치를 수정:

| Pattern | spec.md §3.0 Claim | Actual Count | REQ IDs |
|---------|-------------------|-------------|---------|
| Ubiquitous | 6 | 9 | CI-001, CI-012, CI-018, CI-019, SEC-001, SEC-003, SEC-005, SEC-006, SEC-008 |
| Event-Driven | 14 | 14 | CI-002, CI-003, CI-004, CI-008, CI-009, CI-010, CI-011, CI-013, CI-014, CI-016, CI-017, SEC-004, SEC-007 |
| State-Driven | 4 | 3 | CI-005, CI-020, SEC-009 |
| Optional | 3 | 2 | CI-006, CI-015 |
| Unwanted | 4 | 3 | CI-007, CI-021, SEC-002, SEC-010 |

**Total**: 31 REQs. Event-Driven 실제 13개 (CI-017이 CI-005와 중복 카운트 가능). 정확한 분포는 plan-auditor report 참조.

---

## 6. Test Strategy

### 6.1 TDD Methodology (RED-GREEN-REFACTOR)

모든 task는 TDD 사이클 적용:
1. **RED**: 실패 테스트 작성 (domain logic 인터페이스 정의)
2. **GREEN**: 최소 구현으로 테스트 통과
3. **REFACTOR**: 중복 제거, interface 정리

### 6.2 Test Categories

| Category | Scope | Example |
|----------|-------|---------|
| Unit tests | Domain logic (runner, auth, workflow) | SHA256 검증, 토큰 마스킹, 버전 비교 |
| Integration tests | `gh` CLI 호출 (mock) | `TestRunnerInstall_Integration` |
| Template tests | YAML 구조 검증 | SHA-pin 존재, private guard 포함 |
| E2E tests | `moai github init --dry-run` | 전체 부트스트랩 dry-run |

### 6.3 Coverage Target

- Per-commit minimum: 80% (quality.yaml `tdd_settings.min_coverage_per_commit`)
- Package-level target: 85%+ (TRUST 5)

---

## 7. Dependencies (Execution Order)

```
Wave 1 (T-01 → T-02 → T-03 → T-04 → T-05 → T-06)
                    ↘ T-05 (T-02 병렬 가능)
                    
Wave 2 (T-07 → T-08, T-09, T-10, T-11 병렬)
  T-07은 T-01에 의존 (gh client 필요)

Wave 3 (T-12 독립 → T-13, T-14 → T-15~T-19 병렬 → T-20 → T-21 → T-22)
  T-12은 다른 템플릿의 선행 의존 (detect-language action)
  T-20은 T-15~T-19 완료 후 (panel aggregator)
  T-22은 모든 템플릿 완료 후 (validator)

Wave 4 (T-23~T-26 병렬 가능, T-27~T-29 병렬, T-30 마지막)
  T-23은 Wave 1-3 완료 후
  T-30은 T-23 완료 후 (문서화)
```

---

## 8. Approval Points

| Wave | Gate | Condition |
|------|------|-----------|
| Wave 1 완료 | Go build + test 통과 | `go test ./internal/github/runner/... ./internal/cli/... -run Github` |
| Wave 2 완료 | Auth dry-run 통과 | `moai github auth claude --dry-run` |
| Wave 3 완료 | Template validator 통과 | `go test ./internal/github/workflow/... -run Validate` |
| Wave 4 완료 | 전체 E2E dry-run | `moai github init --dry-run --llm claude,gemini` |

---

Version: 1.0.0
Author: manager-strategy
Last Updated: 2026-04-27
