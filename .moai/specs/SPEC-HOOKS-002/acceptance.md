# SPEC-HOOKS-002 Acceptance Criteria

> **moai_hooks.py Self-contained Hook Script**
>
> Given-When-Then 형식의 수락 기준

---

## Scenario 1: PEP 723 Compliance

**Given** moai_hooks.py 파일이 존재한다  
**When** 파일 헤더를 확인한다  
**Then** PEP 723 inline metadata가 존재해야 한다  
**And** shebang `#!/usr/bin/env python3`가 존재해야 한다  
**And** `requires-python = ">=3.14"`가 존재해야 한다  
**And** `dependencies = []` (빈 배열)이어야 한다  
**And** docstring에 사용법이 명시되어야 한다

---

## Scenario 2: SessionStart Hook - Git 정보 수집

**Given** Git 저장소가 있는 MoAI-ADK 프로젝트  
**And** 현재 브랜치가 "feature/SPEC-HOOKS-002"  
**When** SessionStart 이벤트가 트리거된다  
**Then** Git 브랜치 정보가 출력되어야 한다  
**And** 마지막 커밋 메시지가 표시되어야 한다  
**And** 변경된 파일 수가 표시되어야 한다 (예: "3 files")  
**And** 실행 시간이 500ms 이내여야 한다  
**And** JSON 형식으로 출력되어야 한다 (`success: true`)

---

## Scenario 3: SessionStart Hook - SPEC 진행률

**Given** .moai/specs/ 디렉토리에 5개의 SPEC이 있다  
**And** 그 중 3개의 SPEC에 plan.md가 작성되어 있다  
**When** SessionStart 이벤트가 트리거된다  
**Then** SPEC 진행률이 "3/5 (60%)" 형식으로 표시되어야 한다  
**And** 배너에 SPEC Progress 정보가 포함되어야 한다  
**And** 계산이 정확해야 한다 (completed/total)

---

## Scenario 4: SessionStart Hook - 언어 감지

**Given** pyproject.toml 파일이 있는 Python 프로젝트  
**And** tsconfig.json 파일이 있는 TypeScript 프로젝트  
**When** SessionStart 이벤트가 트리거된다  
**Then** Python이 감지되어야 한다  
**And** TypeScript가 감지되어야 한다  
**And** 배너에 "Languages: Python, TypeScript"가 표시되어야 한다  
**And** 20개 언어 중 해당하는 언어가 모두 감지되어야 한다

---

## Scenario 5: UserPromptSubmit - JIT Context (/alfred:1-spec)

**Given** 사용자가 "/alfred:1-spec 새 기능 추가" 프롬프트를 입력한다  
**And** .moai/memory/spec-metadata.md 파일이 존재한다  
**When** UserPromptSubmit 이벤트가 트리거된다  
**Then** spec-metadata.md가 relevant_docs에 포함되어야 한다  
**And** JSON 결과에 "JIT Context: 1 documents" 메시지가 포함되어야 한다  
**And** 실행 시간이 100ms 이내여야 한다  
**And** `success: true`를 반환해야 한다

---

## Scenario 6: UserPromptSubmit - JIT Context (/alfred:2-build)

**Given** 사용자가 "/alfred:2-build SPEC-AUTH-001" 프롬프트를 입력한다  
**And** .moai/memory/development-guide.md 파일이 존재한다  
**When** UserPromptSubmit 이벤트가 트리거된다  
**Then** development-guide.md가 relevant_docs에 포함되어야 한다  
**And** JSON 결과에 관련 문서 목록이 포함되어야 한다  
**And** 실행 시간이 100ms 이내여야 한다

---

## Scenario 7: UserPromptSubmit - 테스트 키워드 감지

**Given** 사용자가 "run tests for authentication" 프롬프트를 입력한다  
**And** tests/ 디렉토리가 존재한다  
**When** UserPromptSubmit 이벤트가 트리거된다  
**Then** tests/ 디렉토리가 relevant_docs에 포함되어야 한다  
**And** 테스트 파일 패턴 (`*.test.py`, `*.spec.ts`)이 매칭되어야 한다  
**And** 실행 시간이 100ms 이내여야 한다

