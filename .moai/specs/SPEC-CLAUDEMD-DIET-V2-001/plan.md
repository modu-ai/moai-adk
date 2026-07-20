# Plan — SPEC-CLAUDEMD-DIET-V2-001

Tier M. 2nd-round CLAUDE.md diet. Lifecycle: plan → run → sync (3-phase). Track 2 of official-docs-audit-2026-07.

---

## A. Context

본 plan은 `spec.md`의 REQ-CMD2-001..010을 실행 가능한 마일스톤으로 분해한다. 핵심 전제:

1. 6개 다이어트 후보 중 5개(§5, §8, §11, §14, §15)는 기존 rule-SSOT로의 pointer-화만으로 처리된다.
2. §16 Context Search Protocol은 SSOT가 없어 신규 룰 파일 생성 또는 in-place 압축이 필요하다(run-phase 결정 사안).
3. Template-First 원칙에 따라 모든 편집은 `internal/template/templates/CLAUDE.md` 먼저, 그런 다음 로컬 `CLAUDE.md` 동기화 — 언제나 byte-parity 유지.

사전 검증(spec.md §F.1)은 2026-07-08 this session에서 관측되었으며, 모든 SSOT 존재/부재가 확인되었다.

---

## B. Known Issues (filtered to relevant categories per Tier M)

### B.1 §16 SSOT 부재 (CRITICAL)

`find .claude/rules internal/template/templates/.claude/rules -iname '*context*search*'` → no match. §16 추출은 신규 룰 파일 생성 없이 불가능. 이 사항이 본 SPEC의 가장 큰 run-phase 리스크다.

**완화**: M1에서 §16을 가장 먼저 처리하여 신규 룰 파일 생성이 리스크가 되는지 조기 검증. Option A(in-place 압축)를 fallback으로 유지.

### B.2 1st-round KEEP 판결과의 충돌 가능성

1st-round iter-2 D1 감사는 §11 recovery/resumable, §14 operational bullets, §15 CG-Mode ASCII를 "distinctive content UNIQUE → KEEP"으로 분류했다. 본 SPEC의 일부 후보는 이 판결과 겹친다.

**완화**: 각 pointer-화 후보에 대해 AC-CMD2-009(distinctive-content grep)를 run-phase precondition으로 적용 — SSOT에 실제로 distinctive 콘텐츠가 있을 때만 pointer-화 실행.

### B.3 공식 200L 권장 미달성 잔여 위험

§1-§4 canonical을 보존하면 예상 ~300L로 200L 권장에 미달. 사용자 기대와 불일치 가능성.

**완화**: 본 plan과 spec.md가 명시적으로 200L를 SHOULD(비 MUST)로 설정. 3rd-round를 out of scope로 선언하여 잔여 위험을 문서화된 부채로 기록.

---

## C. Pre-flight baselines (run BEFORE any edit; record observed)

run-phase M1 진입 전 반드시 아래 baseline을 관측하고 기록한다:

```bash
# C.1.1 라인 수
L=$(wc -l < CLAUDE.md); T=$(wc -l < internal/template/templates/CLAUDE.md); echo "LIVE=$L TEMPLATE=$T"

# C.1.2 byte-parity
diff CLAUDE.md internal/template/templates/CLAUDE.md; echo "DIFF_EXIT=$?"

# C.1.3 [HARD] / [ZONE:*] / @embed count (D1 fix iter-2 — regex updated)
grep -c '\[HARD\]' CLAUDE.md
grep -c '\[ZONE:' CLAUDE.md
grep -cE '^@\.moai/config/sections/(user|language)\.yaml$' CLAUDE.md  # NOT `^@import` (matched 0)

# C.1.4 section line numbers (edit target anchors)
grep -n '^## [0-9]\+\.' CLAUDE.md

# C.1.5 SSOT 존재 재확인
ls .claude/rules/moai/workflow/spec-workflow.md .claude/rules/moai/workflow/dynamic-workflows.md .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/askuser-protocol.md
ls .claude/rules/moai/workflow/context-search.md 2>/dev/null || echo "NOT-FOUND (expected pre-M1)"

# C.1.6 template neutrality baseline
go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1 2>&1 | tail -5
go test ./internal/template/ -run TestInternalContentLeak -count=1 2>&1 | tail -5
```

모든 값은 이 plan의 현재 측정값(spec.md §F.1)과 일치해야 한다. 불일치 시 run-phase는 중단하고 orchestrator에게 blocker 보고.

### C.2 SSOT verification per candidate (AC-CMD2-009 precondition data — iter-2 baseline run)

