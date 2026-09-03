# SPEC-PLAN-AUDITOR-RESIDUE-001 — 진행 기록

카드: t450 · 브랜치: WT-plan-auditor-residue · Tier: M

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-PLAN-AUDITOR-RESIDUE-001
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
author: manager-spec
tree_sha_at_authoring: "7835148d3"
precondition_verified:
  t367_ancestor_of_develop: true   # merge-base 측정 — 18fc2c9ef ∈ origin/develop 7835148d3
needs_clarification_markers: 0      # t443 우회는 t367 선례(2549f775f)로 확정 — 질의 불필요
hazard_recorded:
  t443_agentemit_drift: sync-auditor.toml sha256 mismatch — go test ./internal/template/agentemit/... FAIL 관측
                                    # t443 소관 — 본 SPEC은 record-and-not-repair (REQ-008)
evidence_path_planned: .moai/reports/t450/verdict.md   # lane이 작성 — plan-phase에서 작성 안 함
```

plan-phase 측정 근거 (모두 이번 실행, 트리 7835148d3):
- 지연 조항 출처: f47d7f5a9 본문 + `.moai/reports/t387/verdict.md` Gaps / 4244c4a06 본문
- 금지 경로 명령: plan-auditor.md:395 (gitignored — `.gitignore:230`, check-ignore exit=0)
- 쌍둥이 기존 드리프트 2 hunk: D7-1 예시 식별자 + Tier-resolved ceiling 문단 (수리 안 함, 기록만)
- agentemit 골든 FAIL: sync-auditor.toml sha256 mismatch (t443 소관)

## §E.2 Run-phase Evidence

실행 트리: 워크트리 t450 / 브랜치 WT-plan-auditor-residue / 시작 HEAD `194fc9439`(develop `6765a75c0` 흡수 머지). 전제 재검증: `git merge-base --is-ancestor 18fc2c9ef develop` exit 0. 원본 명령 출력 전문: `.moai/reports/t450/run-evidence-raw-outputs.md`(브랜치 커밋).

AC 매트릭스(모두 이번 실행, 커밋 직전 작업 트리에서 재측정 — 판정 명령과 출력은 원본 파일 참조):

| AC | 판정 | 근거(요지) |
|----|------|-----------|
| AC-001 반출 조항 존재 | PASS | `grep -c "Export mandate"` 양쪽 사본 각 1, exit 0. 반출 위치 `.moai/reports/<card-id>/plan-audit.md`(또는 `plan-audit-iter<N>.md`, `<SPEC-ID>/`) 명시 |
| AC-002 금지 경로 제거 | PASS | `grep -n "reports/plan-audit/"` 양쪽 사본 매치 0, exit 1 (RED-now :395 매치에서 뒤집힘). 금지 언급은 규약 § Where 참조로 경로 리터럴 없이 대체 |
| AC-003 곁말 규약 반영 | PASS | `grep -c "Side-talk"` 양쪽 각 1 + `measured`/`inferred`/`assumption` 3라벨 모두 존재(양쪽 동일) |
| AC-004 조항 쌍둥이 일치 | PASS | 조항 범위 diff 0 (clause-scoped 추출 비교, AC-004-CLAUSE-SCOPED-DIFF-0). 전체 diff는 기존 2 hunk만(판정 제외 — 적법) |
| AC-005 교차참조 일치 | PASS | (a) 사본 `reports/plan-audit/` 매치 0. (b) spec-workflow § Report Persistence가 review stream을 반출 패밀리로 재서술(양쪽 미러 바이트 동일), run-gate stream만 런타임 기록 디렉터리로. 규약 § stick 감사자 측 서술이 참이 됨. § Cross-references에 plan-auditor 반출 패밀리 명시 |
| AC-006 방출물+카탈로그 | PASS(한정) | `AGENTEMIT_UPDATE=1` 전체 재생성 → `ok` 관측 → sync-auditor.toml만 develop 값 복원(sha `5306b92e…` byte-identical, `git diff develop` 빈 출력) → 골든 적색이 sync-auditor.toml 1건으로 한정(REQ-008). 카탈로그 plan-auditor 해시 갱신 `2403bfb3…` (`gen-catalog-hashes --entry plan-auditor`). TestManifestHashFormat의 sync-auditor CATALOG_HASH_UNSTABLE 1건은 선존재 상속 적색(양변 HEAD 바이트 — 본 실행 미터치), t443/t444 소관 |
| AC-007 t367 :72 보존 | PASS | `the fifth GEARS pattern` / `NOT a GEARS pattern` 양쪽 각 1 — 편집 전후 동일 |
| AC-008 중립성 0매치 | PASS | 신규 템플릿 문면+편집 문서 2종에서 SPEC-ID 일반형·카드 id·40hex SHA·날짜 매치 0. TestTemplateNeutralityAudit / TestTemplateNoInternalContentLeak / TestRuleTemplateMirrorDrift 전부 `ok`. `go build ./...` BUILD-OK |

적용 내용 요약: (1) plan-auditor.md 쌍둥이 § Output Format — sync-auditor 착지 문체(4244c4a06)를 따른 [HARD] Export mandate 단락 + Side-talk discipline 단락 + 두 스트림 재서술로 교체. (2) spec-workflow.md § Report Persistence 양쪽 미러 — "Two report streams coexist deliberately in `.moai/reports/plan-audit/`" 문장을 "두 스트림이 디렉터리를 공유하지 않는다"로 정정. (3) audit-artifact-convention.md 양쪽 미러 § Cross-references — plan-auditor 반출 패밀리를 § Where로 조인. 규약 문서 자체에는 지연(deferral) 문구가 없었음(f47d7f5a9 본문과 t387 verdict Gaps에만 존재) — § stick 감사자 측 문장은 에이전트 정의 조항 착지로 참이 됐고 문서 무변경으로 충족. 곁말 반영 시 금지 경로 리터럴을 문장에 넣지 않아 AC-002/AC-005(a)의 grep 0매치를 유지했다.

작업 중 자체 수정 1건: 템플릿 사본 편집 시 `.mooi/` 오타가 들어갔으나 커밋 전 자체 발견·수정 — AC-004 clause diff 0으로 최종 확인.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-03"
run_commit_sha: "c27cc4d1c"   # run-phase 최종 커밋 — D3 패턴 백필 (선례: SPEC-CODEX-SKILL-PATH-001 등 완료 SPEC 전반, 최종 커밋 기록 관례)
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
ac_conditional_notes: "AC-006은 REQ-008 한정 판정(적색 sync-auditor.toml 1건 기록-미수리)으로 PASS"
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run   # 로컬 병합 창 경유 — 레인은 push 안 함(리드 일괄)
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_all: "BUILD-OK (docs-only 레인 — 로컬 전체 스위트 금지 준수, 전 판정은 origin/develop CI)"
total_run_phase_files: 10
m1_to_mN_commit_strategy: "3커밋 — 조항+교차참조(쌍둥이 동일 커밋) / 방출물+카탈로그 / 증거+진행기록"
not_repaired_inherited_reds:
  - "agentemit golden: sync-auditor.toml sha256 mismatch — t443 소관"
  - "TestManifestHashFormat CATALOG_HASH_UNSTABLE 2건 — 모두 선존재 상속 적색 (본 브랜치 f5b9bb654 직접 측정): (1) sync-auditor catalog hash stale(4244c4a06 이래 — t443/t444 소관) (2) moai whole-tree entry(.claude/skills/moai/) hash stale — 흡수 머지 a261192b7로 develop(lane-5 regen)에서 유입. 상속 검증: 본 실행 저작 커밋(194fc9439..c27cc4d1c)의 catalog.yaml 델타는 plan-auditor 1줄뿐이고 .claude/skills/ 트리 델타는 0"
  - "쌍둥이 기존 2 hunk(D7-1 식별자, Tier ceiling 문단) — SPEC Out of Scope"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec_id: SPEC-PLAN-AUDITOR-RESIDUE-001
sync_complete_at: "2026-09-03"
sync_commit_sha: "4fb90a1cd"   # sync 커밋 실측 SHA — D3 2-commit pattern 백필
sync_status: complete
changelog_entry_position: emitted
b12_self_test_a: "grep -c SPEC-PLAN-AUDITOR-RESIDUE-001 CHANGELOG.md → 0 (exit 1) — 중복 없음 확인 후 1건 추가"
b12_self_test_b: "AC 식별자 8건 (AC-001..AC-008, acceptance.md §D 매트릭스 sort -u 실측) = CHANGELOG 서술 카운트 8/8 일치"
b12_self_test_c: "CHANGELOG 인용 경로 전부 실존 — .claude/agents/moai/plan-auditor.md, internal/template/templates/.claude/agents/moai/plan-auditor.md, .moai/docs/audit-artifact-convention.md ls 확인"
frontmatter_status_transitions:
  - "in-progress → implemented → completed (단일 sync 커밋, spec.md status: completed, updated: 2026-09-03)"
canary_compliance_check:
  spec_body_untouched: true        # spec.md/plan.md/acceptance.md 본문 무변경 — frontmatter status/updated만
  evidence_path_exported: ".moai/reports/t450/ (plan-audit.md, run-evidence-raw-outputs.md — 브랜치 커밋済)"
  docs_site_scheduled: false       # 본 SPEC 계획에 docs-site 작업 없음 — 사용자 대면 제품 동작 변경 없음, 부재는 유효 결과
  mx_tag_validation: sync sub-step — 문서/템플릿 편집 대상 @MX 어노테이션 요건 해당 없음 (마크다운 전용 변경)
```

CHANGELOG 판정: **발행(entry)** — plan-auditor 에이전트 정의는 템플릿 배포물이라 사용자 대면 표면이고, 판정문 반출 계약 변경(Export mandate + Side-talk 조항)은 배포 사용자의 plan-audit 워크플로에 직접 영향을 준다. maintainer-facing 규약 문서 변경은 그에 수반하는 보조 변경이다. `### Changed` 절에 1건 추가.
