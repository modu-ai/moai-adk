# t410 Discovery — `AC-EH3-007` 명제 3 DRIFT 1건, 재현과 원인 재판정

카드: t410 (Class C) · 브랜치 `WT-drift-false-positive` · 트리 `.claude/worktrees/t410`
기준: develop 흡수 후. 바이너리는 **이 트리에서** `go build -o /tmp/moai-t410 ./cmd/moai` (§2.2 도구 출처 귀속).

리드 지시: *"카드 문면을 근거로 쓰지 말고 재현을 근거로 써라."* 그대로 했고, **재현 결과가 t382 의 진단 절반을 뒤집었다.**

---

## Claim

1. **오탐이 맞다.** `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` 은 실제로 닫혔는데 판정기가 `in-progress` 로 추론한다.
2. **그러나 t382 가 적은 원인 2개 중 1개는 거짓이다.** `--grep` 이 본문을 매치하는 것은 원인이 아니다 — subject 필터가 그것을 정상적으로 무력화한다.
3. **실제 기제는 단일 원인이다**: 진짜 close 커밋이 워커에게 **보이지 않아**, 워커가 같은 SPEC 의 더 오래된 `docs(...)` 커밋까지 흘러내려 거기서 `in-progress` 를 집는다.
4. **1건짜리 사건이 아니다.** 같은 형태가 현재 drift 표 203행 중 **15행 내외**에 있다(아래 범위 주의).
5. 따라서 수리 대상은 이 행이 아니라 **판정기 또는 close 규약**이다.

## Evidence

### R1 — 워커 축자 추적 (`r1-walker-trace.log`)

판정기의 필터를 그대로 재적용해 `--grep` 창을 한 줄씩 추적했다(`internal/spec/t410_probe_test.go`).

```
SKIP  7cffb9717  [subject lacks id]  feat(SPEC-HANDOFF-AUTORESUME-001): plan-phase 산출물
                                     extracted=[SPEC-HANDOFF-AUTORESUME-001]
SKIP  e979a4d13  [subject lacks id]  chore(SPEC group C): Mx-phase close (...)
                                     extracted=[]
SKIP  8630b40d0  [subject lacks id]  feat(SPEC-V3R6-SPEC-LINT-CLEANUP-001): ...
SKIP  b6723b495  [subject lacks id]  chore(hook): resolveMemoryDir 실제 구현
ADOPT 97a36b5a2  [status=in-progress]  docs(SPEC-V3R6-SESSION-HANDOFF-AUTO-001): /moai mx Step C
getGitImpliedStatus(...) = "in-progress", err=<nil>
```

**t382 원인 1 반증.** t382 는 *"`--grep` 이 본문까지 매치해 다른 SPEC 의 plan 커밋 `7cffb9717` 을 최신으로 집는다"* 고 적었다. 실제로는 `7cffb9717` 이 **정상적으로 걸러진다** — `commitMatchesSPECID` 가 subject 에서 추출한 토큰 집합(`[SPEC-HANDOFF-AUTORESUME-001]`)에 대상 ID 가 없기 때문이다. `--grep` 의 본문 매치는 후보 창을 넓힐 뿐이고, 본문 전용 매치는 그 다음 필터가 전부 떨어뜨린다.

**t382 원인 2 확증, 그리고 그것이 유일 원인.** `e979a4d13` 은 `extracted=[]` 로 걸러진다 — 결합 범위 `chore(SPEC group C)` 에 완전한 SPEC-ID 토큰이 없다. 워커는 계속 내려가 같은 SPEC 의 `docs(...)` 커밋을 만나고 거기서 `in-progress` 를 집는다.

이 구분은 수리 공간을 바꾼다. 만약 원인 1이 참이었다면 "본문 매치를 끄면" 되지만, 실제로는 **본문을 넓게 읽는 것이 오히려 수리 후보**이고 동시에 원인 1이 상상했던 오탐을 진짜로 만드는 위험이다(§ 수리 후보 A).

### R2 — close 커밋이 실재한다

