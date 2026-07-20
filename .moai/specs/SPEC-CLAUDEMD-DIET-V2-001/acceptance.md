# Acceptance — SPEC-CLAUDEMD-DIET-V2-001

Tier M. 2nd-round CLAUDE.md diet. 본 문서는 spec.md REQ-CMD2-001..010에 대한 AC 매트릭스 + severity + traceability + Definition of Done를 정의한다.

---

## D. Acceptance Criteria Matrix

### AC-CMD2-001 — 라인 수 목표 달성 (MUST)

**검증 명령**:
```bash
L=$(wc -l < CLAUDE.md); T=$(wc -l < internal/template/templates/CLAUDE.md)
echo "LIVE=$L TEMPLATE=$T"
# MUST: $L -le 320 AND $T -le 320 AND $L -eq $T
# SHOULD: $L -le 210 AND $T -le 210
```

**통과 조건**:
- **MUST**: `LIVE ≤ 320 AND TEMPLATE ≤ 320 AND LIVE == TEMPLATE`
- **SHOULD (achievable)**: `LIVE ≤ 210 AND TEMPLATE ≤ 210` — 공식 Claude Code best-practice alignment
- baseline: 405 / 405 / 405 (2026-07-08 측정, spec.md §F.1)

**실패 시**: M1-M3 중 한 섹션의 압축이 불충분했음. 해당 섹션의 재압축 또는 3rd-round 위임 판단.

**근거 REQ**: REQ-CMD2-001, REQ-CMD2-010

---

### AC-CMD2-002 — Template ↔ live byte-parity (MUST)

**검증 명령**:
```bash
diff CLAUDE.md internal/template/templates/CLAUDE.md; echo "DIFF_EXIT=$?"
# MUST: DIFF_EXIT == 0 (no output)
```

**통과 조건**: `DIFF_EXIT == 0` — CLAUDE.md와 template 소스가 byte-identical.

**실패 시**: Template-First 원칙 위반. 두 트리 중 어느 하나가 동기화 누락. run-phase M4에서 반드시 동기화 후 close.

**근거 REQ**: REQ-CMD2-005

---

### AC-CMD2-003 — [HARD] / [ZONE:*] / @embed 보존 (MUST)

**검증 명령** (D1 fix iter-2 — OLD regex `^@import` matched 0 lines, bug; actual syntax is Obsidian-style `@<path>` transclusion):
```bash
grep -c '\[HARD\]' CLAUDE.md           # baseline: 14 (실측 2026-07-08)
grep -c '\[HARD\]' internal/template/templates/CLAUDE.md  # baseline: 14
grep -c '\[ZONE:' CLAUDE.md            # baseline: 14 (실측 2026-07-08)
grep -cE '^@\.moai/config/sections/(user|language)\.yaml$' CLAUDE.md           # baseline: 2 (실측 2026-07-08)
grep -cE '^@\.moai/config/sections/(user|language)\.yaml$' internal/template/templates/CLAUDE.md  # baseline: 2
```

**통과 조건**:
- `[HARD]` count ≥ 14 (1st-round ceiling) AND both trees equal
- `[ZONE:*]` count == 14 (both trees equal)
- `@<path>` embed count == 2 (both trees equal) — `@.moai/config/sections/user.yaml` + `@.moai/config/sections/language.yaml` (Obsidian-style transclusion, NOT markdown `@import`; iter-1 plan-audit D1 caught the regex bug)

**실패 시**: behavior change 발생. 다이어트가 HARD rule / ZONE marker / @embed를 손상함. 즉시 revert.

**근거 REQ**: REQ-CMD2-006, REQ-CMD2-007

---

### AC-CMD2-004 — Always-loaded 예산 감소 (MUST)

**검증 명령**:
```bash
# 1. CLAUDE.md 라인 수 감소 확인 (AC-CMD2-001에 의존)
# 2. 새 룰 파일(Option B)이 path-scoped인지 확인
ls .claude/rules/moai/workflow/context-search.md 2>/dev/null && \
  grep '^paths:' .claude/rules/moai/workflow/context-search.md && \
  echo "PATH-SCOPED: yes" || echo "PATH-SCOPED: n/a (Option A)"
```

**통과 조건**:
- CLAUDE.md 라인 수 감소(AC-CMD2-001)
- Option B 선택 시: `context-search.md`가 `^paths:` 라인을 포함(non-empty) — always-loaded가 아님을 보장
- Option A 선택 시: 새 룰 파일 없음, 이 검증 항목 n/a

**실패 시**: always-loaded 예산이 이동만 될 뿐 감소하지 않음. 새 룰 파일에 path frontmatter 추가 또는 Option A fallback.

**근거 REQ**: REQ-CMD2-004, REQ-CMD2-009

