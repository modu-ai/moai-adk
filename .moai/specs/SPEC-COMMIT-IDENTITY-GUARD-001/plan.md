# Plan — SPEC-COMMIT-IDENTITY-GUARD-001

## §A 맥락

card t1289. 원인 규명(`.moai/reports/t1289/root-cause.md`)이 공유 설정층 오염을 측정으로 확정했고,
운영자는 쓰기 통로 수리와 별개로 **커밋 시점 가드**를 요구했다. 개발 방식은 TDD(RED-GREEN-
REFACTOR) — 새 가드는 새 코드이므로 실패 테스트가 먼저다. 기존 가드 계열의 동작은 바꾸지 않는다.

## §B 알려진 사항 (plan-time 측정)

### B.1 내장 거부 목록 열거 — 측정 명령과 결과

측정 트리: 워크트리 t1289, HEAD `37dc766b9`, 2026-09-28. 명령(단일 호출, 워크트리 루트에서):

```bash
grep -rhoE '(user\.email"?,? *"[^"]+"|user\.email=[^" ,)]+|GIT_(AUTHOR|COMMITTER)_EMAIL=[^" ,)]+)' --include='*_test.go' internal pkg cmd | sed -E 's/.*(=|", *")//; s/"$//' | grep '@' | sort -u
```

결과: 서로 다른 리터럴 **51** 개.

```
a@e.invalid anchor-test@example.com audit@example.invalid board-test@example.com
branch-guard-test@example.com c@e.invalid disposal-test@example.com f@e.com f@example.com
f3@example.com f4@example.com fixture@example.com fixture@example.invalid fixture@example.test
fixture@t516.invalid fx@example.com guard-test@example.invalid m2-test@example.com o@e.x
other@example.com slot-cli-test@example.com slot-lease-test@example.com someone@example.invalid
stopchain-test@example.com sub@e.x sync-gate-test@example.com t@example.com t@example.invalid
t@local t@t t@t.local t@t.t t@t.test t1074@example.invalid t1204@example.invalid t371@example.com
t461@example.test t488@example.invalid t516@test.invalid t560@test.invalid t602@example.invalid
t603@example.invalid t606@example.test t655@test.local t688@example.test t766@example.invalid
t768@example.test test@example.com test@example.invalid test@test.com tier-guard-test@example.com
```

**이 값은 plan 시점 스냅숏이다.** run 페이즈는 run base 에서 같은 명령을 다시 돌려 그 출력을
내장 목록의 근거로 쓰고, 수치를 plan 값에서 옮겨 적지 않는다. 사고 신원 `t@t.t` 와 카드가 예시한
`t@t.test` 는 둘 다 목록에 있다.

열거 명령의 한계: 변수로 넘긴 이메일(`"user.email", email`)과 여러 줄에 걸친 리터럴은 잡지 못한다.
미탐 방향이므로 받아들이며, REQ-CIG-007 의 드리프트 테스트도 **같은 술어**를 쓴다(술어가 달라지면
테스트와 목록이 서로 다른 집합을 재게 된다).

### B.2 형제 가드 관례 (측정)

- 센티널 14 종, `TEST_IDENTITY` 부재: `grep -rhoE '"[A-Z_]+_VIOLATION"' internal/hook --include='*.go' | sort -u`.
- 형제 설정 키(`branch_guard`, `integration_lock`)는 템플릿 `workflow.yaml` 에 실리지 않는다
  (`grep -rln 'integration_lock\|branch_guard' internal/template/templates/.moai/config/` → 0 건).
  이 SPEC 은 D7 결정대로 키를 싣는다.
- 로컬 `.moai/config/sections/workflow.yaml` 에는 `branch_guard.enabled: true` 가 있다.

### B.3 판단 대기 항목

- [NEEDS CLARIFICATION: 저장소 범위 — 가드가 명령의 대상 저장소와 무관하게 거부할 것인가,
  이 프로젝트 저장소(공유 common dir)를 대상으로 하는 커밋만 거부할 것인가.
  (a) 저장소 무관 거부(권장 기본): 파싱이 단순하고 `git -C <다른경로>` 해석이 필요 없다.
  대가 — 레인이 `/tmp` 임시 저장소에서 수동 프로브 커밋을 할 때도 픽스처 신원이면 거부된다.
  (b) 이 저장소 한정: `-C`·`cd` 대상 해석이 필요해 미탐 표면이 넓어지고 코드가 늘어난다.]

## §C 사전 점검

- [ ] 워크트리 격리 세션, base SHA 기록(`git rev-parse --short HEAD`).
- [ ] B.1 열거 명령을 run base 에서 재실행해 출력을 증거로 보존.
- [ ] `internal/hook` 기준선: `go test ./internal/hook/...` 를 변경 전에 돌려 사전 실패 유무 기록
      (develop 에 알려진 사전 실패가 있으면 이 카드 귀속과 분리해 적는다).
- [ ] golangci-lint 판 확인 — CI 판 v2.1.6 으로 잰다.

## §D 제약

- 새 명령 파서를 만들지 않는다. `substituteQuotedArguments`, `substituteHeredocBodies`,
  `substituteShellComments`, `normalizeGitExeSuffix` 등 `branch_guard.go` 도우미를 재사용한다.
