<!-- moai-template:rc1-root@0.2.0; 상세 설계 2026-11-08; 상관관계용 메타데이터일 뿐 권한이 아니다 -->
# [RC1] 2026.11.0-rc.1 전체 범위·gate·영역 상위 Issue

## 목표와 사용자 가치
릴리스 `2026.11.0-rc.1`의 범위·릴리스 gate·미결(HOLD) 목록·영역 상위 Issue 9개(#1976~#1984)를 한곳에서 추적하는 지점이다. 작업 큐가 아니며 owner를 갖지 않는다. 가치는 (1) 설계 정본 §4/§5·00:1286 릴리스 gate가 GitHub 객체(milestone #1, sub-issue 9, blocked_by 49)로 실제 투영됐는지 재조회로 확인할 수 있고, (2) 필수 실패 gate(사용자 파일 손실·무승인 mutation·wrong-worktree·secret 유출·근거 없는 PASS·stale admission·중복 외부 효과) 0건을 영역별 증거 링크로 모아 release review 판정의 입력이 되며, (3) 사용자가 아직 정하지 않은 P04~P15 — ZCode 헤드리스 범위 결정됨(2026-11-08), 그 외 host 지원표 HOLD 조건을 최종 확인하는 체크포인트다.

## 비목표와 범위
비목표: leaf 32개의 구현·검증 내용을 이 Issue에 복제하지 않는다. P04~P15 — ZCode 헤드리스 범위 결정됨(2026-11-08), 그 외 host 지원표 HOLD 목록·영역 링크·verification matrix 갱신(기존 본문 해시 확인 뒤 정확 diff), 발행 manifest `.moai-cli/manifests/github-issues.json` receipt 기록.

## 근거와 확정 결정
| 주장 | 출처 | 확인일 · 버전 | 상태 |
|---|---|---|---|
| 로컬 `v2026.11.0`·`origin/v2026.11.0`·`2026.11.0-rc.1`·`origin/2026.11.0-rc.1` 모두 f6808390bd83c53e44964ed1d4645727945ea1cc. worktree 1개(기본 checkout, 브랜치 v2026.11.0), 열린 PR 0개. | local: git ls-remote origin / git worktree list / gh pr list -R modu-ai/moai-adk --state open | 2026-11-08 · git HEAD f6808390 | 검증 |
| `v2026.11.0` 브랜치에 branch protection 없음(GET …/branches/v2026.11.0/protection → 404 "Branch not protected"). 저장소 ruleset은 `Release tag immutability (v*)` 1개(target tag, active)뿐이라 브랜치 push·force·삭제를 막는 규칙은 현재 없다. | local: gh api repos/modu-ai/moai-adk/branches/v2026.11.0/protection ; gh api repos/modu-ai/moai-adk/rulesets | 2026-11-08 · gh 2.92.0 | 검증 |
| milestone #1 `2026.11.0-rc.1` open, open_issues 43 / closed 0. #1975 sub-issue 9개(#1976~#1984) 연결됨. #1976 하위에 #1985·#1986·#1974·#1987. | local: gh api repos/modu-ai/moai-adk/milestones ; gh api repos/modu-ai/moai-adk/issues/1975/sub_issues ; …/1976/sub_issues | 2026-11-08  | 검증 |
| GitHub sub-issue는 부모당 최대 100개, 중첩 최대 8단계. REST add는 `sub_issue_id`(숫자 id, number 아님) 필요, "The sub-issue must belong to the same repository owner as the parent issue." label·milestone 자동 상속은 문서에 없음. | https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/adding-sub-issues · https://docs.github.com/en/rest/issues/sub-issues | 2026-11-08 · REST X-GitHub-Api-Version 2026-03-10 | 검증 |
| Issue dependencies GA 2025-08-21, "You can link up to 50 issues for each relationship type", REST/GraphQL/webhook 지원. 공식 문서는 dependency가 close·merge를 강제 차단한다고 말하지 않음(시각 표시만). `gh issue create/edit`에 `--blocked-by/--blocking/--add-blocked-by…` 플래그 존재. | https://github.blog/changelog/2025-08-21-dependencies-on-issues/ · https://docs.github.com/en/rest/issues/issue-dependencies · https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-issue-dependencies | 2026-11-08  | 검증 |
| PR의 `Closes #n` 자동 종료는 기본 브랜치(main)로 merge될 때만 동작: "When you merge a linked pull request into the default branch of a repository, its linked issue is automatically closed." 따라서 `v2026.11.0` 대상 PR에서는 auto-close가 일어나지 않으며 Issue 종료는 리더가 증거 확인 후 수동으로 한다(계획 §8의 `Refs #n` 기본 연결과 일치). | https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue | 2026-11-08  | 검증 |
| REST 모범 사례: 조건부 GET(etag/if-none-match, 304는 primary rate limit 미차감), "Conditional requests for unsafe methods, such as POST, PUT, PATCH, and DELETE are not supported", 쓰기 요청은 직렬·1초 간격, `retry-after`·`x-ratelimit-reset` 준수. ETag는 lock이 아니다. | https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api | 2026-11-08  | 검증 |
| REST API 버전: `2026-03-10`(지원 종료 미정)과 `2022-11-28`(2028-03-10 종료). 헤더 생략 시 2022-11-28 기본. | https://docs.github.com/en/rest/about-the-rest-api/api-versions | 2026-11-08  | 검증 |
| Rulesets는 branch protection과 병존·누적 적용("all applicable rules are enforced"), enforcement active/disabled, 저장소당 최대 75개, fnmatch 대상. 브랜치 보호를 둘지는 미결. | https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets | 2026-11-08  | 검증 |
| `2026.11.0-rc.1`은 SemVer 2.0.0에서 유효(선행 0 없음, `rc.1` pre-release), `2026.11.0`보다 낮은 precedence. Git tag 후보 `moai-cli-v2026.11.0-rc.1`(release tag; 브랜치는 `v2026.11.0`)은 저장소 tag ruleset "Release tag immutability (v*)"에 **걸리지 않는다**(접두사가 `moai-cli-`). 따라서 같은 version 재배포 금지는 ruleset이 보장하지 않으며 `moai-cli-v*` 전용 ruleset 추가 여부는 08-D HOLD(운영자 결정) | https://semver.org/ · local ruleset 조회 | 2026-11-08 · SemVer 2.0.0 | 검증 |
| Issue forms는 "currently in public preview and subject to change"이며 기본 브랜치에 commit되어야 선택기에 나타남("Commit your file to the default branch"). RC 브랜치의 form 렌더링은 계속 not_run. | https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms · https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/configuring-issue-templates-for-your-repository | 2026-11-08  | 검증 |
| 릴리스 gate 5항목·필수 실패 gate·M0~M5/RC 승격 완료 조건은 설계 00:1284-1304, 12:149-163(단계별 gate), 12:201-218(RL-01~RL-10), 07:94-110(RC1 acceptance)에 있다. 승인 계획 §11 GH-PLAN-*/GH-CLAIM-*/GH-PR-EVIDENCE/GH-TEMPLATE-*/GH-RC-PATH-ISOLATION 11개가 이 Issue의 verification matrix 행이다. | /Users/goos/MoAI/moai-rc1-handoff-ref/moai-rc1-handoff/design-snapshot/00-page_43faf222414c81919de71f7d7830e26c.md · …/12-page_25fa9b23c690819180a1351d6da22e23.md · …/07-page_c4f994be97548191b41945af0713ecc9.md · /tmp/moai-rc1-approved-plan-20261108.md §11 | 2026-11-08 · snapshot 2026-11-08 04:49Z / plan copy sha256 5b6c011c…a627d | 검증 |
| P04·P05·P06·P08·P09·P10·P11·P12·P14·P15는 설계 체크리스트 §7에서 미체크(미결). P01·P02·P03·P07·P13은 결정 완료라 다시 묻지 않는다. | /Users/goos/MoAI/moai-rc1-handoff-ref/moai-rc1-handoff/design-snapshot/01-page_cfa4c2d5bea481918690cc7e131bd86f.md:254-282 | 2026-11-08  | 검증 |
| 발행 manifest는 링크 49/49(정본 50 중 `00-C←00-B` 1건 HOLD 미연결)를 기록. 이번 세션은 #1974 blocking → #2013, #1973/#1974 blocked_by 없음만 재조회했고 나머지 47건은 manifest 기록값이다. | /Users/goos/MoAI/moai-adk/.moai-cli/manifests/github-issues.json · local gh api …/dependencies/blocked_by, blocking | 2026-11-08 · manifest schema_version 1 (untracked 파일) | 미확인 |

## SPEC 요구(GEARS)
- [REQ-ROOT-01] 시스템(리더)은 항상 #1975 본문에 릴리스 gate 5항목(양 host 발견·호출→SETUP→UPDATE / 서버·DB 없는 Issues online 완주 / API 없는 native 판단 완주 / 필수 실패 gate 0건 / 독립 release review)을 체크리스트로 유지하고, 각 항목은 영역 Issue의 증거(명령·exit code·receipt·exact SHA) 링크 없이는 체크 상태로 바꾸지 않아야 한다.
- [REQ-ROOT-02] WHEN 영역 상위 Issue(#1976~#1984) 하나가 닫히면, 리더는 `closed_as`가 completed인지와 하위 leaf 전부가 completed인지 재조회한 뒤에만 해당 영역 행을 완료로 표시해야 한다(cancelled·not_planned는 완료가 아니다).
- [REQ-ROOT-03] WHILE P04·P05·P06·P08 — claim 방식 ⑤+① 결정됨(2026-11-08), lease·heartbeat 수치 HOLD를 확정하지 않는다.
- [REQ-ROOT-04] WHERE PR 통합 대상이 `v2026.11.0`(기본 브랜치 아님)이면, 시스템은 PR 본문에 `Refs #n`·exact candidate SHA·로컬 build/test 증거·독립 review 링크를 요구하고 `Closes #n` auto-close에 의존하지 않으며, Issue 종료는 리더가 증거 확인 후 수행해야 한다.
- [REQ-ROOT-05] IF 통합 직전 재조회한 `origin/v2026.11.0` SHA가 본문·manifest에 기록된 SHA와 다르거나 보호 규칙(ruleset/branch protection) 상태가 바뀌었으면, 시스템은 통합을 멈추고 차이를 보고해야 한다(force·reset·기존 ref 삭제 금지).
- [REQ-ROOT-06] 시스템은 항상 verification matrix(§11 GH-PLAN-COVERAGE…GH-RC-PATH-ISOLATION 11행 + 12:201 RL-01~RL-10)에 행별 상태를 passed / failed / not_run / blocked / stale 중 하나로 기록하고, 실행하지 않은 행을 passed로 쓰지 않아야 한다.
- [REQ-ROOT-07] WHEN 설계 정본(Page)이 바뀌면 시스템은 기존 Issue를 일괄 덮지 않고 변경안을 다시 계산해 본문 해시 확인 뒤 정확 diff로만 갱신해야 한다.
- [REQ-ROOT-08] 시스템은 항상 본문의 모든 산문을 한국어로, 식별자·명령·경로·URL은 원형으로 유지해야 한다.

## 인수 기준과 검증
- AC-ROOT-01: #1975 하위 sub-issue가 정확히 9개(#1976~#1984)이고 모두 milestone #1에 속한다 · 검증: `gh api -H 'X-GitHub-Api-Version: 2026-03-10' repos/modu-ai/moai-adk/issues/1975/sub_issues --jq '[.[] | {n:.number, m:.milestone.number}]' ` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-ROOT-02: milestone `2026.11.0-rc.1`에 44개(04-D #2016 추가) Issue가 모두 등록되어 있다(open+closed = 44) · 검증: `gh api repos/modu-ai/moai-adk/milestones/1 --jq '{open_issues, closed_issues}'` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-ROOT-03: 정본 §5 선행 관계 50건 중 49건이 blocked_by로 연결되고 `00-C←00-B` 1건만 HOLD로 미연결이다 · 검증: `for n in $(seq 1985 2015) 1974 1973; do echo -n "$n "; gh api repos/modu-ai/moai-adk/issues/$n/dependencies/blocked_by --jq 'map(.number)|tostring'; done` · 결과: not_run (manifest 기록 49/49; 이번 세션은 #1974·#1973 2건만 재조회)
- AC-ROOT-04: 통합 대상 브랜치 `v2026.11.0`가 원격에 존재하고 SHA가 본문 기록과 일치하며 `2026.11.0-rc.1` ref가 보존되어 있다 · 검증: `git ls-remote origin refs/heads/v2026.11.0 refs/heads/2026.11.0-rc.1` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-ROOT-05: `v2026.11.0`의 보호 상태(branch protection·ruleset)가 본문에 기록되어 있고 통합 직전 재조회 값과 같다 · 검증: `gh api repos/modu-ai/moai-adk/branches/v2026.11.0/protection; gh api repos/modu-ai/moai-adk/rulesets --jq '.[]|{name,target,enforcement}'` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-ROOT-06: 릴리스 gate 5항목 각각에 영역 Issue의 증거 링크(명령·exit·receipt·SHA)가 있고, 증거 없는 항목은 미체크 상태다 · 검증: `gh issue view 1975 -R modu-ai/moai-adk --json body --jq .body | grep -c '^- \[x\]'  # 증거 링크 없는 [x]가 0이어야 함` · 결과: not_run (현재 본문 5항목 모두 미체크)
- AC-ROOT-07: verification matrix(§11 11행 + RL-01~RL-10)가 본문에 있고 모든 행에 passed/failed/not_run/blocked/stale 중 하나가 적혀 있다 · 검증: `gh issue view 1975 -R modu-ai/moai-adk --json body --jq .body | grep -E 'GH-(PLAN|CLAIM|PR|TEMPLATE|RC)-|RL-0[1-9]|RL-10' | wc -l  # 21 이상` · 결과: failed (현재 본문에 matrix 없음 → H2에서 추가)
- AC-ROOT-08: HOLD 10건(P04·P05·P06·P08·P09·P10·P11·P12·P14·P15) 각각에 선택지·영향·근거·확인일이 있고 작성자가 결정한 항목이 없다 · 검증: `gh issue view 1975 -R modu-ai/moai-adk --json body --jq .body | grep -c '상태: HOLD'  # 10 이상` · 결과: failed (현재 본문은 P 항목 이름만 나열, 선택지 없음)
- AC-ROOT-09: 독립 release review verdict(작성 context와 분리, 주체·공유 범위 기록)가 링크되어 있다 · 검증: `gh api repos/modu-ai/moai-adk/issues/1975/comments --jq '.[]|select(.body|test("release review"))|.html_url'` · 결과: not_run
- AC-ROOT-10: 필수 실패 gate 7종(사용자 파일 손실·무승인 mutation/전송·wrong-worktree·secret leak·근거 없는 PASS·stale admission·중복 외부 효과) 각각 0건의 근거가 영역별로 연결되어 있다 · 검증: `gh issue view 1975 -R modu-ai/moai-adk --json body --jq .body | grep -c '실패 gate'  # 항목별 링크 존재 수동 확인` · 결과: not_run
- AC-ROOT-11: 본문 산문이 한국어이고 식별자·명령·경로·URL은 원형이며, 갱신 전 기존 본문 sha256을 기록했다 · 검증: `gh issue view 1975 -R modu-ai/moai-adk --json body --jq .body | shasum -a 256` · 결과: not_run (갱신 시 H2가 기록)

## 구현 계획
1. 1. (H2 리더) `gh issue view 1975 --json body`로 현재 본문 취득 → sha256 기록
2. 2. 본문 재구성: 기존 5개 절 유지 + ① 통합 대상 `v2026.11.0`·SHA·보호 상태·`2026.11.0-rc.1` 보존 ② PR 계약 요약(REQ-ROOT-04) ③ verification matrix 21행(상태 전부 not_run으로 시작, AC-ROOT-01/02/04/05는 이번 관측값 기록) ④ HOLD 표 11건(P 10건 + 브랜치 보호) ⑤ 영역 링크 재조회 결과
3. 3. 갱신 직전 본문을 다시 읽어 sha256 동일 확인 후 `gh issue edit 1975 --body-file`(사용자 허용 범위 안에서) — 불일치면 중단 보고
4. 4. manifest에 receipt(before/after sha256, 시각, 명령) 기록
5. 5. agent-2 독립 검토(근거 사실성·HOLD·언어) → 지적 반영
6. 6. 이후 각 영역 Issue 종료 시 REQ-ROOT-02 절차로 행 갱신; release review 전 REQ-ROOT-05 재조회

## 구현·검증 체크리스트
- [ ] 기존 #1975 본문 sha256 기록 후 정확 diff로만 갱신(일괄 덮어쓰기 금지)
- [ ] 릴리스 gate 5항목 유지, 각 항목에 영역 Issue 증거 링크 자리 추가(증거 없으면 미체크)
- [ ] verification matrix 추가: §11 GH-* 11행 + 12:201 RL-01~RL-10, 행별 상태 어휘 passed/failed/not_run/blocked/stale
- [ ] HOLD 표 추가: P04·P05·P06·P08·P09·P10·P11·P12·P14·P15 + `v2026.11.0` 보호 규칙 + `00-C←00-B` 선행, 각 선택지·영향·근거·확인일
- [ ] 통합 대상 `v2026.11.0` SHA f6808390… · 보호 상태(미보호, tag ruleset v*) · `2026.11.0-rc.1` 보존 사실을 본문에 기록
- [ ] PR 계약 요약 기록: 대상 v2026.11.0, `Refs #n`, exact SHA, 로컬 build/test 증거, 독립 review, auto-close 비의존
- [ ] 영역 9개 링크(#1976~#1984) 재조회 결과와 manifest 일치 확인
- [ ] 발행 manifest(`.moai-cli/manifests/github-issues.json`)에 본문 갱신 receipt(기존 해시·새 해시·시각) 기록
- [ ] 독립 검토(agent-2)에서 근거 사실성·HOLD·언어 규칙 확인
- [ ] 완료 판정은 영역 9개 closed_as completed + release review PASS + 최종 SHA/보호/HOLD 재조회 뒤에만

## 의존성(실제 선행)
- #1976~#1984 (영역 상위 9개) — 완료 조건의 입력
- #2015 08-D 문서 지원표와 release 검토 — release review verdict 생산자
- 사용자 결정: P04·P05·P06·P08·P09·P10·P11·P12·P14·P15, `v2026.11.0` 보호 여부

## 파일·공유 interface 범위
- 변경 가능:
  - GitHub Issue #1975 본문(해시 확인 후 정확 diff)
  - .moai-cli/manifests/github-issues.json (receipt 추가)
  - docs/github-plan-proposal.md (사용자 dirty 변경과 충돌 없을 때만, 별도 승인)
- 변경 금지:
  - .moai/ · ~/.moai/ · .claude/ (Go 세대·사용자 자료)
  - Go 세대 .github/* 23개 파일·.gitlab/·references/
  - 원격 main · 2026.11.0-rc.1 ref · 기존 label 삭제/색 변경
  - 사용자 미커밋 변경 5파일(.github/ISSUE_TEMPLATE/moai-*.yml, .github/moai/README.md, docs/github-plan-proposal.md)
  - 영역·leaf Issue 본문(각 그룹 담당)
- 공유 interface:
  - milestone #1 `2026.11.0-rc.1`
  - sub-issue/blocked_by 관계(리더 단독 projection writer)
  - 상태 보고 댓글 형식(stable run marker, 00:1170)

## 테스트와 로컬 빌드
- 코드 변경 없음 → cargo/python 시험 해당 없음
- 읽기 검증: gh api -H 'X-GitHub-Api-Version: 2026-03-10' repos/modu-ai/moai-adk/issues/1975/sub_issues --jq 'length'
- 읽기 검증: gh api repos/modu-ai/moai-adk/milestones/1 --jq '{open_issues,closed_issues}'
- 읽기 검증: git ls-remote origin refs/heads/v2026.11.0 refs/heads/2026.11.0-rc.1
- 읽기 검증: gh api repos/modu-ai/moai-adk/branches/v2026.11.0/protection; gh api repos/modu-ai/moai-adk/rulesets
- 본문 해시: gh issue view 1975 -R modu-ai/moai-adk --json body --jq .body | shasum -a 256

## 완료 조건과 PR 연결
이 Issue 자체는 PR을 만들지 않는다(문서·관계 추적 Issue). 완료 조건: 영역 9개 모두 closed_as completed, 릴리스 gate 5항목 증거 링크 체크, 필수 실패 gate 0건 근거, 독립 release review PASS(주체·context 분리 기록), 통합 직전 `origin/v2026.11.0` SHA·보호 상태·HOLD 잔여를 재조회해 본문에 기록. 현재 기록 SHA는 f6808390bd83c53e44964ed1d4645727945ea1cc(변경 시 갱신). 첨부 증거: 재조회 명령 출력, manifest receipt, 영역별 verdict 링크.
통합 대상 브랜치: `v2026.11.0`. 이슈별 독립 worktree, 로컬 build/test 증거, 독립 review 뒤 PR.

## 미실행·위험
- not_run: blocked_by 47건 재조회(manifest 값 의존), 릴리스 gate 전부, verification matrix 전 행, release review
- not_run: cargo test/clippy/fmt(이 세션은 읽기 전용; Python 22 tests OK는 관측)
- 위험: `v2026.11.0` 미보호 → 실수 force/삭제를 기계적으로 막지 못함(HOLD)
- 위험: Issue forms public preview·기본 브랜치 전용 → RC 브랜치에서 렌더링 계속 not_run
- 위험: milestone 완료율을 시험 통과율로 오인
- 위험: 설계 스냅샷(04:49Z)과 Page 최신본 차이 미대조

## 결정이 필요한 선택지 (미결은 HOLD)
- **P04 첫 RC 인증 순서와 선택 기능(M3 일부·M4·M5)의 milestone 포함 범위** (P04) · 상태: HOLD
  - A. 양 host 필수 gate(M0~M2)만 rc.1 milestone, M3 선택·M4·M5는 후속 milestone · 영향: rc.1 범위 명확, GitLab·ZCode·remote adapter는 지원표에 미검증 표시 · 근거: 설계 00:1286 필수 gate, 01:260 P04 미결, 12:19 범위 변경안은 사용자 확정 대체 불가 (확인일 2026-11-08)
  - B. 현재대로 32 leaf 전부를 rc.1 milestone에 두고 미통과 기능은 release에서 명시 제외 · 영향: milestone 완료율이 시험 통과율로 오인될 위험, 제외 목록 관리 필요 · 근거: 승인 계획 §3 milestone 완료율은 시험 통과율이 아님, §4 선택 기능은 별도 child/후속 후보 (2026-11-08)
- **P05 Rust package 경계와 한시 Go 호환 bridge** (P05) · 상태: HOLD
  - A. 단일 crate `moai-cli`(현재 Cargo.toml) 유지, bridge 없음 · 영향: 구현 단순, Go 기능 재사용 불가 → 모두 재구현 · 근거: /Users/goos/MoAI/moai-adk/Cargo.toml publish=false 단일 패키지, AGENTS.md "No Go bridge is implicitly selected" (2026-11-08)
  - B. domain crate + adapter crate 분리(설계 12:21 Rust domain/process 인터페이스) · 영향: 템플릿 crate 배치(§9 `<Rust-source>/templates/github/`) 결정과 연동 · 근거: 설계 12:21-40, 승인 계획 §9 "실제 crate 배치는 P05와 대조" (2026-11-08)
- **P06 typed protocol의 schema·error/status·version 규칙과 MCP/CLI parity 범위** (P06) · 상태: HOLD
  - A. 현재 구현(exit 0/1, 동작명 상태 어휘)을 00-D에서 설계 어휘(0/1/2, proposed/applied/not_required/blocked/unknown)로 정렬 · 영향: #1973 산출물 변경, 시험 갱신 · 근거: #1973 댓글 revision 1·2 Gap 목록, docs/rust-product-path.md Known gaps (2026-11-08)
  - B. MCP parity를 rc.1에서 제외하고 CLI만 · 영향: 01-C 범위 축소, 기존 MCP 소비자 mapping 지연 · 근거: 설계 01:264 P06 미결 (2026-11-08)
- **P08 — claim 방식 ⑤+① 결정됨(2026-11-08), lease·heartbeat 수치 HOLD
  - A. 로컬 journal + OS lock(단일 리더), 무응답 자동 takeover 없음 · 영향: 여러 Mac 동시 배차 불가, 명시 handoff만 · 근거: 승인 계획 §8, 설계 11:72·11:116, docs/github-plan-proposal.md §3.1 (2026-11-08)
  - B. 원격 Git ref create-if-absent 또는 CI concurrency를 claim primitive로 · 영향: 원격 왕복·권한 필요, exactly-once 주장 불가 · 근거: docs/github-plan-proposal.md §5-10 선택지 ①~④ (2026-11-08)
- **P09 승인 SPEC snapshot·대형 artifact 저장소와 retention** (P09) · 상태: HOLD
  - A. `.moai-cli/contracts.lock` tracked + 작은 Git snapshot · 영향: 이 저장소는 `.moai-cli/` 추적 결정됨(2026-11-08)이라 가능 · 근거: 설계 00:1003, 08:74, 22:62; 사용자 결정 `.moai-cli/` 추적 (#1973 revision 5)
  - B. 기존 CI artifact/Release asset + Issue Artifacts(public preview, 영구 삭제 가능) · 영향: 장기 감사 원본 단독 위탁 불가, pin/export 필요 · 근거: 설계 00:659 Issue Artifacts 주의 (2026-11-08)
- **P10 판단 adapter 기본 모델·threshold·abstain·API 예산 상한** (P10) · 상태: HOLD
  - A. rc.1은 native 판단만, API 판단 adapter는 비활성 기본 · 영향: 07-D 범위 최소, 비용 0 · 근거: 설계 01:272 "API 없는 native 판단은 확정" (2026-11-08)
  - B. 선택 API 판단을 capability·비용 검증 뒤 opt-in · 영향: 비용 예약·abstain 정책 필요 · 근거: 설계 12:149 M3 행 (2026-11-08)
- **P11 token/context·polling·cache·artifact 수치 budget** (P11) · 상태: HOLD
  - A. 측정 전 수치 고정 없음, baseline 측정 후 결정 · 영향: 08-D 지원표에 수치 공란 · 근거: 설계 01:274, 00:1304 "실측 전 향상률 약속 금지" (2026-11-08)
  - B. 임시 운영값(예: polling backoff)만 설정 파일 기본값으로 · 영향: 제품 보장이 아님을 문서에 명시 필요 · 근거: 설계 00:630 polling 기본 (2026-11-08)
- **P12 locale·코드명·문서 채널·지원 OS/arch release matrix** (P12) · 상태: HOLD
  - A. rc.1은 macOS(개발 Mac) 실측만 '지원', 그 외 미검증 표기 · 영향: AGENTS.md "Do not claim Windows/Linux compatibility from Mac fixtures" 준수 · 근거: AGENTS.md 개발 지침, #1973 not_run(Windows/Linux CRLF) (2026-11-08)
  - B. CI matrix로 Linux/Windows 추가 후 지원표 확장 · 영향: GitHub Actions 추가 → 00-C 비목표(Actions 자동화 추가)와 충돌 가능 · 근거: #1974 Non-goals (2026-11-08)
- **P14 후속 Desktop framework·OS·signing** (P14) · 상태: HOLD
  - A. rc.1 범위 밖으로 두고 #1975에는 링크만 · 영향: 없음(확정: Desktop은 후속) · 근거: 설계 01:280, 12:173 (2026-11-08)
- **P15 — ZCode 헤드리스 범위 결정됨(2026-11-08), 그 외 host 지원표 HOLD
  - A. rc.1 미포함, 후속 milestone 후보로만 기록 · 영향: 없음 · 근거: 설계 01:282, 승인 계획 §4 (2026-11-08)
- **`v2026.11.0` 브랜치 보호(ruleset/branch protection) 도입 여부**  · 상태: HOLD
  - A. ruleset 추가: PR 필수·force push 차단·삭제 차단(enforcement active) · 영향: 원격 설정 변경(외부 효과) → 사용자 승인 필요; 리더 FF/merge 절차도 PR 경유로 바뀜 · 근거: 현재 protection 404·tag ruleset만 존재(로컬 조회), about-rulesets 문서 (2026-11-08)
  - B. 보호 없이 운영 규칙(REQ-ROOT-05 재조회·force 금지)으로만 통제 · 영향: 설정 변경 없음, 실수 force·삭제를 기계적으로 막지 못함 · 근거: 사용자 지시 "기존 원격 브랜치 삭제·force·reset 금지" (2026-11-08)

## 관련 논문·기술자료
- Modern Release Engineering in a Nutshell: Why Researchers Should Care (Bram Adams, Shane McIntosh 2016 IEEE SANER 2016 (FOSE track), pp. 78-90, DOI 10.1109/SANER.2016.108) — 릴리스 파이프라인(build·CI·integration·deployment·release)을 단계별 증거로 나누는 관점이 이 Issue의 gate/verification matrix 구성 근거. 서지는 검색 결과 다수 일치(researchr·dblp·Springer 인용)이나 원문 페이지 직접 조회는 연결 거부로 실패. · https://researchr.org/publication/AdamsM16/bibtex · 미확인

## 베스트 프랙티스 채택/비채택
- GitHub REST 쓰기는 직렬·1초 간격, 조건부 GET으로 재조회, `retry-after`/`x-ratelimit-reset` 준수, 쓰기 메서드에 조건부 요청 기대 금지 · 판정: 채택 · 이유: 본문·관계 갱신을 리더 1명이 순차 수행하는 설계(single projection writer)와 일치 · 출처: https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api (2026-11-08)
- REST 호출에 `X-GitHub-Api-Version: 2026-03-10`(또는 2022-11-28) 헤더를 명시해 기본값 변경에 대비 · 판정: 채택 · 이유: 재조회 명령의 재현성 확보 · 출처: https://docs.github.com/en/rest/about-the-rest-api/api-versions (2026-11-08)
- 비기본 브랜치 대상 PR에서는 `Closes #n` auto-close를 기대하지 않고 `Refs #n` + 수동 종료 · 판정: 채택 · 이유: 대상이 `v2026.11.0`이므로 auto-close 불가; Issue 종료는 증거 확인 후 리더가 · 출처: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue (2026-11-08)
- 릴리스 브랜치에 ruleset(PR 필수·force push 차단·삭제 차단) 적용 · 판정: HOLD · 이유: 원격 설정 변경은 외부 효과라 사용자 결정 필요 · 출처: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets (2026-11-08)
- Release engineering: 파이프라인 각 단계(build·test·review·integration·release)의 증거를 분리하고 PR 개설·merge·release를 별개 상태로 취급 · 판정: 채택 · 이유: 승인 계획 §8 "PR 개설 완료가 merge/release 허가를 만들지 않음"과 같은 원칙 · 출처: https://researchr.org/publication/AdamsM16/bibtex (2026-11-08)

## 조사 기록
- 2026-11-08 /tmp/moai-rc1-approved-plan-20261108.md 전문(§1~§12) 읽음
- 2026-11-08 /Users/goos/MoAI/moai-adk/.moai-cli/manifests/github-issues.json(issues·links) 읽음
- 2026-11-08 gh issue view 1975/1974/1973/1976 --json body; gh api …/1974,1973,1975 comments; sub_issues(1975,1976); dependencies(1974,1973); milestones; rulesets; branches/v2026.11.0/protection; pr list; repo default_branch
- 2026-11-08 git ls-remote origin / git worktree list / git status --short / git check-ignore -v / git ls-files
- 2026-11-08 design-snapshot 00(343-360, 600-680, 1170-1215, 1270-1330, 1397-1406), 01(§7 254-282 P 항목), 06 전문, 07(94-124), 12(149-165, 201-220), source-index.json
- 2026-11-08 /Users/goos/MoAI/moai-adk: AGENTS.md(관리 구간 99-111), docs/rust-product-path.md, docs/github-plan-proposal.md, .github/moai/README.md, .github/moai/labels.yml, src/main.rs, skills/moai/SKILL.md, skills/moai/references/routes.json, tests/github_assets_test.py, tests/product_cli.rs(시험 이름), Cargo.toml, git diff(미커밋 5파일)
- 2026-11-08 web: docs.github.com sub-issues REST/UI, issue-dependencies REST/UI, github.blog changelog 2025-08-21, syntax-for-issue-forms, configuring-issue-templates, about-milestones, REST best-practices, api-versions, about-rulesets, linking-a-pull-request-to-an-issue, managing-issue-types; semver.org; code.claude.com/docs/en/memory; learn.chatgpt.com/docs/agent-configuration/agents-md; agentskills.io/specification; doc.rust-lang.org cargo-test; git-scm.com git-check-ignore
- 2026-11-08 web 논문 서지 검색: Adams&McIntosh SANER 2016, Khanna/Kunal/Pierce FSTTCS 2007, Cabot et al. SANER 2015, Bissyandé et al. ISSRE 2013 (dblp·Springer·IEEE 직접 조회는 bot 차단/빈 응답)

## claim
상태: unclaimed. 배정은 리더의 조정 record에서 단일 writer로 기록하고 이 Issue의 소유 댓글에 표시한다. assignee·label은 lock이 아니다.

상태 규칙: Issue label·assignee는 lock·상태가 아니다. 진행·차단·완료는 소유 보고 댓글(stable run marker) 하나로 보고한다. 미결 기술 선택은 HOLD로 두고 템플릿이나 작성자가 확정하지 않는다. 실행하지 않은 검증은 not_run이다.
언어 규칙: 제목·본문·SPEC·체크리스트·설명 문구는 한국어로 쓰고, 코드 식별자·명령·경로·파일명은 원형을 유지한다.

출처: https://chatgpt.com/space/page_66acadc28bdc8191875556986da72e6b (Page updated_at 2026-11-08T05:42:40Z) · 상세 설계 근거: `.moai-cli/manifests/issue-design-research.corrected.json`
> P12 중 결정됨(2026-11-08, docs/decisions-2026-11-08.md): 배포 채널·서명·tag 접두사·binary 경로·Windows 교체. 나머지 P12·P04·P05(일부)·P06·P08 — claim 방식 ⑤+① 결정됨(2026-11-08), lease·heartbeat 수치 HOLD.
