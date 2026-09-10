# SPEC Review Report: SPEC-CI-DOCTOR-BIN-001

Iteration: 1/1 (Tier S ceiling — `harness.plan_audit_tier_ceilings` S=1)
Verdict: PASS
Overall Score: 0.93 (Tier S PASS threshold 0.75 → skip-eligible, artifact-hash unchanged 조건부)

측정 기준점: 브랜치 `WT-ci-doctor-bin` @ `4fdbd55c1`, `bin/moai` 부재 상태(본 감사 세션에서 `ls bin/moai` → No such file 직접 관측). 감사 산출물은 워크트리 절대경로 `.moai/reports/t346/verdict-plan.md`에 기록.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: REQ-CDB-001~004 (`spec.md:66,70,74,78`) — 연속, 공백·중복 0건, 0-padding 일관. 피복표(`spec.md:84-91`)와 1:1 대응.
- [PASS] MP-2 GEARS 형식 (요구 계층 판정): REQ-CDB-001 State-driven "While the embed-axis doctor check is applicable (…) and no readable binary exists (…), the check shall report an informational skip" (`spec.md:66`); REQ-CDB-002 Ubiquitous "The informational skip shall be distinguishable…" (`:70`); REQ-CDB-003 Event-driven "When a readable binary exists…, every existing fail verdict shall be preserved unchanged" (`:74`); REQ-CDB-004 Ubiquitous "The doctor test suite shall pass…" (`:78`). 4/4 패턴 부합. Given-When-Then은 §3 AC(검증 계층)에만 존재 — M3 §Scope 2계층 표 기준 올바른 배치이며 MP-2 위반 아님.
- [PASS] MP-3 YAML frontmatter 유효성: 12 필수 필드 전부 존재·타형 (`spec.md:2-16`) — id/title/quoted version "0.1.0"/status draft/created·updated 2026-08-28/author/priority P1/phase "v3.1.4 target"(전체값 기준 금지 스테이지명 아님)/module internal/cli/lifecycle spec-anchored/tags CSV. snake_case 폐기 별칭(created_at 등) 미사용. 단, 비스키마 여분 필드 1건 → D1 (minor).
- [PASS] MP-4 §22 언어 중립성: N/A — 단일 언어(Go) 프로젝트 스콥, 자동 통과.
- [PASS] MP-5 D7 교차-SPEC 화해: 참조 SPEC은 `SPEC-AGENT-EMIT-LINEAGE-001` 1건(`related_specs` + 본문), 존재 확인, `status: completed`(구 `spec.md:5`) — retired/superseded/archived 아님 → BLOCKING 없음. 부분 대체 선언은 신규 본문에만 존재(`spec.md:68` 대체 조항 + `:117-118` Out of Scope로 구 SPEC 현장 수정 금지) — 지시 조건("old SPEC untouched") 부합. 대체 대상 절의 존재는 직접 확인: 구 `spec.md:82` "**While the judgment point is applicable**, absence of a judgment target is failure, never success." — 신규 SPEC `:37-38` 인용과 전문 일치.
- [PASS] MP-6 D8 크로스플랫폼: SPEC 디렉터리 전체에 `syscall` 0히트(grep rc=1) → D8-4 자동 통과.
- [PASS] MP-7 clarification 게이트: `plan.md`에 `[NEEDS CLARIFICATION` 0히트(grep rc=1); `research.md` 부재(Tier S) → MP-4 선례대로 N/A 구성요소.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 | REQ-CDB-001의 적용가능성이 괄호 술어로 자기완결(`spec.md:66`); 스킵/비활성 구별 근거가 코드 인용으로 명시(`:72`, `uikit/types.go:12-17` — 실측 일치); 각 AC가 잠금 테스트명을 지목(`:93,97`). 요구마다 단일 해석. 서사 정밀도 결함 1건은 D3(요구의 모호성 아님). |
| Completeness | 1.0 | 1.0 | HISTORY(`:21-26`)/WHY(§1)/요구(§2)/AC+피복표(§3)/Out of Scope — `### Out of Scope — <topic>` H3 5개 + 구체 불릿(§4)/제약(§5). frontmatter 완비. HOW는 plan.md(M1/M2·§A.5 PRESERVE·§E 자가검증·§C 사전비행)가 Tier S 규약대로 운반. |
| Testability | 1.0 | 1.0 | 전 AC binary-testable, weasel word 0건. AC-CDB-001은 extractor 미호출 플래그까지 단언(플래그 패턴은 기존 테스트 `doctor_agentemit_embed_test.go:169-171`에서 확립된 기법). AC-CDB-004 RED-now는 본 감사가 독립 재현: 동일 명령 → `--- FAIL: TestRunDoctor_WithExport (5.56s) / runDoctor error: doctor: 1 check(s) failed` — `spec.md:101` 기록과 전문 일치, RED 사유(§1 전파 사슬) 명시, 트리+SHA 핀. |
| Traceability | 1.0 | 1.0 | REQ↔AC 1:1 표(`:84-91`), 미피복 0·고아 0 — 실측 부합. AC-CDB-003의 3개 fail 경로 테스트 존재 확인(`doctor_agentemit_embed_test.go:194,234,258`); 뒤집힐 테스트 `TestAgentEmitEmbed_MissingBinaryFails` `:182` 존재 확인. 교차-SPEC 참조 유효. |

