---
id: SPEC-20261117-installer-skill-agents
title: 설치 스크립트가 skill과 agent까지 깔고, host마다 다음 단계를 알려 줘요
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

2026-11-17 전체 조사(F1·F2·F11)에서 나온 문제예요.

- `install/install.sh`의 `place()`는 `bin/`만 복사해요. `moai update`는 `<version>/skills/moai`와
  `<version>/agents`를 binary 옆에서 찾으니(`src/update.rs` `installed_package_dir`, `installed_agents_dir`), 새로
  설치하면 `no skill package installed with this binary`로 멈춰요.
- 설치 끝의 다음 단계가 skill 설치 전에 `/moai project`로 가요. Codex는 `$moai`, OMP는 `/skill:moai`인데 모두
  `/moai project`로 적혀 있어요. 루트 help도 `update`를 네 번째 단계로 둬요.
- `moai cc`, `moai codex`는 host의 skill이 없어도 아무 말 없이 시작해요.

목표: binary 설치 뒤 설치 스크립트가 제품 자신의 미리 보기→적용 경로(`moai-cli update --json`, 그다음
`moai-cli update --confirm <plan_digest> --json`)를 돌려요. 셸에서 host 폴더로 파일을 복사하는 일은 없어요.
쓰기 전에 어느 host 폴더에 무엇을 쓰는지 보여 주고, 터미널이면 한 번 묻고, 터미널이 아니면 `--yes`가 있을 때만
써요. 끝에는 찾은 host마다 들어가는 말(`/moai`, `$moai`, `/skill:moai`)을 적어요.

## Non-goals

- 릴리스 archive에 `skills/moai`와 `agents`를 넣는 dist 설정(`dist-workspace.toml`의 `include`)과 릴리스
  workflow. 이 SPEC은 archive가 그 둘을 binary 옆에 실어 올 때의 설치 쪽만 다뤄요.
- host 폴더 위치 규칙(`CLAUDE_CONFIG_DIR`, `CODEX_HOME` 등). 다른 branch(host-home resolver)가 고쳐요. 미리 보기에
  나오는 폴더는 제품이 정하고, 설치 스크립트는 host가 있는지만 봐요.
- 없는 host에 쓰는 계획(C-47), PATH 등록(C-27), OMP launcher(C-62), 버전 폴더 이름(HeadVer).

## Requirements

요구 문장은 영어 GEARS로 써요. "release part"는 archive에서 binary 옆에 있는 `skills/` 폴더와 `agents/` 폴더예요.
"host"는 Claude Code(`claude`), Codex(`codex`), OMP(`omp`)이고, 그 명령이 PATH에 있거나 HOME에 그 폴더(`.claude`,
`.codex`, `.omp`)가 있으면 찾은 거예요.

- REQ-1: When an installer places a version folder, the installer shall copy each release part into that version folder beside `bin/`. · failure: none — part가 없는 archive는 binary만 놓아요(예전 동작).
- REQ-2: When an installer finds an existing version folder with the same binary and without a release part that the archive carries, the installer shall add that part and leave `bin/` unchanged. · failure: none — 같은 binary가 아니면 기존 규칙대로 거절해요.
- REQ-3: When an installer finds a host, the installer shall run `moai-cli update --json` from the version folder and print the skill and agent folders of the preview.
- REQ-4: When the preview has a plan to apply and the user agrees, the installer shall run `moai-cli update --confirm <plan_digest> --json` once. · failure: none — 적용 실패는 REQ-8과 같은 안내로 끝나요.
- REQ-5 (REQ-3): IF no host is found, THEN the installer shall run no `update` and name the three hosts and `moai update`.
- REQ-6 (REQ-4): IF the user declines or there is no terminal and no `--yes`, THEN the installer shall write nothing into HOME outside the state folder and name `moai update`.
- REQ-7 (REQ-3): IF the preview has nothing to apply, THEN the installer shall apply nothing and drop the `moai update` step.
- REQ-8 (REQ-3): IF the update preview fails, THEN the installer shall keep the installed binary and exit 0 and print the product's message and `moai update`.
- REQ-9: The next steps shall name `moai update` first while the skill is not installed and then each host found with its entry (`moai cc` then `/moai project`, `moai codex` then `$moai project`, `omp` then `/skill:moai project`).
- REQ-10: When `moai cc`, `moai glm` or `moai codex` starts and the host's skill copy has no `SKILL.md`, the launcher shall print a warning that names the missing folder and `moai update`. · failure: none — 경고만 하고 host는 그대로 시작해요(HOME을 모르면 경고도 없어요).
- REQ-11: The root help shall list `update` before `cc` and `/moai project` in its first steps.

실패 경로: REQ-3 → REQ-5, REQ-7, REQ-8. REQ-4 → REQ-6.

## Acceptance

