---
id: SPEC-WEB-CONSOLE-011
status: draft
created: 2026-07-03
updated: 2026-07-03
---

# Design — SPEC-WEB-CONSOLE-011 (v0.2.1)

## §A yaml.Node Comment-Preserving Patch Seam (M1 핵심)

### §A.1 문제

`ConfigManager.Save()`(internal/config/manager.go:166, 207-219)는 typed struct 재직렬화로 6개 섹션 파일만 기록한다. struct 재직렬화는 (i) yaml 주석 전량 소실, (ii) 미모델링 키 소실을 일으킨다. v0.2.0 전면 확장으로 **Save() 경로가 없는 8개 섹션(workflow, harness, ralph, research, feedback, observability, security, db) 전부**가 웹 편집 대상이 되며, 이들 전부에 대해 node-level 부분 패치 seam이 load-bearing이다 (REQ-WC11-017). 특히 workflow.yaml은 `team.patterns` 의도적 미모델링(EXCL-WSE-004) + role profile `effort` Go-invisible(REQ-WEM-006)의 이중 특이점을 갖는다.

### §A.2 API 형상 (제안 — run-phase에서 확정; v0.2.0 upsert 확장)

배치: `internal/settings/yamlpatch` (신규 서브패키지; `internal/settings`는 010에서 확립된 중립 패키지로 cli/web 양쪽에서 import 가능).

```go
// KeyEdit은 yaml 문서 내 단일 스칼라 교체/생성을 기술한다.
type KeyEdit struct {
    Path  []string // 예: ["team", "role_profiles", "implementer", "model"]
    Value string   // 기록할 스칼라 값 (기존 노드의 Style/Tag 보존)
}

// PatchFile은 파일을 yaml.Node 문서로 로드, Path 탐색으로 대상 스칼라를 교체하고,
// (v0.2.0) 경로가 없으면 mapping 노드를 생성(upsert)한 뒤 재인코딩한다.
// 주석(Head/Line/FootComment)·키 순서·미모델링 키 보존. 삭제는 미지원.
func PatchFile(path string, edits []KeyEdit) error
```

원칙:
- **스칼라 교체 + 누락 경로 upsert** (v0.2.0 확장 — `workflow_agents` 최초 기록이 요구; 2026-07-03 grep 0으로 블록 부재 실측). upsert는 **명시적으로 편집된 경로에만** 적용 — 편집되지 않은 부재 키(예: role profile의 effort 부재)에 빈 값을 주입하지 않는다 (EC-3).
- 노드 **삭제는 seam 미지원** — 유일한 삭제 수요(frontmatter effort-key 제거)는 §C.1의 frontmatter patch layer가 별도 담당.
- 기존 노드의 `Style`/`Tag` 보존 (인용 스타일 유지 시도).
- Encoder indent는 대상 파일의 기존 들여쓰기와 일치하도록 설정 (섹션 yaml은 4-space).

### §A.3 섹션별 쓰기 라우팅 표 (v0.2.0 — 10섹션 전면)

| 섹션 | 쓰기 경로 | 근거 |
|------|----------|------|
| user / language / quality / git-convention | 기존 typed 경로 유지 (변경 없음) | 010 확립 |
| git-strategy | typed Save **dirty-flag** 경로 (manager.go:207-216) | SPEC-GITSTRATEGY-SAVE-ISOLATION-001; 완전 typed라 재직렬화 안전 |
| llm (안전 키) | typed 경로 (LLMConfig 완전 typed, oneof 검증) | types.go:234-283 |
| llm.mode / llm.team_mode | **쓰기 없음** (read-only) | runtime-managed 레이스 |
| workflow (`team.role_profiles` + `workflow_agents`) | **yamlpatch seam 전용** (upsert 포함) | 주석+patterns+effort 보존; REQ-WC11-005/073 |
| harness (스칼라 키) | **yamlpatch seam 전용**; `levels` map 내부는 read-only/collapsed (REQ-WC11-062) | Save() 경로 부재 |
| ralph / research / feedback / observability / security | **yamlpatch seam 전용** | Save() 경로 부재 (run-phase pre-flight §C-8에서 기계 재검증) |
| db | **yamlpatch seam 전용**; `orm`/`multi_tenant`/`migration_tool` 3키만 편집, 5 system 키 read-only (REQ-WC11-019) | Save() 경로 부재 + runtime 키 분리 |
| state / system / project / cache / sunset / tool-policy / lsp / mx / 미지명(constitution, context, design, interview) | **쓰기 없음 + 비노출** | REQ-WC11-018 |
| statusline | 기존 sync.go:103 직접 marshal 경로 유지 + cache_hit 키 추가 | M6은 노출만; 경로 재설계는 범위 밖 |

