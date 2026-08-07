---
title: config 섹션 레퍼런스
weight: 71
draft: false
description: ".moai/config/sections/ 의 주요 설정 파일(handoff/delegation/llm/statusline/security) 키 레퍼런스."
---

MoAI-ADK의 프로젝트 설정은 `.moai/config/sections/` 아래 여러 YAML 파일로 나뉘어 있습니다. [settings.json 가이드](/ko/advanced/settings-json)가 Claude Code 런타임 설정을 다룬다면, 이 페이지는 MoAI-ADK 자체 동작을 제어하는 주요 섹션 파일의 키를 정리합니다.

{{< callout type="info" >}}
**한 줄 요약**: `settings.json`은 Claude Code에게 무엇을 허용할지 정하고, `.moai/config/sections/*.yaml`은 MoAI-ADK가 어떻게 오케스트레이션할지 정합니다.
{{< /callout >}}

## handoff.yaml — auto-resume 핸드오프

세션 경계에서 저장된 핸드오프를 어떻게 처리할지 제어합니다.

```yaml
handoff:
    mode: manual   # manual | auto
    guide: false
```

| 키 | 값 | 설명 |
|----|-----|------|
| `mode` | `manual` (기본) | 저장된 핸드오프를 자동 주입하지 않음 (opt-in 베이스라인 UX) |
| `mode` | `auto` | `/clear` 시 저장된 핸드오프를 세션 컨텍스트로 주입한 뒤 audit-trail 복사본으로 이동 |
| `guide` | `false` (기본) | `true`면 `/clear`가 아닌 세션 시작(startup/resume/compact)에서 대기 중인 핸드오프가 있다는 best-effort stderr 힌트를 띄움. 알림일 뿐 세션을 막지는 않음 |

관련: [자율 연속 루프](/ko/advanced/autonomous-loops), [moai handoff](/ko/cli-reference/handoff).

## delegation.yaml — 에이전트 라우팅 SSOT

`/moai` 서브커맨드마다 기본으로 쓸 스킬과 에이전트를 배정한 맵입니다. 오케스트레이터는 실행 계획을 짤 때(Analyze-First) 이 맵을 읽고 어떤 에이전트를 spawn할지, 어떤 스킬을 주입할지 정합니다.

```yaml
delegation:
    version: 1
    learning:
        observe: routing-ledger
        propose_via: harness-tier-ladder
        auto_apply: false          # Tier-4 게이트 — 사용자 승인 필요
    subcommands:
        plan:
            agents: [manager-spec, plan-auditor, Explore]
            skills: [moai-workflow-spec, moai-foundation-thinking]
        # run / sync / project / fix / loop / ...
    domain_skills:
        backend:  [moai-ref-api-patterns, moai-domain-backend]
        security: [moai-ref-owasp-checklist, moai-ref-llm-security, ...]
    agents:
        manager-spec: [moai-workflow-spec, moai-foundation-thinking]
```

| 블록 | 설명 |
|------|------|
| `learning` | 라우팅 사용 내역을 append-only 원장(`.moai/state/routing-ledger.jsonl`, opt-in·fail-open)에 남기고, 하네스 학습 서브시스템이 4-tier 제안 사다리로 갱신을 제안. `auto_apply: false` — Tier-4 변경은 `AskUserQuestion` 사용자 승인 필요 |
| `subcommands` | 서브커맨드마다 `agents`(spawn할 11개 retained 에이전트) + `skills`(spawn 시 주입할 workflow 스킬). 하나도 배정하지 않아도 유효 (오케스트레이터가 직접 실행) |
| `domain_skills` | 미션 도메인에 맞춰 주입할 스킬 (spawn당 0-3개). 도메인 신호와 매칭 |
| `agents` | 에이전트마다 두는 conditional 스킬 (트리거가 발생하면 on-demand 로드) |

관련: [에이전트 가이드](/ko/advanced/agent-guide), [스킬 가이드](/ko/advanced/skill-guide).

## llm.yaml — 백엔드·프로필 매트릭스

프로필, 프로필 매트릭스, 에이전트별 override, GLM 모델 매핑을 정의합니다.

```yaml
llm:
  profile: "medium"            # high | medium | low (활성 매트릭스 열, max는 high로 읽힘)
  performance_tier: "medium"   # legacy 별칭 (profile 부재 시 읽힘, 동일 어휘)
  profiles:                    # 프로필 열 → 11개 에이전트 → {model, effort}
    high: { ... }              # 상세 표: 프로필 매트릭스 페이지
    medium: { ... }
    low: { ... }
  agent_overrides: {}          # 에이전트별 {model, effort} override (선택)
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.2"          # 1M context — Opus 슬롯
      medium: "glm-4.7"        # 202K context — Sonnet 슬롯
      low: "glm-4.5-air"       # 128K context — 경량 슬롯
      fable: "glm-5.2"
```

