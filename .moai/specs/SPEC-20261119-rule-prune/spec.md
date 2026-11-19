---
id: SPEC-20261119-rule-prune
title: "Verify I: prune the verification rules to patterns seen in ≥2 PRs"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261118-sdd-foundation
---

# Verify I: prune the verification rules to patterns seen in ≥2 PRs

## Intent

검증 규칙은 놔두면 자라기만 해요("they only grow if you let them" — Addy Osmani, 'How the Claude Code team thinks about verification', claude.dev 2026-11-16, 리더 보고 표 2; 사용자 지시 '모두 카드 발행해서 진행'). 이 SPEC은 카드 C-103의 문서 조각: mode-agent.md·internal/core-rules.md·workflows/audit.md(+ 대상에서 빠진 mode-leader.md 도 목록은 작성)의 규칙 문장마다 근거 Issue/PR 을 밝혀, 근거가 하나뿐이고 diff 계기가 없는 규칙은 references/rule-history.md(신규, 리더 결정 ①: internal 밖)로 옮기고, 중복은 합쳐요. 살아남은 규칙은 각자 강제하는 시험·훅·게이트를 이름 붙이거나 'text only' 로 표시해요. 리더 결정(note 2527·2530 수렴, 기본값 ①로 진행): mode-leader.md 편집과 source-manifest.json 해시 갱신은 묶음 2(#3225) 병합 뒤 마지막 단계로 미뤄요.

## Non-goals

- mode-leader.md 편집·source-manifest.json 갱신(마지막 단계, 이 조각 밖).
- 규칙의 의미 변경(삭제·이동·중복 통합만).
- C-96 precheck 의 mode-leader.md 반영(note 2369 — mode-leader.md 편집 단계에서 함께).

## Requirements

- REQ-1: When the rule inventory is written, every rule sentence of mode-agent.md, internal/core-rules.md, workflows/audit.md and mode-leader.md shall be listed in references/rule-history.md with its source issue or PR.
- REQ-2 (REQ-1): IF a rule has exactly one source and no diff trigger (no test, hook or gate reads it), THEN it shall move to the history section of rule-history.md and leave the live file.
- REQ-3: When two live rules state one pattern, the pruned files shall keep one merged sentence and name the merged-away duplicate in the inventory.
- REQ-4 (REQ-3): IF a kept rule is enforced by a test, hook or gate, THEN the inventory shall name that enforcer, and otherwise say `text only`.
- REQ-5 (REQ-1): IF a rule names a path, command or check that no longer exists at the tip, THEN it shall move to the history section too.
- REQ-6: When the slice completes, the inventory shall report each edited file's byte count against its base count (criterion 1's "bytes reported").
- REQ-7 (REQ-6): IF an edited file did not get smaller, THEN the inventory shall name the rule that had to stay and why.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-6): 규칙 목록이 네 파일의 규칙 문장을 빠짐없이 근거와 나열하고(기계 추출+검수), 세 편집 파일이 작아지며 바이트 수를 보고해요.
  verify: `python3 -m unittest tests.repo_scans_test` (스킬 검사 수트) 및 목록 파일 직접 검수; `git diff --stat`
- AC-2 (REQ-2, REQ-4, REQ-5): 근거 1개·계기 없음 규칙과 tip 에 없는 대상을 가리키는 규칙이 history 절로 이동하고, 살아남은 규칙마다 강제자(test·hook·gate 이름) 또는 'text only' 표시가 있어요.
  verify: 목록 검수 + `python3 -m unittest tests.repo_scans_test`
- AC-3 (REQ-1…REQ-6): 이웃 검사 — repo_scans·skill_references·ci_scope 수트와 workflow_size_budget(`cargo test --offline --locked --test workflow_size_budget`), fmt·clippy; source-manifest 갱신은 마지막 단계 전까지 알려진 빨강으로 기록해요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.repo_scans_test tests.skill_references_test tests.ci_scope_test`; `cargo test --offline --locked --test workflow_size_budget`

## Scope

- write: `skills/moai/workflows/mode-agent.md`, `skills/moai/references/internal/core-rules.md`, `skills/moai/workflows/audit.md`, `skills/moai/references/rule-history.md`(신규, internal 밖 — 리더 결정 ①), `.moai/specs/SPEC-20261119-rule-prune/spec.md`
- deferred (리더 2527·기본값 ①): `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json`
- forbidden: 규칙 의미 변경, 실제 HOME·원장·키, 네트워크 쓰기

## Plan

1. 기계 목록: 네 파일에서 규칙 문장(shall/never/only/always/Do not/… 패턴 + Issue #nnnn 인용)을 추출해 target/c103/inventory.md 초안 작성, 근거 Issue 태그와 강제자(시험 파일명 rg 대응)를 붙임.
2. 판정: 근거 1개·계기 없음 → history 이동; 중복 → 통합; tip 부재 대상 → history. 검수 결과를 목록에 반영.
3. 편집: 세 파일 가지치기 + rule-history.md 작성(살은 live 표, 옮긴 것은 history 절).
4. 검사: repo_scans·skill_references·ci_scope·workflow_size_budget + 바이트 보고.

## Risks

- '규칙 문장'의 경계가 주관적이에요 → 기계 추출 패턴을 명시하고 검수 기록을 목록에 남겨요.
- 강제자 매핑이 빠지면 criterion 2 위반 → 근거 없는 규칙은 'text only' 로 명시적으로 표시해요.
- 스킬 파일 축소가 다른 검사의 기대를 깰 수 있어요 → repo_scans·skill_references 수트가 회귀망.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | mode-leader.md 편집·manifest 갱신은 마지막 단계로 연기 | 리더 2527(C-99 경로 충돌), 2530 의 기본값 ① | 조각 전체 통째 대기 |
| D-2 | history 는 references/rule-history.md 한 파일(live 표+history 절; internal 밖 — 리더 결정 ①, 내부 목록 시험·agent reads 요건 회피) | 카드 스코프가 그 경로를 이름 | 파일별 history |
| D-3 | 목록은 기계 추출+검수, 판정 기준(근거 수·계기·tip 대상)을 목록 머리에 명시 | 재검 가능성 | 손목록 |

## Result

- AC-1: met — 네 파일의 규칙 문장이 references/rule-history.md 에 수록돼요: core-rules 21행 수기 판정표 + audit 33·mode-agent 60·mode-leader 41행 기계 목록(추출기·blame·핀 대응 방법 명시). 바이트(wc -c): core-rules.md 3,961→3,747 · audit.md 20,334→19,884 · mode-agent.md 47,928→변경 없음(REQ-7: 문장마다 인라인 근거+86개 핀 문구라 이동 대상 없음 — 목록만) · mode-leader.md 마지막 조각 대기(리더 결정).
- AC-2: met — core-rules 에서 근거 1·계기 없음 11개+꼬리 2개가 history 절로 이동(번호는 핀 때문에 포인터 유지); 생존 규칙마다 강제 시험 이름 또는 'text only' 표기. Web 절 병합은 제외됨(리더 ①; #2517 세-복사 동일 핀).
- AC-3: 부분 — repo_scans 25·skill_references 5·skill_web_pages 4·agent_defs internal_references 1/1·core_rules_2876 6/6 OK; tip(bdb4d3db8c) 빌드 trailer-check findings 0. cargo test --lib 전체·home_isolation_scan·clippy --all-targets·fmt·workflow_size_budget 은 push 전 최종 배터리로 돌려 결과와 함께 LANE-READY 에 적어요; source-manifest 갱신은 마지막 조각(리더 결정)으로 이 시점 알려진 빨강이에요.
