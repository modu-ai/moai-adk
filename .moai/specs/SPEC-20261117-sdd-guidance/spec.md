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
- `moai card spec` 같은 카드↔SPEC 연결 명령: 별도 변경(SDD-4)이 넣어요. 이 SPEC은 그것을 "들어오면 쓸 수 있는
  예정 기능"으로만 적어요.
- mode 파일(`mode-agent.md`, `mode-leader.md`): 이미 카드 원장 기준으로 다시 쓰여 있어 건드리지 않아요.
- `sync.md`의 PR·병합 단계(`v2026.11.0`, `Refs #<n>`): 관문 조각과 함께 바꿔요.
- `@cc:` 표기의 정확한 문법: 별도 SPEC.

## Requirements

요구 문장은 영어 GEARS로 써요. `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: The `moai` skill shall keep the agent-facing SDD rules in one internal reference, `internal/sdd`.
- REQ-2: The reference shall name only `moai spec` actions this build has and mark the card ↔ SPEC link as planned until it lands.
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
  `moai spec` 동작이 모두 `--help`로 열리고, 카드 연결 명령이 없는 동안 그 줄은 모두 "planned … once that change
  lands"라고 말해요.
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
- D2 카드 연결 명령은 이 빌드에 없어서(`moai card spec --help` exit 2) "planned, available once that change
  lands"로 적었어요(리더 결정: 에이전트는 이 참조에서 명령을 배워요). 시험은 명령이 생기면 그 조건을 스스로 풀어요.
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

(구현 뒤 채워요.)
