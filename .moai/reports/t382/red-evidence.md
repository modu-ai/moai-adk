# t382 RED 증거 원장

`verification-completeness.md` §2.1이 요구하는 RED-now 4요소(명령 / 축자 stdout / exit code / 트리 SHA)를 갖춘 항목만 여기 싣는다. `spec.md` §3의 AC 셀은 이 원장을 id로 인용한다.

**항목별 귀속 트리** — 원장 전체에 하나를 걸지 않는다. R1·R2·R3은 `1f10f5e8d`, R4는 `f967089ba`(감사 HEAD)에서 쟀다.

> **R1 귀속 정정 (감사 D6).** 초판은 R1을 `f72c0bf0f`로 귀속했다. 그러나 R1의 명령이 호출하는 `red_probe.py`는 **다음 커밋 `1f10f5e8d`에서 처음 추가됐다** — `git diff --name-only f72c0bf0f..1f10f5e8d`가 `red_probe.py`와 이 파일을 둘 다 신규로 낸다. 측정 당시 프로브가 미추적으로 워킹트리에 있었으므로 관측 자체는 정직하나, `f72c0bf0f`를 체크아웃하면 그 명령은 **실행되지 않는다** — 즉 인용한 SHA가 「이 명령을 재현할 수 있는 트리」를 가리키지 않았다. 프로브가 처음 존재하는 트리 `1f10f5e8d`로 옮긴다. 값은 무사하다: HEAD에서 재실행해도 stdout과 rc가 바이트 동일하다(감사가 독립 재현).

**측정 도구 — 두 좌표** (`verification-claim-integrity.md` §2.2: 읽은 트리와 **판정한 빌드**는 별개 좌표다):

- 도구: `./bin/moai` (트리 로컬 산출물). PATH 바이너리 미사용.
- 빌드 커밋: `./bin/moai version` → `v3.1.2-954-g9328a5242`, built 2026-08-31T10:15:33Z. **즉 바이너리는 `9328a5242`에서 빌드됐고 측정 트리와 다르다.**
- 지연 없음의 근거: `git diff --name-only 9328a5242..HEAD -- '*.go'` → **빈 출력**(0건). 판정 로직이 측정 트리와 동일하므로 이 차이는 무해하다. 그 사실을 여기 적는 이유는, 적지 않으면 읽는 사람이 무해함을 **검증할 수 없기** 때문이다.

기준선 이동 주의: `measurements-9328a5242.md`(M1~M13)는 이 SPEC 디렉터리가 **생기기 전** 트리 `9328a5242`에서 잰 값이다. 아래 R1·R2는 SPEC 산출물이 커밋된 **현재** 트리 `f72c0bf0f`에서 다시 잰 값이며, 그래서 V3R5가 23이 아니라 24다.

---

## 기준선 핀 갱신 이력 — 이 원장의 셀들이 어느 트리에 걸려 있는가

이 카드는 기준선을 트리별로 핀 박는 방식을 쓴다. 그러면 핀은 **낡는다** — 아래는 그 낡음을 감추지 않고 적은 것이다.

| 시점 | 트리 | 무엇이 이 트리에서 측정됐나 |
|---|---|---|
| plan 착수 | `9328a5242` | `measurements-9328a5242.md` M1~M13. **이 SPEC 디렉터리가 생기기 전**이라 V3R5 23 / grandfathered 285 |
| plan 진행 | `f72c0bf0f` · `1f10f5e8d` | 원장 R1~R4. SPEC 산출물이 커밋된 뒤라 V3R5 **24** / grandfathered **286** |
| run 착수 (지금) | 로컬 `dd0cbc5d9` · **원격 `origin/develop` = `297a21ea7`** | 아래 주의 참조 |

**[HARD] run 착수 시점의 원격 기준선은 `297a21ea7`이다** (리드 통지 후 이 워크트리에서 재확인: `git rev-parse --short origin/develop` → `297a21ea7`, 발산 `5 5`). plan 단계가 base 로 적은 `9328a5242` 는 **이미 낡았다** — 그 사이 t377 이 착지했다.

