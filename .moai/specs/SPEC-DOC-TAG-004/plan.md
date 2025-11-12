# Implementation Plan: @SPEC:DOC-TAG-004

**@TAG 자동 검증 및 품질 게이트 (Phase 4)**

---

## 📋 Overview

Phase 4는 @TAG 시스템의 최종 단계로, **자동 검증 및 품질 게이트**를 구축합니다. Phase 1-3에서 구축한 TAG 생성 및 동기화 시스템을 기반으로, 커밋 및 PR 시점에 TAG 무결성을 자동으로 검증하고 문제를 조기에 발견하는 자동화 파이프라인을 완성합니다.

### 핵심 목표

1. **Pre-commit Hooks**: 로컬 커밋 시 빠른 TAG 검증 (3초 이내)
2. **CI/CD Pipeline**: GitHub Actions 통합 및 전체 검증 (5분 이내)
3. **Validation System**: 복잡한 TAG 체인 및 무결성 검증
4. **Documentation**: TAG 인벤토리 및 매트릭스 자동 생성

### 예상 작업량

- **총 작업 시간**: 40-50 시간
- **구현 기간**: 3주 (주당 15-20 시간)
- **팀 구성**: 1명 (풀타임 개발자)

---

## 🗓️ Weekly Breakdown

### Week 1: Pre-commit Hooks + CI/CD Pipeline (20 시간)

**Component 1: Pre-commit Hooks (8 시간)**

**Day 1-2 (5 시간)**:
- [ ] Pre-commit 프레임워크 설치 및 설정
- [ ] `.moai/hooks/` 디렉토리 구조 생성
- [ ] `pre-commit.sh` 스크립트 작성 (기본 골격)
- [ ] Git Hooks 연동 테스트

**Day 3 (3 시간)**:
- [ ] `tag-validator.sh` 검증 로직 구현
  - TAG 중복 감지
  - TAG 형식 검증 (regex 패턴)
  - 변경된 파일만 스캔 (Git staged files)
- [ ] `validation-rules.yml` 설정 파일 작성
- [ ] 성능 최적화 (병렬 처리)

**Component 2: CI/CD Pipeline (12 시간)**

**Day 4-5 (6 시간)**:
- [ ] `.github/workflows/tag-validation.yml` 작성
  - PR 트리거 설정
  - Python 환경 설정 (3.9+)
  - 의존성 설치 스크립트
- [ ] 전체 저장소 스캔 로직 구현
- [ ] TAG 체인 검증 통합

**Day 6 (3 시간)**:
- [ ] 검증 리포트 생성 로직
- [ ] PR 코멘트 자동 추가 스크립트
- [ ] GitHub Actions Artifact 업로드

**Day 7 (3 시간)**:
- [ ] 품질 게이트 설정 (Required Status Check)
- [ ] 워크플로우 테스트 (성공/실패 시나리오)
- [ ] 문서화: CI/CD 워크플로우 가이드 작성

---

### Week 2: Validation System (15 시간)

**Component 3: Validation System (15 시간)**

**Day 8-9 (6 시간)**:
- [ ] `src/moai_adk/core/tags/validator.py` 구현
  - `TAGValidator` 클래스 설계
  - `validate_repository()` 메서드
  - `validate_files()` 메서드
- [ ] `models.py` 작성 (ValidationReport, ValidationIssue)
- [ ] Phase 1 `TAGGenerator` 통합

**Day 10-11 (5 시간)**:
- [ ] `duplicate_detector.py` 구현
  - 중복 TAG 감지 알고리즘
  - 예외 케이스 처리 (문서 참조 허용)
  - 중복 위치 리포트 생성
- [ ] `orphan_detector.py` 구현
  - TAG 체인 그래프 구축
  - 고아 TAG 감지 알고리즘
  - 체인 복구 제안 생성

**Day 12-13 (4 시간)**:
- [ ] `chain_validator.py` 구현
  - SPEC → TEST → CODE → DOC 체인 검증
  - 체인 규칙 설정 파일 로드
  - 체인 끊김 감지 및 리포트
- [ ] 검증 로직 최적화 (캐싱, 병렬 처리)
- [ ] 단위 테스트 작성 (각 검증 모듈)

---

### Week 3: Documentation + Testing (5-10 시간)

**Component 4: Documentation & Reporting (8 시간)**

**Day 14-15 (4 시간)**:
- [ ] TAG 인벤토리 자동 생성 스크립트
  - `docs/tag-inventory.md` 템플릿
  - 파일별, 타입별, 도메인별 그룹핑
  - CI/CD 파이프라인 통합
