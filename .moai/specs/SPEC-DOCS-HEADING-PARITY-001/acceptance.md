# Acceptance: SPEC-DOCS-HEADING-PARITY-001 (card t1328)

모든 AC의 측정은 spec.md §4의 고정 패리티 카운터(`grep -rc '^#\{2,\} '`, H2-이하 전 계층, 펜스 포함, H1 제외, 페이지×로케일 단위)로 한다. 다른 정의의 센 결과는 증거로 인정되지 않는다.

## D. AC Matrix

### AC-DOCS-001 — agent-guide 4로케일 헤딩 수 일치

- **Given** `docs-site/content/{ko,en,ja,zh}/advanced/agent-guide.md`가 각각 존재하고
- **When** 고정 카운터로 네 파일을 측정하면
- **Then** 네 값이 모두 같다 (정렬 전: ko 22 / en 30 / ja 30 / zh 30).

### AC-DOCS-002 — multi-llm 4로케일 헤딩 수 일치

- **Given** `docs-site/content/{ko,en,ja,zh}/multi-llm/_index.md`가 각각 존재하고
- **When** 고정 카운터로 네 파일을 측정하면
- **Then** 네 값이 모두 같다 (정렬 전: ko 11 / en 5 / ja 5 / zh 5).

### AC-DOCS-003 — tokenomics 4로케일 헤딩 수 일치

- **Given** `docs-site/content/{ko,en,ja,zh}/advanced/tokenomics-overview.md`가 각각 존재하고
- **When** 고정 카운터로 네 파일을 측정하면
- **Then** 네 값이 모두 같다 (정렬 전: ko 12 / en 13 / ja 12 / zh 11).

### AC-DOCS-004 — tokenomics en 단독 사실 수리 생존 (로케일별 예상 패턴)

- **Given** `9a53efd24`(t1115 sync-audit 수리)와 `7ad954556`(t1095 Sonnet 4.5/4.6 1M 경계)가 만든 문장의 **로케일별 예상 표면 패턴 표**가 아래와 같고
- **When** 재유도가 끝난 각 로케일의 tokenomics 본문에 자기 로케일의 패턴을 각각 grep하면
- **Then** 표의 모든 패턴이 모든 로케일에서 각각 1회 이상 적중한다(어느 패턴이든 적중 0이면 FAIL).

| 로케일 | 예상 패턴 | 패턴 근거 |
|---|---|---|
| en | `concentrated on the rows that judge` **및** `Sonnet 4\.5 and earlier \(200K\)` | `9a53efd24`(en 단독) / `7ad954556` |
| ko | `Sonnet 4\.5 이하 \(200K\)` | `7ad954556`(en+ko — ko는 커밋 시점에 이미 반영) |
| ja | `Sonnet 4\.5.*\(200K\)` | M2 재유도가 이 사실을 처음 도입 — 현재 0 적중 |
| zh | `Sonnet 4\.5.*\(200K\)` | M2 재유도가 이 사실을 처음 도입 — 현재 0 적중 |

**선결조건(실행 시점)**: ja/zh 다리는 M2(tokenomics 재유도) 착수 전에는 구조적으로 적중 0이다 — 이 AC는 **M2 종료 시점**에 판정한다. M2 이전 측정에서 ja/zh 0은 예상된 사전 상태이지 FAIL이 아니다. (사전 상태 측정: 2026-09-29, 트리 `ee81ac32f` — en L100 / ko L106 적중, ja/zh 0 적중.)

### AC-DOCS-005 — 접기 정렬의 내용 보존

- **Given** agent-guide(en/ja/zh 30→22)와 tokenomics의 접기 정렬에 앞서 각 파생 로케일의 섹션 인벤토리를 기록해 두었고
- **When** 정렬 후 인벤토리 diff를 검토하면
- **Then** 사라진 모든 섹션에 처분 기록(병합 / 이동 / ko-부재 확인)이 존재하고, 무처분 삭제는 0건이다.

### AC-DOCS-006 — 래칫 프루닝 전용

- **Given** 세 표적 페이지가 AC-DOCS-001~003을 충족하고
- **When** `docs-site/.locale-parity-baseline`의 diff를 보면
- **Then** diff는 삭제 전용이고, 삭제된 줄은 표적 3페이지의 줄과 정확히 일치하며, 추가된 줄은 0건이다.

### AC-DOCS-007 — 신규 발산 0

- **Given** hns-oss-docs-verify §4 레시피를 실행하면
- **When** 첫 번째 `comm -23`(신규 발산 집합)의 출력을 보면
- **Then** 출력이 비어 있다(새로 발산한 페이지 없음 — 기존 래칫 분은 유지).

### AC-DOCS-008 — verify 게이트 통과

- **Given** 모든 정렬과 래칫 갱신이 끝났고
- **When** `hns-oss-docs-verify` 레시피 전체를 실행하면
- **Then** must_pass 차원이 모두 충족된다: `hugo --minify --gc` exit 0 + WARN/ERROR 0줄 + `public/sitemap.xml` 존재, locale-parity 1.0, version-sync 1.0, Mermaid TD-only grep 무적중, URL 블랙리스트 grep 무적중.

### AC-DOCS-009 — 증거 착지

- **Given** run phase가 완료되었으면
- **When** `.moai/reports/t1328/`을 확인하면
- **Then** canonical-generation verdict(§2 결정 + 근거 커밋 SHA)와 세 페이지의 전/후 헤딩 테이블(고정 카운터 출력 + 섹션 제목 목록)이 존재한다.

## D.1 엣지 케이스

- **펜스 내부 `## ` 줄**: 코드 펜스 안의 `## 예시` 줄도 카운터에 섞인다. 정렬 중 펜스 추가/삭제로 수치가 흔들리면 같은 정의로 재측정해 비교한다 — "펜스라서 안 센다"는 가정 금지.
- **H1 변동**: 카운터는 H1을 제외하므로 페이지 프론트매터 다음 최상위 제목(`# `) 변경은 패리티에 영향을 주지 않는다. 굳이 H1을 맞출 필요는 없다.
- **ja/zh의 신규 사실 취득**: tokenomics 재유도로 ja/zh가 t1115/t1095 사실을 처음 얻는다 — 번역 누락이 없도록 AC-DOCS-004가 로케일별 패턴 grep으로 닫는다(ja/zh 다리는 M2 종료 시점 판정, AC-DOCS-004 선결조건 참조).
- **수렴 실패 페이지**: 정렬 후에도 발산이 남는 페이지가 생기면(예: ko 구조 자체의 결함 발견) 그 페이지는 프루닝하지 않고, 결함을 블로커 리포트로 orchestrator에 돌려 §2 결정의 재확인을 받는다. 래칫에 줄을 추가해 넘어가지 않는다.

## D.2 품질 게이트 기준

- must_pass: AC-DOCS-001~004, AC-DOCS-006~008 (하나라도 FAIL이면 run phase 종료 불가).
- 권고: AC-DOCS-005 처분 기록의 완전성은 sync-auditor 리뷰 대상.

## D.3 Definition of Done

1. AC-DOCS-001~009 전부 PASS (증거 출력 첨부).
2. `moai spec lint SPEC-DOCS-HEADING-PARITY-001` 0 findings.
3. `.moai/reports/t1328/`에 verdict + 전/후 테이블 착지.
4. 래칫 파일이 3줄 줄어든 상태로 커밋 완료(삭제 전용 diff).
