---
id: SPEC-20261117-atlas-in-read-plan
title: READ와 PLAN이 코드 아틀라스를 읽어요 (재설계 1)
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

사용자 결정(2026-11-17): Go 세대의 개념은 다시 설계하고 파일은 옮기지 않아요. 전수조사에서 Rust 제품에는 코드
아틀라스 명령(`index-build`, `index-query`, `atlas-map`, `atlas-graph`, `search`)이 이미 있는데, 어떤 workflow
지시도 그것을 쓰지 않는다고 나왔어요. PROJECT만 `index-build`를 제안해요(`workflows/project.md` 7단계).

이 SPEC은 READ와 PLAN이 아틀라스가 있고 신선할 때 그것으로 먼저 좁히고, 없거나 낡았으면 검색·읽기로 돌아가며,
스스로 짓지는 않게 해요. 자세한 방법은 짧은 내부 참조 `internal/code-atlas` 한 곳에 두고, 각 workflow에는 한 줄만
둬요.

## Non-goals

- 새 CLI 기능(신선도 플래그 등): 지금 명령에 필요한 것이 다 있어서 만들지 않아요(아래 Result의 측정).
- RUN·AUDIT·SYNC 등 다른 route, `SKILL.md`, mode 파일, `agents/impl.md`·`agents/audit.md`.
- 아틀라스 자동 빌드·갱신(감시자, hook). 짓는 일은 PROJECT의 사용자 동의 뒤에만 있어요.
- 심볼 단위 질문: 기록에 심볼이 없어요(`coverage.symbols`가 `unavailable`이라고 말해요).

## Requirements

요구 문장은 영어 GEARS로 써요. `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When READ or PLAN is about to search the code, the session shall run `moai index-query` on the project before any atlas query.
- REQ-2: When `index-query` reports `fresh`, the session shall use `atlas-graph`, `atlas-map` or `search` to narrow the files before reading them.
- REQ-3 (REQ-1): IF `index-query` reports `partial` or `unavailable`, THEN the session shall search and read the files without the atlas.
- REQ-4 (REQ-2): IF `index-query` reports `stale`, THEN the session shall use the atlas only for files its `changes` do not name and read the named files directly.
- REQ-5 (REQ-2): IF a `stale` answer reports `branch_switched`, THEN the session shall treat the atlas as unavailable.
- REQ-6: The READ and PLAN routes shall never run `index-build`.
- REQ-7: The session shall verify each atlas edge it states by reading the line its `evidence_refs` cite.
- REQ-8: Each READ answer or PLAN result shall name the `generation_id` and `state` it used or say that it searched without the atlas.
- REQ-9: The READ and PLAN workflows shall each point to `internal/code-atlas` in one line.
- REQ-10: The `moai-research` and `moai-plan` agents shall list `internal/code-atlas` in their `reads:`.
- REQ-11: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-atlas-in-read-plan/spec.md`, tracked by git.

실패 경로: REQ-1 → REQ-3. REQ-2 → REQ-4, REQ-5. REQ-6~8은 지시의 내용이고, 시험은 지시 글과 그 명령을 고정해요.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): 참조가 `index-query`와 그 요청, 네 상태와 `branch_switched` 규칙을
  적고, 그 요청들이 임시 저장소에서 실제로 동작해요: 빌드 전 `unavailable`, 빌드 뒤 `fresh`, 파일을 고치면 `stale`과
  `changes.modified`에 그 파일.
  verify: `cargo test --locked --offline --test code_atlas_reference`
- AC-2 (REQ-6): 읽기 명령이 아틀라스를 만들지 않고(빌드 전 읽기 뒤 `index` 폴더 없음), READ·PLAN workflow에
  `index-build`가 없고, PROJECT가 동의 뒤 제안을 그대로 가져요.
  verify: `cargo test --locked --offline --test code_atlas_reference read_and_plan_point_to_the_instruction_in_one_line_and_project_offers_the_build`
- AC-3 (REQ-7, REQ-8): 참조가 `evidence_refs`의 줄을 읽으라고 하고 `generation_id`·`state` 기록을 요구해요; 그래프
  답에 `evidence_refs`가 있는 `file:src/b.rs → file:src/a.rs` 간선이 있어요.
  verify: `cargo test --locked --offline --test code_atlas_reference the_requests_work_and_the_states_follow_the_tree`;
  `grep -c 'generation_id' skills/moai/references/internal/code-atlas.md` → 1 이상
