# Progress — SPEC-REF-SEO-ABSORB-001

## §E.1 Plan-phase Audit-Ready Signal

- 산출 아티팩트: `research.md`, `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M + research.md)
- SPEC ID 정규식 사전 점검: 실행됨, 출력 `PASS`
- 요구사항: GEARS 표기 **23건** (REQ-SEO-001..009 콘텐츠 9 / 010..015 프로토콜 6 / 020..027 배포 8). 016-019는 **의도적 블록 결번**이며 spec.md §B 도입부에 관례로 명시되어 있다. 검증: `grep -o 'REQ-SEO-[0-9]\{3\}' spec.md | sort -u | wc -l` → `23`
- 수용 기준: **24건**, 전부 명령 + 기대 출력 + 실패 형태 명시. 검증: `grep -c '^### AC-SEO-' acceptance.md` → `24`
- 기계 판정 AC가 없는 REQ: **4건**(010 / 014 / 026 / 027) — 전부 §E DoD에 REQ 번호 명시 체크박스 보유. 전수 매핑은 `acceptance.md` §D.1
- 결정 게이트: `plan.md` §B 5건 **전부 확정**(tier=`optional-pack:frontend` / 감사 게이트 중간안 / 접근성 SEO 인과 4종만 / GEO 전면 제외 / docs-site 표 행만 추가). 검증: plan.md·research.md 양쪽 clarification 마커 grep → 각 `0`
- plan-phase 실측 baseline: skill dirs 31, catalog entries 41, Go 상수 31/41/41, 원문 861줄 · `sha256 c088f089…98ee7`, self-trip 8-gram 4600, 대조 skill 8-gram 0, LCS 17-23자, `GOOS=` 보유 템플릿 파일 2, ref 스킬 `## Verification` 체크박스 6-9개, catalog tier 분포 core 29 / devops 4 / backend 3 / frontend 3 / design 1 / harness-generated 1
- 상태: `draft`. Implementation Kickoff Approval 이전에 plan-audit 재감사 필요.

### plan-audit 이력

| iteration | verdict | score | 처리 |
|---|---|---|---|
| 1 | FAIL | 0.72 (Tier M 통과선 0.80) | MUST-FIX 6건 + SHOULD-FIX 일부를 v0.2.0에서 정정, 결정 5건 확정 |
| 2 | FAIL | 0.79 (0.01 미달) | iteration-1 MUST-FIX 6건 **전부 CLOSED 확정** — 재검증 대상 아님. FAIL은 v0.2.0 신규 콘텐츠의 1차 감사 결함 3건(N1/N2/N3)에 의한 것. v0.3.0에서 N1-N3 + SHOULD-FIX N4-N7 + SF-10 정정 |
| 3 | **PASS** | **0.87** (통과선 0.80) | must-pass 7/7, 회귀 0. N1-N3 전부 CLOSED — 감사가 24개 AC 본문에서 REQ→AC 매핑을 독립 재도출해 §D.1의 5건 오프셋·미판정 4건·팬텀 AC 0건을 재현 확인. 잔여 SHOULD-FIX 3건 중 **N10 즉시 정정**, **N8·N9는 사용자 결정으로 run-phase M6 이연** |

v0.3.0 정정 요약 — **N1** AC-SEO-015의 표 헤더 판독 명령 2건이 `grep -- '---' -B1` 옵션 파싱 오류로 실행되지 않던 문제(`-B1`이 파일명으로 해석 → exit 1 + 출력 0행). `-B1`을 `--` 앞으로 옮기고, 원문 측 실측 baseline(섹션 7행 / 표 헤더 10행)과 **네 `count`가 모두 1 이상일 때에만 판정**하는 전제 검사를 추가. **N2** §D.1 추적성 표를 내용 기준 전수 매핑으로 재작성 — §B.2 번호 오프셋(010→011 / 011→012 / 012→013) 명시, 기계 판정 AC 없는 REQ를 2건→**4건**(010/014/026/027)으로 정정, 재부여로 해소하지 않는 이유(다대다 관계라 1:1 불가) 기록. 같은 거짓에 기대던 spec.md §B 번호 관례 근거 문장도 실제 매핑에 맞게 재작성. **N3** AC-SEO-013 Then절 `60자`→`40자`(기대·실패 절과 통일). **N4** AC-SEO-002 H2 상한 근거 정정(ui-polish 실측 13, v0.2.0 기재 14는 research §B.4의 과다 계상 상속). **N5** AC-SEO-005 정규식 `reference[:,]`(9건 중 4건 통과)→`reference\b`(9/9), `NOT for:` 절 **내용** 판정 토큰 2종 신설. **N6** research.md §B.5 skill-routing 표면 열거에 실측 정정 각주. **N7** 접근성 6 vs 7 산술 분해 기록. **SF-10** REQ-SEO-027 라벨 `(Event-detected)`→`(Event-driven)`.