- [ ] TAG 매트릭스 자동 생성 스크립트
  - `docs/tag-matrix.md` 템플릿
  - 체인 매핑 테이블 생성

**Day 16 (2 시간)**:
- [ ] 검증 리포트 템플릿 작성
  - `validation-reports/{date}-validation-report.md`
  - 이슈 상세 리스트 포맷
  - 수정 제안 가이드 포함
- [ ] 자동 커밋 로직 구현 (인벤토리 업데이트 시)

**Day 17 (2 시간)**:
- [ ] 문서화: TAG 검증 가이드 작성
- [ ] 문서화: 트러블슈팅 가이드 작성
- [ ] README 업데이트 (Phase 4 완료 상태 반영)

**Final Testing & Integration (5-10 시간)**

**Day 18-19 (3-5 시간)**:
- [ ] 통합 테스트: Pre-commit Hook + CI/CD 연동
- [ ] 통합 테스트: 전체 검증 플로우 (로컬 → PR → 머지)
- [ ] 성능 테스트: 대규모 저장소 검증 시나리오
- [ ] False Positive 비율 측정 및 개선

**Day 20 (2-5 시간)**:
- [ ] 버그 수정 및 코드 리뷰
- [ ] 최종 문서 정리
- [ ] Phase 4 완료 리포트 작성
- [ ] PR 생성 및 머지

---

## 🔧 Technical Approach

### 1. Pre-commit Hooks Architecture

```
Git Commit Trigger
    ↓
.git/hooks/pre-commit
    ↓
.moai/hooks/pre-commit.sh (Main Hook)
    ↓
.moai/hooks/tag-validator.sh (Validation Logic)
    ↓
Python: TAGGenerator.scan_file()
    ↓
Validation Rules (.moai/hooks/config/validation-rules.yml)
    ↓
Success → Allow Commit
Failure → Block Commit + Error Report
```

**성능 최적화 전략**:
- Git staged files만 검증 (전체 스캔 안 함)
- 병렬 처리: 여러 파일 동시 검증
- 캐싱: 이전 검증 결과 재사용 (파일 수정 시만 재검증)

**Target**: 3초 이내 (100개 파일 이하), 5초 이내 (500개 파일 이하)

---

### 2. CI/CD Pipeline Architecture

```
GitHub PR Trigger
    ↓
.github/workflows/tag-validation.yml
    ↓
Checkout Code + Setup Python
    ↓
Install Dependencies (moai-adk)
    ↓
Run TAGValidator.validate_repository()
    ↓
Generate Validation Report
    ↓
Success:
  - Update TAG Inventory
  - Auto-commit (skip ci)
  - Mark PR as passed
Failure:
  - Comment on PR (error details)
  - Block PR merge
  - Upload report artifact
```

**Workflow Configuration** (`.github/workflows/tag-validation.yml`):

```yaml
name: TAG Validation

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  validate-tags:
    runs-on: ubuntu-latest
    timeout-minutes: 10

    steps:
      - name: Checkout code
        uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for TAG chain validation

      - name: Setup Python 3.9+
        uses: actions/setup-python@v4
        with:
          python-version: '3.9'

      - name: Install dependencies
        run: |
          pip install -e .
          pip install pytest pytest-cov

      - name: Run TAG validation
        id: validation
        run: |
          python -m moai_adk.core.tags.validator \
            --mode=full \
            --output=json \
            --report-path=validation-report.json

      - name: Generate validation report
        if: always()
        run: |
          python scripts/generate_validation_report.py \
            --input=validation-report.json \
            --output=docs/validation-reports/$(date +%Y-%m-%d)-report.md

      - name: Update TAG inventory
        if: steps.validation.outcome == 'success'
        run: |
          python scripts/generate_tag_inventory.py
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add docs/tag-inventory.md docs/tag-matrix.md
          git commit -m "docs(TAG): Update TAG inventory and matrix [skip ci]" || true
          git push

      - name: Comment on PR
        if: github.event_name == 'pull_request' && steps.validation.outcome == 'failure'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const report = fs.readFileSync('validation-report.json', 'utf8');
            const issues = JSON.parse(report).issues;
            let comment = '## ❌ TAG Validation Failed\n\n';
            comment += `Found ${issues.length} issue(s):\n\n`;
            issues.forEach(issue => {
              comment += `- **${issue.severity.toUpperCase()}**: ${issue.message}\n`;
              comment += `  - File: \`${issue.file_path}\` (Line ${issue.line_number})\n`;
              comment += `  - Suggestion: ${issue.suggestion}\n\n`;
            });
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });

      - name: Upload validation report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: tag-validation-report
          path: validation-report.json
