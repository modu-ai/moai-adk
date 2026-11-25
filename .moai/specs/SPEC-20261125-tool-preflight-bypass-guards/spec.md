---
id: SPEC-20261125-tool-preflight-bypass-guards
title: "tool-preflight 가 우회 모드에서도 통합 브랜치 이동, 프로젝트 밖 rm -r, 강제 push 를 거절해요"
status: approved
revision: 1
created: 2026-11-25
approval: {by: leader, at: 2026-11-25, basis: "user pick relayed by the leader in card note 8948 (2026-11-25T03:58:34Z): 'approve, D-1~D-5 defaults (Recommended)' for revision 1 as committed in ee51af960e; the leader read the digest with spec show", digest: "sha256:db767f8d8475c833ee90fe17ff3b14b4ef5b20155cf5b7985b5af9a16be78832"}
---

# tool-preflight 가 우회 모드에서도 통합 브랜치 이동, 프로젝트 밖 rm -r, 강제 push 를 거절해요

## Interview

- Q1: 무엇이 열려 있나요? -> A: SPEC-20261125-agent-team-session-names 의 AC-9 확인(노트 8922, head 2d66176936)에서 권한 우회(`--dangerously-skip-permissions`) 실제 세션의 `git branch -f NeXT HEAD` 가 거절 없이 실행돼 통합 브랜치가 옮겨졌어요. `git update-ref`, `git push origin work:NeXT` 도 설치된 `tool-preflight` 가 허용해요. 절대 경로 `rm -rf` 도 훅은 허용했고, 막은 것은 사용자 설정의 거부 규칙뿐이었어요. `git push origin +NeXT`·`--force origin NeXT` 는 `ask` 라서 우회 모드에서는 물음 없이 지나가요.
- Q2: 사용자가 정한 것은요? -> A: 사용자 결정 2026-11-25(노트 8925): `-l`/`-a` 우회는 유지하고, 전역 빌드 설치는 이 SPEC 의 착지와 git 훅 보호 작업의 착지를 기다려요.
- Q3: 이미 쓸 만한 코드가 있나요? -> A: 있어요. `moai hook bash-boundary`(SPEC-20261123-hook-bash-boundary, 승인됨)의 `hook_bash_boundary::decide` 가 같은 틈을 닫는 판정을 이미 갖고 있어요. 임시 프로젝트에서 설치된 빌드(moai 4.2648.26)로 우회 모드 입력을 먹여 본 결과(2026-11-25T03:53Z 이전에 돌림, `target/c510/probe.sh`): 경계 훅은 `git branch -f NeXT HEAD`, `git update-ref refs/heads/NeXT HEAD`, `git push origin work:NeXT`, `git branch -D NeXT` 를 `protected-ref` 로, `rm -rf /opt/…` 를 `delete-outside` 로 거절하고, `/tmp` 아래와 워크트리 안 상대 경로는 허용해요. 그러나 settings 에 걸려 있는 것은 `hook-dispatch --host claude --event tool-preflight` 뿐이라(훅 설치는 그 SPEC 의 단계 1-e, Non-goals) 이 판정은 세션에 닿지 않고, `tool-preflight` 는 같은 다섯 명령에 빈 답, 강제 push 둘에는 `ask` 를 내요. 강제 push(비보호 브랜치)는 경계 훅도 판정하지 않아요(`ask` 를 쓰지 않는 설계).
- Q4: 우회 모드는 훅 입력에서 어떻게 보이나요? -> A: 훅 입력의 `permission_mode` 값은 `default`, `plan`, `acceptEdits`, `auto`, `dontAsk`, `bypassPermissions` 예요(공식 훅 문서 https://code.claude.com/docs/en/hooks, 2026-11-25T03:53:57Z 에 읽음, 저장한 페이지 sha256 접두 e6e00841acb898b5).

## Intent

우회 모드에서는 권한 물음에 답할 사람이 없고, 설치된 `tool-preflight` 는 통합 브랜치를 손으로 옮기는 명령과 경계 밖 삭제를 보지 못해요. 이 SPEC 은 새 판정 코드를 쓰지 않고, 이미 승인되고 감사받은 경계 판정을 `tool-preflight` 에서 부르게 해서 우회 모드 세션에서 통합 브랜치 이동·경계 밖 `rm -r` 이 거절되게 하고, `ask` 로 끝나던 강제 push 를 거절로 바꿔요. 옛 처리기의 판정 코드는 다시 쓰지 않아요(사용자 지시 msg_3dc06e21df4c, 경계 SPEC Intent). 통합 브랜치의 최종 강제는 git `reference-transaction` 가드가 하고, 이 변경은 그 앞의 빠른 거절이에요. 읽는 사람: 사용자(승인), 감사자, 전역 빌드를 설치하는 리더.

## Non-goals

- 우회 모드가 아닌 모드(`default`, `plan`, `acceptEdits`, `auto`, `dontAsk`, 값 없음)의 동작 변경. 그 모드에서는 사람이 물음에 답해요(D-1).
- 훅 설치와 settings 항목(경계 SPEC 의 단계 1-e), `moai hook bash-boundary` 자체의 판정 변경.
- 새 판정 코드(통합 브랜치 이름 읽기, `rm` 대상 판정, git 명령 분석)를 `src/hook_guards.rs` 에 더하는 일. 판정은 `hook_bash_boundary` 에만 있어요.
- `git reset --hard`, 강제 `git clean` 의 `ask`. 우회 모드에서 물음 없이 지나가는 점은 그대로이고 알려진 틈으로 적어 둬요(Risks).
- 임시 폴더(`/tmp`, `/private/tmp`, `/var/folders`…)의 삭제 거절. 흐름 규칙이 "자기 워크트리나 임시 폴더 안에서만 삭제"라서 경계 SPEC 의 D-3 대로 임시 폴더는 허용이에요(D-2).
- 훅 모드 스위치(`[hooks.modes]`): 그 파일은 퇴역한 `~/.moai-cli/` 에 있고 기본이 observe 라 틈이 열려 있게 돼요(D-4).

## Requirements

<!-- This SPEC adds a check that blocks a tool call in bypass mode: a violation is a deny. It does not start in observe on purpose (D-4): the gap it closes is measured and open today, and the final enforcement of the integration branch stays in the Rust reference-transaction guard. -->
- REQ-1: When `moai hook-dispatch --host claude --event tool-preflight` gets a `Bash` call whose `permission_mode` is `bypassPermissions` from a folder inside a MoAI project, the tool-preflight handler shall judge the command with `hook_bash_boundary::decide` and answer its deny unchanged. · failure: none — a call the old guards deny is denied first and by the same sentence as before.
- REQ-2: When a call in `bypassPermissions` mode is one that the force-push confirmation would ask about and no deny applies to it, the tool-preflight handler shall answer a deny that carries the confirmation's fixed sentence. · failure: none — the sentence names no command byte.
- REQ-3 (REQ-1): IF `permission_mode` is any other string, absent or not a string, THEN the tool-preflight handler shall answer exactly as before this SPEC. · failure: none — nothing new is judged.
- REQ-4 (REQ-1): IF no `.moai/project.yaml` is found at or above the working folder, THEN the tool-preflight handler shall answer exactly as before without calling the boundary judgment. · failure: none — a folder that is no MoAI project is not judged.
- REQ-5: The tool-preflight handler shall make the boundary judgment without starting a process, reading a network or writing a file beyond what `hook_bash_boundary` does itself. · failure: none — the module's own no-side-effect rule (its REQ-11) holds.

## Acceptance

- AC-1 (REQ-1): 임시 프로젝트(통합 브랜치 이름을 프로필에서 읽어요: `NeXT` 와 다른 이름 하나를 더 써서 이름을 박아 넣지 않았음을 보여요)의 `bypassPermissions` 입력에서 그 브랜치에 닿는 `git branch -f/-M/-D`, `git update-ref refs/heads/<b>`, `git push origin x:<b>` 가 모두 `moai-boundary/protected-ref:` 로 시작하는 deny 이고, 같은 명령이 보호되지 않은 레인 브랜치에 닿으면 허용이에요 · verify: `cargo nextest run --test hook_preflight_bypass_cli -E 'test(/^ac1_/)'`
- AC-2 (REQ-1): `bypassPermissions` 입력에서 프로젝트 체크아웃과 임시 폴더 밖의 절대 경로 `rm -r`(`/opt/…`, `/usr/local/…`, HOME 아래)이 `moai-boundary/delete-outside:` deny 이고, 워크트리 안 상대 경로 `rm -r build` 와 임시 폴더 안 경로는 지금과 같이 허용이에요 · verify: `cargo nextest run --test hook_preflight_bypass_cli -E 'test(/^ac2_/)'`
- AC-3 (REQ-2): `bypassPermissions` 입력의 `git push origin +feature`, `git push --force origin feature`, `git push -f origin feature` 가 deny 이고 사유가 `MoAI tool-preflight: a force push` 문장이며 명령 낱말을 담지 않아요; 같은 입력이 `default` 모드에서는 지금처럼 `ask` 예요 · verify: `cargo nextest run --test hook_preflight_bypass_cli -E 'test(/^ac3_/)'`
- AC-4 (REQ-3): AC-1~AC-3 의 모든 명령이 `default`, `plan`, `acceptEdits`, `auto`, `dontAsk`, 값 없음, 문자열 아닌 값에서 이 SPEC 이전과 같은 답(허용 또는 `ask`)이에요 · verify: `cargo nextest run --test hook_preflight_bypass_cli -E 'test(/^ac4_/)'`
- AC-5 (REQ-4): `.moai/project.yaml` 이 없는 폴더에서는 `bypassPermissions` 의 모든 AC-1·AC-2 명령이 이 SPEC 이전과 같은 답이에요 · verify: `cargo nextest run --test hook_preflight_bypass_cli -E 'test(/^ac5_/)'`
- AC-6 (REQ-1, REQ-2): 위험 경로 실행 — 임시 프로젝트(베어 원격 포함)에서 우회 모드의 실제 `claude -p` 세션이 훅을 `hook-dispatch --host claude --event tool-preflight` 한 줄만 설정한 상태에서 `git branch -f <통합 브랜치> HEAD`, `rm -rf <프로젝트 밖 절대 경로>`, `git push --force origin <브랜치>` 를 시도해 세 가지 모두 거절되고, 실행 전후 통합 브랜치 tip, 원격 tip, 대상 폴더가 같음을 읽어 READY 에 기록해요(자동 테스트가 아니라 기록된 실행이에요) · verify: READY 노트의 실행 기록(명령, 전후 읽기)과 `target/c510/` 의 스크립트·출력
- AC-7 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): 이웃 테스트가 그대로 통과해요 · verify: `cargo nextest run --test hook_dispatch_cli --test hook_bash_boundary_cli --test hooks_cli --test hook_entry_cli`
- AC-8 (REQ-5): 우회 모드 허용 명령과 거절 명령의 `hook-dispatch` 지연 p50·p90 을 실제 바이너리로 재서 Result 에 적어요(경계를 두지 않고 숫자만 보고해요) · verify: `cargo nextest run --test hook_preflight_bypass_cli -E 'test(/^ac8_/)' --no-capture`