```
e979a4d13  chore(SPEC group C): Mx-phase close (status implemented→completed, 2026-06-02)
  본문: 그룹 C 5개 SPEC 4-phase 일괄 종료
        - SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Mx verdict EVALUATE-PASS, 1 @MX:TODO deferred
  frontmatter: status: completed · updated: 2026-06-02   ← 커밋 날짜와 정확히 일치
```

커밋 하나가 **5개 SPEC** 을 닫는다. frontmatter 가 앞선 것이 아니라 판정기가 못 보는 것이다.

### R3 — 현재 표에서 재현 (`r3-drift-row.log`)

```
SPEC-V3R6-SESSION-HANDOFF-AUTO-001  completed  in-progress  DRIFT
```

### R4 — 파급 범위 (`r2-blast-radius.log`)

```
drift rows examined:         203
LOOSE body-only close hits:   33
TIGHT (line declares close):  15
```

**두 수치 모두 정확한 값이 아니라 경계값이다.** 프로브(`blast-radius.sh`)는 "본문이 ID 를 언급함"과 "본문이 그 ID 의 close 를 선언함"을 완전히 가르지 못한다.

- LOOSE 는 과대계상이다. 측정된 반례: `a83934d55` 의 본문 `depends_on: SPEC-AUTONOMY-TIERS-001 (completed)` 는 **언급이지 close 가 아니다.**
- TIGHT 도 그 반례를 **여전히 잡는다** — 같은 줄에 `completed` 가 있기 때문이다. 즉 **15는 상한이고, 알려진 오탐이 최소 1건 포함돼 있다.**
- LOOSE-only 18건 중 일부는 진짜 close 일 수 있다(squash 본문의 하위 subject 배치에 따라 갈린다). 즉 **하한도 확정돼 있지 않다.**

정확한 수는 건별 판정으로만 나온다. 이 카드가 그 판정을 범위 안에 둘지는 plan 에서 정한다.

> **[정정 — plan 단계 실측]** 위에서 TIGHT 15 의 알려진 오탐을 **1건**(`a83934d55`)이라고 적었으나,
> plan 단계에서 두 주체가 독립적으로 더 찾았다. `manager-spec` 이 `fb8aff006`×2 · `80dea9684` 를,
> `plan-auditor` 가 같은 3건에 `a83934d55` 를 더해 **최소 4건**을 실측했다. 즉 TIGHT 15 의 상한
> 성격은 내가 적은 것보다 강하고, 실제 대상 수는 15보다 **눈에 띄게 작다**. 하한은 여전히 미확정이다.
> 대상 수는 SPEC-DRIFT-CLOSE-BODY-001 의 건별 판정 AC 가 산출한다 — 15 를 목표치로 쓰지 않는다.

> **[정정 — 후보 B 는 새 기제가 아니다]** 위 §수리 후보 B 를 "코드가 이미 close-infix 개념을 갖고
> 있다" 로만 적었는데 과소 서술이다. `drift.go:223` 에 이미 **FALLBACK-ONLY 결합범위 슬롯**이
> 동작하고 `drift_index.go:212 inMemCombinedScopeClose` 가 그 in-memory 짝이다. `chore(SPEC group C)`
> 가 거기 안 걸리는 이유는 게이트 (a) 가 `deriveScopePrefix(specID)` — SPEC-ID 에서 유도한 접두 —
> 를 요구하는데 `SPEC group C` 는 아무것에서도 유도되지 않기 때문이다(코드 판독으로 확인).
> 따라서 B 는 **기존 슬롯의 세 번째 게이트**이고, 그 자리에 놓이면 후보 A 의 위험(다른 SPEC 의
> 비-completed 상태를 본문에서 채택)이 규율이 아니라 **구성상** 닫힌다 — fallback 의 출력이
> `completed` 하나뿐이기 때문이다.

