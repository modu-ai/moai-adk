---
id: SPEC-20261120-lane-limit-resume
title: "A stalled lane gets one resume line at its usage limit's reset instant"
status: draft
revision: 1
created: 2026-11-20
parent: SPEC-20261119-lane-screen-scan
---

# A stalled lane gets one resume line at its usage limit's reset instant

## Intent

2026-11-20 세션 병목 리뷰(~/.moai/reports/moai-adk/session-bottleneck-review-2026-11-20.md, P7): 네 개의 glm 레인이 "Usage limit reached for 5 hour. Your limit will reset at <시각>"(표기는 UTC+8) 앞에서 35분가량 멈춰 있었고, 사용자가 말해야 움직였어요. `moai lane-screen --scan` 은 이미 멈춤과 해제 시각(UTC)을 읽어 주지만, 그 시각이 지나도 아무도 레인에 한 줄을 넣어 주지 않으면 레인은 계속 서 있어요. 이 SPEC 은 `moai lane-resume` 을 만들어요: 레인 화면을 읽어 한도 멈춤을 알아보고, 해제 순간을 UTC 로 계산해, 그 순간이 지났을 때 레인에게 이어서 한 줄만 보내요. 보내기는 `lane-send` 의 검사 경로(확인 창 거절, 두 번 읽기, 글·핸들 검사)를 따르고, 확인 창에는 절대 넣지 않아요. 모든 MoAI 사용자의 일반 기능이에요.

## Interview

- Q1: 누구에게 보내나요? → A: 기본은 스캔에서 찾은 한도 멈춤 레인 전부, `--terminal` 로 하나만 좁혀요.
- Q2: 보내는 시점은요? → A: 그 레인의 해제 순간이 지난 레인만이요. 아직이면 이름과 UTC 순간을 알려 주고 아무것도 보내지 않아요. `--wait` 은 가장 이른 미래 순간까지 한 번 잠들고 화면을 다시 읽어서 보내요(잠음 상한 8시간, 잠들기는 한 번뿐이에요).
- Q3: `lane-send` 와 뭐가 다른가요? → A: `lane-send` 는 `usage_limit` 은 언제나 거절해요. `lane-resume` 의 일은 바로 그 상태의 "해제 순간이 지난" 경우예요. 확인 창 거절, 보내기 직전의 두 번 읽기, 글·핸들 검사는 같은 규칙을 그대로 적용해요.
- Q4: 두 번 보내기는 어떻게 막나요? → A: 한 번의 실행에서 멈춘 레인마다 최대 한 줄이에요. 상태 파일은 두지 않아요: 보낸 줄로 레인이 깨어나면 화면이 다시 그려져 다음 실행은 멈춤을 찾지 못해요. 문서(operations 조각)에 실행당 한 줄 규칙을 적어요.
- Q5: 화면이 이미 회복됐으면요? → A: 그 레인을 새 상태 이름으로 알리고 아무것도 보내지 않아요.
- Q6: 확인 창이면요? → A: 그 레인은 refused 로 알리고 보내지 않고, 다른 레인은 계속해요.
- Q7: 해제 시각을 못 읽으면요? → A: 그 레인은 skipped(이유) 로 적고 보내지 않아요.
- Q8: 종료 코드는요? → A: 한 줄이라도 보냈으면 0, 아무것도 보내지 않았으면 0 이 아니에요(레인마다 결과는 봉투에 다 있어요).

## Non-goals

- 상태 파일이나 실행 밖의 중복 방지 장부: 기억은 두지 않아요.
- 확인 창에 대신 답하기, 보낸 줄이 무시됐을 때 다시 보내기.
- 한도 자체를 풀거나 되풀이 잠들기: 반복은 레인의 깨우기 주기가 다시 실행하는 몫이에요.
- 한도가 아닌 다른 멈춤(예: 권한 요청)의 해제.

## Requirements

