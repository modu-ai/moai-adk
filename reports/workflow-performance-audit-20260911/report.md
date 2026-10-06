# MoAI project · plan · run · sync 심층 감사

기준: 2026-09-11 · main · 2213871afb7d655f411c46a34fd38bb8153278fc · session 01a08e64-c9d3-76f3-a589-5d5d893e1b62

## Claim

규칙 84개(15,288행·1,131,068바이트), retained agent 11개, workflow 23개를 포함한 157파일을 기계 조사했다. 검토 항목 40개: High 18, Medium 22; 그중 격리 동작 재현 9개. 이는 성능 저하 40건의 실측 확정이나 모든 런타임 분기의 검증을 뜻하지 않는다.

## Baseline-attribution

현재 작업 트리의 파일을 기준으로 하며 dirty 상태를 보존했다. 최초와 최종 HEAD 동일, 157개 파일 해시 재검사 변경 0개. 소스·설정 수정, 실제 4단계 실행, commit/push는 하지 않았다. 설치된 moai v3.2.0-rc.7은 소스 HEAD와 빌드 기준이 다르므로 CLI 결과를 이 소스 동작으로 일반화하지 않았다.

## 단계 및 에이전트 관계

|단계|호출 흐름|산출물|주요 항목|
|---|---|---|---|
|project|모드 판별 → 기존 코드 분석/신규 인터뷰 → docs → plan 감사 → 승인 → 선택적 harness|product.md · structure.md · tech.md · harness-spec|F14 F15 F26|
|plan|문맥/명확성 → 연구 → 계획 → Tier → SPEC → 감사 반복 → 승인/전달|Tier S: spec+plan; M/L: acceptance와 설계/연구 조건부|F01 F12 F13 F16 F17 F19|
|run|plan gate → 문맥/도구 → 재계획/승인 → 작업 분해 → TDD/DDD → 검사/독립 판정 → 전달|코드 · 테스트 · progress Run 증거|F02 F03 F20 F23 F25 F37|
|sync|preflight → 품질/보안/MX/coverage → 차이 분석 → 문서 → 최종 검사 → Git 전달|README · CHANGELOG · Sync 증거 · 종결 상태|F04 F05 F06 F07 F10 F11 F21 F30 F32|

|agent|사용 단계|입력 → 산출물|권한 경계|model / effort|정적 skills|
|---|---|---|---|---|---|
|builder-harness|project 선택|프로젝트 분석·harness-spec → agent/skill/plugin의 메타 구조|본문·업무 코드 작성 제외; docs와 구조/내용 분리|inherit / medium|moai-foundation-cc|
|e2e-tester|run/sync UI 조건부|사용자 경로·앱 실행 정보 → e2e 스크립트·실행·화면 증거|TDD 단위 테스트 작성과 분리|inherit / low|moai-workflow-testing|
|manager-design|plan UI 조건부|UI SPEC·디자인 시스템 → D1~D5 화면·디자인 전달|구현 코드는 develop; SPEC 본문은 spec|inherit / high|moai-domain-frontend|
|manager-develop|run 핵심 / sync 수정 시|승인 SPEC·cycle_type·파일 범위 → 코드·테스트·progress Run 증거|draft → in-progress; 보안/배포 자문과 분리|inherit / medium|moai-foundation-core|
|manager-docs|project / sync 핵심|project 정보 또는 승인된 구현 차이 → project 문서·README·CHANGELOG·Sync 증거|SPEC 본문 금지; sync 단일 commit에서 상태 종결|inherit / low|moai-foundation-core|
|manager-git|plan/run/sync 전달|현재 branch·HEAD·증거·승인 → push·PR·통합 결과|모든 Tier PR 위임 지침; main-direct와 충돌|sonnet / low|moai-foundation-core|
|manager-lead|Tier L / kanban / factory 조건부|milestones·소유 경로·dispatch → 분할 위임·peer 검증·통합 조율|유일하게 Agent 도구 보유; leaf의 재위임 금지|inherit / high|moai-foundation-core, moai-workflow-project|
|manager-spec|plan 핵심 / run 계획 / sync 되돌림|요구·연구·Tier·승인 → spec/plan/acceptance/design/research 본문|초기 draft; docs로 본문 소유권 이전 금지|inherit / medium|moai-foundation-core, moai-workflow-spec|
|plan-auditor|project/plan 및 run 진입|구현 전 문서·Tier·iteration → 독립 PASS/FAIL·review-N·근거|본문 작성자와 분리; project 입력 예외 미연결|inherit / high|없음|
|super-advisor|모든 단계 선택적|좁은 의사결정 질문·근거 → 비구속 자문·대안|판정·구현·SPEC 작성 권한 없음|inherit / high|moai-foundation-core|
|sync-auditor|run 완료 / sync fallback|구현·AC·검증 증거 → 독립 PASS/FAIL·부족 증거|사전 SPEC 감사와 충돌; clean sync는 생략 규정|inherit / high|moai-foundation-quality|

Explore와 general-purpose는 별도 retained 파일이 아닌 runtime agent 타입이다. 4dim은 Explore judge 4개를 생성한다. 보안·성능·devops는 general-purpose 역할 주입이다. manager-lead만 Agent 도구를 갖는다. 이 보고서 제작에는 실제 하위 에이전트를 호출하지 않았다.

## 세부 검토

### F01 · High · 감사 캐시는 sticky라지만 조회는 오늘 날짜 파일을 지시한다

분류: 문서 충돌 / plan → run

run 진입 지침은 시간과 무관한 해시 캐시를 선언하면서 오늘 날짜의 보고서를 읽는다. plan-auditor는 최종 review-N 스트림을 재사용 입력으로, 날짜 파일을 기록 전용으로 규정한다. 어제 PASS와 오늘 첫 run을 연결하는 소비 경로가 문서상 일치하지 않는다.

조치·합격 조건: 최종 review-N 보고서의 해시·점수·버전을 단일 조회 함수로 해결하고 날짜 보고서는 이력으로만 사용한다. 날짜가 바뀌고 파일 내용이 같은 fixture에서 새 감사 호출 0회를 확인한다.

