---
id: SPEC-20261117-legacy-verdicts
title: "레거시 자산 1,107행 판정 — 상태 어휘와 고정 시험"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

`.moai-cli/manifests/legacy-asset-map.json`의 Go 세대 자산 1,107행이 전부 `unchecked`였고,
`tests/product_cli.rs`의 고정 시험이 그 상태를 강제했어요. 2026-11-17 사용자 결정 "개념은 새 설계로, 파일은
안 옮김"(D8 = A 유지)은 행마다 판정을 주고 `unchecked` 행을 0으로 만들라고 해요. 이 SPEC은 상태 어휘
`adopted(<path>)`·`rejected(<reason>)`·`deferred(<design or SPEC id>)`를 넣고, 모든 행을 판정하고, 시험이 어휘와
판정마다 경로·사유·대상을 요구하게 바꿔요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `When` 요구가 없어 실패 경로 요구도 없어요.

- REQ-1: Each row of the legacy asset map shall carry one status of the form `adopted(<path>)`, `rejected(<reason>)` or `deferred(<design or SPEC id>)`.
- REQ-2: The legacy asset map shall hold no row whose status is `unchecked`.
- REQ-3: Each `adopted` verdict shall name a file of this repository that carries the row.
- REQ-4: Each `rejected` verdict shall state a reason of at least three words.
- REQ-5: Each `deferred` verdict shall name one of the four designs of the 2026-11-17 decision or an existing SPEC folder.
- REQ-6: The source manifest of the skill package shall pin the map's sha256 and its verdict counts.
- REQ-7: The change shall create no `.moai-cli/manifests/host-capabilities.json`, `resources/` or `adapters/codex/` path.
- REQ-8: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-legacy-verdicts/spec.md`.

## Acceptance

- AC-1 (REQ-1…REQ-6): 고정 시험이 어휘·경로·사유·대상·개수를 확인해요.
  verify: `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`
- AC-2 (REQ-2): `unchecked` 행이 없어요.
  verify: `grep -c '"status": "unchecked"' .moai-cli/manifests/legacy-asset-map.json` → `0`
- AC-3 (REQ-6): 판정 개수가 manifest와 같아요.
  verify: `python3 -c "import json,collections;print(sorted(collections.Counter(r['status'].split('(')[0] for r in json.load(open('.moai-cli/manifests/legacy-asset-map.json'))['assets']).items()))"` → `[('adopted', 16), ('deferred', 286), ('rejected', 805)]`
- AC-4 (REQ-7): 목적지 충돌 경로를 만들지 않았어요.
  verify: `ls -d .moai-cli/manifests/host-capabilities.json resources adapters/codex` → 셋 다 없음(exit 1)
- AC-5 (REQ-1…REQ-8): 바뀐 곳에 닿는 검사가 통과해요.
  verify: `moai spec check --all --project .`(이 가지에서 빌드한 바이너리); `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `.moai-cli/manifests/legacy-asset-map.json`, `skills/moai/references/source-manifest.json`, `tests/product_cli.rs`,
  `docs/rust-product-path.md`, `.moai/specs/SPEC-20261117-legacy-verdicts/spec.md`
- read-only: `docs/decisions-2026-11-08.md`, `skills/moai/**`, `templates/claude/output-styles/**`, `src/agents.rs`, `agents/**`
- forbidden: `.moai-cli/manifests/host-capabilities.json`, `resources/`, `adapters/codex/` 생성, 실제 HOME, 네트워크 쓰기

## Plan

1. RED: 고정 시험을 새 어휘로 바꾸고, 옛 지도(전부 `unchecked`)에서 실패하는지 확인.
2. 판정: 경로 규칙으로 1,107행을 판정하는 일회용 스크립트(커밋 안 함). 다시 쓴 JSON이 원래 바이트와 같은 모양인지
   먼저 확인하고 `status`와 `purpose`만 바꿔요.
3. manifest의 sha256·`status`·`verdicts`, `docs/rust-product-path.md`의 "Every row stays `unchecked`" 문장.
4. 검사, 커밋, mutation 3개, `## Result`.

## Decisions

