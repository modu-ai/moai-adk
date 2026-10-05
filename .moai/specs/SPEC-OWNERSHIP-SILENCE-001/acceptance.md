# Acceptance: SPEC-OWNERSHIP-SILENCE-001

> AC 코드는 `AC-OWN-nnn` (배차문 AC-1..AC-8 과 1:1 대응 — AC-OWN-001=배차 AC-1, …,
> AC-OWN-008=배차 AC-8). 판정 원칙: 주장이 아니라 관측이다 — 모든 PASS 는 커맨드 + verbatim
> 출력 + 트리 SHA 귀속을 요구한다(`.claude/rules/moai/core/verification-claim-integrity.md`
> §1, §2).

## §D AC 매트릭스

| AC | 요구 (한 줄) | 판정 방법 | 등급 |
|----|-------------|-----------|------|
| AC-OWN-001 | 새 분기는 RED 부터 — 무음 분기를 잡는 테스트가 pre-fix 트리에서 실패가 관측된다 | RED 실행 기록 (커맨드·출력·종료코드·트리 SHA) | High (release-blocking) |
| AC-OWN-002 | 피의자 지점만 발급하고, 세 무음 지점 + 형제 발견은 동작 불변 | 신규·기존 보존 테스트 GREEN + diff 스코프 판독 | High |
| AC-OWN-003 | M4-시절 무음 단언 테스트의 의도적 갱신, 테스트마다 기록된 사유 | 테스트 diff 판독 + 사유 주석 존재 | High |
| AC-OWN-004 | 새 Info 발견은 `--strict` 종료 상태를 바꾸지 못한다 | strict 안전 유닛 테스트 (lint.go:62 미러) | High (release-blocking) |
| AC-OWN-005 | 새 발급 경로의 뮤턴트 최소 1건 검출, 생존 뮤턴트는 왜-허용 메모 | 변이 관측 기록 (주입 → 스위트 판정 → 원복 확인) | High |
| AC-OWN-006 | 규칙 문서 양 사본 정렬 — 템플릿 먼저, 바이트 쌍둥이 유지 | `cmp -s` 일치 + 중립성 가드 통과 + 재작성 산문 판독 | Medium |
| AC-OWN-007 | 스코프된 검증 — `go test ./internal/spec/...` GREEN + `go vet` 청결, 로컬 전수 금지 | 워크트리 안 실행 출력 | High |
| AC-OWN-008 | 모든 검증 증거가 `.moai/reports/t572/` 에 반출돼 있다 | 경로 존재 + 파일 내 커맨드·출력 대조 | Medium |

## §D.1 AC 상세

### AC-OWN-001 — RED 가 pre-fix 트리에서 관측됐다 (배차 AC-1, Reproduction-First)

**GREEN-path 셀(선언)**: `internal/spec/lint_ownership_test.go` 의 신규 테스트 — fake 주입
(`getOwnershipTransitionRunner`, `internal/spec/lint_ownership.go:192`)으로 "발견된 전환 +
`AuthoredByAgent == ""`" 픽스처를 걸고, `OwnershipTransitionUnmeasured` Info 발견 1건을
단언한다. 이 테스트는 구현 전 트리에서 **실패해야 정당하다**.

**RED-now 셀(관측)** — run-phase M1 이 다음 4요소를 `.moai/reports/t572/` 아래 기록한다:

- (a) 커맨드 — 단일 호출 형태, 예: `go test ./internal/spec/ -run TestOwnershipTransitionRule_TrailerAbsent -count=1`
- (b) verbatim stdout — `--- FAIL:` 행이 포함된 원본 출력 (요약 금지)
- (c) 종료 코드 — `1` (별도 필드로 기록; 커맨드 안에 `; echo $?` 를 붙이지 않는다)
- (d) 트리 SHA — pre-fix 커밋 SHA (브랜치 이름 금지)

**RED 사유 명시(의무)**: 실패 사유는 "무음 nil 분기가 발견을 반환하지 않기 때문"이어야 한다.
다른 이유(컴파일 오류, 무관 테스트 충돌)의 RED 는 이 AC 를 만족하지 않는다 — WRONG-reason
red 는 셀을 무효로 만든다(verification-completeness §2).

### AC-OWN-002 — 피의자만 수리됐고 나머지는 불변이다 (배차 AC-2)

