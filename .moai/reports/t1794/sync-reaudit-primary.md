auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: e04ebd5ca6bc441b380c1072d5a1dff8431459ba
required_backend: codex fail rcpt-a658b9c5ab7740d86dd4e057
spec_verdict: SPEC-LOCAL-MAIN-FLOW-001 PASS-WITH-DEBT
spec_verdict: SPEC-RELUP-STRUCT-AUDIT-001 PASS
spec_verdict: SPEC-ROLE-INJECTION-BUDGET-001 FAIL

# t1794 — 1차 체크아웃 병합 후 sync 재감사 (3개 SPEC)

- 감사 대상: 로컬 `main` = `e04ebd5ca6bc441b380c1072d5a1dff8431459ba` (`git rev-parse main`)
- 1차 체크아웃 상태: HEAD 가 `478285ab4` 에 분리(detached)돼 있고 `main` 보다 863 커밋 뒤처져 있다. 따라서 코드 검증은 작업 트리가 아니라 `main` 스냅샷에서 했다 — `git archive main` 추출본(스크래치)과, git 이 필요한 테스트는 `main` 을 분리 체크아웃한 공유 클론(스크래치)에서 실행. 추적 파일은 하나도 수정하지 않았다.
- 판정 빌드: 테스트는 각 스냅샷 소스를 `go test` 로 직접 컴파일했다(설치 바이너리 미사용). `moai verify audit-plan` 만 설치 바이너리(`69c1ffc1f`, `main` 의 조상, 323 커밋 뒤)로 읽었다.

## 종합 판정: FAIL

1. **SPEC-ROLE-INJECTION-BUDGET-001 — FAIL.** 릴리스 블로킹 AC-RIB-007 이 현재 `main` 에서 거짓이다. t1617 이 `190f7db5d` 로 넣은 hooks-system.md 의 두 한도 구분 서술(50K 총 stdout / 10,000자 문자열당 전달)이, 이후 `ec193e477`("chore: commit the primary checkout's pending changes onto main", 10-14)이 1차 체크아웃의 낡은 작업 사본을 그대로 커밋하면서 배포본·템플릿 양쪽에서 지워졌다.
2. **codex 필수 레그 — fail** (영수증 `rcpt-a658b9c5ab7740d86dd4e057`). 다만 지적 대상은 SPEC 표면이 아니라 1차 체크아웃의 **미커밋** 파일이다(아래 Evidence E-C).
3. SPEC-LOCAL-MAIN-FLOW-001 은 PASS-WITH-DEBT, SPEC-RELUP-STRUCT-AUDIT-001 은 PASS.

## 결함 목록 (structured defect-list)

- **F1 [High] [blocking]** `.claude/rules/moai/core/hooks-system.md` 및 `internal/template/templates/.claude/rules/moai/core/hooks-system.md` (Hook stdout 50K 행 부근) — AC-RIB-007/REQ-RIB-009 회귀: 10,000자 per-`additionalContext` 문자열 한도 행과 "이 50K 한도는 TOTAL stdout" 단서가 `ec193e477` 로 삭제됨. 확신도 높음(git 이력으로 직접 관측). — 필요한 수정: `190f7db5d` 가 넣은 두 행을 템플릿 원본부터 복원하고 배포본에 미러(쌍 바이트 동일 유지), 템플릿 중립성(SPEC/카드 id·내부 날짜 없음) 유지. 확인 명령 `grep -c '10,000' .claude/rules/moai/core/hooks-system.md` ≥1.
- **F2 [Medium] [blocking — 프로세스]** 1차 체크아웃이 `478285ab4` 에 분리된 채 `ec193e477` 의 원천인 낡은 미커밋 변경(hooks-system.md 등 14개 파일)을 아직 들고 있다. 이 상태에서 같은 "pending changes 커밋" 이 반복되면 F1 류 회귀(및 t1704 가 지운 `2026-10-09` 날짜 재유입)가 다시 생긴다. 확신도 중간. — 필요한 수정: 리더가 1차 체크아웃 미커밋 변경의 소유 세션을 확인해 정리(이 감사는 손대지 않음).
- **F3 [Low] [optional]** `internal/factory/integration_merge_step_test.go:367` — `exec.Command("git","rev-parse","HEAD^{tree}")` 에 `cmd.Dir` 가 없어 프로세스 cwd 저장소의 트리를 읽는다(카드 트리가 아님). git 저장소 밖에서 실행하면 `TestMergeStepPreMergeCausesReleaseWithDistinctCodes/3_not_a_descendant` 가 실패. card t1479 코드이며 세 SPEC 의 결함은 아님. 확신도 높음(재현됨). — 필요한 수정: `cmd.Dir = f.cardTree` 지정.
- **F4 [Low] [optional]** SPEC-LOCAL-MAIN-FLOW-001 F-11 의 축자 grep(`카드 브랜치는 `main`에서 판다` = 1)이 이후 `76689feaf`(t1698 영문 번역)로 0이 됨. 의미는 영문 `Card branches are cut from `main`.` 로 보존. 문서 증명 명령의 표류(debt).

