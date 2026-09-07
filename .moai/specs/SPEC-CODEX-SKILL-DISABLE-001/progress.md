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

> **재닫기다.** 첫 닫기는 `43e820663` 이었고, 그 뒤 독립 sync-audit 이 **FAIL 89.0** 을 blocking 1건과 함께 돌려보냈다. 수리 커밋 둘(`bc651da28`, `13ae49a05`)이 착지하면서 닫힌 산출물이 **두 커밋 낡은 트리를 기술하는** 상태가 됐다 — 「sync 닫기는 마지막 쓰기여야 한다」는 바로 그 모양이다. 아래는 그 재측정이며, **측정 트리는 `13ae49a05`** 다.

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: pending-backfill-resync   # the re-close lands in the commit that carries it; a SHA cannot name itself
sync_status: complete
resync_of: 43e820663                       # the first close, superseded by this one
sync_audit_verdict: "FAIL 89.0 @ 43e820663 (independent sync-auditor); verdict + all advisories recorded at .moai/reports/t502/sync-audit.md"
sync_audit_disposition:
  f1_blocking: "REPAIRED at bc651da28 - three sites spell the enabled emission, not one; the insert branch (the total-outage shape) had no test. Evidence is a MUTANT-6 contrast, not a RED-then-GREEN: the branch was already correct, so no pre-implementation failure exists"
  f2_advisory: "CLOSED FOR FREE at 13ae49a05 - writing the F1 fixture in CRLF covered reshapeLike's CR branch (332.36: 0 -> 1); MUTANT-7 shows the test discriminates, because line coverage asserts nothing"
  open_advisories: "THIRTEEN remain OPEN and UNOWNED: F3-F12 (ten, round 1) + F13-F15 (three, raised by round 2). Not addressed by this close. Do not read the repairs above as covering them"
  round2: "FAIL 90.8 @ f269edbc2 (.moai/reports/t502/sync-audit-round2.md). Its sole blocking finding F0 - the open-advisory count read nine where F3-F12 is ten - was repaired at 35d6c7404. Round 2 independently re-injected the mutants and reproduced the counterfactual; it also raised F13-F15"
b12_self_test_a: "grep -c 'SPEC-CODEX-SKILL-DISABLE-001' CHANGELOG.md -> 1. NOTE the changed meaning: this is a re-close editing the EXISTING bullet in place, not a fresh emission, so the pre-emission 0 of the first close does not apply. 1 = exactly one entry, still no duplicate"
b12_self_test_b: "grep -c '^### AC-CSD' acceptance.md -> 17 (non-zero; matches the AC matrix row count; recorded Tier-M budget exception per plan.md §F.1, NOT an error). REQ: grep -c '^- \\*\\*REQ-CSD-' spec.md -> 16"
b12_self_test_c: "every path named in the t502 CHANGELOG bullet verified with ls -> 9/9 present"
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added, first bullet (line 12) - edited in place, not re-emitted"
frontmatter_status_transitions:
  spec_md: "status: completed - ALREADY TRANSITIONED at 43e820663 and still correct. This re-close asserts NO new transition; an amendment note is what the tree needed, because what went stale was the EVIDENCE, not the status. Re-asserting completed -> completed would be a no-op dressed as a decision"
  plan_md: "no YAML frontmatter in this artifact - nothing to transition"
  acceptance_md: "no YAML frontmatter in this artifact - nothing to transition"
  progress_md: "no YAML frontmatter in this artifact - nothing to transition"
  updated_field: "spec.md updated: 2026-09-07 - unchanged, and correct: the re-close lands the same calendar day as the first"
canary_compliance_check: not_applicable   # this SPEC defines no forward-looking policy its own sync tests
verification:                              # every row below re-measured at 13ae49a05 by the re-closing agent
  tests: "go test ./internal/cli/... ./internal/codexwiring/... -> GOTEST_EXIT=0; 18 'ok' lines, 0 FAIL lines (the ok count guards against reading a green off an empty file), internal/cli 440.982s"
  vet: "go vet ./... -> exit 0, 0 lines of output (module-wide)"
  spec_lint: "moai spec lint .moai/specs/SPEC-CODEX-SKILL-DISABLE-001/spec.md -> exit 0, 'No findings'"
  spec_audit: "mcp__moai__spec_audit(project_root=<this worktree>, filter_spec=SPEC-CODEX-SKILL-DISABLE-001) -> modern_era_clean 1, drift INFO only (EraAutoDetected V3R6, H-4)"
  lint: "NOT re-run at this HEAD. golangci-lint was exit 0 / '0 issues.' at the first close (43e820663); the two commits since are one source file and one test file. Recorded as carried-over, NOT as a fresh measurement"
  full_suite: "CI's job - no local full suite (CLAUDE.local.md §4)"
