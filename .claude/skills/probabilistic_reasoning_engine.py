#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-004 | @SPEC:SKILL-PROBABILISTIC-REASONING-ENGINE-001 | @TEST: tests/skills/test_probabilistic_reasoning_engine.py
"""Probabilistic Reasoning Engine Skill

확률적 추론 엔진. 불확실한 상황에서 합리적인 결정을 내리기 위한 기능:
1. 확률 계산
2. 리스크 평가
3. 기대값 분석
4. 통계적 추론
5. 의사결정 트리

사용법:
    Skill("probabilistic_reasoning_engine")
"""

import json
import math
import random
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Union
from collections import defaultdict, Counter
from dataclasses import dataclass


@dataclass
class ProbabilityEvent:
    """확률 이벤트 데이터 클래스"""
    name: str
    probability: float
    impact: float
    category: str


@dataclass
class DecisionOption:
    """의사결정 옵션 데이터 클래스"""
    name: str
    expected_value: float
    confidence: float
    risks: List[ProbabilityEvent]
    benefits: List[ProbabilityEvent]


class ProbabilisticReasoningEngine:
    """확률적 추론 엔진 클래스"""

    def __init__(self):
        self.historical_data = self.load_historical_data()
        self.bayesian_network = self.load_bayesian_network()
        self.decision_models = self.load_decision_models()

    def load_historical_data(self) -> Dict[str, Any]:
        """이전 데이터 로드"""
        try:
            data_file = Path(".moai/research/probabilistic/historical_data.json")
            if data_file.exists():
                with open(data_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return {
            "past_decisions": [],
            "success_rates": {},
            "risk_patterns": {},
            "confidence_factors": {}
        }

    def load_bayesian_network(self) -> Dict[str, Any]:
        """베이지안 네트워크 로드"""
        try:
            network_file = Path(".moai/research/probabilistic/bayesian_network.json")
            if network_file.exists():
                with open(network_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return {
            "nodes": {},
            "edges": {},
            "conditional_probabilities": {}
        }

    def load_decision_models(self) -> Dict[str, Any]:
        """의사결정 모델 로드"""
        try:
            models_file = Path(".moai/research/probabilistic/decision_models.json")
            if models_file.exists():
                with open(models_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return {
            "risk_assessment": self.get_default_risk_model(),
            "expected_value": self.get_default_expected_value_model(),
            "confidence_calculation": self.get_default_confidence_model()
        }

    def get_default_risk_model(self) -> Dict[str, Any]:
        """기본 리스크 평가 모델"""
        return {
            "risk_categories": {
                "technical": {"weight": 0.3, "factors": ["complexity", "uncertainty", "dependency"]},
                "business": {"weight": 0.4, "factors": ["market", "competition", "value"]},
                "operational": {"weight": 0.3, "factors": ["process", "resource", "timeline"]}
            },
            "risk_matrix": {
                "low_probability_low_impact": "accept",
                "low_probability_high_impact": "mitigate",
                "high_probability_low_impact": "monitor",
                "high_probability_high_impact": "avoid"
            }
        }

    def get_default_expected_value_model(self) -> Dict[str, Any]:
        """기본 기대값 모델"""
        return {
            "calculation_method": "weighted_average",
            "weights": {
                "positive_outcomes": 0.6,
                "negative_outcomes": 0.4,
                "time_value": 0.8,
                "confidence_factor": 0.7
            }
        }

    def get_default_confidence_model(self) -> Dict[str, Any]:
        """기본 신뢰도 모델"""
        return {
            "confidence_factors": {
                "data_quality": 0.3,
                "historical_accuracy": 0.2,
                "model_suitability": 0.2,
                "expert_consensus": 0.3
            },
            "confidence_thresholds": {
                "high": 0.8,
                "medium": 0.6,
                "low": 0.4
            }
        }

    def analyze_probability(self, events: List[Dict[str, Any]]) -> Dict[str, Any]:
        """확률 분석 수행"""
        analysis_result = {
            "analysis_type": "probability_analysis",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "events_analyzed": len(events),
            "probability_calculations": {},
            "risk_assessment": {},
            "recommendations": [],
            "confidence_level": 0.0,
            "historical_comparison": {}
        }

        # 이벤트 변환
        probability_events = []
        for event_data in events:
            event = ProbabilityEvent(
                name=event_data.get("name", "unknown"),
                probability=float(event_data.get("probability", 0.5)),
                impact=float(event_data.get("impact", 1.0)),
                category=event_data.get("category", "general")
            )
            probability_events.append(event)

        # 개별 확률 계산
        for event in probability_events:
            analysis_result["probability_calculations"][event.name] = self.calculate_event_probability(event)

        # 통계적 분석
        analysis_result["statistical_analysis"] = self.perform_statistical_analysis(probability_events)

        # 리스크 평가
        analysis_result["risk_assessment"] = self.assess_risks(probability_events)

        # 기대값 분석
        analysis_result["expected_value_analysis"] = self.calculate_expected_values(probability_events)

        # 신뢰도 계산
        analysis_result["confidence_level"] = self.calculate_confidence_level(probability_events)

        # 역사적 비교
        analysis_result["historical_comparison"] = self.compare_with_historical_data(probability_events)

        # 추천 생성
        analysis_result["recommendations"] = self.generate_probability_recommendations(analysis_result)

        return analysis_result

    def calculate_event_probability(self, event: ProbabilityEvent) -> Dict[str, Any]:
        """개별 이벤트 확률 계산"""
        # 베이즈 정리 적용
        prior_probability = event.probability

        # 이전 데이터 기반 조건부 확률
        conditional_prob = self.get_conditional_probability(event.category, event.name)

        # 베이즈 업데이트
        updated_probability = self.bayesian_update(prior_probability, conditional_prob)

        # 불확실성 계산
        uncertainty = 1 - self.calculate_certainty_factor(event.probability, event.impact)

        return {
            "prior_probability": prior_probability,
            "conditional_probability": conditional_prob,
            "updated_probability": updated_probability,
            "uncertainty": uncertainty,
            "confidence": 1 - uncertainty,
            "risk_score": self.calculate_risk_score(updated_probability, event.impact)
        }

    def get_conditional_probability(self, category: str, event_name: str) -> float:
        """조건부 확률 가져오기"""
        try:
            category_data = self.bayesian_network["conditional_probabilities"].get(category, {})
            return category_data.get(event_name, 0.5)
        except Exception:
            return 0.5

    def bayesian_update(self, prior: float, conditional: float, evidence: float = 0.5) -> float:
        """베이즈 업데이트"""
        try:
            likelihood = conditional
            marginal = 0.5  # 평균
            posterior = (likelihood * prior) / marginal
            return min(1.0, max(0.0, posterior))
        except Exception:
            return prior

    def calculate_certainty_factor(self, probability: float, impact: float) -> float:
        """확실성 계수 계산"""
        # 확률과 영향도 기반 확실성
        probability_certainty = 1 - abs(0.5 - probability)  # 0.5에 가까울수록 확실성 낮음
        impact_certainty = min(1.0, impact / 10.0)  # 영향도가 클수록 확실성 높음

        return (probability_certainty + impact_certainty) / 2

    def calculate_risk_score(self, probability: float, impact: float) -> float:
        """리스크 점수 계산"""
        return probability * impact

    def perform_statistical_analysis(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """통계적 분석 수행"""
        if not events:
            return {}

        probabilities = [event.probability for event in events]
        impacts = [event.impact for event in events]

        # 기술적 통계량
        mean_prob = sum(probabilities) / len(probabilities)
        std_prob = math.sqrt(sum((p - mean_prob) ** 2 for p in probabilities) / len(probabilities))

        mean_impact = sum(impacts) / len(impacts)
        max_impact = max(impacts)
        min_impact = min(impacts)

        # 분포 분석
        prob_distribution = Counter([int(p * 10) for p in probabilities])
        impact_distribution = Counter([int(i * 2) for i in impacts])

        return {
            "probability_stats": {
                "mean": mean_prob,
                "standard_deviation": std_prob,
                "min": min(probabilities),
                "max": max(probabilities)
            },
            "impact_stats": {
                "mean": mean_impact,
                "min": min_impact,
                "max": max_impact,
                "range": max_impact - min_impact
            },
            "distributions": {
                "probability_distribution": dict(prob_distribution),
                "impact_distribution": dict(impact_distribution)
            },
            "correlation": self.calculate_probability_impact_correlation(probabilities, impacts)
        }

    def calculate_probability_impact_correlation(self, probabilities: List[float], impacts: List[float]) -> float:
        """확률과 영향도 상관계수 계산"""
        if len(probabilities) != len(impacts) or len(probabilities) < 2:
            return 0.0

        n = len(probabilities)
        mean_prob = sum(probabilities) / n
        mean_impact = sum(impacts) / n

        # 상관계수 계산
        numerator = sum((p - mean_prob) * (i - mean_impact) for p, i in zip(probabilities, impacts))
        denominator = math.sqrt(sum((p - mean_prob) ** 2 for p in probabilities) *
                              sum((i - mean_impact) ** 2 for i in impacts))

        return numerator / denominator if denominator > 0 else 0.0

    def assess_risks(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """리스크 평가"""
        risk_assessment = {
            "overall_risk_level": "low",
            "risk_breakdown": defaultdict(list),
            "high_risk_events": [],
            "risk_recommendations": [],
            "risk_mitigation_strategies": []
        }

        # 개별 리스크 평가
        for event in events:
            risk_category = self.categorize_risk(event)
            risk_assessment["risk_breakdown"][risk_category].append({
                "event": event.name,
                "probability": event.probability,
                "impact": event.impact,
                "risk_score": self.calculate_risk_score(event.probability, event.impact)
            })

            # 높은 리스크 이벤트 식별
            if self.calculate_risk_score(event.probability, event.impact) > 0.7:
                risk_assessment["high_risk_events"].append({
                    "name": event.name,
                    "score": self.calculate_risk_score(event.probability, event.impact),
                    "category": risk_category
                })

        # 전체 리스크 레벨 결정
        total_risk_score = sum(self.calculate_risk_score(event.probability, event.impact) for event in events)
        if total_risk_score > 2.0:
            risk_assessment["overall_risk_level"] = "high"
        elif total_risk_score > 1.0:
            risk_assessment["overall_risk_level"] = "medium"
        else:
            risk_assessment["overall_risk_level"] = "low"

        # 리스크 관리 추천
        risk_assessment["risk_recommendations"] = self.generate_risk_recommendations(risk_assessment)
        risk_assessment["risk_mitigation_strategies"] = self.generate_mitigation_strategies(risk_assessment)

        return dict(risk_assessment)

    def categorize_risk(self, event: ProbabilityEvent) -> str:
        """리스크 카테고리 분류"""
        risk_score = self.calculate_risk_score(event.probability, event.impact)

        if risk_score > 0.7:
            return "critical"
        elif risk_score > 0.5:
            return "high"
        elif risk_score > 0.3:
            return "medium"
        else:
            return "low"

    def calculate_expected_values(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """기대값 계산"""
        expected_values = {
            "individual_expected_values": {},
            "total_expected_value": 0.0,
            "positive_expected_value": 0.0,
            "negative_expected_value": 0.0,
            "confidence_adjusted_expected_value": 0.0
        }

        # 개별 기대값 계산
        for event in events:
            expected_value = event.probability * event.impact
            expected_values["individual_expected_values"][event.name] = {
                "expected_value": expected_value,
                "probability": event.probability,
                "impact": event.impact
            }

        # 전체 기대값 계산
        total_expected_value = sum(event.probability * event.impact for event in events)
        expected_values["total_expected_value"] = total_expected_value

        # 양/음 기대값 분리
        positive_ev = sum(event.probability * event.impact for event in events if event.impact > 0)
        negative_ev = sum(event.probability * abs(event.impact) for event in events if event.impact < 0)
        expected_values["positive_expected_value"] = positive_ev
        expected_values["negative_expected_value"] = negative_ev

        # 신뢰도 조정
        avg_confidence = sum(1 - (1 - self.calculate_certainty_factor(event.probability, event.impact)) for event in events) / len(events)
        expected_values["confidence_adjusted_expected_value"] = total_expected_value * avg_confidence

        return expected_values

    def calculate_confidence_level(self, events: List[ProbabilityEvent]) -> float:
        """신뢰도 레벨 계산"""
        if not events:
            return 0.0

        # 기본 신뢰도 요소
        data_quality = 0.8  # 데이터 품질
        model_suitability = 0.7  # 모델 적합성
        historical_accuracy = self.get_historical_accuracy(events)

        # 가중 평균 계산
        confidence_factors = self.decision_models["confidence_calculation"]["confidence_factors"]
        confidence = (
            data_quality * confidence_factors["data_quality"] +
            model_suitability * confidence_factors["model_suitability"] +
            historical_accuracy * confidence_factors["historical_accuracy"]
        )

        return min(1.0, confidence)

    def get_historical_accuracy(self, events: List[ProbabilityEvent]) -> float:
        """역사적 정확도 계산"""
        # 이전 결정 데이터와 비교
        if not self.historical_data["past_decisions"]:
            return 0.5

        # 유사한 이벤트 찾기
        similar_events = []
        for event in events:
            historical_matches = [
                decision for decision in self.historical_data["past_decisions"]
                if self.calculate_event_similarity(event, decision) > 0.7
            ]
            similar_events.extend(historical_matches)

        if not similar_events:
            return 0.5

        # 성공률 계산
        success_count = sum(1 for event in similar_events if event.get("successful", False))
        return success_count / len(similar_events)

    def calculate_event_similarity(self, event: ProbabilityEvent, historical_decision: Dict[str, Any]) -> float:
        """이벤트 유사도 계산"""
        # 이름 유사도
        name_similarity = 1.0 if event.name in historical_decision.get("name", "") else 0.0

        # 확률 유사도
        prob_similarity = 1 - abs(event.probability - historical_decision.get("probability", 0.5))

        # 영향도 유사도
        impact_similarity = 1 - abs(event.impact - historical_decision.get("impact", 1.0))

        return (name_similarity + prob_similarity + impact_similarity) / 3

    def compare_with_historical_data(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """역사적 데이터와 비교"""
        comparison = {
            "similar_scenarios": [],
            "differences": [],
            "trends": [],
            "historical_accuracy": self.get_historical_accuracy(events)
        }

        # 유사한 시나리오 찾기
        for event in events:
            similar_scenarios = [
                scenario for scenario in self.historical_data["past_decisions"]
                if self.calculate_event_similarity(event, scenario) > 0.6
            ]
            if similar_scenarios:
                comparison["similar_scenarios"].append({
                    "event": event.name,
                    "similar_scenarios": len(similar_scenarios),
                    "average_probability": sum(s.get("probability", 0.5) for s in similar_scenarios) / len(similar_scenarios)
                })

        # 차이점 분석
        if events:
            current_avg_prob = sum(event.probability for event in events) / len(events)
            if self.historical_data["success_rates"]:
                historical_avg_prob = sum(self.historical_data["success_rates"].values()) / len(self.historical_data["success_rates"])
                comparison["differences"].append({
                    "probability_difference": current_avg_prob - historical_avg_prob,
                    "impact_difference": sum(event.impact for event in events) / len(events) - 1.0
                })

        return comparison

    def generate_probability_recommendations(self, analysis_result: Dict[str, Any]) -> List[str]:
        """확률 분석 추천 생성"""
        recommendations = []

        # 리스크 기반 추천
        risk_level = analysis_result.get("risk_assessment", {}).get("overall_risk_level", "low")
        if risk_level == "high":
            recommendations.append("높은 리스크 감지 - 추가 분석 및 리스크 완화 전략 필요")
        elif risk_level == "medium":
            recommendations.append("중간 리스크 감지 - 모니터링 권장")

        # 신뢰도 기반 추천
        confidence = analysis_result.get("confidence_level", 0.0)
        if confidence < 0.5:
            recommendations.append("낮은 신뢰도 - 추가 데이터 수집 권장")
        elif confidence > 0.8:
            recommendations.append("높은 신뢰도 - 결정 실행 권장")

        # 기대값 기반 추천
        expected_value = analysis_result.get("expected_value_analysis", {}).get("total_expected_value", 0.0)
        if expected_value > 1.0:
            recommendations.append("긍정적인 기대값 - 결정 실행을 고려해보세요")
        elif expected_value < 0.0:
            recommendations.append("부정적인 기대값 - 결정 재검토 권장")

        # 분포 기반 추천
        stats = analysis_result.get("statistical_analysis", {})
        if "probability_stats" in stats:
            prob_std = stats["probability_stats"].get("standard_deviation", 0)
            if prob_std > 0.3:
                recommendations.append("확률 분산이 큼 - 추가적인 불확실성 관리 필요")

        return recommendations

    def generate_risk_recommendations(self, risk_assessment: Dict[str, Any]) -> List[str]:
        """리스크 관리 추천 생성"""
        recommendations = []

        if risk_assessment["high_risk_events"]:
            recommendations.append(f"높은 리스크 이벤트 {len(risk_assessment['high_risk_events'])}개 즉시 관리 필요")

        for risk_category, events in risk_assessment["risk_breakdown"].items():
            if len(events) > 2:
                recommendations.append(f"{risk_category} 카테고리 리스크 집중 관리 필요")

        if risk_assessment["overall_risk_level"] == "critical":
            recommendations.append("위험한 상황 - 대안 검토 및 실행 계획 재수립")

        return recommendations

    def generate_mitigation_strategies(self, risk_assessment: Dict[str, Any]) -> List[str]:
        """리스크 완화 전략 생성"""
        strategies = []

        # 기본 완화 전략
        mitigation_strategies = {
            "critical": ["프로젝트 중단 또는 완전 재설계", "외부 전문가 자문"],
            "high": ["추가 리소스 할당", "리스크 분산", "완화 조치 즉시 실행"],
            "medium": ["정기적인 모니터링", "경계 설정", "대비 계획 수립"],
            "low": ["정기적 검토", "문서화", "지속적 개선"]
        }

        for risk_category in risk_assessment["risk_breakdown"].keys():
            if risk_category in mitigation_strategies:
                strategies.extend(mitigation_strategies[risk_category])

        return strategies

    def make_decision(self, options: List[Dict[str, Any]], context: Dict[str, Any] = None) -> Dict[str, Any]:
        """의사결정 수행"""
        if context is None:
            context = {}

        decision_result = {
            "decision_type": "probabilistic_decision",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "options_evaluated": len(options),
            "best_option": None,
            "decision_rationale": "",
            "confidence": 0.0,
            "risks": [],
            "recommendations": []
        }

        # 옵션 변환
        decision_options = []
        for option_data in options:
            option = DecisionOption(
                name=option_data.get("name", "unknown"),
                expected_value=option_data.get("expected_value", 0.0),
                confidence=option_data.get("confidence", 0.5),
                risks=[],
                benefits=[]
            )
            decision_options.append(option)

        # 각 옵션 평가
        evaluated_options = []
        for option in decision_options:
            evaluation = self.evaluate_option(option, context)
            evaluated_options.append(evaluation)

        # 최적 옵션 선택
        best_option = self.select_best_option(evaluated_options)
        decision_result["best_option"] = best_option
        decision_result["confidence"] = best_option.get("confidence", 0.0)
        decision_result["decision_rationale"] = self.generate_decision_rationale(best_option, evaluated_options)

        # 리스크 평가
        decision_result["risks"] = self.assess_decision_risks(best_option)

        # 추천 생성
        decision_result["recommendations"] = self.generate_decision_recommendations(decision_result)

        return decision_result

    def evaluate_option(self, option: DecisionOption, context: Dict[str, Any]) -> Dict[str, Any]:
        """개별 옵션 평가"""
        evaluation = {
            "name": option.name,
            "score": 0.0,
            "confidence": option.confidence,
            "weighted_score": 0.0,
            "factors": {}
        }

        # 기대값 점수
        evaluation["factors"]["expected_value"] = option.expected_value
        evaluation["score"] += option.expected_value * 0.4

        # 신뢰도 점수
        evaluation["factors"]["confidence"] = option.confidence
        evaluation["score"] += option.confidence * 0.3

        # 리스크 점수 (리스크가 많을수록 점수 감소)
        risk_score = sum(risk.impact * risk.probability for risk in option.risks)
        evaluation["factors"]["risk_penalty"] = risk_score
        evaluation["score"] -= risk_score * 0.2

        # 이점 점수
        benefit_score = sum(benefit.impact * benefit.probability for benefit in option.benefits)
        evaluation["factors"]["benefit_bonus"] = benefit_score
        evaluation["score"] += benefit_score * 0.1

        # 가중 점수 계산
        weights = context.get("weights", {"expected_value": 0.4, "confidence": 0.3, "risk_penalty": 0.2, "benefit_bonus": 0.1})
        evaluation["weighted_score"] = (
            evaluation["factors"]["expected_value"] * weights.get("expected_value", 0.4) +
            evaluation["factors"]["confidence"] * weights.get("confidence", 0.3) +
            evaluation["factors"]["risk_penalty"] * weights.get("risk_penalty", 0.2) +
            evaluation["factors"]["benefit_bonus"] * weights.get("benefit_bonus", 0.1)
        )

        return evaluation

    def select_best_option(self, evaluated_options: List[Dict[str, Any]]) -> Dict[str, Any]:
        """최적 옵션 선택"""
        return max(evaluated_options, key=lambda x: x["weighted_score"])

    def generate_decision_rationale(self, best_option: Dict[str, Any], all_options: List[Dict[str, Any]]) -> str:
        """의사결정 근거 생성"""
        rationale = f"선택된 옵션: {best_option['name']}\n"
        rationale += f"가중 점수: {best_option['weighted_score']:.3f}\n"
        rationale += f"신뢰도: {best_option['confidence']:.3f}\n"

        # 주요 결정 요인
        factors = best_option.get("factors", {})
        if factors.get("expected_value", 0) > 0:
            rationale += f"긍정적 기대값: {factors['expected_value']:.3f}\n"

        if factors.get("risk_penalty", 0) > 0:
            rationale += f"리스크 요인: {factors['risk_penalty']:.3f}\n"

        return rationale

    def assess_decision_risks(self, best_option: Dict[str, Any]) -> List[Dict[str, Any]]:
        """의사결정 리스크 평가"""
        risks = []

        # 점수 기반 리스크
        if best_option["weighted_score"] < 0.3:
            risks.append({
                "type": "low_score_risk",
                "description": "낮은 점수로 인한 의사결정 불확실성",
                "severity": "medium",
                "mitigation": "추가 분석 수행"
            })

        # 신뢰도 기반 리스크
        if best_option["confidence"] < 0.5:
            risks.append({
                "type": "low_confidence_risk",
                "description": "낮은 신뢰도로 인한 의사결정 리스크",
                "severity": "high",
                "mitigation": "추가 데이터 수집"
            })

        return risks

    def generate_decision_recommendations(self, decision_result: Dict[str, Any]) -> List[str]:
        """의사결정 추천 생성"""
        recommendations = []

        # 기본 추천
        if decision_result["confidence"] > 0.8:
            recommendations.append("높은 신뢰도 - 결정 실행을 권장합니다")
        elif decision_result["confidence"] > 0.6:
            recommendations.append("중간 신뢰도 - 추가 검토 후 결정 권장")
        else:
            recommendations.append("낮은 신뢰도 - 추가 분석 권장")

        # 리스크 기반 추천
        risks = decision_result.get("risks", [])
        if risks:
            recommendations.append(f"{len(risks)}개 리스크 감지 - 관리 전략 수립 필요")

        # 시간 기반 추천
        if "time_sensitivity" in decision_result.get("context", {}):
            recommendations.append("시간 민감성 - 신속한 결정 필요")

        return recommendations


def analyze_probabilities(events: List[Dict[str, Any]]) -> Dict[str, Any]:
    """확률적 추론 엔진으로 확률 분석"""
    engine = ProbabilisticReasoningEngine()
    return engine.analyze_probability(events)


def make_probabilistic_decision(options: List[Dict[str, Any]], context: Dict[str, Any] = None) -> Dict[str, Any]:
    """확률적 추론 엔진으로 의사결정"""
    engine = ProbabilisticReasoningEngine()
    return engine.make_decision(options, context)


def get_probabilistic_engine_status() -> Dict[str, Any]:
    """확률적 추론 엔진 상태 반환"""
    engine = ProbabilisticReasoningEngine()
    return {
        "historical_data_size": len(engine.historical_data["past_decisions"]),
        "bayesian_network_nodes": len(engine.bayesian_network["nodes"]),
        "decision_models_loaded": len(engine.decision_models),
        "confidence_threshold": engine.decision_models["confidence_calculation"]["confidence_thresholds"]
    }


# 표준 Skill 인터페이스 구현
def main() -> None:
    """Skill 메인 함수"""
    try:
        # 인자 파싱
        if len(sys.argv) < 2:
            print(json.dumps({
                "error": "Usage: python3 probabilistic_reasoning_engine.py <action> [args...]"
            }))
            sys.exit(1)

        action = sys.argv[1]

        if action == "analyze":
            if len(sys.argv) < 3:
                print(json.dumps({
                    "error": "Usage: python3 probabilistic_reasoning_engine.py analyze <events_json>"
                }))
                sys.exit(1)

            try:
                events = json.loads(sys.argv[2])
            except json.JSONDecodeError:
                print(json.dumps({
                    "error": "Invalid JSON format for events"
                }))
                sys.exit(1)

            result = analyze_probabilities(events)

        elif action == "decide":
            if len(sys.argv) < 3:
                print(json.dumps({
                    "error": "Usage: python3 probabilistic_reasoning_engine.py decide <options_json> [context_json]"
                }))
                sys.exit(1)

            try:
                options = json.loads(sys.argv[2])
                context = json.loads(sys.argv[3]) if len(sys.argv) > 3 else None
            except json.JSONDecodeError:
                print(json.dumps({
                    "error": "Invalid JSON format for options or context"
                }))
                sys.exit(1)

            result = make_probabilistic_decision(options, context)

        elif action == "status":
            result = get_probabilistic_engine_status()

        else:
            print(json.dumps({
                "error": f"Unknown action: {action}. Available actions: analyze, decide, status"
            }))
            sys.exit(1)

        # 결과 출력
        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        error_result = {
            "error": f"Probabilistic reasoning engine failed: {str(e)}",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()