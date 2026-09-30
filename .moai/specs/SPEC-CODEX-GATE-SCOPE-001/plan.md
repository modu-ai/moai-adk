---
id: SPEC-CODEX-GATE-SCOPE-001
title: "구현 계획 — codex 리뷰 게이트 스코핑"
version: "0.1.0"
created: 2026-10-01
---

# SPEC-CODEX-GATE-SCOPE-001 — 구현 계획

## §A 맥락

트리 `.claude/worktrees/t1383`, 브랜치 `WT-gate-scope-lane`. 카드 t1383(리드 발행 2026-10-01). `quality.yaml` `development_mode: tdd` → cycle_type=tdd(RED-GREEN-REFACTOR). Tier **M** — LOC·파일 수만 보면 S 로도 읽히나, AC 예산(13건)과 판별기+receipt 양축 검증 표면이 M 판정이다(progress.md §E.1).

마일스톤은 **되돌리기 어려운 결정을 앞에 둔다** — §B 의 세 결정이 이 SPEC 의 논쟁점 전부이고, M3 이후는 그 결정들의 기계적 귀결이다.

---

## §B 되돌리기 어려운 결정

### B.1 카드 diff 의 측정 형태 (M3-M4 에서 고정)

리드 지시(카드 본문 R1)는 "merge-base with develop..HEAD plus uncommitted changes"로 합의 형태를 정했다(decision-index Q4 — FOUNDER, 킥오프 확정 대상). 남은 결정은 **측정 형태**다:

| 후보 | 형태 | 근거 | 대가 |
|---|---|---|---|
| **(가) 단일 `git diff <merge-base>`** (권장) | 카드 워크트리에서 `git -C <card-tree> diff $(git merge-base develop HEAD)` | git 의 `git diff <commit>` 정의가 정확히 "커밋분 + 미커밋분" 합집합이다. 두 호출 조립이 필요 없고 §8 merge-base 규율과 같은 식이다 | 커밋분/미커밋분을 따로 관측하는 검사가 없다 — AC 는 합집합으로 잰다 |
| (나) 두 호출 조립 | `git diff base..HEAD` + `git diff HEAD` 를 합침 | 두 범위를 구분 관측 가능 | 경계 파일(커밋 후 추가 수정)이 중복되거나 빠지는 조립 버그의 여지 — (가)에 없는 새 실패 모드 |

권장은 **(가)**. 실제 review/start 에 넘기는 표현(target 객체 확장 또는 `uncommittedChanges` + `cwd=카드 트리`)은 run 이 고르되, **측정식은 (가)를 표준**으로 삼는다 — AC 와 회귀 판정식이 같은 식을 공유해야 서로를 검증한다.

비용 유의: 카드가 길어지면 diff 가 커진다. 완화는 REQ-CGS-007 — 판정이 스코프 상태 키에 묶여 재검토가 상태 변화 시로 제한된다. diff 성장이 900s 예산을 위협하는 수준이 되면 상한 설계는 별도 카드로 논의한다(이 카드에서 상한을 만들지 않는다 — spec.md §E).

### B.2 미식별 세션의 페일오픈 방향 (M3 에서 고정)

두 신호(카드 브랜치·env)가 모두 없을 때(decision-index Q5 — FOUNDER):

| 후보 | 형태 | 근거 | 대가 |
|---|---|---|---|
| **(가) 트리 스코프로 떨어진다** (권장) | REQ-CGS-003 경로 — 현행 동작 | **긍정 증거 활성화**: 새 동작(스코프 축소)은 카드 증거가 있을 때만 켜진다. 오식별 세션은 새 빈-검사 경로가 아니라 현행으로 복귀한다 | 레인 규율(WT- 브랜치)을 따르지 않는 세션은 오늘의 결함을 그대로 겪는다 — 그러나 그 세션은 레인 규율 위반 상태다(gitflow-lane-protocol §1) |
| (나) 카드 스코프로 떨어진다 | 빈 diff → 셀프게이트 허용 | 보수적으로 "덜 검사" | 리더·일반 세션이 오식별되면 **게이트가 조용히 무력**해진다 — 차단해야 할 것을 아예 안 본다. (가)보다 나쁜 실패 모드 |

권장은 **(가)**. REQ-CGS-003 이 이 결정의 요구 표현이다.

### B.3 env 의 역할 (M3 에서 고정)

