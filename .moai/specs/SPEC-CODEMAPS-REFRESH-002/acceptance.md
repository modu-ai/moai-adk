---
id: SPEC-CODEMAPS-REFRESH-002
title: "acceptance.md — codemaps 재발 종결 인정 기준"
version: "0.1.4"
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
tier: M
---

# acceptance.md — SPEC-CODEMAPS-REFRESH-002

## §A. 검증 원칙

합격 판정은 두 층이 **독립적으로** 성립해야 한다.

1. **게이트 층** — `moai graph check`의 codemaps `verdict=fresh`. 이것은 "문서가 최근 트리와 일치하는 시점에 생성됐다"는 최소 신호일 뿐이다.
2. **정확성 층** — 재생성 결과를 실제 트리와 대조한 증거 파일. 게이트가 advisory이고 스탬프는 기계적으로 갱신 가능하므로, **정확성 층 없는 게이트 녹색은 합격이 아니다**(REQ-CM2-011 — 카드의 중심 요건).

두 층이 왜 독립이어야 하는지는 게이트 내부 동작이 말해준다: 재스탬프만으로 `value=0`이 나오므로 **재생성을 건너뛰어도 AC-CM2-010은 통과한다.** AC-CM2-002~008이 그 구멍을 막는 유일한 층이며, 그중 하나라도 약해지면 카드 전체가 "스탬프만 갱신"으로 퇴화한다 — 카드가 명시적으로 금지한 실패 형태다.

모든 AC는 명령 + 기대 출력으로 이분 판정된다. 증거는 `.moai/reports/t475/`(세션 종료 후에도 살아 있는 경로)에 수출한다. 중간 계산의 `/tmp` 사용은 무방하나 **판정 근거로 인용되는 경로는 언제나 `.moai/reports/t475/` 아래여야 한다.**

## §B. RED-now 원장

**문서 수준 트리 핀: `52f863f36`** — 자기 핀이 없는 모든 셀에 이 SHA가 적용된다(`verification-completeness.md` §2.1). 각 셀은 **명령 / verbatim stdout / exit code / 트리 SHA** 네 요소를 담으며, 명령은 단일 호출 형식이다(파이프·리다이렉트·`&&`·`;` 없음 — 그래서 exit code가 별도 필드다).

### RED-1 → AC-CM2-010 (게이트 적색)

- command: `./bin/moai graph check`
- stdout(첫 행): `codemaps  metric=described-source-diff value=64 threshold=40 verdict=stale`
- exit: `1`
- **왜 붉은가**: provenance 앵커 `25a3212a9` 이후 described roots에서 64개 described-worthy 파일이 변경됐고, 그 수가 임계 40을 넘기 때문이다. 문서 부정확이나 파일 부재 때문이 아니다.
- green path: M4 재스탬프가 앵커를 옮겨 값을 0으로 만든다. 통과 출력은 `codemaps ... value=0 threshold=40 verdict=fresh`.

### RED-2 → AC-CM2-004 (누락 단위 부재) · 대표 원소 `internal/template/agentemit`

- command: `/usr/bin/grep -rn -F "internal/template/agentemit" .moai/project/codemaps/`
- stdout: (없음)
- exit: `1`
- **왜 붉은가**: 해당 패키지가 트리에 실재하고 `go list`가 열거하는데(`packages-absent-from-codemaps.txt`) 6문서 어디에도 인용이 없기 때문이다. 오탈자나 경로 오류가 아니다.
- **대표 원소를 `agentemit`으로 바꾼 이유**: 0.1.0의 RED-2는 `internal/stateanchor`였고, 그것도 붉었지만 손 열거된 6개 안의 원소였다. iter-1 D1이 명명한 `agentemit`은 그 6개 **밖**에 있었으므로, 후보 규칙이 실제로 넓어졌음을 이 셀이 증거한다.
- green path: M1이 omission으로 판정하면 M2가 편입한다. 통과 출력은 같은 명령이 1행 이상, exit `0`.
- **조건부성**: M1이 fold로 판정하면 이 AC는 해당 원소에 적용되지 않고 AC-CM2-002가 그 판정을 담당한다. 판정 결과를 미리 못박지 않기 위해 의도한 것이다.

