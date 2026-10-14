# SPEC-CAPACITY-RECLAIM-001 적대적 계획 감사

## Claim

판정: FAIL. 대상은 커밋 `40cc1031c2b18781377f1bdf2c590cd39f08244e`의 계획 문서 다섯 개다. 구현 결함을 확정하는 감사가 아니라, 명시된 구현 계획과 인수조건의 결함을 판정한다.

## 필수 통과 조건

- MP-1 PASS: 요구사항 정의 13개가 REQ-CAP-001부터 013까지 중복과 누락 없이 이어진다.
- MP-2 PASS: spec.md 요구사항 계층을 검사했다. shall/shall not, When, Where와 복합 조건을 사용한다. legacy negative 형태는 현재 호환 기간의 허용 대상으로 판단했다. Given-When-Then인 AC 계층은 GEARS 검사 대상에서 제외했다.
- MP-3 PASS: PyYAML로 파싱한 필수 필드 12개가 모두 존재한다. id/title/version/status/author/priority/phase/module/lifecycle/tags는 문자열, created/updated는 날짜다. spec.md:2의 ID, :4의 quoted semver, :5의 draft, :9의 P0, :12의 lifecycle이 유효하다.
- MP-4 PASS: 템플릿 변경은 ignore 규칙과 hygiene 기본값이다. 특정 프로그래밍 언어의 도구를 우선하는 요구사항은 없다.
- MP-5 PASS: D7 스크립트 exit 0, stdout 없음. 교차 참조한 HYGIENE, SWEEP, REAPER의 spec.md가 존재하며 status는 모두 completed다. usage-log 소유권 변경은 Q4 승인 조건으로 명시되어 있다.
- MP-6 PASS: D8 스크립트 exit 0, stdout 없음. SPEC에 syscall 언급이 없다.
- MP-7 PASS: plan.md의 [NEEDS CLARIFICATION] 표식은 0개다. research.md는 Tier M 문서 집합에 없으며 존재하지 않는다. 아래의 실제 미결정 사항이 표식 부재 때문에 해소되는 것은 아니다.
- MP-8 N/A: acceptance.md:7이 release-blocking AC를 하나도 분류하지 않았다고 명시한다. §2.1의 네 요소 의무를 소급해서 적용하지 않았다. 일반적인 두 셀 채택 및 mutant probe 의무는 별도로 검토했다.
- MP-9 PASS: CN-4에서 6개 milestone, exit binding 0개, ordering candidate 13개를 관측했다. candidate를 수동으로 대조했다. M2 직전/직후 보존 측정과 M3의 AC-012/013 green 이후 M4 bulk 실행(acceptance.md:150)은 계획 순서와 양립한다. 스크립트의 CONFLICT 부재만으로 판정하지 않았다. 아래 승인 거절 경로의 모순은 milestone 순서 충돌이 아닌 완료 조건 결함이다.

## 루브릭 점수

기준: .claude/agents/moai/plan-auditor.md M3. Tier M 통과 문턱은 spec-workflow.md:119의 0.80이다.

- 명확성 0.50: 현재 release 포함 여부, cap 설정 키, apply 확인 표면을 plan.md:37과 spec.md:118에서 미결정 상태로 남긴다.
- 완전성 0.75: History/문제/요구사항/계획/AC/제외 범위/frontmatter는 있지만, home cleanup 인수조건의 실행 명령 및 위 결정 사항이 불완전하다.
- 테스트 가능성 0.50: AC-010/011은 실행 명령과 설정 키가 정해지지 않았다. AC-004/007/014는 요구사항을 어기는 결과를 허용한다.
- 추적성 1.00: 기계적으로 요구사항 13개를 수집했고 UNCOVERED/ORPHAN 출력은 없다. AC 정의 15개가 유효한 요구사항을 참조한다. 이 점수는 ID 연결의 완전성이지, 검사의 충분성을 뜻하지 않는다.
- 네 차원 단순 평균 0.6875. 필수 조건의 실패가 없어도 차단 결함 및 Tier M 문턱 미달로 PASS를 부여하지 않는다.

## 차단 결함

