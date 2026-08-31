# SPEC-ERA-H3-NARROWING-001 — 인수 기준

모든 기준선 수치는 트리 **`9328a5242`** 에서 이 트리의 `./bin/moai`(=`make build` 산출물)로 실측했다. PATH 바이너리는 쓰지 않았다. 재사용하기 전에 다시 잰다.

## §A 방향 선언

카드 지시문의 [HARD] 의무 — 「오분류가 멈췄음만 증명하는 기준 집합은 판정의 절반」 — 을 만족시키기 위해, 각 AC는 자신이 **어느 방향을 재는지**를 머리에 달고 있다.

| 방향 | AC |
|---|---|
| 오분류 중단 (수정이 일한다) | AC-EH3-001, AC-EH3-002, AC-EH3-005, AC-EH3-006 |
| **grandfather 보호 유지 (수정이 과하지 않다)** | AC-EH3-003, AC-EH3-004, AC-EH3-005 |
| **비용 측정 (수정이 새 문제를 만들지 않는다)** | AC-EH3-007 |
| 회귀 가드 (수정이 되돌아가지 않는다) | AC-EH3-008 |

AC-EH3-005는 양쪽에 동시에 선다 — 시대가 바뀐 SPEC과 **바뀌지 않은 SPEC**을 같은 표에서 한 건씩 귀속시키기 때문이다.

## §B 단위 기준 (`internal/spec/era_test.go`)

판정 명령(넷 공통): `go test -run TestClassifyEra ./internal/spec/`

### AC-EH3-001 — 날짜 신호가 있으면 H-3을 통과한다 (오분류 중단)

- **Given** `ProgressMDExists: true`, `ProgressMDContent`에 `## §E.2 Run-phase Evidence`가 있고 `sync_commit_sha`는 비어 있으며, `FrontmatterCreated: "2026-08-25"` (임계 `2026-04-01` 이후)
- **When** `ClassifyEra(signals)`를 호출하면
- **Then** `era == EraV3R6` 이고 rationale이 `"H-5"`로 시작한다

**오늘의 RED (실측).** 이 서브테스트는 아직 존재하지 않으므로 단위 층의 RED는 run-phase가 테스트를 심는 순간 관측된다. 코퍼스 층의 등가 RED는 이미 관측돼 있다 — `./bin/moai spec audit --json`이 `created >= 2026-04-01`인 SPEC **22건**을 `era: "V3R5"`로 보고한다(`.moai/reports/t382/measurements-9328a5242.md` M2·M4). 즉 이 명제는 오늘 카탈로그에서 22번 거짓이다.

### AC-EH3-002 — phase 신호가 있으면 H-3을 통과한다 (오분류 중단)

- **Given** AC-EH3-001과 같은 progress 신호에 `FrontmatterCreated: ""`, `FrontmatterPhase: "v3.0.0"`
- **When** `ClassifyEra(signals)`를 호출하면
- **Then** `era == EraV3R6` 이고 rationale이 `"H-5"`로 시작한다

**오늘의 RED.** 현행 코드에서 H-3이 `phase`를 전혀 읽지 않으므로 `EraV3R5`를 낸다. 코퍼스 등가: V3R5 23건 중 `matchesModernPhase` 참인 것이 5건(M8).

### AC-EH3-003 — 임계 이전 SPEC은 보호를 잃지 않는다 (보호 유지)

- **Given** AC-EH3-001과 같은 progress 신호에 `FrontmatterCreated: "2026-03-15"`, `FrontmatterPhase: ""`
- **When** `ClassifyEra(signals)`를 호출하면
- **Then** `era == EraV3R5` 이고 rationale이 `"H-3"`로 시작한다

**이 기준은 오늘 통과한다. 그래서 뮤테이션으로 결정력을 부여한다.** 수정 코드에서 유예 절(`&& !hasModernEraSignal(signals)`)을 **무조건 스킵**으로 바꾸면 — 즉 H-3을 그냥 지우면 — 이 서브테스트는 반드시 실패해야 한다. run-phase는 그 뮤테이션을 실제로 심어 실패를 관측하고, `.moai/reports/t382/`에 출력을 남긴 뒤 되돌린다. 관측 없이 "통과했다"만 보고하면 이 AC는 아무것도 결정하지 않는다.

