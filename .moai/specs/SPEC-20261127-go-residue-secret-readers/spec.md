---
id: SPEC-20261127-go-residue-secret-readers
title: "go-residue never opens a reader with a secret-shaped name and keeps every Go file when it exists"
status: approved
revision: 1
created: 2026-11-27
parent: SPEC-20261126-go-residue-in-use
approval:
  by: user
  at: 2026-11-27
  basis: "카드 노트 10717 (리더가 옮겨 적은 사용자의 직접 선택 \"승인, 경로 전체 (Recommended)\", 2026-11-27, D-1 은 A). 레인은 승인을 기록하지 않았고 노트를 인용했어요. 승인 digest 는 Interview Q2 를 옮긴 뒤 moai spec show 로 다시 읽었어요(같은 값)"
  digest: "sha256:2eb7ce618092f64f0ed43346cd35c8ef9ff1f3db3a1a4a13833b759211c4385e"
---

# go-residue never opens a reader with a secret-shaped name and keeps every Go file when it exists

## Intent

`/moai update` 의 `go-residue` 조사는 Go 시대 파일을 읽는 곳(지시 파일, `.claude/rules` 파일, 설정 파일, 설정이 이름 부른 스크립트)을 찾으려고 그 읽는 곳 파일을 열어요. 승인된 부모 SPEC(`SPEC-20261126-go-residue-in-use`)의 REQ-3 은 "읽는 곳이 있는데 읽을 수 없으면(읽기 오류, 링크, 4 MiB 초과, JSON 아님, UTF-8 아님, 비밀값처럼 보이는 이름) 모든 `go_same`/`go_changed` 파일을 `keep` 으로 두고 `unread_readers` 에 까닭을 대라"고 해요. 그런데 구현(`src/go_residue/readers.rs`)에는 비밀값처럼 보이는 이름의 가지가 없어요. 그래서 `@.env` 를 가져오는 `CLAUDE.md` 나 `credentials.sh` 같은 이름의 훅 스크립트가 있으면 조사가 그 파일을 열어 읽어요. 리뷰(노트 10598, 머리 `efd5d76d32`)가 이것을 짚었고, 부모 SPEC 의 어떤 AC 도 이 가지를 덮지 않아요.

이 SPEC 은 그 가지를 더해요. 읽는 곳 경로에 비밀값처럼 보이는 이름이 있고 그 항목이 실제로 있으면 파일을 열지 않고(열지 않는 것은 내용을 읽거나 해시하지 않는다는 뜻이고, 있는지는 디렉터리 항목 정보로만 봐요) "읽을 수 없는 읽는 곳"으로 다뤄서 모든 Go 파일을 `keep` 으로 두어요. 실패하는 방향은 계속 "옮기지 않는다"예요. 읽는 곳이 되는 파일의 종류와 읽는 법은 바꾸지 않아요.

읽은 사실(NeXT `e98798e41f`, 이 트리, 코드 읽기만 했고 이 SPEC 을 쓰면서 실행하지는 않았어요):

