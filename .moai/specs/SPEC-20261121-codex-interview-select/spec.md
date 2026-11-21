---
id: SPEC-20261121-codex-interview-select
title: "Codex interview: the Default-mode question-tool limit, documented with the 2026-11-21 sources"
status: draft
revision: 1
created: 2026-11-21
---

# Codex interview: the Default-mode question-tool limit, documented with the 2026-11-21 sources

## Intent

설치 흐름 보고서(2026-11-21 §7)의 관찰: `moai codex` 로 연 세션은 Default 모드라서, 인터뷰 선택지가
실제 선택 화면이 아니라 평문 번호 목록으로 와요. 사용자가 HOW 를 골랐어요(리더 결정 노트, 2026-11-20):
현재 Codex 공식 문서로 Default 모드의 `request_user_input`(또는 같은 기능) 지원을 확인하고, 지원되면
인터뷰가 실제 선택 화면을 쓰게 하고, 아니면 한계를 문서에 적어요. 2026-11-21 재확인 결과 지원은 여전히
실험 플래그 뒤에 있어요(개발 중, 기본 꺼짐, 공식 문서 플래그 표에 없음, `moai codex` 런처가 켜지 않음).
그래서 이 SPEC 은 한계 문서화예요: `questions.md` 의 Codex 행을 오늘 근거로 고치고(플래그 상태, 플래그
없음 때의 거절, 플래그 켬 때의 자동 해소, 인터뷰 재질문 규칙), 그 문구를 `tests/update_route.rs` 핀으로
고정해요.

## Non-goals

- `moai codex` 런처가 `features.default_mode_request_user_input` 을 켜게 하는 일: 이 SPEC 밖이에요.
  이전 조사 카드(2026-11-18 마감)의 마지막 노트 2번 항목이 그 일에 카드가 없다고 적었어요 — 리더가 새
  카드로 만들어요. 그 카드가 착지하고 Codex 가 플래그를 안정 버전으로 올리면 이 행을 다시 고쳐요.
- 사용자 계정으로 모델을 부르는 Codex 실제 대화 실행: 이전 조사의 `not_run` 이유(외부 전송과 사람의
  선택이 필요)가 그대로 유효해요.
- Plan 모드 행동 문구와 case-(2) 평문 규칙의 변경: 오늘 읽음으로도 두 문장은 그대로 맞아요.
- `tests/codex_question_tool.rs` 와 이전 조사 SPEC 의 연구 글: 고치지 않아요. 이번 행 문구가
  `off by default` 와 플래그 이름을 유지하므로 그 시험은 그대로 통과해요.

## Requirements

