# SPEC-DOCS-SITE-001 — G1 Plan Audit Report (Iteration 1)

- **Auditor**: plan-auditor (independent, adversarial mode)
- **Audit Date**: 2026-04-17
- **Audit Inputs**: spec.md (222 LOC), plan.md (370 LOC), acceptance.md (541 LOC)
- **Cross-reference**: migration-inventory.md (825 LOC), spec-project-diff.md (562 LOC)
- **Bias Protocol**: M1 Context Isolation active (author reasoning NOT consumed), M2 Adversarial stance, M4 Evidence citation, M5 Must-Pass firewall, M6 Chain-of-Verification pass executed.

---

## 1. Executive Summary

**Verdict: FAIL** — 수정 후 재심사 필요.

**한 줄 근거**: SPEC이 스스로 정의한 자동 검증 스크립트(AC-G1-02)에 의해 FAIL 판정되고(REQ 수 34 > 상한 30), REQ-DS-32의 Phase 7 시퀀싱이 논리적으로 불일치한다(Vercel 재바인딩과 G4 승인의 선후 관계 모순).

**판정 기준 충족도**:
- Critical 결함 2개 (≥ 1개 → FAIL 조건 만족)
- High 결함 6개 (≥ 6개 → FAIL 조건 만족)
- Medium 11개, Low 6개

판정 기준표:

| 조건 | 결과 |
|------|------|
| Critical 0 + High ≤ 2 → PASS | 해당 없음 |
| Critical 0 + High 3-5 → CONDITIONAL PASS | 해당 없음 |
| Critical ≥ 1 OR High ≥ 6 → FAIL | **트리거됨 (2 Critical, 6 High)** |

---

## 2. 발견된 결함 (Defects)

### CRITICAL (2건)

#### D-001. REQ 개수가 SPEC 자체 상한을 초과 — 자동 검증 실패 확정

- **위치**: spec.md §6 (REQ-DS-01 ~ REQ-DS-34) vs acceptance.md AC-G1-02:L41 vs plan.md G1 기준 §7:L320
- **기대**: REQ 개수 15 ≤ N ≤ 30 (acceptance.md), 또는 15 ≤ N ≤ 25 (plan.md)
- **실제**: REQ-DS-01 ~ REQ-DS-34 총 **34개** (§6.1: 3건, §6.2: 5건, §6.3: 3건, §6.4: 4건, §6.5: 4건, §6.6: 4건, §6.7: 4건, §6.8: 1건, §6.9: 3건, §6.10: 3건)
- **영향**: acceptance.md AC-G1-02 line 41의 awk 스크립트(`$1 >= 15 && $1 <= 30`)가 `34 > 30`이므로 **exit 1 (실패)**. SPEC이 자기 자신의 G1 자동 검증을 통과할 수 없다.
- **권장 수정**:
  1. AC-G1-02 상한을 30 → 35로 상향, 또는
  2. plan.md G1 범위 `15~25`를 `15~35`로 상향 조정 (acceptance.md도 일치시킴), 또는
  3. 요구사항을 통합·감소시켜 30개 이하로 맞춤(비추천 — 내용 손실).
  추천: (1)+(2) 동시 적용. 이유: 34개는 범위가 큰 마이그레이션으로 합리적이며, 억지 통합은 추적성을 해친다.

#### D-002. REQ-DS-32와 Phase 7 시퀀싱 모순 — Vercel 재바인딩 시점 오류

- **위치**: spec.md REQ-DS-32:L172 vs plan.md Phase 7 워크플로우:L228~247 vs acceptance.md AC-G4-01~08:L285~411
- **기대**: Vercel Project 재바인딩이 Preview 검증 전에 완료되어야 AC-G4-01~08 체크리스트(4 locale 홈 + Edge Function 실환경 3시나리오)를 Preview URL 대상으로 수행할 수 있다.
- **실제**: REQ-DS-32는 "**When** the Phase 7 manual approval gate G4 **is granted**, the operator **shall reconfigure** the existing Vercel project..."로 서술 — G4 승인 **후에** 재바인딩한다는 Event-driven EARS 패턴.
  그러나 plan.md Phase 7 단계는 순서가: (1) DNS TTL, (2) **Vercel 재바인딩**, (3) Preview 검증, (4) **G4 승인**, (5) Production 프로모션. 즉 재바인딩이 G4 **이전**에 일어난다.
- **영향**: 
  - REQ와 plan의 이벤트 순서 불일치는 구현 단계에서 해석 충돌을 야기한다.
  - Preview 검증(AC-G4-01~08)은 재바인딩이 완료된 상태에서만 가능하므로, REQ-DS-32의 트리거를 "When G4 is granted"로 둘 수 없다.
