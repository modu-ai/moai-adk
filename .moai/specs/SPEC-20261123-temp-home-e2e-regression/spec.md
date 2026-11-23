---
id: SPEC-20261123-temp-home-e2e-regression
title: "A temp-HOME regression runs the landed Claude Code redesign end to end through the real binary"
status: draft
revision: 1
created: 2026-11-23
---

# A temp-HOME regression runs the landed Claude Code redesign end to end through the real binary

## Intent

Claude Code 재설계(R1-R5)가 조각마다 착지했어요: 공통 rules 설치(`SPEC-20261122-claude-md-home-routing`, `SPEC-20261123-rules-install-followups`), 설치 소유와 옛 상태 폴더 보관(`SPEC-20261123-home-state-retire`), 스킬 구조(`SPEC-20261123-skill-structure`), 훅 진입점 뼈대(`SPEC-20261123-hook-entry-skeleton`). 조각마다 자기 테스트는 있지만 "임시 HOME 에서 처음부터 끝까지 사용자 길로 이어서" 확인한 곳은 없어요. 이 SPEC 은 그 통합 회귀 한 파일(`tests/e2e_temp_home_cli.rs`)을 더해요. 진짜 바이너리 사본, 임시 HOME, 임시 프로젝트만 써요. 실제 `~/.claude` 와 `~/.moai` 는 읽지도 쓰지도 않아요.

덮는 길은 새 설치, 갱신 두 번, 충돌, 되돌리기, 개인 파일 보호, MoAI 가 아닌 프로젝트의 격리, 훅 진입점의 이벤트 입출력, 임시 저장소의 카드 한 바퀴(받기 → 확인 → 커밋 → 감사 평결 → 착지 → 다음 카드)예요.

### 읽은 것 (2026-11-23, NeXT 끝 `1e8114aba0`, 읽기만)

- 갱신은 바이너리 옆 `<버전 폴더>/skills/moai` 와 `<버전 폴더>/agents` 를 패키지로 읽어요(`src/update.rs` `installed_package_dir`, `installed_agents_dir`). 그래서 임시 버전 폴더에 이 트리의 바이너리 사본과 두 폴더를 두면 실제 설치 없이 같은 길이 돌아요(`tests/update_command_cli.rs` 의 `install_version` 이 같은 방식).
- 임시 HOME 도우미 `tests/common/home.rs` 의 `TempHome::command` 는 환경을 비우고 HOME 만 줘요. `tests/claude_md_home_cli.rs` 는 거기에 `MOAI_ACCOUNT_HOME` 을 더해요.
- 되돌리기 명령은 층마다 있어요: `home-routing-revert`, `agents-home-revert`, `home-style-revert`, 끊긴 적용은 `recovery-plan`/`recovery-apply`(`moai help --all`).
- 훅은 지금 진입점 뼈대뿐이에요: `moai hook session-context|bash-boundary|write-boundary` 는 판정 없이 지나가요(`src/hook_entry.rs`). 설치(settings 쓰기)와 H1 맥락 출력은 다음 조각이라 이 테스트는 그 출력 모양을 고정하지 않아요.
- 진짜 `claude` 를 부르는 일은 레인이 하지 않아요(리더 답 2026-11-12, `docs/e2e-matrix.md` 의 Q2 = HOLD 유지: 실제 host 실행, 인증, 비용). 이 SPEC 은 그 결정을 그대로 따라요.
- 시도했으나 막힌 것: 이 워크트리 세션은 `HOME=` 을 명령줄에서 바꾸는 셸 명령을 거절해서(가드가 막음, 돌려 부르지 않았어요), 미리보기 읽기를 셸로 해 보지 못했어요. 테스트 안에서 `Command::env` 로 HOME 을 주는 길은 그대로 쓸 수 있어요.

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
- REQ-7 (REQ-6): IF a managed target is a link or a file that is not UTF-8, THEN the update shall keep it, write nothing through it and report a conflict.
- REQ-8: When an applied layer is undone with its revert command or an interrupted apply is recovered with `recovery-plan` and `recovery-apply`, the command shall bring the HOME tree back to the digest it had before that apply. · failure: none — a digest that differs is what the test measures, and the layer's own tests hold the refusal paths
- REQ-9 (REQ-8): IF the user edited an applied file after the apply, THEN the revert shall keep the user's edit and name the file in its answer.
- REQ-10: When `moai hook session-context`, `bash-boundary` or `write-boundary` gets an event JSON on standard input in a temporary project, the entry point shall answer exit 0 with a standard output that is empty or one valid JSON document (always empty in a project without a MoAI profile) and with no file written into the project.
- REQ-11 (REQ-10): IF the standard input is empty or is not JSON, THEN the entry point shall exit 0 and write no file into the project.
- REQ-12: When one card runs through a temporary git repository and a temporary HOME (take, write check, commit in its own worktree, ready note, a verdict note from another session, `moai card land --by leader`), the card flow shall move the integration branch to the landed head, close the card and name the next ready card as the next one. · failure: none — refusals of the landing (no verdict, stale base, moved tip) belong to the landing's own tests

## Acceptance

