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
| ④ | 구형 모델(4.5~4.7) 상수/슬롯 점검·정리 | — | **잔존 확인** — 4표면: (a) 제공 집합 `closed_sets.go:82-84`(5.1/4.7/4.5-air 노출, web 셀렉트 2곳에 파생), (b) 상수 `defaults.go:255-261`(호환 로딩 보존 주석), (c) statusline `memory.go:36-42`, (d) legacy struct 멤버 `defaults.go:971-974` | **리드 전제와 다름** — "이미 두 모델"은 기본값 기준 참, 제공 집합 기준 불참. 정리 창구는 (a)뿐. [NC-2] 참조 |
| ⑤ | moai web 위젯: gpt-6.1-sol 선택 노출 | — | 감사 codex 모델 필드는 **TypeText 자유 입력**(`schema_sections.go:388`) — 닫힌 집합 부재(verified-absence). `audit.glm.model`만 셀렉트(`:391-392`). `codexServableModelPrefixes`는 `gpt-` 접두사로 이미 통과(`mcp_codex.go:156`) | **변경 불필요 확정** — [NC-3] 참조 |
| ⑥ | statusline glmContextWindows 1M 유지 검증 | memory.go | `memory.go:34-35` — glm-5.3-flash: 1_000_000(명시 항목), glm-5.3: 1_000_000 | **확인 완료** — 유지, 증거 기록만 |
| 유지 | 템플릿 `"model": "sonnet"` 별칭 | — | `settings.json.tmpl:417` 실존 확인 | **보존 확정** |

### §A.2 설계 결정

1. **핀 도달 경로는 2층이다.** (i) Go 기본값 `NewDefaultWorkflowConfig().Audit` — 전체 Loader 경로와 `moai init` 시딩이 소비. (ii) 감사 해석기의 단말 폴백(`resolveCodexAuditModelEffort` → zero, `resolveClaudeAuditModelEffort` → 하드코드 상수) — 섹션 전용 로더(`audit_pin.go`)가 원본 파일만 읽으므로 Go 기본값을 보지 않는다. 요구사항의 "모든 감사 경로가 새 핀으로 수렴"을 충족하려면 **양층을 모두** 고쳐야 한다. M1이 이 점을 명시한다.
2. **codex 핀의 단말 폴백을 zero에서 새 핀으로 바꾸는 것**은 `model_backend_default_test.go:36-37`의 zero-value 단정을 뒤집는다 — 같은 변경 집합에서 테스트를 뒤집는 것이 정상 경로다(REQ-MMU-007).
3. **claude 핀 값 형태** — [NC-1] 참조.
4. **item ④ 정리 방식** — glm-5.2 철수 선례(`closed_sets.go:76-81`) 재사용. [NC-2] 참조.
5. **i18n과 룰 카탈로그**는 감사 기본값을 산문으로 서술하는 표면이라 REQ-MMU-007의 일관성 의무에 든다. 룰 카탈로그는 로컬+템플릿 미러가 byte-parity 검사로 묶여 있어 같은 커밋에 함께 갱신한다.

## §B Known Issues

- REQ-AMP-005(중립성)와 본 SPEC의 신설 핀이 정면으로 충돌한다. 코드 주석의 대체 기록 없이 지나가면 다음 감사가 "drift"로 오판한다 — audit_models.go 주석 갱신을 M1 필수 작업으로 둔다.
- 로컬 도그푸드 workflow.yaml(`gpt-5.6-sol`)을 놓치면, 템플릿·Go 기본값을 갱신해도 **운영자 트리만 구형 모델에 머문다**(프로젝트 핀이 Go 기본값보다 우선). M1 필수 작업.
- `gpt-6.1-sol`의 z.ai/codex 실서빙 가능성은 본 레인에서 실측 불가다(외부 API 호출 금지). 접두사 가드 통과(`gpt-`)까지만 본 트리에서 검증 가능하며, 실서빙은 운영자 환경의 첫 감사 실행에서 확인된다 — Residual-risk로 기록.

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

### M3 (Priority Medium) — 구형 모델 제공 집합 철수

