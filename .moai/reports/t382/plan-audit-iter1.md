# SPEC 감사 보고서: SPEC-ERA-H3-NARROWING-001

Iteration: 1/1 (Tier S — `harness.yaml` `plan_audit_tier_ceilings.S = 1`, 단일 패스)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.825** (Tier S PASS 기준선 0.75 — `spec-workflow.md` §SPEC Complexity Tier)

감사 트리: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t382`, 브랜치 `WT-era-plan-phase`, HEAD `f967089ba`.
측정 도구: `./bin/moai` (트리 로컬). 자기보고 빌드 `v3.1.2-954-g9328a5242`, 빌드 시각 2026-08-31T10:15:33Z.
감사자 프로브: `.moai/reports/t382/audit_verify1.py` · `audit_verify2.py` · `audit_verify3.py` · `audit_verify4.sh` (이 감사가 새로 쓴 재측정 스크립트 — 카드가 만든 것이 아니다).

M1 컨텍스트 격리 고지: 지시문에 담긴 저자 추론은 배제했다. 다만 Tier S 입력 계약에 따라 `spec.md` · `plan.md` · `progress.md`와 카드가 디스크에 남긴 측정 원장은 감사 입력으로 읽었다. `acceptance.md` 부재는 Tier S 규정 산출물 집합이므로 결함으로 보고하지 않는다.

---

## Must-Pass 결과

- **[PASS] MP-1 REQ 번호 일관성** — `spec.md` §2에 `REQ-EH3-001` ~ `REQ-EH3-006` 여섯 개가 연속하며 결번·중복·자릿수 불일치가 없다. Tier S REQ 상한 8 이내.
- **[PASS] MP-2 GEARS 형식 준수** — 요구 계층(`REQ-XXX`)에 대해서만 판정했다. 001·004 unwanted(`shall not`), 002 `Where` 절, 003 `When` 절, 005·006 ubiquitous(`shall`). 여섯 개 모두 GEARS 다섯 패턴 중 하나에 맞는다. 검증 계층(`AC-XXX`, §3 인라인)의 Given-When-Then은 정상 형식이므로 여기서 감점하지 않고 Group 4에서 채점했다. 단 REQ-EH3-002의 패턴 **라벨**은 틀렸다 — D8 참조.
- **[PASS] MP-3 YAML frontmatter 유효성** — 정본 12필드가 모두 있고 타입이 맞다(`id`/`title`/`version:"0.3.0"`/`status:draft`/`created:2026-08-31`/`updated:2026-08-31`/`author`/`priority:P1`/`phase`/`module`/`lifecycle:spec-anchored`/`tags`, 추가로 `tier:S`). 거부되는 snake_case 별칭 없음. 기계 확인: `./bin/moai spec lint --strict --json .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md` → 축자 stdout `[]`, rc 0.
- **[N/A] MP-4 §22 언어 중립성** — 대상이 `internal/spec` Go 코드와 로컬 전용 룰 문서 한 곳으로 한정된 단일 언어 SPEC이다. 템플릿 미러 부재는 `plan.md` §D가 `grep -rln "H-3" internal/template/templates/` 근거로 명시했다. 자동 통과.
- **[PASS] MP-5 D7 교차 SPEC 조정** — 본문이 참조하는 SPEC-ID 다섯 개를 추출해 status를 전부 읽었다. `SPEC-KANBAN-TODO-CLI-001` in-progress · `SPEC-KANBAN-WORKTREE-001` draft · `SPEC-V3R5-INIT-WIZARD-EXPANSION-001` implemented · `SPEC-HOOK-PREEDIT-INVESTIGATE-001` **superseded** · `SPEC-DESIGN-CONST-AMEND-001` draft(`_archive/`). 종결 상태인 두 건 모두 조정돼 있다 — superseded 건은 「이미 `terminal-exempt`라 변화 없다」로 그 종결 상태 자체를 근거로 인용했고, `_archive/` 건은 「감사 대상 밖」으로 명시 배제했다. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall'` 결과 spec.md 0 · plan.md 0. 자동 통과.
- **[PASS] MP-7 clarification 게이트** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-ERA-H3-NARROWING-001/` 무매치. `research.md`는 Tier S라 존재하지 않는다.

---

## 차원별 점수

| 차원 | 점수 | 루브릭 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 | 단위가 값마다 붙어 있고(§1.6 표), 기제를 결론 대신 코드 좌표로 적는다. 감점: §1.2·§1.3이 「오늘 22건」이라 쓰는데 현재 트리 실측은 23건(D5), AC-EH3-008이 「H-5 술어가 참」을 테스트가 **어떻게** 평가하는지 미지정(D3) |
| Completeness | 0.85 | 0.75–1.0 | HISTORY·배경·요구·AC·범위밖(H3 소제목 4개, 각각 불릿 보유)·설계 대안 4종·형제 카드 의존까지 있다. 감점: 무게중심으로 선언한 drift 축에 4요소 RED 셀이 없다(D1) |
| Testability | 0.80 | 0.75–1.0 | AC 여덟 개 전부 판정 명령이 붙고 이진 판정이다. 오늘 이미 통과하는 AC-003·004에는 뮤테이션을 배정해 결정력을 부여했다. 감점: AC-007이 폐기 선언된 기준선 파일에 대고 잰다(D1), AC-002가 인용한 RED가 자기 특징을 구별하지 못한다(D2) |
| Traceability | 0.75 | 0.75 밴드 | REQ 여섯 중 다섯이 AC로 덮인다. REQ-EH3-004(불변 지점 네 곳 미변경)는 어떤 AC도 §3.6 체크리스트 항목도 검사하지 않는다(D4). REQ-EH3-006은 AC가 없으나 §3.6에 grep 판정 항목이 있어 덮인 것으로 본다 |

산술 평균 0.825. Tier S 기준선 0.75 초과.

---

## 확인된 것 — 리드가 지목한 일곱 축

**① 정당화의 크기.** 정확하다. §1.3이 「게이트가 실패를 통과시키고 있다」는 주장을 **명시적으로 반증**하고, §1.2 축 3에서 오늘 rc 델타 0을 스스로 적는다. 재측정으로 확인: 비-terminal V3R5 15건에 `./bin/moai spec lint --strict --json` → **rc 0**, findings 11건 전부 `advisory=true`(`FrontmatterInvalid`×7 / `MovingRefUnpinned`×3 / `StatusGitConsistency`×1). 강등된 ERROR 보유 SPEC이 1건뿐이라는 주장도 맞다(7건 전부 INIT-WIZARD). 반대 방향의 완화(「그러니 별것 아니다」)도 §1.3-2와 §G에서 차단한다. 부풀림도 얼버무림도 없다.

**② RED 셀의 실재성.** R1·R2·R3 셋 다 HEAD에서 축자 재현된다.
- R1: `./bin/moai spec audit --json | python3 .moai/reports/t382/red_probe.py` → stdout `V3R5-classified SPECs: 24 / misclassified: 23 / correctly V3R5: 1 / no-signal set: ['SPEC-V3R5-INIT-WIZARD-EXPANSION-001']`, **rc 1**. 원장과 바이트 동일.
- R2: `./bin/moai spec audit` → `Total SPECs: 715 / Grandfathered: 286 / Modern-era clean: 422 / Drift findings: 500`, rc 0. 파생 429는 715−286이며, 독립 교차검증도 성립한다 — 422(clean) + 7(SyncStatusDrift) = 429.
- R3: `--filter-spec SPEC-ERA-H3-NARROWING-001` → `Grandfathered: 1`, `[INFO] ... (V3R5)`, rc 0. 원장과 동일.

AC-001의 RED 이관은 **판정력을 보존한다**. `go test -run TestClassifyEra ./internal/spec/`는 오늘 실제로 초록이며(재실행 확인), 서브테스트 30개가 도는 비-공허 초록이다. 부재하는 서브테스트가 실패할 수 없다는 사실을 SPEC이 정면으로 적고 RED를 코퍼스로 옮긴 것은 옳다 — 코퍼스 R1은 같은 명제를 실데이터에서 23회 거짓으로 만들므로 오히려 더 강하다. 다만 AC-002는 사정이 다르다(D2).

오늘 통과하는 나머지 기준(AC-003·004)은 전부 뮤테이션을 본체로 갖고, plan §F M2가 뮤테이션 3종을 열거한다. 「아무것도 결정하지 않는 기준」은 D2 한 건을 빼면 없다.

**③ 양방향성.** 갖춰져 있다. AC-005가 총계·원소 일치·건별 근거·표본 5건 네 명제를 요구하고, 명제 3(근거 없이 시대가 바뀐 건이 하나라도 있으면 실패)이 grandfather 손실 벡터를 매 회 재측정하는 장치다. 원소 산술도 맞는다 — `v3r5-population.txt` 23행 = 현재 V3R5 24건 − 이 SPEC 자신, 이동 대상 23건 = 24 − INIT-WIZARD. `EraUnclassified` 0을 기준선으로 실제로 사용한다(AC-004 + AC-005 명제 1). `unclassified` 비용도 코드를 읽어 값을 매겼고 그 판독이 맞다 — `audit.go:311-322`에서 INFO finding 하나를 내고 조기 return 하므로 `checkV3R6Drift`가 돌지 않고, `lint.go:286` `isGrandfatheredSpecDir`와 `drift.go:186`이 `EraFinal()` 거짓으로 노출된다. 게다가 채택 설계는 유예를 **양성 신호에 게이트**하므로 새 H-6 낙하를 원리상 만들지 못한다 — 논증이 맞다.

**④ 자기 사례 주장.** 정직하게 다룬다. §3.2가 좌우 두 열을 나란히 놓고 「왼쪽 열의 수를 사후 근거로 재사용하지 않는다 — 그것이 baseline 귀속 위반이다」를 [HARD]로 적으며, run-phase에 M1 착수 직전 재측정을 지시한다. 핀도 맞다 — `measurements-9328a5242.md`는 SPEC 디렉터리 생성 전, `red-evidence.md`는 산출물 착지 후. 다만 그 규율이 AC-007에서 스스로 깨진다(D1), 서술문 두 곳에서 좌측 값이 「오늘」로 남아 있다(D5), R1의 귀속 트리가 프로브 파일보다 앞선다(D6).

**⑤ 설계 선택.** 논증돼 있다. 옵션 A는 rationale 정밀도 상실이라는 코드가 선언한 가치로, 옵션 B는 측정(23건 중 21건이 이미 §E.4 보유 — 재측정 22/24)으로, 옵션 D는 범위(46건 중 1건) + 자기차폐 영구화로 각각 기각한다. `created` 키잉의 건전성도 확인된다: `created_at` 46건의 시대 분포를 재측정하니 V2.x 31 · V3R2-R4 14 · V3R5 1이다. 즉 45건은 전부 H-1/H-2에서 잡혀 **어떤 날짜 휴리스틱에도 도달하지 않으므로** `created_at`이 era 엔진에 영향을 줄 수 없다. 이것이 기각의 실제 지지 사실인데, SPEC은 대신 「이미 올바르게 분류돼 있다」는 측정하지 않은 정확성 주장을 적었다(D7). 결론은 옳고 근거 문장이 약하다.

**⑥ 가드.** 정직한 한계 서술이다. 2층 분리의 근거(단위=합성 입력, 코퍼스=실 카탈로그, 서로가 못 잡는 것을 명시)가 구체적이고, 「`red_probe.py`는 CI에 배선되지 않으므로 상시 가드가 아니다」를 스스로 적으며 그 이유까지 든다 — 상시 가드로 소개하면 「멈춘 검사가 성공과 구별되지 않는」 결함을 새로 만든다는 것. 상시 층으로 지목한 단위 불변식이 실제로 CI에서 도는 것도 트리에서 확인된다(`.github/workflows/ci.yml:208` `go test ... ./...`). 두 층 모두 known-bad 입력에서 실패를 요구한다 — 코퍼스는 이미 관측된 rc 1, 단위는 뮤테이션. 결함을 포장한 것이 아니다.

**⑦ §6 t371 — 지시문 정정.** **SPEC이 맞고 지시문이 틀렸다.** 두 갈래를 직접 읽어 확인했다.
- `lint.go:1335` — `Advisory: true, // heuristic git-implied signal — never strict-escalated`, severity `SeverityWarning`. 발화 지점에서 이미 advisory다.
- `lint.go:296-312` `applyEraDemotion`의 첫 갈래(`f.Severity == SeverityError && eraDemotableCodes[f.Code]`)는 error를 요구하므로 warning인 이 finding에 닿지 않는다. 둘째 갈래(`lint.go:307` `case f.Severity == SeverityWarning:`)는 **코드로 게이트되지 않아 실제로 닿는다** — 하는 일은 `f.Advisory = true` 대입뿐이고 그 값은 이미 참이다.
- `eraDemotableCodes`는 `lint.go:272-275`에 `MissingExclusions`·`FrontmatterInvalid` 둘뿐.

