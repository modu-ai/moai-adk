# `bundle`

operation `bundle` · 코드 `src/ops/bundle.rs`, `src/bundle.rs`, `src/bundle/build.rs`, `src/bundle/verify.rs` · 카드 C-119, SPEC-20261119-bundle-tool(계획), SPEC-20261119-bundle-ready-check(조각 2), SPEC-20261119-bundle-build(조각 3), SPEC-20261119-bundle-verify(조각 4)

`bundle <동작>` 은 M-Flow 묶음 PR 을 조립하는 리더를 돕는다. 레인이 카드에 `LANE-READY-INTEGRATION` 줄을 남긴 head 를 읽고, head 마다 검사하고, 묶음 PR 본문을 사실에서 만들고, 묶음 가지를 만든다. 봉투의 `operation` 은 항상 `bundle`, 동작은 결과의 `action` 으로 나간다. `ready`·`check`·`body` 의 도움말 첫 문장은 `Reads only.` 로 시작한다(`body --out` 만 `Writes one new file.` 을 덧붙인다). `build` 는 가지와 worktree 를 쓰는 동작이고 도움말 첫 문장이 `Writes a new branch and worktree.` 다. `verify` 는 명령을 실행하는 동작이고 도움말 첫 문장이 `Runs commands.` 다. **어떤 동작도 fetch, push, PR 열기·고치기·병합, 카드 노트와 원장 행 쓰기, 설정 바꾸기를 하지 않는다. `ready`·`check`·`body` 는 ref·가지도 쓰지 않고, `build` 는 새 가지·새 worktree·병합 커밋만 쓴다.** head 고르기, push, PR 은 리더의 몫이고 CI·독립 감사·ruleset 은 그대로다. `build` 를 뺀 동작의 git 은 ref·색인·작업 트리를 쓰지 않는 명령만, 환경을 비운 채 부른다(`audit_delta::command`). 다만 `check`·`body` 가 쓰는 `git merge-tree --write-tree`(git 2.38 이상)는 어느 ref 도 가리키지 않는 tree·blob 객체를 객체 저장소에 남긴다(`git gc` 가 지운다).

## `bundle ready [--project .] [--home H] [--dir .] [--tip <tip>]`

`--tip` 을 비우면 통합 대상은 프로필이 정해요: m-flow 이고 `work_flow.integration_branch` 가 기록돼 있으면 `origin/<branch>`, 기록이 없거나 프로필을 읽을 수 없으면 `origin/v2026.11.0`, 명시한 `--tip` 이 항상 이긴다(SPEC-20261120-mflow-integration-branch-choice).

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
| `audit_tier_low` | head 의 `PASS` 노트가 모두 첫 줄에 tier 를 적었는데(아래) 가장 높은 것이 감사 하한(`audit_floor`)보다 낮다. `detail`: `floor`, `pass_tier`, `note_id` |
| `trailers` | `trailers::check` 가 `<tip>..<head>` 에서 실패를 낸다 |
| `spec` | 이름 붙은 SPEC 의 오류, head 에 spec.md 없음, `Moai-Spec` trailer 없음 |
| `result_missing` | SPEC 의 AC 중 `## Result` 에 줄이 없는 것 |
| `conflict` | tip 과 또는 앞선 head 와 병합하면 충돌한다. `detail.files`·`detail.with`. GitHub 가 돌리는 글자 병합으로 본다(복제본에 등록된 병합 드라이버를 덮는다) |

