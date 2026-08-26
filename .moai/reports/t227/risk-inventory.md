# t227 — 차단 복원 전 위험 선재고

카드 t227 이 [HARD] 로 요구한 "복원 직후 무엇이 막히기 시작하는지" 실측.
트리: `.claude/worktrees/t227`, 브랜치 `WT-deny-restore`, 측정 시각 2026-08-24.

## Claim — 수집 결함만 고치면 이 리포 328개 파일에 대한 Write 가 막힌다

**Evidence** — 배포 룰셋으로 리포 소스를 전수 스캔.

```
$ sg scan --json -c /Users/goos/MoAI/moai-adk-go/.moai/config/astgrep-rules/sgconfig.yml \
    internal/ pkg/ cmd/ > /tmp/t227-scan.json 2>/tmp/t227-scan.err
rc=1
$ cat /tmp/t227-scan.err
Error: 2065 error(s) found in code.
Help: Scan succeeded and found error level diagnostics in the codebase.
```

룰별 분해 (findings 총 2,996 중 error 심각도 2,065):

| ruleId | error findings | 영향 파일 |
|---|---:|---:|
| `go-error-ignored-blank` | **2,025** | **328** |
| `sec-command-injection-shell` | 23 | — |
| `sec-hardcoded-credential` | 13 | — |
| `sec-command-injection-exec` | 2 | — |
| `sec-weak-hash-md5` | 2 | — |
| **`go-error-ignored-blank` 제외 소계** | **40** | **17** |

`ShouldAlert` 는 error 카운트로 판정하므로, 수집 결함이 고쳐지는 순간 위 파일들의
내용을 담은 Write 는 deny 됩니다.

**Baseline-attribution** — 위 명령을 이 트리(`WT-deny-restore`, merge 커밋 `5eb53ada3`)에서
실행하고 그 출력을 파싱한 결과. 프라이머리 체크아웃의 로컬 dogfood 룰셋을 config 로 지정했고,
해당 룰셋은 템플릿 배포본과 동일 내용이다(t217 iter3 에서 확인).

## 원인 — 보안 룰이 아닌 스타일 룰이 error 심각도에 앉아 있다

2,025/2,065 = **98.1%** 가 `go-error-ignored-blank` 한 룰이다. 이 룰이 잡는 것은
`_ = foo()` — 이 코드베이스 전역에서 쓰는 Go 관용구이며 보안 결함이 아니다.

`go/` 디렉터리에서 error 심각도인 룰은 정확히 둘:

```
$ grep -rB4 "^severity: error" .../astgrep-rules/go/ | grep "id:"
  go-error-ignored-blank
  go-defer-in-loop
```

둘 다 보안 룰이 아니다. 심각도가 곧 **쓰기 거부**로 번역되는 게이트에서 스타일 룰이
`error` 를 다는 것은 분류 오류다.

## 남는 40건의 성격

`go-error-ignored-blank` 를 제외한 40건 17파일은 대부분 **테스트와 픽스처**다:
`internal/hook/branch_guard_test.go` 의 셸 문자열(가드 패턴 테스트용), `internal/cli/glm_*_test.go`
의 더미 토큰, `internal/astgrep/testdata/fixtures/`, 그리고 t217 이 만든
`internal/hook/security/testdata/scan-corpus/` — 마지막 것은 **걸리는 게 정상**이다
(deny 를 관측하기 위해 존재하는 픽스처).

프로덕션 코드 경로의 히트는 `internal/navigator/tiers/drift.go:75`
(`sec-command-injection-shell`) 1건이 눈에 띈다.

## 운영자 판정 (2026-08-24)

**수집 결함 수리 + `go-error-ignored-blank`·`go-defer-in-loop` 를 error → warning 으로 강등.**
잘못 분류된 것을 바로잡는 것이지 게이트를 느슨하게 하는 것이 아니다. 강등 후에도 두 룰은
warning 경로로 사용자에게 계속 보인다(warning 은 rc=0·stderr 빈 상태라 지금도 정상 동작).

강등 후 예상 노출: **40건 17파일**.

## Gaps

- 파일 스캔은 Write 페이로드의 근사다. deny 는 쓰기 **전** 내용에 걸리므로, 기존 파일 히트 수가
  곧 차단될 Write 수는 아니다. 다만 같은 파일을 다시 쓰면 같은 히트가 나므로 상한이 아니라
  현실적 추정으로 읽어야 한다.
- 이 리포 한 곳의 측정이다. 배포 사용자 코드베이스의 분포는 관측 범위 밖 — 다만
  `_ = foo()` 는 Go 전반의 관용구라 다른 Go 프로젝트에서도 같은 형태로 나타날 것으로 본다.
- `internal/navigator/tiers/drift.go:75` 가 진짜 결함인지 오탐인지는 판정하지 않았다. 이 카드
  범위 밖이며, 강등 후에도 error 로 남으므로 별도로 다뤄야 한다.

## Residual-risk

- 강등은 배포 룰셋 수정이므로 템플릿 미러 + 중립성 가드가 따라붙는다.
- 프라이머리 체크아웃의 `.moai/config/astgrep-rules/` 는 미추적 로컬 dogfood 사본이라 이
  브랜치가 갱신하지 않는다. 강등이 로컬에도 반영되려면 별도 동기화가 필요하다.
