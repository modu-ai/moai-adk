---
id: SPEC-20261120-project-harness-research
title: "project asks for a researched harness: agents and skills proposed from the detected stack and current sources, with no fixed count"
status: draft
revision: 2
created: 2026-11-20
---

# project asks for a researched harness: agents and skills proposed from the detected stack and current sources, with no fixed count

## Intent

Step 6 of the project route offers at most three project agents and then project skills "the same way", built only from the interview answers. A project's useful harness is not three: it depends on the stack. This change makes step 6 study the detected project (languages, frameworks, test and build tools, CI, folders), search the web for current practice and well-regarded sources for that stack, and propose the agents and skills worth creating, each with its reason and the sources it rests on (URL and read time). The user picks what to create from one selection with the recommended ones first, and nothing is created without the pick. The proposals stay generic and portable and reuse the existing `generation-plan` and `generation-apply` contract. The files that carry the text (`project.md`, the manifest) are held by the card that rewrites the project and update texts, so the text change follows that card; the SPEC, the research notes of this folder and the request limit of `generation-plan` do not wait.

### 개정 2 (2026-11-24): 에이전트 하나만 다시 만들 수 있게 해요

카드 기준 4 가 2026-11-24 에 더해졌어요. `generation-plan` 은 프로젝트 에이전트 하나의 목적만 바꾸는 요청에도 `CLAUDE.md`(conflict), `REVIEW.md`(merge), `.gitignore`(conflict)와 워크플로 파일을 함께 계획해요(카드 측정 2026-11-24). `generation-apply` 는 conflict 가 하나라도 있으면 아무것도 적용하지 않아서, 사용자가 에이전트 하나만 고칠 수 없어요. 이 개정은 그 길을 열어요: 요청이 고르는 `harness-only` 선택(이미 있는 `picks` 목록의 새 항목)이 있으면 계획과 적용은 요청에 적힌 프로젝트 에이전트·스킬과 매니페스트만 건드려요. 6단계 글도 같은 선택을 쓰라고 말해요. 개정 1 에서 Codex CLI 를 이 SPEC 의 대상으로 적었던 곳은 Codex 지원이 2026-11-21 에 빠졌으므로(SPEC-20261121-remove-codex-support) Claude Code 만 말하도록 고쳐요.

## Requirements

- REQ-1: The project route shall tell the agent to study the detected project, its languages, frameworks, test and build tools, CI and folders, before it proposes any agent or skill.
- REQ-2: WHEN web search is available, the project route shall tell the agent to search for current practice and well-regarded sources for the detected stack and to name each source, with its URL and read time, next to the proposal it supports. · failure: none
- REQ-3: The project route shall propose only what the stack and the sources justify, naming no count limit of project agents or project skills.
- REQ-4: Each proposal shall name its purpose, the tools or trigger it needs, its scope, the reason it is proposed and its sources.
- REQ-5: The project route shall offer the proposals as one several-of selection with the recommended ones first, creating nothing that the user did not pick.
- REQ-6: The proposals shall stay generic and portable, naming no internal tool of this project, no secret and no absolute home path.
- REQ-7 (REQ-2): IF web search is unavailable, THEN the project route shall tell the agent to propose from the detected facts alone and to say that no source was read.
- REQ-8: The `generation-plan` operation shall accept up to sixteen agents and sixteen skills in one request.
- REQ-9 (REQ-8): IF a request has more agents or skills than the limit or a name that is not safe, THEN `generation-plan` shall refuse it with the name of the field and the limit, and never with a value of the request.
- REQ-10: WHEN a request carries the pick `harness-only`, `generation-plan` shall list no file but the requested project agents and project skills, the earlier flat agent files they replace and the generation manifest. · failure: none
- REQ-11 (REQ-10): WHEN a request carries the pick `harness-only`, `generation-apply` shall leave the bytes of `CLAUDE.md`, `REVIEW.md`, `.gitignore`, the personal file, every workflow file and every project skill the request does not name, and the manifest's `workflow_id`, `interview_revision` and `profile_sha256`, as they were. · failure: none
- REQ-12 (REQ-10): Step 6 of the project route shall tell the agent to put the pick `harness-only` into every `generation-plan` request it makes for the picked agents and skills. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): step 6 of `skills/moai/workflows/project.md` carries the study, the web search with a source per proposal, the missing-web fallback, the fields of a proposal, the several-of selection and the portability rule, and no longer says "at most three"; the scan tests and `skill_references` pass · verify: `cargo test --locked --offline --test project_harness_step`; `cargo test --locked --offline --test skill_text_citations --test product_cli`; `python3 -m unittest tests.skill_references_test tests.source_manifest_test`
- AC-2 (REQ-8, REQ-9): `generation-plan` accepts sixteen agents and sixteen skills, and refuses seventeen of either, and an unsafe name, with the field named and no value shown · verify: `cargo test --locked --offline --test generation_request_limits`
- AC-4 (REQ-10, REQ-11): in a temporary project whose `CLAUDE.md` managed section, `REVIEW.md` and `.gitignore` block were edited by hand after generation, a request that renames one `prj-` agent's purpose and carries `harness-only` plans only that agent file and the manifest, with no `conflict`; the apply writes that file, leaves the bytes of every other tracked file and the three manifest linkage keys as they were, and the same request without the pick still plans the three conflicts; a second project skill recorded earlier stays when the request names one · verify: `cargo test --locked --offline --test project_harness_pick`
- AC-5 (REQ-12): step 6 names the pick `harness-only` and the request it goes into · verify: `cargo test --locked --offline --test project_harness_step harness_step_asks_for_the_harness_only_pick`
- AC-3 (REQ-1…REQ-7): a dry run of the step on two fixture projects, a Rust command-line tool and a TypeScript web application, gives different proposals, each with at least one source line (URL and read time); the run and its sources are kept in `research.md` of this folder · verify: manual: read `research.md` of this SPEC and compare the two proposal lists

