# SPEC-MOAI-GATEWAY-PICKER-001 — 수용 기준

**AC-GP-001** (REQ-GP-001) — Given 명시 모델·launcher 기본값·빈 기본값을 각각 준비한 새 moai gpt 세션이 있을 때, When 각 조합으로 실행할 때, Then 실제 Claude argv에 초기 모델이 명시되고 첫 turn의 모델은 우선순위 결과와 같다. 빈 기본값에서는 gpt-5.6-sol이며 사용자 표시도 같다.

**AC-GP-002** (REQ-GP-002) — Given 전역 저장 모델을 다른 provider의 ID로 오염시킨 cc·glm·gpt 세션이 있을 때, When 기본값 없이 각각 실행할 때, Then 각 launcher의 첫 실제 요청은 자기 provider 기본 모델로 나가며 오염 모델 endpoint의 요청 계수는 0이다. cc는 Opus 5·Sonnet 5 설정을 각각 검증한다.

**AC-GP-003** (REQ-GP-003) — Given 이전 provider와 대화 기록이 있는 저장 세션이 있을 때, When continue·resume을 각각 명시 모델 있음/없음 조합으로 실행할 때, Then 기록은 이어지고 최초 실제 요청 모델은 이번 launcher 우선순위와 일치한다. 프로세스 재시작에서도 동일하며 새 세션 시험만으로 통과하지 않는다.

**AC-GP-004** (REQ-GP-004) — Given 네 GPT ID가 catalog에 등록된 실제 TUI가 있을 때, When 네 ID를 /model로 각각 선택하고 같은 네 ID를 picker s 경로에서도 선택할 때, Then 각 선택 뒤 실제 turn은 정확한 ID로 도착한다. 존재하지 않는 ID는 명시 거절되고 어떤 upstream에도 전달되지 않는다.

**AC-GP-005** (REQ-GP-005) — Given 사용자·프로젝트 설정에 서로 다른 표지를 두고 두 세션을 동시에 실행할 때, When 각 세션에서 서로 다른 /model을 선택한 뒤 정상 종료·취소를 각각 수행할 때, Then 원본 설정 파일의 내용 해시·존재·권한이 모두 같고 한 세션의 선택이 다른 세션의 요청에 영향을 주지 않는다. 세션 설정 잔존과 cleanup 실패도 명시 기록한다.

**AC-GP-006** (REQ-GP-006) — Given cc·gpt·glm의 초기 기본 모델이 서로 다르고 GLM 슬롯 오염이 있는 TUI일 때, When 다른 모델로 전환한 뒤 Default 행을 선택할 때, Then 화면에 표시한 초기 기본 모델·provider와 다음 실제 turn의 모델·provider가 일치한다. 표시를 바꿀 수 없거나 뜻이 모호한 행은 성공으로 판정하지 않는다.

**AC-GP-007** (REQ-GP-007) — Given 이전 turn의 식별 문장과 도구 결과가 있는 같은 실제 Claude Code 세션이 있을 때, When 네 GPT 모델을 순회하며 회상 질문과 승인된 임시 파일 도구 왕복을 수행할 때, Then 각 응답이 대화 식별 문장·도구 결과를 보존하고 stream 종료가 관측된다. 검증 요청만 기록되거나 mock 응답만 있으면 실제 모델 AC는 미완료다.

**AC-GP-008** (REQ-GP-008) — Given credential이 없는 모델을 가진 TUI와 요청 계수기들이 있을 때, When /model로 선택하고 picker s로도 선택할 때, Then /model 거절 뒤 다음 turn은 이전 선택으로 나가고 s 경로는 첫 turn에서 명시 인증 오류가 난다. 다른 provider로의 fallback 요청은 0이다.

**AC-GP-009** (REQ-GP-009) — Given 코어 launch 정리·모델 선택 UI가 함께 설치된 상태일 때, When GLM tier 슬롯 회귀 시험과 기본 노출 경로 검증을 수행할 때, Then 슬롯 순서·값은 코어 AC-MG-018·025를 충족하고 AUTH 미충족 상태는 전체 목표 성공으로 표시되지 않는다. 코어 단독으로 기본 노출을 켜는 상태는 출시 게이트에서 거절된다.


## 기존 AC의 사전 측정 보강

AC-GP-004의 picker 노출 판정은 `plan.md` A.2의 양성 `--settings` 대조군과 프로젝트/local 음성 대조군을 포함한다.
문서상 discovery 필터에 맞는 Claude 항목과 bare GPT를 같이 제공하고 discovery만으로 GPT 행이 생기지 않는지 확인한다.
custom-header-only의 discovery 생략은 정상 프로토콜 분기이며, 그것만으로 picker 표시나 계정 사용 가능을 판정하지 않는다.
managed/allowlist 변형은 실제 적용 상태·누락 이유·upstream 요청 0을 기록한다. 정책을 무시하여 행을 강제로 만드는 것은 적색이다.

