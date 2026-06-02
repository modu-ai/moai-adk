---
title: Statusline 시스템 — 3-line 레이아웃 완전 가이드
weight: 78
draft: false
---

Claude Code와 moai-adk-go 통합을 위한 **커스텀 statusline 시스템**입니다. Claude Code v2.1.139부터 effort/thinking, v2.1.145부터 workspace.repo + pr 필드가 stdin JSON에 추가되어 풍부한 컨텍스트를 표시할 수 있습니다.

> MoAI 워크플로우는 PR 중심입니다. 모든 SPEC은 plan-PR → run-PR → sync-PR 사이클을 생성하므로, statusline에 현재 PR 번호 + 리뷰 상태 + 컨텍스트 사용률 + handoff 권고를 즉시 노출하면 개발 효율이 크게 높아집니다.

## 개요

### 최종 레이아웃 (3-line v3)

```
🤖 Opus 4.7 │ 🧠 xhigh·t │ 🔅 v2.1.146 │ 🗿 v2.20.0-rc1 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk (🅱️ main ↑5 +2) │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**: 모델 · effort/thinking · Claude Code 버전 · MoAI 버전 · 세션 시간 · output style
- **Line 2 (Usage bars)**: CW (context window) · 5H (rolling) · 7D (rolling) — 각 bar는 이모지 + label + bar + % + reset 정보
- **Line 3 (Git/PR)**: 디렉터리 · 리포지토리+브랜치 통합 · git status · 활성 SPEC task · PR 정보

### 데이터 흐름

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData 파싱)
    ↓
internal/statusline/builder.go (CollectMemory, CollectMetrics, etc.)
    ↓
internal/statusline/renderer.go (3-line v3 layout)
    ↓
.moai/status_line.sh → 터미널 표시
```

## Line 1 — Info (7 segments)

### 🤖 Model

- **포맷**: `🤖 <model display name>`
- **데이터 소스**: stdin `model.display_name` (또는 string shorthand)
- **예시**: `🤖 Opus 4.7`, `🤖 Sonnet 4.6`, `🤖 Haiku 4.5`
- **숨김 조건**: `model` field 부재 또는 `data.Metrics.Model == ""`
- **세그먼트 키**: `model`

### 🧠 Effort / Thinking

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

### 🔅 Claude Code 버전

- **포맷**: `🔅 v<version>` (default) 또는 `🔅 cc v<version>` (full mode)
- **데이터 소스**: stdin `version` 문자열
- **예시**: `🔅 v2.1.146`
- **숨김 조건**: `version` 빈 문자열
- **세그먼트 키**: `claude_version`

### 🗿 MoAI 버전

- **포맷**: `🗿 v<current>` 또는 업데이트 가능 시 `🗿 v<current> -> 🗿 v<latest>`
- **데이터 소스**: `.moai/config/sections/system.yaml` `moai.version` + 백그라운드 update checker 결과
- **예시**:
  - `🗿 v2.20.0-rc1` (최신)
  - `🗿 v2.18.0 -> 🗿 v2.20.0-rc1` (업데이트 권고)
- **세그먼트 키**: `moai_version`

### ⏳ 세션 시간

- **포맷**: `⏳ <X>h <Y>m` (≥1h) / `⏳ <X>m` (<1h) / `⏳ <X>d <Y>h` (≥24h)
- **데이터 소스**: stdin `cost.total_duration_ms`
- **예시**: `⏳ 4h 52m`, `⏳ 35m`, `⏳ 1d 3h`
- **세그먼트 키**: `session_time`

### 💬 Output Style

- **포맷**: `💬 <style name>`
- **데이터 소스**: stdin `output_style.name`
- **예시**: `💬 MoAI`, `💬 R2-D2`, `💬 default`
- **숨김 조건**: `output_style.name` 빈 문자열
- **세그먼트 키**: `output_style`

## Line 2 — Usage Bars (3 segments)

### 🪫/🔋 CW (Context Window)

- **포맷**: `<icon> CW: <bar> <pct>% [(⚠️/clear)]`
- **데이터 소스**:
  - bar: `context_window.context_window_size` × auto-compact threshold (default 85%) → scaled budget
  - 퍼센티지: `context_window.used_percentage` (사전 계산) 또는 `current_usage` tokens 합산
  - (⚠️/clear) 활성 조건: `shouldShowHandoffGuide(data) == true`
