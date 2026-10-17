---
id: SPEC-LAUNCHER-ARGV-001
title: "Launcher worktree-guard argv judgment — one canonical resolver, exec-boundary judgment"
version: "0.1.0"
status: completed
created: 2026-10-15
updated: 2026-10-18
author: manager-spec (MoAI factory, card t1748)
priority: P1
phase: "v3.3.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "launcher, worktree, argv, guard, factory, leader, exec-boundary"
tier: M
related_specs: [SPEC-FACTORY-DISCIPLINE-001]
---

# SPEC-LAUNCHER-ARGV-001 — 런처 워크트리 가드 argv 판정면 재설계

## HISTORY

| 날짜 | 변경 | 근거 |
|---|---|---|
| 2026-10-15 | plan-phase 최초 작성 (lane-22 카드 t1748, Class C). 선행 카드 t1604(SPEC-FACTORY-DISCIPLINE-001)의 iter8 감사 FAIL 뒤 운영자 결정 `d-20261014T143613Z-4764`로 큐된 재설계 카드다. 판정면 방향을 두 기록 방향의 종합으로 확정하고, 기록된 실패 계열 (a)–(d)를 명명된 회귀 기준으로 편입했다 | 카드 본문 + 보존 트리 `.moai/worktrees/t1604/`의 `sync-audit-iter8.md` (영수증 `rcpt-649bc5e11cf68dfd88d64c14`)·`operator-brief.md`·`internal/cli/factory.go` 가드부 판독 + 본 트리 실측 |
| 2026-10-18 | run/sync 중 REQ 정정: REQ-LA-003의 `--` 구분자 문장을 백엔드 실측에 맞게 정정했다. plan 감사 상한 소진으로 plan-auditor 델타 없이 sync 감사 2회차가 판정한다. REQ/AC id 집합은 불변이다(REQ-LA-001..013, AC-LA-001..015) | 실측 원문(claude 2.1.302): 명령 `timeout 25 claude --append-system-prompt -- --version </dev/null` → 출력 `2.1.302 (Claude Code)`, exit code 0 — 값 필수 옵션 뒤의 `--`가 그 옵션의 값으로 소비되고 파싱이 이어진다. 발단은 sync 감사 1회차(`.moai/reports/t1748/sync-audit-iter1.md`) 결함 F1 |

## §0 배경 (측정된 사실)

### §0.1 여섯 번의 같은 계열 순환

선행 카드 t1604는 런처 수준 리더 앵커 가드(팩토리 리더 기동이 카드 워크트리에 앵커되는 것을 거부)를 구현했고, 싱크 감사 여섯 회차가 **같은 argv 해석 계열의 여섯 개 경로**를 연속으로 지적했다 (보존 트리 `sync-audit-iter8.md` "What iter8 confirms" 판독):

> attached F7 → empty-equals O27 → launcher-consumed iter5-F1 → segment iter6-F1 → cross-separator iter7-P2 → consumed-option/cluster/multi-flag iter8

근본 원인은 하나다. 가드가 **실행 전 원시 인수(argv)** 를 자신만의 손으로 쓴 판독기들로 재해석하며, 런처가 무엇을 소비하고 백엔드가 무엇을 선택할지 **예측**하려 했다. 예측 모형은 매 회차 한 철자씩 어긋났고, 수리는 그 철자만 맞췄다.

### §0.2 iter8이 확정한 네 실패 계열

보존 트리 `sync-audit-iter8.md` (판정 대상 `4d07897716`, 영수증 `rcpt-649bc5e11cf68dfd88d64c14`, codex 정적 검토)의 네 건 P2:

