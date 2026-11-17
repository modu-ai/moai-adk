---
id: SPEC-20261118-installer-stdin-null
title: "The installer's moai-cli children read nothing from standard input"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261117-installer-tty
---

# The installer's moai-cli children read nothing from standard input

## Intent

installer-tty 감사의 N1이에요. `curl … | sh`에서 `install.sh`의 stdin은 script 자체이고, `sh`는 그것을 조금씩
읽어요. 그래서 stdin을 물려받은 자식이 그것을 읽으면 아직 실행되지 않은 script 뒷부분이 사라져요. 지금 `moai-cli`는
stdin을 읽지 않고 `main "$@"`가 마지막 줄이라서 문제가 드러나지 않았을 뿐이에요. 이 변경으로 `install.sh`의 모든
`moai-cli` 자식 호출(`install-inspect`, `language-set`, `update` 미리 보기와 적용, `--version`)이 stdin을
`/dev/null`에서 받아요. 답을 `/dev/tty`에서 읽는 경로(`ask_line`)는 그대로예요.

`install.ps1`은 바꾸지 않아요. 문서의 `irm … | iex`와 `& ([scriptblock]::Create((irm …)))`는 script를 메모리의
문자열로 받아 실행하고, `pwsh -File`은 파일에서 읽어요. 어느 경우에도 script가 자식이 물려받는 stdin으로 흘러오지 않아요.
자식이 stdin을 읽어도 콘솔 입력이지 script가 아니에요.

## Requirements

- REQ-1: When `install.sh` runs a `moai-cli` child command, the installer shall give that child `/dev/null` as standard input. · failure: none — `/dev/null`은 항상 열려요. 자식이 실패하면 각 호출의 기존 실패 경로를 그대로 따라요.
- REQ-2: The answer path of `install.sh` shall keep reading answers from standard input or `/dev/tty` as SPEC-20261117-installer-tty states.

## Acceptance

- AC-1 (REQ-1): verify: `python3 -m unittest -k children_read_nothing tests/install_script_test.py` — script를 stdin에 넣고 `sh -s`로 실행해요. 가짜 `moai-cli`는 호출마다 stdin에서 읽은 바이트 수를 적고, 모든 호출이 0바이트예요.
- AC-2 (REQ-2): verify: `python3 -m unittest tests.install_script_test.InstallShTest` — `curl | sh` 터미널 시험(`/dev/tty`)을 포함해 통과해요.
- AC-3 (REQ-1, REQ-2): verify: `sh -n install/install.sh`, `bash -n install/install.sh` → exit 0.

## Scope

- write: `install/install.sh`, `tests/install_script_test.py`, this SPEC.
- `install/install.ps1`: unchanged (see Intent).

## Plan

1. Test first: a fake `moai-cli` that reads its standard input and records the byte count, run through `sh -s` with the script on stdin.
2. Add `</dev/null` to each `moai-cli` child call in `install.sh`, with a comment.
