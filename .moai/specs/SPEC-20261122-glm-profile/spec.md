---
id: SPEC-20261122-glm-profile
title: "The GLM profile: gateway env in the shared config root's user settings, the key only through the user's helper, the real window, the not-available list"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# GLM profile: gateway env in the config folder's user settings, never a key value

## Intent

GLM 레인은 지금 게이트웨이 값을 실행기 환경으로만 받아요(`src/session.rs:90` 의 짝 목록, `src/launcher_terminal.rs:251` 의 env_keys). 공식 문서는 셸로 내보낸 게이트웨이 값이 "doesn't reliably reach background agents hosted by the supervisor" 이라고 적고, 해법으로 "Use a settings file for any gateway that background agents must always route through" 를 줘요(llm-gateway-connect.md, 2026-11-22T02:4xZ 전문 읽음). 또 모르는 모델 id 는 v2.1.323 부터 1M 컨텍스트로 가정해요(llm-gateway-protocol.md § Settings for unrecognized model IDs) — GLM 별칭의 실제 창을 선언하지 않으면 압축 시점이 틀려요. 이 SPEC 이 프로필을 만들어요. 사용자 결정(2026-11-22, 세션 메시지 카드의 USER DECISIONS): 모든 레인이 사용자 세션과 같은 설정 루트에서 돌고, GLM 레인은 그 루트의 settings env 로 GLM 을 타요 — 그래서 이 프로필의 자리는 흐름이 공유하는 바로 그 설정 루트의 사용자 settings env 블록이에요(적용은 사용자 선택). 키 값은 moai 가 절대 쓰지 않아요: 자리는 사용자의 apiKeyHelper 항목이에요. 베타 끄기는 실제 400 Extra inputs 오류의 수리로만, 안 되는 기능 목록은 프로필 글에, 상태줄 5H·7D 는 GLM 페이로드에서 absent 로 나오는 golden 으로 고정해요. 이 조각은 SPEC 만 쓰고 코드는 없어요(카드 본문: 첫 조각은 SPEC).

공유 루트의 긴장도 명시해요: 설정 루트의 사용자 settings env 는 그 루트의 모든 세션에 걸려요. 게이트웨이 자격 증명 변수는 OAuth 로그인보다 앞서므로(authentication 문서의 우선순위: 클라우드 변수 > 토큰 > 키 > helper > OAuth), 이 블록이 활성인 동안 루트 안의 모든 세션 — 사용자가 말하는 세션도 — 게이트웨이로 가요. 같은 루트에 first-party 세션과 GLM 세션을 섞어 두는 방법(세션별 시작 덮어쓰기, 또는 루트 전체 GLM 전환을 사용자가 받아들이는 것)은 이 SPEC 의 열린 질문이에요.

근거(전부 원문): llm-gateway-connect.md 전문(02:4xZ), llm-gateway-protocol.md 전문(02:4xZ), model-config.md 의 창 선언·별칭 고정·modelOverrides 절(03:0xZ까지, 첫 100k 자와 이어읽기), 리더 전체문서 보고서의 16-13 항과 7-5 표(~/.moai/reports/moai-adk/claude-code-native-redesign-2026-11-22.md), 그리고 같은 날 이 레인(agent-24)이 잰 측정: 상태줄 stdin 의 rate_limits 는 문서상 claude.ai 구독·게이트웨이 spend limit 세션에만 존재(statusline.md)하고 서드파티 엔드포인트(api.z.ai)에서는 absent 가 예상, 모델 카탈로그 캐시에 GLM 항목 없음, 아티팩트 도구는 세션에 제공되지만 게시 경로는 게이트웨이에서 불가(제공됨과 사용가능의 구분), Remote Control 없음.

### What exists (tip `6766721f30`)

