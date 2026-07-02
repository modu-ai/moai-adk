---
id: SPEC-WEB-CONSOLE-011
status: draft
created: 2026-07-03
updated: 2026-07-03
---

# Acceptance — SPEC-WEB-CONSOLE-011 (v0.2.1)

> 검증 SSOT. 모든 AC는 가능한 한 기계 검증(go test 타깃 / grep assertion / golden diff). 명령의 파일:라인 앵커는 run-phase에서 content-token 기준으로 재확인 후 사용.
> Severity: **[B]** = blocking (DoD 필수), **[N]** = non-blocking (권장).
> ID 매핑 주의: REQ와 AC는 별개 네임스페이스 — REQ-WC11-062(map-of-structs)의 검증은 AC-WC11-063이다 (AC-WC11-062는 subagent boundary grep에 선점).

## §D AC Matrix

### M1 — Foundation

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-001 | [B] | `grep -rn "SPEC-WEB-CONSOLE-011" internal/web/server.go internal/web/projectconfig.go` | 구 scope contract 주석이 **10섹션 계약**으로 supersede 명기 ≥ 2 매치 |
| AC-WC11-002 | [B] | (v0.2.1 D4 정정 — 구 패턴은 실존 함수 무매치로 `-run` exit 0 vacuous pass) `go test -v -run 'TestWriteProjectConfigSectionIsolation\|TestServer_\|TestScopeContractTenSections\|TestScopeContractExclusions' ./internal/web/` — 실존 테스트(TestWriteProjectConfigSectionIsolation, TestServer_*) + 본 SPEC **신설** guard test 함수명 `TestScopeContractTenSections` / `TestScopeContractExclusions` 명시 바인딩; `-v` 출력의 `=== RUN`으로 실제 실행 확인 | 신설 guard 2종 포함 전부 PASS: 10섹션(git-strategy/llm/workflow/harness/ralph/research/feedback/observability/security/db) 허용 케이스 + 제외군(state/system/project/cache/sunset/tool-policy/lsp/mx/미지명) 거부 케이스 포함 |
| AC-WC11-003 | [B] | `go test -run TestYAMLPatch ./internal/settings/...` (골든 round-trip) | 실제 workflow.yaml fixture: 대상 키만 변경, 주석/`team.patterns`/키 순서 보존 — golden diff가 편집 라인에 국한; upsert 케이스(누락 경로 생성) 포함 |
| AC-WC11-004 | [B] | seam 라우팅 테스트: workflow.yaml 쓰기 경로가 seam 함수 경유 | PASS (unit test로 호출 경로 고정) |
| AC-WC11-005 | [B] | (v0.2.1 D3 정정 — 구 grep은 vacuous: 실제 호출 형상은 `mgr := config.NewConfigManager(); mgr.Save()` @ projectconfig.go:217/244이며 `\|db` 서브스트링은 과매치) `grep -nE 'NewConfigManager\|\.Save\(' internal/web/` + AC-WC11-004 기반 행동 테스트(workflow.yaml 쓰기가 seam으로만 라우팅됨을 고정) | grep 매치가 **명시 allowlist**(projectconfig.go의 승인된 quality/git-convention typed 경로 — projectconfig.go:217/244)에 국한, allowlist 외 신규 `NewConfigManager`/`.Save(` 호출 0; workflow.yaml은 해당 typed 경로로 절대 라우팅되지 않음 (행동 테스트 PASS) |

