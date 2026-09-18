# verdict — card t909: CLAUDE.md §4 에이전트 명단 불일치

`.claude/agents/moai/` · 템플릿 C2 · codex C3 세 사본이 모두 MoAI-custom 12개인데 CLAUDE.md §4 는
11개로 적고 `mission-governor` 가 명단에서 빠져 있다는 카드.

- 측정 시점: 2026-09-18
- 판정 트리: `.claude/worktrees/t909`, 브랜치 `WT-agent-roster-count`
- base: `2ede711ca` = `origin/develop` (fetch 후 `0 0`)
- 리드 판정: **2번 채택** — D1 + `manager-design.md` 쌍둥이를 이 카드로, D2·D3 는 별 카드

## 결론

**카드 전제는 성립한다.** `mission-governor` 는 의도적 제외가 아니라 누락이고, 12 MoAI-custom +
Explore = 13 이 정본이다. 다만 명단은 하나가 아니라 **다섯 곳**이었고 **셋이 서로 달랐다**.
이 카드는 그중 CLAUDE.md 쌍둥이와 `manager-design.md` 쌍둥이를 고쳤고, 코드 쪽 두 곳은 별 카드
초안으로 남긴다(§ 별 카드 초안).

---

## Claim

1. **`mission-governor` 는 retained 에이전트이며 §4 누락은 오버사이트다.** 도입 커밋이 사본 셋과
   `catalog.yaml`·`profile_matrix.go`·docs-site 4로케일·README 4종을 모두 배선하면서 CLAUDE.md 만
   건드리지 않았고, 저장소 자신의 테스트가 이미 12를 정본으로 못박고 있다.
2. **로스터는 다섯 곳에 있고 셋이 서로 다르다.** 파일 실체·`profile_matrix.go`·`catalog.yaml` 은
   12, `profile.go`는 manager-lead 누락, `delegationmap`과 CLAUDE.md 는 mission-governor 누락.
3. **D1 수리 완료** — CLAUDE.md 쌍둥이 2곳을 13/12 로 정정하고 `mission-governor` 를 명단에 넣었다.
4. **범위 확장 1건(리드 승인)** — `manager-design.md` 쌍둥이의 `CLAUDE.md § 4 (11 retained agents
   — manager-design is entry 11 …)` 는 **두 번 낡아** 있었다. 개수와 트리 위치 둘 다 고쳤다.
5. **기계 재생성 2건이 동반됐다** — C2 편집이 codex C3 방출과 `catalog.yaml` 해시를 함께 낡게
   만들었고, 둘 다 명시 동사로 재생성했다. 드리프트 검사는 재생성 전 FAIL → 후 ok 로 양방향 확인.
6. **이 숫자를 지키는 가드는 없다** — CLAUDE.md 프로즈 카운트를 고정하는 테스트가 존재하지 않아
   같은 자리가 다시 낡는다(§ Gaps 1).

## Evidence

### E1 — 의도적 제외인가: 아니다

```
$ git log --format='%h %ad %s' --date=short -- .claude/agents/moai/mission-governor.md
5ec516165 2026-09-15 feat(gtd): add autonomous GTD mission flow (t1)

$ git show --stat --format='' 5ec516165 -- CLAUDE.md internal/template/templates/CLAUDE.md
(무출력 — 사본 셋을 넣으면서 CLAUDE.md 는 양쪽 다 건드리지 않았다)

$ git show --stat --format='' 5ec516165 -- internal/template/templates/.claude/agents/moai/mission-governor.md internal/template/templates/.codex/agents/moai/mission-governor.toml
 .../.claude/agents/moai/mission-governor.md        | 42 ++++++++++++++++++++++
 .../.codex/agents/moai/mission-governor.toml       | 41 +++++++++++++++++++++

$ git ls-tree --name-only main -- .claude/agents/moai/mission-governor.md
(무출력 — main 에는 없다. develop 전용)
```

결정적 증거는 저장소 자신의 테스트다. 문서끼리 대조하면 어느 쪽이 정본인지 원리상 못 가르지만,
기계가 강제하는 값은 가른다.

```
$ sed -n '249,258p' internal/template/catalog_tier_audit_test.go
// 7 retained MoAI-custom agents (4 core + 3 meta) + 1 Anthropic built-in Explore
...
// (11th MoAI-custom retained agent — depth-1 Agent fan-out coordinator),
// plus mission-governor; net +2 = 12.
const expectedAgentCount = 12
```

