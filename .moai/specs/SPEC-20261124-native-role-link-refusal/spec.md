---
id: SPEC-20261124-native-role-link-refusal
title: "Generation refuses to write a native-role file or a skill projection through a linked folder"
status: draft
revision: 3
created: 2026-11-24
---

# Generation refuses to write a native-role file or a skill projection through a linked folder

## Intent

`moai generation-plan` 과 `generation-apply` 는 프로젝트 스킬(`.claude/skills/<id>/SKILL.md`)을 쓸 때 경로의 어느 한 칸이라도 심볼릭 링크(symbolic link)면 `conflict` 로 거절하고 링크를 따라가지 않아요(`src/generation.rs:871` 의 `link_on_path`, apply 쪽 `:1365`). 그런데 같은 파일의 다른 두 분기에는 그 검사가 없어요:

- native-role 에이전트 파일(`.claude/agents/moai/<role>.md`, `:766-796`). 2026-11-24 C-159 감사가 찾은 기존 결함(pre_existing)이에요: `.claude/agents` 나 `.claude/agents/moai` 가 프로젝트 밖 폴더로 걸린 링크면, 계획은 `create`/`merge` 를 내놓고 apply 가 `prj-*.md` 를 그 밖의 폴더에 써요. 하니스 전용 선택 여부와 상관없이요.
- 스킬 투영(skill projection, `.claude/skills/<id>/SKILL.md`, `:798-854`). `.claude` 나 `.claude/skills` 가 프로젝트 밖 빈 폴더로 걸린 링크면, 대상 폴더가 아직 없어 "claimed 아님" 검사(`symlink_metadata` 가 링크를 따라가 실패)도 지나가고, 계획은 `create` 를 내놓고 apply 가 밖에 써요. `.claude/skills/<id>` 자체가 링크일 때만 우연히 `conflict` 가 되고, 그때도 이유는 링크가 아닌 "기준 없는 폴더"예요.

개정 1 은 native-role 만 다뤘어요. 개정 2 는 스킬 투영을 같은 카드에 넣어요(리더 결정, C-453 노트 8415, 2026-11-24): 같은 파일의 같은 구멍이라서, 따로 카드를 열면 두 카드가 `src/generation.rs` 를 같이 고쳐 착지 때 충돌해요. 두 분기 모두 프로젝트 스킬과 같은 거절을 넣어요: 계획은 링크가 있는 경로를 `conflict` 로 표시하고, apply 는 계획 뒤에 생긴 링크도 다시 확인해서 아무것도 더 쓰지 않아요. 근거는 코드 읽기(`src/generation.rs` 의 `link_on_path` 와 세 분기, 2026-11-24 `f5a5c43784` 에서 읽음)와 C-453 카드 본문·기준이에요. 새 공식 문서나 외부 자료는 쓰지 않아요.

개정 3 은 구현 중에 읽은 사실 하나를 고쳐요(2026-11-24, 커밋 `10aef7303d` 에서 실행해 봄): `generation_apply_in` 은 쓰기 전에 계획을 다시 만들어 digest 를 비교해요(`src/generation.rs:1338-1348`). 계획 뒤에 링크가 생기면 이미 계획 쪽 검사(REQ-1, REQ-4)가 그 재계획을 `conflict` 로 바꿔서, apply 는 `stale or altered plan; re-preview required`(`stale_plan`)로 거절하고 아무것도 쓰지 않아요. 쓰기 단계의 링크 재확인(`:1393` 근처)은 apply 가 재계획한 뒤 파일을 쓰기 전 사이에 링크가 생길 때만 닿아요. 그래서 개정 2 의 AC-2·AC-5(메시지에 `symbolic link`)는 CLI 로는 맞출 수 없었고(테스트 두 개가 이 메시지 단언에서 실패해요), 쓰기 단계 검사를 지운 뮤턴트는 CLI 테스트로는 죽지 않아요. 개정 3 은 두 가지를 나눠요: 계획 뒤 링크는 거절(`stale_plan` 이든 링크 `conflict` 든)과 "밖에 아무것도 안 씀"을 CLI 로 보이고(REQ-2, REQ-5), 쓰기 단계의 재확인은 그 검사를 작은 함수로 빼서 단위 테스트로 보여요(REQ-6). 프로젝트 스킬 분기의 같은 재확인에도 테스트가 없었어요(`grep "became a symbolic link"` 가 `src` 한 곳만 찾음); 함수를 같이 쓰므로 그 분기도 같은 테스트가 덮어요.