### M2 — Config 전면 확장 (10섹션)

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-010 | [B] | `go test -run TestGitStrategyFields ./internal/web/ ./internal/settings/` | git-strategy 전체 필드 렌더+저장 round-trip PASS (dirty-flag 경로) |
| AC-WC11-011 | [B] | quality 잔여 키 schema diff 테스트 | quality.yaml 키 전수 대비 미노출 잔여 = 0 (제외 목록 명시분 제외) |
| AC-WC11-012 | [B] | `go test -run TestLLMSafeKeys ./internal/web/` | llm 안전 키 렌더+저장 PASS; oneof 검증 위반 값은 4xx |
| AC-WC11-013 | [B] | `curl POST /save`에 `llm.mode` 변경 포함 시도 (httptest) | mode/team_mode는 무시 또는 4xx — 파일 값 불변 (read-only 계약) |
| AC-WC11-014 | [N] | 빈 tier 렌더 스냅샷 | "(runtime default)" EmptyLabelKey 라벨 렌더 |
| AC-WC11-015 | [B] | i18n.js 키 존재 검사: 신규 필드 키 × 4 locale | en/ko/ja/zh 전부 존재 (누락 0) |
| AC-WC11-016 | [B] | 10섹션 스모크 httptest: 각 섹션 최소 1개 스칼라 필드 렌더 + 저장 round-trip | 10/10 섹션 PASS (git-strategy, llm, workflow, harness, ralph, research, feedback, observability, security, db) |
| AC-WC11-017 | [B] | `go test -run TestYAMLPatchGolden ./...` — **8개 seam 섹션 각각**의 golden fixture round-trip (workflow, harness, ralph, research, feedback, observability, security, db) | 섹션별 8/8 PASS: 주석 + unknown key + 키 순서 보존, diff는 편집 라인 국한 |
| AC-WC11-018 | [B] | 제외군 guard test + UI 부재 검사: excluded 섹션 폼 렌더 0 + write 시도 4xx/무시 | state/system/project/cache/sunset/tool-policy/lsp/mx + constitution/context/design/interview 전부 비노출·쓰기 거부 |
| AC-WC11-019 | [B] | db round-trip 테스트: `orm`/`multi_tenant`/`migration_tool` 편집 반영, 5개 system 키 write 시도 → 파일 불변 | 3키 editable / 5키 read-only PASS |

### M3 — Agent Settings (4표면 전면 쓰기)

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-020 | [B] | httptest GET agent-settings 페이지 | 4표면 전부 렌더: llm tiers + role_profiles(7) + sub-agent frontmatter(7) + workflow_agents(7 purposes) |
| AC-WC11-021 | [B] | llm tier 편집 round-trip 테스트 | 저장 반영 PASS |
| AC-WC11-022 | [B] | role_profile model 편집 → workflow.yaml diff 검사 | 대상 스칼라만 변경; 주석/`team.patterns` 보존 (seam 경유 증거) |
| AC-WC11-023 | [B] | (v0.2.1 D6 정정 — 전 파일 grep은 신설 `WorkflowAgentEntry{Effort string}`(REQ-WC11-071, design.md §C.2)와 자기모순) `awk '/type RoleProfileEntry struct/,/^}/' internal/config/types.go \| grep -c "Effort"` 또는 reflection 컴파일 테스트 `TestRoleProfileEntryHasNoEffortField` | **RoleProfileEntry struct 블록 내** Effort 필드 0 — 신설 WorkflowAgentEntry의 Effort 필드는 본 assertion 범위 밖 |
| AC-WC11-024 | [B] | 잘못된 effort 값(`superhigh`) 제출 테스트 | 4xx; v4manifest closed set 참조 확인 (`grep -rn "v4manifest" internal/web/` ≥ 1) |
| AC-WC11-025 | [B] | (v0.2.0 개정) frontmatter 편집 round-trip httptest: agent 1종 model/effort 변경 → 파일 반영 | 편집 PASS + 렌더에 현재 값 반영 |
| AC-WC11-026 | [B] | (v0.2.0 개정) dynamic-workflow 표면 편집 → workflow.yaml `workflow_agents` 블록 반영 + taxonomy 링크 잔존 | PASS |
| AC-WC11-027 | [B] | frontmatter patch **idempotency + body 보존** 테스트: 동일 패치 2회 → byte-identical; body 구간 원본과 `bytes.Equal` | PASS (body 무접촉 기계 증거) |
| AC-WC11-028 | [B] | agent-settings Templ 렌더에 지속 경고 존재 + i18n ×4 | "moai update가 덮어쓸 수 있음" 경고 렌더 (4-locale 키 존재) |
| AC-WC11-029 | [B] | (i) model=`gpt5` / effort=`superhigh` 제출 → 4xx; (ii) effort 키 부재 agent(manager-docs) 저장 → 키 미주입; (iii) `grep -rn "internal/template/templates" internal/web/` (frontmatter write 경로) | (i) 거부 (ii) 부재 보존 (iii) 0 매치 (dual-write 부재) |