| 키 | 설명 |
|----|------|
| `profile` | 활성 프로필 매트릭스 열 (`high`/`medium`/`low`, 과거 `max`는 `high`의 별칭으로 읽힘). 비어 있으면 `medium`으로 해석. 모든 서브에이전트 spawn의 model+effort 출처 |
| `performance_tier` | legacy 별칭 필드. `profile`이 없을 때만 읽히며, `high`/`medium`/`low` 어휘를 그대로 쓰므로 별도 정규화가 필요 없음 |
| `profiles` | 프로필 열마다 에이전트 → `{model, effort}`를 적은 매트릭스 (에이전트 11개 × 열 3개 = 33셀). 빠진 셀은 Go 기본값(`template.DefaultProfileMatrix`)이 최종 fallback |
| `agent_overrides` | 정규 에이전트 이름을 키로 하는 `{model, effort}` override. 활성 프로필의 에이전트 셀보다 우선 (카탈로그+enum 검증) |
| `glm.base_url` | Z.AI Anthropic 호환 프록시 엔드포인트 |
| `glm.models` | 슬롯별 GLM 모델 매핑. GLM은 Claude의 5단계 effort를 3개 reasoning 상태(thinking-off / reasoning-high / reasoning-max)로 collapse |

관련: [프로필 매트릭스](/ko/advanced/profile-matrix), [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier).

## statusline.yaml — 상태 표시줄

statusline 테마와 16개 세그먼트 토글을 제어합니다.

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  segments:
    model: true
    context: true
    # ... 총 16개 세그먼트 (모두 기본 on)
    task: true
    pr: true
```

| 키 | 설명 |
|----|------|
| `theme` | 테마는 딱 두 개: `catppuccin-mocha`(기본) 또는 `catppuccin-latte` |
| `segments` | 16개 세그먼트를 하나씩 켜고 끄는 토글 (런타임에 조절할 수 있는 유일한 값). 모두 기본 on이며, 꺼진 세그먼트는 아무것도 출력하지 않고 조용히 빠짐 |

세그먼트는 세 줄에 나눠 배치합니다 — 1행(모델·버전·세션 메타), 2행(컨텍스트 윈도우·API 사용량 바), 3행(디렉터리·git·워크플로우·PR).

관련: [Statusline 시스템 및 PR 세그먼트](/ko/advanced/statusline).

## security.yaml — 보안 강화

내장 `DefaultSecurityPolicy` 패턴을 **덧붙이는**(교체가 아닌) 추가 보안 설정입니다. SOLID의 개방-폐쇄 원칙에 따라 core를 건드리지 않고 config만으로 확장합니다.

```yaml
security:
  extra_dangerous_bash_patterns:
    - 'curl\s+.*\|\s*(ba)?sh'
    - 'rm\s+-rf\s+/[^.]'
  extra_deny_patterns: []
  extra_ask_patterns: []
  permission:
    strict_mode: true
    session_rules: []
  sandbox:
    required: false
    network_allowlist: []
    env_scrub_extra: []
    docker_image: "alpine:latest"
```

| 키 | 설명 |
|----|------|
| `extra_dangerous_bash_patterns` | 내장 deny 패턴에 **덧붙일** 위험 Bash 커맨드 정규식 (대소문자 무시) |
| `extra_deny_patterns` / `extra_ask_patterns` | 추가로 둘 파일 deny/ask 패턴 |
| `permission.strict_mode` | `true`면 bypassPermissions 모드의 에이전트 spawn을 거부 |
| `sandbox.required` | `true`면 `sandbox: none` 에이전트를 `sandbox.justification` 없이 거부 (기본 false) |
| `sandbox.network_allowlist` | 기본 8개 호스트에 **덧붙일** 허용 네트워크 호스트 |
| `sandbox.env_scrub_extra` | 기본 scrub 목록에 **덧붙일** env 변수명 (AWS_*, GITHUB_TOKEN 등) |
| `sandbox.docker_image` | docker 백엔드 기본 이미지 |

관련: [보안 노트](/ko/advanced/security-notes), [settings.json 가이드](/ko/advanced/settings-json).

## ralph.yaml — Ralph Engine

`/moai loop` 의 진단 기반 반복 수정 루프(Ralph Engine) 동작을 제어합니다.

```yaml
ralph:
  max_iterations: 5         # 반복 상한 (기본 5; CLI --max 가 우선)
  auto_converge: true       # 정체 감지 시 자동 수렴
  human_review: true        # 리뷰 단계에서 사람 개입 중단
  lint_as_instruction: true # LSP diagnostic 을 다음 턴 지침으로 주입
  warn_as_instruction: false # 에러가 없을 때만 warning 도 주입
