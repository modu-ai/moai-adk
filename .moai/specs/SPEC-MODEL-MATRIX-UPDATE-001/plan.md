---
spec: SPEC-MODEL-MATRIX-UPDATE-001
tier: M
created: 2026-09-30
author: manager-spec
---

# plan.md — SPEC-MODEL-MATRIX-UPDATE-001

## §A Context (plan-phase 검증 결과)

리드 디스패치의 6개 항목 + 유지 항목을 본 워크트리(develop @ 3dd5adf2f 기반)에서 전수 검증했다. 앵커 드리프트와 실제 위치는 아래 표와 같다.

### §A.1 6개 항목 + 유지 항목 검증표

| # | 지시 | 리드 앵커 | 실측 위치(본 트리) | 판정 |
|---|---|---|---|---|
| ① | codex 감사 핀 신설 {gpt-6.1-sol, high} | — | Go 기본값 EMPTY(`defaults.go:1200-1211`), 템플릿 codex 핀 EMPTY(`templates workflow.yaml:114-116`), 해석기 단말 폴백 zero(`mcp_codex.go:206-212` + `model_backend_default_test.go:36-37` 단정), 서빙 접두사 통과(`mcp_codex.go:156`), 로컬 핀 `gpt-5.6-sol`(`.moai/config/sections/workflow.yaml:26`) | **신설 확정** — 3개 기본 표면이 현재 비어 있거나 구형 |
| ② | claude 감사 핀 sonnet/high → {opus-5-5, medium} | defaults.go:1202 | `defaults.go:1200-1211` 일치(드리프트 ~1줄) + `mcp_claude.go:20-21` 상수 2개 + 템플릿 `workflow.yaml:111-113` + 로컬 `workflow.yaml:22-24` + `audit_models.go:79` 주석 + `moai-mcp-tools-catalogue.md:70`(미러 동반) + `i18n.js:586/1481` | **확정** — 총 4+3 표면, 리드 앵커 외 6곳 추가 발견 |
| ③ | `moai glm` effort → max 통일 | defaults.go:246 주석 | flash max-only 주석 `defaults.go:245-248` 일치. 세션 기본값 이미 max(`glm_effort_overlay.go:208-210`). 미통일 표면 = web 티어 기본값(`schema_sections.go:174-183`: high/medium→high, low→low, fable→max) + 그 근거 주석(`:169-172`). llm.yaml에는 tier 기본값 서술이 없음(plan-audit D6 재확인 — collapse map 블록은 변화 불요) | **확정** — 잔여 표면은 web 티어 기본값 |
| ④ | 구형 모델(4.5~4.7) 상수/슬롯 점검·정리 | — | **잔존 확인** — 4표면: (a) 제공 집합 `closed_sets.go:82-84`(5.1/4.7/4.5-air 노출, web 셀렉트 2곳에 파생), (b) 상수 `defaults.go:255-261`, (c) statusline `memory.go:36-42`, (d) legacy struct 멤버 `types.go:361-364`+`defaults.go:971-974`(소비자 `glm.go:746-753`·`:847`) | **운영자 재정(2026-09-30 gate)** — 권장안(철수 전용) 기각, **전면 삭제** 확정(DR-2). 잔존 확인 자체는 유효 — 4표면 모두 삭제 대상으로 편입 |
| ⑤ | moai web 위젯: gpt-6.1-sol 선택 노출 | — | 감사 codex 모델 필드는 **TypeText 자유 입력**(`schema_sections.go:388`) — 닫힌 집합 부재(verified-absence). `audit.glm.model`만 셀렉트(`:391-392`). `codexServableModelPrefixes`는 `gpt-` 접두사로 이미 통과(`mcp_codex.go:156`) | **변경 불필요 확정** — 결정 기록 DR-3 |
| ⑥ | statusline glmContextWindows 1M 유지 검증 | memory.go | `memory.go:34-35` — glm-5.3-flash: 1_000_000(명시 항목), glm-5.3: 1_000_000 | **확인 완료** — 유지, 증거 기록만 |
| 유지 | 템플릿 `"model": "sonnet"` 별칭 | — | `settings.json.tmpl:417` 실존 확인 | **보존 확정** |

### §A.2 설계 결정

