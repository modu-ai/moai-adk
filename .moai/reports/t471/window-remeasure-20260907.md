# t471 — 통합 창 재측정 (lane-3, 2026-09-07)

리드가 lane-4 의 t488 창 반납 뒤 lane-3 을 지명했다. sync 는 창 **전에** 끝나 있었다
(`94a940d03` sync 산출물 + 3-phase close, `903f5a6af` sync_commit_sha 백필, SPEC
`completed`) — t342 선례(run 만 닫고 병합해 창을 다시 받은 건)를 피한 형태다.

## 흡수

| 항목 | 값 |
| --- | --- |
| 흡수 대상 (로컬 develop) | `dfe25bc09` — t488 착지본. 창 잡은 직후 직접 읽음 |
| 흡수 전 카드 tip | `903f5a6af` (ahead 3, dirty 0) |
| 흡수 병합 커밋 | `3c9021ea9` |
| 충돌 | 1건 — `CHANGELOG.md` |

`origin/develop` 은 `615d18c1f` 로 로컬보다 뒤에 있다. 흡수 대상은 로컬 develop 이다.

### 충돌 해소 — 양쪽 다 살렸다

`[Unreleased] > ### Added` 의 첫 항목 자리를 두 카드가 동시에 차지했다: 내 t471 항목
(HEAD 쪽)과 t488 의 `SPEC-PREMERGE-SETTINGS-DRIFT-001` 항목(develop 쪽). 한쪽을 고르는
충돌이 아니라 **나란히 놓이면 되는** 충돌이라, 마커 3줄만 지우고 두 항목을 모두 남겼다.
내용 편집은 0 — 어느 항목의 문장도 고치지 않았다.

## Claim

t471 이 always-loaded 표면에 더하는 `+141` 토큰을 얹고도 예산 가드는 GREEN 을 유지한다.
그리고 그 141 은 이 카드가 `kanban-dispatch.md` 에 더한 분량 전부이며, 다른 always-loaded
파일은 하나도 건드리지 않았다.

## Evidence — 라운드 1 (`3c9021ea9`)

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v -count=1
    token_budget_guard_test.go:69: always-loaded surface = 76775 tokens (budget 77600, headroom 825, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
ok  github.com/modu-ai/moai-adk/internal/config	0.450s

$ go test ./internal/config/... ./internal/template/... -count=1
ok  github.com/modu-ai/moai-adk/internal/config	3.543s
ok  github.com/modu-ai/moai-adk/internal/config/atomicfile	0.442s
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy	1.008s
ok  github.com/modu-ai/moai-adk/internal/template	26.576s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit	1.450s

$ make build   → rc=0   (agents-emit-check 선행 포함)
$ git status --short   → 빈 출력
```

### 141 의 귀속 — 산술이 아니라 실측

리드가 develop `dfe25bc09` 에서 잰 값은 `headroom 966` 이다. **그 값을 인용하지 않고**
내 병합 트리에서 `825` 를 다시 쟀다. 차 141 을 다음 두 측정으로 이 카드에 귀속시킨다:

```
$ git show develop:.claude/rules/moai/workflow/kanban-dispatch.md | wc -c   → 34580
$ git show HEAD:.claude/rules/moai/workflow/kanban-dispatch.md    | wc -c   → 35144
                                                            증가 564 B

$ git diff --name-only develop HEAD -- .claude/rules .claude/output-styles
.claude/rules/moai/workflow/kanban-dispatch-detail.md      ← paths: 제한 → 측정 표면 밖
.claude/rules/moai/workflow/kanban-dispatch.md             ← always-loaded
```

`estimateTokens` 는 `len/4` 이므로 564 B = 141 토큰이고, 966 − 141 = 825 가 실측치와
일치한다. **always-loaded 표면에서 변한 파일은 `kanban-dispatch.md` 하나뿐이므로**
차 141 은 잔여 없이 이 카드 몫이다. 항목 수는 17 로 불변 — 새 항목이 표면에 들어오지 않았다.

## Baseline-attribution

세 측정 모두 이 워크트리(`WT-lead-bottleneck`)의 흡수 병합 커밋 `3c9021ea9` 에서, 위에
적힌 명령 그대로 이번 창에 실행해 관측한 출력이다. 라운드 2 는 이 파일을 커밋한 tip 에서
같은 명령을 다시 돌려 값이 불변임을 확인한 뒤 develop 에 병합한다 — 재측정 트리와 병합
트리를 같게 두기 위해서다. 라운드 2 결과는 창 완료 보고에 싣는다(자기 자신을 담을 수 없는
파일이라서다).

## 재측정 범위

건드린 표면은 `.claude/rules/**`(+ 템플릿 미러), `.claude/agents/moai/manager-lead.md`
(+ 미러 + `make agents-emit` 이 재생성한 `.codex/.../manager-lead.toml`), `catalog.yaml`
해시 1줄이다. 그 표면을 읽는 패키지가 예산 가드(`internal/config`)와 템플릿 임베드·에밋
(`internal/template`, `internal/template/agentemit`)이므로 범위는 이 둘이다. Go 소스는
건드리지 않았다 — 유일한 비-마크다운 diff 는 `catalog.yaml` 의 기계 재생성 해시 줄이다.

## Gaps

- **예산 적자 상환은 이 카드 몫이 아니다.** 카드 판정 당시 `headroom -264` (FAIL) 였고,
  이 창에서 GREEN 인 것은 t473 이 그 사이 develop 에 착지해 여유를 회복했기 때문이다.
  **이 카드가 균형을 맞춘 것이 아니라 다른 카드가 메웠다.** 판정서의 `+141` / `+564 B` 는
  그대로 둔다. 리드에 따르면 t492 의 C2 이관이 추가로 이를 덮는다.
- AC-LDP-001 의 감축 목표(리드 턴 도구 배치 수 ≤ 기준선의 50%)는 여전히 **문서 준비
  상태로만 PASS** 다. 실제 감축은 첫 채택 배치의 회차 보고에서 재야 한다.
- CI 는 이 창에서 돌리지 않았다. 레인은 CI 를 직접 요청하지 않으며 판정은 리드 몫이다.
- `internal/cli` 등 나머지 패키지는 돌리지 않았다 — 이 변경이 닿지 않는다.

## Residual-risk

- 여유 825 는 영구적이지 않다. `estimateTokens` 는 `len/4` 근사(±15%)라 상대 증분 감시용
  이며 절대 회계가 아니다.
- `CHANGELOG.md` 충돌을 마커 제거로 해소했으므로 두 항목의 **순서**는 내 항목이 위다.
  의미상 순서 규약이 있다면 이건 도착 순서일 뿐 우선순위가 아니다.