1. P1, plan.md:11: registry 추가만으로 usage-log 소유권이 이전된다고 가정한다. 실제 CLI는 LogDir를 `.moai/logs`로 설정(clean.go:170)하고 rotator는 `filepath.Join(r.LogDir, name)`으로 순회(rotate.go:163)한다. 따라서 이름만 추가하면 `.moai/logs/usage-log.jsonl`을 검사한다. 실제 파일은 `.moai/harness/usage-log.jsonl`이다. 기존 retention-prune 호출도 hook.go:1331에 남는다. 대상 루트와 기존 prune의 종료 또는 직렬화까지 소유권 전환 계획에 명시해야 한다. 신뢰도 높음.
2. P1, acceptance.md:130: AC-013은 report 모드만 실행하며 live-session fixture가 없다. AC-015는 predicate를 통과하는 두 후보만 입력한다. 따라서 report에서는 PRESERVE하지만 --yes에서는 unpushed/live 트리를 삭제하는 구현이 이 조건을 통과할 수 있다. 실제 삭제 경로에서 owner-confirmed unpushed/live fixture의 보존을 검증해야 한다. 신뢰도 높음.
3. P2, acceptance.md:85: 초과 크기 primary를 그대로 두는 no-op도 primary 1개, rotation 0개, in-place truncate 없음 조건을 만족한다. 요구사항이 요구하는 회전을 관측하지 못한다. 회전 전후 파일의 내용/식별자와 실제 rotated 결과를 검증해야 한다. verification-completeness.md §2 mutant probe에 따라 현재 조건은 채택하기에 너무 얕다. 신뢰도 높음.
4. P2, acceptance.md:153: 모든 AC의 통과를 요구하지만 AC-007과 AC-015의 Given은 Q4/Q2 승인이다. Q4의 공개 Default인 기존 owner 유지나 Q2 거절은 정상 결정인데, 해당 AC를 충족할 수 없어 완료할 수 없다. 승인 여부별 적용 AC 및 N/A 처리를 정의해야 한다. Q1/Q2/Q4의 비어 있는 verdict 자체를 계획 결함으로 판단하지 않았다. 다만 manager-spec.md:119에 따라 자율 kickoff는 차단된다. 신뢰도 높음.
5. P2, plan.md:37: release keep-3의 현재 release 포함 여부, cap 설정 키, apply 확인 표면을 run pre-flight로 미룬다. AC-010/011의 명령도 `the apply runs`뿐이다. 현재 구현은 current를 보호하고 non-current 3개를 유지(clean_home.go:432–445)하므로 총 3개 요구사항과 결과가 다르다. 최종 개수/보호 불변식, 설정 키와 값 검증, 확인 경로, fixture 실행 명령을 계획 단계에서 확정해야 한다. 신뢰도 높음.
6. P2, acceptance.md:62: 세 패턴을 포함하는 주석만으로 grep가 성공한다. 실제 ignore를 검증하지 않는다. 아래 fixture에서 grep exit 0과 git check-ignore exit 1을 함께 관측했다. 대표 경로와 corrupt 변형을 git check-ignore --no-index로 검사해야 한다. 신뢰도 높음.
7. P2, acceptance.md:138: JSON 배열 길이만 같으면 path/branch/size가 전혀 없는 객체 배열도 통과한다. REQ-CAP-013이 요구하는 필드 및 table 출력은 검증되지 않는다. git의 경로/브랜치 집합, 크기, table 출력까지 비교해야 한다. 신뢰도 높음.
8. P2, acceptance.md:74: REQ-CAP-003도 매핑하지만 보존 대상이 두 state 파일뿐이다. 세 corrupt archive를 git rm으로 로컬에서도 지워도 AC-005의 tracked count는 0이고 AC-006은 유지된다. archive 세 개의 존재와 내용 hash도 M2 전후로 비교해야 한다. 신뢰도 높음.
9. P2, acceptance.md:44: awk의 `$2` 및 xargs가 공백을 포함하는 worktree 경로를 분해한다. 공백 경로 fixture는 실제 144 KiB인데 이 명령은 du 오류, 빈 stdout, 최종 exit 0을 반환했다. 경로 전체를 보존해서 읽고 모든 측정 실패를 명시적으로 실패 처리해야 한다. 신뢰도 높음.

## Evidence

명령은 이 워크트리에서 실행했다. 외부 clone이나 실제 worktree의 삭제는 실행하지 않았다.