- **주장**: `lint_ownership.go` 의 트레일러-부재 분기(3ac58b5a1 기준 `:409-416`)가
  REQ-OWN-001 대로 `OwnershipTransitionUnmeasured` Info 를 발급한다. REQ-OWN-002 의 메시지
  5요소(SPEC id / prev → curr, 빈 prev는 "(none)" / 커밋 SHA / subject / 사유)가 메시지에
  모두 나타난다. REQ-OWN-003 의 보존 대상(세 무음 지점 + 형제 발견)은 동작 불변이다.
- **보존 주장**: `:400-402`(창 미스), `:405-407`(매트릭스 미정의), `:419-422`(미인식
  행위자)는 nil, 형제 발견(`OwnershipTransitionSkipped`, `OwnershipTransitionUnreachable`,
  `OwnershipTransitionInvalid`)은 기존 그대로.
- **판정**: (1) 신규 발급 테스트 GREEN(픽스처 2종 — prev 있는 전환과 `(none) → draft`
  전환으로 emptyOrValue 양쪽 경로 관측), (2) 보존 대상 기존 테스트가 무변경으로 통과,
  (3) `git diff` 로 보존 대상 분기의 본문이 건드리지 않았음과 REQ-OWN-009 의 두 주석
  수정(`:179`, `:367`)이 같은 diff 에 함께 착지했음을 판독. 세 가지 모두 기록.
- **주의**: 보존 테스트가 RED 가 되면 본 카드 diff 가 범위를 넘은 것이다 — 테스트를
  "맞추는" 것이 아니라 diff 를 줄인다.

### AC-OWN-003 — 테스트 이동은 의도적이다 (배차 AC-3)

- **주장**: 무음 스킵을 단언하던 기존 테스트(최소
  `lint_ownership_test.go:550-573` `trailer_absent_silent_skip`, 동일 단언 형제 전부)가
  갱신됐고, 각 갱신 위에 한 줄 사유가 기록돼 있다.
- **판정**: 테스트 diff 전수 판독 — (1) 무음 단언을 건드린 모든 테스트에
  `// 의도적 행동 변경 (SPEC-OWNERSHIP-SILENCE-001 REQ-OWN-006): 무음 스킵 → unmeasured
  Info 보고` 형태의 사유 주석이 존재, (2) 사유 없이 바뀐 단언이 0건, (3) 갱신된 테스트 이름
  목록을 증거 파일에 표로 기록.
- **반증 조건**: 사유 없는 무음-단언 삭제, 또는 단언을 약화시켜(발견 수 0 단언 제거 등)
  통과만 만든 diff. 둘 다 이 AC 의 FAIL 이다.

### AC-OWN-004 — Info 는 strict 게이트를 건드리지 않는다 (배차 AC-4)

- **주장**: `OwnershipTransitionUnmeasured` 발견이 포함된 Report 는 `Strict=true` 에서도
  `HasErrors() == false` 다(REQ-OWN-004). REQ-OWN-005 대로 신규 코드는
  `OwnershipTransitionInvalid` 와 별개 코드·별개 등급이다 — 측정 불가 상태가 위반으로
  보고되는 경로가 없다.
- **판정**: 유닛 테스트 1건 — `internal/spec/lint.go:62` 승급 조건(`SeverityWarning &&
  !f.Advisory`)을 미러링하는 단언 + 신규 발견의 코드·등급이 Invalid(Warning)와 다름을
  단언. 신규 Info 발견 단독 Report + 실제 Check() 반환 Report 양쪽으로 관측한다. 커맨드:
  `go test ./internal/spec/ -run 'Strict' -count=1` 계열, 출력 기록.
- **보강(보조)**: 구현 후 `go run ./cmd/moai spec lint --strict` 의 종료 상태가 본 카드
  이전 측정과 동일함을 비교 기록(plan.md §E). 1차 판정자는 유닛 테스트다.

### AC-OWN-005 — 새 분기의 검출은 공허하지 않다 (배차 AC-5, non-vacuity)

- **주장**: REQ-OWN-007 대로 새 발급 경로의 뮤턴트 최소 1건이 스위트에 검출됐고, 생존
  뮤턴트는 왜-허용 메모와 함께 기록됐다.
- **판정**: plan.md §G 의 뮤턴트 표(m1 발급 삭제 / m2 Warning 승급 / m3 emptyOrValue 제거)
  중 최소 1건 주입 → 스위트 실행 → 검출 관측(FAIL 출력 verbatim) → **원복 후 `git diff`
  로 임시 편집 소멸 확인**. 이 4단계 전부가 증거 파일에 순서대로 존재해야 한다.
