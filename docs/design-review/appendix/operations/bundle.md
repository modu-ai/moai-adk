# `bundle`

operation `bundle` · 코드 `src/ops/bundle.rs`, `src/bundle.rs` · 카드 C-119, SPEC-20261119-bundle-tool(계획), SPEC-20261119-bundle-ready-check(조각 2)

`bundle <동작>` 은 M-Flow 묶음 PR 을 조립하는 리더를 돕는다. 레인이 카드에 `LANE-READY-INTEGRATION` 줄을 남긴 head 를 읽고, head 마다 검사하고, 묶음 PR 본문을 사실에서 만든다. 봉투의 `operation` 은 항상 `bundle`, 동작은 결과의 `action` 으로 나간다. 모든 동작의 도움말 첫 문장은 `Reads only.` 로 시작한다(`body --out` 만 `Writes one new file.` 을 덧붙인다). **어떤 동작도 fetch, push, PR 열기·고치기·병합, ref·가지 만들기·옮기기, 카드 노트와 원장 행 쓰기, 설정 바꾸기를 하지 않는다.** head 고르기, push, PR 은 리더의 몫이고 CI·독립 감사·ruleset 은 그대로다. git 은 ref·색인·작업 트리를 쓰지 않는 명령만, 환경을 비운 채 부른다(`audit_delta::command`). 다만 `check`·`body` 가 쓰는 `git merge-tree --write-tree`(git 2.38 이상)는 어느 ref 도 가리키지 않는 tree·blob 객체를 객체 저장소에 남긴다(`git gc` 가 지운다).

## `bundle ready [--project .] [--home H] [--dir .] [--tip origin/v2026.11.0]`

카드마다 첫 줄이 `LANE-READY-INTEGRATION` 으로 시작하는 노트 중 번호가 가장 큰 것(가장 새 줄)을 읽는다. 읽히는 모양은 `LANE-READY-INTEGRATION C-<n> head=<40자리 소문자 16진수>` 다음에 줄 끝이나 공백이 오는 것이고, head 바로 뒤에 ` spec=<SPEC id>` 가 올 수 있다(없어도 된다). 노트가 놓인 카드와 줄의 카드가 같아야 한다. 가장 새 줄이 읽히지 않으면 그 카드는 `unreadable`(카드, 노트 번호)에 올라가고, 더 오래된 줄의 head 는 쓰지 않는다. 결과: `heads[]`(`card`, `note_id`, `head`, `spec`, `specs`(`<tip>..<head>` 의 `Moai-Spec` trailer), `present`, `on_origin`, `already_in_tip`, `ready_lines`), `unreadable[]`, `tip`(이름, SHA, 커밋 시각). 도구는 fetch 하지 않는다: git 은 worktree 가 함께 쓰는 fetch 시각을 남기지 않으므로 tip 의 SHA 와 커밋 시각을 보여 주고, 부르기 전에 fetch 하는 것은 부르는 쪽이다.

## `bundle check HEADS… [--tip …]`

HEADS 는 `C-<n>`(카드의 가장 새 읽히는 READY head) 또는 `C-<n>=<40자리 16진수>`, 합칠 순서대로 준다. head 마다 모든 검사를 하고 발견을 한 결과에 모은다(묶음 5 가 검사 하나가 실패할 때마다 전체가 멈춰 2시간을 잃은 데서 온 규칙). `findings[]` 의 한 항목은 `{code, card, head, detail}`:

| code | 뜻 |
| --- | --- |
| `no_ready_line` | 카드에 읽히는 READY 줄이 없다 |
| `head_missing` | head 가 이 저장소에 없다(fetch 먼저) |
| `not_on_origin` | `origin` 의 어느 가지도 head 를 품지 않는다 |
| `no_merge_base` | head 가 tip 과 합칠 기준(merge base)이 없다(이어지지 않은 역사). 이 head 의 나머지 검사는 건너뛰고 다른 head 는 계속 본다 |
| `already_merged` | head 가 이미 tip 안에 있다 |
| `audit_missing` | `<tip>...<head>`(점 세 개) 의 변경 경로로 `gate::tier_of(Some(0), …)` 가 1 이상인데, 카드 노트 중 첫 줄이 `AUDIT ` 로 시작하고 `gate::parse_verdict` 로 읽히며 판정 `PASS` 이고 head 가 정확히 이 head 인 것이 없다. `detail.ancestor_pass` 는 조상 head 의 가장 가까운 `PASS` |
| `trailers` | `trailers::check` 가 `<tip>..<head>` 에서 실패를 낸다 |
| `spec` | 이름 붙은 SPEC 의 오류, head 에 spec.md 없음, `Moai-Spec` trailer 없음 |
| `result_missing` | SPEC 의 AC 중 `## Result` 에 줄이 없는 것 |
| `conflict` | tip 과 또는 앞선 head 와 병합하면 충돌한다. `detail.files`·`detail.with`. GitHub 가 돌리는 글자 병합으로 본다(복제본에 등록된 병합 드라이버를 덮는다) |

