# SPEC-HARNESS-EVO-RUN-REPORT-001 — Acceptance Criteria

Tier: **M** · development_mode: **tdd** · 총 AC: **10** (AC-HRR-001 ~ AC-HRR-010)

모든 AC는 기계 검증 가능(mechanically-verifiable)을 우선하며, 문서 표면 AC는 grep/파일 존재 기준으로 명시한다.

---

## §D AC 매트릭스

### AC-HRR-001 — manifest `learning` 블록 스키마 수용 (REQ-HRR-001)

- **Given** `learning` 블록(`enabled/tier/confidence_floor/max_findings_per_run`)을 가진 v4 manifest fixture
- **When** `v4manifest` 파서가 manifest를 로드하면
- **Then** `Manifest.Learning`이 nil이 아니고 4개 필드가 정확히 파싱된다
- **And Given** `learning` 블록 없는 기존 8-필드 manifest fixture (legacy)
- **When** 동일 파서가 로드하면
- **Then** `Manifest.Learning == nil` + 파싱 성공(거부 없음) — 하위 호환
- **검증**: `go test -run TestManifestLearningBlock ./internal/harness/v4manifest/` → PASS

### AC-HRR-002 — `learning.tier` 어휘 SSOT 정합 (REQ-HRR-002)

- **Given** `learning.tier`가 `Tier.String()` 어휘(`{observation, heuristic, rule, auto_update}`) 중 하나인 fixture
- **When** 스키마 validation이 실행되면
- **Then** actionable(`{rule, auto_update}`)은 유효+actionable, pre-actionable(`{observation, heuristic}`)은 유효+non-actionable로 처리
- **And Given** `learning.tier: "recommendation"` (PIPE-REPAIR가 제거한 병렬 어휘) fixture
- **When** validation이 실행되면
- **Then** 무효값으로 거부 (별도 병렬 어휘 정의 없음)
- **검증**: `go test -run TestLearningTierVocabulary ./internal/harness/v4manifest/` → PASS; `grep -n "recommendation\|approval_required" internal/harness/v4manifest/*.go` → 어휘 정의 부재(테스트 fixture 제외)

### AC-HRR-003 — Runner return-schema `findings` 표준 계약 (REQ-HRR-003)

- **Given** 개선 신호가 있는 하네스 실행
- **When** Runner의 `run()`이 반환하면
- **Then** 반환 객체에 `findings` 배열이 존재하고 각 원소가 `{surface, kind, summary, confidence, suggested_tier}` 5-필드를 가진다
- **And Given** 개선 신호가 없는 실행
- **When** Runner가 반환하면
- **Then** `findings: []`(빈 배열) — 필드 생략 없음
- **검증**: exemplar Runner 반환 스모크 — `node -e "const {run}=require('./.claude/workflows/harness-release-update-run.js'); ..."` 또는 계약 grep: `grep -n "findings" .claude/workflows/harness-release-update-run.js` → return object에 `findings` 키 존재

### AC-HRR-004 — findings confidence 출처 분리 (REQ-HRR-004)

- **Given** Runner가 findings를 방출하는 코드 경로
- **When** `findings[].confidence`가 산출되면
- **Then** 값이 실행 시점 측정/추정치이며 `learner.go`의 `defaultConfidence`(1.0) 상수를 import/재사용하지 않는다
- **And Given** 근거 없는 confidence 산출 상황
- **When** Runner가 보수 기본값을 방출하면
- **Then** 값이 `confidence_floor` 경계(0.70) + 추정임을 구분 가능
- **검증**: `grep -n "defaultConfidence\|learner" .claude/workflows/harness-*-run.js` → learner.go 상수 참조 부재; findings confidence가 하드코딩 1.0이 아님 확인

### AC-HRR-005 — doctor learning 축 검증 (REQ-HRR-005)

