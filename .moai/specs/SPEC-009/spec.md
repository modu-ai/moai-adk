# SPEC-009: TAG 시스템 SQLite 마이그레이션

**@SPEC:SPEC-009-STARTED** ← 시작점  
**@REQ:TAG-PERFORMANCE-001** ← 성능 요구사항  
**@DESIGN:SQLITE-MIGRATION-001** ← 설계 결정  
**@TASK:DATABASE-SCHEMA-001** ← 구현 태스크

---

## Environment (환경 및 전제 조건)

### @TECH:RUNTIME-ENV-001 실행 환경
- **Python 버전**: ≥3.11 (SQLite3 표준 라이브러리 포함)
- **운영체제**: Windows, macOS, Linux (크로스 플랫폼)
- **메모리**: 최소 100MB 여유 공간 (대용량 JSON 로딩용)
- **디스크**: 최소 500MB 여유 공간 (백업 및 마이그레이션용)

### @TECH:DEPENDENCIES-001 기술 종속성
- **SQLite3**: Python 표준 라이브러리 (외부 종속성 없음)
- **기존 JSON 시스템**: `src/moai_adk/core/tag_system/` 모듈
- **파일 시스템 접근**: `.moai/indexes/` 디렉토리 읽기/쓰기 권한
- **Git 통합**: TAG 추적을 위한 Git 저장소 접근

### @STRUCT:EXISTING-SYSTEM-001 현재 시스템 상태
- **tags.json 크기**: 4,747줄, 136KB, 441개 TAG, 770개 참조
- **성장 예측**: 프로젝트 확장 시 10배 증가 → 1.36MB, 4,410개 TAG
- **성능 병목**: 메모리 로딩, 파일 I/O, Git diff, 검색 속도
- **기존 API**: JSON 기반 인터페이스 완전 호환 유지 필요

---

## Assumptions (가정 사항)

### @VISION:PERFORMANCE-GOAL-001 성능 목표
- **현재 대비 10배 성능 향상**: 검색, 삽입, 업데이트, 삭제 작업
- **메모리 사용량 50% 감소**: 전체 인덱스를 메모리에 로딩하지 않음
- **파일 크기 30% 감소**: SQLite의 효율적인 바이너리 저장
- **Git diff 크기 90% 감소**: 바이너리 파일로 변경되어 diff 최소화

### @REQ:COMPATIBILITY-001 호환성 가정
- **기존 JSON API 100% 호환**: 코드 변경 없이 투명한 전환
- **점진적 마이그레이션**: 무중단 서비스 유지
- **롤백 가능성**: 언제든지 JSON으로 복원 가능
- **외부 도구 호환성**: `validate_tags.py` 등 기존 스크립트 동작 유지

### @TECH:MIGRATION-ASSUMPTIONS-001 마이그레이션 가정
- **데이터 무결성**: 마이그레이션 중 TAG 데이터 손실 없음
- **원자적 전환**: 전환 성공 또는 실패, 중간 상태 없음
- **백업 정책**: 마이그레이션 전 자동 백업 생성
- **검증 가능성**: 마이그레이션 결과 검증 및 비교 도구 제공

---

## Requirements (기능 요구사항)

### @REQ:PERFORMANCE-001 성능 요구사항

**R1. 검색 성능**
- **현재**: O(n) 선형 검색
- **목표**: O(log n) 인덱스 검색
- **측정**: 1,000개 TAG 검색 시 < 10ms

**R2. 삽입 성능**
- **현재**: 전체 JSON 재작성
- **목표**: 개별 레코드 삽입
- **측정**: TAG 추가 시 < 5ms

**R3. 메모리 사용량**
- **현재**: 전체 인덱스 메모리 로딩
- **목표**: 필요 시에만 부분 로딩
- **측정**: 메모리 사용량 < 50MB

### @REQ:FUNCTIONALITY-001 기능 요구사항

**R4. 호환성**
- 기존 JSON API 100% 호환성 유지
- `get_tags()`, `add_tag()`, `remove_tag()` 메서드 시그니처 유지
- JSON 출력 포맷 완전 일치

**R5. 마이그레이션**
- JSON → SQLite 변환 도구
- SQLite → JSON 변환 도구 (롤백용)
- 데이터 무결성 검증 도구
- 성능 비교 벤치마크 도구

**R6. 확장성**
- 10,000개 TAG까지 선형 성능 유지
- 다중 프로젝트 지원 (데이터베이스 분리)
- 백업 및 복원 기능

### @REQ:QUALITY-001 품질 요구사항

**R7. 데이터 무결성**
- ACID 트랜잭션 지원
- 외래키 제약 조건
- 데이터 타입 검증

**R8. 오류 처리**
- 데이터베이스 잠금 상황 처리
- 디스크 공간 부족 상황 대응
- 손상된 데이터베이스 복구

---

## Specifications (상세 명세)

### @DESIGN:DATABASE-SCHEMA-001 데이터베이스 스키마

