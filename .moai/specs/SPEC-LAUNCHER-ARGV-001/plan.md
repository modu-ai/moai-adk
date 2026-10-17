---
id: SPEC-LAUNCHER-ARGV-001
title: "Implementation plan — launcher worktree-guard argv judgment redesign"
version: "0.1.0"
created: 2026-10-15
updated: 2026-10-15
author: manager-spec (MoAI factory, card t1748)
tier: M
---

# SPEC-LAUNCHER-ARGV-001 구현 계획

## §A Context (문맥)

### §A.1 베이스라인

- 작업 트리: 본 카드 워크트리(브랜치 `WT-launcher-argv-redesign`, main `1ced8cab9`에서 분기, plan-phase 기준 HEAD `1ced8cab9`).
- **선행 가드 구현은 main에 없다.** t1604의 구현 전체(SPEC-FACTORY-DISCIPLINE-001 M1+M2)는 보존 브랜치 `WT-lead-discipline` 팁 `7fe9143d9`에만 있고(읽기 전용 증거 — 본 트리에 쓰지 않는다), 본 트리 HEAD와의 병합 공통 조상은 `842160157f`다. 측정: `factory_anchor_guard_test.go` 부재, 가드 기호 grep 0건 (2026-10-15).
- 따라서 본 SPEC의 구현은 **main 기반에서 판정면을 새로 세우되, 비-argv 의미(재정의 플래그·감사 로그 위생·레인 면제·판별자)는 보존 브랜치의 구현을 이식 원으로 삼는다** — 이식 대 세움의 기계적 전략은 M2에서 정한다. argv 판정면 자체는 보존 브랜치의 병렬 판독기들 대신 REQ-LA-001의 단일 해석기로 대체된다(이식하지 않는 유일한 부분 — 그것이 본 재설계다).

### §A.2 PRESERVE 목록 (범위 밖 보존)

- `internal/cli/codex_review_scope.go`의 `cardScopeFromBranch` — 공유 판별자, 소비자로서만 사용(수정 없음).
- 형제 git 명령 가드 면: BranchGuard PreToolUse 훅, 워크트리 세션 가드 — 비간섭.
- `internal/factory/gtd_operation.go:583` — 지정 표면 밖 선존 결함, 편입 금지.
- `internal/cli/lane_resume.go`의 레인 재개 어셈블러(`laneJoinChildArgv`) — 비리더 인접 면, 기존 argv 규약 보존(AC-LA-014). 기동 경로 호출 클로저 **안**이다(plan-repair 2 실측: 호출 cc.go:273 — 레인 케이스가 공용 꼬리로 낙하) — 다만 worktree 값을 읽지 않는 순수 조립자(파일 안 `worktreeFlagValue` 0건)이므로 뷰 편입 대상이 아니며, AC-LA-013 스캔 집합 안에서 "조립 허용·worktree 값 판독 금지"로 검증된다.
- (plan-repair 1) 구판의 "`codex_hooks_seed.go`·`session_worktree.go`·`launch_chain.go` 판독 강제 이관 금지"는 철회됐다 — 본 트리 실측에서 이 파일들의 worktree 판독자는 기동 경로 위이며 M1의 뷰 편입 대상이다(§A.4 설계 노트).

### §A.3 연결 좌표 (본 트리 plan-phase 실측, 커밋 `1ced8cab9` 기준)