읽히지 않는 판정 노트(`AUDIT ` 로 시작하지만 읽히지 않음)는 `unreadable_verdict` 로 알리기만 하고 검사를 실패시키지 않는다. 충돌은 head 와 tip, head 와 앞선 head 를 둘씩 본다; 세 개 이상이 함께 만드는 충돌은 `build`(조각 3)의 몫이다. `PASS` 노트가 있다는 것만 보고, 누가 썼는지·어느 tier 인지는 알 수 없다(`audit_note`). `ok` 는 발견이 없을 때만 참이다. **결과로 알리는 발견은 exit 0 이다**(입력을 못 쓰면 — 없는 tip, 잘못된 head 인자 — exit 1, 사용법 오류는 2): `ok:false` 로 exit 1 을 내는 규칙은 `src/main.rs` 에 연산 이름으로 있고 그 파일은 다른 카드가 잡고 있어, 이 조각은 읽는 쪽이 `ok` 를 본다고 가정한다.

## `bundle body HEADS… --branch B --run FILE [--out FILE] [--tip …]`

묶음 가지 B 의 SHA 를 고정하고 PR 본문을 쓴다. 첫 줄은 `audit-tier: N <경로들>`(경로는 `<tip>...<B>`, N 은 `gate::tier_of(Some(0), 경로들)`), 이어서 SPEC 마다 `Moai-Spec:` 줄, head 표(SPEC, 카드, head, 감사 근거), `Integration:` 줄(가지, tip, 병합 커밋 수, 고정 SHA, 파일·줄 수), `## Self-Audit`, `## Result`. 사실은 head 와 실행 기록 파일에서만 온다.

실행 기록 파일은 JSON 객체 하나다: `commands`(`check`, `command`, `exit`, 선택 `counts`), `not_run`(`check`, `reason`), `self_audit`(`plan`, `red`, `mutation`, `clone`, `host`, `receipt`). 검사 키는 코드에 고정돼 시험이 묶는다: `build`, `lib_tests`, `home_isolation_scan`, `product_cli`, `fmt`, `clippy`, `trailer_check`, `spec_check`, `result_coverage`, `unittest`. 키마다 `commands` 에 있거나 `not_run` 에 이유와 함께 있어야 하고(둘 다는 안 된다), 키 없는 `commands` 항목은 `## Result` 에 덧줄로 간다.

빠진 사실이 하나라도 있으면 본문 전체를 거절하고 빠진 사실을 모두 한 번에 이름 붙인다(자리표시자를 쓰지 않는다): 가지에 없는 head, `check` 의 발견이 하나라도 있는 head(`audit_missing`, `spec`, `trailers`, `conflict`, `no_merge_base` …), 실행도 `not_run` 도 아닌 검사, 줄바꿈이 든 기록 값, exit 가 0 이 아닌 명령, 비어 있는 Self-Audit 줄, `## Result` 에 줄이 없는 AC. 만든 본문이 `self_audit::check` 를 통과하지 못해도 거절한다. `--out` 은 새 파일만 만들고 이미 있는 경로는 거절하며 파일을 그대로 둔다. `Integration:` 줄의 가지 크기는 tip 과의 merge base 에서 센다(가지를 만든 뒤 tip 이 움직였으면 "built from `<base>` (the tip … has moved on since)" 로 적는다). 본문에는 기록·head 에서 오지 않는 고정 문장이 있다: M-Flow 한 줄 소개, `built from`, `Local verification on`, `see the SPEC's Result`, 끝의 Claude Code 표지.
