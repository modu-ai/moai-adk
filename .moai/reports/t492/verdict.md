# t492 판정서 — 창 집행 (lane-6, 2026-09-07)

카드: t492 · 브랜치: `WT-rules-always-loaded-diet` · 창: `moai integration acquire --name lane-6`

## 좌표 — 전부 직접 재판독

리드가 전한 값을 인용하지 않고 이 워크트리에서 다시 읽었다.

| 항목 | 명령 | 관측 |
|---|---|---|
| 흡수 tip | `git rev-parse develop` | `45aaddf1bad62c73459fe5fef2445189136ad9ec` |
| 원격 | `git rev-parse origin/develop` | `615d18c1f990eebd84d96743d81ce1f0ae28bb04` |
| 흡수 전 내 HEAD | `git rev-parse --short HEAD` | `e6b82eb24` |
| **병합 트리** | `git rev-parse --short HEAD` (흡수 후) | **`d64f4d946`** |
| 충돌 | `git rev-parse -q --verify MERGE_HEAD` | 무출력 / rc=1 (충돌 없음, 워킹트리 clean) |
| 미푸시 | `git rev-list --count origin/develop..HEAD` | `83` |

흡수는 `git merge --no-edit develop`. 충돌 0, CHANGELOG 충돌 없음(이 카드는 CHANGELOG 를 건드리지 않는다).

baseline `cc78d1479` 는 흡수 시점에 세 카드 뒤처져 있었다(t488 `dfe25bc09` · t471 `45aaddf1b` · 그리고 t473 `df74b3c9d`). 아래 수치는 전부 **병합 트리 `d64f4d946`** 에서 다시 잰 것이다.

## 병합 트리 재측정

```
go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v

    token_budget_guard_test.go:69: always-loaded surface = 74535 tokens (budget 77600, headroom 3065, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
```

**여유 +3,065 · entries 17 · PASS.**

## 이 카드 몫의 귀속 — 병합 트리에서 직접 측정

산술로 도출하지 않고 두 실측으로 잰다.

```
git show develop:.claude/rules/moai/core/verification-claim-integrity.md | wc -c   →  26629
git show HEAD:.claude/rules/moai/core/verification-claim-integrity.md    | wc -c   →  17668
```

stub **−8,961 B**. 가드는 파일당 `floor(bytes/4)` 를 세므로 `floor(26629/4)=6657` → `floor(17668/4)=4417` = **−2,240 토큰**. 이 값은 흡수 전 내 트리 `e6b82eb24` 에서 잰 것(−2,240)과 **같다** — 흡수가 내 델타를 바꾸지 않았다.

companion 은 `paths: "**/verification-claim-integrity*.md"` 를 갖고 있어 **가드 표면 밖**이다(`head -4 … | grep paths:` 로 확인). 6,860 → 17,278 B 증가는 always-loaded 표면에 기여하지 않는다.

```
git show develop:…-detail.md | wc -c  →   6860
git show HEAD:…-detail.md    | wc -c  →  17278
```

**정합 확인(내 측정이 아님, 대조용)**: 리드가 전한 develop 단독 실측은 `76775 / headroom 825` 였다. `76775 − 2240 = 74535` 로 내 병합 트리 실측과 맞아떨어지므로 잔여 없는 귀속이나, **76,775 는 리드의 값이고 내가 재지 않았다.** 내 주장은 `74,535`(직접) 와 `−2,240`(직접) 둘뿐이다.

## 승인 범위 준수

승인은 **C2 보수층**이었다. R1–R4 구제책 표와 그 비용표는 stub 에 그대로 있으며 손대지 않았다. 공격층 미실행.

## 잔류 의무 — 실질 유실 없음

이관한 9,555 B 안의 의무 문장 개수:

```
grep -c '\[HARD\]' moved-block-A-four-tests.md moved-block-B-instances-limits-divergence.md  →  0, 0
grep -c 'MUST'     moved-block-A-four-tests.md moved-block-B-instances-limits-divergence.md  →  0, 0
```

stub 의 `[HARD]` 총수 **7 → 8**(빠진 것 0, 신설 포인터 1). 잔류 4건의 문면 인용과 행번호는 `analysis.md` §8.4 에 있다.

## 미러 정합

`diff live template | grep -c '^[<>]'` — stub 21 / companion 10, 합 31. 편집 전 29 대비 `+2` 는 의도적으로 중립화한 companion 헤더 1쌍이다. 나머지는 전부 기존 중립화이거나 stub→companion 으로 함께 이동한 중립화다. **모든 diff 라인이 귀속된다.**

## SPEC 미작성 — 리드 판정

배차는 `cmd: /moai plan …` 을 지시했으나 산출물은 분석서였고, **운영자 승인이 그 분석서 위에서 났으며 C2 실행도 그 승인 위에 있다.** 이미 실행된 편집을 사후에 SPEC 으로 감싸는 것은 gate-ordering 위반의 모양이므로 SPEC 을 만들지 않았다. **이는 리드 판정이며 이 카드의 임의 판단이 아니다.** 후속 카드(C1 · C3 · 가드 확장)는 정상 SPEC 경로로 간다.

## 검증 범위와 미검증

**검증한 것** — 이 카드가 바꾼 것이 닿는 패키지:

```
go test ./internal/template/... ./internal/config/...
```

**검증하지 않은 것**:

- **전 패키지 스위트를 로컬에서 돌리지 않았다**(`CLAUDE.local.md` §4.1 / §6). 전수 판정은 CI 몫이다.
- 흡수가 데려온 `internal/cli` · `internal/kanban` 변경(t488 · t471 · t473)은 **각 레인이 자기 창에서 검증했고 CI 가 판정한다.** 이 카드가 재검증하지 않았다.
- 리드가 전한 develop 단독 값 `76,775 / 825` 는 **내 측정이 아니다.**

## 잔여 위험

- always-loaded 표면은 계속 자란다(하루 +493 전례). 여유 +3,065 는 영구적이지 않으며, 다음 상환 수단은 C1(`@AGENTS.md` import 제거, 실측 −3,693)로 남아 있다.
- 이관된 절차를 읽으려면 companion 을 열어야 한다. 그 companion 은 `paths:` 로 `verification-claim-integrity*.md` 편집 시에만 자동 로드되므로, 다른 파일에서 술어를 적용하는 저자는 **stub 의 [HARD] 포인터를 읽고 직접 열어야 한다.** 포인터가 그 지시를 담고 있으나, 지시를 따르지 않는 경우까지 막지는 못한다.
