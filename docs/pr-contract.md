# PR 계약 (RC1)

이 문서는 `v2026.11.0`을 대상으로 여는 PR이 지켜야 할 본문 규칙을 정한다. 소유 Issue는 06-C #2007이다(2026-11-08 사용자 확정, `docs/decisions-2026-11-08.md`). 다른 Issue는 이 문서를 참조만 하고, 바꿀 일이 있으면 06-C에 요청한다.

## 본문은 렌더러가 만든다

GitHub는 기본 브랜치(`main`)에 있는 PR 템플릿만 자동으로 채운다. 그래서 `v2026.11.0` 대상 PR에는 `.github/pull_request_template.md`가 적용되지 않는다. 본문은 `moai pr-body-render`로 만들고, 그 결과를 파일로 저장해 `gh pr create --body-file <파일>`로 연다. 렌더러 출력의 절 구성은 템플릿과 같다.

```text
moai pr-body-render <worktree> (--task <key> | --record <file> --claim-comment <file>) \
  --receipts <receipts.json> --base-sha <40자> --base-branch v2026.11.0 --default-branch main \
  [--not-run <항목>]... [--review-link <url>]
```

## 반드시 들어가는 항목

| 항목 | 규칙 |
| --- | --- |
| Issue 연결 | 첫 줄에 `Refs #<n>`을 쓰고 Issue URL을 함께 적는다. |
| 대상 | 논리 키(예: `06-C`), epoch, claim 토큰 `Moai-Claim: <n>#<claim 댓글 id>` |
| SPEC | `spec_ref`의 id, revision, digest(없으면 `null`) |
| exact SHA | base는 `v2026.11.0` @ 40자 SHA, candidate는 렌더 시점의 worktree HEAD 40자 SHA다. 독립 review가 본 SHA와 같아야 한다. |
| 변경 | base와 candidate의 merge base부터 candidate까지 바뀐 경로와 git 상태(`git diff --name-status <base>...HEAD`) |
| 로컬 시험 증거 | VERIFY receipt(05-C #2003)의 검사마다 명령, exit code, 결과, 통과·실패 수를 적는다. receipt가 없으면 `not_run`으로 적는다. |
| 독립 review | 판정과 판정 댓글 링크를 적을 자리다. 판정이 나오기 전에는 대기로 둔다. |
| not_run / blocked | 실행하지 않은 검사와 그 이유, 기다리는 결정이나 HOLD를 적는다. 표에서 빼지 않는다. |
| 남은 gap | 관측하지 못한 것과, 관측했어도 여전히 틀릴 수 있는 것을 적는다. |
| op marker | 본문 끝에 `<!-- moai-op pr_create:<repository_id>:<task_key>:<epoch> -->`를 둔다. 06-B의 reconcile이 이 marker로 PR을 찾는다. 같은 claim epoch에서 PR을 하나 더 열면(후속 PR·범위 분할) 그 PR은 `pr-body-render --pr-seq N`(N은 2부터)으로 렌더링해 `<!-- moai-op pr_create:<repository_id>:<task_key>-<N>:<epoch> -->`를 쓴다(#2329). 첫 PR의 marker는 그대로이고, 기존 marker도 그대로 읽힌다. 같은 marker를 가진 PR이 둘이면 reconcile은 어느 쪽도 확정하지 않고 `uncertain`으로 남긴다. 이미 확정된 operation에 다른 PR의 ACK가 오면 `pr-confirm`은 충돌로 거부한다. |

## 자동 종료 키워드는 쓰지 않는다

`Closes`, `Fixes`, `Resolves`와 그 활용형은 쓰지 않는다. GitHub는 기본 브랜치 대상 PR에서만 이 키워드를 해석하므로, `v2026.11.0`에 병합해도 Issue는 닫히지 않는다. Issue 종료는 리더가 통합 기록을 보고 판정한다. 렌더러는 본문에서 이 단어들을 찾으면 본문을 내지 않고 거부한다. `--allow-close`를 주어도 키워드는 넣지 않고 경고만 낸다.

## PR 개설은 완료가 아니다

본문을 렌더링하는 것은 PR을 여는 것이 아니다. claim은 PR을 열고 그 ACK를 `pr-confirm`으로 기록할 때 `pr_ready`가 될 뿐이고, `integrated`나 완료로 올라가지 않는다. 통합은 리더가 독립 AUDIT 판정을 읽고, head가 같은지 확인한 뒤 `moai/audit` 상태를 붙여서 진행한다. 병합할 때는 merge REST의 `sha` 파라미터나 `gh pr merge --match-head-commit <candidate>`로 candidate SHA를 고정한다. head가 다르면 409로 거부되며, 이 고정은 08-A #2012가 넘겨받는 인터페이스다.

## claim record 갱신

06-C 명령은 바뀌는 내용을 항상 결과의 `record_update_request`로 돌려준다. record가 로컬 journal에서 왔을 때(`--task`)는 06-B가 공개한 `Journal::commit`으로 journal에도 기록한다. 기록 대상은 다음과 같다.

- 측정한 worktree identity: `worktree_recorded` 이벤트
- `blocked(reason=scope_conflict)` 또는 통합 순서: `scope_check` 이벤트
- 렌더 시점의 `pr_create` effect intent: `effect_intend` 이벤트
- PR ACK(`pr-confirm --task <key> --operation-id <id> --response-id <PR 번호> --observed-state opened`): intended 상태인 effect를 06-B `effect_confirm`으로 확인한다. reconcile이 이미 confirmed로 바꾼 effect는 그대로 받는다. effect가 confirmed일 때만 `pr_ready` 이벤트를 쓴다. 응답을 잃어 `uncertain`이면 상태를 그대로 두고, `effects-reconcile`이 먼저 판정한다.

06-A의 acquire·release 동작은 바꾸지 않고 이벤트만 덧붙인다. 같은 명령을 다시 실행해 record가 그대로이면 이벤트를 쓰지 않는다.

`pr_create` intent의 payload는 본문이 아니라 무엇을 만드는지(task, epoch, Issue, base 브랜치)를 담는다. 그래서 커밋이 늘어 본문이 바뀌어도 같은 intent로 취급된다. 본문 marker(`moai-op`)와 06-B가 기록하는 marker(`moai:op`)는 표기가 다르다. 그래도 06-B reconcile은 marker를 파싱한 operation으로 비교하므로(#2101) 둘을 같은 marker로 본다. 다만 06-B의 기본 reconcile은 Issue 댓글만 조회한다. 그래서 `pr_create`는 판정하지 않고 `uncertain` 그대로 두며, `not_run`(다른 effect가 막혔으면 `blocked`)과 `--remote-probe`가 필요하다는 안내를 낸다(#2101). PR 본문으로 확인하려면 `--remote-probe` 명령을 넘겨야 한다. 판정이 confirmed가 되면 `pr-confirm`으로 `pr_ready`를 기록한다.

## record 파일은 권위가 아니다

리더가 중계한 claim처럼 로컬 journal을 거치지 않은 claim은 `--record <file>`로 넘길 수 있다. 이 파일은 레인이 만든 입력일 뿐 권위가 아니다. 그래서 반드시 `--claim-comment <file>`(Issue claim 댓글의 `id`와 `body`)을 함께 받는다. comment id, session, epoch, lease, claim 토큰 가운데 하나라도 댓글과 다르면 record를 거부한다. 원장은 Issue의 claim 댓글이다.

## 아직 정하지 않은 것

- branch prefix는 확정됐다(리더, 사용자의 "HOLD만 해제" 결정, #1975 댓글 6721834190). `worktree-plan`은 `--branch-prefix`를 기본값 없이 필수로 받고, RC1 레인은 `issue/<n>-<key>`를 쓴다.
- 템플릿을 단일 파일로 둘지 디렉터리로 둘지도 HOLD다(REQ-06C-13). 지금은 단일 파일 `.github/pull_request_template.md`(00-E #2020 1차 작성)를 그대로 쓴다.