이 원장의 R1~R4 는 흡수 **이전** 트리에서 측정된 값이므로, 통합 창에서 `origin/develop` 을 흡수한 뒤에는 **병합 트리에서 다시 재야 한다**. 흡수분이 `internal/spec` 을 건드렸다면 R1~R4 의 수치가 움직일 수 있고, 움직이지 않았더라도 그 사실 자체를 관측해야 한다 — 안 움직였다는 것은 추론이 아니라 측정이다.


## R1 — 코퍼스 RED (AC-EH3-001 · AC-EH3-002 · AC-EH3-005가 인용)

「V3R5로 분류된 SPEC 중 H-5 술어가 modern이라 부를 것이 하나도 없어야 한다」는 명제를 exit code로 판정한다. 오늘 이 명제는 거짓이다.

### 명령 (단일 호출)

```
./bin/moai spec audit --json | python3 .moai/reports/t382/red_probe.py
```

프로브 본체는 `.moai/reports/t382/red_probe.py`. 판정 술어는 `era.go`의 H-5와 같다 — `phase`에 `v3r6` 포함 또는 `v3.0` 접두, 또는 `created >= 2026-04-01`.

### 축자 stdout

```
V3R5-classified SPECs: 24
  carrying a modern-era signal (misclassified): 23
  carrying no modern-era signal (correctly V3R5): 1
  no-signal set: ['SPEC-V3R5-INIT-WIZARD-EXPANSION-001']
```

### exit code

```
1
```

### 무엇이 이것을 초록으로 만드는가 (green path)

M1이 H-3에 유예절을 넣으면 23건이 H-5로 흘러 V3R6이 되고, 프로브는 `misclassified: 0` · exit **0**을 낸다. 남는 1건은 INIT-WIZARD이며 §1.5의 이유로 남는 것이 옳다.

### 뮤턴트 프로브 (이 셀이 공허하지 않다는 근거)

유예절 `&& !hasModernEraSignal(signals)`을 지우면 — 즉 수정 이전 상태로 되돌리면 — 프로브는 다시 exit 1을 낸다. 이 왕복(1 → 0 → 1)을 run-phase가 관측하고 출력을 남긴다. **통과만 관측하면 셀렉터가 0건을 매치한 경우와 구별되지 않는다.**

### 빈 스윕 방어

첫 줄 `V3R5-classified SPECs: N`이 스윕 모집단이다. N이 0이면 프로브는 조사할 대상이 없었다는 뜻이므로 exit 0을 **통과로 읽지 않는다** — 판정 전에 N ≥ 1을 확인한다.

---

## R2 — 감사 총계 기준선 (AC-EH3-005 명제 1이 인용)

### 명령

```
./bin/moai spec audit
```

### 축자 stdout (요약 4행)

```
Total SPECs:        715
Grandfathered:      286 (pre-V3R6 — protected)
Modern-era clean:   422
Drift findings:     500
```

### exit code

```
0
```

### 파생값

`EraAutoDetected` finding의 era 집계: V3R6 205 / V2.x 144 / V3R2-R4 118 / V3R5 24. `EraUnclassified` finding **0건**. V3R6 총계는 715 − 286 = **429**(감사 요약이 직접 내지 않아 뺄셈으로 도출).

수정 후 기대값: grandfathered 286 → **263**, V3R5 24 → **1**, V3R6 429 → **452**, `EraUnclassified` 0 → **0**, V2.x 144 · V3R2-R4 118 **불변**.

---

## R3 — 이 SPEC 자신이 결함의 표본이다

### 명령

```
./bin/moai spec audit --filter-spec SPEC-ERA-H3-NARROWING-001
```

### 축자 stdout

```
Audit summary
=============
Total SPECs:        1
Grandfathered:      1 (pre-V3R6 — protected)
Modern-era clean:   0
Drift findings:     1

Findings:
  [INFO] SPEC-ERA-H3-NARROWING-001 (V3R5) — EraAutoDetected
```

### exit code

```
0
```

`created: 2026-08-31`인 오늘 작성한 SPEC이 「V3R5 시대(2026-03~04)」로 분류되고 grandfather 보호를 받는다. `phase: "v3.2.0 target"`는 `matchesModernPhase`에 걸리지 않으므로(접두가 `v3.0`이 아니고 `v3r6`도 없음) 이 건의 재분류 근거는 `created` 단독이다.

