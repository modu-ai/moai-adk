# `flake-classify`

operation `flake.classify` · 코드 `src/ops/flake_classify.rs`, `src/flake.rs` · Issue #2666 F1

`flake-classify --failed LOG --later LOG [--job NAME]` 는 같은 head 의 실패한 CI run 로그와 뒤의 run 로그(`gh run view <id> --log` 를 저장한 파일, cargo-nextest 출력)를 받아, 앞에서 실패하고 뒤에서 통과한 시험을 `flaky`, 둘 다 실패한 시험을 `still_failing`, 뒤 로그에 없는 시험을 `not_in_later` 로 나눈다. 로그 한 줄은 `<job>\t<step>\t<시각> <글>` 이고 job 마다 따로 본다. 실패 시험에는 `panicked at <위치>` 줄과 그 다음 줄을 함께 적는다. 두 파일(각 64 MiB 이하)만 읽고 아무것도 쓰지 않는다. VERIFY workflow 의 불안정 판정 규칙을 도구로 옮긴 것이다.
