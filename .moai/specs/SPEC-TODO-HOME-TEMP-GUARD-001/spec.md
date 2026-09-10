---
id: SPEC-TODO-HOME-TEMP-GUARD-001
title: "생산 홈 폴백의 임시-디렉터 기원 거부 — SPEC-STATE-ANCHOR-001 §5 결정 이행"
version: "0.1.5"
status: completed
created: 2026-09-08
updated: 2026-09-10
amendment_of: SPEC-TODO-HOME-TEMP-GUARD-001
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "todo-queue, home-pollution, temp-dir-guard, t536, state-anchor-followup"
tier: M
era: V3R6
related_specs: [SPEC-STATE-ANCHOR-001, SPEC-WEB-TODO-QUEUE-001]
---

# SPEC: 생산 홈 폴백의 임시-디렉터 기원 거부

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.5 | 2026-09-10 | manager-spec | **t574 — 제자리 개정(in-place amendment).** AC-THG-006에 양의 방향 절 추가 — 생산 임시 루트 집합 소속(production temp-root set membership): 스텁이 아닌 `defaultTempRoots()`가 `os.TempDir()`·`/tmp`·`/var/folders`를 담고 원소가 정확히 3개임을 AC가 요구한다. **코드 변경 없음(no code change).** 출처: **t536 sync-audit F1**(뮤턴트 M-AUD-2, t574가 M-574로 재현). `status: completed → in-progress`, `amendment_of` 자기 참조. 요구사항 본문·AC 개수(8)·§D.0 매핑은 불변. 아래 § Amendments 참조 |
| 0.1.4 | 2026-09-08 | manager-spec | plan-audit 4차(FAIL — D14·D16 닫힘, D15 미닫힘) 수리 + **리드의 Tier 판정 이행**. ⑴ **접힘 해제(unfold)** — REQ-THG-002에 접었던 판별식 주입 이음매를 **REQ-THG-009로 분리**하고 `tier: S → M`. 근거: 접힌 상태에서 그 절은 주어가 *분류 동작*인 요구사항 안의 종속절이라, REQ-THG-002를 정리하는 후속 편집자가 **어떤 AC도 눈치채지 못한 채** 지울 수 있다 — 그런데 이음매가 없으면 `t.TempDir()` 아래에서 홈 폴백 가지를 밟을 픽스처가 존재하지 않아 **가드가 시험 불가 상태로 출하된다**. 이 SPEC은 공허한 초록을 세 번 냈고(D2·D10·D14), *관측 가능성을 보장하려고 존재하는* 요구사항은 독립적으로 지목 가능해야 한다. D9 선례와는 모양이 다르다: D9의 절은 같은 사건의 **결과**(거부가 무엇을 돌려주는가)였고, 이음매는 그 사건을 **관측하는 도구**다 — 결과와 도구는 다른 축이다. ⑵ **D18** — §C.1의 「전수 8건」이 전수가 아니었다. 인용된 grep 패턴이 대문자 `R`로 시작해 CLI의 소문자 래퍼 `resolveTodoQueueRoot()`(`internal/cli/todo.go:71`)를 **원리상 볼 수 없었다**(실측: `/usr/bin/grep -c 'func Resolve' internal/cli/todo.go` = **0**). 두 축(대소문자 무관 심볼 · 간접 소비자)으로 재열거해 **14건**으로 정정(A행 4→6 · C행 5→7). ⑶ **AC-SA-011 교차-SPEC 판정** — `SPEC-STATE-ANCHOR-001`(completed)의 §5·§5:115·REQ-SA-011 2번째 가지 셋이 같은 방향을 가리킨다: 철회가 아니라 **그 요구사항이 문서화한 두 번째 가지로의 이행**이며, 본 카드가 §5가 위임한 결정의 **이행 주체**다. 그 SPEC은 한 글자도 수정하지 않는다. ⑷ **D19** — C행 판정식이 5건 중 4건에 공허했다(부재 단언 2 · 단언 대상이 가드 반환값과 동일 2). 관측 대상을 반환값에서 **판별식의 직접 단언**(`TempOriginReason(base)`의 `isTemp == false`)으로 옮기고, **그 판정식이 실제로 RED가 되는 조건을 M2가 실연**하도록 종료 조건에 넣었다. ⑸ **D21/D22** — 이음매의 산출 AC 매트릭스 행 정정, 교차 절의 산출 And절 추가. D17은 범위 밖으로 열린 채 |
| 0.1.3 | 2026-09-08 | manager-spec | plan-audit 3차(FAIL 0.8125) 수리 — **D14**(`acceptance.md:134`이 뮤턴트 아래 갈래 (b)의 FAIL 기제를 거꾸로 적었다: 「stat 불가」라 했으나 `Rename`의 목적지가 곧 `BacklogPathForRoot(반환 루트)`라 stat은 **성공**한다 → 실제 FAIL 사유 셋으로 재작성하고 읽힘 단언이 (b)에서 뮤턴트에 공허함을 명시), **D15**(가드가 깨뜨리는 기존 테스트가 1건으로 적혀 있었으나 실측 8건 — 확실히 깨짐 4·공허화 5·배치 의존 1 → `plan.md` §C·§F M2에 전수 등재 + 테스트별 처분, `SPEC-WEB-TODO-QUEUE-001` AC 산출 테스트는 보존 지정. 그 과정에서 **판별식 주입 이음매**가 요구사항임이 드러나 REQ-THG-002에 접었다 — §8 D15 판정 기록), **D16**(「임시 기원」과 「홈 해석 불가」가 배타적이라는 §8의 분리 근거가 거짓 — 교차에서 `base`가 이긴다는 절과 술어 배치 순서 구속을 REQ-THG-001에 추가). §8에 공허한 초록 3연발(D2·D10·D14) 관측 등재 |
| 0.1.2 | 2026-09-08 | manager-spec | plan-audit 2차(FAIL 0.875) 수리 — **D9**(REQ-THG-001의 대체 루트가 「루트」 계층이 아니었다: `resolveStateDir(base,false)`는 이미 상태 디렉터라 소비자가 `BacklogPathForRoot`를 덧붙이면 `base/.moai/state/todo/.moai/state/todo/backlog.json`이 된다 → **대체 루트를 `base`로 정정**하고 성질로 기술, `base`를 「덜 정밀」로 기각했던 종전 판단이 뒤집혔음을 plan.md D12에 정정 기록), **D10**(AC-THG-001이 문자열 동일성만 단언해 D9가 GREEN 통과했다 → `BacklogPathForRoot(반환 루트)`가 픽스처 큐를 가리키는지를 단언), **D11**(`/var/tmp` 제외 근거의 미측정 사전확률을 판단형으로 격하 + 재검토 트리거를 관측 가능한 형태로 교체). **U1 결정 반영**: 운영자가 형태 (b)(안내 후 계속)를 선택(2026-09-07, 리드 경유) → REQ-THG-006·AC-THG-005의 두-형태 표현을 결정된 형태로 **가지치기**(두-형태 표현은 U1이 열려 있던 동안 의도적이었고 지금 불필요해진 것이지 틀렸던 것이 아니다). §8에 `todo_root.go:121`의 기존 계층 불일치를 **관측**으로 등재(본 SPEC은 고치지 않는다 — 별도 카드 후보) |
| 0.1.1 | 2026-09-08 | manager-spec | plan-audit(FAIL 0.75) 수리 — D1(거부 시 대체 루트를 REQ-THG-001이 고정), D2(AC-THG-001 Given 2갈래 분리 + 거짓 RED 서술 정정), D3(REQ-THG-006·AC-THG-005의 U1 편향 제거, Then절과 GREEN 주석 판정 대상 일치), D6(`/var/tmp` 제외 근거 §8 등재), D7(임시 기원 단정을 측정 범위로 축소 + 미복원 1건 Gap 등재), D8(launcher 인용 `456-485`). §4.1에 키 역산 확증 추가. D5는 **기각** — `(Event-detected)`는 GEARS 정본 5패턴이며 §6 서두에 라벨 출처를 명시 |
| 0.1.0 | 2026-09-08 | manager-spec | 최초 작성 — 카드 t536 plan-phase. `SPEC-STATE-ANCHOR-001` §5가 표면화만 하고 결정하지 않은 미결 설계 결정을, 운영자 결정(2026-09-07, 리드 경유)에 따라 **임시-디렉터 기원 한정 거부**로 확정하고 SPEC으로 전사 |

