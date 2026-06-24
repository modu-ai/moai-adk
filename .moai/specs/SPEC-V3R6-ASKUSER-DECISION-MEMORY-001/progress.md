# progress.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001

> 본 파일은 plan-phase에서 §E 스켈레톤만 생성. §E.2/§E.3/§E.4 증거 콘텐츠는 run-phase(manager-develop) 및 sync-phase(manager-docs)에서 채운다. 본 에이전트(manager-spec)는 §E.1만 채운다.

---

## §A. 현재 상태

- **Phase**: run-phase 진행 중 (M1·M2 완료, M3 대기) — handoff 시점 2026-06-25
- **Status**: in-progress (frontmatter — M1 commit `6a42cde91`에서 draft → in-progress 전환)
- **plan-auditor 독립 감사**: iter-2 PASS 0.88 (Tier M threshold 0.80)
- **Implementation Kickoff Approval**: 획득 (사용자 "M1부터 순차 진입" 승인, IGGDA explicit-gate)
- **M1 (REQ-ADM-001~004)**: ✅ 완료 — `internal/cli/preference/` 9 파일 (Entry 7필드/Store/3-tier cascade/atomic upsert), 5 AC PASS (AC-ADM-001~004+NFR-002), coverage 85.7%, commit `6a42cde91`+`ffe162709`, orchestrator 독립 검증 7/7 PASS
- **M2 (REQ-ADM-005~008,017 + NFR-006)**: ✅ 완료 — askuser-protocol.md + moai.md live/template 분할 (+77 ins, 0 del), CI guard PASS (본 SPEC ID 0건), moai.md byte-identical parity, commit `c51839d2f`+`693ee464d`, orchestrator 독립 검증 7/7 PASS
- **origin/main**: synced `0 0` (HEAD = `693ee464d`)
- **다음 (M3)**: PostToolUse advisory 캡처 훅 — `internal/hook/post_tool.go` `user_decision_capture` 서브파이프라인 (REQ-ADM-009/010/018, AC-ADM-009/010/018). advisory/fail-open + Recovery-Signal Carve-Out(SHOULD doctrine-honest). 새 세션 sequential 위임.

---

## §B. 산출물

| 파일 | 상태 |
|------|------|
| spec.md | 작성 완료 (plan-phase) |
| plan.md | 작성 완료 (plan-phase) |
| acceptance.md | 작성 완료 (plan-phase) |
| research.md | 작성 완료 (plan-phase) |
| design.md | 작성 완료 (plan-phase) |
| progress.md | 스켈레톤 (본 파일) |

---

## §C. 다음 단계

1. plan-auditor 독립 감사 (편향 방지)
2. Implementation Kickoff Approval (사용자 명시적 run-phase 진입 승인)
3. Pre-Spawn Sync Check (다중 세션 race 방지)
4. run-phase manager-develop 위임 (M1부터 순차)

---

## §D. PRESERVE-list (중단 시 복구용)

_<pending run-phase>_ — run-phase 진입 후 manager-develop이 채움.

---

## §E.1 Plan-phase Audit-Ready Signal

본 SPEC 디렉터리는 plan-phase 산출물 5종(spec/plan/acceptance/research/design) + 본 progress.md 스켈레톤으로 구성된다. SPEC ID 사전 작성 자체 점검 통과(`SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` → 정준正규식 `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` PASS). frontmatter 12 정준 필드 + era: V3R6 + tier: M.

---

## §E.2 Run-phase Evidence

### M1 — preference 메모리 계층 (`internal/cli/preference/`)

**패키지 산출물**:
- `internal/cli/preference/entry.go` — Entry struct (7 필드) + typed Scope/Confidence enums + Validate()
- `internal/cli/preference/store.go` — Store interface (Upsert/Get/Query) + Tier enum + ErrInvalidEntry/ErrNotFound
- `internal/cli/preference/filestore.go` — fileStore 구현: 3-tier cascade (core.yaml / recall.jsonl / archival/), atomic upsert (temp+rename), core ≤4KB 강제 (demote on overflow), namespace 분리
- 테스트: `entry_test.go`, `store_test.go`, `store_helpers_test.go`, `filestore_coverage_test.go`, `error_paths_test.go`, `atomicwrite_test.go`

**테스트 결과** (`go test ./internal/cli/preference/...`):
```
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	0.430s	coverage: 85.7% of statements
```

**레이스 검출** (`go test -race ./internal/cli/preference/...`):
```
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	1.515s
```

**크로스 플랫폼 빌드**:
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint** (`golangci-lint run --timeout=2m ./internal/cli/preference/...`):
- 0 issues (staticcheck SA9003 empty-branch 수정 후 clean)

**Subagent boundary grep** (C-HRA-008):
- `grep -rn 'AskUserQuestion' internal/cli/preference/ | grep -v _test.go | grep -v "// "` → 0 matches

**commit SHA**: `6a42cde91` (M1 preference memory layer — 본 커밋)

### M2 — askuser-protocol.md + moai.md 추천 배치 원칙 (policy-doc authoring)