## Scope

- write: `.moai/specs/SPEC-20261120-project-harness-research/**`
- write (after the card that rewrites the project texts has returned them): `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/project_harness_step.rs`
- write (widened when the file is free): `src/generation.rs`, `tests/generation_request_limits.rs`
- write (revision 2, by line lease because other cards edit it): `src/agents.rs` (the one new pick id and its skip rule), `tests/project_harness_pick.rs` (new; named with the area word `project` so `tests/areas.toml` maps it), `tests/product_cli.rs` (one line: the step 6 heading pin) and `tools/repo_scans.json` (the reads scan that names `tests/project_harness_step.rs`), the last two by line lease

## Decisions

- D-1: The limit of sixteen agents and sixteen skills stays as a safety bound, not a design count: it exists so one request cannot write an unbounded number of files, it is five times the old advice, and the refusal names it. "No fixed count" is read as no count in the advice the user sees; the leader may lift the bound by one constant.
- D-2: The rebuild record that lets `/moai update` recreate project agents and skills is another card's work (the generation record); this SPEC reuses the existing `generation-plan` and `generation-apply` contract and adds no record.
- D-3: The dry run of AC-3 is a manual criterion: the proposals are the agent's own output from a web search, which a test cannot repeat; the text and the request limit are what the tests pin. The run is recorded in `research.md` with its sources so a reviewer can read it.
- D-4 (revision 2): The official document of the one host that has an agent file, Claude Code subagents and skills, is the source the proposal format stays inside (`research.md`). Codex CLI left the product on 2026-11-21, so step 6 names no other host's file.
- D-5 (revision 2): The way to change one agent alone is a new pick id `harness-only` in the existing `picks` list of the request, not a new request field. Options: (a) a pick id (chosen: the list already travels through the plan, the digest and the apply, an empty list digests as before, and `skip` is an action the agents engine and its receipt already know), (b) a new boolean request field (the request schema and every fixture of it change), (c) infer the mode from a request that has no `picks` (a silent change for every caller today). Undo: remove the id from `PICKS` and the branch that reads it.
- D-6 (revision 2): With the pick, the plan lists no skipped file at all, so the user is shown only what will change; the agents engine still receives the pick and plans every file it owns as `skip`, which is how the apply leaves them alone. A project skill the request does not name is not planned `remove` under the pick, because a request for one agent says nothing about the others.
- D-7 (revision 2): This revision is drafted by the lane. Its approval is a record the leader writes (the user's delegation of 2026-11-23 covers a reversible in-repo pick like this one); the lane does not record it.

## Plan

1. This card, now: this SPEC, `research.md` (the official formats read, the facts a proposal may rely on) and the fixtures' proposals.
2. Tests first: `tests/generation_request_limits.rs` (red: the refusal names no field today) and `tests/project_harness_step.rs` (red against "at most three"); then `src/generation.rs` (the field named in the refusal), then the text of step 6 after the project-text card has returned `project.md` and the manifest.
3. Whole-area run, clippy, rustfmt, spec check, audit, merge.
4. Revision 2: RED first (`tests/project_harness_pick.rs`, and the step-6 phrase test), then the `harness-only` id and its `skip` rule in `src/agents.rs`, then `src/generation.rs` (list only the harness files, skip the workflow files and the no-longer-requested removal, keep the linkage keys), then the step-6 sentence, then the source manifest.

## Result

- AC-1: met — `cargo test --locked --offline --test project_harness_step`: 5 passed (they were red against the old step 6): step 6 studies the detected project, searches for current practice with a URL and read time, has no fixed count and up to sixteen agents, names purpose, scope, reason and sources, offers a several-of pick and creates nothing unpicked, stays portable, falls back to the detected facts when web search is unavailable, and keeps the `native_roles`, `project_skills`, `prj-<name>` and `generation-apply` contract. The 18 test files that read `project.md` pass except the generated-file checks named below.
- AC-2: met — `cargo test --locked --offline --test generation_request_limits` passed (5 tests): sixteen agents and sixteen skills are planned, seventeen of either are refused with the field and the limit, and a name without the `prj-` prefix, a bad host, a tool name with a space and an empty purpose are refused with the field and its index (`native_roles[1].role_id`, `project_skills[0].purpose`) and with no value of the request. Four of the five tests failed on the unchanged code (4 failed, 1 passed) and pass now; the change is in `validate` of `src/generation.rs`.
- AC-3: met by reading — `research.md` holds the dry run on a Rust command-line tool and a TypeScript web application: three proposals each, no name shared, every proposal with a source line (URL and read time 2026-11-20T13:48Z). The sources are search excerpts, not whole pages, and the proposals were made by following the written step by hand; `research.md` says so.
- Area run: `--lib`, `--bins` and 21 reader targets of `generation-plan` (agents_choices_step_cli, agents_apply_linkage, cli_project, empty_parent_prune_cli, generation_own_reference, generation_apply_result_cli, generation_manifest_single_write, generation_request_limits, generation_ref_slash, generation_receipt_order, launcher_auto_init, plan_input_limit_cli, placeholder_profile_guard, project_update_plan_cli, product_cli, project_skills_generation, removed_skill_sources and the four atlas fixtures readers): 24 `test result: ok` lines, 1942 tests passed, 0 failed; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0; rustfmt of the two Rust files clean.
- 개정 2 (2026-11-24, head 는 READY 노트가 적어요): 옛 가지 `c4579f2b38` 을 NeXT `24dceb4e2b` 위에 합쳤어요(`79c6beaa55`). 충돌은 `project.md` 6단계와 `src/generation.rs` 검증 두 곳이었고 NeXT 쪽 문장(CLAUDE.md 5단계, 6a 단계)을 지켰어요. Codex 는 새 6단계 글과 호스트 검사에서 뺐고, 6단계 시험은 6a 앞에서 끊어요. 소스 매니페스트는 `python3 -I tools/source_manifest.py write` 로 다시 만들었고 `check` 가 종료 0 이에요.
- AC-1: met — `cargo test --locked --offline --test project_harness_step`: 6 passed (문구 시험 5개와 `harness-only` 시험 1개), 합친 트리에서 다시 읽었어요.
- AC-2: met — `cargo test --locked --offline --test generation_request_limits`: 5 passed, 합친 트리에서.
- AC-3: met by reading — 변화 없음(`research.md`).
- AC-4: met — `cargo test --locked --offline --test project_harness_pick`: 4 passed. 시험 먼저 돌렸을 때 3개가 `picks holds an id this tool does not know` 등으로 실패했고(RED), 고정 장치는 손으로 고친 CLAUDE.md·REVIEW.md·.gitignore 와 바뀐 프로필로 선택 없는 요청이 conflict 를 내는 것까지 재현해요. 선택이 있으면 계획은 에이전트 파일, 이름 붙인 스킬, 매니페스트만 나열하고 conflict 가 없어요. 적용은 세 파일과 워크플로 파일의 바이트, 이름 없는 스킬 `prj-other`, 매니페스트의 `workflow_id`·`interview_revision`·`profile_sha256` 을 그대로 두어요.
- AC-5: met — `harness_step_asks_for_the_harness_only_pick` 가 6단계의 문장 둘을 읽어요.
- 가드 뮤턴트(코드를 뺀 사본에서 시험이 실패하는지, 각 `cargo test` exit 101): M1 `is_harness_only` 를 늘 false(2 실패), M2 프로필 digest 를 되돌리는 줄 제거(1), M3 `src/agents.rs` 의 skip 규칙 제거(1), M4 이름 없는 스킬 제거 계획을 되살림(2), M5 6단계 문장 제거(1), M6 엔진 파일을 계획에 나열(1), M7 워크플로 파일을 계획에 나열(1). 일곱 모두 복원했고 `git diff` 가 비었어요.
- 되돌리는 법: `PICKS` 에서 `harness-only` 와 그 분기를 지워요(D-5).
