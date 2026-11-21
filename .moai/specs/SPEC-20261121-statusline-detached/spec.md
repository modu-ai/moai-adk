---
id: SPEC-20261121-statusline-detached
title: 상태줄 detached 표시와 조회 실패 미확인, 그리고 사용자가 고른 2·3줄 정리 (C-203 묶음)
status: draft
revision: 1
created: 2026-11-21
---

## Intent

리더 결정(카드 노트 5373, 사용자 확인 2026-11-21T09:22:55Z): statusline의 `repo_groups`가
브랜치 칸이 빈 문자열인(detached HEAD) 때 저장소·GitHub 개수까지 통째로 숨기는 원인을 고쳐요.
지금 코드(`src/statusline/render.rs:376`)는 `branch.is_empty()`면 `None`을 돌려 2줄의
repo·forge·branch 세그먼트가 모두 사라져요. 규칙: 브랜치가 있으면 실제 이름, detached면 짧은
SHA, 저장소·remote·개수 조회 실패는 그 항목만 `미확인`(0으로 표시 금지), GitHub 줄은 항상 보임.

사용자가 고른 2·3줄 정리(카드 C-233 본문, 2026-11-20T21:5xZ mock-up 선택)와 그 뒤의 최신
사용자 지시(리더 노트 5378·5379, 그리고 그보다 새로 이 세션에 입력된 지시): 2줄은
`📡 <repo>, ⬇️ n 🔀 n | 🅱️ <branch> ↑n ↓n +n | 🌳 <worktree>`(worktree 라벨이 3줄 끝에서
2줄 끝으로), 3줄의 Week에는 리셋 카운트다운 `(resets in 6d 21h)`가 붙어요. C-233의
"막대 없이 숫자만"은 더 새로운 사용자 지시(숫자·막대는 그대로)에 덮여 막대와 콜론은
기존 그대로고, `(⚠️ /clear)`·`(🛑/clear!)` 안내만 없애요. 1줄은 그대로. `↑N`은 이미
착지된 SPEC-20261120-statusline-ahead의 동작이라 이 SPEC은 새로 요구하지 않아요.

관계: 이 카드(C-203)의 착지된 레이아웃 작업은 SPEC-20261120-statusline-layout(draft)이고,
이 SPEC은 그 위의 수리와 재정리예요. C-212·C-233의 내용이 이 SPEC 하나로 실려요(리더 묶음 지시).

## Requirements

- REQ-1: When git reports a repository and the branch name is empty (detached HEAD), the renderer shall collect the short HEAD SHA and show it in the branch slot — `🅱️ <short sha>` with the repository segment and its forge counts still shown. · failure: none — a missing SHA leaves the branch slot out and the repository segment stays (REQ-8).
- REQ-2: When git reports a repository but the repository name or its remote cannot be resolved, line 2 shall show `📡 미확인` in the repository slot instead of hiding the segment. · failure: none — a resolved name renders as the name; the unknown form is only for the failed lookup.
- REQ-3: When the forge counts cannot be read (`available` false), the counts shall render `⬇️ 미확인 🔀 미확인` and never `⬇️ 0 🔀 0`, while a user switch that suppresses the counts keeps them hidden. · failure: none — a successful read renders the numbers; the unknown form is only for the failed read.
- REQ-4: When the worktree label is present, the renderer shall close line 2 with it after the branch slot and any pull-request link, leaving line 3 without a worktree segment. · failure: none — an absent label closes nothing; the main checkout has none.
- REQ-5: The context and weekly windows shall keep their rendering as it is — the battery icon switching to `🪫` above 70% and `🔋` otherwise, the label with its trailing colon, the gradient bar and the percentage — with a window whose numbers are unavailable keeping the existing `n/a` placeholder.
- REQ-6: When the weekly window carries a reset time, the weekly segment shall append `(resets in <d>d <h>h)` at a day or more, `(resets in <h>h <m>m)` under a day and `(resets in <m>m)` under an hour, measured from the render's clock. · failure: none — without a reset time the segment ends at the percentage.
- REQ-7: The renderer shall not render the handoff hint `(⚠️ /clear)` or `(🛑/clear!)` on any line.
- REQ-8 (REQ-1): IF the branch name is empty and no short SHA is available, THEN line 2 shall still show the repository segment and its forge counts and omit only the branch slot.

실패 경로: REQ-1 → REQ-8.

## Acceptance

- AC-1 (REQ-1, REQ-8): verify: `cargo test --offline --locked detached_head_shows_the_short_sha_and_keeps_the_repository` — 가지가 빈 head_sha 붙은 상태의 2줄이 `📡 owner/repo, ⬇️ … 🔀 … | 🅱️ a1b2c3d …`를 말하고, SHA 없는 상태에선 repo 세그먼트만 남아요.
- AC-2 (REQ-2, REQ-3): verify: `cargo test --offline --locked lookup_failures_show_unknown_not_zero` — repo가 None이고 github가 unavailable인 2줄이 `📡 미확인 … ⬇️ 미확인 🔀 미확인`을 말하고 0을 갖지 않아요.
- AC-3 (REQ-4, REQ-5, REQ-6, REQ-7): verify: `cargo test --offline --locked the_user_example_gives_exactly_the_three_lines` — 사용자 예시(C-233 본문 값, worktree c135, resets in 6d 21h)가 세 줄로 정확히 나오고, 막대·콜론·/clear 힌트가 어느 줄에도 없어요.
- AC-4 (REQ-6): verify: `cargo test --offline --locked the_reset_countdown_uses_the_three_formats` — 6d21h·5h30m·40m 입력이 세 포맷으로 나와요.
- AC-5 (REQ-1…REQ-8): verify: `python3 -I tools/statusline_golden.py --bin <tree binary> --check` — 골든을 `--write`로 재생성하고 모든 변경 골든을 사람이 리뷰했어요(노트에 나열).
- AC-6 (REQ-1…REQ-8): verify: `moai spec check SPEC-20261121-statusline-detached` → errors 0, warnings 0.

## Scope

- `src/statusline/**`(render.rs, state.rs 수집, segments.rs), `tests/statusline_golden.rs`,
  `tests/fixtures/statusline-claude/**`(재생성), 이 SPEC.
- 이 카드의 이전 SPEC SPEC-20261120-statusline-layout은 그대로(착지된 작업의 기록).

## Plan

1. 시험 먼저: 유닛 4종을 목표 문장으로 RED(detached·미확인·C-233 예시·카운트다운 경계).
2. 수집: symbolic-ref 실패 시 rev-parse --short 한 번(GitStatus.head_sha).
3. 렌더: repo_groups의 None 원인 제거(REQ-1·2·3·8), git_line에 worktree 라벨 편입(REQ-4),
   bars_line 숫자화·카운트다운(REQ-5·6·7).
4. 골든 재생성 + 전수 리뷰, fmt/clippy/--max 2 슬롯 검사, spec check.
