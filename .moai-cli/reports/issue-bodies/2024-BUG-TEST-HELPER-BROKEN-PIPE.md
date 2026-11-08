<!-- moai-template:moai-bugfix@0.3.0; 2026-11-08 -->
```yaml
logical_key: BUG-TEST-HELPER-BROKEN-PIPE
milestone: 2026.11.0-rc.1
parent_key: EP-02
work_kind: bugfix
related: [00-E, 01-B]
claim: {state: unclaimed}
evidence_status: not_run
```
# [RC1][결함] `tests/product_cli.rs` `cli()` 헬퍼가 자식이 먼저 종료하면 BrokenPipe로 panic — rc-ci 첫 실행 실패

## 재현(GitHub Actions, PR #2021 rc-ci run 42464362547, 2026-11-08 08:47Z)
`cargo test --offline --locked` → `pin_precedence_integrity_and_no_fallback` FAILED: `tests/product_cli.rs:28:10 called Result::unwrap() on an Err value: Os { code: 32, kind: BrokenPipe }`. 21 passed · 1 failed. 로컬 Mac에서는 통과(22/22) — 러너에서 자식 `moai-cli`가 인자 검증 실패로 stdin을 읽기 전에 종료해 부모의 `write_all`이 EPIPE를 받는 타이밍 차이.

## 원인(코드 읽기)
`cli()` 헬퍼가 `child.stdin.take().unwrap().write_all(...).unwrap()`로 stdin 쓰기 실패를 panic으로 올린다. 시험 대상은 자식의 exit/stdout/stderr이지 stdin 쓰기 성공이 아니다.

## 수정 방향
- `cli()`에서 stdin 쓰기 오류 중 `ErrorKind::BrokenPipe`는 무시(`let _ =` 또는 match), 다른 오류는 그대로 panic. `wait_with_output()` 결과만 단언.
- 동일 패턴 헬퍼가 다른 테스트 파일에 있으면 함께.
- 회귀 시험: 인자 오류로 즉시 종료하는 호출(예: 존재하지 않는 package 경로)에 큰 stdin(>64KiB)을 써도 panic 없이 비정상 exit를 관찰.

## 인수 기준
- AC-PIPE-1: 위 회귀 시험 통과 · 검증: `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-target-<label> cargo test --offline --locked --test product_cli stdin_broken_pipe_does_not_panic` · 결과: not_run
- AC-PIPE-2: 전체 검사 묶음 통과(22+1) · 결과: not_run
- AC-PIPE-3: 통합 뒤 PR #2021(00-E)이 update-branch 후 rc-ci success · 결과: not_run
- AC-PIPE-4: 독립 AUDIT PASS → 에이전틱 게이트 통합 · 결과: not_run

## 범위
- 변경 가능: `tests/product_cli.rs`(헬퍼 1곳 + 시험 1개), 같은 패턴의 다른 `tests/*.rs`
- 변경 금지: `src/`(제품 동작 변경 없음)

상태 규칙: label·assignee는 lock이 아니다. 실행하지 않은 검증은 not_run.