근거: .claude/skills/moai/workflows/run/phase-execution.md:45; .claude/agents/moai/plan-auditor.md:369. 원문·명령·stdout은 evidence.json의 동일 ID.

### F02 · High · 추적되지 않은 파일의 내용 변경이 검증 키를 무효화하지 않는다

분류: 동작 재현 / 공통 검증 캐시

현재 Key는 HEAD, porcelain-v2, tracked diff만 결합한다. 같은 untracked 경로의 내용을 A에서 B로 바꾼 fixture에서 키가 동일했다. 신규 파일 구현·테스트를 staging 전에 수정하는 경로에서는 오래된 검사 결과가 fresh로 해석될 수 있다.

조치·합격 조건: 검사 대상 untracked 파일의 내용 해시를 포함하고 생성 산출물 제외 규칙을 명시한다. 파일 내용·이름·삭제·ignored 산출물·tracked 변경 각각의 키 불변/변경 시험을 둔다. TTL 단축으로 이 결함을 대체하지 않는다.

근거: internal/verify/key.go:39; internal/verify/key.go:43 / P07. 원문·명령·stdout은 evidence.json의 동일 ID.

### F03 · High · 공백 정규화가 실행 의미가 다른 문서까지 같은 키로 만든다

분류: 동작 재현 / plan 감사 캐시

normalizeWhitespace는 모든 공백을 합친다. YAML block의 두 줄 명령과 한 줄 명령이 같은 문자열이 되는 것을 실제 Go 함수로 재현했다. acceptance의 shell·Python·YAML 예제에서는 줄바꿈과 들여쓰기가 의미이므로 단순 서식 변경으로 볼 수 없다.

조치·합격 조건: 원본 바이트 해시를 기본으로 사용한다. 서식 정규화가 꼭 필요하면 코드 블록을 보존하는 Markdown parser를 사용하되 복잡도와 이득을 먼저 측정한다. 서로 다른 실행 의미의 fixture는 반드시 다른 키여야 한다.

근거: internal/runtime/audit_cache.go:134 / P06. 원문·명령·stdout은 evidence.json의 동일 ID.

### F04 · High · 실패한 검사도 once-per-commit 표식으로 다음 호출에서 사라진다

분류: 동작 재현 / sync Stop 훅

훅은 검사 시작 전에 HEAD를 기록한다. go vet/build가 7로 실패한 첫 호출은 block JSON을 냈으나, 같은 HEAD의 두 번째 호출은 검사 없이 빈 stdout으로 종료했다. 현재 fixture에서 확인한 것은 훅 출력이며 실제 호스트의 Stop 처리까지 재현한 것은 아니다.

조치·합격 조건: running·passed·failed 상태와 tree key를 분리한다. 같은 실패 상태는 기존 block을 재표시하고, 중단된 running 상태는 회수 후 재검사한다. 성공 캐시만 조용히 재사용한다.

근거: .claude/hooks/moai/sync-phase-quality-gate.sh:174; .claude/hooks/moai/sync-phase-quality-gate.sh:185 / P04. 원문·명령·stdout은 evidence.json의 동일 ID.

### F05 · High · 의존성 취약점 검사 지시와 실제 훅의 manifest diff가 다르다

분류: 구현 대조 / sync 보안

sync 지침은 manifest 변경 여부와 무관한 취약점 검사를 Stop 훅이 수행한다고 선언한다. 실제 스크립트는 마지막 커밋의 manifest diff 유무만 기록하며, 소스 파일 변화가 없으면 그 전에 종료한다. 설치된 취약점 데이터베이스를 조회한 증거는 이 스크립트에 없다.

조치·합격 조건: manifest 변경 확인과 취약점 스캔을 서로 다른 검사로 명명한다. lockfile hash·scanner/version·advisory freshness로 스캔을 재사용하고, 실제 실행 명령·출력·누락 언어를 기록한다.

근거: .claude/skills/moai/workflows/sync/quality-gates-quality.md:140; .claude/hooks/moai/sync-phase-quality-gate.sh:310. 원문·명령·stdout은 evidence.json의 동일 ID.

### F06 · High · Critical/High 실패와 High 경고, 실패 후 계속 선택이 공존한다

분류: 문서 충돌 / sync 판정

Phase 7과 sync-auditor는 Critical/High 보안 문제를 FAIL로 정한다. Phase 8은 CRITICAL만 차단하고 HIGH는 경고로 처리한다. CRITICAL에서도 Continue with warning을 제시한다. 같은 취약점이 어느 단계에서 발견되는지에 따라 판정이 달라진다.

조치·합격 조건: severity와 blocking, 사용자가 승인한 예외를 하나의 판정 계약으로 정의한다. High fixture를 Phase 7·8·auditor에 각각 넣어 동일한 중단 결과와 예외 기록을 검증한다.

근거: .claude/agents/moai/sync-auditor.md:38; .claude/skills/moai/workflows/sync/quality-gates-quality.md:137. 원문·명령·stdout은 evidence.json의 동일 ID.

### F07 · High · manager-docs에 SPEC 본문 수정을 맡기면서 금지한다

분류: 문서 충돌 / sync 문서 소유권

sync Step 2.2.1은 requirements·plan·acceptance 본문을 구현에 맞춰 갱신하도록 한다. 실제 수행 주체 manager-docs는 해당 본문을 수정할 수 없고 manager-spec으로 되돌려야 한다. 이 지시쌍은 거절·재위임을 유발하거나 승인된 기준을 사후 변경하게 한다.

조치·합격 조건: divergence를 먼저 분류하고 필요한 SPEC 본문 변경만 manager-spec에 넘긴다. 변경된 AC는 독립 재검토 후 구현과 연결하며 docs는 승인된 차이와 frontmatter만 반영한다.

근거: .claude/skills/moai/workflows/sync/doc-execution.md:166; .claude/agents/moai/manager-docs.md:107. 원문·명령·stdout은 evidence.json의 동일 ID.

### F08 · High · 작은 Tier의 직접 push와 모든 Tier의 PR 의무가 충돌한다

분류: 문서 충돌 / run · sync Git

