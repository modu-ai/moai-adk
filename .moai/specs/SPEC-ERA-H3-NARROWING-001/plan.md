# SPEC-ERA-H3-NARROWING-001 — 구현 계획

Tier S (spec.md + plan.md; AC는 spec.md §3에 인라인). 마일스톤은 **되돌리기 어려운 결정을 앞에** 둔다 — 술어 정의가 먼저, 기계적 문서 갱신이 마지막이다.

## §A 맥락

`spec.md` §1(배경·무게중심)·§3(AC)·§5(설계 선택)이 정본이다. 기준선 측정은 `.moai/reports/t382/measurements-9328a5242.md`(M1~M13). 여기서는 반복하지 않는다.

**Tier S 판정 근거.** `spec-workflow.md` § SPEC Complexity Tier의 기준은 LOC와 파일 수다 — S는 `< 300 LOC` 이고 `< 5 files`. 이 변경은 `era.go` 약 15 LOC, `era_test.go` 약 120 LOC, 문서 몇 줄, **총 3 파일**로 양쪽 기준을 모두 만족한다. REQ 6개 · AC 8개도 Tier S 상한(각 8)에 들어간다. 초안은 Tier M으로 잡았는데, 그 근거가 「AC 집합이 크다」였다 — Tier 표가 키로 삼지 않는 축이고, **측정에 들인 노력에 맞춰 Tier를 올린 것**이었다. 정정한다.

## §B 알려진 함정

1. **`§E.2`는 sync 신호가 아니다.** `hasSyncSection`은 역사적 오칭이며 코드 주석이 이를 명시한다. sync 단계는 `§E.4`다. 술어를 고칠 때 변수명에 이끌려 의미를 뒤집지 않는다.
2. **첫 매치 승리다.** H-3에 조건을 **추가**하는 것과 순서를 바꾸는 것은 전혀 다른 변경이다. 이 SPEC은 전자만 한다(REQ-EH3-003).
3. **기존 H-3 테스트가 왜 통과하는지 이해하고 넘어간다.** `era_test.go`의 H-3 케이스들은 `FrontmatterCreated`/`FrontmatterPhase`를 비워 두고 있다. 이 공백은 우연이 아니라 REQ-EH3-002가 요구하는 성질의 표현이다. 통과했다고 안심하지 말고, AC-EH3-003·004의 뮤테이션으로 그 통과가 공허하지 않음을 보인다.
4. **`terminalStatusEnum`이 era와 병렬로 강등을 건다.** `demote := isGrandfatheredSpecDir(...) || terminalStatusEnum[status]`. V3R5 23건 중 9건이 terminal이라 시대가 바뀌어도 lint 강등은 그대로다. lint 델타를 잴 때 이 9건을 섞으면 수치가 흐려진다.
5. **`MovingRefUnpinned`와 `StatusGitConsistency`는 발화 지점에서 `Advisory: true`를 세운다.** grandfather 여부와 무관하게 `--strict`가 승격하지 못한다. 이 둘을 lint 델타의 근거로 삼으면 안 된다.
6. **`_archive/` 아래에도 spec.md가 있다.** `grep -rl ... .moai/specs --include=spec.md`와 `.moai/specs/*/spec.md` 글롭은 서로 다른 모집단을 센다(47 vs 46). 어느 쪽을 썼는지 라벨 없이 수를 인용하지 않는다.

## §C 사전 점검 (run-phase 착수 전)

```bash
git rev-parse --show-toplevel      # → .../.claude/worktrees/t382
git rev-parse --short HEAD
git branch --show-current          # → WT-era-plan-phase
git fetch origin develop
git rev-list --count --left-right origin/develop...HEAD
make build                         # rc=0, 이후 모든 측정은 ./bin/moai 로만
```

`make build`를 건너뛰고 PATH 바이너리로 재면 이 트리를 재는 것이 아니다.

## §D 제약

- 수정 파일은 §F에 열거된 3개로 한정한다. `internal/spec/lint.go`·`audit.go`·`drift.go`는 **읽되 고치지 않는다**(REQ-EH3-004).
- 로컬에서 `go test ./...` 금지. `go test ./internal/spec/...`만.
- `.claude/rules/local/lifecycle-sync-gate.md`는 로컬 전용 룰이다(`.claude/rules/local/` 아래 — CLAUDE.local.md Local-Only 목록). **템플릿 미러가 없고 필요하지도 않다.** `grep -rln "H-3" internal/template/templates/`가 내는 유일한 히트는 `moai-domain-humanize/modules/copy-review.md`이며 무관하다.
- 결함 주입(AC-EH3-006)은 반드시 임시 사본에서. 실제 SPEC 파일을 훼손하지 않는다.

## §E 자체 검증

각 마일스톤 종료 시 판정 명령과 축자 출력을 `.moai/reports/t382/` 아래에 남긴다. 요약은 증거가 아니다.

## §F 마일스톤

### M1 — 술어 정의와 H-3 게이트 (되돌리기 가장 어려움)

**파일: `internal/spec/era.go`**

| 위치 | 변경 |
|---|---|
| `matchesModernPhase` / `isAfterModernThreshold` 인근 | 새 헬퍼 `hasModernEraSignal(signals EraSignals) bool` 추가 — 본문은 `matchesModernPhase(signals.FrontmatterPhase) \|\| isAfterModernThreshold(signals.FrontmatterCreated)`. H-5의 조건절을 이 헬퍼 호출로 치환해 **두 지점이 같은 술어를 쓴다는 사실을 코드로 못박는다**(AC-EH3-008의 불변식이 이 동일성 위에 선다) |
| H-3 조건절 (`era.go:150` 부근) | `if hasSyncSection && syncSHA == ""` → `if hasSyncSection && syncSHA == "" && !hasModernEraSignal(signals)` |
| H-3 앞 주석 (`era.go:147-149` 부근) | 유예 조건과 그 이유(plan-phase 골격이 `§E.2`를 항상 찍으므로 `§E.2` 유무는 시대 신호가 아니다)를 적는다 |
| 함수 헤더 휴리스틱 순서 목록 (`era.go:96` 부근) | H-3 줄에 「단, modern-era 신호가 없을 때」를 반영 |

