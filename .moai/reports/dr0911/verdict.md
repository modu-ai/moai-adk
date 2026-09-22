# dr0911 — develop 회귀 수리: workflow audit F01~F40 병합이 만든 로컬↔템플릿 드리프트

- 카드: dr0911 (임시 번호, Class B)
- 워크트리: `.claude/worktrees/dr0911` · 브랜치 `WT-develop-mirror-fix`
- 기준 트리: 로컬 develop = origin/develop = `ee99507fb` (재fetch 확인)
- 대조 트리(병합 전): `4c99d973e` — `git archive 4c99d973e`를 세션 스크래치에 풀어 `go -C <dir> test`로 측정
- 적용 규칙: `verification-claim-integrity.md` §2(baseline 귀속), `verification-completeness.md` §4(트리 SHA 고정)

## 1. 결론

리드가 짚은 CI 새 실패 8건은 모두 **F 병합이 로컬 사본만 고치고 템플릿 사본을 한 번도 건드리지 않아서** 생긴 로컬↔템플릿 드리프트다. 병합 범위 `4c99d973e..ee99507fb`의 F 커밋 중 아래 7개 쌍의 템플릿 쪽을 수정한 커밋은 0건이다.

- 수리: 4쌍(로컬이 의도된 새 내용 → 템플릿을 로컬에 맞춤, 되돌리기 없음)
- 보류: 3쌍(템플릿에 복사하면 배포판이 존재하지 않는 파일·설정 키를 가리키거나, 이 저장소의 사설 git-flow 모델이 배포판에 실림 → 판정이 애매해 수정하지 않고 보고)

## 2. 병합 전·후 대조 (RED 원인 증거)

같은 선택자로 두 트리를 쟀다. 병합 전 트리에서 전부 `ok`, 병합 후 트리에서 전부 `FAIL`.

| 테스트 | 병합 전 `4c99d973e` | 병합 후 `ee99507fb` | 실패 지점(병합 후 출력) |
|---|---|---|---|
| `internal/template` TestLateBranchTemplateMirror | `ok` | `FAIL` | `spec-assembly.md` source 32960 B vs mirror 32568 B |
| `internal/template` TestRuleTemplateMirrorDrift | `ok` | `FAIL` | `spec-workflow.md` source 38565 B vs mirror 40124 B |
| `internal/template` TestSanitizedPairParity | `ok` | `FAIL` | `agent-common-protocol.md` net one-sided=11, `plan-auditor.md` net one-sided=17 (tolerance 4) |
| `internal/template` TestPipelineCarrySignal_LocalMirrorMatches | `ok` | `FAIL` | `doc-generation.md` local mirror differs |
| `internal/hook` TestHookWrapperCopiesStayIdentical | `ok` | `FAIL` | `sync-phase-quality-gate.sh` 서브테스트만 FAIL(나머지 5개 PASS) |
| `internal/constitution` TestRegistrySyncGuard | `ok` | `FAIL` | `local` 서브테스트: `[DRIFT] CONST-V3R5-027`, `CONST-V3R5-028` (drift_count=2) — `template` 서브테스트는 PASS |
| `internal/spec` TestACCounterExtractedFromBothCarriers | `ok` | `FAIL` | `manager-docs.md` 로컬·템플릿 byte 불일치 |

측정 명령(병합 후 예): `go test ./internal/template -run TestLateBranchTemplateMirror -count=1`. 병합 전은 같은 명령에 `go -C <archive-dir>`를 붙였다.

Constitution Check CI 단계(`./bin/moai constitution validate`, 2 errors)는 `continue-on-error` 보조 신호다(`.github/workflows/ci.yml` 573-581행 주석). 판정하는 가드는 위의 TestRegistrySyncGuard이며, 두 오류는 같은 CONST-V3R5-027/028 드리프트로 보인다(아래 Gaps 참조).

## 3. 귀속 — 어느 F 병합이 어느 쌍을 깼나

명령: `git log --format='%h %s' --name-only 4c99d973e..ee99507fb -- <로컬 경로> <템플릿 경로>` (쌍마다 실행). 출력에 템플릿 경로는 한 번도 나오지 않았다.

