auditor-model: claude-opus-5-5

# t1702 sync 감사 iter10 (H1 수리 델타 a29bc40be5..f8bf1b6669 확인, Claude 단독)

verdict: PASS
audited_sha: f8bf1b6669a700f7b037b681983dbf7ab5eed719

## 요약

판정은 PASS다. 범위는 운영자 결정 d-20261106T162918Z-afc8대로 H1 수리 델타와 홈 접근 가드 테스트로 한정했다.

- 델타는 테스트 파일 텍스트 네 곳뿐이고 제품 코드는 바뀌지 않았다. 단언(조건식과 피연산자)도 하나도 바뀌지 않았다. 메시지 문자열 셋을 고쳤고, 단언 한 줄 위에 허용 주석 하나를 달았다.
- 허용 주석은 가드가 받아들이는 형식(`//\s*home-guard:allow\s+\S`, 사유 있음)이고, 바로 아랫줄(283행) 하나만 면제한다. 그 줄에는 `UserHomeDir` 호출이 없으므로 다른 것을 함께 덮지 않는다.
- 홈 가드는 여전히 FAIL이다. 다만 남은 "new" 위반은 정확히 2건이고, 둘 다 이 카드가 바꾸지 않은 파일이다(`main`에 같은 줄이 있다, 운영자가 t2137로 발행). 카드 소유 4건은 목록에서 사라졌다. 위반 수는 166에서 162로, new는 6에서 2로 줄었다.
- 지정한 카드 테스트 8개, `internal/userassets` 전체, 파괴 대상 레지스트리 가드 2개, gofmt, go vet이 모두 깨끗하다.

점수: 93/100. Functionality 95 PASS, Security 95 PASS, Craft 92 PASS, Consistency 90 PASS.

Codex leg: not run this round (operator ruling d-20261106T162918Z-afc8; the Codex pass of iter9 covers a29bc40be5, not this delta)

## 저장소 필수 Codex 게이트와 이 판정의 관계

`moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702`는 `"model": "codex"`, codex `"gate": "required"`(config.gates), `"enforced_required": ["codex"]`라고 답한다(E0). 이 회차에는 Codex를 부르지 않았으므로 **f8bf1b6669에 대한 Codex 영수증은 없다.** 기계적으로 보면 필수 codex 게이트는 이 SHA에서 충족되지 않았다. iter9 영수증 `rcpt-2973cf866890155ca03162e5`(pass, findings 없음)는 cf141b5409..a29bc40be5 델타를 본 것이고, 이번 델타는 보지 않았다.

이 PASS는 코드 판정이다. 게이트 충족을 주장하지 않는다. 이 SHA로 착지를 진행할 근거는 운영자 결정 d-20261106T162918Z-afc8(테스트 텍스트뿐인 H1 수리 델타는 Claude sync-auditor 단독으로 확인, Codex 추가 호출 없음)이다. 영수증 가드나 `--result-file` 대조가 이 판정을 막으면 우회하지 않는다. 그 차단을 리더에게 그대로 보고하는 것이 결정의 지시다. 최종 메시지의 판정 줄은 `receipts=none`으로 적는다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 요약 |
|---|---|---|---|---|
| H1 | (닫힘) | 이전 BLOCKING | `internal/cli/update_profile_reach_test.go:110`, `:259`, `:283`; `internal/userassets/profile_reach_test.go:589` | 카드 소유 `~/` 위반 4건이 홈 가드 출력에서 사라졌다(E1) |
| X1 | MINOR | optional, outside(카드 파일 아님, 판정에 영향 없음) | `internal/agenticfactory/core/sanitize_test.go:95`, `internal/cli/factory_fixture_home_test.go:36` | 홈 가드의 남은 new 위반 2건. `main`에 같은 줄이 있다(E5). 이 카드 범위 밖이고 t2137이 맡는다 |

이번 델타에서 새로 생긴 BLOCKING, SHOULD-FIX 결함은 없다. 수리에서 생긴(repair-born) BLOCKING도 없다. iter8, iter9에 이어 연속 회차 수리 중단 규칙은 걸리지 않는다.

