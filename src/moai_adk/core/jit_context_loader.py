#!/usr/bin/env python3
"""
JIT Context Loading System

Phase-based token optimization system that achieves 85%+ efficiency through
intelligent context loading, skill filtering, and budget management.
"""

import hashlib
import json
import logging
import os
import sys
import time
from collections import OrderedDict
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import psutil

logger = logging.getLogger(__name__)


class Phase(Enum):
    """Development phases with different token budgets"""

    SPEC = "spec"  # 30K budget
    RED = "red"  # 25K budget
    GREEN = "green"  # 25K budget
    REFACTOR = "refactor"  # 20K budget
    SYNC = "sync"  # 40K budget
    DEBUG = "debug"  # 15K budget
    PLANNING = "planning"  # 35K budget


@dataclass
class PhaseConfig:
    """Configuration for each development phase"""

    max_tokens: int
    essential_skills: List[str]
    essential_documents: List[str]
    cache_clear_on_phase_change: bool = False
    context_compression: bool = True


@dataclass
class ContextMetrics:
    """Metrics for context loading performance"""

    load_time: float
    token_count: int
    cache_hit: bool
    phase: str
    skills_loaded: int
    docs_loaded: int
    compression_ratio: float = 1.0
    memory_usage: int = 0


@dataclass
class SkillInfo:
    """Information about a skill for filtering decisions"""

    name: str
    path: str
    size: int
    tokens: int
    categories: List[str]
    dependencies: List[str] = field(default_factory=list)
    priority: int = 1  # 1=high, 2=medium, 3=low
    last_used: Optional[datetime] = None


@dataclass
class ContextEntry:
    """Entry in the context cache"""

    key: str
    content: Any
    token_count: int
    created_at: datetime
    last_accessed: datetime
    access_count: int = 0
    phase: Optional[str] = None