- 게이트웨이 값의 지금 길: `src/session.rs:90` 이 (ANTHROPIC_BASE_URL, api.z.ai 경로) 짝을 기본값으로 들고, `src/launcher_terminal.rs:251,265` 가 env_keys 로 터미널에 넘겨요. 셸 내보내기 모양이라 문서의 경고 정확히 해당.
- 키의 지금 길: `glm` 은 ANTHROPIC_AUTH_TOKEN 을 환경에서, 없으면 moai-cli 금고에서 가져오고 값은 호스트 프로세스에만 가요(`tests/launcher_glm_key_file.rs:1-6`). 레거시 키 파일은 읽지 않고 blocked 로 알려요.
- 상태줄 입력은 이미 rate_limits 를 Option 으로 파싱해요(`src/statusline/input.rs:30`, five_hour·seven_day 는 `:143-144`), 있을 때만 상태에 복사돼요(`src/statusline/state.rs:254-256`). absent 렌더의 golden 은 아직 없어요(다섯 시간 창의 렌더 시험은 `tests/usage_limit_cli.rs:92-98`).
- 설정 폴더 주소 계산과 사용자 설정 쓰기는 자리가 있어요: Claude 홈은 항상 CLAUDE_CONFIG_DIR 로 찾는 작업이 진행 중이고(보고서 22번), 일반 내용이 각 설정 폴더에 가는 결정(2026-11-22, 프로젝트 .claude 기본 금지)이 이 SPEC 의 배치 규칙이에요.

## Non-goals

- 이 조각의 코드·시험: SPEC 만 써요.
- 키 값의 저장·출력·대신 쓰기: 어떤 조각에서도 하지 않아요(금지).
- 프로젝트 `.claude/settings.json`·`.claude/settings.local.json` 을 프로필 자리로 쓰는 일: 문서의 경고("Don't put the credential in a project's `.claude/settings.json`. That file is committed and shared")와 프로젝트 .claude 기본 금지 결정 때문에 범위 밖.
- 상태줄 진행 순서 재설계(다른 카드), 원장·흐름 규칙(다른 카드).
- 데스크톱 앱 경로: 데스크톱은 ANTHROPIC_BASE_URL 을 스스로 정해요(desktop.md). GLM 레인은 터미널에서 돌아요.

## Requirements

- REQ-1: When the user picks a GLM profile for the shared config root the flow sessions run under, the generator shall write into that root's user settings exactly one env block holding the gateway base URL, the model alias pins and the real context window declaration, and no credential value · failure: none — a config root without the user's pick is left untouched.
- REQ-2: The generator shall address the credential only as the apiKeyHelper entry the user writes themselves, never writing, printing or storing a credential value · failure: none — a settings file that already holds a raw credential variable is reported as a finding and left unchanged.
- REQ-3: The window declaration shall state the real window of the aliased model with the documented variables (the max-context variable together with the auto-compact window, or an alias mapping when the id maps to a known model), because an unrecognized id is otherwise assumed to have a 1M window · failure: none — an alias that maps to a known Anthropic model id needs no window declaration and gets none.
- REQ-4: The generator shall offer the beta-disable variable only as the repair for an actual 400 Extra inputs error, naming its tool-search side effect in the same text · failure: none — a profile without such an error carries no beta-disable entry.
- REQ-5: The profile text shall carry the not-available list of the gateway route (artifact publishing, claude.ai connectors, Code Mode, advisor, web-fetch subagents, gateway WebSearch, Remote Control, channels, Chrome, voice dictation) with the measured offered-tool versus usable-path distinction · failure: none — a missing item in the list is a documentation defect, not a runtime failure.
- REQ-6: The status line renderer shall show the five-hour and seven-day windows as absent when the payload carries no rate_limits object, pinned by a golden test on a GLM-shaped payload · failure: none — a payload with rate_limits renders exactly as before.
- REQ-7: The generator shall name, as a finding for the user, a leftover credential variable that outranks the helper or that bills unattended headless runs · failure: none — a finding changes nothing itself.

## Acceptance

