# plan.md — SPEC-SPECLINT-GATE-SIGNAL-001

## §A 맥락

- 카드 t525 (Class B, Tier M). 발췌 판정 주체: 리드 2026-09-07.
- 문제: `SPEC Lint` CI 가 develop 에서 5회 연속 실패, `0 error(s), 4,344 warning(s)`
  (run 34088825415, `d4162b368`) → exit 1. 상시 적색 게이트는 신호가 없다.
- 목표: **경고 수 감축이 아니라 신호 회복.** 새 적색이 서 있는 재고와 구분되게 만든다.
- plan 시점 기준 측정 트리 `dd1439502`: `0 error(s), 4368 warning(s)`, rc=0 (`--strict`
  없이). 원문 `.moai/reports/t525/spec-lint-baseline-dd1439502.txt`. 이 수치는 M1 이
  재도출할 때까지 참고용이다.
- 의존: t518(SPEC-SPEC-LINT-BLIND-AXES-001, develop 미착지 — merge-base 로 확인됨).
  순서 근거는 카드 갱신본(2026-09-07): t518 의 axis-1 이 `REQEntry.Widened` 선례로 권고
  확대에 들어가 경고 인구와 advisory 분포가 통째로 움직인다. 초기 카드의 "같은 4,344" 인과
  주장은 lane-5 실측으로 **철회**됐다.

## §B 알려진 이슈 / 전제

1. **인과 철회 기록**: CI 4,344 ↔ t518 4,344 는 우연 일치(다른 단위·다른 커밋). 순서
   판정(t518 먼저)은 유지, 사유는 §A 대로 교체. run 단계에서 이 사유를 어기지 않는다.
2. **수치 동결 금지**: 코퍼스 수(SPEC 디렉터 수 vs `SPEC-*/spec.md` 글롭 수 — 차이는 정확히
   SpecsDirMissingSpecFile 2건)와 경고 총수는 SPEC 착지 때마다 움직인다. 어떤 수도 교리·
   상수·AC 임계값에 못박지 않고, 측정 시점 재도출한다.
3. **t518 axis-2 수치 미확정**: t518 의 axis-2 숫자는 t518 M1 재측정까지 철회 상태다.
   방향만 확정 — 한국어 SPEC 은 "판정 대상이 아니다"이지 "틀렸다"가 아니다.
4. **프런트매터 필드명**: 스키마 문서는 `depends_on`, 코드 바인딩은 `dependencies`
   (`internal/spec/lint.go:500`). 이 SPEC 은 plan 시점에 DAG 간선을 넣지 않는다 — t518 미착지
   상태에서 넣으면 `MissingDependency` **error**(`internal/spec/lint.go:1083`)로 자기 적색.
   간선은 M3 산출물이다.
5. **CC2X ADOPT-001/002 관측(plan 시점)**: 두 디렉터는 `research.md` 만 담는다(프런트매터
   `spec_id`, `phase: research`, child_specs 열거 — 2026-05-12 작성 리서치 우산 문서).
   `discoverSPECs` 가 `SPEC-*/spec.md` 로 글롭하므로 모든 per-SPEC 규칙이 이 디렉터를 방문
   조차 않는다(`SpecsDirMissingSpecFile` 메시지 그대로). M4 가 "왜"를 먼저 기록하고 닫는다.
6. **바이너리 재측정 규율**: 모든 lint 수치는 그때-current 트리에서 그때 빌드한 바이너리로
   잰다. 오래된 바이너리나 다른 트리의 수는 귀속 금지(verification-claim-integrity §2).

## §C 사전 점검 (착수 전)

- [ ] `git fetch origin develop && git rev-parse --short origin/develop` — 측정 대상 트리 확정
- [ ] `git merge-base --is-ancestor` 로 t518 착지 여부 재확인(착지됐으면 M3 게이트 해제, 간선 즉시 추가)
- [ ] `go run ./cmd/moai spec lint --json` 규칙별 재도출 + advisory 분포 집계 → `.moai/reports/t525/` 저장
- [ ] `.moai/spec-lint-baseline.json`(또는 M2 에서 확정할 경로)이 비어 있고 템플릿에 없음 확인
- [ ] 이 SPEC 의 기준선 미포함 상태에서 `--strict` rc=1 재현(REQ-SLGS-002 전제)