- **권장 수정**: REQ-DS-32를 두 개로 분할:
  - REQ-DS-32a (Event-driven): "When Phase 7 enters the cutover preparation step, the operator shall reconfigure the existing Vercel project `prj_EZaVdfE3gJeXVbizafBEECpniINP` with: (a) Git source → modu-ai/moai-adk-go, (b) Root Directory → docs-site, (c) Framework Preset → Hugo, (d) Ignored Build Step → `git diff --quiet HEAD^ HEAD ./docs-site/ || exit 1`."
  - REQ-DS-32b (Event-driven): "When acceptance.md § G4 checklist passes and GOOS grants G4 approval, the operator shall promote the current Preview deployment to Production."
  - DNS TTL 관련 조항(기존 (e))은 D-004와 함께 재검토.

---

### HIGH (6건)

#### D-003. Phase 2 gate 누락 — plan.md §7에 게이트 기준 정의 없음

- **위치**: plan.md §3 Phase 개요 테이블:L24~36, §7 체크포인트:L316~345
- **기대**: Phase 2(Scaffold + Git Subtree) 완료 판정을 위한 게이트 기준 존재.
- **실제**: §3 테이블은 Phase 2 gate를 "G1 (계속)"으로 명시. 그러나 §7에 G1의 정의는 "3-file set 존재 + EARS 수 + D1~D4 이중 명시 + plan-auditor 리뷰 3건 이하"뿐이다. Hugo 빌드, 디렉토리 구조, Nextra 잔재 제거 등 Phase 2의 기술적 완료 기준이 없다.
- **영향**: Phase 2 완료 판정이 불가능. G1(문서 심사)과 Phase 2(scaffold 구현)는 본질이 다른 작업이므로 동일 게이트를 공유할 수 없다.
- **권장 수정**: 새 게이트 G1.5 또는 G2-a 신설:
  ```
  Gate G1.5 (Phase 2 → Phase 3):
  - `docs-site/` 디렉토리 존재
  - `git log --oneline -1 -- docs-site/` squash 커밋 단일 기록 확인
  - `docs-site/hugo.yaml` 존재 + baseURL/languages 정의
  - `docs-site/package.json`, `bun.lock`, `next.config.mjs`, `theme.config.tsx`, `app/`, `components/`, `lib/`, `middleware.ts` 등 Nextra 잔재 0건
  - `cd docs-site && hugo server` 에러 없이 기동 (콘텐츠 0건 상태)
  ```

#### D-004. DNS TTL 300초 하향 (REQ-DS-32(e)) 필요성·리드타임 미검증

- **위치**: spec.md REQ-DS-32:L172 (e), plan.md Phase 7 step 1:L230
- **기대**: DNS 변경이 실제로 필요한지 근거 + 현재 TTL 실측값 + 하향 리드타임 계산
- **실제**: 
  - Exclusions(spec.md:L198)에 따라 "신규 Vercel 프로젝트 생성 금지" — 기존 프로젝트 ID를 재바인딩만 수행.
  - Vercel 프로젝트 내부의 Git 소스 변경은 DNS 레코드(NS/CNAME)를 건드리지 않는다. 도메인 바인딩은 프로젝트에 귀속되므로 **DNS 변경 자체가 불필요할 가능성이 크다**.
  - 현재 `adk.mo.ai.kr`의 TTL 값이 문서에 없다. 만약 현재 TTL이 86400s(24h)라면 "24h 전 하향"만으로는 전파가 완료되지 않는다.
- **영향**: 불필요한 DNS 조작은 서비스 중단 리스크만 키운다. 반대로 실제로 필요한 경우 리드타임이 부족할 수 있다.
- **권장 수정**:
  - Phase 7 사전 조사 AC 추가: "현재 `dig +short TTL adk.mo.ai.kr` 결과를 기록하고, DNS 변경 필요성을 Vercel 지원팀에 문의한 결과를 `.moai/plans/DOCS-SITE/phase-7-dns-assessment.md`에 기록한다."
  - 필요성 확정 시, TTL 하향 리드타임을 `max(24h, current_TTL)`로 수정.
  - 불필요 확정 시, REQ-DS-32(e)를 전면 제거.

#### D-005. REQ-DS-09 ("preserve all 569") vs AC-G2-06 (±5 tolerance) 불일치

