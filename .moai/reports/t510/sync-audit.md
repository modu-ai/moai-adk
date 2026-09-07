# t510 sync-audit — SPEC-STATE-ANCHOR-001

감사자: sync-auditor (독립 판정) · 일자: 2026-09-07
판정 기준 트리: `.claude/worktrees/t510` @ `640baf6a3` (WT-state-write-locus, base `0b1e27877` + develop 흡수 `36b1aff8f`)
방법론: run/sync 보고(§E.2/§E.4)의 모든 주장을 무시하고 재관측. 스코프 테스트만 로컬 실행(전체 스위트 금지 준수), 감사자 쓰기는 본 판정서 1건뿐.

---

## Overall Verdict: **PASS — 95.6/100**

must-pass 방화벽(Functionality + Security) 양 차원 독립 통과. 차단(blocking) 결함 0. 선택(optional) 소견 3건(F1–F3).

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 97/100 | PASS | 아래 측정 로그 1–5항: AC 판정 테스트 6건 + 축 A 3형제 + canary 스윕 + AC-SA-012 grep 집합 + B5/B6 diff-0 전부 재관측 GREEN |
| Security (25%) | 95/100 | PASS | 아래 측정 로그 6항: usableSessionKey 유지, skip 경로 안전, 신뢰 모델 불변, template-embed 가드 유지 — 소견 F1(선택) |
| Craft (20%) | 93/100 | PASS | `go test -cover`: stateanchor **100.0%**, statusline **90.8%**, config 80.6%(선존 baseline — F2); `golangci-lint run` = `0 issues.`; `go vet` clean; `GOOS=windows go build` 통과; 감사 재실행 중 canary HOME 오염 0 (343→343) |
| Consistency (15%) | 96/100 | PASS | 카드 자체 접촉 정확히 17파일 = §E.3 선언 스코프와 일치; CHANGELOG 엔트리 내용이 코드와 사실 일치(builder.go:292, deps.go:178, 4가족, skip 경로); GH #1694 철회 문단이 B7 재판정과 코드로 일치; MX:ANCHOR+REASON+SPEC 시접 부착 |

---

## 측정 로그 (명령 + 관측 출력, this run @ `640baf6a3`)

### 1. 스코프 테스트 (Functionality — 환경 스크럽 복합 형태)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/stateanchor/ ./internal/statusline/ ./internal/config/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/stateanchor	0.889s
ok  	github.com/modu-ai/moai-adk/internal/statusline	15.851s
ok  	github.com/modu-ai/moai-adk/internal/config	2.861s
```

`TestResolveBacklogCounts_LatencyBudget` 플레이크 — 내 3회 statusline 패키지 실행 전부 무재현(알려-good 컨텍스트 대로 미평가).

### 2. AC 판정 명령 개별 재관측

```
$ go test ./internal/statusline/ -run 'TestContextUsageAnchorsToProjectDir|TestBoardRootResolvesThroughStateAnchor|TestGoalArmedReadsFromStateAnchor|TestNoProjectNoState|TestDisplaySegmentUnchangedByAnchorRepair' -count=1 -v
--- PASS: TestContextUsageAnchorsToProjectDir (0.00s)   ← AC-SA-001
--- PASS: TestNoProjectNoState (0.11s)                  ← AC-SA-005
--- PASS: TestDisplaySegmentUnchangedByAnchorRepair (0.00s) ← AC-SA-006
--- PASS: TestBoardRootResolvesThroughStateAnchor (0.19s)   ← AC-SA-002 (B2b 포함)
--- PASS: TestGoalArmedReadsFromStateAnchor (0.18s)     ← AC-SA-003

$ go test ./internal/config/ -run TestConfigCacheAnchorsToProject -count=1 -v
--- PASS: TestConfigCacheAnchorsToProject (0.15s)       ← AC-SA-004
```

### 3. 경계 가드 + sync 위생 (AC-SA-008, D3)

```
$ git diff 0b1e27877..HEAD --stat -- internal/hook/ internal/session/
(빈 출력 — develop 흡수분 포함 기준으로도 0)
$ git diff --name-only 15c6f5857..HEAD -- '*.go'
(빈 출력 — sync 2커밋 코드 무접촉)
$ git show 503aea79d --stat   → gh-1694-reply-draft.md + progress.md + spec.md(2행: status/updated) + CHANGELOG.md(+1)
$ git show 640baf6a3 --stat   → progress.md 1행 (sync_commit_sha 백필 — D3 승인 경로)
```

spec.md frontmatter: `status: completed`, `updated: 2026-09-07` — sync 커밋에서 in-progress→completed 전이 확인(소유자 규정 준수: status+updated만).

### 4. AC-SA-012 단일 시접 grep 집합 (전 6건 재관측)

```
1. grep -rn "resolveProjectDir\|resolveSessionDir" internal/statusline/
   → state_anchor_test.go:29 (RED 이력 주석) 유일 — 유산 제거 확인
