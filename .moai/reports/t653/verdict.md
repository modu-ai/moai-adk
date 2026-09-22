# t653 카드 판정서 (sync-phase, 2026-09-14)

> 카드 t653 — SPEC-MOAI-GATEWAY-001 AS-4 마일스톤(이력·모델·압축·fork). 본 문서는 run-verdict.md
> 초안을 흡수해 싱크 페이즈 마감 시점의 최종 카드 판정으로 확장한 것이다. 초안 원문은
> [run-verdict.md](run-verdict.md)에 보존된다.

## Claim (주장)

카드 t653의 자동화 가능 범위 전체가 착지·검증됐고, 싱크 페이즈 산출물(verdict·§E.4·상태 전이·sync-audit)이 이 단일 커밋에서 발행됐다. 구체적으로:

1. **착지 내용** — M2 소유 thread resume(`Schema=2` + Model/CWD, legacy 거절), M3 idle 전용 모델 전환(waiting 전환 거절 + gateway 거절군), M4 compaction의 정상 요약 turn 처리(PostCompact exact-digest 대조·1회 rebase·public history 재설정·compact RPC 0회), M5 exact completedTurnID 경계 fork(경계 초과 유입 차단·거절군·fork 자식 신규 thread + 신규 프로세스 resume).
2. **실증 이관** — AS-010·011·012의 실세션 양성 실증은 acceptance.md T21에 따라 카드 t844(라이브 계측 세션)로 이관됐다. 이 판정은 실증을 PASS로 세지 않는다.
3. **AS-013 NOT-RUN 유지** — 설치 Claude Code 2.1.270에 native Agent(fork)/subtask 파라미터 부재(M1 probe)로 전제 실패 NOT-RUN이 유지되며, AS4 및 전체 지원 완료는 보류다. 일반 자식·`--fork-session` 성공으로 대체하지 않는다.
4. **상태 전이** — spec.md frontmatter `in-progress → implemented`. `completed`가 아니다: 본 SPEC은 다중 카드 시리즈(Tier L umbrella)로 t654 이후 카드가 남아 있으므로 implemented→completed 전이는 후속 싱크의 몫이다.
5. **CHANGELOG 보류 결정** — 이 시점에서 [Unreleased] 항목을 쓰지 않는다(아래 근거).

## Evidence (증거)

- **커밋 진행** (`git log 92db2cfe7..HEAD`, 이 트리에서 관측): `14dba89c5`(M3 idle 모델) → `e45f50a8d`(M4 compaction) → `64885fa06`(M5 경계 fork) → `9f3dc41e0`(M3-M6 증거 취합) → `4d6837a9d`(T21 이관 문서화) → `c71545213`(run_complete_at 발행) → 본 싱크 커밋. 각 구현 커밋마다 RED 로그가 GREEN 로그에 선행한다(m3-model-red/green, m4-compact-red/green, m5-fork-red/green).
- **싱크 페이즈 재측정** (이 실행, 이 트리):
  - `go test ./internal/gateway/receipt/ ./internal/gateway/conversation/ -count=1 -cover` → `ok`, receipt 88.9% / conversation 80.3%.
  - `go test ./internal/gateway/ -count=1 -cover -skip TestAppServerSubprocessHTTPToolContinuation` → `ok ... coverage: 91.7%`.
  - `go test ./internal/codexbridge/ -count=1` (skip 없음) → 실패 4건, 이름이 사전 존재 lifecycle_subprocess 4건과 정확히 일치 — **신규 실패 0건**.
  - `go test ./internal/codexbridge/ -count=1 -cover -skip <상기 4건>` → `ok ... coverage: 82.5%`; `-covermode=atomic` 병행 시 82.7%.
- **run-phase 측정치** (`.moai/reports/t653/m6-coverage.log`): codexbridge 83.1% / receipt 88.9% / conversation 80.3% / gateway 91.7%. Windows 교차 컴파일 exit 0, `golangci-lint run --timeout=5m` → `0 issues.`(m6-lint.log), gofmt 빈 출력.
- **증거 경로 해소 확인**: `.moai/reports/t653/` 아래 m1~m6 로그·문서 15개 전원이 이 트리에서 존재함을 `ls`로 확인했다.
- **T21 정합성**: acceptance.md:469(T21 행) ↔ spec.md HISTORY 0.12.0 ↔ progress.md §E.3 — 세 표면이 동일한 이관 내용(t844, AS-013 비이관, 전체 통과 보류)을 운반한다.
- **MX 태그**: `internal/codexbridge`·`internal/gateway`에서 @MX:WARN/REASON 11개 파일 관측 — 싱크 서브스텝 검증 통과.

## Baseline-attribution (baseline 귀속)