spec-workflow, run, sync는 Tier S/M에서 main 직접 push를 기본으로 명시한다. manager-git은 모든 Tier에서 push와 PR을 자신에게 위임하도록 한다. 현재 원격 보호 설정은 조회하지 않았으므로 서버에서 거절된다고 단정하지 않는다. 로컬 지침 간 경로 불일치는 확인됐다.

조치·합격 조건: 한 개의 delivery policy를 owner·tier·branch·approval 조건으로 정의하고 모든 호출부가 참조하도록 한다. S/M/L 및 명시적 --pr 조합의 실행계획을 표 기반 시험으로 대조한다.

근거: .claude/rules/moai/workflow/spec-workflow.md:25; .claude/agents/moai/manager-git.md:5. 원문·명령·stdout은 evidence.json의 동일 ID.

### F09 · High · plan 위치·branch switch·reset 지시가 공유 checkout 보호와 충돌한다

분류: 문서 충돌 / 공유 checkout

spec-workflow는 plan을 main checkout에서 실행하도록 하고 cleanup에 reset --hard를 제시한다. sync는 PR 직후 checkout을 지시한다. main-checkout guard와 이번 AGENTS 계약은 공유 트리의 branch 전환을 금지한다. 지침 실행을 위해 보호를 무시할 근거가 없다.

조치·합격 조건: launcher로 진입한 worktree에서 전 단계를 유지하고 별도 integration owner가 원격을 처리한다. 공유 checkout에서는 read-only probe와 허용된 명시 경로 commit만 가능하도록 경로 시험을 만든다.

근거: .claude/rules/moai/workflow/spec-workflow.md:50; .claude/skills/moai/workflows/sync/delivery.md:297; .claude/rules/moai/workflow/main-checkout-branch-guard.md:16. 원문·명령·stdout은 evidence.json의 동일 ID.

### F10 · High · lint 실패를 도구 미설치로 바꾸고 성공 종료한다

분류: 동작 재현 / sync CI 예제

which && golangci-lint || echo SKIP 구조에서 lint가 존재해도 nonzero이면 SKIP이 출력된다. lint=7 fixture는 미설치 문구와 exit 0을 반환했다. 이런 결과를 CI mirror PASS로 소비하면 실패를 숨긴다.

조치·합격 조건: 도구 존재 검사와 실행을 if/else로 분리하고 lint exit code를 보존한다. absent·pass·fail 세 fixture에서 각각 SKIPPED·PASS·FAIL을 확인한다.

근거: .claude/skills/moai/workflows/sync/delivery.md:137 / P02. 원문·명령·stdout은 evidence.json의 동일 ID.

### F11 · High · cross-build의 인자 없는 wait가 실패를 전달하지 않는다

분류: 동작 재현 / sync CI 병렬화

예제는 여러 build를 &로 시작하고 인자 없는 wait로 끝난다. 자식 하나가 7로 종료해도 wait_status=0을 반환했다. 출력 파일의 존재만으로 각 target 성공을 대신할 수도 없다.

조치·합격 조건: 각 PID와 target을 묶어 개별 wait status를 수집하거나 도구 수준의 독립 호출을 사용한다. 한 target 실패와 여러 target 실패 fixture 모두 전체 FAIL이어야 한다.

근거: .claude/skills/moai/workflows/sync/delivery.md:144; .claude/skills/moai/workflows/sync/delivery.md:149 / P03. 원문·명령·stdout은 evidence.json의 동일 ID.

### F12 · High · reconciliation이 명시된 SPEC도 검증 예제가 BLOCKING으로 만든다

분류: 동작 재현 / plan-auditor D7

D7 규칙은 retired/superseded 참조에 명시적 reconciliation이 없을 때 차단한다. 제공된 shell은 상태만 읽고 reconciliation 존재를 검사하지 않는다. explicit reconciliation 문장을 넣은 fixture에서도 BLOCKING이 출력됐다.

조치·합격 조건: 참조 ID별 reconciliation 근거를 문서 구조로 파싱한다. 조정 문구 있음/없음, 다른 ID의 문구, 역사 예시를 구분한 fixture로 오탐과 누락을 검증한다.

근거: .claude/agents/moai/plan-auditor.md:328; .claude/agents/moai/plan-auditor.md:315 / P10. 원문·명령·stdout은 evidence.json의 동일 ID.

### F13 · High · 다른 절의 build tag가 syscall 검증을 통과시킨다

분류: 동작 재현 / plan-auditor D8

규칙은 같은 절이나 문단의 build 제약을 요구하나 shell은 문서 전체에서 한 번만 찾는다. syscall 절에는 제약이 없고 다른 절에 //go:build가 있는 fixture가 아무 차단 출력 없이 끝났다.

조치·합격 조건: 각 syscall 사용과 해당 module의 build 제약을 연결한다. 문서에서는 같은 구조 노드의 명시적 연결을 검사하고, 코드가 생기면 해당 target의 cross-build를 별도로 검증한다.

근거: .claude/agents/moai/plan-auditor.md:343; .claude/agents/moai/plan-auditor.md:357 / P11. 원문·명령·stdout은 evidence.json의 동일 ID.

### F14 · High · project 문서 감사 호출과 auditor의 spec.md 필수 입력이 맞지 않는다

분류: 문서 충돌 / project 감사

project는 product·structure·tech를 만든 뒤 plan-auditor에 project 디렉터리를 넘긴다. agent의 입력 계약은 spec.md가 없으면 보고서 없이 AUDIT BLOCKED를 반환한다. project 전용 rubric을 선택하는 분기와 파일 계약이 agent 본문에 제시되지 않았다.

조치·합격 조건: 문서 유형을 typed input으로 전달하고 project용 필수 입력·평가항목·출력 경로를 정의한다. SPEC을 만들지 않은 정상 project fixture로 실제 감사 완료를 확인한다.

근거: .claude/skills/moai/workflows/project/doc-generation.md:49; .claude/agents/moai/plan-auditor.md:460. 원문·명령·stdout은 evidence.json의 동일 ID.

### F15 · Medium · 구현 전 sync-auditor 호출이 agent의 역할 제한과 겹친다

분류: 문서 충돌 / project · plan thorough