```

| 키 | 설명 |
|----|------|
| `max_iterations` | 반복 상한 (기본 5). 우선순위: CLI `--max` 플래그 > `ralph.max_iterations` > `workflow.yaml loop_prevention.max_iterations` |
| `auto_converge` | N 턴 연속 무진전 시 자동으로 수렴 판정 |
| `human_review` | 리뷰 단계에서 사람이 들여다보게 중단 |
| `lint_as_instruction` | LSP diagnostic 을 `systemMessage` 로 주입해 AI 가 다음 프롬프트로 받게 함 (기본 true) |
| `warn_as_instruction` | 에러가 없을 때 warning 도 함께 주입 (기본 false) |

관련: [/moai loop](/ko/utility-commands/moai-loop), [자율 연속 루프](/ko/advanced/autonomous-loops).

## harness.yaml — 하네스 깊이·평가

하네스 품질 파이프라인 깊이(minimal/standard/thorough)와 자동 감지, 평가자 memory scope, 에스컬레이션을 정의합니다.

```yaml
harness:
  default_profile: "default"  # default | strict | lenient | frontend
  evaluator:
    memory_scope: per_iteration  # FROZEN — 변경 불가 (design-constitution §11.4.1)
  mode_defaults:
    solo: auto
    team: auto
    cg: thorough                 # CG 모드는 항상 thorough
  auto_detection:
    enabled: true
    rules:
      minimal:  # file_count <= 3 AND single_domain, ...
      standard: # file_count > 3 OR multi_domain, ...
      thorough: # security/payment 키워드, critical 우선순위, ...
  escalation:
    enabled: true
```

| 블록 | 설명 |
|------|------|
| `default_profile` | SPEC 에 `evaluator_profile` 이 없을 때 쓸 기본 평가 프로필 |
| `evaluator.memory_scope` | 평가자 메모리 범위. `per_iteration` 으로 고정(FROZEN) |
| `mode_defaults` | 실행 모드(solo/team/cg)별 기본 깊이 |
| `auto_detection.rules` | Complexity Estimator 가 minimal/standard/thorough 로 자동 분류하는 조건 |
| `escalation` | 실패 시 상위 깊이로 에스컬레이션 |

관련: [하네스 프로필과 평가](/ko/advanced/harness-profiles), [moai harness](/ko/cli-reference/harness).

## quality.yaml — 품질 게이트·개발 방법론

개발 모드(DDD/TDD), 커버리지 목표, LSP 품질 게이트 임계값, 품질 게이트 강제 여부를 제어합니다.

```yaml
constitution:
  development_mode: tdd        # tdd | ddd | off
  enforce_quality: true
  test_coverage_target: 85
  lsp_quality_gates:
    enabled: true
    plan: { require_baseline: true }
    run:  { max_errors: 0, max_type_errors: 0, max_lint_errors: 0, allow_regression: false }
    sync: { max_errors: 0, max_warnings: 10, require_clean_lsp: true }
```

| 블록 | 설명 |
|------|------|
| `development_mode` | `tdd` / `ddd` / `off`. `/moai run` 의 cycle_type 기본값 결정 |
| `enforce_quality` | `true` 면 품질 게이트 위반 시 run-phase 가 GREEN 되지 않음 |
| `test_coverage_target` | 패키지 커버리지 최소 목표(%). 임계 패키지(cli/template/hook)는 90%+ 권장 |
| `lsp_quality_gates.plan` | plan-phase: LSP baseline 캡처 여부 |
| `lsp_quality_gates.run` | run-phase: 에러/타입에러/린트에러 0, 회귀 금지 |
| `lsp_quality_gates.sync` | sync-phase: 에러 0, warning ≤ 10, clean LSP 요구 |

관련: [SPEC 워크플로우](/ko/advanced/spec-workflow), [LSP 게이트](/ko/advanced/lsp-gates).

## workflow.yaml — 워크플로우 상태

워크플로우 임계값, branch-state guard opt-in, 세션 worktree opt-in, agentic 루프 상한을 제어합니다.

```yaml
workflow:
  agentic_loop:
    max_iterations: 10        # 파이프라인 수준 completion-loop 상한
  branch_guard:
    enabled: false            # distributed default — hook 이 INERT (SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001)
  session_worktree:
    enabled: false            # distributed default — auto-isolation INERT
  loop_prevention:
    max_iterations: 5         # per-operation 진단 fix-loop 상한 (ralph.max_iterations 과 별개 축)
