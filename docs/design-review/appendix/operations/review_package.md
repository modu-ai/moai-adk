# `review-package`

operation `review.package` · 코드 `src/ops/review_package.rs` · Issue #2614

`review-package PROJECT --base B --head H` 는 감사용 검토 묶음을 결정적으로 만든다: merge-base 에서 HEAD 까지의 diff, 바뀐 파일, 그리고 BASE 에서만 읽은 검토 지시(`REVIEW.md` 와 그 include)·맥락(`AGENTS.md`, `CLAUDE.md`). git 을 읽기만 하고 쓰지 않으며 stdin 은 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R1).