1. **핀 도달 경로는 2층이다.** (i) Go 기본값 `NewDefaultWorkflowConfig().Audit` — 전체 Loader 경로와 `moai init` 시딩이 소비. (ii) 감사 해석기의 단말 폴백(`resolveCodexAuditModelEffort` → zero, `resolveClaudeAuditModelEffort` → 하드코드 상수) — 섹션 전용 로더(`audit_pin.go`)가 원본 파일만 읽으므로 Go 기본값을 보지 않는다. 요구사항의 "모든 감사 경로가 새 핀으로 수렴"을 충족하려면 **양층을 모두** 고쳐야 한다. M1이 이 점을 명시한다.
2. **codex 핀의 단말 폴백을 zero에서 새 핀으로 바꾸는 것**은 `model_backend_default_test.go:36-37`의 zero-value 단정을 뒤집는다 — 같은 변경 집합에서 테스트를 뒤집는 것이 정상 경로다(REQ-MMU-007).
3. **claude 핀 값 형태** — 결정 기록 DR-1 확정(claude-opus-5-5).
4. **item ④ 정리 방식** — 운영자 재정으로 전면 삭제 확정(결정 기록 DR-2). glm-5.2 선례(`closed_sets.go:76-81`)는 철수 기록 주석 문체에만 준용한다.
5. **i18n과 룰 카탈로그**는 감사 기본값을 산문으로 서술하는 표면이라 REQ-MMU-007의 일관성 의무에 든다. 룰 카탈로그는 로컬+템플릿 미러가 byte-parity 검사로 묶여 있어 같은 커밋에 함께 갱신한다.

## §B Known Issues

- REQ-AMP-005(중립성)와 본 SPEC의 신설 핀이 정면으로 충돌한다. 코드 주석의 대체 기록 없이 지나가면 다음 감사가 "drift"로 오판한다 — audit_models.go 주석 갱신을 M1 필수 작업으로 둔다.
- 로컬 도그푸드 workflow.yaml(`gpt-5.6-sol`)을 놓치면, 템플릿·Go 기본값을 갱신해도 **운영자 트리만 구형 모델에 머문다**(프로젝트 핀이 Go 기본값보다 우선). M1 필수 작업.
- `gpt-6.1-sol`의 z.ai/codex 실서빙 가능성은 본 레인에서 실측 불가다(외부 API 호출 금지). 접두사 가드 통과(`gpt-`)까지만 본 트리에서 검증 가능하며, 실서빙은 운영자 환경의 첫 감사 실행에서 확인된다 — Residual-risk로 기록.
- DR-2(전면 삭제)의 수용된 위험: 구형 모델을 지명하던 기존 사용자 설정은 **값**은 폴백+1행 경고로, 별칭 **키**(`opus:`/`sonnet:`/`haiku:`)는 비엄격 로더 의미대로 무음 소멸한다 — 보존이 아니라 관리되는 전환이 운영자의 선택이다(2026-09-30 gate).
- DR-1의 잔여 관찰 항목: claude CLI가 `claude-opus-5-5`를 첫 감사 스폰에서 수용하는지 — 트리 내 증명 불가(외부 실행), §E E8 관찰 항목으로 run-phase가 확인한다.

## §C Pre-flight (run-phase 시작 전 확인)

1. `go build ./...` — 베이스라인 녹색 확인(변경 전).
2. `grep -n "gpt-5.6-sol" -r internal/ .moai/config/` — 구형 codex 핀의 전수 위치 스냅숏.
3. `grep -n '"sonnet"' internal/cli/mcp_claude.go internal/config/defaults.go` — claude 핀 위치 스냅숏.
4. `sed -n '33,43p' internal/statusline/memory.go` — REQ-MMU-005 사전 증거(변경 후 재측정으로 불변 입증).
5. `sed -n '415,419p' internal/template/templates/.claude/settings.json.tmpl` — REQ-MMU-006 사전 증거.

## §D Constraints

- [HARD] 검증 예산: `go build ./...` + grep/sed 허용. **`go test ./...` 금지**(레인 규율 — CI가 전체 판정). 영향 패키지의 테스트는 CI 몫이며, run-phase는 빌드+정적 검증으로 마친다.
- [HARD] Template-First: 템플릿 표면 변경은 `internal/template/templates/` 먼저, `make build`로 임베드 반영. 로컬 미러 동반 갱신.
- [HARD] 룰 카탈로그(`moai-mcp-tools-catalogue.md`)는 로컬 + 템플릿 미러를 **같은 커밋**에(`rule_template_mirror_test.go` byte-parity).
- [HARD] `.claude/agents/**` 무변경 — 감사 핀과 agent 모델 정책은 별개 축(Out of Scope).
- 병합·push는 직렬 통합 창 경유 — 본 SPEC 착지 후 리드 주도(레인 소관 밖).

## §E Self-Verification

run-phase 종료 시 아래를 관측하고 증거를 `.moai/reports/t1368/`에 남긴다:

