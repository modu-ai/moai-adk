# progress — SPEC-ZONE-REGISTRY-HARDEN-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-25
- artifacts: spec.md · plan.md · acceptance.md (Tier M 3종 + 본 progress.md §E 스켈레톤)
- baseline_tree: db1362739 (worktree t268, card t268)
- tier: M · reqs: 8 (REQ-ZRH-001..008) · acs: 9 (AC-ZRH-001..009)
- revised: 0.2.0 — plan-audit iter1 (PASS-WITH-DEBT 0.825) defects D1/D2/D3 applied, 2026-08-25

## §E.2 Run-phase Evidence

> 좌표: plan 재좌표 baseline `db1362739`(실행 트리 HEAD `a739d04b4`의 조상 — `git merge-base --is-ancestor db1362739 HEAD` 관측). 모든 Evidence는 이번 실행·이 트리에서 관측한 원문(VCI §2).

### M1 — F1 rewrap + clause 재선택

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-ZRH-001 | PASS | `grep -c 'clause: "The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report."' .claude/rules/moai/core/zone-registry.md` | `1` |
| AC-ZRH-001 | PASS | 동일 grep · template mirror | `1` |
| AC-ZRH-001 | PASS | `grep -c -F 'The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.' .claude/rules/moai/workflow/ci-autofix-protocol.md` | `1` |
| AC-ZRH-001 | PASS | `grep -c 'test assertion failure) MUST"' .claude/rules/moai/core/zone-registry.md` (절단 clause 잔존) | `0` |
| AC-ZRH-002 | PASS | `git show db1362739:<twin> \| tr '\n' ' ' \| tr -s ' '` vs 현재 트윈 `cmp` — 배포판 | `rc=0` |
| AC-ZRH-002 | PASS | 동일 절차 — 템플릿 원본 | `rc=0` |
| AC-ZRH-003 | PASS | `cmp` 트윈 쌍 / 미러 쌍 | `TWIN rc=0` · `MIRROR rc=0` |
| AC-ZRH-003 | PASS | `go test -run TestRegistrySyncMirrorsIdentical -v ./internal/constitution/` | `--- PASS: TestRegistrySyncMirrorsIdentical (0.00s)` + `mirrors byte-identical: 34970 bytes` |
| AC-ZRH-006(부분) | PASS | `make build && ./bin/moai constitution validate` | `make rc=0` · `exit=0`(retired 4건 skip 안내만, 드리프트 0) |

**M1 가드 통과 원문** (`go test -run TestRegistrySyncGuard -v ./internal/constitution/`, M1 완료 트리):

```
registry_sync_test.go:207: [local mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
registry_sync_test.go:217: [local mirror] clause literal buckets: once=97 zero=0 multi=0 retired_exempt=4 self_reference=0
registry_sync_test.go:207: [template mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
registry_sync_test.go:217: [template mirror] clause literal buckets: once=97 zero=0 multi=0 retired_exempt=4 self_reference=0
--- PASS: TestRegistrySyncGuard (0.12s)
```

**M1 RED 사전 관측** (rewrap 전 baseline, 트리 `a739d04b4`): `grep -c -F 'The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.' .claude/rules/moai/workflow/ci-autofix-protocol.md` → `0` (rc=1) — 신규 clause 문장이 편집 전 원본에 단일 행으로 존재하지 않았음(rewrap 필요 조건).

**Gaps**: AC-ZRH-007(M3 소관) 미측정. **Residual-risk**: 없음(측정 4축 전부 이번 실행 관측).

### M2 — F2 튜플 digest pinning (TDD)

**RED 원문** (구현 전, 스크래치 테스트 `TestRedEvidenceCountPreservingIDSubstitutionPassesCurrentGuard` — 개수 보존 ID 치환 `CONST-V3R2-004 → CONST-V3R2-0040` 을 repo 루트 스크래치 사본에 적용, 현 가드 전체 관측; 트리 `918840f61`, 캡처 후 스크래치 파일 삭제):

```
=== RUN   TestRedEvidenceCountPreservingIDSubstitutionPassesCurrentGuard
    red_evidence_scratch_test.go:56: RED: Validate(mutant) Skipped=false DriftCount=0 Status=ok
    red_evidence_scratch_test.go:66: RED: LoadRegistry(mutant) entries=101 (wantRegistryEntries=101)
    red_evidence_scratch_test.go:71: RED OBSERVED: count-preserving ID substitution (CONST-V3R2-004 -> CONST-V3R2-0040) passes the CURRENT guard: validator DriftCount=0, entry count=101 — no existing check inspects the (id,zone,zone_class,canary_gate) tuple
--- PASS: TestRedEvidenceCountPreservingIDSubstitutionPassesCurrentGuard (0.07s)
```

→ 관측된 탈출 구멍: 프로덕션 validator도 개수 pin도 이 변이를 잡지 못한다(REQ-ZRH-005 RED 근거).

