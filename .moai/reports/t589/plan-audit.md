# Plan-Audit — SPEC-UPDATE-ADD-CODEX-001 (card t589)

- Auditor: plan-auditor (independent)
- Date: 2026-09-09
- Measured tree: `.claude/worktrees/t589` @ `5caddeb2d`, branch `WT-add-codex-verb`
- Iteration: 1/3 (Tier M ceiling 2 — per `harness.yaml plan_audit_tier_ceilings`)

## Verdict

**PASS** — 최종 판정 (Iteration 2 델타 재감사 후; 아래 § Re-Audit 참조). Iteration 1 판정은 PASS-with-fixes(0.86)였고 차단 결함 2건의 수리를 확인해 PASS로 승격됐다.

- Must-pass 7항목 전부 PASS (실측 증거 포함, 아래 표).
- Aggregate score **0.86** (harmonic mean) ≥ Tier M PASS threshold 0.80.
- 단, blocking 분류 결함 2건(F1, F2)이 있다 — run-phase 진입 전 manager-spec 의 소규모 SPEC 수정(각 1~2줄 규모)으로 해소해야 한다. F1은 REQ-UAC-005가 기계적 판정 없이 통과 가능한 유일한 경로이고, F2는 카드 범위의 하위 절 하나가 REQ/AC 어디에도 대응되지 않는다.

## Must-Pass Results

| 항목 | 판정 | 증거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | REQ-UAC-001~014 연속·중복·간극 없음 (spec.md:95-114). 14건 ≤ Tier M 상한 16 |
| MP-2 GEARS 형식 | **PASS** | 14건 전부 GEARS 5패턴 또는 레거시 호환형 적합. REQ-001 복합 Where+When(1.0 동급), REQ-003/012 정준 "shall not", REQ-002 레거시 Unwanted 조건형(이관 기간 내). 판정 대상은 요구 계층(REQ-XXX)만 — AC의 Given-When-Then은 검증 계층 정형이므로 감점 아님(M3 § Scope) |
| MP-3 YAML frontmatter | **PASS** | 12 정규 필드 전부 존재·타입 적합 (spec.md:1-17). snake_case 별칭 없음. `phase: "v3.2.0 target"`은 릴리스 라벨(생명주기 단계명 아님). `tier: M`·`era: V3R6` 선택 필드 유효 |
| MP-4 언어 중립성 | **N/A** | 단일 언어(Go CLI+템플릿 본문) SPEC — 다중 프로그래밍-언어 도구 열거 대상 아님. 자동 통과 |
| MP-5 D7 교차-SPEC 조정 | **PASS** | 참조 4종(SPEC-CODEX-WIRING-001·SPEC-UPDATE-MIRROR-HEAL-001·SPEC-UPDATE-VERSION-FLAG-001·SPEC-INIT-HARNESS-PROMPT-001) 모두 존재, `status: completed` — retired/superseded/archived 없음, 조정 절 불요. 실측: 각 spec.md의 `status:` 행 판독 |
| MP-6 D8 크로스플랫폼 | **PASS** | spec.md 내 `syscall` 0회 → D8-4 자동 통과 |
| MP-7 clarification gate | **PASS** | plan.md 내 `[NEEDS CLARIFICATION]` 0건. research.md 부재(Tier M 3-artifact) → 해당 없음 |

## 실측 검증 요약 (이 감사가 직접 재측정한 값)

