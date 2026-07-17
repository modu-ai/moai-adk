---
title: Statusline 시스템 — 3-line 레이아웃 완전 가이드
weight: 78
draft: false
---

Claude Code와 moai-adk-go 통합을 위한 **커스텀 statusline 시스템**입니다. 토크노믹스는 측정에서 시작합니다. 컨텍스트 사용률(CW%), 프롬프트 캐시 적중률, rate limit 소진율을 터미널 하단에 항상 띄워 두면 토큰 운용 상태를 한눈에 읽을 수 있습니다. Claude Code v2.1.139부터 effort/thinking, v2.1.145부터 workspace.repo + pr 필드가 stdin JSON에 추가되어 더 풍부한 컨텍스트를 표시할 수 있습니다.

> MoAI 워크플로우는 PR 중심입니다. 모든 SPEC은 plan-PR → run-PR → sync-PR 사이클을 생성하므로, statusline에 현재 PR 번호 + 리뷰 상태 + 컨텍스트 사용률 + handoff 권고를 즉시 노출하면 개발 효율이 크게 높아집니다.

## 개요

### 최종 레이아웃 (3-line v3)

```
🤖 Opus │ 🧠 xhigh·t │ ♻️ 87% │ 🔅 v2.1.212 │ 🗿 v3.0.0 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk | 🅱️ main ↑5 +2 │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**: 모델 · effort/thinking · 캐시 히트율 · Claude Code 버전 · MoAI 버전 · 세션 시간 · output style
- **Line 2 (Usage bars)**: CW (context window) · 5H (rolling) · 7D (rolling) — 각 bar는 이모지 + label + bar + % + reset 정보
- **Line 3 (Git/PR)**: 디렉터리 · 리포지토리+브랜치 통합 · git status · 활성 SPEC task · PR 정보

### 데이터 흐름

```
Claude Code (stdin JSON 전달)
    ↓
.moai/status_line.sh (shell wrapper — settings.json statusLine.command)
    ↓
moai statusline (Go binary)
    ↓
internal/statusline/types.go (StdinData 파싱)
    ↓
internal/statusline/builder.go (CollectMemory, CollectMetrics, etc.)
    ↓
internal/statusline/renderer.go (3-line v3 layout)
    ↓
