---
id: DOC-INTEGRATION-001
version: "1.0"
status: "draft"
title: 코드-문서 자동 연동 시스템
description: 코드 스캐닝, AST 분석, 자동 문서 생성, TAG 체인 연동을 통합한 완전한 문서 자동화 시스템
author: "spec-builder"
date: "2025-11-05"
tags: ["documentation", "automation", "integration", "ast", "code-analysis"]
depends_on: ["DOC-ONLINE-001", "DOC-VISUAL-001", "DOC-TABLE-001"]
---


## 환경 (Environment)

- **프로젝트**: MoAI-ADK 문서 시스템
- **파이썬 버전**: 3.13+
- **핵심 의존성**:
  - `tree-sitter` (AST 분석)
  - `networkx` (콜 그래프 분석)
  - `jinja2` (문서 템플릿)
  - `pydantic` (데이터 모델링)
- **연동 모듈**: doc_scanner, doc_generator, tag_manager, doc_server

## 가정 (Assumptions)

- 코드베이스는 AST 분석이 가능한 구조적 프로그래밍 언어로 작성됨
- 기존 DOC-ONLINE-001, DOC-VISUAL-001, DOC-TABLE-001 시스템이 완료되어 있음
- 소스 코드에는 최소한의 문서화 주석이 포함되어 있음
- 동기화된 개발 환경에서 코드 변경과 문서 업데이트가 주기적으로 발생함

## 요구사항 (Requirements)

### 기능 요구사항

1. **자동 코드 스캐닝 시스템**
   - 프로젝트 코드베이스 전체를 주기적으로 스캔
   - 변경된 파일만 선택적으로 재분석 (증분 업데이트)
   - 지원 언어: Python, TypeScript, JavaScript, Go, Rust
   - 파일 시스템 변경 감지 (watchdog 기반)

2. **AST 기반 구조 분석**
   - 함수, 클래스, 모듈 수준의 AST 파싱
   - 의존성 관계 및 콜 그래프 자동 생성
   - 복잡도 메트릭 계산 (순환 복잡도, 결합도)
   - 코드 시그니처 및 타입 정보 추출

3. **지능형 문서 생성**
   - AST 정보를 기반으로 API 문서 자동 생성
   - 코드 예제 및 사용법 자동 추출
   - 타입 정보를 활용한 매개변수/반환값 문서화
   - Markdown 및 HTML 다중 포맷 지원

4. **TAG 체인 자동 연동**
   - TAG 체인 무결성 검증 및 오류 보고
   - 문서 내 TAG 참조 자동 하이퍼링크화
   - 고아 TAG(Orphaned Tags) 자동 감지

5. **실시간 동기화 엔진**
   - 코드 변경 시 문서 자동 업데이트
   - 문서 변경 시 관련 코드 위치 자동 표시
   - 충돌 감지 및 해결 전략 제공
   - 변경 히스토리 추적 및 롤백 기능

### 비기능 요구사항

1. **성능**
   - 전체 프로젝트 분석: 5분 이내 (10,000파일 기준)
   - 증분 업데이트: 30초 이내 (100파일 변경 기준)
   - 실시간 감지: 2초 이내 응답

2. **확장성**
   - 100,000+ 파일 처리 가능
   - 분산 처리 지원 (멀티프로세싱)
   - 플러그인 아키텍처 (새로운 언어 지원)

3. **정확성**
   - AST 파싱 정확도: 99.9%
   - 의존성 분석 정확도: 95%
   - TAG 체인 일관성: 100%

4. **사용성**
   - 단일 명령어로 전체 시스템 실행
   - 직관적인 설정 파일 구조
   - 상세한 진단 로그 및 오류 보고

## 명세 (Specifications)

### 1. 코드 스캐너 모듈 (doc_scanner)

```python
# 1.1 메인 스캐너 인터페이스
class DocumentScanner:
    def __init__(self, config: ScannerConfig):
        self.config = config
        self.parser_factory = ParserFactory()
        self.tag_detector = TagDetector()

    def scan_project(self, project_path: Path) -> ScanResult:
        """
        전체 프로젝트 스캔 수행

        Event: 사용자가 프로젝트 스캔 요청
        Action: 프로젝트 구조 분석 및 파일 수집
        Response: ScanResult 객체 반환 (파일 목록, 메타데이터)
        State: 준비 상태 (스캔 데이터 메모리 로드)
        """

    def scan_incremental(self, changes: List[FileChange]) -> IncrementalResult:
        """
        증분 스캔 수행

        Event: 파일 시스템 변경 감지
        Action: 변경된 파일만 선택적 스캔
        Response: IncrementalResult 객체 반환
        State: 부분 업데이트 상태
        """

# 1.2 언어별 파서 팩토리
class ParserFactory:
    def get_parser(self, language: str) -> LanguageParser:
        """
        언어별 AST 파서 반환

        Event: 언어 타입 식별 요청
        Action: 해당 언어 파서 인스턴스 생성
        Response: LanguageParser 객체 반환
        State: 파서 준비 상태
        """
```

### 2. AST 분석기 모듈 (ast_analyzer)

