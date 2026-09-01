# SPEC-CODEX-REVIEW-TARGET-001 — 진행 기록

카드 **t399** · 이슈 **modu-ai/moai-adk#1632** · 브랜치 `WT-codex-native-branch`

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M 산출물 집합 + progress).
- Tier: **M**. 근거는 AC 예산 — AC 는 iter1 수리 후 11건(001·002·003·004·005·006·006b·007·008·009·010)으로 Tier S 상한(8)을 초과한다. REQ 는 8건으로 Tier M 상한(16) 안. LOC·파일 수만 보면 S 로도 읽히므로 이 분류는 예산 기준의 판단이다.
- 측정 원천: `.moai/reports/t399/discovery.md`, `.moai/reports/t399/schema/v2/ReviewStartParams.json`, 트리 `442da4f06`.
- 카드 전제 정정 1건: "`coerceCodexReviewTarget` 을 덮는 테스트가 없다" → 행동 기준으로는 거짓. 리프트는 이미 검사되고 있으며 검사 지점이 하필 리프트가 옳은 유일한 variant 다(spec.md §A.4).
- 축 2(백엔드 참여 노출)는 t284 소관으로 확정 — 축의 존재·처분은 `.moai/reports/t229/succession.md`, 카드 번호는 큐 판독(spec.md §E).

### plan-audit iter1 → iter2 수리 기록

- iter1 판정: **FAIL 0.75** (Tier M 임계 0.80). must-pass 7/7 통과. 보고서 `.moai/reports/t399/plan-audit.md`.
- blocking 6건(D1~D6) + optional 4건(D7~D10) **전부 반영**. 각 결함의 처리는 spec.md HISTORY v0.2.0 에 항목별로 기록.
- 감사자가 지적한 중심 결함은 "자기가 세운 재사용 규율을 자기 검증 표면에는 적용하지 않았다" — AC-CRT-010(라이브 왕복)이 그 자리를 닫는다.
- 감사자의 3대 실측 주장은 이 세션이 **독립적으로 재확인**했다: 라이브 `review/start` 선례 존재(`codex_live_protocol_probe_test.go:507`), skip 3조건 확립본(`codex_review_gate_live_test.go:33`), `toolErr` 의 `IsError: true` 조기 반환(`mcp_server.go:868`). D2 의 트리 우연 일치(`worktree_base_branch: develop` ↔ `origin/HEAD → origin/develop`)도 직접 측정.
- **운영자 결정 대기 1건**: plan.md §B 의 (가)/(나) — M1 에서 고정. 권장은 (가).
- **운영자 검토 권유 1건**: D2 를 **결정 (a) 정렬**로 닫았다(spec.md §A.7). `worktree_base_branch` 를 기준으로 삼아야 한다면 그것은 GLM 경로까지 함께 바꾸는 별도 카드다.

### 이 세션이 관측하지 않은 것

- 어떤 테스트도 실행하지 않았다. AC-CRT-003 의 "변경 전 초록" 전제는 미확인이며 M1 에서 확인한다.
- 라이브 codex 왕복을 시도하지 않았다. AC-CRT-010 의 RED 예측은 스키마 판독 기반 추론이며, 관측은 M2b 의 몫이다.
- 감사 보고서의 점수·차원 판정은 재계산하지 않았다. 재확인한 것은 그 판정이 근거로 든 파일 사실들이다.

_run-phase 진입 전. 아래 §E.2 ~ §E.4 는 자리표시._

## §E.2 Run-phase Evidence

트리 `.claude/worktrees/t399`, 브랜치 `WT-codex-native-branch`, 기준 `e7746e95d` (= `origin/develop`, fast-forward 흡수 후 `0 0`). 아래 모든 측정은 이 트리에서 이번 실행으로 얻은 것이다.

### 카드 전제의 처분 — 예측이 관측이 됐다

plan-audit iter2 의 N1 이 지적한 대로, AC-CRT-010 의 RED 는 이 시점까지 **스키마 판독에서 유도한 예측**이었고 아무도 라이브 왕복을 돌린 적이 없었다. M2b 가 그것을 돌렸다.

```
--> {"id":3,"jsonrpc":"2.0","method":"review/start","params":{"target":{"type":"baseBranch"},"threadId":"01a05942-1ef4-73e1-b753-0ccfefcd829b"}}
<-- {"error":{"code":-32600,"message":"Invalid request: missing field `branch`"},"id":3}
```