| 쌍 | 로컬만 수정한 커밋 | 판정 |
|---|---|---|
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | `337fa9d5c` (t639 스냅숏 소비자), `de957342a` (t640 Kotlin·다중 언어) | 수리 |
| `.claude/agents/moai/manager-docs.md` | `08284485a` (t623 tier별 AC 원천) | 수리 |
| `.claude/rules/moai/core/agent-common-protocol.md` | `6896eef37` (t635 사전 fetch 순서) | 수리 |
| `.claude/agents/moai/plan-auditor.md` | `ff12d8eb8` (t620 입력 타입), `4eab936f4` (t607 해시·버전 필드) | 수리 |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | `b20e53ebc` (t622), `f6af2bb11` (t625), `df0fc2224` (t646) | 보류 |
| `.claude/skills/moai/workflows/project/doc-generation.md` | `ff12d8eb8` (t620), `cd9d1e47f` (t621), `df0fc2224` (t646) | 보류 |
| `.claude/rules/moai/workflow/spec-workflow.md` | `694be1f01` (t614), `4a33e0905` (t615), `4eab936f4` (t607), `9d34e9875` (t645) | 보류 |

레지스트리 조항 두 개의 제거 커밋: `git log -S 'Step 1 (plan) MUST execute in main checkout' 4c99d973e..ee99507fb -- .claude/rules/moai/workflow/spec-workflow.md` → `4a33e0905`(t615). `Step 4 (cleanup) applies to`도 같은 커밋.

## 4. 수리 내용

어느 쪽이 의도된 새 내용인지는 각 F 커밋의 diff로 판정했다. 네 쌍 모두 로컬 쪽 추가분이 프로그래밍 언어 중립적인 교리·동작 개선이고, 카드 번호·SPEC ID·날짜 같은 템플릿 금지 내용이 없다.

| 파일 | 방식 |
|---|---|
| `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | 로컬을 그대로 복사(byte 동일 쌍). `.sh.tmpl` 짝은 원래 없음 |
| `internal/template/templates/.claude/agents/moai/manager-docs.md` | 로컬을 그대로 복사(byte 동일 쌍) |
| `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | 정제(sanitized) 쌍 — t635 변경 hunk만 템플릿에 반영 |
| `internal/template/templates/.claude/agents/moai/plan-auditor.md` | 정제 쌍 — t607·t620 변경 hunk 3곳만 반영 |
| `internal/template/templates/.codex/agents/moai/{manager-docs,plan-auditor}.toml` | `make agents-emit`로 재생성(손편집 없음) |
| `internal/template/catalog.yaml` | `gen-catalog-hashes.go --entry manager-docs`, `--entry plan-auditor`로만 갱신 |

## 5. 수리 후 측정 (트리: `ee99507fb` + 이 카드의 작업 트리 변경)

| 명령 | 결과 |
|---|---|
| `go test ./internal/hook -run TestHookWrapperCopiesStayIdentical -count=1 -v` | exit 0, `--- PASS` 7줄, `ok` |
| `go test ./internal/spec -run TestACCounterExtractedFromBothCarriers -count=1 -v` | exit 0, `--- PASS: TestACCounterExtractedFromBothCarriers`, `ok` |
| `go test ./internal/template -run TestSanitizedPairParity -count=1 -v` | exit 0, `--- PASS` 10줄, `plan-auditor.md ... net one-sided=1 (tolerance 4)`, `ok` |
| `go test ./internal/template/agentemit/... -count=1` | `ok` |
| `make agents-emit` | exit 0 (골든 테스트 `ok`) |
| TestTemplateNeutralityAudit / TestTemplateNoInternalContentLeak / TestLanguageNeutrality / TestAllAgentsInCatalog / TestManifestHashFormat | 각각 `--- PASS`, `ok` |

## 6. 보류한 3쌍 — 수정하지 않은 이유

미러 테스트가 byte 동일을 요구하므로 일부만 옮기는 부분 수리는 불가능하다. 전부 옮기거나 전부 두어야 한다.