**범위**: REQ-ADM-005~008, 017 (C2 추천 배치 규칙). 본 M2는 policy-document authoring 마일스톤 (Go 코드 아님). 4개 target file에 신규 "Recommendation Placement Principles" / "AskUserQuestion Recommendation Placement" 섹션 추가:

**산출물 (4 파일)**:
- `.claude/rules/moai/core/askuser-protocol.md` (LIVE) — `## Recommendation Placement Principles` 신규 § 추가 (Option Description Standards 뒤, Preview Field Standards 앞). 5원칙 (발화 시점 Fisher 정보 / 질문 순서 정보이익 내림차순 / 통계적 다수 합리적 기본값 + cold-start 공개 / 전제조건 서술 / 적응형 강도) + AC-ADM-005~008,017 관측 증거 명시 + 내부 SPEC-ID 교차 참조.
- `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (TEMPLATE) — 동일 5원칙 neutral form (내부 SPEC-ID/REQ/AC 토큰 0).
- `.claude/output-styles/moai/moai.md` (LIVE) — `### AskUserQuestion Recommendation Placement` 신규 서브섹션 추가 (Localization Contract Pre-emit self-check 뒤, Task Start 배너 앞). 5원칙 렌더 규칙 + cold-start 공개문/전제조건 서술의 conversation_language 번역 의무.
- `internal/template/templates/.claude/output-styles/moai/moai.md` (TEMPLATE) — 동일 렌더 규칙 neutral form. **moai.md는 LIVE=TEMPLATE byte-identical** (REQ-WF006-003 parity test `TestOutputStylesTemplateLiveParity` PASS — design.md §D.1 cross-ref 허용이 있으나 parity test가 byte-identity를 요구하므로 양쪽 동일 내용으로 작성).

**template-neutrality (AC-ADM-NFR-006 S1 Blocker)**: 신규 추가 섹션에서 `SPEC-V3R6|REQ-ADM|AC-ADM|2026-06-24|[0-9a-f]{40}|.moai/(specs|reports)` grep 0건 (새 섹션만 추출 후 grep — 기존 pre-existing 3 매치는 Preview Field Worked Example 옵션 라벨로 M2 범위外 PRESERVE).