| 좌표 | 내용 |
|---|---|
| `internal/cli/cc.go:186-312` | 진입 파싱 → 리더 케이스(194-231)·레인 케이스(232-280, 양 케이스 모두 공용 꼬리로 낙하) → 공용 꼬리: `resolveWorktreeL2Path`(282) → `resolveWorktreeExistingBranch`(289) → `ccWorktreeWriterPrecheck`(296) → `seedAdmittedWorktreeHooks`(302) → `normalizeWorktreeFlag`(304) → `launch(...)`(312) |
| `internal/cli/glm.go:334,339,347,348` | cc 공용 꼬리의 쌍둥이 — `resolveWorktreeL2Path`(334)·`resolveWorktreeExistingBranch`(339)·`seedAdmittedWorktreeHooks`(347)·`normalizeWorktreeFlag`(348). t1604 기준 가드 삽입점 `glm.go:278` |
| `internal/cli/launcher.go:601` | `runLaunchClaude` — 소비 루프 658-692(`-b`·`-c`·`--model`·`--permission-mode`·`--`), 5b settings 동기화 700, `buildArgs` 735-752, `execOrSpawnClaude` 호출 |
| `internal/cli/launcher.go:917` | `normalizeWorktreeFlag` — `-w`/`--worktree` 철자를 정규 2토큰형으로 |
| `internal/cli/launcher.go:1030` | `worktreeFlagValue` — 구분자 앞 첫 값 판독(가드 아닌 소비자와 공유) |
| `internal/cli/session_worktree.go:1012` | `ccWorktreeWriterPrecheck` — 동시 작성자 사전 점검; `worktreeFlagValue` :1015 (호출 cc.go:296) — plan-repair 1 실측으로 기동 경로 위로 재분류 |
| `internal/cli/worktree_branch_flag.go:57,86,121` | `resolveWorktreeExistingBranch`/`splitWorktreeBranchFlag` — `--branch` 소비·materialize(t295); `worktreeFlagValue` :121 |
| `internal/cli/launch_chain.go:34,58` | `injectChainNodeForLaunch`→`extractWorktreeLaunchTarget` — 정규화 뒤 launch args의 `--worktree` 판독(launcher.go:864 경유) — plan-repair 1 실측으로 기동 경로 위로 재분류 |
| `internal/cli/lane_resume_test.go:19` | `TestLaneJoinChildArgv` — 레인 재개 어셈블러 시험(AC-LA-014 선택자가 명시적으로 닿는다 — plan-repair 1 실측) |
| `internal/cli/worktree_branch_flag.go:58,86` | `resolveWorktreeExistingBranch` / `splitWorktreeBranchFlag` — `--branch` 소비·materialize(t295) |
| `internal/cli/codex_review_scope.go:170` | `cardScopeFromBranch` — 공유 카드 접두 판정 |
| 보존 트리 `internal/cli/factory.go:629-868` | 이식 원: `guardLeaderLaunchAnchor` 배치·`stripAllowCardAnchorFlag`·`enforceLeaderAnchorGuard`·`openLeaderAnchorAuditLog`(위생)·`appendLeaderAnchorAuditRow` |
| 보존 트리 `internal/cli/cc.go:197-213` | 가드의 변이 전 배치 원형(REQ-FD-007 주석 포함) |

### §A.4 설계 노트 — 기동 경로 판독자의 뷰 편입 (plan-audit D5, plan-repair 1, 2026-10-15)

