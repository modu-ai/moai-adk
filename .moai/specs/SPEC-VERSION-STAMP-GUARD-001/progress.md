# SPEC-VERSION-STAMP-GUARD-001 — 진행 기록

카드: t388 · Tier S · 워크트리 `WT-version-sync-list`

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- 기준선: `.moai/reports/t388/baseline.md` (트리 `9328a5242`)

### iter-1 (0.1.0)

- 요구 7 / 수락 7. 토큰 술어 가드 + 문서 수정을 한 카드로 묶었다.

### iter-2 (0.2.0) — plan-audit FAIL 대응

- 감사: `.moai/reports/t388/plan-audit.md` — **FAIL 0.77**(Tier S 임계 0.75를 넘겼으나 D2가
  critical/blocking)
- D1·D3·D4·D5-D10 정정. D2는 미해결 blocker로 기록.

### iter-3 (0.3.0) — 운영자 판정으로 카드 분할

- **토큰 술어 가드 축 전체가 카드 t392로 이관**(Tier M, 이 카드 착지에 의존). D2는 수리가
  아니라 **범위 이관으로 해소**됐고, 측정치(2,225줄 / 592파일 + 토큰 히스토그램)는
  `spec.md` §2에 보존했다.
- 삭제된 요구: 0.2.0의 REQ-VSG-004(목록 비의존 스윕) · REQ-VSG-005(토큰 술어) ·
  REQ-VSG-006(토큰 가드 비공허성). 딸린 §3 거부 목록 · §4 `eba919e44` RED 픽스처 ·
  `internal/versionstamp/` 패키지 · AC 3건도 함께 삭제(이월 아님).
- D1·D3는 §3과 함께 소멸. **D4는 살아남아 AC-VSG-001에서 닫았다**(리터럴 7경로가 판정자,
  `61921f1ba`는 출처 인용으로 강등). D7은 살아남은 검사에 재적용(합성 입력
  `docs-site/nonexistent-stamp.toml` 고정), D8은 REQ-VSG-004의 `Where` 절로 이월.
- 남은 산출: Go 파일 하나(`internal/cli/version_sync_list_test.go`) + 문서 하나.
- **회귀 보장은 절반**임을 `spec.md` §4 + REQ-VSG-006 + AC-VSG-006으로 명시. 누락 방향은
  t392 소관.
- 요구 6 / 수락 6 (Tier S 상한 8/8 이내).

### iter-4 (0.4.0) — plan-audit iter-2 FAIL 0.80 대응

- 감사: `.moai/reports/t388/plan-audit-iter2.md` — **FAIL 0.80**(Tier S 임계 0.75를 넘겼으나
  D1이 major/blocking). 결함 9건(D1 major/blocking · D2-D5 minor/blocking ·
  D6-D9 minor/optional).
- **D1 — 주 수리.** 검사의 RED 증거를 유령에 귀속시키던 서술을 걷어냈다. 앵커가 가리키는
  스탬프 소제목은 M2가 만들므로 M1 트리에서 파싱은 0건이고, 그때 우는 것은 **개수 단언**이지
  존재 단언이 아니다. 두 단언이 **각자의 RED을 따로 관측**하도록 재설계했다:
  E3-a(M1, `parsed=0 expected=7`) · E3-b(M2.1, 두 단언 동시 — 보조 관측) ·
  E3-c(M2.3, 치환으로 존재 단언 단독). 기대 RED 리터럴을 측정 **전에** `plan.md` §D의 단언
  메시지 계약으로 못박았다. 새 절 `spec.md` §5.1이 근거를 [HARD]로 적는다.
- **AC-VSG-004는 추가가 아니라 치환**으로 바뀌었다 — 한 줄을 더하면 개수가 8이 되어 원인이
  둘이 된다. `docs-site/hugo.toml` → `docs-site/nonexistent-stamp.toml` 치환은 개수를 7로
  유지해 존재 단언 단독의 빨강을 만든다. D6도 같은 수리로 함께 닫힌다.
- D5: REQ-VSG-005의 순환 비교를 「검사가 상수로 보유한 기대 개수 7」로 교체.
- D4: AC-VSG-006 3항을 판단에서 계기로 교체 — 양성 존재(`partial`, `does not detect`) +
  리터럴 거부 목록 5건, 둘 다 grep 판정.
