# plan.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 구현 계획

> Tier M Standard tier. 6 마일스톤 (M1~M6). 본 문서는 plan-phase 산출물이며, 실제 구현은 run-phase manager-develop 위임 (Implementation Kickoff Approval 이후).

---

## §A. Context

본 SPEC은 AskUserQuestion 의사결정을 지속성 있게 만들고, 과거 의사결정 기반 적응형 추천 배치를 도입한다. 5개 깊은 연구 각도(25 학술 논문 + 6 산업 도구)에서 도출된 4 STRONG 원칙(통합/just-in-time/특성-상태 분리/추천+이유+opt-out 번들)을 Standard tier로 구현한다.

**Standard tier 범위**: policy(규칙 + 출력 스타일) + 제한적 PostToolUse 캡처 훅(advisory/fail-open) + 선호 메모리 계층 + 28일 TTL 감쇠. Power-law 동적 학습, 다중 사용자, RL 정책 학습, 외부 동기화는 "complete" tier로 이월(spec.md §E Out of Scope).

**적응형 추천 강도**: 숙련도 기반 자동 분기 — 전문가=약 추천(info-centric), 일반 사용자=강 추천(기본값-like). 사용자 확인 사항.

---

## §B. Known Issues (사전 식별 리스크)

| ID | 이슈 | 영향 | 완화 |
|----|------|------|------|
| KI-1 | PostToolUse stdin JSON의 AskUserQuestion tool_result 페이로드 구조가 CC 버전별 상이 | 캡처 누락 | advisory/fail-open (REQ-ADM-009); 페이로드 스키마 버전 탐지 + 미인식 시 warn-and-skip |
| KI-2 | 선호 메모리 core 계층이 session-start 로드 비용 가중 | 세션 시작 지연 | NFR-ADM-002 core ≤ 4KB 강제; 초과 시 recall로 강등 |
| KI-3 | 다중 세션이 동시 캡처 시 upsert race | 메모리 손상 | 원자적 upsert (파일 락 또는 single-writer); 다중 세션 race는 Pre-Spawn Sync Check 준수 |
| KI-4 | 추론된 숙련도가 초기 세션에서 부정확 | 잘못된 추천 강도 | cold-start 게이트 (REQ-ADM-014); N 미만 관측 시 neutral 강도 |
| KI-5 | stable/transient misclassification이 감쇠를 왜곡 | 핵심 선호 만료 또는 일시 선호 영구 잔존 | REQ-ADM-003 스키마의 scope 필드 + 정정 루프(REQ-ADM-016)로 교정 |
| KI-6 | template-shipped askuser-protocol.md 카피에 내부 SPEC ID 누출 | 16-언어 템플릿 중립성 위반 | design.md §D 분할 매트릭스 + CI guard `internal_content_leak_test.go` |

---

## §C. Pre-flight (run-phase 진입 전 검증)

run-phase manager-develop 위임 전, orchestrator가 확인할 항목:

- [ ] 본 plan.md가 plan-auditor 독립 감사 PASS (편향 방지)
- [ ] Implementation Kickoff Approval — 사용자 명시적 run-phase 진입 승인 (`.claude/rules/moai/core/askuser-protocol.md §19.1`)
- [ ] Pre-Spawn Sync Check — `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` + `moai session list --json --filter-spec=SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` (다중 세션 race 방지)
- [ ] `moai spec lint` 본 SPEC PASS (frontmatter 12 필드 + GEARS + Out of Scope)
- [ ] `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` 존재 확인 (template-neutrality 분할 대상)

---

## §D. Constraints (구현 제약)

