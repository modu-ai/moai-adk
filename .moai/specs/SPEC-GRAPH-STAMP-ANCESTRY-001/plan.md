# SPEC-GRAPH-STAMP-ANCESTRY-001 구현 계획

## §A. 맥락과 계획 기준

### §A.1 작업 목표

이 계획은 clean codemaps 스탬프의 **객체 존재**와 **현재 checkout `HEAD`에 대한 조상성**을 서로 다른 조건으로 판정한다. 객체는 존재하지만 조상이 아니면 freshness 비교 구간이 성립하지 않으므로 숫자 값을 만들지 않고, `VerdictAbsent`는 호환성용 운반체로만 유지하며 non-nil system error와 CLI exit 2로 닫는다. 조상인 스탬프는 기존 `described-source-diff` 계산을 계속 수행하고, 값이 40 이상이면 stale/exit 1이다.

### §A.2 관측된 기준선

- 기준 트리: `7097e6e214195c45e65cab5fa565b19ca4514c4e` (`7097e6e21`)
- tracked codemaps stamp: `f7b4919541f10b6415173b6bc9fb7192e8824450`
- `git cat-file -e <stamp>^{commit}`: exit 0
- `git merge-base --is-ancestor <stamp> origin/develop`: exit 0
- `git merge-base --is-ancestor <stamp> origin/main`: exit 1
- `moai graph check --root .`: codemaps `value=254 threshold=40 verdict=stale`, process exit 1
- 카드 설명에 있던 값 224는 이 기준 트리에서 재현되지 않았다. 따라서 고정 인수값이나 예상 출력으로 사용하지 않고, 과거 회귀 관측으로만 취급한다.
- 재시도 기준 트리: `b4626c0421d3b21564fcf7ba08967889af52a278` (`b4626c042`). tracked stamp는 동일한 `f7b4919541f10b6415173b6bc9fb7192e8824450`, `dirty=false`, roots=`internal,cmd,pkg`다.
- 재시도에서 `git cat-file -e <stamp>^{commit}`, `git merge-base --is-ancestor <stamp> HEAD`, `... origin/develop`은 모두 무출력 exit 0이고, `... origin/main`은 무출력 exit 1이었다.
- 재시도에서 `moai graph check --root .`은 codemaps `value=254 threshold=40 verdict=stale`를 유지했지만 citations도 `value=1 verdict=stale`였고 전체 exit 1이었다. 따라서 M4의 종결은 codemaps만 보지 않고 다른 어떤 층도 stale이 아님을 확인해야 한다; 이 plan 단계는 전체-green을 주장하지 않는다.

### §A.3 변경 대상

| 구분 | 경로 | 계획 |
|---|---|---|
| checker | `internal/graph/check.go` | clean stamp의 조상성을 content-anchor 해석과 freshness 계산보다 먼저 판정하고, 비조상 오류에서 측정 필드를 비운다. |
| checker tests | `internal/graph/*_test.go` | 객체-비조상, merge/squash/rebase-like 계보, reachable stale/fresh, 기존 anchor/fingerprint 계약을 검증한다. |
| CLI | `internal/cli/graph_check.go` | system error 출력에 `unreachable stamp`, `freshness unmeasured`, genuine regeneration + reachable stamp 복구를 식별 가능하게 싣고 숫자 row를 만들지 않는다. |
| CLI tests | `internal/cli/graph_check*_test.go` | exit 2와 복구 문구, 숫자 출력 부재를 고정한다. |
| workflow | `.github/workflows/graph-freshness.yml` | push의 early-success 경로를 제거하고 target을 `HEAD`로 정해 동일 조상성 검사를 수행한다. ordinary PR와 `release/*` 대상 선택은 보존한다. |
| codemaps | `.moai/project/codemaps/**` | 코드·워크플로 변경을 실제로 반영해 본문을 재생성하고, 도달 가능한 커밋으로 스탬핑한다. |
| evidence | `.moai/reports/t688/**`, 이 SPEC의 `progress.md` | RED/GREEN 명령, 출력, 종료코드, 트리 SHA와 감사 결과를 보존한다. |

### §A.4 반드시 보존할 대상

- provenance schema와 `internal/mx/provenance.go`의 기존 clean/dirty 스탬프 형식
- `VerdictFresh` / `VerdictStale` / `VerdictAbsent` enum과 CLI 0/1/2 종료 구조
- `described-source-diff`, threshold 40, described-worthy predicate
- dirty fingerprint, content-anchor, contribution, driving-path 계약
- body-absent C1의 `VerdictAbsent` + nil error / exit 1 계약
- ordinary PR의 `origin/<base_ref>` 및 `release/*`의 merge-preview `HEAD` 선택
- mx-index, edges, citations 층의 판정

