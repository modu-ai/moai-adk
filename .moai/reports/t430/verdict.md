# t430 정본 판정서 — 레인 push 금지 + 리드 일괄 push (2026-09-02)

lane-1(t430) · 판정자: lane 세션 · 기준선: `origin/develop` = `ad272be20`

## Claim

1. **정본 = `origin/develop` @ `ad272be20` 사본이다.** primary main 사본도, primary 워킹트리 미커밋 사본도 정본이 아니다.
2. 드리프트 표면은 카드 명시 2곳에서 시작해 전수 스윕으로 **5파일 8지점**으로 닫혔다. 이 중 5파일 8지점을 본 카드에서 수정하고, `delivery.md`(2지점)는 보고만 한다.

## Evidence — 판정 근거 4건 (모두 lane 워크트리에서 직접 측정)

| # | 주장 | 측정 |
|---|---|---|
| 1 | 독트린 착지 이력이 develop 쪽 | `git log --oneline -6 origin/develop -- CLAUDE.local.md` → 최상단 `05e58c40a`(t281 M3), `9a161687a` "docs(local): land §4.1 2026-08-29 standing-develop chain + 09-01 WT-push prohibition (t281 prerequisite, operator decision b)". 반면 `git log --oneline -6 origin/main -- CLAUDE.local.md` → 최신이 `ea6706bb9`(ci-loop protocol twins)로 그 이전 — main에는 08-29/09-01 독트린이 커밋된 적 없음 |
| 2 | gitflow-lane-protocol.md는 develop에만 존재 | primary 워킹트리 `ls .claude/rules/local/` → 4파일(해당 파일 없음, 8/15-16 타임스탬프). lane 워크트리(develop 베이스) → 5파일(존재, 14,492바이트). `git ls-tree --name-only HEAD .claude/rules/local/` → 5파일 전부 추적 |
| 3 | primary 워킹트리 사본(660행)은 develop 사본(710행)의 구버전 스냅샷 | `diff origin/develop사본 primary사본` = 176행. primary 유신 라인(`^>`) 35건 **전부 구버전 텍스트**(Last Updated 2026-05-25, 구 `deploy.go:29/122` 행번호, §2.0 에이전트 3사본 절 누락, 2026-08-27 감사 정정 누락, gitflow-lane-protocol.md 포인터 누락). 유신 내용 0건 → 착지 후 primary 사본은 develop 판으로 통째 교체해도 손실 없음 |
| 4 | 리드 실측 라인번호 재확인 | `grep -n "git push origin develop" CLAUDE.local.md`(primary 워킹트리) → 305·306·321행. `git show origin/develop:CLAUDE.local.md \| grep -n ...` → 348·349·364행. 리드 보고와 정확히 일치. 306/349행은 유지 대상인 09-01 [HARD] 조항 본문이고, 305/348(레인 절차 chain)·321/364(운영 절차 code block)가 제거 대상 |

## 드리프트 표면 — 최종 열거: 5파일 8지점 (수정) + delivery.md 2지점 (보고)

열거 방법: 초기 문자열 스윕(`push origin develop`)에 이어, specialist 상호 확인(repo-local-pr-policy·LWD:150)과 **통합 병합+push 동시 언급 패턴 스윕**(git-workflow-doctrine.md:103 추가 발견)으로 닫음 — 동일 패턴 추가 히트 0.