## 1. Claim

| # | 주장 |
|---|---|
| C1 | SPEC-RELUP-STRUCT-AUDIT-001: 검사 패밀리 5개 테스트가 `main` 에서 통과하고, AC-RSA-009(git 호출 0)·AC-RSA-010(`last-codex-version` 0힛)이 유지된다 → PASS |
| C2 | SPEC-ROLE-INJECTION-BUDGET-001: AC-RIB-001/003/004/005/006/009/010 의 테스트·명령은 통과하나, 릴리스 블로킹 AC-RIB-007 이 `main` 에서 실패 → FAIL |
| C3 | SPEC-LOCAL-MAIN-FLOW-001: V-명령 테스트 패밀리(cli 80개·factory 41개·config·template)가 통과하고 F-체크 대부분이 성립, 차이는 카드 종결 시점 한정 조건과 이후 의도된 문서 변경뿐 → PASS-WITH-DEBT |
| C4 | codex 필수 레그는 verdict `fail` 을 냈고 영수증을 발행했으나, 리뷰 범위는 SPEC 병합 표면이 아니었다 |

## 2. Evidence (명령 + 축자 출력)

### E-R (RELUP-STRUCT-AUDIT) — `git archive main` 추출본
```
$ go test -count=1 -run 'TestRelupStructure' ./internal/harness/
ok  	github.com/modu-ai/moai-adk/internal/harness	0.886s        (rc=0)
$ go test -count=1 -v -run 'TestRelupStructure' ./internal/harness/ | grep -E '^--- '
--- PASS: TestRelupStructureWatchNormSentence (0.01s)
--- PASS: TestRelupStructureManifestDomain (0.00s)
--- PASS: TestRelupStructureBPFaces (0.00s)
--- PASS: TestRelupStructureCodexParagraph (0.00s)
--- PASS: TestRelupStructurePhase3URLs (0.00s)
(서브테스트 포함 PASS 행 30개)
$ grep -nE 'os/exec|exec\.Command|"git' internal/harness/relup_structure_audit_test.go   → 출력 없음, rc=1   (AC-RSA-009)
$ grep -rn "last-codex-version" internal/                                                → 출력 없음          (AC-RSA-010)
```

### E-B (ROLE-INJECTION-BUDGET)
```
$ go test -count=1 -run 'TestCompositeBudgetRealAssembly|TestRoleInjectionAssemblyBudget|TestRoleRulesVersionSkew|TestSessionStartRoleRulesSizeGate|RenewInterval|RoleRules|RoleInjection' ./internal/hook/
ok  	github.com/modu-ai/moai-adk/internal/hook	6.617s   (rc=0; -list 매칭 38개)
--- PASS: TestCompositeBudgetRealAssembly / TestRoleInjectionAssemblyBudget / TestSessionStartRoleRulesSizeGate
--- PASS: TestRoleRulesVersionSkew{DiagnosticSkipsFIFO,Predicate,ReadsNestedMoaiStamp,UpdateGuidance,FailureDetailNamesRemedy}
$ awk 'END{print NR; exit !(NR>=36)}' .../relocation-ledger.md        → 58, rc=0   (AC-RIB-002)
$ grep -c 'moai update' internal/hook/role_rules.go                    → 12          (AC-RIB-004)
$ grep -c 'template_version' internal/hook/role_rules.go               → 4           (AC-RIB-005)
$ awk '/REQ-ALB-009 retreat/{n++} END{print n; exit !(n==1)}' internal/hook/role_rules.go → 1, rc=0 (AC-RIB-006)
$ grep -c '10000 characters' internal/hook/role_rules.go → 0 ; grep -c '10,000 characters' → 3
$ grep -c 'stub-delta' .../SPEC-ROLE-INJECTION-BUDGET-001/progress.md  → 2           (AC-RIB-009)
$ grep -c '## Isolation' .claude/rules/moai/workflow/factory-dispatch.md (배포·템플릿) → 1 / 1 (AC-RIB-008)
core 스텁 UTF-16 길이: cross-session-messaging-core.md 8072, factory-dispatch-core.md 9983 (≤10,000, AC-RIB-003)

[AC-RIB-007 — 실패]
$ grep -c '10,000' .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md
.claude/rules/moai/core/hooks-system.md:0
internal/template/templates/.claude/rules/moai/core/hooks-system.md:0
$ git log main --oneline -S'10,000' -- .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md
ec193e477 chore: commit the primary checkout's pending changes onto main
190f7db5d feat(SPEC-ROLE-INJECTION-BUDGET-001): M3 — version-skew detection + moai update guidance, doc dedup, hooks-system dual caps (card t1617)
$ git show ec193e477 -- <두 hooks-system.md> | grep -E '^[-+]' | grep -E '10,000|50K'
-- Hook stdout over 50K characters is saved to disk; ... (v2.1.89+) — this 50K cap is on a hook's TOTAL stdout.
-- A second, distinct cap applies per `additionalContext` string: a single string over 10,000 characters (UTF-16 code units) is not delivered inline — ...
+- Hook stdout over 50K characters is saved to disk; only a file path + preview is injected into context (v2.1.89+)
(배포본·템플릿 각 1쌍, 동일)
$ git merge-base --is-ancestor ec193e477 main → rc=0
```

