---
id: SPEC-EVIDENCE-PATH-EXCEPTION-001
title: "증거 인용 경로 좁히기 — 판정서에 한정한 gitignore 예외와 죽은 negation 정리"
version: "0.4.0"
status: in-progress
created: 2026-09-20
updated: 2026-09-21
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".gitignore, .claude/rules/moai/core, .claude/agents/moai, .claude/output-styles/moai, internal/template/templates, docs-site/content"
lifecycle: spec-anchored
tags: "evidence, citation, gitignore, verdict, verification-claim-integrity, doctrine, docs-site"
tier: L
related_specs: [SPEC-EVIDENCE-CITATION-CANON-001]
---

# SPEC-EVIDENCE-PATH-EXCEPTION-001 — 증거 인용 경로 좁히기

## HISTORY

### 2026-09-20 — 최초 작성 (카드 t1039)

카드는 「독트린이 `.moai/reports/<card-id>/`를 **추적되는** 인용 대상이라고 단언하는데 그 경로는
`.gitignore:227`의 블랭킷으로 무시된다」는 모순으로 열렸다. 즉 규칙이 거짓 전제 위에 서 있고,
그 결과 인용된 증거가 저작 트리 밖에서 해소되지 않는다.

운영자 판정은 **블랭킷 철회가 아니다**. 2026-09-14 블랭킷은 08-31 독트린을 알고 내린 의도적
지시이며 2026-09-20에 재확인됐다. 이 SPEC은 그 지시를 뒤집지 않고 **문장을 좁히고** 좁혀진
문장이 참이 되도록 **좁은 예외**를 뚫는다.

셋째 항목이 함께 들어왔다. `.gitignore:109`의 `!.moai/reports/**/*.log`(2026-09-01 보강)는
`:227`이 부모 디렉터리를 제외하므로 **원리상 발화할 수 없다** — 그런데 그 위의 주석은 여전히
무언가를 고친 것처럼 읽힌다. 죽은 규칙보다 **살아 있다고 읽히는 죽은 주석**이 더 비싸다.

### 2026-09-20 — 레인 측정 §2 정정 (문서 표면 개수)

레인의 리터럴 grep은 `citation target is a tracked path` 3건을 보고했다. `reference.md` 두 사본은
같은 문장을 `a **tracked** path`로 굵게 쓰기 때문에 리터럴에 걸리지 않는다. 볼드 허용 정규식으로
재측정하면 **5건**이다(§1.3).

### 2026-09-20 — iteration 2 (plan-audit FAIL 0.725 대응)

plan-audit iter1이 FAIL(0.725 / Tier M 임계 0.80)을 냈고 blocking 결함 6건을 지목했다. 이
개정이 여섯 가지를 닫는다. 무엇이 어떻게 바뀌었는지는 `progress.md` § iteration 기록에 결함별로
적혀 있다. 이 개정의 두 가지 구조적 변화만 여기 적는다.

1. **docs-site가 범위 안으로 들어왔다**(D1). 같은 거짓 주장이 4개 로케일 × 3개 페이지 계열에
   퍼져 있다는 사실이 측정됐고(§1.3), 공개 문서를 거짓인 채로 두면 이 카드가 고치는 결함이
   가장 많이 읽히는 표면에서 그대로 살아남는다.
2. **Tier M → Tier L**(D3). 편입 후 수정 표면이 Tier M 밴드(5-15)를 넘고, 요구사항·인수조건
   수도 Tier M 예산(16/16)을 넘는다. 판별식은 §0에 적는다. Tier L 산출물 집합에 따라
   `design.md`·`research.md`가 추가됐다.
   *(iteration 2는 이 파일 수를 **16**으로 적었다. 실측은 **20**이며 iteration 3이 정정했다 —
   아래 iteration 3 항목. Tier 결론은 바뀌지 않는다.)*

### 2026-09-20 — iteration 3 (plan-audit PASS-WITH-DEBT 0.85, blocking 3건 대응)

iter2 판정은 PASS-WITH-DEBT 0.85(Tier L 임계 0.85, 여유 0)였고 blocking 결함 3건을 남겼다.
이 개정이 셋을 닫는다. 결함별 변경은 `progress.md` § iteration 기록에 있다. 구조적으로 남길
것은 셋이다.

1. **`AC-EPE-022`의 기계 하위 점검이 공허했고, 그 RED-now 수치가 거짓이었다**(D10). 기록된
   정규식은 이 인벤토리에서 **0행**을 내며(착수 전 초록 = 공허), RED-now로 적혀 있던 **18은
   어떤 명령의 출력도 아니었다** — `20 − 2`로 계산한 값이 fenced 명령+출력 블록 안에
   관측값으로 놓여 있었다. 하위 점검을 실제로 발화하는 두 갈래로 다시 세우고, 인용하는 모든
   수치를 이 트리에서 다시 돌려 출력을 그대로 옮겼다.
2. **수정 파일 수 16이 자기 구성 행과 모순했다**(D11). 층별 행의 합은 **20**이다. 네 자리를
   정정했다. Tier 결론은 그대로이며(20 > 15) 오히려 더 확실해졌다.
3. **`REQ-EPE-002`의 본문 범위와 좌표가 어긋났다**(D12). 본문은 비-판정서 **파일** 3종으로
   범위를 그었는데 좌표 목록에는 디렉터리를 주석하는 `manager-lead.md:63`이 들어 있었고 어떤
   AC도 그 좌표에 닿지 않았다. 같은 형태를 docs-site 쪽 `AC-EPE-022`는 위반으로 세고 있었으므로,
   비대칭을 **본문을 넓히는 쪽**으로 해소했다 — 두 표면이 같은 두 형태를 잡는다.

### 2026-09-21 — 폭 경계 확정 (독법 B) 반영, run 단계 blocker 해소

운영자가 2026-09-21에 **독법 (B)** — 되살리는 것은 `verdict.md` **한 파일**, 깊이는 카드
디렉터리 **한 단계** — 를 확정했다. 출처는 run 단계 배차문이며 `progress.md` §E.2.0이 그
기록이다. 런타임은 이미 (B)로 착지해 있었고(run 단계 M1~M6), 이 개정은 **산문이 코드가 하는
일을 서술하도록** 맞추는 것이다. 다섯 자리를 고쳤다.

1. **`REQ-EPE-006`의 `[UNRESOLVED]` 표식을 해소**했다(§2.2). §3.0의 두 독법 귀결표와 §3.3의
   권고 기록은 **그대로 둔다** — 권고가 (A)였고 운영자가 (B)를 골랐다는 사실 자체가 이 SPEC이
   남겨야 할 기록이며, 치워야 할 발판이 아니다.
2. **(A) 조건부 절들을 「(B) 아래에서는 해당 없음」으로 표시**했다(§3.2,
   `acceptance.md` `AC-EPE-002`의 [HARD] 블록, `design.md` §1.3). (B)에서는 계열 글롭이
   존재하지 않으므로 그 구현 지시들이 발화하지 않는다. **삭제하지 않는다** — 대안이
   조건부로 실려 있었다는 사실이 읽혀야 한다.
3. **`acceptance.md` §B.9의 「표는 command를 싣지 않는다」 [HARD]의 범위를 명시**했다(D20).
   그 문장은 §B.9의 판정식 원장 규약이며 §B.5의 P-프로브 셀에 소급되지 않는다.
