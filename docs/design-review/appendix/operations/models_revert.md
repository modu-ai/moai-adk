# `models-revert`

operation `models.revert` · 코드 `src/ops/models_revert.rs`, `src/update_models.rs` · SPEC-20261119-models-revert-command (card C-47)

`models-revert HOME --receipt ID --permission-reference REF` 는 `moai update` 의 모델 저장(`models.receipt`)이 바꾼 `config/launcher.yaml` 을 저장 전 모양 그대로 되돌린다(저장 전에 파일이 없었으면 파일을 지운다). 파일이 아직 그 저장이 쓴 해시와 같을 때만 되돌리고, 그 뒤 바뀐 파일은 `state: skipped`, `reason: changed_since_apply` 로 건드리지 않는다(표지 파일도 안 쓴다). 사유 없이 거절하는 경우: 빈·자리표시 참조, `models-…` 모양이 아닌 id, 없는 영수증, 다른 operation 의 영수증, 다른 설정 경로, 앞 필드가 빠졌거나 짝이 안 맞는 영수증, 이미 되돌린 영수증. 설정 파일 내용은 어디에도 출력되지 않는다.
