# SPEC-REVIEW-GATE-REPAIR-001 — 설계: 상태 모델과 판정 표

이 문서는 요구사항을 실현하는 상태 모델을 적는다. 함수 이름과 자료 구조의 세부는 run 단계가 정한다. 여기서는 어떤 상태에서 무엇이 일어나야 하는지, 그리고 운영자 판정이 정한 규칙을 고정한다.

## 1. 리뷰 키와 영수증 상태

리뷰 키는 영수증의 결합 필드 다섯 개의 조합이다: head, tree_digest, config_digest, command, tool_version(`internal/verify/receipt.go:70-76`). 키가 하나라도 다르면 영수증은 다른 키의 것이다.

영수증 상태는 여섯 가지다.

| 상태 | 정의 |
|---|---|
| absent | 같은 키의 영수증이 없다 |
| stale | 같은 키의 영수증이 있으나 유효기간(24시간)이 지났다 |
| fresh-pass | 같은 키, 유효기간 안, 판정 pass |
| fresh-fail | 같은 키, 유효기간 안, 판정 fail |
| fresh-inconclusive | 같은 키, 유효기간 안, 판정 inconclusive |
| key-mismatch | 결합 필드 중 하나 이상이 다르다 |

## 2. 상태별 행동: 현재와 목표

목표 열의 "검토 시작"은 진행 중 표식을 존중한다. 진행 중 표식이 있으면 새 검토를 시작하지 않는다(기존 메커니즘). 간격은 없다(Q4 = (a)).

| 상태 | 진입(프롬프트) 현재 | 진입 목표 | 턴 종료 현재 | 턴 종료 목표 |
|---|---|---|---|---|
| absent | 허용 | 허용 | 검토 시작 후 허용 | 검토 시작 후 허용, 메시지 (REQ-RG-015, 016) |
| stale | 허용 | 허용 | 검토 시작 후 허용 | 검토 시작 후 허용, 메시지 (REQ-RG-015, 016) |
| key-mismatch | 허용 | 허용 | 검토 시작 후 허용 | 검토 시작 후 허용, 메시지 (REQ-RG-015, 016) |
| fresh-fail, 전달 기록 없음 | 차단 | 허용, 상세를 추가 맥락으로 한 번 전달, 전달 기록 (REQ-RG-002) | 차단 (stop_hook_active 가 거짓일 때) | 차단 횟수가 C 미만이면 차단(횟수 +1), 아니면 허용 + 메시지 (REQ-RG-003, 004, 016) |
| fresh-fail, 전달 기록 있음 | 차단 | 허용, 반복 없음 (REQ-RG-002) | 차단 (stop_hook_active 가 거짓일 때) | 차단 횟수가 C 미만이면 차단, 아니면 허용 + 메시지 (REQ-RG-003, 004, 016) |
| fresh-pass | 허용 | 허용 | 허용 | 허용, 메시지 없음 (REQ-RG-013) |
| fresh-inconclusive | 허용 | 허용 | 허용(재사용) | 재검토 시작(표식 없을 때), 허용 + 메시지 (REQ-RG-009, 010, 016) |
| 스위치 꺼짐 | 검토 없음 | 검토 없음 | 검토 없음 | 허용 + 메시지 (REQ-RG-011, 014, 016) |
| 검토할 변경 없음 | 검토 없음 | 검토 없음 | 검토 없음 | 허용 + 메시지 (REQ-RG-014, 016) |
| stop-hook 해제 | — | — | 허용 | 허용 + 메시지 (REQ-RG-014, 016) |
| reviewer 없음 | 허용 | 허용 | 허용 | 허용 + 메시지 (REQ-RG-014, 016) |

## 3. 데이터 모델: 전달 기록과 차단 횟수

