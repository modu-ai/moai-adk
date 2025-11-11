#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-002 | @SPEC:SKILL-PATTERN-RECOGNITION-ENGINE-001 | @TEST: tests/skills/test_pattern_recognition_engine.py
"""Pattern Recognition Engine Skill

패턴 인식을 위한 고급 엔진. 다양한 유형의 패턴을 식별하고 분석:
1. 코드 패턴
2. 실행 패턴
3. 에러 패턴
4. 성능 패턴
5. 사용자 행동 패턴

사용법:
    Skill("pattern_recognition_engine")
"""

import json
import re
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Set
from collections import defaultdict, Counter


class PatternRecognitionEngine:
    """고급 패턴 인식 엔진"""

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
        """패턴 데이터베이스 로드"""
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
                        "description": "함수 길이 패턴",
                        "threshold": {"min": 1, "max": 30},
                        "recommendation": "함수는 15-20라인 이내로 유지"
                    },
                    "nesting_level": {
                        "description": "중첩 레벨 패턴",
                        "threshold": {"min": 1, "max": 3},
                        "recommendation": "중첩 레벨은 3레벨 이내로 유지"
                    },
                    "parameter_count": {
                        "description": "매개변수 수 패턴",
                        "threshold": {"min": 1, "max": 7},
                        "recommendation": "매개변수는 5개 이하로 유지"
                    }
                },
                "error_patterns": {
                    "null_pointer": {
                        "description": "널 포인터 패턴",
                        "indicators": ["None", "null", "undefined"],
                        "recommendation": "사전 유효성 검사 추가"
                    },
                    "timeout": {
                        "description": "타임아웃 패턴",
                        "indicators": ["timeout", "timeouterror"],
                        "recommendation": "타임아웃 설정 최적화"
                    }
                },
                "performance_patterns": {
                    "memory_growth": {
                        "description": "메모리 증가 패턴",
                        "indicators": ["memory", "heap", "allocation"],
                        "recommendation": "메모리 누수 점검"
                    },
                    "cpu_intensive": {
                        "description": "CPU 집약적 패턴",
                        "indicators": ["cpu", "processor", "compute"],
                        "recommendation": "캐싱 또는 비동기화 적용"
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
        """패턴 분석 수행"""
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

        # 데이터 유형별 분석
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

        # 패턴 통계 계산
        analysis_result["pattern_statistics"] = self.calculate_pattern_statistics(analysis_result["patterns_detected"])

        # 신뢰도 점수 계산
        analysis_result["confidence_scores"] = self.calculate_confidence_scores(analysis_result["patterns_detected"])

        # 추천 생성
        analysis_result["recommendations"] = self.generate_recommendations(analysis_result["patterns_detected"])

        # 리스크 평가
        analysis_result["risk_assessment"] = self.assess_pattern_risks(analysis_result["patterns_detected"])

        # 유사 패턴 검색
        analysis_result["similar_patterns"] = self.find_similar_patterns(analysis_result["patterns_detected"])

        # 패턴 기록 업데이트
        self.update_pattern_history(analysis_result)

        return analysis_result

    def detect_code_patterns(self, code: str) -> List[Dict[str, Any]]:
        """코드 패턴 감지"""
        patterns = []

        # 함수 길이 패턴
        function_pattern = re.compile(r'def\s+\w+\([^)]*\)\s*:.*?(?=def|\Z)', re.DOTALL)
        for match in function_pattern.finditer(code):
            function_content = match.group()
            lines = function_content.split('\n')
            line_count = len([line for line in lines if line.strip()])

            if line_count > self.analysis_config["min_occurrences"]:
                patterns.append({
                    "pattern_type": "code_function_length",
                    "description": f"긴 함수 패턴 ({line_count} 라인)",
                    "severity": "warning" if line_count < 50 else "error",
                    "details": {
                        "line_count": line_count,
                        "function_name": self.extract_function_name(match.group())
                    },
                    "known_solution": "함수 분리 또는 리팩토링 권장"
                })

        # 중첩 레벨 패턴
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
                    "description": f"높은 중첩 레벨 ({current_nesting} 레벨)",
                    "severity": "warning",
                    "details": {
                        "nesting_level": current_nesting,
                        "line_number": code[:match.start()].count('\n') + 1
                    },
                    "known_solution": "중첩 구조 간소화"
                })

        # 매개변수 수 패턴
        param_pattern = re.compile(r'def\s+(\w+)\(([^)]+)\)')
        for match in param_pattern.finditer(code):
            params = match.group(2).split(',')
            param_count = len([p.strip() for p in params if p.strip()])

            if param_count > 7:
                patterns.append({
                    "pattern_type": "code_parameter_count",
                    "description": f"많은 매개변수 패턴 ({param_count}개)",
                    "severity": "warning",
                    "details": {
                        "param_count": param_count,
                        "function_name": match.group(1)
                    },
                    "known_solution": "매개변수 그룹화 또는 객체 사용"
                })

        return patterns

    def detect_error_patterns(self, error_data: str) -> List[Dict[str, Any]]:
        """에러 패턴 감지"""
        patterns = []

        # 널 포인터 패턴
        null_pattern = re.compile(r'None|null|undefined|nullpointer|nullreference', re.IGNORECASE)
        if null_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_null_pointer",
                "description": "널 포인터 관련 에러 패턴",
                "severity": "high",
                "details": {
                    "indicators": null_pattern.findall(error_data)
                },
                "known_solution": "사전 유효성 검사 및 null 처리 로직 추가"
            })

        # 타임아웃 패턴
        timeout_pattern = re.compile(r'timeout|timeouterror|timedout', re.IGNORECASE)
        if timeout_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_timeout",
                "description": "타임아웃 관련 에러 패턴",
                "severity": "medium",
                "details": {
                    "indicators": timeout_pattern.findall(error_data)
                },
                "known_solution": "타임아웃 설정 최적화 및 재시도 로직"
            })

        # 메모리 관련 패턴
        memory_pattern = re.compile(r'memory|outofmemory|heap|stackoverflow', re.IGNORECASE)
        if memory_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_memory",
                "description": "메모리 관련 에러 패턴",
                "severity": "high",
                "details": {
                    "indicators": memory_pattern.findall(error_data)
                },
                "known_solution": "메모리 관리 개선 및 가비지 컬렉션 최적화"
            })

        # 네트워크 관련 패턴
        network_pattern = re.compile(r'connection|network|timeout|unreachable', re.IGNORECASE)
        if network_pattern.search(error_data):
            patterns.append({
                "pattern_type": "error_network",
                "description": "네트워크 관련 에러 패턴",
                "severity": "medium",
                "details": {
                    "indicators": network_pattern.findall(error_data)
                },
                "known_solution": "네트워크 예외 처리 및 재시도 메커니즘"
            })

        return patterns

    def detect_performance_patterns(self, performance_data: str) -> List[Dict[str, Any]]:
        """성능 패턴 감지"""
        patterns = []

        # O(n^2) 복잡도 패턴
        quadratic_pattern = re.compile(r'nested.*loop|nested.*for|nested.*while', re.IGNORECASE)
        if quadratic_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_quadratic",
                "description": "O(n^2) 복잡도 패턴",
                "severity": "warning",
                "details": {
                    "indicators": quadratic_pattern.findall(performance_data)
                },
                "known_solution": "알고리즘 개선 또는 캐싱 적용"
            })

        # 메모리 누수 패턴
        leak_pattern = re.compile(r'memory.*leak|leak.*memory|growth.*memory|memory.*growth', re.IGNORECASE)
        if leak_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_memory_leak",
                "description": "메모리 누수 패턴",
                "severity": "high",
                "details": {
                    "indicators": leak_pattern.findall(performance_data)
                },
                "known_solution": "메모리 관리 점검 및 객체 생명주기 관리"
            })

        # CPU 집약적 패턴
        cpu_pattern = re.compile(r'cpu.*intensive|compute.*heavy|processor.*load', re.IGNORECASE)
        if cpu_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_cpu_intensive",
                "description": "CPU 집약적 패턴",
                "severity": "medium",
                "details": {
                    "indicators": cpu_pattern.findall(performance_data)
                },
                "known_solution": "비동기 처리 또는 분산 컴퓨팅 적용"
            })

        # I/O 병목 패턴
        io_pattern = re.compile(r'io.*bottleneck|file.*io|disk.*io|network.*io', re.IGNORECASE)
        if io_pattern.search(performance_data):
            patterns.append({
                "pattern_type": "performance_io_bottleneck",
                "description": "I/O 병목 패턴",
                "severity": "medium",
                "details": {
                    "indicators": io_pattern.findall(performance_data)
                },
                "known_solution": "I/O 캐싱 또는 비동기 I/O 적용"
            })

        return patterns

    def detect_execution_patterns(self, execution_data: str) -> List[Dict[str, Any]]:
        """실행 패턴 감지"""
        patterns = []

        # 반복 실행 패턴
        repeated_pattern = re.compile(r'execute.*repeat|repeat.*execute|loop.*execute', re.IGNORECASE)
        if repeated_pattern.search(execution_data):
            patterns.append({
                "pattern_type": "execution_repeated",
                "description": "반복 실행 패턴",
                "severity": "info",
                "details": {
                    "indicators": repeated_pattern.findall(execution_data)
                },
                "known_solution": "배치 처리 또는 캐싱 적용"
            })

        # 순차 실행 패턴
        sequential_pattern = re.compile(r'sequential.*execute|execute.*sequential|serial.*execute', re.IGNORECASE)
        if sequential_pattern.search(execution_data):
            patterns.append({
                "pattern_type": "execution_sequential",
                "description": "순차 실행 패턴",
                "severity": "warning",
                "details": {
                    "indicators": sequential_pattern.findall(execution_data)
                },
                "known_solution": "병렬 처리 또는 파이프라인 적용"
            })

        # 지연 실행 패턴
        delayed_pattern = re.compile(r'delay.*execute|execute.*delay|late.*execute', re.IGNORECASE)
        if delayed_pattern.search(execution_data):
            patterns.append({
                "pattern_type": "execution_delayed",
                "description": "지연 실행 패턴",
                "severity": "medium",
                "details": {
                    "indicators": delayed_pattern.findall(execution_data)
                },
                "known_solution": "실행 계획 최적화 또는 예측 실행"
            })

        return patterns

    def detect_general_patterns(self, data: str) -> List[Dict[str, Any]]:
        """일반 패턴 감지"""
        patterns = []

        # 반복 키워드 패턴
        common_keywords = ["error", "warning", "exception", "fail", "success", "timeout"]
        keyword_counts = defaultdict(int)

        for keyword in common_keywords:
            count = len(re.findall(rf'\b{keyword}\b', data, re.IGNORECASE))
            if count > self.analysis_config["min_occurrences"]:
                keyword_counts[keyword] = count

        if keyword_counts:
            patterns.append({
                "pattern_type": "keyword_frequency",
                "description": "키워드 빈도 패턴",
                "severity": "info",
                "details": {
                    "keyword_counts": dict(keyword_counts),
                    "total_occurrences": sum(keyword_counts.values())
                },
                "known_solution": "키워드 분석을 통한 개선 영역 식별"
            })

        # 길이 패턴
        data_lines = data.split('\n')
        avg_line_length = sum(len(line) for line in data_lines) / len(data_lines) if data_lines else 0

        if avg_line_length > 100:
            patterns.append({
                "pattern_type": "line_length",
                "description": "긴 라인 패턴",
                "severity": "warning",
                "details": {
                    "avg_line_length": avg_line_length,
                    "max_line_length": max(len(line) for line in data_lines)
                },
                "known_solution": "라인 길이 조정 및 코드 포맷팅"
            })

        return patterns

    def calculate_pattern_statistics(self, patterns: List[Dict[str, Any]]) -> Dict[str, Any]:
        """패턴 통계 계산"""
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
        """신뢰도 점수 계산"""
        confidence_scores = {}

        for i, pattern in enumerate(patterns):
            base_confidence = self.analysis_config["confidence_threshold"]

            # 심각도에 따른 조정
            severity_multiplier = {
                "error": 1.2,
                "high": 1.1,
                "warning": 1.0,
                "medium": 0.9,
                "info": 0.8
            }

            multiplier = severity_multiplier.get(pattern.get("severity", "info"), 1.0)
            confidence = min(1.0, base_confidence * multiplier)

            # 패턴 유형에 따른 추가 조정
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
        """개선 추천 생성"""
        recommendations = []

        # 심각도 기반 추천
        high_severity_patterns = [p for p in patterns if p.get("severity") in ["error", "high"]]
        if high_severity_patterns:
            recommendations.append(f"우선 해결할 심각한 패턴 {len(high_severity_patterns)}개 존재")

        # 패턴 유형 기반 추천
        code_patterns = [p for p in patterns if p.get("pattern_type", "").startswith("code_")]
        if code_patterns:
            recommendations.append("코드 스타일 및 구조 개선 필요")

        error_patterns = [p for p in patterns if p.get("pattern_type", "").startswith("error_")]
        if error_patterns:
            recommendations.append("오류 처리 메커니즘 개선 필요")

        performance_patterns = [p for p in patterns if p.get("pattern_type", "").startswith("performance_")]
        if performance_patterns:
            recommendations.append("성능 최적화 필요")

        # 일반 추천
        if not patterns:
            recommendations.append("패턴 감지되지 않음 - 현재 상태 양호")
        else:
            recommendations.append("정기적인 패턴 점검 및 개선 권장")

        return recommendations

    def assess_pattern_risks(self, patterns: List[Dict[str, Any]]) -> Dict[str, Any]:
        """패턴 리스크 평가"""
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
        """유사 패턴 검색"""
        similar_patterns = []

        for pattern in patterns:
            pattern_type = pattern.get("pattern_type")

            # 기존 데이터베이스에서 유사 패턴 검색
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
        """패턴 기록 업데이트"""
        self.pattern_history.append({
            "timestamp": analysis_result["timestamp"],
            "data_type": analysis_result["data_type"],
            "pattern_count": len(analysis_result["patterns_detected"]),
            "risk_level": analysis_result["risk_assessment"]["risk_level"]
        })

        # 최대 100개 항목 유지
        if len(self.pattern_history) > 100:
            self.pattern_history = self.pattern_history[-100:]

    def extract_function_name(self, function_match: str) -> str:
        """함수 이름 추출"""
        match = re.search(r'def\s+(\w+)', function_match)
        return match.group(1) if match else "unknown"


