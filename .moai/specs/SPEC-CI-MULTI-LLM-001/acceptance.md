---
spec_id: SPEC-CI-MULTI-LLM-001
version: 1.0.0
created: 2026-04-27
updated: 2026-04-27
---

# SPEC-CI-MULTI-LLM-001 — Acceptance Criteria

## Given-When-Then Format

모든 31개 REQ에 대한 관측 가능한 수용 기준을 Given-When-Then 형식으로 정의.

---

## 3.1 Core Domain Requirements (REQ-CI-*)

### REQ-CI-001: `moai github` Subcommand Group

**AC-001.1** (Command Structure):
- **Given** `moai` 바이너리가 PATH에 등록되어 있고
- **When** `moai github --help`을 실행하면
- **Then** init, runner, auth, workflow, status 하위 명령이 help 출력에 포함된다

**AC-001.2** (Cobra Consistency):
- **Given** 기존 `moai cc/glm/cg` 명령이 정상 동작하고
- **When** `moai github` 관련 명령을 실행하면
- **Then** cobra 서브커맨드 트리 구조가 기존 패턴과 일치한다 (PersistentFlags, GroupID 등)

**AC-001.3** (Help Output):
- **Given** moai 바이너리가 빌드되어 있고
- **When** `moai github runner --help`을 실행하면
- **Then** install, register, start, stop, status, upgrade 하위 명령과 설명이 출력된다

**AC-001.4** (Dry-Run):
- **Given** `moai github init --dry-run`을 실행하면
- **Then** 실제 파일 변경 없이 실행 계획이 출력된다

---

### REQ-CI-002: `moai github init` Integrated Bootstrap

**AC-002.1** (Interactive LLM Selection):
- **Given** `moai github init`을 실행하고
- **When** LLM 선택 프롬프트가 표시되면
- **Then** claude, codex, gemini, glm 4종 중 다중 선택이 가능하다

**AC-002.2** (Non-interactive Flag):
- **Given** `--llm claude,gemini` 플래그와 함께 실행하면
- **When** init 프로세스가 시작되면
- **Then** LLM 선택 프롬프트가 생략되고 claude, gemini만 배포된다

**AC-002.3** (Selective Deployment):
- **Given** codex를 선택하지 않고 init을 완료하면
- **When** `.github/workflows/`를 확인하면
- **Then** codex-review.yml이 존재하지 않는다

**AC-002.4** (Summary Output):
- **Given** init이 성공적으로 완료되면
- **When** 출력을 확인하면
- **Then** 배포된 파일 목록과 추가된 secret 이름이 요약 표시된다

**AC-002.5** (Conflict Resolution):
- **Given** 사용자 프로젝트에 동일 이름의 워크플로우 파일이 이미 존재하고
- **When** init이 실행되면
- **Then** 덮어쓰기/merge/skip 선택 프롬프트가 표시된다

---

### REQ-CI-003: `moai github runner install`

**AC-003.1** (Auto Detection):
- **Given** 사용자 OS가 macOS arm64이고
- **When** `moai github runner install`을 실행하면
- **Then** `actions-runner-darwin-arm64-{version}.tar.gz`를 다운로드한다

**AC-003.2** (SHA256 Verification):
- **Given** 다운로드한 runner 바이너리의 SHA256이 GitHub 공식 체크섬과 불일치하면
- **When** 검증 단계가 실행되면
- **Then** 다운로드된 파일이 삭제되고 error가 반환된다

**AC-003.3** (Retry Logic):
- **Given** 다운로드가 실패하고
- **When** 자동 재시도가 3회 수행되면
- **Then** 최종 실패 시 명확한 오류 메시지가 출력된다

**AC-003.4** (Existing Directory):
- **Given** `~/actions-runner/`가 이미 존재하고
- **When** install을 실행하면
- **Then** 덮어쓰기 의사 확인 프롬프트가 표시된다