1. `internal/config/closed_sets.go:82-84` — `ValidGLMModels()`를 {DefaultGLM53Flash, DefaultGLM53}로 축소 + glm-5.2 선례 문체로 철수 기록 주석(5.1/4.7/4.5-air 각각).
2. `internal/config/defaults.go:249-254` — "those exposed by ValidGLMModels() (glm-5.1, glm-4.7, glm-4.5-air) are selectable" 주석을 철수 후 기준으로 갱신(plan-audit D5a).
3. `internal/web/glm_tier_test.go:51-56` — `want` 5멤버 리터럴 목록을 2멤버로 갱신 + TestGLMModelSelectOptions 상단 집합 주석 갱신(plan-audit D3 — 축소 시 보장-RED).
4. `internal/config/defaults_test.go:429-435` — `wantSet` 5멤버 목록을 2멤버로 갱신 + :421 "preserved as a still-available model" 주석 갱신(plan-audit D3 — 축소 시 보장-RED).
5. 영향 확인: web `llm.glm.models.*` 셀렉트와 `workflow.audit.glm.model` 셀렉트는 자동 수렴(파생 집합). `settings/audit_pin_fields_test.go`의 glm 셀렉트 관련 단정은 파생 값이라 무변경.

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
- **`glm-4.x` 상수·statusline 항목 삭제** — 호환 로딩 파괴(Out of Scope 명시).
- **`codexServableModelPrefixes` 무분별 확장** — `gpt-`가 이미 커버한다. 집합 변경은 근거 없는 확장이다.

## §H Cross-References

- `.moai/specs/SPEC-MODEL-MATRIX-UPDATE-001/spec.md` — 요구사항 본문
- `.moai/specs/SPEC-MODEL-MATRIX-UPDATE-001/acceptance.md` — AC 매트릭스
- 선행 계열: SPEC-V3R6-AUDIT-MODEL-PIN-001(핀 우선순위·어휘), SPEC-GLM-EFFORT-MAX-001(세션 기본 max), SPEC-MODEL-MATRIX-CORE/CONFIG/SURFACES/DOCS-001(v3.1.0, 전종 superseded — 계열 계보)
- 증거 경로: `.moai/reports/t1368/`

## [NEEDS CLARIFICATION] — 운영자 판단 후보 (권장안 동반, 블로킹 아님)

### [NC-1] claude 핀의 모델 값 형태 (REQ-MMU-002)

리드 지시는 `{opus-5-5, medium}`이지만, "opus-5-5"는 Claude Code가 인식하는 모델 문자열이 아니다(별칭: opus/sonnet/haiku/fable/opusplan, 전체 id: claude-opus-5-5 — `internal/web/validate.go:41-42`, model-policy.md:15-22). 감사 핀 어휘는 백엔드별 모델 id 형태가 선례(gpt-5.6-sol, glm-5.3)다.
- **권장안**: `claude-opus-5-5`(전체 id) — 별칭 재조준(aliased retarget)에 흔들리지 않는 세대 고정이며, 핀 어휘 형태와 일치. 해석기가 값을 검증 없이 그대로 통과시키므로(`mcp_claude.go:181-197`) 유효 형태를 쓰는 것이 안전하다.
- **대안**: 별칭 `opus`(현재 claude-opus-5-5로 조준, 향후 세대 변동을 따라감).
- **기본 진행**: 권장안으로 진행. 리드가 SPEC 검수 때 대안을 택하면 M1 값만 교체.

### [NC-2] item ④ 정리 범위 (REQ-MMU-004)

리드 전제("tier 슬롯이 이미 두 모델")는 기본값 기준 참이지만 **제공 집합 기준 불참**이다 — 5.1/4.7/4.5-air가 web 셀렉트에 노출 중. 그러나 상수·statusline 항목·legacy struct 멤버는 기존 llm.yaml 호환 로딩을 위해 문서화된 의도 보존이다(defaults.go:249-254, closed_sets.go:76-81).
- **권장안**: glm-5.2 선례 재사용 — 제공 집합(`ValidGLMModels`)에서만 철수(두 5.3 모델만 남김), 상수·statusline·legacy 멤버는 호환 보존. 리드 지시의 "정리"를 사용자 가시 표면(선택지)에 집중 적용하는 해석.
- **대안**: 상수까지 전면 삭제 — 기존 llm.yaml 호환 로딩 파괴로 defaults.go 주석의 보존 근거와 충돌. 비권장.
- **기본 진행**: 권장안으로 진행.

### [NC-3] item ⑤ web 위젯 codex 모델 선택 (검증 결과: 변경 불필요)

감사 codex 모델 필드는 TypeText 자유 입력으로 닫힌 집합이 없다(verified-absence: `schema_sections.go:388`). gpt-6.1-sol은 이미 입력 가능하고 Go 서빙 가드도 통과한다. "선택 노출"을 셀렉트 전환으로 읽으면 신규 기능(범위 확장)이 된다.
- **권장안**: 변경 없음 — verified-absence를 증거로 기록하고 종결. M1의 기본값 갱신이 위젯 표시값을 자연 수렴시킨다.
- **대안**: codex 모델 필드를 셀렉트로 전환 — 신규 기능성 확장으로 본 SPEC 밖.
- **기본 진행**: 권장안(변경 없음)으로 진행.
