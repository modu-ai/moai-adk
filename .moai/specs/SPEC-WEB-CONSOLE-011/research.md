---
id: SPEC-WEB-CONSOLE-011
status: completed
created: 2026-07-03
updated: 2026-07-05
---

# Research — SPEC-WEB-CONSOLE-011

> **출처**: §A-§H = 2026-07-03 orchestrator survey (workflow `wf_d19d522a-d39`), 위임 프롬프트로 전달됨. §I = 결정 변경 기록 (사용자 직접 선택, AskUserQuestion, 2026-07-03) + plan-phase 추가 실측.
> **재검증 의무**: 아래 파일:라인 앵커는 survey 시점 실측값이며 라인 번호는 드리프트할 수 있다. run-phase 착수 시 plan.md §C의 content-token grep으로 전수 재확인 후 사용한다 (line-number drift asymmetry 교훈).
> **v0.2.0 주의**: §C-§D의 "staged write / 선택 확장 / read-only" 서술 일부는 §I의 확정 결정으로 **supersede**되었다 — 해당 위치에 개정 표기.

## §A 현재 콘솔 상태 (verified)

- 단일 페이지 HTMX+Templ, loopback 전용. 라우트: `GET /`, `POST /save`, `POST /__shutdown__`, `GET /static/*` (internal/web/app.go:78-88).
- 공유 SSOT = `internal/settings/schema.go` — 6 섹션 / 34 필드 (schema.go:24-31, allFields 233-430). web + TUI wizard 공동 소비 (SPEC-WEB-CONSOLE-010 확립).
- 영속화 4 seam (handleSave, handlers.go:287-316): `profile.WritePreferences` / `profile.SyncToProjectConfig`(sync.go:17 — statusline은 sync.go:103에서 ConfigManager 우회 직접 yaml.Marshal) / `writeProjectConfig`(projectconfig.go:216) / `settings.WriteProjectNestedConfig`(nested.go:102).
- 쓰기 안전 모델 = loopback bind + Host-check 미들웨어. @MX:NOTE app.go:90-92이 CSRF/token 인프라 금지 (REQ-WC-009) — **보존 대상**.

## §B 기반 차단 요소 (M1 대상, verified)

- **Scope contract**: REQ-WC-012 (internal/web/server.go:10-13) + REQ-WC3-007 (projectconfig.go:158 @MX:WARN)이 workflow/harness/git-strategy/llm yaml 접촉 금지. 본 SPEC이 공식 supersede해야 하며 guard tests (`projectconfig_scope_test.go`, `coverage_test.go`)도 함께 갱신 필요. **[v0.2.0: supersede 범위가 10섹션 전체로 확장 — §I 결정 3]**
- **ConfigManager.Save()**: 6개 파일만 기록 (manager.go:166, 207-219 — user/language/quality/git-convention/git-strategy(dirty-flag)/llm). struct 재직렬화는 yaml 주석 + 미모델링 키 파괴.
- **workflow.yaml 특이점**: `RoleProfileEntry`(internal/config/types.go:373-378)는 명시적 결정으로 Effort 필드 부재 (REQ-WEM-006, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001); `team.patterns`는 의도적 미모델링 (EXCL-WSE-004, types.go:359-360). → naive typed save는 첫 쓰기에서 파일 손상. comment/unknown-key-preserving yaml.Node patch seam (gopkg.in/yaml.v3 node surgery)이 M1 필수. **[v0.2.0: seam은 Save() 부재 8섹션 전부의 load-bearing 의존성으로 격상 — §I 결정 3]**

## §C Config 확장 실측 (M2) **[v0.2.0: "선택 확장"은 §I 결정 3으로 "10섹션 전면 확장"으로 supersede]**

- **git-strategy**: 가장 저렴 — 완전 typed + Save dirty-flag 경로 존재 (SPEC-GITSTRATEGY-SAVE-ISOLATION-001, manager.go:207-216).
- **llm.yaml**: 완전 typed (`LLMConfig` types.go:234-283, oneof 검증). claude_models tiers는 현재 빈 문자열 — "빈 값 = runtime default" 시맨틱스를 EmptyLabelKey 패턴(schema.go:225-229) 재사용으로 정의.
- 신규 필드는 기존 `FieldDef` + `PersistTarget` 패턴을 따르고, 필드당 i18n 키 ×4 locale (`internal/web/assets/i18n.js`).

## §D Agent 설정 4표면 실측 (M3) **[v0.2.0: (c)(d)의 read-only 방침은 §I 결정 4로 "전면 쓰기"로 supersede]**