## §D 제약

1. **건드리는 표면**: `internal/spec/lint.go`(또는 신규 `internal/spec/lint_baseline.go`),
   `internal/cli/spec_lint.go`(플래그), `.github/workflows/spec-lint.yml`(배선), 기준선
   파일, 테스트. **경고 부채(SPEC 문서 대량 수정)와 advisory 판정 기준(t518 소관)은 아니다.**
2. **config 섹션 금지**: `.moai/config/sections/` 에 새 파일을 만들지 않는다 — Template-First
   미러 + template-neutrality(C1-C8) + `moai update` 의 `.moai/config` 통째 삭제
   (CLAUDE.local.md §2.3) 삼중 위험. spec.md §2.1 결정.
3. **기준선 파일 위치**: `.moai/config/` 밖이며 템플릿에 미러하지 않는다. 후보
   `.moai/spec-lint-baseline.json`(관리 대상 뿌리 밖이라 update 에 안 지워지고, 배포도 안
   된다). M2 에서 확정.
4. **하위 호환**: 새 플래그 미지정 시 `moai spec lint` 동작 불변. `--strict` 는 유지(다른
   호출자 보존). 기존 exit code 계약(0/1/2/3, `internal/cli/spec_lint.go:39-43`) 유지.
5. **오염 금지**: 기준선 비교 로직은 error-severity 경로(REQ-SLGS-009)를 절대 약화시키지
   않는다. 기존 `HasErrors()` error 단락 먼저, 기준선 판정은 그 다음.
6. **결정적 표면 결정론**: 기준선 JSON 은 키 정렬로 플랫폼 무관 바이트 동일성을 유지해야
   diff/review 가 가능하다(감사 가능성의 전제).

## §E 자가 검증

- [ ] `go test ./internal/spec/... ./internal/cli/...` — 대상 패키지 통과(전체 스위트는
      CI 몫; CLAUDE.local.md §4)