터미널 표시
```

## Line 1 — Info (7 segments)

### Model

- **포맷**: `🤖 <model display name>`
- **데이터 소스**: stdin `model.display_name` (또는 string shorthand)
- **예시**: `🤖 Opus 4.7`, `🤖 Sonnet 4.6`, `🤖 Haiku 4.5`
- **숨김 조건**: `model` field 부재 또는 `data.Metrics.Model == ""`
- **세그먼트 키**: `model`

### Effort / Thinking

- **포맷**: `🧠 <level>[·t]`
- **데이터 소스**: stdin `effort.level` + `thinking.enabled` (Claude Code v2.1.139+)
- **Level 값**: `low` / `medium` / `high` / `xhigh` / `max`
- **`·t` 접미사**: `thinking.enabled == true` 일 때 추가 (extended reasoning 활성)
- **예시**:
  - `🧠 xhigh·t` (xhigh effort + thinking 활성)
  - `🧠 high` (high effort, thinking 없음)
  - `·t` (effort 부재 + thinking만 활성)
- **숨김 조건**: `effort` + `thinking` 모두 부재 (effort.level 빈 문자열 포함)
- **세그먼트 키**: `effort_thinking`

지금 세션이 어느 추론 깊이로 돌고 있는지 항상 확인할 수 있어, 모델 정책이 실제로 적용되는지 검증하는 창이기도 합니다.

### 캐시 히트율

- **포맷**: `♻️ <N>%` (N = cache_read / (cache_read + cache_creation) × 100, 소수점 버림)
- **데이터 소스**: stdin `current_usage.cache_read_tokens` + `current_usage.cache_creation_tokens`
- **예시**: `♻️ 28%` (cache_read 2000, cache_creation 5000 → 2000/7000)
- **숨김 조건**: `current_usage` 부재 · `cache_creation == 0` (fresh cache write 없음) · 둘 다 0 — 값을 지어내지 않고 조용히 생략 (graceful degradation)
- **토글**: `cache_hit: false` in statusline.yaml → 숨김 (default-on)
- **세그먼트 키**: `cache_hit`
- **참고**: 캐시 히트율은 `♻️`, Line 3 Git Status는 `💾`로 이모지가 구분됩니다. prompt-cache 재사용률 모니터링 (SPEC-TOKEN-EFFICIENCY-001 P0-2)

캐시 히트율은 컨텍스트 다이어트의 효과를 보여주는 측정기입니다. 항상 로드되는 지침을 줄이면 이 숫자가 바로 올라가는 것을 볼 수 있습니다.

### Claude Code 버전

- **포맷**: `🔅 v<version>` (3-line 레이아웃에서 실제로 렌더되는 형식)
- **데이터 소스**: stdin `version` 문자열
- **예시**: `🔅 v2.1.212`
- **참고**: 이름 붙은 프리셋(full/compact/minimal)은 폐기되어 세그먼트를 직접 켜고 끕니다 (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001). 과거 full 모드의 `🔅 cc v<version>` 접두 변형은 5-line 레이아웃과 함께 폐기되어 더 이상 렌더되지 않습니다.
- **숨김 조건**: `version` 빈 문자열
- **세그먼트 키**: `claude_version`

### MoAI 버전

- **포맷**: `🗿 v<current>` 또는 업데이트 가능 시 `🗿 v<current> -> 🗿 v<latest>`
- **데이터 소스**: `.moai/config/sections/system.yaml` `moai.version` + 백그라운드 update checker 결과
- **예시**:
  - `🗿 v2.20.0-rc1` (최신)
  - `🗿 v2.18.0 -> 🗿 v2.20.0-rc1` (업데이트 권고)
- **세그먼트 키**: `moai_version`

### 세션 시간

- **포맷**: `⏳ <X>h <Y>m` (≥1h) / `⏳ <X>m` (<1h) / `⏳ <X>d <Y>h` (≥24h)
- **데이터 소스**: stdin `cost.total_duration_ms`
- **예시**: `⏳ 4h 52m`, `⏳ 35m`, `⏳ 1d 3h`
- **세그먼트 키**: `session_time`

### Output Style

- **포맷**: `💬 <style name>`
- **데이터 소스**: stdin `output_style.name`
- **예시**: `💬 MoAI`, `💬 R2-D2`, `💬 default`
- **숨김 조건**: `output_style.name` 빈 문자열
- **세그먼트 키**: `output_style`

## Line 2 — Usage Bars (3 segments)

### CW (Context Window)

- **포맷**: `<icon> CW: <bar> <pct>% [(⚠️/clear) | (🛑/clear!)]`
- **데이터 소스**:
  - bar: `context_window.context_window_size` × auto-compact threshold (default 85%) → scaled budget
  - 퍼센티지: `context_window.used_percentage` (사전 계산) 또는 `current_usage` tokens 합산
  - handoff suffix 활성 조건: `handoffGuideStage(data)` 판정 (아래 2단계 표 참조)
- **배터리 이모지** (`BatteryIcon`, `internal/statusline/gradient.go`):
  - `🔋` (표시 퍼센티지 ≤ 70%)
  - `🪫` (표시 퍼센티지 > 70%)
  - bar 자체는 블록마다 초록 → 노랑 → 빨강 연속 그라디언트 색을 입힙니다 (배터리 임계값과 별개)
- **`(⚠️/clear)` / `(🛑/clear!)` handoff suffix**:
  - 1M context 모델 (Opus 4.8, GLM-5.2): used_percentage ≥50% (raw context_window_size 기준)
  - 200K context 모델 (Sonnet/Haiku): used_percentage ≥90%
  - 의미: 다음 turn 시작 전에 `/clear` 권고 + paste-ready resume message 활용
- **예시**: `🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **세그먼트 키**: `context`