읽히지 않는 판정 노트(`AUDIT ` 로 시작하지만 읽히지 않음)는 `unreadable_verdict` 로 알리기만 하고 검사를 실패시키지 않는다. 충돌은 head 와 tip, head 와 앞선 head 를 둘씩 본다; 세 개 이상이 함께 만드는 충돌은 `build`(조각 3)의 몫이다. `PASS` 노트가 있다는 것을 보고 누가 썼는지는 알 수 없다(`audit_note`). **노트가 적은 tier**(SPEC-20261119-bundle-audit-tier, 리더 결정 D-7): 판정 줄의 마지막 ` head ` 앞에 `tier <N>`(N 은 0~2 한 자리, 앞은 줄 처음·공백·`(`·`,`, 뒤는 줄 끝·공백·`,`·`)`·`;`)이 있으면 읽고(`AUDIT C-<n> (<SPEC>, tier <N>) head <sha> base <sha> — verdict: PASS`) `heads[].pass_tier` 에 head 의 `PASS` 노트 중 가장 높은 값을 보인다. 모든 `PASS` 가 tier 를 적었고 가장 높은 것이 하한보다 낮으면 `audit_tier_low`; tier 가 없는 `PASS`·서로 다른 두 tier·2 를 넘는 숫자는 읽지 않고 이전처럼 하한을 채운다. 판정 줄은 `gate::parse_verdict` 가 그대로 읽는다(`src/gate.rs`·`src/audit_gate.rs` 는 바꾸지 않았다). tier 를 쓰는 일은 강제하지 않는다. `ok` 는 발견이 없을 때만 참이다. **결과로 알리는 발견은 exit 0 이다**(입력을 못 쓰면 — 없는 tip, 잘못된 head 인자 — exit 1, 사용법 오류는 2): `ok:false` 로 exit 1 을 내는 규칙은 `src/main.rs` 에 연산 이름으로 있고 그 파일은 다른 카드가 잡고 있어, 이 조각은 읽는 쪽이 `ok` 를 본다고 가정한다.

## `bundle body HEADS… --branch B --run FILE [--out FILE] [--tip …]`

묶음 가지 B 의 SHA 를 고정하고 PR 본문을 쓴다. 첫 줄은 `audit-tier: N <경로들>`(경로는 `<tip>...<B>`, N 은 `gate::tier_of(Some(0), 경로들)`), 이어서 SPEC 마다 `Moai-Spec:` 줄, head 표(SPEC, head, 감사 근거; 카드 열은 없다), `Integration:` 줄(가지, tip, 병합 커밋 수, 고정 SHA, 파일·줄 수), `## Self-Audit`, `## Result`. 사실은 head 와 실행 기록 파일에서만 온다.

실행 기록 파일은 JSON 객체 하나다: `commands`(`check`, `command`, `exit`, 선택 `counts`), `not_run`(`check`, `reason`), `self_audit`(`plan`, `red`, `mutation`, `clone`, `host`, `receipt`). 검사 키는 코드에 고정돼 시험이 묶는다: `build`, `lib_tests`, `home_isolation_scan`, `product_cli`, `fmt`, `clippy`, `trailer_check`, `spec_check`, `result_coverage`, `unittest`. 키마다 `commands` 에 있거나 `not_run` 에 이유와 함께 있어야 하고(둘 다는 안 된다), 키 없는 `commands` 항목은 `## Result` 에 덧줄로 간다.

머리 표의 감사 칸은 head 의 가장 새 `PASS` 노트가 있을 때 `` `PASS` on `<head>` `` 로 쓰고, 그 노트의 줄이 tier 를 적었으면 `(tier <N>)` 을 뒤에 보인다(적지 않았으면 보이지 않는다). 본문은 밖으로 나가는 글이라 카드 번호, 노트 번호, 세션 id 를 쓰지 않는다(SPEC-20261120-board-bundle-link): 표에 카드 열이 없고 감사 칸은 노트를 인용하지 않으며, 줄마다 `<SPEC id>: …` 로 시작하는 `## Result` 의 SPEC 줄에도 카드가 없고, 실행 기록의 명령 글 속 카드 번호(`bundle check C-1 C-2`)는 자리표시자 `<lane>` 으로 바뀐다. 카드 번호는 `C-12` 와 `c-12` 둘 다 같다(카드 해석기가 둘 다 받는다). 그 밖의 글(Self-Audit 값, not-run 사유, 가지 이름 등)에 카드 번호나 세션 id 꼴이 남아 있으면 본문 전체를 거절하고 아무것도 쓰지 않으며, 거절 글은 어느 필드인지만 말하고 값은 되풀이하지 않는다. 빠진 사실이 하나라도 있으면 본문 전체를 거절하고 빠진 사실을 모두 한 번에 이름 붙인다(자리표시자를 쓰지 않는다): 가지에 없는 head, `check` 의 발견이 하나라도 있는 head(`audit_missing`, `spec`, `trailers`, `conflict`, `no_merge_base` …), 실행도 `not_run` 도 아닌 검사, 줄바꿈이 든 기록 값, exit 가 0 이 아닌 명령, 비어 있는 Self-Audit 줄, `## Result` 에 줄이 없는 AC. 만든 본문이 `self_audit::check` 를 통과하지 못해도 거절한다. `--out` 은 새 파일만 만들고 이미 있는 경로는 거절하며 파일을 그대로 둔다. `Integration:` 줄의 가지 크기는 tip 과의 merge base 에서 센다(가지를 만든 뒤 tip 이 움직였으면 "built from `<base>` (the tip … has moved on since)" 로 적는다). 본문에는 기록·head 에서 오지 않는 고정 문장이 있다: M-Flow 한 줄 소개, `built from`, `Local verification on`, `see the SPEC's Result`, 끝의 Claude Code 표지.

