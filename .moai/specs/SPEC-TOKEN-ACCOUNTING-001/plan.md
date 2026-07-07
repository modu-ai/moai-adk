# Implementation Plan — SPEC-TOKEN-ACCOUNTING-001

> plan.md는 파생 실행 계획이다. WHAT/WHY의 SSOT는 spec.md. 본 문서는 HOW의 골격(마일스톤/제약/리스크)이며 함수명·시그니처 등 세부 설계는 run-phase 소관.

## §A Context

Token-Economy Epic 1/4. per-SPEC 토큰 소비를 영속·감사 가능한 측정값으로 만든다.
전면 재구축이 아닌 **확장(EXTEND)** 전략: 기존 인메모리 Tracker(`internal/runtime/budget.go`)와
transcript 읽기 패턴(`internal/statusline`)을 소스로 참고하되, 누적 과금 토큰 합산은
transcript `message.usage`를 별도로 읽는 신규 소형 패키지로 처리한다.

Tier: **M** (표준). 근거는 §D.

## §B Known Design Challenges (spec.md §B 대응 해소)

### B-attribution — 세션↔SPEC 다대다 귀속 (가장 어려운 설계점)

한 Claude Code 세션이 여러 SPEC을 넘나들 수 있고(transcript는 per-session이지 per-SPEC이
아님), `/clear`로 phase마다 세션이 갈릴 수도 있다. 세 옵션을 평가한다.

| 옵션 | 방식 | 정확도 | 복잡도 | 판정 |
|------|------|--------|--------|------|
| (i) per-SPEC delta | plan-start와 sync-close 두 측정점 delta | 세션 전용 시 정밀 | 두 write-point 필요(run-start baseline) | 후속 정밀화 |
| (ii) session-total snapshot | sync 시점 세션 총합 1점 | 다중-SPEC 세션에서 과다계상 | 최저 | 너무 coarse |
| **(iii→채택) session-set 합산** | SPEC lineage에 기록된 session-UUID 집합의 transcript를 sync-close 단일 시점에 합산 | 세션 전용(`/clear` per-phase 규율) 시 정밀, 공유 시 과다계상 → 신뢰도 플래그 | 낮음(단일 시점, 신규 baseline write 불필요) | **채택 (MVP)** |

**채택: (iii) Sync-close session-set 합산.** 이유:
- 단일 측정 시점(sync-close, manager-docs 소유)만 필요 → run-start baseline write 불요 → 가장 단순.
- SPEC lineage(progress.md의 `source_session_id` 기록 + 활성 sync 세션)로 세션-집합을 구성해
  각 transcript `message.usage`를 합산 → per-SPEC 수치를 얻는다.
- (ii)의 "sync 세션 1개만"보다 정확(전체 lineage 반영), (i)의 두 측정점보다 단순.

**명시된 정확도 한계 (verification-claim-integrity — 미검증 정밀도 주장 금지)**:
1. 한 세션이 `/clear` 없이 **여러 SPEC을 인터리브**하면 그 세션 delta 전체가 close 대상 SPEC에
   과다계상된다. 메커니즘은 세션 내부 인터리브를 구분할 수 없다 →
   `token_attribution_confidence: low` 로 플래그.
2. 세션 lineage가 environment-fallback(`not-available`)이면 활성 sync 세션 1개로 폴백 →
   `low` 신뢰도.
3. **서브에이전트 내부 토큰**이 orchestrator transcript `message.usage`에 포함되는지는
   플랜 시점 **미검증**. 포함 여부에 따라 값이 달라질 수 있음 → residual risk로 명시하고
   "billing-grade 정밀 수치가 아닌 **측정 baseline**"임을 §I 필드와 문서에 못박는다.

`high` 신뢰도 조건: SPEC의 모든 기여 세션이 SPEC-전용(다른 SPEC이 동일 UUID를 참조하지 않음,
`moai spec audit`/lineage 교차확인으로 판정) 일 때.

### B-noleak — 템플릿 중립성 (CLAUDE.local.md §15/§25)

본 SPEC의 코드는 **런타임/개발 도구**이며 배포 템플릿 콘텐츠가 아니다. 따라서:
- 신규 패키지 `internal/tokenusage/**` 와 확장 대상 `internal/spec/**`, `internal/cli/**` 는
  **`internal/template/templates/` 트리에 절대 두지 않는다.** (본 SPEC 코드는 template tree
  OUT 확정.) → 중립성/누출 CI 가드(`internal_content_leak_test.go`,
  `template-neutrality-check.yaml`) 대상 밖.
- progress.md Section Map SSOT 문서(`spec-frontmatter-schema.md`)에 §I를 추가하는 편집은
  `.claude/rules/` (dev-live) 대상이며 template mirror 여부는 run-phase에서 확인(대개 rule 문서는
  template subset 규율 대상이나, SPEC-ID/내부날짜 누출 없이 §I 스키마 설명만 추가하므로 §25 위반 아님).

### B-existing — 중복 금지 (재구축이 아닌 확장)

- SPEC-TOKEN-EFFICIENCY-001 = always-loaded 75K 정적 트립-와이어 가드(회귀 방지). 별개.
- SPEC-TOKEN-001 = skill-count 축소. 별개.
- `internal/runtime/budget.go` Tracker = 인메모리 warning-first per-agent. **EXTEND 참고**(재구축 금지);
  본 SPEC은 이를 변경하지 않고 별도의 누적-합산 baseline을 추가한다.

### B-frontmatter — 12-field canonical schema

spec.md frontmatter는 canonical 12필드(+optional `era: V3R6`) 준수. snake_case alias 없음.
status: draft. (schema 검증 결과 §E.)

### B-scope-discipline — 병렬 세션 미커밋 변경 무접촉