선행 감사(iter 1, FAIL, blocking D5)가 지적한 모순 — REQ-LA-001의 기동 경로 재유도 금지와 Out of Scope의 비가드 소비자 보존의 충돌 — 은 선택지 **(i) 이관**으로 푼다. 근거는 본 트리 실측이다: `worktreeFlagValue`의 비-test 호출자 넷과 정규화 뒤 판독기 하나(`extractWorktreeLaunchTarget`)는 **전부** 기동 경로 위에 있다(§A.3 표; cc·glm의 케이스 스위치가 공용 꼬리 282-312로 낙하하므로 레인 흐름도 같은 꼬리를 지난다). 이 판독자들이 공유하는 `worktreeFlagValue`는 구분자 앞 첫 값만 읽는 모형이라 백엔드의 최후 우선 모형과 어긋나 있다 — 계열 (b)의 모양이 비가드 소비자 안에 이미 살아 있다. 판정면만 뷰로 좁히는 선택지 (ii)는 그 살아 있는 같은-계열 인스턴스를 명명된 예외로 봉인하므로, 여섯 번의 순환을 끝내려는 본 SPEC의 목적과 반대된다. 따라서 다섯 판독자 전부 M1에서 뷰를 소비하도록 재배선하고 Out of Scope는 worktree 값을 읽지 않는 인접 면으로 축소한다(plan-repair 1 시점에는 "호출 클로저 밖"으로 썼으나 plan-repair 2 실측에서 레인 어셈블러가 클로저 안임이 확인돼 정정됐다 — §A.2·§F-M3 참조). 감사가 "session_worktree.go:1015 비기동 흐름"으로 분류한 점은 본 트리에서 정정된다 — :1015는 `ccWorktreeWriterPrecheck`(cc.go:296 호출) 안이다. 다섯 판독자의 현재 판독 도메인은 구분자 앞뿐이다(plan-repair 2 실측: `worktreeFlagValue`는 `--`에서 스캔을 멈춘다 — launcher.go:1026·1033-1035; `extractWorktreeLaunchTarget`은 구분자 뒤 `""` 반환 — launch_chain.go:56-57·:60-61) — M1 재배선의 provenance 구분(구분자 앞 = 비가드 판독자 입력, 구분자 뒤 = M2 가드 판정 입력)은 §F-M1 노트와 acceptance.md §D.4 13열 호환 셀이 묶는다. 판독 해상도 축(plan-audit D8, plan-repair 3): 도메인뿐 아니라 중복 옵션 사이의 값 선택도 오늘과 같아야 한다. `worktreeFlagValue`는 구분자 앞에서 **첫** `-w/--worktree` 값에서 반환하고(launcher.go:1041-1042·1046-1049 — 본 트리 실측), `extractWorktreeLaunchTarget`도 정규화된 args의 첫 일치에서 반환한다(launch_chain.go:64-70). 반면 뷰의 worktree 선택은 백엔드 옵션 모형의 최후 우선이므로 두 모형을 한 필드로 합치면 `-w plain --worktree <허용 접두 밖 경로>`(비리더)가 오늘의 허용에서 거부로 뒤집힌다 — §2 비리더 보존 위반. 해소는 **소비자 범위 첫 값 필드**를 뷰에 따로 두는 것이다: 다섯 비가드 판독자는 이 필드만 소비하고, 백엔드 모형의 최후 우선 선택은 M2 가드 판정 전용으로 남는다. §2에 예외·carve-out을 추가하지 않는다(보드 d-20261015T014443Z-4464, 경로 (i)). 호환 셀 §D.4 14열과 AC-LA-014의 회귀 셀이 이 결정을 묶는다.

## §B Known Issues (알려진 함정)

- **B1 크로스 플랫폼**: 감사 로그 개방의 `O_NOFOLLOW|O_NONBLOCK`은 POSIX 전용 — 보존 브랜치의 `auditLogOpenNoFollow` 빌드 태그 분리(Windows 0) 계승. `GOOS=windows go build` 필수.
- **B2 정책 충돌 사전 점검**: `internal/cli`의 은퇴 SPEC 가드(grep "Retired|superseded") 사전 점검. 본 SPEC은 선행 SPEC의 후속이므로 `related_specs`로 계보를 명시했다.
- **B3 서브에이전트 경계**: 훅/하니스 코드 아님 — 해당 없음(점검은 유지).
- **B4 frontmatter**: `created:`/`updated:`/`tags:` 정규 이름만.
- **B6 소제목 규약**: spec.md의 Out of Scope는 `### Out of Scope — ` 소제목 + `-` 불릿으로(작성 완료).
- **B8 작업 트리 위생**: 커밋은 명시적 pathspec 스테이징만.
- **B9 커밋**: 본 카드는 git-flow 카드 워크플로 — 레인이 카드 브랜치에 커밋, push·PR 없음(레인은 merged-local SHA 보고까지만; 리더 배치가 원격을 처리). Conventional Commits + `🗿 MoAI` + `Authored-By-Agent` 트레일러.
- **B11 서브에이전트 질문 금지**: blocker 보고로만.
- **argv 계열 특유 함정**: ① `:=` 섀도잉 — 리더 케이스에서 `filteredArgs` 재선언 금지(t1604의 cc.go:204 사고 계승, 본 트리 구조에도 동일 위험). ② 감사 로그 개방 경합 — 이식 시 테스트 시암(`leaderAnchorAuditOpenRace`)과 함께 이식. ③ 가드 배치가 변이 뒤로 밀리는 회귀 — M2의 배치 단정 시험이 묶는다.

