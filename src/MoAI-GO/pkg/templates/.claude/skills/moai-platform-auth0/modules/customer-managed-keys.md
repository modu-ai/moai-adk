# Customer Managed Keys

Module: moai-platform-auth0/modules/customer-managed-keys.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

Auth0's Customer Managed Keys feature enables organizations to control encryption key lifecycles and implement their own root keys. This service targets highly regulated environments requiring enhanced cryptographic control.

---

## Key Management Methods

### Control Your Own Key

This method allows Key Management Editors to customize the Tenant Master Key lifecycle through Auth0's Key Management Service using the Management API Rekey endpoint.

When to Use:
- Organizations requiring regular key rotation
- Compliance requirements mandating key lifecycle management
- Scenarios where default Auth0 key management is insufficient

### Bring Your Own Key (BYOK)

This method permits administrators to replace the default Auth0 generated Environment Root Key with a new Customer Provided Root Key by importing wrapped encryption keys to FIPS 140-2 L3 Hardware Security Modules.

When to Use:
- Highly regulated industries (finance, healthcare, government)
- Organizations with existing key management infrastructure
- Requirements for complete cryptographic control

---

## Key Hierarchy Architecture

Auth0 implements a four-tier encryption hierarchy:

### Environment Root Key (Layer 1)

Algorithm: RSA/AES variants (depends on cloud provider)
Storage: FIPS 140-2 Level 3 Hardware Security Modules
Purpose: Protects all lower-tier keys in the hierarchy

### Tenant Master Key (Layer 2)

Algorithm: AES-256-GCM
Storage: Auth0 KMS database (encrypted by Environment Root Key)
Purpose: Protects namespace keys for each tenant

### Namespace Key (Layer 3)

Algorithm: AES-256-GCM
Storage: Auth0 KMS database (encrypted by Tenant Master Key)
Purpose: Protects data encryption keys for specific data categories

### Data Encryption Key (Layer 4)

Algorithm: AES-256-GCM
Storage: Stored with encrypted data
Purpose: Directly encrypts tenant data

---

## HSM Integration

### Infrastructure Design

Auth0 deploys independent Hardware Security Modules in highly available, multi-region configurations.

Environment Root Key storage: Maintained in adjacent HSMs with automatic regional failover capabilities.

Shared infrastructure: Single Environment Root Key supports all tenants unless BYOK is implemented.

### Algorithm Variation by Cloud Provider

Azure: RSA 2048 OAEP
AWS: AES 256 GCM

---

## Key Rotation

### Rotation Process

The Rekey endpoint automates rotation by performing the following operations:

Step 1: Deactivate the current Tenant Master Key.

Step 2: Generate a new Tenant Master Key.

Step 3: Re-encrypt the entire key hierarchy with the new key.

Step 4: Update all references to use the new key.

### Rotation Best Practices

Establish a regular rotation schedule based on compliance requirements.

Document rotation procedures and responsible parties.

Test rotation in non-production environments first.

Monitor for any issues during and after rotation.

Maintain backup and recovery procedures.

---

## Implementation Steps

### Prerequisites

Ensure you have the Key Management Editor role assigned.

Verify your tenant is on an eligible plan (Enterprise with appropriate add-ons).

Understand your compliance requirements for key management.

### Step 1: Access Key Management

Navigate to Auth0 Dashboard, then Security, then Key Management.

### Step 2: Review Current Key Status

Examine the current key hierarchy and active keys.

### Step 3: Configure Key Management Policy

Define rotation schedule and procedures.

Set up monitoring and alerting for key events.

### Step 4: Implement Key Rotation (if using Control Your Own Key)

Use the Management API Rekey endpoint to rotate keys.

Monitor the rotation process and verify completion.

### Step 5: Import Customer Key (if using BYOK)

Prepare the wrapped encryption key according to Auth0 specifications.

Import the key through the designated process.

Verify the key is active and functioning correctly.

---

## Monitoring and Compliance

### Log Events

Auth0 automatically generates log events during key operations:

kms_key_state_changed: Key state transitions (active, deactivated, etc.)

kms_key_management_success: Successful key management operations

kms_key_management_failure: Failed key management operations

### Compliance Considerations

Document key management procedures for audit purposes.

Maintain records of key rotation events.

Ensure key management practices align with regulatory requirements.

Implement separation of duties for key management operations.

---

## Access Control

### Key Management Editor Role

Access to key management functions requires the Key Management Editor role, which must be explicitly assigned to tenant members.

Role Capabilities:
- View key status and configuration
- Initiate key rotation
- Import customer-provided keys
- Configure key management policies

### Best Practices for Access Control

Limit the number of users with Key Management Editor role.

Implement multi-factor authentication for key management access.

Maintain audit logs of all key management activities.

Review access permissions regularly.

---

## Disaster Recovery

### Recovery Considerations

Plan for key recovery scenarios in disaster situations.

Document procedures for key restoration.

Test recovery procedures periodically.

Maintain secure backup of key management configuration.

### BYOK Recovery

When using BYOK, ensure you have procedures to re-import customer keys if needed.

Maintain secure backup of the original customer-provided key.

Document the key import process for recovery scenarios.

---

## Related Modules

- highly-regulated-identity.md: HRI feature integration
- fapi-implementation.md: FAPI compliance requirements
- compliance-overview.md: General compliance considerations
- certifications.md: Auth0 security certifications

---

## Resources

Auth0 Documentation: Customer Managed Keys
FIPS 140-2 Reference: NIST security requirements for cryptographic modules