### 5H (5시간 rolling rate limit)

- **포맷**: `🔋 5H: <bar> <pct>% [(<reset>)]`
- **데이터 소스**: stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset 포맷**:
  - <60분: `(Nm)` (예: `(47m)`)
  - <24시간: `(Nh Nm)` (예: `(2h 15m)`)
  - ≥24시간: `(Mon DD)` (예: `(May 28)`)
- **예시**: `🔋 5H: █████░░░░░ 56% (47m)`
- **데이터 부재**: `rate_limits.five_hour == null` → bar 0%, reset `(rolling)`
- **세그먼트 키**: `usage_5h`

### 7D (7일 rolling rate limit)

- **포맷**: `🔋 7D: <bar> <pct>% [(<reset>)]`
- **데이터 소스**: stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset 포맷**: `(Mon DD)` (절대 날짜)
- **예시**: `🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **세그먼트 키**: `usage_7d`

구독 요금제 사용자에게 5H/7D bar는 사실상 예산 게이지입니다. rate limit이 소진되기 전에 무거운 작업을 배치할지, CG 모드로 GLM 워커에 넘길지를 이 두 bar를 보고 판단할 수 있습니다.

## Line 3 — Git / PR (5 segments)

### Directory

- **포맷**: `📁 <directory name>`
- **데이터 소스**: stdin `workspace.project_dir` (basename) 또는 `cwd`
- **예시**: `📁 moai-adk-go`, `📁 my-project`
- **숨김 조건**: `data.Directory` 빈 문자열
- **세그먼트 키**: `directory`

### Repo + Branch (통합 세그먼트)

- **포맷**: `🔀 <owner>/<name> | 🅱️ <branch>[ ↑N][ ↓N][ +N]`
- **데이터 소스**:
  - `🔀 owner/name`: stdin `workspace.repo.{host, owner, name}` (Claude Code v2.1.145+)
  - `🅱️ branch`: 로컬 git `branch --show-current`
  - `↑N`: ahead count (origin/<branch> 대비)
  - `↓N`: behind count
  - `+N`: dirty count = Modified + Staged + Untracked
- **예시**:
  - `🔀 modu-ai/moai-adk | 🅱️ main ↑3 +2` (repo + branch + ahead + dirty)
  - `🔀 modu-ai/moai-adk | 🅱️ main` (clean branch, no ahead)
- **숨김 조건** (셋 중 하나라도 해당하면 세그먼트 전체 숨김):
  - branch 빈 문자열 또는 git 미가용
  - `workspace.repo` nil (git 미초기화 또는 remote 미설정) — repo 없이 branch만 표시하는 fallback은 없습니다
  - `repo.owner` 또는 `repo.name` 빈 문자열
- **Worktree 모드**: `worktree` segment 활성 + `workspace.git_worktree` 존재 시 branch에 `[WT] ` prefix
- **세그먼트 키**: `git_branch` (combined). `🔀 owner/name` 부분(`repo`)은 이 세그먼트 안에서 렌더되며 16-key 설정 스키마 밖의 17번째 세그먼트입니다 (개별 토글 불가).

### Git Status

- **포맷**: `💾 +<staged> M<modified> ?<untracked>`
- **데이터 소스**: 로컬 git `git status --porcelain` 파싱
- **예시**: `💾 +0 M1 ?1` (staged 0, modified 1, untracked 1)
- **숨김 조건**: git 미가용
- **참고**: 이전 mailbox 4종 emoji (`📬`/`📫`/`📪`/`📭`) 폐기, 통일된 `💾` 사용
- **세그먼트 키**: `git_status`

### Task (활성 SPEC workflow)

- **포맷**: `📋 [<command> <SPEC-ID>-<stage>]`
- **데이터 소스**: `~/.moai/state/last-session-state.json` `active_task` 필드 (해당 파일 작성 시점에만 노출)
- **예시**: `📋 [run SPEC-AUTH-001-run]`
- **숨김 조건**: 활성 task 부재 (`active_task` nil 또는 command 빈 문자열) → segment 숨김
- **세그먼트 키**: `task` (v2.20.0-rc1부터 default-on — 미설정 키는 활성으로 해석)

### PR (활성 GitHub Pull Request)

- **포맷**: `💌 PR #<number> (⌥<review_state>)` (state 있을 때) / `💌 PR #<number>` (state 빈 문자열)
- **데이터 소스**: stdin `pr.{number, url, review_state}` (Claude Code v2.1.145+)
- **Review state 값**: `approved` / `pending` / `changes_requested` / `draft` / 기타 (raw passthrough)
- **색상 코딩** (review_state portion):
  - `approved`: 녹색 (Success)
  - `pending`: 노란색 (Warning)
  - `changes_requested`: 빨간색 (Error)
  - `draft`: 회색 (Muted)
  - 기타: 색상 없음 (raw passthrough)