- **Given** `learning.tier` 오타 / `confidence_floor` 범위 밖 / `enabled:true`인데 Runner findings 미선언 fixture 하네스
- **When** `moai harness doctor`가 실행되면
- **Then** 각 결함이 ERROR-severity finding으로 보고되고 exit code 비-0
- **And Given** `learning` 블록 없는 하네스
- **When** doctor가 실행되면
- **Then** ERROR 아님(INFO note 또는 무보고), exit 0
- **검증**: `go test -run TestDoctor_LearningAxis ./internal/cli/harness/` → PASS

### AC-HRR-006 — specialist improvement-findings 필수 최종 단계 (REQ-HRR-006)

- **Given** harness specialist 에이전트 파일 (Runner 보유 하네스 — release-update 최소)
- **When** specialist 워크플로우를 검사하면
- **Then** 종료 직전 필수 improvement-findings 방출 단계가 존재하고, 방출 findings shape가 REQ-HRR-003(`{surface, kind, summary, confidence, suggested_tier}`)과 정합
- **And** 해당 단계가 AskUserQuestion 직접 호출 없이 구조화 출력(structured findings)만 반환하도록 명시
- **검증**: `grep -n "improvement.findings\|findings" .claude/agents/harness/harness-release-update-specialist.md` → 필수 방출 단계 존재; `grep -n "AskUserQuestion" .claude/agents/harness/harness-release-update-specialist.md | grep -v "blocker\|<!--"` → 직접 호출 부재

### AC-HRR-007 — post-run findings 수집 → push (REQ-HRR-007)

- **Given** 하네스 실행이 종료되고 Runner/specialist가 non-empty findings를 방출한 상태
- **When** 오케스트레이터 post-run 단계가 실행되면
- **Then** post-run doctrine 표면이 findings 수집 → 즉시 AskUserQuestion 제시를 규정하며(pull-only apply 대체), pull `apply` verb는 공존 유지
- **And Given** 빈 findings
- **When** post-run 단계가 실행되면
- **Then** AskUserQuestion 미발화(조용히 진행) 규정
- **검증**: post-run push doctrine 표면(harness.md 또는 rule 파일)에 `grep -n "post-run\|findings\|push"` → 수집+push 규정 존재; `grep -n "empty\|빈\|no findings"` → 빈 findings 미발화 규정 존재

### AC-HRR-008 — AskUserQuestion 오케스트레이터-단독 경계 보존 (REQ-HRR-008)

- **Given** post-run push 배선 + specialist findings 방출 경로
- **When** 전체 배선을 검사하면
- **Then** specialist/Runner는 AskUserQuestion 미호출; findings 방출 → 오케스트레이터 수집 → 오케스트레이터 AskUserQuestion의 비대칭 경계 준수
- **And Given** actionable(`{rule, auto_update}`) findings를 push하는 경우
- **Then** doctrine이 `harness.yaml` `rate_limit`(max_per_week/cooldown_hours SSOT) 준수를 규정
- **검증**: `grep -rn "AskUserQuestion" .claude/agents/harness/ .claude/workflows/harness-*.js` → subagent/Runner 직접 호출 부재; post-run doctrine에 `grep -n "rate_limit\|max_per_week"` → rate-limit SSOT 참조 존재

### AC-HRR-009 — Template-First + §25 중립성 + 격리 (REQ-HRR-009)

- **Given** 본 SPEC이 수정한 template-managed 표면 (post-run doctrine mirror, Builder Runner 계약 등)
- **When** run-phase 완료 후 검증하면
- **Then** template mirror가 live와 정합(`make build` 실행됨) + template content에 SPEC ID/REQ-HRR/감사 인용 부재
- **And** dev-only(exemplar Runner)/user-owned(specialist) 아티팩트가 template로 미유출
- **검증**: `grep -rn "HARNESS-EVO-RUN-REPORT\|REQ-HRR" internal/template/templates/` → 0 matches; `find internal/template/templates -path '*harness*run.js' -o -path '*agents/harness*'` → empty; `go test -run TestSplitHarnessNamespaceNoLeak ./internal/template/` → PASS

### AC-HRR-010 — 하위 호환 (기존 하네스 무영향) (REQ-HRR-010)

