---
title: moai graph
weight: 16
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

코드베이스의 관계를 하나의 산출물로 모아 **거꾸로 질문**에 답하는 도구입니다. "이 패키지를 고치면 어디까지 흔들리는가", "이 SPEC은 실제 코드와 이어져 있는가" — 이런 질문은 grep으로는 답이 안 나오고, 관계가 모여 있어야 답합니다.

{{< callout type="info" >}}
**한 줄 요약**: `moai graph build`가 codemaps·@MX 태그·SPEC·리포트에 흩어진 관계를 `.moai/project/graph/edges.jsonl` 한 파일로 모으고, `moai graph query`가 그 파일에 역방향 질의를 합니다.
{{< /callout >}}

## 왜 필요한가

MoAI-ADK는 관계 정보를 여러 층에 이미 갖고 있습니다 — codemaps의 임포트 그래프, 코드 안의 `@MX:SPEC` 태그, SPEC 문서의 의존 선언, 리포트의 마일스톤 기록. 문제는 이 층들이 저마다 다른 파일에 흩어져 있다는 점입니다. "이 코드를 고치면 어느 SPEC이 영향을 받는가"를 물으려면 임포트 방향(@MX:SPEC 태그가 있는 파일을 임포트하는 파일)과 SPEC-의존 방향을 **같은 그래프에서** 거꾸로 따라가야 합니다. edges.jsonl은 그 하나의 그래프입니다.

## moai graph build

```bash
$ moai graph build
```

임포트 에지 · `@MX:SPEC` 연결 · SPEC 간 의존을 모아 `.moai/project/graph/edges.jsonl`에 기록합니다. 같은 git HEAD에서 두 번 돌리면 같은 내용이 나오도록 결정적으로 작동합니다. 질의는 언제나 이 산출물을 읽으므로, **질의 전에 먼저 build를 돌려** 두어야 합니다.

문서 계층 위에 코드에서 직접 뽑은 에지를 얹습니다 — 함수 호출 에지(code-call)와 임포트 에지(code-import)이며, 기존 문서 에지는 하나도 바뀌지 않습니다. 임포트 대상은 go.mod 모듈 경로를 덜어내 저장소 로컬 패키지로 정규화해서 codemaps의 임포트 그래프와 같은 영역을 가리키게 하고, 호출 해석이 어느 수준(등급)에서 이뤄졌는지 16개 언어 전부에 대해 공표합니다. 두 계층이 같은 관계에 대해 다르게 말하면 어느 한쪽을 버리지 않고 `disagrees_with` 표시와 함께 둘 다 남기며, `--all-disagreements`는 기본 모드에서 숨긴 방향(코드는 발견했는데 문서가 침묵한 로컬 의존)까지 표시합니다.

## moai graph query

한 번의 호출에 셀렉터를 **정확히 하나**만 줍니다.

| 셀렉터 | 질문 | 답 |
|--------|------|-----|
| `--callers <노드>` | 이 패키지/SPEC을 직접 의존하는 대상은? | 역방향 이웃 — 임포트하는 패키지, 의존하는 SPEC, `@MX:SPEC` 태그가 붙은 코드 파일 |
| `--blast <노드>` | 여기서 고치면 어디까지 흔들리는가? | 역방향 에지를 넓게 훑은(BFS) 영향 반경. `@MX:SPEC` 에지는 양방향으로 전파돼 코드 파일이 구현하는 SPEC까지 닿습니다 |
| `--fanin [--limit N]` | 가장 많이 쓰이는 패키지는? | 임포트 팬인 순위 — @MX:DEBT 팬인 질의의 대용품(아직 태그 종류별 에지는 없음) |
| `--specs-no-code` | 코드와 이어지지 않은 SPEC은? | edges.jsonl에 `@MX:SPEC` 에지가 0개인 SPEC 목록 |
| `--milestones-no-card` | 카드 없이 지나간 마일스톤은? | 카드 교차검사 행이 카드를 주장하지 않거나, 주장한 카드가 살아 있는 백로그 큐에 없는 마일스톤 |

```bash
$ moai graph query --callers SPEC-FOO-001
$ moai graph query --blast internal/config
$ moai graph query --fanin --limit 20
$ moai graph query --specs-no-code
$ moai graph query --milestones-no-card
```

