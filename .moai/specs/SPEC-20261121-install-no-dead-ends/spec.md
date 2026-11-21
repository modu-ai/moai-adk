---
id: SPEC-20261121-install-no-dead-ends
title: 설치 흐름의 막다른 길 없애기 (install-flow 보고서 §7 G15)
status: draft
revision: 1
created: 2026-11-21
---

## Intent

install-flow 보고서(`/Users/goos/.moai/reports/moai-adk/install-flow-2026-11-21.md` §7 G15, 로컬
v2026.11.0 a4d4c9a164 기준 2026-11-20T22:54-23:01Z에 읽음)의 네 가지 막다른 길이에요. 사용자 결정
2026-11-20T23:1xZ("모두 다 카드 만들어서 진행하고, 각 어떻게 할 것인지 다시 물어보면 내가 결정")의
HOW는 리더 노트(카드 노트 4684, 2026-11-20T23:28:27Z)로 기록되어 있어요: 막다른 길 없애기 —
비터미널 설치는 정확한 다음 단계 인쇄, PATH 줄 안내, Go 시절 skills 폴더는 막지 말고 안내형
백업·교체, 역할 세션 전체 권한은 처음 한 번 guarded 선택(권장 먼저).

지금 코드에서 읽은 사실(로컬 041536fe07, 워크트리에서 읽음):

- `install/install.sh:515-517` — 답을 읽을 곳이 없으면 "…Install them with: moai update…"를
  인쇄하고 돌아와요. 그런데 `moai update`도 터미널 없으면 다시 묻는 명령이라 막다른 길이 그대로여요.
  설치기는 그 지점에 계획 요약(요약 해시)을 이미 들고 있어요(492 줄).
- `install/install.sh:570-577` — PATH 알림이 export 줄과 함께 다음 단계 목록보다 먼저 인쇄돼요.
  이 동작은 이미 맞아요 — 시험으로 고정만 하면 돼요.
- `src/update.rs:1340-1390` — 내 소유 아닌 skills/moai 패키지에 `--confirm-backup-replace
  <token>` 지우개 플래그를 미리보기가 제안하는 기계가 이미 있어요(#2483). 그런데 이 경로를
  잡는 시험이 하나도 없어요(`src/`, `tests/`에서 confirm-backup-replace를 읽는 파일이 시험 없음).
  이 조각의 시험이 관찰한 실제 동작(2026-11-21): 미리보기는 claude 호스트를 blocked로 두고,
  플래그 없는 apply는 그 호스트를 `status: "blocked"`, `files: []`로 두고 사유 안내가
  "confirm with --confirm-backup-replace backup-replace:claude:<16hex>"까지 이름으로 말하며
  다른 호스트는 그대로 적용돼요(호스트별 독립). 안내 문구(`GO_ERA_GUIDANCE`, 76-78 줄)의
  "wait for a confirmed backup-then-replace procedure" 문장은 이미 뒤에 실제 경로가 이어져
  말해지므로 REQ-6은 그 오래된 문장을 지우는 정리만 남아요.
- `src/session.rs:97-103` — role 세션(`-l`/`-a`)은 전체 권한 기본값(`FULL_ACCESS_DEFAULT`),
  사용자 파일 키 `launcher.role_access`(full|prompt)이 이미 있어요. 어디에서도 질문은 없고
  런처 경고 줄만 있어요.

관계: `src/update.rs`는 카드 C-241이, `src/session.rs`는 카드 C-242가, `install/install.sh`는
카드 C-246이 예약 중이에요(2026-11-21T02:3xZ 기준). 이 SPEC의 구현 조각은 그 예약이 풀린 뒤
해당 파일 소유 카드와의 순서 조정으로 이어가요. C-242는 model·effort 질문을 다루므로 REQ-4의
질문 모양은 그 카드와 조율해요. 시험 파일과 SPEC은 이 조각이 쓰고, 구현 조각은 파일이 풀리는
대로 이 SPEC의 REQ를 따라요.

## Requirements

"답을 읽을 곳"은 설치기의 것과 같은 규칙이에요: stdin이 터미널이면 stdin, 아니면 열 수 있는
/dev/tty, 둘 다 아니면 없음(SPEC-20261117-installer-tty REQ 정의).

- REQ-1: When the installer reaches the skill-install question with no answer source and no
  `--yes`, it shall print the exact non-interactive command that completes the install — the
  working-directory change, the binary path and `update --confirm <plan digest>` the installer
  itself would run — instead of naming the interactive `moai update` alone, and shall keep
  naming `--yes` as the alternative.
- REQ-2: When the linked command folder is not on PATH, the installer shall print the PATH
  notice with the exact export line before the next-steps list that names `moai` commands.
- REQ-3: When the update preview meets a skills/moai folder the install receipt does not own,
  it shall offer the guided backup-and-replace — the plan names its
  `--confirm-backup-replace <token>` apply flag — and an apply without that flag shall leave
  every file of that folder untouched.
- REQ-4: When a role session (`-l`/`-a`) would start with full access because neither
  `launcher.role_access` nor an explicit `--permissions` choice exists, the product shall
  surface exactly one guarded question with the recommended option first (keep full access),
  persist the answer to `launcher.role_access`, and never ask again while the key exists; the
  non-interactive dry-run answer shall report the pending pick as `role_access_pick: "pending"`.
- REQ-5 (REQ-4): IF no answer source exists, THEN the product shall ask nothing and keep
  today's behavior (full access by default with the launcher's warning line).
