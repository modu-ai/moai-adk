auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
score: 0.89
audited_sha: 470547fa2b1d0eeec4d1617ec75bfc2f6c3e0176
card: t2127

# t2127 sync 감사 판정

- 카드: t2127 (Class B, SPEC 없음), 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2127`, 브랜치 `WT-pr-codex-repair-2`
- 감사 대상: HEAD `470547fa2b1d0eeec4d1617ec75bfc2f6c3e0176`, 기준 `04d8c27a7ed77afc09217f8623519806d48df3c4`(= local `main`)
- 판정: **PASS-WITH-DEBT**, 종합 **0.89** (가중 89.5/100). 차단(blocking) 결함 없음. 부채 1건(D1, Low).
- 감사 시작·종료 시 `git status --short` 빈 출력, `git rev-parse HEAD` 동일. 감사 중 쓴 임시 탐침 `internal/hook/zz_audit_probe_t2127_test.go`는 실행 직후 삭제했고, 삭제 뒤 `git status --short` 빈 출력을 확인했다.

## 점수표

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | Codex 지적 형태(`git commit -m 'card work' # cd <primary>`)가 기존 식 true → 새 식 false로 바뀜(탐침 P01). 범위 테스트 318 RUN / 318 PASS / 0 FAIL / 0 SKIP |
| Security (25%) | 85/100 | PASS | 거부→허용 방향 변화 2건(P02·P03)을 재현했으나 둘 다 이미 문서화된 주석 생략 잔여(SPEC-GUARD-COMMENT-SCAN-001 §F)를 물려받은 난독 형태. Critical/High 없음 |
| Craft (20%) | 90/100 | PASS | RED 관측(3건 FAIL, exit 1) 기록, 대조군 7건 의미 있음, `gofmt -l`·`go vet` 무출력, `golangci-lint` `0 issues.` |
| Consistency (15%) | 95/100 | PASS | `normalizeCommandForScan`과 같은 순서(따옴표 축약 → 주석 생략)를 재사용, 새 헬퍼 없음, 변경은 함수 한 줄 + 주석 |

필수 통과 차원(Functionality, Security) 모두 기준 충족.

## Claim

1. 수정이 Codex P2 지적(branch_guard.go:621 @04d8c27a7e)을 정확히 해소한다.
2. 실제 `cd <primary>`는 일반적인 셸 형태 전부에서 여전히 대상으로 잡힌다. 다만 난독 형태 2종에서 거부→허용 변화가 생겼고, 이는 기존에 수용된 잔여의 상속이다(D1).
3. 테스트는 RED가 관측되었고 대조군이 의미 있다.
4. diff 범위는 3개 파일이다.

## Evidence

### E1. diff 범위

```
$ git diff --stat 04d8c27a7e..HEAD
 .moai/reports/t2127/progress.md                    | 58 ++++++++++++++++++++++
 internal/hook/branch_guard.go                      |  7 ++-
 .../hook/branch_guard_cd_comment_t2127_test.go     | 53 ++++++++++++++++++++
 3 files changed, 116 insertions(+), 2 deletions(-)
```

코드 변경은 한 줄이다: `plain := substituteQuotedArguments(command)` → `plain := substituteShellComments(substituteQuotedArguments(command))` (함수 주석 4줄 추가).

### E2. 지정 범위 테스트 (이 감사에서 직접 실행)

명령: `go test ./internal/hook/ -run 'TestCdPrimaryTarget|BranchGuard|ProtectedCommit|Comment' -count=1 -v -timeout 20m` (로그는 세션 scratch 파일)

```
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	83.973s
=== RUN 개수: 318
--- PASS 개수: 318
--- FAIL 개수: 0
--- SKIP 개수: 0
--- PASS: TestCdPrimaryTargetIgnoresCdInShellComment (0.83s)
    --- PASS: .../trailing_comment
    --- PASS: .../comment_after_quoted_message
    --- PASS: .../whole-line_comment_on_a_later_line
    --- PASS: .../cd_before_commit
    --- PASS: .../cd_after_commit
    --- PASS: .../real_cd_then_comment
    --- PASS: .../hash_inside_quotes_then_cd
    --- PASS: .../cd_inside_quotes_is_data
    --- PASS: .../mid-word_hash_then_cd
    --- PASS: .../cd_to_worktree
```

스윕 집합은 비어 있지 않다(318건, 신규 10개 하위 테스트 포함).

### E3. RED 증거 (레인 기록 `red.txt`, 04d8c27a7e + 미추적 테스트 파일)

```
--- FAIL: TestCdPrimaryTargetIgnoresCdInShellComment (4.46s)
    --- FAIL: .../trailing_comment (1.09s)
    --- FAIL: .../comment_after_quoted_message (0.14s)
    --- FAIL: .../whole-line_comment_on_a_later_line (0.17s)
    --- PASS: .../cd_before_commit ... (대조군 7건 PASS)
FAIL	github.com/modu-ai/moai-adk/internal/hook	5.385s
exit=1
```

