# acceptance.md — SPEC-SPECLINT-GATE-SIGNAL-001

모든 AC 는 명령 + 기대 관측치 + 실패 모양을 갖는다. 빈 집합 위의 자신 있는 판정 금지 —
가드형 AC 는 반드시 뮤테이션(잡는 모양·놓치는 모양 양쪽)을 보인다.

## M1 — 판정

### AC-SLGS-001 — 인구통계가 측정 시점에 재도출된다 (maps REQ-SLGS-001)

- Given 그때-current `origin/develop` 트리
- When `go run ./cmd/moai spec lint --json` 을 돌려 규칙별 × severity × advisory 집계를
  `.moai/reports/t525/` 에 저장하면
- Then 기록에 명령 전문과 집계 시점 HEAD SHA 가 함께 있고, SHA 가 `git rev-parse --short
  HEAD` 와 일치한다
- 실패 모양: 수치만 있고 SHA·명령이 없거나, SHA 가 다른 트리의 값을 가리키면 미달.

### AC-SLGS-002 — (b) 가 구조적으로 입증된다 (maps REQ-SLGS-002)

- Given 비-advisory 경고 재고 ≥ 1 인 코퍼스
- When `go run ./cmd/moai spec lint --strict; echo rc=$?` — rc=1, 요약 줄 `0 error(s),
  N warning(s)`(N>0). 그리고 mutation: 같은 코퍼스를 기준선 흡수 경로로 판정하면 rc=0
- Then 양쪽 관측이 모두 기록된다 — "재고만으로 적색"과 "재고 흡수 시 초록"의 대조
- 실패 모양: `--strict` 가 이미 rc=0 이면 (b) 전제가 무너진 것이므로 M1 판정을 다시 쓴다
  (침묵 통과 아님).

### AC-SLGS-003 — 세 축 판정이 각각 증거를 갖는다 (maps REQ-SLGS-002)

- Given M1 판정 기록
- When (a)/(b)/(c) 각 항목을 읽으면
- Then 매 항목에 명령 + 관측 출력(또는 소관 SPEC 참조)이 대응된다
- 실패 모양: "구조상 그렇다"류의 근거 없는 기술이 하나라도 있으면 미달.

### AC-SLGS-004 — 수치 동결이 없다 (maps REQ-SLGS-003)

- Given 이 SPEC 의 문서·코드·테스트 전체
- When `grep -rn "4344\|4,344\|4368\|4,368"` (또는 동치 검사)를 하면
- Then 역사적 인용(출처·트리 명시) 외에 임계값·상수·기대값으로 쓰인 정수가 없다
- 실패 모양: 기대 경고 수를 숫자로 못박은 테스트 단언이 발견되면 미달.

## M2 — 기제

### AC-SLGS-005 — 새 경고가 붉어지고, 지우면 돌아온다 (maps REQ-SLGS-004, REQ-SLGS-005, REQ-SLGS-006 — 핵심)

- Given 기준선이 체크인된 상태
- When `t.TempDir()` 코퍼스에 기준선에 없는 비-advisory 위반 1건을 합성 주입하고 게이트를
  돌리면 — exit 1, 출력에 **어느 규칙이 +1** 인지가 보인다. 그 위반을 지우고 다시 돌리면 —
  rc=0 복귀
- Then 붉어짐과 복귀 양쪽이 관측된다(크로스 프로세스 — 실제 CLI 실행)
- 실패 모양: 주입했는데 초록(공허 가드 — 기준선 비교가 작동 안 함), 또는 주입 없이 적색
  (거짓 양성). 어느 쪽이든 미달.

### AC-SLGS-006 — 서 있는 재고는 초록이다 (maps REQ-SLGS-004, REQ-SLGS-005)

- Given 그때-current develop 전체 코퍼스 + 체크인 기준선
- When `go run ./cmd/moai spec lint --baseline <path>; echo rc=$?`
- Then rc=0 이고 출력에 경고 총수가 그대로 보인다 (재고가 사라진 게 아니라 흡수됐음이 관측됨)
- 실패 모양: 미변경 코퍼스에서 rc=1 이면 기준선 산출이 틀렸거나 비교가 과잉 적발.

### AC-SLGS-007 — 감소는 통과하고, 기준선은 스스로 줄지 않는다 (maps REQ-SLGS-007)

- Given 기준선 통과 상태
- When 스크래치 코퍼스에서 위반 하나를 실제로 고쳐 발견이 줄면 — rc=0 + 감소분 출력.
  이때 기준선 파일의 해시를 before/after 로 비교하면 불변
- Then 통과 + 개선 표시 + 기준선 무변경 세 가지가 모두 관측된다
- 실패 모양: 기준선 파일이 함축적으로 다시 쓰였으면(해시 변화) 미달 — 조용한 수축은
  재기준 감사를 우회한다.

### AC-SLGS-008 — 재기준은 명시적이고 기록을 남긴다 (maps REQ-SLGS-008)

- Given 기준선 파일
- When 명시적 재기준(`--update-baseline` 또는 동치 절차)을 돌리면 — 파일이 그때-current
  재측정값으로 갱신되고, 트리 SHA·날짜·사유 기록이 출력/커밋에 남는다. 반대 mutation:
  명시 단계 없이 게이트를 감소 상태로 돌리면 파일은 불변(AC-SLGS-007 과 동일 관측)
