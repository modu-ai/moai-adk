---
id: SPEC-20261121-remove-codex-support
title: "Remove Codex support: Claude Code and GLM only, with the inventory, the slice order and the restore record"
status: draft
revision: 2
created: 2026-11-21
---

# Remove Codex support: Claude Code and GLM only

## Intent

사용자가 2026-11-21 에 결정했어요: "codex 지원을 철회한다. moai-adk codex 관련 자료 모두 제거하고 claude code, glm 만 지원하도록 하자. card, spec 은 언젠가 codex 가 쓸만해지면 사용하도록 하고…". 그래서 moai-adk 는 Claude Code(GLM 은 Claude Code 실행기를 거치는 모델 선택)만 지원하고, Codex 와 OMP(Oh My Pi, 같은 결정으로 함께 접어요)의 코드·스킬 글·문서·테스트·설치와 업데이트 단계·어댑터를 걷어내요. 이 SPEC 은 조각 0 이에요. 코드는 한 줄도 바꾸지 않고, 걷어낼 곳 전부의 목록(파일:줄, 영역별), 걷어내는 순서, 사용자에게 보일 한 줄, 되살릴 때 쓸 복원 기록을 정해요. 제거 조각 1~3 은 이 SPEC 이 정한 파일 소유와 순서를 따라 따로 SPEC 을 가져요.

revision 2 (2026-11-22): 사용자가 "다음 로컬 판은 Codex·OMP 쓰기를 걷어낸 뒤에 설치한다"고 골랐어요(리더 기록 2026-11-22). 그래서 이 SPEC 은 revision 1(기준 `53cce6c580`)의 목록을 지금 tip 으로 다시 뽑고, 그 사이 착지한 것(호환 층, 런처·모델 조각)을 반영하고, **지금 `moai update` 미리보기가 쓰는 곳**을 따로 표로 만들어 어느 조각이 지우는지 정해요. 코드는 이번에도 한 줄도 바꾸지 않아요.

기준은 `v2026.11.0` 의 `be62c06a3991b96f12a1fdb133b0190ca514e1dc`(revision 2 를 시작한 로컬 통합 가지 tip. 리더가 git 훅 조각의 WIP 커밋을 정리하려고 가지 끝을 한 번 다시 만들었고, `46e831d6f8` 와 파일 내용이 같아요: 두 커밋의 tree 가 같은 `4b6c5e304d`)이에요. 같은 낱말로 센 숫자(`inventory.py` 가 그 커밋의 파일 내용만 읽어 셈해요): 파일 606개, 줄 6585줄. 이 가운데 `codex`(대소문자 무관)를 담은 파일이 584개, `OMP` 낱말만 담은 파일이 19개, 경로 이름에만 Codex 가 있고 내용에는 낱말이 없는 파일이 3개(0줄)예요. 584 + 19 + 3 = 606 이에요.

## Non-goals

- 이 조각에서 코드·테스트·문서·스크립트를 지우거나 고치는 일: 하지 않아요. 목록과 순서만 정해요.
- 사용자의 `~/.codex`, `~/.agents`, 사용자 프로젝트의 `.codex/` 파일을 moai 가 지우는 일: 어느 조각도 하지 않아요(D-3). 이 저장소 자신이 추적하는 `.codex/agents/*.toml` 두 개와 `.codex/` 폴더는 사용자 파일이 아니라 소스라서 조각 3 이 지워요(리더 결정 2026-11-21, D-8).
- 카드와 SPEC 의 글을 지우는 일: 하지 않아요. 언젠가 되살릴 일이라 철회 상태로 남겨요(D-4).
- GLM 지원을 줄이는 일: 하지 않아요. GLM 은 Claude Code 실행기의 모델 선택이라 그대로예요.
- Claude Code 쪽 동작을 바꾸는 일: 하지 않아요. 모든 조각이 Claude·GLM 시험을 그대로 통과해야 해요.

## Requirements

- REQ-1: The inventory shall list every tracked file of the base commit that holds `codex` in any case or the word `omp`, one row per file with its area, owning slice and mentioning line numbers. · failure: none
- REQ-2: When the file lists of `git grep` and `git ls-tree` over the base commit are joined, the result shall hold the same paths as the inventory table. · failure: none
- REQ-3: Every inventory row shall name exactly one owning slice, which is `1`, `1c`, `2`, `3`, `4`, `1>2` or `keep`. · failure: none
- REQ-4: When two slices need the same file, the slice order shall name the slice that edits it first. · failure: none
- REQ-5: When a removal slice is finished, it shall pass the full unnarrowed check with every Claude Code and GLM test green. · failure: none
- REQ-6: The removal shall delete no file in the user's Codex home folders, nor in a project's `.codex` folder. · failure: none
- REQ-7: When `moai update` previews a project or home that still holds Codex files, it shall print one line saying moai no longer manages them. · failure: none
- REQ-8: The SPEC folders and the cards that describe Codex or OMP work shall stay in the repository, withdrawn but not deleted, and the restore record names them with the commits and the steps that bring the work back. · failure: none
- REQ-9: When every slice has landed, a repo scan shall fail on `codex` or `omp` in a shipped path outside the kept folders and the allowed lines. · failure: none
- REQ-10: IF any lane session still runs on Codex, THEN the slice that removes the Codex session-id and lane-loop code shall not start. · failure: none
- REQ-11: Each slice shall land as one merge into the local integration branch, so reverting that merge brings its removal back. · failure: none
- REQ-12: When the plan names a writer that the `moai update` preview or the installer still uses for Codex or OMP files, the plan shall give that writer one owning slice with its source lines. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): 기준 커밋의 `git` 명령 셋을 합친 경로가 606개이고 표의 경로와 한 줄도 다르지 않아요 · verify: `cd <저장소>; { git grep -Iil codex be62c06a39; git grep -Iiwl omp be62c06a39; git ls-tree -r --name-only be62c06a39 | grep -iE 'codex|(^|[/_.-])omp([/_.-]|$)'; } | sed 's/^[^:]*://' | sort -u > /tmp/union.txt; wc -l < /tmp/union.txt; grep -v '^#' .moai/specs/SPEC-20261121-remove-codex-support/inventory.tsv | cut -f3 | sort | comm -3 - /tmp/union.txt | wc -l` 이 606 과 0 을 찍어요
- AC-2 (REQ-1): `codex` 를 담은 파일은 584개예요(여기에 `OMP` 낱말만 담은 19개와 경로 이름에만 Codex 가 있는 3개를 더해 AC-1 의 606) · verify: `git grep -Iil codex be62c06a39 | wc -l` 이 584 를 찍어요
- AC-3 (REQ-3): 모든 행의 소유 조각이 허용 값이에요 · verify: `awk -F'\t' 'substr($1,1,1)!="#" && $2 !~ "^(1|1c|2|3|4|1>2|keep)$"' .moai/specs/SPEC-20261121-remove-codex-support/inventory.tsv | wc -l` 이 0 을 찍어요
- AC-4 (REQ-4): 두 조각이 쓰는 파일은 `1>2` 세 개(`src/host_home.rs`, `src/help.rs`, `src/main.rs`)뿐이에요 · verify: `awk -F'\t' 'substr($1,1,1)!="#" && $2=="1>2" {print $3}' .moai/specs/SPEC-20261121-remove-codex-support/inventory.tsv` 가 그 세 경로를 찍어요
- AC-5 (REQ-8): 복원 기록에 Codex·OMP 를 담은 SPEC 폴더 66개가 있어요 · verify: `wc -l < .moai/specs/SPEC-20261121-remove-codex-support/restore-record.txt` 가 66 을 찍어요
- AC-6 (REQ-7): 사용자에게 보일 한 줄이 SPEC 에 있어요 · verify: `grep -c "moai no longer manages Codex files" .moai/specs/SPEC-20261121-remove-codex-support/spec.md` 가 1 이상이에요
- AC-7 (REQ-9, REQ-10, REQ-11): 조각 순서 표에 조각 1·2·3·4·5 와 시작 조건이 있어요 · verify: `grep -c "^| 조각 [1-5] " .moai/specs/SPEC-20261121-remove-codex-support/spec.md` 가 5 를 찍어요
- AC-8 (REQ-12): 쓰는 곳 표에 설치기까지 포함한 쓰기 9곳이 각각 소유 조각을 가져요 · verify: `grep -c "^| W-[0-9] " .moai/specs/SPEC-20261121-remove-codex-support/spec.md` 가 9 를 찍어요