- **(a) 소비 옵션 그늘** — `moai cc -f -w -b -- card-tree`: 구분자 앞 마지막 토큰만 읽는 교차 구분자 규칙은 `-b`를 읽고, 런처가 `-b`와 `--`을 소비한 뒤 claude는 `--worktree card-tree`를 받는다. 감사자 결론: "The check must reflect the arguments the launch CONSUMES."
- **(b) 다중 worktree 옵션 최후 우선** — `-w plain-tree --worktree --branch main card-tree`: `--branch main` 제거 뒤 claude는 worktree 옵션 두 개를 받고 마지막(card-tree)이 이긴다. `--branch` 다리는 첫 값만 검사했다.
- **(c) 묶음 짧은 옵션** — `moai cc -- -pwcard-tree <프롬프트>`: claude의 불리언 `-p`와 값 옵션 `-w`를 묶은 짧은 옵션 클러스터가 `-w`로 시작하는 토큰만 매칭하는 판독기에 보이지 않고 통과한다 (`lane_resume.go`의 `isShortClusterCarryingR` 선례).
- **(d) 공허한 기록 계수기** — `factory_anchor_guard_test.go`의 run-count 사전 점검이 레거시 backlog.json 존재 검사를 쓰며 새 픽스처에서 항상 0을 돌려, "거부된 기동은 run 기록을 남기지 않는다" 단정이 공허하다. 두 저장 형식을 모두 다루는 정준 판독기(`LoadPure()`)를 직접 호출해야 한다.

### §0.3 베이스라인 위치 (본 트리 실측)

- 본 트리는 main `1ced8cab9`에서 새로 분기되었다(`WT-launcher-argv-redesign`). **선행 가드 구현은 main에 없다** — `factory_anchor_guard_test.go`가 본 트리에 없고, 가드 기호(`guardLeaderLaunchAnchor`·`enforceLeaderAnchorGuard`)가 본 트리 `internal/cli`에 없다(측정: ls·grep, 2026-10-15).
- 선행 구현 전체는 보존 브랜치 `WT-lead-discipline` 팁 `7fe9143d9`(읽기 전용 증거)에만 있고, 본 트리 HEAD와의 병합 공통 조상은 `842160157f`다(측정: `git merge-base`).
- main이 지닌 공유 판별자 `cardScopeFromBranch`는 `internal/cli/codex_review_scope.go:170`에 살아 있다(측정: grep).
- main의 런처 흐름은 이미 판정면 재설계의 접착면을 지닌다: 리더 케이스가 `-w` 해석보다 먼저 실행되고(`internal/cli/cc.go:194` 스위치), 소비 루프와 최종 argv 조립이 한 함수 안에 있다(`runLaunchClaude`, `internal/cli/launcher.go:601` — 소비 루프 658–692, `buildArgs` 735–752, `normalizeWorktreeFlag` 917, `worktreeFlagValue` 1030). `--branch` 소비(t295)도 본 트리에 있다(`internal/cli/worktree_branch_flag.go:86` `splitWorktreeBranchFlag`, `:58` `resolveWorktreeExistingBranch`).

### §0.4 재설계 방향의 확정

운영자가 기록한 두 방향:

1. 앵커 판정을 **exec 경계의 최종 argv**로 옮긴다 — 모든 재작성·전처리 패스가 끝난 뒤 실행될 프로세스에 넘기는 그 argv 정확히를 판정한다.
2. 병렬 손작성 파서들을 **하나의 정준 argv 해석기**로 바꿔 런처 소비 모형·claude 옵션 모형·`--` 구분자 의미를 통일한다.

본 SPEC은 두 방향을 **방향 1에 고정한 종합**으로 채택한다. 근거: (i) 방향 1은 예측을 원천 삭제한다 — 소비가 일어난 뒤의 argv에는 더 이상 "런처가 무엇을 먹을지"의 모형이 필요 없다; (ii) 순수 방향 1만으로는 `--branch` 생성 주석이 보이지 않는다(그 플래그는 exec 전에 소비되므로) — 그러나 그 값은 소비 패스 자체가 이미 알고 있으므로 재해석 없이 주석으로 실려 온다; (iii) 방향 2의 단일 해석기는 종합 안에서 실린다 — 해석기가 뷰를 만들고, 가드와 exec가 **같은 뷰**를 읽는다. 가드는 더 이상 런처를 모방하지 않고 런처의 산출물을 읽는다. iter8이 지적한 세 경로 (a)(b)(c)는 이 구조에서 표현 자체가 사라진다: 소비 뒤 argv에서 `-b`는 이미 없고(→a 해소), worktree 옵션은 모두 최후 우선으로 스캔되고(→b 해소), `--` 뒤 토큰은 그대로 같은 argv 안에서 백엔드 모형으로 분해된다(→c 해소).

