# SPEC-V3R6-CHANGELOG-CLEANUP-001 — Acceptance Criteria

본 SPEC 의 모든 acceptance criterion 은 binary verification command 로 검증된다 (exit 0 = PASS, exit !=0 = FAIL). 수동 해석 없음. 총 5개 AC (AC-CHL-001 ~ AC-CHL-005).

## AC-CHL-001 — Line 65 hallucination entry 전체 삭제

**Given** CHANGELOG.md 의 `[Unreleased]` 섹션에 §A.1 의 7개 환각 사실을 포함한 라인 65 SESSION-HANDOFF-AUTO-001 중복 엔트리가 존재한다,
**When** M1 milestone 이 완료된 후,
**Then** 7개 환각 사실 모두 CHANGELOG.md 에서 제거되어 있어야 한다.

**Binary verify command**:

```bash
# §A.1 7-row 카탈로그를 6 discriminating sub-condition 으로 매핑 (Path 필드는 sub-condition 1 이 unique 하게 cover; bare `internal/handoff/` 토큰은 line 65 단독으로 존재하지 않아 sub-condition 으로 부적합).
test "$(grep -c 'internal/handoff/{package,atomic_write,parser}' CHANGELOG.md)" -eq 0 && \
test "$(grep -c '10 files +556/-3' CHANGELOG.md)" -eq 0 && \
test "$(grep -c '15 functions race-safe' CHANGELOG.md)" -eq 0 && \
test "$(grep -c 'session-handoff.json' CHANGELOG.md)" -eq 0 && \
test "$(grep -c "Block 1-6.*감정" CHANGELOG.md)" -eq 0 && \
test "$(grep -c '\.supersede' CHANGELOG.md)" -eq 0
```

Exit 0 = PASS (6 discriminating sub-condition 모두 0건 매치 — §A.1 7-row 중 Path 필드는 sub-condition 1 이 unique cover, 나머지 6 facts 는 1:1 매핑). Exit !=0 = FAIL.

## AC-CHL-002 — Sibling AC 카운트 acceptance.md SSOT 일치

**Given** CHANGELOG.md 라인 59 (HOOK-ASYNC-EXPAND-001) 가 "AC 12/12 PASS" 로, 라인 61 (HOOK-CWD-LEAK-AUDIT-001) 가 "AC 8/8 PASS" 로 잘못 표기되어 있다,
**When** M2 milestone 이 완료된 후,
**Then** CHANGELOG.md 의 두 라인이 각각의 sibling acceptance.md SSOT 와 일치하도록 정정되어야 한다.

**Binary verify command**:

```bash
# Per-SPEC SSOT vs CHANGELOG entry 카운트 비교 (4 sibling 모두 검증)
HAE_SSOT=$(grep -oE 'AC-HAE-[0-9]+' .moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/acceptance.md | sort -u | wc -l | tr -d ' ')
HCWA_SSOT=$(grep -oE 'AC-HCWA-[0-9]+' .moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/acceptance.md | sort -u | wc -l | tr -d ' ')
HOI_SSOT=$(grep -oE 'AC-HOI-[0-9]+' .moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/acceptance.md | sort -u | wc -l | tr -d ' ')
CBD_SSOT=$(grep -oE 'AC-CBD-[0-9]+' .moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/acceptance.md | sort -u | wc -l | tr -d ' ')

# CHANGELOG.md 에서 잘못된 카운트 패턴은 0건이어야 PASS
# 토큰 순서 주의: CHANGELOG entry 형식은 "<SPEC-ID> ... AC X/X PASS" (SPEC-ID 가 카운트 앞)
test "$(grep -c 'AC 12/12 PASS' CHANGELOG.md)" -eq 0 && \
test "$(grep -c "HOOK-ASYNC-EXPAND.*AC ${HAE_SSOT}/${HAE_SSOT} PASS" CHANGELOG.md)" -ge 1 && \
test "$(grep -c "HOOK-CWD-LEAK-AUDIT.*AC ${HCWA_SSOT}/${HCWA_SSOT} PASS" CHANGELOG.md)" -ge 1 && \
test "$(grep -c "HOOK-OBSERVE-OPT-IN.*${HOI_SSOT}/${HOI_SSOT} AC PASS" CHANGELOG.md)" -ge 1 && \
test "$(grep -c "CI-BASELINE-DRIFT.*AC ${CBD_SSOT}/${CBD_SSOT} PASS" CHANGELOG.md)" -ge 1
```

Exit 0 = PASS (CHANGELOG 의 4 sibling 카운트 모두 acceptance.md SSOT 와 일치). Exit !=0 = FAIL.

## AC-CHL-003 — manager-develop-prompt-template B12 가드 삽입

