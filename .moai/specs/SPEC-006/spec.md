---
spec_id: SPEC-006
status: active
priority: high
dependencies: []
tags:
  - tag-system
  - traceability
  - 16-core
  - indexing
  - automation
  - requirements-tracking
---

# SPEC-006: 16-Core TAG 추적성 시스템 완성

## @REQ:TRACEABILITY-006 프로젝트 컨텍스트

### 배경

MoAI-ADK는 @REQ:PROBLEM-001에서 제기된 "TAG 추적성 공백" 문제를 해결하기 위해 16-Core TAG 시스템을 도입했습니다. 현재 `@REQ → @DESIGN → @TASK → @TEST` 체인의 완전한 추적과 scope creep 방지가 핵심 목표로 설정되어 있습니다.

### 문제 정의

- **현재 상태**: TAG 체인 연결성 검증 시스템 부재로 요구사항 추적 실패
- **핵심 문제**: `.moai/indexes/tags.json` 수동 관리로 인한 동기화 지연
- **비즈니스 영향**: scope creep 발생 및 요구사항-구현 간 불일치로 인한 품질 저하

### 목표

1. `@REQ → @DESIGN → @TASK → @TEST` 체인 자동 추적 및 검증
2. `.moai/indexes/tags.json` 실시간 갱신 시스템 구현
3. 추적성 리포트 자동 생성으로 완전한 요구사항-구현 매핑 제공

## @DESIGN:TAG-SYSTEM-006 환경 및 가정사항

### Environment (환경)

- **TAG 스키마**: 16-Core TAG 시스템 (Primary, Steering, Implementation, Quality)
- **저장소**: `.moai/indexes/tags.json` 중앙 인덱스
- **검색 범위**: 소스 코드, 문서, 커밋 메시지, Issue/PR
- **연동**: Claude Code + Git 작업 자동 동기화

### Assumptions (가정사항)

- 모든 개발자가 TAG 작성 규칙을 이해하고 있음
- Git 커밋 메시지에 TAG가 포함됨
- 문서와 코드에 일관된 TAG 명명 규칙 적용
- `/moai:3-sync` 명령어를 통한 정기적 동기화 수행

## @TASK:IMPLEMENT-006 요구사항 명세

### R1. TAG 체인 자동 추적 및 검증 시스템

**WHEN** 새로운 TAG가 문서나 코드에 추가될 때,
**THE SYSTEM SHALL** 자동으로 TAG 체인의 연결성을 검증하고 무결성을 확인해야 함

**상세 요구사항:**

- Primary Chain 연결성 검증: @REQ → @DESIGN → @TASK → @TEST
- 순환 참조 및 고아 TAG 감지
- TAG 명명 규칙 준수 여부 자동 검사
- 체인 완결성 검증 (시작점부터 종료점까지)

### R2. `.moai/indexes/tags.json` 실시간 갱신

**WHEN** 파일이 수정되거나 Git 작업이 수행될 때,
**THE SYSTEM SHALL** 자동으로 TAG 인덱스를 스캔하고 업데이트해야 함

**상세 요구사항:**

- 파일 변경 감지 시 자동 TAG 추출
- Git 커밋 메시지에서 TAG 파싱
- 중복 TAG 제거 및 정규화
- 인덱스 파일 JSON 스키마 검증

### R3. 추적성 리포트 자동 생성

**WHEN** 추적성 상태 확인이 요청될 때,
**THE SYSTEM SHALL** 완전한 요구사항-구현 매핑 리포트를 생성해야 함

**상세 요구사항:**

- TAG 체인별 구현 상태 매트릭스
- 누락된 TAG 연결 식별 및 리포트
- 구현 완료율 및 커버리지 계산
- 시각적 추적성 다이어그램 생성

### R4. TAG 체인 무결성 검사 도구

**WHEN** 프로젝트 상태 검증이 필요할 때,
**THE SYSTEM SHALL** TAG 체인의 무결성을 종합적으로 검사하고 수정 방안을 제시해야 함

**상세 요구사항:**

- 연결 끊어진 TAG 체인 식별
- 중복되거나 모호한 TAG 감지
- 누락된 구현 영역 식별
- 자동 수정 제안 및 가이드 제공

## @TEST:ACCEPTANCE-006 Acceptance Criteria

### AC1. TAG 체인 자동 추적

**Given** 새로운 요구사항 `@REQ:USER-AUTH-007`이 문서에 추가될 때
**When** TAG 추적 시스템이 실행되면
**Then** 연결된 `@DESIGN:JWT-007`, `@TASK:API-007`, `@TEST:UNIT-007` TAG의 존재 여부가 자동으로 검증되어야 함