작업 트리의 병렬-세션 미커밋 변경(`M internal/cli/glm.go`, `M internal/config/defaults.go`,
`M .moai/config/sections/llm.yaml`, `SPEC-CLI-UIKIT-KERNEL-001/` 등)은 **건드리지·스테이지하지·참조하지 않는다.**
쓰기는 오직 `.moai/specs/SPEC-TOKEN-ACCOUNTING-001/` 내부.

## §C Pre-flight (완료됨)

- [x] `internal/runtime/budget.go` 읽음 — EXTEND base(Tracker, RecordCall, warning-first).
- [x] `internal/statusline/context_usage.go` + `usage.go` + `memory.go` 읽음 — statusline은 **컨텍스트
  창 점유 스냅샷**(current_usage/used_percentage)을 stdin에서 읽지 transcript를 합산하지 않음을 확인.
- [x] `internal/spec/audit.go` + `era.go` 읽음 — 파서 load-bearing 토큰(`§E.2/3/4/5`,
  `sync_commit_sha`, `mx_commit_sha`) 확인. 감사 확장 지점(`AuditResult`, `spec_audit.go`) 확인.
- [x] transcript JSONL 실측 — `~/.claude/projects/<hash>/<uuid>.jsonl` 의 assistant 레코드
  `message.usage` 4필드 확인 + 11턴 합산 실측.
- [x] `.moai/state/context-usage.json` shape 확인.
- [x] Section Map SSOT 확인 — §F/§G/§H 사용중, **§I 미사용(free)** 확인 → §I 배정.

## §D Constraints & Tier 근거

- plan-phase ONLY. Go 코드/테스트/구현 없음.
- **Tier M** 근거: 단일 도메인(internal Go 도구), 산출물 3개(spec/plan/acceptance) + progress skeleton,
  run-phase 파일 표면 ~5–8개(신규 `internal/tokenusage` 패키지 2–3파일 + `audit.go`/`spec_audit.go`
  확장 + 세션 lineage 헬퍼 + Section Map 문서 1편집), ~300–600 LOC. agent/template 표면 없음.
  → Tier S(1–2 파일 trivial)보다 크고, Tier L(다중 도메인·10+ 파일·agent 파일 변경)보다 작음.
- GEARS/EARS 포맷. AC는 기계적 검증 가능해야 함(vacuous 금지).
- `.moai/specs/SPEC-TOKEN-ACCOUNTING-001/` 밖 파일 수정 금지.

## §E Self-Verification (frontmatter schema)

12 canonical 필드 전수 확인:
`id`(regex PASS) · `title`(quoted) · `version`("0.1.0") · `status`(draft, 8-enum) ·
`created`(2026-07-07 ISO) · `updated`(2026-07-07 ISO) · `author` · `priority`(P1) ·
`phase`("v3.1.0") · `module`("internal/tokenusage") · `lifecycle`(spec-anchored) ·
`tags`(comma string). snake_case alias 없음. optional `era: V3R6`(신규 SPEC transient
misclassification 방지 — skeleton §E.2 존재+sync_sha 공백 시 H-3→V3R5 오분류를 pin으로 차단).

## §F Milestones (우선순위 기반, 시간 추정 없음)

- **M1 (선행) — transcript 파서 (`internal/tokenusage`)**: 세션 transcript JSONL을 읽어
  assistant 턴의 `message.usage` 4필드를 합산. malformed 라인/부재 파일 관용(skip, no-panic).
  read-only. cache-hit ratio 계산. TDD. [REQ-TA-001,002,003,004,013]
- **M2 — 귀속 레이어 (attribution)**: progress.md lineage에서 session-UUID 집합 해석 →
  session-set 합산 + 신뢰도(`high`/`low`) 판정 + lineage 부재 시 현재-세션 폴백.
  [REQ-TA-005,006,007] (M1 의존)
- **M3 — progress.md §I writer + Section Map SSOT**: sync-close 시 `## §I Token Accounting`
  섹션에 신규 필드(`tokens_spent` 등) 기록. era.go 파서 무충돌(§E.N/SHA 미변경) 회귀 테스트.
  Section Map SSOT 문서에 §I 행 추가. [REQ-TA-008,009,010] (M2 의존)
- **M4 — audit 표면 (`moai spec audit`)**: `AuditResult`/`DriftFinding`에 `tokens_spent` 노출
  (§I 파싱). 미기록 SPEC은 `null`/omit. JSON 필드 필수, 테이블 컬럼은 nice-to-have.
  [REQ-TA-011,012] (M3 의존)

의존 그래프: M1 → M2 → M3 → M4 (순차). 병렬화 불가(각 단계가 이전 산출물 소비).

## §G Anti-Patterns (회피)

- statusline `context-usage.json`(현재 점유 스냅샷)을 누적 소비 소스로 오용 → transcript
  `message.usage` 합산을 별도로 읽어야 함 (spec.md §A.2 구분).
- `§E.N` 헤딩/`sync_commit_sha`/`mx_commit_sha` 를 재사용·개명 → 파서 침묵 붕괴. §I 신규 letter + 신규 필드명만.
- session-set 합산 값을 "정밀 billing 수치"로 과대주장 → 신뢰도 qualifier + baseline 성격 명시 필수.
- 신규 코드를 `internal/template/templates/`에 배치 → 배포 중립성 위반. template tree OUT 확정.
- 병렬 세션 미커밋 변경 스테이지/참조 → 스코프 이탈.

## §H Cross-References

spec.md §E와 동일. 핵심: `internal/spec/era.go`(무충돌), `internal/runtime/budget.go`(EXTEND base),
`spec-frontmatter-schema.md` §progress.md Section Map(§I SSOT), verification-claim-integrity(정밀도 제약).
