# SPEC-VERSION-STAMP-GUARD-001 — 수락 기준

> `spec.md` §3의 요구를 기계적으로 판정 가능한 형태로 옮긴 것. 이 파일은 **검증 계층**이며
> 각 항목은 Given-When-Then이다. GEARS 요구문은 `spec.md` §3에만 있다.

문서 핀: 이 파일의 모든 RED-now 측정은 트리 **`9328a5242`**(카드 워크트리
`WT-version-sync-list` HEAD, 측정 시점 `origin/develop`과 동일)에서 실행됐다. 항목별 핀이 없는
기준은 이 문서 핀에 구속된다(`verification-completeness.md` §2.1). run-phase는 실행 시점의
HEAD를 다시 적는다.

RED-now / green path 두 셀은 `verification-completeness.md` §2의 채택 규율을 따른다 — RED-now
하나만으로는 「이 작업이 그것을 뒤집을 수 있다」가 증명되지 않고, green path 하나만으로는
출발 관측이 없는 약속이다.

---

## §D 수락 매트릭스

| ID | 요구 | 판정 | RED-now | 뒤집는 마일스톤 |
|---|---|---|---|---|
| AC-VSG-001 | REQ-VSG-001 | 목록 집합 == 리터럴 7경로 | 유령 1 + 누락 4 | M2 |
| AC-VSG-002 | REQ-VSG-002 | 스탬프/릴리스 산출물 분리, 교집합 공집합 | 소제목 1개뿐 | M2 |
| AC-VSG-003 | REQ-VSG-003 | 파생값 단언 0건 + 폴백 서술 존재 | 단언 2건 살아 있음 | M2 |
| AC-VSG-004 | REQ-VSG-004 | 존재 단언이 없는 경로에 실패하고 이름으로 지목 | 검사 부재 | M2.3 (치환 → 되돌림) |
| AC-VSG-005 | REQ-VSG-005 | 개수 단언이 기대값 7과 다를 때 실패(0건 포함) | 검사 부재 | M1 (parsed=0) → M2.2 |
| AC-VSG-006 | REQ-VSG-006 | 절반 보장 명시 + t392 지목 | 문서에 서술 부재 | M2 |

[HARD] AC-VSG-004와 AC-VSG-005는 **한 검사의 서로 다른 두 단언**이며, 각자 자기 RED을 따로
관측한다. 한쪽이 낸 빨강을 다른 쪽의 증거로 적지 않는다 — 완료의 단위는 검사 파일이 아니라
「실패가 관측된 단언」이다(`verification-completeness.md` §1.1). 어느 단언이 울었는지는
`plan.md` §D의 단언 메시지 계약 리터럴로 판정한다.

---

### AC-VSG-001 — 목록이 리터럴 7경로 집합과 일치한다

**Given** `.moai/docs/version-management.md`의 버전 스탬프 소제목 아래 경로 목록,
**When** 그 집합을 아래 **리터럴 7경로 집합**과 대조하면,
**Then** 두 집합이 정확히 같다 — 차집합이 양방향 모두 공집합이다.

    pkg/version/version.go
    .moai/config/sections/system.yaml
    README.md
    README.ko.md
    README.ja.md
    README.zh.md
    docs-site/hugo.toml

[HARD] 판정자는 이 리터럴 집합이며, 저장소 객체를 조회하지 않는다. 0.1.0/0.2.0은 v3.1.4 범프
커밋 `61921f1ba`가 바꾼 경로 집합을 판정자로 삼았는데, 그 커밋은 이 브랜치에서 **도달
불가능**하다(조상 검사 rc=1, `release/v3.1.4`에만 존재). `61921f1ba`는 이 집합의 **출처
인용**으로만 남으며 판정에 관여하지 않는다 — 전말은 `spec.md` §7 R-1.

`internal/template/templates/.moai/config/config.yaml`은 어느 소제목 아래에도 나타나지 않는다.
`system.yaml.tmpl`이 「스탬프가 아님」으로 사유와 함께 언급되는 것은 허용한다 — 목록 항목이
아니라 주석이기 때문이다(`spec.md` §1.2).