- **위치**: spec.md REQ-DS-09:L112 ("preserve all 569 Mermaid code blocks"), acceptance.md AC-G2-06:L156~158 ("569 ± 5")
- **기대**: REQ가 "all"(절대 보존)을 요구하면 AC도 정확히 569을 검증해야 함.
- **실제**: AC는 564~574 허용. 최대 5개 손실까지 Gate 통과 가능.
- **영향**: 콘텐츠 손실이 감지되지 않고 프로덕션에 배포될 수 있음. "all 569"를 주장하는 REQ의 의도와 모순.
- **권장 수정**:
  - 보수적(추천): AC-G2-06을 `test "$COUNT" -eq 569`로 변경.
  - 또는 REQ-DS-09를 "shall preserve at least 564 Mermaid code blocks (tolerance ±5 for formatter whitespace normalization)"으로 약화하고 이유를 명시.

#### D-006. REQ-DS-05 (735건) vs AC-G2-04 (≥700, 약 5% 손실 허용) 불일치

- **위치**: spec.md REQ-DS-05:L102 ("all 735 instances"), acceptance.md AC-G2-04:L127 (`test "$TOTAL" -ge 700`)
- **기대**: 735건 절대 보존 → AC는 정확히 735 또는 해석 가능한 범위 내 허용 범위의 근거 필요
- **실제**: AC는 700 이상만 요구. 35건(약 5%) 손실 허용. 근거 주석 없음.
- **영향**: Callout 35건 손실이 CI를 통과. 사용자 경험 저하 가능.
- **권장 수정**: AC-G2-04를 `test "$TOTAL" -eq 735` 또는 `test "$TOTAL" -ge 730`으로 강화, 그리고 `<Callout ` JSX 잔재 0건 확인(이미 있음) 유지.

#### D-007. Vercel Edge Function 구현 명세 불완전

- **위치**: spec.md REQ-DS-13:L122, plan.md Phase 5 step 1:L162~166, acceptance.md AC-G3-01:L180~197
- **기대**: Vercel Edge Function 임을 런타임 수준에서 보장하는 명세 (platform constraint 수용 포함)
- **실제**: 
  - `docs-site/api/i18n-detect.ts` 경로만 지정. 기본적으로 Vercel은 `/api/*.ts`를 **Serverless Function**(Node.js)으로 처리한다. Edge Function으로 만들려면 파일 상단에 `export const config = { runtime: 'edge' };` 또는 `export const runtime = 'edge';`가 명시되어야 한다. SPEC에 이 요구가 누락됨.
  - Edge Function의 실제 플랫폼 제약(실행 시간 제한, 메모리 제한, 콜드 스타트 지연, npm 패키지 제약)이 검토되지 않았다.
  - `middleware.ts` 164 LOC 가 단순 path/header/cookie 로직이므로 Edge 제약을 초과할 가능성은 낮으나, 확인이 문서화되지 않음.
- **영향**: Phase 5 구현 시 런타임 타입 오분류로 재작업 발생 가능.
- **권장 수정**:
  - REQ-DS-13에 `runtime: 'edge'` 명시 문구 추가.
  - plan.md Phase 5에 "Edge Function platform constraint 사전 검토 단계" 추가 (실행 시간·메모리·npm 패키지).
  - acceptance.md에 AC 추가: "Deploy된 API가 Edge runtime임을 Vercel Dashboard 혹은 `curl -I` 응답 헤더(`x-vercel-cache`, `server`)로 확인."

#### D-008. G1 EARS 범위 plan.md(15~25) vs acceptance.md(15~30) 불일치

- **위치**: plan.md §7 G1 기준:L320, acceptance.md AC-G1-02:L41
- **기대**: 동일한 게이트 기준이 문서 간 일치.
- **실제**: plan.md "15~25개", acceptance.md "15 ≤ N ≤ 30".
- **영향**: 어느 값이 바인딩인지 불명확. D-001과 연계되어 SPEC 자체 자가 모순.
- **권장 수정**: 한 값으로 통일(D-001 해결안에 따라 15~35 권장).

---

### MEDIUM (11건)

#### D-009. 다수 REQ의 AC 커버리지 누락

- **위치**: 
  - REQ-DS-01 (git subtree) — 직접 AC 없음
  - REQ-DS-02 (디렉토리 구조) — 직접 AC 없음
  - REQ-DS-07 (`_meta.ts` 변환 + display:hidden 의미론) — 직접 AC 없음
  - REQ-DS-08 (Go 스크립트 경로) — 직접 AC 없음
  - REQ-DS-10 (Mermaid 전략 문서화) — AC 없음
  - REQ-DS-11 (Optional lazy-load) — AC 없음 (Optional 이라 허용 가능)
  - REQ-DS-16 (버전 URL 구조) — 직접 AC 없음
  - REQ-DS-19 (v2.X 배너) — 직접 AC 없음
  - REQ-DS-20 (커스텀 partial 3종) — 직접 AC 없음
  - REQ-DS-27 상세 (Hugo v0.140+, Mermaid v11+, FlexSearch) — AC-G3-06은 "Hugo"와 "Hextra" 키워드 존재만 확인
  - REQ-DS-29 (forbidden URL CI) — AC 없음
