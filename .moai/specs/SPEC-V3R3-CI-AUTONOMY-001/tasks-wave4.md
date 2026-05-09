---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 4
version: 0.1.0
status: draft
created_at: 2026-05-08
updated_at: 2026-05-08
author: manager-strategy
---

# Wave 4 Atomic Tasks (Phase 1 Output)

> Companion to strategy-wave4.md. SPEC-V3R3-CI-AUTONOMY-001 Wave 4 — T4 Auxiliary Workflow Cleanup.
> Generated: 2026-05-08. Methodology: TDD-adapted for CI/CD config (verify-via-replay; sh -n / yamllint / gh workflow list / manual replay). Wave Base: `origin/main d7bd9c453`.

---

## Atomic Task Table

| Task ID | Description | Files (provisional) | REQ | AC | Status |
|---------|-------------|---------------------|-----|----|--------|
| W4-T01 | Rename `claude-code-review.yml` → `claude-code-review.optional.yml` (git mv 로 history 보존). `llm-panel.yml` 은 origin/main 에 부재 — `.github/required-checks.yml` `auxiliary:` 리스트에서 `llm-panel` 항목 제거 (deprecated placeholder). 트리거 그대로 유지 (workflow `name:` field + `on:` 블록 변경 안 함). PR check 표시명 보존을 `gh workflow list --json name,path` 로 사전 검증. | `.github/workflows/claude-code-review.yml` (git mv → `.optional.yml`), `.github/required-checks.yml` (`auxiliary:` 리스트 정리) | REQ-CIAUT-020 | AC-CIAUT-009 | pending |
| W4-T02 | `docs-i18n-check.yml` 에 job-level `continue-on-error: true` 추가 + advisory PR comment header 갱신 ("ADVISORY ONLY — NOT BLOCKING" 명시). 기존 `DOCS_I18N_STRICT` env var 처리 로직 보존 (additive 변경; phased rollout Phase 1 warn-only 모드 유지). | `.github/workflows/docs-i18n-check.yml` | REQ-CIAUT-022, REQ-CIAUT-023 | AC-CIAUT-009 | pending |
| W4-T03 | `release-drafter-cleanup.yml` 신규 작성 — `schedule: cron: '0 9 * * MON'` (주 1회 월요일 09:00 UTC) + `workflow_dispatch`. `gh release list --json tagName,isDraft,createdAt -L 100` → jq filter 30일+ stale + targetCommitish 가 release/* 미매칭 → `gh release delete <tag> --yes`. Log append to `.github/release-drafter-cleanup-log.md`. `MOAI_CLEANUP_DRY_RUN=1` env var 지원. permissions: `contents: write`. | `.github/workflows/release-drafter-cleanup.yml` (new), `.github/release-drafter-cleanup-log.md` (placeholder) | REQ-CIAUT-021 | AC-CIAUT-010 | pending |
| W4-T04 | Makefile `ci-disable WORKFLOW=<name>` target 추가. `yq -i '.on = {"workflow_dispatch": null}' .github/workflows/$(WORKFLOW).yml` 로 트리거 neutralize → git add + commit (`chore(ci): disable <name> (workflow_dispatch only)`). yq 미설치 fallback 메시지 (`command -v yq`). `--no-commit` flag 검토 (Phase 2 자유도). target description (`## Disable a workflow ...`) 포함. | `Makefile` (extend) | REQ-CIAUT-024 | AC-CIAUT-009 (transitive — 미사용 workflow 정리 도구) | pending |
| W4-T05 | `.github/required-checks.yml` SSoT 무결성 검증 스크립트 작성. (a) `auxiliary:` 항목과 실제 workflow `name:` 매칭 0 mismatch, (b) `branches.main.contexts` 와 `branches.release/*.contexts` 에 auxiliary 항목 미포함 (`comm -12` set intersection), (c) Wave 4 rename 후 SSoT 정합성 회복 검증. `scripts/ci-mirror/validate-required-checks.sh` (POSIX sh) 또는 Makefile `verify-required-checks` target 으로 자동화. `make ci-local` lint 단계에 통합 (Wave 1 framework 활용, additive). | `scripts/ci-mirror/validate-required-checks.sh` (new), `Makefile` (extend with `verify-required-checks`) | REQ-CIAUT-027 (transitive Wave 1 SSoT) | AC-CIAUT-009 + AC-CIAUT-021 transitive | pending |

---

## File Ownership Assignment (Solo 모드 — sub-agent + --branch 패턴, lessons #13)

### Implementer Scope (write access)

```
.github/workflows/claude-code-review.yml                   # git mv to .optional.yml (W4-T01)
.github/workflows/claude-code-review.optional.yml          # rename target (W4-T01)
.github/workflows/docs-i18n-check.yml                      # extend with continue-on-error (W4-T02)
.github/workflows/release-drafter-cleanup.yml              # new (W4-T03)
.github/release-drafter-cleanup-log.md                     # placeholder for cleanup audit log (W4-T03)
.github/required-checks.yml                                # auxiliary: 리스트 정리 (W4-T01)
Makefile                                                   # ci-disable + verify-required-checks targets (W4-T04, W4-T05)
scripts/ci-mirror/validate-required-checks.sh              # new validation script (W4-T05)
```

### Tester Scope (test 파일만 write, production 코드 read-only)

```
scripts/ci-mirror/test/validate-required-checks_test.sh    # new test harness (W4-T05 검증)
```

### Read-Only Scope (cross-Wave consumer / SSoT source)

```
.github/workflows/ci.yml                                   # canonical job names (Lint, Test (...), Build (...), CodeQL) — read-only reference
.github/workflows/auto-merge.yml                           # auto-merge 패턴 (Wave 5/6 참고) — read-only
.github/workflows/release-drafter.yml                      # 기존 Release Drafter (cleanup workflow 의 pair) — read-only
internal/template/templates/.github/required-checks.yml    # user-project SSoT — Wave 4 변경은 dev-only이므로 미러 안 함; 차이 검증을 위한 read-only
internal/template/templates/.github/workflows/             # user-project workflow templates (.tmpl) — Wave 4 와 무관, read-only
scripts/ci-mirror/run.sh                                   # Wave 1 SSoT framework — W4-T05 가 lint 단계에 통합 시 read 만
scripts/ci-mirror/lib/*.sh                                 # Wave 1 language modules — read-only
```

### Implicit Read Access (모든 task)

- `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/` (spec.md, plan.md, acceptance.md, strategy-wave4.md, 본 파일)
- `.claude/rules/moai/**` (auto-loaded rules — agent-common-protocol.md, askuser-protocol.md)
- Wave 1/2/3 산출물 (read-only consumer)
- `CLAUDE.md`, `CLAUDE.local.md` (project rules)

---

## AC Mapping

| Wave 4 Task | Drives AC | Validation |
|-------------|-----------|------------|
| W4-T01 + W4-T05 | AC-CIAUT-009 (auxiliary workflow non-blocking — claude-code-review rename + SSoT 정합성) | rename 후 `gh workflow list --json name,path` 결과 매칭; `.github/required-checks.yml` `auxiliary:` 리스트가 실제 workflow `name:` field 와 1:1 매칭; PR 의 claude-code-review fail 시 ready-to-merge 유지 manual replay |
| W4-T02 | AC-CIAUT-009 (docs-i18n-check non-blocking) | `continue-on-error: true` 추가 후 의도적 fail PR 이 merge 차단 안 함; advisory PR comment header 가 "ADVISORY ONLY" 표기 포함 |
| W4-T03 | AC-CIAUT-010 (Release Drafter stale cleanup) | `workflow_dispatch` 수동 호출 dry-run → log entry 작성; cron 첫 실행 (1주 후) 모니터링; 30일+ stale draft 삭제; targetCommitish=release/* 매칭 시 보호 검증 |
| W4-T04 | AC-CIAUT-009 (transitive — workflow disable 도구 제공) | test fixture workflow 생성 → `make ci-disable WORKFLOW=test-fixture` → `yq` 결과가 `on: workflow_dispatch: null` 로 변경 + commit 메시지 형식 (`chore(ci): disable test-fixture`); yq 미설치 환경 fallback 메시지 검증 |
| W4-T05 | AC-CIAUT-009 + AC-CIAUT-021 transitive (SSoT 무결성) | validation script 가 (a)(b)(c) 3 dimension 모두 검증; `make ci-local` 가 회귀 시 즉시 fail; `make verify-required-checks` 단독 호출 가능 |

---

## TRUST 5 Targets (Wave 4 SPEC-Level DoD)

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | "verify-via-replay" — Go unit test 부적절. (a) yamllint structural test, (b) sh -n syntax check (Makefile body + scripts/ci-mirror/validate-required-checks.sh), (c) AC-CIAUT-009 manual replay (docs-i18n-check fail PR ready-to-merge), (d) AC-CIAUT-010 manual replay (release-drafter-cleanup workflow_dispatch dry-run) | `yamllint .github/workflows/*.yml`; `sh -n scripts/ci-mirror/validate-required-checks.sh`; manual replay 결과를 `.moai/reports/wave4-replay/` 에 기록 |
| **Readable** | yaml indent 2-space + workflow `name:` field 명시 + Makefile target description (`## ...`) 포함; bash script 는 `set -euo pipefail` (POSIX sh 가능 시 `set -eu`) + 함수 단위 분리 + 주석 (한국어 OK per CLAUDE.local.md §3 code_comments 설정 — 단 현재 quality.yaml 기준 영문 코드 주석) | `make help` 출력에 `ci-disable` + `verify-required-checks` 표시; shellcheck (W4-T05 script 대상) clean |
| **Unified** | yaml 스타일이 기존 `.github/workflows/*.yml` 컨벤션 따름 (boolean unquoted, sequence 형식, single-quote for strings with special chars); Makefile target naming `<verb>-<noun>` (ci-local 패턴 계승); validation script 파일 이름 `validate-<noun>.sh` 패턴 | 코드리뷰 + diff 검토; `git diff` 에 indent/quote 일관성 |
| **Secured** | release-drafter-cleanup workflow 의 `permissions: contents: write` 명시 (least privilege); `gh release delete` 가 stale 한정 + targetCommitish 검증 (active release 보호); `MOAI_CLEANUP_DRY_RUN=1` 으로 첫 1주 검증; Makefile `ci-disable` 가 secrets 파일 변경 안 함 (`.github/workflows/*.yml` 만 대상); validation script 가 SSoT 외 파일 modify 안 함 (read-only) | workflow yaml `permissions:` 블록 검증; dry-run log 첫 1주 모니터링; `git diff` 에 secrets 변경 0 |
| **Trackable** | 모든 commit 이 SPEC-V3R3-CI-AUTONOMY-001 W4 reference 포함; Conventional Commits + 🗿 MoAI co-author trailer; `.github/release-drafter-cleanup-log.md` 가 audit trail 보존; W4-T05 validation 결과를 `.moai/reports/wave4-validation/` 에 기록 | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W4'`; cleanup log 파일 존재 + entry append 확인 |

---

## Per-Wave DoD Checklist

- [ ] 모든 5개 W4 task 완료 (위 표)
- [ ] **Template-First mirror 검증 — Wave 4 는 미러 불필요** (strategy-wave4.md §15 참조; `.github/workflows/*` + `Makefile` 은 dev-only)
- [ ] `git mv` 로 `claude-code-review.yml` → `claude-code-review.optional.yml` rename (history 보존 + trigger 영향 없음 검증)
- [ ] `.github/required-checks.yml` `auxiliary:` 리스트에서 `llm-panel` 제거 + 주석 `# llm-panel.yml: removed YYYY-MM-DD, placeholder only`
- [ ] `docs-i18n-check.yml` 에 job-level `continue-on-error: true` 추가 + advisory comment header 갱신
- [ ] `release-drafter-cleanup.yml` 신규 작성 (cron + workflow_dispatch + gh release CLI + targetCommitish 보호 + dry-run env)
- [ ] `Makefile` 에 `ci-disable WORKFLOW=<name>` + `verify-required-checks` target 추가
- [ ] `scripts/ci-mirror/validate-required-checks.sh` 신규 작성 (POSIX sh, 3-dimension 검증)
- [ ] `make ci-local` 통과 (Wave 1 framework 회귀 없음 + W4-T05 validation 통합 후 정상 동작)
- [ ] `yamllint .github/workflows/*.yml` 통과 (모든 yml 파일)
- [ ] `sh -n scripts/ci-mirror/validate-required-checks.sh` 통과 + shellcheck clean
- [ ] AC-CIAUT-009 manual replay 통과 (docs-i18n-check fail PR ready-to-merge)
- [ ] AC-CIAUT-010 manual replay 통과 (release-drafter-cleanup workflow_dispatch dry-run + log entry)
- [ ] `gh workflow list` 출력에 rename 후 workflow 정상 등재 (`Claude Code Review` name 보존; path 만 변경)
- [ ] No release/tag automation 도입 (`gh release create` / `git tag` / `goreleaser` 호출 0 — Wave 4 의 cleanup workflow 는 `gh release delete` 만)
- [ ] No hardcoded URLs/models/paths (validation script + Makefile target 은 변수 + SSoT 참조 only)
- [ ] PR labeled with `type:chore`, `priority:P1`, `area:ci`
- [ ] Conventional Commits + 🗿 MoAI co-author trailer 모든 commit
- [ ] CHANGELOG.md 에 Wave 4 머지 entry

---

## Out-of-Scope (Wave 4 Exclusions)

- Wave 5 (T6 worktree state guard) — 독립적, 별도 Wave
- Wave 6 (T7 i18n validator) — 독립적, 별도 Wave
- Wave 7 (T8 BODP) — 독립적, 별도 Wave
- claude-code-review 의 fix 시도 (org quota 증설, 다른 LLM 으로 대체) — 사용자 별도 SPEC 필요 (Wave 4 는 disable/rename 만)
- llm-panel.yml 재도입 — 별도 SPEC + 사용자 의사결정 필요 (Wave 4 는 SSoT 정리만)
- workflow `name:` field 에 "(advisory)" suffix 추가 — Phase 2 follow-up commit 또는 별도 cleanup wave
- release-drafter-cleanup 의 backfill (이미 80+ stale draft 일괄 삭제) — dry-run 1주 후 사용자 명시 승인 필요
- `internal/template/templates/.github/required-checks.yml` 미러 갱신 — Wave 4 변경은 dev-project-specific (auxiliary 리스트는 user 환경마다 다름)
- 16-language neutrality 검증 — Wave 4 산출물은 GitHub Actions YAML + Makefile 로 언어 무관
- Cross-platform Windows variant of Makefile target — Wave 1/2/3 패턴 계승 (Windows 사용자는 git-bash + GNU make 가정)
- release-drafter-cleanup 의 PR comment notification (close 시 PR 작성자 알림) — over-engineering
- T4 deliverables 의 spec.md REQ wording 정정 — 본 wave scope 외 documentation hygiene (Wave 3 D1 처리 패턴 계승)
- spec.md REQ-CIAUT-020 본문이 "moved to .github/workflows/optional/" 명시 — 본 strategy 의 `.optional.yml` rename 결정과 wording 차이; follow-up commit 으로 spec.md 정정 (Wave 4 implementation 산출물 자체와는 무관)

---

## Honest Scope Concerns

1. **GitHub Actions cached workflow registry stale**: `gh workflow list` 가 origin/main 에 부재한 `llm-panel.yml` 을 active 로 표시하는 현상 — Wave 4 머지 후에도 GitHub UI 가 즉시 갱신되지 않을 수 있음. Mitigation: PR description 에 "expect GitHub UI caching may show llm-panel as removed within 24h" 명시; 사용자 confused 시 manual workflow disable 안내 (`gh workflow disable llm-panel.yml` — 단 파일 부재 시 동작 모호).

2. **`.optional.yml` suffix 가 GitHub UI 에서 시각적으로 구분 안 됨**: PR check tab 은 workflow 의 `name:` field 만 표시 (filename 노출 안 함). reviewer 는 `Claude Code Review` 결과를 advisory 로 인식 못 할 수 있음. Mitigation: workflow `name:` field 자체에 "(advisory)" suffix 추가 — 단, 이는 Wave 4 scope 외 follow-up commit (Out-of-Scope 항목 참조).

3. **release-drafter-cleanup 의 30일 임계값이 long-running release branch 에 false positive 가능**: 6주 이상 진행 중인 release/v3.0.0 의 draft 가 30일 임계 도달 시 false positive 발생 가능. Mitigation: targetCommitish 검증 (release/* 브랜치 매칭 시 예외) + dry-run 첫 1주 모니터링 + 30일 임계값을 `MOAI_CLEANUP_THRESHOLD_DAYS` env var 로 외부화.

4. **Makefile `ci-disable` 의 자동 commit 이 사용자 직접 편집 의도와 충돌 가능**: 사용자가 다른 변경사항과 함께 disable 하려는 경우 자동 commit 이 분리된 commit 만들어 stash/rebase 부담. Mitigation: `--no-commit` flag 검토 (Phase 2 자유도) — 또는 디폴트를 staged-only 로 변경 (`git add` 만 수행).

5. **W4-T05 validation script 의 위치 (`scripts/ci-mirror/`)가 Wave 1 framework 와 namespace 모호**: Wave 1 의 `lib/<lang>.sh` 패턴 (CI mirror execution) 과 Wave 4 의 validation 스크립트는 다른 목적. Mitigation: Phase 2 가 디렉터리 구조 명확화 — 예: `scripts/ci-mirror/lib/` (lang modules), `scripts/ci-mirror/validate/` (validation), `scripts/ci-mirror/run.sh` (entry point).

No hard blockers identified. Wave 4 ready for Phase 2 (manager-tdd) delegation upon strategy + tasks approval.

---

Version: 0.1.0
Status: pending Phase 2 (manager-tdd 위임 대기)
Last Updated: 2026-05-08
