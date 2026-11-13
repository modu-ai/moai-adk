# `flake-classify`

operation `flake.classify` · 코드 `src/ops/flake_classify.rs`, `src/flake.rs` · Issue #2666 F1

`flake-classify --failed LOG --later LOG [--job NAME]` 는 같은 head 의 실패한 CI run 로그와 뒤의 run 로그(`gh run view <id> --log` 를 저장한 파일, cargo-nextest 출력)를 받아, 앞에서 실패하고 뒤에서 통과한 시험을 `flaky`, 둘 다 실패한 시험을 `still_failing`, 뒤 로그에 없는 시험을 `not_in_later` 로 나눈다. 로그 한 줄은 `<job>\t<step>\t<시각> <글>` 이고 job 마다 따로 본다. 실패 시험에는 `panicked at <위치>` 줄과 그 다음 줄을 함께 적는다. 두 파일(각 64 MiB 이하)만 읽고 아무것도 쓰지 않는다. VERIFY workflow 의 불안정 판정 규칙을 도구로 옮긴 것이다.

같은 head 확인(감사 6858491683 B1): job 마다 로그에서 head 를 읽는다. PR run 은 체크아웃 줄 `HEAD is now at <x> Merge <PR head> into <base>` 의 PR head, 그 밖에는 `git log -1 --format=%H` 다음 줄의 SHA 다. 앞에서 실패하고 뒤에서 통과한 시험은 두 head 가 모두 읽히고 같을 때만 `flaky` 이고, 다르면 `head_mismatch`(고친 것이지 흔들린 게 아님), 하나라도 못 읽으면 `head_unverified` 다. 결과에 `first_heads`·`later_heads` 와 로그마다 읽은 상태 줄 수 `status_lines` 를 함께 적는다(0 이면 nextest 출력이 아님). nextest 재시도(`retries`)를 켜면 `TRY 2 PASS`·`FLAKY` 줄이 생기는데 지금 파서는 읽지 않는다 — 이 저장소 `.config/nextest.toml` 은 재시도를 켜지 않았다. 켤 때 파서를 함께 고친다.
