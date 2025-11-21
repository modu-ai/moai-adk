## API Reference

### Core Firebase Operations
- `batch_update_documents(updates)` - Atomic batch document updates
- `subscribe_to_realtime_updates(collection, filters, callback)` - Real-time subscriptions
- `authenticate_user(uid, customClaims)` - User authentication with claims
- `upload_file(filePath, data, metadata)` - Secure file upload
- `call_function(functionName, data)` - Cloud Functions invocation

### Context7 Integration
- `get_latest_firebase_documentation()` - Firebase docs via Context7
- `analyze_firestore_patterns()` - Database patterns via Context7
- `optimize_realtime_architecture()` - Real-time optimization via Context7

## Best Practices (November 2025)

### DO
- Use Firestore for new projects over Realtime Database
- Implement proper security rules for database access
- Use Cloud Functions for serverless backend logic
- Optimize database queries with proper indexing
- Implement proper error handling and retry logic
- Use Firebase Authentication with custom claims
- Monitor performance and costs regularly
- Implement proper backup and recovery procedures

### DON'T
- Use Realtime Database for new projects
- Skip security rules validation
- Ignore database performance optimization
- Forget to implement proper error handling
- Skip Firebase Security Rules testing
- Ignore cost monitoring and optimization
- Forget to implement proper logging
- Skip backup and disaster recovery planning

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture)
- `moai-domain-backend` (Backend Firebase integration)
- `moai-domain-frontend` (Frontend Firebase integration)
- `moai-security-api` (Firebase security implementation)
- `moai-essentials-perf` (Performance optimization)
- `moai-foundation-trust` (Security and compliance)
- `moai-domain-mobile-app` (Mobile Firebase integration)
- `moai-baas-supabase-ext` (PostgreSQL alternative)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Firebase platform updates, and advanced Google Cloud integration
- **v2.0.0** (2025-11-11): Complete metadata structure, Firebase patterns, Google Cloud integration
- **v1.0.0** (2025-11-11): Initial Firebase platform

---

**End of Skill** | Updated 2025-11-13

## Firebase Platform Integration

### Google Cloud Ecosystem
- BigQuery integration for advanced analytics
- Cloud Run for scalable container deployment
- Cloud Logging and Monitoring for observability
- Cloud Build for CI/CD pipeline integration
- Cloud Storage for scalable file management

### Mobile-First Features
- Real-time synchronization across devices
- Offline data persistence and sync
- Push notifications and messaging
- Authentication with social providers
- Global CDN for fast content delivery

---

**End of Enterprise Firebase Platform Expert **