## Non-goals

- 옛 평면 역할 파일(`.claude/agents/<id>.md`)의 제거 경로: 이미 `link_on_path` 로 막혀 있어요(`:1015`). 더 이상 요청되지 않은 프로젝트 스킬의 제거 경로(`:962`)도 마찬가지예요.
- `conflict` 를 해소하는 방법(링크를 지우거나 옮기기)과 사용자 `.claude/` 설정의 변경: 이 검사는 거절만 해요.
- 새 크레이트, 네트워크, 모델 호출. 스킬 투영의 라이선스 검토(`source_reviewed`)와 렌더링은 바뀌지 않아요.

## Requirements

- REQ-1: When `generation-plan` plans a native-role agent file, the plan shall mark the file `conflict` with a detail that names the symbolic link when any component of its path from the project root, the file included, is a symbolic link.
- REQ-2 (REQ-1): IF a symbolic link appears on the path of a native-role agent file after the plan was made, THEN `generation-apply` shall refuse the plan before it writes through that link.
- REQ-3: When no component of a native-role agent file's or a skill projection's path is a symbolic link, `generation-plan` shall plan that file with the action it chose before this change. · failure: none — a path without a link is the existing path and every neighbour test pins it.
- REQ-4: When `generation-plan` plans a skill projection file, the plan shall mark the file `conflict` with a detail that names the symbolic link when any component of its path from the project root, the file included, is a symbolic link.
- REQ-5 (REQ-4): IF a symbolic link appears on the path of a skill projection file after the plan was made, THEN `generation-apply` shall refuse the plan before it writes through that link.
- REQ-6 (REQ-2, REQ-5): IF a symbolic link appears on the path of a native-role agent file, a skill projection file or a project skill file after `generation-apply` made its own plan and before it writes that file, THEN the write step shall stop with a message that names the file and the link before it writes anything more.

## Acceptance

