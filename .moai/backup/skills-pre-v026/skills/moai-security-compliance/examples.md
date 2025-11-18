# Regulatory Compliance - Production Code Examples

## Example 1: Winston Audit Logger

```javascript
// logging/audit-logger.js
const winston = require('winston');
const mongoDb = require('winston-mongodb');

class AuditLogger {
  constructor(config) {
    this.db = config.db;
    this.logger = this.createLogger();
  }
  
  createLogger() {
    return winston.createLogger({
      format: winston.format.combine(
        winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
        winston.format.json(),
        winston.format.printf(({ timestamp, level, message, ...meta }) => ({
          timestamp,
          level,
          message,
          ...meta,
          retention_days: 2555,  // 7 years for GDPR
        }))
      ),
      transports: [
        // File storage
        new winston.transports.File({
          filename: './logs/audit.log',
          maxsize: 5242880,  // 5MB
          maxFiles: 5,
        }),
        // Database storage
        new mongoDb.MongoDB({
          db: this.db,
          collection: 'audit_logs',
          options: { useUnifiedTopology: true },
        }),
      ],
    });
  }
  
  logUserAccess(userId, action, resource, result) {
    this.logger.info('User access', {
      userId,
      action,
      resource,
      result,
      timestamp: new Date().toISOString(),
      compliance: {
        gdpr: true,
        soc2: true,
      },
    });
  }
  
  logDataAccess(userId, dataType, action) {
    this.logger.info('Data access', {
      userId,
      dataType,
      action,
      timestamp: new Date().toISOString(),
      classification: this.classifyData(dataType),
    });
  }
  
  classifyData(dataType) {
    const classifications = {
      'pii': 'SENSITIVE',
      'phi': 'RESTRICTED',
      'payment_card': 'CONFIDENTIAL',
    };
    return classifications[dataType] || 'INTERNAL';
  }
}

module.exports = AuditLogger;
```

## Example 2: GDPR Right-to-Erasure

```javascript
// compliance/gdpr-erasure.js
class GDPRErasureHandler {
  constructor(db) {
    this.db = db;
  }
  
  async requestErasure(userId) {
    // Schedule deletion for 30 days
    await this.db.users.updateOne(
      { _id: userId },
      {
        $set: {
          deletion_requested_at: new Date(),
          deletion_scheduled_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
        },
      }
    );
    
    return { success: true, deleteAt: 30 };  // days
  }
  
  async executeErasure(userId) {
    // 1. Delete user data
    await this.db.users.deleteOne({ _id: userId });
    await this.db.user_profiles.deleteMany({ userId });
    await this.db.user_preferences.deleteMany({ userId });
    
    // 2. Anonymize audit logs
    await this.db.audit_logs.updateMany(
      { userId },
      {
        $set: {
          userId: null,
          anonymized: true,
          anonymized_at: new Date(),
        },
      }
    );
    
    // 3. Log erasure
    await this.db.erasure_logs.insertOne({
      userId,
      erasedAt: new Date(),
      reason: 'GDPR right-to-erasure',
    });
  }
  
  async processScheduledErasures() {
    const now = new Date();
    
    const usersToErase = await this.db.users.find({
      deletion_scheduled_at: { $lte: now },
      deleted: false,
    });
    
    for (const user of usersToErase) {
      await this.executeErasure(user._id);
    }
  }
}

module.exports = GDPRErasureHandler;
```

## Example 3: Data Retention Policies

