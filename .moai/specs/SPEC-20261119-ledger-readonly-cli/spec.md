---
id: SPEC-20261119-ledger-readonly-cli
title: "ledger read-only CLI — card show·list·next leave the HOME unchanged"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-card-reclaim-guard
---

# 원장 조회 명령의 전후 동일성 증명 (카드 C-43 기준 6)

## Intent

카드 C-43 기준 6: "조회 전후 원장·보조 파일·기존 작업 기록·worktree 가 같다. 읽기 전용 연결의 쓰기는 실패".
원장 파일 단위 증명은 tests/ledger_open.rs 가 이미 갖고 있고
(`a_read_only_open_and_read_leaves_the_folder_and_the_file_unchanged`,
`a_write_on_a_read_only_connection_fails_and_changes_nothing`), 조회 명령이 열어 보는 파일이 그뿐이라는
보장은 없어요. 이 조각은 CLI 조회 명령(`card show`·`card list`·`card next`)이 집 전체를 바꾸지 않는다는
증명을 **새 시험 파일 하나**로 넣어요. 세상은 임시 HOME 안에 실제 git 저장소와 그 worktree 를 두고,
기준이 이름한 대상을 전부 담아요: 등록된 원장(잡힌 카드와 메모 포함), 옛 체계의 기록(journal.log·세션
항목·제안), 예약이 기록한 worktree. 비교는 임시 HOME 의 모든 항목(파일 내용·종류·크기·폴더 mtime 포함,
`TempHome::snapshot`)을 조회 전후로 맞추는 방식이라, 조회가 무엇이든 만들거나 고치거나 건드리면 실패해요.
코드 변경은 없습니다(시험만 추가).

## Non-goals

- 읽기 전용 연결의 쓰기 실패(원장 라이브러리 반쪽) — tests/ledger_open.rs
  `a_write_on_a_read_only_connection_fails_and_changes_nothing` 이 이미 증명.
- `src/**` 의 어떤 변경 — 이 조각은 시험 파일과 이 SPEC 만 씁니다.
- `card inbox`·내보내기 등 호스트(네트워크)를 부르는 명령.
- 끝난 리스가 있는 세상의 `card next`(끝난 리스 표시는 설계상 쓰기 — 살아 있는 리스만 있는 세상에서 증명).

## Requirements

- REQ-1: When a session that does not hold the card runs `card show` on an existing card, the command shall answer the card with its notes and leave every entry under the temporary HOME unchanged, folder modification times included · failure: none — a HOME that differs after the command fails the snapshot comparison naming the entry.
- REQ-2: When `card list` runs on a registered project, the command shall answer the card it holds and leave every entry under the temporary HOME unchanged · failure: none — a HOME that differs after the command fails the snapshot comparison naming the entry.
- REQ-3: When a session with nothing offered runs `card next`, the command shall answer a null next and leave every entry under the temporary HOME unchanged · failure: none — a HOME that differs after the command fails the snapshot comparison naming the entry.
- REQ-4: Where the proof compares a before/after snapshot of the HOME, the compared world shall contain a real git repository with a worktree recorded on the reservation, the older system's records (journal.log, a session entry, a proposal) beside the registered ledger's project, and one taken card with a note, so it names every object criterion 6 lists.

## Acceptance

- AC-1 (REQ-1..REQ-4): 시험 파일 전체 통과.
  verify: `cargo test --locked --offline --test ledger_readonly_cli`
- AC-2: 변이 탐침 — 조회 경로에 인위적인 쓰기를 넣으면 이 시험이 실패해요(KILLED 확인 후 복원).
  verify: 변이 실행 로그를 Result 에 기록
- AC-3: 이웃 검사와 문법 검사.
  verify: `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test ledger_open`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `tests/ledger_readonly_cli.rs`, `.moai/specs/SPEC-20261119-ledger-readonly-cli/spec.md`
- read-only: `tests/common/**`, `src/card/command.rs`, `src/card/dispatch.rs`, `src/ledger/open.rs`
- forbidden: 실제 HOME, `src/**` 편집(변이 탐침의 임시 적용은 실행 뒤 복원하며 커밋하지 않음), 네트워크, 스키마 변경, 실제 원격 저장소

