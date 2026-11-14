# `hygiene-plan`

operation `hygiene.plan` · 코드 `src/ops/hygiene_plan.rs`, `src/hygiene.rs` · Issue #2934 (S4a), #2930 R-4

`hygiene-plan PROJECT --integration-ref REF [--tmp DIR] [--now UNIX]` 는 레인 작업이 남긴 것 가운데 보관 폴더 `<HOME>/.moai/holding/<UTC 날짜>/` 로 옮겨도 되는 것을 낸다. 쓰지 않는다. 옮기기(`hygiene-apply`)와 보관 3일 뒤 지우기(`hygiene-purge`)는 따로다.

- `worktree`: 판정은 새로 하지 않고 `worktree-gc-plan`(`src/worktree_gc.rs`)이 `remove` 로 낸 연결 worktree 만 `hold` — 통합 tip 에 병합됨, 커밋 안 된 파일·안전하지 않은 무시 파일 없음, 그 안에서 일하는 프로세스 없음, 최근 변경 없음. 다른 항목은 gc 의 첫 이유로 `keep`. 경로는 절대 경로.
- `tmp_build`: `--tmp`(기본 시스템 임시 폴더) **바로 아래** 폴더(링크 아님) 가운데 꼭대기에 일반 파일 `CACHEDIR.TAG`(cargo 가 모든 target 꼭대기에 씀)가 있는 것. 폴더와 그 바로 아래 항목의 가장 새 변경이 2일(`TMP_MIN_AGE_SECS`, 리더 A 정리 기준) 넘게 지났고, 어떤 프로세스도 그 안을 작업 폴더로 두지 않고 명령 줄에(Linux 는 `/proc/<pid>/environ` 의 환경, 예: `CARGO_TARGET_DIR` 에도) 그 경로를 담지 않을 때만 `hold`. macOS 는 다른 프로세스의 환경을 보여 주지 않아요(2026-11-14 Darwin 27 실측: `CARGO_TARGET_DIR=…` 로 띄운 프로세스가 `ps eww` 에 `/bin/sleep 31` 로만 나옴) — 그래서 macOS 에서 환경으로만 이 폴더를 가리키는 경우는 2일 나이 기준(빌드가 쓰는 동안은 늘 최근)이 막고, 옮기기(`hygiene-apply`)가 직전에 열린 파일을 다시 확인한다. 이유: `recent`, `in_use`, 읽지 못하면 `in_use_unknown`·`age_unknown` 으로 `keep`.
- 결과: `holding`(오늘 보관 폴더), `tmp_min_age_secs`, `holding_keep_secs`(3일), `totals`{`hold`, `keep`, `hold_bytes`}, `items`[{`kind`, `path`, `action`, `reason`, …}], `plan_digest`(옮길 항목의 kind·경로·head), `writes: false`. 크기 재기는 모두 합쳐 20초 안, 넘으면 `bytes_partial`.
