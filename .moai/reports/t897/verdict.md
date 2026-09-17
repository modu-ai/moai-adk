# t897 판정문 — origin/develop CI 적색 2건 수리

카드: t897 (Class B, Tier S) · 브랜치: `WT-ci-red-blockers` (base develop `ba09526be`) · 날짜: 2026-09-18

---

## 1. Claim (주장)

1. **SPEC Lint 실패는 해소됐다.** `.moai/specs/SPEC-GTD-AUTONOMY-001/{plan,acceptance,design}.md` 의 frontmatter 3행 `status: completed` 를 삭제해 `ArtifactStatusFieldForbidden` error 3건이 0건이 됐다 — 수리 전후 실측(§2 E1/E2).
2. **Race Test 실패는 타임아웃이며, 플래그 한 개로 다룬다.** `.github/workflows/ci.yml` 의 `test-race` 실행 스텝에 `-timeout 20m` 을 명시했다. 데이터 레이스 수리가 아니다.
3. 둘 다 다른 동작을 바꾸지 않았다: (1)은 `status:` 행 삭제만이고 `spec.md` 는 손대지 않았다, (2)는 플래그와 주석 추가만이며 다른 스텝·잡은 그대로다(§2 E3).
4. **(2)는 이 카드가 검증하지 못한다.** 워크플로 변경은 로컬에서 재현 불가이고, 판정은 push 후 CI 몫이다(§4 G1).

---

## 2. Evidence (증거 — 명령과 관측 출력)

**E1. 수리 전 (base `ba09526be`, 이 워크트리)**

```
moai spec lint
→ ERROR ArtifactStatusFieldForbidden …/SPEC-GTD-AUTONOMY-001/plan.md        3
→ ERROR ArtifactStatusFieldForbidden …/SPEC-GTD-AUTONOMY-001/acceptance.md  3
→ ERROR ArtifactStatusFieldForbidden …/SPEC-GTD-AUTONOMY-001/design.md      3
→ 3 error(s), 3151 warning(s)
```

error 3건은 전부 같은 규칙, 전부 같은 SPEC, 전부 frontmatter 3행이다.

**E2. 수리 후 (같은 트리, 같은 명령)**

```
moai spec lint
→ 0 error(s), 3151 warning(s)
```

경고 수는 baseline 안이며 판정 대상이 아니다 — 수리 전후 동일(3151)하지만, 이 값이 흔들려도 판정은 바뀌지 않는다.

**E3. 변경 범위 (diff 실측)**

```
git diff --stat
→ .github/workflows/ci.yml                        | 5 ++++-
→ .moai/specs/SPEC-GTD-AUTONOMY-001/acceptance.md | 1 -
→ .moai/specs/SPEC-GTD-AUTONOMY-001/design.md     | 1 -
→ .moai/specs/SPEC-GTD-AUTONOMY-001/plan.md       | 1 -
→ 4 files changed, 4 insertions(+), 4 deletions(-)
```

SPEC 파일 3개는 각각 `-1` 줄뿐 — 삭제 외의 변경이 없다는 기계적 근거다. `spec.md` 는 diff 에 없다.

워크플로 변경 본문:

```
-          go test -json -race -count=1 ./... > test-stream.json || rc=$?
+          # -timeout is per test binary, not per job: Go's 10m default kills
+          # internal/cli, whose -race runtime measures 546-1183s. 20m clears that
+          # ceiling while staying inside this job's timeout-minutes: 25 budget.
+          go test -json -race -count=1 -timeout 20m ./... > test-stream.json || rc=$?
```

**E4. 20m 선택 근거**

- 상한: 같은 잡의 `timeout-minutes: 25`(ci.yml `test-race`) — `-timeout` 이 잡 예산을 넘으면 잡이 먼저 죽어 `-timeout` 의 의미가 없다.
- 하한: `internal/cli` 의 `-race` 실측 구간 546~1183초(auto-memory `feedback_internal_cli_timeout_floor`, 리드 인용). 관측 상한 1183초 ≈ 19.7분 — 20m 은 그 위이고 25분 예산 아래다.
- `-timeout` 은 **테스트 바이너리 1개당** 적용되므로, 잡 전체 벽시간이 아니라 가장 느린 패키지 기준으로 고른다.

---

## 3. Baseline-attribution (이번 런 귀속)

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t897`, 브랜치 `WT-ci-red-blockers`, base develop `ba09526be` (t813 병합 반영본).
- E1·E2 는 이 트리에서 이번 런에 실행한 `moai spec lint` 의 출력이다. 리드가 인용한 CI 실행 번호(`35264881210`, `35264881202`)의 로그는 이 카드가 직접 읽지 않았다 — 리드 관측이다.
- E1 의 error 3건과 CI 가 보고한 error 3건이 같은 것이라는 근거는 개수·규칙명·대상 파일 일치이며, 로그 대조는 하지 않았다(§4 G2).

---

## 4. Gaps (미검증)

- **G1 — 워크플로 수리는 로컬 재현 불가.** `-timeout` 추가가 Race Test 를 초록으로 만드는지는 이 트리에서 확인할 수 없다. GitHub Actions 러너에서만 관측되며, develop push 후 CI 판정에 맡긴다. 카드 지시대로 로컬에서 `-race` 전량을 돌리지 않았다 — 레인 머신 부하 회피이며, 측정 실패가 아니라 측정 안 함이다.
- **G2 — CI 로그 직접 대조 없음.** §3 에 적은 대로 CI 실행 로그는 읽지 않았다. 「DATA RACE 문자열 0건」도 리드 관측이며 이 카드가 재확인하지 않았다.
- **G3 — 경고 3151건.** baseline 안이라 범위 밖이며 손대지 않았다. 이 카드는 그것이 baseline 안이라는 사실도 직접 재지 않았다(리드 관측).
- **G4 — 다른 테스트 영향 없음 미측정.** 코드(`internal/`)를 바꾸지 않았으므로 Go 테스트를 돌리지 않았다. 판정 대상이 없어서이지 측정 실패가 아니다.
- **G5 — 20m 이 충분한지.** 하한 근거는 과거 실측 구간이고, 러너의 실제 `-race` 시간은 이 카드가 재지 않았다. 부족하면 CI 가 다시 타임아웃으로 알려 준다(이번에는 10분이 아니라 20분 지점에서).

---

## 5. Residual-risk (잔여 위험)

- **R1.** `-timeout 20m` 은 증상을 늦추는 값이지 `internal/cli` 의 `-race` 실행 시간을 줄이지 않는다. 그 패키지가 더 느려지면 같은 실패가 20분 지점에서 재발한다 — 근본 대응(패키지 분할·`-race` 범위 축소)은 이 카드 밖이다.
- **R2.** 잡 예산 25분과 `-timeout 20m` 사이 여유는 5분이다. 러너가 느린 날에는 잡 타임아웃이 먼저 걸릴 수 있고, 그때 증상은 「테스트 타임아웃」이 아니라 「잡 취소」로 다르게 보인다.
- **R3.** SPEC 3개의 `status:` 삭제는 lint 를 통과시키지만, 그 SPEC 의 실제 생애주기 상태는 `spec.md` 한 곳에서만 읽힌다. 다른 도구가 `plan.md` 의 상태를 읽고 있었다면 이 변경으로 조용히 값을 잃는다 — 그런 소비자가 있는지는 이 카드가 조사하지 않았다.