- E1 `go build ./...` exit 0 (전사 출력).
- E2 새 핀 존재 grep: `grep -rn "gpt-6.1-sol" internal/ .moai/config/ internal/template/templates/` — REQ-MMU-001 표면 전체 적중.
- E3 claude 핀 grep: `grep -rn "claude-opus-5-5" internal/ .moai/config/ internal/template/templates/` — REQ-MMU-002 표면 적중 + `grep -n "sonnet" internal/cli/mcp_claude.go`가 0행(상수 제거 확인).
- E4 glm 티어 기본값 grep: `grep -n -A 10 "func glmDefaultTierEffort" internal/settings/schema_sections.go` — 전 티어 GLMStateMax 확인.
- E5 제공 집합 grep: `grep -n -A 3 "func ValidGLMModels" internal/config/closed_sets.go` — 두 모델만 확인.
- E6 불변 재측정: memory.go:34-35 (1M 양쪽) + settings.json.tmpl:417 (`"model": "sonnet"`) — 변경 전후 동일 바이트.
- E7 잔존 스캔: `grep -rn "gpt-5.6-sol" internal/ .moai/config/` — 0행 (테스트 픽스처 갱신 포함).
- E8 삭제 스캔 + 관찰 항목: `grep -rn "DefaultGLM45\|DefaultGLM46\|DefaultGLM47\|DefaultGLM45Air\|DefaultGLM51\|DefaultGLM52\|DefaultGLM5Turbo" internal/` — 0행. `grep -rn "Models.Opus\|Models.Sonnet\|Models.Haiku" internal/` — 0행(N2 — cli/ 한정을 internal/로 확대; defaults_test.go:362-379가 cli/ 밖 적중). 관찰(트리 외, run-phase): claude CLI가 `claude-opus-5-5` 핀으로 첫 감사 스폰이 성립하는지 확인(DR-1 잔여 항목 — 증거는 완료 보고에 서술로 기록).

## §F Milestones (결정 가역성 순 — 바뀔 가능성이 큰 결정이 앞선다)

### M1 (Priority High) — 감사 핀 2종: codex 신설 + claude 승격

가장 바뀔 가능성이 큰 결정(모델 값)을 먼저 착지해, 후속 마일스톤이 안정된 값 위에 놓이게 한다.

1. `internal/config/defaults.go` — `Audit.Codex: {gpt-6.1-sol, high}` 신설, `Audit.Claude: {claude-opus-5-5, medium}` 변경.
2. `internal/cli/mcp_codex.go` — `resolveCodexAuditModelEffort` 단말 폴백을 zero → `{gpt-6.1-sol, high}`(상수 추출, mcp_claude.go의 claudeAuditDefault* 패턴 준용).
3. `internal/cli/mcp_claude.go:20-21` — 상수 2개 갱신.
4. `internal/config/audit_models.go` — REQ-AMP-005 대체 기록 + :79 "documented sonnet/high default" 서술 갱신.
5. `internal/template/templates/.moai/config/sections/workflow.yaml` — claude/codex 핀 갱신 + 상단 주석(:15-20, :98-109) 갱신.
6. `.moai/config/sections/workflow.yaml` — 로컬 도그푸드 동기 갱신(claude/codex 핀 + :20 주석 "Claude defaults to sonnet/high" — plan-audit D5b).
7. 테스트 갱신: `audit_models_test.go`, `mcp_audit_config_test.go`, `audit_pin_test.go`, `model_backend_default_test.go`(zero 단정 뒤집기), `mcp_codex_audit_pin_test.go`, `factory_operational_fixture_test.go`, `internal/settings/audit_pin_fields_test.go`.

### M2 (Priority High) — glm effort max 통일

1. `internal/settings/schema_sections.go` — `glmDefaultTierEffort` 전 티어 `GLMStateMax` + 근거 주석 갱신(:164-183).
2. `internal/template/templates/.moai/config/sections/llm.yaml:55-95` — collapse map 블록의 유효성 확인(verify-only — tier 기본값 서술이 없어 갱신 불요, plan-audit D6 재분류).
3. `internal/settings/schema_sections_test.go` — 티어 기본값 단정 갱신(파일 내 glm effort 기본값 관련 케이스).

### M3 (Priority Medium) — 구형 모델 전면 삭제 + 레거시 폴백 (DR-2)

