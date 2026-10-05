# t679 판정서 — 오래된 미병합 브랜치 7개 개별 판정

- 측정일: 2026-09-13
- 측정 트리: `.claude/worktrees/t679` (branch `WT-stale-branch-verdict`, 기점 = develop 팁 `5e0f71175` — 0/0 divergence 실측)
- 측정자: 레인 (카드 t679, 리드 배차)
- [HARD] 병합·브랜치 삭제·push 는 수행하지 않았다 — 권고까지만. 병합 여부는 운영자 판정.

## 측정 방법

- ahead/behind·patch 동등성: `git rev-list --count develop..<branch>`, `git cherry develop <branch>`
- 순내용 델타: `git diff --stat develop...<branch>` (merge-base 기준 3-dot)
- 흡수 여부: `git merge-base --is-ancestor`, 파일 단위 `git diff <branch> develop -- <path>` (빈 출력 = 바이트 동일)
- 카드 상태: `moai todo list --limit 0` (active) + `--dropped`

## 요약 표 (7행)

| # | 브랜치 | 카드 | ahead | 순델타(develop...branch) | 판정 | 병합 권고 |
|---|---|---|---|---|---|---|
| 1 | `WT-ci-flake-series` | t278 · SPEC-CI-FLAKE-SERIES-001 | 12 (전부 t278 자체) | 7파일/350행 — **코드 0**, 문서·장부만 | 거의 완전 흡수. 살아있는 의도는 AC-CFS-007 관측 창 마감인데 그것은 develop 위 새 작업 | **불요 — 폐기 가능** (스테일 progress.md 가 develop 진화본을 되돌림) |
| 2 | `WT-taskstop-name-reclaim` | t267 · SPEC-TEAMMATE-REVIVAL-GUARD-001 | 13 | 23파일/2340행 표기 — 그러나 **핵심 코드 바이트 동일로 이미 흡수** | 완전 초월됨. SPEC 아티팩트·correlation 문서 전부 develop 존재 | **불요 — 폐기 가능** |
| 3 | `WT-orphan-queue-sweep` | t542 (queued — 생존) | 3 | 13파일/962행 — 전부 `.moai/reports/t542/` (develop 에 디렉터리 자체가 없음, **100% 고유**) | 카드 산출물의 증거 단계를 운반. 코드 0 | **병합 권고** (문서 전용 무위험; 실제 정리 실행은 카드 소관로 별도) |
| 4 | `WT-pull-falsifier-window` | t547 (queued — 생존) | 2 (흡수병합 1 + 문서 1) | 5파일/55행 문서 전용 (develop 부재, **100% 고유**) | 「진입 조건 미충족·수집 미시작」 관측 기록 | **병합 권고** (경량 기록 보존) |
| 5 | `WT-worktree-vanish` | t567 | **0** | 0 — **develop 에 완전 흡수** (`--is-ancestor` 실측; 09-13 lane-3 배치 병합과 일치) | 카드 표기 +3 은 스캔 시점 이전 값 | **불요 — 폐기 가능 (이미 병합됨)** |
| 6 | `WT-mx-pending-specref` | t620 (**id 재발행 충돌** — 워크플로 감사 F14 카드가 같은 번호 사용, `git log --all --grep t620` 실측) | 1 | 2파일/316행: `internal/mx/scanner_specref_discard_test.go` 180행 + verdict (develop 부재, **100% 고유**) | 불변식 잠금 테스트 — **본 레인이 develop 팁 트리(5e0f71175)에서 실행해 PASS 실측** (`go test ./internal/mx/ -run TestSpecRefDiscard_UnpairedWarnBlock -count=1` → `ok ... 0.349s`) | **병합 권고** (단일 커밋 — `a44441ced` cherry-pick 도 가능) |
| 7 | `WT-sync-gate-failstate` | t624 | 1 | 1파일/33행 증거문서 (develop 부재) | **브랜치 본체는 과거 develop 에 병합됨** (`a69ff1c03` ancestor 실측). 그 뒤 브랜치에 추가된 문서 1건만 잔존 | **불요(본체는 이미 병합)** — 잔존 1파일은 운영자가 흡수/폐기 선택 |

## 브랜치별 근거

### 1. WT-ci-flake-series (t278)

