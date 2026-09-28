# SPEC-COMMIT-IDENTITY-GUARD-001 — 인수조건

> 각 항목은 Given-When-Then 이며 이진 판정 가능하다. 각 AC 는 머리에 **Covers** 로 자기가 덮는
> `REQ-CIG-XXX` 를 인용한다. "테스트" 는 run 페이즈가 `internal/hook`(설정 표면은
> `internal/config`)에 두는 이름 붙은 Go 테스트이며, 판정은 그 테스트의 실제 실행 출력으로 한다.
>
> **저장소 범위 전제(D8).** AC-CIG-001..012 의 가짜 탐침 테스트는, 따로 적지 않는 한 저장소 범위
> 탐침이 대상과 프로젝트에 **같은** common dir 을 돌려주는 상태(= 이 저장소 대상)를 전제한다. 범위가
> 다른 경우는 AC-CIG-013, 링크드 워크트리는 AC-CIG-014 가 덮는다.

## AC-CIG-001 — 설정층 경유 픽스처 신원의 커밋이 거부된다

**Covers**: maps REQ-CIG-003

**Given** 가드가 켜져 있고, 탐침이 author·committer 이메일로 `t@t.t` 를 돌려주는 상태에서
**When** 셸 도구가 `git commit -m x` 를 실행하려 할 때
**Then** 결정은 deny 이고, 사유는 `TEST_IDENTITY_VIOLATION:` 으로 시작하며 `t@t.t`, 역할
(`author` 또는 `committer`), 처치 방법을 담는다. author 만 일치하는 경우와 committer 만 일치하는
경우 각각도 deny 다(표 테스트 3 행).

## AC-CIG-002 — 트리거 동사 8 종이 모두 검사된다

**Covers**: maps REQ-CIG-002

**Given** 가드가 켜져 있고 탐침이 픽스처 신원을 돌려주는 상태에서
**When** `commit`, `merge`, `cherry-pick`, `revert`, `rebase`, `am`, `commit-tree`, `pull` 을 각각
실행하는 명령(맨 git, `git.exe`, `&&` 연쇄 뒤쪽 위치 포함)이 들어올 때
**Then** 8 동사 모두 deny 다. 표 테스트는 동사 목록을 가드의 트리거 정의에서 읽지 않고 테스트 안에
직접 적는다(가드 쪽 목록에서 동사가 빠지면 테스트가 그 차이를 잡아야 하므로).

## AC-CIG-003 — 트리거가 아닌 명령에는 탐침이 돌지 않는다

**Covers**: maps REQ-CIG-001

**Given** 가드가 켜져 있고 저장소 범위·신원 탐침 seam 이 각각 호출 횟수를 센다
**When** `git status`, `git log --oneline`, `git diff`, `echo 'git commit -m x'`, 주석
`# git commit`, heredoc 본문 안의 `git commit` 이 각각 들어올 때
**Then** 모든 경우 결정은 allow 이고 범위·신원 탐침 호출 수는 각각 **0** 이다.

## AC-CIG-004 — 명령 수준 재정의가 검사된다

**Covers**: maps REQ-CIG-004

