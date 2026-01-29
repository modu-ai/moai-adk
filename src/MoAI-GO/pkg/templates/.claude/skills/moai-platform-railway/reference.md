# Railway Platform Reference

Extended documentation for Railway container deployment configurations and advanced patterns.

---

## Railway Environment Variables

### System-Provided Variables

Railway automatically injects these environment variables into every deployment:

Service Information:
- RAILWAY_SERVICE_NAME: Name of the current service
- RAILWAY_REPLICA_ID: Unique identifier for this replica
- RAILWAY_ENVIRONMENT: Environment name (production, staging)
- RAILWAY_PROJECT_ID: Project identifier
- RAILWAY_DEPLOYMENT_ID: Current deployment identifier

Git Information:
- RAILWAY_GIT_COMMIT_SHA: Full commit SHA
- RAILWAY_GIT_AUTHOR: Commit author
- RAILWAY_GIT_BRANCH: Branch name
- RAILWAY_GIT_REPO_OWNER: Repository owner
- RAILWAY_GIT_REPO_NAME: Repository name

Networking:
- RAILWAY_PRIVATE_DOMAIN: Internal hostname for private networking
- RAILWAY_PUBLIC_DOMAIN: Public hostname (if configured)
- PORT: Port to bind to (Railway expects this)

Volume:
- RAILWAY_VOLUME_NAME: Name of attached volume
- RAILWAY_VOLUME_MOUNT_PATH: Mount path for volume

### Variable Best Practices

Environment-Specific Configuration:
- Use Railway environments (production, staging, development)
- Override variables per environment in dashboard
- Never commit secrets to repository

Secret Management:
- Store secrets in Railway dashboard only
- Use variable references for shared secrets between services
- Rotate secrets through environment variable updates

---

## railway.toml Configuration Reference

### Build Section

```toml
[build]
# Builder selection: DOCKERFILE, RAILPACK, or NIXPACKS (deprecated)
builder = "DOCKERFILE"

# Dockerfile path (relative to railway.toml)
dockerfilePath = "./Dockerfile"

# Build command (for non-Docker builds)
buildCommand = "npm run build"

# Files that trigger rebuild
watchPatterns = ["src/**", "package.json", "Dockerfile"]

# Build arguments
[build.args]
NODE_ENV = "production"
```

### Deploy Section

```toml
[deploy]
# Start command (overrides Dockerfile CMD)
startCommand = "node dist/main.js"

# Health check configuration
healthcheckPath = "/health"
healthcheckTimeout = 60

# Restart policy: ON_FAILURE, ALWAYS, NEVER
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 5

# Number of replicas (1 if using volumes)
numReplicas = 2

# Cron schedule for scheduled services
cronSchedule = "0 * * * *"
```

### Resources Section

```toml
[deploy.resources]
# Memory limit (Mi or Gi)
memory = "512Mi"

# CPU allocation (cores)
cpu = "0.5"
```

### Scaling Section

```toml
[deploy.scaling]
# Replica range
minReplicas = 2
maxReplicas = 10

# Resource-based triggers
targetCPUUtilization = 70
targetMemoryUtilization = 80

# Request-based triggers
targetRequestsPerSecond = 100

# Cooldown settings
scaleDownDelaySeconds = 300
```

### Volume Section

```toml
[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10Gi"
```

### Region Section

```toml
[[deploy.regions]]
name = "us-west1"
replicas = 3

[[deploy.regions]]
name = "europe-west4"
replicas = 2
```

---

## Railway CLI Reference

### Authentication

```bash
railway login              # Interactive login
railway login --browserless  # Token-based login
railway logout             # Logout
railway whoami             # Show current user
```

### Project Management

```bash
railway init               # Initialize new project
railway link               # Link to existing project
railway unlink             # Unlink from project
railway status             # Show project status
```

### Deployment