**구현**: `registryTupleDigest(entries []constitution.Rule)` — 엔트리마다 `fmt.Sprintf("%s|%s|%s|%t", id, zone, zone_class, canary_gate)` 행 → `sort.Strings` → `"\n"` join → `sha256` hex. 상수 `wantTupleDigest = 2edb5384085bccbea2f9fd85e535ac4bc3ec63f7cb0918c075340882b16f51c5`(현 101엔트리 실측). guard의 개수 단정 직후 mirror별 digest 단정 추가. 변이 테이블 `TestRegistryTupleDigestRejectsSubstitution` 4종(id/zone/zone_class/canary_gate — 전부 파서가 살아있는 값으로 치환해 digest 비교 자체가 발화하도록 설계).

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-ZRH-004 | PASS | `go test ./internal/constitution/` | `ok 0.523s` |
| AC-ZRH-004 | PASS | `go test -run TestRegistrySyncGuard -v ./internal/constitution/`(grep) | mirror별 `clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries` + `once=97 zero=0 multi=0 retired_exempt=4 self_reference=0` + `--- PASS: TestRegistrySyncGuard/local` · `/template` — digest 단정 통과(실패 라인 없음) |
| AC-ZRH-005 | PASS | `go test -run TestRegistryTupleDigestRejectsSubstitution -v ./internal/constitution/` | `--- PASS: TestRegistryTupleDigestRejectsSubstitution (0.01s)` — 4 서브케이스 모두 `count preserved (101 entries), digest <서로 다른 hex> != pinned 2edb53…f51c5 — rejected` |
| AC-ZRH-006 | PASS | `make build && ./bin/moai constitution validate` | `rc=0` · `exit=0` · 추적파일 변경은 본 SPEC 파일뿐(임베드 no-op) |
| AC-ZRH-008 | PASS-WITH-DEBT | `golangci-lint run --timeout=2m ./internal/constitution/...` | `0 issues.` |
| AC-ZRH-008 | PASS | `go test -cover ./internal/constitution/` | `ok … coverage: 85.8% of statements`(≥85) |
| AC-ZRH-008 | PASS | `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` | `rc=0` |
| AC-ZRH-008 | PASS-WITH-DEBT | `gofmt -l internal/constitution/` | `canary_test.go`·`validator.go` 2건 출력 — **둘 다 본 SPEC 무편집 baseline 파일**(map 리터럴 정렬차, gofmt 버전 스큐). 본 SPEC 수정 파일 `registry_sync_test.go`는 무출력(정상). `validator.go` 재포맷은 §A.4 PRESERVE 계약 위반이라 불가 — debt로 기록 |
| AC-ZRH-009 | PASS | `grep -c -F 'update wantRegistryEntries and wantTupleDigest in the same change' internal/constitution/registry_sync_test.go` | `2`(≥1) |
| AC-ZRH-009 | PASS | 스크래치 변이(`wantTupleDigest` 말미 1바이트 `5→6`) 후 `go test -run 'TestRegistrySyncGuard/local' -v` | 아래 원문 — 계산 digest 출력 + 갱신 절차 문구 포함 확인 후 원복, `git diff --stat` 당 파일 1건(134 insertions) |

**AC-ZRH-009 실패 메시지 원문** (스크래치 변이 시점):

```
registry_sync_test.go:156: [local mirror] registry tuple digest = 2edb5384085bccbea2f9fd85e535ac4bc3ec63f7cb0918c075340882b16f51c5, want 2edb5384085bccbea2f9fd85e535ac4bc3ec63f7cb0918c075340882b16f51c6 — a pinned (id,zone,zone_class,canary_gate) tuple changed while the entry count stayed 101; if this is a deliberate registry change, update wantRegistryEntries and wantTupleDigest in the same change (REQ-ZRH-004/006)
```

**PRESERVE 검증** (M2 완료 트리): `git diff 1ae6e5c36..HEAD -- internal/constitution/validator.go` → 0행 (loader.go·rule.go 동일 무편집).

**Gaps**: AC-ZRH-007(M3·manager-spec 소관) 미측정 · CI 전 매트릭스 판독 미수행(push 전 — Route B, manager-git 소관). **Residual-risk**: gofmt 스큐 2건은 baseline 부채로 잔존(본 SPEC 스코프 밖).

### M3 — F3 문서 정렬: AC-ZRH-007 측정 (편집 manager-spec `c31c3613d` · 측정 본 에이전트 — plan §F D3 분업)

측정 트리 `c31c3613d`(M3 커밋 HEAD, 측정 시점 워킹트리 clean). Given(M3 완료 후 트리) 충족:

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-ZRH-007 | PASS | `grep -c 'strings.Count(rawFileContent, clause)' .moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md` | `0` (rc=1 — 발생 의미론 서술 제거) |
| AC-ZRH-007 | PASS | `grep -c -i '라인 수 의미론' <동일 파일>` | `2` (≥1 — 실측 라인 수 의미론 기술) |
| AC-ZRH-007 | PASS | `grep -c 'SPEC-ZONE-REGISTRY-HARDEN-001' <동일 파일>` | `1` (≥1 — 정정 출처 erratum 주석) |

