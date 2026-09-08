# SPEC-DRIFT-CLOSE-BODY-001 — 진행 기록

plan_status: audited (PASS-WITH-DEBT 0.80) — run 진입 승인 2026-09-03

## §E.1 Plan-phase Audit-Ready Signal

- 카드: t410 (Class C) · 브랜치 `WT-drift-false-positive` · 트리 `.claude/worktrees/t410`
- Tier: S (spec.md + plan.md; AC는 spec.md §3 인라인). 근거는 `plan.md` §A Tier S 판정 근거
- REQ 7 / AC 7 — Tier S 상한(각 8) 안
- SPEC-ID 정규식 자체 점검: `[[ "SPEC-DRIFT-CLOSE-BODY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`. `.moai/specs/` 충돌 없음
- 조사 원장: `.moai/reports/t410/discovery.md` + `r1-walker-trace.log` · `r2-blast-radius.log` · `r3-drift-row.log`
- 수리 후보 판정: **B 채택**(close 선언 커밋에 한해 본문 조회), A 기각(측정된 위험), C 기각·D 범위 밖 — `spec.md` §5
- 미확정으로 남긴 것: 파급 건수. 조사의 LOOSE 33 / TIGHT 15는 경계값이며 TIGHT에 알려진 오탐 1건 포함. 실제 수는 run-phase M3 전수 대조의 **결과**로 정해진다(AC-DCB-005)
- plan-audit: **PASS-WITH-DEBT · 0.80** (Tier S PASS 임계 0.75, must-pass 7/7, iteration 1/1). 판정본 `.moai/reports/t410/plan-audit-verdict.md`. 남은 부채: blocking 3건은 아래 §E.1.1대로 상환됨, D4·D6은 run-phase 원장 처리로 이월

### §E.1.1 감사 고정본 이후 편집 델타 실측 (리드 [HARD], 2026-09-03)

**Claim** — 감사 고정 이후의 spec.md 편집은 감사가 본 AC/REQ 집합을 바꾸지 않는다. 판정서 결함 상환에 국한된다.

**Evidence**

- 해시 대조: 고정본 `ff4489f3140aaf514d2f9cbc3e4bd6e39c81a12f5c6b07ccb69364d03f46ae96` (판정서 pin, 2026-09-03T06:27:42Z) vs 현재본 `shasum -a 256` → `0bf04b1d05522e1ab30271b2738a830122f7bd79504c99850fc25bad2e926fca` — **다름**. `plan.md`은 `9b69231a…`로 **고정본과 동일**(편집 없음)
- 집합 불변: `grep -c '^\*\*REQ-DCB-00'` → `7`; REQ id `001~007` 연속; AC id `001~007` — 판정서 MP-1 측정과 동일
- 좌표 불변: 판정서 인용 `spec.md:2`(프론트매터 12필드+`tier: S`), `spec.md:84~96`(GEARS 블록, REQ-001~007 형식 동일) — 현재본에서 같은 줄에 그대로
- 델타 위치: 판정서 인용 `spec.md:115`(AC-DCB-002 Given) → 현재 116, `spec.md:164`(AC-DCB-007 술어) → 현재 172. 삽입 +1 / +7, 전부 §3 AC 본문 안
- 델타 방향: spec.md HISTORY `0.3.0` 행이 자기기술 — "plan-audit(t410, PASS-WITH-DEBT 0.80) 차단 결함 3건 + 문서 내 모순 2건 상환". D1 술어 `^[-+].*^status:` → `^[-+]status:` 교체(현재본 AC-DCB-007에서 확인), D2 모양-B 자격+전수 훑기 명시·픽스처 10→12줄, D3은 **AC 신설 없이** AC-DCB-003 (d) 케이스로 병합(AC 7개 유지), D5 §5 표·§5.2 정정, D7 AC-DCB-002 Given 보강. `version:` 0.1.0→0.3.0 정합 수리 동반