**Given** TAG 체인에서 연결이 끊어진 부분이 있을 때
**When** 무결성 검사를 실행하면
**Then** 구체적인 누락 지점과 수정 방법이 제시되어야 함

### AC2. 실시간 인덱스 갱신

**Given** 소스 코드에 새로운 TAG `@FEATURE:LOGIN-001`이 추가될 때
**When** 파일 저장 후 자동 스캔이 실행되면
**Then** `.moai/indexes/tags.json`에 해당 TAG가 올바른 카테고리로 분류되어 추가되어야 함

**Given** Git 커밋 메시지에 TAG가 포함될 때
**When** 커밋이 완료되면
**Then** 커밋 해시와 함께 TAG 인덱스가 자동으로 업데이트되어야 함

### AC3. 추적성 리포트 생성

**Given** 프로젝트에 여러 TAG 체인이 존재할 때
**When** 추적성 리포트를 요청하면
**Then** 요구사항별 구현 완료 상태가 매트릭스 형태로 표시되어야 함

**Given** 누락된 구현이 있는 TAG 체인이 존재할 때
**When** 리포트를 확인하면
**Then** 미완성 영역이 명확히 식별되고 우선순위가 제시되어야 함

### AC4. 무결성 검사 및 수정 제안

**Given** TAG 체인에 중복이나 순환 참조가 있을 때
**When** 무결성 검사를 실행하면
**Then** 문제가 있는 TAG들이 식별되고 구체적인 수정 방안이 제공되어야 함

**Given** 고아 TAG (연결되지 않은 TAG)가 발견될 때
**When** 검사 결과를 확인하면
**Then** 해당 TAG를 적절한 체인에 연결하는 방법이 제안되어야 함

## 범위 및 모듈

### In Scope

- 16-Core TAG 시스템 완전 구현
- TAG 체인 자동 추적 및 검증 엔진
- 실시간 TAG 인덱스 관리 시스템
- 추적성 리포트 생성 도구
- TAG 무결성 검사 및 수정 제안 시스템
- Claude Code + Git 워크플로우 연동

### Out of Scope

- TAG 시각화 GUI 인터페이스 (터미널 기반 리포트만)
- 외부 이슈 트래커 연동 (GitHub Issues 제외)
- TAG 버전 관리 (단순 추가/수정만 지원)
- 다국어 TAG 지원 (영어만 지원)

## 기술 노트

### 구현 기술

- **TAG 파싱**: 정규표현식 + AST 파싱
- **인덱스 관리**: JSON + SQLite 하이브리드
- **파일 감시**: watchdog 라이브러리
- **리포트 생성**: 템플릿 기반 Markdown/HTML 출력

### 의존성

- **watchdog**: 파일 시스템 변경 감지
- **jsonschema**: TAG 인덱스 스키마 검증
- **gitpython**: Git 커밋 메시지 파싱
- **jinja2**: 리포트 템플릿 렌더링

### 16-Core TAG 분류 체계

#### Primary Chain (4개)

- **@REQ**: 요구사항 정의
- **@DESIGN**: 설계 결정사항
- **@TASK**: 구현 작업
- **@TEST**: 테스트 검증

#### Steering (4개)

- **@VISION**: 제품 비전
- **@STRUCT**: 구조 설계
- **@TECH**: 기술 선택
- **@ADR**: 아키텍처 결정

#### Implementation (4개)

- **@FEATURE**: 기능 구현
- **@API**: API 설계
- **@UI**: 사용자 인터페이스
- **@DATA**: 데이터 모델

#### Quality (4개)

- **@PERF**: 성능 최적화
- **@SEC**: 보안 강화
- **@DEBT**: 기술 부채
- **@DOCS**: 문서화

### 성능 고려사항

- TAG 인덱스 증분 업데이트 (전체 스캔 방지)
- 파일 변경 감지 시 배치 처리
- 대용량 프로젝트에서의 스캔 성능 최적화
- 메모리 효율적인 TAG 저장 구조

### 보안 고려사항

- TAG 인덱스 파일 무결성 검증
- 악의적인 TAG 패턴 필터링
- 민감정보 포함 TAG 마스킹

## 추적성

### 연결된 요구사항

- @REQ:PROBLEM-001: TAG 추적성 공백 해결
- @VISION:STRATEGY-001: 16-Core TAG 체계로 완전한 추적성 보장
- @REQ:SUCCESS-001: scope creep 방지 및 요구사항 추적성 확보

### 구현 우선순위

1. TAG 체인 자동 추적 시스템 (High) - 핵심 기능
2. 실시간 인덱스 갱신 (High) - 동기화 필수
3. 추적성 리포트 생성 (Medium) - 가시성 제공
4. 무결성 검사 도구 (Medium) - 품질 보장