```python
# 2.1 AST 분석기 핵심
class ASTAnalyzer:
    def __init__(self, language: str):
        self.language = language
        self.parser = get_tree_sitter_parser(language)
        self.dependency_tracker = DependencyTracker()

    def analyze_file(self, file_path: Path) -> FileAnalysis:
        """
        단일 파일 AST 분석

        Event: 파일 분석 요청
        Action: AST 파싱 및 구조 추출
        Response: FileAnalysis 객체 반환 (함수, 클래스, 의존성)
        State: 분석 완료 상태
        """

    def build_call_graph(self, analyses: List[FileAnalysis]) -> CallGraph:
        """
        콜 그래프 생성

        Event: 여러 파일 분석 결과 입력
        Action: 함수 호출 관계 분석 및 그래프 생성
        Response: CallGraph 객체 반환
        State: 의존성 맵핑 완료
        """

# 2.2 복잡도 분석기
class ComplexityAnalyzer:
    def calculate_complexity(self, node: ASTNode) -> ComplexityMetrics:
        """
        코드 복잡도 계산

        Event: 복잡도 분석 요청
        Action: 순환 복잡도, 결합도, 응집도 계산
        Response: ComplexityMetrics 객체 반환
        State: 메트릭 계산 완료
        """
```

### 3. 문서 생성기 모듈 (doc_generator)

```python
# 3.1 문서 생성기 메인
class DocumentGenerator:
    def __init__(self, template_engine: Jinja2Templates):
        self.template_engine = template_engine
        self.formatter = DocumentFormatter()
        self.tag_integrator = TagIntegrator()

    def generate_docs(self, analysis: ProjectAnalysis) -> GeneratedDocs:
        """
        분석 결과로 문서 생성

        Event: 분석 완료 데이터 수신
        Action: 템플릿 기반 문서 생성
        Response: GeneratedDocs 객체 반환 (HTML, Markdown)
        State: 문서 생성 완료
        """

    def integrate_tags(self, docs: GeneratedDocs, tag_chain: TagChain) -> EnhancedDocs:
        """
        TAG 체인 연동

        Event: 생성된 문서와 TAG 체인 연동 요청
        Action: TAG 참조 검색 및 하이퍼링크 생성
        Response: EnhancedDocs 객체 반환 (TAG 연동 완료)
        State: TAG 통합 완료
        """

# 3.2 템플릿 엔진
class DocumentFormatter:
    def format_api_doc(self, function_info: FunctionInfo) -> APIDocument:
        """
        API 문서 포맷팅

        Event: 함수 정보 포맷팅 요청
        Action: 시그니처, 매개변수, 반환값 포맷팅
        Response: APIDocument 객체 반환
        State: 포맷팅 완료
        """
```

### 4. TAG 연동 모듈 (tag_integrator)

```python
# 4.1 TAG 감지기
class TagDetector:
    def scan_tags_in_code(self, file_path: Path) -> List[TagReference]:
        """
        코드 내 TAG 참조 스캔

        Event: TAG 스캔 요청
        Response: TagReference 리스트 반환
        State: TAG 감지 완료
        """

    def validate_tag_chain(self, tag_refs: List[TagReference]) -> TagValidationResult:
        """
        TAG 체인 무결성 검증

        Event: TAG 무결성 검증 요청
        Action: 참조 존재 여부 및 형식 검증
        Response: TagValidationResult 객체 반환
        State: 검증 완료
        """

# 4.2 TAG 연동기
class TagIntegrator:
    def link_documentation(self, docs: GeneratedDocs, tags: List[TagReference]) -> LinkedDocs:
        """
        문서에 TAG 연동

        Event: 문서-TAG 연동 요청
        Action: TAG 참조에 하이퍼링크 생성
        Response: LinkedDocs 객체 반환
        State: 연동 완료
        """
```

### 5. 동기화 엔진 모듈 (sync_engine)

```python
# 5.1 동기화 코디네이터
class SyncCoordinator:
    def __init__(self, scanner: DocumentScanner, generator: DocumentGenerator):
        self.scanner = scanner
        self.generator = generator
        self.conflict_resolver = ConflictResolver()

    def sync_full_project(self) -> SyncResult:
        """
        전체 프로젝트 동기화

        Event: 전체 동기화 요청
        Action: 스캔 → 분석 → 생성 → 연동 파이프라인 실행
        Response: SyncResult 객체 반환
        State: 동기화 완료
        """

    def sync_incremental(self, changes: List[FileChange]) -> IncrementalSyncResult:
        """
        증분 동기화

        Event: 변경 감지 시 자동 동기화
        Action: 변경 파일만 대상으로 동기화
        Response: IncrementalSyncResult 객체 반환
        State: 부분 동기화 완료
        """

# 5.2 충돌 해결기
class ConflictResolver:
    def resolve_conflicts(self, conflicts: List[SyncConflict]) -> ResolutionResult:
        """
        동기화 충돌 해결

        Event: 충돌 감지
        Action: 충돌 해결 전략 적용 (자동/수동)
        Response: ResolutionResult 객체 반환
        State: 충돌 해결 완료
        """
```

### 6. 웹 서비스 연동 (web_integration)

```python
# 6.1 API 엔드포인트
class DocIntegrationAPI:
    def __init__(self, sync_coordinator: SyncCoordinator):
        self.sync_coordinator = sync_coordinator
        self.status_monitor = StatusMonitor()

    async def trigger_full_sync(self) -> SyncStatus:
        """
        전체 동기화 트리거 API

        Event: HTTP POST /api/sync/full 요청
        Action: 비동기 전체 동기화 실행
        Response: SyncStatus 객체 반환 (진행상황)
        State: 동기화 진행 중
        """

    async def get_sync_status(self) -> SyncStatus:
        """
        동기화 상태 조회 API

        Event: HTTP GET /api/sync/status 요청
        Action: 현재 동기화 상태 반환
        Response: SyncStatus 객체 반환
        State: 상태 조회 완료
        """
```

## 통합 추적성 (Traceability)

- **의존 모듈**: doc_scanner, ast_analyzer, doc_generator, tag_integrator, sync_engine