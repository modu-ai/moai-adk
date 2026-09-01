# Progress: SPEC-SPECLINT-GITBLIND-001

## §E.1 Plan-phase Audit-Ready Signal

### iter-3 (v0.4.0) — develop 흡수 후 인용 재측정 · 린트 기준값 재귀속

측정 트리: 이 워크트리 HEAD **`35bc0715f`**. 흡수: `git merge origin/develop`(`9328a5242`), 91커밋, 충돌 0.
판정 도구는 트리에서 빌드한 `./bin/moai`(PATH 바이너리 미사용). 원문: `.moai/reports/t371/remeasure-35bc0715f.md`.

- **인용 좌표 전량 재측정.** 이동은 `internal/spec/lint.go` 한 파일에 국한된다(총 행수 1323 → 1342).
  `StatusGitConsistencyRule.Check` `:1287-1323` → **`:1306-1342`**, 조용한 skip 블록 `:1305-1308` → **`:1324-1327`**,
  `Advisory: true` `:1316` → **`:1335`**, `terminalStatusEnum` 조기 반환 `:1299-1301` → **`:1318-1320`**,
  `applyEraDemotion` `:284-300` → **`:296-312`**. `lint.go:33-45` · `lint.go:61` 은 불변이고,
  `lint_ownership.go` · `drift.go` · `gitquery_cache.go` · `cli/spec_lint.go` · `drift_characterization_test.go` ·
  `spec-lint.yml` · `ci.yml` 인용은 흡수 전후 동일하다.
- **린트 기준값 = `0 error / 1096 warning`**(`35bc0715f`). **종전에 돌던 "1098"은 warning 수가 아니라
  보고서 파일의 `wc -l` 값이었다** — 단위가 달라 비교 자체가 성립하지 않았다. 흡수 전 트리(`1e5199b88`)가
  스스로 선언한 총계는 `2 error / 1091 warning` 이다.
- **차분 귀속.** rule별 대조에서 움직인 것은 `SyncSHASlotFormat` 0 → **5**(+5, 흡수와 함께 도착한 새 규칙)
  하나뿐이고 나머지 9종은 전부 0 델타다. 검산: 1091 + 5 = 1096. 사라진 error 2건은 둘 다
  `ArtifactStatusFieldForbidden`, 대상은 `SPEC-INTEGRATION-LOCK-ATOMIC-001` 의 `plan.md` / `acceptance.md` —
  **이 카드 소관이 아니며** develop 쪽에서 이미 수리됐다.
- **18건은 개수가 아니라 집합으로 동일하다.** `StatusGitConsistency` 는 18 그대로이고,
  SPEC ID 정렬 목록의 diff 가 비었다. 따라서 `.moai/reports/t371/classification-18.md` 의 분류는 흡수 후에도 유효하다.
- **전제 반증 — t382 는 이 18건을 움직이지 못한다.** `StatusGitConsistency` 는 발화 지점에서 **무조건**
  `Advisory: true` 로 나오고(`internal/spec/lint.go:1335`), `applyEraDemotion`(`:296-312`)의 warning 분기는
  이미 true 인 플래그를 다시 true 로 **설정**할 뿐이며, `eraDemotableCodes`(`:272-275`)에
  `StatusGitConsistency` 는 없다. era 분류는 발화 **이후**에 적용되므로 개수도 억제하지 못한다.
  즉 t382(era.go H-3)가 어느 방향으로 착지하든 이 카드의 18건은 개수도 advisory 여부도 변하지 않으며,
  **병합 순서 제약의 근거로 쓸 수 없다.** 같은 논리가 이 카드가 신설할 `StatusGitUnreachable`(Info)에도 걸린다 —
  `applyEraDemotion` 의 switch 는 Error 와 Warning 만 다루고 Info 는 통과시킨다.
- **잔여 위험.** `internal/spec/lint.go` 를 세 카드가 동시에 만진다(이 카드 M1 ≈ `:1324`, t382 `:272-275`,
  t376 rule 등록부 `:137`). 텍스트 충돌 가능성은 낮으나 **어느 쪽이 먼저 착지하든 나머지의 인용 행번호가 다시 밀린다.**

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

## §F Phase 4 Mode Selection

- 입력 파라미터: tier M · scope 파일 수 ~5(internal/spec 3-4파일 + .github/workflows/spec-lint.yml + 테스트) · 도메인 2(Go 런타임 + CI 워크플로 YAML) · 언어 혼합 Go+YAML · concurrency benefit LOW(coding-heavy) · agent-team 사전요구 없음
- 모드 평가: direct 미선정(다중 파일·신규 테스트 seam) / fanout 미선정(research-heavy 아님, Anthropic coding-task caveat) / sweep 미선정(기계적 균일 변환 아님, ~30파일 미만) / **serial 선택**
- Decision: serial
- 정당화: 구현이 단일 도메인(internal/spec)의 인과 체인(M1 관측가능화 → M2 해소 체인 → M3 워크플로)이고 각 마일스톤이 이전 마일스톤의 seam 위에 세워지므로 병렬 이득이 없다. RED 우선 순서(M1의 픽스처 관측이 M2·M3의 기준선)가 직렬 의존을 만든다.
- 킥오프: 운영자 승인 2026-09-02(lead-1 경유) — 자율 진행 + goal 무장. iter-4 plan-audit PASS 0.87(트리 06d908455) 후 진입.