**AC-003.5** (launchd Registration):
- **Given** 압축 해제가 완료되고
- **When** 사용자가 서비스 설치를 승인하면
- **Then** `./svc.sh install` + `./svc.sh start`가 실행되고 plist 경로가 출력된다

---

### REQ-CI-004: `moai github runner register`

**AC-004.1** (Auto Token):
- **Given** `gh auth login`이 완료된 상태이고
- **When** `moai github runner register`을 실행하면
- **Then** `gh api`로 등록 토큰이 자동 발급된다

**AC-004.2** (Token No Disk Write):
- **Given** 등록 토큰이 발급되고
- **When** 등록 프로세스가 실행되면
- **Then** 토큰이 어떤 파일에도 기록되지 않는다 (테스트로 검증)

**AC-004.3** (Replace Flag):
- **Given** 동일 이름의 기존 runner가 존재하고
- **When** register를 실행하면
- **Then** `--replace` 플래그로 기존 runner가 자동 대체된다

**AC-004.4** (Settings Link):
- **Given** 등록이 성공하면
- **When** 출력을 확인하면
- **Then** `https://github.com/{owner}/{repo}/settings/actions/runners` 링크가 표시된다

---

### REQ-CI-005: Runner 30-Day Deprecation Window

**AC-005.1** (Warning at 25 Days):
- **Given** runner 버전이 25일 이상 경과하고
- **When** `moai github runner status`을 실행하면
- **Then** WARN 상태와 "Upgrade required within N days" 메시지가 표시된다

**AC-005.2** (Error at 30 Days):
- **Given** runner 버전이 30일 초과하고
- **When** `moai github runner status`을 실행하면
- **Then** ERROR 상태로 격상된다

**AC-005.3** (Doctor Warning):
- **Given** runner 버전이 25일 이상 경과하고
- **When** `moai doctor`을 실행하면
- **Then** runner version diagnostic 항목에 WARN/FAIL이 표시된다

**AC-005.4** (SessionStart Notification):
- **Given** runner 버전이 25일 이상 경과하고
- **When** Claude Code 세션이 시작되면
- **Then** SessionStart 훅에서 1일 1회 알림이 표시된다

**AC-005.5** (Upgrade Flow):
- **Given** `moai github runner upgrade`을 실행하면
- **When** 업그레이드가 진행되면
- **Then** svc.sh stop → 다운로드 → 압축 해제 → svc.sh start 순서로 진행된다

---

### REQ-CI-006: GitHub-hosted Runner Fallback

**AC-006.1** (Skip Self-hosted Install):
- **Given** `runner_mode: github-hosted` 설정이 있고
- **When** init을 실행하면
- **Then** self-hosted runner 설치 단계가 생략된다

**AC-006.2** (Ubuntu-latest Rendering):
- **Given** github-hosted 모드이고
- **When** 워크플로우가 렌더링되면
- **Then** `runs-on: ubuntu-latest`로 설정된다

**AC-006.3** (Codex Disabled):
- **Given** github-hosted 모드이고
- **When** Codex 워크플로우를 확인하면
- **Then** 자동 비활성화된다

**AC-006.4** (Cost Warning):
- **Given** github-hosted 모드이고
- **When** init이 실행되면
- **Then** macOS runner 비용 경고가 출력된다

**AC-006.5** (Hybrid Mode):
- **Given** `runner_mode: hybrid` 설정이 있고
- **When** LLM별 runner_mode를 확인하면
- **Then** per-LLM 설정이 우선 적용된다

---

### REQ-CI-007: Public Repo Codex Rejection

**AC-007.1** (Workflow Guard):
- **Given** Codex 워크플로우 템플릿이 존재하고
- **When** YAML을 확인하면
- **Then** `if: github.event.repository.private == true` job-level 가드가 포함된다

**AC-007.2** (Auth Command Guard):
- **Given** repo가 public 상태이고
- **When** `moai github auth codex`을 실행하면
- **Then** 즉시 중단 + OpenAI 정책 링크가 출력된다