1. `internal/config/closed_sets.go:82-84` — `ValidGLMModels()`를 {DefaultGLM53Flash, DefaultGLM53}로 축소 + glm-5.2 선례 문체의 철수·삭제 기록 주석.
2. `internal/config/defaults.go` — `:255-261` 상수 7종(DefaultGLM45/46/47/45Air/51/52/5Turbo) 삭제 + `:249-254` 보존 근거 주석을 삭제 기록으로 대체.
3. `internal/statusline/memory.go` — `:36-42` 구형 항목 7종 삭제 + `:96` "glm-4.5 masking glm-4.5-air" 주석 갱신(`:34-35` 1M 불변은 REQ-MMU-005로 유지).
4. legacy struct 별칭 삭제: `internal/config/types.go:361-364` 필드(Opus/Sonnet/Haiku) + `defaults.go:971-974` 배정 + `internal/cli/glm.go:746-753`·`:847` 레거시 별칭 폴백 사슬 제거 + `internal/settings/schema_sections.go:200-202` REQ-WC12-006 "무접촉 보존" 주석 갱신.
5. 레거시 폴백 구현: tier 슬롯 값이 잔존 알려진 모델 집합 밖이면 티어 기본 모델로 폴백 + 1행 경고(fail-open; 구현 위치는 런처/로더 중 resolved-value 소비 지점 — run 판단 사항).
6. 테스트 전면 갱신(17파일): `glm_tier_test.go`(D3), `defaults_test.go`(D3 + `:421` 주석 + `:362-379` Models.* 접근 — N2), `schema_select_preserve_test.go`(13적중), `glm_persist_gate_test.go`, `glm_effort_overlay_test.go`, `glm_team_test.go`(`:140-141` Models.Opus 접근 — 필드 삭제 시 컴파일 단절, plan-audit iter-2 N1), 그리고 **단정 재작성 3파일** — `glm_autocompact_test.go`(`:16-68`, glm-5.2→1000000 리터럴 단정 + [1m] 접미사 케이스, 주석이 "내장 glmContextWindows 테이블이 검증 대상 baseline"임을 선언), `glm_max_context_test.go`(`:28-30`, 5.2→1M/5.1→200K/4.7→128K), `cg_mode_hardening_test.go`(`:223-227`, glm-5.2 1M 해석 전제의 패리티 테스트) — 이 셋은 열거만으론 부족하고, 윈도우 해석 리터럴이 삭제로 사라지므로 **단정 자체를 삭제 후 동작 기준으로 재작성**해야 한다(plan-audit iter-2 N1).
7. `internal/template/templates/.moai/config/sections/llm.yaml:97-111` — context_windows 주석 블록에 오버라이드 경로 유지 명시 + 내장 테이블이 더 이상 구형 id를 담지 않음을 밝히는 주석 갱신(plan-audit iter-2 N3).
8. 로컬 도그푸드: 삭제 대상 값 0행(DR-2 실측, 2026-09-30) — 갱신 불요, 커밋 직전 재확인만.

### M4 (Priority Medium) — 문서·i18n 일관성

1. `.claude/rules/moai/core/moai-mcp-tools-catalogue.md:70` + 템플릿 미러 — 같은 커밋, sonnet/high → 새 claude 핀 서술.
2. `internal/web/assets/i18n.js` — `f.workflow.audit.claude.model.desc` 4 로케일(en :586, ko :1481, ja :2248, zh :3015)과 `f.workflow.audit.claude.effort.desc` 4 로케일(en :588, ko :1483, ja :2250, zh :3017) — model/effort 기본값 문구 전부 갱신(plan-audit D4 + 감사 발견 누락분인 동급 effort 설명 표면).
3. `make build` (agents-emit-check 통과 — `.md` agent 변경 없음을 확인) + 임베드 재생성.

### M5 (Priority High, 종결 게이트) — 검증·증거

1. §E E1-E7 전 항목 실행, 증거를 `.moai/reports/t1368/`에 기록.
2. REQ-MMU-005/006 불변 재측정 기록.
3. 완료 보고: 카드 id · 브랜치/HEAD · 변경 파일 목록 · 재측정 범위 · 미푸시 커밋 수.

## §G Anti-Patterns

- **Go 기본값만 고치고 해석기 폴백을 놓치는 것** — 섹션 전용 로더는 Go 기본값을 안 읽는다(§A.2-1). 두 층 중 하나만 고치면 일부 경로가 구형/공백으로 남는다.
- **로컬 도그푸드 workflow.yaml 누락** — 프로젝트 핀이 Go 기본값을 이긴다. 운영자 트리만 구형 모델에 머문다(§B).
- **테스트 픽스처 `gpt-5.6-sol` 방치** — E7이 0행을 요구한다.
- **룰 카탈로그 편방향 갱신** — 미러 byte-parity 검사가 build를 중단시킨다.
- **삭제 후 침묵 폴백** — DR-2의 폴백+1행 경고 없이 알 수 없는 모델 값을 침묵 통과시키는 것(운영자가 수용한 위험의 관리 장치가 무력화된다). 반대로 상수·statusline 항목·legacy 필드를 남겨두는 것도 DR-2 위반이다.
- **`codexServableModelPrefixes` 무분별 확장** — `gpt-`가 이미 커버한다. 집합 변경은 근거 없는 확장이다.

