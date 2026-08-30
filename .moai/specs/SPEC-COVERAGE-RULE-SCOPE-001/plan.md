# SPEC-COVERAGE-RULE-SCOPE-001 — 구현 계획

> 되돌리기 어려운 결정을 앞에, 기계적인 작업을 뒤에 둔다. §C(결함 ① 방향)와 §D(결함 ② 심각도)가 이 계획에서 가장 바뀌기 쉬운 결정이며, 검토는 그 두 절에 집중하는 것이 이익이다.

## §A 맥락

측정 트리: 워크트리 `.claude/worktrees/t362`, 브랜치 `WT-coverage-rule-scope`, HEAD `ee50984ab`. 바이너리는 그 트리에서 빌드했다(`go build -o /tmp/moai-t362 ./cmd/moai`, rc=0). 설치본 `~/go/bin/moai`(v3.1.2, `343399d2f`)는 다른 트리라 근거로 쓰지 않았다.

두 결함의 무게는 같지 않다.

| | 결함 ② REQ 파서 협소성 | 결함 ① `acceptance.md` 미판독 |
|---|---|---|
| 위치 | `parseREQs` / `reqLinePattern` (lint.go:452) | `discoverSPECs` / `parseSPECDoc` / `CoverageRule.Check` |
| 코퍼스 영향 | 702개 중 687개(97.9%)에서 세 규칙이 침묵 | 위반 0건 (잠복) |
| 상태 | **활성 — 이 SPEC의 본체** | 잠복 — 관행이 우회하고 있어 아직 아프지 않다 |
| 파급 | 넓히면 수백 건이 새로 발화 | 좁은 수리, 새 finding 거의 없음 |

한 SPEC으로 묶는 근거는 리드 판정이다: 둘이 같은 규칙의 같은 파싱 경로에 있고, 수리가 같은 파서를 건드린다.

## §B 알려진 사실과 측정치

| 항목 | 값 | 재는 명령 |
|---|---|---|
| 전체 `spec.md` | 702 | 코퍼스 스캔 |
| 현행 패턴 매치 파일 | 15 (2.1%) | 위와 동일 |
| 매치 실패 중 REQ 토큰 보유 | 618 / 687 (90%) | 위와 동일 |
| 전 코퍼스 finding 총수 | 177 | `/tmp/moai-t362 spec lint --json` |
| 그중 `CoverageIncomplete` | **0** | `jq -r '.[].code' … \| sort \| uniq -c` |
| 그중 `InvalidREQID` / `DuplicateREQID` | 0 / 0 | 위와 동일 |

finding 내역: MovingRefUnpinned 113 / MissingExclusions 24 / StatusGitConsistency 18 / FrontmatterInvalid 14 / LegacyEARSKeyword 7 / OwnershipTransitionInvalid 1. develop 병합 전(`15453140a`)과 후(`ee50984ab`)에 동일.

**작성 중 관측된 추가 사실 세 건 (모두 이 트리에서 실측).**

1. **이 SPEC 자체가 결함 ①을 재현했다.** 관행대로 AC를 `acceptance.md`에만 두자 `CoverageIncomplete` 8건 / rc=1. 코퍼스 위반이 0이었던 이유가 "규약을 따른 SPEC이 없어서"임이 이로써 확인된다. §2.3의 최소 중복 표로 우회했다.
2. **`MissingExclusions`의 숫자 접두 가설은 반증됐다 (확정).** `OutOfScopeRule.Check`(`lint.go:896`)는 `###` 접두 + "out of scope"를 요구한다. `## 2. Out of Scope`는 H2라 절에 진입하지 못하며, 숫자 접두는 무관하다. 반례: 이 SPEC의 `## 4. Exclusions`(숫자 접두 H2) + `### Out of Scope —` H3 조합이 rc=0 통과. 코퍼스 형태: **H3 형태 684개 / H2뿐 2개**. 미검증 가설이 아니라 코드로 반증된 사실이다. 코퍼스 정리는 범위 밖(spec.md §4).
3. **같은 파싱 경로의 세 번째 협소성 — 협소성은 진짜, 파급은 741 **안**에 있다.** `findACSectionStart`(`parser.go:64`)가 문서의 **첫** `##`+"acceptance" 줄을 AC 절 시작으로 잡고 `extractACLines`가 다음 `##`에서 멈춘다. `acceptance.md`를 언급하는 산문 표제가 앞에 있으면 AC가 0개로 파싱된다. 작성 중 실제 발생했고, 협소성 자체는 실재한다.

   **다만 그 결과를 "741에 없는 증분"으로 적었던 최초 서술은 틀렸다.** 측정으로 반증됐다(`.moai/reports/t362/ac_section_reach.py`, `cross_check.py`, 같은 트리 `ee50984ab`):

   | 측정 | 값 |
   |---|---|
   | `spec.md` 총수 | 703 |
   | `findACSectionStart`가 AC 절에 **도달** | 326 |
   | 도달 못했는데 **AC 토큰은 보유** | **0** |
   | 한국어 AC 표제(인수/수용) 보유 | 47 |
   | 넓은 REQ 패턴에 걸리는 `spec.md` | 63 → 도달 47 / 미도달 16 |
   | 미도달 16개의 AC 토큰 수 | **전부 0** |

   즉 파서가 진입하지 못하는 16개는 **spec.md에 AC를 애초에 두지 않는다.** 파서가 있는 AC를 잃는 것이 아니다. 그리고 741 시뮬레이션은 `maps REQ-…` 출현으로 커버리지를 셌으므로, AC 토큰이 없는 파일은 `maps`도 없어 **그 REQ들은 이미 미커버로 계산돼 있었다.** 이 집단은 741 밖이 아니라 안이다.

   **진짜 오차는 반대 방향이다.** 시뮬레이션은 `maps REQ-…`를 문서 어디에 있든 커버로 셌지만, Go 파서는 AC 절 **안**의 출현만 읽는다(`extractACLines`는 `findACSectionStart`에서 시작해 다음 `##`에서 끊긴다). AC 절 밖의 `maps`는 시뮬레이션에서 커버, 실제로는 미커버다. 이 집단은 파서가 **도달하는** 47개다. 따라서 **741은 하한(under-estimate)으로 행동할 가능성이 크다** — 최초 서술이 말한 방향의 반대다.

