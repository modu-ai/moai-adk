# Plan: SPEC-OWNERSHIP-SILENCE-001

## §A 맥락

- 카드: t572 (Factory Mode lane-11, 배차 2026-09-08). 워크트리
  `.claude/worktrees/t572`, 브랜치 `WT-ownership-lint-silent` (origin/develop tip
  `3ac58b5a1` 기반).
- 실측 baseline: `.moai/reports/t572/measurement-baseline.md` — 커맨드·출력 전문 포함,
  run-phase 좌표 재확인의 1차 근거.
- 판정 기준(리드 성문): "공허한 초록보다 시끄러운 미측정이 낫다". 세 갈래 판정은 spec.md
  §3 에서 (c) 채택으로 확정 — plan.md 는 그 구현 절차만 운반한다.

### 카드 등급 판정

결함 수리 + 의도적 행동 변경 1분기 + 문서 정렬 1축. 설계 판단은 이미 SPEC 이 소유했다.
다만 기존 테스트의 의도적 이동(REQ-OWN-006)과 변이 증거(REQ-OWN-007)가 있어 단순 한 줄
수리로 분류하지 않는다.

### Tier 판정 — M (3 산출물: spec.md / plan.md / acceptance.md)

파일 축: Go 소스 1(`lint_ownership.go`) + Go 테스트 1(`lint_ownership_test.go`) + 규칙 문서
2(템플릿+로컬 바이트 쌍둥이, 동일 내용 1회 편집) + 산출물 3. 마일스톤 3(RED → GREEN → 문서).
Tier M 예산 안에서 압축 없이 진행한다.

## §B 알려진 결함 (이미 측정됨)

| 결함 | 좌표 (3ac58b5a1) | 측정 근거 |
|------|-----------------|-----------|
| 피의자: 전환 발견 + 트레일러 부재 → 조용한 nil | `internal/spec/lint_ownership.go:409-416` | baseline §Evidence-코드 지도 + 본 SPEC §1 |
| 트레일러 관례 실천 단절 — 최근 보유자 2026-09-03 | 전체 104/11,917 건, SPEC 경로 83건 | baseline §Evidence-트레일러 실측 (커맨드+출력 전문) |
| 문서 드리프트 — 구(subject prefix) 메커니즘 서술 | `.claude/rules/moai/development/spec-frontmatter-schema.md:198` + 템플릿 미러 :198 (바이트 쌍둥이, `cmp` 일치 실측) | 본 SPEC §2.2 |
| 낡은 주석 — 죽은 fallback 을 서술 | `lint_ownership.go:179`, `:367` | 본 SPEC §2.2·REQ-OWN-009 |

배차 전제 정정은 spec.md §1.1 에 기록돼 있다: "트레일러 보유 커밋 없음"은 거짓, "2026-09-03
이후 전환에는 없다"가 참인 형태다.

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈추고 보고)

1. **좌표 유효성** — develop 의 `lint.go` 는 t518(axes) 축에서 활발히 변동 중이다(baseline
   Residual-risk). 병합 시점 기준으로 `grep -n "OwnershipTransitionRule" internal/spec/lint.go`
   와 `grep -n "AuthoredByAgent ==" internal/spec/lint_ownership.go` 를 재실행해
   `:409-416`·`:147` 좌표가 이동하지 않았는지 확인한다. 이동했다면 새 좌표로 재귀속하고,
   분기 자체가 사라졌다면(외부 변경) 멈추고 리드에 보고한다.
2. **형제 지점 무변경** — `:375-385`(Skipped), `:388-399`(Unreachable), `:400-402`,
   `:405-407`, `:419-422` 가 baseline 과 동일한 형태인지 판독한다.
3. **쌍둥이 일치** — `cmp -s` 로 규칙 문서 두 사본의 편집 전 일치를 재확인한다(본 카드가
   받은 시점에는 일치 — 256행 동일, 실측 완료).
