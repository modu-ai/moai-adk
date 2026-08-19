# t145 — 워크트리 잔존 감사 (2026-08-20)

카드: t145 (Class B — plan 생략, run → sync 인레인) · 트리 `.claude/worktrees/t145` · 브랜치 `WT-t145`

## 1. 요약

`git worktree list` 등록 **99개**, primary 제외 **98개**. 카드가 준 기준선(94개)보다 4개 늘었다 —
감사 착수 이후 만들어진 t143/t144/t145/t146 이다. 카드의 63/31 도 67/31 로 바뀌었다(그 사이 머지분 반영).

| 분류 | 개수 | 처분 |
|---|---:|---|
| A. 폐기 가능 (origin/main 병합 + 완전 청결 + 비앵커) | 52 | `dispose-merged.sh` 대상 |
| B. 회수 필요 (미추적/미커밋 잔여) | 12 | 산출물 회수 후 폐기 |
| C. 미병합 고아 (main·release 어디에도 없음) | 14 | **삭제 금지** — 유일 사본 |
| D. release/v3.1.1 병합 (배치 12 제외분) | 4 | 배치 PR 머지 후 폐기 |
| E. 배치 12 카드 (카드가 명시한 범위 제외) | 12 | 배치 PR 머지 후 폐기 |
| F. 라이브 세션 앵커 | 4 | **삭제 금지** — 세션 훅이 매 turn 깨짐 |
| 합계 | 98 | |

핵심 발견은 A가 아니라 **C**다. 14개 트리의 커밋이 `origin/main` 에도 `release/v3.1.1` 에도 없고,
그중 **13개는 원격 어느 브랜치에도 없다**(`git branch -r --contains` 결과 NONE). 즉 그 워크트리가 유일 사본이다.

## 2. [HARD] 삭제 금지 — 라이브 세션 앵커 4

감사 시점 실측(`lsof -d cwd`, claude 프로세스 9개):

| 트리 | 비고 |
|---|---|
| `t143` | 라이브 (감사 중 커밋 진행) |
| `t144` | 라이브 |
| `t146` | 라이브 |
| `t145` | 본 카드 트리 |

이 트리를 지우면 해당 세션의 훅이 매 turn 깨진다. 스크립트는 실행 시점에 cwd 를 한 번 더 확인해 건너뛴다.

## 3. 미병합 고아 14 — 유일 사본, 삭제 금지

`origin/main` 미병합 **그리고** `release/v3.1.1` 미병합. `remote` 열은 `git branch -r --contains <tip>`.

| 트리 | 브랜치 | tip | main 앞선 커밋 | remote | 비고 |
|---|---|---|---:|---|---|
| `agent-a8bf881ac39468a30` | `worktree-agent-a8bf881ac39468a30` | `52ad9255c` | 1 | NONE | |
| `agent-ae0671037b886c5a9` | `worktree-agent-ae0671037b886c5a9` | `216dadc2d` | 1 | NONE | |
| `card-astgrep-path` | `worktree-card-astgrep-path` | `813e98b2f` | 2 | NONE | |
| `card-clocal` | `worktree-card-clocal` | `efba57632` | 1 | NONE | |
| `card-provenance` | `worktree-card-provenance` | `a91bb670f` | 4 | NONE | 미추적 1건도 있음 |
| `card-wtiso` | `card-wtiso` | `b1f7118c0` | 13 | NONE | SPEC-FACTORY-CARD-ISOLATION-001 (메모리상 "미커밋 유일 사본"이었으나 현재 커밋됨) |
| `installer-fix` | `worktree-installer-fix` | `222c03d0b` | 3 | NONE | |
| `local-only-relocate` | `worktree-local-only-relocate` | `7f2019a12` | 1 | NONE | |
| `local-tree-cleanup` | `worktree-local-tree-cleanup` | `e585ef26e` | 2 | NONE | |
| `mermaid-audit` | `WT-mermaid-audit` | `b0db10bc7` | 1 | NONE | |
| `readme-redesign` | `worktree-readme-redesign` | `c57d24912` | 1 | `origin/worktree-readme-redesign` | 유일하게 push 됨 (README 4로케일) |
| `statusline-session-name` | `worktree-statusline-session-name` | `ff7983a3b` | 6 | NONE | |
| `t108` | `WT-t108` | `a15bfdc84` | 1 | NONE | |
| `web-shell-chrome` | `feat/web-shell-chrome` | `3aa383c1e` | 3 | NONE | |