카드: **t536** (Class C · **Tier M** — 0.1.4에서 S→M, 이음매 분리에 따른 리드 판정. §8). 측정 기준 트리: `.claude/worktrees/t536`, 브랜치 `WT-home-fallback` @ `412c8cb14`. 아래의 모든 좌표·측정치는 이 트리에서 나온 것이다.

## Amendments

### 개정 1 — AC-THG-006 양의 방향 절: 생산 임시 루트 집합 소속 (카드 t574, 2026-09-10)

| 항목 | 값 |
|---|---|
| 직전 completed 판 | `0.1.4` |
| prior_completed_sha | `029ab039f` — `status: in-progress → completed` 전이와 3-phase close를 실은 sync 커밋. `git show 029ab039f -- .moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md`가 `-status: in-progress` / `+status: completed`를 보인다(t574 트리에서 확인). 이 값을 `progress.md` sync 절의 SHA 칸에 채운 백필 커밋은 `586f26f2a`다 |
| 개정 종류 | 제자리 개정(in-place). 후속 SPEC이 아니므로 `amendment_of`가 자기 자신을 가리킨다 |
| 개정 착수 트리 | `.claude/worktrees/t574` · 브랜치 `WT-temp-roots-ac` · HEAD `95ba9deb2` |

**근거.** t536 sync-audit F1이, `defaultTempRoots()`를 `{os.TempDir()}` 하나로 줄인 뮤턴트 M-AUD-2가 `internal/kanban`·`internal/cli` 스위트를 전부 초록으로 통과한다고 보고했다. 그 수리로 `1d091087b`가 `TestDefaultTempRoots_Membership`을 추가해 이 뮤턴트는 이제 테스트 층에서 잡힌다. 그러나 **그 테스트를 요구하는 인수 기준은 없었다.** 카드 t574가 같은 모양의 뮤턴트 M-574로 재현했다 — AC가 이름 부른 kanban 테스트는 전부 PASS하고 `TestDefaultTempRoots_Membership`만 FAIL한다(`.moai/reports/t574/repro-summary.md`, `.moai/reports/t574/mutant-kanban.txt`).

원인은 AC-THG-006의 형태에 있다. 양의 방향 절(임시 루트 자신과 하위는 임시)은 **스텁한 루트**를 상대로 판정되고, 생산 집합을 상대로 하는 서브테스트는 음의 방향(`/tmpfoo`는 `/tmp` 밖) 하나만 묻는다. 두 판정 모두 뮤턴트가 만족한다. 그래서 그 테스트를 지우거나 약화시키는 후속 편집을 막는 AC가 하나도 없었다. REQ-THG-002는 집합을 {`os.TempDir()`, `/tmp`, `/var/folders`}로 이미 고정하고 있으므로, 결함은 요구사항 층이 아니라 인수 기준 층에 있다.

**결정(운영자).** 새 AC를 만들지 않고 AC-THG-006에 양의 방향 절을 더한다. AC 개수 8과 `acceptance.md` §D.0의 `AC-THG-006 maps REQ-THG-002`는 그대로다.

**범위.**
- `acceptance.md` — AC-THG-006 본문에 생산 집합 소속 절, 확장된 판정 명령, RED 대조 문단을 더하고, §D 매트릭스 AC-THG-006 행의 GREEN 칸과 머리의 판본 표기를 갱신한다.
- `spec.md` — frontmatter(`version`·`status`·`updated`·`amendment_of`), HISTORY 행, 이 § Amendments.
- `progress.md` — plan-phase 신호 절에 개정 신호 블록을 덧붙인다.

**범위 밖.**
- 요구사항 본문(REQ-THG-001..009) — 한 글자도 바꾸지 않는다.
- 코드와 테스트 — `internal/kanban/temp_origin.go`·`internal/kanban/temp_origin_test.go` 무변경. 판정에 쓰는 두 테스트는 이미 트리에 있다.
- AC-THG-006 외의 AC, `plan.md`, 다른 SPEC.
- `progress.md`의 run·sync 절 기존 블록 — 첫 close의 귀속된 기록이므로 고쳐 쓰지 않는다. 개정 이후의 재측정과 재close는 각 단계 소유자가 그 아래에 덧붙인다.

## 1. 문제 — 측정된 형태

`internal/kanban/todo_root.go`가 백로그 큐 루트를 해석한다. `primaryCheckoutRoot(base)`(`:96`)가 git에게 저장소의 common dir을 묻고, git이 답하지 못하면(base가 저장소 안이 아니면) `homeTodoQueueRoot`(`:112`) → `~/.moai/todo/<basename>-<sha256[:4]>/`로 **fail-open**한다.

이 fail-open 자체는 의도된 설계이고 파일 주석이 문서화한다(`:62-63`): "a project without git metadata still gets exactly one queue".

**위험은 fail-open보다 좁다.** 비git base가 **임시 디렉터**(테스트·CI·샌드박스)일 때, 임시 디렉터는 나중에 사라지지만 그 경로를 키로 만든 홈 큐는 영원히 남는다 — 아무도 도달할 수도 정리할 수도 없는 고아다.

### 1.1 도달 가능성 — 세 갈래로 이미 확인됨

가드를 설계하기 전에 그 가지가 실제로 도달 가능한지부터 재야 한다(도달 불가능한 가지를 가드하면 죽은 코드만 는다). 도달한다:

1. **기계적** — `internal/kanban/todo_root_test.go:98` `TestResolveTodoQueueRoot_FallbackNoGit`이 비git base + 스텁 `HomeDirFn`으로 이 가지를 이미 실행한다.
2. **생산 바이너리** — `~/.moai/todo/t203-probe-d7a16ea2`(2026-08-24 01:20)의 기원 경로는 **키 역산으로 복원됐다**. `TodoQueueProjectKey`(`todo_root.go:197-203`)의 유도식 `basename + "-" + sha256(abs)[:4]`을 후보 루트에 대입하면 `sha256("/tmp/t203-probe")[:4] = d7a16ea2`로 정확히 일치한다 — 비git 임시 디렉터 `/tmp/t203-probe`에서 `moai todo`를 호출해 생긴 것이다. 이 1건으로 생산 축의 임시 기원은 추정이 아니라 **측정**이다. 같은 시기의 `~/.moai/todo/proj-325ca0b6`(2026-08-23 20:23)은 후보 루트 대조로 **복원되지 않았고, 임시 기원인지 아닌지는 미검증이다**(§8) — 아니라고 판정한 것이 아니다. (역산은 plan-audit이 먼저 수행했고 본 SPEC이 재유도로 확인했다. 종전 판본은 두 건 **모두**를 임시 기원으로 단정했는데, 그 단정은 `SPEC-STATE-ANCHOR-001` §5에서 근거 없이 옮겨온 것이며 측정보다 넓었다.)
3. **테스트 실행의 규모** — `001-*` 형태 디렉터 341개(Go `t.TempDir()`가 만드는 모양).

