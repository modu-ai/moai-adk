---
id: SPEC-WEB-WRITE-SAFETY-001
created: 2026-09-07
updated: 2026-09-07
---

# Acceptance — SPEC-WEB-WRITE-SAFETY-001 (v0.1.0)

## §A 수용 시나리오 (Given-When-Then)

### AC-WWS-001 — 무저장 기동 시 추적 config 무변경 [부재-가드 · RED-first 필수 · 뮤턴트 필수]

- **Given** 격리 트리에 `.moai/config/sections/`의 스냅샷 복사본이 준비돼 있고
- **When** `moai web`을 기동해 페이지를 열고 Save를 제출하지 않은 채 종료하면
- **Then** `.moai/config/` 하위 모든 git 추적 파일은 스냅샷과 diff 0행이다.

RED-now 셀: M1 실물 재현에서 **커맨드 + 그 출력(verbatim) + exit code + 트리 SHA** 4요소로 기록한다(`verification-completeness.md` §2.1). GREEN 경로 셀: M4 수리가 같은 절차를 diff 0행으로 뒤집는다.

### AC-WWS-002 — GET/탐색 라우트 전부 무쓰기 [부재-가드 · RED-first 필수 · 뮤턴트 필수]

- **Given** 격리 트리와 스냅샷이 준비돼 있고
- **When** GET 라우트 전부(`/`, `/kanban`, `/monitor`, `/todo`, `/settings`, `/specs`, `/events`, `/static/` — 및 라우트 테이블 `internal/web/app.go`의 기타 전 GET 표면)을 순회 요청하면
- **Then** 어떤 요청도 `.moai/config/**` 쓰기를 유발하지 않는다(스냅샷과 diff 0행).

### AC-WWS-003 — 저장 시 비편집 섹션 무변경 [범위 · RED-first 필수]

- **Given** 사용자가 한 섹션의 필드 하나만 바꾸고
- **When** Save를 제출하면
- **Then** 값이 변한 그 섹션 파일 외 나머지 섹션 파일들은 git diff 없음(내용 byte-동일)이다.

RED 근거: 리드 관측 O1이 결함 코드에서 이 시나리오의 실패 형태를 보여주지만, 채택은 M1 트리에서 본 카드가 직접 RED를 관측해 채택한다(carry-over 방지).

### AC-WWS-004 — git-strategy dirty-gate 계약 + 양성 통제 [게이트 · RED-first 필수 · 양성 통제 필수]

- **Given** 세션 내 `SetSection(git_strategy)` 호출 없이 Save가 수행되면
- **When** 저장이 완료되면
- **Then** `git-strategy.yaml`은 byte-동일로 유지된다.

양성 통제(mutant 방향, 필수): 동일 테스트 환경에서 `SetSection(git_strategy)`을 주입하면 재기록이 관측된다. 이 양성 통제가 없으면 본 AC의 GREEN이 "gate가 살아서 지킨 것"인지 "gate이 우회돼 우연히 green인 것"인지 구별 불가능하다 — gate의 비실행이 성공과 구별 불가능한 바로 그 결함 형태다(`verification-completeness.md` §1.3).

### AC-WWS-005 — seam 포맷 충실도 보존 [충실도 · RED-first 필수]

- **Given** 빈 줄·주석·특정 키 순서·Go struct로 모델링되지 않은 키를 포함하는 seam 섹션 파일(예: feedback.yaml)이 있고
- **When** 웹 저장이 그 파일의 스칼라 하나를 바꾸면
- **Then** 대상 키 값 외 표현 요소(빈 줄·주석·키 순서·unknown key)는 golden-file round-trip으로 byte 보존된다.

RED 근거: O1의 빈 줄 삭제가 결함 형태다. M1/M3에서 본 카드가 재관측해 채택.

### AC-WWS-006 — 중복 폼값 명시적 처리 [파서 · RED-first 필수 · 뮤턴트 필수]

- **Given** 동일 `name`을 가진 필드가 2회 제출되는 POST가 있으면
- **When** 파서가 요청을 해석하면
- **Then** 첫 값이 조용히 채택되지 않는다 — 중복 감지(atomic reject 합류) 또는 문서화된 해석 규칙 적용이다.

뮤턴트: 중복을 첫 값으로 처리하는 구(舊) 동작을 재도입하면 본 AC의 가드가 RED가 된다.

### AC-WWS-007 — 가드 품질: 뮤턴트 포착 증명 [가드 품질 · blocking]

- **Given** AC-WWS-001..006의 각 가드가 채택돼 있고
- **When** 결함을 재도입하는 뮤턴트를 각 가드에 적용하면
- **Then** 대응 가드가 RED가 되고, 뮤턴트 이름과 판정 출력이 progress.md §E.2에 기록된다.

**못 잡은 뮤턴트도 남긴다** — 포착 실패 기록은 그 가드의 경계를 그리는 문서다(공허한 초록 방지).

### AC-WWS-008 — M-a/M-b 측정 귀속과 측정-선결 [귀속/게이트 · major]

- **Given** REQ-WWS-008(측정-선결)에 따라 M-a(쓰기 경로 귀속)와 M-b(gate 우회·seam 충실도 원인) 측정이 완료됐고
- **When** 수리(M4)가 설계·구현될 때
- **Then** 수리는 두 측정의 결론을 인용하며, 각 측정은 커맨드 + 관측 출력 + 트리 SHA로 progress.md §E.2에 귀속돼 있다. 두 측정 완료 전에 작성된 수리 코드는 본 AC 위반이다.

