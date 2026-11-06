auditor-model: claude-opus-5-5

# t2087 sync 감사 — 4회차 판정 (N1 수리 델타 + Codex 게이트 마무리)

verdict: PASS-WITH-DEBT
audited_sha: cbf7e3808fd34df391499f04eb35ad561174e491

- 카드: t2087 (Class B, SPEC 없음)
- 점수: 93/100 (3회차 코드 판단을 이번에 다시 확인한 범위 안에서 유지)
- 감사 범위: `a6663e9417..cbf7e3808f` (N1 시험 수리). HEAD `4ed7d468bc`는 여기에 3회차 판정 파일만 더한 커밋이다.
- Codex 영수증: `rcpt-bfde429bbdf8794f4737febe` (verdict `pass`, findings 0)

## 요약

3회차에서 FAIL로 남은 까닭은 Codex 게이트 하나였다. 서버가 실행 전에 `stale_server`로 거부해 리뷰가 돌지 않았다. 사용자가 MCP 서버를 재연결했고, 이번 호출은 설치본(v3.2.1-rc.2, `04d8c27a7e`)이 받아 실제로 리뷰를 실행했다. 결과는 `pass`, 지적 0건이다.

이번 실행에서 다시 확인한 것: 델타 범위(시험 파일 하나와 2회차 판정 파일뿐, 파서 변경 없음), 빌드, vet(네이티브와 `GOOS=windows`), gofmt, 집중 시험. 모두 통과했다. 수리된 시험은 파일이나 디렉터리를 만들지 않고 Windows 제외도 없다.

판정은 PASS-WITH-DEBT다. 이번 델타에 차단 결함은 없고, 앞 회차에서 넘어온 부채(F4·F5·F6·O1·O2)가 남아 있다. 가장 큰 남은 한계는 Windows 실행을 어느 회차에서도 관찰하지 못했다는 점이다(컴파일 확인뿐).

## 결함 목록

| id | 심각도 | 분류 | 위치 | 내용 | 필요한 수정 |
|---|---|---|---|---|---|
| G1 | (해소) | 3회차 게이트 미충족 | MCP 서버 | 재연결 뒤 Codex 리뷰 실행, `pass`, 영수증 `rcpt-bfde429bbdf8794f4737febe`. | 없음 |
| N1 | (해소) | 2회차 차단 결함 | `internal/spec/lint_gitbatch_test.go:502-553` | 이번에 다시 확인: 시험 본문에 파일·디렉터리 생성, Skip, GOOS 분기 없음. | 없음 |
| O2 | Info | optional, in-delta, 3회차 측정 | `internal/spec/lint_gitbatch_test.go:510` | 합성 스트림은 세 번째 US를 머리 줄에 두지만 실제 git은 다음 줄에 둔다. 머리 줄 한 줄 경로만 탄다. | 차단 아님. 원하면 머리 줄을 두 줄로 나눈다. |

이번 델타에서 새로 생긴 차단 결함: 없음. 반복 수리 중단 규칙에도 닿지 않는다.

이월 부채(이번 델타의 결함 아님, pre-existing):
- debt: F4 (Low) `batchedOwnershipTransition`의 root별 pathspec — `_archive` 경로 이름 변경에서 Info 출력이 per-SPEC 경로와 다르다.
- debt: F5 (Info) 집합 단위 이력 단순화와 `--follow`의 차이(잔여 위험).
- debt: F6 (Low) 배치가 실패하면 per-SPEC 경로로 조용히 대체된다(결과는 맞지만 느리다).
- debt: O1 (Low, optional) 같은 패키지의 다른 시험 파일들이 git 백그라운드 유지보수를 끄지 않고 저장소를 만든다.
- debt: O2 (Info, optional) 위 표.

베이스라인 위반 `OwnershipTransitionInvalid` recorded 0 -> current 1은 이번 델타가 건드리지 않은 기존 결함이고 t2105가 소유한다(파서·베이스라인 JSON·워크플로 모두 델타 밖).

## 차원별 판단

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 집중 시험 exit 0. 3회차 돌연변이 결과(F3 이전 파서에서 새 시험 실패)는 같은 SHA에 대한 것이라 유지. Codex도 결함 없음. |
| Security (25%) | 95 | PASS | 델타는 시험 파일뿐이다. 새 입력 경로·비밀값·외부 호출 없음. Codex 지적 0건. |
| Craft (20%) | 88 | PASS | 파일 생성 없음, Windows 제외 없음, build/vet/gofmt 통과. O2와 오라클 대조가 빠진 점은 감점. |
| Consistency (15%) | 92 | PASS | 같은 파일의 시험 이름·주석 관례를 따른다. |

