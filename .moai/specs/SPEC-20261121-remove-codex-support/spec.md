---
id: SPEC-20261121-remove-codex-support
title: "Remove Codex support: Claude Code and GLM only, with the inventory, the slice order and the restore record"
status: draft
revision: 1
created: 2026-11-21
---

# Remove Codex support: Claude Code and GLM only

## Intent

사용자가 2026-11-21 에 결정했어요: "codex 지원을 철회한다. moai-adk codex 관련 자료 모두 제거하고 claude code, glm 만 지원하도록 하자. card, spec 은 언젠가 codex 가 쓸만해지면 사용하도록 하고…". 그래서 moai-adk 는 Claude Code(GLM 은 Claude Code 실행기를 거치는 모델 선택)만 지원하고, Codex 와 OMP(Oh My Pi, 같은 결정으로 함께 접어요)의 코드·스킬 글·문서·테스트·설치와 업데이트 단계·어댑터를 걷어내요. 이 SPEC 은 조각 0 이에요. 코드는 한 줄도 바꾸지 않고, 걷어낼 곳 전부의 목록(파일:줄, 영역별), 걷어내는 순서, 사용자에게 보일 한 줄, 되살릴 때 쓸 복원 기록을 정해요. 제거 조각 1~3 은 이 SPEC 이 정한 파일 소유와 순서를 따라 따로 SPEC 을 가져요.

기준은 `v2026.11.0` 의 `53cce6c580d8a1b59aee04376ad7459f81b096d1`(조각 0 을 시작한 로컬 통합 가지 tip)이에요. 같은 낱말로 센 숫자(`inventory.py` 가 그 커밋의 파일 내용만 읽어 셈해요): 파일 589개, 줄 6503줄. 이 가운데 `codex`(대소문자 무관)를 담은 파일이 576개이고 `OMP` 낱말만 담은 파일이 13개예요(카드가 센 숫자와 같아요: src 86·tests 198·skills/agents/docs/install/templates 139 는 모두 `codex` 파일 수이고, 아래 표의 src 89·tests 204 는 거기에 OMP 만 담은 3·6개를 더한 값이에요).

## Non-goals

- 이 조각에서 코드·테스트·문서·스크립트를 지우거나 고치는 일: 하지 않아요. 목록과 순서만 정해요.
- 사용자의 `~/.codex`, `~/.agents`, 프로젝트의 `.codex/` 파일을 moai 가 지우는 일: 어느 조각도 하지 않아요(D-3).
- 카드와 SPEC 의 글을 지우는 일: 하지 않아요. 언젠가 되살릴 일이라 철회 상태로 남겨요(D-4).
- GLM 지원을 줄이는 일: 하지 않아요. GLM 은 Claude Code 실행기의 모델 선택이라 그대로예요.
- Claude Code 쪽 동작을 바꾸는 일: 하지 않아요. 모든 조각이 Claude·GLM 시험을 그대로 통과해야 해요.

## Requirements

