# Skill 검증 시스템 완전 가이드

## 개요

MoAI-ADK의 검증 시스템은 55개의 Skills가 항상 최고 품질을 유지하도록 보장합니다. 자동화된 검증, 자동 수정 메커니즘, 그리고 TRUST 5 원칙 준수를 통해 프로덕션 환경에서 신뢰할 수 있는 Skills를 제공합니다.

**핵심 기능**:
- **3단계 위험 등급**: SAFE, MEDIUM, HIGH 수정 전략
- **자동 수정**: Duplicate TAG, Format 에러, Whitespace 자동 정규화
- **품질 지표**: S, A, B, C 등급 자동 계산
- **TRUST 5 준수**: 테스트, 가독성, 통합성, 보안, 추적성 검증
- **실시간 모니터링**: CI/CD 파이프라인 통합

## 검증 프레임워크

### 1. 자동 검증 파이프라인

```python
# moai_adk/validation/validator.py
from typing import List, Dict
from dataclasses import dataclass

@dataclass
class ValidationResult:
    """검증 결과"""
    skill_name: str
    status: str  # 'pass', 'warning', 'error'
    score: float  # 0-100
    grade: str  # 'S', 'A', 'B', 'C'
    issues: List[Dict]
    fixed_issues: List[Dict]
    metrics: Dict

class SkillValidator:
    """Skill 검증 엔진"""

    def __init__(self):
        self.checks = [
            self.check_structure,
            self.check_tags,
            self.check_formatting,
            self.check_trust5,
            self.check_content_quality,
        ]

    def validate(self, skill_path: str) -> ValidationResult:
        """Skill 검증 실행"""
        skill = self.load_skill(skill_path)
        issues = []
        fixed_issues = []

        # 모든 검증 실행
        for check in self.checks:
            check_result = check(skill)
            issues.extend(check_result['issues'])

            # 자동 수정 시도
            if check_result['auto_fixable']:
                fixed = self.auto_fix(skill, check_result)
                fixed_issues.extend(fixed)

        # 점수 계산
        score = self.calculate_score(skill, issues, fixed_issues)
        grade = self.assign_grade(score)

        return ValidationResult(
            skill_name=skill.name,
            status=self.determine_status(issues),
            score=score,
            grade=grade,
            issues=[i for i in issues if not i['fixed']],
            fixed_issues=fixed_issues,
            metrics=self.collect_metrics(skill),
        )
```

### 2. 3단계 위험 등급 시스템

```python
# moai_adk/validation/risk_levels.py
from enum import Enum

class RiskLevel(Enum):
    """수정 위험 등급"""
    SAFE = "safe"        # 자동 수정 안전
    MEDIUM = "medium"    # 검토 후 수정
    HIGH = "high"        # 수동 수정 필요

class AutoFixStrategy:
    """자동 수정 전략"""

    @staticmethod
    def classify_issue(issue: Dict) -> RiskLevel:
        """이슈 위험도 분류"""
        if issue['type'] in ['whitespace', 'formatting', 'duplicate_tag']:
            return RiskLevel.SAFE

        if issue['type'] in ['missing_tag', 'incomplete_section']:
            return RiskLevel.MEDIUM

        if issue['type'] in ['logic_error', 'breaking_change']:
            return RiskLevel.HIGH

        return RiskLevel.MEDIUM

    @staticmethod
    def can_auto_fix(issue: Dict) -> bool:
        """자동 수정 가능 여부"""
        risk = AutoFixStrategy.classify_issue(issue)
        return risk == RiskLevel.SAFE

# 실제 예제: v0.23.1 업그레이드
"""
SAFE 수정 (자동):
- 중복된 @TAG 제거
- Trailing whitespace 정규화
- 일관되지 않은 헤딩 레벨 수정

MEDIUM 수정 (검토 후):
- 누락된 @TAG 추가
- 불완전한 섹션 완성

HIGH 수정 (수동):
- 논리 오류 수정
- Breaking change 처리
"""
```

### 3. 자동 수정 메커니즘

#### Duplicate TAG 제거

