---
id: SPEC-20261117-sdd-guidance
title: SDD 지침 정리 — internal/sdd 하나, 순서 PLAN → RUN → VERIFY → SYNC → AUDIT → merge (SDD-5)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

MoAI는 항상 SDD로 일해요(사용자 결정 2026-11-17). 그런데 skill 지침은 아직 GitHub Issue 시대의 말과 순서를
담고 있었어요: PLAN이 "GEARS SPEC"을 약속하지만 쓰는 방법이 없고, 빌드 chain은 audit 뒤에 docs를 두고, 버그 수정
chain에는 PLAN이 없고, 개발 방법을 "DDD"라고 불렀어요. 이 SPEC은 SDD 실행 순서 7단계(조각 SDD-5)예요: 에이전트가
읽는 SDD 규칙을 한 곳(`internal/sdd`)에 두고, workflow와 chain 표가 한 순서를 말하게 하고, "DDD"를 "보존 우선
변경(preserve-then-change)"으로 바꿔요. 지침만 바꾸고 CLI는 바꾸지 않아요.

## Non-goals

- CLI 변경(`moai spec take`·`close`, 관문의 `Moai-Session` 읽기, 커밋 가드의 `spec/` 가지): 다른 조각(SDD-6~8).
- `moai card spec` 같은 카드↔SPEC 연결 명령 자체: 별도 변경(SDD-4)이 넣었어요. 이 SPEC은 그 사용법만 적어요.
- mode 파일(`mode-agent.md`, `mode-leader.md`): 이미 카드 원장 기준으로 다시 쓰여 있어 건드리지 않아요.
- `sync.md`의 PR·병합 단계(`v2026.11.0`, `Refs #<n>`): 관문 조각과 함께 바꿔요.
- `@cc:` 표기의 정확한 문법: 별도 SPEC.

## Requirements

요구 문장은 영어 GEARS로 써요. `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: The `moai` skill shall keep the agent-facing SDD rules in one internal reference, `internal/sdd`.
- REQ-2: The reference shall name only `moai spec` actions and card ↔ SPEC link commands that this build has.
- REQ-3: Each ordering document shall state the order PLAN → RUN → VERIFY → SYNC → AUDIT → merge.
- REQ-4: The build and bug-fix chains shall run PLAN, RUN, VERIFY, SYNC and AUDIT with `moai-docs` before `moai-audit`.
- REQ-5: When SYNC runs before the audit, the SYNC workflow shall write the SPEC's Result and status, the documents and the existing `@cc:` comments on the candidate branch. · failure: none — 맞출 수 없는 차이는 같은 문단이 보고하고 멈추게 해요.
- REQ-6: The AUDIT workflow shall treat drift between the SPEC, the code, the tests, the documents and the `@cc:` comments as a FAIL finding and fix nothing.
- REQ-7: The PLAN workflow shall write or update the SPEC with `moai spec new` and record the user's selection as the approval with its digest.
- REQ-8: The RUN workflow shall describe changes to existing code as preserve-then-change, with characterization tests committed first.
- REQ-9: The skill package and the agent definitions shall not use the word "DDD".
- REQ-10: The `moai-docs` agent shall read `internal/sdd`.
- REQ-11: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-sdd-guidance/spec.md`, tracked by git.

REQ-3의 "ordering document"는 chain 표(`chains/index.md`)와 RUN·VERIFY·SYNC·AUDIT workflow예요.

## Acceptance

- AC-1 (REQ-1, REQ-2): 참조가 사용자 결정을 적고(SPEC 대상, minimal 기본, GEARS + IF…THEN과 `(REQ-n)`, `· failure:
  none`, 답 없는 질문 표시(NEEDS CLARIFICATION), 선택 승인과 digest, 한 SPEC = 한 PR, `Moai-Spec`, `Moai-Session`은 예정), 적은
  `moai spec` 동작이 모두 `--help`로 열리고, 카드 연결 명령(`moai card spec`, `--unlink`, `card next`·`take`의 SPEC
  상태, `--require-approved-spec`)은 이 빌드에 있을 때 "planned" 없이 적혀요(없는 빌드에서는 반대로 "planned"를
  요구해요).
  verify: `cargo test --locked --offline --test sdd_guidance`
- AC-2 (REQ-3, REQ-5, REQ-6, REQ-7, REQ-8): chain 표와 RUN·VERIFY·SYNC·AUDIT에 순서 문장이 있고, 옛 순서 문자열이
  없고, PLAN·RUN·SYNC·AUDIT이 `internal/sdd`를 가리켜요.
  verify: `cargo test --locked --offline --test sdd_guidance every_place_states_one_order`