**Baseline-attribution** — 전부 이 트리(`.claude/worktrees/t410`), origin/develop `7835148d3` + 로컬 develop `6765a75c0` 흡수 후 HEAD `460a7e16d`에서 실측. `git status --porcelain` 0행

**Gaps** — 리드가 지정한 `git diff <감사시점 SHA>..HEAD -- spec.md` 는 **성립하지 않는다**. 고정본은 커밋된 적이 없는 워킹트리 판본이고(`git show HEAD:…spec.md | shasum` → `0bf04b1d…` = 현재본), 감사 시각의 트리 HEAD `4e4607abe`에는 이 SPEC 파일 자체가 없다. 따라서 델타 **원문**은 재구성 불가이며, 위 증거는 원문 diff가 아니라 등가 검사(집합·좌표·자기기술 HISTORY)다

**Residual-risk** — 등가 검사는 AC/REQ의 **개수·id·좌표**와 자기기술 기록을 확인할 뿐, 감사가 읽은 AC **본문 문장**이 그 밖에서 바뀌지 않았음을 증명하지 못한다. 판정서가 좌표를 인용한 3개 AC(002·006·007)는 직접 읽어 대조했고 나머지 4개는 대조하지 않았다

**판정** — 델타가 AC·REQ 집합을 건드리지 않았으므로 리드가 건 정지 조건(집합 변경 시 재감사 요청)은 불성립. run 진행

## §E.2 Run-phase Evidence

원장 정본: `.moai/reports/t410/run-evidence.md` (M1~M4 각 마일스톤 경계에서 5절 형식). 이 표는 AC별 요약이며, 축자 출력은 원장에 있다.

측정 좌표: 트리 `.claude/worktrees/t410` · 브랜치 `WT-drift-false-positive` · 판정 브랜치 `main` = `7ad9f8534` · 코퍼스 측정 시점 HEAD `c7126f526` · 바이너리는 이 트리에서 빌드(`/tmp/moai-t410`, `/tmp/moai-t410-pre`), 설치본 미사용.

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-DCB-001 | **PASS** | `go test ./internal/spec/ -run TestDriftCloseBody -count=1` | `--- PASS: TestDriftCloseBody_BodyDeclaredCloseRecognized (2.21s)`. RED 2단계 실측: RED-1 `undefined: bodyDeclaresClose` rc 1, RED-2 `GitImpliedStatus = "in-progress", want "completed"` rc 1. 공허 방지 선행 단언 `--- PASS: TestDriftCloseBody_PrimaryWalkYieldsInProgress` |
| AC-DCB-002 | **PASS** | 같은 명령 | `--- PASS: TestDriftCloseBody_MentionIsNotClose` + `--- PASS: TestDriftCloseBody_PredicateTable` (실측 12줄 전부). 뮤테이션 의무 3종 + 발견 2종 = **5종 전부 죽음**(원장 M2·M3 표) |
| AC-DCB-003 | **PASS** | 같은 명령 | (a)(b)(c) `--- PASS: TestDriftCloseBody_FallbackOnlyInputGate` 3 subtest. (d) `--- PASS: TestDriftCloseBody_OutputIsCompletedOrNothing` — 1차 워크 값 `in-progress` 유지. 공허 방지 선행 단언 `_DeltaPrimaryWalkPrecondition` PASS |
| AC-DCB-004 | **PASS** | `/tmp/moai-t410 spec drift --no-cache \| grep SPEC-V3R6-SESSION-HANDOFF-AUTO-001` | 전 `completed in-progress DRIFT` → 후 `completed completed aligned`. **행이 존재하면서** DRIFT가 아니다. 양쪽 모두 이 트리에서 재측정 |
| AC-DCB-005 | **PASS** | `diff <(sort drift-before-remeasured.txt) <(sort drift-after.txt)` | ① 비-DRIFT → DRIFT **0건**. ② DRIFT → 해제 **18행 전수**를 커밋 SHA·subject·본문 줄과 함께 원장 §M3 행별 판정표에 기록. ③ 18행 전부 사람 판정 = close 선언. `SPEC-AUTONOMY-TIERS-001` 미해제 확인. Summary 196 → 178 |
| AC-DCB-006 | **PASS** | `go test ./internal/spec/... -count=1 -v` | rc 1 — 유일한 FAIL은 상속 적색 `TestCatalogHashParity`(템플릿 해시, `4244c4a06` 귀속, 이 카드 변경 0). 명명된 가드 전부 실행·통과: `TestGetGitImpliedStatus_ChoreSkip` · `TestGetGitImpliedStatus_SPECIDWordBoundary` · `TestCombinedScopeCloseMatches` · `TestDetectDrift_CombinedScope{Fallback,CollisionGuard,NoCloseInfixNoFallback}` · `TestDetectDrift_Characterization_*` · `TestDetectDrift_ConstantGitLogInvocations` (셀렉터 0매치 아님을 `-v` 이름으로 확인) |
| AC-DCB-007 | **PASS** | `git diff -U0 -- .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md \| grep -cE '^[-+]status:'` | 세 실행: 초록 `0` → 붉음 `2`(`-status: completed` / `+status: implemented`) → 되돌린 초록 `0`. 대조군 `grep -cE '^[-+]'` = `7`(빈 diff의 0이 아님). 변경은 HISTORY 1행 + `version:` + `updated:`. `grep -c 'r1-walker-trace'` = `1` |