각 후보별 distinctive content token과 기대 SSOT 위치. iter-2에서 실측한 baseline grep 결과 포함:

| Candidate | Distinctive token | Expected SSOT | iter-2 baseline hit | Status |
|-----------|-------------------|---------------|---------------------|--------|
| §5 Agent Chain | `Phase 1.*plan-phase.*manager-spec` | spec-workflow.md | **0** | M2 RISK — token 변경 또는 partial-KEEP 필요 |
| §8 AskUserQuestion | `AskUserQuestion is the only user-facing question channel` | askuser-protocol.md | 50 | OK |
| §11 Error Recovery | `Error Recovery Pattern` | agent-common-protocol.md | 1 | OK |
| §14 Parallel Execution | `Read-only verification batching` | agent-common-protocol.md | 1 (`Parallel Execution`) | OK |
| §15 CG Mode | `moai cg` / `CG Mode` | dynamic-workflows.md OR spec-workflow.md | **0 in both** | M2 RISK — CG Mode ASCII가 SSOT에 부재; partial-KEEP 예상 |
| §16 Context Search | `Context Search Protocol` / `session index` | (to be created: context-search.md) | n/a (file absent) | M1이 생성 |

### C.3 Pointer-화 스타일 (1st-round 일관성)

1st-round D4가 확립한 포인터 문체:

```markdown
> Canonical rule: see `.claude/rules/moai/<path>.md` § <section-heading> for <one-line-content-description>.
```

### C.4 Derived target (per-section reduction arithmetic)

| Section | Baseline | Target | Reduction | Mechanism |
|---------|----------|--------|-----------|-----------|
| §16 | 48 | 5 (Opt B) / 15 (Opt A) | −43 / −33 | New rule + pointer / in-place compress |
| §15 | 43 | 20 | −23 | CG Mode ASCII + Dynamic Workflows → pointer |
| §5 | 35 | 15 | −20 | Agent Chain → pointer |
| §11 | 19 | 8 | −11 | Error Recovery 개요 → pointer |
| §14 | 16 | 8 | −8 | Parallel Execution → pointer |
| §8 | 9 | 5 | −4 | Single pointer consolidation |
| **Total (Opt B)** | **170** | **61** | **−109** | 405 → 296L |
| **Total (Opt A)** | **170** | **71** | **−99** | 405 → 306L |

MUST ceiling 320L는 Opt A 달성 시에도 여유(306 < 320). SHOULD 210L는 어느 옵션으로도 본 in-scope 후보만으로는 미달 → 3rd-round 위임(Out of Scope).

---

## D. Key Decisions

### D.1 §16 처리 결정 — Option A vs B (run-phase 판단)

- **Option A (in-place 압축)**: §16을 48L에서 ~15L로 압축. 장점: 새 룰 파일 생성 리스크 없음. 단점: always-loaded 예산 감소 효과 제한적(−33L).
- **Option B (신규 룰 파일 + pointer)**: `.claude/rules/moai/workflow/context-search.md`를 path-scoped로 생성, §16을 5L pointer로 축소. 장점: 최대 감소(−43L) + 공식 권장에 가장 근접. 단점: template-neutrality CI guard 위반 가능성 + 1st-round가 남겨둔 "SSOT 없음" 전제의 정면 부정.

**본 plan의 기본 권장**: Option B. 단, M1 pre-flight에서 template-neutrality 위험이 확인되면 Option A로 fallback. 이 결정은 run-phase M1 진입 시 manager-develop이 내리며, 불확실한 경우 orchestrator에게 re-delegation 루트로 회신.

### D.2 §16 새 룰 파일의 `paths:` 값

Option B 채택 시, 새 `.claude/rules/moai/workflow/context-search.md`의 `paths:` 값은 run-phase가 확정. 후보:
- `paths: "**/.claude/projects/**"` — session transcript 파일에 매칭
- `paths: ".moai/specs/**"` — SPEC 작업 시에만 로드
- on-demand (no paths, Skill 경유 로드)

**본 plan의 기본 권장**: `.moai/specs/**` — context search는 주로 SPEC reference가 누락된 시나리오에서 트리거되므로 SPEC 디렉터리 작업 시에만 로드하는 것이 always-loaded 예산에 가장 효율적.

### D.3 Pointer-화 스타일 (1st-round 일관성 유지)

1st-round D4가 확립한 포인터 문체를 그대로 사용(§C.3 참조). 이 문체는 1st-round AC-CMD-009(distinctive-content grep)의 "POINTER 처리된 섹션이어도 distinctive 콘텐츠는 SSOT에 존재" 검증과 양립한다.

