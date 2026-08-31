# SPEC-EVIDENCE-CITATION-CANON-001 — 구현 계획

## A. 맥락

거짓 전제 하나(“gitignore된 `.moai/state/verify/`가 감사 시점 인용 대상이 된다”)를 규칙에서 걷어내고, 그 자리에 **인용 전 반출** 의무와 **인용 넓이 상한**을 세운다. 여기에 카드 t373이 미룬 ignore 정책 2건이 합류한다. 코드 변경은 가드 테스트 하나뿐이고 나머지는 문서·설정이다.

측정 기준 트리: 워크트리 `.claude/worktrees/t375`, HEAD `b64043481`. 수치 출처는 spec.md §1.1의 측정 스크립트 2개(이 카드가 plan-close 커밋으로 함께 추적한다).

## B. Tier 판정 — M (조건부 재확인)

| 축 | 값 | Tier M 기준 |
|---|---|---|
| 영향 파일 | **11 ~ 18** (§C.3 결정에 의존) | 5–15 |
| 신규 Go LOC | 가드 1개, 대략 200–300 | 300–1000 미만 |
| 요구사항 / AC | 13 / 15 | 상한 16 / 16 |

**파일 수가 범위인 이유**: §C.3의 스킬 3파일이 전부 carve-out이면 **11**, 전부 교체면 그 3파일 + 미러 3 + 방출 1이 더해져 **18**이다. 후자는 Tier M 구간(5–15) 밖이다.

[HARD] **M2가 §C.3을 결정한 직후 Tier를 재확인한다.** 교체가 2건 이상이면 파일 수가 15를 넘으므로, 그 시점에 Tier L 상향 또는 SPEC 분할을 판정하고 결과를 `progress.md` §E.2에 적는다. 초판은 11–13을 적었는데, 그것은 아직 내리지 않은 결정의 유리한 쪽을 가정한 값이었다(iter1 감사 D12).

**Tier S로 내리지 않는 이유**: 변경이 한 파일에 갇히지 않고 **규칙 · 에이전트 정의 · 출력 스타일 · `.gitignore` · 템플릿 미러**의 다섯 표면에 걸친다. **최소 케이스에서 L로 올리지 않는 이유**: 새 실행 기제를 만들지 않고 헌법 문서를 건드리지 않는다.

## C. 손볼 파일과 바뀌는 절

### C.1 규칙 본문 (핵심 — 되돌리기 가장 어려운 결정)

| 파일 | 바뀌는 절 | 무엇이 바뀌나 |
|---|---|---|
| `.claude/rules/moai/core/agent-common-protocol.md` | § Parallel Execution → “Evidence persistence” 불릿 (268행) | 인용 대상 지목 문장을 **스크래치 명명 + 추적 경로 인용 + 인용 전 반출**로 교체. REQ-ECC-001·002·003 |
| `.claude/rules/moai/core/agent-common-protocol-reference.md` | § Evidence persistence obligation (60행 문단) | 같은 교체 + **인용 넓이 상한**(REQ-ECC-004) + **선택 기준**(REQ-ECC-005) + **carve-out 2건**(REQ-ECC-006) |

**M1이 도입하는 고정 문구 3종**은 AC-ECC-002가 판정 명령으로 삼으므로 계약이다: `machine-local scratch`, `export before citing`, `.moai/reports/<card-id>`. 추가로 reference 파일에는 `names one file`과 `never the directory`(또는 `wholesale`), `residual-risk`, `state/verify/snapshots`, `internal/web/events.go`가 들어간다 — 전부 오늘 트리에서 0회이므로 AC가 RED에서 출발한다. 문구를 바꾸려면 AC도 함께 바꾸고, 바꾼 문구가 **수리 전 트리에서 0인지** 다시 확인한다.

**carve-out 문구가 이 SPEC에서 가장 섬세한 부분이다.** 기계 소비자가 **둘**이다(spec.md §1.4): `internal/verify/store.go:15`의 스냅샷 저장소와 `internal/web/events.go:29`의 fsnotify 감시. 후자는 `snapshots`가 아니라 `.moai/state/verify` **디렉터리 전체**를 본다. 교체 문구가 이것들까지 금지하면 attributable diff-check와 웹 콘솔 SSE가 규칙 위반이 된다. 문구는 **“사람이 읽는 인용문의 대상”**과 **“기계 소비자의 저장·감시 위치”**를 명시적으로 갈라야 한다.