---

## Scenario 8: Language Detection - 20개 언어

**Given** 다양한 프로그래밍 언어 시그니처 파일이 있는 프로젝트  
**When** LanguageDetector.detect()가 실행된다  
**Then** 다음 언어들이 정확히 감지되어야 한다:
- Python (pyproject.toml 존재)
- TypeScript (tsconfig.json 존재)
- JavaScript (package.json 존재)
- Ruby (Gemfile 존재)
- PHP (composer.json 존재)
- Go (go.mod 존재)
- Rust (Cargo.toml 존재)
- Java (pom.xml 존재)
- Kotlin (build.gradle.kts 존재)
- Swift (Package.swift 존재)
- Dart (pubspec.yaml 존재)
- C (CMakeLists.txt + *.c 존재)
- C++ (CMakeLists.txt + *.cpp 존재)
- C# (*.csproj 존재)
- Scala (build.sbt 존재)
- Elixir (mix.exs 존재)
- R (DESCRIPTION 존재)
- Julia (Project.toml 존재)
- Perl (*.pl 존재)
- Shell (*.sh 존재)

**And** 언어가 하나도 감지되지 않으면 "Unknown Language"를 반환해야 한다

---

## Scenario 9: Zero Dependencies

**Given** moai_hooks.py 파일  
**When** import statements를 확인한다  
**Then** 모든 import가 Python 표준 라이브러리여야 한다  
**And** sys, os, json, subprocess, glob, enum, typing만 사용해야 한다  
**And** requests, click, rich 등 외부 패키지를 사용하지 않아야 한다  
**And** pip install 명령이 필요하지 않아야 한다  
**And** PEP 723 dependencies = [] (빈 배열)이어야 한다

---

## Scenario 10: SPEC Progress Calculation

**Given** .moai/specs/ 디렉토리에 5개의 SPEC 디렉토리가 있다  
**And** SPEC 디렉토리명이 "SPEC-AUTH-001", "SPEC-USER-001", ... 형식이다  
**And** 그 중 3개에 plan.md 파일이 존재하고 내용이 있다  
**When** SpecProgressTracker.calculate()가 실행된다  
**Then** completed = 3이어야 한다  
**And** total = 5여야 한다  
**And** progress = 60.0%여야 한다  
**And** format() 결과가 "3/5 (60%)" 형식이어야 한다

---

## Scenario 11: Git Timeout Protection

**Given** Git 명령어가 매우 느린 저장소  
**When** GitInfoProvider.get_info()가 실행된다  
**Then** 2초 이내에 타임아웃되어야 한다  
**And** None을 반환하거나 부분 정보를 반환해야 한다  
**And** 프로그램이 중단되지 않아야 한다  
**And** 에러가 stderr에 출력되지 않아야 한다 (gracefully degrade)

---

## Scenario 12: Git 저장소 아닐 때

**Given** Git 저장소가 아닌 디렉토리 (.git/ 없음)  
**When** SessionStart 이벤트가 트리거된다  
**Then** Git 정보 섹션이 생략되어야 한다  
**And** 배너에 SPEC 진행률과 언어 정보만 표시되어야 한다  
**And** 에러가 발생하지 않아야 한다  
**And** `success: true`를 반환해야 한다

---

## Scenario 13: Error Handling - Invalid JSON

**Given** stdin으로 잘못된 JSON이 입력된다  
**And** 입력: `"invalid json"`  
**When** moai_hooks.py가 실행된다  
**Then** stderr에 "Invalid JSON" 에러 메시지가 출력되어야 한다  
**And** exit code 1로 종료되어야 한다  
**And** JSON 형식의 에러 결과를 stderr에 출력해야 한다  
**And** `success: false`를 포함해야 한다

---

## Scenario 14: Error Handling - Unknown Event

