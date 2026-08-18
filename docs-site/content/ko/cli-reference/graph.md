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
