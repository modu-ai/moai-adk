# 다중 에이전트 격리와 소스 관리: 공개 사례와 팩토리 방어 장치 대조

## 1. 이 문서의 성격

이 문서는 `reference-only` 성격의 비교 자료다. 다중 에이전트 시스템을 다룬 공개 글 네 편의 주장과, 이 저장소에 이미 들어 있는 팩토리의 방어 장치를 한 표에 나란히 놓았다. 어떤 규칙, 설정, 코드, 템플릿도 바꾸자는 권고는 담지 않았고, 외부 방식을 팩토리가 따라야 한다는 주장도 하지 않는다. 표에서 `gap`으로 표시한 줄을 메울지는 별도의 결정이며 이 문서의 범위 밖이다.

로컬 개발 자료이므로 사용자 프로젝트로 배포되는 템플릿에는 복제하지 않는다.

## 2. 출처 목록

네 출처의 제목, 주소, 인용 조각, 확인 결과를 아래에 모았다. 인용 조각은 원문의 영어 문장을 그대로 옮긴 것이다.

| ID | Title | URL | Quoted fragment | Flag |
|---|---|---|---|---|
| BP-1 | Scaling Managed Agents: Decoupling the brain from the hands | https://www.anthropic.com/engineering/managed-agents | "each could fail or be replaced independently" | VERIFIED |
| BP-2 | Prompting Claude Opus 5.5 | https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5-5 | "Treat a text-only end of turn as a report rather than as proof the task is done" | VERIFIED |
| BP-5 | Redesigning source control at Anthropic | https://www.anthropic.com/engineering/redesigning-source-control | "post-merge and pre-deploy automated testing" | VERIFIED |
| BP-6 | How we contain Claude across products | https://www.anthropic.com/engineering/how-we-contain-claude | "Design for containment at the environment layer first, then steer behavior at the model layer" | VERIFIED |

`VERIFIED`는 이 문서를 쓰는 실행에서 해당 주소를 직접 내려받아, 인용 조각이 내려받은 원문에 그대로 들어 있음을 확인했다는 뜻이다. 주소를 가져오지 못했거나 조각을 원문에서 찾지 못하면 `UNVERIFIED`로 둔다. `UNVERIFIED` 출처를 인용한 줄은 확정된 사실이 아니라 확인을 기다리는 단서로 읽는다. 지금은 네 출처 모두 `VERIFIED`다.

## 3. 읽는 법

대조표의 열은 Theme, Source, Source claim, MoAI defense, Fit 순서다. Theme은 세 가지다. `correlated-failure`는 한 구성 요소의 실패가 다른 구성 요소로 번지는 문제, `isolation-first`는 경계를 프롬프트가 아니라 실행 환경에 두는 접근, `merge-queue-alternative`는 에이전트가 대량으로 병합하는 환경에서 병합 대기열을 대신하는 방식을 가리킨다.

Fit 칸에는 세 값 가운데 하나가 들어간다.

- `aligned`: 열어 본 본문이 출처의 주장과 같은 방향으로 맞닿는다.
- `differs`: 같은 문제를 다루지만 방식이나 기본 자세가 다르다.
- `gap`: 열어 본 파일에서 출처의 주장에 대응하는 방어 장치를 찾지 못했다. 찾지 못했다는 보고일 뿐 장치가 없다는 단정이 아니다.

MoAI defense 칸은 저장소 파일과 절 이름을 `경로 § 절` 꼴로 가리키고, 그 뒤는 본문을 풀어 쓴 요약이다. 규칙 본문을 옮겨 적지 않았다. 절 이름은 원문 그대로의 영어이며, 규칙 표지가 붙은 절은 표지를 뺀 이름만 적었다.

## 4. 규모에 관한 주의

`merge-queue-alternative` 줄은 규모가 다른 두 방식을 견주는 것이다. 출처 BP-5는 하루에 병합되는 변경이 약 2,500건에서 30,000건 넘게 늘어난 환경을 이야기한다. 그 환경에서는 병합 대기열에 전역 저장소 잠금을 두는 업계 표준 방식을 효과적으로 복제하지 못했고, 대신 병합 후와 배포 전 자동 테스트에 투자했다고 보고한다. 이 물량과 결론은 모두 출처가 밝힌 내용이며 이 문서가 측정한 값이 아니다.