| SPEC 주장 | 재측정 결과 |
|---|---|
| 템플릿 AGENTS.md 15,415 B (EV-5) | ✅ 15,415 B 일치 |
| RED `update --add-codex` → exit 1 "Unknown flag" (EV-1) | ✅ exit=1, 동일 메시지 재현 |
| `update --help`에 `--add-codex` 부재 | ✅ grep -c = 0 |
| EV-2/3/4/8 (add-codex 0·마커 0·Quality/TRUST 0) | ✅ 전부 0으로 재현 |
| EV-6 (템플릿 AGENTS.md `## 1~7` 7개, 행번호 포함) | ✅ 35/59/103/132/158/204/233 바이트 동일 재현 |
| EV-7 (템플릿 CLAUDE.md `^## [0-9]` = 18) | ✅ 18 |
| `TestCodexContractByteCeiling` green | ✅ `ok github.com/modu-ai/moai-adk/internal/config 0.468s` |
| `CodexContractByteCeiling = 24576` (token_budget_guard.go:95) | ✅ 정확 일치. `contractDocuments`가 루트+템플릿 미러 쌍방 측정 확인 |
| `Wire` 서명 (wire.go:43) | ✅ `func Wire(projectRoot string, out, warn io.Writer) (Result, error)` |
| `refreshCodexWiringBestEffort` (update.go:507) | ✅ 정확 — 동 위치에 repairSkillMirrorBestEffort(:519) 형제 선례도 확인 |
| init 좌표 (:83 force, :132 agent, :191 wireCodexUnlessClaude, :335 validateInitFlags, :843 already-initialized) | ✅ 전부 정확. `--non-interactive` 플래그 존재(init.go:82) — AC-UAC-013 명령 실행 가능 |
| init_test.go:52 --force 플래그 목록 | ✅ 정확 |
| merge 보호 목록 (strategies.go:484-485) | ✅ `## MOAI:LEARNED-WORKFLOW`(+`-LOCAL`) 확인 — D3 전립선 성립 |
| curator TierSurfaceMap Tier 4 → CLAUDE.md (dispatch.go:44) | ✅ 확인 — AC-UAC-012 간접 grep(`Path: "CLAUDE.md"`) 도달 가능 |
| update 플래그 면 13종, `--agent`·`--add-codex` 부재 (M4) | ✅ 13종 정확 일치 (update.go:76-97) |
| `wiringFilesExist` 판별 파일 = hooks.json/config.toml (B4) | ✅ codexwiring.go:96-103 |
| 루트 계약 쌍 ↔ 템플릿 쌍 바이트 동일 (plan §H1) | ✅ AGENTS.md 15,415=15,415 / CLAUDE.md 19,766=19,766 |
| 16개 `.agents/skills` 발행 세트 (M6/EV-9) | ✅ 16개 moai-* 디렉터리 |
| `.agents` `@AGENTS.md` 임포트 (템플릿 CLAUDE.md) | ✅ templates/CLAUDE.md:9 존재 |
| 카드 스테일 좌표 정정 (§1.3) | ✅ `internal/cli/validator.go` 부재 확인. 스테일 좌표의 출처(감사 보고서 97행)도 확인 — 정정 경위가 충실함 |
| 바이트 예산 산술 (§D6) | ✅ 15,415+6,500=21,915 ≤ 22,000 목표, 섹션 예산 합 6.0+0.5=6.5 KB 일관 |

## Findings