- **영향**: REQ당 AC 매핑 불완전 — "모든 REQ에 대응 AC" Definition of Done(acceptance.md:L517) 조항과 직접 모순.
- **권장 수정**: 누락된 REQ별 AC 추가. 특히 REQ-DS-19(배너), REQ-DS-20(partial 3종), REQ-DS-29(CI 규칙)는 자동 검증 가능하므로 추가 필수.

#### D-010. v2.12.0 스냅샷 타이밍 모호성

- **위치**: spec.md §4 용어집:L69, §5 D2:L84, REQ-DS-17:L132
- **기대**: v2.13.0 태깅 시점에 "현재 `content/{locale}/` 트리"를 v2.12 폴더로 복사한다. "현재"가 어느 상태인지 명확해야 함.
- **실제**: v2.13 태깅 시 `content/{locale}/`는 v2.13 콘텐츠 업데이트가 이미 반영된 상태일 수 있음(문서가 v2.13 기능을 반영해 먼저 머지되고 태그가 나중에 붙는 워크플로우). 스냅샷은 "v2.12 최종 상태"여야 하는데, REQ-DS-17은 "copies the current content tree"라고 기술. 논리 구멍.
- **영향**: 잘못된 시점에 스냅샷을 뜨면 "v2.12" 폴더 내용이 실제로는 "v2.13 일부 반영"이 되어 역사적 참조성 훼손.
- **권장 수정**: REQ-DS-17에 스냅샷 대상 커밋 명시: "the release automation shall invoke the snapshot script against the commit tagged as the previous release (e.g., the commit tagged v2.12.X), not HEAD of main."

#### D-011. AC-G3-04 테스트 케이스 이름 오류

- **위치**: acceptance.md AC-G3-04:L228~230
- **기대**: 테스트 이름이 실제 릴리스 타입과 일치.
- **실제**: 
  - `TestMajorRelease: v2.12.0 → v2.13.0` — 실제로는 **Minor** 릴리스 (2.12 → 2.13).
  - `TestMinorRelease: v2.12.1 → v2.12.2` — 실제로는 **Patch** 릴리스 (2.12.1 → 2.12.2).
- **영향**: 구현자가 테스트 의도를 오해할 수 있음. 의미론과 이름 불일치.
- **권장 수정**:
  - `TestMinorRelease: v2.12.0 → v2.13.0 입력 시 content/{locale}/v2.12/ 생성`
  - `TestPatchRelease: v2.12.1 → v2.12.2 입력 시 스냅샷 미생성 (patch 감지)`
  - 가능하면 `TestMajorRelease: v2.12.0 → v3.0.0` 케이스 추가.

#### D-012. 플래그 이모지 → 텍스트 전환과 "디자인 리디자인 금지" 상충

- **위치**: spec.md §8 Exclusions:L196 ("디자인 리디자인 금지"), plan.md Phase 4 step 1:L133
- **기대**: 디자인 유지 vs 이모지 제거 원칙의 충돌이 명시적으로 조정됨.
- **실제**: 
  - 기존 moai-docs는 LanguageSelector에 국기 이모지를 사용 (migration-inventory.md §6.2.3).
  - coding-standards.md는 docs-site/layouts, content, i18n에서 이모지 금지 (spec.md §7 제약:L185).
  - plan.md Phase 4: "플래그 이모지 대신 `KO / EN / JA / ZH` 텍스트" — 시각적 변경.
  - Exclusions "디자인 리디자인 금지"와 개념적으로 상충. 정당한 변경이지만 문서에 명시적 조정이 없음.
- **영향**: 이해관계자가 Phase 4 리뷰 시점에 "왜 디자인이 바뀌었나?" 이슈 제기 가능.
- **권장 수정**: spec.md §2 Non-goals 또는 §8 Exclusions에 "단, 이모지 금지 정책(coding-standards.md)에 따라 국기 이모지 → 텍스트 라벨 전환은 필수 허용"을 명시적 예외로 추가.

#### D-013. AC-G4-07 Lighthouse 기준이 Nextra 베이스라인보다 낮을 수 있음