가중합: 0.4×95 + 0.25×95 + 0.2×88 + 0.15×92 = 93.15 → 93.

## 증거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### 0. 트리 확인 (시작 시, Codex 호출 직전, 호출 직후 세 번)
- `git rev-parse --show-toplevel` → `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`
- `git branch --show-current` → `WT-spec-lint-baseline-2`
- `git rev-parse HEAD` → `4ed7d468bc9f8051d803834e540f97fdaf83713f` (세 번 모두 같음)
- `git status --short` → 빈 출력 (세 번 모두)

### 1. 델타 범위
- `git diff --stat a6663e9417 cbf7e3808f` (exit 0):
  ```
   .moai/reports/t2087/sync-audit-verdict-iter2.md | 121 ++++++++++++++++++++++++
   internal/spec/lint_gitbatch_test.go             |  70 +++++++++-----
   2 files changed, 168 insertions(+), 23 deletions(-)
  ```
- `git diff --stat a6663e9417 cbf7e3808f -- internal/spec/lint_ownership_batch.go` → 빈 출력 (exit 0). 파서는 바뀌지 않았다.
- `git diff --stat cbf7e3808f HEAD` (exit 0):
  ```
   .moai/reports/t2087/sync-audit-verdict-iter3.md | 142 ++++++++++++++++++++++++
   1 file changed, 142 insertions(+)
  ```
  감사 대상 SHA와 HEAD의 차이는 보고서 파일 하나뿐이므로, HEAD 트리에서 돌린 검사는 `cbf7e3808f`의 코드에 대한 검사다.

### 2. N1 수리 재확인
- `lint_gitbatch_test.go` 502~553행 읽기: 시험은 `strings.Join`으로 만든 스트림을 `parseOwnershipBatch(strings.NewReader(stream))`에 넣고 `b.resolve(".moai/specs/SPEC-A-001/spec.md")`만 단언한다.
- `grep -n -E 'func TestLintGitBatch_QuotedPathHeaderClosesSegment|os\.|MkdirAll|WriteFile|TempDir|newGBRepo|Skip|GOOS'` 결과를 490~560행으로 거른 출력: `502:func TestLintGitBatch_QuotedPathHeaderClosesSegment(t *testing.T) {` 한 줄뿐. 파일 생성·Skip·GOOS 분기 없음.

### 3. 기존 검사 (이 트리, Go `go1.27.3 darwin/arm64`, 부하 124~155)
- `go build ./...` → exit 0, 출력 0바이트.
- `go vet ./internal/spec/...` → exit 0, 출력 없음.
- `GOOS=windows go vet ./internal/spec/` → exit 0, 출력 없음.
- `gofmt -l internal/spec` → 출력 없음, exit 0.
- `go test -count=1 -run 'TestLintGitBatch|Ownership|ImpliedStatus|StatusGitConsistency|ConstantGitLog|Unreachable' ./internal/spec/` → `ok  	github.com/modu-ai/moai-adk/internal/spec	122.402s`, exit 0 (스크립트 안에서 `echo "exit=$?"`로 받은 실제 종료 코드).
- 판단 빌드(§2.2): 시험·vet·빌드는 이 트리 소스를 컴파일한 것이다. `moai verify audit-plan`, `moai decision read`는 설치본 v3.2.1-rc.2로 돌린 절차 확인이다.