따라서 「demotable 코드가 아니라서 닿지 못한다」는 요약은 부정확하고, 독립성이 실제로 서 있는 곳은 **발화 지점의 `Advisory: true`가 만드는 멱등성**이다. SPEC이 그 구분을 명시적으로 붙들고 요약을 거부한 것이 옳다. 인용 좌표 셋은 HEAD에서 모두 정확하며, `9328a5242` 이후 `.go` 변경이 0이라 `1f10f5e8d` 핀도 유효하다.

---

## Defects Found (심각도 순)

**D1. AC-EH3-007이 스스로 폐기 선언한 기준선에 대고 판정한다 — 그것도 무게중심 축에서**
`spec.md:198-206` (AC-EH3-007; 「오늘의 RED」는 `spec.md:204`) · 참조 파일 `.moai/reports/t382/drift-before-9328a5242.txt`
Severity: **major** — Class: **blocking**

§3.2가 [HARD]로 「아래 AC의 기대값은 전부 오른쪽 열 기준이다. 왼쪽 열의 수를 사후 근거로 재사용하지 않는다」고 선언한다. AC-007은 정확히 왼쪽 열(`9328a5242`)의 파일을 비교 기준으로 지목하고 그 열의 수를 기대값으로 쓴다 — 「`era-exempt`였던 **22행** 중 **21행**이 바뀐다」.

