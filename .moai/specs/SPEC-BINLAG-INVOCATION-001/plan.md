# SPEC-BINLAG-INVOCATION-001 — 구현 계획

순서 원칙: **되돌리기 어려운 결정을 위에** 둔다. §F M0(수리 안 선택)과 M1(면제 집합 = 사용자가 체감하는 거동)이 바뀔 가능성이 가장 크고, 아래로 갈수록 기계적이다.

---

## §A 맥락

카드 t366. 워크트리 `.claude/worktrees/t366`, 브랜치 `WT-lint-binary-lag`, HEAD `d7010f86a`.
근거 문서: `.moai/reports/t366/discovery.md`, `.moai/reports/t366/census.md`. 이 계획은 그 실측을 재도출하지 않고 인용한다.

한 문장 요약: PATH의 설치본(`v3.1.2` / `343399d2f`)이 트리(`d7010f86a`)보다 뒤처져 있어서, 트리를 읽는 하위 명령의 초록이 **통과인지 비실행인지** 갈리지 않는다.

---

## §B 알려진 사실 (재조사 금지)

| # | 사실 | 근거 |
|---|---|---|
| B1 | 지연 판정은 이미 존재하고 정확하다 | `doctor.go:518` + `~/go/bin/moai doctor` 실측 |
| B2 | t326의 세션 시작 권고는 **설치 시점** 때문에 안 뜬다. 설계 결손이 아니다 | `merge-base --is-ancestor` = 0, `strings … grep -ci binlag` = 0 |
| B3 | CI는 이 결함을 드러낼 수 없다 | 모든 워크플로가 체크아웃 소스에서 빌드 |
| B4 | `rootCmd`에 `PersistentPreRun*`가 없다 — seam은 빈 슬롯 | `census.md` |
| B5 | cobra `PersistentPreRunE`는 **연쇄하지 않는다**. `worktree`는 자기 것을 가짐 | `root.go:127` |
| B6 | `hook`은 `trivialCommands`에 없고, 43개 래퍼 중 39개가 `moai hook`을 exec | `spec.md` §1.4 |

---

## §C 착수 전 확인

- [ ] Implementation Kickoff Approval 게이트가 **수리 안**을 결정했다(§F M0). 미결이면 run-phase 진입 금지.
- [ ] 워크트리에서 작업 중이며, 공용 설치본을 갱신할 계획이 없다.
- [ ] `git rev-parse --short HEAD`를 커밋 직전에 다시 읽는다(공유 트리 규율).

---

## §D 제약

- 비교 로직 신규 작성 금지 — `internal/binlag` 소비만(제약 C-1).
- 적용 불가 관용을 문구 그대로 보존(C-2).
- 보편 문장 금지 — `worktree` 예외를 명명하거나 연쇄를 명시(C-3).
- 발화는 stderr, stdout·종료 상태 불변(C-4).
- 로컬 전체 테스트 스위트 실행 금지. 변경된 패키지만 돌리고 전수 판정은 CI.
- 재현은 `/tmp` 아래 자기 바이너리로. `/Users/goos/go/bin/moai` 무변경.

---

## §E 자가 검증

- 양방향 재현이 **둘 다** 관측됐는가? 한쪽만이면 미완.
- 인용한 모든 수치에 명령과 그 명령이 잰 트리가 붙었는가?
- 「모든 하위 명령」이라 쓴 문장이 남아 있지 않은가?
- 절차 전후 `moai version` 출력이 동일한가?

---

## §F 마일스톤

되돌리기 어려운 순.

### M0 — 수리 안 결정 (게이트 소관, 코드 0줄) · 우선순위 High

`spec.md` §6의 A / B / C를 게이트에 제시하고 답을 받는다. 이 마일스톤은 **작성자가 수행하지 않는다** — 결정은 Implementation Kickoff Approval에서 나온다.

- **[NEEDS CLARIFICATION: 수리 안 A / B / C 중 어느 것을 채택하는가]** — A(루트 seam), B(절차만), C(좁은 발화). B는 A·C와 배타적이지 않으므로 「A + B」 같은 조합도 유효한 답이다.
- **[NEEDS CLARIFICATION: Option A 채택 시, 면제 집합에 `hook`을 포함하는가]** — 포함이 §1.4의 실측이 가리키는 답이지만, 포함하면 훅 경로는 영영 지연을 알리지 않는다. 훅은 세션 시작 권고(t326)가 이미 덮는 표면이라는 근거로 포함을 권한다.
- **[NEEDS CLARIFICATION: Option A 채택 시, `worktree` 하위 트리를 예외로 두는가 연쇄를 만드는가]** — 예외로 두면 문서에 명명해야 하고, 연쇄를 만들면 `root.go:127` 핸들러를 수정해야 한다. 둘 다 정당하며 비용이 다르다.