**넓힘 시뮬레이션 (추정치 — baseline 아님).** `.moai/reports/t362/widen_sim.py`가 `^\s*[-*]\s+\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*:` 패턴을 적용하고 `maps REQ-…` 커버리지를 교차한 결과:

| 항목 | 현행 | 넓힌 뒤 (추정) |
|---|---|---|
| REQ 정의 줄 | 231 | **1,077** |
| REQ를 갖는 `spec.md` | 15 | **62** |
| 미커버 REQ를 가진 SPEC | — | **47** |
| 예상 `CoverageIncomplete` | 0 | **741 (추정)** |

상위: SPEC-V3R5-HARNESS-AUTONOMY-001 37, SPEC-HARNESS-CADENCE-BUILD-001 34, SPEC-AGENT-MODEL-ENFORCE-001 28.

**이 741은 두 가지 이유로 추정치다.** ① Python 근사이지 Go 파서가 아니다. ② 넓힌 패턴은 조사자가 고른 하나이며, 다른 넓힘은 다른 수를 낸다. 이 값이 확립하는 것은 자릿수뿐이다 — 0이 아니라 수백. run 단계에서 실제 파서로 재측정한다(REQ-CRS-001-004).

## §C 결함 ① 방향 결정 — 어느 모양으로 고칠 것인가

### C.1 관행과 린터 중 어느 쪽이 옳은가

출처를 다시 읽어 논증한다.

- `spec-assembly.md:55`는 "AC is inline in `spec.md §3`"을 **Tier S의 성질로만** 적는다. 56행(Tier M)과 57행(Tier L)은 `acceptance.md`를 요구하고 spec.md AC 요건을 두지 않는다.
- `manager-develop-prompt-template.md:131`은 AC 계수의 SSOT를 `acceptance.md`로 못박으며, `progress.md`가 아니라고 명시한다.

따라서 린터는 Tier S 전제를 전 Tier에 적용하고 SSOT 대신 관습적 복사본을 읽는다. 권고 방향은 **"관행이 옳다 — lint가 `acceptance.md`를 읽어야 한다"**이다.

반대 방향도 성립한다: Tier M/L에도 spec.md AC 블록을 요구하도록 규약을 고친다. 이 경우 `manager-develop-prompt-template.md:131`의 SSOT 조항도 함께 고쳐야 하고, 62개 SPEC에 AC 중복 기재를 강제하게 된다. 중복은 두 사본이 갈릴 자리를 만들므로 SSOT 원칙과 충돌한다. 권고하지 않는다.

[NEEDS CLARIFICATION: 관행 존치(권고) vs 규약 개정 — 운영자 판정]

### C.2 수리 모양 (i) vs (ii)

`internal/spec/lint_artifact_status.go`(t357, SPEC-ARTIFACT-STATELESS-001 M2)는 `ArtifactStatusFieldForbiddenRule`이 `doc.Body`가 아니라 형제 산출물을 읽는다고 적으며, 같은 일을 하는 선행 규칙으로 `MovingRefUnpinnedRule`과 `HaikuResidualRule`을 지목한다. **단일 파싱 문서 밖으로 손을 뻗는 규칙은 이 패키지의 확립된 패턴이다.**