**Given** stdin으로 알 수 없는 이벤트가 입력된다  
**And** 입력: `{"event":"UnknownEvent","cwd":"."}`  
**When** moai_hooks.py가 실행된다  
**Then** JSON 결과에 `success: false`가 포함되어야 한다  
**And** `error: "Unknown event: UnknownEvent"` 메시지가 포함되어야 한다  
**And** stdout으로 출력되어야 한다 (stderr 아님)  
**And** exit code 0으로 종료되어야 한다 (정상 처리)

---

## Scenario 15: File Size Constraint

**Given** moai_hooks.py 파일  
**When** 파일 라인 수를 확인한다 (`wc -l moai_hooks.py`)  
**Then** 라인 수가 600줄 이하여야 한다  
**And** 빈 줄과 주석을 포함한 총 라인 수여야 한다  
**And** 코드가 간결하고 가독성이 좋아야 한다

---

## Scenario 16: PreCompact Hook - 세션 요약

**Given** 긴 세션이 진행 중이다 (컨텍스트 70% 사용)  
**When** PreCompact 이벤트가 트리거된다  
**Then** 세션 요약이 생성되어야 한다  
**And** SPEC 진행률이 포함되어야 한다  
**And** 최근 변경사항이 포함되어야 한다  
**And** 권장사항 메시지가 포함되어야 한다: "Use /clear or /new for next session"  
**And** JSON 형식으로 출력되어야 한다  
**And** 실행 시간이 100ms 이내여야 한다

---

## Scenario 17: All 9 Hooks Working

**Given** moai_hooks.py가 배포되어 있다  
**When** 9개 hook 이벤트를 순차적으로 테스트한다:
1. SessionStart
2. SessionEnd
3. PreToolUse
4. PostToolUse
5. UserPromptSubmit
6. Notification
7. Stop
8. SubagentStop
9. PreCompact

**Then** 각 hook이 유효한 JSON을 반환해야 한다  
**And** 각 hook이 `success: true`를 반환해야 한다  
**And** 어떤 hook도 예외를 발생시키지 않아야 한다  
**And** 모든 hook이 HookResult 타입을 반환해야 한다

---

## Scenario 18: Performance - SessionStart

**Given** 표준 크기의 프로젝트 (SPEC 10개, 파일 100개)  
**When** SessionStart 이벤트가 트리거된다  
**Then** 실행 시간이 500ms 이내여야 한다  
**And** 메모리 사용량이 50MB 이하여야 한다  
**And** Git 정보 수집, SPEC 진행률 계산, 언어 감지가 모두 완료되어야 한다

---

## Scenario 19: Performance - Other Hooks

**Given** moai_hooks.py가 실행 중이다  
**When** SessionEnd, PreToolUse, PostToolUse, Notification, Stop, SubagentStop, PreCompact 이벤트가 트리거된다  
**Then** 각 hook의 실행 시간이 100ms 이내여야 한다  
**And** 메모리 사용량이 30MB 이하여야 한다  
**And** 빠른 응답 시간을 유지해야 한다

---

## Scenario 20: stdin/stdout Protocol

**Given** stdin으로 JSON payload가 입력된다  
**And** 입력: `{"event":"SessionStart","cwd":"/path/to/project"}`  
**When** moai_hooks.py가 실행된다  
**Then** stdout으로 JSON 결과가 출력되어야 한다  
**And** JSON이 indent=2로 포맷팅되어야 한다  
**And** 유효한 JSON 형식이어야 한다 (`json.loads()` 성공)  
**And** `success`, `message`, `data` 필드가 포함되어야 한다

---

## Scenario 21: uv run Execution

**Given** uv가 설치되어 있다  
**And** Python 3.14가 설치되어 있거나 uv가 자동으로 설치할 수 있다  
**When** `uv run --python 3.14 .claude/hooks/moai_hooks.py SessionStart` 명령을 실행한다  
**Then** Python 3.14가 자동으로 설치되어야 한다 (없으면)  
**And** moai_hooks.py가 정상 실행되어야 한다  
**And** 배너가 출력되어야 한다  
**And** exit code 0으로 종료되어야 한다

---

## Scenario 22: JIT Context - 파일 없을 때