### 4. 감사 계획
- `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2087` → exit 0; `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude·glm `"gate": "off"`, `"cross_model_active": true`.

## Codex 감사 (운영자가 허용한 호출)
- 근거: 카드 결정 `d-20261106T071854Z-643f`("수리 범위만 3회차 허용 (권장)", 시험 수리 델타만), 표준 규칙 `d-20261106T080039Z-305c`(리뷰 시작 전 `stale_server` 등으로 거부돼 영수증이 없는 호출은 횟수에 넣지 않음). 3회차 호출은 영수증 없이 `stale_server`로 거부됐으므로 이번 호출이 허용된 1회다.
- 넘긴 인자(정확히): `project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`, `target=baseBranch`, `mode=native`, `delta_base=a6663e9417718fcd3481c91aa573aefec4617d4d`. `model`·`focus`는 넘기지 않았다.
- 호출 횟수: 정확히 1회. 재시도 없음. `audit_multi`·`glm_audit`·`claude_audit`·app-server 경로는 부르지 않았다.
- 결과(결정적인 부분 그대로):
  - `"verdict":"pass"`
  - `"findings":[]`
  - `"audit_receipt":"rcpt-bfde429bbdf8794f4737febe"`
  - `"build_commit":"04d8c27a7e"`
  - `"review_base":"delta_base (merge base a6663e9417718fcd3481c91aa573aefec4617d4d)"`
  - `"resolved_model":"gpt-6.1-sol"`, `"requested_effort":"high"`, `"error_code":""`
  - summary: `Verdict: pass ... 지정 기준 이후 변경을 정적으로 검토했으나 새로 도입된 결함은 찾지 못했네. 지시에 따라 테스트·빌드·재현 명령은 실행하지 않았네.`
- 주의: Codex는 정적 검토만 했고 시험을 실행하지 않았다고 스스로 밝혔다. 시험 실행은 위 증거 3이 맡는다.
- 리뷰 기준 `delta_base...HEAD`에는 3회차 판정 파일(HEAD 쪽 보고서)도 들어간다. 코드 변경은 `cbf7e3808f`의 시험 파일 하나다.
- 해결되지 않은 Codex 차단 지적: 없음.

## 반복 이력
- 1회차(`140c1ca406`): FAIL 77. F1(Windows PATH 심), F2(`diff.relative`) 차단, F3 권고, F4~F6 부채. Codex `rcpt-265f061edb178a80d58392da`.
- 수리 `fb2c9dbb37`(F1·F2·F3), `a6663e9417`(플레이크).
- 2회차(`a6663e9417`): FAIL 84. 수리에서 생긴 N1(따옴표 디렉터리, Windows 금지). Codex `rcpt-4bf94c4b99b843a5d0c8b2e4`.
- 수리 `cbf7e3808f`: 시험을 합성 스트림 입력으로 바꿈(시험 파일만).
- 3회차(`cbf7e3808f`): 코드 93, N1 해소·F3 방어력 돌연변이로 확인. Codex 호출이 `stale_server`로 거부돼 영수증 없음 → FAIL.
- 4회차(`cbf7e3808f`): 서버 재연결 뒤 Codex `pass`, 영수증 `rcpt-bfde429bbdf8794f4737febe`. 기존 검사 다시 통과 → PASS-WITH-DEBT.

## Gaps (관찰하지 않은 것)
- Windows 실행: 어느 회차에서도 관찰하지 않았다. `GOOS=windows go vet` 컴파일 확인과 정적 판단뿐이다. Windows에서 시험이 실제로 통과하는지는 CI나 Windows 기기에서만 확인된다.
- 도구 거부 1: 집중 시험을 백그라운드 명령으로 처음 보냈을 때 worktree 격리 가드가 거부했다("names git in a form too complex to verify"; 시험 이름 패턴 속 `Git` 글자 때문으로 보인다 — 추정). 같은 명령을 스크래치 스크립트 `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e87d25f1-2d2d-4f0d-b06b-4a890feb4763/scratchpad/focused.sh`(트리로 `cd` 후 실행)로 돌렸다. 결과는 실제 실행이다.
- 3회차의 돌연변이 시험과 실제 git 출력 비교는 이번에 다시 돌리지 않았다. 같은 SHA(`cbf7e3808f`)에 대한 3회차 측정을 인용한다.
- 실행하지 않은 검사: `./internal/spec/...` 전체(지시로 제외), golangci-lint, 커버리지, CI.
- 감사 계획은 `cross_model_active: true`라 정의상 `audit_multi`가 기본 경로지만, 운영자 결정 범위(Codex 1회)와 지시에 따라 부르지 않았다. claude·glm 게이트는 `off`다.
- 원래 시험이 하던 배치 대 per-SPEC 오라클 대조는 새 시험에서 빠졌다(3회차와 같음).

## Residual-risk (관찰했는데도 남는 위험)
- 합성 스트림은 `ownershipBatchFormat`을 손으로 흉내 낸다. 형식 문자열이 바뀌면 시험과 파서를 함께 고쳐야 한다.
- 시험은 부하 124~155에서 돌았다. 시간에 민감한 시험은 CI에서 다르게 나올 수 있다.
- 스크래치 출력(`focused-test.txt`, `build.txt`, `vet*.txt`, `plan.json`)은 이 기기에만 있다. 판정을 가른 줄은 위에 옮겼다.

## 운영 메모 (미검증)
- `measured`: MCP 서버는 v3.2.1-rc.2 `04d8c27a7e`로 떠 있고 이번 호출의 `build_commit`도 `04d8c27a7e`다.
- `assumption`: 첫 백그라운드 명령 거부 원인이 시험 이름의 `Git` 글자라는 것은 확인하지 않은 추정이다. 확인하려면 같은 명령에서 패턴만 바꿔 보면 된다.