**CI guard**:
- `TestTemplateNoInternalContentLeak` PASS (template neutral)
- `TestOutputStylesTemplateLiveParity` PASS (moai.md live=template byte-identical)
- `TestRuleTemplateMirrorDrift` PASS (askuser-protocol.md는 byte-parity allowlist에 없음 — leak-test coverage 범주; moai.md는 parity allowlist에 있음 — byte-identical)
- `TestLanguageNeutrality` PASS
- `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0

**N/A — policy-document milestone**: 본 M2는 Go 코드가 아닌 policy-doc authoring이므로 E3 (coverage)는 N/A. AC의 "관측 증거" 조항은 policy 문서 내에 원칙이 서술되어 있는지 (grep 가능한 문구 존재)로 검증.

**commit SHA**: `c51839d2f` (M2 askuser recommendation-placement rules — 본 커밋, origin/main synced `0 0`)

---

## §E.3 Run-phase Audit-Ready Signal

### M1 AC PASS 매트릭스

| AC | Status | 검증 명령 | 실제 출력 |
|----|--------|-----------|-----------|
| AC-ADM-001 (upsert 멱등성 — replace not append) | PASS | `go test -run TestUpsert_Idempotent_ReplaceNotAppend ./internal/cli/preference/...` | `--- PASS: TestUpsert_Idempotent_ReplaceNotAppend (0.01s)` |
| AC-ADM-002 (네임스페이스 분리 — feedback vs user_decisions 독립 쿼리) | PASS | `go test -run TestNamespaceSeparation_UserDecisionsVsFeedback ./internal/cli/preference/...` | `--- PASS: TestNamespaceSeparation_UserDecisionsVsFeedback (0.01s)` |
| AC-ADM-003 (7-필드 스키마 + 누락 필드 거부) | PASS | `go test -run TestEntry_SevenFieldsPresent ./internal/cli/preference/... && go test -run TestEntry_Validate_RejectsMissingFields ./internal/cli/preference/...` | 2 PASS (8 subtests — missing fact/source_citation/valid_time/last_used/domain/decision_key + invalid scope/confidence 각각 거부) |
| AC-ADM-004 (3-tier cascade — core hit 시 recall/archival 미접근 + recall hit after core miss) | PASS | `go test -run TestCascade_CoreHitSkipsRecallAndArchival ./internal/cli/preference/... && go test -run TestCascade_RecallHitAfterCoreMiss ./internal/cli/preference/...` | 2 PASS (TierCore / TierRecall 반환 검증) |
| AC-ADM-NFR-002 (core ≤4KB 강제 + 초과 시 recall 강등) | PASS | `go test -run TestCoreSizeEnforcement_DemotesOnOverflow ./internal/cli/preference/...` | `--- PASS: TestCoreSizeEnforcement_DemotesOnOverflow (0.03s)` — 20 entries upsert 후 core.yaml ≤4096 bytes + 강등 엔트리 cascade 회수 |

### M2 AC PASS 매트릭스 (policy-doc authoring — REQ-ADM-005~008, 017, NFR-006)

본 M2는 policy-document authoring 마일스톤이므로 "관측 증거" 조항은 policy 문서 내에 원칙이 서술되어 있는지 (문구 grep)로 검증한다.

| AC | Status | 검증 (policy 문서 내 원칙 서술 위치) | 관측 증거 |
|----|--------|--------------------------------------|-----------|
| AC-ADM-005 (정보이익 정렬 발화 — p≈0.5 발화 / p≈0,1 자동 처리+생략) | PASS | `.claude/rules/moai/core/askuser-protocol.md § Recommendation Placement Principles` 원칙 1 (Fisher 정보 I=p(1−p) p=0.5 최대) + moai.md 렌더 규칙 1 | LIVE + TEMPLATE 양쪽에 "p ≈ 0.5 (Fisher 정보 I=p(1−p) 최대, 결정 경계)이면 ... 발화", "p가 0 또는 1에 가까우면(거의 확정) ... 자동 처리하고 질문을 생략" 서술 존재 |
| AC-ADM-006 (질문 순서 — 정보이익 내림차순) | PASS | askuser-protocol.md 원칙 2 + moai.md 렌더 규칙 2 | "각 질문의 추정 정보이익을 내림차순으로 정렬한다 (가장 높은 정보이익 질문이 첫 번째)" 서술 존재 |
| AC-ADM-007 (통계적 다수 합리적 기본값 + cold-start 공개) [S1 Blocker] | PASS | askuser-protocol.md 원칙 3 + moai.md 렌더 규칙 3 | "(권장) 라벨은 ... 관측된 통계적 다수 합리적 기본값 ... 시스템이 밀고 싶은 정책 기본값이 아니어야 한다" + "based on static default, N observations needed for personalization" 공개 의무 서술 존재 |
| AC-ADM-008 (전제조건 서술) | PASS | askuser-protocol.md 원칙 4 + moai.md 렌더 규칙 4 | "추천 옵션의 description은 추천이 성립하는 전제조건을 서술해야 한다" + "Recommended when <precondition>" 형식 권장 서술 존재 |
| AC-ADM-017 (적응형 추천 강도 — 숙련도 기반 자동 분기) | PASS | askuser-protocol.md 원칙 5 + moai.md 렌더 규칙 5 | "고숙련도(전문가) ... 약 추천 강도(info-centric ... (권장) 라벨 override 없이 inferred preference를 공개만)" / "저숙련도(일반 사용자) ... 강 추천 강도 ... (권장) 라벨 + 투명한 이유" + cold-start neutral 강도 서술 존재 |
| AC-ADM-NFR-006 (template 중립성 — 내부 토큰 0) [S1 Blocker] | PASS | `grep -rE 'SPEC-V3R6|REQ-ADM|AC-ADM|2026-06-24|[0-9a-f]{40}|\.moai/(specs|reports)'` on template copies | 신규 추가 섹션에서 0건 (새 섹션만 awk 추출 후 grep exit=1). 기존 pre-existing 3 매치 (Preview Worked Example 라벨)는 M2 범위外 PRESERVE. `TestTemplateNoInternalContentLeak` PASS |

**atomic upsert (SIGKILL 방어, AC-ADM-001 edge case)**: `TestUpsert_AtomicWriteSurvivesPartialState` — recall.jsonl write 후 임시 파일 잔류 0건 검증 PASS.

**coverage**: 85.7% (≥85% threshold PASS per acceptance.md §D.6 quality gate).

**Gaps (미검증 — verification-claim-integrity §3.4 준수)**:
- 다중 프로세스 동시 캡처 race (design.md §C cohabitation) — 본 패키지는 단일 프로세스 내 sequential 안전만 보장; cross-process 동시성은 M3 PostToolUse 캡처 훅 통합 시 advisory/fail-open (REQ-ADM-009)으로 처리됨. 본 M1 unit test는 cross-process 시나리오를 다루지 않음.
- `internal/hook/post_tool.go` user_decision_capture 서브파이프라인 배선 — M3 scope; 본 M1은 패키지만 제공.

**Residual-risk (잔여 위험 — verification-claim-integrity §3.5)**:
- core.yaml 4KB 강제가 매우 긴 단일 Fact(>4KB)에서 데드락 가능성 — 현재 `upsertToCore`는 `len(cf.Entries)==0` 조건으로 루프 탈출하므로 단일 엔트리가 4KB 초과해도 멈추지 않음 (core.yaml 자체가 4KB 초과 상태로 남음). 실제 선호 Fact는 수백 바이트 수준이므로 실용적 한계 내. 단일-엔트리 캡 도입은 complete-tier 후속 검토 권장.

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_ — manager-docs가 sync_commit_sha + CHANGELOG/README 업데이트 증거로 채움.