검사는 이 worktree에서 Python 3.13과 `pwsh`가 PATH에 있을 때 돌려요. 모든 설치 시험은 임시 HOME을 써요.

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): verify: `python3 -m unittest -k carries_the_skill -k without_the_skill -k real_update -k yes_applies tests/install_script_test.py`
  — release 모양의 archive(이 tree의 진짜 binary, skill, agent)를 `--yes`로 설치하면 `versions/<v>/skills/moai`,
  `versions/<v>/agents`, `~/.claude/skills/moai/SKILL.md`, `~/.claude/agents/moai-impl.md`가 생기고 다시 한 미리 보기는
  `nothing_to_apply: true`예요.
- AC-2 (REQ-5, REQ-6, REQ-7, REQ-8): verify: `python3 -m unittest -k no_host -k without_a_terminal -k installed_skill -k failing_preview -k asked_once tests/install_script_test.py`
- AC-3 (REQ-9): verify: `python3 -m unittest -k next_steps -k omp_gets -k without_a_terminal tests/install_script_test.py`
- AC-4 (REQ-10): verify: `cargo test --test launcher_skill_check`
- AC-5 (REQ-11): verify: `cargo test --test help_layout`, `cargo test --lib help::`
- AC-6 (REQ-1…REQ-9): verify: `sh -n install/install.sh`, `bash -n install/install.sh`, `shellcheck install/install.sh` → exit 0.

## Scope

- `install/install.sh`, `install/install.ps1`
- `src/launcher_init.rs`, `src/help.rs`
- `tests/install_script_test.py`, `tests/launcher_skill_check.rs`, `tests/fixtures/help/*.txt`
- `tests/launcher_permissions_setting.rs`: compares launches across two temporary HOMEs, so both get the skill
  copies (the warning names the HOME folder)
- 이 SPEC

## Plan

1. 시험 먼저: 설치 시험(가짜 `update`와 진짜 binary), launcher 시험, help golden.
2. `install.sh`, `install.ps1`: release part 복사, host 찾기, 미리 보기→질문→적용, host별 다음 단계.
3. `launcher_init::prepare`에 skill 확인 한 줄, `help.rs` 첫 단계 순서.

## Risks

- 미리 보기 JSON을 `sh`에서 `grep`/`sed`로 읽어요. `confirm_with`, `projection_root`, `dir` 키 이름이 바뀌면 보여 주는
  목록이 비어요. 적용은 `confirm_with`의 digest가 없으면 하지 않으니 잘못 쓰지는 않아요.
- host 찾기는 `$HOME/.<host>`를 봐요. host-home resolver가 들어오면 제품이 다른 폴더를 쓸 수 있어요; 그때 이 판단을
  제품 명령으로 옮겨요.
- launcher 경고는 `Prepared.note`에 한 줄을 더해요(`src/main.rs`는 범위 밖). 그래서 `--dry-run` JSON의 note에도
  들어가요.

## Decisions

- D1: 터미널 질문은 설치 스크립트가 한 번 해요(언어 메뉴와 같은 터미널 판단). 제품의 터미널 화면에 넘기지 않아서
  `curl | sh`, `pwsh`에서 같은 흐름이에요.
- D2: 미리 보기와 적용은 버전 폴더에서 돌려요. 사용자가 서 있는 프로젝트의 hook 계획이 섞이지 않게 해요.
- D3: 적용 실패와 미리 보기 실패는 경고로 끝나고 exit 0이에요(언어 기록 실패와 같은 규칙). binary 설치는 이미 끝났어요.
- D4: PowerShell 시험의 터미널 답은 처음부터 넣어요. pwsh가 prompt 곁에서 커서 위치(ESC[6n)를 묻고, 시험 pty는
  답하지 않아서 prompt 뒤에 친 `y`가 그 답으로 읽혔어요. 진짜 터미널은 이 질문에 답해요.

## Result

2026-11-17, base `d032d0be1f`, tip `17c19a2759`과 충돌 없음(`git merge-tree` exit 0).

- RED: 새 설치 시험은 바꾸기 전에 26 fail, 2 error(sh·ps1). `launcher_skill_check`는 3 fail(설치된 경우는 처음부터
  통과), help 단위 시험은 fail.
- GREEN: `tests/install_script_test.py` 101개 중 99개 통과·3 skip. 실패 2개 중 ps1 터미널 시험은 D4로 고쳐서 sh·ps1 모두
  통과. `InstallPs1Test.test_the_locale_is_read_from_lc_all_then_lc_messages_then_lang`(LC_ALL='')은 바꾸기 전 base에서도
  실패해요(이 Mac의 `Get-Culture`가 ko).
- launcher 시험 12개 target, banner_golden, product_cli, home_isolation_scan, install_cli, help_layout, `--lib help::`,
  `--lib launcher_init::`, `--lib ui_text::`, `--bin moai`, fmt, clippy `-D warnings`, `sh -n`·`bash -n`,
  repo_scans·ci_scope 통과. shellcheck는 이 기계에 없어서 돌리지 않았어요.
- mutant 3개 모두 잡혔어요: 터미널과 `--yes` 없이 적용, 사용자 폴더에서 미리 보기, `SKILL.md` 없는 폴더를 설치로 봄.
- 후속: 릴리스 archive가 `skills/moai`와 `agents`를 싣도록 dist 설정을 바꿔야 실제 설치에서 AC-1이 돼요(Non-goals).
