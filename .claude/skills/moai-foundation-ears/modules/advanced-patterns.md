# Advanced EARS Requirement Patterns

**Version**: 4.0.0
**Focus**: AI-powered validation, NLP analysis, enterprise automation

---

## Advanced EARS Patterns

### Pattern 1: Automated EARS Pattern Detection with NLP

**AI-Powered Requirement Classification**:
```python
from typing import Optional
import re
from enum import Enum

class EARSType(Enum):
    UBIQUITOUS = "ubiquitous"      # The system SHALL
    EVENT_DRIVEN = "event_driven"  # WHEN... the system SHALL
    STATE_DRIVEN = "state_driven"  # WHILE... the system SHALL
    OPTIONAL = "optional"           # WHERE... the system SHALL
    UNWANTED = "unwanted"          # IF... THEN the system SHALL

class EARSDetector:
    """Detect EARS patterns automatically using NLP and regex"""

    PATTERNS = {
        EARSType.EVENT_DRIVEN: re.compile(
            r'^when\s+(.+?),?\s+(?:the\s+)?system\s+shall\s+(.+)$',
            re.IGNORECASE
        ),
        EARSType.STATE_DRIVEN: re.compile(
            r'^while\s+(.+?),?\s+(?:the\s+)?system\s+shall\s+(.+)$',
            re.IGNORECASE
        ),
        EARSType.OPTIONAL: re.compile(
            r'^where\s+(.+?),?\s+(?:the\s+)?system\s+shall\s+(.+)$',
            re.IGNORECASE
        ),
        EARSType.UNWANTED: re.compile(
            r'^if\s+(.+?),?\s+then\s+(?:the\s+)?system\s+shall\s+(.+)$',
            re.IGNORECASE
        ),
        EARSType.UBIQUITOUS: re.compile(
            r'^(?:the\s+)?system\s+shall\s+(.+)$',
            re.IGNORECASE
        ),
    }

    @staticmethod
    async def detect_pattern(requirement: str) -> Optional[tuple[EARSType, dict]]:
        """Detect EARS pattern in requirement statement"""
        for ears_type, pattern in EARSDetector.PATTERNS.items():
            match = pattern.match(requirement.strip())
            if match:
                groups = match.groups()
                return ears_type, {
                    'trigger': groups[0] if len(groups) > 0 else None,
                    'action': groups[-1],
                    'confidence': EARSDetector._calculate_confidence(
                        requirement, ears_type
                    )
                }
        return None, {}

    @staticmethod
    def _calculate_confidence(requirement: str, ears_type: EARSType) -> float:
        """Calculate pattern match confidence (0.0-1.0)"""
        base_score = 0.8

        # Boost confidence for clear keywords
        if ears_type == EARSType.EVENT_DRIVEN and 'when' in requirement.lower():
            base_score = 0.95
        elif ears_type == EARSType.STATE_DRIVEN and 'while' in requirement.lower():
            base_score = 0.95

        # Check for acceptance criteria presence
        if '(' in requirement and ')' in requirement:
            base_score -= 0.05  # Slightly lower for embedded criteria

        return min(1.0, base_score)

    @classmethod
    async def validate_requirement_set(cls, requirements: list[str]) -> dict:
        """Validate entire requirement set"""
        results = {
            'valid': True,
            'requirements': [],
            'issues': [],
            'pattern_distribution': {}
        }

        for idx, req in enumerate(requirements):
            ears_type, details = await cls.detect_pattern(req)

            results['requirements'].append({
                'index': idx,
                'text': req,
                'pattern': ears_type.value if ears_type else 'unknown',
                'confidence': details.get('confidence', 0.0),
                'valid': ears_type is not None
            })

            if not ears_type:
                results['issues'].append(f'Requirement {idx}: No EARS pattern detected')
                results['valid'] = False

            # Count pattern distribution
            pattern_name = ears_type.value if ears_type else 'unknown'
            results['pattern_distribution'][pattern_name] = \
                results['pattern_distribution'].get(pattern_name, 0) + 1

        return results
```

### Pattern 2: Requirement Completeness Checker

**Multi-Criteria Validation Framework**:
```python
from dataclasses import dataclass
from typing import List, Optional

@dataclass
class RequirementCompleteness:
    has_actor: bool
    has_precondition: bool
    has_action: bool
    has_postcondition: bool
    has_acceptance_criteria: bool
    completeness_score: float

class RequirementCompletenessChecker:
    """Check requirement completeness"""

    REQUIRED_ELEMENTS = {
        'actor': r'(?:user|system|actor|admin|manager)',
        'action': r'(?:shall|must|should|will)',
        'condition': r'(?:when|while|where|if)',
        'measurable': r'(?:\d+|all|none|some)',
    }

    @classmethod
    def check_completeness(cls, requirement: str) -> RequirementCompleteness:
        """Check if requirement has all essential elements"""

        has_actor = bool(re.search(cls.REQUIRED_ELEMENTS['actor'],
                                   requirement, re.IGNORECASE))
        has_action = bool(re.search(cls.REQUIRED_ELEMENTS['action'],
                                    requirement, re.IGNORECASE))
        has_precondition = bool(re.search(r'given|when|while|where',
                                         requirement, re.IGNORECASE))
        has_postcondition = bool(re.search(r'then|result|returned|output',
                                          requirement, re.IGNORECASE))
        has_acceptance_criteria = bool(re.search(r'\(.*?\)|@|acceptance',
                                                 requirement))

        # Calculate completeness score
        score = sum([
            has_actor * 0.2,
            has_action * 0.25,
            has_precondition * 0.15,
            has_postcondition * 0.15,
            has_acceptance_criteria * 0.25
        ])

        return RequirementCompleteness(
            has_actor=has_actor,
            has_precondition=has_precondition,
            has_action=has_action,
            has_postcondition=has_postcondition,
            has_acceptance_criteria=has_acceptance_criteria,
            completeness_score=score
        )

    @classmethod
    def improve_requirement(cls, requirement: str) -> str:
        """Suggest improvements for incomplete requirement"""
        completeness = cls.check_completeness(requirement)

        improvements = []
        if not completeness.has_actor:
            improvements.append("Add actor (who does this): e.g., 'As a user'")
        if not completeness.has_precondition:
            improvements.append("Add precondition: e.g., 'When user is logged in'")
        if not completeness.has_action:
            improvements.append("Add action verb: e.g., 'the system shall validate'")
        if not completeness.has_postcondition:
            improvements.append("Add postcondition: e.g., 'then a token is returned'")
        if not completeness.has_acceptance_criteria:
            improvements.append("Add acceptance criteria: e.g., '(within 200ms)'")

        return "\n".join([f"- {imp}" for imp in improvements])
```