독립 재확인: 아래 E4 탐침이 기존 식(`substituteQuotedArguments`만 거친 스캔)을 같은 트리에서 재구현해 돌렸고, 주석 전용 형태 P01·P08·P12·P13·P17에서 `old=true`를 관측했다. RED는 올바른 이유(주석 안 `cd`를 대상으로 읽음)로 빨갛다.

### E4. 변이 탐침 (임시 `_test.go`, 실행 후 삭제)

명령: `go test ./internal/hook/ -run TestZZAuditProbeT2127 -count=1 -v` → `ok ... 5.442s`. `shell` = POSIX sh에서 실제 `cd <primary>`가 실행되는지, `commitVerbVisible` = `matchProtectedCommitCommand`가 커밋 동사를 보는지.

```
P01 codex single-quoted msg + comment cd               shell=false old=true  new=false commitVerbVisible=true
P02 quote-adjacent # then cd, commit next line         shell=true  old=true  new=false commitVerbVisible=true  <-- DIVERGES FROM SHELL
P03 backslash continuation no space then # line        shell=true  old=true  new=false commitVerbVisible=true  <-- DIVERGES FROM SHELL
P04 heredoc body comment then real cd                  shell=true  old=true  new=true  commitVerbVisible=true
P05 heredoc body names cd (data)                       shell=false old=true  new=true  commitVerbVisible=true  <-- DIVERGES FROM SHELL
P06 escaped hash then cd                               shell=true  old=true  new=true  commitVerbVisible=true
P07 comment inside $( ) then cd                        shell=true  old=true  new=true  commitVerbVisible=true
P08 comment hides separator                            shell=false old=true  new=false commitVerbVisible=true
P09 quoted hash as its own word                        shell=true  old=true  new=true  commitVerbVisible=true
P10 mid-word quoted hash                               shell=true  old=true  new=true  commitVerbVisible=true
P11 ${#x} then cd                                      shell=true  old=true  new=true  commitVerbVisible=true
P12 tab-led comment                                    shell=false old=true  new=false commitVerbVisible=true
P13 indented comment line                              shell=false old=true  new=false commitVerbVisible=true
P14 quote-adjacent # same line (verb also elided)      shell=true  old=true  new=false commitVerbVisible=false <-- DIVERGES FROM SHELL
P15 real cd with trailing comment containing &&        shell=true  old=true  new=true  commitVerbVisible=false
P16 line-start #cd then commit                         shell=false old=false new=false commitVerbVisible=true
P17 pipe then comment                                  shell=false old=true  new=false commitVerbVisible=true
P18 hash after ) (shell: comment)                      shell=false old=true  new=true  commitVerbVisible=true  <-- DIVERGES FROM SHELL
```

해석:
- P01·P08·P12·P13·P17: 지적된 오탐 계열이 모두 해소(old=true → new=false, 셸과 일치).
- P04·P06·P07·P09·P10·P11·P15: 실제 `cd`가 heredoc 뒤, `\#` 뒤, `$( )` 안 주석 뒤, 따옴표 속 `#` 뒤, `${#x}` 뒤, 뒤따르는 주석과 함께 있어도 계속 잡힌다(회귀 없음).
- P05·P18: 셸보다 많이 잡는 쪽(안전 방향), 변경 전과 동일 — 신규 아님.
- P14: 같은 줄의 커밋 동사도 같이 생략되므로 보호 커밋 판정 자체가 변경 전부터 일어나지 않는다 — 결과 변화 없음.
- **P02·P03: 거부→허용 변화.** 커밋 동사가 다음 줄에 보이고(`commitVerbVisible=true`), 기존 식은 `cd`를 잡아 primary로 판정(거부)했지만 새 식은 놓쳐 hook-cwd(워크트리) 판정으로 떨어진다. 원인은 `substituteShellComments`의 문서화된 잔여(branch_guard.go:336-342): 따옴표 축약이 `"a"#b`를 `a X #b`로 만들어 단어 중간 `#`이 단어 시작처럼 보이는 경우, 그리고 공백 없는 `\`-줄바꿈 이어쓰기 다음 줄의 `#`. → D1.

### E5. 정적 검사

```
$ gofmt -l internal/hook        → 무출력, exit 0
$ go vet ./internal/hook/       → 무출력, exit 0
$ timeout 300 golangci-lint run --timeout 5m ./internal/hook/...
0 issues.
```

golangci-lint는 이번에는 거부되지 않고 실행되었다.

### E6. 감사 계획과 Codex 2차 의견 (1회 사용)

`moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, `cross_model_active: true`, `enforced_required: ["codex"]`(claude off, glm off).

`mcp__moai__codex_audit` (native, target baseBranch, delta_base 04d8c27a7e):

(Format-only repair by the lane: the quoted Codex key `verdict:` is renamed `codex-verdict:` so the evidence parser reads a single decision key; the value is verbatim.)

```
codex-verdict: pass
summary: 지정 기준 대비 변경 3개 파일을 정적으로 검토했으나 지적할 결함은 찾지 못했네. 지침대로 테스트·빌드·재현 명령은 실행하지 않았네.
findings: []
resolved_model: gpt-6.1-sol
review_base: delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)
audit_receipt: rcpt-315bad0b1c3c30fb58c11c0c
```

Codex는 D1을 지적하지 않았다(정적 검토만). D1은 이 감사의 탐침에서만 관측되었다.

## Findings

- **D1 [Low] [optional — 부채로 기록] [in-diff, 재현 있음]** `internal/hook/branch_guard.go:623` — `cdPrimaryTarget`이 `substituteShellComments`를 물려받으면서 그 함수의 문서화된 가림(blinding) 잔여도 함께 물려받았다. `echo "a"#b; cd <primary>` + 다음 줄 `git commit -m x`(P02), `echo x\` + 다음 줄 `# ; cd <primary>` + 다음 줄 `git commit`(P03)은 셸에서 primary에 커밋되지만 새 코드는 hook-cwd로 판정한다(변경 전에는 거부). 비차단 사유: 둘 다 의도적으로 꾸민 난독 형태이고, 이 가드의 원칙("의도적으로 난독화한 형태의 과소 매칭은 옳은 실패 방향", branch_guard.go:201-206)과 SPEC-GUARD-COMMENT-SCAN-001 §F가 같은 입력 계열의 잔여를 이미 수용했으며, 같은 파이프라인이 커밋 동사 매처에서도 같은 잔여를 가진다(P14). 근본 수리는 이미 코드 주석이 지목한 `substituteQuotedArguments`(따옴표 축약이 단어 시작 `#`을 만들지 않게) 쪽이다. — Required fix(부채 처분): 후속 카드로 `substituteQuotedArguments`의 단어 경계 보존을 다루거나, 최소한 `cdPrimaryTarget` 주석에 이 상속 잔여를 한 줄로 명시. 이 카드에서 막지 않는다.
- **N1 [Info] [optional] [in-diff]** 새 테스트에 D1 잔여를 고정하는 문서용 케이스가 없다(현 동작을 잔여로 못박는 대조군이 있으면 향후 근본 수리 시 뒤집힘이 보인다). 반영 여부는 재량.

차단 결함 없음. Repair stop rule 해당 없음(첫 감사).

## Baseline-attribution

- 측정 트리: `470547fa2b1d0eeec4d1617ec75bfc2f6c3e0176`(HEAD), 감사 중 작업 트리 깨끗(임시 탐침 파일만 생성 후 삭제).
- E2·E4·E5는 이 감사에서 이 트리에 대해 직접 실행한 `go test`/`gofmt`/`go vet`/`golangci-lint` 출력이다. `go test`는 트리 소스를 직접 컴파일하므로 설치 바이너리 지연과 무관하다.
- E3 RED는 레인이 기록한 `red.txt`(04d8c27a7e + 미추적 테스트)를 읽은 것이며, 독립 재현은 E4 탐침의 `old` 열(기존 식 재구현)로 대신했다.
- 기준 SHA: `04d8c27a7ed77afc09217f8623519806d48df3c4` = `git rev-parse main`.

## Gaps

- 커버리지 수치 미측정(`go test -cover` 전체 패키지는 레인 기록상 1000초 이상). 변경된 한 줄은 신규 10개 하위 테스트가 직접 실행한다.
- 전체 `internal/hook` 패키지와 저장소 전체 테스트는 이 감사에서 재실행하지 않았다(레인 기록 `green-full.txt`: 4242 RUN, 0 FAIL — 레인 주장, 이 감사에서 미재관측). 전체 판정은 CI 몫.
- RED 자체(레인의 원래 FAIL 실행)는 이 감사에서 원래 형태로 재실행하지 않았다(코드 편집 금지). 탐침의 기존 식 재구현으로 갈음.
- `moai verify audit-plan`과 `codex_audit`를 실행한 moai 빌드는 `04d8c27a7e`로, 감사 대상 HEAD `470547fa2`의 조상이다(build_lag 경고). 감사 계획 설정 해석과 Codex 리뷰 전달에만 쓰였고, 변경된 hook 코드는 이 빌드로 판단하지 않았다(코드 판단은 E2·E4의 트리 직접 컴파일).
- Windows에서는 신규 테스트가 `skipWindowsPOSIXCommandText`로 건너뛴다(POSIX 셸 텍스트 테스트). Windows 동작은 관측하지 않았다.

## Residual-risk

- D1의 두 형태 외에도, 따옴표 축약이 단어 경계를 바꾸는 다른 꾸민 형태가 있을 수 있다(예: `$'...'#`). 모두 같은 근본 원인(`substituteQuotedArguments`)으로 귀결된다.
- 가드는 opt-in·fail-open 권고 장치이므로, 이 잔여는 보안 경계 우회가 아니라 조정 신호의 누락이다. 그래도 거부→허용 방향이라 후속 처분 기록이 필요하다.

## 반복 이력

- 1회차(이 판정): PASS-WITH-DEBT 0.89 @470547fa2b.

AUDIT-VERDICT 근거 영수증: rcpt-315bad0b1c3c30fb58c11c0c