## `bundle build HEADS… [--date YYYYMMDD] [--project .] [--home H] [--dir .] [--tip origin/v2026.11.0]`

새 가지 `m-flow/bundle-<날짜>-<n>` 과 새 worktree `<HOME>/.moai/worktrees/<프로젝트>-bundle-<n>` 을 tip 의 커밋에서 만들고, HEADS(`C-<n>` 또는 `C-<n>=<40자리 16진수>`)를 준 순서대로 head 마다 병합 커밋 하나씩 합친다. `<날짜>` 는 오늘(UTC)이고 `--date` 로 정한다(여덟 자리, 월 01~12, 일 01~31). `<n>` 은 1~999 중 다음이 모두 없는 가장 작은 수다: 같은 이름의 로컬 가지, `origin` 가지, 등록된 worktree(폴더가 사라진 것도), 그 폴더; 999 개가 모두 쓰였으면 아무것도 만들지 않고 거절한다. 병합은 GitHub 가 돌리는 글자 병합이다: git 설정(복제본의 것과 사용자의 것)에 이름이 있는 병합 드라이버는 모두 글자 병합으로 덮는다(`git config --get-regexp '^merge\..*\.driver$'`). `check` 의 충돌 검사도 복제본 설정에 이름이 있는 드라이버를 모두 덮는다(전역 설정을 끈 채 돌리므로 복제본의 설정이 적용되는 전부다). **이미 있는 가지·ref·worktree·폴더는 옮기거나 지우거나 강제하지 않는다**: 다시 만들 때는 늘 새 번호의 새 가지다. 쓰는 것은 새 가지, 그 worktree, 병합 커밋뿐이고 fetch·push·PR·카드 노트·원장 행·설정은 건드리지 않는다. 쓰기 git 은 사용자의 작성자 신원과 훅 설정을 따르려고 `identity::scrub_git_env` 로만 환경을 정리해 부른다(`--no-verify` 를 쓰지 않는다).

결과 키: `action`, `number`, `branch`, `worktree`, `tip`, `pinned`(병합을 마친 뒤 `git rev-parse` 로 읽은 가지의 SHA), `merged[]`(`card`, `head`, `merge`), `dropped[]`(`card`, `head`, `files`), `skipped[]`(`card`, `head`, `reason`: `in_tip` 또는 `in_branch`), `rebuild`, `ok`, `writes`. 병합이 충돌한 head 는 `git merge --abort` 로 되돌려 가지에 넣지 않고 `dropped` 에 충돌 파일과 함께 올리며 다음 head 로 간다. `ok` 는 `dropped` 가 비었을 때만 참이고(`skipped` 는 실패가 아니다), `rebuild` 는 떨어뜨린 head 만 뺀 같은 빌드의 명령줄(`moai bundle build C-<n>=<sha> … --tip <tip>`, 셸 문자가 든 낱말은 작은따옴표로 감싼다)이거나, 떨어뜨린 것이 없거나 tip 에 없는 head 가 하나도 남지 않으면 빈 문자열이다. 태그 이름이 head 면 태그가 가리키는 커밋을 합치고 그 SHA 를 보고한다.

만들기 전에 쓸 수 없는 head 를 한 번에 모두 이름 붙여 거절한다(읽히는 READY 줄이 없는 카드, 이 저장소에 커밋이 없는 head, 같은 카드를 두 번 준 것, 모두 이미 tip 에 있어 합칠 것이 없는 경우): 가지도 폴더도 생기지 않는다. 가지를 만든 뒤 git 이 실패하면(작성자 신원 없음, 병합 커밋을 거절하는 훅, 되돌리기 실패) 멈추고 남겨 둔 가지와 worktree 를 이름 붙이며 아무것도 지우지 않는다. `ok:false` 로 exit 1 을 내는 규칙은 연산 이름으로 `src/main.rs` 에 있어 이 동작도 결과로 알리는 발견은 exit 0 이다.