1. 읽는 곳 파일을 여는 곳은 `Collector::read` 한 곳이에요(`src/go_residue/readers.rs`). 지시 파일과 `@` 가져오기, `.claude/rules` 파일, 설정 두 개, 스크립트 따라가기가 모두 이 함수를 지나요. 이 함수는 `go_state::read_bounded_text` 를 불러요. 그 함수는 링크와 크기만 보고 이름은 보지 않아요(`src/go_state.rs:1874-1888`).
2. 비밀값처럼 보이는 이름의 판단은 이미 하나로 있어요: `go_state::secret_name(name)`(= `home_state::is_secret` 이거나 `worktree_include::secret_name`)과 경로 모든 구성 요소를 보는 `go_state::has_secret_component(rel)`(`src/go_state.rs:130-138`). 조사의 다른 곳(`src/go_residue.rs:304`, 매니페스트 경로 검사)이 이미 같은 규칙을 써요. `readers.rs` 는 이 두 함수 중 아무것도 부르지 않아요(`linked_component`, `read_bounded_text` 만 불러요).
3. 읽을 수 없는 읽는 곳을 다루는 길은 이미 있어요: `Collector::unread(file, reason)` 이 `unread_readers` 에 올리고, 그러면 모든 후보가 `keep` 이고 적용 명령이 없어요(부모 SPEC 의 AC-2 테스트 `a_reader_that_cannot_be_read_keeps_every_go_file` 가 지금 일곱 가지 까닭으로 이를 고정해요). 새 까닭 하나만 더하면 돼요.
4. 부모 SPEC 의 조사 쪽(`src/go_residue.rs`)은 비밀값처럼 보이는 이름을 센 적이 있어요(`secret_names`). 이 SPEC 은 그 세는 값이나 답의 모양을 바꾸지 않고, `unread_readers` 에 항목 하나가 더 생길 수 있을 뿐이에요.

읽는 사람: 승인하는 사용자와 리더, 구현 레인, 감사자.

## Interview

- Q1: 무엇이 필요한가요? → A: 카드 본문(리더가 옮긴 리뷰 노트 10598): "SPEC REQ-3 says a secret-shaped name is treated as a place that could not be read, so the file it may reference is kept; readers.rs has no such branch and no AC covers it. Failing direction must stay 'do not move'." (2026-11-27)
- Q2: 아래 Decisions 의 "승인 때 골라요" 항목(D-1)과 이 SPEC 의 첫 판(revision 1)을 승인할까요? → A: "승인, 경로 전체 (Recommended)"(2026-11-27 사용자 직접 선택, 리더가 질문 도구의 답을 카드 노트 10717 에 옮겨 적음). D-1 은 A 예요: 경로의 폴더 이름이든 파일 이름이든 비밀값 모양이면 열지 않아요(`go_state::has_secret_component`). 대가: `credentials-handling.md` 같은 평범한 문서도 걸려서 그 프로젝트의 Go 정리가 멈출 수 있고, 문서 이름을 바꾸면 풀려요.

## Non-goals

- 비밀값처럼 보이는 이름의 규칙 자체(`go_state::secret_name`, `home_state::is_secret`, `worktree_include::secret_name`)를 바꾸는 일. 이 SPEC 은 그 규칙을 부르기만 해요.
- `src/go_state.rs`, `src/go_residue.rs`, `src/go_residue_apply.rs` 의 변경. 조사의 분류와 적용 명령은 그대로예요.
- 읽는 곳이 아닌 파일의 이름 검사(옮길 후보의 비밀값 검사는 이미 `go_residue.rs` 가 해요).
- 비밀값 파일의 내용을 읽거나 해시하거나 답에 싣는 일. 답에는 프로젝트 경로와 고정된 까닭 글만 나가요.

## Requirements

