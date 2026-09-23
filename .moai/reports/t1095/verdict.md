# t1095 — docs-site 컨텍스트 창 수치 정정 판정서

- card: t1095 (Tier S, 운영자 배차 2026-09-23)
- branch: `WT-context-window-docs`, base = 로컬 develop `08113ff0f`
- measured by: agent-34 lane

## Claim

docs-site 의 Fable·Sonnet 5 컨텍스트 창 표기(256K / 200K)가 공식 문서와 어긋났고, 24개 파일(4 로케일)에서
공식 값(1M)으로 정정했다. 모델명 `Claude Fable 5` 는 MoAI 코드가 고정한 ID 와 일치하므로 유지한다.

## Evidence — 공식 재조회 (2026-09-23, WebFetch)

| 출처 | 모델 | Context | 상태 |
|---|---|---|---|
| platform.claude.com/docs/en/models/opus-5-5/overview 「How it compares」 | Claude Fable 5.1 | 1M | — |
| 같은 표 | Claude Opus 5.5 | 1M | — |
| 같은 표 | Claude Sonnet 5 | 1M | — |
| 같은 표 | Claude Haiku 4.5 | 200K | — |
| platform.claude.com/docs/en/models/fable-5/overview | Claude Fable 5 (`claude-fable-5`) | 1M tokens | Legacy (Active, legacy) |
| platform.claude.com/docs/en/models/sonnet-5/overview | Claude Sonnet 5 | 1M tokens (Context windows 카드: "1M tokens by default") | Active (latest) |

→ 현행 라인업에서 200K 는 Haiku 4.5 뿐이다. 256K 창은 공식 문서 어디에도 없다.

## 모델명 판정 (Fable 5 vs 5.1)

- 코드: `internal/template/model_policy.go:80` `"fable": "claude-fable-5"` — MoAI 의 `fable` 별칭은 Fable 5 로 해석된다.
- 공식: Fable 5 는 legacy 이지만 active, 1M. 문서의 `Claude Fable 5` / `claude-fable-5` 표기는 **코드와 일치하므로 유지**하고 수치만 고쳤다.
- 별도 발견(범위 밖, 리드 상신): 별칭이 legacy 모델을 가리킨다. 5.1 로 올릴지는 코드 변경이라 이 카드에서 하지 않았다.

## 기준 규칙과 코드

- 로컬 규칙 `.claude/rules/moai/workflow/context-window-management.md:24-27`: Fable (1M) 50%, Sonnet 5 (1M) 50%,
  Sonnet 4.x / earlier standard (200K) 90%, Haiku (200K) 90%. 문서 표를 이 행 구성에 맞췄다.
- statusline 구현: `internal/statusline/renderer.go:717` `softThresholdPct` — 창 크기 `>= config.HandoffLargeWindowCutoff`(=`500_000`,
  `internal/config/defaults.go:403`) 이면 50%, 아니면 90%. 모델명이 아니라 창 크기로 가르므로, 흐름도는 1M 가지에
  Sonnet 5·Fable 를 넣고 200K 가지를 Haiku·이전 Sonnet 으로 좁혔다.

## 변경 (24 파일, +53/−53)

| 파일 (로케일) | 변경 |
|---|---|
| `multi-llm/model-policy.md` (en/ja/zh/ko) | Fable 5 256K→1M, Sonnet 5 200K→1M |
| `cost-optimization/prompt-caching.md` (en/ja/zh/ko) | 같은 두 행 |
| `advanced/token-budget.md` (en/ja/zh/ko) | 256K 행 → `Fable / Sonnet 5 (1M)` 50%; `Sonnet / Opus 표준 (200K)` → `Sonnet 4.x 이하 표준 (200K)` |
| `advanced/tokenomics-overview.md` (en/ko) | 같은 두 행 (ja/zh 는 이 표가 없음) |
| `advanced/statusline.md` (en/ja/zh/ko) | 흐름도 두 가지 라벨 + 본문 「200K/256K」 2곳 → 「200K」 |
| `advanced/agent-guide.md` (en/ja/zh/ko) | 「200K/256K 계열」 → 「200K 계열」 |
| `multi-llm/_index.md` (ko) | Fable 5 / Sonnet 5 행 (다른 로케일엔 이 표 없음) |
| `claude-code/context-memory/context-window.md` (ko) | Fable 5 1M(「Claude Code에서는 256K 윈도우로 동작」 삭제), Sonnet 5 1M, 본문 「256K·200K 모델」 → 「200K 모델」 |

치환은 적중 횟수를 단언하는 스크립트로 한 번에 적용(49건, 불일치 시 전체 미기록). `256KiB`(moai-design 파일 상한)는 다른 뜻이라 제외.

## 검증

- 잔여: `grep -rnE "256K|256,000" docs-site/content/ | grep -v 256KiB` → 0행. `grep -rnE "(Sonnet 5|Fable)[^\n]{0,40}200K" docs-site/content/` → 0행.
  양성 대조: 같은 계열 패턴을 수정 전 develop 에 `git grep` 으로 걸면 43행.
- 빌드: `hugo --minify -s docs-site -d <스크래치>` → exit 0, `grep -ciE "warn|error"` → 0 (`hugo.log`).
- Mermaid 방향 `flowchart TD` 유지, 본문 이모지 추가 없음, 4 로케일 같은 커밋.

## t1094 겹침 (카드 [HARD])

t1094 (`WT-opus55-docs` @ `a29db12ec`, develop 미착지) 변경 파일과 이 카드 대상 교집합:

- 측정: `git diff --name-only` (이 카드) 와 `git diff --name-only 08113ff0f WT-opus55-docs` 를 정렬 후 `comm`.
  이 카드 24파일 중 **겹침 20**, 비겹침 4 (`advanced/agent-guide.md` ×4 로케일).
- 겹치는 20: statusline ×4, token-budget ×4, prompt-caching ×4, model-policy ×4, tokenomics-overview (en, ko),
  ko context-window, ko multi-llm/_index.
- 같은 줄 충돌 예상: statusline 142행(t1094: Opus 5→5.5), tokenomics 설명 문단(en). 그 외는 인접 행.
- 조치: t1094 가 develop 에 착지하기 전에는 병합 창을 요청하지 않는다. 착지 후 develop 흡수 → 충돌을 양쪽 변경을 합쳐 해소 → 병합 트리에서 hugo 재빌드.

## Gaps

- `Claude Code 는 Claude 슬롯 기준(Opus=1M, Sonnet/Haiku=200K)으로 context_window_size 를 보고` 문장(token-budget, statusline, 로컬 규칙 동일)은
  Claude Code 런타임 동작에 대한 주장이라 공식 API 문서로 판정할 수 없어 손대지 않았다. Sonnet 5 가 1M 인 지금 이 문장이 여전히 맞는지는 미측정.
- ja/zh/en 에 ko 전용 표(context-window 모델 표, multi-llm/_index 표, ja/zh tokenomics 표)가 없는 로케일 간 차이는 기존 상태이며 이 카드 범위 밖.
- 로컬 hugo 서버 렌더 확인(육안)은 하지 않았다 — 빌드 성공과 경고 0만 관측.

## Residual-risk

- t1094 흡수 시 충돌 해소를 잘못하면 Opus 5.5 반영이나 이 카드 수치 중 한쪽이 사라질 수 있다 — 흡수 후 잔여 grep 과 `Opus 5.5` 존재를 함께 재측정해야 한다.