- **예시**:
  - `💌 PR #1234 (⌥approved)` (녹색)
  - `💌 PR #1023 (⌥pending)` (노란색)
  - `💌 PR #7 (⌥changes_requested)` (빨간색)
  - `💌 PR #99 (⌥draft)` (회색)
  - `💌 PR #100` (state 없음)
- **숨김 조건**:
  - `pr` 필드 부재 (PR 없음 또는 v2.1.145 이하)
  - `pr.number == 0`
  - `SegmentPR` config 명시적 false
- **세그먼트 키**: `pr` (default on per v2.20.0-rc1)

## 설정

### 기본 구조

`.moai/config/sections/statusline.yaml`에서 segment 활성화를 관리합니다.

```yaml
statusline:
  theme: catppuccin-mocha    # 색상 테마
  segments:
    # Line 1
    model: true
    effort_thinking: true
    cache_hit: true        # 캐시 히트율 ♻️
    claude_version: true
    moai_version: true
    session_time: true
    output_style: true

    # Line 2
    context: true
    usage_5h: true
    usage_7d: true

    # Line 3
    directory: true
    git_branch: true       # combined repo+branch
    git_status: true
    task: true             # default-on per v2.20.0-rc1
    pr: true               # default on per v2.20.0-rc1
    worktree: false
```

### 새로고침 주기

Statusline의 새로고침 주기는 `settings.json`의 `statusLine.refreshInterval`로 설정합니다 (단위: **초**, 기본값 `10`). `.moai/config/sections/statusline.yaml`이 아닌 Claude Code 런타임 설정에 해당합니다. 값이 너무 낮으면 CPU 사용량이 늘어나고, 너무 높으면 컨텍스트 사용률 변화가 늦게 반영됩니다.

```json
{
  "statusLine": {
    "type": "command",
    "command": "$CLAUDE_PROJECT_DIR/.moai/status_line.sh",
    "refreshInterval": 10
  }
}
```

### Segment 활성 매트릭스

| 세그먼트 | 라인 | 기본 활성 | stdin field |
|---------|------|----------|-------------|
| `model` | L1 | ✓ | `model.display_name` |
| `effort_thinking` | L1 | ✓ | `effort.level` + `thinking.enabled` |
| `cache_hit` | L1 | ✓ | `current_usage.cache_read_tokens` + `cache_creation_tokens` |
| `claude_version` | L1 | ✓ | `version` |
| `moai_version` | L1 | ✓ | (로컬 config) |
| `session_time` | L1 | ✓ | `cost.total_duration_ms` |
| `output_style` | L1 | ✓ | `output_style.name` |
| `context` | L2 | ✓ | `context_window.*` |
| `usage_5h` | L2 | ✓ | `rate_limits.five_hour.*` |
| `usage_7d` | L2 | ✓ | `rate_limits.seven_day.*` |
| `directory` | L3 | ✓ | `workspace.project_dir` |
| `git_branch` (combined) | L3 | ✓ | `workspace.repo.*` + local git |
| `git_status` | L3 | ✓ | local git |
| `task` | L3 | ✓ (v2.20.0-rc1+) | 세션 상태의 `active_task` |
| `pr` | L3 | ✓ (v2.20.0-rc1+) | `pr.*` (Claude Code v2.1.145+) |
| `worktree` | L3 | ✗ opt-in | `workspace.git_worktree` |

