# Acceptance Criteria — SPEC-DIVECC-PAPER-ARCHIVE-001

> SSOT for acceptance criteria. 각 AC는 observable/mechanical (file-existence / grep / diff / git show). 이 AC들은 **run-phase** 결과를 bind하며 plan-phase에서 충족되지 않는다.
> Notation: Given-When-Then (GEARS clause 매핑: Given → Where+While, When → When, Then → shall).

---

## Definition of Done

다음 8개 AC가 모두 PASS이고, run-commit diff가 정확히 2개 산출물(아카이브 파일 + cross-reference 라인) + spec.md frontmatter 전이만 포함하며, Go 변경 0, template mirror 미생성일 때 SPEC은 DONE이다.

---

## AC-PA-001 — 아카이브 파일 존재 + non-stub (binds REQ-PA-001)

- **GIVEN** run-phase가 `.moai/research/` 아래 아카이브 파일을 작성했고
- **WHEN** 파일 존재와 크기를 확인하면
- **THEN** 파일이 존재하고 bare bibliographic stub가 아니다 (substantive — precedent `gears-paper-validation.md` ≈ 10KB house style에 준함).
- **Verification**:
  ```bash
  test -f .moai/research/dive-into-claude-code-archive.md && echo "EXISTS" || echo "MISSING"
  wc -c .moai/research/dive-into-claude-code-archive.md   # 기대: ≥ 4000 bytes (non-stub 임계)
  ```
- **PASS 조건**: `EXISTS` AND byte count ≥ 4000 (non-stub 임계 — 서지정보만 있는 stub는 ~500B 미만; substantive 항목은 §2~§5 포함으로 임계 초과).

## AC-PA-002 — full bibliographic citation 4요소 present (binds REQ-PA-002)

- **GIVEN** 아카이브 파일
- **WHEN** 4개 서지 요소를 grep하면
- **THEN** title / authors(VILA-Lab) / arXiv ID / companion repo URL 모두 present.
- **Verification**:
  ```bash
  for t in "Dive into Claude Code" "VILA-Lab" "2604.14228" "github.com/VILA-Lab/Dive-into-Claude-Code"; do
    grep -qF "$t" .moai/research/dive-into-claude-code-archive.md || echo "MISSING: $t"
  done
  ```
- **PASS 조건**: 출력 없음 (4요소 모두 present).

## AC-PA-003 — consume된 CC-internals 내용 present (binds REQ-PA-003)

- **GIVEN** 아카이브 파일
- **WHEN** consume된 핵심 내용을 grep하면
- **THEN** 5-layer graduated-compaction taxonomy(5개 layer 이름 verbatim) + AI-agent-system design-space taxonomy + query-loop/withheld-recoverable-error framing이 포착됨.
- **Verification (5-layer arm)**:
  ```bash
  for n in "Budget Reduction" "Snip" "Microcompact" "Context Collapse" "Auto-Compact"; do
    grep -qF "$n" .moai/research/dive-into-claude-code-archive.md || echo "MISSING: $n"
  done
  ```
- **Verification (design-space + query-loop arm)**:
  ```bash
  grep -niE "design[- ]space|query[- ]?loop|withheld[- ]recoverable" .moai/research/dive-into-claude-code-archive.md
  ```
- **PASS 조건**: 5-layer for-loop 출력 없음 AND design-space/query-loop arm ≥ 1 match.

## AC-PA-004 — 4개 인용 표면 모두 열거됨 (binds REQ-PA-004)

- **GIVEN** 아카이브 파일
- **WHEN** 4개 canonical 인용 표면 경로를 grep하면
- **THEN** 4개 표면 모두 아카이브에 열거됨.
- **Verification**:
  ```bash
  for p in \
    ".claude/output-styles/moai/moai.md" \
    ".claude/rules/moai/workflow/context-window-management.md" \
    ".claude/rules/moai/workflow/runtime-recovery-doctrine.md" \
    ".claude/rules/moai/development/agent-authoring.md"; do
    grep -qF "$p" .moai/research/dive-into-claude-code-archive.md || echo "MISSING: $p"
  done
  ```
- **PASS 조건**: 출력 없음 (4개 표면 경로 모두 present).

## AC-PA-005 — cross-reference 라인이 단일 아카이브로 연결 (binds REQ-PA-005)

- **GIVEN** run-phase가 cross-reference target file(plan.md §B 제안: `runtime-recovery-doctrine.md` §5)에 라인 1개를 추가했고
- **WHEN** target file에서 아카이브 경로를 grep하면
- **THEN** target file이 단일 durable 아카이브 파일 경로를 인용함.
- **Verification**:
  ```bash
  grep -nF "dive-into-claude-code-archive.md" .claude/rules/moai/workflow/runtime-recovery-doctrine.md
  ```
