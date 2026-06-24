# Implementation Plan — SPEC-DIVECC-PAPER-ARCHIVE-001

> Tier S. doc-only, dev-local(`.moai/research/`), no Go, no template mirror, no behavior change.
> 산출물 2가지: 아카이브 파일 1개 + cross-reference 라인 1개.

## §A. Context

- **작업 위치**: project root `/Users/goos/MoAI/moai-adk-go` (dev-local; `.moai/research/`는 template-distributed 아님).
- **Epic**: Dive-into-CC 후보 N7 (LOW). siblings N1/N2/N3/N5 closed.
- **SPEC 산출물 경로**: `.moai/specs/SPEC-DIVECC-PAPER-ARCHIVE-001/{spec,plan,acceptance,progress}.md`.
- **전제**: trivial (archival). 논문 인용은 N5에서 VERIFIED-by-citation established. 본 SPEC은 in-repo 인용 표면 일관성만 확인 (재검증 금지).
- **PRESERVE (절대 변경 금지)**: 4개 인용 표면 본문(moai.md / context-window-management.md / runtime-recovery-doctrine.md / agent-authoring.md) — 단, §C.2의 cross-reference 라인 1개 추가만 예외; precedent `gears-paper-validation.md`; 그 외 모든 파일.
- **EXTEND**: `.moai/research/` (신규 아카이브 파일 1개).

## §B. 제안: cross-reference target (REQ-PA-005)

