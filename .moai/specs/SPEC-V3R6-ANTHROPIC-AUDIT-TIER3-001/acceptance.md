---
id: SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001
title: "Acceptance — Anthropic Best-Practice Audit Tier 3 (F3+F9+F13)"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules + internal/spec"
lifecycle: spec-anchored
tags: "anthropic-best-practice, audit-tier-3, acceptance, tier-m"
tier: M
---

# Acceptance — SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001

본 acceptance 문서는 10개 mandatory AC를 enumerate한다. Tier M envelope 유지 위해 optional/nice-to-have는 별도 분리 안 함 (전체 mandatory).

---

## §D. AC Matrix (10 mandatory ACs)

### §D.1 Severity classification

- All 10 ACs are **mandatory** — failure of any one blocks run-phase completion
- No optional / no nice-to-have / no Tier-S-style minimal subset
- Each AC traceable to one or more REQ-AAT-### entries in `spec.md §C`

### §D.2 Traceability matrix

| AC | Linked REQ(s) | Verification method | Severity |
|----|--------------|---------------------|----------|
| AC-AAT-001 | REQ-AAT-001 | Filesystem check (ls) | mandatory |
| AC-AAT-002 | REQ-AAT-003, REQ-AAT-004 | wc -l + grep section markers | mandatory |
| AC-AAT-003 | REQ-AAT-005 | Custom diff script | mandatory |
| AC-AAT-004 | REQ-AAT-007 | grep struct definition | mandatory |
| AC-AAT-005 | REQ-AAT-008, REQ-AAT-009 | go test PASS scenarios | mandatory |
| AC-AAT-006 | REQ-AAT-008 | go test FAIL detections | mandatory |
| AC-AAT-007 | REQ-AAT-010, REQ-AAT-011 | go test unreachable-git + observation-only | mandatory |
| AC-AAT-008 | REQ-AAT-013 | diff byte-identical | mandatory |
| AC-AAT-009 | REQ-AAT-015 | 4-platform build + vet + lint | mandatory |
| AC-AAT-010 | REQ-AAT-014 (scope disjointness) | git diff --name-only | mandatory |

### §D.3 Given-When-Then scenarios

#### AC-AAT-001 — 5 subdirectory CLAUDE.md 파일 존재

**Given** run-phase M1 commit has landed on main.

**When** the verifier executes `ls -1 internal/cli/CLAUDE.md internal/template/CLAUDE.md internal/spec/CLAUDE.md internal/hook/CLAUDE.md internal/config/CLAUDE.md` from project root.

**Then** all 5 files MUST be present and accessible (exit code 0, 5 lines of output, no "No such file or directory" error). REQ-AAT-001 satisfied.

#### AC-AAT-002 — 각 subdirectory CLAUDE.md size 60-200 LOC + 4 required sections

**Given** AC-AAT-001 passes.

**When** the verifier executes for each of the 5 files:
```bash
LC=$(wc -l < internal/<module>/CLAUDE.md)
test "$LC" -ge 60 && test "$LC" -le 200
grep -c '^## \(.*Purpose\|.*Key files\|.*Conventions\|.*Cross-references\|.*References\)' internal/<module>/CLAUDE.md
```

**Then** the wc -l result MUST be in [60, 200] range for each of 5 files (REQ-AAT-004), AND the grep count for the 4 required section types (Purpose / Key files or Packages / Conventions / Cross-references) MUST be ≥4 for each file (REQ-AAT-003 — section coverage). Total: 5 file × 2 sub-checks = 10 sub-conditions, all PASS required.

#### AC-AAT-003 — root CLAUDE.md vs subdirectory diff <50% overlap

**Given** AC-AAT-001 passes.

**When** the verifier executes a custom diff script for each subdirectory file:
```bash
for mod in cli template spec hook config; do
  ROOT_LC=$(wc -l < CLAUDE.md)
  SUB_LC=$(wc -l < internal/$mod/CLAUDE.md)
  OVERLAP=$(diff -u CLAUDE.md internal/$mod/CLAUDE.md | grep -E '^[+-]' | grep -v '^[+-][+-][+-]' | wc -l)
  echo "$mod: root_lc=$ROOT_LC sub_lc=$SUB_LC overlap_lines=$OVERLAP"
  PCT=$((OVERLAP * 100 / (ROOT_LC + SUB_LC)))
  test "$PCT" -lt 50  # PASS if <50% overlap
done
```

**Then** each module's overlap percentage MUST be <50% (REQ-AAT-005 — no duplication of root content).

#### AC-AAT-004 — `OwnershipTransitionRule` struct + Check() method 존재

**Given** run-phase M2 commit has landed on main.