종합: rubric 산술 평균 1.0에서 D1-D3(minor, optional) 3건에 대한 감점 −0.07 → **0.93**. 차단 결함 0건. M6에 따라 optional 목록만으로 감점을 부풀리지 않았고, FAIL 요인은 없다.

## 리드 완료 조건 2종 — AC 기계적 판정 가능성 (감사 초점 2)

- 조건 (i) "부재 시 스킵이 실제 발화": AC-CDB-001(status `ok` + extractor 미호출 플래그 + fail 0 기여) + AC-CDB-002(메시지가 부재 경로 + 처방 토큰을 이름으로) + AC-CDB-004 RED→green. 전부 명령·출력으로 판정 가능. **충족.**
- 조건 (ii) "바이너리 존재 시 여전히 fail 가능": AC-CDB-003(extraction error→fail, drift→fail — 4개 fail 분기 `doctor_agentemit_embed.go:130-161` 전수 열거와 대응, 3개 기존 테스트 무수정 통과) + AC-CDB-004 bin-present leg + plan.md M2 `make build` 대조. **충족.** t317 원 목적(재생성 누락이 임베드 증거를 지우는 것 봉쇄)이 REQ-CDB-003으로 보존됨을 확인.

## AC-CDB-004 two-cell 규율 (감사 초점 3)

- RED-now 셀: 트리+SHA 핀(`WT-ci-doctor-bin` @ `4fdbd55c1` — 본 감사에서 동일값 직접 확인), 명령+관측 출력 전문, RED 사유 명시("이 SPEC이 고치는 바로 그 verdict 때문"). wrong-reason red 아님 — fail 카운트 1건이며 원인 브랜치가 코드상 검증됨(`doctor_agentemit_embed.go:119-124` `os.Stat` 실패 → `CheckFail`).
- Green path 셀: 뒤집는 마일스톤 지목(M1 bin-absent 수렴 + M2 `make build` 대조), 통과 출력 형태 명시.
- 두 셀이 쌍으로 존재 — verification-completeness §2 부합.

## Supersession 건전성 (감사 초점 4)

- 대체 대상 절 존재: 구 `spec.md:82` 전문 확인(위 MP-5).
- 구 SPEC 무변경: `status: completed`(`:5`), 신규 SPEC §4가 현장 수정을 금지.
- 선언 위치: 신규 본문만(`:68` + `:117-118`).
- 잔여 모순 없음: REQ-CDB-001(부재=skip)과 REQ-CDB-003(존재=fail 보존)이 구 REQ-AEL-004의 나머지(적용가능성 술어·not-applicable→ok·verb/doctor 도달·CI 미부착)와 충돌하지 않는다 — 부재 케이스만 조준하는 부분 대체로 선언문이 정확히 한정하고 있다. 검증 계층 파생물에 대한 정밀도 보완은 D2.

## Defects Found

D1. non-schema frontmatter field — `spec.md:16` — `related_specs: [SPEC-AGENT-EMIT-LINEAGE-001]`은 정식 12 필수 필드도, 문서화된 optional 집합(issue_number/depends_on/lint.skip/bc_id/amendment_of/tier)도 아니다. 디코더 구조체(`internal/spec/lint.go:390-423`)에 해당 필드가 없어 조용히 무시된다(빈값 FrontmatterInvalid는 아님 — MP-3에는 영향 없음). — Severity: minor — Class: optional — Required fix: `depends_on:`로 옮기거나 스키마 문서에 필드를 등록. 둘 중 어느 쪽이든 run-phase 진입에 영향 없음.

