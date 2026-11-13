# `decision-file-check`

operation `decision.file_check` · 코드 `src/ops/decision_file_check.rs` · Issue #2619

`decision-file-check FILE` 는 결정 묶음 파일(`docs/decisions/*.md`)의 표를 검사한다. 머리글 `| id | what | options | default | reversible | category | deadline | decider | answer | evidence |`, 필수 칸, 선택지와 기본값·답, 범주(`none`·`guarded`·5범주), 기한 형식, 그리고 되돌릴 수 없거나 `guarded`·5범주인 결정은 `decider` 가 user 인지 본다. 읽기만 한다. 문제가 있으면 `ok:false` 이고 줄 번호와 이유를 메시지에 담는다. 통과하면 `rows`.
