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
- `codex_hooks_seed.go`·`session_worktree.go`의 `worktreeFlagValue` 호출 — 가드 아닌 소비자, 강제 이관 금지(Out of Scope).
- `internal/cli/launch_chain.go`·`lane_resume.go`의 기존 argv 규약 — 본 SPEC이 요구하는 것은 기동 경로(launch path)의 단일 해석뿐; 레인 재개 어셈블러 등 인접 면의 기존 판독은 건드리지 않는다.

### §A.3 연결 좌표 (본 트리 plan-phase 실측, 커밋 `1ced8cab9` 기준)

| 좌표 | 내용 |
|---|---|
| `internal/cli/cc.go:186-312` | 진입 파싱 → 리더 케이스(194-231) → `resolveWorktreeL2Path`(282) → `resolveWorktreeExistingBranch`(289) → `ccWorktreeWriterPrecheck`(296) → `seedAdmittedWorktreeHooks`(302) → `normalizeWorktreeFlag`(304) → `launch(...)`(312) |
| `internal/cli/glm.go` | cc의 쌍둥이 흐름(t1604 기준 가드 삽입점 `glm.go:278`) |
| `internal/cli/launcher.go:601` | `runLaunchClaude` — 소비 루프 658-692(`-b`·`-c`·`--model`·`--permission-mode`·`--`), 5b settings 동기화 700, `buildArgs` 735-752, `execOrSpawnClaude` 호출 |
| `internal/cli/launcher.go:917` | `normalizeWorktreeFlag` — `-w`/`--worktree` 철자를 정규 2토큰형으로 |
| `internal/cli/launcher.go:1030` | `worktreeFlagValue` — 구분자 앞 첫 값 판독(가드 아닌 소비자와 공유) |
| `internal/cli/worktree_branch_flag.go:58,86` | `resolveWorktreeExistingBranch` / `splitWorktreeBranchFlag` — `--branch` 소비·materialize(t295) |
| `internal/cli/codex_review_scope.go:170` | `cardScopeFromBranch` — 공유 카드 접두 판정 |
| 보존 트리 `internal/cli/factory.go:629-868` | 이식 원: `guardLeaderLaunchAnchor` 배치·`stripAllowCardAnchorFlag`·`enforceLeaderAnchorGuard`·`openLeaderAnchorAuditLog`(위생)·`appendLeaderAnchorAuditRow` |
| 보존 트리 `internal/cli/cc.go:197-213` | 가드의 변이 전 배치 원형(REQ-FD-007 주석 포함) |

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

- 뷰 구조체(launch view): `finalArgv []string`(exec에 넘겨지는 최종 백엔드 argv), worktree 선택(철자·값·해석된 트리 경로·존재 여부·실효 브랜치), 생성 주석(`--branch` 값), 재정의 플래그 존재.
- 해석기 추출: `runLaunchClaude`의 소비 루프(658-692)를 순수 해석기로 추출 — 런처 소비 옵션 제거, `--` 처리, `normalizeWorktreeFlag` 산물 포함, 백엔드 옵션 모형에 따른 worktree 선택(최후 우선·클러스터 분해·네 철자)을 뷰에 채운다. `runLaunchClaude`는 뷰를 소비하도록 재배선(두 번째 해석 삭제).
- `--branch` 주석: `splitWorktreeBranchFlag`의 판독을 해석기가 재사용(재해석 아님 — 같은 함수 호출)해 주석으로 실고, materialize 실행 자체는 현 위치(cc.go:289)에 그대로 둔다.
- 리더 진입 경로(cc·glm 리더 케이스)는 변이 전에 해석기를 호출해 뷰를 얻고 같은 뷰를 판정·실행에 흘려보낸다(단일 해석 보장 — REQ-LA-001/002).
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

### M3 (Priority Medium) — 병렬 판독기 정리 + 단일 해석 소스 스캔

- 기동 경로(cc/glm → launch → runLaunchClaude)에서 뷰 바깥의 worktree 재유도 제거: 본 트리에 존재하는 이중 판독(해석기 도입으로 생긴 중복 호출) 정리. **보존 브랜치의 병렬 판독기 가족(`worktreeScanAdvance`·`worktreeNextValue`·`launcherConsumedSpan`·`segmentClass` 가드부 사용·교차 구분자 짝짓기)은 본 트리에 존재하지 않으므로 새로 만들지 않는 것이 곧 삭제다** — 소스 스캔이 이를 묶는다.
- `TestLaunchViewSingleParse`: (i) 소스 스캔 — 기동 경로 파일에서 해석기 바깥의 worktree 값 판독 패턴 0건, (ii) 동일성 하위시험 재확인.
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
