# 결정 색인 — SPEC-DEPLOY-STEP-OPTIN-001

> 이 색인은 상태 축을 갖지 않는다. 이 SPEC 의 생명주기는 spec.md 만이 가진다. 각 행은 Detect → Explain → Ask 를 따르며, 판단이 필요한 행은 선호 답을 싣지 않는다. `Default:` 줄은 게시된 Default 규칙(1 현행 동작 보존, 2 이 SPEC 커밋 하나의 되돌리기, 3 사용자에게 보이는 표면이 작은 것)을 순서대로 적용한 정책 결과이며 추천이 아니다. 규칙이 선택지를 가르지 못하는 행에는 `Default:` 가 없다. 이 색인은 plan 단계(spec v0.1.0)에서 작성되었다. 2026-10-19 개정 2(spec v0.2.0)에서 Q1–Q9, Q14, Q15 의 판정이 채워졌다(각 행의 `Operator verdict`). Q11 은 증거 대기로 열려 있다. 판정이 `Default:` 와 다른 쪽을 택하면 판정이 이긴다.

## 개요

| 번호 | 질문 | 라벨 | 분류 | 기본값 (규칙 적용 결과) | 판정 |
|---|---|---|---|---|---|
| Q1 | 단계 키의 이름과 위치 | FOUNDER | product-level | (a) (rule: 순위 3) | DECIDED (2026-10-19) |
| Q2 | 단계 값의 형태 | FOUNDER | product-level | 없음 (규칙이 가르지 못함) | DECIDED (2026-10-19) |
| Q3 | review_repair 의 기본값 | FOUNDER | product-level | (b) (rule: 순위 1) | DECIDED (2026-10-19) |
| Q4 | release/deploy 를 첫 조각에 넣을지 | FOUNDER | product-level | (b) (rule: 순위 3) | DECIDED (2026-10-19) |
| Q5 | 잘못된 값 하나가 지우는 범위 | FOUNDER | product-level | 없음 (규칙이 가르지 못함) | DECIDED (2026-10-19) |
| Q6 | project.continuation 과 todo_creation 의 관계 | FOUNDER | product-level | (a) (rule: 순위 1) | DECIDED (2026-10-19) |
| Q7 | team 프로필 `auto_push: true` 와 읽는 코드 없는 automation 키 | FOUNDER | product-level | (a) (rule: 순위 1) | DECIDED (2026-10-19) |
| Q8 | 이 저장소의 승인된 로컬 factory 설정을 새 키로 옮기는 방법 | FOUNDER | product-level | (a) (rule: 순위 2) | DECIDED — OPERATOR-HELD (2026-10-19; 설정 파일 쓰기는 운영자 승인 게이트) |
| Q9 | 웹 즉시 중지의 의미 | FOUNDER | product-level | (a) (rule: 순위 3) | DECIDED (2026-10-19) |
| Q10 | 감사 기록의 위치와 필드 | FOUNDER | implementation-level | (a) (rule: 순위 3) | DEFAULT-APPLIED |
| Q11 | t1882 의 결과와 merge 단계의 관계 | EVIDENCE-NEEDED | product-level | 없음 (자료가 아직 없음) | 열림 |
| Q12 | 산문 소비자용 해석기 읽기 동사 이름 | FOUNDER | implementation-level | (a) (rule: 순위 3) | DEFAULT-APPLIED |
| Q13 | policy revision 의 형태 | FOUNDER | implementation-level | (a) (rule: 순위 3) | DEFAULT-APPLIED |
| Q14 | dispatch 단계가 막는 범위 | FOUNDER | product-level | 없음 (규칙이 가르지 못함) | DECIDED (2026-10-19) |
| Q15 | local override 가 고정한 단계에 대한 웹 저장의 처리 | FOUNDER | product-level | 해당 없음 (감사 발견 갈림) | DECIDED (2026-10-19) |

