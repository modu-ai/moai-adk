---
id: SPEC-20261124-harness-step-label
title: "generation 단계를 사용자에게는 Project harness 로 보여 주고 식별자는 그대로 두요; 프로젝트 하네스 연구 감사의 뒤따르는 일 네 가지"
status: approved
revision: 1
created: 2026-11-24
parent: SPEC-20261120-project-harness-research
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 (2026-11-24T10:06:12Z, 사용자 위임 2026-11-23T13:57Z, ref note:255:50bdb6de88de495b1efbbfc1ad040462, 승인 요청 노트 8437 에 답함): 판 1 승인(리더가 spec show 로 같은 digest 를 읽음), D-2 는 CLAUDE.md, REQ-6 은 두 곳 다, D-4 는 동작 그대로; 쓰기 목록 열한 개 가운데 아홉 개를 카드 범위에 넣고 src/generation.rs 와 tools/repo_scans.json 은 줄 임대로 받음; 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:452fd94a8f043f2c5240c9bbaa772fbeb41fb668efe229c6660ab3472dab1a05"
---

# generation 단계를 사용자에게는 Project harness 로 보여 주고 식별자는 그대로 두요; 프로젝트 하네스 연구 감사의 뒤따르는 일 네 가지

## Intent

사용자 결정(2026-11-21, 선택 "맞다, 표시 이름만 변경"): `moai project-update-plan` 의 `generation` 단계를 사람이 읽는
자리에서는 한국어 화면에는 '프로젝트 맞춤 하네스', 영어 화면에는 'Project harness' 로 불러요. 호환을 위해 안쪽 이름은
그대로예요: 단계 id `generation`, 명령 `generation-plan` 과 `generation-apply`, 파일 `manifests/generation.json`, JSON
필드. 이 카드에는 프로젝트 하네스 연구(SPEC-20261120-project-harness-research) 감사의 뒤따르는 일 네 가지도
들어와요(리더 결정 2026-11-24T08:09:10Z: 감사의 비차단 N1·N3·N5).

2026-11-21 의 초안(SPEC-20261121-harness-step-label, 가지 spec/harness-step-label-20261121)은 통합 가지에 들어오지
않았고 그 뒤로 저장소가 많이 바뀌었어요(CLAUDE.md, `/moai project` 인터뷰, 하네스 제안 단계, Rust 전용 테스트). 이
판은 오늘 나무에서 요구를 다시 읽은 거예요. 옛 가지에서 가져오는 것은 없어요(D-6).

읽어서 확인한 사실(코드와 글, 통합 가지 `596b525dbc`, 2026-11-24):

- 계획 JSON 에는 표시 이름 칸이 없어요. 단계의 `name` 은 `generation` 이고, 그 계획을 사람에게 보여 주는 화면은
  Rust 가 그리는 게 아니에요. `/moai update` 와 `/moai project` 워크플로 글이 모델에게 계획을 읽어 보여 주라고
  해요(`src/ops/project_update_plan.rs` 는 23줄짜리 감싸기이고 화면을 그리지 않아요). `src/ui_text.rs` 는 영어와
  한국어 칸만 있고 이 단계 이름은 없어요. 일본어·중국어 표 판은 저장소에 없어요.
- 사용자가 단계 이름을 만나는 자리 여덟 곳:
  1. `src/project_update.rs` 의 `INSTALL_REASON` — "profile, generation (AGENTS.md managed section and generated
     files), Claude Code settings, …". 설치 단계 전부가 `blocked` 일 때 이유로 보여요.
  2. 같은 파일의 되돌리기 거절 노트 둘 — "<agents_receipt of the generation step>", "<receipt of the generation
     step>", "refused while the generation step's AGENTS.md is in place". `tests/update_end_to_end_cli.rs` 가
     앞의 둘을 글자 그대로 고정해요.
  3. `skills/moai/workflows/update.md` 1단계 — 각 단계의 미리보기를 "the profile `delta`; the generation `files`;
     the settings `action` …" 로 열거해요.
  4. `skills/moai/workflows/project.md` 5단계 — "use its `generation` step", "Ask about the `generation` step only
     when it reads `ready` in that run". 뒤의 문장은 `tests/project_preview_order.rs` 가 글자 그대로 고정해요.
  5. `project.md` 6단계의 작은 항목 제목 — "Generation (from SETUP): …".
  6. `project.md` 의 최종 보고 문단 — 표의 행을 "(…, hosts, project agents, the official plugins, …)" 로 열거해요.
     `tests/project_applied_table.rs` 는 열과 상태 낱말만 고정하고 행 이름은 고정하지 않아요.