현재 트리 실측(`./bin/moai spec drift`, 감사자 프로브 `audit_verify2.py`): V3R5 24건의 drift 행은 `era-exempt` **23** + `terminal-exempt` **1**(`SPEC-HOOK-PREEDIT-INVESTIGATE-001`). 따라서 갱신된 기준선에서 옳은 기대값은 「23행 중 **22행**이 바뀌고 1행(INIT-WIZARD)이 남는다」이다. AC-007은 전 항목이 한 칸씩 어긋나 있고, 「이 SPEC 자신의 행도 함께 확인한다」는 부기는 그 사실을 인지하되 산술을 고치지는 않는다.

더 무거운 쪽은 증거 등급이다. §3이 「판정이 실제로 기대는 RED는 R1~R3뿐」이라 못박고 M-번호와 배경 파일을 「축자 stdout과 exit code를 갖지 않는다」는 이유로 판정 근거에서 배제한다. 그런데 AC-007의 「오늘의 RED」는 바로 그 배경 등급 파일이다 — `drift-before-9328a5242.txt`는 명령도 exit code도 담지 않은 23행 raw 출력이고, 트리 SHA는 파일명에만 있다. **무게중심으로 선언한 축(§1.2 축 1)이 여덟 AC 중 유일하게 4요소 RED 셀 없이 서 있다.**