### AC-EH3-004 — 신호가 전무하면 H-6으로 떨어지지 않는다 (보호 유지)

- **Given** AC-EH3-001과 같은 progress 신호에 `FrontmatterCreated: ""`, `FrontmatterPhase: ""`, `FrontmatterEra: ""` (`SPEC-V3R5-INIT-WIZARD-EXPANSION-001`의 실제 모양 — frontmatter가 거부된 snake_case 별칭을 써서 두 신호가 모두 빈 값으로 디코딩된다)
- **When** `ClassifyEra(signals)`를 호출하면
- **Then** `era == EraV3R5` 이고 rationale이 `"H-3"`로 시작한다. `EraUnclassified`가 아니다.

**뮤테이션 결정력.** AC-EH3-003과 같은 뮤테이션(H-3 무조건 스킵)에서 이 서브테스트는 `EraUnclassified`를 받아 실패해야 한다 — 그것이 곧 「H-6 낙하가 실제로 일어날 수 있는 세계」의 모습이고, 채택 설계가 그 세계에 있지 않음을 보이는 것이 이 AC의 일이다.

**`unclassified`가 실제로 무엇을 뜻하는지 — 코드를 읽어 확정한 답 (트리 `9328a5242`).** 설계는 이 상태를 만들지 않지만, 만들지 않는 것이 왜 옳은지는 값이 매겨져야 한다. 세 소비 지점이 서로 다르게 반응한다.

| 소비 지점 | `unclassified`의 취급 | 노출인가 보호인가 |
|---|---|---|
| `internal/spec/lint.go` `isGrandfatheredSpecDir` | `EraFinal()`가 거짓 → `applyEraDemotion` 미적용 | **노출** — 구조 게이트 ERROR 그대로, `--strict` 승격 가능 |
| `internal/spec/drift.go` (④ era 정렬) | `EraFinal()`가 거짓 → `era-exempt` 미부여 | **노출** — git 대조 분류 대상이 된다 |
| `internal/spec/audit.go` `auditSpec` | `era == EraUnclassified` 갈래에서 INFO `EraUnclassified` finding 하나를 내고 **조기 return** — `checkV3R6Drift`가 아예 돌지 않는다 | **어느 쪽도 아니다** — 보호받지도, 검사받지도 않는다. `grandfathered`에도 `modern_era_clean`에도 세어지지 않는다 |

요약: `unclassified`는 lint·drift 축에서는 노출이지만 **audit의 sync-drift 축에서는 침묵**이다. 침묵은 INFO finding 하나로 표시되므로 완전한 은폐는 아니나, MUST-FIX가 원리상 생길 수 없다는 점에서 V3R6과 다르다. 이것이 「신호 없음 → 재분류 없음」이라는 보수적 기본값을 택한 이유다.

## §C 코퍼스 기준

### AC-EH3-005 — 시대 변화를 SPEC 단위로 귀속한다 (양방향)

- **Given** 수정 전 트리 `9328a5242`와 수정 후 트리 각각에서
- **When** `./bin/moai spec audit --json`을 돌려 SPEC별 `era` 값을 뽑아 대조하면
- **Then** 다음 넷이 모두 성립한다:
  1. **총계 대조** — `grandfathered` 285 → **263**, V3R6 총계 429 → **451**, V3R5 23 → **1**, `EraUnclassified` finding 0 → **0**. (V2.x 144·V3R2-R4 118은 불변. 263 = 144 + 118 + 1.)
  2. **한 건씩 귀속** — 시대가 바뀐 SPEC은 정확히 **22건**이며, `.moai/reports/t382/v3r5-population.txt`의 23건에서 `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`을 뺀 집합과 **원소 단위로 일치**한다. 총계만 맞추는 것은 이 기준을 만족시키지 않는다.
  3. **바뀐 건마다 근거 제시** — 22건 각각에 대해 `created >= 2026-04-01`(22/22 해당, M4) 또는 `matchesModernPhase(phase)` 참(5/22)이 성립함을 SPEC별 표로 보인다. **근거가 없는데 시대가 바뀐 SPEC이 한 건이라도 있으면 실패**다 — 이것이 grandfather 손실 벡터를 매번 재는 장치다.
  4. **표본 검증** — 22건 중 최소 5건을 뽑아 각각의 `created` 값과 본문 HISTORY 첫 행의 날짜를 대조해 진짜 V3R6임을 확인한다. 표본 목록은 보고서에 이름으로 남긴다.