```javascript
// compliance/retention-manager.js
const cron = require('node-cron');

class RetentionManager {
  constructor(db) {
    this.db = db;
    this.retentionPolicies = {
      'audit_logs': 2555,  // 7 years
      'payment_logs': 2555,  // 7 years (PCI DSS)
      'access_logs': 365,  // 1 year (SOC 2)
      'session_logs': 90,  // 90 days
    };
  }
  
  initialize() {
    // Run daily at 2 AM
    cron.schedule('0 2 * * *', async () => {
      console.log('Running retention cleanup');
      await this.enforceRetention();
    });
  }
  
  async enforceRetention() {
    for (const [collection, days] of Object.entries(this.retentionPolicies)) {
      const cutoffDate = new Date();
      cutoffDate.setDate(cutoffDate.getDate() - days);
      
      const result = await this.db[collection].deleteMany({
        timestamp: { $lt: cutoffDate },
      });
      
      console.log(
        `Retention: Deleted ${result.deletedCount} records from ${collection}`
      );
    }
  }
  
  async archiveToS3(logs) {
    // Archive old logs to S3 Glacier for long-term storage
    const AWS = require('aws-sdk');
    const s3 = new AWS.S3();
    
    const key = `archive/${new Date().getFullYear()}/${logs._id}.gz`;
    
    await s3
      .putObject({
        Bucket: 'compliance-archive',
        Key: key,
        Body: this.compress(logs),
        StorageClass: 'GLACIER',  // Cost-effective long-term storage
      })
      .promise();
  }
  
  compress(data) {
    // Compression implementation
    return data;
  }
}

module.exports = RetentionManager;
```

## Example 4: SOC 2 Evidence Collection

```javascript
// compliance/soc2-evidence-collector.js
class SOC2EvidenceCollector {
  constructor(db) {
    this.db = db;
  }
  
  async collectAccessControlEvidence() {
    return {
      access_policies: await this.getAccessPolicies(),
      mfa_enforcement: await this.checkMFAStatus(),
      access_reviews: await this.getMonthlyAccessReviews(),
      provisioning_logs: await this.getProvisioningLogs(),
    };
  }
  
  async collectSecurityMonitoring() {
    return {
      incident_logs: await this.db.incidents.find({
        timestamp: { $gte: this.auditStartDate },
      }).toArray(),
      
      alert_logs: await this.db.security_alerts.find({
        timestamp: { $gte: this.auditStartDate },
      }).toArray(),
      
      vulnerability_scans: await this.db.vuln_scans.find({
        timestamp: { $gte: this.auditStartDate },
      }).toArray(),
    };
  }
  
  async generateSOC2Report() {
    const evidence = {
      auditPeriod: {
        start: this.auditStartDate,
        end: this.auditEndDate,
      },
      
      domains: {
        'CC': await this.collectAccessControlEvidence(),
        'IL': await this.collectSecurityMonitoring(),
      },
    };
    
    return evidence;
  }
  
  async getMonthlyAccessReviews() {
    return await this.db.access_reviews.find({
      timestamp: { $gte: this.auditStartDate },
    }).toArray();
  }
  
  checkMFAStatus() {
    // Check all users have MFA enabled
    return this.db.users.find({ mfa_enabled: true }).toArray();
  }
}

module.exports = SOC2EvidenceCollector;
```

## Example 5: Context7 MCP Compliance Monitoring

```javascript
// compliance/context7-compliance.js
const { Context7Client } = require('context7-mcp');

class ComplianceMonitoring {
  constructor(apiKey) {
    this.context7 = new Context7Client({ apiKey });
  }
  
  async checkGDPRCompliance() {
    return await this.context7.query({
      type: 'compliance_check',
      framework: 'GDPR',
      checks: [
        'data_classification',
        'consent_management',
        'data_retention',
        'dpia_required',
      ],
    });
  }
  
  async checkSOC2Compliance() {
    return await this.context7.query({
      type: 'compliance_check',
      framework: 'SOC2',
      checks: [
        'access_control',
        'change_management',
        'security_monitoring',
        'incident_response',
      ],
    });
  }
  
  async getComplianceDashboard() {
    const frameworks = ['GDPR', 'HIPAA', 'SOC2', 'ISO27001'];
    
    const results = {};
    
    for (const framework of frameworks) {
      results[framework] = await this.context7.query({
        type: 'compliance_status',
        framework,
      });
    }
    
    return results;
  }
}

module.exports = ComplianceMonitoring;
```

