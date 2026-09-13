# t653 run-phase 판정 초안 (2026-09-14)

## Claim (주장)

t653의 자동화 가능 부분 전체가 구현·검증돼 착지했다: M3 idle 전용 모델 전환(대기 중 전환 차단 + gateway 거절군), M4 compaction의 정상 요약 turn 처리(PostCompact exact-digest 대조·1회 rebase·public history 재설정·compact RPC 0회), M5 exact completedTurnID 경계 fork(경계 초과 유입 차단·거절군·fork 자식 신규 thread와 신규 프로세스 resume). M6 증거(테스트·커버리지·Windows 빌드·lint 0·gofmt)는 모두 이번 실행에서 실측했다. 반면 **실증 항목은 전부 미수행**이며, 특히 M1의 native fork 전제 실패 NOT-RUN이 유지되므로 **AS-013 전체 지원 완료는 계속 보류된다**. AS-011의 실제 turn model 일치, AS-012의 실제 Claude 압축 수집, AS-010의 새 프로세스 실세션 회상은 이 판정에서 PASS로 세지 않는다.

## Evidence (증거)

- 커밋: `14dba89c5`(M3), `e45f50a8d`(M4), `64885fa06`(M5) — 각 커밋마다 RED 로그가 GREEN 로그에 선행한다.
  - M3 RED: `TestIdleModelChangeRepinsAndCarriesNewModel` FAIL — "idle model change rejected" (`.moai/reports/t653/m3-model-red.log`)
  - M4 RED: `CompactBase`/`NewRebaseLedger` 등 미정의 컴파일 실패 (`.moai/reports/t653/m4-compact-red.log`)
  - M5 RED: `ForkAt`/`ChainTo`/`Request.Fork` 미정의 컴파일 실패 (`.moai/reports/t653/m5-fork-red.log`)
- GREEN: `go test` 4패키지 `ok` — codexbridge 83.1% / receipt 88.9% / conversation 80.3% / gateway 91.7% 커버리지 (`.moai/reports/t653/m6-coverage.log`)
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `golangci-lint run --timeout=5m` → `0 issues.` exit 0 (`.moai/reports/t653/m6-lint.log`)
- `gofmt -l` → 빈 출력
- 신규 거절군 테스트: waiting 모델 전환, bare `gpt-6`, 타 provider adapter 진입, foreign/stale/ambiguous/duplicate PostCompact, substring-only summary, 미지·영·사이클·불일치 fork 경계, 미지 원본, Fork+Resume 조합, 경계 이후 사실 유입 차단
- 상세: `.moai/reports/t653/m6-evidence-summary.md`

## Baseline-attribution (baseline 귀속)

- 모든 측정은 worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`의 이번 실행에서 수행됐다. 커밋 진행: `92db2cfe7`(흡수 기준선) → `14dba89c5` → `e45f50a8d` → `64885fa06`.
- 사전 존재 실패 5건(codexbridge lifecycle_subprocess 4건 + gateway `TestAppServerSubprocessHTTPToolContinuation`)은 **작업 변경이 없는 커밋 트리**(`git archive HEAD` 추출본)에서 동일 재현해 이 diff 이전 환경 의존으로 귀속했다. 판정에서 제외한 근거는 위 재현이다.
- lint `0 issues`는 t671이 만든 클린 baseline 위에서의 실측이며, 이번 변경으로 신규 이슈 0건을 유지했다.

## Gaps (미검증)

- **실증 NOT-RUN (환경·권한상 이 세션에서 실행 불가)**:
  - AS-010: 소유 thread로 새 MoAI process 기동 후 실제 정상 답변·합성 사실 회상
  - AS-011: 실제 turn model이 GPT-6 Astra/5.6 선택 ID와 일치하는 양성, 계정별 family 호환 증거
  - AS-012: 실제 Claude 압축(print·대화형·자동·자식)의 요청·응답·후행 hook 수집과 digest 일치, 압축 뒤 ToolSearch 후발 도구
  - AS-013: `--fork-session` 실분기 양성, non-fork Agent 둘 + 중첩 자식 자별 context 격리
- **M1 전제 실패 NOT-RUN 유지**: 설치 Claude Code 2.1.270에 native Agent(fork)/subtask 파라미터 부재 — native fork 양성 의무는 유지되고 AS4/전체 지원 완료는 보류다. 일반 자식·`--fork-session` 성공으로 대체하지 않는다.
- 전체 스위트 판정은 CI 소관(로컬 전체 실행 금지 — 변경 패키지 + gateway만 측정).
- 사전 존재 환경 실패 5건의 원인 규명(하위 프로세스 기동 실패)은 이 카드 범위 밖.

## Residual-risk (잔여 위험)

- `thread/resume`, `turn/start`의 실제 App Server 응답 형태 검증은 fake RPC 기준이다 — 실증에서 서버 계약 일치를 확인해야 한다.
- RebaseLedger는 프로세스 내 상태다 — 재시작 복원은 생성자 `appliedEpoch` 주입에 의존하며, gateway 생산 배선(t654)이 이 값을 어디서 읽어올지는 다음 카드의 설계 대상이다.
- Fork 자식의 inherited prefix는 engine에서 caller-asserted 값이다 — 원장 대조(ChainTo 기반 검증)는 gateway 계층의 책임으로 남고, 이 결합의 생산 배선 검증은 t654 몫이다.
- idle 모델 전환 시 barrier의 model 핀이 저장되는 시점은 turn 성공 후다 — 전환 시도가 실패로 끝나면 이전 핀이 유지되는데, 이것이 운영상 기대와 맞는지는 실증 단계에서 확인 대상이다.
- 사전 존재 환경 실패 5건이 CI에서도 재현되면 별도 결함 카드가 필요하다(이 카드에서는 환경 귀속).