class PhaseDetector:
    """Intelligently detects current development phase from context"""

    def __init__(self):
        self.phase_patterns = {
            Phase.SPEC: [
                r"/moai:1-plan",
                r"SPEC-\d+",
                r"spec|requirements|design",
                r"create.*spec|define.*requirements",
                r"plan.*feature|design.*system",
            ],
            Phase.RED: [
                r"/moai:2-run.*RED",
                r"test.*fail|failing.*test",
                r"red.*phase|tdd.*red",
                r"2-run.*RED|write.*test",
                r"create.*test.*failure",
            ],
            Phase.GREEN: [
                r"/moai:2-run.*GREEN",
                r"test.*pass|passing.*test",
                r"green.*phase|tdd.*green",
                r"2-run.*GREEN|minimal.*implementation",
                r"make.*test.*pass|implement.*minimum",
            ],
            Phase.REFACTOR: [
                r"/moai:2-run.*REFACTOR",
                r"refactor|clean.*code",
                r"quality.*improvement|improve.*code",
                r"optimize.*code|code.*cleanup",
            ],
            Phase.SYNC: [
                r"/moai:3-sync",
                r"documentation.*sync|sync.*docs",
                r"generate.*docs|create.*documentation",
                r"update.*documentation",
            ],
            Phase.DEBUG: [
                r"debug|troubleshoot",
                r"error.*analysis|analyze.*error",
                r"fix.*bug|error.*investigation",
                r"problem.*solving",
            ],
            Phase.PLANNING: [
                r"plan.*implementation|implementation.*plan",
                r"architecture.*design|design.*architecture",
                r"task.*decomposition|breakdown.*task",
                r"plan.*development|development.*planning",
            ],
        }

        self.last_phase = Phase.SPEC
        self.phase_history = []
        self.max_history = 10

    def detect_phase(self, user_input: str, conversation_history: List[str] = None) -> Phase:
        """Detect current phase from user input and conversation context"""
        import re

        # Combine user input with recent conversation history
        context = user_input.lower()
        if conversation_history:
            recent_history = " ".join(conversation_history[-3:])  # Last 3 messages
            context += " " + recent_history.lower()

        # Score each phase based on pattern matches
        phase_scores = {}
        for phase, patterns in self.phase_patterns.items():
            score = 0
            for pattern in patterns:
                matches = len(re.findall(pattern, context, re.IGNORECASE))
                score += matches
            phase_scores[phase] = score

        # Find phase with highest score
        if phase_scores:
            best_phase = max(phase_scores, key=phase_scores.get)
        else:
            best_phase = self.last_phase

        # If no clear winner (all scores 0), use last known phase
        if phase_scores[best_phase] == 0:
            best_phase = self.last_phase

        # Update history
        if best_phase != self.last_phase:
            self.phase_history.append(
                {
                    "from": self.last_phase.value,
                    "to": best_phase.value,
                    "timestamp": datetime.now(),
                    "context": (user_input[:100] + "..." if len(user_input) > 100 else user_input),
                }
            )

            if len(self.phase_history) > self.max_history:
                self.phase_history.pop(0)

        self.last_phase = best_phase
        return best_phase

    def get_phase_config(self, phase: Phase) -> PhaseConfig:
        """Get configuration for a specific phase"""
        configs = {
            Phase.SPEC: PhaseConfig(
                max_tokens=30000,
                essential_skills=[
                    "moai-foundation-ears",
                    "moai-foundation-specs",
                    "moai-essentials-review",
                    "moai-domain-backend",
                    "moai-lang-python",
                    "moai-core-spec-authoring",
                ],
                essential_documents=[
                    ".moai/specs/template.md",
                    ".claude/skills/moai-foundation-ears/SKILL.md",
                ],
                cache_clear_on_phase_change=True,
            ),
            Phase.RED: PhaseConfig(
                max_tokens=25000,
                essential_skills=[
                    "moai-domain-testing",
                    "moai-foundation-trust",
                    "moai-essentials-review",
                    "moai-core-code-reviewer",
                    "moai-essentials-debug",
                    "moai-lang-python",
                ],
                essential_documents=[
                    ".moai/specs/{spec_id}/spec.md",
                    ".claude/skills/moai-domain-testing/SKILL.md",
                ],
            ),
            Phase.GREEN: PhaseConfig(
                max_tokens=25000,
                essential_skills=[
                    "moai-lang-python",
                    "moai-domain-backend",
                    "moai-essentials-review",
                ],
                essential_documents=[".moai/specs/{spec_id}/spec.md"],
            ),
            Phase.REFACTOR: PhaseConfig(
                max_tokens=20000,
                essential_skills=[
                    "moai-essentials-refactor",
                    "moai-essentials-review",
                    "moai-core-code-reviewer",
                    "moai-essentials-debug",
                ],
                essential_documents=[
                    "src/{module}/current_implementation.py",
                    ".claude/skills/moai-essentials-refactor/SKILL.md",
                ],
            ),
            Phase.SYNC: PhaseConfig(
                max_tokens=40000,
                essential_skills=[
                    "moai-docs-unified",
                    "moai-nextra-architecture",
                    "moai-core-spec-authoring",
                    "moai-cc-configuration",
                ],
                essential_documents=[
                    ".moai/specs/{spec_id}/implementation.md",
                    ".moai/specs/{spec_id}/test-cases.md",
                ],
                cache_clear_on_phase_change=True,
            ),
            Phase.DEBUG: PhaseConfig(
                max_tokens=15000,
                essential_skills=["moai-essentials-debug", "moai-core-code-reviewer"],
                essential_documents=[".moai/logs/latest_error.log"],
            ),
            Phase.PLANNING: PhaseConfig(
                max_tokens=35000,
                essential_skills=[
                    "moai-core-practices",
                    "moai-essentials-review",
                    "moai-foundation-specs",
                ],
                essential_documents=[".moai/specs/{spec_id}/spec.md"],
            ),
        }

        return configs.get(phase, configs[Phase.SPEC])


