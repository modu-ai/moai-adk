---
id: SPEC-20261122-star-and-reports
title: "A polite star line at install and update, and improvement reports that leave only on the user's yes under the user's own gh account"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# Star ask and consented reports: nothing leaves the machine without a yes

## Intent

사용자 요청(2026-11-22)과 보낼 방식에 대한 사용자 선택이 이 카드의 규칙이에요. (1) 스타 요청: 설치와 업데이트에서, 사용자가 저장소를 스타하지 않았으면 정중한 한 줄을 사용자 언어로 보여요 — 별이 프로젝트에 실제 힘이라는 뜻과 저장소 링크를 담고요. 사용자를 대신해 스타하는 일은 절대 없어요. 스타 여부는 gh 가 있고 로그인돼 있을 때만 읽기(GET)로 확인해요(응답 204면 스타, 404면 아니에요 — GitHub REST starring 문서); gh 가 없으면 확인 없이 줄만 보여요. 줄은 설치·업데이트마다 최대 한 번이고, 사용자가 스타했다고 하거나 묻지 말아달라 하면 홈 moai 상태에 기억해 다시는 안 물어요. (2) 개선 보고: 사용자는 '사용자 자기 계정으로(익명 아님)'을 골랐어요 — 익명 중계가 없고 조용한 자동 전송이 없어요(이슈에 사용자 이름이 실려요). 설치 때 한 번 물어요: '그때마다 물어보기'(권장) 아니면 '안 보내기'; 선택은 저장되고 명령으로 바꿔요. 보고를 내보낼 때는 사용자가 정확한 제목과 본문을 먼저 봐요(코드 없음, moai 버전과 OS 외 파일 경로 없음, 프로젝트 이름 없음, 비밀 화면 통과), gh 로 사용자 계정에서 예(yes)일 때만 만들어요(POST 이슈 생성); gh 가 없으면 미리 채운 이슈 페이지 URL 을 보여요. 외부 전송은 매번 사용자 선택이에요. 이 조각은 SPEC 만 써요.

읽은 문서(2026-11-22T03:4x–5xZ): GitHub REST starring 문서(GET /user/starred/{owner}/{repo} — 204 스타, 404 아니오, 스코프 없이 읽힘; PUT/DELETE 는 이 카드가 쓰지 않아요), issues 문서(POST /repos/{owner}/{repo}/issues — 제목 필수, 본문 선택, 201 생성, 빠른 연속 생성은 2차 속도 제한), gh api 설명서(GET 기본, --method 로 동사, --silent, 자격 증명은 gh auth login 상태; 나가는 코드는 별도 exit-codes 문서).

### What exists (tip `bad63856f6`)

- gh 호출의 현 precedent: 상태줄 카운트가 gh 를 부르고 그 결과를 `.moai/state/github/counts.json` 에 쌓아요(`src/statusline/counts.rs`); 실행기·운송 계층에도 gh 호출이 있어요(`src/forge/transport.rs`, `src/redact.rs`).
- 비밀 모양 판별이 이미 있어요: 홈 상태의 `is_secret`(`src/home_state.rs:82`)과 원장 노트 검증의 비밀 거부가 같은 규칙의 두 얼굴이에요 — 보고 본문의 검사도 이 규칙을 써요.
- 홈 moai 상태: `src/home_state.rs` 의 계획·적용 구조가 있고, 설치·업데이트의 끝맺음 출력이 줄을 보여줄 자리예요.
- 스타 요청과 동의 저장, 보고 미리보기·전송은 새로 만들어요.

## Non-goals

- 이 조각의 코드·시험: SPEC 만 써요.
- 사용자를 대신한 스타(PUT)나 익명 중계 계정: 어떤 조각에서도 하지 않아요.
- 원격 검사·CI 의 GitHub 쓰기, 티켓 자동 분류 같은 다른 GitHub 기능.
- 보고 본문의 자동 수집 확대(파일 경로·코드·프로젝트 이름은 영구 금지).

## Requirements