- AC-1 (REQ-1, REQ-2): the test's helper builds the version folder and clean environment, a first scenario lists every path of its commands' answers and none lies outside the temporary root, and the test fails when a path is moved out of the root (a mutant of the helper) · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_root`
- AC-2 (REQ-3, REQ-4): a new install through preview and confirm writes the skill, the base agents (the installed agent files parse with this tree's `agent_defs` frontmatter rules), the routing and flow rules files and the style file into the temporary config folder, the written set equals the preview's list, a project folder next to it keeps its digest, and a wrong digest writes nothing · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_new_install`
- AC-3 (REQ-5): a changed package writes only the changed files and the next run of the same package leaves a HOME digest of 0 diff · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_two_updates`
- AC-4 (REQ-6, REQ-7): a hand-edited skill file, a hand-edited rules file, a linked managed target and a non-UTF-8 managed target are kept byte for byte with a conflict answer, and the user's own rules file, own agent, own `CLAUDE.md` text and another tool's settings keys keep their bytes · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_conflict_and_personal`
- AC-5 (REQ-8, REQ-9): each of `home-routing-revert`, `agents-home-revert` and `home-style-revert` returns the HOME digest of before its apply, an interrupted apply (a read-only target folder) is recovered to the pre-apply digest, and an edit made after the apply survives its revert · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_rollback`
- AC-6 (REQ-10, REQ-11): the three hook names exit 0 for a valid event in a MoAI project and in a plain project, print nothing in the plain project, write no file into either, and exit 0 for empty and non-JSON input · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_hook_entry`
- AC-7 (REQ-12): a card goes through take, `card check`, a commit in its worktree, the ready note, the verdict note from a second session and `card land --by leader`; the integration branch tip read with `git rev-parse` equals the landed head, the card is `done`, and `card next` names the second card · verify: `cargo test --locked --test e2e_temp_home_cli -- scenario_card_cycle`
- AC-8 (REQ-1 … REQ-12): the whole file passes and `moai check affected` selects it for its own diff · verify: `cargo test --locked --test e2e_temp_home_cli`

## Scope

- write: `tests/e2e_temp_home_cli.rs`
- write: `.moai/specs/SPEC-20261123-temp-home-e2e-regression/**`

## Risks

- 이 SPEC 은 되돌리기의 "사용자가 그 뒤에 고친 파일을 지킨다"(REQ-9)와 훅의 "JSON 이 아닌 입력은 조용히 지난다"(REQ-11)를 제품이 이미 한다고 가정해요. 코드를 끝까지 읽지 않았어요. 테스트가 다르게 나오면 약한 단언으로 바꾸지 않고, Result 에 그 결함을 적고 `src/**` 를 건드리는 수리 조각을 따로 열어요(반복 문제는 원천을 고쳐요).
- 카드 한 바퀴(REQ-12)의 착지가 임시 저장소에서 SPEC 검사와 기준 검사를 요구할 수 있어요. 필요하면 테스트 안에서 시험용 SPEC 하나를 임시 저장소에만 만들어요. 이 자료는 시험 자료로 표시하고, 실제 저장소 SPEC 의 승인과는 무관해요(승인은 사용자 몫).
- 테스트가 `moai` 를 여러 번 부르고 임시 저장소 git 을 써서 느릴 수 있어요. 한 파일에 시나리오 함수 여섯으로 나누고 공용 도우미로 설치를 한 번만 만들게 해요. 무거운 실행은 `heavy-slot --max 4` 안에서 한 번만 돌려요.
- 훅 불변식(REQ-10)은 H1 이 착지하면 출력이 생겨요. 그래서 "빈 출력이거나 유효한 JSON 한 문서"로만 단언해요. 이 조각이 H1 의 출력 모양을 고정하면 다음 조각이 이 파일을 고쳐야 해서 범위를 넘어요.

## Decisions

- D-1 (진짜 `claude` 실행, 운영자 단계): 주 세션, 에이전트 두 종류 호출, `/memory`·`/hooks`, `stream-json` 의 `system/init`·`plugin_errors`, `--init-only` 훅 시험은 이 테스트에 넣지 않고 Result 에 운영자 단계 명령과 기대 결과를 적고 `not_run` 으로 둬요. 옵션: (a) 그렇게 한다(기본), (b) 사용자가 정한 임시 HOME 에서 레인이 `claude --bare -p` 를 한 번 돌린다(인증과 비용이 들어서 사용자 선택 필요). 되돌릴 수 있어요(Result 에 줄을 더하면 돼요). 분류 `external_contract`·`secret` 에 닿아서 리더가 아니라 사용자 몫이에요. 기본은 (a)예요. 결정자: 사용자(대기). 기한 없음.
- D-2 (이 SPEC 의 승인): 승인은 사용자의 선택으로만 기록해요. 이 레인은 승인을 기록하지 않아요.

## Plan

1. `tests/e2e_temp_home_cli.rs` 에 공용 도우미: 임시 버전 폴더 만들기(바이너리 사본, `skills/moai`, `agents` 복사), 임시 HOME 과 프로젝트, 환경 비우기, HOME 트리 digest. 도우미 테스트(AC-1)를 먼저 써서 실패를 봐요.
2. 시나리오 순서대로 하나씩 실패하는 테스트로 시작해요: 새 설치 → 갱신 두 번 → 충돌·개인 파일 → 되돌리기 → 훅 진입점 → 카드 한 바퀴. 제품 코드는 바꾸지 않아요.
3. 결함이 나오면 그 시나리오 단언을 약하게 만들지 않고 Result 에 적고 수리 카드를 제안해요.
4. 무거운 실행은 `moai heavy-slot-acquire --name <lane> --max 4 -- <command>` 안에서 한 번, `moai check affected <base>..<head>` 는 READY 때 감싸지 않고 한 번 돌려요. 파이썬 테스트는 손으로 돌리지 않아요.
5. SYNC 에서 `## Result` 에 시나리오별 명령, 종료 코드, 센 수, 임시 저장소 카드 착지의 병합 SHA(시험 자료)와 `not_run` 항목을 적어요.

## Result

(draft: 아직 구현 전)
