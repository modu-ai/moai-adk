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
- REQ-4 (REQ-2): IF `index-query` reports `stale`, THEN the session shall drop the deleted files from an atlas answer and add every added, modified or new file that names what it asked about.
- REQ-5 (REQ-2): IF a `stale` answer reports `branch_switched` (HEAD moved since the build), THEN the session shall treat the atlas as unavailable.
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
- AC-8 (REQ-4): 빌드 뒤에 추가한 파일 둘(`src/c.rs`는 `crate::a::alpha`, `src/d.rs`는 묶음 import
  `use crate::{a::alpha, …}`)은 낡은 아틀라스의 의존 목록에 없고, 참조의 `stale` 규칙(그 규칙이 적은 `changes` 키의
  파일을 경로와 모듈 이름 단어 `rg -w`로 검색해 합침)이 둘 다 찾아요. 규칙 문장도 고정돼요.
  verify: `cargo test --locked --offline --test code_atlas_reference a_dependent_added_after_the_build_is_found_by_the_stale_rule`
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
- AC-7 (REQ-1…REQ-11): 패키지와 이웃 검사가 통과해요(AC-8은 아래 AC-2 앞에 있어요).
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
- D2 신선도는 `index-query`의 `state` 하나로 판단해요(이미 있는 기능, 새 플래그 없음). `stale`(감사 B1 뒤 수정):
  아틀라스는 `changes`에 없는 파일을 그대로 설명하지만, 파일 목록 답(특히 의존 목록)은 빠진 것이 있을 수 있어요 —
  바뀐·새 파일이 질문 대상을 새로 쓸 수 있어서요. 그래서 `deleted` 파일을 빼고, `added`·`modified`·`new_files`
  파일을 `rg`로 경로와 import 이름 검색해 합쳐요. 고른 이유: "변경이 있으면 의존 질문에 쓰지 않음"보다 큰 저장소에서
  아틀라스를 계속 쓸 수 있고, 검색 대상이 바뀐 파일뿐이라 짧아요. 되돌리기: 참조 1단계 `stale` 줄을 "쓰지 않음"으로
  바꾸고 AC-8 시험을 지워요. 재감사 N1 뒤: 검색어는 경로와 모듈 이름의 단어 검색(`rg -w x`)이에요 — `crate::x`만
  찾으면 묶음 import(`use crate::{a, x::y}`, 이 저장소 `src/` 17개 파일이 씀)를 놓쳐요. 단어 검색은 넓게 잡으므로
  더한 파일도 4단계대로 그 줄을 읽어 확인해요. 재감사 N2(지운 import가 의존으로 남음)는 보수적이라 그대로 둬요.
- D3 `branch_switched`면 `unavailable`로 봐요: HEAD가 빌드 뒤에 움직였으면(커밋이나 가지 전환, `src/index.rs`가
  HEAD와 `base_commit`이 다를 때 참으로 둠) 바뀐 파일 목록만으로는 범위를 믿을 수 없어서요.
- D4 기본 질의는 1홉 그래프와 이름 붙인 파일이에요(측정: 1홉 4,390바이트, 2홉 93,243바이트; 참조에는 저장소마다
  다른 정확한 배율 대신 "수십 배"라고 적어요). 되돌리기: 참조 3단계의 크기 문장을 지워요.
- D5 참조는 `moai-research`와 `moai-plan`의 `reads:`에 넣고 `chains/index.md`는 바꾸지 않았어요: 부모 세션은
  workflow 한 줄로 읽어요.
- D6 시험은 참조의 요청 문자열을 그대로 고정하고 실제로 돌려요: 지시와 CLI가 어긋나면 시험이 빨개져요.
- D7 `Moai-Spec` 트레일러를 넣고 `Moai-Session`은 관문 조각(SDD-6)이 정할 때까지 넣지 않아요.

## Result

구현 에이전트가 2026-11-17 08:49Z–09:00Z에 채웠어요. SYNC와 AUDIT가 다시 확인해요. macOS arm64(개발 Mac). 설치된
slot이 옛 swap 규칙으로 거절해서(exit 75) cargo는 slot 밖에서 한 번에 하나, `CARGO_BUILD_JOBS=4`,
`CARGO_INCREMENTAL=0`, 빌드 폴더는 이 worktree의 `target/`.

커밋: 기준 `c0b858de8c` 위에 `4252d6ff88`(변경), 그 위에 이 Result.

측정(이 저장소를 `git archive HEAD`로 임시 git 저장소에 풀어서, 1,620 파일, 41 MB, 임시 HOME, debug 바이너리):

| 단계 | 명령 | exit | 시간 | stdout | 상태 |
|---|---|---|---|---|---|
| 빌드 전 | `index-query` | 0 | 0.67 s | 267 B | `unavailable` ("no generation has been built") |
| 첫 빌드 | `index-build` | 0 | 6.78 s | 406 B | `complete` |
| 빌드 뒤 | `index-query` | 0 | 0.47 s | 393 B | `fresh` |
| 다시 빌드(변화 없음) | `index-build` | 0 | 2.39 s | 408 B | `complete` |
| 파일 하나 | `atlas-map` paths `src/heavy_slot.rs` | 0 | 1.30 s | 618 B | `fresh` |
| 폴더 | `atlas-map` prefixes `src`(기본 200개) | 0 | 1.36 s | 71,505 B | `fresh` |
| 폴더 | `atlas-map` prefixes `src`, `max_files` 2000 | 0 | 1.43 s | 114,088 B | `fresh` |
| 1홉 | `atlas-graph` `file:src/heavy_slot.rs` | 0 | 1.38 s | 4,390 B | `fresh` |
| 2홉 | `atlas-graph` 같은 시작, `max_hops` 2 | 0 | 1.34 s | 93,243 B | `fresh` |
| 문자열 | `search` literals `CRITICAL_PRESSURE_LEVEL` | 0 | 2.45 s | 2,704 B | 결과마다 `fresh` |
| 파일 수정 + 새 파일 | `index-query` | 0 | 0.39 s | 437 B | `stale`, `modified` [`src/heavy_slot.rs`], `new_files` [`c11-untracked-probe.txt`] |
| 같은 때 | `atlas-map` | 0 | 0.97 s | 618 B | `stale` |
| 되돌린 뒤 | `index-query` | 0 | 0.39 s | 393 B | `fresh` |