### D.4 §1/§3/§4 negative scope guard

이 섹션들은 본 SPEC의 HARD 제약(REQ-CMD2-006, spec.md §C.1)에 의해 편집 금지. AC-CMD2-010은 이들 섹션의 라인 수가 baseline ± 0임을 run-phase에서 확인. 단, §1/§3/§4 내의 prose trim(행위 변경 없는 wording 정리)은 본 SPEC 범위 밖이며 3rd-round에서 판단.

### D.5 200L SHOULD 목표의 scope-honesty framing (D3 fix iter-2)

공식 200L 권장은 본 SPEC의 **scope choice**로 인해 미달성된다. §1-§4 canonical content를 보존하기로 한 결정(§1 HARD list의 `moai-constitution.md` 추가 pointer-화는 이론적으로 가능하나 본 SPEC이 보호하기로 선택)으로 인해 예상 ~300L. 이것이 물리적 불가능인 것은 아니다 — 범위 설계 결정이다. 본 plan은 ≤ 320L를 MUST로, ≤ 210L를 SHOULD(achievable-if-compression-quality-high)로 설정하며, 200L 미달을 부채로 기록하지 않는다. 추가 감소는 3rd-round(별도 SPEC, out of scope)에서 §1-§4 canonical 재검토 시 판단한다.

---

## E. Self-Verification (plan-phase)

### E.1 REQ → AC 추적표 (완전성)

| REQ | AC | Verification command |
|-----|-----|----------------------|
| REQ-CMD2-001 (line count ≤ 320) | AC-CMD2-001 | `wc -l CLAUDE.md` (both trees) |
| REQ-CMD2-002 (existing-SSOT pointer-ization) | AC-CMD2-009 | `grep <distinctive-content> .claude/rules/moai/<SSOT>.md` per section |
| REQ-CMD2-003 (no-SSOT gate) | AC-CMD2-006 | `ls .claude/rules/moai/workflow/context-search.md` (Option B) OR §16 line count (Option A) |
| REQ-CMD2-004 (§16 path-scoped rule) | AC-CMD2-004 + AC-CMD2-006 | `grep '^paths:' .claude/rules/moai/workflow/context-search.md` (non-empty) |
| REQ-CMD2-005 (byte-parity) | AC-CMD2-002 | `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo $?` |
| REQ-CMD2-006 (no behavior change) | AC-CMD2-005 | `grep -c '@MX:ANCHOR'` 등 sentinel + agent catalog 표 라인 수 비교 |
| REQ-CMD2-007 ([HARD]/@embed preservation) | AC-CMD2-003 | `grep -c '\[HARD\]'; grep -cE '^@\.moai/config/sections/(user\|language)\.yaml$'` (D1 fix) |
| REQ-CMD2-008 (template neutrality) | AC-CMD2-007 | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` |
| REQ-CMD2-009 (always-loaded reduction) | AC-CMD2-004 | line count + 새 룰 파일 path-scoped 확인 |
| REQ-CMD2-010 (derived target arithmetic) | AC-CMD2-001 | plan §C.4 산술 + run-phase 측정 비교 |

빈 행 없음 — 모든 REQ가 최소 1개 AC에 매핑됨.

### E.2 plan-phase 검증 완료 항목

- [x] SPEC ID self-check (spec.md §G)
- [x] 12 canonical frontmatter field 존재 (spec.md YAML)
- [x] `### Out of Scope —` H3 서브헤딩 7개 (spec.md §D)
- [x] GEARS notation (When/While/Where + shall) — REQ 10개 모두 준수
- [x] 1st-round collision 회피 (`SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001` completed 확인)
- [x] 모든 extraction SSOT 존재/부재 spec.md §F.1에서 실측
- [x] `created` / `updated` 사용 (snake_case 아님)
- [x] `tags` comma-separated string (YAML array 아님)

---

## F. Milestones (priority-based, no time estimates)

### M1 — §16 Context Search Protocol extraction (highest-risk-first)

**목표**: §16 48L → 5L (Option B) 또는 15L (Option A). 동시에 always-loaded 예산에서 최대 감소.

**단계**:
1. Option A/B 결정(D.1 참조) — template-neutrality CI guard 사전 검증
2. Option B 선택 시:
   - `internal/template/templates/.claude/rules/moai/workflow/context-search.md` 신규 생성 (path-scoped frontmatter, §25 중립 콘텐츠)
   - §16 콘텐츠 이동 + 5L pointer 잔류
   - CI guard 2종 PASS 확인