---

## R4 — drift 면제 비용 (AC-EH3-007이 인용) — 트리 `f967089ba`

**무게중심 축의 RED다**(`spec.md` §1.2 축 1). 초판은 이 축만 4요소 셀 없이 배경 등급 파일(`drift-before-9328a5242.txt` — 명령도 exit code도 없는 23행 raw 출력, 트리 SHA는 파일명에만)에 기대고 있었다. 감사 D1이 지적한 그 구멍을 이 셀이 닫는다.

판정 명제: **「`era-exempt` 처리된 행 중 modern-era 신호를 가진 것이 하나도 없어야 한다.」** 신호 없이 면제된 행(오늘 INIT-WIZARD)은 **정당한 면제이므로 세지 않는다** — 「면제 0건」을 목표로 잡으면 옳게 면제된 건까지 결함으로 세게 된다.

### 명령 (2단계, 각각 단일 호출)

```
./bin/moai spec drift > .moai/reports/t382/drift-at-f967089ba.txt
./bin/moai spec audit --json | python3 .moai/reports/t382/drift_probe.py .moai/reports/t382/drift-at-f967089ba.txt
```

프로브 본체는 `.moai/reports/t382/drift_probe.py`. 1단계 rc 0.

### 축자 stdout (2단계)

```
V3R5-classified SPECs swept: 24
  era-exempt rows:      23
  terminal-exempt rows: 1  ['SPEC-HOOK-PREEDIT-INVESTIGATE-001']
  other/missing rows:   0  []
  of the era-exempt rows, carrying a modern signal (unearned exemption): 22
  of the era-exempt rows, no modern signal (earned exemption): 1  ['SPEC-V3R5-INIT-WIZARD-EXPANSION-001']
```

### exit code

```
1
```

### green path

M1이 착지하면 **22행이 `era-exempt`에서 git 대조 결과로 바뀌고**, 프로브는 `unearned exemption: 0` · exit **0**을 낸다. 남는 두 행은 그대로다 — INIT-WIZARD(신호 없음, 정당한 era 면제) 1행과 `SPEC-HOOK-PREEDIT-INVESTIGATE-001`(superseded, `terminal-exempt`이므로 시대가 V3R6으로 바뀌어도 행이 움직이지 않는다) 1행.

### 뮤턴트 프로브

유예절을 지우면 다시 `unearned exemption: 22` · exit 1로 돌아온다.

### 빈 스윕 방어

첫 줄 `V3R5-classified SPECs swept: N`이 모집단이다. N = 0이면 exit 0을 통과로 읽지 않는다. `other/missing rows`가 0이 아니면 drift 출력과 audit 집합이 어긋난 것이므로 그 자체를 조사한다.

### 폐기된 기준선과의 대조 (한 칸씩 어긋나 있었다)

| | `drift-before-9328a5242.txt` (배경, 폐기) | **R4 (현재 기준선)** |
|---|---|---|
| 스윕 모집단 | 23 | **24** |
| `era-exempt` | 22 | **23** |
| `terminal-exempt` | 1 | 1 |
| 수정 시 바뀌는 행 | 21 | **22** |
| 잔류 era 면제 | 1 (INIT-WIZARD) | 1 (INIT-WIZARD) |

---

## 인용 불가로 분류한 것

`measurements-9328a5242.md`의 M1~M13은 명령과 관측값을 갖지만 **축자 stdout과 exit code를 갖지 않는다.** 따라서 §2.1 기준으로 release-blocking RED 셀의 근거로는 쓰지 않고, 배경 측정으로만 인용한다. AC가 판정에 기대는 RED는 위 **R1~R4**뿐이다.

`drift-before-9328a5242.txt`도 같은 등급이다 — 23행 raw 출력에 명령도 exit code도 없고 트리 SHA는 파일명에만 있다. **R4가 대체했으므로 어떤 AC의 판정 근거도 아니다.** 파일은 이력으로 남기되 인용하지 않는다.
