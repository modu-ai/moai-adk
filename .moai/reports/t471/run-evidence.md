# RUN EVIDENCE: SPEC-LEAD-DEPUTY-001 (card t471)

- **측정 트리**: `.claude/worktrees/t471`, branch `WT-lead-bottleneck`
- **base pin**: `615d18c1f` (origin/develop 흡수 tip — fast-forward, 충돌 0). 흡수 전 HEAD `bdcec58f0`은 이미 origin/develop의 조상이었고 `origin/develop..HEAD = 0`이었다. run 커밋이 이 브랜치의 첫 고유 커밋이다.
- **일자**: 2026-09-06
- 모든 수치는 위 트리에서 이 실행으로 관측한 것이다. 리드 자체 계수(`acceptance.md` §D.0)는 재유도하지 않았다 — 재유도 불가하며, 그 사실은 §D.0이 이미 명시한다.

## Claim

M1–M4 착지: 상주 deputy 채택 의무 + 완료-보고 `RECOMMEND:` 경로 + 회차 보고 위임·분할 교리 + idle 통지 전환이 세 표면(always-loaded stub / lazy companion / 에이전트 정의)에 문면으로 존재하고, 템플릿 미러·C3 방출·depth seal·중립성이 모두 통과한다. **예산 순증 목표(≤ 0)는 달성하지 못했다 — +141 토큰이다.**

## Evidence (명령 + 원문 출력)

### 흡수와 base pin

```
$ git fetch origin develop
$ git rev-parse --short origin/develop      → 615d18c1f
$ git rev-list --count origin/develop..HEAD → 0        (흡수 전, HEAD=bdcec58f0)
$ git rev-list --count HEAD..origin/develop → 67
$ git merge origin/develop                  → fast-forward, HEAD = 615d18c1f, 충돌 0
$ git status --porcelain | wc -l            → 0
```

### AC-LDP-002 — idle 통지 경계 (RED → GREEN)

RED-now (흡수 tip `615d18c1f`, 편집 전):
```
$ grep -c 'scheduling hint' .claude/rules/moai/workflow/kanban-dispatch.md .claude/agents/moai/manager-lead.md
.claude/rules/moai/workflow/kanban-dispatch.md:0
.claude/agents/moai/manager-lead.md:0
exit=1
```
GREEN (편집 후, 같은 트리):
```
$ grep -c 'scheduling hint' .claude/rules/moai/workflow/kanban-dispatch.md .claude/agents/moai/manager-lead.md
.claude/agents/moai/manager-lead.md:1
.claude/rules/moai/workflow/kanban-dispatch.md:1
exit=0
```

**계수기를 맞추려고 토큰을 놓은 것이 아니다.** `kanban-dispatch.md`의 그 한 건은 § The delegation channel is the queue의 **기존 문단** — 「끝났을 때·권한 프롬프트에서·죽었을 때 idle이 된다 / 언제 읽을지만 알려준다」는 세 갈래 조건절을 이미 운반하던 문단 — 에 `cross-session-messaging.md` § An idle notice is a scheduling hint 상호참조를 **부착**한 결과다. 조건절을 옮기지 않았으므로 축약이 아니라 인용이다. `manager-lead.md`의 한 건도 같은 절을 인용하며 세 갈래를 그 자리에 병기한다.

### AC-LDP-003 / 004 / 006 / 007 — 교리 문면 (protocol readiness)

`kanban-dispatch.md` § Deputy dispatch surface에 [HARD] 3건이 신설됐다:
- 상주 spawn 의무 (배치 첫 디스패치 전, UNNAMED, 정확히 1개) — REQ-LDP-001 / AC-LDP-003
- 완료 보고는 `RECOMMEND:` 요약으로 도달, 경로 명명 의무, 리드 자체 증거 판독 불감소 — REQ-LDP-002 / AC-LDP-004
- 회차 보고 측정·초안은 deputy, 단언 수치는 리드, 무귀속 수치는 결함, 회차별 파일 + 인덱스 — REQ-LDP-005·006 / AC-LDP-006·007

