# Card t708 — Verdict: launcher-side reasoning-envelope repair (`--repair-envelope`)

- card: t708 · SPEC: SPEC-GATEWAY-ENVELOPE-REPAIR-001 (v0.2.2, completed)
- branch: `WT-envelope-persist` (worktree `.claude/worktrees/t708`, base develop `7a7a08f20`, re-absorbed develop twice: `4da5d1c4e` t707 → `febadc784`, `93ae49ce7` t653 → `1aca63da8`)
- sync close: `430cf4429` · re-close after repair: `c9f1477b3`
- date: 2026-09-14, lane session · audits: plan-audit PASS 0.975 (iter2) · sync-audit iter1 FAIL 78 → repair → **iter2 PASS 92** (`.moai/reports/t708/sync-audit.md`)

## Claim (주장)

1. **런처 측 verbatim 봉투 재주입이 t703 stripped-replay 형상의 복구 경로다.** 명시적 `--repair-envelope`(+`--resume`) 동사가 패밀리 네이티브 트랜스크립트에서 게이트웨이 발행 `moai_opaque_v2` 봉투를 찾아, 생존 마커의 자기 증명 다이제스트(`toolu_moai_v1_*` 내장 `opaque_sha256` == sha256(후보 raw)) 대조 후 원 바이트를 원래 경계에 재주입한다. 단발 내구 기록(`families/<FamilyID>/repair/<UUID>.json`), 비파괴 aside(`<transcript>.moai-repair-aside`), 전부-무-아니-전부 거부. 검증기·발행·오류 계약은 0 변경 — 최종 판정은 언제나 변경 없는 `Manifest.Check`가 내린다(합성 거절 의미론, spec v0.2.2 §3.3).
2. **설계는 서버 측 수리를 전제하지 않는다** (리드 지시 준수). t707의 트랜스크립트≠요청 바이트 실증과 t703 사건 트랜스크립트의 봉투 36개 온전 보존(레인 직접 측정)이 전제를 강화한다: 트랜스크립트가 옳은 바이트를 지키고 요청만 어긋나므로, 지속 원천을 올바르게 되돌리는 것이 저장소 안에서 도달 가능한 지렛대다. 클라이언트가 수리된 트랜스크립트에서 다음 요청에 봉투를 싣는지는 저장소 밖 인코딩 시점 동작으로 문서화된 경계다.
3. **경계 준수**: PRESERVE 11파일 zero-diff(머지 후에도 실패 가능한 M4 기계 락), 수신처 저장소 읽기 0(소스 스캔 테스트), 봉투 제조 금지(verbatim 자기발행 바이트만), t700 시접 Reading B(M0 재정 PASSED — REQ-WRR-008은 게이트웨이 상태 변경 경로만 지배).

## Evidence (증거 — 이번 run에서 직접 관측)

- **RED→GREEN (M3)**: `go test ./internal/gateway/conversation/ -run 'TestRepair'` 빌드 실패(`RepairEnvelope` undefined) → 커밋 `7ebf85df6` 후 conversation 7/7·cli 배선 4/4·gateway 표면 38 PASS.
- **바이트 보존 (sync-audit F1 수리, 커밋 `b46d33271`)**: 실제 클라이언트 형상 행(Node 삽입순서·리터럴 `<`/`&`)에 대해 "주입 스팬 제거 시 원 입력 바이트 재현" 불변식 테스트 GREEN — 구 splice의 알파벳 재정렬+HTML 이스케이프는 RED로 재현됐다(`TestRepairInjectLinePreservesRowBytesOutsideInjection`).
- **전 스위트(t653 재흡수 트리 `1aca63da8`)**: conversation `ok 31.839s` · translate `ok 36.059s` · cli `Gateway|Repair` `ok 13.992s`(M4 라이브 락 0 violations) · `GOOS=windows` 빌드·vet exit 0.
- **lint**: 병합 조건 실측 `golangci-lint run --timeout 9m ./internal/gateway/... ./internal/cli/...` → `0 issues.` exit 0 (비범위, translate 포함 — F2의 5 errcheck는 `b46d33271`에서 수정).
- **감사**: plan-audit iter1 COND-FAIL 0.8375(D1/D2/D4) → 1패스 수리 → iter2 PASS 0.975. sync-audit iter1 FAIL 78(F1 재마샬 바이트 훼손·F2 errcheck·F7 합성 거절 명시) → 수리(`b46d33271`/`c9e98df4c`) → iter2 **PASS 92**(기능 92·보안 95·완성도 88·일관성 93).

## Baseline-attribution (baseline 귀속)

모든 측정은 2026-09-13~14 이 레인 세션에서, `.claude/worktrees/t708` 트리(커밋별로 §E.2에 개별 핀)에서 수행. 전체 스위트(`go test ./...`)는 로컬 미실행 — 레인 부하 규율; 전체 판정은 `origin/develop` push 뒤 CI의 몫이다. `internal/cli` 전체 패키지 테스트는 이 머신 기준 ~1583s로 기본 10m 타임아웃 초과가 기존 baseline(게이트웨이 표면 스코프 실행으로 대체).

## Gaps (미검증 — 명시적)

1. **M5 라이브 프로브 스킵** (조정자 결정, t707 후속 소관): (a) 클라이언트 compaction/`--continue` 슬라이싱 후의 봉투 보존 미측정 — 수리는 source-gone에서 거부하므로 설계 비의존; (b) 수리된 트랜스크립트가 실제 클라이언트 재생에 미치는 end-to-end 효과 미검증 — 저장소 내 테스트는 트랜스크립트 측 기제와 불변 Check 조합만 증명.
2. 커버리지 §D.3 게이트(패키지 ≥85%)는 패키지 수준에서 입증되지 않음(conversation 79.6% 전체, 신규 코드 파일별 44–100%) — §E.3에 한계 명기.
3. t700 미착지 — M0 Reading B 유지, 착지 시 §A step 2 재독본 병합 순서는 리드 판정.
4. 운영자 문서의 최상위 명령 구문 예시는 prepareGatewayConversation 수준에서만 테스트 검증(동일 M5 축).

## Residual-risk (잔여 위험)

- 봉투가 트랜스크립트에서 사라진 대화(compaction 등)는 수리 불가 — 거부 후 안내대로 새 대화 or fork. aside-crash 형상은 영구 거절(수동 aside 제거로 복구, 문서화됨).
- 합성 거절 셀(missing-marker·Prefix 불일치)의 최종 거절은 요청 시점 Check가 담당 — 수리만으로 대화가 살아난다고 보장하지 않으며, 그 경우에도 분류된 400 본문이 다음 단계를 알린다.
- 바이트 동일성 테스트는 단위 수지(N1) — 파이프라인 통합 재생 검증은 M5 스킵 축에 눌려 있다.