4. **`acceptance.md` §B.10 regression-guard 행에 빠져 있던 절을 채웠다**(D21). `L-R1`·`L-R2`는
   release-blocking인 `AC-EPE-022`의 PASS 조건이므로 단독으로 릴리스를 막을 수 있다.
5. **`REQ-EPE-005`를 `AC-EPE-001`·`AC-EPE-003`이 이름 붙이도록 했다**(D22). 커버리지는 이미
   실질적이었고 교차참조만 없었다.

---

## 0. Tier 판정 — 적용한 판별식

> **판별식: Tier는 재서 나온 규모가 정하며, 맞추고 싶은 예산이 정하지 않는다.**

`spec-workflow.md` § SPEC Complexity Tier는 예산 초과를 「예산을 완화할 신호가 아니라 tier up
하거나 분할할 신호」로 명시한다. 이 SPEC은 **독립된 두 측정**이 Tier M을 넘는다.

| 축 | 측정 | Tier M 밴드 | 판정 |
|---|---|---|---|
| 수정 파일 수 | 20 (독트린 루트 2 + 도달 표면 루트 2 + 템플릿 미러 4 + docs-site 12, §1.3) | 5-15 | 초과 |
| 요구사항 수 | 18 | ≤16 | 초과 |
| 인수조건 수 | 23 | ≤16 | 초과 |

**요구사항을 지워서 수를 맞추지 않았다.** 접은 것은 한 건뿐이며 그것은 규모 축소가 아니라
**우발적 분리의 복원**이다: iteration 1이 접미 하위 ID(`…-008` 뒤에 소문자 `a`)로 달아 두었던
「픽스처는 baseline이 아니다」는 `REQ-EPE-008`(제자리 재측정)과 같은 하나의 의무 — 「제자리에서
재라, 픽스처는 그 측정이 아니다」 — 이므로 `REQ-EPE-008`의 `shall not` 절로 접었다. 접으면서
금지의 3중 명문화(요구사항 + `acceptance.md` §A + `AC-EPE-006`)는 그대로 유지된다.
이 접기는 iteration 1이 minor로 지목한 접미 하위 ID 문제도 함께 닫으며, 접미 ID가 요구사항
층에서 사라졌으므로 **`REQ-EPE-` 토큰 집계가 `001`~`018` 연속 18건과 정확히 일치한다**.

Tier L의 귀결: 산출물 5종(`spec.md`·`plan.md`·`acceptance.md`·`design.md`·`research.md`),
plan-auditor PASS 임계 **0.85**, REQ/AC 상한 25/25.

---

## 1. 배경과 측정

### 1.0 이 결함의 가장 가까운 실사례 — 이 SPEC을 심사한 감사가 자기 판정서를 반출하지 못했다

> **이 항목이 동기 절에 있는 이유.** 역사적 실사례를 다섯 건 나열하는 것보다, **결함이 이
> 카드 자신의 절차 위에서 재현된 것**이 더 강한 논거다. 감사자가 자기 산출물에 대해 스스로
> 증명한 형태이므로, 이 카드가 고치려는 문장이 거짓이라는 주장에 해석의 여지가 없다.

`plan-auditor.md:601`은 [HARD]로 **"an audit is complete only when its verdict is exported"**
라고 규정하고 반출 목적지를 `.moai/reports/<card-id>/plan-audit.md`로 못 박는다. iter1 감사는
그 자리에 판정서를 썼다. 그런데 그 경로는 `:227` 블랭킷으로 무시되므로 **판정서는 반출되지
않았고, 그 [HARD] 문장은 자신을 판정한 감사에 대해 거짓이다.**

측정(이 트리, HEAD `116820f40`, 이번 실행):

```
$ git status --porcelain --ignored -- .moai/reports/t1039
!! .moai/reports/t1039/                       ← 디렉터리 통째로 ignored

$ git status --porcelain -- .moai/reports/t1039
                                              ← 0행: --ignored 없이는 보이지도 않는다

$ ls -1 .moai/reports/t1039/
lane-measurements.md
plan-audit.md                                 ← 두 파일 모두 디스크에는 있다

$ git check-ignore -v --no-index .moai/reports/t1039/plan-audit.md
.gitignore:227:.moai/reports/*	.moai/reports/t1039/plan-audit.md
```

양성 대조: 같은 `--ignored` 프레임을 `.moai/reports/t338`에 돌리면 negation이 살린 파일이
ignored 목록에 나오지 않는다(무리 C가 발화 — §1.5). 두 방향이 모두 나오므로 계측기는 실패할
수 있다.

읽는 법 두 가지를 분리해 둔다. **두 번째 명령의 0행을 「문제 없음」으로 읽으면 안 된다** —
`--ignored` 없는 `git status`는 무시된 경로에 대해 구조상 침묵하므로, 그 침묵은 부재가 아니라
미측정이다. 그리고 감사자가 잘못한 것이 **아니다**: 규정된 목적지에 규정대로 썼고, 거짓인
것은 목적지가 추적된다는 **전제**다.

이것이 이 카드가 세는 **6번째 실사례**이고, 유일하게 이 카드 자신의 절차에서 나온 것이다.
잔여 위험(§8)에도 같은 사실이 적혀 있는데, 중복이 아니라 역할이 다르다 — 여기서는 **결함이
실재한다는 증거**이고, 거기서는 **워크트리를 폐기하면 이 두 파일의 유일본이 사라진다는 경고**다.

### 1.1 모순의 두 끝

> **[HARD] 아래 `.gitignore` 줄번호는 이 워크트리(`WT-evidence-path`, HEAD `116820f40`, 417줄)
> 기준이다.** `main`에 체크아웃된 트리의 같은 파일은 319줄이며 줄번호가 일치하지 않는다.
> 트리를 명명하지 않은 줄번호 인용은 다른 트리에서 조용히 빗나간다.

| 끝 | 좌표 | 문장 |
|---|---|---|
| 독트린 | `.claude/rules/moai/core/agent-common-protocol.md:274` | "The citation target is a tracked path — in this repository `.moai/reports/<card-id>/`." |
| 독트린 | `.claude/rules/moai/core/agent-common-protocol-reference.md:62` | 같은 주장 + § Export width |
| ignore | `.gitignore:227` | `.moai/reports/*` (운영자 지시 2026-09-14, 주석 `:224-226`) |
| 공개 문서 | `docs-site/content/{en,ko,ja,zh}/advanced/**` | 같은 주장의 4-로케일 번역본 20행 (§1.3) |
| 선행 SPEC | `SPEC-EVIDENCE-CITATION-CANON-001/spec.md:164` REQ-ECC-001 | `.moai/reports/<card-id>/`를 **추적** 정본 위치로 고정 |
| 선행 SPEC | 같은 파일 `:166` REQ-ECC-003 | **임의 산출물**에 대해 인용 전 반출을 구속 |

### 1.2 [HARD] `git check-ignore`의 두 함정 — 이 SPEC의 모든 판별식이 여기 의존한다

레인 측정 §0에서 확립됐고, 이 SPEC의 인수조건은 **예외 없이** 아래 둘 중 하나만 판별식으로 쓴다.

1. `-v`(verbose)는 **negation 적중에도 exit 0**을 낸다. `-v` 아래의 `rc=0`은 「무시됨」이 아니라
   「어떤 규칙이든 적중함」이다.
