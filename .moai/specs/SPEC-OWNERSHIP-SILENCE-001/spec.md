---
id: SPEC-OWNERSHIP-SILENCE-001
title: "OwnershipTransitionRule 무음 통과 수리 — trailer-less 전환을 unmeasured 로 명시 보고"
version: "0.1.0"
status: in-progress
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/spec, .claude/rules/moai/development, internal/template/templates/.claude/rules/moai/development"
lifecycle: spec-anchored
tags: "ownership-lint, silent-nil, unmeasured, authored-by-agent, doc-drift, spec-lint"
tier: M
era: V3R6
---

# SPEC: OwnershipTransitionRule 무음 통과 수리 — trailer-less 전환을 unmeasured 로 명시 보고

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-08 | manager-spec | 최초 작성 — 세 갈래 판정에서 (c) 채택: trailer-less 상태 전환을 무음 통과 대신 Info `OwnershipTransitionUnmeasured` 로 명시 보고. 배차 전제("트레일러 보유 커밋 없음")를 시대 한정 형태로 정정해 기록 |

## 1. 문제 — 측정된 형태

`OwnershipTransitionRule` 은 등록되고 실행되지만, 판정 재료가 있는 상황에서도 조용히 통과하는
분기가 4곳 있다. 실측 baseline(`.moai/reports/t572/measurement-baseline.md`, 트리
`3ac58b5a1` = origin/develop tip 기준)이 확정한 좌표:

| 사이트 | 조건 | 현재 동작 | 본 SPEC 의 판정 |
|--------|------|----------|----------------|
| `internal/spec/lint_ownership.go:400-402` | `rec == nil` — 50커밋 창(`gitLogWindowSize`, `internal/spec/drift.go:302`) 안에 전환 없음 | 조용한 nil | 설계대로 유지 (창 밖 전환은 전환 없는 것과 같다 — 다수 아카이브 SPEC 이 걸린다) |
| `internal/spec/lint_ownership.go:405-407` | `expected == ownerNone` — 매트릭스 미정의 전환 | 조용한 nil | 설계대로 유지 (의미 없는 변화 — 플래그할 위반 자체가 없다) |
| **`internal/spec/lint_ownership.go:409-416`** | **전환은 발견됐는데 commit 에 `Authored-By-Agent:` 트레일러가 없음** | **조용한 nil (본 카드의 피의자)** | **수리 — Info `OwnershipTransitionUnmeasured` 명시 보고** |
| `internal/spec/lint_ownership.go:419-422` | 트레일러 있으나 미인식 행위자(예: manager-git) | 조용한 nil | 설계대로 유지 (전환-관련 행위자가 아니다) |

피의자 지점의 주석은 M4 AC-LSG-004 의 false-positive 가드다 — 당시에는 옳았다. 그러나 그
가드는 "트레일러 관례가 살아있다"는 전제 위에 서 있다. 실측 결과 그 전제는 깨져 있다:

```
$ git log --all --format='%(trailers:key=Authored-By-Agent,valueonly)' | /usr/bin/grep -c '[^[:space:]]'
104
$ git rev-list --all --count
11917
$ git log --all --format='%h%x09%ad%x09%s%x09%(trailers:key=Authored-By-Agent,valueonly)' --date=short | awk -F'\t' '$4 != "" {print; n++; if(n>=3) exit}'
a20fba05f	2026-09-03	docs(SPEC-LANE-PUSH-DOC-001): M2 run-phase evidence export (t463)	manager-develop
a30edfe98	2026-09-03	fix(SPEC-LANE-PUSH-DOC-001): M1 lane push actor sentence repair (t463)	manager-develop
b1bcce4f4	2026-09-01	fix(SPEC-STATUS-TRANSITION-VALIDITY-001): memoize the status-transition history walk (t376)	manager-develop
```

