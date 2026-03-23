# SPEC-HOOK-009 인수 조건

## AC-1: 이벤트 타입 등록

**Given** types.go에 3개 이벤트 상수가 추가된 상태에서
**When** ValidEventTypes()를 호출하면
**Then** PostCompact, InstructionsLoaded, StopFailure가 모두 포함되어야 한다

## AC-2: PostCompact 핸들러

**Given** PostCompact 핸들러가 등록된 상태에서
**When** PostCompact 이벤트가 발생하면
**Then** 세션 복원 상태를 Data 필드에 JSON으로 반환해야 한다

## AC-3: InstructionsLoaded 핸들러

**Given** InstructionsLoaded 핸들러가 등록된 상태에서
**When** InstructionsLoaded 이벤트가 발생하면
**Then** 오류 없이 empty HookOutput을 반환해야 한다

## AC-4: StopFailure 핸들러

**Given** StopFailure 핸들러가 등록된 상태에서
**When** StopFailure 이벤트가 발생하면
**Then** 오류 없이 empty HookOutput을 반환해야 한다

## AC-5: 빌드 및 테스트

**Given** 모든 코드가 추가된 상태에서
**When** go build ./... && go test ./... 을 실행하면
**Then** 제로 에러로 통과해야 한다