3. Option A 선택 시:
   - §16을 15L로 in-place 압축 (When to Search 4개 bullet + Search Process 2-step 요약 + Token Budget 1줄)
4. Template ↔ live byte-parity 확인(REQ-CMD2-005)
5. `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo $?` → 0

**AC 바인딩**: AC-CMD2-001, AC-CMD2-002, AC-CMD2-004, AC-CMD2-006, AC-CMD2-007

**리스크**: template-neutrality 위반 → Option A fallback. 1st-round KEEP 판결 정면 충돌 → run-phase에서 distinctive-content grep으로 제약(이 경우엔 새 SSOT를 만들므로 충돌 아님).

### M2 — §15 Agent Teams + §5 SPEC-Based Workflow compression

**목표**: §15 43L → 20L(−23L), §5 35L → 15L(−20L). 총 −43L 감소.

**단계**:
1. §15 CG Mode ASCII diagram + Dynamic Workflows prose → `dynamic-workflows.md` + `spec-workflow.md` § Agent Teams Variant로 pointer-화
2. §5 Agent Chain detail(Phase 1-6) → `spec-workflow.md`로 pointer-화
3. 각 섹션마다 AC-CMD2-009(distinctive-content grep) 사전 실행 — pointer-화가 안전한지 확인
4. Template ↔ live byte-parity 확인

**AC 바인딩**: AC-CMD2-001, AC-CMD2-002, AC-CMD2-005, AC-CMD2-009

**리스크 (iter-2 강화 — AC-CMD2-009 baseline 실측 결과)**: AC-CMD2-009 사전 grep에서 §5 Agent Chain(`Phase 1.*plan-phase.*manager-spec` in spec-workflow.md)과 §15 CG Mode(`moai cg`/`CG Mode` in both SSOTs)가 **모두 0 hit**으로 관측되었다. 이는 M2의 −43L 감소 목표가 1st-round D1 판결(distinctive content UNIQUE → KEEP)과 동일한 패턴에 직면했음을 의미한다. run-phase M2는 다음 중 하나를 수행해야 한다:
  1. **대체 distinctive-content token 탐색** — 현재 grep 패턴이 너무 좁았을 수 있음. 예: `Phase 1.*manager-spec` (패턴 완화) 또는 `Agent Chain` (문구 변경)으로 재시도.
  2. **Partial-KEEP fallback** — §5/§15의 일부 subsection만 pointer-화하고, CG Mode ASCII처럼 SSOT에 없는 distinctive content는 in-place 유지(1st-round D1 판결 존중). 이 경우 M2 감소량은 −43L에서 −15~−25L로 축소 가능.
  3. M2 감소량 축소 시 MUST ceiling 320L 달성을 위해 M3(§11/§14/§8) 압축 강도를 높이거나, §16 Option B(신규 룰 파일)를 선택하여 −43L를 확보.

  이 리스크는 plan-phase의 blocker가 아니다 — AC-CMD2-009가 run-phase precondition으로 정확히 이 시점에 작동하도록 설계되었다.

### M3 — §11 + §14 + §8 pointer-ization

**목표**: §11 19L → 8L, §14 16L → 8L, §8 9L → 5L. 총 −23L 감소.

**단계**:
1. §11 Error Handling 복구 흐름 → `agent-common-protocol.md` § Error Recovery Pattern로 pointer-화
2. §14 Parallel Execution Safeguards → `agent-common-protocol.md` § Parallel Execution + § Pre-Spawn Sync Check로 pointer-화
3. §8 User Interaction Architecture → `askuser-protocol.md` 단일 pointer로 통합(triple-anchor를 single-pointer로)
4. 각 섹션마다 AC-CMD2-009 사전 실행
5. Template ↔ live byte-parity 확인

**AC 바인딩**: AC-CMD2-001, AC-CMD2-002, AC-CMD2-009

**리스크**: §11 Recovery/Resumable 세부는 1st-round가 KEEP으로 분류한 영역 — 본 M3에서는 Error Recovery **개요**만 pointer-화하고 세부는 유지(D.1의 방어적 접근). 마일스톤 완료 후 예상 감소량 −23L에서 −15L로 축소 가능.

### M4 — Template mirror sync + byte-parity verification + lint + final line-count audit

**목표**: 모든 편집 완료 후 최종 검증. 3-phase close 준비.

