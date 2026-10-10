# decision-index.md — SPEC-LOCAL-MAIN-FLOW-001

> 상태축 없음(stateless). 이 색인은 계획 단계에서 운영자가 보유한 결정 2건을 기록한다. 두 항목 모두 미결이며 `Operator verdict:` 줄은 비워 둔다. 이 파일은 어떤 옵션도 추천하거나 선택하지 않는다. 리더가 전달한 지시와 리더의 결정은 커밋된 문서가 아니므로 권한 앵커로 쓰지 않는다. 운영자가 해당 행의 `Operator verdict:` 를 채우면 `spec.md` §I 가 가리키는 수용 기준의 게이트가 열린다. 재동기화(OQ-8)의 fast-forward 설계는 `plan.md` §B3a 와 REQ-LMF-014 에 있다. 이 색인이 다루는 것은 그 설계가 정하지 않은 갈라진(diverged) 경우의 처리 방법뿐이다. 이 색인의 주장은 `[local]`(로컬 저장소: HEAD, 로컬 main, 기본 체크아웃) 또는 `[origin]`(원격 추적 ref 또는 원격 상태)으로 표시하며, 표시가 없으면 `[local]`이다.

### Q1: `workflow.branch_guard.deny_commits_on` 값(`.moai/config/sections/workflow.yaml:178`, 현재 `[main]`)을 유지·변경·사용자별 설정 중 무엇으로 할 것인가?

Label: FOUNDER
Class: product-level (이 값은 공유 체크아웃에서 에이전트 세션의 커밋 거부라는 사용자에게 보이는 가드 동작을 정한다. 분류가 확정되지 않으면 fail-closed로 product-level이다.)
Authority anchor: (none — 커밋된 트리에 이 값을 정하는 조항이 없다. `AGENTS.local.md:199`(규율 5)은 "로컬 main은 commit-dead다"를 현행 사실로 적을 뿐 운영자의 결정으로 적지 않는다.)
Why unresolved:
- 현재 값은 `.moai/config/sections/workflow.yaml:178`의 `deny_commits_on: [main]`이다. 175–177행 주석은 SPEC-MAIN-COMMIT-BAN-001(카드 t1337)을 가리킨다.
- 템플릿 기본값은 빈 목록이다(`internal/config/defaults.go:1491`, `DenyCommitsOn: []string{}`). 배포되는 사용자에게는 이 거부가 없다.
- 이 목록은 에이전트 세션이 Bash 도구로 낸 `git commit`, `git revert`, `git cherry-pick`만 막는다(`internal/hook/branch_guard.go:434`, `protectedCommitPattern`; `:480`, `checkProtectedCommit`). 훅의 입력은 Bash 도구의 명령 문자열이다(`branch_guard.go:1111-1118`). 운영자 자기 터미널의 명령은 훅 밖이다.
- 이 목록은 `workflow.branch_guard.enabled`(`workflow.yaml:174`, 현재 `true`)에 묶여 있다. 가드가 꺼지면 목록은 효과가 없다.
- 착지 동사의 병합은 이 목록을 읽지 않는다(`plan.md` §B3). 그래서 아래 어느 옵션에서도 착지 동사의 경로는 같다. REQ-LMF-009는 Q1이 미결인 동안 현재 값 `[main]`을 유지한다.

Q1이 정하는 것은 목록 값뿐이다. 분기 상태 차단(`git merge` 등)과 착지 동사의 경로는 어느 옵션에서도 바뀌지 않는다(spec.md REQ-LMF-003).

Options:
- (a) `[main]`을 유지한다. 귀결: 에이전트 세션은 기본 체크아웃의 main에 Bash로 커밋·되돌리기·체리픽을 낼 수 없다. 운영자 터미널은 제한되지 않는다. 근거: `workflow.yaml:178`, `branch_guard.go:434·480`.
- (b) `[]`로 바꾼다. 귀결: 이 저장소의 에이전트 세션이 main에 Bash 커밋을 낼 수 있다. `git merge`, `git switch`, `git checkout`, `git reset` 같은 분기 상태 변경 차단은 목록과 별개이므로 그대로 남는다(`branch_guard.go:155`, `:369`). `AGENTS.local.md:199`(규율 5)은 supersede 표식을 받아야 한다(`plan.md` §B10). 템플릿 기본값(`defaults.go:1491`)은 바뀌지 않는다. REQ-LMF-003의 커밋 거부 조항은 목록에 `main`이 있는 동안에만 유효하므로, 이 옵션에서는 그 조항이 적용되지 않는다(spec.md REQ-LMF-003).
- (c) 사용자별 설정으로 둔다. 템플릿 기본값은 `[]`로 그대로 두고, 저장소별 설정이 목록을 정한다. 귀결: 이 저장소도 값 하나를 골라야 하므로 실제로는 (a) 또는 (b)가 된다. 배포 사용자의 기본 동작은 바뀌지 않는다. AGENTS.md 보편 문구(`internal/template/templates/AGENTS.md.tmpl` 92–95행)는 설정 키를 명명할 뿐이므로 바꿀 필요가 없다.

Default: not ranked. The status quo is not a default: while Q1 is undecided, REQ-LMF-009 keeps the current value `[main]`.

