# SPEC-CODEX-SKILL-DISABLE-001 — 진행 기록

카드: t502 · 브랜치: `WT-codex-skillconfig`

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID 정규식 검사: 실행 `[[ "SPEC-CODEX-SKILL-DISABLE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`
- ID 유일성: `.moai/specs/` 에 동명 디렉터리 없음(생성 전 `ls -d .moai/specs/SPEC-CODEX-*` 로 확인)
- 산출 파일: spec.md · plan.md · acceptance.md · progress.md (Tier M)
- 미해결 `[NEEDS CLARIFICATION]`: **0건**. v0.1.0의 3건(Q-a 경로 표기, Q-b 미러 모드 교차, 후보 루트 집합)은 `.moai/reports/t502/gate-path-shape.md`(19셀, `codex-cli 0.153.4`)로 전부 닫혔다 — 발행 표기는 `<projectRoot>/.agents/skills/<skill>/SKILL.md` 로 확정.
- 선행 증거 2본: `.moai/reports/t504/skills-config-path-shape.md`(`enabled` 필수, 게이트 존재) + `.moai/reports/t502/gate-path-shape.md`(표기 확정, realpath 정규화, skipped 주장 반증).
- 철회 1건: v0.1.0의 「`MirrorModeSkipped` ⇒ 해소 실패」 주장은 측정이 반증해 철회됨(spec.md §A 각주).
- 예산: **REQ 16/16 (상한 안), AC 17 — Tier M 상한 16 초과.** 표제 앵커 `grep -c '^### AC-CSD'` → 17. 이는 **기록된 예외**이며 근거는 plan.md §F.1(원인: D-N2 판정이 세 거절 사례에 서로 다른 종료 코드를 부여해 병합 불가가 됨 / 검토·기각한 병합 후보 명시 / 상한 완화 아님). 조항을 넓혀 수를 맞추지 않았다 — 두 성질을 지는 기준에는 판별 셀도 둘씩 붙였다.
- 남은 [HARD] 게이트: AC-CSD-050(발행 코드 착지 직전, 그 시점 codex 버전에서 2셀 재측정).
- 상태: `draft` — Implementation Kickoff Approval 대기

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t502` @ `WT-codex-skillconfig`. 구현 커밋 `eef6061a6`(부모 `8789e9942`). 아래 모든 판정은 이 트리에서 실행한 명령과 그 출력이다.

### AC 판정 행렬

| AC | 판정 명령 | 관측 출력 | 상태 |
|---|---|---|---|
| AC-CSD-001 | `go test -run TestUpsertCodexSkillDisableAppendsOneEntryWithFalse ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-002 | `go test -run TestResolveCodexSkillMirrorPathPublishesLiteralMirrorPath ./internal/cli/` (symlink·copy 2셀) | `ok` | **PASS** |
| AC-CSD-003 | `bash .moai/reports/t502/e2e-verb.sh copy` · `… symlink` | 두 실행 모두 `verdict=exposed marker=1` → `verdict=gated marker=0`, `expect=… result=MATCH` ×2, `E2E PASS` | **PASS** |
| AC-CSD-010 | `go test -run TestRunCodexSkillDisableDryRunNamesResolvedPath ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-011 | `go test -run TestRunCodexSkillDisableUnresolvedNameFailsWithoutWriting ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-012 | `go test -run TestRunCodexSkillDisableAmbiguousNameFailsWithoutWriting ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-013 | `go test -run TestRunCodexSkillDisableMirrorAbsentExitsZero ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-020 | `go test -run TestUpsertCodexSkillDisableUpdatesExistingEntry ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-021 | `go test -run TestUpsertCodexSkillDisableIdempotentBytes ./internal/cli/` (`bytes.Equal`) | `ok` | **PASS** |
| AC-CSD-022 | `go test -run TestUpsertCodexSkillDisablePreservesSurroundings ./internal/cli/` (3 서브테스트: 무관 엔트리·주석 / CRLF / 말미 개행 없음) | `ok` | **PASS** |
| AC-CSD-023 | `go test -run TestUpsertCodexSkillDisableSkipsUnrecognisedEntry ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-030 | `go test -run 'TestRunCodexSkillDisableDryRunIsDefault\|TestSkillsDisableRequiresCodexFlag' ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-031 | `go test -run 'TestRunCodexSkillDisableBacksUpAndPreservesMode\|TestRunCodexSkillDisableBackupFailureLeavesTargetIntact' ./internal/cli/` | `ok`. e2e 실측도 동반: `pre-run config mode: 644` → `post-run config mode: 644` | **PASS** |
| AC-CSD-032 | `go test -run TestRunCodexSkillDisableFailsOpenOnAbsentConfig ./internal/cli/` | `ok` | **PASS** |
| AC-CSD-033 | `go test -run TestRunCodexSkillDisableReasonsAreDistinct ./internal/cli/` (5상태 전수) | `ok` | **PASS** |
| AC-CSD-040 | `git diff --name-only bf779ecf2..HEAD` | `internal/` 파일 **4개뿐**: `codex_skills_disable.go` · `codex_skills_disable_test.go` · `root.go` · `skills.go`. `internal/codexwiring/skills.go` **없음**, `internal/cli/codex_skills_prune.go` **없음** | **PASS** |
| AC-CSD-050 | `bash .moai/reports/t502/probe.sh selftest` @ `codex-cli 0.153.4` | 2셀 재현 + 음성 대조군. 판정서 `.moai/reports/t502/regate-0.153.4.md` | **PASS** |

### 판별 셀(뮤턴트) — 5종 전부 실제로 RED 관측

성질 하나에 셀 하나. 각 뮤턴트는 넣고 → 실패 출력을 받고 → 되돌렸다.

| 뮤턴트 | 축 | 관측된 RED |
|---|---|---|
| 추가 경로를 no-op으로 | AC-CSD-001 추가 축 | `codex_skills_disable_test.go:165: entries declaring "…" = 0, want exactly 1` |
| `enabled = true` 발행 | AC-CSD-001 값 축 | `codex_skills_disable_test.go:178: enabled = 1, want SkillEnabledFalse` |
| `enabled` 줄 발행 제거 | AC-CSD-001 스키마 축 | `codex_skills_disable_test.go:174: entry declares no 'enabled' key; a codex reading this config fails to start` |
| 해소 `.claude/…` 표기 발행 | AC-CSD-002 표기 축 | symlink 셀: `path = "…/.claude/skills/t502probe/SKILL.md", want "…/.agents/skills/t502probe/SKILL.md"` |
| 대상을 무조건 `0600` 으로 쓰기 | AC-CSD-031 모드 축 | `codex_skills_disable_test.go:486: config mode = 0600, want 0644 preserved` |

세 축이 **서로 다른 단언 줄**에서 실패한다(165 / 178 / 174) — 한 셀이 두 성질을 겸하지 않는다.

**표기 뮤턴트의 범위 한정(불리한 관측도 기록)**: 이 뮤턴트를 잡은 것은 **심링크 셀**이다. 복사 셀은 `/private` 접두 불일치로 실패했을 뿐, 표기 결함 자체로 실패하지 않았다 — 복사 미러에서는 해소 경로가 곧 `.agents` 실파일이라 뮤턴트가 실제로 무해하기 때문이다. 즉 이 결함은 심링크 배포에서만 해롭고, 그것을 잡는 셀도 정확히 심링크 셀 하나다.

### E2E — 대역이 아니라 실제 verb로 닫았다

`gate-path-shape.md` 가 남긴 Gap(「진짜 verb로는 E2E 축을 못 돌렸다」)이 여기서 닫힌다. 계측기는 `--entry-path`/`--enabled` 를 받지 않고 **verb가 쓴 config 를 그대로 읽었다**.

```
== probe BEFORE the verb (expect exposed) ==
verdict=exposed marker=1 name=1 rc=0 out=24946 err=0
expect=exposed result=MATCH
== the verb ==
[force] · Backup: …/config.toml.bak-20260907T094328Z
[force] · Backup sha256: 60c025ec7ebff3afa6594623ab732e1bcb59da7baa321e83faad0f59891f3679
[force] · Disabled t502probe for Codex: …/.agents/skills/t502probe/SKILL.md
== what the verb wrote ==
[[skills.config]]
path = "/tmp/t502-e2e-5QecqrSk/lab/proj/.agents/skills/t502probe/SKILL.md"
enabled = false
post-run config mode: 644
== probe AFTER the verb (expect gated) ==
verdict=gated marker=0 name=0 rc=0 out=24776 err=0
expect=gated result=MATCH
== idempotence: re-run --force must not add a second entry ==
[force-2] · Unchanged: … (it is already disabled)
entries declaring the path: 1
E2E PASS
```

같은 스크립트를 `symlink` 모드로도 돌려 동일 판정(`marker=1 → 0`, `E2E PASS`). 즉 **두 미러 모양 모두**에서 실제 verb의 발행이 게이트를 묶는다.

### 실행 중 발견해 고친 것 2건 (E2E가 아니었으면 안 잡혔다)

1. **프로젝트 루트 앵커** — 처음엔 `findProjectRoot()`(`.moai` 상향 탐색)로 잡았다. E2E가 `/tmp` 픽스처에서 **무관한 `/private/tmp/.moai`** 를 프로젝트로 지목해 미러를 `/private/tmp/.agents/skills` 에서 찾는 것을 드러냈다. codex 는 `.agents/skills` 를 **자기 cwd** 기준으로 푸는 것이 실측이므로, 앵커를 **cwd 고정**으로 바꿨다. 사용자 기계에서도 같은 부류의 오지목이 나올 수 있던 결함이다.
2. **E2E 픽스처 config 가 codex 에게 무효** — `[model]\nname = "x"` 는 codex 파서가 `invalid type: map, expected a string` 로 거절해 rc=1 이었다. 계측기가 이것을 `gated` 가 아니라 **`verdict=ERROR`** 로 보고한 것이 규약대로 동작한 증거다(rc≠0 위의 0은 판정이 아니다). 픽스처를 주석만 있는 config 로 고쳤다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: eef6061a6
run_status: complete
ac_pass_count: 17
ac_fail_count: 0
mutants_observed_red: 5
m1_regate:
  version_stamp: "codex-cli 0.153.4"   # 이 실행의 `codex --version`, 세션 반입 아님
  cells_reproduced: 2                   # symlink literal_false=gated, copy literal_false=gated
  positive_controls: 2                  # control=exposed ×2
  negative_control_absent_token: 0      # ZZZNOTAMARKER, 5개 캡처 전부
  negative_control_real_marker: 1       # 통제 셀 — 카운터가 0이 아닌 값을 낼 수 있다
  report: .moai/reports/t502/regate-0.153.4.md
boundary_check:
  command: "git diff --name-only bf779ecf2..HEAD"
  internal_files: 4
  codexwiring_skills_go_present: false
  codex_skills_prune_go_present: false
prune_regression:
  population_command: "go test -list 'TestPruneCodexSkillEntries|TestJudgeCodexSkillEntry|TestRunCleanCodexSkills' ./internal/cli/..."
  population: 13          # 0이 아님 — 셀렉터가 0개를 고르면 `ok` 를 찍는다
  population_measured_twice: [f6ed23a9c계열, fabb5b000]   # run 시작·착지 tip 양쪽에서 13
  population_pinned_into_criterion: true                   # AC-CSD-040 본문에 값+명령 기입(리드 지시)
  pass_lines: 22          # 서브테스트 포함 실제 `--- PASS` 행 수
  verdict: ok
verification:
  tests: "go test ./internal/cli/... ./internal/codexwiring/... → exit 0 (18 패키지 전부 ok, internal/cli 440s)"
  vet: "go vet ./internal/cli/... ./internal/codexwiring/... → exit 0"
  lint: "golangci-lint run ./internal/cli/... ./internal/codexwiring/... → 0 issues"
  full_suite: "CI 몫 — 로컬 전체 스위트는 돌리지 않았다(CLAUDE.local.md §4)"
new_warnings_or_lints_introduced: 0
cross_platform_build: not_run          # Gaps 참조
m1_to_mN_commit_strategy: "M1(재측정 게이트) 선행 → 구현 단일 커밋 eef6061a6 → 진행기록 커밋"
```

