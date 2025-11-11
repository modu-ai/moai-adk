#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-001 | @SPEC:SKILL-SENIOR-ENGINEER-THINKING-001 | @TEST: tests/skills/test_senior_engineer_thinking.py
"""Senior Engineer Thinking Pattern Skill

Implements senior engineer thinking patterns. Integrates 8 advanced analysis strategies:
1. Root Cause Analysis
2. Pattern Recognition
3. Systematic Elimination
4. First Principles Thinking
5. Cross-Domain Analysis
6. Probabilistic Thinking
7. Resource Optimization
8. Continuous Learning

Usage:
    Skill("senior_engineer_thinking")
"""

import json
import re
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple


class SeniorEngineerThinking:
    """Senior Engineer Thinking Pattern Processor"""

    def __init__(self):
        self.research_strategies = {
            "root_cause_analysis": {
                "description": "Root cause analysis",
                "questions": [
                    "What is the fundamental problem?",
                    "What are the root causes?",
                    "What assumptions are we making?",
                    "What evidence supports this?"
                ],
                "techniques": ["5 Whys", "Fishbone Diagram", "Barrier Analysis"]
            },
            "pattern_recognition": {
                "description": "Pattern recognition",
                "questions": [
                    "What patterns do we observe?",
                    "Have we seen this before?",
                    "What are the common elements?",
                    "What trends are emerging?"
                ],
                "techniques": ["Data Clustering", "Correlation Analysis", "Trend Detection"]
            },
            "systematic_elimination": {
                "description": "Systematic elimination",
                "questions": [
                    "What can we eliminate?",
                    "What is unnecessary complexity?",
                    "What assumptions can we test?",
                    "What can be simplified?"
                ],
                "techniques": ["Occam's Razor", "Minimal Viable Product", "Constraint Analysis"]
            },
            "first_principles": {
                "description": "First principles thinking",
                "questions": [
                    "What are the fundamental truths?",
                    "What do we know for certain?",
                    "What if we start from scratch?",
                    "What are the core assumptions?"
                ],
                "techniques": ["Fundamental Analysis", "Deconstruction", "Reconstruction"]
            },
            "cross_domain_analysis": {
                "description": "Cross-domain analysis",
                "questions": [
                    "What can we learn from other domains?",
                    "What parallel problems exist?",
                    "What are the transferable solutions?",
                    "What are analogous systems?"
                ],
                "techniques": ["Analogical Reasoning", "Cross-Industry Benchmarking", "Knowledge Transfer"]
            },
            "probabilistic_thinking": {
                "description": "Probabilistic thinking",
                "questions": [
                    "What are the probabilities?",
                    "What are the worst/best cases?",
                    "What is the expected value?",
                    "What are the risk factors?"
                ],
                "techniques": ["Risk Assessment", "Monte Carlo Simulation", "Expected Value Calculation"]
            },
            "resource_optimization": {
                "description": "Resource optimization",
                "questions": [
                    "Where are we wasting resources?",
                    "What can be automated?",
                    "What are the bottlenecks?",
                    "How can we scale efficiently?"
                ],
                "techniques": ["Cost-Benefit Analysis", "Resource Allocation", "Efficiency Metrics"]
            },
            "continuous_learning": {
                "description": "Continuous learning",
                "questions": [
                    "What have we learned?",
                    "What patterns repeat?",
                    "What improvements can be made?",
                    "What knowledge can be shared?"
                ],
                "techniques": ["Post-mortem Analysis", "Lessons Learned", "Knowledge Sharing"]
            }
        }

        self.knowledge_base = {}
        self.load_historical_knowledge()

    def load_historical_knowledge(self) -> None:
        """Load previous learning data"""
        try:
            knowledge_dir = Path(".moai/research/knowledge/")
            if knowledge_dir.exists():
                for file_path in knowledge_dir.glob("*_insights.json"):
                    try:
                        import json as json_module
                        with open(file_path, 'r', encoding='utf-8') as f:
                            data = json_module.load(f)
                            if "insights_history" in data:
                                self.knowledge_base[file_path.stem] = data
                    except Exception:
                        continue
        except Exception:
            pass

    def analyze_problem(self, problem: str, context: Dict[str, Any] = None) -> Dict[str, Any]:
        """Perform problem analysis"""
        if context is None:
            context = {}

        analysis_result = {
            "problem": problem,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "strategies_applied": [],
            "insights": [],
            "recommendations": [],
            "knowledge_connections": [],
            "risk_assessment": {},
            "next_steps": []
        }

        # Apply each strategy
        for strategy_name, strategy_info in self.research_strategies.items():
            strategy_result = self.apply_strategy(strategy_name, problem, context)
            analysis_result["strategies_applied"].append({
                "name": strategy_name,
                "description": strategy_info["description"],
                "result": strategy_result
            })
            analysis_result["insights"].extend(strategy_result.get("insights", []))
            analysis_result["recommendations"].extend(strategy_result.get("recommendations", []))

        # Knowledge connection analysis
        analysis_result["knowledge_connections"] = self.find_knowledge_connections(problem, context)

        # Risk assessment
        analysis_result["risk_assessment"] = self.assess_risks(problem, context)

        return analysis_result

    def apply_strategy(self, strategy_name: str, problem: str, context: Dict[str, Any]) -> Dict[str, Any]:
        """Apply individual strategy"""
        strategy_info = self.research_strategies[strategy_name]

        result = {
            "strategy": strategy_name,
            "description": strategy_info["description"],
            "questions": [],
            "insights": [],
            "recommendations": [],
            "techniques": strategy_info["techniques"]
        }

        # Generate questions through problem analysis
        for question_template in strategy_info["questions"]:
            question = question_template.replace("What", f"What about '{problem}'")
            question = question.replace("we", "we" if "we" in problem.lower() else "you")
            result["questions"].append(question)

        # Generate strategy-specific insights
        insights = self.generate_strategy_insights(strategy_name, problem, context)
        result["insights"] = insights

        # Generate strategy-specific recommendations
        recommendations = self.generate_strategy_recommendations(strategy_name, problem, context)
        result["recommendations"] = recommendations

        return result

    def generate_strategy_insights(self, strategy_name: str, problem: str, context: Dict[str, Any]) -> List[str]:
        """Generate strategy-specific insights"""
        insights = []

        # Root cause analysis
        if strategy_name == "root_cause_analysis":
            if "bug" in problem.lower() or "error" in problem.lower():
                insights.append("Root cause likely lies in input validation or error handling")
                insights.append("Consider edge cases and boundary conditions")
            elif "performance" in problem.lower():
                insights.append("Root cause likely in algorithmic complexity or resource usage")
                insights.append("Profile memory usage and execution paths")

        # Pattern recognition
        elif strategy_name == "pattern_recognition":
            if "repeat" in problem.lower() or "again" in problem.lower():
                insights.append("Identify patterns in occurrence timing or conditions")
                insights.append("Look for common triggers or shared characteristics")

        # Systematic elimination
        elif strategy_name == "systematic_elimination":
            if "complex" in problem.lower() or "complicated" in problem.lower():
                insights.append("Remove unnecessary layers of abstraction")
                insights.append("Simplify by focusing on core functionality")

        # First principles thinking
        elif strategy_name == "first_principles":
            if "assumption" in problem.lower() or "assume" in problem.lower():
                insights.append("Question all underlying assumptions")
                insights.append("Break down problem to fundamental components")

        # Cross-domain analysis
        elif strategy_name == "cross_domain_analysis":
            if "new" in problem.lower() or "novel" in problem.lower():
                insights.append("Similar problems exist in other domains")
                insights.append("Transfer solutions from analogous fields")

        # Probabilistic thinking
        elif strategy_name == "probabilistic_thinking":
            if "risk" in problem.lower() or "uncertain" in problem.lower():
                insights.append("Quantify probability of different outcomes")
                insights.append("Assess impact of worst-case scenarios")

        # Resource optimization
        elif strategy_name == "resource_optimization":
            if "slow" in problem.lower() or "fast" in problem.lower():
                insights.append("Identify bottlenecks in resource allocation")
                insights.append("Automate repetitive or resource-intensive tasks")

        # Continuous learning
        elif strategy_name == "continuous_learning":
            if "first time" in problem.lower() or "new" in problem.lower():
                insights.append("Document lessons learned for future reference")
                insights.append("Share insights with team members")

        return insights

    def generate_strategy_recommendations(self, strategy_name: str, problem: str, context: Dict[str, Any]) -> List[str]:
        """Generate strategy-specific recommendations"""
        recommendations = []

        # Strategy-specific concrete recommendations
        if strategy_name == "root_cause_analysis":
            recommendations.append("Implement comprehensive logging and monitoring")
            recommendations.append("Create automated root cause analysis tools")

        elif strategy_name == "pattern_recognition":
            recommendations.append("Develop pattern detection algorithms")
            recommendations.append("Create pattern database for historical reference")

        elif strategy_name == "systematic_elimination":
            recommendations.append("Conduct complexity audit")
            recommendations.append("Implement design principles that encourage simplicity")

        elif strategy_name == "first_principles":
            recommendations.append("Create first principles checklist for major decisions")
            recommendations.append("Document all fundamental assumptions explicitly")

        elif strategy_name == "cross_domain_analysis":
            recommendations.append("Establish cross-domain knowledge sharing sessions")
            recommendations.append("Create analogy library for problem-solving")

        elif strategy_name == "probabilistic_thinking":
            recommendations.append("Implement risk assessment frameworks")
            recommendations.append("Create probabilistic models for critical decisions")

        elif strategy_name == "resource_optimization":
            recommendations.append("Implement resource monitoring and alerting")
            recommendations.append("Automate resource allocation based on demand")

        elif strategy_name == "continuous_learning":
            recommendations.append("Create post-mortem process for all major issues")
            recommendations.append("Establish knowledge sharing platform")

        return recommendations

    def find_knowledge_connections(self, problem: str, context: Dict[str, Any]) -> List[str]:
        """Find related knowledge connections"""
        connections = []

        # Connect with previous insights
        for file_name, data in self.knowledge_base.items():
            for entry in data.get("insights_history", []):
                for insight in entry.get("insights", []):
                    if any(keyword.lower() in insight.lower() for keyword in problem.lower().split()):
                        connections.append(f"Related insight from {file_name}: {insight[:100]}...")
                        break

        # Default connections when none exist
        if not connections:
            connections.append("Create new knowledge entry for this problem")
            connections.append("Document insights for future reference")

        return connections

    def assess_risks(self, problem: str, context: Dict[str, Any]) -> Dict[str, Any]:
        """Assess risks"""
        risks = {
            "high_risks": [],
            "medium_risks": [],
            "low_risks": [],
            "mitigation_strategies": []
        }

        # Risk assessment based on problem type
        if "critical" in problem.lower() or "urgent" in problem.lower():
            risks["high_risks"].append("Time pressure may lead to incomplete analysis")
            risks["mitigation_strategies"].append("Allocate dedicated time for thorough analysis")

        if "unknown" in problem.lower() or "uncertain" in problem.lower():
            risks["high_risks"].append("Uncertainty about root causes")
            risks["mitigation_strategies"].append("Implement incremental approach with monitoring")

        if "complex" in problem.lower() or "complicated" in problem.lower():
            risks["medium_risks"].append("Over-engineering the solution")
            risks["mitigation_strategies"].append("Focus on simple, targeted solutions")

        risks["low_risks"].append("Analysis may take longer than expected")
        risks["mitigation_strategies"].append("Set time limits and prioritize key insights")

        return risks


