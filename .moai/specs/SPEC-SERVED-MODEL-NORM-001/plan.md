# Plan — SPEC-SERVED-MODEL-NORM-001

마일스톤 M1 하나 (Tier S, 코드 파일 2개).

1. RED — `internal/hook/served_model_test.go` 의 `TestServedModel_Classify` 에 케이스 e~h 추가 (접미 달린 ID, 접미 달린 별칭, glm 대조, inherit).
2. GREEN — `internal/hook/served_model.go` 에 `normalizeModelDeclaration` 을 추가하고, `expectedServedModel` 이 정규화된 선언을 쓰며 `inherit` 을 선언 없음으로 취급하게 한다.
3. 재측정 — `go test ./internal/hook/...`, `go test ./internal/cli/ -run 'TestRunDiagnosticChecks|TestServedModelCheck_|TestBinaryLag_'`, golangci-lint v2.1.6, 수리 바이너리로 doctor 스윕 재집계.