- D2: `spec.md`의 걸린 `§3의 단위 고정 조항` 포인터 제거(사유를 인라인으로).
- **D3: 측정 정정.** 「이름에 버전 토큰이 든 파일이 둘」은 거짓이었다 — 아래 재측정 참조.
- D7: AC-VSG-006의 Given을 문서로 한정하고, SPEC 절반은 존재 판정(이미 성립)으로 커버리지 유지.
- D8: `항목 71-78행` → `71-74·77-78행`(75 공백, 76 라벨). D9: REQ-VSG-004의 `Where` 절을
  본절로 접었다.
- 요구 6 / 수락 6 불변. Tier S 상한 8/8 이내.

### iter-4 재측정 (전부 이 워크트리, 트리 `9328a5242`에서 직접 실행)

- 문서 소제목: `### Files Requiring Version Sync`(66) · `**Documentation Files:**`(70) ·
  `**Configuration Files:**`(76). **스탬프/산출물 축은 없다** — D1의 근거
- 목록 항목 행: 71-74(문서 라벨 아래 4건) · 77-78(설정 라벨 아래 2건). 75 공백, 76 라벨 — D8
- 유령은 78행, `**Configuration Files:**` 아래 — 앵커 한정이 읽지 않는 위치
- `test -e internal/cli/version_sync_list_test.go` → 종료 1 (AC-VSG-004/005의 RED-now)
- `test -e docs-site/hugo.toml` → 종료 0 (치환 대상이 실재함)
- **D3 재측정** — 거부 목록 범위에서 이름에 버전 토큰이 든 파일은 **8개**(2가 아니다),
  매치 줄 합 **113**:
  `docs/design/v2.14.0-release-plan.md` 40 · `.moai/release/RELEASE-NOTES-v2.17.0.md` 24 ·
  `.moai/release/MIGRATION-v2.17.0.md` 16 · `.moai/release/RELEASE-NOTES-v2.16.0.md` 12 ·
  `.moai/marketing/awesome-lists/github-release-v2.12.0-enhanced.md` 7 ·
  `.moai/release/RELEASE-NOTES-v2.15.0.md` 6 · `.moai/release/v2.15.0-draft.md` 4 ·
  `.moai/release/RELEASE-NOTES-v2.20.0.md` 4
  - 교차 검증: `-n` 출력에서 뽑은 전체 출현 **2607** − `-h` 기반 **2494** = **113**. 정확히 일치
  - `.moai/release/` 6개가 66줄을 차지한다(`grep -c '^\.moai/release/'` → 66). 이 거부 목록은
    `.moai/release-notes/`만 제외하고 `.moai/release/`는 제외하지 않는다 — t392가 밟을 자리
  - 토큰별: `v2.14.0` 72→112(+40) · `v2.12.0` 83→90(+7) · `v2.17.0` 25→65(+40, 두 파일 합산)
- 면적·히스토그램 재확인: **2225줄 / 592파일**, 출현 총 **2494**
  (v3.0.0 270 · v2.12.0 83 · v3.1.1 80 · v2.1.219 80 · v2.14.0 72 · v2.1.198 68) — 0.3.0과 동일
- AC-VSG-006 RED-now: 문서 전체에 `partial` / `does not detect` **0건**
- REQ-VSG-006 SPEC 절반: `grep -cF '이 카드가 착지해도 목록은 여전히 썩을 수 있다' spec.md` → 1

### iter-3 세션의 재측정 (전부 트리 `9328a5242`에서 직접 실행)

- 스탬프 집합: `61921f1ba` numstat → 7파일 9줄
- 유령: `test -e internal/template/templates/.moai/config/config.yaml` → 종료 1
- 유령이 목록에 존재: 해당 경로 grep → 1건, 종료 0
- 현재 라벨 축: `**Documentation Files:**`(70행) / `**Configuration Files:**`(76행)
- 릴리스 산출물 플레이스홀더 부재: `.moai/release-notes/vX.Y.Z.ko.md` 없음
  (실재 `v3.1.0.ko.md`·`v3.1.3.ko.md`)