### Pattern 3: Requirement Dependency Graph

**AI-Powered Dependency Detection**:
```python
from collections import defaultdict, deque

class RequirementDependencyGraph:
    """Build and analyze requirement dependency relationships"""

    def __init__(self):
        self.graph = defaultdict(set)
        self.requirements = {}

    def add_requirement(self, req_id: str, requirement: str):
        """Add requirement to graph"""
        self.requirements[req_id] = requirement
        if req_id not in self.graph:
            self.graph[req_id] = set()

    def detect_dependencies(self, req_id: str, other_req_id: str) -> Optional[str]:
        """Detect if req_id depends on other_req_id"""
        req1 = self.requirements[req_id]
        req2 = self.requirements[other_req_id]

        # Extract keywords from both requirements
        keywords1 = set(self._extract_keywords(req1))
        keywords2 = set(self._extract_keywords(req2))

        # High overlap suggests dependency
        overlap = keywords1 & keywords2
        similarity = len(overlap) / max(len(keywords1), len(keywords2))

        if similarity > 0.5:
            return f"Potential dependency detected ({similarity:.2%} similarity)"

        return None

    def find_circular_dependencies(self) -> List[List[str]]:
        """Find circular dependency chains"""
        cycles = []
        visited = set()

        def dfs(node, path, rec_stack):
            visited.add(node)
            rec_stack.add(node)

            for neighbor in self.graph[node]:
                if neighbor not in visited:
                    dfs(neighbor, path + [neighbor], rec_stack)
                elif neighbor in rec_stack:
                    cycles.append(path + [neighbor])

            rec_stack.remove(node)

        for req_id in self.requirements:
            if req_id not in visited:
                dfs(req_id, [req_id], set())

        return cycles

    def topological_sort(self) -> List[str]:
        """Sort requirements by dependency order"""
        in_degree = defaultdict(int)

        for req_id in self.requirements:
            if req_id not in in_degree:
                in_degree[req_id] = 0
            for dependent in self.graph[req_id]:
                in_degree[dependent] += 1

        queue = deque([req for req in self.requirements
                      if in_degree[req] == 0])
        sorted_reqs = []

        while queue:
            req_id = queue.popleft()
            sorted_reqs.append(req_id)

            for dependent in self.graph[req_id]:
                in_degree[dependent] -= 1
                if in_degree[dependent] == 0:
                    queue.append(dependent)

        return sorted_reqs
```

### Pattern 4: Requirement Traceability Matrix Generation

**Automated RTM Creation**:
```python
class RequirementTraceabilityMatrix:
    """Generate and maintain requirement traceability"""

    def __init__(self, requirements: dict):
        self.requirements = requirements
        self.rtm = defaultdict(lambda: defaultdict(list))

    def link_to_test_cases(self, req_id: str, test_case_ids: List[str]):
        """Link requirement to test cases"""
        self.rtm[req_id]['test_cases'] = test_case_ids

    def link_to_implementation(self, req_id: str, file_paths: List[str]):
        """Link requirement to implementation files"""
        self.rtm[req_id]['implementation'] = file_paths

    def link_to_user_story(self, req_id: str, story_id: str):
        """Link requirement to user story"""
        self.rtm[req_id]['user_story'] = story_id

    def generate_coverage_report(self) -> dict:
        """Generate traceability coverage report"""
        report = {
            'total_requirements': len(self.requirements),
            'untested': [],
            'untested_coverage': 0.0,
            'unimplemented': [],
            'unimplemented_coverage': 0.0,
            'fully_traced': []
        }

        for req_id, requirement in self.requirements.items():
            traced = self.rtm[req_id]

            if not traced.get('test_cases'):
                report['untested'].append(req_id)
            if not traced.get('implementation'):
                report['unimplemented'].append(req_id)
            if traced.get('test_cases') and traced.get('implementation'):
                report['fully_traced'].append(req_id)

        report['untested_coverage'] = (
            len(report['fully_traced']) / report['total_requirements']
        )
        report['unimplemented_coverage'] = (
            1.0 - len(report['unimplemented']) / report['total_requirements']
        )

        return report
```

---

**Last Updated**: 2025-11-22
