---
id: SPEC-COMMIT-IDENTITY-GUARD-001
title: "커밋 시점 테스트 신원 가드 — 테스트 픽스처 신원으로 커밋을 만드는 셸 명령을 PreToolUse 에서 거부한다"
version: "0.1.4"
status: draft
created: 2026-09-28
updated: 2026-09-28
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook,internal/config"
lifecycle: spec-anchored
tier: M
tags: "hook, pretooluse, guard, git-identity, commit, test-fixture, fail-open"
---

# SPEC-COMMIT-IDENTITY-GUARD-001 — 커밋 시점 테스트 신원 가드

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-28 | manager-spec | 최초 plan-phase 초안 (card t1289). 원인 규명 `.moai/reports/t1289/root-cause.md` 의 측정을 입력으로 삼고, 레인이 이미 내린 설계 결정 7건(§4)을 결정으로 기록했다 |
| 0.1.1 | 2026-09-28 | manager-spec | 레인 결정으로 저장소 범위 `[NEEDS CLARIFICATION]` 해소 — (b) 이 프로젝트와 git common dir 을 공유하는 저장소만 거부(D8). REQ-CIG-010 추가, REQ-CIG-006 실패 원인에 common dir 해석 실패 추가, AC-CIG-013·014 추가 |
| 0.1.2 | 2026-09-28 | manager-spec | 계획 감사 D1-D9 해소: 열거 제외, 탐침 게이트, 역할별 재정의 우선순위, 원문 추출, 배선·감사 경로를 일치시킴 |
| 0.1.3 | 2026-09-28 | manager-spec | 2차 계획 감사 D14 해소: REQ-CIG-001·008·010을 각 한 가지 GEARS 패턴의 단일 응답으로 정리 |
| 0.1.4 | 2026-09-28 | manager-spec | D14 분할 완료(분할, 압축 금지): REQ-CIG-001을 분류(상시)와 트리거 탐침(이벤트) 두 REQ로, REQ-CIG-008을 배선 순서(상시)와 분류 불가 PowerShell 구문(이벤트) 두 REQ로 분리해 REQ-CIG-001..012 연속 번호로 재배치. AC Covers 재귀속, REQ-CIG-010(구)은 단일 State-driven으로 확인해 분할 없음 |

## 1. 문제

develop 이력에서 2026-08-15 부터 09-22 까지 **5,483 건**의 커밋이 author·committer 모두
`t <t@t.t>` — 이 저장소 테스트가 임시 저장소에 쓰는 픽스처 신원 — 으로 기록됐다. 신원이
세션 하나의 환경 변수가 아니라 **모든 레인이 읽는 공유 git 설정층**에 들어갔기 때문에, 서로
다른 워크트리·레인의 커밋이 분 단위로 섞여 전부 같은 가짜 신원을 달았다.

오염이 멈춘 것도 코드 수리가 아니라 공유 설정이 우연히 다시 덮어써진 사건이었다(§2.2).
즉 지금 이 순간 같은 오염이 다시 일어나도 **커밋을 막는 장치는 하나도 없다.** 쓰기 통로를
고치는 일(§7)과 별개로, 커밋이 만들어지는 순간에 신원을 확인하는 마지막 방어선이 필요하다.

이 SPEC 은 그 방어선 — 기존 PreToolUse 가드 계열에 붙는 **커밋 시점 신원 가드** — 를 정의한다.

## 2. 측정 근거 (원인 규명 보고서에서 옮김)

출처: `.moai/reports/t1289/root-cause.md`. 측정 트리 워크트리 `WT-commit-identity-guard`,
HEAD `37dc766b9`(= 당시 로컬 develop), 측정일 2026-09-28. 아래 수치는 그 보고서가 명령과 함께
기록한 값이며, 이 문서가 새로 잰 것이 아니다.

### 2.1 분류

- `git log develop --format=...` 전체 **11,005** 건 중 author 이메일이 `t@t.t` 인 커밋 **5,483** 건.
- 5,483 건 모두 author = committer = `t <t@t.t>`. 둘이 갈라진 커밋 **0** 건 — 두 신원이 같은
  출처(설정 파일의 `user.name`/`user.email`)에서 해석됐다.