- REQ-1: The inventory shall list every tracked file of the base commit that holds `codex` in any case or the word `omp`, one row per file with its area, owning slice and mentioning line numbers. · failure: none
- REQ-2: When `inventory.py` runs on the base commit with `--check 589`, it shall exit 0 and print the file and line total for every area and slice. · failure: none
- REQ-3: Every inventory row shall name exactly one owning slice, which is `1`, `2`, `3`, `4`, `1>2` or `keep`. · failure: none
- REQ-4: When two slices need the same file, the slice order shall name the slice that edits it first. · failure: none
- REQ-5: When a removal slice is finished, it shall pass the full unnarrowed check with every Claude Code and GLM test green. · failure: none
- REQ-6: The removal shall delete no file in the user's Codex home folders, nor in a project's `.codex` folder. · failure: none
- REQ-7: When `moai update` previews a project or home that still holds Codex files, it shall print one line saying moai no longer manages them. · failure: none
- REQ-8: The SPEC folders and the cards that describe Codex or OMP work shall stay in the repository, withdrawn but not deleted. · failure: none
- REQ-9: The restore record shall name the SPEC folders, the cards and the commits that hold the Codex work and the steps that bring it back. · failure: none
- REQ-10: When every slice has landed, a repo scan shall fail on `codex` or `omp` in a shipped path outside the kept folders and the allowed lines. · failure: none
- REQ-11: IF any lane session still runs on Codex, THEN the slice that removes the Codex session-id and lane-loop code shall not start. · failure: none
- REQ-12: Each slice shall land as one merge into the local integration branch, so reverting that merge brings its removal back. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): 목록의 파일 수가 기준 커밋의 `git grep` 과 맞아요 · verify: `python3 -I .moai/specs/SPEC-20261121-remove-codex-support/inventory.py 53cce6c580 --out /dev/null --check 589`
- AC-2 (REQ-1): `codex` 를 담은 파일은 576개예요(여기에 `OMP` 낱말만 담은 13개를 더해 AC-1 의 589) · verify: `git grep -Iil codex 53cce6c580 | wc -l` 이 576 을 찍어요
- AC-3 (REQ-3): 모든 행의 소유 조각이 허용 값이에요 · verify: `awk -F'\t' 'substr($1,1,1)!="#" && $2 !~ "^(1|2|3|4|1>2|keep)$"' .moai/specs/SPEC-20261121-remove-codex-support/inventory.tsv | wc -l` 이 0 을 찍어요
- AC-4 (REQ-4): 두 조각이 쓰는 파일은 `1>2` 두 개(`src/main.rs`, `src/help.rs`)뿐이에요 · verify: `awk -F'\t' 'substr($1,1,1)!="#" && $2=="1>2" {print $3}' .moai/specs/SPEC-20261121-remove-codex-support/inventory.tsv` 가 그 두 경로를 찍어요
- AC-5 (REQ-8, REQ-9): 복원 기록에 Codex·OMP 를 담은 SPEC 폴더 57개가 있어요 · verify: `wc -l < .moai/specs/SPEC-20261121-remove-codex-support/restore-record.txt` 가 57 을 찍어요
- AC-6 (REQ-7): 사용자에게 보일 한 줄이 SPEC 에 있어요 · verify: `grep -c "moai no longer manages Codex files" .moai/specs/SPEC-20261121-remove-codex-support/spec.md` 가 1 이상이에요
- AC-7 (REQ-10, REQ-11, REQ-12): 조각 순서 표에 조각 1·2·3·4·5 와 시작 조건이 있어요 · verify: `grep -c "^| 조각 [1-5] " .moai/specs/SPEC-20261121-remove-codex-support/spec.md` 가 5 를 찍어요

## Scope

- write: 이 SPEC 폴더(`spec.md`, `inventory.py`, `inventory.tsv`, `restore-record.txt`)
- read-only: 저장소 전체(목록을 만들기 위한 `git grep` 만)
- forbidden: 코드·테스트·문서·스킬·스크립트·설치 파일 수정, 카드 상태 변경(카드 철회는 리더가 해요, D-4)

### Inventory

`inventory.tsv` 는 한 파일이 한 줄이에요: 영역, 소유 조각, 경로, 줄 수, 줄 번호 목록(`경로:줄` 이 곧 한 번의 언급). `inventory.py` 는 같은 규칙으로 다시 만들 수 있고(`python3 -I inventory.py <커밋> --out inventory.tsv --check 589`), 그 커밋의 파일 내용만 읽어요.

#### 영역별 (기준 커밋 53cce6c580)