- **위치**: spec.md §7 성능 제약:L180 ("기존 Nextra 베이스라인보다 느려서는 안 된다"), acceptance.md AC-G4-07:L385~395
- **기대**: Lighthouse Performance ≥ Nextra 베이스라인.
- **실제**: AC-G4-07은 Desktop Performance ≥ 85, Mobile ≥ 75 등 절대값만 요구. Nextra 베이스라인이 95라면 85로 떨어져도 AC 통과.
- **영향**: 제약 위반을 AC가 감지하지 못함.
- **권장 수정**:
  - AC-G4-07에 "추가로 Phase 7 사전 측정한 Nextra 베이스라인보다 Performance 점수가 5 이상 하락해서는 안 된다" 조항 추가.
  - phase-7-preview-verification.md 템플릿에 Nextra 베이스라인 기록 필드 신설.

#### D-014. 48h 모니터링 실패 시 롤백 플랜 누락

- **위치**: spec.md REQ-DS-32~34, acceptance.md AC-MON-01:L445~453
- **기대**: 5xx > 0.1%, P1 인시던트 1건 이상, LCP > 2.5s 등 기준 위반 시 조치 경로.
- **실제**: 전진 경로만 기술. 실패 시나리오에 대한 롤백 절차, 책임자, 복구 SLA, 재시도 게이트 없음.
- **영향**: 프로덕션 장애 발생 시 의사결정 지연 → 사용자 영향 확대.
- **권장 수정**: 
  - 신규 REQ-DS-35 (Unwanted): "If any AC-MON-01 threshold is violated during the 48h monitoring window, then the operator shall immediately revert the Vercel project Git source to `modu-ai/moai-adk-docs` main and create an incident record in `.moai/plans/DOCS-SITE/phase-7-rollback.md`."
  - 대응 AC-MON-03 추가.

#### D-015. Hugo module system vs git submodule 결정 미완

- **위치**: plan.md Phase 2 step 3:L77 ("`docs-site/go.mod` (Hugo module system) 또는 `themes/hextra` submodule")
- **기대**: Phase 2 시작 전 단일 방식 확정.
- **실제**: "또는" 표현으로 두 안 제시. "module system 선호"라고 §8에 언급(L351)이 있으나 요구사항으로 잠금되지 않음.
- **영향**: Phase 2에서 임의 결정 → 하류 단계 영향(CI 빌드, Vercel 설정, 종속성 추적).
- **권장 수정**: spec.md §6.1에 추가 REQ 또는 plan.md Phase 2 고정 결정: "Hextra는 Hugo module system(`go.mod` import)으로 통합하며, git submodule은 사용하지 않는다."

#### D-016. REQ-DS-15 "empty skeleton placeholder folders" 내용 형식 모호

- **위치**: spec.md REQ-DS-15:L126
- **기대**: 빈 폴더인지, `_index.md`만 있는지, "Coming soon" 페이지인지 명확.
- **실제**: "empty skeleton placeholder folders"만 기술. Hugo는 빈 디렉토리를 `.gitkeep` 없이 추적하지 않음. 실제 구현 시 `_index.md`(draft: true or "Coming soon") 등 결정 필요.
- **영향**: Phase 3 구현자가 임의 결정 → locale 간 일관성 훼손 가능.
- **권장 수정**: "placeholder folders containing a single `_index.md` with `draft: true` frontmatter and a redirect to the ko locale" 등으로 구체화.

#### D-017. Hextra/Vercel Edge Function 플랫폼 제약 미검증 (D-007과 중복)

- (D-007 참조)
- 별도 항목으로: Hextra가 Hugo module로 통합될 때 `go.mod` + `hugo --gc` 빌드 시간 영향 — Vercel 빌드 timeout(10분 기본) 초과 여부 미검증.
- **권장 수정**: Phase 2 완료 시 로컬에서 clean build 시간 측정 → 5분 이상이면 대안 검토 지침 추가.

#### D-018. Hextra 실제 제약 (내장 versioning 없음, React 컴포넌트 미지원) 문서화 부재

- **위치**: spec.md §2 Goals/Non-goals (누락), plan.md §8 기술 접근 (누락)
- **기대**: 이해관계자가 Hextra의 본질적 한계를 인식할 수 있도록 명시.
- **실제**: 
  - Hextra는 내장 versioning 기능 없음 → SPEC이 `scripts/docs-version-snapshot.go`로 구현 — 하지만 "Hextra에는 versioning이 내장되어 있지 않으므로" 근거 서술 없음.
  - Hextra는 Hugo 기반 → React 컴포넌트 실행 불가. 향후 React 기반 인터랙티브 요소 추가 요구가 발생해도 수용 불가. 이 점이 Non-goals에 명시되지 않음.