- 첫 커밋 `ea3d96266` 2026-08-15T20:47:16+09:00, 마지막 `0d8d543ee` 2026-09-22T16:10:45+09:00.
- 09-22 14:00~16:10 한 구간에 서로 다른 SPEC·워크트리의 커밋이 분 단위로 섞여 모두 `t@t.t` —
  세션 단위 환경 변수로 설명되지 않는다.
- 부수 관측: `t <goos@afamily.kr>` 7 건, `t <namgoos@gmail.com>` 3 건 — 이름과 이메일이 서로 다른
  층에서 해석된 흔적. **이름만으로는 판별이 서지 않는다**는 근거이며, 가드가 이메일을 판별 키로
  쓰는 이유다(§4 D4).

### 2.2 전환점

`0d8d543ee` 16:10:45 (`t@t.t`) → `8f0d658a3` 16:16:03 (실신원). 그 6 분 사이 develop 커밋에
코드 변경이 없다. 누출을 멈춘 것은 수리가 아니라 공유 설정의 재기록이다.

### 2.3 쓰기 통로 (확인된 코드 경로, 동작은 가설)

- 통로 A — `internal/cli/session_worktree.go` `gitConfigSetReal` 이 `--worktree` 없이
  `git config` 를 호출해, `extensions.worktreeConfig` 가 꺼진 이 저장소에서는 값이 모든
  워크트리가 공유하는 `.git/config` 에 들어간다.
- 통로 B — `internal/cli/precommit_relocation_test.go` 가 임시 저장소에 `git config user.email t@t.t`
  를 **설정 파일로** 쓴다. 부모 환경에 `GIT_DIR` 이 남아 있으면 대상이 진짜 저장소가 된다.
  그 선행 조건은 t1232 로 이미 막혔다.

두 통로 모두 **이 SPEC 의 범위 밖**이다(§7). 이 SPEC 은 통로가 무엇이든 결과(가짜 신원 커밋)를
막는 층만 다룬다.

## 3. 목표

- 커밋을 만드는 셸 명령이 알려진 테스트 픽스처 신원으로 실행되려 할 때, 그 명령이 실행되기
  **전에** 거부한다.
- 거부는 오케스트레이터가 문구를 파싱하지 않고 식별할 수 있는 센티널로 시작한다.
- 판별에 확신이 없으면 막지 않고 통과시키되, 통과시킨 사실을 감사 로그에 남긴다.
- 형제 가드와 같은 opt-in 모양을 따른다 — 배포 기본값은 꺼짐, 이 저장소 로컬 설정은 켬.

## 4. 결정 (레인이 내린 설계 결정 — 근거와 함께 기록)

### D1 — 위치: `internal/hook` 의 PreToolUse 셸 가드, 센티널 `TEST_IDENTITY_VIOLATION:`

셸 도구(Bash·PowerShell) 호출은 이미 PreToolUse 가드 계열(`BRANCH_GUARD_VIOLATION`,
`INTEGRATION_LOCK_VIOLATION` 등)이 검사하는 자리다. 새 층을 만들지 않고 그 계열에 한 칸을 더
붙이면 배선·설정·감사 로그 관례를 그대로 물려받는다. 센티널 접두사는 형제들과 같은
`<NAME>_VIOLATION` 모양이며, 측정 시점 `internal/hook` 에 같은 이름은 없다
(`grep -rhoE '"[A-Z_]+_VIOLATION"' internal/hook --include='*.go' | sort -u` → 14 종, `TEST_IDENTITY` 없음).

git 의 pre-commit 훅이 아니라 PreToolUse 에 두는 이유: 이 사고의 커밋은 에이전트의 셸 도구
호출에서 나왔고, git 훅은 설치 여부가 저장소·클론마다 다르며 오염된 공유 설정이 훅 설정까지
건드릴 수 있다. PreToolUse 는 모든 레인 세션에 같은 바이너리로 붙는다.