- AC-3 (REQ-4): verify: `cargo test --locked --offline --test agent_defs_package chains_name_real_routes_agents_and_references`
- AC-4 (REQ-9): verify: `cargo test --locked --offline --test sdd_guidance no_card_number_or_ddd_wording_in_the_package_and_agents`;
  `grep -rn DDD skills agents` → 출력 없음, exit 1
- AC-5 (REQ-10): verify: `cargo test --locked --offline --test agent_defs_package reads_lists_match_the_design_and_render_on_every_host`
- AC-6 (REQ-11): verify: `git ls-files .moai/specs/SPEC-20261117-sdd-guidance` → 이 파일; `moai spec check --all --project .` → exit 0
- AC-7 (REQ-1…REQ-11): 이웃 검사가 통과해요.
  verify: `python3 -I tools/source_manifest.py check`; `moai package-inspect skills/moai` → exit 0;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/skill_references_test.py tests/source_manifest_test.py tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `skills/moai/references/internal/sdd.md`(새 파일), `skills/moai/references/chains/index.md`,
  `skills/moai/workflows/{plan,run,verify,sync,audit,project}.md`, `skills/moai/references/internal/interview.md`,
  `skills/moai/references/source-manifest.json`, `agents/docs.md`, `tests/sdd_guidance.rs`(새 파일),
  `tests/agent_defs_package.rs`, `tests/update_route.rs`, `.moai/specs/SPEC-20261117-sdd-guidance/spec.md`
- read-only: `src/spec_doc.rs`, `src/spec.rs`, `src/run.rs`, `.moai/specs/**`
- forbidden: `AGENTS.md`, `src/**`, `tools/**`, `install/**`, `.github/**`, `internal/build-folder.md`,
  `internal/code-atlas.md`, `workflows/read.md`, atlas 가지가 더한 `plan.md` 줄, `agents/impl.md`, `agents/audit.md`,
  `agents/research.md`, `agents/plan.md`의 `reads:` 줄, mode 파일, 실제 HOME, 네트워크 쓰기

## Plan

1. 도움말 확인: `moai spec new|check|list|show --help`(있음), `moai card spec --help`(없음, exit 2).
2. RED: `tests/sdd_guidance.rs`, `tests/agent_defs_package.rs`(chain 행·`reads:`·내부 참조), `tests/update_route.rs`.
3. GREEN: `internal/sdd`, chain 표, workflow 문장, `agents/docs.md`, 면접 질문, manifest.
4. 검사, 커밋 뒤 mutation(3개 이하), Result.

## Risks

- `plan.md`는 atlas 가지도 고쳐요: 이 SPEC은 Output과 2·5단계만 바꾸고 atlas 가지가 6단계 뒤에 더한 줄과 떨어져
  있어요. 둘 다 병합되면 `git merge-tree`로 다시 확인해요.
- `verify.md`는 지원 플랫폼 가지가 38줄을 고쳐요: 이 SPEC은 Procedure 6단계만 더해요.
- 지침은 강제가 아니에요: 순서·승인 규칙을 기계로 막는 일(hook, 관문)은 SDD-6과 카드 C-86 몫이에요.
- `@cc:` 문법이 아직 정해지지 않아 SYNC는 있는 주석만 맞추고 새로 쓰지 않아요.

## Decisions

- D1 요구 문장은 영어 GEARS, 본문 한국어, 제목 영어(저장소의 다른 SPEC과 같은 형식).
- D2 카드 연결 명령은 처음 기반에 없어서(`moai card spec --help` exit 2) "planned"로 적었다가, 카드 연결
  변경(#3160)이 병합된 tip을 합친 뒤 `--help`로 확인하고 쓸 수 있는 명령으로 고쳤어요(리더 결정: 에이전트는 이
  참조에서 명령을 배워요). 시험은 빌드에 명령이 있으면 "planned"를 금지하고, 없으면 요구해요.
- D3 버그 수정 chain: `moai-research, moai-plan, moai-impl, moai-docs, moai-audit`. 빌드 chain:
  `moai-plan, moai-impl, moai-docs, moai-audit`. VERIFY는 에이전트 없이 부모 세션이 해요.
- D4 SYNC의 글쓰기는 `moai-docs`가 해요: `agents/docs.md`의 쓰기 범위에 SPEC `## Result`·상태와, 범위에 든 코드
  파일의 기존 `@cc:` 주석 줄만 더했어요(새 주석은 쓰지 않아요).