- AC-1 (REQ-1): in a temporary project with a native role, each of `.claude`, `.claude/agents`, `.claude/agents/moai` linked to a folder outside the project, and the role file itself linked to an outside file, the plan marks `.claude/agents/moai/<role>.md` `conflict` with a detail that holds `symbolic link`, `generation-apply` exits non-zero, the outside folder stays empty and the outside file keeps its bytes · verify: `cargo test --locked --offline --test generation_native_role_link a_linked_agents_path_is_a_conflict_and_nothing_is_written_outside`
- AC-2 (REQ-2): a plan made with no link, then `.claude/agents/moai` replaced by a link to an outside folder, gives a `generation-apply` that exits non-zero with a message that holds `symbolic link` or `stale or altered plan`, and the outside folder stays empty · verify: `cargo test --locked --offline --test generation_native_role_link a_link_made_after_the_plan_stops_the_apply`
- AC-3 (REQ-3): a project with plain `.claude/agents` and `.claude/skills` folders gets the role file and the skill projection `create` in the plan and written by the apply, and the neighbour tests of generation stay as they were · verify: `cargo test --locked --offline --test generation_native_role_link a_plain_agents_and_skills_folder_is_planned_and_written_as_before` and `cargo test --locked --offline --test generation_apply_result_cli`
- AC-4 (REQ-4): in a temporary project with a license-reviewed skill projection, each of `.claude`, `.claude/skills`, `.claude/skills/<id>` linked to a folder outside the project, and the `SKILL.md` itself linked to an outside file, the plan marks `.claude/skills/<id>/SKILL.md` `conflict` with a detail that holds `symbolic link`, `generation-apply` exits non-zero, the outside folder stays empty and the outside file keeps its bytes · verify: `cargo test --locked --offline --test generation_native_role_link a_linked_skills_path_is_a_conflict_for_a_projection`
- AC-5 (REQ-5): a plan made with no link, then `.claude/skills` replaced by a link to an outside folder, gives a `generation-apply` that exits non-zero with a message that holds `symbolic link` or `stale or altered plan`, and the outside folder stays empty · verify: `cargo test --locked --offline --test generation_native_role_link a_link_made_after_the_plan_stops_the_projection_apply`
- AC-6 (REQ-1, REQ-4, REQ-6): `refuse_link_after_plan(root, kind, path)` (the write step's check, one function in `src/generation.rs`) returns an error that names the file, the link and the words `symbolic link` for a linked folder or file above a path of the kind `generation:native-role:`, `generation:skill-projection:` or `generation:project-skill:`, and `Ok` for a path without a link and for any other kind · verify: `cargo test --locked --offline --lib generation::tests::a_link_after_the_plan_stops_the_write_step`
- AC-7 (REQ-1, REQ-4, REQ-6): removing the new check from the native-role plan fails AC-1's test, from the projection plan fails AC-4's test, and dropping the kind `generation:native-role:` or `generation:skill-projection:` from `refuse_link_after_plan` fails AC-6's test; each mutant is applied to `src/generation.rs`, the named test is run, and the file is restored; the card's READY carries one `guard-mutant:` line per new guard test, and one `none | <reason>` line for the call of the function in the write step, which only a link made inside `generation-apply` reaches · verify: `cargo test --locked --offline --test generation_native_role_link` and `cargo test --locked --offline --lib generation::tests`

## Scope

- write: `src/generation.rs`, `tests/generation_native_role_link.rs`, `.moai/specs/SPEC-20261124-native-role-link-refusal/**`
- read-only: `tests/project_skills_generation.rs` (the model of the project-skill link tests), `tests/generation_apply_result_cli.rs` (the license-reviewed projection fixture), `tests/areas.toml`
- forbidden: the real HOME and ledger, `.claude/` of the repository, other cards' reserved files, the network

## Plan

Tests first (RED), in a temporary project built like `tests/project_skills_generation.rs` (the `moai` binary through `crate::common::MOAI_BIN`, never the installed state). The projection tests add the reviewed-source fixture of `tests/generation_apply_result_cli.rs` (`manifests/sources.json` with a reviewed license and a fixture skill folder).

1. RED: `tests/generation_native_role_link.rs` with the five tests of AC-1…AC-5 and the plain-folder test of AC-3. The link cases run the real `generation-plan` and `generation-apply` in a throwaway project whose agents or skills folder is linked outside it (the dangerous path itself, not only unit tests).
2. GREEN, native-role plan (REQ-1): in the native-role loop of `src/generation.rs` (`:766-796`) call `link_on_path(&root, &rel)` before the file is read; with a link, `current` is `None`, the action is `conflict` and `detail` says `{link} is a symbolic link; a native role file is never written through it`, as the project-skill branch does.
3. GREEN, projection plan (REQ-4): in the projection loop (`:798-854`) the same call before `fs::read`, the `conflict` action when a link or the existing `unclaimed` case holds, and the link detail first, so a linked `.claude/skills/<id>` names the link, not the missing baseline.
4. Write step (REQ-6): the re-check that the write step already makes for `generation:project-skill:` (`:1393` near) moves into `fn refuse_link_after_plan(root, kind, path) -> Result<()>` with the kinds `generation:project-skill:`, `generation:native-role:` and `generation:skill-projection:` and the same error text; the write step calls it. This part is already in the tree as an inline check (`10aef7303d`); the move is the refactor after the unit test is RED.
5. Test changes after approval of this revision: `moai tdd reopen --reason` for the file `tests/generation_native_role_link.rs` to loosen only the message of AC-2 and AC-5 to `symbolic link` or `stale or altered plan`, then a new RED commit (`Moai-Phase: red`) for the unit test of AC-6 in a `#[cfg(test)] mod tests` of `src/generation.rs`, recorded with `moai tdd red`, then GREEN by the function above.
6. Mutants (AC-7): remove the native-role plan check, run AC-1's test (must fail), restore; the same for the projection plan check and AC-4's test; drop each of the two kinds from `refuse_link_after_plan`, run AC-6's test (must fail), restore. Record each as `guard-mutant: <test> | <code removed> | exit <n>`, plus `guard-mutant: none | the call in the write step is reached only by a link made inside generation-apply`.
7. Neighbours: `moai check affected <base>..<head>` once at READY, `nextest` for the changed targets, `spec check`.

## Risks

- A project that links `.claude/agents` or `.claude/skills` on purpose (for example to share roles or skills between checkouts) now gets `conflict` instead of a write: that is the intended refusal, the same as for project skills, and the detail names the link so the user can decide.
- No test can drop only the call of `refuse_link_after_plan` in the write step: the digest comparison refuses a link made before `generation-apply` starts, and a link made inside it is a race a test cannot place. The unit test pins the function and the READY says so as a `none` mutant line; an audit reads the one call.
- Both new branches share one function with the project-skill branch; a change to `link_on_path` reaches all three, which the tests of all three kinds pin.
