# t600 — 워크트리 이름 `../..` 가 기본 체크아웃 재사용으로 이어지던 결함

- 카드: t600 · [hooks 감사 2026-09-11 · H05 · P1]
- 레인: lane-4
- 워크트리: `.claude/worktrees/t600` · 브랜치 `WT-worktree-name-escape`
- 기준 트리: `eabce7444` (origin/develop, 리드 지정)
- 대상: `internal/hook/worktree_create.go`
- 증거 경로: **`.moai/reports/t600-worktree-name-escape/verdict.md`** (파견이 지정한 `.moai/reports/t600/verdict.md` 가 아님 — 아래 참조)

> **증거 경로 충돌 — 리드 판단 필요.** 파견은 증거 경로로 `.moai/reports/t600/verdict.md` 를 지정했으나, 그 경로에는 **같은 번호를 쓰던 다른 카드**(`WT-cli-prered-four`, internal/cli 선재 적색 4건)의 착지된 증거가 이미 추적 파일로 존재한다(커밋 `29c16a236`, `52a4de177`). 본 레인이 그 파일을 한 번 덮어썼고 즉시 `git restore --source=HEAD` 로 원상 복구했다(`git status --short` 에서 해당 경로에 수정 없음 확인). 이 카드의 증거는 충돌하지 않는 경로에 두었다. 파견 경로를 바꿀지, 옛 카드 증거를 옮길지는 리드 소관이다.

---

## Claim

1. 결함은 기준 트리 `eabce7444` 에 살아 있었고, 카드에 적힌 재현이 그대로 재현된다.
2. 이름 검증(세그먼트 단위), 경로 봉쇄 검사, 재사용 시 등록-워크트리 대조 세 겹을 추가해 결함을 막았다.
3. 합법 입력(단일 세그먼트, 다중 세그먼트 `feat/x`, 등록된 워크트리 재사용)은 계속 통과한다.
4. 가드 6축은 각각 뮤턴트로 무력화했을 때 대응 테스트가 실제로 빨간불이 된다 — 공허한 초록이 아니다.
5. `internal/hook` 패키지 전체 실행에서 남은 실패 1건은 이 변경에 귀속되지 않는다(선재 레드).

---

## Evidence

### E1. 기준 트리 재현 (Claim 1)

명령 (임시 저장소, `t.TempDir`, 경로 이탈 축이므로 임시 경로 한정):

```
go test ./internal/hook/ -run TestProbeT600TraversalReuse -v -count=1
```

출력 (발췌, 그대로):

```
    zz_probe_t600_test.go:18: worktree_name=../.. err=<nil> returned="/private/var/folders/.../TestProbeT600TraversalReuse1761393467/001" repo="/private/var/folders/.../TestProbeT600TraversalReuse1761393467/001" returned_primary_checkout=true
--- PASS: TestProbeT600TraversalReuse (1.05s)
```

`err=<nil>` 이고 반환 경로가 저장소 루트와 같다 — 격리를 요청한 에이전트에게 기본 체크아웃이 돌아갔다. 이 프로브 파일은 정식 테스트로 대체한 뒤 삭제했다.

경로: `filepath.Join(repoRoot, ".claude/worktrees", "../..")` 가 `repoRoot` 로 접히고, 이어지는 `os.Stat` + `IsDir` 재사용 분기가 "디렉터리가 존재한다"는 이유만으로 그것을 워크트리로 반환했다. 등록 여부는 검사하지 않았다.

### E2. 최종 그린 (Claim 2, 3)

```
go test ./internal/hook/ -run 'Worktree|ValidateWorktreeName|EnsureWithin' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	12.263s
```

쓸어 담은 개수 — 빈 스윕이 아님을 먼저 확인:

```
go test ... -v | grep -cE "^--- PASS"
55
```

정적 검사:

```
go vet ./internal/hook/          → exit 0, 출력 없음
golangci-lint run ./internal/hook/...  → 0 issues.
```

### E3. 뮤턴트 6축 (Claim 4)

각 축을 하나씩 무력화하고 해당 테스트만 실행 — 전부 원본 복원 후 `grep -c '__mutant_disabled__'` = 0 으로 잔재 없음 확인.

