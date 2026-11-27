# `release-audit-summary`

operation `release.audit.summary` · 코드 `src/ops/release_audit_summary.rs`, `src/release_audit_summary.rs` · SPEC-20261125-release-audit-summary

`release-audit-summary PROJECT --base <40hex> --head <40hex> [--home DIR] [--post --pr N --repo OWNER/REPO --plan-digest D --permission-reference R] [--allow-incomplete] [--require-complete]` 는 릴리스 PR 범위의 판정을 원장에서 모아 PR 댓글 하나의 글로 만든다. 로컬 git 의 `base..head` first-parent 커밋마다 정확히 한 줄(오래된 것 먼저)을 내고, 줄은 `PASS`(그 커밋이 닫은 카드마다 착지 검사 `card::land::check::judge` 와 같은 규칙으로 독립 감사자의 PASS 가 있고, 줄마다 카드를 쓰거나 예약한 세션을 제외 목록으로 적는다), `ABS`(더 새로운 착지에 흡수된 빨리감기 레인의 중간 커밋; `card.land_started` 의 `tip` 을 알 때만), `NONE`(통과한 판정이 없는 착지, 행이 없는 직접 커밋, 감사자가 작성자나 보유자인 판정), `FAIL`(판정 FAIL 이 서 있는 착지)이다. 한 커밋이 카드 여럿을 닫았으면 한 줄에 모두 나오고, 그중 하나라도 통과하지 못했으면 `PASS` 가 아니다. 줄 문법은 `render` 와 `parse`(`src/release_audit_summary.rs`)가 한 곳에서 정하고, `audit-gate` 는 같은 `parse` 로 읽는다.

기본 실행은 미리보기이다: `text`(올릴 댓글 그대로, 마커 포함), `plan_digest`(그 바이트의 sha256), `landings`, `counts`(`pass`·`abs`·`none`·`fail`), `missing[]`(통과하지 못한 줄의 커밋, 상태, 카드, 이유), `complete`, `writes: false`, `network: "none"` 을 내고 `gh` 를 시작하지 않으며 원장을 쓰지 않는다. `--require-complete` 는 `PASS` 가 아닌 줄이 있으면 오류로 끝낸다. `base` 가 `head` 의 first-parent 줄에 없으면, 범위에 커밋이 없으면, 댓글이 60000바이트를 넘으면(자르지 않고) 거절한다.

`--post` 는 `--pr`, `--repo`, 같은 실행이 낸 `--plan-digest`, 쓸모 있는 `--permission-reference` 가 모두 있어야 한다. 계획을 다시 만들어 digest 를 비교하므로 계획 뒤에 원장이나 범위가 바뀌었으면 아무것도 올리지 않는다. `PASS` 가 아닌 줄이 있는 요약은 `--allow-incomplete` 없이는 올리지 않는다(한 head 의 모든 요약이 통과해야 해서 잘못 올린 것은 고칠 수 없기 때문이다). 올리기는 `GithubConnector::post_summary` 로 한 번이고, 마커 `<!-- moai-release-audit-summary <head> -->` 가 이미 구성원(또는 토큰 계정)의 댓글에 있으면 `already_posted` 로 아무것도 올리지 않는다.