## Scope

- write: 이 SPEC 폴더(`spec.md`, `inventory.tsv`, `restore-record.txt`)
- read-only: 저장소 전체(`git grep`·`git ls-tree` 만)
- forbidden: 코드·테스트·문서·스킬·스크립트·설치 파일 수정, 카드 상태 변경(카드 철회는 리더가 해요, D-4)

### Inventory

`inventory.tsv` 는 한 파일이 한 줄이에요: 영역, 소유 조각, 경로, 줄 수, 줄 번호 목록(`경로:줄` 이 곧 한 번의 언급). 이 표는 스크립트로 한 번 만든 기록 데이터이고, 그 스크립트는 남기지 않아요(이 저장소의 도구는 Rust 로 옮기고 새 Python 은 더하지 않는다는 사용자 결정, D-1). 다시 확인하고 싶으면 AC-1·AC-2 의 `git` 명령을 돌리면 돼요. 줄 수가 0 인 행은 경로 이름에만 Codex 가 있는 파일이에요.

#### 영역별 (기준 커밋 be62c06a39)

| 영역 | 파일 | 줄 | 무엇이 들어 있나 |
| --- | --- | --- | --- |
| launcher-and-session | 5 | 63 | OMP 실행기와 세션 인자 만들기(`src/launcher_init.rs`·`launcher_terminal.rs`·`src/session.rs`)에 남은 줄, 현재 세션 찾기(`CODEX_THREAD_ID`, `src/session_current.rs`). `moai codex` 는 이미 "removed" 안내로 바뀌었어요 |
| home-layer | 19 | 122 | 홈 업데이트의 Codex 단계: `~/.codex` 경로(`src/host_home.rs`), 라우팅 블록(`home_routing.rs`), 스타일(`home_style.rs`), `update*.rs`, 구 스킬(`skills_legacy.rs`) |
| hooks | 6 | 204 | Codex 훅 어댑터와 훅 파일 쓰기(`src/hooks.rs`, `hook_engine.rs`, `hook_guards.rs`, `hook_legacy.rs`, `hook_config.rs`, `update_hooks.rs`) |
| project-layer | 14 | 320 | 프로젝트 단계: `.codex/` 분류(`go_state.rs`), 호스트 단계(`project_hosts.rs`), 생성(`generation.rs`, `agent_defs.rs`, `agent_install.rs`), 프로필(`profile.rs`), 모델·노력 질문 |
| statusline | 2 | 2 | 상태줄 한두 줄 |
| codex-only-files | 32 | 807 | 이름부터 Codex 인 파일: `src/hosts/codex.rs`, `src/hosts/codex_config.rs`, `src/ops/codex_*.rs` 5개, 저장소가 추적하는 `.codex/agents/prj-local-check.toml`·`prj-reviewer.toml`(내용에 낱말 없음, 0줄), `docs/design-review/codex-statusline-official-scope-2026-11-10-sources.json`(0줄), `tests/codex_*`, `tests/fixtures/hooks-pre-p4/codex*`, `docs/**/codex*` |
| other-src | 6 | 28 | 위에 안 든 소스(`src/omp_agents.rs`, `src/hosts/mod.rs` 등)와, 이미 착지한 호환 층(`src/retired_hosts.rs`, `src/hosts/config_backup.rs`: 은퇴 호스트를 안전하게 읽는 코드라 남겨요) |
| lane-and-card-tooling | 29 | 164 | 개발 도구: 레인 루프의 Codex Stop 훅·`/goal` 한도(`src/lane_loop.rs`), 카드 세션 ID(`CODEX_THREAD_ID`, `src/card/dispatch.rs`), `lane_*`, `leader_*`, 결정·트레일러·감사·게이트 |
| cli-surface | 2 | 115 | 명령 선언과 도움말(`src/main.rs`, `src/help.rs`): 호스트 값 `cc\|glm\|codex\|omp`, `codex-*` 명령, 훅 계획 `--host` |
| tests-and-fixtures | 192 | 2055 | 시험과 고정 파일(`tests/**`) |
| skills-and-agent-text | 43 | 129 | `skills/moai/**`(SKILL.md·workflows·references), `templates/**`(번들 팩, GitHub 템플릿) |
| docs | 92 | 924 | `docs/**`, `README*`, `AGENTS.md`, `CONTRIBUTING.md`, `SECURITY.md`, `SUPPORT.md` |
| install-ci-tools | 21 | 65 | `install/*`, `.github/*`, `scripts/*`, `tools/*`, `schemas/*`, `Cargo.toml`, `experiments/*` |
| project-state-files | 13 | 611 | 생성되는 매니페스트(`.moai-cli/manifests/**`)와 이 저장소의 프로필(`.moai/project.yaml`) |
| keep-restore-record | 130 | 976 | 남기는 것: `.moai/specs/**`(SPEC 66개 폴더의 글), `.moai-cli/reports/**`(지난 이슈 본문·보고서) |
| 합계 | 606 | 6585 | |

#### 소유 조각별