### §A.4 yaml.v3 정규화 리스크 (정직한 한계)

yaml.v3 Encoder는 노드 트리 재직렬화 시 일부 포매팅(따옴표 스타일, 빈 줄, 들여쓰기)을 정규화할 수 있다. 완전한 byte-stability는 보증 대상이 아니라 **검증 대상**이다:
- Golden-file round-trip 테스트를 **8개 seam 섹션 각각의 실제 파일 사본**으로 작성 (AC-WC11-017) — 편집 라인 외 diff가 발생하면 테스트가 그 사실을 드러낸다.
- 허용 기준: 주석 전량 보존 + 미모델링 키 전량 보존 + 키 순서 보존. 공백-only 정규화가 관측되면 그 범위를 golden에 명시적으로 고정하고 진행; 주석/키 소실 수준이면 blocker report (라인 단위 수동 패치로의 전환 결정은 orchestrator/사용자 몫).

## §B Team-Profile `effort` — Opaque-Node 결정 (v0.2.0 전면 쓰기 하에서도 유지)

두 옵션 비교:

| | Option A — opaque node (채택) | Option B — RoleProfileEntry에 Effort 추가 |
|---|---|---|
| REQ-WEM-006 | 존중 (reversal 없음) | **reversal — WORKFLOW-EFFORT-MAP-001 SPEC amendment 필요** |
| 코드 변경 | types.go 무변경; seam이 `effort`를 다른 키와 동일하게 노드로 패치 | types.go + 소비 코드 변경 |
| team.patterns 문제 | seam이 어차피 필요 (Option B로도 typed save 불가) | 동일하게 seam 필요 — Effort 추가의 실익 없음 |
| 검증 | 핸들러 계층에서 v4manifest closed set(schema.go:41-73)으로 제출값 검증 | struct 태그 검증 가능하나 중복 |

**결정: Option A 유지.** v0.2.0의 전면 쓰기 요구는 이 결정을 흔들지 않는다 — full write에서도 편집 단위는 여전히 "스칼라 키 패치"이고, seam은 effort를 Go 타입 인지 없이 노드로 읽고/쓰기 때문에 읽기(현재값 표시)·쓰기(패치)·검증(핸들러 closed set) 전부가 opaque-node로 완결된다. Option B가 필요해지는 조건(Go 코드 다른 지점이 effort를 타입-안전하게 소비해야 함)은 본 SPEC 범위에 존재하지 않는다 (REQ-WC11-023).

## §C Agent Settings 뷰 구성 (M3 — v0.2.0 4표면 전면 쓰기)

- 단일 페이지, 4개 카드(표면별) — 전부 편집 가능, 카드별 persistence 경로 배지 표시:
  1. **llm.yaml tiers** — 편집 (typed 경로). 빈 값 = "(runtime default)" (EmptyLabelKey).
  2. **team role profiles** — 7 profiles × {description(ro), model(edit), effort(edit, opaque), isolation(edit), mode(edit)} + `default_model`(edit) + `role_profile_keys`(ro). seam 경유.
  3. **sub-agent frontmatter** — 7 agents 표 (builder-harness inherit/high, manager-develop inherit/xhigh, manager-spec inherit/xhigh, plan-auditor inherit/xhigh, sync-auditor inherit/xhigh, manager-docs haiku(effort 부재), manager-git haiku(effort 부재)). **편집 가능** (§C.1 patch layer) + 지속 경고 배너.
  4. **workflow_agents** — 7 purposes × {model(edit), effort(edit)} (§C.2) + dynamic-workflows.md taxonomy 링크.
- 데이터 소스: (1)(2)(4)는 파일 로드((4)는 typed 읽기); (3)은 `.claude/agents/moai/*.md` frontmatter 파싱 (§C.1 layer의 read 측).
- quality/M2 필드와 동일한 FieldDef 패턴을 재사용하되, role_profiles/workflow_agents/frontmatter처럼 동적·파일단위 구조는 전용 view model 허용 (스칼라 필드 SSOT를 억지로 확장하지 않음).

### §C.1 Frontmatter Patch Layer (v0.2.0 신설 — REQ-WC11-027..029)

**파일 모델**: agent 파일 = `---\n<frontmatter yaml>\n---\n<body>`. Layer는 첫 두 `---` 구분선을 위치 식별 → frontmatter 구간만 yaml.Node 파싱 → 패치 → **원본 body bytes를 그대로 이어붙여** 재조립. body는 파싱조차 하지 않는다 (byte 보존이 구조적으로 보장).

