# plan.md — SPEC-DOCS-DELEGATION-CWD-001

## §A Context

카드 t1387. 레인(팩토리/칸반 companion) 플로우에서 단계 전문가(manager-spec/develop/docs)의 작업 트리 부착이 런타임 결정이라 예측 불가능하고(격리 / 비격리 양상태 — t1373 실측), 싱크 단계의 manager-docs 위임이 t1383 에서 구조적으로 2회 실패해 레인이 소유 예외로 전환한 결함의 문서+가드 수리. 본 SPEC 은 요구·근거·계약 표면을 `spec.md` 가 소유하고, 이 파일은 구현 절차를 소유한다.

개발 모드: `quality.yaml` constitution.development_mode 따름(기본 TDD). 본 SPEC 의 산출물은 문서 2파일(×2 트리) + Go 콘텐츠 가드 테스트 1건 — 테스트가 RED-first 인 유일한 코드 표면이다.

## §B Known Issues (measured)

- 격리 트리거: 에이전트 정의 3종 모두 `isolation:` 미선언(직접 판독), 런타임 자동 격리의 발동 조건 문서 0건(grep 실측) — spec.md §A.4.
- 좌초 증거: gitignore 된 `.moai/reports/` 는 병합에 동행하지 않음 — t1373 progress:137.
- 무음 재앵커링: 비격리 서브에이전트는 호출마다 세션 앵커로 cwd 재해석 — t741 실측, worktree-integration-ops.md §A background 절.
- 소유 예외의 기록은 존재(t1383 §E.4)하나 트리거 조건·절차는 미성문화.

## §C Pre-flight

1. `moai spec lint SPEC-DOCS-DELEGATION-CWD-001` — 0 error 확인.
2. 편집 대상 4파일 존재 확인: `.claude/rules/moai/workflow/kanban-dispatch-mechanics.md`, `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`.
3. 가드 테스트 착지 위치: `internal/template/` — 콘텐츠 스캔 선례(`workflow_rule_paths_pinned_test.go`, `contract_mode_guided_test.go`)와 동일 패키지.

## §D Constraints

- **미러 쌍 편집 [HARD]**: 룰 편집은 C1(로컬)+C2(템플릿) 쌍으로 내용 반영 후 `make build`. 쌍별 실제 기제(계획 단계 프로브로 확정): 템플릿 사본 중립성은 `TestTemplateNoInternalContentLeak`(템플릿 루트 전역 WalkDir), `agent-common-protocol.md` 쌍의 §25 구조 패리티는 `TestSanitizedPairParity`(`sanitized_pair_parity_test.go` 등록). `TestRuleTemplateMirrorDrift` 는 이 두 쌍에 도달하지 않는다 — mechanics 는 어느 바이트 패리티 목록에도 미등록이고 agent-common-protocol 은 §25 정화 쌍으로 바이트 패리티 설계상 제외. 룰 본문에 카드 내력(t1387)·SPEC-ID 직접 인용 금지(§2.1 중립성 — 룰은 "레인 플로우" 일반론으로 서술).
- **항상 적재 예산**: `agent-common-protocol.md` 추가는 1,000 바이트 이내(권장 ≤ 600B, 문장 2-3개). 초과 시 커밋 본문에 비용 산언. 절차 본문은 paths 스코프인 `kanban-dispatch-mechanics.md` 로 부담 0.
- **가드 테스트 형태**: 섹션 존재 + 필수 단계 마커 키워드 검사(금지형 부정 grep 은 양성 대조 포함 — §6 교훈: 0적중 exit grep 은 양성 대조 없이 부재의 증거가 아니다). 스플릿/다이어트로 절차가 옮겨질 경우를 대비해 마커는 파일 경로 의존을 최소화(두 파일 경로만 하드코딩).
- **커밋 규율**: 카드 워크트리에서 pathspec 명시 스테이징, 커밋 직전 `git rev-parse --short HEAD` + `git branch --show-current` 재판독, Conventional Commits + 카드 id 기재. 전체 스위트 금지 — 변경 패키지(`./internal/template/...`)만.

## §E Self-Verification (run-phase 예고)

