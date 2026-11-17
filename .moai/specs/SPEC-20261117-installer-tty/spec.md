---
id: SPEC-20261117-installer-tty
title: curl | sh로 실행해도 터미널에서 묻기 (install.sh)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-installer-skill-agents
---

## Intent

C-78 감사(PASS, `6011fe9c5d`)의 지적 두 개예요.

- N1: 문서의 `curl -fsSL https://adk.mo.ai.kr/install.sh | sh`는 script를 stdin으로 받아요. 그래서 `[ -t 0 ]`이 거짓이고,
  사람이 터미널 앞에 있어도 언어 메뉴와 "Install them now?"를 묻지 않아요. 결국 binary만 설치돼요. rustup처럼 답을
  `/dev/tty`에서 읽어요(열 수 있을 때만, 그 읽기에만).
- N2: `moai-cli update` 호출의 stderr가 파일로 가는 것을 시험으로 고정해요. 터미널로 가면 제품이 거기 낸 말이나 질문이
  사람 화면에서 기다릴 수 있어요.

PowerShell 설치 스크립트는 그대로예요(콘솔에서 읽어서 이미 물어요).

## Requirements

"answer source"는 질문의 답을 읽는 곳이에요: stdin이 터미널이면 stdin, 아니면 열 수 있는 `/dev/tty`, 둘 다 아니면 없음.

- REQ-1: When stdout is a terminal and an answer source exists, `install.sh` shall ask the language menu and the install question once each and read the answers from that source.
- REQ-2 (REQ-1): IF no answer source exists or stdout is not a terminal, THEN `install.sh` shall ask nothing and write nothing into HOME's host folders without `--yes`.
- REQ-3: When `install.sh` runs `moai-cli update`, the installer shall send that command's stderr to a file. · failure: none — 파일을 못 만들면 호출이 실패하고 C-78의 실패 경로(경고, exit 0)로 가요.

실패 경로: REQ-1 → REQ-2.

## Acceptance

- AC-1 (REQ-1): verify: `python3 -m unittest -k curl_pipe_sh_at_a_terminal tests/install_script_test.py` — stdin에 script, 출력과
  controlling terminal은 pty(`setsid` + `TIOCSCTTY`)예요. 메뉴와 질문이 한 번씩 나오고, 친 답(`3`, `y`)이 쓰여요.
- AC-2 (REQ-2): verify: `python3 -m unittest -k without_any_terminal -k no_menu_when -k without_a_terminal tests/install_script_test.py`
- AC-3 (REQ-3): verify: `python3 -m unittest -k never_write_to tests/install_script_test.py` — 미리 보기와 적용 두 호출 모두
  `stderr=redirected`.
- AC-4 (REQ-1…REQ-3): verify: `sh -n install/install.sh`, `bash -n install/install.sh` → exit 0.

## Scope

- `install/install.sh`, `tests/install_script_test.py`, 이 SPEC.

## Plan

1. 시험 먼저: controlling terminal이 있는 `curl | sh` 모양의 실행, 없는 실행, update 호출의 stderr.
2. `install.sh`에 `answer_source`와 `ask_line`. 언어 메뉴와 설치 질문이 같은 규칙을 써요.