def analyze_patterns_with_engine(data: str, data_type: str = "general") -> Dict[str, Any]:
    """패턴 인식 엔진으로 데이터 분석"""
    engine = PatternRecognitionEngine()
    return engine.analyze_patterns(data, data_type)


def get_pattern_database() -> Dict[str, Any]:
    """패턴 데이터베이스 반환"""
    engine = PatternRecognitionEngine()
    return engine.pattern_database


def get_pattern_history() -> List[Dict[str, Any]]:
    """패턴 기록 반환"""
    engine = PatternRecognitionEngine()
    return engine.pattern_history


# 표준 Skill 인터페이스 구현
def main() -> None:
    """Skill 메인 함수"""
    try:
        # 인자 파싱
        if len(sys.argv) < 2:
            print(json.dumps({
                "error": "Usage: python3 pattern_recognition_engine.py <data> [data_type:general|code|error|performance|execution]"
            }))
            sys.exit(1)

        data = sys.argv[1]
        data_type = sys.argv[2] if len(sys.argv) > 2 else "general"

        # 데이터 타입 검증
        if data_type not in ["general", "code", "error", "performance", "execution"]:
            print(json.dumps({
                "error": f"Invalid data_type: {data_type}. Must be: general, code, error, performance, execution"
            }))
            sys.exit(1)

        # 패턴 인식 분석 실행
        result = analyze_patterns_with_engine(data, data_type)

        # 결과 출력
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