- 명령 도움말(`src/help.rs`)은 명령 이름 `generation-plan`, `generation-apply` 만 말하고 단계를 부르지 않아요.
- 뒤따르는 일 네 가지를 읽은 것:
  - `tests/project_harness_pick.rs` 의 선택 없는 시험(`generated_then_edited` 와 그 아래 시험)은 세 추적 파일
    (CLAUDE.md, REVIEW.md, .gitignore)을 손으로 고치지만 "conflict 가 하나 있다" 만 단언해요. 감사는 그 계획의
    conflict 가 REVIEW.md 하나뿐이라고 적었어요(동작은 맞아요).
  - `src/generation.rs` 의 `record_generation`: `harness-only` 선택이 있으면 앞서 적힌 `profile_sha256` 가 있을 때만
    되돌리고 돌아와요. 생성된 적 없는 프로젝트에는 앞서 적힌 값이 없어서 에이전트 엔진이 넣은 현재 프로필의 digest 가
    그대로 남아요(`src/agents.rs` 의 `profile_sha256: digest(profile_bytes)`). 주석은 "the profile digest goes back
    to the one the last full generation recorded" 라고만 적어요.
  - `research.md`(SPEC-20261120-project-harness-research) 의 Rust 픽스처 표 셋째 행은 출처를 "the same testing pages
    as above" 라고만 적고 URL 이 없어요.
  - `tests/generation_request_limits.rs` 의 픽스처 프로필이 `"hosts": ["claude", "codex"]` 를 써요. Codex 는
    2026-11-21 에 제품에서 빠졌어요.

## Interview

- Q1: 표시 이름을 계획 JSON 의 새 칸에 넣을까요? → A: 아니요(D-1). JSON 은 기계가 읽는 출력이고 사람에게는 모델이
  워크플로 글에 적힌 이름으로 보여 줘요. 한국어는 JSON 에 들어가지 않아요.
- Q2: 일본어·중국어 이름은요? → A: 표 판이 없어서 적지 않아요. 문장이 모델에게 다른 대화 언어로는 이름을 옮기라고
  해요(D-7).
- Q3: 설치 이유의 괄호가 말하는 파일은요? → A: 단계가 오늘 쓰는 `CLAUDE.md` 의 관리 구간이에요. 지금 글은 옛 이름
  `AGENTS.md` 를 말해요(D-2).
- Q4: 최종 보고 표의 행 이름은 왜 바꾸나요? → A: 카드 본문이 최종 설정 표를 사용자가 단계를 만나는 자리로
  적었고, 그 행이 오늘 "project agents" 로 적혀 있어요. 하네스의 행에 같은 표시 이름을 써요.
- Q5: 5단계의 문장은 바꾸나요? → A: 아니요. `tests/project_preview_order.rs` 가 고정한 문장들은 글자 그대로
  남기고, 이름 문장 한 개를 그 앞에 더해요(D-3).
- Q6: 생성된 적 없는 프로젝트의 기록은 고치나요? → A: 아니요. 주석과 시험이 오늘 동작을 사실대로 적어요. 동작을
  바꾸는 일은 이 SPEC 밖이에요(D-4).

## Non-goals