- **영향**: 후속 SPEC에서 "왜 React 컴포넌트 못 씀?" 이슈가 재등장 가능.
- **권장 수정**: spec.md §2 Non-goals 끝 또는 §4 용어집 인근에 "Hextra 한계" 소섹션 추가:
  - "Hextra는 내장 versioning 기능이 없어 `scripts/docs-version-snapshot.go` 로 구현한다."
  - "Hextra는 Hugo 기반이므로 React/JSX 컴포넌트 실행이 불가능하다. 대체로 Hugo partial + 바닐라 JS만 허용된다."

#### D-019. aggregateRating 제거의 SEO 영향 평가 부재

- **위치**: spec.md REQ-DS-21:L142, §8 Exclusions:L192, migration-inventory.md R5
- **기대**: Google Rich Results 인덱싱에 미치는 영향 평가 및 대체 지표(선택) 제시.
- **실제**: "Google Search Central policy 위반 소지"만 언급. Search Console에서 기존 `aggregateRating` 기반 Rich Results가 등록되어 있는지, 제거 시 검색 트래픽 감소 가능성이 있는지 조사 없음.
- **영향**: 제거가 SEO 트래픽 저하를 유발할 수 있으나 감지/롤백 계획 없음.
- **권장 수정**: Phase 4 pre-production 단계에 AC 추가: "Google Search Console에서 Rich Results 항목 중 `aggregateRating` 의존 항목을 조회하고 영향도를 `.moai/plans/DOCS-SITE/phase-4-seo-impact.md`에 기록."

---

### LOW (6건)

#### D-020. plan.md §7 G2 표기 오류 — Phase 4 누락

- **위치**: plan.md §7 G2:L324 ("Gate G2 — Hugo Build Gate (Phase 3 → Phase 5)")
- **실제**: Phase 4(커스텀 컴포넌트 재현)도 G2 범위에 포함되나 표기에서 빠짐.
- **권장 수정**: "(Phase 3/4 → Phase 5)"로 수정.

#### D-021. REQ-DS-34 Phase 7 참조 vs plan.md Phase 8 archive 수행 불일치

- **위치**: spec.md REQ-DS-34:L176 ("When Phase 7 G4 passes and 48 hours..."), plan.md §3 테이블:L35~36
- **실제**: REQ-DS-34는 Phase 7 조건 후 동작으로 기술되나, plan.md는 archive 작업을 Phase 8에 배치.
- **권장 수정**: REQ-DS-34 주어를 "the system(Phase 8 automation)"으로 명시 또는 plan.md에서 "Phase 7 조건 충족 시 Phase 8 자동 진입" 조건 명시.

#### D-022. Phase 0 A-unique config 파일 처리 미명시

- **위치**: spec-project-diff.md §4.3 — moai-docs에만 존재하는 `gate.yaml`, `memo.yaml`, `observability.yaml`.
- **기대**: spec.md 혹은 plan.md에서 해당 3개 파일 이관/폐기 결정 명시.
- **실제**: spec.md Exclusions에 별도 언급 없음. Phase 0 산출물은 "이관 생략" 권장했으나 SPEC에 잠기지 않음.
- **권장 수정**: spec.md §8 Exclusions에 "moai-docs의 `.moai/config/sections/{gate,memo,observability}.yaml` 3개는 docs-site 운영에 무관하므로 이관 대상에서 제외한다" 추가.

#### D-023. REQ-DS-19 배너 구현 메커니즘 기술 부재

- **위치**: spec.md REQ-DS-19:L136, plan.md Phase 5 step 2:L169~172
- **기대**: 배너가 어떻게 조건부로 렌더링되는지(Hugo layout/partial/shortcode/front matter 기반)
- **실제**: "상단 배너 partial 자동 연결" 한 줄.
- **권장 수정**: plan.md Phase 5에 "Hextra `banner` partial 또는 `baseof.html`의 `{{ if (hasPrefix .Path "v") }}` 조건 블록으로 구현. 대체 latest URL은 `.File.Path`에서 `v<version>/` prefix 제거로 산출" 등 구체화.

#### D-024. AC-G3-05 CLAUDE.local.md §17 섹션 검출 정규식 취약

- **위치**: acceptance.md AC-G3-05:L243~249
- **실제**: `grep -q "$section\|### ${section#§}" CLAUDE.local.md` — `${section#§}`는 "§17.1" → "17.1"로 prefix 제거. 그러나 `###` 헤더에 공백 규칙 불일치 시 실패 가능(예: `### 17.1. URL`).
- **권장 수정**: `grep -Eq "^### *(§?17\.[1-6])"`로 정규식 강화.

#### D-025. AC-G4-03 `$PREVIEW_URL` 변수 초기화 누락