**RED-now** (트리 `9328a5242`):

    command: grep -c 'internal/template/templates/.moai/config/config.yaml' .moai/docs/version-management.md
    stdout:  1
    exit:    0

즉 유령이 목록에 살아 있다. 누락 4건도 같은 트리에서 확인된다(§1의 대조표).
**green path**: M2가 유령을 제거하고 4건을 추가하면 집합이 일치한다.

---

### AC-VSG-002 — 스탬프와 릴리스 산출물이 분리되어 있다

**Given** 수정된 `.moai/docs/version-management.md`,
**When** 「Files Requiring Version Sync」 절을 읽으면,
**Then** 두 개의 구별되는 소제목이 존재하고, 각각의 아래 항목 집합의 교집합이 공집합이다.

- 스탬프 소제목 아래: AC-VSG-001의 7경로. `CHANGELOG.md`·`.moai/release-notes/`는 **없다**.
- 릴리스 산출물 소제목 아래: `CHANGELOG.md`, `.moai/release-notes/vX.Y.Z.ko.md`. 7경로 중
  어느 것도 **없다**.
- 릴리스 산출물 소제목은 「범프 커밋이 건드리지 않는다」를 명시한다.

**RED-now** (트리 `9328a5242`): 현재 문서는 `**Documentation Files:**` / `**Configuration
Files:**` 두 굵은 라벨을 쓰지만 그 축은 **문서 대 설정**이지 **스탬프 대 산출물**이 아니다 —
`CHANGELOG.md`와 `README.md`가 같은 라벨 아래 있고, 이는 이 AC가 요구하는 분리가 아니다.
**green path**: M2가 축을 바꿔 두 소제목으로 나눈다.

---

### AC-VSG-003 — version.go 모순이 해소되어 있다

**Given** 수정된 문서의 「Single Source of Truth」 절,
**When** `version.go`를 언급하는 모든 행을 읽으면,
**Then** 아래 세 조건이 동시에 성립한다.

1. `reads from git tags at build time` / `via git describe`에 해당하는, `version.go`를
   파생값으로 서술하는 단언이 **0건**이다.
2. `Version`의 **기본값**이 ldflags 부재 시의 폴백이며 매 범프마다 손으로 갱신된다는 서술이
   **1건 이상** 존재하고, 근거로 `Makefile`의 주입 지점과 `.goreleaser.yml`을 함께 가리킨다.
3. 문서가 `Version`을 `constant` / `상수`로 부르는 곳이 **0건**이다 — 패키지 수준 `var`이며,
   `-X` 주입이 성립하려면 그래야 한다(`pkg/version/version.go:7-8`).

문서가 LDFLAGS 주입 사실 자체를 서술하는 것은 그대로 유지된다 — 틀린 것은 「따라서 기본값은
손댈 필요가 없다」는 함의이지 주입 사실이 아니다.

**RED-now** (트리 `9328a5242`): 파생값 단언 2건이 `.moai/docs/version-management.md:8`과 `:12`에
그대로 있다(`spec.md` §1.1에 verbatim 인용). **green path**: M2가 두 행을 고친다.

---

### AC-VSG-004 — 존재 단언이 없는 경로에 실패하고 그 경로를 이름으로 지목한다

**Given** 검사가 착지하고 문서 수정이 끝난 트리(M2.2 이후 — 스탬프 소제목이 존재하고 항목이
7이며 검사가 통과하는 상태),
**When** 스탬프 소제목 아래 항목 한 줄을 존재하지 않는 경로로 **치환**하고 검사를 돌리면,
**Then** 검사가 실패하고 출력이 그 경로를 **이름으로** 포함한다.
**And** 치환을 되돌리면 검사가 통과한다.

[HARD] **추가가 아니라 치환이다.** 한 줄을 더하면 파싱 개수가 8이 되어 개수 단언
(AC-VSG-005)이 함께 울고, 빨강의 원인이 둘이 된다. 치환하면 개수가 7로 유지되므로 **존재
단언 단독의 빨강**이 나오고, 이 AC가 실제로 무엇을 관측했는지가 확정된다.