| 조각 | 파일 | 줄 | 그 가운데 src | tests | 담당 영역 |
| --- | --- | --- | --- | --- | --- |
| 1 | 76 | 791 | 17 | 59 | 홈 업데이트 쓰기(home-layer 의 `update.rs`·`home_routing`·`home_style`·`home_install`·`skills_legacy` 등)와 에이전트 렌더러·설치(`agent_defs`·`agent_install`), 그 시험과 update 고정 파일. `host_home`·`update_cleanup`·`user_assets`·`migrate` 는 제외 |
| 1c | 8 | 150 | 3 | 2 | OMP 실행기와 설치기에 남은 줄: `src/launcher_init.rs`, `launcher_terminal.rs`, `session.rs`, `tests/omp_launcher.rs`, `tests/launcher_agent_entry.rs`, `install/install.sh`, `install.ps1`, `install/README.md`. 조각 1 과 따로 갈 수 있어요(다른 파일) |
| 1>2 | 3 | 149 | 3 | 0 | `src/main.rs`, `src/help.rs`, `src/host_home.rs`(조각 1 이 쓰는 곳을 먼저 지우고, 끝나면 조각 2 가 호스트 variant·명령·도움말 줄을 지워요) |
| 2 | 92 | 1483 | 32 | 57 | hooks, project-layer, statusline, codex-only 파일(src·tests), other-src, `audit.rs`·`contract.rs`, 프로필 스키마, 이 저장소 프로필, 프로젝트 업데이트 시험 |
| 3 | 240 | 2278 | 1 | 65 | 스킬·템플릿 글, 문서, 설치 말고 CI·도구, `labels_check.rs`, 생성 매니페스트, 이 저장소의 `.codex/` 폴더(추적 파일 2개, 리더 결정), 문서·도구를 보는 시험 |
| 4 | 38 | 200 | 25 | 13 | lane·goal·leader 와 카드 도구(`lane_*`, `leader_*`, `goal_block`, 카드 세션 ID `CODEX_THREAD_ID`: `session_current`, `card/*`, `trailers`)와 그 시험 |
| keep | 149 | 1534 | 9 | 10 | 손대지 않아요: 지난 기록·SPEC·리포트 130개 파일과, 남기기로 한 코드·시험 19개(아래) |

`keep` 의 코드·시험 19개: `src/go_state.rs`, `src/update_cleanup.rs`, `src/worktree_gc.rs`, `src/gate.rs`, `src/decision.rs`, `src/user_assets.rs`, `src/migrate.rs` 와 `tests/go_state_codex_cli.rs`, `go_state_classes_cli.rs`, `go_state_cli.rs`, `update_cleanup_cli.rs`, `gate_check_cli.rs`, `migrate_cli.rs`, `worktree_gc_install_dir.rs`, `decision_append_cli.rs`(여기까지가 D-15 의 15개), 그리고 이미 착지한 호환 층 4개: `src/retired_hosts.rs`, `src/hosts/config_backup.rs`, `tests/retired_hosts_layer.rs`, `tests/codex_launcher_removed.rs`(D-20). 이 파일들은 어느 제거 조각도 지우지 않아요. 낱말은 그대로 남으므로 마지막 낱말 검사의 허용 목록에 들어가요(D-15, D-20).

수 맞추기: 파일 76+8+3+92+240+38+149 = 606, src 17+3+3+32+1+25+9 = 90, tests 59+2+0+57+65+13+10 = 206(나머지 310개는 src·tests 밖 파일이에요).

시험은 먼저 `src` 의 소유를 따르는 첫 추정으로 정했고(표를 만든 스크립트가 시험 이름의 낱말로 갈랐어요), 소유 열은 `inventory.tsv` 의 두 번째 열이 기록으로 남아 있어요. 규칙은 제거 영향 보고서 §2·§6 의 카드 배정을 따라요. 부르는 API 를 지우는 조각이 그 시험도 **같은 커밋에서** 고쳐요(순수 삭제만 하는 시험은 아무 때나 지워도 컴파일은 깨지지 않아요). 이름만으로 틀리게 정해진 시험은 그 조각이 `cargo test --no-run` 에서 깨지는 것으로 알게 되고, 소유를 바꾸면 그 조각의 Result 에 적어요.

## Plan

병합 순서는 제거 영향 보고서 §6 과 리더 결정을 따라요: 이 SPEC → 호환 층 → 조각 1 → 조각 2 → 조각 3 → 새 binary 로 이 저장소 자신의 프로필·`generation.json` 다시 만들기 → `clippy --all-targets` → 조각 4 → 조각 5. 한 단계가 착지해 통합 가지의 tip 이 초록인 것을 확인한 뒤에 다음 단계가 병합해요.

