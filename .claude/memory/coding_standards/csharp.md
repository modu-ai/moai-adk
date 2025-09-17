# C#/.NET 규칙

## ✅ 필수
- `dotnet format`, StyleCop 분석, xUnit + coverlet 으로 테스트/커버리지 확보
- 프로젝트는 `<Nullable>enable</Nullable>` 유지, async/await 규칙과 CancellationToken 전파
- DI 컨테이너는 구체 타입 대신 인터페이스 등록, 옵션/설정 바인딩은 `ValidateOnStart()` 활용
- WebApplicationFactory/TestServer 를 사용해 API 통합 테스트, contract tests 수행

## 👍 권장
- minimal APIs + Endpoint filter 조합, MediatR/vertical slice 패턴 적용
- Logging(Serilog) + Observability(Application Insights/OpenTelemetry) 연계
- EF Core: context per scope, migration 분리, seed/test data는 builder 패턴 사용

## 🚀 확장/고급
- Source Generator/Analyzer 로 팀 전용 규칙 확장, Roslyn analyzer 경고를 빌드에 포함
- Multi-target 패키지, NativeAOT, gRPC/SignalR와의 통합
- Azure Functions/AWS Lambda 등 서버리스 실행환경을 고려한 DI/구성 최적화
