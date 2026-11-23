---
id: SPEC-20261123-ready-mutant-line
title: "A READY note carries one mutant line per new guard test, and bundle check warns when it does not"
status: draft
revision: 1
created: 2026-11-23
---

# READY 노트가 새 지킴 테스트마다 뮤턴트 한 줄을 싣고, bundle check 가 없으면 경고

## Intent

2026-11-23 하루에 같은 종류의 문제가 감사에서 네 번 걸렸어요. 새로 넣은 지킴 테스트(guard test: 어떤 코드를 지키려고 쓴 테스트)가 지키는 코드를 빼도 통과했어요: 테스트 병렬 제한의 cap 파싱, `/clear` 뒤 이어받기의 경로 정규화, 테스트 영역 표의 영역 연결, 공식 LSP 규칙의 가짜 서버(PATH 에서 `touch` 를 못 찾아 실행돼도 표식이 안 생김). 레인의 확인은 모두 초록이었어요. 초록 테스트가 "빼도 통과"인지는 지키는 코드를 빼 봐야 알아요.

이 SPEC 은 되풀이 문제를 소스에서 막아요. (1) READY 안내가 새 지킴 테스트마다 뮤턴트(mutant: 지키는 코드를 빼거나 뒤집은 복사본) 한 번과 그 테스트가 실패한 결과를 한 줄로 싣게 하고, (2) `moai bundle check` 가 그 줄이 없거나 읽히지 않거나 뮤턴트가 살아남았으면 경고하고, (3) 가짜 실행 파일 테스트가 PATH 에 기대지 않고 표식을 남기는 도우미를 `tests/common` 에 두어요.

새 검사는 막지 않고 관찰부터 시작해요(`internal/sdd` 11번): 경고는 `ok`·`findings`·종료 코드를 바꾸지 않아요. 막는 쪽으로 바꾸는 일은 관찰 결과를 본 뒤 기록하는 결정이에요(D-3).

## Non-goals

- 뮤턴트를 자동으로 만들거나 돌리는 도구. 레인이 손으로 돌리고 결과 한 줄을 적어요.
- 줄이 사실인지 기계가 검증하는 일. 이 검사는 줄이 있고 읽히고 실패(`exit` 0 이 아님)를 말하는지만 봐요. 사실 확인은 감사의 몫이에요.
- 기존 테스트의 `touch` 를 새 도우미로 바꾸는 일. 도우미와 눈먼 경우를 재현하는 테스트까지만 하고, 옮기는 일은 각 테스트의 주인 변경이 맡아요.
- `bundle check` 의 기존 `findings` 코드 변경.

## Requirements

- REQ-1: The READY guidance of the flow reference shall require in the body of the READY note one line per new guard test, either `guard-mutant: <test> | <code removed or flipped> | exit <n>` with the non-zero exit code of the failing test, or `guard-mutant: none | <reason>` when the change adds no guard test · failure: none — a guidance text is read by the check of AC-1
- REQ-2: The self-check guidance for guards shall name the `guard-mutant:` line as the record of each mutation and a mutant that survives (`exit 0`) as a missing test · failure: none — a guidance text is read by the check of AC-2
- REQ-3: When `moai bundle check` judges a head, it shall read the body after the first line of the newest READY note that names the head and list the warning `mutant_line_missing` when the body holds no `guard-mutant:` line
- REQ-4 (REQ-3): IF a `guard-mutant:` line fits neither form of REQ-1, THEN `moai bundle check` shall list the warning `mutant_line_unreadable` that names the line number
- REQ-5 (REQ-3): IF a `guard-mutant:` line states `exit 0`, THEN `moai bundle check` shall list the warning `mutant_survived` that names the test
- REQ-6 (REQ-3): IF no READY note names the head, THEN `moai bundle check` shall list the warning `mutant_line_missing` with the detail that no note names the head
- REQ-7: The warnings of `moai bundle check` shall be listed under a `warnings` key and leave `ok`, `findings` and the exit status as they are · failure: none — an unchanged value is the check of AC-5
- REQ-8: The test helper `write_marking_executable` in `tests/common` shall write a fake executable whose marker is created by the shell's own redirection, with no external command and no PATH · failure: none — a missing marker is the failing test
- REQ-9: A test shall reproduce the blind case: a fake executable that creates its marker with `touch` under a PATH that holds no `touch` leaves no marker while it ran, and the helper's fake leaves one · failure: none — the test is its own proof
- REQ-10: The rule provenance table of `CLAUDE.md` shall hold one row that names the recurring cause and the date · failure: none — a missing row is read by the check of AC-8

## Acceptance