- 토큰 술어 면적: **2225줄 / 592파일** (단위=줄)
- 토큰 히스토그램(단위=출현, 전체 2494): v3.0.0 270 · v2.12.0 83 · v3.1.1 80 · v2.1.219 80 ·
  v2.14.0 72 · v2.1.198 68
  - 정정: 초판은 `-n` 출력에서 토큰을 뽑아 **파일 이름 속 버전이 중복 계수**됐다
    (`v2.14.0` 112→72 — `docs/design/v2.14.0-release-plan.md` 매치 40줄이 경로에서 유입;
    `v2.12.0` 90→83 — `github-release-v2.12.0-enhanced.md` 매치 7줄). `-h`로 재측정해 교체.
    면적 수치는 영향 없음
- 도달 가능성: `61921f1ba` 조상 rc=1 · `eba919e44` 조상 rc=0
- `spec-lint.yml` fetch-depth 없음(rc=1) · `ci.yml` checkout 7개 중 6개 `fetch-depth: 0`
- 합성 입력 부재 확인: `test -e docs-site/nonexistent-stamp.toml` → 종료 1

### iter-4 (0.5.0) — plan-audit iter-3 PASS-WITH-DEBT 부채 정리

- 감사: `.moai/reports/t388/plan-audit-iter3.md` — **PASS-WITH-DEBT 0.90**(iter-1 0.77 →
  iter-2 0.80 → iter-3 0.90, 단조). run-phase 진입 적격 판정.
- F1(minor/blocking) 닫음 — `plan.md` M2.1에 「유령은 `**Version Stamps:**` 아래에 그대로
  둔다(따라서 8건)」 [HARD] 절 추가. `parsed=8` 핀이 이제 재현 가능하다.
- F2(minor/blocking) 닫음 — `plan.md` §B `항목 71-78행` → `항목 71-74·77-78행`.
  재측정: 항목은 71-74·77-78행, 75행 공백, 76행 `**Configuration Files:**` 라벨.
- F3(major/blocking) 닫음 — plan-phase 산출 4종이 `c6aed3c36`에 커밋됨(전반부). 후반부인
  §E.1 완료 신호를 이 절 하단에 기록(후술).
- F4(minor/optional) 닫음 — AC-VSG-006 3항의 grep 범위를 「Files Requiring Version Sync」
  절 전체로 **한 번만** 고정하고, 판정 명령을 실행 가능한 `sed`+`grep` 조합으로 교체.
- F5(minor/optional) 닫음 — `plan.md` §D에 **앵커 리터럴 계약** 신설.
  `**Version Stamps:**` / `**Release Artifacts:**`를 측정 전에 못박았다. M1·M2가 같은
  문자열을 참조하며 어느 쪽도 기억에서 재구성하지 않는다.
- F6(minor/optional) 닫음 — `acceptance.md` 문서 핀에서 브랜치 등가 주장(「측정 시점
  `origin/develop`과 동일」)을 제거하고 측정 트리 SHA 단일 핀으로 교체. 유효성 확인을
  브랜치 머리 비교가 아니라 대상 경로 diff로 바꿨다.
- F7 — 잔여 기록, 수리 불요.

**iter-4 실측**(2026-09-01, 판정 트리 = 워크트리 HEAD `c6aed3c36`):

- `origin/develop` = `2c18091d1`, `9328a5242`는 그 조상(rc=0). 감사 iter-3 시점의
  `5928095ea`에서 더 나아갔다.
- 측정 대상 8경로(`version-management.md` · `version.go` · `system.yaml` · `hugo.toml` ·
  `README.md` · `README.ko.md` · `Makefile` · `.goreleaser.yml`) diff: `9328a5242`↔`c6aed3c36`
  **빈 출력**, `9328a5242`↔`2c18091d1`도 **빈 출력**. RED-now 측정 전부 유효.
- 예외 1건: `CHANGELOG.md`는 `9328a5242`↔`2c18091d1`에서 **29줄 추가**됐다. 존재만 판정
  대상이고 내용은 어떤 기준도 읽지 않으므로 측정에 영향 없음 — 8경로 목록에서 제외한 이유다.