## `bundle verify HEADS… --branch B [--out DIR] [--self-audit FILE] [--moai-cli PATH] [--python python3] [--slot-dir DIR] [--max N] [--min-free-gb N] [--project .] [--home H] [--dir .] [--tip origin/v2026.11.0]`

도움말 첫 문장은 `Runs commands.` 다. 가지 B(`build` 가 만든 것)가 체크아웃된 worktree 안에서 검사 목록을 모두 돌리고, 돌린 결과를 새 폴더에 남긴다. 폴더 기본값은 `<worktree>/target/bundle-verify/<가지 SHA 앞 10자리>` 이고 `--out` 으로 바꾼다. 명령은 `heavy-slot-acquire`(이 실행 파일 자신의 것, `--slot-dir`·`--max`·`--min-free-gb` 를 그대로 넘긴다) 한 번의 점유 안에서 차례로 돈다. 목록은 `CHECKS` 열 개 키 순서의 데이터다: `build`=`cargo build --locked --offline --bin moai-cli`, `lib_tests`=`cargo test --locked --offline --lib`, `home_isolation_scan` 과 `product_cli`=`cargo test --locked --offline --test <이름>`, 이어서 tip 과 가지 사이에서 추가·변경된 `tests/<이름>.rs` 와 그 사이에 바뀐 SPEC 의 `verify:` 줄이 이름 붙인 `--test` 대상(같은 형식, 키 없음), `fmt`=`cargo fmt --check`, `clippy`=`cargo clippy --locked --offline --all-targets -- -D warnings`, `trailer_check`=`<moai-cli> trailer-check <tip>..HEAD`, `spec_check`=`<moai-cli> spec check --all`, `result_coverage`=`<moai-cli> spec check <SPEC id> --closure`(가지의 `Moai-Spec` trailer 가 가리키는 SPEC 마다 하나), `unittest`=`python3 -m unittest discover -s tests -p '*_test.py'`. `<moai-cli>` 기본값은 `<CARGO_TARGET_DIR 또는 worktree 의 target>/debug/moai-cli` 다.

거절은 아무것도 실행하지 않고 아무것도 쓰지 않으며 모든 이유를 한 번에 말한다(exit 1): 가지가 없음, 가지가 체크아웃된 worktree 가 없음(다른 HEAD 인 worktree 는 세지 않는다), worktree 의 HEAD 가 가지 SHA 와 다름, 추적 파일에 커밋 안 된 변경이 있음, head 가 tip 에도 가지에도 없음, 출력 폴더가 이미 있음, `--self-audit` 가 읽히는 JSON 객체가 아님. 명령 하나가 실패해도 목록은 끝까지 간다. 빌드가 실패했거나 `<moai-cli>` 가 실행 파일이 아니면 그 파일이 필요한 `trailer_check`·`spec_check`·`result_coverage` 는 돌지 않고 `not_run` 에 이유와 함께 오른다(실패가 아니다); 가지가 SPEC 을 하나도 가리키지 않으면 `result_coverage` 도 같고, 파일이 없는 SPEC 대상도 같다.

폴더에는 `run.sh`, 명령마다 `<번호>-<이름>.log`, 그리고 `run.json` 이 생긴다. `run.json` 은 `bundle body --run` 이 읽는 기록이다: `commands[]`(`command`, `exit`, `started`, `ended`, `log`, 키가 있으면 `check`, 읽히면 `counts`), `not_run[]`(`check`, `reason`), `self_audit`(`--self-audit` 파일의 객체, 없으면 `{}`), 그리고 `pinned`·`branch`·`tip`. 결과의 `failed[]` 는 실패한 명령마다 `check`, `command`, `exit`, `log`, `heads` 를 가진다: `heads` 는 출력에 이름이 나온 파일 가운데 가지가 바꾼 파일을 `<tip>...<head>` 로 바꾼 head 의 카드이고, 어느 head 도 아니면 `unattributed` 하나다. `ok` 는 `failed` 가 비었을 때만 참이다. 이 동작은 fetch, push, PR, 카드 노트, 원장 행, 설정, ref 를 쓰지 않는다(`ok:false` 도 exit 0).
