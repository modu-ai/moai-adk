# t92 + t93 Review Verdict — PASS (양 카드)

- t92: Agent Teams 허용 독트린 전환 (`94bc55370`, 35파일 +310/−224) · t93: Task 도구 제거 실측 (`fa9f51efa`, 리포트)
- Branch WT-t92 (merge-base `bc956d7ac` — 이후 release/v3.1.1이 `7b2b3562b`로 전진했으나 통합 시 정상 병합 범위) · ride-along 정확히 2커밋
- Reviewer session: release-v311 (2acd4be4), 2026-08-17

## Claim (주장)

전환 5원칙이 32콘텐츠 파일에 일관 적용됐고, 양면 증거가 정직하며, 중립성이 기계적으로
보장되고, t93은 측정 경계를 준수했다.

## Evidence (증거) — reviewer-executed

| Check | Command | Observed |
|---|---|---|
| 템플릿 전량 | `go -C <t92> test ./internal/template/ -count=1` | **ok 43.227s** |
| leak·sentinel·parity | `MOAI_TEMPLATE_LEAK_STRICT=1 go -C <t92> test ./internal/template/ -run 'Leak\|Sentinel\|Parity'` | **ok 1.575s** — cp 날짜 유입 복원 주장의 기계적 뒷받침(커밋 트리에서 통과) |
| 유령 참조 스윕 | `git grep 'Agent Teams.*RETIRED'` / `git grep -c RETIRED` / `git grep MODE_TEAM_UNAVAILABLE` @94bc55370 | 팀 관련 RETIRED는 settings-management 계보 문단(로컬↔미러 동일문)만; 무관 RETIRED(moai.md:602 분류법, design 헌법 디자인 폐기 등)은 t92 이전 것; 센티널 언급 전부 "retired-era 역사 마커" 프레임 — **현재-불가 프레임 0건** |
| Go 테스트 무변경 | 커밋 파일 목록 | _test.go 부재 — "테스트 변경 0" 주장 성립 (센티널 감사 REQ-WF003-011가 통째로 통과) |
| §C.1·§15 실물 | blob 독회 | Mode 3 "experimental (re-allowed)"·명시 요청 전용·자동선택 금지(anti-pattern §198)·Tier L 자동 라우팅 manager-kanban 유지·양면 증거·제약 조건 목록 — 일관 |
| E1 원시 | raw/e1-taskstop.txt | `task_type: in_process_teammate` 실재 + 반증 관측 병기 — 정직 |
| t93 | report.md + 원시 | 측정만, 수정 후속 (a)/(b) — 경계 준수; 대조군(TaskList 무마커) 설계 타당 |

## 판정 — 디스패치 4개 심사 지점

1. **서술 정합성: 충족.** §A 표·§B 트리·§C.1 전면재작성·§C.2·anti-pattern·비회귀 노트·
   crosswalk·CLAUDE.md §4·§15·run.md 센티널 문단이 동일 방향; 미러 16쌍 패리티(테스트+스팟).
2. **양면 증거 정직성: 적절.** "불일치 미해결 — 신뢰성 미증명, 세션마다 검증"은 올바른
   인식론적 자세. **리뷰어 제3 관측 추가**(아래 기여).
3. **중립성: 기계적으로 확인.** STRICT leak 통과(커밋 트리에서 재실행).
4. **t93 경계: 준수.** 단, 파급(manager-kanban의 TaskList 조율이 백그라운드 경로에서
   조용히 미동작)는 운영상 실질적 — 후속 카드 우선순위 높일 것 권장, 옵션 (b)는 칸반
   "완료는 읽는다" 원칙과 선천 정합.

## 리뷰어 기여 — E1 불일치에 대한 제3 관측 + 가설 (후속 카드 재료)

본 리뷰 세션(GLM) 자체가 t92 직후 named 스폰을 사용: `t79-deep-audit` 스폰이
teammate로 전환(메시지가 teammate-message로 도착), 완료 후 **result 채널 무발송·idle
알림만** — 리뷰어의 SendMessage 요청에 **전체 리포트가 mailbox로 정상 회신**됨.
run/sync/plan 레인의 디스패치도 mailbox 유입. 해서:
- 리드 관측("5-워커 정상 완료")을 지지하는 **제3의 세션 관측** — 전달은 실재하되
  **mailbox 채널** 경유.
- 반증 가능한 가설: teammate 전환 스폰의 **result 채널은 발화하지 않으며, 전달은
  mailbox 기반** — E1 프로브는 발화하지 않을 채널을 기다렸을 가능성. 다만 E1의 2회
  poke가 무응답(본 세션의 poke는 응답)이라 **가설이 전부를 설명하지 못함 — 불일치는
  여전히 미해결**, 원인 규명은 후속 카드가 맞음.

## Findings (advisory)

1. **GLM 상속(치명 gap)**: 후속 측정 카드를 **높은 우선순위로 권장**. 재료 2가지:
   (i) in-process teammate는 동일 프로세스 — 세션의 client/base URL을 구조적으로 공유할
   가능성(별도 client 스폰 여부 미검증), (ii) 본 리뷰 세션이 GLM+live teammate 테스트베드
   (mailbox Q&A로 자가 진단 가능 — E1과 달리 응답 확보 경로 입증됨).
2. `background: false` 회복 여부 미검증(t93 Gaps 명시) — (a) 채택 시 선행 실측 필요.

## Gaps (미검증 — 리뷰어 몫)

- catalog.yaml 해시 재생성의 byte 수준 대조는 개별 확인 안 함(카탈로그 테스트가 전량
  패키지 런에서 커버).
- 32파일 전 혼크 대조가 아닌 핵심 혼크+스윕 기반 판정 — 전수 문장 대조는 아님(스윕이
  유령 클래스를 기계적으로 봉쇄).

## Residual-risk (잔여 위험)

- 배포 문서가 "허용"을 말하나 결과 반환 신뢰성은 미증명 — §C.1의 세션마다 검증 권고가
  유일 완충(기계 게이트 아님). CC 버전업 시 재검증 문구 유지 필요.
- t93 파급(백그라운드 경로 Task 조율 무동작)은 후속 카드 전까지 잔존.

**Verdict: PASS (t92 + t93)** — 자가 통합(merge --no-ff + push) 진행.