| # | 파일 | 지점 | 처분 |
|---|---|---|---|
| 1 | CLAUDE.local.md §4.1 | 348행(레인 chain)·364행(운영 절차 code block) + 규율 2/4(327·329행) 정밀화 + 리드 일괄 push 절차 신설 | **수정** |
| 2 | .claude/rules/local/gitflow-lane-protocol.md | §4 69행 "병합 후 레인이 직접 올린다"·§6 81행·§2 29행(위임 문언)·§7 | **수정** — §2에 delivery.md Step 3.2 제6단계 EXCLUDED 예외 명시, §4 전면 개정, §7 리드 일괄 push 소관 신설. 8행(전환 사기)·34행(push 대상 사실문)·58행(HEAD 재판독 — 리드 push에도 적용)은 유지 |
| 3 | .moai/docs/git-workflow-doctrine.md | §18.3 103행(병합 전략 표 인트로 [HARD])·§18.8 329행(Patch Release 절차 1단계) | **수정** — 살아있는 독트린 2지점. CLAUDE.local.md §References가 §18을 가리키므로 미수정 시 재생산 경로 |
| 4 | .claude/rules/local/repo-local-pr-policy.md | 12행 "; lanes push `origin/develop`" 절 | **수정** — 리드 일괄 문언으로 교체, 판정면 문장 유지 |
| 5 | .moai/docs/git-local-workflow-doctrine.md | §23.7 150행 "합친 뒤 push한다" | **수정** — 리드 일괄 문언으로 교체, RETIRED 마커·enforce_admins 노트 유지. (§23.9(a) 157행은 이미 push-neutral이라 유지) |
| 6 | .claude/skills/moai/workflows/sync/delivery.md | 278행(Step 3.2 WT-* 경로 제6단계)·299행(develop 직접 push 경로) | **보고만, 수정 없음** — (a) 분산 스킬이라 이 리포의 레인 경제를 못박을 수 없음. (b) 템플릿은 t335(`5095e3059`·`41dd0e82c`)가 이미 `<integration-branch>`로 일반화했고 배포본은 t303(`63b4628a6`) 이전 스테일 → 다음 `moai update`에서 템플릿판으로 자가 치유. (c) 2번의 §2 EXCLUDED 예외가 레인 계층에서 봉쇄. (d) 299행 develop 직접 경로는 신설 독트린에서 **리드의 일괄 push**가 바로 그 동작이므로 모순 없음 |

## Baseline-attribution

모든 측정은 lane 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t430/` (베이스 `ad272be20` = `origin/develop`, 진입 직후 `git log --oneline -1` 확인)와 primary 워킹트리 읽기 전용 대상으로 수행. 측정 시점 2026-09-02, 세션 `7b1d3122-f36e-4a4e-abef-128fb8bfdd7a`.

## Gaps

- delivery.md 299행 경로(develop 브랜치 체류 세션)의 실사용 주체 분포는 관측하지 않았다 — 독트린상 리드가 develop 통합 워크트리를 보유하므로 모순 없음으로 판정했을 뿐, 누가 그 경로로 deliver하는지의 이력 통계는 없다.
- primary 워킹트리의 **다른** 미커밋 파일(AGENTS.md, feedback.yaml, git-strategy.yaml, astgrep yml 2건)은 본 카드 대상 밖이라 분류하지 않았다 — CLAUDE.local.md만 전수 diff했다.

## Residual-risk

- 리드가 본 카드 병합과 사이에 `origin/develop`에 추가 push를 하면 ①-③의 행번호가 어긋난다 — 내용 기준 편집이므로 영향은 행번호 표기뿐이다.
- ④의 자가 치유는 "다음 `moai update` 실행"이 전제다. update가 실행되지 않는 한 배포본 delivery.md:278은 레인에게 push를 지시하는 텍스트로 남는다 — ②의 §2 예외가 이를 문서 계층에서 덮지만, delivery.md만 단독으로 읽는 세션(레인 프로토콜 미로딩)에는 노출이 남는다. 완전 봉쇄는 배포본-템플릿 동기화 후 별도 카드 소관.

## 후속

run(manager-docs 위임, 5파일 8지점 편집) → sync(3-phase close, evidence 갱신) → 리드에게 병합 창 요청(로컬 병합 SHA 보고, push 없음). primary 워킹트리 재조정(구버전 CLAUDE.local.md 교체)은 착지 후 운영자/리드 몫 — 유신 내용 0건 확인됐으므로 교체 손실 없음(Evidence #3).