class SkillFilterEngine:
    """Intelligently filters and selects skills based on phase and context"""

    def __init__(self, skills_dir: str = ".claude/skills"):
        self.skills_dir = Path(skills_dir)
        self.skills_cache: Dict[str, Any] = {}
        self.skill_index: Dict[str, SkillInfo] = {}
        self.phase_preferences = self._load_phase_preferences()
        self._build_skill_index()

    def _load_phase_preferences(self) -> Dict[str, Dict[str, int]]:
        """Load phase-based skill preferences"""
        return {
            "spec": {
                "moai-foundation-ears": 1,
                "moai-foundation-specs": 1,
                "moai-essentials-review": 2,
                "moai-domain-backend": 2,
                "moai-lang-python": 3,
                "moai-core-spec-authoring": 1,
            },
            "red": {
                "moai-domain-testing": 1,
                "moai-foundation-trust": 1,
                "moai-essentials-review": 2,
                "moai-core-code-reviewer": 2,
                "moai-essentials-debug": 2,
                "moai-lang-python": 3,
            },
            "green": {
                "moai-lang-python": 1,
                "moai-domain-backend": 1,
                "moai-essentials-review": 2,
            },
            "refactor": {
                "moai-essentials-refactor": 1,
                "moai-essentials-review": 2,
                "moai-core-code-reviewer": 2,
                "moai-essentials-debug": 3,
            },
            "sync": {
                "moai-docs-unified": 1,
                "moai-nextra-architecture": 1,
                "moai-core-spec-authoring": 2,
                "moai-cc-configuration": 2,
            },
            "debug": {"moai-essentials-debug": 1, "moai-core-code-reviewer": 2},
            "planning": {
                "moai-core-practices": 1,
                "moai-essentials-review": 2,
                "moai-foundation-specs": 2,
            },
        }

    def _build_skill_index(self):
        """Build index of all available skills with metadata"""
        if not self.skills_dir.exists():
            logger.warning(f"Skills directory not found: {self.skills_dir}")
            return

        for skill_dir in self.skills_dir.iterdir():
            if skill_dir.is_dir():
                skill_file = skill_dir / "SKILL.md"
                if skill_file.exists():
                    skill_info = self._analyze_skill(skill_file)
                    if skill_info:
                        self.skill_index[skill_info.name] = skill_info

    def _analyze_skill(self, skill_file: Path) -> Optional[SkillInfo]:
        """Analyze a skill file to extract metadata"""
        try:
            stat = skill_file.stat()

            # Read skill content to extract categories
            with open(skill_file, "r", encoding="utf-8") as f:
                content = f.read()

            # Extract skill name from directory
            skill_name = skill_file.parent.name

            # Estimate tokens (rough approximation: 1 token ≈ 4 characters)
            estimated_tokens = len(content) // 4

            # Extract categories from content (look for keywords)
            categories = []
            if any(keyword in content.lower() for keyword in ["python", "javascript", "typescript", "go", "rust"]):
                categories.append("language")
            if any(keyword in content.lower() for keyword in ["backend", "frontend", "database", "security"]):
                categories.append("domain")
            if any(keyword in content.lower() for keyword in ["testing", "debug", "refactor", "review"]):
                categories.append("development")
            if any(keyword in content.lower() for keyword in ["foundation", "essential", "core"]):
                categories.append("core")

            return SkillInfo(
                name=skill_name,
                path=str(skill_file),
                size=stat.st_size,
                tokens=estimated_tokens,
                categories=categories,
                priority=1,
            )

        except Exception as e:
            logger.error(f"Error analyzing skill {skill_file}: {e}")
            return None

    def filter_skills(self, phase: Phase, token_budget: int, context: Dict[str, Any] = None) -> List[SkillInfo]:
        """Filter skills based on phase, token budget, and context"""
        phase_name = phase.value
        preferences = self.phase_preferences.get(phase_name, {})

        # Get all relevant skills for this phase
        relevant_skills = []
        for skill_name, skill_info in self.skill_index.items():
            # Check if skill is relevant for this phase
            if skill_name in preferences:
                # Apply priority from preferences
                skill_info.priority = preferences[skill_name]
                relevant_skills.append(skill_info)

        # Sort by priority and token efficiency
        relevant_skills.sort(key=lambda s: (s.priority, s.tokens))

        # Select skills within token budget
        selected_skills = []
        used_tokens = 0

        for skill in relevant_skills:
            if used_tokens + skill.tokens <= token_budget:
                selected_skills.append(skill)
                used_tokens += skill.tokens
            else:
                break

        return selected_skills

    def get_skill_stats(self) -> Dict[str, Any]:
        """Get statistics about available skills"""
        total_skills = len(self.skill_index)
        total_tokens = sum(skill.tokens for skill in self.skill_index.values())

        categories: Dict[str, int] = {}
        for skill in self.skill_index.values():
            for category in skill.categories:
                categories[category] = categories.get(category, 0) + 1

        return {
            "total_skills": total_skills,
            "total_tokens": total_tokens,
            "categories": categories,
            "average_tokens_per_skill": (total_tokens / total_skills if total_skills > 0 else 0),
        }