처분은 트리 단위가 아니라 **카드 단위 판단**이 필요하다 — 통합할지, 폐기할지. 이 카드의 범위 밖이다.

## 4. 회수 필요 미추적 산출물 — 10개 트리 81 파일

전체 파일 목록은 `recover-list.txt`. primary 에 이미 있는 것은 `[primary 존재]` 로 갈라 놨다.

| 트리 | 회수 필요 | 내용 |
|---|---:|---|
| `agent-a8b0420c635295c5a` | 31 | `.moai/reports/github-issues/` — issue-1568/1570 등 초안·근거·패치 파일 |
| `agent-a62468d0d1a7040cf` | 11 | `.moai/reports/t114/resolved/` — 핸드오프·해소본 룰 파일 |
| `agent-ad55b5fbe632611a7` | 9 | `.moai/reports/t66/` — RED/GREEN 로그 |
| `t74` | 8 | `.moai/reports/t74/` — cwd 재배치 분석 + RED/GREEN |
| `card-relnotes` | 8 | **SPEC 2건** — `SPEC-CHANGELOG-BUDGET-001`, `SPEC-RELEASE-NOTES-ASSEMBLY-001` (spec/plan/acceptance/progress 전부) |
| `t62` | 6 | `.moai/reports/t62/` — cwd 폴백 분석 + RED/GREEN |
| `t80` | 3 | `.moai/reports/t80/` — codex 조사 원자료 |
| `agent-a77cdfb3c00bf2557` | 2 | `.moai/reports/t71/` RED/GREEN |
| `agent-a20f1e50d08875998` | 2 | `.moai/reports/t42/` RED/GREEN |
| `card-provenance` | 1 | `.moai/reports/session-strategy.html` |

우선순위는 `card-relnotes` 의 SPEC 2건이다 — 리포트 로그와 달리 SPEC 문서는 재생성 비용이 크고,
t146(릴리즈 노트 초안)과 주제가 겹친다.

이미 primary 에 있어 회수 불필요: `t46-anchor-guard`(8), `t140`(1).

### 추적 미커밋 2건 (별도 판단)

| 트리 | 브랜치 | 미커밋 | 성격 |
|---|---|---:|---|
| `agent-a2e81b36e7e5646c7` | `worktree-agent-a2e81b36e7e5646c7` | 37 | 칸반/팩토리 계열 대량 편집. `manager-kanban.md` 를 건드리는 등 **개명 이전 시점**의 스테일 작업으로 보인다. 브랜치 tip 자체는 main 병합됨 — 미커밋분만 남은 상태 |
| `agent-af39f39d9c430af36` | `fix/t126-manager-lead-mirror` | 4 | 템플릿 룰 4파일(agent-common-protocol / agent-authoring / agent-patterns / orchestration-mode-selection). t126 미러 수정 계열 |

두 건 다 **내용을 읽고 살릴지 버릴지 판단**해야 한다. 자동 폐기 대상에서 제외했다.

## 5. release/v3.1.1 병합분 — 배치 PR 이후 폐기

카드가 제외한 배치 12(`t127 t130 t131 t132 t133 t134 t135 t137 t139 t140 t141 t142`) 외에 4개가 같은 성격이다:

| 트리 | 브랜치 | tip |
|---|---|---|
| `ossdocs-t118` | `WT-ossdocs-t118` | `c76566817` |
| `ossdocs-v311` | `WT-ossdocs-v311` | `024536e38` |
| `t121` | `WT-t121` | `16899d982` |
| `t128` | `WT-t128` | `207d2e993` |

