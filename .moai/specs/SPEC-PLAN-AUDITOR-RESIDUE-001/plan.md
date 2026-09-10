# SPEC-PLAN-AUDITOR-RESIDUE-001 — 구현 계획

## §A Context

카드 t450 — t387에서 t367 대기로 미뤄둔 plan-auditor 조항 잔여 축. 전제(t367 착지)는 이번 실행에서 merge-base로 검증됐다: 18fc2c9ef는 origin/develop 7835148d3의 조상. 작업 트리: 워크트리 t450, 브랜치 WT-plan-auditor-residue (develop 팁에서 신설히 생성 — 병합 끝난 t387 브랜치에 커밋을 얹지 않는 카드 절차 준수).

### 발견된 지연 조항 (2건, 출처 명시)

1. **t387 곁말 규약의 plan-auditor 반영** — 출처: 커밋 f47d7f5a9 본문("Agent-definition reflection (plan-auditor / sync-auditor) deferred with t386's export-mandate clause until the lane-9 rubric/ownership cards settle") + `.moai/reports/t387/verdict.md` Gaps 절. 내용: `audit-artifact-convention.md` § Side-talk(:87)의 3규칙 — 별도 미검증 절 분리 / 측정 지시 형태 / 3단 상태 라벨(measured·inferred·assumption).
2. **t386 반출 조항의 plan-auditor 측** — 출처: 커밋 4244c4a06 본문("The plan-auditor counterpart stays deferred until t367 (rubric revision) closes"). sync-auditor에 착지한 문면이 모델이다: [HARD] Export mandate — 같은 턴 파일 반출, gitignored 위치 금지. sync-auditor 문면은 템플릿 사본 기준 `sync-auditor.md` Output Format 절 다음 줄에 [HARD] 단락으로 들어가 있다 — plan-auditor도 같은 자리·같은 문체를 따른다.

### t367 :72 상호작용 판정

t367(18fc2c9ef)이 고친 영역은 M3 § Scope의 Score 1.0 루브릭 불릿 목록(현행 :72 Event-detected 불릿, :73 Unwanted legacy-only 불릿)이다. 이 SPEC이 고치는 영역은 § Output Format(현행 :395 부근 쓰기 경로 지시)과 그 뒤에 붙는 반출·곁말 조항 — 텍스트적 겹침이 없다. 당초 충돌 우려는 같은 파일을 다른 레인이 여는 것에 대한 예방이었고, t367이 착지한 지금 그 우려는 소멸했다. AC-007이 :72 문면 보존을 회귀 가드로 고정한다.

### t443 위협 상태 (이번 실행 측정)

`go test ./internal/template/agentemit/...` → `--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)` / `golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch)`. 원인은 4244c4a06이 sync-auditor.md를 고치고 방출물을 재생성하지 않은 것 — 수리 소관은 t443(t444가 보존). `make agents-emit` 전체 실행은 이 FAIL에 막혀 있다.

미해결 질의는 없다 — t443 우회는 카드 지시 + t367 선례(커밋 2549f775f)로 이미 정해졌다. 방출 도구에는 범위 모드가 없으므로 실제 메커니즘은 전체 재생성(`AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission` — 커밋된 .toml 전부 재작성) 후 sync-auditor.toml을 develop 값으로 복원하는 것이고, 카탈로그는 `gen-catalog-hashes --entry plan-auditor`로 갱신한다(커밋 본문 "t443-owned drift deliberately NOT repaired" 기록 포함). AC-006의 한정 합격 판정(plan-auditor 녹색 + 적색 sync-auditor.toml 1건)은 이 대로 유지된다.

## §B Known Issues

- plan-auditor 쌍둥이 사이에 기존 드리프트 2 hunk(D7-1 예시 식별자, Tier-resolved ceiling 문단) — REQ-004/Out of Scope로 기록만.
- `make build` 전 `agents-emit-check`가 t443 드리프트로 실패한다 — 빌드는 t443이 끝나기 전까지 `go build ./...` 등 대체 검증으로 한다(빌드 자체는 방출물 불일치를 막지 않는다 — agents-emit-check가 막는 것).

## §C Pre-flight

- [ ] 작업 시작 시 `git rev-parse --short HEAD` 재판독 — 7835148d3에서 갈라졌는지 확인
- [ ] 템플릿 미러 존재 확인: `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` 와 `internal/template/templates/.moai/docs/audit-artifact-convention.md` (양쪽 모두 존재 측정됨 — :407 상당 / FORBIDDEN 문단)
- [ ] catalog.yaml의 plan-auditor 항목 해시 형식 확인 (t441 카탈로그 커버리지)

## §D Constraints