### C.2 지시로 바뀌는 지점

REQ-ECC-002의 주어가 “규칙 문서”가 아니라 **doctrine 표면 문서**이므로, 아래 둘은 요구 층에 직접 대응한다(초판에서는 plan.md 절 번호에만 매핑돼 요구 근거가 없었다 — iter1 감사 D10).

| 파일 | 바뀌는 절 | 무엇이 바뀌나 |
|---|---|---|
| `.claude/agents/moai/manager-lead.md` | 61 / 85 / 143 / 146-147 / **150** / 157행 | Context-Folding 레시피의 `mkdir -p` + `tee` 대상은 스크래치로 두되 스크래치임을 밝히는 문장을 붙이고, **증거 표 열과 fold-row가 인용하는 경로**는 반출된 `.moai/reports/<card-id>/`로 바꾼다. 150행의 “canonical persistence location” 단정 삭제 |
| `.claude/output-styles/moai/moai.md` | §8 384 / 401 / 587행 | 배너가 박아 둔 경로를 `.moai/reports/<card-id>/`로 교체. `(persistent; … survive /tmp clearance)` 괄호 주석은 거짓 전제의 압축판이므로 함께 정정 |

### C.3 검토 후 결정 — 바꿀 수도, carve-out일 수도 (Tier 재확인 지점)

세 파일은 `.moai/state/verify`를 언급하지만 **기계 소비 문맥**일 가능성이 높다. M2는 파일당 아래 판별식을 적용하고 결과를 `progress.md` §E.2에 적는다(REQ-ECC-011 / AC-ECC-009).

> **판별식**: 이 경로를 최종적으로 읽는 것이 사람인가 기계인가? 사람이 판정 근거로 읽으면 C.1의 교체 대상, `moai` 명령이 키로 조회하면 carve-out.

| 파일 | 지점 | 사전 판단 |
|---|---|---|
| `.claude/skills/moai/workflows/gate.md` | 122행 `moai verify record` | carve-out 유력 — 스냅샷 저장소 |
| `.claude/skills/moai/workflows/loop.md` | 115행 “shared diagnostic snapshot” | carve-out 유력 — 같은 저장소 |
| `.claude/skills/moai/workflows/run.md` | 199행 `--kind verify_path --ref` | **경계 사례.** 원장에 남는 참조이므로 나중에 사람이 읽을 수 있다 |

교체로 결정된 파일은 미러도 함께 바뀌므로 파일 수가 2씩 늘어난다 — §B의 Tier 재확인이 여기 걸린다.

### C.4 `.gitignore`

| 파일 | 무엇이 바뀌나 |
|---|---|
| `.gitignore` (저장소 루트) | `.moai/observability/*.jsonl` **한 줄만** 추가 (REQ-ECC-012) |
| `internal/template/templates/.gitignore` | 같은 한 줄 미러 |

**초판에서 두 가지가 바뀌었다.**

1. `!.moai/observability/.gitkeep` 예외 줄을 **넣지 않는다.** 텔레메트리 형제의 `.gitkeep`은 추적될 뿐 아니라 **템플릿에도 실려 있어**(`internal/template/templates/.moai/evolution/telemetry/.gitkeep`) `moai init`이 모든 사용자 프로젝트에 배포한다. 같은 형태를 따르면 `.moai/observability/`가 배포판 전체에 생기고, `post_tool_duration.go`의 `os.Stat` opt-in이 opt-out으로 뒤집힌다. 근거 전문은 spec.md §4.2.
2. `.moai/project/navigator/fix-drafts/` 항목을 **넣지 않는다.** navigator 아래는 아무것도 무시하지 않는다 — `fix-drafts/`를 처분하는 코드가 없어 잔존이 실패 신호가 아니고(감사 D9의 반대 독해가 여기서 반증된다), 완료 여부를 가르는 `applied.json` 유무는 경로 기반 ignore가 구분하지 못하며, “아직 존재하지 않는 것을 예측으로 무시하지 않는다”는 논거가 다른 navigator 산출물과 동일하게 적용된다. 근거 전문은 spec.md §4.3.

두 판정 모두 `.moai/observability/`·`.moai/project/navigator/` **디렉터리 스캐폴드를 만들지 않는 것**을 포함한다.

