# Sync-audit 판정 — 카드 t622, SPEC-GIT-DELIVERY-PROCEDURE-001 (Tier M, 0.2.6)

감사자: sync-auditor (독립 재실행, 읽기 전용). 재현 산출물: `.moai/reports/t622/sync-audit/`.
적용 규칙: `verification-claim-integrity.md` §1.1(표면 3·4), §2 — 모든 수치는 이 실행에서 이 트리에 대해 잰 값이다.

## 판정

**Overall Verdict: FAIL** — 점수 84/100 (조화평균; 가중평균 83.5)

판정 대상 AC 16개는 현재 HEAD 에서 모두 재현 PASS 다. FAIL 은 AC 가 보지 않는 **차단 결함 1건(F1)** 때문이다. sync 단계에서 고친 `doc-execution.md:34` 가 "뒤 단계는 이 값을 읽지 않는다" 는 사실이 아닌 문장을 새로 넣었고, 이 문장은 같은 워크플로의 `delivery.md` Step 3.2·Phase 14 와 모순된다. 한 줄 수정이고, 재감사는 아래 F1 델타로 좁힌다.

## 기준 트리 귀속

| 항목 | 값 | 명령 |
|---|---|---|
| 워크트리 / 브랜치 | `.claude/worktrees/t622` / `WT-git-procedure-fixes` | `git branch --show-current` |
| HEAD (감사 시작·끝) | `50cebf3c2` · `50cebf3c2` (감사 중 이동 없음) | `git rev-parse --short HEAD` |
| CARD_BASE | `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf` (1줄) | `git merge-base --all develop HEAD` |
| 로컬 develop | `eb50af5a8ec51862e3df76c2e3b08377ce01c4d8`, CARD_BASE 이후 57커밋(흡수 안 됨) | `git rev-parse develop`, `git rev-list --count f1f034bb4..develop` |
| 범위 대조 | `card-range-names.txt` 1118줄; 증거 밖 파일 22개(범위 8쌍 16 + `.toml` + SPEC 4 + CHANGELOG) | `git diff --name-only develop...HEAD` |
| 스냅숏 신선도 | `snapshot-stale.txt` `test -e` 0 / `test -s` 1, `snapshot-stale-mirror.txt` `test -s` 1 | acceptance.md 관례 명령 그대로 |
| 판정 도구 | 이 트리에서 `go test` 가 컴파일한 테스트 바이너리와 Makefile 타깃. 설치된 `moai` 바이너리는 쓰지 않았다 | — |

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 75/100 | PASS (must-pass: AC 16/16 PASS) — 단 차단 결함 F1 | AC 표 아래. F1: `doc-execution.md:34` 가 `delivery.md:284-287`·`:447-450` 과 모순 |
| Security (25%) | 95/100 | PASS | 추가 줄 비밀 탐침 `grep -c -i -E 'token\|secret\|password\|...'` → `0`. 변경은 의도치 않은 자동 병합을 막는 방향. Critical/High 없음 |
| Craft (20%) | 85/100 | PASS | Go 코드 변경 0(커버리지 해당 없음). `make agents-emit-check` exit 0, `make commands-emit-check` exit 0, 표적 테스트 exit 0. 감점: 스크립트 편집(F5)·미기록 사고(F6) |
| Consistency (15%) | 85/100 | PASS | 로컬·템플릿·`.toml` 바뀐 줄 동일(`diff` exit 0 ×3), BASE 사본 차이 본문 보존 8/8. 감점: F3·F4 |

## AC 재실행 (acceptance.md 명령 그대로, 현재 HEAD)