- REQ-1: At install and at update moai shall show one polite line in the user's language asking for a star with the repository link when the star state is unknown or not starred · failure: none — a run whose gh answers starred (204) shows nothing; at most one such line per run; a dismissed ask shows nothing forever.
- REQ-2: The star-state check shall make exactly one read-only GET through the user's gh login and no write call in any path · failure: none — a check that fails or refuses answers unknown and the line still shows.
- REQ-3: A dismissal (the user starred or asked not to be asked) shall be recorded in the home moai state and suppress the line from then on · failure: none — a state write that fails leaves the line showing and never blocks the run.
- REQ-4: At install moai shall ask once how improvement reports are sent — offering ask-me-each-time (recommended) and never — with the stored answer changeable by one command · failure: none — an unanswered ask sends nothing and stores nothing.
- REQ-5: A report shall leave the machine only on the user's yes to its exact full text — shown before the send and carrying no code, no file path beyond the moai version and the OS, no project name and nothing the secret rule blocks · failure: none — the never choice and a no both send nothing and keep the text on the machine.
- REQ-6: A consented report shall be created with gh under the user's account through the repository issue-creation POST with the prefilled issue page URL shown instead of sending when gh is absent · failure: none — a failed send surfaces the error with the exact text preserved for a retry.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): with a stand-in gh answering 204 the line never shows and no write call is recorded; answering 404 the line shows exactly once per run; a dismissed state shows nothing · verify: `cargo test --locked --offline --test star_ask_cli ac1_`
- AC-2 (REQ-4, REQ-5): in a temp home the install ask stores the choice; a report leaves only on a yes to the exact shown text; the never choice sends nothing; a key-shaped value in the body is refused by the secret rule · verify: `cargo test --locked --offline --test report_ask_cli ac2_`
- AC-3 (REQ-6): with the stand-in gh the issue post happens only after the yes and carries exactly the shown title and body; without gh the prefilled URL is shown and nothing is sent · verify: `cargo test --locked --offline --test report_ask_cli ac3_`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-star-and-reports/**`
- later slices, write: the install and update endings (`src/update.rs`, `src/launcher_init.rs`, `src/home_install_apply.rs`), the ask and report module (new), the home-state entries (`src/home_state.rs`), the stand-in gh test harness (`tests/star_ask_cli.rs`, `tests/report_ask_cli.rs`)
- forbidden in every slice: a star on the user's behalf (any PUT to the starred endpoint), an anonymous relay, sending any report without a fresh yes, real network calls in tests

## Plan

1. (done) SPEC (this slice).
2. Star ask: the one-GET check with the stand-in gh, the polite line in the user language, the dismissal state; AC-1.
3. Report ask: the install question and its stored choice, the exact-text preview with the secret rule, the gh send on yes and the prefilled URL without gh; AC-2 and AC-3.

## Risks

- The star line rides the install and update endings, which other cards are reshaping (the update-preview route); the slice keeps to one printed line so the surface stays mergeable.
- The stand-in gh must record its calls so the no-write proof is mechanical; a stand-in that swallows calls proves nothing.
- Secondary rate limiting on issue creation is documented; a send that is refused surfaces the error and keeps the text (REQ-6), never a silent retry.
- The user language of the line and the ask comes from the project's language setting, like every shipped text.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The star check is GET-only and only when gh is present and logged in; otherwise the line shows unchecked. | The card body; the endpoint needs no scopes for the authenticated read. | Skip the line without gh. |
| D-2 | Never star on the user's behalf. | The card body; a star is the user's act. | Ask once for consent. |
| D-3 | The report route is the user's own gh account; no anonymous relay exists in the product. | The user's pick recorded on the card. | Add a relay. |
| D-4 | The dismissal and the report choice live in the home moai state. | The card body names the home moai state; the state module already plans entries. | Move them into a profile. |
| D-5 | The preview runs the secret rule and the no-code, no-path, no-project-name limits before any yes is possible. | The card body; the check must precede the choice, not follow it. | Check after the yes. |

## Result

(Filled when the work is done. This is the SPEC-only slice: no code, no acceptance criterion run yet; the star endpoint facts, the issue endpoint facts and the gh manual behaviors are recorded above.)