- AC-VSG-006 리터럴 재측정: (a) `partial`·`does not detect` → 절 범위 rc=1, 문서 전체 rc=1.
  (b) 거부 5리터럴 → 절 범위 rc=1, 문서 전체 rc=1. 두 범위가 같은 답을 준다.

상태: `in-progress` — Implementation Kickoff Approval 승인됨, run-phase 착지(§E.2/§E.3)

plan_status: audit-ready
plan_complete_at: 2026-09-01
plan_audit_verdict: PASS-WITH-DEBT 0.90 (iter-3, `.moai/reports/t388/plan-audit-iter3.md`)
plan_artifacts_commit: `c6aed3c36`

## §E.2 Run-phase Evidence

증거 전문: `.moai/reports/t388/run-evidence.md` (명령·출력·종료 코드 전량 + 로그 6건).
run-phase 진입 HEAD `6854a9306`, 착지 HEAD는 §E.3.

### AC 판정 매트릭스

| AC | 판정 명령 | Actual Output | Status |
|---|---|---|---|
| AC-VSG-001 | `grep -c '<유령 경로>' .moai/docs/version-management.md` | `0` (rc=1) | PASS |
| AC-VSG-001 | 스탬프 불릿에서 괄호 제거 후 정렬 | 리터럴 7경로와 일치, 양방향 차집합 공집합 | PASS |
| AC-VSG-002 | 스탬프 절에 `CHANGELOG.md\|release-notes` grep | `0` (rc=1) | PASS |
| AC-VSG-002 | 두 절 경로 집합 `comm -12` | 빈 출력 (교집합 공집합) | PASS |
| AC-VSG-003 (1) | `grep -nE 'reads from git tags at build time\|via .git describe.'` | rc=1 (파생값 단언 0건) | PASS |
| AC-VSG-003 (2) | `grep -n 'Makefile:20\|goreleaser.yml:22\|fallback'` | 8·12·13·80행 — 폴백 서술 + 주입 지점 2건 인용 | PASS |
| AC-VSG-003 (3) | `grep -niE 'constant\|상수'` | rc=1 | PASS |
| AC-VSG-004 | E3-c (치환 → 실패 → 되돌림 → 통과) | 아래 E3-c | PASS |
| AC-VSG-005 | E3-a (`parsed=0`) → E4 (7에서 통과) | 아래 E3-a / E4 | PASS |
| AC-VSG-006 (a) | 절 범위에 `partial` / `does not detect` grep | `1` / `1` | PASS |
| AC-VSG-006 (2) | 절 범위에 `t392` grep | `1` | PASS |
| AC-VSG-006 (b) | 거부 5리터럴 `grep -cF` | `0` (rc=1) | PASS |

### 단언별 RED 관측 — 어느 빨강도 다른 단언의 증거로 겸용되지 않음

| 관측 | 트리 | 운 단언 | 원인 | 소용 |
|---|---|---|---|---|
| E3-a | `6854a9306` + 검사 (= `f270d2df5`) | 개수 단독 — `version-stamp entries: parsed=0 expected=7` | **앵커 소제목 부재**로 파싱 0건. 유령 아님 — 존재 단언은 이 실행에서 아무 경로도 출력하지 않았고(`grep -c "does not exist"` → 0, rc=1), 유령은 앵커가 읽지 않는 `**Configuration Files:**` 아래에 있었다 | AC-VSG-005 RED |
| E3-b | `d595faa9d` | 둘 다 — `parsed=8 expected=7` + 유령 경로 이름 지목 | **둘**: 항목이 하나 많고, 그중 하나가 유령 | 보조 관측. 어느 AC의 단일 증거도 아님 |
| E4 | `d7af3d22d` | 없음 (통과) | 항목 7 전부 실재. 통과는 출력이 없으므로 불릿 수 별도 계수(`7`)와 E3-c의 개수 단언 침묵으로 교차 확인 | AC-VSG-005 GREEN |
| E3-c | `d7af3d22d` + 치환(미커밋) | 존재 단독 — `version-sync list names a path that does not exist: docs-site/nonexistent-stamp.toml` | **하나**: 심은 경로. `grep -c "parsed=" m2-3-red.log` → `0` (rc=1)로 치환이 추가가 아님을 기계 확인 | AC-VSG-004 RED → GREEN |