Required fix: `red-evidence.md`에 **R4** 셀을 신설한다 — 명령 `./bin/moai spec drift`, V3R5 24건 행의 축자 발췌(`era-exempt` 23 + `terminal-exempt` 1), exit code, 트리 SHA. AC-007의 기대값을 22행→21행에서 **23행→22행**으로 정정하고 비교 대상을 R4로 바꾼다.

---

**D2. AC-EH3-002가 인용한 RED가 AC-002의 구별 특징을 판정하지 못한다**
`spec.md:156-159` (AC-EH3-002; 「오늘의 RED」는 `spec.md:158`)
Severity: **major** — Class: **blocking**

AC-002는 「`created`가 비고 `phase`만 modern일 때 H-3을 통과한다」는 명제다. 그 구별 특징은 **phase 경로 단독**이다. 그런데 인용된 R1은 phase와 created를 OR로 묶어 판정하며, 현재 카탈로그에서 phase 매치 5건은 **전부 post-threshold `created`도 함께 갖는다**.

재측정(`audit_verify1.py`): `matchesModernPhase` 참 5건 = `SPEC-DWF-CODEMAPS-PILOT-001` · `SPEC-HOOK-PREEDIT-INVESTIGATE-001` · `SPEC-INTERNAL-ARCH-001` · `SPEC-LSPMCP-RETIRE-001` · `SPEC-V3R6-SESSION-HANDOFF-AUTO-001`. `created >= 2026-04-01` 23건 = 24건 중 INIT-WIZARD를 제외한 전부. 즉 phase 매치 집합 ⊂ 날짜 매치 집합이며, **교집합 밖은 공집합**이다.