| 영역 | 파일 | 줄 | 무엇이 들어 있나 |
| --- | --- | --- | --- |
| launcher-and-session | 5 | 117 | `moai codex`/`-l`·`-a` 실행 경로(`src/launcher_init.rs`·`launcher_terminal.rs`), 세션 인자 만들기(`src/session.rs`), 현재 세션 찾기 |
| home-layer | 20 | 124 | 홈 업데이트의 Codex 단계: `~/.codex` 경로(`src/host_home.rs`), 라우팅 블록(`home_routing.rs`), 스타일(`home_style.rs`), `update*.rs`, 구 스킬(`skills_legacy.rs`) |
| hooks | 6 | 210 | Codex 훅 어댑터와 훅 파일 쓰기(`src/hooks.rs`, `hook_engine.rs`, `hook_guards.rs`, `hook_legacy.rs`, `hook_config.rs`, `update_hooks.rs`) |
| project-layer | 14 | 314 | 프로젝트 단계: `.codex/` 분류(`go_state.rs`), 호스트 단계(`project_hosts.rs`), 생성(`generation.rs`, `agent_defs.rs`, `agent_install.rs`), 프로필(`profile.rs`), 모델·노력 질문 |
| statusline | 2 | 2 | 상태줄 한두 줄 |
| codex-only-files | 28 | 775 | 이름부터 Codex 인 파일: `src/hosts/codex.rs`, `src/hosts/codex_config.rs`, `src/ops/codex_*.rs` 5개, `.codex/agents/*`, `tests/codex_*`, `tests/fixtures/hooks-pre-p4/codex*`, `docs/**/codex*` |
| other-src | 4 | 24 | 위에 안 든 소스(`src/omp_agents.rs` 등) |
| lane-and-card-tooling | 29 | 164 | 개발 도구: 레인 루프의 Codex Stop 훅·`/goal` 한도(`src/lane_loop.rs`), 카드 세션 ID(`CODEX_THREAD_ID`, `src/card/dispatch.rs`), `lane_*`, `leader_*`, 결정·트레일러·감사·게이트 |
| cli-surface | 2 | 135 | 명령 선언과 도움말(`src/main.rs`, `src/help.rs`): 호스트 값 `cc\|glm\|codex\|omp`, `codex-*` 명령, 훅 계획 `--host` |
| tests-and-fixtures | 191 | 2060 | 시험과 고정 파일(`tests/**`) |
| skills-and-agent-text | 44 | 131 | `skills/moai/**`(SKILL.md·workflows·references), `templates/**`(번들 팩, GitHub 템플릿) |
| docs | 92 | 924 | `docs/**`, `README*`, `AGENTS.md`, `CONTRIBUTING.md`, `SECURITY.md`, `SUPPORT.md` |
| install-ci-tools | 21 | 65 | `install/*`, `.github/*`, `scripts/*`, `tools/*`, `schemas/*`, `Cargo.toml`, `experiments/*` |
| project-state-files | 13 | 611 | 생성되는 매니페스트(`.moai-cli/manifests/**`)와 이 저장소의 프로필(`.moai/project.yaml`) |
| keep-restore-record | 118 | 847 | 남기는 것: `.moai/specs/**`(SPEC 57개 폴더의 글), `.moai-cli/reports/**`(지난 이슈 본문·보고서) |
| 합계 | 589 | 6503 | |

#### 소유 조각별

| 조각 | 파일 | 줄 | 그 가운데 src | tests | 담당 영역 |
| --- | --- | --- | --- | --- | --- |
| 1 | 91 | 993 | 25 | 66 | launcher-and-session, home-layer 와 그 시험 |
| 1>2 | 2 | 135 | 2 | 0 | `src/main.rs`, `src/help.rs`(조각 1 이 먼저, 끝나면 조각 2) |
| 2 | 92 | 1809 | 33 | 56 | hooks, project-layer, statusline, codex-only 파일(src·tests·`.codex/`), other-src, 프로필 스키마, 이 저장소 프로필 |
| 3 | 238 | 2319 | 0 | 63 | 스킬·템플릿 글, 문서, 설치·CI·도구, 생성 매니페스트, 남는 시험 |
| 4 | 48 | 400 | 29 | 19 | lane-and-card-tooling 과 그 시험 |
| keep | 118 | 847 | 0 | 0 | 손대지 않아요 |

시험은 이름의 낱말로 소유를 처음 정했어요(`inventory.py` 의 `TEST_WORDS`). 시험이 부르는 코드를 지우는 조각이 그 시험도 같이 고쳐요. 이름만으로 틀리게 정해진 시험은 그 조각이 `cargo test --no-run` 에서 깨지는 것으로 알게 되고, 소유를 바꾸면 그 조각의 Result 에 적어요.

## Plan

| 조각 | 무엇 | 언제 시작해요 | 같이 돌 수 있는 조각 |
| --- | --- | --- | --- |
| 조각 1 | 실행기·모델 질문·홈 업데이트 | 지금 | 2, 3 |
| 조각 2 | 프로젝트 단계·훅 어댑터·`.codex/*` 생성·호스트 질문·프로필·Codex 전용 파일 | 지금(`src/main.rs`·`src/help.rs` 는 조각 1 이 반환한 뒤에 잡아요) | 1, 3 |
| 조각 3 | 스킬·문서·설치·CI·도구·템플릿·생성 매니페스트·남는 시험 | 지금(생성 매니페스트와 `source-manifest.json` 해시는 병합 직전에 다시 만들어요) | 1, 2 |
| 조각 4 | 레인·카드 도구의 Codex 경로(카드 세션 ID, 레인 루프 Stop 훅) | 조각 1·2·3 이 모두 착지한 뒤, 그리고 Codex 에서 도는 레인이 하나도 없다고 리더가 확인한 뒤(REQ-11, D-6) | 없음 |
| 조각 5 | 마지막 검사: 저장소 검사에 `codex`·`omp` 가 허용 목록 밖에 없다는 규칙을 더하고 남은 한 줄을 정리 | 조각 4 가 착지한 뒤 | 없음 |

