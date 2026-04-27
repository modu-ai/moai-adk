---
id: SPEC-CI-MULTI-LLM-001
version: 1.0.0
status: Planned
created: 2026-04-27
updated: 2026-04-27
author: manager-spec
priority: P1
issue_number: null
harness_level: standard
language_policy: 16-language-neutral
template_first: true
related_specs:
  - SPEC-CICD-001
  - SPEC-GLM-001
source_report: .moai/reports/devops/multi-llm-self-hosted-actions-2026-04-27.md
---

# SPEC-CI-MULTI-LLM-001: Multi-LLM GitHub Actions Self-Hosted Runner Integration

## HISTORY

| Version | Date       | Author        | Description                                                                |
|---------|------------|---------------|----------------------------------------------------------------------------|
| 1.0.0   | 2026-04-27 | manager-spec  | 초안 생성. expert-devops 보고서 + 사용자 Socratic 인터뷰 2라운드 결과 반영 |

---

## 메타데이터

| 항목         | 값                                                                                      |
|--------------|------------------------------------------------------------------------------------------|
| SPEC ID      | SPEC-CI-MULTI-LLM-001                                                                    |
| 제목         | Multi-LLM GitHub Actions Self-Hosted Runner Integration                                  |
| 상태         | Planned                                                                                  |
| 우선순위     | P1                                                                                       |
| Harness 등급 | standard                                                                                 |
| 생성일       | 2026-04-27                                                                               |
| 담당         | expert-devops (구현), manager-spec (설계)                                                |
| 관련 파일    | `internal/cli/github*.go`, `internal/template/templates/.github/workflows/*.yml.tmpl`    |
| 입력 보고서  | `.moai/reports/devops/multi-llm-self-hosted-actions-2026-04-27.md`                       |

---

## 1. Environment (환경)

### 1.1 대상 시스템

- **CLI 진입점**: `moai` Go 바이너리 (`cmd/moai/`)
- **신규 서브커맨드 그룹**: `moai github`
- **CI 플랫폼**: GitHub Actions (self-hosted runner + GitHub-hosted 하이브리드)
- **Runner 타깃**: 사용자 본인 머신 (macOS arm64 / Linux x86_64 / Linux arm64 동시 지원, 16개 언어 사용자 누구나)
- **Actions Runner 버전**: v2.334.0 이상 (latest at runtime)
- **launchd / systemd 서비스화**: 사용자 OS에 따라 분기
- **템플릿 위치**: `internal/template/templates/.github/workflows/` (현재 비어 있음)
- **배포 모델**: `moai init` 또는 `moai update` 시 사용자 프로젝트로 동기화

### 1.2 통합 LLM 4종 (v1 범위)

| LLM    | 인증 모드                         | 트리거 코멘트       | Action / CLI                                          | Runner                |
|--------|------------------------------------|---------------------|-------------------------------------------------------|------------------------|
| Claude | OAuth (`CLAUDE_CODE_OAUTH_TOKEN`)  | `@claude`           | `anthropics/claude-code-action` (SHA-pin v1.0.72/85+) | self-hosted 권장      |
| Codex  | `auth.json` (조건부 시드)          | `@codex`            | `@openai/codex` CLI (npm)                             | self-hosted 필수      |
| Gemini | `GEMINI_API_KEY`                   | `@gemini-cli`       | `google-github-actions/run-gemini-cli@v0.1.22`        | github-hosted 가능    |
| GLM    | `GLM_AUTH_TOKEN` (Z.ai)            | `@glm`              | `moai glm -- claude -p ...` (Anthropic-compat shim)   | self-hosted 권장      |

### 1.3 현재 시스템과의 차이

| 항목            | 현재 (SPEC-CICD-001)             | 본 SPEC 후                                              |
|-----------------|-----------------------------------|----------------------------------------------------------|
| LLM 개수        | 1 (Claude only)                   | 4 (Claude / Codex / Gemini / GLM)                        |
| Runner 모델     | github-hosted (`ubuntu-latest`)   | self-hosted 우선 + github-hosted 폴백                    |
| 트리거 방식     | `@claude` 코멘트 + PR open        | 코멘트 단수 호출 + PR open 자동 패널 (이중 모드)         |
| 인증 부트스트랩 | 수동 (`gh secret set`)            | `moai github auth <llm>` 대화형 자동화                   |
| 배포 위치       | `.github/workflows/` (직접)       | `internal/template/templates/.github/workflows/` (Template-First) |

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- **A1**: 사용자가 본인 머신에 GitHub Actions runner를 설치할 수 있는 admin 권한을 보유한다.
- **A2**: 사용자 머신의 OS는 macOS (arm64 / x86_64), Linux (x86_64 / arm64) 중 하나이며, Windows는 v1 범위에서 제외 (Codex `drop-sudo` 미지원).
- **A3**: `gh` CLI(>= 2.30.0)가 사용자 머신에 설치되어 있고 `gh auth login` 완료 상태이다.
- **A4**: 사용자가 사용하려는 LLM의 구독/API 키를 사전에 발급받을 수 있다 (Claude Max, Codex 구독, Gemini AI Studio, Z.ai 계정).
- **A5**: `moai` 바이너리가 사용자 머신의 PATH에 등록되어 있다 (GLM 워크플로우가 의존).
- **A6**: GitHub repo는 private이거나 사용자가 private 전환에 동의한다 (Codex 워크플로우 활성화 조건).
- **A7**: `actions/runner` 30일 폐기 정책은 GitHub의 정책이며 변경 시 SPEC 갱신이 필요하다.

### 2.2 비즈니스 가정

