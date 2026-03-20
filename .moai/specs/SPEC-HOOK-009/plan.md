# SPEC-HOOK-009 구현 계획

## 구현 순서

1. `internal/hook/types.go` 편집: 3개 이벤트 상수 추가 + ValidEventTypes 업데이트
2. `internal/hook/post_compact.go` 생성: PostCompact 핸들러
3. `internal/hook/instructions_loaded.go` 생성: InstructionsLoaded 핸들러
4. `internal/hook/stop_failure.go` 생성: StopFailure 핸들러
5. `internal/cli/deps.go` 편집: InitDependencies()에 3개 핸들러 등록
6. 테스트 파일 생성 (각 핸들러별)
7. go build + go test 검증

## 핸들러 설계

### PostCompact (compact.go 패턴 참조)
- compaction 완료 후 상태 복원 로깅
- Data 필드에 compact 결과 JSON 반환

### InstructionsLoaded
- CLAUDE.md/rules 로딩 확인 로깅
- 프로젝트 설정 검증 (`.moai/config/` 존재 여부)
- non-blocking, 항상 empty output 반환

### StopFailure
- API 오류 정보 로깅 (rate limit, auth 등)
- error 필드 파싱하여 복구 가이던스 결정
- non-blocking, 항상 empty output 반환