- **전달 기록**은 리뷰 키 단위로 하나 있다. 내용은 키와 전달 시각이다. 다른 키에는 적용되지 않는다. 키가 바뀌면 새 키의 기록이 새로 생긴다(REQ-RG-002).
- **차단 횟수**도 리뷰 키 단위다. 턴 종료가 그 키를 차단할 때마다 1 을 더한다. 허용은 값을 바꾸지 않는다(보드 판정 D3). 그 키에 fresh pass 영수증이 기록되면 기록을 제거한다. 키가 바뀌면 새 키는 0 에서 시작한다. C = 1(운영자 판정 Q2)(REQ-RG-003, 004).
- 두 기록은 영수증이 저장되는 상태 루트(`.moai/state/verify/` 아래)에 둔다. 같은 구조가 레인 저장소의 `codex-review/` 와 `snapshots/` 에서 관측되었다(research.md §4, 기계 로컬). 위치의 세부는 run 단계가 정한다. F1 이 영수증 결합 기준을 바꾸면 두 기록의 키도 함께 바뀔 수 있다. 이 SPEC 은 그 변경을 하지 않는다.

## 4. 기준 해석 함수 (M4)

입력은 트리 루트와 세 설정 값이다: `git_strategy.worktree_base_branch`(`.moai/config/sections/git-strategy.yaml:5`), `git_strategy.manual.develop_branch`(같은 파일 :16, 통합 대상), `git_strategy.manual.workflow`(같은 파일 :9, 활성 흐름). 모든 조회는 읽기 전용이다.

### 4.1 후보 집합

- 설정 키 K 마다 두 후보를 둔다: 로컬 ref `K` 와 원격 추적 ref `origin/K`.
- 후보 집합은 활성 흐름 절 (a)–(d) 가 정한다. 절은 순서대로 적용하고, 처음 조건이 맞는 절이 결정한다.
  - (a) `manual.develop_branch` 가 `main` 이거나 `worktree_base_branch` 가 `manual.develop_branch` 와 같으면: 단일 기준은 그 브랜치 하나다. 합집합과 develop 후보를 쓰지 않는다.
  - (b) `manual.workflow` 가 `git-flow` 이면: 후보는 `worktree_base_branch` 와 `manual.develop_branch` 의 합집합이다. 이 저장소의 현재 값(`main`, `develop`, `git-flow`)은 이 절에 해당하므로 후보는 `{main, origin/main, develop, origin/develop}` 이다.
  - (c) `manual.workflow` 가 `github-flow` 이면: 단일 기준은 `worktree_base_branch` 다. develop 은 github-flow 에서 legacy 이므로 후보가 아니다.
  - (d) 설정만으로 활성 흐름을 정할 수 없을 때 적용한다. 그 경우는 다음 둘 중 하나다. (i) 두 키(`worktree_base_branch`, `manual.develop_branch`)를 읽을 수 없는 경우이며, 다른 조건과 무관하게 적용한다. (ii) 두 키가 서로 다른 브랜치를 가리키고, `manual.workflow` 가 없거나 알 수 없는 값이거나 두 키와 모순되는 경우다. 절 (a) 가 먼저 적용되며, 절 (a) 는 `manual.workflow` 를 읽지 않는다. 결과는 INCONCLUSIVE 이며, pass 로 캐시하지 않는다(REQ-RG-009).
- 단일 기준(절 (a), (c))의 후보는 그 브랜치의 로컬 ref 와 원격 추적 ref 다. 동점 규칙(§4.3)의 고정 순서는 후보 집합에 든 후보에만 적용한다.
- 후보는 트리에서 커밋으로 해석될 때만 쓰인다. 해석되지 않은 후보는 제외한다. 모든 후보가 해석되지 않을 때에만 REQ-RG-008 이 작동한다.
- `origin/HEAD` 와 설정 밖의 ref 는 후보가 아니다. 대체 기준으로 쓰지 않는다(REQ-RG-008).

### 4.2 거리와 선택

- 해석된 후보 c 마다 mb(c) = `merge-base(c, HEAD)` 를 구하고, d(c) = `rev-list --count mb(c)..HEAD` 를 잰다. d(c) 는 HEAD 에서 mb(c) 를 뺀 커밋 수, 즉 HEAD 가 그 기준 위에 쌓은 커밋 수다.
- mb(c) 가 없는 후보(공통 조상이 없음)는 제외한다.
- d(c) 가 가장 작은 후보를 고른다.

### 4.3 동점 규칙

- d 가 같은 후보가 여럿이면 그 후보들의 mb(c) SHA 를 비교한다.
- SHA 가 모두 같으면 고정 순서의 첫 후보를 이름으로 낸다. 고정 순서는 `worktree_base_branch` 의 로컬 ref, `worktree_base_branch` 의 `origin/` ref, `develop_branch` 의 로컬 ref, `develop_branch` 의 `origin/` ref 다.
- SHA 가 하나라도 다르면 판정은 inconclusive 이며, 후보 이름 둘과 두 SHA 를 낸다. 어느 쪽도 고르지 않는다(AC-RG-020).