### M3 — workflow_agents 신규 typed 표면

> ID 매핑 주의(v0.2.1 D9d, §2.3.1 블록): AC↔REQ 번호가 1 어긋난다 — AC-WC11-070→REQ-WC11-070/071, AC-071→REQ-072, AC-072→REQ-073, AC-073/074→REQ-074 (문서 서두의 062/063 주의문과 동일 취지).

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-070 | [B] | `go test -run TestWorkflowAgents ./internal/config/` — typed loader round-trip | 7-purpose map 파싱 PASS; 블록 부재 시 zero-value 무오류 |
| AC-WC11-071 | [B] | **enum reject test**: workflow_agents에 model=`gpt5` / effort=`ultra` 제출 | 4xx + 파일 불변 (v4manifest closed set) |
| AC-WC11-072 | [B] | workflow.yaml golden: workflow_agents upsert 최초 기록 → 주석/`team.patterns`/기존 키 보존 | golden diff가 신규 블록 추가에 국한 |
| AC-WC11-073 | [B] | `grep -c "workflow_agents" .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` | live + mirror 각각 ≥ 1 (SSOT 참조 갱신); mirror에 SPEC ID 토큰 0 (§25) |
| AC-WC11-074 | [B] | `grep -c "workflow_agents" internal/template/templates/.moai/config/sections/workflow.yaml` + `make build` exit 0 | 템플릿 mirror에 기본 블록 존재 + 재임베드 성공 |

### M4 — Profile CRUD (repro-test-first)

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-030a | [B] | repro test가 수정 이전 트리에서 FAIL함을 커밋 순서로 증거 (RED 커밋 → GREEN 커밋) | `git log` 상 repro test 커밋이 fix 커밋에 선행; RED 시점 출력 인용 |
| AC-WC11-030b | [B] | `go test -run TestProfileNameTraversal ./internal/web/ ./internal/profile/` | 수정 후 PASS: `__profile=../../x` → 4xx + 디렉터리 미생성 |
| AC-WC11-031 | [B] | `grep -rn "isValidProfileName" internal/web/ internal/profile/preferences.go` | 웹 쓰기 경로에서 검증 호출 ≥ 1 |
| AC-WC11-032 | [B] | CRUD httptest: create → list → switch → delete | 전 흐름 PASS; i18n 키 4-locale 존재 |
| AC-WC11-033 | [B] | `default` 및 active profile delete 시도 테스트 | 거부 (4xx) + 프로파일 잔존 |
| AC-WC11-034 | [B] | 불법 이름(`../x`, 공백, 빈 문자열) 제출 테스트 | 4xx + `MkdirAll` side effect 0 (임시 HOME에서 FS 검사) |

### M5 — SPEC Board

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-040 | [B] | httptest GET 보드 (fixture SPEC 디렉터리) | status 분포 + close-debt 열(implemented-not-completed) + MUST-FIX badge 렌더 |
| AC-WC11-041 | [B] | `go doc ./internal/spec ListDocs` | exported wrapper 존재; unit test PASS |
| AC-WC11-042 | [N] | Tier 필드 파싱 테스트 (tier 있는/없는 fixture) | 있으면 badge, 없으면 생략 — 오류 0 |
| AC-WC11-043 | [B] | 보드 렌더 출력에 remediation 문자열 존재 검사 | copyable text로 렌더 (예: `moai spec close <ID> --backfill-only`) |
| AC-WC11-044 | [B] | `grep -rn "exec.Command\|os/exec" internal/web/` (보드 핸들러 범위) | baseline 대비 신규 0 |
| AC-WC11-045 | [B] | `grep -rn "DetectDrift" internal/web/` | 0 매치 |
| AC-WC11-046 | [B] | 보드 라우트 등록 검사 + guard test | specs 경로에 GET만 존재; POST/PUT/DELETE 0 |