```text
git rev-parse HEAD
40cc1031c2b18781377f1bdf2c590cd39f08244e

git status --short
(stdout 없음)
```

Python subprocess로 git show 40cc1031c:<path>의 bytes와 실제 다섯 문서의 bytes를 비교했다.

```text
spec.md matches reviewed commit: True
plan.md matches reviewed commit: True
acceptance.md matches reviewed commit: True
decision-index.md matches reviewed commit: True
progress.md matches reviewed commit: True
```

plan-auditor.md의 trace/CN-4/D7/D8 bash 블록을 그대로 추출하고 문서 경로 placeholder만 치환하여 실행했다.

```text
trace exit: 0
COLLECTED: 13 REQ definitions (acceptance input: read)
ordering exit: 0
COLLECTED: 6 milestones in plan order (M1 M2 M3 M4 M5 M6), 0 exit bindings, 13 ordering candidates
D7 exit: 0
(no stdout)
D8 exit: 0
(no stdout)
```

RED 관측 재실행(일반 AC이며 MP-8 release-blocking 재실행으로 주장하지 않음):

```text
git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1
exit 0, stdout: ' 128526661 total\n'
grep -n -E 'usage-log|lessons-inbox|\.corrupt' internal/template/templates/.gitignore
exit 1, stdout: ''
git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*' | wc -l
exit 0, stdout: '       5\n'
grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go
exit 1, stdout: ''
grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'
exit 1, stdout: ''
```

`/tmp`의 TemporaryDirectory 안에 격리된 fixture를 만들었다. 다음은 이 fixture와 안전한 식 평가에서 얻은 출력이며, 제품 구현을 실행한 결과는 아니다.

```text
AC-CAP-004 comment-only mutant: exit 0 stdout: 1:# usage-log.jsonl
2:# lessons-inbox.jsonl
3:# .corrupt
Actual ignore behavior: exit 1 stdout: ''
AC-CAP-014 count-only mutant: git count= 270 array count= 270 all entries have path/branch/size= False
AC-CAP-007 no-op mutant: primary=1 rotations=0 no_truncate=True; Then accepts: True
```

보조 linter는 `git merge-base --is-ancestor 4f8aba061 WT-web-haiku55-effort-agent-dedupe`의 exit 0으로 빌드 원본 브랜치에 대한 귀속을 확인한 뒤 사용했다.

```text
moai spec lint .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md --format json
0 error(s), 13 warning(s)
```

warning은 모든 REQ에 대한 CoverageIncomplete였다. 실제 sibling acceptance 참조를 수집한 trace script의 결과와 일치하지 않으므로 제안의 추적성 결함으로 채택하지 않았다. format 지정에도 불구하고 이 실행은 텍스트 표를 반환했다.

공백 경로 fixture에 AC-002의 명령을 그대로 실행한 결과:

```text
exit: 0 stdout: '\n' stderr: 'du: /private/tmp/capacity: No such file or directory\n'
Actual fixture size: 144\t/tmp/capacity review-zeq6_oy4
```

## Baseline-attribution

측정 tree는 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631`, HEAD는 `40cc1031c2b18781377f1bdf2c590cd39f08244e`다. 원래 RED pin은 `2aab5f797`이다. 지정한 현재 커밋은 계획 문서만 변경했고, 이번 추시에서 위 RED 수치와 source grep 결과를 관측했다.

## Gaps

- hygiene apply, worktree --yes, home cleanup을 제품에서 실행하지 않았다. 향후 구현에 대한 runtime verdict를 내리지 않았다.
- 전체 worktree의 무거운 용량 traversal과 primary reports의 재측정은 하지 않았다.
- Q1–Q4의 운영자 결정 및 clone별 backup 완료는 관측하지 않았다.
- AC-003/007/010/011/015의 fixture RED와 regression guard mutant의 제품 실행은 미관측이다.
- CLI linter는 다른 브랜치에서 빌드한 설치본이다. 이 커밋의 소스로 새로 build하지 않았다.

## Residual-risk

이 리뷰의 PASS/FAIL은 계획 문서에 대한 판정이다. 지적한 내용을 수정한 뒤에도 격리 fixture로 위험한 실행 경로, owner 이전 시의 배타적 접근, 승인 거절 시 완료 분기, 설정값 경계를 검증하고 재감사해야 한다.