| # | 무력화한 것 | 실행 | 관측 결과 |
|---|---|---|---|
| M1 | `case ".", "..":` 거부 제거 | `-run 'TestValidateWorktreeName\|...RejectsTraversalName'` | FAIL: `parent_traversal` · `dot_segment` · `bare_parent` |
| M2 | `case "":` 거부 제거 | `-run TestValidateWorktreeName` | FAIL: `doubled_separator` · `trailing_separator` · `leading_separator` |
| M3 | `isWorktreeNameRune` 항상 true | `-run TestValidateWorktreeName` | FAIL: `shell_metacharacter` · `space` · `backslash_separator` |
| M4 | `ensureWithinWorktreeParent` 항상 nil | `-run TestEnsureWithinWorktreeParent` | FAIL: `the_parent_itself` · `sibling_escape` · `repository_root` |
| M5 | `ensureRegisteredWorktree` 항상 nil | `-run '...Rejects\|...Reuses'` | FAIL: `RejectsUnregisteredDirectoryReuse` |
| M6 | 심링크 분기 무력화 | `-run ...RejectsSymlinkReuse` | FAIL: `error = ... exists and is not a directory, want the symlink guard to reject it` |

**M5·M6 에서 드러난 것 — 축이 겹친다.** M1 에서 `RejectsTraversalName`(핸들러 경유 `../..`)은 통과했다. 이름 검증이 꺼져도 봉쇄 검사가 잡기 때문이며, 이는 의도한 이중 그물이 실제로 작동한다는 증거다. 다만 겹침은 개별 축을 단정하지 않은 채 지나가게 만들 수 있어, M5 에서 `RejectsSymlinkReuse` 가 여전히 초록인 것을 보고 그 테스트를 **어느 가드가 발화했는지까지 고정**하도록 조였다(`is a symlink` 문자열 단정). M6 은 그 조임 이후에 측정한 것이며, 조이기 전에는 이 축이 무력화돼도 초록이었다.

### E4. 남은 실패 1건의 귀속 (Claim 5)

패키지 전체 실행:

```
go test ./internal/hook/... -count=1
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.60s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	324.516s
ok   (하위 패키지 10개 전부)
```

실행 당시 머신 부하 58.56(1분). 부하가 12 로 내려간 뒤 그 테스트만 단독 재실행:

```
go test ./internal/hook/ -run TestSessionStart_DeferredScanDoesNotBlockReturn -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	1.231s
```

리드 보고에 따르면 lane-3·lane-7 이 각각 독립으로 base(eabce7444 이전 트리 포함)에서도 같은 실패를 실측해 develop **선재 레드**로 귀속했고 발행 목록에 올라 있다. 본 레인의 관측이 세 번째 독립 관측이다.

---

## Baseline-attribution

- 모든 측정은 워크트리 `.claude/worktrees/t600`, 브랜치 `WT-worktree-name-escape`, 기준 커밋 `eabce7444` 에서 이번 실행으로 수행했다.
- E1 은 수정 **이전** 트리(`eabce7444` 원본)에서, E2·E3 은 수정 **이후** 같은 트리에서 측정했다.
- 다른 패키지·다른 트리·다른 시점의 수치를 끌어다 쓰지 않았다.
- E4 의 "lane-3·lane-7 base 실측"은 본 레인이 직접 측정한 것이 **아니라** 리드가 전달한 타 레인 관측이며, 그 사실을 명시한다. 본 레인이 직접 관측한 것은 "단독 재실행 시 통과" 뿐이다.

---

## 설계 결정 — 카드 문구 대비 확대

카드의 개선 범위·완료 조건은 "경로 구분자 거부"라고 적혀 있으나, **세그먼트 단위 검증**으로 넓혔다. `/` 자체는 허용하고, 세그먼트마다 빈 값·`.`·`..`·허용 문자셋 밖 문자를 거부한다.

- **근거**: Claude Code 의 name 계약이 `/` 로 나뉜 다중 세그먼트를 합법으로 규정한다. 코드의 `sanitizeWorktreeBranchSuffix`(`/`→`-`)가 바로 그 지원 목적으로 존재하므로, `/` 전면 거부는 계약이 합법이라 한 입력을 거부하고 그 함수를 도달 불가능한 죽은 코드로 만든다.
- **결정 주체**: 운영자 결정(`AskUserQuestion` 로 확인), 리드 재확인. 리드 전언: "카드 문구와 계약이 부딪히면 계약이 이긴다".
- 봉쇄 검사를 이름 검증과 **독립된 두 번째 그물**로 둔 이유도 여기에 있다 — 이름 규칙이 나중에 느슨해져도 최종 경로가 부모 밖이면 여전히 막힌다.