## §C Pre-flight (구현 전 실행)

- **C-1 백엔드 옵션 모형 재측정 (최우선)** — REQ-LA-003의 모형(최후 우선, 네 철자, 클러스터 분해)은 감사 기록에 근거한 **감사된 모형**이다. 구현 첫 커밋 전에 백엔드 선언 표면(claude CLI 도움말/선언)에서 `--worktree`, `-w`의 값 규약과 묶음 해석을 재확인하고 원문을 run 증거로 남긴다. 측정이 모형과 어긋나면 blocker 보고로 plan-auditor 경유 REQ 정정 — 추측으로 굳히지 않는다.
- **C-2 좌표 재관측** — §A.3 표의 좌표를 구현 시작 시 재확인(파일이 움직였으면 plan 갱신 없이 좌표만 현지화하고 run 증거에 기록).
- **C-3 베이스라인 측정** — `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, 영향 패밀리 베이스라인: `go test ./internal/cli/ -run 'Worktree' -count=1`과 `go test ./internal/cli/ -run 'TestLeader|TestLauncher' -count=1`의 실패 0 확인(신규 결함과 기존 베이스라인 구분).
- **C-4 보존 증거 경로** — 이식 대상 판독은 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1604/` 절대경로 읽기만(쓰기 금지). 본 트리 상대경로 작업 유지.

## §D Constraints (제약)

- spec.md §2 전체. 특히: 비리더 경로 행동 보존(characterize), 템플릿 편집 없음, 새 계수는 `internal/config/defaults.go` 상수, 코드 주석 영어.
- 금지: 원시 진입 argv에서 worktree 선택 재유도(기동 경로), 병렬 판독기 신설, 환경 라벨 판별 입력, `--no-verify`, `git add -A`/`git add .`, 시간 추정.
- `filteredArgs` 재선언(`:=` 섀도잉) 금지 — 함수 수준 변수에 대입.

## §E Self-Verification (마일스톤별 검증)

- 각 마일스톤 종료 시 acceptance.md의 대응 AC 명령을 실행하고 E1-E8 귀속 3조(명령·원문 출력·tree SHA)로 보고한다(`.claude/rules/moai/core/verification-claim-integrity.md` §3 5단 형식).
- M1 종료: AC-LA-005(동일성) 준비 완료 + AC-LA-014(비리더 보존) GREEN.
- M2 종료: AC-LA-001..004·006..012 GREEN + 각 계열 셀의 RED 관측 기록(§D.3 프로토콜).
- M3 종료: AC-LA-013 GREEN(병렬 판독기 삭제 뒤 소스 스캔).
- M4 종료: AC-LA-009 위생 셀 + 크로스 플랫폼 빌드 + 커버리지(85%+) + lint.

## §F Milestones (결정 가역성 순 — 데이터 모형이 맨 앞)

### M1 (Priority High) — 런치 뷰 자료형 + 단일 정준 해석기

