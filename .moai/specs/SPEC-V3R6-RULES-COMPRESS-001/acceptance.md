# SPEC-V3R6-RULES-COMPRESS-001 — Acceptance Criteria

> 본 문서는 `spec.md` §4 EARS Requirements 의 binary verification matrix.
> 모든 AC는 단일 shell 명령으로 실행 가능 (PASS = exit 0 or numeric threshold met).

## 1. Acceptance Criteria (8 binary ACs)

### AC-RC-001 — 종합 word count ≤ 2,300w (-1,103w 이상 감축)

**Verification Command**:

```bash
wc -w \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/verification-batch-pattern.md \
  | tail -1 | awk '{print $1}'
```

**Expected Output**: integer ≤ 2300 (목표 ≤ 1900, 절감 ≥ -1,103w from baseline 3,403w)

**Baseline 비교**:
- Baseline: 3,403w (1,927 + 712 + 764, verified 2026-05-22)
- AC threshold: 2,300w max (-1,103w / -32% minimum)
- Goal: 1,900w (-1,503w / -44% target)

**PASS 조건**: output ≤ 2300

---

### AC-RC-002 — `session-handoff.md` HARD/ZONE 조항 + canonical artifact 보존

**Verification Command (a)** — [ZONE:] [HARD] 마커 카운트 보존:

```bash
grep -c '\[ZONE:Evolvable\] \[HARD\]\|\[ZONE:Frozen\] \[HARD\]' \
  .claude/rules/moai/workflow/session-handoff.md
```

**Expected**: integer ≥ 5 (baseline 5 markers, 압축 후 동일 또는 증가만 허용)

**Verification Command (b)** — 5 Trigger 테이블 cell 보존 grep:

```bash
grep -F 'Context usage crosses model-specific threshold' \
  .claude/rules/moai/workflow/session-handoff.md && \
grep -F 'SPEC phase completion (plan/run/sync)' \
  .claude/rules/moai/workflow/session-handoff.md && \
grep -F 'User explicitly requests session end' \
  .claude/rules/moai/workflow/session-handoff.md && \
grep -F 'PR creation success' \
  .claude/rules/moai/workflow/session-handoff.md && \
grep -F 'multi-milestone task' \
  .claude/rules/moai/workflow/session-handoff.md
```

**Expected**: 5 lines output (5 trigger 모두 grep match)

**Verification Command (c)** — Canonical 6-block format 보존:

```bash
grep -F 'ultrathink.' .claude/rules/moai/workflow/session-handoff.md && \
grep -F 'applied lessons:' .claude/rules/moai/workflow/session-handoff.md && \
grep -F '전제 검증' .claude/rules/moai/workflow/session-handoff.md && \
grep -F '실행:' .claude/rules/moai/workflow/session-handoff.md && \
grep -F '머지 후:' .claude/rules/moai/workflow/session-handoff.md
```

**Expected**: 5 lines output (6-block 핵심 keyword 모두 보존; Block 3 separator는 keyword 미포함이라 4 keyword + ultrathink prefix = 5 match)

**Verification Command (d)** — Worktree Block 0 pattern 보존 (L3 opt-in 케이스):

```bash
grep -F 'Block 0' .claude/rules/moai/workflow/session-handoff.md && \
grep -F 'cd <worktree-absolute-path>' .claude/rules/moai/workflow/session-handoff.md
```

**Expected**: 2 lines output

**PASS 조건**: (a) ≥ 5 AND (b) 5 matches AND (c) 5 matches AND (d) 2 matches

---

### AC-RC-003 — `context-window-management.md` HARD 조항 + model-specific threshold table 보존

**Verification Command (a)** — [ZONE:Evolvable] [HARD] 마커 카운트 보존:

```bash
grep -c '\[ZONE:Evolvable\] \[HARD\]\|\[ZONE:Frozen\] \[HARD\]' \
  .claude/rules/moai/workflow/context-window-management.md
```

**Expected**: integer ≥ 5

**Verification Command (b)** — Context Window Targets 3-row table 보존:

```bash
grep -F 'Opus 4.7 (1M)' .claude/rules/moai/workflow/context-window-management.md && \
grep -F '1,000,000 tokens' .claude/rules/moai/workflow/context-window-management.md && \
grep -F '500,000 tokens' .claude/rules/moai/workflow/context-window-management.md && \
grep -F 'Sonnet/Opus standard (200K)' .claude/rules/moai/workflow/context-window-management.md && \
grep -F '180,000 tokens' .claude/rules/moai/workflow/context-window-management.md && \
grep -F 'Haiku' .claude/rules/moai/workflow/context-window-management.md
```

