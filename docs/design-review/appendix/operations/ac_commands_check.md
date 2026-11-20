# `ac-commands-check`

operation `ac_commands.check` · 코드 `src/ops/ac_commands_check.rs` · tools 슬라이스 2

`ac-commands-check --tree <root>` 는 기록된 `cargo test --test <name> …` 검증 명령을 훑어 존재하지 않는 테스트 대상, 목록이 빈 대상, 하나도 고르지 못하는 필터를 걸어요 — cargo 는 아무것도 고르지 못한 필터에서도 exit 0 ("running 0 tests")이라 무(零)선택 명령이 검증으로 통과한 일이 있었어요(#2352, survey 2.11). `--file <파일>` 은 한 파일만 심판하고 대상 존재는 부른 쪽의 작업 폴더 기준이에요. 훑는 곳은 `docs/**/*.md`, `.moai-cli/reports/*.md`, 2153 번 이상의 issue 본문 기록(얼린 RC1 기록은 제외)이에요. 대상별 cargo 목록(`-- --list`)은 한 번씩만 돌아요(호출별 캐시). 결과는 봉투(envelope)의 result 안에 요약 줄 하나와 위반 줄들 — 문구는 옮겨진 파이썬 도구와 같아요. 쓰는 것은 없어요.