```python
# moai_adk/validation/fixes/duplicate_tags.py
import re
from collections import Counter

class DuplicateTagFixer:
    """중복 TAG 자동 제거"""

    @staticmethod
    def detect_duplicates(content: str) -> List[str]:
        """중복 TAG 탐지"""
        tag_pattern = r'@TAG-[A-Z]+-\d{3}'
        tags = re.findall(tag_pattern, content)

        # 중복 태그 찾기
        tag_counts = Counter(tags)
        duplicates = [tag for tag, count in tag_counts.items() if count > 1]

        return duplicates

    @staticmethod
    def fix(content: str) -> tuple[str, List[str]]:
        """중복 TAG 제거"""
        duplicates = DuplicateTagFixer.detect_duplicates(content)
        fixed_tags = []

        for tag in duplicates:
            # 첫 번째 출현은 유지, 나머지 제거
            pattern = re.escape(tag)
            matches = list(re.finditer(pattern, content))

            # 뒤에서부터 제거 (인덱스 유지)
            for match in reversed(matches[1:]):
                start, end = match.span()
                content = content[:start] + content[end:]
                fixed_tags.append({
                    'tag': tag,
                    'position': start,
                    'action': 'removed_duplicate',
                })

        return content, fixed_tags

# 사용 예제
before = """
## Section 1
Description with @TAG-SKILL-001

## Section 2
Another reference to @TAG-SKILL-001
"""

after, fixed = DuplicateTagFixer.fix(before)
# 두 번째 @TAG-SKILL-001 제거됨
```

#### Format 에러 수정

