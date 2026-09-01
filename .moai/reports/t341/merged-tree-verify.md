# t341 — merged-tree re-measurement (integration window)

카드 **t341** · SPEC-SELECTOR-CENSUS-001 · 브랜치 `WT-selector-census`
흡수 병합 커밋 **`ed3e0fc53`** (`origin/develop` `314a8410c` 24커밋 흡수)

병합 전 각자 초록이던 두 브랜치가 합쳐진 뒤에도 초록인지를 **병합 트리에서** 다시 잰 기록.
병합 전 측정을 병합 후 근거로 재사용하지 않는다.

## Claim

병합 트리 `ed3e0fc53` 에서 이 카드의 범위(`internal/hook`, `internal/telemetry`)가 통과하고,
SPEC Lint 가 error 0 이며, 이 검증 실행 자체가 0-실행이 아니다.

## Evidence

판정 도구 출처를 병합 트리로 고정한 뒤 측정했다. PATH 바이너리를 쓰지 않는다.

| # | 명령 | 결과 |
|---|---|---|
| 1 | `make build` | rc `0` · `BuildID=v3.1.2-901-ged3e0fc53` |
| 2 | `./bin/moai spec lint` | rc `0` · `0 error(s), 1091 warning(s)` |
| 3 | `go test ./internal/hook/... ./internal/telemetry/... -count=1` | rc `0` · 12패키지 전부 `ok` |
| 4 | `go test ./internal/hook/ -run TestZeroExecution -count=1 -v` | rc `0` · `--- PASS` 8 · `=== RUN` 19 |
| 5 | `go vet ./internal/hook/... ./internal/telemetry/...` | rc `0` · 출력 0줄 |
| 6 | `GOOS=windows GOARCH=amd64 go build ./...` | rc `0` · 출력 0줄 |
| 7 | `git status --porcelain` (make build 후) | 0줄 |

바이너리 출처 (명령 1):

```
v3.1.2   v3.1.2-901-ged3e0fc53   built 2026-08-31T05:34:16Z
```

`BuildID` 접미사가 병합 트리 HEAD `ed3e0fc53` 다. 명령 2 의 판정 규칙은 이 트리에서 컴파일된 것이다.

전문: 명령 3 은 `merged-tree-tests.txt`, 명령 4 는 `merged-tree-zeroexec.txt`, 명령 2 의 요약 줄은
`merged-tree-spec-lint-summary.txt` (전문 1095줄 338KB 는 이력에 넣지 않고
`.moai/state/verify/t341/merged-abs/spec-lint.txt` 에 둔다).

### 자기 적용 — 이 검증이 0-실행이 아님

이 카드가 세운 술어를 이 카드의 검증에 적용한다.

```
grep -c 'no tests to run\|no test files'  merged-tree-tests.txt      → 0
grep -c '^--- PASS'                       merged-tree-zeroexec.txt   → 8
grep -c '^=== RUN'                        merged-tree-zeroexec.txt   → 19
grep -c 'no tests to run'                 merged-tree-zeroexec.txt   → 0
```

12개 패키지 어디에도 `[no test files]` 가 없고, 셀렉터는 8 최상위 + 11 서브테스트를 실제로 쓸어담았다.
rc 0 이 "아무것도 안 돌았다" 를 감춘 실행이 아니다.

## Baseline-attribution

- 기준 트리: `ed3e0fc53` — 이 창 안에서 `git fetch origin develop` 후 흡수해 만든 병합 트리
- 흡수 대상: `origin/develop` = `314a8410c` (창 안에서 직접 판독. 창 밖 값이 아니다)
- 흡수 전 카드 head: `3cf724c14` · 발산 `24 2` → 흡수 후 `0 3`
- SPEC Lint `1091` 은 **로컬** 값이다. CI 는 얕은 체크아웃 계통 차분으로 `1072` 가 관측돼 왔다(t371 소관)
- 명령 1-7 은 전부 이 트리에서, 이 창 안에서 실행했다

## 충돌 해소 — CHANGELOG

`ort` 자동 병합이 `### Added` 최상단에서 실패했다(충돌 hunk 1개).

```
ours   (line 13)  - **[SPEC-SELECTOR-CENSUS-001](…)** — sync-phase close … (t341)
theirs (line 15)  - **Worktree-isolated sessions: which guard refused …** (card t287)
```

**양쪽 보존**했다. 어느 쪽도 버리지 않았고, 마커 3줄만 제거했다:
해소 전 1119줄 → 해소 후 1116줄 (정확히 `-3`), 마커 잔량 `grep -c` → 0,
두 항목 각각 `grep -c` → 1 / 1.

## Graph Freshness — 분모 두 개

도구의 `contribution:` 은 흡수 병합 후 first parent 가 자기 브랜치 head 라 **흡수분**을 센다.
자기 기여분은 따로 잰다.

```
git diff --name-only origin/develop...HEAD
  .moai/specs/SPEC-SELECTOR-CENSUS-001/progress.md
  .moai/specs/SPEC-SELECTOR-CENSUS-001/spec.md
  CHANGELOG.md
```

**3파일, `.go` 0건, codemaps 서술 대상 소스 0건.** 이 카드의 sync 기여분은 Graph Freshness 분자에
아무것도 더하지 않는다. codemaps 재생성은 하지 않았다(배치 끝 일괄, 운영자 판정).

두 수를 더하지 않는다.

## Gaps — 관측하지 않은 것

- **CI 미판독.** 이 브랜치는 미푸시다. `origin/develop` 의 CI 도 이 세션이 읽지 않았다
- **전체 스위트 미실행.** `go test ./...` 는 이 저장소에서 금지다. 전 패키지 판정은 CI 몫이다
- **`-race` 미실행.** `TestConcurrencyStress` 축은 이번에도 미측정이다(t372 상속 레드 소관)
- **sync-auditor 미실행**, `spec_audit` 미실행
- **@MX 검증 미수행.** 이 카드의 sync 커밋은 `.go` 무접촉이고, run-phase Go 표면을 재스캔하지 않았다.
  깨끗하다가 아니라 **안 했다**
- **AC-SEC-000 부채 그대로.** `.moai/reports/t341/live-payload.json` 부재, 질문 (b)(c) 미확정
- **흡수한 24커밋의 내용은 각 카드 소관**이다. 이 재측정은 그것들이 옳다고 말하지 않는다 —
  합쳐진 상태에서 내 범위가 붉지 않다고만 말한다
- **범위 밖 패키지 미측정.** `internal/cli`, `internal/spec` 은 흡수분이 건드렸으나 이 배치에서 재지 않았다

## Residual-risk

- 이름 충돌·파일 겹침 판독(창 전 사전 조사: 교집합 공집합)은 **컴파일 축 하나**다. 이번 명령 3 이
  그 축 너머(공유 픽스처·병렬성)를 실제로 덮었지만, 12패키지 rc 0 은 이 머신 이 시점의 값이다
- 로컬 `1091` 과 CI `1072` 의 19 차분은 계통으로 설명돼 있으나(t371), 이 트리에서 CI 를 돌려
  확인한 것은 아니다
- `spec-lint.txt` 전문은 gitignore 경로에만 있다. 워크트리 폐기 시 유실된다 — 요약 줄만 이력에 남는다
