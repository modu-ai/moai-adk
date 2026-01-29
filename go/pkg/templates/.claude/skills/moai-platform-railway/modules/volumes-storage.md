# Volumes and Storage Module

Purpose: Persistent volume configuration, storage patterns, and data management for stateful Railway applications.

---

## Volume Configuration

Basic Volume Definition:
```toml
[deploy]
startCommand = "npm start"

[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10Gi"
```

Multiple Volumes:
```toml
[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10Gi"

[[volumes]]
mountPath = "/app/uploads"
name = "user-uploads"
size = "50Gi"

[[volumes]]
mountPath = "/app/cache"
name = "cache-storage"
size = "5Gi"
```

Volume Size Options:
- 1Gi to 100Gi for standard workloads
- Size can be increased but not decreased
- Cost scales with provisioned size

---

## Storage Patterns

### File Storage Service

```typescript
import { join } from 'path'
import { existsSync, mkdirSync, writeFileSync, readFileSync, unlinkSync } from 'fs'
import { readdir, stat } from 'fs/promises'

const VOLUME_PATH = process.env.RAILWAY_VOLUME_MOUNT_PATH || '/app/data'

class PersistentStorage {
  private basePath: string

  constructor(subdir?: string) {
    this.basePath = subdir ? join(VOLUME_PATH, subdir) : VOLUME_PATH
    if (!existsSync(this.basePath)) {
      mkdirSync(this.basePath, { recursive: true })
    }
  }

  write(filename: string, data: Buffer | string): void {
    const filepath = join(this.basePath, filename)
    writeFileSync(filepath, data)
  }

  read(filename: string): Buffer {
    const filepath = join(this.basePath, filename)
    return readFileSync(filepath)
  }

  exists(filename: string): boolean {
    return existsSync(join(this.basePath, filename))
  }

  delete(filename: string): void {
    const filepath = join(this.basePath, filename)
    if (existsSync(filepath)) {
      unlinkSync(filepath)
    }
  }

  async list(): Promise<string[]> {
    return readdir(this.basePath)
  }

  async getSize(filename: string): Promise<number> {
    const stats = await stat(join(this.basePath, filename))
    return stats.size
  }
}

export const storage = new PersistentStorage()
export const uploadsStorage = new PersistentStorage('uploads')
export const cacheStorage = new PersistentStorage('cache')
```

### Upload Handler

```typescript
import multer from 'multer'
import { v4 as uuid } from 'uuid'
import path from 'path'

const UPLOAD_PATH = process.env.RAILWAY_VOLUME_MOUNT_PATH
  ? `${process.env.RAILWAY_VOLUME_MOUNT_PATH}/uploads`
  : '/app/uploads'

const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, UPLOAD_PATH)
  },
  filename: (req, file, cb) => {
    const ext = path.extname(file.originalname)
    cb(null, `${uuid()}${ext}`)
  }
})

const upload = multer({
  storage,
  limits: { fileSize: 10 * 1024 * 1024 }, // 10MB
  fileFilter: (req, file, cb) => {
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'application/pdf']
    cb(null, allowed.includes(file.mimetype))
  }
})

// Route handler
app.post('/upload', upload.single('file'), (req, res) => {
  if (!req.file) {
    return res.status(400).json({ error: 'No file uploaded' })
  }
  res.json({
    filename: req.file.filename,
    size: req.file.size,
    mimetype: req.file.mimetype
  })
})
```

### SQLite on Persistent Volume

```typescript
import Database from 'better-sqlite3'
import { join } from 'path'

const DB_PATH = process.env.RAILWAY_VOLUME_MOUNT_PATH
  ? join(process.env.RAILWAY_VOLUME_MOUNT_PATH, 'app.db')
  : '/app/data/app.db'

const db = new Database(DB_PATH, {
  verbose: process.env.NODE_ENV === 'development' ? console.log : undefined
})

// Enable WAL mode for better concurrent access
db.pragma('journal_mode = WAL')
db.pragma('synchronous = NORMAL')
db.pragma('cache_size = -64000') // 64MB cache

// Initialize schema
db.exec(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  )
`)

export { db }
```

---

## Cache Storage Patterns

### Disk-Based Cache

```typescript
import { join } from 'path'
import { existsSync, writeFileSync, readFileSync, statSync, unlinkSync } from 'fs'
import { createHash } from 'crypto'

