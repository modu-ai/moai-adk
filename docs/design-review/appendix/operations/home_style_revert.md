# `home-style-revert`

operation `home.style_revert` · 코드 `src/ops/home_style_revert.rs`, `src/home_style.rs` (#2875)

`home-style-revert HOME --receipt ID --permission-reference R` 는 영수증 하나를 되돌린다. 스타일 파일·관리 구간은 적용 뒤 바뀌지 않았을 때만: 만든 파일은 지우고, 바꾼 파일은 before 바이트로 돌리며 소유 기록도 적용 전 기준으로 돌린다. `outputStyle` 은 아직 `MoAI-Easy` 일 때만 그 키 하나를 지우고(다른 키는 그대로), 적용이 만든 `settings.json` 이 그대로면 파일을 지운다. 그 뒤 사용자가 바꾼 것은 `user_changed_since_apply`, 링크가 된 폴더·파일은 그 사유로 건너뛴다. 같은 영수증을 두 번 되돌리지 않는다. 결과: `state`(`reverted`), `receipt`, `reverted`, `skipped`.
