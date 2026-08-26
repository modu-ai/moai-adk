---
title: config 섹션 레퍼런스
weight: 71
draft: false
description: ".moai/config/sections/ 의 주요 설정 파일(handoff/delegation/llm/statusline/security/workflow/crosssession) 키 레퍼런스."
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
      high: "glm-5.3-flash"   # 1M context — Opus 슬롯
      medium: "glm-5.3-flash" # 1M context   — Sonnet 슬롯
      low: "glm-5.3-flash"    # 1M context   — 경량 슬롯
      fable: "glm-5.3-flash"
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

statusline 테마, `github` 세그먼트가 셀 호스팅 서비스, 그리고 16개 세그먼트 토글을 제어합니다.

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  # forge: gitlab             # github | gitlab | none (미지정이면 origin 호스트로 판별)
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
| `forge` | `github` · `gitlab` · `none` 중 하나. `github` 세그먼트가 열린 작업을 어느 호스팅 서비스에서 세는지 정합니다. 값을 비워 두면 origin 원격의 호스트로 판별합니다 — `github.com`이면 `gh`, `gitlab.com`이면 `glab`. 자체 호스팅 인스턴스는 이름에 단서가 없으므로 반드시 직접 적어야 합니다. 인식할 수 없는 값은 판별로 되돌아가지 않고 아무것도 렌더하지 않습니다 — 오타는 틀린 숫자가 아니라 세그먼트 부재로 드러납니다 |
| `segments` | 16개 세그먼트를 하나씩 켜고 끄는 토글 (런타임에 조절할 수 있는 유일한 값). 모두 기본 on이며, 꺼진 세그먼트는 아무것도 출력하지 않고 조용히 빠짐 |

세그먼트는 세 줄에 나눠 배치합니다 — 1행(모델·버전·세션 메타), 2행(컨텍스트 윈도우·API 사용량 바), 3행(디렉터리·git·워크플로우·PR).

관련: [Statusline 시스템 및 PR 세그먼트](/ko/advanced/statusline).

## security.yaml — 보안 강화

내장 `DefaultSecurityPolicy` 패턴을 **덧붙이는** (교체가 아닌) 추가 보안 설정입니다. SOLID의 개방-폐쇄 원칙에 따라 core를 건드리지 않고 config만으로 확장합니다.

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

## workflow.yaml — branch_guard

주 체크아웃(primary checkout)의 브랜치 상태를 지키는 opt-in 가드입니다. 한 체크아웃을 여러 세션이 동시에 쓸 때, 한쪽에서 실행한 `git switch` · `git checkout` · `git reset --hard` · `git stash` · `git rebase` 는 다른 세션의 작업 트리를 아무 신호 없이 바꿔 놓습니다. 이 가드는 그 명령들을 주 체크아웃에서만 거부합니다.

```yaml
workflow:
    branch_guard:
        enabled: false   # 배포 기본값
```

| 키 | 값 | 설명 |
|----|-----|------|
| `enabled` | `false` (기본) | 가드가 완전히 비활성입니다. 판정을 위한 `git rev-parse` 조차 실행하지 않으므로 부가 비용이 없습니다 |
| `enabled` | `true` | 주 체크아웃에서 브랜치 상태를 바꾸는 명령을 거부합니다. 워크트리 안에서는 그대로 허용됩니다 |

**기본값이 꺼짐인 이유.** 이 가드가 막는 위험은 한 체크아웃을 여러 세션이 공유할 때만 생깁니다. 혼자 쓰는 저장소에서는 발생하지 않는 문제이므로, 배포판은 가드를 끈 채로 나갑니다. 여러 세션을 동시에 굴리는 저장소의 관리자가 위 키를 직접 적어 켭니다.

**동작 범위.** 가드는 주 체크아웃과 워크트리를 구분해서, 워크트리 안의 브랜치 조작은 막지 않습니다. `git status` · `git log` · `git diff` · `git fetch` 같은 읽기 명령과 `git stash list` · `git merge-base` 는 켜져 있어도 통과합니다.

**예외와 실패 방향.** 브랜치를 만들어야 하는 git 담당 에이전트는 신원으로 예외 처리되며, `MOAI_BRANCH_GUARD_EXEMPT=1` 환경변수로도 우회할 수 있습니다. 판정이 불확실할 때(저장소가 아님, git 실행 실패 등)는 막지 않고 통과시킨 뒤 감사 로그만 남깁니다 — 확실한 근거가 있을 때만 거부합니다.

브랜치를 바꿔야 하는 작업은 막는 대신 워크트리로 옮기는 것이 정석입니다. 자세한 절차는 [moai worktree](/ko/cli-reference/worktree/) 를 참조하세요.

## workflow.yaml — audit

크로스모델 감사 백엔드(`codex_audit` · `glm_audit` · `audit_multi`)가 실제로 어떤 모델과 effort로 돌지 지정합니다. 백엔드마다 `{model, effort}` 쌍 하나이고, 배포 기본값은 전부 비어 있습니다.

```yaml
workflow:
    audit:
        codex:
            model: ""   # 예: gpt-5.6-sol — codex가 서빙할 수 있는 모델 id
            effort: ""  # 예: high — low | medium | high | xhigh | max
        glm:
            model: ""   # 예: glm-5.3
            effort: ""  # 예: max — low | high | max (z.ai reasoning 상태 이름)
```

