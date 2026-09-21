# t473 — verdict (run phase, B안)

카드: t473 · 레인: lane-11 · 브랜치: `WT-token-budget-overflow` · base `25a3212a9`
운영자 판정: **B안** (A안 이관 + `moai.md` §8 다이어트)

## Claim

always-loaded 예산 초과(123토큰)를 해소했고, 여유를 168 → 1,044 토큰으로 회복했다. 옮긴 내용은 하나도 삭제되지 않았다.

## Evidence

편집 전:
```
always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)
--- FAIL: TestAlwaysLoadedTokenBudget
```

편집 후:
```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v -count=1
always-loaded surface = 76556 tokens (budget 77600, headroom 1044, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.02s)
ok  github.com/modu-ai/moai-adk/internal/config
```

패키지 스위트:
```
$ go test ./internal/config/... -count=1
ok  github.com/modu-ai/moai-adk/internal/config          5.213s
ok  github.com/modu-ai/moai-adk/internal/config/atomicfile 0.807s
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy 1.108s

$ go test ./internal/template/... -count=1
ok  github.com/modu-ai/moai-adk/internal/template          28.609s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit 0.838s

$ make build   → rc=0
```

## 변경 내역 (감축 −1,167 tok)

| 감축 | 파일 | 처분 |
| ---: | --- | --- |
| −216 | `main-checkout-branch-guard.md` (7,800 → 6,939 B) | t467 의 `git branch` 플래그 분류를 동반자로 이관 + Rules 표 행 축약. stub 엔 성질 요약 + 포인터 |
| −951 | `.claude/output-styles/moai/moai.md` (66,255 → 62,451 B) | §8 Localization Contract 의 ko 전용 예시 카탈로그 2본(2,521 B + 2,912 B)을 신규 동반자로 이관. 대표 4행 + 포인터 유지 |

수용처 (둘 다 `paths:` 제한 → 측정 표면 밖, 17 entries 불변):

- `.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md` (8,471 → 9,553 B) — 기존 동반자, description 이 이미 `pattern set` 을 자기 소관으로 선언하고 있었다
- `.claude/rules/moai/core/output-style-localization-catalogue.md` (신규, 6,643 B) — `paths: "**/output-styles/moai/*.md,**/output-style-localization-catalogue.md"` 로 출력 스타일을 편집할 때 로드된다

템플릿 미러 3본 동기화 + 신규 1본 생성. 미러는 `cp` 가 아니라 같은 편집을 각각 적용했고, 기존 중립성 처리(SPEC-ID / REQ 토큰 제거)를 보존했다. 신규 동반자 미러 중립성 스캔: 금지 패턴 0.

## Baseline-attribution

편집 전 77,723 / 편집 후 76,556 은 모두 이 워크트리(`WT-token-budget-overflow`, base `25a3212a9`)에서 같은 명령으로 잰 값이다. 감축 폭 1,167 은 두 파일 실측 감소(216 + 951)와 일치한다.

## Gaps

- **렌더 품질 영향은 측정하지 않았다.** 카탈로그가 always-loaded 에서 빠진 뒤 라벨 번역이 흔들리는지는 이번 트리에서 재지 않았다. 완화책은 (a) [HARD] 번역 의무 자체는 stub 에 남겼고 (b) 대표 4행을 인라인 유지했으며 (c) 전체 카탈로그는 출력 스타일 편집 시 로드된다는 것이다. 관측 창이 필요하다.
- `internal/cli` 등 다른 패키지는 돌리지 않았다(이 변경이 닿지 않는다). 전 패키지 판정은 CI 몫이다.
- `moai.md` §8 Session Handoff(2,391 tok)는 손대지 않았다. `session-handoff.md` 의 드리프트 감시 조항이 "moai.md §8이 4로케일 표를 갖는다"고 못박고 있어, 확인 결과 그 표(line 710)는 그대로 남아 있다.

## Residual-risk

- 여유 1,044 는 t453 이 잰 단일 소조항 단위(~87 tok)의 약 12배다. 상향 사슬을 끊을 만한 폭이지만 영구적이지 않다 — 다음 대형 조항 유입 시 다시 좁아진다.
- `estimateTokens` 는 `len/4` 근사(±15%)다. 상대 증분 감시용이며 절대 회계로 읽으면 안 된다.
- 신규 동반자의 `paths:` 는 출력 스타일 파일 편집 시에만 매칭된다. 렌더 시점에는 로드되지 않는 것이 설계 의도지만, 그 의도가 틀렸다면 위 Gaps 첫 항목이 그 증거로 나타날 것이다.

## 상수는 건드리지 않았다

`AlwaysLoadedTokenBudget = 77600` 을 그대로 두었다. t453 이 *"다음 트립에 자동 정당성은 없다"* 고 적었고, 이 카드는 그 문장이 예고한 다이어트 카드다. 상향 사슬에 한 칸을 더 얹는 대신 사슬을 되돌리는 방향으로 처리했다.
