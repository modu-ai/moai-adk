---
id: SPEC-20261120-leader-round-registration-report
title: "leader-round reports the leader registration lease, ended included"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# leader-round reports the leader registration lease, ended included

## Intent

협업 보드 최종 설계(3판, 2026-11-20) 3.1절이 관찰한 문제: 등록부의 leader 항목 임대가 확인 시점보다 약 38시간 전에 끝나 있었어요. 모드 지침은 리더가 매 감독 라운드마다 `session-register` 를 같은 `instance` 로 갱신하라고 이미 요구해요(mode-leader 배너 단계). 그러나 그 지침을 지켰는지 보여 주는 수단이 없고, 등록부의 공개 읽기(`session_live`·`mode_of`·`show`)는 임대가 살아 있는 항목만 다루며 끝난 항목은 `free` 로 접어 버려서, 언제 끝났는지와 mode 가 leader 인지가 사라져요.

이 조각은 `moai leader-round` 의 공유 답에 `registration` 부분을 더해요: mode 가 leader 인 등록 항목마다 이름·임대 끝 시각·끝났다면 그때부터 흐른 초를 보이고, 갱신 명령 모양을 함께 적어요. 리더는 라운드마다(기본 300초) 이 보고를 보고 곧바로 갱신할 수 있어요. 등록부 임대는 생존 신호가 아니에요(설계 3.1절: 생존 증명은 잠금에 기대요) — 이 보고는 갱신 누락을 드러내는 것이지 세션의 생존을 판단하지 않아요.

읽기 전용이에요: 등록부와 복제 폴더의 그 어떤 파일도 바꾸지 않고, GitHub 스냅숋 없이도 보고해요. 구현 순서는 리더 결정대로: 살아 있는 세션 증명 조각(등록 항목에 세션 id 필드를 더하는 일)이 착지한 뒤, 그 항목 형태 위에서 만료 임대 읽기와 이 보고를 완성해요. 그 전에 이 SPEC 과 RED 시험을 먼저 두어요.

## Non-goals

- 갱신 자체: 리더가 스스로 하는 행위예요(mode-leader 배넝 단계). 이 조각은 보고만 해요.
- 등록 항목 스키마 변경: schema 1 그대로(세션 id 필드는 살아 있는 세션 증명 조각이 다뤄요).
- session-check: 세션 시작 때 한 번만 떠서 누적된 만료를 못 봐요; 보고 자리는 라운드마다 도는 leader-round 로 정했어요.
- 회수 한도 3 과 4시간 임대 기본값, 원장 스키마 v1, 호스트 훅: 모두 그대로.
- 살아 있지 않은 리더 세션의 자동 조치: 보고만 하고 아무 것도 재배차하지 않아요.
- Out of scope: 실제 HOME 의 등록부, 네트워크 조회, 살아 있는 세션 증명 조각의 범위(잠금·세션 id).

## Requirements

- REQ-1: When `leader-round` answers, its shared object shall carry a `registration` part that lists every registry entry whose mode is leader, each with its name, its lease end time and, when that end is past, the seconds elapsed since it. · failure: none — REQ-4 covers an entry that cannot be read.
- REQ-2: While every leader lease the part lists still runs past the round's `now`, the part's status shall be `ok`. · failure: none — the status is the outcome.
- REQ-3: While the registry holds no leader-mode entry, or no registry folder exists, the part's status shall be `absent`. · failure: none — the status is the outcome.
- REQ-4 (REQ-1): IF an entry of the registry cannot be read, THEN the `registration` part shall be `error` with that message alone, leaving every other part of the round as it was.
- REQ-5: While any leader entry the part lists has a lease that ended, the part's status shall be `expired`, naming that entry with the seconds elapsed since its end. · failure: none — the status is the outcome.
- REQ-6: The part's note shall name the renewal step of the mode-leader banner — `session-register` with the entry's `instance` — so a leader reading the round can renew at once.
- REQ-7: Reading the part shall write nothing: the registry folder and the clone state stay exactly as they were.
- REQ-8: The part shall report without a GitHub snapshot — with neither `--repo` nor `--snapshot` given — because it reads only the local registry.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-5): with a leader entry whose lease ended 38 hours before `--now` in the fixture's registry, the round's `registration` is `expired` and names the entry with the elapsed seconds; with a live lease it is `ok`; with no registry folder it is `absent`.
  verify: `cargo test --locked --offline --test leader_round_cli`
- AC-2 (REQ-6): the part's note names `session-register` and `--instance`.
  verify: `cargo test --locked --offline --test leader_round_cli`
- AC-3 (REQ-4): an entry that cannot be read makes the part `error` with the message, and the other shared parts answer as they did.
  verify: `cargo test --locked --offline --test leader_round_cli`
- AC-4 (REQ-7): a snapshot of the registry folder and the clone state before and after the run is unchanged.
  verify: `cargo test --locked --offline --test leader_round_cli`
- AC-5 (REQ-8): without `--repo` and `--snapshot` the answer still carries the `registration` part alongside the `not_run` shared parts.
  verify: `cargo test --locked --offline --test leader_round_cli`
- AC-6: the neighbour round checks and the style checks pass.
  verify: `cargo test --locked --offline --test leader_pulse_cli`
  verify: `cargo test --locked --offline --test leader_cli`
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261120-leader-round-registration-report/spec.md`, `src/leader_round.rs`, `tests/leader_round_cli.rs`, and — after the liveness slice lands — one read helper in `src/lane_registry.rs` (taken then, never while another card holds the file)
- read-only: `src/ops/leader_round.rs`, `src/lane_registry.rs`
- forbidden: 실제 HOME 의 등록부, 네트워크, 호스트 훅, 등록 항목 스키마와 원장

## Plan

1. 이 SPEC 과 RED 시험을 먼저 두요(2026-11-20): 시험은 끝난 임대·살아 있는 임대·등록부 없음·읽을 수 없는 항목·무쓰기·스냅숏 없음의 장면을 고정하고, 지금 코드로는 깨져요.
2. 살아 있는 세션 증명 조각이 READY 를 쓰면 그 head 위에서: 등록부에 만료 임대 읽기 도우미를 하나 두고(그 조각의 항목 형태를 그대로 읽어요) `leader_round::round` 가 다섯째 부분으로 담아요.
3. GREEN 뒤 이웃 시험·여섯 가벼운 검사·변이(상태 판정 뒤집기, 초 계산, 무쓰기)·감사·READY.

## Risks

- 살아 있는 세션 증명 조각이 등록 항목의 형태를 바꿔요(세션 id 필드 추가): 읽기 도우미는 그 조각이 착지한 head 위에서 그 형태를 그대로 읽으므로 이 SPEC 의 REQ 는 변하지 않아요.
- 보고의 갱신 문구가 mode-leader 문서의 명령과 어긋날 수 있어요: 핀은 명령의 핵심 낱말(`session-register`, `--instance`)만 고정해요.
- `round()` 가 등록부를 읽으면 순수성이 옅어져요: 읽기는 건드리지 않는 로컬 한 폴더로 한정하고, 없는 폴더는 `absent` 로 답해 단위 시험이 결정적으로 유지돼요.
