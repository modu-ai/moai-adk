# t1914 — 진행 기록 (t1839 잔여: 카드 정책 재설계 단계 C 인계)

카드: t1914 · Class C · P1 · run tn564n · lane-11. 임대: `moai factory next --card t1914 --run tn564n` exit 0 (worktree=t1914, 신규 트리 WT-design-progress-moai @8249bff8c3).

## 2026-10-21 ~01:1xZ — 배차 ①원 writer 산물 원위치 확인 (전부 확인)

- SPEC-CARD-ELIGIBLE-PULL-001/ 5파일+sweep-union.txt — t1839 트리 원위치.
- progress: `.moai/reports/t1839/progress.md`(46K, merge-ready v29 기록 포함) + SPEC progress.md(68K).
- PR #1844 Windows 분석: `.moai/reports/adk-pr-priority-20261019/continuation/windows-analysis-lane60.md` — 원위치.
- 설계 정본: 주 체크아웃 `.moai/reports/card-policy-redesign-20261017/design.md` — 원위치.
- 선행 상태 실측: WT-10-17-c가 base 대비 28커밋 보유(F10~F14 수선·7 sync-audit 라운드·round-5 PASS-WITH-DEBT→round-7 델타 PASS @ed6c8180b0·F18 untrack @a63a2055f7). lane-60 검증(7시험 셀렉터 exit 0 36.486s @a63a2055f7) 기록 존재. 구현은 사실상 완결 — 잔여는 착지(draft PR).

## 2026-10-21 ~01:2xZ — ②흡수·충돌 해소·검증

- 병합 1: a63a2055f7 → bf89ccf425. 충돌 2건 — factory_card.go: HEAD의 신규 `attempt == 0`(큐락 대기 인수) + 브랜치의 elig 배선 양측 보존으로 해소(중복 `})` 정정 후 gofmt clean); CHANGELOG 양측(t1839·t1668 항목) 보존.
- 검증 1 @bf89ccf425: 고정 7시험 셀렉터 `-count=1` → `ok internal/cli 29.376s`, TEST_EXIT:0.
- 병합 2: 16d03deea3(release head) → b9d27825d4. 충돌 CHANGELOG 1건(양측 t1839·t1809 항목 보존).
- 검증 2 @b9d27825d4: 동일 셀렉터 → `ok internal/cli 28.596s`, TEST_EXIT:0.

## 2026-10-21 ~01:3xZ — ③push·draft PR

- push `WT-design-progress-moai` → origin (new branch, exit 0).
- draft PR #1859: base release/main-batch-20261019, head b9d27825d4, 병합 승인 아님(리더 559a 게이트 소관).

## Gaps

- 패키지 전체 internal/cli 시험·커버리지 미실행(범위 규율 — batch CI 소관, 카드 사슬 동일 처분).
- 새 머지 헤드(b9d27825d4)에 대한 독립 sync-audit은 미실행 — dispatch가 감사가 아닌 draft PR까지를 범위로 명시; 병합 전 559a codex 게이트가 리더 측에서 수행.

## 2026-10-21 ~03:2xZ — codex 재리뷰 발견 처분 (리더 승인분) + 설계 요구사항 전문 기록

- **P2-c 수리(승인)**: SPEC-GLM-JEV-KEY plan.md:149·acceptance.md:35의 선택기를 `-run '^Test(GLM|Glm|Jev|Key|Root).*$'`로 수정 — 이전 수리(f526de7afb)의 `$` 앵커가 접두 패밀리(174건)를 빈 스위프(0건)로 만든 결함 정정. 실측: -list 174건 재입증. 자진 보고가 아니라 codex 재리뷰가 잡은 내 수리의 결함 — 빈 스위프 규정(verification-completeness §1.1) 준수 위해 재수정.
- **stale 앵커 2건 수리(승인)**: PROGRESS-RECORD-IO acceptance.md:50 → 4 preservation 테스트 교대앵커(스위프 4건 실측); :147 → AC 본문이 인용하는 M3 스위프(28-목록, darwin 27건+linux 1건)로 재지향. 기록명 유지 옵션은 리더 처분으로 폐기.
- **P1·P2-a·P2-b — t1914/t1839 설계 요구사항 (구현은 자격 설계 착수 시; #1859 수리 대상 아님 — 리더 처분 ③)**:
  - **P1 (factoryNextPRSnapshot 사전 조회 게이트)**: bare 선택은 factoryNextLeaseOnceEligibility에서 self_dispatch와 5개 자격 조건을 적용하지만, 바로 앞의 factoryNextPRSnapshot이 사용하는 factoryNextMayLease는 여전히 모든 unowned picked 카드를 임대 가능하다고 판정. self_dispatch=false 기본값에서 그런 카드만 남아 실제 선택이 항상 no-card여도 `--wait`는 기본 15분 동안 5초마다 `gh pr list`를 실행해 레인 하나당 약 180회의 불필요한 네트워크 요청 — 여러 레인이 GitHub rate limit을 소진할 수 있음. 사전 조회도 동일한 플래그와 자격 판정을 공유해 실제로 임대 가능한 카드가 없으면 쿼리를 건너뛰어야 함.
  - **P2-a (레코드의 spec_id도 자격 판정에 사용)**: self_dispatch가 켜져 있고 운영자가 `factory assign --spec ...`으로 SPEC을 팩토리 레코드에만 기록한 경우, homestate.Card.SpecID는 유효해도 큐의 BacklogItem.SpecID는 비어 있을 수 있음. 이 조건은 큐 힌트만 검사하므로 해당 카드를 spec-missing으로 거부하고 bare `factory next`가 계속 no-card로 끝남. 레코드 또는 큐 힌트 중 하나에 SPEC이 있으면 통과하도록 arm (b)의 c.SpecID도 판정에 전달해야 함.
  - **P2-b (타 레인 번들 소유권을 선행 대기로 오인 금지)**: self_dispatch 켜진 상태에서 owner label은 비어 있지만 다른 레인이 소유한 번들의 멤버를 평가하면 selectionSkips는 번들 소유권 때문에 true를 반환. 이 값을 그대로 depsMet=false로 넘겨 이미 선행 카드가 병합된 경우에도 refused predecessor-unmerged를 출력 — 자동화나 운영자가 존재하지 않는 병합 대기를 해결하려고 잘못 대응. 외부 레인의 번들 멤버는 게이트 평가 전에 조용히 제외하거나, 번들 소유권 판정과 실제 predecessor/hub 대기 판정을 분리해야 함.
- 출처: PR #1859 codex Code Review (2026-10-21T02:55:19Z, head f526de7afb) 인라인 발견 전문.