**Expected**: 6 lines output (3 model class + 3 absolute ceiling values)

**Verification Command (c)** — Threshold 백분율 모두 보존:

```bash
grep -F '50%' .claude/rules/moai/workflow/context-window-management.md && \
grep -F '90%' .claude/rules/moai/workflow/context-window-management.md && \
grep -F '95%' .claude/rules/moai/workflow/context-window-management.md
```

**Expected**: 3 lines output (50% Opus / 90% Sonnet+Haiku / 95% hard stop)

**PASS 조건**: (a) ≥ 5 AND (b) 6 matches AND (c) 3 matches

---

### AC-RC-004 — `verification-batch-pattern.md` 7-item cross-reference + 2 taxonomy table 보존

**Verification Command (a)** — Cross-reference to `agent-common-protocol.md` §Parallel Execution:

```bash
grep -F 'agent-common-protocol.md' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F '§Parallel Execution' .claude/rules/moai/workflow/verification-batch-pattern.md
```

**Expected**: 2 lines output (canonical 7-item example가 inline 아닌 link로 보존)

**Verification Command (b)** — Verification Class Taxonomy 8-row 보존 (key rows):

```bash
grep -F 'Test execution' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F 'Coverage measurement' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F 'Grep / find / sentinel scan' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F 'CLI smoke' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F 'Lint' .claude/rules/moai/workflow/verification-batch-pattern.md
```

**Expected**: 5 lines output (8-row taxonomy 중 핵심 5 class)

**Verification Command (c)** — Inline 7-item canonical example 부재 (cross-ref 외부화 확인, semantic detector):

```bash
# 7-item canonical example의 numbered headers (# 1. ~ # 7.) inline 여부 검출
# baseline: agent-common-protocol.md §Parallel Execution 가 canonical 소유자, verification-batch-pattern.md 는 cross-ref 만 보유 의무
# `# 7. Lint baseline` 또는 `# 7. Lint` 같은 7번째 마지막 헤더가 inline 되어 있으면 full duplication 신호
grep -cE '^# 7\. ' .claude/rules/moai/workflow/verification-batch-pattern.md
```

**Expected**: integer == 0 (7-item full inline 부재 = canonical reference 외부화 성공). 단순 `go test ./...` 단어 카운트는 §3.2.3 preservation 의무인 Default Grouping 5-row table (Group A row 가 `go test ./...` 포함) 와 직접 충돌하므로 사용 안 함.

**PASS 조건**: (a) 2 matches AND (b) 5 matches AND (c) == 0

**Note (B2 mitigation)**: plan-auditor iter 1 BLOCKING — 이전 `grep -c 'go test ./\.\.\.' ≤ 1` 은 baseline 5 matches (lines 53/68/82/93/102) 이고 §3.2.3 preservation 으로 Group A row (`| A. Functional | \`go test ./...\`, coverage | 30-120 s |`) 는 verbatim 보존 의무 → AC 와 preservation 의무 직접 충돌. semantic 7-item detector (numbered header `^# 7\. `) 로 교체.

---

### AC-RC-005 — Cross-reference link 무결성 (0 dangling, placeholder modulo)

**Verification Command**:

```bash
# 3개 압축된 파일 본문에서 참조하는 `.claude/...` 또는 `.moai/...` 경로 추출
# 단, `<SPEC-ID>` / `<PR-number>` 같은 angle-bracket placeholder 는 false-positive 제외
for path in $(grep -hoE '\.claude/[^ )<]*\.md|\.moai/[^ )<]*\.md' \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/verification-batch-pattern.md \
  | grep -v '<[^>]*>' \
  | sort -u); do
  if [ ! -f "$path" ]; then echo "DANGLING: $path"; fi
done
```

**Expected**: empty output (zero `DANGLING:` lines)

**PASS 조건**: 출력 0 line

**Note (B1 mitigation)**: plan-auditor iter 1 BLOCKING — baseline 에서 placeholder paths (`.moai/specs/<SPEC-ID>/progress.md` 등) 가 illustrative template 으로 paste-ready resume example 에 포함되어 있어 regex `\.claude/[^ )]*\.md|\.moai/[^ )]*\.md` 가 false-positive 추출. 정정: (1) regex character class 에 `<` 추가 (`[^ )<]`) — angle-bracket 시작 시 stop, (2) `grep -v '<[^>]*>'` filter 로 angle-bracket placeholder 라인 추가 제외. 실제 file path 만 검증 대상.