2. `--no-index` 없이는 **이미 추적 중인 파일을 건너뛴다**. 그 결과 `rc=1`이 나오는데, 이는
   「negation이 살렸다」와 겉보기로 구별되지 않는다.

**유효 판별식은 둘뿐이다.**

- `git check-ignore --no-index -q <path>` — `rc=0` = 무시됨, `rc=1` = 무시되지 않음
- `git status --porcelain --untracked-files=all <path>` — 「git add 대상이 되는가」의 지상 진실

`-v`는 **판정이 아니라 「어느 규칙이 결정했는가」를 읽는 용도**로만 쓴다(§1.5가 그 용법이다).

### 1.3 수정 표면 (이 SPEC이 실제로 만지는 개수)

#### (a) 저장소 산문 — 리터럴 주장

측정(이 트리, HEAD `116820f40`):

```
$ grep -rlnE 'citation target is a (\*\*)?tracked(\*\*)? path' --include='*.md' --include='*.toml' .
.moai/specs/SPEC-HIERARCHICAL-TEAM-001/progress.md
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
.claude/rules/moai/core/agent-common-protocol-reference.md
.claude/rules/moai/core/agent-common-protocol.md
# (같은 명령을 지금 다시 돌리면 6행이다 — 6번째가 이 SPEC 자신의 spec.md,
#  바로 이 블록이 그 문자열을 담기 때문이다. 자기-적중은 수정 표면이 아니다.)
```

양성 대조(같은 정규식, 존재하지 않는 토큰): `0`. 따라서 이 5건은 패턴이 만들어낸 수가 아니다.

`progress.md` 1건은 **완료된 SPEC의 진행 기록**이므로 수정 대상이 아니다(기록은 나중 사실에
맞춰 고치지 않는다). 리터럴 주장 기준 수정 표면은 **4건** = 저장소 루트 2 + 템플릿 미러 2.

#### (b) 도달 표면 — 비-판정서 경로를 추적으로 주석한 곳

리터럴 문장은 없지만 같은 거짓 전제를 실어 나르는 곳이 둘 더 있다(§2.1 REQ-EPE-002).

```
$ grep -n 'tracked' .claude/agents/moai/manager-lead.md      → :63, :152, :154
$ grep -n 'tracked' .claude/output-styles/moai/moai.md       → :354, :561 (+ :583, :620 무관)
```

여기에 템플릿 미러 2건이 더 붙는다. 그래서 저장소 측 수정 표면은 **8건**이다
(루트 4 + 템플릿 미러 4).

#### (c) docs-site — 같은 주장의 4-로케일 번역본

[HARD] **인벤토리 grep은 로케일 내성이어야 한다.** `ko`/`ja`/`zh` 페이지는 「tracked」를 번역하고
(`추적 경로` / `追跡対象のパス` / `受版本跟踪的路径`), `ko`는 플레이스홀더 자체도 번역한다
(`.moai/reports/<카드-id>/`). 영어 토큰만 겨눈 grep은 **3파일 5행**만 보고하며, 이것이 iter1이
docs-site를 놓친 경로다.

```
$ grep -rnE '(tracked|추적|追跡|跟踪|追踪)' --include='*.md' docs-site/content/ | grep -E '\.moai/reports' | wc -l
20
$ (같은 파이프, 파일별 집계)
en/advanced/{agent-guide.md 1, manager-lead.md 2, token-budget.md 2}
ko/advanced/{agent-guide.md 1, manager-lead.md 2, token-budget.md 2}
ja/advanced/{agent-guide.md 1, manager-lead.md 2, token-budget.md 2}
zh/advanced/{agent-guide.md 1, manager-lead.md 2, token-budget.md 2}
```

양성 대조: 같은 프레임에 존재하지 않는 토큰 → `0`. 계측기 도달 확인:
`grep -rl 'moai/reports' docs-site/ | wc -l` → `41`(0보다 크다).

**12파일 / 20행 / 4로케일 × 3페이지 계열.** (iter1 감사는 같은 축을 19행으로 보고했다. 차이는
정규식 폭이며 — 이 측정이 `ko/advanced/agent-guide.md`의 「추적 경로」 1행을 더 잡는다 —
양쪽 다 이 트리에서 잰 값이다. 이 SPEC은 이 측정을 쓰고, 어떤 인수조건도 **정확 수치 일치를
단언하지 않는다**.)

#### (d) 합계

| 층 | 파일 | 실제 경로 |
|---|---|---|
| 독트린 루트 | 2 | `.claude/rules/moai/core/agent-common-protocol.md`, 같은 이름 `-reference.md` |
| 도달 표면 루트 | 2 | `.claude/agents/moai/manager-lead.md`, `.claude/output-styles/moai/moai.md` |
| 템플릿 미러 | 4 | 위 네 파일의 `internal/template/templates/` 대응본 |
| docs-site | 12 | 4로케일 × {`agent-guide.md`, `manager-lead.md`, `token-budget.md`} |
| **합계** | **20** | |

이 20이 §0의 Tier 판정 첫 행이다. 여덟 개 저장소 파일의 존재와 대상 문장 보유는 이 트리에서
확인했다(§1.3(a)·(b)의 grep, 그리고 미러 4건의 `ls`).

*iteration 2는 이 합계를 **16**으로 적었다 — 층별 행은 그때도 2+2+4+12였으므로 합계만
어긋나 있었다. `research.md` §7의 다른 산식도 같은 20을 낸다.*

**기계 재생성물은 이 표에 세지 않는다.** `plan.md` M4-5가 재생성하는
`internal/template/templates/.codex/agents/moai/manager-lead.toml`은 손편집 대상이 아니라
`make agents-emit`의 산출물이므로(§2.5 `REQ-EPE-015`), **수정 파일**이 아니다. 그 파일이
바뀌었는지는 파일 수가 아니라 `AC-EPE-015`(`make agents-emit-check` exit 0)가 잰다.

### 1.4 예외 폭 — 측정과 미결

primary 체크아웃에서 측정(스냅샷이며 다른 세션이 쓰는 트리라 드리프트한다):

| 집합 | 명령 | 수 |
|---|---|---|
| `verdict.md` (깊이 2) | `find .moai/reports -mindepth 2 -maxdepth 2 -type f -name 'verdict.md'` | 305 |
| 감사 계열 (깊이 2) | 같은 find, `-name 'plan-audit*.md' -o -name 'sync-audit*.md' -o -name 'review-verdict.md'` | 155 |
| `verdict.md` (깊이 3+) | `find .moai/reports -mindepth 3 -name 'verdict.md'` | 19 |

양성 대조(존재하지 않는 파일명, 같은 find 프레임): `0`.

깊이 3+ 19건의 내역: 18건이 `.moai/reports/t264/rescued/**`(구조 아카이브, 살아 있는 카드
판정서가 아니다), 1건이 `.moai/reports/lead/merge-batch-20260913/verdict.md`(리드 배치 판정서).
즉 **깊이 제한은 실재하는 한 가지 형태를 놓친다** — `reports/lead/<batch>/verdict.md`.

### 1.5 [HARD] `.gitignore` 규칙 순서 — `.moai/reports`를 건드리는 다섯 무리 전량

> **순서가 결과를 정한다.** git은 마지막으로 적중한 규칙을 따르므로, 삽입 위치를 정하려면
> **아래에 있는 규칙까지 전부** 알아야 한다. iter1의 분석은 다섯 무리 중 둘만 담고 있었고,
> 빠진 둘 중 하나(`:294`)가 지금 `AC-EPE-004`가 재는 바로 그 경로를 결정한다.