1. **advisory/fail-open 강제** — 캡처 훅은 어떤 상황에서도 exit 2 금지. 기존 `sync-phase-quality-gate.sh` (Stop) + `status-transition-ownership.sh` (PostToolUse on SPEC body) + `team-ac-verify.sh` (TaskCompleted) 만이 exit 2 권한 보유.
2. **Recovery-Signal Carve-Out 준수 (SHOULD, doctrine-honest)** — 회복 턴(stopReason이 PTL/max_output_tokens/media_size/compact-failure/compact 회복 신호)에서 캡처 훅 exit 0 + 캡처 미실행 행동 정의 (`runtime-recovery-doctrine.md §4` + AP-RR-006). 단, 탐지 메커니즘(stopReason 파싱)은 현재 advisory 훅 layer에서 불가능하며 future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`로 이연됨 — 본 SPEC은 행동만 정의하고 탐지를 조작하지 않는다.
3. **verification-claim-integrity 준수** — 추천은 관측된 메모리 엔트리에 매핑; 관측되지 않은 추론 주장 금지 (`verification-claim-integrity.md §1.1 surface 3`).
4. **템플릿 중립성** — `internal/template/templates/` 하위 산출물은 16-언어 범용 (내부 SPEC ID/REQ 토큰/DATE/SHA 금지). Go 코드(`internal/hook/`, `internal/cli/`)와 SPEC 디렉터리 자체는 내부 전용 (중립성 제약 없음).
5. **GEARS 준수** — 신규 요구사항은 GEARS 패턴; 레거시 IF/THEN 금지 (lint `LegacyEARSKeyword` 경고/엄격 모드 오류).
6. **기존 exit-2 훅과의 충돌 금지** — 캡처 훅은 PostToolUse를 사용하되, `status-transition-ownership.sh`과 동일 페이로드 경쟁 시 cohabitation 규칙 준수 (`internal/hook/CLAUDE.md` cohabitation contract).

---

## §E. Self-Verification (plan-auditor gate 통과 기준)

- [ ] 모든 요구사항(REQ-ADM-001~018)이 GEARS 패턴 사용 (Ubiquitous / When / While / Where / When event-detected)
- [ ] 레거시 IF/THEN 0건
- [ ] Out of Scope 섹션 존재 — `### Out of Scope —` H3 + `-` 불릿 (OutOfScopeRule lint)
- [ ] frontmatter 12 정준 필드 + era: V3R6 + tier: M
- [ ] 5개 컴포넌트(C1~C5) 각각이 ≥1 REQ에 매핑
- [ ] verification-claim-integrity가 REQ-ADM-018로 명시 준수
- [ ] advisory/fail-open이 REQ-ADM-009로 명시 준수
- [ ] Recovery-Signal Carve-Out이 REQ-ADM-010으로 명시 준수 (SHOULD, doctrine-honest — 탐지 메커니즘 future SPEC 이연 명시)
- [ ] 템플릿 중립성 분할이 design.md §D 매트릭스로 명시

---

## §F. Milestones (M1~M6, Tier M)

> 각 마일스톤은 manager-develop 위임 단위. 순서는 의존성 기반(시간 추정 아님). 각 M마다 template-neutrality 분할 명시.

### M1 — 선호 메모리 계층 + 스키마 (C1, REQ-ADM-001~004)

**범위**:
- `internal/cli/preference/` 신규 패키지 — Entry 스키마(fact, source_citation, valid_time, last_used, scope, domain, confidence), Store 인터페이스(Upsert, Get, Query), 3계층 검색(Core/Recall/Archival)
- `~/.claude/projects/{hash}/memory/user_decisions/` 디렉터리 레이아웃 — `core.yaml` (≤4KB 항상 로드), `recall.jsonl` (최근 N 세션), `archival/` (전체 검색)
- 기존 `feedback_*.md` / `MEMORY.md` 계층과의 분리 네임스페이스 검증
- 단위 테스트 — upsert 멱등성, 3계층 검색 계단식 fallthrough, core 크기 제한 강제