팩토리의 병합 물량이나 병목을 잰 기록은 열어 본 파일 어디에도 없었다. 그래서 이 문서는 팩토리의 규모를 두고 어떤 주장도 하지 않는다. 병합 후 테스트를 팩토리가 채택해야 한다는 뜻으로 읽어서도 안 된다.

## 5. 대조표

| Theme | Source | Source claim | MoAI defense | Fit |
|---|---|---|---|---|
| correlated-failure | BP-1 | 출처는 세션 기록을 하네스 바깥에 두므로 하네스가 비정상 종료돼도 새 하네스가 기록을 읽어 마지막 사건부터 이어 갈 수 있다고 설명한다. "nothing in the harness needs to survive a crash" | `.claude/rules/moai/workflow/factory-dispatch.md` § The delegation channel is the queue; § Completion is read, never trusted; `.claude/rules/moai/workflow/auto-semantics.md` § 5.1 The standing awaken carrier. 위임은 메시지가 아니라 디스크의 대기열로 이루어지고, 카드를 넘기는 근거도 디스크에서 읽은 진행 기록과 검증 증거다. 멈춘 레인은 스스로 깨어나지 못하므로 카드를 받을 때 걸어 둔 반복 호출이 깨우고, 깨어난 레인은 메시지보다 디스크를 먼저 읽는다. 범위는 실행 세션 밖에 상태를 둔다는 점까지다. 하네스를 상태 없는 존재로 만든다는 주장에 대응시키지는 않았다. | aligned |
| correlated-failure | BP-1 | 출처는 두뇌(모델과 하네스), 손(샌드박스와 도구), 세션(사건 기록)을 서로 가정이 거의 없는 인터페이스로 갈라 각각 따로 실패하고 따로 교체되게 했다고 설명한다. "each could fail or be replaced independently" | `.claude/rules/moai/workflow/factory-dispatch-gates.md` § Verification load incident record; `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Serializing a heavy run across lanes. 레인마다 작업 트리는 따로 쓰지만 같은 기계를 공유한다. 기록된 사고에서는 여러 레인이 같은 전체 테스트를 동시에 돌려 기계 부하가 크게 치솟았다. 방어는 실패 영역을 나누는 쪽이 아니라 무거운 실행 앞에 자원 임대를 두어 한 번에 하나만 돌리게 하는 쪽이다. 임대를 강제하는 사전 차단은 선택 사항이고 규칙이 말하는 기본 상태는 꺼짐이다. | differs |
| correlated-failure | BP-2 | 출처는 진행 상황을 알리느라 도구 호출 없이 글만 쓰고 끝난 턴을 작업이 끝났다는 증거가 아니라 보고로 다루라고 권한다. "Treat a text-only end of turn as a report rather than as proof the task is done" | `.claude/rules/moai/workflow/cross-session-messaging-core.md` § An idle notice is a scheduling hint; `.claude/rules/moai/workflow/factory-dispatch.md` § Completion is read, never trusted. 세션이 유휴가 되었다는 알림은 일을 마쳤을 때도, 허가를 기다리다 멈췄을 때도, 죽었을 때도 똑같이 오므로 언제 가서 볼지만 알려 줄 뿐 완료의 증거가 아니다. 카드는 직접 읽은 증거로만 다음 단계로 간다. 층위가 다르다. 출처는 모델 한 턴의 끝을 말하고 이 줄은 세션 유휴 알림과 레인의 답신을 다룬다. 열린 항목을 점검표로 돌려보내는 출처의 절차까지 대응시키지는 않았다. | aligned |
| correlated-failure | BP-2 | 출처는 열린 항목이 남은 채 턴이 끝나 자동으로 이어 가게 하더라도, 같은 작업에서는 두세 번 이어 간 뒤 멈추라고 권한다. 정말로 막힌 실행이 끝나고 검토를 받게 하려는 취지다. "stop after two or three automatic continuations on the same task" | `.claude/rules/moai/workflow/goal-directive-detail.md` § What It Is. 자율 이어 가기 목표에는 턴 상한이 있고, 상한에 닿으면 판정 보고서를 남기며 더는 막지 않는다. 런타임에는 연속 차단 상한이 따로 있어 무인 실행에서는 둘 중 낮은 쪽이 먼저 걸리고, 진전 없는 반복이 이어지면 정체 감지가 루프를 멈춘다. 막힌 실행을 끝내 검토할 수 있게 한다는 목적은 같지만 한도의 크기는 출처가 말하는 두세 번보다 크다. | differs |
| isolation-first | BP-6 | 출처는 요약에서 격리를 환경 계층에서 먼저 설계하고 그다음에 모델 계층에서 행동을 유도하라고 정리한다. 확률적 방어가 모두 빗나갈 때 마지막에 걸리는 것은 결정적 경계라는 이유다. "Design for containment at the environment layer first, then steer behavior at the model layer" | `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Rules; § Mechanical Enforcement. 공유 체크아웃에서 분기 상태를 바꾸는 명령을 금지하는 원칙은 항상 로드되는 문서 계층에 있다. 같은 원칙을 명령 단계에서 막는 기계적 계층은 사전 도구 훅으로 따로 있는데, 규칙이 말하는 배포 기본값은 꺼짐이고, 확신이 없으면 통과시키며, 스폰된 에이전트가 신원 예외로 지나갈 수 있는 우회로가 있다는 점도 본문이 인정한다. 방향은 같지만 기본 자세가 다르다. 이 저장소가 실제로 켠 설정값은 열어 보지 않았으므로 그 값을 근거로 한 주장은 하지 않는다. | differs |
| isolation-first | BP-6 | 출처는 환경 계층의 방어를 프로세스 샌드박스, VM, 파일 시스템 경계, 외부 송신 통제로 설명하고, 그 목적을 에이전트가 닿을 수 있는 범위에 단단한 경계를 긋는 것이라고 말한다. "The goal is to set a hard boundary on what an agent can reach" | `.claude/rules/moai/workflow/worktree-integration.md` § Worktree Selection Rules; § Parallel-Session Branch Conflict Auto-Isolation; `.claude/rules/moai/workflow/factory-dispatch.md` § Isolation is provisioned by MoAI, then entered through a launcher. 파일을 쓰는 에이전트는 격리된 작업 트리에서 돌리고, 읽기 전용 에이전트는 격리 대신 도구 목록으로 쓰기 경로를 막는다. 카드 한 장은 실행기로 들어간 자기 작업 트리에서 시작해 거기에 머물며, 같은 체크아웃에 다른 세션이 있으면 작업을 별도 트리로 떼어 낸다. 분리되는 것은 작업 트리와 분기 상태다. 읽은 본문은 이를 쓰기 충돌과 분기 상태 간섭을 막는 수단으로 설명하며, 파일 시스템이나 네트워크, 자격 증명이 닿는 범위를 막는 경계로는 읽지 않았다. | differs |
| isolation-first | BP-6 | 출처는 자체 계측에서 사용자가 권한 확인 요청의 약 93%를 승인했다고 밝히고, 승인이 쌓일수록 하나하나에 쏟는 주의가 줄어 사람의 감독이 약해진다고 해석한다. "users approved roughly 93% of permission prompts" | `.claude/rules/moai/workflow/auto-semantics.md` § 9. The gate inventory and dispositions. 사람이 답하던 관문 대부분은 독립 감사 증거와 기록된 결정으로 자율 처리되고, 사람이 남는 경우는 환경상 불가능한 작업, 운영자가 쥔 작업, 외부 공유 시스템에 닿는 되돌릴 수 없는 작업의 세 부류와 자발적 계약 서명 경로로 한정된다. 줄이는 방향은 맞닿는다. 다만 출처는 촘촘한 경계를 세운 뒤에 감독을 줄이라고 하는데, 여기서 줄이는 근거는 감사 증거이고 경계의 강도는 위 두 줄이 따로 다룬다. 승인율은 출처의 계측값이며 이 저장소의 사실로 옮기지 않는다. | aligned |
| isolation-first | BP-1 | 출처는 자격 증명이 Claude가 만든 코드를 실행하는 샌드박스에서 아예 닿지 않게 하는 것을 구조적 해법으로 든다. 인증을 자원에 묶거나 샌드박스 밖 보관소에 둔다. "the tokens are never reachable from the sandbox where Claude" | 찾지 못함. `.claude/rules/moai/core/moai-constitution.md` § Security Boundaries는 비밀을 버전 관리에 올리지 말고 자격 증명은 환경 변수로 쓰라는 방침 수준이다. `.claude/rules/moai/workflow/ci-autofix-protocol.md` § Secrets and Credentials Protection은 자동 수정 루프가 건드리면 안 되는 파일을 정하고, `.claude/rules/moai/core/glm-web-tooling.md` § Gateway credential and telemetry scope는 게이트웨이 자격 증명이 어디로 전송되는지를 다룬다. 셋 다 자격 증명을 에이전트의 실행 환경에서 구조적으로 떼어 놓는 장치를 설명하지 않는다. 장치가 없다는 증거가 아니라 열어 본 파일에서 찾지 못했다는 뜻이다. | gap |
| merge-queue-alternative | BP-5 | 출처는 하루 병합이 수천 건에서 수만 건으로 늘어난 환경에서 병합 대기열의 전역 저장소 잠금을 효과적으로 복제하지 못해 대신 병합 후와 배포 전 자동 테스트에 투자했다고 보고한다. 물량과 결론은 출처의 보고다. "post-merge and pre-deploy automated testing" | `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Integration into the release branch is self-served (level-2 heading); `.claude/rules/local/gitflow-lane-protocol.md` § 3. Serialization — one lane at a time in the merge window; § 4. Push — the leader does it in bulk; lanes never push. 병합 창은 한 번에 한 레인만 쓰고, 대기열은 도착 순서대로 차례를 넘긴다. 비용이 큰 재측정은 창에 들어가기 전에 후보 트리를 기준으로 끝내 두며, 창 안의 단계는 짧은 동일성 확인과 병합뿐이고 테스트 묶음은 돌리지 않는다. 합친 뒤에는 통합 브랜치의 원격 CI가 최종 판정을 내리고, 직전 묶음이 빨간 동안에는 다음 묶음을 보류한다. 병합 전에 직렬화와 재측정을 두는 점이 출처가 보고한 병합 후 테스트 방식과 다르다. 다만 병합 뒤의 CI 판정도 함께 있어 대비는 부분적이다. 규모가 다른 두 방식의 비교이며, 팩토리의 병합 물량을 잰 값은 열어 보지 않았다. | differs |
| merge-queue-alternative | BP-5 | 출처는 동시에 들어온 변경이 서로 충돌할 수 있음을 인정하면서, 그 충돌은 사람이 알아차리기도 전에 Claude가 고치는 경우가 보통이라고 말한다. 속도는 출처의 관찰이다. "Claude can typically fix these conflicts faster than humans" | `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Integration into the release branch is self-served (level-2 heading; conflicts bullet). 병합을 수행하는 레인이 자기 병합이 일으킨 충돌을 직접 푼다. 다른 레인이 이미 병합한 변경과 의미상 부딪혀 풀 수 없으면 강제로 병합하지 않고 리더에게 차단 보고를 올린다. 속도에 관한 주장은 이 본문에 없어 대응시키지 않았고, 대응하는 것은 에이전트가 충돌을 푼다는 점까지다. | aligned |
| merge-queue-alternative | BP-5 | 출처는 하루 병합 수가 열 배 넘게 늘면서 소스 관리가 개발의 가장 큰 병목이 되었다고 말한다. "source control became our biggest development bottleneck" | 찾지 못함. 팩토리의 병합 물량이나 병목을 잰 기록을 열어 본 파일에서 찾지 못했다. 그래서 이 줄에서는 팩토리의 규모를 주장하지 않는다. | gap |

## 6. 이 대조표가 말하지 못하는 것

- `gap` 두 줄은 열어 본 파일에서 찾지 못했다는 뜻이다. 자격 증명 줄은 관련 규칙 문서 몇 건을 열어 보고도 구조적 분리를 설명하는 대목이 없어서, 병합 물량 줄은 측정 기록이 없어서 그렇게 표시했다.
- 분기 상태 보호 줄은 규칙이 말하는 배포 기본값만 근거로 삼았다. 이 저장소가 실제로 켠 설정값은 열어 보지 않았다.
- 방어 장치 칸은 규칙 본문을 요약한 것이라 요약이 본문보다 세게 또는 약하게 읽힐 수 있다. 판단이 갈리면 가리킨 절을 직접 읽는다.
- 출처 페이지와 규칙 문서는 이후에 바뀔 수 있다. 위 확인은 이 문서를 쓴 시점의 내용에 대한 것이다.