**런을 막는 행(판정 후 상태, 2026-10-19 개정 2).** M2 를 막던 Q1·Q2·Q5, M3 를 막던 Q3·Q6·Q14, M4 의 Q9, M5·M6 의 Q7·Q8 은 판정을 가진다. M3 의 Q11(EVIDENCE-NEEDED)만 열려 t1882 착지까지 M3 을 막는다. M7 은 Q4=(b) 로 철회되었다(마일스톤은 M1–M6). Q8 은 OPERATOR-HELD — 판정(project 계층, 명시 `auto`)은 유효하되 실제 설정 파일 쓰기는 운영자 keep-set 으로 별도 운영자 승인을 게이트로 요구한다(리더 재정 2026-10-19). 승인 전에는 M6 의 (1)(2) 만 진행한다. Q15 는 어떤 마일스톤도 막지 않는다. M2 와 Q3·Q4 의 순서 관계와 재범위 기록은 plan.md M1 을 본다.

**권한 등록부 검토 결과.** DECIDED 와 POLICY-COVERED 로 라우팅할 수 있는 행이 없다.

- 권한 등록부는 제품 문서(`.moai/project/product.md`), 완료 SPEC 의 HISTORY 와 Amendments, `.moai/config/sections/*.yaml` 의 운영자 설정, 헌법이다. 등록부 정의: `.claude/agents/moai/manager-spec.md` 의 결정 색인 절.
- 후보로 본 것과 탈락 이유. SPEC-UPDATE-YAML-PRESERVE-001 은 템플릿에 없는 사용자 키를 병합에서 보존한다고 정했으나, 이 SPEC 의 질문(단계 키의 이름과 위치, 기본값, 관계)과 같은 질문이 아니다. SPEC-PROJECT-CONTINUATION-KEY-001 은 `continuation` 키가 카드 발급 여부를 정하지 않는 표현 키라고 정했으나(코드 주석), 새 단계 정책이 같은 카드 발급을 막는다는 조건은 다룬 적이 없다. 운영자의 카드 문언(큐의 t1895 본문)은 의도를 말하지만 등록부에 속한 문서가 아니다. `AGENTS.md`·`AGENTS.local.md` 와 메모리 항목은 등록부 밖이므로 인용하지 않는다.
- 따라서 14행 모두 FOUNDER 또는 EVIDENCE-NEEDED 를 유지한다. 인용을 검증하지 못한 행을 DECIDED 로 올리지 않는다.

---