- **이모지**:
  - 🔋 (정상, <50% scaled)
  - 🪫 (경고, 50-79% scaled)
  - 🪫 (위험, ≥80% scaled, 색상 추가)
- **(⚠️/clear) handoff suffix**:
  - 1M context 모델 (Opus 4.7): used_percentage ≥50% (raw context_window_size 기준)
  - 200K context 모델 (Sonnet/Haiku): used_percentage ≥90%
  - 의미: 다음 turn 시작 전에 `/clear` 권고 + paste-ready resume message 활용
- **예시**: `🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **세그먼트 키**: `context`

### 🔋 5H (5시간 rolling rate limit)

- **포맷**: `🔋 5H: <bar> <pct>% [(<reset>)]`
- **데이터 소스**: stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset 포맷**:
  - <60분: `(Nm)` (예: `(47m)`)
  - <24시간: `(Nh Nm)` (예: `(2h 15m)`)
  - ≥24시간: `(Mon DD)` (예: `(May 28)`)
- **예시**: `🔋 5H: █████░░░░░ 56% (47m)`
- **데이터 부재**: `rate_limits.five_hour == null` → bar 0%, reset `(rolling)`
- **세그먼트 키**: `usage_5h`

### 🔋 7D (7일 rolling rate limit)

- **포맷**: `🔋 7D: <bar> <pct>% [(<reset>)]`
- **데이터 소스**: stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset 포맷**: `(Mon DD)` (절대 날짜)
- **예시**: `🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **세그먼트 키**: `usage_7d`

## Line 3 — Git / PR (5 segments)

### 📁 Directory

- **포맷**: `📁 <directory name>`
- **데이터 소스**: stdin `workspace.project_dir` (basename) 또는 `cwd`
- **예시**: `📁 moai-adk-go`, `📁 my-project`
- **숨김 조건**: `data.Directory` 빈 문자열
- **세그먼트 키**: `directory`

### 🔀 Repo + Branch (통합 세그먼트)

- **포맷**: `🔀 <owner>/<name> (🅱️ <branch>[ ↑N][ ↓N][ +N])`
- **데이터 소스**:
  - `🔀 owner/name`: stdin `workspace.repo.{host, owner, name}` (Claude Code v2.1.145+)
  - `🅱️ branch`: 로컬 git `branch --show-current`
  - `↑N`: ahead count (origin/<branch> 대비)
  - `↓N`: behind count
  - `+N`: dirty count = Modified + Staged + Untracked
- **예시**:
  - `🔀 modu-ai/moai-adk (🅱️ main ↑3 +2)` (repo + branch + ahead + dirty)
  - `🔀 modu-ai/moai-adk (🅱️ main)` (clean branch, no ahead)
  - `🔀 (🅱️ feat/auth ↑2 ↓1 +6)` (repo 정보 부재 fallback)
- **숨김 조건**:
  - branch 빈 문자열 → 전체 segment 숨김
  - repo nil 시 fallback (괄호 안 branch만 표시)
- **Worktree 모드**: `worktree` segment 활성 시 branch에 `[WT] ` prefix
- **세그먼트 키**: `git_branch` (combined)

### 💾 Git Status

- **포맷**: `💾 +<staged> M<modified> ?<untracked>`
- **데이터 소스**: 로컬 git `git status --porcelain` 파싱
- **예시**: `💾 +0 M1 ?1` (staged 0, modified 1, untracked 1)
- **숨김 조건**: git 미가용
- **참고**: 이전 mailbox 4종 emoji (📬/📫/📪/📭) 폐기, 통일된 💾 사용
- **세그먼트 키**: `git_status`

### 📋 Task (활성 SPEC workflow)

- **포맷**: `📋 [<command> <SPEC-ID>-<stage>]`
- **데이터 소스**: `~/.moai/state/last-session-state.json` `active_task` 필드 (해당 파일 작성 시점에만 노출)
- **예시**: `📋 [/moai run SPEC-V3R5-STATUSLINE-001-implement]`
- **숨김 조건**: 파일 부재 또는 `active_task` nil → segment 숨김
- **세그먼트 키**: `task` (opt-in default off)

### 💌 PR (활성 GitHub Pull Request)

- **포맷**: `💌 PR #<number> (⌥<review_state>)` (state 있을 때) / `💌 PR #<number>` (state 빈 문자열)
- **데이터 소스**: stdin `pr.{number, url, review_state}` (Claude Code v2.1.146+)
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