### M6 — Statusline cache_hit

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-050 | [B] | `for f in internal/settings/schema.go internal/statusline/preset.go internal/profile/sync.go internal/cli/profile_setup.go internal/web/assets/i18n.js .moai/config/sections/statusline.yaml internal/template/templates/.moai/config/sections/statusline.yaml; do grep -c "cache_hit" $f; done` | 각 파일 ≥ 1 |
| AC-WC11-051 | [B] | `go test -run TestSegmentListSSOT ./...` (set-equality) | 3+ 하드코딩 목록 == SSOT 집합 — PASS |
| AC-WC11-052 | [B] | live vs 템플릿 미러 diff: 양쪽 `grep cache_hit` | 동일 키 존재 (Template-First); 미러에 SPEC ID 주석 0 |
| AC-WC11-053 | [B] | (v0.2.0 개정) 파생 카운트 테스트: fieldsets 카운트 라벨이 schema 길이에서 파생됨을 검증 + `grep -nE '"[0-9]+ fields"' internal/web/fieldsets.templ` 하드코딩 총계 리터럴 0 + `go test -run TestStatusline ./internal/web/` (statusline-국소 want 16) | 파생 라벨 PASS + 하드코딩 총계 0 + segment want 16 PASS |
| AC-WC11-054 | [N] | `grep -n "11-segment" internal/cli/profile_setup.go` | 0 매치 (stale comment 정정) |

### Cross-cutting