### 4.4 키 불일치 (운영자 판정 Q1 개정 = (가))

- 절 (b) 에서 두 설정 값이 다른 ref 를 가리키면 4.1–4.3 을 그대로 적용한다. 절 (c) 에서는 `worktree_base_branch` 하나로 판정하므로 `manual.develop_branch` 는 기준이 되지 않는다. 키 불일치 자체는 어느 절에서도 inconclusive 의 사유가 아니다(REQ-RG-007).
- 결과에는 두 설정 값과 `manual.workflow` 값, 적용된 절을 함께 기록한다.

### 4.5 후보 없음

- 해석되는 후보가 하나도 없거나, 해석된 후보 가운데 mb(c) 가 있는 것이 없으면 inconclusive 이며 다른 기준으로 넘어가지 않는다(REQ-RG-008). 절 (d) 의 inconclusive 도 같은 결과다. 두 경우 모두 pass 로 캐시하지 않는다(REQ-RG-009).

### 4.6 예시 (이 저장소의 두 구성)

- 절 (b), 카드가 main 의 커밋 X 에서 잘렸고 develop 이 X 뒤로 전진했다면: mb(main) = X, mb(develop) = X. d 가 같고 SHA 가 같으므로 고정 순서로 `main` 이 선택된다(AC-RG-007).
- 절 (b), 카드가 develop 의 커밋 Y 에서 잘렸고 main 이 Y 의 조상 X 에 있다면: mb(develop) = Y, mb(main) = X. `rev-list --count Y..HEAD` 가 `rev-list --count X..HEAD` 보다 작으므로 `develop` 이 선택된다(AC-RG-008).
- 절 (c), `manual.workflow` 가 `github-flow` 이고 카드가 develop 의 커밋 Y 에서 잘렸다면: 후보는 `main` 뿐이고 develop 은 후보가 아니다. 선택은 `main` 이고 merge-base 는 X 다(AC-RG-023).

### 4.7 출력과 소비자

- 출력: 선택된 ref 이름, mb SHA, d, 두 설정 값과 `manual.workflow` 값, 적용된 절(a–d). 영수증이 이름을 싣는 현재 방식과 호환된다.
- 소비자: 턴 종료 게이트, card 범위 리뷰(`cardMergeBase` 계열), 감사 도구(`resolveReviewBase` 와 `resolveReviewMergeBase` 계열). 이 함수 하나를 거친다. 소비자마다 따로 계산하는 경로를 남기지 않는다(REQ-RG-005).

## 5. 실행 단위 스위치 (M6)

스위치는 run 단위에서 값이 하나다. 운영자 판정 Q3 = (b) 주 체크아웃의 미추적 재정의 파일, 보드 판정 D11 = 커밋된 설정값을 따름.

- **재정의 파일**: 주 체크아웃의 `.moai/state/codex-review-gate/run-switch.json`. `.moai/state/` 는 `.gitignore` 401행으로 무시되는 트리다(미추적). 파일 이름과 형식은 run 단계의 설계 세부다. 파일은 실행 식별자와 활성화 값(on/off)을 싣는다.
- **run 의 정의**: factory run 하나. `--run` 플래그가 가리키는 factory run id 로 지칭되며, 활성 run 은 하나다(`internal/cli/factory_card.go` 의 `--run` 도움말).
- **주 체크아웃 해석**: `IdentifyPrimaryCheckout`(`internal/auditreceipt/storeroot.go:142`, 주 체크아웃을 돌려준다) 또는 같은 결과를 내는 해석기를 쓴다. 목표: 연결된 워크트리에서 읽는 모든 레인이 같은 경로에 도달한다(AC-RG-012, AC-RG-013 이 확인한다). 주 체크아웃을 식별할 수 없으면(연결되지 않은 복제본) 재정의 파일을 찾지 않는다. `StoreRoot` 는 트리에 workflow.yaml 이 있으면 그 트리를 돌려주므로 이 용도로 쓰지 않는다.
- **적용 조건**: 재정의 파일이 싣는 실행 식별자가 단일 활성 run 과 같을 때만 적용한다. 다른 식별자이거나 단일 활성 run 을 해석할 수 없으면 파일을 무시하고 커밋된 값을 따른다.
- **커밋된 값**: 판독 트리의 HEAD 커밋에 있는 `.moai/config/sections/workflow.yaml` 의 `codex.review_gate.enabled`. 작업 트리의 미커밋 값은 쓰지 않는다. 판독은 `git show HEAD:` 같은 읽기 전용 조회다.
- **우선순위**: 적용되는 재정의 값 > 커밋된 값.
- **추적 키의 역할**: 기본값의 원천으로만 남는다. 실행을 전역으로 끄거나 켜는 권한은 없다.
- **두 층**: 셸 래퍼와 Go 처리기가 같은 해석 절차를 쓴다(REQ-RG-012).