Then 3축 모두 충족 → **PASS**. 오케스트레이터 독립 사전측정(strings.Count 0 · HARDEN-001 1 · 라인수의미론 2)과 값 일치. M3 편집 표면(RESYNC-001 plan.md)은 가드가 pin하지 않으므로 비회귀 확인: `go test -count=1 ./internal/constitution/` → `ok … 0.508s`(트리 `c31c3613d`).

**Gaps**: CI 전 매트릭스 판독 미수행(push 전 — Route B, manager-git 소관). **Residual-risk**: 없음(3축 전부 이번 실행·이 트리 관측).


## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-25
run_commit_sha: "c31c3613d (M3 — final milestone commit; all three milestones' edits contained in 918840f61 / 9eafd81f4 / c31c3613d; AC-ZRH-007 measurement evidence rides the follow-up progress.md commit)"
run_status: m1-m3-complete-awaiting-sync
ac_pass_count: 9  # AC-ZRH-001..009 all measured PASS
ac_fail_count: 0
ac_pending_m3: 0
ac_pass_with_debt: 1  # AC-ZRH-008 gofmt axis — pre-existing baseline skew on untouched validator.go/canary_test.go
preserve_list_post_run_count: 0  # violations: git diff 1ae6e5c36..HEAD -- validator.go = 0 lines (measured)
l44_pre_commit_fetch: "0 2 (origin/main...HEAD at 918840f61 — local ahead only, no parallel race)"
l44_post_push_fetch: "pending-push (Route B PR-mandatory: push + PR owned by manager-git)"
new_warnings_or_lints_introduced: 0  # golangci-lint ./internal/constitution/... → "0 issues."; registry_sync_test.go gofmt-clean
cross_platform_build:
  windows_vet_rc: 0  # GOOS=windows GOARCH=amd64 go vet ./internal/constitution/
total_run_phase_files: 6  # M1: 4 (ci-autofix twin pair + zone-registry mirror pair) · M2: 1 (registry_sync_test.go) · M3: 1 (RESYNC-001 plan.md, manager-spec)
m1_to_mN_commit_strategy: "plan-phase artifacts commit + per-milestone commit (M1 fix / M2 feat by manager-develop, M3 docs by manager-spec per §F D3) + measurement-record commit, card id t268 in every subject, no push (Route B)"
```

**커밋 목록(run-phase 전체)**: `0da345433`(plan-phase artifacts) → `918840f61`(M1, manager-develop) → `9eafd81f4`(M2, manager-develop) → `c31c3613d`(M3, manager-spec) → progress.md 측정 기록 커밋(본 커밋, manager-develop).

**M3 소유권(D3) 이행**: RESYNC-001 plan.md 본문 편집과 그 커밋은 manager-spec이 실행(`c31c3613d`), 본 에이전트는 AC-ZRH-007 측정만 담당(§E.2 M3 절).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-25
sync_commit_sha: "pending-backfill-SPEC-ZONE-REGISTRY-HARDEN-001"  # backfilled in the immediately following commit (self-referential-hazard workaround per spec-frontmatter-schema.md D3)
sync_status: completed
b12_self_test_a: "grep -c 'SPEC-ZONE-REGISTRY-HARDEN-001' CHANGELOG.md → 0 pre-emission (no duplicate)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 9; CHANGELOG entry states 9 AC — match"
b12_self_test_c: "all cited paths ls-verified (spec/plan/acceptance/progress.md, internal/constitution/registry_sync_test.go, both template rule paths, RESYNC-001 plan.md)"
changelog_entry_position: "CHANGELOG.md [Unreleased] / Added — top entry"
public_surface_assessment: "no public surface change — internal test-only guard (registry_sync_test.go), registry data clause value, formatting-only template rewrap, closed-SPEC plan.md erratum; CLI behavior / README / docs-site untouched (assessed & recorded at sync, this SPEC)"
template_neutrality_5item: "C1-C5 all PASS on both template diffs (no SPEC-ID/REQ-AC token/audit-citation/date-sha/memory-path in staged template changes); ci-autofix-protocol.md rewrap + zone-registry.md clause swap are content-neutral"
template_catalog_parity: "n/a — catalog.yaml catalogues skills/agents only; the two changed template rule paths are not catalogued (grep 'zone-registry|ci-autofix' catalog.yaml → 0)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (this sync commit; plan.md/acceptance.md carry no frontmatter block — spec.md is the only frontmatter-bearing artifact)"
  updated_refreshed: "2026-08-25 (already current)"
canary_compliance_check: {}  # no forward-looking policy self-tested by this SPEC
```

**Sync-phase surface (this commit, markdown-only — zero code changes)**: CHANGELOG `[Unreleased]` entry · `spec.md` frontmatter close (`status: in-progress → completed`) · this §E.4 signal. Template files and the 6 run-phase implementation files are untouched by sync.