## §1 요구사항 (GEARS)

### §1.1 판정 입력면 — 단일 정준 해석기

- **REQ-LA-001 (단일 정준 해석기)** the launcher entry paths (`moai cc`, `moai glm`) shall 기동당 정확히 하나의 정준 런치 뷰(launch view)를 하나의 해석기(resolver)로 생산해야 한다 — 그 해석기는 ① 런처 소비 모형(런처가 스스로 소비해 제거하는 옵션: `-b`/`--bypass`, `-c`/`--continue`, `--model`/`-m`, `--permission-mode`, `--branch`, 진입 플래그), ② `--` 구분자 의미(첫 구분자는 런처가 소비하고 꼬리는 그대로 통과), ③ 백엔드 옵션 모형(worktree 옵션 철자들과 최후 우선)을 하나로 엮는다. 기동 경로상 그 밖의 어떤 구성요소도 원시 진입 인수에서 worktree 선택을 재유도하지 아니한다.
- **REQ-LA-002 (뷰의 계약)** the launch view shall 백엔드에 넘겨지는 최종 argv(뷰의 `finalArgv`), 그 argv에서 백엔드 옵션 모형으로 읽은 worktree 선택, `--branch` 생성 주석(짧은 이름 `-w`와 `--branch`가 함께 있고 대상 트리가 아직 없을 때 그 트리가 materialize될 브랜치 값), 재정의 플래그 존재 여부를 함께 실어야 한다. **When** 기동이 실행(exec)되는 때, the exec path shall 가드가 판정한 그 뷰를 소비해야 하며 같은 인수를 두 번 해석하지 아니한다 — 단일 해석 보장.
- **REQ-LA-003 (백엔드 옵션 모형 준수)** the worktree selection read from the final argv shall 백엔드(claude)의 옵션 모형을 따라야 한다 — worktree 옵션 여러 개일 때 최후 것이 이기고, 네 철자(`--worktree 값`, `--worktree=값`, `-w 값`, 붙여 쓴 `-w<값>`)와 값 옵션을 묶은 짧은 클러스터(예: `-pwcard-tree` — 불리언 `p` 뒤 값 옵션 `w`가 나머지를 값으로 가짐)는 모형이 규정하는 대로 분해되며, 값을 반드시 취하는 옵션(긴 철자, 마지막 글자가 값 옵션인 짧은 클러스터, 런처 소유 값 플래그) 바로 뒤에 선 `--`는 그 옵션의 값이고 옵션 파싱은 그 뒤에서 계속되며, 그 밖의 `--`는 옵션 파싱을 끝낸다(실측: claude 2.1.302 — `timeout 25 claude --append-system-prompt -- --version </dev/null` → `2.1.302 (Claude Code)`, exit 0). 선택적 값을 취하는 worktree 옵션(`-w`/`--worktree [이름]`) 바로 뒤의 `--` 처리는 바꾸지 않으며 본 요구사항에서 측정하지 않았다(미측정 잔여). 백엔드 모형의 선언은 구현 전 측정으로 고정한다(plan §C-1).

### §1.2 판정면 — 보존 의미

