# SPEC-REPORT-ARTIFACT-DELIVERY-001 — 구현 계획

> 이 문서는 status 축에서 stateless다 — `status:` 필드를 갖지 않는다. SPEC 수명주기 상태는 `spec.md` 단일 소스만이 운반한다.

## §A Context

- **배차**: 카드 t1427 (Class C, 운영자 지시 2026-10-02) — 리더(팩토리 레인) → manager-spec plan-phase.
- **위치**: 카드 워크트리, 브랜치 `WT-artifact-report-delivery` @ 2e700a15e (배차 시점 develop 선두와 동일, 클린 트리).
- **Tier**: M (측정 근거 — 영향 파일 12-15개: Go 3-4 + 웹 자산 1 + mustache 6 + 마크다운 4-5 + 미러, 규모 300-1000 LOC 밴드). plan-auditor PASS 기준 0.80. REQ 상한 16 중 12 사용, AC 상한 16 중 14 사용.
- **방법론**: `constitution.development_mode: tdd` — M1의 Go 변경은 RED-GREEN-REFACTOR. 콘텐츠·템플릿 변경은 기계 grep 검증 + 템플릿 감사 테스트로 검증한다.
- **기존 인프라**: PRESERVE — html 파일 형식 파이프라인(Pretendard jsdelivr, 등급 게이트 mermaid CDN+noscript), `.moai/reports/` 출력 관례, 6모드×3등급 구조, `report.yaml` 기본값, 라이브 `.agents/skills/` 표면, 기존 i18n 키. EXTEND — `reportFormatValues` 폐쇄 집합, 6개 mustache(아티팩트 블록 추가), SKILL.md 전달 단계, skill-routing.md 문구.
- **변경 메커니즘 관례(템플릿)**: 아티팩트 전용 부분은 각 mustache 안에 `<!-- artifact:begin -->` / `<!-- artifact:end -->` 주석으로 구획한다 — 모델 참조 가독성과 기계 grep 검증(블록 내부 sed 범위 스캔)을 동시에 충족한다.

## §B Known Issues (카드 관련성 필터)

- **B4 프론트매터 정식 스키마** — `created:`/`updated:`/`tags:` 정식 명 사용, snake_case 별칭 금지(적용 완료).
- **B6 spec-lint 표제 관례** — `## Out of Scope` H2 단독은 `MissingExclusions` 오류 — `### Out of Scope — <주제>` H3 + 불릿을 갖췄다(적용 완료).
- **B8 워킹 트리 위생** — 커밋은 명시적 pathspec 스테이징으로만(`git add .moai/specs/SPEC-REPORT-ARTIFACT-DELIVERY-001/`), 런타임 관리 파일(`.moai/state/`, `.moai/harness/`) 무변경.
- **B10 미접촉 경로 PRESERVE** — §A의 PRESERVE 목록 밖 파일 무변경. 특히 라이브 `.agents/skills/`는 건드리지 않는다.
- **미러 드리프트(t1381 교훈)** — 편집은 라이브·템플릿 미러 양측에 같은 변경 세트로 반영하고 `diff -rq` 0을 재생성 입증한다. 한쪽만 고치면 다음 `moai update`에서 반대편이 퇴행한다.
- **i18n 거버넌스** — 신규 키(`f.report.format.opt.artifact`, `f.report.format.option.artifact.desc`)는 미번역 채무 대신 4 로캘(en/ko/ja/zh) 번역을 함께 제공해 거버넌스·미번역 허용목록 테스트를 무충돌로 통과시킨다.

## §C Pre-flight (구현 착수 전 1회 배치)

```bash
git branch --show-current          # WT-artifact-report-delivery
git rev-parse --short HEAD         # 배차 기준과 비교
go build ./...                     # 베이스라인 빌드
go test ./internal/settings/ ./internal/web/ ./internal/cli/...   # 베이스라인 그린
diff -rq .claude/skills/moai-domain-html-report internal/template/templates/.claude/skills/moai-domain-html-report   # 시작 패리티 (2026-10-02 실측 OK)
wc -c .claude/skills/moai-domain-html-report/SKILL.md              # 22,516 (예산 베이스라인)
```

