# Card t700 — Verdict

- card: t700 (웨지 안전 재뿌리 정책 — GPT 게이트웨이 receipt-chain 웨지 복구)
- spec: SPEC-GATEWAY-WEDGE-REROOT-001 (Tier M, 3-phase close 완료)
- branch: WT-wedge-reroot-policy (base: local develop `44e56d017`, 흡수: develop `643abfb8c`)
- evidence: .moai/reports/t700/{verdict,sync-audit,plan-audit,plan-audit-iter2}.md
- date: 2026-09-14, lane (t700 전담, 리드 배차)

## Claim (주장)

1. **클라이언트 측 단발 재뿌리 경로가 착지했다.** `moai cc --resume <uuid> --reroot`가 trailing 미발행 assistant boundary와 그 의존 tool_result만 대본에서 제거하고(나머지 바이트 동일 보존 — API-error 표시 행 포함), 영속 단발 마커(`<transcript>.reroot.json`, exclusive-create, 변이 전 기록)로 1회 한도를 프로세스 재시작 너머로 보장하며, 제거분을 `<transcript>.reroot-aside.jsonl`에 비파괴 보존한다. receipt store 핸들은 구조적으로 부재하고(`TestGatewayRerootPathHoldsNoReceiptStore`), 실 store 다이제스트가 전후 동일하다(AC-WRR-010).
2. **인가 우회 불가 판정이 기계적으로 잠겼다.** 수용 술어 불변(REQ-WRR-001) — 회복 후 수용은 무변경 `Manifest.Check`가 이미 부여하는 tail-truncation 형상뿐이다. golden-string 잠금(AC-WRR-007, 3개 오류 메시지 전체 리터럴 핀)이 오류 계약을 byte-identical로 고정하고, 위조 꼬리는 제거→aside 후 잔여 수용·재유입 거절(AC-WRR-016)로 갈라진다. 봉인 5파일 + `internal/gateway/receipt/` 경로는 카드 diff에서 히트 0이다(AC-WRR-012, 구현 13파일 → sync 후 14파일 재측정 동일 판정).
3. **독립 감사 2축 통과.** plan-audit 2차 PASS 1.00(1차 COND-FAIL 0.6875 → D1-D7 수리), sync-audit PASS 94.9/100 — 차단 0, 수리 라운드 불요.

## Evidence (증거 — 이번 run에서 직접 관측)

- **독립 검증 배치(오케스트레이터, HEAD `f5d865d38` 트리)**: `go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1` → `ok ... 1.055s`; `go test ./internal/cli/ -run 'TestGatewayReroot' -count=1` → 11/11 PASS `ok ... 0.982s`; `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `git diff --name-only 643abfb8c..HEAD` 13파일(구현 시점) — 봉인 파일 grep 히트 0; `families.Fork(` 프로덕션 호출점 1→1.
- **sync-auditor 재현**: 동일 스위트 재실행 + 커버리지 수치 재측정(translate 91.9%, rerootGatewayTranscript 85.4%, TranscriptPath 89.5%) 전부 보고치와 일치; merge-base 재유도 = `643abfb8c`로 base 귀속 유효 확인. 상세: `sync-audit.md`.
- **커밋 사슬(미푸시 12)**: `631704655`(plan 산출물) → `a86ff2e3c`(develop 흡수 1) → `10f6be792`(iter-2 감사) → `d025b463c`(재흡수, develop 643abfb8c) → `fc481e293`(M1 pre-flight) → `54a82200f`(M2) → `a2f46eda3`(M3) → `abbbb4eb9`(M4) → `714aa35c6`(M6) → `1c23204d0`(M7) → `f5d865d38`(sync close) → 본 verdict 커밋. 전 커밋 `(t700)` id + `🗿 MoAI` 트레일러.
- **TDD 실측**: E8 RED 증거(8× `undefined: rerootGatewayTranscript`) 캡처 후 GREEN; RED-GREEN 루프가 첫 구현의 API-error 표시 행 오인을 잡아 경계 판별에서 제외시킨 실제 버그 1건 수리.

## Baseline-attribution (baseline 귀속)

- 브랜치 WT-wedge-reroot-policy @ `f5d865d38`, base develop `44e56d017`(배차 명시), 흡수 develop `643abfb8c`(t697 병합 `f45c2dddf` 포함 — 조상 게이트 PASS). 전 증거 this run, this worktree. sync-auditor 측정 시점 로컬 develop `4f5a9006f` 전진 확인 — merge-base 재유도로 귀속 무손상.

## Gaps (미검증 — 명시적)

1. **AC-WRR-013 라이브 라운드트립 미실행** — 오케스트레이터 처분(§E.2 M1 기록): t672 matrix C4b가 수용 형상을 라이브로 이미 증명하고 신규 표면(대본 수술)은 단위 잠금으로 충분, upstream 비용 회피. 라이브 end-to-end는 미관측.
2. **전체 `internal/cli` 스위트 로컬 미판정** — 배경 실행이 600s 기본 타임아웃 중단 형태(해당 패키지 역사적 ~1583s)로 상세 유실, UNRESOLVED로 기록. 변경면 선택자(`TestGatewaySession|TestGatewayConversation|TestGatewayReroot` 17종 + `TestReceiptHistory`) 3회 green. 전체 판정은 리드 일괄 push의 CI 몫.
3. **재개 가능 웨지 형상의 경계** — plain user turn이 phantom boundary 뒤에 있는 웨지는 복구 후 네이티브 completion gate가 재개를 거절한다(제거 의미론은 REQ-WRR-003-1대로 불변, aside 전량 보존). 이 일반형의 정식 경로는 fork 대체로 M6 문서 §5에 기재.

## Residual-risk (잔여 위험)

- 클라이언트는 웨지 꼬리와 위조/추론-제거 꼬리를 구별할 수 없다(구조적) — 단발 마커가 오적용을 1회로 제한하고, 비-웨지형은 분류 사유와 함께 계속 거절된다.
- 착지 승인과 클라이언트 재개 사이의 간극: 일부 형상은 `--reroot` 후에도 fork 대체로 빠진다 — 사용자 체감상 "새 대화"와 유사한 경로가 문서화된 채 남는다.
- sync-audit 선택 발견 F1-F3(마커-aside 창 간 안내 경로, 문서 뉘앙스, 0600 모드) — 재개 불요, 기회 수리 후보로 기록.