## Scope

- write: `.moai/specs/SPEC-20261125-tool-preflight-bypass-guards/**`, `src/hook_engine.rs`(`tool_preflight` 에 한 갈래), `src/hook_guards.rs`(강제 push 확인 문장 상수를 `pub` 으로 한 낱말), `tests/hook_preflight_bypass_cli.rs`(새 파일)
- read-only: `src/hook_bash_boundary.rs`, `src/hook_entry.rs`, `src/hook_config.rs`, `tests/hook_bash_boundary_cli.rs`(세계를 만드는 본보기), `tests/common/**`
- forbidden: `src/hook_bash_boundary.rs` 의 판정 변경, `src/git_hooks/**`(다른 레인이 쥐고 있어요; 겹치면 `outside-scope:` 로 READY 에 적어요), 어떤 settings 파일, `CLAUDE.md`, `skills/**`, 이 저장소의 `.moai/project.yaml`, 실제 HOME

## Plan

1. SPEC 초안을 커밋하고 digest 를 `LANE-BLOCKED kind=decision` 으로 올려요. 레인은 승인을 기록하지 않아요. 승인 전에는 코드를 쓰지 않아요.
2. RED: `tests/hook_preflight_bypass_cli.rs` 에 AC-1~AC-5 를 써요. 임시 프로젝트·임시 HOME 에 진짜 `hook-dispatch` 를 불러요(세계는 `target/` 아래에 만들어요: 시스템 임시 폴더 자체가 경계의 임시 루트라서요). 지금 코드는 빈 답이라 AC-1·AC-2, `ask` 라 AC-3 이 올바른 이유로 실패해요.
3. GREEN: `hook_engine::tool_preflight` 에서 옛 deny 가 없고 Bash 이며 `permission_mode == "bypassPermissions"` 일 때 프로젝트 루트(`.moai/project.yaml` 를 위로 찾아요)를 찾아 `hook_bash_boundary::decide` 를 불러 deny 를 그대로 돌려줘요. 그 뒤 `bash_confirm` 이 강제 push 문장이면 같은 모드에서 deny 로 바꿔요(다른 확인은 그대로).
4. 확인: AC-7 이웃 테스트, `cargo fmt --all`, 새 가드 테스트마다 가드 뮤턴트 한 줄(셸 `sed`/`awk` 로 정확히 한 번 치환하고 `git checkout --` 로 되돌려요, Python 없음), AC-6 의 위험 경로 실행, AC-8 지연, `moai check affected --card` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적어요. tier 2 이므로 이 변경을 쓰지 않은 세션이 전체 감사로 맡아요(`agent-1`).