**template-neutrality 분할**:
- INTERNAL-ONLY: `internal/cli/preference/` (Go 코드, 내부 전용)
- INTERNAL-ONLY: `~/.claude/projects/{hash}/memory/user_decisions/` 레이아웃 (사용자 데이터 경로, 템플릿 아님)
- TEMPLATE-SHIPPED: 없음 (M1은 순수 Go 내부)

**완료 기준**: upsert 멱등성 테스트 PASS, 3계층 검색 테스트 PASS, core ≤ 4KB 강제 테스트 PASS.

### M2 — askuser-protocol.md 추천 규칙 업데이트 (C2, REQ-ADM-005~008, 017)

**범위**:
- `.claude/rules/moai/core/askuser-protocol.md` 수정 — 신규 §"추천 배치 원칙" 서브섹션 추가: (a) 발화 시점 = 정보이익 정렬(p≈0.5); (b) 질문 순서 = 정보이익 내림차순; (c) 추천 옵션 = 통계적 다수 합리적 기본값 + cold-start 공개; (d) 전제조건 서술 의무; (e) 적응형 강도(숙련도 기반 자동 분기)
- `.claude/output-styles/moai/moai.md` §8 (또는 신규 §) — 추천 배치 출력 스타일 렌더 규칙
- **template-shipped 카피 분할 작성** — `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` + `internal/template/templates/.claude/output-styles/moai/moai.md` 내용은 template 중립성 기준에 맞춰 분할 작성 (live 카피는 내부 SPEC ID 교차 참조 허용, template 카피는 내부 ID 0건 — design.md §D.1 준거)

**template-neutrality 분할**:
- TEMPLATE-SHIPPED (중립성 제약): `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` — 범용 추천 배치 원칙, "SPEC-V3R6-ASKUSER-DECISION-MEMORY-001" 등 내부 ID 금지
- TEMPLATE-SHIPPED (중립성 제약): `internal/template/templates/.claude/output-styles/moai/moai.md` — 범용 렌더 규칙
- INTERNAL-ONLY: `.claude/rules/moai/core/askuser-protocol.md` (live 카피, template-shipped와 분할 작성 — live는 내부 SPEC ID 교차 참조 허용, template은 내부 ID 0건)
- CI guard: `internal/template/internal_content_leak_test.go` 통과

**완료 기준**: askuser-protocol.md 양쪽 카피(live + template) 분할 작성 완료 (live는 내부 SPEC ID 교차 참조 허용, template은 내부 ID 0건 — design.md §D.1), template-shipped 카피에서 내부 SPEC ID grep 0, plan-auditor 정책 준수 확인.

### M3 — PostToolUse advisory 캡처 훅 (C3, REQ-ADM-009, 010, 018)

**범위**:
- `internal/hook/post_tool.go` 확장 — 기존 `runHarnessObserve` 경로와 별도로 `user_decision_capture` 서브파이프라인 추가. AskUserQuestion tool_result 감지 시 `internal/cli/preference/` Store.Upsert 호출
- advisory/fail-open 강제 — 모든 오류 경로 exit 0 + `.moai/logs/hook-stderr.log` warn
- Recovery-Signal Carve-Out — stdin JSON의 stopReason/recovery 신호 탐지 시 exit 0 + 캡처 미실행
- verification-claim-integrity — 캡처된 엔트리는 `confidence: observed`로 기록; `inferred`는 별도 경로(orchestrator 추론)에서만 upsert
- 단위 테스트 — 오류 경로 전부 exit 0, 회복 턴 exit 0, 정상 캡처 upsert, 동시 캡처 race

**template-neutrality 분할**:
- INTERNAL-ONLY: `internal/hook/post_tool.go` (Go 코드)
- INTERNAL-ONLY: `.claude/hooks/moai/handle-post-tool.sh` (이미 존재하는 wrapper, 확장 불필요 — Go 핸들러에서 라우팅)
- TEMPLATE-SHIPPED: `internal/template/templates/.claude/hooks/moai/handle-post-tool.sh` (기존 wrapper, 변경 없음 — 이미 범용)