- **(a) llm.yaml tiers**: SMALL, typed (§C 참조).
- **(b) team role profiles**: workflow.yaml `team.role_profiles` map L81-125 (7 profiles × description/effort/isolation/mode/model), `default_model: opus[1m]` L18, `role_profile_keys` L29-32. Go typing 95% 완료 — effort만 예외 (REQ-WEM-006). 편집은 M1 yaml.Node seam 경유만; enum 검증은 `internal/harness/v4manifest/schema.go:41-73` closed sets (5 efforts, 4 model tiers) 재사용 가능.
- **effort 결정 요청**: effort를 Go-invisible로 유지하고 opaque node로 패치 (REQ-WEM-006 유지) vs `RoleProfileEntry`에 Effort 추가 (SPEC amendment). **survey 권고 = 전자** → design.md §B에서 Option A로 채택 (v0.2.0 전면 쓰기 하에서도 유지 확인).
- **(c) sub-agent frontmatter**: `.claude/agents/moai/*.md` 7 agents — builder-harness model:inherit/effort:high; manager-develop inherit/xhigh; manager-spec inherit/xhigh; plan-auditor inherit/xhigh; sync-auditor inherit/xhigh; manager-docs haiku(effort 부재); manager-git haiku(effort 부재). ~~본 SPEC에서 READ-ONLY~~ **[v0.2.0 supersede: 편집 — frontmatter 전용 patch layer + live-only + 지속 경고, §I 결정 4(c)]**. effort 키 부재는 유효 상태 (2 agents 선례).
- **(d) dynamic workflow model/effort**: SSOT는 prose (`.claude/rules/moai/workflow/dynamic-workflows.md` L82-103, 7-purpose taxonomy) + harness.yaml `effort_mapping` L19-22 (typed, `HarnessConfig.EffortMapping` types.go:662-664). ~~READ-ONLY 렌더~~ **[v0.2.0 supersede: 편집 — 신규 typed `workflow_agents:` 표면(workflow.yaml) 경유, §I 결정 4(d). harness.yaml effort_mapping은 harness 섹션 스칼라 키로서 M2 경로에서 편집; machine-readable 신규 표면의 "Out of Scope" 분류는 폐기됨]**

## §E Profile CRUD 실측 (M4)

- 백엔드 primitives 존재: `profile.List` / `Delete` / `EnsureDir` (delete guards profile.go:87-110).
- **보안 전제 (UNVERIFIED HYPOTHESIS)**: 웹 쓰기 경로가 임의 `?profile=` / `__profile` (app.go:133-141, handlers.go:252-254)을 `GetPreferencesPath`/`WritePreferences` (preferences.go:82-88, 150-165)로 흘리며 `isValidProfileName`(profile.go:126-132) 미적용 — path traversal(`__profile=../../x`) + MkdirAll 경유 undocumented implicit-create 가능성. **repro test FIRST 의무** (verification-claim-integrity §1.1 surface 3 — 가설은 도구 검증 전까지 결함이 아님).
- CRUD UI = ~2 신규 POST 라우트 + Templ fragments + i18n; delete는 "default"와 active profile 거부.

## §F SPEC 보드 실측 (M5)

- 데이터 레이어 완비: `spec.Audit` (internal/spec/audit.go:94, pure FS scan, **실측 0.14s / 412 SPECs**, git 무호출), `spec.ExtractFrontmatter`(lint.go:346), `spec.ParseStatus`(status.go:64), 8-value status enum (status.go:13-23).
- 라이브 분포: completed 235 / implemented 138 / archived 31 / superseded 5 / draft 2 / in-progress 1 → 핵심 가치 = **implemented-not-completed close-debt 열** + MUST-FIX SyncStatusDrift badges (audit.go:255-300). 고전적 WIP kanban 아님.
- 전제 코드 2건: `SPECFrontmatter`에 `Tier string yaml:"tier"` 추가 (lint.go:257-287; tier 보유 177/412 — optional badge) + exported `spec.ListDocs(baseDir)` wrapper (discoverSPECs/parseSPECDoc = lint.go:224/311 unexported).
- MUST-FIX remediation 문자열은 **COPYABLE TEXT**로만 렌더 (예: `moai spec close <ID> --backfill-only`) — 핸들러에서 실행 금지.
- git 의존 `DetectDrift` 경로 (**실측 7.9s**)는 동기 렌더에서 제외.
- 쓰기 경로 없음: status 전이는 agent-owned (Status Transition Ownership Matrix, spec-frontmatter-schema.md; hook `status-transition-ownership.sh` owner mismatch 시 exit 2).

## §G Statusline cache_hit 3-way orphan 실측 (M6)

- `cache_hit` 세그먼트는 renderer에서 toggleable (`SegmentCacheHit` types.go:329, default-on, hand-edited yaml honored)이나 다음 표면 전부에서 **부재**:
  - `settings.statuslineSegmentKeys` (settings/schema.go:106-122)
  - `statusline.CanonicalSegments` (preset.go:12-18)
  - profile `defaultStatuslineSegments` (sync.go:155-178)
  - TUI `statuslineAllSegments`
  - statusline.yaml 양쪽 (live + template mirror `internal/template/templates/.moai/config/sections/statusline.yaml` — Template-First 적용)
  - i18n.js
