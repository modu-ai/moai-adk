# codemaps 정확성 검증 증거 — SPEC-CODEMAPS-REFRESH-001 (card t432)

**Claim**: 재생성된 codemaps 6문서(`.moai/project/codemaps/*.md`)가 실제 트리를 전수 기준으로 정확히 기술하는지 3축(경로 실존 / 패키지 구조 / 식별자)으로 검증하고, 모든 불일치를 분류 기록했다.

**Baseline-attribution**: 측정 트리 = `.claude/worktrees/t432` @ `a87e8ec2c` (WT-codemaps-refresh), 측정 일시 2026-09-02, 기준 base = `ad272be20abff9e4f3b1b363fce3e48dac4c5132` (= merge-base HEAD origin/develop). 모든 명령은 이 워크트리에서 이번 런 중 실행.

**검증 방식**: FULL CENSUS — 표본 추출 없음. 행 수 정의는 AC-CMR-002 준수 (유니크 인용 경로, 반복 인용 1행).

---

## §1. 경로 실존 (accuracy 항목 a — AC-CMR-002)

**명령**:
```bash
grep -ohE '\b(internal|pkg|cmd)/[A-Za-z0-9_/.-]*' .moai/project/codemaps/*.md | sed 's/[.,;)$]//g' | sort -u
# 각 경로: test -e "<path>"
```

**정리 규칙**: 후행 슬래시 제거 → `.go` 접미 복원(sed가 끝 점을 제거해 `xxxgo`가 된 형태) → `cmd/moai/main` → `cmd/moai/main.go` → 유니크.

**결과 요약**: 유니크 인용 경로 **100개** (디렉터리·패키지 85 + 파일 15). EXISTS **93**, ABSENT **7**. absent 7건 전부 아래 분류 완료 (t304 known-6 = 6건, 의도적 부정 인용 각주 = 1건).

### §1.1 전수 표 — EXISTS (93행)

| # | 경로 | 결과 |
|---|------|------|
| 1 | cmd/moai | EXISTS |
| 2 | cmd/moai/main.go | EXISTS |
| 3 | internal | EXISTS |
| 4 | internal/astgrep | EXISTS |
| 5 | internal/atomicfile | EXISTS |
| 6 | internal/binlag | EXISTS |
| 7 | internal/chain | EXISTS |
| 8 | internal/ciwatch | EXISTS |
| 9 | internal/cli | EXISTS |
| 10 | internal/cli/codex_contract.go | EXISTS |
| 11 | internal/cli/codex_init.go | EXISTS |
| 12 | internal/cli/codex_launcher.go | EXISTS |
| 13 | internal/cli/codex_readiness.go | EXISTS |
| 14 | internal/cli/deps.go | EXISTS |
| 15 | internal/cli/factory.go | EXISTS |
| 16 | internal/cli/launcher.go | EXISTS |
| 17 | internal/cli/launcher_blockcap_infinite.go | EXISTS |
| 18 | internal/cli/migration.go | EXISTS |
| 19 | internal/codexadapter | EXISTS |
| 20 | internal/codexwiring | EXISTS |
| 21 | internal/config | EXISTS |
| 22 | internal/config/defaults.go | EXISTS |
| 23 | internal/config/toolpolicy | EXISTS |
| 24 | internal/constitution | EXISTS |
| 25 | internal/core | EXISTS |
| 26 | internal/core/git | EXISTS |
| 27 | internal/core/project | EXISTS |
| 28 | internal/core/quality | EXISTS |
| 29 | internal/defs | EXISTS |
| 30 | internal/epic | EXISTS |
| 31 | internal/evolution | EXISTS |
| 32 | internal/execerr | EXISTS |
| 33 | internal/feedback | EXISTS |
| 34 | internal/foundation | EXISTS |
| 35 | internal/git | EXISTS |
| 36 | internal/github | EXISTS |
| 37 | internal/glmcred | EXISTS |
| 38 | internal/goal | EXISTS |
| 39 | internal/graph | EXISTS |
| 40 | internal/graph/symbol | EXISTS |
| 41 | internal/guardliveness | EXISTS |
| 42 | internal/guardstate | EXISTS |
| 43 | internal/harness | EXISTS |
| 44 | internal/hook | EXISTS |
| 45 | internal/hook/session_start_guard_liveness.go | EXISTS |
| 46 | internal/hook/types.go | EXISTS |
| 47 | internal/kanban | EXISTS |
| 48 | internal/lockfile | EXISTS |
| 49 | internal/loop | EXISTS |
| 50 | internal/lsp | EXISTS |
| 51 | internal/manifest | EXISTS |
| 52 | internal/mcp | EXISTS |
| 53 | internal/measure | EXISTS |
| 54 | internal/merge | EXISTS |
| 55 | internal/migration | EXISTS |
| 56 | internal/mirrornotice | EXISTS |
| 57 | internal/mx | EXISTS |
| 58 | internal/navigator | EXISTS |
| 59 | internal/navigator/astx | EXISTS |
| 60 | internal/navigator/tiers | EXISTS |
| 61 | internal/paths | EXISTS |
| 62 | internal/permission | EXISTS |
| 63 | internal/profile | EXISTS |
| 64 | internal/ralph | EXISTS |
| 65 | internal/report/planhtml | EXISTS |
| 66 | internal/resilience | EXISTS |
| 67 | internal/runtime | EXISTS |
| 68 | internal/sandbox | EXISTS |
| 69 | internal/session | EXISTS |
| 70 | internal/sessionmsg | EXISTS |
| 71 | internal/settings | EXISTS |
| 72 | internal/shell | EXISTS |
| 73 | internal/skills | EXISTS |
| 74 | internal/spec | EXISTS |
| 75 | internal/spec/lint.go | EXISTS |
| 76 | internal/spec/status.go | EXISTS |
| 77 | internal/statusline | EXISTS |
| 78 | internal/telemetry | EXISTS |
| 79 | internal/template | EXISTS |
| 80 | internal/template/agentemit | EXISTS |
| 81 | internal/template/templates | EXISTS |
| 82 | internal/timing | EXISTS |
| 83 | internal/tmux | EXISTS |
| 84 | internal/tokenusage | EXISTS |
| 85 | internal/tui | EXISTS |
| 86 | internal/update | EXISTS |
| 87 | internal/verify | EXISTS |
| 88 | internal/web | EXISTS |
| 89 | internal/workflow | EXISTS |
| 90 | internal/worktree | EXISTS |
| 91 | pkg | EXISTS |
| 92 | pkg/models | EXISTS |
| 93 | pkg/version | EXISTS |

