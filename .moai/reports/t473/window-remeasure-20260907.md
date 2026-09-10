# t473 — 통합 창 재측정 (lane-3, 2026-09-07)

리드가 lane-3 에게 t473 창을 지명했다. 카드는 이미 닫혀 있고(판정서 `verdict.md`,
재검증 `reverify-20260907.md`), 남은 것은 창 집행뿐이다. 이 파일은 **병합할 tip 에서**
다시 잰 기록이다 — 병합 전 검증을 병합 후 근거로 재사용하지 않기 위해서다.

## 흡수

| 항목 | 값 |
| --- | --- |
| 흡수 대상 (로컬 develop) | `098f612fa` — t490 착지본 |
| 흡수 전 카드 tip | `30e6a60f4` (ahead 3) |
| 흡수 병합 커밋 | `0d3c0e92b` |
| 충돌 | 0 |

`origin/develop` 은 `615d18c1f` 로 로컬보다 뒤에 있다. 흡수 대상은 로컬 develop 이다.

## Claim

t473 이 회복한 always-loaded 예산 여유가, t490 을 흡수한 병합 트리에서도 그대로
성립한다. 항목 수(17)와 여유(1,044 토큰) 모두 카드 판정 당시와 같다.

## Evidence — 라운드 1 (`0d3c0e92b`)

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v -count=1
    token_budget_guard_test.go:69: always-loaded surface = 76556 tokens (budget 77600, headroom 1044, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
ok  github.com/modu-ai/moai-adk/internal/config	0.396s

$ go test ./internal/config/... ./internal/template/... -count=1
ok  github.com/modu-ai/moai-adk/internal/config	1.788s
ok  github.com/modu-ai/moai-adk/internal/config/atomicfile	0.555s
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.872s
ok  github.com/modu-ai/moai-adk/internal/template	25.780s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit	1.681s

$ make build   → rc=0   (직접 캡처; zsh 는 PIPESTATUS 가 아니라 pipestatus 다)
$ git status --short   → 빈 출력 (catalog.yaml 재생성 후에도 내용 불변)
```

## Baseline-attribution

세 측정 모두 이 워크트리(`WT-token-budget-overflow`)의 흡수 병합 커밋 `0d3c0e92b`
에서, 위에 적힌 명령 그대로 이번 창에서 실행해 관측한 출력이다. 카드 판정 당시
값(76,556 / 1,044 / 17)과 같지만, **그 값을 옮겨 적은 것이 아니라 다시 잰 것이다.**

라운드 2 는 이 파일을 커밋한 tip 에서 같은 명령을 다시 돌려 값이 불변임을 확인한
뒤 develop 에 병합한다 — 재측정 트리와 병합 트리를 같게 두기 위해서다. 라운드 2 의
관측 결과는 창 완료 보고에 싣는다(자기 자신을 담을 수 없는 파일이라서다).

## 재측정 범위

건드린 표면은 `.claude/rules/**`, `.claude/output-styles/**`, 그 템플릿 미러다.
예산 가드(`internal/config`)와 템플릿 임베드(`internal/template`)가 그 표면을 읽는
두 패키지이므로 범위는 이 둘이다. 전 패키지 판정은 develop push 가 일으키는 CI 몫이다.

## Gaps

- `internal/cli` 등 나머지 패키지는 돌리지 않았다 — 이 변경이 닿지 않는다.
- 렌더 품질 영향(카탈로그가 always-loaded 에서 빠진 뒤 라벨 번역이 흔들리는지)은
  이번에도 재지 않았다. 카드 판정서 Gaps 첫 항목이 그대로 유효하다.
- CI 는 이 창에서 돌리지 않았다. 레인은 CI 를 직접 요청하지 않으며, 판정은 리드가
  일괄 push 뒤 읽는다.

## Residual-risk

- 여유 1,044 는 영구적이지 않다. 다음 대형 조항이 always-loaded 로 들어오면 다시 좁아진다.
- `estimateTokens` 는 `len/4` 근사(±15%)다. 상대 증분 감시용이며 절대 회계가 아니다.
