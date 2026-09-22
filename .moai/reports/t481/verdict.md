# t481 — internal/cli 커버리지 억제 재현·측정 판정

- Card: **t481** (Class B investigation · Tier S)
- Branch: `WT-cli-cover-suppression`
- Worktree: `.claude/worktrees/t481`
- Base: local `develop` @ `8b391bc8c` (t477 `b2f98f6aa` 포함 — fast-forward 흡수. 배차 전제였던 "t477 착지 대기"는 레인 착수 시점에 이미 성립하지 않았음)
- 미푸시: 브랜치 전체 미푸시 (push 금지 준수). 측정 시점 카드 커밋 0 — 측정 트리는 `8b391bc8c`와 동일

## Claim

1. t237이 기록한 "실패 패키지는 `go test -cover`의 커버리지 출력을 억제한다"는 **기각** — go 1.26.4에서 거짓.
2. t477 착지 후 `internal/cli`의 커버리지는 **측정된다** — 17/17 패키지 `ok`, exit 0.
3. main `internal/cli` 커버리지는 **80.6%** — §6 목표(패키지 85%, critical 90%) **미달**. 이 카드는 보고 전용으로 수리하지 않음.

## Evidence

### (1) 메커니즘 — 실패 패키지도 수치와 프로파일을 모두 낸다

```
$ go -C .moai/cache/t481-mech test -count=1 -cover ./...   # go1.26.4 darwin/arm64
--- FAIL: TestMul (0.00s)
    bad_test.go:8: Mul(2,3) = 6, want 7 (deliberate failure)
FAIL
coverage: 100.0% of statements
FAIL	t481mech/bad	0.414s
ok  	t481mech/ok	0.727s	coverage: 100.0% of statements
FAIL
```

고의로 실패하는 패키지(`t481mech/bad`)가 패키지 FAIL 줄 **위 별도 줄**로 `coverage: 100.0% of statements`를 출력한다. `-coverprofile`에서도 실패 패키지의 프로파일이 병합된다:

```
mode: set
t481mech/bad/bad.go:4.24,6.2 1 1
t481mech/ok/ok.go:4.24,6.2 1 1
```

원문: `.moai/reports/t481/mechanism-plain.txt`, `mechanism-coverprofile.txt`, `mechanism-coverprofile-data.txt`.

> 측정 과정 기록: 첫 grep을 잘못된 패턴 `t481-mech`(하이픈 포함)으로 돌려 프로파일 0줄로 읽었고, 모듈 경로 `t481mech`로 바로잡은 뒤에 위 판독에 도달했다. 대조 기준(`t481mech/ok` 1줄)과 함께 재확인된 판독이다.

### (2) t237 원문 대조 — "출력 억제"는 부재의 관측이 아니라 도구 체인에 대한 단정이었다

트리 `32319621b` `.moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/progress.md:103`:

> **Unmeasured** — the package FAILs on the inherited red (see AC-PVM-010), and `go test -cover` prints no coverage line for a failing package. Named gap, not a claim.

같은 파일의 AC-PVM-010 행이 인용한 실제 관측: `FAIL github.com/modu-ai/moai-adk/internal/cli 462.872s` — 그리고 서브패키지 16개는 `ok …/internal/cli/worktree 10.857s coverage: 87.1%` 모양으로 기록. 즉 그들의 집계는 **`ok … coverage: X%` 붙은 줄**에서 수치를 읽었고, 실패 패키지의 수치는 그 모양이 아니라 FAIL 줄 위 별도 줄에 나오므로 요약에 잡히지 않았다. 요약에 수치가 없었던 것은 참이지만, 도구가 출력을 억제한 것은 아니었다 — (1)의 측정이 이를 직접 반증한다.

### (3) 착지 후 측정 — 17/17 ok, main 80.6%

```
$ go test -count=1 -cover -timeout 900s ./internal/cli/...   # 트리 8b391bc8c, exit 0 (백그라운드 태스크 출력 직독)
ok  github.com/modu-ai/moai-adk/internal/cli              438.373s  coverage: 80.6% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/agentlint     11.985s  coverage: 86.7% of statements
... (17개 전체 — 원문: .moai/reports/t481/cli-cover-post-t477-sweep.txt)
```