| 단계 | 카드 | 무엇 | 언제 시작해요 |
| --- | --- | --- | --- |
| 호환 층 | C-312 | **착지했어요**(SPEC-20261121-retired-hosts-read-safely, 기준에 들어 있어요). 더하기만 하는 층(기존 동작은 그대로): 프로필의 `codex`·`omp` 를 읽을 때 걸러요(D-10), `launcher.models.codex/omp` 를 은퇴 키로 읽고 무시해요(D-11), 옛 receipt 의 은퇴 호스트 항목은 되돌릴 때 건너뛰고 보고해요(D-12), 생성 manifest 의 Codex 항목은 manifest 에서만 빼요(D-13). 아울러 `codex_config` 의 `Backup`·`BACKUP_DIR`·`write_backup` 과 `CONFIG_LIMIT` 를 공유 모듈로 먼저 옮겨요(`src/hooks.rs`·`src/go_state.rs` 가 쓰기 때문이에요) | 이 SPEC 이 통합 가지에 병합된 뒤 가장 먼저 |
| 조각 1 | C-290 | 둘로 나뉘었어요. **1a 는 착지했어요**(SPEC-20261121-codex-launcher-and-models-removed): `moai codex` 가 "removed" 안내 한 줄과 종료 2 를 돌려주고, 모델 표·질문에서 Codex 가 빠졌어요. **1b 가 남았어요**(소유 열 `1`): 첫 커밋은 이미 빨간 시험 둘을 은퇴시키는 일이에요(D-22). 그다음 홈 업데이트 쓰기 — `update.rs` 의 `HOSTS`·`skill_host_of`·단계 목록, `agent_defs`·`agent_install` 을 **한 커밋에서**, `home_routing`·`home_style`·`home_install`·`skills_legacy`·`update_tux` 와 그 시험·update 골든. 은퇴 호스트의 receipt 는 건너뛰고 `skipped[]` 안내 한 줄을 plan_digest 밖에 내요(D-12). 쓰는 곳 W-1~W-5 | 호환 층·1a 착지 뒤(지금 시작할 수 있어요) |
| 조각 1c | C-316 | 소유 열 `1c`: OMP 실행기(`launcher_init`·`launcher_terminal`·`session` 의 OMP 줄, `tests/omp_launcher.rs`, `tests/launcher_agent_entry.rs`)와 설치기(`install/install.sh`·`install.ps1`·`install/README.md` 의 호스트 목록과 다음 단계 안내). 쓰는 곳 W-9 | 조각 1 과 파일이 겹치지 않아 따로 갈 수 있어요 |
| 조각 2 | C-291 | 프로젝트 단계(`project_hosts`·`project_update`·`project_skills`·`generation`)·훅(`hooks`·`hook_engine`·`hook_legacy`)·`omp_agents`·`hosts/codex*`·`ops` 9개 삭제, 같은 커밋에 컴파일이 깨지는 시험 줄·Codex 전용 시험 파일·`repo_scans.json`·`ci_scope.py`·`cli_text_citations.rs` 항목, 마지막에 `host_home` 의 Codex·OMP variant. 쓰는 곳 W-6~W-8. 이미 있는 브랜치 `spec/20261121-remove-codex-project-layer` 의 일을 이어받을 수 있어요 | 조각 1(1b) 착지 뒤(`src/main.rs`·`src/help.rs` 와 `src/host_home.rs` 는 조각 1 이 반환한 뒤에 잡아요) |
| 조각 3 | C-292 | 스킬 글 → `rule-history`·`source-manifest` 다시 만들기 → 글 시험, 문서+파이썬 시험+src 배열, 설치 스크립트+시험, GitHub 양식+manifest+labels(`host:codex` 는 `not_used` 로 옮겨요), README+`Cargo.toml`, bundle-packs 14개+`packs.json`+`.gitignore` 를 한 커밋에, 이 저장소의 `.codex/` 폴더 | 조각 2 착지 뒤(글만 하는 준비는 먼저 해도 되지만 병합은 뒤예요) |
| 재생성 | 조각 3 의 카드가 이어서 | 새 binary 로 이 저장소 자신의 프로필(`.moai/project.yaml`, `.moai-cli/project.yaml`), `generation.json`·`profile-baseline.json` 과, 그때 착지해 있는 관리 파일(사용자 결정 2026-11-22: moai 는 `CLAUDE.md`·`CLAUDE.local.md` 를 쓰고, 그 일은 C-328 이 맡아요) 다시 만들기. 이 단계는 `AGENTS.md` 를 새로 만들거나 키우지 않아요. 그다음 `cargo clippy --all-targets -D warnings`(죽은 코드) | 조각 3 착지 뒤 |
| 조각 4 | C-313 | 레인·goal·leader 의 Codex 경로(`lane_loop` 의 Codex·OMP arm, `lane-stop-hook` 명령, `lane_duty`, `goal_block` 은 Claude Code 가 쓰는 경우만 남겨요, `leader_round`)와 카드 세션 ID(`CODEX_THREAD_ID`) | 재생성 뒤, 그리고 Codex 에서 도는 레인이 하나도 없다고 리더가 확인한 뒤(REQ-10, D-6) |
| 조각 5 | 새 카드 필요 | 마지막 검사: 저장소 검사에 `codex`·`omp` 가 허용 목록 밖에 없다는 규칙을 더해요. 허용 목록 = D-15 로 남긴 파일, D-16 의 번들 자산, 호환 층이 남긴 읽기 코드, 사용자에게 보이는 한 줄 | 조각 4 착지 뒤 |

각 조각은 `--no-ff` 병합 한 번으로 통합 가지에 들어가요(REQ-11). 낱말 검사(조각 5)를 먼저 켜면 아직 안 지운 조각이 모두 실패하므로, 검사는 마지막이에요. `build.rs` 는 `src/ops/*.rs` 와 `packs.json` 을 자동 등록하므로 파일을 지우는 커밋이 도움말 표·`repo_scans.json`·`ci_scope.py`·`cli_text_citations.rs`(`CLEANED` 17→16)·`packs.json` 을 같은 커밋에서 고쳐요. GLM 이 쓰는 코드(모델 표의 `glm` arm, `glm_token`·`KeySource`, `launcher_terminal` 의 GLM 상수, 세션의 `zai/` 접두어 떼기)는 어느 조각도 지우지 않아요(D-9, 보고서 §4).

조각 1 과 2 의 경계에서 보고서(§2.2)와 리더 노트가 `update.rs` 를 서로 다른 조각에 적어 두었어요. 이 SPEC 은 보고서 표대로 `update.rs`·`agent_defs.rs`·`agent_install.rs` 를 조각 1 에 두되, 리더가 조각 2 로 옮기면 `inventory.tsv` 의 소유 열을 바꿔요(D-19).

### 지금 쓰는 곳 (기준 커밋 be62c06a39)

사용자가 정한 "새 로컬 판은 Codex·OMP 쓰기를 걷어낸 뒤에 설치"를 지키려면, 아래 아홉 곳이 모두 없어져야 해요. 줄 번호는 기준 커밋에서 `git grep -n -i "codex\|\bomp\b"` 로 읽은 것이고, 소유 조각은 위 조각 표와 `inventory.tsv` 의 소유 열이에요(REQ-12).

| Id | 쓰는 곳 | 무엇을 쓰나 | 소유 조각 | 줄(기준 커밋) |
| --- | --- | --- | --- | --- |
| W-1 | 홈 스킬 복사 | `~/.agents/skills/moai*`(Codex 와 OMP 가 읽는 사본), `update.rs` 의 `HOSTS`·`skill_host_of`·단계 목록 | 조각 1(1b) | `src/update.rs` 46, 57, 1515-1531, 2117, 3030-3034, `src/skills_legacy.rs` 1-25, 153, 167, 194 |
| W-2 | 홈 라우팅 블록 | `~/.codex/AGENTS.md` 의 `moai-routing` 블록 | 조각 1(1b) | `src/home_routing.rs` 6, 40, 45, 46, 127 |
| W-3 | 홈 스타일 블록 | `~/.codex/AGENTS.md` 의 스타일 블록(`block_plan`, 되돌리기, `same_writes`) | 조각 1(1b) | `src/home_style.rs` 6, 7, 35-45, 132, 271, 379-381, 441, 469, 760, 872-886 |
| W-4 | 홈 Codex 에이전트 | `~/.codex/agents/*.toml`(`render_codex`) | 조각 1(1b) | `src/agent_defs.rs` 29-31, 320-351, `src/agent_install.rs` 32-35, 51, 145-147 |
| W-5 | 홈 OMP 에이전트 | `~/.omp/agent/agents`(`render_omp`) | 조각 1(1b) | `src/agent_defs.rs` 353-402, 408-565, `src/agent_install.rs` 32-35 |
| W-6 | 프로젝트 훅·설정 | `.codex/hooks.json`, `.codex/config.toml`, `.codex/hooks/moai/`, 훅 어댑터와 훅 파일 쓰기 | 조각 2 | `src/hooks.rs`, `src/hook_engine.rs`, `src/hook_legacy.rs`, `src/hook_guards.rs`, `src/hook_config.rs`, `src/update_hooks.rs` 414, `src/hosts/codex*.rs`, `src/ops/codex_*.rs` |
| W-7 | 프로젝트 Codex 스킬·에이전트 | `.agents/skills`·`.codex/agents` 투영, 생성 manifest 항목 | 조각 2 | `src/project_hosts.rs`, `src/project_skills.rs`, `src/project_update.rs`, `src/generation.rs`, `src/agents.rs` |
| W-8 | 프로젝트 OMP 단계 | OMP 에이전트·훅 단계와 `omp_agents.rs` | 조각 2 | `src/omp_agents.rs`, `src/project_hosts.rs`, `tests/fixtures/hooks-pre-p4/omp.*` |
| W-9 | 설치기 | `install.sh`·`install.ps1` 이 `claude codex omp` 를 찾고 "다음 단계" 줄에 `moai codex`·`omp` 를 안내 | 조각 1c | `install/install.sh` 27, 437, 459, 582-586, `install/install.ps1` 26, 288, 328, 611-616, `install/README.md` 28 |

