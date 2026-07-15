---
title: config 섹션 레퍼런스
weight: 71
draft: false
description: ".moai/config/sections/ 의 주요 설정 파일(handoff/delegation/llm/statusline/security) 키 레퍼런스."
---

MoAI-ADK의 프로젝트 설정은 `.moai/config/sections/` 아래 여러 YAML 파일로 나뉘어 있습니다. [settings.json 가이드](/ko/advanced/settings-json)가 Claude Code 런타임 설정을 다룬다면, 이 페이지는 MoAI-ADK 자체 동작을 제어하는 주요 섹션 파일의 키를 정리합니다.

{{< callout type="info" >}}
**한 줄 요약**: `settings.json` 은 Claude Code에게 무엇을 허용할지 정하고, `.moai/config/sections/*.yaml` 은 MoAI-ADK가 어떻게 오케스트레이션할지 정합니다.
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
| `guide` | `false` (기본) | `true` 시 non-`/clear` 세션 시작(startup/resume/compact)에서 대기 중 핸드오프가 있다는 best-effort stderr 힌트 방출. 정보성일 뿐 세션을 막지 않음 |

관련: [자율 연속 루프](/ko/advanced/autonomous-loops), [moai handoff](/ko/cli-reference/handoff).

## delegation.yaml — 에이전트 라우팅 SSOT

`/moai` 서브커맨드별 기본 스킬/에이전트 배정 맵입니다. 오케스트레이터가 실행 계획을 구성할 때(Analyze-First) 이 맵을 읽어 어떤 에이전트를 spawn하고 어떤 스킬을 주입할지 결정합니다.

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
| `learning` | 라우팅 사용을 append-only 원장(`.moai/state/routing-ledger.jsonl`, opt-in·fail-open)으로 관찰하고, 하네스 학습 서브시스템이 4-tier 제안 사다리로 갱신 제안. `auto_apply: false` — Tier-4 변경은 `AskUserQuestion` 사용자 승인 필요 |
| `subcommands` | 서브커맨드별 `agents` (spawn할 11개 retained 에이전트) + `skills` (spawn 시 주입할 workflow 스킬). 0개 배정도 유효 (오케스트레이터 직접 실행) |
| `domain_skills` | 미션 도메인별 주입 스킬 (spawn당 0-3개). 도메인 신호와 매칭 |
| `agents` | 에이전트별 conditional 스킬 (트리거 발생 시 on-demand 로드) |

관련: [에이전트 가이드](/ko/advanced/agent-guide), [스킬 가이드](/ko/advanced/skill-guide).

## llm.yaml — 백엔드·모델 티어

성능 티어, 청구 플랜, Claude/GLM 모델 매핑을 정의합니다.

```yaml
llm:
  performance_tier: "medium"   # high | medium | low
  plan_type: "subscription"    # api | subscription
  claude_models:
    high: "opus"
    medium: "sonnet"
    low: "sonnet"
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.2"          # 1M context — Opus 슬롯
      medium: "glm-4.7"        # 202K context — Sonnet 슬롯
      low: "glm-4.5-air"       # 128K context — Haiku 슬롯
      fable: "glm-5.2"
```

| 키 | 설명 |
|----|------|
| `performance_tier` | 모든 서브에이전트·팀 에이전트의 모델 선택 제어 (high=복잡 추론, medium=균형, low=빠름/저비용) |
| `plan_type` | 청구 플랜 유형. `api`=태스크당 비용 최적화, `subscription`=주간 쿼터 최적화 (비어 있으면 subscription으로 해석) |
| `claude_models` | 티어별 Claude 모델 매핑. 하네스 레벨이 effort로 연결됨 (thorough→xhigh, standard→high, minimal→medium) |
| `glm.base_url` | Z.AI Anthropic 호환 프록시 엔드포인트 |
| `glm.models` | 티어별 GLM 모델 매핑. GLM은 Claude의 5단계 effort를 3개 reasoning 상태(thinking-off / reasoning-high / reasoning-max)로 collapse |

관련: [plan_type 티어 프로필](/ko/advanced/plan-type-profiles), [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier).

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
| `theme` | 정확히 2개 테마 존재: `catppuccin-mocha` (기본) 또는 `catppuccin-latte` |
| `segments` | 16개 세그먼트 개별 토글 (유일한 런타임 레버). 모두 기본 on이며, 비활성 상태는 graceful no-output으로 처리 |

세그먼트는 3개 라인에 배치됩니다 — 라인 1(모델·버전·세션 메타), 라인 2(컨텍스트 윈도우·API 사용량 바), 라인 3(디렉터리·git·워크플로우·PR).

관련: [Statusline 시스템 및 PR 세그먼트](/ko/advanced/statusline).

## security.yaml — 보안 강화

내장 `DefaultSecurityPolicy` 패턴을 **확장**(교체 아님)하는 추가 보안 설정입니다. SOLID의 개방-폐쇄 원칙을 따라 core 수정 없이 config로 확장합니다.

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
| `extra_dangerous_bash_patterns` | 내장 deny 패턴에 **추가**되는 위험 Bash 커맨드 정규식 (대소문자 무시) |
| `extra_deny_patterns` / `extra_ask_patterns` | 추가 파일 deny/ask 패턴 |
| `permission.strict_mode` | `true` 시 bypassPermissions 모드의 에이전트 spawn 거부 |
| `sandbox.required` | `true` 시 `sandbox: none` 에이전트를 `sandbox.justification` 없이는 거부 (기본 false) |
| `sandbox.network_allowlist` | 기본 8개 호스트에 **추가**되는 허용 네트워크 호스트 |
| `sandbox.env_scrub_extra` | 기본 scrub 목록에 **추가**되는 env 변수명 (AWS_*, GITHUB_TOKEN 등) |
| `sandbox.docker_image` | docker 백엔드 기본 이미지 |

관련: [보안 노트](/ko/advanced/security-notes), [settings.json 가이드](/ko/advanced/settings-json).

## 관련 문서

- [settings.json 가이드](/ko/advanced/settings-json) — Claude Code 런타임 설정
- [하네스 프로필과 평가](/ko/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/ko/cli-reference/doctor) — `moai doctor config` 로 병합 설정 검사
