#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-005 | @SPEC:SKILL-CROSS-DOMAIN-ANALYSIS-ENGINE-001 | @TEST: tests/skills/test_cross_domain_analysis_engine.py
"""Cross-Domain Analysis Engine Skill

도메인 간 분석 엔진. 다른 분야의 지식과 경험을 연결하여 새로운 통찰을 도출:
1. 도메인 매핑
2. 유사성 분석
3. 지식 이전
4. 크로스 도메인 인사이트
5. 창의적 문제 해결

사용법:
    Skill("cross_domain_analysis_engine")
"""

import json
import re
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Set
from collections import defaultdict, Counter
from dataclasses import dataclass


@dataclass
class DomainKnowledge:
    """도메인 지식 데이터 클래스"""
    domain: str
    concepts: Set[str]
    principles: List[str]
    patterns: List[str]
    challenges: List[str]
    solutions: List[str]
    metadata: Dict[str, Any]


@dataclass
class DomainSimilarity:
    """도메인 유사성 데이터 클래스"""
    domain1: str
    domain2: str
    similarity_score: float
    matching_concepts: Set[str]
    transferable_solutions: List[str]
    adaptation_requirements: List[str]


class CrossDomainAnalysisEngine:
    """크로스 도메인 분석 엔진 클래스"""

    def __init__(self):
        self.domain_knowledge = self.load_domain_knowledge()
        self.similarity_matrix = self.load_similarity_matrix()
        self.transfer_patterns = self.load_transfer_patterns()
        self.case_studies = self.load_case_studies()

    def load_domain_knowledge(self) -> Dict[str, DomainKnowledge]:
        """도메인 지식 로드"""
        try:
            knowledge_file = Path(".moai/research/domains/domain_knowledge.json")
            if knowledge_file.exists():
                with open(knowledge_file, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    domains = {}
                    for domain_name, domain_data in data.items():
                        domains[domain_name] = DomainKnowledge(
                            domain=domain_name,
                            concepts=set(domain_data.get("concepts", [])),
                            principles=domain_data.get("principles", []),
                            patterns=domain_data.get("patterns", []),
                            challenges=domain_data.get("challenges", []),
                            solutions=domain_data.get("solutions", []),
                            metadata=domain_data.get("metadata", {})
                        )
                    return domains
        except Exception:
            pass

        # 기본 도메인 지식
        return self.get_default_domain_knowledge()

    def get_default_domain_knowledge(self) -> Dict[str, DomainKnowledge]:
        """기본 도메인 지식"""
        domains = {}

        # 소프트웨어 엔지니어링
        domains["software_engineering"] = DomainKnowledge(
            domain="software_engineering",
            concepts={"agile", "scrum", "ci/cd", "testing", "refactoring", "design_patterns"},
            principles=["KISS", "YAGNI", "DRY", "SOLID"],
            patterns=["MVC", "MVVM", "Observer", "Strategy"],
            challenges=["scalability", "maintainability", "performance"],
            solutions=["microservices", "containerization", "automated_testing"],
            metadata={"complexity": "high", "innovation_rate": "high"}
        )

        # 기계 학습
        domains["machine_learning"] = DomainKnowledge(
            domain="machine_learning",
            concepts={"neural_networks", "deep_learning", "supervised_learning", "unsupervised_learning"},
            principles ["bias_variance_tradeoff", "cross_validation", "feature_engineering"],
            patterns ["CNN", "RNN", "Transformer", "GAN"],
            challenges ["data_quality", "model_interpretability", "overfitting"],
            solutions ["transfer_learning", "ensemble_methods", "regularization"],
            metadata = {"complexity": "very_high", "innovation_rate": "very_high"}
        )

        # 시스템 아키텍처
        domains["system_architecture"] = DomainKnowledge(
            domain="system_architecture",
            concepts={"microservices", "monolith", "distributed_systems", "load_balancing"},
            principles ["scalability", "reliability", "security", "maintainability"],
            patterns ["event_driven", "serverless", "circuit_breaker", "retry_pattern"],
            challenges ["consistency", "availability", "partition_tolerance"],
            solutions ["event_sourcing", "cqrs", "distributed_cache", "service_mesh"],
            metadata = {"complexity": "high", "innovation_rate": "medium"}
        )

        # 사용자 경험 디자인
        domains["ux_design"] = DomainKnowledge(
            domain="ux_design",
            concepts={"user_research", "wireframing", "prototyping", "usability_testing"},
            principles ["user_centered", "accessibility", "consistency", "feedback"],
            patterns ["navigation", "information_architecture", "interaction_design"],
            challenges ["user_adoption", "accessibility", "mobile_responsive"],
            solutions ["user_testing", "iterative_design", "design_systems"],
            metadata = {"complexity": "medium", "innovation_rate": "medium"}
        )

        return domains

    def load_similarity_matrix(self) -> Dict[str, Dict[str, float]]:
        """도메인 간 유사성 매트릭스 로드"""
        try:
            matrix_file = Path(".moai/research/domains/similarity_matrix.json")
            if matrix_file.exists():
                with open(matrix_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return self.calculate_default_similarity_matrix()

    def calculate_default_similarity_matrix(self) -> Dict[str, Dict[str, float]]:
        """기본 유사성 매트릭스 계산"""
        domains = list(self.domain_knowledge.keys())
        matrix = {}

        for domain1 in domains:
            matrix[domain1] = {}
            for domain2 in domains:
                if domain1 == domain2:
                    matrix[domain1][domain2] = 1.0
                else:
                    # 개념 기반 유사도 계산
                    concepts1 = self.domain_knowledge[domain1].concepts
                    concepts2 = self.domain_knowledge[domain2].concepts

                    if concepts1 and concepts2:
                        intersection = concepts1.intersection(concepts2)
                        union = concepts1.union(concepts2)
                        similarity = len(intersection) / len(union)
                    else:
                        similarity = 0.1

                    matrix[domain1][domain2] = similarity

        return matrix

    def load_transfer_patterns(self) -> Dict[str, Any]:
        """지식 이전 패턴 로드"""
        try:
            patterns_file = Path(".moai/research/domains/transfer_patterns.json")
            if patterns_file.exists():
                with open(patterns_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return {
            "common_transfer_patterns": [
                {"pattern": "analogical_reasoning", "description": "유추적 사고"},
                {"pattern": "principle_application", "description": "원리 적용"},
                {"pattern": "pattern_recognition", "description": "패턴 인식"},
                {"pattern": "abstraction_transfer", "description": "추상화 이전"}
            ],
            "transfer_success_factors": [
                "conceptual_similarity",
                "problem_structure_similarities",
                "solution_applicability",
                "adaptation_requirements"
            ]
        }

    def load_case_studies(self) -> List[Dict[str, Any]]:
        """성공 사례 로드"""
        try:
            case_studies_file = Path(".moai/research/domains/case_studies.json")
            if case_studies_file.exists():
                with open(case_studies_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except Exception:
            pass

        return [
            {
                "id": "case_001",
                "source_domain": "software_engineering",
                "target_domain": "machine_learning",
                "problem": "모델 배포 및 관리",
                "solution": "CI/CD 파이프라인 적용",
                "outcome": "모델 배포 자동화 성공",
                "lessons_learned": ["자동화 원칙 적용", "모니터링 시스템 이전"]
            },
            {
                "id": "case_002",
                "source_domain": "ux_design",
                "target_domain": "software_engineering",
                "problem": "복잡한 사용자 인터페이스",
                "solution": "피드백 루프 패턴 적용",
                "outcome": "사용자 만족도 향상",
                "lessons_learned": ["사용자 중심 개발", "반복적 개선 프로세스"]
            }
        ]

    def analyze_cross_domain_connections(self, problem_domain: str,
                                       target_domains: List[str] = None) -> Dict[str, Any]:
        """크로스 도메인 연결 분석"""
        if target_domains is None:
            target_domains = [d for d in self.domain_knowledge.keys() if d != problem_domain]

        analysis_result = {
            "analysis_type": "cross_domain_connections",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "problem_domain": problem_domain,
            "target_domains_analyzed": len(target_domains),
            "domain_similarities": [],
            "transfer_opportunities": [],
            "analogical_connections": [],
            "recommendations": [],
            "case_studies_relevant": []
        }

        # 도메인 유사성 분석
        for target_domain in target_domains:
            similarity = DomainSimilarity(
                domain1=problem_domain,
                domain2=target_domain,
                similarity_score=self.similarity_matrix.get(problem_domain, {}).get(target_domain, 0.0),
                matching_concepts=self.find_matching_concepts(problem_domain, target_domain),
                transferable_solutions=self.find_transferable_solutions(problem_domain, target_domain),
                adaptation_requirements=self.find_adaptation_requirements(problem_domain, target_domain)
            )

            analysis_result["domain_similarities"].append({
                "target_domain": target_domain,
                "similarity_score": similarity.similarity_score,
                "matching_concepts_count": len(similarity.matching_concepts),
                "transferable_solutions": similarity.transferable_solutions,
                "adaptation_requirements": similarity.adaptation_requirements
            })

            # 지식 이전 기회 식별
            if similarity.similarity_score > 0.3:
                analysis_result["transfer_opportunities"].append({
                    "from_domain": problem_domain,
                    "to_domain": target_domain,
                    "score": similarity.similarity_score,
                    "solutions": similarity.transferable_solutions
                })

                # 유추적 연결 생성
                analysis_result["analogical_connections"].append({
                    "source": problem_domain,
                    "target": target_domain,
                    "relationship_type": "conceptual_similarity",
                    "strength": similarity.similarity_score,
                    "analogies": self.generate_analogies(similarity)
                })

        # 관련 사례 검색
        analysis_result["case_studies_relevant"] = self.find_relevant_case_studies(problem_domain, target_domains)

        # 추천 생성
        analysis_result["recommendations"] = self.generate_cross_domain_recommendations(analysis_result)

        return analysis_result

    def find_matching_concepts(self, domain1: str, domain2: str) -> Set[str]:
        """일치하는 개념 찾기"""
        concepts1 = self.domain_knowledge[domain1].concepts
        concepts2 = self.domain_knowledge[domain2].concepts
        return concepts1.intersection(concepts2)

    def find_transferable_solutions(self, domain1: str, domain2: str) -> List[str]:
        """이전 가능한 해결책 찾기"""
        solutions1 = self.domain_knowledge[domain1].solutions
        solutions2 = self.domain_knowledge[domain2].solutions
        challenges1 = self.domain_knowledge[domain1].challenges
        challenges2 = self.domain_knowledge[domain2].challenges

        transferable = []

        # 공통 도전 과제에 대한 해결책 이전 가능성
        common_challenges = set(challenges1).intersection(set(challenges2))
        for challenge in common_challenges:
            # domain2의 해결책이 domain1의 도전 과제에 적용 가능한지 확인
            for solution in solutions2:
                if any(keyword in solution.lower() for keyword in challenge.lower().split()):
                    transferable.append(f"{solution} for {challenge}")

        return transferable

    def find_adaptation_requirements(self, domain1: str, domain2: str) -> List[str]:
        """적응 요구 사항 찾기"""
        requirements = []

        # 도메인별 복잡성 차이
        complexity1 = self.domain_knowledge[domain1].metadata.get("complexity", "medium")
        complexity2 = self.domain_knowledge[domain2].metadata.get("complexity", "medium")

        if complexity1 != complexity2:
            requirements.append(f"복잡도 조정 필요: {complexity1} -> {complexity2}")

        # 혁신 속도 차이
        innovation1 = self.domain_knowledge[domain1].metadata.get("innovation_rate", "medium")
        innovation2 = self.domain_knowledge[domain2].metadata.get("innovation_rate", "medium")

        if innovation1 != innovation2:
            requirements.append(f"혁신 속도 조화 필요: {innovation1} -> {innovation2}")

        return requirements

    def generate_analogies(self, similarity: DomainSimilarity) -> List[str]:
        """유추적 연결 생성"""
        analogies = []

        matching_concepts = similarity.matching_concepts
        if len(matching_concepts) > 0:
            concept_list = list(matching_concepts)
            analogy = f"{similarity.domain1}의 {concept_list[0]}는 {similarity.domain2}의 {concept_list[-1]}와 유사한 문제 해결 접근 방식을 사용할 수 있습니다"
            analogies.append(analogy)

        return analogies

    def find_relevant_case_studies(self, problem_domain: str, target_domains: List[str]) -> List[Dict[str, Any]]:
        """관련 사례 찾기"""
        relevant = []

        for case_study in self.case_studies:
            if (case_study["source_domain"] == problem_domain and case_study["target_domain"] in target_domains) or \
               (case_study["target_domain"] == problem_domain and case_study["source_domain"] in target_domains):
                relevant.append(case_study)

        return relevant

    def generate_cross_domain_recommendations(self, analysis_result: Dict[str, Any]) -> List[str]:
        """크로스 도메인 추천 생성"""
        recommendations = []

        # 유사성 기반 추천
        high_similarity_domains = [
            d for d in analysis_result["domain_similarities"]
            if d["similarity_score"] > 0.5
        ]

        if high_similarity_domains:
            recommendations.append(f"높은 유사성 도메인 {len(high_similarity_domains)}개에서 지식 이전 가능")

        # 이전 기회 기반 추천
        if analysis_result["transfer_opportunities"]:
            recommendations.append(f"{len(analysis_result['transfer_opportunities'])}개의 지식 이전 기회 발견")

        # 사례 기반 추천
        if analysis_result["case_studies_relevant"]:
            recommendations.append(f"{len(analysis_result['case_studies_relevant'])}개의 성공 사례 참고 가능")

        # 일반 추천
        if not analysis_result["transfer_opportunities"]:
            recommendations.append("도메인 간 유사도가 낮음 - 새로운 접근 방식 고려 필요")

        return recommendations

    def transfer_knowledge(self, source_domain: str, target_domain: str,
                         knowledge_type: str = "general") -> Dict[str, Any]:
        """지식 이전 수행"""
        transfer_result = {
            "transfer_type": "knowledge_transfer",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "source_domain": source_domain,
            "target_domain": target_domain,
            "knowledge_type": knowledge_type,
            "transfer_success": False,
            "transferred_elements": [],
            "adaptation_requirements": [],
            "risks": [],
            "recommendations": []
        }

        # 지식 유형별 이전
        if knowledge_type == "principles":
            transferred = self.transfer_principles(source_domain, target_domain)
        elif knowledge_type == "patterns":
            transferred = self.transfer_patterns(source_domain, target_domain)
        elif knowledge_type == "solutions":
            transferred = self.transfer_solutions(source_domain, target_domain)
        else:
            transferred = self.transfer_general_knowledge(source_domain, target_domain)

        transfer_result["transferred_elements"] = transferred["elements"]
        transfer_result["adaptation_requirements"] = transferred["adaptations"]
        transfer_result["risks"] = self.assess_transfer_risks(source_domain, target_domain)
        transfer_result["recommendations"] = self.generate_transfer_recommendations(transfer_result)

        # 이전 성공 여부 판정
        if len(transferred["elements"]) > 0:
            transfer_result["transfer_success"] = True

        return transfer_result

    def transfer_principles(self, source_domain: str, target_domain: str) -> Dict[str, Any]:
        """원리 이전"""
        source_principles = self.domain_knowledge[source_domain].principles
        target_concepts = self.domain_knowledge[target_domain].concepts

        transferred = []
        adaptations = []

        for principle in source_principles:
            # 원리가 대상 도메인의 개념과 관련 있는지 확인
            if any(keyword in principle.lower() for keyword in " ".join(target_concepts).lower().split()):
                transferred.append({
                    "type": "principle",
                    "content": principle,
                    "applicability": "high"
                })
                adaptations.append(f"원리 '{principle}' 적용 시 개념적 조정 필요")

        return {"elements": transferred, "adaptations": adaptations}

    def transfer_patterns(self, source_domain: str, target_domain: str) -> Dict[str, Any]:
        """패턴 이전"""
        source_patterns = self.domain_knowledge[source_domain].patterns
        target_patterns = self.domain_knowledge[target_domain].patterns

        transferred = []
        adaptations = []

        for pattern in source_patterns:
            # 이미 대상 도메인에 존재하는지 확인
            if pattern not in target_patterns:
                transferred.append({
                    "type": "pattern",
                    "content": pattern,
                    "applicability": "medium"
                })
                adaptations.append(f"패턴 '{pattern}' 적용 시 환경별 조정 필요")

        return {"elements": transferred, "adaptations": adaptations}

    def transfer_solutions(self, source_domain: str, target_domain: str) -> Dict[str, Any]:
        """해결책 이전"""
        source_solutions = self.domain_knowledge[source_domain].solutions
        target_challenges = self.domain_knowledge[target_domain].challenges

        transferred = []
        adaptations = []

        for solution in source_solutions:
            # 해결책이 대상 도메인의 도전 과제와 관련 있는지 확인
            for challenge in target_challenges:
                if any(keyword in solution.lower() for keyword in challenge.lower().split()):
                    transferred.append({
                        "type": "solution",
                        "content": solution,
                        "target_challenge": challenge,
                        "applicability": "high"
                    })
                    adaptations.append(f"해결책 '{solution}' 적용 시 도전 과제 '{challenge}' 특성에 맞게 수정 필요")

        return {"elements": transferred, "adaptations": adaptations}

    def transfer_general_knowledge(self, source_domain: str, target_domain: str) -> Dict[str, Any]:
        """일반 지식 이전"""
        source_knowledge = self.domain_knowledge[source_domain]
        target_knowledge = self.domain_knowledge[target_domain]

        transferred = []
        adaptations = []

        # 공통 개� 기반 지식 이전
        common_concepts = source_knowledge.concepts.intersection(target_knowledge.concepts)
        for concept in common_concepts:
            transferred.append({
                "type": "concept",
                "content": concept,
                "applicability": "high"
            })

        # 원리 및 패턴 이전
        transferred.extend([{"type": "principle", "content": p, "applicability": "medium"}
                          for p in source_knowledge.principles])
        transferred.extend([{"type": "pattern", "content": p, "applicability": "medium"}
                          for p in source_knowledge.patterns])

        adaptations.append("도메인별 특성에 맞는 일반화 필요")
        adaptations.append("실제 적용 시 맥락 고려 필요")

        return {"elements": transferred, "adaptations": adaptations}

    def assess_transfer_risks(self, source_domain: str, target_domain: str) -> List[Dict[str, Any]]:
        """이전 리스크 평가"""
        risks = []

        # 도메인 간 차이
        source_metadata = self.domain_knowledge[source_domain].metadata
        target_metadata = self.domain_knowledge[target_domain].metadata

        # 복잡성 차이 리스크
        if source_metadata.get("complexity") != target_metadata.get("complexity"):
            risks.append({
                "type": "complexity_mismatch",
                "description": "도메인 간 복잡성 불일치",
                "severity": "medium",
                "mitigation": "단계적 적용 및 테스트"
            })

        # 혁신 속도 차이 리스크
        if source_metadata.get("innovation_rate") != target_metadata.get("innovation_rate"):
            risks.append({
                "type": "innovation_mismatch",
                "description": "도메인 간 혁신 속도 불일치",
                "severity": "high",
                "mitigation": "혁신 속도 조화 전략 수립"
            })

        # 개념적 차이 리스크
        concept_overlap = len(self.find_matching_concepts(source_domain, target_domain))
        if concept_overlap < 3:
            risks.append({
                "type": "conceptual_gap",
                "description": "개념적 격차가 큼",
                "severity": "high",
                "mitigation": "추가적인 개념 교육 및 연구"
            })

        return risks

    def generate_transfer_recommendations(self, transfer_result: Dict[str, Any]) -> List[str]:
        """이전 추천 생성"""
        recommendations = []

        # 성공적인 이전
        if transfer_result["transfer_success"]:
            recommendations.append("지식 이전 성공 - 단계적 적용 권장")
        else:
            recommendations.append("지식 이전 제한적 - 다른 접근 방식 고려 필요")

        # 적응 요구 사항
        if transfer_result["adaptation_requirements"]:
            recommendations.append(f"{len(transfer_result['adaptation_requirements'])}개의 적응 요구 사항 존재")

        # 리스크 기반 추천
        high_risk_risks = [r for r in transfer_result["risks"] if r["severity"] == "high"]
        if high_risk_risks:
            recommendations.append(f"{len(high_risk_risks)}개의 높은 리스크 감지 - 세심한 계획 필요")

        return recommendations

    def discover_analogies(self, problem_description: str) -> Dict[str, Any]:
        """유사성 기반 유추적 연결 발견"""
        analogy_result = {
            "discovery_type": "analogical_discovery",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "problem_description": problem_description,
            "potential_analogies": [],
            "conceptual_matches": [],
            "solution_transfer_opportunities": [],
            "recommendations": []
        }

        # 문제 설명에서 키워드 추출
        problem_keywords = self.extract_keywords(problem_description)

        # 도메인별 유사성 검사
        for domain_name, domain_knowledge in self.domain_knowledge.items():
            # 도메인 키워드 추출
            domain_keywords = list(domain_knowledge.concepts) + domain_knowledge.principles + domain_knowledge.patterns

            # 키워드 유사도 계산
            similarity_score = self.calculate_keyword_similarity(problem_keywords, domain_keywords)

            if similarity_score > 0.3:
                analogy_result["conceptual_matches"].append({
                    "domain": domain_name,
                    "similarity_score": similarity_score,
                    "matching_keywords": self.find_matching_keywords(problem_keywords, domain_keywords)
                })

                # 해결책 이전 기회
                for solution in domain_knowledge.solutions:
                    if any(keyword in solution.lower() for keyword in problem_keywords):
                        analogy_result["solution_transfer_opportunities"].append({
                            "domain": domain_name,
                            "solution": solution,
                            "relevance_score": similarity_score
                        })

        # 유사성을 기반으로 잠재적 유추적 연결 생성
        if analogy_result["conceptual_matches"]:
            top_matches = sorted(analogy_result["conceptual_matches"],
                               key=lambda x: x["similarity_score"], reverse=True)[:3]

            for match in top_matches:
                analogy = self.generate_domain_analogy(problem_description, match["domain"], match["similarity_score"])
                if analogy:
                    analogy_result["potential_analogies"].append(analogy)

        # 추천 생성
        analogy_result["recommendations"] = self.generate_analogy_recommendations(analogy_result)

        return analogy_result

    def extract_keywords(self, text: str) -> List[str]:
        """문자열에서 키워드 추출"""
        # 간단한 키워드 추출
        words = re.findall(r'\b\w+\b', text.lower())
        # 불용어 제거
        stop_words = {"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
        keywords = [word for word in words if word not in stop_words and len(word) > 3]
        return keywords

    def calculate_keyword_similarity(self, keywords1: List[str], keywords2: List[str]) -> float:
        """키워드 유사도 계산"""
        if not keywords1 or not keywords2:
            return 0.0

        set1 = set(keywords1)
        set2 = set(keywords2)

        intersection = len(set1.intersection(set2))
        union = len(set1.union(set2))

        return intersection / union if union > 0 else 0.0

    def find_matching_keywords(self, keywords1: List[str], keywords2: List[str]) -> List[str]:
        """일치하는 키워드 찾기"""
        set1 = set(keywords1)
        set2 = set(keywords2)
        return list(set1.intersection(set2))

    def generate_domain_analogy(self, problem: str, domain: str, similarity_score: float) -> Optional[Dict[str, Any]]:
        """도메인 유사성 생성"""
        domain_knowledge = self.domain_knowledge[domain]

        # 유사성에 기반한 유추적 연결 생성
        analogy = {
            "domain": domain,
            "similarity_score": similarity_score,
            "analogical_connection": f"현재 문제는 {domain} 도메인의 개념과 유사한 접근 방식을 사용할 수 있습니다",
            "related_concepts": list(domain_knowledge.concepts)[:5],  # 상위 5개 개념
            "potential_solutions": domain_knowledge.solutions[:3]  # 상위 3개 해결책
        }

        return analogy

    def generate_analogy_recommendations(self, analogy_result: Dict[str, Any]) -> List[str]:
        """유추적 연결 추천 생성"""
        recommendations = []

        if analogy_result["potential_analogies"]:
            recommendations.append(f"{len(analogy_result['potential_analogies'])}개의 잠재적 유추적 연결 발견")

        if analogy_result["solution_transfer_opportunities"]:
            recommendations.append(f"{len(analogy_result['solution_transfer_opportunities'])}개의 해결책 이전 기회 발견")

        if not analogy_result["conceptual_matches"]:
            recommendations.append("유사한 도메인 발견되지 않음 - 새로운 접근 방식 필요")

        return recommendations


def analyze_cross_domain_connections(problem_domain: str, target_domains: List[str] = None) -> Dict[str, Any]:
    """크로스 도메인 분석 엔진으로 도메인 연결 분석"""
    engine = CrossDomainAnalysisEngine()
    return engine.analyze_cross_domain_connections(problem_domain, target_domains)


def transfer_knowledge_between_domains(source_domain: str, target_domain: str,
                                    knowledge_type: str = "general") -> Dict[str, Any]:
    """크로스 도메인 분석 엔진으로 지식 이전"""
    engine = CrossDomainAnalysisEngine()
    return engine.transfer_knowledge(source_domain, target_domain, knowledge_type)


def discover_analogies_for_problem(problem_description: str) -> Dict[str, Any]:
    """크로스 도메인 분석 엔진으로 유사성 발견"""
    engine = CrossDomainAnalysisEngine()
    return engine.discover_analogies(problem_description)


def get_cross_domain_engine_status() -> Dict[str, Any]:
    """크로스 도메인 분석 엔진 상태 반환"""
    engine = CrossDomainAnalysisEngine()
    return {
        "domains_loaded": len(engine.domain_knowledge),
        "similarity_matrix_size": len(engine.similarity_matrix),
        "transfer_patterns_available": len(engine.transfer_patterns),
        "case_studies_count": len(engine.case_studies)
    }


# 표준 Skill 인터페이스 구현
def main() -> None:
    """Skill 메인 함수"""
    try:
        # 인자 파싱
        if len(sys.argv) < 2:
            print(json.dumps({
                "error": "Usage: python3 cross_domain_analysis_engine.py <action> [args...]"
            }))
            sys.exit(1)

        action = sys.argv[1]

        if action == "analyze":
            if len(sys.argv) < 3:
                print(json.dumps({
                    "error": "Usage: python3 cross_domain_analysis_engine.py analyze <problem_domain> [target_domains_json]"
                }))
                sys.exit(1)

            problem_domain = sys.argv[2]
            target_domains = json.loads(sys.argv[3]) if len(sys.argv) > 3 else None

            result = analyze_cross_domain_connections(problem_domain, target_domains)

        elif action == "transfer":
            if len(sys.argv) < 4:
                print(json.dumps({
                    "error": "Usage: python3 cross_domain_analysis_engine.py transfer <source_domain> <target_domain> [knowledge_type]"
                }))
                sys.exit(1)

            source_domain = sys.argv[2]
            target_domain = sys.argv[3]
            knowledge_type = sys.argv[4] if len(sys.argv) > 4 else "general"

            result = transfer_knowledge_between_domains(source_domain, target_domain, knowledge_type)

        elif action == "analogies":
            if len(sys.argv) < 3:
                print(json.dumps({
                    "error": "Usage: python3 cross_domain_analysis_engine.py analogies <problem_description>"
                }))
                sys.exit(1)

            problem_description = sys.argv[2]

            result = discover_analogies_for_problem(problem_description)

        elif action == "status":
            result = get_cross_domain_engine_status()

        else:
            print(json.dumps({
                "error": f"Unknown action: {action}. Available actions: analyze, transfer, analogies, status"
            }))
            sys.exit(1)

        # 결과 출력
        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        error_result = {
            "error": f"Cross-domain analysis engine failed: {str(e)}",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()