4. **상속 적색** — `go test ./internal/spec/...` 시작 시점에 이미 붉은 테스트가 있으면
   t577 축 소관과 구분해 기록한다. 본 카드의 GREEN 판정은 "본 카드가 건드린 영역"으로
   스코프한다(AC-OWN-007).

## §D 구속 조건 (재논의 금지)

1. **판정 고정** — (c) 채택 / (a) 후속 카드 / (b) 기각. 구현 중 이 선택을 다시 열지 않는다.
2. **등급·코드 고정** — `SeverityInfo` + `OwnershipTransitionUnmeasured` + Advisory 미설정.
   warning+Advisory 이디엄(lint.go:1056 계열)으로 바꾸지 않는다 — 파일 내 형제(plain Info)
   일관성이 우선이다(spec.md §4 REQ-OWN-001).
3. **보존 고정** — `:400-402`, `:405-407`, `:419-422` 세 무음 지점과 Skipped/Unreachable/
   Invalid 발견은 동작 불변. 이 넷을 건드리는 diff 는 결함이다.
4. **manager-develop 인용 불접촉** — `.claude/agents/moai/manager-develop.md:185`·템플릿
   미러(`:186`)·`.codex/agents/moai/manager-develop.toml:173` 을 만지지 않는다. 문장은
   (c) 채택 후에도 참이다(spec.md §2.2). 에이전트 파일 무변경이므로 **`make agents-emit`
   불필요** — 규칙 문서에는 `.codex` 쌍둥이가 없다.
5. **템플릿 우선** — 규칙 문서는 `internal/template/templates/.../spec-frontmatter-schema.md`
   를 먼저 고치고 동일 내용을 로컬 `.claude/rules/moai/development/spec-frontmatter-schema.md`
   에 적용한다. 두 사본은 편집 후에도 바이트 동일을 유지하고, 신규 산문은 템플릿 중립을
   지킨다(카드 id·내부 SHA·내부 날짜 추가 금지 — drift 정정 기록은 SPEC HISTORY 가 담는다).
   편집 후 `go test ./internal/template/...` 로 중립성·유출 가드를 통과시킨다. 단, 전체
   템플릿 스위트 중 결함 축과 무관한 기존 적색이 있으면 §C.4 와 같이 구분해 기록한다.
6. **도그푸드 트레일러 (run/sync 지시)** — 본 카드의 커밋은 commit body 에
   `Authored-By-Agent:` 트레일러를 단다. 값은 단계별로 고정: plan 커밋 `manager-spec`,
   run 커밋 `manager-develop`, sync 커밋 `manager-docs` (`trailerAgentOwnerKind` 인식
   enum, `internal/spec/lint_ownership.go:160-164`; manager-git 은 미인식이다). 예:
   `fix(SPEC-OWNERSHIP-SILENCE-001): M1 RED — unmeasured finding reproduction` 커밋 body
   마지막에 `Authored-By-Agent: manager-develop`. 커밋 제목은 Status Transition Ownership
   Matrix 의 canonical subject pattern 을 따른다.
7. **시간 추정 금지** — 우선순위 라벨과 단계 순서만 사용한다.

## §E 자가 검증

- `go test ./internal/spec/...` (워크트리 안, `-count=1` 캐시 무시는 flaky 의심 시에만)
- `go vet ./internal/spec/...` (건드린 패키지)
- `cmp -s` 규칙 문서 쌍둥이 일치 (편집 후)
- `go run ./cmd/moai spec lint --strict` 의 종료 상태가 본 카드 이전과 동일한지 비교
  (Info 발견이 종료 상태를 바꾸지 않음의 실측 보강 — 유닛 테스트 AC-OWN-004 가 1차 근거,
  이 실측은 보조)
- 증거 전부: `.moai/reports/t572/` 하위로 반출 (커맨드 + verbatim 출력, AC-OWN-008)

## §F 마일스톤

