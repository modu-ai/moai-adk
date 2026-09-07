# t475 관측 리포트 — codemaps 최신성 재발 종결 (SPEC-CODEMAPS-REFRESH-002)

**측정 트리**: 워크트리 `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`
**작업 시작 HEAD**: `52f863f36` · **재스탬프 리비전**: `52f863f3666c9ec754253a06b96ed1fe844f1590`(= `merge-base HEAD origin/develop`)
**리드 소비용.** 이 리포트는 **아무 설정도 바꾸지 않는다** — 임계값 40, `gate.yaml`, `DefaultThresholds()`, 접힘 정책, Go 코드 전부 무변경이다(§B.2 / REQ-CM2-012).

정책 규칙 적용 기록: 이 실행은 `.claude/rules/moai/core/verification-claim-integrity.md` §1.1(surface 1·3·4)·§2(baseline 귀속)·§3(5절 형식)과 `.claude/rules/moai/development/verification-completeness.md` §1.1(빈 집합 위의 통과)·§2.1(RED-now 4요소)·§4(증거 고정)를 적용해 판정했다.

---

## 관측 ① — 임계 40 대비 값의 거동, 그리고 그 값이 쌓인 **속도**

### 값

| 시점 | codemaps `described-source-diff` | verdict |
|---|---|---|
| 작업 전 (`52f863f36`, 앵커 `25a3212a9`) | **64** / 임계 40 | stale |
| 재스탬프 후 (앵커 `52f863f36`) | **0** / 임계 40 | fresh |

### 속도 — 앵커 귀속이 재료다

임계값 판단에 필요한 것은 값 자체가 아니라 **그 값이 얼마 만에 쌓였는가**이며, 그 답은 앵커를 **누가 언제** 찍었는지에서 나온다. `provenance.json` 직독(작업 전):

```
commit_sha:   25a3212a93b4c811cbb22e3c0b34d43571fa65b4
tree_root:    /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t476
generated_at: 2026-09-03T18:18:34Z
```

**앵커를 찍은 것은 SPEC-CODEMAPS-REFRESH-001(2026-09-02)이 아니라 워크트리 t476 이 2026-09-03T18:18:34Z 에 찍은 스탬프다.** 즉 **64 는 약 5일치 누적분**이다(2026-09-03 → 2026-09-08).

이 속도는 SPEC-GRAPH-FRESHNESS-CADENCE-001 이 산출한 "corrected-40 이 약 1.6일에 교차"와 정합한다 — 5일이면 임계를 여러 번 넘고도 남는다. **즉 64 는 이상값이 아니라 이 리포지터리의 정상 변경 속도가 만드는 예상 값이다.**

### 리드가 읽을 함의 (판단은 운영자 몫 — 이 카드는 아무것도 바꾸지 않는다)

1. **임계 40 은 "재스탬프 주기가 1.6일보다 길면 상시 적색"을 뜻한다.** CADENCE-001 이 통합 축 재유도를 근거로 40 을 유지 판정했으므로 이 카드는 그 판정을 재개하지 않는다. 다만 **관측된 사실은 임계값의 문제가 아니라 재스탬프 주기의 문제**라는 쪽을 가리킨다.
2. **재스탬프는 사람이 해야 풀린다.** 누적 지표이므로 큰 통합 하나가 이후 모든 측정을 붉게 만들고, 게이트는 스스로 회복하지 않는다.
3. **게이트 녹색과 문서 정확성은 별개 축이다.** 재스탬프만으로 `value=0` 이 되므로, 이 카드가 실제로 한 일(누락 15단위 편입 · 7구간 재기술 · 3항목 정확성 검증)은 게이트가 측정하지 않는다. 그래서 정확성 층 증거를 따로 수출했다(`codemaps-accuracy-verification.md` 7섹션).

### 부수 관측 — 게이트 종결의 종료 코드

재스탬프 후 `./bin/moai graph check` 의 **종료 코드는 여전히 1** 이다. **stale 계층은 하나도 없다** — 원인은 mx-index / edges 의 `verdict=absent` 이며, `CheckResult.Failed()`(`internal/graph/check.go:142-149`)가 `VerdictFresh` 가 아닌 **모든** verdict 를 실패로 세기 때문이다. 신규 워크트리에서 두 계층이 absent 인 것은 예상 상태이고 AC-CM2-010 이 그것을 합격 저해 요인이 아니라고 명시한다. **판정은 계층 verdict 로 하고 종료 코드로 하지 않는다** — 종료 코드 0 을 종결 조건으로 읽으면 이 카드는 신규 워크트리에서 영원히 닫히지 않는다.