- **REQ-LA-004 (앵커 거부 — 변이 전)** **When** 지원되는 리더 기동(`moai cc -f`·`moai glm -f`)의 판정이 카드 워크트리 접두(`WT-`) 브랜치를 지닌 앵커 트리로 귀결되는 때, the launcher shall 어떤 세션 변이보다 먼저 기동을 거부해야 하며 감지된 트리와 브랜치, 바로잡음을 거부 사유에 명명해야 한다 — 환경 마커, run 기록, settings 준비, worktree materialize가 모두 판정 뒤에 대기한다(선행 SPEC REQ-FD-007의 배치 보존). 실효 브랜치의 정의: 존재하는 트리는 그 체크아웃 브랜치, 생성 예정 트리는 생성 주석의 브랜치 값, 기동 디렉터리는 그 체크아웃 브랜치다.
- **REQ-LA-005 (재정의 통로)** **Where** 문서화된 명시적 재정의 플래그(`--allow-card-anchor`)가 존재하는 곳에서, the launcher shall 경고와 함께 진행을 허용하고 백엔드·트리·브랜치를 적은 감사 행(`.moai/logs/leader-anchor-audit.log`)을 남겨야 한다. **When** 감사 행을 기록할 수 없는 때, the launcher shall 기동을 거부해야 한다(기록 실패에 대한 fail-closed — 기록되지 않는 우회는 없다). 플래그는 제 자리에서만 재정의로 인정된다 — 다른 옵션의 값으로 서 있는 토큰은 재정의가 아니며 백엔드에 그대로 남는다.
- **REQ-LA-006 (판별자 불변)** the anchor discriminator shall 커밋된 레인 규약 불변값(브랜치의 `WT-` 접두)만을 트리의 공유 카드 스코프 판정으로 읽어야 하며, 세션 환경 라벨을 판별 입력으로 삼지 아니한다(선행 REQ-FD-009 계승; 공유 판정의 현재 좌표는 §0.3).
- **REQ-LA-007 (레인·비팩토리 면제)** **When** 기동이 레인 합류(`-l`/lane-shape 이름)이거나 팩토리 리더 진입이 없는 때, the launcher shall 앵커 가드를 적용하지 아니한다 — 레인의 카드 트리 앵커는 의도된 배치다(선행 REQ-FD-010 보존).
- **REQ-LA-008 (fail-open 방향)** **When** 판정의 구성 요소가 검증 불가능한 때(생성 주석 없이 트리가 부재, 브랜치 판독 실패), the launcher shall 기동을 허용해야 한다 — 거부는 카드 접두의 긍정적 증거 위에서만 한다.
- **REQ-LA-009 (감사 기록 위생)** **While** 재정의 경로가 감사 로그를 여는 동안, the launcher shall 측정된 기록 위생을 보존해야 한다 — 존재하면서 정규 파일이 아닌 경로를 열기 전에 거부(Lstat), no-follow·비차단 플래그로 열기, 연 뒤 서술자의 파일 형식을 재판독해 정규 파일이 아니면 거부(보존 브랜치의 O11·O20·codex-690 계승 — symlink 추종, FIFO 블록, 개방 경합의 세 수리를 모두 유지).

### §1.3 회귀 기준면 — 기록된 실패 계열의 명명된 기준

- **REQ-LA-010 (소비 옵션 그늘 회귀 — 계열 a)** **When** 리더 기동이 런처 소비 뒤에야 이웃이 되는 worktree 옵션을 지니는 때(예: `moai cc -f -w -b -- card-tree` — `-b` 소비 뒤 claude는 `--worktree card-tree`를 받는다), the launcher shall 그 기동을 거부해야 한다. cc·glm 양 팔에서 검증한다.
- **REQ-LA-011 (다중 옵션 최후 우선 회귀 — 계열 b)** **When** 최종 백엔드 argv가 worktree 옵션 둘 이상을 지니는 때(예: `-w plain-tree --worktree --branch main card-tree` → `--branch` 제거 뒤 `--worktree plain-tree --worktree card-tree`), the launcher shall 백엔드가 선택하는 옵션 — 최후 것 — 을 판정해 그것이 카드 트리를 이름댈 때 거부해야 한다.
- **REQ-LA-012 (묶음 짧은 옵션 회귀 — 계열 c)** **When** 백엔드 argv가 worktree 옵션을 묶은 짧은 클러스터를 지니는 때(예: `moai cc -f -- -pwcard-tree <프롬프트>`), the launcher shall 클러스터를 백엔드 모형으로 분해해 묶인 값이 카드 트리를 이름댈 때 거부해야 한다.
- **REQ-LA-013 (기록 계수기 실효성 — 계열 d)** **While** 시험 도구가 거부된 기동 뒤의 기록 부재를 단정하는 동안, the instrument shall 두 저장 형식을 모두 다루는 정준 run-store 판독기로 읽어야 하며, 기록이 존재할 때 계수가 움직이는 긍정 대조(positive control)를 함께 지녀야 한다 — 부재 단정이 공허하지 않도록.

