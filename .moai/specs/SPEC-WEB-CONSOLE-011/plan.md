---
id: SPEC-WEB-CONSOLE-011
status: draft
created: 2026-07-03
updated: 2026-07-03
---

# Plan — SPEC-WEB-CONSOLE-011 (v0.2.1)

## §A Context

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (main checkout, Route B 여부는 Tier L이므로 orchestrator가 Tier-based PR Routing으로 결정).
- **SPEC 산출물**: `.moai/specs/SPEC-WEB-CONSOLE-011/{spec,plan,acceptance,design,research,progress}.md` (Tier L 5-artifact + progress skeleton).
- **스코프 SSOT**: spec.md §1.1의 **확정 4결정 (v0.2.0, 2026-07-03 사용자 직접 선택)**. 재론 금지. v0.1.0의 "선택적 확장 / staged write" 표현은 전부 개정됨.
- **근거 조사**: research.md §A-§H (2026-07-03 orchestrator survey, workflow wf_d19d522a-d39) + §I (결정 변경 기록 + plan-phase 추가 실측). 모든 파일:라인 앵커는 survey 시점 실측 — run-phase 착수 시 §C에서 content-token 기준 재검증.
- **주요 대상 패키지**: `internal/web`, `internal/settings`, `internal/config`, `internal/profile`, `internal/spec`, `internal/statusline`(preset.go만), `internal/cli`(profile_setup.go), `.claude/agents/moai/`(frontmatter만), `.claude/rules/moai/workflow/dynamic-workflows.md`(live+mirror), `internal/template/templates/.moai/config/sections/`(statusline.yaml + workflow.yaml mirror).

### §A.5 PRESERVE 목록 (절대 접촉 금지)

- `internal/statusline/renderer.go` + `internal/statusline/cache_hit_test.go` — **병렬 세션이 수정 중** (git status M). M6은 이 두 파일을 수정하지 않는다 (segment 구현은 이미 존재; M6은 노출 fan-out만).
- `internal/web/app.go:90-92` @MX:NOTE (REQ-WC-009) — CSRF 금지 계약 보존.
- `~/.moai/.env.glm` 및 GLM secrets 경로 일체.
- **agent 파일 body** — frontmatter patch layer는 `.claude/agents/moai/*.md`의 YAML frontmatter만 수정, body는 byte 단위 무접촉 (REQ-WC11-027).
- machine/state 섹션 + 정책 파일 + 미지명 섹션 (state, system, project, cache, sunset / tool-policy, lsp, mx / constitution, context, design, interview) — 웹 노출·쓰기 제외 (REQ-WC11-018).
- runtime-managed 파일 (`.moai/state/*`, `.moai/cache/*`, `.moai/logs/*`).

## §B Known Issues