- REQ-1: When `moai lane-resume` runs, the command shall list every lane it found stopped by a usage limit with that limit's reset instant in UTC.
- REQ-2 (REQ-1): IF a found lane's reset instant has passed, THEN `moai lane-resume` shall send the given resume line to that lane once and no more than once per run.
- REQ-3 (REQ-2): IF a lane's screen is a permission prompt or a choice dialog, THEN `moai lane-resume` shall send nothing to that lane and name it as refused.
- REQ-4 (REQ-2): IF a lane's screen no longer shows the usage limit at the second read just before the send, THEN `moai lane-resume` shall send nothing to that lane and name its new state.
- REQ-5: The command `moai lane-resume` shall run the text and handle checks of `lane-send` before any `orca` call is made.
- REQ-6 (REQ-1): IF a lane's reset instant is still in the future and `--wait` is not given, THEN `moai lane-resume` shall send nothing to that lane and name it as waiting with its reset instant.
- REQ-7 (REQ-6): IF `--wait` is given and the earliest future reset instant is within the sleep cap, THEN `moai lane-resume` shall sleep once to that instant, read the lanes again, and send to each lane whose reset instant has passed.
- REQ-8 (REQ-7): IF `--wait` is given and the earliest future reset instant is beyond the sleep cap, THEN `moai lane-resume` shall sleep for no time and name that lane as waiting.
- REQ-9: The command `moai lane-resume` shall consider exactly the terminal named by `--terminal` when the flag is given and every terminal of the scan otherwise.
- REQ-10 (REQ-1): IF a lane cannot be read or its reset instant cannot be told, THEN `moai lane-resume` shall skip that lane with the reason and continue with the others.
- REQ-11: The command `moai lane-resume` shall exit 0 when at least one line was sent and non-zero when none was sent.
- REQ-12: The command `moai lane-resume` shall send no input to any terminal except at most one resume line per stalled lane of the same run.

## Acceptance

- AC-1 (REQ-1): A fixture screen with the 429 line yields the reset instant in UTC. 가짜 `orca` 의 한도 화면("Your limit will reset at 2026-11-20 18:23 (UTC+8)")에 대해 봉투가 `2026-11-20T10:23:00Z` 를 이름과 함께 알려요.
  verify: `cargo test --locked --offline --test lane_resume_cli`
- AC-2 (REQ-2): the resume send happens once after it. 해제 순간이 지난 화면에서 `sent.txt` 에 정확히 한 줄이 있고, 봉투는 `sent: true` 예요.
  verify: `cargo test --locked --offline --test lane_resume_cli resume_sends_once`
- AC-3 (REQ-3): and never into a prompt. 확인 창 화면에서 `sent.txt` 가 없고 봉투 오류가 `permission_prompt` 를 이름으로 알려요.
  verify: `cargo test --locked --offline --test lane_resume_cli a_prompt_never_receives_the_text`
- AC-4 (REQ-4): 보내기 직전의 두 번째 읽기에서 화면이 회복됐으면 보내지 않고 새 상태를 알려요.
  verify: `cargo test --locked --offline --test lane_resume_cli the_screen_is_read_again_just_before_the_send`
- AC-5 (REQ-5, REQ-12): `-` 로 시작하는 글은 `orca` 를 한 번도 부르지 않고 거절돼요.
  verify: `cargo test --locked --offline --test lane_resume_cli the_text_is_checked_before_orca_runs`
- AC-6 (REQ-6, REQ-7, REQ-8): 미래 순간은 waiting 으로 이름과 UTC 순간이 오고, `--wait` 은 그 순간까지 정확히 한 번 잠들고 한 번 보내며, 상한을 넘으면 잠들지 않고 waiting 이에요.
  verify: `cargo test --locked --offline --test lane_resume_cli wait`
- AC-7 (REQ-9, REQ-10, REQ-11): `--terminal` 로 하나만, 읽지 못한 레인은 skipped(이유), 아무것도 못 보내면 종료 코드 0 이 아니에요.
  verify: `cargo test --locked --offline --test lane_resume_cli`

## Decisions

