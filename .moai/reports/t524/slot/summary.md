# t524 internal/cli 슬롯 결과 (리드 1회 승인분)

- 워크트리: `.claude/worktrees/t524`, 브랜치 `WT-auditor-req-shorthand`, 측정 HEAD `fcb392deb`
- 선택자: `-run '^TestPlanAudit(D7|D8|Traceability)'` — 대상 함수 19개(D7 4 · D8 4 · D7D8 1 · Traceability 10). `TestPlanAuditorFailureClassifiesAsInconclusive` 는 `^TestPlanAudit(` 뒤에 `or` 가 와서 제외된다.
- 모든 실행 앞에 `ps -axo pid,args` 로 다른 internal/cli 컴파일을 쟀다. 판별식은 `/compile .*internal/cli( |$)`, `cli\.test( |$)`, `go (vet|test|build) .*internal/cli` 이고, 대조군은 `(^|/| )claude( |$)` 다.

| 단계 | 사전 확인 (외부 cli / 대조군) | 명령 결과 | 증거 |
|---|---|---|---|
| 1) vet | 0 / 20 | `go_vet_exit=0`, 출력 0바이트 | `precheck-vet.txt`, `vet-internal-cli.{txt,exit}` |
| 2) -run 3묶음 | 0 / 20 | `go_test_exit=0`, PASS 19 · FAIL 0 | `precheck-test.txt`, `test-green.{txt,exit}` |
| 3a) 뮤턴트 m1 (축약 전개 루프 제거, 두 사본 동일 적용) | 0 / 20 | `go_test_exit=1`, FAIL 3 · PASS 16 | `ps-before-m1.txt`, `test-m1.{txt,exit}` |
| 3b) 뮤턴트 m2 (제목형 정의 `defhead` 수집 제거, 두 사본 동일 적용) | 0 / 20 | `go_test_exit=1`, FAIL 1 · PASS 18 | `ps-before-m2.txt`, `test-m2.{txt,exit}` |
| 3c) 복원 후 GREEN | 0 / 20 | `go_test_exit=0`, PASS 19 · FAIL 0 | `ps-before-final.txt`, `test-final-green.{txt,exit}` |

## m1 에서 FAIL 한 테스트

- `TestPlanAuditTraceability_ShorthandMappingIsCovered` — `UNCOVERED: REQ-FIXA-002`
- `TestPlanAuditTraceability_HeadingDefinitionsExposeRealGap` — `UNCOVERED: REQ-FIXD-002` 가 추가로 나왔다
- `TestPlanAuditTraceability_InlineACsWithoutAcceptance` — `UNCOVERED: REQ-FIXS-002` 가 추가로 나왔다

`BareNumbersOutsideListDoNotExpand` 는 m1 에서도 PASS 했다. 이 테스트는 전개하지 않아야 할 자리를 지키는 테스트라, 전개 자체를 없애는 뮤턴트로는 빨갛게 되지 않는다. 과잉 전개 방향은 이 슬롯에서 뮤턴트로 재지 않았다.

## m2 에서 FAIL 한 테스트

- `TestPlanAuditTraceability_HeadingDefinitionsExposeRealGap` — `COLLECTED: 0` · `GAP` 이 출력되고, 기대한 `UNCOVERED: REQ-FIXD-003` 은 나오지 않았다

## 복원 바이트 비교 (`cmp` exit)

| 시점 | 템플릿 vs 저장본 | 로컬 vs 저장본 | 템플릿 vs HEAD | 로컬 vs HEAD |
|---|---|---|---|---|
| m1 복원 후 | 0 | 0 | 0 | 0 |
| m2 복원 후 | 0 | 0 | 0 | 0 |

- 복원 전 해시: `pre-mutant-sha.txt`
- 복원 뒤 확인 결과:
  - `git status --short` 로 두 사본, `catalog.yaml`, `plan-auditor.toml` 을 봤고 출력이 없었다.
  - `make agents-emit-check` exit 0 (`agents-emit-check-after-restore.txt`)

## 커밋하지 않은 것

- `ps-before-*.txt` 원본 덤프는 커밋하지 않았다. 머신 전체 프로세스 명령줄이 들어 있어 다른 세션의 인자(토큰 포함 가능)가 섞일 수 있기 때문이다.
- 표의 사전 확인 계수는 그 덤프에 위 판별식을 적용해 이 슬롯에서 잰 값이다. 원본은 워크트리에만 남아 있어 추적 경로로 재검증할 수 없다(잔여 위험).
- `precheck-vet.txt`, `precheck-test.txt` 는 계수만 담고 있어 커밋했다.

뮤턴트는 이 워크트리 파일에만 적용했다. 커밋은 하지 않았고, 복원을 확인하기 전까지 다른 커밋도 없었다.