**연산 집합** (섹션 seam과 다름): 스칼라 교체 + upsert + **`effort` 키 삭제** (EC-7: "(absent)"로 되돌리기 지원 — manager-docs/manager-git 선례상 effort 부재가 유효 상태이므로 삭제 연산이 필수). 섹션 seam(§A.2)은 삭제 미지원으로 유지 — 삭제 수요는 이 layer에만 존재.

**검증**: 핸들러 계층에서 v4manifest closed set — model ∈ {inherit, haiku, sonnet, opus}, effort ∈ {low, medium, high, xhigh, max} ∪ {absent}. out-of-set은 4xx.

**Idempotency**: 동일 패치 2회 적용 → byte-identical 파일 (AC-WC11-027의 기계 검증 대상).

**Template-mirror policy (결정)**: **live 파일만 기록 + 지속 경고, 자동 dual-write 없음.**
- 근거: `moai web`은 **사용자 프로젝트**에서 실행되며, 사용자 프로젝트에는 `internal/template/templates` 미러가 존재하지 않는다 (그 미러는 moai-adk-go dev repo 전용). dual-write를 구현하면 사용자 환경에서 dead code이고, dev repo에서는 §25 neutrality·Template-First 판단을 기계가 대신하게 되어 위험하다.
- 결과: 편집 UI에 "이 agent 파일은 template-managed — `moai update`가 편집을 덮어쓸 수 있음" 지속 경고 (REQ-WC11-028, i18n ×4). dev repo 메인테이너의 미러 정합은 수동 판단 (Out of Scope).

**견고성**: frontmatter 파싱 실패 파일은 해당 행만 "unavailable" + 편집 비활성 — 페이지 전체 실패 금지.

### §C.2 workflow_agents Typed Surface (v0.2.0 신설 — REQ-WC11-070..074)

**블록 형상** (workflow.yaml — 키 이름은 run-phase §C-9에서 dynamic-workflows.md L82-103의 7-purpose 명칭 실측 후 확정):

```yaml
workflow_agents:
    <purpose-1>: { model: haiku, effort: low }
    # ... 7 purposes
```

**Go 타이핑** (internal/config — REQ-WC11-071):

```go
// WorkflowAgentEntry는 dynamic-workflow purpose별 model/effort 기본값.
type WorkflowAgentEntry struct {
    Model  string `yaml:"model"`
    Effort string `yaml:"effort"`
}
// WorkflowConfig에 추가: WorkflowAgents map[string]WorkflowAgentEntry `yaml:"workflow_agents"`
```

**읽기/쓰기 비대칭**: 읽기는 typed loader (블록 부재 시 zero-value, 무오류); 쓰기는 seam upsert (REQ-WC11-073) — workflow.yaml의 주석/`team.patterns` 보존을 위해 typed Save는 금지 (AP-11).

**SSOT 시맨틱스**: config 블록 = dynamic-workflow model/effort **기본값의 SSOT**; per-script 리터럴 = **override** (우선). dynamic-workflows.md (live + template mirror — 양쪽 실존 2026-07-03 확인)를 이 시맨틱스로 갱신 — 별도 work item, Template-First(`make build`) + §25 neutrality (REQ-WC11-074). 템플릿 workflow.yaml mirror에도 기본 블록 추가.

## §D Profile CRUD + 검증 배치 (M4)

### §D.1 검증 배치 (defense in depth)

- **1차 (필수)**: 웹 경계 — `?profile=` / `__profile` 수신 지점(app.go:133-141, handlers.go:252-254)에서 `isValidProfileName` 즉시 검증, 실패 시 4xx.
- **2차 (권장)**: `profile.GetPreferencesPath` / `WritePreferences` 내부 검증 — 웹 외 호출자도 보호. 단, 기존 호출자 회귀 여부를 전수 확인 후 적용 (2차가 회귀를 만들면 1차만으로 GREEN 처리하고 2차는 별도 커밋/후속 판단).
- repro test는 1차/2차 어느 배치로든 green이 되도록 행동(behavior) 기준으로 작성: "불법 이름 → 4xx + FS 무변화".

### §D.2 라우트

- `POST /profile/create` — name 검증 → `EnsureDir` → 갱신된 목록 fragment 반환 (HTMX).
- `POST /profile/delete` — name 검증 + delete guards(v0.2.1 D2 전제 정정: `default` 거부 = 기존 guard 재사용; **active-profile 거부(4xx) = 웹 경계 신설 로직** — live `profile.Delete`는 active를 stderr 경고 후 삭제 진행, profile.go:98-105 RemoveAll) → 목록 fragment.
- switch — 기존 `?profile=` 로드 경로 재사용 (필요시 `POST /profile/switch`로 명시화; run-phase 판단).
- 모든 fragment 문자열 i18n ×4.

