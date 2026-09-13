# t684 auto-done 실측 오판 — 원인 판정 (lane-4, 2026-09-13)

## 관측 (리드 실측)

`moai todo auto-done --dry-run` (origin/develop e7b93c120 착지본):
scanned=37 closed=8 중 **t654·t656·t657을 form=subject-attribution으로 닫으려 함** — 셋 다 picked(생존) 상태의 재발행 카드이며 미착지. 리드는 실실행 보류, 5장 수동 done.

## 귀속된 커밋 (원/구세대 카드들의 착지물 — origin/develop 전수 제목 검색, 이번 실행)

- **t656 (구세대 = SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001, 병합 30cf7f422)**: 제목 토큰 23건 — `feat(t656): wire settings.json snapshot…` (b5b5883e9), `docs(t656): sync SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001` (0b5a1b3b1), `merge: WT-update-value-merge… (t656)` (30cf7f422) 등
- **t654 (구세대 = WT-guard-quoted-span, 병합 507f67a55)**: 6건 — `merge: WT-guard-quoted-span into develop (card t654)` (507f67a55), `docs(handoff): save the resume body… (t654)` (dd8e77107) 등
- **t657 (구세대 = web 콘솔 카드)**: 4건 — `feat(web): refine console IA and settings details (t657)` (32f97cd75), `merge: integrate t657 host-gate test fix` (987eb7e40), `test(web): honor host gate in t657 surface check` (c3580844e), `merge: integrate t657 web UI UX` (be89d2c43)

합계 33건+ (t654/t656 토큰 grep 41행 — 중복 병합 계보 포함). 전부 **재발행 이전 세대** 카드의 착지물이고, 현재 큐의 t654/t656/t657(새 카드)과는 무관.

## 근본 원인 (2단계)

1. **M1 오판 가드의 도달 범위**: `AutoDoneDistinctTexts`는 **홈 스토어 안의** 동일 id 서로 다른 텍스트만 센다. 구세대 t654/t656/t657은 홈 스토어에 없던 카드(프로젝트 스토어 소속 또는 그 이전 세대) → 홈엔 재발행 새 카드만 존재 → DistinctTexts=1 → `ambiguous-id` 미발동. (sync-audit 잔여 위험 F1 · t684 verdict Residual-risk 에 기록된 그 실패 양상 — "cross-store 재발행은 DistinctTexts=1로 보인다")
2. **제목 귀속의 시간 무경계**: 귀속 술어가 develop **전체 이력**을 훑어 구세대 커밋 33건+을 새 카드의 착지로 귀속. 카드 생성 시각과 커밋 시각의 선후 관계를 보지 않음.

회귀 픽스처가 이걸 못 잡은 이유: AC-AD-004/005의 충돌 픽스처는 **같은 스토어 안에** 전임 레코드가 있는 형태(SQL 주입)로 만들어져 있음 — "전임이 스토어에 아예 없는" 세대교차 재발행은 픽스처 집합 밖이었다.

## 수리 후보 (수리 카드 판정용 — lane-4 의견, 비구속)

1. **시간 경계 (구조적, 권장)**: 제목 귀속은 `커밋 committer date > 카드 created_at` 만 허용. 구세대 커밋은 재발행 카드보다 앞서므로 원리적 배제. 스토어가 단일화된 뒤에도 유효.
2. **재발행 원장**: id 할당기가 재발행 이벤트를 기록 → 재발행 id는 subject 귀속 부적격(recorded-SHA 만 허용).
3. 문서형 제목(`docs(t656):`)까지 잡는 육상 형태 축소 — 부분 완화일 뿐 단독 불충분.

## 관련 카드

- t684 수리(위 후보 1+2 권장) — 운영자 판정 대기
- t657(큐 병합)과 독립 — 병합 후에도 시간 경계는 계속 유효 (t657이 이 결함을 고쳐주지 않음)
- 본 파일은 WT-todo-queue-merge 브랜치에 커밋 예정(run 에이전트 종결 후 — 동일 브랜치 커밋 경합 회피)
