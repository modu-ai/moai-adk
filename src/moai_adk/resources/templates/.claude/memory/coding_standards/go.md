# Go 규칙(요약)

- 도구: gofmt, goimports, golangci-lint, go test -race
- 에러: errors.Is/As, wrap(%w), sentinel 최소
- 동시성: context 전달, worker 제한, channel 누수 방지
- 테스트: table-driven, -race, benchmark 분리
- 구조: internal/ 패턴, 인터페이스 최소화, 사이드이펙트 경계화