- **PASS 조건**: ≥ 1 match (cross-reference 라인 present in §5 Cross-References).
- **Edge case**: run-phase가 plan.md §B와 다른 target을 선택하면, 그 target file에서 동일 grep이 ≥ 1 match이고 그 target이 dev-local(template mirror 없음)이어야 한다 (neutrality 위반 방지). target 변경은 progress.md §E.2에 기록.

## AC-PA-006 — consume-not-implement framing co-located with layer name (binds REQ-PA-006)

- **GIVEN** 아카이브 파일의 5-layer taxonomy 기록 영역
- **WHEN** consume/not-implement 문구가 layer 이름과 co-located되었는지 확인하면
- **THEN** "consumes / does not implement" 의미 문구가 `Budget Reduction` 또는 `graduated-compaction`과 co-located.
- **Verification (non-vacuous co-location anchor, N5 AC-CLN-003 패턴 차용)**:
  ```bash
  grep -niE 'consume[sd]?.{0,80}(Budget Reduction|graduated[- ]compaction)|(Budget Reduction|graduated[- ]compaction).{0,80}(consume[sd]?|does not implement|not implement)' .moai/research/dive-into-claude-code-archive.md
  ```
- **PASS 조건**: ≥ 1 match in the archive.

## AC-PA-007 — out-of-scope 파일 변경 없음; rmirror 미생성 (binds REQ-PA-007)

- **GIVEN** run-phase commit diff
- **WHEN** `git show --stat <run-commit>`을 확인하면
- **THEN** 변경 파일이 {아카이브 파일, cross-reference target file(runtime-recovery-doctrine.md), spec.md frontmatter status 전이} ⊆ 집합이고; `.moai/research/`의 template mirror가 생성되지 않음.
- **Verification**:
  ```bash
  git show --stat <run-commit>   # 파일 목록 ⊆ {archive, runtime-recovery-doctrine.md, progress.md, spec.md}
  find internal/template/templates -path '*moai/research*'   # 기대: empty (mirror 미생성)
  ```
- **PASS 조건**: diff 파일 목록이 허용 집합의 부분집합 AND `find` 출력 empty.

## AC-PA-008 — Go 변경 0; arXiv 재검증 흔적 없음 (binds REQ-PA-008)

- **GIVEN** run-phase commit diff
- **WHEN** `.go` 파일 변경과 WebFetch/WebSearch 흔적을 확인하면
- **THEN** `.go` 파일 변경 0개이고; 동작 변경 없음 (doc-only).
- **Verification**:
  ```bash
  git show --stat <run-commit> | grep -c '\.go'   # 기대: 0
  ```
- **PASS 조건**: `.go` 변경 count = 0. (arXiv 재검증은 인용이 N5에서 established이므로 run-phase에서 수행되지 않음 — REQ-PA-008. 아카이브 항목은 in-repo Read로만 일관성 확인.)

---

## Edge Cases

1. **아카이브 파일명 변경**: plan.md §C.1 제안 파일명(`dive-into-claude-code-archive.md`)을 run-phase가 다른 이름으로 작성하면, AC-PA-001~006의 모든 경로를 그 실제 파일명으로 치환하여 검증한다. 파일명 변경은 progress.md §E.2에 기록.
2. **cross-reference target 변경**: AC-PA-005 edge case 참조 — 새 target은 dev-local(template mirror 없음)이어야 하며, 변경은 progress.md §E.2에 기록.
3. **byte 임계 경계**: AC-PA-001의 4000B 임계는 precedent(≈10KB)의 약 40% — substantive 보장의 보수적 하한. 4000B 미만이면 stub로 간주하여 FAIL.
4. **non-vacuity 확인 (AC-PA-006)**: co-location anchor 패턴은 신규 파일을 대상으로 하므로 pre-edit baseline이 0 match(파일 부재)임이 자명하다 — pass는 run-phase 추가 후에만 가능.

---

## Quality Gate Criteria

- [ ] 8개 AC 모두 binary PASS (위 Verification 명령 실제 출력 첨부).
- [ ] run-commit diff ⊆ {archive, runtime-recovery-doctrine.md, progress.md, spec.md}.
- [ ] `.go` 변경 0개.
- [ ] `find internal/template/templates -path '*moai/research*'` empty (mirror 미생성).
- [ ] Conventional Commits format + `🗿 MoAI` trailer.
- [ ] subagent boundary: AskUserQuestion 호출 0 (blocker는 structured report).