[HARD] 치환 대상과 심는 경로를 **둘 다 고정한다**: `docs-site/hugo.toml` →
`docs-site/nonexistent-stamp.toml`. 「임의의 존재하지 않는 경로」로 두면, 확장자나 디렉터리를
하드코딩한 구현이 우연히 통과할 수 있다. 이 경로는 두 성질을 동시에 만족한다 — 실재하지
않고(`test -e` 종료 1, 이 트리에서 실측), 디렉터리(`docs-site/`)와 확장자(`.toml`)는 스탬프
목록에 실제로 등장해(치환 대상 자신이 그 예다) 「모르는 확장자라서 건너뛴다」는 도피로를
막는다.

[HARD] **기대 RED 출력을 측정 전에 못박는다.** 실패 출력은 `plan.md` §D의 리터럴을 포함한다:

    version-sync list names a path that does not exist: docs-site/nonexistent-stamp.toml

그리고 같은 실행에서 `expected=7`과 다른 `parsed=` 값이 **나오지 않아야** 한다 — 나온다면
치환이 아니라 추가가 된 것이고, 이 관측은 이 AC의 증거가 아니다.

[HARD] 릴리스 산출물 소제목 아래 항목은 판정 대상이 **아니다**. `.moai/release-notes/vX.Y.Z.ko.md`는
플레이스홀더이며 그 이름의 파일은 존재하지 않는다(실측: 실재 파일은 `v3.1.0.ko.md`,
`v3.1.3.ko.md`). 이 한정이 없으면 검사가 정상 항목을 유령으로 지목한다.

**RED-now** (트리 `9328a5242`) — 존재 단언을 낼 주체 자체가 없다:

    command: test -e internal/cli/version_sync_list_test.go
    stdout:  (empty)
    exit:    1

즉 오늘의 트리에서는 어떤 경로도 이름으로 지목되지 않으며, 그것이 이 AC가 뒤집을 출발
상태다.

[HARD] **오늘 트리의 유령은 이 AC의 RED-now가 아니다.** 유령
(`internal/template/templates/.moai/config/config.yaml`, 문서 78행)은 실재하지 않지만
(`test -e` 종료 1) `**Configuration Files:**` 아래에 있고, 검사의 앵커 한정은 그 구간을 읽지
않는다(`spec.md` §5.1). 유령은 **AC-VSG-001**의 RED-now이며 M2가 문서에서 지운다. 이 AC의
관측 대상은 M2.3의 치환이다.

**green path**: M2.2에서 통과 상태를 확인한 뒤, M2.3에서 치환 → 실패(위 리터럴 관측) →
되돌림 → 통과. RED → GREEN 전이가 **한 트리 안에서 문서 한 줄의 차이로** 관측되며, 그
관측이 이 AC의 증거다(`plan.md` E3-c).

**보조 관측**(증거 아님): M2.1 트리(소제목 신설·누락 보충 후 유령 미제거)에서는 존재 단언이
실제 유령을 이름으로 지목한다. 이 검사가 실제 결함을 만나는 유일한 순간이므로 기록하되,
같은 실행에서 개수 단언도 `parsed=8 expected=7`로 울므로 **원인이 둘**이고 이 AC의 단일
증거로 쓰지 않는다(`plan.md` E3-b).

**mutant probe** (`verification-completeness.md` §2): 항상 빈 결과를 반환하는 구현은 이 AC의
「실패한다」 절을 통과하지 못한다. 반대로 항상 실패하는 구현은 「되돌리면 통과한다」 절에
걸린다. 양방향이 함께 있어야 두 뮤턴트가 모두 죽는다. 세 번째 뮤턴트 — 실패는 하되 경로를
출력하지 않는 구현 — 은 위 기대 RED 리터럴의 grep이 죽인다.

---

### AC-VSG-005 — 검사가 몇 개를 훑었는지 말한다

**Given** 착지한 검사,
**When** 검사를 돌리면,
**Then** 파싱한 경로 개수를 보고하고, 그 값이 **검사가 상수로 보유한 기대 개수 7**과 다르면
실패한다.
**And** 파싱 결과가 **0건**인 경우도 그 비교에 걸려 통과가 아니라 실패로 보고된다.

