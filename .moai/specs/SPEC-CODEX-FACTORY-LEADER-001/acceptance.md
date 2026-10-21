# SPEC-CODEX-FACTORY-LEADER-001 — 수용 기준

모든 시나리오는 격리된 임시 프로젝트·홈·큐에서 실행한다. 실제 대화형 모델 실행은 launcher stub으로 대체한다. 실행 결과는 구현자가 관측해 progress.md §E.2에 명령과 출력을 남긴다.

## §A 수용 시나리오

### AC-CFL-001 — 리더 시작과 새 런 (REQ-CFL-001, REQ-CFL-002, REQ-CFL-011)

**Given** 격리된 프로젝트와 종료된 이전 Codex 리더 런 및 child capture stub, **When** launcher 진입점에서 bare `moai codex -f`와 `--factory`를 각각 실행하면, **Then** 각각 한 번 CLI child를 시작하고 새 active 런의 id·리더 역할을 child에 전달하며 저장된 리더 backend는 `codex`이고 이전 런을 인수하거나 레인 claim을 생성하지 않는다. 호출 환경에 기존 레인의 `MOAI_FACTORY_RUN_ID`·worker·역할 값이 있는 경우에도 기존 active 런의 id·backend·소유자가 바뀌지 않는다.

### AC-CFL-002 — 실제 큐의 dispatch와 완료 조회 (REQ-CFL-003, REQ-CFL-004)

이 기준의 레인 완료는 `merge-ready`와 저장된 작업 완료 증거다. 실제 `factory next`·`factory stage` 경로로 정당한 검증 증거를 갖추어 이 상태를 저장하고 리더가 기존 상태 조회로 읽는다. 큐 `CardDone`을 fixture로 만들어 대체하거나 실제 병합·착지를 요구하지 않는다.

**Given** launcher가 만든 Codex 리더 런과 실제 임시 큐의 발행·선택 가능한 카드, 검증된 레인 식별자 및 child launcher stub, **When** 메시징 nudge를 비활성화한 상태에서 stub이 실제 `moai todo engage --dispatch` 경로로 배차하고 레인이 진행·완료 기록을 저장한 뒤 큐·팩토리 저장소를 모두 다시 열어 기존 상태 조회 경로를 실행하면, **Then** 기존 큐 SQLite 파일의 `todo_runtime_assignments`·`todo_dispatch_current`에 저장된 카드·런·레인 식별자가 팩토리 SQLite 파일의 `card_dispatch`에 저장된 `card_id`·`run_id` 및 같은 두 키로 연결한 `cards.owner_label`과 일치하고 조회 결과가 같은 카드의 저장 상태를 반환한다. child launch 이전에 만든 fixture dispatch·fixture 출력과 저장소를 다시 열지 않은 메모리 값은 인정하지 않는다. 별도 dispatch sidecar 파일을 추가하지 않으며 nudge는 한 번도 필요하지 않다.

### AC-CFL-003 — malformed 입력의 무부작용 (REQ-CFL-005)

**Given** 런 DB·레인 registry·작업트리 목록의 이전 상태와 launch counter, **When** `-f 2`, `-f=2`, `--factory=2`, `-f worker`, `-f agent`, `-f lane-2`, `-f -l`, 중복 `-f`, `-f --factory-run x`, `-f app`, `-f status`를 각각 실행하면, **Then** 각각 non-zero 오류이고 child 횟수는 0이며 DB·registry·작업트리 목록과 factory 환경 값은 이전과 같다. 정상 bare -f를 같은 harness에서 실행하는 positive control은 한 번 launch한다.

### AC-CFL-004 — 기존 경로 회귀 (REQ-CFL-006)

**Given** 기존 Claude/GLM 리더·Codex -l·plain Codex의 launch capture harness, **When** 각 경로를 실행하면, **Then** 기존 리더 역할·런 기록·레인 claim·child argv 결과가 유지된다. plain Codex에는 새 리더 모델 인수가 들어가지 않으며 `--` 뒤 사용자 argv는 순서와 값을 보존한다.

### AC-CFL-005 — 모델과 anchor (REQ-CFL-007)

**Given** 기존 리더 기본 모델·effort 설정과 명시적 모델·effort 입력, **When** Codex 리더를 각 경우로 실행하면, **Then** 기본값 또는 명시적 덮어쓰기가 기존 우선순위로 child argv에 반영된다. **Given** 다른 writer가 살아 있거나 판별 불가능한 anchor, **When** 같은 작업트리에 리더를 실행하면, **Then** launch는 0회이고 기존 anchor가 보존된다.

### AC-CFL-006 — 혼합 레인과 교차 리더 경계 (REQ-CFL-008)

**Given** active Codex 리더 런, **When** 유효한 Claude/GLM/Codex 레인 합류를 기존 진입 형태로 실행하면, **Then** 허용된 각 레인이 같은 런에 귀속되고 중복·다른 런의 claim은 기존 검증으로 거부된다. **When** Claude/GLM 리더가 그 런을 인수하려 하면, **Then** non-zero이며 런 소유자와 리더 backend는 유지된다.

### AC-CFL-007 — 선행 문서와 역사 보존 (REQ-CFL-009)

**Given** 수정한 선행 SPEC와 `a9f53c169`, **When** HISTORY·요구사항·제외·관련 AC를 읽고 보존한 t1745 spec.md·plan.md를 원본 blob과 비교하면, **Then** 두 선행 HISTORY에 결정 `d-20261015T000351Z-0dfd`가 있고 bare Codex 리더와 혼합 레인 합류가 과거 금지와 충돌하지 않으며 t1745 두 파일은 byte-identical이고 `rejected`를 유지한다.

### AC-CFL-008 — 운영 안내 (REQ-CFL-004, REQ-CFL-010)

**Given** factory-dispatch 안내와 README, **When** Codex 리더 절을 읽으면, **Then** bare -f 시작·기존 -l 레인 진입·큐 dispatch·저장 완료 조회·지원하지 않는 훅·nudge 없는 확인·새 런 및 resume 제한이 명시되고 Codex를 기본 리더로 바꾸라는 지시는 없다.

### AC-CFL-009 — 실패 경로와 새 런 제한 (REQ-CFL-002, REQ-CFL-005, REQ-CFL-011)

**Given** child launch 실패 또는 실제 큐의 배차 가능한 카드 부재, **When** 리더 경로를 실행하면, **Then** 실패는 오류로 관측되고 launch-pending identity·소유자 자원이 기존 cleanup 정책대로 정리된다. 카드 부재는 dispatch 성공으로 기록되지 않는다. `--factory-run`이나 native resume 입력을 새 리더 재채택 기능으로 해석하지 않는다.

## §B 검증 기준

- 정상/오류/회귀 시나리오를 실제 launcher 경계에서 검사한다.
- AC-CFL-002는 launcher → 실제 todo engage --dispatch → 두 SQLite 저장소 재열기 → 저장 완료 상태 조회를 하나의 실행 경로로 검증한다.
- 변경 범위의 focused tests·vet·독립 감사 결과와 실행하지 않은 전체 CI·실제 모델 검증을 구분한다.
- Definition of Done: 필수 AC가 관측한 증거로 통과하고 동기화·운영자가 요청한 로컬 RC 설치 검증을 완료한다.