배치 PR(`release/v3.1.1` → `main`)이 머지되면 이 4개 + 배치 12 = 16개가 A 분류로 내려온다.

## 6. 폐기 스크립트

```bash
# 확인 (기본값 — 아무것도 지우지 않음)
bash .moai/reports/t145/dispose-merged.sh

# 실제 삭제
APPLY=1 bash .moai/reports/t145/dispose-merged.sh
```

감사 시점 dry-run 실측: **대상 52 / 건너뜀 0**.

스크립트가 트리마다 다시 확인하는 것 3가지 + 1:
1. `git merge-base --is-ancestor <tip> origin/main` — 병합 여부
2. `git status --porcelain -uall` 이 빈 출력 — 추적/미추적 잔여 없음
3. 프로세스 cwd(`lsof -d cwd`)가 그 트리를 물고 있지 않음
4. 디렉터리 부재면 건너뛰고 `git worktree prune` 에 맡김

L1 트리이므로 폐기 동사는 `git worktree remove` 다. `moai worktree done` 은 L2(`~/.moai/worktrees/`) 레지스트리 전용이라
여기 쓰면 범주 오류다. 브랜치 ref 는 `git update-ref -d refs/heads/<branch>` 로 지운다.

## 7. 전체 표 (98행)

`main` = origin/main 조상 여부, `rel` = release/v3.1.1 조상 여부, `미커밋` = 추적 변경 수, `미추적` = untracked 수.
정렬: 폐기 가능 → release 병합 → 배치 12 → 앵커 → 고아 → 회수 필요.

