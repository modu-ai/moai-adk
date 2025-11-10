#!/usr/bin/env python3
# @CODE:SKILL-RESEARCH-003 | @SPEC:SKILL-KNOWLEDGE-INTEGRATION-HUB-001 | @TEST: tests/skills/test_knowledge_integration_hub.py
"""Knowledge Integration Hub Skill

지식 통합 허브. 다양한 지원의 지식을 통합하고 관리:
1. JIT 지식 로딩
2. 지식 베이스 관리
3. 지식 연결 분석
4. 자동 지식 갱신

사용법:
    Skill("knowledge_integration_hub")
"""

import json
import os
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Set, Tuple
from collections import defaultdict, Counter


class KnowledgeIntegrationHub:
    """지식 통합 허브 클래스"""

    def __init__(self):
        self.knowledge_bases = {}
        self.connection_graph = {}
        self.integration_rules = {}
        self.update_timestamps = {}
        self.load_configuration()

    def load_configuration(self) -> None:
        """설정 로드"""
        try:
            config_file = Path(".moai/config.json")
            if config_file.exists():
                with open(config_file, 'r', encoding='utf-8') as f:
                    config = json.load(f)
                    research_config = config.get("tags", {}).get("research_tags", {})

                    self.integration_rules = {
                        "auto_update": research_config.get("auto_discovery", True),
                        "cross_reference": research_config.get("cross_reference", True),
                        "knowledge_graph": research_config.get("knowledge_graph", True),
                        "pattern_matching": research_config.get("pattern_matching", True),
                        "max_knowledge_size": 1000,
                        "update_interval_hours": 24,
                        "confidence_threshold": 0.7
                    }
        except Exception:
            # 기본 설정
            self.integration_rules = {
                "auto_update": True,
                "cross_reference": True,
                "knowledge_graph": True,
                "pattern_matching": True,
                "max_knowledge_size": 1000,
                "update_interval_hours": 24,
                "confidence_threshold": 0.7
            }

    def load_all_knowledge_bases(self) -> Dict[str, Any]:
        """모든 지식 베이스 로드"""
        knowledge_bases = {}
        knowledge_dir = Path(".moai/research/knowledge/")

        if not knowledge_dir.exists():
            return knowledge_bases

        # 모든 지식 파일 로드
        for file_path in knowledge_dir.glob("*.json"):
            if file_path.is_file():
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        knowledge_bases[file_path.stem] = data
                        self.update_timestamps[file_path.stem] = file_path.stat().st_mtime
                except Exception:
                    continue

        return knowledge_bases

    def integrate_knowledge(self, new_knowledge: Dict[str, Any], source_type: str = "manual") -> Dict[str, Any]:
        """지식 통합 수행"""
        integration_result = {
            "integration_success": True,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "source_type": source_type,
            "knowledge_categories": [],
            "connections_made": 0,
            "conflicts_resolved": 0,
            "integrated_sources": [],
            "conflict_sources": [],
            "recommendations": []
        }

        # 모든 지식 베이스 로드
        self.knowledge_bases = self.load_all_knowledge_bases()

        # 지식 카테고리 식별
        categories = self.identify_knowledge_categories(new_knowledge)
        integration_result["knowledge_categories"] = categories

        # 지식 간 충돌 검사 및 해결
        conflicts = self.detect_knowledge_conflicts(new_knowledge)
        if conflicts:
            integration_result["conflicts_resolved"] = self.resolve_conflicts(new_knowledge, conflicts)
            integration_result["conflict_sources"] = [c["source"] for c in conflicts]

        # 지식 연결 생성
        connections = self.create_knowledge_connections(new_knowledge)
        integration_result["connections_made"] = len(connections)

        # 지식 저장
        saved_sources = self.save_integrated_knowledge(new_knowledge, categories)
        integration_result["integrated_sources"] = saved_sources

        # 지식 그래프 업데이트
        if self.integration_rules["knowledge_graph"]:
            self.update_knowledge_graph(connections)

        # 추천 생성
        integration_result["recommendations"] = self.generate_integration_recommendations(
            new_knowledge, connections, conflicts
        )

        return integration_result

    def identify_knowledge_categories(self, knowledge: Dict[str, Any]) -> List[str]:
        """지식 카테고리 식별"""
        categories = []

        # 연구 카테고리 매핑
        category_keywords = {
            "RESEARCH": ["research", "investigate", "analyze", "explore"],
            "ANALYSIS": ["analysis", "evaluate", "assess", "review"],
            "KNOWLEDGE": ["knowledge", "learn", "understand", "comprehend"],
            "INSIGHT": ["insight", "innovation", "optimization", "improvement"]
        }

        knowledge_text = json.dumps(knowledge, ensure_ascii=False).lower()

        for category, keywords in category_keywords.items():
            if any(keyword in knowledge_text for keyword in keywords):
                categories.append(category)

        # 없으면 기본 카테고리
        if not categories:
            categories = ["GENERAL"]

        return categories

    def detect_knowledge_conflicts(self, new_knowledge: Dict[str, Any]) -> List[Dict[str, Any]]:
        """지식 간 충돌 감지"""
        conflicts = []

        if not self.knowledge_bases:
            return conflicts

        new_knowledge_text = json.dumps(new_knowledge, ensure_ascii=False)

        for base_name, base_knowledge in self.knowledge_bases.items():
            if base_name in self.knowledge_bases:
                base_text = json.dumps(base_knowledge, ensure_ascii=False)

                # 유사도 검사
                similarity = self.calculate_text_similarity(new_knowledge_text, base_text)

                if similarity > 0.8:  # 높은 유사도 = 잠재적 중복
                    conflicts.append({
                        "type": "duplication",
                        "source": base_name,
                        "similarity": similarity,
                        "description": f"지식 베이스 '{base_name}'와 유사도 {similarity:.2f}"
                    })

                # 모순 검사
                contradictions = self.check_knowledge_contradictions(new_knowledge, base_knowledge)
                if contradictions:
                    conflicts.append({
                        "type": "contradiction",
                        "source": base_name,
                        "contradictions": contradictions,
                        "description": f"지식 베이스 '{base_name}'와 {len(contradictions)}개 모순 감지"
                    })

        return conflicts

    def calculate_text_similarity(self, text1: str, text2: str) -> float:
        """텍스트 유사도 계산"""
        # 간단한 단어 기반 유사도
        words1 = set(text1.lower().split())
        words2 = set(text2.lower().split())

        if not words1 or not words2:
            return 0.0

        intersection = len(words1.intersection(words2))
        union = len(words1.union(words2))

        return intersection / union if union > 0 else 0.0

    def check_knowledge_contradictions(self, new_knowledge: Dict[str, Any],
                                    existing_knowledge: Dict[str, Any]) -> List[str]:
        """지식 모순 검사"""
        contradictions = []

        # 숫자 값 모순 검사
        def check_numerical_contradictions(new_data, existing_data, path=""):
            if isinstance(new_data, dict) and isinstance(existing_data, dict):
                for key in set(new_data.keys()) | set(existing_data.keys()):
                    new_path = f"{path}.{key}" if path else key
                    if key in new_data and key in existing_data:
                        if isinstance(new_data[key], (int, float)) and isinstance(existing_data[key], (int, float)):
                            if new_data[key] != existing_data[key]:
                                contradictions.append(f"Numerical contradiction at {new_path}: {new_data[key]} vs {existing_data[key]}")
                        elif isinstance(new_data[key], dict) and isinstance(existing_data[key], dict):
                            check_numerical_contradictions(new_data[key], existing_data[key], new_path)
                        elif isinstance(new_data[key], list) and isinstance(existing_data[key], list):
                            # 리스트 길이 모순
                            if len(new_data[key]) != len(existing_data[key]):
                                contradictions.append(f"List length contradiction at {new_path}: {len(new_data[key])} vs {len(existing_data[key])}")

        check_numerical_contradictions(new_knowledge, existing_knowledge)

        return contradictions

    def resolve_conflicts(self, new_knowledge: Dict[str, Any], conflicts: List[Dict[str, Any]]) -> int:
        """지식 충돌 해결"""
        resolved_count = 0

        for conflict in conflicts:
            if conflict["type"] == "duplication":
                # 중복 지식 합병
                self.merge_duplicate_knowledge(new_knowledge, conflict["source"])
                resolved_count += 1
            elif conflict["type"] == "contradiction":
                # 모순 해결
                self.resolve_knowledge_contradiction(new_knowledge, conflict["source"])
                resolved_count += 1

        return resolved_count

    def merge_duplicate_knowledge(self, new_knowledge: Dict[str, Any], source_name: str) -> None:
        """중복 지식 합병"""
        if source_name in self.knowledge_bases:
            existing_knowledge = self.knowledge_bases[source_name]

            # 타임스탬프와 버전 정보 유지
            if "integration_history" not in existing_knowledge:
                existing_knowledge["integration_history"] = []

            existing_knowledge["integration_history"].append({
                "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
                "action": "merged",
                "source": "new_knowledge"
            })

    def resolve_knowledge_contradiction(self, new_knowledge: Dict[str, Any], source_name: str) -> None:
        """지식 모순 해결"""
        if source_name in self.knowledge_bases:
            existing_knowledge = self.knowledge_bases[source_name]

            # 모순에 대한 기록 추가
            if "contradiction_history" not in existing_knowledge:
                existing_knowledge["contradiction_history"] = []

            existing_knowledge["contradiction_history"].append({
                "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
                "resolved": True,
                "method": "manual_integration"
            })

    def create_knowledge_connections(self, new_knowledge: Dict[str, Any]) -> List[Dict[str, Any]]:
        """지식 연결 생성"""
        connections = []

        if not self.integration_rules["cross_reference"]:
            return connections

        new_knowledge_text = json.dumps(new_knowledge, ensure_ascii=False)

        for base_name, base_knowledge in self.knowledge_bases.items():
            base_text = json.dumps(base_knowledge, ensure_ascii=False)

            # 연결 강도 계산
            strength = self.calculate_text_similarity(new_knowledge_text, base_text)

            if strength > self.integration_rules["confidence_threshold"]:
                connections.append({
                    "from": base_name,
                    "to": "new_knowledge",
                    "strength": strength,
                    "type": "semantic",
                    "categories": self.find_common_categories(new_knowledge, base_knowledge)
                })

                # 양방향 연결 추가
                if base_name not in self.connection_graph:
                    self.connection_graph[base_name] = []

                self.connection_graph[base_name].append({
                    "target": "new_knowledge",
                    "strength": strength,
                    "type": "semantic"
                })

        return connections

    def find_common_categories(self, knowledge1: Dict[str, Any], knowledge2: Dict[str, Any]) -> List[str]:
        """공통 카테고리 찾기"""
        categories1 = self.identify_knowledge_categories(knowledge1)
        categories2 = self.identify_knowledge_categories(knowledge2)

        return list(set(categories1) & set(categories2))

    def update_knowledge_graph(self, connections: List[Dict[str, Any]]) -> None:
        """지식 그래프 업데이트"""
        # 연결 정보 그래프에 저장
        for connection in connections:
            source = connection["from"]
            target = connection["to"]

            if source not in self.connection_graph:
                self.connection_graph[source] = []

            self.connection_graph[source].append({
                "target": target,
                "strength": connection["strength"],
                "type": connection["type"],
                "categories": connection["categories"]
            })

    def save_integrated_knowledge(self, knowledge: Dict[str, Any],
                                categories: List[str]) -> List[str]:
        """통합 지식 저장"""
        saved_sources = []
        knowledge_dir = Path(".moai/research/knowledge/")
        knowledge_dir.mkdir(parents=True, exist_ok=True)

        # 타임스탬프 기반 파일명
        timestamp = int(time.time())

        for category in categories:
            filename = f"{category.lower()}_knowledge_{timestamp}.json"
            file_path = knowledge_dir / filename

            try:
                # 지식 메타데이터 추가
                knowledge_with_metadata = {
                    **knowledge,
                    "integration_metadata": {
                        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
                        "categories": categories,
                        "integration_rules": self.integration_rules
                    }
                }

                with open(file_path, 'w', encoding='utf-8') as f:
                    json.dump(knowledge_with_metadata, f, ensure_ascii=False, indent=2)

                saved_sources.append(filename)

            except Exception as e:
                print(f"Failed to save knowledge to {filename}: {str(e)}")

        return saved_sources

    def generate_integration_recommendations(self, knowledge: Dict[str, Any],
                                          connections: List[Dict[str, Any]],
                                          conflicts: List[Dict[str, Any]]) -> List[str]:
        """통합 추천 생성"""
        recommendations = []

        # 연결 수 기반 추천
        if len(connections) > 5:
            recommendations.append(f"강력한 지식 연결 {len(connections)}개 감지 - 지식 그래프 확장")

        # 충돌 수 기반 추천
        if len(conflicts) > 2:
            recommendations.append(f"지식 충돌 {len(conflicts)}개 감지 - 주의 깊은 검토 필요")

        # 카테고리 기반 추천
        categories = self.identify_knowledge_categories(knowledge)
        if "RESEARCH" in categories:
            recommendations.append("연구 결과 추가 문서화 필요")

        if "ANALYSIS" in categories:
            recommendations.append("분석 결과 기반 추천 사항 정리")

        if "KNOWLEDGE" in categories:
            recommendations.append("지식 활용 가이드 제작")

        if "INSIGHT" in categories:
            recommendations.append("인사이트 적용 실행 계획 수립")

        # 저장량 기반 추천
        total_size = sum(len(json.dumps(kb)) for kb in self.knowledge_bases.values())
        if total_size > 10 * 1024 * 1024:  # 10MB
            recommendations.append("지식 베이스 정리 및 보관 필요")

        return recommendations

    def query_knowledge(self, query: str, category: Optional[str] = None,
                      limit: int = 10) -> Dict[str, Any]:
        """지식 쿼리"""
        query_result = {
            "query": query,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "total_results": 0,
            "results": [],
            "categories": [],
            "recommendations": []
        }

        # 지식 베이스 로드
        self.knowledge_bases = self.load_all_knowledge_bases()

        # 쿼리에 맞는 지식 검색
        for base_name, base_knowledge in self.knowledge_bases.items():
            if category and category not in self.identify_knowledge_categories(base_knowledge):
                continue

            # 유사도 검사
            knowledge_text = json.dumps(base_knowledge, ensure_ascii=False)
            similarity = self.calculate_text_similarity(query, knowledge_text)

            if similarity > self.integration_rules["confidence_threshold"]:
                query_result["results"].append({
                    "source": base_name,
                    "similarity": similarity,
                    "knowledge": base_knowledge,
                    "categories": self.identify_knowledge_categories(base_knowledge)
                })

        # 유사도 순 정렬
        query_result["results"].sort(key=lambda x: x["similarity"], reverse=True)

        # 제한 적용
        query_result["results"] = query_result["results"][:limit]
        query_result["total_results"] = len(query_result["results"])

        # 카테고리 수집
        all_categories = set()
        for result in query_result["results"]:
            all_categories.update(result["categories"])
        query_result["categories"] = list(all_categories)

        # 추천 생성
        if query_result["total_results"] == 0:
            query_result["recommendations"].append("관련 지식을 찾을 수 없습니다 - 새로운 지식 생성")
        else:
            query_result["recommendations"].append(f"관련 지식 {query_result['total_results']}개 발견")

        return query_result