결과: `hasModernEraSignal`을 날짜만으로 구현해도 R1은 `misclassified: 0` · rc 0으로 뒤집힌다. AC-002의 RED는 자기가 판정해야 할 phase 경로에 대해 공허하다. SPEC은 이 사실을 절반 알고 있다 — M8 주석에 「5건 모두 M4의 POST 22건 안에 포함」이라 적어 놓고도, 그 함의를 AC-002의 판정력 문제로 잇지 않았다.

구별력을 실제로 갖는 것은 두 가지다: AC-002의 단위 픽스처(`FrontmatterCreated: ""` + `phase: "v3.0.0"`)와 plan §F M2 뮤테이션 3종 중 「헬퍼를 날짜만으로 축소」. 후자는 plan에만 있고 어떤 AC에도 묶여 있지 않다.

Required fix: AC-EH3-002의 「오늘의 RED」에 R1이 phase 경로를 구별하지 못한다는 사실과 그 근거(phase 매치 5건 ⊂ 날짜 매치 23건)를 명시하고, 「헬퍼를 날짜만으로 축소」 뮤테이션을 AC-EH3-002의 판정 본체로 승격해 실패 관측을 요구한다.

---

**D3. AC-EH3-008이 「H-5 술어가 참」을 테스트가 어떻게 평가하는지 지정하지 않아, M1의 술어 동일성 주장이 동어반복으로 무너질 수 있다**
`spec.md:208-213` (AC-EH3-008) · `plan.md` §F M1
Severity: **minor** — Class: **blocking**