**완료 기준**: 모든 오류 경로 exit 0 테스트 PASS, Recovery-Signal Carve-Out 테스트 PASS, 기존 exit-2 훅과 cohabitation 테스트 PASS.

### M4 — 감쇠 정책 + 28일 TTL (C4, REQ-ADM-011, 012)

**범위**:
- `internal/cli/preference/decay.go` 신규 — 멱법칙 감쇠 함수(power-law: `weight = (age_days + 1)^(-α)`, 초기 α=0.5 고정 — "complete" tier에서 동적 학습 이월)
- 일일 백그라운드 감쇠 스캔 (NFR-ADM-004) — `moai preference decay-scan` CLI 서브커맨드 또는 SessionStart 훅에서 1일 1회 실행
- 일시/지속 분리 — stable 엔트리는 pure time-decay 면제, last_used 갱신 시 confidence weight 부양
- 28일 미사용 만료 — transient 엔트리 중 last_used > 28일인 항목 soft-delete (archival로 이동, 즉시 삭제 아님)
- 사용시 리셋 — recall/archival에서 재사용 시 age 카운터 0 + confidence 부양
- 단위 테스트 — stable 면제, transient 28일 만료, 사용시 리셋, 멱법칙 곡선 형태 검증

**template-neutrality 분할**:
- INTERNAL-ONLY: `internal/cli/preference/decay.go` (Go 코드)
- TEMPLATE-SHIPPED: 없음 (M4는 순수 Go 내부)

**완료 기준**: 감쇠 스캔 단위 테스트 PASS, stable/transient 분리 감쇠 테스트 PASS, 28일 TTL 만료 테스트 PASS.

### M5 — 회복 제어 + 맥락적 개인화 게이트 (C5, REQ-ADM-013~016)

**범위**:
- `/moai preference toggle` CLI 서브커맨드 — 세션 단위 개인화 비활성화 (NFR-ADM-005). 상태는 `.moai/state/session-preference-disabled` 센티넬 파일
- 맥락적 강도 저하 게이트 — orchestrator가 민감 도메인(security review, 일회성 쿼리, cold-start) 탐지 시 추천 강도 neutral로 저하 + 공개
- "N일 전 데이터 기반" 공개 — 추천 옵션 description에 신선도 공개 (REQ-ADM-015)
- 정정 루프 — 추론된 선호 공개 시 "이 추론이 틀리면 알려주세요" 채널 제공, 사용자 정정 시 즉시 upsert (REQ-ADM-016)
- 숙련도 추론 — 세션 카운트 + 의사결정 일관성 + 명시적 자가 평가 중 ≥1 방법으로 숙련도 추정 (초기: 세션 카운트 ≥ 임계값 시 전문가)
- 단위 테스트 — 토글 세션 단위 격리, 민감 도메인 강도 저하, 정정 루프 upsert

**template-neutrality 분할**:
- INTERNAL-ONLY: `internal/cli/preference/toggle.go` (Go 코드)
- TEMPLATE-SHIPPED: 없음 (M5 Go 코드는 내부 전용)
- INTERNAL-ONLY: `.moai/state/session-preference-disabled` (상태 파일)

**완료 기준**: 토글 단위 테스트 PASS, 민감 도메인 게이트 테스트 PASS, 정정 루프 테스트 PASS.

### M6 — sync (CHANGELOG + README + frontmatter 전환)

**범위**:
- CHANGELOG.md 엔트리 — 본 SPEC 기능 요약 (사용자 대상)
- README.md / README.ko.md — "AskUserQuestion 의사결정 메모리" 섹션 추가 (사용자 마이그레이션 안내)
- docs-site 4-locale 동기화 (`adk.mo.ai.kr`) — `content/en/workflow-commands/moai-plan.md` + ko/ja/zh 동기화, GEARS 추천 배치 원칙 섹션
- frontmatter `status: draft → planned → in-progress → implemented` 전환 (V3R6 3-phase close: plan→run→sync)
- `progress.md` §E.4 sync_commit_sha 백fill

