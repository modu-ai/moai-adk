# `gate-schedule-apply`

operation `gate.schedule_apply` · 코드 `src/ops/gate_schedule_apply.rs`, `src/gate_schedule.rs` · Issue #2906 S2

`gate-schedule-apply HOME --os macos|linux <plan 과 같은 입력> --plan-digest D --permission-reference R` 는 같은 입력으로 다시 계획해 digest 가 같을 때만 항목 파일 하나를 01-B 발행 엔진으로 쓴다. 권한 참조가 비었거나 digest 가 다르면 아무것도 쓰지 않는다. 쓰기 직전 파일 sha256 이 계획 때와 다르면 멈춘다. 영수증 `<상태 폴더>/schedule/receipts/<id>/receipt.json`(pending → committed, before/after sha256, 권한 참조)과 이전 바이트 사본 `target.before` 를 남긴다. `launchctl`·`crontab` 은 실행하지 않고 켜는·끄는 명령을 글로 돌려준다(결정 Q5). 쓰는 파일은 언제나 `<상태 폴더>/schedule/` 아래다: macOS plist 도 `~/Library/LaunchAgents` 가 아니라 `<상태 폴더>/schedule/<label>.plist` 에 쓴다 — launchd 는 로그인 때 `~/Library/LaunchAgents` 를 읽어 올리므로 거기 쓰는 것이 곧 켜기이고, 켜기는 사람이 출력된 `cp` 와 `launchctl bootstrap` 으로 한다. 대상 파일 자리에 링크가 있으면 거절한다.

경고: `gate-loop --fetch`(S3)가 생기기 전에는 예약 실행이 fetch 하지 않는다(`gate-loop` 은 head 가 이미 fetch 되어 있다고 본다). 그래서 `--merge` 항목을 켜면 오래된 `origin/<base>` 를 기준으로 게이트·병합한다. S3 전까지는 `--dry-run` 항목을 켠다.

결과: `state`(`applied`·`unchanged`), `plan_digest`, `target_path`, `receipt`, `written`, `enable`, `disable`.
