---
id: SPEC-20261120-rule-admission-pruning
title: "Rules earn their place, and the list is pruned on a schedule"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# Rules earn their place, and the list is pruned on a schedule

## Intent

검증 2차 설계(사용자 결정 2026-11-20)의 W2-7: 규칙 후보는 근거가 있어야 들어오고, 들어온 목록은 주기적으로 가지치기돼요. 지금 rule-history.md 에는 내보내는 기준만 있고(5-7행), 들이는 기준·막는 규칙의 반증 단계·정기 가지치기는 없어요. 시험·검증 단계·안내 파일은 그냥 두면 자라기만 해요.

이 조각이 다루는 것 넷: (1) 들이는 기준 문장 — 서로 다른 근거 둘(PR·카드·사고) 또는 데이터 손실·보안 사고 하나, diff 발동 조건이나 고정 시험, 결정 기록 — 을 AGENTS.md 의 진화 절과 rule-history.md 에 적고, (2) 막는 규칙만 최근 병합된 PR diff N개에 되돌려 걸린 수와 오탐 수를 결정 기록에 적는 반증 단계(검사기 old/new 비교와 같은 방식, 에이전트 세 개 쓰지 않음), (3) 릴리스 점검표의 가지치기 단계 — rule-history.md 기계 목록에서 text only·근거 하나인 규칙을 후보로 올리고 삭제는 리더 결정, (4) AGENTS.local.md 크기 시험.

착지 분량(이 SPEC 의 첫 묶음)은 (3)과 (4)예요: (1)과 (2)의 문장 자리인 AGENTS.md 는 다른 카드가, rule-history.md 도 다른 카드가 각각 예약 중이라 그 반납 뒤 scope 를 넓혀 넣어요. Result 는 무엇이 아직 없는지 적어요.

## Non-goals

- 공용 가지치기 명령이나 자동 삭제: 후보 나열까지만, 삭제는 리더 결정.
- 기존 규칙의 문장 변경: 이 조각은 추가만.
- Out of scope: 네트워크, 카드 번호의 출력 사용.

## Requirements

- REQ-1: The AGENTS.md evolution section and rule-history.md shall state the admission rule: a new rule brings two distinct evidences or one data-loss or security incident, a diff trigger or a pinned test, and a decision record.
- REQ-2: The same places shall state the falsification step for blocking rules only: replay the rule over the recent merged PR diffs and record the hit count and the false positives in the decision record before it enforces.
- REQ-3: The release checklist shall carry a pruning step that names rule-history.md as the machine list to scan, selects the text-only single-evidence rules as candidates, and leaves the deletion to the leader's recorded decision.
- REQ-4: A pinned test shall bound AGENTS.local.md's size just above its size at this change, so silent growth is noticed.

## Acceptance

- AC-1 (REQ-1, REQ-2): a pinned test finds the admission sentence and the falsification step in both files.
  verify: `cargo test --locked --offline --test repo_agents_md`
- AC-2 (REQ-3): a pinned test finds the pruning step in the release checklist.
  verify: `cargo test --locked --offline --test repo_agents_md`
- AC-3 (REQ-4): the size test passes with the bound just above the file's current size.
  verify: `cargo test --locked --offline --test repo_agents_md`

## Scope

- write: `.moai/specs/SPEC-20261120-rule-admission-pruning/spec.md`, `docs/release-checklist.md`, `tests/repo_agents_md.rs`
- read-only: `AGENTS.md`, `skills/moai/references/rule-history.md`
- forbidden: 실제 HOME, 네트워크, 카드 번호의 출력 사용

## Plan

1. RED: 점검표 단계와 크기 시험의 고정 시험을 더해 지금 파일로 깨져요.
2. GREEN: 점검표에 (l) 규칙 목록 가지치기 후보 단계를, repo_agents_md.rs 에 AGENTS.local.md 크기 한계(현재 크기 바로 위)를 더해요.
3. AGENTS.md·rule-history.md 의 문장(REQ-1·REQ-2)과 그 고정 시험(AC-1)은 두 예약이 반납하면 같은 가지에서 scope 를 넓혀 완성해요 — Result 가 빠진 것을 적어요.
4. 관련 시험·스캔·trailer-check·tier 1 감사·READY.

## Risks

- 크기 한계를 넉넉히 잡으면 성장을 못 봐요: 현재 31,732 바이트 위의 32 KiB 로 고정해요.
- 점검표 단계가 리더의 실제 흐름과 어긋날 수 있어요: 후보 나열까지만 요구해 삭제 결정은 그대로 리더에게 있어요.