[HARD] 기대 개수 7은 **파싱 결과에서 유도하지 않는다.** 파싱값끼리 비교하면 언제나 참이라
아무것도 단언하지 않으며, 0건 파싱이 통과로 새어 나가는 구멍이 그대로 남는다. 7은
AC-VSG-001의 리터럴 7경로 집합에서 온 상수이고, 문서 항목 수가 바뀌면 사람이 함께 고쳐야
한다(그 잔여 위험이 `spec.md` §7 R-5다).

이 항목이 이 카드의 비공허성 장치다. 훑은 집합이 비어 있는 통과는 아무것도 단언하지 않으며,
그 통과는 「전부 통과」와 출력이 구별되지 않는다(`verification-completeness.md` §1.1).

**RED-now** (트리 `9328a5242`): 검사 자체가 없다 — `internal/cli/version_sync_list_test.go`가
존재하지 않는다(`test -e` 종료 1). 그러므로 「0건을 훑고 통과하는」 상태가 지금의 기본값이다.

**green path**: M1이 개수 보고와 기대값 단언을 함께 세운다. **M1 착지 트리에서 이 단언이
곧바로 운다** — 앵커 소제목을 M2가 만들기 때문에 그 시점 파싱은 0건이다. 기대 RED은 측정
전에 못박혀 있다(`plan.md` §D):

    version-stamp entries: parsed=0 expected=7

이것이 이 AC의 RED 증거이며(`plan.md` E3-a), **AC-VSG-004의 증거가 아니다** — 이 실행에서
존재 단언은 어떤 경로도 보지 못했다(`spec.md` §5.1). M2.2에서 항목이 7이 되면
`parsed=7 expected=7`로 조용해지는 것이 이 AC의 GREEN이다.

**mutant probe**: 존재 검사만 있고 개수 단언이 없는 구현은, 소제목 이름이 바뀌어 파싱이 0건이
되어도 초록으로 남는다 — AC-VSG-004만으로는 그 뮤턴트가 죽지 않으므로 이 AC가 따로 필요하다.

---

### AC-VSG-006 — 절반 보장이 명시되어 있다

**Given** 수정된 `.moai/docs/version-management.md`의 회귀 보장 문단,
**When** 그 문단을 읽으면,
**Then** 아래 세 조건이 동시에 성립한다.

1. 검사가 잡는 방향(목록이 없는 경로를 가리키는 유령)과 **잡지 못하는 방향**(목록에 없는 실제
   스탬프 사이트)이 각각 명시된다.
2. 잡지 못하는 방향의 후속으로 **카드 t392**가 지목된다.
3. 그 문단이 보장의 **부분성을 양성으로 선언**하고, 아래 과다 주장 리터럴을 하나도 담지
   않는다. 둘 다 grep으로 판정한다.

[HARD] 3항의 판정 절차 — 「그런 취지의 서술이 없다」는 판단이 아니라 두 개의 명령이다.

**(a) 양성 존재** — 문단이 아래 두 리터럴을 **모두** 포함한다(각 1건 이상). 부분성 선언은
없는 것을 세는 대신 있는 것을 세어 판정한다.

    partial
    does not detect

즉 「the guarantee this check establishes is **partial**」과 「it **does not detect** a stamp
site absent from the list」에 해당하는 문장이 실제로 쓰여 있어야 한다. 문서가 영어이므로
리터럴도 영어다.

**(b) 리터럴 거부 목록** — 아래 문자열이 이 절 안에 **0건**이다.

    no longer rot
    can no longer
    fully prevent
    guarantees that the list
    ensures the list

    판정: grep -nF -e '<각 리터럴>' <문서의 해당 절> → 전부 rc=1

(a)가 (b)보다 강한 계기다 — (b)는 열거하지 못한 표현을 놓치지만, (a)는 부분성을 적지 않은
문단을 그 사실만으로 떨어뜨린다. 둘을 함께 두는 이유는 (a)를 만족시키면서 옆줄에 과다 주장을
덧붙이는 경우를 (b)가 잡기 때문이다.