1. **spec-assembly.md** (TestLateBranchTemplateMirror)
   - t646이 넣은 주석이 `.claude/hooks/moai/trace-ledger.sh`와 `trace-ledger-contract.md`를 가리키는데, 둘 다 로컬에만 있고 템플릿에는 없다(`ls internal/template/templates/.claude/hooks/moai/trace-ledger.sh` → No such file). 그대로 복사하면 배포판이 없는 파일을 가리킨다.
   - t625가 넣은 "`harness.yaml`의 tier별 천장값(S=1, M=2, L=3)" 문장이 가리키는 `plan_audit_tier_ceilings` 키가 로컬 `harness.yaml`에는 있고(1건) 템플릿 `harness.yaml`에는 없다(0건). 로컬 `.moai/config`는 `moai update`가 템플릿으로 통째로 덮으므로, 이 키는 다음 update 때 로컬에서도 사라진다.
   - 결정 필요: trace-ledger 훅·계약 문서와 tier 천장값 키를 배포판에 함께 싣는가.
2. **doc-generation.md** (TestPipelineCarrySignal_LocalMirrorMatches) — 1과 같은 t646 trace-ledger 의존.
3. **spec-workflow.md** (TestRuleTemplateMirrorDrift + TestRegistrySyncGuard/local + CI Constitution Check)
   - t615가 `[ZONE:Frozen]` 조항을 고쳐 Route A를 main 직행에서 "런처 워크트리 + 로컬 통합" 모델로 바꿨다. 이는 이 저장소의 사설 git-flow 모델이라, 템플릿에 복사하면 모든 사용자 배포판에 실린다.
   - 반대로 로컬을 두면 zone 레지스트리가 제거된 조항 CONST-V3R5-027/028을 계속 가리켜 `local` 가드가 실패한다. 레지스트리를 고치는 것은 Frozen 조항 변경이다.
   - 결정 필요: 운영자 판단(사설 모델을 배포판에 싣지 않는다면 레지스트리 조항을 로컬 쪽 새 문구에 맞춰 개정할지).

## 7. Gaps (관측하지 못한 것)

- CI의 Constitution Check 단계(`./bin/moai constitution validate`)는 직접 돌리지 않았다. 트리에서 바이너리를 빌드하면 `internal/cli`까지 컴파일되는데, 이 카드의 제약상 그 슬롯이 없다. 2 errors가 CONST-V3R5-027/028과 같은 것이라는 판단은 TestRegistrySyncGuard `local` 출력에 근거한 추론이다.
- 네 패키지의 선택자 테스트만 돌렸다. 패키지 전체와 darwin/windows 매트릭스 판정은 develop push 후 CI 몫이다.
- 보류 3쌍의 실패는 병합 후 트리에서 잰 값이고, 이 카드는 그 파일들을 건드리지 않았으므로 그대로 빨간 상태로 남는다.

## 8. Residual-risk

- t646은 trace-ledger 주석을 spec-assembly.md·doc-generation.md 말고도 로컬 스킬 여러 곳(plan.md, run.md, sync.md, context-discovery.md, codebase-analysis.md, mode-detection.md, meta-harness.md, project.md)에 넣었다. 이 파일들의 템플릿 쪽이 같은 드리프트를 갖는지, 그것을 잡는 테스트가 있는지는 확인하지 않았다. 테스트가 없는 쌍이라면 드리프트가 조용히 남는다.
- F01~F40이 템플릿을 한 번도 건드리지 않았다는 사실은, 이번 8건 말고도 테스트가 없는 로컬 전용 변경이 더 있을 수 있다는 뜻이다. 범위 전체의 로컬↔템플릿 대응 점검은 이 카드 범위 밖이다.

---

## 2부 — 운영자 결정 반영 (기준 트리: 로컬 develop `f1f034bb4`로 fast-forward)

### (B) spec-workflow.md — t615 변경분만 병합 전 문구로 복원