각 조각은 `--no-ff` 병합 한 번으로 통합 가지에 들어가요(REQ-12). 조각 3 이 낱말 규칙(조각 5)을 먼저 켜면 아직 안 지운 조각이 모두 실패하므로, 규칙은 마지막이에요.

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

1. **SPEC 폴더 57개**: `restore-record.txt`(한 줄에 하나)에 있어요. 기준 커밋에서 `codex` 나 `OMP` 를 담은 `.moai/specs/*` 폴더 전체이고, 이름이 Codex·OMP 인 것은 `SPEC-20261118-codex-question-tool`, `SPEC-20261119-codex-conflict-choices`, `SPEC-20261118-omp-host`, `SPEC-20261119-omp-host`, `SPEC-20261118-omp-launcher`, `SPEC-20261118-profile-host-omp` 예요. 이 폴더들은 삭제하지 않고 `status` 도 손대지 않아요(SPEC 상태 변경은 이 일의 범위 밖이에요).
2. **카드**: 제목에 `codex`(대소문자 무관)나 낱말 `OMP` 가 든 카드는 2026-11-21 의 `moai card list`(전체 287장)에서 20장이에요: C-8(ready), C-22(draft), C-29(draft), C-38(draft), C-52(done), C-56(done), C-57(ready), C-79(done), C-89(done), C-240(ready), C-243(ready), C-253(review), C-259(ready), C-260(ready), C-264(ready), C-271(active), 그리고 이 일의 C-289~C-292. 제목으로는 Codex 만 다루는 카드가 C-52(done)·C-253(review)·C-271(active)이고, 나머지는 Claude Code 와 함께 Codex 를 말해요. 카드는 원장에 그대로 두고, 아직 시작하지 않은 카드를 철회할지는 카드마다 리더가 정해요(D-4). 철회는 `ready` 카드를 `hold` 로 옮길 뿐 지우지 않아요.
3. **코드**: 기준 커밋 `53cce6c580d8a1b59aee04376ad7459f81b096d1` 에 Codex 코드 전체가 있어요. 각 조각은 `--no-ff` 병합 한 번이라, 조각의 병합 커밋을 역순으로 `git revert -m 1 <병합 커밋>` 하면 그 조각이 돌아와요. 병합 커밋 sha 는 조각이 착지할 때 그 조각 SPEC 의 Result 에 적어요. 되살릴 때는 이 순서를 거꾸로(조각 5 → 1) 되돌려요.
4. **기준 가지(결정 요청)**: 통합 가지가 크게 앞서 가도 기준 커밋을 찾기 쉽도록, 리더가 그 커밋에 이름 있는 가지 한 개를 만들어 두면 좋아요(D-5). 태그가 아니라 가지예요.

## Risks

