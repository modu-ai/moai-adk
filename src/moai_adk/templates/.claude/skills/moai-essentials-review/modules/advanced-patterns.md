# 고급 코드 리뷰 패턴

## Context7 기반 자동 리뷰

```python
class AdvancedCodeReviewer:
    """고급 자동 코드 리뷰."""

    async def comprehensive_review(self, code: str, language: str) -> ReviewReport:
        """종합 코드 리뷰."""

        # Context7 최신 리뷰 패턴
        context7_patterns = await self.context7.get_library_docs(
            context7_library_id="/code-quality/patterns",
            topic="advanced code review analysis patterns",
            tokens=4000
        )

        # 다층 분석
        reviews = await asyncio.gather(
            self.review_quality(code, language, context7_patterns),
            self.review_security(code, language),
            self.review_performance(code, language),
            self.review_architecture(code, language)
        )

        return ReviewReport(
            quality_review=reviews[0],
            security_review=reviews[1],
            performance_review=reviews[2],
            architecture_review=reviews[3],
            overall_score=self.calculate_overall_score(reviews)
        )
```

## AI 학습 기반 패턴 인식

```python
class PatternLearningReviewer:
    """학습 기반 패턴 인식."""

    async def learn_team_patterns(self, previous_reviews):
        """팀의 리뷰 패턴 학습."""

        # 승인된 PR 분석
        approved_patterns = self.extract_patterns(previous_reviews['approved'])

        # 거절된 PR 분석
        rejected_patterns = self.extract_patterns(previous_reviews['rejected'])

        # AI 모델 훈련
        self.ai_model.train(approved_patterns, rejected_patterns)

        return PatternModel(
            approved_characteristics=approved_patterns,
            rejection_reasons=rejected_patterns,
            team_preferences=self.infer_preferences(approved_patterns)
        )

    async def apply_learned_patterns(self, code):
        """학습된 패턴 적용."""

        # AI 모델로 분석
        analysis = await self.ai_model.analyze(code)

        return ReviewSuggestions(
            consistency_issues=analysis['consistency'],
            style_violations=analysis['style'],
            pattern_mismatches=analysis['patterns']
        )
```

---

**Last Updated**: 2025-11-22
