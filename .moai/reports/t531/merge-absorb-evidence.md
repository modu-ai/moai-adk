# t531 — 흡수 병합 증거 (통합 창 안에서 측정)

## 창

`moai integration acquire --name lane-6` → `release-integration window acquired by
e7d89180-… on WT-claudelocal-push-model`, holder `lane-6` (pid 18254),
since `2026-09-08T05:04:28Z`.

## 흡수 직전 tip 재측정 — 리드가 준 수치를 믿지 않고 다시 쟀다

배차문이 로컬 develop 을 `a4855f0b2` 라고 알렸고, 내가 앞서 잰 `ee194493f` 는 낡았다고 경고했다.
흡수 **직전에** 다시 쟀다:

```
$ git fetch origin develop
$ git rev-parse --short develop        → a4855f0b2
$ git rev-parse --short origin/develop → a4855f0b2
$ git rev-list --count --left-right origin/develop...develop
0	0
```

`0 0` 이므로 로컬 develop 이 최신이고 갱신이 필요 없다. **이 확인 자체가 이 카드가 수리한 §4.1 의
첫 단계다** — 흡수 전에 로컬 develop 이 뒤처졌는지부터 본다(F2 수리). 이번에는 최신이었으므로
`gitflow-lane-protocol.md` §11 의 갱신 경로로 갈 필요가 없었다.

## 흡수

```
$ git merge develop --no-edit
CONFLICT (content): Merge conflict in CHANGELOG.md
```

**충돌 1건, `CHANGELOG.md`.** 부가적 충돌이다 — 같은 자리에 HEAD 쪽은 우리 항목을, develop 쪽은
다른 카드 항목을 각각 추가했다. 의미 충돌이 아니므로 **합집합**으로 해소했다.

해소 전 겹침 검사(겹치면 합집합이 성립하지 않는다):

```
HEAD쪽(381-382)  'CLAUDELOCAL-PUSH-MODEL' 적중 = 1
develop쪽(384-402) 'CLAUDELOCAL-PUSH-MODEL' 적중 = 0
develop쪽 항목 = 6건 (CODEX-SKILL-PATH-READBACK / CODEX-SKILL-PATH-SLASH /
                      TODO-HOME-TEMP-GUARD / AC-COLLECTOR-ANCHOR /
                      DOCTOR-STAT-SEAM / SEAM-GREENFIELD)
```

해소는 마커 3줄(`<<<<<<< HEAD` / `=======` / `>>>>>>> develop`)만 제거하고 양쪽 내용을 모두
보존했다. 스크립트가 **세 마커 위치를 각각 단언**하고, HEAD 쪽 우리 항목이 정확히 1건이며
develop 쪽에 우리 항목이 없음을 단언한 뒤에만 썼다 — 하나라도 어긋나면 아무것도 쓰지 않는다.

해소 검증:

```
충돌 마커 잔여      = 0
우리 항목(파일전역) = 1
develop 6건 보존    = 6
충돌 파일 잔여      = 0
```

흡수 병합 커밋: **`aee19460b`** (`Merge branch 'develop' into WT-claudelocal-push-model`).
`MERGE_HEAD` 부재로 병합 완료 확인, 워킹트리 변경 0.

## 병합 트리 재측정 — CARD_BASE 가 이동했으므로 다시 유도했다

```
$ git merge-base origin/develop HEAD → a4855f0b2   (흡수 전 bce6d7e08 에서 이동)
```

흡수 뒤에도 흡수 전 base 를 쓰면 develop 이 가져온 변경까지 이 카드 몫으로 세게 된다.

**develop 이 `CLAUDE.local.md` 를 건드렸는가**: `git diff --stat bce6d7e08 a4855f0b2 --
CLAUDE.local.md` 무출력 — 건드리지 않았다. 이 카드의 프로브는 develop 변경의 영향을 받지 않는다.

| AC | 병합 트리 실측 | 대조군 |
|---|---|---|
| 001 | 슬라이스 **0** · 파일전역 **0** | base 슬라이스 **2** |
| 002 | `분기하는 트리` **2** | 존재형 |
| 003 | restore **1** · arm1 **3** · 창 **5행** · arm2 **2** · arm3 **0** | 뮤턴트 M-003(아래) |
| 004 | `폐기`+`main` 한 줄 동시 **3** | 존재형 |
| 005 | `push origin develop` **2** | 백업 **3**(선행 감사 관측) |
| 006 | `미커밋` **2** | 존재형 |
| 007 | `.go` **0** · 범위 전체 **10** | 재유도된 `a4855f0b2` 기준 |
| 008 | 백업 접근 범위 밖 — **unmeasured** | — |

F2 수리 표지 `뒤처` **1**. SPEC `status: completed`, `version: "0.1.1"`.

## 자기 보고 — 이 창에서 낸 측정 결함 2건

흡수 명령을 낼 때 다음 두 줄을 함께 찍었는데, **둘 다 측정이 아니었다**:

- `--- rc=$? ---` — 파이프의 마지막 명령(`tail`)의 종료코드를 잡았지 `git merge` 의 것이 아니다.
  충돌이 났는데 `rc=0` 으로 찍혔다.
- `echo "(clean = 충돌 없음)"` — 조건 없이 찍히는 고정 문자열이다. 충돌이 있어도 「충돌 없음」이
  나온다.

둘 다 이 카드가 내내 지적해 온 **공허한 초록**의 형태다. 실제 상태는 `git status` 의
`UU CHANGELOG.md` 로 드러났고, 이후 판정은 고정 문자열 대신 `git rev-parse -q --verify MERGE_HEAD`
와 `git diff --name-only --diff-filter=U | wc -l` 로 바꿔 측정했다.