- 뷰 구조체(launch view): `finalArgv []string`(exec에 넘겨지는 최종 백엔드 argv), worktree 선택(철자·값·해석된 트리 경로·존재 여부·실효 브랜치 — 백엔드 옵션 모형의 최후 우선 선택이며 M2 가드 판정 전용), **소비자 범위 첫 값 필드**(plan-audit D8, plan-repair 3 — 구분자 앞 첫 `-w/--worktree` 값과 그 유무; `worktreeFlagValue`가 첫 일치에서 반환하는 해상도를 그대로 보존하며 다섯 비가드 판독자 전용), 생성 주석(`--branch` 값), 재정의 플래그 존재.
- 해석기 추출: `runLaunchClaude`의 소비 루프(658-692)를 순수 해석기로 추출 — 런처 소비 옵션 제거, `--` 처리, `normalizeWorktreeFlag` 산물 포함, 백엔드 옵션 모형에 따른 worktree 선택(최후 우선·클러스터 분해·네 철자)을 뷰에 채운다. `runLaunchClaude`는 뷰를 소비하도록 재배선(두 번째 해석 삭제).
- `--branch` 주석: `splitWorktreeBranchFlag`의 판독을 해석기가 재사용(재해석 아님 — 같은 함수 호출)해 주석으로 실고, materialize 실행 자체는 현 위치(cc.go:289)에 그대로 둔다.
- 리더 진입 경로(cc·glm 리더 케이스)는 변이 전에 해석기를 호출해 뷰를 얻고 같은 뷰를 판정·실행에 흘려보낸다(단일 해석 보장 — REQ-LA-001/002). 해석기 호출은 기동당 한 번이며 리더·레인·비팩토리 모든 흐름이 같은 뷰를 소비한다 — 공용 꼬리 판독자의 입력도 이 뷰다.
- 기동 경로 판독자의 뷰 편입(plan-audit D5 — §A.4 설계 노트): 공용 꼬리와 runLaunchClaude 안의 worktree 선택 판독자 다섯 — `resolveWorktreeL2Path`(cc.go:282·glm.go:334), `resolveWorktreeExistingBranch`→`splitWorktreeBranchFlag`(cc.go:289·glm.go:339), `ccWorktreeWriterPrecheck`(cc.go:296), `seedAdmittedWorktreeHooks`(cc.go:302·glm.go:347), `extractWorktreeLaunchTarget`(launch_chain.go:58, launcher.go:864 경유) — 은 raw args 재판독을 끊고 뷰(또는 뷰가 실은 파싱 값: 트리 이름·해석 경로·생성 주석)를 입력으로 받는다. 케이스별 argv 변이(레이블 치환·레인 조인 조립·settings 부착)도 뷰의 finalArgv에 적용해 판정 시점의 선택과 exec 시점의 선택이 갈라지지 않게 한다(AC-LA-005 동일성의 전제).
- 옵션 근원 구분(provenance — plan-audit D7·D8, plan-repair 2·3): 뷰는 구분자 앞(런처 스코프) 선택과 구분자 뒤(백엔드 스코프) 선택을 구분해 싣고, 구분자 앞 영역 안에서도 값 필드를 둘로 나눈다 — ① 백엔드 모형 선택(최후 우선, M2 가드 판정 전용)과 ② 소비자 범위 첫 값 필드(구분자 앞 첫 `-w/--worktree` 값, `worktreeFlagValue`와 같은 해상도). 다섯 재배선 비가드 판독자에게는 **② 첫 값 필드만** 건넨다 — 이들의 현재 도메인은 구분자 앞뿐이고(본 트리 실측: launcher.go:1026·1033-1035, launch_chain.go:56-57·:60-61), 그 도메인 안의 해상도는 첫 일치에서 멈추는 첫 값이기 때문이다(본 트리 실측: launcher.go:1041-1042·1046-1049의 첫 일치 반환, launch_chain.go:64-70의 첫 일치 반환). 도메인 경계는 기존 보존 시험 `TestLauncherWorktreeL2AfterPassThroughMarker`(launcher_worktree_l2_test.go:166 — AC-LA-014 선택자가 닿는다)가 이미 묶으며, 구분자 뒤 백엔드 모형 선택(§D.4 9열)과 중복 옵션의 최후 우선 선택(§D.4 6열)은 M2 가드 판정의 입력으로 남는다. 두 가지 재배선은 금지다. 구분자 뒤 값에 L2 검증·작성자 사전 점검을 새로 적용하면 오늘 허용되는 꼬리 통과가 거부로 바뀌고(호환 셀 §D.4 13열), 판독자에게 ① 최후 우선 선택을 건네면 `-w plain --worktree <허용 접두 밖 경로>`(비리더)의 판독 값이 `plain`에서 `<경로>`로 바뀌어 오늘 허용되는 launch가 거부된다 — 둘 다 비리더 행동 회귀이며 §2 문자 그대로 보존을 어긴다(호환 셀 §D.4 14열, 회귀 셀 AC-LA-014).
- 검증: 비리더 경로 characterize(AC-LA-014) GREEN, 뷰-exec 동일성 시험(AC-LA-005) GREEN.