| 키 | 설명 |
|----|------|
| `audit.codex.{model, effort}` | codex 감사가 세션 열기와 리뷰 턴에 실어 보내는 쌍. model이 비었거나 codex가 서빙할 수 없는 id면 핀을 버리고 기존 SSOT 해석으로 되돌아갑니다 |
| `audit.glm.{model, effort}` | GLM 감사가 z.ai 요청에 실어 보내는 쌍. effort는 z.ai reasoning 상태 이름 `low` · `high` · `max`만 받으며, 그 외 값을 적으면 reasoning 지시어를 빼고 모델 핀만 적용합니다 |
| 두 쌍 모두 비었을 때 | 이 키가 생기기 전과 동일하게 해석합니다. 핀을 적은 적 없는 프로젝트는 아무 것도 바뀌지 않습니다 |

핀은 감사 진입점에만 적용됩니다. `codex_task` · `glm_task` 같은 작업 위임 경로의 모델 해석은 영향을 받지 않으며, 웹 콘솔의 Audit 패널에서 같은 필드를 직접 편집할 수도 있습니다.

## workflow.yaml — todo

백로그 큐(todo)의 **안내 표면**을 끄는 스위치입니다. 세션이 시작될 때 큐 요약 줄이 뜨는 것, 상태 표시줄의 TODO 세그먼트, 자연어 요청을 todo 워크플로우로 추론 라우팅하는 것 — 이 세 가지가 이 키 하나로 조용해집니다.

```yaml
workflow:
    todo:
        enabled: false   # 명시적 끔 — 키가 없으면 켜진 것으로 봅니다
```

| 키 | 값 | 설명 |
|----|-----|------|
| `todo.enabled` | (키 없음, 기본) | 켜짐. 배포 템플릿에는 이 블록이 아예 없어서, 대부분의 프로젝트가 사는 상태입니다 |
| `todo.enabled` | `false` | 세션 시작 요약 · 상태 표시줄 TODO 세그먼트 · 스킬의 자동 라우팅이 꺼집니다 |

**끄더라도 명령은 지워지지 않습니다.** `moai todo` CLI는 등록된 채로 남고 모든 동사가 동작하며, 이름으로 직접 부른 `/moai todo`도 그대로 실행됩니다. 이 경계는 의도된 것입니다 — 스위치는 "모르는 사용자에게 큐를 보여 주는 표면"만 끄고, 큐를 실제로 쓰는 사람의 경로는 건드리지 않습니다. 설정 파일을 읽지 못하는 실패 경로에서도 켜짐으로 해석됩니다(fail-open).

크기가 크지 않은 일회성 프로젝트에서 세션마다 뜨는 백로그 요약이 소음으로 느껴질 때 이 키를 씁니다. 큐 자체의 운영법은 [moai todo](/ko/utility-commands/moai-todo/) 페이지에 있습니다.

## crosssession.yaml — 세션 간 메시지

내 다른 Claude Code 세션이 보내는 메시지를 이 세션이 어떻게 다룰지 정합니다. `moai cc` · `moai glm` · `moai cg` 런처가 실행 시점에 이 값을 임시 `--settings` 파일로 옮겨 담고, 웹 콘솔은 설정 seam을 통해 이 파일을 편집합니다. 런처를 거치지 않고 맨손으로 `claude`를 실행한 세션은 이 파일을 읽지 않습니다.

```yaml
crosssession:
  inbound: ""             # "" | accept | hold | refuse
  isolate_machines: false # 기본값 — 머신 밖으로 나가는 메시지에 승인을 요구하지 않음
  dialog_expiry: ""       # "" | 60s | 5m | 10m | never
```

| 키 | 값 | 설명 |
|----|-----|------|
| `inbound` | `""`(기본) | Claude Code가 두 세션의 권한 모드 등급을 보고 메시지마다 스스로 판단합니다 |
| `inbound` | `accept` | 메시지를 그대로 전달합니다. 사람이 지켜보지 않는 작업 세션에 메시지를 받게 하려면 이 값이 필요합니다 |
| `inbound` | `hold` | 메시지마다 승인을 받도록 붙잡아 둡니다. 이렇게 붙잡힌 메시지는 만료되지 않고, `accept`가 적용될 때 전달됩니다 |
| `inbound` | `refuse` | 메시지를 버립니다 |
| `isolate_machines` | `false` (기본) | **이 머신 밖의 세션으로 메시지가 나갈 때 승인을 요구하지 않습니다.** 같은 머신 안의 세션끼리는 어떤 값이든 머신을 벗어나지 않지만, 머신 밖의 세션으로 가는 메시지는 Anthropic 서버를 거칩니다 — 그 경로를 승인 없이 허용한다는 뜻이므로 기본값을 그대로 둘지 판단해서 결정하세요 |
| `isolate_machines` | `true` | 메시지가 머신을 벗어나기 전에 반드시 승인을 받습니다(`bypassPermissions` 모드에서도). 어느 설정 범위든 하나가 `true`면 적용되므로, 저장소에 체크인된 프로젝트 파일이 요구를 켤 수는 있어도 끌 수는 없습니다 — 끄려면 `true`를 적은 범위를 전부 지워야 합니다 |
| `dialog_expiry` | `""`(기본) | Claude Code의 5분 기본값을 그대로 씁니다 |
| `dialog_expiry` | `60s` · `5m` · `10m` · `never` | 기본 판단으로 붙잡힌 메시지의 승인 창 기한입니다. `never`는 세션이 끝날 때까지 붙잡아 둡니다. `inbound: hold`로 명시해 붙잡은 메시지에는 적용되지 않습니다 |

## 관련 문서

- [settings.json 가이드](/ko/advanced/settings-json) — Claude Code 런타임 설정
- [하네스 프로필과 평가](/ko/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/ko/cli-reference/doctor) — `moai doctor config`로 병합 설정 검사