### 순서 구속 — 권고가 아니라 의존성

M1 이 RED 를 관측하기 전에 M2 의 구현을 시작하지 않는다(Reproduction-First). 문서 축(M3)은
동작이 확정된 뒤에만 쓴다 — 문서가 먼저 쓰이면 구현과 어긋난 산문이 착지한다.

### M1 — RED: 무음 분기의 재현 (우선순위: High)

- `internal/spec/lint_ownership_test.go` 에 "발견된 전환 + 트레일러 부재 →
  `OwnershipTransitionUnmeasured` Info 발견"을 단언하는 테스트를 추가한다. 기존 fake
  주입 경로(`getOwnershipTransitionRunner`, `:192`)와 기존 픽스처 스타일
  (`trailer_absent_silent_skip` 의 구조, `:550-573`)을 재사용한다.
- pre-fix 트리에서 실행해 **RED 출력을 커맨드+종료코드+트리 SHA 와 함께 기록**한다
  (AC-OWN-001, verification-completeness §2의 RED-now 셀). RED 사유: "nil 만 반환하는
  현재 분기에는 발견이 없다" — 무의미한 RED 가 아님을 명시한다.
- 종료조건: RED 관측 기록이 `.moai/reports/t572/` 에 존재.

### M2 — GREEN: 발급 + 의도적 이동 + 안전망 (우선순위: High)

1. `:409-416` 분기를 Info 발급으로 교체한다(REQ-OWN-001/002). 메시지 형식은 spec.md
   REQ-OWN-002 의 5요소(SPEC id / prev → curr with emptyOrValue / SHA / subject / 사유)를
   전부 담는다. 같은 편집에서 `:179`, `:367` 의 낡은 fallback 주석을 고친다(REQ-OWN-009).
2. M4-시절 무음 단언 테스트를 의도적으로 갱신한다(REQ-OWN-006) — 각 테스트 위에 한 줄
   사유 주석: "의도적 행동 변경 (SPEC-OWNERSHIP-SILENCE-001 REQ-OWN-006) — 무음 스킵이
   unmeasured Info 보고로 대체됨". `trailer_absent_silent_skip` 은
   `trailer_absent_emits_unmeasured` 형태로 대체한다.
3. strict 안전 테스트를 추가한다(AC-OWN-004): 새 Info 발견이 들어간 Report 에 대해
   `HasErrors()`가 `Strict=true` 에서도 false 임을 단언(`lint.go:62` 로직 미러).
4. 보존 테스트를 확인한다(AC-OWN-002): 세 무음 지점·형제 발견에 대한 기존 테스트가
   무변경으로 통과하는지 — 통과하지 못했다면 본 카드 diff 가 범위를 넘은 것이다.
5. 변이 증거(AC-OWN-005): 새 발급 경로 뮤턴트 최소 1건(예: 발급 호출 삭제 / 등급 Warning
   승급 / emptyOrValue 제거)을 주입해 스위트가 잡는지 관측하고, 잡은 뮤턴트·생존 뮤턴트를
   왜-허용 메모와 함께 기록한다. 뮤턴트 주입은 임시 편집이므로 **측정 후 원복을 커밋 전
   `git diff` 로 확인**한다(공유 트리 배타성 — 이 워크트리는 본 레인 단일 작성자).
- 종료조건: `go test ./internal/spec/...` GREEN + `go vet` 청결 + 변이 기록 존재.

### M3 — 문서 정렬 + 최종 검증 (우선순위: Medium)

1. 템플릿 사본의 `## OwnershipTransitionRule Cross-Reference` 절(194-205행 부근)을
   현실에 맞춰 재작성한다(REQ-OWN-008): 트레일러 SSOT, 발견 코드 3종, 무음 사이트의
   설계적 침묵, strict 승급 조건의 정확한 서술. **템플릿 중립 산문**(카드 id·내부
   SHA·내부 날짜 신규 없음).
