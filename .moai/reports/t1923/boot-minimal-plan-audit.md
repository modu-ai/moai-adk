auditor-model: gpt-6
Verdict: fail

- [P1] 리더 공지의 배타적 허용 목록이 완료된 배치 게이트 계약과 충돌합니다 — [spec.md:72](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/spec.md:72)
  - **D1 · blocking · 확신 높음.** `SPEC-AUTONOMY-BATCH-GATE-001` REQ-BGS-016/017은 리더 공지에 `batch gate summary`와 정본 경로·절을 담은 문장을 요구합니다. 새 목록에는 없으며, REQ-FBM-003은 이를 디스크로 이전하도록 명시합니다. 최소 공지 반례에서 기존 검사 조건은 `omit-pointer`, `omit-name-token`을 검출했습니다.
  - **수정:** 해당 문장을 허용 목록에 보존하거나, 기존 계약을 대체하는 범위와 근거를 명시하고 연결 테스트까지 함께 개정하십시오.

- [P1] t85 디스패치 규율 검사의 계약 변경이 계획에서 빠졌습니다 — [plan.md:116](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/plan.md:116)
  - **D2 · blocking · 확신 높음.** 새 목록에 맞춘 최소 공지는 `TestFactoryLeadNoticeCarriesDispatchDiscipline`의 필수 표지 **8/8개**, 한국어 검사의 **4/4개**를 잃습니다. 두 검사는 `session_start_factory_test.go:237,271`에 있으나 변경 대상 목록과 구체적인 계약 개정 계획에 없습니다.
  - **수정:** 두 검사를 명시적으로 포함하고, 공지에서 제거하는 규율이 이전 문서에 보존되었음을 검사하도록 전환하십시오.

- [P1] AC-FBM-001은 허용 목록 밖 문장과 필수 요소 누락을 모두 통과시킵니다 — [acceptance.md:29](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/acceptance.md:29)
  - **D3 · blocking · 확신 높음.** 탐침의 문자열·크기 판정 조건을 메모리에서 평가했습니다. 정상 요소 뒤에 `UNLISTED-PROSE`를 덧붙인 공지는 **fixed=142, composite=6399**로 통과했습니다. `f-probe1 factory-leader https://adk.mo.ai.kr`만 남겨 진입 명령·소켓·이전 문서 포인터를 누락한 공지도 **fixed=44, composite=6301**로 통과했습니다.
  - **수정:** 필수 요소 전체와 허용되지 않은 요소의 부재를 함께 검사하고, `systemMessage` 채널도 포함하십시오.

- [P1] AC-FBM-003은 재배치 없이 삭제한 경우를 통과시킵니다 — [acceptance.md:41](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/acceptance.md:41)
  - **D4 · blocking · 확신 높음.** 관련 문자열을 삭제하고 이전 문서를 전혀 작성하지 않은 메모리 변이는 `grep_count=0`, `grep_exit=1`로 명시된 Then을 만족합니다. 필드명 검색은 REQ-FBM-003의 **모든 제거 요소 보존**, **제거 전 이전**, **포인터 연결**을 관측하지 않습니다.
  - **수정:** 제거 요소별 목적지·보존 내용·포인터를 검사하고, M2 완료 증거를 M3 진입 조건으로 고정하십시오.

- [P1] AC-FBM-002의 성공 조건은 상한 위반 시 실패하는 검사를 보장하지 않습니다 — [acceptance.md:34](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/acceptance.md:34)
  - **D5 · blocking · 확신 높음.** 두 테스트가 칸별 수치와 `swept=72`만 출력하고 실패 단언을 하지 않는 반례는, **fixed=2848, composite=9147**을 출력하면서도 두 `PASS`, SKIP 없음, exit 0이라는 Then을 만족합니다. 문서의 출력 조건을 메모리에서 구성·확인한 반례이며, Go 변이 테스트를 실행했다는 뜻은 아닙니다.
  - **수정:** 고정분·합본 각각의 초과 변이를 실제 검사 경로에 넣어 실패 종료와 원인별 진단을 요구하십시오.

- [P1] E2는 실행된 테스트가 모두 녹색인 기록을 RED-now로 제출합니다 — [acceptance.md:90](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/acceptance.md:90)
  - **D7 · blocking · severity=critical · 확신 높음.** 기록된 원문은 **FAIL 0건, PASS 10건, 마지막 판정 PASS, exit 0**입니다. 이는 기존 검사의 부족함을 보여 주지만, 새 수용 조건을 실행해 실패시킨 기록은 아닙니다. release-blocking AC-FBM-002의 재현 RED가 없습니다.
  - **수정:** 현재 트리에서 실패하는 새 검증기를 작성·실행하고 명령·원문 출력·종료 코드·트리 SHA를 기록하십시오. E1은 별도로 72칸과 FAIL 25줄이 확인되어 이 결함에 해당하지 않습니다.

- [P2] AC-FBM-004·005의 기계 검사는 의무 문단 삭제를 허용합니다 — [acceptance.md:53](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/acceptance.md:53)
  - **D6 · optional · 확신 높음.** 실제 `AGENTS.md` 바이트에서 main 보호 문단 전체를 제거한 메모리 변이는 **23,888바이트**, `live only under` **1건**으로 두 조건을 만족합니다. main 변형 브랜치 금지·착지 시 재동기화·primary HEAD detach 금지 의무는 사라집니다. 계획의 수동 의미 대조는 남아 있지만 이 검사들로 무손실을 증명할 수는 없습니다.
  - **권고:** §B.5 의미 대조 결과를 별도 수용 조건과 증거로 요구하십시오.