| 후보 | 형태 | 근거 | 대가 |
|---|---|---|---|
| **(가) 비결정 — 로그 전용** (권장) | env 라벨은 REQ-CGS-010 관측에만 실린다 | t1373 실측(라이브 큐 — 미커밋 출처, decision-index Q2): /clear 뒤에도 구 라벨이 잔존하고 retire 된 런의 고아 라벨이 남는다. t1378 은 라벨 네임스페이스를 재구성 예정. **값을 읽지 않는 것이 두 카드와 정합하는 유일한 방법**이다 | env 가 유일한 단서인 상황(비 WT- 체제)에서는 판별 불가 — B.2 (가)로 트리 스코프 |
| (나) env 보조 판정 입력 | 브랜치 부재 시 env 로 카드 스코프 | 리드 지시의 "env OR" 문자열 충족 | stale env 가 리더 세션을 카드로 오분류 → B.2 (나)와 조합되면 최악형. t1378 착지 전까지 어떤 env 파싱도 재검토 없이 도입 불가 |

권장은 **(가)**. env 상수는 `internal/config/envkeys.go`(`EnvMoaiFactoryWorker`)로만 참조 — 문자열 리터럴 금지.

---

## §C 새로 필요한 것 — 스코프 판별기와 스코프 키

1. **판별기** — 입력: 세션 cwd 트리. `git -C <cwd> branch --show-current` 가 `WT-` 접두사를 갖는지 판정. git 호출부와 판정 로직을 분리해 주입형 순수 함수로(선례: `reviewGateChangeDetector`). env 는 읽어 로그 맥락으로만 넘긴다(B.3).
2. **카드 워크트리 해상** — `git worktree list --porcelain` 으로 브랜치↔트리 매핑을 얻고, 세션 cwd 가 속한 트리를 검사 대상으로 삼는다. projectDir(freeze 된 primary)는 설정 읽기(`reviewGateConfigRoot` — SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-008 유지)에만 쓴다.
3. **스코프 키** — receipt 바인딩을 스코프 상태로 확장: 트리 스코프는 현행 `verify.Key(root)` 유지, 카드 스코프는 merge-base + HEAD + 카드 트리 더티 digest 로 구성된 별개 키. `ConfigDigest` 에 스코프 클래스를 넣어 **서로 다른 스코프의 receipt 가 절대 매치되지 않게** 한다(REQ-CGS-007).
4. **대상 표현** — 카드 스코프의 review/start 대상 표현은 run 이 고른다. 제약: 트리 스코프 경로의 직렬화는 REQ-CRT-006 shape 유지.

배치: `internal/cli/codex_review_gate.go` + `codex_review_receipt.go` + `codex_stop_chain.go` 멤버 6. 새 파일 필요 여부는 run 이 판단한다 — 기존 주입 seam 선례를 따른다.

**[HARD] 두 경로의 판별은 하나의 함수를 지난다.** `HandleCodexReviewGate` 와 멤버 6 이 서로 다른 판별 구현을 갖는 순간 REQ-CGS-009 가 깨진다 — 같은 세션 상태에 대해 두 경로가 다른 스코프를 보고, receipt 가 영원히 매치되지 않거나 엉뚱한 receipt 를 매치시킨다.

---

## §D 마일스톤

### M1 — 회귀선 고정 (우선순위 High)

- 트리 스코프 현행 경로 검사: 비카드 세션 픽스처(일반 트리 + 미커밋)에서 대상 = uncommittedChanges + cwd = 트리. **변경 전 트리에서 초록** 관측.
- fail-open 회귀(AC-CGS-011) 초록 고정.
- 산출: 회귀 검사 2건 + 초록 출력. 초록이 아니면 즉시 중단·보고.

### M2 — RED 확립 (우선순위 High)

- AC-CGS-001·003~010·012·013 을 구현하는 검사를 **프로덕션 변경 없이** 추가. 픽스처: 카드 워크트리(WT- 브랜치, 커밋 1 + 미커밋 1) + primary 역할 트리의 외부 WIP + env 매트릭스(acceptance.md §B).
- `-v` 로 `=== RUN` + `--- FAIL` 을 관측하고 `.moai/reports/t1383/red/` 에 출력 보존 — progress.md §E.2 인용.
- 산출: RED 로그.

### M3 — 스코프 판별기 + 트리 우선 해상 (우선순위 High)

- §C 1-2 구현: WT- 판별(1차 신호), env 로그 전용(B.3), 미식별 → 트리 스코프(B.2 (가)).
- REQ-CGS-005: cwd 트리 우선, projectDir 는 설정 루트 용도로만.
- REQ-CGS-006: 셀프게이트 스코프 일치.
- M2 RED 중 이 축이 초록으로 뒤집히는지 관측.

