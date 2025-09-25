"""
@FEATURE:CORE-001 Core functionality modules for MoAI-ADK

@REQ:CORE-MODULES-001 → @DESIGN:CORE-ARCHITECTURE-001 → @TASK:CORE-INTEGRATION-001 → @TEST:CORE-MODULES-001

@DESIGN:CORE-ARCHITECTURE-001 Modular core system with clean separation of concerns
@TASK:CORE-INTEGRATION-001 Integration layer for all core components

This package contains the fundamental components:
- @TASK:SECURITY-001 Security validation and safe operations
- @TASK:CONFIG-MGMT-001 Configuration management
- @TASK:DIR-MGMT-001 Directory operations
- @TASK:FILE-MGMT-001 File operations
- @TASK:GIT-INTEGRATION-001 Git integration
- @TASK:SYSTEM-UTILS-001 System utilities
- @TASK:TEMPLATE-ENGINE-001 Template processing
- @TASK:VALIDATION-001 Validation utilities
- @TASK:VERSION-SYNC-001 Version synchronization
"""

from .config_manager import ConfigManager
from .security import SecurityError, SecurityManager
from .template_engine import TemplateEngine
from .validator import validate_claude_code, validate_python_version
from .version_sync import VersionSyncManager

__all__ = [
    'ConfigManager',
    'SecurityError',
    'SecurityManager',
    'TemplateEngine',
    'VersionSyncManager',
    'validate_claude_code',
    'validate_python_version'
]