```

---

### 3. Validation System Architecture

```
TAGValidator (Main Orchestrator)
    ↓
┌─────────────────┬──────────────────┬──────────────────┐
│                 │                  │                  │
DuplicateDetector ChainValidator     OrphanDetector
│                 │                  │
│                 │                  │
└─────────────────┴──────────────────┴──────────────────┘
                  ↓
        ValidationReport (JSON/Markdown)
```

**Key Classes**:

1. **`TAGValidator`** (Main Orchestrator):
```python
class TAGValidator:
    def __init__(self, config_path: Optional[str] = None):
        self.config = self._load_config(config_path)
        self.generator = TAGGenerator()  # Reuse from Phase 1
        self.duplicate_detector = DuplicateDetector()
        self.chain_validator = ChainValidator()
        self.orphan_detector = OrphanDetector()

    def validate_repository(self, repo_path: str) -> ValidationReport:
        """전체 저장소 검증"""
        tags = self._scan_repository(repo_path)
        issues = []

        # Run all validators
        issues.extend(self.duplicate_detector.detect(tags))
        issues.extend(self.chain_validator.validate(tags))
        issues.extend(self.orphan_detector.detect(tags))

        return ValidationReport(
            total_tags=len(tags),
            valid_tags=len(tags) - len(issues),
            issues=issues,
            duplicate_count=len([i for i in issues if i.issue_type == "duplicate"]),
            orphan_count=len([i for i in issues if i.issue_type == "orphan"]),
            chain_breaks=len([i for i in issues if i.issue_type == "chain_break"]),
            execution_time=self._execution_time
        )

    def validate_files(self, file_paths: List[str]) -> ValidationReport:
        """특정 파일 목록 검증 (Pre-commit Hook용)"""
        tags = []
        for file_path in file_paths:
            tags.extend(self.generator.scan_file(file_path))
        return self._validate_tags(tags)
```

2. **`DuplicateDetector`**:
```python
class DuplicateDetector:
    def detect(self, tags: List[TAGInfo]) -> List[ValidationIssue]:
        """중복 TAG 감지"""
        tag_map = defaultdict(list)
        for tag in tags:
            # Skip documentation references (allowed duplicates)
            if self._is_reference(tag):
                continue
            tag_map[tag.id].append(tag)

        issues = []
        for tag_id, occurrences in tag_map.items():
            if len(occurrences) > 1:
                issues.append(ValidationIssue(
                    severity="error",
                    issue_type="duplicate",
                    tag_id=tag_id,
                    file_path=occurrences[0].file_path,
                    line_number=occurrences[0].line_number,
                    message=f"Duplicate TAG '{tag_id}' found in {len(occurrences)} files",
                    suggestion=f"Remove duplicates from: {[t.file_path for t in occurrences[1:]]}"
                ))
        return issues
```

3. **`ChainValidator`**:
```python
class ChainValidator:
    def validate(self, tags: List[TAGInfo]) -> List[ValidationIssue]:
        """TAG 체인 완전성 검증"""
        chain_rules = self._load_chain_rules()  # From config
        chain_graph = self._build_chain_graph(tags)

        issues = []
        for rule in chain_rules:
            # Example: "@SPEC:{ID} -> @TEST:{ID}" (required)
            for tag in tags:
                if tag.type == rule.source_type:
                    expected_target = rule.target_type + ":" + tag.domain + "-" + tag.id
                    if expected_target not in chain_graph:
                        issues.append(ValidationIssue(
                            severity=rule.severity,
                            issue_type="chain_break",
                            tag_id=tag.id,
                            file_path=tag.file_path,
                            line_number=tag.line_number,
                            message=f"Missing chain link: {tag.full_id} -> {expected_target}",
                            suggestion=f"Create {expected_target} or update TAG chain"
                        ))
        return issues
```

4. **`OrphanDetector`**:
```python
class OrphanDetector:
    def detect(self, tags: List[TAGInfo]) -> List[ValidationIssue]:
        """고아 TAG 감지 (체인이 끊긴 TAG)"""
        chain_graph = self._build_chain_graph(tags)
        root_tags = self._find_root_tags(tags)  # SPEC tags

        orphans = []
        for tag in tags:
            if not self._has_path_to_root(tag, chain_graph, root_tags):
                orphans.append(ValidationIssue(
                    severity="warning",
                    issue_type="orphan",
                    tag_id=tag.id,
                    file_path=tag.file_path,
                    line_number=tag.line_number,
                    message=f"Orphan TAG detected: {tag.full_id} (no connection to SPEC)",
                    suggestion=self._suggest_reconnection(tag, chain_graph)
                ))
        return orphans