`profile.go` 의 주석도 같은 방향이다 — *"Newly retained decision roles must be registered here
before an operator can target them with an override"* 아래에 `"mission-governor": true` 가 있다.
즉 그 목록은 mission-governor 를 retained 로 취급한다.

### E2 — 사본 셋 실측 (모두 12)

```
$ ls -1 .claude/agents/moai/*.md | wc -l                                          → 12
$ ls -1 internal/template/templates/.claude/agents/moai/*.md | wc -l              → 12
$ ls -1 internal/template/templates/.codex/agents/moai/*.toml | wc -l             → 12
$ grep -c "agents/moai/" internal/template/catalog.yaml                           → 12
```

세 사본의 파일명 집합은 동일하다: builder-harness, e2e-tester, manager-design, manager-develop,
manager-docs, manager-git, manager-lead, manager-spec, **mission-governor**, plan-auditor,
super-advisor, sync-auditor.

### E3 — 로스터 다섯 곳, 셋이 서로 다름

| 로스터 | 위치 | MoAI-custom | manager-lead | mission-governor |
|---|---|---|---|---|
| 파일 실체 | `.claude/agents/moai/` + C2 + C3 | **12** | ✓ | ✓ |
| 배포 매니페스트 | `internal/template/catalog.yaml` | **12** | ✓ | ✓ |
| 프로필 매트릭스 | `internal/template/profile_matrix.go` `profileMatrixAgentOrder` | **12** | ✓ | ✓ |
| override 검증 집합 | `internal/config/profile.go` `retainedAgentNames` | 11 | **✗** | ✓ |
| 위임맵 스냅샷 | `internal/harness/delegationmap/types.go` `retainedCatalog` | 11 | ✓ | **✗** |
| **문서 (이 카드 대상)** | `CLAUDE.md` §4 + 템플릿 쌍둥이 | 11 → **12** | ✓ | **✗ → ✓** |

```
$ grep -n "manager-lead" internal/config/profile.go
(무출력 — ABSENT)
$ grep -n "mission-governor" internal/harness/delegationmap/types.go
(무출력 — ABSENT)
$ sed -n '242p' internal/template/profile_matrix.go
	"manager-lead",
$ sed -n '240p' internal/template/profile_matrix.go
	"mission-governor",
```

### E4 — D1 수리 (CLAUDE.md 쌍둥이)

```
$ grep -n "retained agents" CLAUDE.md internal/template/templates/CLAUDE.md
CLAUDE.md:55:… exactly **13 retained agents** (12 MoAI-custom + 1 Anthropic built-in `Explore`) …
internal/template/templates/CLAUDE.md:55:… (동일 문안)

$ grep -n "^\*\*Retained agents" CLAUDE.md internal/template/templates/CLAUDE.md
CLAUDE.md:69:**Retained agents (13)**: … `manager-lead`, `mission-governor` (12 MoAI-custom) + Anthropic built-in `Explore`. `mission-governor` carries no Selection Decision Tree row by design — it is the GTD auto-mission decision role, dispatched by that workflow rather than selected by the orchestrator. …
internal/template/templates/CLAUDE.md:69:(동일 문안)
```

`mission-governor` 에 결정 트리 행을 **추가하지 않았고**, 그 부재가 결함으로 읽히지 않도록 한 절을
붙였다. 근거: 그 에이전트는 GTD 자동 미션 흐름의 구성요소로 기계 배차되며(`goal.md:147` —
*"The **mission-governor** may return only a bounded structured decision"*), 오케스트레이터가
선택하는 대상이 아니다. 트리는 오케스트레이터 선택용이므로 행이 없는 것이 맞다.

크기 예산(40,000자) 여유 확인:

```
$ wc -c CLAUDE.md internal/template/templates/CLAUDE.md
   19389 CLAUDE.md
   19451 internal/template/templates/CLAUDE.md
```

### E5 — 범위 확장 (리드 승인): manager-design 쌍둥이

두 번 낡아 있었다 — 개수(`11 retained agents`)와 트리 위치(`entry 11 in the Selection Decision
Tree`, 실제 트리는 7행이고 manager-design 은 6행).

```
$ grep -n "Agent catalog" .claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md
.claude/agents/moai/manager-design.md:208:- **Agent catalog**: `CLAUDE.md` § 4 (13 retained agents — manager-design is row 6 of the Selection Decision Tree).
internal/template/templates/.claude/agents/moai/manager-design.md:210:(동일 문안)
```