- **B1 (병렬 세션 레이스)**: renderer.go / cache_hit_test.go가 다른 세션에서 수정 중. M6 착수 전 해당 작업의 landing 여부를 `git status` + `git log --oneline -5 -- internal/statusline/`으로 확인. 미landing 시 M6 보류 + blocker report.
- **B2 (Cross-SPEC 정책 충돌)**: (i) REQ-WEM-006 — Effort 필드 추가 금지, opaque-node 방식만 허용 (design.md §B — 전면 쓰기 하에서도 유지); (ii) EXCL-WSE-004 — team.patterns 미모델링 유지 (collapsed raw view, REQ-WC11-062); (iii) REQ-WC-012 / REQ-WC3-007 — M1에서 **10섹션 전체 계약**으로 공식 supersede (v0.2.0 확장 — v0.1.0의 "harness 금지 유지" 조항은 폐기됨, 제외군은 REQ-WC11-018로 이동); (iv) STATUSLINE-PRESET-RETIRE-001 — preset 셀렉터 재도입 금지.
- **B3 (Subagent boundary)**: `internal/web`, `internal/spec` 신규 코드에 AskUserQuestion 참조 금지 (C-HRA-008 family grep 0 매치).
- **B4 (Frontmatter schema)**: SPEC 산출물은 canonical 필드명(`created`/`updated`/`tags`)만 사용.
- **B6 (Out of Scope lint)**: spec.md §3은 `### Out of Scope — <topic>` h3 + `-` bullet 구조 준수 (v0.2.0에서 7개 h3로 재편).
- **B8 (Working tree hygiene)**: 커밋은 specific-path `git add`만. 병렬 세션 산출물 (`reports/moai-adk-v3-preview-*.html`, docs-site 수정분, `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/`, `SPEC-HARNESS-EVO-RUN-REPORT-001/`) 절대 staging 금지.
- **B-템플릿 (Template-First + Neutrality)**: statusline.yaml / workflow.yaml / dynamic-workflows.md 템플릿 미러 편집은 live와 같은 마일스톤에서, §25 neutrality (SPEC ID/REQ 토큰/날짜 주석 금지) 준수, `make build`로 재임베드. **미러 실존은 2026-07-03 plan-phase에서 3개 대상 전부 `ls` 확인 완료** (dynamic-workflows.md ✓, workflow.yaml ✓, statusline.yaml ✓ — "Template tree = SUBSET of live" 함정 회피).
- **B-yaml.v3 (정규화 노이즈)**: yaml.v3 Encoder는 노드 트리 재직렬화 시 일부 포매팅을 정규화할 수 있다. **8개 seam 섹션 각각의** golden-file round-trip 테스트로 byte-diff를 검증하고, 허용 불가 수준의 노이즈 발견 시 blocker report (design.md §A.4).
- **B10 (i18n 인플레이션 — v0.2.0 신규)**: 10섹션 폼 + agent settings + 보드 + CRUD로 신규 i18n 키가 수십 개(추정 40-80) × 4 locale 규모. blind sed 금지; 키 네이밍 규약 일관성 + parity 테스트 (AC-WC11-061)로 관리. 마일스톤별 i18n 동반 원칙 (REQ-WC11-015) 엄수.
- **B11 (파생 카운트 전략 — v0.2.0 신규)**: 스키마 총 필드 수가 34를 크게 초과하며 유동. 하드코딩 총계 assertion(구 34→35 pin) 금지 — 카운트 라벨/테스트는 schema 길이에서 파생 (REQ-WC11-053). statusline-국소 카운트(segment 16)는 명시 허용.
- **B12 (섹션 키 미실측 — v0.2.0 신규)**: ralph/research/feedback/observability/security/db의 개별 키 목록과 "Save() 부재 8섹션" 명제는 사용자 확정 입력이다. 파일 실존은 확인(§I)됐으나 키 목록·typed struct 유무는 run-phase pre-flight에서 기계 검증 후 FieldDef 설계 확정 (vci §2 baseline 귀속).
- **B13 (frontmatter patch 견고성 — v0.2.0 신규)**: 7개 agent 파일의 frontmatter 형식 편차(effort 키 부재 2건: manager-docs/manager-git) 대응 — 부재는 유효 상태, 빈 값 주입 금지 (REQ-WC11-029, EC-7).

## §C Pre-flight (run-phase 착수 전 의무 검증)

```bash
# 1. baseline
git branch --show-current && git rev-parse HEAD

# 2. 병렬 세션 statusline 작업 landing 확인 (B1)
git status --porcelain internal/statusline/ && git log --oneline -5 -- internal/statusline/

# 3. cross-platform build baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...

# 4. lint baseline (NEW vs pre-existing 구분용)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 5. survey 앵커 재검증 (content-token 기준; 라인 번호 드리프트 대비)
grep -n "statuslineSegmentKeys" internal/settings/schema.go
grep -n "CanonicalSegments" internal/statusline/preset.go
grep -n "isValidProfileName" internal/profile/profile.go
grep -n "REQ-WC-009" internal/web/app.go
grep -n "REQ-WC3-007\|REQ-WC-012" internal/web/projectconfig.go internal/web/server.go

# 6. spec.Audit 성능 재확인 (M5 동기 렌더 전제)
go test -run TestAudit ./internal/spec/ 2>&1 | tail -3

# 7. (v0.2.0) 10섹션 키 실측 열거 + db 5 system 키 확정 (B12)
for f in git-strategy llm workflow harness ralph research feedback observability security db; do \
  echo "== $f =="; cat .moai/config/sections/$f.yaml; done

# 8. (v0.2.0) typed struct 유무 실측 — "Save() 부재 8섹션" 명제 기계 검증
grep -n "ralph\|research\|feedback\|observability\|security\|db" internal/config/types.go internal/config/manager.go | head -30

# 9. (v0.2.0) 7-purpose taxonomy 키 열거 (workflow_agents map 키 확정)
grep -n "^###\|^- \*\*" .claude/rules/moai/workflow/dynamic-workflows.md | sed -n '1,30p'

# 10. (v0.2.0) 미러 실존 재확인 (plan-phase 2026-07-03 확인 완료 — 재확인)
ls internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md \
   internal/template/templates/.moai/config/sections/workflow.yaml \
   internal/template/templates/.moai/config/sections/statusline.yaml

# 11. (v0.2.0) agent frontmatter 실측 (7 agents model/effort 현황)
head -10 .claude/agents/moai/*.md | grep -n "model:\|effort:\|=="
```