---

### AC-CMD2-005 — 행위 불변 (MUST)

**검증 명령**:
```bash
# 1. 8-retained-agent catalog 라인 수 보존
grep -c '^| `manager-' CLAUDE.md       # baseline: 4 (manager-spec/develop/docs/git — 실측 2026-07-08; iter-1 plan-audit D4-context corrected from erroneous 5)
grep -c '^| `plan-auditor`' CLAUDE.md  # baseline: 1 (실측 2026-07-08)
grep -c '^| `sync-auditor`' CLAUDE.md  # baseline: 1 (실측 2026-07-08)
grep -c '^| `builder-harness`' CLAUDE.md  # baseline: 1 (실측 2026-07-08)
grep -c '^| `Explore`' CLAUDE.md       # baseline: 1 (실측 2026-07-08)

# 2. archived-agent 리스트 보존
grep -c 'archived.*MUST NOT be spawned' CLAUDE.md  # baseline: 1 (실측 2026-07-08)

# 3. command reference (§3) 라인 수 보존
awk '/^## 3\. Command Reference/,/^## 4\./' CLAUDE.md | wc -l  # baseline: 12 (awk heading-inclusive, 실측 2026-07-08)

# 4. Selection Decision Tree 라인 수 보존 (§4)
grep -c 'Use the.*subagent' CLAUDE.md  # baseline: 11 (decision tree items, 실측 2026-07-08; iter-1 plan-audit D4-context corrected from erroneous 9)
```

**통과 조건**: 모든 카운트가 baseline과 동일.

**실패 시**: agent 카탈로그 / archived 리스트 / command set이 변경됨. 즉시 revert.

**근거 REQ**: REQ-CMD2-006

---

### AC-CMD2-006 — §16 추출 완전성 (MUST)

**검증 명령**:
```bash
# Option B (신규 룰 파일)
ls .claude/rules/moai/workflow/context-search.md && echo "OPTION-B: file exists"
diff .claude/rules/moai/workflow/context-search.md \
     internal/template/templates/.claude/rules/moai/workflow/context-search.md && \
  echo "OPTION-B: byte-parity"

# Option A (in-place 압축) — 파일 없음 확인
ls .claude/rules/moai/workflow/context-search.md 2>/dev/null && echo "UNEXPECTED" || echo "OPTION-A: no new file"

# §16 라인 수 (both options)
awk '/^## 16\. Context Search/,/^## 17\./' CLAUDE.md | wc -l
# Option B target: ~7 (5L pointer + 2L section markers)
# Option A target: ~17 (15L compressed + 2L section markers)
```

**통과 조건**:
- Option B: 새 룰 파일 존재 + template과 byte-parity + §16 ≤ 8L
- Option A: 새 룰 파일 부재 + §16 ≤ 17L
- 어느 옵션이든 "When to Search" 4개 trigger / "When NOT to Search" 4개 condition / Search Process 핵심 단계가 보존됨(새 룰 파일 또는 압축된 §16 어느 한쪽에)

**실패 시**: §16 콘텐츠가 손실되었거나, Option A/B 불일치. 재작업.

**근거 REQ**: REQ-CMD2-003, REQ-CMD2-004

---

### AC-CMD2-007 — Template neutrality (MUST)

**검증 명령**:
```bash
cd /Users/goos/MoAI/moai-adk-go
go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1 2>&1 | tail -10
go test ./internal/template/ -run TestInternalContentLeak -count=1 2>&1 | tail -10
```

**통과 조건**: 두 테스트 모두 PASS.

**실패 시**: 새 context-search.md 룰 파일(또는 기존 CLAUDE.md 편집)이 SPEC ID / 내부 날짜 / SHA / 내부 참조를 포함. 해당 콘텐츠 제거 후 재테스트.

**근거 REQ**: REQ-CMD2-008

---

### AC-CMD2-008 — GEARS lint clean (MUST)

**검증 명령** (D2 fix iter-2 — OLD form `moai spec lint SPEC-CLAUDEMD-DIET-V2-001` produced ParseFailure; subcommand expects a FILE PATH):
```bash
moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md
# 실측 2026-07-08: `✓ No findings — all SPEC documents are valid`
```

**통과 조건**: 본 SPEC의 frontmatter 12-field schema + GEARS notation + OwnershipTransition + status 일관성 전부 PASS.

**실패 시**: lint finding 제거 후 재실행. `lint.skip`은 documented debt에만 허용.

**근거 REQ**: (frontmatter schema + GEARS notation canonical requirement)

---

### AC-CMD2-009 — Distinctive-content grep (POINTER precondition, MUST)

**검증 명령** (각 pointer-화 후보에 대해 run-phase 사전 실행):

```bash
# §5 SPEC-Based Workflow — Agent Chain이 spec-workflow.md에 있는지
grep -c 'Phase 1.*plan-phase.*manager-spec' .claude/rules/moai/workflow/spec-workflow.md
# 기대: ≥ 1 (Agent Chain 단계가 SSOT에 존재)