### Gaps — 관측하지 **않은** 것

- **크로스플랫폼 빌드 미실행.** `GOOS=windows`/`linux` 빌드를 돌리지 않았다. 아래 Windows 유보와 함께 읽어야 한다.
- **Windows 에서 이 verb 는 발행을 거절한다.** 발행 경로에 `\` 가 들어가면 병합기가 「이 config 형식이 그대로 담을 수 없는 문자」 사유로 **건너뛴다**. 이유: 이 저장소의 파서는 basic string 값을 **이스케이프 해독 없이** 읽으므로, `\\` 로 이스케이프해 쓰면 codex 는 옳게 읽지만 moai 의 모든 독자는 **다른 경로**로 읽는다 — 그리고 그 중 하나가 「찾지 못한 경로를 지우는」 prune 이다. 유효하지 않은 TOML 을 쓰거나 자기 자신에게 보이지 않는 엔트리를 쓰는 것보다 거절이 낫다고 판단했다. SPEC 이 다루지 않은 축이며, 후속 카드 소관으로 남긴다.
- **홈 미러에만 있는 스킬은 거절한다.** SPEC 은 발행 표기를 프로젝트 미러 경로로 고정했으므로(REQ-CSD-012), 홈 미러에만 있는 이름에 프로젝트 경로를 발행하면 태어날 때부터 죽은 엔트리가 된다. 그래서 「해석 불가」 부류로 거절하되 **사유는 따로** 붙였다. plan.md §B 표가 홈 미러의 지위를 「모호성의 원천」으로만 두는 것과 정합한다.
- **`moai update` 가 만드는 실제 미러 트리 미관측** — 픽스처는 `skill_mirror.go` 모양을 손으로 재현한 것이다(선행 측정과 같은 Gap).
- **단일 codex 버전(0.153.4)** — 버전 드리프트를 관측하지 못했다. 게이트가 답하는 것은 「착지 시점 버전에서 재현되는가」뿐이다.
- **`MirrorModeFailed` / `SKILL.md` 없는 `MirrorModeSkipped` 미측정** — 선행 측정의 Gap 그대로.

### Residual-risk

- **절대경로 발행은 구조적으로 유령 엔트리를 만든다.** 프로젝트를 옮기면 엔트리는 죽고 codex 는 침묵한다. 수거기(`moai clean --codex-skills`, t506)가 **이미 착지해 있다**는 사실이 이 부채를 회수 가능하게 만든다. 도움말 2번 항목이 사용자에게 그 절차를 말한다.
- **`.agents/skills` 는 cwd 에 묶여 있다.** 서브디렉터리에서 codex 를 띄우면 루트 자체가 사라져 발행한 엔트리가 조용히 아무것도 하지 않는다. 이 카드는 앵커를 cwd 로 맞춰 verb 와 codex 가 **같은 답**을 보게 했고, 도움말 1번 항목이 사용자에게 알린다.
- **같은 이름이 다른 루트에서 다른 실파일로 겹쳐 올라오면 노출이 남는다.** 게이트는 파일 단위다. 도움말 3번 항목이 그 경우 무엇을 할지 말한다.
- **정규화는 codex 구현 세부다.** 문자열 비교로 회귀하면 이 게이트를 다시 돌려야 한다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
