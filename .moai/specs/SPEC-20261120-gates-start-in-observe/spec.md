---
id: SPEC-20261120-gates-start-in-observe
title: "New blocking gates start in observe"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# New blocking gates start in observe

## Intent

검증 2차 설계(사용자 결정 2026-11-20)의 W2-5: 새 게이트는 observe 로 시작한다. observe 는 "막았을 것"만 기록해서, 누구도 막히기 전에 오탐을 볼 수 있게 해요. 지금 이 일반 규칙은 어디에도 문장으로 없고, SPEC 양식에도 힌트가 없어요.

이 조각은 규칙 문장 하나와 양식의 한 줄이에요: SDD 안내(`skills/moai/references/internal/sdd.md`)에 규칙을 적고, SPEC 양식(`templates/spec/full.md`)에 한 줄을 더하며, 고정 시험이 둘 다 있음을 검사해요. 공용 모드 기계 장치는 만들지 않아요(지금 필요한 곳은 따로 있어요).

## Non-goals

- 공용 observe/enforce 모드 기계 장치: 만들지 않아요.
- 기존 도구(check-guard·audit-gate·trailer-check·repo-scans)의 동작 변경: 이 조각은 문서와 양식만.
- templates/spec/minimal.md: minimal 은 그대로.
- Out of scope: 네트워크, 카드 번호의 출력 사용.

## Requirements

- REQ-1: The SDD guidance shall state that a new check which blocks starts in observe, that the observation window's would-block count and false positives go into the SPEC's Result, and that the switch to enforce is a recorded decision.
- REQ-2: The full SPEC template shall carry one line reminding the author of the same rule.
- REQ-3 (REQ-1): IF the guidance sentence changes, THEN the manifest entry of the guidance file shall carry its new digest, so the source-manifest check stays green.

## Acceptance

- AC-1 (REQ-1, REQ-2): a pinned test finds the rule sentence in the SDD guidance and the reminder line in the full template.
  verify: `cargo test --locked --offline --test sdd_guidance`
- AC-2 (REQ-3): the manifest check for the guidance file passes with the new digest.
  verify: `python3 -m unittest tests.source_manifest_test`
- AC-3: the reference checks around it and the style checks pass.
  verify: `cargo test --locked --offline --test agent_defs_package`
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261120-gates-start-in-observe/spec.md`, `skills/moai/references/internal/sdd.md`, `templates/spec/full.md`, `tests/sdd_guidance.rs`, `skills/moai/references/source-manifest.json`
- read-only: `templates/spec/minimal.md`
- forbidden: 실제 HOME, 네트워크, 카드 번호의 출력 사용

## Plan

1. RED: sdd_guidance 에 고정 시험을 더해 지금 코드(문장 없음)로 깨져요.
2. GREEN: sdd.md 에 규칙 항목을, full.md 에 양식 줄을 더하고, manifest 도구로 해시를 갱신해요.
3. 관련 시험·스캔·trailer-check·tier 1 감사·READY.

## Risks

- 양식 줄이 생성되는 SPEC마다 등장해요: HTML 주석으로 넣어 렌더 결과가 읽기에 방해되지 않게 해요.
- manifest 해시 갱신을 까먹으면 source_manifest_test 가 잡아요 — 그래서 REQ-3 으로 묶었어요.

## Result

- AC-1: met — `cargo test --locked --offline --test sdd_guidance` 7 개 통과(`new_blocking_gates_start_in_observe` 고정 시험 포함). RED 는 문장을 넣기 전에 같은 시험이 실패한 것으로 확인했어요. 구현 세션의 병합 준비 노트(`ready-integration`)와 tier 1 감사 노트(`verdict-b9111dcba1`)가 같은 수를 기록했어요.
- AC-2: met — `python3 -m unittest tests.source_manifest_test` 통과. sdd.md 의 새 digest 가 source-manifest.json 에 반영돼 있어요(같은 두 노트).
- AC-3: met — `cargo test --locked --offline --test agent_defs_package` 19 개 통과, `cargo fmt --check` 0. 이웃 검사는 구현 세션이 실행했어요(같은 노트).
- 이 Result 를 적는 세션은 문서만 고치므로 위 검사를 다시 돌리지 않았어요(착지와 내용은 git 으로 실측: 규칙 head 는 로컬 통합 가지 tip 의 조상). 변이는 없어요(코드 변경 없음). 상태는 사용자 승인 전까지 draft 로 남아요.