**오늘의 RED (실측).** `./bin/moai spec audit`이 `Grandfathered: 285`를 낸다. 위 명제 1은 오늘 거짓이다.

**[HARD] 기준선이 이 SPEC 자신 때문에 움직인다 — run-phase는 반드시 재측정한다.** 285/23/429라는 수는 이 SPEC 디렉터리가 생기기 **전**의 트리 `9328a5242`에서 잰 값이다. 그런데 이 SPEC 자신이 결함의 표본이다 — `./bin/moai spec audit --filter-spec SPEC-ERA-H3-NARROWING-001`이 `Grandfathered: 1`, `[INFO] SPEC-ERA-H3-NARROWING-001 (V3R5)`를 낸다(plan-phase 골격의 `§E.2` + 빈 `sync_commit_sha` + `created: 2026-08-31`). 따라서 plan-phase 산출물이 커밋된 뒤의 기준선은 **V3R5 24 / grandfathered 286 / V3R6 429**이고, 수정 후 기대값은 **V3R5 1 / grandfathered 263 / V3R6 452**, 시대가 바뀌는 SPEC은 **23건**이 된다. run-phase는 위 수를 그대로 쓰지 말고 M1 착수 직전에 자기 트리에서 다시 재어 이 표를 갱신한 뒤 대조한다. 재측정 없이 `9328a5242`의 수를 사후 근거로 재사용하면 그 자체가 baseline 귀속 위반이다.

**변하지 않아야 할 것.** V2.x 144건과 V3R2-R4 118건은 **한 건도** 움직여선 안 된다. 이 둘의 총계가 변하면 H-3 이외의 휴리스틱이 건드려졌다는 뜻이고, REQ-EH3-003 위반이다.

### AC-EH3-006 — 게이트 복원을 주입 결함으로 증명한다 (오분류 중단)

카드 지시문의 「오늘 이미 통과하는 기준은 아무것도 결정하지 않는다」를 이 AC가 정면으로 다룬다. **코퍼스 lint rc 델타로는 이 AC를 쓸 수 없다** — 재분류 대상 13건(non-terminal)은 강등 대상 ERROR를 오늘 하나도 갖고 있지 않아서, 수정 전후 모두 `--strict` rc가 0이기 때문이다(M11·M12). 그러므로 결함을 심어 잰다.

- **Given** 재분류 대상 22건 중 하나(예: `SPEC-KANBAN-WORKTREE-001`)의 SPEC 디렉터리를 `t.TempDir()` 아래로 복사하고, `spec.md`의 「Out of Scope」 H3 소제목을 제거해 `MissingExclusions`를 인위적으로 성립시킨 뒤
- **When** 그 사본에 대해 `moai spec lint --strict --json`을 돌리면
- **Then** 수정 **전** 바이너리는 rc **0** 이고 해당 finding이 `severity: warning`·`advisory: true`·메시지 말미 `[grandfathered era — downgraded to warning]`를 달고 나온다. 수정 **후** 바이너리는 rc **1** 이고 같은 finding이 `severity: error`·`advisory` 없음으로 나온다.

**오늘의 RED.** 위 「수정 전」 절이 곧 RED이며, run-phase가 실제로 두 바이너리로 각각 돌려 두 출력을 `.moai/reports/t382/`에 남긴다.

**주의 — 원본을 만지지 않는다.** 결함 주입은 반드시 임시 사본에서 한다. 실제 `.moai/specs/` 아래 SPEC을 훼손하면 이 카드의 범위를 벗어난다.

### AC-EH3-007 — drift 축의 비용을 잰다 (비용 측정)