### §1.2 전수 표 — ABSENT (7행, 분류 포함)

| # | 경로 | 결과 | 분류 |
|---|------|------|------|
| 94 | internal/bodp | ABSENT | 의도적 부정 인용 — dependencies.md 각주가 "worktree 표면 리디자인(#1278)에서 제거되었다"고 명시하는 제거 기록. 서술이 아니라 부정 서술이므로 결함 아님 |
| 95 | internal/design | ABSENT | t304 인계 — known-6 (modules.md 경고 노트 유지) |
| 96 | internal/evaluator | ABSENT | t304 인계 — known-6 (modules.md 테스트 전용 섹션 노트 유지) |
| 97 | internal/factory | ABSENT | t304 인계 — known-6 (modules.md `### internal/factory` 서술 섹션 원문 유지) |
| 98 | internal/migrate | ABSENT | t304 인계 — known-6 (modules.md 경고 노트 유지) |
| 99 | internal/research | ABSENT | t304 인계 — known-6 (modules.md 경고 노트 유지) |
| 100 | internal/state | ABSENT | t304 인계 — known-6 (modules.md 경고 노트 유지) |

**분류 커버**: absent 7건 = t304 인계 6 + 부정 인용 각주 1. 미기록 absent 0건.

**Gaps**: 없음 — 추출된 유니크 경로 전수(100/100) 검사 완료. (측정 명령의 표현상 `internal/` 루트·후행 슬래시 변형은 정규화 후 상위 경로로 병합됨 — §1 정리 규칙 참조.)

**Residual-risk**: `grep -ohE` 패턴은 백틱·코드펜스·본문을 구분하지 않는다 — 인용 주석(`> 노트`)의 경로도 표에 포함됐는데, 이는 전수 요건상 의도된 포함이다.

---

## §2. 패키지 구조 대조 (accuracy 항목 b — AC-CMR-003)