### E6 — 기계 재생성 2건 (양방향 확인)

C2 편집이 두 생성물을 함께 낡게 만들었다. 둘 다 재생성 **전 FAIL → 후 ok** 를 보였으므로, 통과가
죽은 검사의 침묵이 아니라 실제 판정이다.

```
$ make agents-emit-check          # 재생성 전
--- FAIL: TestGoldenCommittedArtifactsMatchEmission
    golden_test.go:109: .codex/agents/moai/manager-design.toml: committed artifact differs from
    emission (sha256 mismatch) — regenerate or stop hand-editing
agent-emit drift: … run `make agents-emit`

$ make agents-emit
AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
ok

$ make agents-emit-check          # 재생성 후 (대조)
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.350s
```

```
$ go test ./internal/template/...   # catalog 해시 재생성 전
--- FAIL: TestManifestHashFormat
    catalog_tier_audit_test.go:451: CATALOG_HASH_UNSTABLE: manager-design
    stored hash=3863fa6a…, computed hash=42025003… (source=.claude/agents/moai/manager-design.md)

$ go run ./internal/template/scripts/gen-catalog-hashes.go --all
catalog.yaml updated successfully (13145 bytes)
```

### E7 — 검증 (세정 6변수 단일 호출)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
        MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER && go test ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	56.495s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.420s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	(cached)
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]

$ unset … && go test ./internal/cli/agentlint/... ./internal/harness/delegationmap/...
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.381s
ok  	github.com/modu-ai/moai-adk/internal/harness/delegationmap	0.340s

$ go vet ./internal/template/...
(무출력)
```

변경 파일 6개 (손편집 4 + 기계 재생성 2):

```
$ git status --short
 M .claude/agents/moai/manager-design.md                                  ← 손편집
 M CLAUDE.md                                                              ← 손편집
 M internal/template/catalog.yaml                                         ← 기계 재생성
 M internal/template/templates/.claude/agents/moai/manager-design.md      ← 손편집
 M internal/template/templates/.codex/agents/moai/manager-design.toml     ← 기계 재생성
 M internal/template/templates/CLAUDE.md                                  ← 손편집
```

## Baseline-attribution

| 대상 | 값 | 확인 명령 |
|---|---|---|
| base / `origin/develop` | `2ede711ca` | `git rev-parse --short HEAD` · `git rev-parse --short origin/develop` (동일) |
| develop 동기 상태 | `0 0` | `git rev-list --count --left-right origin/develop...develop` (fetch 직후) |
| 판정 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t909` | `git rev-parse --show-toplevel` |
| 작업 브랜치 | `WT-agent-roster-count` | `git branch --show-current` |
| 도입 커밋 | `5ec516165` (2026-09-15) | `git log -- .claude/agents/moai/mission-governor.md` |
| 정본 개수 강제 지점 | `catalog_tier_audit_test.go:258` `expectedAgentCount = 12` | `sed -n '258p'` |
| 재생성 후 catalog 크기 | 13145 bytes | `gen-catalog-hashes.go --all` 출력 |

배차가 준 기준값(develop `2ede711ca`, 사본 셋 12개, §4 는 11)은 이 트리에서 그대로 재측정돼
일치했다.

## Gaps

- **[핵심] 고친 숫자를 지키는 가드가 없다.** CLAUDE.md 의 `13 retained agents` / `(12 MoAI-custom)`
  프로즈를 파일 실체와 대조하는 테스트는 존재하지 않는다(확인: `grep -rln "retained agents"
  --include='*_test.go' internal/` 가 잡은 3파일 — `embed_test.go`·`catalog_tier_audit_test.go`·
  `agent_lint_test.go` — 는 전부 파일 개수나 lint 규칙을 보고, CLAUDE.md 프로즈는 읽지 않는다).
  **에이전트가 하나 더 들어오면 같은 자리가 또 낡는다.** 이번 수리는 그 재발을 막지 않는다.
- **`internal/template/catalog.yaml` 과 `.codex/…/manager-design.toml` 은 손으로 확인하지
  않았다.** 명시 동사(`make agents-emit`, `gen-catalog-hashes.go --all`)로 재생성하고 드리프트
  검사가 양방향으로 판정했을 뿐, 재생성기의 출력 내용을 줄 단위로 읽지는 않았다.
