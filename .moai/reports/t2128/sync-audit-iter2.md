---
verdict: PASS
audited_sha: c5bc16ab2a9711f4e10abe98b8d3ed48173a661e
card: t2128
score: 0.93
iteration: 2
---

auditor-model: claude-opus-5-5

# t2128 sync 감사 2회차 — 판정 PASS (점수 0.93)

- 카드: t2128 (Class B 결함, SPEC 없음), 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2128`, 브랜치 `WT-pr-codex-repair-6`
- 감사 대상: HEAD `c5bc16ab2a9711f4e10abe98b8d3ed48173a661e` (1회차 감사 SHA `4e7aed6246` 이후 `8aff455aa8` 1회차 판정 파일, `fa2a30ca6e` F1 수리, `c5bc16ab2a` 진행 기록 §5)
- 평가 프로필: 1회차와 같은 `default` (Functionality 40 / Security 25 / Craft 20 / Consistency 15, must-pass = Functionality + Security)
- 범위: 1회차 지적 F1의 수리분(delta)과 패키지 시험·lint·Windows vet 재확인
- Codex: 이 단계 누적 2회(1회차 1회 + 이번 1회, 규칙 `d-20261106T080039Z-305c` 상한 도달), 이번 receipt `rcpt-eb59e780ddb40a6c320d00c0`, 판정 pass, 지적 0건

## 판정 요지

F1은 닫혔다. 수리 커밋은 51행과 54행 두 줄만 `_, _ = os.Stdout.WriteString(...)`으로 바꿨고, 자식 프로세스가 내보내는 `RESULT:` 줄의 내용과 형식은 그대로다. 1회차 이후 `internal/` 아래 다른 변경은 없다.

이번에는 CI와 같은 조합(golangci-lint v2.1.6, Go 1.26.8)으로 lint를 직접 돌렸다. 1회차 커밋에서는 errcheck 2건이 나오고 HEAD에서는 `0 issues`다. 그래서 1회차 때 남겨 둔 "CI 버전에서도 걸리는가" 공백이 양쪽 모두 메워졌다.

로컬 v2.15.0에서 나오는 govet `inline` 7건은 기준 커밋 `04d8c27a7e` 원본 사본에서도 같은 위치에 똑같이 나온다. CI 버전 v2.1.6의 govet에는 `inline` 분석기가 아예 없다. 이 카드와 무관하다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | `go test -count=1 ./internal/config/` → `ok ... 11.580s`(E1). 하위 프로세스 시험 `TestAgenticFactorySettings_RefCycle` PASS(E1). 즉 부모가 읽는 `RESULT:` 줄이 그대로 동작한다. 로직 변경 없음(E2) |
| Security (25%) | 92/100 | PASS | 1회차 이후 소스 변경 없음(E2). 1회차 근거(순환 DoS 제거, 값 미노출) 그대로 유효. 키 위치 secret 공백(F2)은 기존 결함 |
| Craft (20%) | 90/100 | PASS | CI 동등 lint(v2.1.6/go1.26.8) HEAD `0 issues`(E3). 로컬 v2.15.0은 카드 파일 지적 0건이고 govet 7건은 기준선과 같다(E4, E5). gofmt 출력 없음(E1) |
| Consistency (15%) | 92/100 | PASS | 저장소 관례 `_, _ =` 반환값 명시 폐기와 일치. 수리는 두 줄에 한정(E2) |

가중 조화평균: 0.40/95 + 0.25/92 + 0.20/90 + 0.15/92 = 0.0107805 → 92.8 → **0.93**. must-pass 두 차원 통과, BLOCKING 지적 없음.

## 지적 목록

- **F1 [CLOSED]** `internal/config/agentic_factory_settings_cycle_test.go:51`, `:54` — `fa2a30ca6e`에서 `_, _ = os.Stdout.WriteString(...)`로 수리됨. 로컬 v2.15.0과 CI 동등 v2.1.6 모두 이 파일 지적 0건(E3, E4).
- **F2 [MINOR] [optional] [touched-file, 기존 결함]** `internal/config/agentic_factory_settings.go:464` — 키 위치 alias/복합 키 안의 secret을 못 찾는다. 1회차 그대로, 수정 전 코드에서도 같다. 후속 카드 후보.
- **F3 [MINOR] [optional] [in-diff]** `internal/config/agentic_factory_settings.go:463-469` — 같은 route에서 순환이 secret보다 앞이면 오류가 cycle로만 보고된다. 거부는 되고 값 노출 없음. 정보성.
- **F4 [MINOR] [optional] [outside]** `webhook.routes`가 참조하지 않는 순환 앵커는 이 리더가 받아들인다(크래시 없음). 판정 영향 없음.
- **F5 [INFO] [optional] [outside, 기존]** `internal/config/audit_struct_yaml_symmetry_test.go:112`, `harness_ceiling_test.go:61`, `shipped_key_reader_test.go:369/588/625/706/777` — 로컬 golangci-lint v2.15.0의 govet `inline: Constant reflect.Ptr should be inlined` 7건. 기준 커밋 사본에서도 같다(E5). CI v2.1.6에는 해당 분석기가 없다(E6). 이 카드에 귀속되지 않는다. 로컬 lint 버전을 올릴 때 다룰 별도 정리 대상.

수리 중단 규칙: 이번 diff(`4e7aed6246..HEAD`)에서 생긴 BLOCKING 지적 없음 → repair-born 0건.

## 점검 항목별 결론

1. **F1 수리·자식 프로토콜·범위**: `git diff 4e7aed6246 HEAD -- internal/`은 두 줄 변경뿐이다(E2). 출력 문자열 `"RESULT:nil\n"`, `"RESULT:" + err.Error() + "\n"`은 바이트 단위로 같고 반환값 처리만 달라졌다. 부모 쪽 `strings.HasPrefix(line, "RESULT:")` 파싱(77~91행)도 손대지 않았다. 시험은 PASS다.
2. **Lint 귀속**: 로컬 v2.15.0 HEAD에서 7건, 기준 사본에서 같은 7건이 같은 줄·열에 나온다. CI 버전 v2.1.6의 govet 소스에는 `inline` 분석기가 없다. v2.1.6을 CI와 같은 Go 1.26.8로 빌드해 돌리면 HEAD는 0건, 1회차 커밋은 errcheck 2건이다. 그러므로 카드가 만든 lint 지적은 이제 없다.
3. **시험·vet**: 패키지 시험 `ok`, `GOOS=windows GOARCH=amd64 go vet` exit 0.

## Claim / Evidence / Baseline / Gaps / Residual-risk

### Claim

- C1: F1은 수리되었고, 1회차 이후 `internal/` 변경은 그 두 줄뿐이다.
- C2: 자식 프로세스 프로토콜은 바뀌지 않았고 하위 프로세스 시험이 통과한다.
- C3: CI 동등 lint(v2.1.6 + Go 1.26.8)는 HEAD에서 깨끗하고, 1회차 커밋에서는 F1을 잡는다.
- C4: 로컬 v2.15.0의 govet `inline` 7건은 기준 커밋에도 있는 기존 지적이며 CI 버전에는 해당 분석기가 없다.
- C5: 패키지 시험과 Windows 교차 vet은 통과한다.

### Evidence

**E1** — 트리 HEAD `c5bc16ab2a9711f4e10abe98b8d3ed48173a661e`, `GOCACHE=<scratchpad>/gocache-lane28`, go1.27.3 darwin/arm64

```
$ go test -count=1 ./internal/config/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/config	11.580s

