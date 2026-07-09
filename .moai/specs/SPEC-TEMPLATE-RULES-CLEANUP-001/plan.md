# SPEC-TEMPLATE-RULES-CLEANUP-001 — 구현 계획 (plan.md)

## §A Context

62개 템플릿 배포 rule 파일의 5개 finding 그룹(A 깨진 참조 / B 중립성 / C 백포트 / D retired 어휘 / E design 역드리프트)을 정리하고, 재발을 막는 CI 가드 4종을 확장한다. 감사 원천: session 98a5197a (2026-07-09), 본 plan-phase 세션에서 현재 트리 대비 전 anchor 재검증 완료 (research.md).

Tier L / 5-artifact. 요구사항 28건 (REQ-TRC-001..066), AC 27건.

### A.1 파일별 트리-분류 매트릭스 (REQ-TRC-002 — 편집 전 필수 참조)

| 대상 파일 (`.claude/rules/moai/` 기준) | 분류 | 편집 방식 | 근거 |
|---|---|---|---|
| `workflow/spec-workflow.md` | BYTE-PARITY ENROLLED | template + local 동일 편집 (mirror test green 필수) | `rule_template_mirror_test.go` `workflowOptMirroredPaths` |
| `workflow/session-handoff.md` | BYTE-PARITY ENROLLED | template + local 동일 편집 | 상동 |
| `development/manager-develop-prompt-template.md` | SANITIZED PAIR | **template 측만** 편집 (local은 내부 콘텐츠 보유 허용) | mirror test 주석 + `TestSanitizedPairParity` |
| `workflow/runtime-recovery-doctrine.md` | SANITIZED PAIR | **template 측만** 편집 (백포트도 template 방향) | `sanitized_pair_parity_test.go` registry :71-76 |
| `development/agent-authoring.md`, `development/skill-authoring.md` | UNENROLLED (local이 dev-only 문서를 정당 인용) | template 측 prose-rewrite; local은 dev-only 인용 유지 가능 — 파일별 결정을 progress.md에 기록 | local엔 `.moai/docs/dev-only-*` 실존 |
| `core/askuser-protocol.md`, `core/settings-management.md`, `development/sprint-round-naming.md`, `core/agent-common-protocol.md`, `workflow/worktree-integration.md`, `workflow/orchestration-mode-selection.md` | UNENROLLED (기본 동일 편집) | template + local 동일 편집 (드리프트 최소화) | 미래 mirror 등록 용이성 |
| `design/constitution.md` | UNENROLLED → M6에서 byte-parity 복원 | template 편집(M3 날짜/M5 Sprint) → local을 template로 동기화(M6) | Finding E 사용자 결정 |
| `core/zone-registry.md` | TEMPLATE 제거 / LOCAL 유지 | template 파일 삭제 + 참조 정리; local 무접촉 | 사용자 결정 2 |
| `development/spec-frontmatter-schema.md` | UNENROLLED (local에 내부 토큰 존재) | template에 §I row만 백포트 (byte-parity 비목표) | Owning SPEC footer 등 local 내부 콘텐츠 |

### A.2 커밋/푸시 전략 (Hybrid Trunk 보호)

- 커밋은 마일스톤 단위 Conventional Commits, **pathspec 한정** (`git add <명시 경로>` — 공유 체크아웃에 병렬 작업 잔여물 존재).
- M1 가드 커밋은 RED 상태로 로컬 커밋 가능하나, **push는 M7 GREEN 이후 단 1회** — trunk 직push 프로젝트에서 RED 가드를 push하면 main CI가 즉시 적색이 된다.
- RED 증거는 progress.md §E.2에 검증 출력으로 보존 (커밋 순서: M1 가드 커밋이 정리 커밋들보다 선행 → TDD 증거가 히스토리에 남음).

## §B Known Issues / Risks

