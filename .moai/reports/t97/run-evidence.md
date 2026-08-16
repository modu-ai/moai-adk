# t97 — SessionStart 안내 개편 + D1 3단계 이행 — 실행 증거

Worktree: `.claude/worktrees/t97` · Branch: `WT-t97` · Base: `5c3141372` (origin/release/v3.1.1)

## 1. Claim (주장)

1. **D1 코드 이행**: `internal/kanban/bootstrap.go` `CompanionRoles`가 4역할(`plan, run, review, sync`)에서 3역할(`plan, run, sync`)으로 변경됐다. review 라벨(`review`, `review-1`, `review-tjlgt1`)은 더 이상 companion-shape으로 파싱되지 않는다.
2. **진입점 통일 안내**: `session_start_kanban_i18n.go`의 `glmSubstitute`가 "`moai cc -k` = Claude 백엔드 공장장 / `moai glm -k` = GLM 백엔드 공장장 — 런처가 백엔드를, `-k`가 칸반 역할을 정함 + GLM 치환법" 안내로 개편됐다.
3. **이름·백엔드 선택지 안내**: 새 `nameChoices` 필드가 "기본 Agent 작업자 / `judge` (Claude 판정 — GLM 공장장의 유일한 Claude 경로) / `worker-N`"을 안내하고, 4로케일(en/ko/ja/zh) 모두 채웠다.
4. lead notice의 launch 블록이 3라인(`plan/run/sync`)이 됐고, 빌더·주석·CLI help(`cc.go`/`glm.go`/`factory.go`/`kanban.go`)의 "four-role"/"plan -> run -> verify -> sync" 서술이 전부 3단계 표현으로 정리됐다.

## 2. Evidence (증거 — 명령 + 출력)

```
$ go build ./...
BUILD ALL OK

$ go vet ./internal/kanban/ ./internal/hook/ ./internal/cli/
VET OK

$ go test ./internal/kanban/
ok  github.com/modu-ai/moai-adk/internal/kanban  15.211s

$ go test ./internal/hook/        # 2차 실행 (GLM-phrase 검사 갱신 후)
--- FAIL: TestHomeJoinSiteCountIsPinned (0.17s)   ← 선결 결함, §4 Gaps 참조
FAIL  github.com/modu-ai/moai-adk/internal/hook  29.263s
  (t97 변경분 관련 실패 0건 — 1차 실패였던 TestKanbanBootstrapNoticeLead는
   테스트 갱신 후 재실행에서 PASS)

$ go test ./internal/cli/ -run "Kanban|ACFB"
ok  github.com/modu-ai/moai-adk/internal/cli  1.275s

$ golangci-lint run ./internal/kanban/... ./internal/hook/... ./internal/cli/...
0 issues.
```

`t97` 소속 테스트 갱신: `bootstrap_test.go`(역할 고정 테스트 3역일화 + retired-review 비파싱 케이스), `session_start_kanban_test.go`(정규식 3역할화·launch 3라인·retired-review fail-open 케이스), `session_start_kanban_i18n_test.go`(nameChoices 커버리지·프로토콜 토큰 `judge`/`worker-N`·레이아웃 3라인), `session_start_kanban_surface_test.go`, `session_start_factory_test.go`, `kanban_help_test.go`(plan->run->sync 서술 검사로 교체), `kanban_bootstrap_test.go`(살아있는 역할로 교체).

## 3. Baseline-attribution (baseline 귀속)

모든 측정은 본 워크트리(`.claude/worktrees/t97`, branch `WT-t97`)에서, base `5c3141372`(= origin/release/v3.1.1 HEAD)에 본 카드 diff 17파일(+125/−86)을 적용한 트리에 대해 이번 실행으로 관측했다.

## 4. Gaps (미검증)

- **`TestHomeJoinSiteCountIsPinned` 선결 실패 — t97 이전부터 존재**. 지적된 5개 파일(`internal/cli/{memory,migrate_profiles,tokens}.go`, `internal/cli/preference/cmd.go`, `internal/hook/session_end.go`)은 t97 diff와 무교집합이며, base 커밋에서 `git show 5c3141372:internal/cli/tokens.go | grep -c projects` → 5, `preference/cmd.go` → 4로 사이트가 이미 존재한다. 즉 다른 카드가 home-join 사이트를 추가하고 `wantSites`를 갱신하지 않은 것. t97은 수정하지 않음(스코프 밖) — 별도 카드 권장.
- **`internal/cli` 전체 스위트 미실행** — 로컬 부하 규율(레인별 타깃 테스트 → CI가 전수 판정)에 따라 칸반 관련(`-run "Kanban|ACFB"`)만 실행. 전체 판정은 release/v3.1.1 통합 후 CI 몫.
- **카드 범위로 명시적으로 미수정한 표면**: `internal/kanban/column.go` `ColumnReview`(보드 6컬럼의 review 컬럼), `internal/web/viewmodel_ops.go` `ChainRoles`(`lead, plan, run, review, sync`), 칸반 디스패치 룰 문서(`.claude/rules/moai/workflow/kanban-dispatch.md`의 review 컬럼 서술). 디스패치가 `CompanionRoles`와 SessionStart 안내만 범위로 지정했다. 보드 컬럼 폐지는 별도 카드가 필요하다.

## 5. Residual-risk (잔여 위험)

- **진행 중인 기존 run의 review 세션**: review 라벨이 companion-shape에서 빠졌으므로 그 세션의 SessionStart notice는 빈 문자열이 된다(fail-open 설계 — 에러 없음). 해당 세션의 Stop-hook block cap 등 companion 혜택도 더 이상 적용되지 않는다. 현재 v3.1.1 라운드에서 review 역할 세션을 띄운 run이 있다면 영향.
- **`judge` 이름은 안내만 존재**: `judge`를 인식하는 런처/hook 코드 경로는 이 카드에서 추가하지 않았다(디스패치가 안내 문구만 요구). `judge`라는 이름의 세션은 현재 일반 named session으로 동작한다.
- **보드 컬럼 review 잔존**: `moai todo` 보드와 웹 뷰에는 review 컬럼이 남아 있어, 안내(3단계)와 보드(6컬럼) 표면이 일시적으로 어긋난다 — 위 Gaps의 별도 카드로 정리 필요.