## §B. 네 조사 렌즈의 합성 및 모순 해소

이 절이 중단된 synthesizer를 대신하는 계획 단계 합성 기록이다. plan-auditor는 각 행의 **근거 위치**, **채택 결정**, **반증 조건**을 모두 확인해야 한다. 하나라도 맞지 않으면 이 계획을 PASS로 판정하지 않는다.

| 조사 렌즈 | 확인한 생산자·소비자 또는 계약 | 합성 결정 | plan-auditor 반증 조건 |
|---|---|---|---|
| 생산자/소비자 경로 | 스탬프 생산은 `internal/cli/graph_stamp.go` → `internal/mx/provenance.go::StampCodemaps`; 본문 재생성은 `.claude/skills/moai/workflows/codemaps.md`의 5문서; 소비는 `internal/graph/check.go::checkCodemaps` → `internal/cli/graph_check.go` → `.github/workflows/graph-freshness.yml`이다. | 조상성 선판정은 clean stamp 소비 경계에 둔다. 생산자나 schema를 바꾸지 않는다. 복구는 본문 재생성과 reachable stamping을 둘 다 요구한다. | 계획이 stamp writer/schema 변경을 요구하거나, bare restamp만으로 복구한다고 쓰거나, checker와 CLI/CI 소비 경계를 하나라도 누락하면 모순이다. |
| release-merge 조상성 | ordinary PR는 squash 생존성을 위해 `origin/<base_ref>`, `release/*`는 merge-preview `HEAD`, push는 현재 checkout `HEAD`를 판정해야 한다. 현재 push 분기는 객체 존재 뒤 `exit 0`한다. | 기존 PR 분기를 그대로 두고 push만 `TARGET=HEAD`로 동일 ancestry gate에 합류시킨다. merge-commit은 원 스탬프 조상성을 유지하고 squash/rebase-like는 객체가 남아도 비조상으로 거부하는 fixture를 둔다. | release PR를 base ancestry로 되돌리거나 ordinary PR를 `HEAD`로 완화하거나 push가 객체 존재만 확인하면 모순이다. |
| 관련 SPEC 계약 | `SPEC-V3R6-GRAPH-FRESHNESS-001/002`의 0/1/2, `SPEC-GRAPH-FRESHNESS-CADENCE-001`의 predicate/40/귀속, `SPEC-STAMP-REACHABILITY-001`의 object+ancestry/explicit stamp, `SPEC-GRAPH-GATE-RESTAMP-001`의 content anchor/bare-restamp 방지, `SPEC-CODEMAPS-REFRESH-002`의 genuine regeneration을 계승한다. 모두 현재 `completed`다. | 비조상은 숫자 stale이 아니라 미측정 system error다. 구현 호환성을 위해 report row의 `VerdictAbsent`는 유지하되 reason + non-nil error가 실제 본문 부재와 구분한다. reachable stale은 숫자 stale/exit 1을 유지한다. | `VerdictAbsent`를 본문 부재의 증거로 단정하거나, 비조상을 stale/exit 1로 바꾸거나, reachable stale을 exit 2로 바꾸거나, 새 verdict/schema를 만들면 모순이다. |
| 회귀 위험 | 기존 테스트가 bare restamp, genuine regeneration, body absent, dirty fingerprint, reverted churn, described roots, threshold 경계, contribution/driving paths를 잠근다. workflow에는 push early exit가 남아 있다. | 신규 negative/positive 계보 테스트를 기존 회귀 묶음과 함께 실행한다. 미측정 경로는 숫자·anchor·contribution·driving paths가 모두 비어야 하며, genuine regeneration 뒤 별도 freshness 측정을 수행한다. | 신규 테스트만 통과하고 기존 회귀를 실행하지 않거나, 빈 테스트 선택을 PASS로 읽거나, 본문 재생성 없이 provenance만 바꾸면 모순이다. |

### §B.1 교차 렌즈 결론

네 렌즈 사이의 충돌은 다음 구분으로 해소된다.