- **A8**: 단일 LLM 의존 (Claude only)보다 다중 LLM 패널이 코드 리뷰 품질을 향상시킨다 (cross-validation).
- **A9**: 사용자 본인 머신의 self-hosted runner는 GitHub-hosted macOS arm64 runner($0.16/min)보다 비용 효율적이다.
- **A10**: Template-First 배포 모델을 통해 16개 언어 사용자 누구나 동일한 multi-LLM CI를 활용할 수 있어야 한다.

### 2.3 운영 가정

- **A11**: 사용자가 매월 1회 runner 수동 업그레이드를 수행할 의사와 시간이 있다.
- **A12**: 자동 트리거 패널 LLM 세트는 `.moai/config/sections/github-actions.yaml`에서 사용자별로 커스터마이즈 가능하다.

---

## 3. Requirements (요구사항)

### 3.0 EARS 패턴 분포

| 패턴               | REQ 개수 | 비고                                                     |
|--------------------|----------|----------------------------------------------------------|
| Ubiquitous         | 6        | 항상 활성화되는 시스템 속성                              |
| Event-Driven       | 14       | 사용자 명령/PR 이벤트/코멘트 기반                        |
| State-Driven       | 4        | 30일 폐기 윈도우, repo private 상태, 토큰 만료 등        |
| Optional           | 3        | runner_mode 선택, github-hosted 폴백, 패널 LLM 세트 변경 |
| Unwanted Behavior  | 4        | private repo 강제, 토큰 시드 덮어쓰기 금지 등            |

총 31개 REQ (도메인 21 + 보안 10).

---

### 3.1 핵심 도메인 요구사항 (REQ-CI-*)

#### REQ-CI-001 (Ubiquitous): `moai github` 서브커맨드 그룹

[Ubiquitous] 시스템은 **항상** `moai github` 서브커맨드 그룹을 노출하며, 다음 하위 명령을 제공해야 한다:

- `moai github init`
- `moai github runner {install|register|start|stop|status|upgrade}`
- `moai github auth <llm>` (`<llm>` ∈ {claude, codex, gemini, glm})
- `moai github workflow {add|remove|list}`
- `moai github status`

세부 항목:

- **REQ-CI-001.1**: 명령 그룹은 기존 `moai cc/glm/cg` 패턴과 일관된 구조 (코브라 cobra 서브커맨드 트리)를 사용해야 한다.
- **REQ-CI-001.2**: 모든 명령은 `--help` 플래그로 사용법 문서를 출력해야 한다.
- **REQ-CI-001.3**: 명령은 `--dry-run` 플래그를 지원하여 실제 변경 없이 실행 계획을 출력해야 한다.

---

#### REQ-CI-002 (Event-Driven): `moai github init` 통합 부트스트랩

**WHEN** 사용자가 `moai github init`을 실행하면 **THEN** 시스템은 다음을 순차 수행해야 한다:

1. LLM 선택 대화형 프롬프트 (4종 중 다중 선택)
2. 선택된 각 LLM에 대해 `moai github auth <llm>` 자동 호출
3. self-hosted runner 사용 여부 확인 (예/아니오)
4. 워크플로우 템플릿 선택 (Architecture A: 코멘트 라우터 / Architecture B: PR open 패널 / 둘 다)
5. `internal/template/templates/.github/workflows/`에서 사용자 프로젝트 `.github/workflows/`로 동기화
6. `.moai/config/sections/github-actions.yaml` 생성

세부 항목:

- **REQ-CI-002.1**: LLM 선택 시 사용자가 선택하지 않은 LLM의 워크플로우는 배포되지 않아야 한다.
- **REQ-CI-002.2**: 사용자가 `--llm claude,gemini`처럼 비대화형 플래그로 지정 시 프롬프트를 생략해야 한다.
- **REQ-CI-002.3**: 동기화 결과는 사용자에게 요약 출력 (배포된 파일 목록, 추가된 secret 이름)되어야 한다.

---

#### REQ-CI-003 (Event-Driven): `moai github runner install` — Runner 다운로드 및 launchd 등록

**WHEN** 사용자가 `moai github runner install`을 실행하면 **THEN** 시스템은 다음을 수행해야 한다:

1. `https://api.github.com/repos/actions/runner/releases/latest`에서 최신 버전 조회
2. 사용자 OS/아키텍처 자동 감지 (`runtime.GOOS`, `runtime.GOARCH`)
3. 해당 플랫폼 바이너리(`actions-runner-{os}-{arch}-{version}.tar.gz`) 다운로드
4. SHA256 무결성 검증 (GitHub release 페이지의 공식 체크섬과 비교)
5. `~/actions-runner/`에 압축 해제
6. 사용자에게 launchd 서비스 설치 의사 확인 후 `./svc.sh install` + `./svc.sh start` 실행

세부 항목:

- **REQ-CI-003.1**: 다운로드 실패 시 최대 3회 재시도 후 명확한 오류 메시지를 출력해야 한다.
- **REQ-CI-003.2**: `~/actions-runner/`가 이미 존재하면 사용자에게 덮어쓰기 의사를 확인해야 한다.
- **REQ-CI-003.3**: SHA256 불일치 시 즉시 중단하고 다운로드된 파일을 삭제해야 한다.
- **REQ-CI-003.4**: launchd plist 경로(`~/Library/LaunchAgents/actions.runner.{owner}.{repo}.{name}.plist`)를 사용자에게 출력해야 한다.

---

#### REQ-CI-004 (Event-Driven): `moai github runner register` — GitHub 등록

**WHEN** 사용자가 `moai github runner register`를 실행하면 **THEN** 시스템은 다음을 수행해야 한다:

1. `gh api -X POST /repos/{owner}/{repo}/actions/runners/registration-token`로 등록 토큰 자동 발급
2. 사용자에게 runner 이름 입력 요청 (기본값: `{hostname}-{arch}`)
3. `./config.sh --url ... --token ... --labels "self-hosted,{os},{arch},{project-name}" --replace` 실행
4. 등록 결과를 `.moai/config/sections/github-actions.yaml`의 `runner.labels`에 기록