class TokenBudgetManager:
    """Manages token budgets and usage across phases"""

    def __init__(self, max_total_tokens: int = 180000):
        self.max_total_tokens = max_total_tokens
        self.phase_budgets = self._initialize_phase_budgets()
        self.current_usage = 0
        self.usage_history: List[Dict[str, Any]] = []
        self.budget_warnings: List[Dict[str, Any]] = []

    def _initialize_phase_budgets(self) -> Dict[str, int]:
        """Initialize token budgets for each phase"""
        return {
            "spec": 30000,
            "red": 25000,
            "green": 25000,
            "refactor": 20000,
            "sync": 40000,
            "debug": 15000,
            "planning": 35000,
        }

    def check_budget(self, phase: Phase, requested_tokens: int) -> Tuple[bool, int]:
        """Check if requested tokens fit in budget"""
        phase_budget = self.phase_budgets.get(phase.value, 30000)

        if requested_tokens <= phase_budget:
            return True, phase_budget - requested_tokens
        else:
            return False, phase_budget

    def record_usage(self, phase: Phase, tokens_used: int, context: str = ""):
        """Record actual token usage"""
        usage_entry = {
            "phase": phase.value,
            "tokens": tokens_used,
            "timestamp": datetime.now(),
            "context": context[:100],
        }

        self.usage_history.append(usage_entry)
        self.current_usage += tokens_used

        # Keep only recent history (last 50 entries)
        if len(self.usage_history) > 50:
            self.usage_history.pop(0)

        # Check for budget warnings
        phase_budget = self.phase_budgets.get(phase.value, 30000)
        if tokens_used > phase_budget:
            warning = f"Phase {phase.value} exceeded budget: {tokens_used} > {phase_budget}"
            self.budget_warnings.append({"warning": warning, "timestamp": datetime.now()})
            logger.warning(warning)

    def get_efficiency_metrics(self) -> Dict[str, Any]:
        """Calculate and return efficiency metrics"""
        if not self.usage_history:
            return {
                "efficiency_score": 0,
                "average_phase_efficiency": {},
                "budget_compliance": 100,
                "total_usage": 0,
            }

        # Calculate efficiency by phase
        phase_usage = {}
        phase_efficiency = {}

        for entry in self.usage_history:
            phase = entry["phase"]
            tokens = entry["tokens"]
            budget = self.phase_budgets.get(phase, 30000)

            if phase not in phase_usage:
                phase_usage[phase] = {"total": 0, "count": 0, "over_budget": 0}

            phase_usage[phase]["total"] += tokens
            phase_usage[phase]["count"] += 1
            if tokens > budget:
                phase_usage[phase]["over_budget"] += 1

        # Calculate efficiency scores
        for phase, usage in phase_usage.items():
            budget = self.phase_budgets.get(phase, 30000)
            actual = usage["total"] / usage["count"] if usage["count"] > 0 else 0
            efficiency = min(100, (budget / actual * 100) if actual > 0 else 100)
            phase_efficiency[phase] = efficiency

        # Overall efficiency
        overall_efficiency = sum(phase_efficiency.values()) / len(phase_efficiency) if phase_efficiency else 0

        # Budget compliance
        total_entries = len(self.usage_history)
        over_budget_entries = sum(usage["over_budget"] for usage in phase_usage.values())
        budget_compliance = ((total_entries - over_budget_entries) / total_entries * 100) if total_entries > 0 else 100

        return {
            "efficiency_score": overall_efficiency,
            "average_phase_efficiency": phase_efficiency,
            "budget_compliance": budget_compliance,
            "total_usage": self.current_usage,
            "phase_usage": phase_usage,
        }