RED-now 기준(2026-10-02, 2e700a15e 실측 완료 — acceptance.md §D 표 참조): `reportFormatValues` 2값, i18n artifact 키 0, 위저드 옵션 2종, 라우팅 결합 문구 2값, 템플릿 다크모드 블록 0, SKILL.md 전달 단계 auto-open 단일.

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE**: §A 목록 전체 — 특히 아티팩트 블록 밖의 Pretendard jsdelivr 링크와 등급 게이트 mermaid CDN+`<noscript>`는 html 파일 형식의 생명선이므로 제거·변경 금지.
- **금지 명령**: `--no-verify`, `--amend`, main 강제 push. 커밋은 Conventional Commits + 본문에 `(card t1427)` + 최종 트레일러 `🗿 MoAI`.
- **템플릿 중립성**: `internal/template/templates/**`에 카드 토큰(t1427)·SPEC ID·내부 날짜 삽입 금지 — AC-RAD-012가 grep으로 검증.
- **블록 관례**: 아티팩트 전용 내용은 반드시 `artifact:begin`/`artifact:end` 구획 안에 — 블록 밖 오염은 html 경로 훼손이며 AC-RAD-006/008이 적발한다.
- **스킬 교체 금지**: 모드·등급·md 쌍둥이 소유 이동, html+md 기본 경로의 철거는 REQ 위반이다.

## §E Self-Verification (완료 보고 시 필수 산출)

- **E1** AC-RAD-001~014 PASS/FAIL 행렬 — 각 행에 명령 + 축어 출력 (verification-claim-integrity §3 5절 형식: Claim/Evidence/Baseline-attribution/Gaps/Residual-risk).
- **E2** `go build ./...` exit 0.
- **E3** 영향 패키지 커버리지 — `go test -cover ./internal/settings/ ./internal/web/` (기존 커버리지 회귀 없음).
- **E5** `golangci-lint run --timeout=2m` — 신규 이슈 0 (기존 베이스라인과 분리 표기).
- **E6** 커밋 SHA 목록 + push 상태(레인 규율상 push 없음 — 로컬 병합 SHA 보고만).
- **E8** M1 TDD RED 축어 출력(구현 전 실패 테스트 출력) — test-first 위변조 불가 장치.

## §F Milestones (결정 가역성 순 — 바뀔 가능성이 큰 결정부터)

### M1 (Priority High) — `report.format=artifact` 폐쇄 집합 확장 + UI 전파 [REQ-001]
- `internal/settings/schema_sections.go` — `reportFormatValues`에 `artifact` 추가, `reportFields()` 옵션 desc 매핑에 `f.report.format.option.artifact.desc` 분기.
- `internal/web/assets/i18n.js` — 4 로캘 블록 각각 `f.report.format.opt.artifact` + `f.report.format.option.artifact.desc` 2키 추가(번역 제공 — §B i18n 거버넌스).
- `internal/cli/wizard/questions.go` — `report_format` 옵션에 artifact 추가(SSOT 동기 주석 유지).
- 테스트 — schema_sections·schemaform·i18n 거버넌스가 요구하는 갱신. **TDD: RED**(artifact 부재를 단언하는 테스트 선작성·실패 출력 확보) → **GREEN** → REFACTOR.
- 산출: AC-RAD-001, 002, 003. 근거: 폐쇄 집합은 데이터 모델·사용자 대면 변경으로 피드백 가능성이 가장 커서 선두에 둔다.

### M2 (Priority High) — SKILL.md 아티팩트 전달 단계 + 계약 참조 문서 [REQ-002, 003, 004, 006, 010 지침 축]
- Input 섹션 — `report.format` 판독 독트린(라우팅 결합과 정렬): `artifact` 값 의미와 판독 주체.
- "After rendering" 섹션 3분기 개편 — (a) artifact 게시: 링크 제시·auto-open 생략, (b) 폴백: Artifact 도구 부재(Codex/GLM/API-key) → html+md 경로 + 사유 명시, (c) 재게시: 동일 파일 경로 재게시로 URL 유지.
- Diagram Policy — artifact 분기: `pre.mermaid` 코드 블록 출력, CDN 스크립트 금지(뷰어 사전 렌더가 폴백을 대체).
- `references/artifact-contract.md` 신설 — 페이지 계약 요약(§D1 전체), `<title>` 명명 규칙, 다크모드 2블록 스니펫, 16px 거터 규격, CSP 호스트 표, 16MB 상한, **폰트 결정 근거(아래 ③)**.
- Non-goals — artifact는 추가 전달 형식임을 명시.
- 산출: AC-RAD-005, 007, 009(SKILL 축).

### M3 (Priority High) — 6개 모드 템플릿 아티팩트 블록 + fonts.md [REQ-005, 007, 008, 009, 010 템플릿 축]
- 각 mustache에 `artifact:begin/end` 블록 추가 —
  - Google Fonts 전용 링크 세트: Noto Sans KR(본문) + Noto Serif KR(세리프 모드) + JetBrains Mono(코드) + preconnect.
  - 다크모드 2블록: 모드 팔레트의 다크 대응값, `:root:not([data-theme="light"])` 가드 `@media` 블록 + `:root[data-theme="dark"]` 블록, 양 스킴 body 배경.
  - 16px 측면 거터 미디어쿼리(폰 폭, 수평 스크롤 없음).
  - mermaid 방침 주석: `pre.mermaid` 코드 블록 — CDN 스크립트 없음.