- **Given** legacy 하네스 3종 형태: (a) learning 블록 없는 8-필드 manifest, (b) findings 없는 Runner, (c) improvement-findings 단계 없는 specialist
- **When** 스키마 파싱 / Runner 실행 / doctor 검사가 실행되면
- **Then** 모두 유효 legacy로 처리되어 정상 동작(파싱 성공, doctor exit 0, findings는 빈 취급)
- **검증**: `go test -run "TestManifestLearningBlock|TestDoctor_LearningAxis" ./internal/harness/v4manifest/ ./internal/cli/harness/` legacy fixture 케이스 PASS; PIPE-REPAIR doctor 회귀 통과(`go test ./internal/cli/harness/`)

---

## §D.1 REQ ↔ AC 추적표

| REQ | 요약 | AC |
|-----|------|-----|
| REQ-HRR-001 | manifest `learning` 블록 스키마 | AC-HRR-001, AC-HRR-010(a) |
| REQ-HRR-002 | learning tier 어휘 SSOT 정합 | AC-HRR-002 |
| REQ-HRR-003 | Runner return `findings` 계약 | AC-HRR-003, AC-HRR-006(정합), AC-HRR-010(b) |
| REQ-HRR-004 | findings confidence 출처 분리 | AC-HRR-004 |
| REQ-HRR-005 | doctor learning 축 | AC-HRR-005, AC-HRR-010(c) |
| REQ-HRR-006 | specialist improvement-findings 단계 | AC-HRR-006 |
| REQ-HRR-007 | post-run findings 수집 → push | AC-HRR-007 |
| REQ-HRR-008 | AskUserQuestion 경계 보존 | AC-HRR-008 |
| REQ-HRR-009 | Template-First + 중립성 + 격리 | AC-HRR-009 |
| REQ-HRR-010 | 하위 호환 | AC-HRR-010 (+001/005 legacy 케이스) |

10 REQ ↔ 10 AC, 전 REQ 커버.

---

## §D.2 Edge Cases

- **EC-1 (learning 블록 부분 필드)**: `learning`에 일부 필드만(`enabled`만) 존재 — 나머지는 기본값(tier=observation, confidence_floor=0.70, max_findings_per_run=합리적 기본)으로 채우거나 명시 무효 처리. run-phase에서 partial-block 정책 확정
- **EC-2 (findings 상한 초과)**: Runner가 `max_findings_per_run`보다 많은 findings 생성 — 상한까지 절단(truncate) + 절단 사실 note. push는 절단 후 목록만
- **EC-3 (confidence < confidence_floor)**: findings confidence가 floor 미만 — proposal 후보 부적격(non-actionable)으로 분류, push는 하되 actionable 표기 제외
- **EC-4 (Runner 없는 thin 하네스 findings)**: github/release(Runner 없음) — Runner findings 경로 부재. specialist findings 방출(REQ-HRR-006)만으로 push 성립하는지 run-phase 확정(§F 재검증)
- **EC-5 (legacy specialist push)**: improvement-findings 단계 없는 legacy specialist 실행 — findings 빈 목록으로 간주, push 미발화(AC-HRR-007 빈 findings 경로)

---

## §D.3 Quality Gate 기준 (Definition of Done)

- [ ] 10개 AC 전량 PASS (기계 검증 명령 출력으로 증명 — verification-claim-integrity §3.2)
- [ ] `go test ./...` full suite green
- [ ] touched pkg(`v4manifest`, `cli/harness`) coverage ≥85%
- [ ] `go test -race` green (learner/hook 동시성 경로 무회귀)
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] `golangci-lint run` NEW issue 0
- [ ] `moai harness doctor` learning 축 포함, exemplar 0 ERROR-severity findings
- [ ] §25 중립성 grep 0 matches + dev-only/user-owned 격리 grep empty
- [ ] `TestSplitHarnessNamespaceNoLeak` PASS
- [ ] PIPE-REPAIR doctor 테스트 회귀 무손실
- [ ] §E 제외(learner.go confidence 실측화) **미변경** 확인 (`git diff internal/harness/learner.go` → 본 SPEC scope 변경 없음)
- [ ] subagent boundary: specialist AskUserQuestion 직접 호출 0