sync_phase_independent_reproductions:      # re-run at 13ae49a05, not carried over
  e2e: "bash .moai/reports/t502/e2e-verb.sh copy -> exit 0, codex-cli 0.153.4, verdict=exposed marker=1 -> verdict=gated marker=0, result=MATCH x2, pre/post config mode 644, re-run entries declaring the path=1, E2E PASS. Live ~/.codex/config.toml sha256 IDENTICAL before and after (c91a6b73...69598)"
  ac_test_name_existence: "go test -list '<17 test names>' ./internal/cli/ | grep -c '^Test' -> 17 (16 AC-named + the new insert test; a selector matching none also prints ok, so this control is the premise of the package green)"
  boundary: "git diff --name-only bf779ecf2..HEAD | grep '^internal/' -> STILL 4 files; --stat on internal/codexwiring/skills.go and internal/cli/codex_skills_prune.go -> empty. Control run: the same --stat on codex_skills_disable.go DOES report (455 insertions), so the empty result is a real zero, not a silent one"
  regate: "NOT re-run at this HEAD - probe.sh selftest was reproduced at the first close and at run-phase, both @ codex-cli 0.153.4. The e2e above exercises the same gate end to end"
dod_boxes_ticked: 8
dod_boxes_left_unticked: 0
evidence_dir: .moai/state/verify/t502-resync/   # first close: .moai/state/verify/t502-sync/ ; audit: .moai/state/verify/t502-audit/
```

### 판정 이후 트리에 남은 두 사실 — 결함이 아니라 기록

- **`pathLine < 0` bail 은 커버리지 0 으로 의도적으로 남는다.** 유일한 호출자가 만들 수 없는 입력이라(스캔이 접두로 매치하므로 `Path` 가 비지 않으면 그 줄은 반드시 존재한다) 그 분기에 닿는 테스트는 **아무것도 고정하지 못하면서 죽은 분기를 살아있는 것처럼 굳힌다**. 다음 커버리지 판독이 다시 깃발을 꽂지 않도록 이유가 코드 위에 적혀 있다.
- **`upsertCodexSkillDisable` 의 사후조건 가드는 후보이지 간극이 아니다.** 「`enabled` 키 없는 엔트리의 반환을 거부」하면 발행 지점 셋을 한 번에 기계적으로 덮지만, 그것은 **런타임 동작 변경**이라 blocking 최소 수리의 범위 밖이었다. 후속 카드가 집을 수 있도록 후보로 남긴다.

### Gaps — sync-phase가 관측하지 **않은** 것

- **뮤턴트 7종 중 하나도 재주입하지 않았다.** 뮤턴트 1-5 는 인용된 실패 문구가 실재함을 대조했고(줄 번호는 재닫기에서 옮겨졌다 — 모드 축 486 → 544), MUTANT-6·7 은 **구현 커밋 본문의 기록**과 그 단언이 실재함(222·230·233행)이 근거다. 셋 다 재주입해 RED를 본 것이 아니다 — 재주입은 `internal/` 쓰기이고 이 위임의 경계 밖이다. §G 체크 근거 표 2·2b·2c 행이 이 한계를 행별로 명시한다.
- **MUTANT-6·7 은 이제 제3자가 재주입해 확인했다 — 이 간극은 해소됐다.** 재닫기 시점의 근거는 구현 커밋의 기록과 단언의 실재까지였고, 위 두 항목이 그 한계를 적는다. **2차 sync-audit(`.moai/reports/t502/sync-audit-round2.md`, 앵커 `f269edbc2`)이 그 대조를 독립적으로 재현했다**: 발행 지점 **셋 각각**에 뮤턴트를 넣어 저마다 자기 테스트를 붉힌다(MUTANT-6 삽입 분기 → `…InsertsMissingEnabledKey`, MUTANT-R 재작성 분기 → `…UpdatesExistingEntry`, MUTANT-A append 조립 → `…AppendsOneEntryWithFalse` + `…PreservesSurroundings`), MUTANT-7 은 `reshapeLike` 의 CR 계승을 지워 같은 새 테스트를 붉힌다. 기준선은 `EXIT=0 PASS=17 FAIL=0`.
- **결정적인 것은 반사실 대조다.** `43e820663` 시점 테스트 파일을 되돌려 놓고 **같은** MUTANT-6 을 넣으면 `EXIT=0 PASS=16 FAIL=0` — 수리 전에는 삽입 지점 발행을 지워도 초록이었다는 1차의 주장이 **논증이 아니라 실측으로** 재현된다. 이것이 뮤턴트 근거를 「구현 에이전트가 기록함」에서 「제3자가 재현함」으로 끌어올린다.
- **감사관은 감사 트리에 뮤턴트를 넣지 않았다.** `rsync` 스크래치 사본에서 수행했고, 대상 3파일의 sha256 동일을 먼저 확인했다(`codex_skills_disable.go` `8819ef6d…`, `_test.go` `c50a724d…`, `codexwiring/skills.go` `2be7579f…`). `internal/` 무수정 — 그래서 이 재현은 위 「재주입은 위임 경계 밖」과 충돌하지 않는다.
- **advisory 열세 건이 열린 채 소관 미배정이다** — 1차의 **F3-F12 열 건**(`.moai/reports/t502/sync-audit.md`)에 2차가 올린 **F13-F15 세 건**(`sync-audit-round2.md`)이 더해진다. 이 닫기는 그 중 어느 것도 건드리지 않았다 — 1차 blocking 수리와 F2 가 픽스처로 딸려 닫힌 것, 그리고 2차 blocking(F0, 수치 오산)의 수리가 전부다. **2차가 새 advisory 를 올렸으므로 「열 건」은 `35d6c7404` 시점까지만 참이었다**: 같은 절이 같은 이유로 두 번 어긋난 셈이라, 이 수치는 판정서를 다시 세어 고쳤다.
- **`golangci-lint` 판정을 이 HEAD 에서 **어느 쪽도** 갖고 있지 않다.** 닫기의 exit 0 은 `43e820663` 의 측정이라 carried-over 로 적혀 있고, 2차 감사관도 돌리지 않았다(2차 「보지 않은 것」 목록). 즉 이월된 값이 있을 뿐 현재 트리의 lint 관측은 없다. vet 만 이 HEAD 에서 모듈 전체(`./...`)로 재측정됐다.
- **최종 트리에서 E2E 를 재관측한 주체가 없다.** 닫기가 `13ae49a05` 에서, 1차 감사가 `43e820663` 에서 각각 재현했으나 **`f269edbc2` 이후로는 어느 쪽도 돌리지 않았다**. 간극으로 닫지 않고 기록만 하는 이유는 그 사이 소스 diff 가 **주석·문서 전용(실행문 0줄)** 이기 때문이다 — 판단의 근거를 함께 적으므로 다음 독자가 누락만 보고 재실행 여부를 스스로 정할 수 있다. **이 절이 모든 차원을 최종 트리에서 재관측했다는 뜻으로 읽혀서는 안 된다.**
- **이 §E.4 가 인용하는 `.moai/state/verify/t502-resync/` 는 추적되지 않는다**(2차 F15). `.gitignore` 가 `.moai/state/` 를 무시하므로 워크트리를 폐기하면 그 경로는 해소되지 않는다 — 판정서 본문과 이 산출물은 커밋돼 있어 핵심 기록은 남지만, **인용 경로의 내구성은 이 카드 고유가 아닌 계통적 위험**으로 열려 있다.
- **전체 테스트 스위트 미실행 · 크로스플랫폼 빌드 미실행.** 둘 다 CI 몫이며 run-phase Gap 그대로다.
- **문서 표면은 CHANGELOG 하나다.** README·docs-site 4개 로케일에 이 verb 를 싣지 않았다 — 위임이 지목한 산출물이 CHANGELOG 였고, 문서 사이트 항목 추가는 이 카드의 범위 밖이다. 자매 카드 t506(`moai clean --codex-skills`)이 docs-site 페이지를 동반한 전례가 있으므로, **후속 카드로 남길 가치가 있는 간극**으로 기록한다.

### Residual-risk

- **CHANGELOG 항목의 서술은 run-phase 측정에 의존한다.** 게이트 성질(realpath 비교), 금지 표기 3종, `enabled` 필수성은 모두 선행 측정 보고서에서 온 것이고 sync-phase가 재현한 것은 2셀 게이트와 E2E 두 축이다.
- **단일 codex 버전(0.153.4).** sync-phase 재현도 같은 버전이다 — 버전 드리프트는 여전히 미관측이다.
- **한 번 낡았던 닫기는 다시 낡을 수 있다.** 이 닫기가 기술하는 트리는 `13ae49a05` 이며, 이 커밋 뒤에 코드가 착지하면 같은 방식으로 다시 어긋난다. 판별식은 `git diff --name-only <닫기커밋>..HEAD | grep '\.go$'` 이 무출력인지다.
- **advisory 열세 건이 열린 채 남는다.** 그 중 어느 것도 blocking 으로 승격되지 않았으나, F4(Windows 거절 가드 무테스트)와 F5(skip 이 rc=0)는 CHANGELOG 가 이름 붙여 출하한 성질에 걸려 있어 **문서가 주장하는 것과 테스트가 지는 것 사이의 간극**으로 남는다. 2차가 올린 F13·F14 도 같은 축이다 — 주석이 국소 사실에서 전역 안전 성질을 끌어내거나, 헤더가 실제보다 넓게 말한다. 1차 blocking 이 지적한 모양의 잔재이며 거짓 주장은 아니어서 advisory 에 머문다.