### 테스트 전략

- 단위 테스트: TAG 파싱 및 체인 검증 로직 테스트
- 통합 테스트: 파일 변경부터 인덱스 갱신까지 전체 플로우 테스트
- E2E 테스트: 실제 프로젝트에서 TAG 추적성 전 과정 검증

## TAG 시스템 구현 예시

### TAG 파싱 엔진

```python
import re
from typing import Dict, List, Set
from dataclasses import dataclass

@dataclass
class TagInfo:
    category: str          # REQ, DESIGN, TASK, TEST 등
    identifier: str        # USER-AUTH-001
    file_path: str         # 발견된 파일 경로
    line_number: int       # 라인 번호
    context: str           # 주변 컨텍스트

class TagParser:
    TAG_PATTERN = re.compile(r'@([A-Z]+):([A-Z0-9-]+)')

    VALID_CATEGORIES = {
        # Primary Chain
        'REQ', 'DESIGN', 'TASK', 'TEST',
        # Steering
        'VISION', 'STRUCT', 'TECH', 'ADR',
        # Implementation
        'FEATURE', 'API', 'UI', 'DATA',
        # Quality
        'PERF', 'SEC', 'DEBT', 'DOCS'
    }

    def parse_file(self, file_path: str) -> List[TagInfo]:
        """파일에서 모든 TAG 추출"""
        tags = []

        with open(file_path, 'r', encoding='utf-8') as f:
            for line_num, line in enumerate(f, 1):
                matches = self.TAG_PATTERN.findall(line)

                for category, identifier in matches:
                    if category in self.VALID_CATEGORIES:
                        tags.append(TagInfo(
                            category=category,
                            identifier=identifier,
                            file_path=file_path,
                            line_number=line_num,
                            context=line.strip()
                        ))

        return tags
```

### TAG 체인 검증

```python
class TagChainValidator:
    PRIMARY_CHAIN = ['REQ', 'DESIGN', 'TASK', 'TEST']

    def validate_chain(self, tags: List[TagInfo]) -> Dict[str, List[str]]:
        """TAG 체인 유효성 검사"""
        issues = {
            'missing_links': [],
            'orphaned_tags': [],
            'broken_chains': []
        }

        # 식별자별로 TAG 그룹핑
        tag_groups = {}
        for tag in tags:
            base_id = self._extract_base_id(tag.identifier)
            if base_id not in tag_groups:
                tag_groups[base_id] = {}
            tag_groups[base_id][tag.category] = tag

        # 각 그룹의 체인 완결성 검사
        for base_id, group in tag_groups.items():
            chain_issues = self._validate_primary_chain(base_id, group)
            issues['broken_chains'].extend(chain_issues)

        return issues

    def _extract_base_id(self, identifier: str) -> str:
        """식별자에서 기본 ID 추출 (예: USER-AUTH-001 → USER-AUTH)"""
        parts = identifier.split('-')
        return '-'.join(parts[:-1]) if len(parts) > 1 else identifier

    def _validate_primary_chain(self, base_id: str, group: Dict[str, TagInfo]) -> List[str]:
        """Primary Chain 검증"""
        issues = []

        for i, category in enumerate(self.PRIMARY_CHAIN):
            if category not in group:
                issues.append(f"Missing {category} for {base_id}")
            elif i > 0:
                prev_category = self.PRIMARY_CHAIN[i-1]
                if prev_category not in group:
                    issues.append(f"Broken chain: {prev_category} → {category} for {base_id}")

        return issues
```

### 실시간 인덱스 관리

```python
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
import json
from pathlib import Path

class TagIndexManager(FileSystemEventHandler):
    def __init__(self, index_path: Path):
        self.index_path = index_path
        self.parser = TagParser()
        self.validator = TagChainValidator()

        # 인덱스 파일 초기화
        self.load_index()

    def on_modified(self, event):
        """파일 수정 시 TAG 재스캔"""
        if event.is_directory or not self._is_trackable_file(event.src_path):
            return

        self.update_file_tags(Path(event.src_path))

    def update_file_tags(self, file_path: Path):
        """특정 파일의 TAG 정보 업데이트"""
        try:
            tags = self.parser.parse_file(str(file_path))
            self._update_index(file_path, tags)
            self._save_index()
        except Exception as e:
            print(f"TAG 업데이트 실패: {file_path}, 오류: {e}")

    def _is_trackable_file(self, file_path: str) -> bool:
        """추적 대상 파일인지 확인"""
        trackable_extensions = {'.py', '.md', '.txt', '.yml', '.yaml', '.json'}
        return Path(file_path).suffix.lower() in trackable_extensions
```