W-1~W-5 는 `moai update` 의 홈 미리보기와 적용이 곧바로 쓰는 곳이에요. 조각 1(1b) 하나가 지워야 하고, 시험을 먼저 만들어서(은퇴 호스트가 홈에 있어도 Claude·GLM 단계가 그대로 초록이고 Codex·OMP 파일은 한 바이트도 안 바뀌는 시험) 시험이 실패하는 것을 본 뒤에 지워요. W-6~W-8 은 프로젝트 업데이트가 쓰는 곳이라 조각 2 예요. 어느 한 곳이 남으면 새 local 판을 설치할 수 없어요(REQ-12).

### 조각마다 지켜야 하는 것

- 먼저 그 조각의 호출자를 지운 뒤에 호출되던 모듈을 지워요(그래야 중간에도 빌드가 돼요). 다른 조각의 파일에서 그 모듈을 부르는 곳이 나오면 그 조각이 비워 둘 때까지 모듈을 두고, 비운 조각이 지워요. 어느 조각이 지웠는지는 각 조각의 Result 에 파일 목록으로 적어요.
- 걷어낸 뒤 `cargo test --locked --offline`(전체)·`clippy --all-targets -D warnings`·저장소 검사 41개·파이썬 시험이 모두 통과해야 해요(줄이지 않은 검사). Claude Code·GLM 시험은 하나도 지우지 않아요.
- 도움말 고정 파일(golden)과 `skills/moai/references/source-manifest.json` 은 병합 직전에 다시 만들어요(`.moai-cli/manifests/**` 도 같아요).

### 사용자에게 보이는 한 줄

`moai update` 의 미리보기(홈과 프로젝트 모두)에서 Codex 파일이 남아 있는 곳마다 한 줄을 보여요. 파일은 건드리지 않고, 쓰지 않으면 사용자가 직접 지울 수 있다고 알려요.

- English: `moai no longer manages Codex files (~/.codex, ~/.agents/skills, <project>/.codex/). They are left as they are; remove them yourself if you do not use them.`
- 한국어: `moai 는 이제 Codex 파일(~/.codex, ~/.agents/skills, <프로젝트>/.codex/)을 관리하지 않아요. 그대로 두었어요. 쓰지 않으면 직접 지우세요.`

moai 가 쓰던 경로는 기준 커밋의 `git grep` 으로 확인한 것이에요: `.codex/config.toml`, `.codex/hooks.json`, `.codex/AGENTS.md`(홈), `.codex/agents/`, `.codex/hooks/moai/`, `~/.agents/skills/moai*`, 환경 변수 `CODEX_HOME`·`CODEX_THREAD_ID`.

### 복원 기록

Codex 쪽 일을 언젠가 되살릴 때를 위해 아래를 남기고, 어느 것도 지우지 않아요(REQ-8).

1. **SPEC 폴더 66개**: `restore-record.txt`(한 줄에 하나)에 있어요. 기준 커밋에서 `codex` 나 `OMP` 를 담은 `.moai/specs/*` 폴더 전체이고, 이름이 Codex·OMP 인 것은 `SPEC-20261118-codex-question-tool`, `SPEC-20261119-codex-conflict-choices`, `SPEC-20261118-omp-host`, `SPEC-20261119-omp-host`, `SPEC-20261118-omp-launcher`, `SPEC-20261118-profile-host-omp` 예요. 이 폴더들은 삭제하지 않고 `status` 도 손대지 않아요(SPEC 상태 변경은 이 일의 범위 밖이에요).
2. **카드**: 제목에 `codex`(대소문자 무관)나 낱말 `OMP` 가 든 카드는 2026-11-21 의 `moai card list`(전체 287장)에서 20장이에요: C-8(ready), C-22(draft), C-29(draft), C-38(draft), C-52(done), C-56(done), C-57(ready), C-79(done), C-89(done), C-240(ready), C-243(ready), C-253(review), C-259(ready), C-260(ready), C-264(ready), C-271(active), 그리고 이 일의 C-289~C-292. 제목으로는 Codex 만 다루는 카드가 C-52(done)·C-253(review)·C-271(active)이고, 나머지는 Claude Code 와 함께 Codex 를 말해요. 카드는 원장에 그대로 두고, 아직 시작하지 않은 카드를 철회할지는 카드마다 리더가 정해요(D-4). 철회는 `ready` 카드를 `hold` 로 옮길 뿐 지우지 않아요.
3. **코드**: 제거가 시작되기 전 커밋 `53cce6c580d8a1b59aee04376ad7459f81b096d1` 에 Codex 코드 전체가 있어요(revision 1 의 기준이에요). 호환 층과 런처·모델 조각은 그 뒤에 `--no-ff` 병합 하나씩으로 착지했고, revision 2 의 기준 `be62c06a3991b96f12a1fdb133b0190ca514e1dc` 는 그 둘을 포함해요. 각 조각은 `--no-ff` 병합 한 번이라, 조각의 병합 커밋을 역순으로 `git revert -m 1 <병합 커밋>` 하면 그 조각이 돌아와요. 병합 커밋 sha 는 조각이 착지할 때 그 조각 SPEC 의 Result 에 적어요. 되살릴 때는 이 순서를 거꾸로(조각 5 → 1) 되돌려요.
4. **기준 가지(결정 요청)**: 통합 가지가 크게 앞서 가도 기준 커밋을 찾기 쉽도록, 리더가 그 커밋에 이름 있는 가지 한 개를 만들어 두면 좋아요(D-5). 태그가 아니라 가지예요.

## Risks