- E1: `go test ./internal/template/ -run '^TestReconciliationProcedureDocumented$'` GREEN 출력 인용 + 변이 프로브(절차 소절 임시 삭제 → 가드 RED → 복원 → GREEN) 1회 관측. 미러 안전 계측: `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$|^TestSanitizedPairParity$' -count=1` — 착지 쌍의 실제 기제(§D 참조: 원 인용 미러 테스트는 이 쌍에 도달하지 않음이 프로브로 증명됨).
- E2: `moai spec lint` 0 error.
- E3: 룰 미러 `git status --short` — 편집 4파일만.
- E4: 커밋 그래프 — 가드 테스트 커밋이 문서 커밋 뒤에 오고(RED-first 순서), `verification-claim-integrity.md` §2.3 에 따라 순서 주장은 커밋 그래프가 목격.

## §F Milestones (결정 가역성 순 — 바뀔 가능성이 큰 결정부터)

### M1 (Priority High) — 화해 절차 본문 + 항상 적재 의무 문장

결정 밀도가 가장 높은 표면. 내용 결정(절차 단계 구성·예외 트리거 문구)이 여기서 다 확정된다.

1. `kanban-dispatch-mechanics.md` § Isolation 아래 신설 소절 — 제목: `### Reconciling an isolated specialist spawn`. 필수 단계 5개를 소제목/굵은 마커로 고정: (1) 착지 검증 (2) WT- 개명 (3) `git merge --ff-only` 레인 트리 안 plain git (4) gitignored 증거 수확 + hoist-before-dispose (5) run→sync 인접성·단일 싱크 커밋 무수정 병합. 예외 트리거(2회 구조 실패 + 거부 증거)와 기록 의무(§E.4 + 완료 보고 + 리드 보고), 비격리 중 트리 이동 금지 포함. 룰 본문은 카드 내력 없는 일반론.
2. `agent-common-protocol.md` § User Interaction Boundary 의 lane-session 단락에 문장 추가: 스폰 부착은 런타임 결정 → 스폰 뒤 착지 검증 → 격리면 화해 절차(포인터), 산출물 편집 권한 없음. ≤ 600 바이트.
3. C2 템플릿 미러 동일 반영.

### M2 (Priority High) — 콘텐츠 가드 테스트 (RED-first)

`internal/template/docs_delegation_lane_flow_test.go` — `TestReconciliationProcedureDocumented`: 두 파일에서 소절 제목 + 5개 필수 단계 마커 + 의무 문장 키워드 검사. M1 의 문서보다 **먼저 커밋하지 않는다** — RED-first 는 "테스트가 새 문서를 요구한다"는 순서로 커밋 그래프에 새겨진다(M1 커밋 → M2 커밋 순서 유지, §E4).

### M3 (Priority Medium) — 빌드·린트·검증 마무리 (기계적 꼬리)

`make build` (임베드 재생성) → `go vet ./internal/template/...` → 변경 패키지 테스트 → spec lint. 진행 기록 갱신은 매 마일스톤 즉시.

## §G Anti-Patterns

- 레인이 spec/plan/acceptance 본문을 "고쳐 병합"하는 화해 — REQ-DSC-005 위반이자 tk8hce 재현.
- 격리 에이전트에게 `git -C <레인 트리>` 지시 — 가드 7종 거부 형상의 재생 산업.
- 증거 수확 없이 `moai worktree sweep` — gitignored 근거 소멸.
- 문서만 추가하고 가드 없이 종료 — 룰 다이어트에서 조용히 증발(REQ-DSC-011 이 존재하는 이유).
- t1387 카드 내력을 템플릿 미러에 기재 — §2.1 중립성 위반.

## §H Cross-references

- `spec.md` §A (실측 원천), §D (제약), `acceptance.md` §D (AC 매트릭스), `decision-index.md` (Q1-Q3)
- 룰: `kanban-dispatch.md` § Factory Mode (lane spawn authority), `worktree-integration.md` § Hoist·sweep, `worktree-integration-ops.md` § Refused Commands·background subagent, `spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- SPEC: SPEC-CODEX-GATE-SCOPE-001(§E.4 예외 기록 — 계기), SPEC-STALE-RUN-LABEL-001(양상태 실측 — 원천), SPEC-WORKTREE-SWEEP-001(처분 정책 — 인용)
