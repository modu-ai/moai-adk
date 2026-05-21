---
title: Statusline 시스템 및 PR 세그먼트
weight: 78
draft: false
---

Claude Code와 moai-adk-go의 통합을 위한 **커스텀 statusline 시스템**입니다. v2.1.145부터 GitHub PR 정보를 statusline에 표시할 수 있습니다.

> MoAI 워크플로우는 PR 중심입니다. 모든 SPEC은 plan-PR → run-PR → sync-PR을 생성하므로, statusline에 현재 PR 상태를 표시하면 개발 효율이 높아집니다.

## 개요

### 커스텀 statusline이 필요한 이유

Claude Code의 기본 statusline은 일반적인 사용 패턴에 최적화되어 있습니다. 하지만 MoAI-ADK 사용자는 다음과 같은 특화된 정보가 필요합니다:

- **PR 중심 워크플로우**: 현재 작업 중인 PR 번호와 리뷰 상태 (approved/pending/changes_requested)
- **멀티팬 개발**: worktree 기반 병렬 개발 시 현재 SPEC 상태
- **비용 추적**: GLM 환경 사용 시 실시간 비용 모니터링
- **컨텍스트 관리**: 현재 세션의 token 사용률과 누적 비용

커스텀 statusline은 `.moai/status_line.sh` 렌더러를 통해 이러한 정보를 표시합니다.

### Statusline 아키텍처

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData 파싱)
    ↓
internal/statusline/builder.go (세그먼트 구성)
    ↓
internal/statusline/renderer.go (색상 코딩 및 렌더링)
    ↓
.moai/status_line.sh (템플릿 기반 최종 렌더링)
```

## 설정

### 기본 구조

`.moai/config/sections/statusline.yaml`에서 statusline을 설정합니다:

```yaml
statusline:
  mode: default              # default | compact | verbose
  theme: catppuccin-mocha    # 색상 테마 선택
  preset: full               # full | minimal | custom
  segments:
    model: true              # Claude 모델 표시
    context: true            # 컨텍스트 사용률 표시
    directory: true          # 작업 디렉터리 표시
    git_status: true         # Git 상태 표시
    git_branch: true         # Git 브랜치 표시
    worktree: false          # Worktree 정보 표시 (선택적)
    effort_thinking: false   # Effort/thinking 상태 (선택적)
    pr: false                # PR 정보 표시 (선택적, v2.1.145+)
