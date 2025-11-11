# moai-security-authorization: Advanced Authorization Systems

**Expert-Level RBAC, ABAC, and Policy-Based Access Control**  
Trust Score: 9.9/10 | Version: 4.0.0 | Last Updated: 2025-11-11

## 🛡️ Role-Based Access Control (RBAC) Implementation

### Core RBAC Architecture

```python
from enum import Enum
from typing import List, Dict, Set, Optional, Any
from dataclasses import dataclass
import json
from datetime import datetime

class Permission(Enum):
    READ = "read"
    WRITE = "write"
    DELETE = "delete"
    ADMIN = "admin"
    EXECUTE = "execute"
    APPROVE = "approve"
    AUDIT = "audit"

class Role(Enum):
    GUEST = "guest"
    USER = "user"
    MODERATOR = "moderator"
    ADMIN = "admin"
    SUPER_ADMIN = "super_admin"
    DEVELOPER = "developer"
    ANALYST = "analyst"

@dataclass
class Resource:
    id: str
    type: str  # 'document', 'user', 'system', 'api', etc.
    owner_id: Optional[str] = None
    department: Optional[str] = None
    sensitivity_level: str = "public"  # public, internal, confidential, restricted
    metadata: Dict[str, Any] = None
    
    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}

@dataclass
class UserContext:
    user_id: str
    roles: Set[Role]
    department: str
    location: str
    clearance_level: str
    session_id: str
    ip_address: str
    attributes: Dict[str, Any] = None
    
    def __post_init__(self):
        if self.attributes is None:
            self.attributes = {}

class AdvancedRBAC:
    def __init__(self):
        # Role-Permission mapping
        self.role_permissions = {
            Role.GUEST: {Permission.READ},
            Role.USER: {Permission.READ, Permission.WRITE},
            Role.MODERATOR: {Permission.READ, Permission.WRITE, Permission.DELETE},
            Role.ADMIN: {Permission.READ, Permission.WRITE, Permission.DELETE, Permission.AUDIT},
            Role.SUPER_ADMIN: {p for p in Permission},
            Role.DEVELOPER: {Permission.READ, Permission.WRITE, Permission.EXECUTE, Permission.AUDIT},
            Role.ANALYST: {Permission.READ, Permission.AUDIT}
        }
        
        # Department-based permissions
        self.department_resources = {
            "engineering": {"code", "infrastructure", "deployments"},
            "finance": {"financial_reports", "invoices", "budgets"},
            "hr": {"employee_records", "payroll", "benefits"},
            "marketing": {"campaigns", "analytics", "content"}
        }
        
        # Clearance levels
        self.clearance_hierarchy = {
            "public": 0,
            "internal": 1,
            "confidential": 2,
            "restricted": 3
        }
    
    def check_permission(self, user: UserContext, permission: Permission, resource: Resource) -> tuple[bool, str]:
        """
        Check if user has permission for resource using RBAC + contextual rules
        Returns: (has_permission, reason)
        """
        
        # 1. Base permission check from roles
        user_permissions = set()
        for role in.user.roles:
            user_permissions.update(self.role_permissions.get(role, set()))
        
        if permission not in user_permissions:
            return False, f"Permission {permission.value} not granted to any of user's roles"
        
        # 2. Resource ownership check
        if resource.owner_id and resource.owner_id != user.user_id:
            if Role.ADMIN not in user.roles and Role.SUPER_ADMIN not in user.roles:
                return False, "Cannot access another user's resource without admin privileges"
        
        # 3. Department-based access control
        if resource.department and resource.department != user.department:
            if not self._can_cross_departments(user, resource):
                return False, f"Cannot access {resource.department} resources from {user.department}"
        
        # 4. Clearance level check
        required_clearance = self.clearance_hierarchy.get(resource.sensitivity_level, 0)
        user_clearance = self.clearance_hierarchy.get(user.clearance_level, 0)
        
        if user_clearance < required_clearance:
            return False, f"Insufficient clearance level for {resource.sensitivity_level} resource"
        
        # 5. Time-based restrictions
        if not self._check_time_restrictions(user, permission, resource):
            return False, "Access restricted outside business hours"
        
        # 6. Location-based restrictions
        if not self._check_location_restrictions(user, resource):
            return False, "Access restricted from current location"
        
        return True, "Access granted"
    
    def _can_cross_departments(self, user: UserContext, resource: Resource) -> bool:
        """Check if user can access resources from other departments"""
        if Role.SUPER_ADMIN in user.roles:
            return True
        
        # Cross-department access rules
        cross_access_rules = {
            "engineering": ["finance"],  # Engineers can access finance for billing
            "finance": ["engineering"],  # Finance can access engineering for budgets
        }
        
        allowed_departments = cross_access_rules.get(user.department, [])
        return resource.department in allowed_departments
    
    def _check_time_restrictions(self, user: UserContext, permission: Permission, resource: Resource) -> bool:
        """Check time-based access restrictions"""
        hour = datetime.now().hour
        
        # Business hours (9 AM - 6 PM) for sensitive operations
        if permission in [Permission.DELETE, Permission.ADMIN] and resource.sensitivity_level in ["confidential", "restricted"]:
            return 9 <= hour <= 18
        
        # No restrictions for regular operations
        return True
    
    def _check_location_restrictions(self, user: UserContext, resource: Resource) -> bool:
        """Check location-based access restrictions"""
        if resource.sensitivity_level == "restricted":
            # Only allow from office locations
            office_locations = ["US", "UK", "DE"]  # ISO country codes
            return user.location in office_locations
        
        return True

class AttributeBasedAccessControl(ABAC):
    """
    Attribute-Based Access Control implementation
    """
    
    def __init__(self):
        self.policies = []
    
    def add_policy(self, policy: dict) -> None:
        """Add ABAC policy"""
        self.policies.append(policy)
    
    def evaluate_access(self, user: UserContext, resource: Resource, action: str) -> tuple[bool, str]:
        """
        Evaluate access using ABAC policies
        Returns: (allowed, reason)
        """
        
        for policy in self.policies:
            if self._matches_policy(user, resource, action, policy):
                if policy.get("effect") == "deny":
                    return False, f"Access denied by policy: {policy.get('name', 'unnamed')}"
                elif policy.get("effect") == "allow":
                    return True, f"Access granted by policy: {policy.get('name', 'unnamed')}"
        
        # Default deny if no policies match
        return False, "No policy allows this access"
    
    def _matches_policy(self, user: UserContext, resource: Resource, action: str, policy: dict) -> bool:
        """Check if request matches policy conditions"""
        
        # Check subject conditions
        subject_conditions = policy.get("subject", {})
        if not self._match_conditions(user.__dict__, subject_conditions):
            return False
        
        # Check resource conditions
        resource_conditions = policy.get("resource", {})
        if not self._match_conditions(resource.__dict__, resource_conditions):
            return False
        
        # Check action conditions
        action_conditions = policy.get("action", {})
        if not self._match_conditions({"action": action}, action_conditions):
            return False
        
        # Check environment conditions
        env_conditions = policy.get("environment", {})
        env_context = {
            "time": datetime.now().hour,
            "day": datetime.now().weekday(),
            "ip_address": user.ip_address
        }
        
        if not self._match_conditions(env_context, env_conditions):
            return False
        
        return True
    
    def _match_conditions(self, context: dict, conditions: dict) -> bool:
        """Check if context matches all conditions"""
        for key, expected_value in conditions.items():
            actual_value = context.get(key)
            
            if isinstance(expected_value, dict):
                # Handle operators like >, <, contains, etc.
                if not self._evaluate_condition(actual_value, expected_value):
                    return False
            elif actual_value != expected_value:
                return False
        
        return True
    
    def _evaluate_condition(self, actual: Any, condition: dict) -> bool:
        """Evaluate complex conditions with operators"""
        for operator, value in condition.items():
            if operator == "gt" and not (actual > value):
                return False
            elif operator == "lt" and not (actual < value):
                return False
            elif operator == "gte" and not (actual >= value):
                return False
            elif operator == "lte" and not (actual <= value):
                return False
            elif operator == "contains" and value not in str(actual):
                return False
            elif operator == "in" and actual not in value:
                return False
        
        return True

class PolicyEngine:
    """
    Centralized policy engine combining RBAC, ABAC, and custom policies
    """
    
    def __init__(self):
        self.rbac = AdvancedRBAC()
        self.abac = AttributeBasedAccessControl()
        self.custom_policies = []
        
        # Add default ABAC policies
        self._setup_default_policies()
    
    def _setup_default_policies(self) -> None:
        """Setup default security policies"""
        
        # Policy: Allow read access to public resources for all authenticated users
        self.abac.add_policy({
            "name": "Public Read Access",
            "effect": "allow",
            "subject": {"roles": ["user", "admin", "super_admin"]},
            "resource": {"sensitivity_level": "public"},
            "action": {"action": "read"}
        })
        
        # Policy: Deny delete operations outside business hours for sensitive resources
        self.abac.add_policy({
            "name": "Business Hours Delete Restriction",
            "effect": "deny",
            "resource": {"sensitivity_level": ["confidential", "restricted"]},
            "action": {"action": "delete"},
            "environment": {"time": {"lt": 9, "gt": 18}}
        })
        
        # Policy: Allow finance users to access financial data
        self.abac.add_policy({
            "name": "Finance Data Access",
            "effect": "allow",
            "subject": {"department": "finance"},
            "resource": {"department": "finance"},
            "action": {"action": ["read", "write"]}
        })
    
    def check_access(self, user: UserContext, resource: Resource, permission: Permission) -> tuple[bool, str]:
        """
        Unified access check using multiple policy types
        Returns: (access_granted, reason)
        """
        
        # 1. Check RBAC permissions first
        rbac_allowed, rbac_reason = self.rbac.check_permission(user, permission, resource)
        if not rbac_allowed:
            return False, f"RBAC denied: {rbac_reason}"
        
        # 2. Check ABAC policies
        abac_allowed, abac_reason = self.abac.evaluate_access(user, resource, permission.value)
        
        # ABAC deny overrides RBAC allow
        if not abac_allowed:
            return False, f"ABAC denied: {abac_reason}"
        
        # 3. Check custom policies
        for policy in self.custom_policies:
            policy_result = policy.evaluate(user, resource, permission)
            if not policy_result.allowed:
                return False, f"Custom policy denied: {policy_result.reason}"
        
        return True, f"Access granted: {abac_reason}"

class AuditLogger:
    """
    Security audit logging for authorization decisions
    """
    
    def __init__(self):
        self.audit_log = []
    
    def log_access_attempt(self, 
                          user: UserContext, 
                          resource: Resource, 
                          permission: Permission, 
                          granted: bool, 
                          reason: str) -> None:
        """Log authorization decision for audit purposes"""
        
        audit_entry = {
            "timestamp": datetime.now().isoformat(),
            "user_id": user.user_id,
            "user_roles": [role.value for role in user.roles],
            "resource_id": resource.id,
            "resource_type": resource.type,
            "permission": permission.value,
            "granted": granted,
            "reason": reason,
            "ip_address": user.ip_address,
            "session_id": user.session_id,
            "department": user.department,
            "resource_sensitivity": resource.sensitivity_level
        }
        
        self.audit_log.append(audit_entry)
        
        # In production, send to SIEM or audit database
        print(f"AUDIT: {json.dumps(audit_entry)}")
    
    def generate_access_report(self, user_id: str, start_date: datetime, end_date: datetime) -> Dict[str, Any]:
        """Generate access report for compliance audits"""
        
        filtered_logs = [
            entry for entry in self.audit_log
            if entry["user_id"] == user_id
            and start_date <= datetime.fromisoformat(entry["timestamp"]) <= end_date
        ]
        
        total_attempts = len(filtered_logs)
        granted_attempts = sum(1 for entry in filtered_logs if entry["granted"])
        denied_attempts = total_attempts - granted_attempts
        
        resource_access = {}
        for entry in filtered_logs:
            resource_type = entry["resource_type"]
            if resource_type not in resource_access:
                resource_access[resource_type] = {"granted": 0, "denied": 0}
            
            if entry["granted"]:
                resource_access[resource_type]["granted"] += 1
            else:
                resource_access[resource_type]["denied"] += 1
        
        return {
            "user_id": user_id,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat()
            },
            "summary": {
                "total_attempts": total_attempts,
                "granted": granted_attempts,
                "denied": denied_attempts,
                "success_rate": (granted_attempts / total_attempts * 100) if total_attempts > 0 else 0
            },
            "resource_access": resource_access,
            "denied_reasons": [
                entry["reason"] for entry in filtered_logs 
                if not entry["granted"]
            ]
        }

# Example usage
if __name__ == "__main__":
    # Initialize policy engine
    policy_engine = PolicyEngine()
    audit_logger = AuditLogger()
    
    # Create test user
    user = UserContext(
        user_id="user123",
        roles={Role.USER},
        department="engineering",
        location="US",
        clearance_level="internal",
        session_id="sess123",
        ip_address="192.168.1.100"
    )
    
    # Create test resource
    resource = Resource(
        id="doc456",
        type="document",
        owner_id="user123",
        department="engineering",
        sensitivity_level="internal"
    )
    
    # Check access
    granted, reason = policy_engine.check_access(user, resource, Permission.READ)
    audit_logger.log_access_attempt(user, resource, Permission.READ, granted, reason)
    
    print(f"Access granted: {granted}, Reason: {reason}")