```bash
railway up                 # Deploy current directory
railway up --detach        # Deploy without waiting for logs
railway up --service NAME  # Deploy specific service
railway up --environment ENV  # Deploy to specific environment
```

### Environment Variables

```bash
railway variables          # List all variables
railway variables --set KEY=value  # Set variable
railway variables --unset KEY      # Remove variable
railway variables --json   # Output as JSON
```

### Logs and Monitoring

```bash
railway logs               # Stream logs
railway logs --service NAME  # Logs for specific service
railway logs --filter error  # Filter logs
railway logs --lines 100   # Limit lines
```

### Deployment Management

```bash
railway deployments        # List deployments
railway deployments list   # Same as above
railway rollback DEPLOY_ID # Rollback to specific deployment
railway rollback --previous  # Rollback to previous deployment
```

### Database Operations

```bash
railway connect            # Connect to database
railway run COMMAND        # Run command with Railway env
```

---

## Managed Services

### PostgreSQL

Configuration:
- Automatic connection pooling via PgBouncer
- Point-in-time recovery
- Automatic backups

Connection URL Format:
```
postgresql://user:password@host:port/database?sslmode=require
```

### Redis

Configuration:
- Persistence enabled by default
- Memory limits based on plan
- Cluster mode available on higher plans

Connection URL Format:
```
redis://default:password@host:port
```

### MySQL

Configuration:
- InnoDB storage engine
- Automatic backups
- Binary logging enabled

Connection URL Format:
```
mysql://user:password@host:port/database
```

---

## Networking Details

### Port Binding

Railway expects applications to bind to the PORT environment variable:

```typescript
const port = process.env.PORT || 3000
app.listen(port, '0.0.0.0', () => {
  console.log(`Server running on port ${port}`)
})
```

Common Pitfall: Binding to localhost (127.0.0.1) instead of 0.0.0.0 will prevent Railway from routing traffic.

### Private Networking Internals

Private domains use Railway's internal DNS:
- Format: {service-name}.railway.internal
- Resolution only works within Railway network
- No additional configuration needed

Service-to-Service Latency:
- Same region: < 1ms
- Cross-region: 50-150ms depending on distance

### SSL/TLS

Public Domains:
- Automatic Let's Encrypt certificates
- Wildcard certificates for custom domains
- HTTP to HTTPS redirect automatic

Private Networking:
- No TLS by default (internal network is encrypted)
- Can add TLS if compliance requires end-to-end encryption

---

## Limits and Quotas

### Free Tier

- $5/month credit
- 512 MB RAM per service
- 1 vCPU per service
- Limited to 2 services per project

### Pro Tier

- Usage-based pricing
- Up to 32 GB RAM per service
- Up to 8 vCPU per service
- Unlimited services
- Priority support

### Resource Limits

Maximum Values:
- Memory: 32 Gi
- CPU: 8 cores
- Volume size: 100 Gi
- Replicas: 20 per service

---

## Troubleshooting

### Common Issues

Build Failures:
- Check Dockerfile syntax
- Verify base image availability
- Review build logs for errors

Deployment Crashes:
- Check health check path returns 200
- Verify PORT binding
- Review application logs

Networking Issues:
- Confirm private domain format
- Check service is running
- Verify variable references

Volume Issues:
- Volumes require numReplicas = 1
- Check mount path permissions
- Verify volume size sufficient

### Debug Commands

```bash
railway logs --service NAME  # Application logs
railway status              # Service status
railway variables           # Environment check
railway run sh              # Shell access (if available)
```

---

## Version History

Version 2.0.0 (2025-12-30):
- Modular documentation structure
- Context7 MCP integration
- Railpack migration guidance
- Enhanced auto-scaling patterns

Version 1.1.0 (2025-11-15):
- Multi-region deployment
- Volume configuration
- CI/CD integration

Version 1.0.0 (2025-10-01):
- Initial release
- Docker deployment patterns
- Multi-service architecture

---

Status: Production Ready | Updated: 2025-12-30