## Risks

- 우회 모드에서 경계 판정의 전체 분류(`unreadable`, `git-state`, `hook-skip` 포함)가 적용돼요: `rm -r "$DIR"`, `git -C "$VAR" …` 처럼 값을 명령 문장에서 읽을 수 없는 명령은 우회 모드의 `-l`/`-a` 세션에서 거절돼요. 경계 SPEC 이 받아들인 과거절(AC-18)과 같은 성질이고, 레인은 리터럴 경로를 써야 해요(CLAUDE.local.md 도 이미 그렇게 정했어요). 허용 표는 경계 쪽 테스트가 계속 지켜요.
- 알려진 틈(이 SPEC 이 닫지 않아요): 우회 모드에서 `git reset --hard`, 강제 `git clean` 의 `ask` 는 물음 없이 지나가요. 레인 워크트리 안의 정상 작업이라 거절로 바꾸면 흐름이 막혀요. 뒤 카드로 넓힐지는 사용자가 정해요.
- 실제 세션 확인 한계: AC-6 은 설치된 호스트 `claude` 와 `--dangerously-skip-permissions` 로 돌려요. 호스트 버전이 다르면 훅 입력 모양이 달라질 수 있어요(`permission_mode` 는 공식 문서 값 목록에 있어요). 값이 오지 않는 호스트에서는 REQ-3 대로 옛 동작이에요.
- 지연: 경계 판정은 파일 읽기만 해요. 우회 모드 Bash 호출마다 한 번 더 도니 AC-8 이 숫자를 보고해요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 판정은 `permission_mode == "bypassPermissions"` 일 때만 적용해요(기본). 대안: 모든 모드에 적용 | 카드와 세션 이름 SPEC 의 노트가 말한 틈은 우회 모드예요. 다른 모드는 사람이 물음에 답하고, 모든 모드에 적용하면 전역 빌드를 쓰는 일반 사용자의 보통 모드에도 `unreadable` 과거절이 번져요 | 조건 한 줄을 지우면 모든 모드로 넓어져요 |
| D-2 | 임시 폴더 안 삭제는 허용으로 둬요(경계 SPEC D-3). 대안: 절대 경로 `rm -r` 은 임시 폴더도 거절 | 흐름 규칙이 "자기 워크트리나 임시 폴더 안에서만 삭제"예요. 카드 AC-2 는 "프로젝트 밖"이라 임시 폴더 처리를 정하지 않았어요 | 경계 쪽 임시 루트 표를 바꾸는 별도 SPEC |
| D-3 | 경계 판정 전체(다섯 분류)를 그대로 써요. 대안: `protected-ref`·`delete-outside` 만 | 분류를 걸러 쓰면 `git branch -f "$B" HEAD` 가 `unreadable` 로 판정돼 걸러져 틈이 열려요. 감사받은 판정을 쪼개지 않는 편이 안전해요 | 분류 필터를 더해요(경계 모듈 변경 필요) |
| D-4 | 모드 스위치 없이 우회 모드에서 항상 거절해요. 대안: `[hooks.modes]` 규칙(기본 observe) | 그 설정 파일은 퇴역한 `~/.moai-cli/` 에 있고 기본이 observe 라 측정된 틈이 열려 있어요. 되돌리려면 코드를 되돌려요 | 갈래를 지워요 |
| D-5 | 우회 모드의 강제 push 는 모든 브랜치에서 거절해요(`+refspec`, `--force`, `-f`). 대안: 통합·보호 브랜치만 | 카드 기준 3 이 "ask 로 끝나던 강제 push"를 말해요. 레인은 push 하지 않는 흐름(`CLAUDE.md` Lane)이라 막혀도 정상 작업이 없어요 | 확인 문장 비교 한 줄을 지워요 |