## §H Cross-References

- `.moai/specs/SPEC-MODEL-MATRIX-UPDATE-001/spec.md` — 요구사항 본문
- `.moai/specs/SPEC-MODEL-MATRIX-UPDATE-001/acceptance.md` — AC 매트릭스
- 선행 계열: SPEC-V3R6-AUDIT-MODEL-PIN-001(핀 우선순위·어휘), SPEC-GLM-EFFORT-MAX-001(세션 기본 max), SPEC-MODEL-MATRIX-CORE/CONFIG/SURFACES/DOCS-001(v3.1.0, 전종 superseded — 계열 계보)
- 증거 경로: `.moai/reports/t1368/`

## 결정 기록 (Decision Record — operator gate 2026-09-30)

plan-audit iter-1(D1)에 따라 [NC] 마커 절을 결정 기록으로 전환했다. 세 건 모두 운영자 게이트(2026-09-30)에서 확정됐으며, 본 SPEC 어디에도 live 판단 대기 마커는 남지 않는다.

### DR-1 — claude 감사 핀 값: `claude-opus-5-5` 전체 id 확정 (REQ-MMU-002)

- **결정**: 전체 id `claude-opus-5-5` 채택(별칭 `opus` 기각). — 결정일 2026-09-30, 출처: operator gate.
- **D7 접기**: model-policy.md:34의 전체-id 금지는 **agent frontmatter** 스코프다(`[1m]` 컨텍스트 자격 상속 버그 — 하위 에이전트 스폰 실패). 감사 핀은 `-p` 서브프로세스 설정으로 그 결함 경로 밖이며 정책 위반이 아니다. 핀에 `[1m]` 접미사를 붙이지 않는 것은 의도다 — 목적은 감사 등급의 추론 깊이지 컨텍스트 폭이 아니다.
- **잔여 관찰 항목(run-phase)**: claude CLI가 `claude-opus-5-5`를 첫 감사 스폰에서 수용하는지 — 트리 내 증명 불가(외부 실행) 항목으로, M5 관찰 항목(§E E8)에 기록했다.

### DR-2 — item ④: **전면 삭제** (운영자 재정 — 권장안 기각) (REQ-MMU-004)

- **결정**: glm-5.2 선례식 "제공 집합 철수 전용" 권장안을 운영자가 기각하고 **전면 삭제**를 재정했다. — 결정일 2026-09-30, 출처: operator gate. 삭제 범위: `ValidGLMModels` 제공 축소 + `defaults.go:255-261` 상수 7종 + `memory.go:36-42` statusline 항목 + legacy struct 필드(`types.go:361-364`, `defaults.go:971-974`) + 이들을 참조하는 전체 테스트(본 SPEC 열거분 + plan-audit D3의 보장-RED 2파일).
- **수용된 위험의 관리(필수 동반)**: 삭제된 모델을 지명하는 기존 llm.yaml의 동작을 명시한다 — **tier 슬롯 값**이 삭제된 모델 id면 티어 기본 모델로 폴백하고 1행 경고를 낸다(fail-open — 침묵 통과도 하드 오류도 아님, gate-off 경로와 같은 모양). 레거시 별칭 **키**(`opus:`/`sonnet:`/`haiku:`)는 비엄격 로더의 기존 의미대로 무오류 무시된다(구조 필드 소멸의 직접 귀결 — 별도 경고 없음, 이 무음이 수용된 위험의 나머지 절반이다).
- **로컬 도그푸드 점검(본 트리, 2026-09-30 실측)**: `grep -rn "glm-4\.|glm-5\.1|glm-5\.2|glm-5-turbo" .moai/config/` → **0행** — 삭제 대상 값을 지명하는 로컬 설정이 없어 같은 변경에서의 갱신 대상도 없다(verified-absence).
- 폴백의 검증: AC-MMU-012.

### DR-3 — item ⑤ web 위젯: 변경 없음 확정 (verified-absence 유지)

- **결정**: web 위젯 변경 없음 — 감사 codex 모델 필드는 TypeText 자유 입력으로 닫힌 집합이 없고(`schema_sections.go:388`), `codexServableModelPrefixes`가 `gpt-6.1-sol`을 이미 수용한다(`mcp_codex.go:156`). — 결정일 2026-09-30, 출처: operator gate.