- D5 개발 방법 이름: "preserve-then-change". 옛 프로필이 Go 세대 이름으로 적은 값은 같은 뜻으로 읽어요(코드
  `src/run.rs`는 이미 `preserve` 전략으로 읽어요). RUN 3단계의 "bounded context" 문장은 지웠어요(도메인 설계와 혼동).
- D6 `sync.md`는 앞에 "Before the audit" 절을 더하고 Input·Output의 Issue 말을 SPEC으로 바꿨어요. PR·병합 단계의
  Issue·`v2026.11.0` 문장은 관문 조각(SDD-6)이 바꿔요.
- D7 `Moai-Spec` 트레일러를 넣고, `Moai-Session`은 관문 조각이 정할 때까지 넣지 않아요.

## Result

구현 에이전트가 2026-11-17 09:02Z–09:21Z에 채웠어요(09:08Z–09:14Z에는 리더 지시로 atlas 감사 수리를 먼저 하려고
멈췄어요). SYNC와 AUDIT가 다시 확인해요. macOS arm64, cargo는 slot 밖에서 한 번에 하나(설치된 slot이 옛 swap
규칙으로 exit 75), `CARGO_BUILD_JOBS=4`, `CARGO_INCREMENTAL=0`, 빌드 폴더는 이 worktree의 `target/`.

커밋: 기준 `daf8c68c69` 위에 `eb70b57ef0`(WIP, 멈추기 전 저장), `8ae2d8aa2c`(나머지와 SPEC), 그 위에 이 Result.

도움말 확인(이 빌드): `moai spec new|check|list|show --help` 모두 exit 0; `moai card spec --help` exit 2(없음) →
카드 연결은 "planned, available once that change lands"로 적었어요.

- AC-1: met — `sdd_guidance` exit 0, 4 passed. RED(지침 전): `the_reference_states_…`(순서 문장 없음),
  `every_place_states_one_order`, `no_card_number_or_ddd_wording_…`(project.md의 DDD) 실패.
- AC-2: met — 같은 실행. 감사 B2 뒤 고침: 처음 증거는 순서 문장만 고정했고 REQ-8(보존 우선 변경)과 `@cc:` 규칙은
  시험이 없었어요(감사가 문장을 지워도 초록). `preserve_then_change_and_the_cc_rule_are_stated`가 run.md·sdd.md·
  sync.md·agents/docs.md의 다섯 문장을 고정해요(아래 감사 수리).
- AC-3: met — `agent_defs_package` exit 0, 17 passed. RED: chain 행 `"PLAN, RUN"`.
- AC-4: met — 같은 실행; `grep -rn DDD skills agents` → 출력 없음.
- AC-5: met — 같은 실행. RED: `moai-docs`의 `reads:`에 `internal/sdd` 없음.
- AC-6: met — `git ls-files` → 이 파일; `moai spec check --all --project .` exit 0, SPEC 4개 오류 0·경고 0(처음엔
  REQ-3 주어의 쉼표로 `req_gears` 오류와 AC 글의 표시로 경고 1개 → 고침).
- AC-7: met — manifest `write`(9개 파일 바뀜) 뒤 `check` 0; `package-inspect skills/moai` exit 0, 65,536바이트 넘는
  파일 0(`mode-agent.md` 47,998바이트 그대로); `cargo fmt --check` 0(시험 두 파일 `rustfmt` 뒤); clippy 0; Python
  85개 OK(처음 `repo_scans_test`가 새 시험의 `read_dir`을 저장소 순회로 판정 → 이름 붙인 파일만 읽게 바꾼 뒤 OK).
  넓은 실행 28개 대상 모두 exit 0, 387 passed(`--bin moai` 27, `product_cli` 75, `mode_cards_53`,
  `mode_workflow_steps` 42, `core_rules_2876`, `lane_rules_2401·2836·2879`, `gate_collect`, `release_suite` 등).

mutation(커밋 뒤, 셋 다 되돌린 뒤 `git status` 비어 있음): 빌드 chain을 옛 순서로 → 2개 실패; 카드 연결을
"available"로 → 1개 실패; audit.md 순서 문장 삭제 → 1개 실패.

병합 확인(`git merge-tree`): atlas 가지와는 `plan.md`가 충돌 없이 합쳐지고, `source-manifest.json`·
`tests/agent_defs_package.rs`가 충돌해요(둘 다 같은 목록 끝에 항목을 더함 — 늦게 들어가는 쪽이 두 항목을 모두
남기고 `source_manifest.py write`). 지원 플랫폼 가지와는 `verify.md`가 충돌 없이 합쳐지고 `source-manifest.json`만
충돌해요(같은 방법).