- AC-1 (REQ-1, REQ-3): a temporary config root with the user's pick receives exactly the env block (base URL, alias pins, window declaration) and nothing more; without the pick nothing is written; an alias that maps to a known model id plans no window declaration · verify: `cargo test --locked --offline --test glm_profile_cli ac1_`
- AC-2 (REQ-2, REQ-7): no path writes a credential value; a settings file with a raw credential variable is a finding left unchanged; a leftover key variable that outranks the helper is named · verify: `cargo test --locked --offline --test glm_profile_cli ac2_`
- AC-3 (REQ-4): the beta-disable entry appears only in the repair text for the named 400 error and that text names the tool-search side effect · verify: `cargo test --locked --offline --test glm_profile_cli ac3_`
- AC-4 (REQ-5): the profile text carries every item of the not-available list and the offered-versus-usable distinction · verify: `cargo test --locked --offline --test glm_profile_cli ac4_`
- AC-5 (REQ-6): a GLM-shaped payload without rate_limits renders both windows as absent in the golden, and a payload with them renders unchanged · verify: `cargo test --locked --offline --test usage_limit_cli glm_absent`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-glm-profile/**`
- later slices, write: the profile generator (user-settings writer and its picks in the update route), `src/session.rs` (the env pair list becomes the profile's shadow), `src/launcher_terminal.rs` (env_keys), `src/statusline/input.rs`, `src/statusline/state.rs`, `tests/glm_profile_cli.rs`, `tests/usage_limit_cli.rs`, a GLM-shaped fixture under `tests/fixtures/statusline*/`
- forbidden in every slice: writing the real HOME or any real config folder; printing or storing a credential value; a project settings file as the profile target

## Plan

1. (done) SPEC(this slice).
2. Generator: on the user's pick, write the env block into the chosen config folder's user settings through the update preview; findings for raw credentials and leftover key variables; the launcher keeps reading the environment for this release and stops injecting the pair once the profile exists (one migration, scan test at the end).
3. Status line: the absent-5H/7D golden on a GLM-shaped fixture.
4. Capability list: the not-available list in the profile text, shared with the doctor capability table when that card lands.

## Risks

- A settings-file env value beats the shell export (documented). Lanes that still rely on the launcher's injected pair keep working, but the profile becomes the source of truth only after the migration removes the injection — until then two sources can disagree.
- The window value must be the real one per model: a wrong number moves the auto-compact point. The [1m] suffix rules differ per id shape (model-config.md); the generator writes the plain declaration and refuses to guess.
- The helper must print only the key (a banner fails it since v2.1.227); the profile text says so.
- Auth precedence (cloud variables, then the token, then the key, then the helper) means a leftover token or key silently overrides the helper; REQ-7 surfaces it, nothing more.
- Remote Control and voice are unavailable on the gateway route regardless of this profile (claude.ai identity needed); the not-available list prevents rediscovering this per lane.
- The shared root's user-settings env applies to every session under the root. The gateway credential beats a saved OAuth login, so while the block is active the user's own session in that root routes through the gateway too; how a first-party session coexists (or whether the whole root goes gateway) is the open question above.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The profile lives in the chosen config folder's user settings env block, not in the shell or a project file. | The background-agent reach warning and the committed-file warning, both quoted from llm-gateway-connect.md. | Fall back to launcher env. |
| D-2 | The credential path is the user's apiKeyHelper entry; moai names the slot and never the value. | Card body and the never-print rule; the helper is the documented rotating path. | Read the token from the vault as today. |
| D-3 | Beta-disable is a repair for the named 400 error only. | The card body ("400 Extra inputs only then"); the flag also turns tool search off (documented side effect). | Ship it by default. |
| D-4 | The window is declared with the documented variables, not by renaming the alias to a known id. | model-config.md: the declaration applies on its own for an unrecognized id; renaming changes what the gateway sees. | Map the alias instead. |
| D-5 | The not-available list is profile text with the offered-versus-usable distinction, not a runtime check. | The lane measured tools offered while the paths are unavailable; a list is cheap to keep true. | Turn it into a doctor check later. |
| D-6 | The absent 5H/7D is pinned as a golden on a GLM-shaped fixture. | Card criterion 2; the renderer already models the absence structurally. | Leave it unpinned. |
| D-7 | The profile's place is the shared config root the flow sessions run under (the user decision recorded on the session-messages card), applied only on the user's pick. | The same root is what makes lanes reachable to the user's session; a separate GLM-only folder would break that reachability. | A per-lane folder. |

### Open questions (the user's pick before the code slice)

- 루트 안에서 first-party 세션과 GLM 세션의 공존 — 세션별 시작 덮어쓰기로 사용자 세션은 first-party 를 유지할지, 아니면 그 루트 전체가 게이트웨이로 가는 것을 받아들일지. 사용자가 고르기 전까지 코드 조각은 시작하지 않아요.

## Result

(Filled when the work is done. This is the SPEC-only slice: no code, no acceptance criterion run yet; the research base — the two gateway documents and the model-config sections read whole, the report items, and this lane's same-day measurements — is recorded above.)
