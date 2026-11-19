---
id: SPEC-20261120-launcher-liveness
title: "Launcher liveness proof: the registry names the session, the launcher locks a live file, and a read-only lookup answers alive, launcher-exited or unknown"
status: draft
revision: 1
created: 2026-11-20
---

# Launcher liveness proof: the registry names the session, the launcher locks a live file, and a read-only lookup answers alive, launcher-exited or unknown

## Intent

협업 보드 최종 설계(3판) 5.2절 마-1 · BOARD 03 조각 (가)(카드 C-140, 사용자 승인). 회수(reclaim)가 담당의 생존을 모르고 임대 끝났다는 이유만으로 일하는 레인의 카드를 가져가는 것을 막는 첫걸이에요. 세 가지가 합쳐져요: `moai session-register` 가 항목에 호스트 세션 id 를 적어요(결정 노트 2836 의 옵션 가, 스키마 1 유지); 런처가 역할 있는 모든 세션 — 번호(`agent-<n>`)나 이름(`-l`, `--name`)이나 — 에서 겹침 폴더의 `<이름>.live` 배타 잠금을 잡아요(번호 세션은 이미 `agent-<n>.live` 를 잡아요); 읽기 전용 조회가 담당 세션 id → 등록 이름 → 그 이름의 `.live` 잠금 순서로 세 답을 내요 — 살아 있음(잠금 잡힘), 런처 종료(잠금 풀림), 미확인(등록부에 없거나 이름이 다른 세션 id 로 다시 등록했거나 잠금을 조회할 수 없음). "미확인"은 자동 회수를 하지 않아요: 사람이나 리더가 판단해요(설계 5.2절). 런처만 죽고 호스트가 남는 경우는 이 조각에서 "런처 종료"로 두고, 잠금을 호스트가 물려받을지는 별개 결정이에요(설계가 확인하지 못한 것, 8절).

## Non-goals

- 생존 조회가 회수를 하거나 풀어주는 일 — 조회는 읽기만 해요(회수 순서는 워크플로 문서가 말해요).
- 본인 재개(조각 나), 자동 갱신(마-2), 제안·수락 인계(조각 라).
- 원장(SQLite) 스키마 v1 변경, 호스트 훅 설치, 잠금 상속(런처 종료 뒤 호스트가 잠금을 이어받기).
- C-135 의 pid·시작 시각 닻 — 이 설계는 잠금으로 생존을 증명해요.

## Requirements

- REQ-1: The command `moai session-register` shall record the host session id (as `session-current` resolves it) as one more field of its registry entry, under schema version 1, so an older binary that ignores unknown keys still reads the entry.
- REQ-2: A read-only liveness lookup shall answer, for a session id, exactly one of three states — `alive` when an entry names the id and that entry's name still names the same id and the name's `.live` lock is held, `launcher-exited` when the same holds with the lock free, and `unknown` when no entry names the id, or the name now names another session id, or the lock cannot be probed.
- REQ-3: The liveness lookup shall read only — creating, changing or removing no file — and a missing `.live` file answers `launcher-exited` without being created.
- REQ-4: The session module shall offer a named live lock — the exclusive `File` lock on `<name>.live` in the sessions folder, refused on a symlink or a non-regular file — for the launcher to hold for every session it starts with a role, numbered or named alike, releasing it when the host exits.
- REQ-5: The leader workflow shall state the reclaim order — liveness first, work preservation second, remote requests third, reason and reclaim last — and stop describing a reclaim command as missing.
- REQ-6: No part of this change shall take or release a card reservation because of the lookup's answer, while the workflow documents say that an `unknown` answer goes to a person or the leader.

## Acceptance

- AC-1 (REQ-1): 등록한 항목에 세션 id 필드가 있고 스키마 판은 1 이에요.
  verify: `cargo test --locked --offline --test launcher_liveness_cli session_field`
- AC-2 (REQ-2): 세 답이 각각 시험으로 있어요 — 살아 있음·런처 종료·미확인(등록 없음, 이름 재등록, 잠금 조회 불가).
  verify: `cargo test --locked --offline --test launcher_liveness_cli three_answers`
- AC-3 (REQ-3): 조회 앞뒤로 겹침 폴더의 파일 목록·내용·수정 시각이 같고, 없던 잠금 파일이 생기지 않아요.
  verify: `cargo test --locked --offline --test launcher_liveness_cli reads_nothing`
- AC-4 (REQ-4): 이름 잠금 도우미가 배타 잠금을 잡고, 같은 파일의 둘째 잠금은 막히고, 심볼릭 링크는 거부돼요(런처 배선은 남은 조각).
  verify: `cargo test --locked --offline --test launcher_liveness_cli named_lock`
- AC-5 (REQ-5, REQ-6): 리더 워크플로의 회수 서술이 순서를 말하고 "명령이 없다" 문장이 사라져요(읽기 확인).
  manual: 문서 읽기

## Scope

- write: `src/lane_registry.rs`, `src/session.rs`, `src/ops/session_register.rs`, `tests/launcher_liveness_cli.rs`,
  `skills/moai/workflows/mode-leader.md`, `.moai/specs/SPEC-20261120-launcher-liveness/spec.md`
- read-only: `src/main.rs`(배선은 남은 조각, 카드 예약이 풀리면), `src/handoff.rs`, `src/session_current.rs`, `tests/common/**`
- forbidden: 원장 스키마 v1 바꾸기, 조회의 모든 쓰기, 호스트 훅 설치, 회수 한도 3회·기본 임대 4시간 바꾸기

## Plan

1. 시험 먼저(RED): `tests/launcher_liveness_cli.rs` — 세션 id 필드, 세 답, 무쓰기, 이름 잠금 도우미.
2. `src/lane_registry.rs`: 등록이 세션 id 를 적어요(스키마 1 유지); 조회 함수가 세션 id 로 (이름, 지금 세션 id) 를 찾아요.
3. `src/session.rs`: `named_live_lock` 도우미(배타 잠금, 심볼릭 링크·비정규 파일 거부)와 조회가 쓰는 잠금 확인.
4. `src/ops/session_register.rs`: 조회를 내놓는 읽기 전용 플래그.
5. 문서: `skills/moai/workflows/mode-leader.md` 회수 순서·낡은 문장.
6. 남은 조각(따로): `src/main.rs` 런처 배선(카드 예약이 풀리면), `mode-agent.md` 임대 네 시간 한 줄(같은 이유).

## Risks

- 런처만 죽고 호스트가 남으면 잠금은 풀려 "런처 종료"로 나와요 — 회수에는 임대 종료와 워크트리 문이 함께 필요해요(설계 5.2).
- 저장소 밖의 `--name` 세션은 겹침 폴더가 없어 잠금을 못 잡아요 — 조회는 미확인이에요.
- 겹침 폴더의 `.live` 파일 이름이 등록부 이름과 다른 철자(대소문자)면 미확인이에요 — 이름 규칙(1~64자, 안전한 글자)이 같은 규칙을 써요.

## Result

- (채우는 중) 조각 (가) 진행 중; 런처 배선과 mode-agent.md 한 줄은 카드 예약이 풀리는 뒤 남은 조각으로 붙어요.