`manager-lead.md` § Resident mode에 같은 3건의 온전한 서술 + 근거. `kanban-dispatch-detail.md` § Deputy mode에 기전 4문단(상주 근거 / 완료 보고 판독 / 회차 보고 저작 / 폴링 없는 감시).

**§D.2가 정한 2단 검증 중 1단(protocol readiness)만 이 run에서 검증했다.** 실제 감소 측정(AC-LDP-001)은 첫 채택 배치의 회차 보고 소관이다 — 이 카드가 뒤집을 수 있는 셀이 아니다.

### AC-LDP-005 / 010 — PRESERVE 확인

```
$ grep -c 'DEPUTY-RETAINED-BY-LEAD' .claude/agents/moai/manager-lead.md      → 2
$ grep -c 'Delivery-shape verification' .claude/agents/moai/manager-lead.md  → 1
```
위임 5종 표·보유 6항목 열거·delivery-shape 문단은 diff에서 무변경이다(추가 12행은 전부 신설 § Resident mode). `kanban-dispatch.md`의 기존 [HARD] 2건(권한 없음 / 구조 불변)은 보존했고, 후자는 같은 파일 안에서 이미 [HARD]로 서 있는 세 주장을 재진술하던 문장을 압축했을 뿐 주장 3개를 모두 유지한다.

### AC-LDP-008 — 템플릿 중립성

```
$ grep -rn "t471\|리드 자체 계수\|SPEC-LEAD-DEPUTY\|SPEC-LEAD-DEBOTTLENECK" internal/template/templates/ | wc -l
0
```
편집 전 기준선도 0이었다(같은 명령, 흡수 tip).

### AC-LDP-009 — depth seal · 빌드 · 방출

```
$ go build ./...                                    rc=0
$ GOOS=windows GOARCH=amd64 go build ./...          rc=0
$ go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent' -count=1
ok  github.com/modu-ai/moai-adk/internal/template  0.476s
$ go test ./internal/template/agentemit/... -count=1
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.477s
$ go test ./internal/template/ -count=1
ok  github.com/modu-ai/moai-adk/internal/template  25.535s
$ make agents-emit    → .codex/agents/moai/manager-lead.toml 재생성 (C2→C3, 수편집 0)
$ make build          → catalog.yaml updated successfully (12899 bytes) + bin/moai
```

미러 쌍 동기:
```
$ diff -q <로컬> <템플릿>   → kanban-dispatch.md / kanban-dispatch-detail.md / manager-lead.md 3쌍 모두 동일
```

### 표면 델타 (t492 재측정용)