**AC-007.3** (Force-Public Block):
- **Given** `--force-public` 플래그를 사용하고
- **When** `moai github auth codex --force-public`을 실행하면
- **Then** 여전히 차단된다 (HARD compliance)

---

### REQ-CI-008: Claude Auth Bootstrap

**AC-008.1** (Token No Disk):
- **Given** Claude OAuth 토큰이 입력되고
- **When** auth 프로세스가 실행되면
- **Then** 토큰이 임시 파일에 기록되지 않는다

**AC-008.2** (Secret Registration):
- **Given** 토큰이 입력되고
- **When** `gh secret set CLAUDE_CODE_OAUTH_TOKEN`이 실행되면
- **Then** `gh secret list`에 CLAUDE_CODE_OAUTH_TOKEN이 표시된다

**AC-008.3** (Max Plan Notice):
- **Given** 인증이 완료되면
- **When** 출력을 확인하면
- **Then** Max plan 구독 필수 안내가 포함된다

---

### REQ-CI-009: Codex Auth Bootstrap

**AC-009.1** (Private Check):
- **Given** repo가 private이고
- **When** `moai github auth codex`을 실행하면
- **Then** 정상적으로 진행된다

**AC-009.2** (Auth.json Permissions):
- **Given** `~/.codex/auth.json`이 생성되고
- **When** 파일 권한을 확인하면
- **Then** 600 권한이 설정된다

**AC-009.3** (Conditional Seed Pattern):
- **Given** codex-bootstrap composite action이 존재하고
- **When** YAML을 확인하면
- **Then** `if [ ! -f "$CODEX_HOME/auth.json" ]; then ... fi` 가드가 포함된다

---

### REQ-CI-010: Gemini Auth Bootstrap

**AC-010.1** (Key Validation):
- **Given** API key가 입력되고
- **When** 형식 검증이 실행되면
- **Then** 영숫자 + `-_`, 길이 39자 조건이 확인된다

**AC-010.2** (Key Masking):
- **Given** API key `AIzaSyB-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`가 입력되고
- **When** 출력을 확인하면
- **Then** `A***xxxx` 형식으로 마스킹된다

**AC-010.3** (Free Tier Notice):
- **Given** 인증이 완료되면
- **When** 출력을 확인하면
- **Then** 무료 tier 한도 (500 RPD) 안내가 포함된다

---

### REQ-CI-011: GLM Auth Bootstrap

**AC-011.1** (Token Input):
- **Given** `moai github auth glm`을 실행하면
- **When** 프롬프트가 표시되면
- **Then** masked input으로 GLM auth token 입력이 가능하다

**AC-011.2** (Env Variables):
- **Given** setup-glm-env composite action이 실행되고
- **When** 환경변수를 확인하면
- **Then** ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN, DISABLE_BETAS=1, DISABLE_PROMPT_CACHING=1이 주입된다

**AC-011.3** (SPEC-GLM-001 Compliance):
- **Given** GLM 워크플로우가 활성화되고
- **When** 환경변수 값을 확인하면
- **Then** SPEC-GLM-001의 DISABLE_BETAS, DISABLE_PROMPT_CACHING 값과 일치한다

---

### REQ-CI-012: 16-Language Neutral Workflow Templates

**AC-012.1** (Template Location):
- **Given** 워크플로우 템플릿이 존재하고
- **When** 파일 경로를 확인하면
- **Then** `internal/template/templates/.github/workflows/*.yml.tmpl`에 위치한다

**AC-012.2** (Language Detection):
- **Given** detect-language composite action이 실행되고
- **When** 16개 project_markers (go.mod, package.json, pyproject.toml, Cargo.toml, pom.xml, build.gradle, *.sln, Gemfile, composer.json, mix.exs, CMakeLists.txt, build.sbt, *.Rproj, pubspec.yaml, Package.swift, meson.build)를 확인하면
- **Then** 모두 동등하게 감지된다 (PRIMARY 없음)