```python
# moai_adk/validation/fixes/format_errors.py
class FormatErrorFixer:
    """형식 에러 자동 수정"""

    @staticmethod
    def fix_heading_levels(content: str) -> tuple[str, List[Dict]]:
        """헤딩 레벨 일관성 수정"""
        lines = content.split('\n')
        fixed_lines = []
        fixes = []

        expected_level = 1
        for i, line in enumerate(lines):
            if line.startswith('#'):
                # 현재 레벨 계산
                current_level = len(line) - len(line.lstrip('#'))

                # 레벨 차이가 2 이상이면 수정
                if current_level - expected_level > 1:
                    new_level = expected_level + 1
                    new_line = '#' * new_level + line.lstrip('#')
                    fixed_lines.append(new_line)

                    fixes.append({
                        'line': i + 1,
                        'old_level': current_level,
                        'new_level': new_level,
                        'action': 'normalized_heading_level',
                    })
                else:
                    fixed_lines.append(line)
                    expected_level = current_level
            else:
                fixed_lines.append(line)

        return '\n'.join(fixed_lines), fixes

    @staticmethod
    def fix_code_blocks(content: str) -> tuple[str, List[Dict]]:
        """코드 블록 형식 수정"""
        # 언어 지정이 없는 코드 블록 수정
        pattern = r'```\n((?:(?!```)[\s\S])+)```'

        def add_language(match):
            code = match.group(1)

            # 코드 내용으로 언어 추측
            if 'import' in code or 'from' in code:
                return f'```python\n{code}```'
            elif 'function' in code or 'const' in code:
                return f'```typescript\n{code}```'
            else:
                return f'```\n{code}```'

        fixed_content = re.sub(pattern, add_language, content)
        return fixed_content, []
```

#### Whitespace 정규화

```python
# moai_adk/validation/fixes/whitespace.py
class WhitespaceFixer:
    """공백 정규화"""

    @staticmethod
    def fix(content: str) -> tuple[str, List[Dict]]:
        """공백 정규화 실행"""
        fixes = []

        # 1. Trailing whitespace 제거
        lines = content.split('\n')
        cleaned_lines = []
        for i, line in enumerate(lines):
            if line.rstrip() != line:
                fixes.append({
                    'line': i + 1,
                    'action': 'removed_trailing_whitespace',
                })
            cleaned_lines.append(line.rstrip())

        # 2. 연속된 빈 줄 정규화 (최대 2줄)
        normalized_lines = []
        blank_count = 0

        for line in cleaned_lines:
            if line == '':
                blank_count += 1
                if blank_count <= 2:
                    normalized_lines.append(line)
            else:
                blank_count = 0
                normalized_lines.append(line)

        # 3. 파일 끝 줄바꿈 확인
        content = '\n'.join(normalized_lines)
        if not content.endswith('\n'):
            content += '\n'
            fixes.append({
                'action': 'added_final_newline',
            })

        return content, fixes
```

## 품질 지표

### 1. 등급 시스템

```python
# moai_adk/validation/grading.py
class SkillGrading:
    """Skill 등급 산정"""

    THRESHOLDS = {
        'S': 95.0,  # 우수
        'A': 85.0,  # 양호
        'B': 70.0,  # 보통
        'C': 0.0,   # 개선 필요
    }

    @staticmethod
    def calculate_score(skill: Skill, metrics: Dict) -> float:
        """종합 점수 계산"""
        weights = {
            'structure': 0.20,      # 구조 완성도
            'tags': 0.25,           # TAG 체계
            'trust5': 0.30,         # TRUST 5 준수
            'content_quality': 0.15, # 내용 품질
            'formatting': 0.10,     # 형식 준수
        }

        score = 0
        for category, weight in weights.items():
            category_score = metrics.get(f'{category}_score', 0)
            score += category_score * weight

        return score

    @staticmethod
    def assign_grade(score: float) -> str:
        """등급 부여"""
        for grade, threshold in SkillGrading.THRESHOLDS.items():
            if score >= threshold:
                return grade
        return 'C'

# 점수 산정 예시
metrics = {
    'structure_score': 95,      # 모든 필수 섹션 존재
    'tags_score': 90,           # TAG 체계 완벽
    'trust5_score': 85,         # TRUST 5 대부분 준수
    'content_quality_score': 92, # 내용 우수
    'formatting_score': 88,     # 형식 양호
}

score = SkillGrading.calculate_score(skill, metrics)
# score = 89.5

grade = SkillGrading.assign_grade(score)
# grade = 'A'
```

### 2. 세부 메트릭

```python
# moai_adk/validation/metrics.py
class SkillMetrics:
    """Skill 메트릭 수집"""

    @staticmethod
    def collect(skill: Skill) -> Dict:
        """메트릭 수집"""
        return {
            # 구조 메트릭
            'has_all_sections': skill.has_required_sections(),
            'section_count': len(skill.sections),
            'completeness': skill.calculate_completeness(),

            # TAG 메트릭
            'tag_count': len(skill.tags),
            'duplicate_tags': skill.find_duplicate_tags(),
            'orphan_tags': skill.find_orphan_tags(),

            # 내용 메트릭
            'word_count': skill.count_words(),
            'code_example_count': skill.count_code_examples(),
            'diagram_count': skill.count_diagrams(),

            # 품질 메트릭
            'readability_score': skill.calculate_readability(),
            'complexity_score': skill.calculate_complexity(),

            # TRUST 5 메트릭
            'has_tests': skill.has_test_references(),
            'is_readable': skill.check_readability(),
            'is_unified': skill.check_unified_style(),
            'is_secure': skill.check_security_practices(),
            'is_trackable': skill.has_tag_chain(),
        }
```

## TRUST 5 준수 검증

### 1. Test First (테스트 우선)

```python
# moai_adk/validation/trust5/test_first.py
class TestFirstValidator:
    """Test First 원칙 검증"""

    @staticmethod
    def validate(skill: Skill) -> Dict:
        """검증 실행"""
        issues = []

        # 1. 테스트 예제 존재 확인
        if not skill.has_test_examples():
            issues.append({
                'type': 'missing_test_examples',
                'severity': 'high',
                'message': 'Skill should include test examples',
            })

        # 2. TDD 워크플로우 언급 확인
        if not skill.mentions_tdd_workflow():
            issues.append({
                'type': 'missing_tdd_reference',
                'severity': 'medium',
                'message': 'Should reference TDD workflow',
            })

        # 3. 테스트 가능한 코드 예제
        code_examples = skill.extract_code_examples()
        testable_count = sum(1 for ex in code_examples if ex.is_testable())

        if testable_count < len(code_examples) * 0.7:
            issues.append({
                'type': 'low_testability',
                'severity': 'medium',
                'message': f'Only {testable_count}/{len(code_examples)} examples are testable',
            })

        return {
            'passed': len(issues) == 0,
            'issues': issues,
            'score': 100 - (len(issues) * 10),
        }
```

### 2. Readable (가독성)

```python
# moai_adk/validation/trust5/readable.py
class ReadableValidator:
    """Readable 원칙 검증"""

    @staticmethod
    def validate(skill: Skill) -> Dict:
        """검증 실행"""
        issues = []

        # 1. 헤딩 구조 확인
        if not skill.has_clear_heading_structure():
            issues.append({
                'type': 'unclear_structure',
                'severity': 'high',
            })

        # 2. 코드 주석 확인
        code_blocks = skill.extract_code_blocks()
        commented = sum(1 for cb in code_blocks if cb.has_comments())

        if commented < len(code_blocks) * 0.6:
            issues.append({
                'type': 'insufficient_comments',
                'severity': 'medium',
            })

        # 3. 가독성 점수 (Flesch Reading Ease)
        readability = skill.calculate_readability_score()
        if readability < 60:  # 표준 이상 난이도
            issues.append({
                'type': 'low_readability',
                'severity': 'low',
                'score': readability,
            })

        return {
            'passed': len(issues) == 0,
            'issues': issues,
            'score': 100 - (len(issues) * 10),
        }
```

### 3. Unified (통합성)

```python
# moai_adk/validation/trust5/unified.py
class UnifiedValidator:
    """Unified 원칙 검증"""

    @staticmethod
    def validate(skill: Skill) -> Dict:
        """검증 실행"""
        issues = []

        # 1. 일관된 명명 규칙
        naming_style = skill.detect_naming_style()
        if not naming_style.is_consistent():
            issues.append({
                'type': 'inconsistent_naming',
                'severity': 'medium',
            })

        # 2. TAG 체인 연결성
        tag_chain = skill.build_tag_chain()
        if tag_chain.has_breaks():
            issues.append({
                'type': 'broken_tag_chain',
                'severity': 'high',
            })

        # 3. 크로스 레퍼런스
        references = skill.extract_cross_references()
        if len(references) < 3:
            issues.append({
                'type': 'insufficient_references',
                'severity': 'low',
            })

        return {
            'passed': len(issues) == 0,
            'issues': issues,
            'score': 100 - (len(issues) * 10),
        }
```

### 4. Secured (보안)

```python
# moai_adk/validation/trust5/secured.py
class SecuredValidator:
    """Secured 원칙 검증"""

    @staticmethod
    def validate(skill: Skill) -> Dict:
        """검증 실행"""
        issues = []

        # 1. 하드코딩된 시크릿 검사
        secrets = skill.find_hardcoded_secrets()
        if secrets:
            issues.append({
                'type': 'hardcoded_secrets',
                'severity': 'critical',
                'findings': secrets,
            })

        # 2. 보안 베스트 프랙티스 언급
        if not skill.mentions_security_practices():
            issues.append({
                'type': 'missing_security_section',
                'severity': 'medium',
            })

        # 3. 입력 검증 예제
        validation_examples = skill.find_validation_examples()
        if not validation_examples:
            issues.append({
                'type': 'no_validation_examples',
                'severity': 'low',
            })

        return {
            'passed': len(issues) == 0,
            'issues': issues,
            'score': 100 - (len(issues) * 15),  # 보안은 가중치 높음
        }
```

### 5. Trackable (추적성)

```python
# moai_adk/validation/trust5/trackable.py
class TrackableValidator:
    """Trackable 원칙 검증"""

    @staticmethod
    def validate(skill: Skill) -> Dict:
        """검증 실행"""
        issues = []

        # 1. TAG 체계 완성도
        required_tags = skill.get_required_tags()
        actual_tags = skill.extract_tags()

        missing_tags = set(required_tags) - set(actual_tags)
        if missing_tags:
            issues.append({
                'type': 'missing_required_tags',
                'severity': 'high',
                'tags': list(missing_tags),
            })

        # 2. TAG 연결 검증
        for tag in actual_tags:
            if not skill.tag_has_context(tag):
                issues.append({
                    'type': 'tag_without_context',
                    'severity': 'medium',
                    'tag': tag,
                })

        # 3. 변경 이력 추적
        if not skill.has_version_history():
            issues.append({
                'type': 'no_version_history',
                'severity': 'low',
            })

        return {
            'passed': len(issues) == 0,
            'issues': issues,
            'score': 100 - (len(issues) * 10),
        }
```

## 검증 워크플로우

### 1. CI/CD 통합

```yaml
# .github/workflows/skill-validation.yml
name: Skill Validation

on:
  pull_request:
    paths:
      - '.claude/skills/**'
  push:
    branches: [main]
    paths:
      - '.claude/skills/**'

jobs:
  validate-skills:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install MoAI-ADK
        run: pip install moai-adk

      - name: Run Skill Validation
        run: |
          moai-adk validate skills \
            --path .claude/skills \
            --strict \
            --auto-fix safe \
            --report validation-report.json

      - name: Check Quality Gates
        run: |
          moai-adk check quality-gates \
            --min-score 85 \
            --min-grade A \
            --report validation-report.json

      - name: Upload Report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: validation-report
          path: validation-report.json
```

### 2. 로컬 검증

```bash
# 단일 Skill 검증
moai-adk validate skill \
  --path .claude/skills/moai-foundation-tags.md \
  --auto-fix safe \
  --verbose

# 전체 Skills 검증
moai-adk validate skills \
  --path .claude/skills \
  --auto-fix safe \
  --report report.json

# 특정 카테고리만 검증
moai-adk validate skills \
  --path .claude/skills \
  --category foundation \
  --strict

# TRUST 5 검증만 실행
moai-adk validate trust5 \
  --path .claude/skills \
  --detailed
```

## 실제 사례: v0.23.1 업그레이드

### 문제점 발견

```
Validation Results (v0.23.0):
✗ 12 skills with duplicate TAGs
✗ 8 skills with formatting inconsistencies
✗ 5 skills with incomplete TRUST 5 compliance
✗ 3 skills with broken cross-references

Overall Score: 78.5 (Grade B)
```

### 자동 수정 적용

```python
# 자동 실행된 수정
fixes = {
    'safe': [
        'Removed 47 duplicate TAGs',
        'Normalized 156 trailing whitespaces',
        'Fixed 23 heading level inconsistencies',
        'Added 12 missing final newlines',
    ],
    'medium': [
        'Added 8 missing TAG definitions',
        'Completed 5 incomplete sections',
    ],
    'high': [
        'Manually fixed 3 broken cross-references',
        'Updated 2 outdated code examples',
    ],
}
```

### 결과

```
Validation Results (v0.23.1):
✓ All skills passed structure validation
✓ Zero duplicate TAGs
✓ 100% formatting compliance
✓ Full TRUST 5 compliance

Overall Score: 94.2 (Grade A)
Quality Improvement: +15.7 points
```

## Best Practices

### 1. 정기 검증

```bash
# 주간 검증 스케줄
0 0 * * 0 moai-adk validate skills --full --auto-fix safe
```

### 2. Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
moai-adk validate skills \
  --changed-only \
  --auto-fix safe \
  --fail-on-error
```

### 3. 품질 게이트

```python
# CI/CD에서 품질 기준 강제
quality_gates = {
    'min_score': 85,
    'min_grade': 'A',
    'max_critical_issues': 0,
    'max_high_issues': 2,
    'trust5_compliance': True,
}
```

## 다음 단계

- [Skill 개발 가이드](/ko/skills/skill-development) - 새 Skill 만들기
- [TRUST 5 원칙](/ko/skills/trust5) - 품질 원칙 상세
- [Advanced Skills](/ko/skills/advanced-skills) - 고급 기능
- [Foundation Skills](/ko/skills/foundation) - 기본 Skills

## 참고 자료

- [MoAI-ADK Validation API](https://moai-adk.dev/api/validation)
- [TRUST 5 표준](https://moai-adk.dev/standards/trust5)
- [품질 메트릭 정의](https://moai-adk.dev/metrics)