> 위 16개가 정식 설정 스키마 키입니다. `repo`(`🔀 owner/name`)는 `git_branch` 세그먼트 안에서 렌더되는 17번째 세그먼트로, 설정 스키마 밖이라 개별 토글이 없습니다.

## Handoff Guide — `(⚠️/clear)` 권고 기준

CW bar의 handoff suffix는 컨텍스트 사용량이 모델별 임계값을 넘으면 활성화됩니다. 이는 SSE stall 위험을 사전에 방지하고 paste-ready resume message 활용을 권장하는 시각적 마커이며, **2단계**로 동작합니다.

- **soft 단계** `(⚠️/clear)`: 밴드의 soft 임계값 도달 시
- **hard 단계** `(🛑/clear!)`: auto-compact-aware ceiling(`min(cap, auto-compact-threshold + margin)`) 도달 시 (`internal/statusline/renderer.go`). 런타임 auto-compact가 종종 이 ceiling을 선점하므로 hard 단계는 실제로는 드물게 발화되는 상위 신호입니다.

| 모델 클래스 | Context Window | 임계값 | 권고 시점 |
|------------|----------------|--------|----------|
| **1M context** (Opus 4.8) | 1,000,000 tokens | **≥50%** | ~500K 토큰 사용 |
| **256K context** (Fable) | 256,000 tokens | **≥90%** | ~230K 토큰 사용 |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | ~180K 토큰 사용 |
| 기타 / 알 수 없음 | — | 표시 안 함 | (안전 default) |

> 임계값은 `internal/statusline/renderer.go`의 handoff 단계 판정에서 강제됩니다. 이 임계값은 `.claude/rules/moai/workflow/context-window-management.md` HARD rule과 일치합니다.

### GLM 컨텍스트 게이지 보정 (Issue #653)

GLM-5.2는 실제 1M 컨텍스트 모델이지만, Claude Code는 provider와 무관하게 Claude 슬롯 기준으로 `context_window_size`를 보고하므로 GLM 세션에서 raw telemetry(`effectiveWindow`)가 ~180K로 잘못 표시될 수 있습니다. MoAI는 이를 `ResolveGLMContextWindow`(`internal/statusline/memory.go`)로 보정합니다. `MOAI_STATUSLINE_CONTEXT_SIZE` 환경변수(명시적 오버라이드) 또는 `llm.yaml`의 `glm.context_windows` 테이블(glm-5.2 → 1,000,000)에서 값을 해석합니다. GLM 세션에서는 raw `effectiveWindow`가 아니라 MoAI statusline의 CW%를 신뢰하세요.

활성화 시 사용자 흐름은 다음과 같습니다.

1. `(⚠️/clear)` marker 노출
2. 진행 중인 작업을 `progress.md` 등에 저장
3. orchestrator가 paste-ready resume message 생성 (session-handoff.md 6-block 포맷)
4. `/clear` 실행 후 resume message 붙여넣기
5. 새 세션으로 이어 작업

## stdin JSON 스키마 참조