- **`mission-governor` 에 결정 트리 행을 넣지 않은 판단은 문서 판독 근거다.** `goal.md:147` 과
  에이전트 프론트매터(`permissionMode: plan`, `tools: Read, Grep, Glob, Skill`)로 "기계 배차되는
  읽기전용 결정 역할"이라 읽었고, GTD 자동 미션을 **실행해 스폰 경로를 라이브로 관측하지는
  않았다**.
- **D2·D3 를 수리하지 않았다** — 리드 판정에 따라 별 카드. 초안은 아래 절에 있다. 그동안
  `llm.agent_overrides.manager-lead` 는 계속 거절되고, delegationmap 은 mission-governor 위임을
  계속 비카탈로그로 분류한다.
- **docs-site·`delegation.yaml` 주석은 손대지 않았다** — `11 retained`·`10 MoAI-custom`·
  `8 retained` 가 섞여 있다. 4로케일 동반 의무가 붙어 별 카드 규모(§ 별 카드 초안 3).
- **전체 스위트를 돌리지 않았다** — 영향 패키지(`internal/template/…`, `agentlint`,
  `delegationmap`)만 돌렸다. 전 패키지 판정은 CI 몫이고 미관측이다.
- **원격 착지·CI 판정 미관측** — push 는 리드 일괄.

## 범위 확장 2 — 이 카드의 수리가 깨뜨린 파일 (리드 판정으로 포함)

[HARD 명시] **카드는 이 두 파일을 지목하지 않았다. 이 카드의 수리가 깨뜨렸기 때문에 함께
고쳤다.** 조용한 확장이 아니라 선언된 확장이다.

**깨뜨린 기전.** `.claude/rules/moai/development/agent-authoring.md`(+템플릿 쌍둥이) § Agent
Categories 는 스스로 *"…, **aligned with CLAUDE.md §4**"* 라고 선언한다. §4 를 13/12 로 고친
순간 그 선언이 거짓이 된다 — 안 고치고 병합하면 알면서 깨진 상태를 올리는 것이고, 다음 사람은
어느 쪽이 정본인지 또 못 가른다(이 카드가 해결한 바로 그 상황의 재생산).

**같은 줄의 선존재 오류도 함께 고쳤다.** `### Retained MoAI-custom Agents (11)` 제목 아래 목록이
실제로는 **10개만** 나열했다(manager-lead·mission-governor 둘 다 없음) — 제목과 목록이 이미
어긋나 있었다. 어차피 같은 줄을 건드리므로 분리하지 않았다.

| 파일 | 고친 것 |
|---|---|
| `agent-authoring.md`(+쌍둥이) :128 | `12 retained (11 MoAI-custom)` → `13 (12)`; `manager-lead … added later per the hierarchical-team SPEC` → `manager-lead … and mission-governor … were added later` |
| `agent-authoring.md`(+쌍둥이) :130 | 제목 `(11)` → `(12)` |
| `agent-authoring.md`(+쌍둥이) :144-145 | 목록에 `manager-lead` · `mission-governor` 두 항목 추가 (10 → 12) |
| `agent-patterns.md`(+쌍둥이) :234 | `The 11 MoAI-custom retained agents (… manager-lead)` → `The 12 …, mission-governor)` |
| `agent-patterns.md`(+쌍둥이) :269 | `12-agent catalog roles` → `13-agent`; 4-Loop 표가 전 카탈로그를 열거하지 않으며 mission-governor 가 그 표에 없는 이유를 한 절로 명시 |

**쌍둥이 비대칭 보존.** 두 사본은 의도적으로 갈라져 있다 — 로컬 `agent-authoring.md:128` 은
`(SPEC-AGENT-ARCH-V2-001)` 을 담고 템플릿 사본은 담지 않으며(템플릿 중립성), `agent-patterns.md
:269` 도 로컬만 SSOT 절 번호를 인용한다. 각 사본을 **자기 문안 그대로** 고쳐 그 비대칭을 유지
했다. 내가 편집한 구간에 SPEC 토큰이 새로 들어가지 않았음을 확인했다:

```
$ sed -n '126,150p' internal/template/templates/.claude/rules/moai/development/agent-authoring.md | grep -c "SPEC-"
0
```

**검증 (확장분).**

```
$ make agents-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.367s
```

