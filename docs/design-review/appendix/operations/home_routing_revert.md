# `home-routing-revert`

operation `home.routing_revert` · 코드 `src/ops/home_routing_revert.rs`, `src/home_routing.rs` (#2638)

`home-routing-revert HOME --receipt ID --permission-reference R` 는 영수증 하나를 되돌린다. 적용 뒤 바뀌지 않은 파일만: 만든 파일은 지우고, 바꾼 파일은 before 바이트로 돌리며 소유 기록도 적용 전 기준으로 돌린다. 그 뒤 사용자가 바꾼 파일은 `user_changed_since_apply`, 링크가 된 폴더·파일은 그 사유로 건너뛴다. 같은 영수증을 두 번 되돌리지 않는다. 결과: `state`(`reverted`), `receipt`, `reverted`, `skipped`.