이전 회차 항목: R1, N5, N6, B2는 iter9에서 닫혔다. R1 가드는 이번에 다시 돌려 PASS를 확인했다(E3). N7, N8(optional MINOR)은 iter9 기록 그대로 남겨 둔다. 이번 델타는 그 줄들을 건드리지 않았다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. HEAD는 f8bf1b6669a700f7b037b681983dbf7ab5eed719이고, 이 워크트리에서 감사했다.
2. `internal` 아래 델타는 테스트 파일 두 개의 네 곳이고, 단언이나 제품 동작은 바뀌지 않는다.
3. 허용 주석은 가드가 받아들이는 형식이며 283행 하나만 면제한다.
4. 홈 가드에서 카드 소유 위반은 0건이다. 남은 new 2건은 카드가 바꾸지 않은 파일이고, 둘 다 `main`에 있다.
5. 지정한 테스트, 레지스트리 가드, gofmt, go vet이 모두 통과한다.
6. 필수 codex 게이트는 이 SHA에서 영수증으로 충족되지 않았다. 진행 근거는 운영자 결정이다.

### Evidence

**E0 — 트리, 결정, 계획**
```
git rev-parse HEAD             → f8bf1b6669a700f7b037b681983dbf7ab5eed719
git rev-parse --show-toplevel  → /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702
git diff --stat a29bc40be5 f8bf1b6669
 .moai/reports/t1702/progress-r8.md        |   4 +
 .moai/reports/t1702/sync-audit-iter9.md   | 203 ++++++++++++++++++++++++++++++
 internal/cli/update_profile_reach_test.go |   5 +-
 internal/userassets/profile_reach_test.go |   2 +-
 4 files changed, 211 insertions(+), 3 deletions(-)
moai decision read --scope card:t1702 > …/scratchpad/decisions-t1702.txt 다음에 id grep:
 decision record: decided_by=operator via leader to2bbk (AskUserQuestion 2026-11-06) … id=d-20261106T162918Z-afc8 scope=card:t1702 kind=ruling
   body: Operator: the H1 repair delta f8bf1b6669 (test file text only, no product code) is confirmed by a Claude sync-auditor alone (delta + home guard test), no further codex call. If the receipt guard blocks, report back instead of working around it.
moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702
 "config_status": "ok", "model": "codex", claude gate "off", codex gate "required" (config.gates), glm "off",
 "cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
```

**E1 — `internal` 델타 원문(`git diff a29bc40be5 f8bf1b6669 -- internal`)**
```
@@ -107,7 +107,7 @@ func TestEnsureGlobalSettingsEnv_ProfileConfigRootsReach(t *testing.T) {
-		t.Errorf("~/.claude/settings.json should not be created, stat err = %v", err)
+		t.Errorf("the default root's settings.json should not be created, stat err = %v", err)
@@ -256,7 +256,7 @@ func TestUpdateGlobalSettingsEnv_UnreadableProfileStoreErrors(t *testing.T) {
-		t.Fatal("ensureGlobalSettingsEnv must propagate an unreadable profile store, not clean ~/.claude only")
+		t.Fatal("ensureGlobalSettingsEnv must propagate an unreadable profile store, not clean the default root only")
@@ -279,6 +279,7 @@ func TestDoctorSkillReachCustomConfigDirGuidance(t *testing.T) {
+	// home-guard:allow asserts the wording of the doctor advice, not a home access
 	if !strings.Contains(check.Message, "claude-profiles") || !strings.Contains(check.Message, "~/.claude") {
--- internal/userassets/profile_reach_test.go
@@ -586,7 +586,7 @@ func TestRealProfileDirNamesUnreadableBase(t *testing.T) {
-		t.Error("ClaudeConfigRoots must error on an unreadable store, not return the bare ~/.claude")
+		t.Error("ClaudeConfigRoots must error on an unreadable store, not return the bare default root")
```
문맥을 읽은 결과(Read): 110행은 `os.Stat(filepath.Join(tmp, ".claude", "settings.json"))` 단언의 메시지, 259행은 `if err == nil` 아래 메시지, 589행은 `ClaudeConfigRoots(f.home)` 단언의 메시지다. 조건식과 피연산자는 그대로다. 283행 조건식의 `"~/.claude"` 피연산자도 그대로이고, 282행에 주석만 더해졌다.