- 노출 fan-out ≈ 10-12 파일: count-label fieldsets.templ:135 ("16 fields") + web/statusline_test.go:53-54 (want 15→16) 포함. ~~schema 필드 카운트 34→35~~ **[v0.2.0 supersede: M2 전면 확장으로 총 필드 수 유동 — 하드코딩 총계 assertion을 파생 방식으로 전환, §I]**
- 3+ 하드코딩 세그먼트 목록을 단일 SSOT에 묶는 set-equality 테스트 필요 (향후 orphan 재발 방지).
- stale comment profile_setup.go:333 ("11-segment") passing 정정.
- **제외**: env-only knobs (`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, `MOAI_STATUSLINE_CONTEXT_SIZE`), 렌더 상수 (bar width, handoff thresholds — context-window-management.md HARD rule 고정).
- **주의**: `internal/statusline/renderer.go` + `cache_hit_test.go`는 병렬 세션이 수정 중 — 본 SPEC 무접촉.

## §H 참조 소스

- SPEC-WEB-CONSOLE-010 산출물 세트 (`.moai/specs/SPEC-WEB-CONSOLE-010/`) — 시리즈 선례 (frontmatter/FieldDef/i18n 패턴).
- `.moai/config/sections/workflow.yaml` L18/L29-32/L81-125 — role profiles 실측 앵커.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix.
- `.claude/rules/moai/core/verification-claim-integrity.md` — M4 repro-first 의무 + 본 문서 실측치 인용 규율.

## §I 결정 변경 기록 — 사용자 직접 선택 (AskUserQuestion, 2026-07-03) + plan-phase 추가 실측

### §I.1 결정 변경 2건 (v0.1.0 → v0.2.0)

**결정 3 — Config 커버리지: 선택 확장 → 10섹션 전면 확장.**
- 편집 대상 10섹션: git-strategy, llm, workflow, harness, ralph, research, feedback, observability, security, db.
- 존속 제약: (a) machine/state 제외 (state, system, project, cache, sunset); (b) 대형 정책 파일 제외 (tool-policy, lsp, mx); (c) 커버 섹션 내 runtime-managed 키 read-only — llm.mode/team_mode, db 5 system 키(orm/multi_tenant/migration_tool 3개 인터뷰 키만 편집); (d) map-of-structs 서브블록(harness levels 내부, workflow team.patterns)은 read-only/collapsed raw — 스칼라 키는 폼.
- 구조적 귀결: **Save() 경로 없는 8섹션(workflow, harness, ralph, research, feedback, observability, security, db) 전부가 M1 yaml.Node seam에 의존** — seam이 SPEC 전체의 load-bearing 기반으로 격상.

**결정 4 — Agent 설정: staged write → 4표면 전면 쓰기.**
- (a) llm tiers 편집 (기존과 동일); (b) role profiles seam 편집 (opaque-node effort 유지 — design.md §B 재확인); (c) sub-agent frontmatter **편집** — frontmatter 전용 read/patch layer (body 무접촉) + live-only 기록 + "moai update 덮어쓰기 가능" 지속 경고, template dual-write 없음 (미러는 dev repo 전용); 검증 = v4manifest 4-tier model / 5-value effort, effort 부재 유효; (d) dynamic-workflow **편집** — workflow.yaml 신규 typed `workflow_agents:` 블록 (7-purpose map → {model, effort}), internal/config 신규 struct + loader, v4manifest 검증, seam 기록; dynamic-workflows.md (live+mirror)를 "config = 기본값 SSOT / per-script 리터럴 = override"로 갱신 (Template-First + make build, 별도 work item).

결정 1 (SPEC 보드 READ-ONLY) / 결정 2 (Profile CRUD + repro-first)는 무변경.

### §I.2 plan-phase 추가 실측 (2026-07-03, `ls`/`grep` 직접 관측)

- `.moai/config/sections/` 실존 파일: **10개 대상 섹션 전부 실존** ✓ + 제외군(state, system, project, cache, sunset, tool-policy, lsp, mx) 실존 ✓.
- **미지명 섹션 4종 발견**: constitution.yaml, context.yaml, design.yaml, interview.yaml — 사용자 10-목록에도 제외 열거에도 없음 → "지명되지 않은 섹션은 비노출" 원칙으로 REQ-WC11-018에 편입 (내부 불일치 해소 기록).
- `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` **실존** ✓ (17,833B) — REQ-WC11-074의 live+mirror 갱신 AC 유효 ("Template tree = SUBSET of live" 함정 검증 통과).
- `internal/template/templates/.moai/config/sections/workflow.yaml` **실존** ✓ (5,734B) — 템플릿 기본 블록 추가 대상 확정.
- `grep -c workflow_agents .moai/config/sections/workflow.yaml` = **0** — 블록 신설 확정 (최초 기록 upsert 필요, design.md §A.2).

### §I.3 §I가 supersede하는 본문 서술

- §C 표제 "선택 확장" → 10섹션 전면 (결정 3).
- §D (c) READ-ONLY / (d) READ-ONLY + "machine-readable 표면 Out of Scope" → 전면 쓰기 + workflow_agents 신설 (결정 4).
- §G "schema 필드 카운트 34→35" → 파생 assertion (총 필드 수 유동).