측정: `grep -n 'moai/reports' .gitignore` (이 트리, HEAD `116820f40`, 417줄).

| 무리 | 줄 | 내용 | 성격 |
|---|---|---|---|
| **A** | 107-108 | 주석 (카드 t196, 2026-09-01) | 죽은 negation을 살아 있는 것처럼 설명 |
| | 109 | `!.moai/reports/**/*.log` | **죽은 negation** — `:227`이 부모를 제외해 발화 불가 |
| **B** | 224-226 | 주석 (운영자 지시 2026-09-14) | 블랭킷의 근거 |
| | 227 | `.moai/reports/*` | **블랭킷** |
| | 228-230 | `!plan-audit/` → `plan-audit/*` → `!plan-audit/.gitkeep` | plan-audit 스캐폴드 카브아웃 |
| **C** | 232-233 | 주석 (좁은 예외, 2026-09-14 승인) | |
| | 234-236 | t338 3단 관용구 | `!<dir>/` → `<dir>/*` → `!<dir>/<file>` |
| | 237-241 | t528 3단 관용구 (중첩 디렉터리라 5줄) | |
| | 242-244 | t530 3단 관용구 | |
| | 245-247 | t229 3단 관용구 | |
| **D** | 290-293 | 주석 헤더 (Plan Audit Reports) | |
| | 294-295 | `.moai/reports/plan-audit/*.md` + `!plan-audit/.gitkeep` | **B의 중복 카브아웃** |
| **E** | 376 | `.moai/reports/*.md` | **깊이 1 전용** |

#### 무리 D가 지금 무엇을 결정하고 있는가 (측정)

```
$ git check-ignore -v --no-index .moai/reports/plan-audit/verdict.md
.gitignore:294:.moai/reports/plan-audit/*.md	.moai/reports/plan-audit/verdict.md
```

**`:294`는 후보 삽입 지점(`:227` 직후) 모두보다 아래에 있다.** 따라서 `AC-EPE-004`가 재는
`plan-audit/verdict.md`의 현재 판정은 `:227`이 아니라 `:294`가 낸다. `REQ-EPE-007`의 명시적
재-제외가 `:294`와 중복인지 — 혹은 삽입 위치에 따라 `:294`만으로 충분한지 — 는 M1의
측정 항목이다(`plan.md` §F M1).

#### 무리 E는 `<card>/verdict.md`에 닿지 않는다 (측정)

```
$ git check-ignore -v --no-index .moai/reports/top.md
.gitignore:376:.moai/reports/*.md	.moai/reports/top.md          ← 깊이 1: :376이 결정
$ git check-ignore -v --no-index .moai/reports/tZZZ/verdict.md
.gitignore:227:.moai/reports/*	.moai/reports/tZZZ/verdict.md   ← 깊이 2: :227이 결정
```

`*`는 `/`를 넘지 않으므로 `:376`은 `.moai/reports/` **바로 아래의 `.md`**만 매칭하고
`<card>/verdict.md`에는 닿지 않는다. 그래도 이 표에 넣는 이유는, 열거에서 빠지면 다음 사람이
「그럼 `:376`은 왜 상관없나」를 다시 재야 하기 때문이다.

#### 죽은 negation (무리 A)

```
$ git check-ignore --no-index -q .moai/reports/tZZZ/probe.log ; echo $?
0     # 무시됨 — :109의 !.moai/reports/**/*.log는 발화하지 않았다
```

양성 대조(negation이 실제로 발화하는 경로): `.moai/reports/t338/ac-count-baseline.txt` →
`-v` 출력이 `.gitignore:236:!.moai/reports/t338/ac-count-baseline.txt`, 즉 무리 C가 결정.
음성 대조: `README.md` → `rc=1`.

`:109`가 죽은 이유는 문법 오류가 아니라 **순서와 부모 제외**다: git은 제외된 디렉터리 안의
파일을 negation으로 되살리지 않는다. 저장소가 실제로 쓰는 작동 관용구는 무리 C의 3단이다.

---

## 2. 요구사항 (GEARS)

### 2.1 독트린 문장 좁히기

- **REQ-EPE-001** — doctrine 표면 문서는 추적 인용 대상을 **판정서 파일**로 한정해 서술해야 한다
  (`shall`). 「임의 산출물이 추적 경로로 반출된다」는 형태의 서술을 남겨서는 안 된다
  (`shall not`) — 그 형태는 `.gitignore`가 뒷받침하지 않으므로 거짓 전제다.
- **REQ-EPE-002** — Where 독트린 문장이 **비-판정서 좌표**를 추적 경로로 **주석 달고 있는
  경우**, 그 주석은 제거되거나 「머신-로컬 스크래치」로 정정돼야 한다(`shall`). 비-판정서
  좌표는 **두 형태**다 — 이 둘이 곧 `AC-EPE-022`가 docs-site에서 세는 두 형태이며, 같은 형태를
  한 표면에서는 위반으로 세고 다른 표면에서는 비워 두지 않기 위해 여기서 맞춘다:
  - **(i) 비-판정서 파일명** — `<check>.log`, `M<n>-report.md`, `M<n>.<AC-id>.log`
  - **(ii) 파일명 없는 디렉터리 단독** — `.moai/reports/<card-id>/`. 디렉터리는 파일이 아니므로
    「인용은 파일 하나를 이름 붙인다」는 같은 독트린의 요구와도 어긋나고, 그 디렉터리가 추적
    대상이라는 주장은 `:227` 아래에서 거짓이다.

  대상 좌표는 `.claude/output-styles/moai/moai.md:354,561`(형태 i)과
  `.claude/agents/moai/manager-lead.md:152,154,161`(형태 i) 그리고
  `.claude/agents/moai/manager-lead.md:63`(형태 ii)이다.
  **`:161`은 「tracked」라는 낱말을 담지 않는다** — 그 줄은 접기 행 형식 안에서
  `M<n>-report.md`를 **해소 가능한 인용 경로로 제시**하고, 바로 위 `:154`의 「only the tracked
  path does」가 그것을 추적 주장으로 만든다. 그래서 이 요구사항이 `:154`를 함께 지목한다.
- **REQ-EPE-003** — 실제 판정서를 반출하는 표면(`plan-auditor.md:601`, `sync-auditor.md:108`)의
  참/거짓 상태는 좁히기 이후에 **측정되고 기록돼야 한다**(`shall`). 폭 독법 (A)에서는 둘 다
  참이어야 하고, (B)에서는 둘 다 거짓인 채로 남으며 그 잔존이 명시 기록돼야 한다(§3.0). 이
  요구사항은 어느 독법도 전제하지 않는다 — 요구하는 것은 **상태가 알려질 것**이다.
- **REQ-EPE-004** — When 좁히기가 선행 SPEC의 요구사항과 충돌할 때, 그 충돌은 기록돼야 한다
  (`shall`). 기록 방식은 §4를 따른다. 완료된 SPEC의 요구사항 **본문**을 고쳐서는 안 된다
  (`shall not`).

### 2.2 `.gitignore` 좁은 예외

- **REQ-EPE-005** — `.gitignore`는 판정서 파일을 추적 대상으로 되살리는 예외를 가져야 한다
  (`shall`). 예외의 폭은 **판정서 파일에 한정**되며, 카드 본문·로그·바이너리·중간 산출물을
  되살려서는 안 된다(`shall not`).