**Given** 가드가 켜져 있고 탐침은 거부 목록 밖의 신원(예: `dev@real-host.invalid`)을 돌려주는 상태에서
**When** 다음 명령이 각각 들어올 때 —
(a) `GIT_AUTHOR_EMAIL=t@t.t git commit -m x`,
(b) `GIT_COMMITTER_EMAIL=t@t.test git commit -m x`,
(c) `export GIT_AUTHOR_EMAIL=t@t.t && git commit -m x`,
(d) `git -c user.email=t@t.t commit -m x`,
(e) `git commit --author="t <t@t.t>" -m x`
**Then** 다섯 경우 모두 deny 다. 같은 모양에 목록 밖 이메일을 넣은 대조 5 행은 모두 allow 다.
추가 짝 대조는 우선순위를 구분한다: 저장소 `user.email` 설정이 `t@t.t` 이고 명령 밖
`GIT_*_EMAIL` 환경 값은 없는 상태에서
`GIT_AUTHOR_EMAIL=dev@real-host.invalid GIT_COMMITTER_EMAIL=dev@real-host.invalid git commit -m x`
와 `git -c user.email=dev@real-host.invalid commit -m x` 는 두 역할이 모두 대체돼 allow 다.
저장소 `user.email=dev@real-host.invalid` 가 있는 상태에서 `EMAIL=t@t.t git commit -m x` 는
Git 이 그 `EMAIL` 을 쓰지 않으므로 allow, 같은 저장소의 `GIT_AUTHOR_EMAIL=t@t.t git commit -m x`
는 deny 다. 반대로 기본 author 만 `t@t.t` 인 상태에서
`git commit --author="t <dev@real-host.invalid>" -m x` 는 author 가 대체되고 committer 가
목록 밖이면 allow 다. 대조값은 가드 자신의 테스트 파일에만 두고 AC-CIG-010 열거에서 제외한다.

## AC-CIG-005 — 실신원은 통과한다 (양성 대조)

**Covers**: maps REQ-CIG-005, REQ-CIG-011

**Given** 가드가 켜져 있고, `t.TempDir()` 임시 저장소를 **프로젝트 디렉터리이자 명령 cwd** 로 두고
(D8 범위 일치), 비병렬 테스트의 실제 `git var` 탐침 seam 에 넘기는 환경 목록(`cmd.Env`)에만
비픽스처 신원을 준 상태에서 (훅 프로세스 환경·설정 파일 쓰기 없음)
**When** 실제 `git var` 탐침으로 `git commit -m x` 를 평가할 때
**Then** 결정은 allow 이고, 탐침이 실제로 돌려준 author·committer 이메일이 주입한 비픽스처
이메일과 같으며, `.moai/logs/commit-identity-guard-audit.log` 에 추가된 줄은 0 이다. 같은
저장소의 탐침 `cmd.Env` 로 `t@t.t` 를 준 짝 테스트는 deny 다 — 두 테스트가 짝으로 있어야
실제 해석 경로가 판별에 쓰였음이 선다.

## AC-CIG-006 — 정확 일치만 거부한다

**Covers**: maps REQ-CIG-003, REQ-CIG-005

**Given** 가드가 켜져 있다
**When** 탐침 이메일이 `T@T.T`(대소문자), ` t@t.t `(앞뒤 공백), `t@t.tt`, `xt@t.t`, `t@t.t.example.org`
중 하나일 때
**Then** 앞의 두 값은 deny, 뒤의 세 값은 allow 다.

## AC-CIG-007 — 해석 실패는 통과 + 감사 한 줄

**Covers**: maps REQ-CIG-007

**Given** 가드가 켜져 있고, 신원 탐침의 (a) 오류 종료, (b) 시간 상한 초과, (c) `<` `>` 가
없는 출력, (d) 존재하지 않는 cwd, 범위 탐침의 (e) 대상 쪽 실패, (f) 프로젝트 쪽 실패 중
하나를 겪도록 만든 상태에서(유효한 명령 수준 거부 이메일 없음)
**When** `git commit -m x` 가 들어올 때
**Then** 여섯 경우 모두 결정은 allow 이고, `t.TempDir()` 프로젝트 루트의
`.moai/logs/commit-identity-guard-audit.log` 에 원인을 담은 줄이 정확히 1 줄 추가된다.
짝 대조는 범위 탐침이 같은 common dir 을 돌려준 뒤 신원 탐침이 실패하는 상태에서
`GIT_AUTHOR_EMAIL=t@t.t git commit -m x` 를 넣는다. 이 경우 명령 수준 author 거부 이메일이
양성 증거이므로 deny 이고, 위 감사 로그에 탐침 실패 줄은 추가되지 않는다.

## AC-CIG-008 — 꺼져 있으면 가드가 불리지 않는다

**Covers**: maps REQ-CIG-006

