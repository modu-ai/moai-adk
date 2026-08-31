# Progress: SPEC-SPECLINT-GITBLIND-001

## §E.1 Plan-phase Audit-Ready Signal

### iter-2 정정 (v0.3.0)

- **D12 를 필수로 승격**(감사자는 minor 로 접수, 리드가 상향). `printTable` 의 zero-finding
  short-circuit(`internal/cli/spec_lint.go:115-118`, 메시지 `:116`)이 곧 눈감긴 상태가 실제로
  내보내는 출력 — `✓ No findings — all SPEC documents are valid` — 이라는 사실을 `spec.md` §1.2 서사에 편입.
  관측 표면을 기본 표 출력 하나로 못박고, AC-SLGB-001 / 004 에 **그 줄의 부재** 단언과
  RED 기준선 픽스처 전제(schema-valid → M1 이전에 그 줄이 실제로 찍힘)를 추가.
  이 단언이 없으면 Info 를 `--json` 경로에만 내는 구현이 모든 AC 를 통과한다.
- **M2 단계 0 분기 철회.** `cachedMainBranch` 의 cwd 의존은 시그니처 변경을 요구하지 않는다 —
  `chdirForTest`(`drift_characterization_test.go:55`) + `setupDriftCorpusFixture`(`:98` → `:103`)가
  확립된 in-package 선례다. M2 diff 는 `cachedMainBranch` 본문 + 캐시 필드에 국한, 호출부 미접촉.
  go.mod 는 go 1.26.4 라 `t.Chdir` 도 가용하나, 선례가 `os.Chdir` + `t.Cleanup` 이므로 그대로 따른다.
  D9 비병렬 구속은 유지되며, 기존 헬퍼 주석(`:53`)이 이미 같은 문장을 담고 있다.
- **`--json` / `--sarif` 금지 조항 추가.** 감사자의 `--json` 주장은 `Finding` 구조체 태그
  (`internal/spec/lint.go:33-45`)에만 근거하고 실행 검증이 없다. 어떤 AC 도 그 경로에 기대지 않는다(`spec.md` §4).
- 리드 인용 정정 3건(모두 off-by-one, 나머지는 행 일치): `chdirForTest` 는 `:55-70`(리드 `:55-64`),
  `setupDriftCorpusFixture` 는 `:97-105` 이고 `chdirForTest` 호출은 `:103`(리드 `:98-106`),
  short-circuit 블록은 `:115-118`(리드 `:114-117`).

### iter-2 (v0.2.0)

- Tier: **S → M 상향**. 대상 파일 6개(`internal/spec/lint.go`, `internal/spec/drift.go`,
  `internal/spec/gitquery_cache.go`, `.github/workflows/spec-lint.yml` + 테스트 2), AC 11건.
  Tier S 예산(REQ 8 / AC 8)에 맞추려면 AC 를 합쳐야 하는데 그것은 iter-1 D4(추적성 결함)를
  되살리는 방향이라 상향을 택했다. **결과: PASS 문턱 0.75 → 0.80.**
- 산출물: `spec.md` + `plan.md` + `acceptance.md` (Tier M 3-file set) + 이 `progress.md`.
- REQ 9건 / AC 11건 — Tier M 상한(각 16) 이내.
- iter-1 차단 결함 5건 전부 닫음: D1(error 3종 표 + 모양별 발화) · D2(실행당 1건 결정 + AC-SLGB-003 상한) ·
  D4(AC-SLGB-002 ref 이름 추적) · D7(AC-SLGB-005 비-terminal status [HARD] 제약) ·
  D8(AC-SLGB-005 / 008 mutation 절차).
- iter-1 비차단 4건 반영: D3(§F 의존성 정밀화 — M1→M2 는 의존, M1→M3 는 순서 선호) ·
  D5a(§1.2 인용 `:1310-1313` → `:1305-1308`) · D6(`DetectDrift` 를 spec.md §4 잔여 위험으로 승격,
  덮는 AC 없음을 명시) · D9(캐시 전역성 → 비병렬 [HARD] 제약).
- 범위 추가: REQ-SLGB-009(워크플로 trigger paths) + AC-SLGB-010.
- 사전 측정: develop `b9149857c` A/B 3회 (`spec.md` §1.1). 이 SPEC 은 재조사 없이 그 측정 위에 선다.

**미해소 / 의도된 공백**

- AC-SLGB-011 은 착지 후 CI 로그로만 판정 가능하다 — plan-phase 에서 닫히지 않는다.
- `spec.md` §4 의 `DetectDrift` 동작 변화는 어떤 AC 도 덮지 않는다(의도된 공백, 기록됨).
- M2 단계 0(시그니처 유지 vs 디렉터리 파라미터)은 run-phase 착수 시 결정하고 여기 §E.2 에 기록한다.

### iter-1 (v0.1.0)

- Tier S 로 분류, plan-auditor PASS-WITH-DEBT 0.75(문턱과 동일, 여유 0). 차단 결함 5건.
- 판정문: `.moai/reports/t371/plan-audit-iter-1.md`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