- **REQ-EPE-006** [해소 — 운영자 판정 2026-09-21, 독법 (B)] — 「판정서」 집합의 경계는 두 독법
  중 하나로 확정돼야 한다(`shall`): (A) 감사 판정서 계열 전체, (B) `verdict.md` 단독.
  **운영자가 (B)를 확정했다**: 되살아나는 것은 `verdict.md` **한 파일**이고, 깊이는 카드
  디렉터리 **한 단계**(`.moai/reports/<card-id>/verdict.md`)다. 감사 판정서 계열(독법 A)은
  채택되지 않았고, `t264/rescued` 아카이브도 들어오지 않는다.
  **출처**: 2026-09-21 run 단계 배차문(리드 경유 운영자 판정). 기록은 `progress.md` §E.2.0.
  [HARD] **이 선택은 결함을 남기기로 하는 결정이며 실수가 아니다** — §3.0의 귀결표가 그 대가를
  이미 적어 두었고, 잔존 결함(두 [HARD] Export mandate가 거짓인 채로 남는 것)은
  `progress.md` §E.3 Gaps와 별도 카드 **t1059**가 맡는다. 예외를 넓혀 그 문장들을 참으로 만드는
  것은 **결정된 교환을 조용히 수선하는 것**이므로 하지 않는다.
  §3은 그대로 남는다 — §3.0이 두 독법의 수치와 귀결을, §3.3이 권고가 (A)였음을 싣는다.
  **권고가 결정이 아니라는 것이 이 자리에서 실제로 성립했으므로**, 그 기록은 결정이 난 뒤에도
  지워지지 않는다.
- **REQ-EPE-007** — `.moai/reports/plan-audit/`는 예외 **이후에도** 무시된 채로 남아야 한다
  (`shall`). `.moai/docs/audit-artifact-convention.md:50-53`이 그 디렉터리를 판정서 목적지로
  FORBIDDEN으로 표시하며, 예외가 그것을 되살리면 금지된 목적지가 추적 대상이 된다.
  **현재 그 경로를 결정하는 것은 `:227`이 아니라 `:294`다**(§1.5 측정) — 이 요구사항이 명시적
  재-제외를 요구하는지, `:294`가 삽입 위치와 무관하게 충분한지는 M1이 측정으로 가른다.
- **REQ-EPE-008** [HARD] — When 예외 규칙이 `.gitignore`에 삽입될 때, 그 효과는 **제자리에서**
  — 이 트리의 실제 `.gitignore`와 그 규칙 순서(§1.5의 다섯 무리 전량) 아래에서 — 재측정돼야
  한다(`shall`). run 단계의 어떤 인수조건도 §5의 픽스처 결과를 baseline-attribution으로
  인용해서는 안 된다(`shall not`) — 픽스처는 **가설이며 증거가 아니고**, 규칙 순서를 재현하지
  않으므로 결론을 낼 수 없다. 픽스처의 유일한 역할은 무엇을 재야 하는지의 범위를 정하는 것이다.
- **REQ-EPE-009** — 예외는 이미 추적 중인 54개 파일(`git ls-files .moai/reports | wc -l` → 54)과
  충돌해서는 안 된다(`shall not`).

### 2.3 깊이 (이 SPEC이 명시적으로 결정한다)

- **REQ-EPE-010** — 예외는 **한 단계 깊이**(`.moai/reports/<dir>/<파일>`)만 되살려야 한다
  (`shall`). 되살아나는 파일명 집합은 REQ-EPE-006이 정하며 이 요구사항은 그 집합을 **정하지
  않고 참조만 한다** — 이 요구사항이 결정하는 것은 **깊이 하나뿐**이다.
  `.moai/reports/lead/<batch>/verdict.md` 형태(§1.4에서 실재 1건 관측)를 되살릴지는
  §7의 「Out of Scope — 깊이 2 이상 판정서」에서 명시적으로 범위 밖에 둔다.
- **REQ-EPE-011** — When 깊이 제한이 살아 있는 판정서를 놓치는 경우, 그 사실은 `.gitignore`
  주석에 적혀야 한다(`shall`) — 다음 사람이 「깊이가 왜 1단계인가」를 다시 측정하지 않도록.
  주석은 놓치는 **형태**(`reports/lead/<batch>/verdict.md`)를 이름 붙여야 한다(`shall`).

### 2.4 죽은 규칙 정리

- **REQ-EPE-012** — `.gitignore:109`의 `!.moai/reports/**/*.log`는 **철회돼야 한다**(`shall`).
  수리(3단 관용구로 되살리기)는 로그를 추적 대상으로 만들어 REQ-EPE-005의 폭 제한을 위반하므로
  선택지가 아니다.
- **REQ-EPE-013** — `:107-108`의 주석은 정정돼야 한다(`shall`). 정정된 주석은 (a) 그 negation이
  왜 제거됐는지와 (b) 로그가 의도적으로 무시된다는 사실을 서술해야 하며, 무언가를 고친 것처럼
  읽혀서는 안 된다(`shall not`).

### 2.5 미러와 기계 방출

- **REQ-EPE-014** — Where 저장소 루트 사본이 수정되는 경우, `internal/template/templates/`
  아래 대응 미러도 같은 변경을 받아야 한다(`shall`).
- **REQ-EPE-015** — `internal/template/templates/.codex/agents/moai/*.toml`은 **손편집돼서는
  안 되며**(`shall not`), `make agents-emit`으로 재생성돼야 한다(`shall`).
- **REQ-EPE-016** — When 템플릿 측 수정이 제안될 때, 그 내용은
  `.moai/docs/template-internal-isolation-doctrine.md` §25.1의 금지 콘텐츠 클래스(SPEC ID, REQ
  토큰, 감사 인용, 내부 날짜, 커밋 SHA, macOS 편향 경로, CLAUDE.local 참조)를 담아서는 안 된다
  (`shall not`). 이 요구사항은 `internal/template/templates/` 아래에만 적용된다 — docs-site는
  그 뿌리 밖이므로 중립성 규칙의 대상이 아니다.

### 2.6 docs-site 공개 문서 (iteration 2에서 편입)

- **REQ-EPE-017** — Where docs-site 페이지가 좁혀진 독트린 문장의 번역본을 담고 있는 경우, 그
  페이지도 좁혀진 문언으로 갱신돼야 한다(`shall`). 대상은 §1.3(c)가 측정한 12파일이며,
  비-판정서 경로를 추적 경로로 서술하는 문장을 남겨서는 안 된다(`shall not`).
- **REQ-EPE-018** — When docs-site 페이지가 수정될 때, 네 로케일(en / ko / ja / zh)이 **같은
  변경 안에서** 함께 수정돼야 한다(`shall`). 각 로케일의 본문은 그 로케일의 자연스러운
  원어여야 하며, 영어 문장이 비-영어 페이지에 그대로 들어가서는 안 된다(`shall not`).
  근거는 저장소의 docs-site 4-로케일 동기화 독트린(`.moai/docs/docs-site-i18n-rules.md`)이다.

---

## 3. 「판정서」 경계 — 해소된 축 (a)