# §11 Error Handling — recovery flow가 agent-common-protocol.md에 있는지
grep -c 'Error Recovery Pattern' .claude/rules/moai/core/agent-common-protocol.md
# 기대: ≥ 1

# §14 Parallel Execution — verification batch가 agent-common-protocol.md에 있는지
grep -c 'Parallel Execution' .claude/rules/moai/core/agent-common-protocol.md
# 기대: ≥ 1

# §15 Agent Teams — CG Mode가 dynamic-workflows.md 또는 spec-workflow.md에 있는지
grep -c 'CG Mode\|moai cg' .claude/rules/moai/workflow/spec-workflow.md .claude/rules/moai/workflow/dynamic-workflows.md
# 기대: ≥ 1

# §8 User Interaction — AskUserQuestion이 askuser-protocol.md에 있는지
grep -c 'AskUserQuestion' .claude/rules/moai/core/askuser-protocol.md
# 기대: ≥ 1
```

**통과 조건**: 모든 grep이 ≥ 1 hit. 즉, distinctive content가 실제 SSOT에 존재함을 기계적으로 확인.

**실패 시**: pointer-화 안전하지 않음 — distinctive content가 SSOT에 없음. 해당 섹션 KEEP(1st-round D1 판결과 동일). pointer-화 시도 차단.

**run-phase precondition**: 이 검증은 pointer-화 편집 **이전**에 실행. 사후 검증이 아님.

**근거 REQ**: REQ-CMD2-002, REQ-CMD2-003

---

### AC-CMD2-010 — §1/§3/§4 negative scope guard (MUST)

**검증 명령**:
```bash
# §1 Core Identity 라인 수 보존
awk '/^## 1\. Core Identity/,/^## 2\./' CLAUDE.md | wc -l
# baseline: 29 (awk heading-inclusive, 실측 2026-07-08; content body excluding §2 heading = 28L)

# §3 Command Reference 라인 수 보존
awk '/^## 3\. Command Reference/,/^## 4\./' CLAUDE.md | wc -l
# baseline: 12 (awk heading-inclusive; content body = 11L)