$ GOOS=windows GOARCH=amd64 go vet ./internal/config/
vet-windows exit=0

$ go test -count=1 -v -run '^TestAgenticFactorySettings_RefCycle$' ./internal/config/
=== RUN   TestAgenticFactorySettings_RefCycle
--- PASS: TestAgenticFactorySettings_RefCycle (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.354s

$ gofmt -l internal/config/agentic_factory_settings_cycle_test.go internal/config/agentic_factory_settings.go
gofmt-exit=0          (출력 없음)
```

**E2** — 1회차 이후 변경

```
$ git diff --stat 4e7aed6246 HEAD
 .moai/reports/t2128/progress.md                    |  10 +
 .moai/reports/t2128/sync-audit-iter1.md            | 226 +++++++++++++++++++++
 .../config/agentic_factory_settings_cycle_test.go  |   4 +-
 3 files changed, 238 insertions(+), 2 deletions(-)

$ git diff 4e7aed6246 HEAD -- internal/
@@ -48,10 +48,10 @@ func TestAgenticFactorySettings_RefCycleChild(t *testing.T) {
 	_, err := LoadAgenticFactorySettings(root)
 	if err == nil {
-		os.Stdout.WriteString("RESULT:nil\n")
+		_, _ = os.Stdout.WriteString("RESULT:nil\n")
 		return
 	}
-	os.Stdout.WriteString("RESULT:" + err.Error() + "\n")
+	_, _ = os.Stdout.WriteString("RESULT:" + err.Error() + "\n")
 }
```

**E3** — CI 동등 lint. CI 설정(`.github/workflows/ci.yml:861` `go-version-file: go.mod` → `go 1.26.8`, `:865` `golangci-lint@v2.1.6`, `:883` `golangci-lint run --timeout=5m`)을 따라 v2.1.6을 `GOTOOLCHAIN=go1.26.8`로 scratch에 빌드해 실행

```
$ golangci-lint version
golangci-lint has version v2.1.6 built with go1.26.8 from (unknown, modified: ?, mod sum: "h1:LXqShFfAGM5BDzEOWD2SL1IzJAgUOqES/HRBsfKjI+w=") on (unknown)

[트리 HEAD c5bc16ab2a]
$ golangci-lint run --timeout=5m ./internal/config/
lint216-go126-head-exit=0
0 issues.

[1회차 커밋 4e7aed6246의 git archive 사본]
$ golangci-lint run --timeout=5m ./internal/config/
lint216-go126-iter1-exit=1
internal/config/agentic_factory_settings_cycle_test.go:51:24: Error return value of `os.Stdout.WriteString` is not checked (errcheck)
internal/config/agentic_factory_settings_cycle_test.go:54:23: Error return value of `os.Stdout.WriteString` is not checked (errcheck)
2 issues:
* errcheck: 2
```

**E4** — 로컬 golangci-lint v2.15.0, 트리 HEAD

```
$ golangci-lint version
golangci-lint has version 2.15.0 built with go1.27.2 from 25c5369 on 2026-10-16T10:51:54Z
$ golangci-lint run ./internal/config/
lint-exit=1
internal/config/audit_struct_yaml_symmetry_test.go:112:18: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/harness_ceiling_test.go:61:18: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:369:17: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:588:17: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:625:28: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:706:28: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:777:17: inline: Constant reflect.Ptr should be inlined (govet)
7 issues:
* govet: 7
$ grep -n "cycle_test\|errcheck" <lint output>
grep-exit=1           (일치 없음)
```

**E5** — 같은 v2.15.0, 기준 커밋 `04d8c27a7ed77afc09217f8623519806d48df3c4` 전체 트리를 `git -C <tree> archive`로 scratch(`<scratchpad>/t2128-base-i2`)에 풀어 실행

```
$ golangci-lint run ./internal/config/
lint-base-exit=1
internal/config/audit_struct_yaml_symmetry_test.go:112:18: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/harness_ceiling_test.go:61:18: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:369:17: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:588:17: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:625:28: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:706:28: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:777:17: inline: Constant reflect.Ptr should be inlined (govet)
7 issues:
* govet: 7
```

**E6** — golangci-lint v2.1.6 소스(`go mod download github.com/golangci/golangci-lint/v2@v2.1.6`, scratch 모듈 캐시)

```
$ grep -rn "passes/inline\|\"inline\"" .../golangci-lint/v2@v2.1.6/pkg/golinters/govet/
grep-inline-exit=1    (일치 없음)
$ grep -n "passes/" .../v2@v2.1.6/pkg/golinters/govet/govet.go | wc -l
      48
```

**E7** — Codex (`mcp__moai__codex_audit`, mode native, target baseBranch, delta_base `4e7aed62464fc88181a15b675ebc24dba3bcde26`; 호출 전 `git rev-parse --verify` 40자 확인, `merge-base --is-ancestor` exit 0, `git status --short` 0줄; 호출 뒤 HEAD 동일·0줄)

```
"verdict":"pass"
summary: 지정 기준 대비 변경을 정적 검토했으나 새로 도입된 결함은 찾지 못했어. 테스트·빌드·재현 명령은 실행하지 않았어.
findings: []
audit_receipt: rcpt-eb59e780ddb40a6c320d00c0
resolved_model: gpt-6.1-sol
review_base: delta_base (merge base 4e7aed62464fc88181a15b675ebc24dba3bcde26)
build_lag: the installed binary was built from commit 04d8c27a7, an ancestor of this tree's HEAD c5bc16ab2.
```

**E8** — 감사 계획

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2128
"config_status": "ok", "model": "codex", claude gate off, codex gate required, glm gate off,
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
```

### Baseline-attribution

- E1, E2, E4는 이 실행에서 트리 HEAD `c5bc16ab2a9711f4e10abe98b8d3ed48173a661e`(clean)에 대해 직접 돌렸다.
- E5는 기준 커밋 `04d8c27a7e`, E3의 둘째 실행은 1회차 커밋 `4e7aed6246`의 `git archive` 전체 트리 사본에서 돌렸다. 작업 트리에는 이 판정 파일 말고 아무것도 쓰지 않았다.
- `go`·`golangci-lint`는 트리 소스를 직접 컴파일·분석하므로 moai 바이너리 빌드 지연과 무관하다. `moai verify audit-plan`과 MCP 서버(codex_audit)는 설치 바이너리 커밋 `04d8c27a7e`(이 트리 HEAD의 조상)로 돌았다. 이 카드는 moai CLI 동작을 바꾸지 않으므로 계획 판독에 영향이 없다고 보지만, 지연 사실은 기록해 둔다.
- CI 동등 lint 바이너리는 scratch에서 `go install ...@v2.1.6`(GOTOOLCHAIN=go1.26.8)으로 빌드했다. CI도 같은 소스 설치 방식이다.

### Gaps

- CI 자체는 실행하지 않았다. E3은 CI와 같은 lint 버전·Go 버전·명령을 로컬에서 재현한 것이다. 다만 macOS arm64에서 돌렸고 CI는 ubuntu-latest라는 차이가 있다.
- go1.27.3으로 빌드한 v2.1.6은 export data 버전 불일치(`export data version 5 is greater than maximum supported version 2`)로 typecheck 235건을 내며 측정이 불가능했다. 그 결과는 쓰지 않았고 Go 1.26.8 빌드로 바꿨다.
- Windows에서 하위 프로세스 시험을 실제로 실행하지는 않았다(교차 vet만).
- `-race`, 전체 스위트, 다른 패키지는 실행하지 않았다(지시 범위). 이번 회차는 1회차 이후 delta 범위로 좁혔다. 1회차의 기능·보안 탐침(E4~E6 of iter-1)은 다시 돌리지 않았다. 그 사이 소스 변경이 없음을 E2로 확인했으므로 그대로 이어받는다.
- 리더 규칙 `d-20261106T080039Z-305c`에 따라 `audit_multi`는 쓰지 않았다. Codex는 이 단계 상한 2회를 모두 썼다(이번 1회).
- 감사 중 거부된 명령 1건(작업 트리 격리 가드): 모듈 캐시 탐색 `ls`·`grep`을 한 줄로 묶은 명령. 단순 명령으로 나눠 다시 돌렸다. 측정 대체는 없었다.

### Residual-risk

- F2의 키 위치 secret 공백은 남아 있다. 현재 리더는 `repository`만 소비하므로 정책 우회에 그친다.
- 순환이 아닌 아주 깊은 중첩은 재귀 깊이만큼 스택을 쓴다(1회차와 같음).
- 로컬 golangci-lint를 v2.15.0 이상으로 쓰는 개발자는 이 패키지에서 기존 govet `inline` 7건을 계속 본다. CI 버전을 올리면 그때 lint가 실패할 수 있다(이 카드와 무관).
- scratch의 출력 파일(`<scratchpad>/t2128-i2-*.txt`, 빌드 로그)은 기기 로컬이며 인용 대상이 아니다. 판정에 쓴 줄은 모두 위에 옮겨 적었다.

## 반복 이력

| 회차 | 감사 SHA | 판정 | 점수 | 요지 |
|---|---|---|---|---|
| 1 | 4e7aed62464fc88181a15b675ebc24dba3bcde26 | FAIL | 0.87 | F1 errcheck 2건(BLOCKING) 외 수정 정확 |
| 2 | c5bc16ab2a9711f4e10abe98b8d3ed48173a661e | PASS | 0.93 | F1 수리 확인, CI 동등 lint 0건, govet inline 7건은 기준선·CI 미해당 |

Codex receipt: 1회차 `rcpt-e703dc932cb9971ce8459362`, 2회차 `rcpt-eb59e780ddb40a6c320d00c0`.
