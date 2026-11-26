---
id: SPEC-20261123-temp-home-e2e-regression
title: "A temp-HOME regression runs the landed Claude Code redesign end to end through the real binary"
status: draft
revision: 4
created: 2026-11-23
---

# A temp-HOME regression runs the landed Claude Code redesign end to end through the real binary

## Intent

Claude Code 재설계(R1-R5)가 조각마다 착지했어요: 공통 rules 설치(`SPEC-20261122-claude-md-home-routing`, `SPEC-20261123-rules-install-followups`), 설치 소유와 옛 상태 폴더 보관(`SPEC-20261123-home-state-retire`), 스킬 구조(`SPEC-20261123-skill-structure`), 훅 진입점 뼈대(`SPEC-20261123-hook-entry-skeleton`). 조각마다 자기 테스트는 있지만 "임시 HOME 에서 처음부터 끝까지 사용자 길로 이어서" 확인한 곳은 없어요. 이 SPEC 은 그 통합 회귀 한 파일(`tests/e2e_temp_home_cli.rs`)을 더해요. 진짜 바이너리 사본, 임시 HOME, 임시 프로젝트만 써요. 실제 `~/.claude` 와 `~/.moai` 는 읽지도 쓰지도 않아요.

덮는 길은 새 설치, 갱신 두 번, 충돌, 되돌리기, 개인 파일 보호, MoAI 가 아닌 프로젝트의 격리, 훅 진입점의 이벤트 입출력, 임시 저장소의 카드 한 바퀴(받기 → 확인 → 커밋 → 감사 평결 → 착지 → 다음 카드)예요.

설치된 결과(스킬, 기본 에이전트, 공통 rules)도 이제 임시 HOME 에서 `claude` 를 부르지 않고 직접 읽어서 확인해요. 진짜 `claude` 는 맥OS 의 임시 HOME 에서 로그인을 유지하지 못해요(키체인을 `$HOME/Library/Keychains` 아래에서 찾아요). 이 이유는 운영자가 2026-11-25 에 관찰한 것이고(카드 원장의 운영자 실행 노트), 이 SPEC 이 잰 것이 아니에요. 임시 폴더 위에는 `CLAUDE.md` 가 없어야 해요. 부모 폴더의 파일이 실행에서 보이는 내용에 섞이지 않게 하려는 거예요.

### 읽은 것 (2026-11-23, NeXT 끝 `1e8114aba0`, 읽기만)

- 갱신은 바이너리 옆 `<버전 폴더>/skills/moai` 와 `<버전 폴더>/agents` 를 패키지로 읽어요(`src/update.rs` `installed_package_dir`, `installed_agents_dir`). 그래서 임시 버전 폴더에 이 트리의 바이너리 사본과 두 폴더를 두면 실제 설치 없이 같은 길이 돌아요(`tests/update_command_cli.rs` 의 `install_version` 이 같은 방식).
- 임시 HOME 도우미 `tests/common/home.rs` 의 `TempHome::command` 는 환경을 비우고 HOME 만 줘요. `tests/claude_md_home_cli.rs` 는 거기에 `MOAI_ACCOUNT_HOME` 을 더해요.
- 되돌리기 명령은 층마다 있어요: `home-routing-revert`, `agents-home-revert`, `home-style-revert`, 끊긴 적용은 `recovery-plan`/`recovery-apply`(`moai help --all`).
- 훅은 지금 진입점 뼈대뿐이에요: `moai hook session-context|bash-boundary|write-boundary` 는 판정 없이 지나가요(`src/hook_entry.rs`). 설치(settings 쓰기)와 H1 맥락 출력은 다음 조각이라 이 테스트는 그 출력 모양을 고정하지 않아요.
- 진짜 `claude` 를 부르는 일은 레인이 하지 않아요(리더 답 2026-11-12, `docs/e2e-matrix.md` 의 Q2 = HOLD 유지: 실제 host 실행, 인증, 비용). 이 SPEC 은 그 결정을 그대로 따라요.
- 시도했으나 막힌 것: 이 워크트리 세션은 `HOME=` 을 명령줄에서 바꾸는 셸 명령을 거절해서(가드가 막음, 돌려 부르지 않았어요), 미리보기 읽기를 셸로 해 보지 못했어요. 테스트 안에서 `Command::env` 로 HOME 을 주는 길은 그대로 쓸 수 있어요.

설치된 결과를 읽은 것 (2026-11-25, 머리 `b0d0978716`, 읽기만):