AC-GP-006은 Default와 현재 모델 행이 overlay 교체 후에도 남는 조건에서 검사한다. `ANTHROPIC_DEFAULT_MODEL`의 값만
읽어 통과시키지 않고 화면의 뜻과 다음 요청을 비교하며, 조직 기본값·계정 제한 변형도 별도로 판정한다.

AC-GP-005는 overlay 입력 파일뿐 아니라 `/model`이 실제 쓴 목적지를 확인한다. 같은 클라이언트 조건에서 `s` 경로의
저장 없는 선택을 대조군으로 사용한다. 사용자 credential 복사로 설정 격리를 달성한 결과는 수용하지 않는다.

## 공통 실계정 시험 조건

Claude 실서비스 시험은 사용자 지시에 따라 **2026-09-11 19:00 Asia/Seoul 이후** 실행한다. Claude 시험 입력은
`claude-opus-5`와 `claude-sonnet-5`이며 계정의 실제 사용 가능 여부는 별도 관측한다. Sonnet 4.5 과거 캡처는 대체 근거가 아니다.
GPT 시험의 upstream ID는 `gpt-6-astra`, `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`다. 임의 `gpt-6` 별칭을 만들지 않는다.
공식 `gpt-5.6` 별칭을 UI에 추가하더라도 공개 근거와 canonical 매핑 시험이 먼저 필요하며 본 초안의 필수 범위에는 넣지 않는다.
429는 실패한 인증 방식의 증명도 성공도 아니다. Claude OAuth의 코어 T09가 INCONCLUSIVE이면 그 상태를 유지한다.

## 근거와 완료 판단

사용자 범위·공식 모델 문서 위치는 `../../reports/SPEC-MOAI-GATEWAY-001/goal-execution-20260911.md`, 코어 경계는
`../SPEC-MOAI-GATEWAY-001/spec.md` 및 `design.md`가 기준이다. 코어의 기존 lint·mock PASS는 이 SPEC의 완료 근거가 아니다.
각 AC별 실행 명령·그 출력·HEAD·사용한 모델·mock/실계정 구분·Gap·잔여 위험을 보고한다. 관련 구현·통합 시험·독립 감사와
문서 동기화 전 완료로 표시하지 않는다. 이번 계획 작성은 numeric card 생성이나 기존 worktree의 새 카드 착수가 아니다.

## 0.2.0 관측을 반영한 기존 AC 판정 경계

AC-GP-003·005: Given 실제 client에서 같은 UUID로 재개한 합성 양성 관측이 있을 때, When 제품 continue·resume·재시작과
동시 세션을 시험하면, Then 각 경로의 실제 model·공개 history·필요 carrier·설정 해시를 따로 판정한다. 합성 정확한 resume
한 번으로 모든 경로를 PASS하지 않는다. receipt 경로의 삭제·잘못된 귀속은 코어 AC-MG-009와 함께 송신 0으로 검증한다.

AC-GP-004·006: Given 실제 안내의 s=세션 선택, Enter=기본값 선택과 Default 로컬 요청 관측이 있을 때, When 제품 UI를
검증하면, Then 선택 뒤의 실제 turn을 확인한다. Default 직후 snapshot과 후속 /model 또는 재시작 뒤 snapshot을 분리하며
저장 안내만으로 파일 쓰기·fresh session 적용을 판정하지 않는다. 기존 조직 제한·원본 설정 보존 음성군을 삭제하지 않는다.

## RPA-1 연계 — 기존 AC-GP-003·005 보강

Given 소유 대화 인덱스와 retained native profile/transcript·receipt가 있을 때, When new/exact resume/continue/선택
resume/fork를 실행하면, Then 코어 design §4.3 lifecycle 표의 사전 UUID와 root로만 시작해야 한다. 잘못된 UUID·transcript
삭제·누락 manifest는 송신 0이며 첫 metadata를 보고 새 귀속을 만들어서는 안 된다. 취소와 정상 종료 뒤 retained root가
남고 temporary overlay만 정리되어야 한다. 재개한 native의 공개 history·필요 carrier·이번 초기 모델을 직접 판정한다.

Given 같은 family의 두 lead와 다른 family의 두 lead를 각각 준비했을 때, When /model 저장을 시도하면, Then 같은 family는
한 실행만 lease를 가지고 다른 실행은 명시 busy로 끝나며, 다른 family는 서로 설정을 바꾸지 않고 동시 동작해야 한다.
원래 secure-storage namespace 보존과 여섯 원본 설정의 불변은 실제 TUI/OAuth 통합에서 함께 확인한다. source/print의
분리된 관측을 전체 PASS로 대체하지 않는다. fork·선택형 resume의 native 미관측 행은 삭제하지 않고 미완료로 기록한다.