<!-- 이 SPEC 은 옮기는 파일을 줄이는 쪽으로만 바꿔요. 읽는 곳을 확인하지 못하면 옮기지 않아요. -->
- REQ-1: When the go-residue survey is about to read a reader file whose project path holds a secret-shaped name and the entry exists, the survey shall leave the file unopened and record the reader as unread with the reason `has a secret-shaped name`.
- REQ-2 (REQ-1): IF a secret-shaped reader is recorded as unread, THEN the survey shall suggest `keep` for every `go_same` and `go_changed` file, list that reader under `unread_readers`, and print no apply command.
- REQ-3 (REQ-1): IF the secret-shaped path does not exist, THEN the survey shall treat it as no reader and block nothing.
- REQ-4: When the survey decides that a reader path is secret-shaped, the survey shall use only the rule `go_state::has_secret_component`, which tests every component of the path. · failure: none — 이름만 비교하는 규칙이고 파일을 열지 않아요.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 프로젝트에서 `CLAUDE.md` 가 존재하는 `.env`(UTF-8 이 아닌 바이트로 채움)를 `@` 로 가져오면 모든 `go_same`/`go_changed` 가 `keep`, `unread_readers` 가 `.env` 와 까닭 `has a secret-shaped name` 을 대고(UTF-8 이 아니라는 까닭이 아니에요: 열었다면 그 까닭이 나왔을 거예요), 적용 명령이 없고 트리는 그대로예요 · verify: `cargo test --locked --offline --test project_update_go_residue_cli a_reader_with_a_secret_shaped_name_is_never_opened`
- AC-2 (REQ-1, REQ-2): 설정이 이름 부른 스크립트 `.claude/hooks/credentials.sh`(내용이 조사 대상 Go 파일 하나를 이름 부름), `.claude/rules/credentials.md` 와 구성 요소가 비밀값 이름인 폴더 아래의 스크립트 각각에서, 그 스크립트가 열렸다면 생겼을 `go_in_use`(`used_by`)가 없고 모든 후보가 `keep` 이며 `unread_readers` 가 그 경로를 대요 · verify: `cargo test --locked --offline --test project_update_go_residue_cli a_secret_shaped_script_or_rule_is_unread_and_names_no_file`
- AC-3 (REQ-3): 가져오기와 스크립트가 존재하지 않는 비밀값 이름(`.env`, `credentials.sh`)을 가리키면 `unread_readers` 가 비고 증명된 파일이 후보로 남아 적용 명령이 인쇄돼요 · verify: `cargo test --locked --offline --test project_update_go_residue_cli a_missing_secret_shaped_name_blocks_nothing`
- AC-4 (REQ-4): 읽는 곳 코드가 `has_secret_component` 를 부르고 비밀값 이름 목록을 새로 두지 않아요. `home_state::is_secret` 이 대는 이름(`credentials.json`)과 `worktree_include::secret_name` 이 대는 이름(`server.key`)이 모두 막혀요 · verify: `cargo test --locked --offline --test project_update_go_residue_cli a_secret_shaped_name_follows_the_shared_rule`
- AC-5 (REQ-1, REQ-2, REQ-3): 부모 SPEC 의 기존 테스트가 고치지 않고 통과해요 · verify: `cargo test --locked --offline --test project_update_go_residue_cli`

## Scope

- write: `.moai/specs/SPEC-20261127-go-residue-secret-readers/spec.md`
- write: `src/go_residue/readers.rs`, `tests/project_update_go_residue_cli.rs`
- read-only: `src/go_state.rs`(`has_secret_component` 를 불러 써요), `src/go_residue.rs`, `src/go_residue_apply.rs`, `src/home_state.rs`, `src/worktree_include.rs`
- forbidden: `src/go_state.rs` 와 `src/go_residue.rs` 와 `src/go_residue_apply.rs` 의 변경, `CLAUDE.md`, `.moai/project.yaml`, 실제 프로젝트(`mo.ai.kr` 포함)와 실제 HOME 에 쓰기, 파일 삭제

## Plan

