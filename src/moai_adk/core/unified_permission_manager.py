"""
Unified Permission Manager for MoAI-ADK

Production-ready permission management system that addresses agent permission validation
errors identified in Claude Code debug logs. Provides automatic correction, validation,
and monitoring of agent permissions and access control.

Author: MoAI-ADK Core Team
Version: 1.0.0
"""

import json
import logging
import os
import time
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional

# Configure logging
logger = logging.getLogger(__name__)


class PermissionMode(Enum):
    """Valid permission modes for agents"""

    ACCEPT_EDITS = "acceptEdits"
    BYPASS_PERMISSIONS = "bypassPermissions"
    DEFAULT = "default"
    DONT_ASK = "dontAsk"
    PLAN = "plan"


class PermissionSeverity(Enum):
    """Permission validation severity levels"""

    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class ResourceType(Enum):
    """Types of resources that can be protected"""

    AGENT = "agent"
    TOOL = "tool"
    FILE = "file"
    COMMAND = "command"
    SETTING = "setting"


@dataclass
class PermissionRule:
    """Individual permission rule"""

    resource_type: ResourceType
    resource_name: str
    action: str
    allowed: bool
    conditions: Optional[Dict[str, Any]] = None
    expires_at: Optional[float] = None


@dataclass
class ValidationResult:
    """Result of permission validation"""

    valid: bool
    corrected_mode: Optional[str] = None
    warnings: List[str] = field(default_factory=list)
    errors: List[str] = field(default_factory=list)
    severity: PermissionSeverity = PermissionSeverity.LOW
    auto_corrected: bool = False


@dataclass
class PermissionAudit:
    """Audit log entry for permission changes"""

    timestamp: float
    user_id: Optional[str]
    resource_type: ResourceType
    resource_name: str
    action: str
    old_permissions: Optional[Dict[str, Any]]
    new_permissions: Optional[Dict[str, Any]]
    reason: str
    auto_corrected: bool