> **[해소 2026-09-21 — 독법 (B)]** 운영자가 (B)를 확정했다(§2.2 `REQ-EPE-006`, 출처는
> `progress.md` §E.2.0). **아래 절들은 결정 이전의 상태 그대로 보존한다** — 절 제목의
> 「미해소」 표식과 권고 문구를 포함해서다. 결정이 났다고 두 독법의 수치와 권고 기록을 치우면,
> 다음 사람이 「왜 (A)가 아니었는가」를 처음부터 다시 재야 한다. 읽는 법은 하나다:
> **§3은 결정에 이르는 근거이고, 결정 자체는 §2.2에 있다.**
>
> 원문 유지: 폭은 2026-09-14 운영자 지시의 예외 폭이므로 운영자 소관이다. 레인이 리드에
> 에스컬레이트했고 plan 단계는 이 축이 열린 채로 닫혔다. 답은 run 단계 배차문으로 왔다.

### 3.0 [미해소 축 — 운영자 대기] 두 독법의 측정 수치와 귀결

> **이 표가 왜 여기 있는가.** 폭은 운영자 소관이고 답이 없다. 한쪽만 싣고 진행하면 운영자가
> 다른 쪽을 고를 때 SPEC을 다시 써야 하고, 그때는 「왜 그쪽이 아니었는지」의 측정이 이미
> 사라진 뒤다. **두 독법을 나란히 싣는 것은 중복이 아니라, 어느 쪽으로 결정되든 SPEC을 다시
> 쓰지 않기 위한 구조다.** 영향 범위는 §3.4가 전수 열거한다.

| | (A) 감사 판정서 계열 전체 | (B) `verdict.md` 단독 |
|---|---|---|
| 되살아나는 수 (깊이 2, primary 스냅샷) | `verdict.md` 305 + 계열 155 | `verdict.md` 305 |
| `plan-auditor.md:601` [HARD] Export mandate | 참이 된다 | **거짓으로 남는다** |
| `sync-auditor.md:108` [HARD] Export mandate | 참이 된다 | **거짓으로 남는다** |
| 예외가 되살리는 파일명 개수 | 6종(초안) | 1종 |
| `.gitignore` 추가 줄수 | 많음(파일명당 negation 1줄) | 적음 |
| 남는 결함 | 없음(이 카드 범위 내) | 같은 결함이 감사 판정서 층에 잔존 |

(B)를 고르는 것은 **결함을 남기기로 하는 결정**이지 실수가 아니다 — 폭을 좁게 유지하는 대가로
두 [HARD] 문장이 거짓인 채 남는다는 것을 알고 고르는 것이다. 그 경우 REQ-EPE-003은 「참이어야
한다」가 아니라 「거짓임이 기록돼야 한다」로 해소되고, 잔존 결함은 별도 카드가 된다.

### 3.1 권고에 적용한 판별식

> **예외의 폭은, 독트린이 행위자에게 반출을 의무화한 파일 집합과 일치해야 한다.**

근거: 이 카드의 결함은 「독트린이 추적을 단언하는데 ignore가 뒷받침하지 않는다」이다. 의무화된
반출 대상 중 하나라도 예외 밖에 남으면 **같은 결함이 한 층 아래에서 그대로 재생산된다** — 그
표면의 [HARD] 문장이 여전히 거짓이 된다.

### 3.2 권고 판별식이 고르는 집합 (독법 A) — **(B) 확정 아래에서는 해당 없음**

> **[NOT APPLICABLE — 2026-09-21 독법 (B) 확정]** 아래 표는 **(A)를 택했을 경우** 되살아났을
> 집합이다. 운영자가 (B)를 확정했으므로 이 집합은 구현되지 않았고, `verdict.md` 한 파일만
> 되살아난다. **표를 지우지 않는 이유**: 「어느 집합이 검토됐고 어디서 갈렸는가」가 사라지면
> 다음 사람이 `plan-audit.md`·`sync-audit.md`의 누락을 새 결함으로 오인해 같은 측정을 다시
> 한다. 이 표는 그것이 **고려된 뒤 채택되지 않은 것**임을 말한다.

| 파일명 | 의무화 좌표 | 포함? |
|---|---|---|
| `verdict.md` | `kanban-dispatch.md` § Completion is read (리드 판정서) | 예 |
| `plan-audit.md`, `plan-audit-iter<N>.md` | `plan-auditor.md:601` [HARD] Export mandate | 예 |
| `sync-audit.md`, `sync-audit-verdict*.md` | `sync-auditor.md:108` [HARD] Export mandate | 예 |
| `review-verdict.md` | `audit-artifact-convention.md` §Where에 **없음** | 예(경계) |
| `report.md`, `evidence.md`, `pr-body.md`, `*.log` | 의무화 없음 | 아니오 |

`review-verdict.md`는 경계 사례다 — 규약 §Where가 열거하지 않지만 `/moai review` 게이트의
판정서이고 이름 자체가 verdict다. **포함하되 별도로 표시**한다: 운영자가 §Where 열거만을
경계로 삼기를 원하면 이 한 항목만 빼면 된다.

### 3.3 권고와 그 근거 — 그리고 권고가 아닌 것

**권고: 독법 (A).** 근거는 §3.1의 판별식이다 — 독법 (B)는 `plan-auditor.md:601`과
`sync-auditor.md:108`의 두 [HARD] 문장을 거짓인 채로 남기고, 그것이 바로 이 카드가 고치려는
결함의 형태다.

**권고는 결정이 아니다.** 운영자가 (B)를 고르면 §3.0의 귀결표가 그 선택의 대가를 이미 적어
두었고, 그것이 이 절이 존재하는 이유다.

### 3.4 [미해소 축 — 영향 봉쇄] 폭 선택이 영향을 주는 곳, 전수 열거

> **이 절이 없으면 plan 단계가 닫히지 않는다.** 열린 축 하나가 문서 전체를 잠정적으로 만드는
> 것을 막는 유일한 방법이 「어디까지 번지는지를 명시적으로 세는 것」이다.

의존에는 **두 등급**이 있고, 이 구분이 iteration 2의 정정이다. iter1은 아래 셋째 행을 「폭과
무관」 목록에 넣었는데, 그 요구사항이 문언에 `verdict.md`라는 폭 보유 토큰을 담고 있었으므로
과장이었다. 문언은 고쳐졌고(REQ-EPE-010), 관계는 **참조하되 결정하지 않는다**로 명시된다.

| 항목 | 의존 등급 | 형태 |
|---|---|---|
| AC-EPE-002 | **결정 의존** | 되살아나야 하는 파일명 집합이 폭의 함수다. **제거 불가능** — 폭이 곧 그 AC의 입력이다. AC는 「확정된 집합 전원」으로 매개변수화돼 있고 어느 쪽도 선취하지 않는다. |
| AC-EPE-013 | **결정 의존** | 생존 표면의 참/거짓이 폭의 함수다. AC는 **어느 쪽을 요구하지 않고** §3.0 귀결표와 일치하는지만 잰다 — (A)면 참, (B)면 거짓이 기록돼야 PASS. |
| REQ-EPE-010 | **참조 의존 (결정하지 않음)** | 「확정된 집합의 각 파일」을 한 단계 깊이에서 되살리라고 말한다. 집합의 내용은 REQ-EPE-006이 정하고 이 요구사항은 그것을 **읽기만 한다**. 폭이 (A)든 (B)든 이 요구사항의 문언은 바뀌지 않으며, 검증(AC-EPE-005)은 깊이만 잰다. |