- [ ] `golangci-lint run internal/spec/... internal/cli/...` 청결
- [ ] 합성 경고 주입 시나리오(AC-SLGS-005)를 로컬에서 실제 실행 — 붉어짐/복귀 양쪽 관측
- [ ] 서 있는 재고 전체 run 이 기준선으로 rc=0(AC-SLGS-006), 경고 총수 출력 확인
- [ ] 기준선 파일 미수정 불변(AC-SLGS-007) — 감소 시나리오에서 파일 해시 불변 관측
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` — 기준선 직렬화가 크로스 플랫폼인지

## §F 마일스톤

되돌릴 수 없는 결정이 앞에 온다(M1 판정 → M2 기제·운반체 → M3 배선·잠금 → M4 기계적 부수).

### M1 — 판정: (a)/(b)/(c) 를 증거로 못박는다 (가장 되돌리기 어려운 결정)

1. 그때-current origin/develop 트리에서 `go run ./cmd/moai spec lint --json` 재도출 —
   규칙별 × severity × advisory 3차원 집계. 명령 전문 + HEAD SHA 와 함께
   `.moai/reports/t525/census-<sha>.txt|json` 저장(REQ-SLGS-001).
2. (b) 구조 입증: 비-advisory 재고 ≥ 1 상태에서 `--strict` rc=1, 그리고 mutation — 재고를
   기준선으로 흡수하는 판정 경로가 있으면 rc=0 으로 뒤집히는지 대조(AC-SLGS-002).
3. (c)는 t518 소관, (a)는 실재하나 지연 — 각 축 판정 + 명령 + 출력을 M1 판정 기록으로
   정리(REQ-SLGS-002). 수치 동결 없음(REQ-SLGS-003).

산출: M1 판정 기록(`.moai/reports/t525/verdict.md`). **kickoff 승인 전제물이다.**

### M2 — 기제: 규칙별 기준선 래칫 (kickoff 에서 모양 최종 확정)

1. 기준선 파일 스키마 + 직렬화(결정론 정렬), 위치 `.moai/spec-lint-baseline.json` 확정.
2. 비교 정책: error 무조건 적색(REQ-SLGS-009) → 규칙별 비-advisory 경고 수 초과 적색 +
   델타 출력(REQ-SLGS-006) → 이하 통과 + 총수 출력(REQ-SLGS-004).
3. CLI 플래그: `--baseline <path>`, `--update-baseline`(명시적 재기준, SHA·사유 기록
   출력)(REQ-SLGS-005, REQ-SLGS-008). 미지정 시 동작 불변.
4. 감소 방향: 통과 + 감소분 출력, 기준선 파일 불변(REQ-SLGS-007).
5. 테스트: `t.TempDir()` 코퍼스로 합성 경고 주입/제거(AC-SLGS-005), 에러 독립성
   (AC-SLGS-009), 감소 불변(AC-SLGS-007), 재기준 감사(AC-SLGS-008).

파일: `internal/spec/lint.go` 또는 신규 `internal/spec/lint_baseline.go` + `_test.go`,
`internal/cli/spec_lint.go`, 기준선 파일.

### M3 — 배선과 잠금: CI, t518, 재기준

1. `.github/workflows/spec-lint.yml` — `go run ./cmd/moai spec lint --baseline
   .moai/spec-lint-baseline.json`(kickoff 승인 모양에 따름)(REQ-SLGS-010).
2. 기준선 최초 산출 + 체크인(그때-current develop 기준, SHA 기록).
3. t518 잠금 점검 절차: 재기준 단계 앞 `merge-base --is-ancestor` 확인 기록(REQ-SLGS-011).
   t518 착지 전 (a)축 커밋 없음.
4. t518 착지 후: DAG 간선 `dependencies: [SPEC-SPEC-LINT-BLIND-AXES-001]` 추가 + 게이트된
   재기준 1회(§2.3 절차 경유).

파일: `.github/workflows/spec-lint.yml`, 이 SPEC 프런트매터, 기준선 파일.

### M4 — 부수: CC2X ADOPT-001/002

1. "왜 spec.md 가 없는가" 기록 — plan 관측(§B.5)을 run 단계에서 확인하고, 의도 안치인지
   유실인지 판명한다(REQ-SLGS-012).
2. 답에 따라: 리서치 문서가 정체라면 `.moai/reports/` 계열로 이전(SPEC vs Report 분류) 또는
   spec.md 보태기 — 둘 중 하나로 발견 2건 소멸 확인(AC-SLGS-012).

파일: `.moai/specs/SPEC-V3R4-CC2X-ADOPT-001/`, `-002/`(또는 이전처).

## §G 안티패턴

- **경고를 advisory 로 둔갑시켜 없앤다** — (c)축 수리를 이 SPEC 이 몰래 하는 모양. advisory
  경계는 t518 의 몫이며, 여기서 만지면 기준선의 전제가 흔들린다.
- **숫자를 교리로 못박는다** — "4,344"나 "4,368"을 상수·AC 임계·문서 기준값으로 쓴다.
  측정 시점 재도출만이 근거다.
- **재기준을 습관으로 쓴다** — 붉은 게이트를 재기준으로 침묵시키면 기준선은 장식이 된다.
  재기준 커밋에는 사유가 반드시 붙고, 감사에서 사유 없는 재기준은 적발 대상이다.
- **kickoff 없이 모양을 정한다** — (i)-(iv) 기제 선택은 운영자 결정이다. plan-auditor PASS가
  권고를 승인으로 바꾸지 않는다.
- **error 경로를 기준선 뒤로 순서를 바꾼다** — 기준선이 진짜 에러를 가리는 순간 이 게이트는
  거짓 초록 기계가 된다.

## §H 상호참조

- SPEC-SPEC-LINT-BLIND-AXES-001 (t518) — 선행 잠금 대상. 착지 시 `dependencies:` 간선 추가.
- SPEC-SPECLINT-GITBLIND-001 — 같은 CI 잡의 이전 수리(눈멂 관측 가능화). `fetch-depth: 0` +
  main ref fetch 배선의 전제.
- SPEC-V3R4-SPECLINT-DEBT-001/002 — (a)축 부채의 역사적 소관(이 SPEC 은 인수하지 않는다).
- CLAUDE.local.md §2/§2.3 — Template-First, `moai update` 소거 기제.
- `.claude/rules/moai/core/verification-claim-integrity.md` — 모든 수치 귀속 규율.