### E-L (LOCAL-MAIN-FLOW)
```
$ go test -count=1 -run 'TestLocalMain|TestAutoMerge|TestIntegrationSurface|TestIntegrationMergeWorktree|TestFactoryCompletePrimaryTreeGate|TestFactoryCompleteNoIntegrationTreeRefused|TestSD_AC0|TestMWQ19_Scenario5' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	597.488s   (rc=0; -list 매칭 80개)
$ go test -count=1 -run 'TestLocalMainIntegration|TestEmptyTargetGuidance|TestTargetProvenance|TestStructYAMLSymmetry|TestAssertion' ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/config	0.539s
$ go test -count=1 -run 'TestWorkflowWorktreeKeyHonesty|TestAgentsDisclosureCompleteness|TestAutoMerge' ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	36.807s
$ go test -count=1 -run 'TestMergeStep' ./internal/factory/      [git archive 추출본 — git 저장소 아님]
--- FAIL: TestMergeStepPreMergeCausesReleaseWithDistinctCodes (15.70s)
    integration_merge_step_test.go:369: tree: exit status 128: fatal: not a git repository (or any of the parent directories): .git
FAIL	github.com/modu-ai/moai-adk/internal/factory	470.052s
$ go test -count=1 -run 'TestMergeStep' ./internal/factory/      [main 을 분리 체크아웃한 공유 클론, HEAD e04ebd5ca]
ok  	github.com/modu-ai/moai-adk/internal/factory	498.715s   (rc=0; -list 매칭 41개)

F-체크 (추출본 / git 은 main 기준):
F-01  workflow.yaml(템플릿) 248: local_main_integration: / enabled: false            ✔
F-07  '^### B7' 1, 't1621' 7                                                           ✔
      git diff --quiet e32f69c46 main -- internal/homestate → rc=1 (이후 t1707·t1696 커밋; "카드 종결 시점" 한정 조건)
F-08  develop_branch: main, manual: 8행, auto_merge: false, 'auto_merge 유지' 0, 09a42899c ⊂ main ✔
F-09  deny_commits_on: [main]; local_main_integration enabled: true                     ✔
F-10  금지 문구 0, 'remote default branch' 1, 'from local `main`' 0, AGENTS.md.tmpl 무변경 ✔
F-11  '^### §4.0' 1, 'SUPERSEDED by §4.0' 7, '카드 브랜치는 `main`에서 판다' 0 ✘ (76689feaf 영문화 — 의미 보존: 259행 "Card branches are cut from `main`.")
F-12  'Card worktrees branch FROM' 0                                                     ✔
F-13  wc -c 23780 (<24228), wc -m 23680 (<40000)                                        ✔
F-14  1 / 1 / 4                                                                           ✔
F-15  2 / 2 / 2                                                                           ✔
F-16  '^Operator verdict:$' 0 ("결정이 열려 있는 동안" 한정 — 결정 기록됨), GATED-ON-DECISION 7 ✔
F-17  e32f69c46^{tree} = 09a42899c^{tree} = ddc07c9a38...                                 ✔
```