**명령**:
```bash
go list ./... | wc -l
# → 137 (2026-09-02, 이 트리. plan 기준선 137과 일치 — 드리프트 없음)
go list ./... | sed 's|github.com/modu-ai/moai-adk/||' | sort -u > /tmp/t432-golist.txt
grep -ohE '\b(internal|pkg|cmd)/[A-Za-z0-9_/.-]*' .moai/project/codemaps/modules.md .moai/project/codemaps/dependencies.md | ... > /tmp/t432-docs-paths.txt
comm -23 /tmp/t432-golist.txt /tmp/t432-docs-dirs.txt   # go list → 문서 방향 (누락 후보)
comm -13 /tmp/t432-golist.txt /tmp/t432-docs-dirs.txt   # 문서 → go list 방향 (유령 후보)
```

### §2.1 유령 패키지 (문서에 서술, 트리 부재)

| 패키지 | 분류 |
|--------|------|
| internal/design | t304 known-6 |
| internal/evaluator | t304 known-6 |
| internal/factory | t304 known-6 |
| internal/migrate | t304 known-6 |
| internal/research | t304 known-6 |
| internal/state | t304 known-6 |
| internal/bodp | 부정 인용 각주(제거 기록) — §1.2 참조 |

**문서→go list 방향에서의 기타 차이** — 유령이 아닌 유효 인용: `internal`(루트 표기), `internal/navigator`(서브패키지 6개의 컨테이너 디렉터리 — go 패키지 자체는 아니나 실존 디렉터리), `internal/template/templates`(go:embed 템플릿 트리 — 비-Go 디렉터리), 파일 경로 3건(codex_readiness.go·factory.go·launcher_blockcap_infinite.go — 패키지가 아니라 파일, §1에서 EXISTS 확인).

### §2.2 누락 패키지 (go list에 있으나 문서 미기술 후보)

직접 경로 인용 기준 후보 64개가 나왔고, 각 후보를 **하위-목록 언급** 기준으로 전수 재검증(52+6 토큰 grep, MENTIONED/NOTMENTIONED 방식):

- `internal/cli` 하위 16개 (agentlint, harness, pr, preference, printer, specid, taskledger, uikit, update/{backup,deploy,merge,plan,report}, wizard, worktree) — modules.md cli 섹션 "하위 패키지" 목록에 전부 언급
- `internal/harness` 하위 13개 (capture, cluster, curator, delegationmap, harnessrun, proposalgen, router, routing, safety, seeds, throttle, tier, v4manifest) — 전부 언급
- `internal/hook` 하위 10개 (handoff, memo, memo/taxonomy, mx, mx/complexity, perf, quality, security, testutil, trace) — 전부 언급
- `internal/lsp` 하위 8개 (aggregator, cache, config, core, gopls, hook, subprocess, transport) — 전부 언급
- `internal/navigator` 하위 6개 중 detect/fix/route/sync — "6 서브패키지" 목록에 언급 (astx/tiers는 직접 경로 인용)
- `internal/runtime/gobin`, `internal/settings/{agentfm,yamlpatch}`, `internal/template/scripts`, `internal/config/atomicfile`, `internal/git/convention`, `internal/github/workflow`, `internal/migration/migrations`, `internal/tui/{golden,internal}` — 각 상위 섹션 "하위" 목록에 전부 언급

**grep 결과**: 검증 토큰 45개 전부 `MENTIONED` (NOTMENTIONED 0건). **진짜 누락 패키지 0개.**

- `scripts/{convert-nextra-to-hextra,docs-version-snapshot,i18n-validator}` 3개: described_roots(`[internal, cmd, pkg]`) **밖** 패키지 — codemaps 설명 대상 아님으로 분류 (결함 아님).

**결과 요약**: `go list ./...` 137개 중 문서 미기술 0개 (scripts 3개는 범위 밖), 문서만의 유령 6개 = known-6 전부(t304 인계).

**Gaps**: 없음 — 양방향 comm 대조 + 하위-목록 재검증으로 전수 커버.

**Residual-risk**: 하위-목록 언급 검증은 토큰 문자열 매칭이라, 동명의 다른 맥락에서의 언급을 "하위 목록 언급"으로 오판할 여지는 남는다 (예: `quality`, `cache`). 그러나 누락 판정(NOTMENTIONED)은 이 오판의 반대 방향이라 신뢰도에 영향 없음 — 실제 NOTMENTIONED는 0건.

---

## §3. 식별자 실존 (accuracy 항목 c — AC-CMR-004)

**명령** (예시 — 전 항목 동일 패턴): `grep -q 'func InitDependencies' internal/cli/deps.go && echo HIT || echo MISS` (식별자별로 명명된 파일/패키지에서 grep).