- AC-4 (REQ-9): verify: `grep -c 'internal/code-atlas' skills/moai/workflows/read.md skills/moai/workflows/plan.md` → 각 1
- AC-5 (REQ-10): verify: `cargo test --locked --offline --test agent_defs_package`
- AC-6 (REQ-11): verify: `git ls-files .moai/specs/SPEC-20261117-atlas-in-read-plan` → 이 파일;
  `moai spec check --all --project .` → exit 0
- AC-7 (REQ-1…REQ-11): 패키지와 이웃 검사가 통과해요.
  verify: `python3 -I tools/source_manifest.py check`; `moai package-inspect skills/moai` → exit 0, 65,536바이트 넘는
  파일 없음; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/skill_references_test.py tests/source_manifest_test.py tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `skills/moai/references/internal/code-atlas.md`(새 파일), `skills/moai/workflows/read.md`,
  `skills/moai/workflows/plan.md`(각 한 줄), `skills/moai/references/source-manifest.json`, `agents/research.md`,
  `agents/plan.md`(`reads:` 한 항목), `tests/agent_defs_package.rs`, `tests/code_atlas_reference.rs`(새 파일),
  `.moai/specs/SPEC-20261117-atlas-in-read-plan/spec.md`
- read-only: `src/index.rs`, `src/atlas_query.rs`, `src/search.rs`, `src/ops/index_query.rs`,
  `docs/design-review/appendix/operations/`, `skills/moai/workflows/project.md`
- forbidden: `AGENTS.md`, `skills/moai/SKILL.md`, `skills/moai/workflows/mode-*.md`, `agents/impl.md`,
  `agents/audit.md`, 다른 에이전트가 고치는 파일, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 측정: 이 저장소를 `git archive`로 임시 복사본(1,620 파일)에 풀고 임시 HOME에서 다섯 명령의 시간과 출력 크기,
   신선도 변화(빌드 전·뒤·파일 수정·되돌림)를 재요.
2. RED: `tests/agent_defs_package.rs`(`reads:` 목록, 내부 참조 목록)와 새 `tests/code_atlas_reference.rs`.
3. GREEN: `internal/code-atlas` 참조, `read.md`·`plan.md` 한 줄씩, 두 에이전트의 `reads:`, manifest.
4. 검사, 커밋 뒤 mutation, Result.

## Risks

- 아틀라스 답은 추론이에요(간선 `unreviewed`, Rust 파일끼리만): 검증 없이 사실로 쓰면 틀린 답이 나와요 → REQ-7.
- 넓은 질의(최상위 폴더 전체, 2홉 그래프)는 수십 KB라 문맥을 많이 써요 → 참조가 좁게 묻는 방법을 적어요.
- worktree마다 상태 폴더가 달라, 주 checkout에서 지은 아틀라스는 worktree에서 `unavailable`일 수 있어요(그때는
  검색으로 돌아가요).
- `moai` 바이너리가 없는 호스트에서는 명령을 못 돌려요 → `unavailable`과 같이 검색으로 돌아가요.

## Decisions

- D1 요구 문장은 영어 GEARS, 본문은 한국어, 제목은 영어(저장소의 다른 SPEC과 같은 형식).
- D2 신선도는 `index-query`의 `state` 하나로 판단해요(이미 있는 기능, 새 플래그 없음). `stale`은 `changes`에 없는
  파일에만 써요: 낡은 파일의 간선만 의심스럽고, 나머지는 그대로 맞아서요. 되돌리기: 참조 1단계의 `stale` 줄을
  "쓰지 않음"으로 바꿔요.
- D3 `branch_switched`면 `unavailable`로 봐요: 가지가 바뀌면 바뀐 파일 목록만으로는 범위를 믿을 수 없어서요.
- D4 기본 질의는 1홉 그래프와 이름 붙인 파일이에요(측정: 1홉 4,390바이트, 2홉 93,243바이트). 되돌리기: 참조
  3단계의 크기 문장을 지워요.
- D5 참조는 `moai-research`와 `moai-plan`의 `reads:`에 넣고 `chains/index.md`는 바꾸지 않았어요: 부모 세션은
  workflow 한 줄로 읽어요.
- D6 시험은 참조의 요청 문자열을 그대로 고정하고 실제로 돌려요: 지시와 CLI가 어긋나면 시험이 빨개져요.
- D7 `Moai-Spec` 트레일러를 넣고 `Moai-Session`은 관문 조각(SDD-6)이 정할 때까지 넣지 않아요.

## Result

(구현 뒤 채워요.)
