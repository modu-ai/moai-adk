# 리팩토링 최적화 도구

## 자동 리팩토링 도구

### Rope 통합

```python
class RopeRefactoringEngine:
    """Rope를 사용한 자동 리팩토링."""

    def __init__(self, project_path):
        self.project = rope.base.project.Project(project_path)

    def extract_method(self, file_path, start, end, method_name):
        """메서드 추출 자동화."""
        resource = self.project.get_resource(file_path)
        extractor = rope.refactor.extract.ExtractMethod(
            self.project, resource, start, end
        )
        changes = extractor.get_changes(method_name)
        self.project.do(changes)

    def rename_symbol(self, file_path, offset, new_name):
        """기호 이름 변경."""
        resource = self.project.get_resource(file_path)
        renamer = rope.refactor.rename.Rename(
            self.project, resource, offset
        )
        changes = renamer.get_changes(new_name)
        self.project.do(changes)
```

### Black 포매팅 통합

```python
class AutoFormatter:
    """자동 코드 포매팅."""

    def format_project(self, project_path):
        """전체 프로젝트 포매팅."""
        for python_file in glob.glob(f"{project_path}/**/*.py"):
            with open(python_file) as f:
                code = f.read()

            # Black으로 포매팅
            formatted = black.format_str(code, mode=black.FileMode())

            with open(python_file, 'w') as f:
                f.write(formatted)

        return FormattingResult(files_formatted=count)
```

## 코드 품질 분석

### Pylint 기반 품질 체크

```python
class CodeQualityChecker:
    """코드 품질 자동 체크."""

    def check_code_quality(self, file_path):
        """코드 품질 분석."""
        from pylint.lint import Run

        results = Run([file_path], exit=False)

        return QualityMetrics(
            score=results.linter.stats.global_note,
            issues=results.linter.stats.messages_count,
            refactoring_opportunities=self.identify_opportunities(results)
        )
```

## 리팩토링 검증

### 테스트 기반 검증

```python
class RefactoringValidator:
    """리팩토링 검증."""

    async def validate_refactoring(self, original_code, refactored_code):
        """리팩토링 유효성 검증."""

        # 동작 동등성 확인
        original_tests = self.run_tests(original_code)
        refactored_tests = self.run_tests(refactored_code)

        assert original_tests == refactored_tests, "Behavior changed!"

        # 성능 비교
        performance_gain = await self.compare_performance(
            original_code, refactored_code
        )

        return ValidationResult(
            behavior_preserved=True,
            performance_improvement=performance_gain,
            status='valid'
        )
```

---

**Last Updated**: 2025-11-22
