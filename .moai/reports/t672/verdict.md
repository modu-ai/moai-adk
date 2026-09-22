# Card t672 — Verdict

- card: t672 (GPT 게이트웨이 오버레이 서브에이전트 스폰 400)
- branch: WT-receipt-spawn-400 (base: develop 61a9bb57e — t695 투명성 패치 포함)
- evidence: .moai/reports/t672/{verdict,investigation,matrix}.md
- date: 2026-09-13, lane-6

## Claim (주장)

1. **카드 원 가설 (a) 오버레이가 서브에이전트 model 인자를 못 변환 / (b) 프로필 주입 생략 필요 — 측정으로 반증됐다.** 수신 체크는 자식 프로세스당 하나의 receipt 루트(gateway_factory.go:178)를 부모·스폰·팀메이트·요약 포크 전부가 공유하고, 신규 스폰의 첫 턴은 assistant boundary가 0개라 체크를 구조적으로 통과한다. 수리 전 빌드로도 model-arg 스폰 형상 요청이 실요청 200이었다(matrix C2). 원 가설의 두 수리 방향 모두 결함이 아님.
2. **실제 프로덕션 스폰/포크 400의 실체**: 일시 실패(당일의 effort 400·상류 502)가 클라이언트에 "미발행 턴"을 남기고, 이후 모든 재생이 receipt 사슬 검사(Check: Prefix·Previous·Opaque·Items 4속 일치, request.go:272-276 → receipt_history.go → receipt/core.go:99)에서 거절되는 **웨지**다. 거절 자체는 정상(외래 항목·변조 이력 차단이라는 인가 기능)이나, 사유가 "모델 패밀리" 문구로 오인되게 쓰여 있었다.
3. **수리는 수락 완화 0의 원인 분류다**: HistoryReplayError가 ReplayCause(사슬 불일치/혈통 불일치/추론 누락)로 분류돼 400 본문이 스스로 사유를 밝힌다. 기존 안내 문장은 그대로 보존해 과거 장애 서명과의 호환을 유지한다.
4. 마스크드형(이름 붙은 팀메이트 400)은 수리하지 않고 계측으로 남긴다 — 다음 발생 시 본문이 스스로 식별한다.

## Evidence (증거 — 이번 run에서 직접 관측)

- **기제 증명**: gateway_factory.go:178(루트 개방)·:200-204(검증기·메타데이터 양쪽 배선)·:69(외래 세션 메타데이터는 history 이전에 "native receipt authorization required"로 거부 — 관측된 HistoryReplayError는 전부 이 관문 통과분). Check 요구사항 4속 일치.
- **실요청 셀 9종 × 수리 전/후 바이너리**(격리 게이트웨이·에페머럴 포트·실 자격증명 읽기전용): 스폰형(model-arg·무arg)·꼬리절단 포크형 — 수리 **전에도** 200(정상 스폰/포크는 원래 안전); 웨지 재생·추론-제거 재생 — 전후 모두 400이나 수리 후 분류 사유 동봉 확인(예: "…start a new conversation (reason: replayed history does not match the recorded receipt chain…)").
- **단위 테스트 8건**(receipt_history_cause_test.go): 절단 포크 통과 3형태·사슬 중간 절단 거절·짝 중간 절단 거절·위조 봉투 거절·외래 루트 혈통 분류·스폰형 신규 이력 통과(sol·astra)·추론-제거 거절 고정·웨지 형상 거절. 레인 오케스트레이터 재실행: `go vet ./internal/gateway/translate/...` clean, `go test -run TestReceiptHistory` → `ok 0.606s`.
- **커밋**: 715b6b3ea(feat)·4a0c9683e(docs) — diff 5파일(receipt_history.go +71/-4, 신규 테스트 255행, 증거 3) 범위 적정.

## Baseline-attribution (baseline 귀속)

- 브랜치 WT-receipt-spawn-400 @ 4a0c9683e, base develop 61a9bb57e — 이번 run, 이 워크트리.
- 수리 전 대조 바이너리: `git archive HEAD` 빌드(sha256 bf305362…), 수리 후 워크트리 빌드(c61a2af5…).

## Gaps (미검증 — 명시적)

1. 프로덕션 스폰/포크 사고의 정확한 클라이언트측 유발자(절단 vs 추론-제거 vs 선두 편집)는 다음 분류된 400 본문이 나와야 확정 — 현재는 유추.
2. lane-3 원 재현(같은 세션에서 model-arg 스폰만 실패·무arg Explore 성공)은 미재현: 남은 유력 가설은 "요청 형상별 effort/output_config 차이 + 당시 effort 마스크드 400"이나 당시 로그로는 확정 불가. D1 이후 재발 시 본문이 답한다.
3. 웨지의 자동 복구(안전 재뿌리 내기)는 구현하지 않았다 — 현재 복구는 "새 대화 시작" 안내(분류 사유 포함)뿐.
4. upstream 응답 본문은 어댑터 변환 이후로만 관측 가능.

## Residual-risk (잔여 위험)

- 웨지 UX는 여전히 사용자에게 "새 대화"를 강요한다 — 분류로 원인은 보이지만 502 뒤 본 대화 사망(lane-8 사례)은 재발한다. 안전한 재뿌리 정책은 보안 판단이 필요한 별도 카드 후보(리드 결정).
- 마스크드형 팀메이트 400의 근본은 미특정 — 다음 발생 본문이 "native receipt authorization required"(t695의 메타데이터-세션 불일치 가설 확정)·분류된 재생 사유·또는 다른 문자열로 답을 준다.
- 분류 문구는 고정 문자열이라 장애 서명 매칭에 영향 없도록 기존 문장 뒤에 덧붙였으나, 본문 파서를 두고 있는 소비자는 사유 절 추가를 알아야 한다.

## 원 카드 지시 사항 (a)/(b) 판정 기록

(a) 오버레이 매핑 확장 — 불요: model-arg 스폰형 요청이 매핑·수신 전부 정상(실측 200). (b) 프로필 주입 생략·상속 강제 — 불요: 같은 요청이 상속·주입 양쪽에서 통과. 둘 다 t649의 어댑터-소유 설계와 무관한 비결함으로 판정한다.
