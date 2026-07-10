---
id: SPEC-WEB-CONSOLE-012
title: "moai web Stale/Dead Config Surface Cleanup (Track 1)"
version: "0.2.1"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/settings, internal/web"
lifecycle: spec-anchored
tags: "web, console, settings-schema, dead-code, cleanup, llm, glm-fable, research-seam, git-convention"
tier: S
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-010, SPEC-WEB-CONSOLE-011]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.2.1 | 2026-07-10 | manager-spec | plan-audit iter-2 PASS-WITH-DEBT(0.88) N1/N2 해소. N1(major): acceptance.md §D.2 iter-2 증거 블록이 hand-composed였음(vci §3.2 위반 — 좁은 패턴이 매치할 수 없는 `AutoDetectOptions` 라인을 verbatim인 것처럼 수록) — 명령을 실제 실행하여 real observed 출력으로 교체(확장 패턴 `AutoDetection\|AutoDetectOptions` 14줄 기록 + 좁은 패턴 단독은 2줄이라는 사실 병기), whole-struct bind 라인 :237→:234 실측 정정. N2(minor): plan §F M2 REQ 매핑 001-005→001-006 (REQ-WC12-006 편입). |
| 0.2.0 | 2026-07-10 | manager-spec | plan-audit iter-1 FAIL(0.78) D1-D8 정정. D1(CRITICAL): A5 auto_detection 3필드 DEAD→**USED 재분류** — iter-1의 trailing-dot grep(`\.AutoDetection\.`)이 whole-struct bind(`ad := cfg.GitConvention.AutoDetection`)를 구조적으로 매치 불가; live 소비자 = hook_pre_push.go:146→resolveAutoDetectOptions(:223-246)→convention.LoadConvention(manager.go:46), 배선 출처 = SPEC-WEB-CONSOLE-009(**completed** — iter-1의 "미작성" 기재는 stale memory 오류). M3 제거 철회→잔류 확인 반전(REQ-WC12-020 inverted; REQ-WC12-022 withdrawn→A1 보존은 신설 REQ-WC12-006). D2: A1 근거 정정 — GLMModels legacy 필드는 resolveGLMModels(glm.go:720-750, 호출 :794)가 empty-fallback으로 읽음(reader 有); 제거 결정 유지(보존되는 fallback 체인이 곧 하위호환 메커니즘); backfill 인용 :697-704→:663-672 정정. D3: M1 RED-turning 테스트 3건 인벤토리 추가(plan §A.1). D4: AC-WC12-009 `&&` 단락 결함 수정. D5: AC-WC12-017 projectconfig.go 기계 검증 추가. D6: pre-flight §C-5 주석/기계 단언 구분 지침. D7: §D AC 범위 001..020 정합. D8: A5 증거 프로토콜 강화 — field-dot grep + bare-symbol grep 쌍 의무(whole-struct bind/미러 struct 회피 방지). |
| 0.1.0 | 2026-07-10 | manager-spec | 최초 draft. 2026-07-10 orchestrator 감사(Track 1: stale/dead config surface) 5개 발견(A1-A5)을 plan-phase 실측 재검증 후 요구사항화. A4c(bindForm statusline residual)는 실측 no-op 판정 — Out of Scope로 격하. A5 5개 필드는 도구 증거로 USED 2 / DEAD 3 분류 완료 (acceptance.md §D.2). |

---

## §A Context & Motivation

`moai web` 콘솔의 설정 스키마 SSOT(`internal/settings/schema.go` + `schema_sections.go`)와 그 주변 라우팅/문서 표면에 2026-07-10 orchestrator 감사가 5개 stale/dead 결함을 확인했다. 본 SPEC은 그 Track 1 정리분이다. 모든 파일:라인 앵커는 2026-07-10 plan-phase에서 실측 재검증되었다 (plan.md §C에 검증 명령 기록; 라인 번호는 드리프트 가능 — run-phase는 content-token 기준 재확인).