**When** the verifier executes:
```bash
grep -n 'type OwnershipTransitionRule struct' internal/spec/lint.go
grep -n 'func.*OwnershipTransitionRule.*Check' internal/spec/lint.go
grep -n 'OwnershipTransitionInvalid' internal/spec/lint.go
```

**Then** each grep MUST return at least 1 match line (struct definition + Check method + finding code constant). REQ-AAT-007 satisfied.

#### AC-AAT-005 — 7 canonical transition lint PASS scenarios

**Given** M2 implementation complete.

**When** the verifier executes:
```bash
go test -v -run TestOwnershipTransitionRule_Pass ./internal/spec/
```

**Then** the test output MUST show at minimum 7 subtest names (one per canonical transition: `none_to_draft`, `draft_to_in_progress`, `in_progress_to_implemented`, `implemented_to_completed`, `any_to_superseded`, `any_to_archived`, `any_to_rejected`) AND all subtests PASS. REQ-AAT-008 + REQ-AAT-009 satisfied for the PASS half.

#### AC-AAT-006 — 5 lint FAIL scenarios (forbidden crossings + format mismatches)

**Given** M2 implementation complete.

**When** the verifier executes:
```bash
go test -v -run TestOwnershipTransitionRule_Fail ./internal/spec/
```

**Then** the test output MUST show at minimum 5 subtest names covering: (1) manager-docs modifying spec.md body (forbidden crossing per ARR-001), (2) commit message format mismatch (e.g., `chore` instead of `docs` for `in-progress → implemented`), (3) commit message subject lacking SPEC-ID, (4) multi-transition single-commit ambiguity, (5) frontmatter `status:` and commit history disagreement. All 5 MUST PASS (i.e., the rule correctly detects each failure pattern and emits `OwnershipTransitionInvalid` finding). REQ-AAT-008 satisfied for the FAIL half.

#### AC-AAT-007 — git log unreachable graceful degradation

**Given** M2 implementation complete.

**When** the verifier executes the test that simulates non-git environment:
```bash
go test -v -run TestOwnershipTransitionRule_UnreachableGit ./internal/spec/
```

**Then** the test MUST PASS, demonstrating that the rule emits `OwnershipTransitionUnreachable` (Severity Info) and continues without panic when `git log --follow` is unreachable (REQ-AAT-010). Additionally `grep -c 'os.Exec\|file_modify\|ioutil.WriteFile' internal/spec/lint.go` MUST return 0 (REQ-AAT-011 observation-only — no mutation primitives in the rule code).

#### AC-AAT-008 — schema doc cross-ref + template mirror byte-identical

**Given** run-phase M3 commit has landed on main.

**When** the verifier executes:
```bash
diff .claude/rules/moai/development/spec-frontmatter-schema.md \
     internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
echo $?
```

**Then** the diff output MUST be empty AND exit code MUST be 0 (byte-identical mirror per REQ-AAT-013 + CLAUDE.local.md §2 Template-First Rule). Additionally `grep -n 'OwnershipTransitionRule\|OwnershipTransitionInvalid' .claude/rules/moai/development/spec-frontmatter-schema.md` MUST return ≥2 matches (cross-ref subsection present).

#### AC-AAT-009 — 4-platform cross-build + vet + lint

**Given** all M1+M2+M3 commits landed.

**When** the verifier executes 6 commands:
```bash
GOOS=linux   GOARCH=amd64 go build ./...
GOOS=darwin  GOARCH=arm64 go build ./...
GOOS=darwin  GOARCH=amd64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
go vet ./...
golangci-lint run --timeout=2m
```

**Then** all 6 commands MUST exit 0. No new lint findings introduced by the SPEC's run-phase delta (pre-existing baseline counted separately). REQ-AAT-015 satisfied.

#### AC-AAT-010 — Disjoint scope verification (cross-SPEC isolation)

**Given** all 3 M commits landed.

**When** the verifier executes:
```bash
git diff --name-only $(git log --grep="SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001" --reverse --format='%H' | head -1)^..HEAD | sort | uniq > /tmp/aat_files.txt
```

**Then** `/tmp/aat_files.txt` MUST contain ONLY files listed in `plan.md §A.5 EXTEND` (10 file paths). The file MUST NOT contain any of the following forbidden paths (REQ-AAT-014):

```
.claude/rules/moai/core/agent-common-protocol.md
internal/governance/
.claude/agents/core/manager-spec.md
.claude/agents/core/manager-develop.md
.claude/agents/core/manager-docs.md
internal/template/templates/.claude/agents/core/manager-spec.md
internal/template/templates/.claude/agents/core/manager-develop.md
internal/template/templates/.claude/agents/core/manager-docs.md
.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/
.moai/research/
.moai/harness/
.moai/state/
.moai/cache/
.moai/config/sections/
i18n-validator
```