### Q1: 단계 키는 어떤 이름으로 어디에 두는가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 단계 일곱 개를 열거했을 뿐 키 이름과 위치를 정하지 않았다. 새 키는 사용자가 직접 쓰고 웹이 편집하는 공개 설정 표면이므로 이름이 곧 호환 계약이다. 현재 비슷한 표면은 둘이다: `git_strategy.<mode>.automation.{auto_push,auto_pr}` 는 읽는 비테스트 Go 코드가 없고(research.md L20) 모드 프로필마다 따로 있다(L18). `workflow.yaml` 은 이미 worktree·review_gate·candidate_ci 같은 opt-in 토글을 싣는다(L22–L24).
Options:
- (a) `workflow.yaml` 안의 새 중첩 매핑 하나에 일곱 단계를 둔다(작업 이름: `workflow.automation.stages.<stage>`). 한 파일, 한 블록이라 한 번의 원자적 rename 으로 저장된다.
- (b) `git_strategy.<mode>.automation` 아래의 기존 키(`auto_push`, `auto_pr`)를 재사용하고 나머지 다섯을 같은 자리에 더한다. 모드 프로필 셋에 각각 일곱 키가 생긴다.
- (c) 새 섹션 파일(`automation.yaml`)을 만든다. 로더와 구조체 대칭 시험, 웹 스키마 섹션이 함께 늘어난다.
Default: (a) (rule: 순위 3, 사용자에게 보이는 표면이 작은 것). 순위 1: 세 선택지 모두 현재 동작을 바꾸지 않는 키 추가이므로 가르지 못한다. 순위 2: 세 선택지 모두 이 SPEC 의 커밋 되돌리기로 원상복구된다. 순위 3: (a) 는 기존 파일에 블록 하나, (b) 는 세 프로필에 일곱 키씩, (c) 는 새 파일과 로더·스키마 섹션이다. 이 기본값은 추천이 아니다.
Alternate: (b).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) `workflow.automation.stages.<stage>`; jev-1.13.0 p=0.94. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q2: 단계 값은 어떤 형태인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 값의 형태는 fail-closed 판정의 폭을 정한다. 이웃 키는 불리언(`enabled: true|false`, 포인터로 부재를 구별, research.md L17)과 닫힌 문자열 집합(`continuation: none|card|pipeline`, L13)을 둘 다 쓴다.
Options:
- (a) 닫힌 문자열 `manual` / `auto`. 불리언이 아닌 값, 오타, 대소문자 차이가 모두 닫힌 쪽(수동)으로 풀린다. 웹이 두 값을 라디오로 보인다.
- (b) 불리언 `enabled: true|false` 형태(단계마다 하위 키). YAML 불리언 표기(`yes`, `on`)의 해석이 파서마다 달라 오타와 구별이 어렵다.
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) 닫힌 문자열 `manual`/`auto`; jev-1.13.0 p=0.98. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q3: review_repair 의 기본값과 범위는 무엇인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 지금 리뷰 지적 수리 요청의 실행 경로는 codex review gate 의 턴 연장 차단 하나이고(research.md L52), 그 게이트는 템플릿에서 꺼져 있다(L23). 이 저장소의 배포본은 켜 놓았다(L53). 카드는 "범용 배포본 기본값은 모두 수동"과 "단계별 독립"을 요구했다. 게이트를 켠 사실을 수리 요청의 동의로 읽을지가 갈린다.
Options:
- (a) `review_repair` 도 다른 여섯과 같이 기본 수동이다. 게이트를 켜도 단계가 `auto` 가 아니면 검토는 돌고 지적은 보이지만 턴을 잇는 차단은 없다. 게이트를 켠 기존 배포는 단계 키를 따로 적어야 현재 동작을 유지한다.
- (b) 단계 키가 없을 때에 한해 `codex.review_gate.enabled: true` 를 수리 요청의 동의로 읽는다. 게이트를 켠 기존 배포의 동작이 그대로이고 웹에는 출처 `derived` 가 나온다. 단계의 독립성이 한 곳에서 깨진다.
Default: (b) (rule: 순위 1, 현재 동작 보존). 게이트를 켠 배포에서 (b) 만 현재의 차단을 유지한다. 이 기본값은 추천이 아니다.
Alternate: (a).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) 기본 수동(`derived` 출처 없음). 이 저장소의 현재 운영은 Q8=(a) 로 project 계층에 `review_repair: auto` 를 적어 유지한다(M6); jev-1.13.0 p=0.94. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q4: release/deploy 단계를 첫 조각에 넣는가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 배포본 Go 코드에는 릴리스·태그·배포를 게시하는 호출이 없다(research.md L51). 단계를 넣으면 키 하나, 웹 행 하나, 정적 검사 하나가 생기지만 바꾸는 동작은 없다. 넣지 않으면 일곱 단계 중 여섯만 웹에 보인다.
Options:
- (a) 넣는다. 키, 해석기, 웹 행, 새 게시 호출을 막는 정적 검사(AC-DSO-021, M7)까지.
- (b) 넣지 않는다. 이 SPEC 은 여섯 단계로 끝내고 release/deploy 는 게시 경로가 생길 때 후속 SPEC 이 정한다. AC-DSO-021, REQ-DSO-017, M7 이 이 SPEC 에서 빠진다.
Default: (b) (rule: 순위 3, 사용자에게 보이는 표면이 작은 것). 순위 1·2 는 두 선택지를 가르지 못한다. 이 기본값은 추천이 아니다.
Alternate: (a).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (b) release/deploy 단계 제외, 이 SPEC 은 여섯 단계(REQ-DSO-017 · AC-DSO-021 · M7 철회, 후속 SPEC); jev-1.13.0 lean 0.69 → Codex 교차 p=0.90 로 확인. 경로: 운영자 상시 규칙(Jev 판정, 불충분 시 Codex 교차; 리더 run tn564n 중계 2026-10-19).