**A1 — Ghost GLM tier 필드 (0.2.0 근거 정정).** `internal/settings/schema_sections.go` `llmFields()`가 `glm.models.{high, medium, low, opus, sonnet, haiku}` 6개 tier를 노출한다. 환경변수 소비자(`internal/cli/glm.go` `setGLMEnv`, :196-199)는 `Models.{High, Medium, Low, Fable}`을 읽으며 `ANTHROPIC_DEFAULT_OPUS_MODEL`조차 `Models.High`에서 온다. struct의 legacy `Opus/Sonnet/Haiku` 필드(`internal/config/types.go` GLMModels — "Legacy fields for backward compatibility")에는 **reader가 존재한다** (iter-1의 "reader 없음" 기술은 오류): `resolveGLMModels`(glm.go:720-750, live `moai glm` 경로에서 :794 호출)가 high/medium/low 빈 값의 empty-fallback으로 세 필드를 읽고, defaults backfill(glm.go:663-672 — iter-1 인용 :697-704는 라인 드리프트)이 빈 값을 채운다. 그럼에도 **웹 노출 제거 결정은 유지된다**: fallback 체인·struct 멤버·backfill은 전부 무접촉 보존되므로(REQ-WC12-006) legacy yaml만 가진 사용자의 런타임 동작은 불변이고, 웹 콘솔이 같은 유효 슬롯(high/medium/low)의 두 번째 이름(opus/sonnet/haiku)을 편집면으로 병렬 노출하는 것은 잉여·혼동 유발이다(canonical 키만 노출). 반면 실제 소비되는 `glm.models.fable`(`ANTHROPIC_DEFAULT_FABLE_MODEL`)은 노출되지 않는다. → legacy alias 3필드 노출 제거 + fable 노출.

**A2 — Dead `research` seam 쓰기 경로.** `research`는 `SectionResearch`(schema.go:40)로 선언되고 `AllSections()`(schema.go:67)에 등재되며, `RouteSeam` 매핑(sectionroute.go:62) + `SeamSections()`(sectionroute.go:81) + seam 쓰기 whitelist(sectionwrite.go:32)에 배선되어 있다. 그러나 FieldDef 0개, 콘솔 탭(`schemaSectionMetas`, web/schemaform.go:65) 미등재 — 렌더도 편집도 불가능한 유령 쓰기 경로다. `internal/research` 패키지는 삭제됨(실측: 디렉터리 부재; 선행 감사에서 research config 12/12 dead 판정). → research를 seam 라우팅/whitelist/섹션 enum에서 전면 제거 (미등재 → `RouteExcluded`, db 선례와 동일 패턴).

**A3 — Stale doc comment.** `internal/web/server.go`(패키지 doc, :13-25) 및 `internal/web/projectconfig.go`(@MX:REASON, :160 인근)가 여전히 `db`를 편집 가능 10섹션과 8-seam 목록에 열거한다. 실제로는 db가 콘솔 표면에서 제거되어 미등재 → `RouteExcluded`다(sectionroute.go:56-58 주석 실측). A2 완료 후에는 research도 같은 열거에서 빠져야 한다. → 두 doc comment를 최종 scope 계약으로 재작성.