---

### AC-RC-006 — `paths:` frontmatter 추가/수정 부재

**Verification Command**:

```bash
for f in \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/verification-batch-pattern.md; do
  head -10 "$f" | grep -c '^paths:' || true
done | awk '{sum+=$1} END {print sum}'
```

**Expected**: integer == 0 (3 파일 모두 `paths:` frontmatter 부재 = always-loaded 유지)

**PASS 조건**: output == 0

---

### AC-RC-007 — Footer / HISTORY / version 마커 보존

**Verification Command**:

```bash
# session-handoff.md footer
grep -F 'HARD operational rule, applies to all multi-phase MoAI workflows' \
  .claude/rules/moai/workflow/session-handoff.md && \
# context-window-management.md footer
grep -F 'HARD operational rule, applies to all sessions' \
  .claude/rules/moai/workflow/context-window-management.md && \
# verification-batch-pattern.md footer (Version + Classification + Origin)
grep -F 'Version: 1.0.0' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F 'Classification:' .claude/rules/moai/workflow/verification-batch-pattern.md && \
grep -F 'Origin:' .claude/rules/moai/workflow/verification-batch-pattern.md
```

**Expected**: 5 lines output (3 file footer 모두 보존)

**PASS 조건**: 5 matches

---

### AC-RC-008 — 개별 파일 word count threshold 준수

**Verification Command**:

```bash
A=$(wc -w < .claude/rules/moai/workflow/session-handoff.md)
B=$(wc -w < .claude/rules/moai/workflow/context-window-management.md)
C=$(wc -w < .claude/rules/moai/workflow/verification-batch-pattern.md)
echo "session-handoff=$A context-window=$B verification-batch=$C"
test $A -le 1200 && test $B -le 600 && test $C -le 500 && echo "PASS"
```

**Expected**:
- `session-handoff.md` ≤ 1,200w (목표 1,000w + 200w 여유)
- `context-window-management.md` ≤ 600w (목표 500w + 100w 여유)
- `verification-batch-pattern.md` ≤ 500w (목표 400w + 100w 여유)
- 마지막 줄에 `PASS` 출력

**PASS 조건**: 마지막 줄 `PASS`

---

## 2. REQ ↔ AC Traceability Matrix (100% Coverage)

| REQ | AC | Verification Method | Status |
|-----|----|--------------------:|-------:|
| REQ-RC-001 (종합 -1,103w 이상 감축, ≤ 2,300w) | AC-RC-001, AC-RC-008 | `wc -w` 종합 + 개별 | binary |
| REQ-RC-002 (`session-handoff.md` HARD + canonical 보존) | AC-RC-002 (a/b/c/d), AC-RC-007 | grep HARD count + 5 Trigger + 6-block + Block 0 + footer | binary |
| REQ-RC-003 (`context-window-management.md` HARD + threshold table 보존) | AC-RC-003 (a/b/c), AC-RC-007 | grep HARD count + 3-row table + 50%/90%/95% + footer | binary |
| REQ-RC-004 (`verification-batch-pattern.md` 7-item cross-ref + 2 taxonomy 보존) | AC-RC-004 (a/b/c), AC-RC-007 | grep cross-ref + class taxonomy + inline absence + footer | binary |
| REQ-RC-005 (cross-reference 무결성) | AC-RC-005 | grep + file existence loop | binary |
| REQ-RC-006 (`paths:` frontmatter 부재 유지) | AC-RC-006 | head + grep `^paths:` | binary |
| REQ-RC-007 (HISTORY / version footer 보존) | AC-RC-007 | grep footer 5 marker | binary |

**Coverage**: 7 REQs / 8 ACs → **100% (every REQ has ≥1 mapped AC, every AC traces to ≥1 REQ)**

---

## 3. Edge Cases

### EC-RC-001 — `wc -w` 측정 시 frontmatter 포함 여부

**시나리오**: `wc -w` 는 YAML frontmatter (`---` 사이 영역) 포함하여 count. Baseline 3,403w 도 동일 방식으로 측정됨. 압축 후에도 frontmatter 변경 부재이므로 일관성 유지.