- 탐침은 seam(패키지 변수 함수) 뒤에 둔다. 테스트는 실제 `git` 을 부르지 않는 가짜 탐침과,
  `t.TempDir()` 임시 저장소에서 실제 `git var` 를 부르는 통합 테스트를 모두 둔다(후자가 D3 해석
  경로의 이빨이다). 임시 저장소 신원은 `cmd.Env` 로만 준다 — 설정 파일에 쓰지 않는다(통로 B 교훈).
- 탐침 환경: `internal/gitenv` 로 `GIT_DIR`/`GIT_INDEX_FILE`/`GIT_COMMON_DIR` 류를 걷어내되
  신원 변수는 보존한다.
- 테스트에서 OTEL 환경 변수 `t.Setenv` 금지, `HOME` 교체 금지(병렬 오염).
- 템플릿 문언 중립(SPEC ID·카드 번호·날짜·커밋 SHA 금지).

## §E 자체 검증 (run 페이즈가 쓰는 범위)

로컬 전체 스위트는 돌리지 않는다. 명령은 각각 단일 호출로 실행한다.

- E1 `go test -timeout 30m ./internal/hook/...`
- E2 `go test -timeout 30m ./internal/gitenv/...`
- E3 `go test -timeout 30m ./internal/config/...` (설정 키를 더하므로 해당)
- E4 형제 센티널 스윕: `grep -rn '_VIOLATION' internal/hook --include=*.go` — 새 센티널이 한 곳에서
  정의되고 형제와 충돌하지 않음을 확인.
- E5 `golangci-lint run ./internal/hook/... ./internal/config/...` — v2.1.6.
- E6 `make build` exit 0 (템플릿 키 추가 후).
- E7 변이 증거(AC-CIG-011): 세 변이를 각각 적용한 트리에서 E1 의 해당 테스트가 실패함을 기록하고
  변이를 되돌린다. 변이는 커밋하지 않는다.
- E8 AC 기준선 스냅숏이 acceptance.md 변경과 같은 커밋에 있는지 확인(plan 페이즈에서 이미 적용).

## §F 마일스톤 (되돌리기 비용 순 — 바뀔 가능성이 큰 결정부터)

### M1 — 설정 표면과 거부 목록 (우선순위 High)

- `WorkflowConfig` 에 `CommitIdentityGuard` (`enabled bool`, `deny_emails []string`) 추가,
  `defaults.go` 에 `Enabled: false`.
- 내장 목록 상수(§B.1 재측정 결과) + 드리프트 테스트(같은 열거 술어로 저장소 `*_test.go` 를 훑어
  목록 누락 시 실패, 실패 메시지에 누락 리터럴과 처치 방법을 싣는다).
- 로컬 `.moai/config/sections/workflow.yaml` 에 `commit_identity_guard.enabled: true`,
  템플릿 `workflow.yaml` 에 중립 문언의 `enabled: false`.
- §B.3 판단 결과를 반영.

### M2 — 판별 핵심: 트리거 + 신원 해석 + 비교 (우선순위 High)

- 트리거 매처(정규화 후 git 동사 8 종), 명령 수준 재정의 추출기, `git var` 탐침 seam,
  `<name> <email> <epoch> <tz>` 출력에서 이메일 추출, 정확 일치 비교.
- 탐침 시간 상한 기본 설계값 2 초(초과 시 REQ-CIG-006 경로). run 페이즈가 근거를 들어 조정할 수 있다.
- 실패 시 `.moai/logs/commit-identity-guard-audit.log` advisory 한 줄.

### M3 — 배선 (우선순위 Medium)

- `pre_tool.go` 에서 통합 잠금 가드 다음에 `commitIdentityGuardEnabled()` 게이트로 호출.
- PowerShell 분류 불가 구문은 `appendUnclassifiedAudit` 관례로 기록 후 통과.

### M4 — 이빨과 마감 (우선순위 Medium)

- 변이 3 종 증거, 양성 대조, 탐침 0 회 계수 테스트.
- E1-E8 실행, 증거를 `.moai/reports/t1289/` 로 반출.

## §G 반패턴

- 가드가 꺼졌거나 트리거가 아닌데 `git var` 를 돌리는 구현(모든 셸 호출에 서브프로세스 비용).
- 탐침 실패를 거부로 처리하는 구현(fail-closed) — 공유 설정이 잠깐 깨지면 모든 레인의 커밋이 멈춘다.
- 이메일 패턴 휴리스틱(`*@example.*`) — 결정 D4 위반, 실사용자 오탐 위험.
- 테스트 픽스처가 신원을 `git config` 로 설정 파일에 쓰는 것 — 이 사고의 통로 B 그 자체.
- 변이 증거를 실행 없이 "이 테스트가 잡을 것"으로 서술하는 것.

## §H 상호 참조

- `spec.md` §4 결정 D1-D7, §5 REQ-CIG-001..009.
- `acceptance.md` AC-CIG-001..012.
- `.moai/reports/t1289/root-cause.md`.
- `internal/hook/branch_guard.go`, `internal/hook/integration_lock_guard.go`, `internal/hook/pre_tool.go`.
