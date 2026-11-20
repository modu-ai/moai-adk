---
id: SPEC-20261120-mode-flags
title: "Mode flags become --leader / --agent (short -l / -a), old spellings one release"
status: draft
revision: 1
created: 2026-11-20
---

# 모드 옵션 — `--leader`·`--agent`(짧은 `-l`·`-a`), 옛 철자는 한 릴리스 수용

## Intent

사용자 결정(2026-11-20, 리더 대화): `/moai --mode-leader/--mode-agent` 를 `--leader`/`--agent` 로
바꾸고 짧은 별칭 `-l`/`-a` 를 함께 받는다. 짧은 별칭은 `/moai`(`$moai`, `/skill:moai`) 바로 뒤
첫 단어일 때만 모드로 인정해 요청 글 속 `ls -a`·`grep -l` 이 모드로 잡히지 않게 한다. 옛 철자는
안내와 함께 한 릴리스 동안 수용한다. 2026-11-10 #2286(두 철자만)을 대체하는 결정 기록을 남긴다.

## Non-goals

- 워크플로 파일 이름 변경(mode-leader.md·mode-agent.md) — routes.json 은 경로 이름 키라 무변경.
- 과거 SPEC·보고서 기록물의 철자 수정 — 역사 기록은 그대로.
- `--autonomy full|semi` 변경 없음.

## Requirements

- REQ-1: When the request text carries `/moai --leader`, `/moai --agent`, or the short alias as the word right after `/moai` (`$moai`, `/skill:moai` included), the mode route shall be selected the same way `--mode-leader`/`--mode-agent` selects it today · failure: none — a short alias anywhere but right after the skill name selects no mode.
- REQ-2: When a request uses the old spellings `--mode-leader`/`--mode-agent`, they shall keep working for one release, with a notice naming the new spelling · failure: none — after the release window the notice turns into the plain refusal.
- REQ-3: The decisions log shall carry the 2026-11-20 row that names the replaced 2026-11-10 decision (#2286), and no shipped text shall instruct the old spellings · failure: none — a scan test pins both.

## Acceptance

- AC-1 (REQ-1): 신철자 선택 — `/moai --leader`·`--agent`·`-l`·`-a`(바로 뒤)·`/moai ls -a …`(모드 아님) 케이스. verify: `cargo test --locked --offline --test mode_flags_cli`
- AC-2 (REQ-2): 옛 철자 수용 + 안내 문구. verify: `cargo test --locked --offline --test mode_flags_cli`
- AC-3 (REQ-3): 스캔 시험 — 배포 텍스트에 옛 철자 안내 없음 + 결정 로그 행 존재. verify: `cargo test --locked --offline --test mode_flags_cli`; `python3 -m unittest tests.repo_scans_test`
- AC-4 (이웃): 이웃 검사. verify: `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test mode_workflow_steps`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write(현재 예약): `docs/decisions-2026-11-08.md`, `skills/moai/references/routes.json`(무변경 확인), `skills/moai/references/rule-history.md`(행 추가), `src/lane_duty.rs`, `src/lane_idle.rs`, `src/lane_loop.rs`, `src/leader_pulse.rs`, `src/ops/lane_duty_line.rs`, `src/ops/lane_loop_line.rs`, `.moai/specs/SPEC-20261120-mode-flags/**`, `tests/mode_flags_cli.rs`(신설 예정 — widen)
- blocked(다른 카드 예약 — 슬라이스 2): `skills/moai/SKILL.md`(C-200), `skills/moai/workflows/mode-agent.md`·`mode-leader.md` + `source-manifest.json` + mode_cards_53 등(C-151), `src/main.rs`·`src/session.rs`(C-140), `src/leader_round.rs`(C-144), `src/hook_engine.rs`(C-173), `src/help.rs`(C-183), `chains/index.md`(C-217), `AGENTS.local.md`(C-204)
- forbidden: 과거 SPEC·기록물 철자 수정, 실제 HOME, 카드 번호를 SPEC id·가지·커밋에 사용

## Plan

1. 슬라이스 1(지금): 결정 로그 행 + rule-history 행 + routes.json 무변경 기록. 내보내는 쪽(src) 철자 전환은 읽는 쪽(SKILL.md)이 새 철자를 받아들인 뒤 — 순서 제약은 아래 Decisions D-2.
2. 슬라이스 2: SKILL.md·워크플로·manifest(C-151 반환 뒤) — 새 철자 수용 + 짧은 별칭 인접 규칙 + 옛 철자 안내.
3. 슬라이스 3: src 내보내기 전환 + 도움말 + 골든 + 스캔 시험(모든 표면 파일 반환 뒤).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 읽는 쪽(SKILL.md)이 먼저 새 철자를 받아들이고, 내보내는 쪽(src 프롬프트 문장)이 나중에 바뀐다 | 읽는 쪽이 옛 철자만 아는 동안 내보내기가 바뀌면 모드 판정이 깨짐 | 동시 전환(깨짐 위험) |
| D-2 | routes.json 무변경 — MODE-AGENT/MODE-LEADER 는 경로 이름 키이지 철자 키가 아님 | routes.json:46-53 확인 | 키 이름 변경 |
| D-3 | 결정 행은 docs/decisions-2026-11-08.md 에 2026-11-20 으로 추가, 대체 행(#2286·11-10 두 철자) 명시 | 기준 2 — 대체되는 결정을 이름으로 | 행 삭제 |

## Risks

- 표면이 여러 live 카드 예약에 흩어져 있어 전환은 슬라이스 2·3 이 이어져야 완성돼요 — 슬라이스 1 은 결정 기록과 순서 고정만 해요.
- 옛 철자 수용 기간의 안내 문구는 SKILL.md 슬라이스(2)에서 같이 나가야 사용자 혼란이 없어요.