| AC | 재현 결과 (verbatim) | run-report 값과 비교 |
|---|---|---|
| 001 | `local paras=1 af_e=0 af_s=1 lg_s=1` / `template …` 같음 / `toml …` 같음; L·T 절 `diff` exit 0, T·toml 절 `diff` exit 0 | 일치 |
| 002 | 두 블록 모두 `fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS`; 절 52줄 ×4, `local-b=0 tmpl-b=0`, 표 8행 `mL=0 mT=0` | 일치 |
| 003 | Pre-Edit 절 35줄 ×4, `L=0 T=0` | 일치 |
| 004 | squash `0`·`0`, resolved `2`·`2`, 출처 줄 로컬 377 / 템플릿 352 (`src_exit=0`) | 일치 |
| 005 | 나머지 여덟 파일 로컬·템플릿 모두 `0`, `.toml` `1`(26행 기본값 설명); `manager-git.md` 34 / 템플릿 32 기본값 문장뿐 | 일치 |
| 006 | `dl_e=0 dl_s=1 de_grep=1 de_ext_s=1`, `manager-[g]it.md` dl 2 / de 1 (두 사본), `--auto-merge` 4·4·4 | 일치 (자동 검출). F1 은 이 검출식 밖 |
| 013 (a) | `acp=0 pub=0` | 일치 |
| 013 (b) | 8쌍 모두 `body_diff=0 base_nonempty_s=0` (BASE 반출본은 이 감사가 `git show` 로 새로 만듦) | 일치 |
| 013 (c) | `argument-hint: "[SPEC-XXX] [--auto-merge] [--skip-mx]"`, `hint=0` | 일치 |
| 013 (d) | exit 0, 최상위 PASS `2`, acp 흔적 `6`, 빈 선택 토큰 `0` | 일치 |
| 013 (e) | go test exit 1(기준선과 같은 이유), PASS `17`, FAIL = {`TestRuleTemplateMirrorDrift`, `…/spec-workflow.md`}, 정렬 집합 기준선 대비 `setdiff=0`, `nf_e=0 nf_s=1 lp_e=0 lp_s=1` | 일치 |
| 014 (check-only) | `make agents-emit-check` exit 0 (`ok … agentemit 0.459s`, `[no tests to run]` 없음); 범위 변경 `.codex/agents/moai/` = `manager-git.toml` 1줄; 예시 `1`; 절 diff `sync=0 pram=0`(13줄); RED→재생성 순서는 커밋 `cc51d8479`(증거만) → `5708e04d2`(`.toml`) 로 확인 | 일치 |
| 015 | `s=0`, 추가 줄 `39`, SPEC/REQ/날짜/`CLAUDE.local` exit `1`×4, 16진 낱말 `0`, 내부 토큰 탐침(`t[0-9]{3}\|card\|lane\|/Users/\|.moai/reports`) exit 1. 추가로 `TestTemplateNeutralityAudit` exit 0, strict `TestTemplateNoInternalContentLeak` exit 0 (CI 가 도는 두 검사를 로컬에서 표적 실행) | run 38 → sync 39 (sync 편집 1줄), 일치 |
| 016 | `e=0 s=1` (카드 비병합 커밋 38개), 대조 2줄 `97ef8e302…`·`6896eef37…` | 일치 |
| 025 | 범위 diff 16파일 `s=0`, `any=1 lines=1`, `[ZONE:` 적중 exit 1; clause 네 개 사본마다 1; 레지스트리 13·13 / 0·0 | 일치 |
| 026 | (i) 조각마다 ≥1(qgc 1+2), (ii) `ii_e=0 ii_s=1` ×2, (iii) `1 1 1 2` ×2; `--merge` 줄 5개 모두 "deprecated alias of `--auto-merge`" — 읽기로 확인 | 일치 |
| 027 | dl `--no-merge` 1줄(29행), `i_e=0 i_s=1 ii_e=0 ii_s=1` ×2 | 일치 |
| 028 (b) | (a) mg 1 / dl 1 ×2, `028b_e=0 028b_s=1` ×2, 모드 줄 14 ×2 — `ac028-reading.md` 14행과 대응 | 일치 |
| 029 (b) | (a) 1·1 ×2, `029b_e=0 029b_s=1` ×2 | 일치 |
| 030 | `make commands-emit-check` exit 0; 경우 (A): `ch_e=0 ch_s=1`, 원본 커밋 `2f4dfd803…`, `art_s=1`, 경로 대조 `2`, 추적 `2`, `or_e=0 or_s=1`, `pubLT=0`, `flags_s=1` | 일치 |