## Plan

1. 시험 파일 작성 후 실행 — 기존 동작을 고정하는 증명 시험이라 즉시 통과가 예상되고, 실제 결과를 Result 에 기록.
2. 커밋 뒤 변이 탐침: 조회 경로(src/card/command.rs list)에 인위적인 쓰레기 쓰기 한 줄을 넣어 시험이 잡는지(KILLED) 확인하고 복원.
3. 이웃 검사(AC-3)와 가벼운 검사 뒤 push(M-Flow, PR 없음).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 증명 단위는 `TempHome::snapshot`(집 전체, 폴더 mtime 포함) | 기준 6 의 대상 전부를 한 비교로 | 대상별 해시로 |
| D-2 | `card next` 는 살아 있는 리스만 있는 세상에서 증명 | next 의 끝난 리스 표시는 설계상 쓰기 — 무해한 세계에서 증명 | next 를 증명에서 빼기 |
| D-3 | 옛 체계의 기록은 실제 경로 양상(journal.log·proposals 은 프로젝트 상태 옆, 세션 항목은 clone 공용 구역)에 심는다 | 대상이 존재해야 증명이 빈 것으로 통과하지 않음 | 대상을 빼기 |
| D-4 | 저장소는 common::git 가드로 만든 실제 git repo 와 그 worktree | 리더 지시(임시 저장소); 예약이 기록할 worktree 가 실제여야 함 | 평범한 폴더로 |

## Risks

- `card next` 가 끝난 리스 표시 외의 뜻밖의 쓰기를 하면 이 시험이 붙잡아요 — 그 경우 이 조각의 범위(시험만)를
  벗어나므로 카드에 보고하고 수정 조각을 기다립니다.
- mtime 해상도가 낮은 파일시스템에서 우연 통과가 가능해요 — 내용 비교가 함께 있어 완화돼요.

## Result

시험 커밋 `09b016cea7`, SPEC 커밋 `7fdf7cf891`, 형식 수정 `9f075a776b`, base `330c84bbdc`. 시험·검사는
heavy-slot(agent-21-c43) 안에서 이 worktree 의 `target/` 로 돌렸다(2026-11-19T19:5xZ–20:1xZ).

- RED: 첫 실행은 두 시험 모두 실패 — 원인은 시험 준비 결함이었다: `card list`·`card next` 는
  `--project` 플래그가 없고 위치 인자 `[PROJECT]` 를 받는데 초안이 플래그를 썼다(usage 거절로 측정).
  고친 뒤 2 통과. 성질 자체는 처음부터 지켜졌다.
- AC-1: met — `cargo test --locked --offline --test ledger_readonly_cli` 2 passed 0 failed(초록 확인
  재실행 포함 세 번).
- AC-2(변이 탐침): met — `src/card/command.rs` `list()` 에 임시 HOME 아래 `mutation-probe.txt` 를 쓰는
  한 줄을 넣었더니 `show_and_list_leave_the_home_byte_for_byte_as_it_was` 가 실패했다(1 passed 1
  failed; 어긋남 목록에 탐침 파일 이름 등장). `git checkout --` 로 복원해 status 깨끗함. 변이는
  커밋되지 않았다.
- AC-3: met — home_isolation_scan 18 · card_dispatch_cli 25 · card_work_cli 26 · ledger_open 26 통과
  (0 실패), cargo fmt --check 통과(첫 검사 1 → 시험 파일 줄바꿈 수정 뒤 0), rustfmt --edition 2021
  --check src/ops/*.rs 0, clippy --offline --locked --all-targets -- -D warnings 0, git diff --check
  base...HEAD 0, 커밋 꼬리표 interpret-trailers --parse 0, spec check 0 errors 0 warnings(첫 검사
  req_gears 1·req_failure_path 3 → REQ 문형 수정 뒤), spec_cli 18 통과(첫 검사 17 통과 1 실패 → 수정
  뒤), python 네 모듈(repo_scans·source_manifest·local_check·skill_references) 0.
- 돌리지 않은 것: 전체 cargo test(PR 없음 — 리더 묶음 ci 가 인용한다), codex 샌드박스 안 재실행(이
  조각은 실제 HOME·네트워크를 만지지 않는다).