# §4 Agent Catalog 라인 수 보존 (선택 트리 포함)
awk '/^## 4\. Agent Catalog/,/^## 5\./' CLAUDE.md | wc -l
# baseline: 50 (awk heading-inclusive; content body = 49L)
```

**통과 조건**: 각 섹션의 awk-measured 라인 수가 baseline ± 1 이내(줄바꿈 정리 허용치). 내용 토큰(8-retained-agent 표, archived-agent 리스트, Selection Decision Tree)이 보존됨. (iter-1 plan-audit D4 지적 보정: awk는 heading-inclusive이므로 baseline 29/12/50으로 명시)

**실패 시**: 본 SPEC HARD 제약(REQ-CMD2-006, spec.md §C.1, §D Out of Scope) 위반. 즉시 revert.

**근거 REQ**: REQ-CMD2-006

---

## D.1 Severity / blocking classification

| AC | Severity | Blocking? | Notes |
|----|----------|-----------|-------|
| AC-CMD2-001 (line count) | MUST | YES (MUST portion ≤ 320) | SHOULD ≤ 210은 blocking 아님 — 부채 기록 없이 "3rd-round 위임" |
| AC-CMD2-002 (byte-parity) | MUST | YES | Template-First HARD 위반 시 즉시 revert |
| AC-CMD2-003 ([HARD]/@embed) | MUST | YES | behavior change이므로 즉시 revert (D1 fix iter-2: @<path> embed regex) |
| AC-CMD2-004 (always-loaded) | MUST | YES | path-scoped 위반 시 새 룰 파일이 always-loaded가 됨 |
| AC-CMD2-005 (no behavior change) | MUST | YES | agent catalog / archived / command set 변경 시 즉시 revert |
| AC-CMD2-006 (§16 extraction) | MUST | YES | Option A/B 불문 §16 처리 완전성 필수 |
| AC-CMD2-007 (template neutrality) | MUST | YES | CI guard 위반 시 배포 브레이크 |
| AC-CMD2-008 (GEARS lint) | MUST | YES | SPEC 자체 무결성 |
| AC-CMD2-009 (distinctive-content grep) | MUST | YES (run-phase precondition) | 사후가 아닌 사전 검증 — pointer-화 이전 실행 |
| AC-CMD2-010 (§1/§3/§4 guard) | MUST | YES | canonical content 훼손 시 즉시 revert |

**총 10 AC, 전부 MUST-blocking. SHOULD는 AC-CMD2-001의 ≤ 210 부분만.**

---

## D.2 Traceability (REQ → AC)

| REQ | 1차 AC | 2차 AC |
|-----|--------|--------|
| REQ-CMD2-001 (line count ≤ 320) | AC-CMD2-001 | — |
| REQ-CMD2-002 (existing-SSOT pointer) | AC-CMD2-009 | — |
| REQ-CMD2-003 (no-SSOT gate) | AC-CMD2-006 | AC-CMD2-009 |
| REQ-CMD2-004 (§16 path-scoped rule) | AC-CMD2-004 | AC-CMD2-006 |
| REQ-CMD2-005 (byte-parity) | AC-CMD2-002 | — |
| REQ-CMD2-006 (no behavior change) | AC-CMD2-005 | AC-CMD2-010 |
| REQ-CMD2-007 ([HARD]/@embed preservation) | AC-CMD2-003 | — (D1 fix iter-2: @embed NOT @import) |
| REQ-CMD2-008 (template neutrality) | AC-CMD2-007 | — |
| REQ-CMD2-009 (always-loaded reduction) | AC-CMD2-004 | AC-CMD2-001 |
| REQ-CMD2-010 (derived target) | AC-CMD2-001 | — |

모든 REQ가 ≥ 1 AC에 매핑됨. 빈 행 없음.

---

## D.3 Definition of Done

본 SPEC이 "completed"로 전이되기 위해 필요한 조건:

1. **AC 10개 전부 PASS** (MUST 10 + SHOULD 1 — SHOULD는 blocking 아님)
2. **3-phase close** (plan → run → sync) — manager-develop가 run-phase를, manager-docs가 sync-phase를 수행
3. **progress.md §E.2 / §E.3 / §E.4** 채워짐 — run-phase evidence + audit-ready signal + sync-phase audit-ready signal
4. **commit history**:
   - plan-phase: `feat(SPEC-CLAUDEMD-DIET-V2-001): plan-phase artifacts (Tier M, 4 artifacts)` (manager-spec)
   - run-phase M1: `fix(SPEC-CLAUDEMD-DIET-V2-001): M1 §16 extraction...` (manager-develop, `draft → in-progress` 전이)
   - run-phase M2-M4: feat/fix per milestone (manager-develop)
   - sync-phase: `docs(SPEC-CLAUDEMD-DIET-V2-001): sync-phase artifacts` (manager-docs, `in-progress → implemented → completed` 전이를 single sync commit에 통합)
5. **CLAUDE.md 및 template**이 post-diet 상태로 main 브랜치에 존재
6. **1st-round SPEC 훼손 없음** — `SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001`의 AC가 본 SPEC에 의해 회귀되지 않음

---

## D.4 Indirect verification note

일부 AC는 기계적 grep만으로 완전 검증이 어려움:

- **AC-CMD2-005 (no behavior change)**: 카탈로그 라인 수 보존은 기계적이나, agent routing **시맨틱** 보존은 run-phase의 manual spot-check로 보완 필요(예: "Use the manager-develop subagent..." 패턴 무결성).
- **AC-CMD2-009 (distinctive-content grep)**: grep hit 수가 보존을 의미하지는 않음 — hit이 맥락적으로 동등한지 run-phase가 판단. 1st-round D1의 "prose-duplication bar"를 계승.
- **AC-CMD2-001 SHOULD (≤ 210)**: 달성 여부가 압축 품질에 의존하므로 기계적으로 보장 불가. 미달성 시 부채가 아닌 "3rd-round 위임"으로 종료.

이 한계들은 1st-round acceptance.md §D.4와 동일한 구조를 계승하며, verification-claim-integrity.md §1.1 surface 3(defect-claim hazard)의 binding을 받는다.

---

## D.5 Forward-looking checks (post-close 품질 게이트)

본 SPEC close 후 후속 검증 권장(별도 SPEC):

1. **always-loaded 토큰 실측**: 본 SPEC은 라인 수 감소를 목표로 하지만, always-loaded **토큰** 감소는 별도 측정이 필요(statusline 또는 `/cost`로 모니터링).
2. **3rd-round 다이어트 후보 탐색**: ≤ 210 SHOULD 미달성 시, §1-§4 canonical 훼손 없이 추가 감소가 가능한지 별도 SPEC에서 조사.
3. **agent 동작 회귀 모니터링**: 다이어트 후 1-2주간 agent 호출 패턴 이상 유지 관찰(예: hardcoded CLAUDE.md 라인 참조가 있는지 grep).

이 항목들은 본 SPEC의 AC가 아니며, 후속 백로그로만 기록된다.
