---
name: moai-baas-firebase-ext
description: Enterprise Firebase Platform with AI-powered Google Cloud integration. Modular guide for Firestore, Authentication, and Cloud Functions
version: 1.0.0
modularized: true
allowed-tools:
  - Read
last_updated: 2025-11-22
compliance_score: 85
auto_trigger_keywords:
  - api
  - authentication
  - baas
  - ext
  - firebase
  - testing
category_tier: 1
---

## Quick Reference (30 seconds)

# Enterprise Firebase Platform Expert 

---

## When to Use

**Automatic triggers**:
- Firebase architecture and Google Cloud integration discussions
- Real-time database and Firestore implementation planning
- Mobile backend as a service strategy development
- Firebase security and performance optimization

**Manual invocation**:
- Designing enterprise Firebase architectures with optimal patterns
- Implementing comprehensive real-time features and synchronization
- Planning Firebase to Google Cloud migration strategies
- Optimizing Firebase performance and cost management

---

## Quick Reference (Level 1)

### What It Does

Enterprise-grade Firebase platform guidance with AI-powered optimization, focusing on:
- **Firestore**: Real-time database operations, queries, security
- **Authentication**: Multi-provider auth, custom claims, token management
- **Cloud Functions**: Serverless functions, triggers, integrations
- **Storage**: File uploads, metadata, CDN integration
- **Performance**: Optimization patterns, cost management

---

### Core Modules

This skill is modularized for optimal loading:

**Module 1: Firestore Operations** (`SKILL-firestore.md`)
- Advanced Firestore implementation
- Real-time subscriptions
- Data modeling best practices
- Query optimization and indexing
- Security rules

**Module 2: Authentication** (`SKILL-auth.md`)
- User authentication patterns
- Custom claims and roles
- Multi-provider authentication
- Token management
- Security best practices

**Module 3: Cloud Functions** (`SKILL-functions.md`)
- Serverless functions with Python
- Trigger types (Firestore, Storage, Auth, Scheduler)
- Error handling patterns
- Integration examples

---

### Quick Decision Tree

```
Start
  ├─ Need Firestore operations? → SKILL-firestore.md
  ├─ Need Authentication? → SKILL-auth.md
  ├─ Need Cloud Functions? → SKILL-functions.md
  └─ Need complete integration? → All modules
```

---

## Firebase Platform Ecosystem (November 2025)

### Core Firebase Services
- **Firestore**: NoSQL document database with real-time sync
- **Authentication**: Multi-provider user authentication
- **Cloud Functions**: Serverless backend logic (Python, Node.js)
- **Cloud Storage**: File storage with CDN integration
- **Hosting**: Static site hosting with global CDN
- **Analytics**: User behavior and crash reporting

### Latest Features (November 2025)
- **App Check**: Bot and abuse prevention
- **Security Rules**: Fine-grained access control
- **Extensions**: Pre-built backend solutions
- **Emulator Suite**: Local development environment

### Google Cloud Integration
- **Cloud Run**: Container-based functions
- **Cloud SQL**: Relational database integration
- **BigQuery**: Analytics data warehouse
- **Cloud Logging**: Centralized logging

### Performance Characteristics
- **Read Latency**: <100ms (99th percentile)
- **Write Latency**: <200ms (99th percentile)
- **Scalability**: Millions of concurrent connections
- **Availability**: 99.95% uptime SLA

---

## Best Practices Checklist

**Must-Have:**
- ✅ Use Firestore security rules for all collections
- ✅ Implement proper authentication with custom claims
- ✅ Create composite indexes for complex queries
- ✅ Use batch operations for multiple writes
- ✅ Implement error handling in Cloud Functions

**Recommended:**
- ✅ Denormalize data for read performance
- ✅ Use Firestore emulator for local development
- ✅ Implement rate limiting in Cloud Functions
- ✅ Monitor costs with Cloud Billing alerts
- ✅ Use App Check for production apps

**Security:**
- 🔒 Never expose API keys in client code
- 🔒 Validate all inputs in Cloud Functions
- 🔒 Use HTTPS-only for all endpoints
- 🔒 Implement proper token verification
- 🔒 Regular security rules testing

---

## Official References

**Primary Documentation:**
- [SKILL-firestore.md](/moai-baas-firebase-ext/SKILL-firestore.md) – Firestore operations & real-time features
- [SKILL-auth.md](/moai-baas-firebase-ext/SKILL-auth.md) – Authentication & security
- [SKILL-functions.md](/moai-baas-firebase-ext/SKILL-functions.md) – Cloud Functions & triggers

**External Resources:**
- [Firebase Documentation](https://firebase.google.com/docs)
- [Firestore Best Practices](https://firebase.google.com/docs/firestore/best-practices)
- [Firebase Security Rules](https://firebase.google.com/docs/rules)

---

## Version History

**4.0.0** (2025-11-12)
- ✨ Modular structure with 3 sub-skills
- ✨ Enhanced Progressive Disclosure
- ✨ Comprehensive Firestore patterns
- ✨ Advanced Authentication examples
- ✨ Cloud Functions with Python
- ✨ Security best practices

---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (backend-expert)

---

## Integration with Other Skills

**Prerequisite Skills:**
- Skill("moai-backend-api") – API design patterns
- Skill("moai-security-api") – Security best practices

**Complementary Skills:**
- Skill("moai-baas-vercel-ext") – Vercel integration
- Skill("moai-devops-docker") – Containerization

**Next Steps:**
- After setup: Use Skill("moai-frontend-react") for client integration
- For testing: Use Skill("moai-essentials-test") for Firebase emulator

---

**End of Skill** | Updated 2025-11-12