M1은 「H-5의 조건절을 헬퍼 호출로 치환해 두 지점이 같은 술어를 쓴다는 사실을 코드로 못박는다」를 명시 가치로 선언하고, §5 옵션 C가 「그 동일성 위에 핵심 불변식이 선다」고 적는다. 그런데 AC-008의 불변식이 「H-5 술어가 참」을 판정하는 방법이 미지정이다. 테스트가 `hasModernEraSignal(signals)`을 호출해 그 값을 판정에 쓰면, 헬퍼와 H-5 조건절이 나중에 갈라져도 불변식은 계속 통과한다 — 동일성을 검사해야 할 가드가 동일성을 가정하게 된다.

부분적으로는 AC-001·002가 rationale이 `"H-5"`로 시작할 것을 요구해 H-5 갈래 발화를 묶지만, 8축 곱집합을 도는 것은 AC-008이고 AC-008은 「결과가 `EraV3R5`가 아님」만 단언한다.

Required fix: AC-EH3-008에 판정 방식을 못박는다 — 테스트는 H-5 술어를 헬퍼 호출이 아니라 **독립 리터럴**(자체 날짜 비교 + phase 접두 검사)로 평가하거나, 결과 rationale이 `"H-5"`로 시작할 것을 함께 단언한다.

---

**D4. REQ-EH3-004를 검사하는 AC도 완료 정의 항목도 없다**
`spec.md:106` (REQ-EH3-004) · `spec.md` §3.6
Severity: **minor** — Class: **blocking**

REQ-EH3-004는 「`eraDemotableCodes` · `applyEraDemotion` · `Era.EraFinal()` · `Era.IsModern()`을 변경해서는 안 된다」는 unwanted 요구다. AC 여덟 개를 훑으면 이 명제를 판정하는 것이 없다 — AC-005 명제 1은 명시적으로 REQ-EH3-003에 묶여 있고(「변하면 H-3 이외가 건드려졌다는 뜻이며 REQ-EH3-003 위반」), AC-006은 강등 동작을 간접적으로만 지난다. §3.6 완료 정의 여덟 항목에도 해당 줄이 없다. plan §D가 「읽되 고치지 않는다」로 제약하지만 그것은 지시이지 판정이 아니다.

Required fix: §3.6에 판정 가능한 한 줄을 추가한다 — 예: `git diff --stat` 산출이 `internal/spec/era.go` · `era_test.go` · `.claude/rules/local/lifecycle-sync-gate.md` 셋으로 한정됨을 확인하고 출력을 남긴다.

---

**D5. §1.2·§1.3이 폐기된 좌측 열의 수를 「오늘」로 서술한다**
`spec.md:51` (§1.2 축 1 「오늘 22건에 걸려 있다」) · `spec.md:71` (§1.3-1 「22건이 … 통보를 받고 있는데」)
Severity: **minor** — Class: **optional**

두 문장 다 M13(트리 `9328a5242`) 값을 현재형으로 적는다. 현재 트리 실측은 `era-exempt` **23**행이다(D1의 재측정). §3.2가 좌측 열 재사용을 [HARD]로 금지하므로 자기 규율과 어긋난다. 서술문이라 판정에 실리지는 않으나, 다음 읽는 사람이 22를 현재값으로 가져갈 경로가 열려 있다.

Required fix: 두 문장에 트리 라벨을 붙이거나(`9328a5242 기준 22건 — 현재 기준선은 23건`) 갱신된 값으로 바꾼다.

---

**D6. R1의 귀속 트리에 프로브 파일이 존재하지 않는다**
`.moai/reports/t382/red-evidence.md:1` (「tree `f72c0bf0f`」) · R1 명령
Severity: **minor** — Class: **optional**

R1은 트리 `f72c0bf0f`에서 측정했다고 적고, 명령은 `python3 .moai/reports/t382/red_probe.py`를 호출한다. 그런데 `red_probe.py`와 `red-evidence.md` 둘 다 **다음 커밋** `1f10f5e8d`에서 처음 추가됐다(`git diff --name-only f72c0bf0f..HEAD`가 둘을 신규로 낸다). 즉 인용된 트리를 체크아웃하면 그 명령은 실행되지 않는다. 측정 당시 프로브가 미추적 상태로 워킹트리에 있었을 것이므로 관측 자체는 정직하나, 트리 SHA가 「이 명령을 재현할 수 있는 트리」를 가리키지 않는다.