총계(`find`로 측정 — 이 셸의 `ls`는 `ls -la` 별칭이라 파이프 계수를 오염시킨다): `~/.moai/todo` 아래 디렉터 343개, 파일 946개, 11 MB. `backlog.json` 211개의 카드 텍스트는 "first card" ×209, "fix the drift found in t151" ×1, "e2e probe card one" ×1 — 전부 픽스처이고 운영자 데이터는 0건이다.

### 1.2 §5의 「표본 2건」 정정

`SPEC-STATE-ANCHOR-001` §5는 이 상황을 "표본 2건"으로 기술했다. 그 수는 **생산 기원 한 쌍만** 센 것이다. 실제 총계는 위 343개이며, 그중 341개는 테스트 기원이다. 본 SPEC은 §5를 잇는 SPEC이므로 정정된 수치를 배경에 기록한다 — 결론(생산 폴백이 살아 있다)은 그대로이고 규모 서술만 교정된다.

### 1.3 테스트 축은 이미 멈춰 있다 — 남은 것은 생산 축이다

테스트 기원의 유입은 t422 fail-loud 가드가 이미 끊었다: 커밋 `e7a078970`(03:13)이 develop에 착지한 시각은 2026-09-02 04:14(`a1cba5425`). 테스트 기원 고아 중 가장 최신은 2026-09-02 03:53이고 이후 신규는 없다. **본 카드가 닫는 것은 생산 축뿐이다.**

## 2. 소비자 — 두 곳, 그리고 이미 분리된 순수/부작용 경계

`ResolveTodoQueueRoot`에 도달하는 생산 소비자는 둘이다(`/usr/bin/grep -rn 'ResolveTodoQueueRoot' internal/ --include='*.go'`로 측정; 나머지 히트는 전부 주석과 테스트):

| 소비자 | 좌표 | 진입점 |
|---|---|---|
| CLI | `internal/cli/todo.go:72` | `kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())` |
| 웹 콘솔 | `internal/web/todo_queue_read.go:33` | `kanban.ResolveTodoQueueRoot(projectRoot)` |

파일이 문서화한 분할(`todo_root.go:8-17`)이 여기서 결정적이다: `ResolveTodoQueueRoot`은 **순수**하다(어느 가지에서도 `MkdirAll`/`Rename`/`WriteFile` 없음 — 콘솔이 임포트하는 것), `ResolveTodoQueueRootAdopting`만이 adopt-not-shadow 이관을 수행하고 **쓰기 부작용이 도달 가능한 유일한 진입점**이다.

## 3. 운영자 결정 — 범위가 정확하다

**임시-디렉터 기원에 대해서만 홈 큐를 거부한다.** 진짜 비git 프로젝트는 홈 폴백을 **유지한다** — 문서화된 "exactly one queue" 가용성을 회수하지 않는다. (운영자 확인 2026-09-07, 리드 경유. 리드의 1차 전달이 이를 「모든 비git base」로 넓혔던 것은 **기각된 판독**이며 본 SPEC의 어느 요구사항도 그 넓은 형태를 취하지 않는다.)

거부 시의 행동: 홈 큐를 만들지 않고, 호출자에게 **이유를 알린다**(조용한 성공이 아니라 안내).

## 4. 판별식 설계 — 순진한 접두사 비교로는 안 되는 이유

이 기계(macOS)에서 측정된 형태:

```
TMPDIR=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/
readlink /var  → private/var
readlink /tmp  → private/tmp
```

`os.TempDir()`은 **미해석** `/var/folders/...` 형태를 돌려주는 반면, `filepath.EvalSymlinks`를 거친 호출자 경로는 `/private/var/folders/...`를 지닌다. 한쪽만 해석된 두 경로의 접두사 비교는 **조용히 불일치**하고, 가드는 발화하지 않는다. 덧붙여 `TodoQueueProjectKey`(`todo_root.go:197`)는 `filepath.Abs`만 쓰고 심링크를 해석하지 않으므로, 리졸버에 도달하는 형태는 호출자가 무엇을 했느냐에 달려 있다.

따라서 판별식은 **양쪽을 같은 규칙으로 정규화한 뒤** 비교한다. 정규화 규칙은 이 리포에 이미 있는 형태를 그대로 쓴다 — `internal/cli/launcher.go:456-485` `resolveSymlinks`: 경로가 **존재하면** `filepath.EvalSymlinks`, **존재하지 않으면** 어휘적 `filepath.Clean`. 그 함수의 doc과 `@MX:REASON`이 이유를 문서화한다: 비존재 경로에서 `EvalSymlinks`의 동작이 GOOS마다 갈라져(윈도우는 존재하는 접두사를 부분 해석) 접두사 매칭이 깨진다.

비교는 **경로 구성요소 단위 포함**이지 생문자열 접두사가 아니다 — `/tmpfoo`가 `/tmp` 아래로 오분류되지 않게 한다.

불확실할 때의 방향은 **fail-open**이다: 정규화 실패·판별 불가는 "임시 아님"으로 보고한다. 오분류의 두 방향은 대칭이 아니기 때문이다 — 거짓 음성은 오늘의 동작을 그대로 남기지만, 거짓 양성은 진짜 프로젝트의 큐를 회수한다.

**거짓 양성 자세.** 남는 거짓 양성 모집단은 「임시 루트 아래에 사는 **비git** 진짜 프로젝트」로 좁다. git 저장소는 임시 루트 아래에 있어도 `primaryCheckoutRoot`가 먼저 답하므로 폴백 가지에 **도달하지 않는다**(`todo_root.go:64-69`/`:76-83` 직접 판독). 본 SPEC은 이 좁은 잔여 위험을 **수용**하며, 그 대가로 REQ-THG-006의 안내 문구가 사용자에게 무슨 일이 일어났는지와 회복 경로(다른 위치에서 실행하거나 `git init`)를 알린다.

### 4.1 이 설계를 사후 확증한 측정 — 오염은 미해석 `/tmp` 철자로 도달했다

위 두 결정(루트 집합에 `/tmp`를 포함할 것, 양쪽을 정규화할 것)은 §1.1의 키 역산이 **사후적으로 확증한다**. 측정된 유일한 생산 오염 `t203-probe-d7a16ea2`의 기원은 `/private/tmp/t203-probe`가 **아니라 미해석 철자 `/tmp/t203-probe`**다(`sha256("/private/tmp/t203-probe")[:4] = a8822e43`으로 실제 키와 불일치, `sha256("/tmp/t203-probe")[:4] = d7a16ea2`로 일치). 귀결이 둘이다:

1. **루트 집합에 `/tmp`가 반드시 있어야 한다.** 이 기계에서 `os.TempDir()`은 `/var/folders/...`를 돌려주므로 `os.TempDir()`만으로는 이 오염을 잡지 못한다 — 즉 루트 집합의 `/tmp` 항목은 중복 방어가 아니라 유일한 포착 수단이다.
2. **양쪽 정규화가 실제로 일을 한다.** 호출자 경로가 `/tmp/...`로 도달하고 루트가 `/private/tmp`로 해석돼 있으면(또는 그 반대면) 비교는 **조용히 어긋난다** — 방어적 코드가 아니라 이 오염을 잡는 데 필요한 조건이다.

출처: 이 SPEC의 plan-audit(`plan-audit.md` 측정 기록 #4)이 역산을 수행했고, 본 SPEC이 재유도로 같은 값을 얻었다. 저자는 이 역산 없이 §4의 설계에 도달했으므로 이것은 설계의 **근거**가 아니라 사후 **확증**이다 — 그리고 §4가 예측한 실패 형태(미해석 철자로의 도달)가 실제 생산 데이터에서 관측된 형태와 같다는 뜻이다.

## 5. 대표 mutant — 이 SPEC의 AC가 어떻게 공허하게 통과될 수 있는가

1. **미해석 접두사만 비교하는 판별식** — `/private/var/folders/...` 형태로 도달한 호출에서 발화하지 않는다. AC-THG-002가 두 철자 모두를 단언해 잡는다.
2. **생문자열 접두사 매칭** — `/tmpfoo`를 임시로 오분류한다. AC-THG-006.
3. **모든 비git base 거부로의 확대** — 기각된 판독을 그대로 구현. AC-THG-003이 비임시 비git base의 홈 폴백 **유지**를 단언해 잡는다.
4. **가드가 조용히 통과** — 임시 기원에서 홈 디렉터가 안 생기는 것은 가드 없이도 관측될 수 있다. AC-THG-007의 뮤턴트(판별식 무력화 시 오염 재현)만이 판별 증거다. **단, 뮤턴트가 오염을 재현하려면 픽스처가 오염 경로에 실제로 도달해야 한다**: `adoptLocalTodoQueue`는 base에 project-local `backlog.json`이 있을 때만 `MkdirAll`에 도달하므로(`todo_root.go:172-175`), 로컬 큐를 세우지 않은 픽스처에서는 뮤턴트를 걸어도 홈 디렉터가 생기지 않아 **뮤턴트 자체가 공허해진다**. AC-THG-001이 Given을 두 갈래로 나누고 AC-THG-007이 그중 (b)를 쓰는 이유가 이것이다.
5. **순수성 파괴** — 가드를 넣으며 콘솔 경로가 쓰거나 에러를 내면 페이지 렌더가 백로그를 건드린다. AC-THG-005.

## 6. 요구사항 (GEARS)

> **패턴 라벨의 출처.** 아래 괄호 라벨은 `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format의 5패턴 표(`:55-59`)를 따른다 — Ubiquitous / Event-driven / State-driven / Capability gate(Where) / **Event-detected**. `(Unwanted)`는 같은 파일 `:82`의 **legacy EARS** 표에 있는 이름이며, 호환 창(2026-11-22) 안에서 `shall not` 형 금지 요구사항의 관용 라벨로 그대로 쓴다. 두 표를 섞어 쓴다는 사실을 여기 명시해, 라벨 하나만 보고 「문서가 선언한 분류 체계와 어긋난다」고 읽히지 않게 한다.

- **REQ-THG-001** (Event-driven) — **When** the todo queue-root resolution reaches the home-fallback branch (git could not answer) **and** the launch base is classified as a temporary-directory origin, the resolver shall not resolve to, and shall not create, a home queue root under `~/.moai/todo/`; it shall **instead resolve to the launch base itself** — that is, to a root satisfying the property that `BacklogPathForRoot(<resolved root>)` names the project's existing project-local backlog file. The refusal shall carry no relocation side effect on the pure path. Where the temporary-origin condition holds **and** the home directory is simultaneously unresolvable, the temporary-origin refusal shall win and the resolver shall resolve to the launch base — the discriminant shall therefore be evaluated **before** the home-resolution outcome is consulted, so that the returned root is fixed by this requirement and not by the placement of the predicate.
- **REQ-THG-002** (Ubiquitous) — The temporary-origin discriminant shall classify by comparing the normalized absolute launch base against a normalized temp-root set — `os.TempDir()`, `/tmp`, `/var/folders` — using path-component containment, never raw string-prefix matching. (The injection seam this set is reached through is **REQ-THG-009**, split out of this requirement in 0.1.4 — see §8.)
- **REQ-THG-003** (Ubiquitous) — The discriminant shall normalize BOTH sides of every comparison with one rule: `filepath.EvalSymlinks` when the path exists, lexical `filepath.Clean` when it does not (the `internal/cli/launcher.go` `resolveSymlinks` rule, adopted for its documented GOOS determinism).
- **REQ-THG-004** (Event-detected) — **When** normalization fails, or the classification is otherwise uncertain, the discriminant shall report *not temporary* — the guard fails open, preserving the documented fail-open availability.
- **REQ-THG-005** (Unwanted) — The guard shall not withdraw the home fallback for a non-git project outside the temp-root set; "any non-git base" is explicitly NOT the trigger.
- **REQ-THG-006** (Event-driven) — **When** the guard refuses the home queue on the command path, the command shall surface guidance naming the matched temp root **and the substitute root the run continues against**, and the run shall continue rather than exit non-zero; a refusal shall not present as silent success. (U1 is now decided — shape (b), §8. The preceding version carried a two-shape disjunction "…or stating that no queue was used when the run stops"; that half is pruned because the shape it served was not chosen, not because it was wrong.)
- **REQ-THG-007** (Unwanted) — The guard shall not alter `TodoQueueProjectKey` derivation, the git branch (`primaryCheckoutRoot`), the adopt-not-shadow migration on non-temporary fallbacks, or the purity of `ResolveTodoQueueRoot` — the pure entry point shall still perform no `MkdirAll`, `Rename`, or `WriteFile` on any branch.
- **REQ-THG-008** (Unwanted) — The guard shall not delete, move, or migrate any existing directory under `~/.moai/todo/`.
- **REQ-THG-009** (Ubiquitous) — The temp-root set of REQ-THG-002, and the discriminant that reads it, shall be reachable through a package-level injection seam exported from `internal/kanban`, so that a test may declare its own base non-temporary **without leaving the temporary directory this project's [HARD] isolation discipline mandates** (`t.TempDir()`), and may assert the discriminant's verdict on that base directly. Absent the seam, no fixture constructible under that discipline can reach the home-fallback branch at all — so the guard ships unexercisable, and REQ-THG-005 loses its only producing acceptance criterion. Reachability is required from the consuming packages' tests as well as `internal/kanban`'s own (`internal/cli`, `internal/web`), which is why the seam is exported rather than package-private.

## 7. 범위 밖 (Non-goals)

### Out of Scope — 기존 고아 343개의 정리

- `~/.moai/todo/` 아래 343개 디렉터의 삭제·이관·아카이브는 카드 **t542**의 소관이다. 본 SPEC은 그것을 관련 작업으로 참조만 하며, 어떤 요구사항도 정리·삭제·이관을 정의하지 않는다(REQ-THG-008).

### Out of Scope — `stateanchor.Resolve`의 `project_dir` 미검증

- `internal/stateanchor/stateanchor.go`가 `project_dir`을 검증 없이 돌려주는 문제는 카드 **t537**의 소관이다. 다른 파일, 다른 카드다. 본 SPEC의 판독에서 두 작업의 접점은 발견되지 않았다 — `internal/kanban/todo_root.go`는 `internal/stateanchor`를 임포트하지 않는다.

### Out of Scope — 「모든 비git base」로의 확대

- 홈 폴백 자체의 철회는 구현하지 않는다. 운영자 결정(§3)이 임시-디렉터 기원으로 범위를 한정했고, 그 넓은 판독은 기각됐다.

### Out of Scope — 축 B(cwd 앵커) 수리

- `SPEC-STATE-ANCHOR-001`의 축 B(statusline/config 상태-앵커 통일)는 그 SPEC이 이미 닫았다. 본 SPEC은 `internal/statusline` / `internal/config` / `internal/hook` / `internal/session`을 변경하지 않는다.

### Out of Scope — 큐 저장 형식·경로 체계 변경

- `backlog.json`의 형식, `BacklogPathForRoot`의 경로 규칙, `TodoQueueProjectKey`의 키 유도는 불변이다(REQ-THG-007). 바뀌는 것은 「임시 기원일 때 홈 루트로 가지 않는다」 하나뿐이다.

## 8. 미검증 항목 (Gaps)

- **Linux의 `/tmp`·`/private/tmp` 실측이 없다.** §4의 심링크 관측은 이 macOS 기계의 것이다. Linux에서 `/tmp`는 통상 실디렉터이고 `/private/tmp`는 존재하지 않으므로 정규화가 어휘적 `Clean`으로 떨어져 무해하게 비활성이 되리라는 것이 **설계 근거이지 측정이 아니다**. 확인은 run-phase의 CI linux 매트릭스가 수행한다(plan.md §C).
- **U1 — 거부 시 CLI의 형태는 이제 결정됐다: (b) 안내를 내고 대체 루트로 계속한다.** (운영자 결정 2026-09-07, 리드 경유. 종전 판본은 이 항목을 미결로 남겼다.) 결정 근거 세 가지: ⑴ 오염은 **홈을 건드리지 않는 것**만으로 이미 막히므로 비영 종료(a)가 더 막아주는 것은 없다. ⑵ (b)는 REQ-THG-001의 대체 루트 결정과 정합한다 — 순수 경로가 그 루트를 받는 이상 명령 경로만 다른 결론을 낼 이유가 없다. ⑶ 웹 콘솔은 종료 코드를 갖지 않고 어느 쪽이든 루트를 받으므로, (a)를 택하면 **같은 거부에 대해 CLI와 콘솔이 다르게 행동한다** — 소비자 둘이 서로 다른 것을 보게 된다. 덤으로 임시 디렉터에서 도는 기존 테스트·스크립트가 (b)에서는 계속 동작한다.
  - 이 결정은 **순수 리졸버의 반환값과 별개의 축이었다.** 콘솔(`internal/web/todo_queue_read.go:33`)은 종료 코드를 갖지 않으므로 U1을 어느 쪽으로 정하든 순수 경로는 반드시 루트를 반환한다 — `ResolveTodoQueueRoot`의 시그니처가 오류도 옵셔널도 아닌 `string` 하나이기 때문이다(`todo_root.go:64`). 그래서 거부도 루트를 이름 불러야 하고, 그 루트를 REQ-THG-001이 고정한다. 지금은 두 축이 같은 값으로 만난다: 순수 경로가 반환하는 루트와 명령 경로가 계속 쓰는 루트가 같다.
  - **종전 판본의 과장 정정(보존)**: 이 자리에 있던 "AC-THG-005는 두 형태 중 어느 쪽에서도 판정 가능하도록 「홈 디렉터 0개 + 안내 관측」으로 쓰였다"는 주장은 그 시점 AC에 대해 참이 아니었다 — Then절은 3항이었고 GREEN 주석만 2항으로 축소해 중립성을 만들어냈다. 0.1.1이 Then절 자체를 중립으로 고쳤고, 0.1.2가 결정된 형태로 가지치기했다.
- **`/var/tmp`는 임시 루트 집합에서 의도적으로 제외한다.** 이 기계 실측: `/var/tmp`와 `/private/var/tmp`가 둘 다 실재하는 `drwxrwxrwt` 디렉터이고, `os.TempDir()`은 이를 가리키지 않는다(`TMPDIR=/var/folders/kt/.../T/`). 따라서 `/var/tmp` 아래의 비git 실행은 가드를 **통과해** 종전대로 홈 큐를 얻는다 — fail-open 방향이므로 회귀가 아니라 오늘 동작의 유지다. 제외 근거는 §4의 비대칭이다: `/var/tmp`는 재부팅을 견디는 **지속성** 임시 루트라 「나중에 사라진다」는 고아화 전제가 애초에 약하고, 그만큼 그 아래에 오래 사는 진짜 프로젝트가 놓일 여지가 `/tmp`보다 크다고 **판단한다** — 이는 설계 판단이지 측정이 아니다. 두 루트의 거짓 양성 빈도를 재보지 않았다. 측정된 것은 하나뿐이다: 확인된 생산 오염 1건의 기원이 `/tmp`이지 `/var/tmp`가 아니다(§1.1). **재검토 조건(관측 가능한 형태로 교체)**: `/var/tmp` 아래의 비git 실행이 보고되거나 재현되면 루트 집합에 추가한다(한 행짜리 변경). 종전 조건이던 "`/var/tmp` 기원의 고아가 실측되면"은 **영원히 발화하지 않을 수 있으므로 폐기한다** — 고아 디렉터명은 sha256 4바이트라 기원 경로가 복원되지 않을 수 있고, 바로 아래 `proj-325ca0b6`이 그 복원 실패 사례다.
- **`proj-325ca0b6`의 기원 경로는 복원되지 않았다.** 후보 루트 대조로 일치하는 경로를 찾지 못했고, sha256 역상 전수 탐색은 불가능하다. **임시 기원인지 아닌지 판정하지 못했다** — 비임시라고 판정한 것이 아니다. 본 SPEC의 논지(생산 축이 살아 있고 임시 기원이 실재한다)는 `t203-probe-d7a16ea2` 1건의 복원만으로 성립하므로 이 미검증은 결론을 흔들지 않는다.
- **관측 — `homeTodoQueueRoot`의 no-home 가지가 상태 디렉터를 루트로 돌려준다. 본 SPEC은 이를 고치지 않는다.** D9 수리 중에 드러난 것이고, 카드가 묻지 않은 발견이다.
  - **읽은 것**: `BacklogPathForRoot(root)`(`internal/kanban/state_dir.go:129-132`)는 `resolveStateDir(root, false)`에 `backlog.json`을 붙인다 — 즉 넘겨받은 값 아래에 `.moai/state/{todo|kanban}/`을 **한 번 더** 건다(doc `:126` "under a project root"). 그런데 `homeTodoQueueRoot`의 홈 해석 실패 가지(`todo_root.go:121-122`)는 `resolveStateDir(base, false)` — **이미 상태 디렉터인 값** — 을 `ok=false`와 함께 돌려주고, `fallbackTodoQueueRoot`(`:136-137`)와 `ResolveTodoQueueRootAdopting`(`:81-82`)이 그것을 **루트로** 반환한다. 같은 파일의 read-through 가지(`:148`)는 `base`를 돌려주고 머리말(`:22-24`)이 그것을 "PROJECT-LOCAL root"라 부른다 — 두 반환의 계층이 서로 다르다.
  - **읽은 것에서 따라 나오는 것**: 그 가지가 반환한 값에 소비자가 `BacklogPathForRoot`를 합성하면 `base/.moai/state/todo/.moai/state/todo/backlog.json`이 된다 — 운영자의 로컬 큐(`base/.moai/state/todo/backlog.json`)가 아닌 경로다.
  - **주장하지 않는 것**: 이것이 사용자에게 실제로 보이는 결함이라고 **주장하지 않는다.** 트리거는 `HomeDirFn()`이 오류를 내는 것뿐이고, 그 상황이 생산에서 얼마나 발생하는지 측정하지 않았다. 실행으로 관측하지도 않았다 — 위는 두 함수의 전문 판독에서 기계적으로 유도한 합성이다. 의도된 설계인지 잠복 결함인지도 판정하지 못했다(그 가지를 `BacklogPathForRoot`와 합성해 단언하는 테스트도, 계층을 설명하는 주석도 찾지 못했으나 **부재는 부재의 증거가 아니다**).
  - **본 SPEC의 처분**: 고치지 않는다. 두 관심사는 **다르다** — 본 SPEC이 닫는 것은 **임시 기원**이고 이 가지는 **홈 해석 불가**다 — 그러나 **배타적이지는 않다**(plan-audit 3차 D16, 지적이 옳다). 비git 임시 base에서 `HomeDirFn()`이 오류를 내면 두 조건이 **동시에** 성립한다. 종전 판본은 이 자리에 「트리거가 다르다」고 적어 배타성을 함의했고, 그 함의는 거짓이다. 교차가 실재하므로 그 교차에서 무엇을 반환하는지가 정해져 있어야 하며, **REQ-THG-001이 그것을 정한다: `base`가 이긴다.** 그 요구사항은 값만이 아니라 **술어의 평가 순서**도 구속한다 — 판별식을 `fallbackTodoQueueRoot`의 `ok` 검사 **뒤**에 두면 반환값이 `base/.moai/state/todo`(= `:121`의 그 어긋난 형태)가 되어, **본 카드가 본뜨지 않기로 한 형태를 배치 순서만으로 다시 집어 들게 된다.** 순서를 요구사항에 못박는 이유가 이것이다: 결과가 구현자의 배치 선택에 좌우되어서는 안 된다. 관측 지점은 `internal/kanban/todo_root_test.go:204-217` (`TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing` — `t.TempDir()` base + `HomeDirFn` 오류로 정확히 이 교차를 밟는다). 그 테스트는 이 결정으로 **의도적 갱신 대상**이 되며 `plan.md` §C 영향 목록에 등재돼 있다. 이 결정이 하는 일은 교차의 반환값을 고정하는 것뿐이고, `:121` 가지 **자체의 계층 어긋남은 여전히 고치지 않는다**(별도 카드 후보). 요구사항을 더 추가하면 범위가 넓어진다. **별도 카드 후보로 보인다**(발행 여부는 리드 소관). 본 SPEC이 이 관측에서 취하는 것은 하나뿐이다: REQ-THG-001이 그 형태를 **본뜨지 않는다**(0.1.1이 본떴던 것을 0.1.2가 되돌렸다).
- **판정 — `SPEC-WEB-TODO-QUEUE-001` AC-WTQ-008은 철회되지 않는다. 전제 정합이지 철회가 아니다 (plan-audit 3차 D15에서 파생, 리드가 그 SPEC을 직접 읽고 판정).** 가드가 착지하면 `TestResolveTodoQueueRootAdopting_AdoptsLocalQueue`(`internal/kanban/todo_root_test.go:223-249`)가 깨진다. 그 테스트는 AC-WTQ-008의 산출 테스트이므로, 새 동작에 맞춰 고쳐 쓰는 것이 **다른 SPEC의 AC를 조용히 철회하는 경로**가 될 수 있다. 그렇지 않은 이유:
  - **AC-WTQ-008의 주어는 「비git」이지 「임시」가 아니다.** 그 criterion은 전제를 AC-WTQ-006에서 물려받고(`acceptance.md:129-135` ← `:105-111`), AC-WTQ-006의 Given은 "a working directory that **git cannot resolve to a primary checkout**"다. 「임시」라는 말은 그 전제 어디에도 없다.
  - **그것이 봉사하는 REQ-WTQ-004도 마찬가지다.** `spec.md:125-130`은 해석(resolve)과 이관(adopt)의 **분리**를 요구하며, AC-WTQ-008 자신의 baseline 주석이 그 criterion은 이관의 "**call site**를 좁혔을 뿐 동작을 바꾸지 않았다"를 단언한다고 적는다. 임시성은 축에 없다.
  - **테스트의 `t.TempDir()`(`todo_root_test.go:224`)는 부수적이다** — 비git 디렉터를 얻는 관용 수단이지 criterion의 주어가 아니다.
  - **따라서 t536이 하는 일은 홈 폴백에 도달하는 비git base의 *모집단을 좁히는 것*이고**, AC-WTQ-008이 단언하는 보장은 **비임시 비git base에 대해 그대로 참**이다. 본 SPEC은 AC-WTQ-008을 다시 쓰지 않으며 `SPEC-WEB-TODO-QUEUE-001`의 어떤 산출물도 수정하지 않는다. `plan.md` §F M2가 그 테스트의 처분을 **보존(비임시 픽스처로 이관)**으로 지정한다.
- **판정 — `SPEC-STATE-ANCHOR-001` AC-SA-011도 철회되지 않는다. 그 요구사항이 문서화한 **두 번째 가지로의 이행**이다 (plan-audit 4차 D18에서 파생, 0.1.4).** `AC-WTQ-008`과 같은 급의 판정을 여기 남긴다 — 판정의 논리가 옳다는 것과 그 논리가 필요한 곳에 전부 적용됐다는 것은 다른 주장이기 때문이다.
  - **문제의 형태.** 가드가 착지하면 `TestGuardBypassMutant_ObserveHomePollution`(`internal/cli/todo_axisa_guard_test.go:150`)이 깨진다. 그 테스트는 canary HOME + 비git `t.TempDir()`에서 `newTodoCmd()`를 직접 실행해 **홈 오염이 발생하는 것**을 단언하고, 오염이 **없으면** `t.Fatalf`로 죽는다. 본 SPEC의 가드가 하는 일이 정확히 그 오염을 불가능하게 만드는 것이다. 그 테스트는 `SPEC-STATE-ANCHOR-001`(`status: completed`)의 **AC-SA-011 / REQ-SA-011 산출 테스트**다(그 SPEC의 `acceptance.md:23`·`:152`, `spec.md:106`).
  - **⑴ REQ-SA-011 본문이 이 경우를 이미 담고 있다.** 인용(그 SPEC `spec.md:106`, Event-detected): "**When** the Axis-A guard is bypassed (a test helper executing `newTodoCmd()` directly), the canary sweep shall observe the resulting pollution as mutant evidence that the guard — not coincidence — is what keeps the sweep at zero; **a mutant that produces no pollution shall be reported alongside the one that does, naming the guard boundary it reveals.**" 후반부가 **두 번째 가지**다. 따라서 가드가 착지해 뮤턴트가 더는 오염을 내지 않을 때, REQ-SA-011은 **위반되지 않고 그 문서화된 두 번째 가지를 탄다.** 그 테스트의 `t.Fatalf` 문구 자신이 그 요구사항을 되인용한다("REQ-SA-011: a pollution-less mutant is a finding, not a pass") — 요구사항이 기계적으로 구현하지 않은 가지를 사람에게 지시하고 있는 것이다.
  - **⑵ 그 SPEC §5는 이 결정을 명시적으로 위임했다.** §5의 표제와 서두가 홈 폴백의 임시-디렉터 기원 거부 여부를 **운영자 소관으로 유보하고 그 SPEC에서는 구현하지 않는다**고 선언한다. 본 카드는 §5가 이름 부른 「거부하는 안」을 이행하며, 카드의 `운영자 결정 2026-09-07`이 곧 그 유보된 결정이 내려진 것이다. 즉 t536은 완료된 SPEC에 침입하는 것이 아니라, **그 SPEC이 스스로 자기 밖에 둔 결정의 이행 주체**다.
  - **⑶ §5:115가 「이 결정은 본 SPEC의 어느 REQ에도 영향을 주지 않는다」고 단언한다** — 그리고 ⑴의 두 번째 가지가 **그 단언을 참으로 만드는 기제**다. 그 가지가 없었다면 §5:115는 AC-SA-011의 산출 테스트가 발화하는 것과 긴장 관계에 놓였을 것이다.
  - **⑷ 같은 §5:115의 마지막 절이 이 카드의 존재를 예고한다**: "본 결정의 구현은 **별도 후속 카드로 발행하는 것이 적절하다**." t536이 그 카드다.
  - **처분.** `SPEC-STATE-ANCHOR-001`은 **한 글자도 수정하지 않으며 `completed`로 남는다.** 그 SPEC `acceptance.md:23`의 AC-SA-011 증거 칸이 "(뮤턴트 오염은 지금도 재현 가능 — 그것이 증거)"라고 적는 것은 **t536 이전 트리에 시각 고정된(time-indexed) 기재**다 — 그 사실을 여기 적어두는 것이 처분이고, 그 파일을 고치는 것이 아니다. 테스트 자체는 가지 1에서 가지 2로 옮겨가므로, 그 갱신은 **요구사항이 지시한 것**이지 픽스처 편의가 아니다. 갱신 형태와 산출물(경계 보고)은 `plan.md` §C.1 D행과 §F M3가 정한다.
- **판정 — 판별식 주입 이음매는 본 SPEC의 요구사항이고, REQ-THG-002에 접었다 (plan-audit 3차 D15에서 파생).** 결론만 적지 않고 논거를 적는다.
  - **왜 요구사항인가.** 이 리포의 [HARD] 테스트 격리 규율은 모든 테스트 임시 디렉터를 `t.TempDir()`로 만들 것을 요구한다. 그런데 `t.TempDir()`가 돌려주는 경로는 정의상 `os.TempDir()` 아래이므로, 가드가 임시 기원을 거부하는 순간 **표준 격리 도구로 만든 어떤 base도 홈 폴백 가지를 밟지 못한다.** 「비임시 디렉터를 픽스처로 쓰면 된다」는 도피처가 그 규율 아래에서는 존재하지 않는다. 귀결이 구체적이다: AC-THG-003(비임시 비git base가 홈 폴백을 유지한다 — REQ-THG-005의 유일한 산출 AC)은 **픽스처를 가질 수 없어 판정 불가가 되고**, 공허화 5건을 비임시 사본으로 보존하라는 M2 종료 조건도 이행할 수단이 없다. 즉 이음매가 없으면 요구사항 하나가 검증 불가능해진다 — 이것은 구현 취향이 아니라 **가드 자신의 성질**이다.
  - **왜 지금 이 판단이 특히 무거운가.** 시험 불가능한 가드는 이 카드가 하루 종일 만난 실패 계열의 극단이다. 본 SPEC은 같은 자리에서 **공허한 초록을 세 번 냈다**(아래 관측 항목). 그래서 판별식의 시험 가능성을 「run-phase가 알아서」로 미루는 것은 네 번째를 예약하는 것과 같다.
  - **선례는 같은 파일 안에 있다.** `HomeDirFn`(`internal/kanban/todo_root.go:39-46`)이 패키지 수준 함수 변수이고 `stubHome`(`todo_root_test.go:55-61`)이 그것을 스텁한다. 그 doc이 존재 이유를 스스로 적는다 — "without it the fallback root is uncontrollable in tests". 임시 루트 집합이 요구하는 것이 정확히 같은 성질이다.
  - **접기가 정직한 이유 [0.1.3의 논거 — 보존, 기각됨] (당시 Tier 판정 — REQ는 Tier S 상한 8에 있었다).** REQ-THG-002가 도입하는 대상은 **임시 루트 집합**이다. 접은 절은 새 사건도, 새 행위자도, 새 가지도 들여오지 않는다 — 같은 문장이 이미 고정한 **그 집합의 속성**(어떻게 도달되는가)을 말할 뿐이다. D9의 선례가 같은 모양이었다: 대체 루트의 성질 절이 REQ-THG-001에 접힌 것은 그 절이 **그 요구사항이 이미 고정한 반환값**을 구속했기 때문이다. 여기서도 이음매 없는 REQ-THG-002는 「집합은 {`os.TempDir()`, `/tmp`, `/var/folders`}다」이고, 이음매를 넣은 REQ-THG-002는 「그리고 그 집합은 교체 가능하다」다.
  - **반대 논거도 적는다(강변하지 않기 위해).** 이음매는 *시험 가능성* 축이고 REQ-THG-002의 본체는 *분류 동작* 축이라, 다른 관심사를 한 문장에 묶었다고 읽을 여지가 있다. 그 반론을 받아들이면 요구사항은 9개가 되고 **Tier S는 더 이상 맞지 않는다(Tier M)**. 본 SPEC은 접기를 택하되 이 갈림길을 여기 남긴다 — 조용히 9개를 싣지 않기 위해서다. ~~현재 판정: REQ 8 / AC 8, Tier S 유지.~~ **리드 판정 2026-09-08: 반론 채택.** 이음매는 `REQ-THG-009`로 분리되고 Tier는 **M**이다. 갈림길 기록은 지우지 않는다 — 어느 쪽으로 갔는지가 아니라 갈림길이 있었다는 사실만 남으면, 다음 편집자가 같은 자리에서 같은 저울질을 처음부터 다시 한다.
  - **분리 근거 (리드).** 접힌 상태에서 그 절은 **주어가 「분류 동작」인 요구사항 안의 종속절**이다. REQ-THG-002를 정리하는 후속 편집자가 그 절을 지워도 **어떤 AC도 그것을 눈치채지 못한다** — 그런데 이음매가 없으면 `t.TempDir()` 격리 규율 아래에서 홈 폴백 가지를 밟을 픽스처가 존재하지 않으므로 **가드가 시험 불가 상태로 출하된다**. 이 SPEC은 같은 자리에서 공허한 초록을 세 번 냈다(D2 → D10 → D14). *관측 가능성을 보장하려고 존재하는* 요구사항은 독립적으로 지목 가능해야 한다 — 그것이 이 SPEC이 반복해 대가를 치른 실패 계열의 유일한 구조적 방어다.
  - **D9 선례와 모양이 다르다.** D9에서 REQ-THG-001에 접힌 절은 **같은 사건의 결과**였다(「거부가 무엇을 돌려주는가」). 이음매는 그 사건을 **관측하는 도구**다. 결과와 도구는 다른 축이므로 D9를 이 접기의 선례로 쓸 수 없다 — 0.1.3이 그렇게 썼고, 그것이 이 판정이 뒤집은 부분이다.
  - **커버리지 (0.1.4 정정 — D21)**: 이 절의 산출 AC는 AC-THG-003이며, 분리 후 그 AC는 **REQ-THG-005와 REQ-THG-009 둘 다**를 겨눈다(`acceptance.md` §D 매트릭스 행이 그렇게 정정됐다). 새 AC를 만들지 않는다. **현재 판정: REQ 9 / AC 8, Tier M**(Tier M 상한은 REQ 16 / AC 16이므로 여유가 있다).
- **관측 — 판독 도구가 결론을 정한 사례 3건 (0.1.4). 요구사항이 아니라 관측으로 적는다.**
  - ⑴ **대소문자 (D18).** `plan.md §C.1`이 「전수 8건」의 근거로 든 grep은 `'ResolveTodoQueueRoot'` — 대문자 `R`로 시작한다. CLI의 소문자 래퍼 `resolveTodoQueueRoot()`(`internal/cli/todo.go:71`, 본문이 `kanban.ResolveTodoQueueRootAdopting`을 부른다)를 그 패턴은 **원리상 볼 수 없다**. 실측: `/usr/bin/grep -c 'func Resolve' internal/cli/todo.go` → **0**. 그 결과 확실히 깨지는 테스트 2건이 목록 밖에 있었고, 그중 하나는 완료된 형제 SPEC의 AC 산출 테스트였다. 도구가 못 본 것이 「없는 것」으로 기록됐다.
  - ⑵ **잘린 창 (`SPEC-STATE-ANCHOR-001` §5 판독).** 위 AC-SA-011 판정을 준비하며, 한 판독은 §5를 `sed -n '100,112p'`로 읽었다 — §5는 **108-116행**이므로 그 창은 끝에서 네 줄이 모자란다. 그 창에서 115행이 보이지 않았고, **거기 있는 문장이 「없다」로 보고됐다.** 실제로 115행은 「이 결정은 본 SPEC의 어느 REQ에도 영향을 주지 않는다」와 「본 결정의 구현은 별도 후속 카드로 발행하는 것이 적절하다」를 함께 담는다 — 위 판정의 근거 ⑶과 ⑷이 정확히 그 줄이다.
  - ⑶ **원본 대신 요약 (같은 판독의 앞 단계).** 「§5의 근거가 약하고 REQ-SA-011과 긴장 관계에 있다」는 초기 성격 규정은 REQ-SA-011의 **두 번째 가지를 아직 찾기 전에** 쓰였다 — 원본 대신 원본에 대한 요약이 논거 자리에 섰다. 두 번째 가지를 읽고 나니 긴장은 없었고, 네 조각이 같은 방향을 가리켰다.
  - **공통 모양**: 세 번 모두 **판독 도구(패턴·창·요약)가 결론을 정했고**, 세 번 모두 정정은 도구를 바꾸는 것으로 끝났지 판단을 바꾸는 것으로 끝나지 않았다. ⑴은 부재 주장에 대소문자 무관 검색을 요구하고, ⑵·⑶은 **인용은 절의 자기 구분자로 경계 짓는다**(`sed -n '/^## 5\./,/^## 6\./p'`)는 교정을 낳는다 — 추측한 줄 번호로 자르지 않는다. 이 교정은 본 카드 밖으로도 회람된다(리드).
  - **「전수가 전수가 아니었다」의 파급 — 그 열거에 기댄 다른 판단도 함께 재검토했다.** 열거 하나가 틀렸다는 것은 그 열거를 전제로 한 모든 판단이 미확정이 된다는 뜻이므로, 아래를 다시 쟀다:
    - **C행 판정식** — 재검토했고 **뒤집혔다**(D19). 7건 중 6건에 공허했다.
    - **`AC-WTQ-008` 교차 판정** — 재검토했고 **유지된다**. 그 판정의 논거는 그 SPEC 본문(AC-WTQ-006의 Given, REQ-WTQ-004)에서 나왔지 열거에서 나오지 않았다.
    - **`SPEC-STATE-ANCHOR-001` §5** — **절의 자기 구분자로 경계 지어 인용으로 재검증했다**(`## 5.` 108행 ~ `## 6.` 117행 직전 = 108-116). 네 조각이 같은 방향을 가리켰고, 초기의 「긴장 관계」 성격 규정은 철회됐다.
    - **영향 패키지 수** — 3 → **4**로 늘었다(`internal/cli`가 두 갈래로 들어온다). D8의 접촉 금지 목록은 영향받지 않는다(그 목록은 *변경* 대상이고 이쪽은 *테스트* 영향이다).
    - **Tier** — 영향 14건 / 4패키지는 Tier M 판정을 **강화**한다(리드 판정과 독립적으로 같은 방향).
  - **처분**: 각 인스턴스만 고치고 이 반복을 요구사항으로 만들지 않는다 — 위 「공허한 초록 3연발」 항목과 같은 처분이며, 이유도 같다(대상이 이 가드가 아니라 검증·판독 설계 일반이다). 후속 카드 후보이며 발행은 리드 소관이다.
- **관측 — 본 SPEC은 같은 검증 설계에서 공허한 초록을 세 번 냈다. 요구사항이 아니라 관측으로 적는다.**
  - ⑴ **D2**(1차) — AC-THG-001의 And절("canary HOME 아래 디렉터 0")이 자기 Given(로컬 큐 없는 `t.TempDir()`) 아래에서 `MkdirAll`에 도달하지 못해, **가드 없는 현재 트리에서도 이미 참**이었다. ⑵ **D10**(2차) — 반환 루트를 **문자열 동일성**으로만 물어, 계층이 한 단계 어긋난 값이 GREEN으로 통과했다. ⑶ **D14**(3차) — 뮤턴트 아래 갈래 (b)에서 그 읽힘 단언이 **공허하게 통과한다**(아래 acceptance.md 정정).
  - 셋의 공통 모양: **판정문이 겨눈 대상이 그 갈래에 없거나, 이미 만족돼 있다.** 세 번 모두 다른 층에서 났고(단언 대상 부재 → 단언 정밀도 부족 → 뮤턴트 아래 무력화), 세 번 모두 **읽기가 아니라 실행·판독으로만** 드러났다.
  - **본 SPEC의 처분**: 고치는 것은 각 인스턴스뿐이고, 이 반복 자체를 요구사항으로 만들지 않는다 — 그것은 이 SPEC의 범위가 아니라 검증 설계 일반의 문제다. **후속 카드 후보로 보인다**(발행 여부는 리드 소관 — 본 SPEC은 카드를 내지 않는다).
- **`resolveSymlinks` 규칙을 재사용할지 복제할지는 run-phase 결정이다.** 그 함수는 `internal/cli`의 비노출 심볼이라 `internal/kanban`에서 직접 부를 수 없다. **규칙**은 REQ-THG-003으로 고정돼 있고, 공유 위치로의 추출 여부는 구현 판단이다(Tier M 범위 안에서 최소로 — 0.1.4에서 S→M).
- **임시 루트 아래에 사는 진짜 비git 프로젝트의 실제 빈도는 측정하지 않았다.** §4가 그 잔여 위험을 수용한다고 명시한다.