project thorough cross-validation, plan thorough SPEC-review, run의 구현 전 계약 협상은 sync-auditor를 사용한다. agent는 post-implementation only와 pre-implementation 금지를 명시하면서 자기 본문에는 계약 협상도 남아 있다.

조치·합격 조건: 계약 제안은 plan-auditor 또는 비판정 자문으로 이동하고 sync-auditor는 구현 증거 판정에 집중한다. 각 호출이 어떤 입력과 권한으로 실행되는지 계약 시험을 둔다.

근거: .claude/skills/moai/workflows/plan/spec-assembly.md:252; .claude/agents/moai/sync-auditor.md:6; .claude/agents/moai/sync-auditor.md:122. 원문·명령·stdout은 evidence.json의 동일 ID.

### F16 · Medium · Tier를 연구·계획·사용자 검토 뒤에 결정한다

분류: 절차 개선 / plan 비용

연구를 선택한 경로에서는 Phase 6이 research.md를 만들고 Phase 8이 상세 계획을 작성한 뒤 Phase 9에서 Tier를 묻는다. Tier S는 research.md 생략 대상이다. 연구는 권장 단계이므로 항상 발생하는 낭비가 아니라 최종 S인 작업에서의 중복 후보이다.

조치·합격 조건: 요청의 scope와 위험으로 임시 Tier를 먼저 정하고 사용자가 이미 정한 Tier를 재사용한다. 연구 중 근거가 생기면 상향한다. S fixture에서 불필요한 연구 artifact 생성 0회를 측정한다.

근거: .claude/skills/moai/workflows/plan/clarity-interview.md:87; .claude/skills/moai/workflows/plan/spec-assembly.md:29; .claude/skills/moai/workflows/plan/spec-assembly.md:55. 원문·명령·stdout은 evidence.json의 동일 ID.

### F17 · High · 작은 Tier가 생략한 acceptance.md를 하류가 여전히 요구한다

분류: 문서 충돌 / Tier S 하류 소비

Tier S는 AC를 spec.md에 둔다. 반면 manager-docs의 필수 CHANGELOG self-test는 acceptance.md만 읽으며, plan 완료 목록도 acceptance.md를 일률적으로 요구한다. S의 비용 절감 경로가 하류의 파일 누락 처리와 연결되지 않았다.

조치·합격 조건: resolveArtifacts(tier)와 resolveACSource(tier)를 단일 계약으로 사용한다. S는 spec inline, M/L은 acceptance를 읽고 누락과 빈 결과를 구분한다.

근거: .claude/skills/moai/workflows/plan/spec-assembly.md:55; .claude/agents/moai/manager-docs.md:89; .claude/skills/moai/workflows/plan/spec-assembly.md:517. 원문·명령·stdout은 evidence.json의 동일 ID.

### F18 · Medium · 허용된 AC 하위 ID a/b를 CHANGELOG 검사에서 하나로 센다

분류: 동작 재현 / AC 추적

manager-spec은 AC-V3R6-001a/001b를 서로 다른 하위 기준으로 허용한다. manager-docs의 grep은 숫자까지만 추출하여 두 ID를 1개로 계산했다. 통과한 AC 개수를 부정확하게 보고하거나 불필요한 재확인을 유발한다.

조치·합격 조건: AC 식별자 parser 또는 공용 정규식을 사용한다. suffix·domain-less·digit domain·table/headings의 중복을 구분하여 distinct count를 확인한다.

근거: .claude/agents/moai/manager-spec.md:139; .claude/agents/moai/manager-docs.md:89 / P12. 원문·명령·stdout은 evidence.json의 동일 ID.

### F19 · Medium · Tier 상한과 workflow 반복 지시가 다르다

분류: 문서 충돌 / 감사 반복 상한

harness와 auditor는 S=1/M=2/L=3으로 반복 수를 정한다. plan workflow는 standard/thorough에서 3회, cross-validation 불일치에는 추가 1회를 제시한다. 상한을 어느 계층에서 집행하는지 따라 호출 수가 달라진다.

조치·합격 조건: remaining_attempts를 orchestrator가 소유하고 모든 reviewer 호출이 같은 예산을 소비하도록 한다. timeout 재시도와 내용 수정 재감사는 별도 원인으로 기록하되 총 상한을 명시한다.

근거: .moai/config/sections/harness.yaml:75; .claude/skills/moai/workflows/plan/spec-assembly.md:252; .claude/agents/moai/plan-auditor.md:412. 원문·명령·stdout은 evidence.json의 동일 ID.

### F20 · Medium · 현재 phase 번호와 skip/resume 번호가 혼재한다

분류: 문서 충돌 / run 단계 분기

minimal skip 목록은 0.5·2.75·2.8a 등을 쓰지만 본문은 1~20으로 실행한다. Phase 1 cache hit 후 Phase 1로 진행, coverage 완료 후 Phase 1로 진행 같은 문구도 남아 있다. resume는 완료 phase를 건너뛰므로 숫자 혼용은 재실행·누락의 원인이 될 수 있다.

조치·합격 조건: 정수 순서와 별개로 안정적인 phase_id를 두고 skip/resume를 ID로 해석한다. 모든 jump 대상 존재, DAG 순환 없음, deprecated ID 매핑을 기계적으로 검사한다.

근거: .moai/config/sections/harness.yaml:84; .claude/skills/moai/workflows/run/phase-execution.md:54; .claude/skills/moai/workflows/sync/quality-gates-quality.md:317. 원문·명령·stdout은 evidence.json의 동일 ID.

### F21 · High · Phase 7에서 시작할 초안이 Phase 11 결과를 요구한다

분류: 의존관계 충돌 / sync 병렬 초안

docs drafters는 audit와 같은 Phase 7 turn에 출발해야 한다. 입력은 Phase 11 Step 1.5 divergence report다. entry router는 그 문서를 Phase 11에 on-demand로 읽도록 한다. 입력을 먼저 계산하는 명시적 preflight가 없으면 동시 실행 지시를 그대로 만족할 수 없다.

조치·합격 조건: diff·SPEC·divergence를 공통 immutable snapshot으로 먼저 만든 뒤 audit와 draft를 함께 시작한다. missing input으로 재위임한 횟수와 draft 폐기율을 기록한다.