**단계**:
1. `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo $?` → 0 (REQ-CMD2-005)
2. `wc -l CLAUDE.md internal/template/templates/CLAUDE.md` → ≤ 320 (REQ-CMD2-001)
3. `[HARD]` / `[ZONE:*]` / `@<path>` embed count 보존 확인(AC-CMD2-003 — D1 fix iter-2: `^@\.moai/config/sections/(user|language)\.yaml$`)
4. agent catalog / archived-agent 리스트 / command reference 무결성 확인(AC-CMD2-005)
5. `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` PASS(AC-CMD2-007)
6. `go test ./internal/template/ -run TestSplitHarnessNamespaceNoLeak -count=1` PASS (D2 corollary — 실측 test name은 `TestSplitHarnessNamespaceNoLeak`@`split_namespace_test.go:57`; `TestInternalContentLeak`는 "no tests to run" 반환)
7. `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` clean(AC-CMD2-008 — D2 fix iter-2: file path form NOT SPEC-ID)
8. §1/§2/§3/§4 라인 수 baseline ± 1 확인(AC-CMD2-010 — D4/D5 fix iter-2: awk heading-inclusive baselines 29/36/12/50)
9. progress.md §E.2/§E.3 run-phase evidence 채움

**AC 바인딩**: AC-CMD2-001..010 전체

**리스크**: 최종 라인 카운트가 ≤ 320을 초과하면 M1-M3 중 한 섹션을 재압축. ≤ 210 SHOULD 달성 여부는 이 시점에서 확정되며, 미달성 시 부채 기록 없이 "3rd-round 위임"으로 종료.

### Milestone risk notes

- M1이 가장 높은 리스크(신규 룰 파일 + template neutrality). M1 실패 시 전체 SPEC이 run-phase blocker.
- M2-M3은 기존 SSOT가 확인되어 리스크 낮음. 단, 1st-round KEEP 판결과의 충돌은 AC-CMD2-009가 방어.
- M4는 검증 전용 마일스톤. 편집 없음. 실패 시 M1-M3으로 회신.

---

## G. Anti-Patterns

1. **@embed를 감소로 계산**: 1st-round AC-CMD-004가 확립한 원칙 위반. `@<path>` embed (Obsidian-style transclusion)는 구조 전용이며 감소에 기여하지 않는다 (D1 fix iter-2: `@import`가 아니라 `@<path>` embed).
2. **§1-§4 canonical 훼손**: 200L 달성을 위한 유혹이 있으나 본 SPEC HARD 제약 위반.
3. **새 룰 파일을 always-loaded로 생성**: path-scoped가 아니면 always-loaded 예산이 이동만 될 뿐 감소하지 않는다(REQ-CMD2-004 위반).
4. **1st-round KEEP 판결 무시**: §11/§14/§15의 distinctive 콘텐츠를 SSOT 없이 pointer-화하는 것은 1st-round D1의 "0 SSOT hits → KEEP" 판결과 정면 충돌.
5. **template neutrality 위반**: 새 룰 파일에 SPEC ID / 내부 날짜 / SHA 포함은 §25 CI guard 위반.
6. **time estimate 사용**: 마일스톤은 priority-based이며 time estimate 금지(CLAUDE.md §7 + agent-common-protocol.md § Time Estimation).
7. **behavior change 혼입**: prose/pointer refactoring 외의 변경은 본 SPEC 위반(REQ-CMD2-006).
8. **POINTER 처리 전 distinctive-content grep 생략**: AC-CMD2-009를 run-phase precondition으로 사용하지 않으면 1st-round가 보존한 distinctive 콘텐츠가 손실될 수 있다.

---

## H. Cross-References

- **spec.md**: `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` (REQ-CMD2-001..010)
- **acceptance.md**: `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/acceptance.md` (AC-CMD2-001..010)
- **progress.md**: `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/progress.md` (§E.1-E.4 skeleton)
- **1st-round plan.md**: `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/plan.md` (KEEP/CUT/POINTER 방법론 + D1-D6 iter-2 판결 — 본 SPEC의 방법론적 기반)
- **1st-round acceptance.md**: `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/acceptance.md` (AC-CMD-009 prose-duplication precondition — 본 SPEC AC-CMD2-009가 계승)
- **CLAUDE.local.md §2**: Template-First HARD rule
- **CLAUDE.local.md §25**: template internal-content isolation
- **`.claude/rules/moai/development/coding-standards.md` § File Size Limits**: CLAUDE.md 크기 휴리스틱 + 공식 200-line 목표
- **`.claude/rules/moai/core/verification-claim-integrity.md` §5**: "memory→SPEC 검증; verify-before-author" — 본 plan §C의 사전 검증이 이 원칙을 구현