- AC-1 (REQ-1): 흐름 문서가 READY 노트의 `guard-mutant:` 두 형식과 "새 지킴 테스트마다 한 줄"을 적어요 · verify: `rg -c "guard-mutant:" skills/moai/references/flows/next-flow.md`
- AC-2 (REQ-2): 자체 점검 문서의 지킴 항목이 그 줄과 "살아남은 뮤턴트는 빠진 테스트"를 적어요 · verify: `rg -c "guard-mutant:" skills/moai/references/internal/self-check.md`
- AC-3 (REQ-3, REQ-4, REQ-5, REQ-6): 줄 읽기가 두 형식을 읽고 줄 없음·못 읽는 줄·`exit 0`·노트 없음을 각각 구분해요 · verify: `cargo test --offline --locked --lib bundle::mutant_line`
- AC-4 (REQ-3, REQ-4, REQ-5, REQ-6): `bundle check` 가 네 경고를 코드와 함께 내고, 줄이 제대로 있으면 경고가 없어요 · verify: `cargo test --offline --locked --test bundle_check_cli mutant_line`
- AC-5 (REQ-7): 경고가 있어도 `ok` 는 참이고 `findings` 는 비었고 종료 코드는 0 이에요 · verify: `cargo test --offline --locked --test bundle_check_cli mutant_line`
- AC-6 (REQ-8, REQ-9): 도우미의 가짜는 PATH 가 비어도 표식을 남기고, `touch` 가짜는 같은 PATH 에서 실행돼도 표식이 없어요 · verify: `cargo test --offline --locked --test fake_exe_marker`
- AC-7 (REQ-3, REQ-4, REQ-5, REQ-6, REQ-8): 새 지킴 테스트마다 지키는 코드를 뺀 뮤턴트가 의도한 테스트에서 실패해요. Result 의 뮤턴트 표에 적어요 · verify: Result 절의 뮤턴트 표.
- AC-8 (REQ-10): `CLAUDE.md` 의 규칙 출처 표에 원인 줄이 있어요 · verify: `rg -c "guard-mutant" CLAUDE.md`

## Scope

- write: `.moai/specs/SPEC-20261123-ready-mutant-line/**`, `skills/moai/references/flows/next-flow.md`, `skills/moai/references/internal/self-check.md`, `skills/moai/references/source-manifest.json`, `src/bundle.rs`, `src/bundle/mutant_line.rs`(새), `src/ops/bundle.rs`(도움말), `docs/design-review/appendix/operations/bundle.md`, `tests/bundle_check_cli.rs`, `tests/fake_exe_marker.rs`(새), `tests/common/fake_exe.rs`(새), `tests/common/mod.rs`(모듈 한 줄), `CLAUDE.md`
- read-only: `src/gate.rs`, `tests/common/**` 의 나머지, 기존 `touch` 가짜를 쓰는 테스트들
- forbidden: 실제 HOME, 새 파이썬 시험, 기존 테스트의 `touch` 교체

## Plan

1. RED: `src/bundle/mutant_line.rs` 의 줄 읽기 단위 테스트와 `tests/bundle_check_cli.rs` 의 경고 테스트, `tests/fake_exe_marker.rs` 를 먼저 써서 실패를 봐요.
2. GREEN: 줄 읽기(`mutant_line`)를 쓰고 `judge()` 가 경고를 모아 `check` 의 `warnings` 로 내요. 도우미 `write_marking_executable` 을 `tests/common/fake_exe.rs` 에 둬요.
3. 안내: 흐름 문서와 자체 점검 문서, 도움말(`src/ops/bundle.rs`)과 `bundle.md` 를 맞추고 매니페스트를 다시 써요.
4. 뮤턴트: 새 지킴 테스트마다 지키는 코드를 뺀 복사본을 돌려 Result 의 표에 적어요(이 SPEC 이 요구하는 줄의 첫 사용).
5. 규칙 출처 표에 원인 줄을 더해요.

## Risks

- 줄 형식이 노트 글에서 우연히 맞거나 틀릴 수 있어요. 줄은 `guard-mutant:` 로 시작할 때만 읽고(앞의 공백과 `- ` 는 허용), 다른 줄은 보지 않아요.
- 경고가 `ok` 를 바꾸지 않아서 레인이 무시할 수 있어요. 관찰이 끝난 뒤 막는 쪽으로 바꾸는 결정이 D-3 이에요.
- 시험 이름에 `|` 가 들어가는 일은 없지만, 들어가면 줄이 읽히지 않는 줄로 경고돼요.

## Decisions

결정은 되돌릴 수 있고 범주는 `none` 이에요. 결정자는 이 SPEC 의 리더이고, 답이 없으면 기본값으로 진행해요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 줄은 `guard-mutant:` 로 시작하고 필드를 세로 막대로 나눠요: REQ-1 의 두 모양(세 칸 또는 `none` 과 이유 두 칸). 기본값: 이 모양. | 노트 글에서 한 줄로 읽히고 구분자가 시험 이름에 안 나와요. | 문서 한 곳과 줄 읽기 한 함수만 바꾸면 돼요. |
| D-2 | 경고는 `findings` 와 따로 `warnings` 에 담아요. 기본값: 따로. | 기존 `ok`·`findings` 를 읽는 시험과 리더의 읽기를 안 깨요. | 코드 한 곳에서 합쳐요. |
| D-3 | 막는 쪽으로의 전환은 관찰 뒤 기록하는 결정이에요. 관찰 창: 이 검사가 들어간 뒤 `bundle check` 가 도는 번들 열 번 또는 다음 릴리스 중 먼저 오는 것. Result 에 경고 횟수와 거짓 경고를 적어요. 기본값: 관찰 중에는 막지 않음. | `internal/sdd` 11번. | 해당 없음(막지 않음). |
| D-4 | `none` 줄이 새 시험을 더하는 변경에 붙었는지는 기계가 보지 않아요. 기본값: 보지 않고 감사가 봐요. 관찰에서 이 빈틈으로 빠지는 일이 보이면 후속으로 다뤄요. | 규칙 20: 있을 법한 먼 일은 설계하지 않아요. | 해당 없음. |

## Result
