# t586 F1 재현 기록 — S2 가드 뮤턴트 생존과 새 앵커

- 카드: t586 / SPEC: SPEC-INIT-TUX-I18N-001 (3회차 감사 F1)
- 측정 트리: `268cffe2cb47107c8fcf720301df974c04e383e5` (워크트리 `.claude/worktrees/t586`)
- 방법: 3회차 감사 판정 파일 E-4 와 같다. 제품 트리는 건드리지 않고 스크래치 디렉터리에 `internal/cli/profile_setup.go` 사본을 두고, 이 트리로 빌드한 테스트 바이너리를 각 사본 디렉터리에서 실행했다. 가드는 실행 시점 작업 디렉터리의 `profile_setup.go` 를 읽는다(`profile_setup_nested_test.go:28`).

## Claim

1. 기존 S2 가드 `TestTUINestedConfigNoParallelWriter` 는 `persistProjectConfig` 호출만 지운 사본에서도 통과한다. 같은 파일의 함수 정의 줄이 문자열 검사를 계속 참으로 만들기 때문이다. 감사자의 관측이 재현된다.
2. 같은 뮤턴트를 잡은 것은 S9 `TestWizardCarriesStoredSegmentsIntoPrefs` 다. 원본 사본에서 S9 는 통과한다.
3. 이 방법은 뮤턴트를 잡을 수 있다. S3 뮤턴트는 S3 가드를 실패시킨다.
4. SPEC v0.2.3 의 새 S2 앵커(주석이 아니고 `func persistProjectConfig(` 정의가 아닌 `persistProjectConfig(` 호출 줄이 1줄 이상)는 원본 사본에서 호출 줄 하나를 찾고, 호출만 지운 사본에서는 0줄을 낸다. 그 사본에도 정의 줄은 남아 있다.

## Evidence

테스트 바이너리 빌드:

```
go test -c -o <scratch>/t586-f1/cli.test ./internal/cli   → TESTBUILD_EXIT=0 (f1-build-cli-test.txt, 빈 출력)
```

사본과 뮤턴트:

```
diff internal/cli/profile_setup.go <scratch>/t586-f1/ctl/profile_setup.go     → DIFF_CTL_EXIT=0
diff internal/cli/profile_setup.go <scratch>/t586-f1/s2mut/profile_setup.go   → DIFF_S2MUT_EXIT=1
520c520
< 			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
---
> 			if err := error(nil); err != nil {
diff internal/cli/profile_setup.go <scratch>/t586-f1/s3mut/profile_setup.go   → DIFF_S3MUT_EXIT=1
456c456
< 	if permissionMode == defaultPermissionMode {
---
> 	if permissionMode == "acceptEdits" {
```

각 사본 디렉터리에서 `cli.test -test.run '^<이름>$' -test.v -test.count=1`:

| 사본 | 가드 | 종료 코드 | 출력 | 증거 파일 |
|---|---|---|---|---|
| ctl | S2 `TestTUINestedConfigNoParallelWriter` | 0 | `--- PASS` | `f1-repro-ctl-s2.txt` |
| s2mut | S2 `TestTUINestedConfigNoParallelWriter` | 0 | `--- PASS` (뮤턴트 생존) | `f1-repro-s2mut-s2.txt` |
| ctl | S9 `TestWizardCarriesStoredSegmentsIntoPrefs` | 0 | `--- PASS` | `f1-repro-ctl-s9.txt` |
| s2mut | S9 `TestWizardCarriesStoredSegmentsIntoPrefs` | 1 | `profile_setup_removed_questions_test.go:131: persistProjectConfig must be called with an empty convention so git-convention.yaml is preserved` / `--- FAIL` | `f1-repro-s2mut-s9.txt` |
| s3mut | S3 `TestPermissionModeNormalizeAcceptEdits` | 1 | `profile_setup_nested_test.go:86: profile_setup.go must normalize acceptEdits permission mode to empty (REQ-WC10-014)` / `--- FAIL` | `f1-repro-s3mut-s3.txt` |

모든 실행에서 `=== RUN` 이 1줄 나와 선택된 테스트 수는 0 이 아니다.

처음 네 실행을 병렬로 돌렸을 때 S9 실행 한 건에 디렉터리 지정이 빠져 있어 어느 사본을 읽었는지 확정할 수 없었다. 그래서 S9 는 `cd <사본> && pwd` 로 디렉터리를 출력한 뒤 ctl·s2mut 각각 다시 실행했고, 표의 S9 두 줄이 그 결과다.

새 S2 앵커(SPEC `design.md` §10, `acceptance.md` AC-ITI-010 (2)·(4) 의 검색식을 그대로 사용):

```
grep -n -F 'persistProjectConfig(' <copy> | grep -v -E '^[0-9]+:[[:space:]]*//' | grep -v -F 'func persistProjectConfig('

ctl   → ANCHOR_CTL_EXIT=0
520:			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {       (f1-anchor-ctl.txt)
s2mut → ANCHOR_S2MUT_EXIT=1, 0줄                                                       (f1-anchor-s2mut.txt, 빈 파일)
grep -n -F 'func persistProjectConfig(' <scratch>/t586-f1/s2mut/profile_setup.go
      → 160:func persistProjectConfig(projectRoot, devMode, convention string) error {   DEF_STILL_PRESENT_EXIT=0
```

## Baseline-attribution

모든 측정은 이번 실행에서 HEAD `268cffe2c` 의 `internal/cli/profile_setup.go` 사본과, 같은 트리로 `go test -c` 한 테스트 바이너리로 했다. SPEC v0.2.3 수정분은 `.moai/specs/` 문서만 바꾸며 Go 파일은 바꾸지 않는다.

## Gaps

- 새 앵커는 아직 가드 코드로 구현되지 않았다. 4번 주장은 SPEC 이 정한 검색식을 사본에 적용한 결과이고, 가드 테스트가 새 앵커로 바뀐 뒤 실패하는 모습은 run 단계 M5 에서 확인할 몫이다.
- 흡수 뒤(t583 병합 이후) `profile_setup.go` 구성에서는 재현하지 않았다.

## Residual-risk

- 검색식은 한 줄 단위다. 호출이 여러 줄에 걸쳐 `persistProjectConfig(` 와 인자가 갈라지는 형태로 바뀌어도 첫 줄에 호출 토큰이 남으므로 잡힌다. 호출을 변수에 담은 함수 값으로 부르는 형태로 바뀌면 `persistProjectConfig(` 토큰이 사라져 앵커가 0줄이 되므로, 가드는 결함을 놓치는 것이 아니라 거짓 경보로 실패한다(범위 한정 재감사 `plan-audit-iter4-scoped.md` 지적으로 정정).
- 호출을 지우고 `persistProjectConfig(` 를 줄 끝 주석이나 문자열 리터럴에 남기는 뮤턴트는 앵커를 통과한다. 앵커는 줄 전체 주석만 거른다. 옛 앵커도 같은 약점을 가졌고, 명세된 AC 뮤턴트에는 영향이 없다(같은 재감사 O1).
