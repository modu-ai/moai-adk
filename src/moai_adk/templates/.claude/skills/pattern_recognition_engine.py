#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-002 | @SPEC:SKILL-PATTERN-RECOGNITION-ENGINE-001 | @TEST: tests/skills/test_pattern_recognition_engine.py
"""Pattern Recognition Engine Skill

Advanced engine for pattern recognition. Identifies and analyzes various types of patterns:
1. Code patterns
2. Execution patterns
3. Error patterns
4. Performance patterns
5. User behavior patterns

Usage:
    Skill("pattern_recognition_engine")
"""

import json
import re
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Set
from collections import defaultdict, Counter


class PatternRecognitionEngine:
    """Advanced pattern recognition engine"""

    def __init__(self):
        self.pattern_history = []
        self.pattern_database = self.load_pattern_database()
        self.analysis_config = {
            "min_pattern_length": 3,
            "min_occurrences": 2,
            "confidence_threshold": 0.7,
            "pattern_types": ["code", "execution", "error", "performance", "behavior"]
        }

    def load_pattern_database(self) -> Dict[str, Any]:
        """Load pattern database"""
        try:
            database_file = Path(".moai/research/patterns/pattern_database.json")
            if database_file.exists():
                with open(database_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return {
            "known_patterns": {
                "code_patterns": {
                    "function_length": {
                        "description": "Function length pattern",
                        "threshold": {"min": 1, "max": 30},
                        "recommendation": "Keep functions within 15-20 lines"
                    },
                    "nesting_level": {
                        "description": "Nesting level pattern",
                        "threshold": {"min": 1, "max": 3},
                        "recommendation": "Keep nesting level within 3 levels"
                    },
                    "parameter_count": {
                        "description": "Parameter count pattern",
                        "threshold": {"min": 1, "max": 7},
                        "recommendation": "Keep parameters to 5 or fewer"
                    }
                },
                "error_patterns": {
                    "null_pointer": {
                        "description": "Null pointer pattern",
                        "indicators": ["None", "null", "undefined"],
                        "recommendation": "Add pre-validation checks"
                    },
                    "timeout": {
                        "description": "Timeout pattern",
                        "indicators": ["timeout", "timeouterror"],
                        "recommendation": "Optimize timeout settings"
                    }
                },
                "performance_patterns": {
                    "memory_growth": {
                        "description": "Memory growth pattern",
                        "indicators": ["memory", "heap", "allocation"],
                        "recommendation": "Check for memory leaks"
                    },
                    "cpu_intensive": {
                        "description": "CPU intensive pattern",
                        "indicators": ["cpu", "processor", "compute"],
                        "recommendation": "Apply caching or async processing"
                    }
                }
            },
            "pattern_weights": {
                "code_pattern": 0.3,
                "error_pattern": 0.4,
                "performance_pattern": 0.3
            }
        }

    def analyze_patterns(self, data: str, data_type: str = "general") -> Dict[str, Any]:
        """Perform pattern analysis"""
        analysis_result = {
            "data_type": data_type,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "patterns_detected": [],
            "pattern_statistics": {},
            "confidence_scores": {},
            "recommendations": [],
            "risk_assessment": {},
            "similar_patterns": []
        }

        # Analyze by data type
        if data_type == "code":
            analysis_result["patterns_detected"] = self.detect_code_patterns(data)
        elif data_type == "error":
            analysis_result["patterns_detected"] = self.detect_error_patterns(data)
        elif data_type == "performance":
            analysis_result["patterns_detected"] = self.detect_performance_patterns(data)
        elif data_type == "execution":
            analysis_result["patterns_detected"] = self.detect_execution_patterns(data)
        else:
            analysis_result["patterns_detected"] = self.detect_general_patterns(data)

        # Calculate pattern statistics
        analysis_result["pattern_statistics"] = self.calculate_pattern_statistics(analysis_result["patterns_detected"])

        # Calculate confidence scores
        analysis_result["confidence_scores"] = self.calculate_confidence_scores(analysis_result["patterns_detected"])

        # Generate recommendations
        analysis_result["recommendations"] = self.generate_recommendations(analysis_result["patterns_detected"])

        # Assess risks
        analysis_result["risk_assessment"] = self.assess_pattern_risks(analysis_result["patterns_detected"])

        # Find similar patterns
        analysis_result["similar_patterns"] = self.find_similar_patterns(analysis_result["patterns_detected"])

        # Update pattern history
        self.update_pattern_history(analysis_result)

        return analysis_result

    def detect_code_patterns(self, code: str) -> List[Dict[str, Any]]:
        """Detect code patterns"""
        patterns = []

        # Function length pattern
        function_pattern = re.compile(r'def\s+\w+\([^)]*\)\s*:.*?(?=def|\Z)', re.DOTALL)
        for match in function_pattern.finditer(code):
            function_content = match.group()
            lines = function_content.split('\n')
            line_count = len([line for line in lines if line.strip()])

            if line_count > self.analysis_config["min_occurrences"]:
                patterns.append({
                    "pattern_type": "code_function_length",
                    "description": f"Long function pattern ({line_count} lines)",
                    "severity": "warning" if line_count < 50 else "error",
                    "details": {
                        "line_count": line_count,
                        "function_name": self.extract_function_name(match.group())
                    },
                    "known_solution": "Recommend function splitting or refactoring"
                })

        # Nesting level pattern
        nesting_pattern = re.compile(r'(\s+)(if|for|while|try|with)', re.MULTILINE)
        nesting_counts = []
        current_nesting = 0

        for match in nesting_pattern.finditer(code):
            indent_level = len(match.group(1)) // 4
            if indent_level > current_nesting:
                current_nesting = indent_level
            else:
                current_nesting = indent_level

            if current_nesting > 3:
                patterns.append({
                    "pattern_type": "code_nesting_level",
                    "description": f"High nesting level ({current_nesting} levels)",
                    "severity": "warning",
                    "details": {
                        "nesting_level": current_nesting,
                        "line_number": code[:match.start()].count('\n') + 1
                    },
                    "known_solution": "Simplify nested structure"
                })

        # Parameter count pattern
        param_pattern = re.compile(r'def\s+(\w+)\(([^)]+)\)')
        for match in param_pattern.finditer(code):
            params = match.group(2).split(',')
            param_count = len([p.strip() for p in params if p.strip()])

            if param_count > 7:
                patterns.append({
                    "pattern_type": "code_parameter_count",
                    "description": f"High parameter count pattern ({param_count} parameters)",
                    "severity": "warning",
                    "details": {
                        "param_count": param_count,
                        "function_name": match.group(1)
                    },
                    "known_solution": "Group parameters or use object"
                })

        return patterns

    def detect_error_patterns(self, error_data: str) -> List[Dict[str, Any]]:
        """Detect error patterns"""
        patterns = []

        # Null pointer pattern
        null_pattern = re.compile(r'None|null|undefined|nullpointer|nullreference', re.IGNORECASE)
        if null_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_null_pointer",
                "description": "Null pointer related error pattern",
                "severity": "high",
                "details": {
                    "indicators": null_pattern.findall(error_data)
                },
                "known_solution": "Add pre-validation checks and null handling logic"
            })

        # Timeout pattern
        timeout_pattern = re.compile(r'timeout|timeouterror|timedout', re.IGNORECASE)
        if timeout_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_timeout",
                "description": "Timeout related error pattern",
                "severity": "medium",
                "details": {
                    "indicators": timeout_pattern.findall(error_data)
                },
                "known_solution": "Optimize timeout settings and retry logic"
            })

        # Memory related pattern
        memory_pattern = re.compile(r'memory|outofmemory|heap|stackoverflow', re.IGNORECASE)
        if memory_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_memory",
                "description": "Memory related error pattern",
                "severity": "high",
                "details": {
                    "indicators": memory_pattern.findall(error_data)
                },
                "known_solution": "Improve memory management and garbage collection optimization"
            })

        # Network related pattern
        network_pattern = re.compile(r'connection|network|timeout|unreachable', re.IGNORECASE)
        if network_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_network",
                "description": "Network related error pattern",
                "severity": "medium",
                "details": {
                    "indicators": network_pattern.findall(error_data)
                },
                "known_solution": "Add network exception handling and retry mechanism"
            })

        return patterns

    def detect_performance_patterns(self, performance_data: str) -> List[Dict[str, Any]]:
        """Detect performance patterns"""
        patterns = []

        # O(n^2) complexity pattern
        quadratic_pattern = re.compile(r'nested.*loop|nested.*for|nested.*while', re.IGNORECASE)
        if quadratic_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_quadratic",
                "description": "O(n^2) complexity pattern",
                "severity": "warning",
                "details": {
                    "indicators": quadratic_pattern.findall(performance_data)
                },
                "known_solution": "Improve algorithm or apply caching"
            })

        # Memory leak pattern
        leak_pattern = re.compile(r'memory.*leak|leak.*memory|growth.*memory|memory.*growth', re.IGNORECASE)
        if leak_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_memory_leak",
                "description": "Memory leak pattern",
                "severity": "high",
                "details": {
                    "indicators": leak_pattern.findall(performance_data)
                },
                "known_solution": "Check memory management and object lifecycle"
            })

        # CPU intensive pattern
        cpu_pattern = re.compile(r'cpu.*intensive|compute.*heavy|processor.*load', re.IGNORECASE)
        if cpu_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_cpu_intensive",
                "description": "CPU intensive pattern",
                "severity": "medium",
                "details": {
                    "indicators": cpu_pattern.findall(performance_data)
                },
                "known_solution": "Apply async processing or distributed computing"
            })

        # I/O bottleneck pattern
        io_pattern = re.compile(r'io.*bottleneck|file.*io|disk.*io|network.*io', re.IGNORECASE)
        if io_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_io_bottleneck",
                "description": "I/O bottleneck pattern",
                "severity": "medium",
                "details": {
                    "indicators": io_pattern.findall(performance_data)
                },
                "known_solution": "Apply I/O caching or async I/O"
            })

        return patterns

    def detect_execution_patterns(self, execution_data: str) -> List[Dict[str, Any]]:
        """Detect execution patterns"""
        patterns = []

        # Repeated execution pattern
        repeated_pattern = re.compile(r'execute.*repeat|repeat.*execute|loop.*execute', re.IGNORECASE)
        if repeated_pattern.search(execution_data):
            patterns.append({
                "pattern_type": "execution_repeated",
                "description": "Repeated execution pattern",
                "severity": "info",
                "details": {
                    "indicators": repeated_pattern.findall(execution_data)
                },
                "known_solution": "Apply batch processing or caching"
            })

        # Sequential execution pattern
        sequential_pattern = re.compile(r'sequential.*execute|execute.*sequential|serial.*execute', re.IGNORECASE)
        if sequential_pattern.search(execution_data):
            patterns.append({
                "pattern_type": "execution_sequential",
                "description": "Sequential execution pattern",
                "severity": "warning",
                "details": {
                    "indicators": sequential_pattern.findall(execution_data)
                },
                "known_solution": "Apply parallel processing or pipeline"
            })

        # Delayed execution pattern
        delayed_pattern = re.compile(r'delay.*execute|execute.*delay|late.*execute', re.IGNORECASE)
        if delayed_pattern.search(execution_data):
            patterns.append({
                "pattern_type": "execution_delayed",
                "description": "Delayed execution pattern",
                "severity": "medium",
                "details": {
                    "indicators": delayed_pattern.findall(execution_data)
                },
                "known_solution": "Optimize execution plan or predictive execution"
            })

        return patterns

    def detect_general_patterns(self, data: str) -> List[Dict[str, Any]]:
        """Detect general patterns"""
        patterns = []

        # Keyword frequency pattern
        common_keywords = ["error", "warning", "exception", "fail", "success", "timeout"]
        keyword_counts = defaultdict(int)

        for keyword in common_keywords:
            count = len(re.findall(rf'\b{keyword}\b', data, re.IGNORECASE))
            if count > self.analysis_config["min_occurrences"]:
                keyword_counts[keyword] = count

        if keyword_counts:
            patterns.append({
                "pattern_type": "keyword_frequency",
                "description": "Keyword frequency pattern",
                "severity": "info",
                "details": {
                    "keyword_counts": dict(keyword_counts),
                    "total_occurrences": sum(keyword_counts.values())
                },
                "known_solution": "Identify improvement areas through keyword analysis"
            })

        # Line length pattern
        data_lines = data.split('\n')
        avg_line_length = sum(len(line) for line in data_lines) / len(data_lines) if data_lines else 0

        if avg_line_length > 100:
            patterns.append({
                "pattern_type": "line_length",
                "description": "Long line pattern",
                "severity": "warning",
                "details": {
                    "avg_line_length": avg_line_length,
                    "max_line_length": max(len(line) for line in data_lines)
                },
                "known_solution": "Adjust line length and apply code formatting"
            })

        return patterns

    def calculate_pattern_statistics(self, patterns: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Calculate pattern statistics"""
        if not patterns:
            return {"total_patterns": 0, "severity_distribution": {}}

        severity_counts = defaultdict(int)
        pattern_type_counts = defaultdict(int)
        total_severity_score = 0

        for pattern in patterns:
            severity_counts[pattern.get("severity", "info")] += 1
            pattern_type_counts[pattern.get("pattern_type", "unknown")] += 1
            severity_score = {"error": 3, "high": 3, "warning": 2, "medium": 2, "info": 1}
            total_severity_score += severity_score.get(pattern.get("severity", "info"), 1)

        return {
            "total_patterns": len(patterns),
            "severity_distribution": dict(severity_counts),
            "pattern_type_distribution": dict(pattern_type_counts),
            "average_severity_score": total_severity_score / len(patterns) if patterns else 0
        }

    def calculate_confidence_scores(self, patterns: List[Dict[str, Any]]) -> Dict[str, float]:
        """Calculate confidence scores"""
        confidence_scores = {}

        for i, pattern in enumerate(patterns):
            base_confidence = self.analysis_config["confidence_threshold"]

            # Adjust by severity
            severity_multiplier = {
                "error": 1.2,
                "high": 1.1,
                "warning": 1.0,
                "medium": 0.9,
                "info": 0.8
            }

            multiplier = severity_multiplier.get(pattern.get("severity", "info"), 1.0)
            confidence = min(1.0, base_confidence * multiplier)

            # Additional adjustment by pattern type
            type_bonus = {
                "code_function_length": 0.1,
                "code_nesting_level": 0.1,
                "error_null_pointer": 0.2,
                "error_timeout": 0.15,
                "performance_memory_leak": 0.2
            }

            bonus = type_bonus.get(pattern.get("pattern_type"), 0)
            confidence = min(1.0, confidence + bonus)

            confidence_scores[f"pattern_{i}"] = confidence

        return confidence_scores

    def generate_recommendations(self, patterns: List[Dict[str, Any]]) -> List[str]:
        """Generate improvement recommendations"""
        recommendations = []

        # Severity-based recommendations
        high_severity_patterns = [p for p in patterns if p.get("severity") in ["error", "high"]]
        if high_severity_patterns:
            recommendations.append(f"{len(high_severity_patterns)} critical patterns need priority resolution")

        # Pattern type-based recommendations
        code_patterns = [p for p in patterns if p.get("pattern_type", "").startswith("code_")]
        if code_patterns:
            recommendations.append("Code style and structure improvements needed")

        error_patterns = [p for p in patterns if p.get("pattern_type", "").startswith("error_")]
        if error_patterns:
            recommendations.append("Error handling mechanism improvements needed")

        performance_patterns = [p for p in patterns if p.get("pattern_type", "").startswith("performance_")]
        if performance_patterns:
            recommendations.append("Performance optimization needed")

        # General recommendations
        if not patterns:
            recommendations.append("No patterns detected - current state is good")
        else:
            recommendations.append("Regular pattern inspection and improvement recommended")

        return recommendations

    def assess_pattern_risks(self, patterns: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Assess pattern risks"""
        risk_level = {
            "low": 0,
            "medium": 0,
            "high": 0,
            "critical": 0
        }

        for pattern in patterns:
            severity = pattern.get("severity", "info")
            if severity == "error":
                risk_level["critical"] += 1
            elif severity == "high":
                risk_level["high"] += 1
            elif severity == "warning":
                risk_level["medium"] += 1
            elif severity == "medium":
                risk_level["medium"] += 1

        overall_risk = "low"
        if risk_level["critical"] > 0:
            overall_risk = "critical"
        elif risk_level["high"] > 0:
            overall_risk = "high"
        elif risk_level["medium"] > 0:
            overall_risk = "medium"

        return {
            "risk_level": overall_risk,
            "risk_counts": risk_level,
            "total_patterns": len(patterns)
        }

    def find_similar_patterns(self, patterns: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Find similar patterns"""
        similar_patterns = []

        for pattern in patterns:
            pattern_type = pattern.get("pattern_type")

            # Search for similar patterns in existing database
            for known_category, known_patterns in self.pattern_database.get("known_patterns", {}).items():
                if pattern_type in known_patterns:
                    similar_patterns.append({
                        "type": known_category,
                        "pattern_name": pattern_type,
                        "description": known_patterns[pattern_type]["description"],
                        "solution": known_patterns[pattern_type]["recommendation"]
                    })

        return similar_patterns

    def update_pattern_history(self, analysis_result: Dict[str, Any]) -> None:
        """Update pattern history"""
        self.pattern_history.append({
            "timestamp": analysis_result["timestamp"],
            "data_type": analysis_result["data_type"],
            "pattern_count": len(analysis_result["patterns_detected"]),
            "risk_level": analysis_result["risk_assessment"]["risk_level"]
        })

        # Keep maximum 100 entries
        if len(self.pattern_history) > 100:
            self.pattern_history = self.pattern_history[-100:]

    def extract_function_name(self, function_match: str) -> str:
        """Extract function name"""
        match = re.search(r'def\s+(\w+)', function_match)
        return match.group(1) if match else "unknown"


def analyze_patterns_with_engine(data: str, data_type: str = "general") -> Dict[str, Any]:
    """Analyze data with pattern recognition engine"""
    engine = PatternRecognitionEngine()
    return engine.analyze_patterns(data, data_type)


def get_pattern_database() -> Dict[str, Any]:
    """Return pattern database"""
    engine = PatternRecognitionEngine()
    return engine.pattern_database


def get_pattern_history() -> List[Dict[str, Any]]:
    """Return pattern history"""
    engine = PatternRecognitionEngine()
    return engine.pattern_history


# Standard Skill interface implementation
def main() -> None:
    """Skill main function"""
    try:
        # Parse arguments
        if len(sys.argv) < 2:
            print(json.dumps({
                "error": "Usage: python3 pattern_recognition_engine.py <data> [data_type:general|code|error|performance|execution]"
            }))
            sys.exit(1)

        data = sys.argv[1]
        data_type = sys.argv[2] if len(sys.argv) > 2 else "general"

        # Validate data type
        if data_type not in ["general", "code", "error", "performance", "execution"]:
            print(json.dumps({
                "error": f"Invalid data_type: {data_type}. Must be: general, code, error, performance, execution"
            }))
            sys.exit(1)

        # Execute pattern recognition analysis
        result = analyze_patterns_with_engine(data, data_type)

        # Output results
        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        error_result = {
            "error": f"Pattern recognition analysis failed: {str(e)}",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()