### D2 — 트리거: 커밋을 만드는 git 동사, 기존 정규화 도우미 재사용

트리거 동사는 `commit`, `merge`, `cherry-pick`, `revert`, `rebase`, `am`, `commit-tree`, `pull`
이다. `merge`(비 fast-forward)·`pull`(병합 경로)은 커밋을 만들 수 있으므로 포함한다. 명령 문자열
정규화는 `branch_guard.go` 의 기존 도우미(따옴표 구간 치환, heredoc 본문 치환, 셸 주석 치환,
`.exe` 접미사 정규화)를 재사용한다. 트리거 판정에는 정규화된 문자열을 쓰고, 재정의와 대상 경로는
트리거가 확인된 git 실행 세그먼트 및 그 실행에 직접 이어지는 선두 `cd ... &&`·
`export ... &&` 접두부의 **원문**에서만 추출한다. 다른 명령의 재정의·경로는 섞지 않는다.
따옴표를 치환한 문자열에는
`--author="t <t@t.t>"` 의 이메일이 남지 않기 때문이다. git 전역 옵션(`-C <path>`,
`-c <key=value>`, `--git-dir` 등)을 건너뛰어 동사를 찾는 최소 로직과 원문 추출은 새로 필요하다.
미탐(under-match)은 받아들이는
fail-open 방향이다 — 과탐은 정상 작업을 막고, 미탐은 기존 상태(가드 없음)로 돌아갈 뿐이다.

### D3 — 신원 해석: 명령의 작업 디렉터리에서 git 이 **실제로 쓸** 신원을 묻는다

명령의 대상 디렉터리에서 `git var GIT_AUTHOR_IDENT` 와 `git var GIT_COMMITTER_IDENT` 를 실행해
설정층과 훅 환경에서 해석한 기본 author·committer 신원을 얻는다. 명령 문자열 안의 재정의는
트리거가 확인된 git 실행 세그먼트와 직접 이어지는 접두부의 원문에서 읽어 해당 역할의
기본값을 **대체**한다. author 는
`--author` 또는 `GIT_AUTHOR_EMAIL`, committer 는 `GIT_COMMITTER_EMAIL` 이 직접 지정하며,
`git -c user.email` 은 더 높은 우선순위의 역할별 값이 없는 역할을 대체한다. `EMAIL` 은 그
역할에 `user.email` 설정과 더 높은 우선순위의 재정의가 없을 때만 쓰는 최후 값이다. 충돌하는
재정의는 git 의 실제 우선순위로 해석하고 가려진 값은 판정하지 않는다. 따라서 설정에 정상
`user.email` 이 있으면 명령의 `EMAIL=t@t.t` 만으로 거부하지 않는다. 이 우선순위를 판별할
수 없는 경우는 불확실한 역할을 허용하고 감사한다(D5). 탐침은 테스트에서 대체할 수 있는
seam 뒤에 둔다. 실제로 적용될 거부 이메일이 명령 원문에서 확정되면 탐침이 실패하더라도
그 양성 증거로 거부한다. 거부 사유의 처치는 그 역할의 실제 설정 또는 명령 재정의를 고치도록
안내한다.

### D4 — 거부 목록: 정확 일치 이메일, 설정으로 확장, 내장 기본값은 이 저장소 픽스처 전수

판별 키는 이메일이다(§2.1 부수 관측 — 이름은 다른 층에서 따로 해석될 수 있다). 비교는 앞뒤
공백을 걷고 대소문자를 무시한 **정확 일치**이며, 정규식·도메인 패턴 같은 휴리스틱은 쓰지 않는다.
내장 기본 목록은 이 저장소 `*_test.go` 가 실제로 쓰는 픽스처 이메일 리터럴 전수(가드 자신의
`internal/hook/commit_identity_guard*_test.go` 는 양성·음성 대조값 때문에 제외)이고, 열거 명령과
plan-time 측정값은 `plan.md` §B 에 있다. 설정 키 `workflow.commit_identity_guard.deny_emails` 는
내장 목록에 **더하는** 추가 목록이다(내장 목록을 줄이지 않는다).