## Result

- AC-1 met: `hook_preflight_bypass_cli::ac1_a_move_of_the_integration_branch_is_denied_in_bypass_mode` (두 프로필 이름 `NeXT`, `release-line`; 다섯 이동이 `moai-boundary/protected-ref:` deny, 보호되지 않은 브랜치 셋은 허용).
- AC-2 met: `ac2_a_delete_outside_the_checkout_is_denied_in_bypass_mode` (`/opt/…`, `/usr/local/…`, HOME 아래가 `moai-boundary/delete-outside:` deny; 워크트리 안 상대 경로와 `/tmp` 아래는 허용).
- AC-3 met: `ac3_a_force_push_is_denied_in_bypass_mode_and_still_asked_otherwise` (네 강제 push 가 deny, 같은 입력이 `default` 모드에서는 `ask`, 사유는 명령 낱말을 담지 않아요).
- AC-4 met: `ac4_every_other_mode_answers_as_before` (다섯 모드, 값 없음, 숫자, null).
- AC-5 met: `ac5_outside_a_moai_project_nothing_new_is_judged` (프로젝트 파일 없는 git 체크아웃; 가드 뮤턴트 m4 가 이 테스트를 실패시켜요).
- AC-6 met: 위험 경로 실행. 임시 프로젝트(베어 원격 포함)에서 `claude -p --setting-sources project --dangerously-skip-permissions`(claude 2.1.332)의 실제 세션이 훅을 `hook-dispatch --host claude --event tool-preflight` 한 줄만 설정한 상태에서 (a) `git branch -f NeXT HEAD` → `moai-boundary/protected-ref`, (b) `rm -rf <프로젝트 밖 절대 경로>` → `moai-boundary/delete-outside`, (c) `git push --force origin feature` → `MoAI tool-preflight: a force push …` 로 세 가지 모두 `PreToolUse:Bash hook error` 로 거절됐고, 통제 명령 `git status --short` 는 실행됐어요. 실행 전후 브랜치 tip(`rev-parse NeXT feature`), 원격(`ls-remote`), 대상 폴더(`ls -l` 과 `data.txt` 내용)가 같아요. 기록: `target/c510/real.sh`, `real.run.log`.
- AC-7 met: 이웃 테스트 `hook_dispatch_cli`, `hook_bash_boundary_cli`, `hooks_cli`, `hook_entry_cli` 와 새 테스트까지 152개가 통과했어요(nextest, `Summary 152 tests run: 152 passed`).
- AC-8 met: 지연은 debug 빌드에서 허용 명령 p50 12.8 ms · p90 16.0 ms, 거절 명령 p50 15.0 ms · p90 17.8 ms 예요(15회, 로드 평균 약 44; 경계를 두지 않고 숫자만 보고해요).
- 가드 뮤턴트 네 줄은 모두 시험 실패로 잡혔어요(`target/c510/mutants.txt`, `mutants-m4_no_project_check.txt`): m1 경계 갈래 끔 → ac1·ac2, m2 강제 push 갈래 끔 → ac3, m3 모든 모드 적용 → ac3·ac4, m4 프로젝트 검사 끔 → ac5. 첫 m4(`|| true` 를 클로저 안에 둠)는 동작이 같은 뮤턴트라 살아남았고, 조건 전체를 끄는 형태로 고쳐 다시 돌렸어요.
- 알려진 틈: 우회 모드의 `git reset --hard`·강제 `git clean` `ask` 와 직접 PR 머지(`merge_confirm`, `[hooks.modes]` 기본 observe)는 이 SPEC 이 닫지 않아요. 강제 push 거절 사유는 확인 문장(`… confirm it.`)을 그대로 써요(REQ-2 의 글자 그대로).
- 확인하지 않은 것: release 빌드 지연, 다른 호스트 버전의 `permission_mode` 입력 모양, 설치된 전역 빌드의 세션(설치는 리더의 몫).
