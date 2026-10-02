# 절체 런북 — SPEC-GITHUB-FLOW-DEFAULT-001 (M6)

> 이 문서는 git-flow(develop 통합 브랜치)에서 github-flow(main 단일 기준)로 기준 브랜치를 바꾸는 절차를 순서대로 적은 **문서**다. 이 카드(t1453)의 레인은 이 문서를 **실행하지 않는다** — 쓰고, 점검 스크립트와 스크래치 리허설로 git 기계 장치를 검증할 뿐이다(REQ-GFD-020). 실행은 리더와 운영자가 한다. plan 산출물(`spec.md`·`plan.md`·`acceptance.md`·`design.md`·`research.md`)은 감사 해시가 걸려 있어 이 문서가 고치지 않는다. 이 문서는 해시가 걸리지 않는 새 파일이다.
>
> 근거: design D-8(절체 순서)·D-13(develop 퇴역)·D-17(선행 조건)·D-18(폴백)·D-22(Multi-OS 게이트)·D-25(이 카드의 면제)·D-26(헌법 개정)·D-28(CI 증거 고정)·D-29(행마다 주체 하나), REQ-GFD-018·019·020, AC-GFD-018·019·020·023.

## 읽는 법

- **행마다 실행 주체는 하나다**(카드·리더·운영자 중 하나, D-29). 한 단계에 두 주체가 필요하면 접미 번호 행으로 나눈다(`5a` 리더·`5b` 운영자, `9a` 운영자·`9b` 리더).
- **외부 공유 시스템 여부** 열은 `예`나 `아니오`로 시작한다. `예`는 이 단계가 GitHub 나 원격 브랜치처럼 다른 세션이 함께 쓰는 시스템을 바꾼다는 뜻이다. 읽기만 하는 단계는 `아니오`다.
- **기록할 증거** 열은 그 단계를 실행한 쪽이 확인 기록에 남기는 것이다. 명령과 그 출력이 증거이고, 요약은 증거가 아니다. 기록하지 못한 관측은 Gap 으로 적는다.
- **확인 기록**: `.moai/reports/t1453/cutover-confirmation.md`. 단계 이름이나 시각이 아니라 **관측한 선행 조건**(명령과 출력)을 적는다. 이 파일이 없으면 해당 단계는 미실행으로 읽는다(plan.md §E). 0·2단계(develop push)의 선행 조건은 origin/develop 팁 CI 가 녹색이라는 관측이고, 4단계 이후(수렴과 재기동)의 선행 조건은 3단계 사전 점검의 출력이다.
- 시간 예측은 쓰지 않는다. 단계 순서는 시간이 아니라 조건이다.

## 절체 순서