---

## 관측 ② — 후보 20개 전수의 fold / omission 분류 요약

후보 집합은 손으로 열거하지 않고 spec.md §A.3(a) 의 규칙이 산출했다: **A층 5 + B층 14 + C층 1 = 20**(히트-0 패키지 48 기준, 저작 시점 실측과 동일).

| 분류 | 수 | 단위 |
|---|---|---|
| **fold** | **5** | `internal/core/git` · `internal/cli/doctor_hook_delivery.go` · `internal/hook/quality/step_git_env.go` · `internal/kanban/prlink_landedref.go` · `internal/web/fieldsets_codex_templ.go` |
| **omission** | **15** | `internal/chain` · `internal/stateanchor` · `internal/settings/yamlpatch` · `internal/template/{agentemit,commandemit,published_skills.go,skill_mirror_repair.go}` · `internal/cli/{skills.go,codex_skills_disable.go,codex_skills_prune.go,integration_settings_drift.go,update_mirror_heal.go}` · `internal/kanban/settings_drift.go` · `internal/statusline/state_anchor.go` · `internal/web/codexmirror.go` |

omission 15개는 전부 편입돼 재생성 후 `grep -rl` 이 1개 이상을 반환한다. fold 5개는 산문 무변경이 규정이므로 히트 0 을 유지한다.

**패턴 — omission 15개는 흩어져 있지 않고 네 덩어리다.** 이것이 이 카드가 실제로 발견한 것이다:

| 덩어리 | 단위 수 | 앵커 이후 자란 역량 |
|---|---|---|
| **codex 발행 계열** | 5 | 에이전트·명령 정의를 codex 아티팩트로 내보내는 두 방출기와, 그 산물을 배포에서 보호·복구하는 두 seam, 그리고 codex 스킬 노출 제어 |
| **상태 앵커 계열** | 2 | `.moai/state` 를 읽고 쓰는 모든 표면이 공유하는 단일 루트 해석 seam과 statusline 어댑터 |
| **설정 드리프트 계열** | 2 | 병합 창 진입 전 tracked `settings.json` 워킹 사본 드리프트 단정(도메인 절반 + CLI 절반) |
| **나머지** | 6 | 워크트리 계보 원장, YAML 보존 쓰기 경로, 미러 복구, codex 미러 탭 등 |

**`internal/harness/*` 하위 12개는 어느 쪽으로도 갈리지 않았다** — 후보 집합에 들어오지 않았기 때문이다(앵커 이후 무변경). 관측 ③ 의 이관 목록에 전수가 있다. 그 12개가 부모 `internal/harness`(히트 4)로 접힐 개연성이 높다는 것은 **가설이며, 이 카드는 그것을 판정하지 않았다** — 판정하려면 12회의 책임 질문이 더 필요하고, 부모 히트 수는 근거가 아니다.

### 리드가 알아야 할 자기 신고 1건

첫 통과의 재생성이 **fold 5개 전부에 대해서도 산문을 새로 썼다.** §A.3(a1) 은 fold 단위의 산문을 바꾸지 말라고 규정하므로 이는 판정을 편입 허가로 오용한 형태다. §⑥(패키지 대조)에서 **후보 중 히트-0 잔여가 0개**로 나온 것이 그 신호였고, 되돌린 뒤 `internal/core/git` 하나가 잔여로 복귀했다(파일 입도 fold 4개는 패키지 census 에 원래 나타나지 않는다).

**어떤 AC 도 "fold 단위가 여전히 히트 0 인가"를 직접 묻지 않는다.** AC-CM2-007 의 전제로만 간접적으로 걸리므로, 되돌리지 않았어도 12개 AC 는 전부 통과했을 것이다. 후속 카드가 이 간극을 닫을지는 리드 판단이다.

---

## 관측 ③ — 후보에서 제외된 잔여 42개 패키지 전수 목록 (범위 밖 **이관**)

히트-0 패키지 48개 중 **앵커 이후 변경이 없어 후보가 되지 않은 42개**다. 이 카드의 방아쇠는 앵커 이후의 드리프트이므로 이들은 인과가 다르다 — **실재하는 부채이되 이 붉음의 원인이 아니다.** 판정하지 않았고, 버리지도 않는다. 리드가 후속 카드를 정할 재료다.

그중 **12개가 `internal/harness/*` 하위**다(굵게 표시).