```

| 블록 | 설명 |
|------|------|
| `agentic_loop.max_iterations` | 파이프라인 수준 completion-loop 상한 (`AgenticLoopConfig`) |
| `branch_guard.enabled` | Main-Checkout Branch-State Guard 활성화. **분산 기본 `false`** — 공유 체크아웃이 아닌 1인 개발 환경에서는 hook 이 평가하지 않음. 공유 다중 세션 체크아웃 운영자가 opt-in |
| `session_worktree.enabled` | `moai init`/`moai profile`/`moai web` 시 자동 worktree 격리 활성화. **분산 기본 `false`** (SPEC-SESSION-WORKTREE-001). `MOAI_SESSION_WORKTREE` env 로 재정의 |
| `loop_prevention.max_iterations` | per-operation 진단 fix-loop 상한. `ralph.max_iterations` 과 별개 축 — ralph.yaml 이 우선 |

> branch_guard 와 session_worktree 는 모두 분산 기본값이 `false` 입니다. 단일 사용자 리포에서 이 값들이 `false` 인 것은 의도된 동작이지 결함이 아닙니다.

관련: [Main-Checkout Branch Guard](/ko/advanced/branch-guard), [Worktree 통합](/ko/advanced/worktree).

## mx.yaml — @MX 태그 시스템

`@MX` 코드 주석 태그 시스템의 태그 종류, 언어별 감지, 검증 규칙을 정의합니다.

```yaml
mx:
  version: "2.1"
  languages:
    go:      { enabled: auto, patterns: ["*.go"], exclude: ["*_generated.go", "vendor/**"] }
    python:  { enabled: auto, patterns: ["*.py"], exclude: ["**/__pycache__/**"] }
    # ... 16개 언어 동등 (go, python, typescript, javascript, rust, java, kotlin,
    #     csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)
  tags:
    # 각 태그별 설명·활성화 여부
  reason_required:  # @MX:REASON 필수 태그 목록
    - WARN
    - DEBT
```

| 블록 | 설명 |
|------|------|
| `version` | MX 태그 시스템 스키마 버전 |
| `languages` | 16개 언어를 동등하게 나열. 각 언어는 프로젝트 마커(`go.mod`, `pyproject.toml`, `Cargo.toml` 등)로 자동 감지(`enabled: auto`) |
| `tags` | `@MX:NOTE` / `@MX:WARN` / `@MX:ANCHOR` / `@MX:TODO` / `@MX:SPEC` / `@MX:DEBT` / `@MX:LEGACY` 등 태그별 메타데이터. `@MX:SPEC` 은 SPEC 연관 관계를 기록 (SPEC-MX-ASSOCIATION-001) |
| `reason_required` | `@MX:REASON` 필수 태그 목록 (기본: WARN, DEBT) |

> 특정 언어를 "PRIMARY"로 격하하거나 일부만 "enabled" 로 두지 마세요 — 16개 언어는 모두 동등합니다.

관련: [@MX 태그 프로토콜](/ko/advanced/mx-tag-protocol), [moai mx](/ko/cli-reference/mx).

## 환경변수 오버라이드

YAML 섹션 값 일부는 환경변수로 재정의할 수 있습니다. 환경변수가 파일 값보다 우선합니다 (`internal/config/manager.go` 의 `applyEnvOverrides`).

| 환경변수 | 대상 | 설명 |
|----------|------|------|
| `MOAI_DEVELOPMENT_MODE` | `constitution.development_mode` | `tdd`/`ddd`/`off` 중 강제 |
| `MOAI_LOG_LEVEL` | 로그 레벨 | `debug`/`info`/`warn`/`error` |
| `MOAI_LOG_FORMAT` | 로그 포맷 | `text`/`json` |
| `MOAI_NO_COLOR` | 컬러 출력 | `1`/`true` 면 컬러 강제 off |
| `MOAI_CONFIG_DIR` | config 디렉터리 위치 | `.moai/config/` 대신 다른 경로 사용 |

> 위 5개가 config manager 가 실제로 읽는 환경변수 오버라이드의 전부입니다. 상수 정의는 `internal/config/envkeys.go` 에 있습니다. `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` 은 현재 구현되지 않습니다 — 이름과 대화 언어는 `user.yaml`/`language.yaml` 만 읽습니다.

## 관련 문서

- [settings.json 가이드](/ko/advanced/settings-json) — Claude Code 런타임 설정
- [하네스 프로필과 평가](/ko/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/ko/cli-reference/doctor) — `moai doctor config`로 병합 설정 검사