- 쌍둥이 [HARD]: 에이전트 정의와 룰·규약 문서 편집은 항상 로컬 + 템플릿 미러 동시 편집. C3 방출물(.toml)은 손편집 금지 — 템플릿 쪽에서 재생성.
- 템플릿 쪽 신규 문면은 중립성 금지 클래스(C1-C8) 0매치: SPEC-ID 토큰, 카드 id(t450 등), 커밋 SHA, 내부 날짜 금지. 경로 플레이스홀더(`<card-id>`, `<SPEC-ID>`)는 동일 표현이 이미 템플릿 sync-auditor.md에 착지해 있으므로 안전한 선례다.
- 레인 검증 범위: 건드린 패키지 = `internal/template/agentemit` (골든 테스트) + 카탈로그 해시 패리티. 전체 스위트 로컬 실행 금지(CI 몫).
- 커밋마다 카드 id(t450) 명시. 증거는 브랜치에 커밋.

## §E Self-Verification

- 각 AC의 검증 명령은 acceptance.md §D 매트릭스에 RED-now 셀과 함께 고정. 커밋 시점에 재측정해 기록.
- 방출물 검증: `go test ./internal/template/agentemit/...` — plan-auditor.toml 녹색, sync-auditor.toml 적색은 t443 소관 기록으로 남김(범위 재생성 후에도).
- 카탈로그: `go test ./internal/template/...` 중 TestCatalogHashParity / TestManifestHashFormat — plan-auditor 녹색.
- 중립성: `.github/workflows/template-neutrality-check.yaml`의 패턴과 §25.1 카탈로그 기준 0매치 grep.

## §F Milestones

우선순위만 사용 — 시간 예측 없음. 결정 가역성 순(바뀔 가능성 높은 결정부터):

**M1 (Priority High) — 조항 문안 확정(양쪽 쌍둥이 동일 문면)**
가장 바뀔 가능성이 높은 결정: 반출 조항 문안, 곁말 반영 문안, § Output Format 경로 지시의 새 문구. sync-auditor의 착지 문면(4244c4a06)을 문체 선례로 삼아 초안을 만들고, 템플릿 중립성 제약(금지 토큰 배제)을 통과시킨 문면을 확정한다. 산출: 조항 3종 최종 문안 (REQ-001/002/003/004 대응).

**M2 (Priority High) — 쌍둥이 적용 + 교차참조 갱신**
확정 문안을 `.claude/agents/moai/plan-auditor.md`와 템플릿 미러에 동일하게 적용(REQ-001~004). 이어서 `spec-workflow.md` § Report Persistence(양쪽 미러)와 `audit-artifact-convention.md` § What makes the convention stick / § Cross-references(양쪽 미러)를 REQ-005대로 갱신. :72 문면은 건드리지 않는다(REQ-007).

**M3 (Priority High) — 방출물 재생성 + 카탈로그 해시 (t443 우회)**
템플릿 쪽 plan-auditor.md 변경에 맞춰 `AGENTEMIT_UPDATE=1` 전체 재생성을 실행한 뒤 sync-auditor.toml만 develop 값으로 복원한다 — 방출 도구에 범위 모드는 없다(REQ-008, t367 선례 2549f775f). `gen-catalog-hashes --entry plan-auditor`로 카탈로그 해시 갱신. agentemit 골든 테스트에서 plan-auditor 녹색 + sync-auditor 적색(t443 소관) 상태를 증거로 기록.

**M4 (Priority Medium) — 검증 배치 + 증거 커밋**
acceptance.md §D 매트릭스의 명령을 커밋 트리에서 재측정, 결과를 `.moai/specs/SPEC-PLAN-AUDITOR-RESIDUE-001/progress.md` §E.1과 카드 증거에 기록. 증거는 브랜치에 커밋. 최종 판정(verdict.md)은 lane이 `.moai/reports/t450/verdict.md`에 작성 — 이 SPEC 작성자가 아님.

## §G Anti-Patterns

- 금지: `make agents-emit` 전체 재실행으로 t443 드리프트를 몰래 수리 — t443 소관 침범.
- 금지: 기존 쌍둥이 드리프트 2 hunk의 drive-by 수리 — 범위 밖(Out of Scope).
- 금지: .toml 손편집 — C3는 기계 방출물(CLAUDE.local.md §2.0).
- 금지: :72 루브릭 문면 재작성 — t367 착지 내용은 보존 대상.
- 금지: sync-auditor 곁말 반영을 몰래 동봉 — 별도 카드 소관.

## §H Cross-References

- 관련 SPEC: SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 (status: implemented — t367의 루브릭 정비, 이 SPEC의 전제)
- 규약: `.moai/docs/audit-artifact-convention.md` (템플릿 미러 쌍둥이) — § Where / § What / § Side-talk / § What makes the convention stick
- 룰: `.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence (템플릿 미러 쌍둥이)
- 카드 증거 예정 경로: `.moai/reports/t450/verdict.md`