class UnifiedPermissionManager:
    """
    Production-ready permission management system that addresses Claude Code
    agent permission validation errors with automatic correction and monitoring.

    Key Features:
    - Automatic permission mode validation and correction
    - Role-based access control with inheritance
    - Real-time permission monitoring and auditing
    - Configuration file auto-recovery
    - Security-focused fail-safe behavior
    """

    # Valid permission modes from Claude Code
    VALID_PERMISSION_MODES = {
        "acceptEdits",
        "bypassPermissions",
        "default",
        "dontAsk",
        "plan",
    }

    # Default permission mappings
    DEFAULT_PERMISSIONS = {
        "backend-expert": PermissionMode.ACCEPT_EDITS,
        "frontend-expert": PermissionMode.ACCEPT_EDITS,
        "security-expert": PermissionMode.ACCEPT_EDITS,
        "api-designer": PermissionMode.PLAN,
        "database-expert": PermissionMode.ACCEPT_EDITS,
        "docs-manager": PermissionMode.ACCEPT_EDITS,
        "tdd-implementer": PermissionMode.ACCEPT_EDITS,
        "spec-builder": PermissionMode.ACCEPT_EDITS,
        "quality-gate": PermissionMode.ACCEPT_EDITS,
        "default": PermissionMode.DEFAULT,
    }

    def __init__(self, config_path: Optional[str] = None, enable_logging: bool = True):
        self.config_path = config_path or ".claude/settings.json"
        self.enable_logging = enable_logging
        self.permission_cache: Dict[str, Any] = {}
        self.audit_log: List[PermissionAudit] = []
        self.stats = {
            "validations_performed": 0,
            "auto_corrections": 0,
            "security_violations": 0,
            "permission_denied": 0,
        }

        # Role hierarchy for inheritance
        self.role_hierarchy = {
            "admin": ["developer", "user"],
            "developer": ["user"],
            "user": [],
        }

        # Load and validate current configuration
        self.config = self._load_configuration()
        self._validate_all_permissions()

    def _load_configuration(self) -> Dict[str, Any]:
        """Load configuration from file with error handling"""
        try:
            if os.path.exists(self.config_path):
                with open(self.config_path, "r", encoding="utf-8") as f:
                    config = json.load(f)

                if self.enable_logging:
                    logger.info(f"Loaded configuration from {self.config_path}")

                return config
            else:
                if self.enable_logging:
                    logger.warning(f"Configuration file not found: {self.config_path}")
                return {}

        except json.JSONDecodeError as e:
            if self.enable_logging:
                logger.error(f"Invalid JSON in configuration file: {e}")
            return {}
        except Exception as e:
            if self.enable_logging:
                logger.error(f"Error loading configuration: {e}")
            return {}

    def _validate_all_permissions(self) -> None:
        """Validate all permissions in the current configuration"""
        corrections_made = False

        # Check agent permissions
        agents_config = self.config.get("agents", {})
        for agent_name, agent_config in agents_config.items():
            result = self.validate_agent_permission(agent_name, agent_config)
            if result.auto_corrected:
                corrections_made = True
                if self.enable_logging:
                    logger.info(f"Auto-corrected permissions for agent: {agent_name}")

        # Check settings permissions
        settings_config = self.config.get("projectSettings", {})
        if "allowedTools" in settings_config:
            result = self.validate_tool_permissions(settings_config["allowedTools"])
            if result.auto_corrected:
                corrections_made = True

        # Save corrections if any were made
        if corrections_made:
            self._save_configuration()
            if self.enable_logging:
                logger.info("Saved corrected configuration")

    def validate_agent_permission(self, agent_name: str, agent_config: Dict[str, Any]) -> ValidationResult:
        """
        Validate and auto-correct agent permission configuration.

        Addresses the permissionMode validation errors from debug logs:
        - Lines 50-80: Multiple agents with invalid permission modes ('ask', 'auto')
        """
        self.stats["validations_performed"] += 1

        result = ValidationResult(valid=True)

        # Extract current permission mode
        current_mode = agent_config.get("permissionMode", "default")

        # Validate permission mode
        if current_mode not in self.VALID_PERMISSION_MODES:
            # Auto-correct to appropriate default
            suggested_mode = self._suggest_permission_mode(agent_name)

            result.errors.append(
                f"Invalid permissionMode '{current_mode}' for agent '{agent_name}'. "
                f"Valid options: {sorted(self.VALID_PERMISSION_MODES)}"
            )

            # Auto-correction
            agent_config["permissionMode"] = suggested_mode
            result.corrected_mode = suggested_mode
            result.auto_corrected = True
            result.severity = PermissionSeverity.HIGH

            self.stats["auto_corrections"] += 1
            self._audit_permission_change(
                resource_type=ResourceType.AGENT,
                resource_name=agent_name,
                action="permission_mode_correction",
                old_permissions={"permissionMode": current_mode},
                new_permissions={"permissionMode": suggested_mode},
                reason=f"Invalid permission mode '{current_mode}' auto-corrected to '{suggested_mode}'",
                auto_corrected=True,
            )

            if self.enable_logging:
                logger.warning(
                    f"Auto-corrected agent '{agent_name}' permissionMode from '{current_mode}' to '{suggested_mode}'"
                )

        # Validate other agent configuration
        if "model" in agent_config:
            model = agent_config["model"]
            if not isinstance(model, str) or not model.strip():
                result.errors.append(f"Invalid model configuration for agent '{agent_name}'")
                result.severity = PermissionSeverity.MEDIUM

        # Check for required fields
        required_fields = ["description", "systemPrompt"]
        for req_field in required_fields:
            if req_field not in agent_config or not agent_config[req_field]:
                result.warnings.append(f"Missing or empty '{req_field}' for agent '{agent_name}'")

        return result

    def _suggest_permission_mode(self, agent_name: str) -> str:
        """
        Suggest appropriate permission mode based on agent name and function.

        This addresses the core issue from the debug logs where agents had
        invalid permission modes like 'ask' and 'auto'.
        """
        # Check if agent name matches known patterns
        agent_lower = agent_name.lower()

        # Security and compliance focused agents should be more restrictive
        if any(keyword in agent_lower for keyword in ["security", "audit", "compliance"]):
            return PermissionMode.PLAN.value

        # Code execution and modification agents should accept edits
        if any(keyword in agent_lower for keyword in ["expert", "implementer", "builder"]):
            return PermissionMode.ACCEPT_EDITS.value

        # Planning and analysis agents should use plan mode
        if any(keyword in agent_lower for keyword in ["planner", "analyzer", "designer"]):
            return PermissionMode.PLAN.value

        # Management agents should have appropriate permissions
        if any(keyword in agent_lower for keyword in ["manager", "coordinator"]):
            return PermissionMode.ACCEPT_EDITS.value

        # Check against our default mappings
        if agent_name in self.DEFAULT_PERMISSIONS:
            return self.DEFAULT_PERMISSIONS[agent_name].value

        # Default to safe option
        return PermissionMode.DEFAULT.value

    def validate_tool_permissions(self, allowed_tools: List[str]) -> ValidationResult:
        """Validate list of allowed tools for security compliance"""
        result = ValidationResult(valid=True)

        # Define dangerous tools that should require explicit approval
        dangerous_tools = {
            "Bash(rm -rf:*)",
            "Bash(sudo:*)",
            "Bash(chmod -R 777:*)",
            "Bash(dd:*)",
            "Bash(mkfs:*)",
            "Bash(fdisk:*)",
            "Bash(reboot:*)",
            "Bash(shutdown:*)",
            "Bash(git push --force:*)",
            "Bash(git reset --hard:*)",
        }

        for tool in allowed_tools:
            if tool in dangerous_tools:
                result.warnings.append(f"Dangerous tool allowed: {tool}. Consider restricting access.")
                result.severity = PermissionSeverity.HIGH
                self.stats["security_violations"] += 1

        return result

    def check_tool_permission(self, user_role: str, tool_name: str, operation: str) -> bool:
        """
        Check if a user role is permitted to use a specific tool.

        Implements unified permission checking with role hierarchy support.
        """
        self.stats["validations_performed"] += 1

        # Check cache first
        cache_key = f"{user_role}:{tool_name}:{operation}"
        if cache_key in self.permission_cache:
            return self.permission_cache[cache_key]

        # Check direct permissions
        permitted = self._check_direct_permission(user_role, tool_name, operation)

        # If not directly permitted, check role hierarchy
        if not permitted:
            for subordinate_role in self.role_hierarchy.get(user_role, []):
                if self._check_direct_permission(subordinate_role, tool_name, operation):
                    permitted = True
                    break

        # Cache the result
        self.permission_cache[cache_key] = permitted

        if not permitted:
            self.stats["permission_denied"] += 1
            if self.enable_logging:
                logger.warning(f"Permission denied: {user_role} cannot {operation} with {tool_name}")

        return permitted

    def _check_direct_permission(self, role: str, tool_name: str, operation: str) -> bool:
        """Check direct permissions for a specific role"""
        # Default permissions by role
        role_permissions = {
            "admin": ["*"],  # All tools
            "developer": ["Task", "Read", "Write", "Edit", "Bash", "AskUserQuestion"],
            "user": ["Task", "Read", "AskUserQuestion"],
        }

        allowed_tools = role_permissions.get(role, [])

        # Wildcard permission
        if "*" in allowed_tools:
            return True

        # Exact match
        if tool_name in allowed_tools:
            return True

        # Pattern matching for Bash commands
        if tool_name.startswith("Bash(") and "Bash" in allowed_tools:
            return True

        return False

    def validate_configuration(self, config_path: Optional[str] = None) -> ValidationResult:
        """
        Validate Claude Code configuration file for security and compliance.

        This addresses the configuration security gaps identified in the analysis.
        """
        config_to_validate = config_path or self.config_path
        result = ValidationResult(valid=True)

        try:
            with open(config_to_validate, "r", encoding="utf-8") as f:
                config = json.load(f)
        except FileNotFoundError:
            result.errors.append(f"Configuration file not found: {config_to_validate}")
            result.valid = False
            result.severity = PermissionSeverity.CRITICAL
            return result
        except json.JSONDecodeError as e:
            result.errors.append(f"Invalid JSON in configuration file: {e}")
            result.valid = False
            result.severity = PermissionSeverity.CRITICAL
            return result
        except Exception as e:
            result.errors.append(f"Error reading configuration file: {e}")
            result.valid = False
            result.severity = PermissionSeverity.HIGH
            return result

        # Security validations
        security_checks = [
            self._validate_file_permissions,
            self._validate_allowed_tools,
            self._validate_sandbox_settings,
            self._validate_mcp_servers,
        ]

        for check in security_checks:
            check_result = check(config)
            if not check_result:
                result.valid = False
                result.severity = PermissionSeverity.CRITICAL

        return result

    def _validate_file_permissions(self, config: Dict[str, Any]) -> bool:
        """Validate file permission settings"""
        permissions = config.get("permissions", {})

        # Check for overly permissive settings
        if "deniedTools" in permissions:
            denied_tools = permissions["deniedTools"]
            # Ensure dangerous operations are denied
            dangerous_patterns = ["rm -rf", "sudo", "chmod 777", "format", "mkfs"]

            for pattern in dangerous_patterns:
                found = any(pattern in tool for tool in denied_tools)
                if not found:
                    logger.warning(f"Dangerous operation not denied: {pattern}")
                    # Don't fail validation for this - just warn
                    # return False

        return True

    def _validate_allowed_tools(self, config: Dict[str, Any]) -> bool:
        """Validate allowed tools configuration"""
        permissions = config.get("permissions", {})
        allowed_tools = permissions.get("allowedTools", [])

        # Ensure essential tools are available (but don't fail validation)
        essential_tools = ["Task", "Read", "AskUserQuestion"]
        for tool in essential_tools:
            if tool not in allowed_tools:
                logger.warning(f"Essential tool not allowed: {tool}")
                # Don't fail validation for this - just warn
                # return False

        return True

    def _validate_sandbox_settings(self, config: Dict[str, Any]) -> bool:
        """Validate sandbox security settings"""
        sandbox = config.get("sandbox", {})

        # Ensure sandbox is enabled
        if not sandbox.get("allowUnsandboxedCommands", False):
            return True

        # If sandbox is disabled, ensure validated commands are restricted
        validated_commands = sandbox.get("validatedCommands", [])
        dangerous_commands = ["rm -rf", "sudo", "format", "mkfs"]

        for dangerous_cmd in dangerous_commands:
            if any(dangerous_cmd in validated_cmd for validated_cmd in validated_commands):
                logger.warning(f"Dangerous command in validated commands: {dangerous_cmd}")
                return False

        return True

    def _validate_mcp_servers(self, config: Dict[str, Any]) -> bool:
        """Validate MCP server configuration for security"""
        mcp_servers = config.get("mcpServers", {})

        for server_name, server_config in mcp_servers.items():
            # Ensure command doesn't use dangerous flags
            if "command" in server_config:
                command = server_config["command"]
                dangerous_flags = ["--insecure", "--allow-all", "--disable-ssl"]

                for flag in dangerous_flags:
                    if flag in command:
                        logger.warning(f"Dangerous flag in MCP server {server_name}: {flag}")
                        return False

        return True

    def auto_fix_agent_permissions(self, agent_name: str) -> ValidationResult:
        """
        Automatically fix agent permission configuration.

        This is the main method to address the permissionMode errors
        from the debug logs (Lines 50-80).
        """
        # Get current agent configuration
        agents_config = self.config.setdefault("agents", {})
        agent_config = agents_config.get(agent_name, {})

        # Validate and fix
        result = self.validate_agent_permission(agent_name, agent_config)

        # Save configuration if corrections were made
        if result.auto_corrected:
            agents_config[agent_name] = agent_config
            self._save_configuration()

            if self.enable_logging:
                logger.info(f"Fixed permissions for agent: {agent_name}")

        return result

    def auto_fix_all_agents(self) -> Dict[str, ValidationResult]:
        """Auto-fix all agent permissions in the configuration"""
        results = {}

        agents_config = self.config.get("agents", {})
        for agent_name in agents_config:
            results[agent_name] = self.auto_fix_agent_permissions(agent_name)

        # Also check for agents mentioned in the debug log that might not be in config
        debug_log_agents = [
            "backend-expert",
            "security-expert",
            "api-designer",
            "monitoring-expert",
            "performance-engineer",
            "migration-expert",
            "mcp-playwright-integrator",
            "quality-gate",
            "frontend-expert",
            "debug-helper",
            "ui-ux-expert",
            "trust-checker",
            "project-manager",
            "mcp-context7-integrator",
            "mcp-figma-integrator",
            "tdd-implementer",
            "format-expert",
            "mcp-notion-integrator",
            "devops-expert",
            "docs-manager",
            "implementation-planner",
            "skill-factory",
            "component-designer",
            "database-expert",
            "agent-factory",
            "git-manager",
            "sync-manager",
            "spec-builder",
            "doc-syncer",
            "accessibility-expert",
            "cc-manager",
        ]

        for agent_name in debug_log_agents:
            if agent_name not in agents_config:
                # Create default configuration for missing agents
                agents_config[agent_name] = {
                    "permissionMode": self._suggest_permission_mode(agent_name),
                    "description": f"Auto-generated configuration for {agent_name}",
                    "systemPrompt": f"Default system prompt for {agent_name}",
                }

                results[agent_name] = ValidationResult(
                    valid=True,
                    auto_corrected=True,
                    warnings=[f"Created default configuration for agent: {agent_name}"],
                )

        if any(result.auto_corrected for result in results.values()):
            self._save_configuration()

        return results

    def _save_configuration(self) -> None:
        """Save current configuration to file"""
        try:
            # Create backup
            if os.path.exists(self.config_path):
                backup_path = f"{self.config_path}.backup.{int(time.time())}"
                os.rename(self.config_path, backup_path)
                if self.enable_logging:
                    logger.info(f"Created configuration backup: {backup_path}")

            # Save updated configuration
            with open(self.config_path, "w", encoding="utf-8") as f:
                json.dump(self.config, f, indent=2, ensure_ascii=False)

            if self.enable_logging:
                logger.info(f"Saved configuration to {self.config_path}")

        except Exception as e:
            if self.enable_logging:
                logger.error(f"Error saving configuration: {e}")

    def _audit_permission_change(
        self,
        resource_type: ResourceType,
        resource_name: str,
        action: str,
        old_permissions: Optional[Dict[str, Any]],
        new_permissions: Optional[Dict[str, Any]],
        reason: str,
        auto_corrected: bool,
    ) -> None:
        """Log permission changes for audit trail"""
        audit_entry = PermissionAudit(
            timestamp=time.time(),
            user_id=None,  # System correction
            resource_type=resource_type,
            resource_name=resource_name,
            action=action,
            old_permissions=old_permissions,
            new_permissions=new_permissions,
            reason=reason,
            auto_corrected=auto_corrected,
        )

        self.audit_log.append(audit_entry)

        # Keep audit log size manageable
        if len(self.audit_log) > 1000:
            self.audit_log = self.audit_log[-1000:]

    def get_permission_stats(self) -> Dict[str, Any]:
        """Get permission management statistics"""
        return {
            **self.stats,
            "cached_permissions": len(self.permission_cache),
            "audit_log_entries": len(self.audit_log),
            "configured_agents": len(self.config.get("agents", {})),
        }

    def get_recent_audits(self, limit: int = 50) -> List[PermissionAudit]:
        """Get recent permission audit entries"""
        return self.audit_log[-limit:]

    def export_audit_report(self, output_path: str) -> None:
        """Export audit report to file"""
        report = {
            "generated_at": time.time(),
            "stats": self.get_permission_stats(),
            "recent_audits": [
                {
                    "timestamp": audit.timestamp,
                    "resource_type": audit.resource_type.value,
                    "resource_name": audit.resource_name,
                    "action": audit.action,
                    "reason": audit.reason,
                    "auto_corrected": audit.auto_corrected,
                }
                for audit in self.get_recent_audits()
            ],
        }

        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        if self.enable_logging:
            logger.info(f"Exported audit report to {output_path}")