**[2026-09-21] 위 두 결정 의존 항목의 입력이 채워졌다** — 폭은 (B)다. `AC-EPE-002`의 확정
집합은 `{verdict.md}` 단독이므로 **평가 가능**해졌고, `AC-EPE-013`은 (B) 갈래
(전원 `rc=0` + 잔존 결함의 명시 기록)로 평가된다. 이 표의 **의존 등급 자체는 바뀌지 않는다** —
바뀐 것은 입력이 미정이었다가 정해졌다는 것뿐이다.

폭과 완전히 무관한 것: REQ-EPE-001~005, 007~009, 011~018, AC-EPE-001, 003~012, 014~023.
특히 누수(AC-EPE-004)·깊이(AC-EPE-005)·죽은 negation(B.4)·문언 좁히기(B.5)·미러(B.6)·
충돌 기록(B.7)·docs-site(B.9)는 폭이 어느 쪽으로 가도 같은 판정을 낸다.

---

## 4. 선행 SPEC과의 충돌 기록 방식 (결정 (b))

### 4.1 적용한 판별식

> **완료된 SPEC의 요구사항 본문은 그때 무엇이 결정됐는지에 대한 움직이지 않는 증거다. 나중
> 결정이 그것을 좁힐 때, 양쪽 끝에서 발견 가능해야 하되 어느 쪽도 다시 쓰지 않는다.**

관측 기록을 나중 사실에 맞춰 고치면 갱신이 아니라 증거 조작이다. 동시에, 충돌을 적지 않으면
독자가 두 SPEC을 나란히 읽고 스스로 발견해야 한다 — 그건 기록이 아니라 운이다.

### 4.2 채택한 3요소 (전부 수행한다)

1. **새 SPEC의 승계 절** — 이 문서 §4.3이 REQ-ECC-001 / REQ-ECC-003을 지목하고, 각각의 좁혀진
   대체 문언을 적는다. (아래)
2. **완료 SPEC의 HISTORY 행 + frontmatter** — `SPEC-EVIDENCE-CITATION-CANON-001`의 HISTORY에
   **추가만 하는** 행 한 개와, frontmatter에
   `partially_superseded_by: [SPEC-EVIDENCE-PATH-EXCEPTION-001]`. 요구사항 본문은 건드리지
   않는다.
3. **양방향 `related_specs`** — 양쪽 frontmatter에 서로를 적는다.

### 4.3 승계 절 (superseding clause)