**범위 밖 불변 확인**: 1차 워크(`inMemImpliedStatus` 2단 구조)와 기존 combined-scope fallback 미변경 — `TestDriftCloseBody_ExistingFallbackStillFirst` PASS + 기존 combined-scope 3건 PASS. 새 `git log` 서브프로세스 0개(공유 인덱스 재사용) — `TestDetectDrift_ConstantGitLogInvocations` PASS. 본문 조회 출력은 `completed` 또는 무판정 — AC-DCB-003 (d)로 측정.

**계획 대비 델타**: M3 전수 대조에서 §5.3의 12줄 픽스처 **밖의 반례 1건**(`ac3e38a0b`의 amendment 기록)이 나와, §5.3이 정한 대로 술어를 **좁혔다**(모양 A에 한해 커밋 subject의 close 신호 요구, `subjectCloseSignal`). 해제 19 → 18. §5.4 조건 3(실측 close subject 전부 통과)을 측정으로 확인했다. 상세는 원장 §M3.

**이월 부채**: 판정서 D4(REQ-DCB-002 문면이 오류 경로까지 주장) · D6(Tier 파일 모집단·초과 동작) 둘 다 **문서 층 결함**으로 판정하고 원장에 기록했다. 수리에는 `spec.md`/`plan.md` **본문** 편집이 필요한데 manager-develop에게 금지된 표면이므로, 구현을 막지 않는 한 blocker를 올리지 않고 기록으로 남긴다. 상세는 원장 「이월 부채 처리」.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: pending-backfill
run_status: audit-ready
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed
l44_post_push_fetch: not-performed
new_warnings_or_lints_introduced: none-measured
cross_platform_build: not-measured
total_run_phase_files: 10
m1_to_mN_commit_strategy: per-milestone
```

필드 주석 — 값이 무엇을 말하고 무엇을 말하지 않는가:

- `run_commit_sha: pending-backfill` — 커밋은 자기 해시를 인용할 수 없다. sync-phase 이후 후속 커밋이 채운다(`spec-frontmatter-schema.md` § SHA placeholder backfill exemption).
- `ac_pass_count: 7` / `ac_fail_count: 0` — AC-DCB-001~007 전부 PASS. PASS-WITH-DEBT는 없다.
- `preserve_list_post_run_count: 0` — plan.md §D의 범위 밖 다섯 항목을 건드리지 않았다. 변경 파일은 전부 `internal/spec/` + 이 SPEC의 아티팩트 + t410 원장 + `SPEC-ERA-H3-NARROWING-001`의 HISTORY(AC-DCB-007이 요구).
- `l44_*_fetch: not-performed` — 이 카드는 워크트리 격리 안에서 작업했고 push하지 않는다(레인은 develop push를 하지 않는다). fetch/divergence 판정은 병합 창에서 수행한다.
- `new_warnings_or_lints_introduced: none-measured` — `go vet ./internal/spec/` rc 0, `gofmt -l internal/spec/` 무출력. **`golangci-lint`는 돌리지 않았다** — 측정하지 않은 것을 "없음"으로 적지 않기 위해 `none`이 아니라 `none-measured`다.
- `cross_platform_build: not-measured` — `GOOS=windows` 빌드를 재지 않았다. 변경이 순수 Go 문자열/정규식이고 syscall·빌드태그를 건드리지 않아 크로스플랫폼 축이 무관하다고 판단했으나, 판단이지 측정이 아니다.
- `total_run_phase_files: 10` — 소스 3(`drift.go`, `drift_index.go`, `drift_close_body_test.go`) + SPEC 아티팩트 3(이 SPEC의 spec.md·progress.md, t382 spec.md) + 원장·측정 산출물 4. Tier S 파일 상한(`< 5`) 판정의 모집단은 **소스 3개**이며 그 근거는 원장 D6 항목에 있다.
- 전체 스위트(`go test ./...`)는 **의도적으로 돌리지 않았다** — 병렬 레인이 함께 돌려 머신을 마비시킨 전례(2026-08-15) 때문이며, 전 패키지 판정은 CI 몫이다. 재실행 범위는 파일 델타 패키지 ∪ 그것을 import 하는 패키지로 잡았다.

## §E.4 Sync-phase Audit-Ready Signal

**Claim** — 이 SPEC의 문서 표면은 CHANGELOG 항목 1건뿐이다. `docs-site`·`README`·`.moai/docs/`에 drift/close 규약을 서술하는 사용자 표면은 없으며, frontmatter 전이는 `spec.md`에 한정된다(Tier S — `plan.md`는 frontmatter 없음).

**Evidence**

- `grep -c 'SPEC-DRIFT-CLOSE-BODY-001' CHANGELOG.md` (편집 전) → `0` — 중복 없음, B12 self-test 1 통과.
- AC 카운터: `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md | sort -u` → `AC-DCB-001`..`AC-DCB-007` (7개, `AC-LSCSK-003`은 이 SPEC 소유가 아닌 인접 인용문 안 문자열이라 제외). `progress.md` §E.2의 `ac_pass_count: 7`과 일치.
- 파일 경로 검증: CHANGELOG 항목이 인용한 7개 경로(`internal/spec/drift.go`, `internal/spec/drift_index.go`, `internal/spec/drift_close_body_test.go`, `spec.md`, `run-evidence.md`, `plan-audit-verdict.md`, `SPEC-ERA-H3-NARROWING-001/spec.md`) 전부 `ls`로 존재 확인 — B12 self-test 3 통과.
- 문서 표면 조사: `grep -rln "close.*convention\|conventional-commit.*close\|drift.*close\|closeInfixMatch\|GitImpliedStatus" --include="*.md" .moai/docs/ docs-site/ README*` → 2건. 둘 다 직접 읽어 대조: `docs-site/content/en/getting-started/windows-guide.md:9`는 "closes that gap"(WSL이 경로 차이를 없앤다는 뜻, drift와 무관), `docs-site/content/en/core-concepts/verification-claim-integrity.md`의 6건은 전부 `verification-claim-integrity.md`의 "SPEC이 close debt다" 같은 예시 문구(이 SPEC이 아니라 VCI §1.1 surface 3의 일반 예시) — 둘 다 오탐. **drift/close 규약을 서술하는 사용자 문서 표면은 없다.**
- `internal/spec/CLAUDE.md` 존재 확인: `ls internal/spec/CLAUDE.md` → not found. 이 패키지에 로컬 CLAUDE.md가 없으므로 갱신 대상도 없다.

**Baseline-attribution** — 전부 이 워크트리(`.claude/worktrees/t410`, 브랜치 `WT-drift-false-positive`)에서, run-phase 종결 HEAD `2954db755`를 기준으로 측정. `git status --porcelain` 확인은 이 절 작성 직전 실행.

**Gaps** — `internal/spec/*.md` 이외의 코드 주석에 close-subject 규약을 설명하는 부분이 있는지는 grep 패턴 범위 밖이라 전수 스캔하지 않았다(대상은 `.md` 파일로 한정). `spec-frontmatter-schema.md`(항상 로드되는 규칙)의 "Close-subject full-ID mandate" 절은 이 SPEC이 다루는 subject 판정과 인접하지만 본문 조회는 다루지 않으므로 갱신 대상이 아니라고 판단했다 — 판단이지 규칙 전문을 재독해 확정한 것은 아니다.

**Residual-risk** — CHANGELOG 항목은 이 SPEC 저자가 작성한 요약이며, 별도 리뷰어가 원문 대비 과장·누락을 검증하지 않았다. 18행 해제 목록 자체(SHA·subject·본문 줄)는 run-evidence.md 원장에만 있고 CHANGELOG에는 요약 수치(196→178, 18/0)만 반영했다 — 상세 대조가 필요하면 원장을 봐야 한다.

```yaml
sync_complete_at: 2026-09-03
sync_commit_sha: "c1a389036"   # docs(SPEC-DRIFT-CLOSE-BODY-001): sync-phase artifacts — backfilled here; a commit cannot cite its own SHA
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-DRIFT-CLOSE-BODY-001' CHANGELOG.md (pre-emission) -> 0, no duplicate from a parallel session"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' spec.md | sort -u -> 7 distinct AC-DCB-* ids, matching progress.md §E.2 ac_pass_count: 7"
b12_self_test_c: "ls on every path cited in the CHANGELOG entry -> all 7 present (3 source, spec.md, 2 report artifacts, 1 sibling SPEC spec.md)"
changelog_entry_position: "CHANGELOG.md [Unreleased] -> ### Fixed, first entry (immediately after the section header)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (merged 3-phase close on this sync commit); updated: 2026-09-03 (unchanged date, same-day close)"
  plan_md: "no status field (Tier S plan.md carries no frontmatter status axis per spec-frontmatter-schema.md § Artifact Statelessness)"
  progress_md: "not transitioned - progress.md carries no status: field per the schema (body sections only)"
canary_compliance_check:
  applicable: false
  reason: "this SPEC fixes a drift-detector defect; it does not define a forward-looking convention with a self-consuming first-run test"
docs_sync: "no user-facing doc surface describes the drift close-body convention. Scanned: README*.md, .moai/docs/, docs-site/ -> 2 pattern hits, both read by hand and confirmed unrelated (windows-guide.md WSL prose; verification-claim-integrity.md generic defect-claim examples). CHANGELOG.md is the only surface touched"
tests:
  affected_packages: "internal/spec + 12 dependent packages (measured, see §E.4.1); this sync touched no source, only docs + frontmatter"
  full_suite: "NOT RUN locally per instruction; CI owns the full-suite verdict on push"
push_state: "not pushed, not merged - lead pushes in batch and performs the develop merge later, per dispatch"
```

**이월 부채 확인 — D4·D6은 이 sync가 건드리지 않는다.** run-phase 원장이 기록한 두 문서-층 결함(D4: REQ-DCB-002가 `inMemImpliedStatus` 오류 경로까지 주장; D6: Tier 파일 모집단·초과 시 동작)은 `spec.md`/`plan.md` **본문** 편집이 필요하고, manager-docs에게는 금지된 표면이다. 구현을 막지 않으므로 blocker를 올리지 않고 기록으로만 남긴다 — 필요하면 후속 카드가 manager-spec에게 재위임한다.

### §E.4.1 정정 — 의존 패키지 수치 재측정 (sync-audit F2 상환)

**Claim** — §E.4 초판의 `31 dependent packages`는 귀속이 없었을 뿐 아니라 **값이 틀렸다**. 실제는 12개이며, 그중 가장 무거운 `internal/cli`는 첫 배치 실행에서 판정이 나오지 않았다(초록도 적색도 아님).

**Evidence**

- 모집단 산출: `go list -f '{{.ImportPath}} {{join .Deps " "}}' ./... | grep -E ' [^ ]*/internal/spec( |$)' | awk '{print $1}' | wc -l` → `11` (전이 의존). 테스트 전용 임포터는 `.TestImports`/`.XTestImports`로 같은 방식 → `4`. 합집합 `sort -u` → **`12`**
- 12개 전수 실행: `go test ./cmd/moai/... ./internal/cli/... ./internal/codexadapter/... ./internal/codexwiring/... ./internal/epic/... ./internal/feedback/... ./internal/harness/router/... ./internal/hook/... ./internal/migration/migrations/... ./internal/permission/... ./internal/spec/... ./internal/web/... -count=1` → 원장 `.moai/reports/t410/dependents-retest.log`. `grep -c '^ok'` → `35`. 이상 2건:
  - `internal/spec` — `--- FAIL: TestCatalogHashParity`. 상속 적색(카드 diff의 template 파일 0개, `4244c4a06`이 기준선의 조상). 이 카드 소관 아님
  - `internal/cli` — `panic: test timed out after 10m0s` (기본 한도). **시간 초과는 실패가 아니라 미판정이다**