- D1 채택 16행(사용자 요청 목록, 카드 노트): `.claude/skills/moai/SKILL.md`·`workflows/{plan,run,sync}.md`(각 2벌) →
  `skills/moai/…` 같은 이름 파일, `AGENTS.md`·`AGENTS.md.tmpl` → `src/agents.rs`, 출력 스타일 3개(각 2벌) →
  `templates/claude/output-styles/`(카드 노트의 `adapters/claude/output-styles/`는 #2709에서 옮겨졌어요).
- D2 보류 대상 이름: `design:code-atlas-read-plan`(codemaps, 프로젝트 codebase-analysis, plan context-discovery),
  `design:language-domain-guides`(언어 규칙, quality 규칙, coding-standards, `moai-domain-*`, `moai-ref-*`,
  `moai-foundation-quality`), `design:project-agent-skill-generation`(harness 작업 흐름·입구·명령, builder-harness 역할,
  agent/skill 작성 규칙, `moai-meta-harness`, `moai-harness-learner`, `moai-jev-skill-suggestion`),
  `design:edit-time-diagnostics-hook`(LSP·진단 행이 지도에 없어 0행).
- D3 거절 사유 묶음(판정 문자열에 그대로): 입구(단일 `/moai`), 라우터 작업 흐름(`skills/moai/workflows`), `@MX`(ACC
  `@cc:` 설계가 대체), 규칙(AGENTS.md·`skills/moai/references`; host-capabilities.json·`resources/`·`adapters/codex/`
  목적지 없음 — 목적지 충돌 정리), 역할(`agents/`), 작업 방법 skill, 기반 지식 skill, 저장소 전용 도구, 운영 skill,
  보관(은퇴), Go 설치 목록.
- D4 `@MX` 행은 보류가 아니라 거절이에요: 네 설계에 들지 않고, ACC 설계가 옛 표식을 읽기만 한다고 이미 정했어요.
- D5 판정 문자열을 바꾸는 일은 스크립트로 했어요(1,107행이라 손 편집 대신). 스크립트는 커밋하지 않았어요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `c0b858de8c`(origin/v2026.11.0), 변경 커밋
`0cc5c659c4`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는
한 번에 하나. `moai heavy-slot-acquire`가 75(스왑 82%)라 리더 지시대로 슬롯 밖에서 돌렸어요.

판정 개수(종류별): output-style adopted 6 · project-instruction adopted 2 · skill-entry adopted 2, deferred 52,
rejected 83 · support adopted 6, deferred 184, rejected 441 · rule deferred 42, rejected 190 · command deferred 4,
rejected 36 · agent deferred 2, rejected 32 · native-agent-projection deferred 2, rejected 21 · native-mapping-manifest
rejected 1 · distribution-catalog rejected 1. 거절 묶음: 작업 방법 skill 200, 규칙 182, 기반 지식 114, 보관 77,
라우터 작업 흐름 68, 입구 57, 저장소 전용 47, 역할 43, `@MX` 10, 운영 5, 설치 목록 2.

- AC-1: met — RED(지도 바꾸기 전): exit 101, `tests/product_cli.rs:3105` "`unchecked` is not a verdict". GREEN exit 0
  (1 passed). mutation 3개(커밋 뒤, 지도 한 행과 manifest sha256을 함께 바꿔 sha 검사를 지나게 함, 각각 `git restore`
  뒤 tree 깨끗): `rejected()` → exit 101 "is not a verdict"; `adopted(skills/moai/NO-SUCH.md)` → exit 101 "is not a
  file of this repository"; `deferred(later)` → exit 101 "deferred to unknown `later`". 되돌린 뒤 exit 0.
- AC-2: met — `0`.
- AC-3: met — `[('adopted', 16), ('deferred', 286), ('rejected', 805)]`, manifest `verdicts`와 같음.
- AC-4: met — 세 경로 모두 "No such file or directory", exit 1.
- AC-5: met — `moai spec check --all --project .` exit 0(SPEC 3개, errors 0, warnings 0); `product_cli` 0(75);
  `home_isolation_scan` 0(16); `--bin moai` 0(27); `go_state_classes_cli` 0(71)·`launcher_three_hosts` 0(4)
  (`docs/rust-product-path.md`를 읽는 시험); `ac_command_lint` 0(5); `cargo fmt --check` 0(처음 1, 한 줄 줄바꿈 뒤 0);
  Python `repo_scans`·`ci_scope`·`source_manifest` 83개 OK.

실행 안 함: 전체 `cargo test`(바뀐 곳 시험만), clippy, Linux·Windows. 편집 시점 진단 hook 설계로 보류한 행은 0이에요
(지도에 LSP·진단 행이 없음).