```

### 세그먼트 옵션

| 세그먼트 | 기본값 | 용도 | 설명 |
|---------|--------|------|------|
| `model` | true | 현재 모델 | Claude 모델 버전 표시 |
| `context` | true | 컨텍스트 사용률 | 현재 세션의 token 사용률 |
| `directory` | true | 작업 경로 | 현재 작업 디렉터리 |
| `git_status` | true | Git 상태 | 수정된 파일 개수, stash 상태 |
| `git_branch` | true | 현재 브랜치 | 브랜치 이름과 원격과의 차이 |
| `worktree` | false | Worktree 정보 | 현재 worktree 표시 (병렬 개발) |
| `effort_thinking` | false | 사고 모드 | effort 및 thinking 상태 |
| `pr` | false | PR 정보 | GitHub PR 번호와 리뷰 상태 (NEW v2.1.145+) |

## 이용 가능한 세그먼트

### 항상 활성화되는 세그먼트 (4개)

**model** — Claude 모델
- Claude 3.5 Sonnet, Claude 3.7 Opus 등 현재 모델 표시
- 예: `Claude 3.5 Sonnet`

**context** — 컨텍스트 사용률
- 현재 세션의 token 사용률 표시
- 형식: `150K/200K` (사용 중 / 전체)
- 75% 이상 시 경고 색상 표시

**directory** — 작업 디렉터리
- 현재 작업 디렉터리의 상대 경로
- 프로젝트 루트에서의 위치 표시

**git_status** — Git 상태
- 수정된 파일 개수: `M5` (5개 파일 수정)
- Stash 상태: `S2` (2개 스테시)
- 예: `M5 S2`

**git_branch** — 현재 브랜치
- 브랜치 이름
- 원격과의 커밋 차이 표시
- 예: `feat/SPEC-001 +3 -1`

### 선택적 세그먼트 (7개)

**worktree** — Worktree 정보 (선택적)
- L2 worktree 사용 시 표시
- 현재 SPEC 이름
- 활성화: `segments.worktree: true`

**effort_thinking** — Effort/thinking 상태 (선택적)
- Claude 4.7의 thinking 모드 활성화 상태
- effort 레벨 (high/xhigh/max)
- 활성화: `segments.effort_thinking: true`

**output_style** — 출력 스타일 (선택적)
- 현재 출력 스타일 설정
- 활성화: `segments.output_style: true`

**claude_version** — Claude 버전 (선택적)
- Claude Code 버전
- 활성화: `segments.claude_version: true`

**moai_version** — moai 버전 (선택적)
- MoAI-ADK 버전
- 활성화: `segments.moai_version: true`

**session_time** — 세션 경과 시간 (선택적)
- 현재 세션 시작 후 경과 시간
- 활성화: `segments.session_time: true`

**usage_5h** — 5시간 누적 비용 (선택적)
- 지난 5시간의 비용 추적
- GLM 환경에서 유용
- 활성화: `segments.usage_5h: true`

**usage_7d** — 7일 누적 비용 (선택적)
- 지난 7일의 비용 추적
- 활성화: `segments.usage_7d: true`

**task** — Task 정보 (선택적)
- 현재 실행 중인 task 개수
- 활성화: `segments.task: true`

**repo** — 저장소 정보 (선택적, v2.1.145+)
- 현재 작업 중인 GitHub 저장소 소유자/이름 표시
- 예: `modu-ai/moai-adk`
- 활성화: `segments.repo: true`

**long_context** — 긴 컨텍스트 경고 (선택적, v2.1.139+)
- 200K 토큰 초과 시 경고 마크 표시
- 예: `⚠️ 200K+ exceeded`
- 활성화: `segments.long_context: true`

**handoff_guide** — 핸드오프 임계값 안내 (선택적, v2.1.146+)
- 현재 모델의 컨텍스트 창 크기와 권장 핸드오프 임계값 표시
- 1M 모델: 50% 임계값 (≈500K 토큰)
- 200K 모델: 90% 임계값 (≈180K 토큰)
- 예: `[1M: 50% | 200K: 90%]`
- 활성화: `segments.handoff_guide: true`

## NEW v2.1.145: PR 세그먼트

### 개요

Claude Code v2.1.145부터 statusline stdin JSON이 GitHub PR 정보를 포함합니다. MoAI-ADK는 이를 활용하여 statusline에 현재 PR의 리뷰 상태를 표시합니다.

**활성화**: `segments.pr: true`로 설정하면 활성화됩니다. 기본값은 `false`입니다.

### PR 세그먼트 표시 형식

PR 세그먼트는 다음 형식으로 표시됩니다:

```
#1023 ⌥approved
```

- `#1023`: PR 번호
- `⌥`: PR 상태 표시 기호
- `approved`: 리뷰 상태 (색상 코딩)

### 리뷰 상태별 색상

PR의 리뷰 상태에 따라 다른 색상으로 표시됩니다:

| 상태 | 색상 | 의미 |
|------|------|------|
| `approved` | 녹색 | PR이 승인됨 |
| `pending` | 노란색 | 리뷰 대기 중 |
| `changes_requested` | 빨간색 | 변경 요청됨 |
| `draft` | 회색 | 드래프트 상태 |
| (기타/빈 값) | 기본값 | 스타일 미적용 |

### 활성화 방법

1. `.moai/config/sections/statusline.yaml` 파일 편집

```yaml
statusline:
  segments:
    pr: true   # PR 세그먼트 활성화
```

2. Claude Code 세션 재시작

이제 statusline에 현재 PR의 번호와 리뷰 상태가 표시됩니다.

### JSON 입력 스키마 (v2.1.145+)

Claude Code v2.1.145+는 다음 형식의 JSON을 statusline stdin으로 전달합니다:

```json
{
  "pr": {
    "number": 1023,
    "url": "https://github.com/modu-ai/moai-adk/pull/1023",
    "review_state": "pending"
  },
  "workspace": {
    "repo": {
      "host": "github.com",
      "owner": "modu-ai",
      "name": "moai-adk"
    }
  }
}
```

- **pr.number**: PR 번호 (필수)
- **pr.url**: PR URL (선택적)
- **pr.review_state**: 리뷰 상태 (선택적, 기본값: empty)
- **workspace.repo.host**: Git 호스트 (github.com)
- **workspace.repo.owner**: 저장소 소유자
- **workspace.repo.name**: 저장소 이름