---

## §D AC Matrix

| AC | 대응 REQ | 종류 | 심각도 | RED-first | 뮤턴트 | 판정 방법 |
|----|----------|------|--------|-----------|--------|-----------|
| AC-WWS-001 | REQ-WWS-001/002 | 부재-가드 | blocking | **필수 (M1)** | 필수 | 실물 재현 + 스냅샷 diff |
| AC-WWS-002 | REQ-WWS-001 | 부재-가드 | blocking | **필수 (M1)** | 필수 | 라우트 순회 + diff |
| AC-WWS-003 | REQ-WWS-003 | 범위 | blocking | **필수 (M1)** | 권장 | 저장 + git diff |
| AC-WWS-004 | REQ-WWS-004 | 게이트 | blocking | **필수 (M1/M3)** | 양성 통제 **필수** | Save + byte 비교 + SetSection 주입 |
| AC-WWS-005 | REQ-WWS-005 | 충실도 | blocking | **필수 (M1/M3)** | 권장 | golden round-trip |
| AC-WWS-006 | REQ-WWS-006 | 파서 | major | 필수 | 필수 | 중복 폼 POST |
| AC-WWS-007 | REQ-WWS-007 | 가드 품질 | blocking | — | (본 AC가 뮤턴트 검증) | 뮤턴트 실행 기록 |
| AC-WWS-008 | REQ-WWS-008 | 귀속/게이트 | major | — | — | progress.md §E.2 검사 + 수리 설계의 측정 결론 인용 확인 |

"권장" 뮤턴트(003/005)도 수행이 가능하면 수행하고 기록한다. 미수행 시 사유를 progress.md §E.2에 남긴다.

### §D.1 RED-now 셀 4요소 (부재-가드 채택 요건)

`verification-completeness.md` §2.1에 따라, 각 부재-가드의 RED-now 셀은 다음 4요소를 progress.md §E.2에 운반한다:

1. **커맨드** — 단일 호출로 완결되는 판정 명령 (복합 `&&`·파이프 형태 금지)
2. **그 커맨드의 verbatim stdout** — 요약이 아니라 원본 출력
3. **exit code** — 별도 필드로 기록
4. **트리 SHA** — 측정 트리의 커밋 SHA (브랜치 이름 아님)

RED는 "올바른 이유로" red여야 한다 — 무저장 쓰기가 관측돼서 red인 것이지, 셀렉터 무매치·환경 오류로 red인 것이 아니다.

## §D.2 Edge Cases

- **재현 3단계 판별**: (a) 기동만(브라우저 미접속) / (b) 페이지 렌더 / (c) 탐색·폴링 유지 — 어느 단계가 쓰는지가 M-a의 1차 판별 증거다.
- **`/__shutdown__` POST**: 쓰기 라우트이므로 무저장 쓰기 금지 대상에 포함해 확인한다.
- **profile create/delete/rename**: 명시적 동작이므로 허용되지만, 이들이 `.moai/config/` 섹션 파일을 건드리지 않는지 확인한다(profile store는 config 밖).
- **`/static/`·glmkey reveal**: config 쓰기와 무관 표면 — AC-WWS-002 순회에 `/static/`이 포함되고 glmkey reveal도 무쓰기임을 함께 확인한다(REQ-WWS-001 예외 목록에서도 제외 근거).
- **값 불변 저장**: 동일 값 재제출 시에도 섹션 파일이 재기록되면 mtime만 바뀌고 내용은 같다 — 본 SPEC의 판정은 내용 기준 byte-동일(git diff 없음)이며 mtime을 기준으로 삼지 않는다.
- **테스트 격리**: 모든 테스트 fixture는 `t.TempDir()` 아래에 생성한다(프로젝트 루트 오염 금지).

## §D.3 Traceability (REQ → AC)

| REQ | AC |
|-----|-----|
| REQ-WWS-001 (쓰기 시점) | AC-WWS-001, AC-WWS-002 |
| REQ-WWS-002 (탐색 무쓰기) | AC-WWS-001 |
| REQ-WWS-003 (범위 최소화) | AC-WWS-003 |
| REQ-WWS-004 (dirty-gate 계약) | AC-WWS-004 |
| REQ-WWS-005 (포맷 충실도) | AC-WWS-005 |
| REQ-WWS-006 (중복 폼값) | AC-WWS-006 |
| REQ-WWS-007 (회귀 가드) | AC-WWS-001..006 전체 + AC-WWS-007 |
| REQ-WWS-008 (측정-선결) | AC-WWS-008 |

REQ-WWS 전체가 AC 매핑을 가진다. 누락 없음.

## §D.4 Definition of Done

1. 모든 AC PASS — 각 부재-가드는 4요소 RED-now 셀 + green path 셀을 모두 완비
2. M-a/M-b 측정이 progress.md §E.2에 귀속 완료, M4 수리 설계가 결론을 인용
3. M1과 동일한 재현 절차가 수리 후 GREEN
4. git-strategy 양성 통제(SetSection 주입 → 재기록 관측) 관측 완료
5. t509/t510 경계 침범 0
6. primary checkout 무접촉 증명 — 작업 전후 `git status` 비교
7. 변경 패키지(`internal/web`, `internal/config`) 테스트 통과 출력 존재