### M4 — 카드 스코프 검사 + receipt 스코프 키 (우선순위 High)

- §C 3-4 구현: 카드 diff 측정(B.1 (가) 식), 대상 표현, 양 경로(핸들러 + receipt 생산자 + 멤버 6)에 동일 판별 함수 배선(§C [HARD]).
- REQ-CGS-007: 이판식별(다른 스코프 상태 receipt 비매치) + 동일 상태 재차단 유지.
- REQ-CGS-009 패리티 검사 초록.

### M5 — 관측 + 문서 (우선순위 Medium)

- REQ-CGS-010: 스코프 클래스·근거 로그.
- 판별 기준 문서화: `.moai/docs/`(또는 레인 프로토콜 문서)에 판별 기준·페일오픈 방향 요지 추가 — 정본은 spec.md §B 이고 문서는 요지+링크만. sync-phase 소관.
- 기계적 변경. 마지막에 둔다.

---

## §E 자기 검증

| 항목 | 명령 |
|---|---|
| 대상 테스트 | `go test ./internal/cli/ -run '^TestCodexReviewGate\|^TestCodexReviewScope\|^TestProduceCodexReviewReceipt' -count=1 -v` (D1 — 이름 앵커 형태; 오탈자 셀렉터의 0매칭 초록 방지) |
| 정적 검사 (darwin) | `go vet ./internal/cli/...` |
| 정적 검사 (windows) | `GOOS=windows go vet ./internal/cli/...` |
| RED 보존 | `ls .moai/reports/t1383/red/` |
| 범위 침범 확인 | `git diff --stat` — `mcp_convergence.go`·`codex_sync_gate.go` 가 목록에 없어야 한다 |
| 라벨 파싱 부재 | `grep -n 'worker-\|lane-' internal/cli/codex_review_gate.go internal/cli/codex_review_receipt.go` — 로직 매칭 0 (주석 서술은 허용) |

전체 스위트(`go test ./...`)는 로컬에서 돌리지 않는다 — 레인 부하 규율. 전 패키지 판정은 CI(origin/develop) 몫.

---

## §F 안티패턴

1. **verdict 값으로 스코프를 검증하기.** 스텁은 요청과 무관한 verdict 를 돌려준다. 관측 대상은 요청의 target·cwd 와 receipt 상태다.
2. **라벨 정규식 도입.** `^worker-\d+$` 같은 형태 매칭은 t1378 재구성과 충돌한다 — REQ-CGS-004 의 shall not.
3. **env 로 판정 전환.** t1373 이 실측한 stale 신호다. 로그 전용(B.3).
4. **merge-base 핀.** gitflow-lane-protocol §8 — 흡수가 판정식을 오염시킨다. 매 평가 재계산.
5. **fail-close 누설.** 스코프 해상 실패를 차단로 쓰면 REQ-MCP-012 위반이다. 해상 실패는 트리 스코프(B.2)다.
6. **래퍼 쌍둔 미동기화.** 래퍼를 건드리면 템플릿 쌍둔을 같은 커밋에 넣고 `make build`. 안 건드리는 것이 기본 전제다.
7. **`uncommittedChanges` 트리 경로 회귀.** REQ-CRT-006 shape-identical — 카드 스코프 추가가 현행 경로의 직렬화를 바꾸면 안 된다.
8. **0매칭 초록.** 픽스처 워크트리에 WT- 브랜치가 실제 설정됐는지 `-v` `=== RUN` 으로 관측한다. 셀렉터 오탈자는 적색 없이 초록을 낸다.
9. **두 경로에 판별 구현 두 벌.** §C [HARD] — 한 함수.
10. **전체 스위트 로컬 실행.** 레인 부하 규율 — 대상 패키지만.

---

## §G 참조

- spec.md §A(측정) · §B(요구) · §E(범위 밖) · §F(제약)
- acceptance.md §A-§E(AC 규율·매트릭스·DoD)
- decision-index.md — Q1-Q5(B.1-B.3 결정의 권위 표기)
- `.claude/rules/local/gitflow-lane-protocol.md` §1·§8
- 관련 SPEC: SPEC-CODEX-REVIEW-TARGET-001 · SPEC-MOAI-MCP-SERVER-001 · SPEC-DUAL-HARNESS-HOOK-PARITY-001 · SPEC-WORKTREE-STATE-ROOT-001