## 6. 판정 표

| 조건 | 결과 | 기록 |
|---|---|---|
| 키 일치, 기준 해석 성공, fresh pass | 허용, 재검토 없음 | 기준 ref 이름, merge-base SHA |
| 절 (b) 에서 두 설정 값이 다른 ref | 최근접 후보로 판정 (Q1 개정 = (가)) | 두 설정 값, `manual.workflow`, 선택된 이름 |
| 절 (a) 해당 (`manual.develop_branch` = `main`, 또는 `worktree_base_branch` = `manual.develop_branch`) | 단일 기준으로 판정, develop 후보 없음 | 적용된 절, 단일 기준 이름 |
| 절 (c) 해당 (`manual.workflow` = `github-flow`) | `worktree_base_branch` 로 판정, develop 제외 | 적용된 절, 선택된 이름 |
| (d) 설정만으로 활성 흐름을 정할 수 없을 때 적용한다. 그 경우는 다음 둘 중 하나다. (i) 두 키(`worktree_base_branch`, `manual.develop_branch`)를 읽을 수 없는 경우이며, 다른 조건과 무관하게 적용한다. (ii) 두 키가 서로 다른 브랜치를 가리키고, `manual.workflow` 가 없거나 알 수 없는 값이거나 두 키와 모순되는 경우다. 절 (a) 가 먼저 적용되며, 절 (a) 는 `manual.workflow` 를 읽지 않는다. | 결과는 INCONCLUSIVE 이며, pass 로 캐시하지 않는다(REQ-RG-009). | 사유 |
| 해석된 후보 없음 | inconclusive, 다른 기준 없음 (REQ-RG-008) | 사유 |
| 동점이고 merge-base SHA 가 다름 | inconclusive (AC-RG-020) | 두 이름과 두 SHA |
| fresh fail, 차단 횟수 < C | 차단, 차단 횟수 +1 | 차단 횟수 |
| fresh fail, 차단 횟수 = C | 허용 + 메시지 (REQ-RG-004, 016) | 차단 횟수 값 |
| inconclusive | 재검토 시작(표식 없을 때, 간격 없음), 허용 + 메시지 | 시작 시각 |
| absent, stale, key-mismatch | 재검토 시작(표식 없을 때), 허용 + 메시지 (REQ-RG-015) | 상태 이름 |
| 스위치 꺼짐 | 검토 없음, 허용 + 메시지 (REQ-RG-011, 016) | 스위치 출처 (재정의 파일 또는 커밋 값) |
| 검토할 변경 없음 | 허용 + 메시지 (REQ-RG-014, 016) | 사유 |
| stop-hook 해제 | 허용 + 메시지 (REQ-RG-014, 016) | 사유 |

## 7. 관측 가능성

- 메시지 상태 어휘: `absent`, `stale`, `key-mismatch`, `inconclusive`, `차단 횟수 소진`, `스위치 꺼짐`(출처 포함), `검토할 변경 없음`, `stop-hook 해제`, `reviewer 없음`, `기준 판정 inconclusive`(두 이름 포함).
- 메시지는 Stop 훅의 `systemMessage` 로 낸다(`.claude/rules/moai/core/hooks-system.md:103`). 차단 사유 전반의 가시성 확장은 F2 가 맡는다.
- 스위치 출처(재정의 파일인지 커밋된 값인지)를 검토가 꺼진 이유로 함께 기록한다.