- D-1: 새 연산은 파일 하나(`src/ops/lane_resume.rs`)로 등록돼요(#2617 방식, `build.rs` 의 ops 등록부). 판정 재사용은 읽기뿐이에요: `moai_cli::lane_status::lane_screen` 의 `observe`·`reset_of`·`scan` 과 `moai_cli::lane_status::lane_send` 의 `check_handle`·`check_text`·`read_checked`·`Orca`·`ProcessOrca` 를 부르고, 그 파일들은 고치지 않아요.
- D-2: 무상태예요. 실행 밖의 기억 없이, 한 실행에서 멈춘 레인마다 최대 한 줄이에요. 깨어난 레인의 화면은 다시 그려져 다음 실행이 멈춤을 찾지 못하는 것이 중복 막기의 몫이에요.
- D-3: `lane-send` 의 `usage_limit` 거절은 의도적으로 재사용하지 않아요. resume 의 일은 바로 그 상태의 해제-순간-지남 경우라서, 보내기 판정은 이 파일이 같은 공개 조각으로 다시 조립해요(`look`·`refusal`·`read_send_answer` 는 비공개라 작은 재현이 필요해요; `ok:true` 계약은 조각 문서에 핀을 남겨요).
- D-4: `--wait` 의 잠들기는 한 번이고 상한은 8시간(5시간 창 + 여유)이에요. 상한 넘으면 잠들지 않고 waiting 으로 돌아가 호출자(레인의 깨우기 주기)가 다시 오게 해요.
- D-5: 문서 조각 `docs/design-review/appendix/operations/lane_resume.md`, `src/help.rs` 의 Lane and leader 그룹과 예시, 도움말 골든(재생성 후 diff 리뷰)이 이 명령의 사람용 얼굴이에요.

## Scope

- write: `.moai/specs/SPEC-20261120-lane-limit-resume/**`
- write(구현): `src/ops/lane_resume.rs`, `docs/design-review/appendix/operations/lane_resume.md`, `src/help.rs`(그룹·예시), `tests/fixtures/help/**`(골든 재생성), `tests/lane_resume_cli.rs`
- read-only: 그 밖의 모든 것(`src/lane_screen.rs`, `src/lane_send.rs`, 두 파일의 시험과 문서 조각, 다른 operations 포함 — 부르기만 하고 고치지 않아요)

## Plan

1. SPEC 작성과 `spec check`(이 문서).
2. `src/help.rs` 그룹·예시 추가, 도움말 골든 8파일 재생성(`MOAI_HELP_GOLDEN_WRITE=1 cargo test --test help_layout`) 후 diff 리뷰.
3. `src/ops/lane_resume.rs`: 핵심 흐름(스캔/한 터미널 → 멈춤·순간 → 기한 순간 판정 → 잠들기(--wait, 한 번) → 두 번 읽기 → 보내기 1회)과 파일 안 단위 시험(가짜 `Orca`, 기록하는 잠들기).
4. `docs/design-review/appendix/operations/lane_resume.md` 문서 조각.
5. `tests/lane_resume_cli.rs`: 가짜 `orca` 스크립트(`MOAI_ORCA_BIN`, `lane_send_cli.rs` 와 같은 틀)와 화면별 시험(RED 먼저).
6. 검사: `git diff --check`, `cargo fmt --check`, src/ops rustfmt(2021), `clippy --all-targets -- -D warnings`, 대상 시험(lane_resume_cli, help_layout, lane_ 접두, --lib, --bins), python 영역 모듈, trailer-check. 종료 코드를 READY 노트에 기재.

## Risks

- R1: `--terminal` 경로와 스캔 경로의 행동 어긋남 → 한 터미널은 자체 읽기 경로, 스캔은 `screen::scan` 재사용으로 두 경로를 나눠 두고 각자 시험해요.
- R2: 도움말 골든 재생성의 기계적 실수 → 골든 도구로만 재생성하고 diff 를 전수 리뷰해요(상태줄 골든의 교훈).
- R3: 해제 시각의 시간대 해석은 `reset_of` 기존 검사에 맡기고, 제 시험은 UTC+8 픽스처 한 줄로 핀만 남겨요.
- R4: 문서 조각이 자동 검사에 없어 감사에서 문서-구현 어긋남이 나올 수 있어요 → 구현 도움말 문장과 같은 문구를 조각에 써요.