**Given** `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B 에 B1-B11 만 존재한다,
**When** M3 milestone 이 완료된 후,
**Then** Section B 에 신규 bullet B12 "Sync-phase CHANGELOG emission discipline" 이 삽입되어 있어야 하며, B12 본문은 "Read implementation files" + `grep -c '<SPEC-ID>' CHANGELOG.md` 두 가드 명령을 모두 포함해야 한다.

**Binary verify command**:

```bash
test "$(grep -c 'B12. Sync-phase CHANGELOG emission' .claude/rules/moai/development/manager-develop-prompt-template.md)" -ge 1 && \
test "$(grep -c 'Read.*implementation files' .claude/rules/moai/development/manager-develop-prompt-template.md)" -ge 1 && \
test "$(grep -c "grep -c '<SPEC-ID>' CHANGELOG.md" .claude/rules/moai/development/manager-develop-prompt-template.md)" -ge 1 && \
test "$(grep -c 'SPEC-V3R6-CHANGELOG-CLEANUP-001' .claude/rules/moai/development/manager-develop-prompt-template.md)" -ge 1
```

Exit 0 = PASS (B12 heading + 2 가드 명령 + origin SPEC ID cross-reference 모두 존재). Exit !=0 = FAIL.

## AC-CHL-004 — Line 34 정확한 SESSION-HANDOFF-AUTO 엔트리 byte-identical 보존

**Given** CHANGELOG.md 라인 34-39 가 사전 sync commit `1deacc6b3` 에 의해 정확히 추가된 SESSION-HANDOFF-AUTO-001 엔트리이다,
**When** M1+M2+M3 milestone 전체가 완료된 후,
**Then** 라인 34-39 의 sha256 hash 가 baseline (M0 pre-flight 캡처) 와 byte-identical 이어야 한다.

**Binary verify command**:

```bash
# Phase 0 pre-flight 가 EXPECTED_SHA256 을 progress.md 에 capture 했다고 가정.
# 정확한 형식: progress.md 의 line 시작에 'expected_line_34_39_sha256: <64-hex>' (bulleted prose 금지)
EXPECTED_SHA256=$(grep -E '^expected_line_34_39_sha256:' .moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/progress.md | awk '{print $2}')
ACTUAL_SHA256=$(sed -n '34,39p' CHANGELOG.md | sha256sum | awk '{print $1}')
test -n "$EXPECTED_SHA256" && test "$EXPECTED_SHA256" = "$ACTUAL_SHA256"
```

Exit 0 = PASS (sha256 일치). Exit !=0 = FAIL.

**Note**: M0 pre-flight 가 progress.md 에 `expected_line_34_39_sha256: <hash>` 필드를 기록한다. run-phase 시작 시점의 라인 34-39 가 baseline 이며, 본 SPEC 의 어떤 milestone 도 이 영역을 byte-modify 하면 안 된다 (REQ-CHL-004 + REQ-CHL-005 의 binary 검증).

## AC-CHL-005 — Diff scope 라인 65 + 라인 59/61 + B12 영역으로 제한

**Given** 본 SPEC 의 변경 범위는 CHANGELOG.md 의 라인 65 (deletion) + 라인 59 + 라인 61 (edit), 그리고 manager-develop-prompt-template.md 의 B12 영역 (insertion) 만이다,
**When** M1+M2+M3 milestone 전체가 완료된 후,
**Then** `git diff` 결과가 정확히 2개 파일 (CHANGELOG.md + manager-develop-prompt-template.md) 만 포함해야 하며, sibling SPEC 디렉토리는 0건 변경이어야 한다.

**Binary verify command**:

```bash
# 본 SPEC scope 의 정확한 파일 집합만 변경됨 (CHANGELOG + template + 4 SPEC artifact)
CHANGED_FILES=$(git diff --name-only HEAD~3..HEAD | sort -u)
EXPECTED_FILES=$(printf 'CHANGELOG.md\n.claude/rules/moai/development/manager-develop-prompt-template.md\n.moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/plan.md\n.moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/spec.md\n.moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/acceptance.md\n.moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/progress.md\n' | sort -u)

# (B2 fix) CHANGED_FILES 가 EXPECTED_FILES 와 byte-identical 이어야 PASS (out-of-scope 변경 absolute 차단)
UNEXPECTED_CHANGES=$(comm -23 <(echo "$CHANGED_FILES") <(echo "$EXPECTED_FILES") | wc -l | tr -d ' ')
test "$UNEXPECTED_CHANGES" -eq 0 && \

# 5 sibling SPEC 디렉토리 0건 변경 (cross-check, B2 와 중복이지만 explicit)
SIBLING_CHANGES=$(git diff --name-only HEAD~3..HEAD .moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/ .moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/ .moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/ .moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/ .moai/specs/SPEC-V3R6-SESSION-HANDOFF-AUTO-001/ | wc -l | tr -d ' ')
test "$SIBLING_CHANGES" -eq 0
```

Exit 0 = PASS (sibling SPEC 디렉토리 0건 변경). Exit !=0 = FAIL.

## Definition of Done

본 SPEC 의 모든 milestone (M1+M2+M3) 가 완료되고 모든 AC (AC-CHL-001 ~ AC-CHL-005) 가 binary PASS 인 상태가 Definition of Done. plan-auditor 가 PASS 0.75 이상 (Tier S threshold) 을 부여하고, orchestrator 가 §D Quality Gate 5개 명령을 단일 turn parallel batch 로 실행하여 모두 exit 0 이면 본 SPEC 종료.