- [P2] 생산자 예산 회계에서 추가 결합자 2자가 빠졌습니다 — [spec.md:57](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/spec.md:57)
  - **D8 · blocking · 확신 높음.** 표는 `1950+2847=4797`로 계산하지만 탐침은 allowance와 fixed 사이에 `"\n\n"`을 추가합니다. 따라서 생산자는 **4799**이며, 선언된 최대 core를 쓰면 `1950+2+2847+257+4092=9148`입니다. 현재 관측 core 4048에서는 **9104**로 9146·10000 안에 있으므로 현재 넘침으로 판정한 것은 아닙니다.
  - **수정:** 내부 결합자를 회계에 포함하고 생산자·고정분 상한을 일관되게 재도출하십시오.

- [P2] 라이브 스캐너가 넘침 세션을 누락하고 녹색 결과를 출력합니다 — [probe/live_scan.py:39](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py:39)
  - **D9 · blocking · 확신 높음.** 실제 스캐너를 메모리 파일 시스템으로 실행했습니다. 정상 레인 6개·리더 1개 외에, 17번째 줄에 10000자를 넘는 SessionStart를 담은 파일을 넣자 이를 누락하고 **`count 7` / `overflow 0`**을 출력했습니다. AC-FBM-008의 행 수·크기·넘침 조건이 모두 만족됩니다.
  - **수정:** 고정 줄 수 제한을 제거하고, 대상 run·세션 명단·설치 빌드를 연결하며 미관측 대상은 실패로 처리하십시오.

- [P2] 기존 I18N 계약을 무조건적인 두 모드 설명 의무로 잘못 서술했습니다 — [decision-index.md:5](/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1923/.moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/decision-index.md:5)
  - **D10 · optional · 확신 높음.** 실제 REQ-004는 **“Where 안내 문구가 … 2모드 분리를 설명하면”**이라는 조건부 계약입니다. 두 모드 설명을 제거하는 행위 자체를 이 조항 위반으로 볼 수 없습니다. 현재 테스트가 설명의 존재를 요구한다는 사실과 SPEC의 조건부 요구를 구분해야 합니다.
  - **권고:** Q1의 전제를 바로잡고, 유지할 로케일·spawn authority·폴백·문서 동기화 의무와 대체할 테스트 계약을 구분하십시오.

## 감사 기록

**Claim:** 계획 승인 불가입니다. MP-8 실패와 독립적인 계약·검증 결함이 확인되었습니다. 최초 감사이며, 반환 보고서의 목적지는 `.moai/reports/t1923/plan-audit.md`입니다.

**Evidence:** `git rev-parse HEAD`, `git rev-parse HEAD^{tree}`, `git status --porcelain=v1`, 두 diff 검사를 실행했습니다. 시작·종료 HEAD가 같고 status·diff 출력은 비었습니다. 메모리 변이 평가는 위 수치를 출력했으며, 실제 스캐너 반례의 출력은 `count 7`, `overflow 0`이었습니다. 추적 검사 출력은 `COLLECTED: 11 REQ definitions (acceptance input: read)`이고 고아·미연결 ID는 없었습니다.

**Baseline-attribution:** HEAD `b805ac2d2429f414a7b3634d4196711d688289be`, Git tree `6184ce150323dcddb60236549984b4bac9b1ab7f`. 요청 대상 내용 해시는 시작·종료 모두 `b1d4f6b62c4805418c27068b6acd411617c633a7a987724f7e27f0298d2e0252`였습니다.

**Gaps:** Go 테스트와 RED-now 명령은 재실행하지 않았습니다. 공지 변이는 탐침 조건 평가이며 실제 Go 렌더러 실행이 아닙니다. 라이브 설치 검증과 외부 감사 receipt 발급은 수행하지 않았습니다. 파일을 수정하지 않았으며, 보고서 파일은 런처가 저장할 반환 본문입니다.

**Residual-risk:** 수정 후 계약 간 정합성과 실패 변이 검사를 다시 확인해야 합니다.

MP-1 PASS(`spec.md:72–82` 연속 11개), MP-2 PASS(요구사항 계층의 shall 형식·허용된 legacy 부정형), MP-3 PASS(`spec.md:2–13` 필수 12필드), MP-4 N/A(Go hook 범위), MP-5 PASS(참조 SPEC 모두 completed 또는 자기 자신 draft), MP-6 PASS(`syscall` 0건), MP-7 PASS(plan의 미해결 marker 0건; research 부재는 Tier M), **MP-8 FAIL(E2)**, MP-9 PASS(`COLLECTED: 6 milestones … 0 ordering candidates`; M2 이전 후 M3 제거 순서도 일치).

점수: 명확성 0.75, 완전성 1.00, 검증 가능성 0.75, 추적성 0.75. 종합 0.8125이며 must-pass 실패가 우선합니다.

```text
verdict: FAIL
audited_sha: b805ac2d2429f414a7b3634d4196711d688289be
overall_score: 0.8125
must_pass_failed: 1
blocking_count: 8
plan_artifact_hash: b154a7b15d4b2d4b99d8a3b3b56777b63769a95d87a62bab0bff0daf49f49207
scope: full
fix_scope: spec.md#REQ-FBM-001, spec.md#§B.1, plan.md#M1, acceptance.md#AC-FBM-001, acceptance.md#AC-FBM-002, acceptance.md#AC-FBM-003, acceptance.md#E2, probe/live_scan.py#main
defect_class: D1=design, D2=traceability, D3=ac-wording, D4=traceability, D5=ac-wording, D7=other, D8=design, D9=other
convergence_overall: fail
required_backend: codex fail
```

AUDIT-VERDICT: FAIL spec=SPEC-FACTORY-BOOT-MINIMAL-001 receipts=none