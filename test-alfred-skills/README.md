# Alfred Skills Testing Suite

테스트 목적: 4개 Alfred 스킬들의 기능적 완성도와 통합 작동 여부 검증

## 테스트 대상 스킬들

1. **moai-alfred-clone-pattern** - 복잡한 다단계 작업 자율 위임
2. **moai-alfred-personas** - 적응형 커뮤니케이션 및 전문성 탐지
3. **moai-alfred-reporting** - 표준화된 보고서 생성 및 출력 형식
4. **moai-alfred-doc-management** - 문서 관리 규칙 및 위치 결정

## 테스트 구조

```
test-alfred-skills/
├── README.md                    # 이 파일
├── test-data/                   # 테스트용 입력 데이터
├── expected-outputs/            # 예상 출력 결과
├── .moai/                       # MoAI-ADK 프로젝트 구조
│   ├── docs/                    # 가이드 문서 저장
│   ├── reports/                 # 분석 보고서 저장
│   ├── analysis/                # 깊이 분석 결과 저장
│   └── specs/                   # SPEC 문서 저장
├── test-logs/                   # 테스트 실행 로그
└── test-results/                # 최종 테스트 결과
```

## 테스트 시나리오

### 개별 스킬 테스트
- 각 스킬의 핵심 기능별 단위 테스트
- 경계 조건 및 예외 상황 테스트
- 예상 출력과 실제 출력 비교

### 통합 워크플로우 테스트
- 4개 스킬이 함께 작동하는 시나리오 기반 테스트
- Alfred의 4단계 워크플로우와의 연동 테스트
- 실제 개발 시나리오 시뮬레이션

## 테스트 실행 방법

각 테스트는 별도의 Claude Code 세션에서 실행하거나,
자동화된 스크립트를 통해 순차적으로 실행할 수 있습니다.

---

*테스트 시작일: 2025-11-05*
*의도한 동작 검증: 4개 Alfred 스킬들의 완벽한 구현 및 통합*