D2. supersession 선언의 검증-계층 파생 미지목 — `spec.md:68` — 대체 선언이 요구 계층 절만 지목한다. 구 SPEC의 `acceptance.md:49-53`(AC-AEL-003의 "바이너리 부재 — 게이트": `BIN=/nonexistent/moai make embed-check` → exit≠0)은 그 절에서 파생된 검증 조항이며, 신규 행동 하에서 뒤집힌다(verb 경로도 exit 0으로). 라이브 Go 테스트 뒤집기는 명시돼 있다(AC-CDB-001이 `TestAgentEmitEmbed_MissingBinaryFails` 지목 + plan.md B-기대역전/M1 개명). 구 SPEC은 completed·동결이고 embed-check의 CI 부착은 REQ-AEL-004가 금지하므로 살아있는 게이트 충돌은 없다 — 위험은 미래 독자가 t317 수락 기록으로 신규 행동을 재판정하는 것. — Severity: minor — Class: optional — Required fix: `:68` 대체 조항에 한 문장 추가 — "이 대체는 검증 계층 파생물인 AC-AEL-003의 바이너리 부재 게이트(`acceptance.md:49-53`)와 라이브 테스트 `TestAgentEmitEmbed_MissingBinaryFails`의 기대 역전을 포함한다."

D3. CI 전칭 문장 과잉 일반화 — `spec.md:39` — "CI 체크아웃은 `bin/moai`를 절대 갖지 않는다"는 전칭은 `ci.yml`의 `lint` 잡(`:367` `go build -o ./bin/moai`)과 `build` 잡(`:464` `go build … -o bin/moai`)이 각자의 체크아웃에서 정확히 이 경로를 빌드하므로 문자 그대로는 거짓이다. 작동 명제는 옳다 — go test를 돌리는 `test`(`:104`, 커버리지 `:183`)·`test-race`(`:209`, race `:238`) 잡은 바이너리를 빌드하지 않으며, 본 감사가 이 트리에서 RED를 직접 재현했다. 요구·AC·수리 설계에는 영향이 없다(스킵은 부재에서 발화, 판정은 존재에서 — 양방향 견고). — Severity: minor — Class: optional — Required fix: `:39` 문장을 "go test 잡(`test`·`test-race`)은 바이너리를 빌드하지 않는다"로 좁히고 `:367`·`:464` 대조를 각주로.

No blocking defects.

## Gaps (관측하지 못한 것)

- `TestRunDoctor_*` 패밀리 전체(16개 매치 — `coverage_improvement_test.go` 13 + `coverage_test.go` 2 + `target_coverage_test.go` 1)의 개별 RED는 재관측하지 않았다. `TestRunDoctor_WithExport` 1건만 직접 재현했고 나머지는 카드 t346 본문의 9종 기록에 의존한다.
- GREEN 경로는 관측 대상이 아니다(구현 전) — 당연한 미관측.
- `make build` + bin-present 대조(M2 소관)는 실행하지 않았다 — 워크트리에 `bin/`을 쓰는 행위를 감사 범위 밖으로 판단.
- `origin/develop` CI 런 목록을 gh로 재조회하지 않았다 — 카드 본문 + 리드 전달 기록에 의존.
- 교차-백엔드 감사(`audit_multi`)는 이번 감사에 요청·실행되지 않았다(단일 세션 감사).
- `internal/kanban` doctor 참조 0건은 재측정에서 확인했으나(`grep -rln doctor internal/kanban/` → rc=1), 이는 SPEC §4 주장의 근거 확인일 뿐 TestConcurrencyStress 귀속 문제 자체는 범위 밖으로 미해소.

## Residual-risk

- RED 귀속: fail 카운트 1건 + 원인 브랜치 코드 검증 + 카드의 make-build 대조로 귀속됐으나, 어느 검사가 failer인지를 `--check` 필터로 격리해 보지는 않았다. 다른 검사가 기여하더라도 M2의 수렴 명령이 드러낸다 — 낮음.
- `make embed-check` verb 경로의 스킵 시 exit 0은 REQ-CDB-001의 ok 판정에서 함축되지만 SPEC이 verb 경로를 명시하지 않는다(D2와 같은 뿌리) — 낮음.

## Recommendation

PASS. 각 must-pass 근거: MP-1 연속 번호 4건(`:66-78`), MP-2 4/4 GEARS 패턴 부합(요구 계층), MP-3 12필드 완비(`:2-16`), MP-4 N/A(단일 언어), MP-5 구 SPEC completed + 대체 절 전문 확인(`구 spec.md:82`) + 선언의 신규 본문 한정, MP-6 syscall 0히트, MP-7 마커 0히트. 리드 완료 조건 2종 모두 AC에서 기계 판정 가능. Tier S skip-eligible(0.93 ≥ 0.75) — 단, artifact-hash 불변이 깨지면(D1-D3 수리 시) Phase 1 재실행이 정규 경로다.

D1-D3는 optional — run-phase 착수를 막지 않는다. D2+D3는 각 한 문장 수정이므로 run 진입 전 매너 수리로 처리하는 것을 권한다. D1은 스킵해도 무방하다.