## §E SPEC Board 데이터 흐름 (M5)

```
GET /specs → handler
  ├─ spec.ListDocs(baseDir)   ← 신규 exported wrapper (discoverSPECs/parseSPECDoc)
  ├─ spec.Audit(...)          ← pure FS scan (0.14s/412 — 동기 렌더 허용)
  └─ view model 조립:
       {ID, Status, Tier?, Updated, IsCloseDebt(=implemented), DriftFindings[]}
→ Templ 렌더: status 분포 요약 + close-debt 열 + MUST-FIX badge
→ remediation 문자열은 <code> + copy 버튼 (기존 static assets clipboard 패턴)
```

- `DetectDrift`(git 의존, 7.9s)는 호출 금지 — import 자체를 피하고 grep guard로 고정 (AC-WC11-045).
- 쓰기 라우트 없음 — 라우트 등록 guard test (AC-WC11-046).
- `Tier string yaml:"tier"` 필드는 optional — zero value면 badge 생략 (EC-5).
- 캐싱 없음 (0.14s 실측이면 동기 렌더로 충분 — 단순성 우선; 성능 회귀 시 후속).

## §F Statusline Segment-List SSOT 테스트 (M6)

- **SSOT 선정**: `statusline.CanonicalSegments`(preset.go — 세그먼트의 owning package).
- set-equality 테스트가 다음 목록들을 SSOT와 집합 비교:
  1. `settings.statuslineSegmentKeys` (schema.go)
  2. profile `defaultStatuslineSegments` (sync.go)
  3. TUI `statuslineAllSegments` (profile_setup.go)
- 테스트 배치: import 방향 안전한 패키지 (internal/settings 테스트가 statusline/profile을 import — 순환 없음 확인은 run-phase; 필요시 최상위 통합 테스트 패키지).
- 이 테스트가 향후 "renderer에는 있는데 노출 목록에 없는" 3-way orphan 재발을 컴파일 타임+CI에서 차단한다 (REQ-WC11-051).

## §G Risks (v0.2.0 갱신)

| 리스크 | 완화 |
|--------|------|
| yaml.v3 정규화 노이즈로 byte-stability 미달 (8섹션으로 표면 확대) | §A.4 per-section golden ×8 + 허용 기준 명문화 + blocker 경로 |
| 병렬 세션 statusline 작업과 충돌 | plan.md §C pre-flight B1 landing 확인; renderer.go/cache_hit_test.go 무접촉 |
| survey 라인 앵커 드리프트 | content-token grep 재검증 후 편집 (plan.md §C-5) |
| guard-test supersede가 의도보다 넓게 열림 | AC-WC11-002에 제외군(8+4 섹션) 거부 케이스 필수 포함 |
| M4 가설 반증 (검증 갭이 실존하지 않음) | repro test가 반증하면 blocker report → scope 재조정 (검증 갭 수정 항목 제거, CRUD만 진행) |
| frontmatter 파싱(agent 7종) 취약성 | §C.1 견고성 — 실패 행 "unavailable", 페이지 전체 실패 금지 |
| i18n 인플레이션 (신규 키 수십 개 ×4) | B10 — 키 규약 일관성 + parity 전수 테스트; blind sed 금지 |
| 섹션 키/typed struct 미실측 (ralph 등 6종) | plan.md §C-7/8 pre-flight 기계 열거 후 FieldDef 확정; 불일치 발견 시 blocker |
| harness `levels` map 복잡도 | REQ-WC11-062 collapsed raw view — 폼 강제하지 않음 |
| 스키마 총 필드 수 유동으로 카운트 assertion 취약 | B11 파생 assertion 전환 (AC-WC11-053) |

## §H Cross-References

- spec.md §2 (REQ-WC11-001..074), §3 (Exclusions).
- research.md §B (persistence 4-seam 실측), §C-§D (agent 4표면·config 실측), §G (statusline orphan 실측), §I (v0.2.0 결정 변경 + plan-phase 추가 실측).
- `internal/harness/v4manifest/schema.go:41-73` — effort/model closed sets (REQ-WC11-024/029/072 재사용 대상).
- SPEC-WEB-CONSOLE-010 design.md — FieldDef/PersistTarget/EmptyLabelKey 패턴 선례.