- 복원 방법: `git diff 4a33e0905 4a33e0905^ -- .claude/rules/moai/workflow/spec-workflow.md`로 만든 역패치를 `git apply` (사전 `git apply --check` exit 0). 범위는 그 커밋이 이 파일에 넣은 hunk로 한정했다.
- 결과: `go test ./internal/constitution -run TestRegistrySyncGuard -count=1 -v` → `local`·`template` 서브테스트 모두 `--- PASS`, `ok`. CONST-V3R5-027/028 드리프트가 사라졌다.
- **byte 동일은 달성하지 못했다.** 복원 후에도 로컬↔템플릿 diff가 25줄 남는다(`diff` 출력). 남은 차이는 다른 F 커밋 셋이다.
  - t614 `694be1f01`: Route A/B 요약 두 줄 — "protected integration"과 `delivery-policy.md` 참조
  - t607 `4eab936f4`: 감사 캐시 건너뛰기 입력 문단(중립적, 템플릿 반영에 걸림돌 없음)
  - t645 `9d34e9875`: Agent Teams 문단 — `team-capability-resolver.md` 참조
  - `delivery-policy.md`와 `team-capability-resolver.md`는 로컬에만 있고 템플릿에 없다. 따라서 `TestRuleTemplateMirrorDrift/spec-workflow.md`는 여전히 FAIL이다.
- **복원이 만든 파일 내부 모순:** t614가 남긴 Route A 요약은 "phase 에이전트 직접 push 없음, manager-git 통합"이라고 쓰고, 복원된 t615 이전 표는 "`main` (direct) push"라고 쓴다. 한 파일에 두 모델이 공존한다.
- 같은 커밋의 다른 파일 판독(수정하지 않음):
  - `main-checkout-branch-guard.md`: `git worktree add` 예시를 `moai cc -w <name>`으로 바꾼 변경. spec-workflow.md 문구를 전제하지 않는다.
  - `.claude/hooks/tests/test-shared-checkout-contract.sh`(t615가 새로 만든 로컬 전용 셸 테스트): spec-workflow.md에 `launcher worktree`와 `shared primary checkout is not an execution fallback`가 있고 `git reset --hard origin/main` 줄이 없음을 전제한다. **복원된 문구에서는 이 테스트가 실패한다.** 이 스크립트를 부르는 Go 테스트·CI 워크플로·Makefile은 없다(`grep -rn test-shared-checkout-contract`에서 자기 자신 외 0건).

### (A) tier 천장값 키와 trace-ledger 미러

- `internal/template/templates/.moai/config/sections/harness.yaml`에 `plan_audit_tier_ceilings`(S=1, M=2, L=3) 추가. 로컬 주석의 SPEC ID는 걷어냈다.
- trace-ledger 중립성 판정 — **미러함.**
  - `trace-ledger.sh`: 순수 bash + `jq`, 특정 프로그래밍 언어 가정 없음, SPEC ID·날짜·커밋 SHA 없음.
  - `trace-ledger-contract.md`: `paths:` 한정 규칙(항상 로드 아님), 같은 기준으로 내부 내용 없음.
  - 또 두 파일 모두 `moai update`가 통째로 지우는 관리 대상 뿌리(`.claude/hooks/moai`, `.claude/rules/moai`) 안에 있어서, 템플릿에 없으면 로컬에서도 다음 update 때 사라진다.
- 미러: `trace-ledger.sh`(실행 권한 유지), `trace-ledger-contract.md`, `spec-assembly.md`, `doc-generation.md`를 byte 그대로 복사. catalog는 `gen-catalog-hashes.go --entry moai`(스킬 트리 해시).
- 파생 수정(리드 지시 밖, 직접 인과):
  - `internal/config/testdata/shipped_key_inventory.yaml`: 새 키 3개를 P(prose-consumed, 근거 `.claude/agents/moai/plan-auditor.md`)로 분류. 없으면 `TestShippedConfigKeysHaveReaders`가 "3 shipped config key(s) are NOT in the triage inventory"로 실패했다(측정함).
  - `internal/config/loader.go` `knownHarnessTopLevelKeys`에 `plan_audit_tier_ceilings` 한 줄 추가(gofmt가 map 정렬을 다시 맞춰 diff는 21줄). 없으면 이 키를 가진 모든 harness.yaml 로드가 `HRN_SCHEMA_DRIFT` 경고를, `MOAI_CONFIG_STRICT=1`에서는 오류를 낸다. 로컬 harness.yaml은 이미 이 상태였다.
