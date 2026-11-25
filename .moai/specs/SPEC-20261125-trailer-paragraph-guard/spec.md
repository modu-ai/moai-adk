---
id: SPEC-20261125-trailer-paragraph-guard
title: "Moai 트레일러가 마지막 문단 밖에 있으면 커밋할 때 거절하고, 검사기가 그 원인을 이름으로 말해요"
status: draft
revision: 2
created: 2026-11-25
---

# Moai 트레일러가 마지막 문단 밖에 있으면 커밋할 때 거절하고, 검사기가 그 원인을 이름으로 말해요

## Intent

`Moai-Spec`, `Moai-Session`, `Moai-Verified` 줄을 `Co-Authored-By` 와 빈 줄로 갈라 쓰면 git 은 마지막 문단(`Co-Authored-By` 한 줄)만 트레일러 블록으로 읽어요. 그래서 `moai trailer-check` 는 `missing_spec` 와 `missing_verified` 로 실패하는데, commit-msg 훅은 카드 번호만 보기 때문에 커밋은 통과해요. 같은 실수가 세 번 나왔어요(`SPEC-20261124-card-select-batch-land` 수리 커밋, `SPEC-20261125-rust-only-finish` 의 `5b89e3af41`, 아직 착지하지 않은 다음 조각의 세 커밋). 감사나 착지 뒤에야 드러나서 매번 재감사나 예외 판단이 들었어요.

이 SPEC 은 코드를 읽고 실제로 돌려 본 사실 위에서 쓴 거예요(이 저장소 `NeXT` `0befa9cba6`, 빌드 `moai 4.2648.26`, 2026-11-25).

- 원인 재현: `5b89e3af41` 의 메시지는 `Moai-*` 세 줄 다음에 빈 줄, 그다음 `Co-Authored-By` 예요. `git interpret-trailers --parse` 는 `Co-Authored-By` 한 줄만 돌려주고, `moai trailer-check 5b89e3af41^..5b89e3af41 --branch spec/probe` 는 `missing_spec`(fails true)와 `missing_session` 을 내요. 읽는 곳은 모두 git 의 트레일러 블록이에요: `src/trailers.rs` 의 `log_trailers`(`%(trailers:key=…)`), `src/bundle.rs` 의 `spec_ids`, 그리고 자체 파서인 `src/audit_gate.rs` 의 `sessions_of`(마지막 문단만 읽고, 테스트 `trailers_are_read_from_the_last_paragraph_only` 가 그 모양을 고정해요).
- 훅: `install/commit-msg` 는 `moai trailer-check --message-file "$1"` 를 부르고, 그 모드(`message_file_check`)는 카드 번호 검사만 해요. 이 저장소에 설치된 훅(`.git/hooks/commit-msg`)도 같은 내용이에요.
- 이미 착지한 커밋: `origin/v2026.11.0..NeXT`(2183 커밋)에서 메시지에 `Moai-Spec:` 줄이 있는데 git 이 트레일러로 읽지 못하는 커밋이 8개예요(`5b89e3af41`, `b7e1bcfd1c`, `25a75d204a`, `fc41ab568d`, `534617602f`, `bd7cdaf8bd`, `865a894084`, `a4dd2a0864`; 나머지 3개는 `Moai-Spec` 줄 자체가 없는 옛 커밋이에요).
- 막는지 실제로 돌려 본 결과: (1) 이 8개는 모두 `NeXT` 안에 있어서 `moai bundle check <새 head>` 와 `bundle verify` 의 `trailer-check <tip>..HEAD` 범위에 안 들어가요. (2) 원격 기준(`origin/v2026.11.0`) 전체 범위 `trailer-check --branch release/probe` 는 이미 132건 실패인데 8건이 이 8개(`missing_verified`)이고 나머지 124건은 이 문제가 아닌 옛 커밋이에요. `missing_spec` 은 `spec/` 브랜치에서만 실패라 릴리스 브랜치에서는 실패가 아니에요. (3) `.github/workflows/` 의 어느 워크플로도 `trailer-check` 나 `bundle check` 를 부르지 않아요. `audit-gate` 만 커밋 메시지를 읽고(마지막 문단), 이 8개는 저자 목록에 안 들어가요. 막지는 않지만 구멍이 하나 있어요: 세션 `850726f6-eb78-492d-978f-66edb7ab9649` 는 범위 안에서 `bd7cdaf8bd` 와 `865a894084` 두 커밋에만 나와서(둘 다 갈라진 모양), 그 세션이 릴리스 PR 을 감사해도 `audit-gate` 는 저자로 안 읽어요. 나머지 세션 둘(`322b596c-…`, `603b3ad0-…`)은 다른 커밋의 마지막 문단에도 나와요.

