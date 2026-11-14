# `keys-revert`

operation `keys.revert` · 코드 `src/ops/keys_revert.rs`, `src/update_keys.rs` · Issue #2871 (I2 S3)

`keys-revert HOME --receipt ID --permission-reference REF` 는 `moai update` 의 keys 층이 vault 로 가져온 항목(`keys_result.receipt`)을 되돌린다. 영수증에 적힌 봉인 항목의 digest 와 지금 파일이 같을 때만 그 항목 하나를 지우고, 그 뒤 `key-set` 등으로 바뀐 항목은 `changed_since_apply` 로 건너뛴다. 원본 `.env.<provider>` 는 건드리지 않는다. 같은 영수증은 한 번만 되돌린다. 값은 어디에도 나오지 않는다. 결과: `state`(`reverted`), `receipt`, `reverted`, `skipped`, `original_kept`.