## §D Constraints (DO NOT VIOLATE)

- Save() 경로 없는 8개 섹션(workflow, harness, ralph, research, feedback, observability, security, db)에 typed re-marshal 금지 — seam 경유만 (REQ-WC11-005/017).
- REQ-WC11-018 제외군(machine/state 섹션, 정책 파일, 미지명 섹션) 노출·쓰기 금지.
- agent frontmatter patch는 **body byte-무접촉** + **live-only** (template dual-write 금지, REQ-WC11-027).
- CSRF/token 인프라 도입 금지 (REQ-WC11-060).
- 보드 핸들러에서 `DetectDrift` 동기 호출 금지 / 원격 명령 실행 금지 / status 쓰기 금지 (REQ-WC11-044/045/046).
- `RoleProfileEntry`에 `Effort` 필드 추가 금지 (REQ-WC11-023).
- preset 셀렉터 재도입 금지 (STATUSLINE-PRESET-RETIRE-001).
- M4는 repro test RED 확인 전 수정 커밋 금지 (Reproduction-First).
- 신규 UI 문자열은 4-locale i18n 동반 의무 (REQ-WC11-015/061).
- 스키마 **총** 필드 수 하드코딩 assertion 금지 — 파생 방식만 (B11).
- `--no-verify` / force-push 금지; Conventional Commits + `🗿 MoAI` trailer.
- §A.5 PRESERVE 목록 외 무관 파일 변경 금지.

## §E Self-Verification

- acceptance.md §D AC Matrix가 SSOT. manager-develop 완료 보고는 E1(AC PASS/FAIL matrix) ~ E7(blocker report)를 포함해야 한다 (`manager-develop-prompt-template.md` §E).
- 5-section evidence 형식 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) 준수 — verification-claim-integrity.md §3.

## §F Milestones (우선순위 순서 — 시간 추정 없음)

의존 그래프: **M1 → M2a → {M2b, M3}** (M3의 seam 의존 작업은 M2a의 workflow.yaml 라우팅에 의존); M4 / M5 / M6 상호 독립 (M6은 B1 landing gate).

### M1 — Foundation (Priority: Highest; 전체 차단 해소)

- REQ-WC-012 + REQ-WC3-007 scope contract를 **10섹션 계약**으로 공식 supersede: server.go/projectconfig.go 주석 갱신 + guard test 2종 신규 계약 반영 (10섹션 허용 + 제외군 거부 케이스).
- yaml.Node patch seam 구현 (배치: design.md §A — `internal/settings/yamlpatch` 제안): 스칼라 교체 + **누락 경로 upsert** (v0.2.0 확장 — workflow_agents 최초 기록용) + golden-file round-trip 테스트 인프라.
- AC 바인딩: AC-WC11-001..005.

### M2a — 8섹션 seam persistence 인프라 (Priority: Highest; M1 직후)

- 섹션 키 실측 열거(§C-7/8) → 8개 no-Save-path 섹션(workflow, harness, ralph, research, feedback, observability, security, db) 각각의 seam 쓰기 라우팅 + per-section golden fixture round-trip 테스트 (주석/unknown-key 보존).
- db 5 system 키 확정 + read-only 분리 준비.
- AC 바인딩: AC-WC11-017.

### M2b — 10섹션 fieldsets + i18n (Priority: High; M2a 이후)

- git-strategy 전체(typed dirty-flag) + llm 안전 키 + quality 잔여 + 8섹션 스칼라 키 FieldDef/뷰모델 + 폼 렌더.
- llm.mode/team_mode read-only; db 3키 편집/5키 read-only; 빈 tier "(runtime default)"; map-of-structs collapsed raw view (REQ-WC11-062).
- i18n ×4 (B10 규약 준수) + 파생 카운트 라벨 전환 (B11).
- AC 바인딩: AC-WC11-010..016, 018, 019, 063.