감사에서 확인: HEAD(`f967089ba`)에서 R1을 재실행하면 stdout과 rc가 원장과 **바이트 동일**하다. 값 자체는 무사하다.

Required fix: R1의 귀속을 `1f10f5e8d`(프로브가 처음 존재하는 트리) 또는 HEAD로 옮긴다.

---

**D7. 옵션 D의 기각 근거 한 문장이 측정하지 않은 정확성을 주장한다**
`spec.md:287` (§5 옵션 D 「나머지 45건은 다른 경로로 이미 올바르게 분류돼 있다」)
Severity: **minor** — Class: **optional**

M7이 지지하는 사실은 「46건 중 V3R5 버킷 소속은 1건」까지다. 「나머지 45건이 **올바르게** 분류돼 있다」는 분류 정확성 주장이며 어디서도 측정되지 않았다 — 버킷 밖이라는 사실은 분류가 옳다는 증거가 아니다.

감사 재측정(`audit_verify3.py`)이 더 강한 사실을 준다: 46건의 시대 분포는 V2.x **31** · V3R2-R4 **14** · V3R5 **1**이며, V2.x/V3R2-R4는 각각 H-1(progress.md 부재)·H-2(§E.* 마커 부재)에서 잡힌다 — **날짜 휴리스틱에 도달하기 전이다.** 따라서 `created_at`을 읽게 만들어도 이 45건의 분류는 원리상 바뀌지 않는다. 기각 결론은 유지되고 근거만 교체하면 된다.

Required fix: 문장을 측정된 기제로 바꾼다 — 「나머지 45건은 H-1(31건) 또는 H-2(14건)에서 잡혀 날짜 휴리스틱에 도달하지 않으므로 `created_at` 관용이 이들의 분류를 바꾸지 못한다」.

---

**D8. REQ-EH3-002의 GEARS 패턴 라벨이 틀렸다**
`spec.md:104` (REQ-EH3-002 `(capability-gate)`)
Severity: **minor** — Class: **optional**

GEARS의 `Where`는 capability gate / feature flag / static config를 가리킨다. REQ-EH3-002의 조건 「a SPEC이 modern-era 신호를 하나도 갖지 않는 경우」는 능력 게이트가 아니라 **상태 조건**이므로 `While`(state-driven)이 정확한 패턴이다. 구문 형태가 다섯 패턴 중 하나에 맞으므로 MP-2는 통과하나, 라벨은 부정확하다.

Required fix: `(state-driven)`으로 바꾸고 `While` 절로 서술하거나, `Where`를 유지할 근거를 적는다.

---

**D9. plan §F M2가 존재하지 않는 테스트 이름을 지목한다**
`plan.md` §F M2 (「`TestClassifyEra` 테이블에 … 추가한다」)
Severity: **minor** — Class: **optional**

`internal/spec/era_test.go`에 `TestClassifyEra`라는 이름의 테스트는 없다. 실제 이름은 `TestClassifyEra_HeuristicTable`(era_test.go:11)이다. 판정 명령 `go test -run TestClassifyEra ./internal/spec/`는 접두 매치라 정상 동작하므로 실질 피해는 없다. 인용한 테이블 구조(`name` / `signals` / `wantEra` / `wantRule`)는 era_test.go:15-18과 정확히 일치한다.

Required fix: `TestClassifyEra_HeuristicTable`로 정정.

---

**D10. 측정 도구의 빌드 커밋이 인용 트리와 다른데 그 사실이 어디에도 적히지 않는다**
`.moai/reports/t382/red-evidence.md:5` (「이 트리에서 `make build`한 산출물」)
Severity: **minor** — Class: **optional**