**RED-now** (트리 `9328a5242`): 현재 문서에는 회귀 보장을 서술하는 문단이 아예 없다 —
`grep -cF -e partial -e 'does not detect'`를 「Files Requiring Version Sync」 절에 대고 돌리면
0건이며, 검사도 없으므로 잡는 방향/못 잡는 방향의 구분도 없다.
**green path**: M2가 문서에 이 문단을 추가하고, 위 (a)/(b) 두 명령이 각각 성립한다.

이 AC는 **문서만** 판정한다. SPEC 쪽 절반 보장 서술은 `spec.md` §4가 이미 만족하며(0.3.0에
착지), 이미 초록인 대상을 Given에 넣으면 RED-now가 기준 전체에 대한 실패 입력이 되지
못한다(`verification-completeness.md` §2). REQ-VSG-006의 SPEC 쪽 절반은 커버리지를 잃지
않도록 존재로 판정한다 — 이미 성립이며 run-phase가 뒤집을 대상이 아니다:

    grep -cF '이 카드가 착지해도 목록은 여전히 썩을 수 있다' spec.md → 1  (트리 9328a5242 실측)

문서 쪽 절반만이 이 AC의 RED → GREEN 대상이다.

---

## §E 경계 사례

- **소제목만 있고 항목이 0개** — 파싱 0건이므로 AC-VSG-005가 실패시킨다. 통과가 아니다.
- **경로가 백틱으로 감싸인 항목** — 파서가 백틱을 벗겨 읽거나, 못 읽으면 개수 불일치로
  AC-VSG-005가 실패시킨다. 조용히 건너뛰지 않는다.
- **경로 뒤 괄호 주석**(`README.md (Version line)`) — 현재 문서의 실제 서식이다. 파서는 경로
  토큰만 취하고 주석을 버려야 하며, 이 형태에서 7건이 나와야 한다.
- **`system.yaml.tmpl` 언급** — 「스탬프가 아님」 주석이지 목록 항목이 아니므로 파싱 대상에서
  제외된다. 포함하면 개수가 8이 되어 AC-VSG-005가 실패시킨다.
- **앵커 소제목이 아직 없는 트리**(M1 착지 시점) — 파싱 0건이므로 개수 단언이 실패시킨다.
  존재 단언은 이 실행에서 아무 경로도 보지 못하며, 그 빨강은 AC-VSG-004의 증거가 아니다
  (`spec.md` §5.1).
- **두 단언이 함께 실패하는 트리**(M2.1) — 비치명 보고이므로 한 실행에서 두 메시지가 모두
  나온다. 첫 실패에서 중단하는 구현은 이 경계 사례에서 존재 단언의 출력을 잃는다.

---

## §F 완료 정의

- [ ] AC-VSG-001 ~ AC-VSG-006 전부 통과, 각 항목의 명령·출력·종료 코드가 인용됨
- [ ] 단언별 RED이 각자 관측됨 — E3-a(개수, `parsed=0 expected=7`) · E3-c(존재, 치환 경로를
      이름으로) · 보조로 E3-b. 각 관측의 트리 SHA와 출력 전문이 기록되고, 어느 빨강도 다른
      단언의 증거로 겸용되지 않음
- [ ] AC-VSG-004의 RED → GREEN 전이가 M2.3에서 실제로 관측됨 (치환 → 실패 → 되돌림 → 통과)
- [ ] `go test ./internal/cli/... -run TestVersionSyncList` 통과 (착지 HEAD 기록)
- [ ] `go vet ./internal/cli/...` 통과
- [ ] `.moai/docs/version-management.md` 수정본이 AC-VSG-001/002/003/006을 동시에 만족
- [ ] 새 CI 워크플로 파일·job 추가 0건 (기존 `go test ./...`가 실행 주체)
- [ ] 새 패키지 추가 0건 (`internal/cli`에 파일 하나)
- [ ] `spec.md` §7의 잔여 위험이 run-phase 관측으로 갱신됨
