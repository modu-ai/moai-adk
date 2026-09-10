# t473 — always-loaded 예산 초과 귀속 실측 (plan phase)

측정 트리: `25a3212a9` (= `origin/develop`), 워크트리 `.claude/worktrees/t473`
베이스라인 트리: `9636d143d` (t453 상향이 인용한 트리 SHA)
단위: 가드와 동일 — 파일별 `len(bytes)/4`, 대상 = `paths:` 없는 `.claude/rules/moai/**/*.md` + 고정 슬롯 3본

## Claim

always-loaded 표면이 예산을 123토큰 초과했고, 증분 +291토큰은 3개 파일 · 3개 커밋으로 전수 귀속된다.

## Evidence

로컬 재현 (CI 실측과 일치, flake 아님):

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v -count=1
token_budget_guard_test.go:69: always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)
token_budget_guard_test.go:73: ... exceeds budget 77600 (overflow 123); surface has 17 entries
FAIL
```

파일별 분포 (17 entries, 합계 77723 = 가드 총합과 바이트 단위 일치):

| tok | 비중 | 파일 |
| ---: | ---: | --- |
| 16563 | 21.3% | `.claude/output-styles/moai/moai.md` |
| 8567 | 11.0% | `.claude/rules/moai/workflow/kanban-dispatch.md` |
| 6707 | 8.6% | `.claude/rules/moai/core/agent-common-protocol.md` |
| 6657 | 8.6% | `.claude/rules/moai/core/verification-claim-integrity.md` |
| 5455 | 7.0% | `.claude/rules/moai/core/askuser-protocol.md` |
| 5299 | 6.8% | `.claude/rules/moai/workflow/session-handoff.md` |
| 4941 | 6.4% | `CLAUDE.md` |
| 4126 | 5.3% | `.claude/rules/moai/core/moai-constitution.md` |
| 4083 | 5.3% | `.claude/rules/moai/workflow/cross-session-messaging.md` |
| 3693 | 4.8% | `AGENTS.md` |
| 2220 | 2.9% | `.claude/rules/moai/workflow/context-window-management.md` |
| 1950 | 2.5% | `.claude/rules/moai/workflow/main-checkout-branch-guard.md` |
| 1745 | 2.2% | `.claude/rules/moai/workflow/cache-aware-execution.md` |
| 1731 | 2.2% | `.claude/rules/moai/workflow/goal-directive.md` |
| 1600 | 2.1% | `.claude/rules/moai/core/moai-mcp-tools.md` |
| 1398 | 1.8% | `.claude/rules/moai/workflow/skill-routing.md` |
| 988 | 1.3% | `.claude/rules/moai/core/native-idiom-and-register.md` |

증분 귀속 — 베이스라인 77432 → 현재 77723, delta **+291**, 나머지 14파일 무변화:

| 증분 | 바이트 | 파일 | 커밋 | 카드 |
| ---: | ---: | --- | --- | --- |
| +187 | +745 | `main-checkout-branch-guard.md` | `b8acd04cc` | t467 (BRANCH-GUARD-FLAGCLASS M3, doctrine pair v1.3.3) |
| +81 | +325 | `goal-directive.md` | `522116b1d` | t436 (goal prose arm) |
| +23 | +90 | `agent-common-protocol.md` | `8c6eab926` | t216 (HOOK-WIRING-DRIFT M3) |

커밋별 파일 크기 실측 (`git cat-file --batch-check`):

```
7800 b8acd04cc:...main-checkout-branch-guard.md      7055 b8acd04cc^:...  (= 베이스라인)
6925 522116b1d:...goal-directive.md                  6600 522116b1d^:...  (= 베이스라인)
26829 8c6eab926:...agent-common-protocol.md          26739 8c6eab926^2:.. (= 베이스라인, develop 쪽 부모)
```

산술 검산: 187 + 81 + 23 = 291. 77432 + 291 = 77723. 예산 77600 = 77432 + 168(t453 이 남긴 여유). 초과 = 291 − 168 = **123**.

## Baseline-attribution

베이스라인 77432 는 t453 주석에서 옮겨 적은 값이 아니라 이 트리에서 `9636d143d` 를 직접 읽어 재계산한 값이며, 주석이 적어 둔 숫자와 일치한다. 복제 측정기는 현재 트리에서 가드와 동일한 `17 entries / 77723 / overflow 123` 을 재현하므로 공허하지 않다.

## Gaps

- 세 커밋이 각 파일에 더한 **내용의 필요성** 자체는 판정하지 않았다(측정만 했다).
- 예산을 얼마나 회복해야 적정한지의 기준값은 측정 대상이 아니다 — 판정 사항이다.
- `moai.md` §8 축소 시 렌더 품질 영향은 재지 않았다.

## Residual-risk

`estimateTokens` 는 `len/4` 근사(±15%)다. 상대 증분 감시에는 충분하나 절대 토큰 회계로 읽으면 안 된다 — 가드 자신의 주석이 같은 단서를 단다.

## 리드 추정 대조

- **t467 → 확인.** 단일 최대 기여(+187, 증분의 64%).
- **t463 `CLAUDE.local.md` §4.1 → 기각.** `CLAUDE.local.md` 는 측정 표면 17항목에 없다(설계상 부외 — t453 주석이 "열거 밖 주입물…범위 설계상 부외"로 명시). 이 가드에 대한 t463 기여는 **0**이다.
- 리드가 지목하지 않은 기여 2건: t436(+81 — t453 상향이 이미 예산에 반영해 둔 몫), t216(+23).

## 3안 판정 근거

**(b) 예산 상향은 이 카드에서 기각을 권고한다.** 상수 주석의 상향 사슬은 `75,000 → 76,000 → 76,210 → 76,400 → 77,200 → 77,600` 이고, 마지막 상향(t453, 2026-09-03)이 스스로 적었다: *"오늘 하루 성장 +493 전체를 덮지 않는 것은 의도다 — 그 이상 성장은 가드를 다시 트립해 다이어트 카드의 착수 근거로 간다. 다음 트립에 자동 정당성은 없다."* 지금 트립이 그 문장이 예고한 트립이며, 이 카드가 그 다이어트 카드다. 더해 always-loaded 는 전 사용자 매 세션 비용이므로 상향은 영구 비용이다.

**(a) 축소 권고.** 인과적으로 정렬된 최소 처방이 이미 존재한다:

- t467 이 stub 에 넣은 `+745 B` 는 `git branch` 플래그 분류 카탈로그다. 같은 독트린 쌍의 지연 로드 동반자 `main-checkout-branch-guard-detail.md` (8,471 B, `paths:` 제한 있음 → 측정 표면 밖)의 description 이 **이미 `pattern set` 을 자기 소관으로 선언**하고 있다. 즉 이 블록은 처음부터 동반자 몫이었다.
- 이관 시 예상: 77723 − 187 = **77536** (예산 77600 이하, 여유 64).

**다만 여유 64 는 t453 이 측정한 "단일 소조항" 단위(~87 tok)보다 작다.** 이 최소 처방만으로는 다음 독트린 조항 하나에 즉시 재트립한다. 실질 여유를 만들려면 대형 파일 축소가 필요하며, 측정된 최대 후보는 다음과 같다:

`moai.md` **§8 Response Templates = 51,130 B / 12,782 tok — 전체 표면의 16.4%**, 어떤 규칙 파일보다 크다. 이 절은 자기 본문에 *"render-only, not canonical — canonical lives in `session-handoff.md` (SSOT)"* 라고 적고 있어, 정본이 아닌 렌더 표면이 always-loaded 최대 단일 집중을 차지하고 있다.

축소 폭 결정은 운영자 게이트 사항으로 남긴다.