| AC | Sev | 검증 | 기대 결과 |
|----|-----|------|----------|
| AC-WC11-060 | [B] | (v0.2.1 D5 정정 — app.go:84 기존 주석이 패턴에 매치되어 expected-0 불가였음) `grep -rn "csrf\|CSRF\|xsrf" internal/web/ \| grep -v "_test.go" \| grep -v "REQ-WC-009" \| grep -vE ':[0-9]+:[[:space:]]*//'` | 비-주석 코드 매치 0 (기존 주석 라인은 필터로 제외); 신규 CSRF 인프라 0; @MX:NOTE(app.go) 보존 |
| AC-WC11-061 | [B] | i18n 4-locale parity 테스트 (신규 키 전수) | en/ko/ja/zh 누락 0 |
| AC-WC11-062 | [B] | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/web/ internal/spec/ \| grep -v "_test.go" \| grep -v "^[^:]*:[0-9]*:[ \t]*//"` | 0 매치 (subagent boundary) |
| AC-WC11-063 | [B] | (REQ-WC11-062 검증) harness `levels` 내부 / workflow `team.patterns` 렌더 검사 | read-only 또는 collapsed raw view — 폼 input 컨트롤 0; 같은 섹션 스칼라 키는 폼 필드 존재 |

## Given-When-Then 시나리오

### GWT-1 — Profile path traversal 차단 (M4)

- **Given** 웹 콘솔이 loopback에서 실행 중이고 임시 HOME 하에 profile store가 초기화됨,
- **When** `POST /save`(또는 profile route)에 `__profile=../../evil` 이 제출되면,
- **Then** 응답은 4xx이고, `~/.moai/claude-profiles/` 밖에 어떤 디렉터리/파일도 생성되지 않는다.

### GWT-2 — workflow.yaml 주석 보존 편집 (M1+M3)

- **Given** 주석과 `team.patterns`, per-profile `effort` 키를 포함한 실제 workflow.yaml fixture,
- **When** agent-settings 페이지에서 role profile 하나의 `model` 값만 변경 저장하면,
- **Then** 파일 diff는 해당 스칼라 라인에 국한되고 주석/`team.patterns`/`effort`/키 순서는 보존된다.

### GWT-3 — SPEC 보드 동기 렌더 (M5)

- **Given** 412개 SPEC 규모의 `.moai/specs/` fixture,
- **When** 보드 페이지를 GET 하면,
- **Then** pure-FS Audit만으로 렌더되고(git 호출 0), close-debt 열과 MUST-FIX badge, copyable remediation 문자열이 표시되며, 어떤 쓰기 컨트롤도 존재하지 않는다.

### GWT-4 — llm.mode read-only (M2)

- **Given** `llm.yaml`의 `mode: cg` (runtime-managed 상태),
- **When** 폼 제출에 mode/team_mode 변경이 포함되면,
- **Then** 해당 두 키는 파일에서 불변이며 UI는 read-only 표시를 유지한다.

### GWT-5 — cache_hit 세그먼트 노출 (M6)

- **Given** 신규 init된 프로젝트,
- **When** statusline 섹션이 웹/TUI에서 렌더되면,
- **Then** cache_hit 포함 16개 세그먼트가 4-locale 라벨과 함께 표시되고, 저장 round-trip이 hand-edit 값을 보존한다.

### GWT-6 — agent frontmatter 편집 (M3, v0.2.0)

- **Given** `manager-develop.md` (frontmatter `model: inherit` / `effort: xhigh` + 본문 body),
- **When** agent-settings에서 effort를 `high`로 변경 저장하면,
- **Then** frontmatter만 갱신되고 body는 byte-identical이며, 동일 패치를 한 번 더 적용해도 파일이 byte-identical이다 (idempotency). 지속 경고("moai update가 덮어쓸 수 있음")가 표시된다.

### GWT-7 — workflow_agents 최초 기록 (M3, v0.2.0)

- **Given** `workflow_agents` 블록이 없는(2026-07-03 grep 0 실측) 주석 포함 workflow.yaml,
- **When** 특정 purpose의 model을 설정 저장하면,
- **Then** seam upsert가 블록을 생성하되 기존 주석/`team.patterns`는 보존되고, out-of-set 값(model=`gpt5`)은 4xx로 거부된다.

## Edge Cases

- **EC-1**: `default` profile delete 시도 → 거부 (AC-WC11-033).
- **EC-2**: 현재 active profile delete 시도 → 거부 (AC-WC11-033).
- **EC-3**: `effort` 키가 없는 role profile 편집 시 seam이 빈 effort를 주입하지 않는다 — 없는 키는 없는 채로 보존 (명시 편집된 경로만 upsert).
- **EC-4**: claude_models tier 빈 문자열 → "(runtime default)" 렌더, 저장 시 빈 문자열 유지 (다른 값으로 정규화 금지).
- **EC-5**: tier frontmatter 부재 SPEC (235/412) → 보드 badge 생략, 파싱 오류 0.
- **EC-6**: 사용자가 손으로 `cache_hit: false` 편집한 statusline.yaml → 웹 저장이 이 값을 덮어쓰지 않고 보존.
- **EC-7 (v0.2.0)**: effort 키 부재 agent(manager-docs / manager-git) — 편집 화면에서 "(absent)" 상태 표시, 저장 시 키 미주입; 사용자가 값을 지정하면 키 추가, "(absent)"로 되돌리면 키 제거 (frontmatter layer의 effort-key 삭제 지원, design.md §C.1).
- **EC-8 (v0.2.0)**: 제외군 섹션 이름을 위조한 POST(`section=tool-policy`) → 4xx + 파일 무변화 (AC-WC11-018).

## Quality Gate / Definition of Done

- [ ] 모든 [B] AC PASS (E1 matrix에 명령+실제 출력 인용 — verification-claim-integrity §3.2).
- [ ] `go test ./...` 전체 PASS + `go vet ./...` clean.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- [ ] `golangci-lint run` NEW issue 0 (baseline 구분 명시).
- [ ] 터치 패키지 커버리지 ≥ 85% (`go test -cover` 실제 출력 인용).
- [ ] i18n 4-locale parity (신규 키 누락 0 — B10 인플레이션 범위 전수).
- [ ] 템플릿 미러 parity (statusline.yaml + workflow.yaml + dynamic-workflows.md) + `make build` 재임베드 + neutrality CI guard green.
- [ ] agent 파일 body byte-무변경 (`git diff` 상 frontmatter 라인만).
- [ ] renderer.go / cache_hit_test.go 무접촉 (`git diff --name-only`에 부재).
- [ ] `moai spec lint` clean (Out of Scope h3 + frontmatter 12필드).
