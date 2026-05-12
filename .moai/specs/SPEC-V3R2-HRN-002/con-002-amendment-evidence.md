## CON-002 Amendment Evidence — SPEC-V3R2-HRN-002

HRN-002는 FROZEN-zone 조항 `.claude/rules/moai/design/constitution.md §11` (GAN Loop Contract)에 새 하위 조항 `§11.4.1 Evaluator Memory Scope (Principle 4)`를 삽입합니다. CON-002 §5에 따라 5개 안전 레이어 증거가 amendment 착지 전에 제공되어야 합니다.

이 문서는 SPC-001 con-002-amendment-evidence.md (PR #870, 2026-05-13) 패턴을 그대로 따릅니다.

### Layer 1 — Frozen Guard

- **Mechanism**: FROZEN-zone 조항을 수정하는 쓰기 작업 전에 명시적 amendment 권한 확인.
- **Evidence**:
  - Amendment vehicle: SPEC-V3R2-HRN-002 (이 SPEC), spec.md §1에서 `amends design-constitution §11`로 선언.
  - FROZEN-zone 태그: `.claude/rules/moai/design/constitution.md` 삽입 블록에 `<!-- @MX:WARN: FROZEN-zone amendment per CON-002 §5 Layer 1 (Frozen Guard); modifications require full 5-layer cycle -->` + `<!-- @MX:REASON: §11.4.1 inserted by SPEC-V3R2-HRN-002; per-iteration ephemeral evaluator judgment is a constitutional invariant -->` 배치.
  - Zone-registry 등록: `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-153 엔트리 추가 (이 파일과 동일 커밋).
- **Verdict**: PASS — amendment는 명시적이며 FrozenGuard 체크를 통과.

### Layer 2 — Canary Check

- **Mechanism**: 최근 3개 완료 GAN-loop 프로젝트를 fresh-memory evaluator 프로토콜로 재평가; 어떤 프로젝트도 > 0.10 score drop 없음.
- **Evidence**:
  - Canary log: `.moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt` (2026-05-13).
  - 결과: **CanaryUnavailable** — v3 design corpus에 완성된 GAN-loop 평가 아티팩트 없음.
  - 영향: 없음. §11.4.1은 미래 GAN-loop 실행에 대한 preventive rule이며, 기존 평가 실행이 없으므로 회귀 없음.
  - REQ-CON-002-020 준수: 3개 미만 subject 시 CanaryUnavailable 명시적 기록.
- **Verdict**: PASS (CanaryUnavailable by corpus absence — no regression possible).

### Layer 3 — Contradiction Detector

- **Mechanism**: amendment와 기존 FROZEN/EVOLVABLE 조항 간의 충돌 스캔.
- **Evidence**:
  - §11.4.1은 §11.4 Sprint Contract Protocol의 연장선. §11.4의 "Sprint Contract state durable" + "passed criteria carry forward" 조항과 완전히 일치.
  - EvaluatorConfig.MemoryScope FROZEN 값은 evaluator_mode (final-pass / per-sprint), pass_threshold (0.75), escalation_after (3) 등 기존 harness 설정과 독립적 — 충돌 없음.
  - design.yaml + harness.yaml에 동시 추가됨 (T-08); 두 파일의 값이 동일하게 per_iteration.
  - EVOLVABLE zone 조항 (adaptation weights, rubric criteria 등)과 충돌 없음 — §11.4.1은 evaluator context scope만 규정.
- **Verdict**: PASS — 충돌 없음.

### Layer 4 — Rate Limiter

- **Mechanism**: v3.x 릴리스 사이클당 ≤3 FROZEN amendment 상한.
- **Evidence**:
  - v3.x 사이클 FROZEN amendment 인벤토리 (2026-05-13 기준):
    1. SPC-001 (PR #870, 2026-05-13) — hierarchical AC schema; main 착지.
    2. HRN-002 (이 SPEC) — §11.4.1 Evaluator Memory Scope; 착지 예정.
    3. (예비 슬롯; 현재 미사용)
  - Count: 2 of 3 사용. Rate limit 준수.
  - SPC-001 con-002-amendment-evidence.md §Layer 4 참조: HRN-002를 #2로 명시 기록함 (PR #870).
- **Verdict**: PASS — rate cap 내.

### Layer 5 — Human Oversight

- **Mechanism**: 착지 PR description에 maintainer 승인 (타임스탬프 + reviewer 식별자) 기록.
- **Evidence**:
  - Plan-auditor 검증: plan-auditor PASS (2026-05-13, PR #873).
  - Run-phase HumanOversight: maintainer (Goos Kim, bobby@afamily.kr) 승인이 착지 PR description에 타임스탬프와 함께 기록되어야 함.
  - **Approval**: Goos Kim <bobby@afamily.kr> @ 2026-05-13 — orchestrator AskUserQuestion (CON-002 §5 Layer 5 HumanOversight gate) "승인 + PR 생성" 선택. 본 파일과 착지 PR description 양쪽에 기록됨.
- **Verdict**: PASS — maintainer 승인 완료, 5개 레이어 모두 PASS.

## Overall Verdict

5/5 레이어 PASS. CON-002 §5 amendment protocol 완전 충족. main 머지 적격.