- 모든 측정은 worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`에서 수행됐다. run-phase 측정 baseline: 흡수 기준선 `92db2cfe7`(M1/M2 커밋 + develop `033323529` 병합; 리드가 지목한 develop 측 기준선 `74d872aaf`의 후속 병합 지점) → 마일스톤 3커밋. 싱크 재측정 baseline: HEAD `c71545213`.
- 사전 존재 환경 실패 5건(codexbridge lifecycle_subprocess 4건: `TestAuditRealTransportEOFWakesBridge`, `TestAuditCancelBeforeRealStartReplyInterruptsTurn`, `TestCanceledStartBeyondReplyDeadlineStillInterrupts`, `TestAuditCanceledRPCDoesNotExhaustSharedTransport` + gateway `TestAppServerSubprocessHTTPToolContinuation`)의 귀속 근거는 **작업 변경이 없는 커밋 트리**(`git archive HEAD` 추출본)에서의 동일 재현이다(run-phase 수행). 싱크 재측정에서도 동일 4건만 재현돼 귀속이 유지된다. 원인(샌드박스 하위 프로세스 기동 실패) 규명은 이 카드 범위 밖이며 전체 스위트 판정은 CI 소관이다.
- lint `0 issues`는 t671 클린 baseline 위의 run-phase 실측이며, 싱크 페이즈는 문서 전용 변경이므로 lint 재실행을 생략했다(컴파일 대상 무변경).
- 커버리지 재측정 차이(83.1% vs 82.5%): run 로그는 정확한 skip 표현식을 남기지 않았고, 본 재측정은 4건 이름 지정 skip으로 82.5%(atomic 82.7%)를 냈다. 동일 테스트 집합 가정 하에서의 소폭 잔차로 보지만 skip 표현식 미기록 자체를 attribution gap으로 기록한다(Gaps).

## Gaps (미검증)

- **실증 NOT-RUN (환경·권한상 이 세션에서 실행 불가 — T21로 t844 이관)**:
  - AS-010: 소유 thread로 새 MoAI process 기동 후 실제 정상 답변·합성 사실 회상
  - AS-011: 실제 turn model이 GPT-6 Astra/5.6 선택 ID와 일치하는 양성, 계정별 family 호환 증거
  - AS-012: 실제 Claude 압축(print·대화형·자동·자식)의 요청·응답·후행 hook 수집과 digest 일치, 압축 뒤 ToolSearch 후발 도구
- **AS-013 NOT-RUN (이관 대상 아님, 설계된 전제 실패)**: `--fork-session` 실분기 양성, non-fork Agent 둘 + 중첩 자식 자별 context 격리, native fork 양성 의무 — M1 probe 전제 실패로 유지(`m1-native-fork-probe.md`).
- **커버리지 목표 미달**: codexbridge 82.5~83.1%는 quality.yaml 패키지 목표(85%) 미만이다. 부족분의 상당 부분은 이 샌드박스에서 실행 불가한 4건의 환경 의존 테스트가 실행될 때만 측정 가능한 기여분이지만, 그 기여분의 크기 자체는 미측정이므로 "환경 귀속으로 해소됨"이라 단정할 수 없다. CI(전체 스위트 판정 주체)에서의 실측이 남는다.
- **CHANGELOG 미발행**: [Unreleased]에 SPEC-MOAI-GATEWAY-001 항목이 0건인 것이 시리즈의 기존 관행과 일치함을 `grep -c`로 확인했으나, 이는 "발행하지 않기로 한" 결정이지 "사용자 가시 표면이 없음이 검증됨"이 아니다(아래 결정 절 참조).
- **run-phase 커버리지 측정의 skip 표현식 미기록** — 83.1%의 정확한 재현이 불가하다(82.5~82.7% 재측정으로 대체).
- 사전 존재 환경 실패 5건의 원인 규명, 전체 스위트 판정 — 이 카드 범위 밖(CI 소관).

## Residual-risk (잔여 위험)

- `thread/resume`, `turn/start`의 실제 App Server 응답 형태 검증은 fake RPC 기준이다 — t844 실증에서 서버 계약 일치를 확인해야 한다.
- RebaseLedger는 프로세스 내 상태다 — 재시작 복원은 생성자 `appliedEpoch` 주입에 의존하며, gateway 생산 배선(t654)이 이 값을 어디서 읽어올지는 다음 카드의 설계 대상이다.
- Fork 자식의 inherited prefix는 engine에서 caller-asserted 값이다 — 원장 대조(ChainTo 기반 검증)는 gateway 계층의 책임으로 남고, 이 결합의 생산 배선 검증은 t654 몫이다.
- idle 모델 전환 시 barrier의 model 핀 저장은 turn 성공 후다 — 전환 실패 시 이전 핀 유지가 운영상 기대와 맞는지는 실증 단계 확인 대상이다.
- 사전 존재 환경 실패 5건이 CI에서도 재현되면 별도 결함 카드가 필요하다(이 카드에서는 환경 귀속).
- 커버리지 85% 목표: codexbridge가 CI에서도 85% 미만이면 목표 충족을 위한 보강 카드가 필요하다.

## CHANGELOG 결정 (보류 — 근거 기록)

`grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md` → **0** (2026-09-14, 이 트리). 시리즈의 선행 카드(t649~t653, 12커밋) 중 어느 것도 CHANGELOG 항목을 쓴 적이 없다 — 시리즈 관행은 릴리스/런처 배선 시점에 문서화하는 것이다. 카드 t653의 변경면(internal/codexbridge·internal/gateway/receipt·internal/gateway/conversation)은 전부 내부 패키지고 사용자 가시 효과(런처의 resume/model/compaction/fork 동작)는 t654의 런처 배선이 처음 노출한다. 결론: **이번 싱크에서 CHANGELOG 발행을 보류**하고, t654 런처 배선 착지 또는 시리즈 릴리스 시점에 사용자 가시 표면 기준으로 한 번에 발행한다. B12 사전-발행 grep(0건 → 발행 대상 없음 확인)은 수행했다.
