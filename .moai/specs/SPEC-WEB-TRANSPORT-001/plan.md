---
id: SPEC-WEB-TRANSPORT-001
title: "plan — validation-400 boosted-response transport discriminator"
version: "0.1.0"
created: 2026-09-22
author: manager-spec
---

## §A Context

카드 t1080 (class C, plan→run→sync). 근거: t1051 sync-audit F5 + Residual-risk(3). 검증 거부 400 경로(`internal/web/handlers.go:490`)의 hx-boost 수송 결함을 측정-선행-판정-후행으로 다룬다. 본 트리에서 측정 완료된 전제(판단 아닌 측정):

- htmx 는 로컬 임베드(50,917 바이트, `/static/htmx.min.js` 서브, CDN 0) — `internal/web/assets.go:20-24`.
- 핀 버전 2.0.4, 자산 내부에 `responseHandling` 식별자 1회 존재.
- 검증 거부 경로는 배너 + 필드별 오류 풀페이지를 400 으로 렌더(`handlers.go:483-491`) — 본문 포함 여부는 미증명 주장.
- 기존 선례: `internal/web/htmx_test.go` 가 임베드 자산을 Go 테스트에서 grep 하는 패턴이 이미 존재 — 자산 추출 검사는 재사용 가능한 관례다.

## §B Known Issues (측정 전 리스크)

1. **400 본문 피드백 주장은 미증명** — REQ-TR400-001 이 먼저 측정으로 승격해야 판정의 "서버는 피드백을 내보냈다" 쪽이 성립한다.
2. **minified 리터럴 추출 취약성** — 임베드 자산은 minified 이므로 기본 `responseHandling` 테이블의 정확한 문자열 형태가 최종 확인 전까지 가정이다. 추출기는 정규식 단일 패턴에 고정하지 말고, 추출 실패 시 **판정 이연**(REQ-TR400-003 deferred)로 탈출하는 경로를 반드시 갖는다. "추출 못 함"은 실패가 아니라 정당한 판정 등급이다.
3. **문서-자산 불일치 가능성** — htmx 공식 문서의 기본값(4xx → swap:false)이 2.0.4 자산 내부 기본값과 다를 수 있다. 자산 추출값이 우선이다.

## §C Pre-flight

- Base: `0314801c2` (= origin/develop = 로컬 develop, 정렬 확인 완료). Branch: `WT-verify-400-path`.
- 브라우저 불요 — 계약 수준 측정 경로만 사용.
- t1051 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1051/`)는 읽기 전용 참조 — 쓰기 금지.
- 본 SPEC run 단계는 운용 코드 수정 없음 — 테스트 파일과 판정 기록만 산출.

## §D Constraints

- 모든 테스트 임시 디렉터리 `t.TempDir()`.
- 전체 스위트 금지 — 영향 패키지만, 환경스크럽 단일 compound 호출:
  `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/web/...`
- 판정 기록 위치: `.moai/reports/t1080/verdict.md` (카드 증거 경로 관례).
- 커밋은 레인이 수행 — 본 SPEC 산출물만 `.moai/specs/SPEC-WEB-TRANSPORT-001/` 아래.

## §E Self-Verification

- [ ] AC-TR400-001..005 가 acceptance.md 의 Given-When-Then 으로 이원적 검증 가능하게 쓰였는가
- [ ] 판별식이 "본문 문자열 포함 검사"로 축소되지 않았는가 (본문 프로브는 D-400a 특성화일 뿐, 판별식은 D-400b 스왑 결과다)
- [ ] 판정 절차가 [HARD] 측정-선행(REQ-TR400-002)과 [HARD] 비소급(REQ-TR400-005)을 모두 충족하는가

## §F Milestones (결정-가역성 순 — 바뀔 가능성이 큰 결정이 앞에 온다)

### M1 — D-400b 측정 하네스 (임베드 htmx 계약 추출) [Priority High]

`internal/web/htmx_test.go` 선례를 따라 임베드 자산에서 기본 응답 처리 테이블을 추출하고, 상태 코드 패턴 "400" 에 대한 스왑 결과를 평가하는 기계 검사를 만든다. 추출에 성공하면 매칭된 테이블 조각 **전문**을 증거로 기록하고, 실패하면 `deferred-to-browser` 등급으로 탈출한다(REQ-TR400-003). 이것이 본 SPEC 에서 가장 바뀔 가능성이 큰 설계 결정이다(추출 전략 자체) — 먼저 수행해 조기 확정한다.

### M2 — D-400a 특성화 테스트 (400 본문) [Priority High]

httptest 로 검증-실패 POST 를 서브밋하고, 상태 400 + 배너 문구 + 최소 1개 필드별 오류 메시지의 본문 포함을 단언한다(REQ-TR400-001). 현재 코드가 400 을 유지하는 한 이는 특성화(characterization) 테스트다 — 수리 카드가 나오면 그때 명세 테스트로 전환된다.

### M3 — 판정 기록 작성 [Priority Medium]

`.moai/reports/t1080/verdict.md` 에 §2 REQ-TR400-003 사상표에 게이트된 판정 + REQ-TR400-004 신뢰도 등급 + 두 측정 출력 전문 인용을 기록한다. M1·M2 측정이 모두 관측된 뒤에만 작성한다([HARD] 측정-선행).

### M4 — 카드 분할 경계 확인 [Priority Low]

판정 기록에 t1080/t1081 분할(spec.md §1.2)이 유지됐는지 — 본 SPEC 이 브라우저 CDP 범위를 삼키지 않았는지 — 확인하고 한 줄 기록한다. 기계적 확인 단계.

## §G Anti-Patterns

- **본문 문자열 프로브를 판별식으로 쓰기** — D-400a(본문 포함)는 서버 측 특성화일 뿐이다. 판별식은 클라이언트 스왑 결과 D-400b 다. (t1081 쪽 [HARD] 와 같은 축의 경계)
- **문서만으로 판정** — htmx 공식 문서의 기본값을 인용하더라도 자산 추출 없이 판정을 쓰면 관측되지 않은 주장이다.
- **t1051 판정 소급 적용** — "500 이었으니 400 도 결함이다"는 측정이 아니라 추론이다. 판정은 자기 측정에만 근거한다.
- **판정 선작성 후 측정 보강** — REQ-TR400-002 (b) 항 위반.

## §H Cross-References

- SPEC-WEB-CONSOLE-017 — t1051 수송 결함 수리 (감사 종료, 재개 금지)
- 카드 t1081 — 브라우저 CDP 실측 (REQ-A 갭, 본 SPEC 과 분할 경계는 spec.md §1.2)
- `.moai/reports/t1051/sync-audit.md` §F5, Residual-risk(3) — 발단 근거 (t1051 워크트리 내, 읽기 전용)
- `internal/web/htmx_test.go` — 임베드 자산 검사 선례 (M1 이 재사용하는 관례)