| | (i) `parseSPECDoc`이 `acceptance.md`를 읽어 `doc.Criteria`에 병합 | (ii) `CoverageRule.Check`가 형제 `acceptance.md`를 직접 읽음 |
|---|---|---|
| 파급 | `doc.Criteria`를 쓰는 **모든** 규칙에 영향 | `CoverageRule` 한 곳 |
| 선례 | 없음 | 세 건(MovingRef / Haiku / ArtifactStatus) |
| 위험 | `ACIDInvalid` 등 다른 규칙이 acceptance.md 내용에 대해 예상 밖으로 발화 | 낮음 |
| lint.skip / era 강등 | per-SPEC 모양 유지되나 영향 규칙이 늘어남 | 선례와 동일하게 유지 |

**권고: (ii).** 선례에 맞고 파급이 좁다. (i)은 "AC의 정본이 하나가 된다"는 개념적 이점이 있으나, 그 이점을 얻으려면 `doc.Criteria` 소비자 전체를 재측정해야 하며 이 SPEC의 범위를 Tier L로 밀어 올린다.

## §D 결함 ② 심각도 결정 — 코퍼스를 붉히지 않고 넓히는 법

### D.1 t342와 t357은 반대 이유로 같은 자리에 있다

`lint_artifact_status.go:62-71`이 기록한다: `MovingRefUnpinned`(t342)와 `ArtifactStatusFieldForbidden`(t357)은 **둘 다** `eraDemotableCodes` 밖에 있지만 이유가 정반대다.

- **t342**: `warning`을 내고 `Advisory: true`를 발화 지점에서 설정한다. 그 맵은 `SeverityError`에 대해서만 참조되므로 warning은 애초에 맵에 닿지 못한다. 발화 지점이 유일한 지렛대다 — **구조적 필연**.
- **t357**: `error`를 내므로 맵이 닿는다. 밖에 있는 것은 **살아있는 선택**이고, 코퍼스 정리가 같은 SPEC 안에 착지하기 때문에만 감당 가능하다.

**한쪽을 다른 쪽의 선례로 읽으면 둘 다 뒤집힌다.**

### D.2 이 SPEC에 걸리는 결과

넓힌 패턴이 정리 없이 ~741건의 `error`를 착지시키면 코퍼스가 착지 순간 붉어진다. 이것이 정확히 t357의 REQ-AST-001-010이 막으려던 결과다. 선택지 셋:

| 안 | 내용 | 대가 |
|---|---|---|
| **A** | t342 모양 — `warning` + `Advisory: true`를 발화 지점에서 | `--strict`가 warning을 error로 올리므로 strict 실행에서는 여전히 붉다. t357 카드(M2)가 같은 문제를 겪었다 |
| **B** | t357 모양 — `error` 유지 + 같은 SPEC 안에서 코퍼스 정리 | 47개 SPEC의 AC 매핑 작성. Tier를 L로 밀 가능성이 크다 |
| **C** | 넓힘을 단계적으로 — 1차는 파서만 넓히고 `CoverageRule`은 기존 좁은 집합에만 적용, 2차에서 확대 | 두 단계로 나뉘어 카드가 길어진다. 다만 자릿수를 실측한 뒤 심각도를 정할 수 있다 |

**권고: C → (실측 후) A 또는 B.** REQ-CRS-001-004가 재측정을 요구하는 이유가 이것이다. 741이 추정치인 이상, 심각도를 지금 확정하는 것은 재보지 않은 수 위에 정책을 세우는 일이다.

[NEEDS CLARIFICATION: 심각도 안 A/B/C 선택 — 실측 결과를 본 뒤 운영자 판정]

## §E 마일스톤

| M | 내용 | 우선도 | 근거 REQ |
|---|---|---|---|
| **M1** | 넓힌 패턴을 실제 Go 파서로 구현하고, `CoverageRule`을 켜지 않은 상태에서 전 코퍼스 REQ 수집량을 재측정한다. 741의 실측 대응값을 얻는다 | High | REQ-CRS-001-001, -004 |
| **M2** | `reqIDPattern`과 `reqLinePattern`의 집합 정합을 맞춘다. `InvalidREQID`가 코퍼스 전역에서 발화하지 않음을 측정으로 보인다 | High | REQ-CRS-001-002, -005 |
| **M3** | M1 실측값을 근거로 §D의 심각도 안을 확정하고 적용한다. `develop`이 착지 시점에 붉어지지 않음을 CI로 확인한다 | High | REQ-CRS-001-003 |
| **M4** | `CoverageRule.Check`가 형제 `acceptance.md`를 읽도록 한다(§C.2 (ii)안). 회귀 쌍 픽스처를 함께 넣는다 | Medium | REQ-CRS-001-006, -007, -008 |