const CACHE_PATH = process.env.RAILWAY_VOLUME_MOUNT_PATH
  ? `${process.env.RAILWAY_VOLUME_MOUNT_PATH}/cache`
  : '/app/cache'

interface CacheEntry<T> {
  data: T
  expiresAt: number
}

class DiskCache {
  private path: string

  constructor() {
    this.path = CACHE_PATH
  }

  private getFilePath(key: string): string {
    const hash = createHash('md5').update(key).digest('hex')
    return join(this.path, `${hash}.json`)
  }

  set<T>(key: string, data: T, ttlSeconds: number = 3600): void {
    const entry: CacheEntry<T> = {
      data,
      expiresAt: Date.now() + ttlSeconds * 1000
    }
    writeFileSync(this.getFilePath(key), JSON.stringify(entry))
  }

  get<T>(key: string): T | null {
    const filepath = this.getFilePath(key)
    if (!existsSync(filepath)) return null

    const entry: CacheEntry<T> = JSON.parse(readFileSync(filepath, 'utf-8'))
    if (Date.now() > entry.expiresAt) {
      unlinkSync(filepath)
      return null
    }
    return entry.data
  }

  delete(key: string): void {
    const filepath = this.getFilePath(key)
    if (existsSync(filepath)) {
      unlinkSync(filepath)
    }
  }
}

export const cache = new DiskCache()
```

---

## Data Backup Patterns

### Scheduled Backup Script

```typescript
import { exec } from 'child_process'
import { promisify } from 'util'
import { createReadStream } from 'fs'
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3'

const execAsync = promisify(exec)

const s3 = new S3Client({
  region: process.env.AWS_REGION,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!
  }
})

async function backupVolume(): Promise<void> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
  const backupName = `backup-${timestamp}.tar.gz`
  const volumePath = process.env.RAILWAY_VOLUME_MOUNT_PATH || '/app/data'
  const tempPath = `/tmp/${backupName}`

  // Create compressed archive
  await execAsync(`tar -czf ${tempPath} -C ${volumePath} .`)

  // Upload to S3
  await s3.send(new PutObjectCommand({
    Bucket: process.env.BACKUP_BUCKET!,
    Key: `railway-backups/${backupName}`,
    Body: createReadStream(tempPath)
  }))

  // Cleanup
  await execAsync(`rm ${tempPath}`)
  console.log(`Backup completed: ${backupName}`)
}

// Run backup
backupVolume().catch(console.error)
```

### Restore Script

```typescript
import { exec } from 'child_process'
import { promisify } from 'util'
import { createWriteStream } from 'fs'
import { S3Client, GetObjectCommand } from '@aws-sdk/client-s3'
import { Readable } from 'stream'

const execAsync = promisify(exec)

async function restoreVolume(backupKey: string): Promise<void> {
  const volumePath = process.env.RAILWAY_VOLUME_MOUNT_PATH || '/app/data'
  const tempPath = '/tmp/restore.tar.gz'

  // Download from S3
  const response = await s3.send(new GetObjectCommand({
    Bucket: process.env.BACKUP_BUCKET!,
    Key: backupKey
  }))

  // Write to temp file
  const stream = response.Body as Readable
  const writeStream = createWriteStream(tempPath)
  await new Promise((resolve, reject) => {
    stream.pipe(writeStream).on('finish', resolve).on('error', reject)
  })

  // Extract to volume
  await execAsync(`tar -xzf ${tempPath} -C ${volumePath}`)

  // Cleanup
  await execAsync(`rm ${tempPath}`)
  console.log(`Restore completed from: ${backupKey}`)
}
```

---

## Volume Considerations

Replica Limitations:
- Volumes are tied to a single replica
- numReplicas must be 1 when using volumes
- For multi-replica deployments, use external storage (S3, R2)

Performance Characteristics:
- SSD-backed storage
- Low latency for local reads/writes
- Size can be increased dynamically
- Data persists across deployments

Volume vs External Storage Decision:
- Use volumes for: SQLite, local caches, temporary processing
- Use external storage for: User uploads, shared assets, large files

---

## Works Well With

- modules/docker-deployment.md: Container configuration
- modules/multi-service.md: Shared storage patterns
- moai-domain-database: Database storage patterns

---

Version: 1.0.0 | Updated: 2025-12-30