실행 안 함: 전체 `cargo test`, Python 전체, 실제 호스트에서 이 지침대로 PLAN→…→AUDIT를 도는 일, Linux·Windows.

### 감사 수리 (2026-11-17 09:30Z–09:35Z)

독립 감사(head `3223cc9f01`, 09:21:50Z–09:29:13Z)가 FAIL이었어요. B1: tip `25881165be`와
`source-manifest.json`·`tests/agent_defs_package.rs`가 충돌. B2: 두 사용자 규칙(보존 우선 변경의 특성화 시험
먼저, 에이전트는 `@cc:`를 스스로 더하지 않음)을 고정하는 시험이 없음(감사의 S1–S3 문장 삭제가 모두 초록).
N1: 카드 연결 명령이 생기면 "planned"가 남지 않게. N4: `moai spec check (<id> | --all)`.

커밋: `3223cc9f01` 위에 `60d66ed95b`(tip 병합: `agent_defs_package.rs`는 두 쪽 항목 `sdd`와 `build-folder`를 모두
남기고, manifest는 `source_manifest.py write` 뒤 `check` 0) → `e60b25eeb4`(B2·N1·N4) → `084054f0dd`(새 시험
`rustfmt`) → 이 Result.

- B2 증거: 새 시험 `preserve_then_change_and_the_cc_rule_are_stated`. 수리 커밋 뒤 S1(run.md 문장 삭제)·S2(sdd.md
  "Agents never add them on their own" 삭제)·S3(agents/docs.md "never add one" 삭제) 모두 그 시험 실패(exit 101),
  되돌린 뒤 `git status` 비어 있음.
- N1: `moai card spec --help`가 성공하는 빌드에서는 그 줄에 "planned"가 있으면 실패하게 바꿨어요. 이 기반에는 그
  명령이 없어(exit 2) "있을 때" 쪽은 실행되지 않았어요(not_run) — 카드 연결 변경이 병합된 트리에서 돌아요.
- 다시 돌린 검사: `--bin moai` 27, `sdd_guidance` 5, `agent_defs_package` 18, `update_route` 8,
  `mode_workflow_steps` 42, `core_rules_2876` 4, `product_cli` 75 — 모두 exit 0; `cargo fmt --check` 0; clippy 0;
  Python 85개 OK; manifest `check` 0; `package-inspect skills/moai` exit 0, 65,536바이트 넘는 파일 0;
  `moai spec check --all` exit 0(SPEC 5개 오류 0·경고 0).

### 카드 연결 병합 뒤 (2026-11-17 09:53Z–09:58Z)

재감사는 `853f13a811`에서 PASS였지만, 카드 연결 변경(#3160)이 09:45:47Z에 병합돼 병합 트리에서 N1 시험이(의도대로)
실패했어요: "`moai card spec` exists: say it is available".

커밋: `853f13a811` 위에 `c1981a7111`(tip `d032d0be1f` 병합: `source-manifest.json`만 충돌, `write` 뒤 `check` 0) →
이 커밋(`sdd.md` 7번 항목 + 이 Result).

- 병합한 빌드의 도움말: `moai card spec --help` exit 0(`<CARD> [SPEC]`, `--unlink`, SPEC은 `moai spec check` 오류 없이
  통과해야 함, 같은 연결은 아무것도 안 씀); `moai card take --help`에 `--require-approved-spec`(승인된 SPEC이 없으면
  거절, 아무것도 안 잡음); `moai card next --help`가 `spec` 상태와 `warnings`를 적어요. `sdd.md` 7번을 이대로 "쓸 수
  있음"으로 고쳤어요.
- RED(병합 빌드 + 고치기 전 `sdd.md`): `the_reference_names_only_commands_this_build_has` 실패(위 문장). GREEN: exit 0.
- 다시 돌린 검사(병합 트리): `--bin moai` 27, `agent_defs_package` 18, `core_rules_2876` 5, `mode_workflow_steps` 42,
  `product_cli` 75, `sdd_guidance` 5, `update_route` 8 — 모두 exit 0; `cargo fmt --check` 0; clippy 0; Python 85개
  OK; manifest `check` 0; `package-inspect skills/moai` exit 0, 65,536바이트 넘는 파일 0; `moai spec check --all`
  exit 0(SPEC 10개 오류 0·경고 0).