M0의 답이 M1 이하 전체의 형태를 정한다. 답 없이 아래로 내려가지 않는다.

### M1 — 면제 집합 확정 (Option A 계열에서만) · 우선순위 High

사용자가 체감하는 거동이므로 코드보다 먼저 확정한다.

- `trivialCommands`(`root.go:46`)를 출발점으로 삼는다. 새 병렬 목록을 만들지 않는다.
- `hook`을 추가한다(M0 답에 따름).
- 런처 3종은 이미 목록에 있고, 제외 사유는 **비용이 아니라 의미**다(`syscall.Exec`으로 프로세스 교체).
- 산출: 면제 집합의 정의 위치 한 곳과, 그것이 `trivialCommands`에서 파생됐음을 보이는 주석.

### M2 — 발화 seam 배선 · 우선순위 High

- `rootCmd`에 `PersistentPreRunE`를 설치하고 `internal/binlag`의 `Evaluate` / `Advisory`를 소비한다. 조상 비교를 새로 쓰지 않는다.
- 시간 상자는 `binaryLagJoinBound`(250 ms)의 **형태**를 따르되, 그 값이 이 경로의 비용 측정이 아님을 주석에 남긴다.
- 실패·초과·저장소 없음은 전부 침묵(C-2).
- `worktree` 처리는 M0의 답대로(예외 명명 또는 연쇄).

### M3 — 양방향 재현 · 우선순위 High

- `/tmp` 아래에 조상 커밋에서 바이너리를 빌드한다. 공용 설치본 무변경.
- 다른 쪽: 경고 발화 관측. 같은 쪽: 침묵 관측.
- 두 방향의 stdout을 바이트 비교해 오염 없음을 함께 확인한다(AC-BLI-006).
- 관측 결과를 `.moai/reports/t366/` 아래 파일로 남기고 `progress.md` §E.2에서 그 경로를 인용한다.

### M4 — 회귀 테스트 · 우선순위 Medium

- 비연쇄 성질을 실제로 확인하는 테스트(루트 훅 설치 후 `worktree` 계열에서의 거동).
- 면제 집합이 `trivialCommands`에서 파생됐음을 깨뜨리는 변경이 테스트를 빨갛게 만드는지 확인(뮤테이션 1건).
- 적용 불가 관용 보존 테스트(비-git 디렉터리에서 침묵).

### M5 — 템플릿 미러와 문서 · 우선순위 Low

- 배포 템플릿에 닿는 변경이 있으면 `internal/template/templates/` 미러 + `make build`. 코드가 `internal/cli/`에만 있으면 미러 대상 없음.
- CHANGELOG 항목은 sync-phase(manager-docs) 소관이며 이 계획에 포함되지 않는다.

### M-C0 — (Option C가 선택된 경우에만) 명령 단위 census · 우선순위 High

Option C의 **첫 비용**. 223개 `Use:` 정의(107개 파일, 중첩 과다 계상)에서 사람이 타이핑하는 명령을 골라내고, 그중 「답이 컴파일된 규칙에 의존하는」 것을 판정한다. 이 census가 없으면 Option C는 착수할 수 없다.

---

## §G 안티패턴

- **한 방향만 재현하기.** 이 카드가 다루는 결함이 「초록의 두 의미가 안 갈림」인데 재현이 같은 모양을 반복하면 검증이 공허해진다.
- **공용 설치본 갱신으로 문제를 「해결」하기.** 그건 이 세션의 증상만 지우고 아홉 레인의 측정을 움직인다. 재설치는 범위 밖(`spec.md` §7).
- **t326을 결함으로 다시 열기.** §1.1.
- **「모든 하위 명령이 알린다」고 쓰기.** B5 때문에 쓴 대로 거짓.
- **호출당 비교를 훅 경로에 붙이기.** B6.
- **새 건너뛰기 목록 발명하기.** `trivialCommands`가 이미 그 역할을 한다.

---

## §H 상호 참조

- `.moai/reports/t366/discovery.md` — Q1/Q2/Q3 실측
- `.moai/reports/t366/census.md` — CLI 표면 census와 seam 가용성
- `.moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/` — t326. 판정과 세션 시작 권고의 정본
- `internal/binlag/binlag.go` — `Evaluate` / `Verdict` / `Advisory` / `RemedyCommand`
- `internal/hook/session_start_binary_lag.go:29` — 조인 경계의 형태 선례
- `internal/cli/root.go:18` / `:46` / `:127` — seam · 건너뛰기 목록 · 비연쇄 사례