codex-cli **0.150.1**, 실 바이너리 `/Users/goos/.local/bin/codex`. 예측이 관측으로 바뀌었고 카드의 전제는 **반증되지 않았다** — 뒤집혔다면 그 사실을 보고했을 것이다. 이 거절이 이 SPEC 이 확보할 수 있는 가장 강한 RED 이며, 계약 판독이 아니라 실물이다.

### RED 증거 경로 (AC-CRT-008)

| 파일 | 내용 |
|---|---|
| `.moai/reports/t399/red/contract-layer-red.txt` | 계약 층 `-v` 실행. `=== RUN` 17행(셀렉터 0매칭 아님), RED 지정 6 AC 전부 `--- FAIL`, AC-CRT-003 만 `--- PASS` |
| `.moai/reports/t399/red/live-roundtrip-red.txt` | 라이브 왕복 `-v` 실행 — `AC-CRT-010 RED` 로그 + JSON-RPC 거절 본문 |
| `.moai/reports/t399/red/live-roundtrip-transcript.ndjson` | 그 왕복의 양방향 NDJSON 전문 (31행). 6행 = 보낸 요청, 31행 = 거절 응답 |

GREEN 대응본: `.moai/reports/t399/green/contract-layer-green.txt`, `.moai/reports/t399/green/live-roundtrip-green.txt`, `.moai/reports/t399/live/basebranch-roundtrip-green.ndjson`(6행이 `{"branch":"main","type":"baseBranch"}` 로 바뀌었고 id=3 에 error 반환이 없다).

RED 는 프로덕션 변경 **이전** 트리에서 신규 검사만 추가해 얻었다. 신규 검사는 전부 기존 API 표면(`handleCodexAudit` → `sess.sent`)만 쓰므로 컴파일 실패가 아니라 실제 실행 실패로 관측된다 — 컴파일이 깨지면 `=== RUN` 이 하나도 안 나와 §C 규율을 만족하지 못한다.

### §B 후보 결정 (M1)

**(가) 원인을 명명한 `inconclusive`** 로 고정했다. 근거는 plan.md §B 의 권장 그대로다: 호출자가 넘긴 `target=baseBranch` 는 적법한 enum 값이라 고칠 입력이 없고, (나)는 `applyGateUnmet` 을 조기 반환으로 지나쳐 required 게이트 소비자의 관측 표면을 바꾼다.

**이 결정은 운영자 확인을 거치지 않았다.** run-phase 는 사용자에게 질문할 수 없으므로(subagent 경계), plan.md 가 (가)를 권장하고 (나)를 "운영자가 지시하면"의 조건부로 적은 것을 기본값으로 읽고 진행했다. 뒤집는 비용은 작다 — AC-CRT-004 의 표에서 (나) 행으로 갈아끼우고 `runTurn` 의 반환 한 줄을 `toolErr` 로 바꾸는 편집이다.

### 착지한 변경

| 파일 | 변경 |
|---|---|
| `internal/cli/mcp_review_material.go` | `resolveReviewBaseBranchName` + `reviewRefResolves` 신설. `resolveReviewMergeBase` 의 사슬을 이름 층위에서 읽고, 반환 전에 그 이름이 ref 로 해석되는지 확인한다 |
| `internal/cli/mcp_codex.go` | `coerceCodexReviewTarget(v, root) (map, error)` — variant 인지형. `buildCodexReviewParams` 가 error 를 전파하고, `runTurn` 이 조립 실패 시 **아무것도 보내지 않고** 원인을 명명한 `inconclusive` 를 반환 |
| `internal/cli/mcp_server.go` | `codex_audit` 의 `target` 설명에 서버 해석 사실 + 해석 원천 기술 (AC-CRT-009) |
| `internal/cli/mcp_codex_test.go` | `TestCodexAudit_AdversarialDispatchesTurnStart` 가 "adversarial 은 target 을 싣지 않는다"를 단언 (AC-CRT-007) |
| `internal/cli/codex_review_target_test.go` (신규) | AC-CRT-001·002·003·005·006·006b·009 |
| `internal/cli/codex_review_target_live_test.go` (신규) | AC-CRT-010 |

`git_strategy.worktree_base_branch` 는 어느 경로에서도 읽히지 않는다(spec.md §A.7 정렬 결정).

