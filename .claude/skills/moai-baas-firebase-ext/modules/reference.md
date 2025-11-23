# Firebase API Reference & Resources

Complete API reference and official documentation links for Firebase services.

## Firebase Services Reference

### Firestore (Document Database)

**Core Operations**:
- `getDoc(docRef)` - Get single document
- `getDocs(query)` - Get multiple documents
- `setDoc(docRef, data)` - Create or overwrite
- `updateDoc(docRef, data)` - Update fields
- `deleteDoc(docRef)` - Delete document
- `batch()` - Batch operations (max 500)

**Query Operators**:
- `==`, `!=`, `<`, `<=`, `>`, `>=` - Comparison
- `array-contains` - Array membership
- `in`, `not-in` - List membership
- `array-contains-any` - Any array value

**Limitations**:
- Max 20K docs in single query response
- Composite indexes needed for complex queries
- Writes are eventually consistent
- Max 1 MB per document

### Authentication

**Sign-in Methods**:
- Email/Password
- Social (Google, GitHub, Facebook, etc.)
- Anonymous
- Custom tokens
- SAML / OpenID Connect

**Custom Claims**:
- Set per user: `admin`, `role`, `organization`
- Verified in ID token
- Max 1000 bytes per user

### Cloud Functions

**Supported Triggers**:
- `https` - HTTP requests
- `onDocumentCreated` - New document
- `onDocumentUpdated` - Document changed
- `onDocumentDeleted` - Document removed
- `onAuthUserCreated` - New user
- `onObjectFinalized` - File uploaded
- Scheduled - Cron-based (Cloud Scheduler)

**Cold Start Times**:
- TypeScript: ~1-2 seconds
- Python: ~1-3 seconds
- Go: <500ms

### Cloud Storage

**Key Features**:
- 99.999% durability
- Global CDN
- Signed URLs for temporary access
- Custom metadata storage
- Object versioning (optional)

**Limits**:
- Max 5 TB file size
- Max 1 GB upload size (for single request)
- Storage per project: 1 PB

### Firebase Hosting

**Features**:
- Global CDN with edge caching
- SSL certificates (automatic)
- Custom domains
- Rollback capability
- Preview URLs for PRs

**Performance**:
- <50ms global latency
- Automatic compression
- HTTP/2 push

---

## Best Practices Checklist

**Architecture**:
- [ ] Denormalize data appropriately
- [ ] Use subcollections for 1-to-many relationships
- [ ] Plan collection structure before implementation
- [ ] Document your data model

**Security**:
- [ ] Implement Firestore Security Rules
- [ ] Validate input on both client and server
- [ ] Use environment variables for secrets
- [ ] Enable audit logging
- [ ] Regularly audit access patterns

**Performance**:
- [ ] Use pagination for large result sets
- [ ] Cache frequently accessed data
- [ ] Use indexes for complex queries
- [ ] Monitor query performance
- [ ] Limit real-time listeners

**Costs**:
- [ ] Monitor billing dashboard
- [ ] Use bulk operations for batch writes
- [ ] Implement data retention policies
- [ ] Archive old data to Cloud Storage
- [ ] Review Cloud Logging costs

---

## Troubleshooting Common Issues

### Issue: "PERMISSION_DENIED" Error

**Causes**:
- Security Rules blocking operation
- User not authenticated
- Custom claims not set

**Solution**:
```typescript
// Check Security Rules in Firebase Console
// Enable logging: enableLogging(true)
// Verify user authentication before operations
if (user) {
  // Proceed with operation
} else {
  // Show login screen
}
```

### Issue: Slow Queries (>1 second)

**Causes**:
- Missing indexes
- Querying large collection
- Network latency

**Solution**:
```typescript
// Create composite index via Firebase Console
// Use pagination for large datasets
// Add orderBy to existing indexed field
```

### Issue: Real-Time Listener Not Updating

**Causes**:
- Unsubscribe called prematurely
- Security Rules deny read access
- Data not actually changing

**Solution**:
```typescript
// Save unsubscribe function
const unsubscribe = onSnapshot(...)
// Call unsubscribe only when truly done
// not on component re-render
```

---

## Official Documentation

- [Firebase Firestore Documentation](https://firebase.google.com/docs/firestore)
- [Firebase Authentication](https://firebase.google.com/docs/auth)
- [Cloud Functions](https://firebase.google.com/docs/functions)
- [Cloud Storage](https://firebase.google.com/docs/storage)
- [Firebase Hosting](https://firebase.google.com/docs/hosting)

---

## Community Resources

**Official**:
- [Firebase YouTube Channel](https://www.youtube.com/@firebase)
- [Firebase Blog](https://firebase.googleblog.com/)
- [Stack Overflow - firebase tag](https://stackoverflow.com/questions/tagged/firebase)

**Guides**:
- [Firestore Best Practices](https://firebase.google.com/docs/firestore/best-practices)
- [Security Rules Guide](https://firebase.google.com/docs/rules/basics)
- [Performance Optimization](https://firebase.google.com/docs/performance)