- 단계 id, 명령 이름, 매니페스트 경로, JSON 필드 이름, 매니페스트 파일을 말하는 오류 글("generation manifest
  unreadable …") 바꾸기.
- 계획 JSON 에 표시 이름 칸 더하기, `src/ui_text.rs` 에 언어별 이름 표 더하기.
- 일본어·중국어 이름. `Go-generation`, `earlier generation`(도구의 세대)이나 카드 예약 `generation`(세대 번호)이라는
  낱말: 다른 뜻이라 건드리지 않아요.
- 되돌리기 거절 노트의 파일 이름 `AGENTS.md`(병합 단계가 옮기는 옛 파일)와 `src/project_update.rs` 의 다른
  `AGENTS.md` 글: 이 SPEC 은 단계 이름만 바꿔요.
- 생성된 적 없는 프로젝트가 `harness-only` 적용 뒤 무엇을 기록하는지 바꾸기.
- 네이티브 역할 파일 쓰기가 링크된 `.claude/agents/moai` 를 따라가는 문제(감사의 N2): 다른 카드예요.

## Requirements

- REQ-1: The install phase reason of `moai project-update-plan` shall name the generation step `project harness` and no other name.
- REQ-2: The step id `generation` and the commands `generation-plan` and `generation-apply` and the path `manifests/generation.json` and every JSON field name of the plan shall stay unchanged.
- REQ-3: The `/moai update` workflow text and the `/moai project` workflow text shall each tell the model to call the step `Project harness` in English and `프로젝트 맞춤 하네스` in Korean.
- REQ-4: Each of the two workflow texts shall tell the model to translate the display name for any other conversation language and to keep the id `generation` for the commands and the JSON and never to say the id to the user.
- REQ-5: The `/moai update` workflow text shall name the generation step `project harness` where it lists what each step's preview holds.
- REQ-6: The `/moai project` workflow text shall name the generation step `Project harness` in the step that offers the project harness and in the final report table.
- REQ-7: The refusal notes of `moai project-update-plan` that name the generation step as a step shall call it `project harness`.
- REQ-8: The no-pick test of `tests/project_harness_pick.rs` shall assert that the paths planned `conflict` without the pick are exactly `CLAUDE.md` and `REVIEW.md` and `.gitignore`.
- REQ-9: The comment of the `harness-only` branch of `record_generation` in `src/generation.rs` shall say what a never-generated project records in the manifest under the pick.
- REQ-10: A test of `tests/project_harness_pick.rs` shall pin the three linkage keys of the manifest after a never-generated project applies a request with the pick `harness-only`.
- REQ-11: The third proposal row of the Rust fixture in `research.md` of `SPEC-20261120-project-harness-research` shall give the URL of each source it rests on.
- REQ-12: The fixture profile of `tests/generation_request_limits.rs` shall name no retired host.

## Acceptance

- AC-1 (REQ-1): `INSTALL_REASON` 이 "profile, project harness (CLAUDE.md managed section and generated files), …" 로
  읽히고 `generation` 이라는 낱말이 없고 나머지 설치 단계들이 오늘과 같은 순서로 남아요.
  · verify: `cargo nextest run --offline --locked --test project_harness_label -E 'test(install_reason)'`
- AC-2 (REQ-2): `src/project_update.rs` 가 계획 코드에서 단계를 여전히 `generation` 으로 적고 `generation-apply` 를
  찍고 `manifests/generation.json` 을 읽고, `src/help.rs` 가 두 명령을 여전히 적어요.
  · verify: `cargo nextest run --offline --locked --test project_harness_label -E 'test(identifiers_stay)'`
- AC-3 (REQ-3, REQ-4): 워크플로 글 두 파일이 각각 이름 문장 하나를 다섯 조각 통째로 품어요 — 영어 이름, 한국어 이름,
  다른 대화 언어로 옮기는 조각, id 를 명령과 JSON 용으로 두는 조각, id 를 사용자에게 말하지 않는 조각.
  · verify: `cargo nextest run --offline --locked --test project_harness_label -E 'test(workflow_text)'`
- AC-4 (REQ-5, REQ-6): `update.md` 1단계가 각 단계의 미리보기를 열거할 때 그 단계를 `the project harness` 로
  불러요. `project.md` 6단계의 작은 항목 제목과 최종 보고 문단의 행이 `Project harness` 나 `the project harness` 를
  말해요. `tests/project_preview_order.rs` 와 `tests/project_applied_table.rs` 가 고정한 문장은 그대로예요.
  · verify: `cargo nextest run --offline --locked --test project_harness_label --test project_preview_order --test project_applied_table -E 'test(workflow_text_names_the_step_where_it_lists_it) + test(step_5_asks_about_generation_only_when_it_reads_ready) + test(the_final_report_has_an_applied_settings_table)'`
- AC-5 (REQ-7): 되돌리기 거절 노트 둘이 `project harness step` 이라고 적고, `src/project_update.rs` 의 글자열 리터럴
  어디에도 `generation step` 이 없고, `tests/update_end_to_end_cli.rs` 의 두 고정이 새 글을 읽어요.
  · verify: `cargo nextest run --offline --locked --test project_harness_label --test update_end_to_end_cli -E 'test(refusal_notes) + test(the_printed_undo_order_takes_the_chain_back_to_the_start_bytes)'`
- AC-6 (REQ-8): 선택 없는 시험이 conflict 경로 셋을 경로 이름으로 단언해요. 먼저 오늘의 고정 장치로 돌리면 REVIEW.md
  하나만 conflict 라서 빨개요(RED). 고정 장치가 세 파일을 모두 conflict 로 만들면 통과하고, 같은 파일의 선택 있는
  시험 넷은 그대로 통과해요.
  · verify: `cargo nextest run --offline --locked --test project_harness_pick`
- AC-7 (REQ-9, REQ-10): 주석이 생성된 적 없는 프로젝트의 기록을 적어요 — 앞서 적힌 값이 없어서 현재 프로필의
  `profile_sha256` 가 남고 workflow id 와 interview revision 은 엔진이 빈 앞서 기록에서 가져오는 값이에요. 새
  시험이 그 세 칸을 그 값으로 고정해요.
  · verify: `cargo nextest run --offline --locked --test project_harness_pick -E 'test(a_never_generated_project)'`
- AC-8 (REQ-11, REQ-12): `research.md` 의 Rust 픽스처 표 셋째 행이 출처마다 URL 을 적고, 픽스처 프로필에 `codex` 가
  없고, 한도 시험이 통과해요.
  · verify: `grep -n "the same testing pages as above" .moai/specs/SPEC-20261120-project-harness-research/research.md`; `grep -c codex tests/generation_request_limits.rs`; `cargo nextest run --offline --locked --test generation_request_limits`
- AC-9 (REQ-1..REQ-12): 이웃 — 소스 매니페스트가 바뀐 스킬 파일 두 개와 맞고, 워크플로 글 두 파일을 읽는 시험이
  통과하고, SPEC 검사가 오류 0 이에요.
  · verify: `python3 -I tools/source_manifest.py check`; `moai spec check SPEC-20261124-harness-step-label`; `moai check affected <base>..<head>`

여덟 번째 기준의 grep 둘은 출력이 없을 때(0을 찍을 때) exit 1 로 끝나요. 그것이 통과예요.

## Scope

- write: `.moai/specs/SPEC-20261124-harness-step-label/spec.md`, `tests/project_harness_label.rs`, `src/project_update.rs`, `skills/moai/workflows/update.md`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/update_end_to_end_cli.rs`, `tests/project_harness_pick.rs`, `src/generation.rs`, `.moai/specs/SPEC-20261120-project-harness-research/research.md`, `tests/generation_request_limits.rs`, `tools/repo_scans.json`
- read-only: `src/help.rs`, `src/agents.rs`, `src/ops/project_update_plan.rs`, `src/ui_text.rs`, `tests/project_preview_order.rs`, `tests/project_applied_table.rs`, `tests/project_harness_step.rs`, `tests/product_cli.rs`, `tools/source_manifest.py`
- forbidden: 단계 id, 명령 이름, 매니페스트 경로, JSON 필드 이름(`generation`, `generation-plan`, `generation-apply`, `manifests/generation.json`)을 바꾸기, 실제 HOME, push 와 PR, 다른 카드의 가지

공유 파일 `tools/repo_scans.json` 은 줄 임대로 써요(새 테스트 파일을 reads-skills 항목의 검사와 규칙 파일에 한 줄씩).
`src/generation.rs` 는 주석만, `src/project_update.rs` 는 글자열 세 곳만 바꿔요. `tests/areas.toml` 은 쓰지 않아요:
새 테스트 파일의 첫 낱말 `project` 가 이미 표에 있어요(D-5).

## Plan

1. 이 SPEC 을 커밋하고 카드에 연결해요. 승인 노트를 받은 뒤 `moai card scope` 로 쓰기 범위를 넓혀요.
2. RED: `tests/project_harness_label.rs`(소스와 워크플로 글을 읽고 아무 명령도 돌리지 않아요)를 쓰고 reads-skills
   항목에 등록해요. `tests/project_harness_pick.rs` 의 선택 없는 시험에 경로 셋 단언을 넣고, 생성된 적 없는
   프로젝트의 고정 시험을 더해요. 바꾸기 전에 무엇이 왜 빨간지 적어요.
3. GREEN — 글: `INSTALL_REASON`, 거절 노트 둘, `update.md` 와 `project.md` 의 이름 문장과 세 자리, 두 고정.
   소스 매니페스트를 `python3 -I tools/source_manifest.py write` 로 다시 만들어요.
4. GREEN — 뒤따르는 일: 세 파일이 모두 conflict 가 되는 고정 장치, `record_generation` 의 주석, `research.md` 의
   URL, 픽스처 프로필의 호스트.
5. 가드 테스트마다 뮤턴트 하나, 같은 head 의 `moai check affected`, `moai trailer-check`, READY.

## Risks

- 워크플로 글 두 파일을 읽는 시험이 많아요. 바뀐 글 때문에 깨지는 고정은 같은 변경에서 사실대로 맞추고, 검사를
  느슨하게 하지는 않아요.
- `skills/moai` 가 바뀌면 소스 매니페스트의 sha256 도 같은 변경에서 맞춰야 해요. 매니페스트는 Python 도구가 만들고
  Rust 시험(`product_cli` 의 `skill_package`)이 확인해요.
- 한국어 이름은 모델이 화면에 그리는 것이라 끝까지 시험하지 못해요. 시험은 모델에게 주는 문장을 고정할 뿐이에요.
  실제 화면 관측은 not_run 으로 남아요.
- 선택 없는 시험의 고정 장치가 세 파일을 모두 conflict 로 만들려면 셋 다 "양쪽이 움직인" 상태여야 해요. 측정해
  보니 `.gitignore` 관리 구간은 프로필 어느 칸으로도 바뀌지 않아요(D-9). 그래서 그 파일은 매니페스트의 기준선을 옛
  빌드의 값으로 둬요. 선택 있는 시험의 뜻(선택이 있으면 conflict 없음)이 변하지 않게 했어요.
- 기준 커밋 `596b525dbc` 뒤로 통합 가지가 움직였어요. 착지 때 병합은 착지자가 해요. `tools/repo_scans.json` 은
  다른 카드도 자주 고쳐서 병합 충돌이 날 수 있어요(풀이는 두 추가를 모두 남기는 것).
- 설치된 `moai`(4.2648.23)의 SPEC 검사가 팁의 검사보다 느슨할 수 있어요. RUN 에서 나무의 바이너리로
  `spec check` 를 다시 돌려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 표시 이름은 설치 이유 글과 워크플로 글로 전해요. 계획 JSON 에 표시 이름 칸을 더하지 않아요 | 계획 JSON 은 기계용이고 혼자서 화면에 나오지 않아요. 화면은 모델이 워크플로 글대로 그려요. JSON 에 한국어를 넣지 않아요 | 계획의 단계에 `label` 칸을 더하고 글에서 그 칸을 읽게 해요 |
| D-2 | 설치 이유의 괄호는 `CLAUDE.md` 를 말해요(지금은 옛 이름 `AGENTS.md`) | 단계가 오늘 쓰는 파일이 `CLAUDE.md` 예요. 이름을 바꾸는 같은 글자열에 옛 파일 이름을 남기지 않아요. 되돌릴 수 있어요 | 괄호의 파일 이름을 `AGENTS.md` 로 되돌려요 |
| D-3 | `project.md` 5단계에서 시험이 고정한 문장들은 글자 그대로 두고 이름 문장을 5단계 첫문장으로 더해요. `update.md` 에는 1단계의 열거 앞에 더해요 | `tests/project_preview_order.rs` 와 `tests/update_route.rs` 의 고정을 지켜요. 5단계와 1단계가 계획을 사용자에게 읽어 주는 자리예요 | 문장의 자리를 옮겨요 |
| D-4 | 생성된 적 없는 프로젝트의 기록은 바꾸지 않고 주석과 시험이 오늘 동작을 적어요 | 리더가 정한 뒤따르는 일은 "주석이 사실대로 말한다"예요. 동작을 바꾸면 연결 검사가 달라져서 따로 결정할 일이에요 | 동작을 바꾸는 새 SPEC |
| D-5 | 새 테스트 파일은 `tests/project_harness_label.rs` 예요 | 첫 낱말 `project` 가 `tests/areas.toml` 에 있어서 공유 표를 쓰지 않아요. 11-21 의 이름 `harness_step_label` 은 표에 없는 낱말로 시작해요 | 이름을 바꾸고 표에 낱말을 더해요 |
| D-6 | 11-21 초안과 옛 가지는 참고 자료로만 읽어요. 아무것도 합치지 않아요 | 그 가지는 다른 흐름 위에서 만들어졌고 검사 B1 수리 전의 테스트가 있어요. 요구는 오늘 나무에서 다시 읽었어요 | 없음 |
| D-7 | 일본어·중국어 이름은 적지 않고 문장이 "다른 대화 언어로 옮긴다"고 해요 | 저장소에 그 언어의 표 판이 없어요. 사용자 결정도 "언어 표 판이 있으면 그에 맞게"였어요 | 표 판이 생기면 이름을 더해요 |
| D-8 | 거절 노트의 `AGENTS.md`(병합 단계가 옮기는 옛 파일)는 그대로 두고 단계 이름만 바꿔요 | 그 파일 이름은 이 SPEC 의 대상이 아니에요. 파일 이름의 참거짓은 다른 카드가 정해요 | 그 글을 따로 고쳐요 |
| D-9 | 선택 없는 시험의 고정 장치에서 `.gitignore` 는 매니페스트의 기준선을 옛 빌드의 값으로 둬요 | RUN 에서 측정했어요: 프로필 칸 일곱 가지(purpose, tools, development_methods, test_command, protected, review_blocks, build_command)를 하나씩 옮겨도 `.gitignore` 구간은 `unchanged` 예요. CLAUDE.md 는 `tools` 만, REVIEW.md 는 리뷰 규칙만 따라요. 도구가 업데이트되면 무시 목록이 바뀌고 그 전에 생성된 프로젝트의 기준선은 옛 값으로 남아요(감사가 측정한 현장의 `.gitignore` conflict 가 그 모양이에요). 구현 중에 정했어요(승인 뒤에 더한 행) | 기준선을 건드리지 않고 다른 프로필 칸을 찾아요(측정으로는 없어요) |

## Result

2026-11-24 에 이 세션이 구현했어요. 기준 통합 가지 `596b525dbc`, macOS arm64(개발 Mac), 이 워크트리와 그 빌드
폴더 `target/`. 무거운 실행은 모두 `moai heavy-slot-acquire` 를 거쳤고 테스트 빌드는 `RUSTC_WRAPPER=sccache`,
`CARGO_INCREMENTAL=0` 이에요. 로그는 커밋하지 않는 `target/harness-step-label/` 에 있어요.

커밋(기준 위, 순서대로): `b393e5ff9b` SPEC 판 1 → `46f1bfb181` 승인 기록 → `0998b7f4b9` RED → `0cb0309a10`
GREEN → `3e24a6b18c` 식별자 고정 보강 → `df796faaa5` 스캔 표 등록 → 이 Result.

- AC-1: met — install_reason_names_the_step_project_harness 가 빌드한 바이너리로 빈 임시 프로젝트의 계획을 돌려 설치 단계의 reason 이 "profile, project harness (CLAUDE.md managed section and generated files), Claude Code settings, Claude Code hooks, GitHub templates; …" 임을 읽어요. RED(커밋 `0998b7f4b9`, nextest run `8aa1598f-97d9-411c-9290-61da44e09a1b`)에서는 "profile, generation (AGENTS.md managed section …" 로 실패했어요. 뮤턴트 m1.
- AC-2: met — identifiers_stay_as_they_were: 계획이 단계를 `generation` 으로 적고 소스의 `Step` 값 둘과 대기 형태 둘이 id 를 그대로 쓰고 명령과 매니페스트 경로와 도움말이 그대로예요. RED 때부터 통과하는 고정이에요. 뮤턴트 m5.
- AC-3: met — workflow_text_names_the_step_in_english_and_korean 과 workflow_text_tells_to_translate_and_keeps_the_id_for_tools: 두 워크플로 글이 이름 문장을 정확히 한 번씩 품고 다섯 조각이 있어요. RED 에서는 update.md 에 문장이 0 번이었어요. GREEN(nextest run `ca7afe68-77d5-46be-8aba-b81ba33b8a07`, 8개 대상 44개 중 44 통과). 뮤턴트 m2·m3.
- AC-4: met — workflow_text_names_the_step_where_it_lists_it 가 통과하고 project_preview_order 3개와 project_applied_table 2개가 같은 실행에서 통과해요(고정한 문장은 그대로). RED 에서는 "the generation `files`" 로 실패했어요. 뮤턴트 m4.
- AC-5: met — refusal_notes_call_the_step_project_harness 가 빈 프로젝트 계획의 undo_order 를 읽어 통과하고 update_end_to_end_cli 8개가 같은 실행에서 통과해요. RED 에서는 "<agents_receipt of the generation step>" 로 실패했어요. 뮤턴트 m6.
- AC-6: met — without_the_pick_a_hand_edited_tracked_file_blocks_the_change_of_one_agent: 선택 없이 계획한 conflict 경로가 정확히 `.gitignore` 와 `CLAUDE.md` 와 `REVIEW.md` 셋이에요. RED 에서는 `REVIEW.md` 하나였어요(감사가 본 그대로). 고정 장치는 D-9 대로예요. 같은 파일의 선택 있는 시험 셋이 그대로 통과해요. 뮤턴트 m7(tools 를 되돌림)과 m8(`.gitignore` 기준선을 되돌림).
- AC-7: not_run — `record_generation` 의 주석은 아직 못 고쳤어요. `src/generation.rs` 를 다른 카드(네이티브 역할 링크 거절 카드)가 쥐고 있어서 줄 임대가 두 번 거절됐어요(2026-11-24T10:44Z 와 11:01Z 에 시계를 읽었어요). 리더 결정(2026-11-24T10:06:12Z 2번)대로 READY 에 not_run 으로 적어요. 고정 시험 a_never_generated_project_keeps_the_current_profile_digest_under_the_pick 은 RED 실행과 GREEN 실행 모두에서 처음부터 통과해요: 생성된 적 없는 프로젝트에서 선택으로 적용하면 매니페스트의 `profile_sha256` 는 적용에 쓴 프로필의 digest 이고 `interview_revision` 과 `workflow_id` 는 비어 있어요.
- AC-8: met — `research.md` 셋째 행이 URL 둘을 적어요(`grep -n "the same testing pages as above"` 가 출력 없이 끝나요). `grep -c codex tests/generation_request_limits.rs` 가 0 을 찍어요. generation_request_limits 5개가 GREEN 실행에서 통과했어요.
- AC-9: met — `python3 -I tools/source_manifest.py check` exit 0, `moai spec check` 오류 0. `moai check affected` 영수증은 카드의 READY 노트에 있어요. 새 테스트 파일은 reads-skills 와 reads-src 두 항목의 검사와 규칙 파일에 들어갔고(커밋 `df796faaa5`, 줄 임대), 그 검사 표 대상 112개가 통과해요(nextest run `5e92549a-93f3-4fd5-9012-caab9cfbdbbd`).

뮤턴트: 커밋 `3e24a6b18c` 에서 한 자리에서 여덟 개(가드 테스트마다 하나, 선택 없는 가드는 둘)를 돌렸어요.
여덟 모두 exit 100 이고 가드 테스트가 실패 목록에 있어요(일곱은 그 테스트 하나만, m3 은 문장 시험과 함께). 매번
`git checkout --` 로 돌려맞춘 뒤 HEAD 와 같음을 확인했어요. 실행 스크립트의 첫 요약 줄은 "0 of 8 killed" 를 인쇄했어요.
가드 이름에 대상 이름을 붙인 채 실패 목록과 비교한 스크립트의 잘못이었고, 저장된 원본 로그 여덟 개를 읽어 "8 of 8
killed" 를 확인했어요. 스크립트는 고쳤지만 다시 돌리지는 않았어요.

실행 안 함(not_run): `record_generation` 의 주석(위), 한국어 이름이 실제 화면에 보이는 관측(모델이 그려요),
Windows 와 Linux.