### Q5: 잘못된 값 하나는 어디까지 수동으로 만드는가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "누락·잘못된·부분 설정은 지정되지 않은 단계를 수동으로 해석"이라고 썼다. 지정되지 않은 단계만 수동인지, 잘못된 값이 있으면 블록 전체가 수동인지는 문언만으로 정해지지 않는다. 한 단계의 오타가 다른 단계의 명시 동의를 지우면 사용자는 켜 둔 줄 아는 자동화가 멈춘 것을 모르고, 반대로 일부만 살리면 오타 옆의 동의가 읽힌 상태로 남는다.
Options:
- (a) 잘못된 값이 있는 단계만 수동이고 유효한 형제 단계는 그대로 읽는다. 이유 코드가 단계마다 붙는다.
- (b) 블록에 잘못된 값이 하나라도 있으면 일곱 단계 모두 수동이다.
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) 단계별 fallback(잘못된 단계만 수동, 단계별 이유 코드; 블록 전체 수동은 블록 자체를 읽지 못할 때만); Codex p=0.93(Jev 0.54/0.46 동요 → Codex 교차로 결정). 경로: 운영자 상시 규칙(Jev 판정, 불충분 시 Codex 교차; 리더 run tn564n 중계 2026-10-19).

### Q6: project.continuation 과 todo_creation 은 어떤 관계인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: `continuation` 은 `/moai project` Phase 14 에서 카드를 발급할지와 다음 선택지의 범위를 정하는 표현 키이고, 키가 없거나 틀리면 `card` 로 풀려 발급 쪽으로 열린다(research.md L13, L14). 템플릿은 `continuation: card` 로 나간다(L15). 카드 발급은 단계 `todo_creation` 의 대상이기도 하다. 카드가 말한 "project.continuation 키 누락" 항목의 실측은 키의 누락이 아니라 누락이 열린 쪽으로 풀리는 것이다.
Options:
- (a) 발급 여부를 `todo_creation` 단계가 정하고 `continuation` 은 표현 키로 그대로 둔다. 템플릿의 `continuation: card` 값도 그대로다. 웹에서 `continuation: card` 인데 카드가 발급되지 않는 조합이 보일 수 있고, 영향 설명이 그 사실을 알린다.
- (b) (a) 에 더해 템플릿 기본값을 `card` 에서 `none` 으로 바꾼다. 템플릿 기본값 변경이다.
- (c) `continuation` 의 `card`/`pipeline` 값이 곧 카드 발급 동의라고 보고 `todo_creation` 을 따로 두지 않는다. 일곱 단계가 여섯이 된다.
Default: (a) (rule: 순위 1, 현재 동작 보존). (b) 는 템플릿 기본값을 바꾸고 (c) 는 단계 수를 바꾼다. 이 기본값은 추천이 아니다.
Alternate: (b).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) `todo_creation` 이 발급 여부를 정하고 `continuation` 은 표현 키로 유지; jev-1.13.0 p=0.94. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q7: team 프로필의 `auto_push: true` 와 읽는 코드 없는 automation 키는 어떻게 하는가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 템플릿 team 프로필은 `auto_push: true` 로 나가지만(research.md L18, L19) 그 값을 읽는 비테스트 Go 코드가 없다(L20). 새 단계 정책이 푸시를 정하면 같은 설정 파일에 서로 다른 말을 하는 두 키가 남는다.
Options:
- (a) 건드리지 않는다. 두 키가 단계 판정과 무관하다는 주석만 더한다.
- (b) team 프로필 템플릿 값을 `false` 로 바꾼다. 읽는 코드가 없으므로 동작은 같고 템플릿 기본값이 바뀐다.
- (c) 읽는 코드 없는 `auto_push`·`auto_pr` 키를 구조체와 템플릿에서 제거한다. 기존 사용자 파일의 키는 update 가 보존하지만 구조체 대칭 시험과 웹 스키마 정리가 따른다.
Default: (a) (rule: 순위 1, 현재 동작 보존). 이 기본값은 추천이 아니다.
Alternate: (b).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (c) 읽는 코드 없는 `auto_push`·`auto_pr` 을 구조체와 템플릿에서 제거; 기존 사용자 파일의 키는 update 가 보존(REQ-DSO-014); jev-1.13.0 p=0.95. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q8: 이 저장소의 승인된 로컬 factory 설정은 새 키로 어떻게 옮기는가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 새 해석기는 키가 없으면 수동이다. 이 저장소의 factory 레인은 github-flow 전달 경로(push, PR, auto-merge 요청)와 review gate 를 켠 채 운영 중이다(research.md L03, L53). 단계 키를 적지 않으면 run 이 끝나는 순간 이 저장소의 레인 전달이 멈춘다. 이 저장소의 `.moai/config/sections/workflow.yaml` 은 템플릿과 다른 운영자 자신의 파일이다(L25, L56). 같은 값이 어느 계층에 적히는지가 되돌리기 가능성과 공유 범위를 바꾼다.
Options:
- (a) project 계층(`.moai/config/sections/workflow.yaml`, 추적 파일)에 명시 `auto` 를 적는다. 이 SPEC 의 커밋으로 되돌릴 수 있고 모든 체크아웃이 같은 값을 본다.
- (b) local 계층(`.moai/config/local/`, 무시되는 디렉터리, research.md L49)에 적는다. 머신마다 따로 적어야 하고 커밋에 남지 않는다. 웹은 `local-override` 로 표시한다.
- (c) 적지 않는다. run 이후 레인 전달은 운영자가 단계별로 켜기 전까지 멈춘다.
Default: (a) (rule: 순위 2, 이 SPEC 커밋 하나의 되돌리기). 순위 1 은 (a)(b) 를 가르지 못하고 (c) 만 현재 동작을 바꾼다. 이 기본값은 추천이 아니며 이 행은 운영자 보유 변경이므로 자율 판정 대상이 아니다.
Alternate: (b).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) project 계층에 명시 `auto` — **OPERATOR-HELD**(행 자체 선언) + 리더 재정(2026-10-19): 판정 내용은 유효하되 실제 project 계층 설정 파일 쓰기(`.moai/config/sections/workflow.yaml`)는 운영자 keep-set 이며 별도 운영자 승인을 게이트로 요구한다. jev-1.13.0 p=0.95. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q9: 웹의 즉시 중지는 무엇을 하는가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "즉시 중지"를 요구하면서 "바꿔도 진행 중 카드에 소급 안 함"도 요구했다. 즉시 중지가 새 설정 값을 쓰는 일인지, 설정과 별개의 정지 표식인지가 갈린다. 후자는 값의 출처가 하나 늘어난다.
Options:
- (a) project 계층의 일곱 단계를 한 번의 원자적 쓰기로 `manual` 로 쓴다. local override 가 `auto` 인 단계는 효력이 그대로이고 화면이 그 사실을 알린다.
- (b) 설정 파일을 건드리지 않고 정지 표식(`.moai/state/` 아래 파일)을 둔다. 모든 경계가 표식을 먼저 본다. local override 도 이긴다. 해제 동작과 표식 만료 규칙이 새로 필요하다.
- (c) (a) 와 (b) 를 함께 한다.
Default: (a) (rule: 순위 3, 사용자에게 보이는 표면이 작은 것). 이 기본값은 추천이 아니다.
Alternate: (b).
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (b) 설정 없는 정지 표식(`.moai/state/` 아래, 모든 경계가 확인하며 local override 도 이긴다; release·만료 규칙은 design §5·§7); jev-1.13.0 p=0.85. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q10: 웹 저장 감사 기록은 어디에 어떤 필드로 남기는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 감사 기록은 카드의 요구이고 위치와 필드 구성은 사용자 가시 계약이 아니다. 추적되지 않는 위치가 둘 있다: 배포 템플릿의 `.gitignore` 는 `.moai/logs/*` 와 `.moai/state/` 를 모두 무시하고(research.md L50) 이 저장소도 두 디렉터리를 무시한다(L49).
Options:
- (a) `.moai/logs/stage-policy-audit.jsonl` 추가 전용. 한 줄: 시각, 채널, 호출 주체(운영체제 사용자), 단계별 변경 전후, 결과, 새 revision.
- (b) `.moai/state/stage-policy-audit.jsonl`.
Default: (a) (rule: 순위 3, 로그 용도의 기존 디렉터리라 새 상태 위치가 늘지 않는다). 순위 1·2 는 두 선택지를 가르지 못한다.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-19T06:05:24Z lane-4 manager-spec (plan)

