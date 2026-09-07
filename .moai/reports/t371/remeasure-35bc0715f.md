# t371 흡수 후 재측정 — 인용 좌표 · 린트 기준값

- 측정 트리: `.claude/worktrees/t371` @ `WT-lint-shallow-clone`, HEAD **`35bc0715f`**
- 흡수: `git merge origin/develop` (`9328a5242`), 충돌 0, 91커밋
- 대조 기준: 흡수 전 HEAD `1e5199b88`
- 판정 도구: 트리 빌드 `./bin/moai` (`go build -o bin/moai ./cmd/moai`, rc=0). PATH 바이너리 미사용.

## 1. 인용 좌표 재측정 — 이동 4건 / 불변 7건

명령: `grep -n -F '<앵커>' <흡수전 사본> <현재 파일>` (앵커 = 코드 리터럴)

| 앵커 | `1e5199b88` | `35bc0715f` | 델타 |
|---|---|---|---|
| `lint.go` `Advisory: true, // heuristic` | 1316 | **1335** | +19 |
| `lint.go` `StatusGitConsistencyRule) Check` | 1287 | **1306** | +19 |
| `lint.go` §1.2 조용한 skip 블록(`if err != nil` → `return nil`) | 1305-1308 | **1324-1327** | +19 |
| `lint.go` `func applyEraDemotion` | 284 | **296** | +12 |
| `lint.go` `var eraDemotableCodes` | 260 | **272** | +12 |
| `lint.go` `if r.Strict && f.Severity` | 61 | 61 | 0 |
| `lint_ownership.go` `Code: "OwnershipTransitionInvalid"` | 430 | 430 | 0 |
| `drift.go` `cachedMainBranch` 호출부 | 68, 303 | 68, 303 | 0 |
| `gitquery_cache.go` `branch = "master"` | 103, 114 | 103, 114 | 0 |
| `cli/spec_lint.go` `✓ No findings` | 116 | 116 | 0 |
| `drift_characterization_test.go` `chdirForTest` / `setupDriftCorpusFixture` | 55 / 98 | 55 / 98 | 0 |

`lint.go` 총 행수 1323 → 1342 (+19). **이동은 전부 `lint.go` 한 파일에 국한**되고, 그 밖의 인용은 흡수 전후 동일하다.

부수 확인 — `git show-ref --verify refs/heads/main` → `48239c7dc…` (불변).
`grep -rn StatusGitUnreachable internal .github` → rc=1 (여전히 부재).

## 2. 린트 기준값 — 단위 정정 후 귀속

```
./bin/moai spec lint  →  0 error(s), 1096 warning(s)     (35bc0715f)
                          2 error(s), 1091 warning(s)     (1e5199b88, 보고서 자체 선언)
```

**종전 "1098"은 warning 수가 아니라 보고서 파일의 `wc -l` 값이었다.** 단위가 달라 비교 자체가 성립하지 않았다.

rule별 차분 (`awk '$1=="WARNING"{c[$2]++}'` 두 보고서 대조):

| rule | old | new | delta |
|---|---|---|---|
| `SyncSHASlotFormat` | 0 | 5 | **+5** |
| 나머지 9종 (`CoverageIncomplete` 846 · `MovingRefUnpinned` 114 · `LegacyEARSKeyword` 43 · `ModalityMalformed` 25 · `MissingExclusions` 24 · `StatusGitConsistency` 18 · `FrontmatterInvalid` 14 · `InvalidREQID` 6 · `OwnershipTransitionInvalid` 1) | — | — | 0 |

합계 검산: 1091 + 5 = 1096. ✓

사라진 error 2건은 **둘 다 `ArtifactStatusFieldForbidden`, 대상 `SPEC-INTEGRATION-LOCK-ATOMIC-001`** 의 `plan.md` / `acceptance.md`. 이 카드 소관이 아니며 흡수한 develop 쪽에서 수리됐다.

## 3. 18건 집합 동일성 — 개수가 아니라 집합으로 확인

```
diff <(sort statusgit-18-ids.txt) <(현재 린트에서 추출한 StatusGitConsistency SPEC ID 정렬)
→ 차분 없음 (18 = 18, 원소 동일)
```

`classification-18.md` 의 분류는 흡수 후에도 유효하다.

## 4. 리드 제기 전제 하나 반증 — t382 는 이 18건을 움직이지 못한다

리드 지시: *"t382(era.go H-3)가 착지하면 grandfathered 분류가 바뀌어 `StatusGitConsistency` 집합이 움직일 수 있다."*

코드가 그렇지 않다:

- `StatusGitConsistency` 는 **발화 지점에서 무조건** `Advisory: true` 로 나온다 (`lint.go:1335`). era 와 무관하다.
- `applyEraDemotion` (`lint.go:296-312`) 의 warning 분기는 `f.Advisory = true` 를 **설정**할 뿐이다. 이미 true 인 값에 다시 true 를 쓰는 것이라 상태가 바뀌지 않는다.
- `eraDemotableCodes` (`lint.go:272-275`) 는 `MissingExclusions` / `FrontmatterInvalid` 둘뿐 — `StatusGitConsistency` 는 없다.
- era 분류는 발화 **이후**에 적용되므로 finding 의 개수도 억제하지 못한다.

따라서 t382 의 H-3 수정이 어느 방향으로 가든 이 카드의 18건은 개수도 advisory 여부도 변하지 않는다. **병합 순서 제약의 근거로 쓸 수 없다.**

같은 논리가 이 카드가 신설하려는 `StatusGitUnreachable`(Info)에도 적용된다 — `applyEraDemotion` 의 switch 는 Error 와 Warning 만 다루고 Info 는 통과시킨다.

## 5. 잔여 위험

- `lint.go` 는 세 카드가 동시에 만진다(이 카드 M1 ≈ `:1324`, t382 `:272-275`, t376 rule 등록부 `:137`). 텍스트 충돌 가능성은 낮으나 **어느 쪽이 먼저 착지하든 나머지의 인용 행번호가 다시 밀린다.**
- AC-SLGB-011 은 여전히 착지 후 CI 로그로만 판정 가능하다.