## Non-goals

- 이미 착지한 커밋의 메시지를 다시 쓰는 일(착지한 해시가 바뀌어요).
- 검사기와 `audit-gate` 가 마지막 두 문단을 읽게 하는 일(선택지 B). 이유는 D-1 에 있어요.
- `Co-Authored-By` 가 없는 커밋이나 `Moai-*` 가 마지막 문단에 모두 있는 커밋의 판정 변경.
- `moai trailers` 출력 형식 변경과 `audit-gate` 의 저자 읽기 변경(D-3 의 선택지 2 는 따로 SPEC 이 필요해요).
- 원격 기준 전체 범위의 옛 `missing_verified` 124건 정리.

## Requirements

<!-- This SPEC adds a check that blocks. It starts blocking in the commit-msg hook only after AC-5 measures the would-block set on the real range; D-2 records the switch. -->
- REQ-1: When `moai trailer-check --message-file` reads a message with a paragraph before the last one whose lines are all `Key: value` and one key is `Moai-Spec`, `Moai-Session` or `Moai-Verified`, the check shall fail with a `split_trailer_block` finding. · failure: none — 메시지를 읽을 수 없으면 지금처럼 오류 메시지로 끝나요.
- REQ-2: The `split_trailer_block` finding shall name the key, its message line and the one-line fix: remove the blank line so the `Moai-*` lines and any `Co-Authored-By` line are the last paragraph. · failure: none — 값이 없으면 줄이 없어요.
- REQ-3: When `moai trailer-check <range>` reads a commit whose body holds such a paragraph, the check shall report `split_trailer_block` for that commit as a finding that does not fail. · failure: none — 이미 착지한 커밋이 새로 실패하지 않아요.
- REQ-4: When the check reads a message whose `Moai-*` lines are all in the last paragraph or whose earlier paragraphs are prose, it shall report no `split_trailer_block`. · failure: none — 본문이 `Moai-Spec:` 글자를 문장 안에서 인용해도 통과해요.
- REQ-5: The commit-trailers reference shall say that the `Moai-*` lines and `Co-Authored-By` are one last paragraph with no blank line between them. · failure: none.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소에서 `Moai-*` 세 줄과 `Co-Authored-By` 사이에 빈 줄이 있는 메시지 파일을 `trailer-check --message-file` 로 판정하면 exit 1 이고 `split_trailer_block` 이 키 이름과 줄 번호와 고치는 법 한 줄을 담아요 · verify: `cargo test --locked --offline --test card trailer_paragraph_guard_cli::ac1_`
- AC-2 (REQ-1): 설치할 commit-msg 훅(`install/commit-msg`)을 그 메시지로 실제로 실행하면 exit 1 이고, 같은 메시지에서 빈 줄만 지우면 exit 0 이에요 · verify: `cargo test --locked --offline --test card trailer_paragraph_guard_cli::ac2_`
- AC-3 (REQ-3): 갈라진 커밋이 든 범위의 `trailer-check` 는 그 커밋에 `split_trailer_block`(fails false)을 싣고, 그 줄 때문에 `failing` 이 늘지 않아요 · verify: `cargo test --locked --offline --test card trailer_paragraph_guard_cli::ac3_`
- AC-4 (REQ-4): 마지막 문단에 `Moai-*` 가 다 있는 메시지, `Co-Authored-By` 가 없는 메시지, 앞 문단 문장 안에 `Moai-Spec: x` 를 인용한 메시지는 통과해요 · verify: `cargo test --locked --offline --test card trailer_paragraph_guard_cli::ac4_`
- AC-5 (REQ-3, REQ-4): 이 빌드로 `moai trailer-check origin/v2026.11.0..NeXT --branch release/probe` 를 돌려 `split_trailer_block` 이 정확히 위 8개를 이름으로 대고 그 밖의 커밋에는 없어요. 숫자와 해시를 READY 노트에 적어요 · verify: `moai trailer-check origin/v2026.11.0..NeXT --branch release/probe > target/c501/range-after.json`
- AC-6 (REQ-5): 참조 문서에 그 한 문장과 빈 줄이 없는 예시가 있어요 · verify: `cargo test --locked --offline --test card trailer_paragraph_guard_cli::ac6_`
- AC-7 (REQ-1…REQ-5): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test card trailer`
  verify: `cargo test --locked --offline --lib trailers::`
  verify: `cargo test --locked --offline --lib audit_gate::`

## Scope

- write: `.moai/specs/SPEC-20261125-trailer-paragraph-guard/spec.md`, `src/trailers.rs`, `tests/card/trailer_paragraph_guard_cli.rs`, `tests/card/main.rs` (`mod` 한 줄), `install/commit-msg` (주석만), `skills/moai/references/internal/commit-trailers.md`, `skills/moai/references/source-manifest.json` (참조 문서의 sha256 한 줄)
- read-only: `src/audit_gate.rs`, `src/bundle.rs`, `src/bundle/verify.rs`, `src/ops/trailer_check.rs`, `install/git-hooks/pre-push`, `tests/card/trailer_check_cli.rs`
- forbidden: 진짜 HOME 의 Claude 설정 폴더와 HOME 상태 폴더, 설치된 moai 바이너리와 이 저장소에 설치된 훅 파일, 이미 착지한 커밋과 `NeXT` 브랜치, `src/audit_gate.rs` 의 저자 읽기, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1…AC-4, AC-6 을 새 통합 테스트 모듈 `tests/card/trailer_paragraph_guard_cli.rs` 에 써요(`tests/card/main.rs` 에 `mod` 한 줄, 낱말 `trailer` 가 옮겨진 `card` 영역이라 최상위 파일은 `test_area_layout` 이 거절해요; 임시 저장소와 임시 HOME, 이 빌드의 바이너리, `install/commit-msg` 는 `PATH` 에 이 바이너리를 둔 임시 폴더에서 실행). 기준에서는 갈라진 메시지가 통과해서 맞는 이유로 실패해야 해요.
2. `src/trailers.rs`: 메시지 문단을 나눠 마지막 문단 앞에서 `Key: value` 줄만으로 된 문단에 `Moai-*` 키가 있으면 찾는 작은 함수 하나를 두고, `message_file_check`(REQ-1, REQ-2)와 범위 판정의 `findings`(REQ-3, 본문 `body` 에서 찾아 fails false)가 같은 함수를 써요.
3. `install/commit-msg` 의 머리 주석에 두 번째 검사를 적어요(동작은 그대로예요: 이미 `--message-file` 를 불러요). 참조 문서 `commit-trailers.md` 에 한 문장과 예시를 넣고 소스 매니페스트의 해시를 `moai` 가 쓰는 대로 갱신해요(REQ-5).
4. AC-5 실행, 지킴이마다 뮤턴트(판정 함수의 키 확인 제거, `fails` 뒤집기, 마지막 문단 제외 조건 제거), 이웃 테스트, READY 에서 `check affected --card` 한 번.

## Risks

- 설치된 commit-msg 훅은 `PATH` 의 `moai` 를 불러요. 새 규칙은 그 `moai` 가 이 변경을 담은 빌드로 바뀐 뒤에야 커밋 시점에 작동해요(착지 뒤 리더의 `moai update`). 그 전에는 지금처럼 통과해요. 이 SPEC 은 설치본을 건드리지 않아요.
- 문단 규칙이 너무 넓으면 정상 메시지를 막아요. 그래서 "앞 문단의 모든 줄이 `Key: value` 이고 키 하나가 `Moai-*`" 로 좁혔고, AC-4 가 문장 인용을, AC-5 가 실제 2183 커밋을 확인해요. AC-5 가 8개 밖의 커밋을 이름으로 대면 훅은 observe(경고만, exit 0)로 두고 D-2 를 사용자에게 올려요.
- 검사기는 이제 두 가지를 말해요: 원인(`split_trailer_block`)과 증상(`missing_spec`). 둘이 같은 커밋에 같이 보이는 것이 정상이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | **선택지**: (A) 커밋 시점에 거절하고 원인을 이름으로 말해요(기본, 이 SPEC); (B) 검사기가 마지막 두 문단을 읽어요. 기본 A | B 는 git 의 트레일러 블록(`%(trailers)`, `spec_ids`)과 `audit_gate::sessions_of`(마지막 문단만, 테스트로 고정)와 어긋나고, 본문 인용이 저자나 SPEC 으로 읽히는 길을 열어요(아이덴티티 위조 면). A 는 같은 실수를 만드는 순간에 끝내요 | A 의 검사를 빼요 |
| D-2 | **선택지**: 훅이 바로 거절해요(exit 1, 기본) / observe(경고만) 로 시작해요. 기본은 거절이고, AC-5 가 8개 밖을 대면 observe | 이 SPEC 의 막는 검사는 observe 부터라는 템플릿 규칙을 지키되, 실제 범위 측정(AC-5)으로 거짓 양성이 0 임을 먼저 보여요 | observe 로 낮추는 것은 한 줄이에요 |
| D-3 | **선택지 (이미 착지한 8개)**: (1) 다시 쓰지 않아요. 릴리스 PR 감사 기록에 "감사자 세션은 `850726f6-eb78-492d-978f-66edb7ab9649` 가 아니에요" 를 남겨요(기본, 코드 없음); (2) `audit-gate` 가 `Moai-*` 만으로 된 바로 앞 문단도 저자로 읽게 해요(`src/audit_gate.rs` 변경, 별도 SPEC, 위조 면 검토 필요); (3) 메시지를 다시 써요(해시가 바뀌어 재감사, 하지 않아요) | 막는 곳은 없어요: 8개는 `NeXT` 안이라 `bundle check`/`verify` 범위 밖이고 CI 는 `trailer-check` 를 안 불러요. 남는 것은 `audit-gate` 의 저자 목록에서 세션 하나가 빠지는 점이에요 | 기본이라 되돌릴 것이 없어요 |
| D-4 | 개정 2: 새 통합 테스트는 최상위 파일이 아니라 `tests/card/` 모듈이에요. REQ·AC 의 뜻은 그대로고 테스트 위치와 verify 줄(`--test card trailer_paragraph_guard_cli::acN_`)과 Scope 만 바뀌어요 | `tests/areas.toml` 이 낱말 `trailer` 를 `card` 영역에 두고 `tests/card/main.rs` 가 있어 영역이 옮겨졌어요(`test_area_layout` 의 `area_table_maps_every_real_test_file` 이 최상위 새 파일을 거절해요). 개정 1 은 이를 몰랐어요 | 최상위 파일과 영역 표 변경이 필요한데 표는 공유 파일이에요 |

## Interview

- Q1: 이 SPEC 개정 1을 승인하시나요? (리더의 선택 질문, 결정 노트 8879 가 인용해요; 개정 1 의 digest sha256:7e0d5fa8…cf06) → A: 승인, 커밋할 때 거절 (Recommended). 개정 2 는 D-4 의 테스트 위치만 바꿔요.

## Result

(구현 전. 구현하면서 채워요.)