- REQ-6: The Go-era guidance text shall name the existing `--confirm-backup-replace` path
  instead of telling the person to wait for it.

실패 경로: REQ-1 → 계획 요약이 없는 기존 실패 경로(update_failed). REQ-4 → REQ-5.

## Acceptance

- AC-1 (REQ-1, REQ-2): verify: `python3 -m unittest -k install_no_dead_ends tests/install_script_test.py`
  — 터미널 없는 설치(--yes 없음)의 출력이 digest가 들어간 update --confirm 명령을 이름으로 말하고,
  PATH 알림이 export 줄과 함께 다음 단계보다 먼저 나와요. (install.sh 조각이 착지할 때 met)
- AC-2 (REQ-3): verify: `cargo test --offline --locked --test install_no_dead_ends` — 미리보기가
  claude 호스트의 blocked 이유와 `--confirm-backup-replace backup-replace:claude:<16hex>` 플래그를
  제안하고, 플래그 없는 confirm이 외부 폴더의 바이트를 그대로 둬요. (이 조각에서 met)
- AC-3 (REQ-4, REQ-5): verify: `cargo test --offline --locked --test install_no_dead_ends` —
  신규 HOME의 dry-run 답이 `role_access_pick: "pending"`을 갖고, `launcher.role_access`가 있으면
  갖지 않아요. 대화형 질문 자체는 답을 읽을 곳이 필요해 pty 시험(install_script_test.py의 방식)으로
  session.rs 조각이 착지할 때 met.
- AC-4 (REQ-6): verify: `cargo test --offline --locked --test install_no_dead_ends` — 안내 문구
  시험이 `--confirm-backup-replace`를 이름으로 말해요. (update.rs 조각이 착지할 때 met)
- AC-5 (REQ-1, REQ-2): verify: `sh -n install/install.sh && bash -n install/install.sh` → exit 0.
  (install.sh 조각이 착지할 때 met)

## Scope

- 이 조각: `.moai/specs/SPEC-20261121-install-no-dead-ends/**`, `tests/install_no_dead_ends.rs`.
- 이어지는 조각: `install/install.sh`(REQ-1·2, C-246 조정), `src/update.rs`(REQ-3 문구·REQ-6,
  C-241 조정), `src/session.rs`(REQ-4·5, C-242 조정), `tests/install_script_test.py`(AC-1·3의
  pty 부분).

## Plan

1. 시험 먼저(이 조각): AC-2의 두 시험이 오늘 통과하는지 확인(기계가 이미 있으므로 통과가 정상,
   실패는 진짜 발견) — AC-3의 `role_access_pick` 시험은 지금 실패(RED)하는 걸 기록하고 이유와
   함께 올려요.
2. install.sh 조각: REQ-1의 정확한 명령 인쇄, REQ-2 핀 시험(AC-1·5).
3. update.rs 조각: REQ-6 문구(REQ-3 핀은 이 조각이 이미 놓음).
4. session.rs 조각: REQ-4·5 질문과 dry-run 시임(C-242 조율).
