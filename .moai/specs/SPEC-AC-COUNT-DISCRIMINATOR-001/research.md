---
id: SPEC-AC-COUNT-DISCRIMINATOR-001
title: "AC 개수 자가검사 판별자 — 실측과 선행 조사"
version: "0.5.0"
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/agents/moai, .claude/rules/moai/development, internal/template/templates/.claude, internal/template/templates/.codex, internal/spec"
lifecycle: spec-anchored
tags: "b12, changelog, acceptance-criteria, self-test, measurement"
---

# 실측과 선행 조사 — SPEC-AC-COUNT-DISCRIMINATOR-001

> **읽는 법**: 이 문서는 결정을 담지 않는다. 결정은 `spec.md` §2·§3 과 `plan.md` §E M0 에 있고, 여기 있는 것은 그 결정이 선 **측정**과 각 측정의 **방향**(상한인가 하한인가 검증값인가)이다. 방향을 잃은 수치는 이 SPEC 이 반복해 겪은 결함 형태이므로, 모든 표에 방향 열을 둔다.
>
> 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t338`, 브랜치 `WT-ac-count-sweep`, base `da03d9188`(= origin/develop).

## §A 수치의 방향 — 이 SPEC 의 중심 규율

세 스캔은 모두 **불건전하다.** 서로 다른 방향으로 불건전하고, 그래서 **더할 수 없다.**

| 축 | 스캔 | 방향 | 왜 그 방향인가 |
|---|---|---|---|
| 폐기 | `overcount-detector.sh` | **하한** | 행 단위 표지 매칭이라 폐기 기록이 여러 등장에 걸치고 한 곳에 표지가 없으면 놓친다. 실측: `SPEC-UPDATE-DOC-DRIFT-001` 은 폐기 식별자 셋 중 **하나만** 잡혔다 |
| 인용 | `multidomain-scan.sh` / `pre-terminal-scan.sh` | **상한** | 접두사가 둘 이상이면 잡으므로, 정당하게 두 계열을 담는 SPEC(`AC-APO-*` + `AC-DCP-*`)이 같은 모양으로 읽힌다 |
| 별칭 | `alias-shape-scan.sh` | **상한** | 형태만 본다(숫자 꼬리가 같고 한쪽 알파 접두사가 다른 쪽의 진접두사). 참 별칭과 우연한 모양이 섞인다 |

**따라서 합계를 만들지 않는다.** 하한과 상한을 더한 수는 어느 쪽도 뜻하지 않는다. 이 규율은 REQ-ACD-008 로 요구 층에 올라가 있고, 0.3.0 판에서 합계 `134` 를 삭제한 근거다(iter-1 감사 D4).

## §B 모집단 — 같은 트리에서 세 값이 동시에 참이다

| 글롭 | 값 | 무엇 |
|---|---|---|
| `.moai/specs/*/acceptance.md` (깊이 1) | 602 | 이 SPEC 이 채택한 corpus. **이 SPEC 자신의 파일을 포함한다** |
| 재귀 | 603 | `_archive/SPEC-SKILL-001/acceptance.md` 가 더해진다 |
| 최초 실측 시점 | 601 | 이 SPEC 의 `acceptance.md` 가 생기기 전 |

**이것이 수를 얼리지 않고 글롭을 얼린 이유다**(REQ-ACD-006). 스냅샷 판정이 "스냅샷에 기록된 파일은 어떤 차이든 실패 · 기록된 파일이 글롭 밖으로 나가면 실패" 이므로, 저자와 나중 독자가 모집단을 다르게 읽으면 **테스트가 무엇을 단언하는지가 조용히 달라진다** — 좁게 읽으면 이미 잰 파일이 사라짐으로 잡혀 실패하고, 넓게 읽으면 잡힌 파일이 부재 보고로 조용히 흡수된다. (0.5.0 이 부재 파일을 실패에서 뺐어도 이 논거는 그대로다 — 좁힌 것은 처음 보는 파일이고, 글롭 해석이 흔드는 것은 **이미 잰** 파일이다.)

## §C 세 축의 실측

### C.1 폐기 축 (카드가 겨눈 축)

```console
$ bash .moai/reports/t338/overcount-detector.sh   # → overcount-scan.txt
acceptance.md 전수 : 601
플래그된 파일      : 18
플래그된 식별자    : 29
```

**손 판정으로 오탐 8건을 걷어냈다**(`spec.md` §8). 세 파일은 통째로 오탐이고(`SPEC-V3R3-RETIRED-AGENT-001` 5건 · `SPEC-COMPLETION-MARKER-RETIRE-001` 1건 · `SPEC-LSPMCP-RETIRE-001` 1건), `SPEC-V3R2-ORC-001` 은 유일한 플래그가 오탐이라 이 축에서 빠진다 — **오탐 4파일**.

```console
$ grep -nE "AC-RA-02([^0-9-]|$)" .moai/specs/SPEC-V3R3-RETIRED-AGENT-001/acceptance.md
46:### AC-RA-02: manager-tdd.md retired stub has all 5 standardized fields (REQ-RA-002)
```

`AC-RA-02` 는 **유효한 기준**이다. "retired" 는 그 SPEC 의 *주제*이지 그 기준의 *상태*가 아니다. **이 관측이 설계를 결정했다** — 자연어 어휘를 읽는 판별자는 주제와 상태를 구별할 수 없으므로, 판별자는 산문이 만들어낼 수 없는 **예약된 문면 토큰**이어야 한다.

플래그 집합 한정 진짜 과다 계상: **14 파일 / 21 식별자**(하한).

### C.2 인용 축 (이 SPEC 자신에서 처음 관측)

이 SPEC 의 `acceptance.md` 를 기존 스윕에 넣으면 실제 기준 수보다 훨씬 큰 값이 나온다 — 차이는 전부 **다른 SPEC 의 식별자를 인용**한 것이다. 폐기와 무관하다.

```console
$ bash .moai/reports/t338/pre-terminal-scan.sh | tail -4
multi-domain files      : 156
  status=completed      : 122
  pre-terminal          :  34
  no spec.md status     :   0
```

**이 갈라보기가 M0 의 결정을 냈다.** 156 중 122가 `completed` 이고, B12 는 그 SPEC 자신의 sync 에서만 도므로 **다시 돌 sync 가 없다** — 그 122건을 지금 정규화해도 미래의 자가검사 결과는 한 건도 달라지지 않는다. 남는 34건만이 다시 센다.

### C.3 별칭 축 (iter-1 감사가 세우고 이 판이 넓혔다)

```console
$ grep -c '^## AC-ORC-001-' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md
17
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-V3R2-ORC-001/acceptance.md | sort -u | wc -l
      34
```

참값 17, 스윕 34 — **100% 과다 계상**이고 종료코드는 0. 어느 철자도 폐기되지 않았고 남의 SPEC 것도 아니다. **세 번째 축이다.**

전수 형태 스캔의 상한은 **85 파일**(`grep -c '^=== ' alias-shape-scan.txt`). 양쪽을 각각 손으로 확인했다:

| 관측 | 파일 | 판정 |
|---|---|---|
| `AC-001` ~ `AC-HUM-001` | `SPEC-HUMANIZE-001` (`:300` 이 짧은 철자로 같은 기준을 부른다) | **참 별칭** |
| `AC-SL-001` ~ `AC-SL-NF-001` | `SPEC-STATUSLINE-001` | **오탐** — `NF` 는 별개 요구 계열 |

**참과 오탐이 한 스캔 안에 섞여 있다는 것 자체가 결론이다**: 형태 판정은 정지 규칙으로 쓸 수 없고, 85 를 결함 목록으로 쓰면 §2.1 이 어휘 매칭에서 이미 겪은 오류를 목록 층에서 되풀이한다.

## §D iter-2 감사가 낸 실측 (이 판이 더한 것)

### D.1 인접의 두 경계 — 둘 다 실측에서 갈렸다

계수기를 §3.1/§3.2 문면 그대로 두 벌 구현해 결정적 입력(`SPEC-AGENT-PARALLEL-OPT-001/acceptance.md:248`)에 돌렸다. 원본은 `completed` 이므로 **트리 밖 사본**에만 적용했다.

```console
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' orig.md | sort -u | wc -l
      54
$ python3 counter.py inside.md  sameline     # 토큰을 코드 스팬 **안**에 둔 사본
COUNT 53 (live=53 excluded=1 ambiguous=0)
$ python3 counter.py outside.md sameline     # 같은 등장, 토큰만 스팬 **밖** (백틱 개입)
COUNT 54 (live=54 excluded=0 ambiguous=0)
$ python3 counter.py eol.md     sameline     # 토큰을 행 끝으로
COUNT 54 (live=54 excluded=0 ambiguous=0)
$ python3 counter.py eol.md     line         # 같은 입력, 행 단위 구현
LINE live=52
```

**두 결론이 여기서 나온다.**

1. **닫는 백틱은 인접을 깬다**(53 vs 54). 그리고 이 SPEC 의 결정적 예시가 정확히 그 자리에 있다 — 실제 파일에서 그 등장은 코드 스팬 안이고 바로 뒤에 백틱과 괄호가 따른다. 문면이 침묵하면 AC 의 판정이 배치에 따라 뒤집힌다(감사 N4).
2. **행 끝 뮤턴트의 기대값은 54 이지 52 가 아니다.** 52 는 이 SPEC 이 금지하는 **행 단위 구현만이** 내는 값이다. 0.3.0 판이 52 를 적어 두어, 올바른 구현은 뮤턴트 절과 항목 2(53)를 동시에 만족시킬 수 없었다 — 어떤 구현도 통과할 수 없는 요구였다(감사 N2).

줄바꿈 갈래도 별도 fixture 로 갈렸다:

```console
$ python3 counter.py crossline.md sameline   # 줄바꿈은 인접을 깬다
COUNT 3 (live=3 excluded=1 ambiguous=0)
$ python3 counter.py crossline.md ws         # 줄바꿈도 공백으로 허용
COUNT 2 (live=2 excluded=2 ambiguous=0)
```

두 해석 모두 "사이에는 공백만 둔다" 를 지키면서 **다른 수**를 낸다. 문면이 고르지 않으면 구현이 고르게 된다(감사 N5).

### D.2 규약을 예시하는 문서가 규약에 걸린다

전수 corpus 를 돌리면 602 파일 중 **정확히 한 건**이 정지했고, 그것이 이 SPEC 자신의 `acceptance.md` 였다.

```console
$ python3 counter.py <이 SPEC 의 acceptance.md 사본> sameline
AMBIGUOUS AC-SYN-010 AC-SYN-012 (live=22 excluded=0)
```

원인은 AC-ACD-004 의 규약 설명 표다 — fixture 열에서 `AC-SYN-010` 을 토큰이 붙은 표시된 모양으로 보이고, 기대 열에서 같은 식별자를 표시 없이 불렀다. 그것이 정확히 `ambiguous` 의 정의다.

**이 문서 자신도 같은 규율 아래 있다.** 위 진단 출력이 정지시킨 식별자를 이름으로 담으므로, 이 파일은 그 이름들을 **표시된 모양으로 다시 쓰지 않는다**(§D.1 의 스팬 안/밖 비교가 실제 식별자 대신 서술로 주석을 단 이유도 같다 — `spec.md` §3.4 규칙 2). 실측으로 확인한다:

```console
$ python3 .moai/reports/t338/iter2-scratch/counter.py research.md adj ; echo rc=$?
COUNT …   (… ambiguous=0)      rc=0
```

정수 자체는 이 파일이 편집될 때마다 움직이므로 얼리지 않는다(REQ-ACD-008 의 재유도 규율). 여기서 읽는 것은 `ambiguous=0` 과 `rc=0` 이다.

**싼 우회로(코드 스팬 제외)는 REQ-ACD-004 가 막는다** — 마크업 앵커이기 때문이다. 그래서 예외가 아니라 **적용**으로 풀었다(`spec.md` §3.4). 수리 후 재측정:

```console
$ python3 counter.py <수리 후 사본> sameline
COUNT 24 (live=24 excluded=3 ambiguous=0)
```

정지는 사라졌고, 남는 과다 계상은 선택 A 가 명시적으로 받아들인 인용 축 잔여다.

### D.3 배포 반송자는 둘이 아니라 셋

```console
$ git grep -l -F "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+'" -- .
.claude/agents/moai/manager-docs.md
internal/template/templates/.claude/agents/moai/manager-docs.md
internal/template/templates/.codex/agents/moai/manager-docs.toml
CHANGELOG.md                                    # 역사 기록이지 정의 지점이 아니다
(그 외: 과거 SPEC 들의 progress.md / reports — 모두 기록)
```

세 번째가 `.codex` TOML 이고, **계수기 명령을 통째로 담는다**(`manager-docs.toml:68` 이하). 손으로 고치는 대상이 아니라 미러에서 기계 생성되는 대상이다:

```console
$ grep -n "templatesDir\|agentMDRoot" internal/template/agentemit/golden_test.go
30:// templatesDir is the template tree root relative to this package's dir.
31:const templatesDir = "../templates"
34:const agentMDRoot = ".claude/agents/moai"
```

**생성 원본은 템플릿 미러이지 로컬 `.claude/` 가 아니다.** 그리고 게이트가 빌드 앞에 있다:

```console
$ sed -n '23p' Makefile
build: agents-emit-check templ-generate ## Build the binary
```

`agents-emit-check` 는 의도적으로 읽기 전용이라 스스로 재생성하지 않으므로(`Makefile:31-38`), 재생성을 빠뜨린 `make build` 는 sha256 불일치로 **컴파일 전에 정지**한다. 감사가 미러 한 글자를 바꿔 재현했고, 원복 후 초록으로 돌아오는 것까지 확인했다.

`.codex` TOML 은 `catalog.yaml` 과 **별개 경로**다:

```console
$ grep -c '\.codex' internal/template/catalog.yaml
0
```

따라서 catalog 해시 갱신이 이 파일을 대신하지 못하고, §F 에서 별도 파일로 센다 — 이것이 영향 파일을 15 에서 16 으로 올려 **Tier 를 M 에서 L 로 넘긴** 한 건이다.

### D.4 자기가 인용하는 수를 자기가 움직였다

```console
$ bash .moai/reports/t338/collision-scan.sh
lines carrying >=2 distinct AC prefixes: 125     # 0.3.0 판이 문면에 얼린 값: 123
files containing such a line          : 56
```

**같은 HEAD 에서 123 → 125.** corpus 에 이 SPEC 자신의 `acceptance.md` 가 들어 있고 0.3.0 개정이 그 파일의 충돌 행을 늘렸기 때문이다. REQ-ACD-008 이 다른 모든 수치에 요구하는 재유도 규율을 이 SPEC 이 **자기 증거에만** 적용하지 않고 있었다(감사 N7).

**방향**: 이 수는 깨끗한 상한도 하한도 아니다 — 정당한 두 계열 인용 행을 과다 계상하고(상한 방향), 같은 접두사 두 식별자가 한 행에 오는 충돌은 접두사 구분 스캔에 보이지 않아 과소 계상한다(하한 방향). **D1 의 근거는 이 수에 걸려 있지 않다** — 근거는 결정적 사례 1건의 손 검증이고, 이 수는 그 사례가 외딴 것이 아님을 보이는 규모 지표일 뿐이다.

## §E 미러 쌍 baseline (M3-d · M3-e 가 단언할 값)

```console
$ diff internal/template/templates/.claude/agents/moai/manager-docs.md .claude/agents/moai/manager-docs.md
(출력 없음 — 바이트 동일)
$ shasum -a 256 internal/template/templates/.claude/agents/moai/manager-docs.md
27d6252a33131be637294ddced274213bb817012747731d08844becd7a3b7954
```

`manager-develop-prompt-template.md` 쌍은 **171행 SPEC-ID 중립화 1건**만 다르다. verbatim `cp` 는 그 중립화를 되돌리므로 금지다(REQ-ACD-007).

## §F 조사하지 않은 것 (Gaps — 정직하게 적는다)

- **인용 축 122건을 파일별로 검증하지 않았다.** 상한 라벨을 붙인 채 후보 목록으로만 남긴다. 전수 판정 없이 "부채" 로 단언하면 검증되지 않은 결함 주장이 된다
- **별칭 축 85파일의 참값을 재지 않았다.** 참/오탐 각 1건만 손 확인했고, 그 사실 자체가 "형태를 정지 규칙으로 쓸 수 없다" 의 근거다
- **`spec.md` 의 REQ 개수 스윕에 같은 형태의 과다 계상이 있는지 보지 않았다.** 후속 카드 후보(`spec.md` §6)
- **`make build` 를 끝까지 돌리지 않았다.** `.codex` 결합은 `Makefile:23` 의 의존 + `agents-emit-check` 실패 관측으로 세웠고, 전체 빌드 관측은 run-phase 몫이다
- **계수기는 이 조사용 재구현이다.** §3.1/§3.2 문면에서 옮겨 썼을 뿐 이 SPEC 이 만들 구현이 아니다. 다른 구현도 부분 표시에서는 똑같이 정지하지만(그것이 정의다), 인접 해석이 다르면 식별자 집합이 달라질 수 있다 — 그 갈래를 닫은 것이 §D.1 이다

## §G 재현 자산

| 경로 | 무엇 |
|---|---|
| `.moai/reports/t338/measurement.md` | 최초 실측(601 / 18 / 29) |
| `.moai/reports/t338/overcount-detector.sh` · `overcount-scan.txt` | 폐기 축(하한) |
| `.moai/reports/t338/multidomain-scan.sh` · `pre-terminal-scan.sh` · `pre-terminal-scan.txt` | 인용 축(상한) + 상태별 갈라보기 |
| `.moai/reports/t338/alias-shape-scan.sh` · `alias-shape-scan.txt` | 별칭 축(상한) |
| `.moai/reports/t338/collision-scan.sh` | 한 행 다중 접두사(방향 없음 — §D.4) |
| `.moai/reports/t338/repair-scratch/counter.py` | iter-2 수리 검증용 계수기 재구현(인접/행 두 벌) |
| `.moai/reports/t338/repair-scratch/{orig,inside,outside,eol,crossline}.md` | §D.1 의 입력들 |
| `.moai/reports/t338/plan-audit-iter1.md` · `plan-audit-iter2.md` | 감사 기록 |

> `repair-scratch/` 는 **감사 재현 자산**이다. 돌아간 그대로 두며 린트를 위해 고치지 않는다 — 고치면 기록이 달라진다.