**금지:** 휴리스틱 재배치, `modernEraThreshold` 변경, `EraFinal()`/`IsModern()` 수정.

**판정:** `go test ./internal/spec/...` — 기존 테스트 전부 통과(REQ-EH3-002의 1차 증거).

### M2 — 단위 기준과 회귀 가드

**파일: `internal/spec/era_test.go`**

- `TestClassifyEra` 테이블에 AC-EH3-001 ~ AC-EH3-004 네 서브테스트를 추가한다. 기존 테이블 관행(`[]struct{ name string; signals EraSignals; wantEra Era; wantRule string }`)을 그대로 따른다.
- `TestClassifyEra_NoV3R5WhileModernSignal` 불변식 테스트를 추가한다(AC-EH3-008). 조합 순회는 spec.md §3.5에 열거된 4축 곱집합.
- 각 새 서브테스트 위에 **어떤 뮤테이션이 그것을 깨뜨려야 하는지**를 주석으로 적는다 — `lint_movingref_test.go`·`lint_artifact_status_test.go`의 기존 관행.

**판정:** `go test -run TestClassifyEra ./internal/spec/` 통과. 이어서 뮤테이션 3종(H-3 무조건 스킵 / 유예절 제거 / 헬퍼를 날짜만으로 축소)을 하나씩 심어 각각 어떤 테스트가 깨지는지 관측하고 되돌린다. 관측 출력을 남긴다.

### M3 — 코퍼스 귀속과 비용 측정

코드 변경 없음. 측정과 귀속만.

- `make build` 후 `./bin/moai spec audit --json`으로 SPEC별 era를 뽑아 기준선과 대조 → AC-EH3-005 (총계 + 22건 원소 일치 + 건별 근거 + 표본 5건).
- `./bin/moai spec drift`를 돌려 `drift-before-9328a5242.txt`와 23행 대조 → AC-EH3-007. **이 카드의 무게중심 축이므로 M3에서 가장 먼저 잰다**(spec.md §1.2 축 1). 새 `DRIFT` 행이 있으면 **건별로** 진짜/오탐 판정.
- 결함 주입 실험 → AC-EH3-006. 수정 전 바이너리와 수정 후 바이너리 양쪽으로 각각 돌린다(수정 전 바이너리는 M1 착수 전에 `bin/moai-pre-t382`로 보존해 둔다).

**오탐이 1건이라도 나오면 M1으로 돌아간다.**

### M4 — SSOT 문서 갱신 (기계적, 마지막)

**파일: `.claude/rules/local/lifecycle-sync-gate.md`**

- § Era Classification Heuristic의 H-3 행을 새 술어로 갱신: 「`§E.2` present, `sync_commit_sha` 부재 **또한 modern-era 신호(`phase` v3.0/v3R6 계열 또는 `created >= 2026-04-01`) 부재**」.
- H-5 행 옆에, H-3의 유예 조건이 H-5의 술어와 **동일**하다는 불변식을 한 줄로 명시한다.
- § Grandfather Clause Policy에 「신호가 없으면 재분류하지 않는다」는 보수적 기본값과 그 근거(H-6 낙하 회피)를 덧붙인다.

**판정:** `grep -n "H-3" .claude/rules/local/lifecycle-sync-gate.md`으로 갱신 확인. 템플릿 미러 없음 — §D 제약 참조.

## §G 안티패턴

- **총계만 맞추고 귀속을 생략하기.** `Grandfathered: 263`이 나왔다는 사실은 **어느** 22건이 움직였는지 말해 주지 않는다. AC-EH3-005는 원소 단위 일치를 요구한다.
- **뮤테이션 없이 "가드가 통과했다"고 보고하기.** 셀렉터가 0건을 매치해도 초록이다. AC-EH3-003·004·008은 실패 관측을 본체로 갖는다.
- **이 수정을 "게이트가 실패를 통과시키고 있다"로 소개하기.** 측정이 그 주장을 반증한다 — lint rc 델타는 오늘 0이고 강등된 ERROR를 실제로 가진 SPEC은 1건뿐이다(spec.md §1.2 축 3). 이것은 **분류의 정확성 결함**이고, 오늘 실제로 무는 곳은 drift 면제다(축 1, 22건). 반대로 "그러니 별것 아니다"로 흘리는 것도 틀렸다 — 면제 집합이 새 SPEC마다 자란다(축 1 + §1.3).
- **`created_at:` 별칭을 era 엔진이 읽게 만들기.** spec.md §5 옵션 D에서 근거를 적어 기각했다.
- **INIT-WIZARD의 frontmatter를 "지나는 김에" 고치기.** 범위 밖(spec.md §4).

## §H 상호 참조

- `internal/spec/era.go` — 수정 대상
- `internal/spec/lint.go` `applyEraDemotion` / `eraDemotableCodes` / `terminalStatusEnum` — 읽기 전용, 결과 해석의 근거
- `internal/spec/audit.go` `auditSpec` / `internal/spec/drift.go` ④ — 읽기 전용, `unclassified` 비용 판정의 근거
- `.claude/rules/local/lifecycle-sync-gate.md` — SSOT 문서
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Rejected Snake_Case Aliases — INIT-WIZARD 자기차폐의 기제
- 형제 카드 t371 / t376 / t380 — spec.md §4·§6