Verification:
```bash
for forbidden in \
  ".claude/rules/moai/core/agent-common-protocol.md" \
  "internal/governance/" \
  ".claude/agents/core/manager-spec.md" \
  ".claude/agents/core/manager-develop.md" \
  ".claude/agents/core/manager-docs.md" \
  ".moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/" \
  ".moai/research/" \
  ".moai/harness/" \
  ".moai/state/" \
  ".moai/cache/" \
  "i18n-validator" \
; do
  if grep -q "$forbidden" /tmp/aat_files.txt; then
    echo "VIOLATION: $forbidden touched"; exit 1
  fi
done
```

Exit code 0 = AC-AAT-010 PASS. REQ-AAT-014 satisfied.

---

## §D.4 Edge cases coverage

본 SPEC AC 명세는 다음 edge case들을 명시 검증한다:

| Edge case | Covered by AC | Rationale |
|-----------|---------------|-----------|
| 비-git 환경 fallback | AC-AAT-007 | OwnershipTransitionRule must not crash without git |
| Multi-transition single commit (e.g., `(none) → in-progress` in single SPEC creation commit) | AC-AAT-006 (FAIL #4) | rule MUST detect ambiguity |
| Closed SPEC (ARR-001) self-evaluation false-positive | plan.md §6 + M2 verification step 5 | rule must NOT false-flag closed SPECs |
| Subdirectory CLAUDE.md size boundary (59 or 201 LOC) | AC-AAT-002 | exact range [60, 200] enforcement |
| Template mirror drift (1-byte difference) | AC-AAT-008 | `diff` exit 0 strict |
| Cross-SPEC scope bleed (touching `.claude/rules/moai/core/agent-common-protocol.md`) | AC-AAT-010 | hard forbidden list |
| Cross-platform (windows/amd64) build failure | AC-AAT-009 | full 4-platform matrix |

---

## §D.5 Quality Gate Criteria (TRUST 5 mapping)

| TRUST pillar | Coverage in this SPEC's ACs |
|--------------|------------------------------|
| **Tested** | AC-AAT-005 (7 PASS) + AC-AAT-006 (5 FAIL) + AC-AAT-007 (graceful) = 13 explicit test scenarios for new code (`OwnershipTransitionRule`) |
| **Readable** | AC-AAT-002 (section structure) + AC-AAT-003 (root non-duplication) — subdirectory CLAUDE.md readability |
| **Unified** | AC-AAT-008 (template mirror byte-identical) — schema doc consistency |
| **Secured** | (n/a — this SPEC is documentation + lint rule, no security surface) |
| **Trackable** | AC-AAT-004 (struct + finding code grep-able) + AC-AAT-010 (commit history disjoint) — git-log-discoverable change set |

---

## §D.6 Definition of Done

본 SPEC은 다음 조건을 **모두** 만족할 때 run-phase complete로 간주된다:

- [ ] All 10 mandatory ACs PASS (AC-AAT-001..010)
- [ ] M1 + M2 + M3 commits landed on main (max 3 commits per §B.3.6)
- [ ] No file outside `plan.md §A.5 EXTEND` list modified (AC-AAT-010)
- [ ] Cross-platform build (4 platforms) + go vet + golangci-lint 0 new issues (AC-AAT-009)
- [ ] Coverage ≥85% for `internal/spec/` package (plan.md §5 E3)
- [ ] manager-develop self-verification report (Section E) returned to orchestrator
- [ ] Pre-spawn fetch verified clean (`0 0` or `0 N`) for each M commit push
- [ ] Sync-phase completion criteria (separate, post-run):
  - [ ] CHANGELOG.md `[Unreleased]` entry referencing this SPEC
  - [ ] All 4 SPEC artifact frontmatter `status:` transitioned `in-progress → implemented`
  - [ ] `progress.md §E.4 Sync-phase Audit-Ready Signal` populated

---

## §D.7 Forward-looking checks (NOT blocking, observation only)

본 SPEC scope 외이나 운영 중 관찰 가치 있음:

- **AC-FW-001 (observation)** — 운영 후 30일간 `OwnershipTransitionRule` finding 발생 빈도 측정. 빈발 시 (>1 per week) REQ-AAT-009 default-subset 검토 (시작은 4-transition로 축소 가능).
- **AC-FW-002 (observation)** — 5 subdirectory CLAUDE.md 도입 후 agent token consumption 변화 측정 (root + 1 subdirectory load vs root only). 효율성 ROI 검증.
- **AC-FW-003 (observation)** — sibling SPEC (COORD-001) 완료 후 `internal/governance/CLAUDE.md` 추가 SPEC 도입 시점 결정 (Tier 4 backlog).

위 3 forward-looking 항목은 본 SPEC AC 평가에 포함되지 않으며, 운영 관찰 결과는 별도 memory 또는 follow-up SPEC으로 기록.