**template-neutrality 분할**:
- INTERNAL-ONLY: `.moai/specs/SPEC-V3R6-ASKUSER-DECISION-MEMORY-001/` (SPEC 디렉터리 자체)
- TEMPLATE-SHIPPED (중립성 제약): docs-site 콘텐츠는 범용 사용자 대상
- INTERNAL-ONLY: CHANGELOG.md (내부 릴리스 노트)

**완료 기준**: CHANGELOG/README 업데이트, docs-site 4-locale 동기화, `moai spec audit` drift 0, frontmatter `implemented`.

---

## §G. Anti-Patterns (구현 시 금지 패턴)

| AP-ID | 안티 패턴 | 이유 |
|-------|----------|------|
| AP-ADM-001 | 캡처 훅이 exit 2로 워크플로 차단 | advisory/fail-open 위반 (REQ-ADM-009) |
| AP-ADM-002 | 캡처 훅이 회복 턴에서 캡처 시도 (탐지 불가에도 불구하고 proxy 신호 조작으로 회복 턴을 "탐지했다"고 주장) | Recovery-Signal Carve-Out 위반 (REQ-ADM-010 SHOULD), death-spiral 유발 + AP-RR-006 위반 (over-claim) |
| AP-ADM-003 | 추천이 관측되지 않은 메모리 엔트리 기반 | verification-claim-integrity 위반 (REQ-ADM-018) |
| AP-ADM-004 | append-only 메모리 (upsert 아님) | 통합 원칙 위반 (REQ-ADM-001), 토큰 비용 선형 증가 |
| AP-ADM-005 | template-shipped 카피에 내부 SPEC ID/REQ 토큰 포함 | 템플릿 중립성 위반 (NFR-ADM-006) |
| AP-ADM-006 | 순진 time-decay로 stable 선호 만료 | 일시/지속 분리 위반 (REQ-ADM-011), Koren 지속 신호 상실 |
| AP-ADM-007 | 전문가에게 강 추천 배치 | 적응형 강도 위반 (REQ-ADM-017), Loughrey 자율성 침식 |
| AP-ADM-008 | 추천 옵션에 전제조건 서술 누락 | 투명성 번들 위반 (REQ-ADM-008), 기형적 설계 |
| AP-ADM-009 | 정정 채널 없는 추론 공개 | 정정 루프 위반 (REQ-ADM-016), 블랙박스 추론 |
| AP-ADM-010 | 기존 exit-2 훅과의 cohabitation 무시 | sync-phase-quality-gate.sh / status-transition-ownership.sh 충돌 |

---

## §H. Cross-References

- `.claude/rules/moai/core/askuser-protocol.md` — M2 수정 대상 (live + template-shipped)
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md §4` — REQ-ADM-010 Recovery-Signal Carve-Out 준거
- `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3` — REQ-ADM-018 준거
- `.moai/docs/template-internal-isolation-doctrine.md §25` — NFR-ADM-006 템플릿 중립성 준거
- `internal/cli/hook.go runHarnessObserve` — M3 확장 지점
- `internal/hook/CLAUDE.md` cohabitation contract — M3 다중 훅 공존 규칙
- `internal/template/internal_content_leak_test.go` — M2/M6 CI guard
- `SPEC-ASKUSER-ENFORCE-001` (legacy, implemented) — 채널 독점 + Socratic 절차 (본 SPEC은 그 위에 지속성 계층 추가)
- `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` (completed) — runtime-recovery-doctrine SSOT
- `SPEC-EVIDENCE-CLAIM-INVARIANT-001` (completed) — verification-claim-integrity doctrine
- research.md — 5개 연구 각도 + 25 학술 인용
- design.md — 아키텍처 결정 + 템플릿 중립성 분할 매트릭스