def analyze_problem_with_senior_engineer_thinking(problem: str, context: Dict[str, Any] = None) -> Dict[str, Any]:
    """Analyze given problem with senior engineer thinking patterns"""
    thinker = SeniorEngineerThinking()
    return thinker.analyze_problem(problem, context)


def get_available_strategies() -> List[str]:
    """Return list of available strategies"""
    thinker = SeniorEngineerThinking()
    return list(thinker.research_strategies.keys())


def get_strategy_details(strategy_name: str) -> Dict[str, Any]:
    """Return detailed information for specific strategy"""
    thinker = SeniorEngineerThinking()
    return thinker.research_strategies.get(strategy_name, {})


# Standard Skill interface implementation
def main() -> None:
    """Skill main function"""
    try:
        # Parse arguments
        if len(sys.argv) < 2:
            print(json.dumps({
                "error": "Usage: python3 senior_engineer_thinking.py <problem> [context_json]"
            }))
            sys.exit(1)

        problem = sys.argv[1]
        context = {}

        if len(sys.argv) > 2:
            try:
                context = json.loads(sys.argv[2])
            except json.JSONDecodeError:
                pass

        # Execute senior engineer thinking analysis
        result = analyze_problem_with_senior_engineer_thinking(problem, context)

        # Output result
        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        error_result = {
            "error": f"Senior engineer thinking analysis failed: {str(e)}",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()