```

---

### 4. Documentation & Reporting

**TAG Inventory Generation** (`scripts/generate_tag_inventory.py`):

```python
def generate_tag_inventory(tags: List[TAGInfo]) -> str:
    """TAG 인벤토리 Markdown 생성"""
    output = "# TAG Inventory\n\n"

    # Summary
    output += "## Summary\n"
    output += f"- Total TAGs: {len(tags)}\n"
    output += f"- By Type: {_group_by_type(tags)}\n"
    output += f"- By Domain: {_group_by_domain(tags)}\n\n"

    # TAG List by File
    output += "## TAG List by File\n\n"
    files = _group_by_file(tags)
    for file_path, file_tags in sorted(files.items()):
        output += f"### {file_path}\n"
        for tag in file_tags:
            output += f"- {tag.full_id} (Line {tag.line_number})\n"
        output += "\n"

    return output
```

**TAG Matrix Generation** (`scripts/generate_tag_matrix.py`):

```python
def generate_tag_matrix(tags: List[TAGInfo]) -> str:
    """TAG 매트릭스 Markdown 생성"""
    output = "# TAG Matrix\n\n"
    output += "| SPEC | TEST | CODE | DOC | Status |\n"
    output += "|------|------|------|-----|--------|\n"

    domains = _group_by_domain(tags)
    for domain, domain_tags in sorted(domains.items()):
        spec_tag = _find_tag(domain_tags, "SPEC")
        test_tag = _find_tag(domain_tags, "TEST")
        code_tag = _find_tag(domain_tags, "CODE")
        doc_tag = _find_tag(domain_tags, "DOC")

        status = _compute_status(spec_tag, test_tag, code_tag, doc_tag)
        output += f"| {spec_tag or '-'} | {test_tag or '-'} | {code_tag or '-'} | {doc_tag or '-'} | {status} |\n"

    return output