### C.5 가드 (§D)

| 파일 | 성격 |
|---|---|
| `internal/template/evidence_citation_guard_test.go` (신규, 이름은 run 단계 재량) | 스캐너 + 여섯 방향 시연 |

### C.6 템플릿 미러와 방출 (기계적 — 마지막)

C.1·C.2의 편집은 `internal/template/templates/` 아래 같은 경로에 미러가 있다. C.3에서 교체로 결정된 스킬 파일도 미러가 있다. 미러 정합성은 iter1 감사가 재측정해 **누락 없음**을 확인했다(D14).

`.codex/agents/moai/manager-lead.toml`은 **손으로 고치지 않는다** — `make agents-emit`이 `.claude/agents/moai/*.md` 미러에서 방출한다. 편집 후 `make agents-emit` → `make agents-emit-check` → `make build` 순서를 지킨다.

## D. 가드 결정 — 만든다

**결정: 기계적 가드를 만든다.** 근거는 이것이 doctrine 카드이고, 강제 장치가 없으면 t241류(아무도 집행하지 않는 규칙)가 되기 때문이다. 이 저장소에는 그 위험이 가설이 아니라 실측이다 — 오늘 **124개 문서**가 구체 인용을 담고 있다(spec.md §1.1).

### D.1 무엇을 스캔하나

- **범위**: `.claude/rules/`, `.claude/agents/`, `.claude/output-styles/`, `.claude/skills/` 아래 `.md`.
- **제외**: `.moai/reports/**`, `.moai/specs/**` — 그것이 spec.md §5의 부채이고, 여기 넣으면 가드가 첫날부터 빨갛다.
- **양쪽 트리**: 저장소 루트 사본 **과** `internal/template/templates/` 미러.

### D.2 하한은 스캔 모집단에서 도출한다 [수리됨]

실측(이 트리, HEAD `b64043481`) — **양쪽 명령이 같은 4개 하위트리를 열거한다**:

```
find .claude/rules .claude/agents .claude/output-styles .claude/skills -name '*.md' -type f | wc -l   # 363
T=internal/template/templates/.claude
find $T/rules $T/agents $T/output-styles $T/skills -name '*.md' -type f | wc -l                        # 338
```

하한은 **트리별 300**이고, 집계 합계가 아니라 **각 트리에 따로** 건다(AC-ECC-010).

초판의 하한 7은 세 겹으로 잘못됐다: 실제 모집단(합 701)보다 두 자릿수 작아 범위가 한 하위 디렉터리로 붕괴해도 걸리지 않았고, 근거로 든 “오늘 위반 파일 7개”는 저장소 루트만 센 값이며(두 트리 기준 14), 애초에 **위반 파일 수와 스캔 파일 수는 서로 다른 모집단**이다. 수리 후 위반은 0이 되므로 7은 어떤 관측량과도 연결되지 않는다.

**미러 명령도 초판에서 바뀌었다.** 초판은 `templates/.claude` 전체를 훑어 340을 냈는데, §D.1이 정한 스캔 범위는 양쪽 트리 모두 4개 하위트리다. 넓은 모집단에서 도출한 하한은 라벨과 명령이 어긋난 것이고(iter2 감사 N3), 그것이 정확히 이 SPEC이 입법하는 단위 정합 문제다. 범위 밖 2개는 `templates/.claude/loop.md`, `templates/.claude/commands/moai/todo.md`.

### D.2.1 하한이 원리상 못 하는 일 — 하위트리 소실 [수리됨]

하한 300은 루트에 63, 미러에 38의 여유를 남기고, **그 여유 안에 작은 하위트리 둘이 통째로 들어간다**:

| 빠지는 하위트리 | 루트 잔여 | 미러 잔여 | 하한 통과? |
|---|---|---|---|
| `.claude/agents` (루트 21 / 미러 11) | 342 | 327 | **통과한다** |
| `.claude/output-styles` (양쪽 3) | 360 | 335 | **통과한다** |

하필 그 둘이 §1.3의 `manager-lead.md`와 AC-ECC-008이 존재하는 이유인 배너 3지점을 담는다. 범위에서 빠지면 스캔 342건·위반 0으로 초록이다.