되돌림은 눈이 아니라 `git status --short` 빈 출력(= `d7af3d22d`와 바이트 동일)으로 증명했다.

### 검증

| 항목 | 명령 | 결과 |
|---|---|---|
| 패키지 스위트 | `go test ./internal/cli/...` | exit 0, 17개 패키지 ok (`final-test.log`) |
| vet | `go vet ./internal/cli/...` | exit 0 (`final-vet.log`) |
| 포맷 | `gofmt -l internal/cli/version_sync_list_test.go` | 빈 출력, exit 0 |
| 새 패키지 | `git diff --name-only --diff-filter=A 6854a9306 HEAD -- internal/` | `internal/cli/version_sync_list_test.go` 1건 — 기존 패키지, 새 디렉터리 0 |
| 새 CI job · 템플릿 미러 | `git diff --name-only 6854a9306 HEAD -- .github/ internal/template/ .claude/` | 빈 출력 — Template-First 비해당을 가정이 아니라 측정으로 확인 |

로컬 전체 스위트(`go test ./...`)는 돌리지 않았다 — 전 패키지 판정은 CI 몫(`CLAUDE.local.md` §4).

### §7 잔여 위험 갱신

- **R-4는 실제로 발생했고 잡혔다.** M1 실행이 정확히 R-4가 경고한 모양(파싱 0건)이며, 개수
  단언이 그것을 통과가 아니라 실패로 만들었다.
- **R-5는 그대로 열려 있다.** 목록이 정당하게 8건이 되면 `expectedVersionStampEntries`를 사람이
  함께 고쳐야 하고, 그 짝맞춤을 강제하는 것은 없다.
- **닫지 못한 절반은 불변.** 누락 방향은 여전히 보이지 않으며 카드 t392 소관이다(§4).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_entry_head: 6854a9306
run_commit_sha: c161c8965   # M2.3 증거 + 이 기록. 아래 backfill 커밋이 최종 착지
milestone_commits:
  m1: f270d2df5   # 검사 착지 (개수 단언 RED, E3-a) + status draft -> in-progress
  m2_1: d595faa9d # 소제목 신설 + 누락 4건, 유령 잔존 (의도적 RED, E3-b)
  m2_2: d7af3d22d # 유령 제거 (GREEN, E4)
run_status: implemented
ac_pass_count: 6
ac_fail_count: 0
red_observations: 3   # E3-a (개수 단독) / E3-b (보조, 원인 둘) / E3-c (존재 단독)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  note: 순수 Go 테스트 파일 1건 + 마크다운. OS 분기 코드 없음. 매트릭스 판정은 CI.
total_run_phase_files: 8   # Go 1 + 문서 1 + SPEC 2 + 증거 4(+로그)
m1_to_mN_commit_strategy: 마일스톤당 1커밋. M2.1은 의도적 적색으로 커밋한다 — 한 커밋에 접으면
  검사가 실제 유령을 만나는 유일한 관측(E3-b)이 사라진다(`plan.md` M2).
push_state: not pushed — 통합은 리드가 별도로 잡는다
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-01
sync_entry_head: 62ea21945
sync_commit_sha: "<backfilled in the immediately following commit — a commit cannot cite its own SHA>"
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-VERSION-STAMP-GUARD-001' CHANGELOG.md -> 0 (rc=1) before emission; no duplicate entry from a parallel session"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u -> 6 distinct (AC-VSG-001..006), non-zero; the CHANGELOG entry claims no AC count, and §E.2 reports 6/6 PASS"
b12_self_test_c: "every path named in the CHANGELOG entry resolved with ls: spec.md, .moai/docs/version-management.md, internal/cli/version_sync_list_test.go, internal/template/templates/.moai/config/sections/system.yaml.tmpl, Makefile, .goreleaser.yml, cmd/moai/main.go, README.ja.md, README.zh.md, docs-site/hugo.toml, pkg/version/version.go. The one path the entry asserts is ABSENT was checked in that direction: ls internal/template/templates/.moai/config/config.yaml -> rc=1. Line citations re-read in this tree: .goreleaser.yml:22 and Makefile:20 both carry the -X pkg/version.Version injection"
changelog_entry_position: "[Unreleased] -> ### Fixed, appended as the last bullet (line 403 in the post-edit file). Placed under Fixed rather than Added because the deliverable is a documentation-defect repair; the guard test exists to hold that repair, not as a feature"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged onto this single sync commit per the 3-phase close); updated: 2026-09-01"
  plan_md: "no frontmatter — not applicable"
  acceptance_md: "no frontmatter — not applicable"
  progress_md: "no frontmatter — not applicable"
  note: "only spec.md carries frontmatter in this SPEC. That is the intended shape, not an omission: SPEC Lint reports a status: line on a non-spec.md artifact as an error (card t369)"
