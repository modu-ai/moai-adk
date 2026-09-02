# Plan — SPEC-ZONE-REGISTRY-RESYNC-001

## §A 맥락

카드 t232 / 이슈 #1616. 근거 보고서: `.moai/reports/t232/findings.md`(+ `validate-repro.txt`, `analysis-repro.json`, `analyze.py`). 작업 트리 `.claude/worktrees/t232`, 브랜치 `WT-zone-registry-drift`, base `294b4b6ab`.

### 카드 등급 판정

배차 시점 등급은 **Class B**(결함, 원인 미상)였다. 그러나 t232 조사로 원인이 세 층으로 **측정 확정**됐고(`spec.md` §2), 남은 작업에 설계 결정이 하나 있다 — 가드 표면. Class B 는 `plan` 을 건너뛰지만, 이 카드는 결정을 포함하므로 실질 **Class C**(설계 변경)로 다뤄 SPEC 을 세운다. 등급 상향의 근거는 속도가 아니라 결정의 존재다.

### Tier 판정 — M (예산 안에 들어감, 압축하지 않음)

세 축(clause 재동기화 + anchor 탐지·수리 + 가드)을 **전부 넣은 상태로** 예산을 확인했다.

| 축 | 값 | Tier M 상한 | 여유 |
|---|---|---|---|
| REQ | 15 (REQ-ZRR-001..015) | 16 | 1 |
| AC | 14 (AC-ZRR-001..014) | 14 이하 필요 → 16 | 2 |

`spec-workflow.md` § SPEC Complexity Tier 의 상한은 REQ 와 AC 에 **독립적으로** 적용된다(합계가 아님) — Tier M 은 REQ 16 **그리고** AC 16. 15/14 로 둘 다 들어간다.

**요구사항을 병합하거나 조용히 덜어낸 적 없다.** 상한에 맞추려는 압축은 커버리지 축소로 되돌아온다는 것이 오늘 다른 레인의 실측 사례이므로, 만약 세 축이 들어가지 않았다면 Tier L 로 올리고 그 사실을 보고했을 것이다. 여유가 REQ 1칸뿐이라는 점은 그대로 보고한다 — 다음 요구가 하나 더 붙으면 그때는 Tier L 이다.

- **S 아님**: 산출물이 단일 파일 편집이 아니다. 레지스트리 2 미러 + 신규 가드 + CI 배선 + 임베드 재생성이 한 변경 안에 있고, 마일스톤이 3개다.
- **L 아님**: Tier L 은 파일 > 15 또는 constitutional 범위를 요구한다. 이 카드가 만지는 파일은 5개 안팎이다 — 레지스트리 ×2, 가드 테스트 1, `ci.yml`, (선택) `Makefile`. 규칙 문서 본문은 건드리지 않는다(범위 밖). 새 아키텍처도, 되돌리기 어려운 인터페이스도 없다.
- 따라서 **Tier M**: `spec.md` + `plan.md` + `acceptance.md` + `progress.md`. `design.md` / `research.md` 는 만들지 않는다.

## §B 알려진 결함 (이미 측정됨)

1. clause 67건이 DRIFT (패러프레이즈 ~61 + 요약 라벨 6)
2. anchor 17건이 미해석 (그중 9건은 clause 통과 — 검증기가 못 봄)
3. `constitution validate` 를 부르는 곳이 0곳 (Makefile / CI / Go 테스트 전부)
4. CI `constitution-check` job 이 `continue-on-error: true` — 여기 배선하면 가드가 아니라 권고가 된다
5. `grep -F` 기준으로는 검증기 통과분 중 8건도 실패 (다중 행/공백 축약 clause)

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈추고 보고)

| 명령 | 기대값 |
|---|---|
| `moai constitution validate` (fresh init) | exit 1, DRIFT 67 |
| 로컬 리터럴 적중 | 25 / 101 |
| 템플릿 리터럴 적중 | 25 / 101 |
| anchor 해석 실패 | 17 |
| `grep -c "constitution validate" Makefile .github/workflows/*.yml` | 0 |
| `diff -q` 로컬 vs 템플릿 레지스트리 | 동일 |
| `go test ./internal/constitution/...` | ok |