### Q11: t1882 의 결과와 merge 단계는 어떤 관계인가?

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: 없음
Why unresolved: t1882 는 같은 `factory_card_pr.go` 에서 auto-merge 요청을 없애거나 설정 키 뒤로 옮기고 draft 게시를 더한다. 큐 상태는 `picked` 이고 이 트리에는 변경이 없다(research.md L04–L07). 착지한 설계를 보기 전에는 이 SPEC 의 merge 단계가 무엇을 막는지가 정해지지 않는다.
Options:
- (a) t1882 가 auto-merge 요청을 완전히 없앤다. 그 경우 `merge` 단계는 병합을 요청하는 미래의 opt-in 경로를 위한 게이트로 선언만 되고 현재 호출 지점이 없다.
- (b) t1882 가 요청을 자기 키 뒤에 남긴다. 그 경우 이 SPEC 은 그 키를 `merge` 단계로 대체하고 두 번째 키를 만들지 않는다.
- (c) 두 키가 공존한다. 같은 병합 요청에 서로 다른 기본값을 가진 키가 둘이 된다.
Operator verdict:

### Q12: 산문 소비자용 해석기 읽기 동사 이름은 무엇인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: `/moai project` Phase 14 같은 산문 워크플로는 Go 해석기를 부를 수 없으므로 읽기 동사가 필요하다. 이름은 공개 계약이 아니라 스킬 산문과 시험이 쓰는 호출 이름이다. `moai doctor config` 가 이미 있어(doctor_config.go) 진단 계열 동사의 자리가 있으나 부모 동사 연결은 M2 에서 확인한다.
Options:
- (a) `moai doctor stages [--json]`.
- (b) `moai factory stages [--json]`.
Default: (a) (rule: 순위 3, 기존 진단 계열 아래라 최상위 표면이 늘지 않는다).
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-19T06:05:24Z lane-4 manager-spec (plan)