### D5 — 활성화와 실패 방향: 형제 가드와 같은 opt-in, 불확실하면 통과 + 감사 기록

`workflow.commit_identity_guard.enabled` 의 배포 기본값은 `false`(`internal/config/defaults.go`
가 강제), 이 저장소 로컬 설정은 `true`. 꺼져 있으면 저장소 범위·신원 탐침 서브프로세스가
하나도 돌지 않는다.
신원 해석 실패(탐침 오류·시간 초과·출력 파싱 실패·cwd 부재)는 같은 저장소에서 이미 확정된
명령 수준 거부 이메일이 없을 때 통과로 끝나며, 그 사실을
`.moai/logs/commit-identity-guard-audit.log` 에 한 줄로 남긴다(`branch-guard-audit.log` 와 같은
advisory 관례). 저장소 범위 해석 실패는 명령 재정의가 있어도 대상 저장소를 확정할 수 없으므로
통과·감사한다.

### D6 — 이빨: 변이 증거를 인수 기준에 넣는다

거부 검사 제거, 트리거 동사 제거, 신원 해석 파손, 저장소 범위 비교 상수화(항상 같음·항상 다름)의
네 변이 각각이 이름 붙은 테스트 하나 이상을 실패시켜야 한다. 실신원이 통과하는 양성 대조도
함께 둔다. 동작하지 않는 가드가 초록불을 내는
상태(공허 초록)를 막기 위해서다.

### D7 — Template-First

설정 키를 더하므로 `internal/template/templates/.moai/config/sections/workflow.yaml` 에 중립
문언의 키(`enabled: false`)를 싣고, `make build` 를 검증에 넣는다. 템플릿 문언에는 SPEC ID·
카드 번호·날짜를 넣지 않는다(템플릿 중립성 규율). 형제 가드 일부가 템플릿 yaml 에 키를 싣지
않는 것은 알고 있으며, 이 SPEC 은 배포 사용자가 키의 존재를 설정 파일에서 발견할 수 있도록
싣는 쪽을 택한다.

### D8 — 저장소 범위: 이 프로젝트와 git common dir 을 공유하는 저장소만 거부

가드는 명령의 대상 저장소가 이 프로젝트와 **같은 git common dir** 을 공유할 때만 거부한다. 판별은
명령의 대상 디렉터리에서 `git rev-parse --git-common-dir` 를 실행해 절대경로로 정규화한 값과, 훅의
프로젝트 디렉터리에서 같은 명령을 실행해 정규화한 값을 비교한다. 같으면(이 저장소 자체이거나 그
링크드 워크트리) 신원 검사를 진행하고, 다르면(`/tmp` 픽스처 저장소 등) 허용한다. 어느 한쪽이라도
해석에 실패하면 허용하고 감사 한 줄을 남긴다(D5 와 같은 fail-open).

대상 디렉터리는 훅 입력의 cwd 를 기본으로 하되, 트리거가 확인된 실행과 직접 이어지는 원문에
리터럴 `git -C <path>` 나 선두
`cd <path> &&` 가 있으면 그 경로를 쓴다(상대경로는 cwd 기준으로 절대화). 이 해석으로 풀 수 없는
형태는 cwd 로 남으며, 그 결과의 미탐·과탐 방향은 잔여 위험으로 기록한다. 신원 탐침(D3)도 같은 대상
디렉터리에서 실행한다.

근거(레인 결정): 버려지는 임시 저장소에서 픽스처 신원으로 커밋하는 것은 정당한 테스트 활동이다.
측정된 누출은 이 저장소의 공유 설정층에 한정됐다(§2.1). 그 밖의 저장소를 거부해도 보호 가치 없이
실제 작업만 막는다.

## 5. 요구사항 (GEARS)