## §2 제약

- **보존 의미의 완전열거(약화 금지)**: 본 SPEC이 건드리는 판정면의 보존 의미는 §1.2에 열거했다. 위임 본문이 열거한 다섯 항목의 귀속: ① 주 체크아웃에서의 브랜치 상태 변이 거부와 ② git 형식의 워크트리 고립은 **형제 가드 면**(BranchGuard PreToolUse 훅, 워크트리 세션 가드)의 의미로 본 SPEC이 흡수하지도 약화하지도 않는다 — 본 SPEC의 대상은 런처 앵커 가드의 argv 판정면이며, 형제 면은 비간섭으로 보존된다(Out of Scope 참조). ③ fail-open(REQ-LA-008), ④ 면제 행위(REQ-LA-005 재정의·REQ-LA-007 레인), ⑤ 허용 시 감사 기록(REQ-LA-005·009)은 본 SPEC 면 안에서 그대로 요구된다.
- **비리더 기동의 행동 보존**: 단일 해석기로의 리팩터는 비팩토리 기동을 포함한 모든 기동 경로에 영향을 주는 구조 변경이다. 비리더 경로의 관측 행동은 문자 그대로 보존되어야 하며 characterize 시험이 이를 묶는다(acceptance AC-LA-014).
- **백엔드 모형의 측정 고정**: REQ-LA-003의 백엔드 옵션 모형(최후 우선, 클러스터 분해)은 보존 브랜치 감사 기록과 판독기 주석에 근거한 **감사된 모형**이다. 구현 전 plan §C-1의 재측정이 이를 백엔드 선언(도움말/선언 표면)으로 재확인하며, 측정 결과와 어긋나면 본 REQ를 plan-auditor 경유로 정정한다 — 모형을 추측으로 굳히지 않는다.
- **Template-First 해당 없음**: 본 SPEC은 템플릿·배포 설정 면을 바꾸지 않는다(`internal/template/` 편집 없음).
- **상수화**: 감사 로그 상대 경로 등 새 계수는 `internal/config/defaults.go`의 상수로(보존 브랜치 이름 규약 계승 — 하드코딩 금지).
- **언어**: 산출물은 한국어, 식별자·경로·명령은 영어 (저장소 규약).

## §3 성공 기준 (요약)

전체 수용 기준과 검증 명령은 `acceptance.md`를 따른다. 측정 가능한 종결 상태:

1. 기록된 네 실패 계열 (a)–(d)가 명명된 시험 셀로 묶이고, 최종 트리에서 GREEN이며, 각 셀의 RED가 교체 전 상태에서 관측되어 기록된다 (AC-LA-001..004, 015).
2. 가드의 판정 입력과 exec의 입력이 같은 해석의 산물이다 — 뷰의 `finalArgv`와 exec 경계에서 포착한 argv가 같은 철자표에서 동일하다 (AC-LA-005).
3. 보존 의미 여덟(변이 전 거부·레인 면제·재정의+감사·기록 위생·fail-open·환경 독립·`--branch` 재사용·비리더 보존)이 각자의 시험을 지닌다 (AC-LA-006..014).
4. 기동 경로에 worktree 선택의 재유도가 없다 — 소스 스캔 시험이 단일 해석 보장을 묶는다 (AC-LA-013).

