# t452 M0 — 사전 점검 실측 (plan.md §C)

## C.0 BASELINE_SHA (동결)

run-phase 첫 커밋 이전에 동결했다. AC-CSL-005 · 011 · 012 의 변경-집합 판정은 전부 이 값을 기준으로 한다.

```
$ git rev-parse HEAD
c529b2e4aaf5148aee7e6c67649bf392837bbb06
```

## C.1 codex 버전

```
$ codex --version
codex-cli 0.152.1
exit=0
```

## C.3 사용자 계층 사전 대조군 (AC-CSL-010 pre)

측정 대상은 개발자의 실 사용자 계층 `$HOME/.codex` 이며, 프로브가 쓰는 격리 `CODEX_HOME` 과
다른 경로임은 프로브 명령 문면(`CODEX_HOME=/tmp/t452-m0/codexhome`)에서 확인된다.

```
$ shasum -a 256 $HOME/.codex/config.toml
8adec56f13e1cb6baafdb89a406f907f98b002d9136ca1cbf6f40a23194b4ae7

$ find $HOME/.codex/skills -maxdepth 1 -mindepth 1 | sort
/Users/goos/.codex/skills/.system
/Users/goos/.codex/skills/hatch-pet
```

`ls -1` 이 아니라 `find` 로 뜬 이유: 이 셸의 `ls` 는 alias 라 `.` / `..` 를 함께 출력한다.
사후 대조에서 같은 방법을 쓰지 않으면 두 측정이 비교 불가능해지므로, 사후 측정도 같은 `find` 로 뜬다.

## C.4 `codex doctor` 를 측정 표면으로 쓰지 않음

plan.md §C.4 · spec.md §A.5 P4 에 따라 시도하지 않는다 — 그 출력은 스킬 뿌리를 하나도 열거하지 않으며,
그 침묵을 "뿌리가 없다"로 읽는 것이 오독이다.