**Given** 설정의 `workflow.commit_identity_guard.enabled` 가 `false` 인 경우와 키가 없는 경우
**When** 탐침이 픽스처 신원을 돌려주는 상태에서 pre-tool 핸들러가 `git commit -m x` 를 처리할 때
**Then** 두 경우 모두 결정은 allow 이고 범위·신원 탐침 호출 수는 각각 0 이다. 또한 `internal/config` 기본값 테스트가
`Workflow.CommitIdentityGuard.Enabled == false` 를 단언한다.

## AC-CIG-009 — 설정 목록은 내장 목록에 더해진다; 템플릿·로컬 설정

**Covers**: maps REQ-CIG-008, REQ-CIG-006

**Given** `deny_emails: ["ops-bot@corp.invalid"]` 를 설정한 상태에서
**When** 탐침이 `ops-bot@corp.invalid` 를 돌려줄 때와 `t@t.t` 를 돌려줄 때
**Then** 둘 다 deny 다(추가가 내장을 대체하지 않음). 그리고 판별식 둘:
(1) `grep -n 'commit_identity_guard' internal/template/templates/.moai/config/sections/workflow.yaml` 이
1 건 이상 적중하고 그 블록의 `enabled` 값이 `false` 다;
(2) `grep -n 'commit_identity_guard' .moai/config/sections/workflow.yaml` 이 적중하고 그 블록의
`enabled` 값이 `true` 다. 템플릿 파일의 추가 문언에는 SPEC ID·카드 번호·날짜가 0 건이다.

## AC-CIG-010 — 내장 목록이 저장소 픽스처 열거를 전부 덮는다

**Covers**: maps REQ-CIG-008

**Given** run base 에서 `plan.md` §B.1 의 열거 명령을 실행한 출력(리터럴 집합 S)이 증거로 보존돼 있다
**When** 드리프트 테스트가 같은 술어로 저장소 `*_test.go` 중
`internal/hook/commit_identity_guard*_test.go` 를 제외한 파일을 훑을 때
**Then** 테스트는 통과하고, 내장 목록 ⊇ S 이다. 변이 대조로 내장 목록에서 `t@t.t` 한 줄을 지운
트리에서는 드리프트 테스트가 실패하며 실패 메시지가 `t@t.t` 를 이름으로 지목한다. 드리프트
테스트가 0 개 파일을 훑고 통과하는 경우는 실패로 처리한다(빈 결과집합 통과 금지 — 제외 후
훑은 파일 수 ≥ 1 과 찾은 리터럴 수 ≥ 1 을 단언). 대조 이메일은 제외된 가드 테스트 파일에
존재해도 내장 목록으로 승격되지 않는다.

## AC-CIG-011 — 네 변이가 각각 테스트를 실패시킨다

**Covers**: maps REQ-CIG-011

**Given** 초록인 `go test ./internal/hook/...`
**When** 다음 변이를 하나씩 적용한 트리에서 같은 명령을 돌릴 때 —
(a) 거부 목록 비교를 항상 불일치로 바꾼다,
(b) 트리거 동사 중 하나(각 동사에 대해 따로)를 가드의 트리거 정의에서 뺀다,
(c) 탐침이 이메일을 추출하지 못하도록(빈 문자열 반환) 해석을 깨뜨린다,
(d) 저장소 범위 비교를 상수로 바꾼다 — (d1) 항상 같음(AC-CIG-013 이 잡아야 함), (d2) 항상 다름
(AC-CIG-014 가 잡아야 함)
**Then** 각 변이마다 이름 붙은 테스트가 1 개 이상 FAIL 하며, 증거(변이 설명, 실행 명령, 실패한
테스트 이름, 종료 코드)가 `.moai/reports/t1289/` 아래 판정 증거에 기록된다. (b) 는 8 동사 모두에
대해 기록한다. 변이는 커밋되지 않는다.

## AC-CIG-012 — 배선 순서와 PowerShell 분류 불가 구문