### RED-3 → AC-CM2-005 (스탬프 앵커 드리프트)

- command: `cat .moai/reports/t475/described-roots-diff-since-anchor.txt`
- stdout: **수출 파일의 200행 전문**(`M internal/cli/binary_lag_test.go` … 로 시작). 파생값 요약 = `78 A / 1 D / 121 M`.
- exit: `0`
- **왜 붉은가**: 앵커 이후 described roots가 200개 경로만큼 움직였고, 그중 상위 7구간의 서술이 현재 트리를 반영하지 않기 때문이다.
- **집계는 파생값이다** — verbatim stdout은 위 파일이며, `78/1/121`은 그것을 센 결과다. 0.1.0의 RED-3은 집계만 적고 파일 경로를 인용하지 않았다(iter-1 D9-b).
- green path: M2.4가 구간마다 `diff -u`를 남긴다. 통과 판정은 그 diff의 **존재와 내용**이지 이 수치가 아니다.

### RED-4 → AC-CM2-002 · AC-CM2-006 · AC-CM2-007 · AC-CM2-008 (증거 파일 부재)

- command: `ls .moai/reports/t475/codemaps-accuracy-verification.md`
- stdout: `ls: .moai/reports/t475/codemaps-accuracy-verification.md: No such file or directory`
- exit: `1`
- **왜 붉은가**: 네 AC 전부가 이 파일의 섹션을 판정 대상으로 삼는데 파일 자체가 없다. 내용이 틀려서가 아니라 산출물이 아직 없기 때문이다.
- green path: M1/M3이 7섹션 파일을 수출한다. 통과 출력은 파일 경로 1행, exit `0`. **파일 존재는 필요조건일 뿐** — 각 AC의 내용 조건이 별도로 판정된다.

### RED-5 → AC-CM2-005 (재생성 전 사본 부재)

- command: `ls .moai/reports/t475/pre-regen`
- stdout: `ls: .moai/reports/t475/pre-regen: No such file or directory`
- exit: `1`
- **왜 붉은가**: M2.0이 아직 실행되지 않았다. 이 RED는 **복구 불가 방향**이라는 점이 특별하다 — 재생성이 먼저 실행되면 이 디렉터리는 영원히 올바른 내용으로 채워질 수 없다.
- green path: M2.0이 6개 파일을 복사한다. 통과 출력은 6개 파일명, exit `0`.

### RED-6 → AC-CM2-012 (관측 리포트 부재)

- command: `ls .moai/reports/t475/verdict.md`
- stdout: `ls: .moai/reports/t475/verdict.md: No such file or directory`
- exit: `1`
- **왜 붉은가**: M5가 아직 실행되지 않았다.
- green path: M5가 관측 3항목을 수출한다. exit `0`.

### RED-7 → AC-CM2-003a (`docs-truth.md`가 생성기 산출물이 아님)

- command: `/usr/bin/grep -n "docs-truth" .claude/skills/moai/workflows/codemaps.md`
- stdout: (없음)
- exit: `1`
- **왜 붉은가**: 스킬이 `docs-truth`를 한 번도 언급하지 않기 때문이다 — 즉 `/moai codemaps --force`가 그 파일을 만들지 않는다. 이것은 **트리의 사실**이지 이 SPEC이 만든 결함이 아니다.
- **이 셀은 regression-guard다, release-blocking이 아니다.** 작업이 이 명령의 출력을 바꾸지 않는다(스킬 파일은 소관 밖). AC-CM2-003a의 release-blocking 술어는 아래 셀이다.

### RED-7b → AC-CM2-003a (손 갱신 증거 섹션 부재) — release-blocking 술어