- REQ-1: The Codex CLI row of the questions reference shall state that Default mode lists `request_user_input` only when `features.default_mode_request_user_input` is on, that the flag is under development and off by default, and that the `moai codex` launcher does not turn it on in this version.
- REQ-2: The row shall state the two Default-mode limits with their upstream sources and the read date: a call without the flag is refused with `request_user_input is unavailable in Default mode` (openai/codex issue #29104), and a flag-on Default-mode request that goes unanswered resolves on its own after about a minute (openai/codex issue #37472).
- REQ-3: The row shall apply the no-answer rule to the interview: a returned no-answer is not a pick, a case-(2) question never goes on by best judgment (rule 3), and an interview question is re-asked in plain text, one per message, and waited for.
- REQ-4: The row shall cite the installed Codex 0.171.0 (read from its binary, 2026-11-21) for the Default-mode instruction that bans permission requests through the tool.
- REQ-5: The research note of this SPEC shall record the 2026-11-21 re-check: the official config reference (the flag absent from its feature-flags table), the installed 0.171.0 measurements (`codex features list`, the permission-ban sentence), the rust-v0.171.0 source facts with sha256 (stage under development, default false, Plan-only gating, the tool gate), and the open upstream issues with their read times.

## Acceptance

- AC-1 (REQ-1…REQ-4): the Codex row carries the updated availability sentence without the stale
  launcher-will-turn-it-on claim, the two limits with citations, the interview re-ask rule, and the
  0.171.0 citation.
  verify: `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool`
- AC-2 (REQ-5): `research.md` of this SPEC records the re-check with URL, read time and sha256 for
  every fetched source.
  verify: `cargo test --locked --offline --test codex_question_tool research_note_agrees_with_the_hosts_rows_of_questions_md` (the row and the earlier research still agree on the tool name and the off default)
- AC-3: 이웃 검사.
  verify: `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test question_bank_hosts`; `cargo fmt --check`; `python3 tools/source_manifest.py check`

## Scope

- write: `skills/moai/references/internal/questions.md`, `tests/update_route.rs`,
  `.moai/specs/SPEC-20261121-codex-interview-select/spec.md`,
  `.moai/specs/SPEC-20261121-codex-interview-select/research.md`,
  `skills/moai/references/source-manifest.json` (해시 재생성)
- read-only: 설치된 `codex` 바이너리와 공식 문서, openai/codex 이슈, `tests/codex_question_tool.rs`
- forbidden: 실제 HOME 과 Codex 설정·인증 파일의 쓰기, 모델을 부르는 Codex 실행, 런처 코드 변경

## Plan

1. RED: `tests/update_route.rs` 의 Codex 행 핀을 새 문구로 바꾸고(런처 문장 교체, 한계 두 문장 추가,
   인터뷰 재질문 문장 추가, `card C-52` 부정 핀) 실행해 실패를 읽어요.
2. GREEN: `questions.md` 의 Codex 행을 고쳐 통과시켜요.
3. 매니페스트를 재생성하고(`tools/source_manifest.py write`) 이웃 검사와 스캔 시험을 돌려요.

## Risks

- 이슈 본문과 문서 페이지는 언제든 바뀌어요 → 읽은 시각을 적고, 태그 소스는 sha256 으로 굳혀요.
- Codex 가 플래그를 안정 단계로 올리면 문장이 낡아요 → 측정에 버전(0.171.0)과 시각을 적어 두어 낡았는지
  알 수 있게 해요.
- 행의 날짜 문구가 핀과 어긋나면 시험이 실패해요 → 핀은 문장만 묶고 읽은 날짜 문자열은 묶지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 실제 선택 화면 적용(런처가 플래그를 켜기)은 하지 않고 한계 문서화를 해요 | 사용자 선택의 "안 되면" 갈래: 지원이 실험 플래그 뒤(개발 중, 기본 꺼짐, 공식 문서 미기재, 런처 미설정)이고 플래그를 켜도 무응답 자동 해소가 있어요(이슈 #37472) | 리더가 런처 카드를 만들고 Codex 가 플래그를 안정화하면 행을 다시 고쳐요 |
| D-2 | 행의 "card C-52" 귀속 문구를 없애요 | 그 문장이 가리킨 일(런처가 플래그를 켜기)은 카드가 없어요(2026-11-18 마감 노트 2번 항목); 잘못된 귀속은 낡은 문서보다 나빠요 | 없음 |
| D-3 | 행의 인터뷰 재질문 규칙은 case-(2) 규칙과 분리한 한 문장으로 적어요 | 인터뷰 답은 사용자의 선택이라 무응답 자동 해소가 설정을 조용히 정하면 안 돼요; 되돌릴 수 있는 일반 선택은 기존 규칙(도구 유지)을 그대로 둬요 | 후속 카드가 런처 플래그를 켜고 자동 해소가 사라지면 문장을 빼요 |

## Result

- AC-1: met — RED 먼저: `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool` 1 failed, exit 101(새 가용성 핀에서 — 행에 옛 런처 문장 `is to turn it on — card C-52`). GREEN 뒤 같은 시험 ok이고 `--test update_route` 전체 9 passed, exit 0. 행은 새 가용성 문장(런처가 이 판에서 켜지 않음), 한계 두 문장(#29104 거절·#37472 자동 해소, 읽은 날짜), 인터뷰 재질문 문장, 0.171.0 인용을 가지고 `card C-52` 귀속은 없어요.
- AC-2: met — `research.md`가 2026-11-21 재확인을 기록했어요: 공식 config 문서 페이지(sha256 f356e0e77d02370adb1991c9117a76ebe90635a6148663c7793b574dbbf8ed13, 440814 B, 플래그 0회), `codex features list`(0.171.0, `under development false`), rust-v0.171.0 소스 3개 sha256(features-lib 5b191163…, config-types ce9b2622… — 0.170.0과 바이트 동일, tool-config 19259689… — 동일), 이슈 #29104·#37472 읽은 시각. `cargo test --locked --offline --test codex_question_tool` 4 passed(행과 이전 조사 글이 도구 이름·기본 꺼짐에서 같음).
- AC-3: met — `cargo test --locked --offline --test update_route --test codex_question_tool --test question_bank_hosts` 9·4·1 passed; `cargo fmt --check` exit 0; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0; `python3 tools/source_manifest.py write` 후 `check` ok(바뀐 항목 questions.md 1개); `tools/self_check.sh --pre-ready`(2026-11-21T02:49Z, head 041536fe07 위 미커밋 트리)에서 diff-check·fmt·fmt-ops·manifest·py-source-manifest·py-repo-scans·skill-package exit 0, spec-check 항목 exit 1 — 사유는 이 트리가 원래 가진 낡은 SPEC 잔여(closure_status 29개)와 이 SPEC의 Result 미작성(closure_result)이었고, Result 작성 뒤 이 SPEC은 closure_status 잔여만 남겨요(아래 줄).
- 남은 것: 이 SPEC의 closure_status 줄은 승인·마감 절차의 잔여로 남아요(SPEC은 draft — 승인은 사용자의 선택). G14의 행동 재현(평문 대체)은 이 문서화만으로 없어지지 않아요 — 런처가 플래그를 켜는 후속 카드가 필요해요(카드 기준 1의 인수와 연결).