- 기본 에이전트는 `.claude/agents/moai/<name>.md` 로 가요. Claude Code 호스트의 폴더는 `agents/moai` 예요(`src/agent_install.rs:49-54` `host_rel`).
- rules 는 `.claude/rules/moai.md`(소스의 `{invoke}` 를 `/moai` 로 바꾼 것)와 `.claude/rules/moai-flow.md` 두 파일이에요(`src/home_routing.rs:44-54` 상수, `src/home_routing.rs:290-297` `render`). `rules/moai/flow.md` 와 `.claude/CLAUDE.md` 는 설치가 쓰지 않아요. `rules/moai/flow.md` 는 옛 위치라서 새 설치에는 없어야 해요(`LEGACY_FLOW_FILE`).
- `references/routes.json` 의 경로(`routes.<이름>.reference`, `chains`, `core_rules`)는 `references/` 기준 상대 경로예요(`"../workflows/run.md"`, `"chains/index.md"`). 경로 이름은 아홉 개(`AUDIT`, `LEADER`, `PLAN`, `PROJECT`, `READ`, `RUN`, `SYNC`, `UPDATE`, `WORK`)예요.
- 스킬 패키지는 파일 55개예요(`find skills/moai -type f | wc -l`).

## Non-goals

- 진짜 `claude` 실행: 주 세션, 두 종류의 에이전트 호출, `/memory`, `/hooks`, `stream-json` 의 `system/init`·`plugin_errors`, `--init-only` 훅 시험(커버리지 노트 A37). 이것들은 호스트 인증과 비용이 들어서 사용자 몫이에요. 이 SPEC 은 그 절차를 `docs/host-live-session-checklist.md` 에 맞춘 운영자 단계로 Result 에 `not_run` 으로 적기만 해요(Decisions D-1).
- 훅 설치, H1 맥락 출력, H2·H3 판정, 카드 종결 뒤 자동 정리: 각자 다른 조각의 일이에요(리더 결정 2026-11-23). 이 테스트는 그 조각이 착지해도 깨지지 않는 불변식만 재요.
- `src/**` 변경. 테스트가 결함을 찾으면 고치지 않고 Result 에 적어 수리 조각으로 넘겨요.
- Windows·Linux 결과. 맥 한 대에서만 돌고 그렇게 적어요.
- GLM 길(별도 조각의 테스트가 있어요).

## Requirements