## Findings (structured defect-list)

- **F1 [Medium] [blocking] [confidence: High]** `.claude/skills/moai/workflows/sync/doc-execution.md:34` (템플릿 같은 줄) — sync 단계 커밋 `0dc007201` 이 "Store result as `is_worktree_context` boolean as informational context only; no later phase reads it" 로 바꿨다. 같은 워크플로의 뒤 단계가 워크트리 문맥을 쓴다: Phase 13 Step 3.2 github-flow 의 **Worktree context** 전달 경로(`delivery.md:284-287`, 템플릿 259-262)와 Phase 14 다음 단계 선택지의 **If worktree context:** 분기(`delivery.md:447-450`, 템플릿 422-425). 원래 문구 "for use in Phase 13" 은 Step 3.4 트리거 부분만 낡았고, Phase 13 의 Step 3.2 경로에 대해서는 맞는 말이었다. "낡았다" 는 전제를 확인하지 않은 채 사실이 아닌 문장이 배포 템플릿에 들어갔다(VCI §1.1 표면 4).
  - 증거: `/usr/bin/grep -rn 'is_worktree_context\|is_wt_context\|worktree context' .claude/skills …` →
    `delivery.md:287:- Display PR URL and worktree context` · `delivery.md:361:… worktree context alone never triggers a merge.` · `delivery.md:447:**If worktree context:**`; `sed -n '284,287p'` → `**Worktree context** (detected from git directory structure):` / `- Push worktree branch to remote` / `- Create PR if not exists (same as feature branch flow)`.
  - AC 가 못 잡은 이유: AC-GDP-006 검출식은 "워크트리 문맥 = 기본 병합" 모양만 본다. 이 문장은 병합과 무관한 사실 주장이다.
  - Required fix: 두 사본 34행을 같은 문장으로 고친다. 예: "Store result as `is_worktree_context` boolean; later phases may use it for delivery routing (Phase 13 Step 3.2) and next-step options (Phase 14), but it never decides auto-merge." 뒤 36행은 그대로 둔다. 재감사 델타: 34행 L·T 두 사본, AC-GDP-006(de 절)·013(b) doc-execution 쌍·015 재실행.

- **F2 [Low] [optional] [confidence: Medium]** `manager-git.md:172-177`(템플릿 170-175), `delivery.md:371-375`·`:385-391`(템플릿 346-350·360-366) — team 모드의 "전원 승인" 조건은 앞 문장에만 있고 실행 단계 목록에는 승인 확인 단계(예: `gh pr view --json reviewDecision`)가 없다. `delivery.md` 는 실패 절(`:397` "If approvals missing (Team mode) … do NOT merge")로 보완하지만 `manager-git.md` 에는 그 분기가 없다. 카드 이전 모양과 같아 회귀는 아니다. 증거: 위 `cat -n manager-git.md` 166-177행. Required fix(선택): 단계 3 뒤에 "team mode: confirm all required approvals before step 4" 한 줄.

- **F3 [Low] [optional] [confidence: High]** `manager-git.md:139`(템플릿 137) "PR squash + delete-branch", `:123` "squashed remote", `:129` "un-squashed history … squashed remote" — `merge_method` 를 설정할 수 있게 된 뒤에도 산문이 squash 를 전제한다. 실행 예시가 아니라 AC-GDP-005 대상은 아니고, late-branch 절은 t658 소관(spec §D)이다. 증거: `/usr/bin/grep -n -E 'gh pr merge[^|]*--squash|--squash' .claude/agents/moai/manager-git.md` 는 34행만 찍지만, 해당 산문 줄은 `cat -n` 123·129·139행에서 읽힌다. Required fix(선택): t658 입력으로 넘긴다.