```
internal/cli/agentlint            internal/hook/handoff
internal/cli/harness              internal/hook/memo
internal/cli/printer              internal/hook/memo/taxonomy
internal/cli/taskledger           internal/hook/perf
internal/cli/uikit                internal/hook/testutil
internal/cli/worktree             internal/hook/trace
internal/config/toolpolicy        internal/lsp/aggregator
internal/core/project             internal/lsp/cache
internal/git/convention           internal/lsp/core
internal/graph/symbol             internal/lsp/gopls
**internal/harness/capture**      internal/lsp/hook
**internal/harness/cluster**      internal/lsp/subprocess
**internal/harness/curator**      internal/navigator/detect
**internal/harness/delegationmap** internal/navigator/fix
**internal/harness/proposalgen**  internal/navigator/route
**internal/harness/router**       internal/navigator/sync
**internal/harness/routing**      internal/navigator/tiers
**internal/harness/safety**       internal/runtime/gobin
**internal/harness/seeds**        internal/settings/agentfm
**internal/harness/throttle**     internal/tui/internal
**internal/harness/tier**
**internal/harness/v4manifest**
```

**42개 전수 = 12(harness) + 30(그 외).** 이 목록이 A층 변경 필터를 정당화하는 이관 조건이다 — 범위 밖으로 나간 것이 무엇인지 읽을 수 있어야 좁힘이 손실이 아니라 이관이 된다.

### 후속 카드 재료로서의 성질 (판정 아님, 관측)

- **`internal/harness/*` 12개와 `internal/lsp/*` 5개, `internal/navigator/*` 5개는 각각 한 부모 아래 묶음**이다. 22개가 세 부모에 몰려 있으므로, 접힘 정책 질문은 42회가 아니라 사실상 **세 번의 질문**으로 줄어들 개연성이 있다. (개연성이지 판정이 아니다 — 실제 판정은 단위별 책임 질문이다.)
- `internal/git/convention` 은 관측 ④ 와 얽힌다(아래).

---

## 관측 ④ (부수) — 재생성이 드러낸 사실 3건

관측 항목이 아니라 **재기술 과정에서 트리를 다시 재며 나온 것**이다. 어느 것도 이 카드의 수리 대상이 아니었고, 셋 다 문서 본문에 실었다.

1. **직전 판의 "패키지 단위 엣지 1638" 은 재현되지 않는다.** 이 판의 명령은 **345** 를 낸다. 직전 판의 명령 인용이 생략형(`'{{range .Imports}}...'`)이라 무엇을 셌는지 복원할 수 없다. 최상위 집계(205 → 208)는 정합하므로 값 하나만의 문제로 보인다.
2. **"`internal/hook/session_start.go`(67KB)가 트리 최대 비테스트 Go 파일" 은 더 이상 사실이 아니다.** 상위 둘은 `templ` **생성** 산물(`fieldsets_templ.go` 168KB, `screens_templ.go` 121KB)이고 `session_start.go` 는 61KB 다. 크기 순위를 읽을 때 생성/손저작을 갈라야 한다.
3. **`internal/git`(루트 패키지)의 비테스트 import 가 0이다.** 실제로 import 되는 것은 하위 `internal/git/convention` 하나뿐이며, 최상위 집계 fan-in 1 은 전부 그것이다. 소비자가 `core/git` 로 직접 내려가며 중간 계층만 남은 모양으로 읽히지만 **이 카드가 확인하지 않았다** — 관측일 뿐이다. `internal/git/convention` 이 관측 ③ 의 42개 안에 있는 것과 함께 읽을 재료다.

`docs-truth.md` 손 갱신에서도 두 건이 나왔다 — §4.1 의 verb 그룹 분류가 `moai --help` 렌더에 존재하지 않는 손 분류였고 라이브 루트 명령 `skills` 를 빠뜨리고 있었다. 렌더의 실제 4개 그룹으로 교체했다.

---

## 보고하지 않는 것

**`provenance.json` 의 `tree_root` 가 외래 워크트리를 가리키는 것은 관측 항목이 아니다.** `internal/graph/check.go:341-346` 의 주석이 그 설계 의도를 직접 적는다 — codemaps 는 tracked 아티팩트라 스탬프를 찍은 트리 말고는 어디서 봐도 `tree_root` 가 외래인 것이 기본 상태이며, 하드 매칭하면 게이트가 모든 곳에서 무력해진다. 코드도 같은 말을 한다: `pv.TreeRoot != projectRoot` 비교는 `checkMXIndex` 경로에만 있고 `checkCodemaps` 경로에는 없다. 닫힌 조사이므로 리드에게 올리지 않는다.