### Q13: policy revision 은 어떤 형태인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: revision 은 웹 저장 응답과 전환 기록에 남는 값이다(AC-DSO-004, 013, 018). 형태는 내용 해시와 단조 증가 카운터가 있다. 사용자에게 의미 있는 약속은 "단계 관련 입력이 바뀌면 바뀐다"뿐이다.
Options:
- (a) 세 계층의 단계 블록을 정규화해 직렬화한 값의 sha256 앞 12 hex. 상태 파일이 필요 없다.
- (b) 상태 파일에 두는 단조 증가 카운터. 쓰기 경합과 파일 손실 처리가 새로 필요하다.
Default: (a) (rule: 순위 3, 새 상태 파일이 없다).
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-19T06:05:24Z lane-4 manager-spec (plan)

### Q14: dispatch 단계가 막는 범위는 어디까지인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "dispatch/execution"이라고만 썼다. 지금 자가 배차의 경로는 둘이고(research.md L45, L46), 리더나 운영자가 카드를 지명해 레인에 넘기는 경로가 따로 있다. 단계가 자가 배차만 막는지, 지명까지 막는지에 따라 `manual` 의 효과가 달라진다.
Options:
- (a) 자가 배차(인자 없는 `moai factory next`, 레인 시작 스탬프, 세션 시작 규칙 선택)만 막고 `--card` 지명은 운영자 행위로 허용한다.
- (b) 지명까지 막는다. `manual` 이면 레인은 어떤 임대도 스스로 하지 않는다.
- (c) 세션 시작 규칙과 스탬프만 바꾸고 `moai factory next` 는 그대로 둔다. 안내문만 바뀌고 강제는 없다.
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (개정 2) — (a) 자가 배차만 차단(CLI `next`·MCP `factory_next`·`--auto-lane` 임대, 스탬프·세션 시작 규칙), `--card` 지명은 운영자 행위로 허용; jev-1.13.0 p=0.96. 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19).

