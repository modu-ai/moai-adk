---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 4
version: 0.1.0
status: draft
created_at: 2026-05-08
updated_at: 2026-05-08
author: manager-strategy
---

# Wave 4 Execution Strategy (Phase 1 Output)

> Audit trail. manager-strategy output for Wave 4 of SPEC-V3R3-CI-AUTONOMY-001 — T4 Auxiliary Workflow Cleanup.
> Generated: 2026-05-08. Methodology: TDD-adapted for CI/CD config (verify-via-replay; sh -n / yamllint / gh workflow list / manual non-block replay).
> Base: `origin/main d7bd9c453` (Wave 3 PR #790 still OPEN — Wave 4는 plan.md §6에 따라 독립적이며 Wave 1 SSoT만 transitive 의존).

---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-08 | manager-strategy | Initial draft. Resolved OQ1-OQ6 inline (§"Decisions"). Reframed plan.md `optional/` directory pattern to `.optional.yml` rename pattern (OQ2) per GitHub Actions auto-discovery limitation. AC mapping: AC-CIAUT-009 + AC-CIAUT-010. |

---

## 1. Wave 4 Goal

Wave 4는 5-PR sweep (2026-05-05) 사례에서 노출된 R5 (auxiliary workflow noise가 CI signal을 침몰시킴) 를 종결한다. claude-code-review (org quota 부족으로 영구 실패), `llm-panel` (실제 파일 없음, required-checks.yml에 보호 항목으로만 등재), Release Drafter (80+ stale draft 누적), docs-i18n-check (warn-only 의도이나 일부 환경에서 PR red X로 노이즈 누적) 4종 auxiliary workflow를 정리하여 PR signal을 "필수 check만 빨간/초록"으로 회복한다. 본 Wave는 Wave 1의 `.github/required-checks.yml` SSoT와 transitive하게만 의존하고, Wave 2/3 (CI watch + auto-fix) 와 무관하게 독립 적용 가능하다 (plan.md §6 Dependencies). AC-CIAUT-009 (auxiliary non-blocking) 와 AC-CIAUT-010 (Release Drafter stale cleanup) 가 본 Wave 외에서는 충족 불가능하다.

---

## 2. Files Touched

| # | Path | Role | New / Extend / Move |
|---|------|------|---------------------|
| 1 | `.github/workflows/claude-code-review.optional.yml` | renamed from `.github/workflows/claude-code-review.yml` (W4-T01) | Rename (git mv) |
| 2 | `.github/workflows/llm-panel.optional.yml` | placeholder removed from `auxiliary:` (W4-T01 — file does not exist on origin/main) | Spec/SSoT-only update |
| 3 | `.github/workflows/docs-i18n-check.yml` | extend with `continue-on-error: true` job-level guard + advisory comment header (W4-T02) | Extend |
| 4 | `.github/workflows/release-drafter-cleanup.yml` | scheduled cron workflow (gh release CLI) (W4-T03) | New |
| 5 | `Makefile` | add `ci-disable WORKFLOW=<name>` target (W4-T04) | Extend |
| 6 | `.github/required-checks.yml` | reconcile `auxiliary:` list with actual filenames after rename (W4-T05) | Extend (verification-driven) |

**No template mirror required**: All 6 files live under `.github/` or `Makefile` at repo root and are NOT mirrored to `internal/template/templates/`. Verified by inspection of `internal/template/templates/.github/workflows/` (only contains `*.tmpl` for user-project Claude review workflows, not moai-adk-go's own CI). See §15 "Template-First Note" below.

---

## 3. Approach §1 — `*.optional.yml` Rename Pattern (resolves OQ2)

`.github/workflows/optional/` 서브디렉터리 이동 (plan.md §6 W4-T01 원안) 은 GitHub Actions 가 자동 디스커버리하지 않는다 — official docs (`docs.github.com/actions/using-workflows/about-workflows`) 에 따르면 GitHub Actions 는 `.github/workflows/` 디렉터리 직속 `.yml`/`.yaml` 파일만 등록한다. 서브디렉터리 이동 시 워크플로우는 영구적으로 비활성화되어 트리거 자체가 발생하지 않는다.

**채택 방식**: filename suffix 변경 (`<name>.yml` → `<name>.optional.yml`).

이유:
1. 파일은 `.github/workflows/` 직속에 유지되어 GitHub auto-discovery 가능 (트리거 보존)
2. `.optional.yml` suffix 가 사람과 도구 모두에게 "advisory only" 시각 신호 제공
3. Required-checks SSoT (`.github/required-checks.yml`) 의 `auxiliary:` 리스트와 명시적으로 매칭 가능
4. Git history 보존 (`git mv` 한 번)
5. 추가적으로 job-level `if: github.event_name != 'pull_request' || vars.RUN_OPTIONAL == 'true'` 같은 conditional gate 도 결합 가능 (Phase 2 자유도)

**대안 (rejected)**:
- `.github/workflows/optional/` 서브디렉터리 이동 → GitHub Actions가 인식 안 함 (트리거 사라짐)
- `if: false` job-level guard → workflow 가 등록되긴 하나 모든 job 이 skipped로 표시되어 PR check 화면 노이즈 유지
- `workflow_dispatch` 만 트리거로 남김 → 자동 실행 사용 사례 (예: claude-review의 자동 코드 리뷰 코멘트) 손실

**llm-panel 처리**: 본 wave 는 `llm-panel.yml` 파일이 origin/main 에 존재하지 않음을 확인 (`git ls-tree origin/main .github/workflows/`, 2026-05-08). `.github/required-checks.yml` `auxiliary:` 리스트의 `llm-panel` 항목은 deprecated placeholder 로 잔존. W4-T05 에서 SSoT 정합성 정리 (OQ1 §"Decisions" 참조).

---

## 4. Approach §2 — release-drafter-cleanup Mechanism (resolves OQ3)

**채택 방식**: GitHub Actions scheduled workflow (cron) + `gh release` CLI + `actions/github-script@v7` 보조.

설계:
- 트리거: `schedule: cron: '0 9 * * MON'` (매주 월요일 09:00 UTC; KST 18:00) + `workflow_dispatch` (수동 호출 가능)
- 권한: `contents: write` (release delete 권한 필요), `pull-requests: read`
- 핵심 단계:
  1. `gh release list --json tagName,isDraft,createdAt -L 100` 으로 최신 draft 100개 fetch
  2. `jq` 로 `isDraft == true && (now - createdAt) > 30d` 필터
  3. 각 stale draft 에 대해 `gh release delete <tag> --yes` (Release Drafter 가 만든 draft 는 tagName 이 `vNext` 또는 미설정일 수 있음 — `tagName == ""` 케이스 처리 필요)
  4. `release-drafter-cleanup-log.md` 에 timestamp + closed draft 리스트 append (AC-CIAUT-010)

이유 (vs `actions/github-script@v7` 단독):
- `gh` CLI 가 GitHub Actions runner 에 사전 설치되어 있어 추가 의존성 0
- `gh release delete` 는 idempotent + dry-run flag 지원 (`--yes` 생략 시 prompt → 실수 방지)
- `actions/github-script@v7` 는 보조 (jq 출력 검증, log markdown 생성 등 텍스트 작업) 로만 사용

**대안 (rejected)**:
- `actions/github-script@v7` 단독으로 octokit API 직접 호출 → 더 복잡한 권한 관리 + token scope 검토 필요
- 외부 서비스 (예: probot) → external dependency 도입; `feedback_release_no_autoexec.md` 정신과 충돌 가능

**보호 장치**:
- dry-run 환경변수 (`MOAI_CLEANUP_DRY_RUN=1`) 로 `workflow_dispatch` 시 실제 삭제 안 함 (검증용)
- 30일 임계값 + draft 가 release branch 와 연관된 경우 예외 (active release 진행 중일 때 보호) — `gh release view <tag> --json targetCommitish` 로 release/* branch 매칭 확인

---

## 5. Approach §3 — Makefile ci-disable Target (resolves OQ4)

**채택 방식**: `yq -i` (yaml editor) 를 사용한 in-place workflow trigger neutralization + git commit.

설계:
```makefile
ci-disable: ## Disable a workflow (set on: workflow_dispatch only). Usage: make ci-disable WORKFLOW=name
	@test -n "$(WORKFLOW)" || { echo "Usage: make ci-disable WORKFLOW=<workflow-basename>"; exit 1; }
	@test -f .github/workflows/$(WORKFLOW).yml || { echo "Not found: .github/workflows/$(WORKFLOW).yml"; exit 1; }
	yq -i '.on = {"workflow_dispatch": null}' .github/workflows/$(WORKFLOW).yml
	git add .github/workflows/$(WORKFLOW).yml
	git commit -m "chore(ci): disable $(WORKFLOW) (workflow_dispatch only)"
	@echo "Disabled $(WORKFLOW). Re-enable: edit .github/workflows/$(WORKFLOW).yml on: section."
```

이유 (vs sed/awk in-place YAML edit):
- `sed` 로 `on:` 블록을 주석화하는 방식은 multi-line YAML block 에서 brittle (indented mapping vs flow style 처리 불가)
- `yq` 는 YAML AST-aware → workflow 의 다른 필드 (jobs, env, permissions) 를 손상시키지 않음
- `yq` 가 macOS+Linux 에 모두 가용 (`brew install yq`, `apt install yq`); CI runner 도 사전 설치 (`actions/setup-yq` 또는 ubuntu-latest 기본)

**대안 (rejected)**:
- `sed` 주석화 → multi-line `on:` 블록에 fragile, 들여쓰기 손상 위험
- 파일 이동 (`.github/workflows/<name>.yml` → `.github/workflows/disabled/<name>.yml`) → §3에서 reject한 패턴 (GitHub auto-discovery 불가); §3 W4-T01 의 `.optional.yml` 패턴과도 의미 충돌
- 파일 삭제 → re-enable 시 git history 의존; reversibility 약함

**보호 장치**:
- `yq` 미설치 환경 fallback: target 첫 줄에 `command -v yq >/dev/null || { echo "Install yq: brew install yq"; exit 1; }` 추가
- Pre-commit hook (Wave 1 산출물) 이 commit 메시지 형식 (`chore(ci): disable <name>`) 검증
- Re-enable 명령어는 별도 target 추가하지 않음 (사용자가 git revert 또는 수동 yq 편집 — over-engineering 방지)

---

## 6. Approach §4 — required-checks.yml Validation (W4-T05)

**채택 방식**: 검증 스크립트 + manual replay (no new code).

W4-T01 의 rename 후 `.github/required-checks.yml` `auxiliary:` 리스트는 다음 상태가 되어야 한다:

```yaml
auxiliary:
  - claude-code-review        # → 실제 파일은 .github/workflows/claude-code-review.optional.yml
  - docs-i18n-check           # 실제 파일 .github/workflows/docs-i18n-check.yml (continue-on-error 추가)
  # llm-panel removed: file does not exist on origin/main; placeholder only
```

검증 단계:
1. `yq '.auxiliary[]' .github/required-checks.yml` 로 리스트 추출
2. 각 항목에 대해 GitHub workflow 의 `name:` 필드 매칭 확인:
   - `gh workflow list --json name,path` 로 활성 workflow fetch
   - 파일명 (`<name>.optional.yml` 또는 `<name>.yml`) 과 SSoT 항목명 매칭
3. SSoT `branches.main.contexts` 와 `branches.release/*.contexts` 에 auxiliary 항목이 포함되지 않음을 검증 (`comm -12` set intersection)
4. 위 검증을 `scripts/ci-mirror/validate-required-checks.sh` (Wave 1 framework 활용) 또는 Makefile `verify-required-checks` target 으로 자동화

**Phase 2 deliverable**: validation 스크립트는 W4-T05 의 일부로 작성되며 `make ci-local` 의 lint 단계에 추가됨 (Wave 1 framework 회귀 안 시킴, additive).

---

## 7. Decisions (OQ1-OQ6 Resolutions)

### OQ1: llm-panel.yml mapping

**Decision**: 옵션 (c) — `auxiliary:` 리스트에서 `llm-panel` 항목 제거 (deprecated placeholder cleanup).

**Rationale**:
- `git ls-tree origin/main .github/workflows/` 결과 `llm-panel.yml` 파일 부재 확인 (2026-05-08)
- `gh workflow list` 출력에서 `.github/workflows/llm-panel.yml` 가 active 로 표시되는 것은 GitHub 의 cached workflow registry — 파일은 이미 삭제됨
- `internal/template/templates/.github/workflows/llm-panel.yml.tmpl` 는 user-project 용 템플릿 (사용자 프로젝트에 LLM panel 코멘트 봇을 배포할 때 사용); moai-adk-go 자체 CI 와 무관
- `claude.yml` (Claude Code @claude responder) 와 `review-quality-gate.yml` (Claude Code Review severity parser) 는 별개 목적 — `llm-panel` 으로 매핑하지 않음
- SSoT 에서 deprecated 항목 잔존 시 W4-T05 검증 단계가 false positive (매칭 실패) 발생 → 즉시 정리

**Action**: W4-T01 수행 시 `.github/required-checks.yml` `auxiliary:` 리스트에서 `llm-panel` 항목 제거 + 주석 추가 (`# llm-panel.yml: removed 2026-05-08, placeholder only`).

### OQ2: optional/ subdirectory trigger semantics

**Decision**: filename suffix 변경 (`<name>.optional.yml`) 채택; 서브디렉터리 이동 reject.

**Rationale**: GitHub Actions auto-discovery 는 `.github/workflows/` 직속 파일만 처리 (official behavior, docs.github.com/actions). 서브디렉터리 이동 시 workflow 영구 비활성화. `.optional.yml` suffix 는 (a) 트리거 보존, (b) 시각 신호 제공, (c) tooling-friendly (glob `*.optional.yml`).

**Action**: W4-T01 에서 `claude-code-review.yml` → `claude-code-review.optional.yml` (git mv). `llm-panel.yml` 은 부재로 인해 행위 없음.

### OQ3: release-drafter-cleanup mechanism

**Decision**: GitHub Actions scheduled workflow + `gh release` CLI 단독 (옵션 a).

**Rationale**: `gh` CLI 사전 설치 + idempotent + 추가 의존성 0. `actions/github-script@v7` 는 jq 출력 정렬/로그 markdown 생성 등 보조 단계로만 사용 (octokit 직접 호출은 권한 검토 부담 큼).

**Action**: W4-T03 에서 `.github/workflows/release-drafter-cleanup.yml` 신규 작성 — 주 1회 cron, 30일+ stale draft delete, log append, dry-run env var.

### OQ4: Makefile ci-disable mechanism

**Decision**: `yq -i` AST 편집 방식 (옵션 b).

**Rationale**: sed/awk YAML 편집은 multi-line `on:` 블록에서 fragile. yq 는 AST-aware → 다른 필드 손상 0. macOS+Linux 가용. 파일 이동 방식 (옵션 c) 은 OQ2 의 `.optional.yml` 패턴과 의미 충돌.

**Action**: W4-T04 에서 Makefile `ci-disable` target 추가 + yq 미설치 fallback 메시지.

### OQ5: Template-First mirror

**Decision**: 본 Wave 산출물은 `internal/template/templates/` 미러 불필요.

**Rationale**:
- `.github/workflows/*.yml` 는 moai-adk-go 자체 dev project 의 CI 정의 — user project 에 배포되지 않음 (CLAUDE.local.md §1 dev-only)
- `internal/template/templates/.github/workflows/` 는 user-project 용 별개 템플릿 (claude-code-review.yml.tmpl 등 `.tmpl` 확장자 사용; moai-adk-go 의 ci.yml/codeql.yml 은 미러되지 않음을 확인)
- `Makefile` 도 dev-only (user project 는 자체 build system 사용)
- `.github/required-checks.yml` 만 부분 미러됨 (Wave 1 산출물; user project 가 branch protection 적용 시 SSoT 로 사용); Wave 4 의 `auxiliary:` 리스트 변경은 W4-T05 에서 user-facing template 측에도 반영 필요 여부 검토 (현재 `internal/template/templates/.github/required-checks.yml` 은 user project 의 일반적 auxiliary 셋트로 작성됨 — Wave 4 의 dev-project-specific 변경은 미러 안 함)

**Action**: Phase 2 (manager-tdd) 위임 시 "no template mirror required for `.github/workflows/*` and `Makefile`" 명시. 단, `.github/required-checks.yml` 변경은 Phase 2 가 user-template 측에 반영 필요 여부를 별도 판단 (default: 미러 안 함).

### OQ6: AC mapping

**Decision**: AC-CIAUT-009 + AC-CIAUT-010 가 Wave 4 primary AC; 추가 SSoT 검증을 W4-T05 의 acceptance 로 묶음 (별도 AC 신설 없음 — Wave 4 scope 외).

**Rationale**:
- `acceptance.md` §5 (T4) 에 명시된 AC 는 정확히 AC-CIAUT-009 (auxiliary non-blocking) 와 AC-CIAUT-010 (Release Drafter stale cleanup) 두 개
- W4-T05 (required-checks.yml 검증) 는 AC-CIAUT-021 (SSoT 검증) 의 부분 집합 — AC-CIAUT-021 은 Wave 1 의 primary AC 로 이미 매핑됨; Wave 4 는 SSoT 의 변경 후 무결성 유지 책임만 짐
- `auxiliary:` 리스트 변경에 따른 SSoT 무결성은 cross-wave 검증 (Wave 1 W1-T07 + Wave 2 W2-T04 + Wave 4 W4-T05 가 같은 SSoT 를 read/write) 으로 보장; 별도 AC 신설 시 scope 비대화

**Action**: tasks-wave4.md AC Mapping 표에 W4-T01/T02/T03/T05 → AC-CIAUT-009, W4-T03 → AC-CIAUT-010, W4-T04/T05 → cross-wave SSoT 무결성 (AC-CIAUT-021 transitive) 매핑.

---

## 8. Risks & Mitigations

| ID | Risk | Mitigation |
|----|------|-----------|
| W4-R1 | `.optional.yml` rename 후 GitHub UI 의 PR check 표시명 변경 → 기존 branch protection rule 의 `required_status_checks.contexts` 와 mismatch | check 표시명은 workflow 파일의 `name:` 필드 기반 (filename 무관) — rename 만으로 영향 없음을 `gh workflow list --json name,path` 로 사전 검증 |
| W4-R2 | release-drafter-cleanup 이 active release 의 draft 를 잘못 삭제 | 30일 임계값 + `targetCommitish` 가 `release/*` branch 를 가리키는 경우 예외 + dry-run env var 로 첫 1주 검증 |
| W4-R3 | `yq` 미설치 환경에서 `make ci-disable` 실패 | target 첫 줄에 `command -v yq >/dev/null` 체크 + 명확한 install 안내 메시지 |
| W4-R4 | docs-i18n-check `continue-on-error: true` 추가 후 일부 reviewer 가 advisory 결과를 놓침 | Phase 2 에서 advisory PR comment 의 markdown header 명시적으로 "ADVISORY ONLY — NOT BLOCKING" 표기 + sticky comment 패턴 유지 |
| W4-R5 | SSoT 의 `auxiliary:` 항목과 실제 workflow `name:` 매칭 검증 누락 → 후속 Wave 에서 watch loop 가 misclassify | W4-T05 검증 스크립트가 `make ci-local` 단계에 통합되어 회귀 시 즉시 차단 |
| W4-R6 | release-drafter-cleanup workflow 자체가 cron 미작동 (GitHub Actions 60일 inactive 제약) | repo 가 active 한 동안 cron 정상 작동; inactive 60일 임계 도달 시 user 가 수동 dispatch 가능 (workflow_dispatch 트리거 보존) |
| W4-R7 | claude-code-review.optional.yml 로 rename 후 `paths-ignore: [".github/**"]` 패턴 매칭 영향 (자기 자신 변경 시 무한 루프 방지 fragile) | rename 자체는 trigger filter 와 무관 (filter 는 변경된 파일 path 기준; workflow 파일 자체는 항상 trigger 대상) — 회귀 없음 확인 |

---

## 9. Verification Plan

### Per-Task Verification

각 task 별 검증 절차:

- **W4-T01** (rename): `git mv` 후 `gh workflow list` 출력에서 새 파일명으로 등재 확인; `pull_request: [opened]` 트리거가 보존되는지 PR open 시 dispatch 확인
- **W4-T02** (docs-i18n-check non-blocking): Phase 2 가 `continue-on-error: true` 추가 후 의도적 fail 케이스 (advisory 메시지) 가 PR merge 차단 안 함을 manual replay 검증
- **W4-T03** (release-drafter-cleanup): `act` (https://github.com/nektos/act) 또는 GitHub Actions `workflow_dispatch` 수동 호출로 dry-run; 실제 stale draft 삭제는 첫 cron 실행 (1주 후) 모니터링
- **W4-T04** (Makefile ci-disable): test fixture workflow 생성 후 `make ci-disable WORKFLOW=test-fixture` 실행; `yq` 결과가 `on: workflow_dispatch: null` 로 변경되었는지 검증; commit 메시지 형식 확인
- **W4-T05** (SSoT validation): `scripts/ci-mirror/validate-required-checks.sh` 또는 `make verify-required-checks` 실행; auxiliary 항목과 실제 workflow `name:` 매칭 0 mismatch

### Wave-Level DoD Verification

- `sh -n` (shell syntax check) on Makefile target body
- `yamllint -s` on all modified `.yml` files (release-drafter-cleanup.yml, claude-code-review.optional.yml, docs-i18n-check.yml)
- `yq eval '.' <file>` 로 YAML well-formedness 검증
- `gh workflow list` 출력에 모든 workflow 정상 등재
- 5-PR sweep replay (acceptance.md AC-CIAUT-009): 의도적으로 docs-i18n-check 를 fail 시킨 PR 이 ready-to-merge 유지
- `make ci-local` 통과 (Wave 1 framework 회귀 없음)
- `git log --oneline -5` 에 Conventional Commits + 🗿 MoAI co-author trailer 확인

### Manual Replay Scenarios

- **AC-CIAUT-009 replay**: feature branch 에서 docs-site/* 파일 수정 + 의도적 i18n drift 도입 → PR 생성 → docs-i18n-check fail (advisory) → required check 모두 green 시 ready-to-merge 상태 확인
- **AC-CIAUT-010 replay**: 30일+ stale draft 1개를 수동으로 생성 후 `gh workflow run release-drafter-cleanup.yml` dispatch → log 에 close 항목 기록 확인

---

## 10. TRUST 5 Targets (adapted for CI/CD config)

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | "verify-via-replay" — Go 코드 부재로 unit test 대체. structural test (yamllint/yq parse) + manual replay scenarios (AC-CIAUT-009/010) | `yamllint .github/workflows/*.yml`; manual replay 결과를 `.moai/reports/wave4-replay/` 에 기록 |
| **Readable** | yaml indent 2-space + workflow `name:` field 명시 + Makefile target에 `## description` 주석 | `yamllint -d "{rules: {indentation: {spaces: 2}}}"`; `make help` 출력에 ci-disable 표시 |
| **Unified** | yaml 스타일이 기존 `.github/workflows/*.yml` 컨벤션 따름 (boolean unquoted, `on:` 블록 sequence 형식); Makefile target naming `<verb>-<noun>` (ci-local 패턴) | 코드리뷰 + diff 검토; 새 target 명 `ci-disable` 이 기존 `ci-local` 과 동일 prefix |
| **Secured** | `gh release delete` 가 stale 한정 (30일 + non-active-release) — 활성 release draft 보호; `MOAI_CLEANUP_DRY_RUN=1` env var 로 검증 first; release-drafter-cleanup workflow 는 `permissions: contents: write` 명시 (least privilege) | workflow yaml 의 `permissions:` 블록 검증; dry-run 결과를 첫 1주 모니터링 |
| **Trackable** | 모든 commit 이 SPEC-V3R3-CI-AUTONOMY-001 W4 reference 포함; Conventional Commits + 🗿 MoAI co-author trailer; `release-drafter-cleanup-log.md` 가 audit trail 보존 | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W4'`; cleanup log 파일 존재 확인 |

---

## 11. Per-Wave DoD Checklist

- [ ] 모든 5개 W4 task 완료 (W4-T01..T05)
- [ ] `git mv` 로 `claude-code-review.yml` → `claude-code-review.optional.yml` rename (trigger 보존 확인)
- [ ] `.github/required-checks.yml` `auxiliary:` 리스트에서 `llm-panel` 제거 (deprecated placeholder)
- [ ] `docs-i18n-check.yml` 에 job-level `continue-on-error: true` 추가 + advisory comment header 갱신
- [ ] `.github/workflows/release-drafter-cleanup.yml` 신규 작성 (주 1회 cron, gh release CLI, dry-run env var)
- [ ] `Makefile` 에 `ci-disable WORKFLOW=<name>` target 추가 (yq AST edit + 자동 commit)
- [ ] W4-T05 검증 스크립트 (`scripts/ci-mirror/validate-required-checks.sh` 또는 Makefile `verify-required-checks` target) 작성
- [ ] `make ci-local` 통과 (Wave 1 framework 회귀 없음)
- [ ] `yamllint .github/workflows/*.yml` 통과
- [ ] AC-CIAUT-009 manual replay (docs-i18n-check fail PR ready-to-merge) 통과
- [ ] AC-CIAUT-010 manual replay (release-drafter-cleanup workflow_dispatch dry-run) 통과
- [ ] `gh workflow list` 출력에 rename 후 workflow 정상 등재
- [ ] No release/tag automation 도입 (cleanup workflow 는 draft delete 만, release create 안 함)
- [ ] PR labeled with `type:chore`, `priority:P1`, `area:ci`
- [ ] Conventional Commits + 🗿 MoAI co-author trailer 모든 commit
- [ ] CHANGELOG.md 에 Wave 4 머지 entry

---

## 12. Honest Scope Concerns

1. **GitHub Actions 의 cached workflow registry 와 actual workflow file mismatch**: `gh workflow list` 가 origin/main 에 부재한 `llm-panel.yml` 을 active 로 표시하는 현상은 GitHub 측 캐시 stale 가능성 — Wave 4 머지 후 GitHub UI 가 자동 갱신되지 않을 수 있음. Mitigation: PR description 에 "expect GitHub UI caching may show llm-panel as removed within 24h" 명시; 사용자 confused 시 manual workflow disable 안내.

2. **`.optional.yml` suffix 가 GitHub UI 에서 시각적으로 구분되지 않음**: PR check tab 은 workflow 의 `name:` field 만 표시 (filename 노출 안 함). reviewer 는 `claude-code-review.optional.yml` 의 결과를 advisory 로 인식하지 못할 수 있음. Mitigation: workflow `name:` field 자체에 "(advisory)" suffix 추가 — 단, 이는 별도 follow-up commit 으로 처리 (Wave 4 scope 는 filename rename 만).

3. **release-drafter-cleanup 의 30일 임계값이 일부 long-running release branch 에 false positive 발생 가능**: 6주짜리 release/v3.0.0 가 진행 중이고 그 draft 가 30일 임계 도달 시 활성 상태로 잘못 삭제될 수 있음. Mitigation: `targetCommitish` 검증 (release/* 브랜치 매칭 시 예외) + dry-run 첫 1주 모니터링 + 30일 임계값을 `MOAI_CLEANUP_THRESHOLD_DAYS` env var 로 외부화 (default 30).

4. **Makefile `ci-disable` 의 자동 commit 이 사용자 직접 편집 의도와 충돌 가능**: 사용자가 다른 변경사항과 함께 disable 하려는 경우 자동 commit 이 분리된 commit 만들어 stash/rebase 부담. Mitigation: `--no-commit` flag 지원 추가 검토 (Phase 2 자유도) — 또는 디폴트를 staged-only 로 변경 (`git add` 만 수행, commit 은 사용자 책임).

5. **W4-T05 validation script 가 Wave 1/2 framework 와 결합하나 Wave 4 에서 신규 작성**: framework 위치 불명확. Mitigation: validation script 는 `scripts/ci-mirror/validate-required-checks.sh` 에 위치하여 Wave 1 framework 와 인접; Wave 1 의 `lib/<lang>.sh` 패턴과 다른 namespace (validation vs mirror) — Phase 2 가 디렉터리 구조 명확화 필요.

---

## 13. Out-of-Scope (Wave 5/6/7 Deferral, Advanced Features)

- Wave 5 (T6 worktree state guard) — 독립적, 별도 Wave
- Wave 6 (T7 i18n validator) — 독립적, 별도 Wave
- Wave 7 (T8 BODP) — 독립적, 별도 Wave
- claude-code-review 의 fix 시도 (org quota 증설, 다른 LLM 으로 대체) — 사용자 별도 SPEC 필요 (Wave 4 는 disable/rename 만)
- llm-panel.yml 재도입 — 별도 SPEC + 사용자 의사결정 필요
- workflow `name:` field 에 "(advisory)" suffix 추가 — Phase 2 follow-up commit 또는 별도 cleanup wave
- release-drafter-cleanup 의 backfill (이미 80+ stale draft 가 있는 경우 첫 실행에서 일괄 삭제) — dry-run 1주 후 사용자 명시 승인 필요
- 16-language neutrality 검증 — Wave 4 산출물은 GitHub Actions YAML (언어 무관) + Makefile (Go-specific 환경) → 언어 중립성 검증 불필요
- Cross-platform Windows variant of Makefile target — Wave 1/2/3 패턴 계승 (Windows 사용자는 git-bash + GNU make 가정)
- release-drafter-cleanup 의 PR comment notification (close 시 PR 작성자 알림) — over-engineering, 사용자 요청 시 추가

---

## 14. Honest Methodology Adaptation Note

Wave 4 deliverables 는 GitHub Actions YAML + Makefile + bash 검증 스크립트로 구성되어 **Go unit test 가 부적절**하다. quality.yaml 의 `development_mode: tdd` 는 본 Wave 에서 다음과 같이 적응된다:

- **RED 대체**: AC-CIAUT-009/010 manual replay scenario 를 사전 작성 (acceptance.md §5 이미 정의됨)
- **GREEN 대체**: Phase 2 implementation (yaml 작성, Makefile target 추가)
- **REFACTOR 대체**: yamllint + sh -n + manual replay 통과 후 코드 정리

검증은 다음 4 layer 로 구성:
1. **Syntax**: `sh -n` (Makefile body), `yamllint -s` (yaml files), `yq eval '.' <file>` (YAML AST)
2. **Static**: `gh workflow list` 매칭, `.github/required-checks.yml` SSoT 정합성 (W4-T05)
3. **Dynamic**: `act` 또는 GitHub Actions `workflow_dispatch` 수동 호출 (release-drafter-cleanup.yml 검증)
4. **Replay**: AC-CIAUT-009/010 manual replay (reviewer + 작성자 회고)

본 wave 는 Wave 1/2/3 의 Go 코드 + skill 패턴과 다름 — Phase 2 위임 시 manager-tdd 가 본 적응을 인지하고 기존 TDD 사이클 강제 적용을 회피해야 함을 명시.

---

## 15. Template-First Note

[HARD] Wave 4 의 모든 산출물 (`.github/workflows/*`, `.github/required-checks.yml`, `Makefile`) 은 dev-only — **`internal/template/templates/` 미러 불필요**.

검증:
- `internal/template/templates/.github/workflows/` 는 user-project 용 별도 템플릿 (`*.tmpl` 확장자) 만 포함; moai-adk-go 자체 CI workflow 와 무관
- `Makefile` 은 dev-only (user project 는 자체 build system 사용)
- `.github/required-checks.yml` 는 부분 미러 (Wave 1 산출); user-project template (`internal/template/templates/.github/required-checks.yml`) 은 generic 셋트로 작성됨 — Wave 4 의 dev-project-specific 변경 (auxiliary 항목 정리) 은 미러하지 않음 (user-project 의 auxiliary 는 user 환경마다 다름)

Phase 2 manager-tdd 위임 시 "no template mirror required for Wave 4 deliverables" 명시 필요. 만일 Phase 2 에서 다른 판단이 필요하면 manager-tdd 가 blocker report 로 반환.

---

Version: 0.1.0
Status: pending Phase 2 (manager-tdd 위임 대기)
Last Updated: 2026-05-08