```

---

## 🔄 Reuse Strategy (Phase 1-3 Leverage)

Phase 4는 이전 Phase의 성과물을 최대한 재사용하여 개발 시간을 단축합니다:

### Phase 1 재사용 (`@SPEC:DOC-TAG-001`)

**`TAGGenerator` 클래스**:
- `scan_file()`: Pre-commit Hook에서 파일 스캔
- `scan_directory()`: CI/CD에서 전체 저장소 스캔
- `parse_tag()`: TAG 파싱 로직 재사용

**예상 재사용 비율**: 60% (검증 로직만 신규 작성)

### Phase 2 재사용 (`@SPEC:DOC-TAG-002`)

**`TAG-CHAIN-MAP.md`**:
- 체인 검증의 기준 데이터로 사용
- 매트릭스 생성 시 참조

**SPEC 파일 TAG 패턴**:
- SPEC 파일의 TAG 필수 여부 검증에 활용

**예상 재사용 비율**: 30% (체인 규칙 재사용)

### Phase 3 재사용 (`@SPEC:DOC-TAG-003`)

**`doc-syncer` 에이전트**:
- TAG 동기화 로직을 인벤토리 업데이트에 활용
- Markdown 파싱 로직 재사용

**`tag-agent` 기반**:
- TAG 스캔 로직 재사용

**예상 재사용 비율**: 40% (문서 처리 로직 재사용)

### 신규 개발 범위

- **Pre-commit Hook 스크립트**: 100% 신규
- **CI/CD 워크플로우**: 100% 신규
- **검증 로직** (Duplicate/Orphan/Chain): 70% 신규 (기본 스캔 로직은 재사용)
- **리포트 생성**: 50% 신규 (템플릿은 신규, 데이터 수집은 재사용)

---

## 🚨 Risk Assessment & Mitigation

### Risk 1: Pre-commit Hook 실행 시간 초과 (>5초)

**영향**: 개발자 경험 저하, Hook 비활성화 유도
**확률**: 중 (30%)
**대응 전략**:
- 병렬 처리로 성능 최적화
- 경량 검증만 수행 (중복/형식만, 체인 검증 제외)
- 캐싱으로 재검증 최소화
- 성능 테스트로 임계값 사전 확인

### Risk 2: CI/CD 파이프라인 False Positive

**영향**: 정상 PR이 차단되어 개발 속도 저하
**확률**: 중 (40%)
**대응 전략**:
- 철저한 단위 테스트로 검증 로직 검증
- 예외 케이스 명확히 정의 (문서 참조 중복 허용 등)
- 검증 규칙을 설정 파일로 관리하여 유연하게 조정
- Beta 테스트 기간 운영 (2주간 Warning only 모드)

### Risk 3: TAG 검증 규칙 복잡도 증가

**영향**: 유지보수 어려움, 새로운 규칙 추가 시 오류 발생
**확률**: 저 (20%)
**대응 전략**:
- 검증 규칙을 YAML 설정 파일로 분리
- 규칙별 단위 테스트 작성
- 규칙 추가 가이드 문서화
- 규칙 변경 시 회귀 테스트 자동화

### Risk 4: 고아 TAG 자동 복구 실패

**영향**: 수동 개입 필요, 자동화 목표 미달성
**확률**: 중 (50%)
**대응 전략**:
- 자동 복구 대신 복구 제안만 제공
- 복구 제안 정확도 95% 이상 목표
- 수동 복구 가이드 문서 제공
- CLI 유틸리티로 복구 작업 지원

### Risk 5: 대규모 저장소 성능 저하

**영향**: CI/CD 타임아웃 (10분 초과), 검증 불가
**확률**: 저 (15%)
**대응 전략**:
- 증분 검증 옵션 제공 (변경된 파일만)
- 병렬 처리 및 캐싱 최대 활용
- 저장소 크기별 타임아웃 설정 조정
- 성능 벤치마크 테스트로 사전 확인

---

## ✅ Success Criteria

Phase 4 구현 완료 시 다음 조건을 **모두** 충족해야 합니다:

### Performance Metrics

- ✅ **Pre-commit Hook 실행 시간**: 3초 이내 (100개 파일 기준)
- ✅ **CI/CD 파이프라인 실행 시간**: 5분 이내 (전체 저장소)
- ✅ **TAG 중복 감지율**: 100% (모든 중복 TAG 감지)
- ✅ **고아 TAG 감지율**: 95% 이상
- ✅ **False Positive 비율**: 5% 이하
- ✅ **TAG 인벤토리 자동 업데이트 성공률**: 100%

### Functional Requirements

- ✅ **Pre-commit Hook**: 로컬 커밋 시 TAG 검증 자동 실행
- ✅ **CI/CD Pipeline**: PR 생성/업데이트 시 전체 검증 자동 실행
- ✅ **품질 게이트**: 검증 실패 시 PR 머지 차단
- ✅ **검증 리포트**: PR 코멘트 및 Artifact 생성
- ✅ **TAG 인벤토리**: 검증 성공 시 자동 업데이트 및 커밋

### Test Coverage

- ✅ **단위 테스트**: 모든 검증 모듈 95% 이상 커버리지
- ✅ **통합 테스트**: Pre-commit + CI/CD 통합 시나리오 10개 이상
- ✅ **성능 테스트**: 대규모 저장소 시나리오 (1000개 파일)
- ✅ **회귀 테스트**: 기존 Phase 1-3 기능 영향 없음

### Documentation

- ✅ **TAG 검증 가이드**: Pre-commit Hook 설치 및 사용법
- ✅ **CI/CD 워크플로우 가이드**: GitHub Actions 설정 방법
- ✅ **트러블슈팅 가이드**: 검증 실패 시 해결 방법
- ✅ **TAG 인벤토리**: 자동 생성된 최신 TAG 목록
- ✅ **TAG 매트릭스**: 자동 생성된 체인 매핑 테이블

---

## 📝 Notes

### 개발 환경 준비

Phase 4 구현 시작 전에 다음 환경을 준비합니다:

1. **Pre-commit 프레임워크 설치**:
   ```bash
   pip install pre-commit
   ```

2. **GitHub CLI 설치** (PR 코멘트 테스트용):
   ```bash
   brew install gh  # macOS
   # or
   sudo apt install gh  # Ubuntu
   ```

3. **Python 의존성 업데이트**:
   ```bash
   pip install -e .[dev]
   pip install pytest pytest-cov pytest-mock
   ```

4. **Git Hooks 테스트 환경**:
   ```bash
   git config core.hooksPath .moai/hooks
   ```

### 팀 협업 고려사항

- **Pre-commit Hook 설치 강제화**: 팀원 모두가 로컬에서 `pre-commit install` 실행 필요
- **CI/CD 실패 시 대응**: 검증 실패 시 즉시 수정하도록 정책 수립
- **품질 게이트 예외 처리**: 긴급 상황 시 관리자 승인으로 머지 가능하도록 설정

### 향후 확장 가능성

Phase 4 완료 후 추가 가능한 기능:

- **IDE 플러그인**: VSCode/PyCharm에서 실시간 TAG 검증
- **웹 대시보드**: TAG 인벤토리 및 매트릭스 시각화
- **TAG 자동 수정**: 일부 오류 자동 수정 (중복 TAG 제거 등)
- **성능 모니터링**: TAG 검증 성능 추이 추적

---

**문서 종료**