## Out of Scope

### Out of Scope — 형제 git 명령 가드 면

- BranchGuard PreToolUse 훅(주 체크아웃 브랜치 상태 변이 거부)과 워크트리 세션 가드(`git -C <주 체크아웃>` 거부 등)의 판정 로직은 본 SPEC이 건드리지 않는다. 이 면들은 moai 런처 argv를 해석하지 않으므로 argv 판정면 재설계의 대상이 아니며, 비간섭으로 보존된다.

### Out of Scope — `internal/factory/gtd_operation.go:583` (표면 밖 선존)

- dispatch 복구 경로의 `os.Stat` 오류와 부재 혼동 문제는 카드 지정 표면 밖의 기존 결함으로, 본 SPEC에 편입하지 않는다(위임 본문 지정).

### Out of Scope — 선행 SPEC의 M1 면 (Stop 게이트 리더 스킵)

- SPEC-FACTORY-DISCIPLINE-001의 Stop 훅 병목·리더 스킵 면은 본 SPEC의 범위가 아니다. 그 구현은 보존 브랜치에 남아 있으며 본 SPEC이 다루는 유일한 면은 리더 앵커 가드의 argv 판정면이다.

### Out of Scope — 기동 경로 판독 대상이 아닌 인접 면의 argv 규약 (plan-repair 1·2 재획정)

- 기동 경로 위의 worktree 선택 판독자는 범위 **안**이다. 본 트리 실측(HEAD `de846d000`, 2026-10-15)에서 기동 경로의 판독자는 다섯이다 — `resolveWorktreeL2Path`(launcher.go:992, 호출 cc.go:282·glm.go:334), `resolveWorktreeExistingBranch`→`splitWorktreeBranchFlag`(worktree_branch_flag.go:57·86·121, 호출 cc.go:289·glm.go:339), `ccWorktreeWriterPrecheck`(session_worktree.go:1012, 호출 cc.go:296), `seedAdmittedWorktreeHooks`(codex_hooks_seed.go:62, 호출 cc.go:302·glm.go:347), `extractWorktreeLaunchTarget`(launch_chain.go:58, launcher.go:864 경유) — 이며 M1에서 전부 런치 뷰를 소비하도록 재배선된다(REQ-LA-001과 구판 배제 절의 모순 해소 — plan §A.4 설계 노트).
- 강제 이관 금지(범위 밖 보존)는 worktree 값을 읽지 않는 인접 면에 적용된다: 레인 재개 어셈블러(`lane_resume.go` `laneJoinChildArgv` — 비리더 레인 조인 조립, §2 비리더 보존·AC-LA-014가 행동을 묶는다)와 `moai worktree` 같은 세션 워크트리 관리 표면의 비기동 흐름. 어셈블러는 기동 경로 호출 클로저 **안**에 있다(본 트리 실측: `laneJoinChildArgv` 호출 cc.go:273 — 레인 케이스가 공용 꼬리로 낙하, plan-repair 2) — 그러나 측정된 사실은 그것이 worktree 값을 읽지 않는 순수 argv 조립자(세션 이름 쌍 삽입·settings 부착, 파일 안 `worktreeFlagValue` 0건)이라는 것이므로 뷰 편입 대상이 아니고, AC-LA-013의 스캔 집합 **안**에서 "조립은 허용·worktree 값 판독은 금지"로 검증된다.

### Out of Scope — Claude Code 런타임 내부

- 백엔드가 worktree 이름을 해석·생성하는 내부 절차는 수정 대상이 아니라 선언적 모형(REQ-LA-003)의 측정 대상이다.

### Out of Scope — 보존 브랜치의 병합·폐기 결정

- `WT-lead-discipline` @ `7fe9143d9` 트리를 병합할지 재구성할지의 착지 전략은 운영자·리더 결정 사항이며 본 SPEC이 명하지 않는다. 본 SPEC의 구현은 본 트리(main 기반)에서 이루어진다.