- 카드 세션 ID 는 `CODEX_THREAD_ID` 도 읽어요(`src/session_current.rs`, `src/card/dispatch.rs` 의 holder, `src/trailers.rs`). 지금 일부 레인이 Codex 로 도는 중이면, 그 코드를 지우는 순간 그 레인은 카드를 잡거나 놓을 수 없어요. 그래서 조각 4 를 마지막으로 미뤘고 시작 조건에 "Codex 레인 없음"을 넣었어요(D-6, D-18).
- 호환 층(C-312)이 먼저 착지하지 않고 조각 1 이 `launcher.models.codex` 를 검증 대상에서 빼면, 이미 GPT 를 고른 사용자의 `moai cc`·`moai glm` 과 update 의 모델 단계가 거부로 멈춰요(고칠 길까지 같이 막혀요). 그래서 호환 층이 첫 단계예요(D-11).
- 이 SPEC 의 소유 열은 제거 영향 보고서(읽기 전용 조사, 빌드·시험은 돌리지 않았어요)의 카드 배정을 따른 추정이에요. 각 조각이 첫 빌드에서 틀린 곳을 알게 되면 그 조각의 Result 에 소유 변경을 적고 `inventory.tsv` 의 소유 열을 고쳐요.
- 기준 커밋(통합 tip)에서 `tests/retired_hosts_layer.rs` 의 시험 둘이 이미 빨개요: `model_settings::while_the_tools_are_active_a_malformed_value_is_refused_as_before`(180줄)와 `model_settings::saving_a_choice_while_the_tools_are_active_keeps_every_tool_key`(224줄). "도구가 켜진 홈"(`TOOLS`)에서는 아직 `codex`·`omp` 키를 지켜야 한다고 기대하는데, 런처·모델 조각이 모델 표에서 Codex 줄을 지워서 그 길이 더는 닿지 않기 때문이에요(리더가 통합 tip 에서 30 passed·2 failed 로 쟀고, 이 가지에서도 같은 결과예요). 이 파일은 `keep`(D-20)이지만 이 두 시험은 조각 1(1b)이 첫 커밋에서, 시험을 먼저 손봐서 고치거나 켜진 도구의 길과 함께 은퇴시켜요(D-22).
- `src/main.rs`·`src/help.rs`·`src/host_home.rs` 는 조각 1 과 2 가 모두 건드려요. 한꺼번에 고치면 충돌이라 소유를 `1>2` 로 나눴어요(조각 1 이 반환하면 조각 2 가 잡아요).
- 프로필에 이미 `codex`·`omp` 가 적힌 프로젝트: 읽을 때 무시하고 안내 한 줄만 보이게 해요(D-2). 그러면 프로필 스키마와 그 읽기 코드에 `codex` 낱말이 한 곳 남아서, 조각 5 의 허용 목록에 그 한 곳과 사용자에게 보이는 한 줄을 넣어요.
- 시험 소유를 이름의 낱말로 처음 정해서 틀리게 정해진 시험이 있을 수 있어요. 그 조각이 첫 빌드에서 알게 되고 Result 에 소유 변경을 적어요.
- 생성된 매니페스트와 도움말 고정 파일은 병합 순서에 따라 해시가 달라져요. 그래서 각 조각이 병합 직전에 다시 만들어요(조각 3 만의 일이 아니에요).
- `.agents/skills` 는 Codex 와 OMP 말고 다른 도구도 읽을 수 있는 폴더예요. 그래서 moai 가 그 폴더를 지우지 않고, 안내 한 줄만 보여요(REQ-6).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 목록(`inventory.tsv`)은 기준 커밋의 파일 내용에서 스크립트로 한 번 만든 기록 데이터이고, 그 스크립트는 남기지 않아요. 소유 조각은 영역 규칙과 시험 이름으로 정했어요. 다시 확인하는 길은 AC-1·AC-2 의 `git` 명령이에요 | 사용자 결정(이 저장소의 도구는 Rust 로 옮기고 새 Python 은 더하지 않아요, 리더 검토 2026-11-22). 손으로 세면 카드의 숫자(src 86·tests 198)와 어긋나기 쉬워서 같은 낱말로 셀 수 있게 했어요. OMP 낱말만 담은 19개 파일은 카드 숫자에 들지 않아 따로 더했어요 | 규칙 바꾸기 |
| D-2 | 프로필에 적힌 `codex`·`omp` 는 읽고 무시하며 안내 한 줄을 보여요. 조각 2 가 구현해요 | 카드 본문(조각 2)과 리더 결정(2026-11-21: 기존 프로필의 hosts 에 codex/omp 가 있어도 읽혀야 해요)이 정한 방식이에요 | 거절로 바꾸기 |
| D-3 | 사용자의 `~/.codex`, `~/.agents/skills`, 프로젝트 `.codex/` 는 moai 가 지우지 않아요. 관리만 멈추고 안내 한 줄만 보여요 | 사용자 결정(카드 본문) | 없음 |
| D-4 | 카드와 SPEC 은 철회해서 두고 지우지 않아요. 카드 철회는 리더가 해요 | 사용자 결정 "언젠가 codex 가 쓸만해지면 사용하도록" | 철회 풀기 |
| D-5 | 기준 커밋에 이름 있는 가지(태그 아님) 하나를 리더가 만들어 두기를 요청해요. 기본값은 만들지 않음 | 병합 커밋 되돌리기만으로도 복원은 되지만, 기준 커밋을 찾는 일이 쉬워져요. 가지를 밀어 올리는 일은 레인이 하지 않아요 | 만들지 않기 |
| D-6 | 조각 4(카드 세션 ID, 레인 루프 Codex 경로)는 Codex 에서 도는 레인이 없다고 리더가 확인한 뒤에 시작해요 | 레인 도구가 Codex 세션 ID 를 읽는 동안 지우면 그 레인이 카드를 못 다뤄요. 기본값은 대기, 되돌릴 수 있고 분류는 none 이에요 | 조건 없이 시작 |
| D-7 | 카드: 호환 층 C-312, 조각 1 C-290, 조각 2 C-291, 조각 3 C-292, 조각 4 C-313(리더가 만들었어요). 조각 5(낱말 검사)만 아직 카드가 없어요 | 리더 노트 5221 | 조각 5 를 조각 4 에 합치기 |
| D-8 | 이 저장소가 추적하는 `.codex/agents/prj-local-check.toml`·`prj-reviewer.toml` 과 `.codex/` 폴더, `.worktreeinclude`·`.gitignore` 의 `.codex` 줄, 그 파일을 읽는 시험·매니페스트는 조각 3 이 지워요. 사용자 프로젝트의 `.codex/` 는 D-3 대로 건드리지 않아요 | 리더 결정(C-292 노트, 사용자 지시 2026-11-21). 저장소 안의 추적 파일은 사용자 파일이 아니라 소스예요 | 조각 3 병합 되돌리기 |
| D-9 | GLM 이 쓰는 코드는 Codex 와 함께 지우지 않아요. 호스트 값 `glm` 과 모델 선택 코드는 모든 조각에서 남아요. `omp_key`·`is_zai_model`·`OMP_KEY_VAR` 만 지워요 | 리더 결정(C-289 노트 5151, 5221). GLM 은 Claude Code 실행기를 거치는 모델 선택이라 Codex 제거와 따로예요(보고서 §4) | 없음 |
| D-10 | 프로필 `hosts` 의 `codex`·`omp`: 읽을 때 걸러 내요(거부하지 않아요). 거르고 나서 비면 `[claude]` 로 읽어요. 다음 저장 때 사라져요 | 리더 결정 D1(노트 5221, 보고서 §1-2·§5) | 거부로 바꾸기 |
| D-11 | `launcher.models.codex`·`omp`: 은퇴 키로 읽고 무시해요, 다음 저장 때 지워요(선례: `model_effort_questions.rs` 의 OMP 처리). 이게 가장 큰 위험이에요 — 없으면 `moai cc`·`moai glm` 과 update 의 모델 단계가 멈춰요(`session.rs` 의 `apply_saved_choice`, `update_models.rs`) | 리더 결정 D2 | 키를 거부로 바꾸기 |
| D-12 | 옛 receipt 의 은퇴 호스트 항목: 되돌릴 때 건너뛰고 건너뛴 것을 보고해요(`home_routing.rs`, `update_hooks.rs` 의 "unknown host" 거부 제거) | 리더 결정 D3 | 거부로 바꾸기 |
| D-13 | 생성 manifest 의 Codex 항목: manifest 에서만 빼고 파일은 그대로 두어요. "recorded path is not the path of its skill" 충돌이 나지 않아야 해요(`generation.rs`) | 리더 결정 D4 | 없음 |
| D-14 | `hook-dispatch --host codex\|omp`: 한 줄 "host removed" 오류를 내고 호출한 hook 에는 종료 코드 0 을 돌려줘요(충돌하지 않고, 무언가를 조용히 허용하지도 않아요) | 리더 결정 D5(노트 5221). 이 SPEC 의 번호는 D-10~D-18 이 리더 노트의 D1~D9 와 차례로 같아요(D-10=D1 … D-18=D9) | 조용한 통과로 바꾸기 |
| D-15 | 남기는 코드(리더 D6): Go 시절 `.codex` 분류(`go_state.rs`), `update_cleanup.rs` 의 `~/.codex/agents` 읽기 전용 감지, `worktree_gc.rs` 의 `~/.codex/worktrees` 보호, `gate.rs` 보호 경로, `decision.rs` 의 `Host::Codex` 와 `user_assets.rs` 의 slug(옛 기록을 읽는 데 써요). 이 파일과 그 시험 15개는 인벤토리의 `keep` 이고, 호환 층 카드만 `go_state.rs` 의 상수 한 줄을 고쳐요 | 리더 결정 D6, 보고서 §1-5·§2.3 | 일부를 지우기(사용자 기록이 안 읽힐 수 있어요) |
| D-16 | 번들 팩의 `.claude` 자산에 있는 Codex 언급(`templates/bundle-packs`, Go 바이트 스냅샷)은 바이트 그대로 두고 마지막 낱말 검사의 허용 목록에 올려요. 단, `.agents/skills` 14개와 `packs.json` 항목, `.gitignore` 블록은 조각 3 이 한 커밋에서 지워요(`build.rs` 가 `packs.json` 으로 표를 만들어서 따로 지우면 빌드가 실패해요) | 리더 결정 D7, 보고서 §2.4 | 자산도 고치기(Go 설치가 새로고침 대신 충돌이 돼요) |
| D-17 | `lane-loop-line`: Claude 만 남겨요(`/loop` 줄). Codex `/goal` 과 Stop hook 줄은 지워요. `goal_block` 은 Claude Code 가 쓰는 경우에만 남기고, 안 쓰면 조각 4(C-313)에서 지워요(Claude Code 문서로 확인하는 일은 그 카드의 몫이에요) | 리더 결정 D8 | 명령 삭제로 바꾸기 |
| D-18 | 레인·goal·leader 의 Codex 소스(`lane_loop`, `lane-stop-hook`, `lane_duty`, `goal_block`, `leader_round` 의 Codex 부분)는 새 카드 C-313(조각 4)이 맡아요. 이 SPEC 은 카드 세션 ID(`CODEX_THREAD_ID`)도 같은 조각 4 에 두어요 — 보고서 §2.3 은 프로젝트 조각에 적었지만, 지금 Codex 로 도는 레인이 있으면 지우는 순간 카드를 못 다루기 때문이에요(D-6) | 리더 결정 D9, 노트 5221 | 세션 ID 를 조각 2 로 옮기기 |
| D-19 | 조각 1 과 2 의 경계: `update.rs`·`agent_defs.rs`·`agent_install.rs` 는 조각 1 이에요(보고서 §2.2 표). 리더가 조각 2 로 옮기면 `inventory.tsv` 의 소유 열을 바꿔요. 같은 커밋에서 codex 스킬 호스트와 omp agent 호스트를 함께 지워요(`skill_host_of` 가 claude 가 아닌 모든 agent 호스트를 codex 사본에 묶어서 따로 지우면 안 돼요) | 보고서 §2.2·§2.3 과 노트 5221 사이의 차이를 숨기지 않으려고 | 소유 바꾸기 |
| D-20 | 이미 착지한 호환 층(`src/retired_hosts.rs`, `src/hosts/config_backup.rs`, `tests/retired_hosts_layer.rs`)과 런처 조각의 시험 `tests/codex_launcher_removed.rs` 는 은퇴 호스트를 일부러 이름으로 불러요(옛 프로필·receipt·설정을 안전하게 읽고, `moai codex` 안내를 보이려고). 그래서 인벤토리의 `keep` 이고 어느 제거 조각도 지우지 않으며, 마지막 낱말 검사의 허용 목록에 들어가요 | 두 SPEC(SPEC-20261121-retired-hosts-read-safely, SPEC-20261121-codex-launcher-and-models-removed)이 이미 착지했고, 지우면 D-10~D-14 가 깨져요 | 허용 목록에서 빼기(그러면 읽기 코드를 지워야 해요) |
| D-21 | 인벤토리의 소유 조각 `1` 을 둘로 나눠요: `1`(홈 업데이트 쓰기, 1b) 과 `1c`(OMP 실행기와 설치기). 두 곳은 파일이 겹치지 않아서 따로 착지할 수 있고, 새 local 판을 설치하려면 `1` 과 `2` 가 모두 필요해요(W-1~W-8). 설치기(W-9)는 설치 안내 문구라 사용자 데이터를 쓰지 않지만 옛 안내가 남으면 혼란을 주므로 같은 목표 안에 두어요 | 사용자 결정(새 local 판은 Codex·OMP 쓰기를 걷어낸 뒤 설치)과 리더 지시(C-289 → C-290 → C-291 순서) | `1c` 를 조각 3 으로 합치기 |
| D-22 | 기준 커밋에서 이미 빨간 시험 둘: `tests/retired_hosts_layer.rs` 의 `model_settings::while_the_tools_are_active_a_malformed_value_is_refused_as_before`(180줄, `unwrap_err` 가 `Ok` 를 만나요)와 `model_settings::saving_a_choice_while_the_tools_are_active_keeps_every_tool_key`(224줄, `models.codex` 가 null). 둘 다 Codex·OMP 도구가 켜진 홈(`TOOLS`)을 가정하는데, 런처·모델 조각이 모델 표에서 Codex 줄을 지웠으므로(`default_choice("codex")` 가 없어요) 그 길은 더 닿지 않아요. 그 파일과 `src/model_effort_settings.rs` 는 오늘 착지분에서 바뀐 적이 없고 `check affected` 가 그 대상을 고르지 않아서 못 잡았어요. **조각 1(1b)이 첫 커밋에서, 시험을 먼저 손봐서** 두 시험과 `TOOLS` 길을 함께 은퇴시키고(`active` 인자에서 도구가 켜진 경우를 지우거나, 두 시험을 "Codex 줄이 없다"는 새 기대로 고쳐요), 그 READY 노트에 두 이름을 적어요. 이 파일은 `keep` 이지만 이 두 시험의 수리만 예외예요 | 리더 노트(2026-11-22 측정: 통합 tip 에서 30 passed·2 failed)를 이 가지에서 `cargo test --locked --offline --test retired_hosts_layer` 로 다시 쟀어요: 30 passed, 2 failed, 같은 두 시험·같은 줄 | 두 시험만 `#[ignore]` 로 두기(빨간 채로 남는 것보다는 낫지만 지울 때 다시 와야 해요) |