- 카드 세션 ID 는 `CODEX_THREAD_ID` 도 읽어요(`src/card/dispatch.rs` 의 holder 와 `src/session.rs`). 지금 일부 레인이 Codex 로 도는 중이면, 그 코드를 지우는 순간 그 레인은 카드를 잡거나 놓을 수 없어요. 그래서 조각 4 를 마지막으로 미뤘고 시작 조건에 "Codex 레인 없음"을 넣었어요(D-6).
- `src/main.rs`·`src/help.rs` 는 조각 1 과 2 가 모두 건드려요. 한꺼번에 고치면 충돌이라 소유를 `1>2` 로 나눴어요(조각 1 이 반환하면 조각 2 가 잡아요).
- 프로필에 이미 `codex`·`omp` 가 적힌 프로젝트: 읽을 때 무시하고 안내 한 줄만 보이게 해요(D-2). 그러면 프로필 스키마와 그 읽기 코드에 `codex` 낱말이 한 곳 남아서, 조각 5 의 허용 목록에 그 한 곳과 사용자에게 보이는 한 줄을 넣어요.
- 시험 소유를 이름의 낱말로 처음 정해서 틀리게 정해진 시험이 있을 수 있어요. 그 조각이 첫 빌드에서 알게 되고 Result 에 소유 변경을 적어요.
- 생성된 매니페스트와 도움말 고정 파일은 병합 순서에 따라 해시가 달라져요. 그래서 각 조각이 병합 직전에 다시 만들어요(조각 3 만의 일이 아니에요).
- `.agents/skills` 는 Codex 와 OMP 말고 다른 도구도 읽을 수 있는 폴더예요. 그래서 moai 가 그 폴더를 지우지 않고, 안내 한 줄만 보여요(REQ-6).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 목록은 `inventory.py` 가 기준 커밋의 파일 내용에서 만들고, 소유 조각을 영역 규칙과 시험 이름으로 정해요 | 손으로 세면 카드의 숫자(src 86·tests 198)와 어긋나기 쉬워서, 같은 낱말로 두 번 셀 수 있게 했어요. OMP 낱말 13개 파일은 카드 숫자에 들지 않아 따로 더했어요 | 규칙 바꾸기 |
| D-2 | 프로필에 적힌 `codex`·`omp` 는 읽고 무시하며 안내 한 줄을 보여요. 조각 2 가 구현해요 | 카드 본문(조각 2)이 정한 방식이에요 | 거절로 바꾸기 |
| D-3 | 사용자의 `~/.codex`, `~/.agents/skills`, 프로젝트 `.codex/` 는 moai 가 지우지 않아요. 관리만 멈추고 안내 한 줄만 보여요 | 사용자 결정(카드 본문) | 없음 |
| D-4 | 카드와 SPEC 은 철회해서 두고 지우지 않아요. 카드 철회는 리더가 해요 | 사용자 결정 "언젠가 codex 가 쓸만해지면 사용하도록" | 철회 풀기 |
| D-5 | 기준 커밋에 이름 있는 가지(태그 아님) 하나를 리더가 만들어 두기를 요청해요. 기본값은 만들지 않음 | 병합 커밋 되돌리기만으로도 복원은 되지만, 기준 커밋을 찾는 일이 쉬워져요. 가지를 밀어 올리는 일은 레인이 하지 않아요 | 만들지 않기 |
| D-6 | 조각 4(카드 세션 ID, 레인 루프 Codex 경로)는 Codex 에서 도는 레인이 없다고 리더가 확인한 뒤에 시작해요 | 레인 도구가 Codex 세션 ID 를 읽는 동안 지우면 그 레인이 카드를 못 다뤄요. 기본값은 대기, 되돌릴 수 있고 분류는 none 이에요 | 조건 없이 시작 |
| D-7 | 조각 4 와 조각 5 는 새 카드가 필요해요(리더가 만들어요). 지금 있는 카드는 조각 1(C-290)·2(C-291)·3(C-292)뿐이에요 | 목록에서 레인·카드 도구가 4 조각만큼(48 파일) 커서 3 조각에 못 넣어요 | 조각 3 에 합치기 |

## Result

조각 0 은 계획 SPEC 이라 코드는 바뀐 것이 없고, 아래는 기준 커밋 `53cce6c580d8a1b59aee04376ad7459f81b096d1` 에서 각 확인 명령을 실제로 돌린 결과예요(2026-11-21).

- AC-1 met: `python3 -I .moai/specs/SPEC-20261121-remove-codex-support/inventory.py 53cce6c580 --out /dev/null --check 589` → 종료 0, 영역 15개와 조각 6개의 파일·줄 합계가 인쇄되고 총 589 파일 6503 줄.
- AC-2 met: `git grep -Iil codex 53cce6c580 | wc -l` → 576. 같은 방법으로 센 `OMP` 낱말만 담은 파일은 13개(`comm` 으로 576개 목록과 겹치지 않는 것만 센 값)이고 576+13=589 예요.
- AC-3 met: `awk` 로 허용 값 밖의 소유 조각 행을 센 값이 0.
- AC-4 met: `1>2` 행은 `src/help.rs`, `src/main.rs` 두 개.
- AC-5 met: `restore-record.txt` 57줄.
- AC-6 met: `grep -c "moai no longer manages Codex files" spec.md` → 2(영어 한 줄과 SPEC 안에서 그 줄을 말하는 문장).
- AC-7 met: 표의 `| 조각 1 ` … `| 조각 5 ` 행 5개.
- 보조: 파일 소유별 src 25+2+33+29 = 89(= `codex` 86 + OMP 만 3), tests 66+56+63+19 = 204(= 198 + OMP 만 6).

돌리지 않은 것: 조각 1~5 의 코드 제거(각 조각의 SPEC 이 맡아요), 카드 철회, 기준 가지 만들기(리더의 몫, D-4·D-5), 영역·조각 규칙이 시험 이름의 낱말에 기대는 부분이 실제 빌드에서 맞는지(각 조각이 첫 빌드에서 확인해요).