| # | 단계 | 주체 | 외부 공유 시스템 여부 | 기록할 증거 |
|---|---|---|---|---|
| 0 | 선행 조건 확인(D-17): 리더가 미푸시 로컬 develop 커밋을 일괄 push 하고 적색 CI 수리가 끝났는지 본다. origin/develop 팁의 CI 가 success 인지는 **팁 SHA 로 읽는다** — fetch 뒤 `git rev-parse origin/develop` 로 전체 SHA 를 얻고, `gh api "repos/modu-ai/moai-adk/actions/runs?head_sha=<전체 SHA>"` 로 그 SHA 의 런을 읽어 `gh run view <run id>` 로 conclusion 을 확정한다. 목록 호출은 같은 명령이 다른 창을 돌려준 사례가 있어 쓰지 않는다(D-28). 같은 SHA 의 `CI` 워크플로 런이 success 여야 하고, 다른 push 워크플로의 결론도 함께 적는다. 진행 중이면 기다리고, cancelled·failure 가 있으면 녹색이 아니다 | 리더 | 예(develop push) | 팁 전체 SHA, 위 두 명령의 출력(런 id·workflow·conclusion), 일괄 push 직전과 직후의 `git rev-list --count origin/develop..develop` 출력 |
| 1 | 레인 정지: 새 배차를 멈춘다. 각 레인은 진행 카드를 끝내고 증거를 hoist 한 뒤 리더가 `/clear` 를 요청하고(`/clear` 는 운영자가 치는 명령이다), 세션을 종료하고, 워크트리를 정리한다. 브랜치가 push 되고 원격 병합이 착지하기 전에는 어떤 워크트리도 폐기하지 않는다(카드 브랜치는 미푸시이면 워크트리가 유일한 사본이다). 병합되지 않은 picked 카드는 이 단계에서 병합·push 하거나 `unpick` 한다 | 리더 | 아니오 | 레인별 정지 시각이 아니라 정지 사실: `moai session list --json` 에서 카드 워크트리 cwd 의 활성 세션 0건, 레인별 증거 hoist 경로, 폐기하지 않고 남긴 워크트리 목록과 그 사유 |
| 2 | M4·M5 묶음을 현 경로(git-flow)로 develop 에 **마지막으로** 병합하고 push 한다(M1~M3 는 이미 병합돼 있다). 병합 창은 `moai integration acquire --name <lane> --card <card-id>` 로 잡고 보고 뒤에 `moai integration release` 로 돌려준다. push 뒤 develop 팁의 CI 를 0단계와 같은 방식(팁 SHA)으로 관측한다. push 트리거가 살아 있어(M5 가 `develop` 을 빼지 않는다) 관측할 수 있다 | 리더 | 예(develop push) | 병합 커밋 SHA 와 부모, push 뒤 `git rev-parse origin/develop`, 그 SHA 의 CI 런 id·conclusion |
| 3 | 배치 경계 사전 점검(AC-GFD-019): **2단계 병합·push 뒤, 4단계 흡수 앞**에서 `scripts/cutover-precheck.sh --exclude-card t1453` 을 실행한다. 이 카드 자신(D-25)은 2단계 병합 뒤 커밋(증거 hoist 등)이 있으면 브랜치 팁이 origin/develop 의 조상이 아닐 수 있어 한 장만 명시적으로 제외하며, 출력이 제외를 이름으로 적는다. 다른 picked 카드·미푸시 커밋·통합 창·슬롯·활성 세션은 제외되지 않는다. 종료 코드 0 과 `PRECHECK_CLEAR` 여야 4단계로 간다. 점검 스크립트는 origin/develop 을 fetch 하지 않고 있는 그대로 읽으므로, 최신 읽기가 필요하면 리더가 먼저 `git fetch origin develop`(읽기)을 한다 | 리더 | 아니오(읽기 전용) | 점검 출력 전문(종료 코드 포함)과 `EXCLUDED card t1453` 줄. 이 출력이 4단계 이후의 확인 기록 선행 조건이다 |
| 4 | develop 이 `origin/main` 을 흡수한다: 병합 커밋(`git merge --no-ff origin/main`), 충돌 없음(원장 E-23 이 측정한 형태이므로 이 시점에 다시 관측한다). 흡수 뒤 push | 리더 | 예(develop push) | 흡수 커밋 SHA 와 부모 둘(첫째 develop, 둘째 main), 흡수 뒤 트리 SHA, push 뒤 `git rev-parse origin/develop`. 이 SHA 가 5b 병합 직전의 develop 팁이다 |
| 5a | develop→main PR 개설(병합 방식은 병합 커밋으로 지정, squash 금지). 개설 전에 `scripts/cutover-protection-compare.sh --expect baseline` 로 보호 상태가 research.md §2 와 같음을 읽는다(읽기 전용 `gh api`). 이 PR 의 필수 체크(`pull_request: branches: [main]`, `ci.yml:20`)가 두 번째 CI 증거다. `Release PR Multi-OS Gate` 는 release/* 가 아닌 head 에서는 detect 가 건너뛰어 통과하는 경로로 읽히나(워크플로 본문을 읽은 것이고 이 PR 로 실행한 적은 없다) 이 PR 에서 실제로 보고되는지 개설 직후 확인한다. 운영자의 확인 기록을 받은 뒤에만 개설한다. 하나의 PR 이 GitHub 한도에 걸리면 `## 폴백` 으로 간다 | 리더 | 예(PR) | PR 번호, 병합 방식 지정 사실, protection-compare 출력, 필수 체크 5개의 상태 |
| 5b | 필수 체크 통과를 읽고 그 PR 을 **병합 커밋**으로 병합한다(squash·rebase 가 아니다: squash 는 develop 팁을 main 의 조상에서 끊어 이후 두 브랜치의 병합이 영구히 충돌한다). 병합 직전에 `## 되돌리기` 의 병합 전 ref 기록이 확인 기록에 있는지 본다 | 운영자 | 예(main) | 병합 커밋 SHA 와 그 부모 둘, 병합 시점의 필수 체크 5개 상태, 병합 방식(Merge commit)이 화면이나 `gh pr view` 에 보인 사실 |
| 6 | 트리 항등과 조상 확인: fetch 뒤 `git rev-parse origin/main^{tree}` 가 4단계에서 push 한 develop 팁의 `^{tree}` 와 같고, `git merge-base --is-ancestor <4단계 SHA> origin/main` 의 종료 코드가 0 이다. 하나라도 어긋나면 7단계로 가지 않고 `## 되돌리기` 로 간다 | 리더 | 아니오 | 두 트리 SHA 를 낸 명령과 출력, `merge-base --is-ancestor` 의 종료 코드. 이 출력이 7단계의 선행 조건이다 |
| 7 | 레인 재기동: 리더 세션에서 `moai cc -f N` 으로 새 세션을 띄우고, 모든 새 카드 워크트리는 `origin/main` 에서 만든다. 레인은 카드 워크트리 안에서 launcher(`moai cc -w <name>` 또는 `moai cc -w <절대경로>`)로 시작하고, 실행 중인 세션을 한 카드 트리에서 다른 카드 트리로 옮기지 않는다. 한 세션이 카드 전이를 겪으면 `/clear` 는 새 트리로 옮긴 뒤 한 번만 한다(새로 띄운 세션은 첫 카드에 `/clear` 가 필요 없다). 3단계가 병합·push 되지 않은 picked 카드 0 을 보증했으므로 이전 기준에서 만든 미착지 카드 브랜치는 없다. 첫 새 카드 트리에서 `git merge-base HEAD origin/main` 으로 그 트리가 `origin/main` 위에서 분기했음을 확인한다 | 리더 | 아니오 | 재기동한 세션 목록, 첫 새 카드 트리의 `git merge-base HEAD origin/main` 출력, 6단계 출력에 대한 참조 |
| 8 | 첫 카드 관측: 첫 카드가 PR → CI → 병합 → 착지 판정 → sweep 까지 가는지 관측한다. CI 는 PR 번호와 런 id 로 고정해서 읽는다(D-28) | 리더 | 예(PR) | PR 번호, CI 런 id 와 conclusion(`gh run view`), 병합 커밋 SHA, 착지 판정 세 층의 결과, sweep 이 로컬 트리를 정리했는지의 관측 |
| 9a | develop 퇴역의 운영자 몫(D-13): 단계 2 로 `Release PR Multi-OS Gate` 를 main 필수 체크에서 빼고(D-22), develop 보호(읽기 전용)를 건다. 단계 3 의 조건 — 열린 PR 이 develop 을 대상으로 하지 않음, `spec-lint.yml` 의 develop 가져오기 제거 확인(M5), 첫 카드 PR 이 main 까지 완주, main push CI 녹색, 옵션 B 의 `--dry-run` 관측 — 이 서면 develop 브랜치 삭제 | 운영자 | 예(main 보호·develop 브랜치) | 각 조건을 관측한 명령과 출력(조건마다), 필수 체크 목록 변경 뒤의 `scripts/cutover-protection-compare.sh --expect post-cutover` 출력(리더가 읽음), 삭제 뒤 `git ls-remote --heads origin develop` 의 빈 출력 |
| 9b | 삭제 뒤 리더의 몫(D-13 단계 4): 워크플로 push 트리거 목록의 `develop` 을 정리하는 카드(main 대상 PR)와 퇴역 뒤 잔존 점검. 점검은 `git grep -n -w develop -- .github/workflows` 의 종료 코드가 1 이고, `gh run list --workflow=CI --branch develop --limit 3` 의 가장 새 `createdAt` 이 전진하지 않는 것이다. 정리 카드의 발행과 PR 은 리더의 일이다 | 리더 | 예(PR) | 두 점검 명령의 출력과 종료 코드, 점검 시점에 읽은 가장 새 `createdAt` 값 |

## 선행 조건과 확인 기록

| 단계 | 선행 조건(관측) | 누가 기록하나 |
|---|---|---|
| 0·2 | origin/develop 팁의 CI 가 success(팁 SHA 로 읽은 출력) | 리더 |
| 4 이후 | 3단계 사전 점검의 출력(종료 코드 0) | 리더 |
| 5a | 4단계 push 뒤의 `origin/develop` 전체 SHA 와 운영자의 확인 기록 | 리더, 운영자 |
| 5b | 5a 의 PR 에서 필수 체크가 통과한 상태 | 운영자 |
| 7 | 6단계의 트리 항등·조상 출력 | 리더 |
| 9a | 8단계의 첫 카드 완주, main push CI 녹색, 옵션 B `--dry-run`, 열린 PR 의 재대상 | 운영자 |
| 9b | 9a 의 develop 브랜치 삭제 | 리더 |

확인 기록이 없는 단계는 미실행으로 읽는다. 위 표의 선행 조건은 단계 이름이나 시각이 아니라 위 열의 관측이다.

## 되돌리기

REQ-GFD-018 의 "병합 전에 되돌리는 경로를 기록한다"를 이 절이 채운다.

**병합 전 ref 기록.** 5a 개설 직전에 리더가 `git rev-parse origin/main origin/develop` 의 출력을 확인 기록에 적는다(`pre-merge-refs`: main 의 시작 SHA 와 develop 의 시작 SHA). 스크래치 리허설은 같은 형식의 `pre-merge-refs.txt` 를 병합 전에 쓰고 되돌리기를 그 파일에서 읽는다.

**되돌리기 경로는 되돌리기 PR 이다.** 실제 main 은 force-push 가 막혀 있고(`allow_force_pushes: false`, `enforce_admins: true`, research.md §2) 보호 규칙은 PR 을 요구하므로, 수렴 병합을 되돌리는 길은 main 에 병합 커밋의 revert PR(`git revert -m 1 <병합 커밋 SHA>`)을 올려 병합하는 것이다. 되돌린 뒤 main 의 트리가 기록해 둔 시작 main 트리와 같아야 한다. 이 경로는 스크래치 리허설이 되돌린 트리 단언으로 시연한다. 병합 커밋을 되돌린 뒤 같은 develop 을 다시 병합하려면 먼저 그 revert 를 되돌려야 한다는 git 의 알려진 성질이 있으므로 재시도 전에 리더가 이를 확인한다.

**ref 복원은 스크래치 전용이다.** 기록한 ref 로 브랜치를 되돌려 force-push 하는 경로는 스크래치 리허설에서만 쓴다. 실제 main 에는 보호가 막으므로 이 경로를 쓰지 않는다. 리허설이 이를 시연하는 이유는 병합 전 ref 기록이 되돌리기의 완전한 입력임을 보이기 위해서다.

**develop 은 단계 1 동안 삭제하지 않는다.** 되돌린 뒤에도 다시 시도할 원본이 남는다.

## 폴백

하나의 수렴 PR 이 GitHub 한도(7.7천 파일 규모)에 걸려 열리지 않거나 병합되지 않으면(D-18) 순서대로 다음으로 간다. 어느 폴백을 쓸지는 절체 시점의 운영자·리더 판단이고 이 카드의 레인은 고르지 않는다.

1. develop 의 조상 커밋 구간을 순서대로 main 에 병합하는 **분할 PR**. 각 구간이 병합 커밋이어야 조상 관계가 유지된다.
2. 운영자가 보호를 일시 완화하는 방법. 외부 공유 시스템의 되돌리기 어려운 변경이라 운영자가 직접 하고, 완화와 복구 사이의 main 상태를 확인 기록에 적는다.

## 레인 정지와 재기동

kanban-dispatch 규칙(`kanban-dispatch.md`)과 이 절체의 관계를 한곳에 모은다.

- **정지(1단계).** 새 배차를 멈추고, 각 레인이 진행 카드를 끝내 증거를 hoist 한 뒤 리더가 운영자에게 `/clear` 를 요청한다. `/clear` 는 사용자가 치는 명령이라 지시로 보낼 수 없다. 워크트리는 카드 브랜치가 push 되고 원격 병합이 착지하기 전에는 폐기하지 않는다.
- **재기동(7단계).** 새 카드는 새 워크트리에서 시작하고 이전 카드의 트리를 재사용하지 않는다. 카드 전이마다 `/clear` 는 한 번이며 옮긴 뒤에 한다(카드 전이마다 한 번, 그 수를 줄이지 않는다). 카드 워크트리는 `WT-<slug>` 브랜치로 개명하고 카드 id 는 디렉터리와 커밋·증거 경로에 둔다.
- **보증.** 3단계 사전 점검이 병합·push 되지 않은 picked 카드 0, 살아 있는 창·슬롯 보유자 0, 활성 레인 세션 0 을 보증하므로 재기동 시점에 이전 기준에서 만든 미착지 작업이 없다.

## develop 퇴역 단계 (D-13)

단계 순서는 시간이 아니라 조건이다. 삭제 시점은 리더 결정이다.

| 단계 | 상태 | 조건 |
|---|---|---|
| 단계 1 | 수렴과 절체 후. develop 은 남아 있고 더 이상 커밋을 받지 않는다. push 트리거의 `develop` 은 그대로라 develop 팁 CI 가 계속 관측된다 | 되돌리기 원본 유지 |
| 단계 2 | develop 을 읽기 전용으로 만들거나 보호한다(운영자). `Release PR Multi-OS Gate` 를 main 필수 체크에서 뺀다(운영자, D-22) | 열린 PR 이 develop 을 대상으로 하지 않는다. 최근 병합 PR 20건 중 17건이 develop 대상이었으므로 열린 PR 의 재대상 지정이 필요하다. 체크 제거는 M5 의 스위치 켜기가 병합되고 AC-GFD-010 의 두 상태 시험이 통과한 뒤 |
| 단계 3 | develop 삭제(운영자) | `spec-lint.yml` 의 develop 의존 제거 확인(M5), 첫 카드 PR 이 main 까지 완주, main push CI 녹색, 옵션 B 의 `--dry-run` 관측 |
| 단계 4 | 삭제 뒤 정리: 워크플로 push 트리거 목록의 `develop` 제거(리더가 정하는 정리 카드, main 대상 PR) | 잔존 점검 통과 — `git grep -n -w develop -- .github/workflows` 종료 코드 1, 연속 발화 확인 |

`spec-lint.yml` 의 `git fetch origin main:… develop:…` 는 develop 이 사라지면 잡 전체를 실패시키므로 삭제보다 **앞서** 제거한다(M5). 반면 push 트리거의 `develop` 은 삭제 **뒤** 단계 4 에서 정리한다 — 절체 내내 develop 팁 CI 를 관측하기 위해서다(D-17).

## 퇴역 뒤 잔존 점검과 연속 발화

이 점검은 verification-completeness §1.3(연속 발화)의 답이다. 푸시 트리거에서 `develop` 이 사라지면 그 이벤트를 구독하던 가드가 조용히 멈출 수 있고, 멈춘 가드는 실패가 아니라 부재로 보이기 때문이다.

1. **잔존 점검(9b).** `git grep -n -w develop -- .github/workflows` 의 종료 코드가 1 이다(0 은 develop 을 가리키는 죽은 설정이 남았다는 뜻).
2. **develop 쪽 정지.** `gh run list --workflow=CI --branch develop --limit 3` 의 가장 새 `createdAt` 이 삭제 시점 뒤로 전진하지 않는다. 이 명령은 SPEC 이 정한 형태이지만 목록 호출은 같은 명령이 다른 창을 돌려준 사례가 있다(D-28). 그래서 두 번 읽어 같은 값이 나올 때만 증거로 쓰고, 값이 서로 다르면 증거가 아니라 Gap 으로 적는다.
3. **main 쪽 발화(연속 발화).** develop push 로 돌던 가드가 main 의 병합 SHA 에서 실제로 도는지, 병합 SHA 로 고정해(`gh api "repos/modu-ai/moai-adk/actions/runs?head_sha=<SHA>"`) 가드마다 런이 있는지 읽는다. 정리 뒤에 이 가드가 멈췄다면 "아무것도 실패하지 않았고 실패할 것도 없었다"가 되므로 이 읽기가 그 신호를 묻지 않아도 오게 하는 장치다.

## 운영자 전용 단계와 이 카드가 하지 않는 일

REQ-GFD-020 에 따라 이 카드의 레인은 아래를 리더의 확인 기록 없이 실행하지 않으며, 이 문서는 그것들을 카드 단계로 두지 않는다.

- `5b` develop→main PR 병합, `9a` 필수 체크 목록에서 `Release PR Multi-OS Gate` 제거·develop 보호·develop 브랜치 삭제(운영자).
- `moai constitution amend` 두 번(`CONST-V3R5-027`, `CONST-V3R5-028`): 5단째 인간 승인이 대화형 Y/N 이라 운영자가 실행하고, 레인은 `--before`/`--after`/`--evidence` 입력과 `--dry-run` 제안을 준비하고 사후에 `moai constitution validate` 로 검증한다(D-26).
- develop push 세 번(0·2·4단계), 레인 정지와 재기동(리더).
- rc 또는 정식 태그 생성과 push. 이 카드는 태그를 만들지 않는다. 게이트는 태그 앞에 건다(D-7·D-22).
- main 보호·규칙셋·기본 브랜치·merge 설정의 변경. `scripts/cutover-protection-compare.sh` 는 이를 **읽기만** 한다.

## 카드가 준비한 것 (카드 산출물 — 실행하지 않는다)

| ID | 준비물 | 주체 | 외부 공유 시스템 여부 | 위치 |
|---|---|---|---|---|
| P1 | 이 런북 | 카드 | 아니오 | `.moai/specs/SPEC-GITHUB-FLOW-DEFAULT-001/cutover-runbook.md` |
| P2 | 배치 경계 사전 점검(읽기 전용, 읽기 명령 전부 주입 가능) | 카드 | 아니오 | `scripts/cutover-precheck.sh` |
| P3 | 수렴 리허설(스크래치 클론과 스크래치 bare origin 에서만 실행) | 카드 | 아니오 | `scripts/cutover-rehearsal.sh` |
| P4 | 보호 상태 대조(읽기 전용 `gh api`, 주입 가능) | 카드 | 아니오 | `scripts/cutover-protection-compare.sh` |
| P5 | 위 넷의 시험 | 카드 | 아니오 | `internal/template/cutover_*_test.go` |
| P6 | 리허설 증거(트리 항등·조상·되돌리기·ref 복원) | 카드 | 아니오 | `progress.md` §E.2 의 `#### M6` |

## 이 런북이 SPEC 과 만나는 모호한 자리와 못 본 것

SPEC 을 고치지 않고 아는 대로 적는다(개정 A-2 의 입력은 progress.md §J.5).

- **0단계의 CI 읽기와 9b 의 `createdAt` 점검이 서로 다른 읽기 방식을 쓴다.** D-17·D-28 은 목록 호출을 쓰지 말고 SHA 로 고정하라고 하는데, AC-GFD-023 (4) 은 `gh run list --workflow=CI --branch develop --limit 3` 를 문구로 요구한다. 0단계는 SHA 로 고정했고, 9b 는 SPEC 의 문구를 그대로 두되 두 번 읽어 일치할 때만 증거로 삼는다고 적었다.
- **3단계의 "활성 레인 세션"은 세션 등록부의 pid 가 살아 있고 cwd 가 카드 워크트리 안인 항목으로 읽었다**(`moai session list --json` 한 번 읽기). 등록부의 heartbeat 나이는 보지 않으며 pid 재사용으로 오탐할 수 있고, 등록부 밖의 세션은 못 본다.
- **AC-GFD-020 에는 이름 붙은 스크립트가 없다.** 읽기 전용 대조를 돌릴 수 있게 `scripts/cutover-protection-compare.sh` 를 더했다. 그 실제 `gh api` 응답 모양(`required_status_checks.contexts` 등)은 스텁으로만 관측했고 실제 저장소에는 돌리지 않았다.
- **D-8 의 3단계 행은 "카드 스크립트, 리더 실행"으로 두 주체를 적었다.** D-29 의 한 값 규칙에 맞춰 실행 주체인 리더로 적었다(스크립트는 위 카드 산출물 표의 P2).
- **`Release PR Multi-OS Gate` 는 5a 의 PR 에서 필수 체크다.** 워크플로 본문을 읽으면 release/* 가 아닌 head 에서는 detect 가 건너뛰고 게이트가 통과하는 경로이지만 실제 PR 로 관측하지 않았다. 보고되지 않아 병합이 막히면 폴백 2(운영자의 일시 완화)가 쓰인다.
- **수렴 병합의 PR 한도(7.7천 파일 규모)와 실제 GitHub 의 병합 커밋 방식 동작은 스크래치에서 관측할 수 없다.** 리허설은 git 기계 장치(트리 항등·조상·되돌리기)만 증명한다.