- **REQ-ECC-001** ("문서가 인용하는 증거 경로는 추적되는 경로여야 한다. 정본 위치는
  `.moai/reports/<card-id>/`") → **부분 승계**. 좁혀진 대체 문언: *판정서 파일*이 인용되는 경우
  그 경로는 추적돼야 하며 정본 위치는 `.moai/reports/<card-id>/<판정서>`다. 판정서가 아닌
  산출물에 대해서는 추적성이 주장되지 않는다.
- **REQ-ECC-003** ("When 어떤 산출물이 판정 근거로 인용될 때, 인용 이전에 추적 경로로 반출해야
  한다") → **부분 승계**. 좁혀진 대체 문언: 인용 전 반출 의무는 판정서 반출 경로에 대해
  유지되고, 판정서가 아닌 산출물에 대해서는 「인용하지 않는다」쪽으로 해소된다 — 이는
  REQ-ECC-003의 이미 존재하는 후반부(「반출하지 않기로 한 것은 인용하지 않는다」)와 같은 방향이다.
- **승계되지 않는 것**: REQ-ECC-002, 004~011은 이 SPEC이 건드리지 않는다.

### 4.4 명시적으로 기각한 대안

**대안: REQ-ECC-001 본문을 제자리에서 고친다.** 기각 이유 — 2026-08-31에 무엇이 결정됐는지의
기록이 사라지고, 09-14 지시와의 충돌 자체가 보이지 않게 된다. 충돌이 있었다는 사실이 이 카드의
가장 중요한 산출물이다.

---

## 5. 예외 규칙의 알려진 누수 (run 단계가 닫아야 한다)

레인 측정 §7이 픽스처에서 재현한 소박한 와일드카드 형태:

```
reports/*
!reports/*/
reports/*/*
!reports/*/verdict.md
```

| 픽스처 경로 | add 가능? | 의도대로? |
|---|---|---|
| `reports/t100/verdict.md` | 예 | 예 |
| `reports/t100/report.md` | 아니오 | 예 |
| `reports/t100/probe.log` | 아니오 | 예 |
| `reports/t100/sub/verdict.md` | **아니오** | **아니오** — 깊이 제한 |
| `reports/plan-audit/verdict.md` | **예** | **아니오** — FORBIDDEN 디렉터리 재포함 |

**[HARD] 이 표는 가설이지 증거가 아니다.** 4줄짜리 `.gitignore` 픽스처였고, 이 트리의 417줄
파일이 가진 규칙 **순서**(§1.5의 다섯 무리)를 전혀 재현하지 않았다. 특히 픽스처에는 무리 D
(`:294`)가 없는데, 제자리에서는 **그 규칙이 `plan-audit/verdict.md`의 현재 판정을 내고 있다**.
순서가 결과를 정하므로 두 결론(누수·깊이) 어느 쪽도 이 측정으로 확립되지 않는다.
REQ-EPE-008이 제자리 재측정과 **이 표를 baseline으로 인용하는 것의 금지**를 함께 의무화한다.
이 표의 유일한 역할은 **무엇을 재야 하는지 범위를 정하는 것**이다.

- 누수 → REQ-EPE-007이 명시적 요구사항으로 만들고, §1.5가 `:294`의 현재 역할을 측정으로 적는다.
- 깊이 제한 → REQ-EPE-010이 한 단계로 **결정**하고, §7이 깊이 2 이상을 명시적 범위 밖에 둔다.

---

## 6. 수용 기준 요약

전량은 `acceptance.md`. 이 SPEC의 모든 ignore-동작 인수조건은 §1.2의 두 판별식 중 하나를 쓰고,
**실패할 수 있는 양성 대조를 함께 명령한다**. 문서 표면 인수조건은 §1.3의 측정을 재현하되
**정확 수치 일치를 단언하지 않는다** — primary 체크아웃 수치와 docs-site 행 수는 둘 다
드리프트한다.

---

## 7. 범위 밖 (out of scope)

### Out of Scope — 2026-09-14 블랭킷 철회

- `.gitignore:227`의 `.moai/reports/*` 자체를 제거하거나 무력화하는 것. 운영자 지시이며
  2026-09-20에 재확인됐다. 이 SPEC은 그것을 통과하는 **좁은 구멍**만 뚫는다.
- `.moai/reports/` 전체를 추적 대상으로 되돌리는 어떤 변형도 포함되지 않는다.

### Out of Scope — 중복 카브아웃(무리 D) 정리

- `:294-295`가 `:228-230`과 같은 일을 두 번 한다는 사실은 §1.5가 기록한다. **그 중복을 제거하는
  것은 이 카드의 일이 아니다.**
- 이유: 제거하면 `plan-audit/`의 판정이 어느 규칙에서 나오는지가 바뀌고, 그 변화는 이 카드가
  뚫는 예외의 효과 측정과 뒤섞인다. 한 번에 한 가지만 움직여야 제자리 측정(REQ-EPE-008)이
  무엇을 잰 것인지 말할 수 있다. 중복 정리는 별도 카드다.

### Out of Scope — 깊이 2 이상 판정서

- `.moai/reports/lead/<batch>/verdict.md`(§1.4에서 실재 1건) 및
  `.moai/reports/t264/rescued/**`(18건, 구조 아카이브)를 예외로 되살리는 것.
- 이유: 깊이를 한 단계 더 열면 `.gitignore` 규칙이 한 쌍 더 늘고, 되살아나는 집합이
  아카이브 트리(18/19건)까지 포함해 「판정서만」이라는 폭 제한이 사실상 무너진다. 리드 배치
  판정서 1건은 REQ-EPE-011에 따라 주석으로 명시되고, 필요하면 별도 카드가 된다.

### Out of Scope — 이미 추적 중인 54건의 소급 처리

- `git ls-files .moai/reports` → 54건은 블랭킷 이전부터 인덱스에 있다. 이들을 untrack하거나
  재정렬하지 않는다. 예외는 **지금부터 추가되는 파일**의 운명만 바꾼다.

### Out of Scope — 비-판정서 산출물의 추적

- `report.md`(15), `evidence.md`(15), `pr-body.md`(4), `*.log`, 카드 본문, 바이너리.
- 이유: 운영자가 폭을 판정서로 명시했고, 넓히면 블랭킷이 막으려던 것(원격으로 나가는 로컬
  증거 더미)이 그대로 돌아온다.

### Out of Scope — docs-site의 비-주장 언급

- `.moai/reports`를 언급하지만 **추적 주장을 담지 않는** docs-site 페이지. 측정:
  `grep -rl 'moai/reports' docs-site/` → 41파일 중 §1.3(c)의 12파일만이 추적 주장을 담는다.
- 이유: 나머지 29파일은 경로를 예시로 들 뿐 거짓 전제를 싣지 않는다. 손대면 4-로케일 동기화
  부담만 늘고 고치는 결함이 없다.

### Out of Scope — 기계 소비자 경로

- `.moai/state/verify/**`(스냅샷 저장소, SSE 감시 소스). `SPEC-EVIDENCE-CITATION-CANON-001`
  REQ-ECC-006의 명시적 예외로 이미 서술돼 있고 이 SPEC은 건드리지 않는다.

### Out of Scope — 새 Go 코드·훅·CLI

- 예외의 집행은 `.gitignore`와 문서 문언으로 끝난다. ignore 규칙을 기계적으로 지키는 가드가
  필요하다는 판단이 서면 별도 카드다.

### Out of Scope — 서브에이전트 쓰기 범위 가드

- 이 카드의 선행 시도에서 서브에이전트가 지시된 범위 밖(`.gitignore`)을 편집했고, 그것을
  막는 장치가 없다는 사실이 드러났다.
- **그 장치는 여기서 설계하지 않는다.** 카드화 여부가 운영자 판정 대기 중이며, 이 SPEC은
  요구사항·인수조건·계획 단계 어느 것으로도 그것을 제안하지 않는다.
- 사건은 **증거 경로 내구성의 동기**로만 참조된다 — `.gitignore`가 이 카드가 고치려는 바로
  그 파일이었기 때문에, 파괴가 눈에 띄지 않았다면 run 단계의 모든 수치가 스텁을 상대로
  측정됐을 것이다.

### Out of Scope — 폭 경계 결정 자체 (plan 단계 기준; 2026-09-21에 운영자가 결정했다)

- §3의 (A)/(B) 선택은 운영자 소관이며 이 SPEC도 plan 단계도 그것을 확정하지 않는다.
- **[해소]** 운영자가 run 단계 배차문으로 **(B)**를 확정했다(§2.2 `REQ-EPE-006`). 항목을
  지우지 않고 남기는 이유는 **결정의 소유권이 어디였는지**가 이 카드의 기록이기 때문이다 —
  SPEC이 스스로 정하지 않았다는 사실은 결정이 난 뒤에도 참이다.

### Out of Scope — 감사 판정서 계열의 반출 (독법 (B)의 귀결)

- (B) 확정에 따라 `plan-audit.md` · `plan-audit-iter<N>.md` · `sync-audit.md` ·
  `sync-audit-verdict*.md` · `review-verdict.md`는 **무시된 채로 남는다.** 따라서
  `plan-auditor.md:601`과 `sync-auditor.md:108`의 두 [HARD] Export mandate는 이 카드가 착지한
  뒤에도 거짓이다.
- 그 문장들을 참으로 만드는 일(독트린 문언 수리)은 **별도 카드 t1059**의 몫이며, 이 카드는
  예외를 넓혀 그것을 대신하지 않는다. 측정과 귀결 기록은 `progress.md` §E.3 Gaps 1번 항목에
  있다.

---

## 8. 잔여 위험

- **이 SPEC 자신의 증거가 반출돼 있지 않다.** `.moai/reports/t1039/lane-measurements.md`와
  `.moai/reports/t1039/plan-audit.md`는 `:227`로 무시된다 — 이 카드가 열려는 바로 그 규칙이다.
  감사 자신의 판정서가 `plan-auditor.md:601`의 [HARD] Export mandate를 만족하지 못하며,
  이것이 이 결함의 **6번째 실사례**다. **[2026-09-21 정정 — (B) 확정의 귀결]** 이 문장은
  원래 「카드가 착지하기 전까지」라고 적었으나, (B)에서는 `plan-audit*.md`가 **착지한 뒤에도**
  되살아나지 않는다. 즉 이 위험은 이 카드가 닫는 것이 아니라 **남기기로 결정된 것**이며,
  `progress.md` §E.3 Gaps 1번이 다섯 파일을 이름 붙여 측정으로 기록한다.
  **워크트리를 폐기하면 그 다섯의 유일본이 사라진다.**
- §1.4의 primary 체크아웃 수치는 다른 세션이 쓰는 트리의 스냅샷이며 드리프트한다.
- §5의 누수·깊이 측정은 픽스처 측정이다. 제자리 재측정 전까지 결론이 아니다(REQ-EPE-008).
- ~~REQ-EPE-006의 폭은 **미해소**다.~~ **[해소 2026-09-21 — 독법 (B)]** 영향 범위는 §3.4가 두
  등급으로 전수 열거하며, 그 두 항목의 입력이 (B)로 채워졌다. plan 단계는 이 축이 열린 채로
  닫혔고, 답은 run 단계 배차문으로 왔다. **남는 위험은 폭의 미결이 아니라 폭의 귀결이다** —
  두 [HARD] Export mandate가 거짓인 채로 남고, 그 수리는 t1059가 맡는다.
- **`.gitignore` 줄수는 트리마다 다르다.** 이 워크트리(`WT-evidence-path`, `develop` 계열)는
  417줄이고, `main`에 체크아웃된 트리는 319줄이다. 「417줄」을 트리 귀속 없이 인용하면 다른
  트리에서 재는 사람에게 거짓 기준선이 된다. 모든 줄수·줄번호 인용은 트리를 명명한다.
- **docs-site 편입이 Tier를 올렸고, 그것이 감사 임계도 올렸다.** Tier L의 PASS 임계는 0.85이며
  산출물이 5종이다. 범위를 정직하게 잡은 대가가 더 높은 통과선이라는 것은 의도된 교환이지만,
  다음 iteration의 채점 기준이 iter1과 다르다는 사실은 기록해 둔다.
- **docs-site 빌드는 이 카드가 아직 재지 않았다.** 4-로케일 편집이 warning-free hugo build를
  유지하는지는 run 단계에서 처음 측정된다(AC-EPE-022).