근거: .claude/skills/moai/workflows/sync.md:67; .claude/skills/moai/workflows/sync/doc-execution.md:140. 원문·명령·stdout은 evidence.json의 동일 ID.

### F22 · Medium · 5개 docs·4개 judge·MX shard를 동일 turn에 요구한다

분류: 자원 계획 공백 / sync 병렬 수

문서 초안 5개, 품질 judge 4개, MX shard 3~5개를 동시 시작하도록 한다. 합계 12~14개는 사이트마다 인용하는 3~5 범위와 다르다. 최신 규칙은 3~5를 권고로 바꿨으므로 이를 런타임 hard-cap 위반으로 단정할 수 없지만 전체 배치 예산은 없다.

조치·합격 조건: 한 scheduler가 실제 잔여 슬롯을 읽고 전체 동시 수를 제한한다. MX와 구조 검증을 먼저, 이후 가능한 슬롯에 judge/draft를 넣고 queue 대기와 peak concurrency를 측정한다.

근거: .claude/skills/moai/workflows/sync/doc-execution.md:126; .claude/skills/moai/workflows/sync/quality-gates-quality.md:162; .claude/rules/moai/workflow/orchestration-mode-selection.md:32. 원문·명령·stdout은 evidence.json의 동일 ID.

### F23 · Medium · run의 4차원 결과와 sync의 판정 소유자가 서로 다르다

분류: 정책 차이 / run · sync 감사 비용

run은 같은 fan-out script의 결과를 evidence로만 인정하고 cold sync-auditor 판정을 추가한다. sync는 happy path에서 4dim 결과를 binding으로 승격하며 cold auditor를 생략한다. 서로 다른 목적일 수 있으나 전환 조건과 재사용 계약이 분리되어 있다.

조치·합격 조건: 증거 수집과 판정을 독립 계약으로 정의한다. 동일 tree·AC·rubric에서 재사용 가능한 범위를 명시하고, contested/INCOMPLETE에만 추가 검토한다. 실제 호출 감소는 trace로 확인한다.

근거: .claude/skills/moai/workflows/run/task-decomposition.md:110; .claude/skills/moai/workflows/sync.md:81. 원문·명령·stdout은 evidence.json의 동일 ID.

### F24 · Medium · 4차원 judge는 xhigh를 고정하고 model 정책은 xhigh를 배제한다

분류: 설정 불일치 / model · effort

sync-audit-4dim은 네 judge에 xhigh를 직접 지정한다. model-policy는 profile matrix에서 xhigh를 의도적으로 제외하고 목적별 resolver를 정의한다. 비용 차이는 이번에 측정하지 않았지만 선택 정책이 한 경로에 적용되지 않는 것은 확인된다.

조치·합격 조건: purpose=verify-judge의 resolved model/effort를 실행 입력으로 전달한다. 동일 fixture에서 profile별 유효 effort와 품질·출력 토큰·호출 횟수를 비교한 뒤 기준을 선택한다.

근거: .claude/workflows/sync-audit-4dim.js:188; .claude/rules/moai/development/model-policy.md:157. 원문·명령·stdout은 evidence.json의 동일 ID.

### F25 · Medium · 승인된 계획을 run에서 manager-spec이 다시 분석하고 승인받는다

분류: 절차 개선 / plan → run 문맥

plan Phase 8과 10이 계획을 확정한 뒤 run Phase 5는 다시 요구·전략을 추출하고 gate-run-1 승인을 받는다. 이어 Phase 6이 task를 분해한다. 변경 없는 재개에서도 어떤 계획 부분을 재사용하는지 명시된 차이 기반 조건이 없다.

조치·합격 조건: 승인된 plan hash와 execution task graph를 전달한다. 불변이면 입력 검증만 하고, 코드·scope 변화가 있는 부분만 재계획한다. 승인 의미를 유지하되 동일 질문 반복은 합친다.

근거: .claude/skills/moai/workflows/run/phase-execution.md:263; .claude/skills/moai/workflows/run/phase-execution.md:282; .claude/skills/moai/workflows/plan/clarity-interview.md:157. 원문·명령·stdout은 evidence.json의 동일 ID.

### F26 · Medium · 신규 프로젝트의 네 extended 질문을 별도 호출한다

분류: 절차 개선 / project 인터뷰

기존 프로젝트 Stage B는 남은 축을 한 호출에 묶는다. 신규 프로젝트는 네 축 각각을 별도 AskUserQuestion으로 요구한다. 같은 정보 구조에서 호출 수가 최대 4 대 1로 달라진다. 사용자의 실제 대기 시간은 측정하지 않았다.

조치·합격 조건: 신규·기존 모두 미확정 축만 한 batch로 묻는다. 축별 답변 품질과 명시적 empty 처리는 유지한다. 완성된 입력 fixture에서 재질문 수를 세어 확인한다.

근거: .claude/skills/moai/workflows/project/mode-detection.md:216; .claude/skills/moai/workflows/project/codebase-analysis.md:173. 원문·명령·stdout은 evidence.json의 동일 ID.

### F27 · Medium · 5개 기술 키워드면 인터뷰를 건너뛰지만 공통 규칙은 의도 완결을 요구한다

분류: 문서 충돌 / plan 명확성 평가

plan의 skip 조건은 기술 키워드 5개만으로 성립한다. 범위·성공 조건 없이 기술명만 나열한 요청도 해당될 수 있다. 공통 질문 규칙은 의도 명확성을 조건으로 한다. 키워드 개수는 의미 완결의 기계적 증거가 아니다.

조치·합격 조건: scope·constraints·acceptance·authorization의 필드 완결도를 검사한다. 기술명이 많고 목표가 없는 fixture와 짧지만 완전한 요청을 함께 검증한다.

근거: .claude/skills/moai/workflows/plan/context-discovery.md:45; .claude/rules/moai/core/askuser-protocol.md:67. 원문·명령·stdout은 evidence.json의 동일 ID.

### F28 · Medium · paths 없는 규칙 14개가 185,368바이트다

분류: 규모 측정 / 모든 단계 문맥