- REQ-1: When the regression test starts a scenario, the test shall run every `moai` command through a copy of the built binary placed in a temporary version folder (`bin/moai-cli` beside copies of this tree's `skills/moai` and `agents`) with a clean environment whose HOME and `MOAI_ACCOUNT_HOME` name a temporary folder and whose project is a temporary folder. · failure: none — the helper builds the folders before any command runs, so a missing folder fails the setup itself
- REQ-2 (REQ-1): IF a path that a command reports or the test reads lies outside the temporary root, THEN the test shall fail naming that path.
- REQ-3: When `moai update` is previewed and then confirmed with the preview's digest in an empty temporary HOME, the update shall write exactly the files its preview listed (the skill, the base agents, the common rules and the output style) and nothing under any project folder.
- REQ-4 (REQ-3): IF the digest given to the confirmation is not the preview's, THEN the update shall write nothing and leave the HOME tree digest as it was.
- REQ-5: When the package changes and `moai update` runs again, the update shall write only the files whose package bytes changed, and a run of the same package after that leaves the HOME tree digest as it was. · failure: none — a changed byte on the second run is exactly what the test measures
- REQ-6: When the user has edited a file MoAI manages or owns a file MoAI does not manage (a rules file of their own, an agent of their own, their own `CLAUDE.md` text, another tool's settings keys), the update shall keep those bytes and report the edited managed file as a conflict.
- REQ-7 (REQ-6): IF a managed target is a link or a file that is not UTF-8, THEN the update shall keep it, write nothing through it and report `blocked` for the link and `conflict` for the file that is not UTF-8.
- REQ-8: When an applied layer is undone with its revert command or an interrupted apply is recovered with `recovery-plan` and `recovery-apply`, the command shall return every file outside MoAI's own state folders to the bytes it had before that apply, leaving only records inside the state folder (the revert receipt and the layer's state files). · failure: none — a file that differs is what the test measures, and the layer's own tests hold the refusal paths
- REQ-9 (REQ-8): IF the user edited an applied file after the apply, THEN the revert shall keep the user's edit and name the file in its answer.
- REQ-10: When `moai hook session-context`, `bash-boundary` or `write-boundary` gets an event JSON on standard input in a temporary project, the entry point shall answer exit 0 with a standard output that is empty or one valid JSON document (always empty in a project without a MoAI profile) and with no file written into the project.
- REQ-11 (REQ-10): IF the standard input is empty or is not JSON, THEN the entry point shall exit 0 and write no file into the project.
- REQ-12: When one card runs through a temporary git repository and a temporary HOME (take, write check, commit in its own worktree, ready note, a verdict note from another session, `moai card land --by leader`), the card flow shall move the integration branch to the landed head, close the card and name the next ready card as the next one. · failure: none — refusals of the landing (no verdict, stale base, moved tip) belong to the landing's own tests
- REQ-13: When `moai update --confirm` has installed into an empty temporary HOME, the installed skill folder shall equal the package's skill folder byte for byte, name `moai` in the front matter of its `SKILL.md` and hold every file that its `references/routes.json` and its fixed first reads name. · failure: none — a missing or differing file is exactly what the test measures
- REQ-14: When `moai update --confirm` has installed into an empty temporary HOME, the common rules shall be exactly the files `.claude/rules/moai.md` and `.claude/rules/moai-flow.md`, both non-empty, with the `{invoke}` of the routing source replaced and with no `.claude/rules/moai/flow.md`, `.claude/CLAUDE.md` or `CLAUDE.md` present. · failure: none — a missing or differing file is exactly what the test measures
- REQ-15: When `moai update --confirm` has installed into an empty temporary HOME, the base agents shall be exactly the five files `moai-audit`, `moai-docs`, `moai-impl`, `moai-plan` and `moai-research` as `.md` under `.claude/agents/moai/`, each with a front-matter name equal to its file name and a non-empty description, and none directly under `.claude/agents/`. · failure: none — a missing or differing file is exactly what the test measures
- REQ-16 (REQ-1): IF a folder above the temporary root or the temporary HOME holds `CLAUDE.md`, `CLAUDE.local.md` or `.claude/CLAUDE.md`, THEN the test shall fail naming that file.

## Acceptance

- AC-1 (REQ-1, REQ-2): the test's helper builds the version folder and clean environment, a first scenario lists every path of its commands' answers and none lies outside the temporary root, and the test fails when a path is moved out of the root (a mutant of the helper) · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_root`
- AC-2 (REQ-3, REQ-4): a new install through preview and confirm writes the skill, the base agents (the installed agent files parse with this tree's `agent_defs` frontmatter rules), the routing and flow rules files and the style file into the temporary config folder, the written set equals the preview's list, a project folder next to it keeps its digest, and a wrong digest writes nothing · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_new_install`
- AC-3 (REQ-5): a changed package writes only the changed files and the next run of the same package leaves a HOME digest of 0 diff · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_two_updates`
- AC-4 (REQ-6, REQ-7): a hand-edited skill file, a hand-edited rules file, a linked managed target and a non-UTF-8 managed target are kept byte for byte, the preview answers `conflict` for the first, the second and the non-UTF-8 file and `blocked` for the link (the test compares each answer with its exact text), and the user's own rules file, own agent, own `CLAUDE.md` text and another tool's settings keys keep their bytes · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_conflict_and_personal`
- AC-5 (REQ-8, REQ-9): each of `home-routing-revert`, `agents-home-revert` and `home-style-revert` returns every file outside MoAI's state folders to what it was before its apply and leaves only files under the state folder, an interrupted apply (a read-only target folder) is recovered so the skill folder digest is the one before it, and an edit made after the apply survives its revert · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_rollback`
- AC-6 (REQ-10, REQ-11): the three hook names exit 0 for a valid event in a MoAI project and in a plain project, print nothing in the plain project, write no file into either, and exit 0 for empty and non-JSON input · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_hook_entry`
- AC-7 (REQ-12): a card goes through take, `card check`, a commit in its worktree, the ready note, the verdict note from a second session and `card land --by leader`; the integration branch tip read with `git rev-parse` equals the landed head, the card is `done`, and `card next` names the second card · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_card_cycle`
- AC-8 (REQ-1 … REQ-16): the whole file passes and `moai check affected` selects it for its own diff · verify: `cargo test --locked --test e2e_temp_home_cli`
- AC-9 (REQ-13): after a new install the installed skill folder lists the same files as the package copy with the same bytes and no extra file, its `SKILL.md` front matter has `name: moai`, `references/routes.json` names exactly the nine routes, and every route reference, `chains`, `core_rules` and the four first reads (`references/chains/index.md`, `references/internal/core-rules.md`, `references/flows/next-flow.md`, `references/flows/team-messages.md`) are files inside the skill folder · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_installed_skill`
- AC-10 (REQ-14): after a new install `.claude/rules/moai.md` equals the routing source with its one `{invoke}` replaced by `/moai`, `.claude/rules/moai-flow.md` equals the flow source, neither is empty or starts with a front-matter header, and `.claude/rules/moai/flow.md`, `.claude/CLAUDE.md` and `CLAUDE.md` do not exist · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_installed_rules`
- AC-11 (REQ-15): after a new install `.claude/agents/` holds only the folder `moai`, which holds exactly the five base agent files, each with a front-matter name equal to its file name and a non-empty description · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_installed_agents`
- AC-12 (REQ-16): before and after a new install no `CLAUDE.md`, `CLAUDE.local.md` or `.claude/CLAUDE.md` lies in the temporary HOME, the temporary project or any folder above them up to `/` · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_installed_ancestors`
- AC-13 (REQ-13, REQ-14, REQ-15, REQ-16): the four checkers return nothing for a clean made-up tree and a problem that names the planted defect for each defect (a missing, renamed or extra file, a wrong front-matter name, an empty file, an unreplaced `{invoke}`, a context file in an ancestor folder); they run on made-up trees only · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_installed_checkers`

## Scope

- write: `tests/e2e_temp_home_cli.rs`
- write: `.moai/specs/SPEC-20261123-temp-home-e2e-regression/**`

## Risks

- 이 SPEC 은 되돌리기의 "사용자가 그 뒤에 고친 파일을 지킨다"(REQ-9)와 훅의 "JSON 이 아닌 입력은 조용히 지난다"(REQ-11)를 제품이 이미 한다고 가정해요. 코드를 끝까지 읽지 않았어요. 테스트가 다르게 나오면 약한 단언으로 바꾸지 않고, Result 에 그 결함을 적고 `src/**` 를 건드리는 수리 조각을 따로 열어요(반복 문제는 원천을 고쳐요).
- 카드 한 바퀴(REQ-12)의 착지가 임시 저장소에서 SPEC 검사와 기준 검사를 요구할 수 있어요. 필요하면 테스트 안에서 시험용 SPEC 하나를 임시 저장소에만 만들어요. 이 자료는 시험 자료로 표시하고, 실제 저장소 SPEC 의 승인과는 무관해요(승인은 사용자 몫).
- 테스트가 `moai` 를 여러 번 부르고 임시 저장소 git 을 써서 느릴 수 있어요. 한 파일에 시나리오 함수 여섯으로 나누고 공용 도우미로 설치를 한 번만 만들게 해요. 무거운 실행은 `heavy-slot --max 4` 안에서 한 번만 돌려요.
- 훅 불변식(REQ-10)은 H1 이 착지하면 출력이 생겨요. 그래서 "빈 출력이거나 유효한 JSON 한 문서"로만 단언해요. 이 조각이 H1 의 출력 모양을 고정하면 다음 조각이 이 파일을 고쳐야 해서 범위를 넘어요.
- `TMPDIR` 가 `CLAUDE.md` 가 있는 폴더 안(이 저장소 자신)을 가리키면 REQ-16 의 시험은 설계대로 실패하고 그 파일 이름을 대요. 임시 폴더를 위에 `CLAUDE.md` 가 없는 곳에 두면 돼요(Decisions D-3).
- 아홉 개 경로 이름, 첫 읽기 네 파일, 에이전트 다섯 이름은 시험 안에 글자 그대로 있어요(Decisions D-4). 나중 조각이 하나를 바꾸거나 더하면 그 조각이 시험의 목록도 고쳐요.

## Decisions

- D-1 (진짜 `claude` 실행, 운영자 단계): 주 세션, 에이전트 두 종류 호출, `/memory`·`/hooks`, `stream-json` 의 `system/init`·`plugin_errors`, `--init-only` 훅 시험은 이 테스트에 넣지 않고 Result 에 운영자 단계 명령과 기대 결과를 적고 `not_run` 으로 둬요. 옵션: (a) 그렇게 한다(기본), (b) 사용자가 정한 임시 HOME 에서 레인이 `claude --bare -p` 를 한 번 돌린다(인증과 비용이 들어서 사용자 선택 필요). 되돌릴 수 있어요(Result 에 줄을 더하면 돼요). 분류 `external_contract`·`secret` 에 닿아서 리더가 아니라 사용자 몫이에요. 기본은 (a)예요. 결정자: 사용자(대기). 기한 없음.
- D-2 (이 SPEC 의 승인): 승인은 사용자의 선택으로만 기록해요. 이 레인은 승인을 기록하지 않아요.
- D-3 (임시 폴더 위에 `CLAUDE.md` 가 있을 때): 옵션: (a) 시험이 실패하고 그 파일 이름을 댄다(기본), (b) 시험이 위에 `CLAUDE.md` 가 없는 폴더를 골라 거기에 임시 루트를 만든다. (a)는 부모 파일이 보이는 내용에 섞이는 일을 숨기지 않고, (b)는 `TMPDIR` 설정과 상관없이 돌지만 시험이 쓸 폴더를 정하는 새 규칙이 생겨요. 되돌릴 수 있어요(시험 도우미 한 곳). 분류 `none`. 결정자: 리더(되돌릴 수 있는 선택이라서). 기한 없음.
- D-4 (고정 목록): 옵션: (a) 아홉 경로 이름, 첫 읽기 네 파일, 에이전트 다섯 이름을 시험에 글자 그대로 둔다(기본), (b) 패키지에서 뽑는다. (b)는 파일 하나를 패키지와 경로 표에서 함께 빼면 시험이 통과해서 그 누락을 못 잡아요. (a)는 이름이 바뀔 때 목록을 같이 고쳐야 해요. 되돌릴 수 있어요. 분류 `none`. 결정자: 리더. 기한 없음.

## Plan

1. `tests/e2e_temp_home_cli.rs` 에 공용 도우미: 임시 버전 폴더 만들기(바이너리 사본, `skills/moai`, `agents` 복사), 임시 HOME 과 프로젝트, 환경 비우기, HOME 트리 digest. 도우미 테스트(AC-1)를 먼저 써서 실패를 봐요.
2. 시나리오 순서대로 하나씩 실패하는 테스트로 시작해요: 새 설치 → 갱신 두 번 → 충돌·개인 파일 → 되돌리기 → 훅 진입점 → 카드 한 바퀴. 제품 코드는 바꾸지 않아요.
3. 결함이 나오면 그 시나리오 단언을 약하게 만들지 않고 Result 에 적고 수리 카드를 제안해요.
3a. 설치된 결과 시나리오(REQ-13 … REQ-16): 만든 가짜 트리에서 네 검사기를 먼저 시험하는 도우미 시험(`scenario_installed_checkers_*`)을 써서 실패를 보고, 그다음 시나리오 넷(스킬, rules, 에이전트, 위쪽 폴더)을 하나씩 써요. 진짜 `claude` 는 부르지 않아요.
4. 무거운 실행은 `moai heavy-slot-acquire --name <lane> --max 4 -- <command>` 안에서 한 번, `moai check affected <base>..<head>` 는 READY 때 감싸지 않고 한 번 돌려요. 파이썬 테스트는 손으로 돌리지 않아요.
5. SYNC 에서 `## Result` 에 시나리오별 명령, 종료 코드, 센 수, 임시 저장소 카드 착지의 병합 SHA(시험 자료)와 `not_run` 항목을 적어요.

## Result

시험 파일 `tests/e2e_temp_home_cli.rs` 의 시나리오 10개를 구현했어요. 판 1 은 사용자 선택으로 승인됐고(카드 원장 노트 7497, 2026-11-23T12:39:01Z), 판 2(REQ-8 과 AC-5 문구만 바꿈)도 사용자 선택이에요(노트 7527, 2026-11-23T13:15:47Z). 두 선택은 리더 세션이 사용자에게 물어 원장에 옮겨 적었고, 이 레인이 만든 승인이 아니에요. 판 3(REQ-7 과 AC-4 문구만 바꿈: 링크 대상은 `blocked`, UTF-8 이 아닌 파일은 `conflict`)의 근거는 카드 원장 노트 7552 의 `LEADER 결정 (2026-11-23T14:09:18Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 되돌릴 수 있음, 분류 none)` 이에요: 감사 FAIL B1 의 수리로 "링크 대상은 제품이 답하는 `blocked` 를 고정, SPEC 3판에 REQ-7 을 '링크 대상은 blocked' 로". 이 레인은 승인을 기록하지 않고 그 노트를 인용만 해요(머리말 approval 키 없음). 제품 코드(`src/**`)는 바꾸지 않았어요. 범위를 넓힌 파일 `tests/repo_scans_table.rs` 에는 `EXEMPT` 줄 하나만 더했어요.

### 실행 기록 (2026-11-23, 이 워크트리)

- `moai heavy-slot-acquire --name temp-home-e2e --max 4 -- env RUSTC_WRAPPER=sccache cargo test --locked --test e2e_temp_home_cli -- --test-threads=2`: exit 0, 10 passed, 0 failed (25.48s, 에이전트 단언을 더한 시험 파일).
- 감사 FAIL 수리 뒤(2026-11-23): `moai heavy-slot-acquire --name temp-home-e2e --max 4 -- cargo nextest run --locked --test e2e_temp_home_cli --no-capture` (`RUSTC_WRAPPER=sccache`): exit 0, 10 tests run: 10 passed, 0 skipped. 같은 실행에서 `ACTIONS audit=conflict routing=conflict flow=conflict style=blocked`, 세 번째 `--confirm` 의 `result.status` 는 `unchanged`.
- 임시 저장소 카드 한 바퀴의 착지: 시험 자료 병합 SHA 는 실행마다 달라요(한 실행의 예: 착지 head `ceb3dd2f65`, 기준 `42c0ee66ae`). 시험은 `git rev-parse refs/heads/integ` 가 착지 head 와 같은지 단언해요.
- 카드 한 바퀴가 착지에서 거절한 이유 셋(시험을 쓰는 동안 관찰, 모두 제품이 의도한 거절): `authors_unknown`(레인 커밋에 `Moai-Session` 트레일러 없음), `lander_review`(착지 세션이 감사 세션과 달라야 함), `receipt_missing`(PASS 가 이 head 의 영수증을 인용해야 함). 시험은 도구가 찍은 트레일러, 세 번째 세션, 영수증 단계로 이를 채워요.

### 뮤턴트 (스크래치패드 복제본 `d4a992010b`, 한 줄 변이, 각 시나리오만 실행)

- `tools/mutate.py` 는 복제본에서 오프라인 의존성(`cap-fs-ext`)을 못 찾아 네 개 모두 `error` 로 끝났어요. 죽은 것으로 세지 않았어요. 같은 한 줄 변이를 복제본에 손으로 적용해서 cargo 로 돌렸어요.
- `scenario_conflict_and_personal_files_are_kept`: `src/home_routing/rules.rs` 의 편집된 파일 판정 `"conflict"` → `"merge"`: exit 101 (죽음).
- `scenario_rollback_each_layer_returns_the_files_it_changed`: 같은 파일의 `user_changed_since_apply` 건너뛰기 조건에 `false &&`: exit 101 (죽음, 사용자가 고친 파일이 지워져요).
- `scenario_hook_entry_exits_zero_and_writes_nothing_into_the_project`: `src/hook_entry.rs` 의 영수증 실패 갈래 `stdout: String::new()` → `String::from("x")`: exit 101 (죽음).
- `scenario_root_checker_names_a_path_outside_the_root`: 시험 도우미의 `!Path::new(word).starts_with(root)` → `false`: exit 101 (죽음).

감사 FAIL(노트 7552) 수리 뒤의 뮤턴트 (스크래치패드 복제본, 한 줄 변이, `cargo nextest run --locked --test e2e_temp_home_cli`, 복제본은 이 워크트리의 소스와 시험 파일 사본이에요):

- 수리 전 시험 + W1(`src/home_routing/rules.rs:220` `"conflict"` → `"unchanged"`)·W2(`src/update.rs:1412` `Decision::Conflict => "conflict"` → `"preserved"`) 동시 적용: `scenario_conflict_and_personal_files_are_kept` exit 0 (살아남음, ACTIONS audit=preserved routing=unchanged) — 감사자가 본 빈틈을 이 실행에서도 봤어요.
- 수리 뒤 시험 + W1 만: exit 100 (죽음, `rules/moai.md is answered with the wrong action`).
- 수리 뒤 시험 + W2 만: exit 100 (죽음, `skills/moai/references/audit.md is answered with the wrong action`).
- `scenario_root_checker_names_a_path_outside_the_root`: 단어 나누기에서 `:`·`=` 를 뺌: exit 100 (죽음, 나온 값 `["/usr/bin/git"]`).

### 인수 기준

- AC-1: met — 시나리오 `scenario_root_checker_*`, `scenario_root_every_answer_*` 통과, 도우미 변이 exit 101. 경로 찾기는 `:`·`=` 뒤의 경로(`path=/x`, `dir:/x`)도 단어로 봐요(수리 N5).
- AC-2: met — `scenario_new_install_*` 와 `scenario_non_moai_project_*` 통과: 미리보기의 4층(스킬, 에이전트, 라우팅 rules·flow, 스타일 파일과 settings) 파일이 모두 쓰였고 그 밖에는 MoAI 상태 폴더(`.moai`, `.moai-cli`)뿐이며, 틀린 digest 는 아무것도 쓰지 않고, 옆 프로젝트와 안에서 시작한 git 프로젝트의 트리 digest 는 그대로예요. 설치된 에이전트 5개는 이 트리의 `agent_defs::render` 결과와 바이트가 같고 frontmatter 가 `name: moai-…` 로 시작해요.
- AC-3: met — `scenario_two_updates_*`: 바뀐 세 파일만 쓰고, 같은 패키지의 다음 실행은 `nothing_to_apply` 와 digest 0 차이. 세 번째 계획을 `--confirm` 까지 돌려도(`result.status` 가 `unchanged`) HOME 트리 digest 는 그대로예요(수리 N1).
- AC-4: met (판 3 문구) — `scenario_conflict_and_personal_files_are_kept`: 손으로 고친 스킬 파일과 라우팅 파일, UTF-8 이 아닌 flow 파일은 미리보기가 정확히 `conflict`, 링크로 바뀐 스타일 파일은 정확히 `blocked` 라고 답하는 것을 `assert_eq!` 로 고정했어요(이전 판은 `create`·`update` 가 아닌지만 봐서 W1·W2 가 살아남았어요). 모두 바이트 그대로이고 링크 너머 파일에도 쓰지 않았어요. 사용자 자신의 rules, agents, skills, `CLAUDE.md` 와 다른 도구의 settings 키도 그대로예요.
- AC-5: met (판 2 문구) — 세 층의 되돌리기는 MoAI 상태 폴더 밖 파일을 적용 전 바이트 그대로 돌리고, 남는 파일은 모두 상태 폴더 안이에요(시험이 단언). 한 실행에서 남은 것: `.moai/home-routing/receipts/…/receipt.json`·`rules-….json`, `.moai/agents/claude.json`·`receipts/…/receipt.json`, `.moai/home-style/claude.json`·`receipts/…/receipt.json`·`claude-setting.before`. 영수증뿐이 아니라 층 상태 파일도 남아서 문구에 "(되돌리기 영수증과 층 상태 파일)" 을 적었어요(사용자가 고른 문구보다 한 구절 더 정확해요, 리더 확인 요청). 적용 뒤 사용자가 고친 파일은 지켜지고 답에 파일 이름이 나와요. 끊긴 갱신은 `recovery-plan`·`recovery-apply` 로 스킬 폴더 digest 가 이전과 같아졌어요. 판 1 의 "HOME 트리 전체 digest 가 같다"는 거짓이었어요(되돌린 뒤 상태 폴더 기록이 남아 HOME 전체 digest 는 같지 않아요, 매 층에서 관찰).
- AC-6: met — `scenario_hook_entry_*`: 세 이름 × MoAI/일반 프로젝트 × 5가지 입력(이벤트, 빈 입력, JSON 아님, 객체 아님, UTF-8 아님)이 exit 0, 출력은 비었거나 JSON 한 문서이고 프로젝트 digest 는 그대로예요(변수 `CLAUDE_PROJECT_DIR` 경로도). 참고: MoAI 프로젝트 판별은 `.moai/project.yaml`(훅 진입점)이고 카드·장부 쪽 프로필은 `.moai-cli/project.yaml`(`PROJECT_STATE_DIR`)이에요. 이 시나리오는 훅 쪽 파일만 써요.
- AC-7: met — `scenario_card_cycle_*`: 받기, `card check`(범위 안 true, 밖 not true), 레인 워크트리 커밋, 준비 노트, `review`, 다른 세션의 PASS(영수증 포함), 세 번째 세션의 `card land --by leader` 가 통합 브랜치를 착지 head 로 옮기고 카드를 `done` 으로 닫았으며 `card next` 가 두 번째 카드를 `ready` 로 답했어요.
- AC-8: met for the file run — 시험 파일 전체 10 passed (exit 0). `moai check affected --list 1e8114aba0..HEAD` 가 `e2e_temp_home_cli` 를 고른 것을 봤어요. `check affected` 본 실행의 영수증은 READY 노트에 있어요.

### 이 SPEC 이 하지 않은 것 (not_run, 사용자 결정 D-1 = 빼고 운영자 단계로)

운영자가 임시 HOME 과 `CLAUDE_CONFIG_DIR=<임시 HOME>/.claude` 로 `claude` 를 띄워(`docs/host-live-session-checklist.md` §3 의 격리, 실제 `~/.claude`·`~/.moai` 금지) 아래를 직접 확인해요. 이 레인은 실제 `claude` 를 실행하지 않았어요.

1. 임시 HOME 에 `moai update --confirm` 으로 설치한 뒤 `/memory`: `~/.claude/rules/moai.md` 와 `rules/moai-flow.md` 가 보이는지, 프로젝트 `CLAUDE.md` 가 보이는지. not_run.
2. `/hooks`: 프로젝트 훅 설치(다음 조각)가 착지한 뒤에만 의미가 있어요. not_run.
3. 주 세션에서 `/moai` 가 나오고 한 경로가 도는지, `/agents` 에 `moai-plan`·`moai-research`·`moai-impl`·`moai-audit`·`moai-docs` 가 보이는지, 각 종류의 에이전트를 한 번씩 부르는지(하네스 에이전트는 프로젝트가 만든 뒤). not_run.
4. `claude -p --output-format stream-json --verbose` 의 `system/init` 과 `plugin_errors`(비어 있음), `--init-only` 훅 시험(커버리지 노트 A37). not_run.

### 판 4: 설치된 결과 시나리오 (머리 `b0d0978716`, 2026-11-25)

시험 파일 `tests/e2e_temp_home_cli.rs` 에 `// ---- installed result` 구역과 시나리오 다섯(`scenario_installed_checkers_name_what_is_wrong`, `scenario_installed_skill_routes_and_first_reads_resolve`, `scenario_installed_rules_are_the_two_named_files`, `scenario_installed_agents_are_nested_with_matching_names`, `scenario_installed_ancestors_hold_no_context_file`)가 더해졌어요. 아래 실행은 이 SPEC 을 고친 문서 레인이 아니라 시험을 쓴 세션이 돌린 기록이고, 문서 레인은 다시 돌리지 않았어요.

- 빨강 실행: `cargo nextest run --locked --test e2e_temp_home_cli scenario_installed_checkers` 가 `tests/e2e_temp_home_cli.rs:1204` 에서 실패했어요(`left: []`, `right: [".../fake/a/CLAUDE.md"]`, 검사기가 빈 값을 돌려주는 뼈대였어요).
- 파일 전체: `cargo nextest run --locked --no-fail-fast --test e2e_temp_home_cli`: 15 tests run, 15 passed, exit 0 (앞의 시나리오 열 개와 새 다섯 개).
- 인접 스캔: `cargo nextest run --locked --no-fail-fast --test repo_scans_table --test state_folder_literal_scan`: 43 passed, exit 0.
- 시나리오가 찾은 결함은 없어요: 설치된 스킬 폴더는 패키지와 같고, rules, 중첩된 에이전트, 위쪽 폴더 검사가 모두 기대와 맞았어요.
- AC-9 … AC-13: met, 위 실행이 보여 주는 만큼만이에요(스킬 폴더 일치와 경로 해석, rules 두 파일, 중첩 에이전트, 위쪽 폴더, 검사기 만든 트리). AC-8: 시험 파일 전체가 이제 15 passed 예요(앞 줄의 10 passed 는 판 3 때 기록이에요).
- `moai check affected --card` 영수증은 READY 노트에 있어요. clippy 와 뮤턴트는 이 슬라이스에서 `not_run` 이에요(tier 1 이라 요구되지 않아요).
- 진짜 `claude` 실행은 그대로 운영자 단계예요(D-1, `not_run`). 그래서 카드의 기준 1 은 이 슬라이스로 met 표시하지 않아요(진짜 `claude` 부분이 `not_run`).
- 판 4 는 아직 승인되지 않았어요. 머리말에 `approval` 키가 없고 `status: draft` 예요(D-2).