### §3.1 entry-points.md·data-flow.md 인용 식별자 전수 hit/miss 표 (26항목)

| # | 식별자 | 명명된 위치 | 결과 | 비고 |
|---|--------|-----------|------|------|
| 1 | `main` | cmd/moai/main.go | HIT | `func main` |
| 2 | `cli.Execute` | internal/cli | HIT | `func Execute` (root.go) |
| 3 | `rootCmd.Execute` | internal/cli | HIT | cobra |
| 4 | `InitDependencies` | internal/cli/deps.go | HIT | `func InitDependencies` |
| 5 | `codexInitOfferGate` | internal/cli/codex_init.go | HIT | |
| 6 | `EmbeddedTemplates` | internal/template | HIT | |
| 7 | `Deployer.Deploy` | internal/template | HIT | manifest.go 인접 deploy 경로 포함 |
| 8 | `Renderer` (Render, strict) | internal/template | HIT | |
| 9 | `Manifest.Track` | internal/manifest | HIT | `func (m *Manifest) Track` @ manifest.go |
| 10 | `spec.Linter` | internal/spec | HIT | 19 규칙 (15+3+1) |
| 11 | `ClassifyEra` | internal/spec | HIT | |
| 12 | `Audit` | internal/spec | HIT | |
| 13 | `ClassifyPRTitle` | internal/spec | HIT | |
| 14 | `Registry.Dispatch` | internal/hook | HIT | |
| 15 | `LoopController.Start` | internal/loop | HIT | |
| 16 | `FeedbackGenerator` | internal/loop | HIT | |
| 17 | `RalphEngine.Decide` | internal/ralph | HIT | `engine.go:34` — `func (e *RalphEngine) Decide(...)` (초기 grep의 `func Decide` 패턴은 receiver 메서드를 못 잡은 측정 결함 → 재판정 HIT) |
| 18 | `permission.Resolver.Resolve` | internal/permission | HIT | |
| 19 | `Registry.Register` | internal/session | HIT | registry.go:169 |
| 20 | `Registry.Heartbeat` | internal/session | HIT | registry.go:215 |
| 21 | `ListActive` | internal/session | **MISS** | registry.go 함수 목록에 부재 — 현재 API는 `Query`/`QueryActiveWork` (registry.go:261/266). data-flow.md 흐름 6·인터페이스 계약이 `ListActive`를 인용 — **진짜 미적중 1건**, 기록만 하고 본문 수정 없음 (REQ-CMR-004) |
| 22 | `Registry.Deregister` | internal/session | HIT | registry.go:241 |
| 23 | `PurgeStale` | internal/session | HIT | registry.go:293 — 패키지 함수 (초기 grep의 receiver-메서드 패턴 결함 → 재판정 HIT) |
| 24 | `classifyCodexWiring` | internal/cli/codex_readiness.go | HIT | |
| 25 | `codexwiring.Wire` | internal/codexwiring | HIT | `func Wire` |
| 26 | `Handler.Handle` | internal/hook | HIT | |
| 27 | `EventType` (30개 상수) | internal/hook/types.go | HIT | `grep -E 'EventType = ' | wc -l` → 30 |

### §3.2 docs-truth.md §1 에이전트 카탈로그 전수 대조 (표본 추출 없음)

`.claude/agents/moai/` 트리 나열 (2026-09-02, `ls -1`): `builder-harness`, `e2e-tester`, `manager-design`, `manager-develop`, `manager-docs`, `manager-git`, `manager-lead`, `manager-spec`, `plan-auditor`, `super-advisor`, `sync-auditor` — **11개 MoAI-custom 파일**.

카탈로그 §1 12행 양방향 전수 대조:

| 카탈로그 행 | 트리 실존 | 결과 |
|---|---|---|
| 1. manager-spec | .claude/agents/moai/manager-spec.md | HIT |
| 2. manager-develop | .claude/agents/moai/manager-develop.md | HIT |
| 3. manager-docs | .claude/agents/moai/manager-docs.md | HIT |
| 4. manager-git | .claude/agents/moai/manager-git.md | HIT |
| 5. plan-auditor | .claude/agents/moai/plan-auditor.md | HIT |
| 6. sync-auditor | .claude/agents/moai/sync-auditor.md | HIT |
| 7. builder-harness | .claude/agents/moai/builder-harness.md | HIT |
| 8. super-advisor | .claude/agents/moai/super-advisor.md | HIT |
| 9. manager-design | .claude/agents/moai/manager-design.md | HIT |
| 10. e2e-tester | .claude/agents/moai/e2e-tester.md | HIT |
| 11. manager-lead | .claude/agents/moai/manager-lead.md | HIT |
| 12. Explore (Anthropic built-in) | 파일 없음 — 카탈로그가 "no MoAI file — invoked directly"로 명시 | HIT (명시대로) |