| 파일 | 예산 계수 | 편집 전 | 편집 후 | 델타 |
|---|---|---|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | **계수됨** | 34,268 B | 34,832 B | **+564 B** |
| `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | 미계수 (`paths:` 한정) | — | — | +8행 |
| `.claude/agents/moai/manager-lead.md` | 미계수 (에이전트 정의) | — | — | +12행 |

**t492가 알아야 할 것**: 변경 절은 `kanban-dispatch.md` § Deputy dispatch surface 하나와 § The delegation channel is the queue의 한 문장이다. §D의 신설 [HARD] 3건이 그 +564 B의 대부분이고, 같은 절 안에서 5종 열거 문장·근거 산문·별도 idle 블록을 덜어 상쇄한 뒤의 값이다.

### 예산 가드 — 전 / 후 (원문)

편집 전 (흡수 tip `615d18c1f`):
```
always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)
--- FAIL: TestAlwaysLoadedTokenBudget (0.01s)
```
편집 후 (같은 트리):
```
always-loaded surface = 77864 tokens (budget 77600, headroom -264, 17 entries)
--- FAIL: TestAlwaysLoadedTokenBudget (0.01s)
```

**목표 미달을 그대로 적는다: 순증 ≤ 0 목표에 대해 실제는 +141 토큰(+564 B)이다.** 가드는 편집 전에도 FAIL이었고 편집 후에도 FAIL이며, 이 카드가 적자를 123 → 264로 **깊게 했다.**

미달 사유: 남은 상쇄 여지가 PRESERVE 대상뿐이었다. 신설 [HARD] 3건은 **-k/-f 리드 세션을 구속**하므로 always-loaded 표면에 있어야 하고(그 세션은 에이전트 정의를 always-loaded로 읽지 않는다), 기존 [HARD] 2건과 보유 6항목 열거는 spec.md §5 PRESERVE 목록이다. 실질을 덜어 숫자를 맞추는 선택 — 「가드는 항목 수를 세지 내용을 세지 않는다」가 경고하는 바로 그 형태 — 은 취하지 않았다.

## Baseline-attribution

모든 편집 전 값은 흡수 tip `615d18c1f`에서 이 실행으로 관측했다. `acceptance.md` §D.0의 RED-now 표는 리드 세션 자체 계수이며 재유도하지 않았다 — §D.0이 재유도 불가를 명시하고 재측정 레시피를 AC-LDP-001에 위임한다. iter-3 판정서의 핀 `109a4615d`는 대상 파일 2개가 그 사이 무변경이므로 이 tip에서도 유효하다.

## Gaps (관측하지 않은 것)

1. **AC-LDP-001 실측 감소** — 채택 후 운용 창이 필요하다. 이 run은 protocol readiness만 검증했다(§D.2가 정한 2단 검증의 1단).
2. **AC-LDP-004/005/006/007의 실적 확인** — 행위 관찰형이라 교리 문면 존재까지만 검증했다. 실제 `RECOMMEND:` 경로·귀속 표기·회차 파일 분할은 첫 채택 배치에서 확인된다.
3. **CI 재측정 없음** — 전부 로컬 darwin 실행이다. darwin/windows 매트릭스 판정은 develop push가 일으키는 실행 몫이다.
4. **iter-3 optional 2건 미수리** — NEW-1(AC-LDP-002 RED-now 셀의 종료코드 토큰), NEW-2(AC-LDP-007 비교 기준 1절). **수용하되 고치지 않았다.** 사유 둘: ① `acceptance.md`를 지금 고치면 plan-audit skip-eligible의 artifact-hash가 깨져 run-gate가 재감사를 요구한다 ② 같은 배치에서 Tier 상한 예외를 연달아 쓰지 않는다(리드 판정). 종료코드는 본 문서가 실측으로 보완했다 — 편집 전 `exit=1`, 편집 후 `exit=0`.
5. **`.moai/reports/lead/` 실물 미접촉** — REQ-LDP-006은 교리만 쓰고 기존 회차 보고 파일은 옮기지 않았다(리드 D2 판정).

## Residual-risk

1. **예산 적자가 깊어졌다(-123 → -264).** 상환은 이 카드 밖이다: t473 브랜치(미푸시, 창 대기)가 `+1,044`를 들고 있고, t492가 구조 최적화를 맡는다. **t473이 착지하기 전에 이 카드가 develop에 들어가면 가드는 여전히 FAIL이며, 그 FAIL은 이 카드가 만든 것이 아니라 이 카드가 141만큼 더한 것이다** — 두 사실을 함께 읽어야 귀속이 맞다.
2. **t492와 같은 파일을 만졌다.** 충돌 방지선은 절 경계뿐이다 — 이 카드는 § Deputy dispatch surface와 § The delegation channel is the queue의 한 문장만 건드렸다. t492가 그 절들을 다이어트 대상으로 잡으면 재측정이 필요하다.
3. **상주 의무의 비용은 측정되지 않았다.** 배치마다 deputy 1개는 쓰이든 안 쓰이든 스폰된다. 그 비용이 절감보다 큰 배치 형태(카드 1장짜리 배치)가 있을 수 있고, 이 run은 그것을 재지 않았다.
4. **`internal/template/catalog.yaml` 1행이 변경됐다** — `make build`가 manager-lead 해시를 재계산한 기계적 결과다. AC-LDP-009의 문언은 `git diff --stat internal/ pkg/ cmd/`가 비어 있을 것을 요구하나, 에이전트 정의를 고치면 이 파일은 반드시 움직인다. 요구의 취지(Go 소스 비접촉)는 충족한다: `git diff --name-only internal/ pkg/ cmd/ | grep '\.go$' | wc -l` → **0**. 문언 대비 편차를 숨기지 않고 여기 남긴다.
