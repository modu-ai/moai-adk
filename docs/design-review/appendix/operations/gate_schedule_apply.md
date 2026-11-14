# `gate-schedule-apply`

operation `gate.schedule_apply` · 코드 `src/ops/gate_schedule_apply.rs`, `src/gate_schedule.rs` · Issue #2906 S2

`gate-schedule-apply HOME --os macos|linux <plan 과 같은 입력> --plan-digest D --permission-reference R` 는 같은 입력으로 다시 계획해 digest 가 같을 때만 항목 파일 하나를 01-B 발행 엔진으로 쓴다. 권한 참조가 비었거나 digest 가 다르면 아무것도 쓰지 않는다. 쓰기 직전 파일 sha256 이 계획 때와 다르면 멈춘다. 영수증 `<상태 폴더>/schedule/receipts/<id>/receipt.json`(pending → committed, before/after sha256, 권한 참조)과 이전 바이트 사본 `target.before` 를 남긴다. `launchctl`·`crontab` 은 실행하지 않고 켜는·끄는 명령을 글로 돌려준다(결정 Q5). 결과: `state`(`applied`·`unchanged`), `plan_digest`, `target_path`, `receipt`, `written`, `enable`, `disable`.