84개 규칙 전체를 읽어 frontmatter를 검사했다. paths 없는 파일은 14개이며 총 185,368바이트다. 이 값은 디스크의 잠재 상시 로딩 표면이며 실제 주입 토큰·캐시 요금은 아니다. kanban·cross-session·handoff 상세가 단일 작업에도 후보로 남는다.

조치·합격 조건: 상시 규칙은 짧은 dispatch 계약만 남기고 상세를 task-triggered companion으로 옮기는 방안을 우선 평가한다. runtime InstructionsLoaded 목록과 token usage로 실제 절감 여부를 검증한다.

근거: .claude/rules/moai/development/rule-authoring.md:14; .claude/rules/moai/workflow/kanban-dispatch.md:9 / P08. 원문·명령·stdout은 evidence.json의 동일 ID.

### F29 · High · fetch와 그 결과를 읽는 rev-list를 독립 병렬 검사로 묶는다

분류: 의존관계 충돌 / pre-spawn 검증

공통 pre-spawn 지시는 git fetch와 origin/main...HEAD 비교를 parallel batch에 넣는다. rev-list는 fetch가 갱신하는 ref를 소비하므로 둘은 독립이 아니다. 실제 경합을 재현하지 않았으며, 소스상 데이터 의존을 확인한 항목이다.

조치·합격 조건: fetch 완료 뒤 rev-list를 수행한다. 세션 목록 조회만 fetch와 병렬화한다. 지연된 fetch fixture에서 최신 remote ref를 읽는지 확인한다.

근거: .claude/rules/moai/core/agent-common-protocol.md:289; .claude/rules/moai/core/agent-common-protocol.md:296. 원문·명령·stdout은 evidence.json의 동일 ID.

### F30 · High · 읽기 전용 status가 auto-fix·tag 추가보다 뒤에서 종료한다

분류: 문서 충돌 / sync status

status는 read-only health check로 정의되지만 Phase 1 auto-fix, Phase 7 critical auto-fix, Phase 9 Add Missing Tags 뒤에 early exit가 위치한다. mode별 mutation guard가 각 단계에 일관되게 제시되지 않아 status의 안전 계약이 불명확하다.

조치·합격 조건: entry에서 readOnly=true를 전달하고 모든 writer 경로를 차단한다. lint/tag 결함이 있는 fixture의 status 전후 파일 해시와 git diff가 같아야 한다.

근거: .claude/skills/moai/workflows/sync/quality-gates-context.md:50; .claude/skills/moai/workflows/sync/quality-gates-quality.md:211; .claude/skills/moai/workflows/sync/quality-gates-quality.md:267. 원문·명령·stdout은 evidence.json의 동일 ID.

### F31 · Medium · auto sync도 전체 소스 스캔과 coverage test 생성을 요구한다

분류: 범위 확대 / sync 범위

auto는 changed files only로 정의되지만 Phase 11은 ALL source files를 스캔한다. Phase 10은 전체 coverage에서 P1/P2 gap의 테스트를 생성한다. 기존에 낮았던 unrelated package의 coverage까지 문서 동기화 작업으로 들어올 수 있다.

조치·합격 조건: auto는 diff와 영향받는 API·문서만, project는 전체로 분리한다. 기존 coverage debt는 따로 보고하고 변경 때문에 생긴 gap만 동일 scope에서 처리한다.

근거: .claude/skills/moai/workflows/sync/quality-gates-context.md:48; .claude/skills/moai/workflows/sync/doc-execution.md:40; .claude/skills/moai/workflows/sync/quality-gates-quality.md:301. 원문·명령·stdout은 evidence.json의 동일 ID.

### F32 · Medium · full suite·coverage·lint·cross-build를 여러 계층이 다시 요청한다

분류: 절차 개선 / 검증 자원

manager-develop의 최종 full suite, sync Phase 1·3, coverage 전후, local CI mirror가 별도 실행을 명시한다. 일부는 snapshot 소비를 도입했지만 Phase 3·coverage에는 같은 계약이 없다. 동일 run 횟수나 비용을 측정한 수치는 아니다.

조치·합격 조건: tree·command·toolchain·environment·scope를 키로 한 검증 계획을 공유한다. test+coverage를 가능할 때 한 실행으로 묶고, 부하에 민감한 benchmark는 단독 실행한다. 전체 suite는 위험과 CI 정책에 따라 배치한다.

근거: .claude/agents/moai/manager-develop.md:132; .claude/skills/moai/workflows/sync/quality-gates-context.md:158; .claude/skills/moai/workflows/sync/delivery.md:133. 원문·명령·stdout은 evidence.json의 동일 ID.

### F33 · Medium · 문서는 훅·judge의 공유 snapshot 소비를 선언하나 직접 호출 경로가 없다

분류: 구현 대조 / 검증 snapshot 연결

sync 지침은 Stop hook과 4dim judge가 동일 snapshot을 소비한다고 한다. 훅 전문에는 moai verify 호출이 없고 실제 vet/build를 실행한다. 4dim prompt도 명령 실행을 직접 요구한다. 런타임이 외부에서 증거를 주입하는지는 관찰하지 않았다.

조치·합격 조건: 실제 consumer entry에 조회·miss 실행·record를 연결하고 reuse/reexecute 사유를 출력한다. fresh fixture에서 test 도구 호출 0회, dirty fixture에서 1회인 spy 검증을 수행한다.

근거: .claude/skills/moai/workflows/sync/quality-gates-quality.md:43; .claude/hooks/moai/sync-phase-quality-gate.sh:225; .claude/workflows/sync-audit-4dim.js:176 / P09. 원문·명령·stdout은 evidence.json의 동일 ID.

### F34 · Medium · Kotlin Gradle 파일이 Java로 분류되고 monorepo는 첫 언어만 선택한다

분류: 동작 재현 / 언어 라우팅

sync diagnostics는 first match wins이고 Java가 build.gradle.kts를 먼저 소비한다. 훅의 detect_language도 같은 파일을 Java로 반환했다. run은 여러 언어를 모두 준비하도록 하므로 단계별 검사 범위가 달라질 수 있다.

조치·합격 조건: marker별 후보를 수집해 실제 plugin/source suffix로 구분하고 여러 언어를 반환한다. Kotlin-only, Java-only, Go+Node fixture의 검사 목록을 비교한다.