`./bin/moai version` → `v3.1.2-954-g9328a5242`. 즉 바이너리는 `9328a5242`에서 빌드됐고, 원장이 인용하는 측정 트리는 `f72c0bf0f`다. `verification-claim-integrity.md` §2.2는 도구 측정에 **두 좌표**(읽은 트리, 판정한 빌드)를 요구한다.

이 경우 실해는 없다 — `git diff --name-only 9328a5242..HEAD -- '*.go'`가 **빈 출력**이므로 판정 로직이 트리와 동일하다. 그러나 그 비-지연 사실이 원장에 적혀 있지 않아, 읽는 사람은 검증할 수 없다.

Required fix: 원장 측정 도구 줄에 빌드 커밋과 「`9328a5242`..HEAD 사이 `.go` 변경 0」을 함께 적는다.

---

## Recommendation

Must-pass 일곱 항목이 모두 통과하고 총점 0.825가 Tier S 기준선 0.75를 넘으므로 **PASS-WITH-DEBT**다. Tier S 상한이 1회라 재감사 라운드는 없으며, 위 defect 목록이 그대로 수정 경로다.

설계 자체는 건전하다. 채택안(옵션 C)은 유예를 양성 신호에 게이트해 새 `EraUnclassified` 낙하를 원리상 배제하고, 기존 H-3 테스트 세 건이 `FrontmatterCreated`/`FrontmatterPhase`를 비워 둔다는 사실을 직접 읽어 확인했으므로 「기존 테스트를 하나도 깨지 않는다」는 주장도 참이다. §6의 t371 판독은 지시문보다 정확하며 그 정정이 이 SPEC의 가장 좋은 부분이다.

부채는 **증거 배치**에 몰려 있다. 무게중심으로 선언한 축이 유일하게 4요소 RED 없이 서 있고(D1), RED 하나가 자기 특징을 구별하지 못하며(D2), 불변식 가드가 자기가 검사할 동일성을 가정할 수 있다(D3). 셋 다 설계 변경 없이 문서·기준 수정으로 닫힌다.

**run-phase 착수 전 처리 (blocking 4건, 순서대로):**

1. **D1** — `./bin/moai spec drift`를 현재 트리에서 돌려 R4 셀(명령·축자·rc·트리 SHA)을 `red-evidence.md`에 신설하고, AC-EH3-007의 기대값을 `23행 → 22행 전환 + INIT-WIZARD 1행 잔류`로 정정한다.
2. **D2** — AC-EH3-002의 RED 서술에 R1의 phase-경로 비구별성을 명시하고, 「헬퍼를 날짜만으로 축소」 뮤테이션을 이 AC의 판정 본체로 승격한다.
3. **D3** — AC-EH3-008에 H-5 술어의 평가 방식(독립 리터럴 또는 rationale `"H-5"` 접두 단언)을 명시한다.
4. **D4** — §3.6에 REQ-EH3-004를 판정하는 한 줄(변경 파일 3개 한정 확인)을 추가한다.

**optional 6건(D5~D10)**은 정확성과 재현성을 높이지만 run-phase를 막지 않는다. D5·D6·D10은 세 줄 이내 수정이므로 위 네 건과 함께 처리하는 편이 싸다. D7~D9는 판단에 맡긴다 — 특히 D7은 결론이 아니라 근거 문장만 바꾸는 일이므로, 문장을 손대지 않고 두는 선택도 방어 가능하다.

**기준선 만료 주의.** 이 감사의 모든 수치는 HEAD `f967089ba`에서 잰 것이고, SPEC 자신이 §3.2에 적었듯 SPEC이 하나라도 추가되면 715·286·24가 함께 움직인다. run-phase는 M1 착수 직전 자기 트리에서 R1·R2(및 신설 R4)를 다시 돌려 갱신본에 대고 대조한다.
