# 코드 리뷰 최적화

## 병렬 리뷰 처리

```python
class ParallelReviewProcessor:
    """병렬 리뷰 처리."""

    async def review_multiple_files(self, files: List[FileChange]) -> List[FileReview]:
        """여러 파일 병렬 리뷰."""

        # 파일을 배치로 분리
        batches = self.create_batches(files, batch_size=10)

        results = []
        for batch in batches:
            # 배치 병렬 처리
            batch_results = await asyncio.gather(*[
                self.review_file(file) for file in batch
            ])
            results.extend(batch_results)

        return results
```

## 캐싱된 리뷰 패턴

```python
class CachedReviewPatterns:
    """캐싱된 리뷰 패턴."""

    def __init__(self):
        self.pattern_cache = {}

    async def get_cached_review(self, code_signature: str):
        """캐시된 리뷰 조회."""

        if code_signature in self.pattern_cache:
            return self.pattern_cache[code_signature]

        # 리뷰 실행 및 캐시
        review = await self.perform_review(code_signature)
        self.pattern_cache[code_signature] = review

        return review
```

## 리뷰 보고서 최적화

```python
class OptimizedReportGenerator:
    """최적화된 리뷰 보고서 생성."""

    def generate_summary_report(self, reviews: List[Review]) -> SummaryReport:
        """요약 리포트 생성."""

        return SummaryReport(
            critical_issues=self.extract_critical(reviews),
            warnings=self.extract_warnings(reviews),
            suggestions=self.top_k_suggestions(reviews, k=5),
            metrics_summary=self.calculate_metrics(reviews)
        )
```

---

**Last Updated**: 2025-11-22