### Q15: local override 가 고정한 단계에 대한 웹 저장은 거부인가 기록인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why recorded: 1회차 plan-audit(FAIL 0.85, `.moai/reports/t1895/plan-audit.md` D1)이 REQ-DSO-016(거부)과 AC-DSO-020(c)(쓰기+알림)이 같은 입력 — local override 가 고정한 단계에 대한 웹 저장 — 에 반대 결과를 요구하는 미표기 갈림을 발견했다. 원래 14행 어디에도 이 갈림을 묻는 행이 없었다.
Options:
- (a) 거부한다. 쓰기 없이 거절 응답으로 끝낸다.
- (b) 쓰고 알린다. project 값만 쓰고 효력은 local 값 그대로이며, 응답과 감사 기록이 `overridden-by-local` 을 알린다.
Verdict: DECIDED (b) — 쓰기+알림. 판정 경로: 운영자 상시 규칙(이 세션은 Jev 로 판정한다)에 따라 리더 run tn564n 이 중계했다(2026-10-19). Jev 모델 `jev-1.13.0` 의 (b) 답변 확률 0.89. 리더는 개정 전에 1회차 판정문(FAIL)을 통보받았다. 근거: 세 계층 모델에서 project 계층의 쓰기는 local override 를 이길 수 없으므로 never-overwrite 규칙이 보존되고, 거부는 override 해제 뒤 적용될 운영자 의사를 기록하지 못한다. design.md §7 과 AC-DSO-019(즉시 중지의 쓰기+알림)와 같은 규칙이 된다.
Operator verdict: DECIDED 2026-10-19T12:47:02Z lane-52 manager-spec (plan-audit round-1 revision) — (b) 쓰기+알림; 경로: 운영자 상시 규칙(Jev 판정, 리더 run tn564n 중계 2026-10-19); jev-1.13.0 p=0.89

---

## 계획 닫음 처리

plan 닫음 시점(2026-10-19T06:05:24Z)에 `implementation-level` 분류이면서 `Default:` 가 있고 판정이 비어 있던 Q10, Q12, Q13 에 `DEFAULT-APPLIED` 를 기록했다. 제품 수준 행(Q1–Q9, Q14)과 EVIDENCE-NEEDED 행(Q11)은 비어 있으며 `DEFAULT-APPLIED` 를 쓰지 않았다.

2026-10-19T12:47:02Z 개정 2(spec v0.2.0): Q1–Q9, Q14 의 판정을 운영자 상시 규칙(이 세션은 Jev 로 판정하고 불충분하면 Codex 로 교차, 리더 run tn564n 이 중계, 2026-10-19)으로 기록했다. 각 행의 `Operator verdict` 가 모델·선택·확률을 운반한다. Q4=(b) 에 따라 REQ-DSO-017 · AC-DSO-021 · 마일스톤 M7 을 이 SPEC 에서 빼는 재범위를 실행했고 Q9=(b) 에 따라 즉시 중지를 정지 표식 설계로 바꿨다. Q11 은 열려 있다. Q15 는 개정 1에서 더해진 행이다(감사 발견 갈림, DECIDED).