**하한을 올려서 막을 수 있는 문제가 아니다** — 집계 수는 어느 하위트리가 빠졌는지 원리상 구별하지 못한다. AC-ECC-015가 하위트리 **집합 상등**을 따로 단언하고(부분집합이 아니라 상등), `agents`를 빼는 뮤테이션이 **하한을 통과한 채로** 그 단언에서 실패하는 것을 보인다 — 통과·실패가 함께 움직이면 그 단언은 하한의 재서술일 뿐이므로, 독립성 시연이 곧 그 단언이 값한다는 증거다.

### D.3 왜 양쪽 트리이고, 방문 사실을 단언하는가 [수리됨]

기존 `internal/template/gitignore_agents_mirror_test.go`는 45행 `filepath.Join("templates", ".gitignore")`와 임베드 사본만 읽고 저장소 루트를 읽는 코드가 없다. 그 형태에서는 **저장소 루트에서 규칙을 지워도 테스트가 초록**이다.

초판은 이 실패 형태를 이 절에 정확히 서술해 놓고도 그것을 잡는 AC를 세우지 않았다(감사 D2). 그래서 AC-ECC-015가 넷을 단언한다: 방문한 **트리 루트 목록**이 둘 다 포함하는지, 트리마다 방문한 **하위트리 집합이 4개와 상등**인지(§D.2.1), 트리별 스캔 수가 각각 하한을 넘는지, 두 사본이 이 항목에 관해 일치하는지.

방문 단언 자체도 시연 없이는 공허하므로 두 방향의 뮤테이션을 함께 보인다: 트리 루트 한쪽을 빼면 첫째가 실패하고, 하위트리 하나를 빼면 둘째가 **하한을 통과한 채로** 실패한다.

### D.4 허용목록의 단위와 상한 [수리됨]

carve-out(§C.3에서 확정될 기계 소비 문맥)은 허용목록으로 표현하고, 두 가지를 단언한다.

1. **단위는 파일 + 정확한 리터럴.** 파일 경로만으로 된 항목은 금지한다. 파일 단위 면제였다면 §C.3의 세 파일이 통째로 사각지대가 되어, 나중에 그 파일에 진짜 위반 인용이 들어와도 가드가 보지 못한다(저장소 루트 위반 대상 7건 중 3건, 43%).
2. **항목 수를 상수로 단언.** 목록이 조용히 늘어나면 깨진다 — 위반을 지우는 가장 싼 길(“허용목록에 한 줄 추가”)에 값을 매기는 장치다.

둘이 서로를 보완한다. 개수만으로는 항목 하나가 파일 전체를 삼키는 것이 통과하고, 단위만으로는 목록이 길어지는 것이 통과한다.

### D.5 여섯 방향 시연 [HARD]

한 방향만 보이는 시연은 규칙이 꺼진 상태와 구별되지 않는다.