- **Given** `.moai/reports/t382/drift-before-9328a5242.txt`(23행, 오늘 실측: `era-exempt` 22 + `terminal-exempt` 1)
- **When** 수정 후 `./bin/moai spec drift`를 돌려 같은 23건 행을 뽑으면
- **Then**:
  1. `era-exempt`였던 22행 중 **21행**이 git 대조 결과(`completed` / `in-progress` / `implemented` 등)로 바뀐다. 남는 1행은 `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`(V3R5 유지)이다. `SPEC-HOOK-PREEDIT-INVESTIGATE-001`은 이미 `terminal-exempt`이므로 변화가 없다.
  2. 새로 `DRIFT`로 표시된 행이 있으면, **그 각각에 대해** frontmatter status와 git 이력을 직접 대조해 진짜 불일치인지 확인하고, 건별 판정을 보고서에 남긴다. 진짜 불일치는 이 카드의 결함이 아니라 **드러난 기존 상태**이며, 오탐이면 그것은 이 수정이 만든 비용이다.
  3. **오탐이 1건이라도 확인되면 이 AC는 실패**이고, 설계를 다시 연다.

**오늘의 RED.** 위 파일의 22행이 전부 `era-exempt`인 것이 기준선이다.

## §D 회귀 가드

### AC-EH3-008 — 불변식 가드가 known-bad 입력에서 실패한다

- **Given** `internal/spec/era_test.go`에 새 불변식 테스트(가칭 `TestClassifyEra_NoV3R5WhileModernSignal`)를 심는다. 이 테스트는 `{progress §E.2 유무} × {sync_commit_sha 유무} × {created ∈ (빈값, 임계 이전, 임계 이후)} × {phase ∈ (빈값, "v3.0.0", "v3.2.0 target")}`의 조합을 순회하며, **「H-5 술어가 참인데 결과가 `EraV3R5`」인 조합이 하나도 없음**을 단언한다.
- **When** `go test -run TestClassifyEra_NoV3R5WhileModernSignal ./internal/spec/`를 돌리면
- **Then** 수정 후 코드에서 통과한다. 그리고 **뮤테이션에서 반드시 실패한다**: H-3 술어에서 `&& !hasModernEraSignal(signals)`를 제거(=수정 이전 상태로 되돌림)하면 이 테스트가 실패해야 한다.

**뮤테이션이 이 AC의 본체다.** 통과 관측만으로는 가드가 공허한지 알 수 없다. run-phase는 절을 실제로 지우고 실패를 관측한 뒤 되돌리며, 실패 출력을 `.moai/reports/t382/`에 남긴다. 되돌림 후 재통과도 함께 기록한다.

**가드를 선택한 이유.** 이 결함의 본질은 「첫 매치 승리 체인에서 앞선 절이 뒤 절을 굶긴다」이며, 재발 경로는 유예 조건이 어떤 리팩터링에서 조용히 사라지는 것이다. 개별 케이스 테스트 셋(AC-001~004)은 그 리팩터링이 **테스트에 없는 조합**으로 재발하면 잡지 못한다. 조합 순회 불변식은 그 구멍을 닫는다. 이것은 `internal/spec`의 기존 관행과도 일치한다 — `lint_movingref_test.go`·`lint_artifact_status_test.go`가 이미 「이 표식을 지우면 어떤 테스트가 실패해야 하는가」를 주석으로 명시하는 방식을 쓴다.

## §E 완료 정의 (Definition of Done)

- [ ] AC-EH3-001 ~ AC-EH3-008 전부 통과, 각 AC마다 판정 명령과 그 축자 출력이 `.moai/reports/t382/` 아래에 있다
- [ ] AC-003·004·008의 **뮤테이션 실패 관측**이 출력과 함께 기록돼 있고, 뮤테이션이 되돌려졌다
- [ ] `go test ./internal/spec/...` 통과 (전체 스위트 로컬 실행 금지 — 영향 패키지만)
- [ ] `go vet ./internal/spec/...` 통과
- [ ] `golangci-lint run` 통과
- [ ] `.claude/rules/local/lifecycle-sync-gate.md` H-3 행 갱신 (REQ-EH3-006)
- [ ] 모든 수치가 측정 명령과 트리 SHA에 귀속돼 있다