2. 동일 내용을 로컬 사본에 적용하고 `cmp -s` 로 쌍둥이 일치를 확인한다.
3. `go test ./internal/template/...` 중 중립성·유출 가드를 통과시킨다(§D.5).
4. 스코프된 최종 검증(§E 배치)을 단일 턴 병렬로 실행하고, 증거를 `.moai/reports/t572/`
   로 반출한다(AC-OWN-007/008).
5. run/sync 커밋에 도그푸드 트레일러를 확인한다(§D.6).
- 종료조건: 쌍둥이 일치 + 가드 통과 + 증거 반출 완료.

## §G AC별 mutant 노트

| 뮤턴트 | 목표 AC | 예상 판정 | 비고 |
|--------|---------|-----------|------|
| m1: `:414-416` 발급 분기 삭제(원래 nil 로 복원) | AC-OWN-001 | 검출 — M1 RED 테스트가 GREEN 유지를 어긴다 | 도달성 증명: 테스트가 이 분기를 실제로 지난다 |
| m2: 등급을 `SeverityWarning` 로 승급 | AC-OWN-004 | 검출 — strict 테스트가 `HasErrors()==true` 로 뒤집히는 것을 잡는다 | CI 소음 지도 회귀를 잡는 유일 판정자 |
| m3: `emptyOrValue(rec.PreviousStatus)` 제거(원시 값 사용) | AC-OWN-002/메시지 형식 | 검출 — 메시지 5요소 단언이 "(none)" 부재를 잡는다 | (none) → draft 전환 픽스처에서만 관측 가능 — 픽스처 선택이 판정을 만든다 |
| 생존 허용 예: 코드 문자열 오탈자 뮤턴트가 `lint.skip` 매칭 테스트만 건너뛰는 경우 | — | 생존 기록 | `lint.skip` 은 `OwnershipTransitionInvalid` 만 검사하므로 신규 코드와 무관 — why-acceptable 메모로 남긴다 |

판정 원칙: 검출된 뮤턴트는 커밋 전 원복을 확인하고, 생존 뮤턴트는 "왜 잡지 않아도 되는가"를
한 줄이라도 남긴다(메모 없는 생존은 미측정과 같다).

## §H 잔여 위험

1. **develop 병동** — `lint.go` 는 t518 축에서 변동 중. 병합 시 좌표 이동 가능(§C.1 재측정
   이 유일한 방어).
2. **상속 적색** — t577 errcheck 축이 develop CI 에 열려 있다. 본 카드 판정 축과 무관하되,
   판정 문서에 "무관함의 근거"를 남긴다(AC-OWN-007).
3. **소음 볼륨 미확정** — unmeasured Info 의 per-SPEC 건수는 구현 후 실측이다(§7 Gap).
   유한·advisory 성격은 측정돼 있으나, 만약 수백 건 수준으로 나오면 후속 카드에서
   `lint.skip` 안내 문서화로 낮출 수 있다(본 카드 스코프 밖).
4. **트레일러 파서의 한계** — `parseAuthoredByAgent` 가 body 에서 마지막 매치를 취하는
   등 파서 세부는 M4 그대로다. 본 카드는 파서를 건드리지 않는다.

## §I 교차 참조

- 실측 baseline: `.moai/reports/t572/measurement-baseline.md`
- 규칙·행렬 원천 SPEC: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 (M4 AC-LSG-004, §D.1.6 HARD)
- 관례 산문 지점: `.claude/agents/harness/workflow-specialist.md:83` (user-owned
  namespace — 템플릿 미러 없음, 본 카드 불접촉)
- CI: `.github/workflows/spec-lint.yml:58` (`--strict`, paths 트리거, fetch-depth 0)
- strict 승급: `internal/spec/lint.go:62` (Warning + non-Advisory 한정)
- 후속 카드 후보: 관례 기계적 강제(a안) — 에이전트 정의·커밋 게이트·템플릿 미러 전파