- `internal/cli` 재판정: `go test ./internal/cli/... -count=1 -timeout 30m` → `rc=0`, 원장 `.moai/reports/t410/cli-retest.log`, `grep -c '^ok'` → `17`, 비-ok 패키지 줄 0행

**Baseline-attribution** — 전부 이 워크트리(`.claude/worktrees/t410`, 브랜치 `WT-drift-false-positive`), HEAD `3acc8569a`에서 오케스트레이터가 직접 실측. 감사자의 재측정을 인용하지 않고 다시 쟀다.

**Gaps** — 12개 밖 패키지는 돌리지 않았다(전체 스위트는 부하 규율상 의도적 미실행 — 판정은 CI 몫). `internal/cli` 재실행은 30분 한도에서 한 번뿐이라 부하에 따른 재현성은 관측하지 않았다.

**Residual-risk** — `internal/cli`의 첫 미판정이 순수한 시간 초과인지, 이 카드와 무관한 지연 회귀인지는 구분하지 않았다. 30분 한도 재실행이 초록이므로 전자로 본다.

### §E.4.2 재종결 — 0.4.0 amendment (카드 t484)

**재종결 배경** — 이 SPEC은 `status: completed`로 종결된 뒤 카드 t484에서 운영자 승인을 받아 제자리 수정(in-place amendment)을 거쳤다. frontmatter는 `amendment_of: self` + `version: 0.4.0`이며, 수정 사유·범위·간극 조정은 `spec.md` HISTORY `## Amendments` 절에 기록돼 있다(판정서 `.moai/reports/t484/verdict.md`). amendment는 `completed → in-progress` 전환을 만들었으므로, 이 sync 커밋이 소유한 `in-progress → implemented → completed` 전환(`spec-frontmatter-schema.md` § Status Transition Ownership Matrix)으로 SPEC을 다시 `completed`로 닫는다. `version:`은 0.4.0에 그대로 둔다 — 버전은 amendment가 이미 소유했고 재종결은 버전을 올리지 않는다.

- 이 SPEC은 0 소스 파일 변경(문서 전용)이라 CHANGELOG 항목을 만들지 않는다 — develop의 t481 종결이 세운 선례를 따른다
- `updated:`는 2026-09-05(amendment일) 그대로다

```yaml
sync_commit_sha: "76631690b"
```

커밋은 자기 해시를 인용할 수 없으므로 자리표시자를 두고, 후속 backfill 커밋(`chore(SPEC-DRIFT-CLOSE-BODY-001): backfill sync_commit_sha …`)이 실제 SHA로 채운다. §E.4 초판의 `run_commit_sha: pending-backfill`(62·77행)은 run-phase 기록이라 이번 backfill 대상이 아니다.