- command: `ls .moai/reports/t475/codemaps-accuracy-verification.md`
- stdout / exit: RED-4와 동일(`No such file or directory` / `1`)
- green path: M2.2가 `docs-truth.md`를 손 갱신하고 그 사실을 증거 파일 §③에 **M2.1 증거와 분리해** 기록한다.

### §B.1 재실행 불가 항목의 처분 (regression-guard 분류)

`verification-completeness.md` §2.1의 undecidable disposition을 적용한다. 아래는 **release-blocking 자격을 갖지 않고 regression-guard로 분류하며, 통과로 기록하지 않는다.**

| 항목 | 분류 이유 |
|---|---|
| `go list` 136 / 히트-0 48 / A층 5 / B층 14 / 잔여 42 | 트리와 함께 움직인다. 작업 후 같은 값으로 재현되리라 기대할 수 없다 — AC-CM2-002/007은 **수치가 아니라 전수 판정·분류의 존재**로 판정한다 |
| RED-7(스킬의 `docs-truth` 0회 언급) | 이 작업이 스킬 파일을 바꾸지 않으므로 출력이 영원히 같다 — 붉은 채로 남는 것이 정상이다 |
| `citations` 계층 `value=0 verdict=fresh` | **현재 이미 녹색이다.** AC-CM2-006의 RED가 아니라 재생성이 이 값을 깨지 않았는지 보는 회귀 가드다. RED는 RED-4(증거 파일 부재)가 담당한다 |

## §D. AC Matrix

### AC-CM2-001 — 기준선 재측정 기록 (MUST)

- **Given** run이 시작됐고
- **When** plan.md §C Pre-flight 배치가 실행되면
- **Then** progress.md §E.2에 각 명령과 그 verbatim 출력, 그리고 측정한 HEAD SHA가 기록돼 있다. spec.md §A의 저작 시점 값과 다르면 그 차이가 명시돼 있다. 명령 없이 수치만 적힌 기록은 FAIL.
- RED-now: progress.md §E.2가 현재 `_<pending run-phase>_` placeholder다(`52f863f36`). green path: M1 시작 시 채워진다.

### AC-CM2-002 — 후보 산출 + 전수 판별 (MUST)