Claude Code가 statusline 스크립트로 전달하는 stdin JSON 전체 필드 목록은 [공식 docs Available data](https://code.claude.com/docs/en/statusline#available-data)를 참조하세요. moai-adk-go는 다음 필드를 활용합니다.

```json
{
  "session_id": "abc...",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/path/to/cwd",
  "model": {"id": "claude-opus-4-8", "display_name": "Opus"},
  "workspace": {
    "current_dir": "...",
    "project_dir": "...",
    "git_worktree": "feature-xyz",
    "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"}
  },
  "version": "2.1.212",
  "output_style": {"name": "MoAI"},
  "cost": {
    "total_cost_usd": 1.234,
    "total_duration_ms": 17520000,
    "total_lines_added": 156,
    "total_lines_removed": 23
  },
  "context_window": {
    "used_percentage": 62,
    "context_window_size": 1000000,
    "total_input_tokens": 620000,
    "total_output_tokens": 0,
    "current_usage": {
      "input_tokens": 8500,
      "output_tokens": 1200,
      "cache_creation_input_tokens": 5000,
      "cache_read_input_tokens": 605300
    }
  },
  "exceeds_200k_tokens": true,
  "effort": {"level": "xhigh"},
  "thinking": {"enabled": true},
  "rate_limits": {
    "five_hour": {"used_percentage": 56, "resets_at": 1779286800},
    "seven_day": {"used_percentage": 13, "resets_at": 1779832400}
  },
  "pr": {
    "number": 1234,
    "url": "https://github.com/modu-ai/moai-adk/pull/1234",
    "review_state": "approved"
  }
}
```

## 버전 히스토리

- **v2.20.0-rc1 layout v3** (2026-05-22): 3-line layout 재설계 — repo+branch 통합 segment, directory L3 head, `🪫 CW:` emoji 앞으로, `(⚠️/clear)` handoff suffix, `💾` git status 통일, `💌 PR #N (⌥state)` 형식
- **v2.20.0-rc1 STATUSLINE-STDINFIELDS-001** (2026-05-21): `workspace.repo` + `exceeds_200k_tokens` + `pr` stdin 필드 매핑 추가, 1M context handoff threshold 75% → 50%
- **v2.20.0-rc1 STATUSLINE-V2145-001** (2026-05-20): PR segment 추가 (v2.1.145+ stdin), 4-locale docs 동기화
- **v2.1.139** (Claude Code): `effort.level` + `thinking.enabled` stdin JSON 추가
- **v2.1.145** (Claude Code): `workspace.repo` + `pr` stdin JSON 추가

## 트러블슈팅

### Statusline에 PR이 안 나옴

- Claude Code 버전 확인: `🔅 v2.1.145` 이상 필요 (그 이전 버전은 stdin에 `pr` 필드 미포함)
- 현재 branch에 OPEN PR이 있는지 확인: `gh pr view`
- `statusline.yaml`에 `pr: false`로 명시되었는지 확인

### `(⚠️/clear)` 표시 안 됨

- 1M context 모델: used_percentage 50% 미만 → 정상 (아직 임계값 미달)
- 200K context 모델: used_percentage 90% 미만 → 정상
- 임계값 초과인데 표시 안 됨: `shouldShowHandoffGuide` 함수의 `MemoryData.ContextWindowSize` 매핑 확인 (boundary defect 가능성)

### 색상이 표시 안 됨

- 터미널이 ANSI 256-color 지원하는지 확인
- `theme: catppuccin-mocha`가 환경 적합한지 확인
- `NO_COLOR=1` 환경변수 설정 여부 확인

### 검증 명령

```bash
# stdin fixture로 statusline 실 출력 확인
NOW=$(date +%s)
echo '{"session_id":"test","model":{"display_name":"Opus"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.212","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## `/cd` 캐시 보존 디렉터리 전환 (CC 2.1.169+)

Claude Code 2.1.169+는 세션의 작업 디렉터리를 **프롬프트 캐시를 보존하면서** 변경하는 `/cd <path>` 명령을 제공합니다. statusline의 `cwd` 필드가 새 디렉터리를 반영하도록 업데이트되지만, 진행 중인 추론 컨텍스트는 재구축되지 않습니다. 새 터미널 세션을 여는 대신 캐시를 보존하는 방법입니다. `/cd`는 누적된 컨텍스트를 유지하고, 새 터미널은 처음부터 cold-start합니다. 세션 중 컨텍스트 손실 없이 `cwd`를 옮기고 싶을 때(예: 세션 중 L2 worktree로 전환), `/cd`가 마찰이 적은 경로입니다. resume-pattern 통합은 [세션 핸드오프](/ko/workflow-commands/moai-sync)를 참조하세요.

## 관련 문서

- [Settings JSON](/ko/advanced/settings-json) — Claude Code `statusLine` 필드 설정
