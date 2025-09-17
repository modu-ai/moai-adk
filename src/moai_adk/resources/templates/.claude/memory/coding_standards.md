# 코딩 및 아키텍처 기준(빠른 참조)

> 대상: 구현/리뷰 담당 개발자. 작업 전후에 필수 규칙을 확인하고, 언어별 세부 규칙(@imports)을 따라주세요.

## ✅ 핵심 체크리스트
- 작은 단위: 파일 ≤ 300 LOC, 함수 ≤ 50 LOC, 매개변수 ≤ 5, 순환 복잡도 < 10
- 단일 책임: 입력 → 처리 → 반환 구조, 가드절 우선, 부수효과(I/O·네트워크·전역)는 경계층에 격리
- 명시성: 의도를 드러내는 이름, 상수 심볼화, 구조화 로깅(민감정보 금지·요청/상관관계 ID 전파)
- 테스트/PR/보안 점검은 @.claude/memory/shared_checklists.md 를 참조(커버리지, 입력 검증, 시크릿 노출 방지 등)
- 보안 세부 규칙은 @.claude/memory/security_rules.md, TDD 흐름은 @.claude/memory/tdd_guidelines.md 를 따른다

## 🚫 금지 패턴
- 광범위 예외(`except Exception`, catch-all) 남용, 경고/실패 무시, 근거 없는 최적화/추상화
- 비밀/민감 정보의 코드·로그·문서 노출, 하드코딩된 키·비밀번호·토큰
- 순환 의존/Feature Envy/데이터 덩어리(Data Clump) 방치, 추적성(@TAG) 누락

## 📁 언어·플랫폼 프로파일(@imports)
| 문서 | 주요 내용 |
| --- | --- |
| @.claude/memory/coding_standards/python.md | Python 필수/권장/확장 규칙, 도구/타입/예외/테스트 |
| @.claude/memory/coding_standards/typescript.md | TypeScript strict 설정, 타입 전략, 런타임 검증 |
| @.claude/memory/coding_standards/go.md | Go 모듈 경계, 에러 처리, 병행성 패턴 |
| @.claude/memory/coding_standards/java-kotlin.md | Spring/Gradle 구성, null 안전성, 레이어드 구조 |
| @.claude/memory/coding_standards/csharp.md | .NET DI, async/await 패턴, Analyzer 설정 |
| @.claude/memory/coding_standards/rust.md | Ownership/borrowing, Result/Option 처리, cargo fmt/clippy |
| @.claude/memory/coding_standards/swift.md | SwiftUI/Combine, async/await, Xcode 설정 |
| @.claude/memory/coding_standards/sql.md | 스키마 버전, 인덱싱, 마이그레이션/테스트 |
| @.claude/memory/coding_standards/shell.md | set -euo pipefail, 인자 quoting, 안전한 파일 조작 |
| @.claude/memory/coding_standards/terraform.md | 모듈화, workspace/환경 분리, drift 감지 |
| @.claude/memory/coding_standards/frameworks.md | React/Vue/Angular/FastAPI/Spring 등 프레임워크별 주의점 |

각 문서는 “필수/권장/확장” 구조로 정리되어 있으며, 필요한 규칙만 선택적으로 임포트할 수 있습니다.

## 🧱 아키텍처 & 구조
- 레이어드 아키텍처와 모듈 간 의존성 방향은 @.claude/memory/software_principles.md (SOLID, Refactoring, API 패턴)과 @.claude/memory/coding_standards/frameworks.md 를 참고
- 도메인 경계/ADR 작성/기술 선택은 @.claude/memory/project_guidelines.md 및 계획 단계 커맨드(`/moai:3-plan`) 지침을 따름

## 🧪 검증 흐름
1. 구현 전/후 @.claude/memory/shared_checklists.md 로 PR/테스트/보안 항목을 확인
2. 테스트는 @.claude/memory/tdd_guidelines.md 의 Red-Green-Refactor 사이클로 실행
3. 보안/개인정보 처리는 @.claude/memory/security_rules.md 의 ISMS-P 규정을 준수

## 📌 참고 문서
- 운영 전반: @.claude/memory/project_guidelines.md
- 팀 협업/PR: @.claude/memory/team_conventions.md, @.claude/memory/git_workflow.md, @.claude/memory/git_commit_rules.md
- 메모리 구조/임포트 안내: @.claude/memory/README.md