세부 항목:

- **REQ-CI-004.1**: `--replace` 플래그를 기본 적용하여 동일 이름의 기존 runner를 자동 대체해야 한다.
- **REQ-CI-004.2**: 등록 토큰은 메모리 내에서만 사용하고 디스크에 기록하지 않아야 한다 (1시간 유효).
- **REQ-CI-004.3**: 등록 성공 후 GitHub Settings 링크 (`https://github.com/{owner}/{repo}/settings/actions/runners`)를 출력해야 한다.

---

#### REQ-CI-005 (State-Driven): Runner 30일 폐기 윈도우 알림

**WHILE** 등록된 runner의 버전이 latest release 대비 25일 이상 경과한 상태이면 **THEN** 시스템은:

1. `moai github runner status` 실행 시 경고 표시 (`⚠ Runner v2.X.X is N days old. Upgrade required within {30-N} days.`)
2. `moai doctor` 실행 시 동일 경고 노출
3. SessionStart 훅에서 1일 1회 `systemMessage`로 알림

세부 항목:

- **REQ-CI-005.1**: 25일 초과 시 경고, 30일 초과 시 ERROR로 격상해야 한다.
- **REQ-CI-005.2**: 경고는 사용자가 `moai github runner upgrade`를 실행하여 해소해야 한다.
- **REQ-CI-005.3**: `moai github runner upgrade` 실행 시 자동으로 `./svc.sh stop` → 다운로드 → 압축 해제 → `./svc.sh start` 순서로 진행해야 한다.

---

#### REQ-CI-006 (Optional): GitHub-hosted Runner 폴백

[Optional] **WHERE** 사용자가 `.moai/config/sections/github-actions.yaml`에 `runner_mode: github-hosted`를 설정한 경우 시스템은:

- self-hosted runner 설치 단계를 생략하고 모든 워크플로우의 `runs-on`을 `ubuntu-latest`로 렌더링해야 한다.
- Codex 워크플로우는 macOS self-hosted 의존성으로 인해 자동 비활성화되어야 한다 (Linux GitHub-hosted runner에서는 Codex auth.json 영속성 보장 불가).

세부 항목:

- **REQ-CI-006.1**: 폴백 모드에서는 macOS arm64 runner($0.16/min) 비용 경고를 출력해야 한다.
- **REQ-CI-006.2**: 사용자가 `runner_mode: hybrid`로 설정 시 LLM별로 self-hosted/github-hosted를 분리할 수 있어야 한다.

---

#### REQ-CI-007 (Unwanted Behavior): Public Repo에서 Codex 거부

[Unwanted] **IF** GitHub repo가 public 상태이면 **THEN** 시스템은 Codex 워크플로우 실행을 거부해야 한다.

구현:

- 모든 Codex 워크플로우 템플릿은 `if: github.event.repository.private == true` 가드를 포함해야 한다.
- `moai github auth codex` 실행 시 `gh repo view --json visibility`로 repo 가시성을 확인하고, public이면 즉시 중단 + 오류 메시지 출력.
- `moai github init`에서 Codex 선택 시 동일 검증 적용.

세부 항목:

- **REQ-CI-007.1**: 거부 시 OpenAI 공식 정책 링크 (`https://developers.openai.com/codex/auth/ci-cd-auth`)를 출력해야 한다.
- **REQ-CI-007.2**: 사용자가 `--force-public` 플래그를 사용해도 Codex만은 차단해야 한다 (compliance HARD rule).

---

#### REQ-CI-008 (Event-Driven): `moai github auth claude` — Claude OAuth 부트스트랩

**WHEN** 사용자가 `moai github auth claude`를 실행하면 **THEN** 시스템은:

1. `claude` CLI 설치 여부 확인, 없으면 `npm install -g @anthropic-ai/claude-code` 안내
2. `claude setup-token` 실행 안내 → 사용자 토큰 클립보드 복사 후 stdin 입력 요청
3. `gh secret set CLAUDE_CODE_OAUTH_TOKEN --body "<token>"` 자동 실행
4. 등록 검증: `gh secret list | grep CLAUDE_CODE_OAUTH_TOKEN`

세부 항목:

- **REQ-CI-008.1**: 토큰은 메모리 내에서만 처리하고 임시 파일에 기록하지 않아야 한다.
- **REQ-CI-008.2**: 등록 후 사용자에게 Max plan 구독 필수 사실 (보고서 §4.1 USER-VERIFIED)을 안내해야 한다.

---

#### REQ-CI-009 (Event-Driven): `moai github auth codex` — Codex auth.json 부트스트랩

**WHEN** 사용자가 `moai github auth codex`를 실행하면 **THEN** 시스템은:

1. private repo 검증 (REQ-CI-007 적용)
2. `codex` CLI 설치 여부 확인, 없으면 `npm install -g @openai/codex` 안내
3. `codex login` 실행 안내 → `~/.codex/auth.json` 생성 대기
4. `gh secret set CODEX_AUTH_JSON --body "$(cat ~/.codex/auth.json)"` 실행

세부 항목:

- **REQ-CI-009.1**: `~/.codex/auth.json` 파일 권한이 600인지 검증해야 한다.
- **REQ-CI-009.2**: 워크플로우 템플릿은 조건부 시드 패턴 (`[ ! -f "$CODEX_HOME/auth.json" ]`)을 포함해야 한다 (REQ-SEC-002 참조).

---

#### REQ-CI-010 (Event-Driven): `moai github auth gemini` — Gemini API Key 부트스트랩

**WHEN** 사용자가 `moai github auth gemini`를 실행하면 **THEN** 시스템은:

1. Google AI Studio 링크 (`https://aistudio.google.com/apikey`) 출력
2. 사용자에게 API key 입력 요청 (stdin, masked input)
3. 형식 검증 (영숫자 + `-_`, 길이 39자)
4. `gh secret set GEMINI_API_KEY --body "<key>"` 실행

세부 항목:

- **REQ-CI-010.1**: 입력된 키 첫 글자 + 마지막 4글자만 표시 (보안 마스킹).
- **REQ-CI-010.2**: 무료 tier 한도 (Gemini 2.5 Flash 500 RPD) 정보를 안내해야 한다.

---

#### REQ-CI-011 (Event-Driven): `moai github auth glm` — GLM 토큰 부트스트랩

**WHEN** 사용자가 `moai github auth glm`을 실행하면 **THEN** 시스템은:

1. Z.ai 링크 (`https://api.z.ai`) 출력
2. 사용자에게 GLM auth token 입력 요청 (stdin, masked)
3. `gh secret set GLM_AUTH_TOKEN --body "<token>"` 실행

세부 항목:

- **REQ-CI-011.1**: 워크플로우 템플릿은 `~/.moai/.env.glm` 조건부 생성 패턴을 포함해야 한다.
- **REQ-CI-011.2**: 환경 변수 4종 (`ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, `DISABLE_BETAS=1`, `DISABLE_PROMPT_CACHING=1`)을 자동 주입해야 한다 (SPEC-GLM-001 정합성).

---

#### REQ-CI-012 (Ubiquitous): 16개 언어 중립 워크플로우 템플릿

[Ubiquitous] 시스템은 **항상** 16개 지원 언어 (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)를 동등 취급하는 워크플로우 템플릿을 제공해야 한다.

구현:

- 워크플로우 템플릿은 `project_markers`(go.mod, package.json, pyproject.toml, Cargo.toml 등)을 자동 감지하는 step을 포함해야 한다.
- 감지 결과에 따라 빌드/테스트 step을 동적 분기해야 한다 (예: go.mod 감지 시 `go test`, package.json 감지 시 `npm test`).
- 특정 언어를 PRIMARY로 지정하지 않아야 한다.

세부 항목:

- **REQ-CI-012.1**: 템플릿은 `internal/template/templates/.github/workflows/*.yml.tmpl`에 위치해야 한다 (Template-First 강제).
- **REQ-CI-012.2**: 언어 감지 로직은 reusable composite action으로 분리해야 한다 (`internal/template/templates/.github/actions/detect-language/`).
- **REQ-CI-012.3**: 16개 언어 중 하나도 감지되지 않을 경우 빌드/테스트 step을 graceful skip하고 LLM 리뷰만 실행해야 한다.

---

#### REQ-CI-013 (Event-Driven): PR Open 자동 패널 트리거 (Architecture B)

**WHEN** PR이 `opened`, `ready_for_review`, `reopened` 이벤트를 발생시키면 **THEN** 시스템은 `.moai/config/sections/github-actions.yaml`의 `auto_panel.llms` 목록에 정의된 LLM 세트로 병렬 리뷰를 실행해야 한다.

세부 항목:

- **REQ-CI-013.1**: 기본 패널 세트는 `[claude, gemini]` (보고서 §10 권고 시작 세트)로 설정.
- **REQ-CI-013.2**: 각 LLM job은 독립 concurrency group (`group: llm-panel-{llm}-${{ github.event.pull_request.number }}`)을 사용하고 `cancel-in-progress: true` 적용.
- **REQ-CI-013.3**: `paths-ignore`에 `.github/**`를 포함하여 CI 설정 변경 PR에서는 패널을 생략해야 한다.
- **REQ-CI-013.4**: 패널 결과는 PR 코멘트로 게시하며 LLM별 구분 헤더 (`### Claude Review`, `### Gemini Review`)를 사용해야 한다.

---

#### REQ-CI-014 (Event-Driven): PR Comment 단수 LLM 트리거 (Architecture A)

**WHEN** PR 또는 Issue 코멘트에 `@claude`, `@codex`, `@gemini-cli`, `@glm` 중 하나가 포함되면 **THEN** 시스템은 해당 LLM 단일 워크플로우만 트리거해야 한다.

매핑 표:

| 코멘트 패턴            | 트리거되는 워크플로우            |
|------------------------|-----------------------------------|
| `@claude` (PR)         | `claude-code-review.yml`          |
| `@claude` (Issue)      | `claude.yml`                      |
| `@codex`               | `codex-review.yml`                |
| `@gemini-cli /review`  | `gemini-review.yml`               |
| `@glm`                 | `glm-review.yml`                  |

세부 항목:

- **REQ-CI-014.1**: 모든 코멘트 트리거는 `author_association in ('OWNER', 'MEMBER', 'COLLABORATOR')` 가드로 외부 사용자 abuse를 차단해야 한다.
- **REQ-CI-014.2**: 다중 LLM을 동시에 멘션 (`@claude @gemini-cli`)할 경우 모두 병렬 실행해야 한다.
- **REQ-CI-014.3**: 코멘트에 명령 인자 (`@claude review --focus security`)를 포함할 경우 LLM 워크플로우의 prompt에 전달해야 한다.

---

#### REQ-CI-015 (Optional): 사용자 정의 패널 LLM 세트

[Optional] **WHERE** 사용자가 `.moai/config/sections/github-actions.yaml`에서 `auto_panel.llms`를 변경한 경우 시스템은 변경된 세트로 패널을 실행해야 한다.

예시:

```yaml
auto_panel:
  enabled: true
  llms: [claude, gemini, glm]   # codex 제외
  triggers: [opened, ready_for_review]
```

---

#### REQ-CI-016 (Event-Driven): `moai github workflow add/remove/list`

**WHEN** 사용자가 `moai github workflow add <llm>`을 실행하면 **THEN** 시스템은 해당 LLM 워크플로우 템플릿을 사용자 프로젝트로 동기화해야 한다.

`remove`는 역연산, `list`는 현재 배포된 워크플로우 목록 출력.

---

#### REQ-CI-017 (Event-Driven): `moai github status`

**WHEN** 사용자가 `moai github status`를 실행하면 **THEN** 시스템은 다음을 출력해야 한다:

- 등록된 self-hosted runner 목록 (이름, 상태, 버전, 마지막 실행 시각)
- 활성화된 LLM 워크플로우 목록 (파일명, LLM, 트리거 모드)
- 등록된 GitHub secret 목록 (이름만, 값은 마스킹)
- runner 30일 폐기 윈도우 잔여일

---

#### REQ-CI-018 (Ubiquitous): SHA-Pin 강제

[Ubiquitous] 모든 LLM Action은 floating tag (`@v1`, `@latest`)이 아닌 SHA-pin (40자 commit hash)을 사용해야 한다.

세부 항목:

- **REQ-CI-018.1**: Claude Code Action은 v1.0.79–v1.0.84 범위를 회피해야 한다 (#1126 self-hosted 버그).
- **REQ-CI-018.2**: 권장 SHA: v1.0.72 또는 v1.0.85+ (보고서 §4.1).
- **REQ-CI-018.3**: 템플릿 렌더링 시 SHA를 환경 변수 또는 config로 주입할 수 있어야 한다 (`workflows.claude.action_sha`).

---

#### REQ-CI-019 (Ubiquitous): Template-First 강제

[Ubiquitous] 모든 신규 워크플로우 파일은 `internal/template/templates/.github/workflows/`에 우선 추가되고 `make build`로 임베드된 후 사용자 프로젝트에 동기화되어야 한다.

세부 항목:

- **REQ-CI-019.1**: 사용자 프로젝트의 `.github/workflows/`에 직접 파일 추가는 금지.
- **REQ-CI-019.2**: 템플릿 동기화는 `internal/cli/github_init.go`(가칭)에서 `internal/template` 패키지의 `Deploy` API를 호출해야 한다.
- **REQ-CI-019.3**: 사용자 프로젝트에 동일 이름의 파일이 이미 존재할 경우 사용자에게 덮어쓰기 / merge / skip 의사를 확인해야 한다.

---

#### REQ-CI-020 (State-Driven): runner_mode 상태 기반 워크플로우 렌더링

**WHILE** `.moai/config/sections/github-actions.yaml`의 `runner_mode`가 `self-hosted`이면 **THEN** 시스템은 워크플로우의 `runs-on`을 `[self-hosted, {os}, {arch}, {project-name}]`로 렌더링하고, `github-hosted`이면 `ubuntu-latest`로 렌더링해야 한다.

세부 항목:

- **REQ-CI-020.1**: `hybrid` 모드에서는 LLM별 설정 (`workflows.{llm}.runner_mode`)을 우선 적용.

---

#### REQ-CI-021 (Unwanted Behavior): Codex auth.json 매번 덮어쓰기 금지

[Unwanted] **IF** Codex 워크플로우가 `auth.json`을 매 실행마다 시크릿에서 덮어쓰면 **THEN** 시스템은 이를 거부하고 조건부 시드 패턴을 강제해야 한다.

구현:

- 템플릿에 `if [ ! -f "$CODEX_HOME/auth.json" ]; then ... fi` 가드 포함.
- 템플릿 검증 단계에서 무조건 덮어쓰기 패턴 (`echo > auth.json`, `cp ... auth.json`) 탐지 시 빌드 실패.

근거: refresh token 손실 방지 (보고서 §4.2).

---

### 3.2 보안 도메인 요구사항 (REQ-SEC-*)

#### REQ-SEC-001 (Ubiquitous): Private Repo 강제 (Codex)

[Ubiquitous] 모든 Codex 워크플로우는 `if: github.event.repository.private == true` job-level 가드를 포함해야 한다.

(REQ-CI-007과 정합. 본 REQ는 보안 관점에서 별도 검증 가능한 게이트로 명시.)

---

#### REQ-SEC-002 (Unwanted Behavior): 시크릿 디스크 기록 금지

[Unwanted] **IF** `moai github auth <llm>` 명령이 토큰을 임시 파일이나 로그에 기록하면 **THEN** 시스템은 이를 즉시 차단해야 한다.

구현:

- Go 코드에서 토큰은 `os.Stdin` → 바로 `gh secret set`의 stdin으로 파이프하고 변수에 저장하지 않거나 `defer` clear.
- 디버그 로그에 `***`로 마스킹.

---

#### REQ-SEC-003 (Ubiquitous): SHA-Pin Action

[Ubiquitous] (REQ-CI-018 보안 관점 mirror) 모든 third-party Action은 SHA-pin해야 supply chain 공격을 방지할 수 있다.

세부 항목:

- **REQ-SEC-003.1**: GitHub 공식 Action (`actions/*`)은 floating major tag (`@v5`)을 허용하되 `dependabot.yml`에 등록해야 한다.

---

#### REQ-SEC-004 (Event-Driven): 시크릿 회전 정책 안내

**WHEN** 사용자가 `moai github status`를 실행하면 **THEN** 시스템은 각 시크릿의 권장 회전 주기를 표시해야 한다.

| Secret                       | 회전 주기                  |
|------------------------------|----------------------------|
| `CLAUDE_CODE_OAUTH_TOKEN`    | 구독 변경 시 / 6개월       |
| `CODEX_AUTH_JSON`            | 401 발생 시 / 월간 점검    |
| `GEMINI_API_KEY`             | 6개월                      |
| `GLM_AUTH_TOKEN`             | 6개월                      |

---

#### REQ-SEC-005 (Ubiquitous): Workspace Cleanup 강제

[Ubiquitous] 모든 self-hosted runner 워크플로우는 job 시작 시 workspace cleanup step을 포함해야 한다.

구현:

```yaml
- name: Cleanup workspace
  if: always()
  run: |
    find "$GITHUB_WORKSPACE" -name "*.env" -delete 2>/dev/null || true
    # CODEX_HOME/auth.json은 보존 (조건부 시드와 충돌 방지)
```

세부 항목:

- **REQ-SEC-005.1**: `actions/checkout@v5`의 `clean: true` 기본값을 변경하지 않아야 한다.
- **REQ-SEC-005.2**: `~/.moai/.env.glm`은 매 실행마다 시크릿에서 재생성되므로 삭제 안전.

---

#### REQ-SEC-006 (Ubiquitous): Codex `drop-sudo` 강제

[Ubiquitous] 모든 Codex 실행은 `--safety-strategy drop-sudo` 플래그를 포함해야 한다.

세부 항목:

- **REQ-SEC-006.1**: macOS/Linux에서만 지원되며, Windows는 v1 범위에서 Codex 미지원 (A2 가정).
- **REQ-SEC-006.2**: 권장: 별도 `_runner` 사용자 계정을 사용하여 sudoers에서 제외 (defense-in-depth).

---

#### REQ-SEC-007 (Event-Driven): Network Egress 검증

**WHEN** `moai github runner install` 완료 후 **THEN** 시스템은 다음 도메인에 대한 outbound HTTPS 연결을 자동 검증해야 한다:

- `api.github.com`
- `api.anthropic.com`
- `api.openai.com`
- `generativelanguage.googleapis.com`
- `api.z.ai`
- `codecov.io`

세부 항목:

- **REQ-SEC-007.1**: 검증 실패 도메인이 있을 경우 사용자에게 방화벽 설정 안내 메시지 출력.
- **REQ-SEC-007.2**: 사용자가 사용하지 않는 LLM에 해당하는 도메인은 검증 생략.

---

#### REQ-SEC-008 (Ubiquitous): Audit Log 보존

[Ubiquitous] 시스템은 다음 위치에 audit log를 보존해야 한다:

- GitHub Actions 로그: 90일 (GitHub 기본값)
- Self-hosted runner diag: `~/actions-runner/_diag/` (사용자 머신, 자동 정리 비활성화 권장)
- `moai github *` 명령 실행 로그: `.moai/logs/github-cli-{date}.log` (30일 retention, `system.yaml` `document_management.directories.logs` 정책 준수)

---

#### REQ-SEC-009 (State-Driven): 토큰 만료 감지

**WHILE** LLM API가 401 응답을 반환한 상태이면 **THEN** 시스템은:

1. GitHub Actions job을 명확한 오류 메시지 (`AUTH_EXPIRED: rotate {SECRET_NAME} via 'moai github auth {llm}'`)로 종료
2. PR에 review comment로 알림 (Architecture A에서)

세부 항목:

- **REQ-SEC-009.1**: 검출 패턴: HTTP 응답 본문의 `unauthorized`, `invalid_token`, `expired` 키워드.

---

#### REQ-SEC-010 (Unwanted Behavior): External Contributor abuse 방지

[Unwanted] **IF** PR/Issue 코멘트의 `author_association`이 `OWNER`, `MEMBER`, `COLLABORATOR` 중 하나가 아니면 **THEN** 시스템은 LLM 워크플로우 트리거를 거부해야 한다.

(REQ-CI-014.1과 정합, 보안 관점 mirror.)

---

## 4. Specifications (사양)

### 4.1 CLI 명령 트리

```
moai github
├── init                          # 통합 부트스트랩 (REQ-CI-002)
├── runner
│   ├── install                   # REQ-CI-003
│   ├── register                  # REQ-CI-004
│   ├── start                     # launchd start
│   ├── stop                      # launchd stop
│   ├── status                    # 상태 조회 (REQ-CI-005 포함)
│   └── upgrade                   # 30일 폐기 대응 (REQ-CI-005)
├── auth
│   ├── claude                    # REQ-CI-008
│   ├── codex                     # REQ-CI-009
│   ├── gemini                    # REQ-CI-010
│   └── glm                       # REQ-CI-011
├── workflow
│   ├── add <llm>                 # REQ-CI-016
│   ├── remove <llm>              # REQ-CI-016
│   └── list                      # REQ-CI-016
└── status                        # REQ-CI-017
```

### 4.2 워크플로우 파일 매핑

| 템플릿 파일                                                          | 트리거                              | Runner                | LLM    |
|----------------------------------------------------------------------|-------------------------------------|-----------------------|--------|
| `internal/template/templates/.github/workflows/claude.yml.tmpl`      | `@claude` (Issue) / Issue 이벤트    | self-hosted (default) | Claude |
| `internal/template/templates/.github/workflows/claude-code-review.yml.tmpl` | PR open + `@claude` (PR)     | self-hosted           | Claude |
| `internal/template/templates/.github/workflows/codex-review.yml.tmpl`| `@codex` (코멘트) + private 가드    | self-hosted (필수)    | Codex  |
| `internal/template/templates/.github/workflows/gemini-review.yml.tmpl`| `@gemini-cli` (코멘트) / PR open   | github-hosted         | Gemini |
| `internal/template/templates/.github/workflows/glm-review.yml.tmpl`  | `@glm` (코멘트) / PR open           | self-hosted           | GLM    |
| `internal/template/templates/.github/workflows/llm-panel.yml.tmpl`   | PR open (Architecture B)            | mixed                 | 다중   |

### 4.3 Composite Action

| 디렉토리                                                                       | 역할                             |
|---------------------------------------------------------------------------------|----------------------------------|
| `internal/template/templates/.github/actions/detect-language/action.yml.tmpl`  | 16개 언어 자동 감지 (REQ-CI-012) |
| `internal/template/templates/.github/actions/setup-glm-env/action.yml.tmpl`    | GLM 환경 변수 4종 주입            |
| `internal/template/templates/.github/actions/codex-bootstrap/action.yml.tmpl`  | auth.json 조건부 시드            |

### 4.4 Config 파일

`internal/template/templates/.moai/config/sections/github-actions.yaml.tmpl`:

```yaml
github_actions:
  enabled: true
  runner:
    mode: self-hosted        # self-hosted | github-hosted | hybrid
    labels: [self-hosted, macOS, ARM64]
    install_path: ~/actions-runner
  workflows:
    claude:
      enabled: true
      action_sha: <SHA-v1.0.85+>
      runner_mode: self-hosted
      secret_name: CLAUDE_CODE_OAUTH_TOKEN
    codex:
      enabled: false                    # 사용자가 init에서 활성화
      runner_mode: self-hosted
      secret_name: CODEX_AUTH_JSON
      private_repo_required: true
    gemini:
      enabled: false
      action_version: v0.1.22
      runner_mode: github-hosted
      secret_name: GEMINI_API_KEY
    glm:
      enabled: false
      runner_mode: self-hosted
      secret_name: GLM_AUTH_TOKEN
      base_url: https://api.z.ai/v1
  auto_panel:
    enabled: true
    llms: [claude, gemini]              # PR open 시 자동 패널 (REQ-CI-013)
    triggers: [opened, ready_for_review]
  comment_router:
    enabled: true
    patterns:
      "@claude": claude
      "@codex": codex
      "@gemini-cli": gemini
      "@glm": glm
```

### 4.5 Go 패키지 구조

```
internal/cli/
├── github.go                # 루트 cobra command (cmd group "github")
├── github_init.go           # REQ-CI-002
├── github_runner.go         # REQ-CI-003 ~ 005, 017
├── github_auth.go           # REQ-CI-008 ~ 011
├── github_workflow.go       # REQ-CI-016
├── github_status.go         # REQ-CI-017
└── github_test.go           # 모든 _test.go (table-driven)

internal/github/             # 신규 패키지 (CLI와 분리된 도메인 로직)
├── runner/
│   ├── installer.go         # 다운로드 + SHA256 검증
│   ├── registrar.go         # gh API 호출
│   ├── service.go           # launchd / systemd 분기
│   └── version.go           # 30일 폐기 윈도우 검사
├── auth/
│   ├── claude.go
│   ├── codex.go
│   ├── gemini.go
│   └── glm.go
├── workflow/
│   ├── deployer.go          # 템플릿 동기화
│   └── validator.go         # SHA-pin / Codex 가드 검증
└── secret.go                # gh secret CRUD 래퍼
```

---

## 5. Exclusions (What NOT to Build)

[HARD] 본 SPEC의 v1 범위에서 명시적으로 **제외**되는 항목:

- **EX-1**: Windows self-hosted runner 지원 (Codex `drop-sudo` 미지원, A2 가정)
- **EX-2**: GitHub Enterprise Server 지원 (v2: 별도 SPEC)
- **EX-3**: GitLab CI / Bitbucket Pipelines 등 GitHub 외 플랫폼
- **EX-4**: Workload Identity Federation (Gemini Option B) — 본 SPEC은 API key 방식만 (보고서 §4.3)
- **EX-5**: 5번째 이상 LLM (Mistral, Llama 등) — v1은 4종 고정
- **EX-6**: Cost dashboard 자동화 — P3 후속 작업으로 분리 (보고서 §9 P3)
- **EX-7**: LLM 응답 quality scoring 자동화 — `review-quality-gate.yml`(SPEC-CICD-001)와 별개
- **EX-8**: Self-hosted runner cluster (다중 머신) — 단일 사용자 머신 가정
- **EX-9**: 자동 시크릿 회전 (rotation 명령은 안내만, 실행은 사용자 수동)
- **EX-10**: PR review의 LLM간 합의 알고리즘 (consensus voting) — 단순 병렬 게시
- **EX-11**: SaaS 대안 비교 (CodeRabbit, Sourcery 등) — 보고서 §7에 비교 표만 존재, 통합 제외

---

## 6. Risks (위험 등록부)

보고서 §8 Risk Register 11개 항목을 SPEC 위험으로 매핑:

| ID  | 위험                                                                  | 심각도 | 확률           | 대응 REQ                |
|-----|-----------------------------------------------------------------------|--------|----------------|--------------------------|
| R1  | Codex auth.json 손상 → 401 → CI breakage                              | P1     | Medium         | REQ-SEC-009, REQ-CI-021  |
| R2  | Claude Code Action SDK regression (#1126) self-hosted breakage        | P1     | Medium (known) | REQ-CI-018, REQ-SEC-003  |
| R3  | Runner 30일 폐기 윈도우 미준수                                          | P1     | High           | REQ-CI-005, REQ-CI-017   |
| R4  | 단일 머신 SPOF (사용자 본인 머신만 self-hosted)                        | P2     | High           | REQ-CI-006 (github-hosted 폴백) |
| R5  | Network egress 차단 (방화벽이 LLM 도메인 차단)                          | P2     | Low            | REQ-SEC-007              |
| R6  | 구독 계정 정지 → CI breakage                                           | P2     | Low            | REQ-SEC-009, manual fallback |
| R7  | Public 기여자 abuse → self-hosted runner 공격                          | P0     | Very Low       | REQ-CI-007, REQ-SEC-001, REQ-SEC-010 |
| R8  | Codex auth.json 동시 실행으로 토큰 손상                                | P1     | Medium         | concurrency group (REQ-CI-013.2 패턴) |
| R9  | Gemini 무료 tier 한도 초과                                              | P2     | Low            | REQ-CI-010.2 (안내), monitoring (P3) |
| R10 | GLM 엔드포인트 일시 장애                                                | P3     | Low            | `continue-on-error: true` on GLM jobs |
| R11 | `CLAUDE_CODE_OAUTH_TOKEN` 만료                                          | P1     | Low            | REQ-SEC-009              |

---

## 7. Dependencies (의존성)

### 7.1 외부 도구

- `gh` CLI >= 2.30.0 (사용자 환경)
- `actions/runner` v2.334.0+ (다운로드 대상)
- `claude` CLI (npm: `@anthropic-ai/claude-code`)
- `codex` CLI (npm: `@openai/codex`)
- `moai` 바이너리 (PATH 등록 필수, GLM 워크플로우 의존)

### 7.2 GitHub Actions 의존

- `anthropics/claude-code-action` (SHA-pin)
- `google-github-actions/run-gemini-cli` v0.1.22 (또는 SHA-pin)
- `actions/checkout@v5` (SHA-pin: `08c6903cd8c0fde910a37f88322edcfb5dd907a8`)
- `actions/setup-go@v6` (조건부, Go 프로젝트인 경우)

### 7.3 관련 SPEC

- **SPEC-CICD-001** (현행): Claude 단일 LLM CI/CD 재설계. 본 SPEC은 이를 multi-LLM으로 확장.
- **SPEC-GLM-001**: GLM 환경 변수 호환성 (`DISABLE_BETAS`, `DISABLE_PROMPT_CACHING`). REQ-CI-011.2가 정합 강제.

---

## 8. Open Questions — Resolved (2026-04-27)

본 SPEC의 모든 Open Question은 사용자(GOOS) Socratic 인터뷰를 거쳐 해결되었습니다. `/moai run` 진행을 위한 모든 차단 항목 해소.

### OQ1 — Linux self-hosted runner 지원 우선순위 [RESOLVED]

- **결정**: v1.0은 **macOS arm64만 지원**, Linux는 v1.1로 분리
- **근거**: GOOS 환경 우선 검증, 테스트 매트릭스 단축, v1.0 출시 일정 단축
- **영향 REQ**: REQ-CI-003.4 (launchd plist 경로) — 이미 macOS-only로 명시되어 정합. systemd 분기는 v1.1 진입 시 활성화
- **plan.md task 영향**: T-04 (`internal/github/runner/service.go` systemd stub)는 v1.0에서는 stub 유지, v1.1 SPEC에서 구현

### OQ2 — moai init 시 moai github init 자동 호출 여부 [RESOLVED]

- **결정**: **자동 호출하지 않음**. 사용자가 명시적으로 `moai github init` 실행
- **근거**: 책임 분리, `moai init` 단순화 유지, GitHub 통합 미사용 사용자 보호
- **영향 REQ**: 없음 (REQ-CI-002는 명시적 `moai github init` 호출만 다룸)

### OQ3 — Architecture B 패널 결과 코멘트 형식 [RESOLVED]

- **결정**: **단일 통합 코멘트, LLM별 섹션 분리** (`### Claude Review`, `### Gemini Review`, `### GLM Review`, `### Codex Review`)
- **근거**: PR 노이즈 최소화, LLM 간 의견 비교 용이, GitHub UI 가독성
- **영향 REQ**: REQ-CI-013.4 — 이미 LLM별 구분 헤더 명시되어 정합 (수정 불필요)
- **추가 명시**: T-19 (`llm-panel.yml.tmpl`)는 모든 LLM job 결과를 단일 코멘트로 aggregator step에서 합치도록 구현

### OQ4 — 30일 runner 폐기 알림 검증 위치 [RESOLVED]

- **결정**: **로컬 SessionStart 훅에서만 검증**. GitHub Actions cron 워크플로우 추가하지 않음
- **근거**: 인프라 의존 0, 보고서 §3.6 패턴 그대로 적용, moai 사용자가 세션 시작마다 즉시 인지
- **영향 REQ**: REQ-CI-005 — SessionStart 훅 통합은 plan.md M4 (Wave 4)에서 명시 구현
- **알려진 한계**: 사용자가 moai를 30일 이상 미실행 시 알림 누락 가능. v1.x 사용자 피드백 후 cron 추가 여부 재평가

---

## 9. Acceptance Summary

본 SPEC의 acceptance 기준은 별도 `acceptance.md`에 Given-When-Then 형식으로 상세 기술. 핵심 지표:

- **기능 검증**: 31개 REQ 모두 자동 테스트로 검증 가능 (table-driven Go test)
- **보안 검증**: REQ-SEC-001 ~ 010 모두 정적 분석 + 워크플로우 lint 통과
- **언어 중립성**: REQ-CI-012 16개 언어 모두에 대해 detect-language action 동작 확인
- **Template-First**: REQ-CI-019 `internal/template/commands_audit_test.go` 패턴 동일 적용

---

## 10. References

- 입력 분석 보고서: `.moai/reports/devops/multi-llm-self-hosted-actions-2026-04-27.md`
- OpenAI Codex CI/CD Auth: https://developers.openai.com/codex/auth/ci-cd-auth
- GitHub Actions Runner: https://github.com/actions/runner/releases
- Anthropic Claude Code Action: https://github.com/anthropics/claude-code-action
- Google Gemini CLI Action: https://github.com/google-github-actions/run-gemini-cli
- Z.ai GLM Pricing: https://docs.z.ai/guides/overview/pricing
- 관련 규칙: `.claude/rules/moai/development/coding-standards.md` (Template-First, 16-language 중립성)
- CLAUDE.local.md §15 (16개 언어 중립성), §2 (Template-First Rule), §18 (Git workflow)