1. (지금, SPEC 폴더만) 이 SPEC 을 채우고 `moai spec check` 를 돌려요. SPEC-READY 노트에 digest 와 열린 선택(D-1)을 올리고 멈춰요. 승인은 사용자의 것이라 이 레인이 기록하지 않아요.
2. 승인 뒤 `moai card scope` 로 위 쓰기 범위를 더하고 쓰기 전마다 `moai card check` 를 해요.
3. 테스트 먼저(RED): `tests/project_update_go_residue_cli.rs` 에 AC-1 에서 AC-4 의 테스트를 더해요. 프레임은 지금 파일의 `unread_frame` 과 `unreadable_reader` 를 이어 쓰고, 비밀값 파일은 임시 프로젝트 안에만 만들어요. 변경 전 트리에서 AC-1·AC-2·AC-4 는 단언 실패여야 하고(AC-3 은 지금도 통과하는 가드예요) `moai tdd red` 로 기록해요(슬롯 `--max 2`). 셸 스크립트와 Python 은 쓰지 않아요.
4. `src/go_residue/readers.rs`: `Collector::read` 의 맨 앞(읽기 횟수를 세기 전)에서 `go_state::has_secret_component(rel)` 이고 항목이 있으면(`fs::symlink_metadata` 만 써요, 열지 않아요) `unread(rel, "has a secret-shaped name")` 를 하고 `None` 을 돌려줘요. 모듈 머리 주석의 읽기 설명에 이 가지를 한 줄 더해요.
5. 영수증 전 준비: `cargo fmt --all`, 대상 clippy, `project_update_go_residue_cli`·`update_end_to_end_cli` 를 nextest 로 돌려요. 가드 뮤턴트는 새 가드 테스트마다 한 줄씩 Result 에 적어요(편집 한 번, 슬롯 안 nextest, 평범한 `git restore` 한 줄): AC-1(새 가지를 끔), AC-2(스크립트 경로만 검사에서 뺌), AC-4(마지막 구성 요소만 봄). 그다음 `moai trailer-check`, `moai check affected --card` 한 번, READY 노트.

## Risks

- 이름이 `credentials` 로 시작하는 문서(`.claude/rules/credentials-handling.md`)가 있으면 비밀값이 아니어도 읽을 수 없는 읽는 곳이 되어 그 프로젝트의 Go 보관이 모두 멈춰요(과보존). 실패하는 방향이 "옮기지 않는다"이므로 받아들이는 비용이고, D-1 이 이것을 고르는 선택이에요. 사용자는 그 문서 이름을 바꾸면 풀려요.
- 구성 요소 하나만 비밀값 이름이어도 막아요(`.claude/secrets/x.sh`). 규칙이 `has_secret_component` 와 같아 다른 곳과 어긋나지 않지만 너무 넓다고 느낄 수 있어요.
- 파일이 열리지 않았다는 것은 테스트가 "열었다면 났을 결과가 나지 않았다"로만 보여요(UTF-8 이 아닌 바이트, 이름 부른 Go 파일). 열기 자체를 가로채 보지는 않아요(미확인).
- `readers.rs` 의 `FILE_LIMIT`(읽는 파일 1000개)은 비밀값 이름으로 건너뛴 항목을 세지 않아요. 비밀값 이름이 아주 많아도 한도에는 닿지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 (승인 때 골라요) | 경로의 어느 구성 요소든 비밀값 이름이면 읽는 곳 파일을 열지 않고 읽을 수 없는 읽는 곳으로 다뤄요(`has_secret_component`, 기본값 A). 대안 B: 마지막 이름만 봐요(`secret_name`): 비밀값 폴더 안의 평범한 이름 스크립트는 열려요 | 부모 SPEC 의 REQ-3 이 "비밀값처럼 보이는 이름"이라고만 하고 조사의 다른 곳(`go_residue.rs`, `tidy`)은 경로 전체나 구성 요소를 봐요. A 가 같은 규칙이라 어긋나지 않고, 실패 방향이 옮기지 않는 쪽이에요 | 검사 한 줄을 `secret_name(last)` 로 바꾸는 판 |
| D-2 | 비밀값 이름의 항목이 있는지는 `fs::symlink_metadata` 로만 봐요(링크도 있는 것으로 세요). 없으면 막지 않아요 | 부모 SPEC 의 REQ-3 이 "있는 읽는 곳"만 말해요. 메타데이터는 내용을 읽지 않아요 | 해당 없음 |
| D-3 | 새 까닭 글은 `has a secret-shaped name` 이고 답에는 경로와 이 글만 나가요 | 기존 까닭들이 같은 모양의 고정 글이에요(`has a link on its path` 등) | 글 한 곳 |
| D-4 | tier 2 로 다뤄요(리더 지시, 비밀값 파일을 읽는 길을 건드려요). SPEC 승인 전에는 코드를 쓰지 않아요 | 리더 메시지 2026-11-27 | 리더가 tier 1 로 정해요 |
