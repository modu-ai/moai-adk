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