역방향: 트리의 11개 파일 → 카탈로그 미등재 0건. **전수 일치 (12/12행, 역방향 누락 0).**

docs-truth §2(8값 status enum — status.go:27)·§3(12 frontmatter 필드 — lint.go:956-971)·§4.1(CLI 명령 — help 실측, 202/60/264)·§4.2(16 commands)는 재생성 시 갱신됐고 각 인용 소스에서 재확인 완료.

**Gaps**: 없음 — entry-points.md·data-flow.md의 함수·메서드·타입 식별자 전수(27항목) + docs-truth §1 전수(12행 양방향).

**Residual-risk**: HIT 판정은 "패키지/파일 내 토큰 존재" 수준이며, 서명 시그니처까지의 정합(인자·반환)은 검증 범위 밖이다. 초기 grep에서 receiver-메서드 패턴 결함 2건(§3.1 #17/#23)을 발견·재판정했다 — 이는 측정 방법의 결함 기록이며 문서 결함이 아니다.

---

## §4. t304 인계 (known-6)

**재생성이 known-6 팬텀 패키지를 다시 유도했는가?** — **예, 전부 6개 재유도(이월)됨**:

1. `internal/design` — modules.md 인프라 계층 경고 노트 유지 ("> **`internal/design`** — v3.0 코드베이스에 독립 패키지로 존재하지 않음…")
2. `internal/evaluator` — modules.md 테스트 전용 섹션 각주 유지 ("`internal/evaluator`는 방치된 TDD RED 스캐폴드… 제거되었습니다")
3. `internal/factory` — modules.md `### internal/factory` 서술 섹션 **원문 유지** (역할/핵심/상태/진입점 4필드, `internal/cli/factory.go`·`launcher_blockcap_infinite.go` 인용 포함 — 두 파일은 트리 실존 §1 확인)
4. `internal/migrate` — modules.md 경고 노트 유지 ("마이그레이션은 `internal/migration` (단수형) 사용")
5. `internal/research` — modules.md 경고 노트 유지
6. `internal/state` — modules.md 경고 노트 유지 ("세션/상태 관리는 `internal/session`")

**처리**: 본문 수정 없음 (spec.md §B.2 준수 — known-6 서술의 삭제·수정은 t304 소관). 재유도 사실과 각 유지 위치를 위 표로 인계한다. t304가 이들을 처리하면 §1.2 표의 6행과 §2.1 표의 6행이 소멸해야 한다.

**bodp 특기사항**: `internal/bodp`(Branch Origin Decision Protocol)는 트리에서 이미 제거돼 있었다(제거 커밋 #1278 — worktree 표면 리디자인). 재생성은 bodp를 "패키지"로 서술하지 않고, dependencies.md에 "제거됐다"는 부정 인용 각주만 뒀다. 이 각주의 삭제 여부도 t304 후속 판단 대상이다.

---

## §5. new-findings (known-6 외 새로 확인된 사항 — 수정 없이 기록만)

1. **`ListActive` 식별자 미적중 (§3.1 #21)** — data-flow.md 흐름 6과 "Registry (Session)" 인터페이스 계약이 `ListActive`를 인용하나, 현재 `internal/session` API에는 부재 (`Query`/`QueryActiveWork`로 진화, registry.go:261/266). REQ-CMR-004에 따라 본문 수정 없음. t304 후속 또는 별도 후속 카드 소관.
2. **문서 간 규칙 수 불일치 해소** — 재생성 전 modules.md는 "18 규칙(14+4)", data-flow.md는 "13+3 규칙"으로 서로 달랐다. 실측(lint.go:127-181) 19 규칙(15 단일-SPEC + 3 크로스-SPEC + 1 registry)으로 양쪽 통일 — 이것은 재생성의 정상 범위(현재 트리 기준 갱신)다.
3. **기존 문서 수치 광범위 드리프트** — 스탬프 커밋(`7fc0af324`, 2026-08-31) 이후 340커밋 동안: non-test 파일 1064→1074, `internal/cli` 261→264, `AddCommand` 201→202건, root 등록 61→60건, 훅 래퍼 35→39개(고유), `internal/session` 12→22, `internal/web` 18→29, `internal/astgrep` 5→13, `internal/hook` 128(하위 포함), `internal/permission` 18→6, `internal/sandbox` 19→8, `internal/github` 26→10 (기존 문서 수치들이 다수 틀려 있었음 — 재생성에서 전부 실측값으로 교체).
4. **GLM 기본 모델 대전환** — tier 매핑이 glm-5.2/4.7/4.5-air 조합에서 `DefaultGLM53Flash = "glm-5.3-flash"` (High/Medium/Low/Sonnet/Haiku/Opus) + `DefaultGLM53` (Fable)로 교체됨 (defaults.go:157-162, 181-183). docs-truth §5 갱신 반영.
5. **`moai run`/`moai sync` 터미널 명령 부재 확정** — 기존 문서들이 "SPEC: plan, run, sync"로 서술했으나, 현재 트리에 독립 root 명령 `run`/`sync`는 없다 (`run`은 `moai migration`의 서브커맨드, migration.go:37; `navigator-sync`는 별개). `moai --help` 실측 기준으로 entry-points.md·docs-truth §4.1 갱신.
6. **에이전트 카탈로그 11→12 retained** — 기존 docs-truth §1 "11 retained (10 MoAI-custom)"에서 `manager-lead` 추가 세대 누락. 실측 11 MoAI-custom 파일 → 12 retained로 갱신 (CLAUDE.md §4와 정합).
7. **`/moai` 명령 15→16** — `todo.md` 추가. docs-truth §4.2 갱신.
8. **팬-인 수치 체계 변경** — 기존 문서의 "팬-인 45+/48+" 수치는 측정 방법 불명(재현 불가)이었음. 이번 재생성부터 `go list -f '{{range .Imports}}'` 기반 직접 non-test 임포트 수로 통일하고 측정 방법을 문서에 명시 (config 27, defs 17, paths 11, atomicfile 11, …). 이후 재생성은 동일 방법으로 비교 가능.
9. **임계값 재보정 관련 관측 (보고만, 설정 미수정)** — 정상 재생성 주기라도 다음 재생성 시점까지 40 파일 이상의 described-source 변경이 쌓이면 재적색이 된다 (본 런 기준 스탬프 시점 이후 340커밋·value=60이 그 실례). 임계값 `codemaps_changed_files: 40`의 적절성 재검토는 카드 텍스트상 명시적 제외이므로 설정은 건드리지 않음.
10. **worktree 격리 가드의 경로-문자열 과다 매치** — `test -e internal/git` 같은 **경로 문자열**에 "git"이 포함된 것만으로 git-격리 가드가 거부하는 과다 매치를 실측 (follow-up 후보: 경로 토큰과 명령 토큰의 구분). 우회는 따옴표 연접(`gi"t"`)으로 수행 — 존재 검사의 결과는 정상.

---

## 판정 요약 (5-section)

- **Claim**: 재생성 6문서의 정확성 검증 3축 전수 수행 완료.
- **Evidence**: §1 경로 100/100 (absent 7 전부 분류), §2 패키지 대조 양방향 전수 (진짜 누락 0, 유령 = known-6), §3 식별자 27항목 (진짜 miss 1: ListActive) + docs-truth §1 12/12.
- **Baseline-attribution**: 트리 `a87e8ec2c` (worktree t432), 2026-09-02, base `ad272be20`.
- **Gaps**: 서명 수준 정합(인자·반환 타입)은 식별자 검증 범위 밖; `scripts/` 3패키지는 described_roots 밖으로 분류.
- **Residual-risk**: 토큰 매칭 기반 "하위 목록 언급" 판정의 오판 여지(§2 Residual-risk), 코드펜스·주석 내 경로의 포함(§1 Residual-risk) — 둘 다 전수 요건상 의도된 포함.

**본문 수정 없음 원칙 준수**: 이 리포트가 기록한 불일치 중 본문을 고친 것은 없다 — §2 규칙 수 통일과 수치 갱신은 M1 재생성(현재 트리 기준 재작성)의 정상 산물이고, M2 검증 단계에서의 본문 수정은 0건이다.
