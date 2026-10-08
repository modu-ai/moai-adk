---
id: SPEC-RECEIPT-REUSE-001
title: "plan — 감사 영수증 인스턴스 간 재사용 차단"
created: 2026-10-07
updated: 2026-10-07
---

# SPEC-RECEIPT-REUSE-001 — Implementation Plan

> Stateless artifact — lifecycle 상태는 spec.md frontmatter만 운반한다(spec-frontmatter-schema.md § Artifact Statelessness).

## §A Context

- **트리/브랜치**: 카드 워크트리 `.moai/worktrees/t1562`, 브랜치 `WT-receipt-reuse`, HEAD `f97edcc55` (develop 기반 git-flow — 커밋은 이 트리에서만, primary 금지).
- **실행 모드**: `cycle_type=tdd` (quality.yaml `constitution.development_mode: tdd` + 카드 "RED 재현 우선" 명시).
- **Tier**: M (3 산출물 + progress.md + decision-index.md) — 단일 결함 수리, 영향 파일 2-3개 + 테스트.
- **상위 SPEC**: SPEC-CODEX-AUDIT-GATE-AXES-001 (completed) — 이 가드의 모본. 선행 수리 t1544(PR #1763)는 SPEC 없이 카드로 착지해 `depends_on` 없음.
- **결함 1줄 요약**: 파생 키 시작 표식의 keep-earliest 조기 반환(guard.go:116)이 세션+역할의 최초 StartedAt을 영구 경계로 남겨, 이전 인스턴스 수명 동안 주조된 영수증이 다음 인스턴스의 PASS를 증명한다 — 상세와 file:line은 spec.md §1.1.

## §B Known Issues

- **B3 하위경계(C-HRA-008)**: `internal/hook` 금지 어휘 — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` 0 일치 유지.
- **B5 CI 3-tier**: 기저 GREEN/RED 판별 — §C 사전비행에서 영향 패키지 기저를 측정하고, M1의 RED는 "이 SPEC이 창출하는 테스트가 유일한 적색"임을 확인한다(기존 테스트 오염 금지).
- **B7 CWD 해상**: `auditReceiptScope`가 이미 `TreeRootFromCWD`로 처리 — 손대지 않는다(PRESERVE).
- **B8 작업 트리 위생**: 런타임 관리 파일(`.moai/state/`, `.moai/harness/`) 수정 금지; 커밋은 명시 경로식으로만.
- **테스트 간 충돌 없음 확인(M1)**: `TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor`는 r1·r2 **상이 인용**이고 RED 재현 테스트는 **동일 r1 재인용**이라 시나리오가 서로소다 — M1에서 두 테스트가 공존함을 확인한다.

## §C Pre-flight

```bash
git rev-parse --short HEAD        # f97edcc55 — 다르면 정지·보고
git branch --show-current         # WT-receipt-reuse
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m ./internal/hook/... ./internal/auditreceipt/... 2>&1 | tail -5   # 기저 측정 (신규 vs 기존 판별용)
go test -count=1 ./internal/hook/... ./internal/auditreceipt/...                                 # 기저 GREEN 측정
```

## §D Constraints

### §D.1 PRESERVE (건드리지 않는 것)

- `internal/hook/wsr_audit_receipt_tree_test.go` 전부 (트리 루트 해상 테스트)
- `internal/auditreceipt`의 거부 종류(kind) 기계, `StoreRoot`, `TreeRootFromCWD`, `sanitizeKey`
- guard.go의 FAIL 소비 경로, 재진입(`StopHookActive`) 경로, 거부 기록 갱신 규칙(원인 교체·최초 시각 보존)
- 기존 테스트 전부 — 특히 `TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor`·`TestSubagentStop_BackgroundSpawnRecycledReceiptStillRefused`·`TestSubagentStop_BackgroundSpawnAuditorPassIsProvable`·`TestSubagentStop_AgentIDFailKeepsBackgroundMarker`·`TestAuditReceiptGuard_NonRequiredTreesAreInert`는 **문장 변경 없이 GREEN 유지**

### §D.2 금지

- 파생 키 앵커의 종료 시 삭제(미증명-PASS 교착 재연 — t1544 card-review P2 재발)
- 영수증 전역 single-use화 / `WriteReceipt` 시그니처 변경
- Claude Code 런타임 페이로드 변경을 가정하는 수리(spec §1.3 열린 질문이 이것을 가리킨다)
- 비-required 트리로의 게이트 확대, `--no-verify`, 로컬 전체 스위트 실행
- RED 관측 기록 전의 어떤 수리 커밋

### §D.3 설계 결정 — 후보와 긴장 (iter1 D2 재범위 반영, M2 착수 전 결정 기록용)

REQ-RR-001(재범위: 단일-생존 선행 종료의 네 형태)과 REQ-RR-004(재범위: 생존 증명 + 모호 종료 경계 동결)를 동시에 만족하는 기제가 M2의 핵심 결정이다.

**불가결성 (감사자 반례, iter1 D2 채택)** — anonymous 이벤트만 보는 가드는 겹침 구간에서 start↔stop 귀속을 판정할 수 없다: (A 시작 t0, B 시작 t1, B가 r 주조 t2, 무귀속 FAIL stop t3, C 시작 t4, r 인용 PASS stop t5)라는 사건열은 두 판독 — FAIL=A·PASS=B(살아 있는 B의 자기 영수증 → 수락해야, REQ-RR-004/002)과 FAIL=B·PASS=C(죽은 B의 영수증 재사용 → 거부해야, REQ-RR-001) — 에서 바이트 동일하며 상반 판정을 요구한다. payload 귀속 신원이 없는 한(후보 3이 기각된 바로 그 이유) 어느 FIFO·최근-귀속 규칙도 한쪽 판독을 틀리게 한다. 이 창은 기제로 닫히지 않으므로 요구 쪽을 재범위했다.

| 후보 | 요지 | REQ-RR-001 (재범위) | REQ-RR-004 (재범위) | 판정 |
|------|------|-----------|-----------|------|
| 1. 단일-생존 종료-이벤트 경계 | (세션, 역할)별 종료를 기록하되, 종료 시점에 생존 개시가 정확히 1개일 때만 경계를 그 종료 시각으로 전진 — 2개 이상(모호 종료)이면 경계 동결 | 충족 — 단일-생존 선행 종료의 네 형태 전부에서 이후 인용 거부 | 충족 — 모호 종료 동결이 동시성 의미(t1544 card-review P2 핀)를 바이트 등가로 유지 | **선두** |
| 2. 인용-소비 원장 | 수락 시 (세션, 역할, 영수증) 소비 기록 | 부분 — 수락-선행 경로 한정(FAIL·거부·판정 부재 선행에 열림) | 충족 | 보조 |
| 3. 인스턴스 신원 브리지 | start↔stop 상관 토큰 | (충족 가능) | (충족 가능) | 현행 페이로드로 불가(spec §1.3) — 반례의 근본 원인이기도 하다 |
| 4. 앵커 전진 | 수락/종료 시 `StartedAt` 갱신 | 충족 | 위반 — 모호 구분 없는 전진은 AC-RR-003(동시성 핀)을 적색으로 만든다 | 기각 |

- **명명된 잔여 (겹침-구간 재사용)**: 모호 종료 뒤 형제·후행 인스턴스가 생존 인스턴스 시절의 영수증을 인용하는 경로는 재범위된 REQ-RR-001이 보장하지 않는다 — 반례가 증명한 판정 불가능 구간이다. 오늘의 노출("세션 내 최초 인스턴스 이후 영구") 대비 창이 "같은 세션·역할 인스턴스가 겹쳐 살아 있는 동안"으로 좁아지며, AC-RR-009이 M2 의미(단일-생존 전진 / 모호 동결 대조)를 못박는다.
- 기존 잔여도 유효하다: 무-정지 소멸 선행(stop 이벤트 없이 사망하면 종료 기록이 없음) — acceptance.md §E.
- 이 결정은 decision-index.md Q1에 재범위형으로 갱신됐고 DEFAULT-APPLIED다 — M2가 기본값을 뒤집을 근거를 찾으면 blocker 보고로 상향한다.

## §E Self-Verification (manager-develop 보고 의무)

각 항목 verification-claim-integrity 5단 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

- **E1** AC 이진 PASS/FAIL 매트릭스 (acceptance.md §C 전체)
- **E2** 교차 플랫폼 빌드 — `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`
- **E3** 커버리지 — `go test -cover ./internal/hook/... ./internal/auditreceipt/...` (패키지 목표 85%)
- **E4** 하위경계 grep — `internal/hook`에서 AskUserQuestion 0 일치
- **E5** lint — 신규 이슈 vs 기저 분리 보고
- **E6** 브랜치 HEAD + 커밋 SHA 목록
- **E8** RED 실패 출력 원문 — **GREEN 이전에 관측한** M1 재현 테스트의 실패 출력(명령+원문+exit) — test-after 실행은 이 항목을 낼 수 없어 구조적으로 불완전

## §F Milestones

> 결정 가역성 순: M1(RED)이 결함의 존재를 기계적으로 못박고, M2가 최대 변경 가능성(기제 결정)을 운반하며, M3는 기계적 수순이다.

### M1 — RED 재현 (Priority High)
- `internal/hook/audit_receipt_guard_test.go`에 재현 테스트 추가(예: `TestSubagentStop_SequentialAuditorReceiptReuseIsRefused`) — 시나리오: 세션 S + 역할 plan-auditor, ① 인스턴스 1 배경 시작 → 영수증 r1 주조 → r1 인용 PASS 수락(또는 FAIL 종료 — 두 arm), ② 인스턴스 2 배경 시작(동일 S+역할) → **감사 호출 없이** r1 인용 PASS → **차단을 기대**.
- `go test -list` swept-count 상관(신규 테스트명 0 일치 → 1 일치) 선행.
- **RED 관측**: 현재 코드는 수락하므로 테스트 적색 — 명령 + 원문 출력 + exit code + 트리 SHA를 progress.md §E.2에 기록. **이 기록 전에 어떤 수리 커밋도 착지 금지.**
- 산출: progress.md §E.2 기록 + 커밋 `test(SPEC-RECEIPT-REUSE-001): M1 RED reproduction ...`.

### M2 — GREEN 수리 (Priority High)
- §D.3 후보 1(종료-이벤트 경계) 기본으로 REQ-RR-001..003 충족 — 최소 변경(internal/hook + 필요 시 internal/auditreceipt 소폭 추가).
- 제약 테스트 전부 GREEN 유지: AC-RR-003..008 (기존 5 + 신규).
- 근거로 후보를 뒤집을 경우 blocker 보고 → decision-index Q1 갱신 후 진행.
- 산출: 커밋 `fix(SPEC-RECEIPT-REUSE-001): M2 instance-boundary ...`.

### M3 — 검증·정리 (Priority Medium)
- 영향 패키지 `-count=1` 재측정, `go vet`, golangci-lint(신규 0), E1-E8 기록.
- AC 전수 재판정 + progress.md §E.3 작성.
- 산출: 커밋 `test(SPEC-RECEIPT-REUSE-001): M3 verification ...` + §E 기록.

## §G Anti-Patterns

- **앵커 삭제 수리**: 수락/종료 시 파생 표식 삭제 — t1544 P2 교착 재연(§D.2 금지 1항).
- **앵커 전진 수리**: 후보 4 — 동시성 테스트를 적색으로 만드는 수리는 수리가 아니다.
- **전역 single-use화**: MCP 표면을 바꾸는 과잉 수리 — 범위 밖.
- **런타임 변경 기대**: Claude Code가 start 페이로드에 agent_id를 줄 날을 기다리는 수리 — 현행 페이로드 안에서 닫는다.
- **GREEN-우선**: M2 수리를 먼저 하고 M1 테스트를 나중에 쓰는 순서 — RED 원문이 남지 않아 TDD 불가증(E8).
- **기존 테스트 문장 수정으로 충돌 회피**: `ConcurrentBackgroundAuditorsShareEraAnchor` 등은 그대로 GREEN이어야 한다 — 시나리오가 서로소임을 M1에서 확인(§B).

## §H Cross-References

- spec.md §1.1 (검증된 경로) · §1.2 (위협 모델) · §2 (REQ-RR-001..007)
- acceptance.md §B (RED 원장) · §C (AC-RR-001..008)
- decision-index.md Q1-Q3 (기제·원인 어휘·기록 수명)
- `.claude/rules/moai/development/verification-completeness.md` §2 (2칸 규율) · `.claude/rules/moai/core/verification-claim-integrity.md` §3 (5단 보고)
- `.moai/docs/audit-artifact-convention.md` (감사 산출 경로 관행)