| 트리 | 브랜치 | tip | main | rel | 미커밋 | 미추적 | 처분 |
|---|---|---|---|---|---:|---:|---|
| `agent-a05f5722891df4994` | `WT-t50` | `14bc96f90` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a099178d284386869` | `WT-t95` | `8a6da2f66` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a0a46903178bf9897` | `WT-t99` | `557877c49` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a3b7729a864e4caef` | `WT-t51` | `1fbea5ecf` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a55dd6a3e0ddbb399` | `WT-t54` | `2b9ec29a5` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a66650c94c57df3f8` | `WT-t56` | `4419a3744` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a6ad15916b3c82119` | `WT-t49` | `99bc61cd9` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a828e36a6a5eddb09` | `WT-t48` | `30690cfe1` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a88a53c67fbaa94fe` | `WT-t60` | `470c6af40` | O | O | 0 | 0 | 폐기 가능 |
| `agent-a8e645d1c0d614569` | `worktree-agent-a8e645d1c0d614569` | `caf435ec4` | O | O | 0 | 0 | 폐기 가능 |
| `agent-ad93b0b3d713a9364` | `WT-t96` | `948de4836` | O | O | 0 | 0 | 폐기 가능 |
| `agent-adbea08ea93f5f41c` | `WT-t55` | `ae7da8df1` | O | O | 0 | 0 | 폐기 가능 |
| `agent-add5f3f520fd7123d` | `WT-t52` | `da71dbece` | O | O | 0 | 0 | 폐기 가능 |
| `agent-ae3785cfe611541f0` | `WT-t41` | `901e5244f` | O | O | 0 | 0 | 폐기 가능 |
| `agent-af71696d26ebce0bc` | `WT-t63` | `14a3a4c0a` | O | O | 0 | 0 | 폐기 가능 |
| `card-mcpcmd` | `worktree-card-mcpcmd` | `f7de9eb85` | O | O | 0 | 0 | 폐기 가능 |
| `cc234-align` | `WT-cc234-align` | `1c341dd80` | O | O | 0 | 0 | 폐기 가능 |
| `factory-mode` | `worktree-factory-mode` | `b56f5dbc2` | O | O | 0 | 0 | 폐기 가능 |
| `moai-home-paths` | `worktree-moai-home-paths` | `722b7a248` | O | O | 0 | 0 | 폐기 가능 |
| `release-v310` | `release/v3.1.0` | `3d51358f2` | O | O | 0 | 0 | 폐기 가능 |
| `t103` | `WT-t103` | `572c3b732` | O | O | 0 | 0 | 폐기 가능 |
| `t104` | `WT-t104` | `18df36d2a` | O | O | 0 | 0 | 폐기 가능 |
| `t106` | `WT-t106` | `f5297037f` | O | O | 0 | 0 | 폐기 가능 |
| `t107` | `WT-t107` | `7e4b1f590` | O | O | 0 | 0 | 폐기 가능 |
| `t110` | `WT-t110` | `5c9a9715c` | O | O | 0 | 0 | 폐기 가능 |
| `t111` | `WT-t111` | `f5717f9b7` | O | O | 0 | 0 | 폐기 가능 |
| `t113` | `WT-t113` | `327c6453d` | O | O | 0 | 0 | 폐기 가능 |
| `t115` | `WT-t115` | `72cc1b220` | O | O | 0 | 0 | 폐기 가능 |
| `t116` | `WT-t116` | `1f2dd0a0f` | O | O | 0 | 0 | 폐기 가능 |
| `t32` | `WT-t32` | `d3ddd3406` | O | O | 0 | 0 | 폐기 가능 |
| `t36` | `WT-t36` | `3238d9edc` | O | O | 0 | 0 | 폐기 가능 |
| `t39` | `worktree-t39` | `b06cec203` | O | O | 0 | 0 | 폐기 가능 |
| `t40` | `worktree-t40` | `44a85a364` | O | O | 0 | 0 | 폐기 가능 |
| `t45-i18n-keys` | `worktree-t45-i18n-keys` | `a2adaaba9` | O | O | 0 | 0 | 폐기 가능 |
| `t47` | `WT-t47` | `26935922e` | O | O | 0 | 0 | 폐기 가능 |
| `t53` | `WT-t53` | `4fb49ace7` | O | O | 0 | 0 | 폐기 가능 |
| `t58` | `worktree-t58` | `80ac66e38` | O | O | 0 | 0 | 폐기 가능 |
| `t59` | `WT-t59` | `548ca9504` | O | O | 0 | 0 | 폐기 가능 |
| `t61-availability-docs` | `worktree-t61-availability-docs` | `ce7fb16cb` | O | O | 0 | 0 | 폐기 가능 |
| `t69` | `WT-t69` | `3f5705d8d` | O | O | 0 | 0 | 폐기 가능 |
| `t72-worktree-docs` | `worktree-t72-worktree-docs` | `1344d6ce9` | O | O | 0 | 0 | 폐기 가능 |
| `t75` | `WT-t75` | `d497613dd` | O | O | 0 | 0 | 폐기 가능 |
| `t76` | `WT-t76` | `a3b4c12c8` | O | O | 0 | 0 | 폐기 가능 |
| `t84` | `WT-t84` | `5c3141372` | O | O | 0 | 0 | 폐기 가능 |
| `t85` | `WT-t85` | `fc6123df6` | O | O | 0 | 0 | 폐기 가능 |
| `t86` | `WT-t86` | `dd060a191` | O | O | 0 | 0 | 폐기 가능 |
| `t92` | `WT-t92` | `fa9f51efa` | O | O | 0 | 0 | 폐기 가능 |
| `t97` | `WT-t97` | `c5816c927` | O | O | 0 | 0 | 폐기 가능 |
| `wscfg-graph` | `WT-wscfg-graph` | `12a7ad2cf` | O | O | 0 | 0 | 폐기 가능 |
| `wscfg-paths` | `WT-wscfg-paths` | `72ae4b368` | O | O | 0 | 0 | 폐기 가능 |
| `wscfg-timing` | `WT-wscfg-timing` | `4bd8136b6` | O | O | 0 | 0 | 폐기 가능 |
| `wscfg-worktree` | `WT-wscfg-worktree` | `fa76a816c` | O | O | 0 | 0 | 폐기 가능 |
| `ossdocs-t118` | `WT-ossdocs-t118` | `c76566817` | X | O | 0 | 0 | release 병합 — 배치 PR 후 폐기 |
| `ossdocs-v311` | `WT-ossdocs-v311` | `024536e38` | X | O | 0 | 0 | release 병합 — 배치 PR 후 폐기 |
| `t121` | `WT-t121` | `16899d982` | X | O | 0 | 0 | release 병합 — 배치 PR 후 폐기 |
| `t128` | `WT-t128` | `207d2e993` | X | O | 0 | 0 | release 병합 — 배치 PR 후 폐기 |
| `t127` | `WT-t127` | `e7aeec088` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t130` | `worktree-t130` | `8bcfd506d` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t131` | `WT-t131` | `3f31135c3` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t132` | `worktree-t132` | `ad854ac6c` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t133` | `worktree-t133` | `c326eb4e0` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t134` | `worktree-t134` | `2ceeb7b36` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t135` | `worktree-t135` | `ce310ca82` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t137` | `worktree-t137` | `0b4f7d652` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t139` | `worktree-t139` | `34a94cc80` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t140` | `worktree-t140` | `e7aeec088` | X | O | 0 | 1 | 배치 12 — 범위 제외 |
| `t141` | `worktree-t141` | `9a9778b30` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t142` | `WT-t142` | `b796fddf3` | X | O | 0 | 0 | 배치 12 — 범위 제외 |
| `t143` | `WT-t143` | `4100d8767` | O | O | 2 | 0 | **세션 앵커 — 삭제 금지** |
| `t144` | `WT-t144` | `4100d8767` | O | O | 0 | 0 | **세션 앵커 — 삭제 금지** |
| `t145` | `WT-t145` | `4100d8767` | O | O | 0 | 0 | 본 카드 트리(앵커) |
| `t146` | `WT-t146` | `4100d8767` | O | O | 0 | 0 | **세션 앵커 — 삭제 금지** |
| `agent-a8bf881ac39468a30` | `worktree-agent-a8bf881ac39468a30` | `52ad9255c` | X | X | 0 | 0 | **미병합 고아** |
| `agent-ae0671037b886c5a9` | `worktree-agent-ae0671037b886c5a9` | `216dadc2d` | X | X | 0 | 0 | **미병합 고아** |
| `card-astgrep-path` | `worktree-card-astgrep-path` | `813e98b2f` | X | X | 0 | 0 | **미병합 고아** |
| `card-clocal` | `worktree-card-clocal` | `efba57632` | X | X | 0 | 0 | **미병합 고아** |
| `card-provenance` | `worktree-card-provenance` | `a91bb670f` | X | X | 0 | 1 | **미병합 고아** + 미추적 1건 |
| `card-wtiso` | `card-wtiso` | `b1f7118c0` | X | X | 0 | 0 | **미병합 고아** |
| `installer-fix` | `worktree-installer-fix` | `222c03d0b` | X | X | 0 | 0 | **미병합 고아** |
| `local-only-relocate` | `worktree-local-only-relocate` | `7f2019a12` | X | X | 0 | 0 | **미병합 고아** |
| `local-tree-cleanup` | `worktree-local-tree-cleanup` | `e585ef26e` | X | X | 0 | 0 | **미병합 고아** |
| `mermaid-audit` | `WT-mermaid-audit` | `b0db10bc7` | X | X | 0 | 0 | **미병합 고아** |
| `readme-redesign` | `worktree-readme-redesign` | `c57d24912` | X | X | 0 | 0 | **미병합 고아** |
| `statusline-session-name` | `worktree-statusline-session-name` | `ff7983a3b` | X | X | 0 | 0 | **미병합 고아** |
| `t108` | `WT-t108` | `a15bfdc84` | X | X | 0 | 0 | **미병합 고아** |
| `web-shell-chrome` | `feat/web-shell-chrome` | `3aa383c1e` | X | X | 0 | 0 | **미병합 고아** |
| `agent-a20f1e50d08875998` | `worktree-agent-a20f1e50d08875998` | `2df70172f` | O | O | 0 | 2 | 미추적 2건 — 회수 필요 |
| `agent-a2e81b36e7e5646c7` | `worktree-agent-a2e81b36e7e5646c7` | `4100d8767` | O | O | 37 | 0 | 추적 미커밋 37건 — 회수 검토 |
| `agent-a62468d0d1a7040cf` | `WT-t114` | `0e5120590` | O | O | 0 | 11 | 미추적 11건 — 회수 필요 |
| `agent-a77cdfb3c00bf2557` | `worktree-agent-a77cdfb3c00bf2557` | `b018bf1e9` | O | O | 0 | 2 | 미추적 2건 — 회수 필요 |
| `agent-a8b0420c635295c5a` | `WT-t100` | `02d9cac05` | O | O | 0 | 31 | 미추적 31건 — 회수 필요 |
| `agent-ad55b5fbe632611a7` | `WT-t66` | `62c2bf939` | O | O | 0 | 9 | 미추적 9건 — 회수 필요 |
| `agent-af39f39d9c430af36` | `fix/t126-manager-lead-mirror` | `6a64a0bad` | X | O | 4 | 0 | 추적 미커밋 4건 — 회수 검토 |
| `card-relnotes` | `worktree-card-relnotes` | `18d1ddfc8` | O | O | 0 | 8 | 미추적 8건 — 회수 필요 |
| `t46-anchor-guard` | `worktree-t46-anchor-guard` | `08dc10f20` | O | O | 0 | 8 | 미추적 8건 — 회수 필요 |
| `t62` | `worktree-t62` | `4997b9cf2` | O | O | 0 | 6 | 미추적 6건 — 회수 필요 |
| `t74` | `worktree-t74` | `bf2201e5f` | O | O | 0 | 8 | 미추적 8건 — 회수 필요 |
| `t80` | `WT-t80` | `d7f9f3b3a` | O | O | 0 | 3 | 미추적 3건 — 회수 필요 |