| # | 리스크 | 완화 |
|---|--------|------|
| R-1 | zone-registry 런타임 소비자 3곳 (`constitution.go:24`, `doctor.go:580`, `spec_lint.go:159`) — 사용자 프로젝트에서 파일 부재 시 동작 | M3에 scratch `moai init` 프로젝트 실측 검증 단계 포함. doctor/spec_lint는 graceful 확인됨; `moai constitution list`가 비-graceful이면 최소 Go 수정 추가 (REQ-TRC-027) |
| R-2 | date 가드 오탐 — `Last Updated:`/`Version:` footer, 정당한 만료일(예: EARS BC window) | design.md §4: 별도 테스트 + 라인-컨텍스트 제외 + 파일별 allowlist. default-tier `leakClasses`에 추가 금지 (`TestLeakClassNoDateShaInDefaultTier` 계약) |
| R-3 | W# 패턴 오탐 (W3C 등) | RE2 호환 협패턴 우선 (`lessons? #[0-9]+`, `W[0-9] (meta|fix)` 류) + allowlist. design.md §3 |
| R-4 | byte-parity enrolled 파일의 단측 편집 → `RULE_TEMPLATE_MIRROR_DRIFT` | A.1 매트릭스 참조 강제, 마일스톤별 검증 커맨드에 mirror test 포함 |
| R-5 | 공유 체크아웃 병렬 세션 레이스 | pathspec-한정 커밋, spawn 전 `git fetch` + divergence 확인 (agent-common-protocol §Pre-Spawn Sync Check) |
| R-6 | `.moai/sprints` 경로 — design.yaml config 키는 라이브 (`loader_design.go`) | REQ-TRC-040은 rules 프로즈만 처리, config/Go 무접촉 |
| R-7 | askuser-protocol worked example의 SPEC ID들은 기존 pedagogical allowlist 등재 — provenance 문구만 제거해야 allowlist drift가 없음 | M3 편집 시 leak test allowlist entry (LineStart/LineEnd diagnostic) 재확인 |
| R-8 | 정리 후 `make build` 미실행 시 embed 바이너리 stale | M7에서 `make build` + 배포 검증 필수 |

## §C Pre-flight (run-phase 진입 전 확인)

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → 레이스 없음 확인
2. `go test ./internal/template/ -count=1` → 기존 스위트 green (baseline 기록)
3. `grep -c '\[HARD\]' <각 편집 대상 template 파일>` → per-file [HARD] baseline 기록 (REQ-TRC-001 검증용)
4. `ls internal/template/templates/.claude/rules/moai/core/zone-registry.md` → 존재 확인 (제거 대상)
5. research.md §R1-R9 정독 (특히 sanitized-pair 방향, date tier-ownership)

## §D Constraints