**E2 — 허용 표시 규칙(`internal/testguard/scan.go`)**
```
30: var allowComment = regexp.MustCompile(`//\s*home-guard:allow\s+\S`)
84: 			if allowComment.MatchString(c.Text) {
85: 				markerLines[fset.Position(c.Slash).Line] = true
96: 		if seen[key] || markerLines[line] || markerLines[line-1] {
```
표시는 실제 주석 토큰에서만 인식된다. 면제 범위는 표시가 있는 줄과 그 바로 다음 줄이다. 282행 표시는 282행(주석 자체)과 283행만 면제한다. 사유 문자열("asserts the wording …")이 있으므로 `\s+\S`에 맞는다.

**E3 — 테스트와 정적 검사(모두 `GOCACHE=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/671835d7-2501-4df7-8d56-57dfa4e7b2d0/scratchpad/gocache-t1702`)**
```
go test ./internal/testguard -run 'TestHomeGuard_RepositoryHasNoNewHomeAccess' -count=1
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (5.18s)
    repo_guard_test.go:87: NEW home access internal/agenticfactory/core/sanitize_test.go:95: ~/ literal: "~/work/notes.md", // a tilde path names no user
    repo_guard_test.go:87: NEW home access internal/cli/factory_fixture_home_test.go:36: ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
    repo_guard_test.go:96: scanned 3931 test files: 162 violations, seed 184 entries, 2 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
FAIL	github.com/modu-ai/moai-adk/internal/testguard	5.474s
(각 위반 아래 "Use testguard.RedirectHome(t) … or add `// home-guard:allow <reason>` …" 안내 줄은 생략)

go test ./internal/cli -run 'TestEnsureGlobalSettingsEnv_ProfileConfigRootsReach|TestRunUserAssetUpdatePhase_|TestCheckProfileSkillReach|TestDoctorSkillReach|TestUpdateGlobalSettingsEnv_UnreadableProfileStoreErrors|TestUserAssetRootsLineConsumesInstallSnapshot' -count=1 -v
--- PASS: TestEnsureGlobalSettingsEnv_ProfileConfigRootsReach (0.01s)
--- PASS: TestRunUserAssetUpdatePhase_ReachesProfileAndNamesRoots (1.71s)
--- PASS: TestRunUserAssetUpdatePhase_RootsLineWithoutProfiles (1.26s)
--- PASS: TestCheckProfileSkillReach (0.03s)
--- PASS: TestDoctorSkillReachUnreadableProfileStoreWarns (0.00s)
--- PASS: TestUpdateGlobalSettingsEnv_UnreadableProfileStoreErrors (0.00s)
--- PASS: TestDoctorSkillReachCustomConfigDirGuidance (0.60s)
--- PASS: TestUserAssetRootsLineConsumesInstallSnapshot (2.19s)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.531s

go test ./internal/userassets -count=1
ok  	github.com/modu-ai/moai-adk/internal/userassets	7.032s

go test ./internal/cli -run 'TestDestructiveTargetRegistry_|TestCountDestructiveCalls_' -count=1 -v
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.03s)
--- PASS: TestCountDestructiveCalls_ASubRootOfARootHandleIsARootHandle (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.078s

gofmt -l internal/cli internal/userassets          → 출력 없음
go vet ./internal/cli ./internal/userassets        → 출력 없음
```

**E4 — 남은 위반 경로와 카드 변경 파일 대조(`git diff main...HEAD --name-only`, 37개)**
목록에 `internal/cli/update_profile_reach_test.go`와 `internal/userassets/profile_reach_test.go`는 있다. 그러나 이 두 파일은 가드 출력에 더 이상 나오지 않는다. `internal/agenticfactory/core/sanitize_test.go`와 `internal/cli/factory_fixture_home_test.go`는 목록에 **없다**. 그러므로 남은 new 위반 중 카드가 바꾼 파일에 속한 것은 0건이다.

**E5 — 남은 2건이 `main`에 있는지**
```
git grep -n -F "~/" main -- internal/agenticfactory/core/sanitize_test.go internal/cli/factory_fixture_home_test.go
main:internal/agenticfactory/core/sanitize_test.go:95:		"~/work/notes.md", // a tilde path names no user
main:internal/cli/factory_fixture_home_test.go:14:// real ~/.moai. …            (주석, 가드 대상 아님)
main:internal/cli/factory_fixture_home_test.go:17:// real ~/.moai/worktrees/… (주석, 가드 대상 아님)
main:internal/cli/factory_fixture_home_test.go:36:		t.Errorf("MoaiHome() = %q resolves outside … into the real ~/.moai", home, tmp)
```

### 제품 동작에 대한 판단
델타는 `_test.go` 파일 두 개와 `.moai/reports/t1702/` 아래 기록 두 개뿐이다(E0). 테스트 파일은 제품 바이너리에 들어가지 않는다. 바뀐 것은 실패 메시지 문자열과 주석이라서 테스트의 통과·실패 조건도 그대로다. 제품 동작이 바뀔 경로는 없다.

### Baseline-attribution
모든 측정은 이 실행에서, 이 워크트리의 트리 f8bf1b6669에 대해 했다. 테스트 바이너리는 위 비공개 GOCACHE로 이 트리 소스에서 빌드했고, 캐시 오류(`could not import … go-build`)는 없었다. 설치된 `moai`는 측정에 쓰지 않았다. 결정 열람(`decision read`)과 계획 조회(`verify audit-plan`)에만 썼다. iter9 수치(166 violations, 6 new)는 iter9 보고서의 a29bc40be5 측정값이고, 이번 수치(162, 2 new)와 비교하는 데만 인용했다.

### Gaps
- **Codex 미실행**: 운영자 결정에 따라 이번 회차에는 codex, GLM, audit_multi를 호출하지 않았다. f8bf1b6669에 대한 Codex 영수증은 없다. `moai verify audit-plan --result-file` 대조도 하지 않았다.
- `internal/cli`, `internal/hook` 전체 스위트, `go test ./...`, golangci-lint, 커버리지, Windows/Linux 실행은 하지 않았다(지시 범위 밖이고, 델타는 테스트 텍스트뿐이다).
- 홈 가드가 `main` 트리 자체에서 RED인지는 `main` 트리에서 직접 돌리지 않았다. 남은 2건이 `main`의 blob에 있다는 것(E5)과 seed에 없다는 것(iter9 E3)에서 추론한 것이다.
- 홈 가드 외에 iter9 탐색에서 고른 다른 가드는 이번에 다시 돌리지 않았다. 바뀐 두 파일은 문자열과 주석만 바뀌어서 해당 가드들의 판단에 영향이 없다고 본다. 이것은 관측이 아니라 판단이다.
- 거부된 도구 호출은 없다. `moai decision read --output`은 존재하지 않는 플래그라서 사용 오류로 끝났고, 지시된 리디렉션 형태로 다시 실행했다.

### Residual-risk
- 홈 가드는 t2137이 `main`의 2건을 고칠 때까지 RED로 남는다. 착지 판정에서 이 RED를 카드 탓으로 읽으면 안 된다. 이 카드 파일에서 새 위반이 생기면 같은 RED에 묻혀 보이지 않을 수 있으니, 그때는 위반 목록의 경로를 다시 읽어야 한다.
- 283행 허용 표시는 그 줄 전체를 면제한다. 나중에 그 줄에 진짜 홈 접근이 추가되면 가드가 잡지 못한다. 지금은 해당 사항이 없다.
- 운영자 결정이 Codex 게이트를 대신하는 이 경로를 영수증 가드가 받아들일지는 관측하지 않았다.
- 결정 기록 사본은 세션 스크래치패드에만 있다. 판정에 쓴 줄은 이 파일(E0)에 옮겨 두었다.

## 반복 이력

- iter6: FAIL @732fce152d — B1, B2, S1, S2, M1~M4.
- iter7: FAIL @bee9cfb001 — 코드 BLOCKING 없음, 필수 codex 게이트 미충족(stale_server).
- iter8: FAIL @cf141b5409 — BLOCKING R1(iter5 수리에서 생김). 영수증 rcpt-690d37c6f81b6c1daa5c7050.
- iter9: FAIL @a29bc40be5 — R1, N5, N6 닫힘, Codex pass(rcpt-2973cf866890155ca03162e5), 가드 탐색에서 BLOCKING H1(초기 커밋에서 생김).
- iter10(이번): PASS @f8bf1b6669 — H1 닫힘, 새 결함 없음. Claude 단독(d-20261106T162918Z-afc8), Codex 미실행, 이 SHA의 영수증 없음.

## 영수증 가드 차단 기록 (판정 제출 직후)

감사관이 멈출 때 SubagentStop 훅이 PASS를 거부했다. 원문은 다음과 같다. "AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS but the audit could not be corroborated — no receipt cited. Call the codex audit (codex_audit or audit_multi) for this tree, then end with a verdict line citing the receipt id it returns."

운영자 결정 d-20261106T162918Z-afc8은 이 회차에 Codex를 더 부르지 말라고 했고, 영수증 가드가 막으면 우회하지 말고 보고하라고 했다. 그래서 훅이 요구한 codex_audit/audit_multi 호출을 하지 않았다. 훅을 통과시키려고 판정을 FAIL로 바꾸지도 않았다. 코드 판정은 위 근거대로 PASS로 둔다. 이 SHA에 대해 영수증 가드가 받아들이는 PASS 기록은 없다. 여기서 감사를 멈추고, 처리는 리더와 운영자의 결정에 맡긴다(운영자 승인 아래 Codex를 한 번 더 부를지, 또는 다른 처리를 할지).