## 8. 측정 방법과 미검증 항목

**측정 명령** (전부 read-only, 2026-08-20, `origin/main` = `4100d8767`, `release/v3.1.1` = `5798bdc2e`):

- 등록 목록: `git worktree list --porcelain`
- 병합 판정: `git merge-base --is-ancestor <tip> origin/main` / `... release/v3.1.1`
- 원격 도달: `git branch -r --contains <tip>`
- 트리 상태: `git -C <path> status --porcelain -uno` / `-uall`
- 세션 앵커: `pgrep -x claude` → `lsof -a -p <pid> -d cwd -Fn`

**미검증 (Gaps)**

- `moai session list --json` 의 PID/cwd 는 앵커 판정에 쓰지 않았다. 등록 항목 6건이 모두 primary 를 cwd 로 보고하는데
  실제 앵커는 워크트리 3곳이었다 — t144 카드가 지목한 PID 결함과 같은 계열이라 신뢰하지 않고 `lsof` 를 근거로 삼았다.
- 앵커 판정은 `pgrep -x claude` 로 잡힌 9개 프로세스 기준이다. 다른 이름으로 도는 세션이 있으면 놓친다.
  스크립트가 실행 시점에 다시 확인하므로 이 시차는 스크립트 안에서 흡수된다.
- 미커밋 2건(`agent-a2e81b36e7e5646c7` 37파일, `agent-af39f39d9c430af36` 4파일)의 **내용 가치는 판단하지 않았다** — 파일 목록만 제시한다.
- 고아 14건이 담은 작업이 이미 다른 경로로 착지했는지는 확인하지 않았다. tip 커밋의 도달 가능성만 봤다.
- 디스크 사용량은 측정하지 않았다.

**삭제는 이 카드에서 실행하지 않았다.** 산출물은 감사표 + 스크립트 + 회수 목록까지다.
