# Swift 규칙

## ✅ 필수
- SwiftFormat + SwiftLint 로 코드 스타일 유지, CI에 swift test 포함
- async/await 사용 시 MainActor/UI 업데이트 규칙 준수, ConcurrentValue/Sendable 체크
- Value semantics(Struct) 우선, Protocol 기반 설계, 의존성은 주입(Factory/Environment)

## 👍 권장
- SwiftUI + Combine/AsyncSequence 사용 시 상태/뷰모델 분리, Previews로 빠른 검증
- XCTest + XCTExpectations, UI 테스트는 독립 타깃으로 구성, 스냅샷 테스트는 변화에 민감한 요소에만 사용
- Swift Package Manager 로 모듈화, Resources/Localization 분리

## 🚀 확장/고급
- Instrumentation(Time Profiler/Memory Graph)으로 성능/메모리 튜닝, MetricKit 로깅
- DocC 문서화, Swift Charts/ChartsKit 등 최신 프레임워크 탐색
- 서버 사이드 Swift(Vapor) 또는 멀티플랫폼(WatchOS/macOS) 확장 시 빌드 설정 템플릿 정의
