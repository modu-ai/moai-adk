---
id: SPEC-20261121-go-path-removal
title: "moai drops the Go path: inventory refresh, Go-link removal in update, name-switch retirement"
status: draft
revision: 1
created: 2026-11-21
---

# moai drops the Go path: inventory refresh, Go-link removal in update, name-switch retirement

## Intent

사용자 지시(2026-11-21, 리더 전달): "go하고도 연결 모두 끊어라", "go path 완전 제거해라", "모든 moai-adk-go 버전을 사용중단하고 moai update 하면 기존의 go와 연결 고리 제거후 업데이트 하고 moai-adk-go 관련 내용은 제거한다". 제품이 더 이상 어떤 경로로도 Go 빌드를 실행하거나 설치하지 않게 하는 카드다.

기계 상태(2026-11-21 기준): `moai` 는 `~/.local/bin/moai` → `~/.moai/bin/moai` → `~/.moai/versions/4.2647.15/bin/moai` 로 바로 Rust 빌드로 연결된다. 이름 전환 shim(`~/.moai-cli/path/moai`, v2)과 보관 사본 네 개는 PATH 에서 내려온 고아 파일이 됐고, `~/go/bin` 에는 Go moai 바이너리가 없다(사용자가 휴지통으로 치움, 다른 Go 도구는 남음).

전체 계획은 세 조각이다: (1) 읽기 전용 목록 조사 — 이 revision 이 넘기는 것. (2) `moai update` 가 모든 사용자의 Go 연결고리를 제거한다(옛 런처 스크립트와 보관 사본, 그 이름을 부르는 PATH 줄, 프로젝트·홈 설정의 Go 세대 훅 줄 — 미리보기·적용·되돌리기 기록과 함께, 사용자 파일을 조용히 지우지 않고, Go 바이너리 자체는 사용자가 지우도록 목록으로 보여 준다). (3) 제품에서 이름 전환 스크립트와 그 update 단계와 Go 라우팅 낱말 목록을 제거하고 문서·스킬의 Go 시절 글을 정리한다. go-state·migrate 도구는 update 이식이 필요로 하는 동안만 남는다. Go 가드를 내리기 전에 보안 가드 이식 카드가 먼저 착지해야 하고, SPEC 은 영향을 받는 가드를 적어야 한다. 조각 2·3 읊의 세부 AC 는 각 조각 착수 때 이 SPEC 의 다음 revision 에서 갱신한다. HOME 경로 카드(폴더 이동과 설정 줄을 소유)와 겹치는 항목은 그 카드와 조율한다.

기준 목록: SPEC-20261119-rust-only-moai 의 `inventory.md`(가지 spec/20261119-rust-only-moai, head 903061f05233a42fa8c014d3e0819b70871fe19b). 이 revision 의 목록은 그 기준을 2026-11-21 기계 상태와 현 tip 에 맞춰 갱신한다.

조각 순서 결정(2026-11-21, 카드 결정 노트 3건 — 사용자 "진행해라" 승인): 목록 조각 바로 다음에, `moai update` 가 사용자 저장소의 Go 세대 git 훅을 찾아(곁의 해시 파일과 머리말 글로 인식) 미리보기에 보여 주고, 적용 때 되돌리기 기록과 함께 은퇴시키고, 그 자리에 Rust 훅을 설치한다. MoAI 표시가 없는 사용자 작성 훅은 건드리지 않고 "kept" 으로 나열한다. 이 슬라이스가 사용자가 처음 만나는 부분이다(Go 바이너리가 사라진 뒤 실패하는 커밋). 제품 작업: 어떤 프로젝트에도 이 git 훅을 설치·미리보기·제거하는 moai 명령 하나와, 그 명령을 은퇴 직후에 쓰는 update 단계. 트리 안의 `install/commit-msg` 는 옛 런처 경로를 고정하고 있어 plain `moai` 를 부르게 고친다. 모든 훅은 100ms 안쪽이고, 시험은 Rust 로 임시 저장소에서: Go 훅 은퇴 후 교체, 사용자 훅 바이트 그대로 보존, 섞인 폴더, 두 번째 실행은 변화 없음, 되돌리기 복원. pre-push 가 Rust 검사를 돌릴지(예: 밀린 범위의 trailer-check)는 그 조각의 revision 이 이 SPEC 에서 정한다. 리더가 이 기계의 공유 훅 폴더에서 Go 세대 훅을 이미 은퇴시키고 작은 Rust 훅 둘(commit-msg, pre-commit)을 넣었다 — 목록의 기계 상태에 기록했다.

## Non-goals

- 이식·퇴역·shim 제거의 실행: 이 revision 은 읽기만 한다(조각 2·3 의 몫).
- Go 도구체인, `~/go/bin` 의 다른 도구(buf, actionlint 등), 다른 Go 프로젝트, 열린 세션, 작업 트리는 건드리지 않는다.
- 실제 HOME 에 쓰기: 읽기만 한다.
- Go 하위 명령 실행: `--help` 없이는 실행하지 않는다(이 기계엔 Go moai 가 없어 실행 자체가 불가능하다).

## Risks