색인 크기: `.moai-cli/index` 5.8 MB. 결론: 신선도 판정(`state`, `changes`, `branch_switched`)과 각 답의 `state`가 이미
있어서 새 CLI 기능은 필요 없었어요.

- AC-1: 충족 — `code_atlas_reference` exit 0, 3 passed. RED(지시 연결 전): `read.md` 줄 수 0.
- AC-2: 충족 — 같은 실행(빌드 전 읽기 세 번 뒤 `index` 폴더 없음).
- AC-3: 충족 — 같은 실행(`file:src/b.rs` → `file:src/a.rs` 간선, `evidence_refs` 있음); `generation_id` 문장 있음.
- AC-4: 충족 — `read.md` 1줄, `plan.md` 1줄.
- AC-5: 충족 — `agent_defs_package` exit 0, 17 passed. RED: `reads_lists_match_the_design_and_render_on_every_host`·
  `internal_references_say_when_to_read_them` 실패(exit 101).
- AC-6: 충족 — `git ls-files` → 이 파일; `moai spec check --all --project .` exit 0, SPEC 3개 오류 0·경고 0.
- AC-7: 충족 — manifest `write`(read.md·plan.md·code-atlas.md 바뀜) 뒤 `check` exit 0; `package-inspect skills/moai`
  exit 0, 65,536바이트 넘는 파일 없음(`read.md` 3,235·`plan.md` 3,119·`code-atlas.md` 2,415바이트);
  `--bin moai` 27 passed; `cargo fmt --check` 0(새 시험 파일을 한 번 `rustfmt`한 뒤); clippy 0; Python 85개 OK,
  커밋 뒤 `skill_references_test` 2개 OK. 패키지를 읽는 시험도 모두 0: `ac_command_lint` 5, `mode_workflow_steps` 42,
  `core_rules_2876` 4, `update_route` 8, `question_bank_hosts` 1, `install_cli` 7, `cli_name_moai` 5,
  `release_suite` 4, `go_state_codex_cli` 37, `product_cli` 75, `review_package_cli` 12, `update_tux_golden` 14,
  `home_isolation_scan` 16.

mutation(커밋 뒤, 셋 다 되돌린 뒤 `git status` 비어 있음): `read.md` 줄 삭제 → `code_atlas_reference` 1개 실패;
참조의 그래프 요청 `start` → `starts` → 1개 실패; `agents/research.md`의 `reads:`에서 항목 삭제 →
`agent_defs_package` 1개 실패.

실행 안 함: 전체 `cargo test`, Python 전체, 실제 호스트(Claude Code·Codex)에서 READ·PLAN 세션을 돌려 지시를 따르는지
보는 일(호스트 E2E), Linux·Windows.

### 감사 수리 (2026-11-17 09:08Z–09:12Z)

독립 감사(head `4f6a90ee57`, 09:01:54Z–09:07:32Z)가 FAIL이었어요. B1: `stale`에서 `changes`에 없는 파일에만
아틀라스를 쓰라는 규칙은, 바뀐·새 파일에서 나가는 간선을 놓쳐요(감사 탐침: 새 파일이 `crate::worktree_gc`를 썼는데
낡은 1홉 답에는 옛 의존 4개뿐). N1: `branch_switched`는 HEAD가 움직였다는 뜻(커밋 포함). N2: "약 21배" 대신 일반 표현.

커밋: `4f6a90ee57` 위에 `664eccae0f`(수리), 그 위에 이 Result.

- AC-8 (새): 충족 — RED(수리 전 참조로 새 시험 실행): `the_instruction_names_…` 실패(`stale` 규칙에 "An answer that
  lists files … is incomplete" 없음), `a_dependent_added_after_the_build_is_found_by_the_stale_rule` 실패(규칙에 검색
  단계 없음). 같은 실행에서 낡은 아틀라스의 의존 목록은 `src/b.rs`를 담고 `src/c.rs`는 빠져 감사 B1이 재현됐어요.
  GREEN: `code_atlas_reference` exit 0, 4 passed.
- mutation(수리 커밋 뒤): 감사의 M3("stale이면 모든 파일에 사용")로 바꾸면 2개 실패; 검색 키를 `new_files`만으로
  줄이면 1개 실패. 둘 다 되돌린 뒤 `git status` 비어 있음.
- 다시 돌린 검사: `--bin moai` 27, `agent_defs_package` 17, `code_atlas_reference` 4, `ac_command_lint` 5,
  `mode_workflow_steps` 42, `home_isolation_scan` 16, `product_cli` 75 — 모두 exit 0; `cargo fmt --check` 0(새 시험
  `rustfmt` 뒤); clippy 0; Python 85개 OK; `moai spec check --all` exit 0(SPEC 3개 오류 0·경고 0); manifest
  `write`(code-atlas.md 바뀜) 뒤 `check` 0; `package-inspect skills/moai` exit 0, 65,536바이트 넘는 파일 0.