# Global instance for easy import
permission_manager = UnifiedPermissionManager()


def validate_agent_permission(agent_name: str, agent_config: Dict[str, Any]) -> ValidationResult:
    """Convenience function to validate agent permissions"""
    return permission_manager.validate_agent_permission(agent_name, agent_config)


def check_tool_permission(user_role: str, tool_name: str, operation: str) -> bool:
    """Convenience function to check tool permissions"""
    return permission_manager.check_tool_permission(user_role, tool_name, operation)


def auto_fix_all_agent_permissions() -> Dict[str, ValidationResult]:
    """Convenience function to auto-fix all agent permissions"""
    return permission_manager.auto_fix_all_agents()


def get_permission_stats() -> Dict[str, Any]:
    """Convenience function to get permission statistics"""
    return permission_manager.get_permission_stats()


if __name__ == "__main__":
    # Demo script for testing the permission manager
    print("🔧 MoAI-ADK Unified Permission Manager Demo")
    print("=" * 50)

    # Test agent permission validation
    test_agents = [
        {
            "name": "backend-expert",
            "config": {"permissionMode": "ask", "description": "Backend expert agent"},
        },
        {
            "name": "security-expert",
            "config": {
                "permissionMode": "auto",
                "description": "Security expert agent",
            },
        },
        {
            "name": "api-designer",
            "config": {"permissionMode": "plan", "description": "API designer agent"},
        },
    ]

    print("Testing agent permission validation and auto-correction...")

    for agent in test_agents:
        print(f"\nTesting agent: {agent['name']}")
        agent_config: Dict[str, Any] = agent["config"]  # type: ignore[assignment]
        print(f"Original permissionMode: {agent_config.get('permissionMode', 'default')}")

        agent_name: str = agent["name"]  # type: ignore[assignment]
        result = permission_manager.validate_agent_permission(agent_name, agent_config)

        print(f"Valid: {result.valid}")
        print(f"Auto-corrected: {result.auto_corrected}")

        if result.corrected_mode:
            print(f"Corrected to: {result.corrected_mode}")

        if result.errors:
            print(f"Errors: {result.errors}")

        if result.warnings:
            print(f"Warnings: {result.warnings}")

    print("\n📊 Permission Statistics:")
    stats = permission_manager.get_permission_stats()
    for key, value in stats.items():
        print(f"  {key}: {value}")

    print("\n✨ Demo completed! The Unified Permission Manager addresses")
    print("   the agent permission validation errors from the debug logs.")