## 기존 테스트 1건 변경

`TestWorktreeCreateHandler_ReusesExistingDirectory` 가 실패했다. 이 테스트는 `os.MkdirAll` 로 만든 **맨 디렉터리**가 재사용되는 것을 단정하고 있었다 — 즉 이번 카드가 없애라고 지시한 바로 그 동작을 계약으로 못 박고 있었다.

의도(멱등 재사용 + 레지스트리 1회 등록)는 보존한 채, 셋업을 핸들러 경유 실제 워크트리 생성으로 교체했다. 단정 내용은 그대로다.

---

## Gaps — 관측하지 않은 것

- **CI 매트릭스 미측정.** darwin/windows 교차 빌드와 전체 스위트는 돌리지 않았다(로컬 전체 스위트 금지 규율). 통합 판정은 CI 몫이다.
- **Windows 실동작 미측정.** 백슬래시 거부는 `validateWorktreeName` 단위 수준에서만 확인했고, 실제 Windows 런타임에서 훅이 도는 모습은 보지 않았다.
- **심링크 테스트의 조건부 skip.** 심링크를 만들 수 없는 환경에서는 `t.Skipf` 로 건너뛴다. macOS 에서는 실행됐음을 확인했으나, 다른 플랫폼에서 skip 되면 그 축은 그 실행에서 단정되지 않는다.
- **커버리지 수치 미측정.** `go test -cover` 를 돌리지 않았다.
- **실제 Claude Code 런타임 경유 미측정.** 모든 검증은 핸들러 함수 경계에서 했다. 런타임이 실제로 어떤 `name` 을 보내는지는 측정 대상이 아니었다(카드도 "정상 Claude 가 악성 이름을 생성한다는 증거는 없다"고 적고 있다).
- **`.moai/state/worktrees.json` 레지스트리 부작용 미검토.** 거부 경로에서 레지스트리가 건드려지지 않는다는 것은 코드상 자명하지만(`registerEntry` 는 반환 직전에만 호출) 별도 테스트로 단정하지는 않았다.

## Residual-risk — 관측했음에도 남는 위험

- **축 겹침이 다음 축을 가릴 수 있다.** M5·M6 에서 드러났듯, 그물이 여러 겹이면 한 겹이 죽어도 테스트가 초록일 수 있다. 이번에는 심링크 축을 발견해 조였지만, 같은 형태가 남아 있지 않다고 단정할 수는 없다.
- **`EvalSymlinks` 가 실패하면 재사용이 거부된다.** 권한 문제 등으로 해석에 실패하면 기존에 잘 돌던 워크트리도 거부될 수 있다. 안전한 방향(거부)으로 실패하지만, 재사용 가능한 트리를 막는 오탐은 가능하다.
- **`git worktree list` 호출이 재사용 경로에 추가됐다.** 훅 예산(5초)을 쓰는 서브프로세스가 한 번 늘었다. 로컬에서 재사용 테스트가 수 초 내 끝나는 것은 봤으나, 워크트리가 아주 많은 저장소에서의 지연은 측정하지 않았다.
- **선재 레드의 근본 원인은 미해결.** `TestSessionStart_DeferredScanDoesNotBlockReturn` 은 부하에 민감하며, 그 원인은 이 카드 범위 밖이고 별도 카드 소관이다.

---

## 적용한 정책 규칙

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1(관측하지 않은 주장 금지), §2(baseline 귀속), §3(5절 보고 형식) — E4 의 타 레인 관측을 본 레인 측정과 분리해 표기한 근거.
- `.claude/rules/moai/development/verification-completeness.md` §1.1(관측된 실패로만 완료), §1.1 빈 스윕 조항(E2 의 55 PASS 카운트), §2 뮤턴트 프로브(E3), §4(증거는 브랜치명이 아닌 트리 SHA 에 고정).
- `.claude/rules/moai/workflow/kanban-dispatch.md` § 검증 부하는 레인 국소 — 로컬 전체 스위트를 돌리지 않고 변경이 닿는 패키지만 측정, 전체 판정은 CI.
