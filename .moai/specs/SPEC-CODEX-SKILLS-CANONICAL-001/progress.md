# progress — SPEC-CODEX-SKILLS-CANONICAL-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (spec.md + plan.md + acceptance.md). 근거: 예상 변경 파일 8-12개(`internal/template/` 신규 1 + `deployer.go` + `internal/cli/update/deploy/deploy.go` + `internal/defs/dirs.go` + 테스트 4-6 + 템플릿 미러), 예상 LOC 400-700 → Tier M 대역(300-1000 LOC / 5-15 files).
- 요구사항 12개 / 판정 기준 15개 — Tier M 상한 16/16 이내.
- SPEC ID 정규식 검사: `SPEC-CODEX-SKILLS-CANONICAL-001` → `PASS` (Bash `[[ =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`).
- 중복 검사: `.moai/specs/` 에 동일 ID 없음 (`SPEC-CODEX-PHASE2-001` 만 존재).
- 착수 시점 실측(작성 시점): 템플릿 스킬 34개 / core 21 / non-core 13 / 템플릿 트리 심볼릭 링크 0개.
- iter-7 (v0.7.0): plan-audit iter-5 **PASS(0.875)** 후 마감 편집 5건. 요구사항 설계 무변경. **D1**(blocking, 축소가 만든 결함) — 폴백 플랫폼에서 미러가 2회차부터 REQ-CSC-014 실 항목 분기에 걸려 1회차 내용에 고착되는 결과를 §D 축소된 보장에 잔존 고지와 나란히 고지, 수정 (b) 채택(요구사항 무수정). **D2** — plan M1 닫힘 조건 `AC-CSC-006`(M2 산출 seam 전제) → `AC-CSC-003, AC-CSC-014`, M2 에서 AC-CSC-003 제외(거울상 오배치 동시 해소). **D3** — 승계 SPEC 지목: 카드 `t173`, SPEC ID 미발행 사실·부여 시점·갱신 위치 명시. **D4** — §D 첫 H3 에 `-` 불릿 2줄. **D5 수용** — REQ-CSC-006 에 REQ-CSC-001 과 같은 예외 축(011 실패 / 014 선점) 추가.
- iter-7 자기 지적: 분할 예측("미러 절반은 즉시 통과")은 **부분적으로만 성립**했다. 미러 절반이 blocking 을 내지 않았던 것은 건전해서가 아니라 감사 초점이 늘 청소 절반에 있었기 때문이고, 정면으로 읽히자 D1 이 즉시 나왔다. **검사되지 않은 영역의 무결함은 건전성의 근거가 아니다** — iter-6 의 형제 경로 미조사와 같은 계열.
- iter-6 (v0.6.0): **범위 분할** — 청소 계열 전출. 리드가 감사 권고(범위 축소)를 승인. 요구사항 16 → **14**, 판정 16 → **13**. 전출: REQ-CSC-008·009, REQ-CSC-010 의 백업 절, AC-CSC-007·008·009, AC-CSC-012 의 백업 두 팔, §A.4·§A.5·§A.7·§A.10·§A.11, §B.D5·§B.D6, plan M4 + R4·R5·R8·R13·R14·R16 + AP-6·AP-12·AP-13·AP-14·AP-16. **번호는 옮기지 않고 구멍으로 남겼다**(감사 보고서 4건이 인용하므로). REQ-CSC-010 은 manifest 절만 남겨 축소(배포 계열이라 전출하면 REQ-CSC-011 fail-open 과 충돌하는 구멍이 생김), REQ-CSC-015 는 유지하되 근거를 "승계 글롭이 도달할 이름을 생산 시점에 보증"으로 재정립. spec §D 에 **축소된 보장**(개명·은퇴 스킬의 미러가 영구 잔존)을 본문으로 명시.
- iter-6 실측 기록: `copyRegularFile`(`internal/cli/update/deploy/deploy.go:464-473`) 본체가 `os.ReadFile(src)` 임을 직접 확인 — 심볼릭 링크를 따라가므로 디렉터리 링크에서 EISDIR. iter-4 감사 D1(“`os.Lstat` 치환이 `moai update` 를 실패시킨다”)의 코드 경로가 실재함을 재확인했고, 이것이 분할을 확정한 설계 계층 근거다.
- iter-6 자기 지적: §A.6 이 `manifest.Track → HashFile → io.Copy` 에서 **같은 형태**의 EISDIR 위험을 이미 측정해 설계를 뒤집었으면서, 형제 경로인 `copyRegularFile → os.ReadFile` 을 다섯 판본 동안 훑지 않았다. **한 호출 경로에서 측정된 위험은 형제 경로를 훑을 이유**라는 규율을 spec HISTORY 와 §A.6 에 기록했다.
- iter-5 (v0.5.0): iter-4 가 들여온 결함 3건(D2/D3/D3-b) + 인용 오류(D1/D7) + optional 4건 대응. 설계 무변경, 요구사항 계층 무변경. **REQ 16 / AC 16 불변 — 은퇴 없음.** AC-CSC-008 fixture 4형태 → **5형태**(`moai-linkprobe` 링크 추종 탐침 추가), 단언 4 → 6, 판정 syscall 을 `os.Lstat` 으로 [HARD] 고정. AC-CSC-012 2번 팔 측정 폭을 서브트리 전체 → **이름 단위**로 축소. §A.9b 신설(출력 표면 부재 실측 승격) + 출력 seam 인용 4곳 `§B.D6` → `§B.D3` + `§A.9b`. plan M3 본문·닫힘 조건 정정.
- iter-5 실측 기록: (a) dangling 링크 메커니즘 독립 재현 → `Stat err: … no such file or directory  IsNotExist: true` / `Lstat err: <nil>` / `glob: [a/.agents/skills/moai-x]`. (b) clean→deploy 순서 — `update_template_sync.go:297`(clean) < `:323`(deploy). (c) `ManagedCleanTargets` 는 목록 전체를 순회하며 4번째 항목이 `.claude/skills/moai*` → fixture 의 `moai-live` 정본은 링크와 무관하게 제거됨(D2 의 근거).
- iter-5 감사 쟁점: iter-3 감사의 지적 중 반박한 것 없음 — D1~D7 전건이 재현·확인됐고 전부 반영했다. 리드가 제시한 D2 대안 (b) 만 근거를 들어 채택하지 않았다(배포가 손상을 치유해 관측 불가; 사유는 AC-CSC-008 본문).
- iter-4 (v0.4.0): 독립 감사 2건(0.78 / 0.7625) 대응, 리드 상한 예외 승인 하에 수행. **REQ 16 / AC 16 불변 — 은퇴시킨 항목 없음**(모든 수정이 기존 번호 안의 절 추가·수정). 핵심은 §A.10 dangling 결함: REQ-CSC-008 에 `os.Lstat` 판정 + dangling 제거(본체) + 슬라이스 순서(이중 방어), AC-CSC-008 fixture 를 미러 4형태로 확장. 그 밖에 REQ-CSC-001 예외 절, REQ-CSC-010 백업 금지 한정(§A.11), REQ-CSC-005 반환 결과 seam, REQ-CSC-016 범위 축소, §A.9 접두 철자 정정(`moai-` → `moai`), REQ-CSC-007 문구 축소. optional 전건 반영. AC-006·015 를 SHOULD → MUST 로 승격(MUST 13 → 15).
- iter-4 실측 기록: (a) dangling 링크 — 직접 실행 → `Stat err: … no such file or directory  IsNotExist: true` / `Lstat err: <nil>` / `glob: [a/.agents/skills/moai-x]`. (b) 실행 순서 — `update_template_sync.go:297`(clean) < `:323`(deploy). (c) 출력 seam 부재 — `internal/template` 비-테스트 파일에 `io.Writer` 매치 0. (d) 접두 — `grep -cv '^moai-'` → **1**(`moai`), `grep -cv '^moai'` → 0.
- iter-4 감사 쟁점 확인: audit#1 의 N10(M3 모순)·N11(M5 미반영)·N12(백업 소유 마일스톤 부재)는 **디스크 실물 확인 결과 iter-3 에서 이미 닫혀 있었다** — `plan.md` 의 M3 제목은 "미러를 기록·백업 대상에서 제외 (Priority High)", M5 는 "`.gitignore` 에 `.agents/` 등록 (Priority **High**)", AC-CSC-012 는 M3 닫힘 조건에 등록돼 있다. audit#2 가 같은 결론을 냈다. audit#1 이 개정 전 스냅샷을 읽은 것으로 보인다.
- iter-3 (v0.3.0): plan-audit iteration 1 (FAIL 0.775) blocking 8건 대응. 요구사항 12 → **16**, 판정 15 → **16**(둘 다 Tier M 예산 상한 도달). §A 에 실측 4건 추가(§A.6 `manifest.Track` EISDIR, §A.7 pre-clean 백업 비대칭, §A.8 `.gitignore` 부재, §A.9 접두 우연 일치). REQ-CSC-010 방향 반전(기록한다 → 기록·백업하지 않는다). AC-CSC-001/010/013 재작성, AC-CSC-007 경로 구분자 중립화, AC-CSC-011 3-상태 확장, AC-CSC-016 신설. plan M3 방향 반전 · M5 Priority Low → High · M6 신설 · AP 5건 추가. D12 기각(사유 spec §G). 실측 재현 명령과 출력은 아래 iter-3 실측 기록 참조.
- iter-3 실측 기록: (a) `manifest.HashFile` 디렉터리 링크 — 별도 Go 프로그램 실행 → `open err: <nil>` / `copy err: read lnk: is a directory`. (b) 템플릿 `.gitignore` `.agents/` 항목 부재 — `grep` 결과 `.claude/` 계열만. (c) `optional-pack:*` tier — `grep -c 'tier: optional-pack'` → 13, `harness_generated.skills` → `[]`. (d) 비-`moai` 접두 템플릿 스킬 — `find … -exec basename {} \; | grep -cv '^moai'` → 0.
- iter-2 (v0.2.0): 리드 추가 제약 반영 — 청소 글롭 접두 `moai*` 한정을 spec §B.D5 [HARD] 로 고정, `ManagedCleanTargets` 확장이 `moai update` 동작 변경임을 §A.5 에 근거(두 차례 실측된 같은 실패 형태)와 함께 기록, AC-CSC-008 을 제거+생존 **양팔 단일 테스트**로 재작성. 요구사항 12 / 판정 15 불변, `moai spec lint` 무결점.