1. **생산 가능성과 비교 가능성은 별개다.** 스탬프 producer가 유효 commit 객체를 썼다는 사실은 consumer의 현재 `HEAD`에서 비교 가능한 조상임을 보장하지 않는다.
2. **조상성은 freshness보다 먼저다.** release/merge 위상 렌즈가 비교 구간을 정하고, 그 구간이 성립한 경우에만 기존 freshness 계약이 숫자를 낸다.
3. **`VerdictAbsent`는 상태 의미가 아니라 호환 운반체다.** reason과 non-nil error가 본문 부재(C1, nil error)와 비조상(미측정, system error)을 구분한다.
4. **복구는 두 동작의 결합이다.** 관련 SPEC의 bare-restamp 방지와 codemaps-refresh 계약을 함께 만족하려면 genuine regeneration과 reachable stamping을 모두 수행해야 한다.

## §C. 실행 전 점검

run 시작 시 manager-develop은 아래를 다시 측정한다. 계획 단계 수치는 이동하지 않는 기준 SHA에 귀속된 관측이며, run 시점 측정이 실행 판단을 대신한다.

1. `git rev-parse HEAD`와 `git branch --show-current`
2. tracked `provenance.json`의 `commit_sha`, 객체 존재, `HEAD`/`origin/develop`/`origin/main` 조상성
3. `moai graph check --root .`의 codemaps 행과 종료코드
4. 영향 패키지 테스트 선택이 실제 테스트를 1개 이상 실행하는지
5. `.github/workflows/graph-freshness.yml`의 ordinary PR / release PR / push 대상 선택
6. 관련 여섯 SPEC의 `status: completed`와 위 표의 계승 계약

## §D. 기술 접근과 제약

### §D.1 checker 판정 순서

clean `commit_sha` 경로에서는 아래 순서를 지킨다.

1. commit 객체 해석 가능성
2. stamped commit이 checkout `HEAD`의 조상인지 여부
3. content-anchor 해석
4. `described-source-diff`와 귀속 계산
5. threshold 40 비교

1 또는 2에서 실패하면 그 뒤 계산은 실행하지 않는다. 비조상 report는 `Layer=codemaps`, metric 식별자와 호환용 `VerdictAbsent`, 미측정 reason만 운반한다. `Value`, `ContentAnchor`, `ContentAnchorSource`, `Contribution`, `ContributionBase`, `DrivingPaths`, `DrivingPathsOmitted`는 측정 결과로 제시하지 않는다.

### §D.2 CLI와 workflow 경계

- CLI는 checker의 non-nil error를 exit 2로 변환한다.
- 비조상 오류 문구는 `unreachable`과 `freshness unmeasured`를 식별 가능하게 담고, recovery에는 `regenerate`와 reachable `stamp`가 모두 나타난다.
- system-error 경로에서는 일반 per-layer 숫자 표를 렌더링하지 않는다.
- workflow push는 `TARGET=HEAD`로 조상성 검사를 수행한다.
- ordinary PR는 `origin/${GITHUB_BASE_REF}`, `release/*` PR는 merge-preview `HEAD`를 유지한다.

### §D.3 TDD와 증거

프로젝트 설정은 `constitution.development_mode: tdd`다. 신규 동작은 기존 RED 스크립트와 계보 표 RED를 먼저 영구 테스트로 편입한 뒤 GREEN으로 전환한다. 각 release-blocking AC는 명령, 원문 출력, 종료코드, 트리 SHA를 원장에 남긴다. selector가 0개 테스트를 실행한 경우는 PASS가 아니다.

### §D.4 금지 사항

- 새 dependency, verdict enum, provenance field를 추가하지 않는다.
- 자동 restamp 또는 본문 없는 bare restamp를 추가하거나 권하지 않는다.
- threshold/predicate/기존 귀속 계약을 바꾸지 않는다.
- 실제 원격 ref, release merge, PR, push, issue #1661을 변경하지 않는다.
- `.moai/state/**`, 다른 SPEC, 관련 없는 파일을 수정하지 않는다.
- 로컬에서 `go test ./...`를 실행하지 않는다.

## §E. 자체 검증 계획