```sql
-- @DATA:TAGS-TABLE-001 TAG 마스터 테이블
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tag_key TEXT UNIQUE NOT NULL,           -- @REQ:USER-AUTH-001
    tag_type TEXT NOT NULL,                 -- REQ, DESIGN, TASK, TEST
    tag_id TEXT NOT NULL,                   -- USER-AUTH-001
    description TEXT,                       -- 선택적 설명
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- @DATA:REFERENCES-TABLE-001 TAG 참조 테이블
CREATE TABLE tag_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tag_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,                -- 상대 경로
    line_number INTEGER NOT NULL,
    context TEXT,                           -- 해당 라인 내용
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- @PERF:INDEX-001 성능 최적화 인덱스
CREATE INDEX idx_tag_key ON tags(tag_key);
CREATE INDEX idx_tag_type ON tags(tag_type);
CREATE INDEX idx_tag_id ON tags(tag_id);
CREATE INDEX idx_file_path ON tag_references(file_path);
CREATE INDEX idx_line_number ON tag_references(line_number);
CREATE INDEX idx_tag_reference ON tag_references(tag_id, file_path);

-- @DATA:VERSION-TABLE-001 버전 관리
CREATE TABLE schema_version (
    version TEXT PRIMARY KEY,
    migrated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    json_backup_path TEXT
);
```

### @API:ADAPTER-LAYER-001 호환성 어댑터

```python
# @FEATURE:TAG-ADAPTER-001 JSON API 호환성 레이어
class TagIndexAdapter:
    """기존 JSON API를 SQLite 백엔드로 투명하게 변환"""
    
    def __init__(self, db_path: Path):
        self.db_manager = TagDatabaseManager(db_path)
    
    def get_tags(self) -> Dict[str, Any]:
        """기존 JSON 포맷으로 모든 TAG 반환"""
        # SQLite → JSON 변환 로직
        pass
    
    def add_tag(self, tag_key: str, reference: Dict) -> bool:
        """기존 인터페이스로 TAG 추가"""
        # JSON 파라미터 → SQLite 삽입
        pass
    
    def remove_tag(self, tag_key: str) -> bool:
        """TAG 및 모든 참조 삭제"""
        pass
```

### @TASK:MIGRATION-TOOL-001 마이그레이션 도구

```python
# @FEATURE:MIGRATION-001 양방향 마이그레이션
class TagMigrationTool:
    """JSON ↔ SQLite 변환 도구"""
    
    def migrate_json_to_sqlite(self, json_path: Path, db_path: Path) -> bool:
        """JSON → SQLite 마이그레이션"""
        # 1. JSON 파일 파싱 및 검증
        # 2. SQLite 스키마 생성
        # 3. 데이터 변환 및 삽입
        # 4. 무결성 검증
        pass
    
    def migrate_sqlite_to_json(self, db_path: Path, json_path: Path) -> bool:
        """SQLite → JSON 복원 (롤백용)"""
        pass
    
    def verify_migration(self, json_path: Path, db_path: Path) -> bool:
        """마이그레이션 결과 검증"""
        pass
```

### @PERF:BENCHMARK-001 성능 측정 도구

```python
# @TEST:PERFORMANCE-001 성능 벤치마크
class TagPerformanceBenchmark:
    """JSON vs SQLite 성능 비교"""
    
    def benchmark_search(self, tag_count: int) -> Dict[str, float]:
        """검색 성능 측정"""
        pass
    
    def benchmark_insert(self, insert_count: int) -> Dict[str, float]:
        """삽입 성능 측정"""
        pass
    
    def benchmark_memory_usage(self) -> Dict[str, int]:
        """메모리 사용량 측정"""
        pass
```

---

## @TODO:TRACEABILITY-001 추적성 태그 체인

```
@SPEC:SPEC-009-STARTED
├── @REQ:TAG-PERFORMANCE-001
│   ├── @DESIGN:SQLITE-MIGRATION-001
│   ├── @TASK:DATABASE-SCHEMA-001
│   └── @TEST:PERFORMANCE-BENCHMARK-001
├── @REQ:COMPATIBILITY-001
│   ├── @DESIGN:ADAPTER-PATTERN-001
│   ├── @TASK:API-COMPATIBILITY-001
│   └── @TEST:COMPATIBILITY-TEST-001
└── @REQ:MIGRATION-TOOL-001
    ├── @DESIGN:BIDIRECTIONAL-MIGRATION-001
    ├── @TASK:MIGRATION-IMPLEMENTATION-001
    └── @TEST:MIGRATION-VERIFICATION-001
```

---

## 변경 영향 분석

### @STRUCT:IMPACT-ANALYSIS-001 영향받는 모듈

1. **직접 영향**:
   - `src/moai_adk/core/tag_system/index_manager.py` (수정)
   - `.moai/scripts/validate_tags.py` (호환성 유지)

2. **간접 영향**:
   - `/moai:3-sync` 명령어 (성능 개선)
   - Git diff 크기 (바이너리 파일로 변경)

3. **영향 없음**:
   - 기존 Python API 사용자 (투명한 변경)
   - Claude Code 에이전트 (API 호환성 유지)

### @DEBT:MIGRATION-DEBT-001 기술 부채 해결
- **현재**: JSON 파일 크기로 인한 성능 저하
- **해결**: SQLite 인덱스를 통한 O(log n) 검색
- **추가 이익**: ACID 트랜잭션, 데이터 무결성, 동시 접근 제어

---

**완료 조건**: 모든 기존 기능이 10배 빠른 성능으로 동작하며, 기존 코드 변경 없이 투명하게 전환 완료