- 블록 밖(html 파일 형식) 무변경 — Pretendard jsdelivr·기존 CDN 유지.
- `references/fonts.md` — 아티팩트 형식 폰트 매핑 절 추가.
- 블록 경계 주의 — explainer 템플릿의 기존 인라인 `<script>`(2곳, CDN 아님)는 아티팩트 블록으로 흡수하지 않는다(1차 감사 operational note). AC-RAD-008의 pretendard 보존 핀치(explainer 0 · plan 3 · financial/incident/pr/status 각 4 — acceptance.md §E EV-12)도 M3 편집이 지켜야 할 기준선이다.
- **③ 폰트 최종 결정**: artifact 형식은 **Google Fonts Noto 계열**(Noto Sans KR/Noto Serif KR + JetBrains Mono)로 조달한다. 근거 — (1) 계약이 스타일시트 호스트를 Google Fonts로 제한하는데 Pretendard는 Google Fonts 카탈로그에 없다(fonts.md 실측); (2) @font-face 인라인은 `data:` URI로 16MB 상한을 소모하며 한국어 풀패밀리는 수 MB급으로 비현실적; (3) Noto Sans KR·Noto Serif KR·JetBrains Mono는 스킬이 이미 사용 중인 패밀리라 신규 CDN 관계가 0이다(fonts.md 매핑표 실측). 기각 — Pretendard 유지(jsdelivr 차단), Pretendard 자체호스팅 data: URI(용량).
- 기각된 메커니즘 대안 — (a) 아티팩트 전용 템플릿 6종 신설: ~100KB 중복·이중 유지보수 드리프트 위험; (c) 단일 오버레이 문서만 추가: 모델의 심적 병합 오류 위험, 카드 문언("6개 모드 템플릿을 계약에 맞춤")과 검증 용이성에서 블록 방식이 우위.
- 산출: AC-RAD-006, 008, 009(템플릿 축).

### M4 (Priority Medium) — skill-routing.md 개정 [REQ-011]
- :38 결합 라인 — 3값 나열(`html+md` | `md` | `artifact`) + artifact의 의미(아티팩트 게시 경로) 서술.
- :42 안티패턴 문구 개정 — "콘텐츠·렌더링은 moai-domain-html-report, 아티팩트 게시 계약은 artifact-design" 분리 서술로 대체. format=artifact 전달 단계에서 artifact-design의 게시 계약을 참조하는 것은 정상 라우팅임을 명시한다.
- 산출: AC-RAD-004. 근거: 독트린 문구는 산출물이 정의된 뒤 손질하는 것이 충돌이 적어 후순위.

### M5 (Priority Low — 기계적 마무리) — 미러 동기화 + 중립성 + 예산 + 빌드·테스트 [REQ-012]
- M1-M4의 모든 라이브 편집을 `internal/template/templates/` 동일 경로에 반영 → `diff -rq` 0 재생성 입증(라이브 스킬·skill-routing 양측).
- 템플릿 중립성 grep(카드 토큰 0히트) + `go test ./internal/template/` + `TestHtmlReportOutputPathParity` 그린.
- `wc -c` SKILL.md ≤ 26,000B 확인(대형 추가는 references/로 — §B 지침 예산).
- `go build ./...` + `go test ./internal/settings/ ./internal/web/ ./internal/cli/...` 최종 그린.
- 산출(최종 재검증): AC-RAD-010, 011, 013 — AC-RAD-012·014는 항시 유지 조건의 재입증. 전환 시점 귀속의 SSOT는 acceptance.md §D 표다(1차 감사 D5 정렬).

## §G Anti-Patterns

- 미러 한쪽만 편집( t1381 재발 ) — M5에서 `diff -rq` 0으로 강제 입증.
- `report.format=artifact` + 도구 부재를 오류 처리(REQ-002 위반) — 폴백이 기본 동작이다.
- 아티팩트 블록 밖에서 Pretendard·mermaid CDN 제거 — html 파일 형식 훼손.
- 스킬 교체 성격의 변경(모드/등급/md 쌍둥이 소유 이동) — 카드 명시적 결정 위반.
- i18n 키 추가 후 거버넌스 테스트 미갱신 — 4 로캘 번역을 함께 제공.
- 템플릿에 카드 내력 삽입 — 중립성 가드 적색.

## §H Cross-References

- `spec.md` §C 요구사항·§D 제약·§E 가정 / `acceptance.md` §D AC 행렬 / `decision-index.md` Q1-Q3
- `.claude/skills/moai-domain-html-report/references/fonts.md` — 폰트 매핑 SSOT(변경 대상)
- `.claude/rules/moai/workflow/skill-routing.md` §1.1 — 결합 문구(변경 대상)
- `internal/template/html_report_output_path_parity_test.go` — 출력 관례 가드(존중 대상)