canary_compliance_check: "not applicable — this SPEC defines no forward-looking policy that its own sync would have to test"
sync_verification:
  guard_test: "go test ./internal/cli/ -run VersionSyncList -> PASS (exit 0), re-run in this tree at HEAD 62ea21945"
  package_suite: "go test ./internal/cli/... -> exit 0, all packages ok (.moai/reports/t388/sync-test.log)"
  spec_lint: "moai spec lint (whole catalogue) -> 8 error(s), 64 warning(s), NONE of them naming this SPEC (grep -c 'VERSION-STAMP-GUARD' over the ERROR lines -> 0; the SPEC does not appear anywhere in the output). The 8 are inherited from other SPECs and are not this card's to close (.moai/reports/t388/sync-spec-lint.log)"
  full_suite: "not run locally — per-package judgment is CI's (CLAUDE.local.md §4)"
  non_vacuity_remeasured: "the existence assertion was re-observed failing in THIS phase, not carried over from run-phase: substituting the listed docs-site/hugo.toml for docs-site/NOPE-sync-mutant.toml made the check exit 1 with 'version-sync list names a path that does not exist: docs-site/NOPE-sync-mutant.toml' and NO parsed= line (grep -c 'parsed=' -> 0), so the count stayed at 7 and the existence assertion fired alone (.moai/reports/t388/sync-mutation-red.log). Reverted; git status --short showed no entry for the document, i.e. byte-identical; green again"
half_guarantee_wording_intact:
  spec_md: "§4 [HARD] '이 카드가 착지해도 목록은 여전히 썩을 수 있다' + the two-direction table; 13 mentions of t392, 3 of 절반"
  documentation: ".moai/docs/version-management.md — 'The guarantee it establishes is partial' + 'does not detect' + 'Closing that direction is card t392' (grep -c partial -> 1, does not detect -> 1, t392 -> 1)"
  guard_test_header: "internal/cli/version_sync_list_test.go — 'PARTIAL' + the CAUGHT / NOT CAUGHT pair + t392 (grep -c PARTIAL -> 1, NOT CAUGHT -> 1, t392 -> 1)"
  changelog_entry: "the last sub-bullet is the half-guarantee statement; it names the omission direction as the one that caused the card and closes with 'Nothing here should be read as the list no longer being able to rot'"
push_state: "not pushed — integration is scheduled by the lead"
```

### Sync-phase finding — out of scope, reported not repaired

`.claude/skills/hns-moaiadk-dev-reference/SKILL.md` §Version Management carries a **second copy** of
the same material, and that copy still says what this card corrected: `pkg/version/version.go` reads
from git tags at build time (18·22행), a stamp list that omits `README.ja.md` / `README.zh.md` /
`docs-site/hugo.toml` / `pkg/version/version.go`, `CHANGELOG.md` filed as a stamp, and
`internal/template/templates/.moai/config/sections/system.yaml.tmpl` named as a stamp — which §1.2
establishes it is not. Its Release Process section is also the retired manual `git tag` +
`make release` flow that `.moai/docs/version-management.md` explicitly marks as NOT the release path.

This drift **predates this card** and reaches beyond it, so it was not repaired here — a partial fix
would leave the section looking freshly reviewed while half of it stayed wrong. The guard test does
not read this file, so the copy is unguarded either way. Recommended follow-up: replace the
duplicated section with a pointer to `.moai/docs/version-management.md` as its single source, on its
own card.