def integrate_knowledge_with_hub(knowledge: Dict[str, Any], source_type: str = "manual") -> Dict[str, Any]:
    """지식 통합 허브로 지식 통합"""
    hub = KnowledgeIntegrationHub()
    return hub.integrate_knowledge(knowledge, source_type)


def query_knowledge_with_hub(query: str, category: Optional[str] = None,
                           limit: int = 10) -> Dict[str, Any]:
    """지식 통합 허브로 지식 쿼리"""
    hub = KnowledgeIntegrationHub()
    return hub.query_knowledge(query, category, limit)


def get_knowledge_hub_status() -> Dict[str, Any]:
    """지식 허브 상태 반환"""
    hub = KnowledgeIntegrationHub()
    return {
        "knowledge_bases_count": len(hub.knowledge_bases),
        "connection_graph_size": len(hub.connection_graph),
        "integration_rules": hub.integration_rules,
        "update_timestamps": hub.update_timestamps
    }


# 표준 Skill 인터페이스 구현
def main() -> None:
    """Skill 메인 함수"""
    try:
        # 인자 파싱
        if len(sys.argv) < 2:
            print(json.dumps({
                "error": "Usage: python3 knowledge_integration_hub.py <action> [args...]"
            }))
            sys.exit(1)

        action = sys.argv[1]

        if action == "integrate":
            if len(sys.argv) < 3:
                print(json.dumps({
                    "error": "Usage: python3 knowledge_integration_hub.py integrate <knowledge_json> [source_type]"
                }))
                sys.exit(1)

            try:
                knowledge = json.loads(sys.argv[2])
                source_type = sys.argv[3] if len(sys.argv) > 3 else "manual"
            except json.JSONDecodeError:
                print(json.dumps({
                    "error": "Invalid JSON format for knowledge"
                }))
                sys.exit(1)

            result = integrate_knowledge_with_hub(knowledge, source_type)

        elif action == "query":
            if len(sys.argv) < 3:
                print(json.dumps({
                    "error": "Usage: python3 knowledge_integration_hub.py query <query_string> [category] [limit]"
                }))
                sys.exit(1)

            query = sys.argv[2]
            category = sys.argv[3] if len(sys.argv) > 3 else None
            limit = int(sys.argv[4]) if len(sys.argv) > 4 else 10

            result = query_knowledge_with_hub(query, category, limit)

        elif action == "status":
            result = get_knowledge_hub_status()

        else:
            print(json.dumps({
                "error": f"Unknown action: {action}. Available actions: integrate, query, status"
            }))
            sys.exit(1)

        # 결과 출력
        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        error_result = {
            "error": f"Knowledge integration hub failed: {str(e)}",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()