에이전트 `.md` 층을 건드리지 않았으므로 codex 방출·catalog 해시는 이번엔 동반되지 않았다 —
②에서와 달리 재생성이 필요 없다는 것이 이 검사의 판정이다.

```
$ unset <6변수> && go test ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	56.731s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.310s
```

**[HARD] 그러나 이 초록은 두 파일을 보고 얻은 것이 아니다.** 어떤 가드도 이 두 파일을 덮지
않는다 — 실측:

```
$ grep -n "agent-authoring\|agent-patterns" internal/template/catalog.yaml
(무출력 — catalog 해시는 rules 파일을 덮지 않는다)
$ grep -rn "agent-authoring\|agent-patterns" internal/template/rule_template_mirror_test.go
(무출력 — 바이트 패리티 목록에도 없다; 그 목록에는 model-policy.md 등만 있다)
```

즉 이 두 파일이 §4 와 갈라진 것은 **가드가 없어서**이고, 이번 수리도 그 재발을 막지 않는다
(§ Gaps 의 가드 부재 항목과 같은 사안 — 표면이 하나 더 있다는 뜻).

## Residual-risk
- **결정 트리 행 부재가 다음 사람에게 결함으로 읽힐 수 있다.** 붙여 둔 한 절이 그것을 막는
  장치인데, 그 절 자체를 지키는 가드도 없다.
- **`13`이라는 숫자는 mission-governor 를 포함한 현재 상태의 스냅샷이다.** 상한을 다루는 다른
  문장들(`agent-authoring.md` 의 "12-agent retention ceiling" 등)과의 정합은 이 카드에서
  판정하지 않았다 — 위 첫 항목과 같은 사안이다.

---

## 별 카드 초안 (리드 요청 — 운영자 복귀 후 발행)

### 초안 1 — D2: `profile.go` `retainedAgentNames` 에 manager-lead 누락 (기능 결함)

**증상.** `internal/config/profile.go:142` 의 `retainedAgentNames` — 주석이 스스로
*"the closed set of canonical retained-agent names an `llm.agent_overrides` entry may key on"* 라고
선언한 닫힌 집합 — 에 `manager-lead` 가 없다. `profile.go:181` 이 그 맵으로 키를 검증하므로,
운영자가 `llm.agent_overrides.manager-lead` 를 쓰면
`agent manager-lead is not in the retained agent catalog` 로 **거절된다.** 한편
`internal/template/profile_matrix.go` 는 manager-lead 를 3 프로필 전부에서 매핑한다(`:329`·`:344`·
`:359`) — 매트릭스는 알고 있는데 override 검증은 모른다.

**전례 — 같은 계열의 재발.** `profile_matrix.go:230` 주석:

> `manager-lead` was absent from this list until t205 and therefore resolved to the unmapped-agent
> `inherit` sentinel — the Tier L coordinator, the one row that fans out to every other agent, took
> whatever the session happened to be on. It is a mapped row now.

**t205 가 형제 목록 한쪽만 고쳤다.** 이 카드는 나머지 한쪽이다.

**RED 먼저.** `internal/config` 에 `llm.agent_overrides.manager-lead` 를 담은 설정을 로드해
`agent manager-lead is not in the retained agent catalog` 검증 오류가 나오는 것을 재현 테스트로
고정한 뒤 수리한다. 「고친 뒤 초록」은 근거가 안 된다 — 거절이 실제로 일어남을 먼저 보여야 한다.

**수리.** `retainedAgentNames` 에 `"manager-lead": true` 추가. 추가 검토 대상: 두 목록이 다시
갈라지지 않게 하는 가드(한쪽에서 다른 쪽을 파생시키거나, 두 목록의 집합 동일성을 테스트로 고정).

**검증.** `unset <6변수> && go test ./internal/config/...`

---

### 초안 2 — D3: `delegationmap` `retainedCatalog` 에 mission-governor 누락 (분류 누락)

**증상.** `internal/harness/delegationmap/types.go:81` 의 `retainedCatalog` 에
`mission-governor` 가 없다. `IsRetainedAgent` 이 그 맵을 읽고(`:102`), `analyze.go:118` ·
`aggregate.go:86` 이 그 판정으로 비카탈로그 에이전트를 걸러낸다. 그래서 mission-governor 위임은
**"undesignated" 발견에서 조용히 제외된다** — 그 분석기가 존재하는 이유인 발견이 억제된다.

**그 파일이 스스로 적어 둔 자리.** `types.go:79-80`:

> The cost is a snapshot that can go stale against a catalog change; this comment is where that
> change lands.

즉 이 수리는 그 주석이 지목한 자리에 들어간다.

**RED 먼저.** mission-governor 위임 레코드를 넣은 분석 입력으로 `IsRetainedAgent` 이 false 를
반환하고 그 결과 undesignated 판정이 억제되는 것을 재현한 뒤 수리한다.

**수리.** `retainedCatalog` 에 `"mission-governor": {}` 추가. D2 와 같은 「두 목록 동일성 가드」
검토 대상.

**검증.** `unset <6변수> && go test ./internal/harness/delegationmap/...`

**D2 와 한 카드로 묶을 수 있다** — 같은 결함 계열(로스터 목록이 파일 실체와 갈라짐)이고 둘 다
코드 변경 + RED 재현이 필요하다. 다만 거절 표면이 서로 달라(설정 검증 vs 분석 분류) 재현
테스트는 각각 따로 세워야 한다.

---

### 초안 3 — docs-site·delegation.yaml 로스터 스윕 (문서, 4로케일)

**증상.** 배포 문서에 낡은 개수가 섞여 있다. 실측:

| 파일 | 행 | 적혀 있는 값 |
|---|---|---|
| `docs-site/content/ko/core-concepts/what-is-moai-adk.md` | 85, 244 | 12개 (11 MoAI 커스텀) |
| `docs-site/content/en/advanced/agent-guide.md` | 42 | 12 core agents (11 MoAI custom) |
| `docs-site/content/en/advanced/profile-matrix.md` | 7, 32 | 12 retained agents / 36 cells (12 × 3) |
| `docs-site/content/en/advanced/codex-dual-harness.md` | 23 | 11 retained agents |
| `docs-site/content/en/advanced/config-sections.md` | 58 | 11 retained agents |
| `docs-site/content/en/advanced/claude-md-guide.md` | 103 | **11 retained (10 MoAI-custom)** |
| `docs-site/content/en/getting-started/migration.md` | 118 | one of 11 retained agents |
| `docs-site/content/en/core-concepts/what-is-moai-adk.md` | 243 | 11 retained (10 MoAI custom) |
| `.moai/config/sections/delegation.yaml` | 14 | 주석 `the 11 retained agents (CLAUDE.md section 4)` |

`profile-matrix.md` 는 셀 수까지 적고 있어(`36 cells = 12 agents × 3 profiles`) 실제 매트릭스
행 수(13)와 대조가 필요하다.

**범위.** docs-site 는 4로케일 동반 의무가 붙는다(ko 정본 → en → ja/zh). 위 표는 ko·en 적중만
나열했고 ja·zh 는 미조사다. `delegation.yaml` 주석은 별 축(설정 파일, 로케일 무관).

**주의.** 개수만 일괄 치환하지 말 것 — `claude-md-guide.md:103` 처럼 **10 MoAI-custom** 으로 적힌
행은 다른 시점의 스냅샷이라 문맥을 읽고 고쳐야 한다.

---

## 처분 요청

| 항목 | 상태 |
|---|---|
| D1 (CLAUDE.md 쌍둥이) | **수리 완료** — 13/12, mission-governor 명단 추가 |
| 범위 확장 1 (manager-design 쌍둥이) | **수리 완료** — 개수 + 트리 행 위치 둘 다 |
| 범위 확장 2 (agent-authoring · agent-patterns 쌍둥이) | **수리 완료** — 이 카드의 수리가 깨뜨린 파일, 리드 판정으로 포함 (§ 범위 확장 2) |
| 기계 재생성 2건 | 완료 (`make agents-emit`, `gen-catalog-hashes --all`) · 드리프트 양방향 확인 |
| 영향 패키지 테스트 | ok (template / agentemit / commandemit / agentlint / delegationmap) · vet 0 |
| 별 카드 요청 | 가드 부재 — CLAUDE.md 프로즈 카운트와 `agent-authoring`·`agent-patterns` 로스터를 파일 실체와 대조하는 검사가 없다 (리드가 판정 목록에 올림) |
| 별 카드 초안 | D2 · D3 · docs-site 스윕 3건 (위) |
| 이 브랜치 | `WT-agent-roster-count` — 병합 창 요청 |
| 워크트리 | `t909` · `t278-t267` · `t278` · `t267` · `t880` 전부 보존 |
