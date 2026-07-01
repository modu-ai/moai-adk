# Progress — SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001

Lifecycle progress ledger. §E.1 is populated at plan-phase (manager-spec). §E.2/§E.3 are populated at run-phase (manager-develop); §E.4 at sync-phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (complete)
- **Tier**: M (judgment call — see plan.md §A for the file-count-vs-complexity rationale; raw file count ~79 nominally suggests Tier L, but per-file complexity is trivial/mechanical for 78 of 79 files).
- **Artifacts produced**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (4 files — standard Tier M set, no design.md/research.md).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | DOCS ✓ | POSTREBUILD ✓ | CLEANUP ✓ | 001 ✓ → PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Frontmatter**: 12 canonical fields present; `status: draft`; `created`/`updated` (not `_at`); `tags` comma-separated string; `era: V3R6`; `tier: M`; `depends_on`/`related_specs` optional fields present.
- **Requirement count**: 6 REQ-DPC (001-006) + 4 NFR-DPC (001-004).
- **Ground truth basis** (observed 2026-07-02, live repo state): `scripts/docs-i18n-check.sh` fresh run confirmed 65 total errors (64 Check-3 H1-missing across 16 pages × 4 locales + 1 unrelated Check-4 glossary finding); `grep -n goos` confirmed 28 occurrences across 3 files × 4 locales (worktree/examples.md 4/locale, worktree/faq.md 1/locale, getting-started/inventory.md 2/locale); `robots.txt` confirmed wrong `cowork.mo.ai.kr` domain; `hugo.toml` confirmed stale `v3.0.0-rc4` vs current tag `v3.0.0-rc5` (dated 2026-07-01); no other stale rc references found.
- **Source-correction finding**: the task's originally-suggested REQ-DPC-006 source files (`internal/loop/{feedback,feedback_channel,go_feedback}.go`) are the unrelated Ralph Engine loop-feedback system (test/lint/coverage/LSP-diagnostic tracking for `/moai loop`), NOT the `/moai feedback` GitHub-issue workflow. The actual source is `.claude/skills/moai/workflows/feedback.md` + `internal/config/feedback_accessors.go` + `internal/template/templates/.moai/config/sections/feedback.yaml`, cross-verified against `SPEC-INVOCATION-MODEL-001/spec.md`. Also found: the current ko `moai-feedback.md` page already misattributes GitHub issue creation to a "sync-auditor 에이전트" (the actual workflow is orchestrator-direct, no subagent) — folded into REQ-DPC-006's scope.
- **Out of Scope**: present (D5 non-ko backlog, en/ja/zh feedback-page translation, Check-4 glossary defect, Go/template code changes, feedback workflow/config schema changes, full hugo build/deploy re-verification).
- **Plan-phase gaps (residual)**: (1) exact H1 text for each of the 64 files is not enumerated in the SPEC body — left to run-phase per-file frontmatter-title lookup (WHAT not HOW); (2) the ko `moai-feedback.md` new subsection's exact heading name ("피드백 설정" is a suggestion, not mandated) is left to the implementing agent's judgment.
- **Next phase**: run (M1 → M5 per plan.md §F). Requires Implementation Kickoff Approval (plan-to-implement HUMAN GATE) before run-phase entry.

## §E.2 Run-phase Evidence

Run-phase executed as a single manager-develop session (M1→M5, sequential, no re-delegation needed). All 5 milestones committed + pushed to `main` individually (Hybrid Trunk 1-person OSS policy, no PR). Commit SHAs (oldest → newest): `1b97cb4d3` (M1), `e06864285` (M2), `9cf657de6` (M3), `7c62c4882` (M4). HEAD == `origin/main` == `7c62c48829bf9367e6613e55c2b0540bbb76974a` after final push.