**A4 — Dead code 잔재.** (a) `internal/web/assets.go:27-34`의 `errDictKey` keep-alive sentinel — retired html/template `dict` helper의 오류 sentinel(`validate.go:13-15`)을 `unused` linter로부터 지키는 blank reference. 당시 근거였던 "validate.go BYTE-UNCHANGED"(REQ-WC6-004/AC-WC5-010a)는 완결된 선행 SPEC의 시점 제약이며, 상시 byte-guard 테스트 실존 여부는 run-phase 검증 게이트로 둔다(§B.4 Where-gate). (b) `schema_sections.go:62` `WorkflowAgentPurposes()` — 전 리포(테스트 포함) 호출자 0 실측. 함수 주석은 "canonical taxonomy 참조로 유지"를 주장하나(Chesterton's fence 검토됨), taxonomy의 canonical SSOT는 dynamic-workflows.md 표 + `config.Workflow.WorkflowAgents` 맵 키이지 zero-caller Go 함수가 아니다. → 제거.

**A5 — 의심-dead 노출 필드 (0.2.0 재분류: 전원 USED — 제거 대상 0).** verification-claim-integrity.md §1.1 surface 3에 따라 5개 필드의 런타임 reader를 실측했다 (증거 verbatim: acceptance.md §D.2). iter-1은 trailing-dot 패턴(`\.AutoDetection\.`) 단독 grep으로 auto_detection 3필드를 DEAD로 오분류했다 — 이 패턴은 whole-struct bind(`ad := cfg.GitConvention.AutoDetection` 후 로컬 변수로 필드 접근)를 구조적으로 매치할 수 없다(plan-audit iter-1 D1이 반증). 강화 프로토콜(field-dot grep + bare-symbol grep 쌍 — §D.2 preamble)로 재실측한 최종 분류:

| 필드 | 분류 | 근거 (non-test, non-web 행동 reader) |
|------|------|--------------------------------------|
| `quality.tdd_settings.min_coverage_per_commit` | **USED** | `internal/core/quality/trust.go:788` `g.config.TDDSettings.MinCoveragePerCommit` |
| `git_convention.validation.enforce_on_push` | **USED** | `internal/cli/hook_pre_push.go:251-258` (env override 후 config 필드 읽기) — 선행 감사의 dead 반증과 일치 |
| `git_convention.auto_detection.enabled` | **USED** | `hook_pre_push.go:146` → `resolveAutoDetectOptions`(:223-246) whole-struct bind 후 `ad.Enabled` → `convention.LoadConvention`(internal/git/convention/manager.go:46)의 detection gate. 배선 출처 = SPEC-WEB-CONSOLE-009 (completed) |
| `git_convention.auto_detection.confidence_threshold` | **USED** | 상동 — `ad.ConfidenceThreshold` → 감지 수용 임계값으로 전달 |
| `git_convention.auto_detection.sample_size` | **USED** | 상동 — `ad.SampleSize` → Detect 표본 수로 전달 |

→ **제거 대상 0**. M3는 잔류 확인 게이트로 반전(REQ-WC12-020/021): 5필드 전부 노출 유지, A5 관련 스키마/i18n/bridge 무접촉.

---

## §B Requirements (GEARS)

### §B.1 M2 — A1: LLM ghost tier 제거 + fable 노출

**REQ-WC12-001 (Ubiquitous):** The web settings schema (`llmFields()`) shall expose exactly the four GLM model tiers `glm.models.{high, medium, low, fable}` as editable fields.

**REQ-WC12-002 (Ubiquitous — shall not, 0.2.0 rationale 정정):** The web settings schema shall not expose `glm.models.{opus, sonnet, haiku}` — legacy backward-compat aliases whose only runtime reads are the `resolveGLMModels` empty-fallback normalization (glm.go:720-750) and the defaults backfill (glm.go:663-672). Because that fallback chain is preserved untouched (REQ-WC12-006), removing the web exposure changes no runtime behavior; it removes only the redundant second editing surface for the same effective high/medium/low slots.

**REQ-WC12-006 (Ubiquitous — 0.2.0 신설):** The A1 removal shall be schema-surface-only: the `GLMModels` legacy struct members (`Opus`/`Sonnet`/`Haiku`), the `resolveGLMModels` empty-fallback chain, and the legacy defaults backfill shall be preserved unmodified, so that a legacy yaml carrying only `opus`/`sonnet`/`haiku` keys resolves to identical effective models before and after this SPEC.

**REQ-WC12-003 (When):** When a console save submits `llm.glm.models.fable`, the typed apply path (`applyLLMKey`) shall persist the value to `LLMConfig.GLM.Models.Fable`, and shall no longer carry apply branches for the three removed ghost keys.

**REQ-WC12-004 (Ubiquitous):** The interface-i18n dictionary (`internal/web/assets/i18n.js`) shall carry `f.llm.glm.models.fable.{title,desc}` entries in every locale block that carries the sibling `f.llm.glm.models.high.{title,desc}` entries, and shall not carry orphan `f.llm.glm.models.{opus,sonnet,haiku}.*` entries.

**REQ-WC12-005 (Ubiquitous):** The LLM section display meta (`schemaSectionMetas()`, internal/web/schemaform.go) shall describe the real tier set (high/medium/low/fable), not the retired six-tier enumeration.

### §B.2 M1 — A2: research seam 쓰기 경로 폐선

**REQ-WC12-010 (Ubiquitous):** The section routing SSOT shall classify `research` as `RouteExcluded` by non-registration — removed from the `sectionRoutes` map, from `SeamSections()`, and from the `sectionRootKeys` seam write whitelist.

**REQ-WC12-011 (Ubiquitous — shall not):** The settings schema shall not declare a `SectionResearch` section ID nor list it in `AllSections()`; surrounding comments enumerating "7개 seam 섹션" shall be corrected to the remaining six.

**REQ-WC12-012 (When):** When a write to section `research` is attempted via `WriteSectionViaSeam`, the seam shall reject it with the not-seam-writable error and shall not touch `research.yaml` — the same guard behavior already exercised for `db` and other unregistered sections.

### §B.3 M3 — A5: 재분류 결과 전원 USED — 잔류 확인 게이트 (0.2.0 반전)

**REQ-WC12-020 (Ubiquitous — 0.2.0 inverted):** The web settings schema shall retain the three `git_convention.auto_detection.{enabled, confidence_threshold, sample_size}` FieldDefs (schema.go:440-458) unchanged — reclassified USED per the acceptance.md §D.2 evidence (live consumer: `resolveAutoDetectOptions` → `convention.LoadConvention`, wired by completed SPEC-WEB-CONSOLE-009). No FieldDef, i18n key, or bridge entry for these fields shall be removed or altered.

**REQ-WC12-021 (Ubiquitous):** The web settings schema shall retain `quality.tdd_settings.min_coverage_per_commit` and `git_convention.validation.enforce_on_push` — classified USED by the same evidence protocol.

**REQ-WC12-022 — WITHDRAWN (0.2.0):** iter-1의 A5 제거 전제(schema-surface-only removal + struct 보존)가 D1 재분류로 소멸. A1의 struct/fallback 보존 의무는 신설 REQ-WC12-006이 승계. (ID는 재사용하지 않고 철회 기록으로 보존.)

**REQ-WC12-023 (Ubiquitous — shall not):** The cleanup shall not touch the distinct `harness.auto_detection.enabled` field (schema_sections.go:215) — a live harness-section field that shares only the key substring. (0.2.0 note: A5 무접촉 반전으로 이 guard는 방어적 잔류.)

### §B.4 M4 — A4: dead code 잔재 제거

**REQ-WC12-030 (Where):** Where the run-phase pre-flight confirms that no standing test asserts byte-content or hash of `internal/web/validate.go` (the historical REQ-WC6-004 constraint being a closed-SPEC point-in-time assertion), the `errDictKey` sentinel (validate.go:13-15) and its keep-alive blank reference (assets.go:27-34) shall both be removed.

**REQ-WC12-031 (When):** When the pre-flight instead finds a standing byte-content guard on validate.go, the implementation shall retain the sentinel, update the assets.go keep-alive comment to cite the found guard by file:test-name, and report the constraint in the completion report (no silent skip).

**REQ-WC12-032 (Ubiquitous):** The zero-caller function `WorkflowAgentPurposes()` (schema_sections.go:62) shall be removed after a run-phase re-grep confirms the caller count is still zero; the 7-purpose taxonomy remains documented by its canonical SSOT (dynamic-workflows.md + `config.Workflow.WorkflowAgents` map keys).

### §B.5 M5 — A3: stale doc comment 정정

**REQ-WC12-040 (Ubiquitous):** The `internal/web/server.go` package doc and the `internal/web/projectconfig.go` @MX:REASON block shall enumerate the post-cleanup scope contract — typed sections (git-strategy, llm), the six seam sections (workflow, harness, ralph, feedback, observability, security), and shall name `db` and `research` among the excluded set instead of the editable set.

### §B.6 Cross-cutting guards

**REQ-WC12-050 (When):** When any schema FieldDef change lands, both the web-side parity tests (`internal/web` + `internal/settings` schema/i18n tests) and the TUI-side `internal/cli/schema_bridge_test.go` shall be updated together and pass — the SPEC-WEB-CONSOLE-011 M2b regression (web-only update broke 3 TUI tests) shall not recur.

**REQ-WC12-051 (Ubiquitous — shall not):** The implementation shall not modify any file under `internal/statusline/` (owned by a parallel session) nor any file under `internal/template/templates/` (the schema lives in Go code; the template llm.yaml already carries the correct key set).

---

## §C Scope Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — struct/yaml 하위호환 표면

- `GLMModels`의 legacy `Opus/Sonnet/Haiku` struct 필드, `resolveGLMModels` fallback 체인(glm.go:720-750), defaults backfill(glm.go:663-672), `AutoDetectionConfig` struct/defaults/validation — 전부 보존 (yaml 하위호환; REQ-WC12-006 + REQ-WC12-020).
- 라이브 `.moai/config/sections/llm.yaml`에 이미 기록된 ghost 키 값 정리 — 런타임 config 파일 무접촉 (struct가 legacy 필드를 유지하므로 typed re-marshal이 기존 키를 파괴하지 않음).

### Out of Scope — A4c bindForm statusline residual

- plan-phase 실측 결과 no-op: `internal/web/handlers.go` `bindForm`(:449-471) 본문에 statusline 바인딩 분기가 이미 존재하지 않으며, doc comment도 정확하다(M3 redesign 반영). `ProfilePreferences`의 statusline struct 필드는 TUI/CLI + statusline.yaml sync 소비자가 있어 무접촉.

### Out of Scope — git_convention auto-detection 동작 변경

- auto_detection 엔진은 **completed** SPEC-WEB-CONSOLE-009가 이미 배선한 **live 경로**다 (`hook_pre_push.go:146` → `resolveAutoDetectOptions` → `convention.LoadConvention`). 본 SPEC은 이 동작 경로도, 그 노출 필드도 일절 변경하지 않는다 (REQ-WC12-020 잔류; iter-1의 "미래 배선" 기술은 stale memory 오류로 0.2.0에서 정정됨).

### Out of Scope — 병렬 세션 소유 파일 및 템플릿 미러

- `internal/statusline/*` (병렬 세션이 수정 중 — renderer.go, cache_hit_test.go).
- `internal/template/templates/**` 전체 (llm.yaml 템플릿은 이미 정답 키셋 — 변경 불요 실측).

---

## §D Acceptance Criteria

정식 AC 매트릭스(기계 검증 명령 + 기대 출력 + A5 per-field 증거 verbatim)는 `acceptance.md`가 SSOT다. 요약: AC-WC12-001..020, Given-When-Then 시나리오 3건, edge case 3건.