**Covers**: maps REQ-CIG-009, REQ-CIG-010

**Given** 가드가 켜져 있고, (c)는 autonomy `contract` 모드·통합 브랜치 `develop` 설정에서
push 대상 develop 커밋에 추적 중인 카드 SPEC과 보고서 쌍이 들어 있지만 필수 2차 리뷰 증거가
없는 상태다(`checkClosurePush` 가 `second_review_not_performed` 로 거부하는 fixture)
**When** (a) 파괴 명령 거부 목록에도 걸리는 명령, (b) 브랜치 가드에 걸리는 명령,
(c) 통합 잠금 가드 뒤에 배선된 기존 셸 가드 중 push readiness 가 거부하는
`git commit -m x && git push origin develop` 이 픽스처 신원으로
들어올 때, 그리고 (d) PowerShell 도구로 분류 불가 간접 구문(예: 호출 연산자로 계산된 명령 이름)을
통해 커밋하려는 명령이 들어올 때
**Then** (a)(b)(c) 의 거부 사유는 앞선 가드의 것이고 `TEST_IDENTITY_VIOLATION:` 이 아니다;
(d) 는 allow, 범위·신원 탐침 호출 수 각각 0 이며
`.moai/logs/commit-identity-guard-audit.log` 에 분류 불가 구문 한 줄만 추가된다. 또한
`grep -rn '"TEST_IDENTITY_VIOLATION"' internal/hook --include=*.go` 가 비테스트 파일에서 정확히 1 건
(정의 1 곳)을 적중한다.

## AC-CIG-013 — 다른 저장소의 픽스처 신원 커밋은 허용된다 (양성 대조)

**Covers**: maps REQ-CIG-012, REQ-CIG-007

**Given** 가드가 켜져 있고, 프로젝트 디렉터리는 `t.TempDir()` 저장소 P, 명령 대상은 별개의
`t.TempDir()` 저장소 F(`/tmp` 픽스처 저장소 역할, P 와 common dir 이 다름)이며, 실제
`git rev-parse --git-common-dir`·`git var` 탐침을 쓴다
**When** 다음이 각각 들어올 때 — (a) cwd 가 F 인 `GIT_AUTHOR_EMAIL=t@t.t GIT_COMMITTER_EMAIL=t@t.t git commit -m x`,
(b) cwd 가 P 인 `git -C <F 절대경로> -c user.email=t@t.t commit -m x`,
(c) cwd 가 P 인 `cd <F 절대경로> && GIT_AUTHOR_EMAIL=t@t.t git commit -m x`
**Then** 세 경우 모두 결정은 allow 이고 거부 사유가 없으며, 신원 탐침(`git var`) 호출 수는 0 이다
(범위 불일치에서 신원 검사까지 가지 않음). 짝 대조로 같은 명령의 대상을 P 로 바꾼 3 행은 모두
`TEST_IDENTITY_VIOLATION:` 으로 deny 다 — 짝이 있어야 allow 가 범위 판정 덕분임이 선다.

## AC-CIG-014 — 이 저장소의 링크드 워크트리 커밋은 거부된다

**Covers**: maps REQ-CIG-012, REQ-CIG-003

**Given** 가드가 켜져 있고, 프로젝트 디렉터리는 `t.TempDir()` 저장소 P 의 primary 트리, 명령 대상은
테스트 안에서 `git -C <P> worktree add <W>` 로 만든 P 의 링크드 워크트리 W 이며(P 와 W 의
`--git-common-dir` 은 정규화 후 같은 경로), 실제 범위 탐침을 쓴다
**When** 다음이 각각 들어올 때 — (a) cwd 가 W 인 `GIT_AUTHOR_EMAIL=t@t.t git commit -m x`,
(b) cwd 가 P 인 `git -C <W 절대경로> -c user.email=t@t.t commit -m x`
**Then** 두 경우 모두 결정은 deny 이고 사유는 `TEST_IDENTITY_VIOLATION:` 으로 시작하며 `t@t.t` 를
담는다. 테스트는 P 와 W 의 경로 문자열이 서로 **다름**을 함께 단언한다 — 경로가 같으면 링크드
워크트리를 잰 것이 아니다. (macOS `t.TempDir()` 의 `/var` ↔ `/private/var` 별칭이 정규화로 흡수되는지도
이 테스트가 드러낸다.)