**AC-012.3** (Graceful Skip):
- **Given** 16개 언어 중 하나도 감지되지 않고
- **When** 워크플로우가 실행되면
- **Then** 빌드/테스트 step을 skip하고 LLM 리뷰만 실행한다

---

### REQ-CI-013: PR Open Auto Panel Trigger

**AC-013.1** (Default Panel):
- **Given** llm-panel.yml 템플릿이 배포되고
- **When** 기본 설정을 확인하면
- **Then** auto_panel.llms가 `[claude, gemini]`로 설정된다

**AC-013.2** (Concurrency Groups):
- **Given** PR이 열리고
- **When** 다중 LLM job이 실행되면
- **Then** 각 job이 독립 concurrency group (`llm-panel-{llm}-{pr_number}`)을 사용한다

**AC-013.3** (Paths Ignore):
- **Given** `.github/**` 경로만 변경된 PR이 열리고
- **When** 트리거가 평가되면
- **Then** 패널 워크플로우가 실행되지 않는다

**AC-013.4** (Comment Format):
- **Given** 패널 리뷰가 완료되고
- **When** PR 코멘트를 확인하면
- **Then** 단일 코멘트에 LLM별 `### Claude Review`, `### Gemini Review` 섹션 헤더가 포함된다

---

### REQ-CI-014: PR Comment Single LLM Trigger

**AC-014.1** (Author Guard):
- **Given** PR 코멘트의 `author_association`이 OWNER/MEMBER/COLLABORATOR가 아니고
- **When** 워크플로우 트리거가 평가되면
- **Then** LLM 워크플로우가 실행되지 않는다

**AC-014.2** (Multi-Mention):
- **Given** 코멘트에 `@claude @gemini-cli`가 포함되고
- **When** 트리거가 평가되면
- **Then** claude와 gemini 워크플로우가 모두 병렬 실행된다

**AC-014.3** (Command Arguments):
- **Given** 코멘트에 `@claude review --focus security`가 포함되고
- **When** Claude 워크플로우가 실행되면
- **Then** `--focus security` 인자가 prompt에 전달된다

---

### REQ-CI-015: Custom Panel LLM Set

**AC-015.1** (Config Override):
- **Given** `github-actions.yaml`에서 `auto_panel.llms: [claude, gemini, glm]`으로 설정하고
- **When** PR이 열리면
- **Then** claude, gemini, glm 3종이 패널로 실행된다

---

### REQ-CI-016: Workflow Add/Remove/List

**AC-016.1** (Add):
- **Given** `moai github workflow add gemini`을 실행하면
- **When** 완료 후 `.github/workflows/`를 확인하면
- **Then** gemini-review.yml이 존재한다

**AC-016.2** (Remove):
- **Given** `moai github workflow remove gemini`을 실행하면
- **When** 완료 후 `.github/workflows/`를 확인하면
- **Then** gemini-review.yml이 삭제된다

**AC-016.3** (List):
- **Given** 워크플로우가 배포되어 있고
- **When** `moai github workflow list`을 실행하면
- **Then** 파일명, LLM, 트리거 모드 목록이 출력된다

---

### REQ-CI-017: `moai github status`

**AC-017.1** (Runner List):
- **Given** self-hosted runner가 등록되어 있고
- **When** `moai github status`을 실행하면
- **Then** runner 이름, 상태, 버전, 마지막 실행 시각이 출력된다

**AC-017.2** (Workflow List):
- **Given** 워크플로우가 배포되어 있고
- **When** status를 실행하면
- **Then** 활성화된 워크플로우 목록이 출력된다

**AC-017.3** (Secret List):
- **Given** GitHub secret이 등록되어 있고
- **When** status를 실행하면
- **Then** secret 이름만 표시되고 값은 마스킹된다

