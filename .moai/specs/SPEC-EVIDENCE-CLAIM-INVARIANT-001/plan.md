# SPEC-EVIDENCE-CLAIM-INVARIANT-001 — 구현 계획 (plan.md)

> Tier S (minimal). 산출물: 신규 doctrine 규칙 파일 1개 + template mirror. 코드 없음.

## A. Context (배경)

본 SPEC은 IMP-06 — fable-ish(chrisryugj) 13-에이전트 분석 로드맵의 최종 채택 항목이다. SPEC-STOP-EVIDENCE-GATE-001(IMP-02/03)이 제공한 **기계적 advisory detection**(code-session false-success 1개 형태)을, **정책/doctrine codification**으로 보완한다. 동기 결함 클래스는 `L_manager_docs_false_backfill_report`("관측하지 않은 검증을 주장").

run-phase의 단일 deliverable은 markdown doctrine 규칙 파일(local + template mirror)이며, 세 구성요소(invariant / baseline 귀속 / 5-섹션 리포트 포맷)를 codify한다.

## B. Known Issues (알려진 이슈 / 위험)

- **가장 큰 run-phase 위험 = template mirror 중립성**(아래 §D.1). local 사본은 SPEC provenance를 담아도 되지만, mirror 사본은 내부 흔적(SPEC ID, REQ 토큰, 내부 날짜, commit SHA, `feedback_`/메모리 ref, `CLAUDE.local` ref, `/Users/` 경로)이 0이어야 한다. CI 가드 `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml`가 이를 강제한다.
- **내용 복제 금지**(SSOT, coding-standards.md): 기존 규칙(`agent-common-protocol.md` § Skeptical Evaluation Stance, `moai-constitution.md` § Agent Core Behaviors #6, `manager-develop-prompt-template.md` § E, `verification-batch-pattern.md`, `output-styles/moai/moai.md` §8)을 복사하지 말고 cross-reference한다.

## C. Pre-flight (사전 점검)

- `go build ./...` — run-phase가 Go를 건드리지 않으므로 **불요**. 단, template mirror가 `embedded.go` 재생성을 유발하면(`//go:embed all:templates`는 compile-time) `make build`로 재컴파일이 필요할 수 있음(§F M3에서 확인).
- 대상 디렉터리 사전 확인 완료: `.claude/rules/moai/core/` 및 `internal/template/templates/.claude/rules/moai/core/` 모두 존재.
- 신규 파일명 `verification-claim-integrity.md`는 양 경로에서 충돌 없음(미존재 확인 완료).

## D. Constraints (run-phase 제약 — DO NOT VIOLATE)

### D.1 Template 중립성 (단일 최대 위험)

[HARD] template mirror 사본(`internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md`)은 언어/내부-중립이어야 한다. 금지 클래스:

- 내부 SPEC ID(`SPEC-EVIDENCE-CLAIM-INVARIANT-001`, `SPEC-STOP-EVIDENCE-GATE-001` 등)
- REQ/AC 토큰(`REQ-...`, `AC-ECI-...`)
- 내부 작업 날짜(`2026-06-15`)
- commit SHA(`fdf780ae1` 등)
- `feedback_` / 메모리 파일 ref (`L_manager_docs_false_backfill_report` 등)
- `CLAUDE.local` ref
- `/Users/` 절대 경로

**divergence 명시**: local 사본은 SPEC provenance(IMP-06 출처, 선행 SPEC ID, 동기 결함 클래스명)를 담을 수 있다. mirror 사본은 동일 정책을 **generic prose + 메커니즘 설명 + 영구 규칙 인용**으로 재서술한다(예: 동기 결함을 "an agent claiming a verification it did not actually observe"처럼 일반화). 허용 인용: MoAI-ADK 시스템 식별자, 영구 규칙 파일 경로(`agent-common-protocol.md` 등).

### D.2 파일 경로 결정 — `core/` 선택 근거

`core/` vs `quality/` 갈림길에서 **`core/` 선택**. 근거: invariant가 (a) 오케스트레이터 자가-보고(output-style §8)와 (b) manager 에이전트 §E를 **동시에 cross-cutting** 결속한다. `core/`는 `agent-common-protocol.md`·`moai-constitution.md` 같은 cross-cutting 정책의 거처이며, `quality/`는 품질 게이트 전용이라 결속 범위가 좁다. 따라서 `.claude/rules/moai/core/verification-claim-integrity.md` 채택.

### D.3 spec-lint clean

[HARD] doctrine 규칙 파일 자체는 SPEC이 아니므로 spec-lint 대상이 아니다. 단, 본 SPEC의 `spec.md`는 `### X.1 Out of Scope —` (H3) 서브섹션 + ≥1 list item을 가져 `MissingExclusions`(internal/spec/lint.go:681 `OutOfScopeRule`)를 통과한다(검증 완료: H3 `###` prefix + `out of scope` 문자열 + `-` list item 존재).

### D.4 PRESERVE (불변 유지)

- 기존 규칙 파일 내용 변경 금지(M3의 최소 1줄 cross-ref 포인터 추가는 예외 — 양방향 cross-ref 중 한쪽만 택해도 AC-ECI-007 충족 가능).
- `embedded.go`는 자동 생성물 — 직접 편집 금지(`make build`로만 재생성).

## E. Self-Verification (run-phase 완료 시 manager-develop가 제시할 증거)

run-phase 종료 시 §E 자가검증은 AC-ECI-001..007을 binary PASS/FAIL로 제시한다. 특히:

- AC-ECI-006 중립성 grep을 **mirror 파일 대상**으로 실행하고 verbatim 출력(0 매치)을 제시.
- AC-ECI-003/004/005/007 grep 출력을 verbatim 제시(요약 금지 — 본 SPEC이 codify하는 invariant 자체를 run-phase 보고가 준수).

## F. Milestones (Tier S — 최소 3개)

### M1 — local doctrine 규칙 파일 작성

`.claude/rules/moai/core/verification-claim-integrity.md` 생성. 세 구성요소 codify:
1. "no unobserved-verification-claim" invariant + 두 결속 표면(오케스트레이터 self-report + manager-agent §E) 명명 (AC-ECI-003)
2. baseline 무결성 귀속 (AC-ECI-005)
3. 5-섹션 증거-부착 리포트 포맷 (주장/증거/baseline 귀속/미검증/잔여 위험) (AC-ECI-004)
+ ≥4 cross-reference (AC-ECI-007). local 사본은 SPEC provenance 허용.
→ AC-ECI-001 충족.

### M2 — template mirror 작성 (중립)

`internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` 생성. M1과 동일 정책을 **내부-중립 prose**로 재서술(§D.1 금지 클래스 0). → AC-ECI-002 + AC-ECI-006 충족.

### M3 — cross-reference 배선 + 검증

- (택1) output-style `moai.md` §8 또는 manager-develop §E에서 신규 규칙으로의 포인터 1줄 추가, **또는** 신규 규칙이 이들을 참조함을 보장(AC-ECI-007은 신규 규칙→기존 규칙 방향만으로 충족 가능하므로, cross-ref 추가는 SHOULD).
- 검증: AC-ECI-003..007 grep, AC-ECI-006 중립성 grep(mirror 대상).
- template mirror가 `embedded.go` 재생성을 유발하면 `make build` 실행 후 재컴파일 확인(`//go:embed all:templates`는 compile-time이므로 mirror 추가 시 `make build` 필요). 단, 신규 파일 추가만으로 기존 코드 동작 변경은 없음.

## G. Anti-Patterns (피해야 할 것)

- mirror 사본에 SPEC ID/REQ/날짜/SHA를 그대로 복사 (CI 가드 fail).
- 기존 규칙 내용을 신규 규칙에 복사 (SSOT 위반 — cross-ref만).
- doctrine을 unfalsifiable하게 작성 (STOP-EVIDENCE-GATE iter-1 교훈 — 모든 AC를 grep/파일-존재로 검증 가능하게).
- 5-섹션 중 §4 미검증(Gaps)을 생략 (invariant의 핵심 방어선 — 누락 시 침묵 미관측 주장 차단 실패).

## H. Cross-References

- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12-field canonical schema SSOT)
- `internal/spec/lint.go` `OutOfScopeRule` (MissingExclusions, H3/H4 + list item)
- `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml` (mirror 중립성 CI 가드)
- `CLAUDE.local.md` §2 (Template-First + mirror), §15 (언어 중립성), §25 (Template Internal-Content Isolation)
- 선행 SPEC: `.moai/specs/SPEC-HOOK-DISCIPLINE-WIRING-001/`, `.moai/specs/SPEC-STOP-EVIDENCE-GATE-001/`
