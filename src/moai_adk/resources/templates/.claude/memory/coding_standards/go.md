# Go 규칙

## ✅ 필수
- gofmt/goimports 자동 적용, golangci-lint를 PR 파이프라인에 포함, `go test -race` 실행
- 에러 처리: `errors.Is/As`/`fmt.Errorf("%w")`, sentinel 값 최소화, 컨텍스트 정보 포함
- context 전달 필수, goroutine 종료 신호 및 channel 누수 방지, worker 수 제한
- internal/ 디렉터리로 경계를 명확히 하고 인터페이스는 호출 지점 근처에 배치

## 👍 권장
- table-driven 테스트와 벤치마크를 분리, 커버리지 리포트 업로드
- DI 대신 함수 옵션 패턴, zero value 사용 설계, config는 환경 변수+viper 등으로 관리
- tracing/logging은 structured log + slog/zap, opentelemetry 연동 고려

## 🚀 확장/고급
- generics 활용 시 타입 제약 명확화, build tag로 플랫폼별 최적화
- ginkgo/gomega 등 BDD 테스트 프레임워크 병행, fuzz 테스트 도입
- 모듈 버전 관리 시 `go work`/monorepo 전략과 릴리스 태그 정책 정의
