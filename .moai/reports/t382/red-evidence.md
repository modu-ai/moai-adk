# t382 RED 증거 원장 — tree `f72c0bf0f`

`verification-completeness.md` §2.1이 요구하는 RED-now 4요소(명령 / 축자 stdout / exit code / 트리 SHA)를 갖춘 항목만 여기 싣는다. `spec.md` §3의 AC 셀은 이 원장을 id로 인용한다.

측정 도구: `./bin/moai` (이 트리에서 `make build`한 산출물, `bin/moai` mtime 2026-08-31 19:15). PATH 바이너리 미사용.

기준선 이동 주의: `measurements-9328a5242.md`(M1~M13)는 이 SPEC 디렉터리가 **생기기 전** 트리 `9328a5242`에서 잰 값이다. 아래 R1·R2는 SPEC 산출물이 커밋된 **현재** 트리 `f72c0bf0f`에서 다시 잰 값이며, 그래서 V3R5가 23이 아니라 24다.

---

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

## 인용 불가로 분류한 것

`measurements-9328a5242.md`의 M1~M13은 명령과 관측값을 갖지만 **축자 stdout과 exit code를 갖지 않는다.** 따라서 §2.1 기준으로 release-blocking RED 셀의 근거로는 쓰지 않고, 배경 측정으로만 인용한다. AC가 판정에 기대는 RED는 위 R1~R3뿐이다.