> **[정정 — TIGHT/LOOSE 가 가르는 축 자체가 틀렸다]** 내 프로브는 "ID 를 담은 본문 줄이
> `close|completed|Mx verdict` 도 담는가"로 갈랐다. 그런데 **두 부류가 그 토큰을 공유한다** —
> `- SPEC-X: Mx verdict EVALUATE-PASS`(진짜 close)와 `depends_on: SPEC-Y (completed)`(단순 언급)가
> 같은 술어를 만족한다. 실제 판별축은 **어휘가 아니라 위치**다: 그 ID 가 줄의 주어인가, 아니면
> 다른 키(`depends_on:` / `residual_debt` …)가 앞서는가. 스크립트 주석이 TIGHT 를 "act on these"
> 라고 적은 것은 틀렸다 — TIGHT 는 두 부류가 공유하는 토큰으로 거른 상위집합일 뿐이다.
> 결과: **15 = 미판정 11 + 실측 오탐 4**. 미판정 11 은 커밋 3개(`e979a4d13` · `2f449e189` ·
> `7beda68a5`)에서 나온다. LOOSE-only 18 행은 표본 4건만 봤고 전부 언급이었다 — 표본이지 전수가 아니다.

> **[정정 — §1.4 세 번째 위반 형태의 사례가 그 형태를 입증하지 않는다]** "단일 범위인데 형제를 함께
> 닫음"의 사례로 `docs(SPEC-INTERNAL-TEST-001): ... 3-phase close` 를 들었으나, 그 본문은 `-002` 를
> **잔여 부채의 소유자**로 지목할 뿐 닫지 않는다. 즉 이 사례는 그 형태의 증거가 아니라 위 오탐 4건
> 중 하나다. 형태 자체는 존재할 수 있으나 **여기 인용한 사례로는 입증되지 않는다.**

> **[정정 — 지배적인 진짜-close 형태가 조사에서 통째로 빠졌다]** 내가 예로 든 `- SPEC-X: <verdict>`
> 는 소수 형태다. 실제 close 의 다수는 **squash 하위 subject** — 머지 본문이 개별 커밋 subject 를
> 그대로 담은 줄이다:
> ```
> * docs(SPEC-GLM-KEY-INPUT-001): sync-phase artifacts — 3-phase close   (2f449e189)
> * chore(SPEC-V3R6-CODERABBIT-ADOPTION-001): ... — 3-phase close        (7beda68a5)
> ```
> 이 줄들은 **subject 였다면 기존 필터 체인이 이미 받아들였을** 모양이다. 수리의 성격이 바뀐다:
> 후보 B 는 이들을 위한 새 텍스트 술어를 발명할 필요가 없고, 그 줄을
> `shouldSkipCommitTitle` → `commitMatchesSPECID` → `ClassifyPRTitle` 에 **그대로 흘려보내면** 된다.
> 내 조사의 예시만으로 술어를 짰다면 남은 후보 11행 중 **8행을 놓쳤을 것이다.**

> **[추가 실측 — close-infix 어휘가 대상을 떨어뜨린다]** `internal/spec/transitions.go:80-82` 가
> 인정하는 리터럴은 `3-phase close` / `4-phase close` / `mx-phase audit-ready` 셋뿐이다. 확정 대상
> `e979a4d13` 의 subject 는 소문자화하면 `chore(spec group c): mx-phase close` 로 **셋 중 어느 것도
> 담지 않는다.** subject 쪽 close 신호를 후보 게이트로 쓰면 이 카드의 유일한 확정 대상이 구조적으로
> 탈락한다. `Mx-phase close` 는 어휘에 없는 **네 번째 역사적 형태**다.

관측된 위반 subject 형태 3종:
| 형태 | 예 |
|---|---|
| 명시적 결합 범위 | `chore(SPEC group C): Mx-phase close` · `docs(specs): batch sync-phase close — 5 B-grade SPECs` |
| 범위 없는 산문 subject | `Close out 2 SPECs with 3-phase lifecycle completion (doc-only)` |
| 단일 범위인데 형제를 함께 닫음 | `docs(SPEC-INTERNAL-TEST-001): sync-phase artifacts + 3-phase close` (본문에서 `-002` 도 닫음) |

`.claude/rules/moai/development/spec-frontmatter-schema.md` § Close-subject full-ID mandate 가 이 형태를 이미 금지한다. 금지가 있는데도 15건 내외가 존재한다 = **규약이 기계적으로 강제되지 않는다.**

## Baseline-attribution