- **Given** spec.md §A.3(a)의 A층·B층 명령 + C층 이월이 후보 집합을 산출하고(저작 시점 실측 5 + 14 + 1 = **20**)
- **When** M1이 §A.3(a1) 판별식으로 각 후보를 fold / omission으로 판정하면
- **Then** 증거 파일 §①에 (a) 산출된 후보 집합 전문과 (b) **그 집합의 모든 원소**에 대한 판정 행(단위 → 판정 → 근거)이 존재한다.
- **이분 판정 — 하나라도 걸리면 FAIL**:
  1. 후보 집합이 명령의 출력이 아니라 손으로 열거됐으면 FAIL(iter-1 D1 재발).
  2. **판정 행 수 ≠ 후보 수이면 FAIL.** 두 수를 각각 명령으로 세어 대조한다(신규 툴링 없음 — §B.2):

     ```bash
     wc -l < .moai/reports/t475/candidates.txt
     ```
     ```bash
     /usr/bin/grep -cE '^\| `[^`]+` \| (fold|omission) \|' .moai/reports/t475/codemaps-accuracy-verification.md
     ```

     첫 명령은 M1.0이 수출한 후보 목록(P·F 합집합, 한 줄에 하나)의 줄 수, 둘째는 증거 파일 §①의 판정 표에서 `fold`/`omission` 판정을 실은 행 수다. **두 수가 같아야 하며, 적은 쪽이 통과하는 경로는 없다** — 이것이 D1의 공허한 통과가 다시 열리는 문이다.
     판정 표의 행 형식은 위 정규식이 세도록 `| \`<단위>\` | fold\|omission | <근거> |` 로 고정한다.
     저작 시점 기대값은 20이나 **20은 목표가 아니라 그날의 후보 수다** — 실행 시점 후보 수가 다르면 그 수가 기대값이 된다. 기준은 언제나 "같은가"이지 "20인가"가 아니다.
  3. 근거가 **히트 수만** 인용한 행이 있으면 FAIL.
  4. 근거가 **변경 파일 수만** 인용한 행이 있으면 FAIL.
  5. 근거가 그 단위의 **책임을 명명하지 않은** 행이 있으면 FAIL — 3·4와 달리 이것은 양적 지표를 안 썼더라도 걸린다.
  6. 근거가 **검사한 부모 산문의 위치를 인용하지 않은** 행이 있으면 FAIL. `fold`는 책임을 담는 줄을 인용(`<문서>.md:L<n>` + 인용문), `omission`은 검사한 부모 적중 행의 범위를 명시. (같은 부모 산문을 공유하는 묶음은 인용 1회를 여러 행이 공유할 수 있다.)
  7. `internal/template/commandemit` 행이 fold로 판정돼 있으면 FAIL — 운영자가 omission으로 명명한 사례다.
  8. 근거가 **"앵커 이후 무변경"** 만을 인용한 행이 있으면 FAIL — 조건 4(변경 파일 수)의 다른 표현이다. A층 필터는 후보를 한정할 뿐 분류하지 않는다 — 후보에 들어온 이상 판정을 면제받지 않는다.
- 조건 6이 막는 뮤턴트: "`internal/chain`: 체인 노드 구성 책임. 부모 `internal`의 서술이 이를 담지 않음 → omission" — 패키지 이름에서 유도한 책임 + 부모 산문을 한 줄도 읽지 않은 부정. `verification-claim-integrity.md` §1.1 surface 4가 금지하는 형태다.
- RED-now: RED-4.

### AC-CM2-003 — 재생성 완전성 (생성기 5문서) (MUST)

- **Given** M2.0의 사본이 확보된 상태에서
- **When** `/moai codemaps --force`가 완료되면
- **Then** `overview.md` / `modules.md` / `dependencies.md` / `entry-points.md` / `data-flow.md` **5개**가 재생성 대상으로 보고되고, `ls .moai/project/codemaps/`가 7항목(문서 6 + provenance.json)을 낸다.
- **`docs-truth.md`는 이 AC의 대상이 아니다** — AC-CM2-003a가 담당한다. 0.1.0은 "6문서 전부 재생성"이라 적었으나 명시된 실행면이 그것을 만족시킬 수 없었고, `ls` 증거가 `docs-truth.md`를 손대지 않은 채 통과시켰다(iter-1 D3, REFRESH-001에서 계승).
- RED-now: `52f863f36`에서 `provenance.json`의 `generated_at`이 `2026-09-03T18:18:34Z`이고 `commit_sha`가 `25a3212a9`다 — 즉 현재 6문서는 이 run이 만든 것이 아니다. green path: 재생성 후 `generated_at`이 run 시점으로 갱신된다.

### AC-CM2-003a — `docs-truth.md` 손 갱신 (MUST)

- **Given** `docs-truth.md`가 생성기 산출물이 아니고(RED-7)
- **When** M2.2가 그것을 손으로 갱신하면
- **Then** 증거 파일 §③에 (a) 갱신 사실과 (b) **§1 에이전트 카탈로그 표 전수** 대 `.claude/agents/` 트리 나열의 대조 결과가 기록돼 있다. 표본 추출이면 FAIL.
- **§③이 §②(M2.1 재생성 증거)와 분리돼 있어야 한다.** 합쳐지면 "재생성했으니 docs-truth도 됐다"는 오독이 그대로 통과한다 — 그것이 계승된 결함의 형태였다.
- RED-now: RED-7b.

### AC-CM2-004 — omission 단위 편입 (MUST)

- **Given** M1이 어떤 후보를 omission으로 판정했고
- **When** 재생성(+ 필요 시 직접 보정) 후 `/usr/bin/grep -rl -F "<단위>" .moai/project/codemaps/`를 돌리면
- **Then** omission 판정 단위 **전부**가 1개 이상의 파일을 반환한다(exit 0). 직접 보정이 필요했던 단위는 그 사실이 증거 파일에 기록돼 있다. 하나라도 0이면 FAIL.
- **[HARD] "기록 전용"으로 흘려보낼 수 없다.** AC-CM2-007의 기록 전용 분류는 **fold 판정 후보에만** 적용된다. omission 판정 단위가 편입되지 않은 채 AC-CM2-007에 기록만 되고 12개 AC가 전부 통과하는 경로 — iter-1 D1이 명명한 공허한 통과 — 는 이 조항으로 닫힌다.
- fold로 판정된 후보는 이 AC의 대상이 아니다.
- RED-now: RED-2.

### AC-CM2-005 — 구간별 전후 대조 (MUST)

- **Given** M2.0의 사본이 `.moai/reports/t475/pre-regen/`에 존재하고
- **When** §A.3(b) 컷오프 명령이 산출한 구간마다 `diff -u .moai/reports/t475/pre-regen/<doc>.md .moai/project/codemaps/<doc>.md`를 실행하면
- **Then** 증거 파일 §④에 **컷오프가 산출한 모든 구간**에 대한 행이 존재하고, 각 행이 그 `diff -u`의 출력을 담는다.
- **"변경 없음" 행은 빈 diff 출력을 증거로 첨부해야 한다.** 0.1.0은 "바뀌지 않았고 그것이 옳은 이유"라는 자유 서술을 허용했고, 그것은 명령으로 반증되지 않았다 — 재생성이 아무것도 하지 않았어도 전 행에 "변경 없음"을 적으면 통과하는 뮤턴트가 있었다(iter-1 D7). 지금은 빈 diff가 그 주장의 증거다.
- 구간 목록이 손으로 열거됐으면 FAIL — 컷오프 명령의 출력을 채택한다. 저작 시점 실측은 7구간이며 `internal/settings`(6)를 포함한다(0.1.0은 동률인 이 구간을 빠뜨렸다 — iter-1 D5).
- **재생성 전 사본이 없으면 판정 불가이며 FAIL로 처리한다**(사후 복구 불가).
- RED-now: RED-5(사본 부재) + RED-3(드리프트 실재).

### AC-CM2-006 — 인용 경로 실존 (MUST, accuracy a)

- **Given** 재생성된 문서가 `(internal|pkg|cmd)/` 경로를 인용하고
- **When** `check_citations.go`의 정본 규약 3요소 — 정규식 `:23`, 후행 구두점 절삭 `:35`, **blockquote 면제** — 로 경로를 추출해 각각의 존재를 검사하면
- **Then** 증거 파일 §⑤에 (경로 → exists/absent) 전수 표가 존재한다. 행 수는 중복 제거한 유니크 경로 수와 같다. absent 경로는 전부 new-findings로 분류돼 있다.
- **추출 규약이 명명돼 있지 않으면 FAIL.** 0.1.0은 "경로를 추출"이라고만 적었고, 느슨한 추출로 짧은 표를 만들어도 판별할 근거가 없었다(iter-1 D6).
- **blockquote 면제를 적용하지 않았으면 FAIL.** 일부러 부존재를 인용한 줄(제거 기록·rename 이력)이 전부 거짓 `absent`로 분류된다.
- **교차 대조**: `./bin/moai graph check --json`의 `citations` 계층(`positive-cited-path-absence`, threshold 0)이 같은 규약으로 같은 측정을 기계 수행한다. 그 값과 표의 absent 건수가 어긋나면 추출이 규약을 벗어난 것이므로 FAIL.
- RED-now: RED-4. **citations 계층 자체는 현재 녹색이므로 RED가 아니다** — 회귀 가드로 분류한다(§B.1).

### AC-CM2-007 — 패키지 구조 대조 (MUST, accuracy b)

- **Given** M1의 판정 결과가 있고
- **When** 재생성 후 §A.3(a) 히트-0 패키지 명령을 다시 돌리면
- **Then** 증거 파일 §⑥에 잔여 히트-0 패키지 전수가 열거되고, 각각에 M1 판정(fold / omission)이 붙어 있다.
- **판정은 개수가 아니라 분류의 존재로 이분한다.** 미히트 개수 자체는 합격 조건이 아니다(§B.1 — 트리와 함께 움직이는 값이다). 분류 없는 개수 보고만 있으면 FAIL.
- **omission 판정 단위가 이 목록에 남아 있으면 그것은 기록 대상이 아니라 AC-CM2-004 FAIL이다.** 기록 전용 처분은 fold에만 적용된다.
- RED-now: RED-4.

### AC-CM2-008 — 인용 식별자 hit/miss (MUST, accuracy c)

- **Given** 재생성된 `entry-points.md` / `data-flow.md`가 식별자를 인용하고
- **When** REQ-CM2-008이 명명한 추출 명령(백틱 인라인 코드 중 Go 식별자 형태)을 실행해 각 식별자를 명명된 파일·패키지에 대해 해석하면
- **Then** 증거 파일 §⑦에 (식별자 → 명명 위치 → hit/miss) 전수 표가 존재한다. miss는 기록만 하며 인용 본문을 임의로 지우지 않는다.
- **"식별자"의 정의는 명령의 출력이다.** 산문 정의만 있고 명령이 없으면 FAIL(iter-1 D6).
- **추출이 0행을 내면 통과가 아니라 blocker다** — 빈 집합 위의 공허한 통과이며, 추출식이 재생성된 문서 형식과 어긋났다는 뜻이다. 재생성 전 `52f863f36`에서 이 명령은 10행을 낸다.
- RED-now: RED-4.

### AC-CM2-009 — 스탬프 도달성 (MUST)

- **Given** 작업 브랜치가 `WT-codemaps-stale`(통합 브랜치가 아님)이고
- **When** merge-base 명시 재스탬프 후 `provenance.json`의 `commit_sha`를 읽어 `git merge-base --is-ancestor <그 값> origin/develop`를 실행하면
- **Then** exit 0(조상 성립)이다. bare HEAD를 스탬프했으면 FAIL.
- **판정 근거는 `provenance.json`이 기록한 값이다** — 중간 파일에 저장해 둔 "넣었다고 믿는 값"이 아니다.
- **현재는 `merge-base`와 `HEAD`가 같아(둘 다 `52f863f36`) 이 AC가 물지 않는다.** run이 자체 커밋을 쌓는 순간 갈라지므로 명시 형식은 규정으로 남는다. 이 구간에 한해 regression-guard로 읽는다.
- RED-now: 현재 `provenance.json`의 `commit_sha`는 `25a3212a9…`이며 이는 재스탬프 전 값이다(`52f863f36`). green path: M4 후 그 값이 merge-base로 바뀐다.

### AC-CM2-010 — 게이트 종결 (MUST)

- **Given** M2 재생성과 M4 재스탬프가 완료됐고
- **When** `./bin/moai graph check`를 실행하면
- **Then** codemaps 행이 `verdict=fresh`이고 value < 40(기대 0)이며, `citations` 행이 `verdict=fresh`이고, 다른 어떤 계층도 `verdict=stale`이 아니다. mx-index/edges의 `verdict=absent`는 신규 워크트리 예상 상태로 합격을 막지 않는다.
- **이 AC 단독으로는 카드가 닫히지 않는다.** 재스탬프만으로 value가 0이 되므로, 재생성을 건너뛰어도 이 AC는 통과한다. AC-CM2-002~008이 그 구멍을 막는다(REQ-CM2-011).
- RED-now: RED-1.

### AC-CM2-011 — 범위 위생 (MUST)

- **Given** run이 종료 직전이고
- **When** `git status --porcelain`을 실행하면
- **Then** 변경·추가된 모든 경로가 `.moai/project/codemaps/`, `.moai/reports/t475/`, `.moai/specs/SPEC-CODEMAPS-REFRESH-002/` 세 접두사 중 하나에 속한다. `internal/`·`pkg/`·`cmd/`·`gate.yaml`·다른 SPEC 디렉터리에 변경이 하나라도 있으면 FAIL.
- RED-now: 이 AC는 **작업 전 이미 녹색**이다(`52f863f36`에서 `git status --porcelain`이 두 항목만, 둘 다 허용 경로). 따라서 release-blocking이 아니라 **regression-guard**로 분류한다 — 작업이 이 상태를 깨지 않았음을 보는 것이 목적이다(§B.1의 처분 원칙과 같은 이유).

### AC-CM2-012 — 관측 리포트 (MUST)

- **Given** M5가 게이트를 닫았고
- **When** `.moai/reports/t475/verdict.md`를 읽으면
- **Then** 세 관측이 전부 존재한다: ① 임계 40 대비 값 거동 — **앵커 귀속(t476 / 2026-09-03T18:18:34Z)과 5일 누적이라는 사실을 포함해야 한다**; 값만 적혀 있고 속도의 근거가 없으면 FAIL ② 후보(A·B·C층) 20개 전수의 fold/omission 분류 요약 ③ **후보에서 제외된 잔여 42개 패키지의 전수 목록**. 셋 중 하나라도 없으면 FAIL — 특히 ③의 부재는 A층 필터의 정당화 조건이 빠진 것이므로 FAIL이다.
- 그리고 AC-CM2-011이 동시에 성립한다 — **관측을 보고했으되 설정은 하나도 바뀌지 않았다.**
- `tree_root`는 이 리포트의 항목이 **아니다**. spec.md §A.4가 `check.go:341-346` 주석을 인용해 설계 의도임을 확정했으므로, 관측 항목으로 올리는 것은 닫힌 조사를 다시 여는 일이다. verdict.md에 wrong-tree 항목이 실려 있으면 FAIL.
- RED-now: RED-6.

## §D.1 Severity

전 항목 MUST. 하나라도 FAIL이면 카드는 종결되지 않는다.

단, `verification-completeness.md` §2.1의 undecidable disposition에 따라 **AC-CM2-009(현재 구간)와 AC-CM2-011은 regression-guard로 분류하며 "통과"로 기록하지 않는다** — 작업 전 이미 그 상태이므로 통과가 새로운 정보를 담지 않는다. 둘은 깨지지 않았음을 보는 가드다.

## §D.2 Traceability

| AC | REQ |
|----|-----|
| AC-CM2-001 | REQ-CM2-001 |
| AC-CM2-002 | REQ-CM2-002 |
| AC-CM2-003 | REQ-CM2-003 |
| AC-CM2-003a | REQ-CM2-014 |
| AC-CM2-004 | REQ-CM2-004 |
| AC-CM2-005 | REQ-CM2-005 |
| AC-CM2-006 | REQ-CM2-006 |
| AC-CM2-007 | REQ-CM2-007 |
| AC-CM2-008 | REQ-CM2-008 |
| AC-CM2-009 | REQ-CM2-009 |
| AC-CM2-010 | REQ-CM2-010 |
| AC-CM2-011 | REQ-CM2-012 |
| AC-CM2-012 | REQ-CM2-013 |

REQ-CM2-011(증거 독립성)은 단일 AC에 대응하지 않는다 — AC-CM2-006~008이 AC-CM2-010과 **무관하게** 성립해야 한다는 구조 요건이며, 게이트가 녹색이라는 이유로 006~008을 생략하면 셋 모두 FAIL이다.

## §D.3 Definition of Done

- AC-CM2-001~012(003a 포함, 총 13항목) 전부 PASS 또는 regression-guard 성립.
- `.moai/reports/t475/codemaps-accuracy-verification.md`(**7섹션**)와 `verdict.md`(**관측 3항목**)가 존재하고 tracked 경로에 있다. `pre-regen/`에 6개 사본이 있다.
- progress.md §E.2에 각 마일스톤의 명령 + 출력 + HEAD SHA가 기록돼 있다.
- 변경 집합이 허용 3경로에 한정된다.