**Mitigation**: AC-RC-001 / AC-RC-008 모두 파일 전체 `wc -w` (frontmatter 포함) 기준. Baseline 과 동일 측정으로 비교. Frontmatter 자체는 본 SPEC scope 외 (변경 부재).

**Acceptance**: AC 계산은 frontmatter 포함 word count 기준. 본문만 측정해야 한다는 별도 규정 없음.

---

### EC-RC-002 — `verification-batch-pattern.md` 7-item 외부화 시 illustrative reference 허용 경계

**시나리오**: AC-RC-004 (c) 가 `grep -c 'go test ./\.\.\.'` ≤ 1 을 요구. 압축 후 본문에서 7-item canonical example 의 7 commands 가 모두 inline 으로 나열되면 SSOT 분리 risk. 그러나 1-2 commands 정도는 illustrative purpose (e.g., "예: `go test ./...`") 로 허용 — 이 경계를 ≤ 1 로 정량화.

**Mitigation**: 압축 시 7-item 전체 inline 금지, 1 example 정도만 illustrative (e.g., "such as `go test ./...`") 로 허용. `agent-common-protocol.md` §Parallel Execution 인용 link 가 canonical SSOT.

**Acceptance**: AC-RC-004 (c) PASS 시 7-item 분리 SSOT 유지 확인.

---

### EC-RC-003 — HARD 조항 verbatim 검증의 false positive

**시나리오**: `grep -c '\[ZONE:Evolvable\] \[HARD\]\|\[ZONE:Frozen\] \[HARD\]'` 카운트가 baseline 과 같거나 크면 PASS — 그러나 marker 텍스트만 보존되고 조항 본문이 paraphrase 되면 의미 손상 발생 가능. Pure grep count로는 paraphrasing 탐지 불가.

**Mitigation**:
1. AC-RC-002/003/004 의 sub-verification (b/c/d) 가 marker 인접 핵심 keyword (5 Trigger 항목 / threshold 백분율 / class taxonomy 항목명) 까지 verbatim grep 으로 검증.
2. 본 SPEC §3.2 가 보존 의무 canonical artifact 를 enumerate — manager-develop 는 압축 시 §3.2 list 와 baseline diff 로 self-audit.
3. 추가 안전장치: M4 (session-handoff.md 압축) 전후 `diff -u baseline current | grep -E '^[+-].*\[HARD\]'` 로 HARD line 변경 사항 수동 검토 의무 (manager-develop self-verification).

**Acceptance**: AC suite PASS + manager-develop self-report 의 HARD diff 수동 검토 결과 보고 의무.

---

## 4. Definition of Done

본 SPEC 의 run-phase 완료 조건:

- [ ] 8 ACs (AC-RC-001 .. AC-RC-008) 모두 PASS
- [ ] 100% REQ ↔ AC traceability 검증 완료
- [ ] 3 Edge Cases 모두 mitigation 적용 확인
- [ ] 3 files atomic edit 또는 분리 commit (manager-develop 판단)
- [ ] `wc -w` 합계 ≤ 2,300w (-1,103w 이상 감축, 목표 -1,503w / -44%)
- [ ] HARD/[ZONE:] 조항 카운트 baseline 동일 또는 증가 (10건 기준)
- [ ] Canonical artifact (5 Trigger / 6-block / threshold table / class taxonomy / 7-item cross-ref) verbatim 보존
- [ ] Cross-reference 무결성 (0 dangling)
- [ ] spec.md frontmatter `status` `draft` → `implemented` 갱신 + `version` `0.1.0` → `0.2.0`
- [ ] `progress.md` 생성 (run-phase 완료 보고 — manager-develop 책임)

---

## 5. Quality Gate Criteria

본 SPEC의 sync-phase 진입 조건:

- All 8 ACs PASS (자동 verification command 실행 결과 모두 만족)
- Cross-platform 검증 부재 (rule files 압축이므로 OS-independent)
- C-HRA-008 부재 (rule 파일이며 hook/harness code 손대지 않음)
- Subagent boundary 부재 (압축 대상이 .md only)
- 통합 test 부재 (압축은 prose-only 변경)
- 외부 SPEC 의존성 없음 (Wave 1 Lane A 다른 SPEC 과 병렬 독립)

---

Version: 0.1.0
Status: draft
Lifecycle: spec-anchored
Linked spec: `.moai/specs/SPEC-V3R6-RULES-COMPRESS-001/spec.md`