### AC PASS/FAIL 매트릭스

| AC | 판정 | 검증 명령 | 관측 출력 |
|---|---|---|---|
| AC-CRT-001 | **PASS** | `go test ./internal/cli/ -run TestCodexAudit_NativeBaseBranchCarriesBranch -count=1 -v` | `--- PASS ... (0.55s)`. 직렬화된 `params.target` = `{"branch":"main","type":"baseBranch"}` |
| AC-CRT-002 | **PASS** | `go test ./internal/cli/ -run TestCodexAudit_BaseBranchResolutionChain -count=1 -v` | `--- PASS ... (2.14s)`, 4개 하위 케이스 전부 PASS: `step1_remote_default_head` / `step2_main_when_remote_head_absent` / `worktree_base_branch_is_not_read` / `dangling_remote_head_falls_through` |
| AC-CRT-003 | **PASS** (회귀선, 변경 전에도 초록) | `go test ./internal/cli/ -run TestCodexAudit_UncommittedChangesShapeUnchanged -count=1 -v` | 변경 전 `--- PASS (0.44s)` (red/contract-layer-red.txt), 변경 후 `--- PASS (0.45s)`. target 키 1개(`type`), branch/sha/instructions 부재 |
| AC-CRT-004 | **PASS** | `go test ./internal/cli/ -run TestCodexAudit_UnresolvableBaseBranchIsNotASilentOtherReview -count=1 -v` | `--- PASS ... (0.45s)`. review/start 미전송 + `uncommittedChanges` 대체 부재 + `verdict=inconclusive` + `summary` 가 base 해석 불가를 명명하고 `"codex binary not found in PATH"` 와 다름 |
| AC-CRT-005 | **PASS** | `go test ./internal/cli/ -run TestCodexAudit_IncompleteVariantsAreNotSerialized -count=1 -v` | `--- PASS ... (0.45s)`, 하위 `commit` / `custom` PASS |
| AC-CRT-006 | **PASS** | `go test ./internal/cli/ -run TestCodexReviewTarget_SerializableVariantsSatisfyRequiredSet -count=1 -v` | `--- PASS ... (0.54s)`. 순회 행 2 (`uncommittedChanges`, `baseBranch`), 0매칭 가드 통과 |
| AC-CRT-006b | **PASS** | `go test ./internal/cli/ -run TestCodexReviewTarget_UnserializableVariantsLeaveNoTarget -count=1 -v` | `--- PASS ... (0.55s)`. 순회 행 2 (`commit`, `custom`), 두 절(부재 + 대체 부재) 모두 |
| AC-CRT-007 | **PASS** | `go test ./internal/cli/ -run TestCodexAudit_AdversarialDispatchesTurnStart -count=1 -v` | `--- PASS ... (0.00s)`. 저장소 안에서 `baseBranch` 를 든 유일한 codex 검사가 이제 "그 값이 전송되지 않는다"를 단언한다 |
| AC-CRT-008 | **PASS** | 위 RED 증거 경로 표 | `grep -c '=== RUN' red/contract-layer-red.txt` → `17`; RED 6건 전부 `--- FAIL` 관측 |
| AC-CRT-009 | **PASS** | `go test ./internal/cli/ -run TestCodexAuditToolSurface_DescribesServerSideBranchResolution -count=1 -v` | `--- PASS ... (0.00s)`. tools/list 로 읽은 실제 설명에 `server-side` + `remote default head` + `main` |
| AC-CRT-010 | **PASS (관측됨)** | `go test ./internal/cli/ -run TestCodexLive_ReviewStartBaseBranchIsNotRejected -count=1 -v -timeout 300s` | `AC-CRT-010 OBSERVED: live codex accepted the baseBranch review/start; turn.id="01a05943-943f-7413-8b9b-2e95d0799e77"` · `--- PASS (2.16s)`. **skip 아님** — codex-cli 0.150.1 실 바이너리에서 관측 |

### 판정에 쓰지 않은 것

- 스텁의 반환 verdict. 스텁은 요청과 무관한 스크립트를 되돌려주므로(spec.md §A.5) 모든 계약 층 단언은 `sess.sent` 의 직렬화 바이트만 본다.
- `inconclusive` 의 부재. AC-CRT-010 의 판정은 "거절 부재 + `turn/started` 도달"이라는 양성 사실이다.

### 미관측 / 잔여