- ahead 12 커밋은 전부 t278 자체(plan/M1-M4/sync 문서). 코드 수리 3건(324883e AND-gate·5aedc1c p95·def99739d poller TOCTOU)의 내용은 `379b310a6`(#1666) 경유 main→develop 에 이미 반영 — 브랜치가 `2aec09c27` 에서 origin/main 을 병합하며 상쇄돼, 3-dot 순델타에 코드 파일이 0개.
- develop 은 브랜치보다 **앞서** 있다: `.moai/reports/t278/` 에 sync-audit.md·timing-statistic-decision.md 존재(후자는 브랜치에 없음), `progress.md` 17KB 진화본 존재(브랜치판은 구 base 위 +32행).
- 생존 의도: SPEC 상태 "implemented, AC-CFS-007 window pending"(3a7aaab37 커밋 메시지 실측). 창은 2026-08-27 개시 — 마감 작업이 살아있는 의도이며, 그것은 이 브랜치 병합이 아니라 develop 위의 새 카드여야 한다.
- 병합 시 위험: 스테일 `progress.md`·`spec.md` 가 develop 진화본과 충돌 또는 회귀.

### 2. WT-taskstop-name-reclaim (t267)

- 코드 흡수 실측: `git diff WT-taskstop-name-reclaim develop -- internal/hook/agent_stop_guard.go internal/hook/agent_stop_guard_test.go` → **빈 출력(바이트 동일)**. develop 측 착지 커밋 `968ed2acb`(#1667).
- config 도 흡수: develop `workflow.yaml:168` `agent_stop_guard:` 키 + `defaults.go` 존재.
- 브랜치가 "추가"로 보이는 2340행 중 코드부는 구버전 — develop 의 `pre_tool.go` 는 t596(09-12 symlink 수리)·#1658(dangerous_removal 구조화)·t607(slot-lease) 까지 진위. `settings.json.tmpl` 도 develop 이 navigator·pre-tool 와이어링을 더 갖고 있다(브랜치엔 없음).
- 문서 흡수 실측: SPEC 디렉터리 6파일 전부 + `.moai/docs/agent-stop-audit-correlation.md` 가 develop 에 존재(`ls` 실측).

### 5. WT-worktree-vanish (t567) — 카드 표기 정정

- `git rev-list --count develop..WT-worktree-vanish` = **0**, `--is-ancestor` = 참. 리드 스캔의 "+3" 은 09-13 lane-3 배치(t567 병합) 이전 값. **이미 병합된 브랜치를 "미병합"으로 세는 표기 오류 정정.**

### 6. WT-mx-pending-specref (t620) — id 충돌 주의

- `git log --all --grep t620` 실측: `3c2581814`(워크플로 감사 F14 병합)·`ff12d8eb8`(plan-auditor 입력 타입) 등 **다른 카드가 같은 번호 t620 을 쓴 이력**. 본 브랜치의 t620 은 mx specref 판정 카드이며, 큐에는 어느 쪽 레코드도 없음(활성·dropped 모두 부재 — purged).
- 유일 커밋 `a44441ced`(2026-09-12): 테스트 180행 + `.moai/reports/t620/verdict.md` 136행.
- **재측정 실측**: develop 팁 트리에 테스트 파일만 임시 유입 → `ok github.com/modu-ai/moai-adk/internal/mx 0.349s` → 원복 완료. 불변식(대기 WARN 블록의 @MX:SPEC 폐기는 결함 아님)이 현재 트리에서도 성립.

### 7. WT-sync-gate-failstate (t624)

- `git merge-base --is-ancestor a69ff1c03 develop` = 참(과거 병합 착지 실측), `git log a69ff1c03..WT-sync-gate-failstate` = `722fc89a6` 1건. 잔존분은 `.moai/reports/t624/window-cli-3fail-develop-control.txt`(33행) 단독.

## Gaps (미검증)

1. t278·t267·t620·t624 카드 레코드가 큐에 부재(active·dropped 모두 0건, `moai todo pr` → "No backlog item") — 공식 done 여부는 커밋·파일 증거로만 추정했다. t620 은 id 재발행 충돌로 어느 카드가 닫혔는지 큐 단독으로는 복원 불가.
2. t620 테스트는 지정 1함수만 실행했다(패키지 전체 스위트 아님 — 본 카드는 판정 카드이며 전수 검증은 병합 후 CI 몫).
3. 병합 시나리오(충돌 해소 포함)를 실제로 돌리지 않았다 — CHANGELOG 2행 등 미세 충돌 가능성은 미측정.
4. 각 브랜치의 워크트리 존재 여부(폐기 대상 트리)는 확인하지 않았다 — 본 판정은 브랜치 ref 기준.

## Residual-risk

- 권고 3·4·6 의 "병합 권고"는 문서·테스트 델타 기준이며, 병합 자체는 운영자 판정 사항이다(배차 [HARD] 준수 — 미병합).
- t278 의 AC-CFS-007 관측 창이 실제로 만료했는지(마감 요건 충족)는 본 카드 범위 밖 — 창 마감 카드를 새로 발행할 때 판정 필요.