## 목표 판정 — CLAUDE.local.md §6

| 패키지 | 측정 | 85% (패키지 하한) | 90% (critical=cli) |
|---|---|---|---|
| `internal/cli` (main) | **80.6%** | 미달 (−4.4pp) | 미달 (−9.4pp) |
| internal/cli/harness | 80.9% | 미달 (−4.1pp) | 미달 |
| internal/cli/agentlint | 86.7% | 충족 | 미달 |
| internal/cli/preference | 85.8% | 충족 | 미달 |
| internal/cli/update | 88.9% | 충족 | 미달 |
| internal/cli/worktree | 87.1% | 충족 | 미달 |
| internal/cli/pr | 91.7% | 충족 | 충족 |
| internal/cli/taskledger | 92.7% | 충족 | 충족 |
| internal/cli/update/backup | 90.1% | 충족 | 충족 |
| internal/cli/update/deploy | 91.6% | 충족 | 충족 |
| internal/cli/update/merge | 92.1% | 충족 | 충족 |
| internal/cli/update/plan | 95.0% | 충족 | 충족 |
| internal/cli/update/report | 92.9% | 충족 | 충족 |
| internal/cli/wizard | 92.0% | 충족 | 충족 |
| internal/cli/printer | 97.0% | 충족 | 충족 |
| internal/cli/uikit | 98.8% | 충족 | 충족 |
| internal/cli/specid | 100.0% | 충족 | 충족 |

**핵심 수치**: main `internal/cli` **80.6%** — 85% 미달이며 90%(critical) 미달. 85% 하한 미달은 `harness` 80.9% 하나가 추가.

*해석 주석*: §6의 "critical packages (cli, template, hook): 90%+"를 internal/cli 하위 **전체** 패키지에 적용하면 미달 6개, 세 패밀리 **루트**에만 적용하면 main 80.6% 하나다. 두 해석을 모두 표기하며 어느 쪽인지의 판정은 리드/운영자 몫으로 남긴다.

## Baseline-attribution

- 메커니즘 실험: go1.26.4 darwin/arm64, gitignored 모듈 `.moai/cache/t481-mech/` (이 러닝에서 생성한 5파일).
- 스윕: 트리 `8b391bc8c` — local develop 팁과 동일 (fast-forward 흡수 후 카드 커밋 0 상태), 명령 `go test -count=1 -cover -timeout 900s ./internal/cli/...`, 종료코드 0은 백그라운드 태스크 출력의 `exit=0` 직독.
- t237 인용: 커밋 `32319621b`의 blob 직독 (`git show 32319621b:<path>`).

## Gaps (명시적으로 관측하지 않은 것)

- 억제 행동의 Go 버전별 이력(과거 버전에서 참이었는지)은 조사하지 않음 — 현재 툴체인 판정만. 단 t237의 run은 같은 날 같은 머신의 동일 go 1.26.4라, 그들의 런에도 이 기각이 그대로 적용된다.
- CI/청정 환경 재측정 없음 — 로컬 darwin/arm64 단일 실행.
- t237의 실제 집계 방식은 그들 산출물에 기록이 없어, "ok-line 집계 불일치"는 원문과 모순 없는 유력 가설로 표기한다 (측정된 원인이 아님).
- `-coverpkg` 등 다른 cover 플래그 조합은 미측정.

## Residual-risk

- 스윕 단일 실행 — 438s 루트 패키지 런의 환경 변동은 미측정.
- 커버리지 미달 자체의 수리는 이 카드 범위 밖 — 별도 축으로 보고.
- 저장소에 ok-line 집계 스크립트가 실재하는지는 확인하지 않음 (범위 밖).

---

보고 형식 근거: `verification-claim-integrity.md` §3 (5-section evidence-bearing report) + §2 baseline 귀속. 측정 판독 과정의 오독(프로파일 grep 패턴에 하이픈)은 결론 도달 전에 정정했고 그 과정을 §Evidence (1)에 기록했다. lane-10의 원문 기록은 삭제 대상이 아니라 인용-대조 대상이다 — 기록된 기록은 정정으로 대체하지 않고 나란히 둔다.