**AC-017.4** (Deprecation Window):
- **Given** runner가 설치되어 있고
- **When** status를 실행하면
- **Then** 30일 폐기 잔여일이 표시된다

---

### REQ-CI-018: SHA-Pin Enforcement

**AC-018.1** (No Floating Tags):
- **Given** 모든 워크플로우 템플릿이 존재하고
- **When** validator가 실행되면
- **Then** floating tag (`@v1`, `@latest`) 사용이 탐지되지 않는다

**AC-018.2** (Claude Broken Range):
- **Given** Claude Code Action SHA를 확인하고
- **When** v1.0.79-84 커밋 해시와 비교하면
- **Then** 해당 범위의 SHA가 미사용됨이 확인된다

**AC-018.3** (Configurable SHA):
- **Given** `github-actions.yaml`에서 `workflows.claude.action_sha`를 설정하고
- **When** 템플릿이 렌더링되면
- **Then** 해당 SHA가 Action 참조에 사용된다

---

### REQ-CI-019: Template-First Enforcement

**AC-019.1** (Template Location):
- **Given** 워크플로우 파일을 추가하고자 하고
- **When** 파일 위치를 확인하면
- **Then** `internal/template/templates/.github/workflows/`에 추가된다

**AC-019.2** (Deploy API):
- **Given** 템플릿이 templates/에 존재하고
- **When** `moai github init` 또는 `moai update`가 실행되면
- **Then** `internal/template` 패키지의 Deploy API를 통해 동기화된다

---

### REQ-CI-020: runner_mode State-Driven Rendering

**AC-020.1** (Self-hosted):
- **Given** `runner_mode: self-hosted` 설정이 있고
- **When** 워크플로우가 렌더링되면
- **Then** `runs-on: [self-hosted, {os}, {arch}, {project-name}]`으로 설정된다

**AC-020.2** (GitHub-hosted):
- **Given** `runner_mode: github-hosted` 설정이 있고
- **When** 워크플로우가 렌더링되면
- **Then** `runs-on: ubuntu-latest`로 설정된다

---

### REQ-CI-021: Codex auth.json No Overwrite

**AC-021.1** (Conditional Seed):
- **Given** codex-bootstrap composite action이 실행되고
- **When** `$CODEX_HOME/auth.json`이 이미 존재하면
- **Then** 덮어쓰지 않고 건너뛴다

**AC-021.2** (Validator Detection):
- **Given** 템플릿에 `echo > auth.json` 패턴이 포함되고
- **When** validator가 실행되면
- **Then** 무조건 덮어쓰기 패턴으로 error가 보고된다

---

## 3.2 Security Requirements (REQ-SEC-*)

### REQ-SEC-001: Private Repo Enforcement (Codex)

**AC-SEC-001.1** (Job Guard):
- **Given** codex-review.yml.tmpl이 존재하고
- **When** YAML을 확인하면
- **Then** 모든 job에 `if: github.event.repository.private == true` 가드가 포함된다

---

### REQ-SEC-002: Secret No Disk Write

**AC-SEC-002.1** (Token Piping):
- **Given** `moai github auth claude`이 실행되고
- **When** 토큰 처리를 추적하면
- **Then** 토큰이 임시 파일에 기록되지 않고 stdin pipe로 전달된다

**AC-SEC-002.2** (Log Masking):
- **Given** auth 명령이 실행되고
- **When** 로그 출력을 확인하면
- **Then** 토큰 값이 `***`로 마스킹된다

---

### REQ-SEC-003: SHA-Pin Action

**AC-SEC-003.1** (Third-Party SHA):
- **Given** 모든 third-party Action 참조를 확인하고
- **When** validator가 실행되면
- **Then** floating tag가 아닌 40자 commit hash가 사용된다

**AC-SEC-003.2** (GitHub Official):
- **Given** `actions/checkout@v5` 참조를 확인하면
- **Then** floating major tag가 허용되고 dependabot.yml에 등록된다

