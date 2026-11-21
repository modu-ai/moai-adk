---
id: SPEC-20261121-claude-hooks-redesign
title: Claude Code 훅 재설계 R5 — 공식 문서 기준 이벤트·matcher·출력 확인, 미사용 훅 제거, 사용자 훅 보존
status: draft
revision: 1
created: 2026-11-21
---

## Intent

카드 C-310(사용자 확정 설계 2026-11-21T06:3xZ): moai-adk 는 Claude Code 전용(GLM 경로 유지,
Codex 철수)이고, 훅은 공식 문서(code.claude.com/docs/en/hooks — 이 SPEC 작성 시점
2026-11-21T13:1xZ에 읽음, 로컬 Claude Code 2.1.330)을 기준으로 다시 확인해요.

지금 코드에서 읽은 사실(로컬 tip 33f74fccdd 워크트리):

- 설치 세트(src/hooks.rs HOOK_SET): SessionStart(matcher 없음=전체), PreToolUse
  matcher "Bash" → ClaimGuard, PostToolUse matcher "Bash" → Receipt. Stop 은 이미
  빠져 있어요(영수증 읽는 사람이 없고 turn-validation 규칙이 비어 있음).
- 엔진(src/hook_engine.rs)의 임무 5개는 공식 4이벤트에 매핑돼요: SessionCheck→SessionStart,
  ToolPreflight→PreToolUse, ToolObserve·EditDiagnostics→PostToolUse,
  TurnValidation→Stop(미설치). 출력은 최신 형태(hookSpecificOutput.permissionDecision
  deny·additionalContext·Stop decision block — 2026-11-17 문서 기준 기록).
- 호스트 계층(src/hooks.rs): claude 에 더해 codex(.codex/hooks.json)와 omp(.omp/extensions)
  파일 호스트가 있어요 — 사용자 결정(Codex 철수)으로 둘 다 제거 대상이고, 그 본체는
  C-289..292(Codex 철수 카드들)가 가져요. 이 SPEC은 훅 계층 조각만 확인·제거해요.
- 미리보기·백업·되돌리기는 이미 있어요: HostPlan(update_hooks.rs — create/update/
  unchanged/conflict), 확인된 apply 가 바꾼 바이트를 영수증으로 보존, hooks-revert 가
  복원. 사용자 훅은 "user": true 마커와 D1 충돌 규칙(사용자가 고친 항목은 그대로 둠)으로
  보존돼요.

공식 문서에서 확인한 계약(구현 시점 재독 요구): 이벤트 32개 중 moai 가 쓰는 4개의 matcher
값(PreToolUse/PostToolUse 는 도구 이름 — "Bash" 유효; SessionStart 는 시작 이유
startup/resume/clear/compact/fork — matcher 없음=전체 발화), 입력 공통 필드
(session_id·cwd·permission_mode·effort·hook_event_name 등), 출력(exit 0 의 단일 JSON
stdout, exit 2 는 PreToolUse 블록, permissionDecision 은 allow/deny/ask/defer — defer 는
문서에 새로 생긴 값), 5종 핸들러, 설정 5곳 병합, 실행 환경(병렬·타임아웃 600초·cwd·
${CLAUDE_PROJECT_DIR} 은 워크트리에서 원래 프로젝트 루트 고정).

## Requirements

- REQ-1: When the hook layer is planned or applied, it shall target only Claude Code — the codex and omp file hosts, their receipts and their extension or hook files shall not be planned, applied or updated, while the user-written content of those files stays untouched.
- REQ-2: The installed Claude hook set shall be exactly SessionStart (match all), PreToolUse matcher `Bash`, PostToolUse matcher `Bash` — with `MultiEdit` and every other tool name out of the preflight matcher unless the installed Claude Code (2.1.330) still reports that tool, and with no installed entry that the engine does not implement.
- REQ-3: When the engine answers PreToolUse, it shall emit the current `hookSpecificOutput.permissionDecision` shapes (`deny` with a reason; ask as `ask`; allow as exit 0 with empty stdout), and never the retired `approve` spelling.
- REQ-4: When the layer plans a host, user-written hook entries (the `user: true` marker and entries MoAI does not own) shall be preserved byte for byte and merged around the MoAI set, per the documented merge-across-settings rule, with any conflict reported and left as it is.
- REQ-5: The plan and apply shall keep the existing preview, conflict report, receipt backup and `hooks-revert` behavior — a confirmed apply keeps the replaced bytes, and a revert restores a host only while its files still hold what the apply wrote.
- REQ-6: Every hook event the product installs shall be verified against the installed Claude Code 2.1.330 by an event input/output test (stdin fixture → engine answer), and the set shall contain no auto-allow and no auto-done behavior; the WorktreeCreate/WorktreeRemove events shall not be installed by default.

실패 경로: REQ-2 → REQ-6(검증이 도구 이름을 반박하면 matcher 를 고치는 것이 이 SPEC 의 수리 경로).

## Acceptance

- AC-1 (REQ-1): verify: `cargo test --offline --locked --test hooks_apply_cli` (이름은 구현 시 확정) — codex·omp 호스트가 plan 에 나타나지 않고, 그 파일들의 사용자 내용이 없어도 apply 가 그 파일을 만지지 않아요.
- AC-2 (REQ-2): verify: `cargo test --offline --locked --test hooks_plan_cli` — 설치 세트가 3이벤트 그대로이고 preflight matcher 에 MultiEdit 이 없어요.
- AC-3 (REQ-3): verify: `cargo test --offline --locked --test hook_engine_cli` — deny 가 hookSpecificOutput.permissionDecision 로 나가요.
- AC-4 (REQ-4): verify: `cargo test --offline --locked --test hooks_apply_cli user_entries` — 사용자 항목 보존·병합과 충돌 보고.
- AC-5 (REQ-5): verify: 같은 파일의 receipt·revert 시험 — apply 영수증과 revert 조건.
- AC-6 (REQ-6): verify: `cargo test --offline --locked --test hook_events_docs` — 4이벤트 입력 픽스처 → 엔진 응답(2.1.330 계약), auto-allow·auto-done 부재, WorktreeCreate/Remove 미설치.
- AC-7 (전체): verify: `moai spec check SPEC-20261121-claude-hooks-redesign` → errors 0, warnings 0.

## Scope

- 읽기 전용 조각(이것): 이 SPEC + PLAN 노트. 구현 조각: `src/hooks.rs`(설치 세트·호스트
  계층), `src/hook_engine.rs`(C-246 예약 — 조율 필요), `src/update_hooks.rs`·`src/hook_legacy.rs`
  (보존·되돌리기), 위 시험 파일들. 소유 확인은 card scope 로 했고 hook_engine.rs 는 C-246
  예약 확인(20261121T130959Z take 충돌).
- 제외: Codex 철수 본체(C-289..292), 사용자 훅 내용 편집, 훅 규칙 자체의 강도 변경(이 SPEC 은
  문서 대조·제거·보존이지 규칙 추가가 아니에요).

## Plan

1. 문서 재독 기록: 위 Intent 의 읽은 시각·버전을 구현 커밋이 재확인(문서가 바뀌면 이 SPEC 을 고친 뒤 진행).
2. 조각 A(REQ-1·2): hooks.rs 에서 codex·omp 호스트 제거 + matcher 정리(AC-1·2).
3. 조각 B(REQ-3·6): 엔진 출력 확인 시험(AC-3·6) — hook_engine.rs 예약 조율.
4. 조각 C(REQ-4·5): 보존·병합·receipt·revert 시험 보강(AC-4·5).
5. 각 조각: check affected(들어간 것만) → MERGE-READY.