- `provisional_tier`(t622)는 템플릿 clarity-interview에 없는 개념이다. spec-assembly 문구가 "없으면 질문한다"이므로 템플릿에서는 이전 동작으로 떨어질 뿐 깨지지 않는다.

### 2부 측정 (트리: 이 커밋)

| 명령 | 결과 |
|---|---|
| `go test ./internal/config -count=1` | `ok` |
| `go vet ./internal/config/` | exit 0 |
| `go test ./internal/constitution -run TestRegistrySyncGuard -count=1 -v` | PASS 5줄(local·template 포함), `ok` |
| `go test ./internal/template -run '<10개 선택자>' -count=1 -v` | PASS 11 / FAIL 1 (`TestRuleTemplateMirrorDrift/spec-workflow.md` — 위 (B)의 남은 25줄) |
| `go test ./internal/hook -count=1` | 1건 FAIL `TestSessionStart_DeferredScanDoesNotBlockReturn`(731ms, load avg 25.68). 같은 선택자 단독 재실행 `--- PASS` 0.49s → 부하 의존 시간 테스트로 판단, 이 변경과 무관 |

### 2부 Gaps

- `moai constitution validate`는 아직 실행하지 않았다(`internal/cli` 컴파일 슬롯 필요).
- spec-workflow.md byte 동일과 파일 내부 모순 해소는 결정이 필요하다: t614·t645 참조 문서 둘을 템플릿에 미러해 남은 hunk를 전파할지, 아니면 t614·t645도 복원할지.
- t646 trace-ledger 주석이 들어간 나머지 로컬 스킬 14곳의 템플릿 대응은 여전히 점검하지 않았다.

### 2부 추가 — (B) 운영자 결정 반영: spec-workflow.md diff 0

- 결정: 로컬의 t614·t645 hunk는 템플릿 문구로 되돌리고, 중립인 t607 hunk만 템플릿에 전파한다. 참조 문서 `delivery-policy.md`·`team-capability-resolver.md`는 템플릿에 미러하지 않고 로컬 파일은 그대로 둔다. t614·t645의 의도는 후속 검토 카드로 넘긴다.
- 반영: 템플릿 412행에 t607 문단(`runtime.ResolveLatestPlanAudit`, 정확한 바이트 해시, 메타데이터 없는 리뷰 파일은 캐시 미스)을 넣은 뒤, 템플릿을 로컬에 복사했다. 복사 직전 diff는 t614(25-26행)·t645(445-454행) 두 영역뿐이었고, 복사 후 `cmp`가 두 파일이 동일하다고 보고했다.
- 결과: 한 파일에 두 모델이 공존하던 문제(t614 요약 vs 복원된 표)가 사라졌다. Route A는 병합 전 문구 그대로다.
- 잔여(이번에 손대지 않음): `.claude/hooks/tests/test-shared-checkout-contract.sh`는 t615의 새 문구를 전제하므로 현재 문구에서는 실패한다. 호출하는 Go 테스트·CI·Makefile은 0건이다.
- Constitution Check: `a9b15fb62` 트리로 빌드한 바이너리(스크래치 경로)로 `MOAI_CONSTITUTION_REGISTRY=.claude/rules/moai/core/zone-registry.md moai constitution validate` → exit 0, `OK — no drift or violations detected`. 같은 바이너리를 `ee99507fb` 전체 트리 사본에서 실행하면 exit 1, `[DRIFT] CONST-V3R5-027`·`CONST-V3R5-028`(CI의 2 errors와 일치)이 나와 검사기가 실제로 판정함을 확인했다. 출력의 "0 entries checked"는 검사 수가 아니라 문제 항목 수를 가리킨다(`constitution list`는 101개 항목). 이번 (B) 반영은 레지스트리 조항 문구를 바꾸지 않으므로 validate를 다시 돌리지 않고 `TestRegistrySyncGuard`로 재측정했다.
