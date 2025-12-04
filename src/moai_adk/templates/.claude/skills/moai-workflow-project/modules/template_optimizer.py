"""
MoAI Menu Project - Template Optimizer Module

Advanced template analysis, optimization, and performance management system.
Integrates patterns from moai-project-template-optimizer skill with
intelligent analysis and automated optimization capabilities.
"""

import json
import re
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Tuple


class TemplateOptimizer:
    """Advanced template optimization and performance analysis system."""

    def __init__(self, project_root: str, config: Dict[str, Any]):
        self.project_root = Path(project_root)
        self.config = config
        self.templates_dir = self.project_root / ".claude/skills/moai-menu-project/templates"
        self.backups_dir = self.project_root / ".moai-backups"
        self.optimization_cache = {}
        self._ensure_directories()

    def _ensure_directories(self):
        """Ensure all necessary directories exist."""
        self.templates_dir.mkdir(parents=True, exist_ok=True)
        self.backups_dir.mkdir(parents=True, exist_ok=True)
        (self.templates_dir / "optimized").mkdir(exist_ok=True)
        (self.templates_dir / "benchmarks").mkdir(exist_ok=True)

    def analyze_project_templates(self) -> Dict[str, Any]:
        """
        Perform comprehensive analysis of project templates.

        Returns:
            Analysis results with performance metrics and optimization recommendations
        """

        analysis_result = {
            "timestamp": datetime.now().isoformat(),
            "project_root": str(self.project_root),
            "template_files": [],
            "performance_metrics": {},
            "optimization_opportunities": [],
            "backup_recommendations": [],
            "complexity_analysis": {},
            "resource_usage": {},
        }

        # Discover template files
        template_files = self._discover_template_files()
        analysis_result["template_files"] = template_files

        # Analyze each template file
        for template_file in template_files:
            file_analysis = self._analyze_template_file(template_file)
            analysis_result["performance_metrics"][template_file["path"]] = file_analysis

        # Identify optimization opportunities
        analysis_result["optimization_opportunities"] = self._identify_optimization_opportunities(
            analysis_result["performance_metrics"]
        )

        # Analyze overall project complexity
        analysis_result["complexity_analysis"] = self._analyze_project_complexity(template_files)

        # Calculate resource usage patterns
        analysis_result["resource_usage"] = self._calculate_resource_usage(template_files)

        # Generate backup recommendations
        analysis_result["backup_recommendations"] = self._generate_backup_recommendations(analysis_result)

        return analysis_result

    def _discover_template_files(self) -> List[Dict[str, Any]]:
        """Discover all template files in the project."""

        template_files = []

        # Common template file patterns
        template_patterns = [
            "**/*.template",
            "**/*.tmpl",
            "**/*.mustache",
            "**/*.hbs",
            "**/*.handlebars",
            "**/*.jinja",
            "**/*.jinja2",
            "**/*.erb",
            "**/*.liquid",
        ]

        # Search for template files
        for pattern in template_patterns:
            for file_path in self.project_root.glob(pattern):
                if file_path.is_file():
                    relative_path = file_path.relative_to(self.project_root)

                    # Calculate file metrics
                    file_stats = file_path.stat()

                    template_files.append(
                        {
                            "path": str(relative_path),
                            "absolute_path": str(file_path),
                            "size_bytes": file_stats.st_size,
                            "modified_time": datetime.fromtimestamp(file_stats.st_mtime).isoformat(),
                            "type": self._determine_template_type(file_path),
                            "language": self._determine_template_language(file_path),
                        }
                    )

        # Also check common template directories
        template_dirs = [
            "templates",
            "template",
            "views",
            "layouts",
            "_layouts",
            "_includes",
            "_templates",
        ]

        for template_dir in template_dirs:
            dir_path = self.project_root / template_dir
            if dir_path.exists() and dir_path.is_dir():
                for file_path in dir_path.rglob("*"):
                    if file_path.is_file() and file_path.suffix in [
                        ".md",
                        ".html",
                        ".txt",
                        ".yml",
                        ".yaml",
                        ".json",
                    ]:
                        # Check if file contains template markers
                        content = file_path.read_text(encoding="utf-8", errors="ignore")
                        if self._contains_template_markers(content):
                            relative_path = file_path.relative_to(self.project_root)
                            file_stats = file_path.stat()

                            template_files.append(
                                {
                                    "path": str(relative_path),
                                    "absolute_path": str(file_path),
                                    "size_bytes": file_stats.st_size,
                                    "modified_time": datetime.fromtimestamp(file_stats.st_mtime).isoformat(),
                                    "type": "content_template",
                                    "language": "text",
                                }
                            )

        return sorted(template_files, key=lambda x: x["path"])

    def _determine_template_type(self, file_path: Path) -> str:
        """Determine template type based on file extension and content."""

        extension_map = {
            ".mustache": "mustache",
            ".hbs": "handlebars",
            ".handlebars": "handlebars",
            ".jinja": "jinja",
            ".jinja2": "jinja2",
            ".erb": "erb",
            ".liquid": "liquid",
        }

        return extension_map.get(file_path.suffix.lower(), "unknown")

    def _determine_template_language(self, file_path: Path) -> str:
        """Determine the primary language used in template."""

        try:
            content = file_path.read_text(encoding="utf-8", errors="ignore")

            # Look for language-specific markers
            if any(marker in content for marker in ["{{", "{%", "{#"]):
                return "jinja"
            elif any(marker in content for marker in ["{{", "{{#", "{{/"]):
                return "handlebars"
            elif any(marker in content for marker in ["<%", "<%=", "<%-"]):
                return "erb"
            elif any(marker in content for marker in ["{{ ", "{% ", "{{-"]):
                return "liquid"
            else:
                return "text"

        except (UnicodeDecodeError, OSError):
            return "binary"

    def _contains_template_markers(self, content: str) -> bool:
        """Check if content contains template markers."""

        template_markers = [
            "{{",
            "}}",  # Generic template markers
            "{%",
            "%}",  # Jinja-like
            "{{#",
            "{{/",  # Handlebars-like
            "<%",
            "%>",  # ERB-like
            "{{-",
            "-}}",  # Liquid-like
            "{#",
            "#}",  # Comment markers
            "[[",
            "]]",  # Angular-like
            "${",
            "}",  # Variable interpolation
        ]

        return any(marker in content for marker in template_markers)

    def _analyze_template_file(self, template_file: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze individual template file for performance and complexity."""

        file_path = Path(template_file["absolute_path"])

        analysis = {
            "file_path": template_file["path"],
            "size_metrics": self._analyze_file_size(file_path),
            "complexity_metrics": self._analyze_template_complexity(file_path),
            "performance_metrics": self._estimate_template_performance(file_path),
            "optimization_potential": 0,
            "recommendations": [],
        }

        # Calculate overall optimization potential
        analysis["optimization_potential"] = self._calculate_optimization_potential(analysis)

        # Generate specific recommendations
        analysis["recommendations"] = self._generate_file_recommendations(analysis)

        return analysis

    def _analyze_file_size(self, file_path: Path) -> Dict[str, Any]:
        """Analyze file size and content metrics."""

        try:
            content = file_path.read_text(encoding="utf-8")

            lines = content.splitlines()

            return {
                "total_size_bytes": len(content.encode("utf-8")),
                "character_count": len(content),
                "line_count": len(lines),
                "blank_lines": sum(1 for line in lines if not line.strip()),
                "average_line_length": (sum(len(line) for line in lines) / len(lines) if lines else 0),
                "max_line_length": max(len(line) for line in lines) if lines else 0,
            }

        except (UnicodeDecodeError, OSError):
            return {
                "total_size_bytes": file_path.stat().st_size,
                "character_count": 0,
                "line_count": 0,
                "blank_lines": 0,
                "average_line_length": 0,
                "max_line_length": 0,
            }

    def _analyze_template_complexity(self, file_path: Path) -> Dict[str, Any]:
        """Analyze template complexity metrics."""

        try:
            content = file_path.read_text(encoding="utf-8", errors="ignore")

            complexity = {
                "variable_count": 0,
                "loop_count": 0,
                "conditional_count": 0,
                "include_count": 0,
                "nested_level_max": 0,
                "template_depth": 0,
            }

            # Count different template constructs
            complexity["variable_count"] = len(re.findall(r"\{\{[^}]+\}\}", content))
            complexity["loop_count"] = len(re.findall(r"\{%\s*for\s+.+\s+in\s+.+%\}", content))
            complexity["conditional_count"] = len(re.findall(r"\{%\s*if\s+.+%\}", content))
            complexity["include_count"] = len(re.findall(r"\{%\s*(include|extend|import)\s+.+%\}", content))

            # Calculate nesting level
            current_level = 0
            max_level = 0

            for char in content:
                if char == "{":
                    if content[content.index(char) : content.index(char) + 2] in [
                        "{{",
                        "{%",
                    ]:
                        current_level += 1
                        max_level = max(max_level, current_level)
                elif char == "}":
                    if current_level > 0:
                        current_level -= 1

            complexity["nested_level_max"] = max_level
            complexity["template_depth"] = self._calculate_template_depth(content)

        except (UnicodeDecodeError, OSError):
            complexity = {
                "variable_count": 0,
                "loop_count": 0,
                "conditional_count": 0,
                "include_count": 0,
                "nested_level_max": 0,
                "template_depth": 0,
            }

        return complexity

    def _calculate_template_depth(self, content: str) -> int:
        """Calculate template inheritance depth."""

        # Look for extends, include, import statements
        depth_patterns = [
            r'\{%\s*extends\s+["\']([^"\']+)["\']\s*%\}',
            r'\{%\s*include\s+["\']([^"\']+)["\']\s*%\}',
            r'\{%\s*import\s+["\']([^"\']+)["\']\s*%\}',
        ]

        depth = 0
        for pattern in depth_patterns:
            matches = re.findall(pattern, content)
            depth += len(matches)

        return min(depth, 10)  # Cap at reasonable maximum

    def _estimate_template_performance(self, file_path: Path) -> Dict[str, Any]:
        """Estimate template rendering performance."""

        size_analysis = self._analyze_file_size(file_path)
        complexity_analysis = self._analyze_template_complexity(file_path)

        performance = {
            "estimated_render_time_ms": 0,
            "memory_usage_estimate_kb": 0,
            "cpu_intensity": "low",
            "cache_friendly": True,
            "render_complexity_score": 0,
        }

        # Basic performance estimation based on size and complexity
        base_time = 1.0  # Base render time in ms

        # Add time for complexity
        complexity_time = (
            complexity_analysis["variable_count"] * 0.1
            + complexity_analysis["loop_count"] * 2.0
            + complexity_analysis["conditional_count"] * 0.5
            + complexity_analysis["include_count"] * 1.0
        )

        # Add time for file size
        size_time = size_analysis["total_size_bytes"] / 10000  # 10KB per ms

        performance["estimated_render_time_ms"] = base_time + complexity_time + size_time

        # Estimate memory usage
        performance["memory_usage_estimate_kb"] = size_analysis["total_size_bytes"] / 1024

        # Determine CPU intensity
        if performance["estimated_render_time_ms"] > 100:
            performance["cpu_intensity"] = "high"
        elif performance["estimated_render_time_ms"] > 50:
            performance["cpu_intensity"] = "medium"
        else:
            performance["cpu_intensity"] = "low"

        # Check cache friendliness
        performance["cache_friendly"] = (
            complexity_analysis["variable_count"] < 50
            and complexity_analysis["loop_count"] < 10
            and performance["estimated_render_time_ms"] < 100
        )

        # Calculate render complexity score
        performance["render_complexity_score"] = (
            complexity_analysis["variable_count"] * 1
            + complexity_analysis["loop_count"] * 3
            + complexity_analysis["conditional_count"] * 2
            + complexity_analysis["nested_level_max"] * 4
        )

        return performance

    def _calculate_optimization_potential(self, file_analysis: Dict[str, Any]) -> float:
        """Calculate optimization potential score (0-100)."""

        potential = 0.0

        # Size-based potential
        size_metrics = file_analysis["size_metrics"]
        if size_metrics["total_size_bytes"] > 50000:  # 50KB
            potential += 20
        elif size_metrics["total_size_bytes"] > 20000:  # 20KB
            potential += 10

        # Complexity-based potential
        complexity_metrics = file_analysis["complexity_metrics"]
        if complexity_metrics["nested_level_max"] > 5:
            potential += 15
        elif complexity_metrics["nested_level_max"] > 3:
            potential += 8

        if complexity_metrics["variable_count"] > 100:
            potential += 10
        elif complexity_metrics["variable_count"] > 50:
            potential += 5

        # Performance-based potential
        performance_metrics = file_analysis["performance_metrics"]
        if performance_metrics["estimated_render_time_ms"] > 100:
            potential += 20
        elif performance_metrics["estimated_render_time_ms"] > 50:
            potential += 10

        if not performance_metrics["cache_friendly"]:
            potential += 15

        return min(potential, 100)  # Cap at 100

    def _generate_file_recommendations(self, file_analysis: Dict[str, Any]) -> List[str]:
        """Generate optimization recommendations for a file."""

        recommendations = []

        size_metrics = file_analysis["size_metrics"]
        complexity_metrics = file_analysis["complexity_metrics"]
        performance_metrics = file_analysis["performance_metrics"]

        # Size recommendations
        if size_metrics["total_size_bytes"] > 50000:
            recommendations.append("Consider splitting large template into smaller components")

        if size_metrics["max_line_length"] > 200:
            recommendations.append("Break long lines for better readability")

        # Complexity recommendations
        if complexity_metrics["nested_level_max"] > 5:
            recommendations.append("Reduce nesting level by extracting sub-templates")

        if complexity_metrics["variable_count"] > 100:
            recommendations.append("Consider consolidating related variables")

        if complexity_metrics["loop_count"] > 10:
            recommendations.append("Optimize loops and consider pagination for large datasets")

        # Performance recommendations
        if performance_metrics["estimated_render_time_ms"] > 100:
            recommendations.append("Implement template caching for better performance")

        if not performance_metrics["cache_friendly"]:
            recommendations.append("Restructure template for better cache efficiency")

        if performance_metrics["cpu_intensity"] == "high":
            recommendations.append("Consider moving complex logic to template filters or functions")

        return recommendations

    def _identify_optimization_opportunities(self, performance_metrics: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Identify global optimization opportunities across all templates."""

        opportunities = []

        # Analyze patterns across all files
        total_files = len(performance_metrics)
        if total_files == 0:
            return opportunities

        # Calculate aggregate metrics
        total_size = sum(metrics["size_metrics"]["total_size_bytes"] for metrics in performance_metrics.values())
        avg_render_time = (
            sum(metrics["performance_metrics"]["estimated_render_time_ms"] for metrics in performance_metrics.values())
            / total_files
        )
        high_complexity_files = sum(
            1 for metrics in performance_metrics.values() if metrics["complexity_metrics"]["nested_level_max"] > 3
        )

        # Size optimization opportunities
        if total_size > 100000:  # 100KB total
            opportunities.append(
                {
                    "type": "size_optimization",
                    "priority": "high",
                    "description": f"Large total template size ({total_size / 1024:.1f}KB)",
                    "recommendation": "Consider consolidating shared templates and removing unused content",
                    "estimated_impact": "15-25% reduction in load time",
                }
            )

        # Performance optimization opportunities
        if avg_render_time > 50:
            opportunities.append(
                {
                    "type": "performance_optimization",
                    "priority": "medium",
                    "description": f"Average render time is high ({avg_render_time:.1f}ms)",
                    "recommendation": "Implement template caching and optimize complex templates",
                    "estimated_impact": "20-40% improvement in render time",
                }
            )

        # Complexity optimization opportunities
        if high_complexity_files > total_files * 0.3:  # More than 30% files are complex
            opportunities.append(
                {
                    "type": "complexity_optimization",
                    "priority": "medium",
                    "description": f"High complexity in {high_complexity_files} of {total_files} templates",
                    "recommendation": "Extract complex logic into template filters and reduce nesting",
                    "estimated_impact": "Improved maintainability and 10-20% performance gain",
                }
            )

        # Template reuse opportunities
        template_types = {}
        for metrics in performance_metrics.values():
            file_path = Path(metrics["file_path"])
            template_type = file_path.suffix
            template_types[template_type] = template_types.get(template_type, 0) + 1

        common_types = [t for t, count in template_types.items() if count > 3]
        if common_types:
            opportunities.append(
                {
                    "type": "template_reuse",
                    "priority": "low",
                    "description": f"Multiple templates of type: {', '.join(common_types)}",
                    "recommendation": "Create base templates and use template inheritance",
                    "estimated_impact": "Reduced duplication and easier maintenance",
                }
            )

        return sorted(
            opportunities,
            key=lambda x: {"high": 3, "medium": 2, "low": 1}.get(x["priority"], 0),
            reverse=True,
        )

    def _analyze_project_complexity(self, template_files: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Analyze overall project template complexity."""

        complexity_analysis = {
            "total_files": len(template_files),
            "total_size_bytes": sum(f["size_bytes"] for f in template_files),
            "file_types": {},
            "template_languages": {},
            "average_file_size": 0,
            "complexity_distribution": {"low": 0, "medium": 0, "high": 0},
        }

        if template_files:
            complexity_analysis["average_file_size"] = complexity_analysis["total_size_bytes"] / len(template_files)

        # Analyze file types
        for template_file in template_files:
            file_type = template_file["type"]
            complexity_analysis["file_types"][file_type] = complexity_analysis["file_types"].get(file_type, 0) + 1

            lang = template_file["language"]
            complexity_analysis["template_languages"][lang] = complexity_analysis["template_languages"].get(lang, 0) + 1

        return complexity_analysis

    def _calculate_resource_usage(self, template_files: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Calculate resource usage patterns and predictions."""

        resource_usage = {
            "disk_usage": {
                "total_bytes": sum(f["size_bytes"] for f in template_files),
                "average_file_size": 0,
                "largest_file": None,
            },
            "memory_estimates": {"peak_usage_kb": 0, "average_usage_kb": 0},
            "performance_predictions": {
                "concurrent_render_support": 0,
                "recommended_cache_size_mb": 0,
            },
        }

        if template_files:
            resource_usage["disk_usage"]["average_file_size"] = resource_usage["disk_usage"]["total_bytes"] / len(
                template_files
            )
            resource_usage["disk_usage"]["largest_file"] = max(template_files, key=lambda f: f["size_bytes"])["path"]

            # Estimate memory usage (rough approximation: 3x file size for processing)
            total_size_kb = resource_usage["disk_usage"]["total_bytes"] / 1024
            resource_usage["memory_estimates"]["average_usage_kb"] = total_size_kb * 3
            resource_usage["memory_estimates"]["peak_usage_kb"] = total_size_kb * 5

            # Performance predictions
            avg_file_size_kb = resource_usage["disk_usage"]["average_file_size"] / 1024
            resource_usage["performance_predictions"]["concurrent_render_support"] = int(
                1000 / (avg_file_size_kb + 10)
            )  # Rough estimate
            resource_usage["performance_predictions"]["recommended_cache_size_mb"] = max(1, int(total_size_kb / 1024))

        return resource_usage

    def _generate_backup_recommendations(self, analysis_result: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Generate backup recommendations based on analysis."""

        recommendations = []

        template_count = len(analysis_result["template_files"])
        total_size = analysis_result["complexity_analysis"]["total_size_bytes"]
        optimization_opportunities = analysis_result["optimization_opportunities"]

        # Create backup if optimization opportunities exist
        if optimization_opportunities:
            high_priority_ops = [op for op in optimization_opportunities if op["priority"] == "high"]

            if high_priority_ops:
                recommendations.append(
                    {
                        "type": "pre_optimization_backup",
                        "priority": "high",
                        "description": "Create backup before applying optimizations",
                        "reason": "High-impact optimizations detected",
                        "backup_size_estimate": total_size,
                        "recommended_backup_name": f"backup-{datetime.now().strftime('%Y-%m-%d')}-pre-optimization",
                    }
                )

        # Regular backup recommendations
        if template_count > 10:
            recommendations.append(
                {
                    "type": "regular_backup",
                    "priority": "medium",
                    "description": "Regular backup recommended for large template collections",
                    "reason": f"{template_count} template files detected",
                    "backup_frequency": "weekly",
                }
            )

        # Version control recommendations
        if not (self.project_root / ".git").exists():
            recommendations.append(
                {
                    "type": "version_control",
                    "priority": "low",
                    "description": "Consider using Git for template version control",
                    "reason": "No version control system detected",
                    "suggested_action": "Initialize Git repository",
                }
            )

        return recommendations

    def create_optimized_templates(self, optimization_options: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Create optimized versions of templates.

        Args:
            optimization_options: Options for optimization process

        Returns:
            Optimization results and created files
        """

        if optimization_options is None:
            optimization_options = {
                "backup_first": True,
                "apply_size_optimizations": True,
                "apply_performance_optimizations": True,
                "apply_complexity_optimizations": True,
                "preserve_functionality": True,
            }

        optimization_result = {
            "timestamp": datetime.now().isoformat(),
            "options_used": optimization_options,
            "backup_created": False,
            "optimized_files": [],
            "optimizations_applied": [],
            "size_reduction": 0,
            "performance_improvement": 0,
            "errors": [],
        }

        # Create backup if requested
        if optimization_options.get("backup_first", True):
            backup_result = self._create_template_backup("pre-optimization")
            optimization_result["backup_created"] = backup_result["success"]
            optimization_result["backup_info"] = backup_result

        # Analyze current templates
        analysis = self.analyze_project_templates()

        # Apply optimizations
        for template_file in analysis["template_files"]:
            file_optimization = self._optimize_template_file(template_file, optimization_options)

            if file_optimization["success"]:
                optimization_result["optimized_files"].append(file_optimization)
                optimization_result["optimizations_applied"].extend(file_optimization["applied_optimizations"])
            else:
                optimization_result["errors"].append(
                    {
                        "file": template_file["path"],
                        "error": file_optimization.get("error", "Unknown error"),
                    }
                )

        # Calculate overall improvements
        if optimization_result["optimized_files"]:
            original_total_size = sum(f["size_bytes"] for f in analysis["template_files"])
            optimized_total_size = sum(f["optimized_size_bytes"] for f in optimization_result["optimized_files"])

            if original_total_size > 0:
                optimization_result["size_reduction"] = (
                    (original_total_size - optimized_total_size) / original_total_size * 100
                )

        return optimization_result

    def _create_template_backup(self, backup_name: str = None) -> Dict[str, Any]:
        """Create backup of all template files."""

        if backup_name is None:
            backup_name = f"backup-{datetime.now().strftime('%Y-%m-%d-%H-%M-%S')}"

        backup_path = self.backups_dir / backup_name
        backup_path.mkdir(parents=True, exist_ok=True)

        backup_result = {
            "backup_name": backup_name,
            "backup_path": str(backup_path),
            "success": False,
            "files_backed_up": 0,
            "total_size_bytes": 0,
            "created_at": datetime.now().isoformat(),
        }

        try:
            # Discover and backup template files
            template_files = self._discover_template_files()

            for template_file in template_files:
                source_path = Path(template_file["absolute_path"])
                relative_path = Path(template_file["path"])
                backup_file_path = backup_path / relative_path

                # Ensure backup directory exists
                backup_file_path.parent.mkdir(parents=True, exist_ok=True)

                # Copy file
                shutil.copy2(source_path, backup_file_path)
                backup_result["files_backed_up"] += 1
                backup_result["total_size_bytes"] += template_file["size_bytes"]

            # Create backup metadata
            metadata = {
                "backup_name": backup_name,
                "created_at": backup_result["created_at"],
                "files_backed_up": backup_result["files_backed_up"],
                "total_size_bytes": backup_result["total_size_bytes"],
                "project_root": str(self.project_root),
                "template_files": template_files,
            }

            metadata_path = backup_path / "backup-metadata.json"
            metadata_path.write_text(json.dumps(metadata, indent=2, ensure_ascii=False), encoding="utf-8")

            backup_result["success"] = True

        except Exception as e:
            backup_result["error"] = str(e)

        return backup_result

    def _optimize_template_file(self, template_file: Dict[str, Any], options: Dict[str, Any]) -> Dict[str, Any]:
        """Optimize individual template file."""

        file_path = Path(template_file["absolute_path"])
        optimization_result = {
            "file_path": template_file["path"],
            "success": False,
            "original_size_bytes": template_file["size_bytes"],
            "optimized_size_bytes": 0,
            "applied_optimizations": [],
            "error": None,
        }

        try:
            # Read original content
            original_content = file_path.read_text(encoding="utf-8")
            optimized_content = original_content
            applied_optimizations = []

            # Apply size optimizations
            if options.get("apply_size_optimizations", True):
                size_optimized, size_opts = self._apply_size_optimizations(optimized_content)
                if size_optimized != optimized_content:
                    optimized_content = size_optimized
                    applied_optimizations.extend(size_opts)

            # Apply performance optimizations
            if options.get("apply_performance_optimizations", True):
                perf_optimized, perf_opts = self._apply_performance_optimizations(optimized_content)
                if perf_optimized != optimized_content:
                    optimized_content = perf_optimized
                    applied_optimizations.extend(perf_opts)

            # Apply complexity optimizations
            if options.get("apply_complexity_optimizations", True):
                complexity_optimized, complexity_opts = self._apply_complexity_optimizations(optimized_content)
                if complexity_optimized != optimized_content:
                    optimized_content = complexity_optimized
                    applied_optimizations.extend(complexity_opts)

            # Write optimized version
            if optimized_content != original_content:
                # Create optimized version
                optimized_dir = self.templates_dir / "optimized"
                relative_path = file_path.relative_to(self.project_root)
                optimized_file_path = optimized_dir / relative_path

                optimized_file_path.parent.mkdir(parents=True, exist_ok=True)
                optimized_file_path.write_text(optimized_content, encoding="utf-8")

                optimization_result.update(
                    {
                        "success": True,
                        "optimized_size_bytes": len(optimized_content.encode("utf-8")),
                        "applied_optimizations": applied_optimizations,
                        "optimized_file_path": str(optimized_file_path),
                    }
                )
            else:
                optimization_result["success"] = True
                optimization_result["message"] = "No optimizations needed"

        except Exception as e:
            optimization_result["error"] = str(e)

        return optimization_result

    def _apply_size_optimizations(self, content: str) -> Tuple[str, List[str]]:
        """Apply size reduction optimizations."""

        optimized_content = content
        applied_optimizations = []

        # Remove excessive whitespace
        original_lines = content.splitlines()
        optimized_lines = []

        for line in original_lines:
            # Remove leading/trailing whitespace but preserve template indentation
            stripped_line = line.rstrip()
            if stripped_line or line.strip() == "":
                optimized_lines.append(stripped_line)

        optimized_content = "\n".join(optimized_lines)

        if len(optimized_content) < len(content) * 0.95:  # At least 5% reduction
            applied_optimizations.append("whitespace_optimization")

        # Remove redundant template markers
        # (This would be more sophisticated in practice)
        optimized_content = re.sub(r"\{\{\s+\{\{", "{{", optimized_content)
        optimized_content = re.sub(r"\}\}\s+\}\}", "}}", optimized_content)

        if optimized_content != content:
            applied_optimizations.append("template_marker_optimization")

        return optimized_content, applied_optimizations

    def _apply_performance_optimizations(self, content: str) -> Tuple[str, List[str]]:
        """Apply performance optimizations."""

        optimized_content = content
        applied_optimizations = []

        # Optimize loop structures
        # This is a simplified example - real optimization would be more sophisticated
        original_content = optimized_content

        # Replace complex inline conditions with template filters
        optimized_content = re.sub(
            r"\{\{\s*if\s+(.+?)\s*%\}\s*\{\{\s*(.+?)\s*\}\}\s*\{\%\s*else\s*%\}\s*\{\{\s*(.+?)\s*\}\}\s*\{\%\s*endif\s*%\}",
            r'{{ \1 | default("\3") if \2 else "\3" }}',
            optimized_content,
        )

        if optimized_content != original_content:
            applied_optimizations.append("conditional_optimization")

        # Cache expensive operations
        optimized_content = re.sub(r"\{\{\s*(.+?)\|length\s*\}\}", r"{{ \1 | length }}", optimized_content)

        return optimized_content, applied_optimizations

    def _apply_complexity_optimizations(self, content: str) -> Tuple[str, List[str]]:
        """Apply complexity reduction optimizations."""

        optimized_content = content
        applied_optimizations = []

        # Extract complex nested structures into separate templates
        # This is a simplified detection - real implementation would be more advanced

        # Find deeply nested structures (simplified example)
        nesting_level = 0
        max_nesting = 0
        lines = content.splitlines()

        for line in lines:
            open_blocks = line.count("{%") + line.count("{{")
            close_blocks = line.count("%}") + line.count("}}")
            nesting_level += open_blocks - close_blocks
            max_nesting = max(max_nesting, nesting_level)

        if max_nesting > 5:
            applied_optimizations.append("complexity_reduction_needed")
            # In practice, this would extract nested content to separate templates

        return optimized_content, applied_optimizations

    def benchmark_template_performance(self, template_paths: List[str] = None) -> Dict[str, Any]:
        """
        Benchmark template rendering performance.

        Args:
            template_paths: Specific templates to benchmark (all if None)

        Returns:
            Performance benchmark results
        """

        benchmark_result = {
            "timestamp": datetime.now().isoformat(),
            "benchmark_files": template_paths or [],
            "performance_metrics": {},
            "summary": {
                "total_files_tested": 0,
                "average_render_time_ms": 0,
                "fastest_file": None,
                "slowest_file": None,
            },
        }

        # Get files to benchmark
        if template_paths is None:
            template_files = self._discover_template_files()
            template_paths = [f["path"] for f in template_files]

        benchmark_result["benchmark_files"] = template_paths
        render_times = []

        for template_path in template_paths:
            file_path = self.project_root / template_path
            if file_path.exists():
                performance_metrics = self._estimate_template_performance(file_path)

                benchmark_result["performance_metrics"][template_path] = {
                    "estimated_render_time_ms": performance_metrics["estimated_render_time_ms"],
                    "memory_usage_kb": performance_metrics["memory_usage_estimate_kb"],
                    "cpu_intensity": performance_metrics["cpu_intensity"],
                    "cache_friendly": performance_metrics["cache_friendly"],
                    "complexity_score": performance_metrics["render_complexity_score"],
                }

                render_times.append(performance_metrics["estimated_render_time_ms"])

        # Calculate summary
        if render_times:
            benchmark_result["summary"]["total_files_tested"] = len(render_times)
            benchmark_result["summary"]["average_render_time_ms"] = sum(render_times) / len(render_times)

            # Find fastest and slowest files
            min_time_idx = render_times.index(min(render_times))
            max_time_idx = render_times.index(max(render_times))

            benchmark_result["summary"]["fastest_file"] = {
                "path": template_paths[min_time_idx],
                "render_time_ms": min(render_times),
            }

            benchmark_result["summary"]["slowest_file"] = {
                "path": template_paths[max_time_idx],
                "render_time_ms": max(render_times),
            }

        # Save benchmark results
        benchmark_dir = self.templates_dir / "benchmarks"
        benchmark_dir.mkdir(exist_ok=True)

        benchmark_file = benchmark_dir / f"benchmark-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json"
        benchmark_file.write_text(json.dumps(benchmark_result, indent=2, ensure_ascii=False), encoding="utf-8")

        return benchmark_result