근거: .claude/skills/moai/workflows/sync/quality-gates-quality.md:19; .claude/hooks/moai/sync-phase-quality-gate.sh:62 / P05. 원문·명령·stdout은 evidence.json의 동일 ID.

### F35 · Medium · sync의 아무 변경도 없다는 종료 설명이 선행 mutation과 맞지 않는다

분류: 문서 충돌 / 중단·복구

Graceful Exit는 어느 decision point에서 취소해도 문서·Git·branch 변화가 없다고 한다. 하지만 앞 단계에서 auto-fix, tag 추가, coverage test 생성, commit을 수행할 수 있다. 백업도 Phase 12에서야 생긴다.

조치·합격 조건: 각 단계에 적용된 변경과 미적용 변경을 기록하고 취소 시 실제 보존 상태를 보고한다. 자동 rollback은 별도 승인·정확한 범위를 요구하며, 읽기 전용 preview와 apply 경계를 앞에 둔다.

근거: .claude/skills/moai/workflows/sync/delivery.md:399; .claude/skills/moai/workflows/sync/doc-execution.md:109. 원문·명령·stdout은 evidence.json의 동일 ID.

### F36 · Medium · 전체 docs·SPEC 복사와 non-empty 검사는 비용과 무결성을 혼동한다

분류: 절차 개선 / sync 백업

sync는 README·docs 전체·SPEC 전체를 매번 복사하고 non-empty directory check를 무결성 검사라 부른다. 파일 하나만 존재해도 디렉터리가 비어 있지 않으므로 원본 전체의 복구 가능성을 증명하지 못한다.

조치·합격 조건: 승인된 변경 대상만 content hash와 상대 경로 목록으로 백업한다. 기존 Git 객체로 복구 가능한 내용은 중복 복사를 줄이고 uncommitted 원본은 별도 보존한다. 파일별 round-trip 복구 시험을 둔다.

근거: .claude/skills/moai/workflows/sync/doc-execution.md:115; .claude/skills/moai/workflows/sync/doc-execution.md:116. 원문·명령·stdout은 evidence.json의 동일 ID.

### F37 · Medium · 정상 RED와 semantic failure 즉시 중단의 구분이 없다

분류: 문서 충돌 / TDD 오류 처리

manager-develop은 실패하는 테스트를 먼저 실행해야 한다. run의 자가진단 지침은 test assertion failure를 즉시 human escalation 대상으로 적는다. expected RED가 예외라는 명시적 상태 조건이 없어 정상 개발 사이클도 중단 대상으로 읽힌다.

조치·합격 조건: expected_RED, regression, tool_failure를 구분하고 RED 허용은 새 AC 테스트와 해당 cycle에만 한정한다. 기존 테스트 regression은 계속 차단한다.

근거: .claude/agents/moai/manager-develop.md:103; .claude/skills/moai/workflows/run.md:187. 원문·명령·stdout은 evidence.json의 동일 ID.

### F38 · Medium · phase마다 clear하라는 지시와 warm context 유지 지시가 공존한다

분류: 문서 충돌 / 문맥 압축

spec-workflow Plan token strategy는 완료 후 /clear를 지시한다. context-window와 cache-aware 규칙은 같은 작업의 문맥 축약·resume와 warm context 유지를 우선한다. handoff의 고정 문구와 증거 재로딩은 별도의 비용이며 실제 세션 비용은 이번에 측정하지 않았다.

조치·합격 조건: 작업 전환·실제 문맥 한계·감사 독립성 필요 여부로 clear를 결정한다. plan→run이 같은 승인 범위이면 간결한 handoff record를 유지하고 중복 로드를 측정한다.

근거: .claude/rules/moai/workflow/spec-workflow.md:180; .claude/rules/moai/workflow/context-window-management.md:41; .claude/rules/moai/workflow/cache-aware-execution.md:25. 원문·명령·stdout은 evidence.json의 동일 ID.

### F39 · Medium · team을 허용한 entry와 retired로 서술한 하위 경로가 섞여 있다

분류: 문서 충돌 / agent 호출 가능성

run entry와 최신 orchestration rule은 명시적 team을 허용한다. run phase-execution의 scale 설명과 router SKILL Step 3는 agent-team retired/fallback으로 적는다. 같은 요청이 로딩된 문서 위치에 따라 다르게 해석될 수 있다.

조치·합격 조건: 현재 capability와 역사 설명을 분리한다. 하나의 resolver를 두고 explicit team·unsupported runtime·no flag를 fixture로 검증한다. 지원을 단정하기 전에 현재 도구를 확인한다.

근거: .claude/skills/moai/workflows/run/context-loading.md:83; .claude/skills/moai/workflows/run/phase-execution.md:245; .claude/skills/moai/SKILL.md:366. 원문·명령·stdout은 evidence.json의 동일 ID.

### F40 · Medium · 문서의 TRACE PROBE 설명 자체는 실행 trace가 아니다

분류: 측정 공백 / 관측 · 성능 기준

워크플로우 파일의 MOAI_TRACE_PHASES와 enter/exit는 주석·의사 흐름에 있다. 이번에 실제 project→sync를 실행하거나 토큰·agent 대기·재시도 로그를 수집하지 않았다. 따라서 병목의 발생 가능성과 이미 재현한 결함을 실제 시간 절감률로 환산할 수 없다.

조치·합격 조건: 현재 routing ledger와 verification 기록을 먼저 활용해 stage_start/end, agent parent, input hash, cache hit, retry cause, tool duration을 묶는다. S/M/L·cold/warm·실패 주입별 p50/p95와 호출 수로 개선을 비교한다.

근거: .claude/skills/moai/workflows/project.md:30; .claude/skills/moai/SKILL.md:38. 원문·명령·stdout은 evidence.json의 동일 ID.


## 개선 순서