2. grep -n "resolveStateAnchor" builder.go backlog.go state_anchor.go
   → 4매치: 정의(state_anchor.go:22) + B1(builder.go:181) + B3(builder.go:292) + B2(backlog.go:31)
3. grep -rn "Getwd" internal/statusline/*.go | grep -v _test
   → builder.go:439(표시 전용, 불변) + memory.go:73 + version.go:113 — 상태 경로 0
4. grep -rn "CurrentDir" … | grep -v _test
   → 표시 유도(builder.go:428-429) + 시접 입력(state_anchor.go:27)만
5. grep -n "stateanchor\." internal/cli/deps.go → :178 FromDirectory(cwd) — B4 동일 시접
6. stateanchor.go 내 Getwd = 0매치 — 시접 내부 cwd 앵커 부재
```

`extractProjectDirectory` 불변: `git diff 0b1e27877..HEAD -- internal/statusline/builder.go` 전체에서 함수명 0매치 + 발산 입력 골든 테스트 PASS (행위 이중 입증).

### 5. 축 A (AC-SA-009/010/011, regression-guard 채택 증거)

```
$ go test ./internal/cli/ -run 'TestTodoSweepSelectorMatchesFamily|TestGuardBypassMutant_ObserveHomePollution|TestTodoQueueRootGuard' -count=1 -v
selector TestTodo matches 137 test functions (swept-set liveness guard)
mutant pollution observed: 1 entr(ies) under canary HOME/.moai/todo
--- PASS: TestTodoQueueRootGuard_FiresOnLiveRepository / _SilentOnFixture / _SilentOnHomeFallbackFixture

$ MOAI_AXIS_A_CANARY_SWEEP=1 go test ./internal/cli/ -run 'TestAxisACanaryHomeSweep_TodoFamily' -count=1 -v -timeout 20m
sweep verdict: 198 todo tests ran under canary HOME …/TestAxisACanaryHomeSweep_TodoFamily1858186559/001 — 0 directories created under .moai/todo
--- PASS: TestAxisACanaryHomeSweep_TodoFamily (36.60s)
```

내 감사 재실행 전후 canary: `/bin/ls ~/.moai/todo | wc -l` = 343 → 343 (`001-*` 341 불변) — 판정서 수치와 일치, 감사 행위 오염 0.

### 6. 보안 렌즈 (코드 판독 + grep 프로브)

- `SessionTelemetryPath`/`usableSessionKey`(context_usage.go:62-81): 세션 키가 단일 파일명 요소가 아니면("") 거부 — REQ-ST-007 경로 탈출 거부 유지 확인.
- `writeContextUsage`(context_usage.go:176-214): projDir=="" 스킵(REQ-SA-003) + `isTemplateSourceDir` 가드 + write-if-changed throttle + atomic temp+rename + best-effort 무소음 — 전부 수리 후에도 유지.
- `FromDirectory`(stateanchor.go:89-98): `git.ResolveGitDirs` 오류 시 `""` 반환(스킵). `ResolveGitDirs`(internal/core/git/checkout.go:56)는 dir 인자 검증 + 오류 래핑, 셸 경유 없음 — 주입 표면 없음.
- B4 skip 보존: `cache.go` config-dir-exists 가드("Skip the cache write when the config directory itself does not exist", issue #1568) — 비프로젝트 cwd에서 `.moai` 화생성 없음, 코드 직독 확인.
- **적대적 `project_dir` 평가**: 리졸버는 `ProjectDir`를 무검증 반환하나, 신뢰 모델은 수리 전과 동일하다(같은 stdin 필드가 구 리졸버의 `current_dir` 앵커로 쓰였고, 표시 경로는 이미 project_dir 1순위). 쓰기/읽기는 모두 동일 사용자 파일시스템 내부로 제한되고 권한 경계를 넘지 않으며, 유출 능력은 구현 대비 증가하지 않는다(오히려 앵커 선택 집합이 축소됨). skip 남용으로 상태를 끌 수 있는 경로는 "쓰기 대상을 쓰기 불가 지점으로 보내는 것"뿐 — 기존 silent-failure 의미론과 동일한 실패 형태다. → F1(선택 경화 제안).

### 7. sync 산출물 사실 검증 (Consistency)

- CHANGELOG `### Fixed` 첫 항목(#1694 링크): 인용 좌표 전부 코드와 일치(builder.go:292 goal 읽기 / deps.go:178 config 캐시 / 4가족 / skip 경로 / throttle·silent-failure 보존 / B5·B6 diff 0 / 12 AC / canary 198·0 / 뮤턴트 증명 / 정리 레시피). B12: 인용 경로 5건 `ls` 실존, acceptance.md distinct AC = 12, `grep -c 'SPEC-STATE-ANCHOR-001' CHANGELOG.md` = 1(엔트리 본체뿐, 중복 없음).
- **GH #1694 철회 문단 판정: 정확하다.** "session-memo (7)은 이 결함이 아니다 — 훅 사슬이 컴팩트 시점에 프로젝트당 1회 기록, 자체 프로젝트 디렉터를 올바르게 해석" — 코드로 검증: `internal/hook/memo/writer.go` `Write(projectDir string, …)`(호출자 파라미터) + `internal/hook/path_resolve.go` `CLAUDE_PROJECT_DIR` 우선 → Getwd 폴백(`cwd_fallback:true` 마커, :87-88). 방문-디렉터 오염 아님 = B7 재판정(§E.2 M2)과 일치.
- 회신 초안의 gitignore 서술: 배포 템플릿 `internal/template/templates/.gitignore:238` = `.moai/state/`(내부 슬래시 → 리포 루트 고정, 중첩 stray 미포함) — 서술 정확. 정리 레시피("state/만 있으면 stray, config/ 있으면 실초기화 루트")는 보수적으로 올바른 판별식.

### 8. 스코프 드리프트 가드

`git diff --stat 36b1aff8f..HEAD -- internal/` (흡수 병합 이후 = 카드 자체 커밋) = 정확히 17파일 — §E.3 `total_run_phase_files: 17`과 일치. `git diff 0b1e27877..HEAD`에 보이는 codex_skills_prune / web·codexmirror / codexwiring·skills / catalog.yaml 등은 전부 develop 흡수 병합 `36b1aff8f`(타 카드 t506/t509/t526) 유입분으로, 본 카드 소관 아님을 커밋 그래프로 분리 확인.

---

## Findings

- **F1** [Low] [optional] `internal/stateanchor/stateanchor.go:64` — `Resolve`가 `ProjectDir`/`OriginalCwd`를 절대경로·존재 여부 무검증으로 반환한다. 적대적 stdin 페이로드가 상태 쓰기/읽기 앵커를 프로세스가 쓸 수 있는 임의 경로로 보낼 수 있다. 다만 신뢰 모델은 수리 전과 동일하고(같은 필드가 구 리졸버에도 흘렀음), 권한 경계 통과·유출 능력 증가는 없다. Required fix: 없음(승인된 설계). 선택 경화 — 앵커 후보에 `filepath.IsAbs` 요구 또는 디렉터 존재 검증을 후속 SPEC 후보로 기록.
- **F2** [Low] [optional] `internal/config` 패키지 커버리지 80.6% — 85% 문턱 미달이나 **선존 baseline**이다(본 카드는 config에 프로덕션 0행, 테스트 1파일만 추가; M3 수리 본체는 deps.go). 회귀 아님. 후속 카드 후보로 기록.
- **F3** [Info] [optional] 회신 초안의 "resolves its own project directory correctly"는 Getwd 폴백의 존재를 생략한 축약이다 — 폴백은 `cwd_fallback:true`로 표시되고 CLAUDE_PROJECT_DIR이 정당히 없는 운영자 명령 맥락에서 옳은 답(path_resolve.go:75 주석). 이슈 회신 문맥에서 허용 가능한 수준의 단순화로 판정.

차단 결함: **0건**.

---

## Gaps (관측하지 않은 것)

- **RED 셀 전부(C0/M2 2건/M3 2단) — attributed, not re-replicated.** §E.2의 RED 출력 전문은 run 커밋(592e2470e~1150d1f14 구간)에 귀속된 것이고, 재재현은 수리를 되돌리는 것을 요구하므로 읽기전용 감사 범위 밖이다. GREEN 측 재관측 + 삭제된 `resolveProjectDir`의 diff 삭제(-18행) 직독으로 RED의 그럴듯함은 뒷받침된다.
- Linux 빌드/테스트 — 레인 로컬 실행 안 함 규율에 따라 CI 매트릭스 판정 몫(§E.3 `pending-ci`와 동일).
- intended-change 테스트 파일 6건의 갱신은 diff-stat 크기 일치로만 검증했고 줄 단위 리뷰는 하지 않았다.
- spec-lint CoverageIncomplete 경고(#1696) — 알려진 오탐으로 알려-good 컨텍스트, 재판정 안 함.

## Residual-risk

- B3 가시성 변화(cd한 세션이 armed goal을 보게 됨)는 의도된 동작 변경 — cwd 범위 goal 읽기에 의존하던 외부 소비자가 있다면 새 동작을 관측한다. acceptance.md §AC-SA-003과 §E.2 M2에 기록돼 있다.
- 특수 git 토폴로지(서브모듈 워크트리, bare 인접 구성)에서 CommonDir 부모가 기대와 다른 루트를 가리킬 가능성 — 해석 실패 시 `""` 스킵으로 안전하게 무너지고(오염 없음), 정상 리포 형태는 stateanchor 테스트 7건이 고정.
- `TestResolveBacklogCounts_LatencyBudget` p95 스파이크 — 본 감사 3회 실행 무재현. 기계 부하 시 재발 가능성은 잔존(코드 무접촉, 기록된 환경 플레이크).

---

판정 권한 선언: 본 판정은 sync-auditor의 독립 관측이며, run 보고의 AC 매트릭스를 신뢰하지 않고 재관측해 도출했다. 재판정 대상 결함 델타는 위 Findings의 차단 0건 — 재감사가 필요하면 F1–F3 델타만 스코프로 한다.