- codex-cli **0.149.0**(제보자 판번호)에서의 동작. 이 머신에 0.150.1 만 있다.
- codex 가 `baseBranch` 값을 어떻게 해석하는지(로컬 브랜치인지 임의 revision 인지). 픽스처는 `main` 을 로컬 브랜치와 `origin/main` 양쪽으로 만들어 이름 해석 축을 의도적으로 제거했다 — 이 SPEC 이 재는 것은 요청 형태이지 codex 의 이름 규칙이 아니다.
- 리뷰 결과의 품질. 선례와 같이 `turn/started` 에서 세션을 끊는다.
- `worktree_base_branch` 와 `origin/HEAD` 가 갈리는 **실제 사용자 트리**에서의 동작. 픽스처에서는 갈라 놓고 쟀으나(`worktree_base_branch_is_not_read`), 실 트리 관측은 아니다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: 83e50141a  # 선행 커밋 37caf4343 = plan-phase 산출물 착지
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "0 0 (e7746e95d == origin/develop at run entry)"
l44_post_push_fetch: not-applicable (push는 리드가 지정하는 통합 창의 몫)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_vet: "go vet ./internal/cli/... → exit 0"
  windows_vet: "GOOS=windows go vet ./internal/cli/... → exit 0"
package_tests: "go test ./internal/cli/... -count=1 → ok (17 packages, internal/cli 340.9s)"
lint: "golangci-lint run --timeout=5m ./internal/cli/... → 0 issues."
scope_check: "git diff --stat → 4 files; internal/cli/mcp_convergence.go 부재 (DoD 5)"
total_run_phase_files: 6  # 변경 4 + 신규 테스트 2
m1_to_mN_commit_strategy: "M1~M5 를 RED 커밋 1개 + 구현 커밋 1개로 압축 — RED 증거가 커밋 경계로 남는 것이 §C 규율의 요점이므로 그 경계만 보존한다"
operator_decision_pending: "plan.md §B (가)/(나) — (가)로 진행. 운영자 확인 미이행 (subagent 는 질문 채널이 없다). 뒤집기 비용 작음"
```

## §E.4 Sync-phase Audit-Ready Signal

### 착지한 sync-phase 변경

| 파일 | 변경 |
|---|---|
| `CHANGELOG.md` | `[Unreleased] → Fixed` 최상단 항목. 이슈 #1632 링크, 라이브 왕복 관측(예측→관측 전환), variant 인지형 조립, 서버 해석, §A.7 정렬 결정, 미해결 축(t284 · 정책 질문) 명시 |
| `.claude/skills/moai-ref-cross-model-audit/SKILL.md` | `target` 행의 "Passed through unchanged" 가 codex 쪽 서버 해석을 가림 — 문자열이 두 백엔드에 그대로 닿되 codex 는 그 뒤 브랜치 이름을 서버에서 해석한다는 절 추가 |
| `internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md` | 위의 템플릿 미러 (Template-First) |
| `internal/template/catalog.yaml` | `make build` 재생성 — 위 스킬 1행 hash 만 변경 |
| `.moai/reports/t399/issue-reply.md` | #1632 회신 초안 (신규, 미게시 — 게시는 리드 몫). DoD 6 의 3상태 구별 + 제보 5항목 재측정 상태 표 포함 |
| `spec.md` frontmatter | `status: in-progress → completed`, `updated: 2026-09-01` |
| `progress.md` | 이 §E.4 |

### 문서 스윕 결과 (근거 포함)

- **docs-site · README: 해당 없음.** `grep -rn 'baseBranch\|uncommittedChanges' README.md README.ko.md README.ja.md README.zh.md docs-site/content/` → **0 히트**. 4개 README 와 8개 docs-site 페이지가 `codex_audit` 를 언급하지만 전부 도구 카탈로그 행 / `project_root` 절 / fail-open 절이며, `target` 파라미터의 값이나 그 의미를 서술하는 문장은 어디에도 없다. 따라서 4-locale 의무가 발동할 문서가 없고 **아무것도 고치지 않았다** — 이것이 스윕의 결과이지 스윕을 건너뛴 것이 아니다.
- **스킬 표면: 2건 중 1건 변경.** `.claude/skills/moai/workflows/review.md:179` 은 리뷰 범위 → `target` 값 매핑(`--branch B` → `baseBranch`)이며 이 변경에 영향받지 않는다(값 자체는 그대로). `moai-ref-cross-model-audit/SKILL.md:64` 만 갱신했다.
- **범위 밖으로 남긴 관측 1건**: `docs-site/content/{en,ko,ja,zh}/guides/mcp-server.md` 가 `project_root` 를 받는 도구를 **6개**로 적는데, `.claude/rules/moai/core/moai-mcp-tools.md` 는 **9개**로 적는다. 이 카드와 무관한 별개 드리프트이므로 손대지 않았다. 후속 카드 재료.

### 템플릿 미러 검사 결과

- run-phase 변경 4파일은 전부 `internal/cli/` — `find internal/template/templates -name 'mcp_*'` → **0 히트**. 템플릿은 Go 소스를 미러하지 않으므로 **미러 대상 없음**.
- sync-phase 가 만든 미러 의무는 스킬 1건뿐이며 위 표대로 이행 + `make build` 로 `catalog.yaml` 재생성.

### B12 CHANGELOG 방출 자가검사

| 검사 | 명령 | 관측 |
|---|---|---|
| 중복 방출 | `grep -c 'SPEC-CODEX-REVIEW-TARGET-001' CHANGELOG.md` | `0` (방출 전) — 중복 없음 |
| AC 수 일치 | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `10`. **0 아님(공허 아님).** 정본은 **11** — 정규식이 `AC-CRT-006b` 의 후행 문자를 못 잡아 `006` 으로 접기 때문이며, `grep -c 'AC-CRT-006b'` → `2` 로 그 항목의 실재를 별도 확인했다. §E.3 의 `ac_pass_count: 11` 과 CHANGELOG 본문의 "Eleven acceptance criteria" 가 일치 |
| 파일 경로 실재 | `ls` (6경로) | CHANGELOG 가 인용하는 `internal/cli/mcp_codex.go` · `mcp_review_material.go` · `mcp_server.go` · `.moai/reports/t399/red/live-roundtrip-red.txt` · 신규 테스트 2본 전부 실재 |

### 이 세션이 관측하지 않은 것

- 테스트를 재실행하지 않았다. §E.3 의 빌드·vet·테스트·lint 수치와 라이브 왕복 관측은 **run-phase 및 리드의 독립 검증이 이번 실행에서 얻은 것**을 인용한 것이며 sync-phase 의 재측정이 아니다. 이 세션의 변경은 CHANGELOG·스킬 문서·SPEC 산출물뿐이라 Go 동작을 건드리지 않는다(단, `make build` 는 exit 0 으로 컴파일을 관측했다).
- CI 판정. push 하지 않았으므로 이 커밋에 대한 CI 는 존재하지 않는다.
- #1632 회신은 **작성만 했고 게시하지 않았다.**

```yaml
sync_complete_at: 2026-09-01
sync_commit_sha: 390fb7f84  # backfilled in a follow-up commit (a commit cannot name its own hash)
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-CODEX-REVIEW-TARGET-001' CHANGELOG.md → 0 (pre-emission); no duplicate"
b12_self_test_b: "AC regex count 10 (non-zero); canonical 11 — regex folds AC-CRT-006b into 006, verified separately via grep -c 'AC-CRT-006b' → 2. Matches §E.3 ac_pass_count: 11"
b12_self_test_c: "ls over all 6 CHANGELOG-cited paths → all exist"
changelog_entry_position: "[Unreleased] → ### Fixed, first item"
docs_sweep: "docs-site + README: 0 hits for baseBranch/uncommittedChanges → nothing stale, nothing changed. Skills: 1 of 2 updated (moai-ref-cross-model-audit); review.md unaffected"
template_mirror: "run-phase files are all internal/cli/ → no template mirror exists (find → 0 hits). sync-phase skill edit mirrored + catalog.yaml regenerated via make build"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (updated: 2026-09-01)"
  plan_md: "no status/updated fields in frontmatter — nothing to transition (id/title/version/created only)"
  acceptance_md: "no status/updated fields in frontmatter — nothing to transition (id/title/version/created only)"
  progress_md: "no frontmatter block — nothing to transition"
mx_tag_validation: "no @MX annotation added or changed; sync-phase touched no Go source"
canary_compliance_check: not-applicable
push_state: "not pushed — integration is a lead-granted window (per dispatch constraints)"
issue_reply: ".moai/reports/t399/issue-reply.md (drafted, NOT posted)"
```