- 기계 상태는 하루 안에도 움직였다(11-19 목록 이후 shim 고아화, Go 바이너리 퇴장, 자산 번들 착지, 공유 훅 교체). 목록의 수는 읽은 시점의 것이고, 조각 2·3 실행 전에 관련 행을 다시 세야 한다.
- 판정을 갈린 호출자 수에 의존한다. 게이트 호출 12곳처럼 실시간으로 막힌 경로는 목록이 아니라 수리(슬라이스)가 닫는다 — 목록만으로 커밋 막힘이 풀리지 않는다.
- 폴더·파일 제거는 사용자 파일을 다룬다(shim 보관 사본, Go 세대 훅). 조용한 삭제 없이 미리보기·토큰·되돌리기 기록이 없는 조각은 착지하면 안 된다.

## Requirements

- REQ-1: When the inventory is read, it shall list every word on the Go line of the name-switch shim in install/moai-switch.sh and one row per routed spec session and bundle action. · failure: none
- REQ-2: When a row is read, it shall name one disposition rust-has port or retire and carry evidence that makes the disposition checkable today. · failure: none
- REQ-3: When the inventory is read, it shall reflect the machine state of 2026-11-21 and mark what changed since the 2026-11-19 baseline. · failure: none
- REQ-4: When the inventory is read, it shall list every place the product can still execute or install a Go path (installer, update switch step, launcher, hooks, statusline, commit-msg, templates/bundle-packs, tests reading the real HOME) with its disposition. · failure: none
- REQ-5: While the inventory is being made, the work shall change no file outside this SPEC folder and write nothing in the real HOME. · failure: none
- REQ-6: When the git-hook slice runs, moai update shall find Go-generation git hooks by their hash files and header text in the user's repository and show them in the preview. · failure: none
- REQ-7: When the git-hook slice applies, it shall retire the recognised Go hooks with a revert record and install the Rust hooks in their place while listing every hook without MoAI marks as kept. · failure: none

## Acceptance

- AC-1 (REQ-1): every word on the shim's Go line has a row in `inventory.md`.
  verify: `for w in $(awk '/^  af\|agent/{sub(/\)$/,"");sub(/^  /,"");n=split($0,a,"|");for(i=1;i<=n;i++)print a[i]}' install/moai-switch.sh); do grep -q "^| $w " .moai/specs/SPEC-20261121-go-path-removal/inventory.md || { echo "missing $w"; exit 1; }; done`
- AC-2 (REQ-2): every row of the word and action tables names one of the three dispositions.
  verify: `awk '/^## Words/{f=1;next} /^## Execution paths/{f=0} f&&/^\| [a-z]/&&!/^\| (word|action|---)/{n++; if ($0 ~ /\| (rust-has|port|retire) /) d++} END{print "rows="n" with-disposition="d; exit (n==d && n>0 ? 0 : 1)}' .moai/specs/SPEC-20261121-go-path-removal/inventory.md`
- AC-3 (REQ-3): the inventory has a machine-state section naming the resolution chain, the off-PATH shim with its saved copies, and the emptied Go bin folder of the HOME.
  verify: `grep -c -e 'versions/4.2647.15' -e 'moai.before-' -e 'no Go moai binary' .moai/specs/SPEC-20261121-go-path-removal/inventory.md`
- AC-4 (REQ-4): the execution-path table has rows for installer, update, launcher, hooks, statusline, commit-msg, bundle-packs and the tests.
  verify: `for p in installer update launcher hooks statusline commit-msg bundle-packs tests; do awk '/^## Execution paths/{f=1;next} f&&/^\| '"$p"'/{found=1} END{exit !found}' .moai/specs/SPEC-20261121-go-path-removal/inventory.md || { echo "missing $p"; exit 1; }; done`
- AC-5 (REQ-5): the change touches only this SPEC folder.
  verify: `git diff --name-only $(git merge-base HEAD v2026.11.0) HEAD | grep -v '^\.moai/specs/SPEC-20261121-go-path-removal/' | wc -l`

## Scope

- write: `.moai/specs/SPEC-20261121-go-path-removal/**`
- read-only: `install/`, `src/`, `skills/`, `tools/`, `tests/`, `docs/`, `templates/`, the installed Rust binary (`--help` and read-only probes only, temporary HOME), and this machine's state (`which`, `ls`, `shasum`, shell rc reads)
- forbidden: any file outside the SPEC folder, any write in the real HOME, any run of a Go subcommand

## Plan

1. Read the baseline inventory and the current tip sources: `install/moai-switch.sh`, `src/update/switch.rs`, `src/update.rs`, `src/update_tux.rs`, `src/shim_check.rs`, `tools/claude-statusline.sh`, `src/go_hooks.rs`, `src/user_assets.rs`, `src/ops/bundle.rs`, `install/commit-msg`, `tests/diff_go.rs`, the installer and launcher sources.
2. Probe the installed Rust binary with every routed word in a temporary HOME and record what it answers today.
3. Re-verify the machine-side counts the baseline recorded (project hook folders, git pre-commit callers, MCP registrations, shell rc lines) and record what changed.
4. Write `inventory.md` with the machine state, the word table, the action table, the execution-path table and the findings, and keep every claim on a command run for this file.