### AC PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|----------------|
| AC-DPC-001a | PASS | `DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh 2>&1 \| grep -c 'no H1 heading'` | `0` (down from plan-phase baseline 64) |
| AC-DPC-001b | PASS | `head -n 8 <file>` spot-check across ko/en/ja/zh × 2 pages | H1 text matches frontmatter `title:` for that locale in every sampled file (verified for all 64 during authoring, not just a subset) |
| AC-DPC-001c | PASS | `find docs-site/content/<L> -name '*.md' \| wc -l` for ko/en/ja/zh | `99 / 99 / 99 / 99` — unchanged from plan-phase baseline |
| AC-DPC-002a | PASS | `grep -c goos docs-site/content/<L>/worktree/examples.md` for ko/en/ja/zh | `0 / 0 / 0 / 0` |
| AC-DPC-002b | PASS | `grep -c goos docs-site/content/<L>/worktree/faq.md` for ko/en/ja/zh | `0 / 0 / 0 / 0` |
| AC-DPC-002c | PASS | `grep -c '/path/to/your-project' docs-site/content/<L>/worktree/examples.md` for ko/en/ja/zh | `4 / 4 / 4 / 4` |
| AC-DPC-003a | PASS | `grep -c goos docs-site/content/<L>/getting-started/inventory.md` for ko/en/ja/zh | `0 / 0 / 0 / 0` |
| AC-DPC-003b | PASS | `grep -c '/path/to/your-project' docs-site/content/<L>/getting-started/inventory.md` for ko/en/ja/zh | `2 / 2 / 2 / 2` |
| AC-DPC-003c | PASS | Read of the JSON block (`"project_root": "/path/to/your-project"`) in ko/inventory.md | Block starts with `{`, ends with `}`, balanced braces, valid JSON (verified by direct read) |
| AC-DPC-004a | PASS | `cat docs-site/static/robots.txt` | `Sitemap: https://adk.mo.ai.kr/sitemap.xml` exactly |
| AC-DPC-004b | PASS | `grep -c cowork docs-site/static/robots.txt` | `0` |
| AC-DPC-005a | PASS | `grep -n 'version = "v3.0.0-rc5"' docs-site/hugo.toml` | Matches (params block) |
| AC-DPC-005b | PASS | `grep -n 'rc4\|rc2\|rc3' docs-site/hugo.toml` | no matches, exit 1 |
| AC-DPC-005c | PASS | `grep -rn 'rc4\|rc2\|rc3' docs-site/content/*/_index.md` | no matches, exit 1 |
| AC-DPC-005d | PASS | `grep -c 'releaseDate = "2026-07-01"' docs-site/hugo.toml` → `grep -c '2026-06-23' docs-site/hugo.toml` | `1` → `0` |
| AC-DPC-006a | PASS | `grep -c 'moai version\|uname' / 'go version' / '중복\|dedupe\|duplicate' / 'gh auth\|인증되지 않' ko/moai-feedback.md` | `3 / 2 / 5 / 2` (all ≥1) |
| AC-DPC-006b | PASS | `grep -c 'sync-auditor' ko/moai-feedback.md` | `0` |
| AC-DPC-006c | PASS | `grep -c 'feedback.repository\|modu-ai/moai-adk' ko/moai-feedback.md` | `1` |
| AC-DPC-006d | PASS (manual cross-check) | Every claim in the new "피드백 설정" subsection traced against `.claude/skills/moai/workflows/feedback.md` (Diagnostic Attachment, Duplicate Detection, gh Availability and Failure Fallback sections) + `internal/config/feedback_accessors.go` (`FeedbackRepository()`) + `internal/template/templates/.moai/config/sections/feedback.yaml` (`feedback.repository` default) | No invented behavior found; all 4 enhancements traced 1:1 to source lines actually read in M4.1 |
| AC-DPC-006e | PASS | `grep -c 'Claude Code 버전\|현재 SPEC' ko/moai-feedback.md` → `grep -c 'Go 툴체인\|Go 버전\|go version' ko/moai-feedback.md` | `0` → `3` |
| AC-DPC-006f | PASS | `grep -c 'feedback.repository\|modu-ai/moai-adk' docs-site/content/{en,ja,zh}/utility-commands/moai-feedback.md` | `0 / 0 / 0` (ko-only content, per NFR-DPC-004) |