1. **통과 방향** — C.1·C.2 편집 후 실제 트리에서, **트리별로** 스캔 수 ≥ 300이고 위반 0.
2. **합성 뮤테이션** — 새 방식 인용 한 줄과 옛 방식 인용 한 줄이 든 임시 픽스처에서 옛 방식만 잡힌다.
3. **실물 뮤테이션** — 수리 **이전** `agent-common-protocol.md:268` 문장을 픽스처 리터럴로 박아 스캐너가 그것을 잡는다. **이 항목이 진짜 RED다** — 오늘 트리에 실재하는 문장이기 때문이다.
4. **트리 방문 뮤테이션** — 한쪽 트리 루트를 빼면 AC-ECC-015의 트리 방문 단언이 실패한다.
5. **하위트리 방문 뮤테이션** — `agents` 하위트리를 빼면 AC-ECC-015의 하위트리 상등 단언이 실패하고, **같은 실행에서 하한 단언은 통과한다**(342 ≥ 300). 두 단언의 독립성을 보이는 것이 이 시연의 목적이다.
6. **허용목록 단위 뮤테이션** — 리터럴이 빈 항목을 검증자에 먹이면 거부된다(AC-ECC-013 #1).

### D.6 가드가 지키지 못하는 것 (알려진 잔여 위험)

- `.codex/agents/moai/manager-lead.toml`은 `.toml`이라 범위 밖이다. 오늘 그 파일에 `canonical persistence location` 문장이 있고, M6의 `make agents-emit`이 전이적으로 고치지만 **가드가 지키는 것은 아니다**.
- `internal/cli/mcp_glm.go:110`의 **코드 주석**이 `.moai/state/verify/t225/…`를 측정 근거로 인용한다 — 이 SPEC이 없애려는 결함의 살아 있는 실례인데 `.go`는 범위 밖이다. spec.md §5의 후속 카드 후보.

## E. 자체 검증

- 규칙 본문 교체 후 §C.1·C.2 대상 파일에서 남은 `.moai/state/verify` 출현이 전부 허용목록 문맥인지 확인.
- 가드 6방향 실행(`go test ./internal/template/...`).
- carve-out이 기제를 깨지 않는지: `go test ./internal/verify/... ./internal/web/...`.
- `make agents-emit-check` (커밋된 `.toml`과 방출 결과 일치).
- `make build` 후 템플릿 중립성 — 규칙 본문에 SPEC ID·카드 id·내부 날짜를 넣지 않는다. `.moai/reports/<card-id>/`는 **경로 형태**라 중립이지만, 특정 카드 id 예시를 미러에 넣으면 위반이다.

## F. 마일스톤

| # | 내용 | 산출물 |
|---|---|---|
| M1 | C.1 규칙 본문 2파일 교체 + carve-out 2건 문구 확정 | 편집 diff |
| M2 | C.3 판별식 적용, 세 스킬 파일 결정 기록 + **Tier 재확인** | progress.md §E.2 판정 3줄 + Tier 판정 1줄 |
| M3 | C.2 manager-lead.md + output-style 배너 정정 | 편집 diff |
| M4 | D. 가드 + 여섯 방향 시연 | 테스트 파일, `go test` 출력 |
| M5 | C.4 `.gitignore` 2파일 (각 한 줄) | 편집 diff |
| M6 | C.6 미러 + `make agents-emit` + `make build` | `agents-emit-check` rc=0 |

M1이 먼저인 이유: carve-out 문구가 확정돼야 M2의 판별식이 적용 가능하고, M4의 허용목록이 무엇을 담을지 정해진다. M2가 둘째인 이유는 그 결정이 Tier를 되물을 수 있어 뒤로 미룰수록 비싸지기 때문이다. M5·M6은 앞의 결정에 의존하지 않으므로 마지막에 둔다.

## G. 위험과 안티패턴

- **공허한 초록.** 하한을 모집단에서 도출하지 않으면 범위 붕괴가 통과한다 — 초판이 정확히 이 실수를 했다(§D.2).
- **한쪽 트리만 보는 가드.** 이 저장소에 실재하는 형태다(§D.3).
- **허용목록으로 도피.** 개수 단언 + 단위 단언 둘 다 필요하다(§D.4).
- **반출 비대.** 인용 넓이 상한 없이 반출 의무만 세우면 13 MB가 `.moai/reports/`로 옮겨 간다. REQ-ECC-004와 REQ-ECC-005가 **같은 상한**을 가리켜야 한다.
- **소급 확대.** 124개 문서를 “겸사겸사” 고치려는 유혹. 범위 밖이고 후속 카드다.
- **미러 누락.** `.claude/` 아래 편집이 `internal/template/templates/` 미러 없이 커밋되면 다음 `moai update`가 되돌린다.
- **`.gitkeep`을 “형태 일관성”으로 넣기.** 텔레메트리 형제를 그대로 따라 하면 배포판 전체의 훅 opt-in이 뒤집힌다(§C.4, spec.md §4.2).

## H. 상호 참조

- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — 이 SPEC이 지키려는 귀속 규범. **문자열 `.moai/state/verify`를 담고 있지 않으므로 이 SPEC에서 편집하지 않는다**(grep 확인).
- `.claude/rules/moai/workflow/kanban-dispatch.md` 183행 — `.moai/reports/<card-id>/verdict.md`를 증거 경로로 이미 못 박고 있다. 이 SPEC은 그 규약에 프로토콜을 맞추는 것이지 새 규약을 세우는 것이 아니다.
- `/Users/goos/MoAI/moai-adk-go/.moai/reports/t373/verdict.md` 131-147행 — ignore 판정의 선례.
- `.moai/reports/t375/plan-audit-iter1.md` — iter1 감사 판정(FAIL 0.70), 이 판의 수리 대상.
- `CLAUDE.local.md` §2 Template-First Rule, §2.0 에이전트 정의 3사본 규칙.