A. 판정 신뢰성부터 — runtime/verify·hook·auditor 지침 소유자. 이 보고서의 실패 fixture를 회귀 테스트로 편입한다. 실패→PASS 변환과 semantic hash collision을 먼저 막는다. [F02 F03 F04 F10 F11 F12 F13]
B. 권한·입력 계약 통일 — orchestrator·SPEC·Git 정책 소유자. tier/phase/mode/owner를 단일 표로 정의한다. 안전 기준을 낮추지 않고 모순된 호출부만 정렬한다. [F01 F05 F06 F07 F08 F09 F14 F17 F20 F29 F30]
C. 재실행·대기 줄이기 — workflow·scheduler·검증 소비자 소유자. 불변 입력 snapshot을 먼저 만들고 자원 한도 안에서 독립 작업만 병렬화한다. 검사별 재사용 이유를 기록한다. [F16 F19 F21 F22 F23 F24 F25 F26 F31 F32 F33 F34 F36]
D. 문맥·재개·실측 — harness·관측·문서 소유자. 상시 문맥을 조건부 상세로 분리할지 실측한다. 안정 phase ID와 실제 변경 기록을 사용하고 cold/warm 기준을 비교한다. [F15 F18 F27 F28 F35 F37 F38 F39 F40]

## Evidence

명령:
```sh
unset CLAUDE_PROJECT_DIR CLAUDE_SESSION_ID MOAI_SKIP_PLAN_AUDIT MOAI_AUDIT_GATE_T0 && go test ./internal/runtime ./internal/verify -run 'Test.*(Cache|Hash|Snapshot|Key|Fresh|Skip|RecordCheckConcurrent)' -count=1
```
출력(exit 0):
```text
ok  	github.com/modu-ai/moai-adk/internal/runtime	0.476s
ok  	github.com/modu-ai/moai-adk/internal/verify	4.992s
```

격리 probe 12개(P01–P12)의 명령과 원문 출력:
### P01

```text
git status --porcelain=v2; git diff HEAD — untracked new.go content A → B in isolated fixture
porcelain_before=? new.go
porcelain_after=? new.go
diff_before=""
diff_after=""
key_inputs_equal=true
computed_digest_equal=true
```
### P02

```text
bash -c: which() { return 0; }; golangci-lint() { return 7; }; which golangci-lint && golangci-lint run --timeout=5m || echo "SKIP: golangci-lint not installed"
exit=0
stdout=SKIP: golangci-lint not installed
stderr=
```
### P03

```text
bash -c: (exit 7) & (exit 0) & wait; printf "wait_status=%s" "$?"
exit=0
stdout=wait_status=0
stderr=
```
### P04

```text
bash <source>/sync-phase-quality-gate.sh, twice at identical fixture HEAD; fake go exits 7
{
  "first": {
    "exit": 0,
    "stdout": "{\"hookSpecificOutput\":{\"hookEventName\":\"Stop\",\"decision\":\"block\",\"reason\":\"go vet failed\"},\"systemMessage\":\"sync-phase quality gate BLOCKED: go vet failed (go vet=7 go build=7 deps_modified=1). Detail: .moai/logs/sync-quality-gate.log\"}",
    "stderr": ""
  },
  "second": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "tool_calls": [
    "vet ./...",
    "build ./..."
  ]
}
```
### P05

```text
source sync-phase-quality-gate.sh; detect_language <fixture with build.gradle.kts>
exit=0
stdout=java
stderr=
```
### P06

```text
go run normalize.go — normalizeWhitespace function extracted verbatim from internal/runtime/audit_cache.go
exit=0
stdout=different_source=true
normalized_equal=true
normalized="command: | echo first echo second"
stderr=
```
### P07

```text
go run key.go <fixture> — Key unchanged; only package/import and error-format adapter for standalone execution
exit=0
stdout=key_before=1c2af144e463ac5df7c9afc57cd82e847352f26c:a0de81d7346c1cb9
key_after=1c2af144e463ac5df7c9afc57cd82e847352f26c:a0de81d7346c1cb9
key_equal=true errors=<nil>,<nil>
stderr=
```
### P08

```text
Read every rule file; count paths frontmatter absence and UTF-8 bytes
{
  "files": 84,
  "lines": 15288,
  "bytes": 1131068,
  "no_paths_files": 14,
  "no_paths_bytes": 185368
}
```
### P09

```text
Check literal snapshot consumer symbols in full hook and workflow source
{".claude/hooks/moai/sync-phase-quality-gate.sh":{"verify_cli":false,"snapshot_term":false},".claude/workflows/sync-audit-4dim.js":{"verify_cli":false,"snapshot_term":true}}
```
### P10

```text
D7 verification verb extracted verbatim; fixture explicitly reconciles superseded SPEC-OLD-001
exit=0
stdout=BLOCKING: SPEC-OLD-001 has status=superseded but is referenced without reconciliation
stderr=
```
### P11

```text
D8 verification verb; syscall and build-tag occur in different sections
exit=0
stdout=
stderr=
```
### P12

```text
grep -oE AC-([A-Z0-9]+-)*[0-9]+ ac.txt | sort -u | wc -l; two suffix-distinct AC IDs
exit=0
stdout=1
stderr=
```

## 전수 조사 범위

84개 규칙별 paths·행수·hash·헤딩·규범문은 rules-inventory.json, 파일별 연결 분류는 HTML 표에 있다. 157개 파일은 baseline.json에 있다. 전체 바이트/메타데이터/규범문 기계 검사와 핵심 관련 조항 대조를 수행했다. 언어별 도구/API의 최신성, 전체 규칙의 모든 의미와 런타임 조합에 대한 독립 검증은 수행하지 않았다.

## Gaps

실제 project→sync end-to-end, native Stop hook의 block 처리, 외부 취약점 DB, 원격 branch 보호, 실제 모델 가격·토큰·queue/대기 시간, 전체 monorepo 언어별 검사, mirror/template 전체 동등성은 미검증이다. 문서 모순은 해당 문구를 대조한 결과이지 반드시 해당 분기가 실행됐다는 증거가 아니다.

## Residual-risk

157파일 해시가 같아도 범위 밖 설정이나 동시 세션은 변할 수 있다. 기존 선택 테스트의 PASS는 신규 fixture 결함을 반박하지 않는다. 정규식 감사 예제의 오탐/누락을 재현했지만 LLM auditor의 최종 판정까지 재현하지 않았다. 개선안은 아직 적용하지 않았고 실제 성능 향상률은 주장하지 않는다.