**v0.5.0 재측정 (트리 `9ba1e308d`, C안 확정 시점)** — clause 실패 `68`(은퇴 4 포함) / anchor 실패 `17`(은퇴 0)(출처: `progress.md` §F 인용 — 총계는 같은 트리의 `analyze.py`, 은퇴 분리는 `retired-vs-ac.py`(입력 `analysis-postmerge.json`)) / `diff -q` 미러 동일 / `go test ./internal/constitution/...` → `ok  github.com/modu-ai/moai-adk/internal/constitution  1.724s` rc=0(본 커밋 시점 실행). 이 시점부터 clause 집계 기준은 **비은퇴 97** 이다(`spec.md` §1.2 v0.5.0).

## §D 구속 조건 (재논의 금지)

- **D1 — 매처는 불변.** `validator.go` 의 `normalizeWhitespace` / `stripCodeFences` / `Validate` DRIFT 블록에 손대지 않는다. 통과시키는 방향은 오직 데이터 쪽이다.
- **D2 — 엔트리 집합 불변.** 101 → 101. 삭제·추가·번호 재배치 없음. `zone` / `zone_class` / `canary_gate` 값 변경 없음.
- **D3 — clause 는 단일 행 verbatim 구간으로 고른다.** 여러 행에 걸친 인용은 금지 — 정규화 없는 독립 체크(`grep -F`)를 통과해야 하고, 그것이 이 SPEC 의 검증 독립성을 지탱한다.
- **D4 — 교리가 이사했으면 `file:`/`anchor:` 를 옮긴다.** clause 를 요약으로 다시 쓰는 방향은 금지(그게 애초의 원인이다).
- **D5 — 규칙 문서 본문은 고치지 않는다.** 인용을 만들려고 문서를 편집하는 것은 범위 밖이며, 그 순간 이 카드는 교리 변경 카드가 된다.
- **D6 — 가드는 차단 경로에 둔다.** `continue-on-error: true` job 안의 검증은 가드로 인정하지 않는다.
- **D7 — 우회 금지.** clause 길이 임계, 파일 제외 목록, fuzzy/토큰 중첩, 새 환경변수 bypass 를 도입하지 않는다. 기존 `MOAI_CONSTITUTION_SKIP_VALIDATE` 도 가드가 성공으로 읽지 않는다.
- **D8 — slug 규칙은 `spec.md` §2.2 의 6단계로 고정.** 가드 코드에 그대로 적고, 그 규칙 아래에서 anchor 실패 17건이 측정됐음을 주석으로 남긴다(그 선언 의무가 REQ-ZRR-012). **규칙 변경은 구현 세부가 아니라 요구사항 변경으로 다룬다** — 이 프로세스 규율은 요구사항이 아니라 이 구속 조건이 소유하며, 어떤 AC 도 이것을 판정하지 않는다.
- **D9 — `pipeline.go` 스텁을 쓰지 않는다.** `updateSourceFile` / `updateRegistryClause`(`internal/constitution/pipeline.go:256-267`)는 `not yet implemented` 를 반환하는 스텁이다. 수리는 템플릿 레지스트리 + 로컬 미러 **직접 편집**이며, 저 스텁을 구현하거나 호출하는 우회로를 만들지 않는다.
- **D10 — 가드는 붉은 것을 본 뒤에 초록을 인정한다.** 통과만 관측된 가드는 자기가 무엇을 막는지 증명하지 못한다. 일부러 깨뜨린 입력에서 실패하는 것을 **실제 출력으로** 확인한 뒤에야 통과가 근거가 된다(AC-ZRR-007).
- **D12 — 재지정된 `file:` 은 CI paths-filter 안에 머문다.** `detect` job 의 `go_code` 필터(`.github/workflows/ci.yml`)는 `.claude/rules/moai/**` · `CLAUDE.md` · `internal/template/templates/**` · `.moai/**` · `.claude/agents/**` · `.claude/output-styles/moai/**` 를 덮지만 **`.claude/skills/**` 는 덮지 않는다.** 오늘의 17개 파일은 전부 필터 안이지만, 수리가 엔트리를 필터 밖 파일로 옮기면 그 파일 편집에 대해 가드가 조용히 안 돌고 skip-marker 스텁이 초록을 보고한다(`ci.yml:48-58` 주석이 기록한 #1557 실패 형태). 재지정 대상은 필터가 덮는 경로여야 하며, 밖으로 나가야 하면 **같은 변경에서 필터를 확장한다**.
- **D13 — 자기참조 `file:` 금지.** 어떤 엔트리의 `file:` 도 레지스트리 자신을 가리킬 수 없다. 자기참조는 clause 를 정의상 적중시키고 anchor 도 해석시켜, 수리 없이 BLOCKING AC 4건을 한 번에 통과시킨다.
- **D11 — 숫자를 목표로 쓰지 않는다.** 67·17·25 는 `294b4b6ab` 의 측정치이고 규칙 문서가 움직이면 변한다. AC의 GREEN 조건은 전부 "0" 또는 "101/101" 이며, 측정치는 RED 근거로만 인용하고 항상 트리 SHA 를 붙인다.

## §E 자가 검증

각 마일스톤 종료 시 §C 표의 해당 행을 재측정하고 **실제 출력을 그대로** `progress.md` §E.2 에 인용한다. 요약 문장("전부 통과")은 증거가 아니다.

## §F 마일스톤

### 순서 구속 — 권고가 아니라 의존성

[HARD] **anchor 탐지 배선과 anchor 수리는 함께 착지한다.** 나눌 경우 **수리가 먼저**이고 배선이 나중이다. 배선이 먼저 들어가면 17건이 깨진 채로 새 가드가 즉시 붉어지고, 그 붉은 상태가 정상 신호와 구분되지 않아 며칠 안에 "원래 붉은 것"으로 학습된다 — 가드가 태어나자마자 무시당하는 경로다.

같은 이유로 clause 수리도 가드 착지보다 앞선다. 결과적으로 마일스톤 순서는 **M1 수리 → M2 가드 → M3 검증**이며, 이는 되돌리기 어려움 순서가 아니라 **의존성**이다.

대신 이 SPEC 에서 가장 바뀔 가능성이 큰 결정 — 가드 표면 — 은 M2 로 미루지 않고 **§D 구속 조건과 아래 M2 설계표로 지금 확정**한다. 결정을 앞에 두고 착지를 뒤에 두는 형태다.

M1 이 붉은 신호 없이 진행되는 것을 막기 위해, M1 은 §C 사전 점검 표를 반복 측정해 남은 실패 ID 를 줄여 나가는 방식으로 진행한다(측정 스크립트는 `.moai/reports/t232/` 의 기존 분석기를 그대로 쓴다 — 가드가 아직 없어도 RED 는 관측 가능하다).

### M1 — 데이터 수리 (clause + anchor 함께)

- **clause 재동기화**: 각 엔트리에 대해 `file:` 이 지목한 문서에서 **그 교리를 담은 한 줄**을 찾아 그대로 옮긴다. 원문이 그 파일에 없으면 D4 에 따라 `file:`/`anchor:` 를 옮긴다.
- **anchor 수리**: §D8 의 slug 규칙으로 해석되지 않는 17건을 현행 heading 에 맞춘다. 여기에는 clause 가 이미 통과하는 9건이 포함된다 — 이들은 어떤 기존 신호에도 잡히지 않으므로 **명시적으로 목록을 들고 처리한다**.
- 알려진 이사: `CONST-V3R2-008..011` — `CLAUDE.md` §1 → `moai-constitution.md` 의 `## Response Language` / `## Parallel Execution` / `## Output Format`.
- 알려진 heading 개명 7종(`spec.md` §2.2 표) — `ci-autofix-protocol.md` 의 10 엔트리 anchor 를 현행 heading 으로.
- 이중 트리 3건(`CONST-V3R2-004`, `CONST-V3R2-005`, `CONST-V3R5-038`)은 두 판본 공통 구간에서 고른다. **공통 구간 존재는 이미 확인됐다** — 발산 총 4줄, 전부 인용 대상 절 밖(`spec.md` §5). 그럼에도 공통 구간이 없는 엔트리가 나오면 **blocker 보고**(우회 금지).
- 수리는 템플릿 파일 + 로컬 미러 직접 편집이다(D9). 자기참조 `file:` 금지(D13), 재지정은 paths-filter 안에서(D12).
- **유일 적중 실패 시 사다리** (AC-ZRR-002/003 은 각 clause 가 자기 파일에서 **정확히 1회** 적중할 것을 요구한다):
  1. 우선 **더 긴 구간**을 고른다 — 2회 이상 적중은 인용이 짧다는 신호이므로, 같은 교리 문장을 더 넓게 잡으면 대개 유일해진다.
  2. 그래도 유일해지지 않으면 **최소 길이 하한 20자**를 적용해 다시 고른다(현 clause 길이 중앙값 93자, 20자 미만은 3건뿐이므로 이 하한은 사실상 짧은 라벨만 걸러낸다).
  3. 두 단계로도 유일 적중이 불가능한 엔트리가 남으면, 그 사실과 시도한 구간을 `progress.md` §E.2 에 적는다 — 조용히 짧은 clause 로 되돌아가지 않는다.

**M1 종료 조건**: 리터럴 체크 **97/97 비은퇴** (두 트리 각각; 은퇴 4건 clause 는 `[SUPERSEDED …]` 감사 기록으로 불변·면제 — `spec.md` §1.2 v0.5.0), anchor 해석 101/101, 엔트리 집합 diff 0. 전부 기존 분석 스크립트로 측정하며, 이 시점에는 아직 가드가 없다.

### M2 — 가드 착지

이 마일스톤이 끝나면 **가드가 존재하고, 깨끗한 트리에서 초록이며, 일부러 깨뜨린 입력에서 붉은 것이 관측됐다**.

권고 설계(대안은 아래 표):

- `internal/constitution/registry_sync_test.go` (신규 Go 테스트)
  - 두 미러 각각에 대해 `Validate` 를 호출: (로컬 레지스트리, ProjectDir=리포 루트) / (템플릿 레지스트리, ProjectDir=`internal/template/templates`)
  - `result.Skipped == true` 면 **실패**시킨다 (D7 / AC-ZRR-010)
  - `result.DriftCount != 0` 이면 실패, 실패한 엔트리 ID를 전부 출력
  - 같은 테스트 안에서 **anchor 해석**을 검사한다 — 각 엔트리의 `file:` 을 읽어 §D8 의 6단계 규칙으로 heading slug 집합을 만들고 `anchor:` 를 대조. slug 규칙은 코드에 그대로 적고, "이 규칙 아래에서 착지 시점 17건이 실패했다"를 주석으로 남긴다(REQ-ZRR-012). 검증기 코드는 건드리지 않으므로 D1 유지
  - 같은 테스트 안에서 **리터럴 체크**를 검사한다 — 정규화 없이 엔트리별 **라인 수 의미론**(`grep -F -c` 등가 — clause 를 리터럴 부분문자열로 포함하는 **행**의 개수. 같은 행에 2회 적중해도 1로 센다)으로 세어 1회 적중 / 0회 / 2회 이상 / 은퇴 면제 4건의 버킷을 출력한다(빈 clause 도 0회 버킷에서 실패). 이것이 AC-ZRR-002/003 의 기계화이며 검증기보다 엄격하다 — 판정은 boolean 이 아니라 횟수다
  - **[정정 — SPEC-ZONE-REGISTRY-HARDEN-001 (2026-08-25)]** 위 항목의 원문은 리터럴 체크를 `strings.Count`(발생 의미론: 파일 전체 문자열에서의 등장 횟수)으로 기술했으나, 착지된 구현(`internal/constitution/registry_sync_test.go` 의 `literalHitCount`)과 acceptance 판정 규격은 `grep -F -c` **라인 수 의미론**을 잰다(t232 sync-audit verdict F3). 현 데이터는 양쪽 의미론에서 모두 once=97(감사 독립 측정)이라 그 판정은 불변이었다. 본 정정은 문서를 구현·acceptance 의 실측 의미론으로 정렬한다
- CI 배선: 이 테스트는 이미 차단 경로인 `go test ./...` job 에서 돌아간다 → AC-ZRR-008 충족
- 추가로 `.github/workflows/ci.yml` 의 `constitution-check` job 에 `constitution validate` 스텝을 넣되, **그 job 은 `continue-on-error: true` 이므로 보조 신호로만 취급**한다(가드 판정은 Go 테스트가 진다). 이 사실을 스텝 주석에 적는다.

| 대안 | 채택 | 사유 |
|---|---|---|
| Go 테스트(위) | ○ | 이미 차단되는 job 에서 돌고, 두 미러를 한 번에 보며, 검증기 코드를 안 건드린다 |
| `constitution-check` job 에 validate 추가만 | × | job 이 `continue-on-error: true` — 실패해도 PR 이 안 막힌다 (AC-ZRR-008 불충족) |
| `validator.go` 에 `SentinelAnchorNotFound` 배선 | 보류 | 죽은 sentinel 을 정식으로 살리는 방향이며 D1 의 문자를 어기진 않는다(가산적). 다만 검증기 표면을 넓히고 AC-ZRR-005 의 판정 범위를 늘리므로 이번엔 하지 않는다 — 테스트 쪽 해석기로 같은 커버리지를 얻는다. **후속 후보로 명시 기록** |
| pre-commit 훅 | × | 로컬 전용이라 PR 을 막지 못한다 |

**M2 종료 조건 (붉은 것을 먼저 본다 — D10)**:

1. 깨끗한 트리에서 가드 통과를 확인한다
2. 명시 엔트리 `CONST-V3R2-004` 와 무작위 **비은퇴** 1건의 `clause:` 를 각각 일부러 깨뜨려 가드를 돌려 실패를 관측한다 — 실제 출력(종료 코드 + 실패 엔트리 ID)을 `progress.md` §E.2 에 그대로 인용한다
3. anchor 로도 같은 변이를 1회 수행한다
4. 템플릿 미러 쪽으로도 같은 변이를 1회 수행한다
5. `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 을 건 상태에서 변이 주입 가드가 **여전히 실패**하는지 확인한다 — 그리고 **깨끗한 트리(변이 없음)** 에서도 `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 을 건 채 가드를 돌려 "검증 건너뜀" 을 이유로 실패하는지 확인한다(AC-ZRR-010, plan-audit iter3 C5)
6. 변이를 전부 되돌리고 다시 통과를 확인한다
7. 변이 주입 CI 관측 — R1 스크래치 커밋을 PR 에 올려 가드 job 의 결론이 `fail` 인 것을 `gh pr checks` 로 인용한다 (`.moai/reports/t232/guard-failure-scenario.md` §3)

변이 판정의 정본은 `guard-failure-scenario.md` §1–§3 이다 (AC-ZRR-007 준거).

### M3 — 미러·임베드·최종 검증

- 로컬/템플릿 바이트 동일 확인 → `make build` → `git status --porcelain` 클린
- 템플릿 중립성 3종 grep 0 확인
- 새 바이너리로 스크래치 프로젝트 `moai init` → `moai constitution validate` exit 0 / `moai doctor` Fail 0
- `go test ./internal/constitution/... ./internal/template/...` 통과
- (변이 주입 검증은 M2 종료 조건에서 이미 수행 — 여기서는 최종 트리에서 가드 통과만 재확인)

## §G AC별 mutant 노트

각 줄은 "그 AC를 **통과하면서** 그 AC가 지키려는 요구사항을 **위반하는** 구체적 구현"이다. 이 노트가 비면 그 AC는 약한 것이다.

| AC | 이 AC만 보면 통과하는 mutant | 그 mutant 를 잡는 다른 AC |
|---|---|---|
| **AC-002/003/004 합동 (자기참조 `file:`)** | 깨진 엔트리의 `file:` 을 레지스트리 자신으로 재지정한다. clause 는 정의상 그 파일에 있고(측정: `grep -F -c` → 1), heading 50개가 anchor 를 해석시키며, 두 미러가 바이트 동일하니 템플릿 AC 까지 통과 — **수리 0줄로 BLOCKING 4건 통과**. AC-006 은 `file:` 변경을 명시적으로 허용하므로 못 잡는다 | 002/003 의 자기참조 금지 조항 + `file:` 변경 목록 인용 → sync 리뷰 판정 (D13) |
| **AC-002/003 (빈 clause)** | 모든 clause 를 빈 문자열로 만든다. 빈 문자열은 모든 파일에 적중하고(`grep -F -q -- ''` → rc 0), `Validate` 는 `normalizedClause != ""` 로 건너뛰어 AC-001 까지 통과 | 002/003 의 "빈 clause 는 실패로 센다" 조항 |
| **AC-007/008 합동 (`\|\| true`)** | 가드를 차단 job 안에 넣되 `go test ./... \|\| true` 로 감싼다. job 단위 `continue-on-error` 가 없으니 008 을 문자 그대로 만족하고, 007 이 보는 것은 **로컬** 종료 코드라 로컬은 정상적으로 붉어져 007 도 통과 — REQ-007 을 검증하는 AC 가 하나도 남지 않는다 | 008 의 스텝 단위 억제 금지 + 007 의 CI job 결론(`gh pr checks`) 관측 |
| **AC-007 (부분 순회)** | 가드에 제외 목록·조기 반환을 넣어 일부만 검사한다. 변이가 제외 밖 엔트리에 떨어지면 가드는 정상적으로 붉어져 통과 | 007 의 "평가 엔트리 수 = clause 97 / anchor 101(두 미러 각각 분리 보고)" 단언 + 변이 대상 고정(명시 ID 1 + 무작위 비은퇴 1) |
| AC-ZRR-001 (validate exit 0) | 매처를 토큰 중첩으로 바꾼다. 또는 깨진 67 엔트리를 삭제한다. 또는 `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 을 CI/문서 기본값으로 만든다 | 005 (매처 불변), 006 (엔트리 보존), 010 (skip 불허) |
| AC-ZRR-002 (로컬 리터럴 live 97/97) | 로컬 규칙 문서에 레지스트리 문장을 그대로 **붙여 넣어** 인용을 만든다(문서 오염). 또는 로컬만 고치고 템플릿은 방치 | D5 위반 → 003 (템플릿 트리에서도 성립해야 함), 011 (미러 동일성) |
| AC-ZRR-003 (템플릿 리터럴 live 97/97) | 템플릿 규칙 문서를 편집해 인용을 만든다 | 사람 판정: sync-phase 리뷰가 "규칙 본문 diff 0" 을 확인 (§D5). 기계 판정 아님 — Gap 으로 기록 |
| AC-ZRR-004 (anchor 101/101) | anchor 필드를 전부 `#` 같은 항상-해석되는 값으로 치환한다 | 006 (필드 변경 범위 제한) + 014 (slug 규칙 고정) + sync 리뷰. **기계적으로 완전히 막히지 않음** — §H 잔여 위험 |
| AC-ZRR-007 (변이에서 붉은 것을 관측) | 가드를 통과 상태에서만 돌려 보고 "무는 것을 확인했다"고 적는다 | 이 AC 자체가 **실제 실패 출력 인용**을 요구한다(D10). 인용이 없으면 미충족 |
| AC-ZRR-014 (slug 규칙 선언) | 규칙을 코드에 적되 실제 해석기는 다른 규칙을 쓴다 | 004 (그 해석기로 101/101 이 나와야 함) + 007 (anchor 변이에서 실패해야 함) |
| AC-ZRR-005 (매처 불변) | 매처는 그대로 두고 **가드 쪽**에 제외 목록을 넣는다 | 007 의 평가 엔트리 수 단언(clause 97 / anchor 101, 두 미러 각각) — 변이 위치에 의존하지 않는 유일한 방어. 변이 주입만으로는 제외 밖에 떨어질 때 통과한다(위 합동 행) |
| AC-ZRR-006 (엔트리 보존) | 엔트리는 다 남기되 clause 를 전부 빈 문자열로 만든다 — `Validate` 는 빈 clause 를 건너뛴다(`normalizedClause != ""`) | 002/003 의 "빈 clause 는 적중이 아니라 실패로 센다" 조항. **이 방어는 plan 메모가 아니라 AC 본문에 있다** — 식별된 mutant 를 AC 로 승격하지 않으면 방어가 아니다 |
| AC-ZRR-007 (변이 시 실패) | 가드가 clause 변이만 잡고 anchor 변이는 못 잡는다 | 004 + M3 의 변이 주입을 anchor 에도 1회 수행 |
| AC-ZRR-008 (차단 경로) | 가드를 차단 job 에 두되 `|| true` 로 감싼다 | 008 자체가 스텝 단위 억제를 금지하고, 007 이 **CI job 결론**을 관측한다(로컬 종료 코드로는 구분 불가 — 위 합동 행 참조) |
| AC-ZRR-009 (템플릿 커버) | 템플릿 레지스트리를 로컬 것과 동일하다는 이유로 검증을 생략(파일 비교만) | 템플릿 **소스 파일** 3개가 로컬과 다르므로(§5) 비교만으로는 부족 — 변이 주입을 템플릿 소스 쪽에 1회 수행 |
| AC-ZRR-010 (skip 불허) | 가드가 `Skipped` 를 검사하되, 그 검사를 자기 자신도 skip 하는 조건에 넣는다 | 007 의 변이 주입을 skip 환경변수와 **함께** 수행(M3 마지막 항목) |
| AC-ZRR-011 (미러 동일) | 두 파일을 동일하게 유지하되 둘 다 틀린 채로 둔다 | 002/003 |
| AC-ZRR-012 (중립성) | 아무것도 안 한다(이미 0) | 비회귀 AC임을 명시. 채택 근거는 변이(SPEC ID 1개 주입 → 중립성 CI 실패) |
| AC-ZRR-013 (doctor Fail 0) | doctor 의 constitution 체크를 warn 으로 격하한다 | 001 (CLI 직접 판정), 005 |

## §H 잔여 위험

- **의미 정합성은 기계로 못 잡는다.** "verbatim 이고 anchor 가 해석된다"가 "맞는 문장을 골랐다"를 함의하지 않는다. anchor 를 항상-해석되는 값으로 치환하는 mutant(AC-ZRR-004 행)는 기계 판정을 빠져나간다. 이 축은 sync-phase 사람 리뷰가 진다 — 리뷰어에게 "레지스트리 diff 전체를 읽고, 각 clause 가 그 anchor 절 안에 있는지" 를 명시적으로 요청할 것.
- **근접 오답 재지목 1건 — `CONST-V3R2-004`.** 깨진 clause 68건 중 **정확히 이 하나**가 다른 이중 트리 파일 `.claude/rules/moai/NOTICE.md` 안에 리터럴로 존재한다(clause 값 `16-language neutrality`). 이 엔트리를 `NOTICE.md` 로 옮기면 AC-ZRR-002/003 을 **기계적으로 통과한다**. 방어는 새 `file:` 변경 목록이며, **이 엔트리가 `NOTICE.md` 로 이동한 채 그 목록에 나타나면 sync 리뷰어는 거부한다** — 이 교리의 서식지는 `coding-standards.md` `#language-policy` 다. 위 항목의 "레지스트리 diff 전수 읽기" 요청에 **이 엔트리를 이름으로 붙여** 전달할 것. 판정자: sync-phase 사람 리뷰어(기계 판정 없음).
- **"평가 엔트리 수 101" 은 로드 수와 검사 수를 구분하지 못한다.** 레지스트리를 101건 파싱하기만 해도 그 수를 보고할 수 있다. 따라서 manager-develop 은 카운터를 **엔트리별 clause 검사 완료 시점과 anchor 검사 완료 시점에 각각** 증가시키고, §E.2 에 **두 개의 수**(clause 검사 97 / anchor 검사 101)를 따로 인용한다. sync-auditor 는 인용된 출력에 수가 하나가 아니라 **둘**인지 확인한다. 판정자: sync-auditor.
- **규칙 본문 오염 경로.** D5 를 어기고 문서를 고쳐 인용을 만드는 우회는 AC 로 완전히 막히지 않는다. 방어는 PR diff 관찰(규칙 문서 변경 라인 0)이며, 이는 사람/리뷰 판정이다.
- **가드의 anchor 해석기는 재구현이다.** slug 규칙이 렌더러와 다르면 오탐/누락이 생긴다. 17이라는 RED 수치도 이 재구현 기준이다.
- **스냅샷 수리라는 성격.** 이 SPEC 이 고치는 것은 오늘의 101건이고, 가드가 지키는 것은 내일의 편집이다. 가드가 없으면 이 수리는 몇 달 뒤 같은 상태로 돌아온다 — 그래서 M1 이 첫 마일스톤이다.

## §I 교차 참조

- `.moai/reports/t232/findings.md` — 측정 근거 원본
- `internal/constitution/validator.go` — 불변 대상
- `internal/cli/update/deploy/deploy.go` `CleanMoaiManagedPaths` — 왜 템플릿에서 고쳐야 하는지
- `.github/workflows/ci.yml:445-475` — 기존 constitution job (advisory)
- 후속 후보: 프로젝트 로컬 오버레이 기구(#1616 2안), `validator.go` 에 anchor sentinel 정식 추가, 카드 t201(#1595 `[SUPERSEDED]` / `canary_gate` 축)