- **도달성 원칙**: 검출 기록은 "테스트가 통과했다"가 아니라 "뮤턴트가 FAIL 로 뒤집혔다"여야
  한다 — 도달하지 못한 뮤턴트와 진짜 생존자는 같은 `ok` 를 찍는다. 생존 뮤턴트가 있다면
  잡지 못한 이유를 한 줄로 남긴다(메모 없는 생존 = 미측정).

### AC-OWN-006 — 문서가 현실을 서술한다 (배차 AC-6)

- **주장**: REQ-OWN-008 대로 `## OwnershipTransitionRule Cross-Reference` 절이 트레일러
  SSOT + 발견 코드 3종(Invalid/Unreachable/Unmeasured) + 무음 사이트 설계 침묵을 서술하며,
  구(subject prefix) 트리거 문장이 사라졌다.
- **판정**: (1) 템플릿 사본 편집이 로컬 사본보다 먼저임이 커밋 내역으로 확인 — 단일 커밋에
  담을 경우 diff 상 템플릿 경로가 포함되고 내용 동일임으로 대체 판정, (2) `cmp -s` 양 사본
  일치(편집 후) exit 0, (3) `grep -c "subject prefix"` 양 사본에서 0 (구 트리거 잔존 없음),
  (4) `go test ./internal/template/...` 중립성·유출 가드 GREEN, (5) drift 정정 기록이 본
  SPEC HISTORY + progress.md 에 존재(규칙 문서 자체에는 내부 식별자를 새로 쓰지 않는다 —
  템플릿 중립성).
- **불접촉 확인**: `git diff --name-only` 에 `manager-develop.md`와 `.codex/agents/moai/*`
  가 없음 — 인용 문장은 (c) 채택 후에도 참이므로 만지지 않았다(plan.md §D.4).

### AC-OWN-007 — 검증은 스코프돼 있고 청결하다 (배차 AC-7)

- **주장**: 워크트리 안에서 `go test ./internal/spec/...` GREEN, `go vet ./internal/spec/...`
  청결. 로컬 전체 스위트(`go test ./...`)는 실행하지 않았다.
- **판정**: 두 커맨드의 verbatim 출력(패키지 행 + `ok` 행)을 증거로 기록. 전체 스위트 미실행은
  규율 준수다 — 부재가 아니라 의도다.
- **축 분리**: t577(errcheck 축) 등 상속 CI 적색은 본 AC 판정에 집계하지 않는다. 다만
  `internal/spec` 패키지 안에서 본 카드 이전부터 붉었던 테스트가 있다면 §C.4 절차로
  구분-기록한다(숨김이 아니라 귀속이다).

### AC-OWN-008 — 증거가 살아 있는 경로에 있다 (배차 AC-8)

- **주장**: 위 AC 전부의 커맨드·verbatim 출력·종료 코드·트리 SHA가
  `.moai/reports/t572/` 하위에 반출돼 있다.
- **판정**: 증거 파일 경로 목록 + 각 파일이 주장하는 AC 매핑 표를 run-phase가
  progress.md §E.2 에 남긴다. `/tmp` 반출은 이 AC 의 FAIL 이다(감사 시점에 경로가
  살아있어야 한다).

## §D.2 판정 시 유의 — 이 SPEC 고유의 두 함정

1. **"발견이 늘었다"를 결함으로 읽지 않는다.** 수리 후 `moai spec lint` 의 Info 건수는
   **의도적으로 증가**한다(2026-09-03 이후 전환 보유 SPEC 만큼). 이 증가가 본 카드의
   산출이다. INFO 등급 행은 종료 상태를 바꾸지 않으므로 CI 적색과 혼동하지 않는다.
2. **도그푸드 트레일러는 본 카드 자신이 검증 대상이다.** run/sync 커밋에
   `Authored-By-Agent:` 가 붙었는지는 plan.md §D.6 지시 이행의 관측점이다 — 커밋 후
   `git log --format='%(trailers:key=Authored-By-Agent,valueonly)' -1` 로 확인해 증거에
   남긴다. 누락됐다면 FAIL 이 아니라 "관례 미이행"으로 별도 기록한다(REQ-OWN-010 은 SHOULD
   등급).