- **F4 [Low] [optional] [confidence: High]** `.claude/skills/moai/workflows/sync.md:114`(템플릿 104) — 배포 템플릿의 영어 스킬 본문에 한국어 괄호 "(auto-merge 옵트인)" 을 새로 넣었다. 같은 줄의 기존 괄호("PR 생성", "MX 검증 스킵")와 모양은 맞지만 템플릿 스킬 본문 영어 원칙과 어긋난다. 중립성 검사(프로그래밍 언어 축)는 통과한다. Required fix(선택): "(opt-in)" 같은 영어 표기.

- **F5 [Low] [optional, 절차] [confidence: High]** `.moai/reports/t622/run/m1-edit.py`, `m2-edit.py`, `final-judges.sh`, `final-reading-targets.sh` — part 1 은 판정뿐 아니라 **M1·M2 지침 편집 자체도 파이썬 스크립트로** 했다(`m1-edit.py` 는 `delivery.md`·`doc-execution.md`·`SKILL.md`·`reference.md`·`quality-gates-context.md`·`workflows/sync.md` 두 사본, `m2-edit.py` 는 `delivery.md` 두 사본을 덮어씀). 가드는 스크립트 안을 보지 못한다. 결과는 독립 확인했다: 바뀐 줄 L·T 동일(`L-T=0` 스킬 7파일, `manager-git` 3사본), BASE 사본 차이 본문 보존 8/8, 세션 시작 시 primary 체크아웃 상태에 지침 파일 수정 없음. 판정 쪽도 part 2 가 스크립트 없이 `p2-*` 로 다시 쟀고, 이 감사가 16개 AC 를 다시 재 같은 값을 얻었으므로 **스크립트 출력에만 기댄 판정은 없다**. Required fix(선택): 다음 카드에서 지침 편집은 Edit 도구로 한다. 리드 보고의 "읽기 전용 판정만 스크립트로" 는 편집 스크립트 두 개를 빠뜨린 서술이다.

- **F6 [Low] [optional, 기록] [confidence: High]** part 2 의 zsh 따옴표 없는 변수로 인한 공허한 초록과 재실행이 증거 어디에도 기록돼 있지 않다. 증거: `/usr/bin/grep -rl -i -E 'zsh|unquoted|vacuous|word-split|wordsplit|SH_WORD_SPLIT' .moai/reports/t622/run .moai/reports/t622/sync .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001` → 출력 없음, `exit=1`. 인용된 `p2-*` 파일은 기대대로 빈 파일은 빈 판정 파일뿐이고(24개, 모두 기대 빈 결과), 비어 있으면 안 되는 조각·개수 파일은 채워져 있다. 이 감사의 재실행 값과도 같다. Required fix(선택): progress.md §E.2 part 2 에 사고 한 줄(무엇이 공허했고 어느 파일을 다시 만들었는지)을 남긴다.

- **F7 [Info] [optional] [confidence: High]** 추적되지 않은 `.moai/reports/t622/run-stale-b412/`(46파일)가 워크트리에 남아 있다. 판정 근거로 인용되지는 않는다(`preflight-summary.md:17`·`pre1-status.txt:1` 에서 미추적 잔재로만 등장). Required fix(선택): 워크트리 폐기 전에 지우거나 반출 여부를 기록한다.

## 사본·생성물·CHANGELOG 읽기 결과 (항목 2·3)