## Result

이 SPEC 은 계획이라 코드는 바뀐 것이 없고(revision 1 과 2 모두), 아래는 revision 2 의 기준 커밋 `be62c06a3991b96f12a1fdb133b0190ca514e1dc` 에서 각 확인 명령을 실제로 돌린 결과예요(2026-11-22, `date -u` 로 읽은 날). revision 1 의 기준 `53cce6c580` 의 숫자(592 파일, 6503 줄, SPEC 폴더 57개)는 이 개정으로 바뀌었어요.

- AC-1 met: 위 `git` 명령 셋의 합집합이 `wc -l` 로 606, 표의 경로와 `comm -3` 로 비교한 차이가 0, 표의 데이터 행은 `grep -vc '^#'` 로 606. 내용에 낱말이 있는 파일은 603(`codex` 584 + `OMP` 낱말만 19)이고, 합집합에서 그것을 `comm -13` 으로 뺀 경로 이름만의 파일 셋은 `.codex/agents/prj-local-check.toml`, `.codex/agents/prj-reviewer.toml`, `docs/design-review/codex-statusline-official-scope-2026-11-10-sources.json` 이에요. 표의 줄 수 합 6585 와 영역별·조각별 합계는 표를 만든 스크립트가 낸 값이고, 표의 열을 `awk` 로 더해 다시 확인했어요(아래 보조).
- AC-2 met: `git grep -Iil codex be62c06a39 | wc -l` → 584. 같은 방법으로 센 `OMP` 낱말만 담은 파일은 19개(`comm -13` 으로 584개 목록과 겹치지 않는 것만 센 값), 경로 이름에만 Codex 가 있는 파일은 3개(`.codex/agents/*.toml` 2개와 `docs/design-review/codex-statusline-official-scope-2026-11-10-sources.json`)이고 584+19+3=606 이에요.
- AC-3 met: `awk` 로 허용 값 밖의 소유 조각 행을 센 값이 0.
- AC-4 met: `1>2` 행은 `src/host_home.rs`, `src/help.rs`, `src/main.rs` 세 개.
- AC-5 met: `restore-record.txt` 66줄(revision 1 보다 9개 늘었어요: guidance-sweep-core, leader-round-card-loop-signals, mode-guidance-rules, codex-launcher-and-models-removed, go-path-removal, leader-run-lease, retired-hosts-read-safely, rust-only-inventory, rust-only-text-scans).
- AC-6 met: `grep -c "moai no longer manages Codex files" spec.md` → 3(영어 한 줄과 SPEC 안에서 그 줄을 말하는 문장들).
- AC-7 met: 표의 `| 조각 1 ` … `| 조각 5 ` 행 5개(`조각 1c` 행은 따로 있어요).
- AC-8 met: `grep -c "^| W-[0-9] " spec.md` → 9, 모든 행이 소유 조각과 기준 커밋의 줄을 가져요(W-9 의 줄은 `git grep -n -i "codex\|\bomp\b" HEAD -- install/…` 로 다시 읽어 맞췄어요).
- 보조: 소유 조각별 파일은 1:76, 1c:8, 1>2:3, 2:92, 3:240, 4:38, keep:149 = 606. src 17+3+3+32+1+25+9 = 90, tests 59+2+0+57+65+13+10 = 206. 리더 결정 D1~D9(노트 5221)는 D-10~D-18, 이번 개정의 새 결정은 D-20(남기는 호환 층)·D-21(조각 1 을 1·1c 로 나눔)이에요.
- D-22 확인: 이 가지(기준 `be62c06a39` 위)에서 `CARGO_TARGET_DIR=… cargo test --locked --offline --test retired_hosts_layer` 를 heavy slot 안에서 돌렸더니 `test result: FAILED. 30 passed; 2 failed` 이고, 실패는 `saving_a_choice_while_the_tools_are_active_keeps_every_tool_key`(224줄)와 `while_the_tools_are_active_a_malformed_value_is_refused_as_before`(180줄)예요. 이 두 시험은 조각 1(1b)의 첫 커밋이 맡아요(D-22).
- 표 합계 확인: `grep -v '^#' inventory.tsv | awk -F'\t'` 로 소유 열별 파일·줄을 더하면 1:76/791, 1>2:3/149, 1c:8/150, 2:92/1483, 3:240/2278, 4:38/200, keep:149/1534, 합계 606/6585 예요.
- 리더 검토(2026-11-22, 3건) 반영: R1 `inventory.py` 를 지웠고 AC-1 과 D-1·D-19 를 `git` 명령과 표의 소유 열로 바꿨어요. R2 빨간 시험 둘을 Risks 와 D-22 에 이름·이유·주인(조각 1b)과 함께 적었어요. R3 재생성 단계는 `AGENTS.md` 대신 그때 착지해 있는 관리 파일(`CLAUDE.md` 결정)을 쓰고 `AGENTS.md` 를 만들거나 키우지 않는다고 적었어요.
- 기준 커밋 확인: `git rev-parse 46e831d6f8^{tree}` 와 `git rev-parse HEAD^{tree}` 가 모두 `4b6c5e304d93cfc53afc20a1247685612bd5bc5b`(가지 끝만 다시 만들어졌고 내용은 같아요).

돌리지 않은 것: 조각 1b·1c·2~5 의 코드 제거(각 조각의 SPEC 이 맡아요), 카드 철회, 기준 가지 만들기(리더의 몫, D-4·D-5), 영역·조각 규칙이 시험 이름의 낱말에 기대는 부분이 실제 빌드에서 맞는지(각 조각이 첫 빌드에서 확인해요), `cargo` 빌드·시험(이 SPEC 은 코드를 바꾸지 않아서 돌리지 않았어요; 저장소 검사 중 SPEC 을 읽는 것은 `spec check` 와 `check affected` 로 돌려요).