## 경계 사례

- 이름만 `t` 이고 이메일이 실주소인 신원(§2.1 부수 관측) — 이메일이 목록 밖이면 allow. 이름 판별은
  범위 밖이다.
- `git commit --amend` — `commit` 동사로 트리거된다.
- `git merge --ff-only` — 커밋을 만들지 않지만 동사 단위 트리거이므로 검사 대상이다. 실신원이면
  통과하므로 정상 작업 비용은 탐침 1 회다.
- 대상 해석이 풀지 못하는 형태(`pushd`, 변수로 계산한 경로, 서브셸 안의 `cd`) — 대상은 cwd 로 남는다.
  cwd 가 이 저장소이고 실제 대상이 다른 저장소면 과탐, 반대면 미탐이 된다. 잔여 위험으로 기록한다.
- 탐침 cwd 가 primary checkout 이고 명령이 `cd <워크트리> && git commit` 인 경우 — D8 해석이 선두
  `cd` 를 대상으로 삼는다. 이 저장소는
  워크트리 전용 설정이 꺼져 있어 신원 해석이 같다. 워크트리 전용 설정이 켜진 저장소에서는 미탐이
  생길 수 있으며 잔여 위험으로 기록한다.

## 채택 셀 (RED-now / green path)

측정 트리 `37dc766b9`. 가드 코드가 아직 없으므로 모든 AC 의 RED-now 사유는 같다 — 판정 대상
테스트와 가드가 부재하다. 이 사유가 **올바른 이유의 RED** 인 근거: 부재는 이 SPEC 의 run 이 만드는
것만으로 해소되며, 이 SPEC 이 건드리지 않는 파일에 기대는 AC 는 없다(AC-CIG-010 의 드리프트 테스트는
기존 `*_test.go` 를 읽기만 하고, 그 리터럴은 이 SPEC 이 내장 목록으로 흡수한다).

| AC | RED-now 판별 (측정 트리 기준) | green 으로 뒤집는 마일스톤 |
|----|-------------------------------|---------------------------|
| AC-CIG-001, AC-CIG-002, AC-CIG-003, AC-CIG-004, AC-CIG-006, AC-CIG-007 | `grep -rn 'TEST_IDENTITY_VIOLATION' internal/hook` → 0 건 | M2 |
| AC-CIG-005 | 같은 판별 + `git var` 탐침 부재 | M2 |
| AC-CIG-008, AC-CIG-009 | `grep -rn 'commit_identity_guard' internal/config .moai/config/sections internal/template/templates/.moai/config/sections` → 0 건 | M1 (+ M3 배선) |
| AC-CIG-010 | 내장 목록 부재 | M1 |
| AC-CIG-011 | 가드·테스트 부재로 변이 대상 없음 | M4 |
| AC-CIG-012 | `pre_tool.go` 에 가드 호출 부재 | M3 |
| AC-CIG-013, AC-CIG-014 | `grep -rn 'TEST_IDENTITY_VIOLATION' internal/hook` → 0 건 (가드 부재; `git-common-dir` 문자열은 형제 코드 7 파일에 이미 있어 RED 판별로 쓰지 않는다) | M2 |

## 품질 게이트 / 완료 정의

- AC-CIG-001..014 전부 PASS, 각 판정은 실제 실행 출력으로 뒷받침.
- `plan.md` §E 의 E1-E8 실행 기록이 `.moai/reports/t1289/` 에 있음.
- `golangci-lint` v2.1.6 0 건, `make build` exit 0.
- `plan.md` §B.3 저장소 범위 판단이 해소된 상태 유지(레인 결정 (b), spec 0.1.1).