- Template-First: `internal/template/templates/` 우선 편집 → `make build` → local 반영 (분류 매트릭스 예외 준수)
- Rewrite는 normative 내용 비약화 (REQ-TRC-001, [HARD] count 보존)
- 중립성 클래스 SSOT: `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (C1-C8)
- retired 어휘 SSOT: `sprint-round-naming.md`
- 커밋 pathspec 한정, push는 GREEN 후 1회
- 본 SPEC 산출물 외 파일은 plan-phase에서 무접촉 (run-phase에서만 편집)

## §E Self-Verification (run-phase 종료 게이트)

- E1: acceptance.md §D 전 AC PASS/FAIL 매트릭스 (명령 + verbatim 출력)
- E2: `make build` + `go build ./...` 성공 출력
- E3: `go test ./internal/template/ -count=1` 전체 green 출력
- E4: 신규 가드 4종 RED 출력(정리 전) + GREEN 출력(정리 후) 대비 기록
- E5: mirror/sanitized-pair/date-tier 3계약 테스트 green (`TestRuleTemplateMirrorDrift`, `TestSanitizedPairParity`, `TestLeakClassNoDateShaInDefaultTier`)
- E6: push 상태 (GREEN 후 1회 push의 커밋 SHA 목록)
- E7: progress.md §E.2/§E.3 기록 완료

## §F Milestones

| M | 내용 | 주요 파일 | 검증 명령 | 소유자 |
|---|------|-----------|-----------|--------|
| **M1** | CI 가드 4종 작성 (TDD RED) — (a) REQ/AC rules-스코프 확장 + pedagogical allowlist, (b) lessons/W# 클래스, (c) date-provenance 별도 테스트, (d) CONST/SPEC-V3R rules 스코프 확장 + recurrence backstop | `internal/template/internal_content_leak_test.go`, `internal/template/template_neutrality_audit_test.go` 또는 신규 `internal/template/rule_provenance_audit_test.go` (+2~3 테스트 파일 변경) | 신규 가드 실행 → **FAIL(RED) 출력 기록**; `go test ./internal/template/ -count=1`에서 기존 테스트는 green | manager-develop (cycle_type=**tdd**) |
| **M2** | Finding A: 깨진 참조 5건 prose-rewrite | template: `agent-authoring.md`(:25,:38 + 97/98 표기), `skill-authoring.md`(:357), `spec-workflow.md`(:25,:124 — **local 동일 편집**), `manager-develop-prompt-template.md`(:107 — template만) | `grep -rn 'dev-only-commands-isolation\|git-local-workflow-doctrine\|git-workflow-doctrine\|pattern-library' internal/template/templates/.claude/rules/` → 0; mirror test green | manager-develop (cycle_type=ddd, mechanical) |
| **M3** | Finding B: 중립성 정리 + zone-registry 템플릿 제거 + 참조 정리 + CLI graceful 실측 | template: `askuser-protocol.md`, `settings-management.md`, `sprint-round-naming.md`, `session-handoff.md`(local 동일), `spec-workflow.md`(local 동일), `agent-common-protocol.md`, `manager-develop-prompt-template.md`(template만), `worktree-integration.md`, `design/constitution.md`(날짜), `core/zone-registry.md` **삭제**, `runtime-recovery-doctrine.md`(zone-registry 참조 2곳) | acceptance.md AC-TRC-B1..B8 grep 세트 전부 0-hit; scratch `moai init` 프로젝트에서 `moai constitution list`/`moai doctor` graceful 확인 | manager-develop (cycle_type=ddd; graceful 수정 필요 시 해당 부분만 tdd) |
| **M4** | Finding C: 선별 백포트 | template: `spec-frontmatter-schema.md`(§I row), `runtime-recovery-doctrine.md`(용어 ~8곳; sanitized 요소 미복사) | `grep -c 'Token Accounting'` ≥1; MoAI-표기 잔존 grep 0; `grep -n '\.moai/research/\|^Version:\|^Origin:\|CONST-V3R'` → 0; `TestSanitizedPairParity` green | manager-develop (cycle_type=ddd) |
| **M5** | Finding D: retired 어휘 | template: `design/constitution.md`(Sprint Contract 블록 retired 주석 + `.moai/sprints` 문자열 제거), `orchestration-mode-selection.md`(:190 cohort — local 동일 편집) | `grep -n '\.moai/sprints' <template rules>` → 0; `grep -n 'cohort' <template>/orchestration-mode-selection.md` → 0 | manager-develop (cycle_type=ddd) |
| **M6** | Finding E: local 완결 | local: `design/constitution.md` ← template 동기화(M3+M5 반영본), `.moai/config/sections/design.yaml`에서 `design_docs:` 블록 제거 + key-set 정렬 | `diff` constitution 양측 → exit 0; `grep -n 'design_docs' .moai/config/sections/design.yaml` → 0; `go test ./internal/config/ -run 'Symmetry\|Loader' -count=1` green | manager-develop (cycle_type=ddd) |
| **M7** | 재빌드 + 전체 GREEN + 배포 검증 + push | `make build`; 전체 스위트; scratch 배포 재확인; `moai spec lint` | `make build` exit 0; `go test ./internal/template/ ./internal/config/ -count=1` 전부 PASS (신규 가드 **GREEN**); 단일 push | manager-develop + orchestrator 검증 배치 |

마일스톤 순서 의존성: M1(가드 RED 기준선) → M2-M5(template 정리; M3의 zone-registry 참조 정리는 design/constitution.md 편집을 포함하므로 M5와 같은 파일을 만짐 — M3/M5 편집을 같은 파일에서 순차 적용) → M6(local 동기화는 template 확정 후) → M7.

## §G Anti-Patterns (금지)

- `git add -A` / `git add .` — 공유 체크아웃 병렬 잔여물 오염 (pathspec 한정 필수)
- RED 가드 상태로 push — trunk CI 적색
- sanitized-pair 파일의 local 측 강제 byte-parity 맞춤 — `TestSanitizedPairParity` 계약 위반 방향
- date 패턴을 default-tier `leakClasses`에 추가 — `TestLeakClassNoDateShaInDefaultTier` 즉시 FAIL
- rewrite 시 [HARD] 조항 통삭제 — normative 약화 (REQ-TRC-001)
- zone-registry 참조를 다른 template 파일에 새로 만드는 재서술 — REQ-TRC-025 위반
- design.yaml 파일 삭제 — 라이브 config (REQ-TRC-051은 블록 제거+정렬)
- 한국어 문장 blind 치환 — Out of Scope 침범

## §H Cross-References

- spec.md §B (REQ-TRC-*), acceptance.md §D (AC-TRC-*), research.md (증거 R1-R9), design.md (가드 설계)
- `internal/template/rule_template_mirror_test.go` — byte-parity allowlist SSOT
- `internal/template/sanitized_pair_parity_test.go` — sanitized-pair registry
- `internal/template/internal_content_leak_test.go` — leakClasses/pedagogicalAllowlist 기존 구조
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1/§25.3 — 콘텐츠 클래스 + pre-commit self-check
- CLAUDE.local.md §2 (Template-First), §23 (Hybrid Trunk), §25 (내부 콘텐츠 격리)