`--edges <경로>`로 다른 edges.jsonl을 가리키거나, 루트 인자로 다른 프로젝트 루트를 지정할 수 있습니다.

질의 전에 기계적인 층(@MX 인덱스 · edges.jsonl)이 오래했으면 먼저 갱신하고 답합니다. 내용 해시가 달라진 파일만 다시 파싱하므로 커밋하지 않은 편집도 답에 반영되고, 갱신 비용이 gate.yaml의 `update_budget_ms`(기본 2000ms)를 넘으면 경고만 하고 답은 계속합니다. 답을 낸 트리 루트와 커밋(또는 dirty 지문)이 항상 stderr에 함께 찍히므로, 어떤 트리의 답인지 헷갈릴 일이 없습니다.

## moai graph check

```bash
$ moai graph check
```

그래프의 세 층 — codemaps · @MX 인덱스 · edges.jsonl — 이 코드를 제대로 따라가고 있는지 층마다 자기 지표로 측정해 `fresh` / `stale` / `absent`로 판정합니다. codemaps는 스탬프된 생성 커밋 이후 달라진 묘사 대상 파일 수(되돌린 변경은 0으로 셉니다), @MX 인덱스는 내용 해시가 달라진 파일 수, edges.jsonl은 소스 지문 불일치를 봅니다.

각 산출물은 provenance 블록으로 어느 트리·커밋의 산물인지 밝힙니다. 블록이 없으면 판정은 `absent` — 판단할 수 없음을 fresh로 속이지 않고, absent도 실패입니다. 새 워크트리에는 해당 산출물이 애초에 없으므로 전부 absent로 보고됩니다. 종료 코드는 0(모두 fresh) · 1(stale 또는 absent) · 2(시스템 오류)이며, 사전 커밋 품질 게이트의 그래프 신선도 단계와 CI graph-freshness 작업이 이 값을 그대로 소비합니다. 문턱값은 gate.yaml의 `graph_freshness` 섹션에서 조정합니다.

mtime은 어디에서도 읽지 않습니다. 새 체크아웃은 모든 mtime을 초기화하므로 mtime 기반 지표는 방금 재생성된 것으로 오판합니다 — 그래서 모든 지표는 내용 해시, git diff, 지문뿐입니다.

## moai graph stamp codemaps

```bash
$ moai graph stamp codemaps
```

codemaps를 다시 생성한 다음 마지막 단계로 실행합니다. 문서 내용은 `/moai codemaps`가 다듬지만, 그 내용이 **어떤 트리 상태를 묘사하는지**는 이 명령이 `provenance.json`으로 기록합니다. `moai graph check`가 codemaps 층을 판정하는 근거가 이 기록입니다.

## 두 셀렉터를 위한 주의문

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--specs-no-code`**: "미연결"은 "미구현"이 아닙니다. 대부분의 SPEC은 문서·규칙·하네스를 낼 뿐 코드가 없어도 완성입니다. 이 목록은 결함 목록이 아니라 커버리지 지도로 읽으세요.
{{< /callout >}}

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--milestones-no-card`**: 백로그 큐는 카드가 끝나면(`done`) 행을 지웁니다. 그래서 "큐에 없음"은 **끝난 카드와 아예 발급 안 된 카드를 함께** 묶습니다. 각 항목은 `git log --oneline --grep 'merge: tNN'`으로 판별하세요 — 커밋이 있으면 통과, 없으면 새 카드 후보입니다. grep 0건도 "일을 안 했다"는 뜻은 아닙니다. 카드가 새 id로 재발행되었을 수 있으니 새 카드를 끊기 전에 계보를 확인하세요.
{{< /callout >}}

## 관련 문서

- [칸반 모드](/ko/advanced/kanban-mode) — 마일스톤-카드 교차검사가 지켜보는 카드 흐름
- [`/moai mx`](/ko/utility-commands/moai-mx) — @MX 태그와 `@MX:SPEC` 연결의 원본
- [Navigator](/ko/core-concepts/navigator) — 설계 결정·SPEC·심볼을 묶는 또 하나의 그래프(nav-graph.json)