| ID | 렌즈 | 심각도 | 분류 | 위치 | 설명 | 필수 수리 |
|---|---|---|---|---|---|---|
| F1 | 3 모순 + 2 검증가능성 | **major** | **blocking** | spec.md:99 (REQ-UAC-005) · plan.md:57,93 (§D1, M1-2단계) · acceptance.md:150 (§D.4 #1) · internal/cli/update_codex_wiring.go:16-20 | REQ-UAC-005는 화이트리스트 위반 시 fail-loud를 요구한다. 그러나 plan이 구현 모델로 지목한 형제 래퍼 `refreshCodexWiringBestEffortAt`은 **검증 거절(ErrValidationRefused)조차 삼키고 경고만 출력**하는 것이 문서화된 설계다. 그 형제를 그대로 본뜬 mutant(오류 삼킴)는 AC 14건 전부를 통과하면서 REQ-UAC-005를 위반한다 — mutant probe 통과 실패. acceptance §D.4 #1의 판정 수단이 "코드 검토"라 기계적 판정이 없다 | plan §D1에 분기 조항 1줄 추가 — "--add-codex 경로는 ErrValidationRefused를 하드 오류로 상향 전파한다(runUpdate 반환 → exit ≠ 0). IO 계열 실패만 best-effort를 유지한다(codexwiring §F 자세)". acceptance §D.4 #1을 "코드 검토"에서 구체 테스트 명칭(M1 4단계에 추가할 래퍼 전파 단위 테스트)으로 승격 |
| F2 | 1 커버리지 | **minor** | **blocking** | 카드 범위 항목 (2) · acceptance.md:96-101 (AC-UAC-011) | 카드의 "CLAUDE.md keeps `@AGENTS.md` import" 하위 절이 REQ/AC 어디에도 대응되지 않는다. 템플릿 CLAUDE.md:9에 오늘 존재하며, AC-UAC-011의 제목 수(18)만으로는 이 줄의 소실을 잡지 못한다 — 임포트가 빠지면 CLAUDE.md이 AGENTS.md을 아예 로드하지 않는 조용한 파괴다 | AC-UAC-011에 회귀-방어 줄 추가: `grep -c '@AGENTS.md' internal/template/templates/CLAUDE.md` ≥ 1 |
| F3 | 2 검증 환경 | minor | optional | spec.md:29 · plan.md:146 | 인용 증거 `.moai/reports/init-tui-audit-20260909.md`가 비추적 파일이고 primary 체크아웃에만 존재한다 — run-phase가 실행될 이 워크트리에서는 경로가 해결되지 않는다 (원문 확인은 primary에서 수행했고 §7 해당 주장이 실재함을 검증했다) | 워크트리 증거 경로로 사본을 두거나 plan §I에 primary 체크아웃 경로임을 명기 |
| F4 | 5 좌표 정밀도 | minor | optional | spec.md:50 (wire.go:57) · spec.md:52 (161행) | 미세 좌표 오차: RefreshWiring 함수는 wire.go:51-56(:57은 내부 return 행), init_agent_flag_test.go 실측 163행(표기 161). 하중을 지는 좌표는 아니다 | F1/F2 수정 시 함께 정정 |
| F5 | 4 범위 경계 | minor | optional | plan.md:6 (§A) vs plan.md:110 (M3-1단계) | §A가 "update 사이드만 소유"라고 선언하지만 M3는 `internal/cli/init.go`를 직접 고친다(안내+고정 테스트). init.go는 형제 t585의 1차 표면이기도 하다 — 두 카드의 워크트리가 같은 파일을 만지는 병합 충돌 위험이 계획에 기록되어 있지 않다 | plan §A에 조율 줄 1줄 추가(M3의 init.go 변경은 안내 1개+테스트 갱신으로 최소화, t585 착지 순서와 충돌 시 해결 주체 명시) |
| F6 | 3 서술 정확도 | minor | optional | plan.md:57 (§D1) | "같은 조건부 자리" 표현 — 실제 update.go:499-507 슬롯은 무조건 호출 밴드이고 게이트는 RefreshWiring 내부(wire.go:52)에 산다. "조건부"대로 읽으면 무조건 호출이어야 할 신설 동사를 조건 아래 두는 오배치 여지 | "조기 반환 이전의 무조건 밴드, update.go:507 바로 옆"으로 서술 정정 |

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 1.00 | 1.0 | REQ 14건 전부 단일 해석 (spec.md:95-114); 결정 D1-D6 사전 고정 (spec.md:57-66, plan.md:55-72) |
| Completeness | 0.75 | 0.75 | 전 섹션+frontmatter 완비, Out of Scope H3 6종 (spec.md:125-147); 카드 하위 절 1개 미대응 (F2) |
| Testability | 1.00 | 1.0 | AC 14건 전부 실행 명령형 (acceptance.md §D.1); RED 장부 4요소 완비 (§D.5, 전 항목 재측정 성공) |
| Traceability | 0.75 | 0.75 | REQ↔AC 전수 매트릭스 (§D.3); 단 1건 간접 매핑 (REQ-UAC-005 → AC-UAC-002 + 코드 검토, §D.4 #1 — F1) |

**Aggregate: 0.86** (harmonic mean) — Tier M threshold 0.80 이상.

## Gaps (이 감사가 관측하지 않은 것)

1. **GREEN 경로 미관측** — 구현이 착지 전(RED 국면)이므로 AC의 통과 방향 명령은 실행하지 않았다. GREEN 판정은 run-phase 소관이다.
2. **검증 거절의 외부 유발 가능성 미확인** — `codexadapter.ValidateConfig` 거절이 사용자 도달 입력으로 강제 가능한지 확인하지 않았다. 불가능하다면(내부 레지스트리 불일치 시에만 발화) CLI 레벨 블랙박스 AC는 원리적으로 불가하므로, F1의 수리는 단위 테스트/명시 조항 쪽이 옳다.
3. **고정 테스트 6종의 실행 미실시** — `go test ./internal/cli/...` 베이스라인을 돌리지 않았다(바이트 가드 테스트만 실행). plan §C #6이 run 진입 시 재측정한다.
4. **t585 원문 미판독** — t585 경계는 이 카드의 §6 배제 조항과 카드 범위 서술로만 판정했다. t585의 SPEC 산출물은 읽지 않았다.
5. **인용 감사 보고서의 §7 원문 정밀 대조 부분적** — primary 체크아웃에서 CLAUDE.md 덮어쓰기 주장의 실재(97행)는 확인했으나, 보고서 전체를 줄 단위로 대조하지는 않았다.

## Recommendation

1. **F1 수리 (run-phase 진입 전 필수)** — plan §D1에 오류-전파 분기 조항 추가 + acceptance §D.4 #1을 구체 테스트 명칭으로 승격. manager-spec의 1~2줄 규모 수정이다.
2. **F2 수리 (동행)** — AC-UAC-011에 `@AGENTS.md` 임포트 회귀-방어 grep 1줄 추가.
3. F3~F6는 선택 수리 — F1/F2와 같은 편집 기회에 함께 처리하는 것을 권한다.
4. 수리 후 재감사는 이 결함 목록의 델타 범위(F1, F2 해소 확인)로 스코프한다 — 전수 재감사 불요 (Retry Loop Contract).

---

# Re-Audit (Iteration 2) — v0.2.0 델타

- Date: 2026-09-09 · Scope: F1/F2 수리 확인 + F5 판정 (델타 스코프, Retry Loop Contract)
- 코드 트리 불변 확인: HEAD `5caddeb2d` 유지, SPEC 디렉터리만 변경(비추적) — Iteration 1의 EV 장부 귀속 전부 유효
- 버전: 0.1.0 → 0.2.0, HISTORY 행 기록 양호 (spec.md:26 — 수리 근거 + 감사 보고서 인용 포함)

## Regression Check — Iteration 1 결함 대비

| ID | 분류 | 판정 | 검증 증거 |
|---|---|---|---|
| F1 (major, blocking) | — | **RESOLVED** | 세 표면 확인: (1) plan.md:57 §D1에 분기 구속 조항 — "`ErrValidationRefused`는 하드 오류로 전파 — exit ≠ 0, best-effort 허용은 IO뿐" + 삼키는 형제 모델(`update_codex_wiring.go:16-20`)의 **명시적 기각**. (2) acceptance.md:150-152 §D.4 #1이 "코드 검토"에서 **기계 판정 2층**으로 승격 — 기존 `TestWireValidationRefusalWritesNothing` green 유지 + 신규 `TestUpdateAddCodex_ValidationRefusalFailsLoud` 존재 필수(부재 = AC-UAC-002/REQ-UAC-005 FAIL). (3) plan.md:128 §G에 삼킴-mutant 행 등재, 신규 테스트가 판정자. 좌표 실측: `func TestWireValidationRefusalWritesNothing` = wire_test.go:164 정확, `go test ./internal/codexwiring -run TestWireValidationRefusalWritesNothing` → `ok 0.586s`. 삼킴-mutant는 이제 exit-0 전파 요구에 적발되고, 테스트 부재 자체가 FAIL이므로 우회 불가 |
| F2 (minor, blocking) | — | **RESOLVED** | 세 표면 확인: acceptance.md:19 매트릭스 행("18 제목 + `@AGENTS.md` import 행 불변 + 품질 게이트 포인터"), :100 Then 절(import 행 유지 명시 + 헤딩 카운트만으론 못 잡는 이유 기록), :101 판정(`grep -c '@AGENTS.md' internal/template/templates/CLAUDE.md` ≥ 1). 사전 상태는 Iteration 1 실측(templates/CLAUDE.md:9)으로 이미 확정 |
| F5 (minor, optional) | — | **RESOLVED (충분)** | plan.md:6 §A 조율 문장 — M3가 init.go의 유일 축·범위는 안내+테스트 갱신뿐·t585 표면과 비겹침·좌표 이동 시 develop 병합 순서 흡수. 공유 표면·범위 한계·흡수 경로를 다루므로 충분 |
| F3 (optional) | 유지 | UNRESOLVED-optional | 미수리 — 인용 감사 보고서가 primary 체크아웃 전용 비추적 파일. 수용된 부채 (M6: optional 목록은 FAIL 근거 아님) |
| F4 (optional) | 유지 | UNRESOLVED-optional | 미수리 — RefreshWiring :51-56, 163행 vs 표기 161. 하중 좌표 아님. 수용된 부채 |
| F6 (optional) | 유지 | UNRESOLVED-optional | 미수리 — plan §D1 "조건부 자리" 표기 잔존(1건). 단 D1의 신규 하드 전파 조항 + ungated 직접 호출 구속이 배치 모호성의 실위험을 흡수. 수용된 부채 |
| 신규 잔여 (cosmetic) | — | 기록만 | acceptance.md:137 §D.3의 REQ-UAC-005 행이 여전히 "간접, §D.4"로 표기 — §D.4 #1이 이제 직접 기계 판정(신규 테스트 부재=FAIL)을 운반하므로 서술이 과소 기술. 행이 §D.4를 올바로 가리키므로 모순 아님 — 문구 잔여 |

신규 결함 없음. 델타는 REQ/AC 수(14/14)를 변동시키지 않았고, 새 요구는 기존 AC-UAC-011의 확장과 §D.4 판정면 승격으로 흡수됐다 — Tier M 예산 영향 없음. 삼킴-mutant의 재평가: `TestUpdateAddCodex_ValidationRefusalFailsLoud`가 exit ≠ 0 전파를 요구하고 테스트 부재 자체가 FAIL이므로, Iteration 1의 유일한 통과 경로였던 삼킴-mutant는 이제 기계적으로 적발된다.

## Re-Audit Scores

| Dimension | Score | 변동 근거 |
|-----------|-------|----------|
| Clarity | 1.00 | 불변 (F6 서술 잔여는 plan 산문 — REQ 모호성 아님) |
| Completeness | 1.00 | 0.75 → 1.00 (F2 폐쇄 — 카드 하위 절 전부 대응) |
| Testability | 1.00 | 불변 |
| Traceability | 1.00 | 0.75 → 1.00 (F1 폐쇄 — REQ-UAC-005에 직접 기계 판정면 확보; §D.3 문구 잔여는 cosmetic) |

**Aggregate: 1.00** (harmonic mean) ≥ Tier M threshold 0.80.

## Final Verdict

**PASS** — Iteration 2/3 (Tier M ceiling 2). 차단 결함 0건. 수용된 부채: F3·F4·F6 + §D.3:137 문구 잔여 (전부 optional, run-phase 블로커 아님). Skip-eligible 요건(PASS + 0.80 이상) 충족 — artifact-hash 대상은 이 최종 상태(v0.2.0).

### Re-audit Gaps

1. **신규 테스트의 거절-유발 가능성 미검증 (Gap #2 계승)** — `TestUpdateAddCodex_ValidationRefusalFailsLoud`가 거절을 강제할 수단(codexwiring의 시임)이 존재하는지는 확인하지 않았다. 불가하면 run-phase에서 시임 추가가 필요하다 — SPEC 요구(전파·exit ≠ 0)는 유효하며 이는 구현 리스크가지 SPEC 결함이 아니다.
2. `@AGENTS.md ≥ 1`의 사전 상태는 EV 장부 항목이 아니나, Iteration 1 실측(templates/CLAUDE.md:9)으로 확정돼 있다 — 장부 항목 추가는 권장 수준이다.
3. GREEN 경로 미관측은 동일 (구현 착지 전 — run-phase 소관).