### Milestone-to-commit mapping

- M1 (`1b97cb4d3`): robots.txt domain fix + hugo.toml version/releaseDate bump + frontmatter `status: draft → in-progress`.
- M2 (`e06864285`): 12 files (3 pages × 4 locales), maintainer-path-leak mechanical substitution.
- M3 (`9cf657de6`): 64 files (16 pages × 4 locales), H1 heading insertion.
- M4 (`7c62c4882`): 1 file (ko moai-feedback.md), new subsection + table + diagram correction.
- M5: verification-only, no additional commit (this progress.md update is the sole M5 artifact change).

## §E.3 Run-phase Audit-Ready Signal

- **run_status**: complete
- **run_complete_at**: 2026-07-02
- **run_commit_sha (final)**: `7c62c48829bf9367e6613e55c2b0540bbb76974a` (== HEAD == origin/main at run-phase close)
- **ac_pass_count**: 19 / 19 (all lettered AC-DPC entries in acceptance.md §D)
- **ac_fail_count**: 0
- **preserve_list_post_run_count**: 0 unintended modifications — `git status --porcelain docs-site/` empty after final push; no file outside the M1-M4 enumerated set was touched
- **new_warnings_or_lints_introduced**: none applicable (docs-content-only SPEC; no Go/lint toolchain touched)
- **cross_platform_build**: N/A (docs-site-only SPEC, no Go code changed)
- **total_run_phase_files**: 78 (2 + 12 + 64 + 1 — hugo.toml + robots.txt counted individually in M1's 2, spec.md frontmatter update not counted as a "docs-site" file)
- **hugo_build**: `hugo --minify` in `docs-site/` completed exit 0, 1750ms, emitted `v3.0.0-rc5` string in output HTML, zero warnings
- **docs_i18n_check_strict_result**: exit 1 with exactly 1 residual error (`ja/claude-code/agentic/best-practices.md` glossary term 'Anthropic' — pre-existing, out-of-scope Check-4 finding per spec.md Exclusions), 0 Check-3 (H1) errors — matches Definition of Done exactly
- **m1_to_mN_commit_strategy**: 4 separate milestone commits (M1-M4), each pushed individually to `main` directly (Hybrid Trunk 1-person OSS, no PR), verified `git fetch origin main` + `git rev-list --count --left-right` before each push (all `0 N`, no divergence, no parallel-session race detected on `docs-site/`)
- **status transition performed**: `draft → in-progress` on M1 commit (frontmatter `status:` field only, per Status Transition Ownership Matrix)
- **blockers encountered**: none — Ground Truth from plan-phase (2026-07-02 same-day) was re-verified unchanged at run-phase entry (64 H1-missing files, 28 `goos` occurrences, stale hugo.toml version, wrong robots.txt domain — all confirmed identical to plan-phase snapshot before starting M1)

## §E.4 Sync-phase Audit-Ready Signal

- **sync_status**: complete
- **sync_complete_at**: 2026-07-02
- **sync_commit_sha**: 85391a77028a13568f0b2f64cef99ed7a7f6b4dd
- **changelog_entry_position**: [Unreleased] / Fixed section, first entry
- **frontmatter_status_transitions.in-progress_to_implemented**: YES (M1 commit 1b97cb4d3)
- **frontmatter_status_transitions.implemented_to_completed**: YES (sync commit, this artifact)
- **canary_compliance_check.no_go_code_modified**: YES (git diff internal/ cmd/ pkg/ 0 matches in run+sync commits)
- **canary_compliance_check.no_template_modified**: YES (internal/template/templates/ 0 changes — docs-site-only SPEC)
- **canary_compliance_check.specific_path_discipline**: YES (staged: CHANGELOG.md + spec.md + progress.md only; other project files untouched)