## §E.2 Run-phase Evidence

착수 시점 재측정 (worktree `WT-skills-canonical`, 기준 `a338eab1b`):

| 항목 | 명령 | 값 |
|---|---|---|
| 템플릿 스킬 디렉터리 | `find internal/template/templates/.claude/skills -mindepth 1 -maxdepth 1 -type d \| wc -l` | **34** (작성 시점과 동일) |
| 템플릿 트리 심볼릭 링크 | `find internal/template/templates -type l \| wc -l` | **0** |
| 비-`moai-`(하이픈) 이름 | `... -exec basename {} \; \| grep -cv '^moai-'` | **1** — 이름이 정확히 `moai` 인 스킬. plan §C 가 예고한 값 |
| `.gitignore` `.agents` 항목 | `grep -n 'agents' internal/template/templates/.gitignore` | 착수 시점 매치 없음 (M5 의 전제 성립) |

### AC 판정 매트릭스

| AC | 상태 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-CSC-001 | PASS | `go test ./internal/template/ -run TestEmbed_TemplateTreeHasNoSymlinks -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.290s` |
| AC-CSC-002 | PASS | `go test ./internal/template/ -run TestSkillMirror_BothPathsReadableAfterFullDeploy -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.788s` |
| AC-CSC-003 | PASS | `go test ./internal/template/ -run TestSkillMirror_SlimSetEqualsCanonicalAndIsSmaller -count=1` | `ok ... 0.920s` (슬림 21 < 전량 34) |
| AC-CSC-004 | PASS | `go test ./internal/template/ -run TestSkillMirror_LinkIsRelativeAndResolves -count=1` | `ok ...` — darwin 에서 링크 모드 성립, 링크 본체 `../../.claude/skills/<name>` |
| AC-CSC-005 | PASS | `go test ./internal/template/ -run TestSkillMirror_CopyFallbackIsReadable -count=1` | `ok ...` |
| AC-CSC-006 | PASS | `go test ./internal/template/ -run TestSkillMirror_FallbackIsObservable -count=1` | `ok ...` (양방향) |
| AC-CSC-010 | PASS | `go test ./internal/template/ -run TestSkillMirror_ClaudePathUnchangedByMirror -count=1` | `ok ...` |
| AC-CSC-011 | PASS | `go test ./internal/template/ -run TestSkillMirror_PreoccupiedTargets -count=1` | `ok ...` (3-상태 전부) |
| AC-CSC-012 | PASS | `go test ./internal/template/ -run TestSkillMirror_NotTrackedInManifest -count=1` | `ok ... 0.563s` |
| AC-CSC-013 | PASS | `go test ./internal/template/ -run TestSkillMirror_FailOpen -count=1` | `ok ...` |
| AC-CSC-014 | PASS | `go test ./internal/template/ -run TestSkillMirror_SetIsDerivedNotConstant -count=1` | `ok ...` (합성 FS 2개 → 미러 정확히 2개) |
| AC-CSC-015 | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 22.076s` |
| AC-CSC-016 | PASS | `go test ./internal/template/ -run TestEmbed_SkillNamePrefixInvariant -count=1` | `ok ... 0.290s` |

전량 재실행: `go test ./internal/template/... -count=1 -cover` → `ok ... 45.814s coverage: 86.1% of statements`.

### 위증 검사 (각 판정이 지키는 것을 부수고 붉어지는지 확인)

판정이 통과했다는 사실만으로는 그 판정이 무엇을 지키는지 알 수 없으므로, 항목마다 지키는 대상을 일부러 깨뜨려 실패를 관측했다. 전부 되돌린 뒤 재실행해 green 을 재확인했다.

| 부순 것 | 붉어진 판정 | 관측 출력 |
|---|---|---|
| 미러 집합을 상수 3개로 고정 | AC-CSC-014 · 003 · 002 | `mirror set = [moai moai-alpha moai-workflow-tdd], want exactly [moai-alpha moai-beta]` / `slim mirror count 3 must be < full mirror count 3` |
| 링크 본체를 절대 경로로 | AC-CSC-004 | `link body for "moai-alpha" = "/tmp/absolute/.claude/skills/moai-alpha", want "../../.claude/skills/moai-alpha"` |
| 선점 실 디렉터리를 지우고 재생성 (AP-10) | AC-CSC-011(3) | `USER.md was destroyed by the re-deploy: open …/moai-gamma/USER.md: no such file or directory` |
| 미러 실패를 `Deploy` 오류로 전파 | AC-CSC-013(1) | `Deploy failed because mirroring failed: mirror failed: moai-alpha` |
| 폴백을 조용히 처리 | AC-CSC-006 · 005 | `CopyFallbackUsed() = false after an injected symlink failure — the fallback is silent` |
| 미러 단계가 `.claude/` 에 쓰기 | AC-CSC-010 | `.claude tree differs: 5 entries without mirror, 8 with` |
| 미러를 `manifest.Track` 에 전달 | AC-CSC-012 | `track mirror ".agents/skills/moai-alpha": manifest track hash: hash file: read …: is a directory` |
| 템플릿 트리에 심볼릭 링크 주입 | AC-CSC-001 (양팔) | `template source tree contains 1 symlink(s)` + `skill set on disk (35) differs from the embedded set (34)` |
| 위 상태에서 2번 팔 수집을 `IsDir()` 로 (AP-8) | — | 2번 팔이 **침묵**했다(1번 팔만 보고). AC 의 [HARD] 가 경계한 사각이 실재함을 확인 |
| 비-`moai` 이름 스킬 추가 | AC-CSC-016 | `1 skill name(s) without the "moai" prefix: [hns-prefixprobe]` |

두 실측이 SPEC §A 의 주장을 독립적으로 재확인했다. `manifest.Track` 은 디렉터리 링크에서 `is a directory` 로 실패하고(§A.6), `//go:embed` 는 링크를 오류 없이 버린다(§A.2 — 디스크 35 vs 임베드 34, 빌드는 조용).