### M2 (Priority High) — 판정면 교체 + 회귀 셀 (RED→GREEN)

- 가드 신설: 뷰를 입력으로 판정하는 순수 함수 — 기동 디렉터리 브랜치 + 뷰의 worktree 선택(실효 브랜치) + 생성 주석을 `cardScopeFromBranch`로 판정. 리더 케이스 맨 앞(변이 전) 배치 — 보존 브랜치 cc.go:197-213의 배치 원형 이식.
- 비-argv 의미 이식: `--allow-card-anchor` 소비(제 자리 규약), `enforceLeaderAnchorGuard`의 거부/재정의/감사 행, `openLeaderAnchorAuditLog` 위생(Lstat·no-follow·서술자 재판독), 레인·비팩토리 면제. 이식 후 보존 브랜치의 대응 시험 셀을 본 트리 이름 규약으로 옮겨 재생성.
- 회귀 셀 작성 (구속 선택자 — acceptance.md 명령과 일치해야 한다):
  - `TestLauncherAnchorFinalArgv` — 계열 (a) 소비 그늘·(b) 최후 우선·(c) 묶음 클러스터 셀 + 철자표 전체의 뷰-exec 동일성 하위시험. cc·glm 양 팔.
  - `TestLauncherAnchorGuardSemantics` — 변이 전 거부(역할 마커·run 기록·settings 부재), 레인/비팩토리 면제, fail-open, 판별자 환경 독립, 재정의+감사 행, 기록 실패 시 거부.
  - `TestLauncherAnchorAuditHygiene` — symlink/FIFO/서술자 재판독(이식 셀).
  - `TestLauncherAnchorBranchAnnotation` — 기존 트리 실제 브랜치 신뢰 / 신규 생성 주석 브랜치(iter6-F2 보존).
  - `TestLauncherAnchorRunCountInstrument` — 계열 (d): 정준 판독기 기록 계수 + 긍정 대조(기록이 있으면 계수가 움직임을 관측).
- **RED 관측 프로토콜**: (a)-(c) 셀은 **소비 전 argv를 읽는 변형 돌연변이(mutant)** 를 해석기에 잠시 가하고 세 셀 전부가 죽는 것을 관측해 기록(트리 SHA 고정) — 이것이 계열의 재현 원이라 하나의 변형으로 세 셀의 감지력을 증명한다. (d)는 계수기를 레거시 경로로 되돌린 변형으로 죽음을 관측. RED 원문(명령·출력·SHA)은 progress.md §E.2에 남기고 돌연변이를 되돌린다.
- 검증: AC-LA-001..004·006..012 GREEN + RED 기록 완료.

### M3 (Priority Medium) — 기동 경로 판독자 이관 완료 + 단일 해석 소스 스캔