`.moai/config/sections/statusline.yaml`에서 segment 활성화를 관리합니다:

```yaml
statusline:
  mode: default              # default | full
  theme: catppuccin-mocha    # 색상 테마
  preset: custom             # full | minimal | custom
  segments:
    # Line 1
    model: true
    effort_thinking: true
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
    task: true             # opt-in default off in older versions
    pr: true               # default on per v2.20.0-rc1
    worktree: false
```

### Segment 활성 매트릭스

| 세그먼트 | 라인 | 기본 활성 | stdin field |
|---------|------|----------|-------------|
| `model` | L1 | ✅ | `model.display_name` |
| `effort_thinking` | L1 | ✅ | `effort.level` + `thinking.enabled` |
| `claude_version` | L1 | ✅ | `version` |
| `moai_version` | L1 | ✅ | (로컬 config) |
| `session_time` | L1 | ✅ | `cost.total_duration_ms` |
| `output_style` | L1 | ✅ | `output_style.name` |
| `context` | L2 | ✅ | `context_window.*` |
| `usage_5h` | L2 | ✅ | `rate_limits.five_hour.*` |
| `usage_7d` | L2 | ✅ | `rate_limits.seven_day.*` |
| `directory` | L3 | ✅ | `workspace.project_dir` |
| `git_branch` (combined) | L3 | ✅ | `workspace.repo.*` + local git |
| `git_status` | L3 | ✅ | local git |
| `task` | L3 | ⚠️ opt-in | `~/.moai/state/last-session-state.json` |
| `pr` | L3 | ✅ (v2.20.0-rc1+) | `pr.*` (Claude Code v2.1.146+) |
| `worktree` | L3 | ❌ opt-in | `workspace.git_worktree` |

## Handoff Guide — (⚠️/clear) 권고 기준

CW bar의 `(⚠️/clear)` suffix는 컨텍스트 사용량이 모델별 임계값을 넘으면 활성화됩니다. 이는 SSE stall 위험을 사전에 방지하고 paste-ready resume message 활용을 권장하는 시각적 마커입니다.

| 모델 클래스 | Context Window | 임계값 | 권고 시점 |
|------------|----------------|--------|----------|
| **1M context** (Opus 4.7) | 1,000,000 tokens | **≥50%** | ~500K 토큰 사용 |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | ~180K 토큰 사용 |
| 기타 / 알 수 없음 | — | 표시 안 함 | (안전 default) |

> 임계값은 `internal/statusline/renderer.go shouldShowHandoffGuide()` 함수에서 강제됩니다. 이 임계값은 `.claude/rules/moai/workflow/context-window-management.md` HARD rule과 일치합니다.

활성화 시 사용자 흐름:
1. `(⚠️/clear)` marker 노출
2. 진행 중인 작업을 `progress.md` 등에 저장
3. orchestrator가 paste-ready resume message 생성 (session-handoff.md 6-block 포맷)
4. `/clear` 실행 후 resume message 붙여넣기
5. 새 세션으로 이어 작업

## stdin JSON 스키마 참조

Claude Code가 statusline 스크립트로 전달하는 stdin JSON 전체 필드 목록은 [공식 docs Available data](https://code.claude.com/docs/en/statusline#available-data)를 참조하세요. moai-adk-go는 다음 필드를 활용합니다:

```json
{
  "session_id": "abc...",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/path/to/cwd",
  "model": {"id": "claude-opus-4-7", "display_name": "Opus 4.7"},
  "workspace": {
    "current_dir": "...",
    "project_dir": "...",
    "git_worktree": "feature-xyz",
    "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"}
  },
  "version": "2.1.146",
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
- **v2.1.146** (Claude Code): `workspace.repo` + `pr` stdin JSON 추가

## 트러블슈팅

### Statusline에 PR이 안 나옴

- Claude Code 버전 확인: `🔅 v2.1.146` 이상 필요 (v2.1.145는 stdin에 `pr` 필드 미포함)
- 현재 branch에 OPEN PR이 있는지 확인: `gh pr view`
- `statusline.yaml`에 `pr: false`로 명시되었는지 확인

### (⚠️/clear) 표시 안 됨

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
echo '{"session_id":"test","model":{"display_name":"Opus 4.7"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.146","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## 관련 문서

- [Settings JSON](/advanced/settings-json) — Claude Code `statusLine` 필드 설정