M4를 뒤에 둔 것은 잠복 결함이라 급하지 않아서이고, M1-M3이 파서를 넓힌 뒤에 붙어야 두 수리가 같은 파서 위에서 한 번만 검증되기 때문이다.

## §F Tier 제안

**제안: Tier M (spec.md + plan.md + acceptance.md).**

근거: 대상 패키지 하나(`internal/spec`), 수정 파일 소수(`lint.go` + 신규 규칙 파일 1개 + 테스트), 설계 결정 두 건(§C.2, §D)이 있으나 둘 다 이 계획서 안에서 논증 가능하며 별도 `design.md`를 요구하지 않는다.

**Tier L로 밀어 올리는 조건:**

- §D에서 **B안(같은 SPEC 내 코퍼스 정리)**이 채택되는 경우. 47개 SPEC의 AC 매핑 작성은 별도의 설계·연구 문서를 요구한다.
- M1 실측값이 741을 크게 초과해 넓힘 패턴 자체를 재설계해야 하는 경우.
- §C.1에서 규약 개정 방향(반대 방향)이 채택되는 경우. 규약 문서와 62개 SPEC이 함께 움직인다.

Tier 판정은 리드 몫이다. 이 절은 제안이며 프론트매터의 `tier: M`은 작업값이다.

## §G 위험

| 위험 | 영향 | 완화 |
|---|---|---|
| **넓힘이 착지 시점에 코퍼스를 붉힌다** (가장 큰 위험) | develop 적색, 다른 레인 차단 | §D의 단계적 C안 + M1 실측 선행. 심각도 결정을 실측 뒤로 미룬다 |
| 741 추정치를 baseline으로 오독 | 재지 않은 수 위에 정책을 세움 | REQ-CRS-001-004가 재측정을 요구. 문서 전체에서 "추정" 라벨을 매 언급마다 붙였다 |
| 넓힌 추출이 `InvalidREQID`를 대량 발화 | 다른 규칙으로 붉음이 번짐 | M2가 두 패턴의 집합 정합을 명시적으로 다룬다 |
| (i)안 선택 시 `doc.Criteria` 소비자 전체가 영향 | 예상 밖 규칙 발화 | §C.2 권고는 (ii)안 |
| 회귀 쌍의 절반만 관측 | "규칙을 껐다"와 구분 불가 | acceptance.md의 AC-CRS-001-006/007 쌍이 양방향을 강제 |
| **조사 자체가 거짓 형태를 만들 수 있다** | 무효 픽스처가 결론을 뒤집음 | spec.md HISTORY에 실사례 기록. 픽스처는 파서 술어로 직접 검증한 뒤 쓴다 |
| **741은 하한일 가능성이 크다 — 시뮬레이션의 커버리지 술어가 파서보다 넓다** | 실제 파급이 추정보다 크다 | §B 관측 3. 오차원은 AC 절 **밖**의 `maps REQ-…`를 커버로 센 것이며, 해당 집단은 파서가 도달하는 47개다. 미도달 16개는 AC 토큰이 0이라 이미 741 안에 계산돼 있다. M1이 실제 Go 파서로 재측정해 방향(하한)을 확인한다 |
| §2.3 최소 중복 표가 수리 후 남는다 | 결함 ①이 고쳐진 뒤에도 관행 위반 잔재 | M4 완료 시 삭제 대상임을 표 본문에 명시해 두었다 |

## §H 반패턴

- 넓힘의 파급을 "확인했다"고 적으면서 Python 근사를 근거로 대는 것.
- `MovingRefUnpinned`의 warning 처리를 "그러니 우리도 warning이면 된다"의 선례로 읽는 것(§D.1).
- 회귀 검증에서 "관행 픽스처가 통과한다"만 관측하고 끝내는 것 — 규칙을 비활성화해도 같은 결과가 나온다.
- `MissingExclusions` 가설을 확인 없이 결함으로 적는 것.

## §I 교차 참조

- `internal/spec/lint.go` — `discoverSPECs`(313), `reqLinePattern`(452), `parseSPECDoc`(456), `CoverageRule.Check`(682)
- `internal/spec/lint_artifact_status.go` — 형제 산출물 판독 선례 + t342/t357 대비 문단(62-71)
- `internal/spec/parser.go:218` — AC ID 4-분절 패턴
- `.claude/skills/moai/workflows/plan/spec-assembly.md:55-57` — Tier별 산출물 규약
- `.claude/rules/moai/development/manager-develop-prompt-template.md:131` — AC SSOT 조항
- `.moai/reports/t362/widen_sim.py` / `widen-sim-output.txt` — 넓힘 시뮬레이션(추정)
- `.moai/reports/t362/lint-corpus-ee50984ab.json` — 전 코퍼스 측정 원본