v0.2.0 정정 요약 — MF-1 AC-SEO-025가 언어 중립성 가드를 실행하지 않던 문제(강제 파일·`-run` 선택자 오지정), MF-2 AC-SEO-022의 secops 표면 집합 역전(skill-routing 제거 + 워크플로 본문 4건 추가), MF-3 접근성 결정과 REQ-SEO-006의 정면 충돌(개념 4종 편입 + AC 토큰 루프 확장), MF-4 요구사항 개수 오기재(21→23), MF-5 미해소 마커 5건 해소, MF-6 REQ 번호 결번 관례 명시. 추가로 클린룸 판정 대상 트리를 템플릿으로 통일(+AC-SEO-011b), 원문 digest 고정, 구조 발산 판정 신설(AC-SEO-015), tier 유효성 AC 신설(AC-SEO-020c), AC-SEO-006 공허 토큰 제거, AC-SEO-013 임계값 근거 부여, AC-SEO-002 H2 상한 정정, AC-SEO-021b 브랜치 전제 명시, 판정 AC가 없던 REQ-SEO-009에 AC-SEO-009 신설.

**v0.3.1 추가 정정 (N10)**: `research.md` §B.4의 `moai-ref-ui-polish` H2 수치를 `11개` → `명명 10개 + 불변 3 = 총 13개`로 정정하고 실측 각주를 달았다(N6와 동일 처리 — 조사 결과 본문 유지, 수치와 각주만 정정). 이 오기가 v0.2.0 AC-SEO-002의 근거 문장으로 전파되었던 것이 N4였다. **research.md는 형제 SPEC 2건이 재사용하도록 설계된 문서이므로 발원지 정정이 필수다** — acceptance.md만 고치면 아직 착수되지 않은 형제 작업에 같은 수치가 다시 흘러간다.

**run-phase M6 이연 (사용자 결정, 이번 개정에서 손대지 않음)**:

- **N8** — §D.1은 REQ-SEO-020을 완전 판정으로 제시하나, `catalog_tier_audit_test.go:278-279`가 `TrimPrefix`/`TrimSuffix`로 `templates/` 접두사와 후행 슬래시를 벗겨내므로 기계 판정되는 것은 `tier` 뿐이다.
- **N9** — AC-SEO-004는 REQ-SEO-004의 6개 frontmatter 필드 중 길이 절만 판정한다. `name` / `user-invocable` / `metadata` / `progressive_disclosure`가 전부 부재해도 스크립트는 정상 종료한다.

두 건 모두 AC 판별력 공백이며, **수정이 실제로 동작하는지는 `moai-ref-seo/SKILL.md`가 존재해야만 관측 가능**하다. 산출물 없이 AC를 고치면 v0.2.0에서 반복한 "실행해 보지 않은 판정 명령" 패턴을 재생산하므로, M6 검증 단계에서 산출물을 좌변에 놓고 함께 닫는다.

**forward-reference 정합 (해소)**: `research.md` §D.3·§D.5의 두 forward-reference가 v0.1.0 plan.md 마커를 가리키고 있었다. 마커 치환으로 무효가 되었으므로 각각 plan.md §B.5·§B.3 **결정 기록**을 가리키도록 포인터 절만 정정했다(조사 결과 본문은 무수정 — 감사에서 정확성이 확인된 부분이다). 5개 아티팩트 전부 잔여 clarification 마커 0건이며, 재감사의 MP-7 grep은 `plan.md`·`research.md` 양쪽에서 0을 반환한다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