- 트리: `.claude/worktrees/t410`, develop 흡수 후
- 바이너리: `/tmp/moai-t410`, **이 트리에서 빌드** (설치본 아님)
- `spec drift` 는 전부 `--no-cache` (t382 가 관측한 HEAD-SHA 캐시 공허-초록 함정 회피)
- 워커 추적은 판정기 자신의 함수(`commitMatchesSPECID` / `shouldSkipCommitTitle` / `ClassifyPRTitle` / `getGitImpliedStatus`)를 in-package 로 직접 호출

## 측정 환경 주의 — 203 이라는 수를 오독하지 말 것

판정기는 `cachedMainBranch()` → `main` 을 본다. 이 저장소의 작업은 `develop` 에 착지하고 release PR 로만 `main` 에 도달하므로, `main..develop` 이 **1,362 커밋** 벌어져 있다. 203 DRIFT 행 대부분은 "close 가 아직 main 에 없음"이며 **이 카드의 결함이 아니다.** 이 카드의 대상은 그중 "close 가 main 에 있는데도 못 보는" 부분집합뿐이다.

## 수리 후보 (plan 이 판정할 것 — 여기서는 확정하지 않는다)

**A. 워커가 본문에서 ID 를 추출하도록 확대.** `e979a4d13` 이 보이게 된다. 단 `7cffb9717` 도 함께 보이게 되어 **t382 가 상상했던 오탐을 진짜로 만든다** — 다른 SPEC 의 plan 커밋이 최신으로 채택될 수 있다. LSGF-001 이 막으려던 바로 그 형태다. 단독으로는 채택 불가.

**B. close 선언 커밋에 한해 본문 조회.** subject 가 close-infix 를 갖는 커밋에서만 본문의 ID 열거를 권위 있는 목록으로 취급. A 의 위험을 열지 않으면서 `e979a4d13` 을 살린다. 코드가 이미 close-infix 개념을 갖고 있다(`closeInfixMatch`).

**C. 백필 커밋.** SPEC 별 full-ID + close-infix 를 갖는 새 `chore(SPEC-XXX-NNN): ... close` 커밋을 쌓는다. `shouldSkipCommitTitle` 의 D5 가드가 **backfill + close-infix 조합을 일부러 skip 하지 않게** 만들어져 있어, 코드가 이미 이 경로를 상정한다. 판정기 무변경. 대신 합성 커밋이 이력에 남고 15건 내외를 손으로 만들어야 한다.

**D. 규약의 기계적 강제.** 향후 재발만 막고 기존 15건은 안 고친다. B 또는 C 와 병행해야 의미가 있다.

## Gaps — 관측하지 않은 것

- **건별 판정을 하지 않았다.** 15/33 은 프로브의 경계값이고, 각 행이 진짜 오탐인지는 `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` 1건만 커밋 본문·frontmatter 날짜까지 대조해 확정했다. 나머지는 미판정이다.
- `ClassifyPRTitle` 이 `docs(...)` 를 `in-progress` 로 분류하는 것이 그 자체로 옳은지 판정하지 않았다. 다만 이 행에 대해서는 무관하다 — `docs` 를 빈 상태로 만들어도 워커는 더 내려가 `test(...)` 를 만나고, 그래도 `completed` 는 못 나온다.
- 후보 A~D 중 어느 것도 구현·측정하지 않았다. plan 소관.
- `SPEC-ERA-H3-NARROWING-001` 은 `completed` 다. **소급 재판정하지 않는다** — t382 의 착지는 유효하다. 원인 서술 정정이 필요하면 HISTORY 에 사유와 함께 남긴다.
- t469 가 같은 형태를 다룬다는 리드의 언급은 확인하지 않았다(해당 SPEC 미조회).

## Residual-risk

- 파급 범위의 하한이 확정되지 않아, 후보 C(백필)를 고르면 **작업량이 15건보다 많을 수도 적을 수도 있다.**
- 후보 B 는 "close-infix 를 가진 subject" 라는 텍스트 술어에 의존한다. 그 술어를 넓히면 A 의 위험이 뒷문으로 들어온다.
- 판정기가 `main` 을 보는 한, 이 저장소의 git-flow(작업은 develop) 아래에서 drift 표는 계속 대부분 소음이다. 그 축은 이 카드 밖이지만, 이 카드의 수리 효과를 표에서 **보이지 않게** 만든다.