---

### REQ-SEC-004: Secret Rotation Policy

**AC-SEC-004.1** (Rotation Display):
- **Given** `moai github status`을 실행하면
- **When** secret 정보를 확인하면
- **Then** 각 secret의 권장 회전 주기가 표시된다:
  - CLAUDE_CODE_OAUTH_TOKEN: 6개월
  - CODEX_AUTH_JSON: 월간 점검
  - GEMINI_API_KEY: 6개월
  - GLM_AUTH_TOKEN: 6개월

---

### REQ-SEC-005: Workspace Cleanup

**AC-SEC-005.1** (Cleanup Step):
- **Given** 모든 self-hosted 워크플로우 템플릿을 확인하고
- **When** steps를 검사하면
- **Then** workspace cleanup step이 `if: always()`와 함께 포함된다

**AC-SEC-005.2** (Env Preservation):
- **Given** cleanup step이 실행되고
- **When** 삭제 대상을 확인하면
- **Then** CODEX_HOME/auth.json은 삭제되지 않는다

---

### REQ-SEC-006: Codex drop-sudo

**AC-SEC-006.1** (Flag Presence):
- **Given** Codex 실행 명령을 확인하면
- **When** 플래그를 검사하면
- **Then** `--safety-strategy drop-sudo`가 포함된다

**AC-SEC-006.2** (Windows Excluded):
- **Given** Codex 워크플로우 템플릿을 확인하면
- **When** runner 조건을 확인하면
- **Then** Windows runner가 포함되지 않는다

---

### REQ-SEC-007: Network Egress Validation

**AC-SEC-007.1** (Post-Install Check):
- **Given** `moai github runner install`이 완료되고
- **When** egress 검증이 실행되면
- **Then** 6개 도메인에 대한 HTTPS 연결이 확인된다

**AC-SEC-007.2** (Selective Validation):
- **Given** claude와 gemini만 활성화되고
- **When** egress 검증이 실행되면
- **Then** api.anthropic.com, generativelanguage.googleapis.com만 검증된다

**AC-SEC-007.3** (Firewall Guidance):
- **Given** egress 검증이 실패하고
- **When** 출력을 확인하면
- **Then** 방화벽 설정 안내 메시지가 표시된다

---

### REQ-SEC-008: Audit Log Preservation

**AC-SEC-008.1** (CLI Log):
- **Given** `moai github *` 명령이 실행되고
- **When** 로그를 확인하면
- **Then** `.moai/logs/github-cli-{date}.log`에 기록된다

**AC-SEC-008.2** (Retention):
- **Given** audit log가 존재하고
- **When** 30일이 경과하면
- **Then** system.yaml retention 정책에 따라 정리된다

---

### REQ-SEC-009: Token Expiry Detection

**AC-SEC-009.1** (401 Error Handling):
- **Given** LLM API가 401 응답을 반환하고
- **When** 워크플로우가 실행되면
- **Then** `AUTH_EXPIRED: rotate {SECRET_NAME} via 'moai github auth {llm}'` 오류 메시지로 job이 종료된다

**AC-SEC-009.2** (PR Comment):
- **Given** Architecture A (comment trigger)에서 401이 발생하고
- **When** 오류 처리가 실행되면
- **Then** PR에 review comment로 알림이 게시된다

---

### REQ-SEC-010: External Contributor Abuse Prevention

**AC-SEC-010.1** (Author Association Guard):
- **Given** 모든 워크플로우 템플릿을 확인하고
- **When** trigger 조건을 검사하면
- **Then** `author_association in ('OWNER', 'MEMBER', 'COLLABORATOR')` 가드가 포함된다

---

## Summary

| Category | Count |
|----------|-------|
| Total REQs | 31 |
| Total Acceptance Criteria | 60+ |
| Functional ACs | 45+ |
| Security ACs | 15+ |
| Observable (test-output/file-exist/metric) | 100% |