### M3 — Agent Settings 전면 쓰기 (Priority: High; M1+M2a 의존)

- 4표면 편집 뷰: llm tiers(typed) / role_profiles(seam, opaque effort) / sub-agent frontmatter(신규 patch layer) / workflow_agents(신규 typed 표면).
- frontmatter patch layer: frontmatter-only 파싱/패치 + body byte-보존 + idempotency + 지속 경고 UI + v4manifest closed-set 검증 (effort 부재 유효).
- `workflow_agents`: Go struct + loader wiring(internal/config) + seam upsert 기록 + enum 검증 + UI.
- **별도 work item**: dynamic-workflows.md live+mirror SSOT 참조 갱신 + template workflow.yaml mirror 기본 블록 + `make build` (Template-First + §25 neutrality).
- AC 바인딩: AC-WC11-020..029, 070..074.

### M4 — Profile CRUD (Priority: High; 독립)

- Step 1 (RED): path-traversal / implicit-create repro test 작성 → FAIL 관측 (가설 반증 시 scope 재조정 blocker report).
- Step 2 (GREEN): isValidProfileName 강제 (배치는 design.md §D) → repro green.
- Step 3: CRUD routes (~2 POST) + Templ fragments + delete guards + i18n ×4.
- AC 바인딩: AC-WC11-030a/030b, 031..034.

### M5 — SPEC READ-ONLY Board (Priority: Medium; 독립)

- `spec.ListDocs` export + `Tier` frontmatter 필드 추가 → 보드 핸들러/Templ (status 분포 + close-debt 열 + MUST-FIX badge + copyable remediation).
- DetectDrift 제외 / 명령 실행 금지 / 쓰기 경로 부재 guard test.
- AC 바인딩: AC-WC11-040..046.

### M6 — Statusline cache_hit delta (Priority: Medium; B1 landing 확인 후)

- cache_hit 노출 fan-out (~10-12 파일): schema.go / preset.go / sync.go / profile_setup.go(TUI 목록 + stale comment) / statusline.yaml live+mirror / i18n.js / fieldsets.templ 파생 라벨 / statusline_test.go(want 15→16).
- segment-list SSOT set-equality 테스트 신설.
- `make build` (템플릿 재임베드).
- AC 바인딩: AC-WC11-050..054.

## §G Anti-Patterns

- AP-1: no-Save-path 섹션에 naive typed save → 주석/미모델링 키 파괴 (첫 쓰기에서 파일 손상).
- AP-2: i18n/라벨 blind sed 일괄 치환 — 파일별 판단 의무.
- AP-3: 보드 핸들러에서 remediation 명령 서버측 실행.
- AP-4: preset 셀렉터 재도입.
- AP-5: repro test 없이 M4 검증 코드 선반영 (Reproduction-First 위반).
- AP-6: renderer.go / cache_hit_test.go 접촉 (병렬 세션 파일).
- AP-7: RoleProfileEntry에 Effort 필드 추가 (REQ-WEM-006 reversal).
- AP-8: 템플릿 미러에 SPEC ID/내부 날짜 주석 삽입 (§25 neutrality 위반 — CI guard fail).
- AP-9 (v0.2.0): agent 파일 body 접촉 또는 frontmatter의 template dual-write (REQ-WC11-027 위반).
- AP-10 (v0.2.0): 스키마 총 필드 수 하드코딩 assertion (B11 위반 — 파생 방식만).
- AP-11 (v0.2.0): workflow_agents를 typed Save로 기록 (workflow.yaml 주석/team.patterns 파괴 — seam 경유만).

## §H Cross-References

- spec.md §2 (GEARS requirements — v0.2.1: 49 REQ) / §3 (Exclusions 7 h3).
- acceptance.md §D (AC Matrix — 검증 SSOT).
- design.md §A (yaml.Node seam + upsert) / §B (effort opaque-node 결정) / §C.1 (frontmatter patch layer + template-mirror policy) / §C.2 (workflow_agents) / §D (검증 배치).
- research.md §A-§H (survey) + §I (v0.2.0 결정 변경 기록 + plan-phase 추가 실측).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier L 5-section 위임 템플릿 의무).
- CLAUDE.local.md §2 (Template-First) / §25 (Template neutrality).