트레일러 보유 커밋은 전체 11,917 건 중 104 건(SPEC 경로 83 건)이고, **가장 최근 보유자가
2026-09-03** 이다. 즉 그 이후 착지한 모든 카드의 상태 전환은 "판정할 재료(트레일러)가 없는
것"이 아니라 "재료가 있어야 했는데 없는 것"인데도 조용히 통과한다 — 이 규칙은 현재 시대에
구조적으로 침묵하며, 인용하는 문서는 여전히 그 초록을 억지 근거로 인용한다("공백 위에 선
인용").

### 1.1 배차 전제 정정 (본 SPEC 이 기록하는 의무)

배차문의 전제 — "이 저장소의 커밋은 트레일러를 달지 않는다" — 는 **그대로는 거짓**이다.
트레일러 보유 커밋이 104 건 존재한다. 참이 되는 형태는 시대 한정 문장이다:

> 2026-09-03 이후의 모든 상태 전환 커밋에는 트레일러가 없다. 따라서 그 기간의 전환은
> 전부 무음 통과했다.

이 정정은 판정을 바꾸지 않는다 — 오히려 강화한다. 관례가 "원래 없던 것"이 아니라 "죽은
것"이라면, 죽은 관례를 되살리거나(a) 침묵을 깨는(c) 것이 선택지가 되고, 본 SPEC 은 후자를
채택한다(§3).

### 1.2 strict 승급 축 — Info 는 게이트를 건드리지 않는다

```
$ git grep -n "Strict" origin/develop -- internal/spec internal/cli | grep -v _test | grep lint.go
internal/spec/lint.go:62:		if r.Strict && f.Severity == SeverityWarning && !f.Advisory {
```

CI(`.github/workflows/spec-lint.yml:58`, `go run ./cmd/moai spec lint --strict`, 트리거
`.moai/specs/**` + `internal/spec/**`, `fetch-depth: 0`)의 strict 승급은 non-Advisory
**Warning 한정**이다. 본 SPEC 이 추가할 SeverityInfo 발견은 strict 여부와 무관하게 종료
상태를 바꾸지 못한다 — "시끄러운 미측정"은 게이트를 깨지 않는 소음이다.

## 2. 원인

### 2.1 무음 가드가 전제로 삼은 관례가 죽었다

`:409-416` 의 무음 분기는 M4 AC-LSG-004 의 false-positive 가드로, 당시 탈출 기준("moai
spec lint against repo emits no false positives")을 만족시키기 위해 의도된 것이었다. 문제는
그 가드가 **측정 불가 상태와 위반 상태를 같은 nil 로 표현한다**는 점이다. 관례가 살아
있던 시기에는 트레일러 부재가 "레거시/비-MoAI 커밋"을 의미했으므로 무음이 합리적이었지만,
관례가 죽은 지금은 트레일러 부재가 **현행 MoAI 커밋의 기본 상태**가 됐고, 무음은 매
전환마다 측정을 포기한 채 초록을 내보내는 것이 됐다. 판정 기준은 배차문의 한 문장이다:
**공허한 초록보다 시끄러운 미측정이 낫다.**

### 2.2 초록을 인용하는 자리 — 문서 드리프트 (함께 수리)

- `.claude/rules/moai/development/spec-frontmatter-schema.md:198`(로컬)과 바이트 동일 미러
  `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md:198`
  이 `OwnershipTransitionInvalid` 의 발동 조건을 **"commit subject prefix 불일치"**로
  기술한다. 이는 M4 이전 구현을 서술하는 드리프트다 — 현 구현의 WHO 신호 SSOT는
  `Authored-By-Agent:` 트레일러(`internal/spec/lint_ownership.go:160`
  `trailerAgentOwnerKind`)이고, subject-prefix 분류(`commitOwnerKind`, `:112`)는 회귀
  테스트 전용으로만 남아 있다. 같은 절의 발견 코드 목록도 두 개(Invalid, Unreachable)로
  되어 있어 수리 후 세 개가 된다.
- `.claude/agents/moai/manager-develop.md:185` + 템플릿 미러(:186) +
  `internal/template/templates/.codex/agents/moai/manager-develop.toml:173` 은
  `OwnershipTransitionInvalid` 를 진행 넘김의 억지 장치로 인용한다. 이 문장은 (c) 채택
  후에도 거짓이 되지 않는다 — Invalid 가 걸릴 조건(트레일러가 있고 소유자가 어긋나는 경우)은
  여전히 실재하는 경로이기 때문이다. **본 카드는 이 인용을 만지지 않는다**(§6).

## 3. 세 갈래 판정

배차가 제시한 세 선택지를 측정 결과 위에서 판정한다. 이 판정은 본 SPEC 의 첫 마일스톤이며,
plan-auditor 와 이후 단계는 이 기록을 재논의하지 않는다(§5).

### 3.1 (a) 커밋이 실제로 트레일러를 다시 달게 한다 — 후속 카드로 권고 (본 카드 스코프 밖)

유일하게 "새 전환을 실제로 측정 가능"하게 만드는 축이다. 그러나 관례 강제는 전 행위자를
건드리는 하네스 변경이다: 에이전트 정의(커밋 관례 절), 커밋 게이트, 템플릿 미러 전파,
16-프로그래밍-언어 배포판 중립성 검토까지 따라온다. 본 카드(단일 무음 분기 수리)의 스코프를
넘는다. **후속 카드 권고로 기록한다.**

**예외 — 도그푸드는 본 카드 안에 들어간다**: 본 카드 자신의 plan/run/sync 커밋은
`Authored-By-Agent:` 트레일러를 단다(plan.md §D 지시). 비용 0으로 관례를 본보기로
되살리는 길이고, 이 카드의 상태 전환들은 수리 이후 실제로 측정되는 첫 사례가 된다.

### 3.2 (b) 트레일러 없이 판정한다(subject-prefix fallback 부활) — 기각

M4 AC-LSG-004 의 기록된 결정과 정면 충돌한다. subject prefix 는 WHO 신호로 불신뢰라는
이유로 퇴역시킨 것이다(`internal/spec/lint_ownership.go:108-111` 의 NOTE 는 그 기록이다).
fallback 을 되살리면 퇴역 사유가 사라진 것이 아니라 무시된 것일 뿐이고, prefix 위조·불일치
false-positive 를 되돌려 받는다. **기각하며, 이 사유를 본 절에 성문화한다.**

### 3.3 (c) 무음 통과를 unmeasured 로 명시 보고한다 — 채택 (본 카드 구현)

판정 기준「공허한 초록보다 시끄러운 미측정」에 직결된다. M4 가드를 훼손하지 않는다 — Info
발견은 "위반"(`OwnershipTransitionInvalid` Warning)의 위장이 아니라 **측정 상태에 대한
진술**이기 때문이다. CI strict 안전은 §1.2 실측으로 확정됐다. 예상 소음은 2026-09-03 이후
전환을 가진 SPEC 수준으로 유한하며 advisory(게이트 무관)다.

## 4. 요구사항 (GEARS)

### 무음 지점의 명시 보고

- **REQ-OWN-001** — `internal/spec/lint_ownership.go` 의 `:409-416`(트레일러 부재 분기)이
  조용한 nil 대신 발견 하나를 반환해야 한다. Severity `SeverityInfo`, 코드
  `OwnershipTransitionUnmeasured`, Advisory 미설정(파일 내 형제
  `OwnershipTransitionSkipped:375-385`·`OwnershipTransitionUnreachable:388-399` 와 동일한
  plain Info 이디엄 — 파일 내 일관성이 lint.go 다른 곳의 warning+Advisory 관례보다 이긴다).
- **REQ-OWN-002** — 발견 메시지는 다음을 모두 담아야 한다: SPEC id, 전환(prev → curr,
  prev 빈 값은 `emptyOrValue` 로 "(none)"), 커밋 SHA, 커밋 subject, 사유("no
  Authored-By-Agent trailer — ownership transition unmeasured"). 관례의 정의 자리는
  SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 spec.md §D.1.6 HARD 이며 트리 내 산문 지점은
  `.claude/agents/harness/workflow-specialist.md:83` 이다(메시지 본문은 절 번호를 인용하지
  않는다 — 번호는 썩는다).

### 보존 (수리가 파괴해서는 안 되는 것)

- **REQ-OWN-003** — 나머지 세 무음 지점(`:400-402` 창 미스, `:405-407` 매트릭스 미정의,
  `:419-422` 미인식 행위자)과 모든 기존 형제 발견(Skipped, Unreachable, Invalid)의 동작은
  바이트 수준에서 변하지 않아야 한다. 세 지점은 설계적 침묵이며 그 사유는 §1 표에 기록돼 있다.
- **REQ-OWN-004** — 새 Info 발견은 `spec lint --strict` 의 종료 상태를 바꾸지 못해야 한다
  (`internal/spec/lint.go:62` 의 승급 조건이 Warning 한정임을 미러링하는 테스트로 증명).
- **REQ-OWN-005** — 새 발견은 M4 false-positive 결함을 부활시키지 않아야 한다: 코드
  `OwnershipTransitionUnmeasured` 의 Info 는 `OwnershipTransitionInvalid` 의 Warning 과
  별개 코드·별개 등급이어야 하며, 측정 불가 상태가 위반으로 보고되는 경로가 있어서는 안 된다.

### 기존 판정의 의도적 이동

- **REQ-OWN-006** — 트레일러-less 커밋에 대해 무음 스킵을 단언하던 M4-시절 테스트(최소
  `internal/spec/lint_ownership_test.go:550-573` `trailer_absent_silent_skip`, 그 외 동일
  단언 전부)는 **의도적 행동 변경으로서** 갱신해야 하며, 각 테스트마다 한 줄의 기록된 사유를
  남긴다(테스트 휘둘리기가 아님을 남기는 기록).

### 비공허성 (검증 도구 자체의 완결)

- **REQ-OWN-007** — 새 분기에 대한 변이 증거가 있어야 한다: 새 발급 경로의 뮤턴트 최소 1건이
  테스트 스위트에 검출되고, 생존 뮤턴트는 왜 허용 가능한지 메모와 함께 기록된다(도달하지
  못한 뮤턴트와 진짜 생존자는 같은 `ok` 를 찍는다 — 도달성을 증명한다).

### 문서 정렬 ("공백 위에 선 인용" 축)

- **REQ-OWN-008** — `OwnershipTransitionRule Cross-Reference` 절을 현실에 맞춘다: WHO
  신호 SSOT(트레일러), 발견 코드 3종(Invalid Warning / Unreachable Info / 신규
  Unmeasured Info), 무음 사이트의 설계적 침묵과 사유. **템플릿 사본 먼저, 그 다음 로컬
  사본** — 두 사본은 편집 전후 모두 바이트 동일을 유지한다(측정: 편집 전 `cmp` 일치 확인,
  §1 실측 완료). 재작성 산문은 템플릿 중립성을 지킨다(카드 id·내부 SHA·내부 날짜 신규 추가
  없음).
- **REQ-OWN-009** — 본 파일의 낡은 주석 두 곳을 같은 편집에서 고친다: `:179`(트레일러 부재
  시 subject-prefix fallback 이라고 서술 — 현 흐름에는 fallback 경로가 없다)와 `:367`
  (Check() doc-comment 의 "2. fallback" 항). 주석-만 변경이며 동작 변경이 아니다.

### 도그푸드 트레일러

- **REQ-OWN-010** — 본 카드의 커밋은 `Authored-By-Agent:` 트레일러를 단다: plan 커밋
  `manager-spec`, run 커밋 `manager-develop`, sync 커밋 `manager-docs`
  (`trailerAgentOwnerKind` 인식 enum, `internal/spec/lint_ownership.go:160-164`). 세부
  지시는 plan.md §D.

## 5. 알려진 구속 조건 — 재논의 금지

1. **세 갈래 판정(§3)은 확정이다.** (c) 채택·(a) 후속 카드·(b) 기각은 plan-auditor 와
   run/sync 단계에서 다시 열지 않는다. 판정 근거가 되는 측정은
   `.moai/reports/t572/measurement-baseline.md` 에 커맨드·출력 전문으로 보존돼 있다.
2. **등급과 코드는 고정이다.** `SeverityInfo` + `OwnershipTransitionUnmeasured`, Advisory
   미설정. Warning 승급이나 Advisory:true 로의 변경은 CI 소음 지도(lint.go:62)와 파일 내
   이디엄 양쪽을 깬다.
3. **보존 경계는 고정이다.** REQ-OWN-003 의 세 무음 지점을 "침묵 제거" 명분으로 함께
   건드리는 것은 스코프 확장이다 — 각각의 사유는 §1 표에 있다.
4. **M4 결함의 원 근거는 유효하다.** 본 수리는 false-positive 가드의 사유를 폐기하는 것이
   아니라, 그 사유가 커버하던 상태 중 "측정 불가"를 위반과 분리해 보고하는 것이다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — 관례 강제의 전면 부활 (판정 (a))

에이전트 정의·커밋 게이트·템플릿 미러 전파를 포함한 `Authored-By-Agent:` 관례의 기계적
강제. 본 카드는 후속 카드 권고만 기록한다(§3.1). 예외인 본 카드 자신의 커밋 트레일러는
REQ-OWN-010 이라는 요구사항으로 스코프 안에 있다.

### Out of Scope — 다른 무음 지점의 동작 변경

`:400-402`(창 미스), `:405-407`(매트릭스 미정의), `:419-422`(미인식 행위자)의 nil 은 그대로
둔다. 창 미스를 보고하면 아카이브 SPEC 다수가 소음을 내고, 미인식 행위자 보고는 전환-무관
커밋을 건넨다 — 둘 다 본 카드가 고치는 "재료가 있어야 했던 전환"과 다른 상태다.

### Out of Scope — manager-develop 인용 문장 편집

`.claude/agents/moai/manager-develop.md:185`·템플릿 미러·`.codex` toml(:173)의
`OwnershipTransitionInvalid` 억지 인용은 (c) 채택 후에도 참이므로 만지지 않는다(§2.2).
에이전트 파일을 건드리지 않으므로 `make agents-emit` 도 불필요하다.

### Out of Scope — 상속 CI 적색 축

t577(lane-9 errcheck 축) 등 develop 에 이미 열려 있는 CI 적색은 본 카드 판정과 무관하다.
본 카드의 검증은 스코프된 패키지로 한정한다(AC-OWN-007).

### Out of Scope — 로컬 전체 스위트 실행

`go test ./...` 전체를 로컬에서 돌리지 않는다(레인 부하 규율). 전 수트 판정은 develop push
이후 CI 몫이다.

## 7. 미검증 항목 (Gaps)

- **unmeasured 발화 볼륨의 per-SPEC 정확한 수** — 구현 후 `go run ./cmd/moai spec lint`
  실측으로 확정 예정(815 디렉터 중 창 안에 전환을 가진 분만). 유한하다는 것(§3.3)은
  측정됐지만, 건수는 아직 측정 전이다.
- **변경 전 baseline `--strict` 전수 실측** — 머신 부하로 백그라운드 진행 중이었던 항목이며,
  완료 시 `.moai/reports/t572/measurement-baseline.md` 갱신으로 귀속된다. 무음 입증 자체는
  코드·트레일러 측정으로 이미 성립했다(§1).

## Owning SPEC

- 무음 분기 수리: SPEC-OWNERSHIP-SILENCE-001 (본 SPEC)
- 규칙·행렬 원천: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 (M4 AC-LSG-004, §D.1.6)
- 실측 baseline: `.moai/reports/t572/measurement-baseline.md` (카드 t572, 트리 3ac58b5a1)