- REQ-CIG-001 (Ubiquitous): The commit identity guard shall classify each shell call by its normalized command and shall allow a call that invokes no commit-creating git verb without repository-scope or identity probes.
- REQ-CIG-002 (Event-driven): **When** a shell call invokes a commit-creating git verb (`commit`, `merge`, `cherry-pick`, `revert`, `rebase`, `am`, `commit-tree`, `pull`), the commit identity guard shall probe the call's target repository scope and evaluate its identity per REQ-CIG-012's common-dir comparison.
- REQ-CIG-003 (Event-driven): **When** the resolved author email or the resolved committer email exactly matches (whitespace-trimmed, case-insensitive) an entry of the effective deny list, the commit identity guard shall return a deny decision whose reason begins with `TEST_IDENTITY_VIOLATION:` and names the matched email, the identity role (author or committer), and a remedy.
- REQ-CIG-004 (Event-driven): **When** the triggered git invocation or its directly chained `export ... &&` prefix carries an inline or exported `GIT_AUTHOR_EMAIL=` / `GIT_COMMITTER_EMAIL=` / `EMAIL=`, a `git -c user.email=` option, or an `--author=` option, the commit identity guard shall extract that value from the original text, replace only the affected role's probed email according to D3's git precedence, ignore masked values, and deny an effective deny-listed override even if a later identity probe fails.
- REQ-CIG-005 (Ubiquitous): The commit identity guard shall allow a commit-creating command whose effective author and committer emails both fall outside the effective deny list.
- REQ-CIG-006 (Capability gate): **Where** `workflow.commit_identity_guard.enabled` is false or absent, the pre-tool handler shall not invoke the commit identity guard, and no repository-scope or identity probe subprocess shall run.
- REQ-CIG-007 (Event-driven): **When** identity resolution or repository-scope resolution fails — probe error, probe timeout, unparseable probe output, missing working directory, or a `git rev-parse --git-common-dir` failure on either the target side or the project side — and no effective deny-listed command-level override has already been established in an equal-common-dir target, the commit identity guard shall allow the command and append one advisory line naming the cause to `.moai/logs/commit-identity-guard-audit.log`.
- REQ-CIG-008 (Ubiquitous): The effective deny list shall be the union of the built-in fixture list and `workflow.commit_identity_guard.deny_emails`; the built-in list shall contain every fixture email literal the repository's `*_test.go` files assign through `user.email`, `GIT_AUTHOR_EMAIL`, or `GIT_COMMITTER_EMAIL`, excluding `internal/hook/commit_identity_guard*_test.go` because those tests carry out-of-list controls, and an in-repository test shall fail when a literal found by that same nonempty enumeration is missing from the built-in list.
- REQ-CIG-009 (Ubiquitous): The commit identity guard shall handle shell calls after the destructive-command check and all existing shell-tool guards in the pre-tool handler, preserving an earlier deny.
- REQ-CIG-010 (Event-driven): **When** a PowerShell indirection construct cannot be classified as a trigger, the commit identity guard shall allow the command without repository-scope or identity probes and append exactly one unclassified-command line to `.moai/logs/commit-identity-guard-audit.log`.
- REQ-CIG-011 (Ubiquitous): The guard's test suite shall fail under each of four mutations — removal of the deny-list comparison, removal of any single trigger verb, breakage of identity resolution, and replacement of the repository-scope comparison by a constant (always-equal or always-different) — and shall carry a positive control in which a non-fixture identity passes.
- REQ-CIG-012 (State-driven): **While** the triggered git invocation's target directory (the hook input cwd, or the path named by a literal `git -C <path>` or a directly chained leading `cd <path> &&` in the original text) resolves to a git common dir different from the git common dir of the hook's project directory, both compared as absolute cleaned paths, the commit identity guard shall allow the command regardless of its identity and without an identity probe.

## 6. 제약

- 가드는 판별 불가 상황에서 절대 거부하지 않는다(fail-open). 거부는 목록 일치라는 양성 증거가
  있을 때만 난다.
- 범위·신원 탐침은 트리거가 맞은 명령에서만 돈다. 트리거가 아닌 셸 호출의 비용은 문자열
  정규화뿐이다(분류 불가 PowerShell 구문의 감사 기록은 예외).