class ContextCache:
    """LRU Cache for context entries with phase-aware eviction"""

    def __init__(self, max_size: int = 100, max_memory_mb: int = 50):
        self.max_size = max_size
        self.max_memory_bytes = max_memory_mb * 1024 * 1024
        self.cache: OrderedDict[str, ContextEntry] = OrderedDict()
        self.current_memory = 0
        self.hits = 0
        self.misses = 0
        self.evictions = 0

    def _calculate_memory_usage(self, entry: ContextEntry) -> int:
        """Calculate memory usage of a cache entry"""

        # Rough estimation of memory usage
        content_size = sys.getsizeof(entry.content)
        entry_overhead = sys.getsizeof(entry.key) + sys.getsizeof(str(entry.token_count))
        total_size = content_size + entry_overhead

        return total_size

    def get(self, key: str) -> Optional[ContextEntry]:
        """Get entry from cache"""
        if key in self.cache:
            entry = self.cache[key]
            entry.last_accessed = datetime.now()
            entry.access_count += 1
            self.cache.move_to_end(key)
            self.hits += 1
            return entry

        self.misses += 1
        return None

    def put(self, key: str, content: Any, token_count: int, phase: Optional[str] = None):
        """Put entry in cache with LRU eviction"""
        entry = ContextEntry(
            key=key,
            content=content,
            token_count=token_count,
            created_at=datetime.now(),
            last_accessed=datetime.now(),
            phase=phase,
        )

        entry_memory = self._calculate_memory_usage(entry)

        # Check if we need to evict entries
        while len(self.cache) >= self.max_size or self.current_memory + entry_memory > self.max_memory_bytes:
            if not self.cache:
                break

            oldest_key = next(iter(self.cache))
            oldest_entry = self.cache.pop(oldest_key)
            self.current_memory -= self._calculate_memory_usage(oldest_entry)
            self.evictions += 1

        # Add new entry
        self.cache[key] = entry
        self.current_memory += entry_memory

    def clear_phase(self, phase: str):
        """Clear all entries for a specific phase"""
        keys_to_remove = [key for key, entry in self.cache.items() if entry.phase == phase]
        for key in keys_to_remove:
            entry = self.cache.pop(key)
            self.current_memory -= self._calculate_memory_usage(entry)

    def clear(self):
        """Clear all cache entries"""
        self.cache.clear()
        self.current_memory = 0

    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics"""
        total_requests = self.hits + self.misses
        hit_rate = (self.hits / total_requests * 100) if total_requests > 0 else 0

        return {
            "entries": len(self.cache),
            "memory_usage_bytes": self.current_memory,
            "memory_usage_mb": self.current_memory / (1024 * 1024),
            "hit_rate": hit_rate,
            "hits": self.hits,
            "misses": self.misses,
            "evictions": self.evictions,
        }


class JITContextLoader:
    """Main JIT Context Loading System orchestrator"""

    def __init__(self, cache_size: int = 100, cache_memory_mb: int = 50):
        self.phase_detector = PhaseDetector()
        self.skill_filter = SkillFilterEngine()
        self.token_manager = TokenBudgetManager()
        self.context_cache = ContextCache(cache_size, cache_memory_mb)

        self.metrics_history: List[ContextMetrics] = []
        self.current_phase = Phase.SPEC

        # Performance monitoring
        self.performance_stats: Dict[str, Any] = {
            "total_loads": 0,
            "average_load_time": 0,
            "cache_hit_rate": 0,
            "efficiency_score": 0,
        }

    async def load_context(
        self,
        user_input: str,
        conversation_history: List[str] = None,
        context: Dict[str, Any] = None,
    ) -> Tuple[Dict[str, Any], ContextMetrics]:
        """Load optimized context based on current phase and requirements"""
        start_time = time.time()

        # Detect current phase
        self.current_phase = self.phase_detector.detect_phase(user_input, conversation_history or [])
        phase_config = self.phase_detector.get_phase_config(self.current_phase)

        # Generate cache key
        cache_key = self._generate_cache_key(self.current_phase, user_input, context or {})

        # Check cache first
        cached_entry = self.context_cache.get(cache_key)
        if cached_entry:
            metrics = ContextMetrics(
                load_time=time.time() - start_time,
                token_count=cached_entry.token_count,
                cache_hit=True,
                phase=self.current_phase.value,
                skills_loaded=len(cached_entry.content.get("skills", [])),
                docs_loaded=len(cached_entry.content.get("documents", [])),
                memory_usage=psutil.Process().memory_info().rss,
            )

            self._record_metrics(metrics)
            return cached_entry.content, metrics

        # Load fresh context
        context_data = await self._build_context(self.current_phase, phase_config, context or {})

        # Calculate total tokens
        total_tokens = self._calculate_total_tokens(context_data)

        # Check token budget
        within_budget, remaining_budget = self.token_manager.check_budget(self.current_phase, total_tokens)

        if not within_budget:
            # Apply aggressive optimization
            context_data = await self._optimize_context_aggressively(context_data, remaining_budget)
            total_tokens = self._calculate_total_tokens(context_data)

        # Cache the result
        self.context_cache.put(cache_key, context_data, total_tokens, self.current_phase.value)

        # Record usage
        load_time = time.time() - start_time
        self.token_manager.record_usage(self.current_phase, total_tokens, user_input)

        metrics = ContextMetrics(
            load_time=load_time,
            token_count=total_tokens,
            cache_hit=False,
            phase=self.current_phase.value,
            skills_loaded=len(context_data.get("skills", [])),
            docs_loaded=len(context_data.get("documents", [])),
            memory_usage=psutil.Process().memory_info().rss,
        )

        self._record_metrics(metrics)
        return context_data, metrics

    def _generate_cache_key(self, phase: Phase, user_input: str, context: Dict[str, Any]) -> str:
        """Generate unique cache key for context request"""
        key_data = {
            "phase": phase.value,
            "input_hash": hashlib.md5(user_input.encode(), usedforsecurity=False).hexdigest()[:16],
            "context_keys": sorted(context.keys()),
            "timestamp": datetime.now().strftime("%Y%m%d"),  # Daily cache
        }

        return hashlib.md5(json.dumps(key_data, sort_keys=True).encode(), usedforsecurity=False).hexdigest()

    async def _build_context(self, phase: Phase, phase_config: PhaseConfig, context: Dict[str, Any]) -> Dict[str, Any]:
        """Build optimized context for the current phase"""
        context_data: Dict[str, Any] = {
            "phase": phase.value,
            "skills": [],
            "documents": [],
            "metadata": {
                "created_at": datetime.now().isoformat(),
                "max_tokens": phase_config.max_tokens,
                "compression_enabled": phase_config.context_compression,
            },
        }

        # Filter and load skills
        skills = self.skill_filter.filter_skills(phase, phase_config.max_tokens // 2, context)
        for skill in skills:
            skill_content = await self._load_skill_content(skill)
            if skill_content:
                skills_list = context_data["skills"]
                if isinstance(skills_list, list):
                    skills_list.append(
                        {
                            "name": skill.name,
                            "content": skill_content,
                            "tokens": skill.tokens,
                            "categories": skill.categories,
                            "priority": skill.priority,
                        }
                    )

        # Load essential documents
        for doc_path in phase_config.essential_documents:
            doc_content = await self._load_document(doc_path, context)
            if doc_content:
                docs_list = context_data["documents"]
                if isinstance(docs_list, list):
                    docs_list.append(
                        {
                            "path": doc_path,
                            "content": doc_content["content"],
                            "tokens": doc_content["tokens"],
                            "type": doc_content["type"],
                        }
                    )

        return context_data

    async def _load_skill_content(self, skill: SkillInfo) -> Optional[str]:
        """Load content of a skill file"""
        try:
            with open(skill.path, "r", encoding="utf-8") as f:
                content = f.read()
            return content
        except Exception as e:
            logger.error(f"Error loading skill {skill.name}: {e}")
            return None

    async def _load_document(self, doc_path: str, context: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Load and process a document"""
        try:
            # Apply context variable substitution
            formatted_path = doc_path.format(
                spec_id=context.get("spec_id", "SPEC-XXX"),
                module=context.get("module", "unknown"),
                language=context.get("language", "python"),
            )

            if os.path.exists(formatted_path):
                with open(formatted_path, "r", encoding="utf-8") as f:
                    content = f.read()

                return {
                    "content": content,
                    "tokens": len(content) // 4,  # Rough token estimation
                    "type": self._detect_document_type(formatted_path),
                }
        except Exception as e:
            logger.error(f"Error loading document {doc_path}: {e}")
            return None

    def _detect_document_type(self, file_path: str) -> str:
        """Detect document type from file path"""
        if file_path.endswith(".md"):
            return "markdown"
        elif file_path.endswith(".py"):
            return "python"
        elif file_path.endswith(".json"):
            return "json"
        elif file_path.endswith(".yaml") or file_path.endswith(".yml"):
            return "yaml"
        else:
            return "text"

    async def _optimize_context_aggressively(self, context_data: Dict[str, Any], token_budget: int) -> Dict[str, Any]:
        """Apply aggressive optimization to fit token budget"""
        current_tokens = self._calculate_total_tokens(context_data)

        if current_tokens <= token_budget:
            return context_data

        # Strategy 1: Remove low-priority skills
        skills = context_data.get("skills", [])
        skills.sort(key=lambda s: s.get("priority", 3))  # Remove low priority first

        optimized_skills = []
        used_tokens = 0

        for skill in skills:
            if used_tokens + skill.get("tokens", 0) <= token_budget * 0.7:  # Reserve 30% for docs
                optimized_skills.append(skill)
                used_tokens += skill.get("tokens", 0)
            else:
                break

        context_data["skills"] = optimized_skills

        # Strategy 2: Compress documents if still over budget
        if self._calculate_total_tokens(context_data) > token_budget:
            documents = context_data.get("documents", [])
            for doc in documents:
                if doc.get("tokens", 0) > 1000:  # Only compress large documents
                    doc["content"] = self._compress_text(doc["content"])
                    doc["tokens"] = len(doc["content"]) // 4
                    doc["compressed"] = True

        return context_data

    def _compress_text(self, text: str) -> str:
        """Simple text compression by removing redundancy"""
        lines = text.split("\n")
        compressed_lines = []

        for line in lines:
            stripped = line.strip()
            # Skip empty lines and common comment patterns
            if stripped and not stripped.startswith("#") and not stripped.startswith("//") and len(stripped) > 10:
                compressed_lines.append(stripped)

        return "\n".join(compressed_lines)

    def _calculate_total_tokens(self, context_data: Dict[str, Any]) -> int:
        """Calculate total tokens in context data"""
        total_tokens = 0

        # Count skill tokens
        for skill in context_data.get("skills", []):
            total_tokens += skill.get("tokens", 0)

        # Count document tokens
        for doc in context_data.get("documents", []):
            total_tokens += doc.get("tokens", 0)

        # Add overhead (approximate)
        total_tokens += 1000  # Metadata and structure overhead

        return total_tokens

    def _record_metrics(self, metrics: ContextMetrics):
        """Record performance metrics"""
        self.metrics_history.append(metrics)

        # Keep only recent metrics (last 100)
        if len(self.metrics_history) > 100:
            self.metrics_history.pop(0)

        # Update performance stats
        self.performance_stats["total_loads"] += 1
        self.performance_stats["average_load_time"] = sum(m.load_time for m in self.metrics_history) / len(
            self.metrics_history
        )

        cache_hits = sum(1 for m in self.metrics_history if m.cache_hit)
        self.performance_stats["cache_hit_rate"] = cache_hits / len(self.metrics_history) * 100

        # Update efficiency score from token manager
        efficiency_metrics = self.token_manager.get_efficiency_metrics()
        self.performance_stats["efficiency_score"] = efficiency_metrics["efficiency_score"]

    def get_comprehensive_stats(self) -> Dict[str, Any]:
        """Get comprehensive system statistics"""
        return {
            "performance": self.performance_stats,
            "cache": self.context_cache.get_stats(),
            "token_efficiency": self.token_manager.get_efficiency_metrics(),
            "skill_filter": self.skill_filter.get_skill_stats(),
            "current_phase": self.current_phase.value,
            "phase_history": self.phase_detector.phase_history[-5:],  # Last 5 phase changes
            "metrics_count": len(self.metrics_history),
        }


# Global instance for easy import
jit_context_loader = JITContextLoader()


# Convenience functions
async def load_optimized_context(
    user_input: str,
    conversation_history: List[str] = None,
    context: Dict[str, Any] = None,
) -> Tuple[Dict[str, Any], ContextMetrics]:
    """Load optimized context using global JIT loader instance"""
    return await jit_context_loader.load_context(user_input, conversation_history, context)


def get_jit_stats() -> Dict[str, Any]:
    """Get comprehensive JIT system statistics"""
    return jit_context_loader.get_comprehensive_stats()


def clear_jit_cache(phase: Optional[str] = None):
    """Clear JIT cache (optionally for specific phase)"""
    if phase:
        jit_context_loader.context_cache.clear_phase(phase)
    else:
        jit_context_loader.context_cache.clear()


# Initial setup
def initialize_jit_system():
    """Initialize JIT Context Loading System"""
    logger.info("Initializing JIT Context Loading System...")

    stats = get_jit_stats()
    logger.info(f"JIT System initialized with {stats['skill_filter']['total_skills']} skills")
    logger.info(f"Cache configured: {stats['cache']['entries']} entries, {stats['cache']['memory_usage_mb']:.1f}MB")

    return True