| 검증 항목 | 범위 | 통과 조건 |
|---|---|---|
| 신규 checker RED→GREEN | `internal/graph` | 비조상 fixture가 non-nil error와 호환용 `VerdictAbsent`를 반환하고 측정 필드는 비어 있다. |
| 신규 CLI RED→GREEN | `internal/cli` | exit 2, 복구 문구 4종, 숫자 freshness row 부재가 함께 성립한다. |
| 계보 표 | `internal/graph` | merge=ancestor/정상 비교, squash 및 rebase-like=object exists + non-ancestor + 미측정 오류다. |
| workflow push guard | `.github/workflows/graph-freshness.yml` | push 분기에 `TARGET="HEAD"`가 있고 ancestry 전에 `exit 0`이 없다. |
| PR 대상 보존 | workflow fixture/static test | ordinary=`origin/<base_ref>`, release=`HEAD`가 유지된다. |
| 기존 회귀 | 영향 패키지의 이름 지정 테스트와 전체 패키지 | bare restamp, genuine regeneration, dirty fingerprint, body absent, threshold 경계, 귀속 검사가 모두 실행되고 통과한다. |
| genuine regeneration | `.moai/project/codemaps/**` | 생성기 5문서가 현재 트리를 반영하고, reachable stamp + `described-source-diff < 40`을 독립 명령으로 관측한다. |
| SPEC 품질 | SPEC 디렉터리 | strict lint, Tier M artifact set, REQ↔AC 전수 추적, Out of Scope 규칙이 통과한다. |

## §F. 마일스톤

결정이 뒤집힐 가능성이 큰 순서대로 배치한다.

### M1 — 조상성 의미와 계보 fixture 고정 (Priority High)

- 객체 존재와 `HEAD` 조상성을 분리하는 테스트를 영구화한다.
- merge, squash, rebase-like rewritten fixture의 object/ancestry/판정 표를 만든다.
- 비조상에서 호환용 `VerdictAbsent` + non-nil error + 측정 필드 부재를 RED→GREEN으로 전환한다.

### M2 — CLI 및 workflow 경계 종결 (Priority High)

- CLI exit 2, 미측정 의미, genuine regeneration + reachable stamp 복구 안내를 고정한다.
- push `HEAD` ancestry guard를 닫는다.
- ordinary PR 및 `release/*` target 선택 회귀 검사를 함께 실행한다.

### M3 — 기존 freshness 계약 회귀 잠금 (Priority High)

- reachable stale/fresh 숫자 경계와 exit 1/0을 검증한다.
- bare restamp, dirty fingerprint, content-anchor, described-worthy, threshold, contribution/driving-path 테스트를 실행한다.
- 신규 경로가 다른 graph layer나 body-absent C1 처분을 바꾸지 않았음을 확인한다.

### M4 — 실제 codemaps 재생성 및 reachable stamp (Priority High)

- `/moai codemaps --force`의 생성기 계약에 따라 5문서를 현재 변경에 맞게 실제 재생성한다.
- 본문 변경을 검토하고, checkout에서 도달 가능한 commit으로 스탬핑한다.
- ancestry 성공과 `described-source-diff < 40`을 서로 다른 검사로 관측한다. bare restamp만으로는 M4를 완료할 수 없다.

### M5 — 범위 검증과 run 증거 수출 (Priority Medium)

- 영향 패키지 test/vet/lint와 workflow 정적 검사를 실행한다.
- AC별 command/stdout/exit/tree SHA를 `progress.md §E.2/§E.3`과 카드 증거에 기록한다.
- 관련 없는 변경과 `.moai/state/**` 변경이 없음을 확인한다.

## §G. 피해야 할 구현

- `git cat-file` 성공을 ancestry 성공으로 취급하기
- 비조상에서 `git diff`를 실행해 우연히 나온 값을 stale로 보고하기
- `VerdictAbsent` 문자열만 보고 실제 body absence와 system error를 합치기
- push에서 object presence 뒤 성공 종료하기
- release PR를 ordinary PR의 base ancestry 규칙으로 되돌리기
- 오류를 없애기 위해 threshold를 높이거나 predicate를 좁히기
- codemaps 본문을 바꾸지 않은 채 provenance만 다시 쓰기
- selector가 아무 테스트도 실행하지 않은 결과를 green으로 기록하기

## §H. 교차 참조

- `spec.md` — REQ-GSA-001~012, 판정 순서, 제외 범위
- `acceptance.md` — AC-GSA-001~012와 RED-now/GREEN-path 원장
- `.moai/reports/t688/red-ancestry.sh`
- `.moai/reports/t688/red-cli-unreachable.sh`
- `.moai/reports/t688/red-push-guard.py`
- `.moai/reports/t688/red-history-topologies.sh`
- `SPEC-V3R6-GRAPH-FRESHNESS-001`, `SPEC-V3R6-GRAPH-FRESHNESS-002`
- `SPEC-GRAPH-FRESHNESS-CADENCE-001`, `SPEC-STAMP-REACHABILITY-001`
- `SPEC-GRAPH-GATE-RESTAMP-001`, `SPEC-CODEMAPS-REFRESH-002`
- `.claude/rules/moai/development/verification-completeness.md §2, §2.1, §3`