- 로컬·템플릿 바뀐 줄: `local-skills.body` 42줄 = `tmpl-skills.body` 42줄(`diff` exit 0, 7파일씩), `manager-git` 로컬·템플릿·`.toml` 14줄씩 `diff` exit 0 ×2. develop 전용 로컬 덩어리는 AC-013 (b) 본문 비교 8/8 로 보존 확인.
- `.toml` 은 템플릿 md 의 기계 방출과 일치(`make agents-emit-check` exit 0, 두 절 `diff` exit 0).
- 워크트리 문맥 기본 병합 문구: `.claude`·템플릿·`.codex`·`.agents` 전체에서 `auto-merge … default|default to auto-merge|worktree contexts default|merges automatically` 적중 없음. `--no-merge` 는 `delivery.md:383`(템플릿 358) 한 줄뿐이며 "Deprecated no-op … not merging is already the default" — 건너뛰기 스위치로 읽히지 않는다.
- `--squash` 고정 실행 예시: 범위 파일 전체에서 `manager-git.md:34`(템플릿 32)·`.toml:26` 의 기본값 설명 문장만 남음. `delivery.md` 의 squash 언급은 377행 기본값 설명뿐.
- `git_strategy.<mode>.merge_method` 는 실제 설정 키다: `internal/config/types.go:142` `MergeMethod string \`yaml:"merge_method"\``, `internal/config/validation.go:295-297` manual·personal·team.
- Frozen: 범위 diff 에 `[ZONE:` 줄 없음(exit 1).
- CHANGELOG: `[Unreleased]`(8행) 아래 `### Changed`(298행) 첫 항목(300행), SPEC id 1회, "16 judged acceptance criteria" — acceptance.md 판정 대상 수와 일치. 나열한 변경 파일 목록이 범위 diff 와 일치하고 "no Go code changed" 는 범위 목록에 `.go` 0개로 확인. Cf 문자: CHANGELOG 항목 0, 템플릿 추가 줄 0(대조 1).
- 모순 점검: 파일 사이의 유일한 모순이 F1 이다. `delivery.md` 와 `manager-git.md` 의 모드 조건 문장은 바이트 동일, 명령 `argument-hint`·사용법 줄·플래그 목록·CHANGELOG 는 서로 같은 말을 한다.

## Gaps (관측하지 않은 것)

- CI(`origin/develop` 전체 스위트, darwin/windows 매트릭스, CodeRabbit): 브랜치 미푸시, 레인은 push 하지 않는다.
- 교차 모델 감사(codex/glm): 실행하지 않았다. 커밋되지 않은 변경이 없고(`uncommittedChanges` 빈 집합), `baseBranch` 는 서버 쪽에서 원격 기본 헤드(main)로 풀려 develop 커밋 수십 개를 담은 다른 변경 집합을 리뷰하게 된다. 이 카드 범위(`develop...HEAD`)를 겨누는 대상이 도구에 없다.
- `go test ./...`, `make build`, `make embed-check` 미실행(검증 부하 규칙·지시). 설치 바이너리의 임베드 `.toml` 에 대해 아무것도 주장하지 않는다.
- 로컬 develop 이 CARD_BASE 이후 57커밋 앞서 있다. 범위 판정(014·015·016·025·030)은 병합 전 판정이며 통합 창 흡수 뒤 병합 트리에서 다시 재야 한다.
- docs-site 네 로케일은 옛 플래그·옛 기본값을 설명한다(spec §D X4, 후속 카드 소관).

## Residual-risk

- AC-GDP-026~029 는 줄 단위 검출이다. 읽기 기록과 이 감사의 읽기로 보완했지만, 검출식이 예상하지 않은 문장이 다른 절에 들어오면 보이지 않는다(F1 이 그 예).
- AC-GDP-002 뮤턴트 (ix)(절 밖 순서 뒤집기)는 설계상 보지 않는다. 이 카드 커밋은 `agent-common-protocol.md` 를 건드리지 않았다(AC-016).
- `make` 는 실패 레시피에서 exit 2 를 낸다 — acceptance.md 의 "exit 1" 표기와의 차이는 문구 문제로, 검출 자체는 run 의 RED 기록(`ac014-red.txt`)이 보여 준다.

## Recommendations

1. F1 을 고친다(두 사본 34행 한 문장). 재감사는 위 델타만 본다.
2. F2~F7 은 선택 사항이다. F3 은 t658 입력에 붙이고, F6 은 progress.md 에 한 줄 남기는 것을 권한다.