- 탐침 서브프로세스는 시간 상한을 갖는다(기본 설계값은 `plan.md` §F 에 있다). 상한 초과는
  REQ-CIG-007 의 실패로 취급한다.
- 탐침은 `internal/gitenv` 로 저장소 범위 git 환경 변수를 걷어낸 환경에서 실행한다 — 훅 프로세스에
  남은 `GIT_DIR` 이 탐침을 다른 저장소로 돌리지 않게 하기 위해서다. 단, 신원을 정하는
  `GIT_AUTHOR_*`/`GIT_COMMITTER_*`/`EMAIL` 은 걷어내지 않는다(그 값이 곧 git 이 쓸 신원이다).
- 로컬 전체 스위트 실행 금지. 검증 범위는 `plan.md` §E 가 정한다.

## 7. 범위 밖

### Out of Scope — 공개 이력 재작성

- 이미 develop·원격에 들어간 5,483 건의 `t <t@t.t>` 커밋을 고치는 일(rebase·filter-repo·
  mailmap 포함) — 이 SPEC 은 앞으로의 커밋만 다룬다.

### Out of Scope — 쓰기 통로 수리 (후속 카드 후보)

- 통로 A: `internal/cli/session_worktree.go` `gitConfigSetReal` 이 신원을 공유 `.git/config` 에
  쓰는 동작의 수정(`--worktree` 사용 또는 복사 중단) — 사용자 표면 동작 변경이라 별도 카드 후보.
- 통로 A 부수 결함: `gitSafeDirAddReal`(`--global --add`)의 `~/.gitconfig` `safe.directory`
  누적(측정 35,306 줄) — 별도 카드 후보.
- 통로 B: `internal/cli/precommit_relocation_test.go` 가 신원을 설정 파일에 쓰는 방식을 `cmd.Env`
  주입으로 바꾸는 일 — 별도 카드 후보.
- 이미 오염돼 있을 수 있는 공유 설정층의 정리.

### Out of Scope — 셸 도구 밖의 커밋 경로

- `go test` 하위 프로세스, 사람이 연 터미널, IDE, git 자체 훅 등 에이전트 셸 도구 호출을 거치지
  않는 커밋 — PreToolUse 는 그 경로를 보지 못한다. 원격 쪽(pre-receive·브랜치 보호) 신원 검사도
  범위 밖이다.

### Out of Scope — 이름 기반·패턴 기반 판별

- `user.name` 으로 판별하거나, 도메인 패턴(`example.*`, `*.invalid`)·정규식으로 목록을 넓히는 일.

## 8. 성공 기준

- 이 저장소(링크드 워크트리 포함)를 대상으로 한 트리거 명령 × 픽스처 신원(설정층 경유·명령 수준
  재정의 경유)이 모두 `TEST_IDENTITY_VIOLATION:` 로 거부되고, 실신원과 다른 저장소(`/tmp` 픽스처
  저장소)의 픽스처 신원 커밋은 통과한다.
- 가드가 꺼져 있을 때와 트리거가 아닐 때 탐침이 0 회 실행된다(seam 호출 계수로 확인).
- 네 변이(범위 비교 상수화 포함)가 각각 이름 붙은 테스트를 실패시킨다.
- 내장 목록이 저장소 픽스처 열거 결과를 전부 포함하며, 새 리터럴이 들어오면 테스트가 잡는다.
- `make build` 통과, 템플릿 문언 중립, 변경 패키지 테스트·`golangci-lint`(v2.1.6) 초록.

## 9. 의존·선행 사례

- `internal/hook/branch_guard.go` — 센티널 모양, 정규화 도우미, advisory 감사 로그 관례.
- `internal/hook/integration_lock_guard.go` — 작은 가드의 구조와 fail-open 서술 관례.
- `internal/hook/pre_tool.go` — `branchGuardEnabled` / `integrationLockEnabled` 호출부 배선.
- `internal/hook/powershell_indirection.go` — 분류 불가 PowerShell 구문의 감사 기록 관례.
- `internal/gitenv` — 탐침 환경 정리.
- `.moai/reports/t1289/root-cause.md` — 이 SPEC 의 측정 근거.
