#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-004 | @SPEC:SKILL-PROBABILISTIC-REASONING-ENGINE-001 | @TEST: tests/skills/test_probabilistic_reasoning_engine.py
"""Probabilistic Reasoning Engine Skill

Probabilistic reasoning engine. Features for making rational decisions in uncertain situations:
1. Probability calculation
2. Risk assessment
3. Expected value analysis
4. Statistical reasoning
5. Decision trees

Usage:
    Skill("probabilistic_reasoning_engine")
"""

import json
import math
import random
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Union
from collections import defaultdict, Counter
from dataclasses import dataclass


@dataclass
class ProbabilityEvent:
    """Probability event data class"""
    name: str
    probability: float
    impact: float
    category: str


@dataclass
class DecisionOption:
    """Decision option data class"""
    name: str
    expected_value: float
    confidence: float
    risks: List[ProbabilityEvent]
    benefits: List[ProbabilityEvent]


class ProbabilisticReasoningEngine:
    """Probabilistic reasoning engine class"""

    def __init__(self):
        self.historical_data = self.load_historical_data()
        self.bayesian_network = self.load_bayesian_network()
        self.decision_models = self.load_decision_models()

    def load_historical_data(self) -> Dict[str, Any]:
        """Load historical data"""
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
        """Load Bayesian network"""
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
        """Load decision models"""
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
        """Get default risk assessment model"""
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
        """Get default expected value model"""
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
        """Get default confidence model"""
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
        """Perform probability analysis"""
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

        # Convert events
        probability_events = []
        for event_data in events:
            event = ProbabilityEvent(
                name=event_data.get("name", "unknown"),
                probability=float(event_data.get("probability", 0.5)),
                impact=float(event_data.get("impact", 1.0)),
                category=event_data.get("category", "general")
            )
            probability_events.append(event)

        # Calculate individual probabilities
        for event in probability_events:
            analysis_result["probability_calculations"][event.name] = self.calculate_event_probability(event)

        # Statistical analysis
        analysis_result["statistical_analysis"] = self.perform_statistical_analysis(probability_events)

        # Risk assessment
        analysis_result["risk_assessment"] = self.assess_risks(probability_events)

        # Expected value analysis
        analysis_result["expected_value_analysis"] = self.calculate_expected_values(probability_events)

        # Calculate confidence level
        analysis_result["confidence_level"] = self.calculate_confidence_level(probability_events)

        # Historical comparison
        analysis_result["historical_comparison"] = self.compare_with_historical_data(probability_events)

        # Generate recommendations
        analysis_result["recommendations"] = self.generate_probability_recommendations(analysis_result)

        return analysis_result

    def calculate_event_probability(self, event: ProbabilityEvent) -> Dict[str, Any]:
        """Calculate individual event probability"""
        # Apply Bayes' theorem
        prior_probability = event.probability

        # Conditional probability based on historical data
        conditional_prob = self.get_conditional_probability(event.category, event.name)

        # Bayesian update
        updated_probability = self.bayesian_update(prior_probability, conditional_prob)

        # Calculate uncertainty
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
        """Get conditional probability"""
        try:
            category_data = self.bayesian_network["conditional_probabilities"].get(category, {})
            return category_data.get(event_name, 0.5)
        except Exception:
            return 0.5

    def bayesian_update(self, prior: float, conditional: float, evidence: float = 0.5) -> float:
        """Bayesian update"""
        try:
            likelihood = conditional
            marginal = 0.5  # Average
            posterior = (likelihood * prior) / marginal
            return min(1.0, max(0.0, posterior))
        except Exception:
            return prior

    def calculate_certainty_factor(self, probability: float, impact: float) -> float:
        """Calculate certainty factor"""
        # Certainty based on probability and impact
        probability_certainty = 1 - abs(0.5 - probability)  # Lower certainty closer to 0.5
        impact_certainty = min(1.0, impact / 10.0)  # Higher certainty with greater impact

        return (probability_certainty + impact_certainty) / 2

    def calculate_risk_score(self, probability: float, impact: float) -> float:
        """Calculate risk score"""
        return probability * impact

    def perform_statistical_analysis(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """Perform statistical analysis"""
        if not events:
            return {}

        probabilities = [event.probability for event in events]
        impacts = [event.impact for event in events]

        # Descriptive statistics
        mean_prob = sum(probabilities) / len(probabilities)
        std_prob = math.sqrt(sum((p - mean_prob) ** 2 for p in probabilities) / len(probabilities))

        mean_impact = sum(impacts) / len(impacts)
        max_impact = max(impacts)
        min_impact = min(impacts)

        # Distribution analysis
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
        """Calculate probability-impact correlation coefficient"""
        if len(probabilities) != len(impacts) or len(probabilities) < 2:
            return 0.0

        n = len(probabilities)
        mean_prob = sum(probabilities) / n
        mean_impact = sum(impacts) / n

        # Calculate correlation coefficient
        numerator = sum((p - mean_prob) * (i - mean_impact) for p, i in zip(probabilities, impacts))
        denominator = math.sqrt(sum((p - mean_prob) ** 2 for p in probabilities) *
                              sum((i - mean_impact) ** 2 for i in impacts))

        return numerator / denominator if denominator > 0 else 0.0

    def assess_risks(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """Assess risks"""
        risk_assessment = {
            "overall_risk_level": "low",
            "risk_breakdown": defaultdict(list),
            "high_risk_events": [],
            "risk_recommendations": [],
            "risk_mitigation_strategies": []
        }

        # Assess individual risks
        for event in events:
            risk_category = self.categorize_risk(event)
            risk_assessment["risk_breakdown"][risk_category].append({
                "event": event.name,
                "probability": event.probability,
                "impact": event.impact,
                "risk_score": self.calculate_risk_score(event.probability, event.impact)
            })

            # Identify high-risk events
            if self.calculate_risk_score(event.probability, event.impact) > 0.7:
                risk_assessment["high_risk_events"].append({
                    "name": event.name,
                    "score": self.calculate_risk_score(event.probability, event.impact),
                    "category": risk_category
                })

        # Determine overall risk level
        total_risk_score = sum(self.calculate_risk_score(event.probability, event.impact) for event in events)
        if total_risk_score > 2.0:
            risk_assessment["overall_risk_level"] = "high"
        elif total_risk_score > 1.0:
            risk_assessment["overall_risk_level"] = "medium"
        else:
            risk_assessment["overall_risk_level"] = "low"

        # Generate risk recommendations
        risk_assessment["risk_recommendations"] = self.generate_risk_recommendations(risk_assessment)
        risk_assessment["risk_mitigation_strategies"] = self.generate_mitigation_strategies(risk_assessment)

        return dict(risk_assessment)

    def categorize_risk(self, event: ProbabilityEvent) -> str:
        """Categorize risk"""
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
        """Calculate expected values"""
        expected_values = {
            "individual_expected_values": {},
            "total_expected_value": 0.0,
            "positive_expected_value": 0.0,
            "negative_expected_value": 0.0,
            "confidence_adjusted_expected_value": 0.0
        }

        # Calculate individual expected values
        for event in events:
            expected_value = event.probability * event.impact
            expected_values["individual_expected_values"][event.name] = {
                "expected_value": expected_value,
                "probability": event.probability,
                "impact": event.impact
            }

        # Calculate total expected value
        total_expected_value = sum(event.probability * event.impact for event in events)
        expected_values["total_expected_value"] = total_expected_value

        # Separate positive/negative expected values
        positive_ev = sum(event.probability * event.impact for event in events if event.impact > 0)
        negative_ev = sum(event.probability * abs(event.impact) for event in events if event.impact < 0)
        expected_values["positive_expected_value"] = positive_ev
        expected_values["negative_expected_value"] = negative_ev

        # Confidence adjustment
        avg_confidence = sum(1 - (1 - self.calculate_certainty_factor(event.probability, event.impact)) for event in events) / len(events)
        expected_values["confidence_adjusted_expected_value"] = total_expected_value * avg_confidence

        return expected_values

    def calculate_confidence_level(self, events: List[ProbabilityEvent]) -> float:
        """Calculate confidence level"""
        if not events:
            return 0.0

        # Base confidence factors
        data_quality = 0.8  # Data quality
        model_suitability = 0.7  # Model suitability
        historical_accuracy = self.get_historical_accuracy(events)

        # Calculate weighted average
        confidence_factors = self.decision_models["confidence_calculation"]["confidence_factors"]
        confidence = (
            data_quality * confidence_factors["data_quality"] +
            model_suitability * confidence_factors["model_suitability"] +
            historical_accuracy * confidence_factors["historical_accuracy"]
        )

        return min(1.0, confidence)

    def get_historical_accuracy(self, events: List[ProbabilityEvent]) -> float:
        """Calculate historical accuracy"""
        # Compare with previous decision data
        if not self.historical_data["past_decisions"]:
            return 0.5

        # Find similar events
        similar_events = []
        for event in events:
            historical_matches = [
                decision for decision in self.historical_data["past_decisions"]
                if self.calculate_event_similarity(event, decision) > 0.7
            ]
            similar_events.extend(historical_matches)

        if not similar_events:
            return 0.5

        # Calculate success rate
        success_count = sum(1 for event in similar_events if event.get("successful", False))
        return success_count / len(similar_events)

    def calculate_event_similarity(self, event: ProbabilityEvent, historical_decision: Dict[str, Any]) -> float:
        """Calculate event similarity"""
        # Name similarity
        name_similarity = 1.0 if event.name in historical_decision.get("name", "") else 0.0

        # Probability similarity
        prob_similarity = 1 - abs(event.probability - historical_decision.get("probability", 0.5))

        # Impact similarity
        impact_similarity = 1 - abs(event.impact - historical_decision.get("impact", 1.0))

        return (name_similarity + prob_similarity + impact_similarity) / 3

    def compare_with_historical_data(self, events: List[ProbabilityEvent]) -> Dict[str, Any]:
        """Compare with historical data"""
        comparison = {
            "similar_scenarios": [],
            "differences": [],
            "trends": [],
            "historical_accuracy": self.get_historical_accuracy(events)
        }

        # Find similar scenarios
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

        # Analyze differences
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
        """Generate probability analysis recommendations"""
        recommendations = []

        # Risk-based recommendations
        risk_level = analysis_result.get("risk_assessment", {}).get("overall_risk_level", "low")
        if risk_level == "high":
            recommendations.append("High risk detected - additional analysis and risk mitigation strategies needed")
        elif risk_level == "medium":
            recommendations.append("Medium risk detected - monitoring recommended")

        # Confidence-based recommendations
        confidence = analysis_result.get("confidence_level", 0.0)
        if confidence < 0.5:
            recommendations.append("Low confidence - additional data collection recommended")
        elif confidence > 0.8:
            recommendations.append("High confidence - decision execution recommended")

        # Expected value-based recommendations
        expected_value = analysis_result.get("expected_value_analysis", {}).get("total_expected_value", 0.0)
        if expected_value > 1.0:
            recommendations.append("Positive expected value - consider executing decision")
        elif expected_value < 0.0:
            recommendations.append("Negative expected value - decision review recommended")

        # Distribution-based recommendations
        stats = analysis_result.get("statistical_analysis", {})
        if "probability_stats" in stats:
            prob_std = stats["probability_stats"].get("standard_deviation", 0)
            if prob_std > 0.3:
                recommendations.append("High probability variance - additional uncertainty management needed")

        return recommendations

    def generate_risk_recommendations(self, risk_assessment: Dict[str, Any]) -> List[str]:
        """Generate risk management recommendations"""
        recommendations = []

        if risk_assessment["high_risk_events"]:
            recommendations.append(f"{len(risk_assessment['high_risk_events'])} high-risk events require immediate attention")

        for risk_category, events in risk_assessment["risk_breakdown"].items():
            if len(events) > 2:
                recommendations.append(f"{risk_category} category requires focused risk management")

        if risk_assessment["overall_risk_level"] == "critical":
            recommendations.append("Critical situation - review alternatives and revise execution plan")

        return recommendations

    def generate_mitigation_strategies(self, risk_assessment: Dict[str, Any]) -> List[str]:
        """Generate risk mitigation strategies"""
        strategies = []

        # Base mitigation strategies
        mitigation_strategies = {
            "critical": ["Project halt or complete redesign", "External expert consultation"],
            "high": ["Allocate additional resources", "Risk distribution", "Immediate mitigation actions"],
            "medium": ["Regular monitoring", "Set boundaries", "Establish contingency plans"],
            "low": ["Regular review", "Documentation", "Continuous improvement"]
        }

        for risk_category in risk_assessment["risk_breakdown"].keys():
            if risk_category in mitigation_strategies:
                strategies.extend(mitigation_strategies[risk_category])

        return strategies

    def make_decision(self, options: List[Dict[str, Any]], context: Dict[str, Any] = None) -> Dict[str, Any]:
        """Perform decision making"""
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

        # Convert options
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

        # Evaluate each option
        evaluated_options = []
        for option in decision_options:
            evaluation = self.evaluate_option(option, context)
            evaluated_options.append(evaluation)

        # Select best option
        best_option = self.select_best_option(evaluated_options)
        decision_result["best_option"] = best_option
        decision_result["confidence"] = best_option.get("confidence", 0.0)
        decision_result["decision_rationale"] = self.generate_decision_rationale(best_option, evaluated_options)

        # Assess risks
        decision_result["risks"] = self.assess_decision_risks(best_option)

        # Generate recommendations
        decision_result["recommendations"] = self.generate_decision_recommendations(decision_result)

        return decision_result

    def evaluate_option(self, option: DecisionOption, context: Dict[str, Any]) -> Dict[str, Any]:
        """Evaluate individual option"""
        evaluation = {
            "name": option.name,
            "score": 0.0,
            "confidence": option.confidence,
            "weighted_score": 0.0,
            "factors": {}
        }

        # Expected value score
        evaluation["factors"]["expected_value"] = option.expected_value
        evaluation["score"] += option.expected_value * 0.4

        # Confidence score
        evaluation["factors"]["confidence"] = option.confidence
        evaluation["score"] += option.confidence * 0.3

        # Risk score (more risks decrease score)
        risk_score = sum(risk.impact * risk.probability for risk in option.risks)
        evaluation["factors"]["risk_penalty"] = risk_score
        evaluation["score"] -= risk_score * 0.2

        # Benefit score
        benefit_score = sum(benefit.impact * benefit.probability for benefit in option.benefits)
        evaluation["factors"]["benefit_bonus"] = benefit_score
        evaluation["score"] += benefit_score * 0.1

        # Calculate weighted score
        weights = context.get("weights", {"expected_value": 0.4, "confidence": 0.3, "risk_penalty": 0.2, "benefit_bonus": 0.1})
        evaluation["weighted_score"] = (
            evaluation["factors"]["expected_value"] * weights.get("expected_value", 0.4) +
            evaluation["factors"]["confidence"] * weights.get("confidence", 0.3) +
            evaluation["factors"]["risk_penalty"] * weights.get("risk_penalty", 0.2) +
            evaluation["factors"]["benefit_bonus"] * weights.get("benefit_bonus", 0.1)
        )

        return evaluation

    def select_best_option(self, evaluated_options: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Select best option"""
        return max(evaluated_options, key=lambda x: x["weighted_score"])

    def generate_decision_rationale(self, best_option: Dict[str, Any], all_options: List[Dict[str, Any]]) -> str:
        """Generate decision rationale"""
        rationale = f"Selected option: {best_option['name']}\n"
        rationale += f"Weighted score: {best_option['weighted_score']:.3f}\n"
        rationale += f"Confidence: {best_option['confidence']:.3f}\n"

        # Key decision factors
        factors = best_option.get("factors", {})
        if factors.get("expected_value", 0) > 0:
            rationale += f"Positive expected value: {factors['expected_value']:.3f}\n"

        if factors.get("risk_penalty", 0) > 0:
            rationale += f"Risk factor: {factors['risk_penalty']:.3f}\n"

        return rationale

    def assess_decision_risks(self, best_option: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Assess decision risks"""
        risks = []

        # Score-based risk
        if best_option["weighted_score"] < 0.3:
            risks.append({
                "type": "low_score_risk",
                "description": "Decision uncertainty due to low score",
                "severity": "medium",
                "mitigation": "Perform additional analysis"
            })

        # Confidence-based risk
        if best_option["confidence"] < 0.5:
            risks.append({
                "type": "low_confidence_risk",
                "description": "Decision risk due to low confidence",
                "severity": "high",
                "mitigation": "Collect additional data"
            })

        return risks

    def generate_decision_recommendations(self, decision_result: Dict[str, Any]) -> List[str]:
        """Generate decision recommendations"""
        recommendations = []

        # Base recommendations
        if decision_result["confidence"] > 0.8:
            recommendations.append("High confidence - decision execution recommended")
        elif decision_result["confidence"] > 0.6:
            recommendations.append("Medium confidence - additional review recommended before decision")
        else:
            recommendations.append("Low confidence - additional analysis recommended")

        # Risk-based recommendations
        risks = decision_result.get("risks", [])
        if risks:
            recommendations.append(f"{len(risks)} risks detected - management strategy development needed")

        # Time-based recommendations
        if "time_sensitivity" in decision_result.get("context", {}):
            recommendations.append("Time sensitivity - rapid decision needed")

        return recommendations


def analyze_probabilities(events: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Analyze probabilities with probabilistic reasoning engine"""
    engine = ProbabilisticReasoningEngine()
    return engine.analyze_probability(events)


def make_probabilistic_decision(options: List[Dict[str, Any]], context: Dict[str, Any] = None) -> Dict[str, Any]:
    """Make decision with probabilistic reasoning engine"""
    engine = ProbabilisticReasoningEngine()
    return engine.make_decision(options, context)


def get_probabilistic_engine_status() -> Dict[str, Any]:
    """Get probabilistic reasoning engine status"""
    engine = ProbabilisticReasoningEngine()
    return {
        "historical_data_size": len(engine.historical_data["past_decisions"]),
        "bayesian_network_nodes": len(engine.bayesian_network["nodes"]),
        "decision_models_loaded": len(engine.decision_models),
        "confidence_threshold": engine.decision_models["confidence_calculation"]["confidence_thresholds"]
    }


# Standard Skill interface implementation
def main() -> None:
    """Skill main function"""
    try:
        # Parse arguments
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

        # Output result
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