### E-C (codex 필수 레그)
```
mcp__moai__codex_audit(project_root=/Users/goos/MoAI/moai-adk-go, target=baseBranch, mode=native)
verdict: fail ; audit_receipt: rcpt-a658b9c5ab7740d86dd4e057 ; build_commit: 69c1ffc1f4
review_base: "main (merge base 478285ab4c1fc893b020bb464d14cb69c8422523)"
finding [P2]: 배포 템플릿에서 내부 변경일을 제거해 — internal/template/templates/.claude/rules/moai/core/hooks-system.md:368
  추가한 `2026-10-09`는 ... strict `TestTemplateNoInternalContentLeak`의 금지 패턴에 해당 ...

$ git diff HEAD -- internal/template/templates/.claude/rules/moai/core/hooks-system.md   (1차 체크아웃 미커밋 변경)
-| UserPromptSubmit | 5s | 30s | Blocks user interaction; reduced max |
+| UserPromptSubmit | 15s | 30s | Blocks user interaction; reduced max. Raised from 5s (2026-10-09): ...
$ git log main --oneline -1 5fd8ea69e → fix(templates): drop internal work-date provenance from hooks-system timeout table (card t1704)

$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go
config_status ok, model codex, codex gate required (claude off, glm off), cross_model_active true, enforced_required [codex]
```

## 3. Baseline-attribution

- 모든 테스트·grep 은 이번 실행에서 `main` = `e04ebd5ca` 스냅샷에 대해 수행했다(추출본은 `git archive main`, 공유 클론은 `e04ebd5ca` 분리 체크아웃). 1차 작업 트리(HEAD `478285ab4` + 미커밋 변경)는 테스트 대상이 아니다.
- 테스트 바이너리는 각 스냅샷 소스에서 `go test` 가 컴파일한 것. `audit-plan` 판독만 설치 바이너리 `69c1ffc1f`(main 의 조상)로 했다 — 설정 판독이라 코드 판정에는 쓰지 않음.
- codex 레그의 리뷰 범위는 서버가 정한 `main` 과 1차 HEAD `478285ab4` 의 merge base 기준 diff 로, HEAD 가 `main` 의 조상이므로 커밋 diff 는 사실상 비어 있고 실제로는 1차 미커밋 변경을 읽었다.

## 4. Gaps

- **codex 레그가 세 SPEC 의 병합 표면을 리뷰하지 않았다.** 1차 체크아웃이 `478285ab4` 에 분리돼 있어 `baseBranch` 대상이 SPEC 델타를 담지 못했다. 영수증은 이 트리 스토어에 발행됐지만 내용상 SPEC 교차 검토는 아니다. 지시대로 `uncommittedChanges` 는 쓰지 않았고, 결과가 inconclusive 가 아니어서 재시도하지 않았다.
- 감사관 정의상 `cross_model_active: true` 면 `mcp__moai__audit_multi` 경로가 원칙이나, 리더 지시에 따라 `codex_audit` 단일 레그만 실행했다.
- codex 지적(`TestTemplateNoInternalContentLeak` 실패)은 직접 실행해 확인하지 않았다(1차 미커밋 파일 대상, 부하 ~540). 해당 줄에 `2026-10-09` 가 있음만 diff 로 확인했다.
- AC-RIB-011(원장 36행 정합)·AC-RIB-012/013(라이브 레인 재진입 관측)은 이번에 별도 재관측하지 않았다. internal/template 패키지는 LMF 셀렉터 테스트만 돌렸고 전체 패키지는 돌리지 않았다.
- LMF AC-LMF-016 의 첫 병합 실측, F-13 의 커밋별 비용 서술 검사, AC-RSA-011(카드 종결 시 4채널 스코프), AC-RSA-012(인계 표 12행 grep)는 재관측하지 않았다.
- 전체 스위트·커버리지·golangci-lint 는 지시(부하)로 실행하지 않았다 — CI 몫.

## 5. Residual-risk

- F1 수정 전까지 hooks-system.md 독자는 10,000자 문자열당 한도를 모른 채 SessionStart 주입 크기를 판단할 수 있다(역할 주입 넘침 재발 위험 — 코드 게이트는 살아 있어 실제 넘침은 테스트가 잡음).
- `ec193e477` 은 같은 방식으로 다른 파일(config·로그·policies 등 36개)도 덮었다. SPEC 관련 회귀는 hooks-system.md 만 확인했으며, 다른 SPEC 표면의 회귀 여부는 이 감사 범위 밖이다.
- 재감사 결과가 FAIL 이므로 receipt 가드의 sync-auditor 거부 4건은 이 감사로 해제되지 않는다(가드는 증명된 PASS 만 해제).
- 스크래치의 원시 출력(`scratchpad/out/*.txt`)은 기계 로컬이며 인용하지 않았다 — 결정적 줄은 위에 옮겼다.

## 반복 이력

- r1 (이번): 1차 트리 병합 후 재감사, 영수증 `rcpt-a658b9c5ab7740d86dd4e057`.