**Given** 사용자가 "/alfred:1-spec" 프롬프트를 입력한다  
**And** .moai/memory/spec-metadata.md 파일이 존재하지 않는다  
**When** UserPromptSubmit 이벤트가 트리거된다  
**Then** relevant_docs가 빈 배열이어야 한다  
**And** 에러가 발생하지 않아야 한다  
**And** `success: true`를 반환해야 한다  
**And** 실행 시간이 100ms 이내여야 한다

---

## Scenario 23: Multi-language Project

**Given** Python, TypeScript, Go가 혼합된 프로젝트  
**And** pyproject.toml, tsconfig.json, go.mod 파일이 모두 존재한다  
**When** SessionStart 이벤트가 트리거된다  
**Then** 3개 언어가 모두 감지되어야 한다  
**And** 배너에 "Languages: Python, TypeScript, Go"가 표시되어야 한다  
**And** 순서는 LANGUAGE_SIGNATURES 딕셔너리 순서를 따라야 한다

---

## Scenario 24: SPEC 디렉토리 없을 때

**Given** .moai/specs/ 디렉토리가 존재하지 않는다  
**When** SessionStart 이벤트가 트리거된다  
**Then** SPEC 진행률이 "0/0 (0%)"로 표시되어야 한다  
**And** 에러가 발생하지 않아야 한다  
**And** 배너가 정상 출력되어야 한다

---

## Scenario 25: Read-Only Operations

**Given** moai_hooks.py가 실행된다  
**When** 모든 9개 hook을 테스트한다  
**Then** 어떤 파일도 수정하지 않아야 한다  
**And** 새로운 파일을 생성하지 않아야 한다  
**And** Git 저장소를 변경하지 않아야 한다  
**And** 모든 작업이 읽기 전용이어야 한다

---

## Scenario 26: No Network Calls

**Given** moai_hooks.py가 실행된다  
**When** 네트워크 모니터링 도구로 확인한다  
**Then** 어떤 외부 API도 호출하지 않아야 한다  
**And** 네트워크 트래픽이 없어야 한다  
**And** 모든 작업이 로컬 파일 시스템에서만 수행되어야 한다

---

## Scenario 27: Glob Pattern Matching

**Given** 프로젝트에 `*.py` 파일이 여러 개 있다  
**When** LanguageDetector가 Python을 감지한다  
**Then** glob 패턴 `**/*.py`가 재귀적으로 매칭되어야 한다  
**And** 모든 하위 디렉토리의 .py 파일이 감지되어야 한다  
**And** 실행 시간이 100ms 이내여야 한다

---

## Scenario 28: Type Hints Validation

**Given** moai_hooks.py 파일  
**When** mypy로 타입 체크를 실행한다 (`mypy moai_hooks.py`)  
**Then** 타입 에러가 없어야 한다  
**And** 모든 함수/메서드에 타입 힌트가 있어야 한다  
**And** TypedDict가 올바르게 정의되어야 한다  
**And** Optional 타입이 명확히 표시되어야 한다

---

## Scenario 29: Docstring Completeness

**Given** moai_hooks.py 파일  
**When** 모든 클래스와 메서드를 확인한다  
**Then** 모든 public 클래스에 docstring이 있어야 한다  
**And** 모든 public 메서드에 docstring이 있어야 한다  
**And** docstring에 Args, Returns, Examples가 포함되어야 한다 (필요 시)  
**And** 사용법이 명확해야 한다

---

## Scenario 30: Test Coverage

**Given** moai_hooks.py 파일  
**And** tests/unit/test_moai_hooks.py 테스트 파일  
**When** pytest --cov=moai_hooks를 실행한다  
**Then** 테스트 커버리지가 85% 이상이어야 한다  
**And** 모든 핵심 함수가 테스트되어야 한다  
**And** edge case가 테스트되어야 한다 (타임아웃, 파일 없음, 잘못된 JSON 등)  
**And** 모든 테스트가 통과해야 한다

---

**작성자**: @Goos
**작성일**: 2025-10-13
**총 시나리오**: 30개