- M1의 뷰 재배선이 기동 경로 판독자 전부를 덮는지 소스 스캔으로 확인한다: 기동 경로 호출 클로저 파일(cc.go, glm.go, launcher.go, codex_hooks_seed.go, session_worktree.go, worktree_branch_flag.go, launch_chain.go, lane_resume.go — 레인 케이스의 `laneJoinChildArgv` 호출 cc.go:273이 공용 꼬리로 낙하하므로 클로저 안; plan-repair 2)에서 해석기와 그 산책 헬퍼 밖의 worktree 값 판독(`worktreeFlagValue` 호출 포함)은 0건이다 — lane_resume.go의 레인 조인 조립(argv 변이)은 허용, 그 안의 worktree 값 판독(재유도)은 금지이며 스캔이 집계한다 — `worktreeFlagValue`의 첫-값 판독기는 기동 경로에서 비-test 호출자 0이 된다. **보존 브랜치의 병렬 판독기 가족(`worktreeScanAdvance`·`worktreeNextValue`·`launcherConsumedSpan`·`segmentClass` 가드부 사용·교차 구분자 짝짓기)은 본 트리에 존재하지 않으므로 새로 만들지 않는 것이 곧 삭제다** — 소스 스캔이 이를 묶는다.
- `TestLaunchViewSingleParse`: (i) 소스 스캔 — 위 클로저 파일 집합에서 해석기 밖의 worktree 값 판독 패턴 0건(스캔 집합은 호출 클로저 기준으로 고정 — codex_hooks_seed.go 같은 헬퍼 파일 뒤에 숨은 판독도 집합 안에서 센다), (ii) 동일성 하위시험 재확인.
- 검증: AC-LA-013 GREEN.

### M4 (Priority Medium) — 품질 게이트 + 인계

- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + 영향 패밀리 `go test ./internal/cli/ -count=1`(전체가 아니라 영향 패밀리 — 카드 트리 전체 스위트는 구조적으로 적색이라는 저장소 실측 존중) + `golangci-lint run` + `go test -cover ./internal/cli/`(85%+).
- sync 인계 노트: CHANGELOG는 manager-docs 소관, codemap 회전 대상 파일 목록.

## §G Anti-Patterns (금지)

- **AP-1 새 병렬 판독기** — 가드가 뷰 밖에서 argv를 읽는 순간 계열이 부활한다. 판정 입력은 뷰뿐.
- **AP-2 백엔드 모형 추측 고정** — C-1 재측정 없이 모형을 굳히면 여섯 번째 계열을 재현한다.
- **AP-3 배치 이동 회귀** — 가드를 변이 뒤(exec 직전)로 옮기면 "변이 전 거부" 의미가 깨진다. 판정 입력이 최종 argv라는 것과 판정 시점이 변이 전이라는 것은 별개 요구다(REQ-LA-002+004).
- **AP-4 환경 라벨 판별** — 브랜치 접두만(REQ-LA-006).
- **AP-5 형제 면 흡수** — git 명령 가드·gtd_operation.go:583 편입 금지.
- **AP-6 약화** — 거부 조건 축소, 재정의 플래그 묵시화, 감사 행 생략 금지.
- **AP-7 전체 스위트 로컬 실행** — 카드 트리 전체 `go test ./...` 금지(저장소 실측: 구조적 적색·부하 사고 계승). 영향 패밀리 + 원격 CI.
- **AP-8 GEARS legacy** — 산출물에 `IF/THEN` 모달 금지(When-event-detected 사용).

## §H Cross-References

- 선행: SPEC-FACTORY-DISCIPLINE-001(보존 트리 `.moai/worktrees/t1604/.moai/specs/`) — REQ-FD-007..010이 본 SPEC §1.2의 보존 원.
- 감사 증거: 보존 트리 `.moai/reports/t1604/sync-audit-iter8.md`(영수증 `rcpt-649bc5e11cf68dfd88d64c14`)·`operator-brief.md`·`sync-audit-iter7.md`.
- 교리: `.claude/rules/moai/core/verification-claim-integrity.md`(증거 귀속)·`.claude/rules/moai/development/verification-completeness.md`(2셀 채택)·`.claude/rules/local/gitflow-lane-protocol.md`(착지).
- 운영자 결정: `d-20261014T143613Z-4764`(재설계 방향 1·2 기록).