[제안] cross-reference 라인의 target은 **`.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §5 Cross-References (lineage traceability)** 이다.

### 제안 근거 (rationale)

대안 4개를 평가했다:

| 후보 target | 평가 | 채택? |
|-------------|------|-------|
| **runtime-recovery-doctrine.md §5 Cross-References** | 이미 `wquguru/harness-books` book1 + VILA-Lab 논문을 lineage traceability cross-reference로 나열함 (line 외부). N5가 이 §1에 VILA-Lab을 convergent second source로 추가했고, §5에 lineage cross-ref가 이미 있음. dev-local(no mirror, no neutrality 제약) → SPEC-DIVECC 내부 ID/경로 인용 자유. 4개 인용 표면 중 논문 출처(book1 + VILA-Lab 둘 다)를 가장 명시적으로 다루는 표면. | **채택 (권장)** |
| context-window-management.md "Graduated-Compaction Layers" 섹션 | template-distributed → neutrality 제약(C1-C8). `.moai/research/` 내부 경로 인용이 forbidden internal-content class(internal-only path)에 걸릴 위험. 부적합. | 미채택 |
| moai.md §4 (delegation) | template-distributed(추정). neutrality 제약. + delegation 표면에 archive 포인터를 다는 것은 주제 mismatch. | 미채택 |
| agent-authoring.md (extension ladder) | template-distributed. neutrality 제약. extension 표면에 archive 포인터는 주제 mismatch. | 미채택 |

**채택 사유 3가지**:

1. **dev-local 안전성**: `runtime-recovery-doctrine.md`는 N5 plan-phase 관측(§B.1)에서 확인된 대로 template mirror가 **없는** local-only 파일(internal SPEC ID를 이미 carry). 따라서 cross-reference 라인이 dev-local 아카이브 경로(`.moai/research/<archive>.md`)와 내부 SPEC-DIVECC ID를 인용해도 neutrality CI guard에 걸리지 않는다. template-distributed 표면(나머지 3개)은 이 인용을 담을 수 없다.
2. **주제 정합성**: §5는 이미 "lineage traceability"를 위한 cross-reference 집합이며 book1 + VILA-Lab 논문을 둘 다 나열한다. 단일 durable 아카이브로의 포인터는 이 lineage 집합에 자연스럽게 들어간다.
3. **인용 응집(consolidation) 효과**: §5에 추가된 한 라인이 "분산된 4개 인용 → 단일 아카이브" 매핑을 명시하므로, 독자가 어느 인용 표면에서든 §5를 거쳐 durable 아카이브에 도달할 수 있다.

### cross-reference 라인 형태 (run-phase가 작성할 내용 — 예시)

`runtime-recovery-doctrine.md` §5 Cross-References 목록에 한 라인 추가:

```
- `.moai/research/<archive-name>.md` — VILA-Lab "Dive into Claude Code" (arXiv:2604.14228) 논문의 durable 아카이브. 본 논문을 consume하는 in-repo 인용 표면 4곳(moai.md / context-window-management.md / runtime-recovery-doctrine.md §1 / agent-authoring.md)을 단일 항목으로 통합 (SPEC-DIVECC-PAPER-ARCHIVE-001 / Epic Dive-into-CC N7).
```

> run-phase가 `<archive-name>`을 §C.1에서 확정한 실제 파일명으로 치환한다.

## §C. Technical Approach

### C.1 아카이브 파일 (REQ-PA-001~004, 006)

- **경로**: `.moai/research/dive-into-claude-code-archive.md` (제안 파일명; precedent 명명 패턴 `<topic>-<kind>.md`와 정합 — 예: `gears-paper-validation.md`).
- **구조** (precedent `gears-paper-validation.md`를 모델로, house style ≈ 10KB substantive):
  - frontmatter (name / description / created / updated / author / related_spec: SPEC-DIVECC-PAPER-ARCHIVE-001 / related_epic).
  - §1 Full Bibliographic Citation — title / authors(VILA-Lab) / arXiv:2604.14228 / companion repo URL / 발행연도(2026) / 분야(cs.SE) / 대상 버전(Claude Code v2.1.88). (REQ-PA-002)
  - §2 Central Thesis — "98.4% infrastructure, 1.6% AI" (논문 자신의 명제로 framing; moai-adk 측정 아님). (REQ-PA-003)
  - §3 CC-internals consumed by moai-adk — §B.2의 5개 항목 표 (5-layer graduated-compaction taxonomy + consume-not-implement framing(REQ-PA-006) / design-space taxonomy / query-loop·withheld-recoverable-error framing / delegation ~7× token-cost / extension-mechanism context-cost ladder). (REQ-PA-003)
  - §4 In-repo citation surfaces — §B.1의 4개 표면 표 (path / 추가 SPEC / consume한 내용). (REQ-PA-004)
  - §5 Provenance & Framing Boundary — verification-claim-integrity §1.1 surface 3 인용; 논문 명제 vs moai-tree 행동 주장 구분; consume-not-implement boundary 재진술.
  - §6 References — arXiv URL / VILA-Lab repo URL / book1 convergent source / precedent `gears-paper-validation.md`.
- **substantive 보장**: bare bibliographic stub 아님. §2~§5가 consume된 내용·인용 표면·framing boundary를 포착하여 precedent 길이(≈10KB)에 준한다. (AC-PA-001 non-stub byte 임계)

### C.2 cross-reference 라인 (REQ-PA-005)

- §B 제안 target(`runtime-recovery-doctrine.md` §5)에 라인 1개 추가. 그 라인이 아카이브 파일 경로(§C.1)를 인용. 라인 1개만 — 그 외 §5 항목·본문 변경 금지.

### C.3 consume-not-implement framing (REQ-PA-006, AC-PA-006)

- 아카이브 §3에서 5-layer compaction taxonomy를 기록할 때, layer 이름(`Budget Reduction` 또는 `graduated-compaction`)과 **co-located**된 위치에 "moai-adk consumes / does not implement" 문구를 둔다 (N5 AC-CLN-003의 non-vacuous co-location anchor 패턴 차용).

## §D. Constraints (DO NOT VIOLATE)

- **PRESERVE**: 4개 인용 표면 본문 (cross-reference 라인 1개 추가 외 변경 금지); precedent `gears-paper-validation.md`; 다른 SPEC 디렉터리; runtime-managed files(`.moai/state/*`, `.moai/cache/*`, `.moai/logs/*`).
- **금지 산출물**: `.moai/research/` template mirror (dev-local — mirror 생성 금지, REQ-PA-007); 4개 표면 리팩터(라인 1개 초과, REQ-PA-007); arXiv 재검증(WebFetch/WebSearch, REQ-PA-008); Go 변경(REQ-PA-008).
- **금지 명령**: `--no-verify`, `--amend`, force-push to main.
- **사용 의무**: Conventional Commits (`feat(SPEC-DIVECC-PAPER-ARCHIVE-001): ...`), `🗿 MoAI` trailer.
- **subagent boundary (C-HRA-008 / B11)**: AskUserQuestion 금지 — blocker는 structured report로 반환.

## §E. Self-Verification (run-phase가 제출할 항목 요지)

run-phase 완료 보고 시 §D AC 8개 binary PASS/FAIL matrix + 검증 명령 실제 출력을 포함한다 (acceptance.md SSOT). 핵심:

- AC-PA-001: `ls .moai/research/dive-into-claude-code-archive.md` + `wc -c` ≥ 임계.
- AC-PA-002~004: `grep -F` for-loop (citation 4요소 / 5-layer 이름 / 4개 인용 표면 경로) — no MISSING.
- AC-PA-005: target file에서 archive 경로 grep ≥ 1.
- AC-PA-006: co-location anchor `grep -niE` ≥ 1.
- AC-PA-007/008: `git show --stat <run-commit>` 파일 목록 ⊆ {archive, runtime-recovery-doctrine.md} + spec.md frontmatter; `.go` 0개; `find` mirror absent.

## §F. Milestones (priority-ordered, no time estimates)

| Milestone | 내용 | 우선순위 |
|-----------|------|----------|
| M1 | 아카이브 파일 작성 (`.moai/research/dive-into-claude-code-archive.md` — §C.1 구조; substantive ≈10KB; REQ-PA-001~004,006) | High |
| M2 | cross-reference 라인 추가 (runtime-recovery-doctrine.md §5; REQ-PA-005) | High |
| M3 | AC matrix 8개 binary 검증 (acceptance.md SSOT) + commit + push | High |

> M1→M2→M3 순차. doc-only Tier S이므로 단일 위임으로 충분 (minimal delegation form 적용 가능).

## §G. Risks & Anti-Patterns

- **AP-1 — bare stub 아카이브**: 서지정보만 적고 consume 내용·인용 표면을 빠뜨림 → REQ-PA-001 "substantive" 위반. 완화: §C.1 §2~§5 필수, AC-PA-001 byte 임계.
- **AP-2 — neutrality 위반 target 선택**: cross-reference 라인을 template-distributed 표면(나머지 3개)에 두면 `.moai/research/` 내부 경로 인용이 neutrality CI guard에 걸림. 완화: §B에서 dev-local `runtime-recovery-doctrine.md` §5를 target으로 확정.
- **AP-3 — scope creep (문서 리팩터)**: cross-reference 라인 1개를 넘어 4개 인용 표면을 "정리"하려는 충동 → REQ-PA-007 위반. 완화: §D PRESERVE list, AC-PA-007 `git show --stat` ⊆ 2-file 검증.
- **AP-4 — arXiv 재검증**: WebFetch로 논문을 다시 fetch → REQ-PA-008 위반 + 불필요 토큰. 완화: 인용은 established, in-repo Read만.
- **AP-5 — consume-implement 혼동**: 5-layer를 moai-adk가 구현하는 것처럼 기술 → REQ-PA-006 위반. 완화: §C.3 co-location framing, AC-PA-006 anchor.
- **AP-6 — template mirror 생성**: `.moai/research/`에 mirror를 만들려 함 → REQ-PA-007 위반. 완화: dev-local 확인(`find` empty), §D 금지 산출물.

## §H. Cross-References

- `.moai/specs/SPEC-DIVECC-PAPER-ARCHIVE-001/spec.md` — §C requirements (REQ-PA-001~008) SSOT.
- `.moai/specs/SPEC-DIVECC-PAPER-ARCHIVE-001/acceptance.md` — AC matrix + Given-When-Then SSOT.
- `.moai/research/gears-paper-validation.md` — precedent 아카이브 house style.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §5 — cross-reference 라인 target (제안).
- `.moai/specs/SPEC-DIVECC-COMPACTION-LAYER-NAMING-001/` — sibling N5 (인용 provenance + dev-local mirror 관측 모델).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form.