- Then 재기준 전후 diff 가 git 에서 검토 가능하다
- 실패 모양: 기록 없는 갱신 경로가 존재하면 미달.

### AC-SLGS-009 — error 는 기준선과 무관하게 막는다 (maps REQ-SLGS-009)

- Given 기준선 통과 상태의 코퍼스
- When error-severity 위반 1건(예: 현대-era SPEC 에서 Out of Scope 섹션 삭제)을 스크래치
  코퍼스에 주입하고 기준선 있이 게이트를 돌리면 — rc=1, 출력에 error 가 보인다. 지우면
  rc=0
- Then 기준선이 에러를 가리지 않는다는 대조가 관측된다
- 실패 모양: 기준선 통과가 error 를 흡수하면 이 기제는 거짓 초록 기계다 — 즉시 미달.

## M3 — 배선과 잠금

### AC-SLGS-010 — t518 잠금이 기록으로 지켜진다 (maps REQ-SLGS-011)

- Given 이 SPEC 의 run 단계 커밋 목록
- When `git log` 로 SPEC 문서 대량 수정 커밋을 찾고, 재기준 기록의
  `merge-base --is-ancestor` 점검 줄을 읽으면
- Then t518 착지 전 커밋 중에는 (a)축 부채 상환(경고 감축 목적의 SPEC 문서 대량 수정)이
  없고, 기준선 최초 산출 기록에 측정 트리 SHA 가 명시되어 있다. t518 착지 후에는 간선
  추가 + 게이트된 재기준 1회가 기록된다
- 실패 모양: t518 착지 전에 부채 상환 커밋이 있거나, 재기준이 점검 기록 없이 일어나면 미달.

### AC-SLGS-011 — CI 가 새 기제로 돌고 통과 로그에 신호가 남는다 (maps REQ-SLGS-010)

- Given `.github/workflows/spec-lint.yml`
- When 배선 줄을 읽으면 kickoff 승인 모양의 기제(권고: `--baseline`)가 보이고, 착지 후
  develop push 의 녹색 run 로그에 경고 총수(또는 기준선 대비 상태) 줄이 있다
- Then 게이트가 초록인 동안에도 재고 규모가 관측 가능하다
- 실패 모양: 배선이 여전히 벌거벗은 `--strict` 이거나, 녹색 run 의 로그에 상태 줄이
  없으면 미달.

## M4 — 부수

### AC-SLGS-012 — CC2X 두 디렉터가 "왜"와 함께 닫힌다 (maps REQ-SLGS-012)

- Given SPEC-V3R4-CC2X-ADOPT-001/002
- When M4 기록을 읽으면 "왜 spec.md 가 없는가"의 답이 있고(plan 관측: research.md 만
  있는 리서치 우산 문서 — run 단계에서 의도성 확인), 답에 따른 종결 조치가 이뤄져 있다.
  그 후 `go run ./cmd/moai spec lint` 출력에서 `SpecsDirMissingSpecFile` 이 0건이다
- Then 두 디렉터는 spec.md 를 갖거나 `.moai/specs/` 밖으로 이전되어 있다
- 실패 모양: 발견만 사라지고 답이 없거나, lint.skip 등 다른 방법으로 발견을 억눌렀으면 미달.

## 커버리지 표 (REQ → AC)

| REQ | AC |
|-----|----|
| REQ-SLGS-001 | AC-SLGS-001 |
| REQ-SLGS-002 | AC-SLGS-002, AC-SLGS-003 |
| REQ-SLGS-003 | AC-SLGS-004 |
| REQ-SLGS-004 | AC-SLGS-005, AC-SLGS-006 |
| REQ-SLGS-005 | AC-SLGS-005 (플래그 경유), AC-SLGS-006 |
| REQ-SLGS-006 | AC-SLGS-005 |
| REQ-SLGS-007 | AC-SLGS-007 |
| REQ-SLGS-008 | AC-SLGS-008 |
| REQ-SLGS-009 | AC-SLGS-009 |
| REQ-SLGS-010 | AC-SLGS-011 |
| REQ-SLGS-011 | AC-SLGS-010 |
| REQ-SLGS-012 | AC-SLGS-012 |

## 경계 사례

- 기준선 파일 부재 + `--baseline` 지정 → 명확한 인자 오류(exit 3 계열) 또는 안전한 오늘의
  동작 폴백 — M2 에서 정하고 문서화. 침묵 폴백 금지.
- 기준선에 없는 **새 규칙**이 처음 발화 → 기록치 0 대비 증가 = 적색(새 신호다).
- 기준선에만 있고 더 이상 발화하지 않는 규칙 → 감소 방향(AC-SLGS-007), 통과 + 표시.
- 플랫폼별 줄바꿈/키 정렬로 기준선 diff 가 시끄러움 → 결정론 직렬화(§D.6)로 예방.

## Definition of Done

- [ ] AC-SLGS-001..012 전부 관측 근거와 함께 충족
- [ ] develop push 의 `SPEC Lint` run 이 녹색이고, 로그에 재고 상태 줄 존재
- [ ] 합성 경고 주입/제거 시나리오가 로컬에서 재현됨(AC-SLGS-005 양방향)
- [ ] t518 착지 여부와 재기준 게이트 상태가 progress.md 에 기록됨
- [ ] `go test ./internal/spec/... ./internal/cli/...` 통과 + `golangci-lint run` 청결
- [ ] M4 종결 후 `SpecsDirMissingSpecFile` 0건
