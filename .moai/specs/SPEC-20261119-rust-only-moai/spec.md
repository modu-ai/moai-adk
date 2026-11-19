---
id: SPEC-20261119-rust-only-moai
title: "moai runs Rust only: inventory of every Go-delegated word and execution path"
status: draft
revision: 1
created: 2026-11-19
---

# moai runs Rust only: inventory of every Go-delegated word and execution path

## Intent

사용자 결정(2026-11-19): `moai` 는 Rust 만 실행하고, Go 빌드(`moai.old`)로 보내는 일을 전부 없애요. 지금 이름 바꾸기 shim(v2, `install/moai-switch.sh`)은 59개 낱말과 `spec`·`session`·`bundle` 의 일부 하위 동작을 Go 로 보내고, 호스트 훅·상태줄·MCP 서버 같은 실행 경로도 Go 를 부를 수 있어요. 무엇을 옮기고 무엇을 접을지 정하기 전에, 모든 낱말과 경로를 근거와 함께 한 곳에 적는 것이 이 첫 조각이에요. 코드는 바꾸지 않아요.

이 SPEC 의 전체 계획(조각 2.. 이식, 마지막 조각 shim → Rust 링크 switch)은 같은 폴더의 `inventory.md` 가 끝난 뒤 정해져요.

## Non-goals

- 이식·퇴역·shim 교체: 이 조각은 읽기만 해요(뒤 조각의 몫).
- Go 도구체인, `~/go/bin` 의 파일, 다른 Go 프로젝트, 열린 세션, 작업 트리, 설정, 원장은 건드리지 않아요.
- 실제 HOME 에 쓰기: 읽기만 해요.

## Requirements

- REQ-1: When the inventory is read, it shall list every word the v2 shim routes to the Go build, one row per word, and one row per routed `spec`, `session` and `bundle` action. · failure: none
- REQ-2: When a row is read, it shall name one disposition, `rust-has`, `port` or `retire`, and the evidence that makes the disposition checkable. · failure: none
- REQ-3: When a row is `port`, it shall give a size from the counted lines of the Go source and name who calls the word. · failure: none
- REQ-4: When the inventory is read, it shall list every launcher, install and update step, host profile, host hook, status line and MCP registration that can run the Go build, with its disposition. · failure: none
- REQ-5: While the inventory is being made, the work shall change no file outside this SPEC folder and write nothing in the real HOME. · failure: none

## Acceptance

- AC-1 (REQ-1): every word on the shim's Go line has a row in `inventory.md`.
  verify: `for w in $(sed -n '/^  af|agent/p' install/moai-switch.sh | tr -d ' )' | tr '|' ' '); do grep -q "^| $w " .moai/specs/SPEC-20261119-rust-only-moai/inventory.md || { echo "missing $w"; exit 1; }; done`
- AC-2 (REQ-2, REQ-3): every row of the word table ends in one of the three dispositions.
  verify: `n=$(awk '/^## Words/{f=1;next} /^## Execution paths/{f=0} f && /^\| [a-z]/ && !/^\| (word|---)/' .moai/specs/SPEC-20261119-rust-only-moai/inventory.md | grep -c -v -E ' \| (rust-has|port|retire)[^|]* \| [^|]* \| '); echo "rows without a disposition: $n"; test "$n" = 0`
- AC-3 (REQ-4): the path table exists and names the status line, the hooks, the MCP registration, the launcher and the installer.
  verify: `awk '/^## Execution paths/{f=1;next} f && /^\| (statusline|hooks|mcp|launcher|installer)/' .moai/specs/SPEC-20261119-rust-only-moai/inventory.md | wc -l`
- AC-4 (REQ-5): the change touches only this SPEC folder.
  verify: `git diff --name-only $(git merge-base HEAD origin/v2026.11.0) HEAD | grep -v '^.moai/specs/SPEC-20261119-rust-only-moai/' | wc -l`

## Scope

- write: `.moai/specs/SPEC-20261119-rust-only-moai/**`
- read-only: `install/`, `src/`, `skills/`, `tools/`, `tests/`, `docs/`, the Rust help list, the Go help texts, the Go source and the real host settings of this machine
- forbidden: any file outside the SPEC folder, any write in the real HOME, any run of a Go subcommand without `--help`

## Plan

1. Collect the Go help of every routed word and the Rust command list (empty temporary HOME).
2. For each word: the Rust equivalent, the disposition, the Go size and the callers.
3. For each execution path: what runs, how it is wired, what replaces it.
4. Write both tables into `inventory.md`; the leader chooses the slices after it.

## Risks

- A word that looks unused in the repository can be called by a project's own hook or `.mcp.json`: the tables name those callers from the machine's files, and a caller the scan could not see (another machine) is listed as not checked.
- The Go help of a word can differ from what a project calls: the table quotes help only, never runs a subcommand.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Disposition judged by user need (a caller exists) first, size second | Leader dispatch: each word is Rust-has, port or retire, with evidence | Change the row |

## Result

조각 1(읽기만)을 마쳤어요. 목록은 `inventory.md` 에 있어요.

- 낱말 59행(shim Go 줄 60칸에서 `bundle` 은 하위 동작 표로): rust-has 16, port 2(`gate`, `mcp-server`), retire 41. 하위 동작 17행(`spec` 7, `session` 7, `bundle` 3): rust-has 7, retire 10. 실행 경로 12행.
- 실제 호출자가 있는 것: `gate`(git pre-commit 훅이 13개 저장소에서 `moai gate` 를 불러요; Rust 의 `gate` 는 종료 코드 2라서 Go 가 빠지면 그 저장소의 모든 커밋이 막혀요), `mcp-server`(10개 프로젝트의 `.mcp.json`), `verify`(moai-cowork 의 sync-audit 한 곳).
- 판단이 필요한 것은 `inventory.md` 의 "Decisions" 여섯 줄이에요.
- 확인: AC-1 누락 0, AC-2 판정 없는 행 0, AC-3 경로 행 7, AC-4 SPEC 폴더 밖 파일 0(각 명령은 이 가지의 마지막 커밋에서 실행). 음성 시험: `gate` 행을 빼면 AC-1 이 실패하고, 판정을 `maybe` 로 바꾸면 AC-2 가 1을 세요.
- 돌리지 않은 것: Go 하위 명령(모두 `--help` 만), 다른 컴퓨터, 셸 기록.