### §D.4 간접 검증 (판정에는 넣지 않음)

- **커밋 기준선 대조 (1회)**: 변경 전 커밋 `a338eab1b` 를 `git archive` 로 스크래치에 펼쳐, 양쪽 트리에서 동일한 덤프 테스트로 전량 배포 후 `.claude/skills/` 의 (상대경로, SHA-256, 퍼미션) 목록을 뽑아 `diff` 했다. **262 항목, 차이 0** (`diff-exit=0`). 현재 트리 쪽은 미러 기능이 **켜진 기본 상태**로 측정했다. 덤프 헬퍼는 판정 대상이 아니므로 대조 후 삭제했다.
- **Windows**: `GOOS=windows go vet ./internal/...` → exit 0 (테스트 파일 포함 컴파일). `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. 실동작 판정은 CI 매트릭스 몫.
- **Codex 실제 노출**: 관측함. `CODEX_HOME=<scratch>` 로 사용자 홈을 분리한 스크래치 프로젝트에 `.claude/skills/moai-probeskill/` 정본과 `.agents/skills/moai-probeskill -> ../../.claude/skills/moai-probeskill` **상대 디렉터리 링크**를 만들고 `codex debug prompt-input`(모델 호출 0회) 실행 → 스킬 목록에 `moai-probeskill: … (file: <project>/.agents/skills/moai-probeskill/SKILL.md)` 로 나타났다. acceptance §D.4 가 지적한 공백 — M0 관측이 링크의 **형태**를 기록하지 않았다는 것 — 이 형태로 닫힌다: REQ-CSC-003 이 강제하는 상대 디렉터리 링크가 실제로 따라가진다.

### 범위 준수

- 변경 파일 8개 전부 `internal/template/` 안이며, spec §D 의 범위 밖 항목은 하나도 실행하지 않았다. `~/.codex/` 는 무변경이다 — 노출 관측은 `CODEX_HOME` 을 스크래치로 돌려 수행했다.
- `internal/cli/` 호출부는 건드리지 않았다. REQ-CSC-005 는 배포기가 모드·경고를 **반환 결과로 올리는 것**까지를 요구하고 사용자 표시는 호출부 책임으로 규정하며, 이를 판정하는 AC 는 없다. 표시 배선은 미착수 상태다(§E.3 `open_items` 참조).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-22
run_commit_sha: pending-backfill-run-close
run_status: complete
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed  # 리드 지시로 push 및 PR 생성 없음 (worktree-local)
l44_post_push_fetch: not-performed   # 동일
new_warnings_or_lints_introduced: 0  # golangci-lint run --timeout=5m ./internal/template/... → "0 issues."
cross_platform_build:
  darwin_build: pass                 # go build ./... exit 0
  windows_build: pass                # GOOS=windows GOARCH=amd64 go build ./... exit 0
  darwin_vet: pass                   # go vet ./internal/... exit 0
  windows_vet: pass                  # GOOS=windows go vet ./internal/... exit 0
total_run_phase_files: 8
m1_to_mN_commit_strategy: milestone-per-commit  # M1 9c94c6b7a / M2 486a472b2 / M3 8094a5782 / M5 1272a2ef1 / M6 42c0c2167
coverage:
  internal_template: 86.1           # go test ./internal/template/... -cover
open_items:
  - "REQ-CSC-005 의 호출부 표시 배선(internal/cli) 미착수 — 배포기 반환 seam 까지만 구현. 대응 AC 없음."
  - "status draft → in-progress 전환이 M1 커밋이 아니라 run 마감 커밋에 실렸다 (아래 기록 참조)."
deviations:
  - "M1 은 미러 집합 파생 결정을 판정하는 테스트를 먼저 붉게 만든 뒤 구현했다(RED 출력은 위증 검사 표 첫 행과 같은 형태로 §E.2 에 기록). M2·M3 의 구현은 M1 과 같은 커밋 계열에서 함께 착지했으므로, 그 판정들의 근거는 사전 RED 가 아니라 **위증 검사**(지키는 대상을 부수고 붉어짐을 관측) 다. 표로 전부 남겼다."
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-22
sync_commit_sha: pending-backfill-SPEC-CODEX-SKILLS-CANONICAL-001
sync_status: complete
b12_self_test_a: pass    # grep -c 'SPEC-CODEX-SKILLS-CANONICAL-001' CHANGELOG.md → 0 (중복 없음)
b12_self_test_b: pass    # acceptance.md AC 표제 13개 = CHANGELOG 가 인용하는 AC 수와 정합 (13/13 PASS)
b12_self_test_c: pass    # CHANGELOG 가 이름을 든 경로 전부 실재: .claude/skills/, .agents/skills/, internal/template/templates/.gitignore
changelog_entry_position: "[Unreleased] → ### Added (신규 절)"
frontmatter_status_transitions:
  spec_md: in-progress → implemented
  plan_md: n/a          # frontmatter 없음
  acceptance_md: n/a    # frontmatter 없음
  progress_md: n/a      # frontmatter 없음
  completed_transition: deferred   # 리드 지시 — 병합 경로를 리드가 운전하므로 implemented 까지만
docs_changed: none
docs_decision: "README 4-locale · docs-site 무변경. 근거: 배포 트리를 열거하는 문서 페이지가 없고(`.agents/skills` 매치 0건), 미러 수명 관리(복사본 갱신·은퇴 미러 제거)가 승계 카드 소관이라 지금 문서화하면 반쯤 착지한 능력을 완결된 것처럼 기술하게 된다. CHANGELOG 가 이번 변경의 사용자 표면."
open_items_carried:
  - "REQ-CSC-005 호출부 표시 배선(internal/cli) 미착수 — 배포기 계층까지만 충족. CHANGELOG 에 사용자-가시 한계로 명시."
  - "폴백 복사본 고착 — 승계 카드 t173 소관. CHANGELOG 에 사용자-가시 한계로 명시."
```