### 관련 사항

- **SPEC 참조**: [SPEC-V3R5-STATUSLINE-V2145-001](/ko/advanced/statusline#spec-참조)
- **최소 버전**: Claude Code v2.1.145 이상 필요
- **선택적 기능**: 기본값은 `false`이므로 명시적으로 활성화해야 함
- **역호환성**: 이전 버전의 Claude Code는 PR 정보를 전달하지 않음 (세그먼트가 표시되지 않음)

## 문제 해결: Statusline 사라짐 현상

### 증상

- Statusline이 간헐적으로 표시되지 않음
- Claude Code UI에서 statusline 영역이 비어 있음
- `.moai/cache/statusline_debug.log` 파일이 계속 커짐

### 원인 분석 (v2.1.145 M1 수정 이전)

statusline 렌더러는 Claude Code의 **300ms 디바운스 계약**을 준수해야 합니다. 이를 위반하면 진행 중인 실행이 취소됩니다.

이전 코드의 문제점:

```bash
# 문제: DEBUG_STATUSLINE의 기본값이 1 (항상 활성화)
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-1}

# 이로 인해 매 렌더링마다:
# 1. python3 -m json.tool 프로세스 포크 (50-250ms)
# 2. ~/.moai/cache/statusline_debug.log 쓰기 (~10ms)
# 총 60-260ms → 300ms 디바운스 경계선 넘음
# → Claude Code가 진행 중인 statusline 렌더링을 취소
# → 결과: statusline 표시되지 않음
```

### 해결책 (v2.1.145 M1에서 수정됨)

v3.5.0부터는 `DEBUG_STATUSLINE`의 기본값이 **0**입니다:

```bash
# 수정됨: 기본값 0 (비활성화)
DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}

# 디버깅이 필요할 때만 명시적으로 활성화:
export DEBUG_STATUSLINE=1
```

### Padding 조정

이전에는 `echo ""`를 사용하여 statusline 주변 여백을 조정했습니다. 이는 이제 권장되지 않습니다.

**대신** `.claude/settings.json`에서 설정하세요:

```json
{
  "statusLine": {
    "padding": 1
  }
}
```

- `padding: 0`: 여백 없음
- `padding: 1`: 위 아래 1줄 여백 (기본값)
- `padding: 2`: 위 아래 2줄 여백

### 체크리스트

statusline 표시 문제 해결 순서:

1. ✓ `DEBUG_STATUSLINE` 환경 변수 확인
   ```bash
   echo $DEBUG_STATUSLINE  # 기본값은 unset 또는 0이어야 함
   ```

2. ✓ `.moai/status_line.sh` 파일 확인
   ```bash
   grep "DEBUG_STATUSLINE=" ~/.moai/status_line.sh
   # 결과: DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0} 이어야 함
   ```

3. ✓ 명시적 디버깅 활성화 (필요한 경우만)
   ```bash
   export DEBUG_STATUSLINE=1
   # 이제 디버그 정보가 기록됨
   ```

4. ✓ Padding 설정
   ```json
   {
     "statusLine": {
       "padding": 1
     }
   }
   ```

5. ✓ Claude Code 세션 재시작

## 참고 자료

### 공식 문서

- [Claude Code Statusline 공식 문서](https://code.claude.com/docs/en/statusline) — Claude Code의 statusline 계약 및 JSON 스키마

### moai-adk-go 내부

- **패키지**: `internal/statusline/`
  - `types.go`: StdinData, PRInfo, RepoInfo 구조체 정의
  - `builder.go`: 세그먼트 생성 로직
  - `renderer.go`: 색상 코딩 및 최종 렌더링

- **템플릿**: `.moai/status_line.sh.tmpl`
  - 렌더러 호출 및 실행 로직

- **설정**: `.moai/config/sections/statusline.yaml`
  - 세그먼트 활성화/비활성화 설정

### 관련 SPEC

- **[SPEC-V3R5-STATUSLINE-V2145-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-V3R5-STATUSLINE-V2145-001/spec.md)**
  - M1: Statusline 사라짐 현상 수정
  - M2: v2.1.145 PR 세그먼트 추가
  - M3: 문서화 (현재 페이지)