- **위치**: acceptance.md AC-G4-03:L334
- **실제**: `PREVIEW_URL="https://<preview-deployment>.vercel.app"` placeholder. 스크립트를 직접 실행할 수 없음.
- **권장 수정**: "Preview URL은 Vercel Dashboard에서 Preview Deployment URL을 복사해 `export PREVIEW_URL=...` 후 스크립트 실행" 사용 지침을 AC 앞에 명시.

---

## 3. 누락 항목 (Gaps)

1. **Vercel 프로젝트 rebinding 무중단 보장 근거**: "무중단"을 spec.md §2 Goals:L80에 선언했으나 Vercel 공식 정책 확인 근거가 문서에 없음. Phase 0 또는 Phase 7 사전 준비 단계에 검증 AC 추가 필요.
2. **git subtree 압축 후 크기 실측**: D1은 "39 MB → 50 KB"로 단언하나 실측치 없음. Phase 2 step 1 완료 후 `du -sh .git/` 비교 기록 AC 추가 권장.
3. **Hugo 빌드 시간 측정 베이스라인**: Phase 2/3 완료 시점에 `time hugo --minify --gc` 기록이 없음. Vercel 빌드 timeout(10분 기본) 위험 평가 부재.
4. **search 재활성화 결정**: migration-inventory.md §4.2에 `search: false`가 현재 상태로 기록. spec.md §9 후속 작업에 "FlexSearch 최적화 여부 검토"만 언급. 본 SPEC에서 기본 활성/비활성 결정이 없음.
5. **llms.txt 이관 처리**: migration-inventory.md §1.5에 `llms.txt` 존재 언급. spec.md는 `static/`에 복사한다고만 가정(REQ-DS-02). llms.txt의 URL 기준이 새 도메인에서 유효한지 검증 AC 없음.
6. **og.png 최적화 검증 AC**: plan.md Phase 4 step 5 "og.png 7.5 MB → 500 KB 이하 압축". 대응 AC 없음. AC-G4-05에 "static/og.png 크기 ≤ 500 KB" 추가 권장.
7. **`_meta.ts` 파싱 스크립트의 TypeScript 타입 검증**: R8 리스크 완화책으로 "JSON Schema 검증 CI" 언급(plan.md §6 R8:L311)되나 대응 AC 없음.

---

## 4. 모호성 (Ambiguities)

1. **"시각적 동등성 (color/layout/logo) 유지"** (plan.md Phase 4 Gate:L150) — 동등성의 정량 기준 없음. 픽셀 diff? 스크린샷 수동 비교?
2. **"Hugo frontmatter `weight` + `title`로 변환"** (spec.md §4 용어집 `_meta.ts`:L67) — `_meta.ts`의 key 순서가 1, 2, 3 ... 인지 10, 20, 30 ... 인지 스텝 값 미지정.
3. **"경미 제보는 허용"** (acceptance.md AC-MON-02:L459) — 경미/치명 기준 주관적. 명확한 분류표 필요.
4. **"Hextra lazy-load 설정"** (REQ-DS-11:L116) — 실제 Hextra config key 이름 불명. 설정 존재 여부 미검증.
5. **"단일 squash 커밋 1개만 남김"** (spec-project-diff.md §7 D1 권장안, plan.md §8:L349) — squash 커밋 메시지 템플릿 미지정.
6. **"릴리스 태그 시점에 스냅샷 호출"** (plan.md §8 버전 스냅샷:L354) — 호출 주체가 manager-git인지 GitHub Actions인지 모호. REQ-DS-17은 "the release automation"으로만 기술.

---

## 5. 강점 (Strengths)

균형 있는 평가를 위해 잘 처리된 부분 3건:

1. **Phase 0 리스크 완전 수용**: migration-inventory.md의 R1~R10 10개 리스크가 plan.md §6 Risk Register:L302~313에 전부 매핑되고, 각 리스크별 담당 Phase가 지정됨. 추적성 우수.
2. **Dead code 명시적 제외**: Exclusions 섹션(spec.md §8:L190)에서 Phase 0이 식별한 dead code 7개 항목을 각각 열거. "이식 대상 제외"를 REQ-DS-03 CI 검증과 연결. manager-spec이 Phase 0 결과를 성실히 반영.
3. **SPEC-I18N-001 흡수·아카이브 프로세스 명문화**: REQ-DS-28:L160이 (a) 유지할 REQ, (b) 재작성할 REQ, (c) archive 경로를 구체적으로 기술. spec-project-diff.md §2의 판단 매트릭스를 단일 REQ로 압축한 점이 명료함.

---

## 6. Chain-of-Verification Pass (M6)

1차 감사 이후 두 번째 pass에서 추가로 발견한 결함:

- **D-002 (CRITICAL)** 은 1차 pass에서 놓쳤다가 REQ-DS-32의 "When G4 is granted" 문구를 Phase 7 step 순서와 재대조하며 발견. Phase 7 시퀀싱과 REQ 트리거 불일치는 명세의 근본 논리 구멍이다.
- **D-003 (HIGH)** 도 2차 pass에서 발견: Phase 2가 "G1 (계속)"으로 표기되어 있어 정상으로 착시하기 쉬우나, §7에 Phase 2 고유 기준이 아예 없다.
- **D-013 (MEDIUM)** 은 Lighthouse 절대값과 Nextra baseline 상대값을 교차 대조하며 발견.
- **D-014 (MEDIUM)** 48h 모니터링 실패 롤백 플랜 누락은 REQ-DS-34가 전진 경로만 기술함을 재확인하며 발견.

1차 pass에서 spot-check만 한 영역을 재점검:
- REQ 번호 순서: 01~34 전수 확인, 누락·중복 없음.
- Exclusions 항목: 10개 이상 확인.
- 4개 locale 관련 REQ: REQ-DS-04/12/13/15 전수 재점검.
- D1~D4 이중 명시: spec.md §5 vs plan.md §2 전수 비교, 일치 확인.

2차 pass에서 새로 발견된 결함은 defects 리스트에 이미 반영됨.

---

## 7. G1 판정 및 요약

### 판정: **FAIL**

- Critical 2건 + High 6건 → FAIL 조건 확정 트리거.
- Plan Auditor 수정 요청 건수: **25건 (Critical 2 + High 6 + Medium 11 + Low 6)**
- plan.md G1 기준의 "수정 요청 3건 이하"를 대폭 초과.

### 반드시 수정해야 할 항목 (Critical + High, 8건)

1. D-001 — REQ 개수 상한 모순 해소 (AC-G1-02 상한 조정 또는 REQ 통합)
2. D-002 — REQ-DS-32 시퀀싱 수정 (재바인딩은 G4 이전, 프로모션은 G4 이후)
3. D-003 — Phase 2 전용 게이트 정의 신설
4. D-004 — DNS TTL 변경 필요성 조사 및 리드타임 재계산
5. D-005 — REQ-DS-09 vs AC-G2-06 tolerance 일치
6. D-006 — REQ-DS-05 vs AC-G2-04 tolerance 일치
7. D-007 — Vercel Edge Function `runtime: 'edge'` 명시 및 플랫폼 제약 수용 AC 추가
8. D-008 — G1 EARS 범위 plan/acceptance 문서 간 통일

### 수정 후 재심사 요청 경로

1. manager-spec이 본 리포트의 Critical 2건 + High 6건을 최우선으로 수정.
2. Medium·Low 결함은 일괄 또는 선별 수정(최소 50% 해결 권장).
3. 수정 완료 후 spec.md 하단 HISTORY에 `v0.2.0` 엔트리 추가 (수정 내역 요약).
4. Iteration 2 재심사 요청 — 이 리포트의 Critical/High는 Iteration 2에서 resolution 증거(라인 인용 포함)로 해결 확인.
5. Iteration 2에서 Critical 0 + High ≤ 2 달성 시 PASS 가능.

---

## 8. 참고 — 사용된 증거 인용 요약

| 결함 ID | 1차 증거 | 2차 증거 |
|---------|----------|----------|
| D-001 | spec.md §6 REQ-DS-01~34 (34개) | acceptance.md L41 awk `$1 <= 30` |
| D-002 | spec.md L172 REQ-DS-32 "When G4 is granted" | plan.md L228~247 Phase 7 단계 순서 |
| D-003 | plan.md §3 L35 "Gate: G1 (계속)" | plan.md §7 L316~345 Phase 2 기준 없음 |
| D-004 | spec.md L172 REQ-DS-32(e) | spec.md L198 "신규 Vercel 프로젝트 생성 금지" |
| D-005 | spec.md L112 REQ-DS-09 "all 569" | acceptance.md L156~158 "564~574" |
| D-006 | spec.md L102 REQ-DS-05 "735 total" | acceptance.md L127 "-ge 700" |
| D-007 | spec.md L122 REQ-DS-13, plan.md L162 | acceptance.md L180~197 runtime 미검증 |
| D-008 | plan.md L320 "15~25" | acceptance.md L41 "15~30" |

---

**감사 종료**. 재심사를 위해 본 리포트에 대응하는 수정 결과 요약을 `spec.md` HISTORY 항목에 포함하여 Iteration 2를 요청하시기 바랍니다.

- 출력 경로: `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/g1-audit-report.md`
