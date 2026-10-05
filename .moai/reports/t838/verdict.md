# t838 판정문 — t707 관측성 패치 검증 + 라이브 계측 런북

날짜: 2026-09-14 · 브랜치: `WT-receipt-debug-runbook` · 기반: develop `d416f8162`(ff 흡수)
Class B(실행은 운영자) · 수행: lane-1

## Claim (주장)

1. t707의 `observability.patch`(v1)는 **적용 불가**다 — `git apply --check`가 line 26 corrupt를 보고하고, hunk 헤더 줄 수가 본문과 불일치하며, 추가된 코드가 쓰는 `fmt`/`os` import가 패치에 없다(실패해도 컴파일 오류). t707 판정서 갭("검증되지 않은 패치")의 실체 확인.
2. v2 패치(`.moai/reports/t838/observability.patch`)는 `d416f8162`에서 **적용·컴파일·테스트·발화가 모두 실측 확인**됐다. 적용된 작업 diff에서 기계 생성해 hunk 계수가 보장된다.
3. 운영자 런북(`.moai/reports/t838/runbook.md`)은 다음 라이브 CauseChain 400 발생 시 1회 계측의 전 절차(적용→재빌드·설치→재현→stderr 포착→원복→증거 기록)와 출력 판독 가이드를 담는다. 라이브 재현 실행 자체는 운영자 몫으로 남긴다.

## Evidence (증거)

- **v1 부적용 실측**: `git apply --check .moai/reports/t707/observability.patch` → `error: corrupt patch at line 26` (exit 128), `git apply` 동일 실패. 소스 대조로 `fmt`/`os` import 부재 확인(`internal/gateway/translate/receipt_history.go` import 블록).
- **v2 생성 방식**: 적용된 작업 트리에서 `git diff`로 생성 → hunk 줄 수 기계 보장. 헤더에 검증 기록 명기.
- **v2 적용 실측**: `git apply --check` ok → `git apply` → `go build ./internal/gateway/...` ok, `go vet ./internal/gateway/translate/` ok, `go test -count=1 ./internal/gateway/translate/ -timeout 120s` → **ok 1.849s**.
- **발화 실측(verbatim)**: 적용 상태에서 `MOAI_RECEIPT_DEBUG=1 go test -run TestReceiptHistoryRejectsMidChainTruncationWithGuidance -v` →

  ```
  moai-receipt-divergence: observations=2 candidates=3 first_unmatched_boundary=1 root_unmatched=false observations_items_first=1
  --- PASS: TestReceiptHistoryRejectsMidChainTruncationWithGuidance (0.07s)
  ```

  중간 경계 절단 케이스에서 `first_unmatched_boundary=1`(뿌리는 일치)로 의도대로 판독된다. 측정 후 `git restore`로 원복, `git status --porcelain` 추적 수정 0 확인 — **패치는 어느 커밋에도 없음**(커밋된 2파일은 문서뿐).
- **런북 절차**: 적용→`make build && make install`(LDFLAGS 포함·맨손 go install 금지·exit 0 확인)→`MOAI_RECEIPT_DEBUG=1` gpt 세션 재현(Edit 직후 첫 요청, t707 4고유 사례 공통 패턴)→stderr 1행 포착→원복→`.moai/reports/t707/` 아래 증거 기록. 판독 가이드: `root_unmatched=true`→뿌리부터 재생성(t707 갭 3 후보), `first_unmatched_boundary>0`→특정 경계 재인코딩 suspect.

## Baseline-attribution (baseline 귀속)

- 모든 측정은 본 레인 워크트리 `.claude/worktrees/t838`(branch `WT-receipt-debug-runbook`, develop `d416f8162` 기반)에서 이번 세션에 실행한 명령과 출력.
- v1 부적용 판독도 같은 트리의 커밋된 `t707/observability.patch` 원문 대상.

## Gaps (미검증)

1. **라이브 400 재현 미수행** — 운영자 실행 몫. 런북이 준비물·절차·판독만 검증했다.
2. **stderr 도달 경로의 세부** — 게이트웨이를 호스팅하는 moai 프로세스의 stderr에 출력되는 것까지가 코드상 보장(`fmt.Fprintf(os.Stderr, …)`)이고, 운영자 세션 구성별 표시 위치(CC 터미널 vs 로그 파일)는 이 세션에서 확인하지 않았다.
3. 디버그 분기는 `checkObserved` 실패 경로만 감싼다 — `observations` 단계(분류된 에러) 거부에는 출력이 없다. 라이브 400이 어느 쪽인지는 t707 판정상 CauseChain=checkObserved 경로라 커버되나, 재현이 다른 원인이면 무출력이 정상이다.

## Residual-risk (잔여 위험)

- 런북의 `make install`은 운영자 바이너리를 교체한다 — §11 규약(rm+cp clean 재설치, exit 137 시 재시도)을 런북에 명시했으나, 교체 창에서 다른 moai 세션이 돌면 영향을 받는다.
- v2 패치는 `d416f8162` 기준이라, 계측 시점의 develop이 진행돼 있으면 `git apply`가 컨텍스트 불일치로 실패할 수 있다 — 그 경우 런북 헤더의 검증 base를 보고 판단하거나 재생성 요청.
- 디버그 출력은 구조 사실만 담지만 세션 크기·경계 수가 노출되므로 외부 게시 금지(런북 §6).