Operator verdict: (a) `[main]`을 유지한다. Recorded on the board as d-20261010T045752Z-3fbd (card:t1616).

### Q2: 로컬 main에 대한 첫 병합(이 카드의 착지)의 실행 방법과, 로컬 main이 origin/main과 갈라졌을 때(diverged)의 재동기화 방법을 무엇으로 할 것인가?

Label: FOUNDER
Class: product-level (사용자에게 보이는 main 쓰기 경로를 정하므로 fail-closed 규칙에 따라 product-level이다.)
Authority anchor: (none — 커밋된 트리는 첫 main 쓰기의 실행 주체를 정하지 않는다. 재동기화의 fast-forward 설계는 `plan.md` §B3a 와 REQ-LMF-014 에 있으나, 갈라진 경우의 처리는 정하지 않는다.)
Why unresolved:
- 첫 병합: 에이전트의 Bash `git merge`는 기본 체크아웃에서 차단된다(`internal/hook/branch_guard.go:155`, 패턴 `\bgit\s+merge\s`). 운영자 터미널은 이 훅 밖이다.
- 착지 동사 `moai integration merge --card <id>`는 창을 보유한 레인만 쓴다(`internal/cli/integration_merge.go:24–25`, "holder only"). 기본 체크아웃은 `integration_merge.go:137–139`에서 거부되며, REQ-LMF-002의 설정 게이트 뒤에서만 허용된다.
- 창은 `moai integration acquire`와 `release`로 잡고 놓는다(`.claude/rules/local/gitflow-lane-protocol.md` §3).
- `[origin]` 재동기화: 도구 소유의 fast-forward 전용 병합은 pre-flight에 기록한 origin/main 값(BASELINE_SHA)을 대상으로 하며, fast-forward로 맞을 때만 작동한다(`plan.md` §B3a). 로컬 main에 origin/main이 갖지 않은 커밋이 있으면(갈라짐) 도구는 거부한다. 이 경우의 처리 방법이 이 질문의 나머지다.
- 재측정 기록의 키는 후보 트리이며 병합 트리와 같아야 한다(`internal/homestate/card_evidence_readers.go:313–315`). 그래서 첫 병합의 어느 옵션이든 같은 키의 기록을 쓴다. 기록은 `moai integration remeasure`가 `internal/factory/integration_remeasure.go:550`(`RunRemeasure`)에서 쓰며, 쓰는 자리는 `:610`이다.

Options (first merge):
- (a) 운영자가 자기 터미널에서 기본 체크아웃에 카드 끝 SHA를 `--no-ff`로 병합한다. 귀결: 도구는 병합을 만들지 않는다. 완료 경로(`internal/cli/factory_card.go:1924–1960`, adoption)는 병합을 누가 만들었는지 구별하지 않고 채택한다. 착지 동사의 첫 실제 실행은 이 카드에서 검증되지 않은 채 남는다.
- (b) 설계된 착지 동사가 창 안에서 병합한다(`plan.md` §B3, REQ-LMF-003). 귀결: 병합, 재측정 확인, 전이 기록이 도구에 남는다. 대신 첫 실제 main 쓰기가 새 코드 경로를 탄다. 방어는 거부 조건(HEAD 이름 확인, 전체 clean 요구, 상태 집합 비교)이다.
- (c) 설정으로 고르는 별도 통합 브랜치 면을 쓰고, 승격 전까지 main에 쓰지 않는다(REQ-LMF-002의 두 번째 경우). 귀결: 통합 브랜치와 워크트리가 하나 더 필요하다. 승격은 main 쓰기이므로 같은 질문이 승격 시점으로 옮겨진다.

Options (diverged case of the re-sync, OQ-8):
- (i) 도구는 거부하고 안내만 한다. 어떤 도구 경로도 갈라짐을 해소하지 않으며, 릴리스 배치는 Q2가 정해질 때까지 그 상태에서 멈춘다. 귀결: 갈라짐이 생기면 배치가 운영자의 결정을 기다린다.
- (ii) 도구가 창 안에서 `origin/main`을 `--no-ff`로 병합한다(REQ-LMF-003의 가드 처리 적용). 귀결: 원격에 없는 병합 커밋이 다음 릴리스 PR 전까지 로컬 main에 남는다. 이 옵션은 REQ-LMF-003과 REQ-LMF-014를 개정하는 후속 SPEC을 필요로 하며, 그 후속 SPEC은 자체 수용 기준을 가진다.
- (iii) 별도 동기화 브랜치를 워크트리에서 `origin/main`에 fast-forward한 뒤, 도구가 그 브랜치를 로컬 main에 병합한다. 귀결: 브랜치와 워크트리가 하나 더 필요하다. 병합을 실행하는 주체는 여전히 도구다. 이 옵션도 REQ-LMF-003과 REQ-LMF-014를 개정하는 후속 SPEC을 필요로 하며, 그 후속 SPEC은 자체 수용 기준을 가진다.

Default: not ranked. The status quo is not a default and selects no option: the tool path refuses on the primary checkout today, and no first merge has happened.

Operator verdict: 첫 병합 (b) 설계된 착지 동사가 창 안에서 병합한다. 갈라진 경우의 재동기화 (i) 도구는 거부하고 안내만 한다. Recorded on the board as d-20261010T045752Z-3fbd (card:t1616).
